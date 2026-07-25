---
title: Analysis — Play-loop surfaces: what UI the gameplay patterns demand on screen
aliases: [Play Loop Surfaces]
tags: [analysis, gameplay, ui-ux, learning-game]
type: analysis
created: 2026-07-25
updated: 2026-07-25
related: ["[[Game-Gameplay-Patterns]]", "[[Analysis-Learning-Game-Fit]]", "[[Simulation-vs-Director]]", "[[Storyteller-Driven-Pacing]]", "[[Emergent-Narrative-and-Losing-Is-Fun]]", "[[Indirect-Control-and-Divine-Intervention]]", "[[Observation-Driven-Play]]", "[[Difficulty-Dials-and-Dynamic-Depth]]", "[[Brief-and-Assumptions]]"]
---

# Analysis — Play-loop surfaces: what UI the gameplay patterns demand on screen

_Given this branch's gameplay patterns and the ratified learning-game direction, what must the play loop's UI surfaces be — and how do the 2026-07-25 reorientation decisions settle them?_

## Standing on the prior analysis

[[Analysis-Learning-Game-Fit]] is decided context: its ratified recommendations
(scenario runs as the v1 spine, fork-duel as grader, faith-driven mana,
grounded tutor surface, fiction-named stage identities) fed
`docs/design/learning-game-synthesis.md` and are not relitigated here. This
note is the next layer down: the same patterns projected onto **screens,
panels, and controls**, under the reorientation lens (UI must make the village
legible at a glance, make prompting the guardian the central rewarding verb,
teach itself in play, and be fully specifiable page-by-page).

## Given constraints (operator, 2026-07-25 — recorded, not open)

The reorientation steering round ratified eight UI decisions plus thirteen
lead-adopted defaults (`docs/design/reorient-2026-07-25-ui.md`); the ones this
branch's evidence bears on, taken as fixed:

1. **Staged Cockpit** is the headline direction: a first-class guardian
   console page, map-legibility investment (jump-to-source, villager strip,
   condition overlays), stage-shaped **layout defaults** (defaults only —
   everything reachable at every stage; capability locks stay angel-only).
2. **Both big moments seize the screen** (decision 6): the stage-unlock
   ceremony takes over on `curriculum.stage_unlocked`, the postmortem takes
   over on `run.ended`; both dismissable and replayable.
3. **Guardian strip above the minibuffer** (decision 7): the action budget
   (charges, regen, standing orders, faith once TASK-118 lands) always
   visible, paired with the input line.
4. **Exercise panel** (D11): scenario worlds get a fourth dock tab — framing,
   event-derived rubric gauges, pass/fail state, incident forecast, an
   attach-time briefing, and a scenario-cadence narration trigger so the
   chronicle score-narrative renders during short runs.
5. **Incident-schedule visibility is a per-exercise field** (D4): visible
   forecast at stages 1–2, fog from stage 3.
6. **Fork-duel v1 compare surface** (D7): rubric-first scoreboard with
   drill-down, then the shareable HTML retelling; dual side-by-side TUI
   deferred.
7. **Telemetry splits to a systems tab** (D10); the guardian tab carries
   fiction-layer content only. **Report card** renders as a card in the
   guardian console at natural stopping points and inside the postmortem
   (D5). **Unlock attribution voice is the player's authorship** (D6),
   skin-resolved. **Design-ref v2 with a freshness gate** is the UI authority
   (decision 4); **skin tokens ship before new fiction literals** (D2).

## Verdict

The shipped TUI is a first-class observation instrument and the corpus's
watching patterns land on it almost move-for-move — the chronicle ring is
Progress Quest's "events indicate a narrative about the digital person running
on your machine" ([[Observation-Driven-Play]]), the ENDED/grave/morgue posture
is permadeath-as-irreversibility done honestly
([[Emergent-Narrative-and-Losing-Is-Fun]]; `docs/wiki/morgue.md`). What the
play loop lacked was every surface for the **game**: no on-screen goal, no
rubric progress, no visible action budget outside one dock tab, no in-TUI
failure reading beyond a red token, and a stage identity reduced to a header
suffix. The decisions above close each of those gaps in the direction this
branch's evidence supports — the exercise panel is the storyteller pattern's
legible rhythm ([[Storyteller-Driven-Pacing]]), the guardian strip is the mana
economy made ambient ([[Indirect-Control-and-Divine-Intervention]]), the
takeover pair is failure-as-content and progress-as-identity given their
ceremony ([[Emergent-Narrative-and-Losing-Is-Fun]],
[[Difficulty-Dials-and-Dynamic-Depth]]). What remains genuinely open is
smaller and listed at the end: the unscored ambient world's postmortem
contents, rubric-gauge granularity vs metric-gaming, and the forecast's
rendering shape.

