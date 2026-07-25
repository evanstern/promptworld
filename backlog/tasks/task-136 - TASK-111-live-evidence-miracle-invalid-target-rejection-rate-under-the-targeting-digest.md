---
id: TASK-136
title: >-
  TASK-111 live evidence: miracle invalid-target rejection rate under the
  targeting digest
status: To Do
assignee: []
created_date: '2026-07-25 19:37'
labels:
  - mvls
  - guardian-survival
dependencies: []
priority: medium
ordinal: 106000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Follow-up carded from the MVLS sweep (operator request 2026-07-25) so TASK-111's AC3 second clause doesn't rot: 'invalid-target rejections drop to ~0' needs live-world observation the merged code cannot prove. Mechanism shipped in spec 059 / PR #90 (token-bounded positions+passability digest in miracle-capable prompts; door round-trip regression green). Work: run a world on the post-059 binary long enough for survival/miracle turns to accumulate (world-01 after its daemon upgrades, or a seeded scratch world with induced crises), then measure the door-rejection rate on angel miracle coordinates vs world-01's pre-059 baseline (3 of 4 rejected). Tick TASK-111 AC#3 with the evidence and close it out.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A post-059 run with >=5 angel miracle attempts measured; invalid-coordinate door rejections at or near 0, evidence recorded here and on TASK-111
<!-- AC:END -->
