package sim

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// --- test scaffolding (spec 041 US1) -----------------------------------------
//
// Pre-041 tests plant world state and resolve against it omnisciently; the
// knowledge gate now requires the ACTING agent to know the planted places.
// Tests whose subject is not knowledge use these arrangement helpers to grant
// exactly the pre-041 worldview.

// sightAll gives every living agent a current sighting of every other living
// agent (the pre-041 talk_to worldview).
func sightAll(s *State, tick int64) {
	for i := range s.Agents {
		a := &s.Agents[i]
		if a.Dead || a.Map == nil {
			continue
		}
		for j := range s.Agents {
			if j != i && !s.Agents[j].Dead {
				a.Map.sightPeer(j, s.Agents[j].X, s.Agents[j].Y, tick)
			}
		}
	}
}

// grantStructureFacts gives every living agent fresh witnessed facts for
// every current structure and pile (fires carrying FuelUntil as Detail) —
// the pre-041 structure worldview.
func grantStructureFacts(s *State, tick int64) {
	for i := range s.Agents {
		a := &s.Agents[i]
		if a.Dead || a.Map == nil {
			continue
		}
		for _, st := range s.Structures {
			a.Map.upsertFact(PlaceFact{Kind: st.Kind, X: st.X, Y: st.Y, Seen: tick,
				Provenance: ProvenanceWitnessed, Detail: st.FuelUntil})
		}
		for _, p := range s.Piles {
			a.Map.upsertFact(PlaceFact{Kind: "pile", X: p.X, Y: p.Y, Seen: tick,
				Provenance: ProvenanceWitnessed})
		}
	}
}

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

// TestPerceptionSweepWitnessesAndSettles (spec 041 T007): on an agent's
// perception beat the sweep emits agent.saw for ground truth its map lacks —
// resource tiles, structures, piles within the witness radius — with facts
// fully baked (witnessed, Seen = tick, sorted (Kind,X,Y), fires carrying
// FuelUntil as Detail). Once applied, an unchanged world emits nothing more
// for that agent (the diff settles); a Detail change (refuel) re-emits.
func TestPerceptionSweepWitnessesAndSettles(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a := &s.Agents[0]
	// Plant a lit fire and a pile beside agent 0 so structure/pile kinds are
	// exercised alongside the terrain kinds.
	s.Structures = append(s.Structures, Structure{Kind: "fire", X: a.X, Y: a.Y, FuelUntil: 9000})
	s.Piles = append(s.Piles, Pile{X: a.X, Y: a.Y, Wood: 2})

	beat := int64(moveEveryTicks * 8) // (tick + 0*3) % 5 == 0: agent 0's beat
	evs := perceptionEvents(s, m, beat)
	var mine []store.Event
	for _, e := range evs {
		var p SawPayload
		if e.Type != "agent.saw" {
			t.Fatalf("perception sweep emitted %s", e.Type)
		}
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			t.Fatal(err)
		}
		if p.Agent == 0 {
			mine = append(mine, e)
		}
	}
	if len(mine) != 1 {
		t.Fatalf("agent 0 emitted %d agent.saw events on its beat, want 1", len(mine))
	}
	var p SawPayload
	json.Unmarshal(mine[0].Payload, &p)
	var fire, pile bool
	for j, f := range p.Facts {
		if f.Seen != beat || f.Provenance != ProvenanceWitnessed {
			t.Fatalf("fact not baked at emission: %+v", f)
		}
		if abs(f.X-a.X)+abs(f.Y-a.Y) > witnessRadius {
			t.Fatalf("fact beyond the witness radius: %+v", f)
		}
		if j > 0 && factLess(f, p.Facts[j-1]) {
			t.Fatalf("payload facts out of canonical order at %d", j)
		}
		if f.Kind == "fire" && f.X == a.X && f.Y == a.Y {
			fire = true
			if f.Detail != 9000 {
				t.Errorf("fire fact Detail = %d, want the perceived FuelUntil 9000", f.Detail)
			}
		}
		if f.Kind == "pile" {
			pile = true
		}
	}
	if !fire || !pile {
		t.Errorf("sweep missed planted facts: fire=%v pile=%v (%+v)", fire, pile, p.Facts)
	}

	// Apply, then re-sweep the unchanged world on the next beat: settled.
	for _, e := range evs {
		if err := s.Apply(e); err != nil {
			t.Fatalf("apply: %v", err)
		}
	}
	for _, e := range perceptionEvents(s, m, beat+moveEveryTicks) {
		var q SawPayload
		json.Unmarshal(e.Payload, &q)
		if q.Agent == 0 {
			t.Fatalf("settled map re-emitted agent.saw: %s", e.Payload)
		}
	}

	// A Detail change (refuel pushed FuelUntil) is a changed fact: re-emitted.
	s.Structures[len(s.Structures)-1].FuelUntil = 20000
	saw := false
	for _, e := range perceptionEvents(s, m, beat+2*moveEveryTicks) {
		var q SawPayload
		json.Unmarshal(e.Payload, &q)
		if q.Agent == 0 {
			for _, f := range q.Facts {
				if f.Kind == "fire" && f.Detail == 20000 {
					saw = true
				}
			}
		}
	}
	if !saw {
		t.Error("a changed fire Detail was not re-witnessed")
	}

	// Asleep and dead agents never perceive.
	s.Agents[0].Asleep = true
	for _, e := range perceptionEvents(s, m, beat+3*moveEveryTicks) {
		var q SawPayload
		json.Unmarshal(e.Payload, &q)
		if q.Agent == 0 {
			t.Fatal("a sleeping agent perceived")
		}
	}
}

