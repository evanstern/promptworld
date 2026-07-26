---
name: scenario-machinery
description: The spec-054 scenario incident-schedule + rubric core, grown by spec 077 — ArmScenario compiling an authored Schedule, the director-lite incident source landing four incident kinds (gru_emerges, cold_snap, forage_blight, stranger_arrives), and the generalized rubric emitter (per-exercise dawn boundaries, sanctioned-constructor evidence) for curriculum.exercise_passed/stage_unlocked across the nine-exercise catalog. Surfacing/wiring split to [[scenario-machinery-surfacing]].
kind: component
sources:
  - internal/sim/scenario.go
  - internal/sim/curriculum.go
  - internal/sim/executor.go
  - internal/sim/gru.go
  - internal/sim/guardian.go
  - internal/sim/state.go
verified_against: 4c66d240b2715706964f02cfd2396256c9957d8e
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

**Emission — the rubric** (`scenarioRubricEvents`, called from
`stepEvents` immediately AFTER every emitter and BEFORE run-end detection,
only when `s.scenario != nil`; GENERALIZED by spec 077 FR-003 — the
spec-072 first-night-only guard is retired): at the exercise's boundary
dawn (`boundaryDue(def, nextTick)` over `ExerciseDefinition.BoundaryDay` —
N > 0 evaluates at dawn of day N only, a miss emits nothing forever; 0 is
rolling, every dawn from day 2 until a pass lands; always the same
`sim.day_started` arithmetic), with no rubric-violating `agent.died` in
THIS batch (a same-tick death is not yet folded into `s`, so an all-dead
dawn is a fail, never a photo-finish pass — the guard applies to EVERY
exercise) and every `EvaluateRubric` term satisfied, emits
`curriculum.exercise_passed` plus, same batch, `curriculum.stage_unlocked`
when `sim.EvaluateUnlock` grants ([[curriculum-ladder]]) — pass first,
once-only via the state latch `hasCurriculumPass` (belt) ahead of the
reducer's own `StagesUnlocked` duplicate rejection (suspenders). Failure
emits NOTHING: `run.ended` is the sole fail signal (`sim.ExerciseOutcome`
below), never an event of its own.
`EvaluateRubric(s, def, tick) []RubricTerm{Label, Event, Met, Count}` is the
shared derivation this emitter's pass precondition, the exercise panel's
live gauges, AND every report-card surface (spec 072's shared resolver,
[[report-card-renderer]]) read. EVERY cataloged exercise carries a
production arm (spec 077 FR-002, sweep-tested — no cataloged id reaches the
default pending arm): `firstNightRubric`, `theLawRubric` (law term over the
adopted-ever `State.Norms` ledger; charter term = the shared
`charterInForce` helper over `CharterFingerprint`/`CharterCustom`), and
seven spec-077 arms (`coldDawnRubric` … `stewardsChargeRubric`) built from
shared pure helpers — `surviveToDawn`, `noDeaths`, `deathsByCause` (the
spec-044 `DeathRecord.Cause` vocabulary), `charterInForce`, `skillsInForce`
(over the `metatron.skills_observed`-persisted `SkillsFingerprint`),
`nothingTaken` (the zero-wanted `StrangerTakes`-ledger term — Met at
genesis, an empty ledger IS the claim), `storedFoodTotal` (chest stores +
pile batches), and `playerOrderSince` (earliest player order at/after a
tick — toolsmith's acted-under-it conjunct). An exercise without an arm
still renders its `RubricTerms` as unevaluated pending rows (the honest
default, kept for future non-evaluator content).
**Evidence** (spec 077 FR-004) assembles through the sanctioned
constructors ONLY, keyed by satisfied term type (`rubricEvidence`):
`metatron.order_placed` → `OrderPlacedEvidence` over the exercise's
qualifying order (`evidenceOrder`: `firstNightWatch` for the watch-shaped
exercises, `playerOrderSince(SkillsObservedTick)` for toolsmith);
`metatron.charter_observed` → `CharterEvidenceFromState` (reads the
`CharterObservedSeq/Tick` coordinates the charter arm now persists, plus
`Custom: CharterCustom` — spec 072 FR-009's blocker removed);
`metatron.skills_observed` → `SkillsObservedEvidence` (`Custom: true` by
construction — skill files bind only from stage-3 and only players author
them; the stage-3→4 gate's long-deferred evidence design). When a
satisfied charter/skills term's coordinates are not yet on state (a
pre-077 snapshot), the pass WAITS — honest degradation, self-healing on
the next observation; order evidence keeps the shipped skip-not-block
posture. Failure is never emitted: `ExerciseOutcome(s, exercise)` derives
`in_progress` / `passed` / `failed` purely from replica facts
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

Nine exercises ship (`sim.ScenarioExercises`, spec 077 FR-001 — 3/2/2/2 by
stage, seeds 46101–46109 unique): stage-1 `first-night` (46101, day-2
boundary, gru_emerges), `cold-dawn` (46103, day-2, cold_snap 8h),
`stranger-at-the-gate` (46104, day-2, stranger_arrives); stage-2 `the-law`
(46102, ROLLING — its pass emission is real since spec 077: evidence via
`CharterEvidenceFromState`, unlocking stage-3 without `--override`) and
`blighted-larder` (46105, day-4, forage_blight r4 — banked-food floor
`blightedLarderFoodFloor` pinned beside the definition); stage-3
`toolsmith` (46106, rolling — the skills-evidence exercise whose pass first
opens the stage-3→4 gate in production) and `fog-watch` (46107, day-3,
cold_snap + gru under fog visibility); stage-4 `long-winter` (46108, day-4,
all four kinds across three nights) and `stewards-charge` (46109, rolling —
law + charter + skill file + zero deaths; stage-4 passes graduate, no
unlock). `TestScenarioSchedulesCompile` pins every cataloged schedule
compiles at boot (a compile error here is a content bug, never a runtime
one); `TestSchedulePositionsValidPerSeed` pins every authored position
valid on its own seed's map.
