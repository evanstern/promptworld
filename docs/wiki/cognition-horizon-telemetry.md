---
name: cognition-horizon-telemetry
description: Split from [[cognition]] — the operator-facing horizon summaries (HorizonSummary, MaxSafeSpeed, LiveHorizon, SuppressedAt over the fixed speed ladder) every status surface reads, plus the two per-call telemetry payload shapes (CogToolCallPayload, MemoryDivergencePayload) internal/sim/cognition.go records
kind: component
sources:
  - internal/cognition/horizon.go
  - internal/sim/cognition.go
verified_against: 30912a9cd5d2334f76425ac8ca5b74a7a7c90876
---

# Cognition: horizon summaries and telemetry

Split from [[cognition]] (the cognition horizon substrate): the read-only
surfaces that summarize `Route` across the fixed speed ladder for operators,
and the two per-call payload shapes that record cognition's tool-use and
memory-selection behavior for later inspection.

## Horizon summaries

**Horizon summaries** (`horizon.go`, spec 035 R1): `WatchedClasses()` returns
a copy of the fixed `planner`/`conversation`/`meeting` set the operator-facing
surfaces summarize or gate on. `HorizonSummary(secPerPt)` evaluates `Route`
for each watched class across the fixed speed ladder (1/4/8/16/32) at one
seconds-per-point and joins the per-class verdict ("planner suppressed above
16x", "conversation OK at 32x", or "… always suppressed"); its per-class
max-safe-rung loop is extracted (spec 039 FR-002) as
`MaxSafeSpeed(class, secPerPt)` — the highest ladder speed at which the class
still routes to the model, 0 when even 1x suppresses (callers clamp to the
ladder floor), the teaching-posture rung that decides through `Route` so a
posture can never disagree with the router. The summary body was moved verbatim
from `cmd/promptworld/calibrate.go` so the daemon's boot warning, the
`set_speed` warning, and `promptworld calibrate` all read the ONE
implementation (FR-006): the printed horizon can never disagree with the
router. `MaxSafeSpeed(class, secPerPt)` (spec 039 FR-002) is `HorizonSummary`'s
per-class maxOK loop pulled out into its own export: the highest
`horizonLadder` rung at which the named class still routes to the model under
`secPerPt` (0 when even 1x suppresses it, or the class is unregistered) —
`HorizonSummary` now calls it rather than duplicating the loop. It decides
through the same `Route` call as everything else here, so a posture derived
from it can never disagree with the router either; the teaching-posture
default ([[daemon-lifecycle]]) and its `set_speed`/status overrides
([[ipc-server]]) are its callers, clamping its 0 result to the ladder floor via
[[game-clock]]'s `SpeedForRate`.
`SuppressedAt(ticksPerSecond, secPerPtFor)` is the live-estimate twin
— same watched classes and `Route` arithmetic, but at one live
ticks-per-second against a per-class resolver that can exclude a class
entirely (`ok=false`) when its serving provider is calibrated, so a
calibrated class never contributes to the warning; this backs the
`set_speed` uncalibrated-warning composition ([[ipc-server]]).

Spec 037 (`contracts/status-horizon.md`) generalizes the live-estimate twin
into a structured base: `ClassStanding{Class, Suppressed, Verdict}` and
`LiveHorizon(ticksPerSecond, secPerPtFor) []ClassStanding` evaluate the same
watched classes and `Route` arithmetic once, returning each INCLUDED class's
full standing (the verdict's arithmetic string and all) in `WatchedClasses`
order — a class whose resolver returns `ok=false` is excluded entirely (no
entry, not merely never-suppressed), and `ticksPerSecond <= 0` (uncapped max
speed) suppresses every included class via `Route`'s uncapped phrasing.
`SuppressedAt` is now re-based as a suppressed-names filter over
`LiveHorizon` — one watched-class iteration total, so every operator-facing
horizon surface, from the plain warning to the richer per-class status, can
never disagree with the router (spec 035 FR-006 posture, extended).
[[ipc-server]]'s `horizonClasses` composes `LiveHorizon` at the loop's
effective speed into the status wire's structured horizon, rendered per
class by [[cli-promptworld]] and [[tui-client]].

## Tool-call telemetry

**Tool-call telemetry** (`internal/sim/cognition.go`, spec 017 FR-007):
`CogToolCallPayload` (`Job`, `Ordinal`, `Tool`, `Args` capped to 2 KiB,
`Verdict` — the stringified `toolloop.Verdict` enum — `Reason`, `Tier`,
`SnapshotTick`) is one record per tool call a cognition's loop saw: landed,
landed_clamped (spec 058 FR-001/FR-003, `toolloop.VerdictLandedClamped` —
[[tool-loop]]), rejected, read, or unlanded. `{Job, Ordinal}` is the correlation key
(1-based, dense per job, model-emission order). It rides the same reducer-no-op
`cog.*` doctrine as every other cognition event ([[event-types]]).
`NewCogToolCallPayload` assembles the payload sim-side — deliberately with
only plain/stdlib argument types (no `toolloop` or `mind` import) — so both
loop consumers ([[agent-mind]]'s mind, [[guardian]]) unpack their own
`toolloop.CallRecord` and call this one shared constructor rather than each
inventing its own payload shape.

## Memory-relevance divergence telemetry

**Memory-relevance divergence telemetry** (`internal/sim/cognition.go`, spec
042 US2): `MemoryDivergencePayload` (`Agent`, `Tick`, `Mode`, `Legacy`/
`Augmented` — both memory `Seq` lists in window order, `Overlap`,
`Displacement`, `Vectorless`, `SitTick`) is one record per selection the mind
makes while a world's `memory_relevance` flag is `"shadow"` or `"on"`,
recording how the relevance-augmented window (`sim.SelectMemoriesRelevant`)
ranks against the legacy window (`sim.SelectMemories`) it is shadowing —
the evidence the shadow-mode US2→US3 promotion decision is made from. It
rides the same reducer-no-op `cog.*` doctrine as `CogToolCallPayload` above
([[event-types]]). `NewMemoryDivergencePayload` computes `Overlap` (memories
present, by `Seq`, in both windows) and `Displacement` (the summed absolute
rank distance of each shared member) purely over the two selected `[]Memory`
slices — a pre-042, seq-less memory (`Seq` 0) never counts as shared, since
it carries no durable identity. [[memory-retrieval]] owns the selector this
payload audits and the shadow/on posture that gates its emission.

## Connections

Since spec 035, [[daemon-lifecycle]]'s boot warning and [[ipc-server]]'s
`set_speed` warning read `HorizonSummary`/`SuppressedAt`; since spec 039,
[[daemon-lifecycle]]'s boot-time teaching-posture default and [[ipc-server]]'s
`postureWarning`/`postureStatus` read `MaxSafeSpeed` the same way. Since spec
037, [[ipc-server]]'s `horizonClasses` also reads `LiveHorizon` directly (not
just the `SuppressedAt` filter) to compose the status wire's structured
per-class horizon, and [[llm-orchestrator]]'s daemon-lifetime
`SuppressionCounts` (fed by [[agent-mind]]'s `emitSuppressed` through the
`suppressionCounting` seam) rides alongside each entry as its
`SuppressedCount` — a fact this package itself never tracks. Since spec 042,
`cog.memory_divergence` records one shadow/on-mode selection's rank
divergence per emission, purely observational — it gates nothing itself
([[memory-retrieval]]).

Back to [[cognition]] for the decision-class registry, the pure routing
arithmetic, and the delivery-gate doctrine these summaries and payloads
observe.
