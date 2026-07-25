# Phase 1 Data Model: World Tuning Manifest

## TuningState (sim, event-sourced)

The single source of truth for effective dial values, carried on `sim.State`:

```
State.Tuning *TuningState `json:"tuning,omitempty"`
```

| Field | Go type | JSON key | Default | Clamp |
|---|---|---|---|---|
| RefuelDyingBelow | int64 | `refuel_dying_below` | 3600 | [0, 86400] |
| FireBurnPerWood | int64 | `fire_burn_per_wood` | 14400 | [600, 86400] |
| GruEmergePerMille | uint64 | `gru_emerge_per_mille` | 600 | [0, 1000] |
| PlannerCadenceTicks | int64 | `planner_cadence_ticks` | 1800 | [60, 86400] |
| EncounterCooldownTicks | int64 | `encounter_cooldown_ticks` | 7200 | [0, 86400] |

Rules:

- `nil` TuningState ≡ the default set (every pre-048 snapshot and world).
- A non-nil TuningState always carries **all five** fields fully resolved
  (defaults filled in, clamps applied) — never a partial/sparse struct. This is
  what the event payload snapshots and what accessors read.
- `omitempty` keeps pre-048 snapshots byte-identical; **no `format_version`
  bump** (spec-044 precedent, `state.go:96-106`).

## Accessors (nil-safe, on `*State`)

`RefuelDyingBelow()`, `FireBurnPerWood()`, `GruEmergePerMille()`,
`PlannerCadence()`, `EncounterCooldown()` — return the tuned value when
`s.Tuning != nil`, else the `default*` constant. All call sites (reducer and
mind-replica side) go through these; the raw constants are renamed
`default*` and referenced only by `tuning.go` and tests.

## Manifest file (`<world>/tuning.json`, operator-authored)

Sparse by design — set only the dials you want to move:

```json
{
  "fire_burn_per_wood": 28800,
  "gru_emerge_per_mille": 300
}
```

- Absent file → no TuningState change (defaults / whatever state carries).
- Unknown keys or wrong types → boot error (DisallowUnknownFields).
- Out-of-range values → clamped, warning per field.
- Parsing resolves the sparse file against defaults into a **full** effective
  set before comparison/seeding.

## Event: `sim.tuning_applied`

| Aspect | Contract |
|---|---|
| Type | `sim.tuning_applied` |
| Payload | the full effective TuningState (all five fields, post-clamp) |
| Emitted | at boot, by the daemon seed, only when effective ≠ in-effect |
| Reducer | sets `s.Tuning` to the payload set (idempotent re-apply safe) |
| Replay | re-applies like any event; the file is never consulted |

State transitions:

```
Tuning == nil (defaults) ──sim.tuning_applied──▶ Tuning == full set S₁
Tuning == S₁            ──sim.tuning_applied──▶ Tuning == full set S₂
(no transition, no event, when effective set == in-effect set)
```

## Relationships

- `world.World` — gains `TuningPath()` (`<Dir>/tuning.json`); no manifest.json
  change.
- `daemon` boot order — recoverState → load+parse tuning → seedTuning
  (apply+append; before the loop starts and before mind construction, so the
  mind's initial snapshot JSON already carries the tuned state).
- `mind.Mind` — no new fields; reads accessors off `md.replica`, which absorbs
  `sim.tuning_applied` like every other event.
