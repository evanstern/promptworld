---
name: event-types-harvesting-consumption
description: Harvest/consumption event rows split from [[event-types]]: agent.foraged/chopped/hunted/quarried/collected_water, sim.food_rotted, agent.cooked/bathed/refueled/ate, agent.spear_broke/axe_broke, sim.fire_burned_out. Load when tracing yield math, tool durability, food rot, cooking, or fire fuel.
kind: concept
sources:
  - internal/sim/executor.go
  - internal/sim/terrain.go
verified_against: 012f715f55d8d87317e601ad75686c599d277349
---

# Event types — harvesting & consumption events

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
| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `agent.foraged` / `agent.chopped` / `agent.hunted` | `HarvestPayload{agent, x, y}` | work completion (spec 013: skipped entirely — no harvest event — when the taker's free bulk is zero, US1-AS1, so depletion never happens with no room to carry the take; since TASK-196 that skip resolves LOUDLY via `agent.intent_failed` (`"pack full"` — [[event-types-agent-intents]]) rather than the bare `agent.intent_done` it used to share with a successful harvest) | +FoodRaw (forage `forageYieldV2`; hunt `huntYieldBare`, or `huntYieldSpear` + spends `Spears[0]`'s last use if carrying one) / +wood (chop `chopYieldBare` (1) bare-handed, or `chopYieldAxe` (3, spec 032 US2) carrying an axe — re-derived from the same pre-mutation state the emitter checked, spending `Axes[0]`'s last use; a spent-to-zero axe co-emits `agent.axe_broke` in the same batch), each clamped to the taker's pre-event free bulk (`bulkCap − bulk(Inv)`, spec 013 US1-AS2 — the forfeited remainder is lost, not refunded); overlay (harvest/cleared/den cooldown) applies regardless of the clamp, intent cleared. **Spec 081**: a completed `agent.chopped` also removes the `tree` place-fact at `(x,y)` from the actor's mental map and from every awake living in-radius witness's map (`removeHarvestedFact`, [[sim-state-reducer]]), and the executor mints the actor a first-person `agent.memory_added` ("Felled the tree at (x,y).", `salChop`) — see [[mental-map-perception]] |
| `agent.quarried` | `HarvestPayload{agent, x, y}` | work completion (rock outcrop; same zero-free-bulk skip as above, and the same TASK-196 `"pack full"` `agent.intent_failed` resolution) | +Stone: `quarryYieldBare` (1) bare-handed, or `quarryYieldAxe` (3, spec 032 US2) carrying an axe, spending `Axes[0]`'s last use the same way chop does (co-emitting `agent.axe_broke` when spent) — clamped to free bulk; `(x,y)` appended to `State.Quarried` regardless (permanent — [[worldmap-generation]], [[executor]] — the outcrop depletes even when the yield is forfeit), intent cleared. **Spec 081**: mirrors chop — removes the `rock` place-fact at `(x,y)` from the actor's and every awake in-radius witness's map, and mints the actor a first-person "Quarried the outcrop at (x,y)." memory (`salQuarry`) |
| `agent.collected_water` | `HarvestPayload{agent, x, y}` | work completion (any water tile; same zero-free-bulk skip, same TASK-196 `"pack full"` resolution) | +`collectWaterYield` Water clamped to free bulk, intent cleared (no overlay — water never depletes) |
| `sim.food_rotted` | `FoodRottedPayload{x, y, kind, n}` | executor, per-game-minute rot sweep (spec 013 US5; same-kind spoiled batches merged per pile per sweep) | pile's food batches with `spoil_at ≤ tick` matching `kind` removed (up to `n`, oldest first); an emptied pile is removed; chest food never batches, so chests are never reached (FR-010) |
| `agent.cooked` | `CookedPayload{agent, station, consumed, produced, kind}` (station ∈ fire\|oven; kind ∈ food_cooked\|meals) | work completion (cook) | −FoodRaw(consumed), +kind(produced); an oven cook also −1 Wood, intent cleared |
| `agent.bathed` | `BathedPayload{agent, morale_after, warmth_after}` | work completion (bathe, oven only) | −1 Water, −1 Wood, Morale/Warmth set to the absolute post-cap values, intent cleared |
| `agent.refueled` | `RefueledPayload{agent, x, y, fuel_until}` | reflex/planner (instant on arrival) | −1 Wood, the fire at `(x,y)`'s `FuelUntil` set to the absolute (already-capped) deadline, intent cleared |
| `agent.spear_broke` | `SpearBrokePayload{agent}` | work completion (hunt, companion to `agent.hunted` in the same batch) | removes the now-zero `Spears[0]` entry |
| `agent.axe_broke` (spec 032 US2) | `AxeBrokePayload{agent}` | work completion (chop or quarry, companion to `agent.chopped`/`agent.quarried` in the same batch — the `agent.spear_broke` clone) | removes the now-zero `Axes[0]` entry |
| `sim.fire_burned_out` | `FireBurnedOutPayload{x, y}` | `stepEvents`, once per fuel-window transition (`tick-1 < FuelUntil <= tick`) | none — lit-ness stays derived from `FuelUntil`; chronicle/TUI material, plus a low-salience witness memory for nearby living agents |
| `agent.ate` | `AtePayload{agent, meals, cooked, raw, food_after}` | reflex/planner (instant), most-nutritious-first (Meals→FoodCooked→FoodRaw) to satiety (`eatOutcome`) | −Meals/−FoodCooked/−FoodRaw by the consumed counts, Food need set to the absolute `food_after` |
