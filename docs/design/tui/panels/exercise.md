---
title: Panel — exercise (scenario tab)
class: panel
status: shipped
verified_against: 72f82f41f7aa2e345572105894cd0fb7c02fc0aa
sources:
  - internal/tui/exercise.go
  - internal/tui/tui.go
  - internal/tui/views.go
  - internal/sim/scenario.go
  - internal/sim/curriculum.go
---

# Panel: exercise

The fifth dock tab (reorientation D11), present only on scenario worlds:
the framing, live rubric progress, and pass/fail state of the world's
seeded exercise (`sim.ExerciseDefinition`, [[curriculum-ladder]]). Shipped
by spec 054 (TASK-119) on **key `6`**, beside the systems tab's `5`
(spec 053, TASK-125); on ambient worlds the tab does not exist and `6`
falls through inert.

## Mockup — attach-time briefing (first render this attach)

```
┌─ chronicle │ {{skin.guardian.tab_label}} │ villagers │ systems │ EXERCISE ┐
├──────────────────────────────────────────────────────────────────┤
│ FIRST NIGHT                                                       │
│                                                                    │
│ a seeded world tuned so night one is survivable only if the        │
│ guardian is directed well: fuel scarce, the gru active             │
│                                                                    │
│ Incidents ahead are forecast — you'll see the schedule before      │
│ they land.                                                         │
│                                                                    │
│                                          any key — begin           │
└──────────────────────────────────────────────────────────────────┘
```

## Mockup — live gauges (after the briefing is dismissed)

```
┌─ chronicle │ {{skin.guardian.tab_label}} │ villagers │ systems │ EXERCISE ┐
├──────────────────────────────────────────────────────────────────┤
│ FIRST NIGHT · in progress                                         │
│                                                                    │
│ … village survives to dawn of day 2   (sim.day_started: 0)         │
│ ✓ no villager dies                    (agent.died: 0)              │
│ … a watch set before nightfall        (guardian.order_placed: 0)   │
│                                                                    │
│ incidents (forecast): the gru emerges ~22:00 (day 1)               │
└──────────────────────────────────────────────────────────────────┘
```

## Structure

1. **Title line** (`exerciseBody`) — `<NAME> · in progress|passed|failed
   (run ended)`: the exercise id rendered as its display name
   (`exerciseTitle`) plus the outcome posture (`exerciseOutcomeLabel`).
2. **Attach-time briefing** (`exerciseBriefingBody`, the "Portal safe-room"
   reference) — shown once per client attach, before the live gauges:
   `sim.ExerciseDefinition.Framing` + the current incident-visibility mode
   for this stage (below). Dismissed by any key (`handleKey` consumes
   exactly one press, only while this tab is visible —
   `exerciseBriefingShowing`; an open guardian console suppresses the
   eater, being a whole-body takeover — `patterns/keymap.md` "Mode:
   exercise briefing"), then never reshown for that attach; re-attaching
   resets it (`Model.exBriefingDismissed`, cleared on `connectedMsg`) — a
   fresh orientation each session, not a one-time-ever notice.
3. **Rubric gauges** — one row per evaluated rubric term
   (`sim.EvaluateRubric` over the replica — the SAME pure derivation the
   executor's pass precondition AND every report-card surface read (spec
   072), so the panel, the emitter, and the cards can never disagree). Each
   gauge shows the term in plain language, a met/pending marker (`✓`/`…`),
   and the backing event type + count (engineer-facing detail, FR-020 — the
   plain-language callout is the term itself; the raw event reference sits
   alongside it, not instead of it). Every cataloged exercise carries a
   production evaluator arm (spec 077 FR-002 — the nine-exercise catalog
   below; sweep-tested, `TestNoCatalogedExerciseReachesDefaultArm`), so no
   shipped exercise renders permanently pending; a future exercise without
   an evaluator arm gets its terms rendered pending, the honest default.
   Zero-wanted terms ("nothing is taken", counting the `stranger.took`
   ledger) render Met at genesis — an empty ledger IS the claim (the
   spec-072 honest-gauge posture).
