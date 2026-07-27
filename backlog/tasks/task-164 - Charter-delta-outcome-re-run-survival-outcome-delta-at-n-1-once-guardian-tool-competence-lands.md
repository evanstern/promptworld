---
id: TASK-164
title: >-
  Charter-delta outcome re-run: survival outcome delta at n>1 once guardian tool
  competence lands
status: To Do
assignee: []
created_date: '2026-07-27 13:20'
labels:
  - mvls
  - guardian-survival
dependencies:
  - TASK-163
priority: medium
ordinal: 132000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Carded from TASK-137's results (2026-07-27, docs/design/evidence/task-137/results.md): the charter-delta experiment proved charter quality changes guardian BEHAVIOR (default arm 0 landed interventions / 3 attempts vs authored arm 2 landed / 7 attempts, same seed + harsh dials) but the survival OUTCOME delta was null at n=1 (one starvation death each, same villager), and the binding constraint was tool-call competence (80% of privileged attempts gate-rejected — TASK-163). Re-run the A/B once that floor is fixed, so outcome attribution measures the charter surface rather than the model's tool fumbling. Evidence-doc re-run suggestions: sequential arms acceptable; n>=2 seeds or a longer horizon. Feeds TASK-111 AC#5's outstanding outcome-delta caveat and the TASK-112 dispatch checkpoint (docs/design/guardian-directives-runbook.md).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 OPERATOR CHECKPOINT: eval spend approved before dispatch (the TASK-122/TASK-137 treatment)
- [ ] #2 Both arms re-run after the TASK-163 competence fix, same-seed pairing, n>=2 seeds or an operator-approved design; survival outcome delta measured
- [ ] #3 Results doc recorded under docs/design/evidence/; TASK-111 AC#5 outcome-delta caveat updated with the new evidence
<!-- AC:END -->
