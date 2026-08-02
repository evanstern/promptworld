# Spec 109 — Default new worlds to a local model whose engine honors structured outputs

**Board task:** TASK-184
**Status:** specifying
**Origin:** the TASK-174 soak investigation (2026-08-01/02), which spent 12 game-days
chasing what turned out to be a provider-capability problem, not a code defect.

## Problem

promptworld's JSON-reliability work — TASK-58 (planner), TASK-174 / spec 103
(conversation outcome), TASK-183 (utterance route) — is built on **sampler-level
constraint**: the caller sets a JSON Schema, `internal/llm/providers.go` attaches it as
an OpenAI-compat `response_format: {type: json_schema}` envelope, and the provider is
expected to make invalid output unrepresentable.

**Ollama's MLX engine silently ignores that envelope.** Measured 2026-08-02 on the
operator's M1 Max, same harness for every model:

| model | `details.format` | schema-constrained JSON | native tool call |
|---|---|---|---|
| `cogito:3b` | gguf | ✅ valid | ✅ |
| `gemma4:latest` (8B) | gguf | ✅ valid | ✅ |
| **`gemma4:12b-mlx`** | **safetensors (MLX)** | ❌ **returns prose** | ✅ |
| `qwen3.6:latest` (36B MoE) | gguf | ✅ valid | ✅ |

The split tracks the build format, not the model family or size — `gemma4:12b-mlx`
failed head-to-head against its own GGUF sibling. All three constraint mechanisms were
tried against it and all three returned free prose: OpenAI-compat `json_schema` strict,
the same non-strict, and Ollama's own native `/api/chat` `format` parameter.

**Why this was invisible.** The code is correct and does send the envelope. The provider
discards it. Every downstream symptom then looks like a promptworld bug: parse failures,
retries, abandoned conversation scenes. The TASK-174 soak measured 14 of 90 scenes
abandoned this way over 12 game-days before the cause was found. Nothing in the daemon,
the boot log, or the calibration output said "this provider ignores your schema."

**What is actually mis-set today.** Two distinct things, and only the second is a defect
in shipped state:

1. `DefaultConfig`'s local provider (`internal/llm/config.go:467`) is `cogito:3b` +
   `tool_mode: "json"`. This **is not broken** — cogito:3b honors schema constraints and
   tool-calls correctly under `json` mode. Its weakness is output *quality*: in the same
   benchmark it emitted `target: "no"` for a tool argument and wrote a memory field as an
   identifier rather than prose.
2. **`docs/llm-providers.md` actively recommends the broken path.** It names
   `gemma4:12b-mlx` as "a documented upgrade path" (line 29) and uses it as the worked
   example in the v2 registry snippet (line 44). An operator following the guide lands
   exactly on the model that cannot hold a schema. This is what the operator's own
   `llm.json` did, and it is the direct cause of the TASK-174 and TASK-183 symptoms.

## Non-goals

- Fixing the utterance route's failure handling — that is TASK-183, and this spec removes
  its root cause rather than reworking its retry ladder.
- Adding a boot-time provider-capability probe. That is the right long-term guard and is
  cited in Open questions, but it is a behavior change with its own design surface and
  does not belong in a defaults-and-docs change.
- Re-measuring TASK-174's AC #2. That is the separate qwen3.6 soak already running.

## What must be true when this is done

- **FR-001** — `DefaultConfig`'s local provider names a model served in a format whose
  engine honors JSON-Schema constraints, with `tool_mode` set to whatever that model
  actually supports natively.
- **FR-002** — A freshly created world needs no hand editing of `llm.json` for either
  JSON-shaped routes or tool-calling routes to work.
- **FR-003** — `docs/llm-providers.md` no longer recommends or exemplifies an MLX-served
  model, and documents the hazard: how to check `details.format` via
  `/api/show`, what the failure looks like (prose where JSON was demanded, then parse
  failures downstream), and that it is silent.
- **FR-004** — The `promptworld new` guidance line names the new default. It already
  derives from `llm.DefaultConfig()` (`cmd/promptworld/commands.go:318`), so this must
  stay derived rather than hard-coded, and `commands_test.go:167` must keep asserting
  that derivation.
- **FR-005** — Any doc statement about `reasoning_effort` remains accurate. Note the
  existing behavior verified during this investigation: zero-priced providers **already**
  default to `reasoning_effort: "none"` (`providers.go:628-631`, keyed on
  `zeroPriced()` at `config.go:158`), so no local provider pays for hidden reasoning
  today and this spec changes nothing there.

## Success criteria

- **SC-001** — On a freshly created world with an untouched `llm.json`, a
  schema-constrained call returns parseable JSON and a tool-calling route returns a
  well-formed tool call, both verified against a live local endpoint.
- **SC-002** — `grep -n "12b-mlx" docs/llm-providers.md` returns only occurrences that
  are explicitly cautionary, never recommending.
- **SC-003** — `go test ./...` green; `go vet` clean.
- **SC-004** — Every wiki note whose pinned sources this branch touches is re-verified and
  re-pinned in-branch, and `docs/player/` is regenerated if `docs/wiki/` changed
  (spec 069, enforced by the pr gate).

## Decision — which model becomes the default

**RESOLVED (operator, 2026-08-02): `gemma4:latest`.**

The candidates differ mainly in download weight, which is the real cost of a first-run
default. All three honor schema constraints; this was a weight-versus-quality call, not a
correctness one:

| candidate | size | RAM | quality | disposition |
|---|---|---|---|---|
| `cogito:3b` | 2.2 GB | small | weakest — visibly sloppy fields | superseded as default |
| **`gemma4:latest`** | **9.6 GB** | moderate | good | **CHOSEN** |
| `qwen3.6:latest` | 23.9 GB | ~24 GB | best | documented as the recommended upgrade |

Rationale recorded: `gemma4:latest` is clearly better than `cogito:3b` in the same
benchmark — it produced a sensible `gather`/`wood` tool call where cogito produced
`target: "no"` — while keeping "a new world works out of the box" credible without
demanding a 24 GB pull and ~24 GB of RAM. `qwen3.6:latest` becomes the documented upgrade
for capable machines rather than the default.

`gemma4:latest` tool-calls natively (verified in the same benchmark), so the default
`tool_mode` moves from `"json"` — which exists for cogito:3b, which never function-calls
natively (TASK-52) — to `"native"`.

## Open questions

- **A boot-time capability probe.** One schema-constrained request at daemon start, with
  a loud warning when the reply does not validate, converts a silent multi-day failure
  into a startup line. Proposed but deliberately out of scope here; worth its own card.
- **Whether `endpoint_capacity` guidance needs revisiting** now that a recommended local
  model may be a 24 GB MoE rather than a 2 GB dense model.
