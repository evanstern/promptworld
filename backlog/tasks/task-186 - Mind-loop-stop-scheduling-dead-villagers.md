---
id: TASK-186
title: 'Mind loop: stop scheduling dead villagers'
status: To Do
assignee: []
created_date: '2026-08-03 00:15'
labels:
  - mind
  - local-tier
  - scheduling
dependencies: []
ordinal: 168001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The mind loop still plans for villagers who have died, and the sim rejects every one of those intents. TASK-175 closed this hole for sleeping villagers and the dead path never got the same treatment, so those planner calls are still being spent and thrown away.

As a player on a slow local model, I want my limited thinking budget spent on villagers who can actually act, not on ones the gru already killed.

As a player watching the chronicle after a bad night, I do not want the village's attention still going to the dead while the survivors go unattended.

Evidence (TASK-175 T007 soak, 2026-08-01/02, 12.005 game-days, seed 1337 stage-4, recorded in specs/106-sleep-gated-planning/soak.md). The two paths are strikingly asymmetric over the same run:

- '% is asleep' agent.intent_rejected: 0, with 190 sleep-side gate actions (102 'asleep at dequeue' suppressed + 88 'cancelled in flight: agent slept' unusable).
- '% is dead' agent.intent_rejected: 27, with only 2 dead-side gate actions (1 'dead at dequeue' suppressed + 1 'cancelled in flight: agent died' unusable).

So the dead-side machinery EXISTS — spec 106 built both reason strings and both outcome taggings — but it is catching almost nothing, while the sleep side catches everything. That points at the unavailability mirror being updated on sleep/wake edges but not (or not promptly) on death, rather than at a missing gate. specs/106-sleep-gated-planning/soak.md frames dead-agent as a same-shape parity check to the sleep metric; this card is that parity not holding. Not a regression — the dead path simply never received equivalent coverage — and explicitly out of scope for SC-003, which is scoped to sleep. Prior art: TASK-175 / spec 106 (the sleep gate, its unavailability mirror, dequeue gate and in-flight cancel) is the pattern to mirror.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Dead villagers do not consume planner calls: the unavailability mirror reflects death on the same edges and with the same promptness it reflects sleep, so the dequeue gate and in-flight cancel catch dead agents as reliably as sleeping ones
- [ ] #2 A soak of at least 3 game-days shows '% is dead' agent.intent_rejected reduced to near zero from this soak's baseline of 27 over 12 game-days, with dead-side gate actions rising correspondingly
- [ ] #3 Dead-side outcome tagging matches the sleep side exactly: dequeue rows carry outcome=suppressed, in-flight rows carry outcome=unusable
- [ ] #4 The sleep path's measured results are unchanged — no regression in the SC-003 numbers TASK-175 established
<!-- AC:END -->
