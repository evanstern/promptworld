# Phase 1 Data Model: Map Legend Width Policy

**Spec**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md)

This feature persists nothing and introduces no domain entity. The only "model" that
matters is the **composition of the legend line** — the ordered sequence of segments that
truncation eats from the tail of, and which therefore determines what a player loses first
as the terminal narrows.

## Entity: the composed legend line

Assembled in `renderMapGrid` (`internal/tui/views.go:1546-1548`) from the format string
`"%s · [%d,%d–%d,%d of %d×%d] · %s · %s · %s · %s%s%s"`, then wrapped in `styleDim.Render`
— which is why every downstream operation on it must be ANSI-safe.

| # | Segment | Source | Present when | Survives at 80 cols? |
| --- | --- | --- | --- | --- |
| 1 | Day phase (`night` / `day`) | `phase` | always | yes |
| 2 | Viewport extent (`[x0,y0–x1,y1 of W×H]`) | camera origin + viewport tiles | always | yes |
| 3 | Glyph key (`~water ♠wood "forage …`) | `legendGlyphLine()` (`help.go:59`) | always | partially — this is where the cut lands |
| 4 | Agent-glyph note | `agentGlyphNote` (`help.go:45`) | always | no |
| 5 | Map-control note | `mapControlNote` (`help.go:46`) | always | no |
| 6 | Condition-overlay note | `conditionOverlayNote` (`help.go:53`) | always | no |
| 7 | Stockpile-zone summary | `pilesInfo` — `pileZones` | piles in view | no |
| 8 | Chest inspection entries | `chestsInfo` — `describeChest` | chests in view | no |

Segments 7 and 8 build their own leading ` · ` separator, so they append cleanly or vanish
entirely.

## Ordering is the policy

Because this feature truncates the **tail** (spec Assumptions: the operator approved
clamp-plus-ellipsis, not content reflow), the table's row order *is* the priority order,
and it was not chosen for this feature — it is the order the line has always been composed
in. The consequence, stated plainly so it is reviewable rather than incidental:

- A narrow player keeps orientation (phase, viewport) and part of the terrain key.
- A narrow player loses the prose notes and all inspection detail.

That is a defensible ordering — orientation and terrain outrank prose restatements of
controls documented in the help overlay — but it is an *inherited* one, and it is the
reason segment-priority shedding is named as out-of-scope future work rather than
dismissed. Reordering these segments, or shedding 7–8 before cutting into 3, would change
what the legend *says*; this feature only changes how wide it is.

## Invariants

- **I1** — the legend is exactly one row on every path (FR-008). None of the eight
  segments may introduce a newline.
- **I2** — segments are append-only and independently optional; removing segments 7–8
  never invalidates 1–6.
- **I3** — the composed line always carries ANSI styling, so no segment may be measured or
  cut with rune-based arithmetic (FR-006, FR-007).
- **I4** — content is identical across the narrow and widescreen paths; only the width
  budget differs (`m.width` vs. `cols-4`, per research.md R3). One composition site, two
  presentation budgets.
