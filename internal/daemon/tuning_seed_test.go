package daemon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
)

// TestSeedTuningAgainstGenesisPin is spec 057 / TASK-108 US2 (AC3, FR-003): the
// spec-048 boot seed keeps its exact semantics against a genesis-pinned world —
// no code change was needed, and this proves it. A post-057 world boots with
// state whose Tuning already carries the pinned default set (the genesis
// sim.tuning_applied event); seedTuning must:
//   - append NOTHING when no tuning.json exists,
//   - append NOTHING when tuning.json resolves to the same effective set,
//   - append exactly ONE event when tuning.json differs.
//
// LastSeq (the store's monotonic append counter) is the witness: it moves by
// exactly the number of events seedTuning appended.
func TestSeedTuningAgainstGenesisPin(t *testing.T) {
	// newPinnedWorld builds a post-057 world: manifest + store + recovered state
	// carrying the genesis default pin, exactly as `promptworld new` seeds it.
	newPinnedWorld := func(t *testing.T) (*world.World, *store.Store, *sim.State) {
		t.Helper()
		dir := t.TempDir()
		manifest := `{"name":"w","seed":42,"format_version":5,"tick_game_seconds":1}`
		if err := os.WriteFile(filepath.Join(dir, world.ManifestName), []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
		w, err := world.Open(dir)
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		st, err := store.Open(w.DBPath())
		if err != nil {
			t.Fatal(err)
		}
		// Seed the genesis log (world.created + the tuning pin), then recover
		// state from it — the pinned world a post-057 daemon boots into.
		pin := sim.GenesisTuningEvent(0)
		if err := st.AppendEvents([]store.Event{{Tick: 0, Type: "world.created"}, pin}); err != nil {
			t.Fatal(err)
		}
		state, err := recoverState(w, st)
		if err != nil {
			t.Fatalf("recoverState: %v", err)
		}
		if state.Tuning == nil {
			t.Fatal("genesis pin did not establish state.Tuning on recovery")
		}
		if got := state.RefuelDyingBelow(); got != 10800 {
			t.Fatalf("pinned RefuelDyingBelow() = %d, want the 10800 default", got)
		}
		return w, st, state
	}

	writeTuning := func(t *testing.T, w *world.World, body string) {
		t.Helper()
		if err := os.WriteFile(w.TuningPath(), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("absent tuning.json seeds nothing", func(t *testing.T) {
		w, st, state := newPinnedWorld(t)
		defer st.Close()
		before := st.LastSeq()
		if err := seedTuning(w, st, state); err != nil {
			t.Fatalf("seedTuning: %v", err)
		}
		if st.LastSeq() != before {
			t.Errorf("absent file appended an event (seq %d → %d)", before, st.LastSeq())
		}
	})

	t.Run("tuning.json equal to the pinned defaults seeds nothing", func(t *testing.T) {
		w, st, state := newPinnedWorld(t)
		defer st.Close()
		// An empty object resolves to the full default set — equal to the pin.
		writeTuning(t, w, `{}`)
		before := st.LastSeq()
		if err := seedTuning(w, st, state); err != nil {
			t.Fatalf("seedTuning: %v", err)
		}
		if st.LastSeq() != before {
			t.Errorf("a defaults-equal manifest appended an event (seq %d → %d)", before, st.LastSeq())
		}
	})

	t.Run("tuning.json differing from the pin seeds exactly one event", func(t *testing.T) {
		w, st, state := newPinnedWorld(t)
		defer st.Close()
		// Pin the OLD refuel default (3600): differs from the 10800 genesis pin.
		writeTuning(t, w, `{"refuel_dying_below":3600}`)
		before := st.LastSeq()
		if err := seedTuning(w, st, state); err != nil {
			t.Fatalf("seedTuning: %v", err)
		}
		if st.LastSeq() != before+1 {
			t.Errorf("a differing manifest appended %d events, want exactly 1 (seq %d → %d)", st.LastSeq()-before, before, st.LastSeq())
		}
		if got := state.RefuelDyingBelow(); got != 3600 {
			t.Errorf("after applying the manifest, RefuelDyingBelow() = %d, want 3600 (the manifest value wins)", got)
		}
	})
}
