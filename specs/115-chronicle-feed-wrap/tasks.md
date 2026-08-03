# Tasks: Chronicle Raw Feed Wrapping

**Feature**: spec 115 · **Branch**: `task-195-polish-session-1` · **Board task**: TASK-195

**Input**: [spec.md](./spec.md), [plan.md](./plan.md), [research.md](./research.md),
[data-model.md](./data-model.md), [contracts/feed-wrap.md](./contracts/feed-wrap.md),
[quickstart.md](./quickstart.md)

**Tests**: included. Not TDD-by-request but repo convention (tests alongside code) plus SC-006,
which can only be proven by the pre-existing suite passing **unmodified**.

---

## Phase 1: Setup

- [ ] T001 Confirm the frame harness is green before any edit: run `node .claude/skills/tui-frames/scripts/check-frames.mjs --check` from the worktree root and record the result
- [ ] T002 Capture the pre-change baseline frame for comparison: copy `docs/design/tui/frames/mid-game__solo__160x50.txt` to a scratch path outside the repo

---

## Phase 2: Foundational (blocking prerequisites)

**Blocks every user story.** The wrap budget domain must be extended before any call site can ask
for unbounded wrap, and both renderers must move together or the two paths diverge.

- [ ] T003 Declare the `wrapUnbounded = 0` constant with a domain comment (`0` unbounded / `1` truncate / `>1` capped) beside the wrap renderers in `internal/tui/grammar.go`
- [ ] T004 Extend `wrapOrTruncatePlain` in `internal/tui/grammar.go`: change the `maxWrap <= 1` truncate guard to `maxWrap == 1`, remove the upward normalization that would clamp `0` to `1`, and let `0` fall through to the wrap branch with no line cap
- [ ] T005 Extend `styleWrapLine` in `internal/tui/grammar.go` identically: `maxWrap == 1` truncates, `0` wraps unbounded, `> 1` keeps today's cap-and-ellipsis behavior
- [ ] T006 Add the `indent int` parameter to both `wrapOrTruncatePlain` and `styleWrapLine` in `internal/tui/grammar.go`: continuation lines (every line after the first) are prefixed with `indent` spaces and wrap against a text width of `width - indent`
- [ ] T007 Declare `minWrapTextWidth = 24` in `internal/tui/grammar.go` and apply the all-or-nothing fallback in both renderers: when `width - indent < minWrapTextWidth`, set `indent = 0` rather than reducing it
- [ ] T008 [P] Unit-test the extended budget domain in `internal/tui/grammar_test.go`: `maxWrap == 1` truncates with `…`, `maxWrap == 3` caps at three lines with `…`, `maxWrap == 0` emits every line with no ellipsis
- [ ] T009 [P] Unit-test plain/styled equivalence in `internal/tui/grammar_test.go`: for the same input, width, budget and indent, `wrapOrTruncatePlain` and `styleWrapLine` produce identical characters
- [ ] T010 [P] Unit-test the narrow fallback in `internal/tui/grammar_test.go`: an indent leaving under 24 columns of text yields `indent = 0`, never a reduced indent, and no line exceeds `width`

**Checkpoint:** renderers support unbounded wrap and hanging indent; no view behavior has changed yet.

---

## Phase 3: User Story 1 — A thought can be read to its end (P1) 🎯 MVP

**Goal**: long summaries wrap instead of truncating, at every width a player reads at.

**Independent test**: render a fixture with a long prose event at a full-width size; the complete
summary is present across multiple lines with no `…`. Passes without any part of US2.

- [ ] T011 [US1] Switch the chronicle wrap budget from `1` to `wrapUnbounded` for the solo and wide-dock paths in `dockTabContent` in `internal/tui/views.go`, keeping the `width < 60` narrow-dock cap at `3`
- [ ] T012 [US1] Switch the narrow-fallback `chronicleView` in `internal/tui/views.go` from `1` to `wrapUnbounded`
- [ ] T013 [US1] Add one `agent.thought` event with prose long enough to wrap at 160 columns to `midGameFeed` in `internal/tui/fixtures.go`, positioned near the tail so it lands in the visible window at all four committed sizes
- [ ] T014 [US1] Add one `social.conversation_turn` event with prose long enough to wrap at 160 columns to `midGameFeed` in `internal/tui/fixtures.go`, likewise near the tail
- [ ] T015 [US1] Test in `internal/tui/tui_test.go` that a long summary renders in full with no `…` at a full-width size, and that a summary which fits still renders on exactly one line
- [ ] T016 [US1] Test in `internal/tui/tui_test.go` that wrapping breaks between words and never mid-word, except for a single word longer than the text column

**Checkpoint:** thoughts and conversations are readable end to end. Continuation lines still start
at column zero — that is US2's job.

---

## Phase 4: User Story 2 — The feed still reads as a table (P2)

**Goal**: continuation lines begin at the summary column, so the tick/time/type rail stays unbroken.

**Independent test**: render a wrapped row; every continuation line begins at the same column as
the first line's summary, and the columns left of that point are blank.

