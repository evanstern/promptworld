# Quickstart: validating spec 067 (staleness budget scaling)

## Prerequisites

- Go toolchain per `go.mod`; repo root.

## Unit + replay proof (the "proven by test" arm of AC#2)

```sh
go test ./internal/cognition/ -run TestEffectiveBudget -v   # helper table: 1x identity, ladder products, uncapped guard
go test ./internal/sim/ -run 'Stale' -v                     # landing rung: 8x lands the measured regime, 1x regression, reason grammar
go test ./internal/sim/ -run 'StalenessReplay' -v           # speed.set mid-flight replay determinism (US2)
go test ./...                                               # full suite green (SC-002/SC-003)
```

Expected: all pass; the landing tests encode the measured regime
(2000 ticks staleness @ 8x → lands; 1300 @ 1x → rejects; >9600 @ 8x →
rejects with `staleness … > budget 9600 (1200 at 1x × 8x)`).

## Measured-run evidence (the "measured run" arm of SC-001, optional if impractical)

1. Start the TASK-122 measure-world profile (gemma4:12b-mlx all-routes,
   parallel=4, 8 villagers) at 8x.
2. Let it run long enough to accumulate ≥50 plan landings.
3. Count outcomes over the event log:
   `set_plan` landings vs `cog.outcome` `rejected-stale` for class `planner`
   (the TASK-122 notes carry the exact queries).
4. Expected: rejected-stale share < 10% (was ~91%).

Record the numbers as evidence on TASK-141 (board note), per the project's
evidence conventions.

## Surfaces check

- Open the decision-trace view on any rejected-stale event recorded after the
  change: the reason must name the scaled budget with its derivation
  (SC-004) — no renderer change expected.
- `docs/wiki/` cognition-horizon note documents the scheduling-vs-delivery
  split and the residual gap (SC-005); `node scripts/check-merge-drift.mjs
  session` reports grounding not stale after the wiki re-pin.
