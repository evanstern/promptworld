---
id: TASK-192
title: 'Header row overflows the terminal at 80 columns: 20 frames, 81-83 columns'
status: To Do
assignee: []
created_date: '2026-08-03 03:45'
labels:
  - ui
  - tui
  - bug
dependencies: []
ordinal: 174001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The status line across the top of the screen — the one showing the world name, the day and time, whether the sim is running, and how fast — runs a few characters too wide on an 80-column terminal. It wraps onto a second line, pushing everything below it down by a row.

As a player on a standard 80-column terminal, when I start the client, I want the status line to fit on one row so it doesn't steal a row from the map below it.

As a player with a longer world name, I want the status line to shorten gracefully rather than wrapping — the longer the name, the worse the wrap gets today.

## Evidence

Found by the spec 114 frame audit (2026-08-02). 20 of 132 committed frames emit an over-width line 1 at 80 columns:

- `mid-game__*__80x30.txt:1` — 81 columns (9 frames)
- `scenario__inspect__80x30.txt:1`, `scenario__inspect-solo__80x30.txt:1` — 82 columns
- `scenario__*__80x30.txt:1` — 83 columns (9 frames; the `scenario` fixture's world name is the longest, so it overflows worst)

Header shape at 80x30: `Ashgrove — tick 396000 · day 5 20:00 · running · speed 8x (8.0 t/s) [8 villagers]`

These 20 frames are registered in `knownOverWideFrames` in `cmd/promptworld/framewidth_test.go`, the spec 114 width guard, each tagged TASK-192. That registry is bidirectional: once this is fixed the entries MUST be removed, or the guard fails on the stale allowance.

## Where to look

Spec 114 solved the same class of problem for the map legend and is the pattern to follow: `clipLegend` (`internal/tui/views.go`) clamps to a budget with an `…` marker, built on `ansi.Truncate` for ANSI-safety and display-column measurement. See `specs/114-map-legend-width/research.md` R1-R3 for why that primitive and why the clamp lives at the render call site rather than in `clipLine`.

The open question this card must answer that spec 114 did not: whether the header should truncate the tail like the legend does, or shed lower-value segments first (the `[N villagers]` badge and the `(N t/s)` rate are plausible first casualties, and both are recoverable elsewhere in the UI). The header is denser in distinct facts than the legend, so tail truncation is a weaker default here.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The status/header line never exceeds the terminal width at any committed frame size
- [ ] #2 A shortened header signals that it was shortened
- [ ] #3 All 20 TASK-192 entries are removed from knownOverWideFrames and the spec 114 width guard passes without them
<!-- AC:END -->
