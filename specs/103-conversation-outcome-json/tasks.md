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

- [ ] T003 NEW `internal/mind/convo_schema.go`: `convoOutcomeSchema` built
  once from the registry gist cap, topic bounds (≤3, 40), and `sceneCap` —
  flat object, required `["gist","topics","tones","retold"]`, no `anyOf`, no
  integer min/max (parser keeps clamping values); `sayReplySchema` from the
  registry say cap, required `["say"]` (FR-002, FR-003, D3–D5).
- [ ] T004 Stamp the schemas: `Mind.outcome` Submit carries
  `convoOutcomeSchema` (name `conversation_outcome`), `Mind.utterance` Submit
  carries `sayReplySchema` (name `say`); prompts and the TASK-42
  retry/abandon ladder byte-for-byte unchanged (FR-002–FR-004). Zero
  `parse.go` diff (SC-003).
- [ ] T005 Tests alongside: `convo_schema_test.go` (valid JSON round-trip;
  caps equal the registry caps; tones `maxItems == sceneCap`; no `anyOf`);
  `convo_test.go` Request-capture via the existing `Submitter` fakes proves
  both Submit sites carry their schema and non-conversation kinds carry none;
  full existing conversation suite green; `go test -race ./...` green
  (FR-006).

## Phase 3: Measurement + soak

- [ ] T006 Soak per the measurement-run recipe: seeded MEASURE world (never
  the playtest world), conversation kind routed to gemma4:12b via Ollama,
  local-only routes (no paid spend), run until ≥ 20 scenes founded. Compute
  the D6 metrics (outcome parse-failure count; abandoned-scene rate) from
  `cog.outcome` events; record queries + counts + playtest-1 baseline (22
  outcome parse failures, 83 unusable, 62/293 = 21% abandoned) under
  `docs/design/evidence/task-174/` (FR-005, SC-001). Implementer prepares +
  runs; orchestrator reviews evidence.

## Phase 4: Grounding

- [ ] T007 Wiki re-pins in-branch: `social-fabric-conversations` prose
  amendment (constrained-decoding layer above the TASK-42 tolerance) +
  source-touched re-verifies for the llm-family notes pinning
  `llm.go`/`providers.go`/`convo.go` (plan lists them); player-docs freshness
  probe; `node scripts/check-merge-drift.mjs pr` exit 0.
