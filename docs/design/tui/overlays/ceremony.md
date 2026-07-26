---
title: Overlay — unlock ceremony
class: overlay
status: shipped
verified_against: 85f45a121e538d93048f52a1ad22472284ec0938
sources:
  - internal/tui/tui.go
  - internal/tui/views.go
  - internal/skin/skin.go
---

# Overlay: unlock ceremony

The stage-unlock takeover (reorientation decision 6, spec 056/TASK-127):
when `curriculum.stage_unlocked` fires on an attached client, the ceremony
seizes the screen immediately — the operator's chosen maximum-salience
interrupt policy, celebrating a milestone the player earned.

## Mockup (real symbols — `ceremonyView`, stage-3 unlocked by `the-law`)

```
┌ THE CRAFT — unlocked ──────────────────────────────────────────────────┐
│                                                                        │
│  Your play proved The Craft: what the guardian can do now bears your  │
│  own hand in its shaping.                                             │
│                                                                        │
│  ┌─ report card · the-law ──────────────────────────────────────┐    │
│  │ ✓ meeting proposal resolved (meeting.proposal_resolved: 1)     │    │
│  │ ✓ metatron charter observed (metatron.charter_observed: 1)     │    │
│  └────────────────────────────────────────────────────────────────┘   │
│                                                                        │
│                              esc dismiss · q — the world keeps running│
└────────────────────────────────────────────────────────────────────────┘
```

The title and chapter are always the STAGE BEING UNLOCKED's identity
(`replica.StagesUnlocked`'s last entry, `skin.StageName`/`skin.
CeremonyChapter`) — not the stage the proving exercise was played at. A
`first-night` (stage-1) pass unlocking stage-2 renders "THE WRITTEN WORD —
unlocked"; a `the-law` (stage-2) pass unlocking stage-3 (shown above)
renders "THE CRAFT — unlocked".

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

- **Narrated chapter** (`skin.CeremonyChapter(stage)`, tokens
  `skin.stage.<id>.ceremony_chapter`) — a short, skin-resolved paragraph in
  the **player's own authorship voice** (D6): "your play proved `<stage
  name>`," never a third-person system notice. A deliberately generic
  "your play" framing (not "your charter") is used across all three
  unlockable stages: the gate a pass satisfies varies (any pass at stage-1;
  a charter revision at stage-2; a player-granted tool at stage-3), so one
  authored line per stage stays true regardless of which specific evidence
  earned it. Fiction strings render as skin tokens (`patterns/skin-tokens.md`).
- **Rubric checklist (the instrument, authoritative)** — the SAME
  `reportCardView` renderer `overlays/postmortem.md` and
  `pages/guardian-console.md` compose (D5's one shared renderer, three
  sites): the proving exercise's rubric terms, rendered as a plain
  checklist. Content comes from the recorded `CurriculumPass`'s own
  `Evidence` when the pass is still retained (`provingPass`,
  `internal/tui/views.go`) — the authoritative source FR-019 calls "the
  instrument" — falling back to a generic events-ring scan only if the pass
  has aged out of the bounded 32-entry retention.

Both always render together; this is not a toggle — FR-019 rules "both,
instrument authoritative," not "narration by default, instrument on
request."

**Known simplification** (not resolved by this feature): the checklist's
per-term marker is a generic event-presence evaluation (met = the term's
cataloged event type appears at least once in the evidence/ring) — it does
not encode a term's actual pass semantics (some rubric terms want ZERO
occurrences, e.g. `agent.died` on `first-night`; some combine several terms
with OR). `internal/sim/curriculum.go`'s own doc comments name TASK-119's
scenario rubric machinery as the eventual owner of curated per-term
semantics; this renderer is built to stay correct regardless of how that
content evolves (`reportCardFactsFromEvidence`/`reportCardFactsFromEvents`,
`internal/tui/views.go`).

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
2. **`?` overlay** — `overlays/help.md`'s ceremony-replay section
   (`helpSectionCeremonies`, `internal/tui/help.go`; research.md R4
   classifies it `status-derived`: which ceremonies exist depends on run
   history) lists every stage `replica.StagesUnlocked` names, each
   re-rendering the SAME chapter + report card the live ceremony showed
   (`ceremonyReplayLines`) — stored, never regenerated.

Both surfaces existing is the explicit AC (spec.md US2-AS2): a player who
missed or dismissed a ceremony is never permanently denied its content.

## Narrow behavior

Takes over the full screen in narrow exactly as in widescreen — takeovers
are layout-independent (`patterns/layout.md` ruling b). No narrow-specific
rendering exists or is needed: the mockup above applies unchanged below the
112-column breakpoint (content reflows to the narrower width using the same
wrapping every other body-replacement surface in this corpus uses,
`overlays/help.md`'s `helpPanelView` precedent).

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
| ceremony takeover | closed · open | `curriculum.stage_unlocked` (`Model.takeover`, `applyEvent`) | `takeoverView`/`ceremonyView` (`internal/tui/views.go`) | (opens automatically) · — | reorient decision 6 | — |
| narrated chapter | — | `replica.StagesUnlocked`'s last entry + skin lookup | `ceremonyView`, `skin.CeremonyChapter` | — (display-only) | reorient D6/FR-019 | `skin.stage.<id>.ceremony_chapter` |
| rubric checklist (instrument) | — | recorded pass Evidence (`provingPass`) or ring scan (fallback) | `reportCardView`, shared with `overlays/postmortem.md`/`pages/guardian-console.md` | — | reorient FR-019/D5 | — |
| dismiss | open → closed | player action | `handleTakeoverKey` (`internal/tui/tui.go`) | `esc` · — | reorient decision 6 | — |
| q-detach (blessed stopping point) | open → detached | player action | `handleTakeoverKey` → `quit()` (framing from `View()`'s `runEnded()` check) | `q` · — | reorient D13 | — |
| replay via `stages` | — | per-user unlocks record | `cmd/promptworld/stages.go` (unchanged, pre-existing) | — (CLI, no TUI keys) | reorient FR-013 | — |
| replay via `?` overlay | — | `replica.StagesUnlocked`/`CurriculumPasses` | `ceremonyReplayLines` (`internal/tui/help.go`, `helpSectionCeremonies`) | `tab`/`shift+tab` to reach · — | reorient FR-013 / research.md R4 | — |

**Parity rollout**: `esc`/`q` have no mouse target — recorded from birth as
a parity gap (decision 8, formal doctrine in `patterns/keymap.md`, T024).
