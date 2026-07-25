---
title: Analysis — Do the researched gameplay patterns serve promptworld as a learning game?
aliases: [Learning Game Fit]
tags: [analysis, gameplay, learning-game]
type: analysis
created: 2026-07-25
updated: 2026-07-25
related: ["[[Game-Gameplay-Patterns]]", "[[Simulation-vs-Director]]", "[[Indirect-Control-and-Divine-Intervention]]", "[[Emergent-Narrative-and-Losing-Is-Fun]]", "[[Observation-Driven-Play]]", "[[Difficulty-Dials-and-Dynamic-Depth]]", "[[Storyteller-Driven-Pacing]]", "[[Brief-and-Assumptions]]"]
---

# Analysis — Do the researched gameplay patterns serve promptworld as a learning game?

_Which of this branch's patterns fit promptworld's actual purpose — an LLM prompting → agents-and-tools LEARNING GAME that is fun, engaging, and addictive in a good way — and what gameplay follows from them?_

## Lens correction (supersedes the brief's assumption)

[[Brief-and-Assumptions]] framed promptworld as "a TUI ambient simulation of an LLM-agent village where the player mostly reads a story feed." That assumption is now superseded (operator, 2026-07-24/25): promptworld's purpose is a **learning game** that teaches prompting, agent design, and tool use experientially; the ambient sim is the medium, not the point. The brief and [[_grounding]] stay immutable per vault discipline — this note declares the updated lens, and every verdict below is taken under it. Where the corpus's framing ("watching is a viable primary verb," [[Observation-Driven-Play]]) conflicts with the learning purpose, the learning purpose wins: **writing prompts is the primary verb; watching is the reward.**

## Operator decisions (given constraints, not open questions)

Recorded 2026-07-25; the recommendations below assume them:

1. **Session shape: staged.** Short seeded scenario runs teach the early curriculum; the long-lived ambient world is the endgame a player graduates into.
2. **Director-lite first.** No new director subsystem in v1. Scenario worlds get **pre-scripted incident schedules** through existing genesis/injection machinery; a live state-watching storyteller is a later graduation.
3. **Scoring: hybrid.** Scenario runs and fork-duels get explicit event-derived rubrics and pass signals; the ambient endgame world stays unscored, chronicle-only.
4. **Mana: faith-driven endogenous regen.** Metatron's charge regeneration becomes a pure reducer function of village faith state — landed true visions build faith, failed prophecies decay it.

Also decided (sibling player-docs branch reconciliation, duplicated here per vault isolation): the angel IS allowed to be the tutor, backed by a read-only registry-derived `explain` tool so mechanics answers are grounded, not confabulated; villager deaths stay unsoftened everywhere, with DF-style "your first village will probably freeze" reassurance up front; onboarding (`?` overlay, first-occurrence lessons) is a TUI-level property of every world, auto-retiring, with curriculum stages a separate opt-in layer.

## Verdict

promptworld already owns the best *instrument* in this design space: its control surface literally IS assistant configuration (charter.md / skills/ / capabilities.json — CLAUDE.md- and SKILL.md-shaped, TASK-64, docs/wiki/metatron.md), its feedback signal is a persisted per-tool-call verdict trail (TASK-52/63), and its intervention economy is the god-game mana pattern made of real LLM turns ([[Indirect-Control-and-Divine-Intervention]]). Four of the branch's five pattern families map directly onto shipped systems. What was missing — a game loop with goals, player-attributable failure, pacing, and a grade — is now largely resolved by the operator's decisions: **staged scenario runs with scripted incidents and event-derived rubrics are the v1 game**, and the ambient world is the DF-style unscored endgame. The remaining structural risk is the angel: as it gains autonomy (TASK-111/112) *and* the tutor role, the design must keep the player's authored text the binding variable, or the game grades itself.

## Reasoning

### Patterns that fit, and what they attach to

