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
// a pass — a rubric term's satisfying event, or (stage-2 gate) the
// metatron.charter_observed fingerprint in force at pass time (spec 044). The
// (type, seq, tick) triple is enough to re-locate the event in the log, so the
// claim stays independently auditable (FR-008).
type EvidenceRef struct {
	Type string `json:"type"`
	Seq  int64  `json:"seq"`
	Tick int64  `json:"tick"`
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
