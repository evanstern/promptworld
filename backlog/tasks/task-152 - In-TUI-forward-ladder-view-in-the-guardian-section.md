---
id: TASK-152
title: In-TUI forward-ladder view in the ? guardian section
status: To Do
assignee: []
created_date: '2026-07-26 17:57'
labels:
  - game-ui
  - pedagogy
dependencies: []
references:
  - docs/design/reorient-2026-07-26-ui.md
ordinal: 122000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient 2026-07-26 decision 6. The forward ladder (all stages: identity · concept · earned/next · unlock evidence, matching stages --json) renders only in the CLI today — a TUI player can see where they are but never what's next. Render a ladder block in the ? guardian section (deterministic floor, model-free), built on world.StagesLadder + worlds.LoadUnlocks (relocated by spec 063 T014 for exactly this). Docs rider: the view is status-derived, so overlays/help.md's byte-identity table gains a row in the same PR.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 ? guardian section renders the four-stage ladder with earned/next state and unlock evidence, parity with stages --json
- [ ] #2 overlays/help.md byte-identity table row added same-PR; check-tui-design.mjs --changed passes
<!-- AC:END -->
