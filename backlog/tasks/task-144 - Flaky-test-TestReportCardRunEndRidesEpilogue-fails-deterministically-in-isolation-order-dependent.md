---
id: TASK-144
title: >-
  Flaky test: TestReportCardRunEndRidesEpilogue fails deterministically in
  isolation (order-dependent)
status: In Progress
assignee: []
created_date: '2026-07-26 15:14'
updated_date: '2026-07-26 16:06'
labels:
  - flaky-test
dependencies: []
priority: medium
ordinal: 114000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
CORRECTED DIAGNOSIS (2026-07-26, planning-tier investigation; supersedes the original framing): NOT order-dependent shared state — no test seeds state for another. Root cause is a goroutine-shutdown race in the test fixture: guardian.New (internal/guardian/guardian.go:262) spawns reportCardWorker; newTestGuardian (guardian_test.go:141) calls Close() which is just close(mt.done) with NO JOIN (guardian.go:274). If the worker hasn't parked in its select (reportcard.go:166-175) before the test enqueues into cardQ, its first select sees BOTH cases ready and Go picks uniformly at random — ~50% chance it steals the job, leaving drainCard empty → 'no card job queued'. Probabilistic (~5%/run), NOT deterministic: -count=1 passes 10/10; -count=5 fails ~iteration 3; -count=3 full-package fails SIBLING tests (TestReportCardProducerStoresValidatedNote, TestReportCardRejectsUnrecordedCitation). Smoking gun: produceCard log lines emitted after the test's t.Fatal — only the worker consumes cardQ besides the test. digestWorker (digest.go:186) and triggerWorker (orders.go:453) share the identical select shape and latent exposure. Fix shape: make Close a join — sync.WaitGroup around the four goroutines started in New, Close = close(done) then wg.Wait(). Test-hygiene class, not a live production bug (production never Closes then drives queues), but the fix lands in production shutdown code. Caveat: a job enqueued BEFORE Close may still be randomly processed during shutdown — pre-existing, out of scope.

Spec: specs/070-guardian-test-order
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Test passes in isolation with -count=5 and in the full suite
- [ ] #2 Root cause (the shared state) named in task notes
- [ ] #3 internal/guardian report-card/digest/trigger tests pass -count=50 clean in isolation and full-package (was -count=5 — too weak, 5 iterations pass by luck)
- [ ] #4 Spec phase: Setup
- [ ] #5 Spec phase: User Story 1 + 2 — deterministic worker shutdown (P1)
- [ ] #6 Spec phase: Grounding (in-branch, per the in-PR doctrine)
- [ ] #7 Spec phase: Polish
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Claim per spec 065: specs/070-guardian-test-order/ stubbed. Sweep lane 1 (runbook docs/design/pdlc-hardening-runbook.md, signed off 2026-07-26). Tier: Sonnet (single-package test fix, routine); escalation to Opus is an operator checkpoint if root cause lands in concurrency machinery.

Spec 070 authored (compact — diagnosis pinned, fix prescribed: WaitGroup join in Close). Tier checkpoint resolved by operator 2026-07-26: Sonnet holds (concurrency analysis complete at planning tier; fix surgical). AC strengthened to -count=50.

Spec: specs/070-guardian-test-order
<!-- SECTION:NOTES:END -->
