# Feature Specification: World Tuning Manifest (tuning.json)

**Feature Branch**: `048-tuning-manifest`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "World tuning manifest: a boot-loaded, clamp-validated, event-logged tuning.json in the world dir as the promotion path for doctrine constants (docs/design/control-surface-and-calibration.md §6, TASK-107). Every field defaults to the current doctrine constant; absent file means current behavior exactly. Per-field validation clamps follow the max_tokens/loop_max_rounds pattern in llm/config.go. Applied values are emitted as events at boot (and on change) so replays reproduce behavior — the calibration.json/format_version discipline. First five promoted dials: refuelDyingBelow, fireBurnPerWood, gruEmergePerMille, PlannerCadenceTicks, conversation pair cooldown — these consume the manifest instead of consts. Design report §6 updated to point at the mechanism. Follow-on (recorded decision): once world-01 is dialed in, tuned values become the standard default for new worlds."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Operator tunes a dial without editing code (Priority: P1)

The operator wants to change a behavior dial on an existing world (e.g., make fires
burn longer per wood, or make villagers refuel earlier). They place or edit a
`tuning.json` file in the world directory, restart the daemon, and the world runs
with the tuned values. No file present means the world behaves exactly as it does
today — byte-for-byte identical behavior to the current doctrine constants.

**Why this priority**: This is the whole point of the feature — the promotion path
from "hand-edit a constant and rebuild" to "edit a per-world manifest and restart."
Every other story depends on the manifest existing and being read.

**Independent Test**: Boot a world with no `tuning.json` and confirm behavior is
unchanged; boot the same world with a `tuning.json` overriding one dial and confirm
the new value is in effect.

**Acceptance Scenarios**:

1. **Given** a world directory with no `tuning.json`, **When** the daemon boots,
   **Then** every dial has its current doctrine-constant value and the world's
   behavior is indistinguishable from today's.
2. **Given** a `tuning.json` that sets one dial to a valid in-range value, **When**
   the daemon boots, **Then** that dial takes the manifest value and all other
   dials keep their defaults.
3. **Given** a `tuning.json` with a value outside its documented bounds, **When**
   the daemon boots, **Then** the value is clamped to the nearest bound, the
   clamped (effective) value is what gets applied and recorded, and the operator
   sees a warning naming the field, the raw value, and the clamp applied.
4. **Given** a `tuning.json` that is unreadable (malformed JSON) or contains an
   unrecognized field name, **When** the daemon boots, **Then** boot fails with a
   clear error naming the problem — the world never silently runs with values the
   operator didn't intend.

---

### User Story 2 - Replays reproduce tuned behavior (Priority: P1)

A world that ran under tuned values must replay identically from its event log,
even if `tuning.json` has since been edited, deleted, or the world directory was
copied elsewhere. The effective tuning values are recorded in the event log at the
moment they take effect, and replay consumes the logged values — never the file.

**Why this priority**: The world is event-sourced; a dial that changes reducer
behavior without leaving a trace in the log breaks replay determinism, which is a
hard invariant of the project. This is co-P1 with Story 1 because shipping the
manifest without it would be unsafe-by-default — exactly what §6 exists to prevent.

**Independent Test**: Run a world with a tuned dial past events that the dial
affects, then delete `tuning.json` and replay the log; the replayed state must
match the live state hash-for-hash.

**Acceptance Scenarios**:

1. **Given** a boot where the effective tuning values differ from the values
   currently in effect in the world's event-sourced state, **When** the daemon
   applies them, **Then** an event recording the full set of effective values is
   appended to the log before any tick runs under them.
2. **Given** a boot where the effective tuning values are identical to those
   already in effect, **When** the daemon boots, **Then** no redundant tuning
   event is appended (the log does not grow on every restart).
3. **Given** an event log containing tuning events, **When** the world is replayed
   with `tuning.json` absent or changed, **Then** replay applies the values from
   the events and reproduces the original run exactly.
