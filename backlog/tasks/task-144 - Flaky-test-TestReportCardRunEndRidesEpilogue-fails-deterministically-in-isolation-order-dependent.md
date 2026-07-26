---
id: TASK-144
title: >-
  Flaky test: TestReportCardRunEndRidesEpilogue fails deterministically in
  isolation (order-dependent)
status: In Progress
assignee: []
created_date: '2026-07-26 15:14'
updated_date: '2026-07-26 15:54'
labels:
  - flaky-test
dependencies: []
priority: medium
ordinal: 114000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Found during TASK-141 implementation (2026-07-26): internal/guardian TestReportCardRunEndRidesEpilogue passes in full-package runs but fails deterministically when run in isolation (go test ./internal/guardian/ -run 'TestReportCardRunEndRidesEpilogue' -count=5) — reproduced on untouched origin/main at repo root, so it predates the TASK-141 branch. Order-dependent test defect in the guardian report-card queue: the test appears to rely on state another test establishes. Diagnosis needed: find the shared state and make the test self-contained.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Test passes in isolation with -count=5 and in the full suite
- [ ] #2 Root cause (the shared state) named in task notes
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Claim per spec 065: specs/070-guardian-test-order/ stubbed. Sweep lane 1 (runbook docs/design/pdlc-hardening-runbook.md, signed off 2026-07-26). Tier: Sonnet (single-package test fix, routine); escalation to Opus is an operator checkpoint if root cause lands in concurrency machinery.
<!-- SECTION:NOTES:END -->
