---
id: TASK-110
title: >-
  Tool surface hygiene: clamp expressive text instead of rejecting; prune dead
  verbs
status: In Progress
assignee: []
created_date: '2026-07-25 03:00'
updated_date: '2026-07-25 18:50'
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
- [ ] #1 expressive text overruns truncate and land, with a truncation notice in the verdict
- [ ] #2 set_plan step overflow clamps to first 3 instead of rejecting
- [ ] #3 structural malformations still reject strictly
- [ ] #4 collect_water and bathe removed from the villager loop roster
- [ ] #5 malformed rate measured before/after on world-01 (expect ~11% -> ~1%)
- [ ] #6 Spec phase: Foundational
- [ ] #7 Spec phase: User Story 1 — Expressive clamping (P1)
- [ ] #8 Spec phase: User Story 2 — set_plan step clamp (P1)
- [ ] #9 Spec phase: User Story 3 — Roster prune (P2)
- [ ] #10 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
MVLS sweep dispatch (2026-07-25): implementer tier Sonnet — constitution V rubric: routine/mechanical single-mechanism changes, complete file:line diagnosis on the task; no concurrency or doctrine arbitration. Lane 1; merges FIRST.
<!-- SECTION:NOTES:END -->
