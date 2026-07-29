---
id: TASK-162
title: >-
  merge-drift pr gate: docs-stale probe fires on all pinned sources and after
  history moves
status: Done
assignee: []
created_date: '2026-07-27 00:31'
updated_date: '2026-07-29 19:12'
labels: []
dependencies: []
priority: medium
ordinal: 130000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Sibling gap documented by praxis TASK-57 (spec's non-goals explicitly parked this as promptworld's task): the pr-mode docs-stale probe in scripts/check-merge-drift.mjs only invokes the player-docs freshness checker when the branch's own diff touches docs/wiki/ (scripts/check-merge-drift.mjs:1645). Two blind spots: (1) other pinned sources are missed — design-reference pins under docs/design/tui/* get only the non-blocking tui-surface warn (lines 1704-1718), and the checker's own non-wiki inputs (README.md, docs/llm-providers.md, spec 046 sources) never trigger it; (2) history moves (e.g. merging main into a pin-carrying branch) can stale freshness without any pinned-source path appearing in the branch diff, so the probe never runs (observed hazard — see merge-main-into-pin-carrying-branches memory / spec 069 lifecycle). Fix: the probe fires on all pinned sources and after every history move.

Spec: specs/088-pr-gate-docs-stale-probe
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 pr-mode invokes the player-docs freshness checker when the branch touches ANY of the checker's pinned inputs (docs/wiki/, README.md, docs/llm-providers.md, spec 046 quickstart sources), not just docs/wiki/
- [x] #2 design-reference pins (docs/design/tui/*) are freshness-gated in pr mode with a blocking finding, not only the warn-level tui-surface notice
- [x] #3 the probe also runs after history moves: a branch whose HEAD moved (e.g. merge of main into it) since the last probe is re-checked even when its own diff vs origin/main touches no pinned source
- [x] #4 gate behavior covered by the script's existing test/fixture approach (synthetic branch cases for each new trigger)
- [x] #5 Spec phase: Foundational (trigger plumbing)
- [x] #6 Spec phase: User Story 1 — non-wiki pinned inputs gate (P1)
- [x] #7 Spec phase: User Story 3 — history moves re-trigger (P2)
- [x] #8 Spec phase: User Story 2 — design-reference pins block (P2)
- [x] #9 Spec phase: Polish & verification
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Triage on main @ ce41355 (2026-07-27): blind spot is LATENT, not live — node scripts/check-tui-design.mjs passes all pin checks and the player-docs freshness probe reports 13 fresh / 0 stale. No remediation backlog; the deliverable is the gate fix alone. Origin of the report: praxis TASK-57 (upstream) fixed this in its own doctrine and parked the promptworld sibling as our task.

board-sweep-2026-07-29 lane 0: spec 088 authored on branch task-162-pr-gate-docs-stale-probe (spec/plan/tasks committed 1d4a26f); linked via spec-bridge. Implementation tier: Sonnet — single-script tooling change with fixture tests (routine slice per constitution Principle V); escalation trigger: none foreseen.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #131 (merge commit 259dd0f). pr-mode docs-stale probe now fires on all player-docs pinned inputs (derived from promptworld-docs:source tags), after history moves (merge in origin/main..tip), and design-reference pin drift blocks via check-tui-design delegation (tui-design-stale/env-error). Fixture matrix F1-F9; 31/31 + 10/10 tests; exit-code contract unchanged. Spec 088 all 8 tasks done (spec-bridge derived Done-eligible).
<!-- SECTION:FINAL_SUMMARY:END -->
