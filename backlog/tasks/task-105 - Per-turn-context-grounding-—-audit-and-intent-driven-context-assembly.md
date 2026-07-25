---
id: TASK-105
title: Per-turn context grounding — audit and intent-driven context assembly
status: In Progress
assignee: []
created_date: '2026-07-25 02:41'
updated_date: '2026-07-25 04:18'
labels:
  - goal-quality
dependencies: []
references:
  - specs/043-context-grounding
priority: high
ordinal: 14000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Direction C from spike TASK-101 — Evan: '100% on this, almost more important than A/B.' Two parts. (1) AUDIT: produce a durable, complete inventory of exactly what each villager receives in context per thought (system prompt + userPrompt blocks, internal/mind/prompt.go:73-145) and what is notably absent. Known gaps: own last/current intent + source (LastGoal is TUI-only), need TRAJECTORIES (level+direction, not just level), active-plan echo so a thought continues rather than restarts. (2) REDESIGN: assemble context efficiently and with intent — self-history block, trajectories, plan echo, plus richer grounding via relevant-memory retrieval and selective journal-entry stuffing (dovetails with the embedding-memory retrieval work, spec 042 / TASK-98). Budget note: thoughts run 4-5 loop turns max, so per-turn context stuffing is affordable on a moderately hostable model. Non-trivial: full Spec Kit before implementation.

Spec: specs/043-context-grounding
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Written per-turn context inventory (present vs absent) exists as a durable artifact
- [ ] #2 Self-history, need trajectories, and active-plan echo added to the decision prompt
- [ ] #3 Relevant-memory/journal retrieval feeds the prompt with measured token budget
- [x] #4 Spec phase: Setup
- [x] #5 Spec phase: Foundational (blocking prerequisites)
- [ ] #6 Spec phase: US5 — Operators can see exactly what an agent knew (P1) 🎯 co-MVP
- [ ] #7 Spec phase: US1 — An agent knows what it was just doing (P1) 🎯 co-MVP
- [ ] #8 Spec phase: US2 — An agent feels which way its needs are moving (P2)
- [ ] #9 Spec phase: US3 — An agent continues its plan instead of restarting it (P3)
- [ ] #10 Spec phase: US4 — What an agent remembers is chosen for the moment (P4)
- [ ] #11 Spec phase: Polish & cross-cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Synergies (2026-07-24 board pass): TASK-110 prunes dead verbs from the roster — shrinks the tool surface this task's context budget pays for; do 110's roster prune before or with the context redesign. Relevant-memory retrieval leans on the embedding work: TASK-98 (in progress, spec 042) provides record-at-emission vectors + relevance term; TASK-102 (embed preflight warning bug) should land so embedding-path signal is clean.

Spec drafted and committed to main (996d503): specs/043-context-grounding/spec.md — 5 user stories (P1 self-history + P1 context inventory audit, P2 need trajectories, P3 plan echo, P4 relevance retrieval under token budget), 10 FRs, 7 measurable SCs incl. flip-rate reduction vs the world-01 baseline. Requirements checklist passes; zero NEEDS CLARIFICATION (defaults in Assumptions). Next: speckit-plan → speckit-tasks → spec-bridge:link, then delegated implementation.

Implementation dispatch (constitution Principle V): Foundational + US1 slice (T001-T012, T013 stretch) → Opus 4.8 spec-implementer. Rubric: touches internal/sim reducer state and internal/mind orchestration (doctrine-adjacent, cross-package), shadow-invariant byte-identity constraint — senior tier required. US5 (T006-T007, wiki note + capture) → Sonnet, dispatched after the code slice lands on the branch.

Foundational + US1 slice landed on task-105-context-grounding (commits 9de0665, d2bbd7f, 1e1b97d; Opus 4.8 implementer). Assembler + byte-identity wrap, cog.thought sizes, IntentRecord ring + 5 reducer arms, self_history block. Whole-repo tests, vet, gofmt green; shadow invariant holds. Gate-checked by orchestrator (targeted sim/mind tests re-run). T013 stretch: SC-004 confirmed via read-only probe of world-01 (Sage tick 265,864 ring shows instinct forage + alternation); committed harness deferred to T024/T027 pattern (needs replay-to-tick helper + env-guarded skip). Deviations accepted: intent_rejected now state-mutating per data-model (split from cog.* no-op arm); IntentRecord ticks KEEP in rebase taxonomy; self_history always renders (empty state line); future-dating line owned by frame block.
<!-- SECTION:NOTES:END -->
