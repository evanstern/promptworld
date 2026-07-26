# Tasks: Village lens completion — villager strip + map condition overlays

**Input**: Design documents from `/specs/060-village-lens/`
**Prerequisites**: spec.md (the authored design pages are the detailed contract: panels/villager-strip.md, panels/map.md Wave-5 stub)
**Board**: TASK-129 · one branch (`task-129-village-lens`), one PR

Note: this Wave-5 slice is small and page-driven; plan-level decisions live
in the spec's standing resolutions — no separate plan/research artifacts
(the spec IS the plan; precedent: the design pages carry the mockups,
budgets, and fold rules verbatim).

## Phase 1: Setup

- [x] T001 Verify baseline in the task worktree on fresh `origin/main`: build + `go test -race ./...` + `node scripts/check-tui-design.mjs --changed` green

## Phase 2: User Story 1 — villager strip (P1)

- [x] T002 [US1] `villagerStripView(width)` in internal/tui/views.go per panels/villager-strip.md: `N villagers` + roster-ordered name-initial glyph run reusing the map's exact per-state styling; end-drop overflow with `…N`; display-only
- [x] T003 [US1] Row-budget integration: extend the spec-050 fold machinery in internal/tui/layout.go — strip row under the header, folds SECOND (after map legend, before later chrome) to the `[N villagers]` header badge (pages/home.md form); narrow renders badge only; height/width sweep tests in internal/tui/layout_test.go + views_test.go (SC-003)
- [x] T004 [US1] Roster-parity tests in internal/tui/render_test.go: strip matches the villagers-tab roster 1:1 across awake/asleep/dead fixtures (SC-001)

## Phase 3: User Story 2 — map condition overlays (P1)

- [x] T005 [US2] Overlay derivation in internal/tui/views.go (renderMapGrid agents loop): needs-critical (roster critical-threshold parity), suppressed-mind (latest decision-trace outcome suppressed — existing client projection), dying-fire (within the dying-fuel window of `FuelUntil`, steady warn style); priority needs-critical > suppressed-mind; legend entries for the new styles
- [x] T006 [US2] Fixture render suite (FR-004/SC-002): every overlay state × set/clear × priority × color-profile distinguishability (reuse the family-tint test discipline)

## Phase 4: Polish & Cross-Cutting Concerns

- [x] T007 [P] Design-page amendments: villager-strip.md specified → shipped (symbols; overflow/display-only rulings); map.md Wave-5 stub → real overlay rows + priority rules + look-cursor deferral ruling; layout.md + home.md re-verified; re-pin all touched pages
- [x] T008 Run gates: `go test -race ./...`, `node scripts/check-tui-design.mjs --changed`, `node scripts/check-merge-drift.mjs pr`, gofmt/vet clean
- [x] T009 Pre-PR: rebase onto fresh `origin/main`, re-run gates post-rebase

## Dependencies & Execution Order

T002 → T003 → T004; T005 → T006 [P with US1 after T002]; T007 → T008 → T009.
