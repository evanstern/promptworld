# Tasks: Report-card truth — unify all card surfaces on sim.EvaluateRubric

**Input**: Design documents from `/specs/072-report-card-truth/`
**Prerequisites**: spec.md, plan.md

**Tests**: included alongside code (constitution V routine posture; tier is Opus 4.8 per
the board record — cross-package + doctrine-adjacent).

**Organization**: phases map to the board ACs — Phase 3 ↔ AC #1 (unified surfaces, ✗ on
failure), Phases 2+4 ↔ AC #2 (the-law evaluator, flag persisted, live gauges), Phase 5 ↔
AC #3 (design pages + check-tui-design). Phases 6–7 are the wiki-in-PR obligations and
close-out.

## Phase 1: Setup

- [ ] T001 Cut the task worktree from fresh origin/main and baseline: `node scripts/check-merge-drift.mjs worktree --spec 072 --task TASK-149`, then `git worktree add .worktrees/task-149 -b task-149-report-card-truth origin/main`; `go test ./internal/sim/ ./internal/tui/` green before changes (worktree `.worktrees/task-149`)

## Phase 2: Foundational — persist the charter authorship flag (blocks Phase 4)

- [ ] T002 Add `CharterCustom bool` (`json:"charter_custom,omitempty"`) beside `CharterFingerprint` with the spec-072 zero-value doc comment, in `internal/sim/state.go` (~line 121) (plan D1, spec FR-006/008)
- [ ] T003 Set `s.CharterCustom = !p.Default` in `applyGuardian`'s `metatron.charter_observed` arm, in `internal/sim/guardian.go` (~line 434); reducer test: `Default: true` → false, `Default: false` → true, in `internal/sim/guardian_test.go` (plan D1, spec FR-006)

## Phase 3: User Story 1 — every card surface derives verdicts from EvaluateRubric; ✗ on failure (P1, board AC #1)

**Goal**: postmortem, ceremony, and console card share one fact resolver; the false ✓ dies.

**Independent Test**: run-ended fixture with 2 deaths renders `✗ no villager dies (agent.died: 2)` on all surfaces.

- [ ] T004 [US1] Fact builders from the evaluator: `reportCardFactsFromRubric([]sim.RubricTerm)` and `reportCardFactsFromPass(terms, evidence)` (all-met, evidence-backed — re-read, never re-grade), in `internal/tui/reportcard.go` (plan D3, spec FR-001/002/004)
- [ ] T005 [US1] Shared resolver `resolveReportCardFacts(def, pass) (facts, mode)` — precedence pass → concluded rubric (runEnded) → live rubric; nil replica → nothing — in `internal/tui/reportcard.go`, generalizing `buildChecklistCard`'s switch (plan D3, spec FR-001/002)
- [ ] T006 [US1] Rewire the three surfaces through the resolver: `postmortemReportCard` (`internal/tui/views.go:659`), `ceremonyReportCardFor` (`views.go:756`, `provingPass` kept; aged-out → nil-pass fallback), `buildChecklistCard` (`internal/tui/reportcard.go:87`, stopping-point gate unchanged) (plan D3, spec FR-001)
- [ ] T007 [US1] Delete `reportCardFactsFromCounts`/`FromEvents`/`FromEvidence` and `humanizeEventType` from `internal/tui/views.go` (843-885) plus their direct tests (plan D3, spec FR-003)
- [ ] T008 [US1] TUI truth tests: the motivating regression (postmortem `✗` on `agent.died: 2` — SC-001), recorded-pass all-met on ceremony + console card, live `…` markers, cross-surface identical rows; update `TestReportCardChecklistOnly`, `TestReportCardBothComposeChecklistAboveNote`, `TestConsoleCardSeamComposesReportCard` and ceremony/postmortem fixtures asserting old generic labels, in `internal/tui/{views,takeover,reportcard,console}_test.go` (plan D4, spec US1 scenarios)

## Phase 4: User Story 2 — the-law production evaluator; gauges stop rendering permanently pending (P2, board AC #2)

**Goal**: `EvaluateRubric` evaluates the-law from state facts; the exercise tab needs no change to benefit.

**Independent Test**: `TestTheLawRubricTable` matrix + genesis-replay equivalence.

