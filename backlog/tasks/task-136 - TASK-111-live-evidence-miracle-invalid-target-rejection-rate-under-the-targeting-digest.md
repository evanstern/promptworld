---
id: TASK-136
title: >-
  TASK-111 live evidence: miracle invalid-target rejection rate under the
  targeting digest
status: In Progress
assignee: []
created_date: '2026-07-25 19:37'
updated_date: '2026-07-27 08:27'
labels:
  - mvls
  - guardian-survival
dependencies: []
priority: medium
ordinal: 106000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Follow-up carded from the MVLS sweep (operator request 2026-07-25) so TASK-111's AC3 second clause doesn't rot: 'invalid-target rejections drop to ~0' needs live-world observation the merged code cannot prove. Mechanism shipped in spec 059 / PR #90 (token-bounded positions+passability digest in miracle-capable prompts; door round-trip regression green). Work: measure the door-rejection rate on angel miracle coordinates on a post-059 run vs world-01's pre-059 baseline (3 of 4 rejected); evidence rides the TASK-137 crisis experiment. Tick TASK-111 AC#3 with the evidence.

REORIENT 2026-07-26 reframe (docs/design/reorient-2026-07-26-ui.md): under the teaching-game lens this measurement is the noise-floor check for prompt attribution. Invalid-target rejections are mechanical noise sitting between the player's charter wording and the guardian's outcomes — if they aren't ~0, the charter-delta experiment (TASK-137) and the fork duel's scoreboard (TASK-67) attribute targeting failures to prompting skill: a false-✗ twin of the report-card-truth problem (synthesis decision 1's honesty doctrine — feedback only teaches if the evidence is true). The measurement plan is unchanged; the number it produces is now the confound bound that prompt-vs-outcome comparisons must quote, not just TASK-111 AC3 close-out.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A post-059 run with >=5 angel miracle attempts measured; invalid-coordinate door rejections at or near 0, evidence recorded here and on TASK-111
- [ ] #2 Rejection rate quoted as the mechanical-noise/confound bound in TASK-137's evidence doc, so prompt-attribution comparisons (TASK-137, TASK-67 duel) can cite it
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Evidence rides the TASK-137 crisis experiment (2026-07-26): harsh-dial worlds force survival turns -> miracle attempts; door rejection rate measured across both arms vs the world-01 pre-059 baseline (3 of 4 rejected).

TASK-137 run produced the rejection-rate sample (2026-07-27, docs/design/evidence/task-137/results.md): 10 privileged-action attempts across both arms, 8 rejected by the targeting/validity door (80%; 3/3 default arm, 5/7 authored arm). Taxonomy: send_vision place grants missing place_x/place_y (2), hallucinated tree coords (1), work_miracle give_item with non-item kind 'fire'/empty (3), work_miracle move with empty entity class (2). All model-side (gemma4:12b-mlx native tool-calls). Zero invalid-target rejections converted to landed actions after retry within the same turn except Sage's vision (retried with coords — but hallucinated, re-rejected). Event evidence in ~/.promptworld/measure/task-137-{default,authored}/world.db (cog.tool_call verdict/reason fields).
<!-- SECTION:NOTES:END -->