## Reasoning

### What the corpus pins to each decided surface

- **Exercise panel (D11) ← director-lite + hybrid scoring.** A learning game
  needs lessons on schedule and a legible arc; RimWorld's Cassandra makes the
  rhythm knowable (cooldowns, pressure/release — [[Storyteller-Driven-Pacing]])
  while DF's pole leaves drama unscheduled ([[Simulation-vs-Director]]). D4's
  per-exercise visibility field is exactly the two poles as a dial: visible
  forecast at stages 1–2 (the teaching register), fog from stage 3 (the DF
  register as an earned graduation). The rubric gauges must be event-derived
  projections — every rubric term is a cataloged event type
  (`docs/wiki/curriculum-ladder.md`), so the panel is the decision-trace
  pattern applied to the run itself, never a parallel scoring engine. The
  scenario-cadence narration trigger fixes a real mismatch this analysis
  found: the narrator closes chapters at day/night boundaries (~2/game-day,
  `docs/wiki/chronicle.md`), so a minutes-scale scenario at watchable speed
  could end with zero narrated score-story; without the trigger, the
  chronicle-as-score-narrative decision rides a surface that never fires.
- **Takeover pair (decision 6) ← losing-is-fun + identity framing.** The
  corpus says the down-period after failure is where the lesson lands and the
  celebrated story object is the retelling
  ([[Emergent-Narrative-and-Losing-Is-Fun]] — Boatmurdered; Wichman's
  irreversibility-not-pain). Today `run.ended` renders one bold-red token and
  the morgue is a disk file (`docs/wiki/tui-client.md`,
  `docs/wiki/morgue.md`): the postmortem takeover gives the run's legacy
  document an in-game reading surface — run summary, epitaphs with the
  charter-revision alignment ("the angel's watch at that moment" is the
  teaching payload), rubric outcome, report card (D5), and jump-offs
  (retry / fork) on the reserved `⏎` seam (D3). The unlock ceremony is the
  same machinery pointed at success: Cogmind's rebrand evidence says identity
  lives in prominent, informed moments, and its 9.5%-adoption failure was a
  *discoverability* failure ([[Difficulty-Dials-and-Dynamic-Depth]]) — a
  one-line digest entry scrolling at 16x reproduces it; a seized, replayable
  ceremony does not. One deliberate asymmetry to keep: the ceremony speaks in
  the player's-authorship voice (D6) while the postmortem stays inside the
  morgue's no-blame vocabulary contract — competence is attributed on
  success, evidence is presented without judgment on failure. Both registers
  are corpus-grounded and they must not bleed into each other.
- **Guardian strip (decision 7) ← the mana economy.** The genre's loop only
  teaches restraint and timing if the budget is ambient like the clock
  ([[Indirect-Control-and-Divine-Intervention]]: intervention "budgeted and
  endogenous"). The charge bank currently hides in one tab's pane header —
  invisible exactly while the player watches the chronicle. Pairing budget
  with the input line also makes the minibuffer read as THE verb, which is the
  lens's central-verb clause in one row. Faith (TASK-118) joins the strip when
  it lands, closing the endogenous feedback loop visually: better prompting →
  truer visions → visible faith → visible charges.
- **Fork-duel scoreboard (D7) ← retold, not rendered.** The comparison's
  product is the narrative of what your change did, not simultaneity
  ([[Emergent-Narrative-and-Losing-Is-Fun]]: emergent narrative is "retold" by
  players); hence rubric-first with drill-down, then the HTML retelling — the
  dual live TUI was the expensive option the evidence didn't demand. A lost
  duel is a postmortem: the scoreboard and the takeover should share the
  rubric-rendering and glossary discipline (no raw enums in a grade).
- **Systems-tab split (D10) ← legibility + the skin boundary.** The guardian
  tab had accreted provider tables, horizon rows, and spend beside the
  transcript (`docs/wiki/tui-client.md`); mixing fiction-layer content with
  engine telemetry both muddies at-a-glance reading and would have made
  TASK-121's skin sweep ambiguous control-by-control. The split is the skin
  boundary drawn as UI regions.

### Reconciliation with the sibling branches (prose only, per vault isolation)

- **Unlock ceremony (Learning-Game-Design branch) vs postmortem takeover
  (this branch): unified, adopted.** The sibling proposed the stage-unlock
  ceremony with a badge-until-pause interrupt policy; this branch proposed the
  postmortem takeover with an ambient/scenario split. The operator chose
  maximum salience for both (decision 6), overriding both branches' caution.
  Adopted as one **takeover surface family**: shared body-replacement
  rendering (the help overlay precedent), shared dismiss/replay contract
  (`?`, `stages`, the morgue file), skin-token copy (D2), and the two voice
  registers above. The interrupt-policy worry mostly dissolves on inspection:
  unlocks fire on exercise passes, which happen in scenario worlds where the
  ceremony IS the lesson's climax; `run.ended` interrupts nothing (the world
  is over). Residual tension recorded below.
