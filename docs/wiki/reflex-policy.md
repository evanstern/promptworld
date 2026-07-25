---
name: reflex-policy
description: Deterministic survival decisions for idle agents (eat → food/search → night warmth/fire-refuel → rest → prep → wander) plus BFS pathfinding with fixed neighbor order; resolveGoal resolves the full spec-012 planner goal vocabulary (quarry, water, crafting, stations, refuel, cook, bathe) plus spec-013's storage goals, spec-032's walls/axes/paths goals, and spec-041's knowledge-gated resolution (nearestKnown, search/frontier, last-known-sighting talk_to) to coordinates
kind: component
sources:
  - internal/sim/policy.go
  - internal/sim/path.go
verified_against: cc514f7ff456fefbcfe289471c5a1467b8e724df
---

# Reflex policy & pathfinding

`decideIntent` is the deterministic pure function that gives idle, awake agents
something to do — since TASK-7, only agents idle past the 120-tick grace (the
planner's injection window). It is the permanent degraded mode: when no planner
thoughts arrive, this keeps bodies alive. `resolveGoal` (same file) is the shared
target resolver: planner goals from [[agent-mind]] resolve to coordinates through
the exact same nearest-X helpers the reflex uses. Spec 012 widened `resolveGoal`'s
goal set considerably (quarrying, water, crafting, an oven, refueling, cooking,
bathing) while trimming the reflex ladder itself down to one addition — refueling
a dying fire — and one removal — shelter-building dropped out of the reflex
entirely once it was re-costed in planks. Spec 013 (inventory & storage v1) widened
`resolveGoal` again — a chest to build, goods to drop/pick up/deposit/withdraw —
and left the reflex ladder itself completely untouched: all five new goals are
planner/plan-only (FR-014), added to `resolveGoal` but never reachable from
`decideIntent`. Spec 032 (walls, axes, paths) widened it once more — two wall
builds, a demolish/repair pair on an existing wall, a fourth hand-craft
(`craft_axe`), and a path build — every one planner/plan-only, same pattern:
the reflex ladder itself gains nothing from spec 032 (an axe or a wall is
never something `decideIntent` reaches for on its own). Spec 014 (TASK-53) restructured `resolveGoal` from one large
switch into `goalResolvers`, a name-keyed resolver table with the old per-verb
bodies verbatim — the [[tool-registry]]'s boot-time coverage gate
(`sim.ValidateToolCoverage`) asserts every World tool on the villager roster has
a table entry, so a registered verb can never lack its resolver. The plan-step
accept set that once lived beside it (`planGoals`) is gone: the sim door now
derives it from the registry ([[sim-loop]]).

Spec 041 ([[mental-maps]]) changed WHAT every resolver above searches, not
the search mechanics: every goal that targets a place now resolves against
the acting agent's own FRESH facts (`nearestKnown`/`nearestKnownAdjacentTo`,
`path.go` — knowledge-gated twins of `nearest`/`nearestAdjacentTo` that keep
the identical BFS geometry and tie-breaking, only the match closure differs),
not the world's ground truth. A resolver that knows of nothing of the right
kind fails with an epistemic reason ("you know of no forage") rather than
resolving to a place it has never perceived. Availability that is not itself
place knowledge — a harvested forage spot, a cooling den, wall damage, chest
contents, quarry depletion — stays layered on top as an ordinary ground
condition: the agent knows the PLACE and walks there; whether the walk pays
off is checked at arrival, exactly as before. A new goal, `search`
(US4), resolves to the nearest exploration frontier
(`nearestFrontier`) instead of any resource — the deliberate answer to
knowing of nothing. `talk_to`/`seek` resolves to the target's last KNOWN
sighting (the mental map's peer record), never its live coordinates.

## How it works

Priority ladder (first match wins):

1. **Eat** — hungry (`Food < hungryAt`, 350) and carrying any edible unit
   (`hasAnyFood`: `Inv.Meals + Inv.FoodCooked + Inv.FoodRaw > 0`) → instant
   `agent.ate`. The triplet check replaces the old raw-food-only check (T018) so an
   agent carrying only meals or only cooked food still eats reflexively.
2. **Get food** — hungry, nothing carried → `foodIntent`: nearest KNOWN
   fresh forage fact (spec 041 — availability still checked as a ground
   condition), else nearest KNOWN ready den (`hunt`); knowing of neither
   (US4, T026, FR-013 parity without omniscience) falls back to the nearest
   exploration frontier (`search`) — hunger-only, so a fed villager never
   mounts an expedition just to top up the larder.
3. **Night, cold** (`!warmAt`) — reach KNOWN warmth (`goto_warmth`,
   `warmKnownPredicate`: a remembered-lit fire or a known shelter) if
   reachable; else `reflexRefuelIntent`, the reflex's one new rule (T020,
   FR-012): if carrying any wood and a KNOWN fire is remembered cold or has
   less than `refuelDyingBelow` (3600 ticks, one game-hour) of remembered
   fuel left, relight or top it up (`refuel_fire`) — cheaper than a fresh
   build; else build a fire if carrying `fireWoodCost` (2) wood; else chop
   the nearest KNOWN standing tree (yes, chopping in the cold dark — that's
   the day-1 drama working as designed).
