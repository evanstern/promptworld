---
name: event-types
description: The event taxonomy — namespaced types, their payload structs, who emits them, and their reducer effects
kind: concept
sources:
  - internal/sim/state.go
  - internal/sim/agents.go
  - internal/sim/executor.go
  - internal/sim/recipes.go
  - internal/sim/gru.go
  - internal/sim/loop.go
  - internal/sim/landing.go
  - internal/sim/miracles.go
  - internal/sim/journal.go
  - internal/sim/consolidate.go
  - internal/sim/terrain.go
  - internal/sim/metatron.go
  - internal/sim/morgue.go
  - internal/daemon/daemon.go
verified_against: dee5f4bf60093cb5d775e10c8ced41c7e5b385ec
---

# Event types

Every event has a namespaced `type` and a canonical-JSON payload defined as a Go
struct in `internal/sim` (structs, never maps, so bytes are deterministic; core
payloads live in `state.go`, families note their own file below).
This catalog is the contract downstream consumers (chronicle, Metatron digests, the
TUI) will read. Spec 012 (resources/food/crafting) bumped the world save format to
v2 ([[world-save-directory]]): a widened `Inventory`, a new `agent.ate` payload
shape, and nine new event types. Spec 013 (inventory & storage — bulk cap, ground
piles, builder-owned chests, theft, food rot) bumped it again to **v3**: six more
event types below (drop/pick_up/deposit/withdraw completions, a theft record, and
the food-rot sweep), plus changed semantics on several existing event types (no new
types for these — yield clamping on gathers, net-bulk re-validation on crafts, a
zero-bulk give guard, inventory death-spill) that a v2 replay under this build would
get wrong; the format gate shields old logs from the new semantics. A v1 world
cannot boot under this build — it must run `promptworld migrate` first
([[world-migration]]), chaining 1→2→3 in one run; its sole output event,
`world.migrated`, is also cataloged here. Spec 019 (grounded memories — situated
episodic memories + agent journal) adds **no** format bump: every addition is
`omitempty`, byte-stable against pre-019 logs. `MemoryAddedPayload` gains
situated context (`where`/`why`/`conv`), `IntentSetPayload` gains `reason`
(carried onto the intent so the executor can bake it into a memory's `why` at
completion), and TWO new whitelisted event types — `journal.entry_written` /
`journal.entry_deleted` — drive the agent-authored journal ([[agent-mind]]).
Spec 028 (adaptive throttle) likewise adds **no** format bump: `State` gains
`RequestedSpeed` (`omitempty` — absent means ungoverned, so every pre-028
snapshot is a valid ungoverned state), and two new reducer-applied types,
`clock.governor_shed`/`clock.governor_recovered`, land the governor's
speed-ladder decisions ([[cognition]]).
Spec 029 (Metatron agency — [[metatron]], [[metatron-orders]]) likewise adds
**no** format bump: `State` gains `MetatronOrders []MetatronOrder`
(`omitempty` — an empty order set is genuinely zero-value, unlike
`MetatronCharges`'s spent-to-zero precedent, so a pre-029 snapshot with the
field absent unmarshals to nil), and FOUR new event types drive a standing
order's lifecycle — `metatron.order_placed` (monitor_and_act), one-shot
`metatron.order_triggered` (the trigger worker, live-only, never replayed),
`metatron.order_cancelled` (cancel_order), and executor-emitted
`metatron.order_expired` (the `charge_regenerated` pattern: a pure function
of state + tick, so it reproduces on replay with no angel running). The same
spec retires `nudge_dream`/`nudge_omen` from the tool registry in favor of
`send_vision`, so `metatron.nudged`'s `form` domain is now `vision` (exactly
one living target, any hour) / `omen` (≥1 living targets, night-only,
`State.Night`) / `dream` (legacy, grandfathered: accepted on replay for
historical events, but no live tool can produce a new one).
Spec 030 (epistemic hygiene — honest
provenance, hearsay decay, attribution-preserving gists) is likewise format-stable:
`MemoryAddedPayload` gains `Origin` (`omitempty`, the closed-vocabulary
provenance class — `action`/`witness`/`report`/`omen`/`gist`/`digest`, defined in
`internal/sim/memory.go` — stamped at every emission site; absent classifies as
secondhand, the conservative default, and `DirectPerception` is the only test the
belief validator runs against it), `BeliefRevisedPayload` gains `Evidence`
(`omitempty`, the resolved `MemoryRef{tick, hash}` identities a revision cites)
and `Direct` (`omitempty`, whether any cited evidence is direct perception — the
revision only refreshes the belief's decay anchor when true), and
`ConsolidatedPayload` gains `Coerced` (`omitempty`, telemetry: how many beliefs
the validator downgraded off `"witnessed"` for lack of direct evidence, never a
rejection). One new whitelisted type, `agent.belief_reinforced`
(`BeliefReinforcedPayload{agent, belief_id}`), re-anchors a held belief's decay
clock at the grounded-observation seam — spec 030 ships the consumer (whitelist +
reducer arm) only, no in-tree emitter yet ([[nightly-consolidation]]). Spec 032
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
Spec 034 (provider health conditions + preflight — [[llm-provider-health]]) is
also format-stable: one new whitelisted type, `daemon.llm_warning`
(`LLMWarningPayload{provider, kind, detail, remedy, active}`), rides the
daemon's operator-event door ([[sim-loop]]'s `InjectOperator`) on every
provider-health transition (raise/reclassify/clear); it is operator-facing
only, alongside `daemon.started`/`stopped` under the existing `daemon.*`
no-op convention.
Spec 039 (teaching-world speed posture — [[daemon-lifecycle]], [[cognition]])
adds no new event type and no new emission door: on a teaching world with an
orchestrator, boot computes the planner-safe posture rung and applies it
through the loop's ordinary `set_speed` command, so it lands as an EXISTING
`clock.speed_set` (below) with an ordinary `SpeedSetPayload{speed}` — a new
emitter of that type, not a new type, which is why the format stays put.
Spec 041 ([[mental-maps]] — per-agent spatial knowledge) is likewise
format-stable: `Agent` gains `Map *MentalMap` (`omitempty`, the Journal/Hail
pointer precedent — a pre-041 snapshot with the field absent round-trips
byte-identically), and FOUR new whitelisted/reducer event types drive it —
`agent.saw` (perception sweep witnesses), `agent.map_corrected` (a remembered
place found gone), `social.place_told` (directions exchanged in talk), and
`metatron.place_revealed` (a vision's optional place grant) — full shapes and
reducer effects in the table below.
Spec 042 ([[memory-retrieval]] — embedding-augmented memory retrieval) is also
format-stable: `Memory` gains `omitempty` `Seq`/`Vec`/`VecModel` (`Seq` the
emitting `agent.memory_added` event's store seq — a pre-042 memory's field is
absent/0) and `Agent` gains `omitempty` `SitVec`/`SitVecModel`/`SitVecTick`,
so every pre-042 snapshot round-trips byte-identically. THREE new whitelisted
event types drive it: `agent.memory_embedded` and `agent.situation_embedded`
(the mind-side embedder's two state-mutating vector companions) and
`cog.memory_divergence` (a reducer-no-op telemetry record, the `cog.*`
convention) — full shapes and reducer effects in the table below. A world's
`memory_relevance` manifest flag (`""`/`shadow`/`on`, validated at
`world.Open`) gates whether the embedder and divergence recording run at all;
it carries no event of its own.
Spec 043 (context grounding — [[decision-context]]) adds NO new event type at
all: `Agent` gains `omitempty` `IntentLog []IntentRecord` (the recent-intent
ring, cap 8) and `NeedsAnchor`/`NeedsAnchorTick` (the need-trajectory window
anchor), both maintained by EXISTING intent-lifecycle and needs arms (rows
below note each), so every pre-043 snapshot round-trips byte-identically —
and `agent.intent_rejected`, formerly a reducer no-op, becomes state-mutating
(it appends an already-closed ring record). `CogThoughtPayload` gains three
additive LAST `omitempty` fields — `prompt_bytes`/`block_bytes`/
`dropped_blocks`, the assembled decision-prompt size, per-block byte
breakdown, and budget-dropped blocks — stamped ONLY on planner-class
emissions, so pre-043 events and every non-planner emission marshal
byte-identically.
Spec 044 (run outcomes, morgue, gru escalation, graves — [[morgue]]) is also
format-stable: `State` gains `omitempty` `Deaths` (the run's death ledger),
`Ended`/`RunEnd` (the terminal run-over latch and summary),
`CharterFingerprint`, and the bounded `MorgueEpilogues` ring, and `Structure`
gains a `grave` kind (reducer-placed at a death tile, never player-built).
THREE new event types land the feature — executor-emitted `run.ended` (never
injectable: it opens two whole new namespaces, `run.*` and, via the
whitelisted pair, `morgue.*`) plus whitelisted `metatron.charter_observed` and
`morgue.epilogue`, the only two additions `TestWhitelistDiffIdentical` accepts
against the spec's declared boundary widening. Two existing rows changed
semantics rather than shape: `agent.died`'s cause domain gains `"gru"` (an
escalated kill emitted by `gruStep` itself) and its reducer arm gains the
ledger append + grave placement, and `gru.attacked`'s `health` may now be 0
when the target was already weakened below `nearDeathBelow` ([[gru]]).

## How it works

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `world.created` | `WorldCreatedPayload{name, seed}` | CLI `new`, tick 0 | none (genesis marker) |
| `world.migrated` | `WorldMigratedPayload{from_format, source_events, source_tick, state}` (`state` embeds the full canonical `sim.State`) | `promptworld migrate` (client-side, offline — [[world-migration]]), once, right after a fresh `world.created` | replaces `State` wholesale (after checking `state.Seed` matches — a foreign payload is a no-op); the log alone (`world.created` → `world.migrated`) reproduces the migrated world with zero snapshots |
| `clock.paused` / `clock.resumed` | `{}` | loop command | pause flag (+ snapshot on pause) |
| `clock.speed_set` | `SpeedSetPayload{speed}` | loop command (player `set_speed`, or — since spec 039 — the daemon's teaching-posture default applying its computed rung through the same command at boot, [[daemon-lifecycle]]) | `Speed` updated; since spec 028 also clears `State.RequestedSpeed` — a player command always collapses governed state (FR-009) |
| `clock.governor_shed` / `clock.governor_recovered` (spec 028 FR-008) | `GovernorPayload{requested, from, to, debt, jobs}`, shared by both | the daemon's governor sampler via the loop's `govern` command ([[cognition]], [[daemon-lifecycle]]) | `Speed = to`; `RequestedSpeed = requested` (shed) or cleared when `to == requested` (recovered reaching the ceiling); `EffectiveRate` follows `to` unless `Degraded` — never silent, so an operator can reconstruct every governed interval from the log alone (SC-005) |
| `clock.degraded` / `clock.recovered` | `DegradedPayload{effective_rate}` / `{}` | loop auto-slow | degradation flags |
| `sim.day_started` / `sim.night_started` | `DayPayload{day}` | executor, 06:00/22:00 | `Night` flag only — waking is explicit |
| `sim.forage_regrown` | `RegrownPayload{x, y}` | executor, regrow tick | harvest overlay removed |
| `agent.intent_set` | `IntentSetPayload{agent, goal, target, res, source, kind?, qty?, job?, reason?}` | reflex (grace-gated), planner injection, or a plan step firing | intent installed; `source` (`reflex`/`planner`/`plan`/`meeting`) says which mind chose it; also stamps `Agent.LastGoal`/`LastGoalTick` (spec 015 — never cleared by any event, the villagers tab's past-objective line, [[tui-client]]); `job` (spec 017, omitempty) is set ONLY at the `inject_intent` landing site from `InjectArgs.JobID` — a planner-loop landing carries its job id, reflex/executor-authored intents carry none; `reason` (spec 019, omitempty, now the LAST field) is likewise set ONLY at that landing site from `InjectArgs.Reason` — the planner's free-text reason, copied onto `Intent.Reason` by the reducer so it survives to completion where the executor bakes it into a memory's `why`; both `omitempty` tails stay empty on reflex/executor emissions, so those marshal byte-identically to pre-feature; spec 043 US1: also appends an `IntentRecord{goal, source, reason, tick}` to `Agent.IntentLog` (ring, cap 8) — a previous still-open record stays open, so an override reads as open-then-new ([[decision-context]]) |
| `agent.work_started` | `WorkStartedPayload{agent, tick}` | executor at target | `WorkStart` stamped |
| `agent.intent_done` | `AgentPayload{agent}` | executor (done/invalid/unreachable) | intent cleared — but since spec 038, a **build** goal (`build_fire`/`build_shelter`/`build_oven`/`build_chest`/`build_path`/`build_wall_plank`/`build_wall_stone`) whose mid-work re-validation fails no longer funnels through here; it emits the distinct `agent.build_failed` below instead. `intent_done` remains the resolution for successful non-build completion paths, non-build no-ops (craft/cook/bathe/deposit contested re-checks), and every non-build goal's invalid/unreachable exit; spec 043 US1: also closes the newest still-open `IntentLog` record `"done"` |
| `agent.build_failed` (spec 038) | `BuildFailedPayload{agent, goal, reason}` | executor, mid-work re-validation of a build goal (the seven `build_*` above) — emitted **instead of** a bare `agent.intent_done` when the build genuinely fails: `reason` is `site no longer buildable` (any build goal whose `buildSite` re-check fails) or `site blocked too long` (walls only, once a reserved-tile occupant outlasts `wallOccupancyGraceTicks` past the due tick — [[executor]]) | intent cleared (`Intent = nil`, `IdleSince` stamped) — identical to `agent.intent_done`, so no material spent and no structure stands; a paired same-tick `agent.memory_added` (`OriginAction`, `salShelter`) rides along stating the build did NOT complete and why, so the builder can falsify a phantom-structure belief; the builder's mind re-arms its planner exactly as for `agent.intent_done` ([[agent-mind]]). Distinct from `agent.intent_rejected` (up-front landing refusal) — this clears an accepted intent; spec 043 US1: also closes the newest still-open `IntentLog` record `"failed"`. TUI digest renders it as a failure line (builder, goal, reason), never "finished" ([[tui-client]]) |
| `agent.moved` | `AgentMovedPayload{agent, x, y}` | executor pathing | position updated; spec 041: the mover's surroundings (perception radius) are marked explored in its private mental map, and mover + awake bystanders within the witness radius record each other's positions as peer sightings (`MentalMap.Peers` — what `talk_to`/`seek` resolve against) — silent derived bookkeeping, no companion event (`agent.woke` and `metatron.entity_moved` villager moves run the same sighting pass) |
| `agent.saw` (spec 041) | `SawPayload{agent, facts}` — `facts` are fully-baked `PlaceFact`s (kind, x, y, seen, prov, src?, detail?), sorted (kind, x, y) at emission | executor perception sweep, on the agent's movement-cadence beat: ground truth within the witness radius (structures, ground piles, standing trees / unharvested forage / unquarried rock, water shoreline, dens) diffed against the agent's mental map — new, changed (a fire's `FuelUntil` detail), horizon-staled, or provenance-upgraded facts only, so a settled map emits nothing | each fact upserted verbatim into `Agents[agent].Map.Facts` (provenance `witnessed`, `Seen` = event tick); digest-only — deliberately NO chronicle line (too chatty) and not an absorb trigger |
| `social.place_told` (spec 041 US5) | `PlaceToldPayload{from, to, facts}` — up to 2 facts per direction, fully baked at emission: told provenance, the TELLER's `seen` (secondhand is never fresher — staleness is the trust model), `src` = the immediate teller, `detail` carried verbatim; canonical (kind, x, y) order | executor talk sidecar (`talkEvents`, beside the `TellableFor` rumor slot, every founded talk incl. hail-founded): per direction, the teller's fresh facts the listener lacks-or-holds-staler, selected freshest → nearest-to-listener → coordinate order | upsert into the RECEIVER's map only where absent-or-staler (`Seen` compared — fresher knowledge never loses to secondhand); companion situated memories both sides ride the same batch ("Told Birch about the fire at (x,y)." / "Birch told you of a fire at (x,y)."); chronicle grammar line; not an absorb trigger |
| `agent.map_corrected` (spec 041 US3) | `MapCorrectedPayload{agent, gone}` — `gone` carries the facts AS REMEMBERED (verbatim from the agent's map, canonical order; context baked at emission for narration) | the same perception sweep, right after the agent's `agent.saw`: remembered FRESH facts within the witness radius that are ABSENT from ground truth (`groundFactPresent` — absence of the PLACE, not of its availability: a harvested forage spot or cooling den persists; a chopped tree, quarried-out outcrop, drained pile, or removed structure is gone) | each gone fact removed from the agent's map; a situated first-person discovery memory rides the SAME batch per fact as a companion `agent.memory_added` (`mapCorrectedText`, `salMapCorrected`, Origin witness — memories accrete only via `agent.memory_added`); chronicle grammar line; absorb trigger when a removed fact matches the agent's current intent target (the planner re-arms — [[agent-mind]]) |
| `agent.foraged` / `agent.chopped` / `agent.hunted` | `HarvestPayload{agent, x, y}` | work completion (spec 013: skipped entirely — no event — when the taker's free bulk is zero, US1-AS1, so depletion never happens with no room to carry the take) | +FoodRaw (forage `forageYieldV2`; hunt `huntYieldBare`, or `huntYieldSpear` + spends `Spears[0]`'s last use if carrying one) / +wood (chop `chopYieldBare` (1) bare-handed, or `chopYieldAxe` (3, spec 032 US2) carrying an axe — re-derived from the same pre-mutation state the emitter checked, spending `Axes[0]`'s last use; a spent-to-zero axe co-emits `agent.axe_broke` in the same batch), each clamped to the taker's pre-event free bulk (`bulkCap − bulk(Inv)`, spec 013 US1-AS2 — the forfeited remainder is lost, not refunded); overlay (harvest/cleared/den cooldown) applies regardless of the clamp, intent cleared |
| `agent.quarried` | `HarvestPayload{agent, x, y}` | work completion (rock outcrop; same zero-free-bulk skip as above) | +Stone: `quarryYieldBare` (1) bare-handed, or `quarryYieldAxe` (3, spec 032 US2) carrying an axe, spending `Axes[0]`'s last use the same way chop does (co-emitting `agent.axe_broke` when spent) — clamped to free bulk; `(x,y)` appended to `State.Quarried` regardless (permanent — [[worldmap-generation]], [[executor]] — the outcrop depletes even when the yield is forfeit), intent cleared |
| `agent.collected_water` | `HarvestPayload{agent, x, y}` | work completion (any water tile; same zero-free-bulk skip) | +`collectWaterYield` Water clamped to free bulk, intent cleared (no overlay — water never depletes) |
| `agent.crafted` | `CraftedPayload{agent, kind}` (kind ∈ planks\|refined_stone\|spear\|axe) | work completion (hand-craft; completion re-validation extends to the net output−input bulk delta, spec 013 US1 — a craft whose net wouldn't fit is not emitted, `agent.intent_done` only; only `craft_planks` has a positive net — `craft_axe`'s 1 planks + 1 stone → axe (spec 032 US2) nets like the spear) | recipe delta from `recipes.go` by goal (re-derived from `kind`); a fresh spear appends `spearDurability` (3) to `Spears`; a fresh axe appends `axeDurability` (10) to `Axes`; both kept sorted ascending (harvests/hunts spend the most-worn first), intent cleared |
| `agent.built` | `BuiltPayload{agent, kind, x, y}` (kind ∈ fire\|shelter\|oven\|chest\|wall_plank\|wall_stone\|path) | work completion (site pre-validated as buildable; since spec 013, `buildSite` additionally rejects a tile holding a pile — FR-007; since spec 032 US1, a wall build additionally re-validates `(x,y)` holds no living agent — never entomb the builder or anyone else — and lands ADJACENT: the wall stands on the Res tile while the builder stands on Target; `build_path` (US3) is stand-on-target like fire/oven/chest) | structure added, recipe's inputs spent (via `recipes.go`'s `build_<kind>`); a fresh fire also gets `FuelUntil = tick + 2×fireBurnPerWood`; a fresh chest (spec 013 US3) gets `Owner = agent` (permanent, no transfer in v1) and an empty `Store`; a fresh wall (spec 032 US1) gets `HP = wallMaxHP(kind)` — full health, derived from kind, never stored as a separate max (fire lit-ness doctrine); a path gets no HP (not a wall — never blocks passage), intent cleared |
| `agent.wall_chipped` (spec 032 US1) | `WallWorkPayload{agent, x, y}` | work completion (`demolish`, when the chip would leave the wall standing, `HP − demolishChipHP ≥ 1`; research R5 — multi-cycle demolish) | the wall at `(x,y)`'s `HP -= demolishChipHP`, clamped to never go below 1 (a standing wall never serializes ≤ 0); the agent's `Intent.WorkStart` resets to 0, re-arming the executor's work gate for the next demolish cycle under the same intent — no new scheduling |
| `agent.wall_destroyed` (spec 032 US1) | `WallWorkPayload{agent, x, y}` | work completion (`demolish`, the final chip that would take `HP` to ≤ 0) | removes the wall structure at `(x,y)` — its tile becomes passable again by construction; intent cleared. `metatron.entity_removed` reaches the same end through the miracle path |
| `agent.wall_repaired` (spec 032 US1) | `WallWorkPayload{agent, x, y}` | work completion (`repair`; validity requires a damaged, still-standing wall plus 1 unit of its `wallRepairMaterial(kind)` carried) | consumes 1 unit of that material and restores `HP` by `repairHPPerUnit`, clamped to `wallMaxHP(kind)`; if still damaged AND material remains, `Intent.WorkStart` resets to 0 to re-arm another cycle (intent kept) — otherwise intent cleared |
| `agent.dropped` | `DroppedPayload{agent, x, y, kind, n}` | executor, `drop` completion (instant, agent's current tile — spec 013 US2, planner/plan-only) | `Inv[kind] −= n`; the tile's pile created-or-merged `+= n` (food becomes/merges a batch stamped `spoil_at = tick + rotWindowTicks`; spears AND axes (spec 032 US2) move most-worn-first with their durabilities), intent cleared |
| `agent.picked_up` | `PickedUpPayload{agent, x, y, kind, n}` | executor, `pick_up` completion (instant on arrival at a pile on/adjacent tile; one event per kind moved in the batch) | pile `−= n` (food oldest-batch-first), `Inv[kind] += n`; an emptied pile is removed; intent cleared on the last event of the batch |
| `agent.deposited` | `DepositedPayload{agent, x, y, kind, n}` | executor, `deposit` completion at a chest (instant on arrival — spec 013 US3) | `Inv[kind] −= n`, chest `Store[kind] += n`, both clamped to the chest's free space (`chestCap − bulk(*Store)`); intent cleared |
| `agent.withdrew` | `WithdrewPayload{agent, x, y, kind, n, owner}` | executor, `withdraw` completion at a chest (instant on arrival) | chest `Store[kind] −= n`, `Inv[kind] += n`, clamped to the taker's free bulk; intent cleared; a non-owner taker co-emits the theft companion batch (`social.chest_taken` + a reason-`"theft"` `social.relation_changed` + owner/witness `agent.memory_added`, all in the same batch — [[social-fabric]]) |
| `sim.food_rotted` | `FoodRottedPayload{x, y, kind, n}` | executor, per-game-minute rot sweep (spec 013 US5; same-kind spoiled batches merged per pile per sweep) | pile's food batches with `spoil_at ≤ tick` matching `kind` removed (up to `n`, oldest first); an emptied pile is removed; chest food never batches, so chests are never reached (FR-010) |
| `agent.cooked` | `CookedPayload{agent, station, consumed, produced, kind}` (station ∈ fire\|oven; kind ∈ food_cooked\|meals) | work completion (cook) | −FoodRaw(consumed), +kind(produced); an oven cook also −1 Wood, intent cleared |
| `agent.bathed` | `BathedPayload{agent, morale_after, warmth_after}` | work completion (bathe, oven only) | −1 Water, −1 Wood, Morale/Warmth set to the absolute post-cap values, intent cleared |
| `agent.refueled` | `RefueledPayload{agent, x, y, fuel_until}` | reflex/planner (instant on arrival) | −1 Wood, the fire at `(x,y)`'s `FuelUntil` set to the absolute (already-capped) deadline, intent cleared |
| `agent.spear_broke` | `SpearBrokePayload{agent}` | work completion (hunt, companion to `agent.hunted` in the same batch) | removes the now-zero `Spears[0]` entry |
| `agent.axe_broke` (spec 032 US2) | `AxeBrokePayload{agent}` | work completion (chop or quarry, companion to `agent.chopped`/`agent.quarried` in the same batch — the `agent.spear_broke` clone) | removes the now-zero `Axes[0]` entry |
| `sim.fire_burned_out` | `FireBurnedOutPayload{x, y}` | `stepEvents`, once per fuel-window transition (`tick-1 < FuelUntil <= tick`) | none — lit-ness stays derived from `FuelUntil`; chronicle/TUI material, plus a low-salience witness memory for nearby living agents |
| `agent.ate` | `AtePayload{agent, meals, cooked, raw, food_after}` | reflex/planner (instant), most-nutritious-first (Meals→FoodCooked→FoodRaw) to satiety (`eatOutcome`) | −Meals/−FoodCooked/−FoodRaw by the consumed counts, Food need set to the absolute `food_after` |
| `agent.slept` / `agent.woke` | `AgentPayload{agent}` | executor | sleep flag (slept clears intent) |
| `agent.needs_changed` | `NeedsPayload{agent, …}` | per-game-minute heartbeat | needs set to absolute values; spec 043 US2: once `tick − NeedsAnchorTick ≥ trajectoryWindowTicks` (1800) the arm snapshots the current needs into `Agent.NeedsAnchor`/`NeedsAnchorTick` — the window-edge anchor the decision prompt diffs against for rising/falling/steady arrows ([[decision-context]]); anchor unset (tick 0) until the first full window elapses |
| `agent.died` | `DiedPayload{agent, cause}` — cause ∈ `starvation`\|`exposure`\|`collapse`\|`gru` (spec 044 US3) | heartbeat at 0 health, **or** `gruStep` (spec 044 US3) on an escalated gru kill — emitted immediately after that attack's `gru.attacked`, cause `"gru"`, with an inline witness-death memory loop (gru attacks land off the %60 needs heartbeat, so the executor's own witness-death block never runs for them) | `Dead`, intent cleared; spec 013 (US2, FR-006, research R7): the agent's entire carried inventory spills into a pile at the death tile (created/merged, food batches stamped `tick + rotWindowTicks`), emptying `Inv` — reducer-internal, no new event; spec 044 US1: the death is also appended to the `State.Deaths` ledger (`{agent, tick, cause}`, application = event order) so the run-end declaration can carry the run's full death history without a log scan; spec 044 US4 (FR-017): a `Structure{Kind:"grave"}` is placed at the death tile — unconditional (no dedup against whatever else already occupies the tile; structures already coexist per-tile by kind) — visible on the map ([[tui-client]]), knowable to villagers ([[agent-mind]] known-places), and blocking future `buildSite` on that tile (deliberate, research R10) |
| `run.ended` (spec 044 US1) | `RunEndedPayload{tick, deaths, final_cause}` — `deaths` is the whole run's ledger (`{agent, tick, cause}`, event order) | executor (`stepEvents`), in the same batch as the run's final `agent.died`, ordered after every same-tick death and its witness memories, guarded by `!State.Ended` — exactly once per world, ever; never injectable (no whitelist entries) | sets `State.Ended` (terminal latch — no event clears it, so restart/replay lands back ended) and `State.RunEnd`; the loop idles in the paused-mode posture (mutating commands refused, reads served), status surfaces gain `ended`/`ended_day`, and the TUI enters postmortem posture ([[tui-client]], [[sim-loop]]) |
| `agent.talked` | `TalkedPayload{a, b}` | executor, adjacent pair (chat-while-working) | +morale both, talk cooldown; both remember |
| `agent.memory_added` | `MemoryAddedPayload{agent, text, salience, subject, tone, where?, why?, conv?, origin?}` | executor/social/governance/gru heuristics (situated by the acting-or-witnessing agent's tile, and — since spec 030 — stamped with the closed-vocabulary `Origin` class at that same emission site, a required constructor parameter so no new site can compile unstamped); convo gists (injected, `origin: gist`) | append to `Memories`; subject/tone mark gossip seeds; spec 019 (US1) copies the situated context verbatim onto the reduced `Memory` — `where` (`*MemoryPlace{x,y,desc}`, the tile plus a `describePlace`-baked terrain/feature phrase, nil = unsituated), `why` (the driving intent's `Reason`, verbatim; witness memories carry none), `conv` (a conversation ref = founding-talk tick, set by convo gists) — all `omitempty`, so a pre-019 payload reduces to a pre-019-shaped memory (baked at emission, never re-derived, [[agent-mind]], [[social-fabric]]); spec 030 copies `origin` the same way (`omitempty`, absent = `""` = secondhand) — the ONLY signal `DirectPerception` (`internal/sim/memory.go`), and so the belief validator, reads to classify a memory as direct perception (`action`/`witness`/`omen`) vs secondhand (`report`/`gist`/`digest`/absent) |
| `agent.memory_embedded` (spec 042 US1) | `MemoryEmbeddedPayload{agent, mem_seq, vec, model}` in `internal/sim/agents.go` | the mind-side embedder driver, injected via `InjectSocial` after observing the target's `agent.memory_added` commit ([[memory-retrieval]]) | scans the agent's memories newest-first for the one whose `Seq` equals `mem_seq` and copies `vec`/`model` onto it verbatim (the reducer never computes or inspects a vector); a `mem_seq` of 0 never matches (a pre-042, seq-less memory has no target identity); a vanished target (agent died, memory consolidated away) is a deliberate no-op, never an error |
| `agent.situation_embedded` (spec 042 US1) | `SituationEmbeddedPayload{agent, tick, text, vec, model}` in `internal/sim/agents.go` | the mind-side embedder driver, injected via `InjectSocial` at planner cadence ([[memory-retrieval]]) | overwrites `Agent.SitVec`/`SitVecModel`/`SitVecTick` with `vec`/`model`/`tick` — later events simply replace earlier ones, no history kept; `text` (the rendered situation template) rides the payload as an audit surface only, stored nowhere in state |
| `agent.thought` | `ThoughtPayload{agent, text, source}` | `inject_intent` command (planner); `inject_social` (musing) | none (chronicle material) |
| `journal.entry_written` | `JournalWrittenPayload{agent, text}` (`journal.go`) | mind journal tool (`write_journal_entry`, injected via `InjectSocial` — spec 019 US3) | the ONLY journal-growth path: appends a reducer-id'd `JournalEntry{id, tick, text}` to the agent's `Journal` via `appendEntry`, which enforces the per-agent `journalBudgetRunes` (4000) rune budget INSIDE `Apply` — the `InjectSocial` dry-run turns an over-budget append into a door rejection, so no over-budget event lands (SC-005, [[agent-mind]]) |
| `journal.entry_deleted` | `JournalDeletedPayload{agent, entry}` (`journal.go`) | mind journal tool (`delete_from_journal`, injected) | removes the entry with that id from the agent's `Journal` (survivor order preserved, ids never reused or renumbered so freed runes are immediately reclaimable); a missing id errors at the door |
| `daemon.started` / `daemon.stopped` | `DaemonStartedPayload` / `DaemonStoppedPayload` | daemon lifecycle | none |
| `daemon.llm_warning` (spec 034) | `LLMWarningPayload{provider, kind, detail, remedy?, active}` | daemon condition hook, via [[sim-loop]]'s `InjectOperator` door, on a provider-health raise/reclassify/clear ([[llm-provider-health]]) | none — operator-facing only, same no-op class as `daemon.started`/`stopped` |
| `social.*` family | see `specs/003-social-fabric/contracts/social-events.md` | executor rules, genesis, convo driver (injected) | edges, ledger, rumors, secrets; `social.conversation` appends the bounded record ring (TASK-22, [[social-fabric]]); `social.gave` (spec 013 US1) is additionally skipped by the executor when the receiver has zero free bulk and the reducer clamps defensively (never over `bulkCap`) |
| `social.chest_taken` | `ChestTakenPayload{owner, taker, x, y}` (`social.go`) | executor, same batch as a non-owner `agent.withdrew` (spec 013 US4, FR-011) | none beyond the record itself — the distinct taking happening; chronicle/TUI material ([[social-fabric]]) |
| consolidation family: `agent.memory_promoted` / `agent.memory_faded` / `agent.belief_revised` / `agent.narrative_set` / `agent.consolidated` | payload structs in `internal/sim/consolidate.go`; contract in `specs/004-nightly-consolidation/contracts/` (spec 030 additions in `specs/030-epistemic-hygiene/contracts/`) | consolidation driver (injected) | salience boost / memory removal / belief create-or-revise / narrative replace / once-per-night ledger ([[nightly-consolidation]]); all reducer-total (vanished targets no-op); spec 030 threads two payload additions through — `belief_revised`'s `evidence` (the validator's resolved `MemoryRef{tick, hash}` citations) and `direct` (whether any cited evidence is direct perception; only a `direct` revision refreshes the belief's `Reinforced` decay anchor — a myth retold nightly on hearsay alone never re-anchors), and `consolidated`'s `coerced` (telemetry: beliefs the validator downgraded off `"witnessed"` for lack of direct evidence, never a rejection) |
| `agent.belief_reinforced` (spec 030 US2, FR-008) | `BeliefReinforcedPayload{agent, belief_id}` in `internal/sim/consolidate.go` | whitelisted through `InjectSocial`'s injection door (the grounded-observation seam) — ships as consumer only; no in-tree emitter yet, the perception-of-absence work is the intended future producer | re-anchors the named belief's `Reinforced` decay-clock field to `now` (`e.Tick`); a vanished belief id no-ops, reducer-total like its siblings |
| `gru.emerged` / `gru.moved` / `gru.sighted` / `gru.attacked` / `gru.withdrew` | payload structs in `internal/sim/gru.go`; `GruAttackedPayload.Health` (spec 044 US3): >= `gruWoundFloor` when pre-attack health was >= `nearDeathBelow` (healthy targets never die from one attack); may be 0 when the target was already weakened (pre-attack health < `nearDeathBelow`) — immediately followed, same batch, by `agent.died{cause:"gru"}` | `gruStep` (executor tick) | `State.Gru` lifecycle/position; sighting latch; attack sets absolute post-wound health, wakes victim, clears intent ([[gru]]); reducer-total (vanished gru no-ops); a killing blow (health 0) additionally lands the standard `agent.died` fallout in the same batch |
| `chronicle.entry` | `ChronicleEntryPayload{day, from_tick, to_tick, text, thread, agents}` in `internal/sim/chronicle.go` | narrator driver (injected, TASK-11) | appends the bounded `State.Chronicle` ring ([[chronicle]]) |
| `metatron.charge_regenerated` | `ChargeRegeneratedPayload{}` in `internal/sim/metatron.go` | executor, absolute 6-game-hour boundaries below cap | `MetatronCharges` +1, cap 3 ([[metatron]]) |
| `metatron.nudged` | `MetatronNudgedPayload{form, targets, text}` | Metatron console turn (injected, TASK-12) | validates (charges > 0, form ∈ vision\|omen\|dream, living targets, text cap) then `MetatronCharges` −1; `vision` (spec 029, replaces `dream` as the live one-target form) needs exactly one living target at any hour; `omen` needs ≥1 living targets AND `State.Night`; `dream` is legacy-only (grandfathered exactly-one-target validation so historical events replay, but no tool can emit a new one); villager memories ride companion `agent.memory_added` events in the same atomic batch |
| `metatron.place_revealed` (spec 041 FR-014) | `PlaceRevealedPayload{agent, facts}` in `internal/sim/mentalmap.go` — the emitter bakes only the place identity (kind, x, y, prov `revealed`); `seen` and `detail` are the reducer arm's NORMATIVE stamps | `send_vision`'s optional place grant (`place_kind`/`place_x`/`place_y` argument triple, all-or-none), riding the vision's atomic `InjectSocial` batch after the nudge memories | validates (living target, every fact names a REAL place — `groundFactPresent`, so the god reveals what is, never what isn't) then upserts into the target's mental map with `Seen` = landing tick, provenance `revealed`, `Detail` = ground truth at landing (a fire's `FuelUntil`); the companion Origin-omen memory ("The vision showed you the fire at (x,y).") rides the same batch as `agent.memory_added`; chronicle grammar line; absorb rides the batch's `metatron.nudged` (paused-authoring trigger) |
| `metatron.order_placed` (spec 029, [[metatron-orders]]) | `MetatronOrder{id, origin, condition, action, event_types, agent, keywords?, confirm?, placed_tick, expires_tick, status}` | Metatron's `monitor_and_act` tool, injected via `InjectSocial` | validates (non-empty id not reused by any past order regardless of status, `origin` ∈ player\|system, non-empty `event_types`, ttl 1..7 game days, `agent` index valid or −1 for any, `condition` ≤300 chars, `action` ≤400 chars, and — player-origin only — fewer than 3 already-active player orders, `MetatronPlayerOrderCap`; system-origin deferral orders are exempt from the cap); the payload's `status` is ignored — a landed order is always `active`; `MetatronOrders` appended then pruned to every active order plus the most recent 32 non-active (`pruneMetatronOrders`) |
| `metatron.order_triggered` | `OrderTriggeredPayload{id, matched_type, matched_tick}` | the angel's trigger worker (injected, live-only — NEVER emitted during replay, since the matching runs off live events the replica sees post-batch) | the named order transitions active → triggered (one-shot consumption); rejects an unknown id or one not currently active |
| `metatron.order_cancelled` | `OrderIDPayload{id}` | Metatron's `cancel_order` tool, injected | the named order transitions active → cancelled; same rejection rule as triggered |
| `metatron.order_expired` | `OrderIDPayload{id}` | executor, `stepEvents`, once per order once `nextTick >= expires_tick` for an active order (the `charge_regenerated` pattern — a pure function of state + tick, so replay reproduces it without any angel running) | the named order transitions active → expired, freeing its slot against the player cap |
| `metatron.charter_observed` (spec 044 US2) | `CharterObservedPayload{fingerprint, default}` in `internal/sim/metatron.go` — `fingerprint` is a short content hash (12 hex chars) of the EFFECTIVE charter text a turn ran under; `default` marks the authored default | the Metatron turn pipeline (`observeCharter`, injected via `InjectSocial` at charter load), only when the fingerprint differs from `State.CharterFingerprint` — the first turn always emits; skipped on ended worlds | sets `State.CharterFingerprint`; the log's observation sequence is the event-sourced charter-revision timeline the morgue aligns each death against (most recent observation ≤ the death, [[metatron]]) |
| `morgue.epilogue` (spec 044 US2) | `MorgueEpiloguePayload{agent, text}` in `internal/sim/morgue.go` — `agent` is the mourned villager, or −1 for the run-end epilogue | mind narrator worker (injected via `InjectSocial` on absorbing `agent.died`/`run.ended`; LLM-gated, so structurally absent in no-model worlds); one of the two prose types an ENDED world's door still accepts (`endedProseWhitelist`) | appends the bounded `State.MorgueEpilogues` ring (chronicle pattern) for replica/scribe rendering; the morgue's factual render never depends on it — narrator absence or failure is a gap, never a stall |
| `metatron.time_snapped` | `TimeSnappedPayload{to_tick, gratis}` in `internal/sim/miracles.go` | angel's turn reply or the `promptworld miracle` CLI/IPC door (spec 016, [[metatron-miracles]]), injected via `InjectSocial` | rejects a target at or before the current tick (forward-only); spends 2 charges (the dearest miracle) unless `gratis`; `rebaseTicks` shifts every relative-duration field forward by the jump so remaining durations are preserved, then `State.Tick = to_tick`; the skipped regeneration boundaries mint no charges |
| `metatron.item_granted` | `ItemGrantedPayload{agent, kind, qty, gratis}` | angel's turn reply or the CLI/IPC door, injected | validates a living villager, a known item kind, positive qty, and the bulk cap (reject-whole, never clamp); spends 1 charge unless `gratis`; adds `qty` units (a spear grant appends `qty` fresh `spearDurability` entries, kept sorted) |
| `metatron.entity_moved` | `EntityMovedPayload{class, x, y, to_x, to_y, gratis}` (`class` ∈ villager\|structure\|pile) | angel's turn reply or the CLI/IPC door, injected | validates presence at the source and the destination's placement rule (villager/pile → passable, structure → buildSite); spends 1 charge unless `gratis`; relocates the entity (a moved villager drops its intent and goes idle at the landing tick; a moved structure carries its `FuelUntil`/`Owner`/`Store`; a moved pile merges onto any pile already at the destination) |
| `metatron.entity_removed` | `EntityRemovedPayload{class, x, y, gratis}` (`class` ∈ structure\|pile\|terrain; villager is rejected — never removable) | angel's turn reply or the CLI/IPC door, injected | validates presence; spends 1 charge unless `gratis`; deletes the structure (a chest first spills its `Store` to a ground pile — goods are never silently destroyed) or the pile (with contents), or overlays the terrain through the executor's own vocabulary (tree→`Cleared`, forage→`Harvested` with regrow, rock→`Quarried`; an already-overlaid tile is rejected as a no-op target) |
| `meeting.*` / `norm.*` families (TASK-13) | payload structs in `internal/sim/governance.go`; contract in `specs/006-norms-and-votes/contracts/governance-events.md` | all executor beats (`governanceEvents`) EXCEPT `meeting.proposal_rephrased`, the one injected governance type (mind phrasing driver), and a config-declared `meeting.convention_established`, seeded by the daemon on boot | meeting lifecycle on `State.Meeting`, norms enact/amend/repeal on `State.Norms`, reducer-internal voter/witness edge deltas; rephrase validates (norm exists, text ≤ 280) then swaps text only ([[governance]]) |
| `meeting.convention_established` (TASK-36) | `MeetingConventionPayload{convene_second, open_second, x, y, source}` in `internal/sim/governance.go` | executor emergent-gathering detector (`source: emergent`) or daemon boot seed from `world.json`'s `meeting` block (`source: config`) | one-shot: sets `State.MeetingConvention` (first source wins) and seeds `MeetingPlace`; clears the gathering watch ([[governance]]) |
| `sim.gathering_observed` (TASK-36) | `GatheringObservedPayload{x, y, start}` in `internal/sim/governance.go` | executor per-minute watch while no convention exists (start/break of a sustained gathering; all-zero = reset) | `Meeting.GatherStart/GatherX/GatherY` set, so replay reconstructs the emergent watch |
| `cog.thought` | `CogThoughtPayload{job, class, agent, snapshot_tick, generation, trigger_seq, points, predicted_wall_ms, predicted_land_tick, prompt_bytes?, block_bytes?, dropped_blocks?}` in `internal/sim/cognition.go` — the spec 043 tail (FR-009/FR-010) is additive, LAST, `omitempty`: assembled user-prompt size, per-block byte breakdown keyed by contract block name, and the blocks the size budget dropped in drop order; stamped ONLY on planner-class thoughts, zero-valued on every other class ([[decision-context]]) | mind driver (injected) when a call passes the router; `trigger_seq` is the log seq of the arming stimulus (0 = pure cadence) | none (telemetry, TASK-32, [[cognition]]) |
| `cog.outcome` | `CogOutcomePayload{job, class, agent, outcome, snapshot_tick, landing_tick, staleness_ticks, predicted_wall_ms, actual_wall_ms, kind?, reason?}` | loop landing ladder (landed/adapted/rejected-* /superseded) or mind driver (suppressed/expired/unusable — router suppressions have no matching `cog.thought`); the `retried` outcome is the one NON-terminal use — a marker for a consumed one-shot retry (conversation sites since TASK-42; the tool-loop's transport retry since spec 025, emitted by mind's `runPlan` and metatron's `Turn` alongside whatever terminal the run earns) | none — the terminal record of every thought (plus the non-terminal `retried` marker); rejections carry `kind` `prediction-miss` or `world-change` |
| `agent.intent_rejected` | `IntentRejectedPayload{agent, goal, reason, staleness_ticks}` in `internal/sim/cognition.go` | loop, when the landing ladder refuses a metered intent (alongside its `cog.outcome`) | since spec 043 US1 STATE-MUTATING (split out of the `cog.*` no-op arm): appends an already-closed `"rejected"` `IntentRecord` (source `planner`) to the ring — the refused intent never landed, so `Intent`/`IdleSince` stay untouched, but the next thought sees the attempt was made and refused; still its own type so the villagers tab/chronicle can notice refused intentions without parsing `cog.*` |
| `cog.recalibration_recommended` | `RecalibrationPayload{tier, estimate_s_per_pt, spike_rate, window}` in `internal/sim/cognition.go` | mind driver (injected) when a tier's live estimator breaches the spike-rate drift threshold (once per breach episode) | none (telemetry) |
| `cog.tool_call` (spec 017, FR-007) | `CogToolCallPayload{job, ordinal, tool, args?, verdict, reason?, tier, snapshot_tick}` in `internal/sim/cognition.go` | mind/metatron (injected), one per tool call a cognition's [[tool-loop]] saw — landed, rejected, read, or unlanded; `{job, ordinal}` is the correlation key (1-based, dense per job, model-emission order) | none — recorded observability, reducer no-op, whitelisted alongside the other `cog.*` types |
| `cog.memory_divergence` (spec 042 US2) | `MemoryDivergencePayload{agent, tick, mode, legacy, augmented, overlap, displacement, vectorless, sit_tick}` in `internal/sim/cognition.go` | mind driver (injected via `InjectSocial`), one per memory selection while the world's `memory_relevance` flag is `shadow` or `on` — `legacy`/`augmented` are both windows' memory `Seq`s in window order ([[memory-retrieval]]) | none — recorded observability, reducer no-op, whitelisted alongside the other `cog.*` types |
| `agent.plan_set` | `PlanSetPayload{agent, job, steps}` in `internal/sim/plan.go` | loop, on a guarded plan landing (TASK-32 US4) | `Agent.Plan` replaced with the steps |
| `agent.plan_step_started` / `agent.plan_expired` | `PlanStepPayload{agent, job, step, reason?}` in `internal/sim/plan.go` | executor (`planStepEvents`) on an idle agent's head step firing / window closing or resolve failing | head step popped / whole remaining plan cleared (a broken sequence is not resumed); spec 043: `plan_expired` also stamps the expired step into the `IntentLog` ring (`stampOrAppendExpired` — an open record matching the step's goal closes `"expired"`, otherwise a closed record is appended for a step that expired before ever firing) |
| hail family (TASK-47): `social.hailed` / `social.hail_met` / `social.hail_expired` | `HailedPayload{from, to, until}` / `HailMetPayload{from, to}` / `HailExpiredPayload{from, to}` in `internal/sim/agents.go`; contract in `specs/010-hail-protocol/contracts/events.md` | loop (`inject_intent` talk_to landing) and executor (`planStepEvents` talk_to firing) emit `hailed`; the executor's per-tick `hailStep` sweep emits `met` (hailer adjacent, accompanied by the `agent.talked` talk shape bypassing the ambient cooldown) or `expired` (window closed) | `hailed` sets `Agent.Hail{By, Until}` (the movement-only pause); `met`/`expired` clear it — `agent.died` and `agent.slept` also clear it. World-emitted only, never model-injectable |

Conventions: `clock.*` are applied player/scheduler commands; `sim.*`, `agent.*`,
and (spec 044) `run.*` are world happenings (pure functions of state + seed +
tick — `run.ended` is executor-emitted, never injectable); `morgue.*` (spec 044)
is injected narrator prose about the run, the chronicle's pattern; `daemon.*` are process
bookkeeping, wall-time dependent, and excluded from determinism comparisons (as are
`clock.*` in the binary-level test, since their ticks depend on command timing).
Payloads record **outcomes** (positions reached, absolute need values), never dice
rolls, so replay needs no RNG. Unknown types are no-ops in the reducer, so adding
types is backward-compatible with old replay code. The `cog.*` family
(TASK-32, [[cognition]]) is recorded observability —
explicit reducer no-ops whose wall-time fields are recorded input, so no failure
is silent and thought chains are walkable from the log alone; their payload
field order is canonical per `specs/007-cognition-horizon/contracts/events.md`.
`agent.intent_rejected` shares that observability role but since spec 043 is
no longer a no-op — its ring append (deterministic from the event alone) keeps
it replay-safe.
`world.migrated` (spec 012 US6) is the one exception to "payloads are small
outcomes" — its payload embeds the entire canonical `sim.State`, by design: it is
the single record standing in for the whole pre-break history, and the reducer's
`state.Seed` check keeps it total (a mismatched payload no-ops rather than erroring).

## Connections

[[sim-state-reducer]] applies these; the [[executor]], [[reflex-policy]], and
[[sim-loop]] emit the sim/agent/clock families — since spec 039,
[[daemon-lifecycle]]'s teaching-posture boot default is a second caller of the
loop's `set_speed` command that lands `clock.speed_set`, alongside the player;
`promptworld migrate`
([[cli-promptworld]], [[world-migration]]) emits `world.migrated`; the mind driver and the loop's
landing ladder emit the `cog.*` telemetry ([[cognition]]); [[daemon-lifecycle]]
emits `daemon.*`; [[event-log]] stores them;
[[ipc-protocol]] pushes them to subscribers verbatim. [[llm-provider-health]]
emits `daemon.llm_warning` through [[sim-loop]]'s `InjectOperator` door. The `metatron.*` miracle
family is emitted through [[metatron]]'s two doors and reduced in
`internal/sim/miracles.go` — see [[metatron-miracles]] for the cost table,
gratis doctrine, and the shift-semantics re-base taxonomy. The standing-order
lifecycle (spec 029) reduces in `internal/sim/metatron.go` alongside
`charge_regenerated`/`nudged` — see [[metatron-orders]] for the placement
validation, trigger-matching, and confirm/degradation mechanics; `order_placed`/
`order_triggered`/`order_cancelled` are whitelisted in [[sim-loop]]'s
`InjectSocial` door exactly like the miracle types, while `order_expired`
needs no whitelist entry (executor-emitted, the `charge_regenerated`
precedent). [[mental-maps]] owns the `agent.saw`/`agent.map_corrected`/
`social.place_told`/`metatron.place_revealed` family end to end — the
executor's perception sweep and talk sidecar emit the first three, the
`send_vision` door the fourth, and `internal/sim/mentalmap.go` reduces all
four. [[memory-retrieval]] owns `agent.memory_embedded`/
`agent.situation_embedded`/`cog.memory_divergence` end to end — the mind-side
embedder driver emits all three through [[sim-loop]]'s `InjectSocial` door,
and [[sim-state-reducer]] reduces the first two (the third is a `cog.*`
no-op). [[morgue]] owns the spec 044 surface — the [[executor]] emits
`run.ended`, the Metatron turn pipeline injects `metatron.charter_observed`
(reduced in `internal/sim/metatron.go`), and the mind narrator injects
`morgue.epilogue` (reduced in `internal/sim/morgue.go`); [[gru]] covers the
escalated-kill semantics on `gru.attacked`/`agent.died`. [[decision-context]]
covers the spec 043 ring/anchor surfaces the intent-lifecycle and needs rows
maintain, and the prompt whose size `cog.thought`'s `prompt_bytes` tail
records.

## Operational notes

The outcome-payload
convention ([[deterministic-rng]]) is load-bearing — keep it; `gru.attacked`
carrying absolute post-wound health (never the wound roll) is the pattern.
