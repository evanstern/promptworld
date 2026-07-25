package sim

// Spec 062 US4 (T010, SC-001): the world-01 forage<->goto_warmth thrash
// regression. The layer fight was: the planner sends a cold, fed villager to
// warmth; on arrival-idle the old reflex counter-schedules a larder forage,
// dragging it back off the fire; the planner re-issues warmth; repeat every
// ~200-320 ticks (Sage: 436 flips, 334 within <=200 ticks). This one test
// encodes BOTH proofs:
//
//   - the OLD flip: the prep the new gate holds back IS a larder forage away
//     from the fire — absent the gate's inputs (window unarmed, warmth healthy —
//     the pre-062 world had neither the yield window nor a day warmth danger
//     band) the reflex forages, the counter-schedule;
//   - the NEW hold + recovery: with the planner intent completing inside the
//     yield window, ZERO prep intents fire and the warmth trajectory recovers.
//
// prepYieldTicks and the danger bands are deliberately immutable consts (R4,
// FR-006), so the "doctrine zeroed" inverse is expressed by nullifying the
// gate's INPUTS (unarmed window + healthy warmth), not by mutating the
// constants — the faithful, const-respecting encoding of the pre-062 world.

import (
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

func isPrepGoal(goal string) bool {
	// The PREP rungs' goals (prepDecision): first-fire build, refuel top-up, and
	// the larder forage/hunt. build_fire/refuel_fire also appear as SURVIVAL
	// warmth actions, but in the Sage scenario the agent is AT the fire (warmAt),
	// so any of these firing would be a counter-schedule.
	switch goal {
	case "forage", "hunt", "build_fire", "refuel_fire":
		return true
	default:
		return false
	}
}

// sageScenario builds the post-arrival Sage state: one live villager, cold, fed,
// rested, larder below stock, standing where a planner goto_warmth just landed
// it — beside a real lit fire it knows — with a known forage tile tempting the
// larder rung. Returns state, map, and the warm tile the agent stands on.
func sageScenario(t *testing.T, seed uint64) (*State, *worldmap.Map, *Agent, int64) {
	t.Helper()
	const now int64 = 5000
	m := testMap(seed)
	s := NewState(seed, m)
	s.Night = false // daytime: the world-01 larder rung is a day rung
	isolateAgents(s)
	a := &s.Agents[0]
	a.Dead = false
	a.Asleep = false
	a.Needs = Needs{Health: 1000, Food: 600, Rest: 600, Warmth: 250, Morale: 600}
	a.Inv = Inventory{FoodRaw: 0} // larder below stockFoodRawTo

	// A real lit fire one tile east; the agent stands beside it (warmAt true).
	fx, fy := a.X+1, a.Y
	s.Structures = append(s.Structures, Structure{Kind: "fire", X: fx, Y: fy, FuelUntil: now + 100000})
	if !warmAt(s, a.X, a.Y, now) {
		t.Skipf("seed %d: agent not warm beside the placed fire", seed)
	}
	a.Map.upsertFact(PlaceFact{Kind: "fire", X: fx, Y: fy, Seen: now, Provenance: ProvenanceWitnessed, Detail: now + 100000})
	// A known forage tile ⇒ the larder rung has somewhere to counter-schedule to.
	spot, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return effectiveKind(m, s, x, y) == worldmap.Forage })
	if !ok {
		t.Skipf("seed %d: no reachable forage for the larder counter-schedule", seed)
	}
	a.Map.upsertFact(PlaceFact{Kind: "forage", X: spot.X, Y: spot.Y, Seen: now, Provenance: ProvenanceWitnessed})
	return s, m, a, now
}

