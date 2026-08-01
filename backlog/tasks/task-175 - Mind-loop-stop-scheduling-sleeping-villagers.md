---
id: TASK-175
title: 'Mind loop: stop scheduling sleeping villagers'
status: In Progress
assignee: []
created_date: '2026-07-30 16:42'
updated_date: '2026-08-01 19:32'
labels: []
dependencies: []
ordinal: 143000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The mind loop keeps planning for villagers who are asleep, and the sim rejects every one of those intents — hundreds of wasted planner cycles on a tier where each call is expensive.

As a player on a slow local model, I want my limited LLM throughput spent on villagers who are awake and can act on the plan.

Evidence (playtest-1): 905 of 1,486 agent.intent_rejected events were "X is asleep" (Ash 131, Birch 127, Cedar 124, Oak 110, ...). Each represents a planner round-trip (~37s avg wall on local gemma) whose output was dead on arrival. The mind should gate planning on sleep state (or the scheduler should skip sleeping agents until wake).

Spec: specs/106-sleep-gated-planning
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Sleeping villagers do not consume planner calls (mind gates on sleep state or scheduler skips until agent.woke)
- [ ] #2 A soak shows 'is asleep' intent rejections reduced to near zero from playtest-1's baseline of 905/29 days
- [x] #3 Spec phase: Unavailability mirror + pre-submit gate (US1 layer 1)
- [x] #4 Spec phase: In-flight cancel (US1 layer 2)
- [x] #5 Spec phase: Wake resumption + regression (US2)
- [ ] #6 Spec phase: Soak evidence (card AC #2)
- [x] #7 Spec phase: Reconcile with spec 102 + grounding + gates
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
sweep dispatch (runbook playtest-1-findings-sweep): tier Opus 4.8 — scheduling logic in internal/mind orchestration, named explicitly by constitution P.V's Opus rubric; small diff but edits the mind driver TASK-112's branch also touches. Spec 106-sleep-gated-planning. LANE 2: PR merges only after TASK-112's PR lands (operator ruling), after TASK-172.

Implementation complete on branch task-175-sleep-gated-planning @ d62fb2fe (Opus 4.8, gated by orchestrator): T001-T006+T009 done — unavailability mirror + dequeue gate + in-flight cancel, all in internal/mind/mind.go (+132/-1), 401-line test file; full suite 23 pkgs green, -race green, internal/sim diff EMPTY (SC-004), 12 wiki notes re-pinned, player docs regenerated, pr gate exit 0. OPEN: T007 live soak (queries+recipe in specs/106-sleep-gated-planning/soak.md, runs post-merge before Done), T008 post-112 reconcile. PR held per operator ruling.

spec-bridge sync post-merge (PR #148, merged 2026-07-30): mirror+gate 3/3, in-flight cancel 2/2, wake resumption 1/1, reconcile+grounding 2/2 — phases ticked; card AC#1 satisfied (both post-enqueue windows closed, sim ladder byte-unchanged). OPEN: Soak evidence 0/1 (T007) — live ≥3-game-day soak vs the 905/29-day baseline, recipe+queries in specs/106-sleep-gated-planning/soak.md; deferred until TASK-164 arm B frees the shared local host (~2026-07-31 morning). Task stays In Progress until soak evidence lands.

T007 soak STARTED 2026-08-01 15:31 (orchestrator, runbook playtest-1-findings-sweep resume). Blocker cleared: TASK-164 Done (PR #150), local host idle. Running as ONE combined soak world shared with TASK-174's T006 (operator approved 2026-08-01) — the two query sets are independent reads of the same event log, so one >=3-game-day run satisfies both. Deviation from specs/106-sleep-gated-planning/soak.md recorded: stage-4 defaults instead of the recipe's 'harsh dials'. Rationale — SC-003's metric ('is asleep' agent.intent_rejected per game-day vs the 31.2 baseline, plus zero planner cog.thought for agents asleep at submit) is dial-independent: villagers keep the same sleep cycle either way, while harsh dials would starve the conversation volume TASK-174's co-hosted T006 needs. World: throwaway scratchpad (never ~/.promptworld/), seed 1337, stage-4 --override, all kinds on local gemma4:12b-mlx, no cloud provider declared. Embedding omitted (no /v1/embeddings on this ollama) — vectorless; sleep gating does not depend on vector memory. Speed 16x, 16.0 ticks/s effective; target >=259200 ticks (3 game-days), ETA ~4.5h. Queries per soak.md sections 1-3; counts land on this card when the run completes. Note: a first attempt died to a config error (tool_mode:'json' vs gemma's native tool-calling) and was torn down; the run of record is the tool_mode:'native' restart — see TASK-174's note for the finding.
<!-- SECTION:NOTES:END -->
