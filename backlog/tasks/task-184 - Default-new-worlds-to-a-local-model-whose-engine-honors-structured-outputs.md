---
id: TASK-184
title: Default new worlds to a local model whose engine honors structured outputs
status: Done
assignee: []
created_date: '2026-08-02 16:03'
updated_date: '2026-08-02 17:56'
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
- [x] #1 DefaultConfig's local provider names a gguf-served model that honors JSON-schema constraints, and sets reasoning_effort and tool_mode so both JSON routes and tool calls work on a freshly created world with no hand editing
- [x] #2 docs/llm-providers.md documents the MLX-versus-gguf structured-output hazard, how to check details.format before trusting a model, and the reasoning_effort knob with its measured speed effect
- [x] #3 The 'promptworld new' guidance line names the new default model rather than a stale one
- [x] #4 Tests covering DefaultConfig updated; every wiki note whose pinned sources this branch touches is re-verified and re-pinned in-branch; docs/player regenerated if docs/wiki changed
- [x] #5 Spec phase: Default config
- [x] #6 Spec phase: Tests
- [x] #7 Spec phase: Operator documentation
- [x] #8 Spec phase: Grounding (spec 069 — pr gate blocks without this)
- [x] #9 Spec phase: Verification and PR
<!-- AC:END -->



## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Dispatch tier: Sonnet — claude-sonnet-5, via the .claude/agents/spec-implementer.md definition. Rubric justification (constitution P.V): single-package change in internal/llm plus doc reconciliation and test updates; no cross-package or architectural surface, no concurrency/scheduling/governor logic, and the diagnosis is already complete at file:line so no design exploration is required. Escalation to spec-implementer-opus is NOT indicated. Carded and claimed 2026-08-02 out of the TASK-174 soak investigation.

OPERATOR DECISION 2026-08-02, resolved before implementation dispatch: default local model becomes gemma4:latest (8B gguf, 9.6GB), with tool_mode moving from 'json' to 'native' since gemma4:latest function-calls natively. qwen3.6:latest (36B MoE, 23.9GB) is documented as the recommended upgrade for capable machines rather than shipped as the default — a 24GB first pull needing ~24GB RAM is too heavy a default. cogito:3b superseded: it is CORRECT for structured output but visibly poorer in the same benchmark (emitted target:'no' as a tool argument, wrote a memory field as an identifier rather than prose). Recorded in specs/109-default-local-model/spec.md under 'Decision'. Two corrections to the earlier framing, both verified in code: (1) reasoning_effort needs NO change — zero-priced providers already resolve it to 'none' (internal/llm/providers.go:628-631, keyed on zeroPriced() at config.go:158), so no local provider pays for hidden reasoning today; (2) the SHIPPED default (cogito:3b) was never broken for structured output — the actual defect is docs/llm-providers.md, which names gemma4:12b-mlx as 'a documented upgrade path' (line 29) and uses it as the worked registry example (line 44), steering operators onto the one model that cannot hold a schema. That is what the operator's own llm.json did and the direct cause of the TASK-174/TASK-183 symptoms. Spec artifacts complete on branch task-184-default-local-model @ 6ec54b16: spec.md + plan.md + tasks.md, 13 tasks across 5 phases.

PR #155 opened 2026-08-02 — https://github.com/evanstern/promptworld/pull/155 — awaiting operator review. All 13 tasks across 5 phases complete on branch task-184-default-local-model. Implementation dispatched Sonnet (claude-sonnet-5, spec-implementer) in two phase-scoped dispatches: phases 1-3 (config + tests + docs) and phases 4-5 (grounding + verification); model that actually served: claude-sonnet-5 both times, no escalation. Verified by orchestrator independently of the implementer reports: go build/vet clean, 23 packages ok, check-merge-drift pr verdict=pass with no findings and baseLag=0. FR-004 guard holds — cmd/promptworld/commands_test.go:167 passes UNMODIFIED, so the 'ollama pull' guidance line still derives from DefaultConfig() rather than being hard-coded. SC-001 verified live on a throwaway world outside ~/.promptworld/ with an untouched llm.json: guidance line names gemma4:latest; schema-constrained call returned parseable JSON; tool call returned finish_reason=tool_calls with well-formed arguments. Re-pin honesty audited by the orchestrator against the actual diffs, not taken on report: llm-orchestrator and llm-provider-health classified NEEDS-REVIEW and had prose amended BEFORE re-pinning (both quoted the old cogito:3b/'json' default); llm-provider-registry, guardian-report-card and guardian-order-triggering were RE-PIN-ONLY with no default-model claims; llm-chain-walk-dispatch RE-PIN-ONLY, retaining its cogito:3b mention because it is a still-true TASK-52 measurement about the tool_mode mechanic rather than a claim about what ships. 5 player pages regenerated. CORRECTION to plan.md: it listed nightly-consolidation among the notes pinning internal/llm/config.go — it does not (its sources are internal/sim/consolidate.go and internal/mind/*), the orchestrator's grep matched body text rather than the sources frontmatter; the implementer caught this and left the note untouched, and the gate never flagged it.

spec-bridge sync: Default config: 2/2 · Tests: 2/2 · Operator documentation: 4/4 · Grounding (spec 069 — pr gate blocks without this): 2/2 · Verification and PR: 3/3 — status In Progress → Done
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Default config: 2/2 · Tests: 2/2 · Operator documentation: 4/4 · Grounding (spec 069 — pr gate blocks without this): 2/2 · Verification and PR: 3/3). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
