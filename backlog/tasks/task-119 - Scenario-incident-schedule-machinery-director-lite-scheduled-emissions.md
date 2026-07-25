---
id: TASK-119
title: Scenario incident-schedule machinery (director-lite scheduled emissions)
status: In Progress
assignee: []
created_date: '2026-07-25 04:43'
updated_date: '2026-07-25 18:22'
labels:
  - learning-game
  - design-session
dependencies: []
ordinal: 90000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Learning-game synthesis Wave 2 (operator decision 2: director-lite first). The one honest new design step under the staged-session plan: nothing today injects at a future tick on schedule. Nearest shipped primitives are executor emissions that are pure functions of (state, tick) — charge regen, order expiry — so v1 most plausibly rides an executor-style scenario block in world config (deterministic, replay-safe by the same argument), NOT the InjectSocial door. Powers scenario worlds: promptworld new --scenario first-night = seeded world + authored incident schedule + event-derived rubric + morgue epitaph on failure; each scenario is one prompting lesson in fiction (first-night teaches visions+orders; curfew-repeal teaches omens+governance — TASK-68(b) sketches both). Scheduled incidents double as lesson triggers for the first-occurrence projection. The live state-watching storyteller (watchers over player state, persona-named difficulty dial) is the post-v1 graduation of this machinery. Grounding: Analysis-Learning-Game-Fit recs 1 and 6.

Spec: specs/054-scenario-machinery
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
- [ ] #8 Spec phase: Setup
- [ ] #9 Spec phase: Foundational — the sim machinery
- [ ] #10 Spec phase: User Story 3 — promptworld new --scenario (P2, small; unblocks manual testing)
- [ ] #11 Spec phase: User Story 1+2 e2e (P1)
- [ ] #12 Spec phase: User Story 4 — the exercise tab (P2)
- [ ] #13 Spec phase: User Story 5 — narration + morgue (P3)
- [ ] #14 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reorient 2026-07-25 rescope (D11/D4): scenario worlds get an exercise dock tab (framing line, LIVE event-derived rubric gauges via the decision-trace projection pattern, pass/fail state) plus an attach-time briefing; incident-schedule visibility is a per-exercise VOCABULARY field, not a boolean (forecast at stages 1-2, fog from stage 3 as defaults); a scenario-cadence narration trigger closes chronicle chapters on rubric beats so the score narrative renders during short runs (the narrator's ~2 chapters/game-day would otherwise produce zero entries). curriculum.* pass emissions double as the ceremony trigger (TASK-127). Panel page authored in TASK-123 before build. Open (parked in synthesis): headline-live gauges vs full-breakdown-at-postmortem.

Model tier: Opus 4.8 (spec-implementer, model=opus). Rubric: sim-loop/executor emission class + determinism doctrine, cross-package (sim/daemon/world/ipc/mind/scribe/tui) — senior tier per constitution Principle V and the runbook Lane 2 assignment. Standing resolution: live rubric gauges per authored panels/exercise.md (parked question resolved by the page). Dispatched by UI-sweep orchestrator.

[merge-drift session] warn: task-119-scenario-machinery is cleanup-eligible (ancestor): git worktree remove /Users/evanstern/evan/promptworld/.worktrees/task-119 && git branch -d task-119-scenario-machinery
evidence: /Users/evanstern/evan/promptworld/.worktrees/task-119, task-119-scenario-machinery
fingerprint: 0c8de6eeb45c

[merge-drift session] warn: task-119-scenario-machinery and task-121-skinnable-guardian will conflict on internal/tui/views.go whichever merges first
evidence: internal/tui/views.go, task-119-scenario-machinery, task-121-skinnable-guardian
fingerprint: 67484f01df7f
<!-- SECTION:NOTES:END -->