- **Exercise goal panel (Learning-Game-Design branch) vs exercise panel
  (this branch): converged, merged.** Same surface from two evidence bases
  (their GameFlow clear-goals criterion; this branch's storyteller
  legibility). Their proposed placement (a guardian-pane header extension)
  lost to the dock tab; their **attach-time briefing** (the scenario's
  fiction as the opening beat) won and is folded into D11. Their dwell-lesson
  row (decision 5) is complementary, not competing: lessons teach the UI, the
  exercise panel states the goal.
- **Staged Cockpit (Game-UI-UX branch): adopted as the host.** This branch's
  surfaces are now rooms in their house — the guardian console hosts the
  report card cards (D5) and the tutor lane; the villager strip (D12) is the
  needs-floor legibility this branch's rubric terms read against (a rubric
  gauge over needs floors is only teachable if needs are glanceable); the
  chronicle `⏎` jump-to-source (D3) is the same reserved seam this branch's
  postmortem jump-offs use. Their stage-shaped **layout defaults** (decision
  3) subsume this branch's "locked-capability rendering as fiction"
  recommendation in part; the remainder (the next-stage teaser in fiction
  voice) rides D9's model-free "the guardian" help-overlay section.
- **Design-ref v2 with control tables (Game-Player-Docs branch): adopted;
  this branch's staleness findings are its evidence.** This analysis
  independently found `docs/design/tui/panels/dock.md` specifying a
  transcript-only guardian tab while the shipped pane carries roughly six
  specs of accreted content — the page-by-page authority had already quietly
  failed. The sibling's structure (pages/ · panels/ · overlays/ · patterns/,
  uniform control tables with skin-token columns, `verified_against` pins, a
  same-PR freshness gate) is the fix, ratified as decision 4. This branch's
  contribution lands there as spec-before-build pages: `panels/exercise.md`,
  the takeover pair under `overlays/`, and the guardian strip in the
  minibuffer panel page.
