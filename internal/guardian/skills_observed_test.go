package guardian

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// Skills observation tests (spec 077 FR-006): the observeCharter twin —
// emitted at turn time on fingerprint change of the BOUND set, never for an
// empty set, structurally silent at stages 1–2 (stageSkills refuses to bind).

// skillsObservations extracts every landed metatron.skills_observed from the
// injector's batches, in landing order (the charterObservations shape).
func skillsObservations(t *testing.T, inj *stateInjector) []sim.SkillsObservedPayload {
	t.Helper()
	inj.mu.Lock()
	defer inj.mu.Unlock()
	var out []sim.SkillsObservedPayload
	for _, batch := range inj.batches {
		for _, e := range batch {
			if e.Type != "metatron.skills_observed" {
				continue
			}
			var p sim.SkillsObservedPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				t.Fatal(err)
			}
			out = append(out, p)
		}
	}
	return out
}

// TestSkillsObservationEmitted: a bound skill set is observed on the first
// turn it runs under (fingerprint + composition-ordered names), an unchanged
// set emits nothing further, and an edit is observed on the next turn with a
// new fingerprint — fingerprint-at-effect, the charter observation's
// semantics exactly.
func TestSkillsObservationEmitted(t *testing.T) {
	mt, _, inj, dir := newTestGuardian(t, "I am here.")
	writeSkill(t, dir, "10-watch.md", "Keep the night watch doctrine.")
	writeSkill(t, dir, "20-tone.md", "Speak plainly.")

	// Turn 1 (pre-ladder stage "": skills bind): first observation.
	if _, err := mt.Turn(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	obs := skillsObservations(t, inj)
	if len(obs) != 1 {
		t.Fatalf("after turn 1: %d observations, want 1", len(obs))
	}
	if len(obs[0].Names) != 2 || obs[0].Names[0] != "10-watch.md" || obs[0].Names[1] != "20-tone.md" {
		t.Errorf("names = %v, want composition order [10-watch.md 20-tone.md]", obs[0].Names)
	}
	skills, _ := loadSkills(dir)
	if obs[0].Fingerprint != skillsFingerprint(skills) {
		t.Errorf("fingerprint = %q, want the composed set's hash", obs[0].Fingerprint)
	}
	// (Seq stamping is the real loop's stampSeqs contract — the sim-side arm
	// test covers it; this fixture's injector applies with Seq 0.)
	if inj.state.SkillsFingerprint != obs[0].Fingerprint {
		t.Errorf("State.SkillsFingerprint = %q, want %q", inj.state.SkillsFingerprint, obs[0].Fingerprint)
	}

	// Turn 2: unchanged set — no re-emission.
	if _, err := mt.Turn(context.Background(), "again"); err != nil {
		t.Fatal(err)
	}
	if obs = skillsObservations(t, inj); len(obs) != 1 {
		t.Fatalf("unchanged set re-emitted: %d observations, want still 1", len(obs))
	}

	// Player edits a skill file; turn 3 observes the revision.
	writeSkill(t, dir, "10-watch.md", "Keep the DAY watch too.")
	if _, err := mt.Turn(context.Background(), "what changed?"); err != nil {
		t.Fatal(err)
	}
	obs = skillsObservations(t, inj)
	if len(obs) != 2 {
		t.Fatalf("after the edit: %d observations, want 2", len(obs))
	}
	if obs[1].Fingerprint == obs[0].Fingerprint {
		t.Error("edited set re-recorded the old fingerprint")
	}
}

// TestSkillsObservationEmptySetSilent: a world with no skills/ folder — or
// with files present but a stage that refuses to bind them (stages 1–2) —
// never emits: absence is not an observation, and the structural silence at
// stages 1–2 is what makes Custom-by-construction honest.
func TestSkillsObservationEmptySetSilent(t *testing.T) {
	// No skills folder at all.
	mt, _, inj, _ := newTestGuardian(t, "quiet")
	if _, err := mt.Turn(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	if obs := skillsObservations(t, inj); len(obs) != 0 {
		t.Fatalf("skill-less world emitted %d observations, want 0", len(obs))
	}

	// Files present, stage-2 (skills do not bind): still silent — the lock
	// notice is the player's signal, never a recorded observation.
	mt2, _, inj2, dir2 := newTestGuardian(t, "locked")
	mt2.SetStage("stage-2", "")
	writeSkill(t, dir2, "10-watch.md", "Keep the night watch doctrine.")
	if _, err := mt2.Turn(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	if obs := skillsObservations(t, inj2); len(obs) != 0 {
		t.Fatalf("stage-2 world emitted %d observations, want 0 (skills do not bind)", len(obs))
	}

	// Stage-3: the same files bind and are observed.
	mt3, _, inj3, dir3 := newTestGuardian(t, "unlocked")
	mt3.SetStage("stage-3", "")
	writeSkill(t, dir3, "10-watch.md", "Keep the night watch doctrine.")
	if _, err := mt3.Turn(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	if obs := skillsObservations(t, inj3); len(obs) != 1 {
		t.Fatalf("stage-3 world emitted %d observations, want 1", len(obs))
	}
}

// TestSkillsObservationSkippedWhenEnded: an ended world's evidence timeline
// is closed — the charter observation's gate, shared.
func TestSkillsObservationSkippedWhenEnded(t *testing.T) {
	mt, _, inj, dir := newTestGuardian(t, "over")
	writeSkill(t, dir, "10-watch.md", "Keep the night watch doctrine.")
	mt.replica.Apply(store.Event{Tick: 1, Type: "run.ended", Payload: mustJSON(sim.RunEndedPayload{
		Tick: 1, Deaths: sim.DeathRefs([]sim.DeathRecord{{Agent: 0, Tick: 1, Cause: "starvation"}}), FinalCause: "starvation"})})
	mt.mirrorState()
	if _, err := mt.Turn(context.Background(), "what remains?"); err != nil {
		t.Fatal(err)
	}
	if obs := skillsObservations(t, inj); len(obs) != 0 {
		t.Fatalf("ended world emitted %d observations, want 0", len(obs))
	}
}
