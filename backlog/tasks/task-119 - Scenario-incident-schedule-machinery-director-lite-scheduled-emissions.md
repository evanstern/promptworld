---
id: TASK-119
title: Scenario incident-schedule machinery (director-lite scheduled emissions)
status: To Do
assignee: []
created_date: '2026-07-25 04:43'
updated_date: '2026-07-25 14:45'
labels:
  - learning-game
  - design-session
dependencies: []
ordinal: 90000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Learning-game synthesis Wave 2 (operator decision 2: director-lite first). The one honest new design step under the staged-session plan: nothing today injects at a future tick on schedule. Nearest shipped primitives are executor emissions that are pure functions of (state, tick) — charge regen, order expiry — so v1 most plausibly rides an executor-style scenario block in world config (deterministic, replay-safe by the same argument), NOT the InjectSocial door. Powers scenario worlds: promptworld new --scenario first-night = seeded world + authored incident schedule + event-derived rubric + morgue epitaph on failure; each scenario is one prompting lesson in fiction (first-night teaches visions+orders; curfew-repeal teaches omens+governance — TASK-68(b) sketches both). Scheduled incidents double as lesson triggers for the first-occurrence projection. The live state-watching storyteller (watchers over player state, persona-named difficulty dial) is the post-v1 graduation of this machinery. Grounding: Analysis-Learning-Game-Fit recs 1 and 6.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A deterministic scheduled-emission primitive exists (pure function of state+tick), replay-safe
- [ ] #2 Scenario definitions carry authored incident schedules + event-derived rubrics + a pass/fail signal in world config
- [ ] #3 At least one runnable scenario (e.g. first-night) demonstrates schedule -> incidents -> rubric -> morgue epitaph end to end
- [ ] #4 Design leaves a documented seam for the post-v1 live director
- [ ] #5 Exercise panel: framing + live rubric gauges + pass/fail in scenario worlds
- [ ] #6 Per-exercise visibility vocabulary (not boolean); attach-time briefing
- [ ] #7 Scenario-cadence narration renders score-narrative chapters during short runs
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reorient 2026-07-25 rescope (D11/D4): scenario worlds get an exercise dock tab (framing line, LIVE event-derived rubric gauges via the decision-trace projection pattern, pass/fail state) plus an attach-time briefing; incident-schedule visibility is a per-exercise VOCABULARY field, not a boolean (forecast at stages 1-2, fog from stage 3 as defaults); a scenario-cadence narration trigger closes chronicle chapters on rubric beats so the score narrative renders during short runs (the narrator's ~2 chapters/game-day would otherwise produce zero entries). curriculum.* pass emissions double as the ceremony trigger (TASK-127). Panel page authored in TASK-123 before build. Open (parked in synthesis): headline-live gauges vs full-breakdown-at-postmortem.
<!-- SECTION:NOTES:END -->
