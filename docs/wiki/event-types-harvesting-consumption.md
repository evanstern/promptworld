---
name: event-types-harvesting-consumption
description: Harvest/consumption event rows split from [[event-types]]: agent.foraged/chopped/hunted/quarried/collected_water, sim.food_rotted, agent.cooked/bathed/refueled/ate, agent.spear_broke/axe_broke, sim.fire_burned_out. Load when tracing yield math, tool durability, food rot, cooking, or fire fuel.
kind: concept
sources:
  - internal/sim/executor.go
  - internal/sim/terrain.go
verified_against: b6a20eaa4da1073a69959a5aff69591d931103a9
---

# Event types — harvesting & consumption events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `agent.foraged` / `agent.chopped` / `agent.hunted` | `HarvestPayload{agent, x, y}` | work completion (spec 013: skipped entirely — no event — when the taker's free bulk is zero, US1-AS1, so depletion never happens with no room to carry the take) | +FoodRaw (forage `forageYieldV2`; hunt `huntYieldBare`, or `huntYieldSpear` + spends `Spears[0]`'s last use if carrying one) / +wood (chop `chopYieldBare` (1) bare-handed, or `chopYieldAxe` (3, spec 032 US2) carrying an axe — re-derived from the same pre-mutation state the emitter checked, spending `Axes[0]`'s last use; a spent-to-zero axe co-emits `agent.axe_broke` in the same batch), each clamped to the taker's pre-event free bulk (`bulkCap − bulk(Inv)`, spec 013 US1-AS2 — the forfeited remainder is lost, not refunded); overlay (harvest/cleared/den cooldown) applies regardless of the clamp, intent cleared |
| `agent.quarried` | `HarvestPayload{agent, x, y}` | work completion (rock outcrop; same zero-free-bulk skip as above) | +Stone: `quarryYieldBare` (1) bare-handed, or `quarryYieldAxe` (3, spec 032 US2) carrying an axe, spending `Axes[0]`'s last use the same way chop does (co-emitting `agent.axe_broke` when spent) — clamped to free bulk; `(x,y)` appended to `State.Quarried` regardless (permanent — [[worldmap-generation]], [[executor]] — the outcrop depletes even when the yield is forfeit), intent cleared |
| `agent.collected_water` | `HarvestPayload{agent, x, y}` | work completion (any water tile; same zero-free-bulk skip) | +`collectWaterYield` Water clamped to free bulk, intent cleared (no overlay — water never depletes) |
| `sim.food_rotted` | `FoodRottedPayload{x, y, kind, n}` | executor, per-game-minute rot sweep (spec 013 US5; same-kind spoiled batches merged per pile per sweep) | pile's food batches with `spoil_at ≤ tick` matching `kind` removed (up to `n`, oldest first); an emptied pile is removed; chest food never batches, so chests are never reached (FR-010) |
| `agent.cooked` | `CookedPayload{agent, station, consumed, produced, kind}` (station ∈ fire\|oven; kind ∈ food_cooked\|meals) | work completion (cook) | −FoodRaw(consumed), +kind(produced); an oven cook also −1 Wood, intent cleared |
| `agent.bathed` | `BathedPayload{agent, morale_after, warmth_after}` | work completion (bathe, oven only) | −1 Water, −1 Wood, Morale/Warmth set to the absolute post-cap values, intent cleared |
| `agent.refueled` | `RefueledPayload{agent, x, y, fuel_until}` | reflex/planner (instant on arrival) | −1 Wood, the fire at `(x,y)`'s `FuelUntil` set to the absolute (already-capped) deadline, intent cleared |
| `agent.spear_broke` | `SpearBrokePayload{agent}` | work completion (hunt, companion to `agent.hunted` in the same batch) | removes the now-zero `Spears[0]` entry |
| `agent.axe_broke` (spec 032 US2) | `AxeBrokePayload{agent}` | work completion (chop or quarry, companion to `agent.chopped`/`agent.quarried` in the same batch — the `agent.spear_broke` clone) | removes the now-zero `Axes[0]` entry |
| `sim.fire_burned_out` | `FireBurnedOutPayload{x, y}` | `stepEvents`, once per fuel-window transition (`tick-1 < FuelUntil <= tick`) | none — lit-ness stays derived from `FuelUntil`; chronicle/TUI material, plus a low-salience witness memory for nearby living agents |
| `agent.ate` | `AtePayload{agent, meals, cooked, raw, food_after}` | reflex/planner (instant), most-nutritious-first (Meals→FoodCooked→FoodRaw) to satiety (`eatOutcome`) | −Meals/−FoodCooked/−FoodRaw by the consumed counts, Food need set to the absolute `food_after` |
