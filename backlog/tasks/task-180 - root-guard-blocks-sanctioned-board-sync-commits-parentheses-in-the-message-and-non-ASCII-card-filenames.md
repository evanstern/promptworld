---
id: TASK-180
title: >-
  root-guard blocks sanctioned board-sync commits: parentheses in the message,
  and non-ASCII card filenames
status: To Do
assignee: []
created_date: '2026-08-01 04:25'
updated_date: '2026-08-03 03:59'
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

TWO MORE INSTANCES observed 2026-08-02 during TASK-188's board sync, same parser, same family as this card's parenthesis and non-ASCII-filename cases. Both blocked a legitimately-scoped board-sync commit at root; neither has a bypass; each cost a retry.

(a) Shell REDIRECTION and PIPE tokens parse as pathspecs. A board-sync commit using the -F message-file form, followed by a stderr redirect and a pipe into tail, was blocked - even though -F is correctly listed in COMMIT_SHORT_WITH_VALUE (scripts/hooks/root-guard-hook.mjs:238) and the staged set was exactly one backlog/ card file. isBoardSyncCommit() walks the token list after the subcommand; -F correctly consumes its value, but the trailing redirect, pipe, and tail tokens then fall through the "not a flag" branch into pathspecs, and boardPathspecOk() rejects them as non-backlog. Workaround in use: run board-sync commits bare, with no redirection and no pipe.

(b) Heredoc PAYLOAD text is scanned as if it were commands. Authoring a plain notes file with a shell heredoc was blocked because the PROSE inside the heredoc quoted a git-commit example. The guard matched quoted text that is data, never executed. Workaround in use: author note files with the Write tool instead of a heredoc.

Shared root cause with this card's existing cases: the guard tokenizes the raw bash string rather than parsing shell grammar, so anything that is not an argument to the git invocation - redirections, pipe segments, heredoc bodies, message text - can be mistaken for one. A fix that only special-cases each new token shape will keep accreting; the durable fix is to isolate the actual git argv (stop at the first redirection or pipe operator, ignore heredoc bodies) before the pathspec walk.

NON-ASCII FILENAME CASE REPRODUCED AND ROOT-CAUSED (2026-08-02, committing TASK-189's card, whose title carries an em dash).

Mechanism, confirmed by reading the hook: with no explicit pathspecs the guard falls back to the staged set (root-guard-hook.mjs:311-315), which it reads via `git diff --cached --name-only`. Git C-QUOTES any path containing non-ASCII bytes, so the em-dash card comes back as a token that literally begins with a double-quote character. The check is files.every((f) => f.startsWith('backlog/')), which is false for a C-quoted path, so a correctly-scoped single-card board-sync commit blocks. Fix candidate: pass -z (git already does the right thing with it, and taskCardsOnOriginMain in check-merge-drift.mjs:471 ALREADY uses -z for exactly this reason and cites it in a comment), or set core.quotePath=false on the read, or strip C-quoting before the prefix test.

WORKAROUND THAT WORKS, verified: pass the card as an EXPLICIT double-quoted pathspec after --. parseGitInvocation's tokenizer (root-guard-hook.mjs:186-189) keeps a double-quoted token whole via /("[^"]*"|'[^']*'|\S+)/ and stripQuotes removes the wrapper, so the real path reaches boardPathspecOk with its leading backlog/ intact and passes. The explicit-pathspec branch (line 305-308) never consults the staged set, so C-quoting never enters the picture.

A WORKAROUND THAT DOES NOT WORK: retitling the task to remove the non-ASCII character. The backlog CLI updates the title inside the card but does NOT rename the file on disk, so the em dash stays in the filename and the block persists.

MECHANISM FOR THE REDIRECTION CASE (a) above, now also root-caused: parseGitInvocation stops the segment at the first character in /[;|&\n`)]/. For a commit followed by a stderr redirect, that boundary lands on the ampersand INSIDE the redirect, leaving a truncated trailing fragment as a token. It does not start with a dash, so it falls through to the pathspec list, fails boardPathspecOk, and blocks. The redirect was never an argument to git at all.

All three cases on this card share one cause: the guard tokenizes the raw bash string instead of parsing shell grammar, and separately trusts porcelain output that git deliberately quotes. Suggested fix order: (1) -z on the staged-set read, cheapest and closes the non-ASCII case outright; (2) stop the git segment at a redirection operator rather than at a bare ampersand; (3) ignore heredoc bodies.

Confirmed 2026-08-02 during TASK-191, and a workaround found.

Reproduction: TASK-189's card filename carries a non-ASCII em-dash. Staging it and committing with no pathspec —

  git add "backlog/tasks/task-189 - ...—...md"
  git commit -F <message-file>

— is blocked with "blocked `git commit` — direct commits at the root checkout are forbidden EXCEPT board-sync commits scoped entirely to backlog/". The commit WAS scoped entirely to backlog/ (a single staged path, verified with git diff --cached --name-only). With no pathspec on the command line the hook falls back to inspecting staged paths, and git renders a non-ASCII path quoted and octal-escaped ("backlog/tasks/task-189 - ...\342\200\224...md"). The leading double-quote defeats the under-backlog/ prefix test, so a sanctioned commit is rejected.

WORKAROUND (no config change, no bypass): pass the path as an explicit unquoted glob pathspec, so the hook parses the command-line argument instead of git's escaped output —

  git commit -F <message-file> backlog/tasks/task-189*.md

This succeeded on the identical staged state that had just been rejected. The glob is expanded by the shell into a literal argument beginning with backlog/, which the prefix test accepts. It is the sanctioned shape ("every pathspec under backlog/"), merely expressed in a form the parser can read.

Note the workaround is fragile in exactly the way the bug is: quoting the pathspec to survive the spaces in the filename reintroduces a leading quote character. The glob happens to avoid needing quotes.

Suggested fix: have the hook enumerate staged paths with -z (git diff --cached --name-only -z), which emits raw unquoted NUL-delimited paths and sidesteps quotePath entirely, rather than parsing the human-facing quoted form.
<!-- SECTION:NOTES:END -->
