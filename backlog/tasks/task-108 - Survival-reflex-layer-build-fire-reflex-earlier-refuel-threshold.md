---
id: TASK-108
title: 'Survival reflex layer: build-fire reflex + earlier refuel threshold'
status: To Do
assignee: []
created_date: '2026-07-25 02:59'
updated_date: '2026-07-25 03:10'
labels: []
dependencies:
  - TASK-107
priority: high
ordinal: 13000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Doctrine decision (user, 2026-07-24): self-preservation is table stakes — villagers are simple people who at minimum know how to not die; err toward survival instinct wherever the reflex layer has a gap. World-01 evidence (docs/design/control-surface-and-calibration.md §3.1): 8 fires built vs 42 burnouts over 6 days, warmth 848->82 on day 7, Oak died of exposure, 425 intents rejected 'no warmth anywhere'. Changes: (a) raise refuelDyingBelow 3600->~10800 (sim/agents.go:571); (b) new deterministic cold-reflex: cold night + no reachable warmth + carrying >=2 wood -> build fire (reflex block sim/agents.go:515-545); (c) leave fireBurnPerWood alone for now — one lever at a time. Thresholds should ride the tuning manifest (TASK-107).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 cold villager with wood and no reachable warmth builds a fire via reflex, no model call
- [ ] #2 refuel reflex arms below ~3h fuel instead of 1h
- [ ] #3 reflex-layer audit: note any other survival gap found (eat/sleep/warmth parity)
- [ ] #4 thresholds sourced from tuning manifest when TASK-107 lands (const fallback until then)
<!-- AC:END -->
