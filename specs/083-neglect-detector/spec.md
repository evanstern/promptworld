# Feature Specification: Neglect detector — critical need with zero intents in its class

**Feature Branch**: `083-neglect-detector` (task branch: `task-133-neglect-detector`)

**Created**: 2026-07-26

**Status**: Draft

**Input**: TASK-133 / TASK-106 research (`docs/design/thrash-detection-research.md` §1.3,
§3, §4; operator decision 2026-07-25). Reorient 2026-07-26 board move 13 relabel: a
learning-game prerequisite — postmortem attribution can't teach unless the sim can name
neglect when it happens; the alert enters through the shipped severity grammar
(chronicle whole-line alert + map overlay), never a new channel.

## Problem (pinned)

Oak died of exposure on world-01 day 7 (04:04) while warmth drained 636→0 over ~6 hours
— and in that whole window the villager emitted only reflex `chop` and planner `wander`
intents: **zero warmth-class intents during the fatal slide**
(`docs/design/thrash-detection-research.md` §1.3 finding 3;
`docs/design/evidence/task-106/summary.json` false_positive_notes[3]). No oscillation
detector can catch this shape — it is not thrash, it is **death-by-neglect**: a need
below its critical band for a long stretch with the agent's mind never once scheduling
anything in that need's class. Today nothing in the sim names this state: the villager
gets no percept ("I am freezing and I have done nothing about it"), the player gets no
alert until `agent.died`, and the postmortem can only say *exposure*, never *neglect* —
so the teaching moment the learning game rests on has no vocabulary for the single
ambient failure shape world-01 actually produced.

The detector design is substantially pre-decided by the research (§1.3 names the
definition, §3 sketches the injection design, §4 orders it above the thrash percept);
this spec encodes it against the current codebase rather than re-inventing it.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The villager perceives its own neglect (Priority: P1)

A villager's warmth has been below the danger band for two game-hours and in that time
it has set no warmth-class intent — it chopped wood and wandered while freezing (Oak's
shape). The sim emits one deterministic `sim.neglect_detected` event for that villager
and need, and in the same batch injects a high-salience observation memory in the
villager's own voice-of-evidence ("I am dangerously cold and I have done nothing to
warm myself for hours."). The memory's salience sits in the near-death interrupt band,
so the generation bump supersedes any in-flight thought and the next planner prompt is
guaranteed to carry the percept — the shipped interrupt machinery IS the research's
"planner beat" (§3), no new nudge mechanism. One injection per episode: the latch
clears only when the need recovers above the band.

**Why this priority**: board AC #2 — the percept is the mechanism; everything else
(validation, alert surfaces) observes it. Without the injection the detector is
telemetry, not a survival intervention.

**Independent Test**: fold a scripted event history (needs slide + chop/wander intent
records) into a state, run the executor sweep at the firing heartbeat, and assert the
event + companion memory emit exactly once; replay the full log from genesis and assert
byte-identical state (the standard live-vs-replay hash idiom).

**Acceptance Scenarios**:

1. **Given** a villager whose warmth entered the danger band at tick S and has stayed
   below it, with no warmth-class intent set since before S, **When** the needs
   heartbeat reaches S + `neglectWindowTicks`, **Then** ONE `sim.neglect_detected`
   fires for (agent, "warmth") with the pre-tick level and the band-entry tick, and a
   companion `agent.memory_added` carries the fixed per-need text at `salNeglect`,
   situated (`PlaceAt`) — both in the same batch, event first.
2. **Given** the same slide but a `goto_warmth` (or `warm_up`/`build_fire`/
   `refuel_fire`) intent set 30 game-minutes ago, **Then** nothing fires — the
   zero-intent clock reset (the class intent proves the mind engaged, whatever the
   outcome).
3. **Given** a fired detection, **When** the need stays critical for another full
   window with no class intent, **Then** nothing more fires (one injection per episode
   — the fired latch holds until recovery).
4. **Given** a fired detection, **When** the need recovers to/above its band and later
   re-enters and completes a fresh window with zero class intents, **Then** the
   detector fires again (a new episode).
