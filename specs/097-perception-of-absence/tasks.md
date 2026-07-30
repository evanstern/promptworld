# Tasks: Perception of absence (TASK-80)

**Input**: `specs/097-perception-of-absence/spec.md` (D1–D5 ratified there).

## Phase 1: Observation channel

- [X] T001 `agent.place_observed`: executor emission on intent-completing
  arrivals, exhaustive within placeScanRadius, deterministic; reducer apply arm;
  event catalog entry (FR-001).
- [X] T002 Situated observation memory: low base salience, dedup window,
  "observed" provenance (FR-002); dials in the tuning manifest (FR-004).

## Phase 2: Belief reconciliation (mind-side)

- [X] T003 TASK-79-seam reconciliation: confirmation boost, bounded
  disconfirmation decay (faster than silence), silence unchanged; matching in
  internal/mind only (FR-003).

## Phase 3: Surfaces + tests

- [X] T004 Digest grammar entry + event-types.md; TestCatalogSweep green
  (FR-005).
- [X] T005 Tests: emission determinism, replay byte-identity on existing
  fixtures, dedup, all three belief paths (FR-006); go test -race ./... green.

## Phase 4: Soak + grounding

- [X] T006 One-game-day soak at 8x on a seeded MEASURE world (never playtest):
  bounded observation-memory counts + unchanged survival behavior; evidence at
  docs/design/evidence/task-80/ (US3, SC-002). Implementer prepares + runs with
  local-only LLM routes (no paid spend); orchestrator reviews evidence.
- [ ] T007 Wiki re-pins (prose amendments where perception/memory behavior is
  described); player-docs probe; merge-drift pr gate exit 0.
