---
id: TASK-180
title: >-
  root-guard blocks sanctioned board-sync commits: parentheses in the message,
  and non-ASCII card filenames
status: To Do
assignee: []
created_date: '2026-08-01 04:25'
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
