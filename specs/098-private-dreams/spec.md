# Feature Specification: Consolidation clustering + habituation — the private dream phase

**Feature Branch**: `task-99-private-dreams`

**Created**: 2026-07-29

**Status**: Draft

**Input**: TASK-99 — build-order step 3 of the memory-retrieval vault analysis
(research/Agent-Memory-Retrieval/Analysis-Embedding-Retrieval-Adoption.md;
Novelty-Gates-and-Habituation.md; Consolidation-and-Clustering.md). Depends on
TASK-98 (Done): record-at-emission vectors exist. Framing (2026-07-24): this
phase is dream-like — nightly, reweights and rewrites memories — and dreams must
be PRIVATE.

## Design decisions

- **D1 Privacy by construction.** Clustering operates on ONE agent's memory
  store per pass; the pass's inputs are that agent's vectors alone. No shared
  vector table, no cross-agent index anywhere in the pipeline — proven by a test
  in which two agents with overlapping experiences consolidate to independent
  outcomes (card AC#1).
- **D2 Cheap geometry first (SAGE/RecMem patterns).** Embedding-space density
  detection finds near-duplicate clusters; clear-cut cluster/no-cluster cases
  resolve by geometry alone (thresholds as dials); only the ambiguous band may
  consult the consolidation LLM slot the nightly phase already owns. No new LLM
  call classes.
- **D3 Outcomes land as recorded events.** Habituation (down-weight) and merge
  outcomes are events (salience-revision / memory-merge under the existing
  cognition-telemetry family shape), replay-safe, emitter-computes. The reducer
  applies recorded outcomes; it never re-derives cluster decisions.
- **D4 Seeded noise — ADOPTED, minimally (card AC#4 decision).** A small
  rngAt-seeded jitter is applied to cluster-boundary scoring during the dream
  pass (purpose-keyed: seed, "dream", night-tick, agent index), so borderline
  clusters occasionally merge/survive differently run-to-run ACROSS SEEDS while
  remaining byte-identical within a log (rngAt is deterministic per seed+tick).
  Rationale: dream-like variance the framing asked for, at zero replay cost;
  bounded to boundary band only so habituation of true duplicates is stable.
  Recorded here as the design decision the AC demands; the dial can be zeroed.

## User Scenarios & Testing *(mandatory)*

### US1 - Repetition fades, novelty stays (Priority: P1)

As a villager who foraged the same berry patch forty times, I want my nightly
consolidation to detect that dense cluster and down-weight/merge it, so my
working window and retrieval stop drowning in near-duplicates while the one
night the gru chased me stays vivid.

**Acceptance Scenarios**:

1. **Given** a dense cluster of near-duplicate memories (embedding geometry),
   **When** the nightly pass runs, **Then** habituation down-weights/merges them
   via recorded events, replay-safe (D3); distinct memories are untouched.
2. **Given** a clear-cut case, **Then** geometry alone decides (no LLM call);
   **Given** an ambiguous-band case, **Then** at most the existing consolidation
   slot is consulted (D2), and the decision still lands as a recorded event.
3. **Given** the same log, **When** replayed, **Then** byte-identical (D4's
   noise is rngAt-seeded — deterministic per log).

---

### US2 - Dreams are private (Priority: P1)

As the epistemic-hygiene doctrine, I require that two agents with overlapping
experiences never influence each other's consolidation — no shared dreams.

**Acceptance Scenarios**:

1. **Given** two agents with substantially overlapping experience streams,
   **When** both consolidate, **Then** each outcome derives only from that
   agent's own store (test proves independence by perturbing one agent's
   memories and observing zero effect on the other's consolidation — card AC#1).

---

### US3 - The dials are tunable and the behavior observable (Priority: P2)

**Acceptance Scenarios**:

1. **Given** the tuning manifest, **Then** cluster density threshold, ambiguous
   band width, habituation weight factor, merge cap per night, and dream-jitter
   amplitude (zeroable) are dials (spec 048).
2. **Given** the chronicle/decision surfaces, **Then** salience-revision/merge
   events are documented (event-types.md) and digest-rendered.

## Requirements *(mandatory)*

- **FR-001**: Per-agent clustering pass in the nightly consolidation slot over
  TASK-98's recorded vectors; strictly single-store inputs (D1).
- **FR-002**: Geometry-first routing with ambiguous-band LLM fallback within the
  existing consolidation slot only (D2); no new cognition classes.
- **FR-003**: Habituation/merge outcomes as recorded events (existing telemetry
  family shape), reducer applies without re-derivation (spec 092 doctrine);
  additive types, no format-version bump (spec 094 doctrine).
- **FR-004**: rngAt-seeded boundary jitter, purpose-keyed, dial-zeroable (D4).
- **FR-005**: Dials in the tuning manifest; event-types.md + digest entries;
  TestCatalogSweep green.
- **FR-006**: Tests: privacy independence (US2), habituation/merge correctness,
  geometry-vs-LLM routing, replay byte-identity, -race green.

## Success Criteria *(mandatory)*

- **SC-001**: On a seeded world with a manufactured duplicate-heavy stream, the
  post-night store shows the cluster collapsed (weights/merges recorded) and
  retrieval surfaces the distinct memory over the habituated mass.
- **SC-002**: The privacy perturbation test proves zero cross-agent influence.
- **SC-003**: Existing replay fixtures byte-identical; noise dial zeroed
  reproduces pre-noise outcomes exactly.

## Assumptions

- TASK-98 delivered record-at-emission vectors and the retrieval relevance term;
  the nightly consolidation slot exists (TASK-9 lineage) with an LLM budget this
  feature reuses rather than extends.
- Tier: **Opus** — internal/mind consolidation orchestration + replay-doctrine
  surface (recorded-outcome events, seeded noise).
