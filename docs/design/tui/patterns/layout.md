---
title: Pattern — layout
class: pattern
status: shipped
verified_against: 7e3c2b5f5f23eb8e5fcb37d0f867dbc6f46a289b
sources:
  - internal/tui/layout.go
  - internal/tui/views.go
---

# Pattern: layout

Breakpoints, width math, row budget, and style tokens for the widescreen
composite. Inputs are `tea.WindowSizeMsg` width/height, as today. This
feature (Wave 0 rulings a/b) re-derives the row budget and fold order for
the new permanent chrome (villager strip, lesson row, guardian strip); the
breakpoint, column budget, and style tokens are unchanged and preserved
verbatim below.

## Breakpoint

```
width ≥ 112 cols  →  widescreen composite (pages/home.md)
width <  112 cols →  narrow fallback: today's single-pane UI, unchanged
```

112 is the narrowest width where an even 50/50 split (see "Column budget"
below) still leaves both the map and the dock genuinely usable — below it,
halving the terminal would starve both regions rather than either one, so
the narrow single-pane fallback takes over instead of shrinking further.
(Implementation note, TASK-34: two earlier drafts of this section described
a fixed-44-col dock with the map holding a 73-col target as the terminal
narrowed toward the breakpoint; a planning decision during implementation
replaced that with a straight 50/50 split — see "Column budget". The
breakpoint value itself (112) did not need to change: at 112 the split gives
map=56 / dock=55, both still comfortably usable.)
Resizing across the breakpoint swaps layouts live without losing state.
Height has no breakpoint: rows get scarce → panels shed their lowest-priority
rows (map legend first) and, since this feature, the new permanent chrome
rows fold in the ruled order below (ruling a) once the map legend alone
isn't enough.

## Column budget (widescreen)

```
totalCols
├─ gutter: 1 col
├─ dock:   (totalCols − gutter) / 2            (floors down on an odd split)
└─ map:    totalCols − gutter − dock           (takes the odd leftover column;
                                                 viewport tiles = (mapCols − 4) / 2 —
                                                 2 cols border + 2 cols padding)
```

Map and dock split the terminal 50/50 — a planning decision (TASK-34)
superseding an earlier fixed-44-col dock. The map takes the extra column
when `totalCols − gutter` is odd, so it is never smaller than the dock.

## Row budget (widescreen) — re-derived, Wave 0 ruling (a)

The pre-reorientation budget was 5 fixed rows (header 1 · body remainder ·
minibuffer 3 · footer 1). This feature adds three new permanent chrome rows
at their stage-1–2 defaults (`patterns/stage-defaults.md`): the villager
strip (D12), the lesson row (decision 5), and the guardian strip
(decision 7). At full stage-1–2 defaults:

```
totalRows
├─ header:         1
├─ villager strip:  1   (D12; widescreen default-on, all stages —
│                        patterns/stage-defaults.md)
├─ body:           remainder (map ∥ dock; both full body height)
├─ lesson row:     2   (decision 5; default-on at stages 1–2 only —
│                       0 rows when stage-defaulted off; ≤2 lines, fixed 2
│                       when on; borderless — a plain 2-line block, unlike
│                       the bordered minibuffer/guardian-strip, to fit
│                       exactly this 2-row budget)
├─ guardian strip:  1   (decision 7; always-on, all stages)
├─ minibuffer:      3   (unchanged; bordered single line)
└─ footer:          1
```

Fixed chrome at stages 1–2: **9 rows** (was 5) → `body = totalRows − 9`; a
30-row terminal keeps 21 body rows, a 24-row terminal 15. At stage 3+ or
pre-ladder-defaulted-off, the lesson row starts already folded (badge form,
0 rows), so fixed chrome there is **7 rows** by default before any fold
pressure applies.

**Reconciliation note (spec 060, TASK-129)**: this section's full stage-1–2
target is now the shipped reality — all three new chrome rows exist
(guardian strip: spec 050; lesson row: spec 055/TASK-117; villager strip:
spec 060/TASK-129), and `computeRows` (`internal/tui/layout.go`) implements
the complete four-field row budget (`Header`, `VillagerStrip`, `Lesson`,
`Strip` [guardian], `Body`, `Minibuffer`, `Footer`) folding all three in the
ruled order below. Nothing about the ruled total order changed when the
villager strip landed — it was always going to fold second, before the
lesson row.

### Fold order (a total order over the collapsible rows)

Chrome folds when body rows would drop below `bodyMin = 10`, reclaiming in
this total order, one step at a time, until body ≥ 10 or the floor is
reached:

1. **map legend** (existing body-internal shed — stays first; unchanged
   from pre-reorientation).
2. **villager strip → folds into a header count badge** — e.g.
   `[12 villagers]`, appended to the header row (`pages/home.md`).
   **Shipped** (spec 060, TASK-129): `computeRows`'s `VillagerStrip` field
   drops to 0 at exactly this threshold, and `Model.villagerCountBadge`
   (`internal/tui/views.go`) composes the header segment.
3. **lesson row → folds to a header badge** (`[lesson]`); content remains
   reachable via `?` (pull) — the exact seam `overlays/help.md`'s lessons
   section already specifies. **Shipped** (spec 055, TASK-117):
   `computeRows`'s `Lesson` field drops to 0 at exactly this threshold.
