---
id: TASK-106
title: 'Research: thrash detection as a percept — define and detect goal oscillation'
status: To Do
assignee: []
created_date: '2026-07-25 02:42'
updated_date: '2026-07-25 18:18'
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
- [ ] #1 Detection definition chosen with false-positive/false-negative analysis against world-01 replay
- [ ] #2 Percept/memory injection design sketched (salience, wording, planner-trigger interaction)
- [ ] #3 Go/no-go recommendation for implementation task
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
CLARIFICATIONS (Evan, 2026-07-24): (1) SHUTTLING IS REQUIRED — spatial distinctness of the two goal targets is a mandatory conjunct, not optional. Detection = alternation AND futility AND shuttling. (2) GOAL CLASSES are more than two; establish a first-class need/goal-class taxonomy (candidate home: tool registry metadata, each tool tagged with the need-classes it serves, so detector + prompts + reflex share one taxonomy). Initial classes: sustenance (forage/hunt/eat/cook), warmth (goto_warmth/build_fire/refuel_fire), shelter-safety (build_shelter/build_wall_*/repair/flee-type), relationship (talk_to/social verbs), and WONDER — the why-seeking class: wanting to know why things happen; proto-inquiry that is simultaneously proto-science and proto-religion and must never collapse into exclusively one or the other (nearest existing verbs: muse, journal writing, investigating rumors/lore; this class may need new affordances). (3) Depends on TASK-105: the detector's go/no-go baseline should be measured POST-context-grounding, since self-history + trajectories may let the model catch some loops unaided.
<!-- SECTION:NOTES:END -->
