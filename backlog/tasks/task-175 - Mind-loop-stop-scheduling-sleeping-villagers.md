---
id: TASK-175
title: 'Mind loop: stop scheduling sleeping villagers'
status: To Do
assignee: []
created_date: '2026-07-30 16:42'
labels: []
dependencies: []
ordinal: 143000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The mind loop keeps planning for villagers who are asleep, and the sim rejects every one of those intents — hundreds of wasted planner cycles on a tier where each call is expensive.

As a player on a slow local model, I want my limited LLM throughput spent on villagers who are awake and can act on the plan.

Evidence (playtest-1): 905 of 1,486 agent.intent_rejected events were "X is asleep" (Ash 131, Birch 127, Cedar 124, Oak 110, ...). Each represents a planner round-trip (~37s avg wall on local gemma) whose output was dead on arrival. The mind should gate planning on sleep state (or the scheduler should skip sleeping agents until wake).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Sleeping villagers do not consume planner calls (mind gates on sleep state or scheduler skips until agent.woke)
- [ ] #2 A soak shows 'is asleep' intent rejections reduced to near zero from playtest-1's baseline of 905/29 days
<!-- AC:END -->
