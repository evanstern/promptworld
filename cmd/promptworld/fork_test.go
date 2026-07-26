package main

// cmdFork tests (spec 076 T011): argument classification follows `new`'s
// conventions exactly, --at accepts only latest-snapshot, refusals are
// informed, and the summary carries every US1-scenario-5 fact.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worlds"
)

// forkableWorld creates a name-form world in the isolated home and gives it
// a snapshot to fork at (a synthetic covered log — no daemon run needed).
func forkableWorld(t *testing.T, name string) string {
	t.Helper()
	if err := cmdNew([]string{name, "--seed", "9"}); err != nil {
		t.Fatal(err)
	}
	home, err := worlds.WorldsHome()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, name)
	w, err := world.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	state := sim.NewState(w.Manifest.Seed, w.Map())
	if err := st.ReplayEvents(0, func(e store.Event) error { return state.Apply(e) }); err != nil {
		t.Fatal(err)
	}
	state.Tick = 100
	if err := st.SaveSnapshot(100, st.LastSeq(), state.Marshal()); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCmdForkNameForm(t *testing.T) {
	home := isolatedHome(t)
	forkableWorld(t, "aria")
	if err := cmdFork([]string{"aria", "aria-b"}); err != nil {
		t.Fatalf("cmdFork: %v", err)
	}
	forkDir := filepath.Join(home, "worlds", "aria-b")
	w, err := world.Open(forkDir)
	if err != nil {
		t.Fatalf("fork should be a readable world in the worlds home: %v", err)
	}
	if w.Manifest.Name != "aria-b" || w.Manifest.Seed != 9 {
		t.Errorf("fork manifest = name %q seed %d, want aria-b / 9 (seed carried)", w.Manifest.Name, w.Manifest.Seed)
	}
	if w.Manifest.Lineage == nil || w.Manifest.Lineage.Parent != "aria" {
		t.Errorf("fork lineage = %+v, want parent aria", w.Manifest.Lineage)
	}
}

func TestCmdForkPathForm(t *testing.T) {
	isolatedHome(t)
	forkableWorld(t, "aria")
	dest := filepath.Join(t.TempDir(), "duel-b")
	if err := cmdFork([]string{"aria", dest}); err != nil {
		t.Fatalf("cmdFork path form: %v", err)
	}
	w, err := world.Open(dest)
	if err != nil {
		t.Fatalf("fork should exist at the exact path: %v", err)
	}
	if w.Manifest.Name != "duel-b" {
		t.Errorf("path-form fork name = %q, want the basename duel-b", w.Manifest.Name)
	}
}

func TestCmdForkRefusals(t *testing.T) {
	isolatedHome(t)
	forkableWorld(t, "aria")

	// --at anything but latest-snapshot: informed refusal naming the follow-on.
	err := cmdFork([]string{"aria", "aria-b", "--at", "tick-500"})
	if err == nil || !strings.Contains(err.Error(), "follow-on") {
		t.Errorf("--at tick-500 = %v, want the documented-follow-on refusal", err)
	}

	// A bad bare name fails worlds.ValidateName, the `new` convention (the
	// other invalid shapes — leading '-'/'.', a '/' — classify as flags or
	// paths before validation, exactly as they do for `new`).
	if err := cmdFork([]string{"aria", ""}); err == nil || !strings.Contains(err.Error(), "must not be empty") {
		t.Errorf("an invalid bare <new-name> must refuse via worlds.ValidateName, got %v", err)
	}

	// Wrong arity: usage error.
	if err := cmdFork([]string{"aria"}); err == nil || !strings.Contains(err.Error(), "usage") {
		t.Errorf("missing <new-name> = %v, want a usage error", err)
	}
}

// TestForkSummaryContract pins US1 scenario 5: the summary names the
// boundary (game day/time and tick), the events carried, the truncated
// tail, the ended warning and wallet line when set, and the start-both
// next steps.
func TestForkSummaryContract(t *testing.T) {
	res := &world.ForkResult{
		Name: "aria-b", Dir: "/tmp/aria-b", ParentName: "aria",
		ForkTick: 97200, ForkSeq: 5000, TruncatedTail: 3,
		BoundaryEnded: true, SpendCarried: true,
	}
	out := forkSummary(res, "aria", "aria-b")
	for _, want := range []string{
		"day 2, 09:00", "tick 97200", // the boundary in game time
		"1..5000",                  // events carried
		"3 parent events past",     // the truncated tail
		`forked from "aria"`,       // lineage
		"born ended",               // the ended-boundary warning
		"never mints fresh budget", // the AC5 wallet line
		"promptworld start aria",   // start both
		"promptworld start aria-b",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("summary missing %q:\n%s", want, out)
		}
	}

	// The quiet path: no tail, live boundary, no spend — those lines vanish.
	quiet := forkSummary(&world.ForkResult{Name: "b", ParentName: "a", ForkTick: 0, ForkSeq: 10}, "a", "b")
	for _, absent := range []string{"truncated", "born ended", "wallet"} {
		if strings.Contains(quiet, absent) {
			t.Errorf("quiet summary should not mention %q:\n%s", absent, quiet)
		}
	}
}

// TestCmdForkLeavesNoRegistryState (spec 076 edge case): fork writes no
// registry entry — a home fork is scan-owned; an outside-home fork
// self-registers at its first daemon boot, like every world.
func TestCmdForkLeavesNoRegistryState(t *testing.T) {
	home := isolatedHome(t)
	forkableWorld(t, "aria")
	dest := filepath.Join(t.TempDir(), "outside-b")
	if err := cmdFork([]string{"aria", dest}); err != nil {
		t.Fatal(err)
	}
	reg := filepath.Join(home, "known_worlds.json")
	if data, err := os.ReadFile(reg); err == nil {
		var m map[string]any
		if json.Unmarshal(data, &m) == nil {
			if _, ok := m["outside-b"]; ok {
				t.Error("fork must not write registry state (the daemon self-registers at boot)")
			}
		}
	}
}
