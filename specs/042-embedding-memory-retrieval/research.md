# Phase 0 Research: Embedding Memory Retrieval

**Feature**: specs/042-embedding-memory-retrieval · **Date**: 2026-07-24

Upstream grounding: `research/Agent-Memory-Retrieval/` (vault branch: 13-search fan-out +
Analysis-Embedding-Retrieval-Adoption). This file resolves the plan-level unknowns against
the actual codebase (pins from the 2026-07-24 exploration pass).

## D1 — Where embedding happens: an async mind-side driver, never the executor

**Decision**: Embedding runs as a new asynchronous driver on the mind side (peer of the
consolidation daemon), never inside the sim executor or reducer. It watches
`agent.memory_added` events, computes vectors, and injects them back as recorded
companion events through the existing `InjectSocial` door (whitelist at
`internal/sim/loop.go:193-256`). The reducer attaches vectors to memories; replay just
re-applies the recorded events.

**Rationale**: The sim layer is model-free by doctrine (spec 030, FR-002 "no text
inspection, no heuristics"); memories are emitted by the deterministic executor inside
the tick loop, which must not block on HTTP. The project already has exactly one pattern
for model output entering deterministic space: recorded events through the two doors
(`InjectIntent` loop.go:137-148, `InjectSocial`). An async companion event follows it
verbatim. It also gives the spec's degraded mode for free: a memory is vectorless (neutral
relevance) until its vector lands, and stays vectorless if the embedder is down — loud but
non-fatal (FR-010), tick throughput untouched (SC-005 satisfied by construction).

**Spec refinement noted**: FR-001's "recorded at emission" is realized as *requested at
emission, recorded as a companion event moments later*. Durability and never-re-embed-on-
replay — the requirement's intent — hold exactly.

**Alternatives considered**: (a) synchronous embedding inside the situated constructors —
violates model-free sim, blocks the tick loop on HTTP, rejected. (b) Embedding inside the
reducer at apply time — worse: the reducer must be a pure function of the event stream,
rejected outright.

## D2 — Embedding transport: extend internal/llm with an `embedding` kind over openai_compat

**Decision**: Add an embeddings call path to the existing `openai_compat` transport
(POST `endpoint + "/embeddings"`, sibling of the `/chat/completions` path at
`internal/llm/providers.go:440-441`), a new route kind `embedding` in the v2 kinds enum
(`internal/llm/llm.go:33-53`), and a provider entry in `llm.json`. Default pin:
**`all-minilm` served by Ollama** (384-dim MiniLM-L6-v2 class, matching the spec
assumption). Route absent → embedding subsystem off → vectorless world; no boot error
(post-v2 kind backfill pattern, but deliberately *without* backfilling a default route:
absence is the off switch, mirroring "deleting llm.json disables the orchestrator",
docs/llm-providers.md:199-200).

**Rationale**: The project already runs local models through Ollama via openai_compat
(TASK-89: cogito:3b / gemma4 tiers); Ollama serves OpenAI-compatible `/v1/embeddings` for
embedding models. Reusing the provider registry gives model pinning (the provider's
`model` field is the recorded identity), routing, and operator ergonomics for free.
Grounding: hosted embedding APIs are non-deterministic and network-bound (vault
_grounding §7) — excluded; local small models embed in ~4 ms class (§8).

**Alternatives considered**: (a) hugot / ONNX Runtime in-process — new C bindings,
cross-compile burden, a second inference stack beside the existing one; rejected. (b)
llama.cpp sidecar — a new external process to document and manage when Ollama already
fills the local-inference role here; rejected. (c) Hosted APIs — non-deterministic,
rejected by grounding.

## D3 — Vector storage shape: JSON float32 arrays in the companion event payload

**Decision**: Vectors ride as plain JSON `[]float32` arrays plus a model-identity string
in the companion event payload, copied verbatim by the reducer into
`Memory.Vec []float32` / `Memory.VecModel string` (omitempty additive fields). No
quantization, no base64, no FormatVersion bump.

**Rationale**: Every payload in the store is human-readable JSON with zero binary
precedent (store: SQLite `payload TEXT`, `internal/store/schema.go:11-24`; no base64
anywhere outside tests). A 384-float JSON array is ~4–6 KB; at village scale (tens of
agents, thousands of memories) that is tens of MB in SQLite — immaterial (vault grounding
§6). Additive omitempty fields keep pre-feature events and snapshots byte-identical — the
established spec-019/030 pattern (`internal/sim/agents.go:159-178`). Go's `json.Marshal`
of float32 slices is deterministic (shortest round-trip repr), and cosine is computed in a
sequential float64 loop — no parallel-reduction nondeterminism (the ONNX-envelope caveat
from grounding §7 applies to *producing* vectors, which happens once and is recorded, not
to consuming them).

**Alternatives considered**: int8 quantization + base64 (~512 B/memory) — breaks the
human-readable-payload doctrine for a saving nobody needs at this scale; revisit only if
stores grow orders of magnitude. Separate SQLite table for vectors — splits a memory's
identity across stores and complicates snapshot/replay; rejected.

## D4 — Memory identity for the companion event: the emission event's store seq

**Decision**: The reducer stamps each memory with the `seq` of its `agent.memory_added`
event (`Memory.Seq int64`, omitempty additive). The companion event
(`agent.memory_embedded`) targets `{agent, mem_seq}`; the reducer attaches the vector to
the matching memory, no-op if absent (e.g. agent died in the gap).

