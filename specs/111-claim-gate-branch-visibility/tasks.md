# Tasks: Claim gate sees spec numbers held by pushed branches

**Spec**: `specs/111-claim-gate-branch-visibility/` · **Task**: TASK-188

## Phase 1 — Gate implementation

- [ ] **T001** Add `branchHeldSpecNumbers(cwd)` beside `takenSpecNumbers()` in
  `scripts/check-merge-drift.mjs`: pure read of `refs/remotes/origin/task-*` via
  `for-each-ref` + per-branch `ls-tree`, returning `Map(number -> {dir, branch})`,
  degrading to an empty map / skipped branch on any non-zero git status.
  *(FR-001, FR-006, FR-008)*
- [ ] **T002** Add `nextFreeSpecNumber(...maps)` computing `max(all keys) + 1` over
  the union, guarded to return `1` when every map is empty. *(FR-005)*
- [ ] **T003** Rewrite `runClaim()`'s collision decision to the main-then-branch order
  from plan §2, emitting the branch-held message shape from plan §4 while leaving the
  main-held message byte-identical. *(FR-002, FR-003, FR-004)*

## Phase 2 — Regression tests

- [ ] **T004** Extend `scripts/claim-protocol.test.mjs`: branch-held collision blocks
  and names the branch; owner re-claim passes with the holder on its own branch;
  next-free skips branch-held numbers; main-held collision still blocks unchanged.
  *(SC-002, SC-003)*
- [ ] **T005** Run `node --test scripts/check-merge-drift.test.mjs` and confirm no
  other mode regressed. *(SC-003)*

## Phase 3 — Evidence and grounding

- [ ] **T006** Re-pin any `docs/wiki/` note listing `scripts/check-merge-drift.mjs`
  as a source; regenerate `docs/player/` if the wiki changed. *(spec 069, pr gate)*
- [ ] **T007** Reproduce the live TASK-173/TASK-187 spec-110 collision against the
  real repo with the fixed gate; record the transcript on TASK-188. *(SC-001, AC#6)*
- [ ] **T008** Run `node scripts/check-merge-drift.mjs pr` from the worktree; open the
  PR (merge-commit only).
