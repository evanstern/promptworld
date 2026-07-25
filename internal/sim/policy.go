package sim

import (
	"fmt"

	"github.com/evanstern/promptworld/internal/worldmap"
)

// The reflex policy: a deterministic pure function deciding what an idle,
// awake agent does next. It is the stand-in for the LLM planner until TASK-7
// — and stays forever as the degraded-mode fallback (the executor must keep
// bodies alive when no planner thoughts arrive).
//
// Priority order: eat → get food → survive the night (fire/warmth) → rest →
// prepare (wood, fire, shelter, stockpile) → wander.
//
// It returns either a direct event (eating happens instantly) or an intent.

type decision struct {
	directEvent string // "agent.ate" or ""
	intent      *Intent
}

// decideIntent is the reflex ladder, explicitly split into two classified rung
// groups (spec 062 FR-001, R2 — the 057 audit's rung table is the authority):
//
//   - SURVIVAL (survivalDecision) — instinct that keeps the body alive: eat,
//     get food, the night warmth ladder + terminal sleep, and the daytime nap.
//     These run first and are UNCONDITIONED by the yield window and danger
//     bands (a life-saving reflex never defers).
//   - PREP (prepDecision) — opportunistic village upkeep: first-fire prep,
//     the dying-fire refuel top-up, and the larder stock-up. These are
//     "instinct that yields to intelligence": once T004 lands they defer to a
//     recent non-reflex intent and to any need in its danger band.
//
// The classification lives in this function STRUCTURE (FR-001), not in prose —
// which rungs are survival vs prep is exactly which helper they live in. When
// neither group decides, the agent wanders (idle filler, neither survival nor
// prep).
func decideIntent(s *State, m *worldmap.Map, idx int, tick int64) decision {
	a := &s.Agents[idx]
	if d, ok := survivalDecision(s, m, a, tick); ok {
		return d
	}
	if d, ok := prepDecision(s, m, a, tick); ok {
		return d
	}
	return wanderDecision(s, m, a, idx, tick)
}

// survivalDecision is the SURVIVAL rung group (spec 062 R2): the life-saving
// instinct that runs before — and unconditioned by — the PREP gate. It reports
// (decision, true) when a survival rung owns the agent this round, else
// (zero, false) to fall through to prep. The night branch always decides
// (warmth ladder, then terminal sleep); by day it decides only for hunger and
// exhaustion. Behavior is byte-identical to the pre-062 ladder's rungs 1–4.
func survivalDecision(s *State, m *worldmap.Map, a *Agent, tick int64) (decision, bool) {
	// Eat from inventory the moment hunger bites (most-nutritious-first, T018).
	if a.Needs.Food < hungryAt && hasAnyFood(a) {
		return decision{directEvent: "agent.ate"}, true
	}

	// Hungry with nothing carried: get food (forage first, hunt as backup),
	// or — knowing of no available food source — go LOOKING for one (spec 041
	// US4 T026, FR-013 parity without omniscience). The search fallback is
	// hunger-only: the larder top-up in prep stays opportunistic (a fed
	// villager doesn't mount expeditions — constant frontier treks desync the
	// village from its forage-regrow rotation and starve marginal maps).
	if a.Needs.Food < hungryAt {
		if d, ok := foodIntent(s, m, a, tick); ok {
			return decision{intent: d}, true
		}
		if p, ok := nearestFrontier(m, s, a); ok {
			return decision{intent: &Intent{Goal: "search", TargetX: p.X, TargetY: p.Y}}, true
		}
	}

	// Spec 041 (US1/T014, reflex parity): every rung below resolves targets
	// through the SAME map predicates the goal resolvers use — no omniscient
	// fallback (clarify Q2). Standing warmth (warmAt at the agent's own tile)
	// stays ground truth: feeling warm where you stand is perception, not
	// place knowledge.
	warmKnown, _ := warmKnownPredicate(a, tick)
	if s.Night {
		if !warmAt(s, a.X, a.Y, tick) {
			// Reach KNOWN warmth, or make it, or get the wood to make it.
			if p, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return warmKnown(x, y) && passable(m, s, x, y) }); ok {
				return decision{intent: &Intent{Goal: "goto_warmth", TargetX: p.X, TargetY: p.Y}}, true
			}
			// The one reflex addition (T020, FR-012): a cold/dying KNOWN fire
			// and wood in hand — relight it (cheaper than a fresh build).
			if in, ok := reflexRefuelIntent(s, m, a, tick); ok {
				return decision{intent: in}, true
			}
			if a.Inv.Wood >= fireWoodCost {
				if p, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return buildSite(m, s, x, y) }); ok {
					return decision{intent: &Intent{Goal: "build_fire", TargetX: p.X, TargetY: p.Y}}, true
				}
			}
			if in, ok := chopIntent(s, m, a, tick); ok {
				return decision{intent: in}, true
			}
		}
		// Warm (or nothing to be done about it): sleep where you stand.
		return decision{intent: &Intent{Goal: "sleep", TargetX: a.X, TargetY: a.Y}}, true
	}

	// Daytime. Exhausted agents nap somewhere warm if possible.
	if a.Needs.Rest < tiredAt {
		tx, ty := a.X, a.Y
		if !warmAt(s, tx, ty, tick) {
			if p, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return warmKnown(x, y) && passable(m, s, x, y) }); ok {
				tx, ty = p.X, p.Y
			}
		}
		return decision{intent: &Intent{Goal: "sleep", TargetX: tx, TargetY: ty}}, true
	}

	return decision{}, false
}

