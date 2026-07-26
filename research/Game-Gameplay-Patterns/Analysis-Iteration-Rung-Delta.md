---
title: Analysis — The iteration rung: play-loop delta after the Staged Cockpit shipped
aliases: [Iteration Rung Delta]
tags: [analysis, gameplay, learning-game, delta]
type: analysis
created: 2026-07-26
updated: 2026-07-26
related: ["[[Game-Gameplay-Patterns]]", "[[Analysis-Learning-Game-Fit]]", "[[Analysis-Play-Loop-Surfaces]]", "[[Simulation-vs-Director]]", "[[Storyteller-Driven-Pacing]]", "[[Emergent-Narrative-and-Losing-Is-Fun]]", "[[Indirect-Control-and-Divine-Intervention]]", "[[Observation-Driven-Play]]", "[[Difficulty-Dials-and-Dynamic-Depth]]", "[[Brief-and-Assumptions]]"]
---

# Analysis — The iteration rung: play-loop delta after the Staged Cockpit shipped

_The 2026-07-25 reorientation's entire course of action has shipped. With the
cockpit and scenario surfaces real, how well does the PLAY LOOP serve the
teaching-game lens now — and what does this corpus say the next
highest-leverage gameplay work is?_

## Standing on the prior layers

[[Analysis-Learning-Game-Fit]] (the pattern→purpose verdicts and the eight
ratified learning-game decisions) and [[Analysis-Play-Loop-Surfaces]] (the
same patterns projected onto screens, under the 2026-07-25 reorientation
decisions) are immutable decided context. This note is the **delta layer**:
what of those layers is now verified shipped, which of their open questions
got ruled, and what the corpus says about the loop that now actually exists.
[[Brief-and-Assumptions]]'s ambient-sim framing remains superseded, as
declared in [[Analysis-Learning-Game-Fit]].

## Given constraints (operator, 2026-07-26 — ratified, not open)

The 2026-07-26 steering round (`docs/design/reorient-2026-07-26-ui.md`
§Decisions, items 1–10) fixed what this note's draft round left open:

1. **Report-card truth (decision 1, NEW card, HIGH):** every report-card
   surface unifies on `sim.EvaluateRubric` NOW — a failed term renders ✗ at
   the postmortem — and `the-law`'s real evaluator is authored (persist the
   charter `Default` flag into state, the documented blocker at
   `internal/sim/scenario.go:277`). Sequenced after TASK-144.
2. **Design-gate semantic lint (decision 2, NEW card, small):**
   `scripts/check-tui-design.mjs` learns to catch `unbuilt (wave` cells on
   `status: shipped` pages; `overlays/postmortem.md`'s seven stale cells and
   `panels/exercise.md:110` are amended in the same PR.
3. **TASK-67 fork duel (decision 3): promoted Medium → HIGH**, reframed as
   the loop's **iteration rung** — all D7 prerequisites shipped; v1 =
   rubric-first scoreboard sharing `reportCardView` + `sim.EvaluateRubric`,
   then the HTML retelling; dual-TUI stays deferred.
4. **Exercise catalog wave (decision 5):** ladder v1 = 2–3 hand-authored
   exercises per stage; incident vocabulary grows to ~3 kinds (cold snap,
   forage blight, stranger/trickster arrival), each a reducer-valid event
   shape indistinguishable from an ambient cause.
