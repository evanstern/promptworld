---
title: Analysis — Pedagogy to UI
aliases: [Teaching-In-Play UI Projection, Learning-Game UI Analysis]
tags: [analysis, learning-games, pedagogy, tui, teaching-game]
type: analysis
created: 2026-07-25
updated: 2026-07-25
related: ["[[Learning-Game-Design]]", "[[Learning-Helper-Anatomy]]", "[[Teaching-Through-Play]]", "[[Healthy-Engagement-vs-Dark-Patterns]]", "[[Meta-Progression-and-Failure]]", "[[Puzzle-Pedagogy-Patterns]]", "[[Observe-Intervene-Onboarding]]", "[[Brief-and-Assumptions]]"]
---

# Analysis — Pedagogy to UI

_How this branch's evidence on teaching-through-play projects onto promptworld's
screens, under the 2026-07-25 reorientation lens: promptworld is pivoting into a
staged prompting-skills teaching game (curriculum ladder TASK-68, skinnable guardian
TASK-121) while remaining an ambient, terminal-first LLM-agent world sim, and its
UI must make the village legible, make prompting the guardian the central verb,
teach itself in play, and be fully specifiable page by page. Written after the
operator's steering round; the eight numbered Decisions and thirteen lead-adopted
defaults in `docs/design/reorient-2026-07-25-ui.md` are treated throughout as FIXED
constraints, not options._

## Thesis

**The substrate is unusually ready to teach itself; the curriculum has no living
screen.** promptworld already implements this corpus's hardest prescriptions —
progressive disclosure that is *structural* (a stage-locked tool is absent from the
angel's declared schema, not prose-forbidden, matching the idle-game grammar where
the UI itself is the tutorial, [[Observe-Intervene-Onboarding]]), a `?` overlay with
a deliberate empty lessons seam (`helpLessons`, spec 045, `docs/wiki/tui-client.md`),
plain-language glossary discipline, and an unlock chain that is automatic and
artifact-gated with no meta-currency, exactly the Slay-the-Spire-shaped
designer-paced tutorial ([[Meta-Progression-and-Failure]]). What was missing — and
what the steering round has now decided into existence — is the *rendering* of the
pedagogy: before the decisions, the moment of earning a stage was one chronicle line
that scrolls away at 16x, exercises had no on-screen goal, and the ladder's front
door was a CLI command outside the game (`promptworld stages`), reproducing the
WorldBox tutorial-bear discoverability failure ([[Observe-Intervene-Onboarding]]:
"I would have never thought to look there"). The decisions close every one of these
gaps; this analysis records the corpus case for each surface, the reconciliation
with the three sibling branches, and the residual risks the evidence flags.

## The operator's decisions as given constraints

From `docs/design/reorient-2026-07-25-ui.md` (verbatim intent, fixed): the headline
direction is the **Staged Cockpit** (Decision 1 — guardian console page + village
lens + stage-shaped layout *defaults*); first-occurrence lessons render in a
**dedicated lesson row** above the minibuffer, one active lesson, ≤2 lines, dwell
until done/dismissed, pointer at a key/tab, anti-spam with opportunity decay,
default-on at stages 1–2 and badge+overlay-only at stage 3+/pre-ladder (Decision 5);
**both big moments seize the screen immediately** — the stage-unlock ceremony on
`curriculum.stage_unlocked` and the postmortem on `run.ended` — dismissable and
replayable (Decision 6, overriding the badge-until-pause recommendation); a
**guardian strip** pairs the action budget with the minibuffer (Decision 7); the
report card renders as a card in the guardian console at natural stopping points and
inside the postmortem takeover (D5); unlock attribution goes to **the player's
authorship** — "your charter proved The Written Word", skin-resolved (D6);
seen-lessons state lives **per-user** (D8); the `?` overlay gains a **guardian
section** (D9); skin tokens ship **before** new fiction literals (D2); scenario
worlds get an **exercise panel** with attach-time briefing and staged incident-forecast
visibility (D11, D4); `q`-detach is the blessed stopping point (D13).

