---
title: Panel — villager strip
class: panel
status: shipped
verified_against: 2f1044b326023d26fda33a3c70de61f6e6563b0c
sources:
  - internal/tui/views.go
  - internal/tui/layout.go
---

# Panel: villager strip

The village-lens glanceability win (reorientation D12): a one-row,
colonist-bar-style strip under the header giving an at-a-glance read of the
whole roster without opening the villagers tab. **Shipped** (spec 060,
TASK-129) — `villagerStripView`/`villagerStripGlyph` (`internal/tui/views.go`),
wired into the row budget by `computeRows`'s `VillagerStrip` field
(`internal/tui/layout.go`).

## Mockup

```
 promptworld — tick 8801 · day 4 08:12 · running · speed 1x (1.0 t/s)
 12 villagers   A a R † S b K L m N o P
 ┌─ MAP ──────────────────────────────┐ ┌─ chronicle │ … │ villagers ┐
```

**No border** — one plain line, exactly 1 row (`patterns/layout.md`'s row
budget), directly under the header and above the map∥dock body.

## Structure

- **Count** — `N villagers` (living + dead, matching the roster the
  villagers tab shows).
- **Glyph run** — one name-initial glyph per villager, styled exactly as
  the map's agent glyphs already are (`panels/map.md`): bright for awake,
  dim/lowercase for asleep, faint `†` for dead — the SAME priority/style
  rules, so a player who reads the map's glyphs already reads this strip.
  Order is the roster's stable order (matching `panels/villagers.md`'s
  roster, so a position here corresponds to a roster row).
- Width overflow (more villagers than columns allow): drop from the end
  with a trailing `…N` count (`styleDim`, e.g. `…8`), the same "shed, never
  silently truncate a glyph mid-render" discipline every other panel in
  this corpus follows — never a partial/mid-glyph cut.

## Behavior

- **Widescreen default-on, all stages** (`patterns/stage-defaults.md` is
  the authority — this page cites it rather than restating a number): R2
  rules this strip default-on across every stage and pre-ladder in
  widescreen; it is not stage-varying the way the lesson row is.
- **Display-only** — no selection cursor, no drill-down; the villagers tab
  (`panels/villagers.md`) remains the one place to inspect an individual
  villager. This strip answers "how is the village doing right now", never
  "tell me about Ash."
- **Fold behavior** (`patterns/layout.md` ruling a): folds SECOND (after the
  map legend, before the lesson row) under height pressure, to a header
  count badge — `[12 villagers]` — appended to the header row
  (`pages/home.md`).
- **Not carried in narrow** (`patterns/layout.md` ruling b): the narrow
  fallback shows the `[12 villagers]` header-count-badge form only; the
  villagers tab/pane is the narrow drill-down, exactly as it is today.

## Linear-stream / CLI projection (D1)

`promptworld status`/the villagers tab's own data already expose the full
roster with per-agent awake/asleep/dead state — an `attach`/CLI observer
loses nothing this strip shows; the strip is a purely visual glanceability
layer over facts already surfaced elsewhere.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| villager count | N | `replica.Agents` (length, `Model.villCount`) | `villagerStripView` | — (display-only) | reorient D12, spec 060 | — |
| villager glyph run | awake · asleep · dead, per agent | `replica.Agents` | `villagerStripGlyph` | — | reorient D12, spec 060 | — |
| fold to header count badge | shown · folded | fold pressure (`patterns/layout.md`) / narrow (never carried) | `Model.villagerCountBadge`, `computeRows.VillagerStrip` (`patterns/layout.md`) | — | reorient D12, spec 060 | — |

**Parity rollout**: this page is display-only end to end — no actionable
control exists on it, so no parity gap to track.
