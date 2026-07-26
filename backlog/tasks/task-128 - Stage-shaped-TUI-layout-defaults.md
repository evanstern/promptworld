---
id: TASK-128
title: Stage-shaped TUI layout defaults
status: Done
assignee: []
created_date: '2026-07-25 14:45'
updated_date: '2026-07-26 02:56'
labels:
  - learning-game
  - tui
dependencies:
  - TASK-123
priority: medium
ordinal: 98000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient 2026-07-25 (decision 3), Wave 4. Which panels/tabs/chrome are visible BY DEFAULT is stage-resolved (stage-1 boots map + narrated chronicle + guardian line + lesson row; traces/raw feed/systems surface as stages arrive, each announced by a first-occurrence lesson). Defaults only: everything reachable at every stage and via ?; pre-ladder worlds get everything; capability locks stay angel-only (spec 046 doctrine untouched).

Spec: specs/066-stage-defaults
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Stage-resolved default visibility exists; every surface remains reachable at every stage
- [x] #2 Pre-ladder worlds byte-identical to ungated full layout
- [x] #3 Spec phase: Setup
- [x] #4 Spec phase: Foundational (blocking all user stories)
- [x] #5 Spec phase: User Story 1 — A stage-1 player boots into the focused layout (P1) 🎯 MVP
- [x] #6 Spec phase: User Story 2 — Pre-ladder worlds are untouched (P1)
- [x] #7 Spec phase: User Story 3 — Surfaces arrive with the stage (P2)
- [x] #8 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
spec-bridge link: specs/066-stage-defaults attached (spec+plan+tasks complete, 16 tasks; derived status In Progress — spec phase done, implementation not dispatched). Dispatch gated on TASK-119's merge per runbook Lane 4 (128 runs after the tabs/rows it governs exist: 125 ✓, 117 ✓, 119 pending).

Dispatched (UI-sweep orchestrator): spec-implementer on Sonnet. Rubric: single-package TUI layout/view code with tests alongside — routine tier per constitution Principle V; WATCH ITEM: touches every mode's layout — escalate one-way to Opus if gates fail (recorded per runbook Lane 4). Worktree .worktrees/task-128 cut from post-119 main (828686b) — everything the spec governs now exists except the villager strip (TASK-129, tolerated-absent by design). Expect rebases over #99/#100 merges.

spec-bridge sync: Setup: 1/1 · Foundational: 4/4 · US1: 4/4 · US2: 1/1 · US3: 4/4 · Polish: 2/2 — status In Progress → Done

Merged via PR #102 (squash 24ae434). Human ACs on merge evidence: #1 stage-resolved starting set exists (resolveStageDefaults + parity sweep vs the authority page) with every surface reachable at every stage (reachability sweep test, SC-003); #2 pre-ladder byte-identity proven by 7 pre-wiring sha256 golden frames still identical post-wiring (SC-002) + unrecognized-stage fail-open. Gate-reviewed deviations recorded: T007 partial (exercise-tab/incident-vocabulary stay on existing independently-tested mechanisms, table-equivalence proven by test, documented on the authority page); US3 machinery is wired+tested forward-compatible plumbing (today no numbered stage turns a surface ON, and no toggle UX exists yet — the override structure is precedence-proven against direct input). Sonnet tier, no escalation needed. Authority page shipped + re-pinned to squash on main.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Setup: 1/1 · Foundational (blocking all user stories): 4/4 · User Story 1 — A stage-1 player boots into the focused layout (P1) 🎯 MVP: 4/4 · User Story 2 — Pre-ladder worlds are untouched (P1): 1/1 · User Story 3 — Surfaces arrive with the stage (P2): 4/4 · Polish & Cross-Cutting Concerns: 2/2). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
