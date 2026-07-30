---
id: TASK-174
title: >-
  Local-tier JSON robustness: conversation outcomes fail on prose replies from
  small models
status: To Do
assignee: []
created_date: '2026-07-30 16:42'
labels: []
dependencies: []
ordinal: 142000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Small local models sometimes answer the conversation-outcome call with plain prose instead of JSON, and the scene's outcome is thrown away after one retry. Constrained decoding (or a cloud fallback for the outcome step alone) would make local-tier conversations land reliably.

As a player running the village on a home machine, I want conversations between villagers to reliably leave their mark (relationship shifts, memories), even when my local model is a small one.

Evidence (playtest-1, conversation routed to local gemma4:12b via ollama): 22 conversation outcomes failed on bad JSON — replies opening with prose ("invalid character 'F' looking for beginning of value", "no JSON object in reply"); 62 of 293 conversations abandoned (21%); 83 cog.outcome=unusable. Related prior art: TASK-58 fixed the same failure shape for the local planner with JSON-schema structured outputs; the conversation-outcome (and meeting) routes never got that treatment. Cost note: cloud fallback for outcome parsing alone is trivial — the entire 29-day run spent $3.25 of a $100 budget.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Conversation-outcome calls on the local tier use constrained/JSON-mode decoding (as TASK-58 did for the planner) or fall back to cloud on parse failure
- [ ] #2 Outcome parse-failure rate is measurable, and a soak on a small local model shows abandoned-outcome rate materially reduced from playtest-1's baseline (22 failed outcomes / 21% scenes abandoned)
<!-- AC:END -->
