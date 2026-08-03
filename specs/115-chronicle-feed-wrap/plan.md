# Implementation Plan: Chronicle Raw Feed Wrapping

**Branch**: `task-195-polish-session-1` | **Date**: 2026-08-03 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/115-chronicle-feed-wrap/spec.md`

## Summary

The raw chronicle feed truncates long summaries instead of wrapping them, so a villager's thought
or a line of conversation — the content players read the feed for — is cut mid-sentence and lost.
Two causes: the wrap budget is one line at every width a player reads at, and where wrapping does
occur it has no hanging indent, so continuation lines collide with the tick column.

The approach is deliberately small. The wrap budget gains one new value (`0` = unbounded), the two
existing wrap renderers gain an `indent` parameter derived from each row's own prefix, and a
fixture gains two long-prose events so the behavior becomes visible in the committed frame matrix.
No new types, no state, no payload changes. See [research.md](./research.md) for how each of these
was chosen.

## Technical Context

**Language/Version**: Go (module `github.com/evanstern/promptworld`)

**Primary Dependencies**: Bubble Tea + lipgloss (`internal/tui`); no new dependency

**Storage**: N/A — the feed is a pure projection of the in-memory event ring

**Testing**: `go test ./internal/tui/`, plus the frame harness
(`go run ./cmd/promptworld frames --dump`) as the layout oracle

**Target Platform**: terminal client, all supported sizes; committed frame sizes are 80x30,
112x30, 113x30, 160x50

**Project Type**: single Go module, terminal application

**Performance Goals**: per-frame cost stays O(visible rows) — the existing window-then-format order
(spec 094 R8) is preserved, and the wrap operates only on rows already selected for display

**Constraints**: no emitted line may exceed the pane width at any size; rows that fit on one line
must render byte-identically to today

**Scale/Scope**: two files of production change (`internal/tui/grammar.go`,
`internal/tui/views.go`), one fixture file, one design page, one contract

## Constitution Check

*GATE: evaluated before Phase 0, re-evaluated after Phase 1. Both passes recorded.*

| Principle | Pre-Phase-0 | Post-Phase-1 | Notes |
| --- | --- | --- | --- |
| **I. Artifact-Grounded Action** | PASS | PASS | Diagnosis pinned to file:line on TASK-195 before any code; spec, research, contract, and this plan are the durable trail. The frame diff is the review artifact. |
| **II. One Task, One PR** | PASS | PASS | Rides TASK-195's single branch and single PR alongside the session's other decisions. No new task, no second PR. |
| **III. Gates Over Assertions** | PASS | PASS | `check-tui-design.mjs --changed` gates the design-authority amendment; `check-merge-drift.mjs pr` gates wiki re-pin and player docs; the frame matrix is regenerated, never hand-edited. |
| **IV. Grounding Freshness** | PASS | PASS | `internal/tui/views.go` (13 notes) and `grammar.go` (4) go stale on merge; re-pin rides this PR per spec 069. Footprint recorded on TASK-195. |
| **V. Model-Tiered Workflow** | **DEVIATION** | **DEVIATION** | Implemented inline rather than dispatched to `spec-implementer`. Ratified by the operator as TASK-195 decision 4 and justified in Complexity Tracking below. |

**Spec rigor:** full Spec Kit was run — specify → plan → tasks — with the spec linked to the board
before implementation. `speckit-clarify` was skipped deliberately: the single genuinely ambiguous
question (unbounded vs capped wrap depth) was put to the operator directly and ratified, and no
`[NEEDS CLARIFICATION]` marker survived into the spec.

## Project Structure

### Documentation (this feature)

```text
specs/115-chronicle-feed-wrap/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 — seven resolved unknowns
├── data-model.md        # Phase 1 — entities and invariants
├── quickstart.md        # Phase 1 — validation guide
├── contracts/
│   └── feed-wrap.md     # Phase 1 — normative row shape, budget, indent
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
internal/tui/
├── grammar.go       # styleWrapLine, wrapOrTruncatePlain, chronicleLinePrefix — the wrap renderers
├── views.go         # dockTabContent, chronicleView, chronicleRawBody, renderChronicleRow — budget call sites
├── fixtures.go      # midGameFeed — gains one long thought and one long conversation turn
├── grammar_test.go  # wrap/indent unit coverage
└── tui_test.go      # feed-level behavior

docs/design/tui/
├── panels/chronicle.md   # design authority — amended (FR-013)
└── frames/*.txt          # regenerated evidence, never hand-edited
```

**Structure Decision**: Single Go module, existing package layout. The change is confined to the
terminal client's chronicle rendering; nothing outside `internal/tui` and the design docs is
touched. No new files are created in production code.

## Phase sequencing

1. **Renderers first** — extend the wrap budget domain and add the indent parameter to both
   `styleWrapLine` and `wrapOrTruncatePlain`, keeping them equivalent. Unit-testable with no view
   involvement.
2. **Call sites** — switch the solo and narrow-fallback budgets to unbounded; pass the prefix
   width as the indent; keep the narrow dock's cap of 3.
3. **Fixture** — add the two long-prose events, so the frames can show the behavior.
4. **Regenerate frames** — the review artifact.
5. **Docs** — amend the chronicle design page and land the contract.

Steps 1–2 are the feature; 3–4 are what make it reviewable; 5 is what the spec-047 gate requires.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Principle V — implementation runs inline in the planning session rather than dispatched to `spec-implementer` | Operator ruling, recorded as TASK-195 decision 4: this session's harness forbids subagent dispatch unless requested, and the polish loop's value is the tight discuss → diagnose → implement → live-prove cycle | Dispatching the pinned agent is the simpler-by-doctrine path, but it is unavailable under this session's harness constraints. Mitigation: the model serving the work is `claude-opus-5`, the tier Principle V assigns to hard slices — the work is not served below its rubric tier, only without the delegation hop. Scoped to this session; not a precedent. |
| `maxWrap == 0` as a sentinel for "unbounded" rather than an explicit policy type | Preserves every existing budget value by meaning, so the dock's capped wrap and the truncate path keep their current tests unmodified | A `wrapPolicy{MaxLines, Indent}` struct reads better at call sites but rewrites both renderers and every caller for what two integers express, against the constitution's simplicity constraint. Mitigation: named constant `wrapUnbounded` plus contract §2 documenting the domain. |
