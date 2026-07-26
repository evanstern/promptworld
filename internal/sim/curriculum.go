package sim

import (
	"encoding/json"
	"fmt"

	"github.com/evanstern/promptworld/internal/store"
)

// The curriculum ladder's world-visible facts (spec 046, contracts/events.md):
// exercise passes and stage unlocks, recorded as events so a world's history is
// the auditable proof of what was earned in it (FR-007/FR-008 — the per-user
// unlocks record is a projection of these, never an input). Both types are the
// EXECUTOR emission class (the metatron.order_expired precedent): pure
// functions of (state, tick), never mind- or operator-injected, so NO
// whitelist entries exist for them. The production emitter is the spec-054
// scenario rubric machinery (scenario.go, TASK-119): this file ships the
// payload contracts, the reducer arms, and the exercise content it consumes.

const (
	// curriculumPassRetain bounds retained pass records (the guardianOrderRetain
	// precedent): the most recent 32 passes keep the audit trail readable without
	// unbounded state growth. StagesUnlocked needs no cap — it latches at most
	// one entry per ladder stage.
	curriculumPassRetain = 32
)

// EvidenceRef points at one event in THIS world's history that contributed to
// a pass — a rubric term's satisfying event, or (stage-2/stage-3 gates) the
// player-authored fact in force at pass time (spec 044/046). The (type, seq,
// tick) triple is enough to re-locate the event in the log, so the claim
// stays independently auditable (FR-008).
type EvidenceRef struct {
	Type string `json:"type"`
	Seq  int64  `json:"seq"`
	Tick int64  `json:"tick"`
	// Custom marks this evidence entry as a PLAYER-AUTHORED/PLAYER-GRANTED
	// fact, as opposed to the world's default — the single flag both gate
	// conjuncts above stage-1 read (contracts/unlocks-record.md "Gate
	// conjuncts"): for a stage-2 pass, Type=="metatron.charter_observed" with
	// Custom==true means a player-authored charter revision was in force at
	// pass time; for a stage-3 pass, any evidence entry with Custom==true
	// names a player-granted tool's contributing act (as opposed to a tool
	// granted by the default/stage manifest alone). Absent/false = a
	// default-fact evidence entry, which never satisfies a gate conjunct
	// (SC-004's negative case).
	//
	// For charter evidence, Custom is NEVER asserted freehand: it is derived
	// from the recorded event's real payload as the inverse of
	// CharterObservedPayload.Default (spec 044 US2 — Default==true means the
	// world's default/preset charter was in force, i.e. NOT player-authored)
	// by CharterObservedEvidence below, the single sanctioned constructor, so
	// the gate conjunct and the recorded payload can never disagree.
	Custom bool `json:"custom,omitempty"`
}

// ExercisePassedPayload is curriculum.exercise_passed: a seeded exercise's
// event-derived rubric reached its pass signal. Outcome-shaped — Evidence
// lists the satisfying events the unlock derivation (and any later audit)
// reads.
type ExercisePassedPayload struct {
	Exercise string        `json:"exercise"`
	Stage    string        `json:"stage"`
	Tick     int64         `json:"tick"`
	Evidence []EvidenceRef `json:"evidence,omitempty"`
}

// StageUnlockedPayload is curriculum.stage_unlocked: a recorded pass satisfied
// the gate conjuncts to the named stage. Emitted exactly once per (world,
// stage) — the reducer rejects a duplicate latch.
type StageUnlockedPayload struct {
	Stage    string `json:"stage"`
	Exercise string `json:"exercise"`
	Tick     int64  `json:"tick"`
}

// CurriculumPass is one recorded pass on state (bounded ring, newest last) —
// the replay-derived record the unlock derivation evaluates and status
// surfaces read.
type CurriculumPass struct {
	Exercise string        `json:"exercise"`
	Stage    string        `json:"stage"`
	Tick     int64         `json:"tick"`
	Evidence []EvidenceRef `json:"evidence,omitempty"`
}