- **Skin-token sequencing (Game-Player-Docs branch, D2): adopted, with a
  consequence for this branch's surfaces.** Every fiction string the exercise
  panel, ceremony, and postmortem introduce (framing text, ceremony copy —
  the morgue's frame lines are already contract-pinned and stay put) is
  authored as skin tokens from day one, or TASK-121's sweep grows unboundedly.

No irreconcilable conflicts among the four branches: every overlap resolved
into adoption, merger, or an operator decision. The one position this branch
holds against the grain of decision 6 is recorded as a tension, not a
conflict.

### Recommendations as they now stand (post-decisions, build-relevant order)

1. **Spec the exercise panel first** (`panels/exercise.md` in design-ref v2,
   control table included): framing, rubric gauges (event-derived, glossary
   discipline), pass/fail state, forecast per D4's per-exercise field,
   attach-time briefing, scenario-cadence narration trigger. TASK-119's spec
   should absorb the UI acceptance criteria.
2. **Spec the takeover family as one artifact** (`overlays/postmortem.md` +
   `overlays/stage-unlock.md` sharing a pattern page): shared machinery,
   dismiss/replay contract, the two voices (authorship on success, no-blame
   evidence on failure), report card placement per D5, `⏎` jump-offs per D3.
3. **Guardian strip** (decision 7): one line — charges, regen countdown,
   order count, faith slot reserved for TASK-118. Cheapest of the set;
   unblocks the "budget is ambient" teaching immediately.
4. **Fork-duel scoreboard** (D7) sharing the rubric renderer with the
   postmortem; the HTML retelling reuses the same fold (the Boatmurdered
   export from [[Analysis-Learning-Game-Fit]], now sequenced second).
5. **Systems-tab split (D10)** rides whichever PR first touches the guardian
   tab — it is a precondition for TASK-121's clean sweep.

## Tensions & tradeoffs

- **Decision 6 vs the never-interrupt evidence.** The operator chose salience
  over the corpus-cautious badge-until-pause. Mostly safe (see
  reconciliation), but two edges want watching: an unlock earned while the
  player is mid-read at high speed seizes the screen away from the thing they
  were reading (the replay contract must be flawless), and if unlock-capable
  exercises ever run inside the ambient endgame world, a seizure there
  strains that world's DF-pole contract. Watch-item, not relitigation.
- **Live rubric gauges vs gaming the metric.** [[Analysis-Learning-Game-Fit]]
  warned event-derived scores invite optimizing the metric over the village;
  D11 makes the metric maximally visible. The mitigation is rubric *content*
  (score outcomes the fiction cares about), but panel design has a lever too:
  headline gauges live, full term-by-term breakdown reserved for the
  postmortem. Open below.
- **Forecast fog (D4, stage 3+) is not free.** Fog restores the DF register
  but removes the schedule legibility that made stage-1 scenarios teachable;
  a stage-3 player who has only ever seen forecasts meets unscheduled drama
  for the first time inside a scored run. The per-exercise field permits a
  middle register (incidents named, timings hidden) — worth an explicit
  vocabulary in TASK-119's spec rather than a boolean.
- **The takeover pair adds two screens outside the five-region anatomy.** The
  body-replacement precedent (help overlay) keeps chrome and exact-height
  invariants, but the design ref's anatomy page must document the overlay
  class explicitly or the "fully specifiable" promise degrades again.

## Confidence & open questions

High confidence in the pattern→surface mapping and the reconciliation — every
overlap with the siblings resolved cleanly, and each decided surface traces to
sourced corpus claims. Moderate confidence in the exercise panel's gauge
design: this corpus has no learning-game-proper sources, and the commissioned
follow-up vault pass should inform gauge/feedback granularity before TASK-119
finalizes. Genuinely open, decision-shaped:

1. **What does the ambient endgame world's postmortem contain?** Decision 6
   applies the takeover to every `run.ended`, but the ambient world is
   unscored per the synthesis's hybrid-scoring decision. Morgue evidence +
   epilogues only, or also the TASK-115 report card (charter-attribution is
   not a score, so it may pass the boundary)? Bounds what "unscored" means on
   screen.
2. **Rubric-gauge granularity:** all rubric terms live in the exercise panel,
   or headline terms live with the full breakdown at the postmortem? (The
   metric-gaming tradeoff above; per-exercise field or global doctrine?)
3. **Forecast rendering vocabulary:** countdown lines ("night falls in 2
   hours"), a schedule list, or the middle register (incidents named, timing
   fogged)? TASK-119 spec-level, but it decides how much of the storyteller's
   rhythm the player can plan against at each stage.
4. **Scenario-cadence narration trigger:** chapter close on exercise
   resolution only, or per landed incident? Decides whether the score
   narrative reads as one verdict-story or a running commentary — and its
   call cost.
5. **Carried from [[Analysis-Learning-Game-Fit]], still open:** the survival
   lane's competence ceiling (it bounds what the postmortem may honestly
   attribute) and the audience question (it decides whether the HTML
   retelling is v1 or deferred).

## Basis

- [[_grounding]] — all sourced pattern claims: RimWorld storyteller rhythm and
  legibility, DF/roguelike failure doctrine and Boatmurdered, god-game mana
  economy, Progress Quest event-feed retention, Cogmind difficulty
  rebrand/discoverability
- [[Analysis-Learning-Game-Fit]] — the decided recommendation layer this note
  builds on (immutable; not relitigated)
- [[Storyteller-Driven-Pacing]], [[Simulation-vs-Director]] — the two pacing
  poles behind D4's visibility dial and the forecast tradeoff
- [[Emergent-Narrative-and-Losing-Is-Fun]] — the postmortem takeover, the
  retelling-first compare surface, the no-blame register
- [[Indirect-Control-and-Divine-Intervention]] — the guardian strip's
  always-visible budget
- [[Observation-Driven-Play]] — the chronicle's retention role and the
  watch-is-the-reward framing the exercise panel must not displace
- [[Difficulty-Dials-and-Dynamic-Depth]] — stage identity, the ceremony's
  discoverability argument
- [[Brief-and-Assumptions]] — the superseded ambient-sim lens (still
  superseded)
- Project artifacts cited in prose: `docs/design/reorient-2026-07-25-ui.md`
  (the eight decisions + D1–D13, given constraints),
  `docs/design/learning-game-synthesis.md` (the eight learning-game
  decisions), `docs/wiki/tui-client.md`, `docs/wiki/curriculum-ladder.md`,
  `docs/wiki/morgue.md`, `docs/wiki/chronicle.md`, `docs/wiki/metatron.md`,
  `docs/design/tui/` (INDEX, panels/dock.md staleness finding), and board
  tasks TASK-67/68/115/117/118/119/121; the three sibling reorientation
  reports (Game-UI-UX, Learning-Game-Design, Game-Player-Docs branches)
  reconciled in prose, never linked, per vault isolation
