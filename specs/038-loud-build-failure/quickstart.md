# Quickstart: validating Loud Build Failure & Occupancy Tolerance

**Spec**: [spec.md](spec.md) · **Contract**: [contracts/agent-build-failed.md](contracts/agent-build-failed.md)

## Prerequisites

- Go toolchain (per go.mod); repo root or the task worktree
  (`.worktrees/task-91`).
- `export PATH="/opt/homebrew/bin:$PATH"` for go/gh.

## Run the regression suite

```sh
go test ./internal/sim/ -run 'Wall|BuildFailed|Replay' -v
go test ./...
```

Expected green tests proving each story:

| Story / FR | Test (internal/sim/wall_test.go unless noted) | Asserts |
|------------|----------------------------------------------|---------|
| US3/FR-004 passerby tolerance | `TestWallBuildToleratesPasserby` (new) | second agent crosses res tile mid-work and leaves → `agent.built`, wall stands, planks spent; zero `agent.build_failed` |
| US3/FR-005/FR-006 squatter | `TestWallOccupancyGuard` (rewritten at :300) | agent parked on res tile at due tick → no `agent.built`, completion defers, then exactly one `agent.build_failed` (reason `site blocked too long`) + failure memory after `wallOccupancyGraceTicks`; planks unspent; no wall |
| US1/FR-001/FR-007 site vanished | `TestWallBuildSiteVanishedFailsLoud` (new) + one non-wall build case | site invalidated mid-work → `agent.build_failed` (reason `site no longer buildable`) + memory, never bare `intent_done` |
| US2/FR-003 memory | asserted inside the failure tests | same-tick `agent.memory_added`, `OriginAction`, text states build did NOT complete |
| FR-009 replay | `whole_feature_test.go` `TestReplayByteIdentityWallsAxesPaths` (:617) | expected-event sets include `agent.build_failed` + paired memory; byte-identical replay |
| FR-008 re-arm | existing planner re-arm coverage + failure tests | intent cleared, `IdleSince` stamped, builder plans again |

## Manual smoke (optional)

1. Run a world; watch the TUI digest.
2. Steer/wait for a wall build with bystanders: a crossing villager no longer
   kills the build; a blocked build eventually prints a failure line naming
   the builder, the goal, and the reason — never "finished".

## Documentation check

- `docs/wiki/event-types.md` has the `agent.build_failed` row and the amended
  `agent.intent_done` row per the contract.
- After merge: `/grounding-wiki:wiki-update` re-pins notes sourcing
  `executor.go`/`state.go`/`digest.go`; then player-docs freshness check.
