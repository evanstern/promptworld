---
id: TASK-108
title: 'Survival reflex layer: build-fire reflex + earlier refuel threshold'
status: Done
assignee: []
created_date: '2026-07-25 02:59'
updated_date: '2026-07-25 22:09'
labels:
  - instinct-layer
  - mvls
dependencies:
  - TASK-107
priority: high
ordinal: 13000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Doctrine decision (user, 2026-07-24): self-preservation is table stakes — villagers are simple people who at minimum know how to not die; err toward survival instinct wherever the reflex layer has a gap. World-01 evidence (docs/design/control-surface-and-calibration.md §3.1): 8 fires built vs 42 burnouts over 6 days, warmth 848->82 on day 7, Oak died of exposure, 425 intents rejected 'no warmth anywhere'. Changes: (a) raise refuelDyingBelow 3600->~10800 (now the defaultRefuelDyingBelow dial default, sim/tuning.go per spec 048); (b) cold build-fire reflex — verified against the post-041 ladder (night rung exists; spec 057 proves it and closes exposed gaps); (c) leave fireBurnPerWood alone — one lever at a time. Sweep grounding 2026-07-25: task's code pins predated specs 041/043/048; spec 057 carries the reality-checked scope incl. the genesis tuning pin (first default change since 048).

Spec: specs/057-survival-reflex-gaps
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 cold villager with wood and no reachable warmth builds a fire via reflex, no model call
- [x] #2 refuel reflex arms below ~3h fuel instead of 1h
- [x] #3 reflex-layer audit: note any other survival gap found (eat/sleep/warmth parity)
- [x] #4 thresholds sourced from tuning manifest when TASK-107 lands (const fallback until then)
- [x] #5 Spec phase: Foundational
- [x] #6 Spec phase: User Story 1 — Refuel arms below 3 game-hours (P1)
- [x] #7 Spec phase: User Story 2 — Genesis tuning pin (P1)
- [x] #8 Spec phase: User Story 3 — Cold build-fire reflex proven (P2)
- [x] #9 Spec phase: User Story 4 — Survival audit (P3)
- [x] #10 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
MVLS sweep dispatch (2026-07-25): implementer tier Opus 4.8 — constitution V rubric: doctrine-adjacent survival behavior in the sim reducer, replay-affecting genesis change. Lane 1; merges after TASK-110.

spec-bridge sync: Foundational: 1/1 · User Story 1 — Refuel arms below 3 game-hours (P1): 3/3 · User Story 2 — Genesis tuning pin (P1): 4/4 · User Story 3 — Cold build-fire reflex proven (P2): 2/2 · User Story 4 — Survival audit (P3): 1/1 · Polish & Cross-Cutting: 1/2

PR #89 squash-merged as 483e90c. Matrix exposed no surgical gap; three non-surgical gaps carded in audit.md (night search + day warmth -> TASK-103; wake-to-cold -> TASK-104). T013 (player-docs refresh) batched to end of lane-1 merge train.

spec-bridge sync: Foundational: 1/1 · User Story 1 — Refuel arms below 3 game-hours (P1): 3/3 · User Story 2 — Genesis tuning pin (P1): 4/4 · User Story 3 — Cold build-fire reflex proven (P2): 2/2 · User Story 4 — Survival audit (P3): 1/1 · Polish & Cross-Cutting: 2/2 — status In Progress → Done
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Foundational: 1/1 · User Story 1 — Refuel arms below 3 game-hours (P1): 3/3 · User Story 2 — Genesis tuning pin (P1): 4/4 · User Story 3 — Cold build-fire reflex proven (P2): 2/2 · User Story 4 — Survival audit (P3): 1/1 · Polish & Cross-Cutting: 2/2). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
