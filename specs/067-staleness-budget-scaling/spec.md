# Feature Specification: Staleness budget scaling — planning must survive clock speed

**Feature Branch**: `067-staleness-budget-scaling`

**Created**: 2026-07-26

**Status**: Draft

**Input**: User description: "Plan staleness budget doesn't scale with clock speed — planning silently dies above ~4x on local tiers (TASK-141)."

## Diagnosis (AC#1 — pinned)

The landing gate and the admission gate measure the same hazard with different
rulers, and only one of them knows about clock speed:

- **Landing gate (the killer):** `internal/sim/landing.go:rungStale` rejects an
  intent when `staleness > dc.BudgetTicks`, where staleness =
  `state.Tick - in.SnapshotTick` (actual game-ticks elapsed while the thought
  ran) and `BudgetTicks` is the **fixed** per-class constant from
  `internal/cognition/registry.go` (`planner` = 1200 ticks). The rejection
  surfaces as `OutcomeRejectedStale` (`internal/sim/cognition.go:17`).
- **Admission gate (the honest-looking one):** `internal/cognition/route.go:Route`
  predicts drift as `points × calibrated s/pt × ticksPerSecond` and allows when
  predicted drift ≤ the same `BudgetTicks`. It scales with speed — but it
  predicts from the **calibrated per-call latency**, which excludes endpoint
  queue wait.

**The latency math** (TASK-122 measure world, gemma4:12b-mlx, parallel=4,
8 villagers): a plan thought's true wall time is ~250s once queueing is
included. 1 tick = 1 game second (`internal/clock/clock.go:1`); speed is game
seconds per real second. So the same 250s thought accrues:

| Speed | Actual staleness | vs budget 1200 | Outcome |
|-------|-----------------|----------------|---------|
| 4x    | ~1000 ticks     | ≤ 1200         | lands   |
| 8x    | ~2000 ticks     | > 1200         | dies    |
| 32x   | ~8000 ticks     | > 1200         | dies    |

Meanwhile Route, predicting from calibrated ~12 s/pt (3pt × 12 × 8 = 288 ticks
≤ 1200), keeps **admitting** planner thoughts at 8x, and every horizon/status
surface (`internal/cognition/horizon.go:LiveHorizon` and everything built on
it) reports planner as healthy — the horizon gates *scheduling*, never
*landing*. Result on the measure world: 27 `set_plan` landings vs 287
rejected-stale (~91% of admitted plan thoughts burned model budget and were
discarded), while the TUI said "planner thinking."

## Decision (the fix space, decided)

TASK-141 names three candidate mechanisms. This spec chooses **(a) scale the
landing budget with the event-sourced clock speed**, and rejects the others:

- **(b) derive the landing budget from the calibrated s/pt** — rejected: the
  latency estimator is daemon runtime state (persisted out-of-band, TASK-113),
  **not** event-sourced. A landing gate reading it is not a pure function of
  event-sourced state and breaks replay determinism. (The horizon may read it
  because the horizon only schedules; landing outcomes are recorded events.)
- **(c) promote the budget to a tuning.json dial (spec 048 path)** — rejected
  as the fix: a bigger constant is still a fixed constant; the structural
  death just moves to a higher speed. (Nothing here forecloses later dial
  promotion of the *base* budget.)

**Chosen semantics:** a class's `BudgetTicks` is reinterpreted as its staleness
budget **at 1x** — i.e., wall-clock patience. At landing, the effective budget
is `BudgetTicks × ticksPerSecond(state.Speed)` using the event-sourced
effective speed in force at the landing tick (`internal/sim/state.go:26`,
reduced from `speed.set` events). The admitted thought that lands when the
router predicted it would always lands; a 250s thought passes at every capped
speed (250s wall ≤ 1200s wall patience).

**What deliberately does not change:** `Route` and the cognition-horizon
doctrine keep the fixed fiction budgets. Scheduling still paces cognition
against the fiction (the horizon); landing forgives delivery against the wall.
Scaling `Route` identically would cancel speed out of the admission arithmetic
and dissolve the horizon doctrine (specs 007/032/035/037) — explicitly out of
scope.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Planning survives 8x on a calibrated local tier (Priority: P1)

