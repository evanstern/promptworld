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
		Payload: mustPayload(AgentMovedPayload{Agent: Ref(0), X: nx, Y: ny})}); err != nil {
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
		Payload: mustPayload(AgentMovedPayload{Agent: Ref(1), X: s.Agents[1].X, Y: s.Agents[1].Y})}); err != nil {
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
		if p.Agent.ID == 0 {
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
		if q.Agent.ID == 0 {
			t.Fatalf("settled map re-emitted agent.saw: %s", e.Payload)
		}
	}

	// A Detail change (refuel pushed FuelUntil) is a changed fact: re-emitted.
	s.Structures[len(s.Structures)-1].FuelUntil = 20000
	saw := false
	for _, e := range perceptionEvents(s, m, beat+2*moveEveryTicks) {
		var q SawPayload
		json.Unmarshal(e.Payload, &q)
		if q.Agent.ID == 0 {
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
		if q.Agent.ID == 0 {
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
			Payload: mustPayload(AgentMovedPayload{Agent: Ref(0), X: x, Y: y})}); err != nil {
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
			if p.Agent.ID == 0 {
				corrected = &p
			}
		case "agent.memory_added":
			var p MemoryAddedPayload
			mustUnmarshal(t, e.Payload, &p)
			if p.Agent.ID == 0 && p.Origin == OriginWitness {
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
					Payload: mustPayload(AgentMovedPayload{Agent: Ref(0), X: x, Y: y})}); err != nil {
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

// --- spec 041 US5: spatial knowledge spreads through talk (T031) -------------

// TestPlaceTellTransfer (SC-006): a founded talk passes at most placeTellCap
// facts per direction — told provenance, the TELLER's Seen tick, Source = the
// teller — the receiver's map gains them, companion memories ride both sides,
// and the receiver can act on the told fact at resolver level.
func TestPlaceTellTransfer(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a, b := &s.Agents[0], &s.Agents[1]
	b.X, b.Y = a.X, a.Y+1 // adjacent (a founded talk's shape)
	a.Map.Facts = nil
	b.Map.Facts = nil
	b.Inv.Wood = 2

	// A knows three fresh places; B knows nothing.
	a.Map.upsertFact(PlaceFact{Kind: "fire", X: 30, Y: 30, Seen: 900, Provenance: ProvenanceWitnessed, Detail: 999999})
	a.Map.upsertFact(PlaceFact{Kind: "tree", X: 12, Y: 12, Seen: 800, Provenance: ProvenanceWitnessed})
	a.Map.upsertFact(PlaceFact{Kind: "forage", X: 14, Y: 14, Seen: 700, Provenance: ProvenanceWitnessed})
	s.Structures = []Structure{{Kind: "fire", X: 30, Y: 30, FuelUntil: 999999}}

	const tick = int64(1000)
	evs := talkEvents(s, 0, 1, tick)
	var told *PlaceToldPayload
	tellerMem, listenerMem := false, false
	for _, e := range evs {
		switch e.Type {
		case "social.place_told":
			var p PlaceToldPayload
			mustUnmarshal(t, e.Payload, &p)
			if p.From.ID == 0 && p.To.ID == 1 {
				if told != nil {
					t.Fatal("more than one A→B place_told in a single talk")
				}
				q := p
				told = &q
			} else if p.From.ID == 1 {
				t.Fatalf("empty-map B told A something: %+v", p)
			}
		case "agent.memory_added":
			var p MemoryAddedPayload
			mustUnmarshal(t, e.Payload, &p)
			if p.Agent.ID == 0 && strings.HasPrefix(p.Text, "Told Birch about the") {
				tellerMem = true
			}
			if p.Agent.ID == 1 && strings.HasPrefix(p.Text, "Ash told you of a") {
				listenerMem = true
			}
		}
	}
	if told == nil {
		t.Fatal("no A→B social.place_told on a founded talk")
	}
	if len(told.Facts) != placeTellCap {
		t.Fatalf("transfer carried %d facts, want the cap %d", len(told.Facts), placeTellCap)
	}
	// Selection: freshest first → the fire (Seen 900) and the tree (Seen 800).
	kinds := map[string]int64{}
	for _, f := range told.Facts {
		kinds[f.Kind] = f.Seen
		if f.Provenance != ProvenanceTold || f.Source != 0 {
			t.Errorf("fact not baked told-by-teller: %+v", f)
		}
	}
	if kinds["fire"] != 900 || kinds["tree"] != 800 {
		t.Errorf("selection should carry the two freshest with the TELLER's Seen: %+v", told.Facts)
	}
	if !tellerMem || !listenerMem {
		t.Errorf("companion memories missing: teller=%v listener=%v", tellerMem, listenerMem)
	}

	// Apply the batch: B holds the told facts and acts on the fire.
	for _, e := range evs {
		if err := s.Apply(e); err != nil {
			t.Fatalf("apply %s: %v", e.Type, err)
		}
	}
	got, ok := b.Map.factAt("fire", 30, 30)
	if !ok || got.Provenance != ProvenanceTold || got.Seen != 900 || got.Source != 0 {
		t.Fatalf("receiver's told fact wrong: %+v", got)
	}
	in, _, err := resolveGoal(s, m, 1, "refuel_fire", -1, "", 0, tick+10)
	if err != nil || in.TargetX != 30 || in.TargetY != 30 {
		t.Fatalf("receiver should act on the told fire: %+v %v", in, err)
	}
}

// TestPlaceTellStalerNeverOverwrites (SC-006): a fact the listener holds at
// least as freshly is never offered, and a staler told fact applied through
// the reducer never overwrites fresher knowledge.
func TestPlaceTellStalerNeverOverwrites(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a, b := &s.Agents[0], &s.Agents[1]
	a.Map.Facts = nil
	b.Map.Facts = nil
	a.Map.upsertFact(PlaceFact{Kind: "fire", X: 30, Y: 30, Seen: 500, Provenance: ProvenanceWitnessed, Detail: 700})
	b.Map.upsertFact(PlaceFact{Kind: "fire", X: 30, Y: 30, Seen: 900, Provenance: ProvenanceWitnessed, Detail: 999})

	if facts := tellablePlaces(s, 0, 1, 1000); len(facts) != 0 {
		t.Fatalf("a staler fact was offered to a fresher holder: %+v", facts)
	}
	// Defensive reducer half: a staler payload still never overwrites.
	e := store.Event{Tick: 1000, Type: "social.place_told", Payload: mustPayload(PlaceToldPayload{
		From: Ref(0), To: Ref(1), Facts: []PlaceFact{{Kind: "fire", X: 30, Y: 30, Seen: 500, Provenance: ProvenanceTold, Source: 0, Detail: 700}},
	})}
	if err := s.Apply(e); err != nil {
		t.Fatal(err)
	}
	if got, _ := b.Map.factAt("fire", 30, 30); got.Seen != 900 || got.Provenance != ProvenanceWitnessed {
		t.Fatalf("fresher firsthand knowledge was overwritten by staler hearsay: %+v", got)
	}
}

// TestPlaceTellSecondhandKeepsOriginalSeen (SC-006): a relayed fact keeps the
// ORIGINAL observer's Seen (secondhand-of-secondhand is never fresher) while
// Source updates to the immediate teller.
func TestPlaceTellSecondhandKeepsOriginalSeen(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	b, c := &s.Agents[1], &s.Agents[2]
	b.Map.Facts = nil
	c.Map.Facts = nil
	// B holds a fact it was TOLD by A (agent 0), original Seen 400.
	b.Map.upsertFact(PlaceFact{Kind: "oven", X: 20, Y: 20, Seen: 400, Provenance: ProvenanceTold, Source: 0})

	facts := tellablePlaces(s, 1, 2, 1000)
	if len(facts) != 1 {
		t.Fatalf("relay offered %d facts, want 1", len(facts))
	}
	if facts[0].Seen != 400 {
		t.Errorf("relay freshened the fact: Seen %d, want the original 400", facts[0].Seen)
	}
	if facts[0].Source != 1 || facts[0].Provenance != ProvenanceTold {
		t.Errorf("relay must name the immediate teller: %+v", facts[0])
	}
}

// TestPlaceRevealedArm (spec 041 FR-014, T032): the metatron.place_revealed
// reducer arm validates rather than clamps (living target, real place) and
// stamps Seen/Provenance/Detail normatively — Seen = the landing tick,
// Provenance revealed, Detail = ground truth at landing (a fire's FuelUntil)
// — regardless of what the emitter baked beyond the place identity.
func TestPlaceRevealedArm(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	fx, fy := s.Agents[0].X, s.Agents[0].Y
	s.Structures = append(s.Structures, Structure{Kind: "fire", X: fx, Y: fy, FuelUntil: 9000})

	reveal := func(agent int, facts []PlaceFact, tick int64) error {
		return s.Apply(store.Event{Tick: tick, Type: "metatron.place_revealed",
			Payload: mustPayload(PlaceRevealedPayload{Agent: Ref(agent), Facts: facts})})
	}

	// A real place lands with the arm's normative stamps.
	if err := reveal(1, []PlaceFact{{Kind: "fire", X: fx, Y: fy, Provenance: ProvenanceRevealed}}, 500); err != nil {
		t.Fatalf("reveal: %v", err)
	}
	f, ok := s.Agents[1].Map.factAt("fire", fx, fy)
	if !ok {
		t.Fatal("revealed fact missing from the target's map")
	}
	if f.Seen != 500 || f.Provenance != ProvenanceRevealed || f.Detail != 9000 || f.Source != 0 {
		t.Fatalf("revealed fact not normatively stamped: %+v", f)
	}

	// The dry-run's rejections: unknown target, dead target, empty grant, and
	// a place that is not really there (the god reveals what is).
	if err := reveal(99, []PlaceFact{{Kind: "fire", X: fx, Y: fy}}, 600); err == nil {
		t.Error("out-of-range target accepted")
	}
	s.Agents[2].Dead = true
	if err := reveal(2, []PlaceFact{{Kind: "fire", X: fx, Y: fy}}, 600); err == nil {
		t.Error("dead target accepted")
	}
	if err := reveal(1, nil, 600); err == nil {
		t.Error("empty fact grant accepted")
	}
	if err := reveal(1, []PlaceFact{{Kind: "oven", X: fx, Y: fy}}, 600); err == nil {
		t.Error("a place absent from ground truth was revealed")
	}

	// A map-less living agent skips the upsert without error (reducer total).
	s.Agents[3].Map = nil
	if err := reveal(3, []PlaceFact{{Kind: "fire", X: fx, Y: fy}}, 700); err != nil {
		t.Fatalf("map-less reveal errored: %v", err)
	}
}

// TestPlaceRevealedThroughDoor (spec 041 FR-014, T032): the whole vision-grant
// path through the REAL InjectSocial door — the landVision batch shape (nudge,
// vision memory, place grant, companion omen memory) lands atomically, the
// fact re-stamped to the loop tick; a false place rejects the WHOLE batch,
// spending nothing.
func TestPlaceRevealedThroughDoor(t *testing.T) {
	h := newLadderHarness(t, func(s *State) {
		s.Structures = append(s.Structures, Structure{Kind: "fire", X: 7, Y: 7, FuelUntil: 99999})
	})

	visionBatch := func(kind string, x, y int) []store.Event {
		return []store.Event{
			{Type: "metatron.nudged", Payload: mustPayload(GuardianNudgedPayload{
				Form: "vision", Targets: []int{0}, Text: "Fire, beyond the ridge."})},
			{Type: "agent.memory_added", Payload: mustPayload(MemoryAddedPayload{
				Agent: Ref(0), Text: "You saw a vision: Fire, beyond the ridge.",
				Salience: SalDream, Subject: Ref(-1), Origin: OriginOmen})},
			{Type: "metatron.place_revealed", Payload: mustPayload(PlaceRevealedPayload{
				Agent: Ref(0), Facts: []PlaceFact{{Kind: kind, X: x, Y: y, Provenance: ProvenanceRevealed}}})},
			{Type: "agent.memory_added", Payload: mustPayload(MemoryAddedPayload{
				Agent: Ref(0), Text: "The vision showed you the fire at (7,7).",
				Salience: SalDream, Subject: Ref(-1), Origin: OriginOmen})},
		}
	}

	// A false place (no fire at (30,30)) rejects the whole batch at the door.
	if err := h.loop.InjectSocial(visionBatch("fire", 30, 30)); err == nil {
		t.Fatal("a false reveal passed the door")
	}
	stateJSON, _, err := h.loop.DoState()
	if err != nil {
		t.Fatal(err)
	}
	var before State
	if err := json.Unmarshal(stateJSON, &before); err != nil {
		t.Fatal(err)
	}
	if before.GuardianCharges != GuardianGenesisCharges {
		t.Fatalf("rejected batch spent a charge: %d", before.GuardianCharges)
	}
	if len(before.Agents[0].Memories) != 0 {
		t.Fatal("rejected batch landed a memory")
	}

	// The real place lands atomically, Seen re-stamped to the loop tick.
	if err := h.loop.InjectSocial(visionBatch("fire", 7, 7)); err != nil {
		t.Fatalf("vision grant rejected: %v", err)
	}
	stateJSON, _, err = h.loop.DoState()
	if err != nil {
		t.Fatal(err)
	}
	var after State
	if err := json.Unmarshal(stateJSON, &after); err != nil {
		t.Fatal(err)
	}
	f, ok := after.Agents[0].Map.factAt("fire", 7, 7)
	if !ok {
		t.Fatal("revealed fact missing after the door")
	}
	if f.Seen != after.Tick || f.Provenance != ProvenanceRevealed || f.Detail != 99999 {
		t.Fatalf("door-landed fact not stamped at the loop tick: %+v (tick %d)", f, after.Tick)
	}
	if after.GuardianCharges != GuardianGenesisCharges-1 {
		t.Errorf("charges = %d after the vision, want %d", after.GuardianCharges, GuardianGenesisCharges-1)
	}
	mems := after.Agents[0].Memories
	if len(mems) != 2 || mems[1].Text != "The vision showed you the fire at (7,7)." {
		t.Fatalf("companion omen memory missing: %+v", mems)
	}
	if mems[1].Origin != OriginOmen {
		t.Errorf("companion memory Origin = %q, want %q", mems[1].Origin, OriginOmen)
	}
}

// --- spec 041 polish: dead-agent hygiene (T033) -------------------------------

// TestDeadAgentKnowledgeHygiene (data-model invariant): death excludes an
// agent's map and peer sightings from every read path — the perception sweep,
// talk founding (and with it the place-knowledge transfer), and the seek
// resolver's live Dead gate — while the map DATA stays in state for
// historical fidelity (nothing un-knows the dead).
func TestDeadAgentKnowledgeHygiene(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a, b := &s.Agents[0], &s.Agents[1]
	b.X, b.Y = a.X, a.Y+1 // adjacent — a founded talk's shape, were both alive

	// The dead agent holds unique fresh knowledge the living one lacks, and
	// the living one holds a fresh sighting of the dead one.
	a.Map.Facts = nil
	b.Map.Facts = nil
	a.Map.upsertFact(PlaceFact{Kind: "fire", X: 30, Y: 30, Seen: 900, Provenance: ProvenanceWitnessed, Detail: 999999})
	b.Map.sightPeer(0, a.X, a.Y, 900)
	if err := s.Apply(store.Event{Tick: 950, Type: "agent.died",
		Payload: mustPayload(DiedPayload{Agent: Ref(0), Cause: "starvation"})}); err != nil {
		t.Fatal(err)
	}

	const tick = int64(1000)

	// The perception sweep never runs for the dead: no agent.saw, no
	// agent.map_corrected, on any beat.
	for probe := tick; probe < tick+moveEveryTicks; probe++ {
		for _, e := range perceptionEvents(s, m, probe) {
			var p struct {
				Agent AgentRef `json:"agent"` // dual-shape (spec 086)
			}
			mustUnmarshal(t, e.Payload, &p)
			if p.Agent.ID == 0 {
				t.Fatalf("the dead perceived: %s %s", e.Type, e.Payload)
			}
		}
	}

	// The social beat never founds a talk with the dead — so the dead agent's
	// map is neither told from (teller) nor told into (receiver).
	for _, e := range socialEvents(s, tick) {
		switch e.Type {
		case "agent.talked", "social.place_told", "social.rumor_told":
			t.Fatalf("the dead joined the social fabric: %s %s", e.Type, e.Payload)
		}
	}

	// A living agent's remembered sighting of the dead does not resolve: the
	// live Dead gate refuses at the door (death knowledge is out of the fact
	// model's scope — the accepted deviation).
	if _, _, err := resolveGoal(s, m, 1, "talk_to", 0, "", 0, tick); err == nil {
		t.Fatal("seek toward a dead agent resolved")
	} else if !strings.Contains(err.Error(), "dead") {
		t.Fatalf("seek refusal should name death, got: %v", err)
	}

	// Historical fidelity: the dead agent's map survives in state and across
	// a snapshot round trip — excluded from reads, never erased.
	var round State
	if err := json.Unmarshal(s.Marshal(), &round); err != nil {
		t.Fatal(err)
	}
	dead := round.Agents[0]
	if !dead.Dead || dead.Map == nil {
		t.Fatal("the dead agent's map was erased")
	}
	if f, ok := dead.Map.factAt("fire", 30, 30); !ok || f.Seen != 900 {
		t.Fatalf("the dead agent's knowledge was mutated: %+v", f)
	}

	// notePresence records nothing about (or for) the dead: a living walker
	// arriving beside the corpse neither refreshes its own sighting of the
	// dead nor stamps the dead agent's peer list.
	deadPeersBefore := len(round.Agents[0].Map.Peers)
	if err := s.Apply(store.Event{Tick: tick + 5, Type: "agent.moved",
		Payload: mustPayload(AgentMovedPayload{Agent: Ref(1), X: a.X, Y: a.Y})}); err != nil {
		t.Fatal(err)
	}
	if sight, ok := peerSightingOf(&s.Agents[1], 0); ok && sight.Seen > 900 {
		t.Fatalf("a living walker refreshed its sighting of the dead: %+v", sight)
	}
	if len(s.Agents[0].Map.Peers) != deadPeersBefore {
		t.Fatal("the dead agent's peer list grew after death")
	}
}
