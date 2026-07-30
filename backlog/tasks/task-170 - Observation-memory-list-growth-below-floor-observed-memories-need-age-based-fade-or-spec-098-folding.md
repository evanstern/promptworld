---
id: TASK-170
title: >-
  Observation-memory list growth: below-floor observed memories need age-based
  fade or spec-098 folding
status: To Do
assignee: []
created_date: '2026-07-30 02:38'
labels:
  - memory
  - debt
dependencies: []
priority: medium
ordinal: 138000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Spec 097's soak measured ~245 observation memories/agent/day accreting in the memory LIST (working window bounded, list not); consolidation faded only 56/day. Candidate fixes: age-based auto-fade for below-floor OriginObserved memories, or fold them into spec 098's habituation clustering.

As a villager, my memory of ordinary arrivals should fade with time instead of accumulating forever.

Carded from TASK-80's implementation report (spec 097, PR #141) + evidence at docs/design/evidence/task-80/results.md.
<!-- SECTION:DESCRIPTION:END -->
