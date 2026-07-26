# Tasks: Mouse-parity sweep test

**Input**: Design documents from `/specs/073-mouse-parity-sweep/`
**Prerequisites**: spec.md (grounded cell inventory + handler-evidence decision), plan.md

**Tests**: the deliverable IS a test; proof = the sweep passing on the shipped
corpus + the SC-001 mutation check failing loudly.

## Phase 1: Setup

- [ ] T001 Cut the worktree per protocol (from repo root:
  `node scripts/check-merge-drift.mjs worktree --spec 073 --task TASK-154`,
  then `git worktree add .worktrees/task-154 -b task-154-mouse-parity-sweep origin/main`);
  confirm baseline `go test ./internal/tui/` green in
  `.worktrees/task-154`

## Phase 2: US1 — the sweep test (board AC #1: parses control tables, fails on any non-'—' mouse cell without a handler)

- [ ] T002 [US1] Parser + classifier in `internal/tui/mouseparity_test.go`
  (new file): walk `../../docs/design/tui`, match the canonical control-table
  header, extract column 5, classify display-only / tracked gap / shipped
  claim with last-` · `-segment mouse-half semantics (plan D1/D2, spec
  FR-001/FR-006 — `t.Fatalf` on unreadable corpus, non-vacuity asserted)
- [ ] T003 [US1] Oracle + live-dispatch proof in the same file: one entry —
  `panels/chronicle.md` / `jump-to-source` / `click line` — whose check
  drives `widescreenModel` into inspect mode, renders to populate
  `chronHit`, sends `mouseLeftRelease` through `Update`, and asserts
  selection + `jumpToSource` effect (plan D3, spec FR-003; crib
  `internal/tui/tui_test.go:872`)
- [ ] T004 [US1] `TestMouseParitySweep` assembly: direction 1 (claim without
  oracle entry → fail naming page + control, and run every oracle check as a
  subtest), direction 2 (stale oracle entry → fail), rollout-note presence
  check for pages with tracked-gap rows (plan D4, spec FR-002/004/005)
- [ ] T005 [US1] Proof: `go test ./internal/tui/ -run TestMouseParity -v`
  passes on the shipped corpus; SC-001 mutation check — temporarily flip one
  `—` mouse half in a control table to a fake claim, confirm the sweep fails
  naming that page and control, revert (plan D6)

## Phase 3: US2 — mechanized-tracking note (board AC #2: patterns/keymap.md rollout note updated)

- [ ] T006 [US2] Amend doctrine rule 3 in
  `docs/design/tui/patterns/keymap.md`: tracking is mechanized by
  `TestMouseParitySweep` (`internal/tui/mouseparity_test.go`); state the
  graduation contract (real mouse cell + oracle entry + passing live proof,
  same PR); no other rule changes (plan D5, spec FR-007)
- [ ] T007 [US2] Re-verify + re-pin `docs/design/tui/patterns/keymap.md`
  (`verified_against` → a branch commit); run
  `node scripts/check-tui-design.mjs --changed` from the worktree until it
  passes (spec 047 same-PR gate, spec SC-002)

## Phase 4: Grounding + gates (in-branch, per the wiki-in-PR lifecycle)

- [ ] T008 Wiki probe: no note pins the touched files (spec Assumptions;
  candidates checked: `docs/wiki/testing-strategy.md`,
  `docs/wiki/tui-input-help.md`, `docs/wiki/tui-client.md` — none list
  `mouseparity_test.go` or `docs/design/tui/*` as sources); if the pr gate
  nonetheless reports `wiki-repin-missing`, re-verify and re-pin the named
  note ON THIS BRANCH, then (only if `docs/wiki/` changed) regenerate
  `docs/player/` via the player-docs skill and confirm
  `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
  clean (constitution IV, spec SC-004)
- [ ] T009 Full-suite + pr gate: `go test ./...` green; `gofmt -l` clean on
  the new file; `node scripts/check-merge-drift.mjs pr` from
  `.worktrees/task-154` exits 0; PR opens with test + keymap.md amendment +
  re-pin together, merged with `gh pr merge --merge` (merge-commit-only —
  the keymap.md pin is a branch hash)

## Phase 5: Post-merge bookkeeping (root, derived state only)

- [ ] T010 From repo root after merge: spec-bridge sync (TASK-154 → Done as
  artifacts prove), tick this tasks.md, runbook execution-log row
  (`docs/design/reorient-2026-07-26-sweep-runbook.md`), worktree + branch
  cleanup (spec 065 / constitution IV boundary — no grounding content in
  post-merge commits)

## Dependencies

T001 → T002 → T003 → T004 → T005 → (T006 → T007) → T008 → T009; T010
post-merge. MVP = Phase 2 (board AC #1); Phase 3 is board AC #2 and rides the
same PR.