// TestMentalMapReplayByteIdentical (spec 041 T010, SC-003; the sim-level twin
// of internal/mind's TestJournalAndSituatedReplayByteIdentical): a live run
// whose moved/saw sequences fill the villagers' maps replays from genesis to
// BYTE-IDENTICAL per-agent mental maps — explored bitmap and fact list both —
// and an identical whole-state hash. Genesis seeding is reconstructed from
// seed (never replayed) and every fact rides a recorded event, so the log
// alone must reproduce every map.
func TestMentalMapReplayByteIdentical(t *testing.T) {
	const seed, ticks = 99, 6 * 3600 // a working morning: moves, saws, builds
	m := testMap(seed)
	live := NewState(seed, m)
	log := driveTicks(t, live, m, ticks, nil)

	// The run must actually exercise the write paths this test certifies.
	saws := 0
	for _, e := range log {
		if e.Type == "agent.saw" {
			saws++
		}
	}
	if saws == 0 {
		t.Fatal("no agent.saw events in a working morning — the sweep never fired")
	}
	grew := false
	for i := range live.Agents {
		if len(live.Agents[i].Map.Facts) > 0 {
			grew = true
		}
	}
	if !grew {
		t.Fatal("no villager holds any place-fact after a working morning")
	}

	replayed := NewState(seed, m)
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replayed.Tick = e.Tick
	}
	driveTicks(t, replayed, m, ticks, nil) // re-live the quiet tail, as recovery does

	for i := range live.Agents {
		lm, rm := mustPayload(live.Agents[i].Map), mustPayload(replayed.Agents[i].Map)
		if !bytes.Equal(lm, rm) {
			t.Errorf("agent %d mental map diverged on replay:\nlive:   %s\nreplay: %s", i, lm, rm)
		}
	}
	if live.Hash() != replayed.Hash() {
		t.Fatal("whole-state hash diverged on replay")
	}
}

