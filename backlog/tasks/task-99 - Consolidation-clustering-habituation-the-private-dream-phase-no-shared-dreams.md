---
id: TASK-99
title: >-
  Consolidation clustering + habituation: the private dream phase (no shared
  dreams)
status: Done
assignee: []
created_date: '2026-07-24 19:45'
updated_date: '2026-07-29 23:48'
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

Spec: specs/098-private-dreams
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Clustering and habituation operate strictly within one agent's memory store; no cross-agent bleed is possible by construction, and a test proves two agents with overlapping experiences never influence each other's consolidation
- [x] #2 Habituation: dense clusters of near-duplicate memories are detected via embedding geometry and down-weighted/merged at consolidation; outcomes land as recorded events (salience-revision / merge events), replay-safe
- [x] #3 Cheap-geometry-first economics: clear-cut cluster/no-cluster cases resolved by geometry alone; LLM judgment reserved for the ambiguous band (SAGE routing pattern)
- [x] #4 Noise exploration (design decision recorded in spec): whether/how to inject seeded noise (rngAt) into the clustering pass for dream-like variance — decided, justified, and documented even if the answer is no
- [x] #5 Spec rigor: full Spec Kit with spec-bridge:link BEFORE implementation; builds on TASK-98's recorded-vector infrastructure
- [x] #6 Spec phase: Clustering + habituation core
- [x] #7 Spec phase: Noise + dials
- [x] #8 Spec phase: Surfaces + tests
- [x] #9 Spec phase: Evidence + grounding
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
board-sweep-2026-07-29 lane 4: spec 098 landed + linked (AC5 spec rigor satisfied); AC4 noise decision RECORDED in spec D4 — adopted minimally, rngAt-seeded zeroable boundary jitter. Tier: Opus — internal/mind consolidation orchestration + replay-doctrine surface. Dispatch after TASK-95 merges (shared memory/event surfaces); may run parallel with TASK-80, merges serial.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #139. Per-agent dream phase in nightly consolidation: single-store density clustering (privacy proven by byte-identical perturbation test), geometry-first with ambiguous-band consult in the existing LLM slot, recorded salience_revised/memory_merged events (no re-derivation, additive types), rngAt-seeded zeroable boundary jitter per the recorded AC4 decision, five manifest dials with genesis pin. Live demo evidence at docs/design/evidence/task-99/results.md (world preserved). Opus tier; spec 098 all tasks done.
<!-- SECTION:FINAL_SUMMARY:END -->
