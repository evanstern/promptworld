---
id: TASK-184
title: Default new worlds to a local model whose engine honors structured outputs
status: In Progress
assignee: []
created_date: '2026-08-02 16:03'
updated_date: '2026-08-02 16:05'
labels:
  - llm
  - local-tier
  - defaults
dependencies: []
ordinal: 166001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
<!-- SECTION:DESCRIPTION:BEGIN -->
A freshly created world currently points its local tier at a model that cannot be trusted to return JSON, so villager conversations and plans fail in ways that look like game bugs. Change what a new world ships with, so the local tier works out of the box and does not spend twenty times longer than it needs to on reasoning tokens nobody reads.

As a player creating my first world, I want villagers to talk and plan correctly straight away, without having to learn which local models format their replies properly.

As an operator running the sim, I want the shipped default to be fast enough that the game does not start skipping villagers' thinking at ordinary speeds.

Evidence (measured on the operator's M1 Max, 2026-08-02, same harness for every model). Structured output via Ollama's schema constraint: cogito:3b OK, gemma4:latest OK, qwen3.6:latest OK, gemma4:12b-mlx FAILS — it returns prose. The split is the build format, not the model family or size: Ollama's MLX engine (details.format=safetensors) silently ignores schema constraints, while every gguf build honors them. gemma4:12b-mlx failed head-to-head against its own gguf sibling. Second finding: gemma4 and qwen3.6 are thinking models that burn 300-1000 reasoning tokens per call; setting reasoning_effort=none cut a schema-constrained call from 28.6s to 0.7s on qwen3.6 and 14.4s to 1.5s on gemma4:latest, with tool calls going 23.2s to 1.0s, and JSON stayed valid throughout. promptworld ALREADY transmits reasoning_effort (internal/llm/providers.go:146 and :273, passed through verbatim by resolveReasoningEffort at config.go:442), so this is reachable from llm.json with no code change — what needs changing is the DEFAULT. Calibration on qwen3.6:latest with reasoning_effort=none measures 1.0 s/point against the 20 s/point bootstrap estimate, clearing every cognition class at high speed; the same calibration could not complete at all on gemma4:12b-mlx. Note internal/llm/config.go:455 already records that gemma4:12b-mlx 'never function-called reliably out of the box', which is why DefaultConfig ships cogito:3b — this card extends that finding to structured outputs and picks a stronger default. Related: TASK-174 (outcome route), TASK-183 (utterance route) — both were chasing what is really this one provider-capability problem.
<!-- SECTION:DESCRIPTION:END -->

Spec: specs/109-default-local-model
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 DefaultConfig's local provider names a gguf-served model that honors JSON-schema constraints, and sets reasoning_effort and tool_mode so both JSON routes and tool calls work on a freshly created world with no hand editing
- [ ] #2 docs/llm-providers.md documents the MLX-versus-gguf structured-output hazard, how to check details.format before trusting a model, and the reasoning_effort knob with its measured speed effect
- [ ] #3 The 'promptworld new' guidance line names the new default model rather than a stale one
- [ ] #4 Tests covering DefaultConfig updated; every wiki note whose pinned sources this branch touches is re-verified and re-pinned in-branch; docs/player regenerated if docs/wiki changed
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Dispatch tier: Sonnet — claude-sonnet-5, via the .claude/agents/spec-implementer.md definition. Rubric justification (constitution P.V): single-package change in internal/llm plus doc reconciliation and test updates; no cross-package or architectural surface, no concurrency/scheduling/governor logic, and the diagnosis is already complete at file:line so no design exploration is required. Escalation to spec-implementer-opus is NOT indicated. Carded and claimed 2026-08-02 out of the TASK-174 soak investigation.
<!-- SECTION:NOTES:END -->
