# Tasks: Guardian strip — always-visible action budget line

**Input**: Design documents from `/specs/050-guardian-strip/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/guardian-strip.md, quickstart.md
**Board**: TASK-126 · one branch (`task-126-guardian-strip`), one PR

**Organization**: US1 (the strip) is the MVP; US2 (honesty rules) and US3
(fold/narrow) harden it.

## Phase 1: Setup

- [x] T001 Verify baseline in the task worktree: `go build ./...`, `go test -race ./...`, `node scripts/check-tui-design.mjs --changed` all green on a fresh `origin/main` fork

## Phase 2: Foundational

- [x] T002 Expose the charge-regen cadence read-only from internal/sim (exported const or accessor mirroring the MetatronChargeCap pattern, internal/sim/metatron.go:20-22 / research R2); unit test asserting it matches the executor's firing rule (internal/sim/executor.go:53-56)
- [x] T003 Extend `rowBudget`/`computeRows` in internal/tui/layout.go with the `Strip` row per data-model.md (pure function of (totalRows, stripVisible); strip folds last at the ruled threshold; body floor preserved); sweep tests over heights in internal/tui/layout_test.go (SC-003)

## Phase 3: User Story 1 — The budget is one glance from the verb (P1) 🎯 MVP

**Goal**: the strip renders above the minibuffer on every dock tab with live values.

**Independent Test**: quickstart.md manual steps 1–3; `go test -race ./internal/tui -run Strip`.

- [x] T004 [US1] Implement `guardianStripView(width)` in internal/tui/views.go per contract §2: charge bank (reuse the glyph-run form of views.go:1509-1512), regen forecast (`next +1 @ <game time>` via T002's cadence + existing clock formatting), standing-order count (`len(m.status.Orders)`), ` · ` separators; no faith segment
- [x] T005 [US1] Insert the strip row into the widescreen composite between body and minibuffer (views.go:284 vicinity), wired to T003's row budget; assert composite height invariance in internal/tui/views_test.go
- [x] T006 [US1] Live-update test in internal/tui/render_test.go: charge spend/regen and order add/expiry reflected next frame (fixture replica mutations; US1 AS-2)

## Phase 4: User Story 2 — The strip never lies (P2)

**Goal**: segment presence matrix exactly per contract §2.

**Independent Test**: `go test -race ./internal/tui -run StripPresence`.

- [x] T007 [US2] Implement presence rules in `guardianStripView`: full bank → regen omitted; no status snapshot → blank row; zero orders → true `👁 0`; faith never rendered (research R4); right-to-left truncation (faith→orders→regen→bank) with `…`, never wrapping past 1 row
- [x] T008 [US2] Fixture presence-matrix sweep in internal/tui/render_test.go (SC-002): {no status, partial bank, full bank, 0 orders, N orders} × width pressure — every rendered segment true, every absent segment absent

## Phase 5: User Story 3 — The strip survives pressure (P3)

**Goal**: fold-last relocation + narrow carry per layout.md rulings a/b.

**Independent Test**: quickstart.md manual steps 4–5; `go test -race ./internal/tui -run 'Fold|Narrow'`.

- [x] T009 [US3] Fold relocation: when T003's budget drops the strip, `minibufferView`'s dormant branch (views.go:1692) renders the relocated prefix form (contract §3), truncated to width; focused/busy branches byte-identical — regression test asserting all three minibuffer states with strip folded/unfolded
- [x] T010 [US3] Narrow-fallback carry: insert the strip above the minibuffer in the narrow stack (views.go:1558 vicinity) sharing `guardianStripView`; test at sub-breakpoint widths in internal/tui/views_test.go

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T011 [P] Amend docs/design/tui/panels/guardian-strip.md: `status: specified → shipped`; control-table renderer cells → real symbols (`guardianStripView` etc.); record spec 050's rulings (full-bank regen omission, pre-status blank row, right-to-left truncation, zero-orders-is-true); re-pin `verified_against`
- [x] T012 [P] Re-verify + re-pin docs/design/tui/patterns/layout.md (row-budget/fold rows now real) and docs/design/tui/panels/minibuffer.md (dormant relocation form recorded)
- [x] T013 Run gates in the worktree: `go test -race ./...` and `node scripts/check-tui-design.mjs --changed` green; fix anything surfaced
- [x] T014 Pre-PR: rebase onto fresh `origin/main`, re-run both gates post-rebase

## Dependencies & Execution Order

- Phase 2: T002 and T003 are independent [P].
- US1: T004 needs T002; T005 needs T003+T004; T006 needs T005.
- US2: T007 needs T004; T008 needs T007. Can start once T004 exists (parallel with T005/T006).
- US3: T009 needs T003+T004; T010 needs T004. Parallel with US2.
- Polish: T011/T012 [P] after behavior final; T013 → T014 last.

## Implementation Strategy

MVP = Phases 1–3. US2/US3 are hardening on the same renderer. One worktree,
subtasks as commits, one PR with code + doc amendments together.
