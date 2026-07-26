---
id: TASK-154
title: 'Mouse-parity sweep test: control tables become a gate'
status: Done
assignee: []
created_date: '2026-07-26 17:57'
updated_date: '2026-07-26 18:32'
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
- [x] #1 A test parses control tables and fails on any non-'—' mouse cell without a handler
- [x] #2 patterns/keymap.md rollout note updated to reflect mechanized tracking
- [x] #3 Spec phase: Setup
- [x] #4 Spec phase: US1 — the sweep test (board AC #1: parses control tables, fails on any non-'—' mouse cell without a handler)
- [x] #5 Spec phase: US2 — mechanized-tracking note (board AC #2: patterns/keymap.md rollout note updated)
- [x] #6 Spec phase: Grounding + gates (in-branch, per the wiki-in-PR lifecycle)
- [x] #7 Spec phase: Post-merge bookkeeping (root, derived state only)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Sweep claim (runbook docs/design/reorient-2026-07-26-sweep-runbook.md): spec 073-mouse-parity-sweep. Tier: Sonnet — tooling/test, single surface (control-table parser + handler assertion, keymap.md rollout-note burn-down).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #112 (merge sha 86b776d, merge commit). TestMouseParitySweep gates the design corpus's mouse claims: canonical control-table parsing, bidirectional oracle, live tea.MouseMsg dispatch proof (initial entry: chronicle jump-to-source). keymap.md rule 3 records the graduation contract; mutation check proved the gate bites. Player-docs pin refreshed in-branch. Full suite, check-tui-design --changed, freshness 13/13, merge-drift pr gate all green.
<!-- SECTION:FINAL_SUMMARY:END -->
