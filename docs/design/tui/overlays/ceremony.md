---
title: Overlay — unlock ceremony
class: overlay
status: specified
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
---

# Overlay: unlock ceremony

The stage-unlock takeover (reorientation decision 6): when
`curriculum.stage_unlocked` fires on an attached client, the ceremony seizes
the screen immediately — the operator's chosen maximum-salience interrupt
policy, celebrating a milestone the player earned. **Not built** — specified
spec-before-build for Wave 4.

## Mockup

```
┌ THE WRITTEN WORD — unlocked ──────────────────────────────────────────┐
│                                                                        │
│  Your charter proved The Written Word: a law that outlives the        │
│  conversation, written once and honored by every turn since.          │
│                                                                        │
│  ┌─ report card · the-law ──────────────────────────────────────┐    │
│  │ ✓ norm adopted under a player-authored charter revision       │    │
│  │ ✓ charter_observed{custom: true} recorded before the vote      │    │
│  └────────────────────────────────────────────────────────────────┘   │
│                                                                        │
│                              esc dismiss · q — the world keeps running│
└────────────────────────────────────────────────────────────────────────┘
```

## Layering & trigger

- **Trigger**: `curriculum.stage_unlocked`, on every attached client, the
  instant the event lands — no confirmation, no queue (decision 6:
  "the operator chose maximum salience over the never-interrupt-play rule").
- **Takeover**: body-replacement in the solo-zoom slot, chrome (header/
  minibuffer/footer) stays visible, output remains exactly terminal-height
  — the same discipline every other body-replacement surface in this corpus
  follows (`overlays/help.md`'s `helpPanelView` precedent).
- **Non-stacking + precedence vs. postmortem**: takeovers never stack. If
  `run.ended` fires while this ceremony is already open, the ceremony is
  dismissed immediately and replaced by `overlays/postmortem.md` — a
  run-ended state is more final and encompassing than an in-flight
  celebration, so **postmortem always wins**. If (the far rarer ordering)
  a stage-unlock fires while a postmortem is already open, the ceremony
  trigger is deferred rather than interrupting the postmortem — postmortem
  still wins, and the deferred ceremony remains fully reachable afterward
  via the replay surfaces below (this precedence rule is stated identically
  on `overlays/postmortem.md`).

## Content: two voices, instrument authoritative (FR-019)

- **Narrated chapter** — a short, skin-resolved paragraph in the **player's
  own authorship voice** (D6): "your charter proved `<stage name>`," never a
  third-person system notice. Fiction strings render as skin tokens
  (`patterns/skin-tokens.md`).
- **Rubric checklist (the instrument, authoritative)** — the SAME report-
  card artifact `overlays/postmortem.md` and `pages/guardian-console.md`
  render (D5's one shared renderer, three sites): the exercise's rubric
  terms that earned the unlock, rendered as a plain checklist. When the
  narrated chapter and the checklist could ever be read as disagreeing, the
  checklist is the source of truth — the narration is presentation, never
  a second scoring computation.

Both always render together; this is not a toggle — FR-019 rules "both,
instrument authoritative," not "narration by default, instrument on
request."

## Dismissal and the blessed stopping point (D13)

- `esc` — dismiss one layer, same as every other overlay; returns to
  whatever was beneath.
- `q` — the ceremony's own **blessed stopping point** (D13): detaches
  directly from the ceremony, rendering the same "the world keeps running"
  affordance every detach carries. A milestone moment is a deliberately
  natural place to stop playing, not a forced continue.

## Replayability (explicit acceptance criterion, FR-013)

Replayable from **both** pull surfaces, independently:

1. **`promptworld stages`** — already surfaces the earned stage, the
   proving world, and the evidence pointer ([[curriculum-ladder]]) —
   enough facts to identify which ceremony to revisit, model-free, from the
   CLI.
2. **`?` overlay** — `overlays/help.md` gains a ceremony-replay entry point
   (research.md R4 classifies this section `status-derived`: which
   ceremonies exist depends on run history; replayed content is stored, not
   regenerated) — **not built** in this slice; recorded here as the pull
   surface this page's replayability AC depends on, with its own
   `unbuilt (wave 4)` row in the control table below.

Both surfaces existing is the explicit AC (spec.md US2-AS2): a player who
missed or dismissed a ceremony is never permanently denied its content.

## Interrupt-policy watch item

Recorded per the reorientation's own open question 5 (carried forward, not
resolved here): decision 6 stands as designed, but the reopening signal is
named explicitly — **playtest evidence of ceremony fatigue, or mid-crisis
seizure complaints** (a ceremony firing while the player is mid-emergency
elsewhere in the village) — would reopen this interrupt policy for revision.
Nothing in this feature acts on that signal; it is a watch item, not a
predicted outcome.

## Linear-stream / CLI projection (D1)

`curriculum.stage_unlocked` already renders on the raw chronicle feed
(digest: "the village's watcher earned `<stage>`") and as a narrated
chronicle chapter line ([[chronicle]]) — an `attach`/`tail` client never
misses the milestone; `promptworld stages` carries the durable facts. This
overlay is a celebratory TUI presentation layer over facts already visible
to a non-visual observer.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| ceremony takeover | closed · open | `curriculum.stage_unlocked` | `unbuilt (wave 4)` | (opens automatically) · — | reorient decision 6 | — |
| narrated chapter | — | skin-resolved unlock text (D6 voice) | `unbuilt (wave 4)` | — (display-only) | reorient D6/FR-019 | `skin.guardian.name` (stage-name resolution reuses `skin.StageName`, not a new token) |
| rubric checklist (instrument) | — | exercise rubric evidence (shared with report card, D5) | `unbuilt (wave 4)`, shared with `overlays/postmortem.md` | — | reorient FR-019/D5 | — |
| dismiss | open → closed | player action | `unbuilt (wave 4)` | `esc` · — | reorient decision 6 | — |
| q-detach (blessed stopping point) | open → detached | player action | `unbuilt (wave 4)` | `q` · — | reorient D13 | — |
| replay via `stages` | — | per-user unlocks record | `unbuilt (wave 4)` (CLI already exposes the facts today) | — (CLI, no TUI keys) | reorient FR-013 | — |
| replay via `?` overlay | — | run-history-derived ceremony list | `unbuilt (wave 4)` | — | reorient FR-013 / research.md R4 | — |

**Parity rollout**: `esc`/`q` have no mouse target — recorded from birth as
a parity gap (decision 8, formal doctrine in `patterns/keymap.md`, T024).
