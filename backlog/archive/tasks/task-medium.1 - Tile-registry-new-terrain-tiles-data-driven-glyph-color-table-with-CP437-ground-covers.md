---
id: TASK-MEDIUM.1
title: >-
  Tile registry + new terrain tiles: data-driven glyph/color table with CP437
  ground covers
status: To Do
assignee: []
created_date: '2026-07-26 14:45'
labels:
  - game-ui
  - tui
dependencies: []
parent_task_id: TASK-MEDIUM
ordinal: 113000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Execute the Game-UI-UX tile-vocabulary-expansion analysis (research/Game-UI-UX/Analysis-Tile-Vocabulary-Expansion.md; briefing https://claude.ai/code/artifact/05415ad3-3efd-4693-9330-c626f4435731). Two-part deliverable ratified by the operator (2026-07-26): (1) extract the shipped map glyph/color style table (internal/tui/views.go:1581-1644 + tile switch 1290-1330) into a data-driven tile registry — one tile table feeding renderer, legend line, and ? overlay, colors as skin-style tokens not literals (Rule 4: one grid model, swappable skins; Cogmind externalized-colors precedent; TASK-121 token contract precedent); (2) add 2-3 new ground-cover terrain kinds from the analysis's CP437 shading tier (e.g. marsh, sand) wired through worldgen and the TUI via the new registry. Constraints from the analysis: semantics on the themeable ANSI 16, materials on 256; states recolor, never re-glyph; never color alone for actionable distinctions; every glyph must clear the existing set clustered at small sizes; legend and ? overlay stay one shared source. Non-trivial: full Spec Kit before implementation.
<!-- SECTION:DESCRIPTION:END -->
