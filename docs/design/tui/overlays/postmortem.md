---
title: Overlay — postmortem
class: overlay
status: specified
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
---

# Overlay: postmortem

The run-end takeover (reorientation decision 6): when `run.ended` fires on
an attached client, the postmortem seizes the screen immediately — the same
maximum-salience interrupt policy as the ceremony, at the opposite emotional
pole: the morgue's no-blame evidence register, never a scold. **Not
built** — specified spec-before-build for Wave 4.

## Mockup — ambient (unscored) world

```
┌ THE RUN HAS ENDED ─────────────────────────────────────────────────────┐
│                                                                        │
│  The last villager died of exposure. The village stands empty — the   │
│  run has ended.                                                       │
│                                                                        │
│  morgue — no-blame evidence                                           │
│  Ash · day 6 · exposure · charter observed: default                   │
│  Rowan · day 6 · exposure · charter observed: default                 │
│                                                                        │
│                                                    esc dismiss · q quit│
└──────────────────────────────────────────────────────────────────────┘
```

## Mockup — scored/scenario run

```
┌ THE RUN HAS ENDED ─────────────────────────────────────────────────────┐
│                                                                        │
│  The last villager died of exposure. The village stands empty — the   │
│  run has ended.                                                       │
│                                                                        │
│  ┌─ report card · first-night ──────────────────────────────────┐    │
│  │ ✗ village survives to dawn      (agent.died: 2)                │   │
│  │ ✓ watch placed before nightfall (metatron.order_placed: 1)     │   │
│  └──────────────────────────────────────────────────────────────┘    │
│                                                                        │
│  morgue — no-blame evidence                                           │
│  Ash · day 1 · exposure · charter observed: default                   │
│                                                                        │
│                                                    esc dismiss · q quit│
└──────────────────────────────────────────────────────────────────────┘
```

## Layering & trigger

- **Trigger**: `run.ended`, on every attached client, the instant the event
  lands. A client attaching AFTER the fact also sees it — `Model.runEnded()`
  is already dual-source ([[tui-client]]: the replica's `State.Ended` plus
  the pushed event plus the status poll), so a fresh attach to an already-
  ended world opens this takeover automatically on connect, not only on the
  live transition.
- **Takeover**: body-replacement, chrome stays visible, exactly
  terminal-height — the same discipline as every other takeover in this
  corpus.
- **Non-stacking + precedence vs. ceremony**: identical rule, stated on
  both pages for consistency — **postmortem always wins**. If a ceremony is
  open when `run.ended` lands, the ceremony is dismissed and replaced by
  this takeover; a stage-unlock arriving while this takeover is already open
  never interrupts it (deferred, replayable later via `overlays/
  ceremony.md`'s pull surfaces). See `overlays/ceremony.md` for the same
  rule from the other side.

## Content: the FR-018 ambient/scored boundary

The hybrid-scoring boundary stays crisp on screen — this is the ruling
FR-018 requires, restated precisely:

- **Ambient (unscored) world**: **morgue evidence only** — the same
  no-blame register `morgue.md` already renders on disk ([[morgue]]), now
  as a live TUI takeover: the run-ending factual line, then one row per
  death (name, day, cause, the charter-observation timeline entry closest
  to that death). **No report card renders** — there is nothing to score;
  an ambient world was never given a rubric.
- **Scored/scenario run**: the report card renders FIRST (the exercise's
  pass/fail outcome — sharing the exact same rubric-checklist renderer
  `pages/guardian-console.md`'s inline cards and `overlays/ceremony.md`'s
  instrument use, D5's one-renderer-three-sites rule), followed by the
  SAME morgue evidence section every world gets. Scored status never
  removes the morgue register — it adds the report card in front of it.

## Report card (D5 — the shared renderer, authored here)

One rubric-checklist renderer, three call sites:

1. **This page** — a scored run's postmortem, immediately after the
   narrated run-end line.
2. **`pages/guardian-console.md`** — inline cards at natural stopping
   points (run end, pause, exercise resolution).
3. **`overlays/ceremony.md`** — the "instrument authoritative" checklist
   inside an unlock celebration.

Content: one row per rubric term (from the exercise's event-derived rubric,
[[curriculum-ladder]]), a met/pending or met/missed marker depending on
call site (a still-running exercise shows met/pending; a concluded one
shows met/missed), and the backing event reference. The renderer is
identical everywhere it appears; only its surrounding takeover/inline/
celebratory chrome differs per call site.

## Dismissal

- `esc` — dismiss, return to whatever rendering would otherwise show (the
  header's `ENDED` posture and read-only clock keys persist regardless,
  per spec 044 — dismissing the takeover never re-enables clock control on
  an ended world).
- `q` — quits/detaches normally. **No special "world keeps running"
  stopping-point framing here** (unlike the ceremony's D13 framing): the
  run has already ended, so that reassurance would be actively misleading;
  `q` behaves exactly as the existing postmortem-posture footer already
  documents ("run ended (read-only)").

## Replayability (explicit acceptance criterion, FR-013)

Replayable from the **morgue**, as its own explicit AC: re-attaching a
client to an already-ended world re-opens this takeover automatically on
connect (above). It is also reachable **on demand** without a fresh
reconnect: while `Model.runEnded()` is true, a new global key **`p`**
(mnemonic: postmortem) reopens the takeover from any dismissed state,
anywhere in the client — the dedicated on-demand pull surface this page's
replayability AC requires, distinct from the ceremony's `?`/`stages` pair
(the postmortem's natural pull surface IS the morgue content itself, which
this key re-surfaces directly rather than through a separate lookup step).

## Linear-stream / CLI projection (D1)

`morgue.md` (rendered to the world's save directory by the scribe,
[[morgue]]) and the raw chronicle feed's `run.ended`/`agent.died`/
`morgue.epilogue` lines ([[chronicle]]) already carry every fact this
takeover shows, with no TUI dependency; `promptworld status` reports the
ended state model-free. This overlay adds no fact a linear observer lacks.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| postmortem takeover | closed · open | `run.ended` / `State.Ended` (dual-source, matching `Model.runEnded()`) | `unbuilt (wave 4)` | (opens automatically) · — | reorient decision 6 | — |
| run-end narrated line | — | `run.ended` payload (final cause) | `unbuilt (wave 4)`, shares wording with [[chronicle]]'s existing digest/narrated line | — (display-only) | reorient decision 6 (surface); TASK-60/spec 044 (line text, pre-existing) | — |
| morgue evidence rows | — | morgue no-blame register ([[morgue]]) | `unbuilt (wave 4)` (content pre-exists in `morgue.md`) | — | reorient decision 6 / spec 044 | — |
| report card (scored runs only) | absent (ambient) · shown (scored) | exercise rubric evidence | `unbuilt (wave 4)`, shared with `pages/guardian-console.md`/`overlays/ceremony.md` | — | reorient FR-018/D5 | — |
| dismiss | open → closed | player action | `unbuilt (wave 4)` | `esc` · — | reorient decision 6 | — |
| quit/detach | — | player action | existing quit path (unchanged) | `q` · — | pre-existing | — |
| replay via reopen key | closed → open | `Model.runEnded()` | `unbuilt (wave 4)` | `p` · — | reorient FR-013 | — |
| replay via reattach | — | `State.Ended` at connect | `unbuilt (wave 4)` (reuses existing `runEnded()` dual-source posture) | (automatic on connect) · — | reorient FR-013 | — |

**Parity rollout**: `esc`/`p` have no mouse target — recorded from birth as
a parity gap (decision 8, formal doctrine in `patterns/keymap.md`, T024).
