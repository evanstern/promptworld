---
id: TASK-158
title: >-
  Guardian missions: accept, decompose, pursue, report (the player's standing
  order to the guardian)
status: In Progress
assignee: []
created_date: '2026-07-26 20:25'
updated_date: '2026-07-31 01:14'
labels:
  - learning-game
dependencies:
  - TASK-112
  - TASK-157
priority: medium
ordinal: 127000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The player-facing loop on top of the TASK-157 substrate: 'Guardian, direct the villagers to settle near (x,y)' becomes a durable, event-sourced MISSION artifact. The guardian decomposes it — survey the land, place designations, issue directives — then monitors its own directives via the existing standing-order machinery (directive.* events are in observableEventTypes per TASK-157) and reports done/failed with event evidence through the report-card producer. Mission completion is derived from the designation/directive fulfillment predicates, never self-graded prose.

Doctrine: a mission is durable pre-authorization — the same legal shape as a standing order — so NO initiative-frame relaxation is needed; pursuit turns are pre-authorized by the mission the player issued. Requires TASK-112 (agentization/scheduled cognition) for multi-turn follow-through: today the guardian only thinks on player chat or order match.

EASY-MODE DEFAULT (operator decision, 2026-07-26): the default guardian does what the player asks without question — obedience lives in the compiled default charter ('execute the player's missions without editorializing'); personality, refusals, and counsel-first behavior are skinned-guardian data (TASK-121 substrate), never the default. Anti-self-grading guard carries over from TASK-111/112: charter quality must measurably change mission outcomes.

Grounding: docs/wiki/guardian-watch-workers.md (monitor machinery), docs/wiki/guardian-report-card.md (evidence-cited reporting), docs/design/learning-game-synthesis.md (three-lane initiative frame; missions extend the ambition lane's pre-authorization contract). Feature ideation session 2026-07-26.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 spec exists under specs/ and is linked to this task via spec-bridge before implementation starts
- [ ] #2 mission artifact is durable and event-sourced; completion derives from designation/directive fulfillment predicates, not model prose
- [ ] #3 guardian decomposes and pursues a mission across multiple turns with no player in the loop (rides TASK-112 scheduled cognition)
- [ ] #4 failure is honestly reported with recorded-event evidence via the report-card path
- [ ] #5 default charter executes missions without editorializing; refusal/personality demonstrably arrives only via skinned charters
- [ ] #6 anti-self-grading guard: charter quality measurably changes mission outcomes on a seeded world
<!-- AC:END -->
