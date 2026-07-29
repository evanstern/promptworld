# Feature Specification: Move-miracle target freshness — re-resolve the villager at the door

**Feature Branch**: `task-166-move-miracle-freshness`

**Created**: 2026-07-29

**Status**: Draft

**Input**: TASK-166 — carded from TASK-163's evidence
(docs/design/evidence/task-163/results.md, PR #128): 3 of 5 residual
privileged-action rejections on the fixed binary were position-freshness races.

## Decision (card AC #1)

**Chosen: (a) door-side name re-resolution — x/y become advisory when a villager
name is supplied.** When a `work_miracle` call has kind=move, class=villager, and a
villager NAME (the `villager` field already exists in `miracleArgs`,
internal/guardian/turn.go), the door re-resolves the villager's LIVE position at
validation/emission time and uses it as the move's source coordinates; the surveyed
x/y are advisory (used only when no name is given, preserving today's
coordinate-addressed form).

**Replay/determinism analysis** (the AC's required artifact):
- Emission stays emitter-computes: the RECORDED `metatron.entity_moved` event
  carries the RESOLVED source coordinates, so the reducer arm
  (`applyEntityMoved`, internal/sim/miracles.go:497 — "no living villager at
  (x,y)" check at :506) is byte-unchanged. Old logs replay identically (their
  events carry the coordinates that were valid at their tick); new events carry
  coordinates valid at emission. No format_version bump (no persisted-name change,
  no reducer re-derivation — TASK-75/134 doctrine).
- Validation precedes the charge exactly as today: resolution happens before the
  dry-run, so a dead/missing villager still refuses BEFORE charging.
- Determinism: resolution reads live sim state at the emission tick through the
  existing snapshot path — same inputs, same outcome; nothing wall-clock enters
  the event.

**Rejected alternatives**:
- (b) freshness token binding the call to the survey tick with a grace window —
  adds protocol state and a tunable window for a problem name-resolution solves
  outright; the token still races beyond its window at high speed.
- (c) guidance-only ("prefer name-addressed moves") — leaves the race in place;
  the evidence shows the model already forms well-formed calls, so the gap is
  architectural, not instructional. (A one-line guidance nudge still ships with
  (a) so models SUPPLY the name — see FR-004.)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A move lands on the villager, not the coordinates (Priority: P1)

As the guardian, when I decide to move Ash to the hearth and Ash keeps walking
while I think (at 8x, 30–60s of latency is 240–480 ticks), I want my move to follow
Ash — resolved by name at the door — so the act I was charged for lands on the
villager I named instead of refusing with "no living villager at (x,y)".

**Acceptance Scenarios**:

1. **Given** a move call with class=villager, a valid living villager name, and
   surveyed coordinates the villager has since left, **When** the door validates,
   **Then** the move lands using the villager's live position as source; the
   recorded event carries the resolved coordinates.
2. **Given** the same call naming a dead or unknown villager, **When** the door
   validates, **Then** it refuses BEFORE charging with a message naming the
   problem (existing refusal shape).
3. **Given** a villager-move with coordinates only (no name), **When** the door
   validates, **Then** today's exact coordinate-addressed behavior applies
   (including the race — coordinates remain a legal address form).
4. **Given** structure/pile moves, **When** validated, **Then** behavior is
   byte-identical to today (name resolution applies to villagers only).

---

### User Story 2 - Old worlds replay cleanly (Priority: P1)

As an operator with recorded worlds (including the live playtest), I want replay of
every pre-fix recorded move to apply exactly as before, so this fix is invisible to
history.

**Acceptance Scenarios**:

1. **Given** a pre-fix event log containing entity_moved events, **When** replayed
   on the fixed binary, **Then** the state hash sequence is byte-identical
   (reducer arm untouched).

---

### User Story 3 - Probe evidence on a live world (Priority: P2)

As the sweep's evidence trail, a live probe on a MEASUREMENT world (never the
running playtest world) demonstrates name-addressed villager moves landing at 8x.

**Acceptance Scenarios**:

1. **Given** a seeded measure world at 8x with the guardian on the TASK-163 probe
   recipe, **When** a name-addressed villager move is attempted after the villager
   moves during model latency, **Then** it lands; evidence appended to the
   TASK-166 card and docs/design/evidence/task-166/.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: kind=move + class=villager + villager name ⇒ the door resolves the
  villager's live position at validation/emission and uses it as the move source;
  x/y advisory. Name lookup uses the existing living-villager resolution
  (agentIndexByName semantics: unknown or dead ⇒ pre-charge refusal).
- **FR-002**: The recorded event carries resolved coordinates
  (emitter-computes); `applyEntityMoved` is unchanged; replay of pre-fix logs is
  byte-identical (regression-tested).
- **FR-003**: Coordinate-only villager moves and ALL structure/pile moves behave
  exactly as today.
- **FR-004**: The move guidance gloss (internal/tool/derive.go, TASK-163 pattern)
  gains one line telling the model to supply the villager's name for villager
  moves; the raced-refusal message suggests it too.
- **FR-005**: Tests: unit coverage for raced-move-lands (entity moved after
  survey), dead/unknown-name refusal before charge, coordinate-only path
  unchanged, structure/pile unchanged, and a replay byte-identity regression over
  a log containing pre-fix moves.

## Success Criteria *(mandatory)*

- **SC-001**: A raced villager-move (target moved after survey) lands under the
  chosen mechanism in tests and in the live probe.
- **SC-002**: Replay byte-identity proven over pre-fix recorded moves.
- **SC-003**: Probe evidence recorded (docs/design/evidence/task-166/) with the
  card updated — the TASK-163 residual class "position-freshness races" is
  extinct or reduced to coordinate-only calls.

## Assumptions

- The live probe uses the measurement-run recipe (seeded measure world,
  9router-proxied guardian route) — the TASK-14 playtest world is NEVER touched.
- The `villager` field in miracleArgs is already parsed; no tool-schema change is
  needed (guidance gloss only).
