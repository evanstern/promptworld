---
id: TASK-122
title: Re-measure SC-007 flip-rate at full length after instinct work lands
status: Done
assignee: []
created_date: '2026-07-25 13:58'
updated_date: '2026-07-26 13:43'
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
- [x] #1 ≥4 game-day run on the compound build, spike counting method, results recorded
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Operator checkpoint resolved (2026-07-25): run on a FRESH seeded throwaway world (path-form creation, never touching the worlds registry — the SC-007 Sample A discipline), created with the post-104 binary so it carries the genesis tuning pin + full MVLS stack (062 arbitration, 064 recovery, 061 damper, 058 clamps, 059 watches). Same-tier = gemma4:12b-mlx planner route (the Sample B tier that cleared by only 8%); calibrate before starting; speed 4x for strict Sample-B comparability (~24h wall for 4 game-days) unless calibration proves a higher rung planner-safe — record whichever runs. Measurement: the SC-007 flip-count method verbatim (food {forage,hunt} vs warmth {goto_warmth,build_fire,refuel_fire,warm_up — NOTE: classify the new warm_up verb into the warmth class}, flips + <=200-tick flips + per-game-day normalization); script committed to docs/design/evidence/task-122/ this time, not /tmp. Bar: worst agent <=36 flips/game-day over >=4 game-days; compare against the world-01 baseline AND both SC-007 samples to attribute the compound 043+062+064 effect. Launch blocked on TASK-104's merge (operator will signal).
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
MVLS sweep: checkpoint 5 resolved — fresh seeded world, approved by operator 2026-07-25. Awaiting TASK-104 merge signal to launch.

Measurement RUN LAUNCHED (2026-07-25): fresh world ~/.promptworld/measure/task-122 (path-form, off-registry), seed 4242, created on the post-104 binary (genesis tuning pin: full MVLS defaults). All routes gemma4:12b-mlx @ mbpro-m1 (Sample-B planner tier, zero cloud spend), calibrated (--tier local), daemon pid 15570, speed 4x from tick 1 — >=4 game-days ≈ 24h wall (ETA ~2026-07-26 evening). Counting script committed: docs/design/evidence/task-122/flip_count.py (SC-007 method + warm_up in the warmth class), validated against the world-01 baseline (Sage 73.94/day ≈ spike's ~72).

Speed bumped 4x -> 8x at tick ~5,819 (day 1 07:36, operator decision 2026-07-25): horizon verified green at 8x for all three classes (planner/conversation/meeting thinking, none suppressed). Rationale: bias direction is conservative (staler planner reactions at higher compression push flips UP, so clearing the bar at 8x is stronger evidence); span was 1.6 game-hours old at the switch — effectively single-speed; wall ETA halves to ~11h (~2026-07-26 morning). Evidence doc will footnote the mixed span. Revert condition: any horizon suppression appears.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Full-length re-measure complete (2026-07-26): 5.204 game-days on the compound MVLS build (043+057+058+059+061+062+064), gemma4:12b-mlx all routes @ 8x, seed 4242. Worst agent 6.73 flips/game-day — clears the <=36 bar by 81%, 90.7% below the world-01 baseline (~72). Fast flips: 20 total across 8 agents (baseline Sage alone: 334+). Zero deaths; all villagers at full health+warmth at span end; survival watches never triggered; conversation volume down ~15x. Evidence: docs/design/evidence/task-122/results.md (+ flip_count.py, baseline-validated). Regime caveats recorded (8x plan-staleness — carded TASK-141). Measurement world retained at ~/.promptworld/measure/task-122; daemon stopped.
<!-- SECTION:FINAL_SUMMARY:END -->
