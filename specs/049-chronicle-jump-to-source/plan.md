# Implementation Plan: Chronicle jump-to-source + input-parity retrofit start

**Branch**: `049-chronicle-jump-to-source` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/049-chronicle-jump-to-source/spec.md`

## Summary

Fill the chronicle inspect-mode reserved `⏎` seam (reorient D3): `⏎` — and,
per the newly ratified input-parity doctrine (decision 8), a mouse click on a
chronicle line — resolves the selected event's subject to a map position and
centers the map camera there; unlocatable events get an honest visible hint in
the detail pane's previously-reserved actions slot. Technical approach: wire
the existing `detailActions(e store.Event)` hook (`internal/tui/tui.go:1166`,
returns `nil` today) to a subject resolver (live replica agent position,
falling back to explicit payload coordinates), express "center on (x,y)" as
the existing `panX/panY` centroid-offset camera (offset = target − wanderer
centroid, render-time clamping already handles bounds), and enable Bubble Tea
mouse events app-wide (`tea.WithMouseCellMotion()` at
`cmd/promptworld/commands.go:745`) binding only the chronicle-line click.

## Technical Context

**Language/Version**: Go 1.24 (repo toolchain; no new language deps)

**Primary Dependencies**: Bubble Tea (existing) — adds `tea.WithMouseCellMotion()` program option + `tea.MouseMsg` handling; no new modules

**Storage**: N/A (pure client-side navigation over the existing event replica)

**Testing**: `go test -race ./...`; table-driven unit tests in `internal/tui`; catalog-sweep-style test over every event type in the digest catalog proving jump-or-hint totality (SC-002, TestCatalogSweep precedent)

**Target Platform**: terminal client (darwin/linux), attached to a running world daemon

**Project Type**: single Go module; feature is one package (`internal/tui`) plus a one-line program-option change in `cmd/promptworld`

**Performance Goals**: no render-path regression; subject resolution is O(payload fields) per keypress, never scans full oversized payloads (windowing discipline, spec edge case)

**Constraints**: keyboard behavior byte-identical for all existing keys (FR-005); narrow fallback must land on the map view after a jump (FR-007); same-PR design-doc amendment gate (`node scripts/check-tui-design.mjs --changed`)

**Scale/Scope**: ~5 files in `internal/tui` touched + 1 line `cmd/promptworld` + 2–3 design pages amended; est. ≤ 500 LOC delta including tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Artifact-Grounded Action** — PASS: spec/plan/tasks under `specs/049-*`; board TASK-124 linked via spec-bridge before implementation; decisions cite reorient D3/decision 8 + spec-047 pages.
- **II. One Task, One PR** — PASS: TASK-124 = one worktree (`.worktrees/task-124`), one branch, one PR carrying code + same-PR design-doc amendments.
- **III. Gates Over Assertions** — PASS: `check-tui-design.mjs --changed` + `go test -race ./...` gate the PR; spec-bridge gate holds board status to artifacts.
- **IV. Grounding Freshness** — PASS (planned): merge touches `internal/tui` sources listed by `docs/wiki/tui-client.md` → wiki-update + player-docs freshness check in the re-ground step.
- **V. Model-Tiered Workflow** — PASS: planned on Fable 5; implementation dispatched to `spec-implementer` on **Sonnet** (routine tier: single-package view/rendering code, tests alongside; no concurrency/doctrine surface). No escalation trigger anticipated.

**Post-Phase-1 re-check**: PASS — design adds no new packages, no new abstractions beyond the already-documented `detailAction` hook; no Complexity Tracking entries needed.

## Project Structure

### Documentation (this feature)

```text
specs/049-chronicle-jump-to-source/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── jump-to-source.md  # UI contract: resolution rules, key/mouse grammar, actions-bar states
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```text
cmd/promptworld/
└── commands.go          # tea.NewProgram(..., tea.WithMouseCellMotion())  [1 line]

internal/tui/
├── tui.go               # detailActions() real impl; handleInspectKey ⏎; tea.MouseMsg routing;
│                        #   jumpToSource(); subject resolver entry; centerCameraOn(x, y)
├── digest.go            # payload subject/position extraction helpers (beside existing
│                        #   per-type digest knowledge; reuse its payload accessors)
├── views.go             # detail-pane actions bar rendering (replaces "[future: actions]");
│                        #   chronicle visible-row window base recorded for click hit-testing
├── layout.go            # (read-only reference) pane rectangles for mouse hit-testing
├── tui_test.go          # key/mouse behavior tests; narrow-fallback jump test
├── digest_test.go       # subject-resolution table tests; catalog sweep jump-or-hint totality
└── render_test.go       # actions-bar rendering states

docs/design/tui/
├── panels/chronicle.md  # jump-to-source row: unbuilt → real symbols; parity-rollout note
├── patterns/keymap.md   # inspect ⏎ row: reserved → jump; first mouse binding recorded
└── (any page whose parity-rollout note this changes) + verified_against re-pins
```

**Structure Decision**: single-project layout; all logic lands in the existing
`internal/tui` package beside the seams spec 047 documented for exactly this
feature (`detailActions` hook, `[future: actions]` slot, reserved `⏎`).
No new packages, no new files expected beyond tests.

## Complexity Tracking

No constitution violations — table intentionally empty.
