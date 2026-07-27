---
name: report-card-renderer
description: The shared rubric-checklist renderer (D5) and the spec-072 fact resolver — exported replica-parametric by spec 076 (tui.ResolveRubricFacts/RecordedPassFor/RenderReportCard): reused by the postmortem, the ceremony, the guardian console, and promptworld compare's duel (the fourth consumer); plus the help overlay's ceremony-replay section and the skin's per-stage D6 ceremony chapter. Read when touching views.go, reportcard.go, help.go, or skin.go's ceremony-chapter table.
kind: component
sources:
  - internal/tui/views.go
  - internal/tui/reportcard.go
  - internal/tui/help.go
  - internal/skin/skin.go
verified_against: 657c770f87404b936a0587db1f6b00e81b9f0ee6
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
guardian-console card seam — spec 056 shipped the renderer and the wrapper
(`type reportCard struct{title, facts, mode}`, `var _ consoleCard =
reportCard{}` the compile-time proof), [[grounded-feedback]] (TASK-115)
the production wiring into `Model.consoleCards`.

- `reportCardFact{Term, Met, Backing}` is one already-resolved rubric row —
  facts are computed once at open time by the shared resolver, never inside
  the renderer, so its output is identical at every call site by construction.
- `reportCardMode` selects the marker vocabulary only: `reportCardConcluded`
  (met `✓` / missed `✗` — a run or exercise that's over) vs `reportCardLive`
  (met `✓` / pending `…` — still running); the same rows and backing
  references either way.
- `reportCardView(title, facts, mode, width)` is the one renderer every
  site calls — a bordered box, one line per fact, `"✓/✗/… <term> (<backing>)"`.

## The shared fact resolver (spec 072 — report-card truth)

`ResolveRubricFacts(state, def, pass)` (`internal/tui/reportcard.go` —
exported replica-parametric by spec 076; `Model.resolveReportCardFacts` and
`Model.recordedPassFor` are now thin wrappers passing `m.replica`, the old
`runEnded()` conjunct folded into the state-driven `s.Ended` read) is the
ONE precedence switch every card surface derives facts through — the
postmortem (`postmortemReportCard`), the ceremony (`ceremonyReportCardFor`),
the console checklist (`buildChecklistCard`), and (the fourth consumer)
`promptworld compare`'s duel scoreboard ([[world-forking]]), which reaches
it through the exported surface: `ReportCardFact`/`ReportCardMode` type
aliases with `ReportCardConcluded`/`ReportCardLive` consts,
`RecordedPassFor(state, exercise)` for the instrument lookup, and
`RenderReportCard(title, facts, mode, width)` wrapping `reportCardView` —
the duel card and the postmortem card are the same artifact by
construction, and a second precedence switch anywhere is a spec violation:

1. **Recorded pass** (`RecordedPassFor`) — the instrument: every term
   renders met (the emitter only fires when every term held; re-read, never
   re-grade), backing preferring the pass's own `Evidence` refs
   (`reportCardFactsFromPass`, `"<type> · seq <n>"`, first match by event
   type).
2. **Run ended** (`s.Ended`) — concluded `sim.EvaluateRubric` facts over the
   state (`reportCardFactsFromRubric`): a failed term renders `✗` with its
   honest backing count (`"agent.died: 2"` on a two-death run), never a
   presence-derived `✓`.
3. **Still running** — the same rubric facts in live (`…` pending) mode.

Labels are always `sim.RubricTerm.Label`'s hand-authored plain language
("no villager dies") — the evaluator is the same pure derivation the pass
emitter and the exercise tab's gauges read ([[scenario-machinery]]), so no
two surfaces can disagree. A nil state yields no facts (no card). The
pre-072 presence-based builders (`reportCardFactsFromCounts/FromEvents/
FromEvidence`) and the mechanical `humanizeEventType` gloss are deleted:
they graded a term met the first time its event type appeared at all —
backwards for zero-wanted terms (two deaths rendered `✓ agent died`).
The ceremony's aged-out-pass edge falls back to path 2 (concluded rubric
over current state), a pinned decision rather than an accident.

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
