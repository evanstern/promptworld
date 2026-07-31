---
id: TASK-158
title: >-
  Guardian missions: accept, decompose, pursue, report (the player's standing
  order to the guardian)
status: Done
assignee: []
created_date: '2026-07-26 20:25'
updated_date: '2026-07-31 02:33'
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

Spec: specs/107-guardian-missions
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 spec exists under specs/ and is linked to this task via spec-bridge before implementation starts
- [x] #2 mission artifact is durable and event-sourced; completion derives from designation/directive fulfillment predicates, not model prose
- [x] #3 guardian decomposes and pursues a mission across multiple turns with no player in the loop (rides TASK-112 scheduled cognition)
- [x] #4 failure is honestly reported with recorded-event evidence via the report-card path
- [x] #5 default charter executes missions without editorializing; refusal/personality demonstrably arrives only via skinned charters
- [x] #6 anti-self-grading guard: charter quality measurably changes mission outcomes on a seeded world
- [x] #7 Spec phase: Mission substrate
- [x] #8 Spec phase: Pursuit + doctrine
- [x] #9 Spec phase: Surfaces + tests
- [x] #10 Spec phase: Evidence + grounding
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
board-sweep-2026-07-29 lane 2 final: spec 107 landed + linked (AC1 satisfied). Rulings encoded: EASY-mode default (2026-07-26 firm); IN-BRANCH obedience eval gate (operator, 2026-07-30, TASK-73 precedent — old default must reproduce the TASK-166-observed counsel-loop, new default must execute directly). Doctrine: missions are pre-authorization — full competence at any spec-102 ceiling. Tier: Opus (initiative-frame doctrine, cross-package; draft-runbook A5 call carried).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #149. Mission artifacts (084 discipline, forgery-refused terminals) accepted in plain words, decomposed and pursued on the steward lane via existing verbs at full competence under the ceiling (pre-authorization doctrine, test-pinned + live-proven), completion/failure predicate-derived into the report card. EASY-mode default charter shipped behind the passing in-branch obedience eval (old default counsel-loop reproduced, class-dependent nuance recorded); legacy worlds keep game-authored charter. Live demo msn-11524-0 completed end-to-end with no player in loop; eval+demo worlds preserved. Ratified: TTL 7d default, mission_id on plan verbs, deadline-only failure, stage-1 per 084 precedent. AC6's outcome measurement rides the 164 instrument's mission scenario (harness recorded). Opus tier; spec 107 all tasks done.
<!-- SECTION:FINAL_SUMMARY:END -->
