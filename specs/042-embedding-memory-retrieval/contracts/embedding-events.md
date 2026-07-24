# Contract: Embedding Events & Reducer Semantics

**Feature**: specs/042-embedding-memory-retrieval · applies to `internal/sim` (reducer,
whitelist), `internal/mind` (embedder driver), `internal/llm` (embedding kind).

## 1. Who may emit what

| Event | Emitter | Path | Reducer action |
|---|---|---|---|
| `agent.memory_added` | executor (unchanged) | deterministic reduce | build Memory; **stamp `Seq` = event seq** |
| `agent.memory_embedded` | mind embedder driver ONLY | `InjectSocial` (whitelist +) | attach `Vec`/`VecModel` to the `{agent, mem_seq}` memory; no-op if absent |
| `agent.situation_embedded` | mind embedder driver ONLY | `InjectSocial` (whitelist +) | set agent's `SitVec`/`SitVecModel`/`SitVecTick` |
| `cog.memory_divergence` | mind plan path ONLY | `InjectSocial` (whitelist +) | **no-op** (telemetry, cog.* class) |
| `daemon.llm_warning` | embedder driver on failure | existing channel | existing semantics (loud, non-fatal) |

The sim layer NEVER computes an embedding, and the reducer NEVER inspects vector
contents — copy-verbatim only (model-free sim doctrine, spec 030 lineage).

## 2. Embedder driver contract (`internal/mind/embedder.go`)

- Watches the absorbed event stream for `agent.memory_added`; for each, embeds the
  memory's **exact recorded text** (FR-011: no normalization beyond a fixed-byte
  truncation cap; the cap and its boundary are constants).
- At each `PlannerCadenceTicks` bucket edge, renders each live agent's situation string
  (deterministic template over replica state: time phase · position + place desc · worst
  needs · active intent verb + reason · nearby agent names), embeds it, injects
  `agent.situation_embedded`.
- Batching allowed; ordering within one agent MUST be emission-ordered (a memory's
  companion never lands before the memory itself — guaranteed by the door's event
  ordering).
- On transport failure: debounced `daemon.llm_warning`; NEVER retries into the tick path;
  NEVER blocks absorb/plan; skipped items stay vectorless forever (no backfill pass in
  this feature).
- Runs ONLY when the `embedding` route exists in llm.json. Absent route → driver not
  wired (vectorless world).

## 3. Determinism rules (the non-negotiables)

1. **Replay performs zero embedding computation.** All vectors enter state via recorded
   events; `recoverState`/`ReplayEvents` reproduce them byte-identically.
2. **Additive-omitempty only.** `Memory.Seq/Vec/VecModel`, `Agent.SitVec*`, and the
   `memory_relevance` world flag serialize with omitempty; pre-feature worlds,
   snapshots, and events are byte-identical under the new binary. No FormatVersion bump.
3. **Seq stamping is replay-stable**: the reducer stamps `Memory.Seq` from the event's
   store seq, which `CheckContiguity` guarantees gapless and identical in replay.
4. **Model identity travels with every vector** (`model` field); consumers MUST treat
   vectors with mismatched models as incomparable (neutral), never as garbage cosine.
5. **JSON float32 arrays** are the only vector wire/state format (no base64, no
   quantization in this feature).

## 4. llm embedding kind (`internal/llm`)

- `openai_compat` gains `Embed(ctx, texts []string) ([][]float32, model string, err)`
  posting `{"model", "input"}` to `endpoint + "/embeddings"` (OpenAI-compatible; Ollama
  serves this for embedding models).
- New route kind `embedding` in the kinds enum; validated like other kinds; **no default
  backfill** — a missing route means OFF, logged once at boot as info (not a warning
  spam), mirroring the "no llm.json → reflex-only" posture.
- The provider entry's `model` string is authoritative and is the string recorded as
  `model` on every emitted vector.
- Anthropic transport does NOT implement Embed (returns a typed unsupported error);
  routing an embedding kind to it is a config validation error at boot.
