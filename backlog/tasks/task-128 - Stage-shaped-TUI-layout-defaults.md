---
id: TASK-128
title: Stage-shaped TUI layout defaults
status: In Progress
assignee: []
created_date: '2026-07-25 14:45'
updated_date: '2026-07-25 23:58'
labels:
  - learning-game
  - tui
dependencies:
  - TASK-123
priority: medium
ordinal: 98000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient 2026-07-25 (decision 3), Wave 4. Which panels/tabs/chrome are visible BY DEFAULT is stage-resolved (stage-1 boots map + narrated chronicle + guardian line + lesson row; traces/raw feed/systems surface as stages arrive, each announced by a first-occurrence lesson). Defaults only: everything reachable at every stage and via ?; pre-ladder worlds get everything; capability locks stay angel-only (spec 046 doctrine untouched).

Spec: specs/066-stage-defaults
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Stage-resolved default visibility exists; every surface remains reachable at every stage
- [ ] #2 Pre-ladder worlds byte-identical to ungated full layout
- [ ] #3 Spec phase: Setup
- [ ] #4 Spec phase: Foundational (blocking all user stories)
- [ ] #5 Spec phase: User Story 1 — A stage-1 player boots into the focused layout (P1) 🎯 MVP
- [ ] #6 Spec phase: User Story 2 — Pre-ladder worlds are untouched (P1)
- [ ] #7 Spec phase: User Story 3 — Surfaces arrive with the stage (P2)
- [ ] #8 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
spec-bridge link: specs/066-stage-defaults attached (spec+plan+tasks complete, 16 tasks; derived status In Progress — spec phase done, implementation not dispatched). Dispatch gated on TASK-119's merge per runbook Lane 4 (128 runs after the tabs/rows it governs exist: 125 ✓, 117 ✓, 119 pending).
<!-- SECTION:NOTES:END -->
