---
id: TASK-83
title: 'gofmt drift on main: five files fail gofmt -l'
status: Done
assignee: []
created_date: '2026-07-23 21:19'
updated_date: '2026-07-24 15:17'
labels: []
dependencies: []
ordinal: 3000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Found by spec 028's T017 gate (2026-07-23): gofmt -l internal/ flags five files whose drift PRE-EXISTS task-33 (not in its diff): internal/metatron/digest.go, internal/sim/bulk_cap_test.go, internal/sim/ground_pile_test.go, internal/tui/grammar.go, internal/tui/render_test.go. Surgical fix: gofmt -w those files. Consider adding a gofmt check to the standing test/CI gate so drift can't accumulate silently. Trivial-exemption candidate (surgical, diagnosis pinned, ACs here).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 gofmt -l internal/ prints nothing on main
- [x] #2 A standing gate (test or hook) fails on future gofmt drift, or a decision not to add one is recorded
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Worktree task-83 off origin/main. 2. gofmt -w currently drifted files (drift moved since filing: now bulk_cap_test.go, ground_pile_test.go, recipes_test.go, grammar.go). 3. Add standing gate: a Go test (gofmt_test.go) that fails go test ./... when any tracked .go file is unformatted — no CI workflows exist, so the test suite is the standing gate. 4. go test ./... green, PR, tick ACs. Tier: Sonnet spec-implementer (routine single-purpose slice per Principle V rubric; trivial exemption applies — surgical fix, diagnosis pinned on task, ACs on task).
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-07-24: drift re-verified on fresh main: gofmt -l internal/ now flags internal/sim/bulk_cap_test.go, internal/sim/ground_pile_test.go, internal/sim/recipes_test.go, internal/tui/grammar.go (digest.go + render_test.go fixed by later merges; recipes_test.go newly drifted). Confirms AC#2's premise that drift accumulates silently.

2026-07-24: implemented on branch task-83-gofmt-drift, Sonnet spec-implementer, single commit 3e4e9df; PR #64 (https://github.com/evanstern/promptworld/pull/64) merged to main (af13190). AC#1 proven on main post-merge: gofmt -l internal/ prints nothing and internal/lint TestGofmt passes on main. AC#2: standing gate internal/lint/gofmt_test.go TestGofmt — walks all .go files from module root, diffs against go/format.Source, fails naming every offender; rides go test ./... (project's standing gate; no CI/Makefile). Deliberate-failure sanity check passed during implementation. (Note: an earlier copy of this note briefly landed on the worktree's copy of this file due to a stray cwd; discarded there and re-recorded here from the root.)
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Fixed pre-existing gofmt drift (4 files at fix time: internal/sim/bulk_cap_test.go, ground_pile_test.go, recipes_test.go, internal/tui/grammar.go) and added a standing gate: internal/lint/gofmt_test.go TestGofmt diffs every .go file against go/format.Source and fails naming offenders, riding go test ./... . Landed via PR #64 (merge af13190). Verified on main: gofmt -l internal/ empty, TestGofmt green.
<!-- SECTION:FINAL_SUMMARY:END -->
