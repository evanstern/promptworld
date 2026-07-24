# Data Model: Embedding Memory Retrieval

**Feature**: specs/042-embedding-memory-retrieval · **Date**: 2026-07-24
**Rule of the model**: every addition is omitempty-additive — pre-feature events,
snapshots, and worlds round-trip byte-identically; no FormatVersion bump (spec-019/030
precedent, `internal/sim/agents.go:159-178`).

## Extended entities

### Memory (`internal/sim/agents.go`)

| Field | Type | New? | Semantics |
|---|---|---|---|
| Text, Salience, Tick, Subject, Tone, Where, Why, Conv, Origin | (existing) | — | unchanged (FR-008) |
| `Seq` | `int64` `json:"seq,omitempty"` | NEW | store seq of the emitting `agent.memory_added` event; stamped by the reducer at apply time; the memory's stable identity for companion events (research D4) |
| `Vec` | `[]float32` `json:"vec,omitempty"` | NEW | embedding of the memory text; attached by the reducer from `agent.memory_embedded`; nil = vectorless (neutral relevance) |
| `VecModel` | `string` `json:"vec_model,omitempty"` | NEW | identity of the producing model (provider `model` string); comparability boundary (FR-009) |

**Invariants**: `Vec`/`VecModel` are set together or not at all. Reducer never computes
either — copy-verbatim only. `Seq` is deterministic in replay (seq is part of the stream).

### Agent (no changes)

Per-agent isolation (FR-005) is structural: vectors live on `Agent.Memories` entries;
no shared store exists.

## New event payloads (all JSON, `internal/sim`)

### `agent.memory_embedded` — reducer-armed companion

| Field | Type | Semantics |
|---|---|---|
| `agent` | int | owner agent index |
| `mem_seq` | int64 | `Seq` of the target memory's `agent.memory_added` event |
| `vec` | []float32 | 384-dim class vector, JSON array |
| `model` | string | producing model identity |

Reducer: find `Memories[i].Seq == mem_seq` for `agent`; attach `Vec`/`VecModel`;
**no-op if not found** (agent died / memory consolidated away). Emitted only by the
mind-side embedder driver via `InjectSocial` (whitelisted).

### `agent.situation_embedded` — reducer-armed, per-agent rolling query vector

| Field | Type | Semantics |
|---|---|---|
| `agent` | int | agent index |
| `tick` | int64 | tick the situation text was rendered at |
| `text` | string | the deterministic situation string (audit surface for divergence) |
| `vec` | []float32 | its embedding |
| `model` | string | producing model identity |

Reducer: store as the agent's current situation vector (`Agent.SitVec/SitVecModel/
SitVecTick`, omitempty additive). Selection uses the latest at-or-before the selection
tick; absent → legacy fallback. Cadence: one per agent per `PlannerCadenceTicks` bucket
(only while the embedder runs; gaps are legal and handled by fallback).

### `cog.memory_divergence` — reducer NO-OP telemetry (whitelisted like other `cog.*`)

| Field | Type | Semantics |
|---|---|---|
| `agent` | int | agent index |
| `tick` | int64 | selection tick |
| `mode` | string | `"shadow"` or `"on"` at time of recording |
| `legacy` | []int64 | legacy window as memory `Seq`s, window order |
| `augmented` | []int64 | three-term window as memory `Seq`s, window order |
| `overlap` | int | count of seqs in both windows |
| `displacement` | int | sum of rank distance for shared members |
| `vectorless` | int | candidates lacking a comparable vector |
| `sit_tick` | int64 | tick of the situation vector used (0 = none → rankings identical by definition) |

## Configuration

### `world.json` (`internal/world/world.go`)

| Field | Type | Semantics |
|---|---|---|
| `memory_relevance` | string, omitempty | `""` (off, default) · `"shadow"` (record divergence, prompts get legacy window) · `"on"` (prompts get augmented window; divergence still recorded). US2→US3 flip is an operator edit gated on the documented threshold decision (FR-007) |

### `llm.json` (`internal/llm/config.go`)

- New route **kind `embedding`** + a provider entry (transport `openai_compat`, endpoint
  = local Ollama, model = pinned embedding model; default recommendation `all-minilm`).
- **Absent route = subsystem off** (vectorless world; no boot error, no backfill —
  deliberate deviation from the warn-backfill pattern because absence is the feature
  switch, mirroring "no llm.json → reflex-only").
- The provider `model` string **is** the recorded `model` identity on every vector.

## State-transition summary

```
memory emitted (executor, deterministic)          → Memory{Seq, Vec:nil}
  └─ embedder driver observes memory_added (async)
       ├─ ok    → InjectSocial(agent.memory_embedded)  → reducer attaches Vec/VecModel
       └─ fail  → daemon.llm_warning (debounced); memory stays vectorless

planner cadence bucket edge (embedder, async)     → render situation text (deterministic
  from replica) → embed → InjectSocial(agent.situation_embedded) → reducer stores SitVec

plan job (mind, absorb goroutine)                 → mode?
  ""       → SelectMemories (unchanged)
  "shadow" → both rankings; prompt ← legacy; emit cog.memory_divergence
  "on"     → prompt ← SelectMemoriesRelevant; emit cog.memory_divergence
```

## Validation rules

- Replay applies only recorded events — zero embed calls (FR-002; SC-001 via existing
  replay harness).
- Cross-model guard: cosine contributes only when `Memory.VecModel == situation model`;
  else neutral (FR-009).
- Vectorless memory or absent situation vector → neutral relevance / legacy fallback,
  deterministically (edge cases; FR-010).
- Selection never mutates memory state; recency decays from `Memory.Tick` (creation)
  only (FR-004).
- Text-to-embedding preparation is deterministic: exact memory text as recorded;
  truncation (if the model's input cap requires it) at a fixed byte boundary (FR-011).
