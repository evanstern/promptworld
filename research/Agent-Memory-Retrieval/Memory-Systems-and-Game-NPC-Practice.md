---
title: Memory Systems and Game NPC Practice
aliases: [Systems Landscape]
tags: [systems, npc, games, landscape]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Memory-Retrieval]]", "[[Retrieval-Scoring-Functions]]", "[[Small-Scale-Vector-Search-and-Latency]]"]
---

# Memory Systems and Game NPC Practice

The current (2025–2026) landscape of agent-memory systems, and what game/NPC deployments
actually do.

## Taxonomy

[CoALA (arXiv:2309.02427)](https://arxiv.org/pdf/2309.02427) is the shared vocabulary:
**episodic** (instance-specific experience from earlier decision cycles), **semantic**
(abstracted facts), **procedural** (skills/code). Letta, Mem0, and LangChain build on it
([[_grounding]] §5).

## Major systems

| System | Architecture | Reported results |
|---|---|---|
| Letta (ex-MemGPT) | hierarchical core vs archival memory; context management as an OS problem | — |
| Mem0 | key-value memories extracted from conversations, dense-embedding retrieval | Recall@5 ≈ 0.5–0.6 on LongMemEval-S |
| Zep | temporal knowledge graph over dense retrieval | 63.8% vs Mem0's 49.0% on LongMemEval (GPT-4o) |
| A-MEM | dynamic memory-to-memory links, supersede detection | — |

The Zep-vs-Mem0 gap is attributed to temporal structure on top of embeddings — pure dense
similarity loses time-aware queries ([[_grounding]] §5).

## Game NPC practice

Production NPC systems keep **two separate vector stores per NPC** — conversation memory
(continuity with the player) and world knowledge (role-relevant facts) — each queried
independently by cosine similarity against the incoming input, top-k from each merged into
the prompt ([Fixed-Persona SLMs, arXiv:2511.10277](https://arxiv.org/html/2511.10277v1)).
Shipping games store conversation summaries as embeddings and retrieve by semantic relevance
rather than chronology ([[_grounding]] §5). Retrieval-latency practice is covered in
[[Small-Scale-Vector-Search-and-Latency]].

## Grounding

- [[_grounding]] §5 (all claims)
- [CoALA, arXiv:2309.02427](https://arxiv.org/pdf/2309.02427)
- [Fixed-Persona SLMs with Modular Memory, arXiv:2511.10277](https://arxiv.org/html/2511.10277v1)
