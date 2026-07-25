---
id: TASK-130
title: >-
  player-docs: refresh keys-reference.html after spec 047 keymap.md
  restructuring
status: Done
assignee: []
created_date: '2026-07-25 16:04'
updated_date: '2026-07-25 17:05'
labels:
  - player-docs
dependencies: []
priority: medium
ordinal: 100000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
check-freshness.mjs reports docs/player/keys-reference.html stale: it pins docs/design/tui/patterns/keymap.md@5536b2f, which spec 047 (TASK-123) restructured (help overlay extracted to overlays/help.md; input-parity doctrine added). Run the player-docs skill to regenerate/re-pin the page once TASK-123's PR merges. Found by spec 047 slice C acceptance sweep (quickstart §5).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 check-freshness.mjs --check exits 0: keys-reference.html re-pinned to current keymap.md (beafe69)
- [x] #2 Help-overlay section sourced from docs/design/tui/overlays/help.md with its own source meta tag
- [x] #3 Shipped-only content: no unbuilt keys (G/x/p), no skin tokens; keyboard-only input-parity note added
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Confirm freshness gate shows only keys-reference.html stale (done)
2. Cut worktree .worktrees/task-130 from origin/main
3. Run player-docs skill to regenerate/re-pin keys-reference.html against restructured keymap.md (help overlay → overlays/help.md, input-parity doctrine)
4. Verify check-freshness.mjs --check passes (13 fresh)
5. One PR from task-130 branch
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Regenerated via player-docs skill in .worktrees/task-130. Only stale page rewritten; other 12 pages untouched. PR #81 opened from task-130-player-docs-keys-reference.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
keys-reference.html refreshed after spec 047's keymap.md restructuring, shipped via PR #81 (squash-merged as ef696aa). Re-pinned keymap.md@beafe69; added overlays/help.md@54964eb as second source for the help-overlay section spec 047 extracted there; added the keyboard-only input-parity note. No shipped bindings changed; unbuilt keys (G/x/p) and skin tokens excluded. check-freshness.mjs on merged main: 13 fresh / 0 stale (exit 0). Worktree and branch removed.
<!-- SECTION:FINAL_SUMMARY:END -->
