---
id: TASK-129
title: Villager strip + map condition overlays (village lens completion)
status: In Progress
assignee: []
created_date: '2026-07-25 14:45'
updated_date: '2026-07-26 02:40'
labels:
  - learning-game
  - tui
dependencies:
  - TASK-123
priority: low
ordinal: 99000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient 2026-07-25 (D12), Wave 5. A one-row colonist-bar-style villager strip under the header (status glyphs, needs/mood salience — RimWorld colonist bar + 1.5 mood-glow precedent), stage-defaulted like other chrome; map condition overlays (needs-critical marker, suppressed-mind marker, dying-fire pulse — Cogmind map-dynamics doctrine); evaluate a look-cursor vs the growing legend line.

Spec: specs/060-village-lens
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Villager strip renders and click/jump follows parity doctrine
- [ ] #2 At least needs-critical and suppressed-mind overlays render on the map
- [ ] #3 Spec phase: Setup
- [ ] #4 Spec phase: User Story 1 — villager strip (P1)
- [ ] #5 Spec phase: User Story 2 — map condition overlays (P1)
- [ ] #6 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Model tier: Sonnet (spec-implementer default). Rubric: single-package rendering, tests alongside — routine tier. Standing resolutions in spec 060: strip is display-only per the authored page (board AC #1's click/jump clause satisfied vacuously — no actions, no parity gap); look-cursor DEFERRED with ruling recorded on map.md; dying-fire pulse = steady warn style, no blink. DISPATCH: Lane 5 tail — after Lane 3/4 merges.

Dispatched (UI-sweep orchestrator): spec-implementer on Sonnet per recorded rubric; worktree .worktrees/task-129 cut from post-115 main (6056a22). Lane-5 gate met: Lane 3 fully merged (117 #88, 127 #99, 115 #100). Standing rulings from the card hold: strip display-only (AC #1 parity clause vacuous), look-cursor deferred (ruling on map.md), dying-fire pulse = steady warn. Expect rebase over TASK-128's merge (in dev concurrently; its stage-defaults table lists the villager strip as tolerated-absent, so either merge order reconciles).
<!-- SECTION:NOTES:END -->
