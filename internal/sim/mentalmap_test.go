package sim

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// TestExploredCodecRoundTrip (spec 041 T003): marked bits survive the
// base64 encode/decode round trip exactly — no neighbors bleed, re-encoding
// an unchanged map is byte-identical, and marking is monotone.
func TestExploredCodecRoundTrip(t *testing.T) {
	const w, h = 64, 64
	mm := newMentalMap(w, h)

	// A fresh map knows nothing.
	for _, p := range []Point{{0, 0}, {63, 63}, {31, 40}} {
		if mm.ExploredAt(w, h, p.X, p.Y) {
			t.Errorf("fresh map claims (%d,%d) explored", p.X, p.Y)
		}
	}

	mm.MarkExplored(w, h, 10, 10, 2)
	before := mm.Explored
	// Everything within Manhattan distance 2 is set; the diamond's corners'
	// diagonal neighbors are not.
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			want := abs(dx)+abs(dy) <= 2
			if got := mm.ExploredAt(w, h, 10+dx, 10+dy); got != want {
				t.Errorf("ExploredAt(10%+d,10%+d) = %v, want %v", dx, dy, got, want)
			}
		}
	}
	// Round trip through JSON (the snapshot path) preserves the bits.
	b, err := json.Marshal(mm)
	if err != nil {
		t.Fatal(err)
	}
	var back MentalMap
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.Explored != before {
		t.Error("explored bitmap changed across a JSON round trip")
	}
	// Monotone: re-marking a subset changes nothing.
	mm.MarkExplored(w, h, 10, 10, 1)
	if mm.Explored != before {
		t.Error("re-marking already-explored tiles changed the bitmap (must be monotone)")
	}
}

// TestMarkExploredClipsBounds (spec 041 T003): marking near an edge clips to
// the map — no panic, no wraparound onto the far edge.
func TestMarkExploredClipsBounds(t *testing.T) {
	const w, h = 8, 8
	mm := newMentalMap(w, h)
	mm.MarkExplored(w, h, 0, 0, 3)
	if !mm.ExploredAt(w, h, 0, 0) || !mm.ExploredAt(w, h, 3, 0) || !mm.ExploredAt(w, h, 1, 2) {
		t.Error("in-bounds tiles within the radius should be explored")
	}
	// Row-major wraparound hazard: bit (y=0, x=-1) must not land on (y=... )
	// any other tile. The far edge of row 0 and the ends of rows 1..3 stay dark.
	for _, p := range []Point{{7, 0}, {4, 0}, {7, 1}, {0, 4}} {
		if mm.ExploredAt(w, h, p.X, p.Y) {
			t.Errorf("(%d,%d) explored — out-of-bounds marking leaked", p.X, p.Y)
		}
	}
	if mm.ExploredAt(w, h, -1, 0) || mm.ExploredAt(w, h, 0, 8) {
		t.Error("out-of-bounds queries must report unexplored")
	}
}

// TestFactUpsertOrderingDeterminism (spec 041 T003, research D1): upserting
// the same facts in two different orders yields byte-identical maps — the
// sorted (Kind, X, Y) invariant makes canonical bytes independent of
// discovery order.
func TestFactUpsertOrderingDeterminism(t *testing.T) {
	facts := []PlaceFact{
		{Kind: "tree", X: 5, Y: 9, Seen: 100, Provenance: ProvenanceWitnessed},
		{Kind: "fire", X: 12, Y: 34, Seen: 100, Provenance: ProvenanceWitnessed, Detail: 9000},
		{Kind: "fire", X: 3, Y: 40, Seen: 200, Provenance: ProvenanceWitnessed, Detail: 4000},
		{Kind: "tree", X: 5, Y: 2, Seen: 150, Provenance: ProvenanceWitnessed},
		{Kind: "pile", X: 1, Y: 1, Seen: 300, Provenance: ProvenanceWitnessed},
	}
	a, b := newMentalMap(8, 8), newMentalMap(8, 8)
	for _, f := range facts {
		a.upsertFact(f)
	}
	for i := len(facts) - 1; i >= 0; i-- { // reversed insertion order
		b.upsertFact(facts[i])
	}
	ab, bb := mustPayload(a), mustPayload(b)
	if !bytes.Equal(ab, bb) {
		t.Fatalf("insertion order leaked into canonical bytes:\n%s\n%s", ab, bb)
	}
	// Sorted invariant holds.
	for i := 1; i < len(a.Facts); i++ {
		if factLess(a.Facts[i], a.Facts[i-1]) {
			t.Fatalf("facts out of order at %d: %+v", i, a.Facts)
		}
	}
	// Upsert replaces in place: re-seeing the (12,34) fire refreshes it
	// without duplicating.
	a.upsertFact(PlaceFact{Kind: "fire", X: 12, Y: 34, Seen: 500, Provenance: ProvenanceWitnessed, Detail: 20000})
	if len(a.Facts) != len(facts) {
		t.Fatalf("upsert duplicated a (Kind,X,Y) slot: %d facts", len(a.Facts))
	}
	if f, ok := a.factAt("fire", 12, 34); !ok || f.Seen != 500 || f.Detail != 20000 {
		t.Errorf("upsert did not replace: %+v", f)
	}
	// removeFact deletes exactly its slot and preserves order.
	a.removeFact("tree", 5, 2)
	if _, ok := a.factAt("tree", 5, 2); ok {
		t.Error("removed fact still present")
	}
	if len(a.Facts) != len(facts)-1 {
		t.Errorf("remove dropped %d facts, want 1", len(facts)-len(a.Facts))
	}
}

