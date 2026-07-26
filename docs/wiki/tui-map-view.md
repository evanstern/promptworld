---
name: tui-map-view
description: The map camera region: terrain/agent/structure glyph rendering (water, woods, marsh/sand, walls, graves, piles) resolved through the tile registry, condition overlays, camera pan and jump-to-source, the look-cursor tile-inspection mode, and the legend's stockpile-zone/chest inspection additions. Split from [[tui-client]]; read when touching views.go's renderMapGrid, tiles.go, look.go, or digest.go's subjectRegistry.
kind: component
sources:
  - internal/tui/views.go
  - internal/tui/tiles.go
  - internal/tui/digest.go
  - internal/tui/tui.go
verified_against: 0fd2104c59c54be8e8071d319fa4ce192083faf3
---

# TUI map view

Split from [[tui-client]] (corpus-spec v2 size-budget split, summary-style):
this note covers the map camera region — terrain and overlay rendering, camera
control, and the legend's inspection content. See [[tui-client]] for the dock
tabs, chronicle feed, and input/help overlay.

## The map region, camera, and jump-to-source

Regions: the **map** is a camera window over the generated terrain from
`Model.gameMap` (regenerated locally via `world.Map()`,
[[worldmap-generation]]): water ~, wood ♠, forage ", rock outcrops ^, dens
ᴥ, and — since spec 068, on a world generated with the marsh/sand pass — marsh
░ and sand ▒ glyphs, plus dynamic overlay state read off the replica (never
part of the
static tile) — a quarried-out rock outcrop renders as a faint `,` ahead of the
static terrain check — with the replica's agents on top (by initial,
lowercase asleep, † dead — since spec 060, [[village-lens]], a living
agent's STYLE additionally carries two condition overlays over the same
glyph: needs-critical when Health/Food/Warmth/Rest crosses its existing
danger band, else suppressed-mind when its latest decision trace is a
router suppression, needs-critical winning when both apply; neither
condition changes what glyph or case renders, only its style) plus built
structures: fires render lit ▲ while the
current tick is before the structure's `FuelUntil` — since spec 060 a THIRD,
still-lit "dying" style applies inside `State.RefuelDyingBelow()`'s window
before that, distinct from both plain lit and cold — and fall back to a faint,
hollow cold glyph △ once fuel runs out, shelters ⌂, ovens ▣, chests ☐ (spec
013 US3), and the [[gru]] as a red G while it is abroad; ground piles (spec
013 US2, `Model.replica.Piles`) render as a dedicated overlay `%`, layered
like structures rather than folded into them so a coincidental tile overlap
loses neither glyph's priority silently. Since spec 032, a wall structure
(`wall_plank`/`wall_stone`) renders as a solid barrier glyph — `▤` plank, `▩`
stone — dim (`styleWallDamaged`) whenever its `HP` is below `sim.WallMaxHP`,
the same faded-glyph treatment as a burnt-out fire, so a wall under
demolition reads at a glance; a path structure renders at TERRAIN level
(below agents/structures/piles, its own `paths` set rather than the
structures map) as a warm-tan `·` distinct from plain grass's dim `·`, so an
agent or a dropped pile standing on a path tile still shows through. Since
spec 044 (US4), a `grave` structure renders as `✝` in faint gray
(`styleGrave`, the cold-fire precedent for spent/inert glyphs), recorded
both in the structures map and in a dedicated `graves` set that the agents
loop consults for one deliberate priority carve-out: a dead agent standing
on a tile that also holds a grave renders the grave glyph instead of the
plain dead marker — the body becomes the grave — because every post-044
death places its grave at the dead agent's own frozen tile, so the usual
agent-over-structure priority would otherwise permanently hide the glyph. A
graveless dead agent (pre-044 replay/history) still renders the plain `†`.
Since spec 068 ([[tile-registry]]), every glyph and style named in this
paragraph — including `styleWallDamaged`/`styleGrave` above — resolves
through the tile registry's classed style tokens (`internal/tui/tiles.go`)
rather than a per-tile literal: `views.go`'s old style-literal block is gone,
`tile()` now calls `tileKey(...).render(state)`, and the named styles are
token-derived aliases kept for direct call sites — the glyphs, colors, and
priority described here are unchanged bytes, only their home moved.
The camera follows the living agents' centroid, arrow keys pan, `c` recenters.
Since spec 049 (TASK-124, reorient D3) the camera gains one computed writer:
**jump-to-source** — in inspect mode, `⏎` (or a mouse click on a chronicle
list line; `tea.WithMouseCellMotion` is enabled in `cmdUI` and `handleMouse`
binds ONLY this control, everything else falling through inert) resolves the
selected event's subject (`resolveSubject` + a per-type `subjectRegistry` in
digest.go: primary actor's live replica position if alive, else explicit
payload coordinates, else unlocatable; `world.migrated` is never decoded) and
sets the pan via `centerCameraOn` (`pan = target − wandererCentroid()`), so a
jump IS a pan — same clamps, same `c` recovery. The detail pane's bottom row
is now a permanent actions bar (`detailActions`, exactly one action per
event): the jump affordance `⏎ jump to <name> (x,y)` or the honest
`no location for this event`; a catalog sweep asserts jump-or-hint totality.
In the narrow fallback a successful jump lands on the map pane with the
paused selection preserved. Click hit-testing reads a per-frame
`chronHitRegion` the inspect-list renderer records — running-clock,
out-of-region, help-open, and minibuffer-focused clicks are all no-ops.

