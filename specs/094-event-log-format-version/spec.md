# Feature Specification: Event-log format_version + translating migration — and the guardian rename through it

**Feature Branch**: `task-134-event-log-format-version`

**Created**: 2026-07-29

**Status**: Draft

**Input**: TASK-134 (HIGH) + operator rulings: metatron.* persisted names get
MIGRATED, not aliased (2026-07-25 team review); TASK-134 ships the REAL rename +
TASK-121 shim removal (orchestrator checkpoint, 2026-07-29, board-sweep runbook).

## What exists vs what's missing (grounded 2026-07-29, this branch's base)

- EXISTS: world-manifest `FormatVersion` (currently 5, internal/world/world.go:31)
  with `promptworld migrate` as driver and the snapshot-cut migration pattern
  (fresh log whose single `world.migrated` event carries the full transformed
  state; old DB archived, never overwritten).
- MISSING (the card's gap): the LOG itself carries no schema stamp — a log is not
  self-describing; and no migration mode can TRANSLATE a log (preserving event
  history) rather than snapshot-cutting it away. Renaming persisted event types
  without that is a one-way replay-compat door: the only guard today is a value
  pin in recipes_test.go:75-76.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The log says what schema it speaks (Priority: P1)

As the daemon (and any tool opening a log), I want the event log to carry a
format_version stamp written at genesis — and consulted on load — so a log is
self-describing independent of its manifest, and version-mismatched logs are
refused with the migrate hint instead of mis-replaying.

**Acceptance Scenarios**:

1. **Given** a fresh world, **When** genesis writes the log, **Then** the log
   carries the current log-format stamp (card AC#1: written at genesis and on
   every append path that needs it — if the stamp is a header/meta row, appends
   need nothing; if per-segment, each segment carries it).
2. **Given** a pre-stamp log (e.g. world-01-shaped), **When** opened by the fixed
   binary, **Then** it is recognized as the implicit legacy version and handled
   per US2 — never silently replayed under wrong semantics.
3. **Given** a log stamped NEWER than the binary understands, **When** opened,
   **Then** refusal with a clear upgrade hint (the world.go:51 posture: a future
   generation must never be mis-generated).

---

### User Story 2 - Translating migration preserves history (Priority: P1)

As an operator with recorded worlds, I want `promptworld migrate` to gain a
TRANSLATION mode — rewrite an old log into the new schema (here: event-type
renames), preserving every event, tick, and payload — so history survives schema
breaks that are pure renames, unlike the snapshot-cut which archives it.

**Acceptance Scenarios**:

1. **Given** an existing pre-rename log, **When** migrated, **Then** the new log
   contains the same event sequence with translated type names and the new stamp;
   the source DB is archived per the existing never-overwrite guard.
2. **Given** the migrated log, **When** replayed, **Then** the state-hash sequence
   is byte-identical to replaying the source log on the pre-rename binary (card
   AC#2 — the migration provably changes nothing but names).
3. **Given** an already-migrated world, **When** migrate runs again, **Then** the
   idempotence/already-migrated guards behave like the existing migration modes.

---

### User Story 3 - The guardian rename ships through it (Priority: P1)

As the skinnable-guardian program (TASK-121), I want the ~13 persisted metatron.*
event types renamed to guardian.* IN PRODUCTION — reducer arms, emitters,
whitelists, digest grammar, expected-event sets — with old logs migrating via US2
and TASK-121's chronicle Type-column display alias REMOVED (the interim shim's
stated retirement condition).

**Acceptance Scenarios**:

1. **Given** the renamed binary, **When** a fresh world runs guardian activity,
   **Then** only guardian.* types are persisted; no metatron.* string remains in
   any emit/apply/whitelist path (the freeze annotations from spec 052 are
   retired).
2. **Given** a pre-rename world (seeded, with metatron.* events), **When**
   migrated and replayed, **Then** state hashes match the pre-rename replay
   (US2.2) and the world runs forward on the new binary.
3. **Given** the chronicle, **When** rendering both migrated and fresh worlds,
   **Then** the Type column shows guardian.* natively and the display-alias shim
   is deleted.
4. **Given** an UNMIGRATED pre-rename world, **When** the new binary opens it,
   **Then** refusal with the migrate hint (never a silent wrong-name replay).

---

### User Story 4 - The doctrine is recorded where changers will trip over it (Priority: P2)

As a future contributor, I want the doctrine — persisted-name changes and
reducer-re-derivation changes REQUIRE a format-version bump + migration — recorded
in the wiki (reconciled with TASK-75/spec 092's doctrine note, same sweep) and as
comments at the definition sites, so the next rename can't skip the machinery.

**Acceptance Scenarios**:

1. **Given** the wiki event-log/sim-state-reducer notes, **Then** the doctrine
   names the bump requirement, the translation-vs-snapshot-cut decision rule
   (pure rename ⇒ translate; semantic break ⇒ snapshot-cut), and cross-links
   spec 092's emitter-computes doctrine.
2. **Given** the event-type definition sites, **Then** comments state the bump
   requirement (replacing the spec-052 freeze annotations).

### Edge Cases

- The live playtest world (TASK-14, day 22) is NEVER migrated or opened by test
  tooling in this task; the pre-rename fixture is a disposable seeded world.
  Post-merge, the playtest daemon keeps running its current binary — migration of
  real worlds happens only when the operator chooses to restart/migrate them.
- recipes_test.go:75-76 value pin: superseded by the real guard — update the test
  to assert through the new versioning instead of a bare name pin.
- state.go hunt-yield re-derivation (the OTHER hazard on the card): documented by
  spec 092 (TASK-75) and listed in its audit; NOT migrated here (no behavior
  change to it in this task — its future change would use this machinery).
- Store DDL comment ("format_version 1 schema", internal/store/schema.go:3):
  reconcile terminology so world FormatVersion, store schema version, and the new
  log stamp are named distinctly in docs.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A log-level format-version stamp, written at genesis, readable
  without replay; pre-stamp logs resolve to an implicit legacy version.
- **FR-002**: Load-time enforcement: mismatched (older) logs refuse with the
  migrate hint; newer-than-binary logs refuse with the upgrade posture.
- **FR-003**: A translating migration mode in the existing migrate driver:
  event-type rename maps, every event/tick/payload preserved, archived source,
  never-overwrite + already-migrated guards, refuses a live daemon.
- **FR-004**: Byte-identity proof: replay(source log, old semantics) ==
  replay(translated log, new semantics) as state-hash sequences, test-enforced on
  a seeded fixture world.
- **FR-005**: The metatron.*→guardian.* rename of all persisted guardian event
  types, applied across emitters/reducer/whitelists/digest/tests; spec-052 freeze
  annotations retired; TASK-121 display-alias shim removed.
- **FR-006**: Doctrine recorded (wiki + definition-site comments), reconciled
  with spec 092; version-history comments in world.go pattern followed for the
  log stamp.
- **FR-007**: `go test -race ./...` green; replay/determinism harness green;
  TestCatalogSweep (story feed) green with the renamed types.

## Success Criteria *(mandatory)*

- **SC-001**: A seeded pre-rename world migrates and replays with byte-identical
  state hashes, then runs forward on the new binary (the end-to-end AC#4 demo, on
  the REAL rename).
- **SC-002**: Zero metatron.* persisted-type strings in emit/apply paths
  post-rename (grep-clean except migration maps + historical docs).
- **SC-003**: An unmigrated old world cannot be mis-opened (refusal tested both
  directions: old-log-new-binary, new-log-old-binary posture documented).
- **SC-004**: Doctrine present in wiki with the decision rule; TASK-75's spec 092
  note cross-links it.

## Assumptions

- The ~13 affected types are those carrying the metatron.* prefix at the freeze
  annotations (spec 052); the implementer enumerates them exactly from the freeze
  list, not from memory.
- The stamp's physical shape (header row vs meta event vs sqlite pragma/table) is
  the implementer's call within FR-001/FR-002 constraints — recorded in a short
  research note in this spec dir.
- World-manifest FormatVersion may or may not bump alongside (implementer's call:
  if the log stamp lives inside the DB, a manifest bump may be redundant; if
  bumped, the existing Open gating carries it) — decision recorded in the
  research note.
