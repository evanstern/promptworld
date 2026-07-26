---
id: TASK-156
title: >-
  Wiki body-budget debt: four notes over 8000 chars under the 0.21.0-tightened
  gate
status: To Do
assignee: []
created_date: '2026-07-26 20:13'
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