## What the corpus says about each decided surface

**The lesson row (Decision 5) is RimWorld's learning helper, correctly transplanted.**
The corpus's lesson unit bundles *content + trigger + UI pointer + decay policy* in
one def, scheduled with anti-spam gates — at most a few active, spaced, desire
threshold — and retired automatically when the player performs the interaction
([[Learning-Helper-Anatomy]]). The decided row carries all four fields plus the
one-active constraint. The stage-defaults refinement (on at 1–2, badge-only at 3+)
is the corpus's adaptive-messaging rule made cheap: George Fan's rule 7 — tips for
players doing the wrong thing, skipped for competent players — approximated by
stage as the competence proxy ([[Teaching-Through-Play]]). Lesson prose must keep
Fan's register: minimal words, unobtrusive, never a wall (rules 5/6/8 — "the little
boy who cried wolf"); the ≤2-line budget encodes this. The dwell-until-retired
behavior is the load-bearing difference from a feed line: a pushed lesson that can
scroll away fails the Cogmind caution this branch's own board projection recorded
("hot pink and blinking… and still people sometimes miss them" — TASK-117's
description), which is also why every pushed lesson keeps a pull path in the `?`
overlay (the RimWorld one-corpus/two-deliveries shape, [[Learning-Helper-Anatomy]]).

**The unlock ceremony (Decision 6) is the meta-progression reframe given a screen.**
Meta-progression's teaching function is the *felt* conversion of effort into
progress — "death delivers two signals: you made an error, and you still made
progress" ([[Meta-Progression-and-Failure]]) — and Hades' design goal was explicitly
to reduce the sting of failure by making every ending advance something. A stage
unlock rendered as one digest line cannot carry that beat. The takeover ceremony
can, and D6's attribution voice matters as much as the surface: SDT's competence
satisfaction — not the system's, the player's — is what predicts enjoyment and
return ([[Healthy-Engagement-vs-Dark-Patterns]]); "your charter proved The Written
Word" attributes the competence to the authored text, which is simultaneously the
skill being taught. The ceremony should also carry the natural-stopping-point copy:
the corpus finds players who experience a session as *completed* at a major moment
are less frustrated than those cut mid-quest, and first sessions should end "in the
best possible state to return" ([[Healthy-Engagement-vs-Dark-Patterns]] §session).
An unlock is precisely such a down-period; "a good place to rest — your next world
can begin at The Written Word" costs one line.

**The overridden interrupt caution, recorded factually.** This analysis originally
recommended badge-until-pause; the operator chose immediate seizure for both
takeovers. The corpus evidence on the residual risk is narrow and worth stating
precisely: Fan's rule 6 (messaging that "never pauses or interrupts play") and the
stopping-point research (frustration concentrates on being cut *mid-task*) both
warn against interruption during active play. However, both decided triggers are
themselves major-moment down-periods — `run.ended` interrupts nothing by
definition, and `curriculum.stage_unlocked` fires at an exercise pass, a resolution
beat. The genuine residual exposure is the case where the player is attending to
something *else* on screen when the unlock lands (e.g. mid-inspect on a paused
chronicle, or watching an unrelated crisis in a still-running scenario world): the
seizure discards their visual context. The decided mitigations — dismissable in one
key, replayable from `?`/`stages`/morgue — cap the cost at one context switch.
Verdict: the override spends a small, bounded amount of Fan-rule-6 compliance to
buy maximum salience for the single most missable, most pedagogically valuable
moment in the game; on this corpus's evidence that trade is defensible, and the
risk should be revisited only if playtesting (the Portal method — validate
sequencing empirically, weekly, [[Teaching-Through-Play]]) shows ceremony fatigue.

