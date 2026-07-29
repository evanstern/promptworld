package main

// compare tests (spec 076 T016 + T018): SC-004 — the decided duel renders
// the winner all-✓ from its recorded pass and the loser's concluded ✗ with
// honest backing, no raw enum anywhere in the output, facts identical to
// the shared resolver's; SC-005 — divergence lands on the first differing
// STORY event (machinery-only differences render NO divergence), the
// zero-divergence line is pinned, the interleave is labeled and
// tick-ordered, and the fork-lineage default window (with --since override)
// holds.

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tui"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worlds"
)

// duelWorld builds a fixture world: created on disk, optionally running a
// scenario, with the given events appended after the genesis marker.
func duelWorld(t *testing.T, name, scenario string, events []store.Event) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), name)
	w, err := world.Create(dir, name, 5)
	if err != nil {
		t.Fatal(err)
	}
	if scenario != "" {
		if err := world.SetScenario(dir, scenario); err != nil {
			t.Fatal(err)
		}
	}
	st, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	genesis := []store.Event{{Tick: 0, Type: "world.created",
		Payload: duelJSON(t, sim.WorldCreatedPayload{Name: name, Seed: 5})}}
	if err := st.AppendEvents(append(genesis, events...)); err != nil {
		t.Fatal(err)
	}
	return dir
}

func duelJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// deathAndEnd is the loser's history: two deaths, then the run-over
// declaration.
func deathAndEnd(t *testing.T) []store.Event {
	t.Helper()
	deaths := []sim.DeathRecord{
		{Agent: 0, Tick: 40000, Cause: "gru"},
		{Agent: 1, Tick: 50000, Cause: "exposure"},
	}
	return []store.Event{
		{Tick: 40000, Type: "agent.died", Payload: duelJSON(t, map[string]any{"agent": 0, "cause": "gru"})},
		{Tick: 50000, Type: "agent.died", Payload: duelJSON(t, map[string]any{"agent": 1, "cause": "exposure"})},
		{Tick: 50000, Type: "run.ended", Payload: duelJSON(t, sim.RunEndedPayload{
			Tick: 50000, Deaths: sim.DeathRefs(deaths), FinalCause: "exposure"})},
	}
}

// passEvents is the winner's history: the recorded pass instrument.
func passEvents(t *testing.T) []store.Event {
	t.Helper()
	return []store.Event{
		{Tick: 30000, Type: "guardian.order_placed", Payload: duelJSON(t, map[string]any{})},
		{Tick: 86400, Type: "curriculum.exercise_passed", Payload: duelJSON(t, sim.ExercisePassedPayload{
			Exercise: "first-night", Stage: "stage-1", Tick: 86400,
			Evidence: []sim.EvidenceRef{{Type: "guardian.order_placed", Seq: 2, Tick: 30000}},
		})},
	}
}

// TestCompareDecidedDuel is SC-004: winner all-✓ from its recorded pass,
// loser concluded-✗ with the honest agent.died count, plain-language
// outcomes, and a raw-enum sweep over the whole rendered output.
func TestCompareDecidedDuel(t *testing.T) {
	winner := duelWorld(t, "champ", "first-night", passEvents(t))
	loser := duelWorld(t, "fallen", "first-night", deathAndEnd(t))

	r, err := buildDuelReport(winner, loser, -1)
	if err != nil {
		t.Fatal(err)
	}
	out := renderDuelReport(r)

	// Winner: every row ✓ (the pass is the instrument), evidence-backed
	// where the pass's own Evidence names the term's event.
	if r.A.Mode != tui.ReportCardConcluded {
		t.Errorf("winner mode = %v, want concluded", r.A.Mode)
	}
	for _, f := range r.A.Facts {
		if !f.Met {
			t.Errorf("winner fact %q unmet — a recorded pass must grade all-met", f.Term)
		}
	}
	if !strings.Contains(out, "champ — passed") {
		t.Errorf("winner outcome line missing:\n%s", out)
	}
	if !strings.Contains(out, "guardian.order_placed · seq 2") {
		t.Errorf("pass evidence backing missing:\n%s", out)
	}

	// Loser: the postmortem card — concluded ✗ on the death term with the
	// honest backing count, outcome in the no-blame register.
	if !strings.Contains(out, "✗ no villager dies (agent.died: 2)") {
		t.Errorf("loser's failed term must render ✗ with its honest count:\n%s", out)
	}
	if !strings.Contains(out, "fallen — did not make it through") {
		t.Errorf("loser outcome must render in the postmortem register:\n%s", out)
	}

	// The raw-enum sweep (FR-019/SC-004): the enum tokens never print.
	for _, raw := range []string{"in_progress", "— failed", ": failed"} {
		if strings.Contains(out, raw) {
			t.Errorf("raw enum token %q leaked into the duel output:\n%s", raw, out)
		}
	}

	// One resolver (FR-018): the report's facts are byte-equal to a direct
	// call of the exported spec-072 resolver over the same offline state.
	ww, err := world.Open(winner)
	if err != nil {
		t.Fatal(err)
	}
	state, _, err := worlds.OfflineState(ww)
	if err != nil {
		t.Fatal(err)
	}
	def, _ := sim.ExerciseByID("first-night")
	facts, mode := tui.ResolveRubricFacts(state, def, tui.RecordedPassFor(state, def.ID))
	if len(facts) != len(r.A.Facts) || mode != r.A.Mode {
		t.Fatalf("compare facts diverge from the shared resolver: %d/%v vs %d/%v",
			len(r.A.Facts), r.A.Mode, len(facts), mode)
	}
	for i := range facts {
		if facts[i] != r.A.Facts[i] {
			t.Errorf("fact %d diverges from the shared resolver:\ncompare:  %+v\nresolver: %+v", i, r.A.Facts[i], facts[i])
		}
	}
}

