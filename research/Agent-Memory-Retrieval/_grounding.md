---
title: Agent-Memory-Retrieval — Grounding
aliases: []
tags: [grounding]
type: source
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Memory-Retrieval]]"]
---

# Agent-Memory-Retrieval — Grounding

> Source-of-truth artifact. This is the raw, cited output of a research pass (the `deep-research`
> skill, or a direct web-search fan-out). Keep it close to verbatim — do not editorialize, prune,
> or draw conclusions here. Knowledge notes and analyses cite *into* this file.

**Research question:** What is known about embedding-based memory retrieval for LLM agent
simulations — retrieval scoring (recency + importance + relevance), open-ended memory text,
novelty/habituation via vector distance, consolidation-time clustering and reweighing, and the
determinism/reproducibility properties of embedding models and APIs — such that a deterministic,
replayable, tick-based sim could adopt them?
**Method:** web-search fan-out (13 parallel searches across 2 rounds) · 2026-07-24

---

## §1 The generative-agents retrieval function (Park et al. 2023)

The canonical architecture is the Stanford "Generative Agents" memory stream
([Park et al. 2023, arXiv:2304.03442](https://ar5iv.labs.arxiv.org/html/2304.03442)):

- The **memory stream** records events the agent perceives as natural-language descriptions
  with an event timestamp and a most-recent-access timestamp.
- Retrieval scores every memory on three **normalized** dimensions and takes the top-ranked
  to fit the prompt context:
  - **Recency** — exponential decay on the most recent *access* (not creation): decay factor
    **0.995 per hour** in the reference implementation.
  - **Importance** — an integer score assigned by prompting the language model at write time
    ("poignancy," 1–10, where 1 is brushing teeth and 10 is a breakup).
  - **Relevance** — **cosine similarity between the memory's embedding vector and a query
    embedding** derived from the current situation.
- The composite is a weighted sum, equal weights in the paper:
  `score = α_recency·recency + α_importance·importance + α_relevance·relevance`
  ([paper HTML](https://dl.acm.org/doi/fullHtml/10.1145/3586183.3606763);
  [Ruder newsletter summary](https://newsletter.ruder.io/p/generative-agents-forums-for-foundation);
  [AgentPatterns writeup](https://agentpatterns.ai/agent-design/generative-agents-memory-stream/)).
- The paper's ablation found all three components contribute; the multi-signal score
  outperforms pure cosine similarity
  ([Lukyanenko review](https://artgor.medium.com/paper-review-generative-agents-interactive-simulacra-of-human-behavior-cc5f8294b4ac)).

**Reflection trigger:** reflections (higher-order synthesized memories) are generated **when
the sum of importance scores of the latest perceived events exceeds a threshold — 150 in the
implementation** — which in practice fired roughly two or three times per simulated day.
Reflection takes the ~100 most recent memories, asks the LLM for salient questions, retrieves
against those questions, and writes synthesized insights back into the same stream, where they
are retrieved alongside raw observations
([paper](https://ar5iv.labs.arxiv.org/html/2304.03442);
[AgentPatterns](https://agentpatterns.ai/agent-design/generative-agents-memory-stream/) — which
also notes the threshold must be tuned empirically: too high and reflection never fires).

## §2 Write-time importance scoring: LLM ratings vs heuristics vs learned models

- The Park-style pattern — **LLM rates each observation 1–10 at write time, the score is
  stored and reused at retrieval** — works but "adds a model call per write and the ratings
  drift across model versions, making it expensive for high-throughput agents"
  ([Hindsight/Vectorize on consolidation](https://hindsight.vectorize.io/blog/2026/05/21/agent-memory-consolidation)).
- Later systems compute importance from **measurable signals instead of a one-shot LLM
  rating**: e.g. a weighted mix of information density, recency, and **access frequency**,
  normalized with a sigmoid
  ([Memory Mechanisms in LLM Agents, EmergentMind](https://www.emergentmind.com/topics/memory-mechanisms-in-llm-based-agents)).
- A 2026 line of work builds **cognitively grounded multi-factor value models** for what to
  remember — learned scoring over several factors rather than a single hand-set scale
  ([Learning What to Remember, arXiv:2606.12945](https://arxiv.org/pdf/2606.12945)).
- Surveys report a trend from fixed heuristics toward **learned/RL-tuned memory operations**
  (A-MEM, Memory-R1), which "consistently outperform static or heuristic pipelines,
  particularly for multi-hop reasoning and update-intensive tasks"
  ([Memory for Autonomous LLM Agents survey](https://arxiv.org/html/2603.07670v1);
  [Rethinking Memory in LLM Agents, arXiv:2505.00675](https://arxiv.org/pdf/2505.00675)).

## §3 Novelty, surprise, and habituation via embedding distance

- **SAGE (novelty gate)** ([arXiv:2605.30711](https://arxiv.org/abs/2605.30711)): write-side
  control for agent memory. Scores each candidate fact with a **von Mises–Fisher density
  estimator over the existing memory embeddings** (i.e., directional density on the unit
  sphere — how "crowded" the neighborhood is) and routes with an adaptive threshold that
  tracks the store's geometry: clearly novel → ADD, clearly redundant → NOOP, uncertain →
  LLM merge step. Cuts add-phase API cost ~3.4× and latency ~2.5× vs Mem0 on GPT-4o-mini;
  as a drop-in gate for A-MEM it skips 16–18% of LLM calls with minimal quality change.
  Code: [github.com/swang1024/SAGE](https://github.com/swang1024/SAGE).
- **Surprise-gated episodic memory** in robotics stores an episode only when prediction
  error / unexpectedness is high ([Worth Remembering, arXiv:2606.03787](https://arxiv.org/pdf/2606.03787));
  prediction-error-as-salience goes back to curiosity-driven intrinsic motivation in RL
  ([survey](https://arxiv.org/pdf/2407.21338)).
- **Habituation** is the classical mechanism: "a form of simple memory that suppresses
  response to repeated, neutral stimuli … critical in guiding attention toward the most
  salient and novel features." Implemented computationally as early as habituating SOMs on
  mobile robots — the distance of the winning neuron from recently-fired neighborhoods
  counted as novelty beyond a threshold
  ([Marsland et al., arXiv:cs/0006007](https://arxiv.org/pdf/cs/0006007);
  [novelty detection overview](https://en.wikipedia.org/wiki/Novelty_detection)).
- The literature distinguishes **novelty** (not previously encountered → far from everything
  stored) from **surprise** (violates an expectation), though both stem from
  observation-vs-model mismatch
  ([Joint Agent Memory and Exploration Learning, arXiv:2606.01528](https://arxiv.org/html/2606.01528)).

## §4 Consolidation: reflection, clustering, summarization, reweighting

- The dominant pattern: agents accumulate time-stamped observations in a persistent log,
  **periodically run a reflection step that clusters related observations and synthesizes
  higher-order insights via LLM prompting**, then retrieve a recency/importance/relevance
  mixture of observations + reflections + plans
  ([EmergentMind topic page](https://www.emergentmind.com/topics/memory-mechanisms-in-llm-based-agents-c6936a2e-2296-46de-b469-040d6767712a)).
- **RecMem** ([arXiv:2605.16045](https://arxiv.org/html/2605.16045v1)): a "subconscious
  memory layer" buffers raw interactions **via lightweight embeddings**; LLM-based
  consolidation runs **only when an incoming item finds a sufficient number of semantically
  similar prior items** (recurrence-driven), extracting episodic summaries + semantic facts
  from the cluster. Explicitly motivated by reducing per-write LLM cost in long-running
  agents.
- **CLAG** ([arXiv:2603.15421](https://arxiv.org/pdf/2603.15421)): agent-driven clustering
  as the memory-organization primitive for small-model agents.
- Summarization-based consolidation (MemoryBank, ChatGPT-RSum, hierarchical/reflective
  chunking around subgoals) is the other common family
  ([Rethinking Memory survey](https://arxiv.org/pdf/2505.00675);
  [apxml course chapter](https://apxml.com/courses/agentic-llm-memory-architectures/chapter-3-designing-memory-systems/memory-consolidation-summarization)).
- The consolidation problem framed as its own design axis — what to merge, supersede, decay —
  in [Hindsight's consolidation essay](https://hindsight.vectorize.io/blog/2026/05/21/agent-memory-consolidation)
  and staleness/supersede detection in **A-MEM** (dynamic links between memory entries)
  ([Awesome-Agent-Memory list](https://github.com/TeleAI-UAGI/Awesome-Agent-Memory)).

## §5 Systems landscape and game-NPC practice

- **CoALA** taxonomy (episodic / semantic / procedural memory) is the shared vocabulary;
  Letta, Mem0, and LangChain build on it
  ([Cognitive Architectures for Language Agents, arXiv:2309.02427](https://arxiv.org/pdf/2309.02427)).
- **Letta (ex-MemGPT)**: hierarchical core-vs-archival memory, context management as an OS
  problem. **Mem0**: key-value memories extracted from conversations, dense-embedding
  retrieval (Recall@5 ≈ 0.5–0.6 on LongMemEval-S). **Zep**: temporal knowledge graph over
  dense retrieval; 63.8% vs Mem0's 49.0% on LongMemEval with GPT-4o — the gap attributed to
  temporal structure
  ([MemMachine paper's related-work numbers, arXiv:2604.04853](https://arxiv.org/html/2604.04853v1);
  [Awesome-Agent-Memory](https://github.com/TeleAI-UAGI/Awesome-Agent-Memory)).
- **Game NPCs**: production systems keep **two separate vector stores per NPC** —
  conversation memory (player-interaction continuity) and world knowledge (role-relevant
  facts) — each queried independently by cosine similarity against the incoming player
  input, top-k from each merged into the prompt
  ([Fixed-Persona SLMs with Modular Memory, arXiv:2511.10277](https://arxiv.org/html/2511.10277v1)).
  Games in the wild store conversation summaries as embeddings and retrieve by semantic
  relevance rather than chronology
  ([Wanderfolk roundup](https://wanderfolk.ai/games-where-npcs-remember-you/)).
- **Latency budgets** (production guidance, May 2026): two-stage retrieval — fast approximate
  search to a candidate pool in <10 ms, then a precise reranker — totals **40–80 ms** vs
  200 ms+ for full-precision search; a typical setup (1536-dim, HNSW, top-k 20) holds
  **sub-50 ms at multi-million scale**
  ([Supermemory latency-budget post](https://supermemory.ai/blog/latency-budgets-memory-retrieval)).

## §6 Vector search at small scale: brute force vs ANN

- Brute-force exact kNN computes similarity against every stored vector; it guarantees 100%
  recall and "is ideal where the dataset is relatively small (thousands to low millions of
  vectors) and dimensionality is moderate"
  ([Weaviate explainer](https://weaviate.io/blog/vector-search-explained);
  [Ozkaya overview](https://mehmetozkaya.medium.com/vector-search-algorithms-knn-ann-and-disk-ann-803c620bf73e)).
- ANN indexes (HNSW etc.) become the preferred method "when a dataset reaches hundreds of
  thousands of vectors"; latency of brute force grows linearly with dataset size
  ([Vectroid KNN-vs-ANN overview](https://www.vectroid.com/resources/knn-vs-ann-comprehensive-overview);
  [SQLite.ai intro](https://blog.sqlite.ai/intro-to-vector-search-nearest-neighbor-search-algorithms-for-rag-applications)).
- For datasets of **thousands of vectors, brute-force kNN is squarely in the feasible zone**:
  exact results, no index maintenance, no external database
  ([Vectroid](https://www.vectroid.com/resources/knn-vs-ann-comprehensive-overview)).

## §7 Determinism and reproducibility of embeddings

- **Hosted embedding APIs are not reliably deterministic.** OpenAI's `text-embedding-ada-002`
  and successors return **slightly different vectors for identical input across calls** —
  reported repeatedly ([openai-python issue #868](https://github.com/openai/openai-python/issues/868);
  [community thread: different embeddings for exact same text](https://community.openai.com/t/different-embeddings-for-exact-same-text/411223);
  [thread on -3-small](https://community.openai.com/t/are-vectors-generated-by-text-embedding-3-small-always-the-same-for-the-same-text-input/739757)).
  Contributing factors named: model version changes, batching/ordering effects, GPU
  non-determinism ([Opened AI explainer](https://openedai.io/are-openai-embeddings-deterministic/)).
  Reproducibility limits of RAG systems as a whole are studied in
  [arXiv:2509.18869](https://arxiv.org/pdf/2509.18869).
- **Local inference is not automatically bit-exact either.** ONNX "operators promise the
  mathematically correct result assuming real-number arithmetic, but do not guarantee
  bit-exact results": summation order, fused ops, extended-precision registers, and SIMD /
  multi-core / GPU parallelism all vary results; ONNX Runtime (unlike PyTorch flags) offers
  no determinism switch
  ([emmtrix wiki: Numerical Precision in ONNX](https://www.emmtrix.com/wiki/Numerical_Precision_in_ONNX_and_AI_Inference);
  [floating-point non-associativity study, arXiv:2408.05148](https://arxiv.org/pdf/2408.05148)).
- The nondeterminism in LLM inference at temperature 0 was traced to **batch-size-dependent
  reduction kernels**; batch-invariant kernels achieve bit-identical outputs across 1,000
  runs — i.e., determinism is achievable but requires controlling kernels, batch size, and
  hardware ([Thinking Machines: Defeating Nondeterminism](https://thinkingmachines.ai/blog/defeating-nondeterminism-in-llm-inference/);
  [HEAL, arXiv:2606.21023](https://arxiv.org/pdf/2606.21023)).
- Practical implication documented across these sources: **same binary + same hardware +
  same batch shape + CPU single-thread** is the reproducible envelope; cross-machine
  bit-exactness is not promised by any mainstream runtime.
- **Nomic Embed** is explicitly positioned as a **reproducible, auditable** open-weights
  embedding model (training code + weights released for end-to-end reproducibility)
  ([Nomic Embed paper, arXiv:2402.01613](https://arxiv.org/pdf/2402.01613)).

## §8 Local embedding inference: models, sizes, latency, Go options

- Small strong open models: **all-MiniLM-L6-v2** (384-dim, the default compact
  sentence-transformers model), **bge-small-en-v1.5** (384-dim, 512-token context, GGUF
  builds exist), **nomic-embed-text** (768-dim, long context, GGUF; v2 is MoE multilingual)
  ([sbert efficiency docs](https://sbert.net/docs/sentence_transformer/usage/efficiency.html);
  [bge-small GGUF](https://huggingface.co/ChristianAzinn/bge-small-en-v1.5-gguf);
  [nomic GGUF](https://huggingface.co/nomic-ai/nomic-embed-text-v2-moe-GGUF);
  [open-source embedding model roundups](https://www.baseten.co/blog/the-best-open-source-embedding-models/)).
- CPU inference of these small models is fast: int8-quantized MiniLM-class models reach
  **~266 queries/sec on CPU** (≈4 ms/query) via ONNX/OpenVINO dynamic quantization
  ([philschmid: Optimum + ONNX Runtime](https://www.philschmid.de/optimize-sentence-transformers);
  [Spring AI ONNX embeddings](https://docs.spring.io/spring-ai/reference/api/embeddings/onnx.html)).
- **Go paths:** [hugot](https://github.com/knights-analytics/hugot) runs Hugging Face
  transformer pipelines (incl. feature-extraction/embeddings) in Go over ONNX Runtime;
  [all-minilm-l6-v2-go](https://pkg.go.dev/github.com/clems4ever/all-minilm-l6-v2-go)
  packages MiniLM directly for Go
  ([worked example](https://medium.com/@clement.michaud/text-embeddings-in-go-huggingface-power-without-python-143623938a0f));
  [onnx-gomlx](https://github.com/gomlx/onnx-gomlx) runs ONNX graphs in GoMLX; and
  **llama.cpp** serves any GGUF embedding model over an OpenAI-compatible HTTP endpoint
  usable from Go ([llama.cpp](https://github.com/ggml-org/llama.cpp)).

## Sources

1. Park et al., Generative Agents — https://ar5iv.labs.arxiv.org/html/2304.03442 · https://dl.acm.org/doi/fullHtml/10.1145/3586183.3606763
2. AgentPatterns, Generative Agents Memory Stream — https://agentpatterns.ai/agent-design/generative-agents-memory-stream/
3. Ruder newsletter on generative agents — https://newsletter.ruder.io/p/generative-agents-forums-for-foundation
4. Lukyanenko paper review — https://artgor.medium.com/paper-review-generative-agents-interactive-simulacra-of-human-behavior-cc5f8294b4ac
5. Hindsight/Vectorize, The Consolidation Problem in Agent Memory — https://hindsight.vectorize.io/blog/2026/05/21/agent-memory-consolidation
6. EmergentMind, Memory Mechanisms in LLM Agents — https://www.emergentmind.com/topics/memory-mechanisms-in-llm-based-agents
7. Learning What to Remember (arXiv:2606.12945) — https://arxiv.org/pdf/2606.12945
8. Memory for Autonomous LLM Agents survey (arXiv:2603.07670) — https://arxiv.org/html/2603.07670v1
9. Rethinking Memory in LLM Agents (arXiv:2505.00675) — https://arxiv.org/pdf/2505.00675
10. SAGE novelty gate (arXiv:2605.30711) — https://arxiv.org/abs/2605.30711 · https://github.com/swang1024/SAGE
11. Worth Remembering: surprise-gated robot episodic memory (arXiv:2606.03787) — https://arxiv.org/pdf/2606.03787
12. Marsland et al., Novelty Detection on a Mobile Robot Using Habituation (arXiv:cs/0006007) — https://arxiv.org/pdf/cs/0006007
13. Novelty detection (overview) — https://en.wikipedia.org/wiki/Novelty_detection
14. Joint Agent Memory and Exploration Learning (arXiv:2606.01528) — https://arxiv.org/html/2606.01528
15. RecMem (arXiv:2605.16045) — https://arxiv.org/html/2605.16045v1
16. CLAG (arXiv:2603.15421) — https://arxiv.org/pdf/2603.15421
17. apxml, Memory Consolidation & Summarization — https://apxml.com/courses/agentic-llm-memory-architectures/chapter-3-designing-memory-systems/memory-consolidation-summarization
18. Awesome-Agent-Memory — https://github.com/TeleAI-UAGI/Awesome-Agent-Memory
19. CoALA (arXiv:2309.02427) — https://arxiv.org/pdf/2309.02427
20. MemMachine (arXiv:2604.04853) — https://arxiv.org/html/2604.04853v1
21. Fixed-Persona SLMs with Modular Memory (arXiv:2511.10277) — https://arxiv.org/html/2511.10277v1
22. Wanderfolk, games where NPCs remember you — https://wanderfolk.ai/games-where-npcs-remember-you/
23. Supermemory, Memory Retrieval Latency Budgets — https://supermemory.ai/blog/latency-budgets-memory-retrieval
24. Weaviate, Vector Search Explained — https://weaviate.io/blog/vector-search-explained
25. Vectroid, KNNs vs ANNs — https://www.vectroid.com/resources/knn-vs-ann-comprehensive-overview
26. Ozkaya, Vector Search Algorithms — https://mehmetozkaya.medium.com/vector-search-algorithms-knn-ann-and-disk-ann-803c620bf73e
27. SQLite.ai, Intro to Vector Search — https://blog.sqlite.ai/intro-to-vector-search-nearest-neighbor-search-algorithms-for-rag-applications
28. openai-python issue #868 (ada-002 not deterministic) — https://github.com/openai/openai-python/issues/868
29. OpenAI community threads on embedding nondeterminism — https://community.openai.com/t/different-embeddings-for-exact-same-text/411223 · https://community.openai.com/t/are-vectors-generated-by-text-embedding-3-small-always-the-same-for-the-same-text-input/739757
30. Opened AI, Are OpenAI Embeddings Deterministic — https://openedai.io/are-openai-embeddings-deterministic/
31. On the Reproducibility Limitations of RAG Systems (arXiv:2509.18869) — https://arxiv.org/pdf/2509.18869
32. emmtrix, Numerical Precision in ONNX and AI Inference — https://www.emmtrix.com/wiki/Numerical_Precision_in_ONNX_and_AI_Inference
33. Floating-point non-associativity & reproducibility (arXiv:2408.05148) — https://arxiv.org/pdf/2408.05148
34. Thinking Machines, Defeating Nondeterminism in LLM Inference — https://thinkingmachines.ai/blog/defeating-nondeterminism-in-llm-inference/
35. HEAL (arXiv:2606.21023) — https://arxiv.org/pdf/2606.21023
36. Nomic Embed: Training a Reproducible Long Context Text Embedder (arXiv:2402.01613) — https://arxiv.org/pdf/2402.01613
37. sbert.net, Speeding up Inference — https://sbert.net/docs/sentence_transformer/usage/efficiency.html
38. philschmid, Accelerate Sentence Transformers with Optimum — https://www.philschmid.de/optimize-sentence-transformers
39. Spring AI, Transformers (ONNX) Embeddings — https://docs.spring.io/spring-ai/reference/api/embeddings/onnx.html
40. hugot (Go ONNX transformer pipelines) — https://github.com/knights-analytics/hugot
41. all-minilm-l6-v2-go — https://pkg.go.dev/github.com/clems4ever/all-minilm-l6-v2-go
42. Michaud, Text Embeddings in Go — https://medium.com/@clement.michaud/text-embeddings-in-go-huggingface-power-without-python-143623938a0f
43. onnx-gomlx — https://github.com/gomlx/onnx-gomlx
44. llama.cpp — https://github.com/ggml-org/llama.cpp
45. bge-small-en-v1.5 GGUF — https://huggingface.co/ChristianAzinn/bge-small-en-v1.5-gguf · nomic-embed GGUF — https://huggingface.co/nomic-ai/nomic-embed-text-v2-moe-GGUF
46. Baseten, best open-source embedding models — https://www.baseten.co/blog/the-best-open-source-embedding-models/
