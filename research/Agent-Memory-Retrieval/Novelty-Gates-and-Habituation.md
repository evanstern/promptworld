---
title: Novelty Gates and Habituation
aliases: [Surprise-Gated Memory, Embedding Novelty]
tags: [novelty, habituation, surprise, embeddings]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Memory-Retrieval]]", "[[Write-Time-Importance-Scoring]]", "[[Consolidation-and-Clustering]]"]
---

# Novelty Gates and Habituation

What embedding *distance* can and cannot tell you about a new memory. The literature is
consistent: distance-to-neighbors measures **novelty/typicality**, which is a different
signal from **importance** ([[Write-Time-Importance-Scoring]]) and from query-time
**relevance** ([[Retrieval-Scoring-Functions]]).

## Novelty vs surprise vs importance

- **Novelty**: not previously encountered — the new item embeds far from everything stored.
- **Surprise**: violates an expectation — prediction error against a learned model; used as
  intrinsic salience in RL and robotics since curiosity-driven controllers.
- Both stem from observation-vs-model mismatch, but they are distinguished in the
  literature ([[_grounding]] §3). Neither equals importance: a tenth occurrence of a
  life-threatening event is *less novel* than the first but not less important.

## SAGE: the density-gate design

[SAGE (arXiv:2605.30711)](https://arxiv.org/abs/2605.30711) is the clearest worked example
of write-side embedding gating: each candidate memory is scored by a **von Mises–Fisher
density estimator over the existing store's embeddings** — how crowded its neighborhood on
the unit sphere is — with an adaptive threshold tracking the store's geometry. Routing is
three-way: clearly novel → ADD, clearly redundant → NOOP, uncertain → escalate to an LLM
merge. Reported effect: ~3.4× lower add-phase API cost and 2.5× lower latency vs Mem0;
16–18% of LLM calls skipped as a drop-in gate for A-MEM ([[_grounding]] §3). The pattern:
**cheap geometry handles the clear cases; the model is reserved for the ambiguous band.**

## Habituation

Habituation — suppressing response to repeated neutral stimuli so attention goes to salient
novel ones — is the classical biological mechanism, computationally realized as early as
habituating self-organizing maps on robots (novelty = winning neuron's distance from
recently-fired neighborhoods, thresholded) ([Marsland et al.](https://arxiv.org/pdf/cs/0006007),
[[_grounding]] §3). Surprise-gated episodic memory in robotics stores an episode **only**
when unexpectedness is high ([Worth Remembering](https://arxiv.org/pdf/2606.03787)).

An implication documented in this literature: a dense cluster of near-identical stored
memories is the *signature of a habituated (routine) event class* — density can therefore
be read as a down-weighting signal for repeats, which is the inverse of reading proximity
as importance ([[_grounding]] §3).

## Grounding

- [[_grounding]] §3 (all claims)
- [SAGE, arXiv:2605.30711](https://arxiv.org/abs/2605.30711)
- [Worth Remembering, arXiv:2606.03787](https://arxiv.org/pdf/2606.03787)
- [Marsland et al., arXiv:cs/0006007](https://arxiv.org/pdf/cs/0006007)
