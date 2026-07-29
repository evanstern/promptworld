# Feature Specification: Determinism scope note + reducer-constants replay-hazard doctrine

**Feature Branch**: `task-75-determinism-doctrine`

**Created**: 2026-07-29

**Status**: Draft

**Input**: TASK-75 — docs/doctrine PR, minimal code (2026-07-22 team review items;
re-verified by the 2026-07-25 review: no format_version exists, README.md:80 and
deterministic-rng.md:50 still claim per-seed determinism).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The determinism claim readers rely on is the true one (Priority: P1)

As a developer (or future task) about to build on "same seed ⇒ same world", I want
the wiki and README to state the ACTUAL guarantee — determinism is PER-LOG, not
per-seed across machines — so nobody builds a cross-machine determinism check on a
claim the code does not make (EffectiveRate is wall-clock-measured, lands in
clock.degraded events, and is baked into the canonical state hash).

**Acceptance Scenarios**:

1. **Given** the updated docs, **When** a reader consults deterministic-rng.md,
   sim-loop.md, or the README determinism paragraph, **Then** each states: replay
   of a given log is exact; two machines on the same seed may diverge (wall-clock
   EffectiveRate enters the hash); the per-seed claim is corrected everywhere it
   appears.
2. **Given** the doc changes, **When** the wiki freshness gate runs, **Then**
   every touched note is re-pinned and green.

---

### User Story 2 - The reducer-constants hazard is named doctrine, not tribal memory (Priority: P1)

As a contributor changing a gameplay constant, I want recorded doctrine —
**emitter-computes / payload-carries-the-outcome is the default; reducer-re-derives
is the exception requiring an explicit format_version note** — plus code comments
at the existing exception sites, so a retune can't silently break old-log replay.

**Acceptance Scenarios**:

1. **Given** the doctrine note (event-log and/or sim-state-reducer wiki note),
   **When** a reader checks it, **Then** it names the default, the exception, the
   TASK-134 pointer (the migration machinery that future work rides), and
   reconciles with the existing spec-019 emitter-computes precedent recorded at
   sim-state-reducer.md.
2. **Given** the audited exception sites, **When** a reader opens them, **Then**
   each carries a comment naming the hazard and pointing at the doctrine
   (hunt-yield re-derivation, state.go ~:596-600 per the card's drift audit —
   re-verify line positions at implementation time).

---

### User Story 3 - The audit list is complete (Priority: P2)

As TASK-134 (the migration machinery task, queued in this same sweep), I want an
audit list of ALL reducer-re-derive sites in the wiki note, so the migration work
knows its full surface (audit only — migrating is TASK-134's job, explicitly out
of scope here).

**Acceptance Scenarios**:

1. **Given** the wiki note, **When** TASK-134's implementer reads it, **Then** it
   lists every site where the reducer re-derives outcomes from mutable gameplay
   constants (found by systematic sweep of internal/sim reducer paths), each with
   file:line and the constant(s) involved.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Per-log vs per-seed determinism limit documented in
  docs/wiki/deterministic-rng.md, docs/wiki/sim-loop.md (or the note that owns
  EffectiveRate/clock.degraded), and README's determinism paragraph; all incorrect
  per-seed claims corrected.
- **FR-002**: Reducer-constants doctrine recorded in the event-log or
  sim-state-reducer wiki note: emitter-computes default, reducer-re-derives
  exception requires format_version bump + migration (TASK-134 pointer), spec-019
  precedent reconciled, spec-048 genesis-tuning-pin mitigation and its residual
  scope (pre-057/migrated worlds) noted per the card's 2026-07-25 pointer.
- **FR-003**: Code comments at each audited re-derive site naming the hazard and
  doctrine home. Comment-only changes; no behavior change (go test ./... green,
  gofmt clean).
- **FR-004**: Audit list of all reducer-re-derive sites in the wiki note
  (file:line + constants), produced by sweeping the reducer's Apply arms.
- **FR-005**: All touched wiki notes re-pinned; docs/player regenerated if the
  probe demands (README is a pinned player-docs input); merge-drift pr gate green.

## Success Criteria *(mandatory)*

- **SC-001**: No doc in the repo claims per-seed cross-machine determinism.
- **SC-002**: The doctrine + audit list exist in exactly one wiki home each,
  cross-linked; TASK-134 can consume the list as-is.
- **SC-003**: Zero behavior change: comment-only code diff; tests green.

## Assumptions

- Card drift-audit anchors (2026-07-23/25): EffectiveRate loop.go:567/578,
  State.EffectiveRate state.go:35 reducer :369, hunt-yield state.go:596-600,
  README.md:80, deterministic-rng.md:50 — implementer re-verifies before editing.
- TASK-134 (same sweep, later in lane 1) supplies the machinery; this task's
  doctrine lands first per the cards' cross-reference.
