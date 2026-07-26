---
id: TASK-155
title: >-
  check-merge-drift: paused-label lanes downgrade to info — praxisflux TASK-55
  counterpart
status: In Progress
assignee:
  - '@claude'
created_date: '2026-07-26 20:11'
updated_date: '2026-07-26 20:19'
labels: []
dependencies: []
ordinal: 125000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A task paused by the operator (label 'paused' in the task file's frontmatter labels: list, set/cleared only via backlog task edit --labels; provenance as a 'paused by <who> <date>: <why>' append-note — the praxisflux paused-lane convention) is parked state, not a live lane. check-merge-drift.mjs must downgrade a paused task's branch/worktree findings from blocking/warn to info in all three gate modes (session/worktree/pr), with the pause cited as evidence, and must never prescribe cleanup of a paused task's branch or worktree. Counterpart of praxisflux TASK-55 / specs/015-paused-lane-marker (praxis repo).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Paused detection reads the 'paused' label from the task file's frontmatter labels: list
- [x] #2 session mode: a paused task's branch findings (textual/pairwise conflict, stale-base, branch-unpushed, backlog-overlap, spec-number-collision, cleanup-eligible) downgrade to info with the pause as evidence
- [x] #3 session mode never emits cleanup prescriptions for a paused task's branch/worktree (--apply-cleanup untouched)
- [x] #4 pr mode: drift findings for a paused task's branch downgrade to info; spec-069 grounding gates (wiki-repin-missing, player-docs-*) stay blocking
- [x] #5 worktree mode: the downgrade is wired mode-uniformly; ownership protections (spec-number-collision on a paused task's claimed dir, claim mode) stay blocking
- [x] #6 Covered by scripts/check-merge-drift.test.mjs regression tests
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Claim: card In Progress + specs/080-paused-label-lanes stub (this commit)
2. check-merge-drift.mjs: taskLabels/isTaskPaused helpers reading frontmatter labels:; downgrade-to-info wrapper applied at finding construction in session (per-branch, pairwise = either side) and pr (gated branch) modes; worktree mode wired through the same helper; cleanup prescriptions skipped for paused branches; pause cited as evidence
3. Regression tests in scripts/check-merge-drift.test.mjs (bare-origin + clone fixture per existing convention): paused branch downgrades in session + pr, cleanup not prescribed, unpaused control unchanged, spec-069 gates still block
4. Gates: node --test scripts/, pr mode from this worktree exit 0
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented on branch paused-lane-marker (worktree .worktrees/paused-lane-marker). check-merge-drift.mjs: parseTaskLabels/isTaskPaused (frontmatter labels:, block + inline forms, read from the main worktree's working tree) + makeDriftFinding wrapper; session sites (textual/pairwise conflict, branch-unpushed, backlog-overlap, spec-number-collision, cleanup-eligible with prescription exclusion) and pr sites (textual-conflict block→info, stale-base, backlog-overlap, spec-number-collision) route through it; worktree/claim ownership protections and pr-mode spec-069 grounding blocks deliberately unchanged. Six spec-080 regression tests in scripts/check-merge-drift.test.mjs — node --test: 21/21 pass (plus claim-protocol 10/10). Spec: specs/080-paused-label-lanes (claim stub + spec.md; praxisflux TASK-55 counterpart).
<!-- SECTION:NOTES:END -->
