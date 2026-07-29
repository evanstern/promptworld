# Tasks: Private dreams — consolidation clustering + habituation (TASK-99)

**Input**: `specs/098-private-dreams/spec.md` (D1–D4 ratified there).

## Phase 1: Clustering + habituation core

- [X] T001 Per-agent density clustering over TASK-98 vectors in the nightly
  slot; geometry-first thresholds + ambiguous band routing to the existing
  consolidation LLM slot (FR-001/FR-002).
- [X] T002 Habituation/merge outcomes as recorded events (salience-revision /
  memory-merge), reducer apply arms, memory-store weight application
  (FR-003).

## Phase 2: Noise + dials

- [X] T003 rngAt-seeded boundary jitter (purpose-keyed, zeroable dial) + all
  dials in the tuning manifest (FR-004, D4).

## Phase 3: Surfaces + tests

- [X] T004 event-types.md + digest entries; TestCatalogSweep green (FR-005).
- [X] T005 Tests: privacy perturbation independence, habituation/merge
  correctness, routing, replay byte-identity, noise-zeroed equivalence;
  go test -race ./... green (FR-006).

## Phase 4: Evidence + grounding

- [ ] T006 Seeded-world demonstration (local-only LLM routes, never the
  playtest): duplicate-heavy stream collapses, distinct memory survives;
  evidence at docs/design/evidence/task-99/ (SC-001).
- [ ] T007 Wiki re-pins (consolidation/memory/embedding notes); player-docs
  probe; merge-drift pr gate exit 0.
