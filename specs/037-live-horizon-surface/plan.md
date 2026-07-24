# Implementation Plan: Live Cognition-Horizon Surface

**Branch**: `task-41-live-horizon-surface` | **Date**: 2026-07-24 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/037-live-horizon-surface/spec.md`

## Summary

At high speed the router silently suppresses thought classes; the only traces
are raw `cog.outcome` payloads and the per-villager "didn't think" chain
entries. This feature makes the horizon live and aggregate: the daemon counts
every router suppression per class (a mind→orchestrator seam), composes a
per-watched-class standing (suppressed-or-not at the current EFFECTIVE speed,
router arithmetic verbatim, calibration flag, count) on the polled status
reply, and the TUI (header badge + metatron-pane block) and CLI status render
it in plain language with a calibrate-vs-slow-down remedy. All arithmetic stays
in `internal/cognition` (spec-035 single-implementation doctrine); clients
render verbatim and never re-derive.

## Technical Context

**Language/Version**: Go 1.26 (module `github.com/evanstern/promptworld`)

**Primary Dependencies**: stdlib; Bubble Tea + Lipgloss (TUI, existing); no new
dependencies

**Storage**: none — counters are process-lifetime in-memory (daemon);
no schema/file changes

**Testing**: `go test ./...`; table-driven unit tests per package (existing
convention); TUI render tests assert on plain-text lines

**Target Platform**: darwin/linux daemon + terminal client (existing)

**Project Type**: single Go module, package-per-concern (`internal/*`)

**Performance Goals**: status composition stays O(watched classes) per poll
(3 classes × pure arithmetic); counter increment is a mutex map bump on the
mind's absorb path — must never block (telemetry doctrine)

**Constraints**: no-LLM world status bytes byte-identical (additive
`omitempty`, spec 028/034/035 precedent); no new polling loops or push
channels (FR-010); raw enum strings never reach the TUI screen (verdict
glossary doctrine)

**Scale/Scope**: 6 packages touched (`cognition`, `llm`, `mind`, `ipc`,
`tui`, `cmd/promptworld`); ~4 new wire fields; 2 render surfaces + 1 CLI
section; wiki re-pins for 5 notes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|-----------|-------|--------|
| I. Artifact-Grounded Action | Spec/plan/tasks under `specs/037-*`; board task TASK-41 linked via spec-bridge before implementation; decisions recorded in research.md | PASS |
| II. One Task, One PR | TASK-41 ↔ branch `task-41-live-horizon-surface` in `.worktrees/task-41`; one PR; spec docs commit to main at root | PASS |
| III. Gates Over Assertions | spec-bridge gate mirrors phase criteria; status advances only with artifacts (commits, passing tests) | PASS |
| IV. Grounding Freshness | Touched files are pinned sources of `cognition`, `tui-client`, `ipc-protocol`, `ipc-server`, `llm-orchestrator`, `agent-mind`, `cli-promptworld` wiki notes → `/grounding-wiki:wiki-update` + player-docs freshness check before Done | PASS (planned re-pin) |
| V. Model-Tiered Workflow | Planning on Fable 5 (this plan). Implementation delegated to `spec-implementer`. Tier: **Opus 4.8** — cross-package change (6 packages) touching `internal/cognition` and `internal/mind`/`internal/llm` telemetry seams, explicitly named senior-tier territory in the rubric | PASS |

**Post-Phase-1 re-check** (2026-07-24): design adds one optional interface
seam, one pure helper beside its spec-035 siblings, additive wire fields, and
render-only client code — no new projects, no new dependencies, no doctrine
changes. No Complexity Tracking entries needed. PASS.

## Project Structure

### Documentation (this feature)

```text
specs/037-live-horizon-surface/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0: decisions + rationale
├── data-model.md        # Phase 1: entities and wire shapes
├── quickstart.md        # Phase 1: end-to-end validation guide
├── contracts/
│   └── status-horizon.md  # Phase 1: status-reply horizon contract
├── checklists/
│   └── requirements.md  # Spec quality checklist (done)
└── tasks.md             # Phase 2 (/speckit-tasks — not created by plan)
```

### Source Code (repository root)

```text
internal/cognition/
├── horizon.go           # + ClassStanding, LiveHorizon(); SuppressedAt re-based on it
└── horizon_test.go      # + LiveHorizon table tests (incl. uncapped, exclusion)

internal/llm/
├── llm.go               # + Orchestrator.RecordSuppression / SuppressionCounts
└── llm_test.go          # + counter tests (concurrent bumps, copy semantics)

internal/mind/
├── telemetry.go         # + suppressionCounting seam; emitSuppressed hooks it
└── (existing tests)     # + seam test with fake orchestrator

internal/ipc/
├── protocol.go          # + StatusData.Horizon []HorizonClass (omitempty)
├── server.go            # + horizon composition in statusDataFull
└── *_test.go            # + composition tests; no-LLM byte-identity test

internal/tui/
├── views.go             # + header suppression badge; metatron-pane horizon block
└── views_test.go        # + badge/block render tests (both layouts, narrow)

cmd/promptworld/
├── status.go            # + horizon section for LLM worlds
└── (existing tests)     # + render test
```

**Structure Decision**: existing package-per-concern layout; every change is
an additive extension of a file that already owns the concern. No new
packages, no new directories.

## Design Decisions (Phase 0 summary — full rationale in research.md)

- **D1 — Arithmetic home**: new `cognition.LiveHorizon(ticksPerSecond,
  secPerPtFor) []ClassStanding` beside `SuppressedAt` in `horizon.go`;
  `SuppressedAt` becomes a name-filter over it so exactly ONE watched-class
  iteration exists (spec-035 FR-006 posture, spec FR-002).
- **D2 — Counters**: process-lifetime per-class counts owned by
  `llm.Orchestrator`; the mind reports through an optional
  `suppressionCounting` interface seam on `md.orch` (the `estimating`
  pattern), called from `emitSuppressed` — the single suppression terminal.
  Counts ALL classes; the wire surfaces watched ones.
- **D3 — Wire shape**: `StatusData.Horizon []HorizonClass` (additive,
  `omitempty`), composed in `statusDataFull` only when `s.llm != nil`;
  structured fields (class, suppressed, verdict arithmetic, calibrated,
  count) — remedy PHRASING is a client concern, facts ride the wire.
- **D4 — Calibrated classes stay included**: unlike the spec-035 set_speed
  warning (uncalibrated-only gate), the live surface reports actual
  suppression regardless of calibration; the calibrated flag only changes the
  remedy phrase (calibrate vs slow down).
- **D5 — TUI placement**: compact header badge (the `[llm: …]` badge pattern,
  warn-styled) listing suppressed classes; detailed per-class block in the
  metatron dock pane adjacent to the existing LLM provider table
  (`llmProviderLines`), where cognition-ops surfaces already live.
- **D6 — Effective speed source**: `sim.Status.Speed` (post-governor) via the
  existing status door; `Route` itself handles uncapped (`tps ≤ 0`) with
  dedicated phrasing, so FR-003 needs no special casing.

## Complexity Tracking

No Constitution Check violations — table intentionally empty.
