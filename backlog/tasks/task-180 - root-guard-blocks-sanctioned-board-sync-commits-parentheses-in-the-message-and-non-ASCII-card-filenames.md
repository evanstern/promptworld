---
id: TASK-180
title: >-
  root-guard blocks sanctioned board-sync commits: parentheses in the message,
  and non-ASCII card filenames
status: To Do
assignee: []
created_date: '2026-08-01 04:25'
updated_date: '2026-08-03 00:56'
labels:
  - debt
dependencies: []
priority: medium
ordinal: 148000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The root-guard hook rejects two shapes of the board-sync commit it is supposed to allow, so a legitimate card move gets blocked and the operator has to reword the commit message to get past a gate that was never aimed at them.

As the operator, when I move a card and commit it with the message style this repo already uses — 'board-sync: card TASK-178 (wiki size-debt wave 2)' — I want the commit to land, not to be told the root checkout is read-only.

As a session claiming a task, when the card's title contains an em dash, I want 'git add' followed by a bare 'git commit' to be recognized as board-sync, instead of silently failing the staged-path check.

Two independent defects in scripts/hooks/root-guard-hook.mjs, both hit live while claiming TASK-179:

1. Parentheses in the commit message. Reproduced: 'git commit -m "board-sync: card TASK-178 (wiki size-debt wave 2)" -- backlog/x.md' exits 2 (blocked); the same command with the parens removed exits 0. The command splitter treats '(' / ')' as shell grouping, so the message's words become argv tokens and land in the pathspec list, where they fail boardPathspecOk. Note the repo's own history uses parenthesized board-sync messages (bfdb57c2), so this blocks the house style.

2. Quoted non-ASCII staged paths. On the no-pathspec branch, isBoardSyncCommit reads 'git diff --cached --name-only' and tests files.every(f => f.startsWith('backlog/')) — root-guard-hook.mjs:311-315. Git C-quotes any path with non-ASCII bytes, so a card whose title has an em dash comes back as '"backlog/tasks/task-179 - ...\342\200\224..."' and the check fails on the leading quote. Card titles in this repo routinely contain em dashes. Fix is -z (or core.quotePath=false) plus NUL splitting.

Neither defect is a security hole — both fail closed, blocking allowed work rather than permitting forbidden work — but each forces the operator to fight a gate, which is exactly the posture CLAUDE.md tells sessions not to adopt.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A board-sync commit whose message contains parentheses is allowed, with a regression test pinning the exact TASK-178-style message
- [ ] #2 A board-sync commit staging only a card file whose name contains non-ASCII characters is allowed via the no-pathspec staged-set branch, with a regression test using an em-dash filename
- [ ] #3 Both fixes preserve fail-closed behavior: no commit touching a path outside backlog/ becomes allowed, covered by existing or new negative tests
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
THIRD + FOURTH defect, same hook, hit live 2026-08-02 while claiming TASK-173 (sweep). Both belong in this card's one fix pass.

3. Shell-redirection tokens parsed as pathspecs. parseGitInvocation stops the segment at the first char in /[;|&\n`)]/ (root-guard-hook.mjs:182). A sanctioned board-sync commit written as 'git commit -m "board-sync: ..." -- backlog/x.md 2>&1 | tail -5' is truncated at the '&' of '2>&1', leaving '2>' as a trailing token that isBoardSyncCommit reads as a pathspec; boardPathspecOk('2>') is false, so the commit blocks. Reproduced three times in a row; the identical command with the redirection removed exits 0. This is the SAME root cause as defect 1 (the boundary regex doubling as a tokenizer) and should share its fix and its regression test — pin a message-with-parens case AND a trailing-redirection case.

4. The worktree carve-out does not recognise the harness's isolation root. pre-write blocks any write under the root checkout that is not gitignored; the sanctioned escape is '.worktrees/<task-branch>'. Claude Code background jobs isolate via EnterWorktree, which always creates worktrees under '.claude/worktrees/<name>' - a TRACKED subtree here - so every write in a harness-created worktree is blocked as a root-checkout write. praxisflux's pdlc:sweep documents exactly that path in its 'Background-job / no-main-push execution mode', so the two doctrines do not overlap at all. Workaround used: 'git worktree add .worktrees/task-173 ...' at root, then EnterWorktree with path: pointing at it (satisfies the harness isolation guard and root-guard both). Durable fix: treat any path under a directory listed by 'git worktree list' as worktree-authored, or at minimum add '.claude/worktrees/' to the carve-out and to .gitignore.
<!-- SECTION:NOTES:END -->
