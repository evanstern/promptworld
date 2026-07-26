---
id: TASK-97
title: >-
  Bundle effects: target addressing for structures/piles/terrain
  (move_entity/remove_entity beyond villagers)
status: In Progress
assignee: []
created_date: '2026-07-24 19:39'
updated_date: '2026-07-26 21:43'
labels:
  - idea
dependencies: []
priority: high
ordinal: 7000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Follow-up from TASK-85 (specs/036-scriptable-agent-tools). The v1 effect compiler addresses move_entity/grant_item targets by living-villager name only; remove_entity is structurally present but inert (reducer rejects villager removal by doctrine). The built-in work_miracle addresses structures/piles/terrain by class+tile (MiracleParams{Class,X,Y}); bundle effects have no equivalent grammar — no contract specifies one and no fixture exercises it. Design a target-addressing grammar (class+tile or an id scheme) for bundle effects, extend contracts/bundle-manifest.md + the compiler, and add fixtures. Anchor: internal/bundle/effects.go (resolve* helpers), internal/metatron/miracle_batch.go villager-vs-structure paths.

Spec: specs/082-target-addressing
<!-- SECTION:DESCRIPTION:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Realigned 2026-07-26 (guardian-directives ideation): promoted to HIGH — now on the critical path of TASK-157 (guardian directives/designations), which depends on this task. The target-addressing grammar designed here (class+tile / region / id scheme) must serve BOTH bundle effects AND designation tile/region addressing — design it once for both consumers. Anchor list unchanged; add internal/tool registry designation params as a third consumer when TASK-157's spec lands.

Sweep claim (runbook docs/design/faith-directives-sweep-runbook.md, signed-off 2026-07-26): spec 082-target-addressing. Tier: Opus 4.8 — cross-package (bundle effect compiler + tool registry + metatron miracle paths); the grammar binds two consumers (bundle effects + TASK-157 designations), designed once. TASK-160 claim flow: claim authored on the task branch, landed on main by merge.
<!-- SECTION:NOTES:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Spec phase: Setup
- [ ] #2 Spec phase: Foundational — the grammar package (blocks all user stories)
- [ ] #3 Spec phase: US1 — structures & piles by class+tile; remove_entity becomes real (P1) 🎯 MVP
- [ ] #4 Spec phase: US2 — terrain removal (P2)
- [ ] #5 Spec phase: US3 — the designation seam, contract-named (P3)
- [ ] #6 Spec phase: Grounding + gates (in-branch, wiki-in-PR lifecycle)
- [ ] #7 Spec phase: Post-merge bookkeeping (derived state, lands by merge — TASK-160)
<!-- AC:END -->
