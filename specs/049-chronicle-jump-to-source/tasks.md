# Tasks: Chronicle jump-to-source + input-parity retrofit start

**Input**: Design documents from `/specs/049-chronicle-jump-to-source/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/jump-to-source.md, quickstart.md
**Board**: TASK-124 · one branch (`task-124-chronicle-jump-to-source`), one PR

**Organization**: grouped by user story; US1 (⏎ jump) is the MVP; US2 (mouse)
and US3 (actions bar) are additive on the same foundation.

## Phase 1: Setup

- [x] T001 Verify baseline in the task worktree: `go build ./...` and `go test -race ./...` green on a fresh `origin/main` fork; `node scripts/check-tui-design.mjs --changed` green (no changes yet)

## Phase 2: Foundational (blocking prerequisites for all stories)

- [x] T002 Extract the live-wanderer centroid computation from `renderMapGrid` (internal/tui/views.go:444-461) into a shared `Model` helper (e.g. `wandererCentroid()`) used by both the renderer and the new camera writer — behavior byte-identical, existing render tests stay green
- [x] T003 Add `centerCameraOn(x, y int)` to `Model` in internal/tui/tui.go: sets `panX/panY = target − wandererCentroid()` per research R1; unit test in internal/tui/tui_test.go proving pan-equivalence (map title flips to panned, `c` recenter restores follow, clamping at edges matches manual pan)
- [x] T004 Implement `resolveSubject(e store.Event) (name string, x, y int, ok bool)` beside the digest catalog in internal/tui/digest.go per contract §2: primary-actor field per event type → live replica position (alive only), else explicit payload coordinates, else not locatable; bounded to known top-level fields (never scans `world.migrated` state) — table-driven tests in internal/tui/digest_test.go covering actor-alive, actor-dead-with-payload-position, unlocatable, and multi-participant (speaker wins) cases

## Phase 3: User Story 1 — Jump from an event to its subject on the map (P1) 🎯 MVP

**Goal**: `⏎` on a selected chronicle event centers the map on its subject; unlocatable events are honest no-ops.

**Independent Test**: quickstart.md manual steps 1–4; `go test -race ./internal/tui -run Jump`.

- [x] T005 [US1] Wire `handleInspectKey`'s `enter` case (internal/tui/tui.go:986) to jump: resolve subject via T004; on ok, `centerCameraOn`; on !ok, no state change (the actions bar is the visible explanation per contract §1); keep the key swallowed (handled=true) in both branches
- [x] T006 [US1] Narrow-fallback landing (FR-007): on successful jump when the layout is the narrow single-pane fallback, switch the visible pane to the map view preserving paused clock, chronicle selection, and detail scroll; test the round-trip (jump → map pane → back to chronicle → selection intact) in internal/tui/tui_test.go
- [x] T007 [US1] Catalog-sweep totality test (SC-002) in internal/tui/digest_test.go: for every event type in the digest catalog, `resolveSubject` + `detailActions` yield exactly jump-or-hint — no panics, no empty outcomes, bounded work on oversized payloads (`world.migrated` fixture)

## Phase 4: User Story 2 — Click a chronicle line to select and jump (P2)

**Goal**: first mouse-bound control; click = select + jump while paused; inert while running; keyboard untouched.

**Independent Test**: quickstart.md manual step 5; `go test -race ./internal/tui -run Mouse`.

- [x] T008 [US2] Enable mouse reporting: add `tea.WithMouseCellMotion()` to the `tea.NewProgram` call in cmd/promptworld/commands.go:745 (research R2)
- [x] T009 [US2] Record chronicle render geometry (data-model `chronHitRegion`): chronicle body renderers in internal/tui/views.go fill panel rectangle + per-row event index each `View()` (wrapped rows map to their event; invalid when chronicle not rendered); overwrite/invalidate rules per data-model.md
- [x] T010 [US2] Handle `tea.MouseMsg` in `Update` (internal/tui/tui.go): left-button release inside a valid `chronHitRegion` row while paused → set `chronSelected` to that row's event (resetting detail scroll, same as `j`/`k`) and apply T005's jump rules; running clock or out-of-region clicks → no-op (contract §1); tests in internal/tui/tui_test.go for click-select-jump, running-clock no-op, border/detail-pane click no-op
- [x] T011 [US2] Keyboard-regression guard (FR-005): test asserting every existing inspect/global key handling is byte-identical with mouse enabled (reuse focus_test.go's exhaustive-key harness pattern)

## Phase 5: User Story 3 — The seam advertises itself honestly (P3)

**Goal**: the detail pane's reserved slot becomes the live affordance/hint bar.

**Independent Test**: quickstart.md manual steps 2 & 4 (bar text); `go test -race ./internal/tui -run DetailAction`.

- [x] T012 [US3] Populate `detailActions(e store.Event)` (internal/tui/tui.go:1166): return exactly one action per event — `⏎ jump to <name> (x,y)` when locatable (live-resolved via T004), `no location for this event` when not; never nil (data-model totality)
- [x] T013 [US3] Render the actions bar: replace the `"[future: actions]"` literal (internal/tui/views.go:1216) with the `detailActions` label for the selected event, preserving the pane's row budget and truncation discipline; rendering-state tests in internal/tui/render_test.go (locatable, unlocatable, truncated-width)

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T014 [P] Amend docs/design/tui/panels/chronicle.md: jump-to-source control-table row `unbuilt` → real renderer symbols with `⏎` · click bindings; parity-rollout note records the corpus's first shipped mouse target; detail-pane prose/mockup `[future: actions]` → actions bar; re-pin `verified_against` (research R7)
- [x] T015 [P] Amend docs/design/tui/patterns/keymap.md: inspect-mode `⏎` row `reserved — no-op` → jump action; parity doctrine rule 3's "zero controls have a real mouse target" statement updated to name this control; migration note; re-pin `verified_against`
- [x] T016 Run gates in the worktree: `go test -race ./...` and `node scripts/check-tui-design.mjs --changed` both green; fix anything they surface
- [x] T017 Pre-PR: rebase onto fresh `origin/main`, re-run both gates post-rebase, open the PR per the runbook's standard footer

## Dependencies & Execution Order

- Phase 2 blocks everything: T002 → T003 (same camera seam); T004 independent of T002/T003 → **T002+T004 can run in parallel**.
- US1 (T005–T007): T005 needs T003+T004; T006 needs T005; T007 needs T004+T012's totality shape (write against `resolveSubject` first, extend when T012 lands).
- US2 (T008–T011): T008 anytime [P]; T009 independent [P]; T010 needs T005+T009; T011 after T010.
- US3 (T012–T013): T012 needs T004; T013 needs T012. US3 can proceed in parallel with US2 after Phase 2.
- Polish: T014/T015 [P] once behavior is final; T016 → T017 last.

## Implementation Strategy

MVP = Phases 1–3 (keyboard jump with honest no-op). US2 and US3 layer on
without touching US1's semantics. Single worktree, subtasks as commits, one
PR carrying code + design-doc amendments together (gate rule 1).
