package daemon

// Spec 046 US3 T015/T019: the daemon-side curriculum observer, tested
// directly against a world fixture and an isolated unlocks-record home —
// no daemon boot, no LLM orchestrator required (FR-014: this machinery is
// model-free by construction; nothing here touches llm.json).

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worlds"
)

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestCurriculumObserverUpsertsUnlockWithEvidence (spec 046 T013/T015): on
// observing curriculum.stage_unlocked in the SAME batch as the
// curriculum.exercise_passed that proved it, the observer upserts the
// per-user unlocks record with a pointer at this world and the pass event's
// (type, seq, tick) — no orchestrator, no daemon boot required.
func TestCurriculumObserverUpsertsUnlockWithEvidence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("PROMPTWORLD_HOME", home)

	dir := t.TempDir()
	w, err := world.Create(dir, "proofworld", 1)
	if err != nil {
		t.Fatal(err)
	}

	obs := curriculumObserver(w)
	pass := sim.ExercisePassedPayload{Exercise: "first-night", Stage: "stage-1", Tick: 86400}
	unlock := sim.StageUnlockedPayload{Stage: "stage-2", Exercise: "first-night", Tick: 86400}
	batch := []store.Event{
		{Seq: 41, Tick: 86400, Type: "curriculum.exercise_passed", Payload: mustJSON(t, pass)},
		{Seq: 42, Tick: 86400, Type: "curriculum.stage_unlocked", Payload: mustJSON(t, unlock)},
	}
	obs(batch)

	u := worlds.LoadUnlocks()
	if !u.Earned("stage-2") {
		t.Fatal("expected stage-2 recorded as earned")
	}
	entry := u.Entries["stage-2"]
	if entry.World != "proofworld" || entry.Path != dir || entry.Exercise != "first-night" {
		t.Errorf("unlock entry = %+v", entry)
	}
	if len(entry.Evidence) != 1 || entry.Evidence[0].Type != "curriculum.exercise_passed" ||
		entry.Evidence[0].Seq != 41 || entry.Evidence[0].Tick != 86400 {
		t.Errorf("unlock evidence = %+v, want a pointer at the pass event (seq 41)", entry.Evidence)
	}
	if entry.EarnedAt == "" {
		t.Error("expected a non-empty earned_at timestamp")
	}
}

// TestCurriculumObserverIgnoresOtherEvents: a batch with no
// curriculum.stage_unlocked is a complete no-op — no write, no touch of the
// unlocks record (advisory doctrine: only observes what it's told to).
func TestCurriculumObserverIgnoresOtherEvents(t *testing.T) {
	home := t.TempDir()
	t.Setenv("PROMPTWORLD_HOME", home)

	dir := t.TempDir()
	w, err := world.Create(dir, "quietworld", 1)
	if err != nil {
		t.Fatal(err)
	}
	obs := curriculumObserver(w)
	obs([]store.Event{
		{Seq: 1, Tick: 100, Type: "curriculum.exercise_passed",
			Payload: mustJSON(t, sim.ExercisePassedPayload{Exercise: "first-night", Stage: "stage-1", Tick: 100})},
		{Seq: 2, Tick: 200, Type: "sim.day_started", Payload: mustJSON(t, sim.DayPayload{Day: 2})},
	})
	u := worlds.LoadUnlocks()
	if len(u.Entries) != 0 {
		t.Errorf("expected no unlock recorded without a stage_unlocked event, got %v", u.Entries)
	}
}

// TestCurriculumObserverToleratesMissingPassInBatch: a stage_unlocked event
// with no matching exercise_passed in the same batch still upserts an entry
// (degraded — no evidence pointer) rather than dropping the unlock or
// panicking; advisory-never-authority means the world's own history remains
// the real proof regardless of what this convenience record captured.
func TestCurriculumObserverToleratesMissingPassInBatch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("PROMPTWORLD_HOME", home)

	dir := t.TempDir()
	w, err := world.Create(dir, "solow", 1)
	if err != nil {
		t.Fatal(err)
	}
	obs := curriculumObserver(w)
	unlock := sim.StageUnlockedPayload{Stage: "stage-2", Exercise: "first-night", Tick: 86400}
	obs([]store.Event{{Seq: 9, Tick: 86400, Type: "curriculum.stage_unlocked", Payload: mustJSON(t, unlock)}})

	u := worlds.LoadUnlocks()
	if !u.Earned("stage-2") {
		t.Fatal("expected stage-2 recorded even with no evidence-bearing pass in the batch")
	}
	if len(u.Entries["stage-2"].Evidence) != 0 {
		t.Errorf("expected no evidence pointer without a matching pass, got %v", u.Entries["stage-2"].Evidence)
	}
}

// TestCurriculumObserverPathMatchesWorldFixture: confirms the recorded Path
// is exactly the fixture world's Dir, so `promptworld stages`'s audit
// pointer resolves back to this exact directory (contract rule 5).
func TestCurriculumObserverPathMatchesWorldFixture(t *testing.T) {
	home := t.TempDir()
	t.Setenv("PROMPTWORLD_HOME", home)

	dir := filepath.Join(t.TempDir(), "namedworld")
	w, err := world.Create(dir, "namedworld", 1)
	if err != nil {
		t.Fatal(err)
	}
	obs := curriculumObserver(w)
	obs([]store.Event{{Seq: 1, Tick: 1, Type: "curriculum.stage_unlocked",
		Payload: mustJSON(t, sim.StageUnlockedPayload{Stage: "stage-2", Exercise: "first-night", Tick: 1})}})

	u := worlds.LoadUnlocks()
	if got := u.Entries["stage-2"].Path; got != dir {
		t.Errorf("recorded path = %q, want %q", got, dir)
	}
}
