# Tasks: Design-gate semantic lint

**Input**: Design documents from `/specs/075-design-gate-semantic-lint/`
**Prerequisites**: spec.md (pinned 8-cell inventory), plan.md; spec 072's PR
merged (shared pages: `overlays/postmortem.md`, `panels/exercise.md`)

**Tests**: the script is its own test vehicle — red run (exit 1, exactly 8
findings) before the cell fixes, green run (exit 0) after.

## Phase 1: Setup

- [X] T001 Preconditions at root: confirm spec 072 (`task-149-report-card-truth`) is merged to `origin/main`; `git fetch origin && git pull --ff-only`; cut the worktree via `node scripts/check-merge-drift.mjs worktree --spec 075 --task TASK-150` then `git worktree add .worktrees/task-150 -b task-150-design-gate-semantic-lint origin/main`; push the claim per spec 065
- [X] T002 Baseline in the worktree: current `node scripts/check-tui-design.mjs` exits 0 (structurally green pre-change); record the 8-cell inventory with `grep -rn "unbuilt (wave" docs/design/tui/` (7 renderer cells in `overlays/postmortem.md`, 1 in `overlays/help.md`, prose excluded) and confirm `panels/exercise.md` carries no stale "TASK-127, unbuilt" claim on the post-072 base (plan D6, spec FR-007)

## Phase 2: Board AC #1 — the lint (US1, P1)

- [X] T003 [US1] Add the `semantic-cells` check to `scripts/check-tui-design.mjs`: for every page with frontmatter `status: shipped`, flag any canonical control-table row whose renderer column (column 4) contains the literal `unbuilt (wave` — violation with file, 1-based line, cell text, remedy; structural pass, existing exit/JSON/human conventions untouched; header gains the spec-075 contract line (plan D1/D2/D7, spec FR-001/002/003/009)
- [X] T004 [US1] Red-run proof: the extended script exits 1 with exactly 8 `semantic-cells` violations (7 postmortem, 1 help) and no other new findings; `--json` shape unchanged; `status: specified` pages and `unbuilt (pending TASK-118)` (guardian-strip.md) not flagged — capture the output as SC-001's artifact for the notes/PR body (plan D3, spec FR-008, SC-001)

## Phase 3: Board AC #2 — the corpus stops lying (US2, P1)

- [X] T005 [US2] Amend `docs/design/tui/overlays/postmortem.md`: rename the seven wave-marked renderer cells (rows *postmortem takeover*, *run-end narrated line*, *morgue evidence rows*, *report card (scored runs only)*, *dismiss*, *replay via reopen key*, *replay via reattach* — by `control/region`, not line number) to the real symbols verified against code in-branch (`postmortemView`, `reportCardView`/`reportCard` via `buildChecklistCard`, `Model.runEnded()`, the verified dismiss/reopen key handlers), preserving all other columns and sharing annotations; re-pin `verified_against` to a branch commit (plan D4, spec FR-005)
- [X] T006 [US2] Amend `docs/design/tui/overlays/help.md`: badge deep-link renderer cell → `unbuilt (pending TASK-142, layer-2)`; update the hybrid-status prose paragraph to the pending-task form; re-pin `verified_against` (plan D5, spec FR-006)
- [X] T007 [US2] Verify `docs/design/tui/panels/exercise.md` by content on this branch: no stale "TASK-127, unbuilt" claim remains (expected — spec 072 amended it); record the verification; on residue only, correct and re-pin (plan D6, spec FR-007)
- [X] T008 [US2] Green-run self-test: `node scripts/check-tui-design.mjs` and `node scripts/check-tui-design.mjs --changed` both exit 0 at the branch tip — the lint does not fail its own PR (spec FR-008, SC-003)

## Phase 4: Grounding + gates (in-branch, per the in-PR doctrine)

- [X] T009 Wiki + player-docs probes in-branch: no `docs/wiki/` note lists `scripts/check-tui-design.mjs` or any touched page as a source (expected no re-pin, spec SC-005); `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` clean (`docs/wiki/` untouched); `go test ./...` green per doctrine (spec SC-004)
- [X] T010 PR: `node scripts/check-merge-drift.mjs pr` passes from the worktree; PR body carries the red/green proof; merge with `gh pr merge --merge` (merge-commit-only — branch-hash pins) (plan D8) — local gate verified in-branch (exit 0, warnings only); PR open/merge left to the orchestrator

## Phase 5: Post-merge bookkeeping (derived state only)

- [ ] T011 At root after the merge: spec-bridge sync, tasks.md ticks, runbook execution-log row; worktree removed, branch deleted, root ff-pulled (plan D8)

## Dependencies

T001 → T002 → T003 → T004 → (T005, T006, T007 — parallel, distinct files) →
T008 → T009 → T010; T011 post-merge. MVP = T003+T004 (AC #1) + T005–T008
(AC #2); both board ACs must land in the one PR.
