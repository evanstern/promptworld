---
id: TASK-133
title: >-
  Neglect detector: critical need with zero intents in its class (the shape that
  killed Oak)
status: Done
assignee: []
created_date: '2026-07-25 18:57'
updated_date: '2026-07-27 00:51'
labels:
  - mvls
  - thrash-detection
dependencies: []
priority: high
ordinal: 103000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Follow-on carded from TASK-106 research (docs/design/thrash-detection-research.md §1.3, §4; operator decision 2026-07-25). World-01 evidence: Oak died of exposure day 7 while warmth drained 636->0 over ~6h — the reflex chopped wood and the planner wandered; zero warmth-class intents the whole slide. No oscillation detector can catch death-by-neglect. Detector: need below critical threshold for T game-time with zero intents in that need's class over the same window -> emit a deterministic percept event + high-salience observation memory (same injection design as the thrash percept sketch, §3). Simpler than the thrash detector; composes with TASK-111 (a survival watch the angel can act on) and TASK-108/103 (reflex/arbitration backstop). Thresholds promoted-dial-ready consts, not tuning.json entries until earned.

Spec: specs/083-neglect-detector
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Detector defined and validated against Oak's death window in the world-01 v3 log (fires) and healthy windows (silent)
- [x] #2 Deterministic percept event + high-salience observation memory injection, replay-visible
- [x] #3 Composition with survival watches (TASK-111) considered in the spec
- [x] #4 Spec phase: Setup
- [x] #5 Spec phase: Foundational — derived state, class dictionary, reducer arms (blocks all)
- [x] #6 Spec phase: User Story 1 — the deterministic percept + high-salience memory (P1, board AC #2)
- [x] #7 Spec phase: User Story 2 — validated against Oak's death window, silent on healthy (P1, board AC #1)
- [x] #8 Spec phase: User Story 3 — shipped severity channels: chronicle alert + map overlay (P2, reorient move-13 obligation)
- [x] #9 Spec phase: Grounding — wiki-in-PR obligations (in-branch, pr-gate enforced)
- [x] #10 Spec phase: Polish & close-out
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reorient 2026-07-26 (board move 13): relabeled a learning-game prerequisite — neglect is the observed ambient failure shape (Oak's warmth slide), and postmortem attribution can't teach unless the sim can name neglect when it happens. Its alert enters through the shipped severity grammar (chronicle whole-line alert + map overlay, patterns/chronicle-grammar.md), never a new channel.

Sweep claim (runbook docs/design/faith-directives-sweep-runbook.md, signed-off 2026-07-26): spec 083-neglect-detector. Tier: Opus 4.8 — reducer/percept event + high-salience memory injection + world-01 log validation; cognition-adjacent. TASK-160 claim flow: claim authored on the task branch, landed on main by merge.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #125. Neglect detector live: need below the spec-062 danger band for 7200 ticks with zero class intents fires sim.neglect_detected once per episode plus a salience-9 witness memory (the one recorded keep-below-9 exception); alerts ride shipped channels only (chronicle whole-line + existing critical map styling, zero new render code). Validated against the REAL world-01 archived v3 log: Oak's window confirmed, with the empirical finding that Oak slept through it (detector fires at wake ~21min pre-death; awake fixture retains ~5h runway) and the migrated world.db log starts post-death — probe targets world.v3.db. Live-replay hash identity, snapshot byte-identity, six SHIFT rebase anchors. AC3 satisfied by the spec's composition design (named unbuilt fourth watch kind on the wiki seam). Tier: Opus 4.8 as recorded.
<!-- SECTION:FINAL_SUMMARY:END -->