// validLadderStage is the reducer-side closed stage vocabulary — the
// deterministic twin of world.ValidStage's manifest check (kept local: the
// deterministic core does not import the save-directory package). "" is NOT
// valid here — an event must always name its stage.
func validLadderStage(s string) bool {
	switch s {
	case "stage-1", "stage-2", "stage-3", "stage-4":
		return true
	}
	return false
}

// applyCurriculum is the reducer arm for curriculum.* events. It validates
// rather than clamps (the guardian arm's contract): the executor emits these
// as pure functions of recorded state, so a recorded event always re-applies
// cleanly at the same position in replay, and a malformed fixture is rejected
// at the door.
func (s *State) applyCurriculum(e store.Event) error {
	switch e.Type {
	case "curriculum.exercise_passed":
		var p ExercisePassedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if p.Exercise == "" {
			return fmt.Errorf("apply %s: empty exercise id", e.Type)
		}
		if !validLadderStage(p.Stage) {
			return fmt.Errorf("apply %s: unknown stage %q", e.Type, p.Stage)
		}
		s.CurriculumPasses = append(s.CurriculumPasses, CurriculumPass{
			Exercise: p.Exercise, Stage: p.Stage, Tick: p.Tick, Evidence: p.Evidence,
		})
		if drop := len(s.CurriculumPasses) - curriculumPassRetain; drop > 0 {
			s.CurriculumPasses = append([]CurriculumPass(nil), s.CurriculumPasses[drop:]...)
		}
	case "curriculum.stage_unlocked":
		var p StageUnlockedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		// stage-1 is the ladder's floor — every player holds it unearned — so
		// only stages 2..4 are ever unlocked.
		if !validLadderStage(p.Stage) || p.Stage == "stage-1" {
			return fmt.Errorf("apply %s: %q is not an unlockable stage", e.Type, p.Stage)
		}
		if p.Exercise == "" {
			return fmt.Errorf("apply %s: empty exercise id", e.Type)
		}
		// Once per (world, stage) — a duplicate latch is rejected at the door.
		// Deliberately NO cross-check against CurriculumPasses: that record is
		// bounded (pruned past 32), so the gate-conjunct evaluation happens at
		// emission time (TASK-119 / the US3 derivation), never on re-apply.
		for _, st := range s.StagesUnlocked {
			if st == p.Stage {
				return fmt.Errorf("apply %s: %s already unlocked in this world", e.Type, p.Stage)
			}
		}
		s.StagesUnlocked = append(s.StagesUnlocked, p.Stage)
	}
	return nil
}

// nextLadderStage returns the stage a pass at stage unlocks, and whether one
// exists (stage-4 is graduation — nothing unlocks past it, per the ladder's
// synthesis decision 3).
func nextLadderStage(stage string) (string, bool) {
	switch stage {
	case "stage-1":
		return "stage-2", true
	case "stage-2":
		return "stage-3", true
	case "stage-3":
		return "stage-4", true
	}
	return "", false
}

