---
id: TASK-155
title: >-
  check-merge-drift: paused-label lanes downgrade to info — praxisflux TASK-55
  counterpart
status: In Progress
assignee:
  - '@claude'
created_date: '2026-07-26 20:11'
updated_date: '2026-07-26 20:11'
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
- [ ] #1 Paused detection reads the 'paused' label from the task file's frontmatter labels: list
- [ ] #2 session mode: a paused task's branch findings (textual/pairwise conflict, stale-base, branch-unpushed, backlog-overlap, spec-number-collision, cleanup-eligible) downgrade to info with the pause as evidence
- [ ] #3 session mode never emits cleanup prescriptions for a paused task's branch/worktree (--apply-cleanup untouched)
- [ ] #4 pr mode: drift findings for a paused task's branch downgrade to info; spec-069 grounding gates (wiki-repin-missing, player-docs-*) stay blocking
- [ ] #5 worktree mode: the downgrade is wired mode-uniformly; ownership protections (spec-number-collision on a paused task's claimed dir, claim mode) stay blocking
- [ ] #6 Covered by scripts/check-merge-drift.test.mjs regression tests
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Claim: card In Progress + specs/080-paused-label-lanes stub (this commit)
2. check-merge-drift.mjs: taskLabels/isTaskPaused helpers reading frontmatter labels:; downgrade-to-info wrapper applied at finding construction in session (per-branch, pairwise = either side) and pr (gated branch) modes; worktree mode wired through the same helper; cleanup prescriptions skipped for paused branches; pause cited as evidence
3. Regression tests in scripts/check-merge-drift.test.mjs (bare-origin + clone fixture per existing convention): paused branch downgrades in session + pr, cleanup not prescribed, unpaused control unchanged, spec-069 gates still block
4. Gates: node --test scripts/, pr mode from this worktree exit 0
<!-- SECTION:PLAN:END -->
