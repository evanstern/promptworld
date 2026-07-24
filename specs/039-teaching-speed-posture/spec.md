# Feature Specification: Teaching-World Speed Posture (Calibrated Soft Cap)

**Feature Branch**: `039-teaching-speed-posture`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Teaching-world speed posture: calibrated soft cap with
horizon-arithmetic warning (classroom mode). Per decision-6 and TASK-78: teaching-mode
worlds carry a per-world config posture that defaults their speed to the highest
calibrated planner-safe ladder rung, derived from the world's calibration profile —
never hard-coded. SOFT cap: exceeding it succeeds and surfaces the horizon arithmetic.
An uncalibrated teaching world prompts for calibrate. Consumable by TASK-68 stage
presets; non-teaching worlds unchanged. NOT an engine rule (decision-4 stands)."

## Context

A learner's loop is *tweak the charter → speed up → watch the effect*, but the cognition
horizon (decision-4) deterministically suppresses villager planner thoughts at exactly
the speeds a learner reaches for (16–32× on typical hosts). Decision-6 (client,
2026-07-23) resolved this for ambient running: teaching worlds get a **soft speed cap**
pinned to the highest calibrated planner-safe ladder rung — the number calibration
already computes — with **warn-with-override** semantics, so overriding the cap is
itself a lesson about the horizon. This is a *teaching posture protecting feedback
legibility*, never an engine rule: the engine still refuses nothing (decision-4).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Teaching world runs at the fastest honest speed by default (Priority: P1)

An educator (or stage preset) marks a world as a teaching world. From then on, the
world's default speed is the highest ladder rung at which villager planners still reach
the model — computed from the world's own calibration profile for the provider that
actually serves planner thoughts, never from a hard-coded number. On a fast calibrated
host that default is high (e.g. 32×); on a slow one it is lower (e.g. 8×). The learner
watching at default speed always watches a village that is genuinely thinking.

**Why this priority**: this is the core deliverable of decision-6's soft-cap mechanism —
without a calibrated default, a teaching world either crawls needlessly or silently
stops thinking, and both anti-lessons defeat classroom mode.

**Independent Test**: create a world, mark it as teaching, give it a calibration profile
with a known seconds-per-point, and observe that the world's effective default speed is
exactly the highest rung the planner arithmetic allows for that profile — and that it
changes when the profile changes, with no code or config number edited.

**Acceptance Scenarios**:

1. **Given** a teaching world whose calibration profile yields planner-safe speeds up to
   16×, **When** the world starts without an explicit speed choice, **Then** it runs at
   16× (not the generic default, not a hard-coded cap).
2. **Given** the same world re-calibrated on a faster host so planners clear 32×,
   **When** it next starts, **Then** its default is 32× — the posture followed the
   profile with no other change.
3. **Given** a non-teaching world on the same host, **When** it starts, **Then** its
   default speed and every speed behavior are exactly as today (unchanged).

---

### User Story 2 - Exceeding the posture teaches the horizon instead of blocking (Priority: P2)

