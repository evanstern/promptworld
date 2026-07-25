# Data Model: Guardian strip (spec 050)

No persistent data, no wire changes. Client-side render values only.

## Strip segments (per-frame render values)

| Segment | Source | Present when | Rendered form |
|---|---|---|---|
| charge bank | `status.Clock.MetatronCharges`, `sim.MetatronChargeCap` | status snapshot exists | `⚡…·… (N/cap)` |
| regen forecast | current tick + regen cadence (R2 accessor) | status exists AND charges < cap | `next +1 @ <game time>` |
| standing orders | `len(status.Orders)` | status snapshot exists | `👁 N standing orders` (0 is a true value) |
| faith | TASK-118 future field | never (this feature) | — omitted entirely |

**Validation**: presence rules are exactly spec FR-003/research R4; the
fixture sweep in `render_test.go` enumerates the presence matrix (no status /
partial bank / full bank / 0 orders / N orders) and asserts absent segments
render nothing (no dashes, no zeros for missing mechanics).

## rowBudget (extended)

`internal/tui/layout.go` `rowBudget` gains `Strip int` (0 or 1). Invariant:
`Header + Strip + Body + Minibuffer + Footer == totalRows` whenever
`totalRows` covers the fixed chrome; `Strip` drops to 0 only at the ruled
fold threshold (layout.md ruling a step 4), and body never goes negative.

**State transitions**: pure function of `(totalRows, stripVisible)` — no
persistent state; recomputed per `tea.WindowSizeMsg` as today.

## Relocated dormant line

When `Strip == 0`, `minibufferView`'s dormant branch composes:
`<bank glyphs> · 👁N · <existing dormant placeholder>` — truncated to width,
dormant state only (focused/busy branches untouched).
