# Tasks: Entity-lookup seam (TASK-76)

**Input**: `specs/099-entity-lookup-seam/spec.md` (D1/D2 ratified there).

## Phase 1: Seam

- [ ] T001 Audit all positional scans (pileAt/chestAt/structureAt + rot sweep +
  any others); define the accessor type; v1 = existing scans, tie-break
  identical (FR-001).
- [ ] T002 Route all call sites through it; grep-clean for direct scans
  (FR-002).

## Phase 2: Proof + decision

- [ ] T003 Determinism: bit-identical replay of existing fixtures;
  go test -race ./... green (FR-003).
- [ ] T004 D2 store-error decision note (wiki operational note + loop.go site
  comment; no retry code) (FR-004).

## Phase 3: Grounding

- [ ] T005 Wiki re-pins; player-docs probe; merge-drift pr gate exit 0 (FR-005).
