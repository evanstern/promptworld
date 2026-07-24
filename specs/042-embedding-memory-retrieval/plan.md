# Implementation Plan: Embedding Memory Retrieval

**Branch**: `042-embedding-memory-retrieval` | **Date**: 2026-07-24 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/042-embedding-memory-retrieval/spec.md`

## Summary

Give every episodic memory a recorded embedding vector and give memory-window selection a
query-conditioned relevance term — without touching the deterministic core. Embedding runs
as an async mind-side driver (peer of consolidation) that watches `agent.memory_added`,
calls a pinned local model through a new `embedding` kind in `internal/llm`, and injects
recorded companion events (`agent.memory_embedded`, `agent.situation_embedded`) through
the existing `InjectSocial` door. The reducer attaches vectors verbatim; replay never
re-embeds. Selection gains a pure sibling `SelectMemoriesRelevant` in `internal/sim`,
gated by a three-state `memory_relevance` world flag (`""`/`"shadow"`/`"on"`): shadow mode
records rank-divergence (`cog.memory_divergence` reducer no-ops) while prompts still get
the legacy window; `on` is flipped only after the documented divergence threshold
decision (spec US2 → US3 gate). Full decisions with rationale: [research.md](./research.md).

## Technical Context

**Language/Version**: Go (module at repo root; pure-Go dependency posture — modernc SQLite)

**Primary Dependencies**: `internal/llm` (openai_compat transport → Ollama, new
`/embeddings` path + `embedding` route kind, default pin `all-minilm` 384-dim);
`internal/mind` (new embedder driver + shadow scoring at plan time); `internal/sim`
(Memory fields, `SelectMemoriesRelevant`, reducer arms, event whitelist);
`internal/store` (no schema change — payloads are JSON TEXT); `internal/world`
(`memory_relevance` flag)

**Storage**: existing SQLite event store; vectors as JSON `[]float32` in event payloads
(~4–6 KB/memory; tens of MB at village scale — immaterial). No FormatVersion bump:
all payload/snapshot additions are omitempty additive (spec-019/030 precedent)

**Testing**: `go test ./...`; replay byte-identity harness (existing gate); new
isolation test (FR-005), vectorless/cross-model fallback tests, shadow-mode
no-behavior-change test, US3 scenario test (relevance-promoted memory)

**Target Platform**: darwin/linux dev machines running the daemon + Ollama locally

**Project Type**: single Go project, event-sourced sim daemon

**Performance Goals**: tick throughput within 10% of baseline (SC-005 — satisfied by
construction: zero model calls on the tick path); embeds are cadence-bounded
(≤ agents × 1 situation embed per planner-cadence bucket + 1 per new memory)

**Constraints**: byte-identical replay (zero embedding computation during replay);
model-free sim layer (all model I/O behind the mind-side doors); per-agent isolation by
construction; salience semantics untouched (FR-008)

**Scale/Scope**: village scale (tens of agents, thousands of memories/agent over long
runs); ~6 packages touched; no new external services beyond the already-local Ollama

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Artifact-Grounded Action** — PASS. Vectors, situation texts, and divergence
  metrics are all recorded events; the shadow→on decision is gated on a recorded summary
  and a board-task artifact (FR-007). No decision lives only in chat.
- **II. One Task, One PR** — PASS. TASK-98 ↔ this spec ↔ one branch
  (`.worktrees/task-98`) ↔ one PR. TASK-99 (consolidation clustering) is explicitly out
  of scope.
- **III. Gates Over Assertions** — PASS. Replay byte-identity harness, spec-bridge gate,
  and the new divergence threshold gate (US2→US3) are all artifact-checked; shadow mode
  is designed so behavior cannot change before the gate passes.
- **IV. Grounding Freshness** — PASS (planned). `docs/wiki/` notes pinning
  `internal/sim/memory.go`, `internal/mind/prompt.go`, `internal/llm/*` sources must be
  re-pinned via `/grounding-wiki:wiki-update` before Done; player-docs freshness check
  after; **plus the TASK-98 AC: a codebase-to-course run on this feature before the PR
  ships**.
- **V. Model-Tiered Workflow** — PASS (recorded here for the implement phase): this
  feature is cross-package (llm/mind/sim/world), touches cognition scheduling and
  doctrine-adjacent replay semantics → **Opus 4.8 via spec-implementer** for the core
  slices (embedder driver, reducer arms, selector, shadow scoring); Sonnet acceptable for
  leaf work (docs reconciliation, quickstart validation). Tier choice + rubric
  justification to be recorded on TASK-98 at delegation time.

**Post-Phase-1 re-check (2026-07-24)**: no new violations introduced by the design.
Complexity Tracking stays empty — the design reuses every existing seam (doors, kinds,
omitempty payloads, cog.* no-ops) rather than adding new architecture.

## Project Structure

### Documentation (this feature)

```text
specs/042-embedding-memory-retrieval/
├── plan.md              # This file
├── research.md          # Phase 0 — decisions D1–D8 with rationale
├── data-model.md        # Phase 1 — entities, payloads, flags
├── quickstart.md        # Phase 1 — end-to-end validation guide
├── contracts/
│   ├── embedding-events.md    # event schemas + reducer semantics + determinism rules
│   └── relevance-scoring.md   # three-term score, fallbacks, shadow/on gating
├── checklists/requirements.md # spec quality checklist (passing)
└── tasks.md             # Phase 2 — /speckit-tasks output (not yet)
```

### Source Code (repository root)

```text
internal/
├── llm/
│   ├── llm.go           # + Kind "embedding" in the kinds enum/routes
│   ├── config.go        # + provider/route validation for the embedding kind
│   └── providers.go     # + Embed() on openai_compat (POST endpoint+"/embeddings")
├── mind/
│   ├── embedder.go      # NEW: async driver — watches memory_added, embeds memory
│   │                    #      text + cadence situation text, injects companions
│   ├── prompt.go        # window selection: mode-aware (legacy/shadow/on)
│   ├── convo.go         # same mode-awareness for the k=5 scene snapshot
│   └── telemetry.go     # + cog.memory_divergence emitter
├── sim/
│   ├── memory.go        # + Memory.Seq/Vec/VecModel; + SelectMemoriesRelevant (pure)
│   ├── state.go         # reducer: stamp Seq on memory_added; arm memory_embedded /
│   │                    #          situation_embedded attach
│   ├── loop.go          # InjectSocial whitelist: + the two agent.* companions,
│   │                    #          + cog.memory_divergence no-op
│   └── agents.go        # Memory struct fields (omitempty additive)
├── world/
│   └── world.go         # + MemoryRelevance string flag (omitempty additive)
└── daemon/
    └── daemon.go        # boot: wire embedder driver when the embedding route exists

docs/llm-providers.md    # + embedding kind operator docs
internal/sim/*_test.go, internal/mind/*_test.go   # tests per Testing above
```

**Structure Decision**: single Go project, existing package boundaries. The only new file
is `internal/mind/embedder.go` — every other change extends an existing seam. The
sim/mind boundary rule of the design: *model I/O and mode policy live in mind; scoring
math, memory state, and event semantics live in sim as pure functions of recorded data.*

## Complexity Tracking

No constitution violations to justify — table intentionally empty.