func TestThrashRegressionSageShape(t *testing.T) {
	// ---- Proof 1: the OLD flip exists (doctrine's inputs nullified). ----
	s, m, a, now := sageScenario(t, 42)
	a.LastMindIntentDone = 0 // window unarmed (pre-062 had no yield window)
	a.Needs.Warmth = 600     // warmth healthy (pre-062 had no day warmth band)
	if prepYields(s, a, now) {
		t.Fatal("with the window unarmed and warmth healthy, prep must NOT yield (pre-062 world)")
	}
	if g := goalOf(decideIntent(s, m, 0, now)); g != "forage" {
		t.Fatalf("the pre-062 reflex should counter-schedule a larder forage (the flip), got %q", g)
	}

	// ---- Proof 2: the NEW hold — same cold arrival, window armed. ----
	s, m, a, now = sageScenario(t, 42)
	a.LastMindIntentDone = now - 100 // a planner goto_warmth completed 100t ago
	if !prepYields(s, a, now) {
		t.Fatal("inside the yield window, prep must yield")
	}
	if g := goalOf(decideIntent(s, m, 0, now)); isPrepGoal(g) {
		t.Fatalf("the new arbitration fired prep (%q) in the yield window — the layer fight is not over", g)
	}
}

// TestThrashRegressionDrivenRecovery is SC-001 end-to-end through the executor:
// a planner goto_warmth completes at a fire; across the whole yield window the
// reflex fires ZERO prep intents for that agent, and its warmth recovers.
func TestThrashRegressionDrivenRecovery(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.Night = false
	isolateAgents(s)
	a := &s.Agents[0]
	a.Dead = false
	a.Needs = Needs{Health: 1000, Food: 600, Rest: 600, Warmth: 250, Morale: 600}
	a.Inv = Inventory{FoodRaw: 0}

	// A real lit fire two tiles east; the warm target is the tile between.
	fx, fy := a.X+2, a.Y
	tx, ty := a.X+1, a.Y
	if !passable(m, s, tx, ty) || !passable(m, s, fx, fy) {
		t.Skip("seed 42: fire/warm tiles not passable for the driven scenario")
	}
	s.Structures = append(s.Structures, Structure{Kind: "fire", X: fx, Y: fy, FuelUntil: 200000})
	a.Map.upsertFact(PlaceFact{Kind: "fire", X: fx, Y: fy, Seen: 1, Provenance: ProvenanceWitnessed, Detail: 200000})
	spot, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return effectiveKind(m, s, x, y) == worldmap.Forage })
	if ok {
		a.Map.upsertFact(PlaceFact{Kind: "forage", X: spot.X, Y: spot.Y, Seen: 1, Provenance: ProvenanceWitnessed})
	}
	warmthBefore := a.Needs.Warmth

	// Inject a planner goto_warmth toward the warm tile at tick 1.
	cmds := map[int64][]store.Event{
		1: {{Tick: 1, Type: "agent.intent_set", Payload: mustPayload(IntentSetPayload{
			Agent: 0, Goal: "goto_warmth", TargetX: tx, TargetY: ty, Source: "planner"})}},
	}
	// Drive less than prepYieldTicks so the whole drive is inside the window.
	log := driveTicks(t, s, m, 1500, cmds)

	// Find when the planner goto_warmth completed (arms the window).
	var doneTick int64 = -1
	for _, e := range log {
		if e.Type != "agent.intent_done" {
			continue
		}
		var p AgentPayload
		json.Unmarshal(e.Payload, &p)
		if p.Agent == 0 {
			doneTick = e.Tick
			break
		}
	}
	if doneTick < 0 {
		t.Fatal("the planner goto_warmth never completed — scenario did not arm the window")
	}
	if s.Agents[0].LastMindIntentDone != doneTick {
		t.Fatalf("window armed to %d, want the completion tick %d", s.Agents[0].LastMindIntentDone, doneTick)
	}

	// SC-001: ZERO prep intents for agent 0 after arrival, inside the window.
	for _, e := range log {
		if e.Type != "agent.intent_set" || e.Tick <= doneTick {
			continue
		}
		if e.Tick-doneTick >= prepYieldTicks {
			continue
		}
		var p IntentSetPayload
		json.Unmarshal(e.Payload, &p)
		if p.Agent == 0 && isPrepGoal(p.Goal) {
			t.Fatalf("reflex fired prep %q at tick %d (%d into the window) — the thrash loop is not dead",
				p.Goal, e.Tick, e.Tick-doneTick)
		}
	}

	// The warmth trajectory recovers.
	if s.Agents[0].Needs.Warmth <= warmthBefore {
		t.Fatalf("warmth did not recover: %d -> %d", warmthBefore, s.Agents[0].Needs.Warmth)
	}
}