- [ ] T017 [US2] Pass the row's own prefix rune-width as the `indent` argument from `renderChronicleRow` in `internal/tui/views.go` into both the `styleWrapLine` and `styleWholeLine` paths, so ordinary rows and the alert/labeled-voice tiers align identically
- [ ] T018 [US2] Test in `internal/tui/tui_test.go` that every continuation line of a wrapped row starts at that row's summary column and contains no tick, time, or type text
- [ ] T019 [US2] Test in `internal/tui/tui_test.go` that the indent tracks recomputed column widths: two windows whose widest tick or longest type differ produce different indents, both correct for their own window
- [ ] T020 [P] [US2] Test in `internal/tui/tui_test.go` that an alert-tier row (e.g. `agent.died`) and a labeled-voice row wrap with the same indent as an ordinary row, keeping their whole-line styling on every physical line
- [ ] T021 [P] [US2] Test in `internal/tui/tui_test.go` that a selected wrapped row renders reversed on every one of its physical lines (R6 — currently unexercised)

**Checkpoint:** the feed reads as a table again, with prose in one aligned block.

---

## Phase 5: User Story 3 — Narrow panes degrade sensibly (P3)

**Goal**: the indent yields before the text column collapses; nothing overflows or crashes.

**Independent test**: render a wrapped row at the narrowest supported size; the text column keeps a
usable width and no line exceeds the pane.

- [ ] T022 [US3] Test in `internal/tui/tui_test.go` that at a pane width where the indent would leave under 24 columns, rows wrap at full width with zero indent
- [ ] T023 [US3] Test in `internal/tui/tui_test.go` that no emitted feed line exceeds the pane width at 80, 112, 113 and 160 columns
- [ ] T024 [US3] Test in `internal/tui/tui_test.go` that the text column is never zero or negative and rendering never panics at degenerate widths, including a single word longer than the text column

**Checkpoint:** all three stories complete and independently proven.

---

## Phase 6: Row budget and evidence

- [ ] T025 Verify — do not rebuild — the physical-line trim in `chronicleRawBody` in `internal/tui/views.go` still honors the row budget with multi-line rows, and test in `internal/tui/tui_test.go` that the body never exceeds its allotted rows and the newest row stays visible at the bottom
- [ ] T026 Regenerate the frame matrix with `go run ./cmd/promptworld frames --dump` and commit the result; never hand-edit a `.txt` frame
- [ ] T027 Confirm from the regenerated frames that at least one committed frame shows a visibly wrapped, indented row (SC-007), and diff against the T002 baseline to confirm the fixture churn is only the expected tick/window shift

---

## Phase 7: Polish and cross-cutting

- [ ] T028 Amend `docs/design/tui/panels/chronicle.md` to describe the wrap budget, the hanging indent, and the narrow-pane fallback, and re-pin it (FR-013)
- [ ] T029 Run `node scripts/check-tui-design.mjs --changed` and resolve any finding until it exits 0
- [ ] T030 Run the full suites: `go test ./...`, `gofmt -l internal/`, `go vet ./internal/...`, and `node --test scripts/check-merge-drift.test.mjs`
- [ ] T031 Confirm SC-006 by checking that no pre-existing `internal/tui` test needed modification; if any did, treat it as a defect in the change rather than a stale test and fix the change
- [ ] T032 Record on TASK-195 the frame-churn scope from T027, so the PR reviewer knows why many `mid-game` frames changed

---

## Dependencies

```
Phase 1 (T001-T002)
   ↓
Phase 2 (T003-T010)  ← blocks everything; renderers move together
   ↓
Phase 3 US1 (T011-T016)  🎯 MVP — shippable alone
   ↓
Phase 4 US2 (T017-T021)  ← needs the indent parameter from T006
   ↓
Phase 5 US3 (T022-T024)  ← needs the fallback from T007 and the indent from T017
   ↓
Phase 6 (T025-T027)      ← frames must be regenerated after all rendering changes
   ↓
Phase 7 (T028-T032)
```

**Story independence**: US1 ships alone and delivers the whole complaint's value. US2 and US3
depend on Phase 2's parameter but not on each other's tests. US3's fallback is only observable
once US2 supplies a non-zero indent.

## Parallel opportunities

- **Phase 2**: T008, T009, T010 are three separate test functions in one file — parallel-safe in
  authoring, though they land in the same commit.
- **Phase 3**: T013 and T014 are two independent fixture events; T015 and T016 are separate tests.
- **Phase 4**: T020 and T021 touch different behaviors and are independent of each other.
- **Not parallel**: T004/T005/T006/T007 all edit the same two renderers and must be sequenced, and
  T026 must follow every rendering change.

## Implementation strategy

**MVP is Phase 2 + Phase 3 (US1).** That alone makes thoughts and conversations readable, which is
the entire complaint. Stopping there would leave continuation lines at column zero — worse-looking
but not worse-informed.

Ship in phase order. The frame regeneration (T026) is deliberately late: running it before the
rendering is final produces churn that must be redone, and a frame file is evidence, never a draft.

**Total**: 32 tasks — 8 foundational, 6 (US1), 5 (US2), 3 (US3), 3 budget/evidence, 5 polish, 2 setup.
