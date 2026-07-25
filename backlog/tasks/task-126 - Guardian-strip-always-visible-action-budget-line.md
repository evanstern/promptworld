---
id: TASK-126
title: 'Guardian strip: always-visible action budget line'
status: Done
assignee: []
created_date: '2026-07-25 14:44'
updated_date: '2026-07-25 17:54'
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
- [x] #1 Strip visible in composite regardless of dock tab; content stage/skin-resolved; collapses per Wave-0 fold-order ruling
- [x] #2 Spec phase: Setup
- [x] #3 Spec phase: Foundational
- [x] #4 Spec phase: User Story 1 — The budget is one glance from the verb (P1) 🎯 MVP
- [x] #5 Spec phase: User Story 2 — The strip never lies (P2)
- [x] #6 Spec phase: User Story 3 — The strip survives pressure (P3)
- [x] #7 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Model tier: Sonnet (spec-implementer default). Rubric: single-package rendering + pure layout arithmetic in internal/tui (one read-only const export in sim), tests alongside — routine tier per constitution Principle V. Dispatched by UI-sweep orchestrator per runbook Lane 1.

spec-bridge sync: 14/14 tasks done — merged via PR #83 (f18e9a4) after clean rebase; gates green post-rebase (race suite + check-tui-design). Deviations recorded on the design page: replica-sourced order count; ⚡⚡· (2/3) segment form; pre-status empty fold prefix; narrow carry scoped to the guardian pane (the only narrow composer).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Guardian strip shipped via spec 050 + PR #83 (merge f18e9a4). One borderless budget row above the minibuffer (charge bank ⚡+(N/cap), next-regen forecast via newly exported sim.MetatronChargeRegenTicks, replica-sourced standing-order count); honest degradation (no faith segment, regen omitted at full bank, blank pre-status row); folds last by relocating into the dormant minibuffer line; carried in narrow (guardian pane). Design pages guardian-strip.md→shipped, layout/minibuffer/anatomy re-pinned in-PR; wiki re-verified (6 notes → 4b15038); player-docs refresh dispatched. All gates green post-rebase.
<!-- SECTION:FINAL_SUMMARY:END -->
