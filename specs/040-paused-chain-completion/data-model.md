# Phase 1 Data Model: Paused Authoring Chain-Completion

No new entities, event types, or payload fields. The feature consumes and
produces existing shapes; this document pins which fields matter and the one
state transition the feature adds.

## Consumed

### `metatron.nudged` (event, existing — internal/sim/metatron.go:37-43)

| Field | Type | Use here |
|-------|------|----------|
| `form` | string (`"dream"`/`"vision"`/`"omen"`) | not consulted — all forms wake |
| `targets` | `[]int` (agent indices) | the villagers to arm, one round each |
| `text` | string | not consulted (the dream memory carries it to the prompt) |
| event `Seq` | int64 | recorded as the arming stimulus (`pendingSeq`) — the causality edge on the eventual `cog.thought` |

### `State.Paused` (replica field, existing — internal/sim/state.go:25)

Reduced deterministically from `clock.paused` / `clock.resumed`
(state.go:337-340). The single source of paused truth for both fixes; never
wall-clock.

### `DecisionClass` (registry, existing — internal/cognition/registry.go)

`Points`, `BudgetTicks` feed the paused verdict's recorded fields unchanged.
Budgets are doctrine (decision-6): no values change.

## Produced

### `cognition.Verdict` (in-memory, existing shape — internal/cognition/route.go:7-15)

The paused variant produced by the new pure `RoutePaused`:

| Field | Paused value |
|-------|--------------|
| `Allow` | `true` (always — 0 ≤ every budget) |
| `Class`, `Points`, `BudgetTicks` | from the registry, as today |
| `PredictedWallMs` | still predicted (`Points × secondsPerPoint × 1000`) — wall time passes while frozen |
| `PredictedDriftTicks` | `0` |
| `Arithmetic` | names the paused state — exact format in [contracts/recorded-events.md](contracts/recorded-events.md) |

### `cog.thought` / `cog.outcome` (events, existing shapes — internal/sim/cognition.go)

No field changes. While paused: `PredictedLandTick == SnapshotTick` (truth fix
D3), `TriggerSeq` = the nudge event's Seq, landing/outcome tick = frozen tick,
`StalenessTicks` = 0.

## State transitions (mind-internal, existing fields)

```
pending[i]=false ──(metatron.nudged targeting i, replica.Paused)──▶ pending[i]=true, pendingSeq[i]=nudge.Seq
pending[i]=true  ──plan(): debounce open, awake, alive, no meeting──▶ job queued; lastPlanned[i]=frozenTick
pending[i]=true  ──plan(): debounce closed (< 300 game ticks)──▶ stays pending (no thought while frozen; designed bound)
```

While running (`replica.Paused == false`) the `metatron.nudged` case does not
fire: the transition graph is byte-identical to today's.
