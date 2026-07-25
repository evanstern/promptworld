package sim

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

func memAt(tick int64, sal int, text string) Memory {
	return Memory{Text: text, Salience: sal, Tick: tick}
}

// TestWindowBound is AC#3's core: the window never exceeds K, no matter how
// large the soul.
func TestWindowBound(t *testing.T) {
	a := &Agent{Name: "Ash"}
	for i := int64(0); i < 500; i++ {
		a.Memories = append(a.Memories, memAt(i*60, 1+int(i%10), "m"))
	}
	w := SelectMemories(a, 42, 0, 500*60, WindowK)
	if len(w) != WindowK {
		t.Fatalf("window = %d memories, want exactly %d", len(w), WindowK)
	}
}

// TestWindowDeterministic: same state + tick → identical selection.
func TestWindowDeterministic(t *testing.T) {
	a := &Agent{Name: "Ash"}
	for i := int64(0); i < 100; i++ {
		a.Memories = append(a.Memories, memAt(i*600, 1+int(i%10), "m"))
	}
	w1 := SelectMemories(a, 7, 3, 100_000, WindowK)
	w2 := SelectMemories(a, 7, 3, 100_000, WindowK)
	if len(w1) != len(w2) {
		t.Fatal("selection sizes differ")
	}
	for i := range w1 {
		// Memory carries a []float32 vector since spec 042, so struct equality
		// no longer compiles; these vectorless fixtures compare identity fields.
		if w1[i].Tick != w2[i].Tick || w1[i].Text != w2[i].Text || w1[i].Salience != w2[i].Salience {
			t.Fatalf("selection differs at %d: %+v vs %+v", i, w1[i], w2[i])
		}
	}
	// Same cadence bucket → same serendipity picks even at nearby ticks.
	w3 := SelectMemories(a, 7, 3, 100_000+60, WindowK)
	if len(w3) != len(w1) {
		t.Fatal("bucketed selection changed size within a cadence window")
	}
}

// TestWindowFavorsSalienceAndRecency: a fresh 10★ beats an old 1★; the
// serendipity quota still surfaces something old.
func TestWindowFavorsSalienceAndRecency(t *testing.T) {
	a := &Agent{Name: "Ash"}
	// 40 old low-salience, then recent high-salience ones.
	for i := int64(0); i < 40; i++ {
		a.Memories = append(a.Memories, memAt(i*60, 1, "old-noise"))
	}
	a.Memories = append(a.Memories,
		memAt(90_000, 10, "watched-death"),
		memAt(91_000, 9, "near-death"),
		memAt(92_000, 3, "talk"))
	w := SelectMemories(a, 42, 0, 93_000, WindowK)

	var texts []string
	for _, m := range w {
		texts = append(texts, m.Text)
	}
	joined := strings.Join(texts, ",")
	if !strings.Contains(joined, "watched-death") || !strings.Contains(joined, "near-death") {
		t.Errorf("high-salience recents missing: %v", texts)
	}
	if !strings.Contains(joined, "old-noise") {
		t.Errorf("serendipity tail pick missing: %v", texts)
	}
	// Reverse-chronological presentation.
	for i := 1; i < len(w); i++ {
		if w[i].Tick > w[i-1].Tick {
			t.Fatalf("window not reverse-chronological at %d", i)
		}
	}
}

// TestMemoriesAccrete: a running village generates memories via events, and
// they land in state (AC#2's accretion half at sim level).
func TestMemoriesAccrete(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	log := driveTicks(t, s, m, 12*3600, nil)

	var memEvents int
	for _, e := range log {
		if e.Type == "agent.memory_added" {
			memEvents++
			var p MemoryAddedPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				t.Fatal(err)
			}
			if p.Salience < 1 || p.Salience > 10 || p.Text == "" {
				t.Errorf("bad memory payload: %+v", p)
			}
		}
	}
	if memEvents == 0 {
		t.Fatal("half a game-day produced no memories (fires/talks should mark souls)")
	}
	var inState int
	for _, a := range s.Agents {
		inState += len(a.Memories)
	}
	if inState != memEvents {
		t.Errorf("state carries %d memories, log emitted %d", inState, memEvents)
	}
}

