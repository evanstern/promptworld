---
id: TASK-163
title: >-
  Guardian tool-call competence: raise the privileged-action landing rate above
  the 80% rejection floor
status: In Progress
assignee: []
created_date: '2026-07-27 13:20'
updated_date: '2026-07-27 14:17'
labels:
  - mvls
  - guardian-survival
dependencies: []
priority: high
ordinal: 131000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Carded from TASK-136's live measurement (2026-07-27, docs/design/evidence/task-137/results.md): 8 of 10 guardian privileged-action attempts were rejected at the targeting/validity door (80%; 3/3 default arm, 5/7 authored arm) — the spec-059 targeting digest shipped and functions, but rejections did not drop toward 0 because every failure was model-side tool-call incompetence on gemma4:12b native tool-calls: send_vision place grants missing place_x/place_y (2), hallucinated tree coords (1), work_miracle give_item with non-item kind 'fire'/empty (3), work_miracle move with empty entity class (2). This floor is the mechanical-noise/confound bound for all prompt-attribution comparisons (TASK-137 charter delta, TASK-67 fork duel) and a checkpoint for TASK-112 dispatch (docs/design/guardian-directives-runbook.md). Candidate mitigations from the evidence doc, not pre-decided: stronger model tier for metatron_watch; prompt shaping so the guardian reads coordinates from its watch-context digest before granting places; schema/argument validation with in-turn repair retry. Event evidence: ~/.promptworld/measure/task-137-{default,authored}/world.db (cog.tool_call verdict/reason fields).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Failure taxonomy triaged to a chosen mitigation (model tier, prompt shaping, arg-repair retry, or a combination) with the decision and rationale recorded on this card
- [ ] #2 Post-fix run with >=10 privileged-action attempts: rejection rate measured, materially below the 80% baseline, evidence doc recorded
- [ ] #3 The new rate replaces 80% as the quoted confound bound for prompt-attribution comparisons (TASK-137 evidence doc updated; TASK-67 duel scoreboard cites it)
<!-- AC:END -->
