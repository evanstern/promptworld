# Implementation Plan: World Tuning Manifest (tuning.json)

**Branch**: `048-tuning-manifest` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/048-tuning-manifest/spec.md`

## Summary

Add an optional, boot-loaded `tuning.json` to the world directory as the promotion
path for doctrine constants (control-surface report §6). Five dials are promoted:
`refuelDyingBelow`, `fireBurnPerWood`, `gruEmergePerMille`, `PlannerCadenceTicks`,
and the conversation pair (encounter) cooldown. The technical approach composes
three patterns the codebase already ships:

1. **Clamp-don't-reject validation** (`llm/config.go` `normalizeTokenBudget` /
   `workers()` pattern): every field has a default equal to the current constant
   and a documented min/max; out-of-range values clamp with a warning; malformed
   or unknown fields fail boot.
2. **Boot-seeded event** (`seedMeetingConvention` pattern, `daemon.go:481`): at
   boot, if the manifest's effective values differ from what event-sourced state
   carries, the daemon applies + appends one `sim.tuning_applied` event before
   the loop starts; replay re-applies it, so the seed is one-shot and
   replay-deterministic.
3. **Event-sourced state as the single consumption point**: a new
   `*TuningState` field on `sim.State` (pointer, `omitempty` — pre-048 snapshots
   stay byte-identical, no `format_version` bump; the `Deaths`/`Ended`/`RunEnd`
   precedent at `state.go:96-106`). Reducer-side call sites read accessors on
   `State`; the mind layer reads the same accessors off its replica
   (`md.replica`), which already absorbs every event — both consumption paths
   ride one source of truth with no new plumbing.

## Technical Context

**Language/Version**: Go 1.26.4 (single module, `github.com/evanstern/promptworld`)

**Primary Dependencies**: stdlib only for this feature (`encoding/json`); existing
internal packages `internal/sim`, `internal/world`, `internal/daemon`,
`internal/mind`, `internal/store`

**Storage**: world-dir files (`tuning.json` alongside `manifest.json`,
`calibration.json`, `llm.json`) + the append-only event log (`internal/store`)

**Testing**: `go test ./...`; determinism/replay suites in `internal/sim` and
`internal/daemon` (e.g. `embed_replay_test.go`, `governor_replay_test.go`)

**Target Platform**: the promptworld daemon (macOS/Linux), local worlds

**Project Type**: single Go module — CLI + daemon + TUI

**Performance Goals**: none new — boot-time file read; zero per-tick overhead
beyond a field read (accessors are nil-checked struct reads)

**Constraints**: replay determinism is a hard invariant (values must come from
events, never the file, during replay); pre-048 snapshots and logs must load
byte-compatibly; absent file must be behaviorally indistinguishable from today

**Scale/Scope**: 5 dials, ~6 files touched in `internal/sim`, 1 in
`internal/mind`, 1–2 in `internal/daemon`/`internal/world`, plus tests and the
design-report §6 update

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Artifact-Grounded Action** — PASS. Spec/plan/tasks live in
  `specs/048-tuning-manifest/` on main; the feature is TASK-107 on the board,
  linked via spec-bridge before implementation; tuning values themselves become
  event-log artifacts (that is the feature).
- **II. One Task, One PR** — PASS. TASK-107 → one branch (`task-107-tuning-manifest`)
  in `.worktrees/task-107`, one PR. Spec phases are internal breakdown.
- **III. Gates Over Assertions** — PASS. Board status moves only via
  `spec-bridge:sync` against ticked tasks.md; the SC-001/SC-003 replay checks are
  physical evidence, not assertions.
- **IV. Grounding Freshness** — PASS with follow-through: `docs/wiki/` notes
  listing `internal/sim/state.go`, `agents.go`, `gru.go`, `internal/mind/mind.go`,
  or `internal/daemon/daemon.go` as sources must be re-pinned via
  `/grounding-wiki:wiki-update` after merge; `docs/player/` freshness check after
  that.
- **V. Model-Tiered Workflow** — PASS. Planning (this document) on Fable 5;
  implementation delegated to the `spec-implementer` agent. **Tier: Opus 4.8** —
  the slice is cross-package (sim reducer + mind scheduling + daemon boot),
  doctrine-adjacent (changes how doctrine constants bind), and touches
  planner-cadence scheduling in `internal/mind` — three independent rubric
  triggers. Justification to be recorded on TASK-107.

**Post-Phase-1 re-check**: PASS — design added no projects, no new dependencies,
no violations; Complexity Tracking stays empty.

## Project Structure

### Documentation (this feature)

```text
specs/048-tuning-manifest/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0: decisions + rationale
├── data-model.md        # Phase 1: TuningState, manifest schema, event payload
├── quickstart.md        # Phase 1: end-to-end validation guide
├── contracts/
│   └── tuning.md        # Phase 1: tuning.json schema + sim.tuning_applied contract
├── checklists/
│   └── requirements.md  # Spec quality checklist (done)
└── tasks.md             # Phase 2 (/speckit-tasks — not created by plan)
```

### Source Code (repository root)

```text
internal/
├── sim/
│   ├── tuning.go          # NEW: TuningState, defaults, clamps, ParseTuning,
│   │                      #      NewTuningEvent, accessors on State
│   ├── tuning_test.go     # NEW: clamp table, parse failures, accessor defaults,
│   │                      #      apply/replay determinism, snapshot compat
│   ├── state.go           # State.Tuning field + sim.tuning_applied reducer arm;
│   │                      #      fireBurnPerWood call site → accessor
│   ├── agents.go          # consts become defaults (defaultRefuelDyingBelow, …)
│   ├── policy.go          # refuelDyingBelow call site → accessor
│   ├── executor.go        # fireBurnPerWood call site → accessor
│   └── gru.go             # gruEmergePerMille call site → accessor
├── mind/
│   ├── mind.go            # PlannerCadenceTicks + encounterCooldownTicks reads
│   │                      #      → replica accessors
│   └── embedder.go        # PlannerCadenceTicks bucket reads → replica accessor
├── world/
│   └── world.go           # TuningPath() helper (CalibrationPath pattern)
└── daemon/
    └── daemon.go          # boot: load file → ParseTuning (clamp/warn or fail)
                           #      → seedTuning (apply+append if differs)

docs/design/tui? — no. docs/design/control-surface-and-calibration.md §6 update
```

**Structure Decision**: everything doctrine lives in `internal/sim/tuning.go` —
defaults, clamps, parsing, the event constructor, and the accessors — because
the defaults ARE sim doctrine constants and the reducer must apply the event
deterministically. The daemon only does I/O (read file, warn, seed); the mind
only reads accessors off its replica. `internal/world` contributes a path
helper. No new packages, no new dependencies.

## Complexity Tracking

No constitution violations — table intentionally empty.
