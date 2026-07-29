---
name: scenario-machinery
description: The spec-054 scenario incident-schedule core, grown by spec 077 — ArmScenario compiling an authored Schedule into the boot-frozen unserialized runtime, and the director-lite incident source (scheduleSource) landing four incident kinds (gru_emerges, cold_snap, forage_blight, stranger_arrives) as reducer-valid unmarked events. The rubric emitter/evidence/exercise-catalog half is split to [[scenario-rubric]]; surfacing/wiring to [[scenario-machinery-surfacing]].
kind: component
sources:
  - internal/sim/scenario.go
  - internal/sim/curriculum.go
  - internal/sim/executor.go
  - internal/sim/gru.go
  - internal/sim/state.go
verified_against: 1603d5ac22d9be35469ec88bf2355b7d2f9500bc
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
This note covers the incident half; the rubric emitter, its evidence
assembly, and the nine-exercise catalog are split to [[scenario-rubric]].
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
nowhere else. The closed incident vocabulary is FOUR kinds (spec 077 FR-009
grew spec 054's one): `IncidentGruEmerges`, `IncidentColdSnap` (entry param
`Hours` [1,24]), `IncidentForageBlight` (`X,Y` center + `Radius` [1,8]),
`IncidentStrangerArrives` (`X,Y` entry tile) — `IncidentScheduleEntry`
carries the two additive kind-specific fields. `compileIncident` resolves
each to an absolute tick (`clock.ParseTimeOfDay`/`clock.TickAt`) plus a
per-kind `windowEnd`: the next dawn for the night-shaped kinds
(gru/blight/stranger), the snap's OWN end (`tick + Hours*3600`) for
cold_snap. A state latch per kind (`s.Gru != nil` / `coldSnapActive` /
blightable-tiles-empty / `s.Stranger != nil`) plus the closed window keeps
"fires late, never twice" pure — no internal mutable flag exists, so a
time-snap past the whole window skips the incident silently (US2 AS-2),
never retried retroactively.

**Emission — incidents** (`scenarioIncidentEvents`, called from
[[executor]]'s `stepEvents` immediately BEFORE `gruStep`, only when
`s.scenario != nil`): asks the armed source what's due, validates each
incident's kind-specific preconditions against pre-tick state, and emits
reducer-valid event shapes carrying NO authored/scenario marker (spec 077
FR-013) — a scheduled emission is indistinguishable in artifact from an
ambient cause. Per kind: `gru_emerges` (no gru abroad, position passable +
unprotected) emits `gru.emerged`; `cold_snap` (`coldSnapActive(s, tick)`
false) emits `sim.cold_snap{night, until_tick}` with the end derived from
the AUTHORED coordinates (tick + hours — a late firing still expires on
schedule); `forage_blight` (`blightableTiles(m, s, x, y, r)` non-empty —
an exhausted patch skips silently) emits ONE merged
`sim.forage_blighted{x, y, radius, tiles, regrow_tick}` with tiles in
deterministic row-major walk order and `regrow_tick = tick +
blightRegrowTicks` (4 game days); `stranger_arrives` (`strangerEntryValid`:
passable + unprotected, no stranger abroad) emits `stranger.arrived` — the
entity itself then lives in `stranger.go` ([[executor-tick-subsystems]]).
The preconditions are NAMED predicates written to be called verbatim by a
future ambient emitter — TASK-28's recorded seam, documented beside
`gruScheduledTonight` ([[gru]]'s `gru.go`), which remains the only
preemption twin because `gru_emerges` is the only kind with an ambient dice
path today: on a night with a scheduled `gru_emerges` entry the 22:00
random roll is skipped entirely — never two spawn mechanisms in one night —
and the skip consumes no RNG draw (`rngAt` is coordinate-seeded, no
stream), so ambient nights and unscheduled scenario nights roll exactly as
before. The two `sim.*` incident kinds' reducer arms live here too
(`applyIncident`): `sim.cold_snap` latches `State.ColdSnapUntil` (read-time
expiry — no end event; the needs heartbeat reads `coldSnapActive` for the
harsher `warmthLossColdSnap` rate, [[executor-needs-survival]]);
`sim.forage_blighted` appends the EXISTING `Harvest{X, Y, Regrow}` overlay
per tile (idempotent on re-apply — already-harvested tiles skip), so a
blighted tile IS a harvested tile with a long regrow: perception,
mental-map correction, and `sim.forage_regrown` all work unchanged.

**Emission — the rubric** ([[scenario-rubric]], the split-off second half):
`scenarioRubricEvents`, called from `stepEvents` immediately AFTER every
emitter and BEFORE run-end detection, watches the same replica for the
exercise's boundary dawn and emits `curriculum.exercise_passed` (plus
`curriculum.stage_unlocked` when `sim.EvaluateUnlock` grants) once every
`EvaluateRubric` term is satisfied — pass-only emission, sanctioned-
constructor evidence, `ExerciseOutcome` as the shared outcome word. The
rubric arms, evidence rules, and the nine-exercise catalog live in the
child note.

**Incident visibility** (reorientation D4): `IncidentVisibilityFor(def,
stage)` resolves the forecast/fog vocabulary — detail in
[[scenario-machinery-surfacing]].

**Surfacing, wiring, and the exercise tab** — split into
[[scenario-machinery-surfacing]]: once armed, the exercise id rides
status/CLI, the manifest, boot wiring, chronicle/morgue narration, and a
fifth TUI dock tab present only on a scenario world.

## Connections

[[curriculum-ladder]] owns the `ExerciseDefinition` content consumed here
(`ScenarioExercises`, `Schedule`, `IncidentVisibility`); [[scenario-rubric]]
is the split-off rubric half (arms, evidence, catalog); [[executor]] hosts
both emission call sites inside `stepEvents`; [[gru]] shares the
night-emergence preemption (`gruScheduledTonight`); [[event-types]] catalogs
the `curriculum.*` and `sim.*` incident event shapes;
[[sim-state-reducer]] owns the unexported `State.scenario` field, sharing
the `State.m` precedent; [[scenario-machinery-surfacing]] is the split-off
child covering every downstream surface.

## Operational notes

The nine-exercise catalog (seeds 46101–46109), including each exercise's
schedule and boundary, lives in [[scenario-rubric]]'s operational notes,
alongside the schedule-compile and per-seed position-validity tests
(`TestScenarioSchedulesCompile`, `TestSchedulePositionsValidPerSeed`).