**The postmortem takeover + report card (D5) puts feedback at the down-period.**
Zachtronics' histograms appear at solve time and motivate without rewards
([[Puzzle-Pedagogy-Patterns]] §3); the decided placement — report card at stopping
points and inside the postmortem, badges between — is that timing discipline
applied to a sim. The card's charter-attributed critique ("your charter never
mentions coordinates…", TASK-115) stays on the intrinsic pole: comparative,
evidence-derived, reward-free, and it attributes to the text, never the person.

**The guardian section in `?` (D9) extends the deterministic floor to prompting.**
The overlay's model-free, byte-identical-with-nil-status property is the
charter-independent teaching floor; before D9 it taught keys and glyphs but zero
prompting — the game's actual subject. Static-per-stage granted-verbs content
(stage identity, concept, one example ask per verb) is renderable from data the
same way `stagesLadder` already is.

**Session shape (D13) is already healthy.** No streaks, no appointments, no FOMO
anywhere on the board — the dark-pattern taxonomy's temporal categories
([[Healthy-Engagement-vs-Dark-Patterns]]) have no instances to remove. Blessing
`q`-detach with "the world keeps running" copy converts the ambient design's
natural session-agnosticism into an explicit reassurance, the corpus's
"make quitting at one's own volition easy" recommendation.

## Reconciliation with the sibling branches

_(Sibling positions are restated in prose per branch isolation; their drafts are the
Game-UI-UX, Game-Gameplay-Patterns, and Game-Player-Docs evaluator reports of this
run, converged into `docs/design/reorient-2026-07-25-ui.md`.)_

**Adopted — the UI branch's Staged Cockpit as host.** The UI sibling proposed
promoting the guardian conversation to a first-class console page, splitting engine
telemetry to a systems tab, and letting stage shape layout *defaults* (never locks).
This analysis's lesson row and ceremony are not free-standing surfaces; they land
*inside* that shell: the lesson row in the bottom chrome the cockpit defines, the
report card as a console card (D5), the ceremony as a body-replacement takeover
using the same slot machinery the help overlay proved. The cockpit's stage-shaped
defaults are also the mechanism that makes Decision 5's stage-1–2/3+ split
implementable as one rule rather than a special case. Adopted without reservation —
it is this corpus's progressive-disclosure principle applied to the layout itself.

**Merged — the gameplay branch's exercise panel absorbs this branch's goal panel.**
The gameplay sibling specified a fourth dock tab for scenario worlds (framing,
rubric gauges, pass/fail, incident forecast, attach-time briefing — now D11); this
branch had independently proposed an exercise goal panel from the GameFlow
clear-goals/feedback criteria ([[Teaching-Through-Play]] §flow). Same surface; the
merge assigns the pedagogy contract to their panel: rubric terms render as *live*
event-derived gauges (feedback during, not only after), the framing line is the
Portal safe-room opening delivered at attach, and D4's staged forecast visibility
(visible at stages 1–2, fog from 3) is flow-zone widening — reduce uncertainty for
novices, restore it as skill grows. Their scenario-cadence narration trigger also
rescues this branch's "score narrative" concern: without it the chronicle's
ambient cadence can produce zero narrated entries in a minutes-scale run.

**Merged — the docs branch's taxonomy with this branch's ConceptDef anatomy.** The
docs sibling split TASK-117's lesson set into mechanics-tier and prompting-tier
lessons (first rejected tool call, first custom charter observed — the
prompting-verb first-occurrences the lens cares most about) and ratified
skin-token-first sequencing (D2). This branch supplies the per-lesson *behavioral*
contract from [[Learning-Helper-Anatomy]]: trigger condition, opportunity decay,
UI pointer, one-active anti-spam scheduling, auto-retire on performance. The two
compose cleanly — taxonomy says *what* lessons exist, anatomy says *how each
behaves* — and both land in TASK-117's spec. Skin-token-first is accepted as
sequencing discipline: lesson strings, ceremony copy, and the D6 attribution line
are all fiction-layer strings and must be tokens from day one, or TASK-121's sweep
grows unboundedly. D8 (per-user seen-lessons) additionally makes the corpus's
knowledge-decay idea *implementable* — resurfacing long-unused lessons requires
state that follows the player, not the world — though decay itself remains
unadopted (below).