- **Indirect control + mana economy → the charge bank.** The genre template ([[Indirect-Control-and-Divine-Intervention]]: no unit control, world-level verbs, budgeted power) maps 1:1 onto omens/visions/miracles priced from one event-sourced bank (docs/wiki/metatron.md, metatron-miracles.md; TASK-59/27). Budgeting is what makes prompting a skill: world-01 showed a vague context wastes charges on door-rejected miracles (3 of 4 rejected on bad coordinates, TASK-111). The grounding's loop is *endogenous* — "power derived from the size and prosperity of the population of worshipers" ([[_grounding]] § God games) — and the operator has now adopted exactly that (decision 4, faith-driven regen). This closes the genre's positive feedback loop: better prompting → truer prophecies → more faith → more power. It is the macro-arc the learning game was missing.
- **Observation play → the chronicle, and the Smallville cascade is already live.** Progress Quest's retention hook — "events indicate a narrative about the digital person running on your machine" ([[Observation-Driven-Play]]) — is the chronicle ring + moments queue (TASK-11, docs/wiki/chronicle.md). The single-nudge cascade ([[_grounding]] § Generative Agents, the Valentine's party) has been *reproduced in production*: one player omen (world-01 seq 50664) birthed the village-wide Thornspire myth, 271 events (TASK-81). Strongest evidence the substrate can teach "small prompt, big consequence."
- **Losing is fun → permadeath already exists; stakes don't yet.** `agent.died` is permanent and irreversibility-not-pain is the shipped doctrine ([[Emergent-Narrative-and-Losing-Is-Fun]]); but the gru "wounds, never kills" (docs/wiki/gru.md — zero deaths in 1257 game days) and healing is passive, so nothing is at stake and intervention is decorative. TASK-31 (run outcomes, morgue file, gru lethality) and TASK-30 (labor budget) are not survival tuning — they are prerequisites for the rubric meaning anything. The operator's "deaths stay unsoftened, reassure up front" decision commits to this pole; the morgue file is the retellable artifact (the Boatmurdered lesson: the celebrated story object is a *retelling*, [[_grounding]] § DF).
- **Difficulty rebranding → capability stages as identity.** Cogmind's Rogue/Adventurer/Explorer rebrand plus first-startup choice ([[Difficulty-Dials-and-Dynamic-Depth]]) maps onto TASK-68 stage presets over TASK-64's manifest: a stage is "what your angel can do," never "easy mode." The teaching-speed posture (TASK-78/decision-6 — exceeding the cap prints the horizon arithmetic) is difficulty-framing that *teaches the core system*, the RimWorld storyteller-as-difficulty move done with real numbers.
- **Dynamic depth → already real.** Chronicle-only reading → decision-trace view (TASK-63) → raw event log; charter-only play → skills → capability manifests → Starlark tool bundles (TASK-85). Same run, self-selected depth — and the depth axis *is* the curriculum axis (prompting → instruction files → agents → tools).
- **Smallville emergence → exceeded.** Memory stream/reflection/planning parallels memories → consolidation → planner (docs/wiki/agent-mind.md); the social fabric and governance stack (TASK-8/13: rumors with provenance decay, self-legislation, exile) go beyond the paper. This is the drama generator the scenarios will narrate.

### Patterns that conflicted, and how the decisions resolve them

