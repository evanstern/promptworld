---
id: TASK-113
title: Persist LLM latency estimator across daemon restarts
status: Done
assignee: []
created_date: '2026-07-25 03:00'
updated_date: '2026-07-25 05:21'
labels: []
dependencies: []
priority: medium
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
World-01 evidence: the live per-provider sec/pt estimate (cognition/estimate.go:50, EWMA-adaptive) is process-lifetime; world-01 restarted 36 times, so the estimator reset to the optimistic calibration floor 36 times, re-triggering the staleness storm each boot (92 recalibration_recommended events; planner decisions landing 17-23 ticks stale; calibrate output itself notes measured s/pt is a single-call floor while live load runs N agents concurrently). Fix: persist learned per-provider estimates into the world dir on shutdown (and periodically), reseed at boot from max(calibration seed, persisted estimate). Consider BreachRate 0.3->0.2 for faster median adoption while in there (cognition/estimate.go:13) — separate commit, measurable.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 learned s/pt survives a daemon restart (persisted in world dir, reseeded at boot)
- [x] #2 reseed takes the max of calibration seed and persisted value
- [x] #3 staleness-storm-after-restart no longer reproduces on world-01
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Delivered in PR #75 (squash-merged 2026-07-25). estimator_state.json in the world dir, calibration.json posture (boot-loaded, daemon-written every 5min + shutdown, never event-sourced, never read during replay — replay safety verified, not asserted). Boot reseed = max(calibration seed, persisted) via cognition.ReseedValue + Orchestrator.SeedPersisted. Restart-storm regression test drives the world-01 drift shape (10->60 s/pt) with zero re-adoptions after reseed. BreachRate 0.3->0.2 taken as separate characterized commit (adoption bar 7/20 -> 5/20 spikes, one-shot-rejection regression unchanged). Full -race suite green. Implemented by Sonnet spec-implementer per tier rubric; reviewed and merged by orchestrator. Wiki re-pin follows (cognition.md claims now stale).
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-07-25 04:54
---
Implementation started 2026-07-25. Tier: Sonnet spec-implementer (rubric: single-package routine slice — persistence + reseed in internal/cognition estimator, complete file:line diagnosis and ACs pinned on card; trivial-track exemption applies). Worktree .worktrees/task-113.
---
<!-- COMMENTS:END -->
