---
id: TASK-109
title: 'Conversation loop damper: find pair-cooldown leak, add novelty gate'
status: In Progress
assignee: []
created_date: '2026-07-25 02:59'
updated_date: '2026-07-25 19:44'
labels:
  - behavior-hygiene
  - mvls
dependencies: []
priority: high
ordinal: 17000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
World-01 evidence: Birch<->Sage held 219 conversation scenes / 343 talks in 6.2 days despite encounterCooldownTicks=7200 and talkCooldownSec=7200. DIAGNOSED (2026-07-25, docs/design/evidence/task-109/): the planner talk_to -> hail -> hailStep founding path carries NO pair-frequency gate — 99.1% of Birch<->Sage scenes and 97.8% of ALL world-01 scenes were hail-founded; canTalk gates only the ambient beat, pairSeen gates only encounter arming, hailStep bypasses canTalk by TASK-47 design. Systemic, all pairs. Operator decision (2026-07-25): pair cooldown lives SIM-SIDE on the hail founding path (event-sourced PairTalks state, reusing the encounter_cooldown_ticks dial); novelty gate stays mind-side as a clearly-marked removable SHIM (user 2026-07-24: compensates for weak model-side variety — first place to look if conversations feel less dynamic; remove when model tiers make it unnecessary). Related: TASK-89 (gist confabulation is model-tier).

Spec: specs/061-conversation-loop-damper
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 root cause of the cooldown bypass identified and written on this task
- [x] #2 pair re-conversation gated by cooldown AND novelty (new memory above salience floor)
- [x] #3 last-gist anti-repeat context in scene prompt
- [x] #4 shim documented as removable with a pointer comment at the gate site
- [x] #5 Spec phase: Foundational — pair record (US2)
- [x] #6 Spec phase: User Story 1 — sim-side hail cooldown (P1)
- [x] #7 Spec phase: User Story 3 — novelty SHIM (P2)
- [ ] #8 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
MVLS sweep dispatch (2026-07-25): implementer tier Opus 4.8 — constitution V rubric: internal/mind orchestration (scene founding) + doctrine-adjacent sim gate + cross-package event-sourced state. Diagnosis-first checkpoint cleared with operator (sim-side gate chosen). Lane 1; merges after 108.

spec-bridge sync: Foundational — pair record (US2): 2/2 · User Story 1 — sim-side hail cooldown (P1): 3/3 · User Story 3 — novelty SHIM (P2): 4/4 · Polish & Cross-Cutting: 1/2

PR #91 squash-merged as 1debe18. Root cause on this task + docs/design/evidence/task-109/; shim removal note also at specs/061-conversation-loop-damper/shim-note.md.
<!-- SECTION:NOTES:END -->
