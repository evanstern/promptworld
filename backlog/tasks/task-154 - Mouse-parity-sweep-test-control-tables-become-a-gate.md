---
id: TASK-154
title: 'Mouse-parity sweep test: control tables become a gate'
status: In Progress
assignee: []
created_date: '2026-07-26 17:57'
updated_date: '2026-07-26 18:15'
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

Spec: specs/073-mouse-parity-sweep
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A test parses control tables and fails on any non-'—' mouse cell without a handler
- [ ] #2 patterns/keymap.md rollout note updated to reflect mechanized tracking
- [ ] #3 Spec phase: Setup
- [ ] #4 Spec phase: US1 — the sweep test (board AC #1: parses control tables, fails on any non-'—' mouse cell without a handler)
- [ ] #5 Spec phase: US2 — mechanized-tracking note (board AC #2: patterns/keymap.md rollout note updated)
- [ ] #6 Spec phase: Grounding + gates (in-branch, per the wiki-in-PR lifecycle)
- [ ] #7 Spec phase: Post-merge bookkeeping (root, derived state only)
<!-- AC:END -->



## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Sweep claim (runbook docs/design/reorient-2026-07-26-sweep-runbook.md): spec 073-mouse-parity-sweep. Tier: Sonnet — tooling/test, single surface (control-table parser + handler assertion, keymap.md rollout-note burn-down).
<!-- SECTION:NOTES:END -->
