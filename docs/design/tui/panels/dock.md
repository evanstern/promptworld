---
title: Panel — dock (tab container)
class: panel
status: shipped
verified_against: 9905c70f326d027a34425081442270bdb80648f7
sources:
  - internal/tui/tui.go
  - internal/tui/views.go
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

**Parity rollout**: no control on this page has a mouse target today — the
whole dock (`internal/tui`) predates the input-parity doctrine (decision 8).
Every keyed row above is a parity gap tracked here honestly rather than
silently omitted; the formal doctrine and rollout plan land in
`patterns/keymap.md`.
