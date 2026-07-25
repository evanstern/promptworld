---
id: TASK-98
title: >-
  Embedding memory retrieval: record-at-emission vectors + relevance term in
  memory selection
status: Done
assignee: []
created_date: '2026-07-24 19:45'
updated_date: '2026-07-25 03:25'
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
ordinal: 500
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Adopt embedding-based memory retrieval per the vault analysis (research/Agent-Memory-Retrieval/Analysis-Embedding-Retrieval-Adoption.md): add a query-conditioned relevance term (cosine similarity) to the working-memory window selection, with each memory's vector recorded at emission so replay never re-embeds. Covers build-order steps 1+2 of the analysis (embed-at-emission infrastructure + relevance in selection; open-ended/varying memory text rides along as the enabler). Step 3 (consolidation clustering + habituation) is a separate follow-on task. The salience table stays: salience remains the deterministic write-time control-plane signal (GenerationBumpSalience, rumorMinSalience, governance evidence); embeddings only affect which memories enter prompts. Write-time kNN salience scoring is explicitly deferred per the analysis.

Spec: specs/042-embedding-memory-retrieval
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Each new episodic memory's embedding vector is recorded at emission (event payload or companion event); replay is byte-identical and never re-embeds — determinism doctrine holds (record-at-emission, pure functions over recorded floats)
- [x] #2 Memory selection scores with a relevance term (cosine of memory vector vs recorded query/situation vector) alongside existing salience and recency; selection stays a pure function of recorded data
- [x] #3 Guardrail (analysis finding 1): recency continues to decay from creation time — recency-from-last-access is rejected; no state mutation on read. If access-refresh is ever wanted it must be a recorded event (documented, not implemented)
- [x] #4 Guardrail (analysis finding 2): rank-divergence instrumentation ships BEFORE relevance drives prompts — log divergence between salience×recency and the three-term score; a documented threshold decides whether relevance stays on
- [x] #5 Per-agent isolation: embedding storage and retrieval are strictly per-agent — no cross-agent vector reads possible by construction, covered by a test (no shared memory pool)
- [x] #6 Embedding model choice is pinned and documented (local small model, e.g. 384-dim MiniLM/bge class via hugot or llama.cpp sidecar); model file/version is part of replay hygiene
- [x] #7 codebase-to-course run tagged on this feature before the PR ships (docs/course refreshed to teach the new memory-retrieval mechanic)
- [x] #8 Spec rigor: full Spec Kit (specify → clarify → plan → tasks) with spec-bridge:link BEFORE implementation; wiki re-ground + player-docs freshness after merge
- [x] #9 Spec phase: Setup
- [x] #10 Spec phase: Foundational (blocking prerequisites for all stories)
- [x] #11 Spec phase: User Story 1 — vectors from birth, replay never recomputes (P1) 🎯 MVP
- [x] #12 Spec phase: User Story 2 — shadow-mode divergence instrumentation (P2)
- [x] #13 Spec phase: User Story 3 — relevance shapes the window (P3)
- [x] #14 Spec phase: Polish & cross-cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
spec-bridge sync: Setup: 0/3 · Foundational (blocking prerequisites for all stories): 0/4 · User Story 1 — vectors from birth, replay never recomputes (P1) 🎯 MVP: 0/4 · User Story 2 — shadow-mode divergence instrumentation (P2): 0/7 · User Story 3 — relevance shapes the window (P3): 0/4 · Polish & cross-cutting: 0/3

Model-tier decision (constitution V): core slices delegated to spec-implementer @ Opus 4.8 — rubric: cross-package (llm/mind/sim/world/daemon), concurrency (async embedder driver vs absorb/plan goroutines), doctrine-adjacent (replay byte-identity, InjectSocial door semantics, reducer arms). Leaf tests/docs may run Sonnet. Implementation in .worktrees/task-98 (branch task-98-embedding-memory-retrieval), phased: (1) T001-T011 Setup+Foundational+US1 MVP, (2) T012-T018 US2 shadow, (3) T019-T025 US3+Polish. Fable gates between phases.

Ops decision (2026-07-24): embedding model kept perma-loaded on the serving host (localhost Ollama — all world endpoints are local). Mechanism: embedder driver warm-pin via native /api/embed keep_alive:-1 at start + hourly re-warm (empirically: compat traffic preserves the pin; compat-body keep_alive ignored). Scope: embed model only (~46MB); chat models keep default eviction (24GB RAM). Contract §2 + T008/T023 updated, committed to main.

