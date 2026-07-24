# Feature Specification: Paused Authoring Chain-Completion

**Feature Branch**: `040-paused-chain-completion`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Paused authoring chain-completion: nudge wakes the nudged villager + pause-aware routing (classroom mode) — TASK-77"

## Context

Classroom mode's authoring loop is *pause → edit the charter → ask Metatron to
nudge a villager → watch that villager's one thought land under the new charter
→ resume*. Decision-6 (client, 2026-07-23) found that while paused this
mediated chain already works almost end to end: Metatron chat has no pause gate
(internal/ipc/server.go:334-355), the angel's nudge lands atomically at the
frozen tick as a `metatron.nudged` event plus a dream memory per target
(blessed by decision-4's landing semantics), and the villager remembers it. It
breaks at exactly the last two links:

1. **No wake** — a landed nudge is not on the planner's wake-stimulus list
   (`absorb()`'s arm switch, internal/mind/mind.go:206-228), so the nudged
   villager never thinks about it while the world is frozen.
2. **False arithmetic** — `routeVerdict` (internal/mind/telemetry.go:61-71)
   computes predicted drift at the world's SET speed even while frozen. A world
   paused from 32x still suppresses the planner, even though a thought taken
   while frozen drifts by exactly zero ticks.

This feature is exactly those two fixes — no new mode, no new verbs, no
single-stepping (explicitly deferred by the client). Doctrine door (decision-6,
extending decision-4): pause changes meaning from "the minds are quiet" to "the
world is frozen, but responds to the angel." Villagers stay sealed; influence
stays mediated through Metatron; staleness budgets remain doctrine.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A nudge wakes the nudged villager while paused (Priority: P1)

A classroom learner pauses the world, edits Metatron's charter, and tells the
angel "nudge Aldric." The vision lands at the frozen tick, and Aldric — and only
Aldric — thinks exactly once about it under the new charter, at the frozen tick,
while the world stays paused. The learner watches the thought and its outcome
appear in the decision surfaces without ever resuming.

**Why this priority**: this is the missing link that makes the paused authoring
sandbox exist at all — without the wake, the nudge is a memory nobody reads
until resume, and the learner's loop has no feedback.

**Independent Test**: pause a world, land a vision on one villager via
Metatron, and confirm exactly one planner thought for that villager is
attempted at the frozen tick; confirm a second nudge to the same villager while
still frozen produces no second thought.

**Acceptance Scenarios**:

1. **Given** a paused world and a villager whose planner debounce window is
   open, **When** Metatron lands a vision on that villager, **Then** that
   villager's planner is armed and attempts exactly one thought at the frozen
   tick, with the nudge event recorded as the thought's arming stimulus.
2. **Given** the same paused world after that thought lands, **When** Metatron
   nudges the same villager again without resuming, **Then** no second thought
   occurs — the game-time planner debounce cannot reopen while the clock is
   frozen (the bound is by construction, not by a counter).
3. **Given** a paused world, **When** Metatron lands a night omen on several
   villagers, **Then** each nudged villager gets its own single bounded round;
   villagers not targeted are not armed.
4. **Given** a paused world and a nudged villager, **When** the thought's
   effects land, **Then** they land at the frozen tick at zero staleness and
   are fully recorded as ordinary events (thought, outcome, and any resulting
   intents), exactly as decision-4's blessed catch-up round.

---

### User Story 2 - Paused routing tells the truth (Priority: P1)

The same learner's paused world was set to 32x — a speed at which the planner
class is suppressed on this host. When the nudged villager's thought is routed
while paused, the router recognizes that a frozen world cannot drift: predicted
drift is zero, zero is within every staleness budget, and the thought is
allowed. The recorded verdict arithmetic says the world was paused, so anyone
reading the record later sees why the thought was allowed.

**Why this priority**: co-equal with the wake — without truthful routing, the
wake arms a thought that the router then suppresses using stale arithmetic, and
the loop still ends in silence. It also independently fixes decision-4's
already-blessed catch-up rounds, which today can be falsely suppressed while
paused on worlds set to high speed.

**Independent Test**: pause a world whose set speed suppresses the planner,
route a planner thought, and confirm the verdict is allow with a recorded
arithmetic string that names the paused state; resume and confirm verdicts
revert to the set-speed arithmetic unchanged.

**Acceptance Scenarios**:

1. **Given** a world paused from a speed that suppresses the planner class,
   **When** any thought is routed while paused, **Then** the verdict is allow
   with zero predicted drift, and the recorded arithmetic names the paused
   state rather than showing set-speed drift math.
2. **Given** a paused world, **When** a routed thought's verdict is recorded,
   **Then** replaying the event log reproduces the identical verdict — the
   paused state is derived from recorded clock events, so paused verdicts are
   reproducible arithmetic.
3. **Given** a world paused from uncapped max speed, **When** a thought is
   routed while paused, **Then** the paused rule wins: the verdict is allow
   (a frozen world does not drift, whatever the set speed).

---

### User Story 3 - The running world is untouched (Priority: P2)

A player running a world normally (not paused) sees behavior byte-identical to
today: nudges landed while running do not arm any new planner wake, routing
verdicts use the same set-speed arithmetic as before, and the existing replay
determinism harness stays green on unpaused logs.

**Why this priority**: the feature is scoped to the paused authoring sandbox;
any leak into running behavior would change villager cadence and horizon
doctrine for every world, which decision-6 explicitly does not authorize.

**Independent Test**: run the full existing test suite plus the replay
determinism harness on unpaused scenarios and confirm no behavioral change;
land a nudge while running and confirm the nudged villager's planner cadence is
unchanged from today.

**Acceptance Scenarios**:

1. **Given** a running (unpaused) world, **When** Metatron lands a nudge,
   **Then** the nudged villager's planner is not newly armed by it — the
   villager thinks on today's existing stimuli and cadence only.
2. **Given** a running world at any speed, **When** thoughts are routed,
   **Then** verdicts and their arithmetic strings are byte-identical to
   today's.
3. **Given** the existing replay determinism harness, **When** it runs on this
   feature's build, **Then** it stays green, including on logs that contain
   paused-nudge sessions.

---

### Edge Cases

- **Nudge inside a closed debounce window**: a villager who thought less than
  one debounce interval before the pause is nudged. The memory lands, but no
  thought occurs while frozen — the game-time debounce is the designed bound
  and stays honest (decision-6: "one nudge buys exactly one round" means *at
  most* one). The learner's remedy is to resume briefly or nudge another
  villager; making this legible is TASK-41's surface, not new mechanism here.
- **Nudged villager is asleep or dead**: existing planner semantics hold — dead
  and sleeping villagers do not think (a night omen to a sleeper still lands as
  a memory that surfaces on wake). Villagers stay sealed; no new wake path for
  sleepers.
- **Nudged villager is at a meeting**: existing semantics hold — the arm stays
  pending until the meeting closes; while frozen the meeting cannot close, so
  no thought occurs until resume.
- **A plan already in flight for the nudged villager**: single-flight per agent
  holds; the arm stays pending and does not stack a second concurrent thought.
- **Multi-target omen while paused**: each living, awake target gets its own
  single round; the per-villager debounce bounds each independently.
- **Resume immediately after a frozen-tick thought**: the debounce window
  carries over into running time from the frozen tick — no double-think burst
  on resume.
- **Pause → nudge → resume → pause again**: game time advanced between pauses;
  if the debounce window reopened, a fresh nudge buys a fresh single round —
  consistent with the bound being game-time, not per-pause.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: While the world is paused, a landed Metatron nudge (vision or
  omen) MUST arm the nudged villager's planner at the frozen tick, with the
  nudge's landing event recorded as the arming stimulus (the causality edge on
  the eventual thought record).
- **FR-002**: The wake MUST apply only to the villagers the nudge targets, and
  each armed round MUST remain bounded by the existing game-time planner
  debounce — no counter, no new bounding mechanism; while frozen the debounce
  cannot reopen, so one nudge buys at most one round per villager.
- **FR-003**: The nudge-wake MUST NOT arm anything while the world is running:
  unpaused arming stimuli are exactly today's set.
- **FR-004**: While the world is paused, routing MUST treat predicted drift as
  zero — zero is within every staleness budget, so the verdict is allow for
  every decision class, including at set speeds (or uncapped max) that would
  suppress while running.
- **FR-005**: A verdict computed while paused MUST record an arithmetic string
  that names the paused state (the truth of why drift is zero), in place of the
  set-speed drift arithmetic; verdicts computed while running MUST be
  byte-identical to today's.
- **FR-006**: The paused state consulted by both fixes MUST be derived from the
  recorded event log (the clock's pause/resume events as applied to the mind's
  replica), never from wall-clock or out-of-band state, so replayed verdicts
  and wakes are deterministic.
- **FR-007**: Thoughts attempted at the frozen tick MUST be fully recorded
  through the existing telemetry (thought and outcome records, and any landed
  effects as ordinary events) with zero staleness, and the replay determinism
  harness MUST remain green on logs containing them.
- **FR-008**: Existing pause doctrine boundaries MUST hold: no new operator
  verbs, no single-stepping, no new mode flag, no per-class budget overrides —
  villagers remain sealed and operator influence remains mediated through
  Metatron's existing landing door.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On a paused world, the loop "edit charter → nudge villager →
  watch the villager's thought land under the new charter" completes without
  resuming, every time the nudged villager's debounce window is open — where
  today it completes never.
- **SC-002**: One nudge while paused yields at most one thought per nudged
  villager: repeated nudges to the same villager during one pause never produce
  a second thought.
- **SC-003**: 100% of verdicts recorded while paused show drift zero / allow
  and name the paused state in their recorded arithmetic; 0 planner
  suppressions occur while paused.
- **SC-004**: Replaying an event log containing paused nudge-thought sessions
  reproduces the log byte-identically (existing determinism harness green).
- **SC-005**: Unpaused behavior is byte-identical to today: the existing test
  suite passes unchanged, and no new wake stimuli or verdict changes are
  observable in running-world logs.

## Assumptions

- The mind's replica already observes pause state deterministically via the
  recorded clock pause/resume events (the reducer maintains a paused flag), so
  no new event types are needed.
- Metatron chat remaining available while paused is existing, blessed behavior
  (no pause gate on the chat handler; decision-6 confirms), and is out of scope
  to change.
- The landed-nudge batch (one `metatron.nudged` event plus one dream memory per
  target) is the existing, stable landing shape for both visions and omens;
  this feature consumes it, not changes it.
- Legibility of "why didn't my nudge produce a thought" (closed debounce,
  sleeping villager) is TASK-41's live horizon surface, already shipped —
  this feature adds mechanism only, no new UI.
- The teaching-world soft speed cap from decision-6 is a separate follow-up
  (TASK-68's stage presets), not part of this feature.
- Doctrine tier: this is a doctrine-adjacent behavior change in the mind's
  routing and wake semantics — constitution Principle V assigns implementation
  to the Opus rubric tier, with this spec linked to board TASK-77 before
  implementation begins.