// TestReflexGrace: idle agents act only after the grace window, and
// IdleSince tracks intent completion.
func TestReflexGrace(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)

	// From genesis (IdleSince 0), nothing may happen before the grace.
	log := driveTicks(t, s, m, reflexGraceTicks-1, nil)
	for _, e := range log {
		if e.Type == "agent.intent_set" {
			t.Fatalf("reflex acted at tick %d, inside the grace window", e.Tick)
		}
	}
	// Shortly after the grace, reflexes fire (staggered).
	log = driveTicks(t, s, m, reflexGraceTicks+40, nil)
	var acted bool
	for _, e := range log {
		if e.Type == "agent.intent_set" {
			var p IntentSetPayload
			json.Unmarshal(e.Payload, &p)
			if p.Source != "reflex" {
				t.Errorf("expected reflex source, got %q", p.Source)
			}
			acted = true
		}
	}
	if !acted {
		t.Fatal("reflex never fired after the grace window")
	}
}

// TestInjectedPlannerIntent: a planner-style command timeline (intent_set
// source planner + thought) replays deterministically like any input, and
// the executor acts on it.
func TestInjectedPlannerIntent(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)

	// Compute what the resolver would emit for agent 0 "forage" at tick 30.
	pre := NewState(seed, m)
	driveTicks(t, pre, m, 30, nil)
	intent, direct, err := resolveGoal(pre, m, 0, "forage", -1, "", 0, 30)
	if err != nil || direct != "" || intent == nil {
		t.Fatalf("resolveGoal: %v %q %v", err, direct, intent)
	}

	timeline := map[int64][]store.Event{
		30: {
			{Tick: 30, Type: "agent.thought", Payload: mustPayload(ThoughtPayload{Agent: 0, Text: "The bushes call to me.", Source: "planner"})},
			{Tick: 30, Type: "agent.intent_set", Payload: mustPayload(IntentSetPayload{
				Agent: 0, Goal: intent.Goal, TargetX: intent.TargetX, TargetY: intent.TargetY, Source: "planner"})},
		},
	}
	logA := driveTicks(t, s, m, 3600, timeline)

	// The injected intent leads to actual foraging by agent 0.
	var foraged bool
	for _, e := range logA {
		if e.Type == "agent.foraged" {
			var p HarvestPayload
			json.Unmarshal(e.Payload, &p)
			if p.Agent == 0 {
				foraged = true
			}
		}
	}
	if !foraged {
		t.Fatal("planner-injected forage intent never completed")
	}

	// Determinism with injections in the timeline (SC-005).
	s2 := NewState(seed, m)
	logB := driveTicks(t, s2, m, 3600, timeline)
	if s.Hash() != s2.Hash() {
		t.Fatal("state hash diverged with identical injected timeline")
	}
	if len(logA) != len(logB) {
		t.Fatal("event count diverged with identical injected timeline")
	}
}

// TestResolveGoalErrors: impossible goals are refused with no event.
func TestResolveGoalErrors(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	sightAll(s, 0) // spec 041: talk_to resolves from the actor's sightings

	if _, _, err := resolveGoal(s, m, 0, "summon_gru", -1, "", 0, 0); err == nil {
		t.Error("unknown goal should error")
	}
	if _, _, err := resolveGoal(s, m, 0, "eat", -1, "", 0, 0); err == nil {
		t.Error("eat with no food should error")
	}
	if _, _, err := resolveGoal(s, m, 0, "build_fire", -1, "", 0, 0); err == nil {
		t.Error("build_fire with no wood should error")
	}
	if _, _, err := resolveGoal(s, m, 0, "talk_to", 0, "", 0, 0); err == nil {
		t.Error("talking to yourself should error")
	}
	if in, _, err := resolveGoal(s, m, 0, "talk_to", 1, "", 0, 0); err != nil || in.Goal != "seek" {
		t.Errorf("talk_to should resolve to seek: %v %v", in, err)
	}
}

// --- spec 042 T014: SelectMemoriesRelevant scorer tests ---

// memVec builds a vectored memory in the given model.
func memVec(tick int64, sal int, text string, vec []float32, model string) Memory {
	return Memory{Text: text, Salience: sal, Tick: tick, Vec: vec, VecModel: model}
}

