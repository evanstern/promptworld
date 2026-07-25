package daemon

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/mind"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// TestContextReplayByteIdentical (spec 043 T024a) extends the replay harness
// (TestOnWorldReplayByteIdentical pattern) to the spec-043 decision-context
// state: a seeded world runs UNPAUSED with no embedder wired, its reflexes
// issuing intents and its executor decaying needs, so the log accrues the
// intent-lifecycle and needs-changed events that reduce into IntentLog rings and
// the NeedsAnchor/NeedsAnchorTick trajectory window. The log then replays through
// both recovery paths (snapshot recovery + genesis replay) byte-identically, and
// the assembled decision prompt — a pure function of that state — reproduces
// byte-for-byte from the recovered world. All new spec-043 state is
// reducer-derived, so replay determinism holds by construction.
func TestContextReplayByteIdentical(t *testing.T) {
	dir := t.TempDir() + "/w"
	w, err := world.Create(dir, "ctx-replay", 42)
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	state := sim.NewState(w.Manifest.Seed, w.Map())
	loop := sim.NewLoop(state, w.Map(), st, nil) // no embedder, no notify consumer

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- loop.Run(ctx) }()

	// Spin at max speed through the recorded set_speed door (never poke state
	// directly around a running loop — the replay-harness rule) until the
	// spec-043 state has accrued: some agent carries several intent records AND a
	// refreshed needs anchor. The anchor refreshes only after a full
	// trajectoryWindowTicks (1800) of game time, so this also guarantees the run
	// spans more than one trajectory window.
	if _, err := loop.Do("set_speed", "max"); err != nil {
		t.Fatalf("set_speed: %v", err)
	}
	sampleState := func() *sim.State {
		raw, _, err := loop.DoState()
		if err != nil {
			t.Fatalf("state: %v", err)
		}
		var s sim.State
		if err := json.Unmarshal(raw, &s); err != nil {
			t.Fatalf("unmarshal state: %v", err)
		}
		return &s
	}
	ready := func(s *sim.State) bool {
		if s.Tick <= 1800 {
			return false
		}
		for i := range s.Agents {
			if len(s.Agents[i].IntentLog) >= 3 && s.Agents[i].NeedsAnchor != nil {
				return true
			}
		}
		return false
	}
	deadline := time.Now().Add(30 * time.Second)
	for !ready(sampleState()) {
		if time.Now().After(deadline) {
			s := sampleState()
			t.Fatalf("spec-043 state never accrued within 30s (tick %d); no agent reached ≥3 intents + a needs anchor", s.Tick)
		}
		time.Sleep(20 * time.Millisecond)
	}
	if _, err := loop.Do("pause", ""); err != nil {
		t.Fatalf("pause: %v", err)
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("loop.Run: %v", err)
	}
	live := state.Marshal()

	// Snapshot-recovery leg: the daemon's own recoverState path reproduces the
	// state byte-identically.
	replayed, err := recoverState(w, st)
	if err != nil {
		t.Fatalf("recoverState: %v", err)
	}
	if got := string(replayed.Marshal()); got != string(live) {
		t.Fatalf("snapshot recovery diverged from the live state:\nlive:     %s\nreplayed: %s", live, got)
	}

	// Genesis-replay leg: a full log replay from tick 0 also reproduces it.
	fresh := sim.NewState(w.Manifest.Seed, w.Map())
	if err := st.ReplayEvents(0, func(e store.Event) error {
		if e.Tick > fresh.Tick {
			fresh.Tick = e.Tick
		}
		return fresh.Apply(e)
	}); err != nil {
		t.Fatalf("genesis replay: %v", err)
	}
	if got := string(fresh.Marshal()); got != string(live) {
		t.Fatalf("genesis replay diverged from the live state:\nlive:     %s\nreplayed: %s", live, got)
	}

	// The spec-043 reducer-derived fields specifically survive replay — asserted
	// per-agent against the live state object, not just implied by the whole-state
	// byte-identity above, so a regression that dropped one field would name it.
	// At least one agent must carry a populated ring + anchor (the ready() gate
	// guaranteed it), and every agent's ring/anchor must match field-for-field.
	sawRing, sawAnchor := false, false
	for i := range state.Agents {
		lv := state.Agents[i]
		rp := replayed.Agents[i]
		if !intentLogEqual(lv.IntentLog, rp.IntentLog) {
			t.Errorf("agent %d IntentLog diverged on replay:\nlive:     %+v\nreplayed: %+v", i, lv.IntentLog, rp.IntentLog)
		}
		if !needsAnchorEqual(lv.NeedsAnchor, lv.NeedsAnchorTick, rp.NeedsAnchor, rp.NeedsAnchorTick) {
			t.Errorf("agent %d NeedsAnchor diverged on replay: live (%v,%d) replayed (%v,%d)",
				i, lv.NeedsAnchor, lv.NeedsAnchorTick, rp.NeedsAnchor, rp.NeedsAnchorTick)
		}
		if len(rp.IntentLog) > 0 {
			sawRing = true
		}
		if rp.NeedsAnchor != nil {
			sawAnchor = true
		}
	}
	if !sawRing || !sawAnchor {
		t.Fatalf("replayed state carries no IntentLog ring (%v) or needs anchor (%v) — the spec-043 fields did not survive replay", sawRing, sawAnchor)
	}

	// The assembled decision prompt (spec 043 context.go) is a pure function of
	// world state, so the prompt assembled from the recovered world is
	// byte-identical to the one assembled from the live running state — for every
	// agent. Compared against the live *object* (not the marshalled bytes): the
	// prompt reads the agent's runtime-reconstructed mental map, which recoverState
	// rebuilds through the same reducer the live loop used.
	for i := range replayed.Agents {
		wantP := mind.AssembleUserPrompt(state, i, sim.WindowK, "")
		gotP := mind.AssembleUserPrompt(replayed, i, sim.WindowK, "")
		if gotP != wantP {
			t.Errorf("agent %d assembled prompt diverged on replay:\nlive:     %q\nreplayed: %q", i, wantP, gotP)
		}
	}
}

