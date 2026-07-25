---
id: TASK-104
title: Needs-conditioned recovery intents with parameters
status: To Do
assignee: []
created_date: '2026-07-25 02:41'
updated_date: '2026-07-25 18:18'
labels:
  - goal-quality
  - instinct-layer
  - mvls
dependencies:
  - TASK-103
priority: high
ordinal: 16000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Direction B from spike TASK-101. Recovery goals complete on the NEED, not the location: e.g. warm_up(until_warmth>=N) loiters at the fire until the condition holds, mirroring eat-to-satiety. Generalize: parameterized intent arguments (tool args carry the completion condition) rather than new one-off verbs — flexible and generalizable per Evan's note. Kills the arrive→idle→reflex-vacuum cycle that manufactures the oscillation. Non-trivial: full Spec Kit before implementation.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 At least warm_up (and pattern for rest/food analogs) completes on a need condition passed as an argument
- [ ] #2 Idle-at-recovery-site no longer triggers instinct dispatch mid-recovery
- [ ] #3 Deterministic sim test covers recover-then-release behavior
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Design together with TASK-103: a villager loitering-to-recover must not register as idle to the instinct layer (the 120-tick idle grace is what let the larder rule hijack Sage at the fire).
<!-- SECTION:NOTES:END -->
