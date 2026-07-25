# Implementation Plan: Guardian strip — always-visible action budget line

**Branch**: `050-guardian-strip` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/050-guardian-strip/spec.md`

## Summary

Render one borderless action-budget row directly above the minibuffer in the
widescreen composite (and carried in narrow): charge bank (glyph run +
`(N/cap)`), next-regen forecast, standing-order count; faith omitted until
TASK-118 ships. Honest degradation per segment; folds last under height
pressure by relocating into the minibuffer's dormant line (layout.md ruling
a step 4). Technical approach: a new `guardianStripView(width)` renderer in
`internal/tui/views.go` fed by the same `m.status`/replica fields the
guardian tab header already reads (`views.go:1509-1512`, `Status.Orders`);
the composite (`views.go:284`) and narrow fallback gain the row; the row
budget in `internal/tui/layout.go` (`computeRows`) grows a strip row with a
fold threshold; regen boundary derived from the existing
`chargeRegenTicks` cadence exposure.

## Technical Context

**Language/Version**: Go 1.24 (repo toolchain; no new deps)

**Primary Dependencies**: Bubble Tea + lipgloss (existing); no new modules

**Storage**: N/A (pure render over existing replica/status)

**Testing**: `go test -race ./...`; render-state fixture tests in `internal/tui/render_test.go` (segment presence/absence sweep — SC-002); row-budget/fold-order sweep in `internal/tui/layout_test.go` (SC-003)

**Target Platform**: terminal client (darwin/linux)

**Project Type**: single Go module; one package (`internal/tui`) only

**Performance Goals**: no render-path regression; strip assembly is O(segments) string work per frame

**Constraints**: exactly 1 row at every width (truncate, never wrap); fold relocates content (dormant minibuffer state only); no new fiction literals (skin-tokens.md rule 5 — all segments non-fiction chrome); same-PR design-doc amendment gate

**Scale/Scope**: ~3 files in `internal/tui` + tests; 3–4 design pages amended; est. ≤ 400 LOC delta including tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Artifact-Grounded Action** — PASS: spec dir `specs/050-*`; board TASK-126 linked via spec-bridge pre-implementation; behavior contract is the authored `panels/guardian-strip.md`.
- **II. One Task, One PR** — PASS: TASK-126 = `.worktrees/task-126`, one branch, one PR (code + doc amendments).
- **III. Gates Over Assertions** — PASS: `check-tui-design.mjs --changed` + `go test -race ./...`; spec-bridge gate on board status.
- **IV. Grounding Freshness** — PASS (planned): `internal/tui` sources → wiki-update + player-docs freshness in re-ground.
- **V. Model-Tiered Workflow** — PASS: planned on Fable 5; implementation on **Sonnet** (routine tier: single-package rendering + layout arithmetic, tests alongside). No escalation trigger anticipated.

**Post-Phase-1 re-check**: PASS — no new packages/abstractions; one new renderer + row-budget extension. No Complexity Tracking entries.

## Project Structure

### Documentation (this feature)

```text
specs/050-guardian-strip/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── guardian-strip.md  # segment grammar, degradation rules, fold relocation
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
internal/tui/
├── views.go             # guardianStripView(width); composite row insertion (views.go:284);
│                        #   narrow-fallback carry (views.go:1558 vicinity); dormant-line
│                        #   relocation in minibufferView (views.go:1692)
├── layout.go            # rowBudget gains Strip row; computeRows fold threshold
│                        #   (strip folds last, per layout.md ruling a step 4)
├── tui.go               # (minor) regen-boundary helper if clock math lives beside Model
├── render_test.go       # segment presence/absence fixture sweep (SC-002); truncation
├── layout_test.go       # row-budget arithmetic + fold-order sweep over heights (SC-003)
└── views_test.go        # composite/narrow carry assertions

internal/sim/            # (read-only) chargeRegenTicks / MetatronChargeCap exports —
                         #   if chargeRegenTicks is unexported, export a read-only
                         #   accessor mirroring the MetatronChargeCap pattern (agents.go:843)

docs/design/tui/
├── panels/guardian-strip.md   # specified → shipped; real renderer symbols; spec 050's
│                              #   three added rulings recorded (full-bank omission,
│                              #   pre-status blank, truncation order)
├── patterns/layout.md         # row-budget/fold rows re-verified, re-pinned
├── panels/minibuffer.md       # dormant-state relocation form recorded, re-pinned
└── (INDEX/anatomy re-pins only if the check script requires)
```

**Structure Decision**: single-project layout; rendering + layout arithmetic
in `internal/tui` beside the surfaces the authored page names. The only
possible out-of-package touch is exporting the regen cadence read-only from
`internal/sim` (established export pattern), decided in research R2.

## Complexity Tracking

No constitution violations — table intentionally empty.
