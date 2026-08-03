---
id: TASK-191
title: >-
  Map legend overflows the terminal: 355 runes into 80 columns, mid-token cut at
  112
status: In Progress
assignee: []
created_date: '2026-08-03 03:15'
updated_date: '2026-08-03 03:16'
labels:
  - ui
  - tui
  - bug
dependencies: []
priority: high
ordinal: 173001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The map legend — the strip along the bottom of the map naming what every symbol means — is unreadable at the two terminal sizes players actually use. On an 80-column terminal it renders at roughly 4.5x the screen width and soft-wraps into about five rows, shoving the map itself off the top of the screen. On wider terminals it is chopped mid-word after only three symbols, with no marker that anything was cut.

As a player on a standard 80-column laptop terminal, when I open the map, I want the symbol key to fit on screen so it doesn't wrap and push the map out of view.

As a player, when the key is too long to fit, I want a visible sign that it was cut short — otherwise I can't tell whether the village only has three kinds of terrain or whether I'm being shown a fragment.

As a player who widens the window, I want the key to reveal more symbols as room appears, so a bigger terminal is rewarded with more information.

## Evidence (frame harness, spec 112)

31 of 132 committed frames carry lines wider than their declared terminal width. This card covers the legend class (21 frames); the header (11 frames) and scenario-footer (2 frames) overflows are separate findings, not in scope here.

- `docs/design/tui/frames/mid-game__home__80x30.txt:29` — 355 runes into an 80-column terminal.
- `scenario__*__80x30.txt:29` — 356 runes (7 frames, worst case).
- `empty__*__80x30.txt:26` — 354 runes (7 frames).
- `docs/design/tui/frames/mid-game__home__112x30.txt:22` — hard cut mid-token: `night · [21,8–46,24 of 64×64] · ~water ♠wood "forage` — three glyphs of roughly twenty, no ellipsis.

Distinct from the spec 112 FR-008 narrow-fallback caveat: that caveat blesses an 80x30 frame having FEWER ROWS than 30 (no fold arithmetic below the breakpoint). It says nothing about column overflow. These are horizontal, and they are bugs.

## Diagnosis (complete, file:line)

- `internal/tui/views.go:1750` — the narrow fallback returns `styleBox.Render(grid) + "\n" + legend`. The legend is concatenated OUTSIDE the box with no `clipLine`/`clipContent` and no `.Width()`, so the raw legend reaches the terminal unclamped. The comment at `views.go:1500` asserts "clipContent already clips an over-wide legend" — true on the widescreen path (`views.go:937`), false here. That stale assumption is the bug.
- `internal/tui/views.go:1784` — `clipLine` crops via `lipgloss.NewStyle().MaxWidth(width).Render(s)`, a hard crop with no ellipsis, producing the mid-token cut on the widescreen path.

## Open design question for the spec

`truncateRunes` (`internal/tui/digest.go:94`) is the package's idiomatic `…` truncator but is rune-based; the legend carries ANSI from `styleDim.Render`, which is why `clipLine` uses lipgloss `MaxWidth`. The fix needs an ANSI-safe ellipsis clip. Whether that lands in `clipLine` globally (about 18 call sites, many goldens rewritten) or as a legend-local helper is the spec's call. The chronicle already truncates with its own `…` (`grammar.go:284`, `grammar.go:554`), which is precedent for per-surface handling.

Interacts with the documented shed order at `docs/design/tui/panels/map.md:130` ("drop the legend before shrinking the viewport when rows get scarce") and the spec 060 deferral at `map.md:209`, which named "legend-line overflow pain" as its own re-open trigger. This card is that trigger firing.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 At every committed frame size, the map legend line never exceeds the declared terminal width (80/112/113/160)
- [ ] #2 The 80x30 narrow-fallback legend is clamped at the render site, not left to the terminal to wrap
- [ ] #3 A legend truncated for width ends in a visible ellipsis marker, never a bare mid-token cut
- [ ] #4 Widening the terminal reveals strictly more legend content
- [ ] #5 docs/design/tui/panels/map.md records the width policy and resolves the spec-060 legend-overflow deferral at map.md:209
- [ ] #6 A regression guard fails if any committed frame emits a line wider than its declared width
<!-- AC:END -->
