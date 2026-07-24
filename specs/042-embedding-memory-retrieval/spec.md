# Feature Specification: Embedding Memory Retrieval

**Feature Branch**: `042-embedding-memory-retrieval`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Embedding memory retrieval: record-at-emission vectors + relevance term in memory selection (TASK-98). Each episodic memory gets an embedding vector recorded at emission (replay never re-embeds); memory window selection gains a query-conditioned relevance term (cosine vs recorded situation vector) alongside existing salience and recency; recency stays creation-time-based (reject last-access decay); rank-divergence instrumentation ships before relevance drives prompts; per-agent store isolation by construction; pinned local embedding model (384-dim class) as part of replay hygiene. Salience table unchanged as the write-time control-plane signal. Grounded in research/Agent-Memory-Retrieval/Analysis-Embedding-Retrieval-Adoption.md."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Memories carry meaning vectors from birth, and replay never recomputes them (Priority: P1)

Every episodic memory an agent forms is stamped, at the moment it is recorded, with a
compact numeric representation of its meaning (an embedding vector) and the identity of
the model that produced it. When an operator replays a recorded world, the replay reads
those stamped vectors as data — no embedding computation of any kind happens during
replay, and the replayed world is byte-identical to the original.

**Why this priority**: This is the foundation every other story stands on, and it is the
story where the project's core invariant (deterministic, byte-identical replay) is at
stake. Nothing relevance-related can ship until vectors exist as recorded, replayable
data. It is independently valuable even alone: the memory stream becomes semantically
indexed for any future consumer.

**Independent Test**: Run a seeded world for several game days; verify every new episodic
memory carries a vector and model identity; replay the world from its event log and
verify byte-identity with zero embedding computations performed during replay.

**Acceptance Scenarios**:

1. **Given** a running world, **When** any episodic memory is emitted (action, witness,
   gossip, gist, digest, omen), **Then** the recorded event carries the memory's
   embedding vector and the producing model's identity.
2. **Given** a recorded world whose memories carry vectors, **When** the world is
   replayed from its event log, **Then** the replayed output is byte-identical to the
   original and no embedding model is invoked at any point during replay.
3. **Given** the embedding provider is unavailable at emission time, **When** a memory
   is emitted, **Then** the memory is still recorded (without a vector), the failure is
   loudly visible to the operator, and the sim does not stall or crash.

---

### User Story 2 - Shadow-mode relevance scoring with rank-divergence instrumentation (Priority: P2)

Before relevance is allowed to change what agents think about, the system scores memory
selection both ways — the existing salience-and-recency ranking, and the new ranking that
adds situation-relevance — and records how much the two rankings disagree. An operator can
inspect this divergence over a run and make an evidence-based decision (against a
documented threshold) about whether the relevance term earns its place.

**Why this priority**: This is the analysis's second guardrail: in a world where days
repeat, relevance may just echo recency. Shipping the measurement before the behavior
change means the decision to turn relevance on is made from recorded evidence, not hope.
It is independently valuable: even if relevance were never enabled, the divergence report
answers a real design question.

**Independent Test**: Run a seeded world with shadow scoring enabled; confirm agent
behavior is bit-for-bit unchanged from a run without it; confirm a per-selection
divergence record exists and a summary can be produced for the run.

**Acceptance Scenarios**:

1. **Given** shadow scoring is active, **When** an agent's memory window is selected,
   **Then** the selection presented to the agent is exactly the legacy selection, and the
   divergence between the legacy and relevance-augmented rankings is recorded.
2. **Given** a completed run with shadow records, **When** the operator asks for the
   divergence summary, **Then** they receive a per-agent and whole-run view sufficient to
   judge the documented go/no-go threshold.
3. **Given** shadow scoring is active, **When** the world is replayed, **Then** replay
   remains byte-identical (divergence records are themselves recorded artifacts, not
   recomputed).

---

### User Story 3 - Relevance shapes the working-memory window (Priority: P3)

Once the divergence evidence clears the documented threshold, an agent facing a situation
recalls what *matters to that situation* — a memory semantically close to the moment at
hand can enter the agent's working memory even when it is neither recent nor loud. An
agent who was robbed at the river weeks ago, returning to the river, has that memory
surface; the same memory stays dormant across unrelated days.

**Why this priority**: This is the payoff — but it is deliberately last, gated on the
guardrail evidence from Story 2 and the foundation from Story 1. Turning it on early
would risk changing agent behavior on an unmeasured hypothesis.

**Independent Test**: Construct a scenario where an old, low-salience memory is
semantically close to the agent's current situation but excluded by the legacy ranking;
verify it enters the window when relevance is enabled, and verify windows on unrelated
ticks are unchanged or near-unchanged.

**Acceptance Scenarios**:

1. **Given** relevance is enabled and an agent holds an old, situation-matching memory
   that the legacy ranking excludes, **When** the agent's window is selected for a
   matching situation, **Then** the memory appears in the window.
2. **Given** relevance is enabled, **When** windows are selected, **Then** recency
   continues to decay from each memory's creation time only — selecting or surfacing a
   memory never alters that memory's future ranking (no state mutation on read).
3. **Given** two agents with semantically similar private memories, **When** either
   agent's window is selected, **Then** only that agent's own memories are ever
   candidates — cross-agent retrieval is impossible by construction.
4. **Given** relevance is enabled, **When** a world is replayed, **Then** the replay is
   byte-identical: every selection is a pure function of recorded vectors.

---

### Edge Cases

- **Legacy memories without vectors** (worlds recorded before this feature, or emissions
  during a provider outage): selection must handle vectorless memories deterministically
  — they participate with neutral relevance rather than being excluded or crashing.
