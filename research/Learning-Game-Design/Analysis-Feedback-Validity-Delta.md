---
title: Analysis — Feedback Validity Delta
aliases: [Post-Cockpit Pedagogy Delta, Report-Card Truth Analysis]
tags: [analysis, learning-games, pedagogy, feedback-validity, tui, teaching-game]
type: analysis
created: 2026-07-26
updated: 2026-07-26
related: ["[[Learning-Game-Design]]", "[[Analysis-Pedagogy-To-UI]]", "[[Learning-Helper-Anatomy]]", "[[Teaching-Through-Play]]", "[[Healthy-Engagement-vs-Dark-Patterns]]", "[[Meta-Progression-and-Failure]]", "[[Puzzle-Pedagogy-Patterns]]", "[[Observe-Intervene-Onboarding]]", "[[Brief-and-Assumptions]]"]
---

# Analysis — Feedback Validity Delta

_The 2026-07-26 re-run's delta evaluation, written at Phase 4 (cross-grounding)
after the operator ratified the ten decisions in
`docs/design/reorient-2026-07-26-ui.md` — treated throughout as FIXED
constraints, not options. This layer sits on top of [[Analysis-Pedagogy-To-UI]]
(immutable; its superseded assumptions are declared below, never edited) and
reconciles with the three sibling evaluator drafts of this run — Game-UI-UX,
Game-Gameplay-Patterns, and Game-Player-Docs — restated in prose per branch
isolation._

## Thesis

**The teaching chrome shipped and matches this corpus; the problem class has
moved from surface existence to feedback validity.** Every ranked
recommendation in [[Analysis-Pedagogy-To-UI]] except one (the Portal
validation cadence, rec 7) now has a shipped, spec'd, design-pinned surface:
the lesson row is a faithful RimWorld ConceptDef transplant
([[Learning-Helper-Anatomy]] → `docs/design/tui/panels/lesson-row.md`, spec
055/TASK-117 — trigger + ≤2-line content + pointer + decay + one-active
spacing + per-user seen state at `~/.promptworld/lessons-seen.json`); the
takeover pair carries the meta-progression reframe and the no-blame failure
register ([[Meta-Progression-and-Failure]] → `overlays/ceremony.md` /
`overlays/postmortem.md`, spec 056/TASK-127, including the D13 blessed
stopping point — and the postmortem *deliberately drops* the "world keeps
running" reassurance because it would be a lie on an ended run, a finer
application of the stopping-point research
([[Healthy-Engagement-vs-Dark-Patterns]] §session) than this branch asked
for); the exercise panel is the GameFlow clear-goals surface
([[Teaching-Through-Play]] §flow → `panels/exercise.md`, spec 054/TASK-119);
and spec 063's tutor stack — deterministic explain tool, compiled-in tutor
guide, evidence-cited report card, `?` guardian section — is a stronger match
to this corpus's feedback findings than the branch dared recommend (below).

What the delta exposes is a subtler failure the corpus is equally clear
about: **a teaching surface that renders an untruth teaches the untruth.**
The report-card renderer that fires at the postmortem, the ceremony, and the
console card grades by generic event *presence*
(`reportCardFactsFromEvents/Evidence`, `internal/tui/views.go:863–880`),
while the live exercise tab grades by real semantics (`sim.EvaluateRubric`,
`internal/sim/scenario.go:280`, where zero deaths means MET at count 0). The
result is documented in the shipped design corpus's own mockup
(`overlays/postmortem.md`): a first-night failure with two deaths renders
**"✓ agent died (agent.died: 2)"** — a checkmark on the run's failing
outcome, at the single most emotionally salient teaching moment the game has.
Histogram-style feedback works because it is true and evidence-derived
([[Puzzle-Pedagogy-Patterns]] §3); competence satisfaction — the player's,
grounded in what actually happened — is what predicts enjoyment and return
([[Healthy-Engagement-vs-Dark-Patterns]] §SDT); Hades' whole failure design
is honest accounting with the sting removed ([[Meta-Progression-and-Failure]]
§Hades). A ✓ on a failure violates all three at once. **Decision 1 takes this
finding in full**: a new HIGH card unifies every report-card surface on
`sim.EvaluateRubric`, authors `the-law`'s real evaluator (persisting the
charter `Default` flag into state — the documented blocker at
`internal/sim/scenario.go:277`), and sequences after TASK-144 (the flaky
report-card test touches the same code).

