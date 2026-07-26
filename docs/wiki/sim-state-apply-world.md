---
name: sim-state-apply-world
description: sim.State.Apply's world/governance dispatch arms — the unexported map/scenario fields, mental-map knowledge growth, gru/governance/miracle/guardian-order dispatch, curriculum unlocks, world-tuning, and world.migrated's wholesale replace
kind: component
sources:
  - internal/sim/state.go
  - internal/sim/agents.go
  - internal/sim/miracles.go
  - internal/sim/morgue.go
  - internal/sim/curriculum.go
verified_against: 510a3c3133e120d84cd50525dbc4ee0d3ec01cdc
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
bare `State` have none until called — so miracle reducer
arms consult the terrain vocabulary (`passable`/`buildSite`/`effectiveKind`)
identically live, in dry-run, and in replay. Spec 054 added a second such
field: an unexported, never-serialized
`scenario *armedScenario` (exercise definition + compiled incident
source), attached once at daemon boot via `ArmScenario` — a world with no
scenario block (nil here) is byte-identical to pre-054 on every path
([[scenario-machinery]]). `world.migrated`'s wholesale
`*s = p.State` replacement preserves the receiver's map AND armed
scenario (the unmarshalled payload state carries neither). `State.MapDims()`
(spec 041) exposes the attached map's `(W, H)` — 0,0 when unattached — so the
mind's prompt renderer ([[agent-mind]]) can size a mental-map bitmap read
without `State` ever serializing the map.