// EvaluateUnlock is the gate-conjunct decision (spec 046 US3, T012,
// contracts/unlocks-record.md "Gate conjuncts"): given a recorded
// exercise_passed payload and the fold state it was recorded against, it
// decides whether the pass earns the next stage. Pure over (state, pass) —
// the executor emission class — so it is safe to call from a test fixture
// today and from TASK-119's rubric machinery once it lands, with identical
// semantics either way:
//
//	stage-1 -> stage-2: ANY stage-1 exercise pass (the ladder's floor asks
//	           nothing more than attempting it — FR-007).
//	stage-2 -> stage-3: the pass's evidence must include a player-authored
//	           charter revision in force at pass time — Type ==
//	           "metatron.charter_observed" AND Custom == true, where Custom
//	           is derived (CharterObservedEvidence) as the inverse of the
//	           recorded CharterObservedPayload.Default (spec 044 US2).
//	           SC-004: a default/preset-charter pass must NOT satisfy this —
//	           a stage-1 tutor-preset world's observation records
//	           Default==true (the preset is the game's authorship, not the
//	           player's), so it never opens this gate.
//	stage-3 -> stage-4: the pass's evidence must include a player-granted
//	           tool's contributing act — any evidence entry with Custom ==
//	           true (a fixed event type isn't pinned by the contract: which
//	           tool contributed is TASK-119's exercise design).
//
// Reconciled with spec 044 US2 (T022): the real metatron.charter_observed
// event (CharterObservedPayload{Fingerprint, Default} — metatron.go, spec
// 044 contracts/events.md) landed on main while this feature was in flight.
// The gate keeps reading EvidenceRef.Custom (the pass payload stays
// self-contained and replay-pure — no log scan here), and honesty is pushed
// to construction: CharterObservedEvidence is the only sanctioned way to
// build a charter evidence entry, and it sets Custom = !payload.Default.
//
// Already-unlocked stages are NOT re-unlocked (StagesUnlocked is consulted)
// — exactly once per (world, stage), matching the reducer's own duplicate
// rejection (belt-and-suspenders: this is the pre-emission check; the
// reducer arm is the door).
func EvaluateUnlock(s *State, pass ExercisePassedPayload) (stage string, ok bool) {
	next, hasNext := nextLadderStage(pass.Stage)
	if !hasNext {
		return "", false
	}
	for _, st := range s.StagesUnlocked {
		if st == next {
			return "", false
		}
	}
	switch pass.Stage {
	case "stage-1":
		return next, true
	case "stage-2":
		for _, ev := range pass.Evidence {
			if ev.Type == "metatron.charter_observed" && ev.Custom {
				return next, true
			}
		}
		return "", false
	case "stage-3":
		for _, ev := range pass.Evidence {
			if ev.Custom {
				return next, true
			}
		}
		return "", false
	}
	return "", false
}

// CharterObservedEvidence derives the gate-facing EvidenceRef from a recorded
// metatron.charter_observed event (spec 044 US2; reconciled here per 046
// T022). EvidenceRef.Custom is the honest INVERSE of the payload's Default
// flag: Default==true means the world's default/preset charter was in force
// (authored by the game — a stage-1 tutor-preset world records this), so
// Custom==false and the stage-2→3 gate stays shut; Default==false means a
// player-authored revision was in force, so Custom==true. This is the ONLY
// sanctioned constructor for a charter evidence entry — the spec-054 rubric
// machinery (scenario.go) is its production consumer once a stage-2 exercise
// gains a production evaluator (the first-night evaluator needs no charter
// conjunct; test fixtures use it today) — so EvaluateUnlock's conjunct and
// the recorded payload can never disagree.
func CharterObservedEvidence(e store.Event) (EvidenceRef, error) {
	if e.Type != "metatron.charter_observed" {
		return EvidenceRef{}, fmt.Errorf("charter evidence: %q is not metatron.charter_observed", e.Type)
	}
	var p CharterObservedPayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return EvidenceRef{}, fmt.Errorf("charter evidence: %w", err)
	}
	return EvidenceRef{Type: e.Type, Seq: e.Seq, Tick: e.Tick, Custom: !p.Default}, nil
}

// CharterEvidenceFromState is the STATE-SOURCED charter evidence constructor
// (spec 077 FR-004/FR-005, research R7) — the third sanctioned EvidenceRef
// constructor, beside CharterObservedEvidence (event-sourced; fixtures) and
// OrderPlacedEvidence. It reads the coordinates the metatron.charter_observed
// reducer arm persists (CharterObservedSeq/Tick — the PlacedSeq apply-time
// stamp precedent) plus the same-arm authorship flag, so the honesty
// derivation stays in sanctioned-constructor land: Custom is CharterCustom
// (already the inverse of the recorded payload's Default — spec 072), never
// asserted freehand. ok=false when no observation's coordinates are on state
// (Seq 0 — a pre-077 snapshot whose charter arm predates the stamp): the
// evidence entry is honestly ABSENT and the caller's pass waits for the next
// charter observation to stamp them — self-healing degradation, never a
// fabricated coordinate (spec edge "pre-077 the-law world").
func CharterEvidenceFromState(s *State) (EvidenceRef, bool) {
	if s.CharterObservedSeq == 0 {
		return EvidenceRef{}, false
	}
	return EvidenceRef{Type: "metatron.charter_observed", Seq: s.CharterObservedSeq,
		Tick: s.CharterObservedTick, Custom: s.CharterCustom}, true
}

