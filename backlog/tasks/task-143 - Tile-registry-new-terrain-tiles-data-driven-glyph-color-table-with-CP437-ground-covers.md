---
id: TASK-143
title: >-
  Tile registry + new terrain tiles: data-driven glyph/color table with CP437
  ground covers
status: Done
assignee: []
created_date: '2026-07-26 14:45'
updated_date: '2026-07-26 16:14'
labels:
  - game-ui
  - tui
dependencies: []
priority: medium
ordinal: 113000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Execute the Game-UI-UX tile-vocabulary-expansion analysis (research/Game-UI-UX/Analysis-Tile-Vocabulary-Expansion.md; briefing https://claude.ai/code/artifact/05415ad3-3efd-4693-9330-c626f4435731). Two-part deliverable ratified by the operator (2026-07-26): (1) extract the shipped map glyph/color style table (internal/tui/views.go:1581-1644 + tile switch 1290-1330) into a data-driven tile registry — one tile table feeding renderer, legend line, and ? overlay, colors as skin-style tokens not literals (Rule 4: one grid model, swappable skins; Cogmind externalized-colors precedent; TASK-121 token contract precedent); (2) add 2-3 new ground-cover terrain kinds from the analysis's CP437 shading tier (e.g. marsh, sand) wired through worldgen and the TUI via the new registry. Constraints from the analysis: semantics on the themeable ANSI 16, materials on 256; states recolor, never re-glyph; never color alone for actionable distinctions; every glyph must clear the existing set clustered at small sizes; legend and ? overlay stay one shared source. Non-trivial: full Spec Kit before implementation.

Implementation tier (constitution V): Opus 4.8 — cross-package (tui+worldmap+world+sim), FormatVersion 4→5 break with migration, determinism-sensitive generation (rubric: cross-package/architectural).

Spec: specs/068-tile-registry
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A tile registry (data, not code literals) is the single source for map glyph+style; renderMapGrid, legendGlyphLine, and the ? overlay all read from it; no per-tile style literals remain in views.go
- [x] #2 2-3 new ground-cover terrain kinds (CP437 shading tier) generate in new worlds, render via the registry, appear in legend and ? overlay, and are covered by worldmap + TUI tests
- [x] #3 Existing worlds/replays render byte-identically for the pre-existing vocabulary (no regression in golden/TUI tests)
- [x] #4 Spec linked on the board via spec-bridge:link before implementation; wiki re-pinned after merge
- [x] #5 Spec phase: Setup — pin current behavior BEFORE touching anything
- [x] #6 Spec phase: Foundational — the registry substrate
- [x] #7 Spec phase: User Story 1 — one tile table drives map, legend, overlay (P1) 🎯 MVP
- [x] #8 Spec phase: User Story 2 — marsh and sand (P2)
- [x] #9 Spec phase: Polish & cross-cutting
<!-- AC:END -->



## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
spec-bridge sync: Setup — pin current behavior BEFORE touching anything: 2/2 · Foundational — the registry substrate: 1/1 · User Story 1 — one tile table drives map, legend, overlay (P1) 🎯 MVP: 5/5 · User Story 2 — marsh and sand (P2): 6/6 · Polish & cross-cutting: 1/2

spec-bridge sync: Setup — pin current behavior BEFORE touching anything: 2/2 · Foundational — the registry substrate: 1/1 · User Story 1 — one tile table drives map, legend, overlay (P1) 🎯 MVP: 5/5 · User Story 2 — marsh and sand (P2): 6/6 · Polish & cross-cutting: 2/2 — status In Progress → Done
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Setup — pin current behavior BEFORE touching anything: 2/2 · Foundational — the registry substrate: 1/1 · User Story 1 — one tile table drives map, legend, overlay (P1) 🎯 MVP: 5/5 · User Story 2 — marsh and sand (P2): 6/6 · Polish & cross-cutting: 2/2). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
