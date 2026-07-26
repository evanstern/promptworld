---
title: Panel — dock (tab container)
class: panel
status: shipped
verified_against: 6e83f579db2b448c9c59b15575bf564b1e9b1852
sources:
  - internal/tui/tui.go
  - internal/tui/views.go
  - internal/tui/look.go
---

# Panel: dock

The right-hand tab **container** in the widescreen composite — chrome only.
Since this feature (D10, the skin-boundary-as-file-boundary ruling), tab
*content* is documented on its own per-tab page, not here:

- [guardian.md](guardian.md) — fiction-layer tab content (transcript, standing
  orders, instruction surface, working feedback)
- [systems.md](systems.md) — engine telemetry (never skinned)
- [villagers.md](villagers.md) — the villagers roster/detail/decisions tab
- [exercise.md](exercise.md) — scenario exercise progress (spec 054):
  **conditional** — the row gains this tab only when the attached world
  carries a `Manifest.Scenario` block; ambient worlds keep the 4-tab row
  (chronicle · guardian · villagers · systems) byte-identically

This page owns only the tab row, its badges, tab-switching, and the solo-zoom
seam — the same container regardless of which tab is active.

## Mockup

```
┌─ chronicle │ {{skin.guardian.tab_label}} │ villagers │ systems ─┐   ← tab row doubles as the panel title
├───────────────────────────────────────────────────────────────────┤
│                                                                   │
│  active tab content — see the owning tab page                    │
│                                                                   │
└───────────────────────────────────────────────────────────────────┘
```

- Tab row: active tab bright and UPPERCASE, inactive dim. The guardian tab
  shows an unseen-reply badge dot (`{{skin.guardian.tab_label}} •`) whenever
  its tab isn't currently visible and a reply has arrived — the systems tab
  never carries this badge (D10: "no second badge system"; it carries no
  conversational content to badge).
- Keys `2` chronicle · `3` guardian · `4` villagers · `5` systems · `6`
  exercise (scenario worlds only — spec 054) select tabs; the **same key
  again**, while already selected, zooms that tab to full width — solo zoom
  ([../pages/solo-views.md](../pages/solo-views.md)).
- Each tab keeps its own state (scroll, filters, selection, input history)
  across switches — one dock-content renderer dispatches by active tab
  (`dockTabContent`), so a tab never re-mounts on selection.
- Adding a future tab = a new label in the row + a content renderer; no new
  layout — spec 053's systems tab (key `5`) proved the pattern, and spec
  054's exercise tab proved its conditional form: an entry in `dockTabsRow`
  present only when the world carries a scenario + `exerciseBody`, with
  `tab`/`shift+tab` cycling (`nextDockTab`/`prevDockTab`, now Model methods)
  including it exactly when present.

## Borrow seam: the TILE view (spec 074-look-cursor)

While the look-cursor mode ([../patterns/keymap.md](../patterns/keymap.md)
"Mode: look-cursor") is active, the dock body is **borrowed** by a transient
**TILE view** — the tile the map cursor sits on, in DF's fixed hierarchy
(agents → piles/chests → structures → terrain → recent events), a header
carrying the tile's registry whatis prose plus warmth/light levels
(`internal/tui/tiles.go`'s `meaning` rows, spec 068). This is the same
"one component, two widths" borrow shape the solo zoom already uses, applied
along a different axis (content, not width):

- **Not a tab**: no `pane` enum value, no key digit (`2`–`6` exit the mode
  and select an actual tab instead of navigating to this), no membership in
  `nextDockTab`/`prevDockTab`'s cycle.
- **`m.dockTab` is never written** while borrowed — `dockTabsRow` renders
  every normal label dim-inactive and appends a highlighted `TILE (x,y)`
  pseudo-label in place of the usual active highlight; `dockTabContent`
  short-circuits to the TILE renderer before ever reading `m.dockTab`. The
  underlying tab's state (chronicle selection/scroll, villager selection,
  scroll offsets — this page's own "each tab keeps its own state" rule)
  therefore survives **by construction**: nothing about it ever changed,
  there is nothing to restore.
- **State preservation, concretely**: entering the mode while paused with
  the chronicle selected and inspecting an event leaves that selection and
  detail-pane scroll exactly where they were; exiting the mode (`esc`/`v`)
  or pressing a tab digit (which exits first, then selects) shows the
  chronicle exactly as it was left, inspect mode resuming automatically
  (its own `chronicleVisible()`-gated entry, unaffected by this feature).
- **Focus, drawn**: `⏎`/`tab` from the map cursor moves keyboard focus into
  the TILE pane — its border renders amber (the `panelFocus` token,
  [../patterns/focus-contract.md](../patterns/focus-contract.md) rule 2,
  the same token the guardian console's own focused sub-panels use);
  `j`/`k` select a row, `⏎` drills into an agent's detail (the villager-
  detail renderer family), an event's raw-JSON inspector (`formatInspector`/
  `chronicleDetailPane` — the FR-020 boundary, `panels/chronicle.md`), or a
  chest/pile's contents.