- [ ] T009 [US2] `case "the-law": return theLawRubric(s)` in `EvaluateRubric` + `theLawRubric` (law term over `len(s.Norms)`, charter term over `CharterFingerprint`+`CharterCustom` — plan D2 shapes); rewrite the blocker doc comment (~lines 276-279) and update `scenarioRubricEvents`' first-night-only comment to cite spec FR-009, in `internal/sim/scenario.go` (spec FR-007/009)
- [ ] T010 [US2] `TestTheLawRubricTable` (default-charter unmet / custom met / norm adoption met+count / nothing pending) and replay-equivalence assertion (live fold vs genesis replay → identical terms, US2-5), in `internal/sim/scenario_test.go` (plan D4, spec SC-002)
- [ ] T011 [US2] Gauge proof on a `the-law` fixture: exercise tab rows flip per the table, no permanently-pending term (extends `TestExerciseGaugesTrackReplica`), in `internal/tui/exercise_test.go` (spec SC-002)

## Phase 5: User Story 3 — design reference amended, authority gate green (P3, board AC #3)

- [ ] T012 [US3] Amend `docs/design/tui/overlays/postmortem.md` (rewrite the Known-simplification block as the shipped rubric contract; fix the scored mockup to truthful glyphs + real labels; leave `unbuilt (wave 4)` cells to TASK-150) and `docs/design/tui/overlays/ceremony.md` (its identical note: pass = instrument all-met, aged-out = rubric fallback) (plan D5, spec FR-010/011)
- [ ] T013 [US3] Amend `docs/design/tui/panels/exercise.md`: stale "(TASK-127, unbuilt)" pointer at line 110; note both cataloged exercises now carry production evaluators, non-evaluator content renders pending (plan D5, spec FR-010)
- [ ] T014 [US3] `node scripts/check-tui-design.mjs --changed` from the worktree: re-verify + re-pin every flagged page (views.go alone pins ~11), amendments only where behavior changed; gate passes (spec SC-003)

## Phase 6: Grounding — wiki-in-PR obligations (in-branch, pr-gate enforced)

- [ ] T015 `/grounding-wiki:wiki-update` reconciliation over the branch diff; review-work re-pins expected on `docs/wiki/report-card-renderer.md`, `scenario-machinery.md`, `scenario-machinery-surfacing.md`, `sim-state-world-fields.md`, `sim-state-reducer.md`, `event-types-guardian-morgue.md`, `guardian-report-card.md`, `tui-dock-tabs.md`, `takeover-surfaces.md`, `curriculum-ladder-progression.md`; computed re-pins for other notes listing `views.go`/`state.go`/`guardian.go`/`scenario.go` — all pinned to branch commits (plan D6, spec SC-005)
- [ ] T016 Regenerate `docs/player/` via the `player-docs` skill (wiki changed in T015); `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` passes in-branch (plan D6, spec SC-005)

## Phase 7: Polish & close-out

- [ ] T017 Full proof: `gofmt -l` clean; `go test ./...` green (existing snapshots load byte-identical — FR-008/SC-004); `node scripts/check-merge-drift.mjs pr` from the worktree exits 0; PR opens carrying code + design + wiki + player docs together; merge via `gh pr merge --merge` only
- [ ] T018 Post-merge (root): spec-bridge sync, board AC ticks, tasks.md ticks, runbook execution-log row — derived state only, no grounding content on main

## Dependencies & Execution Order

- T001 → everything. T002 → T003 → {T009, T010} (the flag blocks the evaluator).
- T004 → T005 → T006 → T007 → T008 (US1 chain; independent of Phase 2 — may run in
  parallel with T002/T003).
- T009 → T010 → T011. T011 also needs T003 (fixture folds a charter event).
- T012/T013 after T006/T009 (pages describe shipped behavior); T014 after T012/T013 and
  all code tasks. T015 after all code; T016 after T015; T017 after everything; T018
  post-merge.
- **MVP** = Phases 2–4 (the truth change); Phase 5 completes board AC #3.

**Parallel opportunities**: Phase 2 ∥ Phase 3 (different packages); T012/T013 authoring
can start once T006/T009 shapes are fixed.
