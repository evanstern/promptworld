---
title: Analysis — In-Game-First Teaching
aliases: [In-Game vs Out-of-Game Docs, Teaching Surface Analysis]
tags: [analysis, player-docs, teaching-game]
type: analysis
created: 2026-07-25
updated: 2026-07-25
related: ["[[Game-Player-Docs]]", "[[Quickstart-Guide-Patterns]]", "[[Explaining-The-Screen]]", "[[In-Game-Contextual-Help]]", "[[Manual-Structure-Conventions]]", "[[Brief-and-Assumptions]]"]
---

# Analysis — In-Game-First Teaching

_Should promptworld's help be primarily in-game or out-of-game — the question
[[Brief-and-Assumptions]] deferred — now answered under the updated lens the operator has
since stated: promptworld is a **learning game** (prompting → agents → tools), so
"player documentation" and "the game teaching you" are one concern, and contextual help,
tutorials, and curriculum are gameplay._

## Verdict

**In-game primary, out-of-game reference — operator-ratified 2026-07-25.** The angel
leads: Metatron becomes the tutor (in-fiction help), grounded by a read-only,
registry-derived `explain` tool so mechanics facts come from data, never from the model's
vibes. `docs/player/` remains the reference layer (the DF-wiki role), gaining a
screen-orientation page and a keys reference card immediately. Onboarding is a property
of **every world** at the TUI level — a `?` overlay and auto-retiring first-occurrence
lessons — while curriculum stages stay a separate opt-in layer. Failure stays unsoftened
in all modes and is reassured up front, DF-style: your first village will probably freeze,
and that's the story.

## Reasoning

### Why in-game wins

1. **The corpus's own trajectory.** Every retrofit case moved help *into* the game — DF
   Steam added its tutorial + help section, Caves of Qud 1.0 shipped its tutorial together
   with a UI overhaul — and RimWorld never had a manual at all ([[In-Game-Contextual-Help]],
   [[_grounding]] §§ RimWorld, Dwarf Fortress, Caves of Qud).
2. **The ambient-sim attention problem.** The branch MOC's open question — the corpus is
   command-driven; how do observe-mostly games document the observe/intervene split? —
   cuts in-game's favor: an ambient player has no natural moment to leave for a manual;
   help must arrive where attention already is.
