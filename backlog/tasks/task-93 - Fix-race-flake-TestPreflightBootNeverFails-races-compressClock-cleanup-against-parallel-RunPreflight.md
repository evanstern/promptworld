---
id: TASK-93
title: >-
  Fix -race flake: TestPreflightBootNeverFails races compressClock cleanup
  against parallel RunPreflight
status: To Do
assignee: []
created_date: '2026-07-24 18:10'
updated_date: '2026-07-25 03:10'
labels:
  - tests
dependencies: []
priority: high
ordinal: 11000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Pre-existing race found during TASK-41 (2026-07-24), reproduced on main with zero 037 code: go test -race ./internal/llm/ fails — compressClock cleanup (internal/llm/preflight_test.go:83) writes a package clock var while a PARALLEL test's RunPreflight goroutine (internal/llm/preflight.go:229) still reads it. Passes in isolation; only surfaces under -race with parallel scheduling. Same family as the TASK-69 TestQueueBackpressure flake fix — mirror that approach. Surgical fix; diagnosis pinned.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 go test -race ./internal/llm/ passes repeatedly (≥5 consecutive runs) on main
<!-- AC:END -->