spec-bridge sync: Setup: 3/3 · Foundational (blocking prerequisites for all stories): 4/4 · User Story 1 — vectors from birth, replay never recomputes (P1) 🎯 MVP: 4/4 · User Story 2 — shadow-mode divergence instrumentation (P2): 0/7 · User Story 3 — relevance shapes the window (P3): 0/4 · Polish & cross-cutting: 0/3

spec-bridge sync: Setup: 3/3 · Foundational (blocking prerequisites for all stories): 4/4 · User Story 1 — vectors from birth, replay never recomputes (P1) 🎯 MVP: 4/4 · User Story 2 — shadow-mode divergence instrumentation (P2): 7/7 · User Story 3 — relevance shapes the window (P3): 0/4 · Polish & cross-cutting: 0/3

spec-bridge sync: Setup: 3/3 · Foundational: 4/4 · US1 (P1 MVP): 4/4 · US2 (P2): 7/7 · US3 (P3): 4/4 · Polish: 3/3 — spec artifacts Done-eligible. Status Done DEFERRED (stays In Progress): deliverable is one merged PR (not yet opened); AC#4 divergence threshold decision, AC#7 course run before PR, AC#8 post-merge re-ground still open. Done at merge per repo precedent (TASK-91/41 pattern). SC-005 measured: one game day at max speed, embedding off 0.90-0.95s vs on 0.95-1.00s = +5.0-5.7% (budget 10%); ~300-330 embed calls; delta shrinks at paced speeds.

AC#7 done pre-PR: codebase-to-course run shipped on the branch (docs/course/, commit be5dabf) — "How a Villager Remembers", 5 modules, course gate green. Branch ready for PR: 15 commits, all three stories + polish gated, full suite green except pre-existing TestCatalogSweep.

PR opened: https://github.com/evanstern/promptworld/pull/70 (branch task-98-embedding-memory-retrieval, 20 commits incl. docs/course). Remaining to Done: merge, then AC#8 (wiki re-ground + player docs); AC#4 (shadow→on threshold decision) is post-merge operational.

Rebase onto main (041 mental maps, PR #69) complete: both features intact, unions at all shared seams (Agent fields, whitelist, prompt sections, boot wiring); one semantic reconciliation — 041 made InjectSocial dry-runs expensive, embedder now batches/coalesces (contract-sanctioned) restoring SC-005 to +0.4-3.9% (was +60-74% naive post-rebase; note 041 itself moved the off-baseline 0.92s→1.30s via its per-beat sweep). Independent gate green incl. 041 headline tests. PR #70 MERGEABLE/CLEAN @ ec9a7a0.

Post-merge close-out: ACs 1-3,5,6 proven by merged code+tests (record-at-emission/replay meter, three-term selection, creation-time recency, isolation test, pinned all-minilm:latest). AC#4 guardrail satisfied as specified: instrumentation shipped and relevance drives no prompts anywhere (flag off in all worlds); the shadow→on flip per world follows the documented gate procedure (docs/llm-providers.md + promptworld divergence) and gets recorded here when made. AC#8 done: wiki re-ground (26 notes + new memory-retrieval note) + player docs 7/7 fresh, commit 6a1cd67. AC#7 course shipped in PR. Worktree/branch cleaned; merge verified via gh api.

spec-bridge sync: Setup: 3/3 · Foundational (blocking prerequisites for all stories): 4/4 · User Story 1 — vectors from birth, replay never recomputes (P1) 🎯 MVP: 4/4 · User Story 2 — shadow-mode divergence instrumentation (P2): 7/7 · User Story 3 — relevance shapes the window (P3): 4/4 · Polish & cross-cutting: 3/3 — status In Progress → Done
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Setup: 3/3 · Foundational (blocking prerequisites for all stories): 4/4 · User Story 1 — vectors from birth, replay never recomputes (P1) 🎯 MVP: 4/4 · User Story 2 — shadow-mode divergence instrumentation (P2): 7/7 · User Story 3 — relevance shapes the window (P3): 4/4 · Polish & cross-cutting: 3/3). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