- **Simulation vs director** ([[Simulation-vs-Director]]): promptworld sits at the DF pole — events from systems, no arc guarantee — while a learning game needs lessons to arrive on schedule ([[Storyteller-Driven-Pacing]]'s watchers + incident generators). Resolution (decision 2): **phased**. V1 scenario worlds carry authored incident schedules; the live director is a graduation. One honesty note: "existing machinery" is *almost* true — nothing today injects at a future tick on schedule. The nearest shipped primitives are executor emissions that are pure functions of (state, tick) (charge regen, order expiry — docs/wiki/metatron-orders.md) and system-origin standing orders. Director-lite v1 most plausibly rides an executor-style scenario block in world config (deterministic, replay-safe by the same argument as charge regen), not the InjectSocial door. Small design work, not a new subsystem — but it should be scoped eyes-open.
- **Ambient 24/7 posture vs the learner's minutes-scale loop.** The substrate optimizes for "days pass unattended" (docs/wiki/overview.md); classroom mode (TASK-66/77/78) already fights this correctly. The staged decision completes the resolution: scenario runs are short by construction; the ambient posture is the endgame's feature, not the tutorial's bug.
- **Unscored purism vs pedagogy.** DF's "no win condition" and Cogmind's "tight design destabilized by easier settings" ([[_grounding]]) argue against scoring; tutorials need pass signals. Decision 3 (hybrid) threads it: rubrics live where the fiction is a lesson (scenarios, fork-duels), the graduated world keeps the DF contract — the chronicle is the only mirror. This matches the sources rather than fighting them.
- **The angel: tutor, autonomous actor, and unreliable manual — one reconciled position.** The sibling player-docs branch established that asking Metatron is simultaneously the help command *and* a rep of the skill being taught, and warned the angel is an unreliable manual without deterministic grounding; this branch's concern was the mirror image — TASK-111/112 autonomy must not erase the player's authorship. These converge on one design position for the initiative frame: **three lanes.** (a) *Tutor lane* — always-on, charge-free, grounded by the registry-derived read-only `explain` tool (decided), so mechanics answers are derived facts, never charter-colored guesses; explaining is exempt from initiative restrictions. (b) *Survival lane* — autonomous per TASK-111 (genesis watch orders, act-on-own for near-death), still charge-gated. (c) *Ambition lane* — everything beyond survival (world-shaping, clock control, standing orders) stays player-requested or pre-authorized, exactly the current `metatronInitiativeFrame` contract. The lane boundaries are where the pedagogy lives: the charter must remain the binding variable in lanes (b) and (c) — a default-charter angel should demonstrably underperform an authored one, or the learning signal is zero. Where the survival lane's competence *ceiling* sits is genuinely unresolved (open question below): a tutor-angel that both warns and acts leaves little room for player-attributable failure.

### Reconciled recommendations (re-ranked under the decisions + sibling report)

1. **Scenario runs, v1 spine** (merges the round-1 director idea, reworked director-lite, with scenario worlds). `promptworld new --scenario first-night`: seeded world + authored incident schedule (deterministic scheduled emissions) + event-derived rubric + morgue epitaph on failure. Each scenario is one prompting lesson in fiction ("survive the first night" teaches visions + orders; "get the curfew repealed" teaches omens + governance reading — TASK-68(b) already sketches both). Builds on: instance manager, world.json presets, TASK-31's run.ended, TASK-68.
2. **Fork-duel as grader and postmortem** (merged with the sibling's fork-compare-as-pedagogy item). TASK-67's fork+compare plus a scored rubric over events (deaths, needs floors, norms passed, charge efficiency, rejected-call rate) — "here's what your prompt change did," side by side. The tightest honest answer to "did my edit work?", enabled by replay determinism.
3. **Faith-driven endogenous mana** — now decided, so it graduates from idea to spec work: define faith as event-sourced state (belief provenance from spec 030 already distinguishes omen-origin), regen as a pure reducer function, and the prophecy-verification rule (what makes a vision "true"). This is the ambient endgame's unscored score.
4. **Registry-grounded tutor surface** (merges this branch's "angel's report card" with the sibling's explain-tool + tutor-charter recommendations — one work item, not three). The `explain` tool grounds mechanics; a default `skills/guide.md` makes the base angel a competent tutor; the report-card critique ("your charter never mentions coordinates; the miracle was rejected twice for them") rides the same grounded surface plus the TASK-63 trace, on a cheap chain.
5. **Fiction-named, artifact-gated capability ladder.** Stages as identities (e.g. Messenger → Guardian → Archangel), unlocked by evidence in the event log, with TASK-81's canonization as the capstone tool. The Cogmind rebrand + this repo's own gate doctrine applied to the player.
6. **Live storyteller-director** — post-v1 graduation of #1: watchers over player state (stage, charge efficiency, failure history) scheduling teachable incidents through the injection door, Cassandra-style cooldowns, persona-named directors as the difficulty dial.
7. **Boatmurdered export** — one command renders chronicle + morgue + charter history as a shareable HTML story artifact. Cheap, and it is the retention/social loop the DF sources say matters most.
8. **Prompt-injection trickster** — a bundle persona (TASK-85) that tries to socially engineer the angel into wasting charges; teaches injection awareness in-fiction (TASK-68 names the concept).
9. **Appointment loop, ethically** — "while you were away" ritual + daily forecast on the ambient world; healthy compulsion rooted in the 24/7 clock, not manufactured FOMO. Complements (does not duplicate) the sibling's first-occurrence learning-helper projection, which is TUI-level teaching, not retention.

### Proposed backlog moves

- **TASK-68**: promote to High — it is the spine. Absorb: scenario definitions with event-derived pass signals (decision 3), fiction-named stage identities, artifact-gated unlocks, and the sibling branch's ACs (per-stage quickstart; stage-1 in-game orientation; pass signal surfaced in-game).
- **TASK-67**: add the scored rubric + postmortem framing to scope; reframe from "teaching feature" to core game verb.
- **New task**: faith-driven charge regeneration (decision 4) — spec-first; touches reducer doctrine.
- **New task**: scenario incident-schedule machinery (director-lite) — the deterministic scheduled-emission design noted above.
- **TASK-111/112**: encode the three-lane initiative frame; add an AC that charter quality measurably changes autonomous performance on a seeded world (the anti-self-grading guard).
- **TASK-31/30**: relabel as learning-game prerequisites; TASK-31 absorbs the shareable-epitaph framing; TASK-30 must reconcile its ~8h/day break-even against scenario session lengths.
- **TASK-14**: add the learning-game question to the proving run: "did prompt iteration measurably change outcomes, and could the player tell why from in-game surfaces alone?"
- **Vault follow-up**: this corpus has zero sources on learning games proper — no Zachtronics/TIS-100 (puzzle-as-programming-pedagogy), no tutorial/onboarding-curve literature, no healthy-retention design, no roguelike meta-progression. Commission a companion research pass before the TASK-68 spec.

## Tensions & tradeoffs

- **Tutor + survival autonomy vs player-attributable failure.** The one reconciliation that is only partial: an angel that both *teaches* and *acts on its own to prevent deaths* narrows the space in which the player can fail informatively. The three-lane frame contains it structurally, but the survival lane's competence ceiling is a real design choice, not a detail — set too high and scenarios can't be lost; too low and TASK-111's own evidence (Ash starved, Oak froze while charges sat at cap) recurs. Flagged, not papered over.
- **Rubrics can be gamed and can flatten emergence.** Event-derived scores invite optimizing the metric over the village; the hybrid decision (ambient world unscored) is the mitigation, but rubric design should score *outcomes the fiction cares about*, never call counts.
- **Director-lite's determinism story needs one honest design step** (scheduled emissions), despite the "existing machinery" framing.
- **Faith-driven mana adds a failure spiral**: low faith → fewer charges → less ability to rebuild faith. Genre-authentic ([[_grounding]]'s positive feedback loop cuts both ways) and roguelike-appropriate for scenarios, but the ambient endgame may want a floor.

## Confidence & open questions

High confidence in the pattern mapping (every claim above cites shipped systems or the grounding) and in the re-ranking under the operator's four decisions. Moderate confidence in the director-lite machinery sketch — it needs its spec. Remaining open (operator's, deliberately kept open):

1. **What failure state can the player cause, and how is it attributed?** (all-dead run end? angel discredited / faith collapse?) Losing only teaches if it traces to the player's text.
2. **The angel's deliberate-incompetence ceiling**: after TASK-112, what must the angel *never* do well without a good charter? This now gates both this branch's and the sibling branch's recommendations — answer before the TASK-112 spec.
3. **Audience**: self-directed engineers vs classroom students — decides self-serve vs artifact-gated stage unlocks, and whether the export/sharing loop (and deferred TASK-65 identity work) matters for v1.

## Basis

- [[_grounding]] — all sourced claims: RimWorld storyteller mechanics, DF/roguelike failure doctrine, god-game mana economy, Progress Quest retention, Generative Agents cascade, Cogmind difficulty rebrand
- [[Simulation-vs-Director]] — the poles promptworld sits between; the director-lite phasing
- [[Indirect-Control-and-Divine-Intervention]] — the mana-economy template the faith decision adopts
- [[Emergent-Narrative-and-Losing-Is-Fun]] — permadeath, retold artifacts, stakes
- [[Observation-Driven-Play]] — chronicle/idle framing and the Smallville precedent
- [[Difficulty-Dials-and-Dynamic-Depth]] — stage naming, dynamic depth, first-startup choice
- [[Storyteller-Driven-Pacing]] — the incident-generator pattern behind scenarios and the future director
- [[Brief-and-Assumptions]] — the superseded ambient-sim lens this note corrects
- Code-grounded evidence cited in prose: docs/wiki/ (metatron, metatron-miracles, metatron-orders, chronicle, gru, agent-mind, tool-loop, governance, social-fabric, overview) and backlog TASK-11/13/27/30/31/52/59/63/64/66/67/68/78/81/85/111/112, all verified 2026-07-24/25; the sibling player-docs branch's report (not linked per vault isolation)
