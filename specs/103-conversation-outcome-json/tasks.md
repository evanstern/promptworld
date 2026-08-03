# Tasks: Conversation-outcome JSON robustness (TASK-174)

**Input**: `specs/103-conversation-outcome-json/spec.md` (D1–D6 ratified there).

## Phase 1: Transport — restore structured outputs (internal/llm)

- [X] T001 Restore additive `Request.ResponseSchema` (`json.RawMessage`) +
  `SchemaName` (`string`), both omitempty, documented (Anthropic ignores; not
  attached beside Tools); `openai_compat.callNative` attaches
  `response_format {type: json_schema, json_schema: {name, schema}}` iff a
  schema is set and `len(req.Tools) == 0` — TASK-58's `f6bd31ae` shape
  (FR-001).
- [X] T002 Provider tests in `internal/llm/providers_test.go`: envelope
  present and well-formed iff schema set; payload byte-identical (no
  `response_format` key) when unset; no attach when Tools ride the same
  request (FR-001, FR-006).

## Phase 2: Conversation schemas (internal/mind)

- [X] T003 NEW `internal/mind/convo_schema.go`: `convoOutcomeSchema` built
  once from the registry gist cap, topic bounds (≤3, 40), and `sceneCap` —
  flat object, required `["gist","topics","tones","retold"]`, no `anyOf`, no
  integer min/max (parser keeps clamping values); `sayReplySchema` from the
  registry say cap, required `["say"]` (FR-002, FR-003, D3–D5).
- [X] T004 Stamp the schemas: `Mind.outcome` Submit carries
  `convoOutcomeSchema` (name `conversation_outcome`), `Mind.utterance` Submit
  carries `sayReplySchema` (name `say`); prompts and the TASK-42
  retry/abandon ladder byte-for-byte unchanged (FR-002–FR-004). Zero
  `parse.go` diff (SC-003).
- [X] T005 Tests alongside: `convo_schema_test.go` (valid JSON round-trip;
  caps equal the registry caps; tones `maxItems == sceneCap`; no `anyOf`);
  `convo_test.go` Request-capture via the existing `Submitter` fakes proves
  both Submit sites carry their schema and non-conversation kinds carry none;
  full existing conversation suite green; `go test -race ./...` green
  (FR-006).

## Phase 3: Measurement + soak

- [X] T006 Soak per the measurement-run recipe: seeded world (never the
  playtest world), conversation kind routed to a local model, local-only
  routes (no paid spend), run until >= 20 scenes founded. Compute the D6
  metrics (outcome parse-failure count; abandoned-scene rate) from
  `cog.outcome` events; record queries + counts + playtest-1 baseline (22
  outcome parse failures, 83 unusable, 62/293 = 21% abandoned) under
  `docs/design/evidence/task-174/` (FR-005, SC-001).
  **DONE 2026-08-03 — SC-001 demonstrated.** Full results, both soaks, and
  the three-way comparison table in
  `docs/design/evidence/task-174/results.md`; `queries.sql` is the reusable
  D6 query file and runs unchanged against either world.
  Headline (soak B, `qwen3.6:latest`, the spec 109 default): **92 founded
  scenes over 9.37 game-days, 0 outcome parse failures, 0 abandoned
  scenes** — against a playtest-1 baseline of 22 parse failures and
  62/293 = 21% abandoned.
  Soak A (`gemma4:12b-mlx`, 90 scenes / 11.97 game-days) measured 10 parse
  failures and 3 scenes killed by one, which led to the root cause: Ollama's
  MLX engine silently discards schema constraints, so spec 103's correct code
  never reached the sampler. That produced TASK-184/spec 109 (default moved
  to a gguf model, merged PR #155) and TASK-185 (daemon-start capability
  probe). Soak B is the re-run on the new default.
  Residual, re-scoping TASK-183: the say/utterance route's one remaining
  retry changed character from "prose, no raw" to "well-formed JSON,
  truncated" — a token-budget problem, not a schema-adherence one.
## Phase 4: Grounding

- [X] T007 Wiki re-pins in-branch: `social-fabric-conversations` prose
  amendment (constrained-decoding layer above the TASK-42 tolerance) +
  source-touched re-verifies for the llm-family notes pinning
  `llm.go`/`providers.go`/`convo.go` (plan lists them); player-docs freshness
  probe; `node scripts/check-merge-drift.mjs pr` exit 0.
