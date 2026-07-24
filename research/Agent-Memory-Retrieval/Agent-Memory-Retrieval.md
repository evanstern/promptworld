---
title: Agent-Memory-Retrieval
aliases: [Embedding Memory Retrieval, Memory Retrieval]
tags: [moc, agent-memory, embeddings, retrieval]
type: moc
created: 2026-07-24
updated: 2026-07-24
related: []
---

# Agent-Memory-Retrieval

Embedding-based memory retrieval for LLM agent simulations. Gathered to ground a design
question in promptworld: whether/how to move beyond the hand-tuned salience table and
salience×recency window selection toward open-ended memory text with embedding relevance at
retrieval time — under the sim's determinism/replay constraints ([[Brief-and-Assumptions]]).

## Scope

**In:** retrieval scoring functions (the generative-agents recency + importance + relevance
family); write-time importance scoring (LLM-rated, signal-computed, learned); novelty and
habituation via embedding distance (write-side gates); consolidation-time clustering,
reflection, and reweighting; the 2025–26 memory-systems landscape and game-NPC practice;
brute-force kNN vs ANN at per-agent scale and production latency budgets; determinism and
reproducibility of hosted and local embedding inference, including Go integration paths.
**Out:** any recommendation of what promptworld should build (analysis phase); general RAG
document-retrieval literature beyond what agent memory borrows; spatial/world-model memory
(that is the Agent-Mental-Maps branch's topic — no cross-links).

## What is known

The canonical retrieval function scores every memory as a weighted sum of **recency**
(exponential decay from last access), **importance** (a stored write-time score), and
**relevance** (query-conditioned cosine similarity) — each term contributes in ablation
([[Retrieval-Scoring-Functions]]). The stored importance term has three generations:
LLM-rated at write time (costly, drifts across model versions), computed from measurable
signals (density, access frequency), and learned value models; importance also serves as an
integrated threshold trigger for reflection (~150 summed points)
([[Write-Time-Importance-Scoring]]). Embedding *distance* to the existing store measures
novelty/typicality — not importance — and the worked systems use it as a cheap write-side
gate that reserves LLM judgment for the ambiguous band (SAGE's vMF density gate), with
habituation as the classical frame for down-weighting dense clusters of repeats
([[Novelty-Gates-and-Habituation]]). Reweighting and rewriting happen at consolidation:
reflection synthesizes higher-order memories back into the same stream, and RecMem-style
systems trigger consolidation *from* embedding-cluster formation
([[Consolidation-and-Clustering]]). The systems landscape (Letta, Mem0, Zep, A-MEM; CoALA
taxonomy) shows temporal structure layered on dense retrieval winning benchmarks, and game
NPCs shipping per-NPC dual vector stores queried by cosine similarity
([[Memory-Systems-and-Game-NPC-Practice]]). At per-agent scale (thousands of vectors),
brute-force exact kNN is squarely feasible with no vector database; production latency
budgets (40–80 ms two-stage) address million-scale stores
([[Small-Scale-Vector-Search-and-Latency]]). Hosted embedding APIs are not deterministic;
local ONNX inference is reproducible only within a same-binary/same-hardware/single-thread
envelope; small models (384-dim MiniLM/bge class) embed in ~4 ms on CPU and are runnable
from Go ([[Embedding-Determinism-and-Local-Inference]]).

## Notes

- [[Brief-and-Assumptions]] — the request, promptworld's constraints, scoped question
- [[Retrieval-Scoring-Functions]] — recency + importance + relevance; decay constants; ablation
- [[Write-Time-Importance-Scoring]] — LLM poignancy vs computed signals vs learned models; threshold triggers
- [[Novelty-Gates-and-Habituation]] — what embedding distance measures; SAGE's density gate; habituation
- [[Consolidation-and-Clustering]] — reflection, RecMem's cluster-triggered consolidation, supersede/staleness
- [[Memory-Systems-and-Game-NPC-Practice]] — CoALA taxonomy; Letta/Mem0/Zep/A-MEM; NPC dual stores
- [[Small-Scale-Vector-Search-and-Latency]] — brute force vs ANN crossover; production latency budgets
- [[Embedding-Determinism-and-Local-Inference]] — API nondeterminism; ONNX envelope; models; Go paths

## Analyses

- [[Analysis-Embedding-Retrieval-Adoption]] — should promptworld adopt embedding-based
  memory retrieval, and in what form? (Verdict: yes — read-time relevance + consolidation
  trigger, recorded at emission; never a write-time salience replacement.)
  - Rendered: `embedding-retrieval-adoption-briefing.html` —
    [live page](https://claude.ai/code/artifact/cc1f9eb9-4357-497a-904d-acaefc1ed1e9)

## Open questions

- Reported per-write embedding latencies are for MiniLM-class models on server CPUs; no
  source measured the marginal cost inside a tick-based Go sim loop specifically.
- No source directly addresses replaying a sim whose retrieval used embeddings — the
  record-at-emission vs pinned-model tradeoff is undocumented territory.
- kNN **label propagation** for importance (scoring a new memory from its neighbors' stored
  scores) is implied by the learned-value-model line but no gathered source evaluates it
  directly against seed-exemplar tables.
- Whether relevance-weighted retrieval measurably changes *behavior* in village-scale sims
  (vs chat assistants) — the generative-agents ablation is the closest evidence.

## Grounding

- [[_grounding]] — web-search fan-out, 13 searches across 2 rounds, 2026-07-24
- [Generative Agents, arXiv:2304.03442](https://ar5iv.labs.arxiv.org/html/2304.03442)
- [SAGE, arXiv:2605.30711](https://arxiv.org/abs/2605.30711)
- [RecMem, arXiv:2605.16045](https://arxiv.org/html/2605.16045v1)
