package mind

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Spec 110 Phase 1 (T005/T006): the harvest ledger and its classifier. This
// phase is deliberately INERT — it records and classifies, and changes nothing
// about what chronicleNote emits; TestChronicleNoteCorrectionLineInertPhase1
// below is what pins that.

// TestHarvestLedgerPopulatedFromAbsorb (T003/T005, FR-001): the existing
// agent.chopped / agent.quarried absorb arm feeds the ledger, and the spec-081
// witness re-arm that shares the arm still fires.
func TestHarvestLedgerPopulatedFromAbsorb(t *testing.T) {
	state := sim.NewState(42, worldmap.Generate(42, 64, 64))
	state.Paused = false
	state.Tick = 5000
	state.Agents[0].X, state.Agents[0].Y = 20, 20
	state.Agents[1].X, state.Agents[1].Y = 21, 20
	state.Agents[1].Intent = &sim.Intent{Goal: "chop", TargetX: 20, TargetY: 20}
	md := &Mind{replica: state}

	md.absorb([]store.Event{
		{Tick: 5000, Seq: 7, Type: "agent.chopped",
			Payload: mustJSON(t, sim.HarvestPayload{Agent: sim.Ref(0), X: 20, Y: 20})},
		{Tick: 5100, Seq: 8, Type: "agent.quarried",
			Payload: mustJSON(t, sim.HarvestPayload{Agent: sim.Ref(3), X: 40, Y: 41})},
	})

	if who, ok := md.attributedHarvest(20, 20, 5000); !ok || who != 0 {
		t.Errorf("chop at (20,20): got (%d,%v), want (0,true)", who, ok)
	}
	if who, ok := md.attributedHarvest(40, 41, 5100); !ok || who != 3 {
		t.Errorf("quarry at (40,41): got (%d,%v), want (3,true)", who, ok)
	}
	// The arm's pre-existing behaviour is undisturbed.
	if !md.pending[0] {
		t.Error("harvester was not re-armed by its own chop")
	}
	if !md.pending[1] {
		t.Error("spec-081 witness re-arm no longer fires")
	}
}

// TestHarvestLedgerAgeEvictionAtWindowEdge (T002/T005, FR-001): the window is
// inclusive at its edge and evicts oldest-first past it — both on lookup (a
// stale entry never matches) and on the ledger's own eviction pass.
func TestHarvestLedgerAgeEvictionAtWindowEdge(t *testing.T) {
	var l harvestLedger
	l.record(3, 4, 2, 1000)

	if _, ok := l.lookup(3, 4, 1000+harvestLedgerWindow); !ok {
		t.Error("harvest exactly harvestLedgerWindow old must still explain a correction")
	}
	if _, ok := l.lookup(3, 4, 1000+harvestLedgerWindow+1); ok {
		t.Error("harvest one tick past the window must not explain a correction")
	}

	// A harvest at the window edge leaves the older entry in place...
	l.record(9, 9, 5, 1000+harvestLedgerWindow)
	if len(l.at) != 2 {
		t.Fatalf("entries = %d, want 2 (edge is inclusive)", len(l.at))
	}
	// ...and one tick past it evicts the older entry outright.
	l.record(9, 8, 5, 1000+harvestLedgerWindow+1)
	if _, present := l.at[harvestKey{3, 4}]; present {
		t.Error("entry older than the window was not evicted")
	}
	if _, present := l.at[harvestKey{9, 9}]; !present {
		t.Error("in-window entry was wrongly evicted")
	}
}

// TestHarvestLedgerCapEviction (T002/T005, FR-001): the hard entry cap bounds
// the ledger on a long run, dropping oldest-first.
func TestHarvestLedgerCapEviction(t *testing.T) {
	var l harvestLedger
	const extra = 5
	for i := 0; i < harvestLedgerCap+extra; i++ {
		// Distinct coordinates, strictly increasing ticks, all inside the
		// window so only the cap can evict.
		l.record(i%1000, i/1000, i%sim.AgentCount, int64(i))
	}
	if len(l.at) != harvestLedgerCap {
		t.Fatalf("entries = %d, want the cap %d", len(l.at), harvestLedgerCap)
	}
	// The `extra` oldest records are gone; the newest survives.
	for i := 0; i < extra; i++ {
		if _, present := l.at[harvestKey{i % 1000, i / 1000}]; present {
			t.Errorf("oldest entry %d survived cap eviction", i)
		}
	}
	last := harvestLedgerCap + extra - 1
	if _, present := l.at[harvestKey{last % 1000, last / 1000}]; !present {
		t.Error("newest entry was evicted by the cap")
	}
}

