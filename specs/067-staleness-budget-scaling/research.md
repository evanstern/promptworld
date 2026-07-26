# Research: Staleness budget scaling (spec 067)

## R1 — Scale source: event-sourced effective speed at the landing tick

- **Decision**: The scale factor is `state.Speed.TicksPerSecond()` read from the
  reducer's own state at the landing tick (`internal/sim/state.go:26`,
  `internal/clock/clock.go:100`). Effective budget =
  `BudgetTicks × ticksPerSecond`, with `ticksPerSecond ≤ 0` (uncapped `max`)
  falling back to the unscaled base budget.
- **Rationale**: `Speed` reduces from recorded `speed.set` events, so the gate
  stays a pure function of event-sourced state (spec FR-002). 1x has
  `TicksPerSecond() == 1`, so the 1x behavior is bit-identical (FR-007). The
  ladder values (1, 4, 8, 16, 32) are small exact floats; `float64(BudgetTicks)
  × tps` is exact for every registry value, so the int64 conversion is
  deterministic across platforms.
- **Alternatives considered**:
  - *Requested speed* — wrong: the governor can hold effective speed below the
    request; the fiction ran at the effective rate.
  - *Speed at admission (snapshot tick)* — requires carrying admission speed in
    the intent; landing-tick speed is simpler, already available, and the
    mid-flight-change semantics (operator slows world → tolerance tightens) are
    acceptable and deterministic (spec Edge Cases).
  - *Calibrated s/pt* — rejected in the spec (estimator is not event-sourced;
    replay hazard).

## R2 — Consumption-site map for `BudgetTicks` (the TASK-141 "first task")

| # | Site | Role | Disposition |
|---|------|------|-------------|
| 1 | `internal/cognition/route.go:32` (`Route`), `route.go:50` (`RoutePaused`) | Admission: predicted drift vs budget — the cognition-horizon doctrine | **Unchanged** (spec FR-003; scaling here would cancel speed out and dissolve the horizon) |
| 2 | `internal/cognition/governor.go:91` | Adaptive-throttle debt: `seconds × tps / BudgetTicks` | **Unchanged** — already speed-aware by construction; debt measures scheduling pressure against the fiction budget, same doctrine family as Route |
| 3 | `internal/sim/landing.go:214-217` (`rungStale`) | Landing gate: actual staleness vs budget | **Scaled** — the defect site |
| 4 | `internal/mind/convo.go:412` | Mind-side scene-staleness pre-abort for conversation continuations (same predicate as the landing gate, evaluated early to save a model call) | **Scaled identically** — a pre-check that disagrees with the gate it fronts re-creates the class of bug this spec kills ("prompt and gate never disagree" ethos); mind-side, so no replay surface, and the replica carries the same event-sourced `Speed` |
| 5 | `internal/cognition/horizon.go` (all surfaces via `Route`) | Operator-facing horizon | **Unchanged** (FR-003) |
| 6 | Registry doctrine comment `internal/cognition/registry.go:20-22` | Doctrine text | **Reworded** (FR-005): `BudgetTicks` = staleness budget at 1x (wall-clock patience), enforced scaled at landing |

- **Rationale**: the split is principled, not incidental — sites 1/2/5 gate
  *scheduling* against the fiction; sites 3/4 gate *delivery* against the wall.
- **Alternatives considered**: scaling everything (dissolves the horizon
  doctrine, spec-level rejection); scaling only site 3 (leaves the convo
  pre-check able to abort scenes the landing gate would accept).

## R3 — Where the scaling helper lives

- **Decision**: a small pure helper on `DecisionClass` in `internal/cognition`
  (e.g. `EffectiveBudgetTicks(ticksPerSecond float64) int64`), consumed by both
  `rungStale` and the convo pre-check.
- **Rationale**: `internal/cognition` is the leaf package that owns the budget
  doctrine; both consumers already hold a `DecisionClass`. One implementation
  means the two delivery gates can never drift apart, mirroring how all horizon
  surfaces route through `Route`. `internal/sim` and `internal/mind` both
  already depend on `internal/cognition`.
- **Alternatives considered**: inline arithmetic at each site (two copies of
  the uncapped guard — drift risk); a helper in `internal/clock` (wrong owner:
  the budget is cognition doctrine, clock only supplies the rate).

## R4 — Reason-string grammar for scaled rejections

- **Decision**: `staleness %d > budget %d (%d at 1x × %gx)` — e.g.
  `staleness 9800 > budget 9600 (1200 at 1x × 8x)`. The convo pre-abort keeps
  its `scene staleness …` prefix with the same budget clause.
- **Rationale**: spec FR-004/SC-004 — recorded audits must show the effective
  budget, and the derivation makes the decision-trace view self-explanatory
  with no renderer change (`internal/tui/decisions.go` maps outcomes, it does
  not parse reasons).
- **Alternatives considered**: keeping the terse `staleness %d > budget %d`
  with the scaled number only (audit can't tell base from scaled); structured
  payload fields for base/scale (schema change to `cog.outcome` — heavier than
  the string, and nothing machine-consumes the split today).

## R5 — Replay-determinism proof shape

- **Decision**: a reducer-level test alongside
  `internal/sim/governor_replay_test.go` that records a run whose event log
  contains a `speed.set` mid-flight of a pending thought, replays the log, and
  asserts identical landing outcomes; plus table-driven `rungStale` (or
  successor) unit tests covering the spec's acceptance scenarios (4x/8x/32x,
  1x regression, uncapped guard, scaled-reason grammar).
- **Rationale**: mirrors the existing governor replay proof pattern; the
  measured-run half of SC-001 is evidence work on the TASK-122 measure world,
  kept out of the test suite (assumption in spec).
