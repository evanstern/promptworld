---
id: TASK-141
title: >-
  Plan staleness budget doesn't scale with clock speed — planning silently dies
  above ~4x on local tiers
status: To Do
assignee: []
created_date: '2026-07-26 02:33'
updated_date: '2026-07-26 02:34'
labels:
  - mvls
  - behavior-hygiene
dependencies: []
priority: medium
ordinal: 111000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Found during the TASK-122 measurement run (2026-07-25, world ~/.promptworld/measure/task-122, gemma4:12b-mlx all-routes @ 8x): set_plan landed 27 times vs 287 rejected-stale (~85% dead) — 'rejected-stale: staleness ~1900-2300 > budget 1200'. Cause: the staleness budget is a FIXED game-tick constant while thought latency is wall-clock — a ~250s gemma plan thought (incl. endpoint queueing at parallel=4, 8 villagers) is ~1000 game-ticks at 4x (lands) but ~2000 at 8x (dies). Above ~4x, planning is structurally dead for local tiers while the horizon still reports 'planner thinking' (the horizon gates scheduling, not landing staleness). Sample-A (32x) had the same regime. Fix space (spec should decide): scale the budget with clock speed, derive it from the calibrated s/pt like the horizon does, or promote it as a tuning.json dial (spec 048 path) — mind the replay implications either way (landing outcomes are events; a speed-dependent gate must stay a pure function of event-sourced state, and the clock speed IS event-sourced). Evidence queries in the TASK-122 notes; reproducible against the measure world's log.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Diagnosis pinned: the exact constant + consumption site named (file:symbol) with the latency math
- [ ] #2 Chosen mechanism keeps planning viable at 8x on a calibrated local tier (proven by test or measured run) without breaking replay determinism
- [ ] #3 Horizon/status surfaces stop reporting 'planner thinking' as healthy when plan landing is structurally dead (or the gap is explicitly documented)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Diagnosis pin (AC#1 progress): the rejection fires at internal/sim/landing.go rungStale ('staleness %d > budget %d', dc.BudgetTicks) via OutcomeRejectedStale (internal/sim/cognition.go). The per-class BudgetTicks value (set_plan class = 1200) is the constant that doesn't scale with clock speed — full consumption-site map is the spec's first task.
<!-- SECTION:NOTES:END -->
