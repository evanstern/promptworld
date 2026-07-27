---
title: Panel — guardian strip
class: panel
status: shipped
verified_against: fe87ae046b3cf13160443d8bb358cb44f651eaf6
---

# Panel: guardian strip

The always-visible action-budget line (reorientation decision 7): one row
pairing the guardian's charge economy with the minibuffer immediately below
it, so the minibuffer reads as *the* verb — the budget for that verb is
never more than one glance away. **Shipped** (spec 050, Wave 2):
`guardianStripView` in `internal/tui/views.go`.

## Mockup

```
 ⚡⚡· (2/3) · next +1 @ 12:00 · 👁 2 standing orders · faith 62
┌─────────────────────────────────────────────────────────────────────────┐
│ ⏎ m — speak with the angel…                                            │
└─────────────────────────────────────────────────────────────────────────┘
```

**No border** — one plain line, exactly 1 row (`patterns/layout.md`'s
row-budget line item), directly above the minibuffer box. Reconciliation
note: the shipped renderer drops the mockup's original illustrative
`charges` word (spec 050 FR-002 — the numeric `(N/cap)` already reads as a
charge count without a label word); the faith segment shipped with spec 085
(TASK-118) in the populated `faith N` form, with `faith —` as the
older-daemon skew state (see "Structure" below).

## Structure

Four segments, left to right, each degrading to absence — or, for faith,
to the honest dashed form — rather than a misleading zero when its
underlying data isn't available:

1. **Charge bank** — the same `⚡`-filled/`·`-empty glyph run
   `panels/guardian.md`'s pane header already renders
   (`Status.Clock.GuardianCharges` / `sim.GuardianChargeCap`), plus the
   numeric form `(N/cap)` for at-a-glance reading without counting glyphs.
   Present whenever a status snapshot exists.
2. **Regen** — `next +1 @ <time>`, the next absolute boundary of the
   EFFECTIVE regen cadence. Since spec 085 the cadence is a pure function of
   village faith (`sim.FaithRegenCadenceTicks` — the band table: fervent 4h /
   steady 6h / wavering 12h / forsaken 24h-floor-or-stopped), so the strip
   reads it off the wire (`Status.Clock.FaithRegenTicks`) rather than a
   compiled constant; against an OLDER daemon that serves no faith field it
   falls back to the legacy exported steady-band constant
   (`sim.GuardianChargeRegenTicks`) — exactly what that daemon fires on.
   **Omitted at a full bank** (spec 050 ruling, research R4.1) **and when no
   regen is scheduled at all** (spec 085: wire cadence 0 — the scenario
   forsaken band; the R4.1 honesty rule generalized): the executor only fires
   `charge_regenerated` below cap and on a live cadence, so forecasting an
   arrival that isn't scheduled would be a lie.
3. **Standing-order count** — `👁 N standing orders`, the length of the
   client-side replica's `MetatronOrders` (the same underlying data
   `panels/guardian.md`'s full standing-orders block counts via its own IPC
   peek — both are projections of the same event stream and can never
   disagree) — the glanceable summary, not a competing surface (open the
   guardian tab or console for the full per-order rows). **Zero is a true
   value** here (spec 050 ruling): `👁 0 standing orders` renders rather than
   being omitted — the mechanic exists and its count is genuinely 0, unlike
   the absence cases below.
4. **Faith** — `faith N` (spec 085, TASK-118): the village's event-sourced
   devotion score (0..100), read from `Status.Clock.Faith` — a pointer field
   a spec-085 daemon always serves (the sim accessor is nil-safe, genesis
   50). Renders `faith —` (present, dashed) exactly when the pointer is nil
   — a TUI carrying this renderer against a daemon that predates the field —
   claiming nothing, inventing no zero (the strip's standing honesty rule).
   Positioned LAST so it is the FIRST segment dropped under width pressure
   (the drop-order contract below). In-fiction number only — no badge, no
   streak, no delta arrow (the overjustification caution: faith is a world
   fact, not a score surface).

**Pre-status (connecting)**: before the first status snapshot arrives, the
row is present — occupying its row budget — but renders blank (spec 050
ruling, research R4.2): no invented zeros for a bank or order count the
client doesn't know yet.

**Width pressure**: segments truncate from the right with `…` (spec 050
ruling, research R4.3) — drop order is faith first (spec 085: positioned
last), then standing-orders, then regen, then the charge bank (the headline,
and the only segment ever hard-clipped mid-segment, at widths too narrow
even for it alone) — never wraps to a second row.

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

`promptworld status`/`metatron_status` exposes the charge bank, the regen
timing, the standing orders count, and (since spec 085) the faith score and
the effective regen cadence (`ClockStatus.Faith` / `FaithRegenTicks`, both
served from the sim's own `FaithScore`/`FaithRegenCadenceTicks` — the daemon
never re-derives bands) model-free — a non-TUI observer loses nothing this
strip surfaces; the strip is a glanceable TUI-side restatement of facts the
CLI/IPC protocol carries. Spec 050 added no wire field; spec 085 added
exactly the two above.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| charge-bank segment | 0..cap charges | `Status.Clock.GuardianCharges`, `sim.GuardianChargeCap` | `guardianStripView` | — (display-only) | reorient decision 7 / spec 050 | — |
| regen segment | next boundary time · omitted at full bank or when no regen is scheduled (wire cadence 0) | `Status.Clock.Tick`, `Status.Clock.FaithRegenTicks` (spec 085; `sim.GuardianChargeRegenTicks` fallback against a pre-085 daemon) | `guardianStripView` | — | reorient decision 7 / spec 050 / spec 085 | — |
| standing-order count segment | 0..N (0 is a true value) | client replica `MetatronOrders` (length) | `guardianStripView` | — | reorient decision 7 / spec 050 | — |
| faith segment | populated `faith N` · present-dashed `faith —` (nil pointer — older daemon) | `Status.Clock.Faith` (spec 085) | `guardianStripView` | — | reorient decision 7 / spec 085 | — |
| fold-relocation into minibuffer | strip row · relocated-into-dormant-line | fold pressure (`patterns/layout.md`, `computeRows`) | `guardianBudgetPrefix` (composed in `minibufferView`'s dormant branch) | — | reorient decision 7 / `patterns/layout.md` ruling a / spec 050 | — |

**Parity rollout**: this page is display-only end to end — no actionable
control exists on it (the strip has no keys of its own), so no parity gap
to track.