4. **guardian strip → folds its content into the minibuffer's dormant
   line** (the budget text prefixes the dormant hint,
   `panels/guardian-strip.md`'s fold-relocation rule) — folds LAST because
   decision 7 says the budget is always visible; the fold keeps it visible,
   one row cheaper, never hides it outright. **Shipped** (spec 050):
   `computeRows`'s `Strip` field drops to 0 at exactly this threshold, and
   `minibufferView`'s dormant branch (`internal/tui/views.go`) composes the
   relocated prefix via `guardianBudgetPrefix`.

**Floor layout**: header + body(≥10) + minibuffer(3) + footer — the
pre-reorientation stack. Terminals too short even for that (< 15 rows) keep
the existing behavior (panels shed lowest-priority rows; no new rule).

**Rationale**: folding in inverse order of doctrine strength — the legend is
already sheddable; the villager strip is Wave 5 glanceability (weakest
claim); the lesson row has an explicit designed fallback (badge+overlay is
its own stage-3+ default, so the fold reuses a designed state rather than
inventing one); the guardian strip is the only element a decision says is
*always* visible, so it never disappears — it relocates. `bodyMin = 10`
keeps the dock's smallest useful tab (villagers roster header + a few rows)
and the map viewport genuinely usable; below that the composite was already
degenerate pre-reorientation.

**Alternatives considered**: fixed height breakpoints (e.g. fold at < 28
rows) — rejected: the budget-driven rule (`bodyMin`) adapts to any chrome
combination and matches the existing "rows get scarce → shed lowest-priority"
doctrine; hiding the guardian strip entirely at the floor — rejected:
contradicts decision 7.

Solo views: the solo panel takes the whole body; the villager strip, lesson
row, guardian strip, minibuffer, and footer all persist unchanged around it
— solo zoom only ever replaces the map∥dock body.

## Narrow-fallback chrome rules — Wave 0 ruling (b)

In the narrow (< 112 cols) single-pane layout:

- **Guardian strip**: **carried** — 1 row above the minibuffer, identical
  content (decision 7's "always visible" is width-independent). **Shipped**
  (spec 050) reconciliation: narrow's only minibuffer instance lives inside
  `guardianView` (the guardian pane) — narrow's other panes have no
  composer at all, pre-existing this feature — so the carry lands there,
  unconditionally; narrow has no `computeRows`/fold arithmetic of its own to
  drop the strip against.
- **Lesson row**: **carried** with the same stage defaults as widescreen
  (on at stages 1–2, badge+overlay at 3+/pre-ladder); the same fold rule
  applies against `bodyMin` once narrow's own row budget is worked out
  (narrow has no map∥dock split to shed from first, so its own legend/
  content shedding absorbs pressure before these new rows fold).
- **Villager strip**: **NOT carried** — narrow shows the header count badge
  form only; the villagers solo/pane view is the drill-down.
- **Guardian console / systems tab / exercise panel**: reachable as solo
  views (the existing narrow pattern, `pages/solo-views.md`); no new
  narrow-specific rendering.
- **Ceremony / postmortem**: take over the full screen in narrow exactly as
  in widescreen (takeovers are layout-independent); linear-stream
  projections (D1) are unaffected by layout in either width.

**Rationale**: narrow's contract is "today's single-pane UI, never
deleted" — additive chrome must justify each row it takes from an already-
short layout. The strip and lesson row carry doctrine (decisions 5/7); the
villager strip carries none in narrow (its value is glanceability *beside*
the map, which narrow doesn't render).

**Alternatives considered**: overlay-only lesson delivery in narrow —
rejected: narrow terminals plausibly host new players (SSH sessions), and
stages 1–2 is exactly where pushed delivery matters.

## Style tokens

One named Lipgloss style per role, defined once beside the existing styles in
`views.go` — panels refer to roles, never to raw colors:

| Token | Role | Today's anchor |
|---|---|---|
| `panelBorder` | dormant panel/dock borders | existing `styleBox` rounded border |
| `panelFocus` | focused minibuffer border + title | yellow, same hue as `PAUSED` |
| `tabActive` / `tabInactive` / `tabBadge` | dock tab row | badge dot = `{{skin.guardian.tab_label}} •` |
| `feedDim` | seq, time, default payloads | existing faint/dim style |
| `feedType` | event type column | — |
| `feedName` | `{"A"→"B"}` speaker pairs | — |
| `feedSpeech` | quoted utterances (brightest text on screen) | — |
| `feedClock` | `clock.*` events | yellow |
| `feedSelect` | inspect-mode selected row | background highlight |

Map glyph colors are unchanged (existing terrain/agent styles, night
dimming).

## Composition notes

- Panels are composed with `lipgloss.JoinHorizontal/JoinVertical` over
  independently rendered strings — same technique `View()` uses today, two
  columns instead of one.
- Every panel is handed its exact `(width, height)` and must render to it; no
  panel measures the terminal itself. This is the contract that makes
  dock-tab vs. solo "same component, two widths" work, and the same contract
  the new villager-strip/lesson-row/guardian-strip rows must honor once
  built.
- Implementation note (TASK-34, B1): rendering to *exactly* the handed
  height is a hard requirement, not an aspiration — Bubble Tea scrolls a
  taller-than-terminal `View()` up, which pushes the header off the top of
  the screen. Two lipgloss facts make this easy to violate by accident:
  `Style.Height()` only *pads* short content, it never truncates tall
  content, so one overlong content line silently grows a panel instead of
  erroring; and a style's own `Padding(0,1)` eats 2 columns out of whatever
  `Width()` was set to, before any text renders, so the truly safe content
  width is `Width − 2`, not `Width`. Every panel body in `views.go` clips
  each content line to that true width before handing it to a
  bordered/padded box (`clipContent`) rather than relying on lipgloss's own
  wrapping to stay in bounds. The new chrome rows (villager strip, lesson
  row, guardian strip) will need the identical discipline once built —
  this exact-height invariant is corpus-wide, not per-panel.
