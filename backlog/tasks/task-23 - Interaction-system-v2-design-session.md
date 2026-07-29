---
id: TASK-23
title: 'Interaction system v2: design session'
status: In Progress
assignee: []
created_date: '2026-07-19 22:27'
updated_date: '2026-07-29 19:15'
labels:
  - design
dependencies: []
ordinal: 9000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The full agent<=>agent interaction system needs a ground-up design (user, 2026-07-19: 'that's a whole system we need to design'). Socratic/spec session covering: interaction primitives beyond talk (argue, trade, teach, comfort, conspire), scene formation and dissolution for groups, how conversation records become long-term relationship memory (interplay with TASK-9 consolidation), LLM budget shaping across interaction kinds, and what the chronicle (TASK-11) needs from interactions. Output: a spec under specs/ linked to the board via spec-bridge. Builds on evidence from Conversations v1.5.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A grounding/design session produces a spec directory for interactions v2, linked on the board via spec-bridge
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Re-grounding 2026-07-22: reframed onto the tool substrate — new interaction primitives (argue/trade/teach/comfort/conspire) should be authored as tool-registry entries (TASK-53) with per-agent rosters, invoked through the TASK-52 loop, not as a bespoke parallel system. The design session should start from the registry's tool classes (world/expressive/read). Ordered after Metatron v2 (TASK-27), which exercises the same substrate first.

Reorient 2026-07-26 (board move 12): reframed as the DF-pole drama generator — social incidents are the ambient endgame's retention content (Thornspire cascade proved the substrate); chronicle requirements should include rubric-legibility for future social exercises.

Reordered 2026-07-26 (guardian-directives ideation, operator: 'ok on 23'): the TASK-27 ordering note is obsolete (Metatron v2 is Done). New ordering: AFTER TASK-157 (guardian directives/designations) — the guardian→villager directive channel (durable checkable goals, decision-context block, DIRECTIVE reflex rung) will inform the villager↔villager interaction primitives this design session covers; teach/order-shaped interactions should reuse the directive substrate's completion-predicate vocabulary where applicable.
<!-- SECTION:NOTES:END -->