// TestMapCorrectionOnArrival (spec 041 US3, SC-005): a villager learns a fire
// by witnessing it, walks away, the fire is removed while unobserved, and on
// walking back the perception beat emits agent.map_corrected carrying the
// fact AS REMEMBERED plus a companion situated memory — after application the
// fact is gone, the memory is stamped, and the next resolution uses the
// corrected knowledge (the honest "you know of no fires").
func TestMapCorrectionOnArrival(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a := &s.Agents[0]
	a.Inv.Wood = 2

	far, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool {
		return abs(x-a.X)+abs(y-a.Y) >= 25 && buildSite(m, s, x, y)
	})
	if !ok {
		t.Skip("seed lacks a far fire site")
	}
	stand, ok := nearest(m, s, far.X, far.Y, func(x, y int) bool {
		return passable(m, s, x, y) && abs(x-far.X)+abs(y-far.Y) <= 2
	})
	if !ok {
		t.Skip("no stand tile near the fire site")
	}
	move := func(tick int64, x, y int) {
		if err := s.Apply(store.Event{Tick: tick, Type: "agent.moved",
			Payload: mustPayload(AgentMovedPayload{Agent: 0, X: x, Y: y})}); err != nil {
			t.Fatal(err)
		}
	}
	sweep := func(tick int64) []store.Event { // agent 0's beat: tick%5 == 0
		evs := perceptionEvents(s, m, tick)
		for _, e := range evs {
			if err := s.Apply(e); err != nil {
				t.Fatalf("apply %s: %v", e.Type, err)
			}
		}
		return evs
	}

	// Learn the fire by witnessing it.
	s.Structures = []Structure{{Kind: "fire", X: far.X, Y: far.Y, FuelUntil: 900000}}
	move(100, stand.X, stand.Y)
	sweep(105)
	remembered, ok := a.Map.factAt("fire", far.X, far.Y)
	if !ok || remembered.Detail != 900000 {
		t.Fatalf("fire not witnessed: %+v", remembered)
	}

	// Walk far away; the fire is removed while unobserved — no correction
	// fires from out of range.
	awayX, awayY := s.Agents[0].X, s.Agents[0].Y
	if home, ok := nearest(m, s, far.X, far.Y, func(x, y int) bool {
		return passable(m, s, x, y) && abs(x-far.X)+abs(y-far.Y) > witnessRadius+2
	}); ok {
		awayX, awayY = home.X, home.Y
	}
	move(200, awayX, awayY)
	s.Structures = nil // demolished/removed while away
	for _, e := range sweep(205) {
		if e.Type == "agent.map_corrected" {
			t.Fatal("correction fired outside the witness radius")
		}
	}
	if _, ok := a.Map.factAt("fire", far.X, far.Y); !ok {
		t.Fatal("unwitnessed removal must not touch the map")
	}

	// Walk back: arrival's beat perceives the absence.
	move(300, stand.X, stand.Y)
	evs := sweep(305)
	var corrected *MapCorrectedPayload
	var memText string
	for _, e := range evs {
		switch e.Type {
		case "agent.map_corrected":
			var p MapCorrectedPayload
			mustUnmarshal(t, e.Payload, &p)
			if p.Agent == 0 {
				corrected = &p
			}
		case "agent.memory_added":
			var p MemoryAddedPayload
			mustUnmarshal(t, e.Payload, &p)
			if p.Agent == 0 && p.Origin == OriginWitness {
				memText = p.Text
			}
		}
	}
	if corrected == nil || len(corrected.Gone) != 1 {
		t.Fatalf("arrival did not emit the correction: %+v", corrected)
	}
	if corrected.Gone[0] != remembered {
		t.Errorf("Gone fact %+v != the fact as remembered %+v", corrected.Gone[0], remembered)
	}
	// situateText composes the place clause onto the base text (spec 019), so
	// assert the discovery stem, not the full composed line.
	stem := strings.TrimSuffix(mapCorrectedText(remembered), ".")
	if !strings.HasPrefix(memText, stem) {
		t.Errorf("companion memory %q does not carry the discovery %q", memText, stem)
	}
	if _, ok := a.Map.factAt("fire", far.X, far.Y); ok {
		t.Error("corrected fact still in the map")
	}
	found := false
	for _, mm := range a.Memories {
		if strings.HasPrefix(mm.Text, stem) && mm.Origin == OriginWitness {
			found = true
		}
	}
	if !found {
		t.Error("discovery memory not stamped on the soul")
	}
	// The next plan uses the corrected knowledge.
	if _, _, err := resolveGoal(s, m, 0, "refuel_fire", -1, "", 0, 310); err == nil ||
		err.Error() != "you know of no fires" {
		t.Errorf("post-correction resolution = %v, want the honest knowledge failure", err)
	}
}

