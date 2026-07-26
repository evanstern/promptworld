---
name: scenario-machinery
description: The spec-054 scenario incident-schedule + rubric core — ArmScenario compiling an authored Schedule, the director-lite incident source landing gru_emerges pressure, and the rubric evaluator that is spec 046's awaited production emitter for curriculum.exercise_passed/stage_unlocked. Surfacing/wiring split to [[scenario-machinery-surfacing]].
kind: component
sources:
  - internal/sim/scenario.go
  - internal/sim/curriculum.go
  - internal/sim/executor.go
  - internal/sim/gru.go
  - internal/sim/guardian.go
  - internal/sim/state.go
verified_against: 8ec9aefc624396325c0083d2be207d5fcb057420
---

# Scenario machinery (incident schedule + rubric)

Spec 054 (TASK-119) is spec 046's awaited production half:
[[curriculum-ladder]]'s `ExerciseDefinition`s and their
`curriculum.exercise_passed`/`stage_unlocked` events existed since spec 046,
but nothing emitted them outside fixtures — this feature adds the authored
incident scheduler landing deterministic pressure into a scenario world,
plus the rubric evaluator watching the same replica for the exercise's
pass boundary. Both are the EXECUTOR EMISSION CLASS (the
`metatron.order_expired`/`charge_regenerated` precedent): pure functions of
(state, boot-frozen scenario config, tick), no LLM, no injection door — the
recorded events are the only latches, so a restart or replay resumes exactly.
Downstream surfacing (status/CLI, manifest, boot wiring, narration, TUI):
[[scenario-machinery-surfacing]].

## How it works

