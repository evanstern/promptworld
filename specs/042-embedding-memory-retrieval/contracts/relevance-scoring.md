# Contract: Relevance Scoring & Mode Gating

**Feature**: specs/042-embedding-memory-retrieval · applies to `internal/sim/memory.go`
(pure scoring) and `internal/mind` (mode policy).

## 1. The three-term score (`SelectMemoriesRelevant`)

Pure function; signature-compatible sibling of `SelectMemories`:

```
SelectMemoriesRelevant(a *Agent, seed uint64, agentIdx int, tick int64, k int,
                       query []float32, queryModel string) []Memory
```

Per-memory score:

```
decayed  = float(m.Salience) halved per whole game-day of (tick - m.Tick)   // EXACTLY today's weight
sal01    = decayed / MaxSalience                                            // [0,1]
rel01    = (cosine(m.Vec, query) + 1) / 2   if m.Vec != nil && m.VecModel == queryModel
         = 0.5 (neutral)                     otherwise                       // vectorless / cross-model
score    = sal01 + rel01                                                     // equal weight (reference default)
```

- **Ties**: newer `Tick` wins (unchanged from `SelectMemories`).
- **Recency**: lives ONLY inside `decayed`, from `m.Tick` (creation). Selection mutates
  NOTHING — no last-access stamps, no counters (FR-004, guardrail 1).
- **Serendipity tail**: the top `k−2` come from `score`; the 2 seeded old-half picks are
  byte-for-byte today's algorithm (same rng stream key `"serendipity"`, same cadence
  bucket) so the tail's replay behavior is unchanged.
- **Cosine**: sequential float64 accumulation over float32 inputs, fixed index order —
  deterministic on all platforms; zero-magnitude vectors → neutral (0.5).
- **Isolation**: the only memory source is `a.Memories` (FR-005). No global/store access.

## 2. Fallback ladder (all deterministic)

| Condition | Behavior |
|---|---|
| `query == nil` (no situation vector recorded yet) | return `SelectMemories(...)` — legacy, exactly |
| memory vectorless / model mismatch | that memory scores with neutral `rel01 = 0.5` |
| `n <= k` | all memories, reverse-chronological (unchanged) |

Neutral = 0.5 (the cosine midpoint) so vectorless memories are neither promoted nor
punished relative to vectored ones — only genuine similarity signal moves ranks.

## 3. Mode gating (mind side)

| `world.json memory_relevance` | Prompt window | `cog.memory_divergence` |
|---|---|---|
| `""` (default) | `SelectMemories` | not emitted |
| `"shadow"` | `SelectMemories` (legacy — behavior MUST be bit-identical to `""` for agents) | emitted per plan job |
| `"on"` | `SelectMemoriesRelevant` | emitted per plan job |

- Applies to both mind prompt paths: planner window (prompt.go) and conversation scene
  snapshot (convo.go, k=5). Scribe soul.md, TUI, and consolidation are out of scope and
  unchanged.
- **Shadow-mode invariant (testable)**: with `"shadow"`, the bytes of every prompt are
  identical to `""` mode on the same recorded world. Divergence events are additional
  recorded telemetry only.
- **US2→US3 gate (FR-007)**: flipping to `"on"` requires a divergence summary spanning
  ≥ 1 full game day (SC-003) and a recorded threshold decision on TASK-98. The summary
  is computed from `cog.memory_divergence` events (auditable: seqs → memory texts).
  Suggested primary metric: mean overlap@K and the share of selections with ≥1
  relevance-promoted memory; the decision note must state the numbers and the call.

## 4. Success-criteria hooks

- SC-004 scenario test: construct an agent with an old low-salience memory whose text is
  semantically close to a crafted situation; assert it enters the `"on"` window and is
  absent from the legacy window.
- SC-006 isolation test: two agents with near-identical memories; mutate agent A's
  memories; assert agent B's selection output (all modes) is byte-identical before/after.