- **Model change mid-lineage**: vectors from different embedding models are not
  comparable. If the pinned model changes, the system must refuse to compare
  cross-model vectors (neutral relevance across the boundary) rather than silently
  producing garbage similarity.
- **Memory text exceeding the model's input limit**: truncation must be deterministic so
  the same memory always yields the same vector.
- **No situation vector available** for a selection (e.g., a consumer without a recorded
  query): selection falls back to the legacy ranking, deterministically.
- **Agent with fewer memories than the window size**: behavior unchanged from today (all
  memories returned).
- **Embedding latency spike at emission**: memory emission must not block the tick loop
  indefinitely; the loud-failure path (US1 scenario 3) bounds the wait.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Every episodic memory MUST carry an embedding vector and the identity of
  the model that produced it, recorded at emission as part of the durable event record.
- **FR-002**: Replay MUST NOT invoke any embedding computation; all vectors and all
  scores derived from them MUST be pure functions of recorded data, preserving
  byte-identical replay.
- **FR-003**: Memory window selection MUST support a relevance term defined as the
  similarity between a memory's recorded vector and a recorded situation (query) vector,
  combined with the existing salience and recency terms.
- **FR-004**: Recency MUST continue to decay from a memory's creation time only.
  Selection and retrieval MUST NOT mutate any memory state (no last-access decay, no
  access counters affecting rank).
- **FR-005**: Embedding storage and retrieval MUST be per-agent by construction: no
  selection, scoring, or similarity computation may read another agent's memories or
  vectors. A test MUST prove two agents with overlapping experiences cannot influence
  each other's selection.
- **FR-006**: Rank-divergence instrumentation MUST ship before, and operate without,
  any behavior change: while in shadow mode, agents receive exactly the legacy
  selection, and the divergence between legacy and relevance-augmented rankings is
  recorded per selection.
- **FR-007**: The decision to let relevance drive prompts MUST be gated on a documented
  divergence threshold and recorded as a durable decision artifact (the gate may pass or
  fail; both outcomes are recorded).
- **FR-008**: The salience table and every salience-consuming threshold (cognition bump,
  gossip gating, governance evidence) MUST be unchanged in semantics and value by this
  feature.
- **FR-009**: The embedding model MUST be pinned per world lineage (identity recorded);
  vectors from different models MUST never be compared — cross-model comparisons resolve
  to neutral relevance.
- **FR-010**: Failure to obtain a vector at emission MUST be loud (operator-visible) and
  non-fatal: the memory is still recorded, vectorless, and participates in selection
  with neutral relevance.
- **FR-011**: Memory text preparation for embedding (including any truncation) MUST be
  deterministic: identical text always produces one identical, recorded vector per
  pinned model.

### Key Entities

- **Episodic memory**: an agent's recorded experience; gains a meaning vector and
  producing-model identity at emission. Existing attributes (text, salience, tick,
  subject, tone, where, why, origin) are unchanged.
- **Situation (query) vector**: a recorded representation of an agent's current moment
  (what it perceives and intends), captured when cognition is prepared; the anchor the
  relevance term compares against.
- **Divergence record**: a durable per-selection artifact capturing how the legacy and
  relevance-augmented rankings differ, aggregable into a per-run summary.
- **Model identity**: the pinned embedding model's name/version, recorded with every
  vector and with the world lineage; the comparability boundary for vectors.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A world recorded with this feature replays byte-identically, with zero
  embedding computations during replay — verified by the existing replay gate.
- **SC-002**: In a standard seeded run, 100% of newly emitted episodic memories carry a
  vector and model identity (absent deliberate provider-outage injection).
- **SC-003**: A rank-divergence summary covering at least one full simulated day exists
  and is reviewable before any run where relevance affects agent prompts.
- **SC-004**: In the relevance-enabled scenario test, an old, low-salience,
  situation-matching memory enters the agent's working-memory window that the legacy
  ranking provably excluded.
- **SC-005**: Sim throughput with embedding-at-emission enabled stays within 10% of the
  baseline wall-clock per game-day on the reference machine.
- **SC-006**: The isolation test demonstrates that removing or altering one agent's
  memories has zero effect on any other agent's selection output.

## Assumptions

- The situation (query) vector is derived from the same situation text the agent's
  cognition already sees (perception summary plus active intent), so "relevant to now"
  means relevant to what the agent is currently experiencing and trying to do. The exact
  composition is a planning-phase decision; the spec only requires it be recorded.
- A small local embedding model (384-dimension class) is assumed; hosted embedding APIs
  are out per the grounding (non-deterministic, network-dependent). Model choice and
  integration mechanism are planning-phase decisions.
- Vector storage rides in the event log / world state (roughly 1.5 KB per memory
  uncompressed at 384 dims); village-scale worlds make this immaterial. Compression/
  quantization is a planning-phase option, not a requirement.
- Build-order step 3 of the analysis (consolidation-time clustering and habituation —
  TASK-99) is explicitly out of scope, as is any change to write-time salience
  assignment (deferred indefinitely per the analysis).
- Further open-ended (LLM-authored) memory text is out of scope; today's situated,
  varying memory texts are assumed sufficient for the relevance term to be meaningful,
  with the divergence instrumentation (US2) as the check on that assumption.
- The go/no-go divergence threshold is deliberately not fixed by this spec; it is set
  and documented as part of the US2→US3 gate decision.
- Consumers of the working-memory window other than cognition prompts (e.g., soul.md
  rendering) continue to receive the legacy selection until explicitly migrated.
