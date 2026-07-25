---
id: TASK-108
title: 'Survival reflex layer: build-fire reflex + earlier refuel threshold'
status: In Progress
assignee: []
created_date: '2026-07-25 02:59'
updated_date: '2026-07-25 18:46'
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
- [ ] #1 cold villager with wood and no reachable warmth builds a fire via reflex, no model call
- [ ] #2 refuel reflex arms below ~3h fuel instead of 1h
- [ ] #3 reflex-layer audit: note any other survival gap found (eat/sleep/warmth parity)
- [ ] #4 thresholds sourced from tuning manifest when TASK-107 lands (const fallback until then)
- [ ] #5 Spec phase: Foundational
- [ ] #6 Spec phase: User Story 1 — Refuel arms below 3 game-hours (P1)
- [ ] #7 Spec phase: User Story 2 — Genesis tuning pin (P1)
- [ ] #8 Spec phase: User Story 3 — Cold build-fire reflex proven (P2)
- [ ] #9 Spec phase: User Story 4 — Survival audit (P3)
- [ ] #10 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
MVLS sweep dispatch (2026-07-25): implementer tier Opus 4.8 — constitution V rubric: doctrine-adjacent survival behavior in the sim reducer, replay-affecting genesis change. Lane 1; merges after TASK-110.
<!-- SECTION:NOTES:END -->
