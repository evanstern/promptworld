---
id: TASK-107
title: 'World tuning manifest: tuning.json promotion path for doctrine constants'
status: In Progress
assignee: []
created_date: '2026-07-25 02:59'
updated_date: '2026-07-25 17:46'
labels: []
dependencies: []
priority: high
ordinal: 12000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Keystone from docs/design/control-surface-and-calibration.md §6. A boot-loaded, clamp-validated, event-logged tuning.json in the world dir; every field defaults to the current doctrine constant. First promoted dials: refuelDyingBelow, fireBurnPerWood, gruEmergePerMille, PlannerCadenceTicks, conversation pair cooldown. Values logged as events at boot/change so replays reproduce behavior (calibration.json pattern). Follow-on goal (user decision 2026-07-24): once dialed in on world-01, fold the tuned values back as the standard default for new worlds.

Spec: specs/048-tuning-manifest
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 tuning.json read at boot with per-field clamps; absent file == current constants
- [ ] #2 applied values emitted as events so replay is deterministic
- [ ] #3 the five named dials consume the manifest instead of consts
- [ ] #4 docs/design report §6 updated to point at the mechanism
- [x] #5 Spec phase: Foundational (Blocking Prerequisites)
- [x] #6 Spec phase: User Story 1 — Operator tunes a dial without editing code (P1) 🎯 MVP
- [x] #7 Spec phase: User Story 2 — Replays reproduce tuned behavior (P1)
- [x] #8 Spec phase: User Story 3 — The five earned dials consume the manifest (P2)
- [x] #9 Spec phase: User Story 4 — Design report points at the mechanism (P3)
- [ ] #10 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Full Spec Kit run: specs/048-tuning-manifest (spec/plan/research/data-model/contracts/quickstart/tasks committed to main 02b3009..92bb134). Implementation: 20 tasks in 7 phases per tasks.md — TuningState on sim.State (omitempty, no format_version bump), boot-seeded sim.tuning_applied full-set event (seedMeetingConvention pattern), clamp-values/reject-structure parsing (llm config pattern), accessors consumed by reducer + mind replica. One branch task-107-tuning-manifest in .worktrees/task-107, one PR.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementer tier: Opus 4.8 (constitution V rubric — three independent triggers: cross-package slice (sim reducer + mind scheduling + daemon boot + world), doctrine-adjacent behavior change (rebinds five doctrine constants), and planner-cadence/scheduling logic in internal/mind). Delegated via spec-implementer agent, model=opus.

spec-bridge sync: Foundational (Blocking Prerequisites): 3/3 · User Story 1 — Operator tunes a dial without editing code (P1) 🎯 MVP: 4/4 · User Story 2 — Replays reproduce tuned behavior (P1): 5/5 · User Story 3 — The five earned dials consume the manifest (P2): 4/4 · User Story 4 — Design report points at the mechanism (P3): 1/1 · Polish & Cross-Cutting: 2/3
<!-- SECTION:NOTES:END -->
