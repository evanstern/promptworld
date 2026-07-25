# Implementation Plan: Run outcomes, the morgue file, death escalation, and graves

**Branch**: `044-run-outcomes-morgue` | **Date**: 2026-07-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/044-run-outcomes-morgue/spec.md`

## Summary

Make permadeath consequential at the run level, on four existing seams. (1) **Run end**:
`stepEvents` detects the last death same-batch and emits a `run.ended` event; its reducer
arm sets an event-sourced `State.Ended` flag; `Loop.Run` gains an ended branch modeled on
the paused branch (no timer, reads served, mutating commands refused), so restart-comes-
back-ended falls out of snapshot+replay. (2) **Morgue**: a new scribe document
(`morgue.md`) — always-on, whole-file re-render from the scribe replica plus a typed
event scan; per-death epitaphs (facts + charter-fingerprint/orders evidence) and a
run-end summary; optional narrator epilogue lands as a recorded `morgue.epilogue` event
so the file stays a pure render. A minimal `metatron.charter_observed` fingerprint event
gives charter revisions an identity for the evidence timeline. (3) **Escalation**: the
gru's survival floor at `gru.go:131` becomes conditional on pre-attack health
`< nearDeathBelow`; kills emit the standard `agent.died` (cause `"gru"`) from `gruStep`
with an inline witness loop, flowing through the unchanged death path. (4) **Graves**:
the `agent.died` reducer arm places a `Structure{Kind: "grave"}` at the death site,
inheriting perception, place-telling, map rendering, and vision-grant plumbing for free;
grief rumors ride the shipped witness-death memory (verified: already the strongest
possible rumor seed). Full decision log with alternatives: [research.md](./research.md).

## Technical Context

**Language/Version**: Go 1.24 (single module, `github.com/evanstern/promptworld`)

**Primary Dependencies**: stdlib + modernc.org/sqlite (store), charmbracelet/bubbletea
(TUI), internal packages only — no new dependencies

**Storage**: append-only SQLite event log + snapshots in `world.db`
(`internal/store`); derived documents as whole-file renders in the world save dir
(`internal/scribe`); no schema change (new event types are payload structs; `State`
gains `omitempty` fields — no `format_version` bump)

**Testing**: `go test ./...`; sim determinism/replay harnesses
(`TestDeterminismSameSeedSameTimeline`, `TestReplayRebuildsState`, `driveTicks`),
`gruTestState` scenario pattern, scribe golden files, TUI digest `TestCatalogSweep` +
header badge tests

**Target Platform**: macOS/Linux daemon + terminal client (unchanged)

**Project Type**: single Go project — daemon (`internal/daemon`, `internal/sim`,
`internal/store`), mind (`internal/mind`), scribe (`internal/scribe`), IPC
(`internal/ipc`), TUI (`internal/tui`), CLI (`cmd/promptworld`)

**Performance Goals**: no regression to tick cadence; morgue render is per-death/boot
(rare) — a bounded event scan per render is acceptable; ended worlds consume no timer
wakeups (idle loop, paused-mode posture)

**Constraints**: replay determinism (canonical JSON, fixed iteration orders, no fresh
RNG rolls — escalation is a pure predicate on recorded state); store single-writer
(run-end emitted from `stepEvents`, never a daemon goroutine); snapshot byte-compat via
`omitempty`; facts never depend on the model (morgue works LLM-off); fixed-frame /
injection whitelists updated explicitly (`morgue.epilogue`)

**Scale/Scope**: 8-villager worlds; ~6 new event-log types
(`run.ended`, `morgue.epilogue`, `metatron.charter_observed`, cause `"gru"` on
`agent.died`, structure kind `"grave"`); touches `internal/sim` (executor, gru, state,
loop), `internal/scribe`, `internal/mind` (narrate), `internal/ipc`, `internal/tui`,
`internal/world`, `internal/metatron` (fingerprint emission), `cmd/promptworld` (status)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.* (v1.1.0)

- **I. Artifact-Grounded Action — PASS.** The plan derives from the six operator
  decisions recorded on TASK-31 (comments #3/#4) and the spec; it produces durable
  artifacts (this plan, research.md, data-model.md, contracts, tasks.md) and the feature
  itself is an artifact-producing one (event log records, morgue document).
- **II. One Task, One PR — PASS.** TASK-31 is the linked deliverable; all user stories
  land as commits on one `task-31` branch in `.worktrees/task-31`, one PR. Spec phases
  are internal breakdown mirrored as criteria via spec-bridge.
- **III. Gates Over Assertions — PASS.** spec-bridge gates TASK-31's status against
  these artifacts; `TestCatalogSweep` and the determinism harnesses are mechanical gates
  the feature must satisfy; no derived state is hand-edited (morgue is a regenerable
  render, never a source of truth — the doctrine is the design).
- **IV. Grounding Freshness — PASS (deferred obligation).** Implementation touches files
  pinned by wiki notes (gru.md, executor.md, event-types.md, world-save-directory.md,
  chronicle.md, scribe/social/mental-maps notes). The task is not done until
  `/grounding-wiki:wiki-update` re-pins after merge; event-types.md gains catalog rows as
  part of the work itself (TestCatalogSweep enforces).
- **V. Model-Tiered Workflow — PASS.** This plan was produced on the planning tier.
  Implementation is delegated to `spec-implementer` subagents; tier per slice recorded on
  TASK-31: US1 (loop/executor/reducer halt semantics — concurrency/scheduling logic) and
  the US2 narrator/fingerprint slice (doctrine-adjacent: injection whitelist, turn
  pipeline) meet the **Opus 4.8** rubric; US3 (single-file gru conditional + tests) and
  US4 (grave structure kind + TUI rendering) are **Sonnet**-eligible routine slices.

**Post-design re-check (after Phase 1): PASS** — no new violations introduced; no
Complexity Tracking entries needed (zero new dependencies, no new subsystems, all four
pieces ride existing seams).

## Project Structure

### Documentation (this feature)

```text
specs/044-run-outcomes-morgue/
├── plan.md              # This file
├── research.md          # Phase 0 — R1..R13 decision log
├── data-model.md        # Phase 1 — entities, state fields, event payloads
├── quickstart.md        # Phase 1 — end-to-end validation walkthrough
├── contracts/
│   ├── events.md        # New/changed event types, payloads, ordering guarantees
│   ├── morgue-document.md  # morgue.md structure contract (export-ready, FR-012)
│   └── status.md        # Status/IPC additive fields, postmortem posture contract
└── tasks.md             # Phase 2 (/speckit-tasks — not created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/sim/
├── executor.go      # run-end detection + emission (R1); witness loop reuse
├── gru.go           # conditional survival floor, gru-kill emission (R4, R5)
├── state.go         # State.Ended (omitempty), run.ended + charter-fingerprint
│                    #   reducer arms, grave placement in agent.died arm (R2, R10)
├── loop.go          # ended branch in Run, handleCommand gating (R2)
├── agents.go        # DiedPayload cause "gru" doc; livingCount helper
└── *_test.go        # gru escalation, run-end, grave, determinism/replay tests

