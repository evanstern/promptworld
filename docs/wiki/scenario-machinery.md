---
name: scenario-machinery
description: The spec-054 scenario incident-schedule + rubric machinery — the director-lite incident source that lands authored pressure (gru_emerges) into a scenario world, the rubric evaluator that is spec 046's awaited production emitter for curriculum.exercise_passed/stage_unlocked, the boot-frozen ArmScenario runtime, and the TUI exercise tab that surfaces both
kind: component
sources:
  - internal/sim/scenario.go
  - internal/sim/curriculum.go
  - internal/sim/executor.go
  - internal/sim/gru.go
  - internal/sim/guardian.go
  - internal/sim/state.go
  - internal/sim/loop.go
  - internal/world/world.go
  - internal/daemon/daemon.go
  - internal/scribe/scribe.go
  - internal/scribe/morgue.go
  - internal/mind/mind.go
  - internal/mind/narrate.go
  - internal/tui/exercise.go
  - internal/tui/tui.go
  - internal/tui/views.go
  - cmd/promptworld/commands.go
  - internal/ipc/protocol.go
  - internal/ipc/server.go
verified_against: aedcf52f680ed68910e185c3ccde44bd320517b6
---

# Scenario machinery (incident schedule + rubric)

Spec 054 (TASK-119) is spec 046's awaited production half: the curriculum
ladder's [[curriculum-ladder]] `ExerciseDefinition`s and their
`curriculum.exercise_passed`/`stage_unlocked` events existed since spec 046,
but nothing emitted them outside test fixtures until this feature — an
authored incident scheduler that lands deterministic pressure into a scenario
world, and a rubric evaluator that watches the same world's replica for the
exercise's pass boundary. Both are the EXECUTOR EMISSION CLASS (the
`metatron.order_expired`/`charge_regenerated` precedent): pure functions of
(state, boot-frozen scenario config, tick), no LLM, no injection door — the
recorded events are the only latches, so a restart or replay resumes exactly.

## How it works

