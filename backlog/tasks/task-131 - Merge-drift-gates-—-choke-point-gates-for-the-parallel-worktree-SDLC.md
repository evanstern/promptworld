---
id: TASK-131
title: Merge-drift gates — choke-point gates for the parallel worktree SDLC
status: In Progress
assignee: []
created_date: '2026-07-25 17:26'
updated_date: '2026-07-25 17:27'
labels: []
dependencies: []
ordinal: 101000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Deterministic gate script run at three choke points (session start, worktree cut, PR open) that predicts textual merge conflicts, surfaces clean-merging semantic collisions (backlog/, wiki-pinned sources, internal/tui/, spec-number collisions), prescribes post-merge janitor cleanup, and records findings as board notes — no daemon, no GitHub Actions; gates never touch a live task's branch.

Spec: specs/051-merge-drift-gates
<!-- SECTION:DESCRIPTION:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
spec-bridge link: spec 051 specify phase complete (spec.md + requirements checklist, all items pass, 0 clarification markers). Design decisions settled pre-spec with Evan: gates-only (no daemon/CI), no external annotation, gates never touch live branches, findings land as board notes. No tasks.md yet — phase ACs will seed at sync after speckit-tasks.
<!-- SECTION:NOTES:END -->
