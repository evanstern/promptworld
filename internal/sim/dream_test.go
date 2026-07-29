package sim

import (
	"bytes"
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// Spec 098 (private dreams) test suite — FR-006:
//   - privacy perturbation independence (US2, SC-002 — the crown jewel),
//   - habituation/merge correctness (US1),
//   - geometry-vs-LLM routing (D2),
//   - replay byte-identity (US1-AC3, SC-003),
//   - noise-zeroed equivalence + cross-seed variance (D4).

// dreamTestDials is the doctrine default set spelled out, so these tests pin
// intent independently of the tuning constants.
func dreamTestDials() DreamTuning {
	return DreamTuning{
		DensityPerMille:       900,
		AmbiguousBandPerMille: 30,
		HabituationPerMille:   500,
		MergeCapPerNight:      4,
		JitterPerMille:        0, // most tests want exact geometry; noise tests set it
	}
}

// vecAt returns a unit 3-vector whose cosine against vecAt(0) — (1,0,0) — is
// exactly cos(theta): the test's dial for cluster cohesion.
func vecAt(theta float64) []float32 {
	return []float32{float32(math.Cos(theta)), float32(math.Sin(theta)), 0}
}

// dreamMem builds one vectored memory.
func dreamMem(text string, sal int, tick int64, vec []float32) Memory {
	return Memory{Text: text, Salience: sal, Tick: tick, Subject: -1, Vec: vec, VecModel: "test-model"}
}

// TestPlanDreamClearClusterHabituatesAndMerges (US1-AC1): a dense
// near-duplicate cluster collapses by geometry alone — the newest
// most-salient member stays vivid, the oldest members merge under the cap,
// the rest habituate — while distinct and vectorless memories are untouched.
func TestPlanDreamClearClusterHabituatesAndMerges(t *testing.T) {
	same := vecAt(0) // identical vectors: cohesion 1.0, clearly above 0.93
	mems := []Memory{
		dreamMem("berries 1", 4, 1000, same),
		dreamMem("berries 2", 4, 2000, same),
		dreamMem("berries 3", 4, 3000, same),
		dreamMem("berries 4", 4, 4000, same),
		dreamMem("berries 5", 4, 5000, same),
		dreamMem("berries 6", 4, 6000, same),
		dreamMem("the gru chased me", 9, 3500, vecAt(math.Pi/2)), // orthogonal: distinct
		{Text: "vectorless old memory", Salience: 5, Tick: 50, Subject: -1},
	}
	d := dreamTestDials()
	d.MergeCapPerNight = 2
	plan := PlanDream(mems, 42, 1, 0, d)

	if len(plan.Ambiguous) != 0 {
		t.Fatalf("clear-cut cluster routed to the consult: %+v", plan.Ambiguous)
	}
	// Representative: equal salience → newest tick wins (berries 6).
	if len(plan.Merges) != 1 {
		t.Fatalf("merges = %d, want 1", len(plan.Merges))
	}
	mg := plan.Merges[0]
	if mg.Kept != (MemoryRef{Tick: 6000, Hash: MemoryHash("berries 6")}) || mg.Salience != 4 {
		t.Errorf("kept = %+v sal %d, want berries 6 at 4", mg.Kept, mg.Salience)
	}
	// Cap 2: the two oldest non-representatives merge.
	wantMerged := []MemoryRef{
		{Tick: 1000, Hash: MemoryHash("berries 1")},
		{Tick: 2000, Hash: MemoryHash("berries 2")},
	}
	if !reflect.DeepEqual(mg.Merged, wantMerged) {
		t.Errorf("merged = %+v, want %+v", mg.Merged, wantMerged)
	}
	// The three surviving non-representatives habituate 4 → 2.
	if len(plan.Revisions) != 3 {
		t.Fatalf("revisions = %d, want 3: %+v", len(plan.Revisions), plan.Revisions)
	}
	for i, wantTick := range []int64{3000, 4000, 5000} {
		rv := plan.Revisions[i]
		if rv.Ref.Tick != wantTick || rv.Salience != 2 {
			t.Errorf("revision %d = %+v, want tick %d sal 2", i, rv, wantTick)
		}
	}
	// The distinct and vectorless memories appear nowhere.
	for _, rv := range plan.Revisions {
		if rv.Ref.Tick == 3500 || rv.Ref.Tick == 50 {
			t.Errorf("distinct/vectorless memory touched: %+v", rv)
		}
	}
}

// TestPlanDreamRouting (US1-AC2, D2): clearly-dense clusters decide by
// geometry (no consult), an in-band cluster routes to the consult with its
// fold outcome precomputed, and scattered memories produce nothing.
func TestPlanDreamRouting(t *testing.T) {
	d := dreamTestDials() // member bar 0.87, clear bar 0.93
	inBand := math.Acos(0.90)

	t.Run("ambiguous band routes to the consult", func(t *testing.T) {
		mems := []Memory{
			dreamMem("routine 1", 3, 400, vecAt(0)), // leader
			dreamMem("routine 2", 3, 500, vecAt(inBand)),
			dreamMem("routine 3", 3, 600, vecAt(inBand)),
			dreamMem("routine 4", 3, 700, vecAt(inBand)),
		}
		plan := PlanDream(mems, 42, 1, 0, d)
		if len(plan.Revisions) != 0 || len(plan.Merges) != 0 {
			t.Fatalf("in-band cluster decided by geometry alone: %+v", plan)
		}
		if len(plan.Ambiguous) != 1 {
			t.Fatalf("ambiguous groups = %d, want 1", len(plan.Ambiguous))
		}
		g := plan.Ambiguous[0]
		if g.Size != 4 || len(g.Examples) != dreamGroupExamples {
			t.Errorf("group size %d examples %d, want 4 and %d", g.Size, len(g.Examples), dreamGroupExamples)
		}
		// Fold outcome precomputed: rep = newest (routine 4), 3 merged under
		// the cap of 4, no survivors left to habituate.
		if g.Merge.Kept.Tick != 700 || len(g.Merge.Merged) != 3 || len(g.Revisions) != 0 {
			t.Errorf("precomputed fold = %+v", g)
		}
	})

	t.Run("scattered memories produce nothing", func(t *testing.T) {
		mems := []Memory{
			dreamMem("a", 3, 100, vecAt(0)),
			dreamMem("b", 3, 200, vecAt(math.Pi/3)),
			dreamMem("c", 3, 300, vecAt(2*math.Pi/3)),
			dreamMem("d", 3, 400, vecAt(math.Pi)),
		}
		plan := PlanDream(mems, 42, 1, 0, d)
		if len(plan.Revisions)+len(plan.Merges)+len(plan.Ambiguous) != 0 {
			t.Fatalf("scattered store dreamed something: %+v", plan)
		}
	})

	t.Run("cross-model vectors never cluster", func(t *testing.T) {
		mems := []Memory{
			dreamMem("x1", 3, 100, vecAt(0)),
			dreamMem("x2", 3, 200, vecAt(0)),
			dreamMem("x3", 3, 300, vecAt(0)),
		}
		mems[2].VecModel = "other-model"
		plan := PlanDream(mems, 42, 1, 0, d)
		if len(plan.Revisions)+len(plan.Merges)+len(plan.Ambiguous) != 0 {
			t.Fatalf("cross-model memories clustered: %+v", plan)
		}
	})
}

// TestPlanDreamPrivacyPerturbation is SC-002, the card's AC#1 (US2): perturb
// one agent's memories arbitrarily and the OTHER agent's consolidation —
// plan and reduced post-night state — is byte-identical. Privacy holds by
// construction: the pass's only input is the one agent's store.
func TestPlanDreamPrivacyPerturbation(t *testing.T) {
	seed := uint64(7)
	m := testMap(seed)
	d := dreamTestDials()
	d.JitterPerMille = 15 // noise on: privacy must hold with the dial live too

	overlap := func() []Memory {
		return []Memory{
			dreamMem("foraged the patch 1", 4, 1000, vecAt(0)),
			dreamMem("foraged the patch 2", 4, 2000, vecAt(0)),
			dreamMem("foraged the patch 3", 4, 3000, vecAt(0)),
			dreamMem("foraged the patch 4", 4, 4000, vecAt(0)),
			dreamMem("met the stranger", 8, 3500, vecAt(math.Pi/2)),
		}
	}

	build := func(perturbB bool) (*State, DreamPlan) {
		s := NewState(seed, m)
		s.Agents[0].Memories = overlap()
		s.Agents[1].Memories = overlap() // substantially overlapping experiences
		if perturbB {
			// Perturb agent 1 every way that could plausibly leak: vectors,
			// saliences, membership, extra dense mass.
			s.Agents[1].Memories[0].Vec = vecAt(math.Pi / 5)
			s.Agents[1].Memories[2].Salience = 9
			s.Agents[1].Memories = append(s.Agents[1].Memories,
				dreamMem("foraged the patch 5", 4, 5000, vecAt(0)),
				dreamMem("foraged the patch 6", 4, 6000, vecAt(0)))
		}
		// Agent 0's pass reads agent 0's store — the seam the mind driver
		// snapshots (consolJob.mems).
		plan := PlanDream(s.Agents[0].Memories, seed, 1, 0, d)
		for _, e := range DreamEvents(0, plan.Revisions, plan.Merges) {
			e.Tick = 80000
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		return s, plan
	}

	sBase, planBase := build(false)
	sPert, planPert := build(true)

	if !reflect.DeepEqual(planBase, planPert) {
		t.Fatalf("agent 0's dream plan changed when agent 1's memories were perturbed:\n%+v\nvs\n%+v", planBase, planPert)
	}
	a0Base, _ := json.Marshal(sBase.Agents[0])
	a0Pert, _ := json.Marshal(sPert.Agents[0])
	if !bytes.Equal(a0Base, a0Pert) {
		t.Fatal("agent 0's post-night state changed when agent 1's memories were perturbed")
	}
	// Sanity: the pass actually did something for agent 0 (a vacuous zero
	// outcome would prove nothing).
	if len(planBase.Revisions)+len(planBase.Merges) == 0 {
		t.Fatal("agent 0's dream plan was empty — the perturbation proof is vacuous")
	}
}

// TestDreamNoiseZeroedEquivalence (SC-003, D4): with the jitter dial at 0 the
// pass is a pure function of the store and dials — the seed contributes
// NOTHING, reproducing pre-noise outcomes exactly.
func TestDreamNoiseZeroedEquivalence(t *testing.T) {
	// Borderline cluster: cohesion 0.92, near the 0.93 clear bar — exactly
	// where jitter WOULD matter if it were live.
	edge := math.Acos(0.92)
	mems := []Memory{
		dreamMem("edge 1", 3, 100, vecAt(0)),
		dreamMem("edge 2", 3, 200, vecAt(edge)),
		dreamMem("edge 3", 3, 300, vecAt(edge)),
		dreamMem("edge 4", 3, 400, vecAt(edge)),
	}
	d := dreamTestDials() // JitterPerMille: 0
	base := PlanDream(mems, 1, 1, 0, d)
	for seed := uint64(2); seed <= 32; seed++ {
		if got := PlanDream(mems, seed, 1, 0, d); !reflect.DeepEqual(got, base) {
			t.Fatalf("seed %d changed a zero-noise plan", seed)
		}
	}
}

// TestDreamNoiseVariesAcrossSeeds (D4): with the dial up, a borderline
// cluster classifies differently across seeds — dream-like variance between
// runs — while any single (seed, night, agent) triple stays deterministic.
func TestDreamNoiseVariesAcrossSeeds(t *testing.T) {
	edge := math.Acos(0.92) // 0.92 ± 0.2 straddles the 0.93 clear bar
	mems := []Memory{
		dreamMem("edge 1", 3, 100, vecAt(0)),
		dreamMem("edge 2", 3, 200, vecAt(edge)),
		dreamMem("edge 3", 3, 300, vecAt(edge)),
		dreamMem("edge 4", 3, 400, vecAt(edge)),
	}
	d := dreamTestDials()
	d.JitterPerMille = 200
	outcomes := map[bool]int{} // true = decided clear, false = ambiguous
	for seed := uint64(1); seed <= 64; seed++ {
		plan := PlanDream(mems, seed, 1, 0, d)
		outcomes[len(plan.Ambiguous) == 0]++
		// Determinism per triple: the same coordinates always agree.
		if again := PlanDream(mems, seed, 1, 0, d); !reflect.DeepEqual(again, plan) {
			t.Fatalf("seed %d: two calls disagreed", seed)
		}
	}
	if outcomes[true] == 0 || outcomes[false] == 0 {
		t.Fatalf("no cross-seed variance at the boundary: %v", outcomes)
	}
}

// TestApplyDreamReducer: the two arms apply recorded outcomes verbatim
// (clamped), and are total — vanished targets no-op, never error.
func TestApplyDreamReducer(t *testing.T) {
	seed := uint64(9)
	m := testMap(seed)
	s := NewState(seed, m)
	s.Agents[0].Memories = []Memory{
		{Text: "one", Salience: 6, Tick: 100, Subject: -1},
		{Text: "two", Salience: 6, Tick: 200, Subject: -1},
		{Text: "three", Salience: 6, Tick: 300, Subject: -1},
	}
	ev := func(typ string, p any) store.Event {
		b, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		return store.Event{Tick: 80000, Type: typ, Payload: b}
	}

	// Salience revision applies the absolute recorded value.
	if err := s.Apply(ev("agent.salience_revised", SalienceRevisedPayload{
		Agent: Ref(0), MemTick: 100, TextHash: MemoryHash("one"), Salience: 3, Reason: DreamReasonHabituation})); err != nil {
		t.Fatal(err)
	}
	if s.Agents[0].Memories[0].Salience != 3 {
		t.Errorf("salience = %d, want 3", s.Agents[0].Memories[0].Salience)
	}
	// Out-of-range recorded values clamp defensively.
	if err := s.Apply(ev("agent.salience_revised", SalienceRevisedPayload{
		Agent: Ref(0), MemTick: 100, TextHash: MemoryHash("one"), Salience: 99})); err != nil {
		t.Fatal(err)
	}
	if s.Agents[0].Memories[0].Salience != MaxSalience {
		t.Errorf("clamped salience = %d, want %d", s.Agents[0].Memories[0].Salience, MaxSalience)
	}
	// Vanished target (wrong hash): no-op, no error.
	if err := s.Apply(ev("agent.salience_revised", SalienceRevisedPayload{
		Agent: Ref(0), MemTick: 200, TextHash: "ffffffff", Salience: 1})); err != nil {
		t.Fatal(err)
	}
	if s.Agents[0].Memories[1].Salience != 6 {
		t.Errorf("vanished-target revision mutated state")
	}

	// Merge removes the absorbed members and sets the kept salience; a
	// vanished member no-ops while the rest of the batch still applies.
	if err := s.Apply(ev("agent.memory_merged", MemoryMergedPayload{
		Agent: Ref(0),
		Kept:  MemoryRef{Tick: 300, Hash: MemoryHash("three")},
		Merged: []MemoryRef{
			{Tick: 100, Hash: MemoryHash("one")},
			{Tick: 999, Hash: "ffffffff"}, // vanished
		},
		Salience: 7})); err != nil {
		t.Fatal(err)
	}
	a := &s.Agents[0]
	if len(a.Memories) != 2 {
		t.Fatalf("memories = %d, want 2 (one removed, vanished ref ignored)", len(a.Memories))
	}
	if a.Memories[0].Text != "two" || a.Memories[1].Text != "three" || a.Memories[1].Salience != 7 {
		t.Errorf("post-merge store = %+v", a.Memories)
	}
}

// TestDreamReplayByteIdentity (US1-AC3, SC-003): a timeline carrying both
// dream event types replays to byte-identical state and survives the
// snapshot round-trip — replay applies recorded outcomes with zero
// re-derivation and zero model calls.
func TestDreamReplayByteIdentity(t *testing.T) {
	seed := uint64(13)
	m := testMap(seed)
	mkEvents := func(t *testing.T) []store.Event {
		ev := func(tick int64, typ string, p any) store.Event {
			b, err := json.Marshal(p)
			if err != nil {
				t.Fatal(err)
			}
			return store.Event{Tick: tick, Type: typ, Payload: b}
		}
		return []store.Event{
			ev(10, "agent.memory_added", MemoryAddedPayload{Agent: Ref(0), Text: "picked berries", Salience: 4, Subject: Ref(-1)}),
			ev(20, "agent.memory_added", MemoryAddedPayload{Agent: Ref(0), Text: "picked berries again", Salience: 4, Subject: Ref(-1)}),
			ev(30, "agent.memory_added", MemoryAddedPayload{Agent: Ref(0), Text: "picked berries once more", Salience: 4, Subject: Ref(-1)}),
			ev(80000, "agent.memory_merged", MemoryMergedPayload{
				Agent: Ref(0), Kept: MemoryRef{Tick: 30, Hash: MemoryHash("picked berries once more")},
				Merged:   []MemoryRef{{Tick: 10, Hash: MemoryHash("picked berries")}},
				Salience: 4}),
			ev(80001, "agent.salience_revised", SalienceRevisedPayload{
				Agent: Ref(0), MemTick: 20, TextHash: MemoryHash("picked berries again"),
				Salience: 2, Reason: DreamReasonHabituation}),
			ev(80002, "agent.consolidated", ConsolidatedPayload{
				Agent: Ref(0), Night: 1, UpTo: 30, Outcome: ConsolidationAccepted,
				DreamFolded: 1, DreamKept: 1}),
		}
	}

	run := func() *State {
		s := NewState(seed, m)
		for _, e := range mkEvents(t) {
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		s.Tick = 80002
		return s
	}
	live, replayed := run(), run()
	if !bytes.Equal(live.Marshal(), replayed.Marshal()) {
		t.Error("replayed state differs from live state")
	}
	if live.Hash() != replayed.Hash() {
		t.Error("replay hash diverged")
	}
	var thawed State
	if err := json.Unmarshal(live.Marshal(), &thawed); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(live.Marshal(), thawed.Marshal()) {
		t.Error("snapshot round-trip changed state")
	}
}

// TestParseTuningDreamDials (US3-AC1): the five spec-098 dials ride the
// spec-048 manifest — sparse keys resolve against defaults, out-of-range
// values clamp with a warning, absent keys leave the block nil ≡ defaults,
// and Equal compares by resolved value.
func TestParseTuningDreamDials(t *testing.T) {
	parsed, warns, err := ParseTuning([]byte(`{
		"dream_density_per_mille": 950,
		"dream_jitter_per_mille": 999
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Dream == nil {
		t.Fatal("dream keys present but block nil")
	}
	want := defaultDream()
	want.DensityPerMille = 950
	want.JitterPerMille = maxDreamJitterPerMille // clamped
	if *parsed.Dream != want {
		t.Errorf("resolved dream = %+v, want %+v", *parsed.Dream, want)
	}
	if len(warns) != 1 || !strings.Contains(warns[0], "dream_jitter_per_mille") {
		t.Errorf("warns = %v, want one jitter clamp warning", warns)
	}

	// No dream keys → nil block, and Equal treats nil as the default set.
	parsed2, _, err := ParseTuning([]byte(`{"fire_burn_per_wood": 7200}`))
	if err != nil {
		t.Fatal(err)
	}
	if parsed2.Dream != nil {
		t.Errorf("dream block resolved without any dream key: %+v", parsed2.Dream)
	}
	explicit := *parsed2
	d := defaultDream()
	explicit.Dream = &d
	if !parsed2.Equal(explicit) {
		t.Error("nil dream block != explicit default block under Equal")
	}

	// The dial reaches the pass: event round-trip through the reducer.
	seed := uint64(3)
	s := NewState(seed, testMap(seed))
	if s.DreamDials() != defaultDream() {
		t.Errorf("untuned DreamDials = %+v, want defaults", s.DreamDials())
	}
	if err := s.Apply(NewTuningEvent(0, *parsed)); err != nil {
		t.Fatal(err)
	}
	if got := s.DreamDials(); got != want {
		t.Errorf("post-apply DreamDials = %+v, want %+v", got, want)
	}
}
