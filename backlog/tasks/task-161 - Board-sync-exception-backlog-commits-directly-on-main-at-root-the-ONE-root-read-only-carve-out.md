---
id: TASK-161
title: >-
  Board-sync exception: backlog/ commits directly on main at root (the ONE
  root-read-only carve-out)
status: In Progress
assignee: []
created_date: '2026-07-26 22:10'
updated_date: '2026-07-26 22:10'
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
- [ ] #1 root-guard pre-bash allows 'git commit' at root when NO -a/--all/--interactive/--patch flag is present AND every explicit pathspec argument (if any) is under backlog/ AND, when no pathspecs are given, the staged set is non-empty and entirely under backlog/; all other root commits (mixed staged paths, non-backlog pathspecs, empty stage, -a) stay blocked; MERGE_HEAD carve-out unchanged
- [ ] #2 pre-write still blocks Write/Edit of backlog/ at root (hand-editing the board stays forbidden; only the CLI path via Bash is sanctioned)
- [ ] #3 CLAUDE.md: root-read-only block gains the board-sync exception bullet (edit via CLI at root, commit scoped to backlog/ only, push immediately, branches never commit backlog/); claim-before-work block rewritten (card move commits directly at root + push = mutex event; spec stub rides branch + immediate merge); wiki-in-PR step 7 updated (board moves via the exception; other derived state still branch+merge)
- [ ] #4 Fixture suite extended and green: staged-backlog-only commit at root allowed; mixed staged blocked; non-backlog staged blocked; empty stage blocked; -a blocked; explicit backlog/ pathspec allowed; non-backlog pathspec blocked; -m message text never parsed as pathspec; worktree commits unaffected
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Tier: spec-implementer @ Opus 4.8 — doctrine-adjacent enforcement amendment (same rubric as TASK-160).
<!-- SECTION:NOTES:END -->
