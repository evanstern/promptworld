---
id: TASK-130
title: >-
  player-docs: refresh keys-reference.html after spec 047 keymap.md
  restructuring
status: To Do
assignee: []
created_date: '2026-07-25 16:04'
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
