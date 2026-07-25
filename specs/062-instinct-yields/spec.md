# Feature Specification: Instinct Yields to Intelligence (reflex/planner arbitration)

**Feature Branch**: `062-instinct-yields`

**Created**: 2026-07-25

**Status**: Draft

**Input**: TASK-103 — Direction A from spike TASK-101 (decisions 2026-07-24,
full evidence on the spike task). Root cause of world-01's forage↔goto_warmth
thrash (Sage 436 flips; 334 within ≤200 ticks): **layer fight** — the reflex's
daytime larder-stocking prep rule fires the moment a planner intent completes
(120-tick idle grace), never checks warmth, and counter-schedules the agent
away from the fire the planner just sent it to. Reframe: the reflex ladder is
**instinct that yields to intelligence** — a safety net, not a scheduler.
Additional scope from the spec-057 survival audit: the day-branch warmth gap
(Gap B) and the night search-for-warmth fallback (Gap A) land here. TASK-106's
research confirms the blast radius: day-4/5 thrash storms were village-wide.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Prep instinct yields (Priority: P1)

The reflex's PREP rungs (larder top-up, opportunistic wood/first-fire prep,
non-urgent refuel) stop counter-scheduling: they do not fire (a) within a
yield window after a planner-sourced intent completed, nor (b) while any need
is in its danger band. Survival rungs (eat, seek/make warmth when cold, sleep
when exhausted) are unaffected — instinct still saves lives; it just stops
micromanaging fed, planner-driven villagers. In a world with no planner
(degraded mode), nothing changes: no planner intents means no yield window,
and prep behaves exactly as today.

**Why this priority**: this is the structural saboteur — world-01's loop was
prep-reflex vs planner, alternating every ~200–320 ticks.

**Independent Test**: complete a planner intent, advance less than the yield
window with the agent idle → no prep intent; advance past it → prep resumes.
Put a need in its danger band → prep suppressed regardless of window.

**Acceptance Scenarios**:

1. **Given** a planner-sourced intent that just completed, **When** the agent
   idles past the reflex grace within the yield window, **Then** no PREP rung
   fires (the agent may still eat/sleep/seek-warmth via survival rungs).
2. **Given** any need in its danger band, **When** the reflex evaluates prep,
   **Then** prep yields (the survival rung for that need decides instead).
3. **Given** a world with no planner activity ever (no-LLM), **When** the
   reflex evaluates, **Then** behavior is identical to today (degraded-mode
   fallback preserved — the reflex must keep bodies alive).
4. **Given** the yield window expired with the agent still idle, **When** the
   reflex evaluates, **Then** prep fires as today (yield decays; instinct
   never starves the idle loop forever).

---

### User Story 2 - Cold villagers act on cold in daytime (Priority: P1)

The reflex's day branch gains a warmth rung: a villager whose warmth is in
the danger band during the day seeks known warmth, refuels/relights a known
dying fire, or builds one (the night branch's existing ladder, applied by
day) — BEFORE any prep rung runs. World-01's "Sage forages while freezing"
shape becomes impossible at the reflex layer.

**Why this priority**: the day branch's missing warmth rule is the named gap
(spike cause 1; audit Gap B) that let prep dispatch cold villagers.

**Independent Test**: daytime, warmth in danger band, known fire nearby →
reflex yields a warmth intent, not forage/wander.

**Acceptance Scenarios**:

1. **Given** daytime and warmth in the danger band and reachable known warmth,
   **When** the reflex evaluates, **Then** it produces a warmth-seeking intent.
2. **Given** the same but no known warmth and wood ≥ build cost, **When** the
   reflex evaluates, **Then** it builds (day mirror of the night rung).
3. **Given** daytime and warmth healthy, **When** the reflex evaluates,
   **Then** the day branch behaves as today (rest → prep → wander).

---

### User Story 3 - Cold night with nothing known: search, don't surrender (Priority: P3)

A cold villager at night who knows no warmth, carries insufficient wood, and
finds nothing to chop searches toward the frontier (the existing hungry-search
shape) instead of lying down to sleep cold — the audit's Gap A. Bounded: the
search is one rung above "sleep where you stand", not an endless night trek;
if no frontier is reachable either, the villager still sleeps (today's
terminal behavior).

**Why this priority**: P3 — a real gap (it is how exposure deaths finish),
but rarer than the day-branch storm and riskier to tune (night wandering);
droppable to a follow-on card if implementation shows it needs its own
evidence loop.

**Acceptance Scenarios**:

1. **Given** cold night, no known warmth, no wood, nothing choppable known,
   and a reachable frontier, **When** the reflex evaluates, **Then** a search
   intent toward the frontier results.
