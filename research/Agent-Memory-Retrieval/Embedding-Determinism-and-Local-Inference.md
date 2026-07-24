---
title: Embedding Determinism and Local Inference
aliases: [Reproducible Embeddings, Go Embedding Options]
tags: [determinism, reproducibility, embeddings, golang, onnx]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Memory-Retrieval]]", "[[Small-Scale-Vector-Search-and-Latency]]"]
---

# Embedding Determinism and Local Inference

What is known about getting the *same vector for the same text, every time* — the property a
replay-deterministic sim would need if embeddings were ever recomputed rather than recorded —
and the practical options for running embedding models locally, including from Go.

## Hosted APIs: not deterministic

OpenAI embedding endpoints (`ada-002`, `text-embedding-3-*`) return **slightly different
vectors for identical input across calls** — reported repeatedly and acknowledged in
community threads ([openai-python #868](https://github.com/openai/openai-python/issues/868),
[[_grounding]] §7). Named causes: model-version changes, batching/ordering effects, GPU
nondeterminism. RAG-system reproducibility limits are studied as a general phenomenon
([arXiv:2509.18869](https://arxiv.org/pdf/2509.18869)).

## Local inference: reproducible only within an envelope

- ONNX operators promise mathematically correct results **but not bit-exact ones**:
  summation order, fused ops, extended-precision registers, and SIMD/multi-core/GPU
  parallelism all vary outputs; ONNX Runtime exposes **no determinism switch** (unlike
  PyTorch's flags) ([emmtrix](https://www.emmtrix.com/wiki/Numerical_Precision_in_ONNX_and_AI_Inference),
  [[_grounding]] §7).
- Root cause is floating-point **non-associativity** under parallel reduction
  ([arXiv:2408.05148](https://arxiv.org/pdf/2408.05148)); the Thinking Machines work traced
  temperature-0 LLM nondeterminism to batch-size-dependent reduction kernels and achieved
  bit-identical outputs with batch-invariant kernels ([[_grounding]] §7).
- The documented reproducible envelope: **same binary + same hardware + same batch shape +
  single-threaded CPU**. Cross-machine bit-exactness is not promised by any mainstream
  runtime ([[_grounding]] §7).
- **Nomic Embed** is explicitly positioned as a reproducible, auditable open-weights
  embedding model (training code + weights released)
  ([arXiv:2402.01613](https://arxiv.org/pdf/2402.01613)).

## Local models and cost

Small strong open models: **all-MiniLM-L6-v2** (384-dim), **bge-small-en-v1.5** (384-dim,
512-token context, GGUF builds), **nomic-embed-text** (768-dim, long context, GGUF).
Int8-quantized MiniLM-class models reach **~266 queries/sec on CPU (≈4 ms/query)** via
ONNX/OpenVINO quantization ([[_grounding]] §8).

## Go paths

- [hugot](https://github.com/knights-analytics/hugot) — Hugging Face transformer pipelines
  (incl. embeddings) in Go over ONNX Runtime.
- [all-minilm-l6-v2-go](https://pkg.go.dev/github.com/clems4ever/all-minilm-l6-v2-go) —
  MiniLM packaged directly for Go.
- [onnx-gomlx](https://github.com/gomlx/onnx-gomlx) — ONNX graphs in GoMLX.
- [llama.cpp](https://github.com/ggml-org/llama.cpp) — serves any GGUF embedding model over
  an OpenAI-compatible HTTP endpoint callable from Go.

([[_grounding]] §8.)

## Grounding

- [[_grounding]] §7 (determinism), §8 (models, latency, Go options)
- [emmtrix, Numerical Precision in ONNX](https://www.emmtrix.com/wiki/Numerical_Precision_in_ONNX_and_AI_Inference)
- [Nomic Embed, arXiv:2402.01613](https://arxiv.org/pdf/2402.01613)