4. **Incident-visibility vocabulary** (D4) — a **vocabulary, not a
   boolean**, per `patterns/stage-defaults.md` (the authority) and
   `sim.IncidentVisibilityFor` (definition override wins, else the
   stage-keyed default): `forecast` (the authored schedule is shown ahead
   of time with approximate game times — `exerciseIncidentLine`; stages 1–2
   and pre-ladder) or `fog` (the line is omitted entirely — incidents are
   revealed only as they happen; stage 3+). The gauges section never
   changes shape between the two.
5. **Pass/fail banner** — `in progress` (default) → `passed` (the recorded
   `curriculum.exercise_passed` reached the replica — the pane adds a pass
   banner carrying the definition's score narrative) or `failed (run
   ended)` (a `run.ended` landed with no preceding pass — derived by
   `sim.ExerciseOutcome` over replica facts, dual-sourced with the status
   poll's `Ended` exactly like the postmortem posture, `Model.runEnded`).
   The exercise's own outcome is distinct from but consistent with the
   world's postmortem (`overlays/postmortem.md`).

## Behavior

- **Scenario-cadence narration trigger**: while a scenario is armed, the
  chronicle narrator ([[chronicle]]) gains an ADDITIONAL narration trigger
  at the exercise's own pass/fail boundary (`internal/mind/narrate.go`
  `chronicleNote`: `curriculum.exercise_passed` always closes a chapter;
  `run.ended` closes one only when the scenario is armed) — independent of
  the day/night chapter cadence, so a short scenario run still produces at
  least one chronicle entry narrating its outcome. Additive: the day/night
  cadence is unchanged for ambient worlds and continues alongside this
  trigger for scenario worlds.
- **Ceremony-trigger linkage**: a landed `curriculum.exercise_passed`
  followed by `curriculum.stage_unlocked` is exactly `overlays/ceremony.md`'s
  takeover trigger — this panel and the ceremony overlay read the same two
  event types; the exercise tab is the *ambient* progress view, the
  ceremony is the *celebratory* takeover the same evidence earns (shipped —
  spec 056/TASK-127; no coupling here, and since spec 072 both surfaces
  also share the same `sim.EvaluateRubric` verdict source).

## The exercise catalog (spec 077 — nine exercises, 3/2/2/2 by stage)

`sim.ScenarioExercises` (`internal/sim/curriculum.go`) is the single source
this panel, `promptworld new --scenario`, and `world.ValidScenarioExercise`
all derive from — no second hand-maintained list anywhere. Seeds
46101–46109, unique; boundaries are dawn-shaped (`BoundaryDay` N = dawn of
day N; 0 = rolling, every dawn from day 2).

| ID | Stage | Seed | Boundary | Schedule |
|---|---|---|---|---|
| `first-night` | stage-1 | 46101 | day 2 | gru_emerges d1 22:00 |
| `cold-dawn` | stage-1 | 46103 | day 2 | cold_snap d1 22:00, 8h |
| `stranger-at-the-gate` | stage-1 | 46104 | day 2 | stranger_arrives d1 23:00 |
| `the-law` | stage-2 | 46102 | rolling | — |
| `blighted-larder` | stage-2 | 46105 | day 4 | forage_blight d2 08:00, r4 |
| `toolsmith` | stage-3 | 46106 | rolling | — |
| `fog-watch` | stage-3 | 46107 | day 3 | cold_snap d1 22:00 6h · gru_emerges d2 22:00 |
| `long-winter` | stage-4 | 46108 | day 4 | cold_snap 8h · forage_blight r4 · stranger_arrives · gru_emerges (days 1–3) |
| `stewards-charge` | stage-4 | 46109 | rolling | — |

## Incident vocabulary (spec 077 — four kinds, kind-specific params)

The closed `IncidentScheduleEntry` kind vocabulary the forecast line
renders (`incidentNoun`, `exerciseIncidentLine`). Every kind lands as
reducer-valid events indistinguishable from an ambient cause — no
authored/scenario marker in any payload; ambient dice paths are TASK-28's
recorded scope.

| Kind | Params | Emits | Surfaced as |
|---|---|---|---|
| `gru_emerges` | X,Y (border tile) | `gru.emerged` | "the gru emerges" |
| `cold_snap` | Hours [1,24] | `sim.cold_snap` | "a cold snap" |
| `forage_blight` | X,Y + Radius [1,8] | `sim.forage_blighted` | "the forage blights" |
| `stranger_arrives` | X,Y (border tile) | `stranger.arrived` (+ `stranger.moved/took/departed` from the entity) | "a stranger arrives" |

## Stage defaults

Per `patterns/stage-defaults.md` (the authority): this tab's presence is
**world-shaped, not stage-shaped** — present whenever the attached world
carries a `Manifest.Scenario` block (`Model.exerciseID`, validated against
the compiled catalog), regardless of `Stage`. The incident-visibility
vocabulary value IS stage-keyed: `forecast` at stages 1–2 and pre-ladder,
`fog` from stage 3; a definition override wins (D4).

## Narrow behavior

No narrow-specific rendering: the exercise tab is reachable as a solo/narrow
pane exactly like every other dock tab (`exerciseView`, the villagersView
shape; `patterns/layout.md` ruling b; [pages/solo-views.md](../pages/solo-views.md)).
The attach-time briefing and rubric gauges render at narrow width using the
same column-dropping discipline every other dock tab follows.

## Linear-stream / CLI projection (D1)

The rubric's underlying events are cataloged event types on the raw log —
an `attach`/`tail` client sees every rubric-relevant event as it lands, and
`promptworld status` surfaces the exercise's outcome model-free
(`WorldStatus.ScenarioExercise`/`ScenarioOutcome`, the same posture
`Status.Stage`/`StageOverridden` already have; `scenarioStatusLine` renders
the human line). No exercise-tab content is TUI-exclusive.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| exercise tab (dock selection) | absent · selected · solo | `Manifest.Scenario` via `Model.exerciseID` | `dockTabsRow`, `tabsView` (narrow) | `6` (again: solo; again: home; inert on ambient worlds) · — | spec 054 (reorient D11) | — |
| attach-time briefing | shown once per attach · dismissed | `sim.ExerciseDefinition.Framing`, `sim.IncidentVisibilityFor` | `exerciseBriefingBody` | any key (consumed, one press, this tab only) · — | spec 054 (reorient D11) | — |
| title line | in progress · passed · failed (run ended) | `sim.ExerciseOutcome` over the replica + `Model.runEnded` | `exerciseBody`, `exerciseOutcomeLabel` | — | spec 054 (reorient D11) | — |
| rubric gauge row | met · pending | `sim.EvaluateRubric` (replica state facts + counts) | `exerciseBody` | — | spec 054 (reorient D11) | — |
| incident-visibility line | forecast · omitted (fog) | `sim.ExerciseDefinition.Schedule`, `sim.IncidentVisibilityFor` (`patterns/stage-defaults.md`) | `exerciseIncidentLine` | — | spec 054 (reorient D4) | — |
| pass/fail banner | none (in progress) · passed · failed (run ended) | `curriculum.exercise_passed` / `run.ended` folded into the replica | `exerciseBody` | — | spec 054 (reorient D11) | — |
| scenario-cadence narration trigger | fires at pass/fail boundary | exercise outcome events | `internal/mind/narrate.go` `chronicleNote` (background, not a UI control) | — | spec 054 (reorient D11) | — |

**Parity rollout**: the tab key (`6`) and the briefing's any-key dismiss are
keyboard-only — recorded here as parity gaps from birth (decision 8, formal
doctrine in `patterns/keymap.md`), as this page promised while unbuilt: a
mouse target for tab selection and briefing dismissal graduates them out of
this note when it lands.
