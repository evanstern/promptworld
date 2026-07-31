# Tasks: Guardian missions (TASK-158)

**Input**: `specs/107-guardian-missions/spec.md` (D1–D5 + rulings ratified).

## Phase 1: Mission substrate

- [X] T001 Mission artifacts (084 discipline): guardian.mission_* events,
  reducer arms, door-validated acceptance, one-way terminals, prune (FR-001).
- [X] T002 Derived completion/failure from spec-084 predicates + recorded
  events; report-card integration (FR-003).

## Phase 2: Pursuit + doctrine

- [X] T003 Scheduled-lane pursuit: mission context in steward prompts;
  full-competence pursuit at any ceiling scoped to the active mission;
  order-door single-arbiter preserved (FR-002, SC-003).
- [X] T004 EASY-mode default charter clause; skinned-refusal separation
  (FR-004).

## Phase 3: Surfaces + tests

- [X] T005 Digest, event-types.md, decision trail; TestCatalogSweep (FR-005).
- [ ] T006 FR-006 test suite; go test -race ./... green; replay byte-identity.

## Phase 4: Evidence + grounding

- [ ] T007 In-branch obedience eval (FR-008): old vs new default, scripted
  direct-mission prompts via the measurement proxy; results in evidence doc.
- [ ] T008 Live demo (FR-007, SC-001) on a seeded measure world; evidence doc
  complete; 164-instrument mission-scenario harness prepared (US3).
- [ ] T009 Wiki re-pins (+ new guardian-missions.md), player-docs probe,
  merge-drift pr gate exit 0.
