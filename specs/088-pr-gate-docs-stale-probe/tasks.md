# Tasks: merge-drift pr gate — docs-stale probe on all pinned sources + history moves

**Input**: Design documents from `specs/088-pr-gate-docs-stale-probe/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: fixture tests are REQUIRED (spec FR-006 / card AC #4).

**Organization**: grouped by user story; every change lives in
`scripts/check-merge-drift.mjs` + `scripts/check-merge-drift.test.mjs`, so tasks
within a phase are sequential (same files), not [P].

## Phase 1: Foundational (trigger plumbing)

- [ ] T001 Extract the probe-trigger computation into a named helper in
  `scripts/check-merge-drift.mjs`: collect `promptworld-docs:source` tags from
  `docs/player/*.html` at the branch tip, union with the `docs/wiki/` prefix rule;
  return the matched trigger reasons for a given `branchFiles` list (research D1;
  FR-001 one-named-place requirement).
- [ ] T002 Add the history-move predicate (`git rev-list --merges origin/main..<tip>`
  non-empty) to the same helper (research D2; FR-003).
- [ ] T003 Restructure the probe call site (~line 1645) to invoke the player-docs
  checker at most once when ANY trigger matches, threading trigger reasons into the
  finding message; preserve existing finding rules/severities (FR-004, FR-005).

## Phase 2: User Story 1 — non-wiki pinned inputs gate (P1)

- [ ] T004 [US1] Fixtures F1, F2, F9 in `scripts/check-merge-drift.test.mjs`:
  README-only branch with stale checker blocks (`player-docs-stale`); declared
  spec-046 source with fresh checker passes; checker missing on README-only trigger
  blocks (`player-docs-env-error`). Existing 069/US2 no-trigger test must still pass
  (F8 / SC-004).

## Phase 3: User Story 3 — history moves re-trigger (P2)

- [ ] T005 [US3] Fixtures F3, F4: branch tip containing a merge of main with no
  pinned-source diff paths invokes the checker (stale blocks; fresh passes — no
  false blocking).

## Phase 4: User Story 2 — design-reference pins block (P2)

- [ ] T006 [US2] Add tui-design delegation to pr mode in
  `scripts/check-merge-drift.mjs`: on the trigger set from data-model.md, run
  `check-tui-design.mjs --changed <range> --json` (env-overridable path
  `CHECK_MERGE_DRIFT_TUI_DESIGN_CHECKER`), map exit 1 → blocking `tui-design-stale`
  (pages named from JSON), exit 2/missing → blocking `tui-design-env-error`; keep
  `tui-surface` warn (research D3; FR-002).
- [ ] T007 [US2] Fixtures F5, F6, F7: design-pin drift blocks; re-pinned branch
  passes with warn-only reminder; combined pinned-input + history-move run emits
  each finding at most once (FR-004).

## Phase 5: Polish & verification

- [ ] T008 Run the full quickstart validation: `node --test
  scripts/check-merge-drift.test.mjs` green; manual `pr` run from this worktree
  (which itself has a claim + spec commits) behaves per quickstart.md; re-verify +
  re-pin any wiki note listing `scripts/check-merge-drift.mjs` as a source
  (constitution IV — the gate self-applies).