4. **Night, warm** — sleep where you stand.
5. **Exhausted by day** (`Rest < tiredAt`, 250) — nap, preferring a warm tile.
6. **Prep** — knows of no fire (spec 041: `!knowsAnyFresh(a, "fire", tick)`,
   the agent's own belief rather than `!hasStructure("fire")` — a village fire
   someone else built and this agent has never seen still triggers this rung
   for THIS agent) → build/chop toward one; then `reflexRefuelIntent` again,
   unconditionally, to keep a known fire from dying down; then stock the
   larder to `stockFoodRawTo` (8) units of raw food (`Inv.FoodRaw`).
   Shelter-building is gone from this ladder (T020): since spec 012 re-costed
   it in `Planks` (`shelterPlankCost`, 8) instead of raw wood, it joined the
   crafting economy and became planner-only — the reflex never enters
   `resolveGoal`'s `build_shelter`, `craft_*`, or `build_oven` cases.
7. **Wander** — a seeded short stroll (`rngAt(seed, "wander", tick, idx)`).

Waking (`wakeReason` in executor.go) mirrors this: day + decent rest
(`Rest >= 600`), or a hunger emergency the agent can act on — `Food < 150` and
`hasAnyFood`, the same triplet check as the eat rule above. Fully-rested agents
sleep through the night — the live-run sleep/wake churn bug is documented in the
TASK-5 notes. Actually eating the food (most-nutritious form first — `Meals` then
`FoodCooked` then `FoodRaw`, stopping at `satietyAt`) is the executor's
`eatOutcome`, detailed in [[executor]].

## resolveGoal's goal vocabulary

`resolveGoal` grew from the original handful (`eat`, `forage`, `hunt`, `chop`,
`build_fire`, `build_shelter`, `sleep`, `goto_warmth`, `wander`, `talk_to`/`seek`)
to cover spec 012's full economy, still resolving every goal to a concrete
`Intent` or an error through the same `nearest`/`nearestAdjacentTo` helpers the
reflex uses:

- **`eat`** now refuses on two grounds — nothing to eat (`!hasAnyFood`) or already
  sated (`Needs.Food >= satietyAt`, 900) — so a planner-chosen eat is never wasted
  at the ceiling.
