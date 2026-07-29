# Tasks: Event-log format_version + translating migration + guardian rename (TASK-134)

**Input**: `specs/094-event-log-format-version/spec.md` (operator rulings: migrate
not alias; ship the REAL rename here).

## Phase 1: Log stamp + enforcement

- [X] T001 Research note (specs/094-event-log-format-version/research.md): stamp's
  physical shape (header row / meta event / sqlite table-pragma), whether the
  world-manifest FormatVersion bumps alongside, and the exact enumeration of
  affected metatron.* persisted types from the spec-052 freeze annotations —
  decisions + rationale (spec Assumptions).
- [X] T002 Implement the log-level format-version stamp: written at genesis,
  readable without replay, implicit legacy version for pre-stamp logs (FR-001).
- [X] T003 Load-time enforcement: older ⇒ refuse with migrate hint; newer ⇒
  refuse with upgrade posture (FR-002; tests both directions).

## Phase 2: Translating migration

- [X] T004 Translation mode in the migrate driver: type-rename maps, every
  event/tick/payload preserved, archive + never-overwrite + already-migrated +
  live-daemon guards (FR-003).
- [X] T005 Byte-identity harness: replay(source, old) == replay(translated, new)
  as state-hash sequences on a seeded fixture world (FR-004).

## Phase 3: The guardian rename

- [X] T006 Rename all persisted metatron.* types to guardian.* across
  emitters/reducer arms/injection whitelists/digest grammar/expected-event sets;
  retire spec-052 freeze annotations; update recipes_test value pin to assert
  through the versioning (FR-005).
- [X] T007 Remove TASK-121's chronicle Type-column display-alias shim; chronicle
  renders guardian.* natively (US3.3).
- [X] T008 End-to-end: seeded pre-rename world → migrate → byte-identical replay →
  runs forward on new binary; unmigrated world refused with hint (SC-001/SC-003).

## Phase 4: Doctrine + grounding

- [X] T009 Doctrine in wiki (event-log / sim-state-reducer notes): bump
  requirement, translate-vs-snapshot-cut decision rule, cross-link spec 092;
  definition-site comments (FR-006).
- [ ] T010 Full gates: go test -race ./..., replay harness, TestCatalogSweep;
  re-pin all touched wiki notes; player-docs refresh if probed stale; merge-drift
  pr gate exit 0 (FR-007).
