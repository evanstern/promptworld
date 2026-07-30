# Feature Specification: Local-tier JSON robustness — conversation outcomes survive prose replies

**Feature Branch**: `task-174-conversation-outcome-json`

**Created**: 2026-07-30

**Status**: Draft

**Input**: TASK-174 (sweep runbook playtest-1-findings-sweep; spec dir claimed by
the stub commit on this branch). Evidence from playtest-1 (29 game-days,
conversation routed to local gemma4:12b via Ollama): 22 conversation outcomes
failed on bad JSON — replies opening with prose ("invalid character 'F' looking
for beginning of value", "no JSON object in reply"); 62 of 293 conversations
abandoned (21%); 83 `cog.outcome=unusable`. Prior art: TASK-58 fixed the same
failure shape for the local planner with JSON-schema structured outputs; the
conversation route never got that treatment (and TASK-71 later dropped the
transport support as dead code once spec 017 moved the planner onto tool calls).

## Design decisions (resolved)

- **D1 Mechanism — sampler-level JSON-schema constrained decoding, mirroring
  TASK-58.** Restore the additive `Request.ResponseSchema` (`json.RawMessage`)
  + `SchemaName` (`string`) fields on `internal/llm.Request` (removed as dead
  code by TASK-71, commit `db45c44b`; the mechanism itself is proven —
  TASK-58 commit `f6bd31ae` drove planner unusable replies to 0/30 live). The
  `openai_compat` transport attaches `response_format {type: "json_schema"}`
  in `callNative` **iff** a schema is set **and** the request carries no
  `Tools` (a tools-carrying json-mode request already owns `response_format`
  for its envelope; conversation calls never carry tools). With no schema the
  payload stays byte-identical. The Anthropic transport ignores both fields —
  the caller's parser remains the final gate, so a request never fails for
  carrying a schema on a tier that can't use it.
- **D2 No cloud fallback.** The card's AC #1 is disjunctive (constrained
  decoding OR cloud fallback); constrained decoding is the doctrine-consistent
  arm. A cloud fallback for the outcome step would break the ratified scene
  provider pin (spec 024 US3: every utterance and the outcome call stamp the
  scene's one pinned provider; mid-scene failure flows into the TASK-42
  tolerance path, **never** a re-resolve or provider switch —
  [[social-fabric-conversations]]), and would add routing/config surface for a
  failure class TASK-58 already proved the schema eliminates. The TASK-42
  tolerance ladder (one retry per site, `lenientOutcome` repair,
  transport-never-retried, all-or-nothing abandonment) stays byte-for-byte as
  the safety net.
