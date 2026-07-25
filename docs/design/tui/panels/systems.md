---
title: Panel — systems (engine telemetry)
class: panel
status: specified
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
sources:
  - internal/tui/views.go
  - internal/tui/tui.go
---

# Panel: systems

The dock's **never-skinned telemetry** content (D10 — the skin boundary is a
file boundary; this page carries zero skin tokens, by design, unlike
[guardian.md](guardian.md)).

**Hybrid status, stated plainly**: `status: specified` because the *systems
tab itself* — a fourth dock tab an operator selects with its own key — does
not exist yet in `internal/tui` (it lands in a later reorientation wave,
per D10's "systems tab" plan). But every control this page describes is
**already shipped**, rendering today inside the single "metatron" tab
alongside the fiction-layer content [guardian.md](guardian.md) now owns. This
page is written spec-before-build for the *tab*, reconciliation-accurate for
the *content*: the `renderer` column below names real, currently-executing Go
symbols — not `unbuilt (wave N)` — because there is nothing left to build for
the content itself, only the tab it will eventually move into.

## Mockup

```
┌─ chronicle │ metatron │ villagers │ SYSTEMS (Wave 2–3) ┐
├─────────────────────────────────────────────────────────┤
│ local      cogito              ● q0  1/2   $0.42        │
│   uncalibrated — suppressed at 32x — calibrate or slow  │
│ cloud      gemma4:12b-mlx       ● q2  2/4 ⏳ $1.18        │
│ (unattributed)                             $0.05         │
│ spend $1.65 of $100                                      │
│                                                           │
│ 🜂 cognition horizon                                     │
│   planner thinking at 32x                                │
│   conversation suppressed at 32x — slow down · skipped 4 │
└─────────────────────────────────────────────────────────┘
```

## Structure

1. **Provider table** (spec 024 US6) — one row per LLM provider: name, model,
   an up (`●`)/down (`○`) glyph, queue depth, inflight/slots, a `⏳` contended
   marker, and this provider's dollar spend share of the month, sorted by
   name. A trailing `(unattributed)` row covers legacy-metered spend the
   per-provider rows don't attribute, followed by a `spend $X of $Y` wallet
   line (and a budget-exhausted notice once spend meets the cap).
2. **Health-condition rows** (spec 034) — a provider currently carrying an
   active health condition gains an indented continuation line, immediately
   below its row, rendering the condition's plain-language detail and remedy
   in the pane's error style.
3. **Horizon block** (spec 037 US1/US2) — a `🜂 cognition horizon` header,
   then one row per watched cognition class: a thinking class reads "`<class>`
   thinking at `<speed>`"; a suppressed class is warn-styled with its remedy
   (`horizonRemedy` — "calibrate or slow down" for an uncalibrated class,
   "slow down" for a calibrated one **— this is spec 035's calibration-UX
   voice inside the TUI**, the same warn-loudly doctrine the CLI's
   speed-change reply and boot warning carry) plus the router's own verdict
   arithmetic as a dim trailing detail. A trailing "· skipped N" appears on
   every suppressed row and on a thinking row only once it has ever been
   suppressed. Absent entirely on a no-LLM world.

## Related telemetry rendered elsewhere (not owned by this page)

Specs 028 (adaptive throttle), 031 (estimator breach-adoption), and 033
(governor accrued debt) are the same engine-telemetry family this tab
conceptually houses (D10), but their shipped TUI surfaces today render
**outside** the dock, so their control-table rows are owned by those pages,
not duplicated here:

- The header's governed-speed suffix (`"asked 32x — 3 minds in flight, debt
  140%"`, `governedSpeedSuffix`) — owned by
  [../pages/home.md](../pages/home.md)'s header row.
- The raw chronicle feed's `clock.governor_shed`/`clock.governor_recovered`
  digest lines — owned by
  [../patterns/chronicle-grammar.md](../patterns/chronicle-grammar.md).

`anatomy.md` cross-references both from this page's entry so an implementer
looking for "the governor" finds all three surfaces in ≤2 hops regardless of
which one they start from.

## Narrow behavior

No narrow-specific rendering: once built, the systems tab is reachable as a
solo/narrow pane exactly like every other dock tab
(`patterns/layout.md` ruling b — "systems tab: reachable as solo views, no
new narrow-specific rendering"; [pages/solo-views.md](../pages/solo-views.md)).
Content does not reflow differently below the 112-column breakpoint beyond
the existing width-aware column-dropping every dock tab already does.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| systems tab (dock selection) | — | — | `unbuilt (wave 2-3)` — no 4th dock-tab key exists yet | `unbuilt` · — | reorient D10 | — |
| provider table row | up · down · contended | `Status.LLM.Providers[i]` | `llmProviderLines` | — | spec 024 | — |
| provider health continuation line | absent · shown | `ProviderStatus.Condition`/`ConditionDetail`/`ConditionRemedy` | `llmProviderLines` | — | spec 034 | — |
| `(unattributed)` spend row | absent · shown | `Status.LLM.Spent − Σ(provider.SpentUSD)` | `llmProviderLines` | — | spec 024 | — |
| spend/budget wallet line | under budget · exhausted | `Status.LLM.Spent`/`.Budget` | `metatronView` (spend line) | — | spec 024 | — |
| horizon header | absent (no LLM) · shown | `Status.Horizon` | `horizonLines` | — | spec 037 | — |
| horizon class row | thinking · suppressed | `ipc.HorizonClass` | `horizonRow` | — | spec 037 | — |
| horizon remedy detail | calibrated · uncalibrated | `HorizonClass.Calibrated` | `horizonRemedy` | — | spec 035 | — |
| horizon skipped-count | absent · N>0 | `HorizonClass.SuppressedCount` | `horizonRow` | — | spec 037 US2 | — |

**Parity rollout**: this page is display-only end to end — no actionable
control exists on it today (no tab key yet), so no parity gap to track.