## The operator's decisions as given constraints

From `docs/design/reorient-2026-07-26-ui.md` §Decisions (ratified 2026-07-26;
verbatim intent). The ones this branch's evidence bears on:

- **Decision 1 — report-card truth**: NEW HIGH card; unify on
  `sim.EvaluateRubric` everywhere; ✗ renders at the postmortem; `the-law`
  evaluator authored; after TASK-144. (This analysis's #1 recommendation,
  taken.)
- **Decision 2 — design-gate semantic lint**: `check-tui-design.mjs` gains a
  shipped-page/`unbuilt (wave` cell check; the stale `overlays/postmortem.md`
  cells and `panels/exercise.md:110` amended in the same PR.
- **Decision 3 — TASK-67 fork duel promoted to HIGH**, reframed as the
  loop's iteration rung; v1 = rubric-first scoreboard sharing
  `reportCardView` + `sim.EvaluateRubric`.
- **Decision 5 — exercise catalog wave**: 2–3 hand-authored exercises per
  stage; incident vocabulary grows to ~3 kinds.
- **Decision 6 — in-TUI ladder view**: the forward ladder (identity ·
  concept · earned/next · unlock evidence, matching `stages --json`) renders
  in the `?` guardian section — deterministic floor, model-free.
- **Decision 7 — quickstart first-prompt step**: `getting-started.html`
  gains an "ask your guardian one thing" step sourced from the
  `skin.guardian.example_ask.*` token family; per-stage first-session blocks.
- **Decision 9 — lane order**: duel (TASK-67) before faith (TASK-118).
- **Decision 10 — parked watch items** (recorded, NOT scheduled; kept
  genuinely open below): teaching-signal instrumentation consent; fire-once
  lesson doctrine (revisit at >15 lessons); retry accommodation;
  rubric-gauge exposure doctrine; ambient-drama-floor posture.

## Reconciliation with the sibling branches

**Game-Gameplay-Patterns — same defect, reconciled severity and ownership;
their catalog wave adopted as this branch's delivery vehicle.** The gameplay
sibling found the identical report-card simplification and filed it as a
watch item ("should not survive the next content wave"); this branch ranks
it the #1 remaining gap. The severity difference is a lens difference, not a
factual dispute: under their loop-completeness lens the card is a cosmetic
lag behind a correct evaluator; under this branch's teach-itself lens the
postmortem *is* the failure lesson, and a lesson that shows ✓ on death
mis-teaches now, on every failed run, at the moment the corpus says feedback
lands hardest ([[Puzzle-Pedagogy-Patterns]] §3 — feedback at solve time;
[[Meta-Progression-and-Failure]] — death must deliver *true* signals).
Decision 1 resolves it at this branch's severity (HIGH, now), while adopting
their two sharpenings: `the-law`'s evaluator rides the same card (their
finding that its rubric renders permanently pending), and the fix sequences
after TASK-144. The ownership finding stands as a process lesson worth
recording: the gap had been parked on TASK-119, which closed Done without
owning it — the same status-exceeded-ownership shape TASK-135 was carded to
fix. Their **exercise catalog wave (decision 5) is adopted as the delivery
vehicle for this branch's remaining lesson work**: lesson tranche 2 (first
explain answer, first report card, first skill file at stage 3, first faith
event post-TASK-118) and the first wrong-thing detector (repeated
same-cause tool-call rejections — Fan rule 7's *adaptive* messaging, tips
for players doing the wrong thing, [[Teaching-Through-Play]]) should land as
content riding that wave rather than as a separate teaching-chrome feature —
new exercises and new lessons are the same authoring motion over the same
event vocabulary. Their metric-gaming concern about all-terms-live rubric
gauges and this branch's live-gauges endorsement are both parked correctly
under decision 10 (rubric-gauge exposure: all-terms-live stands until
multi-exercise content exists). Their duel-before-faith ordering is
decision 9 and consistent with this branch's TASK-67 elevation.

