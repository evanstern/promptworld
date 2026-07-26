---
name: cognition-governor-debt
description: Split from [[cognition]] — the adaptive-throttle governor (spec 028) that turns the player's speed setting into a ceiling, not a promise; the pure Debt staleness signal (spec 033 accrued-debt revision) and the Governor hysteresis state machine that sheds/recovers one speed notch at a time
kind: component
sources:
  - internal/cognition/governor.go
verified_against: 30912a9cd5d2334f76425ac8ca5b74a7a7c90876
---

# Cognition: adaptive-throttle governor

Split from [[cognition]] (the cognition horizon substrate): the governor
extends the horizon from the other side — instead of only scoping what a
model may decide at a given speed, the world governs its own effective speed
when the player wants both high speed and high thought fidelity.

## How it works

**Adaptive-throttle governor** (`governor.go`, spec 028, doctrine research R6):
the player's speed setting becomes a CEILING, not a promise: `Debt` and
`Governor` are pure, stdlib-only functions of a snapshot the daemon supplies
every sample; nothing here calls a model, reads a wall clock, or is
config-tunable at runtime.

- **Debt** (`Debt(pending []PendingDebtInput, ticksPerSecond) (debt float64,
  jobs int)`): the aggregate staleness signal (spec 033, revising spec 028
  FR-001/FR-002) — for each pending thought the seconds are piecewise:
  `PredictedSec − ElapsedSec` while within prediction (remaining work, drains
  as today), `ElapsedSec` once overrun (full accrued drift, grows) — then
  `× ticksPerSecond / BudgetTicks(class)`. An overdue thought's elapsed time
  IS its grounded debt; the pre-033 floor-to-zero inverted the signal under
  overload (worst drift → least debt → no shed, the world-01 defect). The
  boundary jump at `ElapsedSec == PredictedSec` is doctrine
  (specs/033-governor-accrued-debt/contracts/debt-formula.md). `debt` is the
  sum, `jobs` counts only positive-contributing entries (overdue thoughts now
  contribute, so they count). Unknown kinds are skipped (they cannot reach a
  model) and `ticksPerSecond ≤ 0` (uncapped max) yields zero — pure
  arithmetic, no randomness, identical inputs always yield identical debt.
- **`Governor`** (a hysteresis state machine, one instance owned by the
  daemon's sampler, [[daemon-lifecycle]]): `Sample(debt, jobs, paused,
  effective, requested) Decision` counts consecutive qualifying SAMPLES, not
  wall durations, at the daemon's fixed `GovernorCadence`. Breach accrues
  while `debt > ShedThreshold` and the effective speed sits above the 1x
  floor (`clock.LadderIndex`); a continuous `BreachWindow` sheds exactly one
  notch (`clock.CappedLadder()[idx-1]`). Recovery accrues only while governed
  with room to climb — the debt PROJECTED at the next notch up (current debt
  scaled by that notch's tick-rate ratio, FR-006) stays under `ShedThreshold ×
  RecoverHeadroom`; a continuous `RecoveryWindow` (deliberately longer than
  `BreachWindow` — asymmetric hysteresis, US3) recovers one notch, never
  above the requested ceiling. Any decision, pause, or a speed change between
  samples resets both windows; a paused sample is a no-op (FR-013 — elapsed
  pause time never counts). At the 1x floor with debt still over threshold,
  the governor saturates silently — no decision, visible only via status.
- **Doctrine constants** (versioned with the code, never runtime knobs,
  FR-007 — same posture as the registry's points/budgets): `GovernorCadence`
  1 s (the daemon's sampling interval), `ShedThreshold` 1.0 (budget-fractions
  above which breach accrues), `BreachWindow` 5 s, `RecoverHeadroom` 0.5
  (scales `ShedThreshold` to the recovery ceiling), `RecoveryWindow` 20 s.
- **Who owns/calls it**: the daemon builds a `governorSampler` only when an
  orchestrator exists (a no-LLM world constructs zero governor machinery,
  FR-003/SC-004) and runs it in its own goroutine, sampling
  `llm.Orchestrator.PendingCognition()` ([[llm-orchestrator]]) and the
  [[sim-loop]]'s non-blocking status door every cadence, storing the debt
  reading for status, and issuing any resulting shed/recover `Decision`
  through the loop's `Govern` door — which lands it as a recorded
  `clock.governor_shed`/`clock.governor_recovered` event or drops it silently
  if it no longer applies ([[event-types]], [[sim-state-reducer]]). The
  package itself owns no goroutine, no wall clock, and no loop reference —
  only the pure `Debt` function and the `Governor` decision struct.

## Connections

The daemon's governor sampler ([[daemon-lifecycle]]) drives `Debt`/`Governor`
from [[llm-orchestrator]]'s `PendingCognition` snapshot and the
[[sim-loop]]'s status/`Govern` doors; the router ([[cognition]]'s `Route`,
consumed at the [[sim-loop]]'s landing ladder) reads the EFFECTIVE speed the
governor may have shed, so shedding speed deterministically widens what the
model may own and recovery narrows it again (spec 028 FR-010, extending
[[cognition]]'s decision-4 from the other side).

Back to [[cognition]] for the decision-class registry, the pure routing
arithmetic, and the delivery-gate doctrine the governor's shed/recovered
speed feeds into.
