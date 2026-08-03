---
name: event-types-crafting-building
description: Crafting/building/goods-movement event rows split from [[event-types]]: agent.crafted/built, wall_chipped/destroyed/repaired, dropped/picked_up/deposited/withdrew. Load when tracing the spec 012/013/032 resource-economy bump (v2/v3 world format), recipe deltas, wall HP, or pile/chest transfers.
kind: concept
sources:
  - internal/sim/executor.go
  - internal/sim/recipes.go
  - internal/sim/terrain.go
verified_against: 012f715f55d8d87317e601ad75686c599d277349
---

# Event types — crafting, building & goods-movement events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.


Spec 086 (agent-named payloads): every agent-referencing field in this
family's payloads is a `sim.AgentRef` — the wire carries
`{"id":N,"name":"…"}` objects (lists element-wise), the name stamped at
emission from the fixed roster via `Ref`/`Refs`; sentinels marshal
`{"id":-1,"name":""}`. Legacy bare-int rows decode through the dual-shape
unmarshal forever and reducer arms fold `.ID`s only — the conventions and
the normative back-compat matrix live on [[event-types]] ("Agent
references are named refs").
Spec 012 (resources/food/crafting) bumped the world save format to
v2 ([[world-save-directory]]): a widened `Inventory`, a new `agent.ate` payload
shape, and nine new event types.

Spec 013 (inventory & storage — bulk cap, ground
piles, builder-owned chests, theft, food rot) bumped it again to **v3**: six more
event types below (drop/pick_up/deposit/withdraw completions, a theft record, and
the food-rot sweep), plus changed semantics on several existing event types (no new
types for these — yield clamping on gathers, net-bulk re-validation on crafts, a
zero-bulk give guard, inventory death-spill) that a v2 replay under this build would
get wrong; the format gate shields old logs from the new semantics. A v1 world
cannot boot under this build — it must run `promptworld migrate` first
([[world-migration]]), chaining 1→2→3 in one run; its sole output event,
`world.migrated`, is also cataloged here.

Spec 032
(walls, axes, paths — terrain) is also format-stable: `Inventory` and `Pile` both
gain `Axes []int` (`omitempty`, a `Spears` clone — remaining harvest uses per
carried axe, sorted ascending, tripling chop/quarry yield), `Structure` gains
`HP` (`omitempty`, walls only — a derived-from-kind max, never stored separately,
the fire lit-ness doctrine) and three new `Kind` values, `wall_plank`/
`wall_stone` (blocking, demolishable, repairable) and `path` (walkable, doubles
movement speed for an agent stepping off it). Four new event types land the
feature — `agent.axe_broke` (a `spear_broke` clone) and the wall work cycle
`agent.wall_chipped`/`agent.wall_destroyed`/`agent.wall_repaired`; `craft_axe`,
`build_wall_plank`/`build_wall_stone`/`build_path`, `demolish`, and `repair` are
new goals reusing the existing `agent.crafted`/`agent.built` types (no new event
types for those).

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `agent.crafted` | `CraftedPayload{agent, kind}` (kind ∈ planks\|refined_stone\|spear\|axe) | work completion (hand-craft; inputs re-validated via `hasItems`, and completion re-validation extends to the net output−input bulk delta, spec 013 US1 — a craft with insufficient inputs, or one whose net wouldn't fit, is not emitted; since spec 096 either resolves LOUDLY via `agent.intent_failed` (`"contested"`) instead of the old bare `agent.intent_done`; only `craft_planks` has a positive net — `craft_axe`'s 1 planks + 1 stone → axe (spec 032 US2) nets like the spear) | recipe delta from `recipes.go` by goal (re-derived from `kind`); a fresh spear appends `spearDurability` (3) to `Spears`; a fresh axe appends `axeDurability` (10) to `Axes`; both kept sorted ascending (harvests/hunts spend the most-worn first), intent cleared |
| `agent.built` | `BuiltPayload{agent, kind, x, y}` (kind ∈ fire\|shelter\|oven\|chest\|wall_plank\|wall_stone\|path) | work completion (site pre-validated as buildable; since spec 013, `buildSite` additionally rejects a tile holding a pile — FR-007; since spec 032 US1, a wall build additionally re-validates `(x,y)` holds no living agent — never entomb the builder or anyone else — and lands ADJACENT: the wall stands on the Res tile while the builder stands on Target; `build_path` (US3) is stand-on-target like fire/oven/chest) | structure added, recipe's inputs spent (via `recipes.go`'s `build_<kind>`); a fresh fire also gets `FuelUntil = tick + 2×fireBurnPerWood`; a fresh chest (spec 013 US3) gets `Owner = agent` (permanent, no transfer in v1) and an empty `Store`; a fresh wall (spec 032 US1) gets `HP = wallMaxHP(kind)` — full health, derived from kind, never stored as a separate max (fire lit-ness doctrine); a path gets no HP (not a wall — never blocks passage), intent cleared |
| `agent.wall_chipped` (spec 032 US1) | `WallWorkPayload{agent, x, y}` | work completion (`demolish`, when the chip would leave the wall standing, `HP − demolishChipHP ≥ 1`; research R5 — multi-cycle demolish) | the wall at `(x,y)`'s `HP -= demolishChipHP`, clamped to never go below 1 (a standing wall never serializes ≤ 0); the agent's `Intent.WorkStart` resets to 0, re-arming the executor's work gate for the next demolish cycle under the same intent — no new scheduling |
| `agent.wall_destroyed` (spec 032 US1) | `WallWorkPayload{agent, x, y}` | work completion (`demolish`, the final chip that would take `HP` to ≤ 0) | removes the wall structure at `(x,y)` — its tile becomes passable again by construction; intent cleared. `guardian.entity_removed` reaches the same end through the miracle path |
| `agent.wall_repaired` (spec 032 US1) | `WallWorkPayload{agent, x, y}` | work completion (`repair`; validity requires a damaged, still-standing wall plus 1 unit of its `wallRepairMaterial(kind)` carried) | consumes 1 unit of that material and restores `HP` by `repairHPPerUnit`, clamped to `wallMaxHP(kind)`; if still damaged AND material remains, `Intent.WorkStart` resets to 0 to re-arm another cycle (intent kept) — otherwise intent cleared |
| `agent.dropped` | `DroppedPayload{agent, x, y, kind, n}` | executor, `drop` completion (instant, agent's current tile — spec 013 US2, planner/plan-only) | `Inv[kind] −= n`; the tile's pile created-or-merged `+= n` (food becomes/merges a batch stamped `spoil_at = tick + rotWindowTicks`; spears AND axes (spec 032 US2) move most-worn-first with their durabilities), intent cleared |
| `agent.picked_up` | `PickedUpPayload{agent, x, y, kind, n}` | executor, `pick_up` completion (instant on arrival at a pile on/adjacent tile; one event per kind moved in the batch) | pile `−= n` (food oldest-batch-first), `Inv[kind] += n`; an emptied pile is removed; intent cleared on the last event of the batch |
| `agent.deposited` | `DepositedPayload{agent, x, y, kind, n}` | executor, `deposit` completion at a chest (instant on arrival — spec 013 US3) — a vanished/full chest, or nothing left to give, resolves LOUDLY via `agent.intent_failed` (spec 096, `"contested"`); an empty `Kind` resolves LOUDLY too but distinctly (`"invalid"` — [[event-types-agent-intents]]) | `Inv[kind] −= n`, chest `Store[kind] += n`, both clamped to the chest's free space (`chestCap − bulk(*Store)`); intent cleared |
| `agent.withdrew` | `WithdrewPayload{agent, x, y, kind, n, owner}` | executor, `withdraw` completion at a chest (instant on arrival) — a vanished chest, or nothing available to take, resolves LOUDLY via `agent.intent_failed` (spec 096, `"contested"` — [[event-types-agent-intents]]) | chest `Store[kind] −= n`, `Inv[kind] += n`, clamped to the taker's free bulk; intent cleared; a non-owner taker co-emits the theft companion batch (`social.chest_taken` + a reason-`"theft"` `social.relation_changed` + owner/witness `agent.memory_added`, all in the same batch — [[social-fabric]]) |