**Game-Player-Docs — their semantic lint guards the same property this
branch's finding depends on; adopted as one honesty doctrine with two
enforcement points.** The docs sibling caught `overlays/postmortem.md`
carrying `status: shipped` with seven `unbuilt (wave 4)` renderer cells —
the *documentation* telling an untruth about what exists, where this
branch's finding is the *surface* telling an untruth about what happened.
These are the same failure class: a gate that checks structure but not
meaning lets falsehood ride a green check. The corpus's version of the
lesson is Fan rule 8's noise warning inverted — a reference that cries wolf
(or worse, cries all-clear) trains the player, and the implementer, to tune
out ([[Teaching-Through-Play]]). Decisions 1 and 2 should be understood as
one doctrine — *gates must verify meaning where meaning is checkable* —
enforced at two layers: the design corpus (decision 2's lint) and the
rendered game (decision 1's one-evaluator rule). Their **quickstart
first-prompt step (decision 7) is adopted with this corpus's grounding
attached**: it is Fan rules 2 and 4 verbatim — doing over reading, and "get
players to perform actions once… Once they see the results of their action,
that's often all it takes" ([[Teaching-Through-Play]]) — applied to the
pivot's central verb, which the current minimal session (build → create →
start → watch → detach) never exercises. The sample ask should come from the
same `skin.guardian.example_ask.*` family the `?` guardian section renders,
so the doc, the overlay, and the tutor teach one phrasing vocabulary
([[Observe-Intervene-Onboarding]] — players learn the interaction grammar
itself; consistency is the tutorializer).

**Game-UI-UX — no collision; a scheduling difference the decisions resolve
explicitly.** Their #1 (the TASK-142 look-cursor; lens clause (a),
map-interrogability) and this branch's #1 (feedback validity; lens clause
(c)) name different biggest-gaps from different corpora — the operator
scheduled both (decisions 1 and 4) rather than forcing a rank, which is the
correct resolution and is recorded here as such rather than papered over:
under *this* branch's lens, validity outranks interrogability, and decision
1's HIGH vs decision 4's next-UI-lane sequencing reflects that. One genuine
composition adopted from them and the docs sibling jointly: the look-cursor's
TILE pane rendering the spec-068 tile registry's `meaning` rows makes the
map a third in-place whatis surface (after `?` and explain) — progressive
disclosure through the UI itself, this corpus's idle-game grammar
([[Observe-Intervene-Onboarding]]) extended to terrain. Their mouse-parity
sweep test (decision 8) is the same convention→mechanism move as decision 2;
no pedagogy stake beyond endorsing the pattern.

**One tension carried consciously, not a conflict.** The gameplay sibling
wants *more* ambient drama supply (seasons, the neglect detector — their
TASK-133/28 reframings; their evidence: 0 deaths in 5.2 game-days); this
branch proposed retry accommodation — *softening* repeated scenario failure
(Hades' God Mode shape, +2%/death, loop preserved,
[[Meta-Progression-and-Failure]] §Hades). These pull opposite directions on
failure frequency, but they are both flow-zone widening
([[Teaching-Through-Play]] §flow — anxiety above the zone, boredom below):
drama supply fixes the boredom edge of the ambient endgame; retry
accommodation would fix the anxiety edge of the scenario ladder. The
existing reconciliation instrument is the stage-keyed forecast/fog
vocabulary (D4, `patterns/stage-defaults.md`) — difficulty presentation
already varies by demonstrated competence. Retry accommodation is parked
under decision 10; if it is ever taken, the unlock record must carry an
honesty marker (the `stage_overridden` precedent) so accommodated passes
stay auditable — the skill-dilution debate's legitimate worry
([[Meta-Progression-and-Failure]] §skill-dilution) answered the way this
project answers everything: with evidence markers.

