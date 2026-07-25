---
id: TASK-131
title: Merge-drift gates — choke-point gates for the parallel worktree SDLC
status: Done
assignee: []
created_date: '2026-07-25 17:26'
updated_date: '2026-07-25 18:19'
labels: []
dependencies: []
ordinal: 101000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Deterministic gate script run at three choke points (session start, worktree cut, PR open) that predicts textual merge conflicts, surfaces clean-merging semantic collisions (backlog/, wiki-pinned sources, internal/tui/, spec-number collisions), prescribes post-merge janitor cleanup, and records findings as board notes — no daemon, no GitHub Actions; gates never touch a live task's branch.

Spec: specs/051-merge-drift-gates
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Spec phase: Setup
- [x] #2 Spec phase: Foundational (Blocking Prerequisites)
- [x] #3 Spec phase: User Story 1 — PR gate: no doomed PR gets opened (Priority: P1) 🎯 MVP
- [x] #4 Spec phase: User Story 2 — Session-start janitor (Priority: P2)
- [x] #5 Spec phase: User Story 3 — Worktree-cut gate (Priority: P3)
- [x] #6 Spec phase: User Story 4 — Findings become board artifacts (Priority: P3)
- [x] #7 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
spec-bridge link: spec 051 specify phase complete (spec.md + requirements checklist, all items pass, 0 clarification markers). Design decisions settled pre-spec with Evan: gates-only (no daemon/CI), no external annotation, gates never touch live branches, findings land as board notes. No tasks.md yet — phase ACs will seed at sync after speckit-tasks.

speckit-plan complete: plan.md (Constitution Check PASS x5, Sonnet tier justified — single-package script, no concurrency/doctrine logic), research.md (11 decisions incl. merge-tree --write-tree verified live on git 2.50.1, exit-code contract 0/1/2, squash-detection via empty-contribution tree equality), data-model.md, contracts/ (gate-cli, detection-rules, report-schema), quickstart.md (5 fixture-repo validation scenarios). Next: speckit-tasks.

spec-bridge sync: Setup: 0/1 · Foundational (Blocking Prerequisites): 0/3 · User Story 1 — PR gate: no doomed PR gets opened (Priority: P1) 🎯 MVP: 0/3 · User Story 2 — Session-start janitor (Priority: P2): 0/5 · User Story 3 — Worktree-cut gate (Priority: P3): 0/2 · User Story 4 — Findings become board artifacts (Priority: P3): 0/2 · Polish & Cross-Cutting: 0/3

spec-bridge sync: Setup: 1/1 · Foundational (Blocking Prerequisites): 3/3 · User Story 1 — PR gate: no doomed PR gets opened (Priority: P1) 🎯 MVP: 3/3 · User Story 2 — Session-start janitor (Priority: P2): 5/5 · User Story 3 — Worktree-cut gate (Priority: P3): 2/2 · User Story 4 — Findings become board artifacts (Priority: P3): 2/2 · Polish & Cross-Cutting: 2/3

Implementation complete: PR #85 open (branch task-131-merge-drift-gates, 3 commits, base f5e5855). All quickstart scenarios validated by spec-implementer (Sonnet): S1 conflict block exit 1 naming file; S2 squash cleanup (empty-contribution) + --apply-cleanup scoped correctly, dirty worktree excluded; S3 all three worktree-gate exits; S4 notes append-once + fingerprint dedup; S5 session ~7-9s, no foreign reflog entries (FR-009 held). Contract §4 amended on main (conditional -d/-D). Dogfooded: pr gate on its own branch → exit 0/warnings (stale-base 36, merge clean). T019 + Done await merge.

spec-bridge sync: Setup: 1/1 · Foundational (Blocking Prerequisites): 3/3 · User Story 1 — PR gate: no doomed PR gets opened (Priority: P1) 🎯 MVP: 3/3 · User Story 2 — Session-start janitor (Priority: P2): 5/5 · User Story 3 — Worktree-cut gate (Priority: P3): 2/2 · User Story 4 — Findings become board artifacts (Priority: P3): 2/2 · Polish & Cross-Cutting: 3/3 — status In Progress → Done
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Setup: 1/1 · Foundational (Blocking Prerequisites): 3/3 · User Story 1 — PR gate: no doomed PR gets opened (Priority: P1) 🎯 MVP: 3/3 · User Story 2 — Session-start janitor (Priority: P2): 5/5 · User Story 3 — Worktree-cut gate (Priority: P3): 2/2 · User Story 4 — Findings become board artifacts (Priority: P3): 2/2 · Polish & Cross-Cutting: 3/3). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
