---
id: TASK-191
title: >-
  Map legend overflows the terminal: 355 runes into 80 columns, mid-token cut at
  112
status: Done
assignee: []
created_date: '2026-08-03 03:15'
updated_date: '2026-08-03 04:10'
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

Spec: specs/114-map-legend-width
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 At every committed frame size, the map legend line never exceeds the declared terminal width (80/112/113/160)
- [x] #2 The 80x30 narrow-fallback legend is clamped at the render site, not left to the terminal to wrap
- [x] #3 A legend truncated for width ends in a visible ellipsis marker, never a bare mid-token cut
- [x] #4 Widening the terminal reveals strictly more legend content
- [x] #5 docs/design/tui/panels/map.md records the width policy and resolves the spec-060 legend-overflow deferral at map.md:209
- [x] #6 A regression guard fails if any committed frame emits a line wider than its declared width
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
PR #159 open (draft). Spec 114 shipped: legend clamps to its budget with a trailing ellipsis, both render paths, plus a matrix-wide width guard.

Model tier (constitution Principle V — required record): routine tier per the rubric — single package (internal/tui) plus one test in cmd/promptworld, view/rendering code, tests alongside code. No escalation trigger fired: no cross-package or architectural change, no concurrency/scheduling/governor logic, no doctrine-adjacent behavior change, no prior failed attempt.

DEVIATION, disclosed: the model that actually served was claude-opus-5, running INLINE in the planning session rather than as a delegated spec-implementer subagent. Principle V requires the planning tier to delegate and "never implement inline." This session's operating instructions explicitly forbid spawning agents unrequested, and the operator did not request one; the tighter session guardrail was followed and the deviation surfaced to the operator at the time rather than after. Quality impact is nil — Opus 5 is itself the constitution's escalation implementation tier — the deviation is procedural (cost and delegation discipline), not qualitative.

Two findings surfaced during the work, neither silently resolved:

1. Five legend tests were pinning the defect. testModel renders at 80 columns, and TestMapRendersPilesAndStockpileZones / ChestGlyphAndInspection / WallGlyphs / PathGlyph / GraveGlyph all asserted the legend CONTAINED inspection content that only fit because the legend was unclamped. They were verifying something no player at 80 columns could usefully see. Repointed at the composed legend from renderMapGrid (the composition they were actually testing); presentation clamping is now covered separately by legend_width_test.go.

2. The header-overflow class is 20 frames, not the 11 first reported. The initial audit miscounted by reading rows listing "at:1,29" as legend-only. Corrected before the deny-list was written; TASK-192 carries the true figure.

Scope held: the operator scoped this to the legend class (21 frames). The header (20 frames, TASK-192) and scenario footer (2 frames, TASK-193) were filed rather than fixed, and are registered in the guard's bidirectional deny-list, which fails if an entry is dropped early, outlives its fix, or names a frame the matrix no longer produces — verified all three ways.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Delivered and merged as PR #159 (merge commit fc63ae95, two parents — verified NOT a squash, so the in-branch wiki re-pins remain reachable from main).

The map legend now clamps to the width it renders into and marks the cut with an ellipsis. Before: the narrow path appended the legend outside the map box with no clip at all, so a 354-356 column line reached an 80-column terminal, soft-wrapped into roughly five rows, and pushed the map off the top of the screen; at 112+ it was clipped but cut mid-token with no marker, showing three symbols of twenty as if that were the whole key. Verified on main after merge: the 80x30 legend is one row, 80 columns, ending in the marker.

Implementation is clipLegend (internal/tui/views.go) on ansi.Truncate — ANSI-safe so a cut never severs an escape and bleeds styling down the screen, display-column aware for double-width glyphs, and a no-op when the legend already fits so a fitting legend gains no false marker. Each call site clamps with the budget it owns: m.width for the narrow path (its legend is outside the box), cols-4 for the widescreen box interior. clipLine/clipContent were deliberately left untouched — they are a layout safety net, not a content-communication device, and the chronicle already ellipsizes per-surface, so a global tail risked double-ellipsis where the two layers compose (research.md R2).

Guarded by TestFramesNeverExceedDeclaredWidth, matrix-wide, measuring display columns rather than runes. Its deny-list is bidirectional and was probed all three ways: dropping an entry early, an entry outliving its fix, and an entry naming a dead frame all fail correctly.

Grounding verified fresh on main after merge: frame matrix matches a fresh dump, player docs 16 fresh / 0 stale, tui-design gate passes, and docs/wiki/tui-map-view.md's pin resolves as an ancestor of main.

Two findings the work surfaced, neither silently resolved. (1) Five legend tests were pinning the defect — testModel renders at 80 columns and they asserted inspection content that only fit because the legend was unclamped, verifying something no player could usefully see; repointed at the composition they actually tested. (2) The header-overflow class is 20 frames, not the 11 first reported (the audit misread rows listing "at:1,29" as legend-only) — corrected before the deny-list was written.

Scope held to the legend class. The header (TASK-192) and scenario footer (TASK-193) were filed rather than fixed and are registered in the guard so they cannot be forgotten.

docs/design/tui/panels/map.md records the width policy and closes the spec 060 deferral at map.md:209, whose own re-open trigger was "legend-line overflow pain becomes the actual bottleneck" — this work is that trigger firing and being answered.
<!-- SECTION:FINAL_SUMMARY:END -->