2. **Given** the same but no reachable frontier, **When** the reflex
   evaluates, **Then** sleep (today's behavior — the fallback of the fallback).

---

### User Story 4 - Thrash regression corpus (Priority: P1)

A deterministic sim test reproduces the world-01 loop shape — planner sends
the agent to warmth; on arrival-idle the old reflex counter-schedules a
larder forage; the planner re-issues warmth — and proves the loop no longer
occurs under the new arbitration: after the planner intent completes, the
reflex holds prep, the agent stays put, and needs recover.

**Why this priority**: AC #3 on the board task; the proof that the layer
fight is over, not just re-tuned.

**Acceptance Scenarios**:

1. **Given** the scripted Sage-shape scenario (cold, fed, planner goto_warmth
   completing, larder below stock target), **When** run under the new
   arbitration, **Then** zero prep intents fire within the yield window and
   the warmth trajectory recovers (under old behavior the test demonstrates
   the flip — encode both as one regression).

---

### Edge Cases

- Danger bands: defined once, per need, from existing doctrine constants
  where they exist (hunger threshold, cold threshold, exhaustion threshold);
  any NEW threshold is promoted-dial-ready (single named constant) but NOT a
  tuning.json dial yet (earned, not speculative).
- Yield window length: same discipline — one named constant, dial-ready.
  It must be long enough to cover planner cadence (the planner gets a chance
  to follow up its own intent) and short enough that a dead planner never
  strands prep for more than a beat.
- Plan-sourced and meeting-sourced intents count as "intelligence" for the
  yield window exactly like planner-sourced ones (all non-reflex sources).
- The yield state must be event-sourced (replay-deterministic) — whatever the
  reflex consults must be derivable from the log, snapshot-compatible
  (omitempty), and rebase-taxonomy classified if tick-anchored.
- Reflex-sourced intents do NOT arm the yield window (instinct yielding to
  itself would deadlock prep in no-LLM worlds).
- Survival rungs and the TASK-108 refuel reflex (dying fire) are NOT prep —
  enumerate rung classification explicitly in the plan (the 057 audit's rung
  table is the base).
- Interaction with TASK-104 (needs-conditioned recovery, separate branch): no
  file-level dependency; the yield window makes arrival-idle safe even before
  warm_up exists. Do not implement recovery semantics here.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The reflex ladder's rungs MUST be explicitly classified
  SURVIVAL vs PREP (from the 057 audit's rung table); classification lives in
  code as structure or named grouping, not prose.
- **FR-002**: PREP rungs MUST NOT fire within the yield window after the most
  recent non-reflex-sourced intent completed, and MUST NOT fire while any
  need is in its danger band. SURVIVAL rungs are exempt from both conditions.
- **FR-003**: The yield signal MUST be event-sourced state (derivable from
  the log; snapshot-compatible via omitempty; rebase-classified), armed by
  non-reflex intent completion and never by reflex intents.
- **FR-004**: The reflex day branch MUST handle warmth-in-danger before any
  prep rung, mirroring the night branch's seek → refuel-dying → build ladder.
- **FR-005**: The night branch MUST gain the bounded frontier-search fallback
  (US3) below the existing rungs and above terminal sleep; if US3 is dropped
  mid-implementation it MUST be dropped by explicit runbook amendment, not
  silently.
- **FR-006**: Danger bands and the yield window MUST be named constants in a
  single doctrine home, promoted-dial-ready, NOT added to tuning.json.
- **FR-007**: A no-planner world's reflex behavior MUST be byte-identical to
  today except where danger bands suppress prep (deterministic and testable);
  existing reflex/determinism suites stay green.
- **FR-008**: The Sage-shape regression (US4) MUST encode the old loop and
  the new non-loop in one deterministic test.

### Key Entities

- **Rung classification**: SURVIVAL vs PREP over the decideIntent ladder.
- **Yield window**: the post-intelligence quiet period for prep instinct.
- **Danger band**: per-need threshold below which prep yields and survival
  rungs own the agent.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The Sage-shape regression proves the loop dead: zero prep
  intents inside the yield window post-planner-completion; warmth recovers in
  the scripted scenario.
- **SC-002**: Every US1–US3 acceptance scenario has a deterministic test;
  full suite green; replay determinism suites green.
- **SC-003**: No-LLM degraded mode proven: a planner-free drive of the
  reflex matches today's intents except where a danger band suppresses prep.
- **SC-004**: TASK-122's later full-length re-measure has a defined baseline:
  the flip-count methodology from TASK-106's evidence
  (`docs/design/evidence/task-106/analyze.py`) runs unchanged against a
  post-062 world log.

## Assumptions

- Danger-band values: reuse the sim's existing per-need thresholds where they
  exist (hunger/cold/exhaustion trigger points already gate survival rungs);
  where a "danger" grade doesn't exist, set it at the survival-rung trigger
  (the band where instinct already acts) — plan-phase confirms per need.
- The yield window default: one planner cadence (1800 ticks) is the natural
  starting value — the planner gets one beat to follow up before instinct
  resumes prep; plan-phase may adjust with rationale but it stays dial-ready.
- Arrival-idle loitering (staying warm until recovered) is TASK-104's scope;
  this spec only stops the reflex from dispatching the arrived agent away.
- The spec-057 audit's rung table is current and authoritative for
  classification (it was verified against the post-041 ladder today).
