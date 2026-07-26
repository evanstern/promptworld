---
title: Analysis — Semantic Honesty Delta
aliases: [Post-Cockpit Docs Delta, Help-Stack Delta 2026-07-26]
tags: [analysis, player-docs, teaching-game, reference-honesty]
type: analysis
created: 2026-07-26
updated: 2026-07-26
related: ["[[Game-Player-Docs]]", "[[Manual-Structure-Conventions]]", "[[Explaining-The-Screen]]", "[[In-Game-Contextual-Help]]", "[[Quickstart-Guide-Patterns]]", "[[Analysis-UI-Reference-And-Help-Stack]]", "[[Analysis-In-Game-First-Teaching]]"]
---

# Analysis — Semantic Honesty Delta

_The re-run delta evaluation of 2026-07-26 (`docs/design/reorient-2026-07-26-ui.md`),
cross-grounded against the three sibling branch drafts (Game-UI-UX,
Game-Gameplay-Patterns, Learning-Game-Design). The 2026-07-25 course of action has
shipped in full; this note evaluates what the shipped help/docs layer still owes the
lens, reconciles this branch's findings with the siblings', and records the operator's
2026-07-26 steering decisions as fixed constraints. Prior layers
([[Analysis-In-Game-First-Teaching]], [[Analysis-UI-Reference-And-Help-Stack]]) are
immutable context; superseded framings are declared below, never edited._

## Constraints taken as given

The operator's steering decisions of 2026-07-26 (`## Decisions` items 1–10 in
`docs/design/reorient-2026-07-26-ui.md`) are FIXED. The ones binding this branch:

