---
id: TASK-132
title: Merge-drift gate hooks — harness enforcement of the spec 051 choke points
status: Done
assignee: []
created_date: '2026-07-25 18:34'
updated_date: '2026-07-25 18:47'
labels: []
dependencies: []
ordinal: 102000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Enforce the spec 051 merge-drift gates at the harness level via project hooks in .claude/settings.json: a PreToolUse hook on Bash that intercepts 'gh pr create' (runs pr gate from the effective cwd, blocks on exit 1/2) and 'git worktree add' (runs worktree gate, blocks on exit 1/2), plus a SessionStart hook (startup|clear) that runs the session gate and injects its report as context, never blocking.

Trivial exemption (constitution Dev Workflow): surgical — two new files (scripts/hooks/merge-drift-hook.mjs wrapper + .claude/settings.json) consuming the documented exit codes of scripts/check-merge-drift.mjs (specs/051-merge-drift-gates/contracts/gate-cli.md). Diagnosis: spec 051 deliberately scoped v1 enforcement to CLAUDE.md convention ('no hook wiring in scope for v1, though the gates' exit codes are designed so hooks could consume them later' — spec.md Assumptions); the gap is that nothing physically blocks a session that skips the gate. This task closes it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 PreToolUse hook blocks 'gh pr create' when the pr gate exits nonzero, and allows it on exit 0, verified by piping synthetic hook-input JSON
- [x] #2 PreToolUse hook blocks 'git worktree add' when the worktree gate exits nonzero, allows on exit 0, same verification
- [x] #3 Hook resolves the effective directory (input cwd, honoring a 'cd <path> &&' prefix in the command) and no-ops for commands outside this repo or not matching the two patterns
- [x] #4 SessionStart hook (matcher startup|clear) runs the session gate at repo root and emits the report to stdout for context injection; it never blocks session start
- [x] #5 Hook wrapper is zero-dependency single-file ESM per repo script conventions; missing gate script or malformed stdin fails open (exit 0)
- [x] #6 settings.json validated with jq against the hooks schema; hook fires proven via sentinel test
- [x] #7 CLAUDE.md merge-drift section notes that the gates are hook-enforced
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation tier: Sonnet (default) per Principle V rubric — single-file wrapper script + settings JSON, no concurrency/governor logic. Delegated to spec-implementer in .worktrees/task-132.

Implementation complete: PR #86 open (branch task-132-merge-drift-hooks, commit 26fc753, base 7405d23). All 7 ACs verified via synthetic hook-input JSON (block on gate exit 1/2, allow on 0, cd-prefix redirect, fail-open paths, session report, jq schema). Gated by planning tier: diff vs true base = exactly 3 intended files; pr gate dogfooded → exit 0/warnings (stale-base 4, merge clean). Post-merge note: hooks go live for NEW sessions; current sessions need /hooks or restart to load .claude/settings.json.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merge-drift gates are now harness-enforced: PR #86 merged as 30a359c. scripts/hooks/merge-drift-hook.mjs + checked-in .claude/settings.json wire PreToolUse (blocks PR-create and worktree-add commands on gate exit 1/2, fail-open otherwise, cd-prefix-aware, jurisdiction-checked) and SessionStart (session gate report injected as context, never blocks). All 7 ACs verified with synthetic hook input. Went live immediately in the implementing session and produced its first true block within minutes. Worktree and branch cleaned via the session gate's own prescription. Known limitation (first live firing): the command matcher is naive to shell quoting — a quoted string argument containing the literal trigger phrase fires the gate (observed: a board-note text mentioning the command pattern). Fail-safe direction (blocks, never bypasses); candidate follow-up: skip matches inside quoted segments.
<!-- SECTION:FINAL_SUMMARY:END -->
