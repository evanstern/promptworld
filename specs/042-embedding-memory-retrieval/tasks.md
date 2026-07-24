# Tasks: Embedding Memory Retrieval

**Input**: Design documents from `/specs/042-embedding-memory-retrieval/`

**Prerequisites**: plan.md, spec.md, research.md (D1–D8), data-model.md,
contracts/embedding-events.md, contracts/relevance-scoring.md, quickstart.md

**Tests**: included — the spec explicitly demands proof tests (FR-005 isolation,
shadow invariant, SC-004 scenario, SC-001 replay) and the project's replay harness is a
standing gate.

**Organization**: grouped by user story; each phase is an independently testable
increment. US1 alone is the MVP (semantically indexed, replay-safe memory stream).

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

**Purpose**: config surface + wire-level plumbing both drivers and scoring depend on

- [x] T001 Add `embedding` route kind to the kinds enum and route validation (absent
      route = subsystem OFF, boot info line, no backfill; anthropic transport routed to
      embedding = boot config error) in `internal/llm/llm.go` and
      `internal/llm/config.go` (contracts/embedding-events.md §4)
- [x] T002 Implement `Embed(ctx, texts []string) ([][]float32, model string, err)` on
      the openai_compat transport, POST `endpoint+"/embeddings"`, in
      `internal/llm/providers.go`; typed unsupported error on anthropic
- [x] T003 [P] Add `memory_relevance` three-state flag (`""`/`"shadow"`/`"on"`,
      omitempty, byte-identical round-trip for old worlds — Teaching precedent) in
      `internal/world/world.go` + validation test in `internal/world/world_test.go`

---

## Phase 2: Foundational (blocking prerequisites for all stories)

**Purpose**: memory identity + payload shapes every story reads/writes

- [x] T004 Add `Seq int64`, `Vec []float32`, `VecModel string` (all omitempty) to
      `Memory`, and `SitVec []float32` / `SitVecModel string` / `SitVecTick int64`
      (omitempty) to `Agent`, in `internal/sim/agents.go`; assert pre-feature snapshot
      bytes unchanged in `internal/sim/state_test.go` (data-model.md invariants)
- [x] T005 Reducer stamps `Memory.Seq` from the emitting event's store seq on
      `agent.memory_added` apply in `internal/sim/state.go`; replay-stability test
      (same stream twice → same seqs) in `internal/sim/state_test.go`
- [x] T006 Define `MemoryEmbeddedPayload` and `SituationEmbeddedPayload` + reducer arms
      (attach by `{agent, mem_seq}`, no-op if absent; set agent SitVec*) in
      `internal/sim/state.go` (+ payload structs beside `MemoryAddedPayload` in
      `internal/sim/agents.go`), per contracts/embedding-events.md §1
- [x] T007 Whitelist `agent.memory_embedded`, `agent.situation_embedded` (reducer-armed)
      and `cog.memory_divergence` (no-op) in the `InjectSocial` whitelist in
      `internal/sim/loop.go`, with door-ordering comment (companion never precedes its
      memory)

**Checkpoint**: `go test ./internal/sim ./internal/world ./internal/llm` green; replay
harness green; no behavior change anywhere (all additions dormant).

---

## Phase 3: User Story 1 — vectors from birth, replay never recomputes (P1) 🎯 MVP

**Goal**: every new memory gains a recorded vector via the async embedder; replay is
byte-identical with the embedder absent.

**Independent test**: quickstart.md US1 — embedded count converges to added count;
replay passes with Ollama stopped; kill-Ollama run stays loud + non-fatal.

- [x] T008 [US1] Create the embedder driver: watch absorbed `agent.memory_added`, embed
      the exact recorded text (fixed-byte truncation cap constant), inject
      `agent.memory_embedded` via `InjectSocial`; debounced `daemon.llm_warning` on
      transport failure; never blocks absorb/plan; wired only when the `embedding`
      route exists; warm-pins the embedding model at driver start + slow re-warm
      (native `/api/embed` `keep_alive:-1`, best-effort, per contract §2) — new file
      `internal/mind/embedder.go` (contracts/embedding-events.md §2)
- [x] T009 [US1] Boot wiring: construct/start the embedder alongside the consolidation
      driver when llm.json has the `embedding` route, in `internal/daemon/daemon.go`
- [x] T010 [P] [US1] Driver unit tests: emission-ordered per-agent companions, failure
      debounce, vectorless-forever on skip (no backfill), in
      `internal/mind/embedder_test.go`
