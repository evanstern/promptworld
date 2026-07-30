---
name: sim-state-apply-agents
description: sim.State.Apply's core per-agent arms — genesis placement, clock/night/forage ticks, intent/movement/eating/talk/needs/death, and the spec-083 neglect anchor/latch arms. v2 crafting yields, v3 storage events, and the wall demolish/repair family split to [[sim-state-apply-agents-resources]].
kind: component
sources:
  - internal/sim/state.go
  - internal/sim/agents.go
verified_against: cf65debb44c1e17b54c0f3421d11e1e8cc28576c
---

# Sim state: core agent Apply arms

Split from [[sim-state-reducer]] (summary-style, corpus-spec v2): genesis
placement and the bulk of `Apply`'s per-agent event arms — the ones every
tick actually exercises: clock/night/forage ticks, intents, movement,
eating, talk, needs, death, the v2 crafting/gather yields, the v3 storage
events, and the wall demolish/repair family.

`NewState(seed, m)` is genesis: tick 0 (day 1 06:00), `DefaultSpeed` (4x), eight
named agents on distinct passable tiles via `genesisPlacement` ([[deterministic-rng]]),
with deliberately imperfect needs — day 1 must demand foraging, wood, and a fire
before dark. `genesisPlacement` (spec 012 US6) is factored out so [[world-migration]]
can re-place carried souls on a regenerated v2 map byte-identically to a fresh
genesis of the same seed.

`Apply` switches on event type: `clock.*` maintain pause/speed/degradation —
since spec 028, `clock.speed_set` additionally clears `RequestedSpeed` (a
player command always collapses governed state, FR-009), and two new arms,
`clock.governor_shed`/`clock.governor_recovered`, apply the daemon governor's
ceiling-preserving notch decisions: both set `Speed = to` and follow
`EffectiveRate` from it unless `Degraded`; shed additionally sets
`RequestedSpeed = requested`, recovered sets it only when `to != requested`
(clearing it once the climb reaches the ceiling) — see [[cognition]] for the
governor's decision logic and [[event-types]] for the payload shape;
`sim.night_started`/`sim.day_started` flip `Night` (waking is an explicit
`agent.woke`, never implicit); `sim.forage_regrown` clears a harvest overlay; the
`agent.*` family ([[event-types]]) drives intents (`agent.intent_set` carries a
storage goal's `Kind`/`Qty` onto the `Intent`, spec 013 R4 — since spec 064 R1
it also carries an OPTIONAL completion condition (`UntilNeed`/`UntilValue`)
onto the intent, but ONLY when `UntilNeed` names a valid closed-set need
(`isRecoveryNeed` — `warmth`/`rest`/`food`); a malformed or absent need leaves
both fields zero, the pre-064 arrive-and-done shape — and also stamps
`Agent.LastGoal`/`LastGoalTick` — spec 015 R1, `omitempty`, written here and
never cleared by any event, so the [[tui-client]] villagers tab can show an
idle villager's most recent objective from any snapshot; since spec 017 the
payload carries `Job` (`omitempty`), the tool-use loop's job id when a
planner-loop landing set it, and since spec 019 the payload's LAST field,
`Reason` (`omitempty`), the planner's free-text reason — copied onto
`Intent.Reason` so it survives to completion, where the executor bakes it into
the memory's `Why`; reflex/executor-authored intents carry neither `omitempty`
tail, so those emissions marshal byte-identically to before; since spec 043
US1 the arm also appends an `IntentRecord` to the agent's ring via
`appendIntent` — source verbatim, oldest dropped past `intentLogCap` into a
fresh backing array so canonical bytes never alias dead capacity; a previous
record still open stays open, so an override reads as open-then-new; and
since spec 083 the arm stamps the neglect detector's last-class-intent clock
— `needClassOf(goal)` non-empty (the `needClassGoals` dictionary beside the
goal registry, [[reflex-policy]]) stamps that need's `*Intent = tick` on
`Agent.Neglect`, source-agnostic on purpose: any scheduled class intent
proves the mind engaged, whatever the outcome
([[executor-needs-survival]])),
movement, work
products (inventory + overlays + structures), eating (`agent.ate`'s `AtePayload`
sets the absolute post-eat food need and decrements each carried food form by its
consumed count — no reducer-side arithmetic), sleep, talk (since spec 061
the `agent.talked` arm also upserts the unordered `PairTalks` ledger via
`recordPairTalk` — the pair-scoped companion to the per-agent `LastTalk`
write, both hail-founded and ambient talks updating the one record; the
conversation loop damper's hail founding gate reads it back via
`PairLastTalk`, [[social-fabric]]), needs (absolute
values; since spec 043 US2 the `agent.needs_changed` arm also rolls the
trajectory anchor — once `tick − NeedsAnchorTick ≥ trajectoryWindowTicks`
(1800, one planner cadence) it snapshots the current needs into
`NeedsAnchor`/`NeedsAnchorTick`, so direction is measured over roughly the
last window rather than instant-to-instant noise; on a fresh world the first
window carries no anchor and renders steady; and since spec 083 the same arm
maintains the neglect detector's band-entry anchors on `Agent.Neglect`
([[sim-state-agent-fields]]) — per need in food/warmth/rest, a value below
its spec-062 danger band with no anchor set stamps `*Since = tick` (the
downward crossing), and a value at/above the band clears the anchor AND the
episode's fired latch together (episode over, detector re-armed); a third
NEW arm, `sim.neglect_detected`, sets exactly the payload need's fired latch
(one injection per episode — the executor sweep's emission,
[[executor-needs-survival]]), nothing else), and death. The v2
resource/crafting events, the v3 storage economy, and the spec-032 wall
demolish/repair HP family split into
[[sim-state-apply-agents-resources]].

## Connections

Back to [[sim-state-reducer]] and its other split-off notes.
[[sim-state-apply-agents-resources]] is this note's own split-off child —
crafting/gather yields, storage, and the wall HP family.
[[deterministic-rng]] owns `genesisPlacement`; [[world-migration]] reuses it
to re-place migrated souls; [[executor]] owns the needs/intent/death
semantics these arms enact; [[event-types]] catalogs every payload shape
here. [[sim-state-intent-lifecycle]] builds on `agent.intent_set`'s
completion condition. `agent.talked`'s `talkMoraleBonus` and
`agent.needs_changed`'s classification constants (`nearDeathBelow`/
`nearDeathResetAt`, `trajectoryWindowTicks`, the neglect band) are audited
by [[sim-state-reducer]]'s spec-092 reducer-constants replay-hazard doctrine
(TASK-75).
