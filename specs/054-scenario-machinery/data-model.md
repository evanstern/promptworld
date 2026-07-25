# Data Model: Scenario machinery (spec 054)

## Incident schedule entry (authored content, on ExerciseDefinition)

| Field | Type | Meaning |
|---|---|---|
| Kind | closed enum, v1: `gru_emerges` | which emission |
| Day, TimeOfDay | int, "HH:MM" | game time; compiled to absolute tick at arm time |
| X, Y | ints (kind-specific) | authored position for gru_emerges |

**Validation**: kinds outside the enum refuse at world.Open (catalog
validation); times parse via the existing clock arithmetic. The schedule is
compiled-in exercise content (like RubricTerms), not player data.

## Incident source (the seam)

`incidentsDue(s *State, nextTick int64) []incident` — pure; the recorded
event is the only latch (derived from state, e.g. gru-abroad checks), never
internal mutable flags. v1 impl: the compiled authored schedule. Documented
future impl: the live state-watching director.

## Rubric evaluation (derived, per tick)

Pure function (state, definition, nextTick) → {term satisfaction[], at
boundary?, pass?}. No persisted state: CurriculumPasses/StagesUnlocked (the
existing reducer latches) provide once-only emission; evidence payloads are
built by the sanctioned constructors (`CharterObservedEvidence` et al.).
Outcome vocabulary: `in_progress` → `passed` (exercise_passed on log) /
`failed` (run.ended with no pass). Failure emits nothing — run.ended is the
signal.

## Manifest.Scenario (consumed, was reserved)

`ScenarioConfig{Exercise string}` — validated at Open against
`sim.ScenarioExercises` ids; boot-frozen into the loop (SetStage
discipline). Absent block = ambient world, all machinery dormant.

## Status additions (wire, additive omitempty)

`scenario_exercise` (id), `scenario_outcome` (`in_progress|passed|failed`).
Absent on ambient worlds and old daemons; clients treat absence as
no-exercise.

## Exercise tab state (client)

| Field | Meaning |
|---|---|
| briefing dismissed | per attach; reset on reconnect |
| gauge projections | derived per frame from replica (event counts, state facts) + compiled definition |
| visibility mode | definition override else stage default (forecast stages 1–2/pre-ladder, fog 3+) |
| banner | from replica CurriculumPasses / runEnded() |

## Visibility vocabulary (D4)

`forecast | fog` — a per-exercise optional override + stage-keyed default;
extensible (never a boolean in any signature or wire field).
