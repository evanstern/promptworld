---
id: TASK-103
title: Instinct yields to intelligence — reflex/planner arbitration
status: In Progress
assignee: []
created_date: '2026-07-25 02:41'
updated_date: '2026-07-25 20:06'
labels:
  - goal-quality
  - instinct-layer
  - mvls
dependencies:
  - TASK-107
  - TASK-108
priority: high
ordinal: 15000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Direction A from spike TASK-101. Reframe the reflex ladder as 'instinct' that YIELDS to intelligence: reflex prep rules (larder-stocking, refuel top-up, first-fire prep) must not counter-schedule against a recent planner intent or fire while any need is in a danger band; add a warmth rung to the reflex's day branch (the gap that manufactured the Sage loop); night frontier-search fallback (057 audit Gap A); prune/soften hard-coded prep behavior where the planner owns it (the yield gate IS the softening). Evidence: world-01 forage<->goto_warmth thrash (Sage 436 flips, 334 within <=200 ticks) is reflex-vs-planner counter-scheduling, not LLM indecision; TASK-106 research shows the storms were village-wide (days 4-5, 6 of 8 villagers). Full Spec Kit: spec 062.

Spec: specs/062-instinct-yields
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Reflex never overrides a live/recent planner decision except on survival-threshold breach
- [x] #2 Day-branch warmth gap closed
- [x] #3 Replay/sim test demonstrates Sage-style thrash episode no longer occurs
- [x] #4 Spec phase: Foundational
- [x] #5 Spec phase: User Story 1 — Prep yields (P1)
- [x] #6 Spec phase: User Story 2 — Day warmth rung (P1)
- [x] #7 Spec phase: User Story 3 — Night search fallback (P3, droppable by amendment)
- [x] #8 Spec phase: User Story 4 — Thrash regression (P1)
- [ ] #9 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
RECONCILIATION with TASK-108 (2026-07-24): no conflict once the ladder is split into two kinds of instinct. SURVIVAL instinct (don't die: eat-when-starving, warmth-at-night, 108's new cold-reflex build-fire, refuel) gets MORE authority per the table-stakes doctrine; HOUSEKEEPING instinct (larder stocking, prep chopping, wander) is what yields to intelligence — it was a housekeeping rule (day-branch larder stock, policy.go:96) that caused the world-01 thrash by preempting survival recovery. Sequence: 107 (tuning manifest) → 108 (survival gaps, surgical) → this task (yield semantics for housekeeping + day-branch warmth check). 108's thresholds and this task's danger bands both ride the tuning manifest. Design together with TASK-104: recovery-loitering must not read as 'idle' to any instinct rule.

MVLS sweep dispatch (2026-07-25, lane 2 — forked after TASK-108's merge): implementer tier Opus 4.8 — constitution V rubric: doctrine-adjacent change to the sim reducer's decision ladder; the slice the MVLS program hinges on.

OPERATOR DECISION (2026-07-25, team review) — SEQUENCING CONFLICT TO RESOLVE. The operator accepted 108 -> 104 -> 103. This task was dispatched In Progress against spec 062 before that decision was recorded. Reconcile with the MVLS session before implementing the warmth half.

WHY 104 FIRST: internal/sim/policy.go contains ZERO reads of .Warmth (verified by grep). Warmth is not a need to the instinct layer — every warmth rung keys on warmAt(tile), a location predicate, plus s.Night. AC#2 ('day-branch warmth gap closed') is therefore unwritable until TASK-104 makes warmth a need the ladder can read.

AC#1 AS WRITTEN DESCRIBES A RACE THAT CANNOT HAPPEN. The reflex runs only inside 'if a.Intent == nil' (executor.go:240), gated on nextTick - a.IdleSince >= reflexGraceTicks (:256). It structurally CANNOT override a live planner intent. The real mechanism is a VACUUM, not a race: intent completes -> intent_done stamps IdleSince (~25 reducer sites) -> 120 ticks later the reflex acts -> the planner's next beat is PlannerCadence() = 1800 ticks away. The reflex owns ~93% of the idle window by construction. Written against 'override', this task ships a guard that never fires; written against 'vacuum', the levers are the grace window and the cadence. Recommend rewording AC#1 before implementation.

THE SURVIVAL/HOUSEKEEPING SPLIT DOES NOT EXIST IN THE CODE. decideIntent is one flat ladder (policy.go:24-126) returning an untagged 'decision'; IntentSetPayload.Source (agents.go:102) is provenance, not authority. There is no seam to toggle — this task must BUILD one. The minimal honest shape is a class tag on the decision threaded into IntentSetPayload, which is an EVENT-PAYLOAD CHANGE and therefore REPLAY-AFFECTING. The operator wanted it to ride TASK-108; 108's PR #89 has already merged, so it now rides this task — treat it as replay-affecting scope and pair it with TASK-134 (event-log format_version).

HAZARD NO TASK NAMES: Agent.LastGoalTick (agents.go:201) is classified 'keep' in the rebase taxonomy and read only by the TUI (tui/views.go:2518). The moment this task uses it as a recency test (nextTick - LastGoalTick < N) it becomes a duration anchor and MUST flip to 'shift', or a Metatron timeline rebase silently corrupts arbitration. Same for IntentRecord.Tick/OutcomeTick. TestRebaseTaxonomyComplete will NOT catch this — it verifies a field IS classified, never that the classification is CORRECT.

Any new threshold ships as a tuning.json dial (spec 048 / 057 US2), not a bare const.

spec-bridge sync: Foundational: 3/3 · User Story 1 — Prep yields (P1): 2/2 · User Story 2 — Day warmth rung (P1): 2/2 · User Story 3 — Night search fallback (P3, droppable by amendment): 2/2 · User Story 4 — Thrash regression (P1): 1/1 · Polish & Cross-Cutting: 1/2

PR #93 squash-merged as 46b1841. Human ACs proven: #1 prep yields to window+danger bands (survival-threshold breach = survival rungs, unconditioned); #2 day warmth rung shipped (gated deviation: no day chop tail — degraded-mode 8/8 sacred, pinned by test); #3 Sage-shape regression proves the loop dead both directions. Runbook 104-before-103 amendment reconciled with evidence (premise disproven — Needs.Warmth was always readable; the gap was policy not reading it, which IS this fix).
<!-- SECTION:NOTES:END -->