**Arming** (`internal/sim/scenario.go`): `ArmScenario(def)` compiles an
`ExerciseDefinition`'s authored `Schedule` (`[]IncidentScheduleEntry{Kind,
Day, Time, X, Y}`) into absolute-tick `compiledIncident`s and attaches the
result as `armedScenario{def, source}` to `State.scenario` — unexported and
NEVER serialized, exactly the `State.m *worldmap.Map` precedent: canonical
state bytes are unchanged by arming, and replay needs no scenario runtime at
all (the recorded events are the persistence). Called exactly once, at
daemon boot (`armScenario` in `internal/daemon/daemon.go`, the `SetStage`
discipline), from the manifest's `Scenario` block; a world that never calls
it is byte-identical to pre-054 on every path. `world.migrated`'s wholesale
state replacement carries the scenario runtime across the swap the same way
it carries `m` ([[sim-state-reducer]]). `State.ScenarioExerciseID()` reports
the armed exercise id, `""` for an ambient world — the loop's status
composer and the daemon's `armScenario` return value both read it.

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
shared derivation both this emitter's pass precondition and the exercise
panel's live gauges read — v1 implements `first-night`'s rubric end-to-end
(FR-004: survive to dawn of day 2, zero deaths, a player-origin standing
order placed before night one's fall — `firstNightWatch` scans
`State.GuardianOrders` earliest-first by `(PlacedTick, PlacedSeq)`); every
other exercise renders its `RubricTerms` as unevaluated pending rows (content
work, not faked). Evidence rides `OrderPlacedEvidence(order)` — curriculum's
second sanctioned `EvidenceRef` constructor beside `CharterObservedEvidence`
— which re-locates the watch's placement event via the reducer-stamped
`GuardianOrder.PlacedSeq` (`(Seq, Tick)` coordinates, the `Memory.Seq`
precedent; [[guardian-orders]] owns the field and its stamping). Failure is
never emitted: `ExerciseOutcome(s, exercise)` derives `in_progress` /
`passed` / `failed` purely from replica facts (`hasCurriculumPass` vs
`s.Ended`) — shared by the loop's status composer and the exercise panel's
banner, so every surface reports the same word.

**Incident visibility** (reorientation D4): `IncidentVisibilityFor(def,
stage)` resolves `VisibilityForecast` ("the schedule is shown ahead of time")
or `VisibilityFog` ("incidents are revealed only as they happen") — a
definition's own override wins, else the stage-keyed default (forecast at
stages 1–2 and pre-ladder, fog from stage 3, `docs/design/tui/patterns/
stage-defaults.md`). A vocabulary in every signature, never a boolean.

**Surfacing — status/CLI** ([[sim-loop]], [[ipc-protocol]],
[[cli-promptworld]]): `sim.Status` gains additive `omitempty`
`ScenarioExercise`/`ScenarioOutcome`, composed inside the loop goroutine
(`status()`) so the pair is always coherent with `Tick` — absent for an
ambient world, so its status bytes are unchanged. `ipc.WorldStatus` mirrors
the same two fields (`ipc/protocol.go`), folded in by `ipc/server.go`'s
`statusData` straight from the loop's snapshot; `promptworld status`'s
`scenarioStatusLine` renders `exercise: <id> — <outcome>` (`failed` renders
`failed (run ended)`), empty for an ambient world or an old daemon.
`promptworld new --scenario <id>` (`cmd/promptworld/commands.go`) resolves
the id against `sim.ExerciseByID` (unknown ids refuse, listing the whole
`sim.ScenarioExercises` catalog), IMPLIES the exercise's stage and pins its
authored seed (an explicit `--stage`/`--seed` may only agree, never
override), rides the existing earned-stage gate unchanged (a scenario never
bypasses the earn gate), and stamps the manifest's `Scenario` block
set-after-create via `world.SetScenario` (the `SetStage` pattern — write-once,
no toggle command).

**Manifest consumption** ([[world-save-directory]]): `world.Manifest.Scenario`
(`*ScenarioConfig{Exercise}`) is no longer the spec-046 reserved/unconsumed
seam — `world.Open` now validates a present block against
`ValidScenarioExercise` (a local mirror of `sim.ScenarioExercises`' id set,
the `validLadderStage` twin-list precedent — the deterministic core and the
save-directory package deliberately don't import each other;
`TestScenarioVocabularyMirrorsSimCatalog` pins the two in sync), refusing an
unknown exercise id loudly rather than silently booting ambient.
`world.SetScenario(dir, exercise)` is the write-once stamp `promptworld new
--scenario` calls right after `Create`.

**Boot arming** ([[daemon-lifecycle]]): `daemon.armScenario(w, state)` reads
the manifest's `Scenario` block, resolves it via `sim.ExerciseByID` (a
catalog miss here is a real corruption — `world.Open` already validated it —
refused loudly, never silently booted ambient), calls `state.ArmScenario`,
and returns the armed exercise id. Called after `seedSurvivalWatches`, before
the sim loop exists, so no tick ever runs unarmed. When armed, the scribe
(`scr.SetScenario(exercise)`) and the mind (`md.SetScenario(exercise)`) each
receive the exercise id the same boot-frozen way, both before their
respective consumers start.

**Narration and the morgue** ([[chronicle]], [[morgue]]): `Scribe.
SetScenario(exercise)` installs the exercise id and re-renders the morgue
immediately, so an already-ended scenario world's run summary carries the
exercise line from the very first boot render on restart.
`writeRunSummary` (`internal/scribe/morgue.go`) gains a `scenarioExercise`
parameter: on a scenario world it appends "**The exercise**: `<id>` —
`<outcome>`. _Stated as evidence; the reader draws the lesson._" to the
run-end section, in the same no-blame evidence register as the rest of the
morgue — failure is stated, never scored. `Mind.SetScenario(exercise)`
(`internal/mind/mind.go`) arms the narrator's additional chapter trigger:
`chronicleNote` (`internal/mind/narrate.go`) gains a
`curriculum.exercise_passed` case ("The watcher's exercise — `<id>` — was
passed: the village made it through.") and a scenario-cadence trigger that
closes ONE additional chapter at the exercise's pass/fail boundary —
`curriculum.exercise_passed` always closes a chapter; `run.ended` closes one
only when `md.scenarioExercise != ""` — additive to the ambient day/night
cadence, which stays untouched (`exercise_passed` only ever lands on
scenario worlds, and the `run.ended` half is gated on the armed scenario), so
a sub-one-game-day scenario run still yields a narrated chapter carrying the
outcome.

**The exercise tab** (`internal/tui/exercise.go`, [[tui-client]]): a fifth
dock tab, `paneExercise` (key `6`), present ONLY on a scenario world
(`Model.exerciseID()` reads `m.w.Manifest.Scenario` plus a live
`sim.ExerciseByID` re-check — world-shaped, not stage-shaped; absent on every
ambient world, and the tab, its help row, and its footer hint all vanish with
it rather than rendering inert). An attach-time briefing (framing +
incident-visibility mode) shows once per attach, dismissed by any key while
it's visible (`exerciseBriefingShowing`, reset on reconnect); after dismissal
the body renders one gauge row per `sim.EvaluateRubric` term over the live
replica (met/pending marker, backing event count — the SAME pure function
the executor's pass precondition reads, so panel and emitter can never
disagree), an incident line under `VisibilityForecast` (omitted entirely
under fog, never blanked), and a pass/fail banner once `sim.ExerciseOutcome`
resolves (dual-sourced with the TUI's own `runEnded()` posture for a live
transition the replica snapshot hasn't folded yet).

## Connections

[[curriculum-ladder]] owns the `ExerciseDefinition` content this note's
machinery consumes (`ScenarioExercises`, `Schedule`, `IncidentVisibility`)
and the `EvaluateUnlock` gate conjuncts `scenarioRubricEvents` calls;
[[executor]] hosts both emission call sites inside `stepEvents`; [[gru]]
shares the night-emergence preemption (`gruScheduledTonight`); [[event-types]]
catalogs `curriculum.exercise_passed`/`stage_unlocked` and `GuardianOrder.
PlacedSeq`; [[guardian-orders]] owns `GuardianOrder.PlacedSeq`'s stamping
(the reducer arm, `stampSeqs`) that `OrderPlacedEvidence` re-locates;
[[sim-state-reducer]] owns the unexported `State.scenario` field, sharing the
`State.m` unserialized-boot-frozen-state precedent; [[sim-loop]] and
[[ipc-protocol]]/[[ipc-server]] carry the additive `ScenarioExercise`/
`ScenarioOutcome` status facts; [[world-save-directory]] validates and
carries the manifest's `Scenario` block; [[daemon-lifecycle]] arms the
runtime at boot and hands the exercise id to the scribe and mind;
[[chronicle]] narrates the pass event and the additional chapter trigger;
[[morgue]] names the exercise outcome in the run summary; [[cli-promptworld]]
fronts `new --scenario` and the status line; [[tui-client]] hosts the
exercise dock tab.

## Operational notes

Two exercises ship (`sim.ScenarioExercises`): `first-night` (stage-1, seed
46101) has a full production rubric evaluator and an authored `Schedule`
(one `gru_emerges` entry, night one at (44,0)); `the-law` (stage-2, seed
46102) has no production evaluator yet — its charter conjunct is not
state-derivable today (state retains only the fingerprint, not the `Default`
flag) — so its rubric renders as pending content work, never faked.
`TestScenarioSchedulesCompile` pins every cataloged schedule compiles at
boot (a compile error here is a content bug, never a runtime one).