[[mental-maps]]'s two knowledge-growth arms mutate `Agent.Map` directly:
`agent.saw` upserts the perception sweep's fully-baked facts verbatim
(`Map.upsertFact`), `agent.map_corrected` removes facts the sweep found gone
(`Map.removeFact`) — both no-op on a map-less agent (a pre-041 world mid-
migration), keeping the reducer total; `social.place_told` (the talk
sidecar's directions exchange) and `metatron.place_revealed` (a vision's
optional place grant) route through the `applySocial`/`applyGuardian`
dispatchers below, upserting into the RECEIVER's map only where the fact is
absent or its own knowledge staler. Several
EXISTING arms gained silent DERIVED bookkeeping with no new event: `agent.moved`,
`agent.woke`, and a `villager`-class `metatron.entity_moved` each call
`markExplored`/`notePresence` — a mover's surroundings become explored
terrain and mover-and-bystanders record each other's positions — a pure
function of (state, event) with no chronicle noise, so a mind-map-populating
step never needs its own event type (research D2);
the `gru.*` family dispatches to
`applyGru` in `gru.go` ([[gru]]); the spec-077 incident kinds dispatch
likewise — `sim.cold_snap`/`sim.forage_blighted` to `applyIncident` in
`scenario.go` (latch `ColdSnapUntil`; append `Harvest` overlays,
idempotent on re-apply) and the `stranger.*` family to `applyStranger` in
`stranger.go` (entity lifecycle + the clamped, reducer-total take arm —
[[event-types-scenario-incidents]]);
the `meeting.*`/`norm.*` families — plus `meeting.convention_established` and
the `sim.gathering_observed` watch event (TASK-36) — dispatch to
`applyGovernance` in `governance.go` ([[governance]]); the four miracle types
`metatron.time_snapped`/`metatron.item_granted`/`metatron.entity_moved`/
`metatron.entity_removed` (spec 016, [[guardian-miracles]]) dispatch to
`applyMiracle` in `miracles.go`, alongside `metatron.charge_regenerated`/
`metatron.nudged`'s `applyGuardian` — which since spec 029 also arms the
standing-order lifecycle: `metatron.order_placed` validates and appends (id
uniqueness, origin, non-empty `event_types`, a 1..7-game-day ttl, valid agent
index, condition/action length caps, and — player-origin only — the 3-order
active cap) then prunes to the active set plus the most recent 32 non-active;
`metatron.order_triggered`/`metatron.order_cancelled`/`metatron.order_expired`
each transition one order from active to a terminal status via shared
`transitionGuardianOrder`, rejecting an unknown id or one not active
([[guardian-orders]]); since spec 044 (US2) `applyGuardian` also carries
`metatron.charter_observed`, validating a non-empty fingerprint (so the
`InjectSocial` dry-run refuses a blank one at the door) then sets
`State.CharterFingerprint` — state keeps only the CURRENT fingerprint, the
full revision timeline being the log's observation sequence the [[morgue]]
aligns each death against. `morgue.epilogue` dispatches to
`applyMorgueEpilogue` in `morgue.go` (spec 044 US2): validate the agent
index (`-1` = run-end epilogue) and non-empty text, then append the
bounded `State.MorgueEpilogues` ring (`morgueEpilogueCap` 32).
`guardian.report_card` (spec 063, [[grounded-feedback]]) dispatches to
`applyReportCard` in `reportcard.go`: validate-not-clamp —
non-empty fingerprint, non-empty note capped at 1200 runes, and
every cited seq strictly less than the event's own seq (a card can never
cite the future) — then keeps only the LATEST card on
`State.GuardianReportCard`; the log alone carries every prior card.
The `curriculum.*` pair (spec 046, [[curriculum-ladder]]) dispatches to
`applyCurriculum` in `curriculum.go` — validate-not-clamp, the guardian arm's
contract, both types being the executor emission class (pure functions of
recorded state: a landed event always re-applies cleanly in replay, a
malformed fixture is rejected at the door): `curriculum.exercise_passed`
checks a non-empty exercise id and the closed stage vocabulary
(`validLadderStage`, `stage-1`..`stage-4` — the reducer-side twin of
`world.ValidStage`, kept local so the deterministic core never imports the
save-directory package) then appends the bounded pass ring;
`curriculum.stage_unlocked` also rejects `stage-1` (the ladder's
unearned floor — only stages 2..4 ever unlock) and any stage already latched
(once per world per stage), and does NOT cross-check
`CurriculumPasses` — that ring is pruned past 32, so the gate-conjunct
evaluation (`EvaluateUnlock`) happens at emission time, never on re-apply.
`sim.tuning_applied` (spec 048, [[world-tuning]]) joins this
validate-not-clamp family: the payload is always the FULL effective five-dial
set (never a delta, never re-clamped here — clamping is `ParseTuning`'s job
daemon-side), so the arm is a pure, idempotent `s.Tuning = &TuningState{...}`
assignment — replay re-applies it identically and the daemon boot seed never
double-counts. `State.Tuning *TuningState` (`omitempty`, no `format_version`
bump) is nil until the first such event; nil reads as the default dial set
through the nil-safe accessors (`RefuelDyingBelow()`, `FireBurnPerWood()`,
`GruEmergePerMille()`, `PlannerCadence()`, `EncounterCooldown()`) every other
promoted call site (`agents.go`'s fire-fuel arm ([[sim-state-apply-agents]]), [[reflex-policy]],
[[gru]], [[agent-mind]]'s cadence/encounter scheduling) reads through instead
of the retired raw constants.
`world.migrated` (spec 012 US6) is the one arm not mutating fields
incrementally: after checking the payload's `State.Seed` matches (mismatch
no-ops, keeping `Apply` total), it replaces `*s` wholesale with the embedded state —
[[world-migration]] is the only producer. `world.forked` (spec 076,
[[world-forking]]): an explicit no-op arm — provenance stays off `State`.

## Connections

Back to [[sim-state-reducer]] and its other five split-off notes.
[[mental-maps]] owns `Agent.Map`'s type and the perception sweep;
[[gru]], [[governance]], [[guardian-miracles]], [[guardian-orders]], and
[[grounded-feedback]] own the deep mechanics their dispatch targets
implement; [[morgue]] owns `applyMorgueEpilogue`; [[curriculum-ladder]]
owns `EvaluateUnlock` and the unlock evidence constructor; [[world-tuning]]
owns the manifest and dial defaults; [[scenario-machinery]] owns the
unexported `scenario` field's lifecycle; [[worldmap-generation]] is what
[[world-migration]]'s `world.migrated` payload replaces state around.
