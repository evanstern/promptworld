---
name: sim-state-apply-world
description: sim.State.Apply's world/governance dispatch arms — the unexported map/scenario fields, mental-map knowledge growth, gru/governance/miracle dispatch, the spec-084/085 designation/directive/faith/prophecy plan dispatch (applyPlan), and world.migrated's wholesale replace. Guardian-order/curriculum/tuning records split to [[sim-state-apply-guardian-records]].
kind: component
sources:
  - internal/sim/state.go
  - internal/sim/agents.go
  - internal/sim/miracles.go
verified_against: b35a7ffec46ba996741cdba4af9652fcfd163b32
---

# Sim state: world & governance dispatch arms

Split from [[sim-state-reducer]] (summary-style, corpus-spec v2): `State`'s
unexported, never-serialized map/scenario fields, plus the `Apply` arms
that dispatch into sibling files or apply world-scale, validate-not-clamp
events — mental-map knowledge growth through `world.migrated`'s wholesale
replace.

`State` also carries an unexported `m *worldmap.Map` (spec 016): the static
generated map, attached at construction, never serialized (canonical state
bytes unchanged). `SetMap` attaches it to a `State` built outside
`NewState` — the loop's dry-run probe and any replica unmarshalled into a
bare `State` have none until called — so miracle reducer arms consult the
terrain vocabulary (`passable`/`buildSite`/`effectiveKind`) identically
live, in dry-run, and in replay. Spec 054 added a second such field: an
unexported, never-serialized `scenario *armedScenario` (exercise definition
+ compiled incident source), attached once at daemon boot via
`ArmScenario` — a world with no scenario block (nil here) is byte-identical
to pre-054 ([[scenario-machinery]]). `world.migrated`'s wholesale `*s =
p.State` replacement preserves the receiver's map AND armed scenario (the
unmarshalled payload state carries neither). `State.MapDims()` (spec 041)
exposes the attached map's `(W, H)` — 0,0 when unattached — so the mind's
prompt renderer ([[agent-mind]]) can size a mental-map bitmap read without
`State` ever serializing the map.

[[mental-maps]]'s two knowledge-growth arms mutate `Agent.Map` directly:
`agent.saw` upserts the perception sweep's fully-baked facts verbatim
(`Map.upsertFact`), `agent.map_corrected` removes facts the sweep found gone
(`Map.removeFact`) — both no-op on a map-less agent (a pre-041 world
mid-migration), keeping the reducer total; the spec-097 `agent.place_observed`
arm records only the agent's `LastObs` dedup anchor
(`ObservationMark{x, y, kinds, tick}` — [[executor-perception-observation]]),
never touching map or beliefs; `social.place_told` (the talk sidecar's directions exchange) and
`guardian.place_revealed` (a vision's optional place grant) route through
the `applySocial`/`applyGuardian` dispatchers below — upserting into the
RECEIVER's map only where the fact is absent or its own knowledge staler.
Since spec 084 the seven `designation.*`/`directive.*` types dispatch to
`applyPlan` in `plans.go` likewise — validate-not-clamp arms, one-way
transition doors, the `designation.placed` arm's all-villager grant;
[[guardian-designations]]. Since spec 085 `faith.changed` dispatches to
`applyFaith` in `faith.go` (the clamping fold, the ONLY faith writer) and
the three `prophecy.*` types to `applyProphecy` in `prophecy.go` (the
declaration door spends the charge stake; terminals re-validate and
transition one-way); [[guardian-faith]]. Several EXISTING arms gained
silent DERIVED bookkeeping with no new event: `agent.moved`, `agent.woke`,
and a `villager`-class `guardian.entity_moved` each call
`markExplored`/`notePresence` — a mover's surroundings become explored
terrain and mover/bystanders record each other's positions — a pure
function of (state, event) with no chronicle noise, so map-populating never
needs its own event type (research D2). The `gru.*` family dispatches to
`applyGru` in `gru.go` ([[gru]]); the spec-077 incident kinds dispatch
likewise — `sim.cold_snap`/`sim.forage_blighted` to `applyIncident` in
`scenario.go` (latch `ColdSnapUntil`; append `Harvest` overlays, idempotent
on re-apply) and the `stranger.*` family to `applyStranger` in
`stranger.go` (entity lifecycle + the clamped, reducer-total take arm —
[[event-types-scenario-incidents]]); the `meeting.*`/`norm.*` families —
plus `meeting.convention_established` and the `sim.gathering_observed`
watch event (TASK-36) — dispatch to `applyGovernance` in `governance.go`
([[governance]]); the four miracle types
`guardian.time_snapped`/`guardian.item_granted`/`guardian.entity_moved`/
`guardian.entity_removed` (spec 016, [[guardian-miracles]]) dispatch to
`applyMiracle` in `miracles.go`, alongside `guardian.charge_regenerated`/
`guardian.nudged`'s `applyGuardian`. `applyGuardian` and its neighbors also
own the standing-order lifecycle, charter observation, morgue epilogues,
the guardian report card, curriculum unlocks, and the tuning-dial snapshot
— split into [[sim-state-apply-guardian-records]]. `world.migrated` (spec
012 US6) is the one arm not mutating fields incrementally: after checking
the payload's `State.Seed` matches (mismatch no-ops, keeping `Apply`
total), it replaces `*s` wholesale with the embedded state —
[[world-migration]] is the only producer. `world.forked` (spec 076,
[[world-forking]]): an explicit no-op arm — provenance stays off `State`.

## Connections

Back to [[sim-state-reducer]] and its other split-off notes.
[[sim-state-apply-guardian-records]] is this note's own split-off child —
standing orders, charter observation, morgue epilogues, the report card,
curriculum unlocks, and tuning. [[mental-maps]] owns `Agent.Map`'s type and
the perception sweep; [[gru]], [[governance]], and [[guardian-miracles]]
own the deep mechanics their dispatch targets implement;
[[scenario-machinery]] owns the unexported `scenario` field's lifecycle;
[[worldmap-generation]] is what [[world-migration]]'s `world.migrated`
payload replaces state around.
