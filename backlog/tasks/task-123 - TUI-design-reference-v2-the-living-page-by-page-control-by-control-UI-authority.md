---
id: TASK-123
title: >-
  TUI design reference v2: the living page-by-page, control-by-control UI
  authority
status: To Do
assignee: []
created_date: '2026-07-25 14:43'
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
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 docs/design/tui reconciled with specs 013-046: every shipped surface documented where an implementer would look
- [ ] #2 v2 taxonomy in place: anatomy.md, per-tab panels, overlays/, patterns/ incl. skin-tokens and stage-defaults
- [ ] #3 Uniform control tables with keys+mouse and skin-token columns across all panel/overlay pages
- [ ] #4 verified_against pins + check script exist; same-PR amendment enforced by a gate
- [ ] #5 All ten new-surface pages authored (spec-before-build) with mockups + control tables
- [ ] #6 Rulings recorded: bottom-chrome row budget/fold order, narrow fallback, overlay no-LLM invariant
<!-- AC:END -->
