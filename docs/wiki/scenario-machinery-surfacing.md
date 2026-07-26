---
name: scenario-machinery-surfacing
description: How the scenario runtime surfaces once armed — boot-frozen ArmScenario wiring at daemon boot, the ipc/status/CLI ScenarioExercise/Outcome facts, the manifest's consumed Scenario block (nine exercise ids since spec 077), chronicle/morgue narration of the pass/fail boundary, and the TUI's fifth exercise dock tab with its four-kind incident forecast nouns.
kind: component
sources:
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
verified_against: b6a20eaa4da1073a69959a5aff69591d931103a9
---

# Scenario machinery — surfacing, wiring, and the exercise tab

Split off [[scenario-machinery]] (spec 071): everything downstream of an
armed scenario — how the manifest carries it, how the daemon boots it, how
status/CLI/TUI/narration surfaces it — as distinct from the incident/rubric
emission core that stays in the parent.

## How it works

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
`ValidScenarioExercise` (a local mirror of `sim.ScenarioExercises`' id set
— nine ids since spec 077's catalog wave,
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

**Incident visibility** (reorientation D4, `IncidentVisibilityFor(def,
stage)` in `internal/sim/scenario.go`): resolves `VisibilityForecast` ("the
schedule is shown ahead of time") or `VisibilityFog` ("incidents are
revealed only as they happen") — a definition's own override wins, else the
stage-keyed default (forecast at stages 1–2 and pre-ladder, fog from stage
3, `docs/design/tui/patterns/stage-defaults.md`). A vocabulary in every
signature, never a boolean; the tab below and the attach briefing are its
consumers.

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
the executor's pass precondition AND, since spec 072, every report-card
surface read ([[report-card-renderer]]), so panel, emitter, and cards can
never disagree; both cataloged exercises evaluate for real — the-law's
gauges flip on adopted norms and the persisted charter authorship flag,
[[scenario-machinery]]), an incident line under `VisibilityForecast`
(omitted entirely under fog, never blanked), and a pass/fail banner once
`sim.ExerciseOutcome` resolves (dual-sourced with the TUI's own `runEnded()`
posture for a live transition the replica snapshot hasn't folded yet).

## Connections

[[scenario-machinery]] is the parent — owns the incident/rubric emission
core this wiring surfaces; [[sim-loop]] and [[ipc-protocol]]/[[ipc-server]]
carry the additive `ScenarioExercise`/`ScenarioOutcome` status facts;
[[world-save-directory]] validates and carries the manifest's `Scenario`
block; [[daemon-lifecycle]] arms the runtime at boot and hands the exercise
id to the scribe and mind; [[chronicle]] narrates the pass event and the
additional chapter trigger; [[morgue]] names the exercise outcome in the run
summary; [[cli-promptworld]] fronts `new --scenario` and the status line;
[[tui-client]] hosts the exercise dock tab.