5. **Given** the fired event's memory, **Then** the agent's `Generation` bumps
   (salience ≥ `GenerationBumpSalience`) and the memory wins the working window on
   salience+recency — replay reproduces the identical memory, seq, and bump.

---

### User Story 2 - The detector is validated against Oak's death and silent on health (Priority: P1)

A developer (or the pr gate) runs `go test ./internal/sim/`. A recorded-fixture test
derived from the documented world-01 Oak death window (warmth 636→0 over ~6h, only
reflex `chop` + planner `wander` intents, death day 7 04:04) fires the detector well
before the death tick; fixtures derived from the labeled healthy windows (Oak's
productive day-4 shuttling, +723 food/+902 warmth with class intents present; a routine
night dip that recovers) stay silent. An env-guarded probe test (the spec-043 SC-004
`PROMPTWORLD_WORLD01_DB` idiom) replays the machine-local world-01 log itself and
evaluates the detector predicate over the replayed state across Oak's window (fires)
and sampled healthy windows (silent).

**Why this priority**: board AC #1 — the detector's definition is only as good as its
behavior on the evidence that motivated it. The world-01 log is the regression corpus
(research §4).

**Independent Test**: the fixture tests run in CI unconditionally; the probe test skips
without the env var and is recorded as evidence when run.

**Acceptance Scenarios**:

1. **Given** the Oak-shaped fixture (warmth crossing the band, decaying 4/min to 0,
   zero warmth-class intents, chop/wander records accruing), **When** the sweep runs,
   **Then** `sim.neglect_detected` fires at the first heartbeat ≥ band-entry +
   `neglectWindowTicks` — with warmth already 0 but health ≈ 900, roughly five
   game-hours before the death the same trajectory produces (runway to act).
2. **Given** the healthy-shuttling fixture (need dips below band but class intents land
   within every window), **Then** the detector never fires.
3. **Given** the recovery fixture (need dips below band, recovers above it before the
   window completes), **Then** the detector never fires and the anchors reset.
4. **Given** `PROMPTWORLD_WORLD01_DB` pointing at a copy of the world-01 log, **When**
   the probe replays to sampled ticks inside Oak's final ~6h, **Then** the predicate
   holds for (Oak, warmth); at sampled ticks inside labeled healthy windows (and for
   the never-thrashing agents Ash/Hazel), it does not.
