---
id: TASK-124
title: Chronicle jump-to-source + input-parity retrofit start
status: In Progress
assignee: []
created_date: '2026-07-25 14:44'
updated_date: '2026-07-25 17:17'
labels:
  - learning-game
  - tui
dependencies:
  - TASK-123
priority: medium
ordinal: 94000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient 2026-07-25 (docs/design/reorient-2026-07-25-ui.md, D3 + decision 8), Wave 2 quick win. Fill the chronicle's reserved ⏎ seam: selected event centers the camera on its subject (RimWorld letter / DF zoom-to-event precedent); with parity, click-a-line jumps too. Ratify input parity in patterns/keymap.md (every action keyboard AND mouse, keyboard primary); control tables' mouse column enables a future parity sweep test.

Spec: specs/049-chronicle-jump-to-source
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 ⏎ (and click) on a chronicle event with a located subject centers the map on it; subject-less events are honest no-ops with a hint
- [ ] #2 keymap.md carries the ratified parity rule; new bindings documented keys+mouse
- [ ] #3 Spec phase: Setup
- [ ] #4 Spec phase: Foundational (blocking prerequisites for all stories)
- [ ] #5 Spec phase: User Story 1 — Jump from an event to its subject on the map (P1) 🎯 MVP
- [ ] #6 Spec phase: User Story 2 — Click a chronicle line to select and jump (P2)
- [ ] #7 Spec phase: User Story 3 — The seam advertises itself honestly (P3)
- [ ] #8 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Model tier: Sonnet (spec-implementer default). Rubric: single-package view/rendering slice (internal/tui + 1-line cmd option), tests alongside code, no concurrency/governor/doctrine surface — routine tier per constitution Principle V. Dispatched by UI-sweep orchestrator per runbook Lane 1.
<!-- SECTION:NOTES:END -->
