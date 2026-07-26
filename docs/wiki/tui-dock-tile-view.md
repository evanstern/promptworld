---
name: tui-dock-tile-view
description: Split from [[tui-dock-tabs]] — the look-cursor mode's transient TILE-view borrow of the dock body (spec 074, TASK-142): the deliberately-not-a-tab pseudo-label, the dockTabContent short-circuit to tileBody, the tile row hierarchy (agents → piles/chests → structures → terrain → recorded-position events), drill-ins, mouse hit regions, and the borrow's end conditions. Read when touching look.go's tileBody or the dock borrow seam in views.go.
kind: component
sources:
  - internal/tui/look.go
  - internal/tui/views.go
  - internal/tui/tui.go
verified_against: b6a20eaa4da1073a69959a5aff69591d931103a9
---

# TUI dock TILE view (look-cursor borrow)

Split from [[tui-dock-tabs]] (corpus-spec v2 size-budget split,
summary-style): this note covers the look-cursor mode's transient borrow of
the dock body (spec 074-look-cursor, TASK-142). [[tui-map-view]]'s own
section owns the mode's entry/movement/camera; [[tui-input-help]]'s
keymap-doctrine parent covers the mode's key layer; this note covers the
borrow itself, since it lands on the dock [[tui-dock-tabs]] owns.

## The borrow

It is deliberately NOT a sixth tab: no `pane` enum value, no
key digit, no membership in `nextDockTab`/`prevDockTab`'s cycle. `dockTabsRow`
renders every real tab label dim-inactive and appends a highlighted
`TILE (x,y)` pseudo-label in place of the usual active highlight;
`dockTabContent` short-circuits to `tileBody` (`look.go`) before ever
reading `Model.dockTab` — so a tab's own state (chronicle selection/scroll,
villager selection, whatever the previously-active tab was showing)
survives the borrow **by construction**, never merely by careful
bookkeeping: nothing about it changed while borrowed.

## The TILE body

A header (`TILE (x,y) · <tile registry meaning>` plus
warmth/light meter lines derived from `sim.EnvAt` — [[gru]]/
[[executor-world-state]] own the sim-side derivation) followed by rows in
Discworld's fixed hierarchy — agents (needs bars + current intent; the
[[gru]] joins this band non-drillably when abroad on the tile) → piles/
chests (reusing `summarizePileContents`/`describeChest`, [[tui-map-view]]'s
own inspection-line renderers) → structures (registry names, fire lit/
dying/cold, wall damaged+HP, grave) → terrain (one row, the same effective-
kind resolution `tile()` uses) → recent recorded-position events
(`tileEvents`, filtered through the SAME `subjectRegistry` jump-to-source
uses, but keeping the event's own recorded payload coordinates rather than
preferring a live actor position — an event belongs to where it happened,
not wherever its actor has since wandered). `⏎`/`tab` from the map cursor
moves keyboard focus into this pane (amber border,
[[tui-input-help]]'s focus-contract parent, rule 2 — drawn focus, never a
text-capture client); `j`/`k` select a row, `⏎` drills into an agent
(reusing the villager-detail renderer family), an event (reusing
`formatInspector`/`chronicleDetailPane`'s raw-JSON family — the FR-020
"plain language by default, raw behind an explicit drill" boundary), or a
chest/pile's contents. A `tileHitRegion` (the `chronHitRegion` pattern)
records the rendered row-list geometry each frame for the click half of
mouse parity: a row click selects it (acquiring pane focus), a second click
on the already-selected row drills in.

## End conditions

Opening the guardian console (`G`) or a solo zoom (the mode's own digit
exit-and-select) ends the borrow first — it never survives underneath a
body-replacing surface; the help overlay and a takeover layer sit ABOVE it
instead, restoring it unchanged on dismissal.

## Back to parent

[[tui-dock-tabs]] owns the dock this view borrows — its tab bar, guardian
console, and chronicle/guardian/systems tab contents; [[tui-map-view]] owns
the look-cursor's map-side mechanics and the legend/inspection renderers
this body reuses.