5. **Given** the full fixture event log (including the detector's own emitted events),
   **When** replayed from genesis onto a fresh state, **Then** the state hash equals
   the live-driven state's hash (replay determinism, reducer-only writes).

---

### User Story 3 - The player sees the alert through shipped severity channels (Priority: P2)

A player watching the chronicle sees the neglect line rendered whole-line in the alert
role (bold red) — `sim.neglect_detected` joins `agent.died`/`gru.attacked`/
`social.chest_taken`/`norm.violated`/`stranger.took` in `isAlertType`, exactly the
spec-077 `stranger.took` precedent: membership in the existing tier, no new tier, no
new channel. On the map, the neglected villager is already painted with the shipped
needs-critical overlay (`styleAgentCritical`, bold red underline) — by construction,
since the detector's critical bands are the same sim constants the overlay predicate
reads; a test pins that a neglect-firing agent renders critical.

**Why this priority**: the reorient move-13 relabel (card implementation note) — the
alert MUST enter through the shipped severity grammar. It is P2 only because the
mechanism (US1) must exist first; it ships in the same PR.

**Independent Test**: chronicle render test asserts the whole-line alert styling and
the deterministic wording; map render test asserts the critical overlay on a
neglect-state fixture; `TestCatalogSweep` passes with the new type.

**Acceptance Scenarios**:

1. **Given** a `sim.neglect_detected` event in the feed, **When** the chronicle
   renders, **Then** the row renders whole-line in `styleFeedAlert` with the
   deterministic per-need wording (name + peril + inaction), per the amended
   digest-grammar contract table.
2. **Given** the new event type, **Then** `TestCatalogSweep` passes: a `digestRegistry`
   entry, a `catalogFixture` row, and the backticked mention in
   `docs/wiki/event-types.md` all exist (the three synchronized additions the sweep
   enforces).
3. **Given** a villager in the neglect-firing state, **When** the map grid renders,
   **Then** the glyph carries `styleAgentCritical` — no new token, no new glyph, no new
   legend row (the spec-060 overlay already covers the condition; the test makes the
   subsumption explicit).
4. **Given** the branch touches `internal/tui/`, **Then**
   `node scripts/check-tui-design.mjs --changed` passes with
   `docs/design/tui/patterns/chronicle-grammar.md` (alert-tier list + color-roles row),
   `panels/chronicle.md`, and `panels/map.md` (condition-overlay section names neglect)
   re-verified and re-pinned in the same PR.

---

### Edge Cases

- **Asleep villager**: the detector skips sleeping agents at firing time (their
  inaction is sleep; the spec-064 wake ladder owns sleeping emergencies — hunger wake
  at 150, exposure wake at 150). Anchors keep accruing; if the villager wakes still
  critical and idle past the window, the next heartbeat fires. A sleeping villager's
  rest episode self-clears anyway (`restRegenSleep` +4/min lifts rest above 250 within
  minutes).
- **Direct eating (`agent.ate`)**: the reflex eat path emits no intent and so never
  stamps the food class tick — but eating raises food (+40/+80/+100 per unit) above the
  350 band, which resets the episode via the needs arm. A villager whose eating does
  NOT lift food above the band has no food, and a fire that fails to warm is dead —
  in both cases the eventual firing is honest.
- **Health and morale**: excluded. Health has no intent class (it derives from food/
  warmth — serving those IS serving health); morale is not a survival need. The
  detector covers exactly the `recoveryNeeds` closed set {food, warmth, rest}.
- **Kind-parameterized transfers** (`pick_up`/`withdraw` of food): NOT in the food
  class in v1 — `IntentRecord`/`IntentSetPayload` class membership is goal-name-only,
  and counting every pick_up would mask genuine neglect while hauling wood. Accepted
  false-fire mode: a villager hauling food home for two hours while starving gets a
  percept — a nudge, not a punishment; the memory text is still true. Revisit only on
  live evidence (see Assumptions).
- **Day rescue (warmth)**: warmth drifts up +2/min by day, so a night slide that
  survives to dawn recovers above 350 in ≤ ~3h and the episode honestly resets — the
  detector only fires on slides the day cannot rescue in time (Oak died at 04:04).
- **Never-any-class-intent villager** (fresh world, `Neglect` nil / class tick 0): tick
  0 means "never" and satisfies the zero-intent clause — correct: a villager that has
  never once scheduled for a need it has been critical in for a full window is the
  purest neglect. The band anchor still requires a full window below the band, so
  genesis worlds (needs start healthy) never fire spuriously.
- **Detector event vs. same-tick death**: the sweep reads pre-tick state (the
  `recoveryHoldEvents` purity precedent); a villager dying in the same heartbeat batch
  emits `agent.died` from the heartbeat block regardless. Both events land; `run.ended`
  (if any) still closes the batch. No ordering contract between the two is claimed
  beyond batch membership.
- **Ended world**: `stepEvents`' top guard already freezes everything; the detector
  emits nothing after `run.ended` by construction.
- **Pre-083 snapshots and logs**: the new `Neglect` pointer is `omitempty` (the
  Journal/Hail/Map precedent) — a snapshot without it round-trips byte-identically.
  Replaying a pre-083 log on the new binary populates the derived anchors (the
  spec-043 IntentLog precedent: derived-state additions change replayed state versus
  the old binary; live-vs-replay identity on the SAME binary is the invariant).

## Requirements *(mandatory)*

### Functional Requirements

Mapped to the three board ACs: AC #1 ↔ FR-007..009 (US2), AC #2 ↔ FR-001..006 (US1),
AC #3 ↔ the Composition section below (spec-level, no code). US3 ↔ FR-010..013.

- **FR-001**: The sim MUST maintain, per agent, reducer-derived neglect anchors — per
  need in {food, warmth, rest}: the tick the need entered its critical band
  (`agent.needs_changed` arm: set on downward crossing, cleared with the fired latch on
  recovery to/above the band) and the tick a class intent last landed
  (`agent.intent_set` arm: stamped when the goal is in that need's class) — plus a
  per-need one-per-episode fired latch (set by the new `sim.neglect_detected` reducer
  arm). Reducer-only writes; no other code path may touch them (replay determinism).
- **FR-002**: The critical bands MUST be the existing spec-062 danger-band constants —
  `dangerFoodBelow` (350), `dangerWarmthBelow` (350), `dangerRestBelow` (250) — reused,
  not renamed and not re-valued. These are the same constants the reflex survival
  rungs, the recovery preemption, and (via their exported aliases) the map's
  needs-critical overlay already read: one home.
- **FR-003**: The need→goal class dictionary MUST live in `internal/sim/policy.go`
  beside the goal-resolver registry (research §2: "classes must live next to the goal
  registry or they rot"), v1 membership: food = {`forage`, `hunt`, `cook`}, warmth =
  {`goto_warmth`, `warm_up`, `build_fire`, `refuel_fire`}, rest = {`sleep`}. A test
  MUST pin that every class member is a registered resolvable goal (anti-rot).
- **FR-004**: The detector MUST run in `stepEvents` on the per-game-minute needs
  heartbeat (`nextTick%60 == 0`), as a pure function of pre-tick state and the tick,
  and fire for (agent, need) exactly when ALL hold: agent living and awake; pre-tick
  need value below its band; band-entry anchor non-zero and ≥ `neglectWindowTicks` old;
  last class-intent tick zero or ≥ `neglectWindowTicks` old; fired latch clear. It
  emits ONE `sim.neglect_detected` event carrying {agent, need, level, since} plus, in
  the same batch immediately after it, ONE companion `agent.memory_added` (the
  buildFailed/map-corrected companion shape — memories accrete ONLY via
  `agent.memory_added`, `TestMemoriesAccrete`).
- **FR-005**: `neglectWindowTicks` (T = 7200, two game-hours) and `salNeglect` (9) MUST
  be promoted-dial-READY constants — named, single doctrine home in `internal/sim`,
  with rationale comments — and MUST NOT be tuning.json entries (the spec-062/064
  "dials are earned, not speculative" posture; explicit card instruction).
- **FR-006**: The companion memory MUST be situated (`situatedMemoryEvent`: `PlaceAt`,
  `Why` empty, Origin `OriginWitness` — perceiving one's own condition is direct
  perception) with a fixed deterministic per-need text in the villager's
  voice-of-evidence, at `salNeglect` = 9 ≥ `GenerationBumpSalience` — a deliberate
  join of the near-death/exile interrupt band (deviation from the keep-below-9 default
  for texture memories, recorded in research.md R6): the generation bump IS the
  research §3 planner beat, and the one-per-episode latch bounds the interrupt rate.
- **FR-007**: An in-repo recorded-fixture test MUST validate the detector against the
  documented Oak death window shape — warmth 636→0 at the night decay rate with only
  `chop` (reflex) / `wander` (planner) intent records — asserting it fires at
  band-entry + T (before the trajectory's death), fires once, and re-arms only on
  recovery; and against healthy-window fixtures (class intents present; recovery before
  the window completes) asserting silence. The raw world-01 v3 log is NOT in-repo
  (~106 MB, machine-local at `~/.promptworld/worlds/world-01/`); the fixture derived
  from its documented shape is the binding CI validation, stated honestly.
- **FR-008**: An env-guarded probe test (the spec-043 SC-004 `PROMPTWORLD_WORLD01_DB`
  copy-and-replay idiom, `internal/daemon/context_replay_test.go` precedent) MUST
  replay the machine-local world-01 log and evaluate the detector predicate over the
  replayed state: true at sampled ticks inside Oak's death window, false at sampled
  ticks inside labeled healthy episodes and for Ash/Hazel. It skips without the env
  var. The predicate MUST therefore be factored as a small pure function callable
  outside the sweep.
- **FR-009**: Live-vs-replay byte-identity MUST hold with the detector active: driving
  a world to a firing and replaying its log from genesis produce identical state hashes
  (the `governor_replay_test.go` idiom). Every new tick-bearing anchor MUST be added to
  the `rebaseTicks` SHIFT taxonomy (`internal/sim/miracles.go`) — duration anchors,
  non-zero only (the `NeedsAnchorTick`/`IntentRecord.Tick` shape).
- **FR-010**: `sim.neglect_detected` MUST join `isAlertType`
  (`internal/tui/grammar.go`) — the whole-line alert tier, the spec-077 `stranger.took`
  precedent: one case added to the existing switch, no new tier, no new channel, no
  family-table change (`sim` → `familySim` exists).
- **FR-011**: The type MUST satisfy `TestCatalogSweep`: a `digestRegistry` entry
  (`internal/tui/digest.go`) with deterministic per-need wording, a `catalogFixture`
  row (`internal/tui/digest_test.go`), and the backticked mention in
  `docs/wiki/event-types.md`; the digest-grammar contract table
  (`specs/018-chronicle-digest/contracts/digest-grammar.md` §3) MUST be amended with
  the wording (the spec-077 precedent).
- **FR-012**: The map presentation MUST be the shipped needs-critical overlay and
  nothing new: a test pins that an agent in the neglect-firing state renders
  `styleAgentCritical` (subsumption by construction — FR-002's bands are the overlay
  predicate's bands). No new token, glyph, or legend row.
- **FR-013**: `docs/design/tui/patterns/chronicle-grammar.md` (alert-tier enumeration +
  color-roles `alert` row), `panels/chronicle.md`, and `panels/map.md` (condition
  overlays name neglect) MUST be amended, re-verified, and re-pinned in this PR;
  `node scripts/check-tui-design.mjs --changed` MUST pass on the branch.

### Key Entities

- **`NeglectState`** — new per-agent derived struct, one `omitempty` POINTER field on
  `Agent` (`neglect`, the Journal/Hail/Map precedent — pre-083 snapshots round-trip
  byte-identically): per-need band-entry anchor, last-class-intent tick, fired latch.
  Written only by the three reducer arms in FR-001. See data-model.md.
- **`sim.neglect_detected`** — new executor-emitted event: `NeglectDetectedPayload
  {Agent, Need, Level, Since}`. Executor-emitted ⇒ needs NO injection-door whitelist
  entry (`internal/sim/loop.go` doctrine: pure function of state + tick, like
  `charge_regenerated`).
- **Need class dictionary** — `internal/sim/policy.go`, beside the goal-resolver
  registry: goal-name → need-class membership (FR-003).
- **Doctrine constants** — `neglectWindowTicks` (7200), `salNeglect` (9): the spec-083
  const block, dial-ready, not dialed.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001** (board AC #1): the Oak-shaped fixture fires `sim.neglect_detected` for
  (Oak, warmth) at band-entry + 7200 ticks — with ≈5 game-hours of health runway before
  the trajectory's death — and the healthy fixtures produce zero firings; both pinned
  by CI tests in `internal/sim/`.
- **SC-002** (board AC #1, opt-in): the env-guarded world-01 probe, when run, holds for
  Oak's window and is silent on sampled healthy windows and on Ash/Hazel; its run is
  recorded as evidence on the task.
- **SC-003** (board AC #2): the firing batch carries event + companion memory (salience
  9, generation bump); genesis replay reproduces identical state hash and identical
  memory seq — replay-visible by construction.
- **SC-004** (board AC #2, one-per-episode): a full second window without recovery adds
  zero further events; recovery then relapse fires exactly once more.
- **SC-005**: `TestCatalogSweep` and the chronicle/map render tests pass; the chronicle
  row renders whole-line alert; the map fixture renders `styleAgentCritical`.
- **SC-006**: `go test ./...` green; existing snapshot fixtures load byte-identical
  (`omitempty`); `node scripts/check-tui-design.mjs --changed` passes.
- **SC-007**: the pr gate passes end-to-end: wiki notes touching changed sources
  re-pinned in-branch, `docs/player/` regenerated
  (`node .claude/skills/player-docs/scripts/check-freshness.mjs --check` clean),
  `node scripts/check-merge-drift.mjs pr` exits 0; merged with `gh pr merge --merge`.

## Composition with survival watches — TASK-111 / spec 059 (board AC #3, considered)

The neglect detector and the guardian's survival watches are DISJOINT layers that
compose without contact in this spec, and the seam for a future joint is named here:

- **Different thresholds, deliberate ordering**: the three boot-seeded watches fire at
  emergencies — near-death (health < 200), starvation (Food == 0), exposure
  (Warmth == 0) — while neglect fires at the danger band (350/350/250) plus two hours
  of proven inaction. Neglect is therefore UPSTREAM: on Oak's trajectory it fires
  ≈5 game-hours before the exposure watch's Warmth == 0 moment. The villager gets the
  first chance to save itself (the percept + generation bump); the angel's watches
  remain the later, harder backstop. No double-alarm: they are different events on
  different surfaces at different times.
- **What composes today with zero new code**: the `sim.neglect_detected` event is in
  the world log and the chronicle — visible to the player who commands the guardian,
  and present in any event window a guardian turn reads. The watches' own
  `matchSurvival` path (live-only, `agent.needs_changed`-keyed, hysteresis-latched) is
  untouched.
- **The considered follow-on (NOT built here)**: a fourth system-origin watch kind
  (`neglect`) matching `sim.neglect_detected` — unlike the three needs-band watches it
  would be a structural event-type match (the `orderMatches` shape), giving the
  guardian an autonomous survival turn with real runway instead of a deathbed one. It
  inherits spec 059's origin-keyed exemptions wholesale (cap-exempt, non-expiring,
  non-cancellable). Deliberately out of scope: TASK-111 AC #5 (does charter quality
  change autonomous survival performance?) is still open, and a new watch kind is a
  guardian-authority change that must ride its own deliverable. This section satisfies
  board AC #3 as spec consideration; if the operator wants the watch, it is a one-card
  follow-on with this seam as its input.
- **Backstop relationship (TASK-108/103)**: the reflex/arbitration layer attacks the
  CAUSE (why no warmth intent was scheduled); this detector names the SYMPTOM whatever
  the cause. Post-103/104 re-measures may shrink how often it fires; the detector
  remains correct as the honest last-line namer either way (research §4's sequencing
  argument, inverted for neglect: it lands first because it is simpler and its shape
  actually killed a villager).

## Assumptions

- The research doc's detection definition for THRASH (§2: W=4h, K=8, need-progress
  clause) is explicitly NOT this spec — §4 splits neglect out as its own simpler
  detector; only §1.3's definition ("need below critical threshold for T with zero
  intents in that need's class"), §3's injection design, and §6's dial doctrine bind
  here.
- T = 7200 (two game-hours) is a chosen doctrine constant, not a research-derived
  number (the research names "T" abstractly). Rationale: long enough that every reflex
  survival rung (which triggers AT the same bands) has had dozens of staggered
  opportunities to produce a class intent — so a firing proves the whole mind stack
  failed to engage, not a slow beat; short enough that on Oak's real trajectory the
  percept lands with ≈5 game-hours of health runway. Dial-ready if live evidence wants
  a different value.
- Class membership is goal-name-only (the `IntentRecord.Goal` granularity); `eat` is a
  direct event, not an intent goal, and self-resolves via the needs reset (Edge Cases).
  `chop` is deliberately NOT warmth-class — Oak's fatal window was full of reflex chops
  (research §1.3), and wood-in-pouch is not warmth-seeking.
- No village-level aggregation (the research §2 aggregation was for thrash's 6×
  same-tick storms): neglect is per-agent and episode-latched, so the alarm rate is
  bounded at one per agent per episode by construction.
- The world-01 probe targets the machine-local archived log via the existing
  `PROMPTWORLD_WORLD01_DB` env-guard convention; if the archived `world.v3.db`'s
  events cannot fold through the current reducer, the probe targets the migrated
  `world.db` (the spec-043 Sage-window precedent does exactly this) — the death window
  predates the migration cutoff in either file's history.
- The map overlay needs no code because `needsCritical` (`internal/tui/views.go`)
  already paints `styleAgentCritical` from the SAME exported band constants
  (`SurvivalStarvingRearm`/`SurvivalFreezingRearm` = 350, `DangerRestBelow` = 250);
  FR-012's test converts that coincidence into a pinned contract.
- Tier (recorded on the board card): Opus 4.8 — reducer/percept event + high-salience
  memory injection + world-01 log validation; cognition-adjacent.
