---
id: TASK-103
title: Instinct yields to intelligence — reflex/planner arbitration
status: To Do
assignee: []
created_date: '2026-07-25 02:41'
updated_date: '2026-07-25 18:18'
labels:
  - goal-quality
  - instinct-layer
  - mvls
dependencies:
  - TASK-107
  - TASK-108
priority: high
ordinal: 15000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Direction A from spike TASK-101. Reframe the reflex ladder as 'instinct' that YIELDS to intelligence: reflex prep rules (larder-stocking, refuel) must not counter-schedule against a recent planner intent or fire while any need is in a danger band; add a warmth check to the reflex's day branch (currently absent — policy.go day branch never considers warmth); prune/soften hard-coded prep behavior where the planner can own it. Evidence: world-01 forage↔goto_warmth thrash (Sage 436 flips) is reflex-vs-planner counter-scheduling, not LLM indecision. Non-trivial: full Spec Kit before implementation.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Reflex never overrides a live/recent planner decision except on survival-threshold breach
- [ ] #2 Day-branch warmth gap closed
- [ ] #3 Replay/sim test demonstrates Sage-style thrash episode no longer occurs
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
RECONCILIATION with TASK-108 (2026-07-24): no conflict once the ladder is split into two kinds of instinct. SURVIVAL instinct (don't die: eat-when-starving, warmth-at-night, 108's new cold-reflex build-fire, refuel) gets MORE authority per the table-stakes doctrine; HOUSEKEEPING instinct (larder stocking, prep chopping, wander) is what yields to intelligence — it was a housekeeping rule (day-branch larder stock, policy.go:96) that caused the world-01 thrash by preempting survival recovery. Sequence: 107 (tuning manifest) → 108 (survival gaps, surgical) → this task (yield semantics for housekeeping + day-branch warmth check). 108's thresholds and this task's danger bands both ride the tuning manifest. Design together with TASK-104: recovery-loitering must not read as 'idle' to any instinct rule.
<!-- SECTION:NOTES:END -->
