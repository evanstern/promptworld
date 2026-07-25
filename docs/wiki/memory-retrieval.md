---
name: memory-retrieval
description: The spec-042 episodic-memory retrieval stack — embedding vectors recorded at emission (async mind-side embedder, warm-pinned local model), per-agent situation vectors, the pure three-term relevance selector, the memory_relevance shadow/on mode gate, and divergence instrumentation
kind: pipeline
sources:
  - internal/mind/embedder.go
  - internal/sim/memory.go
  - internal/sim/cognition.go
  - internal/mind/prompt.go
  - internal/mind/convo.go
  - internal/mind/telemetry.go
  - internal/llm/providers.go
  - internal/world/world.go
  - cmd/promptworld/divergence.go
verified_against: cc514f7ff456fefbcfe289471c5a1467b8e724df
---

# Memory retrieval (embedding relevance)

Spec 042 (TASK-98, PR #70) gives every episodic memory a recorded embedding vector and
gives working-memory selection an optional, evidence-gated relevance term. The design
splits cleanly across the sim/mind boundary: model I/O and mode policy live on the mind
side; scoring math, memory state, and event semantics stay in `internal/sim` as pure
functions of recorded data, so replay stays byte-identical with zero model calls.

## How it works

**Vectors at emission (async, recorded).** The `Embedder` driver
(`internal/mind/embedder.go`) watches absorbed `agent.memory_added` events, embeds each
memory's exact recorded text (fixed-byte cap `embedTextCapBytes`), and injects
`agent.memory_embedded {agent, mem_seq, vec, model}` through the `InjectSocial` door. The
reducer attaches `Vec`/`VecModel` to the memory whose `Seq` matches — copy-verbatim,
no-op if absent, and a zero `mem_seq` (pre-042 memory) never matches. `Memory.Seq` is
stamped from the event's store seq by the reducer, pre-stamped in the live loop by
`stampSeqs` (`internal/sim/loop.go`) because the loop applies events before the store
assigns sequence numbers. `Observe` never blocks the absorb path (drop-on-overflow), the
worker is FIFO single-flight with batching/coalescing, and per-agent companion order is
emission order.

**Situation vectors.** At each `PlannerCadenceTicks` bucket edge the driver renders a
deterministic situation string per live agent — `renderSituation`: day/night, position +
`sim.PlaceAt` description, worst needs, active intent, nearby agents (Dead and Asleep
agents are skipped) — embeds it, and injects `agent.situation_embedded`, which the
reducer stores as `Agent.SitVec/SitVecModel/SitVecTick`. The recorded text is the audit
surface for divergence review.

**Selection.** `sim.SelectMemoriesRelevant` scores `sal01 + rel01`: `sal01` is exactly
`SelectMemories`' salience-halved-per-game-day weight normalized by `MaxSalience`;
`rel01` is `(cosine+1)/2` via `relevance01` — neutral `0.5` when the memory is
vectorless, cross-model, dimension-mismatched, or zero-magnitude. A nil query delegates
to `SelectMemories` verbatim; ties break newer-first; the two serendipity tail picks run
the legacy algorithm on the same `"serendipity"` rng stream. Recency counts from
creation only — selection mutates nothing (the reference design's last-access decay was
deliberately rejected as hostile to pure selection and replay). Isolation is structural:
the only memory source is `a.Memories`.

**Mode gate.** `world.json`'s `memory_relevance` (`""`/`"shadow"`/`"on"`, unknown values
refused at `world.Open`) threads boot-static through `mind.New` into `selectWindow`
(`internal/mind/prompt.go`) and the conversation snapshot (`internal/mind/convo.go`).
Shadow computes both rankings, serves the legacy window bit-identically, and records
`cog.memory_divergence` (payload authority `sim.NewMemoryDivergencePayload`: both
windows as seqs, overlap, rank displacement, vectorless count, sit_tick); `"on"` serves
the augmented window and keeps recording. `promptworld divergence <world> [--json]`
aggregates per-agent/per-game-day evidence for the documented shadow→on gate decision
(FR-007) — the flip is an operator edit recorded on the board, never automatic.

**Transport.** The llm layer's `embedding` route kind is a valid route key outside the
cognition-kind registry (embedding is not a cognition; `Submit` rejects it).
`openaiCompat.Embed` posts to `endpoint + "/embeddings"`; `WarmEmbed` pins the model
resident via the Ollama-native `/api/embed` with `keep_alive: -1` (best-effort — non-
Ollama endpoints degrade to cold loads) at driver start and hourly. The anthropic
transport returns a typed unsupported error, and routing embedding to it fails config
validation at boot. Orchestrator `Embed` bypasses queue/breaker/estimator and serves
head-of-chain only, so one lineage never mixes model identities.

## Connections

- [[agent-mind]] — the prompt window this selector fills; the planner driver that
  triggers selection.
- [[sim-state-reducer]] — the reducer arms that attach vectors and situation state.
- [[event-types]] — `agent.memory_embedded`, `agent.situation_embedded`,
  `cog.memory_divergence`.
- [[llm-orchestrator]] — the `embedding` kind, `Embed`/`WarmEmbed` transport surface.
- [[world-save-directory]] — the `memory_relevance` manifest flag.
- [[executor]] — where memories are emitted (situated constructors, salience table).
- [[cli-promptworld]] — the `divergence` subcommand.
- Upstream design record: `specs/042-embedding-memory-retrieval/` and the research vault
  branch `research/Agent-Memory-Retrieval/`.

## Operational notes

- Absent `embedding` route in `llm.json` = subsystem off: memories land vectorless,
  selection stays legacy, one boot info line. Embedder transport failure = debounced
  `daemon.llm_warning` (kind `embedding-unavailable`); memories still land, vectorless
  forever (no backfill pass).
- Pin the fully tagged model id (`all-minilm:latest`) — a bare alias embeds fine but
  trips a persistent spurious provider-health `model-missing` warning that embed traffic
  cannot clear (TASK-102).
- Measured cost (SC-005): +0.4–3.9% wall-clock per game day at max speed post-041-merge
  (~8–9 coalesced embed calls); the 10% budget holds with headroom.
- Divergence records fire per enqueued plan job — on an uncalibrated rig run ≤16x (or
  calibrate first) or the planner never enqueues and no divergence lands.
- Replay proof: `TestEmbeddingReplayByteIdentical` / `TestOnWorldReplayByteIdentical`
  meter zero embedding endpoint calls during replay.
