# Implementation Plan: Map Legend Width Policy

**Branch**: `task-191-map-legend-width` | **Date**: 2026-08-02 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/114-map-legend-width/spec.md`

**Board Task**: TASK-191

## Summary

The map legend overflows the terminal it renders into, in two opposite ways: unclamped at
354–356 columns into an 80-column terminal (where it wraps into roughly five rows and
pushes the map off-screen), and clamped-but-unmarked at 112+ columns (where it is cut
mid-token after three symbols with no sign anything was omitted).

The fix is a legend-local width clamp built on `ansi.Truncate`, applied at each of the two
render call sites with the width budget already in scope there, plus a matrix-wide
regression guard in the frame test suite. `clipLine`/`clipContent` — the layout safety net
— are deliberately left untouched; the reasoning is in [research.md](./research.md) R2.

## Technical Context

**Language/Version**: Go (see `go.mod`)

**Primary Dependencies**: `github.com/charmbracelet/lipgloss` v1.1.0,
`github.com/charmbracelet/bubbletea` v1.3.10, and
`github.com/charmbracelet/x/ansi` v0.10.1 — currently indirect, promoted to direct by this
feature (R1)

**Storage**: N/A — this is a render-path change with no persisted state

**Testing**: `go test ./...`; the frame matrix harness in `cmd/promptworld/frames_test.go`;
generated frames under `docs/design/tui/frames/` as the review artifact

**Target Platform**: terminal (Bubble Tea TUI), widths from 80 columns upward

**Project Type**: single Go module, CLI/TUI

**Performance Goals**: N/A — one additional bounded string operation per rendered frame

**Constraints**: no line may exceed the width of the terminal it renders into; the legend
stays one row on every path (FR-008); behavior above the breakpoint must not otherwise
regress (FR-012)

**Scale/Scope**: 2 render call sites, 1 new helper, 1 new test, 1 deny-list, 39 committed
frames regenerated, 1 design-authority page amended

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
| --- | --- | --- |
| I — Artifact-Grounded Action | PASS | Spec carries a reproducible frame audit and file:line diagnosis; TASK-191 exists, is claimed In Progress, and is linked to this spec dir. |
| II — One Task, One PR | PASS | TASK-191 is the single deliverable; one branch, one PR. The follow-up cards this work files are separate deliverables and get their own PRs later, not diffs here. |
| III — Gates Over Assertions | PASS | FR-009's guard converts the fix into a test. The pr gate, the tui-design gate, and the frame-freshness checker all run before merge. |
| IV — Grounding Freshness | PASS | `docs/design/tui/panels/map.md` is amended in-branch (FR-011). Wiki re-pin and player-docs regeneration run in-branch if the pr gate reports either stale. |
| V — Model-Tiered Workflow | PASS | Routine tier — `.claude/agents/spec-implementer.md` (claude-sonnet-5). Rubric checked in research.md; no escalation trigger fires. |

**Post-Phase-1 re-check**: PASS, unchanged. The design added no new package, no new module,
and no cross-package coupling. The one judgment call the design surfaces — the deny-list in
R4 — is registered debt with follow-up cards, not an unjustified violation, and is recorded
in Complexity Tracking below.

## Project Structure

### Documentation (this feature)

```text
specs/114-map-legend-width/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 output — R1..R4 decisions
├── data-model.md        # Phase 1 output — legend segment composition
├── quickstart.md        # Phase 1 output — validation guide
├── contracts/
│   └── legend-width.md  # Phase 1 output — the width invariant as a testable contract
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/tui/
├── views.go             # mapView (narrow, :1737), mapPanelView (widescreen, :907),
│                        # renderMapGrid (legend composition, :1207/:1546),
│                        # clipLine (:1777) and clipContent (:1800) — both UNTOUCHED
├── help.go              # legendGlyphLine + the three prose notes — content, untouched
└── views_test.go        # unit coverage for the new helper

cmd/promptworld/
├── frames.go            # frameSizes (:73) — the declared-width source the guard reads
└── frames_test.go       # + the new matrix-wide width guard

docs/design/tui/
├── panels/map.md        # width policy recorded; the :209 deferral resolved
└── frames/*.txt         # 39 frames regenerated — the review artifact
```

**Structure Decision**: no structural change. The feature touches one existing package
(`internal/tui`) and adds one test to an existing test file in `cmd/promptworld`. The new
helper lives beside the code it serves in `views.go` rather than in a new file, matching
how `clipLine`, `clipContent`, `describeChest`, and `summarizeInventoryContents` are all
already colocated there.

## Implementation Approach

### 1. The helper

A single map-panel-local function that clamps a composed legend to a column budget,
appending `…` only when it actually removes content. Built on `ansi.Truncate` (R1), which
supplies ANSI-safety (FR-006), display-column measurement (FR-007), and the
already-fits no-op (FR-004) without further work.

It must be defensive at the low end (Edge Cases): a budget smaller than the ellipsis plus
one column must yield something renderable rather than a negative length or a panic.

### 2. The two call sites

- `mapView` (`views.go:1750`) — clamp to `m.width` before concatenating. This is the
  defect: today the legend is appended raw, outside the box, with no clip in the
  expression at all.
- `mapPanelView` (`views.go:924-927`) — clamp to `cols-4` before the legend joins
  `content`. After this, `clipContent`'s pass over the legend line is a no-op; the safety
  net remains, but stops being what does the cutting.

### 3. The stale comment

`views.go:1500` asserts that `clipContent` "already clips an over-wide legend." That belief
is what let the narrow path ship unclamped and must be corrected in the same change, not
left to mislead the next reader. The comment at `views.go:928-936` describing `clipContent`
as load-bearing for the legend needs the same correction.

### 4. The guard

A matrix-wide test in `cmd/promptworld/frames_test.go`: for every fixture × state × size,
assert no line exceeds that size's declared width, measured in display columns (not runes —
the audit's own rune-based measurement was known to understate width on rows carrying
double-width glyphs).

It carries an explicit deny-list of the 22 frames failing for the two out-of-scope defect
classes (header row, scenario footer). The deny-list is bidirectional: a listed frame that
*starts* passing must also fail the test, so the debt cannot linger unnoticed after someone
fixes it. Each entry names its follow-up card.

### 5. Regeneration and authority

Regenerate the matrix with `promptworld frames --dump`; the resulting diff across 39 frames
is the review artifact. Amend `docs/design/tui/panels/map.md` to state the width policy and
to resolve the `:209` deferral, whose own re-open trigger — "legend-line overflow pain
becomes the actual bottleneck" — is what this feature answers. Per the skill's authority-vs-
evidence rule, the doc and the frames must be brought into agreement deliberately and in the
same PR, never one silently bent to the other.

### 6. Follow-up cards

File two cards for the classes the guard will detect but this feature does not fix: the
header-row overflow (20 frames) and the scenario keybind footer (2 frames). They are what
the deny-list entries point at.

## Complexity Tracking

> Filled because the design carries one deliberate, registered compromise.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| The FR-009 guard ships with a 13-frame deny-list rather than a clean matrix | FR-010 requires the guard to cover every line of every frame, but two unrelated over-width defect classes (header row, scenario footer) are explicitly out of this feature's scope — the operator scoped the work to the legend. A guard that cannot be introduced without them is a guard that never gets introduced. | Scoping the guard to the legend line only would satisfy FR-009 while gutting FR-010 and leaving the rest of the matrix permanently unguarded. Fixing all three classes here would silently widen scope past what was approved and mix three unrelated causes into one review. The deny-list is visible, can only shrink, fails if an entry starts passing, and each entry names the card that retires it. |
