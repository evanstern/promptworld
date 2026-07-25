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
// whitelist entries exist for them. The production emitter is TASK-119's
// scenario rubric machinery; until it lands, only test fixtures emit them —
// this file ships the payload contracts and the reducer arms.

const (
	// curriculumPassRetain bounds retained pass records (the metatronOrderRetain
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
	// Custom==true means a player-authored charter revision (a fingerprint ≠
	// default) was in force at pass time; for a stage-3 pass, any evidence
	// entry with Custom==true names a player-granted tool's contributing act
	// (as opposed to a tool granted by the default/stage manifest alone).
	// Absent/false = a default-fact evidence entry, which never satisfies a
	// gate conjunct (SC-004's negative case).
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
// rather than clamps (the metatron arm's contract): the executor emits these
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
//	           "metatron.charter_observed" AND Custom == true (SC-004: a
//	           default-charter pass must NOT satisfy this).
//	stage-3 -> stage-4: the pass's evidence must include a player-granted
//	           tool's contributing act — any evidence entry with Custom ==
//	           true (a fixed event type isn't pinned by the contract: which
//	           tool contributed is TASK-119's exercise design).
//
// RECONCILIATION NOTE (T022, for the 044 US2 rebase): 044 (task-31, in
// flight at 046 plan time) is landing the REAL
// metatron.charter_observed fingerprint event. This function references it
// by the event TYPE STRING and a Custom/default flag on EvidenceRef ONLY —
// it does not import or depend on 044's payload type — so once 044 merges,
// the only reconciliation needed is confirming the real event's fingerprint
// semantics map onto Custom (true iff the observed charter differs from the
// world's default/preset) and, if 044 exposes a richer payload, optionally
// reading it directly instead of trusting the evidence flag. No logic here
// should need to change.
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
