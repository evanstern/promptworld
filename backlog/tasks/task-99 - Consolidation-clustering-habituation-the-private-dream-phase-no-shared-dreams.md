---
id: TASK-99
title: >-
  Consolidation clustering + habituation: the private dream phase (no shared
  dreams)
status: In Progress
assignee: []
created_date: '2026-07-24 19:45'
updated_date: '2026-07-29 21:47'
labels:
  - memory
  - embeddings
  - consolidation
dependencies:
  - TASK-98
references:
  - research/Agent-Memory-Retrieval/Analysis-Embedding-Retrieval-Adoption.md
  - research/Agent-Memory-Retrieval/Novelty-Gates-and-Habituation.md
  - research/Agent-Memory-Retrieval/Consolidation-and-Clustering.md
priority: medium
ordinal: 8000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Build-order step 3 of the vault analysis (research/Agent-Memory-Retrieval/Analysis-Embedding-Retrieval-Adoption.md): embedding-cluster detection and habituation reweighting in the nightly consolidation slot, following the cheap-geometry-first patterns (SAGE density gate; RecMem cluster-triggered consolidation). Framing from design discussion 2026-07-24: this phase is dream-like (nightly, reweights and rewrites memories) — and dreams must be PRIVATE. Clustering must never pool memories across agents; a shared vector table would produce shared dreams. Also explore deliberately injecting noise into the clustering pass (dream-like distortion) — which, in this codebase, must be seeded noise (rngAt pattern) to preserve byte-identical replay.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Clustering and habituation operate strictly within one agent's memory store; no cross-agent bleed is possible by construction, and a test proves two agents with overlapping experiences never influence each other's consolidation
- [ ] #2 Habituation: dense clusters of near-duplicate memories are detected via embedding geometry and down-weighted/merged at consolidation; outcomes land as recorded events (salience-revision / merge events), replay-safe
- [ ] #3 Cheap-geometry-first economics: clear-cut cluster/no-cluster cases resolved by geometry alone; LLM judgment reserved for the ambiguous band (SAGE routing pattern)
- [ ] #4 Noise exploration (design decision recorded in spec): whether/how to inject seeded noise (rngAt) into the clustering pass for dream-like variance — decided, justified, and documented even if the answer is no
- [ ] #5 Spec rigor: full Spec Kit with spec-bridge:link BEFORE implementation; builds on TASK-98's recorded-vector infrastructure
<!-- AC:END -->
