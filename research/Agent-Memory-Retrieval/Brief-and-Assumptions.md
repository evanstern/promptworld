---
title: Brief and Assumptions
aliases: []
tags: [brief, agent-memory]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Memory-Retrieval]]"]
---

# Brief and Assumptions

## The request (restated)

Research **embedding-based memory retrieval for LLM agent simulations**, vis-à-vis the
promptworld use case. The user's framing, near-verbatim:

1. Promptworld today assigns episodic-memory importance from a **small hand-tuned salience
   table** (discrete event types → integer 1–10), and selects a working-memory window
   (`SelectMemories`) by **salience × recency decay** only.
2. The user proposes: expand event/memory text to be **open-ended (non-templated)** — or at
   least carry varying meta-information — and use an **embedding system** (RAG-style) so that
   a new memory is embedded, compared against previously recorded memories (cosine
   similarity / kNN), scored from its neighbors, and planted back in a vector store for
   future lookups. Over time similar memories cluster.
3. **Memory retrieval is the primary use-case** — embedding relevance at retrieval time
   (ranking which memories enter the prompt window), more than replacing the write-time
   salience score.

## Constraints from the codebase (current state, verified 2026-07-24)

- **Determinism / replay is load-bearing.** The sim layer is deterministic and model-free by
  doctrine (spec 030: origin stamped at emission, "no text inspection, no heuristics").
  Derived values are baked into event payloads at emission so replay is byte-identical and
  never recomputes.
- **Salience is a control-plane signal, not just a rating.** Hard thresholds consume it
  synchronously: `GenerationBumpSalience = 9` (interrupts cognition, checked in the
  reducer), `rumorMinSalience = 4` (gossip gating; belief confidence = salience × 10),
  governance evidence selection (max salience), and the `SelectMemories` window
  (salience halved per game-day of age; top K−2 plus 2 seeded serendipity picks; K = 10).
- **Consolidation already exists** as the layer licensed to reweigh/rewrite
  (`sim/consolidate.go` promotion boosts capped at 10; `mind/consolidate.go` LLM nightly
  day-gist).
- Three memory kinds are already open-ended (LLM-written): conversation gists, day
  digests, and Metatron omens. The rest are templated strings.
- Implementation language is **Go**; the sim is tick-based; agent counts are village-scale
  (tens, not thousands); per-agent memory streams grow over long runs.

## Research question (scoped)

What is known about embedding-based memory retrieval for LLM agent simulations —
retrieval scoring functions (recency + importance + relevance), open-ended memory text,
novelty/habituation via vector distance, consolidation-time clustering and reweighing,
and the determinism/reproducibility properties of embedding models and APIs — such that
a deterministic, replayable, tick-based sim could adopt them?

## Assumptions

- "Embedding system" means dense text embeddings + cosine kNN, not sparse/BM25 (though
  hybrid retrieval evidence is in scope as context).
- The vector store would be per-agent (each agent's private memory stream), small-N
  (hundreds to low thousands of memories per agent), so exhaustive kNN is plausible and
  ANN indexes may be unnecessary — this is a question to ground, not assume.
- Determinism could be satisfied either by (a) recording embedding outputs/scores into the
  event log at emission (replay reads, never recomputes), or (b) a pinned local embedding
  model that is bit-reproducible. Which is realistic is a grounding question.

## Open questions / ambiguities flagged

- Does retrieval relevance need to be replay-deterministic at all, if retrieval happens in
  the **mind** layer (which already makes non-deterministic LLM calls that are recorded as
  events)? The sim/mind boundary may absorb the determinism concern — needs a design
  decision, not research, but the research should establish what others do.
- How would relevance be queried — against the current situation/plan text? The "query" is
  a design choice downstream of this research.
- Whether write-time salience should *also* become learned (kNN label propagation from
  seed exemplars) is secondary; the primary use-case is retrieval ranking.
