# Implementation Plan: Paused Authoring Chain-Completion

**Branch**: `task-77-paused-chain-completion` | **Date**: 2026-07-24 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/040-paused-chain-completion/spec.md`

## Summary

Complete the last two links of the paused authoring chain (decision-6, TASK-77):
(1) a landed Metatron nudge arms the nudged villager's planner for one
debounce-bounded round at the frozen tick — a new `metatron.nudged` case in
`absorb()`'s arm switch, gated on the replica's paused flag; (2) pause-aware
routing — a paused world routes at zero predicted drift (allow, every class),
with the recorded arithmetic naming the paused state, via a new pure
`cognition.RoutePaused` consulted by `routeVerdict` before any speed math. A
third, subordinate truth fix falls out of FR-004: `newMeta` predicts the land
tick at set speed even while frozen, so paused thoughts predict landing at the
snapshot tick (which also correctly suppresses the future-dating prompt prefix
— `futureDated` returns "" when landing ≤ now). Unpaused code paths are
byte-identical; all paused behavior derives from the replica's event-reduced
`Paused` flag, so replay determinism holds.

## Technical Context

**Language/Version**: Go 1.26.4 (module `github.com/evanstern/promptworld`)

**Primary Dependencies**: standard library only in the touched packages
(`internal/cognition` is pure arithmetic; `internal/mind` uses the in-repo
`internal/sim`, `internal/store`, `internal/llm` seams)

**Storage**: append-only event log via `internal/store` (existing; no schema
changes — the feature consumes existing event types `metatron.nudged`,
`clock.paused`/`clock.resumed`, and emits existing `cog.thought`/`cog.outcome`)

**Testing**: `go test ./...`; mind harness (`newHarnessAt`, scripted model —
internal/mind/telemetry_test.go) for pause scenarios; pure table tests in
internal/cognition; byte-identical replay pattern from
internal/sim/governor_replay_test.go for determinism

**Target Platform**: daemon host (darwin/linux); no user-facing surface changes

**Project Type**: single Go module, event-sourced sim daemon

**Performance Goals**: none new — arming is O(targets) per nudge event in the
absorb loop; the paused route branch is one bool check before existing math

**Constraints**: unpaused behavior byte-identical (FR-003/FR-005/SC-005);
replay determinism (FR-006/FR-007); no new event types, verbs, modes, or
budget changes (FR-008); absorb goroutine ownership rules unchanged

**Scale/Scope**: 3 production files touched (~40 lines), 2 test files extended
plus 1 pure test; no migrations, no UI

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Artifact-Grounded Action | PASS | Spec 040 + this plan derive from decision-6 (accepted) and TASK-77's drift-audited pins; implementation produces commits + PR on the board task. |
| II. One Task, One PR | PASS | TASK-77 → one branch `task-77-paused-chain-completion` in `.worktrees/task-77` → one PR; spec phases are internal breakdown. |
| III. Gates Over Assertions | PASS | Spec linked via spec-bridge before implementation (task AC #5); bridge gate mirrors phase criteria; status advances only on artifacts. |
| IV. Grounding Freshness | PASS (post-merge obligation) | `internal/mind/mind.go`, `internal/mind/telemetry.go`, `internal/cognition/route.go` are wiki-note sources (agent-mind, cognition-horizon et al.) — `/grounding-wiki:wiki-update` + player-docs freshness check required after merge. |
| V. Model-Tiered Workflow | PASS | Doctrine-adjacent behavior change in `internal/mind` routing/wake semantics ⇒ **Opus 4.8 senior tier**, delegated via `spec-implementer` with `model: opus`; planning tier (this document) writes no implementation code. Tier + rubric justification recorded on TASK-77. |

**Post-Phase-1 re-check**: PASS — design adds one pure function, one absorb
case, and two guarded branches; no new packages, seams, events, or modes. No
Complexity Tracking entries needed.

## Project Structure

### Documentation (this feature)

```text
specs/040-paused-chain-completion/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0: decisions + verified grounding
├── data-model.md        # Phase 1: event/verdict shapes consumed & produced
├── quickstart.md        # Phase 1: validation guide
├── contracts/
│   └── recorded-events.md  # Phase 1: recorded-event & arithmetic contracts
├── checklists/
│   └── requirements.md  # Spec quality checklist (complete)
└── tasks.md             # Phase 2 (/speckit-tasks — not created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/cognition/
├── route.go             # + RoutePaused(dc, secondsPerPoint) Verdict (pure)
└── route_test.go        # + paused-verdict table cases

internal/mind/
├── mind.go              # absorb(): + case "metatron.nudged" (paused-gated arm)
├── telemetry.go         # routeVerdict(): paused branch before speed math;
│                        # newMeta(): paused ⇒ predictedLandTick = snapshotTick
├── mind_test.go         # + arming tests (paused arms targets; running arms nothing)
└── telemetry_test.go    # + paused-route/land/second-nudge harness tests
                         #   (alongside TestPauseStartsNoNewThoughts, which stays green)

internal/sim/            # untouched (payloads, reducer, clock events all exist)
```

**Structure Decision**: two existing packages only — `internal/cognition` gains
the pure paused verdict (keeping ALL routing arithmetic in the cognition
package, where spec 007 put it), and `internal/mind` gains the paused-gated
wake + the two truth branches. No daemon, IPC, TUI, or sim changes: the chat
handler already lacks a pause gate, the nudge landing batch already exists, and
the replica already reduces `clock.paused`/`clock.resumed` into `State.Paused`.
