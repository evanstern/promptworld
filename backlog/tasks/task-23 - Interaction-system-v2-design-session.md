---
id: TASK-23
title: 'Interaction system v2: design session'
status: Done
assignee: []
created_date: '2026-07-19 22:27'
updated_date: '2026-07-30 00:40'
labels:
  - design
dependencies: []
ordinal: 9000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The full agent<=>agent interaction system needs a ground-up design (user, 2026-07-19: 'that's a whole system we need to design'). Socratic/spec session covering: interaction primitives beyond talk (argue, trade, teach, comfort, conspire), scene formation and dissolution for groups, how conversation records become long-term relationship memory (interplay with TASK-9 consolidation), LLM budget shaping across interaction kinds, and what the chronicle (TASK-11) needs from interactions. Output: a spec under specs/ linked to the board via spec-bridge. Builds on evidence from Conversations v1.5.

Spec: specs/093-interactions-v2
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A grounding/design session produces a spec directory for interactions v2, linked on the board via spec-bridge
- [x] #2 Spec phase: Design authoring
- [x] #3 Spec phase: Ratification
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Re-grounding 2026-07-22: reframed onto the tool substrate — new interaction primitives (argue/trade/teach/comfort/conspire) should be authored as tool-registry entries (TASK-53) with per-agent rosters, invoked through the TASK-52 loop, not as a bespoke parallel system. The design session should start from the registry's tool classes (world/expressive/read). Ordered after Metatron v2 (TASK-27), which exercises the same substrate first.

Reorient 2026-07-26 (board move 12): reframed as the DF-pole drama generator — social incidents are the ambient endgame's retention content (Thornspire cascade proved the substrate); chronicle requirements should include rubric-legibility for future social exercises.

Reordered 2026-07-26 (guardian-directives ideation, operator: 'ok on 23'): the TASK-27 ordering note is obsolete (Metatron v2 is Done). New ordering: AFTER TASK-157 (guardian directives/designations) — the guardian→villager directive channel (durable checkable goals, decision-context block, DIRECTIVE reflex rung) will inform the villager↔villager interaction primitives this design session covers; teach/order-shaped interactions should reuse the directive substrate's completion-predicate vocabulary where applicable.

board-sweep-2026-07-29 lane 5: design session run autonomously per sign-off; spec 093 authored on the tool substrate per the card's re-grounding notes; OQ-1..OQ-5 flagged for the ratification PR review. No implementer dispatch — planning-model authoring.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Design ratified: PR #132 merged by operator (714f8a08e). Spec 093 delivers interactions v2 on the tool substrate — five primitives (argue/trade/teach/comfort/conspire) as registry entries, scenes generalizing hails, typed relationship deltas into consolidation, cognition-registry budgeting, rubric-legible chronicle entries; OQ-1..OQ-5 recorded (operator ratified at review; unresolved OQs carry to implementation slicing). Implementation is future work.
<!-- SECTION:FINAL_SUMMARY:END -->
