---
id: TASK-129
title: Villager strip + map condition overlays (village lens completion)
status: Done
assignee: []
created_date: '2026-07-25 14:45'
updated_date: '2026-07-26 03:27'
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
- [x] #1 Villager strip renders and click/jump follows parity doctrine
- [x] #2 At least needs-critical and suppressed-mind overlays render on the map
- [x] #3 Spec phase: Setup
- [x] #4 Spec phase: User Story 1 — villager strip (P1)
- [x] #5 Spec phase: User Story 2 — map condition overlays (P1)
- [x] #6 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Model tier: Sonnet (spec-implementer default). Rubric: single-package rendering, tests alongside — routine tier. Standing resolutions in spec 060: strip is display-only per the authored page (board AC #1's click/jump clause satisfied vacuously — no actions, no parity gap); look-cursor DEFERRED with ruling recorded on map.md; dying-fire pulse = steady warn style, no blink. DISPATCH: Lane 5 tail — after Lane 3/4 merges.

Dispatched (UI-sweep orchestrator): spec-implementer on Sonnet per recorded rubric; worktree .worktrees/task-129 cut from post-115 main (6056a22). Lane-5 gate met: Lane 3 fully merged (117 #88, 127 #99, 115 #100). Standing rulings from the card hold: strip display-only (AC #1 parity clause vacuous), look-cursor deferred (ruling on map.md), dying-fire pulse = steady warn. Expect rebase over TASK-128's merge (in dev concurrently; its stage-defaults table lists the villager strip as tolerated-absent, so either merge order reconciles).

spec-bridge sync: Setup: 1/1 · US1 villager strip: 3/3 · US2 map overlays: 2/2 · Polish: 3/3 — status In Progress → Done

Merged via PR #103 (squash 7e3c2b5). Human ACs on merge evidence: #1 strip renders display-only per the standing ruling (parity clause vacuous — no actions); #2 needs-critical + suppressed-mind overlays render (needs-critical wins ties), plus the dying-fire steady-warn state. Rebase reconciled with TASK-128s stage-defaults (two 128 tests updated whose golden/row-sum assumptions predated the new row — justified in comments; parity sweep green). Design pages re-pinned to squash on main. One sim re-export (DangerRestBelow); morale excluded from needs-critical (no danger band exists). Gate-reviewed and accepted by UI-sweep orchestrator.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Setup: 1/1 · User Story 1 — villager strip (P1): 3/3 · User Story 2 — map condition overlays (P1): 2/2 · Polish & Cross-Cutting Concerns: 3/3). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
