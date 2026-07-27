package world

// Fork ceremony tests (spec 076 T008–T010): the happy path (prefix
// contiguity, verbatim boundary snapshot, exact lineage event, manifest
// carry, sidecar copy/skip), the refusal table, partial-failure cleanup,
// the AC4 determinism proofs (genesis replay of the fork log reproduces the
// boundary snapshot's state_hash — which IS the parent's state hash at the
// same (tick, seq)), and the AC5 wallet-inheritance proof.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// forkParent builds a current-format world with a real event history: a
// genesis marker plus state-mutating sim events, a boundary snapshot
// covering the prefix (state folded through the reducer, exactly as
// recovery would), and a two-event uncovered tail past the boundary. The
// snapshot state's hash is what every determinism proof keys on.
func forkParent(t *testing.T, dir string) (w *World, boundaryTick, boundarySeq int64) {
	t.Helper()
	const seed = 42
	w, err := Create(dir, "parent", seed)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	st, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	events := []store.Event{
		{Tick: 0, Type: "world.created", Payload: forkJSON(t, sim.WorldCreatedPayload{Name: "parent", Seed: seed})},
		{Tick: 100, Type: "agent.moved", Payload: forkJSON(t, sim.AgentMovedPayload{Agent: sim.Ref(0), X: 3, Y: 4})},
		{Tick: 200, Type: "clock.speed_set", Payload: forkJSON(t, map[string]string{"speed": "4x"})},
		{Tick: 900, Type: "chronicle.entry", Payload: forkJSON(t, sim.ChronicleEntryPayload{Day: 1, FromTick: 0, ToTick: 900, Text: "the village woke"})},
		{Tick: 1000, Type: "agent.moved", Payload: forkJSON(t, sim.AgentMovedPayload{Agent: sim.Ref(1), X: 5, Y: 6})},
	}
	if err := st.AppendEvents(events); err != nil {
		t.Fatal(err)
	}

	// The boundary snapshot: state folded from the full prefix, tick set the
	// way recovery sets it (max of snapshot/last-event tick).
	state := sim.NewState(seed, w.Map())
	if err := st.ReplayEvents(0, func(e store.Event) error { return state.Apply(e) }); err != nil {
		t.Fatal(err)
	}
	boundaryTick, boundarySeq = int64(1000), st.LastSeq()
	state.Tick = boundaryTick
	if err := st.SaveSnapshot(boundaryTick, boundarySeq, state.Marshal()); err != nil {
		t.Fatal(err)
	}

	// An uncovered tail past the boundary — what "truncated to the snapshot
	// boundary" deliberately leaves behind.
	tail := []store.Event{
		{Tick: 1100, Type: "agent.moved", Payload: forkJSON(t, sim.AgentMovedPayload{Agent: sim.Ref(0), X: 7, Y: 8})},
		{Tick: 1200, Type: "agent.moved", Payload: forkJSON(t, sim.AgentMovedPayload{Agent: sim.Ref(1), X: 9, Y: 9})},
	}
	if err := st.AppendEvents(tail); err != nil {
		t.Fatal(err)
	}
	return w, boundaryTick, boundarySeq
}

func forkJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestForkCeremonyHappyPath(t *testing.T) {
	base := t.TempDir()
	parentDir, destDir := filepath.Join(base, "parent"), filepath.Join(base, "fork-a")
	_, boundaryTick, boundarySeq := forkParent(t, parentDir)

	// Sidecars: some that must copy, some that must not.
	for name, body := range map[string]string{
		"llm.json":             `{"monthly_budget_usd":5}`,
		"charter.md":           "keep everyone alive",
		"tuning.json":          `{}`,
		"chronicle.md":         "scribe view — must NOT copy",
		"morgue.md":            "scribe view — must NOT copy",
		"village_charter.md":   "scribe view — must NOT copy",
		"daemon.log":           "runtime — must NOT copy",
		"world.v1.db":          "migration archive — must NOT copy",
		"metatron/soul.md":     "guardian soul",
		"agents/notes.txt":     "drop-in",
		"bundles/b1/tool.json": `{}`,
	} {
		p := filepath.Join(parentDir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	res, err := Fork(parentDir, destDir, "fork-a")
	if err != nil {
		t.Fatalf("Fork: %v", err)
	}
	if res.Name != "fork-a" || res.Dir != destDir || res.ParentName != "parent" {
		t.Errorf("ForkResult identity = %+v", res)
	}
	if res.ForkTick != boundaryTick || res.ForkSeq != boundarySeq {
		t.Errorf("ForkResult boundary = (%d, %d), want (%d, %d)", res.ForkTick, res.ForkSeq, boundaryTick, boundarySeq)
	}
	if res.TruncatedTail != 2 {
		t.Errorf("TruncatedTail = %d, want 2", res.TruncatedTail)
	}
	if res.BoundaryEnded {
		t.Errorf("BoundaryEnded = true on a live-boundary parent")
	}
	if res.SpendCarried { // no llm_spend_* meta on this parent
		t.Errorf("SpendCarried = %v, want false", res.SpendCarried)
	}

	// The fork log: contiguous 1..N+1 — the parent's prefix verbatim plus
	// exactly one world.forked.
	parentStore, err := store.Open(filepath.Join(parentDir, "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer parentStore.Close()
	forkStore, err := store.Open(filepath.Join(destDir, "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer forkStore.Close()
	if err := forkStore.CheckContiguity(); err != nil {
		t.Errorf("fork log contiguity: %v", err)
	}
	forkEvents, err := forkStore.EventsSince(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if int64(len(forkEvents)) != boundarySeq+1 {
		t.Fatalf("fork log has %d events, want %d", len(forkEvents), boundarySeq+1)
	}
	parentEvents, err := parentStore.EventsSince(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	for i := int64(0); i < boundarySeq; i++ {
		p, f := parentEvents[i], forkEvents[i]
		if p.Seq != f.Seq || p.Tick != f.Tick || p.Type != f.Type || string(p.Payload) != string(f.Payload) {
			t.Errorf("prefix event %d diverged:\nparent: %+v\nfork:   %+v", i+1, p, f)
		}
	}
	forked := forkEvents[boundarySeq]
	if forked.Type != "world.forked" || forked.Tick != boundaryTick || forked.Seq != boundarySeq+1 {
		t.Errorf("lineage event = %+v, want world.forked at (tick %d, seq %d)", forked, boundaryTick, boundarySeq+1)
	}
	var lp sim.WorldForkedPayload
	if err := json.Unmarshal(forked.Payload, &lp); err != nil {
		t.Fatal(err)
	}
	if lp.ParentName != "parent" || lp.ParentSeed != 42 || lp.ForkTick != boundaryTick || lp.ForkSeq != boundarySeq || lp.ParentCreatedAt == "" {
		t.Errorf("WorldForkedPayload = %+v", lp)
	}
	// Exactly one world.forked in the whole log.
	count := 0
	for _, e := range forkEvents {
		if e.Type == "world.forked" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("fork log carries %d world.forked events, want exactly 1", count)
	}

	// The boundary snapshot rides verbatim: same (tick, seq), same hash.
	parentSnap, err := parentStore.LatestValidSnapshot()
	if err != nil || parentSnap == nil {
		t.Fatalf("parent snapshot: %v, %v", parentSnap, err)
	}
	forkSnap, err := forkStore.LatestValidSnapshot()
	if err != nil || forkSnap == nil {
		t.Fatalf("fork snapshot: %v, %v", forkSnap, err)
	}
	if forkSnap.Tick != parentSnap.Tick || forkSnap.Seq != parentSnap.Seq || forkSnap.Hash != parentSnap.Hash {
		t.Errorf("fork snapshot (tick %d, seq %d, %s) != parent (tick %d, seq %d, %s)",
			forkSnap.Tick, forkSnap.Seq, forkSnap.Hash, parentSnap.Tick, parentSnap.Seq, parentSnap.Hash)
	}

	// Meta stamped for validateMeta's first-boot check.
	if v, _ := forkStore.GetMeta("seed"); v != "42" {
		t.Errorf("fork meta seed = %q, want 42", v)
	}
	if v, _ := forkStore.GetMeta("format_version"); v != strconv.Itoa(FormatVersion) {
		t.Errorf("fork meta format_version = %q, want %d", v, FormatVersion)
	}

	// The manifest: name/created_at/lineage new, EVERYTHING else verbatim.
	parentW, err := Open(parentDir)
	if err != nil {
		t.Fatal(err)
	}
	forkW, err := Open(destDir)
	if err != nil {
		t.Fatalf("fork manifest fails Open: %v", err)
	}
	fm, pm := forkW.Manifest, parentW.Manifest
	if fm.Name != "fork-a" {
		t.Errorf("fork name = %q", fm.Name)
	}
	if fm.Lineage == nil || fm.Lineage.Parent != "parent" || fm.Lineage.ForkTick != boundaryTick || fm.Lineage.ParentCreatedAt != pm.CreatedAt {
		t.Errorf("fork lineage = %+v", fm.Lineage)
	}
	// Field-by-field carry: neutralize the deliberately-new fields, then the
	// rest must compare equal wholesale.
	fm.Name, fm.CreatedAt, fm.Lineage = pm.Name, pm.CreatedAt, pm.Lineage
	if forkJSONString(t, fm) != forkJSONString(t, pm) {
		t.Errorf("manifest carry diverged:\nfork:   %+v\nparent: %+v", forkW.Manifest, pm)
	}
	if fm.Seed != 42 {
		t.Errorf("fork seed = %d, want the parent's 42 (seed is CARRIED — identity is name/dir/socket)", fm.Seed)
	}

	// Sidecar catalog (research R9): copies present, skips absent.
	for _, name := range []string{"llm.json", "charter.md", "tuning.json", "metatron/soul.md", "agents/notes.txt", "bundles/b1/tool.json"} {
		if _, err := os.Stat(filepath.Join(destDir, name)); err != nil {
			t.Errorf("sidecar %s should have copied: %v", name, err)
		}
	}
	for _, name := range []string{"chronicle.md", "morgue.md", "village_charter.md", "daemon.log", "daemon.pid", "daemon.sock", "world.v1.db"} {
		if _, err := os.Stat(filepath.Join(destDir, name)); err == nil {
			t.Errorf("%s must NOT copy into the fork", name)
		}
	}
}

// TestForkRefusals is the refusal table (spec 076 US1 scenario 4 + edge
// cases): no valid snapshot, non-empty destination, live source daemon —
// each refused with the world untouched and no partial destination.
func TestForkRefusals(t *testing.T) {
	t.Run("no snapshot", func(t *testing.T) {
		base := t.TempDir()
		dir := filepath.Join(base, "parent")
		w, err := Create(dir, "parent", 7)
		if err != nil {
			t.Fatal(err)
		}
		st, err := store.Open(w.DBPath())
		if err != nil {
			t.Fatal(err)
		}
		if err := st.AppendEvents([]store.Event{{Tick: 0, Type: "world.created",
			Payload: forkJSON(t, sim.WorldCreatedPayload{Name: "parent", Seed: 7})}}); err != nil {
			t.Fatal(err)
		}
		st.Close()
		dest := filepath.Join(base, "fork")
		_, err = Fork(dir, dest, "fork")
		if err == nil || !strings.Contains(err.Error(), "no valid snapshot") {
			t.Fatalf("Fork without a snapshot = %v, want the no-snapshot refusal", err)
		}
		if !strings.Contains(err.Error(), "start") || !strings.Contains(err.Error(), "stop") {
			t.Errorf("refusal should name the start-and-stop remedy, got %q", err)
		}
		if _, statErr := os.Stat(dest); statErr == nil {
			t.Error("refusal left a destination behind")
		}
	})

	t.Run("non-empty destination", func(t *testing.T) {
		base := t.TempDir()
		parentDir := filepath.Join(base, "parent")
		forkParent(t, parentDir)
		dest := filepath.Join(base, "occupied")
		if err := os.MkdirAll(dest, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dest, "keep.txt"), []byte("mine"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Fork(parentDir, dest, "fork"); err == nil || !strings.Contains(err.Error(), "not empty") {
			t.Fatalf("Fork into a non-empty dir = %v, want the not-empty refusal", err)
		}
		if _, err := os.Stat(filepath.Join(dest, "keep.txt")); err != nil {
			t.Error("the refusal must not touch the occupied destination")
		}
	})

	t.Run("running daemon", func(t *testing.T) {
		base := t.TempDir()
		parentDir := filepath.Join(base, "parent")
		forkParent(t, parentDir)
		if err := os.WriteFile(filepath.Join(parentDir, "daemon.pid"),
			[]byte(strconv.Itoa(os.Getpid())+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := Fork(parentDir, filepath.Join(base, "fork"), "fork")
		if err == nil || !strings.Contains(err.Error(), "daemon is running") {
			t.Fatalf("Fork of a live world = %v, want the running-daemon refusal", err)
		}
		if !strings.Contains(err.Error(), "promptworld stop") {
			t.Errorf("refusal should name the stop command, got %q", err)
		}
	})
}

// TestForkPartialFailureCleanup: a ceremony failure past destination
// creation removes the partial destination best-effort (research R9), so a
// retry is clean.
func TestForkPartialFailureCleanup(t *testing.T) {
	base := t.TempDir()
	parentDir := filepath.Join(base, "parent")
	forkParent(t, parentDir)
	// A metatron that is a FILE makes the sidecar dir copy fail mid-ceremony.
	if err := os.WriteFile(filepath.Join(parentDir, "metatron"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(base, "fork")
	if _, err := Fork(parentDir, dest, "fork"); err == nil {
		t.Fatal("Fork should fail on an uncopyable sidecar")
	}
	if _, err := os.Stat(dest); err == nil {
		t.Error("a failed fork must remove its partial destination")
	}
}

// TestForkDeterminismProofs is board AC #4 / FR-010 (a)+(b): replaying the
// fork's log from genesis through the reducer reproduces the boundary
// snapshot's state_hash exactly — and that hash IS the parent's state hash
// at the same (tick, seq), because the world.forked arm is a recorded-
// history no-op (byte-identity, US1 scenario 3).
func TestForkDeterminismProofs(t *testing.T) {
	base := t.TempDir()
	parentDir, destDir := filepath.Join(base, "parent"), filepath.Join(base, "fork")
	_, boundaryTick, _ := forkParent(t, parentDir)
	if _, err := Fork(parentDir, destDir, "fork"); err != nil {
		t.Fatal(err)
	}

	forkW, err := Open(destDir)
	if err != nil {
		t.Fatal(err)
	}
	forkStore, err := store.Open(forkW.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer forkStore.Close()

	// (a) genesis replay of the fork's FULL log (world.forked included).
	replayed := sim.NewState(forkW.Manifest.Seed, forkW.Map())
	var lastTick int64
	if err := forkStore.ReplayEvents(0, func(e store.Event) error {
		if e.Tick > lastTick {
			lastTick = e.Tick
		}
		return replayed.Apply(e)
	}); err != nil {
		t.Fatal(err)
	}
	if replayed.Tick < lastTick {
		replayed.Tick = lastTick // recovery's Tick = max(snapshot tick, last event tick)
	}
	if lastTick != boundaryTick {
		t.Fatalf("fork log's last tick = %d, want the boundary %d", lastTick, boundaryTick)
	}

	forkSnap, err := forkStore.LatestValidSnapshot()
	if err != nil || forkSnap == nil {
		t.Fatalf("fork snapshot: %v, %v", forkSnap, err)
	}
	if got := replayed.Hash(); got != forkSnap.Hash {
		t.Errorf("genesis replay of the fork log hashes %s, want the boundary snapshot's %s", got, forkSnap.Hash)
	}

	// (b) the same hash is the PARENT's state hash at (fork tick, fork seq).
	parentStore, err := store.Open(filepath.Join(parentDir, "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer parentStore.Close()
	parentSnap, err := parentStore.LatestValidSnapshot()
	if err != nil || parentSnap == nil {
		t.Fatalf("parent snapshot: %v, %v", parentSnap, err)
	}
	if replayed.Hash() != parentSnap.Hash {
		t.Errorf("fork state at the fork tick (%s) != parent state at the same (tick, seq) (%s) — byte-identity broken",
			replayed.Hash(), parentSnap.Hash)
	}
}

// TestForkInheritsWallet is board AC #5 / SC-006: a fork of a world with
// recorded month spend opens its meter at the parent's spend — total AND
// per-provider attribution — never a fresh grant.
func TestForkInheritsWallet(t *testing.T) {
	base := t.TempDir()
	parentDir, destDir := filepath.Join(base, "parent"), filepath.Join(base, "fork")
	forkParent(t, parentDir)

	month := time.Now().UTC().Format("2006-01")
	parentStore, err := store.Open(filepath.Join(parentDir, "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range map[string]string{
		"llm_spend_" + month:            "1.25",
		"llm_spend_" + month + ":cloud": "1",
		"llm_spend_" + month + ":local": "0.25",
	} {
		if err := parentStore.SetMeta(k, v); err != nil {
			t.Fatal(err)
		}
	}
	parentStore.Close()

	res, err := Fork(parentDir, destDir, "fork")
	if err != nil {
		t.Fatal(err)
	}
	if !res.SpendCarried {
		t.Error("SpendCarried = false, want true (llm_spend_* keys were present)")
	}

	forkStore, err := store.Open(filepath.Join(destDir, "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer forkStore.Close()
	meter, err := llm.NewMeter(forkStore, 5.0, []string{"cloud", "local"})
	if err != nil {
		t.Fatal(err)
	}
	gotMonth, spent, budget, perProvider := meter.Snapshot()
	if gotMonth != month || spent != 1.25 || budget != 5.0 {
		t.Errorf("fork meter opens at (%s, %v, %v), want (%s, 1.25, 5.0) — forking never mints fresh budget",
			gotMonth, spent, budget, month)
	}
	if perProvider["cloud"] != 1 || perProvider["local"] != 0.25 {
		t.Errorf("per-provider attribution = %v, want cloud=1 local=0.25", perProvider)
	}
}

// forkJSONString marshals a manifest for wholesale field comparison.
func forkJSONString(t *testing.T, m Manifest) string {
	t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
