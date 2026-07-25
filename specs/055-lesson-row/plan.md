# Implementation Plan: First-occurrence lessons projection (lesson row)

**Branch**: `task-117-lesson-row` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/055-lesson-row/spec.md`

## Summary

A purely client-side teaching projection in the TUI: a static, skin-tokened lesson
catalog whose entries fire on first occurrence of cataloged events (mechanics +
prompting tiers), render in a new two-line lesson row above the guardian strip
(one-active / dwell / dismiss / opportunity-decay), persist per-user seen-state beside
`unlocks.json`, and feed the same catalog into the help overlay's existing
`helpLessons` seam as the pull half. No daemon, IPC, or event-schema changes.

## Technical Context

**Language/Version**: Go 1.24 (module `github.com/evanstern/promptworld`)

**Primary Dependencies**: Bubble Tea + Lip Gloss (existing TUI stack); no new
dependencies

**Storage**: one new per-user JSON file `~/.promptworld/lessons-seen.json`
(atomic-write, load-tolerant — `internal/worlds/unlocks.go` precedent)

**Testing**: `go test -race ./...`; table-driven render tests
(`internal/tui/render_test.go` conventions); event-fixture sweeps for the
first-occurrence invariant

**Target Platform**: terminal client (`internal/tui`), all platforms the TUI ships on

**Project Type**: single Go module; feature is TUI-client + one small `internal/worlds`
addition

**Performance Goals**: zero added model calls (static strings); projection cost
O(events polled), same order as the existing decision-trace ingest

**Constraints**: lesson row ≤ 2 rows always; no raw `{{…}}` literals rendered;
seen-state advisory (failed load/write never blocks boot or play); no new event
types; keyboard-first (`x` binding + documented no-op fallthrough)

**Scale/Scope**: 8-lesson minimum catalog; ~1 new file in `internal/tui`, ~1 in
`internal/worlds`, wiring edits in ≤ 5 existing TUI files; 2 design pages re-pinned +
keymap page flip

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Artifact-grounded**: spec/plan/tasks committed under `specs/055-lesson-row/`;
  board task TASK-117 linked via spec-bridge before implementation. PASS.
- **II. One task, one PR**: TASK-117 → worktree `.worktrees/task-117` → branch
  `task-117-lesson-row` → one PR. Spec phases are internal breakdown. PASS.
- **III. Gates over assertions**: spec-047 TUI gate (`check-tui-design.mjs --changed`)
  runs in the worktree; `panels/lesson-row.md` + `patterns/keymap.md` amended in the
  same PR (`status: specified → shipped`, real renderer symbols, re-pin). Bridge gate
  governs board status. PASS.
- **IV. Grounding freshness**: merge touches `internal/tui/*` (sources of
  `docs/wiki/tui-client.md`) → wiki-update + player-docs freshness check in the
  re-ground step. PASS (planned).
- **V. Model-tiered workflow**: planning on Fable 5 (this document); implementation
  dispatched to `spec-implementer` on **Sonnet** — single-package view/rendering code
  with tests, the rubric's routine tier; no concurrency/governor logic, no
  doctrine-adjacent behavior change (runbook Lane 3 concurs). Escalation one-way per
  rubric if gates fail. PASS.

Post-Phase-1 re-check: no new violations introduced by the design (no new packages
beyond the two files above, no new dependencies, no daemon surface). PASS.

## Project Structure

### Documentation (this feature)

```text
specs/055-lesson-row/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   ├── lessons-catalog.md    # catalog entry contract + minimum taxonomy
│   └── seen-state-file.md    # per-user record: path, schema, semantics
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```text
internal/tui/
├── lessons.go           # NEW: catalog, trigger projection, row state machine,
│                        #      queue/decay, skin-token resolution seam
├── lessons_test.go      # NEW: invariant sweep, render tests, seen-state tolerance
├── tui.go               # wire ingest into the poll loop; `x` key dispatch;
│                        #      populate helpLessons from the catalog at init
├── layout.go            # lesson-row rows in chrome budget + fold order
│                        #      (patterns/layout.md ruling (a))
├── views.go             # render the two-line row above the guardian strip;
│                        #      [lesson] header badge (folded/stage-3+ state)
├── help.go              # (likely no edit — helpLessons seam already renders
│                        #      entries; populate-only from tui.go/lessons.go)
└── render_test.go       # boundary cases if conventions demand co-location

internal/worlds/
├── lessons.go           # NEW: LessonsSeenPath/LoadLessonsSeen/MarkLessonSeen —
│                        #      unlocks.go pattern (load-tolerant, atomic write)
└── lessons_test.go      # NEW

docs/design/tui/
├── panels/lesson-row.md # same-PR amendment: specified → shipped, renderer
│                        #      symbols, re-pin
└── patterns/keymap.md   # same-PR amendment: `x` moves from "specified, unbuilt"
                         #      to the global table; re-pin
```

**Structure Decision**: follow the decision-trace precedent exactly — one
self-contained projection file in `internal/tui` (as `decisions.go` is) plus a
persistence sibling in `internal/worlds` (as `unlocks.go` is). No new packages.

## Complexity Tracking

No constitution violations; table not needed.