// SkillsObservedEvidence is the fourth sanctioned EvidenceRef constructor
// (spec 077 FR-006, research R8): the stage-3 gate's long-deferred
// "player-granted tool's contributing act" evidence. It reads the
// coordinates the metatron.skills_observed reducer arm persists. Custom is
// true BY CONSTRUCTION — the honest twin of the charter's derived-inverse
// rule: no game-shipped skill files exist and stages 1–2 lock binding out
// entirely (stageSkills), so a bound, observed skill set is player-authored
// by structural necessity, not by assertion. ok=false when no observation is
// on state (Seq 0) — the evidence entry is honestly absent.
func SkillsObservedEvidence(s *State) (EvidenceRef, bool) {
	if s.SkillsObservedSeq == 0 {
		return EvidenceRef{}, false
	}
	return EvidenceRef{Type: "metatron.skills_observed", Seq: s.SkillsObservedSeq,
		Tick: s.SkillsObservedTick, Custom: true}, true
}

// ExerciseDefinition is a seeded scenario exercise — CONTENT, not machinery
// (spec 046 US4, contracts/exercises.md): stage, seed, framing, an
// event-derived rubric, the pass signal shape, and the chronicle framing for
// its score narrative. The spec-054 scenario/rubric machinery (scenario.go)
// is the consumer; spec 046 guaranteed the definitions parse and every
// RubricTerm names a cataloged event type (proven in internal/tui, which
// owns the digest catalog — see
// TestExerciseRubricTermsAreCatalogedEventTypes).
type ExerciseDefinition struct {
	// ID is the exercise's stable content id (curriculum.exercise_passed's
	// Exercise field names it).
	ID string
	// Stage is the ladder stage this exercise teaches and gates (world.Stage1
	// etc.) — the exercise is played AT this stage and, on pass, may unlock
	// the next (sim.EvaluateUnlock).
	Stage string
	// Seed tunes a deterministic world for this exercise (contracts/
	// exercises.md: "exact seed pinned at implementation").
	Seed uint64
	// Concept is the one prompt-engineering concept this exercise teaches
	// (the ladder's one-concept-per-stage discipline, spec.md).
	Concept string
	// Framing is the in-fiction incident setup text.
	Framing string
	// RubricTerms are the event-derived terms a pass must satisfy — every
	// entry MUST be a cataloged event type (a digestRegistry key).
	RubricTerms []string
	// PassSignal documents the curriculum.exercise_passed shape this
	// exercise's rubric drives toward — descriptive content, not executable:
	// the spec-054 machinery (scenarioRubricEvents) is what actually emits it.
	PassSignal string
	// ScoreNarrative frames how the chronicle should tell the attempt's
	// story, win or lose (FR-011 — failure is a story, not a scold).
	ScoreNarrative string
	// Schedule is the exercise's authored incident schedule (spec 054 US2):
	// deterministic pressure landed by the executor's incident source at
	// authored game times — compiled to absolute ticks at arm time
	// (ArmScenario), validated by TestScenarioSchedulesCompile. Content, like
	// RubricTerms; an exercise with no schedule simply arms an empty source.
	Schedule []IncidentScheduleEntry
	// IncidentVisibility optionally overrides the stage-keyed default for how
	// much of the schedule the exercise panel forecasts (spec 054 FR-009,
	// reorientation D4): "" = the stage default; else a visibility vocabulary
	// value (VisibilityForecast/VisibilityFog). A vocabulary, never a boolean.
	IncidentVisibility string
	// BoundaryDay is the exercise's pass-boundary declaration (spec 077
	// FR-003, research R6 — a two-value vocabulary, always dawn-shaped):
	// N > 0 evaluates the rubric exactly at dawn of day N (first-night: 2);
	// 0 is rolling — every dawn from day 2 until a pass lands. Content,
	// like RubricTerms; boundaryDue (scenario.go) is the one reader.
	BoundaryDay int
}

