---
name: executor-agent-state
description: The agent body's persisted shape — eight named villagers, integer Needs/Inventory, the mental-map/vector/intent-log/needs-anchor pointer fields other subsystems own end to end, and where tuning constants live. Load when tracing what an Agent struct actually carries or where a decay/yield/cost constant is defined.
kind: component
sources:
  - internal/sim/agents.go
verified_against: cffd9a79bbed61ccac573d97c6cf544565b40336
---

# Executor — agent state

Child of [[executor]] — the persisted shape of an agent body: named villagers,
the Needs/Inventory economy, and the pointer fields other subsystems own end
to end but the executor threads through.

## How it works

**Agents** (`agents.go`): eight named bodies (`sim.AgentNames`) with authored
personas ([[agent-mind]]). Since spec 041 each agent also carries `Map
*MentalMap` (`omitempty`, the Journal/Hail pointer precedent — a pre-041
snapshot round-trips byte-identically) — the private spatial-knowledge store
[[mental-maps]] owns end to end; the executor's only stake in it is emitting
the events that grow and correct it (below). Since spec 042 each agent also
carries `omitempty` `SitVec`/`SitVecModel`/`SitVecTick`, and each `Memory` the
executor's salience table appends (below) carries `omitempty` `Seq`/`Vec`/
`VecModel` — the embedding-retrieval identity and vector [[memory-retrieval]]
owns end to end; the executor has no stake in either, it neither computes nor
reads a vector, only emits the `agent.memory_added` events the reducer later
stamps a `Seq` onto. Since spec 043 each agent further carries an `IntentLog`
ring (`omitempty`, cap `intentLogCap` = 8) of `IntentRecord{Goal, Source,
Reason, Tick, Outcome, OutcomeTick}` — the recent-intent history the decision
prompt's self-history block renders — and a `NeedsAnchor`/`NeedsAnchorTick`
window-edge needs snapshot (a POINTER with `omitempty`, the Journal/Hail/Map
precedent) the `agent.needs_changed` arm rolls forward once
`trajectoryWindowTicks` (1800) of game time has elapsed, backing the prompt's
rising/falling/steady need trajectories; both are maintained entirely by
reducer arms (`appendIntent`/`stampIntentOutcome`/`stampOrAppendExpired`,
agents.go) — replay-safe by construction, and pre-043 snapshots round-trip
byte-identically ([[decision-context]] owns the rendering surface). Since
spec 062 each agent also carries `LastMindIntentDone` (`omitempty`) — the
reflex PREP gate's yield-window anchor ([[reflex-policy]]) — armed by the
SAME `agent.intent_done` completion arm reading `stampIntentOutcome`'s
(now dual-valued) closed-record `Source` return; the executor itself has no
stake in it beyond emitting the completion event the arm reads.
`Needs{Health, Food, Rest, Warmth, Morale}` are integers 0..1000 —
integer math keeps decay byte-deterministic across platforms. `Inventory` (v2,
format_version 2 — [[world-save-directory]]) carries `Wood`, `Stone`, `Water`,
`Planks`, `RefinedStone`, `FoodRaw`, `FoodCooked`, `Meals` (all `omitempty` ints),
and `Spears []int` — remaining uses per carried spear, sorted ascending so a hunt
always spends the most-worn one first. `Axes []int` (spec 032) mirrors `Spears`
exactly: remaining harvest uses per carried axe, sorted ascending, spent
most-worn-first on chop/quarry. The legacy `Food int` field is gone; a v1
world must run `promptworld migrate` ([[world-migration]]) before it can boot under
this build. Most tuning constants (decay rates, action durations, yields, costs,
thresholds) sit at the top of `agents.go`; the v2 economy's constants (food
restores, spear durability, gather/craft/build/station magnitudes) are
grouped under their own "spec 012" block there — except the two fire-fuel
dials (`refuelDyingBelow`, `fireBurnPerWood`), which spec 048 promoted to
per-world dials and relocated to `internal/sim/tuning.go` as
`defaultRefuelDyingBelow`/`defaultFireBurnPerWood` behind the nil-safe
`State.RefuelDyingBelow()`/`State.FireBurnPerWood()` accessors ([[world-tuning]])
— and a separate "spec 032" block
holds the walls/axes/paths tuning surface (`wallPlankHP` 200, `wallStoneHP` 600 —
at least 2x the plank wall per FR-003 — `axeDurability` 10 harvest uses,
`chopYieldBare`/`chopYieldAxe` 1/3, `quarryYieldBare`/`quarryYieldAxe` 1/3,
`demolishChipHP` 100 per work cycle, `repairHPPerUnit` 100 per material spent,
`pathStoneCost` 1); spec 038 adds `wallOccupancyGraceTicks` (120) to the same
block — ticks past a wall's due tick (`WorkStart + workDuration`) that
completion may defer on an occupied reserved tile before failing loudly, a
pure function of `WorkStart` so no new persisted state is needed. The legacy flat `chopWood`/`quarryYield` (2) constants are
deleted, replaced by the bare/axe pairs. The recipe table itself lives
in `recipes.go` (mirroring `specs/012-resources-food-crafting/contracts/recipes.md`
and `specs/032-walls-axes-paths/contracts/recipes.md` — `recipes_test.go` asserts
the tables agree).

## Connections

Parent note: [[executor]]. [[agent-mind]] authors the personas this section's
eight named bodies carry; [[mental-maps]] owns `Map`; [[memory-retrieval]]
owns the embedding fields on `SitVec\*` and each `Memory`'s `Seq`/`Vec`/
`VecModel`; [[decision-context]] owns the rendering surface for `IntentLog`/
`NeedsAnchor`; [[reflex-policy]] owns the yield-window anchor
`LastMindIntentDone` arms; [[world-tuning]] owns the per-world tuning dials
relocated out of `agents.go`; [[world-save-directory]] tracks the
`Inventory` format version.
