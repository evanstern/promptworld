---
id: TASK-101
title: 'Spike: foster better-composed agent goals (intent thrash: forage vs warmth)'
status: In Progress
assignee: []
created_date: '2026-07-24 21:45'
updated_date: '2026-07-24 21:51'
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
- [x] #1 Evidence from logs confirms or refutes the oscillation pattern, with excerpts recorded on the task
- [x] #2 Root-cause hypothesis recorded (context vs commitment vs composition)
- [x] #3 Candidate design directions written up with tradeoffs
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Map the intent/goal selection system (code + prompts)\n2. Hunt logs for forage↔warmth oscillation evidence\n3. Diagnose: context gap vs commitment gap vs composition gap\n4. Write up candidate design directions on the task
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Code map (explorer 1): Intent selection is an LLM tool-use loop (one acting tool per thought) over the tool registry (internal/tool/registry.go:280-316); deterministic reflex ladder fallback (internal/sim/policy.go:24-112, first-match-wins: eat > forage > warmth > sleep). Decision prompt (internal/mind/prompt.go:73-145) includes all 5 needs, inventory, structures, nearby agents, social/law context, top-K memories. Notably ABSENT: LastGoal (exists but TUI-only, never fed back), any echo of an active plan's intent. Prompt framing literally says 'choose the one action that best fits your situation and needs right now' (prompt.go:51-56) — myopic-reactive by construction. set_plan exists: max 3 steps (PlanStepCap, sim/plan.go:16), guards from closed vocab, 2-game-hour default expiry, broken plan cleared entirely (no resume). No goal persistence/commitment/hysteresis mechanism. Fire: build_fire = 2 wood, 600 ticks; warmth sources: fire radius 2, daytime, oven bath. Reflex ladder puts hunger unconditionally above warmth.

LOG EVIDENCE (explorer 2, world-01 @ ~/.promptworld/worlds/world-01/world.db, 261,722 events / 523,670 ticks): oscillation CONFIRMED. forage↔goto_warmth flips per agent: Sage 436 (334 within ≤200 ticks), Fern 350, Oak 294, Rowan 204, Birch 202. Excerpt — Sage ticks 265,411–266,631: intent alternates every ~200–320 ticks between forage→(3,20) [source=reflex] and goto_warmth→(4,11) [source=planner], planner reasons 'Warmth is dangerously low at 45… must reach fire immediately' while the reflex re-issues forage moments after warmth ticks up (438→516 at fire, then drains again; food drifts 816→795). Sources split cleanly: forage is reflex-issued (4098 reflex vs 993 planner), goto_warmth is planner-issued (1089 planner vs 42 reflex).

ROOT CAUSE (AC#2): not primarily a comprehension gap — planner reasoning shows the agent KNOWS it is cold and where the fire is. Three structural causes:
(1) LAYER FIGHT: reflex and planner counter-schedule each other. Sage's food was 816 (not hungry; night forage impossible) ⇒ the reflex forage came from the daytime larder-stocking prep rule (policy.go:96-100, stock to 8 raw), which never checks warmth — the day branch of the ladder has no warmth rule at all. Reflex fires whenever the agent idles past 120-tick grace, i.e. the moment goto_warmth COMPLETES on arrival.
(2) RECOVERY IS NOT REPRESENTABLE: goto_warmth is arrive-and-done; warming up requires LOITERING, but idleness at the fire is exactly what invites the reflex to dispatch the agent elsewhere. No intent form 'stay until warmth ≥ N' — needs-recovery goals complete on location, not on the need.
(3) SELF-BLINDNESS + WEAK COMPOSITION: the prompt never includes the agent's own recent intents (LastGoal exists, TUI-only), need trajectories, or active-plan echo; framing is 'best fits right now'. build_fire resolves to nearest build site from current position (policy.go:277-285), so 'fire at the berry patch' is only expressible as a composed plan (forage → build_fire), which nothing encourages; plans expire in 2 game-hours and clear entirely on any break.

DESIGN DIRECTIONS (AC#3), ordered by leverage/cost:
A. STOP THE LAYER FIGHT (deterministic, cheap, no LLM cost): (i) reflex prep rules (larder-stock, refuel) must yield when any need is in a danger band or when a planner intent landed recently — reflex as safety net, not scheduler; (ii) add a warmth check to the reflex's day branch; (iii) make reflex respect/extend the planner's last decision instead of overwriting it at first idle.
B. NEEDS-CONDITIONED RECOVERY INTENTS (deterministic, cheap): recovery goals complete on the NEED, not the location — 'warm_up until warmth ≥ 800' loiters at the fire; mirrors eat-to-satiety. Kills the arrive→idle→reflex-vacuum cycle directly.
C. SELF-HISTORY IN THE PROMPT (prompt-side, medium): feed back (a) current/last intent + source ('you were foraging; a reflex sent you'), (b) need trajectories (warmth 52↓ falling, food 81→ steady), (c) active plan echo so the next thought CONTINUES rather than restarts. Addresses 'lack of understanding of their place' — the model can't reason about a loop it cannot see.
D. THRASH DETECTION AS PERCEPT (medium): detect k intent-alternations within a window in-sim and inject as an observation/memory: 'you have walked between the fire and the berry patch 5 times; neither need has improved.' Lets the LLM's actual competence engage the pattern; pairs with a prompt nudge 'when two needs compete, prefer a plan that co-locates them (e.g. build a fire near food).'
E. STRONGER PLAN MECHANICS (bigger): plan resume instead of clear-on-break, longer/renewable expiry, and possibly location anchors on build verbs — though forage→build_fire already composes 'fire at the food site' since build_fire builds near where you stand.
Tradeoffs: A+B are sim-only and testable deterministically but don't make the LLM smarter; C+D spend prompt tokens per thought and need tuning to avoid nagging; E touches plan.go semantics that spec 014/017 deliberately kept conservative. Recommended order: A/B first (removes the structural saboteur), then C, then D; E only if composition still doesn't emerge.
<!-- SECTION:NOTES:END -->
