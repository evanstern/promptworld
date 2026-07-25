---
id: TASK-105
title: Per-turn context grounding — audit and intent-driven context assembly
status: To Do
assignee: []
created_date: '2026-07-25 02:41'
updated_date: '2026-07-25 03:28'
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
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Written per-turn context inventory (present vs absent) exists as a durable artifact
- [ ] #2 Self-history, need trajectories, and active-plan echo added to the decision prompt
- [ ] #3 Relevant-memory/journal retrieval feeds the prompt with measured token budget
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Synergies (2026-07-24 board pass): TASK-110 prunes dead verbs from the roster — shrinks the tool surface this task's context budget pays for; do 110's roster prune before or with the context redesign. Relevant-memory retrieval leans on the embedding work: TASK-98 (in progress, spec 042) provides record-at-emission vectors + relevance term; TASK-102 (embed preflight warning bug) should land so embedding-path signal is clean.

Spec drafted and committed to main (996d503): specs/043-context-grounding/spec.md — 5 user stories (P1 self-history + P1 context inventory audit, P2 need trajectories, P3 plan echo, P4 relevance retrieval under token budget), 10 FRs, 7 measurable SCs incl. flip-rate reduction vs the world-01 baseline. Requirements checklist passes; zero NEEDS CLARIFICATION (defaults in Assumptions). Next: speckit-plan → speckit-tasks → spec-bridge:link, then delegated implementation.
<!-- SECTION:NOTES:END -->
