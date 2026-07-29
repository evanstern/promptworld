# Feature Specification: Entity-lookup seam + store-error posture decision

**Feature Branch**: `task-76-entity-lookup-seam`

**Created**: 2026-07-29

**Status**: Draft

**Input**: TASK-76 — 2026-07-22 team review improvement 6 (latent scaling walls);
drift audit 2026-07-23: pileAt at state.go:240, chestAt/structureAt at
internal/sim/terrain.go:107/:85, exactly 7 executor.go call sites
(97,545,570,640,644,786,838 — re-verify, lines have moved since), rot sweep at
executor.go:151. Fatal store write loop.go:~398, no retry seam.

## Decisions

- **D1 — Seam only, no index.** All positional entity lookups (piles, chests,
  structures, and any other O(n) positional scan found during the audit) route
  through ONE accessor type, so a grid/spatial index can drop in later without
  touching call sites. The accessor's v1 implementation is the existing linear
  scan, byte-for-byte semantics INCLUDING tie-break ordering.
- **D2 — Store-error posture: fatal-by-doctrine STANDS, with a recorded
  rationale note.** Decision (derived from the card's own framing and the
  determinism doctrine): an unwritable log must never silently diverge from
  state; a bounded-retry seam adds a liveness/consistency tradeoff nobody has
  needed at current scale (no observed incident on this repo). Deliverable is
  the durable decision note (wiki operational note), naming the re-open
  trigger: the first real-world transient-write incident, or multi-world
  hosting. NO retry implementation ships.

## User Scenarios & Testing *(mandatory)*

### US1 - The future index has a socket to plug into (Priority: P1)

As a future spatial-indexing task, I want every positional lookup behind one
accessor, so dropping in a grid index is a one-file change.

**Acceptance Scenarios**:

1. **Given** the refactor, **Then** zero raw positional slice scans remain at
   the former call sites (grep-clean for direct pileAt/chestAt/structureAt
   usage outside the accessor; the rot sweep iterates through the accessor).
2. **Given** the determinism harness, **Then** replay of existing logs is
   bit-identical across the refactor (accessor returns identical results incl.
   tie-break ordering).

---

### US2 - The fatal posture is a decision, not an accident (Priority: P2)

**Acceptance Scenarios**:

1. **Given** the wiki operational note, **Then** it records D2 (fatal stands),
   the rationale, and the named re-open triggers; the loop.go site comments
   point at it.

## Requirements *(mandatory)*

- **FR-001**: One accessor type owning all positional entity lookups; v1 = the
  existing scans, semantics-identical (D1).
- **FR-002**: All former call sites (executor + rot sweep + any audit finds)
  routed through it; grep-clean.
- **FR-003**: Determinism proof: bit-identical replay of existing fixtures;
  go test -race ./... green.
- **FR-004**: D2 decision note in the wiki + site comment; no retry code.
- **FR-005**: Wiki re-pins; player-docs probe; merge-drift pr gate exit 0.

## Success Criteria *(mandatory)*

- **SC-001**: Replay fixtures bit-identical pre/post refactor.
- **SC-002**: A one-file index prototype could implement the accessor interface
  without touching any call site (demonstrated by the interface shape, not a
  shipped prototype).
- **SC-003**: The store-error posture is durably recorded with re-open triggers.

## Assumptions

- Tier: Sonnet — mechanical seam refactor + harness + doc note (routine).
- Tail task per the runbook: merges LAST among sim-touching PRs; droppable.
