---
id: TASK-101
title: 'Spike: foster better-composed agent goals (intent thrash: forage vs warmth)'
status: In Progress
assignee: []
created_date: '2026-07-24 21:45'
updated_date: '2026-07-24 21:47'
labels:
  - spike
dependencies: []
ordinal: 84000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Exploration/discussion spike. Problem: agents pursue single-intent goals (chop wood, forage) and oscillate between competing needs — e.g. forage while cold, abort to seek warmth, abort back to food — instead of composing a solution (build a fire where the food is, bring wood). Investigate: (1) does the thrash show up in logs; (2) is the cause missing context (needs/world/self understanding) at intent-selection time, missing goal persistence/commitment, or missing plan composition (multi-step goals); (3) sketch candidate mechanisms to encourage better problem solving.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Evidence from logs confirms or refutes the oscillation pattern, with excerpts recorded on the task
- [ ] #2 Root-cause hypothesis recorded (context vs commitment vs composition)
- [ ] #3 Candidate design directions written up with tradeoffs
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Map the intent/goal selection system (code + prompts)\n2. Hunt logs for forage↔warmth oscillation evidence\n3. Diagnose: context gap vs commitment gap vs composition gap\n4. Write up candidate design directions on the task
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Code map (explorer 1): Intent selection is an LLM tool-use loop (one acting tool per thought) over the tool registry (internal/tool/registry.go:280-316); deterministic reflex ladder fallback (internal/sim/policy.go:24-112, first-match-wins: eat > forage > warmth > sleep). Decision prompt (internal/mind/prompt.go:73-145) includes all 5 needs, inventory, structures, nearby agents, social/law context, top-K memories. Notably ABSENT: LastGoal (exists but TUI-only, never fed back), any echo of an active plan's intent. Prompt framing literally says 'choose the one action that best fits your situation and needs right now' (prompt.go:51-56) — myopic-reactive by construction. set_plan exists: max 3 steps (PlanStepCap, sim/plan.go:16), guards from closed vocab, 2-game-hour default expiry, broken plan cleared entirely (no resume). No goal persistence/commitment/hysteresis mechanism. Fire: build_fire = 2 wood, 600 ticks; warmth sources: fire radius 2, daytime, oven bath. Reflex ladder puts hunger unconditionally above warmth.
<!-- SECTION:NOTES:END -->