// ExerciseByID looks def up in the shipped catalog — the single resolution
// path the daemon's boot arming, `promptworld new --scenario`, and the
// exercise panel all share (spec 054). The manifest-side twin of this
// catalog's id set is world.ValidScenarioExercise (kept local there: the
// save-directory package and the deterministic core do not import each
// other); TestScenarioCatalogMirrorsWorldVocabulary pins the two in sync.
func ExerciseByID(id string) (ExerciseDefinition, bool) {
	for _, def := range ScenarioExercises {
		if def.ID == id {
			return def, true
		}
	}
	return ExerciseDefinition{}, false
}

// OrderPlacedEvidence derives the gate-facing EvidenceRef for a watch that
// contributed to a pass, from the standing order's state record (spec 054
// FR-004) — the second sanctioned evidence constructor, beside
// CharterObservedEvidence above. The (Seq, Tick) coordinates re-locating the
// metatron.order_placed event come from the reducer's own apply-time stamp
// (GuardianOrder.PlacedSeq/PlacedTick — the Memory.Seq precedent), so the
// claim stays independently auditable without a log scan. Custom stays
// false: a stage-1 watch is placed through a stage-manifest-granted tool
// (monitor_and_act is in the ratified stage-1 ceiling), and only a
// player-GRANTED tool's contributing act may assert Custom (the stage-3
// gate conjunct's vocabulary above).
func OrderPlacedEvidence(o GuardianOrder) (EvidenceRef, error) {
	if o.PlacedSeq == 0 {
		return EvidenceRef{}, fmt.Errorf("order evidence: %s carries no recorded placement seq", o.ID)
	}
	return EvidenceRef{Type: "metatron.order_placed", Seq: o.PlacedSeq, Tick: o.PlacedTick}, nil
}

// FirstNightExercise is the stage-1 shipped exercise (contracts/
// exercises.md): keep the village alive through night one, teaching the
// player to ask for the right watch — visions and orders through
// conversation. The ratified stage-1 ceiling amendment (monitor_and_act/
// cancel_order in the ceiling) is what makes the order-evidence rubric term
// below actually playable at stage-1.
var FirstNightExercise = ExerciseDefinition{
	ID:          "first-night",
	Stage:       "stage-1",
	Seed:        46101,
	BoundaryDay: 2, // dawn of day 2 — the shipped boundary, now declared (spec 077 FR-003)
	Concept:     "asking for the right watch — visions and orders through conversation",
	Framing:     "a seeded world tuned so night one is survivable only if the guardian is directed well: fuel scarce, the gru active",
	RubricTerms: []string{
		"sim.day_started",       // dawn of day 2 reached, run not ended
		"agent.died",            // rubric wants ZERO of these before dawn
		"metatron.nudged",       // a vision/omen landed before nightfall
		"metatron.order_placed", // OR a watch was set before nightfall (ratified amendment)
	},
	PassSignal:     `curriculum.exercise_passed{exercise: "first-night", stage: "stage-1"}`,
	ScoreNarrative: "the night's chronicle chapter is the telling — the village's first night under a new watcher",
	// The authored pressure (spec 054 US2): the gru emerges at nightfall of
	// night one, at the northern map edge beside the seed-46101 tree belt
	// ("near the north woods") — (44,0) is a passable, unprotected border
	// tile on this exercise's own map (pinned by
	// TestFirstNightSchedulePositionValid), exactly the tile class the random
	// emergence path draws from. The schedule preempts that night's random
	// roll (scenario.go, research R3).
	Schedule: []IncidentScheduleEntry{
		{Kind: IncidentGruEmerges, Day: 1, Time: "22:00", X: 44, Y: 0},
	},
}

