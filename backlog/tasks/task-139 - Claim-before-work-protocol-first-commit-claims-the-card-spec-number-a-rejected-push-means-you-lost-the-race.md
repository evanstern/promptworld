---
id: TASK-139
title: >-
  Claim-before-work protocol: first commit claims the card + spec number; a
  rejected push means you lost the race
status: In Progress
assignee: []
created_date: '2026-07-25 20:22'
updated_date: '2026-07-25 21:55'
labels:
  - gates
  - process
  - review-2026-07-25
dependencies: []
priority: high
ordinal: 109000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Operator + review agreement (2026-07-25). The spec-number race has now fired FIVE times in one day, the last one caused by the fix for the fourth: the review session renumbered grounded-feedback to 063 (ff9f8e1) while the MVLS session was concurrently claiming 063 for needs-conditioned-recovery (e9e75c1), forcing a second reclaim to 064 (4387d32). Earlier the same evening two sessions independently executed the SAME 059 renumber, and a board-notes push collided the same way. In every case neither session could see the other's in-flight decision at the moment it acted.

THE INSIGHT: git push rejection IS the compare-and-swap. There is no lock file in git, but there IS a serialization point — whoever pushes first wins and the loser gets a non-fast-forward rejection. Today a rejected push reads as 'annoying, rebase and carry on', which is precisely how the duplicate work happened. Treating the rejection as SIGNAL turns a convention into mutual exclusion.

THE PROTOCOL:
1. The FIRST commit of any task claims it — move the board card to In Progress AND claim the spec number by creating its directory (empty is fine) — before any spec authoring or code.
2. Push immediately. Never force-push.
3. A REJECTED PUSH MEANS YOU LOST THE RACE. Fetch, re-read the board and specs/. If another session now holds that task or that number, STOP the lane and surface it to the operator. This is the existing no-duplication doctrine with a mechanical trigger instead of a polite request.

WHAT IT FIXES, AND WHAT IT DOES NOT — do not oversell:
- Two sessions claiming the same TASK: yes, exactly this.
- Two sessions claiming the same SPEC NUMBER: yes, but ONLY because step 1 folds the number claim into the same push. Without that clause it does nothing for numbers.
- In-flight CODE invisible from other clones: NO. Card moves already propagate today (seven tasks read In Progress in every clone while their branches existed on one machine only). That needs the separate push-on-first-commit rule for task branches — pair the two, do not conflate them.

Enforcement home: scripts/check-merge-drift.mjs (session + worktree modes) plus the existing PreToolUse hook. Cutting a worktree for a task whose card is not already In Progress on origin/main should warn; creating a spec directory whose number is already taken on origin/main should block. The takenSpecNumbers() helper already computes the right thing — it just runs too late to prevent anything.

Spec: specs/065-claim-before-work
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Doctrine in CLAUDE.md + runbook template: first commit claims card and spec number, pushed immediately; a rejected push is a stop-the-lane signal
- [ ] #2 check-merge-drift worktree mode warns when the task card is not In Progress on origin/main
- [ ] #3 Creating a specs/NNN-* whose number is already taken on origin/main is blocked at claim time, not detected after
- [ ] #4 Task branches push on first commit so in-flight work is auditable from any clone
- [ ] #5 Two-session simulation shows the second session is stopped rather than duplicating (test or documented manual run)
- [ ] #6 Spec phase: Setup
- [ ] #7 Spec phase: Foundational (blocking prerequisites for the gate stories)
- [ ] #8 Spec phase: User Story 2 — the gates stop the second session mechanically (P1) 🎯 MVP
- [ ] #9 Spec phase: User Story 3 — in-flight work auditable from any clone (P2)
- [ ] #10 Spec phase: User Story 1 — doctrine: the protocol itself (P1)
- [ ] #11 Spec phase: User Story 4 — two-session race simulation (P2)
- [ ] #12 Spec phase: Polish & cross-cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Tier decision: Opus 4.8 (rubric: doctrine-adjacent behavior change — modifies spec 051 merge-drift gate semantics + CLAUDE.md doctrine; hook enforcement layer). Spec 065 authored, planned, tasked, linked (7 phase ACs). Contract wrinkle caught at plan review: worktree --spec must become claim-aware (--spec NNN --task TASK-n passes on Spec-marker attribution) or every claimed task's own worktree would be blocked. Dispatching spec-implementer on task-139-claim-before-work worktree.
<!-- SECTION:NOTES:END -->
