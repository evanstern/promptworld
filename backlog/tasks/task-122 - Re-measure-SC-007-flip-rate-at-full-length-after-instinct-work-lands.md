---
id: TASK-122
title: Re-measure SC-007 flip-rate at full length after instinct work lands
status: To Do
assignee: []
created_date: '2026-07-25 13:58'
updated_date: '2026-07-25 20:49'
labels:
  - goal-quality
  - thrash-detection
  - mvls
dependencies: []
priority: low
ordinal: 9000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Follow-on from TASK-105/spec 043 evidence (specs/043-context-grounding/evidence/sc-007-fliprate.md): both post-043 samples cleared the ≤36 flips/game-day bar but spans were sub-1-game-day and Sample B (gemma4:12b-mlx at 4x) cleared by only 8%. After TASK-103 (instinct yields to intelligence) and TASK-104 (needs-conditioned recovery) land, run a full-length (≥4 game-day) same-tier measurement to tighten the estimate and attribute the compound effect.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 ≥4 game-day run on the compound build, spike counting method, results recorded
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Operator checkpoint resolved (2026-07-25): run on a FRESH seeded throwaway world (path-form creation, never touching the worlds registry — the SC-007 Sample A discipline), created with the post-104 binary so it carries the genesis tuning pin + full MVLS stack (062 arbitration, 064 recovery, 061 damper, 058 clamps, 059 watches). Same-tier = gemma4:12b-mlx planner route (the Sample B tier that cleared by only 8%); calibrate before starting; speed 4x for strict Sample-B comparability (~24h wall for 4 game-days) unless calibration proves a higher rung planner-safe — record whichever runs. Measurement: the SC-007 flip-count method verbatim (food {forage,hunt} vs warmth {goto_warmth,build_fire,refuel_fire,warm_up — NOTE: classify the new warm_up verb into the warmth class}, flips + <=200-tick flips + per-game-day normalization); script committed to docs/design/evidence/task-122/ this time, not /tmp. Bar: worst agent <=36 flips/game-day over >=4 game-days; compare against the world-01 baseline AND both SC-007 samples to attribute the compound 043+062+064 effect. Launch blocked on TASK-104's merge (operator will signal).
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
MVLS sweep: checkpoint 5 resolved — fresh seeded world, approved by operator 2026-07-25. Awaiting TASK-104 merge signal to launch.
<!-- SECTION:NOTES:END -->
