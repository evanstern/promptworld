---
id: TASK-54
title: >-
  resolveWorld swallows manifest-open errors — version mismatch reports as 'no
  world named'
status: In Progress
assignee: []
created_date: '2026-07-22 04:54'
updated_date: '2026-07-24 15:47'
labels:
  - bug
dependencies: []
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Found live (2026-07-22, post-PR-#31): a stale v1 binary running 'scriptworld ui myworld-01' against the migrated v2 world reported "no world named myworld-01" while ps (IPC-based, no manifest read) listed it fine. Root cause: name resolution treats any world.Open failure as not-a-world and falls through to the not-found message, hiding the real 'format_version N unsupported (run scriptworld migrate)' error. Fix: resolution should distinguish 'directory exists but unopenable' from 'no such world' and surface the Open error verbatim. Applies both directions (old binary/new world, new binary/future-format world). Surgical: internal/worlds resolution path + a test.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Resolution distinguishes 'directory exists but unopenable' from 'no such world' and surfaces the world.Open error verbatim (both directions: old binary/new world, new binary/future world)
- [x] #2 Test covering a format_version-mismatch world resolving to the migrate-hint error, not not-found
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Trivial exemption (surgical fix + file:line diagnosis on task + ACs on task). 1) In .worktrees/task-54: change internal/worlds/resolve.go so candidate probing distinguishes 'not a world' from 'world dir exists but world.Open fails' (e.g. format_version mismatch) and surfaces the Open error verbatim, home-first. 2) Test: format_version-mismatch world resolves to the migrate-hint error, not ErrNotFound. Tier: Sonnet via spec-implementer — single-package (internal/worlds) routine fix, no concurrency/doctrine surface, no prior failed attempt.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Drift audit 2026-07-23: still real. Fresh pins — isReadableWorld returns world.Open(dir)==nil at internal/worlds/resolve.go:106-109, so a format-version-mismatch world is indistinguishable from a missing one; falls through to ErrNotFound ('no world named %q', resolve.go:44-45, 82-103). CLI path via cmd/promptworld/commands.go:68-73.

Tier choice: Sonnet (default) per constitution Principle V rubric — single-package surgical fix in internal/worlds + test; not cross-package, not concurrency/governor, not doctrine-adjacent.

Implemented on branch task-54-resolve-surface-open-errors (worktree .worktrees/task-54), PR #66 open. resolve.go: isReadableWorld → probeWorld (ok/unopenable/not-a-world; unopenable = manifest exists on disk but world.Open failed, version-agnostic so it covers both binary/world direction mismatches). New ErrUnopenable{Name,Path,Err} with Unwrap surfaces the world.Open migrate-hint error verbatim. Home-first hardened: unopenable home candidate beats healthy registry candidate. 3 new tests in resolve_test.go; go build/vet/test green across internal/worlds, internal/world, cmd. Gated by planning session: full diff reviewed, tests re-run independently. Awaiting merge; wiki-update after merge (resolve.go is a pinned source).
<!-- SECTION:NOTES:END -->