// TestCompareLiveWorld is US2 scenario 3: a still-running scored world
// renders live markers (… pending) and the still-running outcome.
func TestCompareLiveWorld(t *testing.T) {
	live := duelWorld(t, "alive", "first-night", []store.Event{
		{Tick: 1000, Type: "agent.moved", Payload: duelJSON(t, sim.AgentMovedPayload{Agent: sim.Ref(0), X: 1, Y: 1})},
	})
	other := duelWorld(t, "other", "first-night", deathAndEnd(t))
	r, err := buildDuelReport(live, other, -1)
	if err != nil {
		t.Fatal(err)
	}
	if r.A.Mode != tui.ReportCardLive {
		t.Errorf("live world mode = %v, want live", r.A.Mode)
	}
	out := renderDuelReport(r)
	if !strings.Contains(out, "alive — still running") {
		t.Errorf("live outcome must read plainly, never in_progress:\n%s", out)
	}
	if !strings.Contains(out, "…") {
		t.Errorf("live card should carry pending markers:\n%s", out)
	}
}

// TestCompareAmbientHonesty is US2 scenario 4: an ambient world gets an
// honest no-scorecard note, never an invented card — the chronicle sections
// still render.
func TestCompareAmbientHonesty(t *testing.T) {
	ambient := duelWorld(t, "plainland", "", []store.Event{
		{Tick: 1000, Type: "chronicle.entry", Payload: duelJSON(t, sim.ChronicleEntryPayload{
			Day: 1, FromTick: 900, ToTick: 1000, Text: "a quiet morning"})},
	})
	scored := duelWorld(t, "scored", "first-night", deathAndEnd(t))
	r, err := buildDuelReport(ambient, scored, -1)
	if err != nil {
		t.Fatal(err)
	}
	if r.A.Exercise != nil {
		t.Error("ambient side resolved an exercise")
	}
	out := renderDuelReport(r)
	if !strings.Contains(out, "plainland — an ambient world: no rubric exists") {
		t.Errorf("ambient honesty line missing:\n%s", out)
	}
	if !strings.Contains(out, "a quiet morning") {
		t.Errorf("chronicle drill-down should still render for an ambient world:\n%s", out)
	}
}

// TestCompareDifferentExercises is US2 scenario 5: each card renders under
// its own exercise, with the not-head-to-head note.
func TestCompareDifferentExercises(t *testing.T) {
	fn := duelWorld(t, "night-w", "first-night", nil)
	law := duelWorld(t, "law-w", "the-law", nil)
	r, err := buildDuelReport(fn, law, -1)
	if err != nil {
		t.Fatal(err)
	}
	out := renderDuelReport(r)
	if !strings.Contains(out, "different exercises") || !strings.Contains(out, "not head-to-head") {
		t.Errorf("different-exercise note missing:\n%s", out)
	}
	if !strings.Contains(out, "report card · first-night") || !strings.Contains(out, "report card · the-law") {
		t.Errorf("each side must render under its own exercise:\n%s", out)
	}
}

