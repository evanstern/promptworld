# promptworld as a learning game — synthesis (2026-07-25)

**Status:** design synthesis; plan-of-record candidate pending board sign-off.
**Inputs:** `research/Game-Gameplay-Patterns/Analysis-Learning-Game-Fit.md` and
`research/Game-Player-Docs/Analysis-In-Game-First-Teaching.md` (both gate-verified
2026-07-25), cross-grounded against each other under eight operator decisions recorded
below. This document is the cross-branch connective tissue the vault's isolation rule
forbids the analyses from holding themselves.

## TL;DR

promptworld already owns the hard part: a control surface that literally *is* assistant
configuration (charter / skills / capability manifests — TASK-64), a persisted
per-tool-call verdict trail as the feedback signal (TASK-52/63), a budgeted god-game
intervention economy (metatron-miracles), and production proof that one prompt can move a
world (the Thornspire cascade: 271 events from one omen, TASK-81). What it lacks is the
game around that instrument — goals, player-attributable failure, paced lessons, a grade —
and the docs to receive a stranger. **No massive refactor is necessary or advisable**: the
event-sourced substrate is the asset, and every high-payoff move below lands as contained
design steps on existing seams (one reducer-doctrine change, one scheduled-emission
primitive, one initiative-frame clarification). The single load-bearing risk is the angel:
as it gains autonomy (TASK-111/112) *and* the tutor role, the player's authored text must
remain the binding variable — otherwise the game grades itself and the lesson evaporates.

## The eight ratified decisions (operator, 2026-07-25)

1. **Session shape: staged.** Scenario runs teach the early curriculum; the ambient world
   is the endgame you graduate into.
2. **Director-lite first.** V1 scenarios carry pre-scripted incident schedules
   (deterministic scheduled emissions); a live state-watching storyteller is post-v1.
3. **Scoring: hybrid.** Scenarios and fork-duels get event-derived rubrics and pass
   signals; the ambient world stays unscored — the chronicle is its only mirror.
4. **Mana: faith-driven endogenous regen.** Charge regeneration becomes a pure reducer
   function of village faith; landed true visions build it, failed prophecies decay it.
5. **The angel is the tutor**, grounded by a read-only registry-derived `explain` tool so
   mechanics facts come from data, never model vibes.
6. **Screen-orientation page + keys card ship now** as a content task (approved; carded).
7. **Failure tone: unsoftened everywhere, reassured up front** — deaths stay real in all
   modes; `getting-started.html` gains the DF-style "your first village will probably
   freeze — that's the story" paragraph.
8. **Onboarding is every-world, TUI-level** — `?` overlay + auto-retiring
   first-occurrence lessons run everywhere; curriculum stages remain a separate opt-in
   layer.

## What we like (keep, and build on)

- **The control surface is the curriculum.** Charter → skills → capability manifests →
  Starlark tool bundles (TASK-85) is already the prompting → instruction-files → agents →
  tools ladder; the depth axis and the lesson plan are the same object.
- **"Make help unnecessary first" is shipped**: digest grammar with sweep-test coverage
  (TASK-60), verdict glossary (TASK-63), suppression rows carrying their own remedy
  (TASK-40/41), the teaching-speed posture that prints horizon arithmetic (TASK-78).
- **Small prompt, big consequence is proven in production** (Thornspire), and the fork
  machinery (TASK-67) plus replay determinism make the comparison honest.
- **Docs honesty machinery exists on both sides**: `docs/player/` freshness gating
  (TASK-82) and described ≡ declared tool guidance derived from the registry (TASK-64).
- **Failure is already permanent** (`agent.died`); it just isn't reachable (gru wounds,
  never kills — zero deaths in 1257 game days). Stakes are a tuning decision away, not a
  systems rebuild.

## The unified design position: three lanes, split by channel

The two branches' angel tensions resolve into one frame:

- **Tutor lane** — always-on, charge-free, faith-free, ungraded. Rides the existing
  converse (speech) channel plus one read-only `explain` tool grant; excluded from every
  rubric. Explaining is speech, not an act: **no initiative-frame relaxation needed.**
- **Survival lane** — autonomous per TASK-111 (genesis watch orders), still charge-gated.
  Its competence *ceiling* is the top open question below.
- **Ambition lane** — world-shaping, clock control, standing orders: player-requested or
  pre-authorized, exactly the current contract.

The anti-self-grading guard (AC for TASK-111/112): **charter quality must measurably
change autonomous performance on a seeded world.** A default-charter angel should
demonstrably underperform an authored one, or the learning signal is zero. Deliberate
incompetence, if adopted, applies to world-acting only — never the tutor voice or the
`explain` tool's facts.

## Course of action (build-ordered)

**Wave 0 — ship now (content only, zero design risk)**
- Screen-orientation player page (NetHack-chapter-3-shaped: screen regions, glyph table)
  + keys reference card + losing-is-fun paragraph in `getting-started.html`. Consider a
  registry-generated reference section so glyph/cost/key tables cannot rot (CDDA pattern).

