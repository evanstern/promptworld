---
id: TASK-114
title: >-
  Player docs: screen-orientation page, keys reference card, losing-is-fun
  paragraph
status: Done
assignee: []
created_date: '2026-07-25 04:26'
updated_date: '2026-07-25 05:06'
labels: []
dependencies: []
ordinal: 85000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Approved 2026-07-25 (learning-game synthesis, docs/design/learning-game-synthesis.md — Wave 0). Content-only extension of the player-docs skill machinery (TASK-82 stays Done). NetHack-chapter-3-shaped eighth page: headed by the player's question ('what do all those things on screen mean?'), organized by screen region (header anatomy incl. speed suffix and suppression badges, map glyph table, dock tabs), glyph table prefaced 'you need not memorize these'. Plus a one-page keys reference card (Analog Game Studies reference-card pattern), and the DF-style losing-is-fun paragraph in getting-started.html ('your first village will probably freeze — that's the story'; world-01 Day-7 exposure death is the precedent). Consider a registry-generated reference section so glyph/cost/key tables cannot rot (CDDA Hitchhiker's Guide pattern). Grounding: research/Game-Player-Docs/Analysis-In-Game-First-Teaching.md recommendation 1.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 An eighth player-docs page exists answering 'what do all those things on screen mean', organized by screen region, with a complete map glyph table
- [x] #2 A keys reference card page/section exists, pure controls, unmixed with lore
- [x] #3 getting-started.html opens expectations with the losing-is-fun reassurance paragraph
- [x] #4 New pages pass the player-docs freshness gate (check-freshness.mjs --check)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Delivered in PR #73 (squash-merged 2026-07-25). Two new pages (understanding-the-screen.html, keys-reference.html — standalone card by design, unmixed with concepts), losing-is-fun paragraph in getting-started.html, nav cross-links across all pages, player-docs skill machinery updated 7→9 pages. Freshness gate: 9 fresh / 0 stale / 0 missing. Implemented by Sonnet spec-implementer per tier rubric; reviewed and merged by orchestrator.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-07-25 04:50
---
Implementation started 2026-07-25. Tier: Sonnet spec-implementer (rubric: single-surface docs/content work — player-docs HTML pages via existing skill machinery, no engine code, no concurrency/doctrine surface; matches 'doc reconciliation / routine slice'). Worktree .worktrees/task-114, branch task-114-player-docs-screen-orientation. Spec rigor: content task executing pre-approved synthesis decision 6 with complete ACs on this card (trivial-track: no code, inputs and outputs fully enumerated in description).
---
<!-- COMMENTS:END -->
