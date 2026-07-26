---
id: TASK-145
title: >-
  Wiki grounding moves inside the PR cycle: blocking gate, no post-merge main
  commits
status: Done
assignee: []
created_date: '2026-07-26 15:38'
updated_date: '2026-07-26 16:16'
labels:
  - process
  - pdlc
dependencies: []
priority: high
ordinal: 115000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Operator direction (2026-07-26, during TASK-141): the lifecycle becomes (1) design (2) code (3) approval (4) wiki grounding (5) PR (6) merge (7) close task + commit main — wiki updates BELONG TO THE PR; 'PR merges then we add more to main afterward' is explicitly rejected. Spec 069 decides: pin-vs-branch blocking predicate (wiki-repin-missing), in-PR player-docs freshness block, merge-commit-only doctrine, step 7 named as spec 065 + pdlc:sweep re-ground (derived-state-only post-merge commits), no bypass flag. See specs/069-wiki-in-pr-gate/ for the full decision record.

Spec: specs/069-wiki-in-pr-gate
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Lifecycle documented: CLAUDE.md PDLC block + constitution Principle IV updated to the in-PR wiki gate
- [x] #2 check-merge-drift pr-mode gate BLOCKS on wiki-sources-overlap when the branch itself does not re-pin the touched notes
- [x] #3 player-docs placement decided and documented (in-PR vs derived post-merge)
- [x] #4 Step 7 (close task + commit main) reconciled with the existing planned design, named by artifact
- [x] #5 Spec phase: Setup
- [x] #6 Spec phase: Foundational
- [x] #7 Spec phase: User Story 1 — A code PR cannot open without its wiki grounding (P1)
- [x] #8 Spec phase: User Story 2 — Player docs cannot go stale through a merge (P1)
- [x] #9 Spec phase: User Story 3 — The doctrine says what the gate enforces (P2)
- [x] #10 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Claim per spec 065: specs/069-wiki-in-pr-gate/ stubbed. Sweep lane 1 (runbook docs/design/pdlc-hardening-runbook.md, signed off 2026-07-26). Tier: Opus 4.8 for gate/hook code (doctrine-adjacent SDLC-critical infra; a defect blocks every future PR); constitution amendment stays planning-tier via speckit-constitution. Operator decisions recorded: player-docs regenerate IN the PR; praxisflux upstream out of scope.

spec-bridge sync: Setup: 1/1 · Foundational: 1/1 · User Story 1 — A code PR cannot open without its wiki grounding (P1): 2/2 · User Story 2 — Player docs cannot go stale through a merge (P1): 2/2 · User Story 3 — The doctrine says what the gate enforces (P2): 2/2 · Polish & Cross-Cutting: 1/2

spec-bridge sync: Setup: 1/1 · Foundational: 1/1 · User Story 1 — A code PR cannot open without its wiki grounding (P1): 2/2 · User Story 2 — Player docs cannot go stale through a merge (P1): 2/2 · User Story 3 — The doctrine says what the gate enforces (P2): 2/2 · Polish & Cross-Cutting: 2/2 — status In Progress → Done
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Setup: 1/1 · Foundational: 1/1 · User Story 1 — A code PR cannot open without its wiki grounding (P1): 2/2 · User Story 2 — Player docs cannot go stale through a merge (P1): 2/2 · User Story 3 — The doctrine says what the gate enforces (P2): 2/2 · Polish & Cross-Cutting: 2/2). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
