# Phase 1 Data Model: Live Cognition-Horizon Surface

No persistent state is added. All entities are in-memory (daemon process
lifetime) or wire-only (status reply). No event-log schema changes: the
existing `cog.outcome` suppression records are untouched.

## ClassStanding (internal/cognition)

One watched class's live standing at a given speed — the structured base the
spec-035 string surfaces filter from.

| Field | Type | Meaning |
|-------|------|---------|
| `Class` | `string` | Watched decision-class name (`planner`, `conversation`, `meeting`) |
| `Suppressed` | `bool` | `!Route(...).Allow` at the given ticks/sec and live estimate |
| `Verdict` | `cognition.Verdict` | The router's full verdict — `Arithmetic` carries the auditable string verbatim |

**Construction**: `LiveHorizon(ticksPerSecond, secPerPtFor)` — iterates
`WatchedClasses()` in doctrine order; a class where `secPerPtFor` returns
`ok=false` is EXCLUDED (no entry), matching `SuppressedAt`'s exclusion
semantics. `ticksPerSecond ≤ 0` yields every included class suppressed with
`Route`'s uncapped phrasing.

**Invariant**: `SuppressedAt(tps, f)` ≡ names of `LiveHorizon(tps, f)` entries
with `Suppressed == true` (single-iteration doctrine, spec FR-002).

## Suppression counters (internal/llm)

Process-lifetime per-class counts on the `Orchestrator`.

| Element | Type | Meaning |
|---------|------|---------|
| `suppressions` | `map[string]int64` + mutex | Count of router-suppressed thoughts per decision class since daemon start |
| `RecordSuppression(class string)` | method | O(1) increment; safe from the mind's absorb goroutine |
| `SuppressionCounts() map[string]int64` | method | Defensive copy for the status composer |

**Rules**: counts ALL classes reaching `emitSuppressed` (watched or not);
monotonic; reset only by daemon restart; no persistence. A world without an
orchestrator has no counter (FR-009 by construction).

## suppressionCounting seam (internal/mind)

Optional interface the mind type-asserts on its orchestrator handle
(`md.orch`), mirroring the existing `estimating` seam:

```go
type suppressionCounting interface {
    RecordSuppression(class string)
}
```

Called from `emitSuppressed` before the detached event emit. A fake/nil
orchestrator lacking the seam is a silent no-op — telemetry never blocks and
never fails the absorb loop.

## HorizonClass (internal/ipc, wire)

One entry of `StatusData.Horizon` — additive, `omitempty`, present only for
LLM worlds; entries follow `WatchedClasses()` order; classes with no
admissible serving provider are absent.

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| Class | `class` | `string` | Decision-class name |
| Suppressed | `suppressed` | `bool` | Router verdict at current EFFECTIVE speed and live estimate |
| Verdict | `verdict` | `string` | `Verdict.Arithmetic` verbatim (audit string) |
| Calibrated | `calibrated` | `bool` | Serving provider has a calibration-profile entry (`CalibratedAt != ""`) |
| SuppressedCount | `suppressed_count` | `int64` | This class's suppression count since daemon start |

**Composition point**: `statusDataFull` (`internal/ipc/server.go`) — resolver
`EstimateForKind(llm.Kind(class))` supplies (provider, estimate, ok);
`ok=false` excludes the class; speed from `sim.Status.Speed.TicksPerSecond()`.

**State transitions**: none — recomputed from scratch on every status request
(pure function of live estimate × effective speed × counters).

## Render-side (no new state)

- **TUI**: header badge and metatron-pane block are pure renders of
  `m.status.Horizon`; no client-side projection, no re-derivation, nothing to
  reset on reconnect.
- **CLI**: status section is a pure render of the same field.
