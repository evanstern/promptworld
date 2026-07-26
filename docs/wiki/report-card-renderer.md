---
name: report-card-renderer
description: The shared rubric-checklist renderer (D5): reportCardFact/reportCardView/reportCardMode and the two fact-builders, reused by the postmortem, the ceremony, and (via consoleCard) the guardian console; the help overlay's ceremony-replay section reusing the same helpers; and the skin's per-stage D6 ceremony chapter. Split from [[takeover-surfaces]]; read when touching views.go, help.go, or skin.go's ceremony-chapter table.
kind: component
sources:
  - internal/tui/views.go
  - internal/tui/help.go
  - internal/skin/skin.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# The shared report-card renderer (D5)

Split from [[takeover-surfaces]] (corpus-spec v2 size-budget split,
summary-style): this note covers the shared rubric-checklist renderer spec
056 (D5) shipped, and the help overlay's ceremony-replay section that reuses
it. See [[takeover-surfaces]] for the ceremony/postmortem takeovers
themselves.

## The renderer

One rubric-checklist implementation, several sites: the postmortem
(concluded — a finished run), the ceremony (concluded — "the instrument
authoritative", FR-019), and, via a `consoleCard` wrapper, spec 053's
guardian-console card seam — spec 056 shipped only the renderer and the
wrapper (`type reportCard struct{title, facts, mode}`, `var _ consoleCard =
reportCard{}` the compile-time proof), leaving production wiring into
`Model.consoleCards` for [[grounded-feedback]] (TASK-115) to complete.

- `reportCardFact{Term, Met, Backing}` is one already-resolved rubric row —
  facts are computed once at open time by small helpers, never inside the
  renderer, so its output is identical at every call site by construction.
- `reportCardMode` selects the marker vocabulary only: `reportCardConcluded`
  (met `✓` / missed `✗` — a run or exercise that's over) vs `reportCardLive`
  (met `✓` / pending `…` — still running); the same rows and backing
  references either way.
- `reportCardView(title, facts, mode, width)` is the one renderer every
  site calls — a bordered box, one line per fact, `"✓/✗/… <term> (<backing>)"`.
- `reportCardFactsFromEvents`/`reportCardFactsFromEvidence` derive facts
  from, respectively, the client's own bounded chronicle ring (the fallback
  source) or a recorded `CurriculumPass`'s own `Evidence` list (the
  ceremony's preferred, authoritative source — exactly the satisfying
  events `sim.EvaluateUnlock` itself read). `humanizeEventType` is a
  deliberately generic mechanical gloss ("agent.died" → "agent died") —
  placeholder copy pending curated per-term prose from a future scenario
  rubric evolution ([[scenario-machinery]]).

## Replayability

**Replayability** (`help.go`'s `?` overlay, a fourth section): every stage
`replica.StagesUnlocked` names gets a `ceremonyReplayLines` entry —
re-rendering the SAME chapter + report card the live ceremony showed
(shared helpers, so the two can never disagree), stored/derived rather than
regenerated. A missed or dismissed ceremony is never permanently lost. The
overlay's section enum gains `helpSectionCeremonies` (title "ceremonies"),
reached by `tab`/`shift+tab` like every other section; a nil/empty replica
degrades to an honest "no stage has unlocked yet in this world" placeholder.

## The skin's ceremony chapter

**The skin's ceremony chapter** (`internal/skin/skin.go`): `defaultCeremonyChapters`
is one authored D6 "your play proved <identity>" line per UNLOCKABLE stage
(stage-2 through stage-4 — stage-1 is the ladder's floor and is never
unlocked, so it carries no entry, nothing ever resolves it); `Skin.
CeremonyChapter(stage)` is a plain token lookup (`skin.stage.<id>.
ceremony_chapter`) like every other fiction string, so a world skin.json may
re-voice it per stage. Deliberately generic "your play" phrasing throughout,
since the gate a pass satisfies varies by stage (any pass at stage-1; a
charter revision at stage-2; a player-granted tool at stage-3) — the
chapter's subject stays true regardless of which specific evidence earned
it.

## Back to parent

[[takeover-surfaces]] links here for the report-card renderer, its help-
overlay replay surface, and the skin's ceremony chapter; that note's own
Connections section lists [[scenario-machinery]] as the rubric data source
both the live and replayed report cards read, and [[skin]] as the source of
every other fiction string the ceremony/postmortem pages render.
