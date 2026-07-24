---
id: TASK-78
title: >-
  Teaching-world speed posture: calibrated soft cap with horizon-arithmetic
  warning (classroom mode)
status: In Progress
assignee: []
created_date: '2026-07-23 17:00'
updated_date: '2026-07-24 19:05'
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
ordinal: 7000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Follow-up cut from TASK-66 / decision-6 (client decision 2026-07-23). Teaching-mode worlds default their speed to the highest calibrated planner-safe rung of the ladder — the number calibrate already computes (horizonSummary, cmd/promptworld/calibrate.go:173-197: 'planner suppressed above 16x'). SOFT posture, decided explicitly over a hard cap: exceeding the default is allowed and surfaces the horizon arithmetic (e.g. '3pt × 17.0s/pt × 32x = 1632 ticks > budget 1200 — villagers will stop deep-thinking'), so overriding the cap is itself a lesson about the horizon.

Shape (decision-6): a per-world teaching/config posture consumed by TASK-68's stage presets — NOT an engine rule (decision-4 stands: the engine never caps speed to protect cognition; this caps a teaching posture to protect feedback legibility). Derived per world from the calibration profile at world creation/speed-change time, never hard-coded — must survive spec 024's per-provider seconds-per-point divergence (recompute from the profile the planner class actually routes to).

Interactions: TASK-40 (uncalibrated worlds silently over-suppress — an uncalibrated teaching world must prompt calibrate before the posture can be honest; bootstrap 20 s/pt would cap at 16x pessimistically); TASK-68 (stage presets carry the posture field); TASK-41 (horizon legibility in the TUI is the always-on counterpart of the one-shot warning). Stage 1 worlds (conversational Metatron) need no posture — the metatron class never suppresses at watchable speeds.

Spec: specs/039-teaching-speed-posture
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Teaching worlds default to the highest calibrated planner-safe ladder speed, derived from the world's calibration profile (not hard-coded)
- [ ] #2 Setting a speed above the posture succeeds and surfaces the horizon arithmetic for the classes it suppresses
- [ ] #3 An uncalibrated teaching world prompts for calibrate rather than silently adopting the pessimistic bootstrap cap (aligns with TASK-40)
- [ ] #4 Posture lives as per-world config consumable by TASK-68 stage presets; non-teaching worlds are unchanged
- [ ] #5 Spec phase: Setup
- [ ] #6 Spec phase: Foundational (Blocking Prerequisites)
- [ ] #7 Spec phase: User Story 1 - Teaching world runs at the fastest honest speed by default (Priority: P1) 🎯 MVP
- [ ] #8 Spec phase: User Story 2 - Exceeding the posture teaches the horizon instead of blocking (Priority: P2)
- [ ] #9 Spec phase: User Story 3 - Uncalibrated teaching worlds are told to calibrate (Priority: P2)
- [ ] #10 Spec phase: User Story 4 - The posture is a per-world fact other features can read (Priority: P3)
- [ ] #11 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Spec Kit flow complete: specs/039-teaching-speed-posture (spec + plan + research + data-model + contracts/posture.md + quickstart + tasks.md, 18 tasks / 7 phases). Clarify skipped: decision-6 + docs/design/horizon-vs-learner-iteration-speed.md pre-answer the shape questions (soft cap, planner-safe rung, uncalibrated prompt, per-world config not engine rule). Design: optional Manifest.Teaching bool (omitempty, no FormatVersion bump); exported cognition.MaxSafeSpeed extracted from HorizonSummary's maxOK loop; boot default applied as a recorded clock.speed_set via the normal loop command path (replay byte-identity); posture warning composes with spec 035's uncalibratedWarning on the one StatusData.Warning set_speed field; additive StatusData.Posture for TASK-68. Implementation: one worktree .worktrees/task-78, branch task-78-teaching-speed-posture, one PR; T001-T017 delegated to spec-implementer; T018 wiki re-pin + player-docs at root after merge, then sync to Done.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Drift audit 2026-07-23: content holds; pin moved — horizonSummary is calibrate.go:241 with the 'suppressed above Nx' arithmetic at :261 (:173-197 is orchSampler). No teaching posture exists anywhere yet, confirmed.

Model tier: Opus 4.8 (constitution V rubric): slice is cross-package (world/cognition/ipc/daemon/cmd), touches internal/cognition (rubric-listed), and the boot-time recorded speed_set event is replay-determinism/doctrine-adjacent (decision-4/-6 boundary must stay warn-never-block). Recorded per plan.md Constitution Check + research.md R7.
<!-- SECTION:NOTES:END -->