// prepDecision is the PREP rung group (spec 062 R2): opportunistic daytime
// village upkeep — a fire before the first night, then keep it burning, then a
// full larder. Shelter-building is planner-only (T020, FR-012): it costs planks,
// and the reflex never enters the crafting economy. Spec 041: "the village has
// a fire" is now the agent's own belief — it builds one when it KNOWS of none,
// not when the world holds none. Reports (decision, true) when a prep rung
// fires, else (zero, false). Behavior is byte-identical to the pre-062 ladder's
// rungs 5–7 (T004 adds the yield/danger gate in front of it).
func prepDecision(s *State, m *worldmap.Map, a *Agent, tick int64) (decision, bool) {
	if !knowsAnyFresh(a, "fire", tick) {
		if a.Inv.Wood >= fireWoodCost {
			if p, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return buildSite(m, s, x, y) }); ok {
				return decision{intent: &Intent{Goal: "build_fire", TargetX: p.X, TargetY: p.Y}}, true
			}
		}
		if d, ok := chopIntent(s, m, a, tick); ok {
			return decision{intent: d}, true
		}
	}
	// Keep the fire alive (T020): top up a dying/cold fire while carrying wood.
	if in, ok := reflexRefuelIntent(s, m, a, tick); ok {
		return decision{intent: in}, true
	}
	if a.Inv.FoodRaw < stockFoodRawTo {
		if d, ok := foodIntent(s, m, a, tick); ok {
			return decision{intent: d}, true
		}
	}
	return decision{}, false
}

// wanderDecision is the idle filler (neither survival nor prep): a little
// seeded, tick-pure wander when no rung above owns the agent. Byte-identical to
// the pre-062 ladder's terminal wander block.
func wanderDecision(s *State, m *worldmap.Map, a *Agent, idx int, tick int64) decision {
	r := rngAt(s.Seed, "wander", tick, idx)
	for try := 0; try < 8; try++ {
		dx := int(r.Uint64N(9)) - 4
		dy := int(r.Uint64N(9)) - 4
		if dx == 0 && dy == 0 {
			continue
		}
		if passable(m, s, a.X+dx, a.Y+dy) {
			return decision{intent: &Intent{Goal: "wander", TargetX: a.X + dx, TargetY: a.Y + dy}}
		}
	}
	return decision{} // stay idle this round
}

// hasAnyFood reports whether the agent carries any edible unit (T018): the
// eat/wake checks key on the full triplet, not raw food alone.
func hasAnyFood(a *Agent) bool {
	return a.Inv.Meals+a.Inv.FoodCooked+a.Inv.FoodRaw > 0
}

