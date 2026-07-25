---
title: Panel — guardian strip
class: panel
status: specified
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
---

# Panel: guardian strip

The always-visible action-budget line (reorientation decision 7): one row
pairing the guardian's charge economy with the minibuffer immediately below
it, so the minibuffer reads as *the* verb — the budget for that verb is
never more than one glance away. **Not built** — specified spec-before-
build for a later reorientation wave (Wave 2).

## Mockup

```
 ⚡⚡· charges (2/3) · next +1 @ 12:00 · 👁 2 standing orders · faith —
┌─────────────────────────────────────────────────────────────────────────┐
│ ⏎ m — speak with the {{skin.guardian.epithet}}…                        │
└─────────────────────────────────────────────────────────────────────────┘
```

**No border** — one plain line, exactly 1 row (`patterns/layout.md`'s
row-budget line item), directly above the minibuffer box.

## Structure

Four segments, left to right, each degrading to absence rather than to a
misleading zero when its underlying data isn't available:

1. **Charge bank** — the same `⚡`-filled/`·`-empty glyph run
   `panels/guardian.md`'s pane header already renders
   (`Status.Clock.MetatronCharges` / `sim.MetatronChargeCap`), plus the
   numeric form `(N/cap)` for at-a-glance reading without counting glyphs.
2. **Regen** — `next +1 @ <time>`, the next absolute 6-game-hour boundary at
   which `metatron.charge_regenerated` fires ([[metatron]]'s charge-economy
   rule) — a plain restatement of a mechanic that already exists, not a new
   one.
3. **Standing-order count** — `👁 N standing orders`, the length of
   `Status.Orders` — the same data `panels/guardian.md`'s full
   standing-orders block already lists in detail; this segment is the
   glanceable summary, not a competing surface (open the guardian tab or
   console for the full per-order rows).
4. **Faith** — reserved for TASK-118 (not yet landed, per the reorientation
   move list). Renders as `faith —` (present, dashed) once TASK-118's field
   exists on the wire but before this strip has anything meaningful to show,
   and is **omitted entirely** (not even the dash) while TASK-118 hasn't
   shipped at all — the strip must never claim a mechanic that doesn't
   exist yet.

## Behavior

- **Always visible, all stages** (`patterns/stage-defaults.md`) — decision 7
  states the budget is always visible; no stage default ever turns this row
  off.
- **Fold-last relocation** (`patterns/layout.md` ruling a, step 4): under
  height pressure this is the LAST row to fold, and folding never hides its
  content — it **relocates into the minibuffer's dormant-state line**,
  prefixing the existing placeholder hint:
  ```
  ⚡⚡· 👁2 · ⏎ m — speak with the {{skin.guardian.epithet}}…
  ```
  This relocation applies to the **dormant** state only; the focused and
  busy minibuffer states keep their existing content unchanged (the input
  line and the busy notice already have no room to spare, and the budget
  context remains one tab away in `panels/guardian.md`/the guardian
  console).
- **Carried in narrow** (`patterns/layout.md` ruling b) — decision 7's
  "always visible" is width-independent; the narrow fallback keeps this row
  exactly as widescreen does.

## Linear-stream / CLI projection (D1)

`promptworld status`/`metatron_status` already exposes the charge bank,
regen timing (via the clock/status fields), and `Status.Orders` count
model-free — a non-TUI observer loses nothing this strip surfaces; the strip
is a glanceable TUI-side restatement of facts the CLI/IPC protocol already
carries.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| charge-bank segment | 0..cap charges | `Status.Clock.MetatronCharges`, `sim.MetatronChargeCap` | `unbuilt (wave 2)` | — (display-only) | reorient decision 7 | — |
| regen segment | next boundary time | clock/charge-regen timing | `unbuilt (wave 2)` | — | reorient decision 7 | — |
| standing-order count segment | 0..N | `Status.Orders` (length) | `unbuilt (wave 2)` | — | reorient decision 7 | — |
| faith segment | absent (pre-TASK-118) · present-dashed · populated | TASK-118's future status field | `unbuilt (wave 2, pending TASK-118)` | — | reorient decision 7 | — |
| fold-relocation into minibuffer | strip row · relocated-into-dormant-line | fold pressure (`patterns/layout.md`) | `unbuilt (wave 2)` | — | reorient decision 7 / `patterns/layout.md` ruling a | — |

**Parity rollout**: this page is display-only end to end — no actionable
control exists on it (the strip has no keys of its own), so no parity gap
to track.
