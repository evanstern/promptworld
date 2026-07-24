---
title: Write-Time Importance Scoring
aliases: [Poignancy Rating, Salience Scoring]
tags: [importance, salience, scoring]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Memory-Retrieval]]", "[[Retrieval-Scoring-Functions]]", "[[Novelty-Gates-and-Habituation]]"]
---

# Write-Time Importance Scoring

How systems assign the stored "importance" number a memory carries from birth — the term
promptworld currently supplies from its hand-tuned salience table.

## Three generations of approach

**1. LLM-rated at write time (Park et al.).** Prompt the model to rate each observation 1–10
("1 = brushing teeth, 10 = a breakup"); store the integer; reuse it at every retrieval.
Documented costs: **one model call per memory write**, and **ratings drift across model
versions**, which makes it expensive and unstable for high-throughput agents
([Hindsight](https://hindsight.vectorize.io/blog/2026/05/21/agent-memory-consolidation),
[[_grounding]] §2).

**2. Computed from measurable signals.** Importance as a weighted mix of information
density, recency, and **access frequency**, sigmoid-normalized — no model call, fully
deterministic given the inputs
([EmergentMind](https://www.emergentmind.com/topics/memory-mechanisms-in-llm-based-agents),
[[_grounding]] §2).

**3. Learned value models.** Cognitively grounded multi-factor value models trained to
predict what is worth remembering ([arXiv:2606.12945](https://arxiv.org/pdf/2606.12945));
surveys report RL-tuned memory operations (A-MEM, Memory-R1) consistently beating static
heuristic pipelines on multi-hop and update-heavy tasks ([[_grounding]] §2).

## Importance as a trigger, not just a rank

In the generative-agents architecture the stored importance also drives a **threshold
trigger**: when the sum of importance over recent events exceeds ~**150**, a reflection pass
fires (2–3×/simulated day in practice). The threshold is tuned empirically — too high and
reflection never fires ([[_grounding]] §1). This is structurally the same pattern as a
salience value gating downstream machinery (promptworld's `GenerationBumpSalience`), except
integrated over recent events rather than tested per-event.

## Grounding

- [[_grounding]] §1 (poignancy scale, reflection threshold 150), §2 (costs of LLM rating;
  signal-based and learned alternatives)
- [Hindsight, The Consolidation Problem](https://hindsight.vectorize.io/blog/2026/05/21/agent-memory-consolidation)
- [Learning What to Remember, arXiv:2606.12945](https://arxiv.org/pdf/2606.12945)