- **Decision 2** — the design-gate semantic lint (this branch's recommendation, taken):
  `scripts/check-tui-design.mjs` gains a warn/fail when a `status: shipped` page
  carries `unbuilt (wave` renderer cells; `overlays/postmortem.md`'s seven stale cells
  and `panels/exercise.md:110` are amended in the same PR.
- **Decision 1** — report-card truth unified on `sim.EvaluateRubric` (NEW card, HIGH),
  sequenced after TASK-144.
- **Decision 4** — TASK-142 look-cursor runs in the next UI lane, amended with the DF
  fixed tile-content hierarchy and the spec-068 tile registry's `meaning` rows as the
  tile pane's whatis content, plain-language per FR-020 (this branch's framing, taken
  jointly with the Game-UI-UX branch's).
- **Decision 6** — the forward ladder renders in the `?` guardian section
  (deterministic floor, model-free).
- **Decision 7** — `getting-started.html` gains the "ask your guardian one thing" step
  (sample ask from the `skin.guardian.example_ask.*` token family); each stage page
  gains a first-session do-this-then-this block (this branch's recommendation, taken).
- **Decision 8** — the mouse-parity sweep test is commissioned.
- **Decision 10** — parked as recorded watch items, genuinely open: player-docs pin
  de-churn; TASK-146 → Medium reassessment after TASK-145 merges; fire-once lesson
  doctrine (revisit at >15 lessons); teaching-signal instrumentation consent;
  chronicle/world search `/`; among others.

## Verdict

Every ranked recommendation in [[Analysis-UI-Reference-And-Help-Stack]] except the
badge deep-link is now a shipped, pinned, code-verified artifact — the design corpus v2
(spec 047/TASK-123), the extracted `overlays/help.md` with all five sections, the
lesson row (spec 055), the takeover pair with the replayability invariant as explicit
FR-013 ACs (spec 056), skin tokens (spec 052), the D9 guardian section teaching
example asks at the no-LLM floor (spec 063), and the NetHack-shaped player pages
(TASK-114: `understanding-the-screen.html` titled with the player's own question and
carrying the "you need not memorize" preface; the losing-is-fun paragraph in
`getting-started.html`). Cogmind's four layers ([[In-Game-Contextual-Help]]) all exist
in code.

**The delta gap this branch names is semantic honesty — and cross-grounding shows it is
one property with two faces, found independently by two branches.** This branch found
the reference lying about the code: `overlays/postmortem.md` carries `status: shipped`
and a fresh pin while seven control-table renderer cells still read `unbuilt (wave 4)`
— though `internal/tui/render_test.go:923` exercises `postmortemView`, `takeover_test.go`
tests the takeover family, and `tui.go:939` implements the `p` reopen key those cells
deny. The Learning-Game-Design branch found the screen lying about the world: the
shipped report-card renderer grades by generic event presence
(`reportCardFactsFromEvents`), so a first-night failure with two deaths renders
"✓ agent died (agent.died: 2)" — a checkmark on the failing outcome at the most
salient teaching moment — while the exercise tab grades by real semantics
(`sim.EvaluateRubric`). The Game-Gameplay-Patterns branch flagged the same inversion
as a known-simplification that "should not survive the next content wave." These are
the same corpus failure mode: convention-maintained truth rots; only derivation from
data holds ([[Manual-Structure-Conventions]] "Keeping reference honest" — CDDA's
generated-from-source guide vs hand-written wiki rot; [[_grounding]] § Cataclysm).
Decisions 2 and 1 respectively fix the two faces, and this note's job is to state the
unified property so future gates treat both as one class.

## The honesty property, unified

The corpus records exactly two mechanisms that keep game reference true: generate it
from the game's own data, or put the canonical answer in the game and let the doc
teach only the lookup key ([[Manual-Structure-Conventions]], [[Explaining-The-Screen]]
whatis delegation). promptworld's stack now applies both — but unevenly, and the two
shipped failures land precisely where neither mechanism reached:

1. **Doc → code (decision 2's face).** The keymap is un-rottable because
   `help_test.go` sweeps advertised rows against real handlers — generation-from-data.
   The control tables' *renderer* columns had no such check: `check-tui-design.mjs`
   validates pins and table headers, and INDEX.md rule 4 concedes semantic drift is "a
   same-PR review responsibility." Gate rule 3 ("flip status AND fill in real renderer
   symbols") was half-executed on postmortem.md — status flipped, pin bumped, cells
   left stale — and the structural gate passed. The lint closes the cheap 80%: a
   shipped page containing `unbuilt` cells is machine-detectable. A residue remains
   that no cell-lint catches: **stale ownership pointers in design prose** — both
   overlay pages' "known simplification" notes assigned the report-card fix to
   TASK-119, which is Done; the debt ended up owned by a closed task, which is how it
   survived until this re-run. Decision 1's new card repairs the instance; the class
   (prose that names a task as future owner outlives the task) is worth a review-
   checklist line in the gate rules when decision 2's PR touches INDEX.md.
2. **Screen → world (decision 1's face).** The report card is a *reference document
   rendered at runtime* — it is the postmortem's morgue-register text, the ceremony's
   evidence, the console card. Grading a failure ✓ violates the same rule as a manual
   that mis-documents a command, but at higher stakes: the Learning-Game-Design
   branch's grounding (histogram feedback works because it is true and
   evidence-derived; competence attribution predicts return) converges with this
   branch's [[Analysis-In-Game-First-Teaching]] tension note — "a charter-voiced angel
   confidently misexplaining the horizon would be anti-teaching" — which is why the
   explain tool was built registry-derived. The report card must sit on the same
   footing: one truth derivation (`EvaluateRubric`) for all four surfaces.

The mouse-parity sweep (decision 8, the Game-UI-UX branch's idea) belongs to the same
family and this branch endorses it in those terms: the v2 control tables were built to
be "the AI-parseable unit" precisely so sweeps like this could exist
([[Analysis-UI-Reference-And-Help-Stack]] § control tables). Every column that gains a
mechanical consumer stops being able to lie.

## Reconciliations with the siblings

- **Look-cursor as whatis (Game-UI-UX #1 + this branch's idea 4 → decision 4, joint
  framing confirmed).** The UI branch frames TASK-142 as the inspector chain's missing
  middle — the map is legible but not interrogable at a point (their Smallville/DF-`k`
  grounding). This branch supplies what the tile pane *says*: the spec-068 tile
  registry's glyph/name/meaning rows are NetHack's `/` whatis content
  ([[Explaining-The-Screen]]), making the look-cursor the third in-place lookup after
  `?` and the explain tool. The joint framing: TASK-142 serves lens (a) as
  interrogability and lens (c) as layer-2 help simultaneously, and
  `understanding-the-screen.html`'s "you need not memorize" preface — currently backed
  in-game only by the `?` overlay's glyph walkthrough — gains its promised
  point-and-ask answerer for map tiles. The DF fixed content hierarchy (agents →
  piles/chests → structures → terrain) is adopted; FR-020 keeps the pane
  plain-language.
- **In-TUI ladder view (Learning-Game-Design #2 → decision 6) lands in this branch's
  surface, and carries an invariant obligation.** The forward ladder renders in the
  `?` guardian section. One correction this branch owns: earned/next state reads the
  per-user unlocks record, which is *status-derived*, not static-per-stage — so
  `overlays/help.md`'s byte-identity classification table must gain a row (like the
  lessons registry: static catalog, live state columns) in the same PR. This is
  exactly the erosion path [[Analysis-UI-Reference-And-Help-Stack]] warned about
  ("ceremony replay + the lessons registry pull further in the dynamic direction"):
  each dynamic addition to the overlay is fine iff the no-LLM floor guarantee is
  restated deliberately, never silently.
- **Quickstart first-prompt (this branch → decision 7), sequenced against the exercise
  catalog (Game-Gameplay-Patterns #1 → decision 5).** The stage pages' first-session
  do-this-then-this blocks should, for scenario stages, walk the player into an
  exercise briefing — but the catalog wave will add 2–3 exercises per stage and
  three incident kinds. Author the blocks now against the shipped `first-night`
  content (content-only, per the decision), and expect one revision pass when the
  catalog lands; the DF quickstart's "minimal viable session" scoping rule
  ([[Quickstart-Guide-Patterns]]) says the block names the *smallest* exercise, not
  the catalog.
- **Search convergence (Game-UI-UX #5 + this branch's help-search idea → one parked
  doctrine).** The UI branch wants chronicle/world `/`; this branch wanted help-overlay
  search (Cogmind's manual is *searchable*, [[_grounding]] § Cogmind). Decision 10
  parks both under one watch item — correctly. Reconciled position: when the villager-
  count signal reopens it, search should be specified once, corpus-wide (chronicle
  lines, villager names, help sections, lesson catalog), not accreted per surface —
  the corpus's segment-and-tag rule applied to a single lookup grammar.
- **Instrumentation (Learning-Game-Design #4, parked).** Their finding — the ceremony-
  fatigue and fire-once watch items name reopening signals nobody can observe — is
  accepted as true and stays parked per decision 10. This branch adds why the parked
  state is *safe*: the push/pull invariant (every pushed teaching is pull-reachable —
  [[In-Game-Contextual-Help]]'s missed-message caution, enforced from the lesson row
  through FR-013) bounds the cost of unobservable push effectiveness. Unobservable
  *validity* (the report card) was not safe to park, and wasn't — decision 1.
- **No unreconcilable conflicts.** The nearest candidate — the Gameplay branch's
  metric-gaming mitigation (headline-live vs all-terms-live) against this branch's
  layer-4 transparency preference — was ruled by decision 10 (all-terms-live stands
  until multi-exercise content exists), and this branch accepts the ruling with the
  Gameplay branch's own reopening signal (playtest evidence of counter-optimization).

## Superseded framings, declared

- [[Analysis-UI-Reference-And-Help-Stack]]'s verdict clause "the remaining biggest gap
  is lens clause (d): the design reference is not yet the buildable authority it
  claims to be" is **superseded**: clause (d) passes as of spec 047 + the shipped gate.
  The successor gap is the semantic-honesty property above.
- That note's open questions are now resolved: audience→advanced-tier by FR-020
  (INDEX.md, plain-language default, debug *mode* not tier); vertical row budget and
  narrow fallback by `patterns/layout.md`'s shipped rulings (bodyMin, fold order:
  villager strip → lesson row → guardian strip last); the overlay no-LLM invariant by
  `overlays/help.md`'s byte-identity classification — which decision 6 now obligates
  to grow one row (above).
- [[Brief-and-Assumptions]]'s assumption that "the eventual consumer of this research
  is `docs/player/` and possibly in-app help" is superseded in emphasis, per
  [[Analysis-In-Game-First-Teaching]]'s ratified in-game-primary verdict — now fully
  built out.

## Still open from this branch (not covered by decisions 1–10)

- **Badge deep-link** — the one unshipped item of this branch's seven, still recorded
  honestly as `unbuilt (wave 4, layer-2)` in `overlays/help.md`'s control table, and
  named by no 2026-07-26 decision. Cheap, specified, unowned. Recommendation: ride the
  next UI lane alongside TASK-142 (both are layer-2 point-of-question help; decision
  4's lane is the natural home) rather than getting its own card.

## Parked items, kept open — with sharpenings from this pass

- **Player-docs pin de-churn** (decision 10). Evidence from this pass:
  `keys-reference.html` went stale on main *again* (spec 068's `help.md` re-pin, a
  tile-registry change that touched no key content), after TASK-130 fixed the
  identical staleness once. Sharpening: spec 069 (TASK-145, In Progress) moves the
  cost — the pr gate will BLOCK on `player-docs-stale`, so staleness can no longer
  reach main, but every TUI-touching PR now pays a regeneration it may not
  semantically need. The watch item's trigger is therefore concrete: if post-069
  sweeps show repeated no-content-change regenerations of the keys card, narrow the
  pins (section-hash or wiki-note-level) then.
- **TASK-146 (CAPSULES.md / corpus-spec v2)** — reassess after TASK-145, per decision
  10. Cross-grounding datum worth recording: **all four evaluators of this re-run
  independently chose INDEX just-in-time grounding and completed their deltas loading
  between zero and one full wiki notes each** — the INDEX capsules are already
  load-bearing routing artifacts. TASK-146's real product is making that mode official
  and budget-enforced; its weakest row today is `guardian.md`'s 1286-char capsule —
  the note for the pivot's central verb — and the doc-generation chain (player-docs,
  codebase-to-course) consumes the same routing. This strengthens the Medium-after-145
  case without deciding it.
- **Fire-once lesson doctrine** (>15 lessons trigger) and **retry accommodation** —
  the Learning-Game-Design branch's items; no docs-layer position beyond the push/pull
  safety argument above.

## Corpus gaps carried forward

Unchanged from [[Game-Player-Docs]]'s MOC: how observe-mostly games document the
observe/intervene split remains ungrounded (the corpus is command-driven), and — new
from this pass — the corpus has no grounding on *runtime-rendered reference validity*
(the report-card class of document). Both fit one future companion research pass.

## Basis

- [[_grounding]] — CDDA's generated-from-data reference; Cogmind's searchable manual
  and missed-message caution; NetHack whatis; DF quickstart scoping
- [[Manual-Structure-Conventions]] — the two reference-honesty mechanisms this note
  generalizes
- [[Explaining-The-Screen]] — whatis delegation; the look-cursor's docs role
- [[In-Game-Contextual-Help]] — push/pull invariant grounding the instrumentation park
- [[Quickstart-Guide-Patterns]] — minimal-viable-session scoping for decision 7
- [[Analysis-UI-Reference-And-Help-Stack]], [[Analysis-In-Game-First-Teaching]] —
  the immutable prior layers whose superseded clauses are declared above
- `docs/design/reorient-2026-07-26-ui.md` `## Decisions` 1–10 — fixed constraints
- Sibling drafts (Game-UI-UX, Game-Gameplay-Patterns, Learning-Game-Design branches),
  reconciled in prose per vault isolation
- Code/design evidence from the read-only pass: `overlays/postmortem.md` (7 stale
  cells vs `internal/tui/render_test.go:923`, `takeover_test.go`, `tui.go:939`),
  `panels/exercise.md:110`, `overlays/help.md` byte-identity table, `INDEX.md` gate
  rules + FR-020, `patterns/layout.md` rulings, `internal/tui/tiles.go`,
  `docs/player/` (13 pages; keys-reference staleness), `check-freshness.mjs` output,
  TASK-67/114/116/117/119/123/127/130/142/143/144/145/146 board views