## What this corpus says about the newly decided surfaces

**The ladder view (decision 6) completes the discoverability repair.**
[[Analysis-Pedagogy-To-UI]]'s thesis line — "the curriculum has no living
screen" — is hereby declared **partially superseded**: the *current* stage
now lives everywhere in-game (header segment, console header, `?` guardian
section via `helpGuardianLines`, `internal/tui/help.go:451`), and unlocked
stages replay from `?` (`ceremonyReplayLines`). What remained was the
*forward* ladder — all four stages, earned state, and what evidence unlocks
the next — visible only in `promptworld stages`, outside the game: the
WorldBox tutorial-bear failure ([[Observe-Intervene-Onboarding]] — "I would
have never thought to look there") half-standing. The corpus case for
decision 6 is direct: Slay the Spire's ladder paces learning precisely
because the not-yet-earned pool is *visible* and unlocks are automatic and
designer-controlled ([[Meta-Progression-and-Failure]] §StS), and clear goals
are a GameFlow criterion ([[Teaching-Through-Play]]). The implementation is
nearly free by prior design: `world.StagesLadder` was relocated to
`internal/world` by spec 063 T014 explicitly so multiple surfaces could
render one table — decision 6 is the third renderer, and placing it in the
`?` guardian section keeps it on the deterministic, model-free floor
(stage-keyed content, the shipped byte-identity classification in
`overlays/help.md`). Content contract this branch recommends: render exactly
the `stages --json` fields (identity, concept, earned, unlock evidence,
proving world when earned) — an informed identity table, never a difficulty
menu, the spec-046 framing preserved.

**Spec 063's tutor stack, assessed against the corpus (recorded here because
the prior layer predates it).** The explain tool is TIS-100's
manual-as-game-artifact made mechanical: deterministic, registry-derived
("zero model-generated bytes," SC-001), in-fiction, pull-based
([[Puzzle-Pedagogy-Patterns]] §1) — fused with RimWorld's adaptive delivery
because the *guardian* chooses when to consult it
([[Learning-Helper-Anatomy]]). The report card's attribution note grades
only from recorded evidence, cites event seqs and charter fingerprints,
attributes to the text never the person, degrades silently without the
chain, and appears only at stopping points with a badge between (spec 063
FR-005/FR-006) — every clause on the intrinsic pole of
[[Healthy-Engagement-vs-Dark-Patterns]]'s dividing line, and the
Zachtronics timing discipline (feedback at the down-period, no rewards)
applied to prose critique. Tutor-lane doctrine — explain is charge-free,
faith-free, rubric-excluded, "explaining is speech, not an act" (FR-002/
FR-003) — is an SDT-autonomy protection this corpus implies but never states
so crisply: asking questions must never become a costed or graded behavior,
or the game punishes the exact curiosity it exists to teach. Decision 1
completes this stack rather than amending it: the checklist half of the
composed card becomes as honest as the note half already is.

## Recommendations as they now stand (ranked, post-decisions)

1. **Execute decision 1 with the term-semantics vocabulary made explicit.**
   The unification card should give `RubricTerm` an honest tri-state
   (met/missed/pending by call-site, as `overlays/postmortem.md` §Report card
   already specifies) and per-term direction (wanted-present vs
   wanted-absent), so no future exercise can reintroduce presence-grading.
   Evidence: [[Puzzle-Pedagogy-Patterns]] §3; [[Meta-Progression-and-Failure]].
