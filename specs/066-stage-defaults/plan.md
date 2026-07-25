# Implementation Plan: Stage-shaped TUI layout defaults

**Branch**: `task-128-stage-defaults` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/066-stage-defaults/spec.md`

## Summary

Resolve each TUI surface's *starting* visibility from the world's curriculum
stage (`internal/world.World.Stage`, `""` = pre-ladder), with every per-surface,
per-stage value carried in one code table that a sweep test asserts is identical
to the authority page `docs/design/tui/patterns/stage-defaults.md`. Defaults
feed the existing fold pipeline (`internal/tui/layout.go` `rowBudget`) as its
starting set; reachability, fold order, capability gating, and takeover firing
are all untouched. Pre-ladder worlds resolve to the union posture and must
render byte-identical to today.

## Technical Context

**Language/Version**: Go 1.24 (module github.com/evanstern/promptworld)

**Primary Dependencies**: bubbletea/lipgloss (existing TUI stack); no new deps

**Storage**: N/A — no new persisted state; reads the existing `World.Stage`
field (spec 046) off the status snapshot the TUI already receives

**Testing**: `go test -race ./...`; frame/golden assertions in `internal/tui`
(existing `render_test.go`/`layout_test.go` conventions); authority-page sweep
test in the `TestCatalogSweep` style (parse the md table, assert parity with
the code table)

**Target Platform**: terminal (same as existing TUI)

**Project Type**: single Go monorepo; this feature is single-package
(`internal/tui`) plus design-doc amendments

**Performance Goals**: no frame-render regression; default resolution is a map
lookup per surface at boot/stage-change (negligible)

**Constraints**: pre-ladder frames byte-identical (SC-002); fold order
provably unchanged (SC-004); spec 046 capability doctrine untouched (FR-007)

**Scale/Scope**: ~1 new file + test in `internal/tui`; touches `layout.go`,
`views.go`, `tui.go`, `lessons.go`, `help.go` wiring; amends
`docs/design/tui/patterns/stage-defaults.md` (specified → shipped) and the
per-surface pages whose default posture it governs

## Constitution Check

*GATE: v1.1.0. Evaluated before Phase 0; re-checked after Phase 1.*

- **I. Artifact-grounded**: spec 066 + this plan + tasks.md on main; board
  TASK-128 linked via spec-bridge before implementation. PASS
- **II. One task, one PR**: TASK-128 → worktree `.worktrees/task-128`, branch
  `task-128-stage-defaults`, one PR. PASS
- **III. Gates over assertions**: spec-047 design gate
  (`check-tui-design.mjs --changed`) + authority-page parity sweep test +
  race suite + merge-drift gates; the sweep test IS the mechanism that keeps
  the code table from drifting off the page. PASS
- **IV. Grounding freshness**: post-merge wiki re-pin (`tui-client.md`,
  `curriculum-ladder.md` list touched sources) + player-docs freshness check.
  PASS (planned in tasks)
- **V. Model-tiered**: Sonnet spec-implementer (single-package view/layout
  code, tests alongside — routine tier). Watch item: the slice touches every
  mode's layout; escalate one-way to Opus if gates fail. Tier + rubric
  recorded on TASK-128 at dispatch. PASS

Post-Phase-1 re-check: no new violations introduced by the design (no new
packages, no new persisted state, no doctrine-adjacent behavior). PASS

## Project Structure

### Documentation (this feature)

```text
specs/066-stage-defaults/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── stage-defaults-table.md   # code-table ↔ authority-page parity contract
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```text
internal/tui/
├── stagedefaults.go        # NEW: the code table + resolution (stage → starting set)
├── stagedefaults_test.go   # NEW: authority-page parity sweep + per-stage frame tests
├── layout.go               # starting-set input to rowBudget (fold order untouched)
├── views.go                # currentStage consumption; tab presence defaults
├── tui.go                  # stage-change re-resolution; explicit-toggle overrides
├── lessons.go              # first-occurrence announcements on newly-appearing surfaces
└── help.go                 # stage-variant guardian section default (table row)

docs/design/tui/
├── patterns/stage-defaults.md  # specified → shipped; re-pin
└── (pages whose defaults the table governs: re-pin only where amended)
```

**Structure Decision**: single-package feature inside `internal/tui`, one new
source file + test so the table and its resolution live in one place; all
other files are wiring edits at existing seams (`rowBudget` inputs,
`currentStage`, lesson catalog).

## Complexity Tracking

No Constitution Check violations — table intentionally empty.
