package metatron

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
	mt, _, inj, dir := newTestAngel(t, "I am here.")
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

// TestCharterObservationSkippedWhenEnded (spec 044): an ended world's
// evidence timeline is closed — a turn on it emits no observation (the
// narrowed door would refuse it anyway; the skip keeps the log quiet).
func TestCharterObservationSkippedWhenEnded(t *testing.T) {
	mt, _, inj, _ := newTestAngel(t, "It is over.")
	mt.charterFP = ""
	mt.replica.Apply(store.Event{Tick: 1, Type: "run.ended", Payload: mustJSON(sim.RunEndedPayload{
		Tick: 1, Deaths: []sim.DeathRecord{{Agent: 0, Tick: 1, Cause: "starvation"}}, FinalCause: "starvation"})})
	mt.mirrorState()

	if _, err := mt.Turn(context.Background(), "what happened?"); err != nil {
		t.Fatal(err)
	}
	if obs := charterObservations(t, inj); len(obs) != 0 {
		t.Errorf("ended world emitted %d observations, want 0", len(obs))
	}
}