// TheLawExercise is the stage-2 shipped exercise (contracts/exercises.md):
// get a norm adopted, teaching durable instruction — policy that must
// outlive the conversation lives in the charter. Its rubric's charter-fingerprint
// term is the SC-004 gate conjunct: a default-charter pass must not unlock
// stage-3 (sim.EvaluateUnlock enforces this over the pass's evidence, not
// this content definition).
var TheLawExercise = ExerciseDefinition{
	ID:      "the-law",
	Stage:   "stage-2",
	Seed:    46102,
	Concept: "durable instruction — policy that must outlive the conversation lives in the charter",
	Framing: "a seeded world with a norm-shaped problem (a nighttime curfew vs. fuel gathering) requiring sustained, consistent guardian behavior across several days",
	RubricTerms: []string{
		"meeting.proposal_resolved", // the village norm/vote resolves in the instructed direction
		"metatron.charter_observed", // a player-authored charter revision in force across the window (spec 044 US2; Custom derived via CharterObservedEvidence)
	},
	PassSignal:     `curriculum.exercise_passed{exercise: "the-law", stage: "stage-2"}`,
	ScoreNarrative: "the governance arc as narrated by the chronicle — the charter as the law behind the law",
	// BoundaryDay 0 = rolling (spec 077 FR-003): the pass lands at the first
	// dawn (from day 2) the rubric holds — sustained-behavior exercises have
	// no fixed judgment day. This completes spec 072 FR-009's deferral: the
	// evidence blocker is gone (CharterEvidenceFromState reads the persisted
	// CharterObservedSeq/Tick), so the-law now EMITS.
}

// --- the spec-077 exercise catalog wave (FR-001, data-model §5 normative) ---
//
// Nine hand-authored exercises, 3/2/2/2 by stage, seeds 46101–46109 (unique;
// TestScenarioExerciseSeedsUnique). Every rubric term names a cataloged event
// type (TestExerciseRubricTermsAreCatalogedEventTypes); every Met derivation
// is a pure state fact (EvaluateRubric's arms, scenario.go); every schedule
// compiles at boot (TestScenarioSchedulesCompile) with authored positions
// pinned valid on each exercise's own seed's map
// (TestSchedulePositionsValidPerSeed).

// ColdDawnExercise — stage-1: the cold-snap variant of the first night. The
// snap doubles the night's bite (warmthLossColdSnap), so an undirected
// village freezes; the taught move is the same watch-and-visions
// conversation first-night teaches, under harder weather.
var ColdDawnExercise = ExerciseDefinition{
	ID:          "cold-dawn",
	Stage:       "stage-1",
	Seed:        46103,
	BoundaryDay: 2,
	Concept:     "asking for the right watch when the pressure is weather, not a predator",
	Framing:     "a seeded world where a cold snap grips the first night — fuel is everything, and the guardian must be told what to watch",
	RubricTerms: []string{
		"sim.day_started",       // dawn of day 2 reached, run not ended
		"agent.died",            // zero exposure-cause deaths (no villager freezes)
		"metatron.order_placed", // a watch set before nightfall (the first-night predicate, shared)
	},
	PassSignal:     `curriculum.exercise_passed{exercise: "cold-dawn", stage: "stage-1"}`,
	ScoreNarrative: "the chronicle tells the night the cold came early — and who kept the fires fed",
	Schedule: []IncidentScheduleEntry{
		{Kind: IncidentColdSnap, Day: 1, Time: "22:00", Hours: 8},
	},
}

