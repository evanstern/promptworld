# Feature Specification: Live Cognition-Horizon Surface

**Feature Branch**: `037-live-horizon-surface`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Surface the cognition horizon live in the TUI/status (TASK-41). Remaining scope per the board task's drift audit: (1) a live per-class horizon verdict at the current effective speed — e.g. a header/status indicator like 'conversations suppressed at 32x — calibrate or slow down' — so a player at high speed sees WHY the world has gone quiet instead of 'nothing is happening'; (2) suppression counters, so accumulated suppressions are visible rather than living only in raw cog.outcome payloads. Natural home: the TASK-34 dock (status strip or existing tab) plus the 1-second status poll."

## Context

At high speed the cognition horizon (spec 007) deterministically suppresses
decision classes whose predicted drift exceeds their staleness budget: villagers
stop planning, conversations stop happening, meetings degrade to templates. At
32x even the narrator is router-gated, so the world goes visibly quiet — and
today the only trace of WHY lives in raw `cog.outcome` telemetry payloads and
the per-villager decisions sub-view's "didn't think" chain entries. There is no
aggregate, live answer to "what is the model allowed to think about right now,
and how much thought has been skipped?"

This legibility is also a prerequisite for classroom mode (TASK-66/decision-6):
a suppressed planner with no visible verdict reads to a learner as "my prompt
did nothing."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Live per-class verdict at the current speed (Priority: P1)

A player runs their world at 32x and notices villagers have stopped talking and
planning. Glancing at the always-visible status header, they see a compact
indicator that one or more thought classes are currently suppressed. The
detailed surface names each affected class and the remedy in plain language —
e.g. "conversation suppressed at 32x — slow down", or, on an uncalibrated
world, "— calibrate or slow down" — so the silence has a visible cause and a
next step.

**Why this priority**: this is the core defect — "the world went silent with no
explanation." Everything else in this feature elaborates on this verdict.

**Independent Test**: run a world with a slow (or bootstrap-seeded) provider at
32x, attach the TUI, and confirm the header shows a suppression indicator and
the detail surface names the suppressed class(es) with speed and remedy;
drop to 1x and confirm the indicator clears within one status poll.

**Acceptance Scenarios**:

1. **Given** an LLM world whose current estimates suppress `conversation` at
   32x, **When** the player watches at 32x, **Then** the TUI header shows a
   suppression indicator and the detail surface reads out the class, the
   current speed, and a remedy, within one status-poll interval.
2. **Given** the same world, **When** the player slows to a speed where nothing
   is suppressed, **Then** the indicator disappears and the detail surface
   shows all watched classes as thinking, within one status-poll interval.
3. **Given** a governed world (requested 32x, governor shed to 16x), **When**
   the player views the horizon, **Then** the verdict reflects the EFFECTIVE
   speed (16x), consistent with the speed the router actually applies.
4. **Given** a world running at uncapped max speed, **When** the player views
   the horizon, **Then** all watched classes read as suppressed with phrasing
   that makes clear max speed always suppresses model thought.

---

### User Story 2 - Suppression counters (Priority: P2)

The same player wants to know whether suppression is an ongoing drain or a
one-off: the horizon surface shows, per watched class, how many thoughts have
been suppressed since the daemon started, and the counts visibly grow while the
world runs hot.

**Why this priority**: the live verdict says "suppressed right now"; counters
give it weight — "and it has already cost 214 conversations" — turning a state
into a trend the player can act on.

**Independent Test**: run hot at 32x for a minute, confirm the per-class counts
on the horizon surface increase; slow down, confirm they stop increasing but
are not reset.

**Acceptance Scenarios**:

1. **Given** a world suppressing `planner` at high speed, **When** villager
   planning cadences fire, **Then** the planner suppression count on the
   horizon surface increases accordingly.
2. **Given** accumulated counts, **When** the player slows to an unsuppressed
   speed, **Then** counts stop growing but retain their values (they reset only
   with the daemon).
3. **Given** a fresh daemon on a world that has never suppressed, **When** the
   player views the horizon, **Then** counts read zero.

---

### User Story 3 - Headless status parity (Priority: P3)

An operator (or a classroom facilitator scripting around the CLI) runs the
status command against a hot world and sees the same live horizon: per-class
verdict at the current effective speed plus suppression counts — without
attaching the TUI.

**Why this priority**: the daemon already knows; every attached surface should
say. The CLI is the cheapest surface and serves scripting/classroom tooling,
but the TUI is where the "silent world" confusion actually happens.

**Independent Test**: run `promptworld status <world>` against a suppressed
world and confirm the horizon lines appear; run it against a no-LLM world and
confirm output is unchanged from today.

**Acceptance Scenarios**:

1. **Given** an LLM world with a suppressed class, **When** the operator runs
   the status command, **Then** the output includes the per-class horizon
   verdict and suppression counts.
2. **Given** a no-LLM world, **When** the operator runs the status command,
   **Then** the output is unchanged from before this feature.

---

### Edge Cases

- **No-LLM world**: no orchestrator exists — the status payload carries no
  horizon fields (byte-identical to today), the TUI shows no indicator, and no
  counter machinery is constructed (the spec-028 governor precedent).
- **Paused world**: the verdict renders at the configured effective speed — it
  answers "what may think at this speed," which is also what resuming will
  apply. Pausing neither hides nor resets the surface.
