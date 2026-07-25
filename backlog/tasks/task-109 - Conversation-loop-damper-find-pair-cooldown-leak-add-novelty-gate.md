---
id: TASK-109
title: 'Conversation loop damper: find pair-cooldown leak, add novelty gate'
status: To Do
assignee: []
created_date: '2026-07-25 02:59'
updated_date: '2026-07-25 03:10'
labels: []
dependencies: []
priority: high
ordinal: 17000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
World-01 evidence: Birch<->Sage held 177 conversations (74 in one day, one per ~20 game-min) despite encounterCooldownTicks=7200 (mind/mind.go:36) and talkCooldownSec=7200 (sim/agents.go:543) — something arms scenes around the cooldown; diagnose which path (planner talk_to intents vs encounter arming) before designing. Then: (1) fix the leak; (2) novelty gate — a pair cannot re-converse until one of them has a new memory above a salience floor since their last exchange, and the last convo gist enters the scene prompt as 'this already happened'. NOTE (user, 2026-07-24): the novelty gate is a SHIM compensating for weak model-side variety — mark it clearly in code and docs so that if conversations later feel less dynamic than wanted, this is the first place to look, and remove it when model tiers make it unnecessary. Related: TASK-89 (gist confabulation is model-tier).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 root cause of the cooldown bypass identified and written on this task
- [ ] #2 pair re-conversation gated by cooldown AND novelty (new memory above salience floor)
- [ ] #3 last-gist anti-repeat context in scene prompt
- [ ] #4 shim documented as removable with a pointer comment at the gate site
<!-- AC:END -->