// StrangerAtTheGateExercise — stage-1: the trickster variant. Nothing hunts
// the villagers; the larder is the stake, and firelight is the answer.
var StrangerAtTheGateExercise = ExerciseDefinition{
	ID:          "stranger-at-the-gate",
	Stage:       "stage-1",
	Seed:        46104,
	BoundaryDay: 2,
	Concept:     "directing attention — the guardian can only guard what it is told matters",
	Framing:     "a seeded world where a stranger slips in on the first night, after whatever the village leaves unattended in the dark",
	RubricTerms: []string{
		"sim.day_started", // dawn of day 2 reached
		"agent.died",      // zero deaths
		"stranger.took",   // rubric wants ZERO of these (zero-wanted — renders honestly per spec 072)
	},
	PassSignal:     `curriculum.exercise_passed{exercise: "stranger-at-the-gate", stage: "stage-1"}`,
	ScoreNarrative: "the chronicle tells of the visitor no one invited — and what was, or wasn't, missing at dawn",
	// (3,0): a passable, unprotected border tile on the seed-46104 map
	// (pinned by TestSchedulePositionsValidPerSeed) — the same tile class
	// the gru's spawn path draws from.
	Schedule: []IncidentScheduleEntry{
		{Kind: IncidentStrangerArrives, Day: 1, Time: "23:00", X: 3, Y: 0},
	},
}

// blightedLarderFoodFloor is blighted-larder's banked-food threshold —
// CONTENT, pinned beside the definition (data-model §5): enough stored food
// (chests + piles, storedFoodTotal) to ride out the blighted patch, tuned
// only by editing this constant.
const blightedLarderFoodFloor = 10

// BlightedLarderExercise — stage-2: durable instruction against a slow
// pressure. The blight lands on day 2 and the boundary is day 4 — only a
// standing policy (charter) that keeps the village banking food survives the
// gap; conversation alone won't hold for two days.
var BlightedLarderExercise = ExerciseDefinition{
	ID:          "blighted-larder",
	Stage:       "stage-2",
	Seed:        46105,
	BoundaryDay: 4,
	Concept:     "policy that outlives the conversation — a charter rule against a slow-burning pressure",
	Framing:     "a seeded world whose forage belt blights on the second morning — the larder you banked under your charter is what remains",
	RubricTerms: []string{
		"metatron.charter_observed", // a player-authored charter in force
		"agent.died",                // zero starvation-cause deaths
		"agent.deposited",           // a larder banked (storedFoodTotal >= blightedLarderFoodFloor)
	},
	PassSignal:     `curriculum.exercise_passed{exercise: "blighted-larder", stage: "stage-2"}`,
	ScoreNarrative: "the chronicle tells the lean days — the law behind the law, and the chest that held",
	// (24,41) r4: the seed-46105 forage belt's densest patch (pinned by
	// TestSchedulePositionsValidPerSeed — blightable at genesis).
	Schedule: []IncidentScheduleEntry{
		{Kind: IncidentForageBlight, Day: 2, Time: "08:00", X: 24, Y: 41, Radius: 4},
	},
}

// ToolsmithExercise — stage-3: the capability-design stage's own concept —
// author a skill file, see the guardian observed running under it, and see
// it act (a player order placed under the observed set). Rolling boundary:
// the pass lands at the first dawn the whole chain holds.
var ToolsmithExercise = ExerciseDefinition{
	ID:      "toolsmith",
	Stage:   "stage-3",
	Seed:    46106,
	Concept: "shaping what the guardian can do — a skill file is capability you author",
	Framing: "a seeded world with no scripted pressure: the trial is the workshop — bind a skill file, and put the guardian to work under it",
	RubricTerms: []string{
		"metatron.skills_observed", // your skill file guides the guardian
		"metatron.order_placed",    // the guardian acts under it (a player order since the observation)
		"agent.died",               // zero deaths
	},
	PassSignal:     `curriculum.exercise_passed{exercise: "toolsmith", stage: "stage-3"}`,
	ScoreNarrative: "the chronicle tells the guardian learning your craft — the first working under your own tools",
}