func intentLogEqual(a, b []sim.IntentRecord) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func needsAnchorEqual(a *sim.Needs, at int64, b *sim.Needs, bt int64) bool {
	if at != bt {
		return false
	}
	if (a == nil) != (b == nil) {
		return false
	}
	if a == nil {
		return true
	}
	return *a == *b
}

// TestSageThrashWindowContextReplay (spec 043 T013/T024b, SC-004) is the env-
// guarded evidence test for the thrash-window replay: pointed at a COPY of
// world-01's world.db (PROMPTWORLD_WORLD01_DB), it reconstructs Sage's (agent 7)
// state at tick 265864 via replayToTick and asserts the assembled decision
// context surfaces the reflex-issued forage redirection (source "instinct") and
// the forage/goto_warmth alternation the TASK-101 spike documented — inspection
// of assembled text only, no model in the loop. Skipped unless the env var is
// set. The test copies the world into its own temp dir before opening, so the
// real world-01 save is never touched even if store.Open migrates.
func TestSageThrashWindowContextReplay(t *testing.T) {
	const (
		sageIdx  = 7 // sim.AgentNames[7] == "Sage"
		thrashAt = int64(265864)
	)
	dbPath := os.Getenv("PROMPTWORLD_WORLD01_DB")
	if dbPath == "" {
		t.Skip("set PROMPTWORLD_WORLD01_DB to a COPY of world-01's world.db to run the SC-004 evidence replay")
	}
	if sim.AgentNames[sageIdx] != "Sage" {
		t.Fatalf("agent %d is %q, not Sage — the world roster changed", sageIdx, sim.AgentNames[sageIdx])
	}

	// Copy the world (world.json + world.db, plus any WAL sidecars) into an
	// isolated temp dir and open only the copy, so a schema-migrating store.Open
	// can never mutate the real save (task constraint: world-01 is READ-ONLY).
	srcDir := filepath.Dir(dbPath)
	tmp := t.TempDir()
	mustCopy(t, filepath.Join(srcDir, world.ManifestName), filepath.Join(tmp, world.ManifestName))
	mustCopy(t, dbPath, filepath.Join(tmp, "world.db"))
	for _, sidecar := range []string{"world.db-wal", "world.db-shm"} {
		if _, err := os.Stat(filepath.Join(srcDir, sidecar)); err == nil {
			mustCopy(t, filepath.Join(srcDir, sidecar), filepath.Join(tmp, sidecar))
		}
	}

	// Parse the manifest directly rather than through world.Open: world.Open
	// gates on format_version (world-01 is a legacy save older than this build),
	// but a read-only historical replay needs only the seed + map dimensions —
	// the current reducer replays the (format-stable) intent-lifecycle events
	// faithfully. The map is regenerated deterministically from the seed.
	var man struct {
		Seed      uint64 `json:"seed"`
		MapWidth  int    `json:"map_width"`
		MapHeight int    `json:"map_height"`
	}
	manBytes, err := os.ReadFile(filepath.Join(tmp, world.ManifestName))
	if err != nil {
		t.Fatalf("read manifest copy: %v", err)
	}
	if err := json.Unmarshal(manBytes, &man); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	m := worldmap.Generate(man.Seed, man.MapWidth, man.MapHeight)

	st, err := store.Open(filepath.Join(tmp, "world.db"))
	if err != nil {
		t.Fatalf("open store copy: %v", err)
	}
	defer st.Close()

	state, skipped, err := replayToTick(man.Seed, m, st, thrashAt)
	if err != nil {
		t.Fatalf("replayToTick %d: %v", thrashAt, err)
	}
	if state.Tick > thrashAt {
		t.Fatalf("reconstructed tick %d overshot the cutoff %d", state.Tick, thrashAt)
	}
	// The intent-lifecycle events that build IntentLog must all have applied —
	// only then is the self_history reconstruction trustworthy. Legacy events the
	// current invariants reject (e.g. omens) may be skipped, but never these.
	for _, typ := range []string{"agent.intent_set", "agent.intent_done", "agent.intent_rejected", "agent.build_failed", "agent.plan_expired"} {
		if skipped[typ] > 0 {
			t.Fatalf("%d %s event(s) were rejected on replay — the self_history reconstruction is not trustworthy", skipped[typ], typ)
		}
	}
	if len(skipped) > 0 {
		t.Logf("SC-004 replay tolerated skipped legacy events (rejected by current reducer invariants, unrelated to intent history): %v", skipped)
	}

	sage := state.Agents[sageIdx]
	if len(sage.IntentLog) == 0 {
		t.Fatalf("Sage has no intent history at tick %d — wrong tick or wrong db", thrashAt)
	}

	prompt := mind.AssembleUserPrompt(state, sageIdx, sim.WindowK, "")

	// SC-004 shape assertions on the assembled text.
	// 1. The reflex-issued forage shows as an instinct-sourced record.
	if !strings.Contains(prompt, "instinct drove this") {
		t.Errorf("assembled self_history shows no reflex/instinct-sourced record (SC-004 redirection):\n%s", prompt)
	}
	// 2. Both forage and goto_warmth appear in the recent history — the
	//    alternation the spike documented.
	if !strings.Contains(prompt, "forage") {
		t.Errorf("assembled self_history names no forage intent (SC-004 alternation):\n%s", prompt)
	}
	if !strings.Contains(prompt, "goto_warmth") {
		t.Errorf("assembled self_history names no goto_warmth intent (SC-004 alternation):\n%s", prompt)
	}
	// 3. The alternation is visible across the recent-intent lines: at least one
	//    forage and one goto_warmth among the self_history "Recently you:" block.
	sh := selfHistoryRegion(prompt)
	if sh == "" {
		t.Fatalf("no self_history block in the assembled prompt:\n%s", prompt)
	}
	if !strings.Contains(sh, "forage") || !strings.Contains(sh, "goto_warmth") {
		t.Errorf("forage/goto_warmth alternation not visible in the self_history block (SC-004):\n%s", sh)
	}

	// Emit the evidence for the human record (visible under `go test -v`).
	t.Logf("SC-004 replay — Sage (agent %d) at tick %d\n=== assembled decision context ===\n%s\n=== IntentLog ring ===\n%s",
		sageIdx, thrashAt, prompt, formatRing(sage.IntentLog))
}

