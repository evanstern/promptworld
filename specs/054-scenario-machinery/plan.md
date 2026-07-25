# Implementation Plan: Scenario incident-schedule machinery (director-lite)

**Branch**: `054-scenario-machinery` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/054-scenario-machinery/spec.md`

## Summary

Arm the spec-046 curriculum substrate with its awaited production machinery,
all in the executor emission class (pure over state + boot-frozen config +
tick): (1) an incident scheduler compiled from the exercise definition's
authored schedule (v1 vocabulary: `gru_emerges`, preempting that night's
random roll), behind a one-method incident-source seam the post-v1 live
director can implement later; (2) a rubric evaluator that emits
`curriculum.exercise_passed` + same-batch `curriculum.stage_unlocked`
(existing unlock gate + evidence constructors); (3) `promptworld new
--scenario <id>` stamping the reserved `Manifest.Scenario` block + the
exercise's stage/seed/preset; (4) status carrying scenario facts additively;
(5) a scenario-worlds-only **exercise** dock tab (key `6`) implementing
`panels/exercise.md` as authored (attach-once briefing, live gauges,
forecast/fog vocabulary, pass/fail banner); (6) one extra narrator chapter
trigger at the pass/fail boundary + the failed-run exercise line in the
morgue's run summary.

## Technical Context

**Language/Version**: Go 1.24 (repo toolchain; no new deps)

**Primary Dependencies**: existing packages — `internal/sim` (executor + curriculum), `internal/world` (manifest), `internal/daemon` (boot wiring), `internal/mind` (narrator trigger), `internal/scribe` (morgue line), `internal/ipc` (status fields), `internal/tui` (exercise tab), `cmd/promptworld` (--scenario)

**Storage**: no new files; consumes the reserved `Manifest.Scenario` block (world.go:87-96); events are the persistence

**Testing**: `go test -race ./...`; determinism twins (two genesis runs byte-identical; replay equivalence — the sim-loop test idioms); rubric-evaluator table tests; ambient-regression (zero fixture changes); TUI render/briefing tests; e2e first-night compressed-clock pass

**Target Platform**: daemon + terminal client

**Project Type**: single Go module, cross-package (sim core + daemon + client)

**Performance Goals**: per-tick evaluation is O(schedule remaining + rubric terms) over already-in-memory state; no LLM calls added (the narrator trigger reuses the existing queue)

**Constraints**: executor purity (stepEvents must not mutate state; no new RNG streams — authored schedules are deterministic by construction); reducer remains the semantic authority (incidents propose, dry-run disposes); ambient worlds byte-identical; same-batch pass+unlock (daemon observer contract, internal/daemon/curriculum.go:44-57); same-PR design-doc amendments

**Scale/Scope**: est. 900–1,500 LOC delta incl. tests; 4+ design pages touched

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Artifact-Grounded Action** — PASS: spec dir `specs/054-*`; TASK-119 linked pre-implementation; standing resolutions cite the authored exercise page + spec-046 seam comments.
- **II. One Task, One PR** — PASS: `.worktrees/task-119`, one branch, one PR.
- **III. Gates Over Assertions** — PASS: determinism twins + replay tests + design gate + spec-bridge gate.
- **IV. Grounding Freshness** — PASS (planned): touches sources of `executor.md`, `curriculum-ladder.md`, `sim-loop.md`, `tui-client.md`, `cli-promptworld.md`, `morgue.md`, `chronicle.md` → wiki-update + player-docs in re-ground.
- **V. Model-Tiered Workflow** — PASS: planned on Fable 5; implementation on **Opus 4.8** — rubric: sim-loop/executor emission class + determinism doctrine (cross-package, doctrine-adjacent) per the runbook Lane 2 assignment and constitution Principle V senior-tier profile.

**Post-Phase-1 re-check**: PASS — one new interface (incident source, deliberately one-method), no new packages; the manifest block was pre-reserved. No Complexity Tracking entries.

## Project Structure

### Documentation (this feature)

```text
specs/054-scenario-machinery/
├── plan.md
├── research.md          # R1–R7: scheduler shape, rubric evaluation, preemption,
│                        #   boot wiring, status/tab, narrator/morgue hooks
├── data-model.md        # schedule entry, incident source seam, rubric eval, status fields
├── quickstart.md
├── contracts/
│   └── scenario-machinery.md  # determinism contract, event contracts, panel grammar
└── tasks.md
```

### Source Code (repository root)

```text
internal/sim/
├── scenario.go          # NEW: incident source seam + authored-schedule impl
│                        #   (compile game-times → ticks at boot); rubric evaluator
│                        #   (pure per-tick over state + definition); pass emission
├── scenario_test.go     # determinism twins, preemption, rubric tables
├── executor.go          # stepEvents consults the armed scenario (charge-regen idiom);
│                        #   gru emergence roll preempted on scheduled nights
├── curriculum.go        # (read-mostly) evidence constructors + EvaluateUnlock consumed
└── gru.go               # emergence-roll preemption hook (smallest honest seam)

internal/world/world.go  # Scenario block: validation at Open (exercise id ∈ catalog)
internal/daemon/         # boot arming from Manifest.Scenario (SetStage discipline)
internal/ipc/            # Status: exercise id + outcome (additive omitempty)
internal/mind/narrate.go # chapter trigger at pass/fail boundary (new switch cases)
internal/scribe/morgue.go# run-summary exercise-outcome line (failed scenarios)
cmd/promptworld/         # new --scenario flag (commands.go new FlagSet); catalog refusal
internal/tui/            # exercise tab (key 6, scenario worlds only): briefing,
│                        #   gauges, forecast/fog, banner; dock enum extension
└── (tests beside each)

docs/design/tui/
├── panels/exercise.md       # specified → shipped, real symbols, tab key 6
├── panels/dock.md           # conditional 5th tab row
├── patterns/keymap.md       # key 6 + briefing dismiss + parity notes
├── patterns/stage-defaults.md # visibility vocabulary now real; re-verify
└── re-pins on all touched pages
```

**Structure Decision**: the machinery concentrates in one new `internal/sim/
scenario.go` beside the substrate it consumes; every other package gets the
smallest honest hook. The incident-source seam is an interface so the live
director lands post-v1 without touching the executor again.

## Complexity Tracking

No constitution violations — table intentionally empty.
