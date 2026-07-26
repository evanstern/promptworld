# Contract: Delivery-gate staleness rule (spec 067)

## The rule

For any classed intent landing at tick T in a world whose event-sourced
effective speed at T has tick rate `tps = state.Speed.TicksPerSecond()`:

```
staleness      = max(0, T − SnapshotTick)
effectiveBudget = BudgetTicks                       if tps <= 0   (uncapped)
                = int64(float64(BudgetTicks) × tps) otherwise
outcome         = rejected-stale                    iff staleness > effectiveBudget
```

Properties the implementation MUST preserve:

1. **Purity** (spec FR-002): inputs are event-sourced state (`T`,
   `state.Speed`) and recorded intent fields (`SnapshotTick`, class). No wall
   clock, no estimator, no daemon runtime state.
2. **1x identity** (FR-007/SC-003): `tps == 1` ⇒ `effectiveBudget ==
   BudgetTicks`; every existing 1x outcome is bit-identical.
3. **Scheduling gates untouched** (FR-003): `Route`, `RoutePaused`, governor
   debt, and all horizon surfaces keep the unscaled `BudgetTicks`.
4. **One predicate, two consumers** (research R2/R3): the reducer landing rung
   (`internal/sim/landing.go:rungStale`) and the mind-side conversation
   scene-staleness pre-abort (`internal/mind/convo.go`) both evaluate the rule
   through the same `EffectiveBudgetTicks` helper.

## Reason-string grammar (recorded in `agent.intent_rejected` / `cog.outcome`)

- Scaled rejection (capped speeds):
  `staleness <S> > budget <E> (<B> at 1x × <M>x)`
  where `S` = staleness ticks, `E` = effective budget, `B` = base
  `BudgetTicks`, `M` = tick-rate multiplier (e.g. `8`).
  Example: `staleness 9800 > budget 9600 (1200 at 1x × 8x)`
- Uncapped speed: unscaled form `staleness <S> > budget <B>` (theoretical —
  Route suppresses all classes at uncapped speed).
- Convo pre-abort keeps its `scene staleness …` prefix with the same budget
  clause.

Consumers: TUI decision-trace and digest render these strings opaquely (they
map outcome codes, never parse reasons) — no renderer contract change.

## Reference table (registry values, capped ladder)

| Class (base) | 1x | 4x | 8x | 16x | 32x |
|--------------|-----|-----|-----|------|------|
| planner (1200) | 1200 | 4800 | 9600 | 19200 | 38400 |
| conversation (7200) | 7200 | 28800 | 57600 | 115200 | 230400 |
| meeting (3600) | 3600 | 14400 | 28800 | 57600 | 115200 |

Wall-clock reading: every class's wall patience is constant across capped
speeds (`BudgetTicks` seconds — planner 20 wall-minutes). Residual structural
death (documented per FR-006): wall latency above `BudgetTicks` seconds still
dies at every speed.
