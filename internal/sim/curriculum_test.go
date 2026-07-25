package sim

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// Fixture emission (spec 046 R5): the production emitter of curriculum.* is
// TASK-119's rubric machinery; these tests emit fixtures to prove the reducer
// contract — payloads apply, records bound, replays reproduce.

func curriculumEvent(t *testing.T, typ string, tick int64, payload any) store.Event {
	t.Helper()
	return store.Event{Tick: tick, Type: typ, Payload: mustPayload(payload)}
}

func passPayload(exercise, stage string, tick int64) ExercisePassedPayload {
	return ExercisePassedPayload{Exercise: exercise, Stage: stage, Tick: tick,
		Evidence: []EvidenceRef{{Type: "metatron.nudged", Seq: 41, Tick: tick - 5}}}
}

// TestExercisePassedRecordsPass (spec 046 T004): a pass event lands on state
// with its evidence pointers intact, and survives a snapshot round trip
// (omitempty additive fields, no format bump).
func TestExercisePassedRecordsPass(t *testing.T) {
	s := NewState(1, testMap(1))
	e := curriculumEvent(t, "curriculum.exercise_passed", 100, passPayload("first-night", "stage-1", 100))
	if err := s.Apply(e); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(s.CurriculumPasses) != 1 {
		t.Fatalf("passes = %d, want 1", len(s.CurriculumPasses))
	}
	p := s.CurriculumPasses[0]
	if p.Exercise != "first-night" || p.Stage != "stage-1" || p.Tick != 100 {
		t.Errorf("recorded pass = %+v", p)
	}
	if len(p.Evidence) != 1 || p.Evidence[0].Seq != 41 || p.Evidence[0].Type != "metatron.nudged" {
		t.Errorf("evidence pointers lost: %+v", p.Evidence)
	}

	// Snapshot round trip: the recorded pass reconstructs from bytes alone.
	restored := restoreState(t, s.Marshal())
	if len(restored.CurriculumPasses) != 1 {
		t.Fatalf("pass did not round-trip: %+v", restored.CurriculumPasses)
	}
	rp := restored.CurriculumPasses[0]
	if rp.Exercise != p.Exercise || rp.Stage != p.Stage || rp.Tick != p.Tick ||
		len(rp.Evidence) != 1 || rp.Evidence[0] != p.Evidence[0] {
		t.Errorf("restored pass = %+v, want %+v", rp, p)
	}
}

func restoreState(t *testing.T, b []byte) *State {
	t.Helper()
	var s State
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatalf("restore: %v", err)
	}
	return &s
}

// TestExercisePassedValidates (spec 046 T004): malformed fixtures are rejected
// at the door — empty exercise, unknown stage.
func TestExercisePassedValidates(t *testing.T) {
	s := NewState(1, testMap(1))
	cases := []ExercisePassedPayload{
		{Exercise: "", Stage: "stage-1", Tick: 1},
		{Exercise: "first-night", Stage: "stage-5", Tick: 1},
		{Exercise: "first-night", Stage: "", Tick: 1},
	}
	for _, p := range cases {
		if err := s.Apply(curriculumEvent(t, "curriculum.exercise_passed", 1, p)); err == nil {
			t.Errorf("payload %+v should be rejected", p)
		}
	}
	if len(s.CurriculumPasses) != 0 {
		t.Errorf("rejected events must not mutate state: %+v", s.CurriculumPasses)
	}
}

// TestCurriculumPassesBounded (spec 046 T004): the pass record is a bounded
// ring — only the most recent curriculumPassRetain (32) passes are retained,
// oldest dropped first, deterministically.
func TestCurriculumPassesBounded(t *testing.T) {
	s := NewState(1, testMap(1))
	for i := 0; i < curriculumPassRetain+8; i++ {
		p := passPayload("first-night", "stage-1", int64(i))
		if err := s.Apply(curriculumEvent(t, "curriculum.exercise_passed", int64(i), p)); err != nil {
			t.Fatalf("apply %d: %v", i, err)
		}
	}
	if len(s.CurriculumPasses) != curriculumPassRetain {
		t.Fatalf("passes = %d, want the %d cap", len(s.CurriculumPasses), curriculumPassRetain)
	}
	if got := s.CurriculumPasses[0].Tick; got != 8 {
		t.Errorf("oldest retained pass tick = %d, want 8 (oldest dropped first)", got)
	}
}

// TestStageUnlockedLatches (spec 046 T004): an unlock latches the per-world
// stage fact; a duplicate latch for the same stage is rejected (once per
// (world, stage)); stage-1 and unknown stages are never unlockable.
func TestStageUnlockedLatches(t *testing.T) {
	s := NewState(1, testMap(1))
	unlock := StageUnlockedPayload{Stage: "stage-2", Exercise: "first-night", Tick: 200}
	if err := s.Apply(curriculumEvent(t, "curriculum.stage_unlocked", 200, unlock)); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(s.StagesUnlocked) != 1 || s.StagesUnlocked[0] != "stage-2" {
		t.Fatalf("StagesUnlocked = %v, want [stage-2]", s.StagesUnlocked)
	}

	if err := s.Apply(curriculumEvent(t, "curriculum.stage_unlocked", 300, unlock)); err == nil {
		t.Error("duplicate stage-2 unlock should be rejected")
	} else if !strings.Contains(err.Error(), "already unlocked") {
		t.Errorf("duplicate refusal should say so, got: %v", err)
	}

	for _, bad := range []StageUnlockedPayload{
		{Stage: "stage-1", Exercise: "x", Tick: 1},
		{Stage: "stage-7", Exercise: "x", Tick: 1},
		{Stage: "stage-3", Exercise: "", Tick: 1},
	} {
		if err := s.Apply(curriculumEvent(t, "curriculum.stage_unlocked", 1, bad)); err == nil {
			t.Errorf("payload %+v should be rejected", bad)
		}
	}
	if len(s.StagesUnlocked) != 1 {
		t.Errorf("rejected unlocks must not latch: %v", s.StagesUnlocked)
	}
}

// TestCurriculumReplayDeterministic (spec 046 contracts/events.md): replaying
// a history carrying curriculum events reproduces identical state — the
// events are replay-deterministic like every other recorded type.
func TestCurriculumReplayDeterministic(t *testing.T) {
	m := testMap(3)
	timeline := map[int64][]store.Event{
		1000: {curriculumEvent(t, "curriculum.exercise_passed", 1000, passPayload("first-night", "stage-1", 1000))},
		1500: {curriculumEvent(t, "curriculum.stage_unlocked", 1500,
			StageUnlockedPayload{Stage: "stage-2", Exercise: "first-night", Tick: 1500})},
	}
	a, b := NewState(3, m), NewState(3, m)
	driveTicks(t, a, m, 3_000, timeline)
	driveTicks(t, b, m, 3_000, timeline)
	if a.Hash() != b.Hash() {
		t.Fatal("curriculum events broke same-seed determinism")
	}
	if len(a.CurriculumPasses) != 1 || len(a.StagesUnlocked) != 1 {
		t.Fatalf("timeline events not applied: passes=%d unlocks=%v",
			len(a.CurriculumPasses), a.StagesUnlocked)
	}
}
