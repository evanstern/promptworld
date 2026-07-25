---
id: TASK-124
title: Chronicle jump-to-source + input-parity retrofit start
status: Done
assignee: []
created_date: '2026-07-25 14:44'
updated_date: '2026-07-25 18:05'
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
- [x] #1 ⏎ (and click) on a chronicle event with a located subject centers the map on it; subject-less events are honest no-ops with a hint
- [x] #2 keymap.md carries the ratified parity rule; new bindings documented keys+mouse
- [x] #3 Spec phase: Setup
- [x] #4 Spec phase: Foundational (blocking prerequisites for all stories)
- [x] #5 Spec phase: User Story 1 — Jump from an event to its subject on the map (P1) 🎯 MVP
- [x] #6 Spec phase: User Story 2 — Click a chronicle line to select and jump (P2)
- [x] #7 Spec phase: User Story 3 — The seam advertises itself honestly (P3)
- [x] #8 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Model tier: Sonnet (spec-implementer default). Rubric: single-package view/rendering slice (internal/tui + 1-line cmd option), tests alongside code, no concurrency/governor/doctrine surface — routine tier per constitution Principle V. Dispatched by UI-sweep orchestrator per runbook Lane 1.

spec-bridge sync: 17/17 tasks done — merged via PR #84 (c388c41) after rebase over the guardian-strip and spec-048 merges (one test-file conflict, both-appended-tests, resolved keep-both). Gates green post-rebase. Deviations recorded in-PR: actions bar now a permanent detail-pane row; intent_set jumps to target; entity_moved jumps to destination.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Chronicle jump-to-source shipped via spec 049 + PR #84 (merge c388c41). ⏎/click on a selected chronicle event centers the map on its subject (65-type subject registry: live actor position → payload coordinates → honest hint); pan-equivalent camera write; permanent detail-pane actions bar (jump affordance or 'no location'); first mouse-bound control (parity doctrine ratified in keymap.md — rule 3 names it); narrow fallback lands on the map pane. Design pages chronicle.md/keymap.md amended + re-pinned in-PR; wiki re-verified (tui-client jump prose, 2bb6ab9); player-docs refresh dispatched. Gates green post-rebase over the guardian-strip and spec-048 merges.
<!-- SECTION:FINAL_SUMMARY:END -->
