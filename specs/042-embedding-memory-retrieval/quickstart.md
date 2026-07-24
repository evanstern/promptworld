# Quickstart: Embedding Memory Retrieval — validation guide

**Feature**: specs/042-embedding-memory-retrieval. Proves the three user stories
end-to-end. Details: [data-model.md](./data-model.md), [contracts/](./contracts/).

## Prerequisites

- Ollama running locally with the pinned embedding model:
  `ollama pull all-minilm && curl -s localhost:11434/v1/embeddings -d '{"model":"all-minilm:latest","input":"hello"}' | head -c 120`
- **Model id in llm.json must be the fully tagged `all-minilm:latest`** (as `ollama list`
  prints it), not bare `all-minilm` — the alias embeds fine, but the provider-health
  preflight compares ids and raises a persistent spurious `model-missing` warning that
  embed traffic cannot clear (tracked as a future finding).
- A test world save dir; `llm.json` with the `embedding` provider + route added
  (see docs/llm-providers.md); `world.json` with `"memory_relevance": "shadow"`.
- **Speed**: run at ≤16x on an uncalibrated rig, or `promptworld calibrate` first —
  above the suppression threshold the planner never enqueues, and divergence records
  fire per enqueued plan job (at 32x uncalibrated you'd see zero divergence events).

## US1 — vectors at emission, replay never re-embeds

```sh
go test ./internal/sim -run 'Memory' -count=1          # unit: Seq stamping, attach, fallbacks
# run the daemon against the test world for a few game hours, then:
sqlite3 <save>/world.db "select count(*) from events where type='agent.memory_added';"
sqlite3 <save>/world.db "select count(*) from events where type='agent.memory_embedded';"
```

**Expected**: embedded count converges to added count (SC-002); each embedded payload
carries `vec` (384 floats) + `model`.

**Replay byte-identity (SC-001)**: `go test ./internal/daemon -run Replay -count=1` —
the harness is httptest-stubbed and its endpoint-call meter asserts **zero embedding
calls during replay** (`TestEmbeddingReplayByteIdentical`, `TestOnWorldReplayByteIdentical`);
the manual equivalent on a real world is a daemon stop→restart (recovery replay) with
no new embed traffic.

**Loud failure (US1-S3)**: stop Ollama, run the world, emit memories.
**Expected**: memories land vectorless, one debounced `daemon.llm_warning` in the feed,
no tick stall.

## US2 — shadow mode changes nothing, records divergence

```sh
go test ./internal/mind -run 'Shadow' -count=1
# shadow invariant: same seed, memory_relevance "" vs "shadow" → identical agent behavior
sqlite3 <save>/world.db "select count(*) from events where type='cog.memory_divergence';"
```

**Expected**: divergence events per plan job; prompts byte-identical to off-mode
(the shadow invariant test asserts this). After ≥1 full game day, produce the summary
with **`promptworld divergence <world> [--json]`** (mean overlap@K, promoted-memory
share, displacement, per agent/game-day) — the recorded input to the US3 gate decision
(FR-007, SC-003).

## US3 — relevance shapes the window (after the gate decision)

Set `"memory_relevance": "on"`.

```sh
go test ./internal/sim -run 'RelevancePromoted|Isolation' -count=1
```

**Expected**:
- SC-004: the scenario test's old, low-salience, situation-matching memory is in the
  `on` window and provably absent from the legacy window.
- SC-006: isolation test — altering agent A's memories leaves agent B's selection
  byte-identical.
- Replay of an `on` world remains byte-identical (selection is pure over recorded
  vectors).

## Throughput check (SC-005)

Compare wall-clock per game-day, embedding on vs off, same seed/world:
**expected within 10%** (embeds are fully off the tick path; expect ≈0 delta).

## Done-ness beyond code (TASK-98 ACs)

- Divergence threshold decision recorded on TASK-98 (AC #4 / FR-007).
- `/grounding-wiki:wiki-update` re-pins memory/prompt/llm notes; player-docs freshness.
- **codebase-to-course run on this feature before the PR ships (AC #7).**
