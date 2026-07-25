---
title: Panel — guardian strip
class: panel
status: shipped
verified_against: ab212309e8fd13aa069e7c6d7c7c7c5821213835
---

# Panel: guardian strip

The always-visible action-budget line (reorientation decision 7): one row
pairing the guardian's charge economy with the minibuffer immediately below
it, so the minibuffer reads as *the* verb — the budget for that verb is
never more than one glance away. **Shipped** (spec 050, Wave 2):
`guardianStripView` in `internal/tui/views.go`.

## Mockup

```
 ⚡⚡· (2/3) · next +1 @ 12:00 · 👁 2 standing orders
┌─────────────────────────────────────────────────────────────────────────┐
│ ⏎ m — speak with the angel…                                            │
└─────────────────────────────────────────────────────────────────────────┘
```

**No border** — one plain line, exactly 1 row (`patterns/layout.md`'s
row-budget line item), directly above the minibuffer box. Reconciliation
note: the shipped renderer drops the mockup's illustrative `charges` word
and `faith —` segment (spec 050 FR-002/assumptions carry higher authority
here than this page's original mockup did before it shipped) — the numeric
`(N/cap)` already reads as a charge count without a label word, and the
faith segment doesn't exist at all yet (see "Structure" below).

## Structure

Three segments today, left to right, each degrading to absence rather than
a misleading zero when its underlying data isn't available (a fourth,
faith, is reserved but not yet implemented — see below):

1. **Charge bank** — the same `⚡`-filled/`·`-empty glyph run
   `panels/guardian.md`'s pane header already renders
   (`Status.Clock.GuardianCharges` / `sim.GuardianChargeCap`), plus the
   numeric form `(N/cap)` for at-a-glance reading without counting glyphs.
   Present whenever a status snapshot exists.
2. **Regen** — `next +1 @ <time>`, the next absolute 6-game-hour boundary at
   which `metatron.charge_regenerated` fires (`[[metatron]]`'s charge-economy
   rule, cadence exported read-only as `sim.MetatronChargeRegenTicks`) — a
   plain restatement of a mechanic that already exists, not a new one.
   **Omitted at a full bank** (spec 050 ruling, research R4.1): the executor
   only fires `charge_regenerated` below cap, so forecasting an arrival that
   isn't scheduled would be a lie.
3. **Standing-order count** — `👁 N standing orders`, the length of the
   client-side replica's `MetatronOrders` (the same underlying data
   `panels/guardian.md`'s full standing-orders block counts via its own IPC
   peek — both are projections of the same event stream and can never
   disagree) — the glanceable summary, not a competing surface (open the
   guardian tab or console for the full per-order rows). **Zero is a true
   value** here (spec 050 ruling): `👁 0 standing orders` renders rather than
   being omitted — the mechanic exists and its count is genuinely 0, unlike
   the absence cases below.
4. **Faith** — reserved for TASK-118 (not yet landed, per the reorientation
   move list). Will render as `faith —` (present, dashed) once TASK-118's
   field exists on the wire but before this strip has anything meaningful to
   show, and is **omitted entirely** (not even the dash, not even the
   segment) while TASK-118 hasn't shipped at all — the strip must never
   claim a mechanic that doesn't exist yet. Nothing in `guardianStripView`
   references a faith field today; there is no code path to activate.

**Pre-status (connecting)**: before the first status snapshot arrives, the
row is present — occupying its row budget — but renders blank (spec 050
ruling, research R4.2): no invented zeros for a bank or order count the
client doesn't know yet.

**Width pressure**: segments truncate from the right with `…` (spec 050
ruling, research R4.3) — drop order is standing-orders, then regen, then the
charge bank (the headline, and the only segment ever hard-clipped
mid-segment, at widths too narrow even for it alone) — never wraps to a
second row.

## Behavior

- **Always visible, all stages** (`patterns/stage-defaults.md`) — decision 7
  states the budget is always visible; no stage default ever turns this row
  off. (`patterns/stage-defaults.md` itself is a later wave; this page's
  renderer has no stage conditional at all today, which trivially satisfies
  "always visible.")
- **Fold-last relocation** (`patterns/layout.md` ruling a, step 4): under
  height pressure this is the LAST row to fold (`internal/tui/layout.go`'s
  `computeRows`, `Strip` field, `bodyMin` threshold), and folding never
  hides its content — it **relocates into the minibuffer's dormant-state
  line** (`guardianBudgetPrefix`), prefixing the existing placeholder hint:
  ```
  ⚡⚡· 👁2 · ⏎ m — speak with the angel…
  ```
  This compact relocated form carries the bank glyphs alone (no `(N/cap)`
  numeric — that detail stays one tab away) and the order count with no
  words (`👁2`, not `👁 2 standing orders`) — deliberately terser than the
  full strip, since it now shares a line with the dormant hint. This
  relocation applies to the **dormant** state only; the focused and busy
  minibuffer states keep their existing content unchanged (verified
  byte-identical by regression test, `internal/tui/render_test.go`). Spec
  050 ruling (extending the strip's own honesty rule to its folded form):
  the relocated prefix is itself blank (omitted, plain placeholder only)
  before the first status snapshot — no invented zeros in the folded form
  either.
- **Carried in narrow** (`patterns/layout.md` ruling b) — decision 7's
  "always visible" is width-independent; the narrow fallback keeps this row
  exactly as widescreen does. Reconciliation note: the narrow fallback's
  only minibuffer instance lives inside `guardianView` (the guardian pane) —
  narrow's other panes (map/chronicle/villagers) have no composer at all
  pre-existing this feature — so "carried, above the minibuffer" lands
  there, unconditionally (narrow has no `computeRows`/fold arithmetic of its
  own to drop it against).

## Linear-stream / CLI projection (D1)

`promptworld status`/`metatron_status` already exposes the charge bank,
regen timing (via the clock/charge-regen cadence fields), and the standing
orders count model-free — a non-TUI observer loses nothing this strip
surfaces; the strip is a glanceable TUI-side restatement of facts the
CLI/IPC protocol already carries. This feature added no new wire field.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| charge-bank segment | 0..cap charges | `Status.Clock.GuardianCharges`, `sim.GuardianChargeCap` | `guardianStripView` | — (display-only) | reorient decision 7 / spec 050 | — |
| regen segment | next boundary time · omitted at full bank | `Status.Clock.Tick`, `sim.MetatronChargeRegenTicks` | `guardianStripView` | — | reorient decision 7 / spec 050 | — |
| standing-order count segment | 0..N (0 is a true value) | client replica `MetatronOrders` (length) | `guardianStripView` | — | reorient decision 7 / spec 050 | — |
| faith segment | absent (pre-TASK-118) · present-dashed · populated | TASK-118's future status field | `unbuilt (pending TASK-118)` | — | reorient decision 7 | — |
| fold-relocation into minibuffer | strip row · relocated-into-dormant-line | fold pressure (`patterns/layout.md`, `computeRows`) | `guardianBudgetPrefix` (composed in `minibufferView`'s dormant branch) | — | reorient decision 7 / `patterns/layout.md` ruling a / spec 050 | — |

**Parity rollout**: this page is display-only end to end — no actionable
control exists on it (the strip has no keys of its own), so no parity gap
to track.