// selfHistoryRegion returns the "Recently you:" block of the assembled prompt
// (its "- " lines), empty when the first-thought empty-state rendered instead.
func selfHistoryRegion(prompt string) string {
	const head = "Recently you:\n"
	i := strings.Index(prompt, head)
	if i < 0 {
		return ""
	}
	rest := prompt[i+len(head):]
	var b strings.Builder
	b.WriteString(head)
	for _, line := range strings.Split(rest, "\n") {
		if !strings.HasPrefix(line, "- ") {
			break
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func formatRing(ring []sim.IntentRecord) string {
	var b strings.Builder
	for _, r := range ring {
		b.WriteString(r.Goal)
		b.WriteString(" [")
		b.WriteString(r.Source)
		b.WriteString("] @")
		b.WriteString(strconv.FormatInt(r.Tick, 10))
		if r.Outcome != "" {
			b.WriteString(" -> ")
			b.WriteString(r.Outcome)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// mustCopy streams src to dst, failing the test on any error — used to isolate a
// real world save from the test's opens.
func mustCopy(t *testing.T, src, dst string) {
	t.Helper()
	in, err := os.Open(src)
	if err != nil {
		t.Fatalf("open %s: %v", src, err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		t.Fatalf("create %s: %v", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		t.Fatalf("copy %s -> %s: %v", src, dst, err)
	}
	if err := out.Close(); err != nil {
		t.Fatalf("close %s: %v", dst, err)
	}
}