// TestFactFreshnessHorizons (spec 041 T003, research D6): volatile kinds
// (fire, pile) stale within the short horizon, durable kinds hold to the long
// one; the boundary is exclusive (now − Seen == horizon is stale); KnownFresh
// filters by both kind and freshness.
func TestFactFreshnessHorizons(t *testing.T) {
	const seen = int64(1000)
	fire := PlaceFact{Kind: "fire", X: 1, Y: 1, Seen: seen, Provenance: ProvenanceWitnessed}
	shelter := PlaceFact{Kind: "shelter", X: 2, Y: 2, Seen: seen, Provenance: ProvenanceWitnessed}

	if !factFresh(fire, seen) || !factFresh(fire, seen+factHorizonVolatileTicks-1) {
		t.Error("a fire fact inside the volatile horizon must be fresh")
	}
	if factFresh(fire, seen+factHorizonVolatileTicks) {
		t.Error("a fire fact at the volatile horizon must be stale")
	}
	if !factFresh(shelter, seen+factHorizonVolatileTicks) {
		t.Error("a shelter fact is durable — the volatile horizon must not stale it")
	}
	if factFresh(shelter, seen+factHorizonDurableTicks) {
		t.Error("a shelter fact at the durable horizon must be stale")
	}

	mm := newMentalMap(8, 8)
	mm.upsertFact(fire)
	mm.upsertFact(shelter)
	mm.upsertFact(PlaceFact{Kind: "fire", X: 3, Y: 3, Seen: seen + 5000, Provenance: ProvenanceWitnessed})
	now := seen + factHorizonVolatileTicks // first fire stale, second fresh
	fresh := mm.KnownFresh("fire", now)
	if len(fresh) != 1 || fresh[0].X != 3 {
		t.Errorf("KnownFresh(fire) = %+v, want only the (3,3) fire", fresh)
	}
	if got := mm.KnownFresh("shelter", now); len(got) != 1 {
		t.Errorf("KnownFresh(shelter) = %+v, want the (2,2) shelter", got)
	}
	if got := mm.KnownFresh("oven", now); got != nil {
		t.Errorf("KnownFresh(oven) = %+v, want none", got)
	}
}

// TestGenesisSeedsSpawnExploration (spec 041 T006, research D7): NewState
// gives every villager a map with its spawn surroundings explored at the
// perception radius — and nothing else: no facts, no knowledge of far tiles.
// Pure function of (seed, map): two genesis runs seed identical maps.
func TestGenesisSeedsSpawnExploration(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	for i := range s.Agents {
		a := &s.Agents[i]
		if a.Map == nil {
			t.Fatalf("agent %d has no genesis map", i)
		}
		if !a.Map.ExploredAt(m.W, m.H, a.X, a.Y) {
			t.Errorf("agent %d does not know its own spawn tile", i)
		}
		// The diamond edge is in; just past it is out (clip-aware).
		if x := a.X + witnessRadius; x < m.W && !a.Map.ExploredAt(m.W, m.H, x, a.Y) {
			t.Errorf("agent %d spawn radius edge unexplored", i)
		}
		if x := a.X + witnessRadius + 1; x < m.W && a.Map.ExploredAt(m.W, m.H, x, a.Y) {
			t.Errorf("agent %d knows terrain beyond the perception radius", i)
		}
		if a.Map.Facts != nil {
			t.Errorf("agent %d granted genesis facts %+v — cold-start worlds grant none", i, a.Map.Facts)
		}
	}
	// Determinism: genesis maps are a pure function of (seed, map).
	s2 := NewState(seed, m)
	if !bytes.Equal(s.Marshal(), s2.Marshal()) {
		t.Error("two genesis runs of the same seed diverged")
	}
}

// TestMovedMarksExplored (spec 041 T005, research D2): the agent.moved
// reducer arm marks the mover's new surroundings explored — silent derived
// bookkeeping, monotone (the old area stays known). A map-less agent (dead at
// migration on a pre-041 world) is skipped without error.
func TestMovedMarksExplored(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a := &s.Agents[0]
	oldX, oldY := a.X, a.Y

	// Walk far enough that fresh terrain enters the radius.
	nx, ny := oldX, oldY
	for _, d := range neighborOrder {
		if m.Passable(oldX+d[0], oldY+d[1]) {
			nx, ny = oldX+d[0], oldY+d[1]
			break
		}
	}
	if nx == oldX && ny == oldY {
		t.Skip("spawn tile has no passable neighbor on this seed")
	}
	if err := s.Apply(store.Event{Tick: 10, Type: "agent.moved",
		Payload: mustPayload(AgentMovedPayload{Agent: 0, X: nx, Y: ny})}); err != nil {
		t.Fatalf("apply agent.moved: %v", err)
	}
	if !a.Map.ExploredAt(m.W, m.H, nx, ny) {
		t.Error("mover's new tile not explored")
	}
	if !a.Map.ExploredAt(m.W, m.H, oldX, oldY) {
		t.Error("exploration must be monotone — the old area was forgotten")
	}
	// The new radius edge in the step direction is now known.
	ex, ey := nx+(nx-oldX)*witnessRadius, ny+(ny-oldY)*witnessRadius
	if ex >= 0 && ey >= 0 && ex < m.W && ey < m.H && !a.Map.ExploredAt(m.W, m.H, ex, ey) {
		t.Errorf("radius edge (%d,%d) around the new position unexplored", ex, ey)
	}

	// Map-less agents skip the bookkeeping without error (nil-safe).
	s.Agents[1].Map = nil
	if err := s.Apply(store.Event{Tick: 11, Type: "agent.moved",
		Payload: mustPayload(AgentMovedPayload{Agent: 1, X: s.Agents[1].X, Y: s.Agents[1].Y})}); err != nil {
		t.Fatalf("apply agent.moved with nil map: %v", err)
	}
}