// TestCompareDivergence is SC-005's placement half: the marker lands on the
// first differing story event, rendered in game time.
func TestCompareDivergence(t *testing.T) {
	shared := store.Event{Tick: 1000, Type: "agent.moved", Payload: duelJSON(t, sim.AgentMovedPayload{Agent: sim.Ref(0), X: 1, Y: 1})}
	a := duelWorld(t, "same-a", "", []store.Event{shared,
		{Tick: 2000, Type: "agent.moved", Payload: duelJSON(t, sim.AgentMovedPayload{Agent: sim.Ref(0), X: 2, Y: 2})},
	})
	b := duelWorld(t, "same-b", "", []store.Event{shared,
		{Tick: 2000, Type: "agent.foraged", Payload: duelJSON(t, map[string]any{"agent": 0, "x": 2, "y": 2})},
	})
	r, err := buildDuelReport(a, b, 500) // window past genesis (the differing world.created names)
	if err != nil {
		t.Fatal(err)
	}
	if r.Divergence == nil {
		t.Fatal("expected a divergence")
	}
	if r.Divergence.Tick != 2000 {
		t.Errorf("divergence tick = %d, want 2000 (the first differing story event)", r.Divergence.Tick)
	}
	if r.Divergence.A != "agent.moved" || r.Divergence.B != "agent.foraged" {
		t.Errorf("divergence sides = %q / %q", r.Divergence.A, r.Divergence.B)
	}
	out := renderDuelReport(r)
	if !strings.Contains(out, "the stories diverge at day 1") || !strings.Contains(out, "tick 2000") {
		t.Errorf("divergence line missing or not in game time:\n%s", out)
	}
}

// TestCompareMachineryNeverDiverges is SC-005's exclusion half: a pair
// differing ONLY in machinery classes (daemon.*, clock.*, cog.*, llm.*)
// and chronicle wording renders NO divergence — two runs that told the
// same story at different wall speeds do not falsely diverge.
func TestCompareMachineryNeverDiverges(t *testing.T) {
	story := store.Event{Tick: 1000, Type: "agent.moved", Payload: duelJSON(t, sim.AgentMovedPayload{Agent: sim.Ref(0), X: 1, Y: 1})}
	a := duelWorld(t, "wall-a", "", []store.Event{story,
		{Tick: 1100, Type: "daemon.started", Payload: duelJSON(t, sim.DaemonStartedPayload{Tick: 1100, RecoveryMs: 12})},
		{Tick: 1200, Type: "clock.speed_set", Payload: duelJSON(t, map[string]string{"speed": "4x"})},
		{Tick: 1300, Type: "chronicle.entry", Payload: duelJSON(t, sim.ChronicleEntryPayload{Day: 1, FromTick: 1000, Text: "dawn broke gently"})},
	})
	b := duelWorld(t, "wall-b", "", []store.Event{story,
		{Tick: 1150, Type: "daemon.started", Payload: duelJSON(t, sim.DaemonStartedPayload{Tick: 1150, RecoveryMs: 99})},
		{Tick: 1250, Type: "clock.speed_set", Payload: duelJSON(t, map[string]string{"speed": "max"})},
		{Tick: 1300, Type: "chronicle.entry", Payload: duelJSON(t, sim.ChronicleEntryPayload{Day: 1, FromTick: 1000, Text: "the sun rose over the marsh"})},
	})
	r, err := buildDuelReport(a, b, 500)
	if err != nil {
		t.Fatal(err)
	}
	if r.Divergence != nil {
		t.Errorf("machinery/chronicle-only differences must not diverge, got %+v", r.Divergence)
	}
	out := renderDuelReport(r)
	if !strings.Contains(out, "identical since") {
		t.Errorf("zero-divergence line missing:\n%s", out)
	}
}

