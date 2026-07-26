# Tasks: First-person harvest memory (mental-map update at chop/quarry time)

**Input**: Design documents from `/specs/081-first-person-harvest-memory/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/events.md, quickstart.md
**Board**: TASK-159 | **Branch**: `task-159-first-person-harvest-memory` (worktree `.worktrees/task-159`)
**Tier**: spec-implementer pinned Opus 4.8 (research.md D7)

**Tests**: included — the spec's success criteria and contract §4 invariant are
test-shaped, and house convention is tests alongside code.

**Organization**: tasks grouped by user story; US1 (actor) is the MVP slice.

## Phase 1: Setup

- [ ] T001 Baseline: in `.worktrees/task-159`, run `go test ./internal/sim/ ./internal/mind/` and record the green baseline in the task notes (no code changes)

## Phase 2: Foundational (blocking prerequisites)

- [ ] T002 Add `salChop = 4` and `salQuarry = 4` constants (salHunt band, comment the operator decision 2026-07-26 superseding the no-memory chop posture) in `internal/sim/memory.go`
- [ ] T003 Add first-person act-memory texts — chop `"Felled the tree at (%d,%d)."`, quarry `"Quarried the outcrop at (%d,%d)."` — beside `mapCorrectedText` in `internal/sim/memory.go`

**Checkpoint**: constants/texts compile; nothing behavioral yet.

## Phase 3: User Story 1 — My own harvest is mine, in the first person (P1) 🎯 MVP

**Goal**: actor's map fact removed at act time + first-person act memory; no self-discovery corrections ever.

**Independent Test** (spec US1): one villager chops one tree → exactly one "Felled the tree" memory, zero "had been felled when you looked" for that (agent, tile) at any tick.

- [ ] T004 [US1] In the `agent.chopped` reducer arm (`internal/sim/state.go:1057` region): after existing mutations, `a.Map.removeFact("tree", p.X, p.Y)` (nil-map/absent-fact safe); mirror in the `agent.quarried` arm (`state.go:1176` region) with `"rock"` — actor only in this task
- [ ] T005 [US1] At the executor chop/quarry emit sites (`internal/sim/executor.go`, chop completion and `case "quarry"` ~line 1330): append companion `situatedMemoryEvent(nextTick, i, salChop|salQuarry, where, in.Reason, OriginAction, …)` for the actor, same batch as the act (hunt precedent at executor.go:1221)
- [ ] T006 [US1] Test in `internal/sim/mentalmap_test.go`: chop removes actor's tree fact; quarry removes actor's rock fact; actor-never-knew-fact chop is a no-op removal that still mints the memory
- [ ] T007 [US1] Test in `internal/sim/memory_test.go` (or adjacent): completed chop/quarry accretes exactly one first-person actor memory via `agent.memory_added` (TestMemoriesAccrete posture preserved); salience below generation-bump band
- [ ] T008 [US1] Test (sweep integration, `internal/sim/mentalmap_test.go`): after a chop, no later perception pass emits `agent.map_corrected` naming (actor, tree, x, y) — SC-001 shape

**Checkpoint**: US1 independently shippable — every guaranteed self-correction is gone.

## Phase 4: User Story 2 — Watching a neighbor harvest is not a later mystery (P2)

**Goal**: awake in-radius witnesses' facts removed silently at act time; absorb parity for lost intent premises.

**Independent Test** (spec US2): two adjacent villagers, one chops → bystander's fact gone at act tick, zero memories minted for them, no later correction.

- [ ] T009 [US2] Extend both reducer arms in `internal/sim/state.go`: for every other villager with `!Dead && !Asleep && Map != nil` within `witnessRadius` of (x,y) (pre-mutation positions, the axe-yield idiom), `removeFact` the matching kind — provenance-blind (data-model §transitions)
- [ ] T010 [US2] Extend the absorb switch in `internal/mind/mind.go` (~line 263 case list / map_corrected arm at ~290): on `agent.chopped`/`agent.quarried`, additionally arm any villager within `witnessRadius` of (x,y) whose live intent has `(TargetX,TargetY)==(x,y)` or `(ResX,ResY)==(x,y)` (contract §5)
- [ ] T011 [US2] Test in `internal/sim/mentalmap_test.go`: awake in-radius witness fact removed with zero memory events; asleep in-radius witness keeps fact; `told`-provenance fact also removed (hearsay edge case); dead agents untouched
- [ ] T012 [US2] Test in `internal/sim/mentalmap_test.go`: same-tick sweep — a perception beat on the act tick emits no correction for the acted tile, and the next beat finds nothing to correct (FR-006, research D5)
- [ ] T013 [US2] Test in `internal/mind/` (absorb test file beside `mind.go`): witness with intent targeting the felled tile re-arms on the act event; witness with unrelated intent stays quiet (map_corrected parity)

**Checkpoint**: all on-scene parties absorb the act at act time.

## Phase 5: User Story 3 — Genuine return-discovery still works (P3, regression guard)

**Independent Test** (spec US3): A witnesses tree, leaves radius; B chops; A returns → exactly one correction + discovery memory for A.

- [ ] T014 [US3] Test in `internal/sim/mentalmap_test.go`: out-of-radius agent keeps the fact at act time and corrects exactly once on return with the situated discovery memory; asleep-at-act-tick agent likewise corrects later (spec US2 AS3 / US3)
- [ ] T015 [US3] Test (contract §4 invariant): fold a scripted log with actor + on-scene witness + absent agent; assert the only `agent.map_corrected` names the absent agent

## Phase 6: Polish & Cross-Cutting

- [ ] T016 Replay determinism case (SC-005): extend the existing fold/replay harness (`internal/sim/fork_event_test.go` shape) with a chop+witnesses log; canonical state bytes (mental maps included) identical across fold and replay
- [ ] T017 Full gates in worktree: `go test ./...`; `node scripts/check-tui-design.mjs --changed` (expect no-op)
- [ ] T018 Live validation per quickstart.md §3 on a fresh world (~1 game day): SC-001/SC-002/SC-004 SQL checks land at expected values; record counts vs the worldy baseline (103/103 self-corrections, 75% loss memories) in TASK-159 notes; spot-check journals/chronicle (SC-006)
- [ ] T019 Wiki re-pins ON THE BRANCH (spec 069): re-verify + re-pin every note listing touched sources — at minimum `docs/wiki/mental-map-perception.md`, `mental-maps.md`, `event-types-mental-map.md`, `mental-map-model.md`, plus any memory/salience note listing `internal/sim/memory.go` or `internal/mind/mind.go` (`/grounding-wiki:wiki-update` scoped to the branch diff)
- [ ] T020 Regenerate `docs/player/` in-branch (player-docs skill; `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` must pass)
- [ ] T021 Pre-PR gate from the worktree: `node scripts/check-merge-drift.mjs pr` exits 0

## Dependencies & Execution Order

- Phase 1 → Phase 2 → Phase 3 (US1) → Phase 4 (US2) → Phase 5 (US3) → Phase 6
- US1 is independently testable/shippable (MVP). US2 depends on US1's reducer
  edits only textually (same arms); US3 is tests-only and depends on US1+US2
  behavior being in place.
- [P] opportunities: T002/T003 (same file — sequential in practice); T006/T007
  touch different test files and may run [P] after T004+T005; T011-T013 after
  T009+T010 (T013 in `internal/mind` is [P] with T011/T012); T019/T020
  sequential (player docs read the re-pinned wiki).

## Implementation Strategy

MVP = Phase 1-3 (US1): kills all 103-of-103 guaranteed self-corrections.
Ship the rest in the same PR (one task, one PR) — US2/US3 are small and the
board task's ACs span all three stories. The implementer works entirely in
`.worktrees/task-159`; board/tasks.md ticks happen from repo root (worktree
cwd bookkeeping pitfall).