4. **Given** an existing world whose log predates this feature, **When** it is
   replayed, **Then** the doctrine-constant defaults apply until the first tuning
   event (if any), and replay of the old segment is unchanged.

---

### User Story 3 - The five earned dials consume the manifest (Priority: P2)

The five constants that earned promotion on world-01 evidence are actually driven
by the manifest: the refuel-dying threshold, fire burn time per wood, nightly gru
emergence chance, planner cadence, and the conversation pair (encounter) cooldown.
Turning any of them in `tuning.json` changes the corresponding behavior.

**Why this priority**: These are the concrete dials the operator has been wanting
to turn; without them the mechanism is scaffolding with nothing attached. P2 only
because the mechanism (Stories 1–2) must exist first.

**Independent Test**: For each dial, set a non-default value in `tuning.json`,
boot, and observe the changed behavior (e.g., a fire's fuel deadline moves with
the tuned burn value; planners run on the tuned cadence).

**Acceptance Scenarios**:

1. **Given** a tuned refuel threshold, **When** a fire's remaining fuel drops
   below the tuned value (not the old constant), **Then** the refuel reflex
   triggers.
2. **Given** a tuned burn-per-wood value, **When** a fire is built or refueled,
   **Then** its fuel deadline is computed from the tuned value.
3. **Given** a tuned gru emergence chance, **When** night falls, **Then** the
   emergence roll uses the tuned per-mille value.
4. **Given** a tuned planner cadence, **When** the mind layer schedules planner
   turns, **Then** the tuned cadence drives the schedule and agent stagger.
5. **Given** a tuned conversation pair cooldown, **When** two agents who recently
   conversed become adjacent again, **Then** the tuned cooldown gates the new
   encounter.

---

### User Story 4 - The design report points at the mechanism (Priority: P3)

A future reader of the control-surface report's §6 (the promotion path) finds a
pointer to the shipped mechanism: where the manifest lives, which dials are
promoted, and what the promotion steps now look like in practice.

**Why this priority**: Documentation follow-through; cheap, but the report is the
governance record and must not describe the promotion path as hypothetical once
it exists.

**Independent Test**: Read §6 of the report and confirm it names `tuning.json`,
the five promoted dials, and the event-logging discipline as shipped mechanism
rather than proposal.

**Acceptance Scenarios**:

1. **Given** the shipped feature, **When** §6 of
   `docs/design/control-surface-and-calibration.md` is read, **Then** it points
   at the manifest mechanism and the five promoted dials as implemented.

---

### Edge Cases

- Manifest present but empty (`{}`): all dials keep their defaults; boot succeeds;
  no tuning event is appended if defaults are already in effect.
- A dial set exactly to its default value: treated as "in effect" like any other
  value — no event if the world is already running defaults.
- Out-of-bounds value (e.g., a zero or negative cadence): clamped to the
  documented bound, warning printed, clamped value logged and applied.
- Malformed JSON, wrong types, or unknown field names: boot fails loudly with the
  file path and the specific problem. Typos must not silently no-op.
- `tuning.json` edited while the daemon is running: no effect until next boot.
  Hot (mid-run) exposure is explicitly out of scope — it is step 3 of the §6
  promotion path and comes later, if ever, per dial.
- World copied/moved with its `tuning.json`: behaves the same on the new host;
  replay is already file-independent via the logged events.
- Replay tooling and any secondary event consumers encounter the new tuning event
  kind in logs: they must tolerate it (at minimum, not crash; ideally surface it).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST read an optional `tuning.json` from the world
  directory at daemon boot. When the file is absent, every dial MUST have exactly
  its current doctrine-constant value and observable behavior MUST be unchanged.
- **FR-002**: Every manifest field MUST have a documented default (the current
  doctrine constant) and a documented validation clamp (min/max). Out-of-range
  values MUST be clamped to the nearest bound with an operator-visible warning;
  the clamped value is the effective value.