An operator runs a calibrated local-tier world (the TASK-122 regime:
gemma4:12b-mlx all-routes, parallel=4, 8 villagers) at 8x. Villagers keep
planning: admitted plan thoughts land instead of being silently discarded
after burning model budget.

**Why this priority**: This is the defect itself — above ~4x, planning is
structurally dead on local tiers while every surface says it is healthy. 8x is
a supported ladder rung, not an exotic posture.

**Independent Test**: A landing-gate unit test drives `rungStale` (or its
successor) with the measured regime's numbers at 4x/8x/32x; a measured run
against the measure-world profile confirms the landing rate.

**Acceptance Scenarios**:

1. **Given** effective speed 8x and a plan thought whose actual staleness is
   2000 ticks (the measured 250s regime), **When** the intent lands, **Then**
   it is accepted (2000 ≤ 1200 × 8).
2. **Given** effective speed 1x and a thought 1300 ticks stale, **When** it
   lands, **Then** it is rejected exactly as today (1300 > 1200 × 1) — 1x
   behavior is unchanged.
3. **Given** effective speed 8x and a thought whose wall time exceeded the
   wall patience (staleness > 9600 ticks), **When** it lands, **Then** it is
   rejected `rejected-stale` with the **effective** (scaled) budget named in
   the reason string.

---

### User Story 2 - Replay determinism across speed changes (Priority: P1)

A recorded event log spanning mid-run speed changes replays to the identical
state: every landing outcome reproduces, because the gate reads only
event-sourced state (tick, snapshot tick, effective speed) and intent fields.

**Why this priority**: Landing outcomes are recorded events; a
non-deterministic gate corrupts every replay-derived surface. This is the
constraint that killed option (b).

**Independent Test**: Extend the existing replay-determinism coverage (e.g.
alongside `internal/sim/governor_replay_test.go`) with a scenario that lands
thoughts across a `speed.set` boundary and asserts identical outcomes on
replay.

**Acceptance Scenarios**:

1. **Given** an event log where a thought is admitted at 4x, the speed changes
   to 16x mid-flight, and the intent lands, **When** the log replays, **Then**
   the landing outcome is byte-identical (the gate used the landing-tick
   effective speed both times).

---

### User Story 3 - Status surfaces stop calling structural death "thinking" (Priority: P2)

