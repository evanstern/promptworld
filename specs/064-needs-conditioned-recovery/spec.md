# Feature Specification: Needs-Conditioned Recovery Intents

**Feature Branch**: `064-needs-conditioned-recovery`

**Created**: 2026-07-25

**Status**: Draft

**Input**: TASK-104 — Direction B from spike TASK-101 (decision 2026-07-24):
recovery goals complete on the NEED, not the location. `goto_warmth` is
arrive-and-done; warming up requires LOITERING, but idleness at the fire is
exactly what invited dispatch elsewhere (the arrive→idle→vacuum cycle).
Generalize as parameterized intent arguments — the completion condition rides
the intent — rather than one-off verbs. Additional scope from the spec-057
survival audit: Gap C (sleepers don't wake to cold) is carded here. Spec 062
(merged) stopped the reflex from counter-scheduling; this spec makes the
recovery itself hold the agent until it has actually recovered.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - warm_up completes on warmth, not arrival (Priority: P1)

A villager who needs warmth takes a `warm_up` intent that walks to warmth and
then LOITERS: the intent stays active — the agent visibly warming, not idle —
until the warmth need reaches the intent's completion threshold, then
completes. The planner can pass the threshold as a tool argument
(`until_warmth`); absent, a doctrine default applies. The reflex's warmth
rungs (day and night, spec 062) issue the same needs-conditioned intent, so
reflex-driven recovery also holds at the fire instead of arriving, idling,
and wandering off cold.

**Why this priority**: the core of Direction B — world-01's Sage reached the
fire over and over and left before recovering (warmth 438→516→drain).

**Independent Test**: drive a cold agent through warm_up at a fire: intent
stays active across ticks while warmth climbs, completes exactly when the
threshold is crossed, agent is released.

**Acceptance Scenarios**:

1. **Given** a cold villager and a warm_up intent targeting known warmth,
   **When** it arrives, **Then** the intent remains active (no completion
   event) while warmth is below the threshold.
2. **Given** the same, **When** warmth crosses the threshold, **Then** the
   intent completes (normal completion event; spec 062's yield window arms if
   the source was the planner).
3. **Given** a planner warm_up with an explicit `until_warmth`, **When** it
   lands, **Then** that threshold (clamped to sane bounds) governs completion;
   absent, the doctrine default governs.
4. **Given** a reflex-issued warmth recovery (day or night rung), **When** the
   agent reaches warmth, **Then** it holds until the doctrine default
   threshold — no arrive-idle-wander-away-cold cycle.

---

### User Story 2 - The condition machinery is generic (Priority: P2)

The completion condition is a general mechanism on intents — a need plus a
threshold — not warm_up-private plumbing. The pattern is proven by at least
one analog (rest: sleep-until-rested already conceptually exists — align it,
or a `recover`-style parameterization of an existing rest/food behavior),
demonstrating a second consumer without inventing speculative verbs.

**Why this priority**: AC #1 on the board task demands the pattern, not just
the instance; P2 because warm_up is the evidenced need.

**Acceptance Scenarios**:

1. **Given** the intent model, **When** a second need-conditioned consumer is
   wired (rest analog), **Then** it reuses the same condition fields and
   completion check — no parallel mechanism.
2. **Given** an intent with no condition, **When** it executes, **Then**
   behavior is exactly today's arrive-and-done (the mechanism is opt-in;
   every existing intent is untouched).

---

### User Story 3 - Recovery is interruptible for the right reasons (Priority: P1)

A loitering recovery doesn't become a trap: (a) if the condition becomes
unreachable — the warmth source dies and the agent is no longer gaining —
the intent aborts (a distinct, honest outcome) and the agent re-decides;
(b) survival still preempts: a recovery for one need never holds an agent
through a worse emergency in another (the existing preemption/decision
machinery must get its chance at the cadence it has today); (c) recovery
intents respect the existing staleness/expiry discipline so a stuck
condition cannot pin an agent forever.

**Why this priority**: co-P1 — an un-interruptible loiter re-creates the
death-by-neglect shape (TASK-133) from the other side.

**Acceptance Scenarios**:

1. **Given** a warm_up at a fire that burns out mid-recovery, **When** the
   agent stops gaining warmth, **Then** the intent aborts with a distinct
   outcome and the agent re-decides (reflex or planner).
2. **Given** a recovery in progress and another need crossing into its danger
   band, **When** the sim evaluates, **Then** the existing
   preemption/decision path can interrupt exactly as it can interrupt any
   active intent today (no new immunity).
3. **Given** a recovery whose condition never advances, **When** the existing
   staleness window elapses, **Then** the intent ends by the existing
   mechanism (no infinite loiter).

---

### User Story 4 - Sleepers wake to cold (Priority: P2)