- **FR-003**: A manifest that cannot be parsed, has wrong-typed values, or
  contains unrecognized field names MUST fail boot with an error naming the file
  and the problem (fail-closed; no silent fallback to defaults).
- **FR-004**: The effective tuning values MUST be recorded in the world's event
  log whenever they differ from the values currently in effect in the
  event-sourced state, before any simulation work runs under the new values.
  Identical values MUST NOT append a redundant event on restart.
- **FR-005**: Replay MUST derive tuning values exclusively from the event log
  (defaults until the first tuning event), never from `tuning.json`, so that
  replays reproduce the original run regardless of the file's current contents
  or absence.
- **FR-006**: The five promoted dials MUST consume the manifest-driven values
  instead of hard constants: refuel-dying threshold, fire burn per wood, gru
  nightly emergence per-mille, planner cadence, and conversation pair (encounter)
  cooldown.
- **FR-007**: Worlds and event logs created before this feature MUST load and
  replay unchanged (defaults in effect until a tuning event appears, which for
  old logs is never).
- **FR-008**: §6 of `docs/design/control-surface-and-calibration.md` MUST be
  updated to point at the shipped mechanism (manifest location, promoted dials,
  event discipline), replacing the hypothetical description.

### Key Entities

- **Tuning manifest (`tuning.json`)**: an optional per-world file of named dials;
  partial by design (set only what you want to change). Fields: the five promoted
  dials. Each field has a default (current doctrine constant) and a clamp range.
- **Tuning event**: an event-log record of the full set of effective dial values
  at the moment they take effect; the replay-authoritative source of tuning state.
- **Promoted dial**: a former doctrine constant now driven by the manifest. The
  initial five and their defaults: refuel-dying threshold (1 game-hour of fuel
  remaining), fire burn per wood (4 game-hours), gru emergence chance (600 per
  mille per night), planner cadence (30 game-minutes), conversation pair cooldown
  (2 game-hours).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A world with no `tuning.json` produces behavior indistinguishable
  from before the feature: existing determinism/replay test suites pass unchanged
  and no new events appear in its log at boot.
- **SC-002**: The operator can change any of the five promoted dials on an
  existing world with a file edit plus restart — zero code changes, zero
  rebuilds — and observe the changed behavior within the first affected tick.
- **SC-003**: A tuned world's event log replays to the same state hash as the
  live run even after `tuning.json` is deleted or altered post-run.
- **SC-004**: 100% of out-of-range manifest values are clamped and warned about;
  100% of malformed/unknown-field manifests are rejected at boot with an error
  that names the offending field or parse problem.
- **SC-005**: All five dials demonstrably respond to the manifest (one test per
  dial proving the tuned value, not the old constant, drives the behavior).

## Assumptions

- "Conversation pair cooldown" refers to the per-pair encounter cooldown in the
  mind layer (2 game-hours), per the report's §2.6 inventory — not the sim-side
  ambient talk cooldown, which stays a constant for now.
- Fail-closed on malformed/unknown fields is chosen over warn-and-continue: a
  typo'd field name that silently no-ops would defeat the manifest's purpose
  (the operator believes a dial moved when it didn't). Clamping (not rejecting)
  handles merely out-of-range values of correctly named fields.
- Tuning events record the full effective set (not deltas) so a replay can
  establish tuning state from any single tuning event without scanning history.
- Boot-time application only. Hot exposure (IPC/angel/TUI) is §6 step 3 and out
  of scope; so is folding tuned world-01 values back into new-world defaults
  (recorded follow-on decision, a separate future change).
- The manifest governs per-world behavior; `promptworld new` does not write a
  `tuning.json` (absent file is the default state for new worlds).
- The daemon's existing boot logging is the operator-visible channel for clamp
  warnings and applied-value reporting; no new UI surface is required.
