---
id: TASK-188
title: >-
  Claim gate is blind to spec numbers held by pushed-but-unmerged branches
  (TASK-173/TASK-187 both hold 110)
status: Done
assignee: []
created_date: '2026-08-03 01:16'
updated_date: '2026-08-03 02:01'
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
- [x] #1 claim mode BLOCKS (exit 1) when the requested number is held by a different spec dirname on a pushed remote task branch, naming the branch, the taken dirname, and the owning task
- [x] #2 claim mode stays idempotent for the owner: re-claiming the SAME dirname passes whether it lives on origin/main, on the caller's own branch, or nowhere yet
- [x] #3 the block message reports a next-free number computed from BOTH origin/main and branch-held claims, so following the advice cannot land on a second collision
- [x] #4 the check is a pure read that fails closed on fetch failure, matching the existing claim-mode contract; no new writes, no new mutation surface
- [x] #5 regression tests cover: branch-held collision blocks, owner re-claim passes, main-held collision still blocks, and next-free skips branch-held numbers
- [x] #6 the live TASK-173/TASK-187 spec-110 collision is reproduced by the new check and recorded as evidence on this card
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
IMPLEMENTED + PR OPEN (2026-08-02). Spec 111 claimed and stubbed to main per spec 065 (stub merge 1b17cd33); full spec/plan/tasks on branch task-188-claim-gate-branch-visibility; PR #156 open, merge-commit only.

Change (scripts/check-merge-drift.mjs): new pure-read branchHeldSpecNumbers(cwd) scans already-fetched refs/remotes/origin/task-* via for-each-ref plus per-branch ls-tree, returning Map(number -> {dir, branch}); new nextFreeSpecNumber(...maps) takes next-free over the UNION of main-held and branch-held (also fixes a latent -Infinity when no specs exist); runClaim() now decides main-first then branch, so origin/main keeps attribution when both hold a number and its pre-111 message stays byte-identical. Scoped to origin/task-* so release/experiment branches never lock numbers. No new fetch, no writes, still fails closed (exit 2) on an unreachable remote; any git failure degrades to today's main-only behavior.

Blast radius bonus: scripts/hooks/merge-drift-hook.mjs invokes the gate as a subprocess in claim mode, so BOTH hook layers - pre-write into a colliding specs/NNN-*/ and pre-bash mkdir of one - start blocking with no hook change.

Verification: node --test scripts/claim-protocol.test.mjs 17/17 (7 new cases: branch-held collision blocks and names the branch, next-free skips branch-held numbers, owner re-claim passes with the holder on its own branch, main wins attribution when both hold, non-task remote branches are not claims). node --test scripts/check-merge-drift.test.mjs 31/31, no other mode regressed. pr gate: verdict=pass, no findings.

AC#6 live reproduction against this repo - all four of these PASSED before the fix:
  claim --dir 110-something-new  -> exit 1, names origin/task-173-absence-attribution (task=TASK-173)
  claim --dir 112-other-thing    -> exit 1, names origin/task-187-frame-harness (task=TASK-187)
  claim --dir 111-claim-gate-branch-visibility -> exit 0 (owner re-claim, holder on main plus own branch)
  claim --dir 113-fresh          -> exit 0 (unheld number)
NUANCE: the 110 DOUBLE-claim resolved itself mid-flight - TASK-187 renumbered off 110 to specs/112-tui-frame-harness while this task was being built, so only TASK-173 holds 110 now. The defect class is unchanged and both lanes still reproduce it branch-side.

Wiki grounding: no-op. No docs/wiki note lists any scripts/ path as a source (the corpus grounds the Go codebase, not the harness scripts); pr gate confirms - no wiki-repin-missing, no player-docs-stale.

Model tier (constitution Principle V v1.3.0): DEVIATION, recorded here and in plan.md. Implemented INLINE on the planning model (Opus 5, claude-opus-5) rather than dispatched to spec-implementer. Reason: this session's operator instructions bar spawning subagents unless requested, and the operator directed this session to perform the fix directly. Rubric-wise this was a Sonnet-tier slice - single file, surgical, complete file:line diagnosis pinned on this card before work started.

