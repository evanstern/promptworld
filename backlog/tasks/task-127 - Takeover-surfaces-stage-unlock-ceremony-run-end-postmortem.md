---
id: TASK-127
title: 'Takeover surfaces: stage-unlock ceremony + run-end postmortem'
status: In Progress
assignee: []
created_date: '2026-07-25 14:44'
updated_date: '2026-07-25 23:50'
labels:
  - learning-game
  - tui
dependencies:
  - TASK-123
priority: medium
ordinal: 97000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient 2026-07-25 (decision 6, D6), Wave 4. One takeover-surface family (body-replacement slot precedent): the ceremony seizes the screen on curriculum.stage_unlocked (identity earned, what it grants, proving evidence, player-authorship voice per D6 — 'your charter proved The Written Word'); the postmortem seizes on run.ended (rubric outcome, report card, epitaphs with charter alignment, retry/fork jump-offs). Voice asymmetry: success speaks player authorship; failure speaks the morgue's no-blame evidence register. Both dismissable and replayable from pull surfaces (?, stages, morgue) — explicit AC. Open questions parked in the synthesis: ambient postmortem contents; ceremony score voice.

Spec: specs/056-takeover-surfaces
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Ceremony takeover on unlock while attached; player-authorship voice; skin-tokened
- [ ] #2 Postmortem takeover on run.ended; morgue-evidence register; retry/fork jump-offs
- [ ] #3 Both replayable from pull surfaces; dismiss is one keypress; never stack
- [ ] #4 Spec phase: Setup
- [ ] #5 Spec phase: Foundational
- [ ] #6 Spec phase: User Story 1 — Postmortem takeover (P1)
- [ ] #7 Spec phase: User Story 2 — Ceremony takeover (P1)
- [ ] #8 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Model tier: Sonnet (spec-implementer default). Rubric: single-package overlay state machine + rendering, tests alongside — routine tier per constitution Principle V. Both parked operator questions verified RESOLVED in the authored overlay pages (ambient postmortem = morgue-only; ceremony = both voices, instrument authoritative) — runbook checkpoint condition not met, proceeding per 'the pages win'. DISPATCH GATED on TASK-121's skin-contract merge (Lane 3 ordering); spec complete and ready.

Dispatched (UI-sweep orchestrator, handoff 2026-07-25b step 3): spec-implementer on Sonnet per recorded rubric; worktree .worktrees/task-127 fast-forwarded to 9386e6a before dispatch. Gate condition met (TASK-121 merged, PR #94). Parallel with TASK-115 (Opus); merge order: smaller first, serial, re-ground between. Implementer warned of pre-existing red TestCatalogSweep on main (TASK-140 hotfix in flight).
<!-- SECTION:NOTES:END -->
