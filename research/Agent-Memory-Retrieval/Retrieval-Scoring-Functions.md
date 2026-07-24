---
title: Retrieval Scoring Functions
aliases: [Recency-Importance-Relevance]
tags: [retrieval, scoring, generative-agents]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Memory-Retrieval]]", "[[Write-Time-Importance-Scoring]]", "[[Consolidation-and-Clustering]]"]
---

# Retrieval Scoring Functions

How agent-simulation systems decide *which* memories enter the prompt window. The canonical
form is the Stanford generative-agents score ([[_grounding]] §1).

## The three-term score

Every memory in the stream is scored on three normalized dimensions, combined as a weighted
sum (equal weights in the reference implementation):

```
score = α_recency · recency + α_importance · importance + α_relevance · relevance
```

| Term | Computed | When |
|---|---|---|
| Recency | exponential decay, factor **0.995/hour**, on last *access* time | read time |
| Importance | integer 1–10, LLM-rated "poignancy" | write time, stored |
| Relevance | **cosine similarity** of memory embedding vs a query embedding of the current situation | read time |

([Park et al. 2023](https://ar5iv.labs.arxiv.org/html/2304.03442), [[_grounding]] §1.)

Notable details:

- **Recency decays from last access, not creation** — retrieving a memory refreshes it, so
  frequently-recalled memories stay warm. This makes retrieval itself state-mutating.
- **Relevance is query-conditioned**: it does not exist as a stored property of the memory;
  it is recomputed against whatever the agent is currently perceiving/planning.
- The paper's ablation found each term contributes; the composite outperforms pure cosine
  similarity ([[_grounding]] §1).

## What retrieval draws from

Retrieval ranks over the union of raw observations, reflections (synthesized higher-order
memories, [[Consolidation-and-Clustering]]), and plans — all in one stream, all scored by the
same function ([[_grounding]] §1, §4).

## Variations in later systems

Later work keeps the multi-signal shape but changes where the signals come from: importance
from measurable statistics (information density, access frequency) instead of an LLM call
([[Write-Time-Importance-Scoring]]); relevance from temporal knowledge graphs (Zep) or
memory-to-memory links (A-MEM) layered over dense similarity
([[Memory-Systems-and-Game-NPC-Practice]], [[_grounding]] §5).

## Grounding

- [[_grounding]] §1 (score definition, decay constant, ablation), §4 (reflections in the
  same stream), §5 (variations)
- [Generative Agents, arXiv:2304.03442](https://ar5iv.labs.arxiv.org/html/2304.03442)
- [AgentPatterns memory-stream writeup](https://agentpatterns.ai/agent-design/generative-agents-memory-stream/)
