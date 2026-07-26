# Data Model: Staleness budget scaling (spec 067)

No new persisted state, no schema changes. Entities touched:

## DecisionClass (internal/cognition/registry.go)

- **Fields**: unchanged (`Class`, `Points`, `BudgetTicks`, `Degrade`,
  `FutureDated`). Registry values unchanged.
- **Semantics change**: `BudgetTicks` is reinterpreted as the staleness budget
  **at 1x** (wall-clock patience). New derived accessor
  `EffectiveBudgetTicks(ticksPerSecond float64) int64`:
  - `ticksPerSecond <= 0` → `BudgetTicks` (uncapped guard)
  - else → `int64(float64(BudgetTicks) × ticksPerSecond)`
- **Validation**: existing `Validate()` invariants unchanged (positive base
  budgets, Fibonacci points).

## Landing intent (internal/sim — LandIntent input)

- Unchanged shape. `SnapshotTick` remains the staleness numerator
  (`staleness = state.Tick − SnapshotTick`, clamped ≥ 0).

## World state (internal/sim/state.go)

- `Speed clock.Speed` (event-sourced via `speed.set` reductions) — read-only
  scale source at the landing tick. No new fields.

## Recorded events

- `agent.intent_rejected` / `cog.outcome`: payload **shapes unchanged**; only
  the `Reason` string for `rejected-stale` gains the scaled-budget derivation
  (see contracts/landing-gate.md). `StalenessTicks` fields unchanged.

## State transitions

- Landing outcome decision: `OutcomeRejectedStale` now fires iff
  `staleness > EffectiveBudgetTicks(tps at landing tick)`; all other rungs and
  outcome orderings unchanged.
