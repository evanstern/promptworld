---
id: TASK-77
title: >-
  Paused authoring chain-completion: nudge wakes the nudged villager +
  pause-aware routing (classroom mode)
status: In Progress
assignee: []
created_date: '2026-07-23 17:00'
updated_date: '2026-07-24 19:09'
labels:
  - teaching-game
  - classroom-mode
dependencies: []
references:
  - >-
    backlog/decisions/decision-6 -
    Classroom-mode-curriculum-staged-horizon-posture-—-paused-chain-completion-for-authoring-calibrated-soft-speed-cap-for-ambient-running-budgets-stay-doctrine.md
  - docs/design/horizon-vs-learner-iteration-speed.md
priority: medium
ordinal: 8000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Follow-up cut from TASK-66 / decision-6 (client decision 2026-07-23). While paused, the operator's mediated chain already works up to the nudge becoming a memory: metatron_chat has no pause gate (ipc/server.go:312), the angel's landed effects inject at the frozen tick (blessed by decision-4), and the nudge lands as a SalDream memory. It breaks at the last two links, and this task is exactly those two fixes — no new mode, no new verbs, no single-stepping (explicitly deferred by the client):

(1) Wake: a landed nudge arms the nudged villager's planner for one round at the frozen tick — add the nudge event to absorb()'s arm switch (internal/mind/mind.go:203-228). Bounded by construction: the 300-tick planner debounce is game-time and cannot reopen while frozen, so one nudge buys exactly one round (same shape as decision-4's blessed catch-up round).

(2) Truth: pause-aware routing — routeVerdict (internal/mind/telemetry.go:61-71) computes drift at the world's SET speed even while frozen, suppressing a thought whose real drift is zero. Paused ⇒ predicted drift 0 ≤ any budget ⇒ allow; the recorded arithmetic string should say the world was paused. Not an override of the horizon — it makes the arithmetic tell the truth.

Doctrine door (decision-6): extends decision-4's landing-triggered catch-up blessing to landings the operator caused via Metatron — pause changes meaning from 'the minds are quiet' to 'the world is frozen, but responds to the angel.' Villagers stay sealed; influence stays mediated. Replay determinism holds: paused verdicts are reproducible arithmetic; frozen-tick thoughts enter the log as recorded events.

DOCTRINE-ADJACENT BEHAVIOR CHANGE in internal/mind — Opus 4.8 rubric tier per constitution V; full Spec Kit (specify → plan → tasks → spec-bridge:link) before implementation. The learner loop it buys: pause → edit charter → 'Metatron, nudge Aldric' → watch Aldric's one thought land under the new charter → resume. Diagram from the design session: docs/design/horizon-vs-learner-iteration-speed.md.

Spec: specs/040-paused-chain-completion
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A landed nudge arms the nudged villager's planner for exactly one bounded round at the frozen tick; the debounce still prevents any second round while frozen
- [ ] #2 routeVerdict treats a paused world as zero predicted drift (allow); the verdict's recorded arithmetic names the paused state
- [ ] #3 Frozen-tick thoughts land at zero staleness, fully recorded (cog.thought/cog.outcome); replay determinism harness green
- [ ] #4 Unpaused behavior is byte-identical to today: no new wake stimuli or routing changes apply while running
- [x] #5 Spec Kit spec written and linked via spec-bridge before implementation (doctrine-adjacent, non-trivial)
- [ ] #6 Spec phase: Setup
- [ ] #7 Spec phase: User Story 1 — A nudge wakes the nudged villager while paused (Priority: P1) 🎯 MVP
- [ ] #8 Spec phase: User Story 2 — Paused routing tells the truth (Priority: P1)
- [ ] #9 Spec phase: User Story 3 — The running world is untouched (Priority: P2)
- [ ] #10 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Spec Kit 040 (specs/040-paused-chain-completion; renumbered from 039 after a concurrent session claimed 039 for teaching-speed-posture/TASK-78): spec → plan → tasks generated 2026-07-24; linked via spec-bridge (gate green). Design (research.md D1-D5): (1) absorb() gains case metatron.nudged, paused-gated, arming each Targets index with the nudge's Seq as causality edge; (2) new pure cognition.RoutePaused (allow, drift 0, arithmetic '…while paused = 0 ticks <= budget N') consulted by routeVerdict before the tps<=0 branch so paused wins at uncapped; (3) newMeta paused ⇒ predictedLandTick=snapshotTick (kills the future-dating prefix via existing futureDated no-op; prompt/gate/record agree). No new events/verbs/modes; debounce is the bound. 16 tasks (tasks.md), tests-first; MVP=US1 wake at 16x.

TIER: Opus 4.8 (senior) via spec-implementer — rubric: doctrine-adjacent behavior change in internal/mind routing/wake semantics (constitution V explicit trigger); touches cognition-horizon doctrine (decision-4/6 pause semantics). One-way escalation not applicable (starts at Opus).
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Drift audit 2026-07-23: pins re-verified; one moved — metatron_chat handler (no pause gate) is ipc/server.go:334-355, not :312 (:312 is llm_call). absorb() arm switch mind.go:206-228 and routeVerdict telemetry.go:61-71 hold exactly.

2026-07-24: Spec 039 authored + linked (AC#5 satisfiable once committed); plan/research/data-model/contracts/quickstart/tasks on disk. Implementation delegated to spec-implementer @ Opus 4.8 in .worktrees/task-77.

2026-07-24: renumbered spec dir 039→040 (concurrent TASK-78 session claimed 039 on main first); marker, feature.json, and spec docs updated.
<!-- SECTION:NOTES:END -->