// windowsEqual compares two selections field-by-field (Memory carries a slice,
// so struct equality does not compile).
func windowsEqual(a, b []Memory) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Tick != b[i].Tick || a[i].Text != b[i].Text || a[i].Salience != b[i].Salience || a[i].Seq != b[i].Seq {
			return false
		}
	}
	return true
}

// TestRelevantNilQueryDelegates (contract §2 row 1): a nil query — no
// situation vector recorded yet — returns SelectMemories' output verbatim.
func TestRelevantNilQueryDelegates(t *testing.T) {
	a := &Agent{Name: "Ash"}
	for i := int64(0); i < 60; i++ {
		a.Memories = append(a.Memories, memAt(i*600, 1+int(i%10), "m"))
	}
	legacy := SelectMemories(a, 7, 0, 40_000, WindowK)
	got := SelectMemoriesRelevant(a, 7, 0, 40_000, WindowK, nil, "all-minilm")
	if !windowsEqual(legacy, got) {
		t.Errorf("nil query must delegate to SelectMemories verbatim:\nlegacy %+v\ngot    %+v", legacy, got)
	}
}

// TestRelevantNeutralMatchesLegacy (contract §1 tail + §2 neutrality): when
// every memory scores the 0.5 neutral relevance (all vectorless here), the
// score is a monotonic transform of the legacy weight — head ordering, tie
// behavior, AND the seeded serendipity tail come out byte-identical to
// SelectMemories. This is the no-signal ⇒ no-behavior-change guarantee.
func TestRelevantNeutralMatchesLegacy(t *testing.T) {
	a := &Agent{Name: "Ash"}
	for i := int64(0); i < 80; i++ {
		a.Memories = append(a.Memories, memAt(i*700, 1+int(i%9), "m"))
	}
	query := []float32{1, 0, 0}
	for _, tick := range []int64{50_000, 60_000, 100_000} {
		legacy := SelectMemories(a, 42, 3, tick, WindowK)
		got := SelectMemoriesRelevant(a, 42, 3, tick, WindowK, query, "all-minilm")
		if !windowsEqual(legacy, got) {
			t.Errorf("tick %d: all-neutral relevance diverged from legacy:\nlegacy %+v\ngot    %+v", tick, legacy, got)
		}
	}
}

