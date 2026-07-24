---
id: TASK-92
title: >-
  Fix TestCatalogSweep: event-types.md backticks daemon.llm_warning but digest
  fixture lacks a row
status: To Do
assignee: []
created_date: '2026-07-24 18:10'
labels:
  - tests
dependencies: []
priority: high
ordinal: 78000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Pre-existing failure found during TASK-41 (2026-07-24), reproduced on main with zero 037 code: internal/tui/digest_test.go:207 — docs/wiki/event-types.md backticks "daemon.llm_warning" (a spec-034 event type) but catalogFixture has no row for it, so TestCatalogSweep fails on every run. Docs↔fixture drift: either add a fixture row (+ digest entry if the type lacks one) or correct the wiki catalog. Surgical fix; diagnosis pinned.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 go test ./internal/tui/ -run TestCatalogSweep passes on main
<!-- AC:END -->
