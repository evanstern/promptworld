# Tasks: Map Legend Width Policy

**Feature**: `specs/114-map-legend-width` | **Board Task**: TASK-191
**Branch**: `task-191-map-legend-width` | **Date**: 2026-08-02

**Input**: [spec.md](./spec.md), [plan.md](./plan.md), [research.md](./research.md),
[data-model.md](./data-model.md), [contracts/legend-width.md](./contracts/legend-width.md),
[quickstart.md](./quickstart.md)

Tests are included because the spec requires them: FR-009 and FR-010 make a regression
guard part of the deliverable, and contract clauses C2/C3 are stated as testable
properties.

---

## Phase 1: Setup

- [x] T001 Promote `github.com/charmbracelet/x/ansi` from indirect to direct in `go.mod` (already present at v0.10.1 via lipgloss; no new module enters `go.sum`), then run `go mod tidy` and confirm `go build ./...` is clean

---

## Phase 2: Foundational

**Blocking prerequisite for every user story below.** The helper is the one piece both
render paths call; nothing else can proceed until it exists.

- [x] T002 Add the legend-local clamp helper to `internal/tui/views.go`, near `clipLine`/`clipContent`: it takes the composed legend and a column budget, returns it unchanged when it already fits, and otherwise returns it truncated to the budget with a trailing `…`. Build it on `ansi.Truncate(s, budget, "…")` per research.md R1 — do NOT use `truncateRunes` (rune-based, breaks ANSI) and do NOT modify `clipLine` or `clipContent` (research.md R2)
- [x] T003 Make the helper defensive at the low end per spec.md Edge Cases: a budget at or below the width of the ellipsis itself must return something renderable — never a negative length, never a panic
- [x] T004 [P] Unit-test the helper in `internal/tui/views_test.go`: fits-exactly returns unchanged with no `…` and no padding (C2 second half); over-budget gains a trailing `…` and measures at most the budget (C2 first half); ANSI-styled input is not severed (C5); double-width glyph input is measured in display columns, not runes (FR-007); degenerate budgets (0, 1, negative) do not panic

**Checkpoint**: helper exists and is unit-tested; no render path uses it yet.

---

## Phase 3: User Story 1 — The legend stops destroying the narrow layout (P1) 🎯 MVP

**Goal**: at 80 columns the legend occupies exactly one row and the map grid stays fully
visible, instead of the legend wrapping into roughly five rows and pushing the map
off-screen.

**Independent test**: render any fixture at 80x30; no line exceeds 80 columns and the map
box is fully present. Requires nothing from US2 or US3.

- [x] T005 [US1] In `mapView` (`internal/tui/views.go:1737`), clamp the legend with the T002 helper to a budget of `m.width` before concatenating it, replacing the bare `styleBox.Render(grid) + "\n" + legend` at `views.go:1750`. `m.width` is the correct budget because the narrow legend renders outside the map box, so the full terminal width is genuinely available (research.md R3)
- [x] T006 [US1] Regenerate the narrow frames with `go run ./cmd/promptworld frames --dump` and confirm every `*__80x30.txt` legend line is now at most 80 columns, and that the map grid and its box border are fully present in the frame
- [x] T007 [P] [US1] Verify FR-008 holds on this path: the legend is still exactly one rendered row at 80 columns — the fix must clamp, never wrap to a second row (contract C4)

**Checkpoint**: the 80-column layout is usable. This alone is a shippable increment.

---

## Phase 4: User Story 2 — Truncation is honest (P2)

**Goal**: a legend that was shortened says so, at every size, so a player cannot mistake a
fragment for the whole symbol key.

**Independent test**: render at 112x30; the legend ends in `…` when content was dropped,
and carries no `…` when everything fit. Independent of US1's narrow path.

- [x] T008 [US2] In `mapPanelView` (`internal/tui/views.go:907`), clamp the legend with the T002 helper to a budget of `cols-4` before it joins `content` at `views.go:924-927`. `cols-4` is the true usable width: the box is given `.Width(cols-2)` and `Padding(0,1)` consumes two more, which is the same arithmetic `clipContent` performs internally (research.md R3)
- [x] T009 [US2] Correct the stale comment at `internal/tui/views.go:1500` asserting that `clipContent` "already clips an over-wide legend" — true of the widescreen path, false of the narrow one, and that false belief is what let the defect ship. Also correct the `views.go:928-936` block describing `clipContent` as load-bearing for the legend: after T008 the legend arrives already within budget and `clipContent` is a no-op for it, still present as the layout safety net but no longer what does the cutting
- [x] T010 [US2] Regenerate frames and confirm `mid-game__home__112x30.txt:22` now ends in `…` rather than cutting mid-token after `"forage`
- [x] T011 [P] [US2] Verify C2's second half at a width where the whole legend fits: no `…` appears and the line is not padded — an implementation that always appends an ellipsis must fail this

**Checkpoint**: truncation is visible wherever it happens.

---

## Phase 5: User Story 3 — A wider terminal is rewarded (P3)

**Goal**: every additional column reveals more legend, up to the full untruncated text.

**Independent test**: render one fixture and state across increasing widths; visible legend
content is monotonically non-decreasing.

- [x] T012 [US3] Verify contract C3 across the committed sizes and several off-matrix widths (80, 100, 112, 113, 140, 160) using `go run ./cmd/promptworld frames --fixture mid-game --state home --size WxH`, confirming visible legend content never decreases as width grows — this rules out a fixed cap that would satisfy C1 and C2 while ignoring available space
- [x] T013 [P] [US3] Add a regression test asserting monotonicity of rendered legend width across at least three increasing widths, so a future fixed-cap change cannot silently reintroduce the problem

