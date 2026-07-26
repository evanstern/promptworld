---
id: TASK-160
title: >-
  Root checkout is read-only: worktree+merge-only rule, hook-enforced (no
  rebases)
status: In Progress
assignee: []
created_date: '2026-07-26 21:19'
updated_date: '2026-07-26 21:20'
labels: []
dependencies: []
priority: high
ordinal: 128000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Operator directive (2026-07-26): NOTHING is modified in the root checkout directly — every change is authored on a branch in .worktrees/ and lands on main only by merge (PR or manual 'git merge --no-ff' at root). Rebases are forbidden repo-wide. Enforce at the harness level (PreToolUse hooks) plus CLAUDE.md doctrine.

Trivial-exemption rationale (constitution Dev Workflow): surgical change with complete diagnosis — files are CLAUDE.md (worktrees block + claim protocol + wiki-in-PR step 7), .claude/settings.json (hook wiring), and a new scripts/hooks/root-guard-hook.mjs modeled on scripts/hooks/merge-drift-hook.mjs. ACs on this task.

Supersedes the 'derived state commits directly to main' practice: board/spec-bridge/tasks.md bookkeeping is authored on a branch and merged back. Claim protocol (spec 065) reconciliation: the claim commit is authored on the task branch and lands on main via an immediate manual merge from root — push rejection on that merge-push remains the mutual-exclusion signal.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 scripts/hooks/root-guard-hook.mjs blocks at root: git commit (except concluding a conflicted merge, MERGE_HEAD present), cherry-pick, revert, am, merge --squash, and branch creation (checkout -b / switch -c); blocks git rebase anywhere in the repo; blocks force-pushes; fails open on malformed stdin, out-of-jurisdiction paths, or internal errors
- [ ] #2 Write/Edit/NotebookEdit to tracked (non-gitignored) files in the root checkout are blocked; paths under .worktrees/ and gitignored paths pass
- [ ] #3 .claude/settings.json wires the guard: PreToolUse on Bash and on Write|Edit|NotebookEdit, alongside the existing merge-drift hooks
- [ ] #4 CLAUDE.md: worktrees block rewritten as the iron-clad root-read-only rule (worktree + merge only, no rebases, sanctioned root git ops enumerated); claim-before-work and wiki-in-PR step-7 blocks amended for consistency
- [ ] #5 Guard verified against stdin fixtures in a scratch repo: root commit blocked, worktree commit allowed, rebase blocked everywhere, merge at root allowed, merge-concluding commit allowed, root Edit blocked, gitignored/worktree Edit allowed
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Claim: this card move commits on task-160-root-read-only-guard and lands on main via immediate manual merge from root (dogfoods the new claim flow).
2. Delegate implementation (root-guard-hook.mjs + settings.json wiring + CLAUDE.md doctrine edits) to spec-implementer on Opus 4.8 — tier justification: doctrine-adjacent enforcement change; the hook gates every future session's git behavior, so fail-open/fail-closed correctness is critical.
3. Verify with stdin fixtures in a scratch repo (AC5), tick ACs, Done with final summary.
4. Land: manual 'git merge --no-ff' at root + push (operator ratified in-session; no PR needed per operator: 'either a PR or a manual merge').
5. Post-merge: reconcile auto-memory notes that encode the superseded direct-to-main bookkeeping practice.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Tier: spec-implementer @ Opus 4.8. Rubric: doctrine-adjacent behavior change (SDLC enforcement hook); not routine single-package feature work.
<!-- SECTION:NOTES:END -->