// FogWatchExercise — stage-3: fog visibility by stage default — the schedule
// is NOT forecast; the skill file must be in force before trials the player
// cannot see coming (a snap, then the gru).
var FogWatchExercise = ExerciseDefinition{
	ID:          "fog-watch",
	Stage:       "stage-3",
	Seed:        46107,
	BoundaryDay: 3,
	Concept:     "designing for the unseen — capability that holds when you don't know what's coming",
	Framing:     "a seeded world under fog: trials land unannounced across two nights, and only what you authored beforehand guards the village",
	RubricTerms: []string{
		"sim.day_started",          // dawn of day 3 reached
		"agent.died",               // zero deaths
		"metatron.skills_observed", // a skill file in force before the trials
	},
	PassSignal:     `curriculum.exercise_passed{exercise: "fog-watch", stage: "stage-3"}`,
	ScoreNarrative: "the chronicle tells two nights weathered blind — the craft, proven in fog",
	// (2,0): a passable, unprotected border tile on the seed-46107 map.
	Schedule: []IncidentScheduleEntry{
		{Kind: IncidentColdSnap, Day: 1, Time: "22:00", Hours: 6},
		{Kind: IncidentGruEmerges, Day: 2, Time: "22:00", X: 2, Y: 0},
	},
}

// LongWinterExercise — stage-4: every pressure kind at once, across three
// nights — the mastery capstone's endurance half.
var LongWinterExercise = ExerciseDefinition{
	ID:          "long-winter",
	Stage:       "stage-4",
	Seed:        46108,
	BoundaryDay: 4,
	Concept:     "stewardship under compound pressure — everything you've built, all at once",
	Framing:     "a seeded world where the winter turns feral: a snap, a blight, a stranger, and the gru — three nights, nothing spared",
	RubricTerms: []string{
		"sim.day_started", // dawn of day 4 reached
		"agent.died",      // zero deaths
		"stranger.took",   // nothing is taken (zero-wanted)
	},
	PassSignal:     `curriculum.exercise_passed{exercise: "long-winter", stage: "stage-4"}`,
	ScoreNarrative: "the chronicle tells the long winter — the village that came through whole, or the one that didn't",
	// Bottom-border stranger entry (0,63) and gru tile (2,63); blight center
	// (20,19) r4 — all pinned on the seed-46108 map.
	Schedule: []IncidentScheduleEntry{
		{Kind: IncidentColdSnap, Day: 1, Time: "22:00", Hours: 8},
		{Kind: IncidentForageBlight, Day: 2, Time: "08:00", X: 20, Y: 19, Radius: 4},
		{Kind: IncidentStrangerArrives, Day: 2, Time: "23:00", X: 0, Y: 63},
		{Kind: IncidentGruEmerges, Day: 3, Time: "22:00", X: 2, Y: 63},
	},
}

// StewardsChargeExercise — stage-4: the whole instruction ladder in force at
// once — village law, player charter, player skill file — with no one lost.
// Rolling boundary: graduation lands the first dawn the stewardship holds.
var StewardsChargeExercise = ExerciseDefinition{
	ID:      "stewards-charge",
	Stage:   "stage-4",
	Seed:    46109,
	Concept: "the full stewardship — law, charter, and craft held together",
	Framing: "a seeded world with no scripted pressure: the trial is governance itself — a law adopted, a charter in force, a skill file guiding, and every villager alive at dawn",
	RubricTerms: []string{
		"meeting.proposal_resolved", // a village law adopted
		"metatron.charter_observed", // a player-authored charter in force
		"metatron.skills_observed",  // your skill file guides the guardian
		"agent.died",                // zero deaths
	},
	PassSignal:     `curriculum.exercise_passed{exercise: "stewards-charge", stage: "stage-4"}`,
	ScoreNarrative: "the chronicle tells a village governed well — the steward's charge, kept",
}

// ScenarioExercises is the shipped exercise catalog (spec 077 FR-001: nine
// hand-authored exercises, 3/2/2/2 by stage — grown from spec 046's two).
var ScenarioExercises = []ExerciseDefinition{
	FirstNightExercise, ColdDawnExercise, StrangerAtTheGateExercise,
	TheLawExercise, BlightedLarderExercise,
	ToolsmithExercise, FogWatchExercise,
	LongWinterExercise, StewardsChargeExercise,
}
