---
id: TASK-193
title: 'Scenario keybind footer overflows at 112/113 columns: 121 columns, 2 frames'
status: To Do
assignee: []
created_date: '2026-08-03 03:46'
labels:
  - ui
  - tui
  - bug
dependencies: []
ordinal: 175001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The row of keyboard shortcuts along the bottom of the screen runs off the edge in scenario worlds, which carry extra lesson keybinds that ordinary worlds don't. It wraps onto a second line at the exact widths a scenario is most likely to be played at.

As a learner working through a scenario, when I glance at the shortcut row to find the key I need, I want it to sit on one line instead of wrapping and shifting the layout under it.

As a learner, when there are more shortcuts than fit, I want to see that the row was shortened rather than assume I've been shown all of them.

## Evidence

Found by the spec 114 frame audit (2026-08-02). 2 of 132 committed frames:

- `scenario__home__112x30.txt:30` — 121 columns into 112
- `scenario__home__113x30.txt:30` — 121 columns into 113

Only the `scenario` fixture's `home` state, and only at the two breakpoint widths — the 160x50 frame is roomy enough to fit it, and the 80x30 narrow fallback renders a shorter footer variant. That narrow band is why it went unnoticed.

Both frames are registered in `knownOverWideFrames` in `cmd/promptworld/framewidth_test.go`, the spec 114 width guard, tagged TASK-193. The registry is bidirectional: once this is fixed the entries MUST be removed, or the guard fails on the stale allowance.

## Where to look

Smallest of the three over-width classes the audit found (legend 21 frames, fixed by spec 114; header 20 frames, TASK-192; this, 2 frames). Overflow is 8-9 columns, so a clamp may be all it needs.

Worth checking first whether the scenario footer should instead drop the least-used binding at these widths — unlike the legend, every entry in this row is a key the player might press, so truncating the tail silently removes a discoverable action rather than a description. That argues for shedding a named low-value binding over a blind tail cut. Follow spec 114's `clipLegend` (`internal/tui/views.go`) for the ANSI-safe, display-column-aware mechanics either way.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The scenario keybind footer never exceeds the terminal width at any committed frame size
- [ ] #2 No keybinding is silently removed without a visible signal
- [ ] #3 Both TASK-193 entries are removed from knownOverWideFrames and the spec 114 width guard passes without them
<!-- AC:END -->
