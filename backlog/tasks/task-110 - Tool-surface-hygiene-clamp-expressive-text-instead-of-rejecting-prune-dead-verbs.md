---
id: TASK-110
title: >-
  Tool surface hygiene: clamp expressive text instead of rejecting; prune dead
  verbs
status: In Progress
assignee: []
created_date: '2026-07-25 03:00'
updated_date: '2026-07-25 20:15'
labels:
  - behavior-hygiene
  - mvls
dependencies: []
priority: high
ordinal: 18000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Diagnosis (2026-07-24, world-01 event log): 807 rejected_malformed tool calls, ~93% are 'exceeds text cap' across ALL tiers including cloud haiku — muse (403), talk_to (308), plus reason-cap overruns on world verbs. This is cap design, not model grammar: every model writes past the 200-rune caps. Fix: truncate-with-notice instead of reject for EXPRESSIVE text fields only (say.text/gist/muse.text/reason — sweep grounding 2026-07-25: talk_to carries no text param in the current registry; rejection site is internal/toolloop/loop.go, step-cap site internal/sim/landing.go); clamp set_plan to first PlanStepCap steps instead of rejecting; keep strict rejection for structural failures. Decision (user 2026-07-24): conversation STAYS on gemma — no re-route, no tool_mode experiment; loop damper (TASK-109) will cut volume, re-evaluate after. Also decided: shrink the roster — remove collect_water and bathe from LoopRosterVillager (water has no consumer); revisit if a thirst need is ever designed.

Spec: specs/058-tool-surface-hygiene
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 expressive text overruns truncate and land, with a truncation notice in the verdict
- [x] #2 set_plan step overflow clamps to first 3 instead of rejecting
- [x] #3 structural malformations still reject strictly
- [x] #4 collect_water and bathe removed from the villager loop roster
- [x] #5 malformed rate measured before/after on world-01 (expect ~11% -> ~1%)
- [x] #6 Spec phase: Foundational
- [x] #7 Spec phase: User Story 1 — Expressive clamping (P1)
- [x] #8 Spec phase: User Story 2 — set_plan step clamp (P1)
- [x] #9 Spec phase: User Story 3 — Roster prune (P2)
- [ ] #10 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
MVLS sweep dispatch (2026-07-25): implementer tier Sonnet — constitution V rubric: routine/mechanical single-mechanism changes, complete file:line diagnosis on the task; no concurrency or doctrine arbitration. Lane 1; merges FIRST.

spec-bridge sync: Foundational: 2/2 · User Story 1 — Expressive clamping (P1): 3/3 · User Story 2 — set_plan step clamp (P1): 2/2 · User Story 3 — Roster prune (P2): 2/2 · Polish & Cross-Cutting: 1/2

PR #92 squash-merged as 7e76246. Gated deviations accepted: say/gist clamp in the scene parser (already clamping; rune-safety bug fixed there), set_plan schema maxItems removed so the landing clamp is reachable, name-keyed reason clamp for set_plan. AC#5 before-rate measured (world-01 v3: 807 rejections, ~93% text-cap); after-rate pends world-01 runtime on the post-058 binary — revisit before sweep close.

AC#5 after-measurement (2026-07-25, world-01 live on the lane-1 binary, 3,612 events post-restart): 200 most recent cog.tool_call verdicts = landed 85, landed_clamped 18, read_ok 58, rejected_gate 39, rejected_malformed 0. Zero malformed rejections vs ~11% pre-058; the 18 landed_clamped are exactly the text-cap class (93% of old rejections) now landing. Modest sample, decisive direction.
<!-- SECTION:NOTES:END -->
