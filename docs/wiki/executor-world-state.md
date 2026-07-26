---
name: executor-world-state
description: The v3 storage economy (carried bulk cap, ground piles, drop/pick_up, builder-owned chests, food rot) and the terrain/structure overlays (cleared/harvested/quarried tiles, walls blocking movement, paths, fire/shelter/oven/chest/grave). Load for inventory-bulk, pile, chest, or terrain-passability questions.
kind: component
sources:
  - internal/sim/executor.go
  - internal/sim/terrain.go
  - internal/sim/agents.go
verified_against: b6a20eaa4da1073a69959a5aff69591d931103a9
---

# Executor — world state: storage economy and terrain

Child of [[executor]] — goods once they leave an agent's hands (bulk cap,
ground piles, chests, rot) and the map's event-sourced overlay state
(cleared/harvested/quarried tiles, structures, walls, paths).

## How it works

**Carried bulk & the v1 storage economy** (spec 013): every kind of carried good
counts toward a per-villager `bulk` — one unit per inventory count, one per
carried spear or (since spec 032) axe — capped at `bulkCap` (24), derived via `bulk()`/`freeBulk()` and
never stored. Every gather completion (`forage`/`chop`/`hunt`/`quarry`/
`collect_water`) clamps its yield to the taker's pre-event free bulk and is
skipped entirely — no event at all, so no depletion — when free bulk is already
zero (US1-AS1/AS2); a hand-craft's completion additionally re-validates its net
output-minus-input bulk delta the same way (only `craft_planks` is net-positive;
crafts don't truncate, they simply don't happen if the net won't fit). The
give-to-starving social rule (`repayable`/`giveable`) likewise requires the
receiver have free bulk before a give is offered.

Ground goods live as `State.Piles`, one per tile (event-sourced overlay state,
like `Quarried`). `drop`/`pick_up` are instant-on-arrival, planner/plan-only
goals (never in the reflex ladder, FR-014): `drop` moves a named `Kind`/`Qty`
(`Qty` 0 = all carried) from inventory onto the agent's own tile, creating or
merging the tile's pile; `pick_up` targets the nearest pile (on or adjacent) and
moves goods in, truncated to free bulk, emitting one `agent.picked_up` per kind
actually moved — `Kind` "" sweeps every kind in canonical field order (wood,
stone, water, planks, refined_stone, food_raw, food_cooked, meals, spears, and
since spec 032, axes). Food
on the ground is batch-tracked (`FoodBatch{Kind, N, SpoilAt}`, drop order, same
`(Kind, SpoilAt)` merges); every non-food kind is a flat count; spears and axes
carry their remaining uses, always sorted ascending so the most-worn moves first
on either side of a transfer (axes move the exact same way as spears — build,
drop/pick-up, deposit/withdraw, and death-spill all clone the spear path).
`agent.died` additionally spills the dead agent's
entire inventory onto a pile at the death tile (reducer-internal, no new event —
research R7's debt-opening precedent), and `buildSite` (`terrain.go`) rejects any
tile already holding a pile (FR-007 — goods aren't buried).

**Builder-owned chests** (`build_chest`, spec 013 US3): a fifth structure kind
alongside fire/shelter/oven, gated on `chestPlankCost` (6) planks with a
fire-comparable build duration. The builder is recorded as the chest's `Owner`
permanently (no transfer or inheritance in v1) and the chest gets an empty
`Store`, capped at `chestCap` (48, the same derived `bulk()`). `deposit`/
`withdraw` are instant-on-arrival, planner/plan-only goals resolving to the
nearest chest (`withdraw` with a named `Kind` targets the nearest chest actually
holding it); their completions re-validate the chest still stands and truncate
the move to whichever side is tighter — the chest's free space on deposit, the
taker's free bulk on withdraw. A non-owner `withdraw` is theft: never blocked,
always marked — the executor co-emits a companion batch in the same tick
(`social.chest_taken`, a reason-`"theft"` `social.relation_changed`, the owner's
gossip-seed memory, and witness memories for nearby villagers — [[social-fabric]]
has the full mechanics).

**Food rot** (spec 013 US5): on the same per-game-minute boundary the needs
heartbeat uses, `stepEvents` also sweeps every pile's food batches for ones whose
`SpoilAt` has arrived, emitting one `sim.food_rotted` per (pile, kind) with
same-kind spoiled batches merged — a pure function of (state, tick), the
fuel-sweep pattern. Chest food carries no batches and never rots (FR-010).

**Terrain overlays** (`terrain.go`): chopped trees and harvested forage are
event-sourced state over the static map — `effectiveKind`/`passable` merge
[[worldmap-generation]] with `State.Cleared`/`Harvested`/`Quarried`; forage regrows
after 12 game-hours (`sim.forage_regrown`), dens cool down 6 game-hours after a hunt.
A quarried rock outcrop (spec 012) is different from the other two: it does NOT
revert to Grass — `effectiveKind` renders it as `worldmap.Depleted` permanently (no
regrow in v1), `passable` allows walking across it, but it is neither buildable
(`buildSite`) nor quarryable again. Spec 068's marsh and sand ground covers
([[worldmap-generation]]) deliberately have NO overlay arm here at all: nothing
clears, harvests, or quarries them, so `effectiveKind` always renders whatever the
generator produced, and `passable` admits both alongside grass/forage/Depleted
(marsh/sand carry no resource affordance and are never buildable, same as forage).
Structures (`fire`, `shelter`, `oven`, `chest`)
exist only in state; `warmAt` is a *lit* fire within Manhattan radius 2, or standing on a
shelter (ovens grant no warmth) — decomposed (spec 074-look-cursor, TASK-142)
into a private `warmthSource(s, x, y, tick) (warm bool, source string)` core
that attributes WHICH structure (fire vs. shelter, in the same scan order
`warmAt` always used) provided the warmth; `warmAt` is now a thin wrapper, and
`internal/sim/env.go`'s exported `EnvAt` reads the same core so the TUI's
look-cursor TILE view can show a warmth level/source without a duplicated
radius constant (`gru.md`'s `lightSource` is the sibling decomposition for
light). Behavior and byte output are unchanged — a pure read-side export, no
reducer/tuning change. `fireStructAt`/`litFireAt` locate a fire by
coordinate and test lit-ness for the refuel/cook completion checks. Spec 032
adds two more structure kinds and the first one to affect pathing: `isWall`
names the wall family (`wall_plank`, `wall_stone`); `passable` now checks
`wallAt(s, x, y) != nil` FIRST and refuses the tile if so — a standing wall is
the one structure family that blocks movement (`buildSite`'s generic
"no structure on this tile" scan already kept walls, like every other
structure, un-buildable-over). `wallMaxHP(kind)` derives each kind's ceiling
(`wallPlankHP` 200, `wallStoneHP` 600) for the build stamp, the repair clamp,
and the TUI's damage styling — never stored separately (`WallMaxHP` exports it
for [[tui-client]]). `agentAt(s, x, y)` backs the wall-build occupancy guard
(FR-007: a wall may never land on a tile holding a living agent) — since spec
038 checked only at the completion moment (deferring, then bounded-loud-
failing on a lingering occupant), no longer during mid-work re-validation.
`pathAt(s, x,
y)` reports a `path` structure underfoot — the movement dual-phase cadence's
per-step predicate (above); a path has no `HP` and never blocks (`isWall` is
false for it). Spec 044 (US4) adds an eighth structure kind, `grave` — placed
by the reducer at a death tile, never built by any goal (no recipe, no build
verb); it never blocks movement but, like every structure, blocks `buildSite`
on its tile, and the perception sweep witnesses it like any other structure
([[sim-state-reducer]], [[mental-maps]]).

## Connections

Parent note: [[executor]]. [[worldmap-generation]] supplies the static
terrain (Rock quarry sites, marsh/sand ground covers) this section's overlays
merge with; [[social-fabric]] carries the theft companion batch a non-owner
chest `withdraw` triggers; [[tui-client]] renders wall HP dimming, path
tiles, fire lit/cold state, ground piles, and chest contents;
[[world-migration]] re-places carried souls and over-cap carry on a fresh map
across format-version cuts; [[executor-goal-completions]] is where a build/
demolish/repair/craft goal's completion event is defined.

## Operational notes

The v3 storage economy (spec 013) is exercised by its own suite —
`bulk_cap_test.go`, `ground_pile_test.go`, `chest_test.go`, `theft_test.go`,
`rot_test.go`, `migrate_test.go` — plus an extended `whole_feature_test.go`
pass. Spec 032 (walls, axes, paths) adds `wall_test.go` (blocking/rerouting,
occupancy guard, HP stamping, multi-cycle demolish/repair math and re-arm,
replay determinism), `axe_test.go` (bare-vs-axe yield, ten-use breakage,
bulk truncation, storage round-trip, replay), and `path_speed_test.go` (a
paved corridor halves traversal ticks vs. unpaved) — plus an extended
`whole_feature_test.go` pass exercising all three together.
