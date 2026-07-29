---
name: scenario-rubric
description: Split from [[scenario-machinery]] — the generalized rubric emitter (scenarioRubricEvents, spec 077 FR-003): per-exercise boundary dawns, EvaluateRubric production arms for every cataloged exercise, sanctioned-constructor evidence assembly (rubricEvidence), ExerciseOutcome derivation, and the nine-exercise catalog (seeds 46101–46109). Read when touching rubric arms, evidence constructors, or the exercise catalog in internal/sim/scenario.go / curriculum.go.
kind: component
sources:
  - internal/sim/scenario.go
  - internal/sim/curriculum.go
  - internal/sim/guardian.go
  - internal/sim/state.go
verified_against: 1603d5ac22d9be35469ec88bf2355b7d2f9500bc
---

# Scenario rubric (pass boundary, evidence, exercise catalog)

Split from [[scenario-machinery]] (corpus-spec v2 size-budget split,
summary-style): this note covers the rubric emitter — the evaluator watching
the replica for an exercise's pass boundary and landing
`curriculum.exercise_passed`/`stage_unlocked` — plus its evidence assembly,
the outcome derivation, and the nine-exercise catalog. The incident-scheduler
half (arming, the incident source, incident emission) stays in
[[scenario-machinery]]; both halves share that note's EXECUTOR EMISSION CLASS
purity contract (pure functions of state, boot-frozen scenario config, and
tick — no LLM, no injection door; the recorded events are the only latches).

## How it works

**The emitter** (`scenarioRubricEvents`, called from [[executor]]'s
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

**The arms**: `EvaluateRubric(s, def, tick) []RubricTerm{Label, Event, Met,
Count}` is the shared derivation this emitter's pass precondition, the
exercise panel's live gauges, AND every report-card surface (spec 072's
shared resolver, [[report-card-renderer]]) read. EVERY cataloged exercise
carries a production arm (spec 077 FR-002, sweep-tested — no cataloged id
reaches the default pending arm): `firstNightRubric`, `theLawRubric` (law
term over the adopted-ever `State.Norms` ledger; charter term = the shared
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
posture.

**Outcome**: failure is never emitted — `ExerciseOutcome(s, exercise)`
derives `in_progress` / `passed` / `failed` purely from replica facts
(`hasCurriculumPass` vs `s.Ended`), shared by the loop's status composer
and the exercise panel's banner, so every surface reports the same word.

## Connections

[[scenario-machinery]] is the parent — arming, the incident source, and
incident emission; [[curriculum-ladder]] owns the `ExerciseDefinition`
content and the `EvaluateUnlock` gate called here; [[executor]] hosts the
`stepEvents` call site; [[report-card-renderer]] reads the same
`EvaluateRubric` derivation; [[event-types]] catalogs
`curriculum.exercise_passed`/`stage_unlocked` and `GuardianOrder.PlacedSeq`;
[[guardian-orders]] owns `PlacedSeq`'s stamping (`stampSeqs`) that
`OrderPlacedEvidence` re-locates; [[scenario-machinery-surfacing]] renders
the terms as the exercise tab's gauges.

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