- **`quarry`** and **`collect_water`** are planner-only (never in the reflex
  ladder): both resolve via `nearestKnownAdjacentTo` (spec 041), the
  same beside-the-resource pattern `chop` uses, matching a KNOWN fresh
  `"rock"`/`"water_edge"` fact instead of ground-truth `worldmap.Rock`/
  `worldmap.Water` — knowing of neither fails honestly ("you know of no rock
  outcrops"/"you know of no water") before the search even runs; quarry
  depletion stays a ground condition layered on top (an outcrop's fact
  persists until US3's correction, the forage-overlay precedent).
- **`build_fire`** is unchanged: gated on `fireWoodCost` wood, resolved to the
  nearest `buildSite`.
- **`build_shelter`** is re-costed to `Planks` (`shelterPlankCost`, 8, was wood)
  and is planner-only now that the reflex dropped it.
- **`build_oven`** is new: gated on `recipeFor("build_oven")`'s inputs (refined
  stone plus planks, checked via `hasItems`) and resolved to a `buildSite` the
  same way as fire and shelter.
- **`craft_planks`**, **`craft_stone`**, and **`craft_spear`** are new hand-crafts
  that need no travel — each resolves to the agent's own tile once
  `recipeFor(goal)`'s inputs are satisfied.
- **`refuel_fire`** is the one goal both the reflex (`reflexRefuelIntent`) and the
  planner can choose (FR-020): it targets the nearest KNOWN fire (spec 041,
  `nearestKnown`) regardless of remembered lit state — a cold fire is relit
  on arrival, a dying one topped up. See
  [[executor]] for the fuel window (`fireBurnPerWood`, `fireFuelCap`) the
  completion applies.
- **`cook`** targets the nearest station the agent KNOWS is valid (spec
  041): an oven fact, or a fire fact remembered lit — its remembered `Detail`
  (`FuelUntil` as last seen) still ahead of now, predicting burnout from the
  agent's own knowledge rather than reading the world's live fuel state; the
  fixed BFS neighbor order still makes the tie-break deterministic, and the
  station reached determines the output and duration (`food_cooked` vs.
  `meals`) at the executor.
- **`bathe`** is new and oven-only, gated on `recipeFor("bathe")`'s water/wood
  inputs — water's only v1 consumer; since spec 041 the oven itself must
  also be KNOWN (`knowsAnyFresh`/`nearestKnown`).
- **`build_chest`** (spec 013 US3) is planner/plan-only, gated on
  `chestPlankCost` (6) planks and resolved to the nearest `buildSite` — the same
  pattern as `build_fire`/`build_oven` (the pile-tile exclusion, FR-007, already
  lives in `buildSite`).
- **`drop`**, **`pick_up`**, **`deposit`**, and **`withdraw`** (spec 013 US2/US3)
  are the storage goals, all planner/plan-only and instant-on-arrival (like
  `eat`): `drop` targets the agent's own tile (no place knowledge needed);
  `pick_up`, `deposit`, and `withdraw` (spec 041) target the nearest KNOWN
  pile/chest — `pick_up` a fresh `"pile"` fact, `deposit` any KNOWN chest
  (still no ownership gate), `withdraw` a KNOWN chest whose `Store` holds
  `Kind` (or, with `Kind` "", any KNOWN chest holding anything) — pile
  presence and chest contents stay ground conditions on top of the knowledge
  gate (what's inside a chest, or whether a pile has drained, is not itself
  place knowledge). All four carry `Kind`/`Qty` (`Qty` 0 = all of kind, or as
  much as fits) onto the resolved `Intent`, threaded through to the
  completion at [[executor]] — see there for the truncation/re-validation
  rules and the theft consequences of a non-owner `withdraw`.
- **`craft_axe`** (spec 032 US2) shares the same hand-craft closure as
  `craft_planks`/`craft_stone`/`craft_spear` — no travel, resolves once
  `recipeFor("craft_axe")`'s inputs (1 plank + 1 stone) are satisfied.
- **`build_wall_plank`** and **`build_wall_stone`** (spec 032 US1) share a
  `wallBuild` closure, gated on `recipeFor(goal)`'s inputs, that resolves via
  `nearestAdjacentTo` over `buildSite` — unlike every other build (which
  resolves the agent's own standing tile as the target), a wall build stands
  the agent on the neighboring passable tile (`Target`) and puts the wall on
  the adjacent buildable one (`Res`), the same stand/build split `chop`/`quarry`
  use beside a resource: building where you stand would entomb the builder the
  instant the wall lands (FR-007).
- **`demolish`** (spec 032 US1) resolves via `nearestAdjacentTo` over a KNOWN
  wall (spec 041: a fresh `"wall_plank"` or `"wall_stone"` fact, either kind
  — "you know of no walls" when neither) — adjacent-stand like the wall
  builds, since a wall tile is impassable. No material is required to tear
  one down; damage itself stays ground truth, checked at arrival.
- **`repair`** (spec 032 US1) resolves via `nearestAdjacentTo` over a KNOWN
  wall that is ALSO damaged and affordable (`w.HP < wallMaxHP(w.Kind)` and
  `invField(a.Inv, wallRepairMaterial(w.Kind)) >= 1`, both still ground
  conditions — damage is not in the fact model, so a wall mended behind the
  agent's back simply no-ops at arrival) — a wall already at full health
  never resolves; there is nothing to repair.
- **`build_path`** (spec 032 US3) is stand-on-target like `build_fire`
  (resolves via plain `nearest` over `buildSite`, not adjacency), gated on
  `pathStoneCost` (1) stone — unaffected by spec 041 (a buildable site is
  never itself a place-fact).
- **`search`** (spec 041 US4) is new: resolves to the nearest exploration
  frontier (`nearestFrontier` — an explored, passable tile adjacent to
  unexplored land, Yamauchi-style) regardless of resource kind; a fully
  explored reachable world fails honestly ("nothing left unexplored").
  Wander-class completion (below, [[executor]]) — the walk itself does the
  exploring.
- `sleep` and `wander` are unchanged. `goto_warmth` (spec 041) now resolves
  against `warmKnownPredicate` — a remembered-lit fire or a KNOWN shelter,
  never a live warmth read — failing honestly ("you know of no warm place")
  rather than falling through to build/chop when nothing known is warm.
  `talk_to`/`seek` (spec 041, T013) resolves to the target's LAST KNOWN
  sighting (`peerSightingOf`, the mental map's peer record) rather than the
  target's live coordinates — a stale sighting walks honestly to where the
  target was last seen, and the landing/arrival guards
  (`GuardTargetPresent`) cover a miss; liveness (`Dead`) stays a live check,
  since death-knowledge honesty is beyond this feature's place-fact scope.

Pathfinding (`path.go`, unchanged in its own geometry by spec 012, spec 032's
path-the-tile-improvement feature (a naming coincidence), or spec 041):
breadth-first search with **fixed neighbor order (N, E, S, W)** and FIFO
frontier, so shortest paths and nearest-match searches are identical on every
run. `nextStep` re-derives one hop per move from the shortest path (paths are
never stored in state — movement outcomes are evented, so replay needs no
path data); a standing wall (spec 032) makes its tile impassable via
[[executor]]'s `passable`, so BFS routes around walls with no change to
`path.go` itself — walls are just another obstacle the same search already
handles. `nearest` finds the closest reachable tile matching a predicate in
BFS order; `nearestAdjacentTo` finds a standing tile beside a resource —
chopping a tree, quarrying rock, drawing water, and (spec 032)
building/demolishing/repairing a wall all resolve through it. Spec 041 adds
three knowledge-gated wrappers on the SAME geometry, never touching the BFS
itself: `nearestKnown`/`nearestKnownAdjacentTo` layer a fresh-fact check onto
`nearest`/`nearestAdjacentTo`'s match closure (so knowledge-gated resolution
keeps every ground-truth search's exact tie-breaking), and `nearestFrontier`
(US4, Yamauchi-style) finds the closest reachable tile the agent's map marks
EXPLORED that 4-neighbors at least one UNEXPLORED in-bounds tile, decoding
the agent's `Explored` bitmap once per search — [[mental-maps]] owns the
bitmap and fact-freshness semantics these wrappers read. The escape clause
lets an agent standing on impassable terrain (pre-terrain saves) step out.

## Connections

[[executor]] invokes decisions on a staggered cadence and executes the resulting
intents, including the fire-fuel and cooking/crafting mechanics several of the new
goals above key on; passability comes from [[executor]]'s terrain overlays over
[[worldmap-generation]]; randomness only via [[deterministic-rng]] purpose tags
(`wander`, plus `genesis` placement in [[sim-state-reducer]]); [[mental-maps]]
is the per-agent knowledge store every goal resolver now reads through —
`nearestKnown`/`nearestKnownAdjacentTo`/`nearestFrontier`, `knowsAnyFresh`,
`warmKnownPredicate`, and `peerSightingOf` all live in this note's files but
are gated entirely on facts [[mental-maps]] and the executor's perception
sweep populate.

## Operational notes

BFS over a 64×64 map per decision/move is the current throughput ceiling — the
executor still clears >200k ticks/sec in the test harness, and auto-slow
([[sim-loop]]) degrades honestly under load. TASK-7 replaces this ladder with
planner-chosen goals; the ladder itself must remain reachable as the fallback.
