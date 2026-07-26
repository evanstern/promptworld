---
id: TASK-156
title: >-
  Wiki body-budget debt: four notes over 8000 chars under the 0.21.0-tightened
  gate
status: In Progress
assignee: []
created_date: '2026-07-26 20:13'
updated_date: '2026-07-26 21:40'
labels:
  - wiki
  - tooling
dependencies: []
ordinal: 125000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Surfaced by TASK-67's lane (spec 076 close-out): executor-world-state.md (+414), gru.md (+67), tui-dock-tabs.md (+1667), tui-map-view.md (+882) exceed the corpus body budget as merged by PR #115 (look-cursor grounding) under the tightened 0.21.0 freshness gate. Fix = split/trim per the corpus-spec child-note pattern (tui-dock-tabs and tui-map-view are the heavy ones). Pre-existing main debt — blocks the corpus freshness gate's budget check, not the merge-drift pr gate.
<!-- SECTION:DESCRIPTION:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Operator directive 2026-07-26: this work moves into a worktree on a task branch — wiki note splits/trims are grounding content and ride a PR per spec 069 (never root commits); root stays clean on main. The in-root WIP observed during the reorient sweep should be migrated to .worktrees/task-156 before continuing.

2026-07-26: orphaned in-root WIP (12 modified notes + 4 new child notes: scenario-rubric, sim-state-outcome-fields, tui-dock-tile-view, tui-look-cursor) migrated off root onto this branch per the operator directive on this card and the TASK-160 root-read-only rule. Originating session lost; salvaged via git stash from root. Doneness not yet verified — body-budget gate + INDEX consistency + player-docs regen still to be checked before PR.

2026-07-26: salvage complete — WIP committed on task-156-wiki-body-budget, CAPSULES.md regenerated, getting-started.html re-pinned (gru.md moved, facts unchanged). All gates green: corpus 167 fresh / zero budget violations, player-docs 13 fresh, pr gate pass. PR #121 open; merge awaits operator review of the (unreviewed, salvaged) wiki content.
<!-- SECTION:NOTES:END -->
