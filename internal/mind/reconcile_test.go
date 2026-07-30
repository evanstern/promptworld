package mind

import (
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Spec 097 (TASK-80, FR-003/FR-006): the mind-side belief reconciliation's
// three paths — confirmation boost, bounded disconfirmation decay (faster
// than silence), and silence untouched — plus the matcher's conservatism
// (out-of-disc coordinates, no coordinates, no recognized feature ⇒ nothing)
// and the single D4 surprise bump per disconfirming observation.

const (
	testConfirmBoost = 10
	testRetainPct    = 70
)

func obsPayload(x, y int, kinds ...string) sim.PlaceObservedPayload {
	if kinds == nil {
		kinds = []string{}
	}
	return sim.PlaceObservedPayload{Agent: sim.Ref(0), X: x, Y: y, Radius: 2, Kinds: kinds}
}

func testBelief(id int, statement string, conf int, reinforced int64) sim.Belief {
	return sim.Belief{ID: id, Statement: statement, Confidence: conf,
		Provenance: sim.ProvenanceTold, Source: -1, Subject: -1, Tick: reinforced, Reinforced: reinforced}
}

func decodeReinforced(t *testing.T, payload json.RawMessage) sim.BeliefReinforcedPayload {
	t.Helper()
	var p sim.BeliefReinforcedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestReconcileConfirm: the believed feature is present at the believed place
// ⇒ one confirmed movement carrying effective confidence + boost.
func TestReconcileConfirm(t *testing.T) {
	a := &sim.Agent{Beliefs: []sim.Belief{testBelief(3, "There is a fire at the woods (10,10).", 80, 1000)}}
	evs := reconcileObservation(a, 0, obsPayload(10, 10, "fire", "tree"), 1000, testConfirmBoost, testRetainPct)
	if len(evs) != 1 || evs[0].Type != "agent.belief_reinforced" {
		t.Fatalf("events = %v, want one agent.belief_reinforced", evs)
	}
	p := decodeReinforced(t, evs[0].Payload)
	if p.BeliefID != 3 || p.Kind != sim.BeliefConfirmed {
		t.Errorf("payload = %+v, want belief 3 confirmed", p)
	}
	// At the anchor tick effective == stored (80) ⇒ boosted to 90.
	if p.Confidence != 90 {
		t.Errorf("confidence = %d, want 90 (80 effective + %d boost)", p.Confidence, testConfirmBoost)
	}

	// Boost caps at 100 (both named features — fire and woods — present).
	a.Beliefs[0].Confidence = 95
	evs = reconcileObservation(a, 0, obsPayload(10, 10, "fire", "tree"), 1000, testConfirmBoost, testRetainPct)
	if p := decodeReinforced(t, evs[0].Payload); p.Confidence != 100 {
		t.Errorf("capped confidence = %d, want 100", p.Confidence)
	}
}

// TestReconcileDisconfirm: the place was observed and the feature is absent ⇒
// bounded decay (retain% of effective) — strictly faster than silence, which
// would have left the effective value untouched — plus exactly one surprise
// bump for the observation's companion memory.
func TestReconcileDisconfirm(t *testing.T) {
	const tick = int64(1000)
	a := &sim.Agent{
		Beliefs: []sim.Belief{testBelief(5, "The tendrils and stones of Thornspire stand at the forest's edge (20,20).", 80, tick)},
		Memories: []sim.Memory{{
			Tick: tick, Text: "Looked around: standing trees at the woods (20,20) — find Thornspire.",
			Salience: 2, Origin: sim.OriginObserved,
		}},
	}
	// Trees present (the "forest" token confirms), but no rock ⇒ "stones"
	// disconfirms: ANY absent expectation disconfirms an exhaustive scan.
	evs := reconcileObservation(a, 0, obsPayload(20, 20, "tree"), tick, testConfirmBoost, testRetainPct)
	if len(evs) != 2 {
		t.Fatalf("events = %d, want 2 (belief movement + surprise bump)", len(evs))
	}
	p := decodeReinforced(t, evs[0].Payload)
	if p.Kind != sim.BeliefDisconfirmed || p.BeliefID != 5 {
		t.Fatalf("first event = %+v, want belief 5 disconfirmed", p)
	}
	eff := sim.EffectiveConfidence(a.Beliefs[0], tick) // == 80 at the anchor
	want := int((int64(eff)*testRetainPct + 50) / 100) // 56
	if p.Confidence != want {
		t.Errorf("disconfirmed confidence = %d, want %d", p.Confidence, want)
	}
	if p.Confidence >= eff {
		t.Errorf("disconfirmation (%d) not faster than silence (%d)", p.Confidence, eff)
	}
	if evs[1].Type != "agent.memory_promoted" {
		t.Fatalf("second event = %s, want agent.memory_promoted", evs[1].Type)
	}
	var mp sim.MemoryPromotedPayload
	if err := json.Unmarshal(evs[1].Payload, &mp); err != nil {
		t.Fatal(err)
	}
	if mp.MemTick != tick || mp.TextHash != sim.MemoryHash(a.Memories[0].Text) || mp.Boost != disconfirmSalienceBoost {
		t.Errorf("surprise bump = %+v, want the observation memory at t%d boosted by %d", mp, tick, disconfirmSalienceBoost)
	}
}

// TestReconcileSilence: beliefs the observation cannot judge produce NOTHING —
// out-of-disc coordinates, no coordinates at all, and no recognized feature
// vocabulary all stay silent (their ambient decay is untouched by construction:
// no event, no clock movement).
func TestReconcileSilence(t *testing.T) {
	a := &sim.Agent{Beliefs: []sim.Belief{
		testBelief(1, "There is a fire at (40,40).", 80, 0),               // other place
		testBelief(2, "Rowan is a generous soul.", 60, 0),                 // no coords, no feature
		testBelief(3, "Something ancient sleeps beneath (10,10).", 50, 0), // in-disc, no recognized feature
	}}
	evs := reconcileObservation(a, 0, obsPayload(10, 10, "fire", "tree"), 1000, testConfirmBoost, testRetainPct)
	if len(evs) != 0 {
		t.Fatalf("unjudgeable beliefs produced %d events, want 0", len(evs))
	}
}

// TestReconcileOneBumpManyBeliefs: two disconfirmed beliefs about the same
// place produce two belief movements but only ONE surprise bump.
func TestReconcileOneBumpManyBeliefs(t *testing.T) {
	const tick = int64(500)
	a := &sim.Agent{
		Beliefs: []sim.Belief{
			testBelief(1, "A chest sits at (7,7).", 70, tick),
			testBelief(2, "An oven stands at (8,7).", 70, tick),
		},
		Memories: []sim.Memory{{Tick: tick, Text: "Looked around: nothing of note here at (7,7).", Salience: 2, Origin: sim.OriginObserved}},
	}
	evs := reconcileObservation(a, 0, obsPayload(7, 7), tick, testConfirmBoost, testRetainPct)
	var movements, bumps int
	for _, e := range evs {
		switch e.Type {
		case "agent.belief_reinforced":
			movements++
		case "agent.memory_promoted":
			bumps++
		}
	}
	if movements != 2 || bumps != 1 {
		t.Errorf("movements/bumps = %d/%d, want 2/1", movements, bumps)
	}
}

// TestReconcileMythDiesSlowly pins SC-001's shape end to end through the REAL
// reducer: an implanted false place-belief loses confidence on every recorded
// visit, survives several of them (myths die slowly), and trends below the
// confidence floor — through recorded events alone.
func TestReconcileMythDiesSlowly(t *testing.T) {
	const seed = 42
	s := sim.NewState(seed, worldmap.Generate(seed, 64, 64))
	a := &s.Agents[0]

	// Implant the myth: high-confidence, freshly anchored.
	revised, err := json.Marshal(sim.BeliefRevisedPayload{
		Agent: sim.Ref(0), BeliefID: 0, Statement: "The tendrils and stones of Thornspire stand at (30,30).",
		Confidence: 90, Provenance: sim.ProvenanceTold, Source: sim.Ref(1), Subject: sim.Ref(-1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(store.Event{Tick: 100, Type: "agent.belief_revised", Payload: revised}); err != nil {
		t.Fatal(err)
	}

	visits := 0
	tick := int64(200)
	for range [12]struct{}{} {
		b := a.Beliefs[0]
		if sim.EffectiveConfidence(b, tick) < sim.BeliefConfidenceFloor {
			break
		}
		evs := reconcileObservation(a, 0, obsPayload(30, 30, "tree"), tick, testConfirmBoost, testRetainPct)
		if len(evs) == 0 {
			t.Fatal("a visit to the believed place moved nothing")
		}
		for _, e := range evs {
			if e.Type == "agent.belief_reinforced" {
				if err := s.Apply(store.Event{Tick: tick, Type: e.Type, Payload: e.Payload}); err != nil {
					t.Fatal(err)
				}
			}
		}
		visits++
		tick += 7200 // the dedup window paces repeat visits
	}
	if visits < 2 {
		t.Errorf("myth died in %d visit(s) — want it to survive multiple (dials, not cliffs)", visits)
	}
	if got := sim.EffectiveConfidence(a.Beliefs[0], tick); got >= sim.BeliefConfidenceFloor {
		t.Errorf("after %d visits effective confidence = %d, want < floor (%d)", visits, got, sim.BeliefConfidenceFloor)
	}
}
