---
id: TASK-123
title: >-
  TUI design reference v2: the living page-by-page, control-by-control UI
  authority
status: In Progress
assignee: []
created_date: '2026-07-25 14:43'
updated_date: '2026-07-25 15:06'
labels:
  - learning-game
  - tui
  - design
dependencies: []
priority: high
ordinal: 93000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient run 2026-07-25 (docs/design/reorient-2026-07-25-ui.md), Wave 0 — the deliverable. Reconcile docs/design/tui with shipped reality (specs 013-046; panels/dock.md is ~6 specs stale vs the shipped metatron pane); adopt the v2 taxonomy (pages/ · panels/ per dock tab · overlays/ · patterns/ + top-level anatomy.md region index); uniform control tables (control/region · states · data source · renderer · keys+mouse · introduced-by · skin-token) as the AI-parseable unit; verified_against pins + check script + same-PR freshness gate (TASK-82 precedent, decision 4). Author new-surface pages spec-before-build: guardian console, systems tab, exercise panel, ceremony overlay, postmortem overlay, lesson row, guardian strip, villager strip, stage-defaults pattern, ? guardian section. Wave 0 rules on: bottom-chrome row budget + fold order (~4 fixed rows at stages 1-2), narrow-fallback behavior for new chrome, and the ? overlay's no-LLM byte-identity invariant restated for status-derived sections. Grounding: research/Game-Player-Docs/Analysis-UI-Reference-And-Help-Stack.md (structure), research/Game-UI-UX/Analysis-Teaching-Game-TUI.md (staleness audit + build order).

Spec: specs/047-tui-design-reference-v2
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 docs/design/tui reconciled with specs 013-046: every shipped surface documented where an implementer would look
- [ ] #2 v2 taxonomy in place: anatomy.md, per-tab panels, overlays/, patterns/ incl. skin-tokens and stage-defaults
- [ ] #3 Uniform control tables with keys+mouse and skin-token columns across all panel/overlay pages
- [ ] #4 verified_against pins + check script exist; same-PR amendment enforced by a gate
- [ ] #5 All ten new-surface pages authored (spec-before-build) with mockups + control tables
- [ ] #6 Rulings recorded: bottom-chrome row budget/fold order, narrow fallback, overlay no-LLM invariant
- [ ] #7 Spec phase: Setup
- [ ] #8 Spec phase: Foundational (blocking prerequisites)
- [ ] #9 Spec phase: User Story 1 — reconcile shipped reality + taxonomy (P1)
- [ ] #10 Spec phase: User Story 2 — ten new-surface pages spec-before-build (P1)
- [ ] #11 Spec phase: User Story 5 — the Wave 0 rulings recorded (P2)
- [ ] #12 Spec phase: User Story 3 — uniform control tables + parity doctrine (P2)
- [ ] #13 Spec phase: User Story 4 — freshness mechanization (P2)
- [ ] #14 Spec phase: Polish & cross-cutting
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Full Spec Kit: specify (spec 047) → clarify (operator-reserved Qs from reorient doc) → plan → tasks
2. spec-bridge:link TASK-123 ↔ specs/047 before implementation
3. Implement via spec-implementer agent in .worktrees/task-123 (one PR)
4. spec-bridge:sync + wiki/player-docs freshness checks at close
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Grounding read: docs/design/reorient-2026-07-25-ui.md (decisions 1-8, D1-D13), research/Game-UI-UX/Analysis-Teaching-Game-TUI.md, research/Game-Player-Docs/Analysis-UI-Reference-And-Help-Stack.md. Next spec number verified free: 047 (local + origin/main).

Spec Kit complete: spec 047 specified (07a11bb), planned (565addc), tasks (2c903d0) — 32 tasks / 8 phases. Operator rulings 2026-07-25: ambient postmortem = morgue evidence only; ceremony voice = both, instrument authoritative; raw registry values behind debug toggle. Tier: Sonnet via spec-implementer (rubric: doc reconciliation + single-file zero-dep tooling; all judgment resolved at planning tier in research.md R1-R8). Linked via spec-bridge.
<!-- SECTION:NOTES:END -->
