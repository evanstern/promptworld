---
id: TASK-174
title: >-
  Local-tier JSON robustness: conversation outcomes fail on prose replies from
  small models
status: In Progress
assignee: []
created_date: '2026-07-30 16:42'
updated_date: '2026-08-01 19:32'
labels: []
dependencies: []
ordinal: 142000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Small local models sometimes answer the conversation-outcome call with plain prose instead of JSON, and the scene's outcome is thrown away after one retry. Constrained decoding (or a cloud fallback for the outcome step alone) would make local-tier conversations land reliably.

As a player running the village on a home machine, I want conversations between villagers to reliably leave their mark (relationship shifts, memories), even when my local model is a small one.

Evidence (playtest-1, conversation routed to local gemma4:12b via ollama): 22 conversation outcomes failed on bad JSON — replies opening with prose ("invalid character 'F' looking for beginning of value", "no JSON object in reply"); 62 of 293 conversations abandoned (21%); 83 cog.outcome=unusable. Related prior art: TASK-58 fixed the same failure shape for the local planner with JSON-schema structured outputs; the conversation-outcome (and meeting) routes never got that treatment. Cost note: cloud fallback for outcome parsing alone is trivial — the entire 29-day run spent $3.25 of a $100 budget.

Spec: specs/103-conversation-outcome-json
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Conversation-outcome calls on the local tier use constrained/JSON-mode decoding (as TASK-58 did for the planner) or fall back to cloud on parse failure
- [ ] #2 Outcome parse-failure rate is measurable, and a soak on a small local model shows abandoned-outcome rate materially reduced from playtest-1's baseline (22 failed outcomes / 21% scenes abandoned)
- [x] #3 Spec phase: Transport — restore structured outputs (internal/llm)
- [x] #4 Spec phase: Conversation schemas (internal/mind)
- [ ] #5 Spec phase: Measurement + soak
- [x] #6 Spec phase: Grounding
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
sweep dispatch (runbook playtest-1-findings-sweep): tier Sonnet — routine single-subsystem robustness following TASK-58's established structured-outputs pattern (constitution P.V: single-package feature, no concurrency/doctrine surface). Spec 103-conversation-outcome-json.

spec-bridge sync post-merge (PR #144, merged 2026-07-30): Transport 2/2, Conversation schemas 3/3, Grounding 1/1 — phases ticked. Card AC#1 satisfied (constrained decoding shipped, no cloud fallback per spec-024 pin doctrine). OPEN: Measurement + soak 0/1 (T006) — card AC#2 needs the live soak vs playtest-1 baseline; queries committed at docs/design/evidence/task-174/queries.sql; run when the shared ollama host is uncontended. Task stays In Progress until soak evidence lands.

Shared-machine coordination (relayed from the TASK-112/164 sibling session, 2026-07-30): TASK-164 arm B runs at 8x until ~tomorrow morning (matched 6.76-game-day horizon, holding 8.0 ticks/s); arm B's effective rate dropping below ~8 is the first starvation symptom on the shared local model host. T006 evidence soak therefore DEFERRED until arm B completes — it was exactly this contention that starved the first soak attempt. Sequence: run T006 after ~tomorrow morning, then sync the soak phase.

T006 combined soak STARTED 2026-08-01 15:31 (orchestrator, runbook playtest-1-findings-sweep resume). Blocker cleared: TASK-164 Done (PR #150), local host idle. ORCHESTRATOR RULING — ONE combined soak world serves both TASK-174 T006 and TASK-175 T007: same world shape, independent query sets over the same event log; operator approved 2026-08-01. Deviation recorded: stage-4 defaults, NOT specs/106 soak.md's 'harsh dials' — harsh dials suppress conversation volume that T006's >=20-founded-scenes bar needs, while sleep-cycle scheduling (T007's metric) is dial-independent. World: throwaway scratchpad (never ~/.promptworld/), seed 1337, stage-4 --override, all 10 kinds routed to a single local provider gemma4:12b-mlx @ localhost:11434/v1 — no cloud provider declared, zero paid-spend surface. Embedding route omitted (this ollama serves no /v1/embeddings) — vectorless world; neither metric depends on vector memory. Speed 16x, 16.0 ticks/s effective; target >=259200 ticks (3 game-days) AND >=20 founded scenes, ETA ~4.5h. FIELD FINDING (config, not a code defect): first attempt with tool_mode:'json' (the cogito:3b-tuned world default) made gemma4:12b-mlx fence its tool-call envelope in markdown backticks — 'invalid character backtick looking for beginning of value' — which tripped the local circuit breaker (17 planner unusable, 16 'loop: admission refused') and produced zero usable cognition. gemma4:12b-mlx advertises native tools capability; tool_mode:'native' fixed it outright (3/3 planner landed, 31 intents set, zero failures). Recommend documenting the model->tool_mode pairing in docs/llm-providers.md.
<!-- SECTION:NOTES:END -->
