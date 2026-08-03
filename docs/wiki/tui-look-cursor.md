---
name: tui-look-cursor
description: Split from [[tui-map-view]] — the spec-074 look-cursor tile-inspection mode's map side (TASK-142): the v-toggled second camera writer, cameraOrigin/mapViewportDims extracted helpers, hjkl/HJKL cursor movement with camera push, the reverse-video cursor rendering, the mode's key layer, and mapHitRegion mouse parity. The dock-body TILE view it feeds is [[tui-dock-tile-view]]. Read when touching look.go's cursor/camera math or handleLookKey.
kind: component
sources:
  - internal/tui/look.go
  - internal/tui/views.go
  - internal/tui/tui.go
verified_against: 9f7df6137c78506f9d5ab48809f6c2e4855da782
---

# TUI look-cursor mode (map side)

Split from [[tui-map-view]] (corpus-spec v2 size-budget split,
summary-style): this note covers the look-cursor mode's map-side mechanics
(spec 074-look-cursor, TASK-142) — entry, movement, camera interplay,
rendering, and mouse parity. The transient TILE view it projects into the
dock body is [[tui-dock-tile-view]]; [[tui-input-help]]'s keymap doctrine
covers where its key layer sits.

## Entry and the extracted camera helpers

`v` toggles a second, independent camera writer over the same
`wandererCentroid`+pan substrate jump-to-source uses: a **look-cursor**
tile, entered at the camera-center tile (or, via a map-tile click, at the
clicked tile — the map's first real mouse target, decision 8 rule 1). Two
extracted helpers keep every consumer honest against the same numbers
`renderMapGrid` draws: `cameraOrigin(vw, vh)` (the world-space top-left tile
for a viewport, wanderer centroid + pan, clamped to the map — literally
`renderMapGrid`'s own inline math, pulled out) and `mapViewportDims()` (a
second, independent read of whichever viewport formula the live layout
uses — widescreen's column-budget math or narrow's `vw/vh` formula — so the
key handler and mouse hit-testing need no cached geometry).

## Movement and the camera

`hjkl`/arrows move the cursor one tile, `H`/`J`/`K`/`L` jump 8 (clamped to
`[0,W)×[0,H)`, never wrapping); moving pushes the camera (`panX`/`panY`) by
whatever overshoot keeps the cursor at least 2 tiles inside the current
viewport, degrading at the world edge (the camera stops, the cursor may
reach the viewport border) — `c` snaps the camera onto the cursor
(`centerCameraOn`, the identical jump-to-source formula, one more caller).
Exiting (`esc`/`v` again, or a dock-tab digit) resets `panX`/`panY` to 0 —
"resume centroid-following" is literally the pre-existing following state.

## Rendering

The cursor tile gets a background-highlight style transform
(`styleLookCursor`, `lipgloss.Reverse(true)`) over whatever glyph `tile()`
resolves — the `styleFeedSelect` precedent, never a new glyph (spec-068
FR-003 discipline extended); mode-off rendering never reaches this branch,
so `TestTilesIdentityPin`'s goldens are untouched. The map panel's title
gains a third state, `MAP · cursor (x,y) · c center · esc exit`, replacing
the following/panned title while active.

## Key layer and mouse parity

While active, the mode's own key layer (`handleLookKey`, layered between
the console and inspect checks in `handleKey` — [[tui-input-help]]) claims
the whole contested key set and the dock body is borrowed by the transient
TILE view — [[tui-dock-tile-view]] owns that half. `mapHitRegion` (a
`chronHitRegion`-shaped per-frame pointer, recorded by
`mapPanelView`/`mapView` each render, invalidated by default) is the
click-geometry half of mouse parity: a release inside it maps
`(x0 + (col)/2, y0 + row)` back to a world tile (2 screen columns per tile,
the existing glyph+space stride) and either enters the mode there or moves
an already-active cursor.

## Back to parent

[[tui-map-view]] owns the map region this mode rides — glyph rendering,
camera following, jump-to-source, and the legend's inspection additions.
