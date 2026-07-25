---
id: TASK-114
title: >-
  Player docs: screen-orientation page, keys reference card, losing-is-fun
  paragraph
status: To Do
assignee: []
created_date: '2026-07-25 04:26'
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
- [ ] #1 An eighth player-docs page exists answering 'what do all those things on screen mean', organized by screen region, with a complete map glyph table
- [ ] #2 A keys reference card page/section exists, pure controls, unmixed with lore
- [ ] #3 getting-started.html opens expectations with the losing-is-fun reassurance paragraph
- [ ] #4 New pages pass the player-docs freshness gate (check-freshness.mjs --check)
<!-- AC:END -->