## Look-cursor mode (spec 074-look-cursor, TASK-142)

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

`hjkl`/arrows move the cursor one tile, `H`/`J`/`K`/`L` jump 8 (clamped to
`[0,W)×[0,H)`, never wrapping); moving pushes the camera (`panX`/`panY`) by
whatever overshoot keeps the cursor at least 2 tiles inside the current
viewport, degrading at the world edge (the camera stops, the cursor may
reach the viewport border) — `c` snaps the camera onto the cursor
(`centerCameraOn`, the identical jump-to-source formula, one more caller).
Exiting (`esc`/`v` again, or a dock-tab digit) resets `panX`/`panY` to 0 —
"resume centroid-following" is literally the pre-existing following state.

**Rendering**: the cursor tile gets a background-highlight style transform
(`styleLookCursor`, `lipgloss.Reverse(true)`) over whatever glyph `tile()`
resolves — the `styleFeedSelect` precedent, never a new glyph (spec-068
FR-003 discipline extended); mode-off rendering never reaches this branch,
so `TestTilesIdentityPin`'s goldens are untouched. The map panel's title
gains a third state, `MAP · cursor (x,y) · c center · esc exit`, replacing
the following/panned title while active.

**Borrowing the dock**: while active, the mode's own key layer
(`handleLookKey`, layered between the console and inspect checks in
`handleKey` — [[tui-input-help]]) claims the whole contested key set and the
dock body is borrowed by a transient TILE view — [[tui-dock-tabs]] owns
that half. `mapHitRegion` (a `chronHitRegion`-shaped per-frame pointer,
recorded by `mapPanelView`/`mapView` each render, invalidated by default)
is the click-geometry half of mouse parity: a release inside it maps
`(x0 + (col)/2, y0 + row)` back to a world tile (2 screen columns per tile,
the existing glyph+space stride) and either enters the mode there or moves
an already-active cursor.

## Inspection: legend additions

Inspection (spec 013 T021/T026, SC-006): the map legend — its one designated
inspection surface, content grows the line rather than adding a second row —
appends, for whatever's currently in view, a stockpile-zone summary per pile
cluster and an owner+contents+fullness entry per chest. Piles in view are
grouped into **stockpile zones** by 4-neighbor Manhattan adjacency
(`pileZones`, a render-side-only flood fill — no zone state, matching
spec.md's "an observability grouping of adjacent piles, not a state entity");
each zone renders as `pile(x,y) contents` (single pile) or
`zone[n](x0,y0)-(x1,y1) contents` (multi-pile, bounding box + count), where
contents (`summarizePileContents`) is non-food resource counts plus a spear
count plus a `food Nr/Nc/Nm` batch total when any food is held. Each visible
chest renders as `chest(x,y) [Owner] contents n/48` (`describeChest`, owner
resolved through the same `agentName` helper the chronicle grammar uses,
contents via `summarizeInventoryContents`, capacity `sim.ChestCap`) — a
chest's `Store` is a plain counts inventory rather than dated batches,
because chests preserve food indefinitely (no rot deadlines to track).

## Back to parent

[[tui-client]] links here for the map region; that note's own Connections
section lists [[tile-registry]], [[worldmap-generation]], and [[village-lens]]
as the map's underlying data sources.