5. **Lane order (decision 9): duel before faith** — TASK-67 lands before
   TASK-118, which lands before TASK-112 (per TASK-112's AC5/AC6).
6. **TASK-142 look-cursor runs in the next UI lane** (decision 4), with the
   DF fixed tile-content hierarchy and the tile registry's `meaning` rows as
   whatis content. Decisions 6–8 (in-TUI ladder view in the `?` guardian
   section; quickstart first-prompt step; mouse-parity sweep test) are
   likewise fixed.
7. **Parked as recorded watch items (decision 10), genuinely open:**
   proving-run timing vs the ambient drama floor (TASK-14/28/133 posture);
   rubric-gauge exposure doctrine (all-terms-live stands until
   multi-exercise content exists); teaching-signal instrumentation consent;
   retry accommodation; and the rest of the decision-10 list.

## Verdict

The 2026-07-25 course of action shipped essentially in full, and it landed
where this corpus pointed. Verified against code and the design corpus, not
claims: the exercise panel with live event-derived rubric gauges
(`internal/tui/exercise.go`, `docs/design/tui/panels/exercise.md`,
`status: shipped`) is [[Storyteller-Driven-Pacing]]'s legible rhythm on
screen; the director-lite scheduler with a real director seam
(`internal/sim/scenario.go` — the `incidentSource` interface,
`scheduleSource` v1; `docs/wiki/scenario-machinery.md`) is
[[Simulation-vs-Director]]'s two poles as a per-exercise forecast/fog
vocabulary (`sim.IncidentVisibilityFor`), built exactly on the
executor-emission precedent [[Analysis-Learning-Game-Fit]] demanded instead
of the injection door; the takeover pair carries the precise voice asymmetry
[[Analysis-Play-Loop-Surfaces]] specified — the morgue's no-blame register on
failure ("_Stated as evidence; the reader draws the lesson_",
`internal/scribe/morgue.go`), the player's-authorship voice on success
([[Emergent-Narrative-and-Losing-Is-Fun]], [[Difficulty-Dials-and-Dynamic-Depth]]);
the guardian strip makes the mana budget ambient with a faith slot reserved
under an honest never-claim-an-unshipped-mechanic rule
(`docs/design/tui/panels/guardian-strip.md`;
[[Indirect-Control-and-Divine-Intervention]]); and the Cogmind
identity-not-easiness move shipped as skin-resolved stage identities ("The
Voice" et al., `docs/wiki/curriculum-ladder.md`, `docs/wiki/skin.md`). The
chronicle's retention role ([[Observation-Driven-Play]]) is intact, and the
scenario-cadence narration trigger (`internal/mind/narrate.go`) fixed the
mismatch this branch found — a minutes-scale run now always yields a narrated
score-story. The anti-self-grading guard is encoded on the board (TASK-137
charter-delta experiment In Progress; TASK-112 AC5/AC6).

**The single biggest remaining gap: the loop is built but the game is one
lesson long, and the iteration rung is missing.** `sim.ScenarioExercises`
holds exactly two exercises; only `first-night` has a production rubric
(`the-law`'s charter conjunct is not state-derivable —
`scenario-machinery.md` §Operational notes); the incident vocabulary is one
kind (`gru_emerges`). And the corpus's tightest answer to "did my prompt edit
work?" — the fork duel — remained unbuilt even though everything it was
waiting on (rubric evaluator, report-card renderer, glossary discipline,
postmortem register) had shipped. Decisions 3/5/9 now close exactly this:
duel promoted HIGH as the iteration rung, catalog wave ratified, duel before
faith. A player can pass one exercise and see one ceremony; after this wave
they can *iterate* — which is the verb the learning game exists to teach.

## Reasoning

### What shipped, pattern by pattern (delta verification)

- **Director-lite, honestly built.** [[Analysis-Learning-Game-Fit]]'s honesty
  note ("nothing today injects at a future tick; ride the executor-emission
  precedent") was followed literally: `scenarioIncidentEvents` is a pure
  function of (state, boot-frozen config, tick), replay-safe, with the
  post-v1 live director named as a second `incidentSource` implementation —
  the Cassandra graduation has a seam waiting for it (spec 054, TASK-119
  Done). A scheduled `gru.emerged` is indistinguishable in kind from a rolled
  one, which decision 5 now elevates to doctrine for the new incident kinds.
- **Losing-is-fun, on screen.** The postmortem takeover renders the morgue's
  no-blame evidence live, opens automatically on re-attach to an ended world,
  and is replayable via `p` (`docs/design/tui/overlays/postmortem.md`
  FR-013) — permadeath-as-irreversibility given its reading surface, as
  [[Emergent-Narrative-and-Losing-Is-Fun]] required. FR-018 ruled the prior
  analysis's open question 1: the **ambient postmortem is morgue evidence
  only** — the hybrid-scoring boundary drawn on screen exactly where
  [[Analysis-Play-Loop-Surfaces]] left it open.
- **Three of the prior layer's five open questions are now ruled**: ambient
  postmortem contents (morgue-only, FR-018); forecast rendering (a
  vocabulary, never a boolean — `forecast`/`fog`, stage-keyed with
  per-definition override, D4); narration trigger (chapter at the pass/fail
  boundary only, additive to the day/night cadence). Question 2
  (rubric-gauge granularity) was decided *against* this branch's caution —
  all terms render live in the exercise panel — and is now a decision-10
  watch item (see Tensions). Question 5's survival-lane competence ceiling
  is becoming empirical rather than a priori via TASK-136/137.

### Reconciliation with the sibling branches (prose only, per vault isolation)

- **Report-card truth (Learning-Game-Design branch): adopted; primacy
  ceded.** This branch's draft flagged the same defect as content debt (the
  design corpus's own "known simplification" note: `agent.died: 2` renders ✓
  on a failing run); the pedagogy sibling found the sharper consequence — a
  checkmark on the run's failing outcome **at the most salient teaching
  moment the game has**, mis-teaching the rubric vocabulary the ladder rests
  on — and correctly ranked unification #1. This corpus reinforces their
  claim from its own ground: the postmortem's no-blame register
  ([[Emergent-Narrative-and-Losing-Is-Fun]] — evidence presented without
  judgment) only works if the evidence is TRUE; a false ✓ is not neutral, it
  is a lie in the one register the morgue contract promises never lies.
  Ratified as decision 1 (NEW HIGH card, `EvaluateRubric` everywhere, plus
  `the-law`'s evaluator — which this branch's catalog concern also needed).
  Ownership resolved: the pedagogy branch owns the finding; this branch's
  catalog wave (decision 5) depends on it and builds on top.
- **Wave ordering (Game-UI-UX branch): no residual conflict — parallel
  lanes.** The UI sibling ranks TASK-142 look-cursor first (the map's
  inspection dead-end, lens clause a); this branch ranked catalog + duel
  (lenses b/c). The decisions dissolve the collision: TASK-142 runs in the
  next **UI lane** (decision 4) while TASK-67 is promoted HIGH in the
  **game-loop lane** (decisions 3/9) — they don't contend for the same slot.
  This branch's residual stance, recorded not relitigated: if the lanes ever
  DO contend, the iteration rung goes first — the observe side of the
  observe/intervene rhythm is already corpus-validated strong (the UI
  sibling's own verdict), while the learn-iterate loop is incomplete without
  the duel; a map you can interrogate teaches reading, a duel teaches the
  game's central verb.
- **Fork duel: one surface, three corpus lenses, one framing ratified.** The
  pedagogy branch grounds it as the Opus-Magnum-histogram
  intrinsic-optimization engine; the player-docs branch as the last
  unshipped teaching surface and the losing-is-fun postmortem artifact; this
  branch as the iteration rung completing learn→try→compare→retry. These
  compose — decision 3 ratifies the iteration-rung framing verbatim and the
  shared-renderer requirement (`reportCardView` + `sim.EvaluateRubric`)
  binds all three. The HTML retelling stays sequenced second, which keeps
  the Boatmurdered export ([[Emergent-Narrative-and-Losing-Is-Fun]] — the
  celebrated story object is a retelling) alive as duel phase 2 rather than
  a separate task.
- **Semantic lint (Game-Player-Docs branch): adopted.** Their #1 finding —
  the gate validates structure, not semantics, and `overlays/postmortem.md`
  ships seven `unbuilt (wave 4)` renderer cells on a `shipped` page — sits
  on THIS branch's flagship surface. Lens clause (d) protects this corpus's
  surfaces too; decision 2 mechanizes it. Adopted without reservation.
- **First-prompt quickstart step (Game-Player-Docs): adopted, with this
  corpus's grounding.** A getting-started walkthrough that never has the
  player prompt the guardian is the same discoverability failure class as
  Cogmind's 9.5% non-default adoption ([[Difficulty-Dials-and-Dynamic-Depth]]
  — the settings existed; players never met them). Decision 7 fixes it
  content-side; the in-TUI ladder view (decision 6) fixes the forward-ladder
  half — StS-style visible-next-identity is the informed-choice menu this
  corpus's rebrand evidence demands, rendered at the deterministic floor.
- **TASK-133 framing: merged.** The UI sibling routes the neglect alert
  through the shipped severity grammar (chronicle whole-line alert + map
  overlay, never a new channel); this branch frames neglect as the observed
  failure shape the postmortem cannot yet explain (Oak's warmth slide,
  zero warmth-class intents — death the no-blame register can display but
  not attribute). Both go on the card; no conflict.
- **Retry accommodation (pedagogy branch idea 5): genuinely contested,
  correctly parked.** Their Hades-God-Mode proposal collides with this
  corpus's Cogmind evidence that easier settings "somewhat destabilize" a
  tight design ([[Difficulty-Dials-and-Dynamic-Depth]]) and with the faith
  failure-spiral floor question carried by TASK-118. The honesty-marker
  mitigation (the `stage_overridden` precedent) is the right shape IF it
  ever ships; whether it ships is an identity question for the operator.
  Parked under decision 10 — kept open here, deliberately.

### Recommendations as they now stand (post-decisions, build-relevant order)

1. **Report-card truth first** (decision 1, after TASK-144) — everything
   downstream (duel scoreboard, catalog, postmortem) renders through it.
2. **Fork duel v1** (decision 3, TASK-67 HIGH): rubric-first scoreboard on
   the shared renderer; a lost duel IS a postmortem — same register, same
   glossary. HTML retelling second.
3. **Exercise catalog wave** (decision 5): 2–3 exercises per stage, ~3
   incident kinds, every kind reducer-valid and ambient-indistinguishable.
   Cold snap should be designed WITH TASK-28's seasons session in view —
   one mechanic, two duties (authored incident + ambient drama supply).
4. **Faith** (TASK-118, after the duel per decision 9, before TASK-112):
   the strip's reserved segment and the tutor-lane exclusion (TASK-112 AC6)
   are already specified; the failure-spiral floor is the spec's one open
   design choice.
5. **Live director persona** stays post-catalog: the `incidentSource` seam
   is built; do not start before decision 5's catalog exists, or the
   director has nothing to schedule ([[Storyteller-Driven-Pacing]]).

## Tensions & tradeoffs

- **All-terms-live rubric gauges vs metric-gaming** — this branch's caution
  lost to the pedagogy corpus's histogram evidence (visible, true,
  evidence-derived feedback drives intrinsic optimization), and the shipped
  panel renders every term live. Decision 10 parks it with all-terms-live
  standing until multi-exercise content exists. Held consciously: the
  reopening signal is playtest evidence of players optimizing counters over
  the village — which requires the instrumentation question (next item) to
  be answerable.
- **The decision-6 ceremony watch item is still unobservable.** The prior
  layer named "playtest evidence of ceremony fatigue" as the reopening
  signal; the pedagogy sibling showed nothing can produce that evidence (no
  dismiss-latency or dwell instrumentation exists). Instrumentation consent
  is parked under decision 10 — until it's granted or refused, both this
  watch item and the gauge-exposure one above are gates without gauges.
- **The ambient drama floor** — uniquely this branch's concern, unaddressed
  by any sibling and parked by decision 10. The DF pole needs deep
  interacting systems producing unscheduled trouble; today's ambient roster
  is the gru (wounds-mostly) plus night cold, with zero deaths in 5.2
  game-days of TASK-122 evidence. A graduate arrives in an endgame that is
  Phoebe Chillax forever ([[Storyteller-Driven-Pacing]]). TASK-28 (seasons),
  TASK-23 (interaction v2 — the social drama generator the Thornspire
  cascade proved the substrate can host, [[Observation-Driven-Play]]), and
  TASK-133 are the un-worked supply; TASK-14's proving run will measure a
  calm world until some of it lands. Parked, not resolved — the catalog wave
  buys time (scenarios carry the drama for now), but the endgame's contract
  with a graduated player is unfunded.
- **This branch's MOC needs a framing update it cannot give itself** (notes
  are immutable; recorded here): [[Game-Gameplay-Patterns]]'s open question
  3 — "does any shipped game combine a director AI with an LLM-agent
  simulation?" — is now answered by this project: spec 054 is exactly that
  combination (authored director-lite over an LLM-agent village). A future
  corpus pass should record promptworld as the existence proof. Likewise
  [[Brief-and-Assumptions]] stays superseded; its "watching is primary"
  assumption must only be read through [[Analysis-Learning-Game-Fit]]'s lens
  correction.

