package sim

// spec 057 / TASK-108 US3 (SC-003): the cold build-fire reflex, proven cell by
// cell against the CURRENT ladder with NO planner/model involvement — every
// assertion calls decideIntent directly and checks the resulting reflex Goal.
//
// The matrix is cold-night × {wood level} × {warmth knowledge}:
//
//	                | warmth known-reachable | none known    | known-but-stale/dead
//	  wood ≥ 2      | goto_warmth            | build_fire     | refuel_fire
//	  wood = 1 +tree| goto_warmth            | chop           | refuel_fire
//	  wood = 0      | goto_warmth            | search         | search
//
// Doctrine reading of each column:
//   - known-reachable warmth: walk to the fire you remember as lit (cheapest).
//   - none known: make warmth — build with wood in hand, else go get the wood
//     (chop) if a tree is known; with NEITHER, spec 062 US3 (057 audit Gap A)
//     now SEARCHES toward the frontier for warmth/wood rather than lying down
//     cold — sleep is the floor only when no frontier is reachable either. The
//     reflexAgent's genesis map on seed 42 has a reachable frontier, so the
//     wood=0/nothing-known cells resolve to search (was sleep pre-062); the
//     no-frontier fall-through to sleep is pinned by TestNightSearchFallback.
//   - known-but-stale (a fire you remember as cold/out): relight it with any
//     wood you carry (refuel beats a fresh build); with no wood and no tree,
//     the US3 frontier search is again the floor above sleep. The refuel rung
//     is what closes the "stale warmth belief" candidate gap named in spec 057
//     US3 — the agent acts on the fire it knows rather than livelocking on a
//     warmth it can't reach.

import (
	"testing"

	"github.com/evanstern/promptworld/internal/worldmap"
)

// reflexAgent configures agent 0 of a fresh world as an idle, fed, awake
// villager on a cold night with no warmth underfoot — the state in which the
// night reflex ladder is the sole decider. Returns the state and the agent.
func reflexAgent(t *testing.T, seed uint64) (*State, *worldmap.Map, *Agent) {
	t.Helper()
	m := testMap(seed)
	s := NewState(seed, m)
	s.Night = true
	a := &s.Agents[0]
	a.Dead = false
	a.Asleep = false
	// Fed (skip the eat / get-food rungs that precede the night branch) and
	// rested; nothing warm underfoot (no structure placed on the agent's tile,
	// so warmAt is false and the cold branch is entered).
	a.Needs = Needs{Health: 1000, Food: 600, Rest: 600, Warmth: 300, Morale: 600}
	a.Inv = Inventory{}
	return s, m, a
}

// grantKnownLitFire gives the agent a fresh fact of a fire it remembers as LIT
// (Detail ahead of now) on its own tile — warmth it can plan on. No real
// structure is placed, so warmAt (ground truth) stays false and the reflex must
// choose goto_warmth to reach the remembered fire.
func grantKnownLitFire(a *Agent, now int64) {
	a.Map.upsertFact(PlaceFact{Kind: "fire", X: a.X, Y: a.Y, Seen: now,
		Provenance: ProvenanceWitnessed, Detail: now + 100000})
}

// grantKnownColdFire gives the agent a fresh fact of a fire it remembers as
// COLD/dying (Detail at or behind now) on its own tile — known, but not warmth
// it can plan on. The refuel rung keys on this (relight the fire you remember).
func grantKnownColdFire(a *Agent, now int64) {
	a.Map.upsertFact(PlaceFact{Kind: "fire", X: a.X, Y: a.Y, Seen: now,
		Provenance: ProvenanceWitnessed, Detail: now})
}

// grantKnownTree finds a real standing tree, plants it into ground truth if the
// seed's spawn lacks one nearby, and grants the agent a fresh fact of it — a
// choppable tree the wood-acquisition rung can resolve. Returns false if no
// passable adjacency exists (caller skips).
func grantKnownTree(s *State, m *worldmap.Map, a *Agent, now int64) bool {
	tx, ty, ok := nearestTreeTile(m, s, a)
	if !ok {
		return false
	}
	a.Map.upsertFact(PlaceFact{Kind: "tree", X: tx, Y: ty, Seen: now, Provenance: ProvenanceWitnessed})
	return true
}

