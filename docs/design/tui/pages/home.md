---
title: Page — home (widescreen composite)
class: page
status: shipped
verified_against: 72f82f41f7aa2e345572105894cd0fb7c02fc0aa
sources:
  - internal/tui/views.go
  - internal/tui/layout.go
---

# Page: home (widescreen composite)

The default view whenever the terminal is wide enough (see
[../patterns/layout.md](../patterns/layout.md) for the breakpoint). Replaces
the "one pane at a time" model as the resting state of the app.

## Mockup (today's shipped chrome)

```
 promptworld — tick 8801 · day 4 08:12 · running · speed 1x (1.0 t/s)
 ┌─ MAP · following centroid ────────────────┐ ┌─ chronicle │ {{skin.guardian.tab_label}} │ villagers ┐
 │ ~ ~ ~ ~ " " ♠ ♠ ♠ ♠ ♠ " " . . . . ▲ . .   │ │  08:09 agent.foraged           │
 │ ~ ~ ~ " " ♠ ♠ A ♠ ♠ " " . . ⌂ ⌂ . . . .   │ │    Ash foraged at (14,9)       │
 │ ~ ~ " ♠ ♠ ♠ R ♠ " " . . . ⌂ . B . . .     │ │  08:11 social.conv_turn        │
 │ ~ ~ . " " ♠ ♠ " . . . . . . . . . . .     │ │    Ash→Rowan "the fire's low   │
 │ ~ . . ᴥ . " " . . . . S . . . " " . .     │ │    again"                     │
 │ ~ . . . . . . . . . . . . " " ♠ ♠ . .     │ │  08:11 social.conv_turn        │
 │ ~ ~ . . . . " " . . . . . ♠ ♠ ♠ ♠ . .     │ │    Rowan→Ash "I stacked wood   │
 │                                           │ │    at dawn"                   │
 │ ~ water ♠ wood " forage ᴥ den ▲ fire ⌂ sh │ │                               │
 └───────────────────────────────────────────┘ └───────────────────────────────┘
 ┌─────────────────────────────────────────────────────────────────────────────┐
 │ ⏎ m — speak with the {{skin.guardian.epithet}}…                            │
 └───────────────────────────────────────────────────────────────────────────┘
  2 chronicle 3 {{skin.guardian.tab_label}} 4 villagers (again: solo) · m ask · space pause · q quit
```

**Reconciliation correction**: the pre-v2 mockup showed a right-aligned
`N villagers awake` header segment and raw inline-JSON feed lines — neither
exists in the shipped renderer. `headerView` composes no villager-count
segment at all (there is no such element to map in `anatomy.md`), and the
raw feed has rendered through the digest grammar (spec 018/TASK-60,
[../patterns/chronicle-grammar.md](../patterns/chronicle-grammar.md)) since
before this feature — inline JSON payloads only ever appear in the paused
inspect-mode detail pane
([../panels/chronicle.md](../panels/chronicle.md) "Mode 2"), never inline in
the running feed. This mockup is corrected to match; see "Header segments"
below for the full, reconciled header inventory (specs 028/034/037/039/044).

## Mockup (full teaching chrome — shipped)

Three permanent chrome rows exist around this same composite — a villager
strip under the header, a lesson row above the guardian strip, and the
guardian strip itself above the minibuffer (decisions 5/7/12). **All three
are now shipped** (guardian strip: spec 050; lesson row: spec 055/TASK-117;
villager strip: spec 060/TASK-129 — each row's own page is
`status: shipped`); the row math and fold order are ruled in
[../patterns/layout.md](../patterns/layout.md) and each row is fully
specified on its own page (mockup, control table, stage defaults, linear
projection). No header villager-count segment renders while the strip
itself is showing — the two are mutually exclusive (FR-002; see
"Header segments" below for the folded badge form):

```
 promptworld · attached · day 4 · 08:12 · 1×
 ┌─ VILLAGER STRIP — one row, stage-defaulted ───────────────────────── (D12) ┐
 ┌─ MAP ──────────────────────────────────────┐ ┌─ chronicle │ … │ villagers ┐
 │  … unchanged …                             │ │  … unchanged …            │
 └─────────────────────────────────────────────┘ └───────────────────────────┘
 ┌─ LESSON ROW — one active lesson, ≤2 lines ───────────────────── (decision 5) ┐
 ┌─ GUARDIAN STRIP — budget: charge bank · regen · orders · faith ─ (decision 7) ┐
 ┌─────────────────────────────────────────────────────────────────────────────┐
 │ ⏎ m — speak with the {{skin.guardian.epithet}}…                            │
 └───────────────────────────────────────────────────────────────────────────┘
  2 chronicle 3 guardian 4 villagers (again: solo) · m ask · space pause · q quit
```

If height pressure instead folds the villager strip to its header badge
form (`patterns/layout.md` ruling a step 2), the header row reads:

```
 promptworld · attached · day 4 · 08:12 · 1×                  [4 villagers]
```

- [panels/villager-strip.md](../panels/villager-strip.md) (D12) —
  colonist-bar glyph run, folds to a header count badge.