## Confidence & open questions

High confidence in the shipped-state verification (every claim traces to a
pinned design page, a wiki note, code, or a board artifact) and in the
sibling reconciliation — every overlap resolved into adoption, ratified
decision, or a recorded parked tension; no conflict papered over. The one
position held against a sibling (iteration rung before look-cursor if lanes
contend) is moot under the current parallel-lane plan and recorded only as a
contingency stance. Genuinely open, all parked by decision 10 with named
resurfacing moments:

1. **Proving-run timing vs the ambient drama floor** — run TASK-14 on the
   calm world and accept a Phoebe-register finding, or land TASK-28/133
   first? Resurfaces when TASK-14 is next scheduled.
2. **Rubric-gauge exposure doctrine** — all-terms-live stands; revisit when
   decision 5's multi-exercise catalog exists and (if consented)
   instrumentation can show gaming.
3. **Teaching-signal instrumentation consent** — gates two watch items;
   local-only event emission is the proposal on the table.
4. **Retry accommodation** — in or out of the game's identity; the corpus
   tension (Cogmind's destabilization cost vs Hades' completion evidence)
   is recorded above for whenever it resurfaces.
5. **Carried, still open from [[Analysis-Learning-Game-Fit]]:** the
   survival-lane competence ceiling — now expected to be answered
   empirically by TASK-136/137 evidence rather than by decree.

