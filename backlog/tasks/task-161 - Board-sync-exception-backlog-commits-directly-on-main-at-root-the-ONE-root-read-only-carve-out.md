---
id: TASK-161
title: >-
  Board-sync exception: backlog/ commits directly on main at root (the ONE
  root-read-only carve-out)
status: Done
assignee: []
created_date: '2026-07-26 22:10'
updated_date: '2026-07-26 23:57'
labels: []
dependencies: []
priority: high
ordinal: 129000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Operator directive (2026-07-26, amending TASK-160): the Backlog.md board must stay in sync ON MAIN — it is the plan of record and the concurrent-session mutex, and it cannot lag on task branches waiting for PR merges. Ratified as the ONLY exception to the root-read-only rule: backlog/ board state is edited at root via the backlog CLI, committed at root scoped to backlog/ paths only, and pushed immediately. Task branches never commit backlog/ going forward (single home for board state; avoids card-file merge conflicts). Claim protocol simplifies back toward original spec 065: the card→In Progress commit+push at root is the mutual-exclusion event; the spec dir stub still rides the task branch and lands via immediate merge (specs are NOT excepted).

Trivial-exemption rationale: surgical amendment with complete diagnosis — scripts/hooks/root-guard-hook.mjs (commit rule carve-out), CLAUDE.md (root-read-only block, claim protocol, wiki-in-PR step 7), fixtures. Tier: spec-implementer @ Opus 4.8, same rubric as TASK-160 (doctrine-adjacent enforcement).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 root-guard pre-bash allows 'git commit' at root when NO -a/--all/--interactive/--patch flag is present AND every explicit pathspec argument (if any) is under backlog/ AND, when no pathspecs are given, the staged set is non-empty and entirely under backlog/; all other root commits (mixed staged paths, non-backlog pathspecs, empty stage, -a) stay blocked; MERGE_HEAD carve-out unchanged
- [x] #2 pre-write still blocks Write/Edit of backlog/ at root (hand-editing the board stays forbidden; only the CLI path via Bash is sanctioned)
- [x] #3 CLAUDE.md: root-read-only block gains the board-sync exception bullet (edit via CLI at root, commit scoped to backlog/ only, push immediately, branches never commit backlog/); claim-before-work block rewritten (card move commits directly at root + push = mutex event; spec stub rides branch + immediate merge); wiki-in-PR step 7 updated (board moves via the exception; other derived state still branch+merge)
- [x] #4 Fixture suite extended and green: staged-backlog-only commit at root allowed; mixed staged blocked; non-backlog staged blocked; empty stage blocked; -a blocked; explicit backlog/ pathspec allowed; non-backlog pathspec blocked; -m message text never parsed as pathspec; worktree commits unaffected
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Tier: spec-implementer @ Opus 4.8 — doctrine-adjacent enforcement amendment (same rubric as TASK-160).

Field bugs found by the faith-directives sweep orchestrator (2026-07-26), root-guard-hook.mjs board-sync path: (1) staged-set check tests startsWith backlog/ against git diff --cached --name-only, but git C-quotes non-ASCII paths (em-dash card filenames), so the quoted line starts with a doublequote and the exception fails CLOSED on exactly the cards it should allow — workaround applied at root: git config core.quotePath false; durable fix: strip C-quoting or run the subprocess with -c core.quotepath=false. (2) The pre-eval staged-set semantics mean 'git add X && git commit' compounds always fail (nothing staged at evaluation time) — stage and commit must be separate tool invocations; document or evaluate post-add. (3) Tokenizer appears to fail closed on parens in commit messages / parenthesized subshells containing git commit.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented by spec-implementer @ Opus 4.8 (commits 0cbea6c, a540b21), gated by the planning session (--amend denial added on review, checked before both allow paths). root-guard commit rule at root now: --amend denied outright; then MERGE_HEAD merge-conclusion allow; then the board-sync exception — no -a/--all/-p/--patch/-i/--include/--pathspec-from-file, explicit pathspecs (if any) all under backlog/, else staged set non-empty and entirely backlog/. Clustered short flags parsed char-wise (-am hole closed). pre-write unchanged: hand-editing backlog/ at root still blocked. CLAUDE.md: root-read-only block carve-out bullet, claim protocol as two immediate pushes (card move at root = mutex event; spec stub via branch + immediate merge), step 7 split (board via exception; other derived state branch+merge). 35/35 fixtures green incl. TASK-160 regressions.
<!-- SECTION:FINAL_SUMMARY:END -->
