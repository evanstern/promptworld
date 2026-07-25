# Tasks: `?` help overlay in the TUI (every world)

**Input**: Design documents from `/specs/045-tui-help-overlay/`

**Prerequisites**: plan.md, spec.md, research.md (R1–R8), data-model.md,
contracts/help-content.md

**Tests**: included — SC-003's keymap sweep and the exact-height invariant are explicit
mechanical gates.

**Organization**: grouped by user story; all work lands on one `task-116` branch in
`.worktrees/task-116` (one task, one PR).

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

- [ ] T001 Create worktree `.worktrees/task-116` from fresh origin/main; confirm
      `go build ./... && go test ./internal/tui` green at base

## Phase 2: Foundational (blocking prerequisites)

- [ ] T002 Extract the map legend into a shared glyph table (`glyphEntry` rows) in
      internal/tui/views.go (or help.go), and render `renderMapGrid`'s legend line
      (views.go:615-617) from it — behavior-identical (existing render tests stay
      green); this is the FR-005 anti-drift substrate
- [ ] T003 [P] Add `helpOpen`/`helpSection`/`helpTier`/`helpMode`/`helpScroll` fields
      to Model in internal/tui/tui.go (state only, no behavior yet)

**Checkpoint**: shared glyph table live; model carries overlay state.

## Phase 3: User Story 1 — `?` answers "what can I press right now" (P1) 🎯 MVP

**Goal**: `?` from every non-entry mode opens mode-frozen key help, basic tier first,
advanced tier reachable; dismissal releases exactly one layer, side-effect-free.

**Independent Test**: from each documented mode press `?`; basic page lists that mode's
keys first; advanced tier reachable; dismissal restores prior screen/focus/selection.

- [ ] T004 [US1] Create internal/tui/help.go: per-mode key tables (global/home,
      minibuffer, inspect, villagers roster/detail, solo/narrow) with basic/advanced
      tier per row, sourced from docs/design/tui/patterns/keymap.md (basic ≈
      footer-hinted set)
- [ ] T005 [US1] Overlay key dispatch at the head of `handleKey` (after ctrl+c guard
      tui.go:524, before mbFocused): `?` opens when `!m.mbFocused` (freezing
      `helpMode`); while open — tier-advance key, section navigation, `J`/`K` scroll
      (pager idiom tui.go:815-821), `esc`/`?` dismiss; all other keys inert —
      internal/tui/tui.go (R1; contracts layering rules)
- [ ] T006 [US1] `helpPanelView` body replacement in `widescreenView` (solo-zoom
      precedent, views.go:220-239) and `narrowView` (views.go:81-97), chrome kept,
      styleBox/clipContent sizing, pager overflow footer — internal/tui/views.go +
      help.go (R2, R4)
- [ ] T007 [P] [US1] Footer hints gain `· ? help` in every `footerView` branch
      (views.go:197-216) and solo titles if applicable; update the keymap doc's footer
      table + overlay keys — internal/tui/views.go,
      docs/design/tui/patterns/keymap.md (R5, FR-011)
- [ ] T008 [US1] Tests in internal/tui/help_test.go + extensions: `?` opens from every
      mode incl. narrow fallback; minibuffer `?` still types (focus-contract rule-4
      sweep stays green); tier advance; esc releases exactly one layer
      (villagers-detail-under-overlay case); dismissal restores selection/scroll;
      `TestWidescreenViewExactHeight` gains help/help-advanced states; keymap sweep —
      every listed binding handled, every handled binding listed once (SC-003)

**Checkpoint**: US1 demonstrable per quickstart §2 (keys portion) — MVP.

## Phase 4: User Story 2 — the screen explained (P2)

**Goal**: walkthrough covering header anatomy (all conditional badges), full glyph
legend, dock tabs.

**Independent Test**: overlay walkthrough covers every header element/badge, every map
glyph, every dock tab, one concept per line, scrollable at small sizes.

- [ ] T009 [US2] Walkthrough content in help.go: `headerAnatomy` rows (every headerView
      element + badges: running/PAUSED, governed-speed suffix, [degraded], [llm: …],
      [suppressed: …], disconnected — plus ENDED if spec 044 has landed by rebase
      time), glyph page rendered from the T002 shared table, `dockTabEntry` rows from
      paneNames/dockTabKey — internal/tui/help.go (R3)
- [ ] T010 [US2] Tests: content-presence assertions (views_test.go style) — every
      current header badge string has a walkthrough row; glyph page enumerates exactly
      the shared table (drift impossible by construction, assert anyway); dock tabs
      complete; scroll reaches all content at 80x24 — internal/tui/help_test.go

**Checkpoint**: a newcomer can decode the whole screen from the overlay.

## Phase 5: User Story 3 — the floor holds with no angel (P3)

**Goal**: byte-identical overlay on no-LLM worlds; live-model conditions irrelevant.

**Independent Test**: nil-status/nil-replica model exercises US1/US2 tests verbatim.

- [ ] T011 [US3] Tests: overlay renders and navigates with nil `status` and nil
      `replica` (tui.go:151-153 tolerance); content bytes identical to a
      status-carrying model; no code path in help.go reads LLM/status state (enforce
      by construction review + test) — internal/tui/help_test.go (R6, SC-004)

**Checkpoint**: charter-independent floor proven.

## Phase 6: User Story 4 — pushed lessons findable again (P4)

**Goal**: the pull-reference seam ships as structure + contract.

**Independent Test**: reference section navigable; adding a `helpLesson` entry requires
no structural change.

- [ ] T012 [US4] `helpLesson{id,title,body}` table (empty) + contract comment
      referencing contracts/help-content.md; lessons section renders entries or the
      placeholder line — internal/tui/help.go
- [ ] T013 [US4] Test: inject a fixture lesson entry in-test → it renders and is
      navigable with zero non-content changes (SC-006) — internal/tui/help_test.go

**Checkpoint**: all four stories functional.

## Phase 7: Polish & Cross-Cutting

- [ ] T014 Full gate: `go test ./...`, gofmt/vet; SC-005 soak test (open/close loop
      causes zero model-state diff outside help fields)
- [ ] T015 Rebase over main; if spec 044's TUI changes (ENDED token, grave glyph,
      footer hints) have landed, reconcile: ENDED row in headerAnatomy, grave glyph
      arrives via the shared table automatically; re-run sweeps (plan.md Known
      collision)
- [ ] T016 [P] Run quickstart.md live walkthrough (§2–§4); record outcomes in the PR
      description; note post-merge wiki-update obligation (Principle IV)

## Dependencies & Execution Order

- Setup → Foundational → US1 → US2 → US3 → US4 → Polish. US2 depends on T002 (table)
  and US1's panel host. US3/US4 are small and ride US1's structures.
- Parallel: T003 with T002; T007 alongside T005/T006; T010/T011/T013 parallel per
  test file once content lands.

## Implementation Strategy

**MVP = US1** (the universal `?` with tiered keys). US2 adds the walkthrough content,
US3 is a proof obligation, US4 a seam. Single Sonnet-tier slice per plan.md
Constitution Check V (single-package view code); escalate only if the dispatch-head
refactor unexpectedly destabilizes the focus contract.
