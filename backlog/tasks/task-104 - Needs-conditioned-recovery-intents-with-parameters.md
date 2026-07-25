---
id: TASK-104
title: Needs-conditioned recovery intents with parameters
status: In Progress
assignee: []
created_date: '2026-07-25 02:41'
updated_date: '2026-07-25 21:49'
labels:
  - goal-quality
  - instinct-layer
  - mvls
dependencies:
  - TASK-103
priority: high
ordinal: 16000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Direction B from spike TASK-101. Recovery goals complete on the NEED, not the location: warm_up(until_warmth>=N) loiters at the fire until the condition holds, mirroring eat-to-satiety. Generalize: parameterized intent arguments (tool args carry the completion condition) rather than new one-off verbs — flexible and generalizable per Evan's note. Kills the arrive->idle->reflex-vacuum cycle that manufactures the oscillation (062 stopped the counter-scheduling; this makes recovery itself hold). Also carries 057 audit Gap C: sleepers wake to cold (Oak's final night). Full Spec Kit: spec 064.

Spec: specs/064-needs-conditioned-recovery
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 At least warm_up (and pattern for rest/food analogs) completes on a need condition passed as an argument
- [x] #2 Idle-at-recovery-site no longer triggers instinct dispatch mid-recovery
- [x] #3 Deterministic sim test covers recover-then-release behavior
- [x] #4 Spec phase: Foundational
- [x] #5 Spec phase: User Story 1 — warm_up (P1)
- [x] #6 Spec phase: User Story 3 — interruptibility (P1)
- [x] #7 Spec phase: User Story 2 — generic mechanism proof (P2)
- [x] #8 Spec phase: User Story 4 — wake to cold (P2)
- [x] #9 Spec phase: Integration proof
- [ ] #10 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Design together with TASK-103: a villager loitering-to-recover must not register as idle to the instinct layer (the 120-tick idle grace is what let the larder rule hijack Sage at the fire).

REVIEW FINDING (2026-07-25) — THIS TASK OWNS A LIVE, UNINTERRUPTIBLE DEATH PATH. Raised to High.

TASK-108's survival audit routed 'wake-to-cold' here. The review independently confirmed the mechanism and it is worse than a gap: wakeReason (internal/sim/executor.go) returns true ONLY for day-plus-rested, or a hunger emergency with food in hand. COLD IS NOT A WAKE CONDITION. decayNeeds drains warmth 4/min at night when not warm; at Warmth==0 health drains 3/min. policy.go:74 will put a cold agent with no wood and no known warmth to sleep anyway. A sleeping villager therefore freezes to death with no wake path — almost certainly world-01's Oak.

Root cause shared with TASK-103: internal/sim/policy.go contains ZERO reads of .Warmth (verified by grep; Needs.Food/Needs.Rest have four). Warmth is not a need to the instinct layer at all — every warmth rung keys on warmAt(tile), a LOCATION predicate, plus s.Night. That is why this task is foundational rather than the third step: warm_up(until warmth>=N) is the first time warmth becomes a need the ladder can see.

OPERATOR DECISION (2026-07-25): sequence 104 BEFORE 103 — 103's AC#2 ('day-branch warmth gap closed') is unwritable until this task exists, because there is nothing to check in the day branch except a need the ladder cannot read. CONFLICT TO RESOLVE: TASK-103 was dispatched In Progress against spec 062 before this decision was recorded. Reconcile with the MVLS session.

Also: any new threshold this task introduces must ship as a tuning.json dial (spec 048 / 057 US2 genesis pin), not a bare const.

MVLS sweep dispatch (2026-07-25, lane 3 — forked after TASK-103's merge #93): implementer tier Opus 4.8 — constitution V rubric: cross-package (sim executor + tool registry + mind handler), doctrine-adjacent intent-completion semantics.

Implementer report gated (2026-07-25, read-only — merge held for operator signal): T001-T014 done, 22 packages green, vet clean, TUI gate no-op. Four deviations ACCEPTED: (1) rest analog mechanism-proven in tests only (plan R6 escape hatch; sleep untouched); (2) exposureWakeBelow=150 NEW constant, not the R5-nominated 350 — faithful mirror of the hunger EMERGENCY floor (350 roused sleepers on routine dips, regressed degraded-mode); (3) survival-preemption yield on holds (US3 AS2/FR-004, keeps 8/8, less sticky); (4) governance emergent-gathering exclusion for hold-pinned villagers (Asleep/Exiled parallel, byte-inert pre-064) — flagged for operator review. Constants: warmthRecoverTo=800, recoveryStallTicks=300. Survival-neutral across 30 seeds. Branch NOT yet rebased across the guardian rename (PR #94 post-dates its fork) — rebase+gates+PR pend operator go.

spec-bridge sync: Foundational: 2/2 · User Story 1 — warm_up (P1): 4/4 · User Story 3 — interruptibility (P1): 2/2 · User Story 2 — generic mechanism proof (P2): 2/2 · User Story 4 — wake to cold (P2): 2/2 · Integration proof: 1/1 · Polish & Cross-Cutting: 1/2

PR #96 squash-merged as 5acb5b5 (clean rebase across the guardian rename, full gates re-run green). Human ACs proven: #1 warm_up completes on the need condition passed as an argument, rest analog proves the pattern; #2 hold-at-target kills idle-at-recovery-site dispatch (TestReflexHoldNoArriveIdleWander + TestSageWarmUpHeldToThresholdThenReleased); #3 TestRecoverThenRelease deterministic recover-then-release.
<!-- SECTION:NOTES:END -->
