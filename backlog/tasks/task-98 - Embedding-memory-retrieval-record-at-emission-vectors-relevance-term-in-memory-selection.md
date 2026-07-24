---
id: TASK-98
title: >-
  Embedding memory retrieval: record-at-emission vectors + relevance term in
  memory selection
status: In Progress
assignee: []
created_date: '2026-07-24 19:45'
updated_date: '2026-07-24 20:33'
labels:
  - memory
  - embeddings
  - cognition
dependencies: []
references:
  - research/Agent-Memory-Retrieval/Analysis-Embedding-Retrieval-Adoption.md
  - research/Agent-Memory-Retrieval/_grounding.md
  - 'https://claude.ai/code/artifact/cc1f9eb9-4357-497a-904d-acaefc1ed1e9'
priority: medium
ordinal: 82000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Adopt embedding-based memory retrieval per the vault analysis (research/Agent-Memory-Retrieval/Analysis-Embedding-Retrieval-Adoption.md): add a query-conditioned relevance term (cosine similarity) to the working-memory window selection, with each memory's vector recorded at emission so replay never re-embeds. Covers build-order steps 1+2 of the analysis (embed-at-emission infrastructure + relevance in selection; open-ended/varying memory text rides along as the enabler). Step 3 (consolidation clustering + habituation) is a separate follow-on task. The salience table stays: salience remains the deterministic write-time control-plane signal (GenerationBumpSalience, rumorMinSalience, governance evidence); embeddings only affect which memories enter prompts. Write-time kNN salience scoring is explicitly deferred per the analysis.

Spec: specs/042-embedding-memory-retrieval
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Each new episodic memory's embedding vector is recorded at emission (event payload or companion event); replay is byte-identical and never re-embeds — determinism doctrine holds (record-at-emission, pure functions over recorded floats)
- [ ] #2 Memory selection scores with a relevance term (cosine of memory vector vs recorded query/situation vector) alongside existing salience and recency; selection stays a pure function of recorded data
- [ ] #3 Guardrail (analysis finding 1): recency continues to decay from creation time — recency-from-last-access is rejected; no state mutation on read. If access-refresh is ever wanted it must be a recorded event (documented, not implemented)
- [ ] #4 Guardrail (analysis finding 2): rank-divergence instrumentation ships BEFORE relevance drives prompts — log divergence between salience×recency and the three-term score; a documented threshold decides whether relevance stays on
- [ ] #5 Per-agent isolation: embedding storage and retrieval are strictly per-agent — no cross-agent vector reads possible by construction, covered by a test (no shared memory pool)
- [ ] #6 Embedding model choice is pinned and documented (local small model, e.g. 384-dim MiniLM/bge class via hugot or llama.cpp sidecar); model file/version is part of replay hygiene
- [ ] #7 codebase-to-course run tagged on this feature before the PR ships (docs/course refreshed to teach the new memory-retrieval mechanic)
- [ ] #8 Spec rigor: full Spec Kit (specify → clarify → plan → tasks) with spec-bridge:link BEFORE implementation; wiki re-ground + player-docs freshness after merge
- [ ] #9 Spec phase: Setup
- [ ] #10 Spec phase: Foundational (blocking prerequisites for all stories)
- [ ] #11 Spec phase: User Story 1 — vectors from birth, replay never recomputes (P1) 🎯 MVP
- [ ] #12 Spec phase: User Story 2 — shadow-mode divergence instrumentation (P2)
- [ ] #13 Spec phase: User Story 3 — relevance shapes the window (P3)
- [ ] #14 Spec phase: Polish & cross-cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
spec-bridge sync: Setup: 0/3 · Foundational (blocking prerequisites for all stories): 0/4 · User Story 1 — vectors from birth, replay never recomputes (P1) 🎯 MVP: 0/4 · User Story 2 — shadow-mode divergence instrumentation (P2): 0/7 · User Story 3 — relevance shapes the window (P3): 0/4 · Polish & cross-cutting: 0/3

Model-tier decision (constitution V): core slices delegated to spec-implementer @ Opus 4.8 — rubric: cross-package (llm/mind/sim/world/daemon), concurrency (async embedder driver vs absorb/plan goroutines), doctrine-adjacent (replay byte-identity, InjectSocial door semantics, reducer arms). Leaf tests/docs may run Sonnet. Implementation in .worktrees/task-98 (branch task-98-embedding-memory-retrieval), phased: (1) T001-T011 Setup+Foundational+US1 MVP, (2) T012-T018 US2 shadow, (3) T019-T025 US3+Polish. Fable gates between phases.
<!-- SECTION:NOTES:END -->