When a class's landings are structurally dying, the operator-facing surface
says so instead of reporting healthy cognition (TASK-141 AC#3).

**Why this priority**: With US1 fixed, the observed 8x gap closes — landings
succeed, so "planner thinking" becomes true again. What remains is the
*residual* gap: any future regime where actual latency exceeds even the scaled
budget (wall latency > `BudgetTicks` seconds, ≈20 wall-minutes for planner)
would again die silently. The AC allows fixing the surface **or** explicitly
documenting the gap; this spec requires the documentation plus the cheap
honest signal already carried by telemetry.

**Independent Test**: Documentation review + a test that the rejection
telemetry (`cog.outcome`) carries the effective budget, so the decision-trace
and systems surfaces render the true arithmetic.

**Acceptance Scenarios**:

1. **Given** a rejected-stale landing at 8x, **When** the `cog.outcome` /
   `agent.intent_rejected` events are rendered (decision trace, systems tab),
   **Then** the reason names the scaled budget (e.g. `staleness 9800 > budget
   9600 (1200 × 8x)`), never the misleading base constant.
2. **Given** the shipped docs/wiki notes for the cognition horizon, **When**
   the reader consults them, **Then** the horizon-gates-scheduling /
   landing-gates-delivery split and the residual structural-death regime are
   explicitly documented.

---

### Edge Cases

- **Speed change mid-flight**: the gate uses the effective speed at the
  *landing* tick, not the admission tick. A drop 8x→1x mid-flight can shrink
  the budget under an in-flight thought and reject it — deterministic and
  accepted (the operator slowed the world; fiction-staleness tolerance
  tightened with it).
- **Paused world**: ticks freeze, so staleness stops accruing; no special
  case needed (RoutePaused already admits everything).
- **Uncapped max speed**: `TicksPerSecond() == 0`. Route already suppresses
  every class at uncapped speed, so no admitted thought should reach landing;
  the gate MUST NOT multiply by zero — at uncapped speed the base budget
  applies unscaled (theoretical branch, mirrors Route's posture).
- **Governor-held speed**: `state.Speed` is the *effective* speed (the
  governor holds it below `RequestedSpeed`); scaling reads the effective
  speed, which is the one the fiction actually ran at.
- **Negative staleness**: already clamped to 0 (`landing.go:42-44`); scaling
  changes nothing.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The landing staleness gate MUST compare actual staleness against
  an effective budget of `BudgetTicks × ticksPerSecond(effective speed at the
  landing tick)`, where `BudgetTicks` is the class's registered 1x budget.
  At uncapped speed (ticksPerSecond ≤ 0) the base budget applies unscaled.
- **FR-002**: The gate MUST remain a pure function of event-sourced state and
  the landing intent's recorded fields — no wall-clock reads, no latency-
  estimator reads, no daemon runtime state.
- **FR-003**: The admission router (`Route`), `RoutePaused`, and every horizon
  surface built on them MUST be left unchanged: fixed fiction budgets keep
  gating scheduling. Only the landing consumption site changes.
- **FR-004**: Rejection telemetry (`agent.intent_rejected` reason and
  `cog.outcome` reason) MUST name the effective budget and its derivation
  (base × speed), so recorded audits never show the misleading base constant.
- **FR-005**: The doctrine comment on `DecisionClass`
  (`internal/cognition/registry.go`, decision-4) MUST be updated to state the
  new interpretation: `BudgetTicks` is the staleness budget at 1x (wall-clock
  patience), enforced scaled at landing; values remain reviewed-code doctrine,
  not runtime tuning.
- **FR-006**: The residual structural-death regime (wall latency exceeding
  `BudgetTicks` seconds still dies at every speed while the horizon reports
  healthy) MUST be explicitly documented in the cognition-horizon wiki note(s)
  re-pinned by this task.
- **FR-007**: Existing landing behavior at 1x MUST be bit-identical to today
  (1x scale factor is 1.0); no recorded historical log's replay may change
  outcome. (New scaled behavior only manifests in events recorded after this
  change; replay of old logs re-runs the same pure gate over the same state
  and reaches the same outcomes because their speeds and stalenesses are
  unchanged — the gate change itself MUST be verified against a pre-change
  fixture log if one exists.)

### Key Entities

- **DecisionClass.BudgetTicks**: per-class staleness budget, reinterpreted as
  the 1x (wall-clock) budget; registry values unchanged.
- **Landing intent**: carries `SnapshotTick` (staleness numerator) and class;
  unchanged shape.
- **Effective speed** (`sim.State.Speed`): event-sourced via `speed.set`
  reductions; supplies the scale factor at the landing tick.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In the measured 8x local-tier regime (250s plan thoughts ≈ 2000
  ticks staleness), admitted `set_plan` thoughts land instead of dying: unit
  test proves acceptance at 8x with the measured numbers, and a measured run
  against the TASK-122 profile shows the rejected-stale share of plan landings
  drops from ~91% to under 10%.
- **SC-002**: Landing outcomes are replay-deterministic across speed changes:
  the replay test of US2 passes, and the full existing replay suite stays
  green.
- **SC-003**: 1x landing behavior is unchanged: existing `rungStale` /
  landing tests pass without weakening any 1x assertion.
- **SC-004**: Every rejected-stale event recorded after the change names the
  scaled budget in its reason; the decision-trace view renders it without
  modification (string-only change) or with its renderer updated in the same
  task.
- **SC-005**: The cognition-horizon wiki note documents the
  scheduling-vs-delivery gate split and the residual structural-death regime
  (FR-006), and the wiki freshness gate passes.

## Assumptions

- The TASK-122 measure world (`~/.promptworld/measure/task-122`) remains
  available as the reference regime for the measured-run half of SC-001; if
  rerunning it is impractical, the unit test with its recorded numbers plus
  the landing-rate arithmetic satisfies AC#2's "proven by test" arm.
- `state.Speed` at the landing tick is the correct scale source (effective,
  governor-held speed) — confirmed present and event-sourced at
  `internal/sim/state.go:26`.
- Queue-aware admission prediction (making `Route`'s s/pt include endpoint
  queue wait) is out of scope — that is estimator territory (TASK-86 lineage)
  and touches scheduling doctrine, not landing.
- No migration is needed: no persisted artifact stores an effective budget;
  budgets are recomputed from the registry constant and event-sourced speed.
