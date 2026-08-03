---
id: TASK-185
title: Warn at daemon start when a provider ignores structured-output constraints
status: To Do
assignee: []
created_date: '2026-08-02 16:23'
updated_date: '2026-08-03 05:47'
labels:
  - llm
  - local-tier
  - observability
  - provider-health
dependencies: []
ordinal: 167001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The game asks its local model to answer in a fixed JSON shape and then quietly accepts whatever comes back. If the model's engine ignores that request, nothing anywhere says so — the damage only shows up game-days later as conversations that die part-way through for no visible reason. Check once when the world starts, and say so plainly if the provider cannot do what the game needs.

As a player starting a world with a model that cannot hold the shape the game asks for, I want to be told at startup, in one line I can act on, rather than discovering it days later through villagers whose conversations keep evaporating.

As an operator debugging a world that is behaving strangely, I want to rule the provider in or out in seconds instead of instrumenting the event log.

Evidence and motivation: this cost 12 game-days of soak to find once already (TASK-174, 2026-08-01/02). Ollama's MLX engine accepts an OpenAI-compat response_format json_schema envelope, and Ollama's own native format parameter, and silently discards both — verified on gemma4:12b-mlx across three separate mechanisms, while its own gguf sibling honored all of them. promptworld sent the envelope correctly the whole time (internal/llm/providers.go:168-179), so every downstream symptom looked like a promptworld bug: parse failures, retries, then 14 of 90 conversation scenes abandoned over 12 game-days. Neither the boot log, nor promptworld status, nor promptworld calibrate reported anything wrong. A single schema-constrained request at daemon start would have caught it immediately. Related: TASK-184 (spec 109) changes the shipped default away from an affected model and documents the hazard, but documentation only helps operators who read it before choosing — this card is the mechanical guard that does not depend on that. Also relevant to TASK-183, whose utterance-route failures share this root cause.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 At daemon start, each openai_compat provider that serves a schema-constrained route is probed once with a trivial JSON-Schema request, and the reply is checked for validity against that schema
- [ ] #2 A provider that fails the probe produces a loud, actionable warning naming the provider, the model, and what will degrade — in the same place other provider warnings surface — without preventing the world from starting
- [ ] #3 The probe is cheap and bounded: one small request per provider at most, with a timeout, and it never runs per-call or on a hot path
- [ ] #4 A provider that passes, or a world with no llm.json, produces no new output; the probe result is observable via promptworld status
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
EVIDENCE 2026-08-03 from the spec 110 (TASK-173) re-narration, a clean quantification of the cost this warning would prevent.

Narrating 24 replayed chapters with gemma4:12b-mlx through the PRODUCTION narrator path - production prompts, production request body, production parseNarration, production truncation ladder 800/1600/3200 with maxTruncationRetries=2:

- 7 of 24 chapters (29%) spent the ENTIRE ladder on the model's 'reasoning' field and returned empty or truncated JSON. parseNarration correctly treated each as unusable, so those chapters are narration GAPS - three model calls each, all wasted, three chapters' worth of story lost.
- The same 24 chapters through qwen3.6 with reasoning_effort: none had 1 parse failure and completed the whole log in 142 seconds, against gemma's 2.6 hours.
- The trait is identical on current main and was in force during the original live soak, so it is an environment property, not a regression.

This is the exact failure mode this card proposes to warn about, and it is silent today: the operator sees missing chronicle chapters, not a provider-capability problem. Note it compounds with the ladder - a provider that ignores structured-output constraints does not just fail, it fails three times as expensively, because every retry re-pays the same reasoning tokens. Worth considering whether the warning should also short-circuit the truncation ladder for a provider known to ignore the constraint. Cross-ref: local-model default is TASK-184; the spec 110 evidence artifact is specs/110-absence-attribution/evidence.md on branch task-173-absence-attribution.
<!-- SECTION:NOTES:END -->