- [panels/lesson-row.md](../panels/lesson-row.md) (decision 5) — one active
  lesson, dwell/dismiss, anti-spam, per-user seen state.
- [panels/guardian-strip.md](../panels/guardian-strip.md) (decision 7) —
  charge bank/regen/order-count/faith, fold-last relocation into the
  minibuffer's dormant line.
- Stage-shaped defaults for all three, and the fold order once the terminal
  runs short: [../patterns/stage-defaults.md](../patterns/stage-defaults.md)
  (the authority table) and [../patterns/layout.md](../patterns/layout.md)
  (ruling a/b — row budget re-derivation happens here regardless of when
  the rows themselves ship, so the arithmetic is never retrofitted).

## Composition

| Region | Spec | Notes |
|---|---|---|
| header | 1 row | see "Header segments" below |
| map | left, all remaining columns | [../panels/map.md](../panels/map.md) |
| dock | right, fixed width | [../panels/dock.md](../panels/dock.md); chronicle tab is the default on launch |
| minibuffer | full width, above footer | [../panels/minibuffer.md](../panels/minibuffer.md) |
| footer | 1 row of key hints | hints change with mode — see [../patterns/keymap.md](../patterns/keymap.md) |

## Header segments (reconciled)

The header (`headerView`) composes, left to right: world name (or a
disconnected/retrying error), `tick N`, game time, clock state, speed, then
zero or more conditional badges. Every segment renders byte-identically to
before its introducing spec when the condition it guards is false/absent —
no existing world's header changes because this feature landed.

| Segment | States | Introduced-by |
|---|---|---|
| clock state | `running` · `PAUSED` (amber) · `ENDED` (bold red, outranks `PAUSED` regardless of the clock state the run ended under) | TASK-34 / spec 044 |
| speed | `speed <N>x (<rate> t/s)` | TASK-34 |
| governed-speed suffix | absent (ungoverned) · `asked <N>x — <j> minds in flight, debt <d>%` | spec 028 (the same surface spec 039's teaching soft-cap warning rides — no separate segment) |
| `[degraded]` | absent · shown | pre-existing (`Clock.Degraded`) |
| `[llm: provider kind]` | absent · shown (first name-sorted affected provider) | spec 034 |
| `[suppressed: class, class]` | absent · shown (wire-ordered watched classes) | spec 037 |
| `[lesson]` | absent · shown (stage-3+/pre-ladder default, or once the lesson row folds) | spec 055 |
| `[N villagers]` | absent (the villager strip itself is showing) · shown (narrow, ruling b — or a widescreen fold, ruling a step 2) | spec 060 |

The **postmortem posture** (spec 044): `ENDED` outranking `PAUSED` also makes
the clock keys (space, `[`, `]`) client-side no-ops and swaps the footer's
pause/resume hint for `run ended (read-only)` — every reading surface stays
fully functional (see [../panels/minibuffer.md](../panels/minibuffer.md) and
`patterns/keymap.md`'s footer table).

## Behavior

- This page is always "underneath": solo views and the narrow fallback are
  the only things that replace it, and `1` / `esc` (when nothing is focused)
  always return here.
- Map and dock update live simultaneously — no region ever freezes because
  another has attention.
- Pausing the clock changes chronicle-tab behavior (inspect mode) but never
  the layout.
- Arrow keys pan the map from this page as long as the minibuffer is not
  focused, regardless of which dock tab is selected.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| world name / disconnect | connected · disconnected (retrying) | `Model.connected`, `Model.lastErr` | `headerView` | — (display-only) | TASK-34 | — |
| clock state segment | running · paused · ended | `Status.Clock.Paused`, `Model.runEnded()` | `headerView` | `space` pause/resume · — | TASK-34/spec 044 | — |
| speed segment | N steps, governed/ungoverned | `Status.Clock.Speed`/`RequestedSpeed`/`GovernorDebt`/`GovernorJobs` | `headerView`, `governedSpeedSuffix` | `[`/`]` · — | TASK-34/spec 028/039 | — |
| `[degraded]` badge | absent · shown | `Status.Clock.Degraded` | `headerView` | — | pre-existing | — |
| `[llm: …]` badge | absent · shown | `Status.LLM` (`firstLLMCondition`) | `headerView` | — | spec 034 | — |
| `[suppressed: …]` badge | absent · shown | `Status.Horizon` (`suppressedHorizonClasses`) | `headerView` | — | spec 037 | — |
| `[lesson]` badge | absent · shown | `Model.lessonBadgeVisible` | `headerView` | — | spec 055 | — |
| `[N villagers]` badge | absent · shown | `Model.villCount`, `computeRows.VillagerStrip` | `Model.villagerCountBadge` | — | spec 060 | — |

**Parity rollout**: pause/resume (`space`) and speed (`[`/`]`) have no mouse
target today; tracked here rather than omitted (decision 8, formal doctrine
in `patterns/keymap.md`, T024). Every badge above is display-only (no keys,
no gap).
