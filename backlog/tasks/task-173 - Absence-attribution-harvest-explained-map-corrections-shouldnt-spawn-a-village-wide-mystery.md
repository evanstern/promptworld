---
id: TASK-173
title: >-
  Absence attribution: harvest-explained map corrections shouldn't spawn a
  village-wide mystery
status: To Do
assignee: []
created_date: '2026-07-30 16:41'
labels: []
dependencies: []
ordinal: 141000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
When a villager finds a remembered tree or rock gone, nothing lets them (or the narrator) infer the mundane explanation — that a neighbor harvested it — so ordinary resource consumption reads as an unexplained phenomenon and can take over the story.

As a player, I want real mysteries to stand out, not be drowned by my villagers panicking about firewood their neighbors chopped.
As a villager in the game, when I hear Cedar has been felling trees all week, arriving at a stump should connect those dots instead of feeding cosmic dread.

Evidence (playtest-1, 29 game-days): the dominant chronicle thread all 29 days was a "vanishing landscape" horror storyline. Cross-check: ALL 780 distinct "vanished" locations match an agent.chopped/agent.quarried event exactly — zero genuine anomalies. 2,932 agent.map_corrected events (~100/day, never declining) each fed the narrative, while chop-rumors were simultaneously circulating socially (social.rumor_told, social.place_told).

Scope note: spec 097 (perception of absence — dedup, disconfirmation decay) merged after this run. First step is a v6 re-run of the same scenario to measure what 097 already absorbs; the remaining gap is attribution — grounding a correction against known harvest activity (own memories, rumors) before it earns mystery-grade salience/narration.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A v6 re-run of the playtest-1 scenario is measured: rate of map corrections narrated as anomalies, before/after comparison recorded on this task
- [ ] #2 A map correction explainable by known harvest activity (witnessed or rumored) is attributed as mundane and does not earn mystery-grade narrative weight
- [ ] #3 Genuinely unexplained absences still surface as noteworthy (the guardian's real mysteries are not suppressed)
<!-- AC:END -->