**Arming** (`internal/sim/scenario.go`): `ArmScenario(def)` compiles an
`ExerciseDefinition`'s authored `Schedule` (`[]IncidentScheduleEntry{Kind,
Day, Time, X, Y}`) into absolute-tick `compiledIncident`s and attaches the
result as `armedScenario{def, source}` to `State.scenario` — unexported and
NEVER serialized, exactly the `State.m *worldmap.Map` precedent: canonical
state bytes are unchanged by arming, and replay needs no scenario runtime at
all (the recorded events are the persistence). Called exactly once, at
daemon boot (the `SetStage` discipline, [[scenario-machinery-surfacing]]),
from the manifest's `Scenario` block; a world that never calls it is
byte-identical to pre-054 on every path. `world.migrated`'s wholesale state
replacement carries the scenario runtime across the swap the same way it
carries `m` ([[sim-state-reducer]]). `State.ScenarioExerciseID()` reports
the armed exercise id, `""` for an ambient world.

**The incident source** (`incidentSource` interface, `incidentsDue(s,
nextTick) []incident`): the director seam (contract §6). v1 ships exactly one
implementation, `scheduleSource` — the compiled authored schedule, carrying
no mutable state of any kind; a post-v1 live state-watching director is a
documented second implementation of the same interface, attaching here and
nowhere else. The closed v1 incident vocabulary is one kind,
`IncidentGruEmerges` ("gru_emerges"): compiled via `compileIncident` to an
absolute tick (`clock.ParseTimeOfDay`/`clock.TickAt`) plus a `windowEnd` —
the next dawn after that tick, since a gru emergence's window is exactly the
night it's authored for; a state latch (`s.Gru != nil`) plus the closed
window is what keeps "fires late, never twice" pure — no internal mutable
flag exists, so a time-snap past the whole window skips the incident
silently (US2 AS-2), never retried retroactively.

**Emission — incidents** (`scenarioIncidentEvents`, called from
[[executor]]'s `stepEvents` immediately BEFORE `gruStep`, only when
`s.scenario != nil`): asks the armed source what's due, validates each
incident's kind-specific preconditions against pre-tick state (no gru
abroad, the authored position passable and unprotected — the same class of
tile the random emergence path would itself draw from), and emits the
identical reducer-valid event shape (`gru.emerged`) the ambient dice path
uses — a scheduled emergence is indistinguishable in kind from a rolled one.
`gruScheduledTonight` ([[gru]]'s `gru.go`) is the companion preemption check:
on a night with a scheduled `gru_emerges` entry, [[gru]]'s own 22:00 random
roll is skipped entirely — never two spawn mechanisms in one night — and the
skip consumes no RNG draw (`rngAt` is coordinate-seeded, no stream), so
ambient nights and unscheduled scenario nights roll exactly as before.

**Emission — the rubric** (`scenarioRubricEvents`, called from
`stepEvents` immediately AFTER every emitter and BEFORE run-end detection,
only when `s.scenario != nil`): at the exercise's boundary tick (first-night:
dawn of day 2, the same `sim.day_started` arithmetic), with no
rubric-violating `agent.died` in THIS batch (a same-tick death is not yet
folded into `s`, so an all-dead dawn is a fail, never a photo-finish pass)
and every `EvaluateRubric` term satisfied, emits `curriculum.exercise_passed`
plus, same batch, `curriculum.stage_unlocked` when `sim.EvaluateUnlock`
grants ([[curriculum-ladder]]) — pass first, once-only via the state latch
`hasCurriculumPass` (belt) ahead of the reducer's own `StagesUnlocked`
duplicate rejection (suspenders). Failure emits NOTHING: `run.ended` is the
sole fail signal (`sim.ExerciseOutcome` below), never an event of its own.
`EvaluateRubric(s, def, tick) []RubricTerm{Label, Event, Met, Count}` is the
shared derivation this emitter's pass precondition, the exercise panel's
live gauges, AND every report-card surface (spec 072's shared resolver,
[[report-card-renderer]]) read. Both cataloged exercises carry production
arms: `first-night` (FR-004: survive to dawn of day 2, zero deaths, a
player-origin watch placed before night one's fall — `firstNightWatch`
scans `State.GuardianOrders` earliest-first) and `the-law` (spec 072
FR-007, `theLawRubric`: law term met on non-empty `State.Norms` — the
adopted-ever ledger `resolveProposal` appends — charter term on
`CharterFingerprint != "" && CharterCustom`, persisted as `!Default` by the
`metatron.charter_observed` arm; latest observation wins, so a revert to
the default charter flips it back off). An exercise without an arm
renders its `RubricTerms` as unevaluated pending rows (the honest default,
not faked). Emission stays first-night-only (spec 072 FR-009): the-law's
boundary tick, evidence assembly (`CharterObservedEvidence` needs the
observed event's `Seq`/`Tick`, which state does not retain), and pass
emission remain exercise-catalog content work. Evidence rides `OrderPlacedEvidence(order)` — curriculum's
second sanctioned `EvidenceRef` constructor — re-locating the watch's
placement event via the reducer-stamped `GuardianOrder.PlacedSeq`
([[guardian-orders]] owns the stamping). `ExerciseOutcome(s, exercise)`
derives `in_progress`/`passed`/`failed` purely from replica facts
(`hasCurriculumPass` vs `s.Ended`) — shared by the loop's status composer
and the exercise panel's banner, so every surface reports the same word.

**Incident visibility** (reorientation D4): `IncidentVisibilityFor(def,
stage)` resolves the forecast/fog vocabulary — detail in
[[scenario-machinery-surfacing]].

**Surfacing, wiring, and the exercise tab** — split into
[[scenario-machinery-surfacing]]: once armed, the exercise id rides
status/CLI, the manifest, boot wiring, chronicle/morgue narration, and a
fifth TUI dock tab present only on a scenario world.

## Connections

[[curriculum-ladder]] owns the `ExerciseDefinition` content consumed here
(`ScenarioExercises`, `Schedule`, `IncidentVisibility`) and the
`EvaluateUnlock` gate `scenarioRubricEvents` calls; [[executor]] hosts both
emission call sites inside `stepEvents`; [[gru]] shares the night-emergence
preemption (`gruScheduledTonight`); [[event-types]] catalogs
`curriculum.exercise_passed`/`stage_unlocked` and `GuardianOrder.PlacedSeq`;
[[guardian-orders]] owns `PlacedSeq`'s stamping (`stampSeqs`) that
`OrderPlacedEvidence` re-locates; [[sim-state-reducer]] owns the unexported
`State.scenario` field, sharing the `State.m` precedent;
[[scenario-machinery-surfacing]] is the split-off child covering every
downstream surface.

## Operational notes

Two exercises ship (`sim.ScenarioExercises`): `first-night` (stage-1, seed
46101) has a full production rubric evaluator, pass emission, and an
authored `Schedule` (one `gru_emerges` entry, night one at (44,0));
`the-law` (stage-2, seed 46102) has a production EVALUATOR since spec 072
(the charter-authorship blocker removed by persisting `State.CharterCustom`)
but no pass emission yet (FR-009, content work).
`TestScenarioSchedulesCompile` pins every cataloged schedule compiles at
boot (a compile error here is a content bug, never a runtime one).
