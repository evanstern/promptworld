# Contract: status-reply horizon block

**Surface**: IPC `status` / `pause` / `resume` / `set_speed` replies
(`StatusData`, `internal/ipc/protocol.go`) — the shared response shape gains
one additive field.

## Field

```json
{
  "world":  { ... },
  "clock":  { ... },
  "daemon": { ... },
  "log":    { ... },
  "llm":    { ... },
  "horizon": [
    {
      "class": "planner",
      "suppressed": false,
      "verdict": "3pt x 17.0s/pt x 8x = 408 ticks <= budget 1200",
      "calibrated": true,
      "suppressed_count": 0
    },
    {
      "class": "conversation",
      "suppressed": true,
      "verdict": "13pt x 17.0s/pt x 8x = 1768 ticks > budget 7200",
      "calibrated": false,
      "suppressed_count": 214
    }
  ]
}
```

## Rules

1. **Presence**: `horizon` appears ONLY when the world has an orchestrator
   (`llm.json`). A no-LLM world's reply is byte-identical to pre-037 output
   (`omitempty`; the composer never runs). Empty-slice is never emitted —
   either absent or ≥1 entry.
2. **Membership**: entries follow `cognition.WatchedClasses()` order
   (`planner`, `conversation`, `meeting`). A class whose kind has no
   admissible serving provider (`EstimateForKind` ok=false) is omitted.
   Calibrated classes are INCLUDED (contrast: the spec-035 `warning` field,
   which stays gated to uncalibrated providers and is unchanged by this
   feature).
3. **Verdict source**: `suppressed` and `verdict` come from
   `cognition.LiveHorizon` → `cognition.Route` at the loop's EFFECTIVE speed
   (`sim.Status.Speed.TicksPerSecond()`) and the serving provider's live
   estimate. `verdict` is `Verdict.Arithmetic` VERBATIM — clients must not
   parse it, only display it. At uncapped speed (`tps ≤ 0`) every entry is
   suppressed with `Route`'s uncapped phrasing.
4. **Counts**: `suppressed_count` is the daemon-lifetime count of router
   suppressions for that class (source: `Orchestrator.SuppressionCounts()`,
   fed by the mind's `emitSuppressed`). Monotonic while the daemon runs;
   resets only on daemon restart; zero for a class never suppressed.
5. **Refresh**: recomputed per status request; clients read it on their
   existing poll cadence. No push channel.
6. **Stability**: additive field on `StatusData`; every existing field's bytes
   are unchanged. Consumers must tolerate the field's absence (older daemon)
   and unknown future fields per entry.

## Consumer obligations

- **TUI**: header badge iff ≥1 entry has `suppressed == true`; metatron-pane
  block renders one row per entry; remedy phrase from `suppressed` ×
  `calibrated` (uncalibrated → mention calibration; calibrated → slow down
  only). Raw wire strings other than `verdict` never reach the screen
  unphrased.
- **CLI status**: renders the section for LLM worlds; no output change for
  no-LLM worlds.