// TestCorrectionSparesAvailabilityLapses (spec 041 US3): a harvested forage
// spot and a cooling den are places whose AVAILABILITY lapsed, not places
// gone — the correction never removes them (they regrow/re-arm; the
// resolvers' ground conditions cover the gap). A chopped tree is genuinely
// gone (the cleared overlay is permanent) and corrects.
func TestCorrectionSparesAvailabilityLapses(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a := &s.Agents[0]

	// Stand the agent somewhere with both a forage spot and a tree in radius
	// (the tiles themselves need not be passable — the sweep notes any tile
	// within the radius). Deterministic row-major scan.
	inRadius := func(cx, cy int, kind worldmap.TileKind) (Point, bool) {
		for y := cy - witnessRadius; y <= cy+witnessRadius; y++ {
			for x := cx - witnessRadius; x <= cx+witnessRadius; x++ {
				if m.InBounds(x, y) && abs(x-cx)+abs(y-cy) <= witnessRadius && m.At(x, y) == kind {
					return Point{X: x, Y: y}, true
				}
			}
		}
		return Point{}, false
	}
	var spot, treeTile Point
	placed := false
	for y := 0; y < m.H && !placed; y++ {
		for x := 0; x < m.W && !placed; x++ {
			if !passable(m, s, x, y) {
				continue
			}
			f, ok1 := inRadius(x, y, worldmap.Forage)
			tr, ok2 := inRadius(x, y, worldmap.Tree)
			if ok1 && ok2 {
				spot, treeTile = f, tr
				if err := s.Apply(store.Event{Tick: 1, Type: "agent.moved",
					Payload: mustPayload(AgentMovedPayload{Agent: 0, X: x, Y: y})}); err != nil {
					t.Fatal(err)
				}
				placed = true
			}
		}
	}
	if !placed {
		t.Skip("map has no vantage with both forage and tree in radius")
	}

	// Witness both, then harvest the spot and fell the tree.
	for _, e := range perceptionEvents(s, m, 5) {
		if err := s.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	if _, ok := a.Map.factAt("forage", spot.X, spot.Y); !ok {
		t.Fatal("forage spot not witnessed")
	}
	if _, ok := a.Map.factAt("tree", treeTile.X, treeTile.Y); !ok {
		t.Skip("tree tile not witnessed on this seed")
	}
	s.Harvested = append(s.Harvested, Harvest{X: spot.X, Y: spot.Y, Regrow: 999999})
	s.Cleared = append(s.Cleared, treeTile)

	for _, e := range perceptionEvents(s, m, 10) {
		if err := s.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	if _, ok := a.Map.factAt("forage", spot.X, spot.Y); !ok {
		t.Error("a harvested spot was corrected away — availability is not absence")
	}
	if _, ok := a.Map.factAt("tree", treeTile.X, treeTile.Y); ok {
		t.Error("a felled tree survived correction — the cleared overlay is permanent")
	}
}

// TestStalePlanStepExpiresWithoutOmniscience (spec 041 US3, T022): a plan
// step whose target kind was corrected away fails via agent.plan_expired
// carrying the knowledge reason — it never re-resolves omnisciently to a fire
// the agent has not learned.
func TestStalePlanStepExpiresWithoutOmniscience(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a := &s.Agents[0]
	a.Inv.Wood = 2
	// An unknown fire exists in the world; the agent's own fire knowledge was
	// just corrected away (map holds no fire facts).
	s.Structures = []Structure{{Kind: "fire", X: 40, Y: 40, FuelUntil: 900000}}
	a.Map.Facts = nil
	a.Plan = []PlanStep{{Job: "p", Goal: "refuel_fire", Target: -1, Until: 999999}}

	evs := planStepEvents(s, m, 0, 1000)
	var expired *PlanStepPayload
	for _, e := range evs {
		if e.Type == "agent.intent_set" {
			t.Fatalf("stale plan step re-resolved omnisciently: %s", e.Payload)
		}
		if e.Type == "agent.plan_expired" {
			var p PlanStepPayload
			mustUnmarshal(t, e.Payload, &p)
			expired = &p
		}
	}
	if expired == nil || expired.Reason != "you know of no fires" {
		t.Fatalf("plan step should expire with the knowledge reason, got %+v", expired)
	}
}
