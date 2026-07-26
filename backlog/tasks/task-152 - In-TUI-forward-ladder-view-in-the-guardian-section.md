---
id: TASK-152
title: In-TUI forward-ladder view in the ? guardian section
status: Done
assignee: []
created_date: '2026-07-26 17:57'
updated_date: '2026-07-26 20:29'
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

Spec: specs/078-tui-ladder-view
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 ? guardian section renders the four-stage ladder with earned/next state and unlock evidence, parity with stages --json
- [x] #2 overlays/help.md byte-identity table row added same-PR; check-tui-design.mjs --changed passes
- [x] #3 Spec phase: Setup
- [x] #4 Spec phase: Foundational — the shared earned rule (blocks US1)
- [x] #5 Spec phase: User Story 1 — the ladder block (P1, board AC #1) 🎯 MVP
- [x] #6 Spec phase: User Story 2 — design authority, same PR (P2, board AC #2)
- [x] #7 Spec phase: Grounding (in-branch, per the wiki-in-PR lifecycle)
- [x] #8 Spec phase: Polish & gates
- [x] #9 Spec phase: Post-merge bookkeeping (root, derived state only)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Sweep claim (runbook docs/design/reorient-2026-07-26-sweep-runbook.md): spec 078-tui-ladder-view. Tier: Sonnet — single-package deterministic TUI view (world.StagesLadder + worlds.LoadUnlocks, relocated by spec 063 T014 for exactly this). Sequenced after TASK-142's merge (shared help.go/help.md).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #118 (merge commit fb8c865). The ? guardian section renders the four-stage forward ladder (identity · concept · earned/next · unlock evidence) with stages --json parity BY CONSTRUCTION: the earned rule relocated to (*worlds.Unlocks).StageEarned, both surfaces are projections; parity test computes expectations at runtime (zero hardcoded catalog — survives TASK-151). StageOverridden annotates without laundering into earned. help.md byte-identity row + control row same-PR; 24 wiki notes re-pinned (7 amended); player docs regenerated. Tier: Sonnet as recorded.
<!-- SECTION:FINAL_SUMMARY:END -->
