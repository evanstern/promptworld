package sim

// Spec 062 US2 (T007): the daytime warmth rung — a cold-but-not-tired villager
// by day seeks/relights KNOWN warmth (AS1), builds with wood in hand (AS2), and
// leaves the day branch byte-identical when warmth is healthy (AS3). The rung
// deliberately stops before the night ladder's chop tail (the flagged deviation
// on dayWarmthLadder): by day warmth passively regenerates, so trekking to chop
// firewood is unjustified subsistence-time theft — proven by the no-chop case.

import (
	"testing"

	"github.com/evanstern/promptworld/internal/worldmap"
)

// dayColdAgent configures agent 0 as an idle, fed, RESTED, DAYTIME villager
// whose warmth is in the danger band with nothing warm underfoot — the state in
// which the new day warmth rung is the sole decider (it runs after the nap rung,
// which a rested agent never triggers, and before all prep).
func dayColdAgent(t *testing.T, seed uint64) (*State, *worldmap.Map, *Agent, int64) {
	t.Helper()
	const now int64 = 5000
	m := testMap(seed)
	s := NewState(seed, m)
	s.Night = false
	a := &s.Agents[0]
	a.Dead = false
	a.Asleep = false
	// Fed and rested (skip the hunger + nap survival rungs), warmth in its
	// danger band, nothing warm underfoot.
	a.Needs = Needs{Health: 1000, Food: 600, Rest: 600, Warmth: dangerWarmthBelow - 50, Morale: 600}
	a.Inv = Inventory{}
	return s, m, a, now
}

// TestDayWarmthSeeksKnownWarmth (US2 AS1): day, cold, reachable known warmth ⇒
// the reflex walks to the fire it remembers as lit, not forage/wander.
func TestDayWarmthSeeksKnownWarmth(t *testing.T) {
	s, m, a, now := dayColdAgent(t, 42)
	grantKnownLitFire(a, now) // a fire remembered as lit on the agent's tile

	if g := goalOf(decideIntent(s, m, 0, now)); g != "goto_warmth" {
		t.Fatalf("day + cold + known warmth: reflex chose %q, want goto_warmth", g)
	}
}

// TestDayWarmthRelightsKnownDyingFire (US2 AS1, the refuel arm): day, cold, a
// KNOWN cold/dying fire and wood in hand ⇒ relight it (cheaper than a build).
func TestDayWarmthRelightsKnownDyingFire(t *testing.T) {
	s, m, a, now := dayColdAgent(t, 42)
	a.Inv.Wood = 3 // enough to build, but relighting a known fire wins
	grantKnownColdFire(a, now)

	if g := goalOf(decideIntent(s, m, 0, now)); g != "refuel_fire" {
		t.Fatalf("day + cold + known dying fire + wood: reflex chose %q, want refuel_fire", g)
	}
}

// TestDayWarmthBuildsWithWood (US2 AS2): day, cold, no known warmth, wood ≥ build
// cost ⇒ build a fire (the day mirror of the night build rung).
func TestDayWarmthBuildsWithWood(t *testing.T) {
	s, m, a, now := dayColdAgent(t, 42)
	a.Inv.Wood = fireWoodCost // no fire facts granted ⇒ nothing known to reach

	if g := goalOf(decideIntent(s, m, 0, now)); g != "build_fire" {
		t.Fatalf("day + cold + no known warmth + wood: reflex chose %q, want build_fire", g)
	}
}

// TestDayWarmthHealthyIsByteIdenticalToPre062 (US2 AS3): with warmth healthy the
// day warmth rung never fires, so the day branch behaves exactly as today —
// rest → prep → wander. Here: fed, rested, warm, empty larder with known forage
// ⇒ the larder prep rung forages, exactly as pre-062.
func TestDayWarmthHealthyIsByteIdenticalToPre062(t *testing.T) {
	s, m, _, now := prepAgent(t, 42) // warmth 600, window unarmed
	if g := goalOf(decideIntent(s, m, 0, now)); g != "forage" {
		t.Fatalf("day + warmth healthy: reflex chose %q, want the pre-062 larder forage", g)
	}
}

// TestDayWarmthDoesNotChopTheDeviation pins the flagged deviation (dayWarmthLadder
// omits the night ladder's chop tail): day, cold, no known warmth, NO wood, but a
// KNOWN choppable tree ⇒ the day rung does NOT chop (unlike night); the warmth
// danger band then holds prep, so the villager wanders while it passively rewarms.
// The night ladder, by contrast, WOULD chop here (proven by TestColdNightReflexMatrix).
func TestDayWarmthDoesNotChopTheDeviation(t *testing.T) {
	s, m, a, now := dayColdAgent(t, 42)
	a.Inv.Wood = 0
	if !grantKnownTree(s, m, a, now) {
		t.Skip("seed 42 has no reachable tree for the no-chop case")
	}
	// An empty larder + known forage would tempt prep — but warmth is in danger,
	// so prep yields and the day rung (no chop) leaves only wander.
	a.Inv.FoodRaw = 0
	spot, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return effectiveKind(m, s, x, y) == worldmap.Forage })
	if ok {
		a.Map.upsertFact(PlaceFact{Kind: "forage", X: spot.X, Y: spot.Y, Seen: now, Provenance: ProvenanceWitnessed})
	}

	g := goalOf(decideIntent(s, m, 0, now))
	if g == "chop" {
		t.Fatal("the DAY warmth rung must not chop for firewood (the flagged deviation)")
	}
	if g == "forage" {
		t.Fatal("warmth in the danger band must hold the larder prep forage")
	}
	if g != "wander" {
		t.Fatalf("day + cold + no wood + no reachable known warmth: want wander, got %q", g)
	}
}
