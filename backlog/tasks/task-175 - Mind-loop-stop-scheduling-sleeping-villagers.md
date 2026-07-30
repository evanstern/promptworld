---
id: TASK-175
title: 'Mind loop: stop scheduling sleeping villagers'
status: In Progress
assignee: []
created_date: '2026-07-30 16:42'
updated_date: '2026-07-30 18:36'
labels: []
dependencies: []
ordinal: 143000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The mind loop keeps planning for villagers who are asleep, and the sim rejects every one of those intents — hundreds of wasted planner cycles on a tier where each call is expensive.

As a player on a slow local model, I want my limited LLM throughput spent on villagers who are awake and can act on the plan.

Evidence (playtest-1): 905 of 1,486 agent.intent_rejected events were "X is asleep" (Ash 131, Birch 127, Cedar 124, Oak 110, ...). Each represents a planner round-trip (~37s avg wall on local gemma) whose output was dead on arrival. The mind should gate planning on sleep state (or the scheduler should skip sleeping agents until wake).

Spec: specs/106-sleep-gated-planning
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Sleeping villagers do not consume planner calls (mind gates on sleep state or scheduler skips until agent.woke)
- [ ] #2 A soak shows 'is asleep' intent rejections reduced to near zero from playtest-1's baseline of 905/29 days
- [ ] #3 Spec phase: Unavailability mirror + pre-submit gate (US1 layer 1)
- [ ] #4 Spec phase: In-flight cancel (US1 layer 2)
- [ ] #5 Spec phase: Wake resumption + regression (US2)
- [ ] #6 Spec phase: Soak evidence (card AC #2)
- [ ] #7 Spec phase: Reconcile with spec 102 + grounding + gates
<!-- AC:END -->



## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
sweep dispatch (runbook playtest-1-findings-sweep): tier Opus 4.8 — scheduling logic in internal/mind orchestration, named explicitly by constitution P.V's Opus rubric; small diff but edits the mind driver TASK-112's branch also touches. Spec 106-sleep-gated-planning. LANE 2: PR merges only after TASK-112's PR lands (operator ruling), after TASK-172.
<!-- SECTION:NOTES:END -->
