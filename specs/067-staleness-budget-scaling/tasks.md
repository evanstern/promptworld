# Tasks: Staleness budget scaling — planning must survive clock speed

**Input**: Design documents from `/specs/067-staleness-budget-scaling/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/landing-gate.md, quickstart.md

**Tests**: included — the spec's success criteria explicitly demand the test proofs (SC-001..SC-003).

**Organization**: grouped by user story; US1/US2 are both P1 (US2's replay proof gates the mechanism US1 introduces).

## Phase 1: Setup

- [x] T001 Baseline: from the task worktree, run `go test ./...` and record the suite green before any change (worktree `.worktrees/task-141`)

## Phase 2: Foundational (blocking prerequisites)

- [x] T002 Add `EffectiveBudgetTicks(ticksPerSecond float64) int64` to `DecisionClass` and rewrite the decision-4 doctrine comment (BudgetTicks = budget at 1x, wall-clock patience, enforced scaled at delivery gates; values stay reviewed-code doctrine) in `internal/cognition/registry.go` (plan D1, research R1/R3)
- [x] T003 [P] Table tests for the helper — 1x identity, 4x/8x/16x/32x products for every registry class (contract reference table), uncapped `tps <= 0` fallback — in `internal/cognition/registry_test.go`

## Phase 3: User Story 1 — Planning survives 8x on a calibrated local tier (P1)

**Goal**: admitted plan thoughts land at 8x in the measured regime instead of dying rejected-stale.

**Independent Test**: `go test ./internal/sim/ -run 'Stale'` proves the acceptance scenarios with the TASK-122 measured numbers.

- [x] T004 [US1] Scale the landing gate: `rungStale` takes the tick rate from `l.state.Speed.TicksPerSecond()` at its one call site, compares staleness to `dc.EffectiveBudgetTicks(tps)`, and emits the scaled reason grammar `staleness %d > budget %d (%d at 1x × %gx)` (unscaled form at uncapped speed) in `internal/sim/landing.go` (plan D2, contract grammar)
- [x] T005 [US1] Update/extend landing tests for the spec's US1 acceptance scenarios — 2000 ticks @8x lands, 1300 @1x rejects unchanged, >9600 @8x rejects with the scaled-reason grammar asserted verbatim, uncapped guard, existing 1x assertions untouched — in `internal/sim/landing_test.go` (+ `internal/sim/cognition_test.go` if its rejected-stale fixture asserts the old reason string)
- [x] T006 [US1] Scale the mind-side conversation scene-staleness pre-abort through the same helper (replica speed's `TicksPerSecond()`), keeping the `scene staleness …` prefix with the contract budget clause, in `internal/mind/convo.go` (plan D3, research R2 #4)

**Checkpoint**: US1 independently testable — landing gate + pre-abort share one predicate; measured-regime scenarios pass.

## Phase 4: User Story 2 — Replay determinism across speed changes (P1)

**Goal**: landing outcomes reproduce byte-identically when replaying a log with a mid-flight `speed.set`.

**Independent Test**: `go test ./internal/sim/ -run 'StalenessReplay'`.

- [x] T007 [US2] New reducer replay test: record a run where a classed thought is admitted at 4x, `speed.set` to 16x lands mid-flight, the intent lands; replay the log and assert identical landing outcomes (mirror the `internal/sim/governor_replay_test.go` proof pattern) in `internal/sim/staleness_replay_test.go` (plan D4.3)

**Checkpoint**: mechanism proven pure — spec FR-002/SC-002.

## Phase 5: User Story 3 — Status surfaces stop calling structural death "thinking" (P2)

**Goal**: recorded audits show the true arithmetic; the residual structural-death regime is documented.

**Independent Test**: reason-grammar assertions (T005) + doc review against spec FR-006.

- [x] T008 [US3] Verify the decision-trace/digest surfaces render the scaled reason without change (they map outcome codes, never parse reasons — confirm no test in `internal/tui/` asserts the old string; update any that does) in `internal/tui/decisions.go` / `internal/tui/digest.go` test files only
- [x] T009 [US3] Document the scheduling-vs-delivery gate split and the residual structural-death regime (wall latency > BudgetTicks seconds dies at every speed while the horizon reports healthy) in the cognition-horizon wiki note(s) via `/grounding-wiki:wiki-update` re-pin — runs at repo root after the PR merges (spec FR-006/SC-005, plan D5)

## Phase 6: Polish & Cross-Cutting

- [x] T010 Full verification from the worktree: `gofmt -l` clean on touched files and `go test ./...` green (SC-002/SC-003); run `node scripts/check-merge-drift.mjs pr` before opening the PR
- [x] T011 Measured-run evidence (SC-001 measured arm, optional if impractical): rerun the TASK-122 measure-world profile at 8x, count `set_plan` landings vs planner `rejected-stale`, expect <10% rejected share (was ~91%); record numbers as a TASK-141 board note (quickstart §measured-run)

## Dependencies & Execution Order

- Phase 2 blocks everything (both delivery gates consume the helper).
- US1 (T004–T006) and US2 (T007) both depend on T002; T007 also depends on T004 (it exercises the scaled gate) — execute US1 then US2.
- US3: T008 after T004 (grammar exists); T009 after merge (root-side re-ground).
- Polish: T010 after all code tasks; T011 independent of T010, post-merge.

**Parallel opportunities**: T003 ∥ T004 (different packages); T005 ∥ T006 (different packages); T008 ∥ T007.

## Implementation Strategy

MVP = Phase 2 + US1 (the defect fix) with US2's proof immediately after — both P1, one worktree, one PR (TASK-141). T009/T011 are the post-merge grounding/evidence tail the PDLC requires before the board task closes.