2. **Ride decision 5's catalog wave with the teaching content:** lesson
   tranche 2, the first wrong-thing detector (Fan rule 7), and `the-law`'s
   evaluator (decision 1's scope where the wave lands first). One authoring
   motion, one event vocabulary. Evidence: [[Teaching-Through-Play]],
   [[Learning-Helper-Anatomy]].
3. **Decision 6's ladder view renders `stages --json` verbatim in the `?`
   guardian section** — model-free, identity-framed, evidence-forward.
   Evidence: [[Meta-Progression-and-Failure]] §StS,
   [[Observe-Intervene-Onboarding]].
4. **TASK-67 (decisions 3/9) is this corpus's comparison surface** — Opus
   Magnum's histogram-without-rewards for prompting: fork, diverge the
   charter, rubric-first scoreboard through the now-shared renderer. The
   last unserved corpus pillar; intrinsic motivation by comparison, never by
   reward. Evidence: [[Puzzle-Pedagogy-Patterns]] §3,
   [[Healthy-Engagement-vs-Dark-Patterns]] §overjustification.
5. **Decision 7's first-prompt step uses the example-ask token family** so
   docs, overlay, and tutor teach one phrasing vocabulary. Evidence:
   [[Teaching-Through-Play]] (Fan rules 2/4/10).
6. **When TASK-118 (faith) is eventually taken (after the duel, decision
   9):** keep faith an in-fiction resource, never a badge/streak surface
   (the overjustification caution and the streak-anxiety evidence,
   [[Healthy-Engagement-vs-Dark-Patterns]]); the failure-spiral floor
   question is Hades' proprietary-difficulty reasoning and should cite it.

## Superseded-assumptions ledger (prior layers immutable; declared here)

- [[Analysis-Pedagogy-To-UI]] **thesis** ("the curriculum has no living
  screen"): partially superseded — current-stage surfaces shipped (specs
  055/056/063/066); the forward-ladder half is decided into existence
  (decision 6) but unbuilt at this writing.
- Prior **open question 1** (bottom-chrome vertical budget): RESOLVED —
  `patterns/layout.md` rules `bodyMin` and the fold order (villager strip →
  lesson row → guardian strip last); the lesson row folds to a designed
  badge state.
- Prior **open question 2** (score voice in the ceremony): RESOLVED — FR-019
  ruled both voices, instrument authoritative (`overlays/ceremony.md`).
- Prior **open question 3** (knowledge decay vs fire-once): NOT resolved —
  parked under decision 10 with a concrete trigger (revisit at >15 lessons).
  One shipped nuance recorded: queue-decayed lessons are NOT marked seen
  (`panels/lesson-row.md` §anti-spam), a deliberate softening of fire-once.
- Prior **open question 4** (ceremony-fatigue reopening trigger): still
  named (`overlays/ceremony.md` §watch item) but **unobservable** — nothing
  instruments dismiss latency or dwell time; instrumentation consent is
  parked under decision 10. A named signal nobody can observe is not yet a
  gate.
- Prior **open question 5** (audience tier / raw registry exposure):
  RESOLVED — the FR-020 corpus-wide ruling (plain language by default; raw
  values behind a debug mode, `docs/design/tui/INDEX.md`).
- [[Brief-and-Assumptions]]'s framing of promptworld as observe-mostly with
  sparse intervention: stands, but the intervention verb now has the largest
  page in the app (guardian console, spec 053) — the watch/act split this
  corpus studied is now *taught* rather than merely designed.

## Genuinely open (decision-10 parked items, kept open)

1. **Fire-once lesson doctrine** — revisit at >15 catalog entries; the
   corpus's resurfacing mechanism ([[Learning-Helper-Anatomy]]'s
   KnowledgeDecayRate) remains the candidate design when the trigger fires.
2. **Teaching-signal instrumentation consent** — local-only event emission
   (lesson dismiss latency, ceremony dwell) is the cheapest way to make the
   decision-6-era watch items decidable by the Portal method
   ([[Teaching-Through-Play]] §playtesting); whether ANY behavioral
   telemetry is in-doctrine is the operator's call, unmade.
3. **Retry accommodation** — parked; if taken, honesty-marked passes (the
   `stage_overridden` precedent) are this branch's condition for it.
4. **Rubric-gauge exposure** — all-terms-live stands until multi-exercise
   content exists (the gameplay sibling's metric-gaming watch item, shared).
