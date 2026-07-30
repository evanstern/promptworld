---
id: TASK-172
title: >-
  Consolidation truncation: nightly memory consolidation silently collapses as
  worlds age
status: In Progress
assignee: []
created_date: '2026-07-30 16:41'
updated_date: '2026-07-30 19:13'
labels: []
dependencies: []
priority: high
ordinal: 140000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Villagers stop forming long-term memories as their world grows older: the nightly consolidation call outgrows its fixed output budget, comes back as truncated JSON, and is silently rejected — night after night, with nothing surfacing the failure.

As a player, I want my villagers to keep remembering and growing over a long run, so a 30-day world feels like week one but deeper.
As an operator, I want a loud signal when a background cognitive pipeline starts failing every night, instead of discovering a ten-day blackout in the daemon log after the fact.

Evidence (playtest-1, v5, 29 game-days, consolidation routed to cloud Sonnet): acceptance degrades monotonically — night 2: 7/9 accepted; night 11: 3/13; night 17: 1/15; nights 20–29: 0/8 every night (ten straight nights, all 8 villagers). Logged invalid sample is JSON cut mid-field; day-29 narration also died with "unterminated JSON object". Root cause still live on main: internal/mind/consolidate.go:166 defaults max_tokens.consolidation to 1024 with no truncation-aware retry. Only 232 agent.consolidated events landed all run.

Spec: specs/105-consolidation-truncation
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Consolidation output truncation is detected and retried with a larger budget (or input is bounded so output fits) — a 30-day-old villager's nightly consolidation succeeds
- [ ] #2 Per-night consolidation acceptance is observable (telemetry/log summary), and sustained failure is loud, not silent
- [ ] #3 A regression test covers consolidation at large memory volume (late-world shape), not just day-1 shape
- [ ] #4 Spec phase: Detection + ladder helper
- [ ] #5 Spec phase: Consolidation integration
- [ ] #6 Spec phase: Narrator generalization
- [ ] #7 Spec phase: Per-night observability
- [ ] #8 Spec phase: Docs + grounding
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
sweep dispatch (runbook playtest-1-findings-sweep): tier Opus 4.8 — cross-package (internal/mind consolidate worker + internal/llm token-budget seam), failure-handling in async mind orchestration per constitution P.V hard-slice rubric. Spec 105-consolidation-truncation. LANE 2: PR merges only after TASK-112's PR lands (operator ruling).

Implementation complete on branch task-172-consolidation-truncation @ ada23d09 (Opus 4.8, gated by orchestrator): T001-T008 done — truncation ladder (parse-first detection, 1024→2048→4096 clamp llm.MaxTokenBudget, ≤2 retries), late-world day-29 regression fixture (AC#3), narrator generalization (800→1600→3200), per-night acceptance summary + ≥2-night WARNING escalation; race suite green (2 e2e daemon-start flakes re-verified PASS in isolation — load starvation, not branch defect), parse.go zero-diff, 21 wiki notes re-pinned, 4 player pages regenerated, pr gate exit 0. OPEN: T009's post-112 merge-in clause. PR held per operator ruling. Deviations recorded in-branch: synchronous retried-record injection, exhausted-ladder chapter lands as gap per spec FR-008, docs/event-types.md path corrected to wiki event-type notes.
<!-- SECTION:NOTES:END -->
