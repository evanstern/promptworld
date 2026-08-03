---
id: TASK-174
title: >-
  Local-tier JSON robustness: conversation outcomes fail on prose replies from
  small models
status: Done
assignee: []
created_date: '2026-07-30 16:42'
updated_date: '2026-08-03 06:11'
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
- [x] #2 Outcome parse-failure rate is measurable, and a soak on a small local model shows abandoned-outcome rate materially reduced from playtest-1's baseline (22 failed outcomes / 21% scenes abandoned)
- [x] #3 Spec phase: Transport — restore structured outputs (internal/llm)
- [x] #4 Spec phase: Conversation schemas (internal/mind)
- [x] #5 Spec phase: Measurement + soak
- [x] #6 Spec phase: Grounding
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
sweep dispatch (runbook playtest-1-findings-sweep): tier Sonnet — routine single-subsystem robustness following TASK-58's established structured-outputs pattern (constitution P.V: single-package feature, no concurrency/doctrine surface). Spec 103-conversation-outcome-json.

spec-bridge sync post-merge (PR #144, merged 2026-07-30): Transport 2/2, Conversation schemas 3/3, Grounding 1/1 — phases ticked. Card AC#1 satisfied (constrained decoding shipped, no cloud fallback per spec-024 pin doctrine). OPEN: Measurement + soak 0/1 (T006) — card AC#2 needs the live soak vs playtest-1 baseline; queries committed at docs/design/evidence/task-174/queries.sql; run when the shared ollama host is uncontended. Task stays In Progress until soak evidence lands.

Shared-machine coordination (relayed from the TASK-112/164 sibling session, 2026-07-30): TASK-164 arm B runs at 8x until ~tomorrow morning (matched 6.76-game-day horizon, holding 8.0 ticks/s); arm B's effective rate dropping below ~8 is the first starvation symptom on the shared local model host. T006 evidence soak therefore DEFERRED until arm B completes — it was exactly this contention that starved the first soak attempt. Sequence: run T006 after ~tomorrow morning, then sync the soak phase.

T006 combined soak STARTED 2026-08-01 15:31 (orchestrator, runbook playtest-1-findings-sweep resume). Blocker cleared: TASK-164 Done (PR #150), local host idle. ORCHESTRATOR RULING — ONE combined soak world serves both TASK-174 T006 and TASK-175 T007: same world shape, independent query sets over the same event log; operator approved 2026-08-01. Deviation recorded: stage-4 defaults, NOT specs/106 soak.md's 'harsh dials' — harsh dials suppress conversation volume that T006's >=20-founded-scenes bar needs, while sleep-cycle scheduling (T007's metric) is dial-independent. World: throwaway scratchpad (never ~/.promptworld/), seed 1337, stage-4 --override, all 10 kinds routed to a single local provider gemma4:12b-mlx @ localhost:11434/v1 — no cloud provider declared, zero paid-spend surface. Embedding route omitted (this ollama serves no /v1/embeddings) — vectorless world; neither metric depends on vector memory. Speed 16x, 16.0 ticks/s effective; target >=259200 ticks (3 game-days) AND >=20 founded scenes, ETA ~4.5h. FIELD FINDING (config, not a code defect): first attempt with tool_mode:'json' (the cogito:3b-tuned world default) made gemma4:12b-mlx fence its tool-call envelope in markdown backticks — 'invalid character backtick looking for beginning of value' — which tripped the local circuit breaker (17 planner unusable, 16 'loop: admission refused') and produced zero usable cognition. gemma4:12b-mlx advertises native tools capability; tool_mode:'native' fixed it outright (3/3 planner landed, 31 intents set, zero failures). Recommend documenting the model->tool_mode pairing in docs/llm-providers.md.

OPERATOR RULING 2026-08-01, binding, AC#2 interpretation: 'abandoned-outcome rate' means conversations abandoned BECAUSE THE OUTCOME COULD NOT BE PARSED — not all-cause abandonment. Under that reading the soak measures ZERO at 3.032 game-days: both outcome-call parse failures (reason prefix 'outcome: ', has-raw) were NON-TERMINAL 'retried' markers whose scenes went on to complete; no scene died to an outcome parse failure. AC#2 therefore PASSES. Interim counts @ tick 261994 / 3.032 game-days via docs/design/evidence/task-174/queries.sql: outcome_parse_failures=2, founded_scenes=23, abandoned_scenes=2, abandoned_pct=8.7 all-cause, breakdown landed=21 retried=4 unusable=2. Secondary/all-cause note, recorded as caveat NOT claim: 21.2% -> 8.7% is directionally better but not statistically established at n=23 (interval overlaps baseline). Soak CONTINUES OVERNIGHT per operator to build n; final counts and the evidence-doc rewrite land at close-out and supersede these interim numbers. NEW FINDING, to be carded at close-out: the utterance/say route still emits non-JSON despite sayReplySchema shipping in PR #144 — 'utterance turn 3: no JSON object in reply' (no-raw) x2 retried plus x2 terminal 'abandoned at turn 3 after retry'. BOTH of the soak's abandoned scenes died here, making the say route the dominant abandonment cause now that the outcome route is fixed. Same failure class TASK-174 fixed, different route.

CORRECTION + ROOT CAUSE 2026-08-02, supersedes the 2026-08-01 interim note. The overnight soak reached 90 founded scenes / 11.97 game-days and OVERTURNS the n=23 reading: AC#2's quantity under the operator's ruling — scenes abandoned BECAUSE the outcome could not be parsed — is 3, NOT 0. The n=23 zero was a small-sample artifact. Final counts via queries.sql: outcome_parse_failures=10, founded_scenes=90, abandoned_scenes=14, abandoned_pct=15.6 all-cause; breakdown landed=75 retried=18 unusable=14 suppressed=1. Of the 10 outcome-route parse failures, 7 were recovered by retry and 3 killed the scene (terminal unusable, reason prefix 'outcome: ', raw present). Of the 14 abandonments: 3 outcome-route, 11 utterance-route. Normalized against playtest-1 (22 failures / 293 scenes / 29 game-days): outcome parse failures 7.5 per 100 scenes baseline vs 11.1 per 100 scenes soak — NOT reduced; all-cause abandonment 21.2% vs 15.6% — modestly reduced. NOTE the baseline is not cleanly comparable for AC#2's specific quantity: playtest-1's '22' conflated routes, since one of its two cited signatures ('no JSON object in reply') is the utterance-route shape, not the outcome-route shape.

ROOT CAUSE, verified by direct probe against the rig: ollama's MLX engine does NOT honor structured-output constraints for gemma4:12b-mlx. Three independent probes all returned free prose — (a) OpenAI-compat response_format {type: json_schema} with strict, (b) same without strict, (c) ollama NATIVE /api/chat with a 'format' JSON-Schema. The code is correct and DOES send the envelope (internal/llm/providers.go:168-179 gated on len(Tools)==0; internal/mind/convo.go:618 sets convoOutcomeSchema), but the provider ignores it, so the constrained decoding AC#1 credits is INERT on this rig — which is the operator's always-on local default. This single cause explains both the outcome-route and utterance-route (TASK-183) failures. AC#1 is therefore true of the code and false of the observed behavior; AC#2 does not pass. RECOMMEND: do not close TASK-174 on this evidence — operator decision required on whether to amend scope, pin a provider/model that honors constraints, or add a parse-repair layer.

SECONDARY FINDING: gemma4:12b-mlx is a THINKING model — it spends completion tokens on a 'reasoning' field before 'content'. Under a small max_tokens budget the reasoning exhausts the budget and content comes back EMPTY, which is the likely source of the soak's 4 'empty utterance' failures (2 retried, 2 terminal). Token budgets on this route do not account for reasoning tokens.

OPERATOR RULING 2026-08-02: route chosen = pin a constraint-honoring model. RESOLVED WITH NO REPO CHANGE. Spec 109 / TASK-184 (Done, PR #155) had already moved the fresh-world default to gemma4:latest, which spec 109's spec.md benchmark measured live on the operator's M1 Max as honoring JSON-Schema constraints and tool-calling natively — the same harness that caught gemma4:12b-mlx returning prose. So the shipped default already satisfies 'pin a model that honors constraints'; promoting qwen3.6:latest to default was explicitly rejected in spec 109 on download weight, NOT correctness, and remains the documented upgrade path. Reversing that here would cross a spec boundary as well as a same-day operator ruling.

The MLX failure was operator world config, not the default: ~/.promptworld/worlds/myworld-01/llm.json pinned gemma4:12b-mlx in the legacy v1 shape, which never receives the spec 109 default since that applies only to worlds created by promptworld new. Operator confirms myworld-01 is defunct; no config change made. world-01 likewise still runs the superseded cogito:3b with tool_mode json.

AC#2 re-measurement therefore rides the running qwen3.6 soak, which spec 109 names as exactly that. Controlled A/B — same seed 1337, same binary, same stage-4 dials, only the local provider differs. At 6.54 game-days / 70 founded scenes it reads outfail=0 outkill=0 uttkill=0, against the gemma4:12b-mlx run's 10 outcome parse failures / 3 outcome kills / 11 utterance kills at 90 scenes. Needs ~20 more founded scenes for like-for-like parity before AC#2 can be closed on it; the watcher exits at 90.

Raw evidence for both runs preserved out of perishable job scratch to ~/Claude/soak-evidence/2026-08-02/ — see board doc-1. Both world.db captured via sqlite3 .backup, integrity_check ok.

spec-bridge sync: Transport — restore structured outputs (internal/llm): 2/2 · Conversation schemas (internal/mind): 3/3 · Measurement + soak: 1/1 · Grounding: 1/1 — status In Progress → Done

T006 FINAL soak evidence 2026-08-03 — SC-001 DEMONSTRATED, AC#2 PASSES at 0. Soak B on qwen3.6:latest (the spec 109 default merged in PR #155): 92 founded scenes over 9.37 game-days, outcome_parse_failures=0, abandoned_scenes=0, abandoned_pct=0.0. Breakdown: landed=90, suppressed=2 ('nothing new since last exchange' — the novelty gate working as designed, not a failure), retried=1 (utterance-route truncation, recovered, scene completed). Exactly one imperfect event in 92 scenes and it was neither terminal nor on the route this spec fixed. Against the playtest-1 baseline of 22 outcome parse failures and 62/293 = 21.2% abandoned. Under the operator's binding AC#2 ruling (scenes abandoned BECAUSE the outcome could not be parsed): 0/92 here vs 3/90 on the old model. Full three-way comparison table, the root-cause narrative, and the reproduction note in docs/design/evidence/task-174/results.md, which replaces the earlier PARTIAL write-up.

SUPERSEDES the 2026-08-01 interim note and its correction: the n=23 reading of soak A reported AC#2 as 0 and was WRONG (true value 3 at n=90). That error is recorded in results.md rather than quietly dropped, because it nearly closed this task on a false negative. Soak B's zero is credible for two independent reasons: it is a zero over 92 scenes, and it was PREDICTED in advance by a direct mechanism probe (the constraint provably reaches the sampler on gguf and is provably discarded on MLX) rather than discovered by looking at outcomes.

Spun out of this task's investigation: TASK-184/spec 109 (default local model → gguf, MLX hazard documented, merged PR #155), TASK-185 (daemon-start capability probe), TASK-186 (dead-path scheduling leak), and TASK-173 re-opened (absence storyline resurfaces past the 4-day window that justified dropping it). TASK-183 re-scoped by this evidence: the utterance route's failure changed character from 'prose, no raw payload' to 'well-formed JSON, truncated' — a token-budget problem, not a schema-adherence one.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Transport — restore structured outputs (internal/llm): 2/2 · Conversation schemas (internal/mind): 3/3 · Measurement + soak: 1/1 · Grounding: 1/1). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