internal/scribe/
└── scribe.go        # renderMorgue: epitaphs + run summary + recorded epilogues (R6, R7)

internal/mind/
└── narrate.go       # epilogue job on the narrator worker; gru-kill chronicle line (R9)

internal/metatron/
└── turn.go / charter.go  # charter content fingerprint, metatron.charter_observed (R8)

internal/world/
└── world.go         # MorguePath() helper

internal/ipc/
└── protocol.go / server.go  # additive ended field on ClockStatus/StatusData (R3)

internal/tui/
├── views.go         # ENDED header token, grave glyph + legend (R12, R10)
├── digest.go        # digest rows for run.ended / morgue.epilogue / gru death
└── tui.go           # postmortem posture from replica flag + push + poll (R12)

internal/tool/
└── registry.go      # placeFactKinds gains "grave" (R10)

cmd/promptworld/
└── commands.go      # status human/JSON + offline snapshot ended field (R3)

docs/wiki/event-types.md  # catalog rows (mechanically gated by TestCatalogSweep)
```

**Structure Decision**: single Go project, existing package boundaries; no new packages.
The one architecturally load-bearing split is chronicle-shaped and already shipped: the
mind writes prose only as recorded events; the scribe renders files only from the replica.

## Complexity Tracking

No Constitution Check violations — table intentionally empty.
