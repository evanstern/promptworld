---
title: Panel — systems (engine telemetry)
class: panel
status: shipped
verified_against: a30ee798ff6cc6316256d7833aead1e8a4c9a849
sources:
  - internal/tui/views.go
  - internal/tui/tui.go
---

# Panel: systems

The dock's **never-skinned telemetry** content (D10 — the skin boundary is a
file boundary; this page carries zero skin tokens, by design, unlike
[guardian.md](guardian.md)).

**Built** (spec 053, TASK-125): the fourth dock tab (key `5`) now exists in
`internal/tui` — `paneSystems`/`dockTabKey[paneSystems]` (tui.go),
`systemsContentBody`/`systemsView` (views.go, the dock/solo and narrow-
fallback call sites). Every renderer this page describes was already
shipped before this feature (spec 024/034/035/037) and moves here
unchanged — a relocation, not a rewrite: `llmProviderLines`, `horizonLines`/
`horizonRow`/`horizonRemedy` are the exact same functions the guardian tab's
`guardianView` used to call directly.

## Mockup

```
┌─ chronicle │ {{skin.guardian.tab_label}} │ villagers │ SYSTEMS (Wave 2–3) ┐
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
4. **No-LLM absence, stated honestly** (spec 053 SC-002) — a world with no
   `Status.LLM` at all (no `llm.json`) renders neither an empty provider
   table nor blank chrome: `systemsContentBody` states "no LLM configured
   for this world" outright, the tab's one honesty rule now that it's a
   whole dedicated destination rather than a silently-empty corner of the
   fiction-layer tab.

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

No narrow-specific rendering: the systems tab is reachable as a solo/narrow
pane exactly like every other dock tab (`patterns/layout.md` ruling b —
"systems tab: reachable as solo views, no new narrow-specific rendering";
[pages/solo-views.md](../pages/solo-views.md)) — `systemsView` (narrow
fallback) shares the exact same `systemsContentBody` the dock/solo call
sites use. Content does not reflow differently below the 112-column
breakpoint beyond the existing width-aware column-dropping every dock tab
already does.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| systems tab (dock selection) | selected · unseen (never — no badge) | `Model.dockTab`, `paneSystems` | `dockTabsRow`, `dockTabContent`/`systemsView` | `5` select · — | spec 053 | — |
| provider table row | up · down · contended | `Status.LLM.Providers[i]` | `llmProviderLines` | — | spec 024 | — |
| provider health continuation line | absent · shown | `ProviderStatus.Condition`/`ConditionDetail`/`ConditionRemedy` | `llmProviderLines` | — | spec 034 | — |
| `(unattributed)` spend row | absent · shown | `Status.LLM.Spent − Σ(provider.SpentUSD)` | `llmProviderLines` | — | spec 024 | — |
| spend/budget wallet line | under budget · exhausted | `Status.LLM.Spent`/`.Budget` | `systemsContentBody` (spend line) | — | spec 024 | — |
| no-LLM absence notice | absent (LLM configured) · shown | `Status.LLM == nil` | `systemsContentBody` | — | spec 053 | — |
| horizon header | absent (no LLM) · shown | `Status.Horizon` | `horizonLines` | — | spec 037 | — |
| horizon class row | thinking · suppressed | `ipc.HorizonClass` | `horizonRow` | — | spec 037 | — |
| horizon remedy detail | calibrated · uncalibrated | `HorizonClass.Calibrated` | `horizonRemedy` | — | spec 035 | — |
| horizon skipped-count | absent · N>0 | `HorizonClass.SuppressedCount` | `horizonRow` | — | spec 037 US2 | — |

**Parity rollout**: every control above has a key (`5`) but no mouse
target — tracked here rather than omitted, same as every other pre-decision-8
dock control (`panels/dock.md`).