A learner sets a teaching world's speed above the posture. The change **succeeds** —
the cap is soft — and the response surfaces the horizon arithmetic for each affected
cognition class (e.g. "3pt × 17.0s/pt × 32x = 1632 ticks > budget 1200 — villagers
will stop deep-thinking"), so the learner understands exactly what they traded away.

**Why this priority**: warn-with-override is the decided posture shape; without the
arithmetic surfaced, an override silently produces the reflex-only village the feature
exists to prevent.

**Independent Test**: on a teaching world with a known profile, request a speed above
the posture; verify the speed change applies AND the response carries the per-class
arithmetic naming the suppressed classes with the exact numbers the router uses.

**Acceptance Scenarios**:

1. **Given** a teaching world with posture 16×, **When** the operator sets speed 32×,
   **Then** the speed becomes 32× and the response includes the planner's horizon
   arithmetic (points × seconds-per-point × speed vs budget) and its consequence in
   plain language.
2. **Given** the same world, **When** the operator sets a speed at or below 16×,
   **Then** no posture warning appears.
3. **Given** a speed that suppresses both planner and conversation classes, **When**
   the operator sets it, **Then** the arithmetic for each suppressed watched class is
   surfaced, not just the planner's.
4. **Given** a non-teaching world, **When** the operator sets any valid speed, **Then**
   the response is exactly as today (no posture warning; the pre-existing uncalibrated
   warning behavior is untouched).

---

### User Story 3 - Uncalibrated teaching worlds are told to calibrate, not silently capped (Priority: P2)

A teaching world whose planner-serving provider has never been calibrated would, if the
posture were computed silently, adopt the pessimistic bootstrap estimate and cap far
below what the host can actually afford (e.g. 16× capped on a host that measures safe at
32×). Instead, the world tells the operator the posture cannot yet be honest and prompts
them to run calibration — aligning with the existing uncalibrated-world warning
(spec 035), not duplicating or contradicting it.

**Why this priority**: an over-pessimistic silent cap punishes exactly the first-run
classroom experience; the decision explicitly requires the calibrate prompt (TASK-78
AC#3, TASK-40 alignment).

**Independent Test**: create a teaching world with no calibration profile; verify that
starting it and changing its speed each surface a prompt to calibrate, and that the
posture-based default is presented as provisional/pessimistic rather than silently
adopted as if calibrated.

**Acceptance Scenarios**:

1. **Given** an uncalibrated teaching world, **When** it starts or its speed is set,
   **Then** the operator is prompted to run calibration, with the message identifying
   the estimate as the pessimistic bootstrap.
2. **Given** the operator then calibrates and restarts, **When** the world runs again,
   **Then** the calibrate prompt is gone and the posture reflects the measured profile.
3. **Given** an uncalibrated non-teaching world, **When** speed is raised into
   suppression territory, **Then** it warns exactly as spec 035 already does — no new
   behavior.

---

### User Story 4 - The posture is a per-world fact other features can read (Priority: P3)

Curriculum stage presets (TASK-68) and other surfaces need to know whether a world is a
teaching world and what its current posture speed is. The teaching marker lives in the
world's own configuration; the computed posture rung is queryable at runtime without
re-deriving calibration arithmetic by hand.

**Why this priority**: the posture is the carrier TASK-68's stage presets consume;
without a readable home it cannot be composed, but it has no learner-facing value on
its own.

**Independent Test**: read the world's configuration and runtime status: the teaching
marker is present and true for teaching worlds, absent/false otherwise; the effective
posture rung is visible in the world's status surface.

**Acceptance Scenarios**:

1. **Given** a teaching world, **When** its configuration is inspected, **Then** a
   durable per-world teaching marker is present (no global or hard-coded state).
2. **Given** a running teaching world, **When** status is queried, **Then** the current
   posture rung (and whether it is calibrated or provisional) is reported.
3. **Given** an existing world created before this feature, **When** it is loaded,
   **Then** it loads cleanly as a non-teaching world (backward compatible).

---

### Edge Cases

- **Even 1× is not planner-safe** (pathologically slow host/profile): the posture
  defaults to the lowest ladder rung and the arithmetic is surfaced — the world still
  runs; the engine never refuses a speed (decision-4).
- **Planner-serving provider missing from an otherwise-present profile** (per-provider
  divergence, spec 024): that provider is bootstrap-seeded, so the world is treated as
  uncalibrated for posture purposes (US3 path), even though other providers are
  calibrated.
- **Recalibration while the world exists**: the posture is recomputed from the profile
  at each consumption point (start, speed change, status); no stale stored number can
  disagree with the profile.
- **Speed `max` (uncapped)**: already rejected for worlds with an orchestrator; the
  posture adds nothing to that path and must not weaken it.
- **Teaching marker on a pure-sim world (no LLM)**: legal but inert — no posture
  warning can fire because there is no cognition to suppress; no calibrate prompt loops.
- **Posture equals the generic default**: no warning noise — messages appear only when
  a requested speed exceeds the posture or calibration is missing.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST support a durable per-world teaching posture marker,
  settable when a world is created and toggleable afterwards; worlds without the marker
  behave exactly as today (default off; backward compatible with existing worlds).
- **FR-002**: A teaching world's default speed MUST be the highest ladder rung at which
  the planner class routes to the model, computed from the calibration profile of the
  provider that actually serves planner-class work — recomputed at each point of use,
  never stored as a number or hard-coded (survives per-provider divergence and
  recalibration).
- **FR-003**: Setting a teaching world's speed above the posture MUST succeed (soft
  cap) and MUST surface the horizon arithmetic — points × seconds-per-point × speed vs
  budget, plus a plain-language consequence — for every watched cognition class the
  requested speed suppresses.
- **FR-004**: The arithmetic surfaced MUST be computed by the same rule the cognition
  router uses, so the warning and actual routing can never disagree.
- **FR-005**: When a teaching world's planner-serving provider is uncalibrated, the
  system MUST prompt the operator to run calibration (at world start and on speed
  changes) and MUST identify the posture as provisional/pessimistic rather than
  silently adopting the bootstrap-derived cap as if measured.
- **FR-006**: The teaching marker and the effective posture rung (with its
  calibrated-vs-provisional provenance) MUST be readable by other features and by the
  operator: the marker from the world's configuration, the effective posture from the
  world's status surface.
- **FR-007**: All posture messaging MUST be non-blocking and advisory: no speed change
  that is valid today becomes an error, the engine's routing/suppression behavior is
  byte-identical for teaching and non-teaching worlds at the same speed, and existing
  validation (including the `max`-speed rejection for LLM worlds) is unchanged
  (decision-4: the engine never caps speed to protect cognition).
- **FR-008**: Non-teaching worlds MUST be observably unchanged: no new fields required
  in their configuration, no new warnings, no default-speed change.

### Key Entities

- **Teaching posture marker**: a per-world configuration fact ("this world is a
  teaching world"); the only durable state this feature adds.
- **Posture rung**: the derived, never-stored value — the highest planner-safe ladder
  speed for this world right now, with provenance (calibrated vs provisional
  bootstrap).
- **Horizon arithmetic message**: the advisory surfaced on override — per suppressed
  class: points, seconds-per-point, speed, predicted drift ticks, budget ticks, and the
  behavioral consequence.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On a calibrated host, a teaching world at its default speed never has
  planner-class thoughts suppressed by the horizon — 100% of planner routing decisions
  at default speed are allowed, across any calibration profile.
- **SC-002**: 100% of speed changes that exceed the posture on a teaching world both
  apply the requested speed and surface per-class arithmetic; 0% of speed changes at or
  below the posture produce posture warnings.
- **SC-003**: An operator of an uncalibrated teaching world encounters the calibrate
  prompt on first start and on every speed change until calibration exists — and never
  after it does.
- **SC-004**: Worlds without the teaching marker show zero behavior difference from
  before the feature (default speed, warnings, status output for non-teaching worlds
  unchanged).
- **SC-005**: After recalibration, the very next posture consumption reflects the new
  profile with no world-configuration edits.

## Assumptions

- The teaching marker is operator-set (world creation flag or config edit); automatic
  stage-based setting is TASK-68's job, which consumes this posture rather than being
  built here.
- "Planner-safe" defines the posture rung (the planner class is the pain point per
  decision-6); the override warning covers all watched classes, but conversation-class
  suppression does not lower the default rung.
- The always-on, in-TUI legibility of live suppression is TASK-41's scope (spec 037);
  this feature's messaging is the one-shot advisory at speed-change/start time, in the
  same channel spec 035's uncalibrated warning already uses.
- Ladder rungs are the existing capped ladder; the posture picks among them and
  introduces no new speeds.
- Existing worlds on disk predate the marker and must load unchanged (format
  compatibility preserved).