**Checkpoint**: all three user stories complete; the width policy is fully realized.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T014 Add the matrix-wide width guard to `cmd/promptworld/frames_test.go`: for every fixture × state × size, assert no rendered line exceeds that size's declared width, reading widths from `frameSizes` (`cmd/promptworld/frames.go:73`) rather than parsing filenames, and measuring **display columns** (`ansi.StringWidth`) rather than runes — rune counting understates width wherever double-width glyphs appear, which is exactly how the original audit produced false negatives (contract C1)
- [x] T015 Give the T014 guard a bidirectional deny-list of the 22 known-failing frames from the two out-of-scope defect classes — the header row (20 frames at 81–83 columns, all `*__80x30.txt` line 1) and the scenario keybind footer (2 frames, `scenario__home__{112,113}x30.txt` line 30). A frame not on the list that fails is a failure; a frame **on** the list that passes is also a failure, so the debt cannot silently outlive its fix. Each entry names its follow-up card from T018
- [x] T016 Amend `docs/design/tui/panels/map.md` to state the width policy (one row always; clamp to the available budget; truncate the tail with `…`; the segment order in data-model.md is what the truncation eats from) and to resolve the spec 060 deferral at `map.md:209`, whose own re-open trigger — "legend-line overflow pain becomes the actual bottleneck" — is precisely what this feature answers. Reconcile against `map.md:125-132`, which says the legend "grows the line rather than adding a second row": that intent is preserved, but now has a stated terminus
- [x] T017 Run `node scripts/check-tui-design.mjs --changed` and re-verify + re-pin every affected page under `docs/design/tui/`, per the spec 047 UI authority gate — required before any PR touching `internal/tui/`
- [x] T018 [P] File two follow-up Backlog cards for the out-of-scope defect classes the T014 guard detects but this feature does not fix: the header-row overflow (20 frames) and the scenario keybind footer overflow (2 frames). Each card carries its frame evidence; their ids are what T015's deny-list entries reference
- [x] T019 Run the full validation sweep from [quickstart.md](./quickstart.md): the frame audit reports only the 22 deny-listed out-of-scope frames; `go test ./...` is green; `node .claude/skills/tui-frames/scripts/check-frames.mjs --check` passes; the 39-frame diff is coherent and is the review artifact
- [x] T020 Confirm the guard actually bites before trusting it: temporarily widen a line in a committed frame and re-run T014 (must fail); remove a deny-list entry without fixing its frame (must fail); add a deny-list entry for a passing frame (must fail). Revert all three probes
- [x] T021 Run `node scripts/check-merge-drift.mjs pr` from the worktree and clear every blocking finding in-branch. `wiki-repin-missing` and `player-docs-stale` have **no bypass flag**: re-verify and re-pin any `docs/wiki/` note whose pinned sources this branch touched, and regenerate `docs/player/` if `docs/wiki/` changed
- [x] T022 Record on TASK-191, per constitution Principle V, the implementation tier that actually served, with its rubric justification — including the disclosed deviation that implementation ran inline on the planning model rather than via a delegated `spec-implementer` subagent, because this session's operating instructions forbid spawning agents unrequested

---

## Dependencies

```text
T001 (dep) ─▶ T002 (helper) ─▶ T003 (defensive) ─┬─▶ T004 (unit tests) [P]
                                                  │
                                                  ├─▶ Phase 3 US1: T005 ─▶ T006 ─▶ T007 [P]
                                                  ├─▶ Phase 4 US2: T008 ─▶ T009 ─▶ T010 ─▶ T011 [P]
                                                  └─▶ Phase 5 US3: T012 ─▶ T013 [P]
                                                                        │
Phases 3–5 complete ─▶ T014 ─▶ T015 ─▶ T016 ─▶ T017 ─▶ T019 ─▶ T020 ─▶ T021 ─▶ T022
                                  ▲
                        T018 [P] ─┘  (card ids feed T015's deny-list)
```

**Story independence**: US1 and US2 touch different functions (`mapView` vs.
`mapPanelView`) and can proceed in parallel once T003 lands. US3 is verification over both
and is best run after them, though T013's test can be written earlier.

**Hard ordering**: T014 depends on all three stories, because the guard cannot pass until
the legend class is fixed. T015 depends on T018 only for the card ids it cites. T021 is
last because the pr gate audits the finished branch.

## Parallel Execution Opportunities

- T004 (unit tests) runs alongside the first render call site
- T005–T007 (US1) and T008–T011 (US2) are parallelizable across the two call sites
- T007, T011, T013 are independent verifications within their stories
- T018 (follow-up cards) is independent of all code work and can be filed at any point
  before T015

## Implementation Strategy

**MVP = Phase 1 + Phase 2 + Phase 3 (US1).** That stops the severe failure — the legend
destroying the 80-column layout — and is independently shippable. US2 and US3 then upgrade
truncation from silent to honest, and prove the policy scales with width.

**Do not skip Phase 6.** T014/T015 are what convert this from a fix into a guarantee, and
T016/T017 are gate-enforced: the spec 047 UI authority gate blocks any PR touching
`internal/tui/` without an amended design reference, and the spec 069 pr gate blocks on
stale grounding with no bypass.

**Review artifact**: per the project's TUI design loop, this change is reviewed as the
before/after diff of the 21 regenerated frames — not as a prose description of one.
