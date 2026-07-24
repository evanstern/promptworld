package sim

// US1 tests (spec 041 T015): knowledge-gated target resolution with honest
// "you know of none" failures, known-far-beats-unknown-near, the
// witness-then-resolve flow, and full reflex parity (no omniscient fallback).

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// TestKnowledgeRejectionWording (US1 scenario 1, contracts §4): a gated verb
// with NO fresh matching fact fails with the knowledge phrasing — distinct
// from the existing reachability phrasing, which stays reserved for "knows
// some, reaches none".
func TestKnowledgeRejectionWording(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a := &s.Agents[0]
	a.Inv.Wood = 2
	a.Inv.FoodRaw = 1

	// World truth is irrelevant: plant a lit fire the agent has never seen.
	s.Structures = []Structure{{Kind: "fire", X: 30, Y: 30, FuelUntil: 100000}}

	for goal, want := range map[string]string{
		"forage":      "you know of no forage",
		"refuel_fire": "you know of no fires",
		"goto_warmth": "you know of no warm place",
		"cook":        "you know of no lit fire or oven",
	} {
		_, _, err := resolveGoal(s, m, 0, goal, -1, "", 0, 100)
		if err == nil || err.Error() != want {
			t.Errorf("%s with an empty map: err = %v, want %q", goal, err, want)
		}
	}

	// Knows some, reaches none: a forage fact whose tile is harvested keeps
	// the reachability phrasing — the two failure classes must stay distinct.
	fx, fy := 5, 5
	a.Map.upsertFact(PlaceFact{Kind: "forage", X: fx, Y: fy, Seen: 100, Provenance: ProvenanceWitnessed})
	s.Harvested = append(s.Harvested, Harvest{X: fx, Y: fy, Regrow: 200000})
	_, _, err := resolveGoal(s, m, 0, "forage", -1, "", 0, 100)
	if err == nil || err.Error() != "no forage reachable" {
		t.Errorf("known-but-unavailable forage: err = %v, want the reachability phrasing", err)
	}
	if strings.Contains("no forage reachable", "you know of no") {
		t.Fatal("phrasings must be distinguishable")
	}
}

// TestKnowledgeRejectionThroughDoor (US1 scenario 1): the knowledge rejection
// reaches the landing ladder verbatim as a recorded guard rejection — what
// the model sees on its next cycle.
func TestKnowledgeRejectionThroughDoor(t *testing.T) {
	l := landingLoop(func(s *State) {
		s.Agents[0].Inv.Wood = 2
		s.Agents[0].Map.Facts = nil // no place knowledge at all
	})
	emit, evs := captureEmit()
	args := meteredArgs(0, "refuel_fire")
	err := l.landIntent(&args, emit)
	if err == nil || err.Error() != OutcomeRejectedGuard+": you know of no fires" {
		t.Fatalf("door error = %v, want %q", err, OutcomeRejectedGuard+": you know of no fires")
	}
	found := false
	for _, e := range *evs {
		if e.typ == "agent.intent_rejected" {
			found = true
		}
	}
	if !found {
		t.Error("knowledge rejection was not recorded as agent.intent_rejected")
	}
}

// TestKnownFarBeatsUnknownNear (US1 scenario 3): two lit fires — one near
// but UNKNOWN, one far but KNOWN — a warmth verb resolves to the known far
// one, never the unknown-but-closer one.
func TestKnownFarBeatsUnknownNear(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a := &s.Agents[0]

	near, ok1 := nearest(m, s, a.X, a.Y, func(x, y int) bool {
		return abs(x-a.X)+abs(y-a.Y) >= 5 && buildSite(m, s, x, y)
	})
	far, ok2 := nearest(m, s, a.X, a.Y, func(x, y int) bool {
		return abs(x-a.X)+abs(y-a.Y) >= 25 && buildSite(m, s, x, y)
	})
	if !ok1 || !ok2 {
		t.Skip("seed lacks suitable fire sites")
	}
	s.Structures = []Structure{
		{Kind: "fire", X: near.X, Y: near.Y, FuelUntil: 100000},
		{Kind: "fire", X: far.X, Y: far.Y, FuelUntil: 100000},
	}
	// The agent knows ONLY the far fire.
	a.Map.upsertFact(PlaceFact{Kind: "fire", X: far.X, Y: far.Y, Seen: 100,
		Provenance: ProvenanceWitnessed, Detail: 100000})

	in, _, err := resolveGoal(s, m, 0, "goto_warmth", -1, "", 0, 100)
	if err != nil || in == nil {
		t.Fatalf("goto_warmth with a known fire failed: %v", err)
	}
	if d := abs(in.TargetX-far.X) + abs(in.TargetY-far.Y); d > fireWarmRadius {
		t.Errorf("target (%d,%d) is not at the known far fire %v", in.TargetX, in.TargetY, far)
	}
	if d := abs(in.TargetX-near.X) + abs(in.TargetY-near.Y); d <= fireWarmRadius {
		t.Errorf("target (%d,%d) warmed by the UNKNOWN near fire %v — omniscient resolution", in.TargetX, in.TargetY, near)
	}

	// refuel_fire picks the known far fire too, not the closer unknown one.
	a.Inv.Wood = 2
	in, _, err = resolveGoal(s, m, 0, "refuel_fire", -1, "", 0, 100)
	if err != nil || in.TargetX != far.X || in.TargetY != far.Y {
		t.Errorf("refuel target = %+v (err %v), want the known far fire %v", in, err, far)
	}
}

