---
title: Small-Scale Vector Search and Latency
aliases: [Brute-Force kNN, Latency Budgets]
tags: [vector-search, knn, ann, latency]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Memory-Retrieval]]", "[[Memory-Systems-and-Game-NPC-Practice]]", "[[Embedding-Determinism-and-Local-Inference]]"]
---

# Small-Scale Vector Search and Latency

Whether a per-agent memory store (hundreds to low thousands of vectors) needs a vector
database at all, and what latency budgets production systems work to.

## Brute force vs ANN: the crossover

- **Exact brute-force kNN** (cosine against every stored vector) guarantees 100% recall and
  is "ideal where the dataset is relatively small (thousands to low millions of vectors) and
  dimensionality is moderate" ([Weaviate](https://weaviate.io/blog/vector-search-explained),
  [[_grounding]] §6).
- **ANN indexes** (HNSW etc.) become the preferred method at **hundreds of thousands of
  vectors**; brute-force latency grows linearly with store size
  ([Vectroid](https://www.vectroid.com/resources/knn-vs-ann-comprehensive-overview),
  [[_grounding]] §6).
- For **thousands of vectors** the sources place brute force squarely in the feasible zone:
  exact results, no index maintenance, no external database ([[_grounding]] §6). At 384
  dimensions (MiniLM/bge-small class, [[Embedding-Determinism-and-Local-Inference]]), a
  scan over a few thousand vectors is a few million multiply-adds — sub-millisecond on a
  modern CPU core; the documented linear-growth concern only bites orders of magnitude
  later.

## Production latency budgets

- Two-stage retrieval (fast approximate candidate pool in <10 ms, then a precise reranker)
  totals **40–80 ms**, vs 200 ms+ for full-precision single-stage search
  ([Supermemory](https://supermemory.ai/blog/latency-budgets-memory-retrieval),
  [[_grounding]] §5).
- A typical production configuration — 1536-dim embeddings, HNSW, top-k 20 — holds
  **sub-50 ms at multi-million scale** ([[_grounding]] §5).

These budgets are for million-scale shared stores; per-agent small-N stores sit far below
the regime where any of this machinery is needed ([[_grounding]] §6).

## Grounding

- [[_grounding]] §5 (latency budgets), §6 (brute force vs ANN)
- [Weaviate, Vector Search Explained](https://weaviate.io/blog/vector-search-explained)
- [Vectroid, KNNs vs ANNs](https://www.vectroid.com/resources/knn-vs-ann-comprehensive-overview)
- [Supermemory, Latency Budgets](https://supermemory.ai/blog/latency-budgets-memory-retrieval)
