---
id: TASK-93
title: >-
  Fix -race flake: TestPreflightBootNeverFails races compressClock cleanup
  against parallel RunPreflight
status: Done
assignee: []
created_date: '2026-07-24 18:10'
updated_date: '2026-07-25 03:34'
labels:
  - tests
dependencies: []
references:
  - 'https://github.com/evanstern/promptworld/pull/72'
priority: high
ordinal: 11000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Pre-existing race found during TASK-41 (2026-07-24), reproduced on main with zero 037 code: go test -race ./internal/llm/ fails — compressClock cleanup (internal/llm/preflight_test.go:83) writes a package clock var while a PARALLEL test's RunPreflight goroutine (internal/llm/preflight.go:229) still reads it. Passes in isolation; only surfaces under -race with parallel scheduling. Same family as the TASK-69 TestQueueBackpressure flake fix — mirror that approach. Surgical fix; diagnosis pinned.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 go test -race ./internal/llm/ passes repeatedly (≥5 consecutive runs) on main
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Tier: Sonnet (spec-implementer) — routine test-only race fix per rubric: surgical, diagnosis pinned (preflight_test.go:83 vs preflight.go:229), mirrors TASK-69 fix pattern. Dispatched 2026-07-24, worktree .worktrees/task-93.

Implemented (Sonnet spec-implementer): test-only fix — startPreflight helper launches RunPreflight and registers t.Cleanup that cancels AND blocks on a done channel until the goroutine returns; LIFO cleanup ordering guarantees the goroutine finishes before compressClock restores the package clock vars. Race reproduced 5/5 pre-fix, 5/5 -race passes post-fix; production code untouched. Mirrors the TASK-69 fix family (dc3780a). PR: https://github.com/evanstern/promptworld/pull/72. AC#1 ('on main') ticks after merge. NOTE: implementer worked via Bash inside .worktrees/task-93 because the harness Edit sandbox was pinned to the dispatching session's worktree — dispatch future implementers with Agent-tool worktree isolation instead of pre-made .worktrees dirs.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #72 (b9640a4). startPreflight helper synchronizes the RunPreflight goroutine's lifetime with test cleanup (cancel + block on done channel; LIFO cleanup ordering guarantees goroutine exit before compressClock restores clock vars). Verified on merged main: go test -race -count=5 ./internal/llm passes (AC#1). Production code untouched.
<!-- SECTION:FINAL_SUMMARY:END -->
