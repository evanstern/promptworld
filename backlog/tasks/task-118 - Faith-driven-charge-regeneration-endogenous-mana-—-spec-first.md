---
id: TASK-118
title: Faith-driven charge regeneration (endogenous mana) — spec-first
status: To Do
assignee: []
created_date: '2026-07-25 04:43'
updated_date: '2026-07-26 17:58'
labels:
  - learning-game
  - metatron
  - design-session
dependencies: []
ordinal: 89000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Learning-game synthesis Wave 3 (operator decision 4, ratified 2026-07-25). Metatron's charge regen becomes a pure reducer function of village faith state, closing the god-game positive feedback loop the corpus documents (power derived from the population of worshipers): better prompting -> truer prophecies -> more faith -> more power. This is the ambient endgame's unscored score. Spec must define: faith as event-sourced state (belief provenance from spec 030 already distinguishes omen-origin beliefs), the prophecy-verification rule (what makes a vision 'true'), and regen as a pure function (replay-safe, same shape as today's clock regen). Known tradeoff to design around: the failure spiral (low faith -> fewer charges -> less ability to rebuild) — genre-authentic and roguelike-appropriate in scenarios, but the ambient world may want a floor. Touches reducer doctrine: constitution spec rigor applies (Spec Kit + spec-bridge:link before implementation). Grounding: Analysis-Learning-Game-Fit rec 3.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Spec Kit spec produced and linked via spec-bridge before implementation
- [ ] #2 Faith is event-sourced state with a defined prophecy-verification rule
- [ ] #3 Regen is a pure reducer function of faith; replay determinism demonstrated
- [ ] #4 Failure-spiral posture decided explicitly (scenario vs ambient floor)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reorient 2026-07-26 decision 9: lane order — after TASK-67 (duel first: completes the learn-iterate loop), before TASK-112 (agentization changes what earns faith; TASK-112 AC5/AC6 already legislate the tutor-lane exclusion). Strip integration pre-specified at panels/guardian-strip.md §4 (dashed faith segment contract). Corpus riders: failure-spiral AC grounded in the Hades God-Mode reasoning (Meta-Progression-and-Failure); overjustification caution — faith stays an in-fiction resource, never a badge/streak surface.
<!-- SECTION:NOTES:END -->