- **Watched class with no serving provider** (e.g. its chain has no admissible
  provider): the class is excluded from the verdict rather than guessed at —
  the health badge (spec 034) already covers provider-down legibility.
- **Calibrated vs uncalibrated remedy**: an uncalibrated (bootstrap-seeded)
  class's remedy mentions calibration; a calibrated class's remedy only offers
  slowing down — never telling an already-calibrated player to recalibrate as
  if it were a fix.
- **Estimate adoption mid-run** (spec 031): the verdict follows the live
  estimate automatically on the next poll; no special handling.
- **Narrow terminal (single-pane fallback)**: the header indicator still
  renders; the detailed per-class surface lives where the layout hosts it in
  both modes.
- **Counter overflow**: counts are since-daemon-start integers; realistic runs
  cannot overflow a 64-bit count.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The status response for an LLM world MUST carry a live per-class
  horizon: for each watched decision class, whether the router would suppress
  it at the current effective speed under the current live estimates, with the
  router's arithmetic verdict available for display.
- **FR-002**: The horizon verdict MUST be derived from the same routing
  arithmetic the router itself applies (single implementation — the spec-035
  FR-006 posture): the surfaced verdict can never disagree with what the router
  actually does at that speed.
- **FR-003**: The verdict MUST be evaluated at the EFFECTIVE speed (post-
  governor), not the requested ceiling, and MUST treat uncapped max speed as
  suppressing all watched classes.
- **FR-004**: The system MUST count, per watched decision class, every thought
  the router suppresses (the suppression floors: skip/reflex/template) since
  daemon start, and surface the counts on the status response alongside the
  verdict.
- **FR-005**: The TUI header MUST show a compact suppression indicator whenever
  at least one watched class is suppressed at the current effective speed, and
  MUST NOT show it otherwise. The indicator follows the existing header-badge
  pattern (degraded / llm-health badges).
- **FR-006**: A TUI dock surface MUST render the per-class detail — each
  watched class's verdict at the current speed plus its suppression count — in
  plain language (raw enum strings never reach the screen).
- **FR-007**: The remedy phrasing MUST distinguish an uncalibrated class
  (mention calibration, mirroring the spec-035 set_speed warning) from a
  calibrated one (offer slowing down only).
- **FR-008**: The CLI status command MUST render the horizon (verdict +
  counts) for an LLM world and MUST leave no-LLM world output unchanged.
- **FR-009**: A no-LLM world MUST carry no horizon fields on the wire
  (additive/omitempty — pre-feature status bytes are identical) and construct
  no counting machinery.
- **FR-010**: The surface MUST refresh at the existing status-poll cadence;
  no new polling loops or push channels are introduced for it.

### Key Entities

- **Horizon class status**: one watched decision class's live standing — class
  name, suppressed-or-not at the current effective speed, the router's verdict
  arithmetic (display string), whether its serving provider is calibrated, and
  its suppression count since daemon start.
- **Suppression counter**: monotonic per-class count of router-suppressed
  thoughts, owned daemon-side (reset only by daemon restart), so every attached
  client and the CLI read the same numbers.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A player watching a hot world can name which thought classes are
  suppressed, at what speed, and what to do about it, using only the TUI —
  within one status-poll interval (~1 s) of the condition becoming true, and
  without opening raw telemetry.
- **SC-002**: The surfaced verdict agrees with the router's actual routing
  decision for every watched class at every ladder speed (verifiable by
  comparing surface output against router outcomes in tests).
- **SC-003**: Suppression counts strictly increase while a class is suppressed
  and its cadence fires, and never reset while the daemon runs.
- **SC-004**: A no-LLM world's status output and TUI render are byte-identical
  to pre-feature behavior.
- **SC-005**: A learner in a suppressed-planner world can answer "why did my
  prompt do nothing?" from the visible surface alone (classroom-mode
  prerequisite, decision-6).

## Assumptions

- **Watched classes are the surface's scope**: the fixed spec-035 watched set
  (`planner`, `conversation`, `meeting`) is what players reason about;
  slow-cadence classes (consolidation, chronicle, metatron) are deliberately
  out of scope for the live surface, as they are for the boot warning and
  set_speed warning.
- **Counters are daemon-side and process-lifetime**: parity with the estimator
  (process-lifetime) and with "every client sees the same numbers." Persisting
  counts across restarts is out of scope.
- **The verdict is computed daemon-side and shipped on status**: clients render
  verbatim and never re-derive routing arithmetic (the metatron-orders "no
  client-side re-derivation" posture). The TUI's 1-second status poll is the
  refresh vehicle.
- **Home of the detailed surface**: the dock's existing panes (where the LLM
  provider table already lives) — a new dock tab is not required; the header
  indicator + a per-class block on an existing pane satisfies the task's
  "status strip or existing tab" framing. Final placement is a plan decision.
- **The spec-035 set_speed warning stays as-is**: it fires once at the moment
  of a speed change and is gated to uncalibrated providers; this feature is the
  continuous complement, not a replacement.
- **Dependency**: spec 035's live-estimate suppression arithmetic
  (`SuppressedAt`-family, single-implementation doctrine) and the existing
  status plumbing (1 s poll, additive-omitempty fields, spec 028/034
  precedents).
