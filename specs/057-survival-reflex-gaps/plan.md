# Implementation Plan: Survival Reflex Gaps (fire)

**Branch**: `task-108-survival-reflex` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

## Summary

Three moves, all riding shipped machinery: (1) `defaultRefuelDyingBelow` 3600 →
10800 in `internal/sim/tuning.go` (the dial's doctrine home since spec 048);
(2) genesis tuning pin — `promptworld new` world creation seeds one
`sim.tuning_applied` event with the full default set, using spec 048's existing
`NewTuningEvent` + reducer arm, so future default changes never reach into
post-057 worlds' replays; (3) a reflex proof matrix + survival audit over the
existing ladder (`internal/sim/policy.go` `decideIntent`), fixing only
fire-adjacent surgical gaps.

## Technical Context

**Language/Version**: Go 1.26.4 · **Dependencies**: none new · **Storage**:
event log (genesis event) · **Testing**: `go test ./...`; matrix tests in
`internal/sim` · **Target**: daemon + `promptworld new` path ·
**Constraints**: replay determinism; no format_version bump; spec 048 compat
guarantees intact · **Scope**: `internal/sim/tuning.go` (default), the world
genesis path (find it: where `promptworld new` seeds the initial log — likely
`internal/world` + genesis seeding in daemon or cmd), `internal/sim/policy.go`
tests, docs listed in FR-006.

## Constitution Check

- **I Artifact-grounded** — PASS: spec/plan/tasks on main; audit artifact is a
  deliverable (FR-005).
- **II One task, one PR** — PASS: TASK-108 → `.worktrees/task-108`, one PR.
- **III Gates** — PASS: spec-bridge drives status; SC tests are the evidence.
- **IV Grounding freshness** — PASS with follow-through: `world-tuning`,
  `reflex-policy`, `daemon-lifecycle` (if genesis path touches it) wiki notes
  re-pin post-merge; player-docs check after.
- **V Model-tiered** — PASS: implementation delegated to spec-implementer at
  **Opus 4.8** (doctrine-adjacent survival behavior in the sim reducer;
  replay-affecting genesis change). Recorded on the task at dispatch.

**Post-Phase-1 re-check**: PASS — no new packages, no violations.

## Phase 0 research decisions

- **R1 — default change location**: `defaultRefuelDyingBelow` in
  `internal/sim/tuning.go:27` is the single home (048 relocated it); the dial
  contract table and report rows move in the same PR (FR-006); wiki re-pin
  post-merge.
- **R2 — genesis pin mechanism**: reuse `sim.NewTuningEvent(tick=0/genesis
  tick, defaultTuning())` appended with the other genesis events at world
  creation. Find the genesis seam: `promptworld new` → world creation +
  first events (grep `genesis` in cmd/ and internal/). The boot seed
  (`seedTuning`) needs NO change: it already compares against
  `state.EffectiveTuning()`, and a pinned state simply makes the comparison
  explicit. `migrate` is explicitly out (spec assumption).
- **R3 — proof matrix placement**: extend `internal/sim/policy_test.go` (or a
  new `reflex_matrix_test.go`) — table-driven over the US3 cells; reuse
  existing test fixtures for known-fact injection (see
  `food_fire_test.go` for the fixture idiom).
- **R4 — audit artifact home**: `specs/057-survival-reflex-gaps/audit.md`,
  linked from the task notes; carded gaps get board ids inline.

## Project Structure

```text
specs/057-survival-reflex-gaps/  spec.md, plan.md, tasks.md, audit.md (US4), checklists/
internal/sim/tuning.go           default 3600 → 10800 (+ comment)
<genesis seam>                   one NewTuningEvent at world creation
internal/sim/reflex_matrix_test.go  US3 proof matrix
internal/sim/tuning_test.go      genesis-pin replay test (SC-002)
specs/048-tuning-manifest/contracts/tuning.md  default column update (FR-006)
docs/design/control-surface-and-calibration.md §2.4/§6 rows (FR-006)
docs/design/grounded-assumptions.md or TASK-75's target doc — hazard note (FR-007; put it
  where the determinism scope note lives today; if TASK-75's doc doesn't exist yet, the
  hazard note lands in the control-surface report §6 and TASK-75 gets a pointer note)
```

## Complexity Tracking

None — no violations.
