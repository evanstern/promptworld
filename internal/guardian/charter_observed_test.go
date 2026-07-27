package guardian

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// charterObservations extracts every landed metatron.charter_observed from
// the injector's batches, in landing order.
func charterObservations(t *testing.T, inj *stateInjector) []sim.CharterObservedPayload {
	t.Helper()
	inj.mu.Lock()
	defer inj.mu.Unlock()
	var out []sim.CharterObservedPayload
	for _, batch := range inj.batches {
		for _, e := range batch {
			if e.Type != "metatron.charter_observed" {
				continue
			}
			var p sim.CharterObservedPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				t.Fatal(err)
			}
			out = append(out, p)
		}
	}
	return out
}

// TestCharterObservationEmitted (spec 044 US2, T014 / FR-008): the first turn
// records the effective charter's fingerprint; an unchanged charter emits
// nothing further; an edit is observed on the NEXT turn with a new
// fingerprint and default=false — fingerprint-at-effect semantics.
func TestCharterObservationEmitted(t *testing.T) {
	mt, _, inj, dir := newTestGuardian(t, "I am here.")
	mt.charterFP = "" // undo the fixture pre-seed: this test IS the observation

	// Turn 1: first observation — the default charter, by its content hash.
	if _, err := mt.Turn(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	obs := charterObservations(t, inj)
	if len(obs) != 1 {
		t.Fatalf("after turn 1: %d observations, want 1", len(obs))
	}
	if !obs[0].Default || obs[0].Fingerprint != charterFingerprint(persona.DefaultCharter) {
		t.Errorf("observation = %+v, want the default charter's fingerprint with default=true", obs[0])
	}
	if inj.state.CharterFingerprint != obs[0].Fingerprint {
		t.Errorf("State.CharterFingerprint = %q, want %q", inj.state.CharterFingerprint, obs[0].Fingerprint)
	}

	// Turn 2: unchanged charter — no re-emission.
	if _, err := mt.Turn(context.Background(), "hello again"); err != nil {
		t.Fatal(err)
	}
	if obs = charterObservations(t, inj); len(obs) != 1 {
		t.Fatalf("after an unchanged turn 2: %d observations, want still 1", len(obs))
	}

	// Player edits the charter; turn 3 observes the revision.
	if err := os.WriteFile(filepath.Join(dir, "charter.md"), []byte("Be bold. Guard the fire."), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := mt.Turn(context.Background(), "what changed?"); err != nil {
		t.Fatal(err)
	}
	obs = charterObservations(t, inj)
	if len(obs) != 2 {
		t.Fatalf("after the edit: %d observations, want 2", len(obs))
	}
	if obs[1].Default || obs[1].Fingerprint == obs[0].Fingerprint {
		t.Errorf("edited observation = %+v, want a new fingerprint with default=false", obs[1])
	}
	if obs[1].Fingerprint != charterFingerprint("Be bold. Guard the fire.") {
		t.Errorf("fingerprint = %q, want the hash of the effective edited text", obs[1].Fingerprint)
	}
}

// TestCharterObservationTutorPresetIsDefault (spec 046 T022 reconciliation,
// step 4a — evidence honesty across 044/046): a stage-1 tutor-preset world's
// effective charter is the preset constant the stage lock serves
// (persona.TutorCharter) — the GAME's authorship, not the player's — so the
// observation must fingerprint that effective text and record default=true.
// The morgue's evidence timeline and the stage-2→3 unlock gate
// (sim.CharterObservedEvidence derives Custom = !default) both read this flag:
// preset text masquerading as player-authored would let a tutor world satisfy
// SC-004's conjunct without the player ever writing a charter. Once the world
// reaches an authoring stage and the player actually revises charter.md, the
// next observation records default=false — that contrast is the point.
func TestCharterObservationTutorPresetIsDefault(t *testing.T) {
	mt, _, inj, dir := newTestGuardian(t, "Watch with me.")
	mt.SetStage("stage-1", "tutor")
	mt.charterFP = ""

	// Stage-1: charter.md holds the genesis default text, but the lock serves
	// the tutor preset — the observation records the preset as default=true.
	if _, err := mt.Turn(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	obs := charterObservations(t, inj)
	if len(obs) != 1 {
		t.Fatalf("after turn 1: %d observations, want 1", len(obs))
	}
	if obs[0].Fingerprint != charterFingerprint(persona.TutorCharter) {
		t.Errorf("fingerprint = %q, want the tutor preset's (the stage-effective text, not the file's)", obs[0].Fingerprint)
	}
	if !obs[0].Default {
		t.Errorf("observation = %+v: a stage-1 tutor-preset charter is the game's authorship and must record default=true", obs[0])
	}

	// Stage-2 (authoring unlocked), same tutor preset: an untouched tutor
	// charter is still default; a player revision is observed as authored.
	mt.SetStage("stage-2", "tutor")
	if err := os.WriteFile(filepath.Join(dir, "charter.md"), []byte(persona.TutorCharter), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := mt.Turn(context.Background(), "still watching?"); err != nil {
		t.Fatal(err)
	}
	if obs = charterObservations(t, inj); len(obs) != 1 {
		t.Fatalf("unchanged tutor text at stage-2 re-emitted: %d observations, want still 1", len(obs))
	}
	if err := os.WriteFile(filepath.Join(dir, "charter.md"), []byte("Guard the fire above all."), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := mt.Turn(context.Background(), "new law"); err != nil {
		t.Fatal(err)
	}
	obs = charterObservations(t, inj)
	if len(obs) != 2 {
		t.Fatalf("after the player's revision: %d observations, want 2", len(obs))
	}
	if obs[1].Default {
		t.Errorf("observation = %+v: a player-authored revision must record default=false", obs[1])
	}
}

// TestCharterObservationEndedStageOneCoexists (T022 reconciliation, step 4b):
// spec 044's run-end gate and spec 046's stage gate ride the same turn path —
// on an ENDED stage-1 tutor-preset world the stage lock still composes the
// preset charter for the turn, and the closed evidence timeline still wins:
// no observation is emitted. Neither gate disturbs the other.
func TestCharterObservationEndedStageOneCoexists(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "The watch is over.")
	mt.SetStage("stage-1", "tutor")
	mt.charterFP = ""
	mt.replica.Apply(store.Event{Tick: 1, Type: "run.ended", Payload: mustJSON(sim.RunEndedPayload{
		Tick: 1, Deaths: sim.DeathRefs([]sim.DeathRecord{{Agent: 0, Tick: 1, Cause: "starvation"}}), FinalCause: "starvation"})})
	mt.mirrorState()

	if _, err := mt.Turn(context.Background(), "what remains?"); err != nil {
		t.Fatal(err)
	}
	if obs := charterObservations(t, inj); len(obs) != 0 {
		t.Errorf("ended stage-1 world emitted %d observations, want 0", len(obs))
	}
}

// TestCharterObservationSkippedWhenEnded (spec 044): an ended world's
// evidence timeline is closed — a turn on it emits no observation (the
// narrowed door would refuse it anyway; the skip keeps the log quiet).
func TestCharterObservationSkippedWhenEnded(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "It is over.")
	mt.charterFP = ""
	mt.replica.Apply(store.Event{Tick: 1, Type: "run.ended", Payload: mustJSON(sim.RunEndedPayload{
		Tick: 1, Deaths: sim.DeathRefs([]sim.DeathRecord{{Agent: 0, Tick: 1, Cause: "starvation"}}), FinalCause: "starvation"})})
	mt.mirrorState()

	if _, err := mt.Turn(context.Background(), "what happened?"); err != nil {
		t.Fatal(err)
	}
	if obs := charterObservations(t, inj); len(obs) != 0 {
		t.Errorf("ended world emitted %d observations, want 0", len(obs))
	}
}
