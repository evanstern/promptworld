---
id: TASK-183
title: >-
  Utterance route JSON robustness: villager spoken lines fail on prose replies
  from small models
status: To Do
assignee: []
created_date: '2026-08-02 00:32'
updated_date: '2026-08-02 13:37'
labels:
  - local-tier
  - llm
  - conversation
dependencies: []
ordinal: 165001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
When two villagers talk, every spoken line is its own AI call that has to come back as JSON. Small local models sometimes answer with plain prose instead, and after a single retry the whole conversation is abandoned part-way through — the villagers simply stop mid-scene and nothing they said leaves a mark. This is the same failure TASK-174 fixed for the conversation-outcome call, but on the route that produces the dialogue itself.

As a player running the village on a home machine, I want a conversation that has already started to play out to its end, rather than silently dying half-way through, even when my local model is a small one.

As a player reading the chronicle, I want the scenes I see to be whole ones, so the story does not have holes in it where a villager stopped talking for no reason I can see.

Evidence (TASK-174/175 combined soak, 2026-08-01, 3.03 game-days, gemma4:12b-mlx local, seed 1337 stage-4): of 23 founded conversation scenes, 2 were abandoned and BOTH died on this route — terminal cog.outcome{unusable} with reason 'abandoned at turn 3 after retry: no JSON object in reply' — plus 2 further non-terminal cog.outcome{retried} markers with reason 'utterance turn 3: no JSON object in reply'. All four carry NO raw payload, distinguishing them from the outcome-call parse failures (reason prefix 'outcome: ', raw present) that TASK-174 addressed. In the same soak the outcome route lost ZERO scenes, so the say route is now the DOMINANT cause of abandoned conversations. Note that sayReplySchema shipped in PR #144 (spec 103) yet these calls still return non-JSON — so the fix is not simply 'declare a schema', and the first step is diagnosing why the existing schema is not constraining this route. Prior art: TASK-58 (local planner, structured outputs), TASK-174 / spec 103 (conversation-outcome, constrained decoding).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Utterance/say calls on the local tier reliably produce parseable replies (diagnose why sayReplySchema does not currently constrain this route, then fix at the transport or schema layer as the diagnosis directs)
- [ ] #2 A soak on a small local model shows scenes abandoned at the utterance turn reduced to near zero from this soak's baseline of 2/23 founded scenes
- [ ] #3 Utterance parse-failure and utterance-abandonment counts are measurable from the event log alone, in the shape of docs/design/evidence/task-174/queries.sql
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
AC#1's diagnosis is ALREADY ANSWERED by the TASK-174 overnight soak (2026-08-02) — the cause is NOT the schema code. Verified by direct probe against the rig: ollama's MLX engine does not honor structured-output constraints for gemma4:12b-mlx. Three probes, all returning free prose: (a) OpenAI-compat response_format {type: json_schema, strict: true}, (b) the same non-strict, (c) ollama's NATIVE /api/chat 'format' JSON-Schema parameter. sayReplySchema IS sent correctly (internal/mind/convo.go:570; internal/llm/providers.go:168-179 attaches it whenever the request carries no tools) — the provider simply ignores it. So this card's scope shifts from 'fix the schema' to 'the local tier cannot rely on sampler-level constraint on this engine': the fix is a provider/model capability probe plus a fallback strategy (parse-repair pass, a model or engine that honors constraints, or a retry ladder), not a schema declaration. NOTE this is the SAME root cause as TASK-174's residual outcome-route failures, so the two are one defect surfacing on two routes — consider merging scope or sequencing them together.

SECOND CAUSE for a subset of this card's failures: gemma4:12b-mlx is a THINKING model that spends completion tokens on a 'reasoning' field before emitting 'content'. Under a tight max_tokens budget the reasoning exhausts the budget and content returns EMPTY — the soak's 'empty utterance' failures (2 retried, 2 terminal) are this, not a JSON-shape problem. Full soak baseline for this card @ 90 founded scenes / 11.97 game-days: 11 of 14 total abandonments died on the utterance route — 7 'abandoned at turn 3 after retry: no JSON object in reply', 2 at turn 2, 1 'abandoned at turn 3 after retry: empty utterance', 1 at turn 5 — plus 11 non-terminal retried markers on the same route.
<!-- SECTION:NOTES:END -->
