---
id: TASK-141
title: >-
  Plan staleness budget doesn't scale with clock speed — planning silently dies
  above ~4x on local tiers
status: In Progress
assignee: []
created_date: '2026-07-26 02:33'
updated_date: '2026-07-26 15:16'
labels:
  - mvls
  - behavior-hygiene
dependencies: []
priority: medium
ordinal: 111000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Found during the TASK-122 measurement run (2026-07-25, world ~/.promptworld/measure/task-122, gemma4:12b-mlx all-routes @ 8x): set_plan landed 27 times vs 287 rejected-stale (~85% dead) — 'rejected-stale: staleness ~1900-2300 > budget 1200'. Cause: the staleness budget is a FIXED game-tick constant while thought latency is wall-clock — a ~250s gemma plan thought (incl. endpoint queueing at parallel=4, 8 villagers) is ~1000 game-ticks at 4x (lands) but ~2000 at 8x (dies). Above ~4x, planning is structurally dead for local tiers while the horizon still reports 'planner thinking' (the horizon gates scheduling, not landing staleness). Sample-A (32x) had the same regime. Fix space (spec should decide): scale the budget with clock speed, derive it from the calibrated s/pt like the horizon, or promote it as a tuning.json dial (spec 048 path) — mind the replay implications either way (landing outcomes are events; a speed-dependent gate must stay a pure function of event-sourced state, and the clock speed IS event-sourced). Evidence queries in the TASK-122 notes; reproducible against the measure world's log.

Spec: specs/067-staleness-budget-scaling
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Diagnosis pinned: the exact constant + consumption site named (file:symbol) with the latency math
- [ ] #2 Chosen mechanism keeps planning viable at 8x on a calibrated local tier (proven by test or measured run) without breaking replay determinism
- [ ] #3 Horizon/status surfaces stop reporting 'planner thinking' as healthy when plan landing is structurally dead (or the gap is explicitly documented)
- [x] #4 Spec phase: Setup
- [x] #5 Spec phase: Foundational (blocking prerequisites)
- [x] #6 Spec phase: User Story 1 — Planning survives 8x on a calibrated local tier (P1)
- [x] #7 Spec phase: User Story 2 — Replay determinism across speed changes (P1)
- [ ] #8 Spec phase: User Story 3 — Status surfaces stop calling structural death "thinking" (P2)
- [ ] #9 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Diagnosis pin (AC#1 progress): the rejection fires at internal/sim/landing.go rungStale ('staleness %d > budget %d', dc.BudgetTicks) via OutcomeRejectedStale (internal/sim/cognition.go). The per-class BudgetTicks value (set_plan class = 1200) is the constant that doesn't scale with clock speed — full consumption-site map is the spec's first task.

Claiming per spec 065: spec dir specs/067-staleness-budget-scaling/ stubbed; full Spec Kit flow (specify → plan → tasks → implement) to follow.

Spec 067 authored (spec/plan/research/data-model/contracts/quickstart/tasks). Diagnosis pinned in spec.md with consumption-site map in research.md R2 (AC#1 ✓). Mechanism decided: landing-side speed scaling — BudgetTicks reinterpreted as 1x budget, effective = base × TicksPerSecond(state.Speed) at the landing tick; Route/horizon/governor untouched. Implementer tier: Opus 4.8 per constitution V rubric — cross-package (cognition+sim+mind), scheduling/cognition doctrine-adjacent behavior change.

spec-bridge sync: Setup: 1/1 · Foundational (blocking prerequisites): 2/2 · User Story 1 — Planning survives 8x on a calibrated local tier (P1): 3/3 · User Story 2 — Replay determinism across speed changes (P1): 1/1 · User Story 3 — Status surfaces stop calling structural death "thinking" (P2): 1/2 · Polish & Cross-Cutting: 1/2
<!-- SECTION:NOTES:END -->
