# Tasks: Wiki grounding moves inside the PR cycle

**Input**: Design documents from `/specs/069-wiki-in-pr-gate/`
**Prerequisites**: plan.md, spec.md, research.md

**Tests**: included — the gate's correctness IS the deliverable (FR-009).

**Organization**: US1/US2 are the gate (P1, implementer); US3 is doctrine (P2,
split implementer/orchestrator per plan D4).

## Phase 1: Setup

- [X] T001 Baseline from the task worktree: `node --test scripts/check-merge-drift.test.mjs scripts/claim-protocol.test.mjs` green before changes (worktree `.worktrees/task-145`)

## Phase 2: Foundational

- [X] T002 Note-at-tip reader `loadWikiNotesAt(ref, notePaths, cwd)` reusing the existing frontmatter parser, in `scripts/check-merge-drift.mjs` (plan D1)

## Phase 3: User Story 1 — A code PR cannot open without its wiki grounding (P1)

**Goal**: pr mode blocks unless the branch carries its own re-verification.

**Independent Test**: fixture matrix in `scripts/check-merge-drift.test.mjs`.

- [ ] T003 [US1] Predicate + finding swap in `gatePr`: evaluate pin-vs-branch per overlapped note; fail → block `wiki-repin-missing` (message names note, matched sources, remedy); pass → no finding; scoped `wiki-note-malformed` escalation for predicate-needed notes, in `scripts/check-merge-drift.mjs` (plan D2, spec FR-001/002/004/005)
- [ ] T004 [US1] Fixture-repo matrix: overlap-no-repin blocks; repinned-pass; source re-touched after pin blocks; pin unreachable from tip blocks; no-overlap branch unchanged; malformed-needed-note blocks — in `scripts/check-merge-drift.test.mjs` (plan D5, spec US1 scenarios + edge cases)

## Phase 4: User Story 2 — Player docs cannot go stale through a merge (P1)

**Goal**: wiki-touching branches must carry fresh player docs.

**Independent Test**: stub-checker fixture tests.

- [ ] T005 [US2] Player-docs checker spawn in `gatePr` gated on `docs/wiki/` changes; env-var override `CHECK_MERGE_DRIFT_PLAYER_DOCS_CHECKER` for tests; exit 1 → `player-docs-stale` block, exit 2 → `player-docs-env-error` block, no invocation without wiki changes, in `scripts/check-merge-drift.mjs` (plan D3, spec FR-003)
- [ ] T006 [US2] Stub-checker tests (exit 0/1/2 + not-invoked case) in `scripts/check-merge-drift.test.mjs` (plan D5)

## Phase 5: User Story 3 — The doctrine says what the gate enforces (P2)

**Goal**: a fresh session learns the lifecycle from the docs alone.

**Independent Test**: doc review vs TASK-145 ACs; constitution Sync Impact Report.

- [ ] T007 [US3] CLAUDE.md rewrite: PDLC loop diagram (wiki grounding in-branch, pre-PR), Grounding-freshness + Player-docs rules at the pr choke point, merge-commit-only (`gh pr merge --merge`), step-7 derived-state-only paragraph citing spec 065 + pdlc:sweep re-ground, in `CLAUDE.md` (spec FR-007)
- [ ] T008 [US3] Constitution Principle IV amendment via `speckit-constitution` (MINOR bump to v1.2.0, Sync Impact Report) — ORCHESTRATOR task, committed onto the task branch, in `.specify/memory/constitution.md` (spec FR-008)

## Phase 6: Polish & Cross-Cutting

- [ ] T009 Full harness green (`node --test scripts/*.test.mjs`); self-apply: `node scripts/check-merge-drift.mjs pr` from the worktree passes under the NEW logic before the PR opens (SC-003, SC-004 first half)
- [ ] T010 Post-merge (root): spec-bridge sync, tasks.md ticks, runbook execution-log row — derived state only, per the very doctrine this spec lands

## Dependencies & Execution Order

- T002 blocks T003; T003 blocks T004; T005 ∥ T003 (different gatePr regions, same file — implementer sequences within one dispatch); T007 ∥ T003-T006 (different files); T008 after T007 (constitution cites the CLAUDE.md wording), orchestrator-executed.
- MVP = Phases 2–4 (the gate). US3 completes the TASK-145 ACs.

**Parallel opportunities**: T007 alongside the script work; T004/T006 as one fixture-building pass.
