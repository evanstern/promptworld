# Contract: tuning.json and sim.tuning_applied

## File contract — `<world-dir>/tuning.json`

Operator-authored, optional, sparse JSON object. Peer of `manifest.json`,
`calibration.json`, `llm.json`.

### Schema

```json
{
  "refuel_dying_below": 10800,
  "fire_burn_per_wood": 14400,
  "gru_emerge_per_mille": 600,
  "planner_cadence_ticks": 1800,
  "encounter_cooldown_ticks": 7200
}
```

All keys optional. All values JSON numbers (integers). Units: game-seconds
(= ticks) for all fields except `gru_emerge_per_mille` (probability per mille
per night, 0–1000).

### Bounds (clamped, warn-not-error)

| Key | Min | Default | Max | Notes |
|---|---|---|---|---|
| `refuel_dying_below` | 0 | 10800 | 86400 | 0 disables the refuel reflex trigger window (default raised 3600 → 10800 in spec 057 / TASK-108) |
| `fire_burn_per_wood` | 600 | 14400 | 86400 | effective deadline still truncated by the unpromoted `fireFuelCap` |
| `gru_emerge_per_mille` | 0 | 600 | 1000 | 0 = gru never emerges; 1000 = every night |
| `planner_cadence_ticks` | 60 | 1800 | 86400 | divides stagger/bucket math; min guards degenerate schedules |
| `encounter_cooldown_ticks` | 0 | 7200 | 86400 | 0 = pair encounters gate only on adjacency/debounce |

### Boot behavior

| Input | Outcome |
|---|---|
| File absent | Silent; effective set = defaults (or whatever state carries) |
| Valid, in-range | Values applied |
| Known key, out of range | Clamped to bound; per-field warning: field, raw value, clamped value |
| Unknown key | Boot fails: file path + offending key |
| Malformed JSON / wrong type | Boot fails: file path + parse error |
| Edited while daemon runs | No effect until next boot |

Warning/error text follows the `llm.json` style, e.g.
`tuning.json fire_burn_per_wood 999999 out of range (max 86400) — clamped to 86400`.

## Event contract — `sim.tuning_applied`

```json
{
  "tick": 123456,
  "type": "sim.tuning_applied",
  "payload": {
    "refuel_dying_below": 3600,
    "fire_burn_per_wood": 28800,
    "gru_emerge_per_mille": 300,
    "planner_cadence_ticks": 1800,
    "encounter_cooldown_ticks": 7200
  }
}
```

- Payload is the **full effective set** (all five fields, post-clamp), never a
  delta.
- Appended by the daemon boot seed only when the effective set differs from the
  in-effect set (`nil` state tuning compares as the default set). Restarting
  with an unchanged file appends nothing.
- Reducer arm: sets `State.Tuning` to the payload set. Pure, idempotent under
  re-application.
- Replay: values come exclusively from this event; `tuning.json` is never read
  during replay.
- Ordering: seeded after state recovery, before the sim loop starts and before
  mind construction (`daemon.go` recoverState → seed → `mind.New(state.Marshal())`),
  so no tick and no planner schedule ever runs ahead of the tuned values.

## Compatibility

- Pre-048 snapshots: `State.Tuning` is `omitempty` → old snapshots unmarshal to
  `nil` (defaults); re-marshaled snapshots without tuning stay byte-identical.
  No `format_version` bump.
- Pre-048 logs: contain no `sim.tuning_applied`; replay uses defaults
  throughout — behavior unchanged.
- Secondary event consumers (TUI digest, chronicle, morgue render): must not
  crash on the new event type; no rendering obligation in this spec.
