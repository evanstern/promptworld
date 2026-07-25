---
id: TASK-126
title: 'Guardian strip: always-visible action budget line'
status: In Progress
assignee: []
created_date: '2026-07-25 14:44'
updated_date: '2026-07-25 17:23'
labels:
  - learning-game
  - tui
dependencies:
  - TASK-123
priority: medium
ordinal: 96000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient 2026-07-25 (decision 7), Wave 2. One line above the minibuffer pairing the action budget with the input: charge bank, regen countdown, standing-order count, faith gauge once TASK-118 lands. Makes the minibuffer read as THE verb; today the budget hides in one tab's pane header.

Spec: specs/050-guardian-strip
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Strip visible in composite regardless of dock tab; content stage/skin-resolved; collapses per Wave-0 fold-order ruling
- [ ] #2 Spec phase: Setup
- [ ] #3 Spec phase: Foundational
- [ ] #4 Spec phase: User Story 1 — The budget is one glance from the verb (P1) 🎯 MVP
- [ ] #5 Spec phase: User Story 2 — The strip never lies (P2)
- [ ] #6 Spec phase: User Story 3 — The strip survives pressure (P3)
- [ ] #7 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Model tier: Sonnet (spec-implementer default). Rubric: single-package rendering + pure layout arithmetic in internal/tui (one read-only const export in sim), tests alongside — routine tier per constitution Principle V. Dispatched by UI-sweep orchestrator per runbook Lane 1.
<!-- SECTION:NOTES:END -->