// reflexRefuelIntent is the reflex's one new rule (T020, FR-012): when carrying
// wood, refuel the nearest KNOWN fire the agent remembers as cold or dying —
// remembered Detail (FuelUntil as last seen) under refuelDyingBelow from now
// (spec 041: the agent predicts burnout from its own knowledge; a fire someone
// refueled out of view is still worth the walk — the completion no-ops at the
// cap). Returns no intent when the agent has no wood or knows no such fire —
// the reflex never chops just to refuel.
func reflexRefuelIntent(s *State, m *worldmap.Map, a *Agent, tick int64) (*Intent, bool) {
	if a.Inv.Wood < 1 {
		return nil, false
	}
	if p, ok := nearestKnown(m, s, a, "fire", tick, func(x, y int) bool {
		f, _ := knownFreshFact(a, "fire", x, y, tick)
		return f.Detail-tick < s.RefuelDyingBelow()
	}); ok {
		return &Intent{Goal: "refuel_fire", TargetX: p.X, TargetY: p.Y}, true
	}
	return nil, false
}

// foodIntent is the reflex's shared food lookup: the nearest KNOWN forage
// spot, else the nearest KNOWN ready den (spec 041 T014 — same predicates as
// the forage/hunt resolvers). The US4 search fallback lives on the HUNGRY
// call site in decideIntent, not here — the larder rung shares this lookup
// but never searches.
func foodIntent(s *State, m *worldmap.Map, a *Agent, tick int64) (*Intent, bool) {
	if p, ok := nearestKnown(m, s, a, "forage", tick, func(x, y int) bool {
		return effectiveKind(m, s, x, y) == worldmap.Forage
	}); ok {
		return &Intent{Goal: "forage", TargetX: p.X, TargetY: p.Y}, true
	}
	if p, ok := nearestKnown(m, s, a, "den", tick, func(x, y int) bool {
		return denReadyAt(s, x, y, tick)
	}); ok {
		return &Intent{Goal: "hunt", TargetX: p.X, TargetY: p.Y}, true
	}
	return nil, false
}

// goalResolver produces a concrete, deterministic intent (or a direct-event
// tag, or an error) for one goal against live state. The signature carries
// everything the per-verb bodies need; unused params in a given resolver are
// fine (the bodies are the old switch arms, moved verbatim).
type goalResolver func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error)

// goalResolvers is the name-keyed resolution table (spec 014, R2): the former
// resolveGoal switch, one arm per entry. It replaces the switch so startup can
// verify — against the tool registry — that every world tool has a resolver
// (internal/sim/toolcheck.go). The per-verb semantics are byte-identical to the
// old arms; only the dispatch shape changed. The reflex ladder (decideIntent)
// is keyed by intent, not by this table, and stays hand-written (R6).
var goalResolvers = buildGoalResolvers()

