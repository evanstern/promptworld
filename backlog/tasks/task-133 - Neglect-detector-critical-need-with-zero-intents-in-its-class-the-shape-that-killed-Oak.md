---
id: TASK-133
title: >-
  Neglect detector: critical need with zero intents in its class (the shape that
  killed Oak)
status: To Do
assignee: []
created_date: '2026-07-25 18:57'
updated_date: '2026-07-26 17:58'
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
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Detector defined and validated against Oak's death window in the world-01 v3 log (fires) and healthy windows (silent)
- [ ] #2 Deterministic percept event + high-salience observation memory injection, replay-visible
- [ ] #3 Composition with survival watches (TASK-111) considered in the spec
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reorient 2026-07-26 (board move 13): relabeled a learning-game prerequisite — neglect is the observed ambient failure shape (Oak's warmth slide), and postmortem attribution can't teach unless the sim can name neglect when it happens. Its alert enters through the shipped severity grammar (chronicle whole-line alert + map overlay, patterns/chronicle-grammar.md), never a new channel.
<!-- SECTION:NOTES:END -->