// TestCompareForkPair is US3 scenarios 1 + 4 over a REAL fork: the default
// window is the fork tick read from lineage, identical post-fork histories
// render the exact identical-since-the-fork line, and --since overrides
// the window.
func TestCompareForkPair(t *testing.T) {
	base := t.TempDir()
	parentDir := filepath.Join(base, "aria")
	w, err := world.Create(parentDir, "aria", 5)
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := st.AppendEvents([]store.Event{
		{Tick: 0, Type: "world.created", Payload: duelJSON(t, sim.WorldCreatedPayload{Name: "aria", Seed: 5})},
		{Tick: 1000, Type: "agent.moved", Payload: duelJSON(t, sim.AgentMovedPayload{Agent: sim.Ref(0), X: 1, Y: 1})},
	}); err != nil {
		t.Fatal(err)
	}
	state := sim.NewState(5, w.Map())
	if err := st.ReplayEvents(0, func(e store.Event) error { return state.Apply(e) }); err != nil {
		t.Fatal(err)
	}
	state.Tick = 1000
	if err := st.SaveSnapshot(1000, st.LastSeq(), state.Marshal()); err != nil {
		t.Fatal(err)
	}
	st.Close()

	forkDir := filepath.Join(base, "aria-b")
	res, err := world.Fork(parentDir, forkDir, "aria-b")
	if err != nil {
		t.Fatal(err)
	}

	// Identical post-fork histories: nothing appended to either side.
	r, err := buildDuelReport(parentDir, forkDir, -1)
	if err != nil {
		t.Fatal(err)
	}
	if r.Lineage == nil || r.Lineage.Parent != "aria" || r.Lineage.Child != "aria-b" {
		t.Fatalf("lineage = %+v, want aria-b forked from aria", r.Lineage)
	}
	if r.Since != res.ForkTick {
		t.Errorf("default window = %d, want the fork tick %d", r.Since, res.ForkTick)
	}
	if r.Divergence != nil {
		t.Errorf("identical post-fork histories must not diverge (world.forked is the ceremony's own marker), got %+v", r.Divergence)
	}
	out := renderDuelReport(r)
	if !strings.Contains(out, "aria-b forked from aria at day 1") {
		t.Errorf("lineage header missing:\n%s", out)
	}
	if !strings.Contains(out, "the two runs are identical since the fork — the change made no observable difference yet") {
		t.Errorf("the honest zero-divergence line missing:\n%s", out)
	}

	// --since overrides the lineage window.
	r2, err := buildDuelReport(parentDir, forkDir, 0)
	if err != nil {
		t.Fatal(err)
	}
	if r2.Since != 0 {
		t.Errorf("--since 0 should override the lineage window, got %d", r2.Since)
	}
}

// TestCompareInterleaveOrderAndLabels is SC-005's interleave half: entries
// merge by from_tick with per-world labels, ties stable (A before B), and
// the divergence marker sits in timeline position.
func TestCompareInterleaveOrderAndLabels(t *testing.T) {
	entry := func(fromTick int64, text string) store.Event {
		return store.Event{Tick: fromTick + 10, Type: "chronicle.entry", Payload: duelJSON(t, sim.ChronicleEntryPayload{
			Day: 1, FromTick: fromTick, ToTick: fromTick + 10, Text: text})}
	}
	a := duelWorld(t, "tale-a", "", []store.Event{
		entry(1000, "first light"),
		{Tick: 1600, Type: "agent.moved", Payload: duelJSON(t, sim.AgentMovedPayload{Agent: sim.Ref(0), X: 1, Y: 1})},
		entry(2000, "the gathering"),
	})
	b := duelWorld(t, "tale-b", "", []store.Event{
		entry(1500, "a stranger arrives"),
		{Tick: 1600, Type: "agent.foraged", Payload: duelJSON(t, map[string]any{"agent": 0, "x": 1, "y": 1})},
	})
	r, err := buildDuelReport(a, b, 500)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Entries) != 3 {
		t.Fatalf("interleave = %d entries, want 3: %+v", len(r.Entries), r.Entries)
	}
	wantOrder := []struct{ world, text string }{
		{"tale-a", "first light"},
		{"tale-b", "a stranger arrives"},
		{"tale-a", "the gathering"},
	}
	for i, want := range wantOrder {
		if r.Entries[i].World != want.world || r.Entries[i].Text != want.text {
			t.Errorf("entry %d = [%s] %q, want [%s] %q", i, r.Entries[i].World, r.Entries[i].Text, want.world, want.text)
		}
	}
	out := renderDuelReport(r)
	if !strings.Contains(out, "[tale-a] day 1 — first light") || !strings.Contains(out, "[tale-b] day 1 — a stranger arrives") {
		t.Errorf("interleave labels missing:\n%s", out)
	}
	// Marker position: after the from_tick-1500 entry (divergence at 1600),
	// before the from_tick-2000 entry.
	markerIdx := strings.Index(out, "the stories diverge here")
	strangerIdx := strings.Index(out, "a stranger arrives")
	gatheringIdx := strings.Index(out, "the gathering")
	if markerIdx < 0 || !(strangerIdx < markerIdx && markerIdx < gatheringIdx) {
		t.Errorf("divergence marker out of timeline position (stranger=%d marker=%d gathering=%d):\n%s",
			strangerIdx, markerIdx, gatheringIdx, out)
	}
}