- **Covered by a body-replacing surface, the mode exits first**: opening the
  guardian console (`G`) or a solo zoom (the exit-and-select digits) ends
  the borrow before the new surface opens — it never survives underneath
  one. The help overlay and a takeover layer ABOVE the borrow instead
  (restoring it on dismissal by construction, same as every other mode).

## Behavior

- **Default tab on launch:** chronicle ([chronicle.md](chronicle.md)).
- **Reply arrival while the guardian tab isn't visible:** the tab row badges
  (`{{skin.guardian.tab_label}} •`); the dock never steals the currently
  selected tab out from under the player (guardian.md covers the transcript
  side of this).
- **Solo zoom**, narrow fallback, and cross-width rendering ("one component,
  two widths") are specified in
  [../pages/solo-views.md](../pages/solo-views.md); this page only owns the
  seam (same-key-twice) that reaches it.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| tab row | active · inactive | `Model.dockTab` | `dockTabsRow` | `2`/`3`/`4`/`5` (+`6` on scenario worlds) select · — | TASK-34 | — |
| chronicle tab label | active · inactive | `Model.dockTab` | `dockTabsRow` | `2` · — | TASK-34 | — |
| guardian tab label | active · inactive · unseen-badge | `Model.dockTab`, `Model.guardianUnseen` | `dockTabsRow` | `3` · — | TASK-34 | `skin.guardian.tab_label` |
| villagers tab label | active · inactive | `Model.dockTab` | `dockTabsRow` | `4` · — | spec 015 | — |
| systems tab label | active · inactive | `Model.dockTab` | `dockTabsRow` | `5` · — | spec 053 | — |
| exercise tab label (scenario worlds only) | absent · active · inactive | `Model.dockTab`, `Model.exerciseID` | `dockTabsRow` | `6` (inert on ambient worlds) · — | spec 054 | — |
| unseen-reply badge dot | shown · hidden | `Model.guardianUnseen` | `dockTabsRow` | — (display-only) | TASK-34 | — |
| tab-switch → solo zoom | home,tab=k · solo(k) | `Model.solo`, `Model.dockTab` | `selectTab`, `widescreenView` | same key twice · — | TASK-34 | — |
| solo → home / switch | solo(k) · home | `Model.solo` | `selectTab` | `1`/`esc` (home) · a different tab key (switch) · — | TASK-34 | — |
| dock panel chrome | dormant border | — (static box) | `dockPanelView` | — | TASK-34 | — |
| TILE view pseudo-label | shown (borrow active) | `Model.lookActive`, `lookX`/`lookY` | `dockTabsRow` (`lookDockTabsRow`) | — (display-only; entry is the map's own `v`/click) | spec 074 | — |
| TILE pane row select/drill | none selected · selected · drilled | `Model.lookSel`, `lookFocus`, `lookDrill` | `tileBody` (`look.go`) | `j`/`k` select · `⏎` drill · click a row selects, a second click drills | spec 074 | — |

**Parity rollout**: the pre-existing tab row/switch/badge controls above
still have no mouse target — the whole dock (`internal/tui`) predates the
input-parity doctrine (decision 8), and this feature didn't touch them, so
it didn't close that gap either. The TILE pane's row click is a NEW control
this feature adds, and it ships with both a key and a mouse target together
(decision 8 rule 1) — the second leg of spec 074's click-tile/click-row
parity landing (`panels/map.md` is the first). The formal doctrine and
rollout plan live in `patterns/keymap.md`.
