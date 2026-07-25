# Tasks: Survival Reflex Gaps (fire)

**Input**: spec.md, plan.md (specs/057-survival-reflex-gaps/)

**Tests**: included — SC-001..003 demand test evidence; replay determinism is a
hard invariant.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

*None — existing module; worktree is TASK-108 process.*

---

## Phase 2: Foundational

- [ ] T001 Locate the world-genesis seam: where `promptworld new` creates the
      world and seeds its first events (grep genesis/new-world paths in `cmd/`,
      `internal/world`, `internal/daemon`); record the exact function in the PR
      description. No behavior change yet.

---

## Phase 3: User Story 1 — Refuel arms below 3 game-hours (P1)

- [ ] T002 [US1] Change `defaultRefuelDyingBelow` 3600 → 10800 in
      `internal/sim/tuning.go` with a doctrine comment citing TASK-108 evidence
      (42 burnouts vs 8 builds).
- [ ] T003 [P] [US1] Tests: refuel intent fires at 2.5 game-hours remaining
      under default; does NOT fire under a `TuningState` pinning 3600 (dial
      still wins) — extend `internal/sim/tuning_test.go` or
      `food_fire_test.go` (SC-001).
- [ ] T004 [US1] Same-PR doc moves (FR-006):
      `specs/048-tuning-manifest/contracts/tuning.md` default column,
      `docs/design/control-surface-and-calibration.md` §2.4 fire row + §6 dial
      table.

---

## Phase 4: User Story 2 — Genesis tuning pin (P1)

- [ ] T005 [US2] Seed one `sim.tuning_applied` event (full
      `defaultTuning()` set via `sim.NewTuningEvent`) in the world-genesis
      seam found in T001, ordered with the other genesis events.
- [ ] T006 [US2] Verify the spec-048 boot seed against a pinned world: present
      unchanged manifest → no event; differing manifest → one event
      (`seedTuning` should need no change — prove it with a test, adjust only
      if the proof fails).
- [ ] T007 [P] [US2] Replay-independence test (SC-002): create a pinned state,
      apply events, replay the log into a fresh state whose accessor defaults
      are DIFFERENT (simulate a future default change by constructing the
      expected state from the pin, not the const) — hashes equal; plus a
      pre-057 fixture (no pin) still replays under compiled defaults (FR-007).
- [ ] T008 [US2] Document the pre-057/migrated-world default-shift hazard where
      the determinism scope note lives (plan R4 note: control-surface report §6
      + pointer on TASK-75 if its doc is unbuilt).

---

## Phase 5: User Story 3 — Cold build-fire reflex proven (P2)

- [ ] T009 [US3] Proof matrix test `internal/sim/reflex_matrix_test.go`: cold
      night × {wood≥2, wood=1+choppable-known, wood=0} × {warmth known-reachable,
      none-known, known-but-stale/dead} — assert each cell's doctrine outcome
      reflex-only (SC-003).
- [ ] T010 [US3] Fix any fire-adjacent surgical gap the matrix exposes (e.g.
      stale-warmth belief loop); anything non-surgical or out of the fire
      boundary is carded, not fixed — evidence in audit.md.

---

## Phase 6: User Story 4 — Survival audit (P3)

- [ ] T011 [US4] Write `specs/057-survival-reflex-gaps/audit.md`: every
      `decideIntent` rung — need protected, thresholds keyed, gap disposition
      (fixed here / carded TASK-n / no gap); cite the TASK-103 day-branch
      warmth gap as carded (SC-004).

---

## Phase 7: Polish & Cross-Cutting

- [ ] T012 Full gates in worktree: `go test ./...`, `go vet ./...`,
      `node scripts/check-tui-design.mjs --changed` (expected no-op); re-run
      after rebase.
- [ ] T013 Post-merge (root): wiki re-pin (`world-tuning`, `reflex-policy`,
      genesis-path notes), player-docs freshness check, spec-bridge sync.
