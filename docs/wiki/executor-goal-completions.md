---
name: executor-goal-completions
description: What each v2/spec-032 goal's completion emits — gather yields, hand-craft outputs, build/demolish/repair on walls and other structures, cook/bathe/refuel — one bullet per goal's event and reducer effect. Load when tracing a specific goal's completion payload or re-arm behavior.
kind: component
sources:
  - internal/sim/executor.go
  - internal/sim/recipes.go
  - internal/sim/terrain.go
verified_against: 376afd4cee54839a545bc88409f3c485c2f5149d
---

# Executor — goal completions

Child of [[executor]] — the completion-time behavior of every goal in
[[executor-goals-and-intents]]'s catalog: the event each emits and the state
change the reducer applies.

## How it works

Completion behavior per goal (since spec 084, `heed_directive` — the
DIRECTIVE rung's walk-to-site leg, [[guardian-designations]] — completes
instantly on arrival via bare `agent.intent_done`, the `search`/`seek`
class):
- `chop` → `agent.chopped` (+`chopYieldBare` (1) Wood bare-handed, or
  `chopYieldAxe` (3) with a carried axe — spec 032 US2, replacing the old flat
  `chopWood` (2)). `quarry` → `agent.quarried` (+`quarryYieldBare`/`quarryYieldAxe`,
  same 1/3 split), and the outcrop is added to `State.Quarried` (below).
  `collect_water` → `agent.collected_water` (+`collectWaterYield` Water); water
  sources never deplete. A carried axe (`Axes[0]`, checked pre-mutation like the
  spear/hunt precedent) spends its last use on either harvest — spending it to
  zero emits a companion `agent.axe_broke` right after, in the same batch, plus
  a memory ("My axe broke at the work…"), the exact `agent.spear_broke` pattern.
  Since spec 081, `chop`/`quarry` completion also mints the ACTOR a first-person
  act memory in the same batch ("Felled the tree at (x,y)." / "Quarried the
  outcrop at (x,y).", `salChop`/`salQuarry`, the hunt-memory shape) — superseding
  the old "completed chops mint no memory" posture (operator decision
  2026-07-26) — and the reducer arm removes the felled/quarried place-fact from
  the actor's and every awake in-radius witness's mental map at the act event
  ([[mental-map-perception]]).
- `hunt` → `agent.hunted`; a carried spear (`Spears[0]`, checked pre-mutation) raises
  the yield to `huntYieldSpear` (vs. `huntYieldBare` bare-handed) and spends that
  spear's last use — spending it to zero emits a companion `agent.spear_broke` right
  after, in the same batch, plus a memory ("My spear broke on the hunt…").
- `craft_planks`/`craft_stone`/`craft_spear`/`craft_axe` (spec 032 US2 adds the
  last: 1 plank + 1 stone → one axe holding `axeDurability` (10) uses) → inputs
  re-validated against `recipes.go`'s table at completion (`hasItems`), and the
  net bulk delta re-validated against `freeBulk` (US1 T012); either failing
  now resolves LOUDLY (`agent.intent_failed`, spec 096, `"contested"`), no
  craft or inventory change. A
  satisfied craft emits `agent.crafted{Kind}`; the reducer applies the recipe's
  delta, appending to `Inv.Spears`/`Inv.Axes` (sorted ascending) for the two
  durability-slice crafts, or a flat `Inv` field otherwise.
- `build_oven` → `agent.built{Kind: "oven"}`; the first oven in the village gets
  distinct memory text ("Raised the village's first oven — meals and baths, at
  last."), and nearby living agents get a witness memory, same pattern as a
  witnessed death.
- `build_wall_plank`/`build_wall_stone` (spec 032 US1; occupancy tolerance spec
  038 US3) → `agent.built{Kind: "wall_plank"|"wall_stone"}` landing on the
  intent's `Res` tile — walls are the one build family that lands ADJACENT to
  the builder (`Target`) rather than on it, so a builder can never wall itself
  in (FR-007). Mid-work, only the site is re-validated (`buildSite(Res)`) —
  spec 038 dropped the `!agentAt` term from the mid-work guard, so a passerby
  crossing the reserved tile no longer cancels the build. At the completion
  moment `agentAt(Res)` IS checked: an occupied tile DEFERS completion (no
  event that tick, never entomb) rather than cancelling, and the deferral is
  bounded — once `nextTick - WorkStart >= workDuration + wallOccupancyGraceTicks`
  (120) the build fails LOUDLY via `buildFailedEvents` with `reason:
  buildFailSiteBlocked` instead of waiting forever. A vanished site (mid-work
  or at the due tick) fails loudly immediately with `buildFailSiteUnbuildable`
  — site loss is never waited out. On a clean completion the reducer stamps
  the new wall's `HP` at `wallMaxHP(kind)` (`wallPlankHP` 200, `wallStoneHP`
  600 — derived from kind, never stored separately, the fire-lit-ness
  doctrine) and spends the recipe's planks/refined-stone. A builder memory
  rides at the shelter salience tier ("Built a wall."); walls emit no witness
  memory.
- `demolish` (spec 032 US1) → one chip per completed work cycle against the
  still-standing wall at `Res` (re-validated; a vanished wall — someone else
  finished it first — resolves LOUDLY via `agent.intent_failed`, spec 096,
  `"target gone"`): `agent.wall_chipped`
  removes `demolishChipHP` (100), and the reducer re-arms the SAME intent
  (`WorkStart` reset to 0) for another cycle when the wall would still stand
  (`HP - chip >= 1`); the cycle that would take it to zero instead emits
  `agent.wall_destroyed`, which removes the structure and clears the intent. A
  plank wall takes 2 cycles, a stone wall 6. No memory (spam-avoidance, the
  forage/path precedent — chop/quarry left that precedent in spec 081).
- `repair` (spec 032 US1) → one cycle mends a still-damaged wall at `Res` with 1
  unit of its matching carried material (`wallRepairMaterial`: planks for
  `wall_plank`, refined stone for `wall_stone`), restoring `repairHPPerUnit`
  (100) HP clamped to `wallMaxHP`; `agent.wall_repaired` re-arms the same intent
  for another cycle if the wall is still damaged AND material remains, else
  clears it. Re-validated at completion — wall gone, mended, or material spent
  resolves LOUDLY via `agent.intent_failed` (spec 096, `"target gone"`).
- `build_path` (spec 032 US3) → `agent.built{Kind: "path"}` landing on the
  intent's own `Target` tile (stand-on-target, like fire/oven/chest), spending
  `pathStoneCost` (1) raw stone; a path carries no `HP` (`isWall` is false for
  it) and emits no builder memory (not formative, same spam-avoidance
  precedent as forage). Standing on a path tile is what grants the 2x
  movement bonus described above (`pathAt`), not the build itself.
- `cook` → up to `ovenBatchSize` FoodRaw converts to `agent.cooked`: at a fire,
  fuel-free, producing `food_cooked`; at an oven, additionally burning 1 carried
  wood, producing `meals` (mirrors the fire's own fuel). A cold/gone station
  fails LOUDLY mid-work (`agent.intent_failed`, spec 096, `"target gone"`); an
  oven with no wood or no raw food at completion fails LOUDLY too, distinctly
  (`"contested"`).
- `bathe` (oven only) → re-validates carried water + wood at completion (water's
  only consumer); missing either fails LOUDLY (`agent.intent_failed`, spec 096,
  `"contested"`). A satisfied bathe emits `agent.bathed` with absolute post-cap
  Morale/Warmth (`bathMorale`/`bathWarmth` bumps, gru-pattern) and a
  positive-toned memory.
- `refuel_fire` → re-validated on arrival (fire present, wood carried); a refuel
  granting no gain over the current deadline is a no-op (`agent.intent_done`
  only — not in spec 096's list).

## Connections

Parent note: [[executor]]. [[executor-goals-and-intents]] is this note's
companion — the state machine and goal catalog these completions belong to.
[[sim-state-reducer]] applies every event this section lists;
[[social-fabric]] documents the theft companion batch a non-owner
`withdraw`-adjacent completion can trigger (see [[executor-world-state]]);
[[tui-client]] renders wall HP and fuel/lit state these completions produce.

## Operational notes

Spec 038 (loud build failure & occupancy tolerance) rewrites `wall_test.go`'s
occupancy-guard coverage into a defer-then-fail matrix — `TestWallOccupancyGuard`
(a permanent squatter fails loudly at the grace bound), `TestWallBuildToleratesPasserby`
(mid-work crossing no longer cancels), and `TestWallBuildDefersThenCompletes` (a
mid-window departure lets completion land on the first clear tick, never
during occupancy) — plus `TestWallBuildSiteVanishedFailsLoud` for the
site-loss path, and an extended `whole_feature_test.go` pass proving
`agent.build_failed` and its paired failure memory replay byte-identically.
