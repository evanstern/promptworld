package e2e

// Fork e2e (spec 076 T012 — SC-001 + FR-010c): a run-and-stopped world
// forks at its shutdown snapshot; parent and fork start SIDE BY SIDE, both
// answer status and appear running in ps; and the fork — having then RUN on
// its own — passes the determinism harness independently: its full log
// replays from genesis to its own final snapshot's state_hash, not by
// inheritance from the parent.

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
)

func TestForkE2E_SideBySideAndIndependentDeterminism(t *testing.T) {
	isolatedHome(t)

	// A pure-sim parent, run past the first snapshot cadence and stopped
	// gracefully (the stop cuts the covering shutdown snapshot).
	parentDir := filepath.Join(t.TempDir(), "duel-a")
	run(t, "new", parentDir, "--name", "duel-a", "--seed", "4242")
	os.Remove(filepath.Join(parentDir, "llm.json"))
	run(t, "start", parentDir)
	run(t, "speed", parentDir, "max")
	waitTick(t, parentDir, 4000) // past SnapshotEveryTicks (3600)
	run(t, "stop", parentDir)

	// Fork it (path form; name from the basename).
	forkDir := filepath.Join(t.TempDir(), "duel-b")
	out := run(t, "fork", parentDir, forkDir)
	if !strings.Contains(out, `forked "duel-a" → "duel-b"`) {
		t.Fatalf("fork summary = %q", out)
	}

	// SC-001: both worlds run simultaneously, each on its own daemon.
	run(t, "start", parentDir)
	t.Cleanup(func() { stopHard(parentDir) })
	run(t, "start", forkDir)
	t.Cleanup(func() { stopHard(forkDir) })
	ps := status(t, parentDir)
	fs := status(t, forkDir)
	if ps.World.Name != "duel-a" || fs.World.Name != "duel-b" {
		t.Errorf("status names = %q / %q, want duel-a / duel-b", ps.World.Name, fs.World.Name)
	}
	if ps.World.Seed != 4242 || fs.World.Seed != 4242 {
		t.Errorf("seeds = %d / %d, want 4242 for both (seed is carried)", ps.World.Seed, fs.World.Seed)
	}
	running := map[string]bool{}
	for _, row := range psJSON(t) {
		if row.State == "running" {
			running[row.Name] = true
		}
	}
	if !running["duel-a"] || !running["duel-b"] {
		t.Errorf("ps --json running set = %v, want both duel-a and duel-b", running)
	}

	// FR-010(c): let the FORK run on past its boundary at max, then stop it
	// and prove its own log replays from genesis to its own final snapshot.
	run(t, "speed", forkDir, "max")
	forkBoundary := fs.Clock.Tick
	waitTick(t, forkDir, forkBoundary+3000)
	run(t, "stop", forkDir)
	run(t, "stop", parentDir)

	w, err := world.Open(forkDir)
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	snap, err := st.LatestValidSnapshot()
	if err != nil || snap == nil {
		t.Fatalf("fork final snapshot: %v, %v", snap, err)
	}
	replayed := sim.NewState(w.Manifest.Seed, w.Map())
	if err := st.ReplayEvents(0, func(e store.Event) error {
		if e.Seq > snap.Seq {
			return nil // the covering snapshot's window; later bookkeeping (daemon.stopped) is past it
		}
		return replayed.Apply(e)
	}); err != nil {
		t.Fatal(err)
	}
	if replayed.Tick < snap.Tick {
		replayed.Tick = snap.Tick // quiet ticks advance the clock without rows (recovery's rule)
	}
	sum := sha256.Sum256(replayed.Marshal())
	if got := hex.EncodeToString(sum[:]); got != snap.Hash {
		t.Errorf("fork's genesis replay hashes %s, want its own final snapshot's %s — the fork must pass the determinism harness independently", got, snap.Hash)
	}

	// The fork's log carries exactly one world.forked, and it survives into
	// the run history (SC-002's e2e face).
	events, err := st.EventsSince(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	forked := 0
	for _, e := range events {
		if e.Type == "world.forked" {
			forked++
		}
	}
	if forked != 1 {
		t.Errorf("fork log carries %d world.forked events, want exactly 1", forked)
	}
}
