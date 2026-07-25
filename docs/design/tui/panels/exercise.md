---
title: Panel — exercise (scenario tab)
class: panel
status: specified
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
---

# Panel: exercise

The fourth dock tab (reorientation D11), present only on scenario worlds:
the framing, live rubric progress, and pass/fail state of the world's
seeded exercise (`sim.ExerciseDefinition`, [[curriculum-ladder]]). **Not
built** — specified spec-before-build for Wave 4 (TASK-119).

## Mockup — attach-time briefing (first render this attach)

```
┌─ chronicle │ {{skin.guardian.tab_label}} │ villagers │ EXERCISE ┐
├──────────────────────────────────────────────────────────────────┤
│ FIRST NIGHT                                                       │
│                                                                    │
│ Keep the village alive through night one by directing the         │
│ {{skin.guardian.epithet}}: visions, omens, and the watch.          │
│                                                                    │
│ Incidents ahead are forecast — you'll see the schedule before      │
│ they land (stage 1).                                               │
│                                                                    │
│                                          any key — begin           │
└──────────────────────────────────────────────────────────────────┘
```

## Mockup — live gauges (after the briefing is dismissed)

```
┌─ chronicle │ {{skin.guardian.tab_label}} │ villagers │ EXERCISE ┐
├──────────────────────────────────────────────────────────────────┤
│ FIRST NIGHT · in progress                                         │
│                                                                    │
│ ✓ village survives to dawn         (agent.died: 0 so far)         │
│ … watch placed before nightfall     (metatron.order_placed: —)    │
│                                                                    │
│ incidents (forecast): gru sighted ~22:00 · cold snap ~03:00        │
└──────────────────────────────────────────────────────────────────┘
```

## Structure

1. **Framing line** — the exercise's name and `sim.ExerciseDefinition
   .Framing` text, rendered plainly (the same "failure is a story, not a
   scold" framing the chronicle score-narrative already carries,
   [[curriculum-ladder]]).
2. **Attach-time briefing** (the "Portal safe-room" reference) — shown once
   per client attach, before the live gauges: framing + the current
   incident-visibility mode for this stage (below). Dismissed by any key,
   then never reshown for that attach (re-attaching shows it again — a
   fresh orientation each session, not a one-time-ever notice).
3. **Rubric gauges** — one row per rubric term; every term is, by
   construction, a cataloged event type ([[curriculum-ladder]]:
   "an event-derived rubric whose every term must be a cataloged event
   type"). Each gauge shows the term in plain language, a met/pending
   marker, and the event count/condition backing it (engineer-facing detail,
   FR-020 — plain-language callout is the term itself; the raw event
   reference sits alongside it, not instead of it).
4. **Incident-visibility vocabulary** (D4) — a **vocabulary, not a
   boolean**, per `patterns/stage-defaults.md` (the authority): `forecast`
   (the incident schedule is shown ahead of time, stages 1–2 and
   pre-ladder) or `fog` (the schedule is hidden — incidents are revealed
   only as they happen, stage 3+). The gauges section never changes shape
   between the two; only whether the "incidents ahead" line is populated
   or omitted changes.
5. **Pass/fail state** — `in progress` (default) → `passed` (the
   `curriculum.exercise_passed` event landed — the pane switches to a
   pass banner) or `failed (run ended)` (a `run.ended` event landed with no
   preceding pass — the exercise's own outcome, distinct from but consistent
   with the world's postmortem, `overlays/postmortem.md`).

## Behavior

- **Scenario-cadence narration trigger**: while this tab's exercise is
  active, the chronicle narrator ([[chronicle]]) gains an ADDITIONAL
  narration trigger at the exercise's own pass/fail boundary — independent
  of the day/night chapter cadence — so a short scenario run (which might
  span less than one ~2-chapters/game-day cadence window) still produces at
  least one chronicle entry narrating its outcome, rather than zero. This is
  additive: the day/night cadence is unchanged for ambient worlds and
  continues alongside this trigger for scenario worlds.
- **Ceremony-trigger linkage**: a landed `curriculum.exercise_passed`
  followed by `curriculum.stage_unlocked` is exactly `overlays/ceremony.md`'s
  takeover trigger — this panel and the ceremony overlay read the same two
  event types; the exercise tab is the *ambient* progress view, the
  ceremony is the *celebratory* takeover the same evidence earns.

## Stage defaults

Per `patterns/stage-defaults.md` (the authority): this tab's presence is
**world-shaped, not stage-shaped** — present whenever the attached world
carries a `Manifest.Scenario` block, regardless of `Stage`. The
incident-visibility vocabulary value IS stage-keyed: `forecast` at stages
1–2 and pre-ladder, `fog` from stage 3.

## Narrow behavior

No narrow-specific rendering: the exercise tab is reachable as a solo/narrow
pane exactly like every other dock tab (`patterns/layout.md` ruling b —
"guardian console / systems tab / exercise panel: reachable as solo views…
no new narrow-specific rendering"; [pages/solo-views.md](../pages/solo-views.md)).
The attach-time briefing and rubric gauges render at narrow width using the
same column-dropping discipline every other dock tab follows.

## Linear-stream / CLI projection (D1)

The rubric's underlying events are cataloged event types on the raw log —
an `attach`/`tail` client sees every rubric-relevant event as it lands, and
`promptworld status` can surface the exercise's pass/fail state model-free
(the same posture `Status.Stage`/`StageOverridden` already have,
[[curriculum-ladder]]). No exercise-tab content is TUI-exclusive.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| exercise tab (dock selection) | — | world carries `Manifest.Scenario` | `unbuilt (wave 4)` — no 4th dock-tab key exists yet | new tab key (TBD by TASK-119) · — | reorient D11 | — |
| attach-time briefing | shown once per attach · dismissed | `sim.ExerciseDefinition.Framing` | `unbuilt (wave 4)` | any key · — | reorient D11 | `skin.guardian.epithet` |
| framing line | — | `sim.ExerciseDefinition` | `unbuilt (wave 4)` | — | reorient D11 | — |
| rubric gauge row | met · pending | cataloged event types (rubric terms) | `unbuilt (wave 4)` | — | reorient D11 | — |
| incident-visibility line | forecast · fog · omitted (n/a) | stage-keyed vocabulary (`patterns/stage-defaults.md`) | `unbuilt (wave 4)` | — | reorient D4 | — |
| pass/fail banner | in progress · passed · failed (run ended) | `curriculum.exercise_passed` / `run.ended` | `unbuilt (wave 4)` | — | reorient D11 | — |
| scenario-cadence narration trigger | fires at pass/fail boundary | exercise outcome events | `unbuilt (wave 4)`, [[chronicle]]'s narrator worker | — (background, not a UI control) | reorient D11 | — |

**Parity rollout**: this page is entirely unbuilt, so its eventual tab-key
and briefing-dismiss key are recorded here from birth as needing a mouse
target once implemented (decision 8, formal doctrine in
`patterns/keymap.md`, T024) rather than being a gap discovered later.
