# Tasks: Wiki corpus-spec v2 adoption

**Input**: spec.md (decisions 1–6 are the design; no separate plan.md — the
"plan" is the worklist the gate itself emits, re-derived in-branch per spec
assumption)

## Phase 1: Setup

- [x] T001 Worktree `.worktrees/task-146`; re-derive the worklist: run the freshness gate with a throwaway CAPSULES.md to get the current failure list (then delete it); record counts in the task notes

## Phase 2: User Story 1 — the corpus passes v2 (P1)

- [x] T002 [US1] Split/tighten/exempt every over-budget note body per spec Decisions 1–3 (Sonnet fan-out in batches; orchestrator reviews each batch's INDEX/link/pin discipline), in `docs/wiki/`
- [x] T003 [US1] Rewrite every over-budget capsule (≥ the six named) + any capsule whose note's coverage changed, ≤500 chars routing text, in `docs/wiki/*.md` frontmatter
- [x] T004 [US1] Update `docs/wiki/INDEX.md` (one line per child; grouped placement)
- [x] T005 [US1] Generate `docs/wiki/CAPSULES.md` via capsules.mjs; freshness gate exit 0 in v2 failure mode

## Phase 3: User Story 2 — downstream intact (P1)

- [x] T006 [US2] Bump `docs/player/*.html` wiki source pins to each note's current `verified_against` (paths unchanged); player-docs checker 13/13 in-branch
- [x] T007 [US2] Content-loss spot audit: per split note, parent+children substance accounting vs original; record in task notes

## Phase 4: Polish

- [x] T008 `node scripts/check-merge-drift.mjs pr` pass from the worktree; PR (merge commit) with corpus + capsules + player pins together
- [x] T009 Post-merge (root): spec-bridge sync, tasks.md ticks, runbook log row — derived state only

## Dependencies

T001 → T002 → (T003 ∥ T004) → T005 → T006/T007 → T008 → T009.