// TestRelevance01FallbackLadder (contract §2): every neutral row — vectorless,
// cross-model, zero-magnitude (either side), dimension mismatch — scores
// exactly 0.5; a comparable identical vector scores 1.0 and an opposite one 0.
func TestRelevance01FallbackLadder(t *testing.T) {
	q := []float32{1, 0}
	cases := []struct {
		name  string
		m     Memory
		query []float32
		want  float64
	}{
		{"vectorless", memAt(0, 5, "m"), q, 0.5},
		{"cross-model", memVec(0, 5, "m", []float32{1, 0}, "other-model"), q, 0.5},
		{"zero-magnitude memory", memVec(0, 5, "m", []float32{0, 0}, "all-minilm"), q, 0.5},
		{"zero-magnitude query", memVec(0, 5, "m", []float32{1, 0}, "all-minilm"), []float32{0, 0}, 0.5},
		{"dimension mismatch", memVec(0, 5, "m", []float32{1, 0, 0}, "all-minilm"), q, 0.5},
		{"identical vector", memVec(0, 5, "m", []float32{1, 0}, "all-minilm"), q, 1.0},
		{"opposite vector", memVec(0, 5, "m", []float32{-1, 0}, "all-minilm"), q, 0.0},
		{"orthogonal vector", memVec(0, 5, "m", []float32{0, 1}, "all-minilm"), q, 0.5},
	}
	for _, c := range cases {
		if got := relevance01(c.m, c.query, "all-minilm"); got != c.want {
			t.Errorf("%s: relevance01 = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestRelevantPromotesSituationMatch (the SC-004 mechanism at unit scale): an
// old, low-salience memory whose vector matches the query enters the relevant
// window while the legacy ranking provably excludes it; unrelated old
// memories stay out.
func TestRelevantPromotesSituationMatch(t *testing.T) {
	a := &Agent{Name: "Ash"}
	query := []float32{1, 0}
	// One old, low-salience, situation-matching memory…
	a.Memories = append(a.Memories, memVec(100, 2, "Robbed at the river.", []float32{1, 0}, "all-minilm"))
	// …buried under plenty of newer, louder, unrelated ones. Salience 5: the
	// equal-weight form's maximum relevance advantage is +0.5 — five decayed
	// salience points — so a perfect match (rel 1.0 vs the crowd's neutral-ish
	// 0.5) overcomes exactly this class of gap (contract §1 reference default).
	for i := int64(1); i <= 40; i++ {
		a.Memories = append(a.Memories, memVec(100_000+i*600, 5, "Routine day.", []float32{0, 1}, "all-minilm"))
	}
	tick := int64(130_000)

	contains := func(w []Memory, text string) bool {
		for _, m := range w {
			if m.Text == text {
				return true
			}
		}
		return false
	}
	// The legacy head (top k−2 by decayed salience) can never hold it: at this
	// age its weight is 0. Serendipity could luck into it, so assert on the
	// deterministic head by checking the legacy window at a bucket where the
	// seeded tail misses it.
	legacy := SelectMemories(a, 42, 0, tick, WindowK)
	if contains(legacy, "Robbed at the river.") {
		t.Skip("seeded serendipity tail surfaced the old memory in the legacy window at this bucket; scenario needs the deterministic-exclusion baseline")
	}
	got := SelectMemoriesRelevant(a, 42, 0, tick, WindowK, query, "all-minilm")
	if !contains(got, "Robbed at the river.") {
		t.Errorf("situation-matching old memory missing from the relevant window:\n%+v", got)
	}
}

// TestRelevantTiesNewerWins (contract §1): equal scores order newer-first —
// same salience, same age bucket, same relevance ⇒ the later tick leads.
func TestRelevantTiesNewerWins(t *testing.T) {
	a := &Agent{Name: "Ash"}
	// 20 identical-score memories (same salience, all inside one half-life,
	// all neutral relevance) at ascending ticks.
	for i := int64(0); i < 20; i++ {
		a.Memories = append(a.Memories, memAt(1000+i, 5, fmt.Sprintf("m%d", i)))
	}
	got := SelectMemoriesRelevant(a, 7, 0, 2000, WindowK, []float32{1}, "all-minilm")
	if len(got) == 0 || got[0].Tick != 1019 {
		t.Fatalf("tie-break did not favor the newest: head %+v", got[:1])
	}
}

// TestRelevantPure (FR-004): selection mutates nothing — the agent's memory
// list is byte-identical before and after, and two runs agree exactly.
func TestRelevantPure(t *testing.T) {
	a := &Agent{Name: "Ash"}
	for i := int64(0); i < 50; i++ {
		a.Memories = append(a.Memories, memVec(i*900, 1+int(i%10), "m", []float32{float32(i), 1}, "all-minilm"))
	}
	before, err := json.Marshal(a.Memories)
	if err != nil {
		t.Fatal(err)
	}
	q := []float32{3, 1}
	w1 := SelectMemoriesRelevant(a, 9, 1, 60_000, WindowK, q, "all-minilm")
	w2 := SelectMemoriesRelevant(a, 9, 1, 60_000, WindowK, q, "all-minilm")
	if !windowsEqual(w1, w2) {
		t.Errorf("two identical selections diverged:\n%+v\n%+v", w1, w2)
	}
	after, _ := json.Marshal(a.Memories)
	if string(before) != string(after) {
		t.Errorf("selection mutated memory state (FR-004):\nbefore %s\nafter  %s", before, after)
	}
}

// TestRelevantSmallSoulUnchanged (contract §2): n ≤ k returns every memory
// reverse-chronologically, exactly as SelectMemories does.
func TestRelevantSmallSoulUnchanged(t *testing.T) {
	a := &Agent{Name: "Ash"}
	for i := int64(0); i < 5; i++ {
		a.Memories = append(a.Memories, memAt(i*100, 3, fmt.Sprintf("m%d", i)))
	}
	got := SelectMemoriesRelevant(a, 7, 0, 1000, WindowK, []float32{1}, "all-minilm")
	if len(got) != 5 || got[0].Tick != 400 || got[4].Tick != 0 {
		t.Errorf("small soul must return all memories reverse-chronologically: %+v", got)
	}
}

// TestSC004RelevancePromotedMemory (spec 042 T020, SC-004): the acceptance
// scenario end-to-end at the selector — an old, low-salience,
// situation-matching memory enters the "on" window AND is provably absent
// from the legacy window (head by decayed-salience arithmetic, tail by the
// seeded serendipity draw at this exact seed/bucket — all deterministic, so
// the absence is a hard assertion, not luck).
func TestSC004RelevancePromotedMemory(t *testing.T) {
	a := &Agent{Name: "Ash"}
	query := []float32{1, 0}
	const marker = "Robbed at the river."
	a.Memories = append(a.Memories, memVec(100, 2, marker, []float32{1, 0}, "all-minilm"))
	for i := int64(1); i <= 40; i++ {
		a.Memories = append(a.Memories, memVec(100_000+i*600, 5, "Routine day.", []float32{0, 1}, "all-minilm"))
	}
	const seed, agentIdx, tick = 42, 0, 130_000

	contains := func(w []Memory) bool {
		for _, m := range w {
			if m.Text == marker {
				return true
			}
		}
		return false
	}
	legacy := SelectMemories(a, seed, agentIdx, tick, WindowK)
	if contains(legacy) {
		t.Fatalf("legacy window contains the old memory at this seed/bucket — pick fixture constants where the exclusion is provable:\n%+v", legacy)
	}
	on := SelectMemoriesRelevant(a, seed, agentIdx, tick, WindowK, query, "all-minilm")
	if !contains(on) {
		t.Fatalf("SC-004: the situation-matching memory never entered the relevance window:\n%+v", on)
	}
}

// TestSC006IsolationByConstruction (spec 042 T021, SC-006/FR-005): two agents
// with near-identical private memories; mutating agent A's memories leaves
// agent B's selection output byte-identical in EVERY mode — legacy, the
// shadow pair (both rankings), and "on". Selection's only memory source is
// the agent's own list, so cross-agent influence is impossible by
// construction; this proves it at the output bytes.
func TestSC006IsolationByConstruction(t *testing.T) {
	build := func() *State {
		s := NewState(42, testMap(42))
		s.Tick = 130_000
		for _, idx := range []int{0, 1} {
			ag := &s.Agents[idx]
			ag.Memories = append(ag.Memories, memVec(100, 2, "Robbed at the river.", []float32{1, 0}, "all-minilm"))
			for i := int64(1); i <= 30; i++ {
				ag.Memories = append(ag.Memories, memVec(100_000+i*600, 5, "Routine day.", []float32{0, 1}, "all-minilm"))
			}
			ag.SitVec = []float32{1, 0}
			ag.SitVecModel = "all-minilm"
			ag.SitVecTick = 129_000
		}
		return s
	}
	// Every selection surface B has, marshalled: legacy, relevant (B's own
	// query), and relevant with nil query (the fallback row).
	selectAllModes := func(s *State) string {
		b := &s.Agents[1]
		out := [][]Memory{
			SelectMemories(b, s.Seed, 1, s.Tick, WindowK),
			SelectMemoriesRelevant(b, s.Seed, 1, s.Tick, WindowK, b.SitVec, b.SitVecModel),
			SelectMemoriesRelevant(b, s.Seed, 1, s.Tick, WindowK, nil, ""),
		}
		j, err := json.Marshal(out)
		if err != nil {
			t.Fatal(err)
		}
		return string(j)
	}

	s := build()
	before := selectAllModes(s)

	// Mutate agent A every way that could plausibly leak: rewrite vectors to
	// match B's query perfectly, inflate salience, then drop memories entirely.
	aAg := &s.Agents[0]
	for i := range aAg.Memories {
		aAg.Memories[i].Vec = []float32{1, 0}
		aAg.Memories[i].Salience = 10
	}
	if got := selectAllModes(s); got != before {
		t.Fatal("mutating agent A's vectors/salience changed agent B's selection (SC-006 broken)")
	}
	aAg.Memories = nil
	aAg.SitVec = nil
	if got := selectAllModes(s); got != before {
		t.Fatal("removing agent A's memories changed agent B's selection (SC-006 broken)")
	}
}
