# Tasks: Takeover surfaces — ceremony + postmortem

**Input**: Design documents from `/specs/056-takeover-surfaces/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/takeovers.md, quickstart.md
**Board**: TASK-127 · one branch (`task-127-takeover-surfaces`), one PR

## Phase 1: Setup

- [ ] T001 Verify baseline in the task worktree on fresh `origin/main`: build + `go test -race ./...` + `node scripts/check-tui-design.mjs --changed` green; note which of TASK-121/125/119 have merged (skin lookup, consoleCard seam, scenario status fields — research R4/R6 sequencing postures)

## Phase 2: Foundational

- [ ] T002 Takeover owner state machine per data-model.md (internal/tui/tui.go): enum + ceremonyDeferred + per-attach postmortemDismissed; transitions incl. postmortem-always-wins, same-kind replacement, connect auto-open, `p` (ended-only), esc; precedence interleaving tests (SC-003) in internal/tui/tui_test.go
- [ ] T003 Shared report-card renderer `reportCardView(def, facts, mode, width)` per contract §4 (internal/tui/views.go): concluded/live marker modes over rubric terms + backing references (exercise-panel gauge vocabulary); consoleCard wrapper + seam-composition test (spec 053 interface); three-site row-equivalence test (SC-005) in internal/tui/render_test.go

## Phase 3: User Story 1 — Postmortem takeover (P1)

- [ ] T004 [US1] postmortemView per contract §2 (internal/tui/views.go): title + run-end line; morgue rows derived at open from replica facts (death ledger + closest charter observation, unknown-honest on ring rotation — research R3); body-replacement slot (help precedent), exact-height in both layouts
- [ ] T005 [US1] Scored-run card: render reportCardView (concluded) above the morgue rows ONLY when scenario exercise + rubric facts are present (research R4); ambient/scored/missing-data matrix tests (SC-002)
- [ ] T006 [US1] Dismiss/replay wiring: esc; `p` global ended-only; auto-open on connect to ended worlds; ENDED-posture regression tests (clock keys stay read-only after dismissal)

## Phase 4: User Story 2 — Ceremony takeover (P1)

- [ ] T007 [US2] ceremonyView per contract §3: skin-resolved title + D6 authorship chapter (skin lookup per research R6; add any missing tokens to the default table per the skin contract §4) + reportCardView (instrument); q-detach with the D13 "world keeps running" framing; fixture-event tests (production emission may be TASK-119's, either way fixtures drive the tests)
- [ ] T008 [US2] Replay surfaces: `?` overlay ceremony-replay section (internal/tui/help.go, spec 045 content contract amended deliberately) listing earned unlocks from replica facts, re-rendering stored content; test that a dismissed/deferred ceremony is retrievable (SC-004); `promptworld stages` unchanged (verify only)

## Phase 5: Polish & Cross-Cutting Concerns

- [ ] T009 [P] Design-page amendments (research R7): ceremony.md + postmortem.md → shipped w/ real symbols; help.md replay entry; keymap.md (`p`, takeover keys, parity gaps); guardian-console.md seam note names reportCardView; re-pin every touched page to the final code commit
- [ ] T010 Run gates: `go test -race ./...`, `node scripts/check-tui-design.mjs --changed`, gofmt/vet clean
- [ ] T011 Pre-PR: rebase onto fresh `origin/main` (expect tui.go/views.go/help.go/keymap.md conflicts with Lane-2 siblings; take main's side for anything not intentionally changed), re-run all gates post-rebase

## Dependencies & Execution Order

- T002 → everything; T003 after T002 [P with T004's shell].
- US1: T004 → T005 → T006. US2: T007 after T003; T008 after T007. US1 ∥ US2 after Phase 2.
- Polish: T009 → T010 → T011.

## Implementation Strategy

State machine + renderer first (they are the family's shared spine), then
the two takeovers in parallel. One worktree, one PR; PR body calls out the
precedence rules and the D5 one-renderer contract.
