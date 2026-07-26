---
name: executor-world-state
description: The v3 storage economy (carried bulk cap, ground piles, drop/pick_up, builder-owned chests, food rot) and the terrain/structure overlays (cleared/harvested/quarried tiles, walls blocking movement, paths, fire/shelter/oven/chest/grave). Load for inventory-bulk, pile, chest, or terrain-passability questions.
kind: component
sources:
  - internal/sim/executor.go
  - internal/sim/terrain.go
  - internal/sim/agents.go
verified_against: cffd9a79bbed61ccac573d97c6cf544565b40336
---

# Executor — world state: storage economy and terrain

Child of [[executor]] — goods after leaving an agent's hands (bulk cap,
ground piles, chests, rot) and the map's event-sourced overlays
(cleared/harvested/quarried tiles, structures, walls, paths).

## How it works

**Carried bulk & the v1 storage economy** (spec 013): every carried good
counts toward a per-villager `bulk` — one unit per inventory count, one per
carried spear or (since spec 032) axe — capped at `bulkCap` (24), derived via
`bulk()`/`freeBulk()`, never stored. Every gather completion (`forage`/`chop`/
`hunt`/`quarry`/`collect_water`) clamps its yield to the taker's pre-event free
bulk and is skipped — no event, so no depletion — at zero free bulk
(US1-AS1/AS2); a hand-craft completion re-validates its net
output-minus-input bulk delta the same way (only `craft_planks` is net-positive;
crafts don't truncate — they don't happen if the net won't fit). The
give-to-starving social rule (`repayable`/`giveable`) likewise requires
receiver free bulk before offering a give.

Ground goods live as `State.Piles`, one per tile (event-sourced overlay,
like `Quarried`). `drop`/`pick_up` are instant-on-arrival, planner/plan-only
goals (never in the reflex ladder, FR-014): `drop` moves a named `Kind`/`Qty`
(`Qty` 0 = all carried) from inventory onto the agent's tile, creating or
merging its pile; `pick_up` targets the nearest pile (on or adjacent) and
moves goods in, truncated to free bulk, emitting one `agent.picked_up` per kind
moved — `Kind` "" sweeps every kind in canonical field order (wood,
stone, water, planks, refined_stone, food_raw, food_cooked, meals, spears, and
since spec 032, axes). Ground
food is batch-tracked (`FoodBatch{Kind, N, SpoilAt}`, drop order, same
`(Kind, SpoilAt)` merges); every non-food kind is a flat count; spears and axes
carry their remaining uses, always sorted ascending — most-worn moves first
on either side of a transfer (axes clone the spear path exactly — build,
drop/pick-up, deposit/withdraw, death-spill).
`agent.died` also spills the dead agent's
entire inventory onto a death-tile pile (reducer-internal, no new event —
research R7's debt-opening precedent), and `buildSite` (`terrain.go`) rejects any
tile holding a pile (FR-007 — goods aren't buried).

**Builder-owned chests** (`build_chest`, spec 013 US3): a fifth structure kind
alongside fire/shelter/oven, gated on `chestPlankCost` (6) planks with a
fire-comparable build duration. The builder is the chest's `Owner`
permanently (no transfer or inheritance in v1); the chest gets an empty
`Store`, capped at `chestCap` (48, the same derived `bulk()`). `deposit`/
`withdraw` are instant-on-arrival, planner/plan-only goals resolving to the
nearest chest (`withdraw` with a named `Kind` targets the nearest one
holding it); their completions re-validate the chest still stands and truncate
the move to the tighter side — chest free space on deposit, taker free bulk on
withdraw. A non-owner `withdraw` is theft: never blocked,
always marked — the executor co-emits a same-tick companion batch
(`social.chest_taken`, a reason-`"theft"` `social.relation_changed`, the owner's
gossip-seed memory, witness memories for nearby villagers; full mechanics in
[[social-fabric]]).

**Food rot** (spec 013 US5): on the needs heartbeat's per-game-minute boundary,
`stepEvents` also sweeps every pile for food batches whose
`SpoilAt` has arrived, emitting one `sim.food_rotted` per (pile, kind) with
same-kind batches merged — a pure function of (state, tick), the
fuel-sweep pattern. Chest food carries no batches and never rots (FR-010).

**Terrain overlays** (`terrain.go`): chopped trees and harvested forage are
event-sourced state over the static map — `effectiveKind`/`passable` merge
[[worldmap-generation]] with `State.Cleared`/`Harvested`/`Quarried`; forage regrows
after 12 game-hours (`sim.forage_regrown`), dens cool 6 game-hours after a hunt.
A quarried rock outcrop (spec 012) differs from the other two: it does NOT
revert to Grass — `effectiveKind` renders it `worldmap.Depleted` permanently (no
regrow in v1), `passable` admits it, but it is neither buildable
(`buildSite`) nor quarryable again. Spec 068's marsh and sand ground covers
deliberately have NO overlay arm here: nothing
clears, harvests, or quarries them, so `effectiveKind` always renders what the
generator produced, and `passable` admits both alongside grass/forage/Depleted
(no resource affordance, never buildable — same as forage).
Structures (`fire`, `shelter`, `oven`, `chest`)
exist only in state; `warmAt` is a *lit* fire within Manhattan radius 2, or standing on a
shelter (ovens grant no warmth) — decomposed (spec 074-look-cursor, TASK-142)
into a private `warmthSource(s, x, y, tick) (warm bool, source string)` core
attributing WHICH structure (fire vs. shelter, in `warmAt`'s original scan
order) provided the warmth; `warmAt` is now a thin wrapper;
`internal/sim/env.go`'s exported `EnvAt` reads the same core so the TUI's
look-cursor TILE view shows warmth level/source without a duplicated
radius constant (sibling decomposition for light: `gru.md`'s `lightSource`). Behavior and byte output unchanged — a pure read-side export, no
reducer/tuning change. `fireStructAt`/`litFireAt` locate a fire by
coordinate and test lit-ness for refuel/cook completion checks. Spec 032
adds two more structure kinds — the first to affect pathing: `isWall`
names the wall family (`wall_plank`, `wall_stone`); `passable` checks
`wallAt(s, x, y) != nil` FIRST, refusing the tile if so — a standing wall is
the one structure family blocking movement (`buildSite`'s generic
"no structure on this tile" scan already kept walls, like every
structure, un-buildable-over). `wallMaxHP(kind)` derives each kind's ceiling
(`wallPlankHP` 200, `wallStoneHP` 600) for the build stamp, repair clamp,
and TUI damage styling — never stored separately (`WallMaxHP` exports it
for [[tui-client]]). `agentAt(s, x, y)` backs the wall-build occupancy guard
(FR-007: a wall may never land on a living agent's tile) — since spec
038 checked only at completion (deferring, then bounded-loud-
failing on a lingering occupant), not during mid-work re-validation.
`pathAt(s, x,
y)` reports a `path` structure underfoot — the movement dual-phase cadence's
per-step predicate; a path has no `HP` and never blocks (`isWall` is
false for it). Spec 044 (US4) adds an eighth structure kind, `grave` — placed
by the reducer at a death tile, never built by any goal (no recipe, no build
verb); never blocks movement but, like every structure, blocks `buildSite`
on its tile; the perception sweep witnesses it like any structure
([[sim-state-reducer]], [[mental-maps]]).

## Connections

Parent note: [[executor]]. [[worldmap-generation]] supplies the static
terrain (Rock quarry sites, marsh/sand ground covers) these overlays
merge with; [[social-fabric]] carries the theft companion batch a non-owner
chest `withdraw` triggers; [[tui-client]] renders wall HP dimming, path
tiles, fire lit/cold state, ground piles, and chest contents;
[[world-migration]] re-places carried souls and over-cap carry on a fresh map
across format-version cuts; [[executor-goal-completions]] defines a build/
demolish/repair/craft goal's completion event.

## Operational notes

The v3 storage economy (spec 013) has its own suite —
`bulk_cap_test.go`, `ground_pile_test.go`, `chest_test.go`, `theft_test.go`,
`rot_test.go`, `migrate_test.go`. Spec 032 (walls, axes, paths) adds `wall_test.go` (blocking/rerouting,
occupancy guard, HP stamping, multi-cycle demolish/repair math and re-arm,
replay determinism), `axe_test.go` (bare-vs-axe yield, ten-use breakage,
bulk truncation, storage round-trip, replay), and `path_speed_test.go` (a
paved corridor halves traversal ticks vs. unpaved); both specs also extend the
`whole_feature_test.go` pass — spec 032's exercising all three together.
