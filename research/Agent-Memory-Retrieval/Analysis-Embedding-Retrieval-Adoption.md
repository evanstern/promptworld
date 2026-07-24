---
title: Analysis — Should promptworld adopt embedding-based memory retrieval, and in what form?
aliases: [Embedding Retrieval Adoption]
tags: [analysis, agent-memory, embeddings]
type: analysis
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Memory-Retrieval]]", "[[Brief-and-Assumptions]]"]
---

# Analysis — Should promptworld adopt embedding-based memory retrieval, and in what form?

Given promptworld's constraints — deterministic replay, salience as a control-plane signal
with hard thresholds consumed synchronously in the reducer, a Go tick-based sim, and
village-scale agent counts ([[Brief-and-Assumptions]]) — which of the five candidate pieces
should be built (relevance in `SelectMemories`, open-ended memory text, novelty/habituation
gating, consolidation-time clustering, write-time kNN salience scoring), in what order, and
how is replay determinism satisfied?

## Verdict

**Yes — adopt embeddings, but as a read-time relevance term and a consolidation trigger,
never as a write-time replacement for the salience table.** Build in this order:

1. **Embed-at-emission infrastructure + a relevance term in `SelectMemories`** — the primary
   payoff, and the corpus's best-evidenced intervention.
2. **Open-ended memory text** — ship it *with or after* (1), because it is what makes the
   relevance term earn its keep; promptworld's situated where/why texts already vary enough
   to start.
3. **Consolidation-time clustering + habituation** — the RecMem/SAGE patterns, dropped into
   the nightly digest that already exists.
4. **Write-time kNN salience scoring — defer indefinitely.** The salience table stays, as
   seed truth and control-plane signal.

Determinism is satisfied by one rule: **embeddings are recorded at emission and replay never
re-embeds.** Vectors are payload data, like `describePlace` output; every score derived from
them is a pure function of recorded data.

## Reasoning

**Relevance is the missing third term, and the evidence for it is the strongest in the
corpus.** The canonical score is recency + importance + relevance, and the ablation —
run in Smallville, the closest published analog to promptworld — found each term
contributes and the composite beats pure cosine ([[Retrieval-Scoring-Functions]]).
promptworld already has two of the three terms (salience = importance, half-life decay =
recency); it is exactly one term short of the reference architecture. Meanwhile the cost
side collapses at promptworld's scale: per-agent stores of hundreds-to-thousands of
memories sit far below the brute-force/ANN crossover, so exact cosine scan is
sub-millisecond and **no vector database is needed at all**
([[Small-Scale-Vector-Search-and-Latency]]). A 384-dim model embeds in ~4 ms on CPU and is
callable from Go ([[Embedding-Determinism-and-Local-Inference]]). High evidence of benefit,
trivially cheap at this scale — that is the piece to build first.

**Determinism forces record-at-emission, and conveniently promptworld already works that
way.** Hosted embedding APIs return different vectors for identical input; local ONNX
inference is bit-reproducible only within a same-binary/same-hardware/single-thread
envelope, and no mainstream runtime promises cross-machine exactness
([[Embedding-Determinism-and-Local-Inference]]). So "pin a model and recompute on replay"
is not a safe strategy — but "stamp the derived value into the event and never re-derive"
is promptworld's existing doctrine (`describePlace`, origin stamping;
[[Brief-and-Assumptions]]). Concretely: an embedding event (or payload field) records each
memory's vector at emission; the cognition job records its query vector; the
relevance-augmented selection is then a **pure function over recorded floats** — exactly as
replayable as today's `SelectMemories`. Storage is the only cost: 384 float32 ≈ 1.5 KB per
memory (int8 quantization cuts that 4×), trivial at village scale.

**One reference-design feature must be explicitly rejected: recency-from-last-access.**
Park's recency term decays from the *last access* and so mutates state on read
([[Retrieval-Scoring-Functions]]) — hostile to a pure selection function and to replay.
Keep creation-time decay. If access-refreshing is ever wanted, accesses must become
recorded events; that is a real cost and nothing in the corpus shows the access-based
variant is load-bearing.

**Open-ended text is the enabler, not a separate bet.** Embeddings over fully templated
strings degenerate toward a type lookup — but promptworld's situated texts (place, why,
subject, tone) already vary within a type, and three memory kinds (gists, digests, omens)
are already free text ([[Brief-and-Assumptions]]). So the relevance term is not useless on
day one, and every step toward open-ended text (LLM-authored memory phrasing, richer
metadata in the string) increases its resolution. Sequence them together: (1) makes (2)
valuable, (2) makes (1) stronger.

