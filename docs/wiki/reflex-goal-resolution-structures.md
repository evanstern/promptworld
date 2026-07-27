---
name: reflex-goal-resolution-structures
description: Child of [[reflex-policy]] — resolveGoal's storage and structural goals (build_chest, drop/pick_up/deposit/withdraw, craft_axe, build_wall_plank/build_wall_stone, demolish, repair, build_path), each naming its resolver, gate, and spec-013/032/041 knowledge-gating. See [[reflex-goal-resolution]] for consumable/survival/exploration goals.
kind: component
sources:
  - internal/sim/policy.go
  - internal/sim/path.go
verified_against: fc9566d527941d3950fdd307168556820bd0875b
---

# Reflex goal resolution — storage, walls, and paths

Child of [[reflex-policy]]: `resolveGoal`'s storage and structural goals —
spec 013's chest/pile goals and spec 032's wall/axe/path goals. See
[[reflex-goal-resolution]] for the consumable, survival, and exploration
goals sharing the same `resolveGoal` catalog, and [[reflex-pathfinding]] for
the `nearestAdjacentTo` geometry every adjacency-gated goal below uses.

## resolveGoal's goal vocabulary (storage, walls, and paths)

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
  `craft_planks`/`craft_stone`/`craft_spear` (see [[reflex-goal-resolution]])
  — no travel, resolves once `recipeFor("craft_axe")`'s inputs (1 plank + 1
  stone) are satisfied.
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

## Connections

Parent [[reflex-policy]] summarizes this catalog and links every sibling
child; [[reflex-goal-resolution]] covers the remaining goal vocabulary;
[[reflex-pathfinding]] is the `nearestAdjacentTo` geometry every
adjacency-gated goal above uses; [[mental-maps]] is the knowledge store
gating every KNOWN-prefixed lookup; [[executor]] applies the completions,
including the storage-goal truncation/theft rules.
