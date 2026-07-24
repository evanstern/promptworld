# Data Model: Teaching-World Speed Posture

## Durable state (the only addition)

### `world.Manifest.Teaching` — teaching posture marker

| Field | Type | JSON | Default | Notes |
|---|---|---|---|---|
| `Teaching` | `bool` | `teaching,omitempty` | `false` | Per-world; absent on every pre-existing manifest and on non-teaching worlds (byte-identical files, FR-008). FormatVersion unchanged (3). |

- **Written by**: `promptworld new --teaching` (creation); `promptworld teaching
  <world> on|off` (offline toggle via world-package helper). Never by the daemon.
- **Read by**: daemon boot (posture default + boot prompts), IPC server (postureWarning,
  status Posture block), CLI status rendering.
- **Validation**: none needed — any bool is legal; teaching + no orchestrator is legal
  and inert (edge case: pure-sim teaching world).

## Derived values (never stored)

### Posture rung

```
postureRung = cognition.MaxSafeSpeed("planner", est)
  where (name, est, ok) = Orchestrator.EstimateForKind(Kind("planner"))
  provenance: calibrated = Orchestrator.CalibratedAt(name) != ""
```

- Domain: `{0, 1, 4, 8, 16, 32}` — a `horizonLadder` rung, or `0` when even 1×
  suppresses the planner (edge: posture defaults to lowest rung `1` for the boot
  default; the warning arithmetic still surfaces).
- Recomputed at: daemon boot (default application), every set_speed (warning gate),
  every status (Posture block). No cache, no persistence — SC-005 by construction.

### State transitions

None. The world state machine is untouched; the boot-time posture default rides the
existing `clock.speed_set` recorded event (loop command path), so replay reproduces it
byte-identically.

## Wire additions (additive, omitempty — see contracts/posture.md)

| Type | Field | Present when |
|---|---|---|
| `ipc.StatusData` | `Posture *PostureStatus` | teaching world AND orchestrator, status-family replies |
| `ipc.PostureStatus` | `Rung string`; `Calibrated bool` | (new type) |
| `ipc.StatusData.Warning` | (existing, spec 035) | now ALSO carries posture-override text on set_speed replies for teaching worlds; composition newline-joins with the uncalibrated warning when both fire |

## Relationships

- Posture rung **depends on** the planner-serving provider's live estimate (spec 024
  per-provider estimators) → recalibration/provider-failover automatically moves it.
- `PostureStatus.Calibrated` **is the same predicate** as spec 035/037's
  `CalibratedAt != ""` — one provenance rule everywhere.
- TASK-68 stage presets **consume** `Manifest.Teaching` (config) + `StatusData.Posture`
  (runtime); nothing here depends on TASK-68.
