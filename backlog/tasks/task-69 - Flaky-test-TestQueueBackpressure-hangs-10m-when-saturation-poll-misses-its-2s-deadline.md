---
id: TASK-69
title: >-
  Flaky test: TestQueueBackpressure hangs 10m when saturation poll misses its 2s
  deadline
status: Done
assignee: []
created_date: '2026-07-23 04:58'
updated_date: '2026-07-24 17:02'
labels: []
dependencies: []
priority: medium
ordinal: 5000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Found while gating TASK-48 (commit 67c648b worktree, identical to main for this file region). Under CPU contention the test's saturation poll (llm_test.go:291-297, 2s deadline waiting for Queue >= queueCap) can expire before the 33 goroutine submits saturate the tier. The overflow Submit (llm_test.go:302) then ENQUEUES instead of getting ErrQueueFull and blocks forever in the reply select (llm.go:421, background ctx). The single worker grinds release-blocked jobs at workerCallCap=2min each (llm.go:225) — goroutine dump showed handler arrivals at ~2min cadence (9/7/4/2 min waits) until the go test 10m timeout panics. Fix direction: make the overflow submit carry a short context timeout, or t.Fatal when the saturation poll times out instead of proceeding, so a missed race fails in seconds not 10 minutes. Evidence: goroutine dump in TASK-48 session, 2026-07-23.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Root cause confirmed against a reproduced hang or reasoned trace
- [x] #2 Test fails fast (seconds) when saturation is not reached; go test ./internal/llm/ -count=10 green under load
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Trivial-exemption path (surgical fix, file:line diagnosis pinned on task, ACs on task — no spec). 1) Confirm root cause by reasoned trace against current pins (done: poll llm_test.go:307-313 falls through silently on timeout; overflow Submit :314 w/ background ctx enqueues and blocks at reply select llm.go:1010; workerCallCap=2min llm.go:302/:1087 matches dump cadence). 2) Fix in worktree .worktrees/task-69 branch task-69-flaky-queuebackpressure: raise saturation poll deadline (generous, exits early when saturated) + t.Fatal when saturation not reached, + short-timeout ctx on overflow Submit as backstop — missed race fails in seconds, not 10m. 3) Gate: go test ./internal/llm/ -count=10 green under load. 4) One PR.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Drift audit 2026-07-23: facts hold; line pins moved. Saturation poll 2s deadline now llm_test.go:307-309 (test at :288); overflow Submit with background ctx llm_test.go:303-304; blocking reply select is llm.go:721-722 (NOT :421); workerCallCap=2min at llm.go:284 (used :799).

Tier: Sonnet (spec-implementer default). Rubric: 'tests alongside code' — test-only surgical fix with design pinned on the task; no production concurrency/scheduling logic touched (internal/llm escalation line covers logic changes, not this).

Fix landed on branch task-69-flaky-queuebackpressure (commit dc3780a), PR #65 open. Test-only: saturation poll deadline 2s→10s with t.Fatalf(observed queue depth) on miss; overflow Submit bounded by 5s ctx timeout as regression backstop. Gates run twice (implementer + orchestrator): go build/vet/gofmt clean, go test ./internal/llm/ -count=10 green (~21.5s). No wiki impact: no docs/wiki note lists llm_test.go as a source. Remaining: merge PR #65, then Done + worktree cleanup.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Flake fixed test-only in PR #65 (merged as 71c17ff). Root cause confirmed by trace: saturation poll fell through silently on its 2s deadline, letting the overflow Submit (background ctx) enqueue and block forever in the reply select while the worker ground release-blocked jobs at workerCallCap=2min — hence the 10m go-test panic. Fix: poll deadline 10s + t.Fatalf(observed queue depth) when saturation is missed (fails in seconds), overflow Submit bounded by 5s ctx as regression backstop. Gates green twice: go test ./internal/llm/ -count=10 (~21.5s), build/vet/gofmt clean. No wiki impact (llm_test.go not a pinned source). Worktree and branch cleaned up post-merge.
<!-- SECTION:FINAL_SUMMARY:END -->
