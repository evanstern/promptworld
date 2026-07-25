---
id: TASK-106
title: 'Research: thrash detection as a percept — define and detect goal oscillation'
status: Done
assignee: []
created_date: '2026-07-25 02:42'
updated_date: '2026-07-25 18:57'
labels:
  - goal-quality
  - research
  - thrash-detection
  - mvls
dependencies:
  - TASK-105
priority: medium
ordinal: 3000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Direction D from spike TASK-101 — starts with research and fleshing out per Evan. Question: how do we detect goal thrash in-sim, and what IS a detection? Candidate definition to evaluate: within a sliding window (last N intents or W ticks), the agent alternates between two need-serving goal CLASSES (e.g. food-acquisition {forage,hunt,eat} vs warmth {goto_warmth,build_fire,refuel_fire}) with >=K A→B→A transitions, AND neither underlying need shows net improvement over the window (the need-progress clause separates thrash from healthy interleaving), optionally AND the two goal targets are spatially distinct (shuttling). Detector runs on the per-agent intent history already in the event stream (agent.intent_set carries source/goal/target/tick). On detection: inject a high-salience observation memory ('you have walked between the fire and the berry patch 5 times; neither need improved') rendered into the next prompt, and optionally trigger a planner beat. Research method: replay world-01 (~/.promptworld/worlds/world-01/world.db) as ground truth — tune W/K so Sage/Fern/Oak episodes fire and healthy interleaving does not; also survey alternative metrics (switch-rate, wasted-travel ratio).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Detection definition chosen with false-positive/false-negative analysis against world-01 replay
- [x] #2 Percept/memory injection design sketched (salience, wording, planner-trigger interaction)
- [x] #3 Go/no-go recommendation for implementation task
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
CLARIFICATIONS (Evan, 2026-07-24): (1) SHUTTLING IS REQUIRED — spatial distinctness of the two goal targets is a mandatory conjunct, not optional. Detection = alternation AND futility AND shuttling. (2) GOAL CLASSES are more than two; establish a first-class need/goal-class taxonomy (candidate home: tool registry metadata, each tool tagged with the need-classes it serves, so detector + prompts + reflex share one taxonomy). Initial classes: sustenance (forage/hunt/eat/cook), warmth (goto_warmth/build_fire/refuel_fire), shelter-safety (build_shelter/build_wall_*/repair/flee-type), relationship (talk_to/social verbs), and WONDER — the why-seeking class: wanting to know why things happen; proto-inquiry that is simultaneously proto-science and proto-religion and must never collapse into exclusively one or the other (nearest existing verbs: muse, journal writing, investigating rumors/lore; this class may need new affordances). (3) Depends on TASK-105: the detector's go/no-go baseline should be measured POST-context-grounding, since self-history + trajectories may let the model catch some loops unaided.

MVLS sweep dispatch (2026-07-25): research lane. Data-crunching delegated to a Sonnet agent (read-only against world-01 archives); detector definition + synthesis stays on the planning tier per constitution V. NOTE: world-01 was migrated v3→v4 today; the day-1..7 ground-truth events are in ~/.promptworld/worlds/world-01/world.v3.db (and the pre-migration backup world-01.bak-task89/).

Research complete (2026-07-25): docs/design/thrash-detection-research.md + evidence/task-106/ (reproducible analyze.py against world.v3.db). Detector: W=4h, K=8, need-progress clause + cap guard — 10/11 labeled-bad caught, 0 healthy-interleave FPs, 0 firings on never-thrashing agents. Key discoveries: thrash is daytime VILLAGE-WIDE (days 4-5 storms, 6 of 8 villagers); flip volume != pathology (need clause is load-bearing); Oak's death was NEGLECT not thrash — separate detector class recommended (critical need + zero intents in class for T). Recommendation: GO, split — card neglect detector (higher value, composes with 111/108/103); implement thrash percept after 103/104 + TASK-122 re-measure. Operator checkpoint pending on the go/no-go + carding.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Research complete (operator-accepted 2026-07-25). Deliverables: docs/design/thrash-detection-research.md + reproducible evidence (docs/design/evidence/task-106/). Detector chosen: W=4h sliding window, K=8 A->B->A class transitions, need-progress clause + cap guard — 10/11 labeled-bad episodes caught, 0 healthy-interleave false positives, 0 firings on never-thrashing agents, validated against world-01 v3 log. Key findings: thrash is daytime village-wide; flip volume != pathology (need clause load-bearing); Oak's death was neglect not thrash. Dispositions per operator: neglect detector carded as TASK-133 (High); thrash-percept implementation deferred until after TASK-103/104 + TASK-122 re-measure.
<!-- SECTION:FINAL_SUMMARY:END -->
