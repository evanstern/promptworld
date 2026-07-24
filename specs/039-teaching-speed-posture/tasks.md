# Tasks: Teaching-World Speed Posture (Calibrated Soft Cap)

**Input**: Design documents from `specs/039-teaching-speed-posture/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/posture.md, quickstart.md

**Tests**: Included — project convention is table tests beside code (plan.md Testing),
and spec 035/037 precedent ships byte-identity regression tests with every additive
wire field.

**Organization**: Grouped by user story; US1 (calibrated teaching default) is the MVP.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

- [X] T001 Create worktree `.worktrees/task-78` on branch `task-78-teaching-speed-posture` from fresh `origin/main`; confirm `go build ./...` and `go test ./...` green at baseline

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: the manifest marker and the exported rung arithmetic — every story reads both

- [X] T002 [P] Add `Teaching bool` with `json:"teaching,omitempty"` to `Manifest` in internal/world/world.go (doc comment: teaching posture marker, decision-6/spec 039); add a `SetTeaching(dir string, on bool) error` read-modify-write helper beside Load/Create
- [X] T003 [P] Extract `MaxSafeSpeed(class string, secPerPt float64) float64` in internal/cognition/horizon.go from HorizonSummary's maxOK loop (highest horizonLadder rung where `Route(...).Allow`; 0 when none); refactor `HorizonSummary` to call it — output strings byte-identical
- [X] T004 [P] Table tests: manifest round-trip (absent field ⇒ false; non-teaching world.json byte-identical on rewrite; old manifest loads) in internal/world/world_test.go; `MaxSafeSpeed` rungs at 5.0/17.0/20.0/1000.0 s/pt + HorizonSummary unchanged-output regression in internal/cognition/horizon_test.go

**Checkpoint**: marker + arithmetic exist; all stories can proceed

---

## Phase 3: User Story 1 - Teaching world runs at the fastest honest speed by default (Priority: P1) 🎯 MVP

**Goal**: teaching worlds boot to the posture rung derived live from the planner-serving provider's estimate; recalibration moves it at next boot; non-teaching worlds untouched

**Independent Test**: quickstart.md §2 — seed calibration.json at 17.0 s/pt ⇒ boot defaults 16x (recorded `clock.speed_set`); re-seed 5.0 ⇒ 32x; plain world unchanged

### Implementation for User Story 1

- [X] T005 [US1] Daemon boot posture default in internal/daemon/daemon.go: for `w.Manifest.Teaching` worlds with an orchestrator, after calibration seeding compute rung via `cognition.MaxSafeSpeed("planner", est)` from `orch.EstimateForKind(llm.Kind("planner"))` (rung 0 ⇒ clamp to lowest capped rung), issue the loop's normal `set_speed` command so it lands as a recorded `clock.speed_set` event; print the posture line per contracts/posture.md §2 (calibrated flavor: rung + s/pt + CalibratedAt). Pure-sim teaching worlds: no-op
- [X] T006 [US1] `--teaching` flag on `promptworld new` in cmd/promptworld/commands.go threading to `world.Create` (extend Create's signature or set-after-create via `world.SetTeaching`) so the manifest carries the marker from birth
- [X] T007 [US1] Tests: boot-default table test (teaching+calibrated ⇒ posture speed event + stdout line; teaching+no-LLM ⇒ no-op; non-teaching ⇒ byte-identical boot output) in internal/daemon/daemon_test.go; `new --teaching` manifest assertion in cmd/promptworld tests beside existing cmdNew coverage

**Checkpoint**: US1 fully functional — MVP

---

## Phase 4: User Story 2 - Exceeding the posture teaches the horizon instead of blocking (Priority: P2)

**Goal**: set_speed above the posture succeeds and the reply Warning carries per-class Route arithmetic + consequence; at/below posture and non-teaching: no posture warning

**Independent Test**: quickstart.md §3 — `speed teach 32` applies AND warns with planner arithmetic verbatim; `speed teach 8` silent; plain world replies byte-identical

### Implementation for User Story 2

- [X] T008 [US2] `(*Server).postureWarning(speed clock.Speed) string` in internal/ipc/server.go beside uncalibratedWarning: teaching + orchestrator + requested rung > posture rung ⇒ for each watched class whose `Route` at the requested speed disallows, emit `Verdict.Arithmetic` verbatim + degrade consequence phrase (contracts/posture.md §3); compose with uncalibratedWarning newline-joined into the one set_speed reply Warning; widen the Warning doc comment in internal/ipc/protocol.go to name the posture case (still set_speed-only, still never blocks, max-gate untouched)
- [X] T009 [US2] Tests in internal/ipc/server_test.go (or existing warning test home): above-posture warns with exact arithmetic; at/below posture no posture text; non-teaching unchanged; uncalibrated teaching world composes both texts; speed always applied; `max` still errors

**Checkpoint**: US1 + US2 independently verifiable

---

## Phase 5: User Story 3 - Uncalibrated teaching worlds are told to calibrate (Priority: P2)

**Goal**: bootstrap-seeded planner provider ⇒ boot + speed changes prompt `promptworld calibrate`, posture presented as provisional; calibrating clears it; spec 035 behavior for non-teaching worlds untouched

**Independent Test**: quickstart.md §4 — fresh teaching world: provisional boot line + calibrate prompt; after calibrate + restart both gone

### Implementation for User Story 3

- [X] T010 [US3] Teaching flavor of the uncalibrated boot path in internal/daemon/daemon.go: when the teaching posture derives from a bootstrap-seeded provider (`CalibratedAt == ""`), mark the posture line provisional and extend `uncalibratedBootWarning` (or add the teaching variant beside it) with the explicit "posture cannot yet be honest — run `promptworld calibrate <world>`" prompt per contracts/posture.md §2; the posture rung is still applied
- [X] T011 [US3] Tests in internal/daemon/daemon_test.go: uncalibrated teaching boot output (provisional + prompt, golden-style like the existing uncalibratedBootWarning test); calibrated teaching boot has neither; uncalibrated NON-teaching boot output byte-identical to pre-039

**Checkpoint**: all warning/prompt paths honest

---

## Phase 6: User Story 4 - The posture is a per-world fact other features can read (Priority: P3)

**Goal**: `posture` block on status-family replies for teaching+LLM worlds only; CLI status line; offline toggle command

**Independent Test**: quickstart.md §5 — status shows `posture {rung, calibrated}` and the CLI line; `promptworld teaching <world> on|off` toggles the manifest; non-teaching status byte-identical

### Implementation for User Story 4

- [X] T012 [US4] Add `PostureStatus{Rung string; Calibrated bool}` + `StatusData.Posture *PostureStatus` (`json:"posture,omitempty"`, doc comment per contracts/posture.md §4) in internal/ipc/protocol.go; compose in the status path in internal/ipc/server.go only when `s.w.Manifest.Teaching && s.llm != nil`, recomputed per reply from the planner-serving provider
- [X] T013 [P] [US4] `promptworld teaching <world> [on|off]` subcommand in cmd/promptworld/commands.go using `world.SetTeaching` (print-current with no arg; note "applies at next daemon start"); register in the command table/help
- [X] T014 [US4] Render the posture line in `promptworld status` output in cmd/promptworld/commands.go (calibrated vs provisional wording per contracts/posture.md §5)
- [X] T015 [US4] Tests: status reply carries posture only for teaching+LLM (byte-identity for non-teaching and pure-sim replies) in internal/ipc/server_test.go; toggle round-trip + status line rendering in cmd/promptworld tests

**Checkpoint**: all stories independently functional

---

## Phase 7: Polish & Cross-Cutting Concerns

- [X] T016 Run quickstart.md end-to-end (all six sections) in the worktree; `go vet ./... && go test ./...` green; fix fallout
- [X] T017 Replay determinism check: teaching world's log replays byte-identical including the boot `clock.speed_set` event (reuse spec 036's replay harness pattern)
- [X] T018 Open PR from `.worktrees/task-78` (one TASK, one PR); after merge at root: `/grounding-wiki:wiki-update` re-pin (notes sourcing world.go, horizon.go, protocol.go, server.go, daemon.go, commands.go), then player-docs freshness (`node .claude/skills/player-docs/scripts/check-freshness.mjs --check`) and regenerate if stale
  — landed as a direct `--no-ff` merge to main (a733aa3) at user direction during the 2026-07-24 GitHub PR-API outage (branch pushed + review-passed first; PR body preserved in the merge commit). Wiki re-pinned at d120aef (13 notes); player-docs refreshed at c0221aa (7/7 fresh).

---

## Dependencies & Execution Order

- **Phase 1 → Phase 2**: T002/T003/T004 need the worktree
- **Phase 2 blocks all stories**: T002 (marker) feeds T005/T006/T008/T010/T012/T013; T003 (arithmetic) feeds T005/T008/T012
- **US1 (T005-T007)**: only needs Phase 2
- **US2 (T008-T009)**: only needs Phase 2 (posture warning is independent of the boot default)
- **US3 (T010-T011)**: builds on T005's boot posture line (same daemon.go region) — run after US1
- **US4 (T012-T015)**: only needs Phase 2; T013 parallel with T012/T014
- **Polish (T016-T018)**: after all stories; T018's wiki/player-docs legs run at root post-merge

### Parallel Opportunities

- T002 ∥ T003 (different packages); T004 splits across both test files
- After Phase 2: US1, US2, US4 can proceed in parallel (different files except commands.go — serialize T006/T013/T014 edits); US3 follows US1
- T013 ∥ T012/T014

---

## Implementation Strategy

MVP = Phases 1-3 (US1): a teaching world that boots at the honest fastest speed is
demonstrable value alone. Then US2 (the lesson-on-override), US3 (honesty when
uncalibrated), US4 (consumer surface), polish. Single implementer (Opus 4.8
spec-implementer per plan.md Constitution Check V) — execute sequentially in task-ID
order; commit per task or logical group on the one branch.