// TestWitnessThenResolve (US1 scenario 2 / independent test): a fire far
// outside the villager's known area rejects a warmth verb; after the villager
// travels within perception range and the sweep runs, the same verb resolves
// to that fire.
func TestWitnessThenResolve(t *testing.T) {
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
	s.Structures = []Structure{{Kind: "fire", X: far.X, Y: far.Y, FuelUntil: 500000}}

	if _, _, err := resolveGoal(s, m, 0, "refuel_fire", -1, "", 0, 100); err == nil ||
		err.Error() != "you know of no fires" {
		t.Fatalf("unwitnessed fire should reject with the knowledge phrasing, got %v", err)
	}

	// Walk the villager beside the fire (position via the reducer, as the
	// executor would) and run its perception beat.
	stand, ok := nearest(m, s, far.X, far.Y, func(x, y int) bool {
		return passable(m, s, x, y) && abs(x-far.X)+abs(y-far.Y) <= 2
	})
	if !ok {
		t.Skip("no stand tile near the fire")
	}
	if err := s.Apply(store.Event{Tick: 200, Type: "agent.moved",
		Payload: mustPayload(AgentMovedPayload{Agent: 0, X: stand.X, Y: stand.Y})}); err != nil {
		t.Fatal(err)
	}
	const beat = 205 // (205 + 0*3) % moveEveryTicks == 0: agent 0's beat
	for _, e := range perceptionEvents(s, m, beat) {
		if err := s.Apply(e); err != nil {
			t.Fatal(err)
		}
	}

	in, _, err := resolveGoal(s, m, 0, "refuel_fire", -1, "", 0, beat)
	if err != nil || in == nil || in.TargetX != far.X || in.TargetY != far.Y {
		t.Fatalf("witnessed fire should resolve: in=%+v err=%v want %v", in, err, far)
	}
}

// TestReflexParityNoOmniscientFallback (T014, clarify Q2): a hungry villager
// whose map holds no place-facts never resolves to an unseen food source —
// the get-food rung dead-ends honestly (until US4's search) and the ladder
// falls through to wander/idle.
func TestReflexParityNoOmniscientFallback(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a := &s.Agents[0]
	a.Needs.Food = hungryAt - 50 // hungry, nothing carried
	a.Map.Facts = nil            // knows no places at all

	d := decideIntent(s, m, 0, 1000)
	if d.directEvent != "" {
		t.Fatalf("nothing to eat, got direct event %q", d.directEvent)
	}
	if d.intent != nil {
		switch d.intent.Goal {
		case "forage", "hunt", "chop", "goto_warmth", "refuel_fire":
			t.Fatalf("empty-map reflex resolved %q to (%d,%d) — omniscient fallback",
				d.intent.Goal, d.intent.TargetX, d.intent.TargetY)
		}
	}

	// Same villager granted a forage fact: the SAME rung now resolves — the
	// gate is knowledge, not the rung.
	spot, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool {
		return effectiveKind(m, s, x, y) == worldmap.Forage
	})
	if !ok {
		t.Skip("no forage on this seed")
	}
	a.Map.upsertFact(PlaceFact{Kind: "forage", X: spot.X, Y: spot.Y, Seen: 1000, Provenance: ProvenanceWitnessed})
	d = decideIntent(s, m, 0, 1000)
	if d.intent == nil || d.intent.Goal != "forage" {
		t.Fatalf("known forage should resolve the get-food rung, got %+v", d)
	}
}

// TestTalkResolvesFromSighting (T013): talk_to/seek walk toward the acting
// agent's LAST SIGHTING — a stale sighting resolves to where the target was
// last seen (arrival guards cover the miss), and no sighting at all is the
// honest knowledge failure.
func TestTalkResolvesFromSighting(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a := &s.Agents[0]
	a.Map.Peers = nil // never seen anyone

	if _, _, err := resolveGoal(s, m, 0, "talk_to", 1, "", 0, 100); err == nil ||
		err.Error() != "you do not know where Birch is" {
		t.Fatalf("sightless talk_to: err = %v, want the knowledge phrasing", err)
	}

	// Sight Birch at an old position, then move Birch elsewhere: seek targets
	// the SIGHTING, not the live coordinates.
	oldX, oldY := s.Agents[1].X, s.Agents[1].Y
	a.Map.sightPeer(1, oldX, oldY, 100)
	s.Agents[1].X, s.Agents[1].Y = (oldX+20)%m.W, (oldY+20)%m.H
	in, _, err := resolveGoal(s, m, 0, "talk_to", 1, "", 0, 200)
	if err != nil || in == nil || in.Goal != "seek" {
		t.Fatalf("sighted talk_to failed: %+v %v", in, err)
	}
	if in.TargetX != oldX || in.TargetY != oldY {
		t.Errorf("seek target (%d,%d), want the last sighting (%d,%d) — live coordinates leaked",
			in.TargetX, in.TargetY, oldX, oldY)
	}
}
