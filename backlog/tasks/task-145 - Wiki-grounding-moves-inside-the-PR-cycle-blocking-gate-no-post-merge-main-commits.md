---
id: TASK-145
title: >-
  Wiki grounding moves inside the PR cycle: blocking gate, no post-merge main
  commits
status: To Do
assignee: []
created_date: '2026-07-26 15:38'
labels:
  - process
  - pdlc
dependencies: []
priority: high
ordinal: 115000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Operator direction (2026-07-26, during TASK-141): the lifecycle becomes (1) design (2) code (3) approval (4) wiki grounding (5) PR (6) merge (7) close task + commit main — wiki updates BELONG TO THE PR; 'PR merges then we add more to main afterward' is explicitly rejected. Current state: wiki-update runs at root after merge (CLAUDE.md PDLC block, plan D5 pattern), and scripts/check-merge-drift.mjs in pr mode only WARNS on wiki-sources-overlap (TASK-141 shipped with 10 such warnings and a 3-commit post-merge tail on main — the motivating case). Design surface: (a) the task branch re-pins/rewrites affected docs/wiki notes BEFORE the PR is opened, so one PR carries code+wiki atomically; (b) escalate the gate's wiki-sources-overlap from warn to BLOCK when the branch touches pinned sources without re-pinning those notes in the same branch (branch-local verified_against must equal branch HEAD or an ancestor — needs a pin-vs-branch rule since the merge commit hash doesn't exist yet at PR time); (c) decide player-docs placement (derived from wiki — same PR, or a sanctioned post-merge derivation?); (d) concurrent-lane collision posture: two branches re-pinning the same note merge serially (sweep doctrine) — spell out the loser's re-pin obligation; (e) step 7 'close task + commit main' — operator believes this is already designed somewhere (check spec 065 claim protocol + pdlc:sweep runbook merge-serial flow) — reconcile rather than re-invent. Upstream: grounding-wiki skill prose ('commit together with or immediately after') and the merge-drift gate live in praxisflux (~/neumo/projects/praxis) — plugin change, version-lockstep laws apply.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Lifecycle documented: CLAUDE.md PDLC block + constitution Principle IV updated to the in-PR wiki gate
- [ ] #2 check-merge-drift pr-mode gate BLOCKS on wiki-sources-overlap when the branch itself does not re-pin the touched notes
- [ ] #3 player-docs placement decided and documented (in-PR vs derived post-merge)
- [ ] #4 Step 7 (close task + commit main) reconciled with the existing planned design, named by artifact
<!-- AC:END -->
