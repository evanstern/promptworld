# Tasks: Stage-shaped TUI layout defaults

**Input**: Design documents from `/specs/066-stage-defaults/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/stage-defaults-table.md, quickstart.md

**Tests**: included — the spec's success criteria (SC-001..SC-005) are test-anchored and the parity contract's enforcer IS a test.

**Organization**: tasks grouped by user story; US1 (stage-resolved boot) and US2 (pre-ladder identity) are both P1 — US2's golden-frame test is the regression net for everything US1 wires, so its capture task runs FIRST (Foundational).

## Phase 1: Setup

- [x] T001 Create `internal/tui/stagedefaults.go` skeleton: governed-surface id constants (lesson-row, guardian-strip, villager-strip, exercise-tab, incident-vocabulary, systems-tab, guardian-console, help-guardian-section, ceremony, postmortem), posture value type, and the `resolve(stage string, hasScenario bool)` signature returning the starting visible set (data-model.md entities), compiling but unwired.

## Phase 2: Foundational (blocking all user stories)

- [x] T002 Capture the PRE-FEATURE pre-ladder golden frames in `internal/tui/stagedefaults_test.go`: render a representative pre-ladder world frame corpus at fixed sizes through the existing pipeline and commit the goldens BEFORE any wiring lands — this is US2's byte-identity baseline (SC-002) and the regression net for every later task.
- [x] T003 Fill the code table in `internal/tui/stagedefaults.go` mirroring `docs/design/tui/patterns/stage-defaults.md` cell-for-cell (surface × stage-1..4, pre-ladder, narrow; the page's cell vocabulary).
- [x] T004 [P] Write the parity sweep test in `internal/tui/stagedefaults_test.go` per contracts/stage-defaults-table.md: parse the authority page's "Per-surface stage defaults" markdown table (TestCatalogSweep precedent, path-relative read) and assert cell-for-cell parity with the code table; unknown rows/columns on either side fail.
- [x] T005 [P] Implement `resolve()` in `internal/tui/stagedefaults.go`: stage column selection; `""`/unrecognized stage → pre-ladder union (fail-open, research R3); world-shaped exercise-tab axis and stage-keyed incident vocabulary resolved independently (FR-006); absent-surface tolerance (a row with no built surface is inert). Unit tests alongside in `internal/tui/stagedefaults_test.go` covering all four stages, pre-ladder, unrecognized stage, scenario × stage combinations.

## Phase 3: User Story 1 — A stage-1 player boots into the focused layout (P1) 🎯 MVP

**Goal**: the starting visible set is stage-resolved at boot; every surface stays reachable with stage-independent content.

**Independent test**: boot each stage; assert the frame matches the authority table's column; reach every non-default surface via help/solo/pull path.

- [x] T006 [US1] Wire the resolved starting set into `internal/tui/layout.go` as the `rowBudget` starting wants (lesson row, strips) WITHOUT touching the fold order or `bodyMin` logic (research R2); pre-ladder resolution must reproduce today's unconditional wants exactly.
- [x] T007 [US1] Wire tab/page presence and section defaults in `internal/tui/views.go` (exercise tab world-shaped presence; systems tab; incident forecast/fog vocabulary selection) and `internal/tui/help.go` (stage-variant guardian section default), reading only the resolved set — never capability machinery (FR-007).
- [x] T008 [P] [US1] Per-stage starting-set frame tests in `internal/tui/stagedefaults_test.go`: boot frames for stage-1..4 match the authority table's columns exactly (SC-001), including scenario and non-scenario worlds.
- [x] T009 [P] [US1] Reachability sweep test in `internal/tui/stagedefaults_test.go`: every governed surface × every stage reaches full content via help overlay / solo view / pull path without a stage change, and the reached content is stage-independent (SC-003, FR-002).

## Phase 4: User Story 2 — Pre-ladder worlds are untouched (P1)

**Goal**: pre-ladder worlds render byte-identical to the pre-feature layout; unrecognized stages fail open.

**Independent test**: golden-frame byte comparison against T002's baseline.

- [x] T010 [US2] Assert the T002 pre-feature goldens still match post-wiring: pre-ladder worlds' frames byte-identical across the corpus (SC-002); add the unrecognized-stage case asserting the same fail-open frames (FR-003). Existing `internal/tui/layout_test.go` fold tests must pass UNMODIFIED (SC-004) — if any needs edits, the wiring is wrong, not the test.

## Phase 5: User Story 3 — Surfaces arrive with the stage (P2)

**Goal**: live stage change re-resolves defaults; explicit in-session toggles win; new surfaces announce exactly once.

**Independent test**: drive a stage unlock with the TUI attached; observe re-resolution, announcement, toggle preservation.

- [x] T011 [US3] Stage-change re-resolution in `internal/tui/tui.go`: diff the stage id between status snapshots (existing `consoleStage`/`currentStage` plumbing, research R4) and re-resolve the starting set on change; takeovers (ceremony/postmortem) remain layout-independent (FR-008).
- [x] T012 [US3] Session-only `surfaceOverrides` in `internal/tui/tui.go`/`internal/tui/stagedefaults.go`: record explicit player visibility toggles for governed surfaces; re-resolution never overwrites an overridden surface; overrides are never persisted (data-model.md rules). Unit tests for override-vs-re-resolution precedence.
- [x] T013 [US3] Route stage-driven surface appearance through the existing first-occurrence lesson machinery in `internal/tui/lessons.go` (spec 055 catalog): newly default-on surfaces announce exactly once — no duplicate when the surface later appears key-driven, no suppression (SC-005, research R5). Exactly-once test in `internal/tui/lessons_test.go`.
- [x] T014 [P] [US3] Fold-pressure composition test in `internal/tui/stagedefaults_test.go`: a stage unlock on a short terminal admits the newly default-on row through `patterns/layout.md` ruling (a) only — body never dips below `bodyMin`, fold order unchanged (spec edge case).

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T015 Amend `docs/design/tui/patterns/stage-defaults.md`: `status: specified → shipped`, fill real renderer symbols (stagedefaults.go names), re-pin `verified_against` to the implementation commit; re-verify + re-pin any other `docs/design/tui/` page whose prose this feature's wiring touched; `node scripts/check-tui-design.mjs --changed` green.
- [x] T016 Full-suite validation per quickstart.md: `go test -race ./...` green in the worktree post-rebase onto fresh origin/main; quickstart's manual stage-1 / pre-ladder / unlock walkthrough spot-checked against a live world.

## Dependencies

- Phase 2 blocks everything (T002 goldens must predate wiring; T003→T004/T005).
- US1 (T006–T009) after Phase 2; T006 before T007; T008/T009 after T006+T007.
- US2 (T010) after US1's wiring (it validates the wiring changed nothing pre-ladder); its baseline (T002) is Foundational.
- US3 (T011–T014) after US1; T011 before T012/T013; T014 after T011.
- Polish (T015–T016) last.

## Parallel opportunities

- T004 ∥ T005 (different concerns, same new files — coordinate but no logical dependency).
- T008 ∥ T009 once T006/T007 land.
- T014 ∥ T012/T013 once T011 lands.

## Implementation strategy

MVP = Phases 1–4 (US1 + US2): stage-resolved boot with the pre-ladder identity
proven. US3 (live arrival) is a self-contained increment on top. The authority
page is never edited to match code — code moves to the page (contract), and any
default change starts on the page.
