# Tasks: Scenario incident-schedule machinery (director-lite)

**Input**: Design documents from `/specs/054-scenario-machinery/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/scenario-machinery.md, quickstart.md
**Board**: TASK-119 · one branch (`task-119-scenario-machinery`), one PR

## Phase 1: Setup

- [x] T001 Verify baseline in the task worktree on fresh `origin/main`: `go build ./...`, `go test -race ./...`, `node scripts/check-tui-design.mjs --changed` green; note whether TASK-125's systems tab (key 5) has merged (dock-enum rebase expectation)

## Phase 2: Foundational — the sim machinery

- [x] T002 Incident schedule on the exercise definition (internal/sim/curriculum.go content + internal/sim/scenario.go): schedule entry type (kind enum v1 `gru_emerges`, day/HH:MM, position per data-model.md); FirstNightExercise gains its authored schedule; compile-to-ticks at arm time via existing clock arithmetic
- [x] T003 Incident source seam in internal/sim/scenario.go per research R2: `incidentsDue(s, nextTick)` pure, state-latched (recorded event is the latch); authored-schedule implementation; the director seam documented at the interface (contract §6)
- [x] T004 Executor integration (internal/sim/executor.go, charge-regen idiom): consult the armed scenario in stepEvents; gru_emerges emission with precondition checks; random-roll preemption on scheduled nights (internal/sim/gru.go smallest honest hook, research R3); determinism twin tests + preemption tests + ambient-regression (zero fixture changes) in internal/sim/scenario_test.go
- [x] T005 Rubric evaluator in internal/sim/scenario.go per research R4: pure per-tick term satisfaction + boundary detection for first-night (dawn-of-day-2, zero deaths, watch evidence, charter evidence via sanctioned constructors); emits exercise_passed + same-batch stage_unlocked via EvaluateUnlock; once-only via existing reducer latches; table tests + same-batch ordering test + replay-equivalence test
- [x] T006 Manifest + boot wiring: world.Open validates Scenario.Exercise against the catalog (internal/world/world.go, ValidStage idiom); daemon arms the loop from the manifest at boot (internal/daemon, SetStage discipline); armed-vs-ambient boot tests

## Phase 3: User Story 3 — promptworld new --scenario (P2, small; unblocks manual testing)

- [x] T007 [US3] `--scenario <id>` flag on the new command (cmd/promptworld/commands.go): stamp Scenario + definition Stage/Seed/preset; unknown-id refusal listing the catalog; earned-stage gate unchanged; CLI tests

## Phase 4: User Story 1+2 e2e (P1)

- [x] T008 [US1] Compressed-clock e2e: first-night world runs to pass — both events one batch, unlocks record updated (daemon observer), replay reproduces (SC-001/SC-003 backbone); and to fail — run.ended, no pass, outcome derivable
- [x] T009 [US2] Status additions (internal/ipc): scenario_exercise + scenario_outcome additive omitempty per research R6; server composition + protocol tests; model-free outcome via `promptworld status`

## Phase 5: User Story 4 — the exercise tab (P2)

- [x] T010 [US4] Dock tab `exercise` (key 6) present only when status carries an exercise id (internal/tui): enum/paneNames extension via existing dispatch; ambient worlds show no tab, `6` inert; tab-grammar tests
- [x] T011 [US4] Panel content per contract §4: title/banner (in progress/passed/failed from replica), gauge rows (term plain-language + met/pending + backing count, live), incident line per visibility vocabulary (definition override else stage default; forecast lists compiled schedule with approx times, fog omits); render tests across {stage-1 forecast, stage-3 fog, passed, failed, ambient-absent}
- [x] T012 [US4] Attach-time briefing: once per attach, framing + visibility mode, any-key dismiss consuming exactly one key while the exercise tab is visible; reset on reconnect; focus-contract regression tests

## Phase 6: User Story 5 — narration + morgue (P3)

- [x] T013 [US5] Narrator chapter trigger (internal/mind/narrate.go): exercise_passed case + scenario-world run.ended case call closeChapter at the boundary; ambient cadence byte-identical (regression test)
- [x] T014 [US5] Morgue run-summary exercise-outcome line on failed scenario runs (internal/scribe/morgue.go writeRunSummary), no-blame register; fold test

## Phase 7: Polish & Cross-Cutting Concerns

- [x] T015 [P] Design-page amendments: panels/exercise.md specified → shipped (real symbols, key 6); panels/dock.md (conditional 5th tab); patterns/keymap.md (key 6 + dismiss + parity notes); patterns/stage-defaults.md re-verified (visibility vocabulary real); re-pin every touched page to the final code commit
- [x] T016 Run gates: `go test -race ./...`, `node scripts/check-tui-design.mjs --changed`, gofmt/vet clean
- [x] T017 Pre-PR: rebase onto fresh `origin/main` (expect dock-enum/keymap conflicts with TASK-125's systems tab — take main's side for anything not intentionally changed; renumber the tab key if collision), re-run all gates post-rebase

## Dependencies & Execution Order

- Phase 2 strictly ordered T002 → T003 → T004 → T005 → T006 (each consumes the prior).
- T007 after T006; T008 after T005+T007; T009 after T005.
- US4: T010 after T009; T011 after T010; T012 after T011.
- US5: T013/T014 [P] after T005.
- Polish: T015 → T016 → T017.

## Implementation Strategy

Sim machinery first and fully tested (Phases 2–4) — it is the deliverable;
the tab (Phase 5) projects it; narration/morgue (Phase 6) polish it. One
worktree, one PR; the PR body calls out the determinism contract (§1) and
the director seam for reviewers.