**Rationale**: Memories currently have no id; (agent, tick) is not unique (one tick can
emit several memories). Store seq is unique, totally ordered, gapless-enforced
(`store.CheckContiguity`, store.go:210-227), available to both the reducer and the
mind's absorb loop (which already records stimulus seqs, mind.go:261-266). Deterministic
in replay because seq is part of the recorded stream.

**Alternatives considered**: content hash of (agent,tick,text) — collision-prone for
repeated identical memories and more code; UUID minted at emission — new nondeterminism
surface; both rejected.

## D5 — The situation (query) vector: cadence-embedded situation text, recorded per agent

**Decision**: The embedder driver also maintains a per-agent **situation vector**: at
planner cadence (the existing `PlannerCadenceTicks` bucket used for serendipity seeding,
`internal/sim/memory.go:336`), it renders a compact deterministic situation string from
the mind's replica state — time-of-day phase, position + place description, worst needs,
active intent verb + reason, nearby agent names — embeds it, and injects
`agent.situation_embedded {agent, tick, text, vec, model}`. Selection uses the latest
recorded situation vector at or before the selection tick; if none exists, selection
falls back to the legacy ranking (spec edge case).

**Rationale**: The relevance term needs a query vector *before* prompt assembly, but
prompt assembly happens in the absorb goroutine (mind.go:351-357) and must not block on
HTTP. A cadence-refreshed, recorded situation vector keeps selection a pure function of
recorded data (replay-safe by construction) at a cost of one embed per agent per cadence
bucket — trivial for a local 384-dim model at village scale. Recording the situation
*text* alongside the vector makes divergence audits explainable. The composition
deliberately includes the active intent so "relevant to now" means "relevant to what I am
doing and where I am", not merely "similar to my last memory".

**Alternatives considered**: (a) synchronous query embed at plan time — blocks absorb;
rejected. (b) Query = centroid of the agent's most recent memories' vectors (zero extra
embeds) — collapses relevance toward recency, which is precisely the failure mode the
US2 guardrail exists to detect; rejected as the primary design (noted as a cheap
fallback experiment the divergence instrumentation could compare someday).

## D6 — Selector: stays in internal/sim as a pure function; the mind chooses the mode

**Decision**: `SelectMemories` keeps its exact current behavior. A new pure sibling in
`internal/sim/memory.go` — `SelectMemoriesRelevant(a, seed, agentIdx, tick, k, query
[]float32, queryModel string)` — implements the three-term score. Per-memory:
`score = decayedSalience/10 + cosine(mem.Vec, query)` where `decayedSalience` is exactly
today's halving-per-game-day weight, normalized to [0,1] by the salience ceiling
(`MaxSalience`), and cosine contributes 0 (neutral) when the memory is vectorless or its
`VecModel != queryModel` (cross-model guard, FR-009). Ties → newer wins (unchanged). The
two serendipity tail picks are untouched. The mind (prompt.go:135, convo.go:177) picks
legacy vs shadow vs relevant per the world's mode flag; scribe/TUI/consolidation
consumers are untouched (they never call SelectMemories today, confirmed by call-site
sweep).

**Rationale**: Per-agent isolation is preserved by construction — the function's only
memory input remains `a.Memories` (FR-005). Keeping it sim-side keeps the shared
test surface (the "selection is a pure function shared by the mind's prompts and the
tests" doctrine, memory.go:13-16). Equal-weight normalized terms follow the
generative-agents reference (vault grounding §1); recency stays inside the salience-decay
term, creation-time only (FR-004 / guardrail 1).

**Alternatives considered**: mind-side selector — duplicates scoring logic away from its
tests; rejected. Weighted-α tuning surface — premature before US2 evidence; the
normalized equal-weight form is the reference default, revisit with divergence data.

## D7 — Shadow-mode divergence instrumentation: a recorded reducer no-op event

**Decision**: A three-state world flag `memory_relevance: "" | "shadow" | "on"` in
`world.json` (additive omitempty string; `Teaching bool` precedent, world.go:50). In
shadow mode the mind computes both rankings at plan time, hands the *legacy* window to
the prompt, and emits `cog.memory_divergence` — a reducer no-op whitelisted like the
other `cog.*` telemetry (loop.go:241-248) — carrying: agent, tick, both top-K lists (as
mem seqs), overlap count, rank-displacement sum, and how many candidates were vectorless.
In `on` mode the augmented window feeds the prompt and the divergence event still
records. The go/no-go threshold decision (FR-007) is made from a summary over ≥1 full
game day (SC-003) and recorded on the board task; flipping shadow→on is a world.json
edit, recorded in the task notes.

**Rationale**: `cog.*` reducer no-ops are the established, replay-safe home for cognition
telemetry (cog.thought/tool_call/outcome, telemetry.go). Recording both lists as seqs
keeps events small and auditable against `agent.memory_added` events. Keeping the flag in
world.json (not llm.json) is deliberate: llm.json can be deleted to disable the
orchestrator without silently changing memory-selection semantics.

**Alternatives considered**: log-file instrumentation — violates artifact doctrine
(decisions derive from recorded artifacts); rejected. Auto-flip on threshold — the
gate decision is the operator's, recorded per constitution Principle I; rejected.

## D8 — Failure loudness

**Decision**: Embedder failures (endpoint down, model missing, timeout) emit the existing
`daemon.llm_warning` event (the loud-failure channel TASK-91 established), at most once
per failure episode (debounced), while memories continue to land vectorless.

**Rationale**: Reuses the established operator-visible failure surface; satisfies US1
scenario 3 (loud, non-fatal, no stall).