func buildGoalResolvers() map[string]goalResolver {
	// craft handles the three hand-crafts (planks/stone/spear), which shared
	// one switch arm — keyed on `goal`.
	craft := func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
		// T026: hand-crafts anywhere — target is the agent's own tile, no
		// travel. Planner-only (FR-020); never enters the reflex ladder.
		r, ok := recipeFor(goal)
		if !ok {
			return nil, "", fmt.Errorf("unknown recipe %q", goal)
		}
		if !hasItems(a.Inv, r.Inputs) {
			return nil, "", fmt.Errorf("%s lacks inputs for %s", a.Name, goal)
		}
		return &Intent{Goal: goal, TargetX: a.X, TargetY: a.Y}, "", nil
	}
	// wallBuild resolves both wall builds (spec 032 US1, research R2). Unlike
	// fire/shelter/oven/chest (which build on the tile the agent stands on),
	// walls build ADJACENT: the builder stands on a passable tile (Target) and
	// the wall lands on the neighboring buildable tile (Res). Building where you
	// stand would entomb the builder the instant the wall lands (FR-007), so
	// nearestAdjacentTo over buildSite gives the stand/build pair — the chop/
	// quarry adjacency pattern.
	wallBuild := func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
		r, ok := recipeFor(goal)
		if !ok {
			return nil, "", fmt.Errorf("unknown recipe %q", goal)
		}
		if !hasItems(a.Inv, r.Inputs) {
			return nil, "", fmt.Errorf("%s lacks inputs for %s", a.Name, goal)
		}
		if stand, res, ok := nearestAdjacentTo(m, s, a.X, a.Y, func(x, y int) bool { return buildSite(m, s, x, y) }); ok {
			return &Intent{Goal: goal, TargetX: stand.X, TargetY: stand.Y, ResX: res.X, ResY: res.Y}, "", nil
		}
		return nil, "", fmt.Errorf("no build site reachable")
	}
	// talk resolves both talk_to (the planner verb) and its internal "seek"
	// alias — they shared one switch arm. Spec 041 (T013): the walk target is
	// the acting agent's LAST SIGHTING of the target (its mental map's peer
	// record), never the target's live coordinates — a stale sighting resolves
	// honestly to where the target was last seen, and the landing/arrival
	// guards (GuardTargetPresent) cover the miss. The Dead check stays on live
	// state: the landing door's GuardTargetAlive gates it identically, and
	// death-knowledge honesty is beyond this feature's place-fact scope.
	talk := func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
		if targetAgent < 0 || targetAgent >= len(s.Agents) || targetAgent == idx {
			return nil, "", fmt.Errorf("no such agent to seek")
		}
		t := &s.Agents[targetAgent]
		if t.Dead {
			return nil, "", fmt.Errorf("%s is dead", t.Name)
		}
		sight, ok := peerSightingOf(a, targetAgent)
		if !ok {
			return nil, "", fmt.Errorf("you do not know where %s is", t.Name)
		}
		return &Intent{Goal: "seek", TargetX: sight.X, TargetY: sight.Y}, "", nil
	}

	return map[string]goalResolver{
		"eat": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// T018: eat over the full food triplet, most-nutritious-first to
			// satiety. Refuse if empty-handed or already sated (no unit is ever
			// consumed at satiety — the eating-overshoot edge case).
			if !hasAnyFood(a) {
				return nil, "", fmt.Errorf("%s has nothing to eat", a.Name)
			}
			if a.Needs.Food >= satietyAt {
				return nil, "", fmt.Errorf("%s is already sated", a.Name)
			}
			return nil, "agent.ate", nil
		},
		"forage": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// Spec 041 (US1, research D3): knowledge-gated — the candidate set is
			// the agent's fresh forage facts, not the world's forage tiles.
			// AVAILABILITY at a known place (the harvested overlay) stays a
			// ground condition on top, the den-readiness class: regrowth is not
			// in the fact model, and without it the reflex livelocks on the
			// nearest just-harvested spot (its fact persists until US3's
			// correction). Knowing none is the epistemic failure (contracts §4),
			// distinct from reaching none.
			if !knowsAnyFresh(a, "forage", tick) {
				return nil, "", fmt.Errorf("you know of no forage")
			}
			if p, ok := nearestKnown(m, s, a, "forage", tick, func(x, y int) bool {
				return effectiveKind(m, s, x, y) == worldmap.Forage
			}); ok {
				return &Intent{Goal: "forage", TargetX: p.X, TargetY: p.Y}, "", nil
			}
			return nil, "", fmt.Errorf("no forage reachable")
		},
		"hunt": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// Spec 041: gated on known dens; readiness stays a ground condition
			// on top (cooldowns are not in the fact model — the same class as
			// chest contents below).
			if !knowsAnyFresh(a, "den", tick) {
				return nil, "", fmt.Errorf("you know of no dens")
			}
			if p, ok := nearestKnown(m, s, a, "den", tick, func(x, y int) bool {
				return denReadyAt(s, x, y, tick)
			}); ok {
				return &Intent{Goal: "hunt", TargetX: p.X, TargetY: p.Y}, "", nil
			}
			return nil, "", fmt.Errorf("no ready den reachable")
		},
		"chop": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			if !knowsAnyFresh(a, "tree", tick) {
				return nil, "", fmt.Errorf("you know of no trees")
			}
			if in, ok := chopIntent(s, m, a, tick); ok {
				return in, "", nil
			}
			return nil, "", fmt.Errorf("no tree reachable")
		},
		"quarry": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// Planner-only (research R5, FR-020): never added to decideIntent's
			// reflex ladder. Spec 041: gated on known outcrops; depletion (the
			// Quarried overlay) stays a ground condition — availability, not
			// place knowledge (forage's harvested-overlay rationale).
			if !knowsAnyFresh(a, "rock", tick) {
				return nil, "", fmt.Errorf("you know of no rock outcrops")
			}
			if stand, res, ok := nearestKnownAdjacentTo(m, s, a, "rock", tick, func(x, y int) bool {
				return effectiveKind(m, s, x, y) == worldmap.Rock
			}); ok {
				return &Intent{Goal: "quarry", TargetX: stand.X, TargetY: stand.Y, ResX: res.X, ResY: res.Y}, "", nil
			}
			return nil, "", fmt.Errorf("no rock outcrop reachable")
		},
		"collect_water": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// Planner-only, same as quarry. Spec 041: water_edge facts sit on the
			// shoreline water tiles, matching the old matchRes coordinates.
			if !knowsAnyFresh(a, "water_edge", tick) {
				return nil, "", fmt.Errorf("you know of no water")
			}
			if stand, res, ok := nearestKnownAdjacentTo(m, s, a, "water_edge", tick, nil); ok {
				return &Intent{Goal: "collect_water", TargetX: stand.X, TargetY: stand.Y, ResX: res.X, ResY: res.Y}, "", nil
			}
			return nil, "", fmt.Errorf("no water reachable")
		},
		"build_fire": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			if a.Inv.Wood < fireWoodCost {
				return nil, "", fmt.Errorf("%s lacks wood (%d < %d)", a.Name, a.Inv.Wood, fireWoodCost)
			}
			if p, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return buildSite(m, s, x, y) }); ok {
				return &Intent{Goal: goal, TargetX: p.X, TargetY: p.Y}, "", nil
			}
			return nil, "", fmt.Errorf("no build site reachable")
		},
		"build_shelter": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// T036: re-costed to planks (was wood) — shelter joins the plank
			// economy; planner-only (FR-012), never reflex-chosen.
			if a.Inv.Planks < shelterPlankCost {
				return nil, "", fmt.Errorf("%s lacks planks (%d < %d)", a.Name, a.Inv.Planks, shelterPlankCost)
			}
			if p, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return buildSite(m, s, x, y) }); ok {
				return &Intent{Goal: goal, TargetX: p.X, TargetY: p.Y}, "", nil
			}
			return nil, "", fmt.Errorf("no build site reachable")
		},
		"build_oven": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// T030: the flagship stone-cost station, on-site like fire/shelter.
			r, _ := recipeFor("build_oven")
			if !hasItems(a.Inv, r.Inputs) {
				return nil, "", fmt.Errorf("%s lacks inputs for an oven (%d refined stone + %d planks)", a.Name, r.Inputs[0].N, r.Inputs[1].N)
			}
			if p, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return buildSite(m, s, x, y) }); ok {
				return &Intent{Goal: goal, TargetX: p.X, TargetY: p.Y}, "", nil
			}
			return nil, "", fmt.Errorf("no build site reachable")
		},
		"build_chest": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// T023 (spec 013 US3): the first owned container — 6 planks on the
			// nearest buildable tile (build_oven pattern; the pile-tile exclusion
			// already lives in buildSite, T019). Planner/plan-only (FR-014). Timed
			// work; the completion re-validates inputs + site (contested pattern).
			r, _ := recipeFor("build_chest")
			if !hasItems(a.Inv, r.Inputs) {
				return nil, "", fmt.Errorf("%s lacks planks for a chest (%d < %d)", a.Name, a.Inv.Planks, r.Inputs[0].N)
			}
			if p, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return buildSite(m, s, x, y) }); ok {
				return &Intent{Goal: goal, TargetX: p.X, TargetY: p.Y}, "", nil
			}
			return nil, "", fmt.Errorf("no build site reachable")
		},
		"craft_planks": craft,
		"craft_stone":  craft,
		"craft_spear":  craft,
		"craft_axe":    craft, // spec 032 US2: same shared hand-craft closure
		"build_path": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// Spec 032 US3 (research R3): a path is built ON the tile the agent
			// stands on (stand-on-target, the build_fire pattern) — paths are
			// walkable, so there is no entombment risk, unlike walls. The generic
			// build completion + reducer arm handle the rest.
			r, _ := recipeFor("build_path")
			if !hasItems(a.Inv, r.Inputs) {
				return nil, "", fmt.Errorf("%s lacks stone for a path (%d < %d)", a.Name, a.Inv.Stone, pathStoneCost)
			}
			if p, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return buildSite(m, s, x, y) }); ok {
				return &Intent{Goal: goal, TargetX: p.X, TargetY: p.Y}, "", nil
			}
			return nil, "", fmt.Errorf("no build site reachable")
		},
		"build_wall_plank": wallBuild,
		"build_wall_stone": wallBuild,
		"demolish": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// Spec 032 US1 (research R5): tear down the nearest wall. Adjacent-
			// stand (wall tiles are impassable), so the adjacent search gives the
			// stand tile (Target) beside the wall tile (Res). No material needed
			// to demolish. Spec 041: gated on KNOWN walls (either kind).
			if !knowsAnyFresh(a, "wall_plank", tick) && !knowsAnyFresh(a, "wall_stone", tick) {
				return nil, "", fmt.Errorf("you know of no walls")
			}
			knownWall := func(x, y int) bool {
				return knownFactAt(a, "wall_plank", x, y, tick) || knownFactAt(a, "wall_stone", x, y, tick)
			}
			if stand, res, ok := nearestAdjacentTo(m, s, a.X, a.Y, knownWall); ok {
				return &Intent{Goal: "demolish", TargetX: stand.X, TargetY: stand.Y, ResX: res.X, ResY: res.Y}, "", nil
			}
			return nil, "", fmt.Errorf("no wall reachable to demolish")
		},
		"repair": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// Spec 032 US1 (research R5): mend the nearest DAMAGED wall the agent
			// can afford — HP below the derived max AND at least 1 unit of that
			// wall's build material carried. A wall already at full health never
			// resolves. Adjacent-stand, like demolish. Spec 041: gated on known
			// walls; HP/material stay ground conditions (damage is not in the
			// fact model — a mended-behind-your-back wall no-ops at arrival).
			if !knowsAnyFresh(a, "wall_plank", tick) && !knowsAnyFresh(a, "wall_stone", tick) {
				return nil, "", fmt.Errorf("you know of no walls")
			}
			if stand, res, ok := nearestAdjacentTo(m, s, a.X, a.Y, func(x, y int) bool {
				if !knownFactAt(a, "wall_plank", x, y, tick) && !knownFactAt(a, "wall_stone", x, y, tick) {
					return false
				}
				w := wallAt(s, x, y)
				return w != nil && w.HP < wallMaxHP(w.Kind) && invField(a.Inv, wallRepairMaterial(w.Kind)) >= 1
			}); ok {
				return &Intent{Goal: "repair", TargetX: stand.X, TargetY: stand.Y, ResX: res.X, ResY: res.Y}, "", nil
			}
			return nil, "", fmt.Errorf("%s has no damaged wall reachable to repair (with the right material)", a.Name)
		},
		"refuel_fire": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// T020: planner OR reflex (the one shared goal, FR-020). Target the
			// nearest KNOWN fire (spec 041), lit or cold — the completion
			// relights a cold one.
			if a.Inv.Wood < 1 {
				return nil, "", fmt.Errorf("%s lacks wood to refuel a fire", a.Name)
			}
			if !knowsAnyFresh(a, "fire", tick) {
				return nil, "", fmt.Errorf("you know of no fires")
			}
			if p, ok := nearestKnown(m, s, a, "fire", tick, nil); ok {
				return &Intent{Goal: "refuel_fire", TargetX: p.X, TargetY: p.Y}, "", nil
			}
			return nil, "", fmt.Errorf("no fire reachable to refuel")
		},
		"cook": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// T031: cook raw food at the nearest valid station, whichever is
			// nearer (the shared BFS's fixed neighbor order makes the tie-break
			// deterministic). Station-specific duration/output is resolved from
			// the target at the executor. Spec 041: the stations are the agent's
			// KNOWN ovens and the fires it REMEMBERS as lit — remembered Detail
			// (FuelUntil as last seen) against now, so burnout is predicted from
			// the agent's own knowledge, never read off the world.
			if a.Inv.FoodRaw <= 0 {
				return nil, "", fmt.Errorf("%s has no raw food to cook", a.Name)
			}
			station := func(x, y int) bool {
				if f, ok := knownFreshFact(a, "fire", x, y, tick); ok && f.Detail > tick {
					return true
				}
				return knownFactAt(a, "oven", x, y, tick)
			}
			if !knowsLitFire(a, tick) && !knowsAnyFresh(a, "oven", tick) {
				return nil, "", fmt.Errorf("you know of no lit fire or oven")
			}
			if p, ok := nearest(m, s, a.X, a.Y, station); ok {
				return &Intent{Goal: "cook", TargetX: p.X, TargetY: p.Y}, "", nil
			}
			return nil, "", fmt.Errorf("no lit fire or oven reachable to cook at")
		},
		"bathe": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// T032: water's only v1 consumer — bathe at a KNOWN oven (spec 041).
			r, _ := recipeFor("bathe")
			if !hasItems(a.Inv, r.Inputs) {
				return nil, "", fmt.Errorf("%s lacks water/wood to bathe", a.Name)
			}
			if !knowsAnyFresh(a, "oven", tick) {
				return nil, "", fmt.Errorf("you know of no ovens")
			}
			if p, ok := nearestKnown(m, s, a, "oven", tick, nil); ok {
				return &Intent{Goal: "bathe", TargetX: p.X, TargetY: p.Y}, "", nil
			}
			return nil, "", fmt.Errorf("no oven reachable to bathe at")
		},
		"drop": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// Planner/plan-only (FR-014). Instant on the agent's current tile; the
			// completion emits agent.dropped with the actual post-clamp counts. An
			// empty Kind or nothing carried resolves via intent_done at completion
			// (executor) — resolveGoal creates the intent regardless (the goal is
			// possible; the re-validation is where it may become a no-op).
			return &Intent{Goal: "drop", TargetX: a.X, TargetY: a.Y, Kind: kind, Qty: qty}, "", nil
		},
		"pick_up": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// Planner/plan-only. Target the nearest KNOWN pile tile (spec 041) —
			// piles sit on passable ground, so the agent walks onto it; the
			// completion re-validates a pile on/adjacent and moves goods
			// truncated to free bulk (Kind "" sweeps every kind in canonical
			// order). Pile presence stays a ground condition (a drained pile is
			// REMOVED from state while its fact lives until US3's correction —
			// the forage-overlay availability class).
			if !knowsAnyFresh(a, "pile", tick) {
				return nil, "", fmt.Errorf("you know of no piles")
			}
			if p, ok := nearestKnown(m, s, a, "pile", tick, func(x, y int) bool {
				return s.pileAt(x, y) != nil
			}); ok {
				return &Intent{Goal: "pick_up", TargetX: p.X, TargetY: p.Y, Kind: kind, Qty: qty}, "", nil
			}
			return nil, "", fmt.Errorf("no pile reachable")
		},
		"deposit": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// T024 (spec 013 US3): planner/plan-only. Target the nearest KNOWN
			// chest (spec 041; any owner — the commons/ownership split is social,
			// not mechanical). The completion re-validates the chest and
			// truncates to its free space; Kind "" or nothing that fits resolves
			// via intent_done at completion.
			if !knowsAnyFresh(a, "chest", tick) {
				return nil, "", fmt.Errorf("you know of no chests")
			}
			if p, ok := nearestKnown(m, s, a, "chest", tick, nil); ok {
				return &Intent{Goal: "deposit", TargetX: p.X, TargetY: p.Y, Kind: kind, Qty: qty}, "", nil
			}
			return nil, "", fmt.Errorf("no chest reachable")
		},
		"withdraw": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// T024: planner/plan-only. Target the nearest KNOWN chest (spec 041)
			// whose Store holds Kind (Kind "" ⇒ anything in it) — contents stay
			// a ground condition on top of the knowledge gate (what's inside is
			// not in the fact model; den-readiness class). The completion
			// truncates to the taker's free bulk and to what the chest holds; a
			// non-owner take co-emits the theft companion batch (US4, T029).
			if !knowsAnyFresh(a, "chest", tick) {
				return nil, "", fmt.Errorf("you know of no chests")
			}
			if p, ok := nearestKnown(m, s, a, "chest", tick, func(x, y int) bool {
				ch := s.chestAt(x, y)
				if ch == nil || ch.Store == nil {
					return false
				}
				if kind == "" {
					return bulk(*ch.Store) > 0
				}
				return carriedCount(*ch.Store, kind) > 0
			}); ok {
				return &Intent{Goal: "withdraw", TargetX: p.X, TargetY: p.Y, Kind: kind, Qty: qty}, "", nil
			}
			return nil, "", fmt.Errorf("no chest with those goods reachable")
		},
		"sleep": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			return &Intent{Goal: "sleep", TargetX: a.X, TargetY: a.Y}, "", nil
		},
		"goto_warmth": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// Spec 041: warmth the agent can PLAN on — near a fire it remembers
			// as lit, or on a shelter it knows (warmKnownPredicate). Terrain
			// passability stays ground truth (D3).
			warm, any := warmKnownPredicate(a, tick)
			if !any {
				return nil, "", fmt.Errorf("you know of no warm place")
			}
			if p, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return warm(x, y) && passable(m, s, x, y) }); ok {
				return &Intent{Goal: "goto_warmth", TargetX: p.X, TargetY: p.Y}, "", nil
			}
			return nil, "", fmt.Errorf("no warmth anywhere")
		},
		"search": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			// Spec 041 US4 (research D4): walk to the nearest frontier of the
			// KNOWN world — explored, passable, adjacent to unexplored land.
			// Arrival is an instant goal (wander-class); the perception sweep
			// then grows the map and the planner re-arms with fresh knowledge.
			// A fully-explored reachable world fails honestly (contracts §4).
			if p, ok := nearestFrontier(m, s, a); ok {
				return &Intent{Goal: "search", TargetX: p.X, TargetY: p.Y}, "", nil
			}
			return nil, "", fmt.Errorf("nothing left unexplored")
		},
		"wander": func(s *State, m *worldmap.Map, a *Agent, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
			r := rngAt(s.Seed, "wander", tick, idx)
			for try := 0; try < 8; try++ {
				dx, dy := int(r.Uint64N(9))-4, int(r.Uint64N(9))-4
				if (dx != 0 || dy != 0) && passable(m, s, a.X+dx, a.Y+dy) {
					return &Intent{Goal: "wander", TargetX: a.X + dx, TargetY: a.Y + dy}, "", nil
				}
			}
			return nil, "", fmt.Errorf("nowhere to wander")
		},
		"talk_to": talk,
		"seek":    talk,
	}
}