// TestAttributedHarvestExactCoordinate (T004/T005, FR-002): attribution is by
// exact tile. A neighbouring harvest explains nothing — the failure direction
// D4 chose deliberately, since a false attribution would silence a real
// mystery.
func TestAttributedHarvestExactCoordinate(t *testing.T) {
	md := &Mind{}
	md.harvests.record(12, 7, 4, 2000)

	if who, ok := md.attributedHarvest(12, 7, 2500); !ok || who != 4 {
		t.Errorf("exact hit: got (%d,%v), want (4,true)", who, ok)
	}
	for _, c := range [][2]int{{11, 7}, {13, 7}, {12, 6}, {12, 8}} {
		if _, ok := md.attributedHarvest(c[0], c[1], 2500); ok {
			t.Errorf("(%d,%d) must not be attributed to a harvest at (12,7)", c[0], c[1])
		}
	}
}

// TestAttributedHarvestMiss (T004/T005, FR-002 / SC-003): coordinates with no
// harvest in the ledger classify as unexplained — the genuine-anomaly path.
func TestAttributedHarvestMiss(t *testing.T) {
	md := &Mind{}
	if who, ok := md.attributedHarvest(1, 1, 100); ok || who != 0 {
		t.Errorf("empty ledger: got (%d,%v), want (0,false)", who, ok)
	}
	md.harvests.record(1, 1, 6, 100)
	if _, ok := md.attributedHarvest(1, 1, 100+harvestLedgerWindow+1); ok {
		t.Error("a harvest outside the window must classify the correction as unexplained")
	}
}

// TestChronicleNoteCorrectionLineInertPhase1 (T006): Phase 1 changes NOTHING
// about what chronicleNote emits. A correction whose coordinates the ledger
// explains still contributes its own per-event line, wording and position
// unchanged. Phase 2 is where the attributed case stops emitting a line and
// folds into the chapter tally instead — this test is the before-picture that
// proves the ledger arrived on its own.
func TestChronicleNoteCorrectionLineInertPhase1(t *testing.T) {
	md, _, _ := narrMind(t)
	// A harvest at the very coordinates the correction is about.
	md.harvests.record(12, 7, 0, 1000)
	if _, ok := md.attributedHarvest(12, 7, 2000); !ok {
		t.Fatal("test setup: the correction's coordinates should be attributed")
	}

	md.chronicleNote(mustEvent(t, 2000, "agent.map_corrected", sim.MapCorrectedPayload{
		Agent: sim.Ref(1),
		Gone:  []sim.PlaceFact{{Kind: "pine", X: 12, Y: 7}},
	}))
	if len(md.narrLines) != 1 {
		t.Fatalf("lines = %d, want 1 — Phase 1 must not change chronicleNote's output", len(md.narrLines))
	}
	const want = "Birch went looking for the pine at (12,7) and found it gone."
	if !strings.HasSuffix(md.narrLines[0], want) {
		t.Errorf("correction line = %q, want it to end with %q", md.narrLines[0], want)
	}

	// The unexplained case is identical today, and stays identical in Phase 2.
	md.chronicleNote(mustEvent(t, 2100, "agent.map_corrected", sim.MapCorrectedPayload{
		Agent: sim.Ref(1),
		Gone:  []sim.PlaceFact{{Kind: "pine", X: 60, Y: 60}},
	}))
	if len(md.narrLines) != 2 {
		t.Fatalf("lines = %d, want 2 (unexplained correction keeps its line)", len(md.narrLines))
	}
}
