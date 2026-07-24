---
id: TASK-94
title: >-
  TestCatalogSweep red on main: daemon.llm_warning backticked in event-types.md
  but uncovered by catalog fixture
status: To Do
assignee: []
created_date: '2026-07-24 18:27'
labels:
  - bug
dependencies: []
priority: medium
ordinal: 78000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Pre-existing spec-034 drift, surfaced while gating TASK-91/spec 038: docs/wiki/event-types.md backticks daemon.llm_warning (lines ~94/147/200) but internal/tui digest_test.go TestCatalogSweep finds no catalog fixture row for it, and as an operator-facing state-no-op event it has no digest renderer either — so it cannot just be added to the fixture (the fixture→registry check would then fail). Needs a design decision: render it in the digest, un-backtick it in the doc, or exempt it in the sweep. Verified failing on pristine main (go test ./internal/tui/ -run TestCatalogSweep). Keeps internal/tui red for every branch.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 go test ./internal/tui/ is green on main
- [ ] #2 Chosen resolution (render/un-backtick/exempt) recorded on this task
<!-- AC:END -->
