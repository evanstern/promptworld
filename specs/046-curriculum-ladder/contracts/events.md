# Contract: curriculum event types

New namespace `curriculum.*` — requires a `familyByNamespace` row, digestRegistry
entries, catalog fixture rows, and `docs/wiki/event-types.md` rows (TestCatalogSweep
enforces all of it).

## `curriculum.exercise_passed`

- **Emitter (production)**: TASK-119's rubric machinery — executor emission class,
  pure function of (state, tick), the `metatron.order_expired` precedent. **No
  whitelist entries** (never mind- or operator-injected). Until 119 lands, only test
  fixtures emit it; this feature ships the type, reducer, catalog wiring, and
  consumers.
- **Payload**: `{exercise, stage, tick, evidence []EvidenceRef}` — outcome-shaped;
  evidence lists the satisfying events (rubric terms; for stage-2 gates, the
  charter-fingerprint observation in force).
- **Reducer**: records the pass on state (bounded), enabling replay-derived audit.
- **Charter evidence derivation (reconciled with spec 044, T022)**: spec 044 US2's
  `metatron.charter_observed` (specs/044-run-outcomes-morgue/contracts/events.md on main) carries
  `CharterObservedPayload{fingerprint, default}` where `default == true` means the
  world's default/preset charter was in force. An `EvidenceRef` for it keeps this
  contract's `custom` flag, but `custom` is DERIVED, never asserted freehand:
  `sim.CharterObservedEvidence` is the single sanctioned constructor and sets
  `custom = !payload.default` — inverted polarity, so a stage-1 tutor-preset
  world's observation (`default: true`, the game's authorship) can never satisfy
  the stage-2→3 gate conjunct (SC-004).

## `curriculum.stage_unlocked`

- **Emitter**: same emission class, derived from a pass satisfying a gate's conjuncts;
  exactly once per (world, stage).
- **Payload**: `{stage, exercise, tick}`.
- **Reducer**: latches the per-world unlocked fact.
- **Consumers**: chronicleNote case (narrated in-game — FR-009), TUI digest row,
  daemon-side observer that upserts the per-user unlocks record, status surfaces.

## Ordering / determinism

Both types are replay-deterministic (pure emission over recorded state); unknown-type
no-op reducers keep old binaries safe; replaying a world reproduces its passes and
unlocks (the per-user record is a projection of them, never an input).