// resolveGoal turns a planner-chosen goal into a concrete, deterministic
// intent at the tick boundary (research R5). The model steers; the sim
// drives. Errors mean the goal is impossible right now — nothing is emitted.
// Dispatch is now table-driven (spec 014, R2); the per-verb semantics are
// unchanged, and an unknown goal errors exactly as the old switch default did.
func resolveGoal(s *State, m *worldmap.Map, idx int, goal string, targetAgent int, kind string, qty int, tick int64) (*Intent, string, error) {
	a := &s.Agents[idx]
	if r, ok := goalResolvers[goal]; ok {
		return r(s, m, a, idx, goal, targetAgent, kind, qty, tick)
	}
	return nil, "", fmt.Errorf("unknown goal %q", goal)
}

// chopIntent targets the nearest KNOWN standing tree (spec 041 — shared by
// the chop resolver and the reflex's firewood rung, so parity is by
// construction). The cleared overlay stays a ground condition (availability,
// not place knowledge — forage's harvested-overlay rationale; a chopped
// tree's fact would otherwise livelock the firewood rung until US3 corrects).
func chopIntent(s *State, m *worldmap.Map, a *Agent, tick int64) (*Intent, bool) {
	stand, res, ok := nearestKnownAdjacentTo(m, s, a, "tree", tick, func(x, y int) bool {
		return effectiveKind(m, s, x, y) == worldmap.Tree
	})
	if !ok {
		return nil, false
	}
	return &Intent{Goal: "chop", TargetX: stand.X, TargetY: stand.Y, ResX: res.X, ResY: res.Y}, true
}
