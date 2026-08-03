---
id: TASK-188
title: >-
  Claim gate is blind to spec numbers held by pushed-but-unmerged branches
  (TASK-173/TASK-187 both hold 110)
status: In Progress
assignee: []
created_date: '2026-08-03 01:16'
updated_date: '2026-08-03 01:16'
labels:
  - gate
  - spec-065
  - concurrency
dependencies: []
priority: high
ordinal: 170001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The gate that stops two sessions from grabbing the same spec number only looks at work already merged into the main line, so two sessions running at the same time can both take the same number and neither is told. This makes the gate also look at numbers claimed on branches that have been pushed but not yet merged, so the collision is caught at claim time instead of days later.

As a session about to start work, when I claim a spec number, I want to be told immediately if another in-flight session already took it — not discover it after I have written a whole spec against that number.

As the operator, when two lanes do collide, I want the gate to have caught it at claim time, so I am not asked to arbitrate a renumber after both lanes already carry full spec/plan/tasks artifacts.

Live evidence (2026-08-02): TASK-173 and TASK-187 BOTH hold spec number 110 — task-173-absence-attribution carries specs/110-absence-attribution and task-187-frame-harness carries specs/110-tui-frame-harness, each with spec.md + plan.md + tasks.md, both branches pushed to origin, neither claim stub merged to main. Both passed the claim gate, and a probe confirms a third session would pass too:

    node scripts/check-merge-drift.mjs claim --dir 110-something-new  ->  verdict=pass, no findings

Diagnosis pinned: scripts/check-merge-drift.mjs takenSpecNumbers() (line 614) reads ONLY the origin/main tree, and runClaim() (line 1387) consults nothing else. The comment at lines 461-465 states the design intent explicitly — "The claim protocol defines ownership by presence on origin/main". Spec 065's protocol compensates by requiring the claim stub to be merged to main immediately via git merge --no-ff at root; when that step is skipped or delayed, the gate has no fallback and the exclusion window is wide open. Branch-vs-main collision detection already exists (specNumberCollisions(), line 635, used by session and pr modes) — branch-vs-branch is the gap.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 claim mode BLOCKS (exit 1) when the requested number is held by a different spec dirname on a pushed remote task branch, naming the branch, the taken dirname, and the owning task
- [ ] #2 claim mode stays idempotent for the owner: re-claiming the SAME dirname passes whether it lives on origin/main, on the caller's own branch, or nowhere yet
- [ ] #3 the block message reports a next-free number computed from BOTH origin/main and branch-held claims, so following the advice cannot land on a second collision
- [ ] #4 the check is a pure read that fails closed on fetch failure, matching the existing claim-mode contract; no new writes, no new mutation surface
- [ ] #5 regression tests cover: branch-held collision blocks, owner re-claim passes, main-held collision still blocks, and next-free skips branch-held numbers
- [ ] #6 the live TASK-173/TASK-187 spec-110 collision is reproduced by the new check and recorded as evidence on this card
<!-- AC:END -->
