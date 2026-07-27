---
id: TASK-118
title: Faith-driven charge regeneration (endogenous mana) — spec-first
status: Done
assignee: []
created_date: '2026-07-25 04:43'
updated_date: '2026-07-27 01:31'
labels:
  - learning-game
  - metatron
  - design-session
dependencies: []
priority: medium
ordinal: 89000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Learning-game synthesis Wave 3 (operator decision 4, ratified 2026-07-25). Metatron's charge regen becomes a pure reducer function of village faith state, closing the god-game positive feedback loop the corpus documents (power derived from the population of worshipers): better prompting -> truer prophecies -> more faith -> more power. This is the ambient endgame's unscored score. Spec must define: faith as event-sourced state (belief provenance from spec 030 already distinguishes omen-origin beliefs), the prophecy-verification rule (what makes a vision 'true'), and regen as a pure function (replay-safe, same shape as today's clock regen). Known tradeoff to design around: the failure spiral (low faith -> fewer charges -> less ability to rebuild) — genre-authentic and roguelike-appropriate in scenarios, but the ambient world may want a floor. Touches reducer doctrine: constitution spec rigor applies (Spec Kit + spec-bridge:link before implementation). Grounding: Analysis-Learning-Game-Fit rec 3.

Spec: specs/085-faith-regen
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Spec Kit spec produced and linked via spec-bridge before implementation
- [x] #2 Faith is event-sourced state with a defined prophecy-verification rule
- [x] #3 Regen is a pure reducer function of faith; replay determinism demonstrated
- [x] #4 Failure-spiral posture decided explicitly (scenario vs ambient floor)
- [x] #5 Spec phase: Setup
- [x] #6 Spec phase: Foundational — faith state, fold, curve (blocks all user stories)
- [x] #7 Spec phase: US1 — the faith accounting sweep (P1) 🎯 MVP
- [x] #8 Spec phase: US2 — regen as a pure function of faith + the posture decision (P1)
- [x] #9 Spec phase: US3 — prophecy: declare, verify, judge (P2)
- [x] #10 Spec phase: US4 — the visible surface (P2)
- [x] #11 Spec phase: Polish, grounding, gates
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reorient 2026-07-26 decision 9: lane order — after TASK-67 (duel first: completes the learn-iterate loop), before TASK-112 (agentization changes what earns faith; TASK-112 AC5/AC6 already legislate the tutor-lane exclusion). Strip integration pre-specified at panels/guardian-strip.md §4 (dashed faith segment contract). Corpus riders: failure-spiral AC grounded in the Hades God-Mode reasoning (Meta-Progression-and-Failure); overjustification caution — faith stays an in-fiction resource, never a badge/streak surface.

Realigned 2026-07-26 (guardian-directives ideation, operator-endorsed): FULFILLED DIRECTIVES are the natural endogenous faith source — villager compliance with the guardian's directives closes the god-game mana loop (prosperity of the flock funds the power that shapes it; research/Game-Gameplay-Patterns/Indirect-Control-and-Divine-Intervention.md). Order after TASK-157 (guardian directives/designations); the spec-first pass here should define faith earned on directive fulfillment events (directive.* vocabulary lands in TASK-157) alongside any other faith sources it identifies.

Rider from TASK-151 close-out (spec 077 FR-020): when faith-driven charge regen ships, add the 'first-faith-event' lesson to the tranche — deliberately NOT stubbed in spec 077 because no faith event type exists yet; the lesson catalog taxonomy test pins its absence until then.

Sweep claim (runbook docs/design/faith-directives-sweep-runbook.md, signed-off 2026-07-26): spec 085-faith-regen. Tier: Opus 4.8 — reducer doctrine (faith as event-sourced state, regen as pure function), doctrine-adjacent by definition. Dependencies satisfied: TASK-67 (duel) and TASK-157 (directives — directive.fulfilled is the named faith seam, specs/084-guardian-directives/contracts/events.md §3) both merged. Board claim at root per TASK-161; spec stub rides the branch.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #126. Endogenous mana live: faith as event-sourced state (closed five-reason delta table, clamping reducer arm the only writer, no retroactive minting); regen a pure function of faith with the genesis band byte-identical to the old 6h schedule (pinned); prophecy via a charge-priced uncancellable tool over a closed predicate vocabulary judged from recorded state (charge spend event-sourced in the arm); FR-005 posture decided — scenarios spiral (forsaken), ambient floors at 24h, reversal lever recorded; strip §4 faith segment ships per the pre-specified contract; first-faith-event lesson closes the spec-077 rider; faith.* rubric-banned (tutor lane never earns faith). Delta magnitudes + 1-charge pricing + no-cancel recorded as review assumptions. Tier: Opus 4.8 as recorded.
<!-- SECTION:FINAL_SUMMARY:END -->
