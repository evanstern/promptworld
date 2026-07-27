package sim

import (
	"fmt"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Spec 081 (first-person harvest memory): a completed chop/quarry removes the
// place-fact from the actor's and every awake in-radius witness's map at the
// act event, mints a first-person act memory for the actor, and never later
// "corrects" a harvest its owner or an on-scene witness already knows about.
// Genuine return-discovery (absent/asleep at act time) still fires.

// chopEvent / quarryEvent build a bare harvest event for direct reducer tests.
func chopEvent(agent, x, y int, tick int64) store.Event {
	return store.Event{Tick: tick, Type: "agent.chopped", Payload: mustPayload(HarvestPayload{Agent: Ref(agent), X: x, Y: y})}
}

func quarryEvent(agent, x, y int, tick int64) store.Event {
	return store.Event{Tick: tick, Type: "agent.quarried", Payload: mustPayload(HarvestPayload{Agent: Ref(agent), X: x, Y: y})}
}

// TestHarvestReducerRemovesActorFact (T006, US1): the chop/quarry reducer arms
// remove the actor's matching fact at the act event; a chop for a tile the
// actor holds no fact for is a silent no-op removal (the perception-cadence gap
// edge case — the first-person memory, minted by the executor, is unaffected).
func TestHarvestReducerRemovesActorFact(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	isolateAgents(s)
	a := &s.Agents[0]
	a.Dead = false

	a.Map.upsertFact(PlaceFact{Kind: "tree", X: 5, Y: 6, Seen: 1, Provenance: ProvenanceWitnessed})
	if err := s.Apply(chopEvent(0, 5, 6, 10)); err != nil {
		t.Fatal(err)
	}
	if _, ok := a.Map.factAt("tree", 5, 6); ok {
		t.Error("chop left the actor's felled tree in the map")
	}

	// Never-knew: no fact at the tile — the removal is a no-op and must not panic.
	if err := s.Apply(chopEvent(0, 7, 8, 11)); err != nil {
		t.Fatal(err)
	}
	if _, ok := a.Map.factAt("tree", 7, 8); ok {
		t.Error("a never-known tile gained a fact from the reducer")
	}

	// Quarry parity with kind rock.
	a.Map.upsertFact(PlaceFact{Kind: "rock", X: 9, Y: 10, Seen: 1, Provenance: ProvenanceWitnessed})
	if err := s.Apply(quarryEvent(0, 9, 10, 12)); err != nil {
		t.Fatal(err)
	}
	if _, ok := a.Map.factAt("rock", 9, 10); ok {
		t.Error("quarry left the actor's quarried outcrop in the map")
	}
}

// TestHarvestReducerWitnessSet (T011, US2): a completed chop removes the fact
// from every awake living in-radius villager's map — provenance-blind (a told
// fact goes too) — while an asleep, dead, or out-of-radius villager keeps it.
func TestHarvestReducerWitnessSet(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	// Configure five agents around the felled tile at (30,30). Positions are
	// read straight from state, so no real terrain is needed for this arm.
	const tx, ty = 30, 30
	seedFact := func(i, x, y int, prov string) {
		s.Agents[i].Dead = false
		s.Agents[i].Asleep = false
		s.Agents[i].X, s.Agents[i].Y = x, y
		s.Agents[i].Map.upsertFact(PlaceFact{Kind: "tree", X: tx, Y: ty, Seen: 1, Provenance: prov})
	}
	seedFact(0, tx, ty, ProvenanceWitnessed)                 // actor
	seedFact(1, tx+3, ty, ProvenanceTold)                    // awake, in radius, hearsay
	seedFact(2, tx+2, ty, ProvenanceWitnessed)               // asleep, in radius
	seedFact(3, tx+1, ty, ProvenanceWitnessed)               // dead, in radius
	seedFact(4, tx, ty+witnessRadius+4, ProvenanceWitnessed) // awake, out of radius
	s.Agents[2].Asleep = true
	s.Agents[3].Dead = true

	if err := s.Apply(chopEvent(0, tx, ty, 20)); err != nil {
		t.Fatal(err)
	}

	if _, ok := s.Agents[0].Map.factAt("tree", tx, ty); ok {
		t.Error("actor kept the felled tree")
	}
	if _, ok := s.Agents[1].Map.factAt("tree", tx, ty); ok {
		t.Error("awake in-radius witness kept the felled tree (told provenance must be removed too)")
	}
	if _, ok := s.Agents[2].Map.factAt("tree", tx, ty); !ok {
		t.Error("asleep witness lost the fact — they did not see it fall")
	}
	if _, ok := s.Agents[3].Map.factAt("tree", tx, ty); !ok {
		t.Error("dead villager's map was touched")
	}
	if _, ok := s.Agents[4].Map.factAt("tree", tx, ty); !ok {
		t.Error("out-of-radius villager lost the fact — they must discover it on return")
	}
}

// TestHarvestMintsFirstPersonMemory (T007, US1): driving a chop then a quarry
// to completion accretes exactly one first-person act memory per act via
// agent.memory_added, in the salChop/salQuarry (salHunt) band, carrying the
// intent reason — the TestMemoriesAccrete posture (never appended to state).
func TestHarvestMintsFirstPersonMemory(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	isolateAgents(s)
	a := &s.Agents[0]
	a.Dead = false
	const reason = "the hearth needs wood"

	setIntent := func(goal string, tx, ty, rx, ry int) {
		if err := s.Apply(store.Event{Tick: s.Tick, Type: "agent.intent_set",
			Payload: mustPayload(IntentSetPayload{Agent: Ref(0), Goal: goal, TargetX: tx, TargetY: ty,
				ResX: rx, ResY: ry, Source: "planner", Reason: reason})}); err != nil {
			t.Fatal(err)
		}
	}

	stand, res := treeStand(t, m, s, a.X, a.Y)
	setIntent("chop", stand.X, stand.Y, res.X, res.Y)
	log := driveTicks(t, s, m, s.Tick+8000, nil)

	qstand, qres := rockStand(t, m, s, a.X, a.Y)
	setIntent("quarry", qstand.X, qstand.Y, qres.X, qres.Y)
	log = append(log, driveTicks(t, s, m, s.Tick+8000, nil)...)

	assertOneActMemory(t, log, salChop, reason,
		fmt.Sprintf("Felled the tree at (%d,%d)", res.X, res.Y))
	assertOneActMemory(t, log, salQuarry, reason,
		fmt.Sprintf("Quarried the outcrop at (%d,%d)", qres.X, qres.Y))
}

// assertOneActMemory checks exactly one agent.memory_added carries the given
// stem (base text minus the situated place/why clauses), at the given salience,
// origin action, with the driving reason as Why.
func assertOneActMemory(t *testing.T, log []store.Event, salience int, reason, stem string) {
	t.Helper()
	count := 0
	for _, e := range log {
		if e.Type != "agent.memory_added" {
			continue
		}
		var p MemoryAddedPayload
		mustUnmarshal(t, e.Payload, &p)
		if !strings.HasPrefix(p.Text, stem) {
			continue
		}
		count++
		if p.Agent.ID != 0 {
			t.Errorf("act memory %q went to agent %d, not the actor", p.Text, p.Agent.ID)
		}
		if p.Salience != salience {
			t.Errorf("act memory %q salience = %d, want %d", p.Text, p.Salience, salience)
		}
		if p.Origin != OriginAction {
			t.Errorf("act memory %q origin = %q, want %q", p.Text, p.Origin, OriginAction)
		}
		if p.Why != reason {
			t.Errorf("act memory %q Why = %q, want the driving reason %q", p.Text, p.Why, reason)
		}
	}
	if count != 1 {
		t.Errorf("first-person act memory %q… minted %d times, want exactly 1", stem, count)
	}
}

// TestNoSelfCorrectionAfterChop (T008, SC-001): once the actor's own fact is
// removed at the act event, no later perception beat corrects the actor's map
// for the tile it just chopped.
func TestNoSelfCorrectionAfterChop(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	isolateAgents(s)
	a := &s.Agents[0]
	a.Dead = false

	stand, res := treeStand(t, m, s, a.X, a.Y)
	applyEvents(t, s, store.Event{Tick: 1, Type: "agent.moved",
		Payload: mustPayload(AgentMovedPayload{Agent: Ref(0), X: stand.X, Y: stand.Y})})
	applyEvents(t, s, perceptionEvents(s, m, 5)...) // agent 0 beat: tick%5 == 0
	if _, ok := a.Map.factAt("tree", res.X, res.Y); !ok {
		t.Skip("tree not witnessed on this seed/stand")
	}

	applyEvents(t, s, chopEvent(0, res.X, res.Y, 6))
	if _, ok := a.Map.factAt("tree", res.X, res.Y); ok {
		t.Fatal("actor's felled tree survived the chop reducer")
	}

	for _, tick := range []int64{10, 15, 20, 25} {
		for _, e := range perceptionEvents(s, m, tick) {
			assertNoCorrection(t, e, 0, res.X, res.Y)
			applyEvents(t, s, e)
		}
	}
}

// TestSameTickSweepNoCorrection (T012, FR-006, research D5): a perception beat
// landing on the act tick emits no correction for the acted tile — the sweep
// reads pre-batch state where the tree still stands — and the next beat finds
// the fact already removed, so it too stays silent.
func TestSameTickSweepNoCorrection(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	isolateAgents(s)
	a := &s.Agents[0]
	a.Dead = false

	stand, res := treeStand(t, m, s, a.X, a.Y)
	applyEvents(t, s, store.Event{Tick: 1, Type: "agent.moved",
		Payload: mustPayload(AgentMovedPayload{Agent: Ref(0), X: stand.X, Y: stand.Y})})
	applyEvents(t, s, perceptionEvents(s, m, 5)...)
	if _, ok := a.Map.factAt("tree", res.X, res.Y); !ok {
		t.Skip("tree not witnessed on this seed/stand")
	}

	// Act tick 10 (agent 0 beat): the sweep runs BEFORE the chop lands (the
	// stepEvents ordering), tree still in ground truth → no correction.
	for _, e := range perceptionEvents(s, m, 10) {
		assertNoCorrection(t, e, 0, res.X, res.Y)
		applyEvents(t, s, e)
	}
	applyEvents(t, s, chopEvent(0, res.X, res.Y, 10))
	if _, ok := a.Map.factAt("tree", res.X, res.Y); ok {
		t.Fatal("fact survived the same-tick chop")
	}
	// Next beat: nothing left to correct.
	for _, e := range perceptionEvents(s, m, 15) {
		assertNoCorrection(t, e, 0, res.X, res.Y)
		applyEvents(t, s, e)
	}
}

// TestReturnDiscoveryStillCorrects (T014, US3 / US2-AS3): an agent outside
// witness radius at the act tick, and one asleep at the act tick, both keep the
// fact and later correct exactly once — the genuine return-discovery narrative.
func TestReturnDiscoveryStillCorrects(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	isolateAgents(s)
	for _, i := range []int{0, 1, 2} {
		s.Agents[i].Dead = false
	}
	actor := &s.Agents[0]
	stand, res := treeStand(t, m, s, actor.X, actor.Y)

	move := func(i, x, y int, tick int64) {
		applyEvents(t, s, store.Event{Tick: tick, Type: "agent.moved",
			Payload: mustPayload(AgentMovedPayload{Agent: Ref(i), X: x, Y: y})})
	}
	// All three co-located at the stand, then a full sweep window so each of
	// agents 0,1,2 hits its beat once and witnesses the tree.
	move(0, stand.X, stand.Y, 1)
	move(1, stand.X, stand.Y, 1)
	move(2, stand.X, stand.Y, 1)
	sweepWindow(t, s, m, 5)
	for _, i := range []int{0, 1, 2} {
		if _, ok := s.Agents[i].Map.factAt("tree", res.X, res.Y); !ok {
			t.Skipf("agent %d did not witness the tree on this seed", i)
		}
	}

	// Agent 1 walks out of radius; agent 2 falls asleep in place. Agent 0 chops.
	away, ok := nearest(m, s, res.X, res.Y, func(x, y int) bool {
		return passable(m, s, x, y) && abs(x-res.X)+abs(y-res.Y) > witnessRadius+2
	})
	if !ok {
		t.Skip("no out-of-radius stand near the tree")
	}
	move(1, away.X, away.Y, 20)
	s.Agents[2].Asleep = true
	applyEvents(t, s, chopEvent(0, res.X, res.Y, 21))

	if _, ok := s.Agents[1].Map.factAt("tree", res.X, res.Y); !ok {
		t.Fatal("out-of-radius agent lost the fact at act time")
	}
	if _, ok := s.Agents[2].Map.factAt("tree", res.X, res.Y); !ok {
		t.Fatal("asleep agent lost the fact at act time")
	}

	// Agent 1 returns; agent 2 wakes. Both correct exactly once on their beat.
	move(1, stand.X, stand.Y, 40)
	s.Agents[2].Asleep = false
	corr := map[int]int{}
	for tick := int64(45); tick < 60; tick++ {
		for _, e := range perceptionEvents(s, m, tick) {
			if e.Type == "agent.map_corrected" {
				var p MapCorrectedPayload
				mustUnmarshal(t, e.Payload, &p)
				for _, f := range p.Gone {
					if f.Kind == "tree" && f.X == res.X && f.Y == res.Y {
						corr[p.Agent]++
					}
				}
			}
			applyEvents(t, s, e)
		}
	}
	if corr[1] != 1 {
		t.Errorf("returning out-of-radius agent corrected %d times, want exactly 1", corr[1])
	}
	if corr[2] != 1 {
		t.Errorf("woken in-radius agent corrected %d times, want exactly 1", corr[2])
	}
	if corr[0] != 0 {
		t.Errorf("actor corrected %d times, want 0 (their own act was first-person)", corr[0])
	}
}

// TestOnlyAbsentAgentCorrects (T015, contract §4 invariant + SC-004): with an
// actor, an on-scene awake witness held in radius, and an absent agent, the
// only map correction for the tile names the absent agent, and the on-scene
// witness accrues no memory at all from the act.
func TestOnlyAbsentAgentCorrects(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	isolateAgents(s)
	for _, i := range []int{0, 1, 2} {
		s.Agents[i].Dead = false
	}
	actor := &s.Agents[0]
	stand, res := treeStand(t, m, s, actor.X, actor.Y)

	move := func(i, x, y int, tick int64) {
		applyEvents(t, s, store.Event{Tick: tick, Type: "agent.moved",
			Payload: mustPayload(AgentMovedPayload{Agent: Ref(i), X: x, Y: y})})
	}
	move(0, stand.X, stand.Y, 1)
	move(1, stand.X, stand.Y, 1) // on-scene witness, stays put
	move(2, stand.X, stand.Y, 1) // absent-to-be
	sweepWindow(t, s, m, 5)
	for _, i := range []int{0, 1, 2} {
		if _, ok := s.Agents[i].Map.factAt("tree", res.X, res.Y); !ok {
			t.Skipf("agent %d did not witness the tree on this seed", i)
		}
	}

	// Snapshot the on-scene witness's memory count; the act must add nothing.
	memsBefore := len(s.Agents[1].Memories)

	away, ok := nearest(m, s, res.X, res.Y, func(x, y int) bool {
		return passable(m, s, x, y) && abs(x-res.X)+abs(y-res.Y) > witnessRadius+2
	})
	if !ok {
		t.Skip("no out-of-radius stand near the tree")
	}
	move(2, away.X, away.Y, 20)
	applyEvents(t, s, chopEvent(0, res.X, res.Y, 21))
	move(2, stand.X, stand.Y, 40)

	corrections := 0
	for tick := int64(45); tick < 60; tick++ {
		for _, e := range perceptionEvents(s, m, tick) {
			if e.Type == "agent.map_corrected" {
				var p MapCorrectedPayload
				mustUnmarshal(t, e.Payload, &p)
				for _, f := range p.Gone {
					if f.Kind == "tree" && f.X == res.X && f.Y == res.Y {
						corrections++
						if p.Agent != 2 {
							t.Errorf("correction for the felled tile named agent %d, want only the absent agent 2", p.Agent)
						}
					}
				}
			}
			applyEvents(t, s, e)
		}
	}
	if corrections != 1 {
		t.Errorf("the felled tile drew %d corrections, want exactly 1 (the absent agent)", corrections)
	}
	if got := len(s.Agents[1].Memories); got != memsBefore {
		t.Errorf("on-scene witness accrued %d memories from the act, want 0", got-memsBefore)
	}
}

// TestChopWitnessReplayByteIdentical (T016, SC-005): folding a scripted log
// with a chop and an in-radius witness into two fresh genesis states yields
// byte-identical canonical state — mental maps included — and the maps actually
// changed (both actor and witness lost the fact), so the check is meaningful.
func TestChopWitnessReplayByteIdentical(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	scratch := NewState(seed, m)
	stand, res := treeStand(t, m, scratch, scratch.Agents[0].X, scratch.Agents[0].Y)

	treeFact := PlaceFact{Kind: "tree", X: res.X, Y: res.Y, Seen: 1, Provenance: ProvenanceWitnessed}
	log := []store.Event{
		{Tick: 1, Type: "agent.moved", Payload: mustPayload(AgentMovedPayload{Agent: Ref(0), X: stand.X, Y: stand.Y})},
		{Tick: 1, Type: "agent.moved", Payload: mustPayload(AgentMovedPayload{Agent: Ref(1), X: stand.X, Y: stand.Y})},
		{Tick: 2, Type: "agent.saw", Payload: mustPayload(SawPayload{Agent: 0, Facts: []PlaceFact{treeFact}})},
		{Tick: 2, Type: "agent.saw", Payload: mustPayload(SawPayload{Agent: 1, Facts: []PlaceFact{treeFact}})},
		chopEvent(0, res.X, res.Y, 3),
	}

	fold := func() *State {
		s := NewState(seed, m)
		for _, e := range log {
			if err := s.Apply(e); err != nil {
				t.Fatalf("fold %s: %v", e.Type, err)
			}
			s.Tick = e.Tick
		}
		return s
	}
	a, b := fold(), fold()

	if _, ok := a.Agents[0].Map.factAt("tree", res.X, res.Y); ok {
		t.Fatal("fold did not remove the actor's fact — the test would be vacuous")
	}
	if _, ok := a.Agents[1].Map.factAt("tree", res.X, res.Y); ok {
		t.Fatal("fold did not remove the in-radius witness's fact")
	}
	if a.Hash() != b.Hash() {
		t.Fatalf("chop+witness fold diverged:\na: %s\nb: %s", string(a.Marshal()), string(b.Marshal()))
	}
}

// --- local helpers ----------------------------------------------------------

// applyEvents applies each event to s, failing the test on any reducer error.
func applyEvents(t *testing.T, s *State, evs ...store.Event) {
	t.Helper()
	for _, e := range evs {
		if err := s.Apply(e); err != nil {
			t.Fatalf("apply %s: %v", e.Type, err)
		}
	}
}

// sweepWindow runs the perception sweep across five consecutive ticks so every
// awake agent's staggered beat fires exactly once, applying what it emits.
func sweepWindow(t *testing.T, s *State, m *worldmap.Map, start int64) {
	t.Helper()
	for tick := start; tick < start+5; tick++ {
		applyEvents(t, s, perceptionEvents(s, m, tick)...)
	}
}

// assertNoCorrection fails if e is a map correction naming (agent, tree, x, y).
func assertNoCorrection(t *testing.T, e store.Event, agent, x, y int) {
	t.Helper()
	if e.Type != "agent.map_corrected" {
		return
	}
	var p MapCorrectedPayload
	mustUnmarshal(t, e.Payload, &p)
	if p.Agent != agent {
		return
	}
	for _, f := range p.Gone {
		if f.Kind == "tree" && f.X == x && f.Y == y {
			t.Errorf("agent %d was corrected for the tree at (%d,%d) it was on-scene for", agent, x, y)
		}
	}
}
