package sim

import (
	"fmt"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Scenario machinery (spec 054): the production half of the spec-046
// curriculum substrate — an authored incident scheduler and the rubric
// evaluator that emits curriculum.exercise_passed / curriculum.stage_unlocked.
// Everything here is the EXECUTOR emission class (the charge-regen /
// order-expired precedent): pure functions of (state, boot-frozen scenario
// config, next tick) — no LLM, no injection door, no new RNG purpose tags,
// and NO mutable machinery state (contracts/scenario-machinery.md §1). The
// recorded events are the only latches; a restart or replay resumes exactly
// because nothing lives in memory that isn't derivable from (manifest,
// recorded events, tick).

// Incident kinds — the closed v1 vocabulary (FR-002). Growing it means a new
// constant here, a compile arm in compileIncident, and an emission arm in
// scenarioIncidentEvents; nothing outside this file may assume "schedule"
// specifically (contract §6).
const (
	// IncidentGruEmerges lands gru.emerged at an authored night-time tick and
	// position, preempting that night's random emergence roll (research R3 —
	// never two spawn mechanisms in one night).
	IncidentGruEmerges = "gru_emerges"
)

// Incident-visibility vocabulary (reorientation D4): how much of the incident
// schedule the exercise panel shows ahead of time. A VOCABULARY, never a
// boolean, in every signature and wire field (FR-009) — extensible beyond
// these two values without a shape change.
const (
	VisibilityForecast = "forecast" // the schedule is shown ahead of time
	VisibilityFog      = "fog"      // incidents are revealed only as they happen
)

// IncidentVisibilityFor resolves the effective visibility mode for an
// exercise on a world at the given manifest stage: the definition's own
// override wins; otherwise the stage-keyed default (forecast at stages 1–2
// and pre-ladder, fog from stage 3 — docs/design/tui/patterns/
// stage-defaults.md, the authority table).
func IncidentVisibilityFor(def ExerciseDefinition, stage string) string {
	if def.IncidentVisibility != "" {
		return def.IncidentVisibility
	}
	switch stage {
	case "stage-3", "stage-4":
		return VisibilityFog
	}
	return VisibilityForecast // stages 1–2 and pre-ladder ("")
}

// IncidentScheduleEntry is one authored incident on an exercise definition —
// compiled-in CONTENT, like RubricTerms, never player data (data-model.md).
// Day/Time are game time ("HH:MM"), compiled to an absolute tick at arm time
// via the existing clock arithmetic; X/Y are kind-specific parameters (the
// authored position for gru_emerges).
type IncidentScheduleEntry struct {
	Kind string
	Day  int64
	Time string
	X, Y int
}

// incident is one due emission an incident source proposes for a tick.
type incident struct {
	Kind string
	Tick int64 // the authored absolute tick (may be < nextTick after a snap)
	X, Y int
}

// incidentSource is THE director seam (spec 054 FR-002, contract §6): the
// interface producing due incidents for a tick. Its contract is fixed by the
// determinism contract (§1): implementations must be pure over (state,
// nextTick) and state-latched — the recorded event is the only "already
// fired" record, derived by observing state (e.g. a scheduled gru emergence
// is due only while its night still stands and no gru is abroad), NEVER an
// internal mutable flag, or restarts and replays desync (executor purity,
// executor.go). v1 has exactly one implementation, the compiled authored
// schedule below; the post-v1 live state-watching director is a documented
// SECOND implementation of this same interface — it attaches here and
// nowhere else, proposing incidents the reducer-valid world still disposes.
type incidentSource interface {
	incidentsDue(s *State, nextTick int64) []incident
}

// compiledIncident is one schedule entry compiled to absolute-tick
// coordinates at arm time. windowEnd is the tick at which the incident
// lapses: for gru_emerges, the dawn that ends its authored night — the
// state latch (s.Gru == nil) plus this closed window is what makes "fires
// late, never twice" pure: after an emergence the gru stays abroad until
// that same dawn, and past the dawn the entry is no longer due at all. A
// time-snap past the whole window skips the incident silently (the
// precondition-failed class, US2 AS-2 — never retried retroactively, never
// invented on replay).
type compiledIncident struct {
	incident
	windowEnd int64
}

// scheduleSource is v1's only incidentSource: the exercise definition's
// authored schedule, compiled once at arm time (boot-frozen — the SetStage
// discipline). It carries no mutable state of any kind.
type scheduleSource struct {
	entries []compiledIncident
}

func (src *scheduleSource) incidentsDue(s *State, nextTick int64) []incident {
	var due []incident
	for _, c := range src.entries {
		if nextTick < c.Tick || nextTick >= c.windowEnd {
			continue // not yet due, or lapsed with its window
		}
		switch c.Kind {
		case IncidentGruEmerges:
			// State latch: a gru already abroad means either this entry
			// already fired tonight (it holds the woods until the dawn that
			// closes this window) or the night is otherwise occupied — in
			// both readings the incident is not due. The recorded gru.emerged
			// IS the latch, observed through state.
			if s.Gru != nil {
				continue
			}
		}
		due = append(due, c.incident)
	}
	return due
}

// armedScenario is the boot-frozen scenario runtime attached to a State by
// ArmScenario: the exercise definition plus its compiled incident source.
// Unexported and never serialized (the State.m precedent) — canonical state
// bytes are unchanged by arming, and replay needs no scenario at all (the
// recorded events are the persistence).
type armedScenario struct {
	def    ExerciseDefinition
	source incidentSource
}

// ArmScenario compiles def's authored schedule and attaches the scenario
// runtime to s (spec 054 FR-006) — called exactly once, at daemon boot, from
// the manifest's Scenario block (the SetStage discipline: boot-frozen,
// never mutated afterward). A world that never calls this is byte-identical
// to pre-054 in every code path (contract §1.3). Compile errors are content
// bugs — TestScenarioSchedulesCompile pins every cataloged schedule, so a
// boot-time error here is a can't-happen belt.
func (s *State) ArmScenario(def ExerciseDefinition) error {
	src := &scheduleSource{}
	for i, e := range def.Schedule {
		c, err := compileIncident(e)
		if err != nil {
			return fmt.Errorf("scenario %s schedule[%d]: %w", def.ID, i, err)
		}
		src.entries = append(src.entries, c)
	}
	s.scenario = &armedScenario{def: def, source: src}
	return nil
}

// ScenarioExerciseID reports the armed scenario's exercise id, "" when none
// is armed — the loop's status composer reads it (FR-007).
func (s *State) ScenarioExerciseID() string {
	if s.scenario == nil {
		return ""
	}
	return s.scenario.def.ID
}

// compileIncident turns one authored entry into absolute-tick coordinates
// via the existing clock arithmetic (data-model.md "compiled to absolute
// ticks at boot").
func compileIncident(e IncidentScheduleEntry) (compiledIncident, error) {
	switch e.Kind {
	case IncidentGruEmerges:
	default:
		return compiledIncident{}, fmt.Errorf("unknown incident kind %q", e.Kind)
	}
	h, min, err := clock.ParseTimeOfDay(e.Time)
	if err != nil {
		return compiledIncident{}, err
	}
	if e.Day < 1 {
		return compiledIncident{}, fmt.Errorf("day %d out of range (1-based)", e.Day)
	}
	tick := clock.TickAt(e.Day, h, min, 0)
	if tick < 0 {
		return compiledIncident{}, fmt.Errorf("time %q on day %d precedes genesis", e.Time, e.Day)
	}
	return compiledIncident{
		incident:  incident{Kind: e.Kind, Tick: tick, X: e.X, Y: e.Y},
		windowEnd: nextDawnTick(tick),
	}, nil
}

// nextDawnTick is the first tick strictly after t whose second-of-day is the
// dawn boundary — the close of the night containing (or following) t.
func nextDawnTick(t int64) int64 {
	delta := (int64(dayStartSecond) - clock.SecondOfDay(t) + 86400) % 86400
	if delta == 0 {
		delta = 86400
	}
	return t + delta
}

// gruScheduledTonight reports whether the armed schedule has a gru_emerges
// entry landing in the night that begins at nightfallTick (research R3): on
// such a night the random emergence roll is skipped — the schedule preempts
// the dice, so exactly one spawn mechanism exists per night. Skipping the
// roll consumes no RNG draw (rngAt is coordinate-seeded, no stream), so the
// preemption is deterministic by construction: the schedule is config.
func gruScheduledTonight(s *State, nightfallTick int64) bool {
	if s.scenario == nil {
		return false
	}
	src, ok := s.scenario.source.(*scheduleSource)
	if !ok {
		return false // a future non-schedule source owns its own preemption story
	}
	dawn := nextDawnTick(nightfallTick)
	for _, c := range src.entries {
		if c.Kind == IncidentGruEmerges && c.Tick >= nightfallTick && c.Tick < dawn {
			return true
		}
	}
	return false
}

// scenarioIncidentEvents is the incident half of the executor's scenario
// consultation (spec 054 US2): ask the armed source what is due, validate
// each incident's kind-specific preconditions against the pre-tick state,
// and emit the same reducer-valid event shapes the ambient paths use. The
// schedule proposes; the reducer-valid world disposes — a failed
// precondition skips the incident silently, recorded nowhere, retried never
// (US2 AS-2). Pure over (state, map, next tick), stepEvents doctrine.
func scenarioIncidentEvents(s *State, m *worldmap.Map, nextTick int64) []store.Event {
	var events []store.Event
	for _, inc := range s.scenario.source.incidentsDue(s, nextTick) {
		switch inc.Kind {
		case IncidentGruEmerges:
			// Preconditions at emission (research R3): no gru abroad (also
			// the source's latch — belt here, matching EvaluateUnlock's
			// belt-and-suspenders posture) and an authored position the
			// random path could itself have chosen (passable, unprotected) —
			// the emitted event is indistinguishable in kind from the dice's.
			if s.Gru != nil || !passable(m, s, inc.X, inc.Y) || gruProtected(s, inc.X, inc.Y) {
				continue
			}
			events = append(events, store.Event{Tick: nextTick, Type: "gru.emerged",
				Payload: mustPayload(GruEmergedPayload{Night: gruNightIndex(nextTick), X: inc.X, Y: inc.Y})})
		}
	}
	return events
}

// RubricTerm is one evaluated rubric row: the plain-language term, the
// cataloged event type backing it, whether it is currently satisfied, and
// the backing count — the shared derivation behind both the executor's pass
// precondition and the exercise panel's live gauges (research R6: the
// gauges read the replica through this same pure function, so the panel and
// the emitter can never disagree).
type RubricTerm struct {
	Label string
	Event string
	Met   bool
	Count int
}

// EvaluateRubric derives per-term satisfaction for def from state facts the
// reducer already maintains — pure over (state, definition, tick), no log
// scan (research R4). v1 implements the first-night rubric end-to-end
// (FR-004); other exercises get their content terms rendered pending —
// the-law's charter conjunct is not state-derivable today (state retains
// only the fingerprint, not the Default flag), so its production evaluator
// is future content work, documented here rather than faked.
func EvaluateRubric(s *State, def ExerciseDefinition, tick int64) []RubricTerm {
	switch def.ID {
	case "first-night":
		return firstNightRubric(s, tick)
	}
	terms := make([]RubricTerm, 0, len(def.RubricTerms))
	for _, ev := range def.RubricTerms {
		terms = append(terms, RubricTerm{Label: ev, Event: ev})
	}
	return terms
}

// firstNightRubric is FR-004's condition set over state facts: the
// dawn-of-day-2 boundary, the zero-deaths ledger, and a player watch order
// placed before night one fell (the ratified stage-1 ceiling amendment makes
// monitor_and_act playable at stage-1). The definition's metatron.nudged
// term has no durable state trace (the nudged reducer arm only spends the
// charge), so the direction condition reads order evidence alone — exactly
// the conditions FR-004 pins.
func firstNightRubric(s *State, tick int64) []RubricTerm {
	dawn2 := clock.TickAt(2, 6, 0, 0)
	day, _, _, _ := clock.GameTime(tick)
	daysStarted := int(day - 1)
	if daysStarted < 0 {
		daysStarted = 0
	}
	watch, watchCount := firstNightWatch(s)
	return []RubricTerm{
		{Label: "village survives to dawn of day 2", Event: "sim.day_started",
			Met: tick >= dawn2 && len(s.Deaths) == 0 && !s.Ended, Count: daysStarted},
		{Label: "no villager dies", Event: "agent.died",
			Met: len(s.Deaths) == 0, Count: len(s.Deaths)},
		{Label: "a watch set before nightfall", Event: "metatron.order_placed",
			Met: watch != nil, Count: watchCount},
	}
}

// firstNightWatch finds the earliest player-origin standing order placed
// before night one fell (nightfall of day 1), plus the count of qualifying
// placements — read from State.GuardianOrders, which retains consumed orders
// (bounded at 32) well past one game day. Earliest-first (PlacedTick, then
// PlacedSeq) keeps the evidence pick deterministic.
func firstNightWatch(s *State) (*GuardianOrder, int) {
	nightfall := clock.TickAt(1, 22, 0, 0)
	var found *GuardianOrder
	count := 0
	for i := range s.GuardianOrders {
		o := &s.GuardianOrders[i]
		if o.Origin != GuardianOriginPlayer || o.PlacedTick >= nightfall {
			continue
		}
		count++
		if found == nil || o.PlacedTick < found.PlacedTick ||
			(o.PlacedTick == found.PlacedTick && o.PlacedSeq < found.PlacedSeq) {
			found = o
		}
	}
	return found, count
}

// hasCurriculumPass reports whether a pass for exercise is already on state —
// the emitter's once-only latch (research R4). The CurriculumPasses ring is
// bounded (32), but a same-exercise pass is emitted at most once per world by
// construction: this pre-emission check is consulted at every boundary tick,
// and the boundary can only recur after a pass if the ring were pruned past
// it — 32 distinct passes deep, which one-exercise-per-world (v1) never
// reaches.
func hasCurriculumPass(s *State, exercise string) bool {
	for _, p := range s.CurriculumPasses {
		if p.Exercise == exercise {
			return true
		}
	}
	return false
}

// Exercise outcome vocabulary (FR-007, data-model.md): in_progress until the
// pass lands; failed only when run.ended lands with no prior pass — failure
// is NEVER emitted as an event, run.ended IS the fail signal (contract §2).
const (
	OutcomeInProgress = "in_progress"
	OutcomePassed     = "passed"
	OutcomeFailed     = "failed"
)

// ExerciseOutcome derives the exercise's model-free outcome from replica
// facts (CurriculumPasses vs Ended) — shared by the loop's status composer
// and the exercise panel's banner, so every surface reports the same word.
func ExerciseOutcome(s *State, exercise string) string {
	switch {
	case hasCurriculumPass(s, exercise):
		return OutcomePassed
	case s.Ended:
		return OutcomeFailed
	}
	return OutcomeInProgress
}

// scenarioRubricEvents is the rubric half of the executor's scenario
// consultation (spec 054 US1, research R4): at the exercise's boundary tick,
// with every term satisfied and no rubric-violating death in THIS batch
// (the run-end detector's pure-function-of-(pre-tick-state, batch) idiom —
// same-tick deaths are not yet folded into s, and an all-dead dawn must be
// a fail, not a photo-finish pass), emit curriculum.exercise_passed and,
// when the existing unlock gate grants, curriculum.stage_unlocked — SAME
// batch, pass first (the daemon observer's contract, internal/daemon/
// curriculum.go). Once-only via the state latch (hasCurriculumPass); the
// reducer's own StagesUnlocked duplicate rejection is the door behind this
// pre-emission check. Failure emits NOTHING — run.ended is the fail signal.
func scenarioRubricEvents(s *State, nextTick int64, batch []store.Event) []store.Event {
	def := s.scenario.def
	if def.ID != "first-night" {
		return nil // v1: only first-night has a production evaluator (FR-004)
	}
	// Boundary: dawn of day 2 — the same arithmetic the sim.day_started
	// emission uses, so the pass rides the exact tick the day boundary lands.
	day, _, _, _ := clock.GameTime(nextTick)
	if clock.SecondOfDay(nextTick) != int64(dayStartSecond) || day != 2 {
		return nil
	}
	if hasCurriculumPass(s, def.ID) {
		return nil // once-only latch (state, not memory)
	}
	for _, e := range batch {
		if e.Type == "agent.died" {
			return nil // a dawn-tick death is a rubric violation (edge case: all-dead dawn)
		}
	}
	for _, term := range EvaluateRubric(s, def, nextTick) {
		if !term.Met {
			return nil
		}
	}
	// Evidence via the sanctioned constructors (curriculum.go): the watch
	// order's placement event, re-located by the reducer-stamped coordinates.
	var evidence []EvidenceRef
	if watch, _ := firstNightWatch(s); watch != nil {
		if ref, err := OrderPlacedEvidence(*watch); err == nil {
			evidence = append(evidence, ref)
		}
	}
	pass := ExercisePassedPayload{Exercise: def.ID, Stage: def.Stage, Tick: nextTick, Evidence: evidence}
	events := []store.Event{{Tick: nextTick, Type: "curriculum.exercise_passed", Payload: mustPayload(pass)}}
	if stage, ok := EvaluateUnlock(s, pass); ok {
		events = append(events, store.Event{Tick: nextTick, Type: "curriculum.stage_unlocked",
			Payload: mustPayload(StageUnlockedPayload{Stage: stage, Exercise: def.ID, Tick: nextTick})})
	}
	return events
}