- [x] T011 [US1] End-to-end + replay proof: run a seeded world, assert SC-002
      (100% coverage absent outage) and SC-001 (replay byte-identical with zero embed
      calls — embedder not wired during replay), extending the existing replay harness
      test in `internal/daemon/` (or `internal/sim/replay_test.go` if that's where the
      harness lives — follow the existing test's home)

**Checkpoint**: US1 demonstrable end-to-end per quickstart; MVP complete.

---

## Phase 4: User Story 2 — shadow-mode divergence instrumentation (P2)

**Goal**: both rankings computed at plan time; prompts unchanged; divergence recorded.

**Independent test**: quickstart.md US2 — shadow invariant (prompt bytes identical to
off-mode) + divergence events per plan job + a ≥1-game-day summary.

- [x] T012 [US2] Situation-vector leg of the embedder: deterministic situation template
      (time phase · position + place desc · worst needs · active intent verb+reason ·
      nearby agent names) rendered from replica state at each `PlannerCadenceTicks`
      bucket edge, embedded, injected as `agent.situation_embedded` — in
      `internal/mind/embedder.go` (research D5)
- [x] T013 [US2] Pure three-term scorer `SelectMemoriesRelevant` (equal-weight
      normalized sal01 + rel01; neutral 0.5 for vectorless/cross-model; nil query →
      legacy delegation; serendipity tail byte-identical to today; ties newer-wins;
      sequential float64 cosine) in `internal/sim/memory.go`
      (contracts/relevance-scoring.md §1–2)
- [x] T014 [P] [US2] Scorer unit tests: fallback ladder rows, neutrality, tail
      byte-identity vs `SelectMemories`, determinism across runs, in
      `internal/sim/memory_test.go`
- [x] T015 [US2] `cog.memory_divergence` payload + emitter (agent, tick, mode, both
      windows as seqs, overlap, displacement, vectorless count, sit_tick) in
      `internal/sim/cognition.go` + `internal/mind/telemetry.go`
      (data-model.md event table)
- [x] T016 [US2] Mode gating at both prompt paths: read `memory_relevance`; `"shadow"` →
      compute both rankings, prompt gets legacy, emit divergence; in
      `internal/mind/prompt.go` (planner window) and `internal/mind/convo.go` (k=5
      scene snapshot)
- [x] T017 [US2] Shadow-invariant test: same seed, `""` vs `"shadow"` → byte-identical
      prompts/behavior, divergence events present, replay green — in
      `internal/mind/` beside the existing mind tests
- [x] T018 [P] [US2] Divergence summary tooling: aggregate `cog.memory_divergence`
      into mean overlap@K + promoted-memory share per agent/day (a small CLI or test
      helper reading the store — match the project's existing tooling pattern, e.g.
      under `cmd/` or as a `go test -run Summary` helper); document invocation in
      `specs/042-embedding-memory-retrieval/quickstart.md`

**Checkpoint**: shadow mode running; SC-003 summary producible; US2 gate evidence
collectable. **STOP — the shadow→on flip is an operator decision recorded on TASK-98
(FR-007), not a task.**

---

## Phase 5: User Story 3 — relevance shapes the window (P3)

**Goal**: `"on"` mode feeds the augmented window to prompts, guarded by the pinned
invariants.

**Independent test**: quickstart.md US3 — SC-004 scenario + SC-006 isolation + replay.

- [x] T019 [US3] `"on"` mode: prompt paths consume `SelectMemoriesRelevant` (divergence
      still recorded) in `internal/mind/prompt.go` + `internal/mind/convo.go`
- [x] T020 [P] [US3] SC-004 scenario test: old low-salience situation-matching memory
      enters the `on` window, provably absent from legacy window, in
      `internal/sim/memory_test.go`
- [x] T021 [P] [US3] SC-006 isolation test: mutate agent A's memories → agent B's
      selection byte-identical in every mode, in `internal/sim/memory_test.go`
      (FR-005)
- [x] T022 [US3] Replay byte-identity of an `"on"` world (selection pure over recorded
      vectors) added to the harness test from T011

**Checkpoint**: all three stories independently proven; feature complete pending gate
decision + polish.

---

## Phase 6: Polish & cross-cutting

- [x] T023 [P] Operator docs: `embedding` kind + provider example + `memory_relevance`
      flag + shadow→on gate procedure + warm-pin/keep-alive note (perma-loaded embed
      model; chat models untouched) in `docs/llm-providers.md` and the event
      additions in `contracts/events.md` (the repo's event catalog — keep
      TestCatalogSweep green)
- [x] T024 Throughput check (SC-005): wall-clock per game-day, embedding on vs off,
      same seed — record numbers in TASK-98 notes (expect ≈0 delta; budget 10%)
- [x] T025 Quickstart walkthrough executed top-to-bottom against a real world; fix any
      drift in `specs/042-embedding-memory-retrieval/quickstart.md`

**Post-merge (TASK-98 ACs, not code tasks)**: divergence threshold decision recorded on
the board (AC#4); `/grounding-wiki:wiki-update` re-pin + player-docs freshness;
**codebase-to-course run on this feature before the PR ships (AC#7)**.

---

## Dependencies

```
Setup (T001–T003)
  └─▶ Foundational (T004–T007)   [T004 → T005 → T006 → T007]
        └─▶ US1 (T008–T011)      [T008 → T009 → {T010, T011}]
              └─▶ US2 (T012–T018) [T012,T013 parallel → T015 → T016 → T017; T014,T018 parallel]
                    └─▶ US3 (T019–T022) [T019 → {T020,T021,T022}]  ← gated on FR-007 decision
                          └─▶ Polish (T023–T025)
```

US2 depends on US1 only for live vectors (the scorer T013/T014 is independently
developable against fixture vectors). US3 depends on US2's instrumentation by design
(the guardrail), not by code.

## Parallel opportunities

- Phase 1: T003 ∥ (T001→T002)
- US1: T010 ∥ T011 once T008–T009 land
- US2: T012 ∥ T013–T014 (different packages); T018 ∥ everything after T015
- US3: T020 ∥ T021 ∥ T022
- Polish: T023 ∥ T024

## Implementation strategy

MVP = Phase 1–3 (US1): semantically indexed, replay-safe memory stream — shippable and
valuable alone. Then US2 to collect gate evidence during normal runs. US3 flips only
after the recorded threshold decision. Model-tier note (constitution V, recorded in
plan.md): core slices (T004–T009, T012–T013, T015–T016, T019) are cross-package/
doctrine-adjacent → **Opus 4.8 spec-implementer**; leaf tests/docs/tooling (T010, T014,
T017–T018, T020–T025) may run on Sonnet.