**Wave 1 — the grounded feedback layer + the floor (v1 teaching substrate)**
- `explain` tool (deterministic, registry-derived, grant-gated, read-only) + default
  `skills/guide.md` + tutor-charter preset. The report-card critique ("your charter never
  mentions coordinates; the miracle was rejected twice for them") rides the same grounding
  + the TASK-63 trace on a cheap chain: explain is pull, the report card is push, one data
  source so the grader never grades on vibes.
- `?` overlay in the TUI (every world). Load-bearing: a no-LLM world has no tutor, so the
  overlay is the charter-independent floor, not redundancy.
- Stakes prerequisites: TASK-31 (run outcomes, morgue, gru lethality) and TASK-30 (labor
  budget) — the rubric means nothing until failure is reachable.

**Wave 2 — the game (v1 spine)**
- Scenario runs: `promptworld new --scenario first-night` — seeded world + authored
  incident schedule + event-derived rubric + morgue epitaph on failure. Each scenario is
  one prompting lesson in fiction. Requires the one honest design step: a deterministic
  scheduled-emission primitive (pure function of (state, tick), the charge-regen/order-
  expiry precedent) — small, but real; spec it eyes-open.
- Fork-duel as grader and postmortem (TASK-67): charter A vs B on a seeded fork, rubric
  over events (deaths, needs floors, norms passed, charge efficiency, rejected-call rate),
  rendered plain-language (glossary discipline — no raw enums in a grade).
- First-occurrence lessons projection (every world, auto-retiring; scenario incidents
  double as scheduled lesson triggers). Every pushed lesson also reachable from `?`.
- TASK-68 curriculum ladder as the spine tying these together: fiction-named stages as
  identities (Cogmind's rebrand lesson), artifact-gated unlocks from event-log evidence,
  per-stage quickstart pages, pass signals surfaced in-game.

**Wave 3 — the endgame arc**
- Faith-driven charge regen (spec-first; touches reducer doctrine). Failure-spiral
  tradeoff is genre-authentic for scenarios; the ambient world may want a floor.
- Capability ladder capstone: TASK-81 canonization as the top-stage tool.
- Boatmurdered export: chronicle + morgue + charter revision history as one shareable
  HTML retelling — the retention/social loop the DF sources say matters most.
- Appointment loop, ethically: "while you were away" ritual + daily forecast.

**Post-v1**
- Live storyteller-director (watchers over player state scheduling teachable incidents;
  persona-named directors as the difficulty dial).
- Prompt-injection trickster persona (TASK-85 bundle) — injection awareness in-fiction.

**Research follow-up (before the TASK-68 spec)**
- Companion vault pass on learning-game design proper: Zachtronics/TIS-100
  puzzle-as-pedagogy, tutorial/onboarding-curve literature, healthy-retention design,
  roguelike meta-progression, RimWorld learning-helper per-lesson anatomy.

## On "massive refactors"

Explicitly assessed and rejected. The candidates examined: a director subsystem (deferred
to post-v1 by decision 2; v1 rides scheduled emissions), a scoring engine (rubrics are
event-log projections — the decision-trace precedent), the tutor (charter/skill content +
one tool grant, no doctrine change), faith mana (a reducer function over event-sourced
state, the same shape as charge regen today). The substrate's event-sourced,
replay-deterministic discipline is precisely what makes every one of these cheap. The
expensive path would be abandoning the ambient identity for a pure puzzle game — the
staged decision keeps both and sequences them instead.

## Proposed board moves (pending sign-off; item 6 of the decisions already approved)

| Move | What |
|---|---|
| TASK-68 → High | The spine. Absorb: event-derived pass signals surfaced in-game, fiction-named stage identities, artifact-gated unlocks, per-stage quickstarts, stage-1 orientation via tutor charter. |
| TASK-67 rescope | Core game verb: scored rubric + postmortem framing, plain-language output. |
| TASK-111/112 | Encode the three-lane frame + channel split; AC: charter quality measurably changes autonomous performance. Frame 112 as "the player programs an agent" (stage 3). |
| TASK-30/31 relabel | Learning-game prerequisites; 31 absorbs shareable-epitaph framing; 30 reconciles break-even vs scenario session lengths. |
| TASK-14 extend | Add: "did prompt iteration measurably change outcomes, and could the player tell why from in-game surfaces alone?" |
| NEW (carded) | Screen-orientation page + keys card + losing-is-fun paragraph. |
| NEW | Grounded feedback layer: `explain` tool + guide skill + tutor-charter preset + report card (shared grounding contract). |
| NEW | `?` overlay (every world, no-LLM worlds included). |
| NEW | First-occurrence lessons projection. |
| NEW | Faith-driven charge regeneration (spec-first). |
| NEW | Scenario incident-schedule machinery (director-lite, scheduled-emission primitive). |
| NEW | Vault research pass: learning-game design sources. |

## Open questions (genuinely the operator's)

1. **Player-attributable failure state** — what loss can the player cause, and how is it
   attributed (all-dead run end? faith collapse / angel discredited)? Bounds what the
   report card may honestly attribute; until answered it attributes to charter text only,
   never blame.
2. **The angel's deliberate-incompetence ceiling** — after TASK-112, what must the angel
   *never* do well without a good charter? Gates the TASK-112 spec; the channel-split
   position (incompetence applies to world-acting only) needs ratifying as doctrine.
3. **Audience** — self-directed engineers vs classroom students. Decides self-serve vs
   gated stage unlocks, whether the export/sharing loop (and TASK-65) matters for v1, and
   whether `?`'s advanced tier exposes raw registry values.
4. *(minor)* Seen-lessons persistence home (per-user file vs per-world client state) and
   reset semantics.
