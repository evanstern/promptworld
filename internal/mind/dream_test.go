package mind

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// Spec 098 (private dreams) — the mind-side seam: the consolidation worker
// runs the geometry pass first and independently, and consults the existing
// consolidation slot for the ambiguous band only (D2), landing every outcome
// as recorded events (D3).

// dreamVec returns a unit 3-vector at cosine cos(theta) from dreamVec(0).
func dreamVec(theta float64) []float32 {
	return []float32{float32(math.Cos(theta)), float32(math.Sin(theta)), 0}
}

// setupDreamConsol is setupConsol with a vectored memory store: the three
// base memories plus a six-member cluster at the given angle from its leader.
func setupDreamConsol(t *testing.T, model Submitter, theta float64) (*harness, *Mind) {
	t.Helper()
	h := newHarness(t, "")
	if _, err := h.loop.Do("pause", ""); err != nil {
		t.Fatal(err)
	}
	m := h.m
	state := sim.NewState(42, m)
	mems := []sim.Memory{
		{Text: "Saw a wolf at the treeline.", Salience: 7, Tick: 100, Subject: -1},
		{Text: "Ate two berries.", Salience: 1, Tick: 200, Subject: -1},
		{Text: "Cedar promised me firewood.", Salience: 5, Tick: 300, Subject: 2, Tone: 20},
	}
	for i, tick := range []int64{1000, 2000, 3000, 4000, 5000, 6000} {
		vec := dreamVec(theta)
		if i == 0 {
			vec = dreamVec(0) // the leader
		}
		mems = append(mems, sim.Memory{
			Text: "Foraged the berry patch " + string(rune('a'+i)) + ".", Salience: 4,
			Tick: tick, Subject: -1, Vec: vec, VecModel: "test-embed",
		})
	}
	state.Agents[0].Memories = mems

	md, err := New(model, h.loop, h.loop, m, 42, state.Marshal(), [sim.AgentCount]string{}, testLoopRounds, testPlannerTokens, testConsolidationTokens, "", noopLoop)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(md.Close)
	return h, md
}

// TestDreamGeometryLandsDespiteRejectedNight (US1-AC1/AC2 + D3): a clear-cut
// cluster (identical vectors) is decided by geometry ALONE — its recorded
// habituation/merge batch lands even when the night's LLM output is garbage
// and the consolidation itself is rejected.
func TestDreamGeometryLandsDespiteRejectedNight(t *testing.T) {
	h, md := setupDreamConsol(t, &scriptedModel{replies: []string{"the model hums, no json"}}, 0)

	md.maybeConsolidate(sleptEvent(t, 80000, 0))

	markers := h.waitEvents(t, 10*time.Second, func(e store.Event) bool {
		return e.Type == "agent.consolidated"
	})
	if len(markers) != 1 {
		t.Fatalf("markers = %d, want 1", len(markers))
	}
	var mp sim.ConsolidatedPayload
	json.Unmarshal(markers[0].Payload, &mp)
	if mp.Outcome != sim.ConsolidationRejected {
		t.Fatalf("marker outcome = %q, want rejected", mp.Outcome)
	}

	all, _ := h.st.EventsSince(0, 0)
	counts := map[string]int{}
	for _, e := range all {
		counts[e.Type]++
	}
	// Six identical-vector members, default dials (cap 4, habituation 500):
	// newest kept vivid, four oldest merged, one habituated 4 → 2.
	if counts["agent.memory_merged"] != 1 || counts["agent.salience_revised"] != 1 {
		t.Fatalf("geometry batch = merged %d revised %d, want 1 and 1 (counts %v)",
			counts["agent.memory_merged"], counts["agent.salience_revised"], counts)
	}
}

// dreamConsolidationReply is goodConsolidation plus a routine verdict: g1
// folds; g9 and a mangled label are mechanical slack, coerced away.
func dreamConsolidationReply() string {
	return strings.Replace(goodConsolidation(), `"narrative":`,
		`"routine": ["g1", "g9", "berries"], "narrative":`, 1)
}

// TestDreamAmbiguousConsultFolds (US1-AC2, D2/D3): an in-band cluster is NOT
// decided by geometry; it rides the existing consolidation call as a labeled
// group, and the model's fold verdict lands the precomputed merge/habituation
// events in the accepted atomic batch, with the keep/fold counts on the
// marker.
func TestDreamAmbiguousConsultFolds(t *testing.T) {
	inBand := math.Acos(0.90) // cohesion 0.90 ± 0.015 jitter: always ambiguous
	h, md := setupDreamConsol(t, &scriptedModel{replies: []string{dreamConsolidationReply()}}, inBand)

	md.maybeConsolidate(sleptEvent(t, 80000, 0))

	markers := h.waitEvents(t, 10*time.Second, func(e store.Event) bool {
		return e.Type == "agent.consolidated"
	})
	if len(markers) != 1 {
		t.Fatalf("markers = %d, want 1", len(markers))
	}
	var mp sim.ConsolidatedPayload
	json.Unmarshal(markers[0].Payload, &mp)
	if mp.Outcome != sim.ConsolidationAccepted {
		t.Fatalf("marker = %+v, want accepted", mp)
	}
	if mp.DreamFolded != 1 || mp.DreamKept != 0 {
		t.Errorf("marker dream counts = folded %d kept %d, want 1 and 0", mp.DreamFolded, mp.DreamKept)
	}

	all, _ := h.st.EventsSince(0, 0)
	counts := map[string]int{}
	for _, e := range all {
		counts[e.Type]++
	}
	if counts["agent.memory_merged"] != 1 || counts["agent.salience_revised"] != 1 {
		t.Fatalf("fold batch = merged %d revised %d, want 1 and 1 (counts %v)",
			counts["agent.memory_merged"], counts["agent.salience_revised"], counts)
	}
}

// TestDreamConsultPromptShape: the ambiguous block and the routine contract
// line appear ONLY when groups ride the call — a group-less night's prompt is
// byte-identical to the pre-098 prompt.
func TestDreamConsultPromptShape(t *testing.T) {
	job := consolJob{
		agent: 0, name: "Ash", anchor: persona.Anchors["Ash"],
		buffer: []sim.Memory{{Text: "Saw a wolf.", Salience: 7, Tick: 100, Subject: -1}},
	}
	bare := consolidateUserPrompt(job, nil)
	if strings.Contains(bare, "routine") || strings.Contains(bare, "[g1]") {
		t.Fatal("group-less prompt mentions the dream consult")
	}
	groups := []sim.DreamGroup{{Size: 6, Examples: []string{"Foraged the berry patch a.", "Foraged the berry patch b."}}}
	withGroups := consolidateUserPrompt(job, groups)
	for _, wantSub := range []string{"[g1] 6 similar memories", `"Foraged the berry patch a."`, `"routine"`} {
		if !strings.Contains(withGroups, wantSub) {
			t.Errorf("consult prompt missing %q", wantSub)
		}
	}
}

// TestParseRoutineRefs: mechanical slack — unknown, mangled, duplicate, and
// out-of-range labels are dropped, never a rejection.
func TestParseRoutineRefs(t *testing.T) {
	got := parseRoutineRefs([]string{"g2", "G1", "g2", "g9", "m1", "", "gx"}, 3)
	want := []int{1, 0}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("parseRoutineRefs = %v, want %v", got, want)
	}
	if out := parseRoutineRefs(nil, 0); out != nil {
		t.Errorf("nil refs parsed to %v", out)
	}
}
