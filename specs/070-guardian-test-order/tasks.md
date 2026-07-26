# Tasks: Guardian worker shutdown joins

**Input**: Design documents from `/specs/070-guardian-test-order/`
**Prerequisites**: spec.md (pinned diagnosis), plan.md

**Tests**: proof-by-repetition matrices (no new test files expected).

## Phase 1: Setup

- [x] T001 Baseline in the worktree (`.worktrees/task-144`): reproduce the flake once (`go test ./internal/guardian/ -run 'TestReportCard' -count=20` — expect ≥1 failure) and record it

## Phase 2: User Story 1 + 2 — deterministic worker shutdown (P1)

- [x] T002 [US1] WaitGroup join: `wg.Add(1)` at each of the four spawn sites in `New`, `defer wg.Done()` first line of each worker, `Close` = `close(done)` + `wg.Wait()`, in `internal/guardian/guardian.go` (plan D1/D2, spec FR-001/002)
- [x] T003 [US1] Proof matrices: `-run 'TestReportCard' -count=50` isolation clean; `go test ./internal/guardian/ -count=10 -race` clean; full `go test ./...` green (plan D3, spec FR-004)

## Phase 3: Grounding (in-branch, per the in-PR doctrine)

- [x] T004 Re-verify `docs/wiki/guardian.md` against the Close-semantics change and re-pin to a branch commit; run `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` in-branch (expected clean — no page cites guardian.md) (plan D4, spec SC-004)

## Phase 4: Polish

- [x] T005 `gofmt -l` clean; `node scripts/check-merge-drift.mjs pr` from the worktree passes; PR opens with code + grounding together
- [x] T006 Post-merge (root): spec-bridge sync, tasks.md ticks, runbook execution-log row (derived state only)

## Dependencies

T001 → T002 → T003 → T004 → T005; T006 post-merge. MVP = T002+T003.