func nearestTreeTile(m *worldmap.Map, s *State, a *Agent) (int, int, bool) {
	// Tree tiles are impassable (you stand ADJACENT to chop), so the nearest
	// passable-only search never lands on one — use the adjacent-stand search
	// chopIntent itself uses, and grant the fact at the tree (res) tile.
	_, res, ok := nearestAdjacentTo(m, s, a.X, a.Y, func(x, y int) bool {
		return m.InBounds(x, y) && effectiveKind(m, s, x, y) == worldmap.Tree
	})
	return res.X, res.Y, ok
}

func TestColdNightReflexMatrix(t *testing.T) {
	const now int64 = 100

	// warmth-knowledge setup per column.
	const (
		warmKnownReachable = "known-reachable"
		warmNoneKnown      = "none-known"
		warmKnownStale     = "known-but-stale"
	)
	// wood setup per row.
	const (
		woodTwo     = "wood>=2"
		woodOneTree = "wood=1+tree"
		woodZero    = "wood=0"
	)

	cases := []struct {
		warmth string
		wood   string
		want   string
	}{
		{warmKnownReachable, woodTwo, "goto_warmth"},
		{warmKnownReachable, woodOneTree, "goto_warmth"},
		{warmKnownReachable, woodZero, "goto_warmth"},

		{warmNoneKnown, woodTwo, "build_fire"},
		{warmNoneKnown, woodOneTree, "chop"},
		{warmNoneKnown, woodZero, "search"}, // spec 062 US3: was "sleep"; Gap A frontier search

		{warmKnownStale, woodTwo, "refuel_fire"},
		{warmKnownStale, woodOneTree, "refuel_fire"},
		{warmKnownStale, woodZero, "search"}, // spec 062 US3: was "sleep"; Gap A frontier search
	}

	for _, c := range cases {
		name := c.warmth + "/" + c.wood
		t.Run(name, func(t *testing.T) {
			s, m, a := reflexAgent(t, 42)

			// Wood dimension.
			switch c.wood {
			case woodTwo:
				a.Inv.Wood = 2
			case woodOneTree:
				a.Inv.Wood = 1
				if !grantKnownTree(s, m, a, now) {
					t.Skip("seed 42 has no reachable tree for the chop cell")
				}
			case woodZero:
				a.Inv.Wood = 0
			}

			// Warmth-knowledge dimension.
			switch c.warmth {
			case warmKnownReachable:
				grantKnownLitFire(a, now)
			case warmNoneKnown:
				// no fire facts
			case warmKnownStale:
				grantKnownColdFire(a, now)
			}

			d := decideIntent(s, m, 0, now)

			// Reflex-only: never a direct event, never a nil where an intent is
			// expected.
			if d.directEvent != "" {
				t.Fatalf("%s: reflex produced a direct event %q, want intent %q", name, d.directEvent, c.want)
			}
			if d.intent == nil {
				t.Fatalf("%s: reflex produced no intent, want %q", name, c.want)
			}
			if d.intent.Goal != c.want {
				t.Fatalf("%s: reflex chose %q, want %q", name, d.intent.Goal, c.want)
			}
		})
	}
}

// TestColdNightRefuelBeatsBuildForKnownColdFire pins the doctrine tie-break the
// matrix's known-but-stale/wood≥2 cell asserts: with wood enough to build AND a
// known cold fire in reach, the reflex relights the known fire (refuel) rather
// than building a second one — the cheaper survival move, and the guard against
// a village sprouting redundant fires next to dead ones.
func TestColdNightRefuelBeatsBuildForKnownColdFire(t *testing.T) {
	const now int64 = 100
	s, m, a := reflexAgent(t, 42)
	a.Inv.Wood = 3 // plenty to build
	grantKnownColdFire(a, now)

	d := decideIntent(s, m, 0, now)
	if d.intent == nil || d.intent.Goal != "refuel_fire" {
		t.Fatalf("with wood and a known cold fire, reflex should refuel (not build); got %+v", d.intent)
	}
}
