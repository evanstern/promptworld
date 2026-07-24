# Recorded-Event Contracts: Paused Authoring Chain-Completion

These are the durable, replay-audited surfaces this feature commits to. They
are contracts because they enter the append-only event log and must reproduce
byte-identically on replay (FR-006/FR-007).

## C1: Paused verdict arithmetic (in `cog.outcome.reason` and allow-path verdicts)

Format (pure, from `cognition.RoutePaused`):

```
{points}pt x {secondsPerPoint:.1f}s/pt while paused = 0 ticks <= budget {budgetTicks}
```

Example (planner, bootstrap 20 s/pt): `3pt x 20.0s/pt while paused = 0 ticks <= budget 1200`

Properties:
- Contains the literal word `paused` (FR-005: the arithmetic names the paused state).
- Never produced while `State.Paused == false`; running-world strings are
  byte-identical to today's `Route` output (SC-005).
- Deterministic given (class registry, calibrated seconds-per-point, paused
  flag) — all event-log-derived or config-pinned; reproducible on replay.

## C2: Arming causality edge

A `cog.thought` for a paused nudge-wake round records
`trigger_seq == Seq(metatron.nudged)` — the nudge landing is the recorded
cause of the thought (existing FR-020 edge, new stimulus).

## C3: Frozen-tick truth on thought records

While paused, for any routed thought:
- `cog.thought.predicted_land_tick == snapshot_tick`
- `cog.outcome.landing_tick == snapshot_tick == frozen tick`
- `cog.outcome.staleness_ticks == 0`
- prompt carries **no** future-dating prefix (`futureDated` no-ops at landing ≤ now)

## C4: No new event types or payload fields

The feature emits only existing types (`cog.thought`, `cog.outcome`, and
whatever the landed plan produces through existing doors). A replay of any log
written before this feature reduces identically after it (backward
byte-compatibility); a log containing paused sessions replays identically on
any build containing this feature.