Accepted risk (recorded in plan.md): a stale unmerged branch now permanently blocks its number. Mitigation is in the message, which names the branch and prescribes deleting it on origin. A blocked claim is visible and one command from resolution; a duplicate claim costs two lanes' spec work.

OUT OF SCOPE, follow-up candidate: worktree --spec claim-awareness (check-merge-drift.mjs around line 1485) still reads main only. Claim mode is the first choke point so a blocked number never reaches worktree mode, but the asymmetry is real.

TWO MORE ROOT-GUARD PARSER DEFECTS FOUND, both belong on TASK-180 (same family as its parenthesis and non-ASCII-filename cases). Neither has a bypass; both cost a retry.
  (a) Shell redirection and pipe tokens parse as pathspecs. A board-sync commit using the -F message-file form, followed by a stderr redirect and a pipe into tail, was BLOCKED at root - even though -F is correctly listed in COMMIT_SHORT_WITH_VALUE and the staged set was a single backlog/ card file. The trailing redirect, pipe, and tail tokens were read as non-backlog pathspecs. Workaround: run board-sync commits bare, with no redirection and no pipe.
  (b) Heredoc PAYLOAD text is scanned as if it were commands. Writing this very notes file with a shell heredoc was blocked because the prose inside the heredoc quoted a git-commit example. The guard matched the quoted text, not an executed command. Workaround: author note files with the Write tool instead of a heredoc.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Closed the hole that let two concurrent sessions claim the same spec number. The claim gate defined spec-number ownership solely by presence on origin/main, so a number claimed on a pushed-but-unmerged branch was invisible to it — and on 2026-08-02 TASK-173 and TASK-187 both took 110, each with complete spec/plan/tasks, both gates green. It surfaced only when the spec-bridge Stop hook complained the cards exceeded their artifacts, two removes from the cause.

Fix (scripts/check-merge-drift.mjs): branchHeldSpecNumbers(cwd) reads already-fetched refs/remotes/origin/task-* via for-each-ref plus per-branch ls-tree; nextFreeSpecNumber(...maps) takes next-free over the union of main-held and branch-held (also retiring a latent -Infinity on an empty spec set); runClaim() decides main-first then branch, so origin/main keeps attribution when both hold a number and its pre-111 message is byte-identical. Scoped to origin/task-* so release and experiment branches never lock numbers. Pure read, no new fetch, still fails closed on an unreachable remote, degrades to main-only behavior on any git failure. Both merge-drift hook layers inherit the block for free, since they invoke the gate as a subprocess.

Verified: claim-protocol.test.mjs 17/17 with 7 new cases, check-merge-drift.test.mjs 31/31, pr gate pass. Live before the fix, all four passed; after, the two branch-held numbers block and name their holding branch while owner re-claim and a free number pass.

POST-MERGE CONFIRMATION on main (merge commit fe9fc10b, PR 156, merge-commit not squash, 2 parents verified): claim --dir 110-competing now exits 1 naming origin/task-173-absence-attribution and attributing TASK-173 — the branch-held path is live and actively protecting an in-flight lane. claim --dir 111-anything-else exits 1 via the unchanged main-held path. Spec 111 landed complete on main; worktree removed, branch deleted, root fast-forwarded.

Both colliding lanes resolved on their own during the work: TASK-187 renumbered to specs/112-tui-frame-harness and has since merged, leaving 110 cleanly TASK-173's. No card was ever set back and no spec content was moved to main early.

Tier: implemented inline on the planning model (Opus 5, claude-opus-5) rather than delegated, a recorded deviation from constitution Principle V — this session's operator instructions bar spawning subagents unless requested, and the operator directed this session to perform the fix. Rubric-wise a Sonnet-tier slice.

Two follow-ups left on the board rather than folded in: TASK-180 gained three root-caused parser defects found while doing this work (C-quoted staged paths, redirect-boundary tokens, heredoc bodies scanned as commands) with a verified workaround; TASK-189 cards the spec-bridge gate reading main while spec artifacts ride the PR, which is what raised the original alarm. Also noted out of scope: worktree --spec claim-awareness still reads main only.
<!-- SECTION:FINAL_SUMMARY:END -->