**No unreconcilable conflicts.** The one prior disagreement — interrupt policy —
was resolved by operator override and is recorded above as a constraint with its
factual risk note.

## Recommendations as they now stand (ranked, post-decisions)

1. **TASK-117 spec carries the full lesson anatomy.** Lesson row per Decision 5;
   per-lesson fields = trigger, ≤2-line content (skin tokens, D2), UI pointer,
   opportunity-decay flag, priority; scheduler = one active, spaced, auto-retire on
   the taught interaction; taxonomy = mechanics tier + prompting tier (docs-branch
   merge); seen-state per-user (D8); every lesson names its `?` pull path.
   Evidence: [[Learning-Helper-Anatomy]], [[Teaching-Through-Play]] (Fan 5/6/7/8).
2. **The exercise panel (D11) is the clear-goals surface.** Live rubric gauges,
   attach-time briefing, D4 forecast staging, scenario-cadence narration.
   Evidence: GameFlow criteria ([[Teaching-Through-Play]]); [[Puzzle-Pedagogy-Patterns]]
   on feedback at solve time.
3. **The unlock ceremony ships with stopping-point copy and D6 attribution.**
   Immediate takeover per Decision 6; player-authorship voice; replayable;
   one-line "good place to rest / what your next world gains" close. Evidence:
   [[Meta-Progression-and-Failure]], [[Healthy-Engagement-vs-Dark-Patterns]] §SDT+§session.
4. **The postmortem takeover hosts the report card.** Rubric outcome,
   charter-attributed critique, epitaphs; badges between stopping points (D5).
   Evidence: [[Puzzle-Pedagogy-Patterns]] §3, [[Meta-Progression-and-Failure]] §Hades.
5. **The `?` guardian section (D9)** — static-per-stage, model-free prompting
   reference; the deterministic floor finally teaches the game's subject.
6. **Protect the opening composition.** The dormant minibuffer line plus
   TutorCharter is already the corpus-correct first minute (single salient action,
   safe room — [[Observe-Intervene-Onboarding]], [[Teaching-Through-Play]] §Portal);
   the cockpit's stage-1 default layout should keep the screen that calm.
7. **Adopt the Portal validation cadence for the teaching surfaces**: the lesson
   row, ceremony, and exercise panel are exactly the features whose sequencing the
   corpus says cannot be authored correctly without playtest iteration.

## Genuinely open questions (kept open)

1. **Bottom-chrome vertical budget.** Decision 5 (lesson row) + Decision 7
   (guardian strip) + minibuffer + footer stack up to four fixed rows at stages
   1–2. Stacking order, and whether the lesson row collapses to zero height when
   no lesson is active, needs a layout ruling in the tui-v2 design doc — the
   exact-height rendering invariant makes this a real budget, not a nicety.
2. **Score voice inside the ceremony.** Instrument voice (rubric checklist) vs
   fiction voice (narrated chapter) at the unlock moment — the gameplay sibling's
   question, now sharpened because Decision 6 makes the ceremony the moment the
   player remembers. Which voice leads the takeover is undecided.
3. **Knowledge decay.** D8's per-user store enables RimWorld-style resurfacing of
   long-unused lessons ([[Learning-Helper-Anatomy]]'s KnowledgeDecayRate); v1 is
   fire-once. Adopt decay post-v1, or is fire-once-forever the doctrine?
4. **Ceremony fatigue check.** Decision 6 stands; the open item is only the
   playtest trigger — what observed behavior (dismiss-within-N-frames rates,
   operator feedback) would reopen the interrupt policy?
5. **Audience tier (carried from the synthesis).** Whether the overlay's advanced
   tier and the exercise panel's gauges may expose raw registry numbers depends on
   the self-directed-engineer vs classroom answer, still open at the synthesis
   level (`docs/design/learning-game-synthesis.md` open question 3).
