---
id: TASK-131
title: Merge-drift gates — choke-point gates for the parallel worktree SDLC
status: In Progress
assignee: []
created_date: '2026-07-25 17:26'
updated_date: '2026-07-25 17:45'
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
- [ ] #1 Spec phase: Setup
- [ ] #2 Spec phase: Foundational (Blocking Prerequisites)
- [ ] #3 Spec phase: User Story 1 — PR gate: no doomed PR gets opened (Priority: P1) 🎯 MVP
- [ ] #4 Spec phase: User Story 2 — Session-start janitor (Priority: P2)
- [ ] #5 Spec phase: User Story 3 — Worktree-cut gate (Priority: P3)
- [ ] #6 Spec phase: User Story 4 — Findings become board artifacts (Priority: P3)
- [ ] #7 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
spec-bridge link: spec 051 specify phase complete (spec.md + requirements checklist, all items pass, 0 clarification markers). Design decisions settled pre-spec with Evan: gates-only (no daemon/CI), no external annotation, gates never touch live branches, findings land as board notes. No tasks.md yet — phase ACs will seed at sync after speckit-tasks.

speckit-plan complete: plan.md (Constitution Check PASS x5, Sonnet tier justified — single-package script, no concurrency/doctrine logic), research.md (11 decisions incl. merge-tree --write-tree verified live on git 2.50.1, exit-code contract 0/1/2, squash-detection via empty-contribution tree equality), data-model.md, contracts/ (gate-cli, detection-rules, report-schema), quickstart.md (5 fixture-repo validation scenarios). Next: speckit-tasks.

spec-bridge sync: Setup: 0/1 · Foundational (Blocking Prerequisites): 0/3 · User Story 1 — PR gate: no doomed PR gets opened (Priority: P1) 🎯 MVP: 0/3 · User Story 2 — Session-start janitor (Priority: P2): 0/5 · User Story 3 — Worktree-cut gate (Priority: P3): 0/2 · User Story 4 — Findings become board artifacts (Priority: P3): 0/2 · Polish & Cross-Cutting: 0/3
<!-- SECTION:NOTES:END -->
