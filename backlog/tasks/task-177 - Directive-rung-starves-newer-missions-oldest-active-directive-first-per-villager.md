---
id: TASK-177
title: >-
  Directive rung starves newer missions: oldest-active-directive-first per
  villager
status: To Do
assignee: []
created_date: '2026-07-31 02:33'
labels:
  - guardian
  - enhancement
dependencies: []
priority: medium
ordinal: 145000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
When a villager holds directives from two missions, the DIRECTIVE reflex rung picks the OLDEST active directive until it expires, so a newer mission's directive waits behind it — observed live in TASK-158's demo world. Candidate fixes: priority/recency-aware directive selection, or per-mission fairness.

As a player, when I give the guardian a new mission, I expect its directives to compete fairly with older standing work instead of queueing behind it.

Carded from TASK-158's implementation report (spec 107, PR #149), board-sweep-2026-07-29.
<!-- SECTION:DESCRIPTION:END -->