**Consolidation is where clustering and habituation belong, and the slot already exists.**
The corpus's clear pattern is cheap-geometry-first: SAGE's density gate routes the obvious
cases and reserves the LLM for the ambiguous band; RecMem fires LLM consolidation only when
an embedding cluster forms ([[Novelty-Gates-and-Habituation]],
[[Consolidation-and-Clustering]]). promptworld's nightly digest is off the hot path,
latency-tolerant, and already licensed to reweigh (promotion boosts). Dropping
cluster-detection and dense-cluster down-weighting (habituation) there requires no new
architectural seams — consolidation outcomes are already events, so replay holds.

**Write-time kNN salience scoring loses on every axis it is measured on.** (a) Salience is
consumed synchronously in the reducer (`GenerationBumpSalience`) and gates gossip and
governance — a control-plane signal whose thresholds demand cross-type comparability and
write-time availability ([[Brief-and-Assumptions]]). (b) The corpus warns that per-write
model scoring is costly and drifts across model versions
([[Write-Time-Importance-Scoring]]), and neighbor-label propagation with self-inserted
labels is a feedback loop with no published evaluation ([[Agent-Memory-Retrieval]] open
questions). (c) The literature's own direction for importance is *measurable signals and
learned models*, not kNN over self-labels. The table is small, legible, deterministic, and
free; nothing gathered suggests replacing it pays. If its coverage ever strains, the
grounded upgrade path is signal-computed importance (density/access-frequency style), not
embedding neighbors.

**Where retrieval should live:** relevance scoring belongs on the mind side of the
sim/mind boundary — prompts are already assembled from recorded state there — leaving the
sim's pure `SelectMemories` intact as the fallback and test surface. The mind-side selector
is still a pure function (recorded memory vectors × recorded query vector), so the shared-
by-tests property survives; it just gains an input.

## Tensions & tradeoffs

- **The relevance term might just echo recency at village scale.** Days repeat; the current
  situation often resembles the recent past, so cosine may re-rank little. The Smallville
  ablation is the counter-evidence, but it measured believability in interviews, not
  behavioral divergence in a survival sim. This is the real risk of building (1) first —
  mitigate by instrumenting: log rank-divergence between salience×recency and the
  three-term score before letting it drive prompts.
- **Operational surface.** A local embedding model (ONNX runtime C bindings via hugot, or a
  llama.cpp sidecar) is a new moving part in a pure-Go project — version-pinning the model
  file becomes part of replay hygiene even with record-at-emission (old vectors and new
  vectors must at least share a space, or re-embedding events must be modeled).
- **Zep's benchmark lesson cuts slightly against dense similarity**: temporal structure
  layered over embeddings beat embeddings alone by 15 points
  ([[Memory-Systems-and-Game-NPC-Practice]]). promptworld's tick/subject/origin metadata is
  already strong temporal-relational structure; the honest reading is that relevance
  *complements* it, and a metadata-aware retrieval (filter by subject, then rank by cosine)
  may beat naive cosine over everything.
- **What the verdict gives up:** the "one embedding pipeline scores everything" elegance of
  the original proposal. Salience and relevance stay separate signals with separate
  mechanisms. The corpus says that separation is not a compromise but the actual shape of
  the systems that work ([[Retrieval-Scoring-Functions]], [[Novelty-Gates-and-Habituation]]).

## Confidence & open questions

**Moderately high** on the ordering and on record-at-emission as the determinism strategy —
both rest on multiple independent sources and on constraints verified in the codebase.
**Medium** on the magnitude of behavioral benefit from the relevance term at village scale;
no gathered source measures it in a survival-sim setting ([[Agent-Memory-Retrieval]] open
questions). Would change the verdict: evidence that rank-divergence between the two scoring
functions is negligible in practice (would demote (1) below (3)); a published evaluation of
kNN label propagation for importance that beats seed tables (would soften the "defer
indefinitely"); a bit-reproducible embedding runtime becoming mainstream (would relax
record-at-emission to pin-and-recompute).

## Basis

- [[_grounding]] — §1 (scoring, ablation), §3 (SAGE, habituation), §4 (RecMem,
  consolidation), §5 (Zep gap, NPC practice, latency), §6 (brute-force crossover),
  §7 (determinism), §8 (models, Go paths)
- [[Retrieval-Scoring-Functions]] · [[Write-Time-Importance-Scoring]] ·
  [[Novelty-Gates-and-Habituation]] · [[Consolidation-and-Clustering]] ·
  [[Memory-Systems-and-Game-NPC-Practice]] · [[Small-Scale-Vector-Search-and-Latency]] ·
  [[Embedding-Determinism-and-Local-Inference]] · [[Brief-and-Assumptions]]
