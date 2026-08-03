# Tasks: Claim gate sees spec numbers held by pushed branches

**Spec**: `specs/111-claim-gate-branch-visibility/` · **Task**: TASK-188

## Phase 1 — Gate implementation

- [x] **T001** Add `branchHeldSpecNumbers(cwd)` beside `takenSpecNumbers()` in
  `scripts/check-merge-drift.mjs`: pure read of `refs/remotes/origin/task-*` via
  `for-each-ref` + per-branch `ls-tree`, returning `Map(number -> {dir, branch})`,
  degrading to an empty map / skipped branch on any non-zero git status.
  *(FR-001, FR-006, FR-008)*
- [x] **T002** Add `nextFreeSpecNumber(...maps)` computing `max(all keys) + 1` over
  the union, guarded to return `1` when every map is empty. *(FR-005)*
- [x] **T003** Rewrite `runClaim()`'s collision decision to the main-then-branch order
  from plan §2, emitting the branch-held message shape from plan §4 while leaving the
  main-held message byte-identical. *(FR-002, FR-003, FR-004)*

## Phase 2 — Regression tests

- [x] **T004** Extend `scripts/claim-protocol.test.mjs`: branch-held collision blocks
  and names the branch; owner re-claim passes with the holder on its own branch;
  next-free skips branch-held numbers; main-held collision still blocks unchanged.
  *(SC-002, SC-003)*
- [x] **T005** Run `node --test scripts/check-merge-drift.test.mjs` and confirm no
  other mode regressed. *(SC-003)*

## Phase 3 — Evidence and grounding

- [x] **T006** Re-pin any `docs/wiki/` note listing `scripts/check-merge-drift.mjs`
  as a source; regenerate `docs/player/` if the wiki changed. *(spec 069, pr gate)*
  **Result: no-op.** No wiki note lists any `scripts/` path as a source — the corpus
  grounds the Go codebase, not the harness scripts. Confirmed by the pr gate.
- [x] **T007** Reproduce the live collision against the real repo with the fixed gate;
  record the transcript on TASK-188. *(SC-001, AC#6)*
  **Result: reproduced.** Note the situation moved while this task was in flight —
  TASK-187 renumbered itself off 110 to `specs/112-tui-frame-harness`, so the
  double-claim on 110 no longer exists. Both lanes' branch-held claims still
  reproduce the defect class the fix targets:

  ```
  $ node scripts/check-merge-drift.mjs claim --dir 110-something-new
  verdict=blocked  exit=1
    [block] spec-number-collision: specs/110-absence-attribution is already claimed
      on branch origin/task-173-absence-attribution for number 110 — claim
      specs/110-something-new is a collision; next free number is 113 (if that
      branch is abandoned, delete it on origin and re-run) (task=TASK-173)

  $ node scripts/check-merge-drift.mjs claim --dir 112-other-thing
  verdict=blocked  exit=1
    [block] spec-number-collision: specs/112-tui-frame-harness is already claimed
      on branch origin/task-187-frame-harness ... (task=TASK-187)

  $ node scripts/check-merge-drift.mjs claim --dir 111-claim-gate-branch-visibility
  verdict=pass  exit=0      # owner re-claim, holder on main + own branch
  $ node scripts/check-merge-drift.mjs claim --dir 113-fresh
  verdict=pass  exit=0      # unheld number
  ```

  Before the fix all four of these passed.
- [ ] **T008** Run `node scripts/check-merge-drift.mjs pr` from the worktree; open the
  PR (merge-commit only).