- **D3 Outcome schema shape — flat, required-complete, values clamped by the
  parser.** `convoOutcomeSchema` is generated programmatically (never
  hand-copied) from the existing single sources of truth: `gist` (string,
  `maxLength` = the tool registry's gist cap, 200), `topics` (array,
  `maxItems` 3, items string `maxLength` 40), `tones` (array, `maxItems` =
  `sceneCap` (4), items integer), `retold` (string, `maxLength` 300 — a
  token-budget bound per TASK-58's truncated-JSON lesson, not a game rule);
  `required`: all four keys. TASK-58's live findings are binding: no `anyOf`
  (llama.cpp's schema-to-grammar converter bails and enforces nothing), no
  integer `minimum`/`maximum` (tone clamping to −2..2 stays in `parseOutcome`
  — the schema constrains SHAPE, the parser clamps VALUES). `maxLength`
  counts code points while the registry caps are bytes; the parser's
  rune-safe byte clamp remains the authoritative cap. The legacy
  `tone_a`/`tone_b` pair shape still parses (unconstrained cloud replies may
  use it); the outcome prompt is unchanged.
- **D4 Utterance calls get the same treatment.** `sayReplySchema` —
  `{"say": string, maxLength = say cap (300)}`, `required: ["say"]` — stamps
  the utterance Submit. Same mechanism, same failure class, same tier;
  abandoned-at-turn parse failures are part of the 21% abandoned baseline the
  card's AC #2 measures. Strictly additive: parse, retry budget, and
  round-robin semantics unchanged.
- **D5 Zero `parse.go` diff.** Both schema builders live in a NEW file
  `internal/mind/convo_schema.go` (reading the package's existing
  `gistCapBytes`/`sayCapBytes`/`sceneCap`); `parse.go` is a cross-task hotspot
  (spec 105 touches it) and this feature does not modify it.
- **D6 Measurability — defined over existing telemetry, no payload change.**
  The `cog.outcome` contract (spec 011/TASK-42) already makes outcome parse
  failures countable: transport errors are never retried, so a
  `cog.outcome{retried}` with reason prefix `"outcome: "` is exactly one
  outcome parse failure; a terminal `cog.outcome{unusable}` with reason prefix
  `"outcome: "` and non-empty `raw` is a parse-killed scene (transport
  abandonment carries no raw). Metrics: **outcome parse-failure count** =
  retried events with the `outcome: ` prefix plus parse-killed unusable
  events; **abandoned-scene rate** = terminal `unusable` conversation-job
  events / founded scenes (terminal `cog.outcome` events with job prefix
  `conversation-`). The soak evidence records the queries and counts under
  `docs/design/evidence/task-174/`.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Conversations land on a small local model (Priority: P1)

As a player running the village on a home machine with a small local model, I
want conversations between villagers to reliably leave their mark —
relationship shifts, gist memories, rumors — even when the model would rather
answer the summary request with a paragraph of prose than with JSON.

**Independent Test**: run a scene against a scripted/live local provider and
observe the outcome request carrying the schema envelope and the scene landing.

**Acceptance Scenarios**:

1. **Given** a conversation scene pinned to an `openai_compat` provider,
   **When** the outcome call is submitted, **Then** the wire payload carries
   `response_format {type: json_schema}` with the generated outcome schema,
   and a schema-conformant reply parses and lands the scene's atomic batch.
2. **Given** a request with no `ResponseSchema` (every non-conversation kind),
   **When** it is submitted through `openai_compat`, **Then** the payload is
   byte-identical to today (no `response_format` key).
3. **Given** a conversation pinned to the Anthropic transport, **When** the
   outcome call is submitted, **Then** the schema fields are ignored and
   behavior is unchanged (`parseOutcome` + `lenientOutcome` remain the gate).
4. **Given** a reply that still fails to parse, **Then** the TASK-42 ladder is
   unchanged: one re-request, `cog.outcome{retried}` with `raw`, then terminal
   `cog.outcome{unusable}` and all-or-nothing abandonment.

---

### User Story 2 - The operator can measure the failure rate (Priority: P2)

As the operator running a soak, I want the outcome parse-failure rate and the
abandoned-scene rate countable from the event log alone, so "materially
reduced from playtest-1" is a recorded number, not an impression.

**Acceptance Scenarios**:

1. **Given** a completed soak world, **Then** the D6 metrics are computed from
   `cog.outcome` events alone and recorded (queries + counts) under
   `docs/design/evidence/task-174/`, alongside the playtest-1 baseline (22
   outcome parse failures, 83 unusable, 62/293 = 21% abandoned).
2. **Given** the soak (small local conversation model, ≥ 20 founded scenes),
   **Then** zero scenes are abandoned by outcome parse failure and the
   abandoned-scene rate is materially below the 21% baseline.

---

### User Story 3 - Utterances stop killing scenes too (Priority: P3)

As a villager in the game mid-conversation, I want my turns constrained to the
`{"say": ...}` shape on the local tier, so a prose-shaped turn doesn't burn the
scene's single utterance retry (or abandon the whole scene) before the summary
is ever reached.

**Acceptance Scenarios**:

1. **Given** an utterance call on an `openai_compat` provider, **Then** the
   payload carries the say schema; **Given** a parse failure anyway, **Then**
   the one-retry-per-scene utterance budget behaves exactly as today.

### Edge Cases

- Schema + `Tools` on one request: the schema is NOT attached (the json-mode
  tool envelope owns `response_format`); documented on the field. Conversation
  calls never carry tools, so this is a doctrine guard, not a live path.
- Empty reply / reply with no `{`: still a parse failure with `raw` possibly
  empty; the retry ladder covers it (schema enforcement makes it rare, not
  impossible — Ollama honors `json_schema`; a router that silently ignores it
  degrades to exactly today's behavior).
- Cloud models reached over an `openai_compat` router receive the schema too —
  transport-level decision, same doctrine as TASK-58's planner schema.
- Legacy `tone_a`/`tone_b` replies (unconstrained tiers): still parse.
- `maxLength` (code points) vs registry caps (bytes): the parser's rune-safe
  byte clamps remain authoritative; the schema bound only keeps a reply inside
  the call's token budget (TASK-58's truncated-JSON lesson).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `internal/llm.Request` regains additive, `omitempty`
  `ResponseSchema`/`SchemaName` fields; `openai_compat.callNative` attaches the
  `response_format {type: json_schema}` envelope iff a schema is set and the
  request carries no Tools; payload byte-identical when unset; the Anthropic
  transport ignores the fields (D1).
- **FR-002**: `convoOutcomeSchema` generated in new
  `internal/mind/convo_schema.go` from the registry gist cap, topic bounds,
  and `sceneCap` (D3, D5); `Mind.outcome` stamps it (+ a schema name) on its
  Submit. Prompt text unchanged.
- **FR-003**: `sayReplySchema` generated from the registry say cap;
  `Mind.utterance` stamps it on its Submit (D4).
- **FR-004**: The TASK-42 tolerance ladder is byte-for-byte unchanged:
  retry-once per site, `lenientOutcome` repair, transport errors never
  retried, all-or-nothing abandonment, `raw`/`retried` telemetry (D2). No
  reducer, event-type, or payload changes anywhere — replay untouched.
- **FR-005**: The D6 metrics are computed for the soak and recorded with their
  queries, soak counts, and the playtest-1 baseline under
  `docs/design/evidence/task-174/` (AC #2).
- **FR-006**: Tests alongside the code: provider-level envelope-iff-schema
  (set / unset / with-Tools), schema single-source assertions (valid JSON;
  caps equal the registry caps; tones `maxItems` == `sceneCap`; no `anyOf`),
  outcome and utterance Submits carry the schema (existing `Submitter` fakes
  capture the Request), full existing conversation suite green, `go test
  -race ./...` green.

## Success Criteria *(mandatory)*

- **SC-001**: On the soak (small local conversation model, ≥ 20 founded
  scenes): zero scenes abandoned by outcome parse failure; abandoned-scene
  rate materially below playtest-1's 21%; evidence recorded per FR-005.
- **SC-002**: Every request that sets no schema produces a byte-identical wire
  payload (proven at the provider test level); Anthropic-path behavior
  unchanged; zero reducer/event changes, so replay fixtures are untouched.
- **SC-003**: `internal/mind/parse.go` has an empty diff on this branch.

## Assumptions

- Soak setup follows the measurement-run recipe: seeded MEASURE world (never
  the playtest world), conversation kind routed to gemma4:12b via Ollama,
  local-only routes (no paid spend).
- Ollama's OpenAI-compat endpoint honors `response_format {type: json_schema}`
  (proven live by TASK-58 on this repo).
- Tier: **Sonnet** — routine single-subsystem robustness following TASK-58's
  established structured-outputs pattern (recorded on the board card by the
  sweep dispatch). Escalation rubric untriggered.
