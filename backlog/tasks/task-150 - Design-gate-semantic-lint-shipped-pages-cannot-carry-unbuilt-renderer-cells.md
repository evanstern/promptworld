---
id: TASK-150
title: 'Design-gate semantic lint: shipped pages cannot carry unbuilt renderer cells'
status: In Progress
assignee: []
created_date: '2026-07-26 17:57'
updated_date: '2026-07-26 18:33'
labels:
  - tooling
  - design-corpus
dependencies: []
references:
  - docs/design/reorient-2026-07-26-ui.md
ordinal: 120000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient 2026-07-26 decision 2. The freshness gate validates pins and table headers, never cell content: overlays/postmortem.md shipped with seven 'unbuilt (wave 4)' renderer cells for renderers that exist and are tested. Extend scripts/check-tui-design.mjs to warn/fail when a status: shipped page contains 'unbuilt (wave' in a renderer cell (optional: grep-level check that named renderer symbols exist in internal/tui). Fix postmortem.md ×7 and panels/exercise.md:110 in the same PR. Stale ownership pointers in prose remain a review responsibility (recorded residue class).

Spec: specs/075-design-gate-semantic-lint
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 check-tui-design.mjs flags shipped pages containing 'unbuilt (wave' renderer cells
- [ ] #2 overlays/postmortem.md seven stale cells and panels/exercise.md:110 corrected same-PR
- [ ] #3 Spec phase: Setup
- [ ] #4 Spec phase: Board AC #1 — the lint (US1, P1)
- [ ] #5 Spec phase: Board AC #2 — the corpus stops lying (US2, P1)
- [ ] #6 Spec phase: Grounding + gates (in-branch, per the in-PR doctrine)
- [ ] #7 Spec phase: Post-merge bookkeeping (derived state only)
<!-- AC:END -->



## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Sweep claim (runbook docs/design/reorient-2026-07-26-sweep-runbook.md): spec 075-design-gate-semantic-lint. Tier: Sonnet — single-script lint extension + design-doc cell fixes. Lane A: implementation starts only after TASK-149's PR merges (shared pages: overlays/postmortem.md, panels/exercise.md).
<!-- SECTION:NOTES:END -->