3. **The learning-game clincher, unique to promptworld.** The intervention surface is
   conversation with an LLM agent. "Ask the angel what that means" is simultaneously the
   whatis command (NetHack's `/`, [[Explaining-The-Screen]]) and a rep of the exact skill
   the game teaches. No comparator could make its help channel the curriculum;
   promptworld can. This is the strongest single argument and the operator has ratified
   building on it.

### What already fits (evaluation against the corpus)

- **Cogmind layer 1 — "make help unnecessary first"** ([[_grounding]] § Cogmind) is
  promptworld's shipped strength: the per-event digest grammar with a sweep test
  (TASK-60), the plain-language verdict glossary where raw enums never reach the screen
  (TASK-63), suppression rows that carry their own remedy verbatim ("suppressed at 32x —
  calibrate or slow down", TASK-40/41), and the governed-speed suffix. This is the
  researched principle, implemented.
- **Cogmind layer 2 — context help at the point of question** is half-built: the
  chronicle detail pane with a documented jump-off extension point, the map legend's
  inspection line, the villager decisions sub-view (TASK-63 — "why did my agent do
  that", the prompting learner's core feedback surface). These answer "what happened"
  (instance) but not yet "what IS this" (concept).
- **DF quickstart structure** ([[Quickstart-Guide-Patterns]]): `getting-started.html`
  follows it — numbered do-this-then-this, plain-terms definitions at first mention,
  why-alongside-how, a trailing quick-reference block, explicit hand-off to deeper pages.
  `playing-via-metatron.html` answers the brief's founding question ("what is the damn
  angel thing") in its first two paragraphs.
- **Keeping reference honest** ([[Manual-Structure-Conventions]]): both researched
  mechanisms already exist — pages pinned to wiki notes + commit with a checkable
  freshness gate (TASK-82, the CDDA Hitchhiker's-Guide posture), and in-game
  described ≡ declared by construction (Metatron's tool guidance derived from the
  registry, TASK-64).

### What's missing (the gaps the verdict addresses)

- **Cogmind layers 3–4 are absent entirely**: no tutorial, no in-game searchable
  reference, no `?` key anywhere in the TUI (verified against `internal/tui/`).
- **NetHack chapter 3 has no counterpart**: the map's dense glyph vocabulary is
  enumerated in no player page — `docs/player/` contains no legend or symbol table.
- **No losing-is-fun reassurance** despite real, permanent death and documented
  village-scale collapses; DF leads its quickstart with exactly this
  ([[Quickstart-Guide-Patterns]]).
- **The unreliable-manual hazard**: Metatron's fixed frame forbids inventing *events*
  but nothing grounds it in *mechanics facts* — a charter-voiced angel confidently
  misexplaining the horizon would be anti-teaching. Hence the `explain` tool as the
  deterministic backstop (the NetHack pattern of delegating to a lookup the docs merely
  teach, adapted for an LLM intermediary).

### Reconciliation with the sibling gameplay-patterns branch

The sibling analysis's verdict — promptworld has the instrument but not the game:
"nothing manufactures situations that demand prompting skill, and nothing tells the
player whether their prompt made a difference" — is complementary, not competing. This
branch owns the second half of that sentence. Three reconciliations:

1. **One grounded feedback layer, two directions.** The sibling's "angel's report card"
   (post-turn cheap-chain critique attributing outcomes to charter text — "your charter
   never mentions coordinates") and this branch's `explain` tool converge: **explain is
   pull** (deterministic facts from the registries — tool rosters, charge/miracle costs,
   decision classes, glyphs), **the report card is push** (attribution). They should
   share grounding: the report card must cite event-log evidence and registry facts
   through the same data the explain tool serves, so the grader never grades on vibes.
   RimWorld's learning helper — the same lesson corpus serving both push (triggered) and
   pull (searched) ([[In-Game-Contextual-Help]]) — is the precedent for one corpus, two
   deliveries.
2. **"Writing is the primary verb; watching is the reward" vs feed-anchored help.** The
   operator's staged-session decision dissolves this: in scenario runs (the early
   curriculum), teaching attaches to the *write* surface — the report card after a turn,
   the explain tool in conversation, the tutor charter's orientation; in the ambient
   endgame, teaching attaches to the *watch* surface — first-occurrence lessons in the
   feed, the `?` overlay over the screen the player is reading. Same lesson corpus, two
   trigger regimes. Director-lite (pre-scripted incident schedules in scenario worlds)
   also resolves the sibling's "drama on no schedule means lessons on no schedule": in
   scenarios, the scheduled incident IS the lesson trigger — the RimWorld trigger model
   with a deterministic event source; in ambient worlds, first-occurrence triggers
   suffice because the lesson corpus auto-retires.
3. **One coherent initiative-frame position** (the sibling's TASK-111/112 concern —
   autonomy must not erase the player or let the angel grade itself — vs this branch's
   round-1 "the autonomous angel is the natural adaptive tutor"): **split the roles by
   channel, not by entity.** The angel's *world-acting* autonomy is the graded artifact —
   the player's program — and the sibling's AC stands: charter quality must measurably
   change autonomous performance. The angel's *tutor voice* is a help channel that must
   never be part of the graded run: it spends no charges, lands no world events, earns no
   faith, and is excluded from any rubric. The existing architecture already draws this
   line structurally — converse is the final-text channel and deliberately not a callable
   tool, while acting tools land through gated doors — so the initiative frame ("acts
   in-world only when player-asked or pre-authorized") needs **no relaxation** for
   tutoring: explaining is speech, not an act. Tutor-mode is a charter/skill content
   change plus one read-only tool grant, not a doctrine change.

### Operator decisions recorded as constraints

1. **Angel as tutor: yes** — tutor + registry-derived read-only `explain` tool;
   in-fiction help leads.
2. **Screen-orientation page + keys card: approved** as an immediate content task (to be
   carded at synthesis).
3. **Failure tone: unsoftened everywhere + reassure up front** — deaths stay real in all
   modes (the honest grade); `getting-started.html` gains the DF-style paragraph.
4. **Onboarding is every-world, TUI-level** — `?` overlay and first-occurrence lessons
   run everywhere and auto-retire; seen-lessons state lives client-side; curriculum
   stages remain opt-in.

Sibling-side decisions this analysis takes as given: staged session shape (scenarios
early, ambient endgame), director-lite first, hybrid scoring (event-derived rubrics for
scenarios/duels, chronicle-only for ambient), faith-driven endogenous charge regen.

## Recommendations (reconciled ranking)

1. **Screen-orientation page + keys reference card** (approved; content-only; the
   player-docs skill's generation/freshness machinery is the home). NetHack
   chapter-3-shaped: the player's question as the heading, organized by screen region,
   glyph table prefaced "you need not memorize these" — because items 2–3 will answer in
   place. Include the losing-is-fun paragraph in `getting-started.html` in the same pass.
2. **The grounded feedback layer**: the `explain` tool (deterministic, registry-derived,
   grant-gated, read-only) + a default `skills/guide.md` and tutor-charter preset that
   teach the angel to answer how-do-I-play questions through it + the sibling's report
   card sharing the same grounding. This is the operator-ratified lead and the substrate
   for curriculum stage 1.
3. **`?` overlay in the TUI** (every world): context-sensitive per pane — keys first
   (Cogmind's basic/advanced tiering), then the screen-region walkthrough. Load-bearing
   detail: a **no-LLM world has no tutor** — reflex-only villages are a first-class
   posture — so the overlay is not redundant with item 2; it is the out-of-fiction floor
   beneath an angel that may be absent, down, or mid-repair.
4. **First-occurrence lessons projection** (every world, auto-retiring, seen-state
   client-side): the RimWorld trigger model over the event stream, using the same
   client-side-projection pattern as the decision-trace view. In scenario worlds,
   director-lite incidents become scheduled triggers. Heed Cogmind's recorded caution —
   "hot pink and blinking… and still people sometimes miss them" — every pushed lesson
   must also be reachable from the `?` overlay's pull reference.
5. **Per-stage quickstarts** riding the curriculum ladder (DF's separate-quickstarts-
   per-audience pattern, [[Quickstart-Guide-Patterns]]): each stage ships a one-pager;
   pass signals surface in-game with the chronicle as the score narrative.
6. **Postmortem docs**: fork-compare/duel rubric output rendered through the same
   plain-language glossary discipline as the TUI (no raw enums in a grade), and the
   morgue/retelling exports framed as the losing-is-fun artifact.

Sequencing is deferred to the synthesis phase; **recommendation only**: 1 (approved
content, zero design risk) → 2 (the ratified lead; unblocks stage 1) → 3 → 4.

## Proposed backlog moves

- **New task (approved)**: screen-orientation page + keys card + losing-is-fun paragraph
  (player-docs skill extension; consider a registry-generated reference section so
  glyph/cost/key tables cannot rot — the CDDA pattern).
- **New task**: `explain` tool + default guide skill + tutor-charter preset; note the
  shared-grounding contract with the sibling's report card.
- **New task**: `?` overlay (every world, no-LLM worlds included).
- **New task**: first-occurrence lessons projection (every world; scenario-incident
  triggers when director-lite lands).
- **TASK-68 → High** (concurring with the sibling: it is the spine). Add ACs: a
  per-stage quickstart page; the stage pass signal surfaces in-game; stage-1 orientation
  delivered via the tutor charter.
- **TASK-111/112**: adopt the sibling's AC (charter quality measurably changes
  autonomous performance) and add the channel-split corollary: the tutor voice is
  excluded from the graded surface; document world-acting vs converse as the
  initiative-frame boundary. Frame 112 as "the player programs an agent" (stage 3).
- **TASK-67**: compare/duel output rendered plain-language (glossary discipline);
  framed as the postmortem teaching artifact.
- **TASK-14 (proving run)**: add the sibling's question, whose second clause is this
  branch's test — "did prompt iteration measurably change outcomes, **and could the
  player tell why from in-game surfaces alone**?"

## Tensions & tradeoffs

- **Tutor vs deliberate incompetence** — the one conflict this analysis could NOT fully
  reconcile. The sibling side leaves open an "angel's deliberate-incompetence ceiling"
  (an angel that is bad on purpose so the player learns to fix it). The same entity
  being the trusted tutor is contradictory unless the channel split above is enforced
  architecturally: incompetence may apply to *world-acting* only, never to the tutor
  voice or the explain tool's facts. This analysis takes that position, but the operator
  has not ruled on the incompetence ceiling itself — flagged, not papered over.
- **In-fiction help can still mislead in tone** even with grounded facts: a surly custom
  charter voices the tutor too. Acceptable — the `?` overlay and `docs/player/` are the
  charter-independent floors — but worth stating: the game's *guaranteed* help is
  deterministic; the angel's help is only as good as the player's prompt, which is
  itself the lesson.
- **Cost**: tutor turns and report-card critiques are model calls. The report card is
  cheap-chain by design; explain-tool reads are free (no model needed to serve them);
  first-occurrence lessons are model-free strings. The expensive path (tutor
  conversation) is player-initiated, which is the correct incentive shape.

## Confidence & open questions

Confidence: high on the verdict (operator-ratified) and on recommendations 1–3 (they
map one-to-one onto grounded patterns and existing machinery); medium on 4's trigger
taxonomy (which event types are lesson-worthy is content design, not yet grounded).

Open (genuinely undecided, drives the next steering round):

- The **player-attributable failure state** (open on the sibling side) bounds what the
  report card can honestly attribute — until it exists, the card should attribute to
  charter text only, never to blame ("you lost Ash").
- The **deliberate-incompetence ceiling** (see Tensions) — needs an operator ruling
  that the channel split is doctrine.
- **Audience** (self-directed engineers vs classroom) decides whether the `?` overlay's
  advanced tier exposes raw registry values or stays plain-language-only.
- **Seen-lessons persistence specifics**: "client-side/TUI-level" is decided; the exact
  home (a per-user file vs per-world client state) and its reset semantics are not.
- **Corpus gaps carried forward** (from the branch MOC, still ungrounded): the RimWorld
  learning helper's per-lesson anatomy, and how observe-mostly games document the
  observe/intervene split; the sibling flagged zero learning-game-design sources
  vault-wide. One companion research pass should cover all three.

## Basis

- [[_grounding]] — the cited research pass (RimWorld learning helper; DF quickstart
  structure and losing-is-fun; NetHack Guidebook chapters 2–3 and whatis; Cogmind's four
  layers and "make help unnecessary first"; CDDA's data-driven reference; manual-writing
  separation rules)
- [[Quickstart-Guide-Patterns]] — first-session structure and tone devices
- [[Explaining-The-Screen]] — screen-region orientation and the whatis delegation pattern
- [[In-Game-Contextual-Help]] — the RimWorld trigger model and Cogmind's layer ordering
- [[Manual-Structure-Conventions]] — reference honesty mechanisms and reference cards
- [[Brief-and-Assumptions]] — the original brief whose deferred question this note resolves
- The sibling gameplay-patterns branch's analysis (referenced in prose only; vault
  isolation) — the "instrument but not the game" verdict, the angel's report card, the
  writing-is-the-verb inversion, and the TASK-111/112 gradeability AC reconciled above
- Codebase/board evidence cited during the read-only pass: `docs/player/` (7 pages,
  TASK-82/spec 026), `internal/tui/` (no `?` binding; digest grammar, verdict glossary,
  detail pane, decisions sub-view), TASK-40/41/60/63/64/66/67/68/78,
  `docs/design/horizon-vs-learner-iteration-speed.md` (decision-6),
  `docs/design/control-surface-and-calibration.md` (world-01 collapse evidence)
