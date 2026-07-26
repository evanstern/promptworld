# Tasks: In-TUI forward-ladder view

**Input**: Design documents from `/specs/078-tui-ladder-view/`
**Prerequisites**: spec.md (parity contract + TASK-151 armor), plan.md (D1–D8)

**Tests**: the parity test (plan D4) IS a deliverable (spec FR-005); the
byte-identity/degradation suite extensions (D5) prove FR-008/FR-010.

**Board AC mapping**: Phase 2+3 → TASK-152 AC #1 (ladder + `stages --json`
parity); Phase 4 → AC #2 (help.md byte-identity row + design gate).

## Phase 1: Setup

- [ ] T001 Cut the worktree per protocol (from a fresh repo root:
  `git fetch origin && git pull --ff-only`, then
  `node scripts/check-merge-drift.mjs worktree --spec 078 --task TASK-152`,
  then `git worktree add .worktrees/task-152 -b task-152-tui-ladder-view origin/main`;
  push the branch on first commit); confirm baseline
  `go test ./internal/tui/ ./internal/worlds/ ./cmd/promptworld/` green in
  `.worktrees/task-152`. If TASK-151 has merged, note any
  substrate-symbol drift against spec.md's Grounding section before coding.

## Phase 2: Foundational — the shared earned rule (blocks US1)

- [ ] T002 Relocate the earned rule into `internal/worlds/unlocks.go`:
  nil-safe `(u *Unlocks) StageEarned(stage string) bool` (stage-1
  unconditional floor ∨ record entry), doc comment naming the spec 063 T014
  one-source-two-surfaces precedent (plan D1, spec FR-003)
- [ ] T003 `cmd/promptworld/stages.go` consumes it: delete local
  `stageEarned`, point `cmdStages` + `highestEarnedStage` at the method;
  assert zero output change (`stages` text + `--json`) via existing tests /
  a quick fixture run; unit-test the method in
  `internal/worlds/unlocks_test.go` (nil receiver, floor, entry) (plan D1)

## Phase 3: User Story 1 — the ladder block (P1, board AC #1) 🎯 MVP

- [ ] T004 [US1] Model plumbing in `internal/tui/tui.go`:
  `unlocks *worlds.Unlocks` field, loaded once in `New()` (boot precedent
  `populateHelpLessons`); render-time earned predicate = record ∨
  `replica.StagesUnlocked` membership (plan D2, spec FR-006)
- [ ] T005 [US1] `helpLadderLines` in `internal/tui/help.go`, appended by
  `helpGuardianLines`: iterate `world.StageOrder`; per stage render skin
  identity + `(id)`, `teaches:` concept, `unlocked by:` evidence
  (graduation wording when empty), earned/next/not-yet state with the
  audit pointer for record-earned stages, you-are-here marker with the
  `StageOverridden` annotation; `wrapText`/`clipLine` throughout (plan D3,
  spec FR-001/002/004/007/009)
- [ ] T006 [US1] Parity test in `internal/tui/help_test.go`: fixture
  unlocks record under `setHome(t)` (one record-earned stage with
  world+exercise); expected rows computed at runtime from the SAME
  substrate `stages --json` marshals (`StageOrder` × `StagesLadder` ×
  `StageEarned` × `skin.Stage`); assert every field surfaces in the
  rendered guardian section; zero hardcoded stage ids/counts/prose — the
  TASK-151 armor (plan D4, spec FR-005/SC-001)
- [ ] T007 [US1] Byte-identity + degradation tests: extend
  `TestHelpGuardianByteIdenticalPerStatus` to the ladder inputs (fixed
  stage/override/unlocks/replica ⇒ constant bytes); nil status + nil
  replica + no unlocks file ⇒ floor ladder, non-empty, no panic
  (`TestHelpContentReadsNoStatusOrReplica` stays green); replica-only
  mid-session unlock shows earned without audit pointer; overridden world
  annotated but not earned-laundered; 80×24 scroll reachability (the
  `TestHelpWalkthroughScrollsAt80x24` precedent) (plan D5, spec
  FR-008/010, edge cases)

## Phase 4: User Story 2 — design authority, same PR (P2, board AC #2)

- [ ] T008 [US2] Amend `docs/design/tui/overlays/help.md`: Section 5 "The
  ladder" content contract; byte-identity classification table row
  (unlocks-record-derived, model-free — exact wording per plan D6); control
  table row naming `helpLadderLines` + its data sources (plan D6, spec
  FR-011)
- [ ] T009 [US2] Re-verify + re-pin help.md (`verified_against` → branch
  commit); run `node scripts/check-tui-design.mjs --changed` from the
  worktree until it exits 0 (spec 047 same-PR gate, spec SC-002)

## Phase 5: Grounding (in-branch, per the wiki-in-PR lifecycle)

- [ ] T010 Re-verify + re-pin every wiki note whose pinned sources this
  branch touched: `docs/wiki/tui-input-help.md` (help.go/tui.go),
  `cli-world-lifecycle.md` + `curriculum-ladder.md` (stages.go),
  `curriculum-ladder-progression.md` (unlocks.go) — fold the new
  StageEarned/ladder facts where the notes state behavior; then (since
  `docs/wiki/` changed) regenerate `docs/player/` via the player-docs skill
  and confirm `node .claude/skills/player-docs/scripts/check-freshness.mjs
  --check` exits 0 — note the 073 lesson: the freshness checker also tracks
  `docs/design/tui/*` plain-file sources, so re-check after T009's re-pin
  (plan D7, spec SC-004, constitution IV)

## Phase 6: Polish & gates

- [ ] T011 `gofmt -l` clean; full `go test ./...` green; `node
  scripts/check-merge-drift.mjs pr` from `.worktrees/task-152` exits 0
  (wiki-repin-missing / player-docs-stale both clear); PR opens carrying
  code + parity test + help.md amendment + wiki re-pins + player-docs
  together; merged with `gh pr merge --merge` ONLY (merge-commit-only —
  in-branch pins are branch hashes) (plan D8, spec SC-003)

## Phase 7: Post-merge bookkeeping (root, derived state only)

- [ ] T012 From repo root after merge: spec-bridge sync (TASK-152 → Done as
  artifacts prove), tick this tasks.md, runbook execution-log row
  (`docs/design/reorient-2026-07-26-sweep-runbook.md`), worktree + branch
  cleanup, ff-pull root (spec 065 / constitution IV boundary — no grounding
  content post-merge)

## Dependencies & Execution Order

T001 → (T002 → T003) → (T004 → T005 → T006 → T007) → (T008 → T009) → T010 →
T011; T012 post-merge. T006 and T007 may proceed in parallel once T005
lands; T008 may draft in parallel with Phase 3 but T009's re-pin must
follow the final code commit it verifies. MVP = Phases 2+3 (board AC #1);
Phase 4 is board AC #2 and rides the same PR.