A sleeping villager whose warmth is draining toward the exposure band wakes
(the audit Gap C — Oak's final night: warmth 636→0 while asleep with the wake
gate only watching daybreak-with-rest and hunger emergency). Waking to cold
puts the villager back into the decision ladder, where spec 062's warmth
rungs (and this spec's warm_up) take over.

**Why this priority**: it is the audit gap explicitly carded here, and the
remaining half of the exposure-death story (062 handles the awake half).

**Acceptance Scenarios**:

1. **Given** a sleeping villager whose warmth falls below the cold-emergency
   threshold, **When** the wake gate evaluates, **Then** the villager wakes
   (mirroring the existing hunger-emergency wake).
2. **Given** a sleeping villager cozy at a fire, **When** night passes,
   **Then** sleep is uninterrupted (the wake fires on the band, not on
   "night").

---

### Edge Cases

- Threshold bounds: `until_warmth` clamps to a sane range (above the danger
  band floor, at or below the need cap); clamp-with-notice, consistent with
  the spec-058 posture for planner-facing args.
- Already satisfied on arrival (or at issue): completes immediately —
  condition checked, not assumed.
- Completion threshold vs need cap: a threshold at cap is legal (loiter to
  full); the abort path (US3) covers a fire that can't get it there.
- Replay determinism: the condition rides the intent's event payload; the
  per-tick completion check is a pure function of state — replay-identical.
- Snapshot compat: new intent fields omitempty; pre-064 snapshots and logs
  load and replay unchanged; rebase taxonomy for any tick-anchored addition.
- The reflex issuing warm_up must not change the no-LLM parity contract
  beyond what it fixes: a planner-free world's villagers now hold at fires
  until recovered — this is an intended behavior change (the arrive-idle
  vacuum was the defect), enumerated in tests like 062's suppression cases.
- Sources and 062 interplay: reflex-issued warm_up completions still never
  arm the yield window (source discipline unchanged).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The intent model MUST support an optional completion condition
  (need + threshold); absent condition ≡ today's arrive-and-done for every
  existing intent (byte-compatible events and snapshots, omitempty).
- **FR-002**: A `warm_up` behavior MUST complete on the warmth need crossing
  its threshold, loitering at the warmth source while active; the planner
  tool surface MUST accept `until_warmth` (clamped, defaulted by doctrine).
- **FR-003**: The reflex's day and night warmth-recovery rungs (spec 062)
  MUST issue the needs-conditioned form with the doctrine default.
- **FR-004**: A recovery whose condition stops advancing (source dead) MUST
  abort with a distinct outcome; recoveries MUST remain interruptible by the
  existing preemption and staleness machinery (no new immunity).
- **FR-005**: The condition machinery MUST be generic across needs, proven by
  a second consumer (rest analog) sharing the same fields and check.
- **FR-006**: The sleep wake gate MUST wake on warmth crossing the
  cold-emergency band (the hunger-emergency-wake shape); cozy sleep stays
  uninterrupted.
- **FR-007**: New thresholds (recovery default, wake band if new) MUST be
  named doctrine constants, promoted-dial-ready, NOT tuning.json entries.
- **FR-008**: Replay determinism and the full existing suite MUST stay green;
  no format_version bump.

### Key Entities

- **Completion condition**: optional (need, threshold) on an intent.
- **warm_up**: the warmth recovery intent/verb consuming it (planner + reflex).
- **Recovery abort**: the distinct outcome for condition-unreachable.
- **Cold-emergency wake**: the new wake-gate arm.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Deterministic recover-then-release test: warm_up holds across
  the recovery span and completes on the exact threshold-crossing tick
  (board AC #3).
- **SC-002**: The arrive-idle-vacuum shape is dead end-to-end: extending the
  062 Sage-shape scenario with warm_up shows the agent held at the fire to
  the threshold, then released — zero mid-recovery dispatches (board AC #2).
- **SC-003**: Abort, preemption, and staleness paths each proven by a test;
  no recovery can pin an agent past the existing staleness window.
- **SC-004**: Second-consumer proof (rest analog) shares the mechanism in
  tests; existing no-condition intents byte-identical through the suite.
- **SC-005**: Wake-to-cold: the Oak final-night shape (sleeping, warmth
  draining to the band) wakes the sleeper in a test; cozy-sleep control stays
  asleep.

## Assumptions

- The completion-condition default for warmth recovery: recover to a healthy
  margin above the danger band (plan-phase picks the constant against the
  existing needs scale; the spike's example used 800/1000).
- "Stopped advancing" (abort trigger) is plan-phase precision — simplest
  honest form: no net gain over a short doctrine window while active at the
  target (must not false-abort on the first flat tick).
- The rest analog for US2 is chosen at plan phase from existing behavior
  (sleep already ends on conditions — aligning it to the shared mechanism may
  BE the proof, with no behavior change).
- Gap C's wake band reuses the exposure/cold constants from specs 059/062
  where they fit (one doctrine, one home).
- TUI: an active warm_up renders via existing intent/goal surfaces; no new UI
  is owed in this spec (check-tui-design gate expected no-op unless goal
  labels appear in `internal/tui` string tables — if they do, the same-PR
  design-doc amendment applies).
