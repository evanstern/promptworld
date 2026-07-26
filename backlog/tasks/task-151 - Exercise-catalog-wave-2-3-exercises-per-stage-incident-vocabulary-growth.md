---
id: TASK-151
title: 'Exercise catalog wave: 2-3 exercises per stage + incident vocabulary growth'
status: Done
assignee: []
created_date: '2026-07-26 17:57'
updated_date: '2026-07-26 20:36'
labels:
  - gameplay
  - content
dependencies:
  - TASK-149
references:
  - docs/design/reorient-2026-07-26-ui.md
ordinal: 121000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient 2026-07-26 decision 5. Ladder v1 = 2-3 hand-authored exercises per stage (today: two exercises, one production rubric). Grow the incident vocabulary beyond gru_emerges to ~3 kinds (cold snap, forage blight, stranger/trickster arrival) — each a reducer-valid event shape indistinguishable from an ambient cause (the gru.emerged precedent), entering through the shipped severity grammar (no new channels). Lesson catalog tranche 2 (first explain answer, first report card, first skill file, first faith event post-TASK-118) plus the first wrong-thing detector lesson (repeated same-cause tool rejections) ride this wave as content. Depends on TASK-149 for rubric truth.

Spec: specs/077-exercise-catalog
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Each stage has 2-3 exercises with production evaluators
- [x] #2 At least 3 incident kinds, replay-safe and ambient-indistinguishable, entering via shipped severity channels
- [x] #3 Lesson tranche 2 + one wrong-thing-detector lesson land as catalog content
- [x] #4 Spec phase: Setup
- [x] #5 Spec phase: Foundational — state, reducer arms, evidence coordinates (blocks all later phases)
- [x] #6 Spec phase: User Story 2 — incident vocabulary grows to three new kinds (P1, board AC #2)
- [x] #7 Spec phase: User Story 1a — emitter generalization (blocks 4b; board AC #1)
- [x] #8 Spec phase: User Story 1b — nine exercises with production evaluators (P1, board AC #1)
- [x] #9 Spec phase: User Story 3 — lesson tranche 2 + wrong-thing detector (P2, board AC #3)
- [x] #10 Spec phase: Design authority — pages amended, gate green
- [x] #11 Spec phase: Grounding — wiki-in-PR obligations (in-branch, pr-gate enforced)
- [x] #12 Spec phase: Polish & close-out
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Sweep claim (runbook docs/design/reorient-2026-07-26-sweep-runbook.md): spec 077-exercise-catalog. Tier: Opus 4.8 — new reducer-valid event kinds across sim reducer + chronicle digest (TestCatalogSweep) + scenario evaluators; replay-safety doctrine-adjacent. Dependency satisfied: TASK-149 merged (rubric truth + the-law evaluator live).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #119 (merge commit fb8c865). Nine exercises shipped (3/2/2/2 across stages, seeds 46101–46109), every one with a production evaluator and pass emission — the-law's emission completes spec 072 FR-009 via persisted CharterObservedSeq/Tick; metatron.skills_observed provides the first production stage-3→4 unlock. Three incident kinds (cold snap, forage blight, stranger entity) enter as reducer-valid ambient-indistinguishable events through shipped channels; replay/snapshot/rebase proofs green. Lesson catalog 8→12 incl. the same-refusal wrong-thing detector; first-faith-event honestly absent pending TASK-118 (rider noted). Tier: Opus 4.8 as recorded.
<!-- SECTION:FINAL_SUMMARY:END -->
