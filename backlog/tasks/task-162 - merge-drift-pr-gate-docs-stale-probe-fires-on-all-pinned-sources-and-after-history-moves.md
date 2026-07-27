---
id: TASK-162
title: >-
  merge-drift pr gate: docs-stale probe fires on all pinned sources and after
  history moves
status: To Do
assignee: []
created_date: '2026-07-27 00:31'
updated_date: '2026-07-27 00:35'
labels: []
dependencies: []
priority: medium
ordinal: 130000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Sibling gap documented by praxis TASK-57 (spec's non-goals explicitly parked this as promptworld's task): the pr-mode docs-stale probe in scripts/check-merge-drift.mjs only invokes the player-docs freshness checker when the branch's own diff touches docs/wiki/ (scripts/check-merge-drift.mjs:1645). Two blind spots: (1) other pinned sources are missed — design-reference pins under docs/design/tui/* get only the non-blocking tui-surface warn (lines 1704-1718), and the checker's own non-wiki inputs (README.md, docs/llm-providers.md, spec 046 sources) never trigger it; (2) history moves (e.g. merging main into a pin-carrying branch) can stale freshness without any pinned-source path appearing in the branch diff, so the probe never runs (observed hazard — see merge-main-into-pin-carrying-branches memory / spec 069 lifecycle). Fix: the probe fires on all pinned sources and after every history move.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 pr-mode invokes the player-docs freshness checker when the branch touches ANY of the checker's pinned inputs (docs/wiki/, README.md, docs/llm-providers.md, spec 046 quickstart sources), not just docs/wiki/
- [ ] #2 design-reference pins (docs/design/tui/*) are freshness-gated in pr mode with a blocking finding, not only the warn-level tui-surface notice
- [ ] #3 the probe also runs after history moves: a branch whose HEAD moved (e.g. merge of main into it) since the last probe is re-checked even when its own diff vs origin/main touches no pinned source
- [ ] #4 gate behavior covered by the script's existing test/fixture approach (synthetic branch cases for each new trigger)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Triage on main @ ce41355 (2026-07-27): blind spot is LATENT, not live — node scripts/check-tui-design.mjs passes all pin checks and the player-docs freshness probe reports 13 fresh / 0 stale. No remediation backlog; the deliverable is the gate fix alone. Origin of the report: praxis TASK-57 (upstream) fixed this in its own doctrine and parked the promptworld sibling as our task.
<!-- SECTION:NOTES:END -->
