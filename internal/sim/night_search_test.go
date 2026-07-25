package sim

// Spec 062 US3 (T009): the bounded night frontier-search fallback (057 audit
// Gap A). A cold villager at night who knows no warmth, carries insufficient
// wood, and finds nothing to chop searches toward the frontier instead of
// lying down cold — but only one rung above terminal sleep: with no reachable
// frontier either, it still sleeps (today's floor of the fallback).

import "testing"

// TestNightSearchFallbackSearchesWhenFrontierReachable (US3 AS1): cold night, no
// known warmth, no wood, nothing choppable known, a reachable frontier ⇒ the
// reflex searches toward the frontier rather than sleeping cold.
func TestNightSearchFallbackSearchesWhenFrontierReachable(t *testing.T) {
	const now int64 = 100
	s, m, a := reflexAgent(t, 42) // cold night, warmth underfoot false
	a.Inv.Wood = 0                // insufficient to build
	// No fire facts (no known warmth), no tree facts (nothing choppable). The
	// genesis map on seed 42 has explored terrain with a reachable frontier.
	if _, ok := nearestFrontier(m, s, a); !ok {
		t.Skip("seed 42 genesis has no reachable frontier for the search case")
	}

	if g := goalOf(decideIntent(s, m, 0, now)); g != "search" {
		t.Fatalf("cold night, nothing known/carried/choppable, frontier reachable: chose %q, want search", g)
	}
}

// TestNightSearchFallbackSleepsWithoutFrontier (US3 AS2): the same cold, empty
// villager but with NO reachable frontier (its mental map holds no explored
// tile bordering the unknown) ⇒ sleep, today's terminal floor — the fallback of
// the fallback.
func TestNightSearchFallbackSleepsWithoutFrontier(t *testing.T) {
	const now int64 = 100
	s, m, a := reflexAgent(t, 42)
	a.Inv.Wood = 0
	// Erase the explored map: with nothing explored, no frontier exists.
	a.Map.Explored = ""
	if _, ok := nearestFrontier(m, s, a); ok {
		t.Fatal("clearing Explored should leave no reachable frontier")
	}

	if g := goalOf(decideIntent(s, m, 0, now)); g != "sleep" {
		t.Fatalf("cold night, nothing known, no frontier: chose %q, want the terminal sleep", g)
	}
}

// TestNightSearchDoesNotFireWithWoodOrTree pins the gate: the search rung is one
// rung above sleep and BELOW the make-warmth rungs, so a villager who can still
// build (wood in hand) or chop (a known tree) does that instead of searching —
// search is only the truly-nothing floor.
func TestNightSearchDoesNotFireWithWoodOrTree(t *testing.T) {
	const now int64 = 100

	// Wood in hand ⇒ build, not search.
	s, m, a := reflexAgent(t, 42)
	a.Inv.Wood = fireWoodCost
	if g := goalOf(decideIntent(s, m, 0, now)); g != "build_fire" {
		t.Fatalf("wood in hand should build, not search; got %q", g)
	}

	// A known tree ⇒ chop, not search.
	s, m, a = reflexAgent(t, 42)
	a.Inv.Wood = 1
	if !grantKnownTree(s, m, a, now) {
		t.Skip("seed 42 has no reachable tree")
	}
	if g := goalOf(decideIntent(s, m, 0, now)); g != "chop" {
		t.Fatalf("a known tree should chop, not search; got %q", g)
	}
}
