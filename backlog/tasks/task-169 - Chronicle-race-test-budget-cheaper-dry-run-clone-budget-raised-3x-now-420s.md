---
id: TASK-169
title: >-
  Chronicle -race test budget: cheaper dry-run clone (budget raised 3x, now
  420s)
status: To Do
assignee: []
created_date: '2026-07-30 02:38'
labels:
  - debt
  - testing
dependencies: []
priority: medium
ordinal: 137000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The chronicle -race test's wall-clock budget has been raised three times (240s -> 420s in spec 097's branch; branch runtime 262s vs main's 161s). The test's own comment marks the cheaper-dry-run-clone follow-up as DUE. Deliverable: make the dry-run clone cheap enough to restore a tight budget.

Carded from TASK-80's implementation report (spec 097, PR #141), board-sweep-2026-07-29.
<!-- SECTION:DESCRIPTION:END -->
