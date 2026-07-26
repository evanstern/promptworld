---
id: TASK-154
title: 'Mouse-parity sweep test: control tables become a gate'
status: To Do
assignee: []
created_date: '2026-07-26 17:57'
labels:
  - tooling
  - game-ui
dependencies: []
references:
  - docs/design/reorient-2026-07-26-ui.md
ordinal: 124000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient 2026-07-26 decision 8. Input parity (2026-07-25 decision 8) is doctrine-with-a-worklist: patterns/keymap.md's Parity rollout note admits internal/tui predates it outside jump-to-source. Parse the design corpus control tables' keys+mouse column, assert every non-'—' mouse cell has a real handler (help_test.go keymap-sweep precedent), and burn down the rollout note as targets land.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A test parses control tables and fails on any non-'—' mouse cell without a handler
- [ ] #2 patterns/keymap.md rollout note updated to reflect mechanized tracking
<!-- AC:END -->