## Basis

- [[_grounding]] — all sourced pattern claims: RimWorld storyteller rhythm,
  DF/roguelike failure doctrine and Boatmurdered, god-game mana economy,
  Progress Quest retention, Smallville cascade, Cogmind rebrand
- [[Analysis-Learning-Game-Fit]], [[Analysis-Play-Loop-Surfaces]] — the
  immutable decided layers this delta stands on
- [[Storyteller-Driven-Pacing]], [[Simulation-vs-Director]] — director-lite
  verification and the live-director graduation
- [[Emergent-Narrative-and-Losing-Is-Fun]] — the postmortem register, the
  report-card-truth adoption, the retelling-as-duel-phase-2
- [[Indirect-Control-and-Divine-Intervention]] — the guardian strip and the
  faith sequencing
- [[Observation-Driven-Play]] — chronicle retention, the narration-trigger
  fix, the interaction-v2 drama framing
- [[Difficulty-Dials-and-Dynamic-Depth]] — stage identities shipped, the
  first-prompt/ladder-view discoverability grounding, the retry tension
- [[Brief-and-Assumptions]] — superseded (still superseded)
- Project artifacts cited in prose:
  `docs/design/reorient-2026-07-26-ui.md` (decisions 1–10, given
  constraints), `docs/design/reorient-2026-07-25-ui.md` (the prior run),
  `docs/design/tui/` (INDEX.md, panels/exercise.md, panels/guardian-strip.md,
  overlays/postmortem.md, overlays/ceremony.md, patterns/stage-defaults.md),
  `docs/wiki/scenario-machinery.md`, `docs/wiki/curriculum-ladder.md`,
  `docs/wiki/skin.md`, `docs/wiki/morgue.md`, `docs/wiki/chronicle.md`,
  `internal/sim/scenario.go`, `internal/tui/exercise.go`,
  `internal/mind/narrate.go`, `internal/scribe/morgue.go`; board tasks
  TASK-14/23/28/67/112/118/119/122/127/133/136/137/144; the three sibling
  2026-07-26 evaluations (Game-UI-UX, Learning-Game-Design,
  Game-Player-Docs branches) reconciled in prose, never linked, per vault
  isolation. Wiki-grounding mode: INDEX.md routing + just-in-time full
  notes (scenario-machinery; targeted reads of curriculum-ladder/skin); no
  CAPSULES.md exists.
