---
name: cognition-estimator-calibration
description: Split from [[cognition]] — per-provider live seconds-per-point estimation (EWMA, spike rejection, breach-adoption, persisted estimator state surviving restarts) and the human-authored calibration profile (calibration.json, provider-keyed SeedFor/Calibrated, bootstrap defaults) that seeds it
kind: component
sources:
  - internal/cognition/estimate.go
  - internal/cognition/estimator_state.go
  - internal/cognition/calibration.go
verified_against: 30912a9cd5d2334f76425ac8ca5b74a7a7c90876
---

# Cognition: estimator and calibration

Split from [[cognition]] (the cognition horizon substrate): how a provider's
live seconds-per-point is estimated at runtime, and the human-authored
baseline that seeds it before any live samples exist.

## Estimation

**Estimation** (`estimate.go`): `Estimator` holds one provider's live
seconds-per-point as an EWMA (`EWMAAlpha` 0.2) over per-point-normalized call
durations. A sample beyond `SpikeFactor` (3.0) times the current estimate is
excluded from the EWMA but retained — with its value — in a `WindowSize`-20
ring of `{secPerPoint, spike}` slots; on the sample that first drives the
rolling spike rate over a full window past `BreachRate` (0.2, lowered from
0.3 in TASK-113 — adoption now needs ≥5 spikes per 20-window instead of ≥7,
faster median adoption with the one-shot-rejection property intact), the estimator
ADOPTS (spec 031, breach-adoption): it re-seeds `estimate` to the window
median (all retained values, spike and non-spike alike), zeroes the ring, and
`Sample` returns a non-nil `Adoption{Prior, Adopted, SpikeRate}` — the
`cog.recalibration_recommended` episode, which now has an actor instead of
being an unread signal. Re-arm is structural: a fresh window must refill
before any further verdict, and post-adoption samples in the new regime are no
longer spikes against the adopted estimate. One-shot lag spikes (too few to
breach) are thus still rejected while systemic drift — including a step change
larger than `SpikeFactor`, which pre-031 froze the estimate forever — is
followed within one window. Learned estimates survive restarts (TASK-113):
`EstimatorState` persists every provider's live seconds-per-point to
`estimator_state.json` in the world dir (`estimator_state.go` — daemon-written
every 5 minutes and at shutdown, unlike the human-only `calibration.json`),
and boot reseeds each provider to `ReseedValue` = max(calibration/bootstrap
seed, persisted estimate) — the persisted value only ever RAISES a seed, never
undercuts a fresher human calibration or the pessimistic bootstrap floor.
Pre-113, the estimator was process-lifetime: world-01's 36 restarts reset it
to the optimistic calibration floor 36 times, re-triggering the staleness
storm each boot. The file never enters the event log and is never read during
replay — the estimator sits outside the reducer entirely.

## Calibration

**Calibration** (`calibration.go`): `Profile` is `calibration.json` in the
world save directory (`World.CalibrationPath()`), written only by the
`promptworld calibrate` subcommand (full-file replace via `Save`) — the daemon
never writes it, so the recorded baseline moves only under a human's hand.
`LoadProfile` treats a missing file as legal (nil, nil) and a malformed one as
a warning the daemon downgrades to bootstrap defaults.
`SeedFor(p, name, zeroPriced)` returns a provider's recorded seconds-per-point
— the profile is keyed by PROVIDER NAME since spec 024; legacy worlds derive
providers named `local`/`cloud`, so pre-024 tier-keyed profiles keep matching
with no translation — or, on a miss, a bootstrap by pricing class: zero-priced
providers seed `BootstrapLocalSecPerPt` (20.0), priced ones
`BootstrapCloudSecPerPt` (10.0) — an
uncalibrated world fails toward reflex, never toward stale action.
`TierProfile.SecondsPerPoint`'s unit is doctrine, spelled out since spec 017:
for a single-shot kind it is one model call's wall time per point; for a loop
cognition (the villager planner) it is the WHOLE tool-use
loop's wall time per point — the same unit [[llm-orchestrator]]'s live
estimator observes via `Orchestrator.ObserveCognition` ([[tool-loop]]), so a
seeded baseline and a live observation stay directly comparable and the
router's suppression arithmetic stays truthful when a cognition spends N
model calls, not one.
`profileEntry(p, name)` (spec 035 R3) is the single presence test ("a usable,
positive-`SecondsPerPoint` entry exists for this provider") that both
`SeedFor`'s seed-value choice and the exported `Calibrated(p, name) bool`
apply, so "what seed?" and "is this provider calibrated?" can never disagree.
`Calibrated` is the provenance a caller outside this package records —
`llm.Orchestrator.SeedCalibration` stamps each provider's `calibratedAt` from
it ([[llm-orchestrator]]), and the CLI/status surfaces render "uncalibrated
(bootstrap)" for a provider where it's false ([[cli-promptworld]],
[[llm-provider-health]]).

## Connections

[[llm-orchestrator]] owns one `Estimator` per provider (spec 024), feeds it
each completed call's duration normalized by the kind's point cost (successes
only), and exposes the live estimate back to the mind via `EstimateForKind` —
the kind's currently admissible chain-head provider's estimate, so a fast
small model is never averaged with a slow quality model. Since spec 017 the
planner's per-round calls inside [[tool-loop]] each opt out of that per-call
feed (`Request.SkipObserve`) and the loop itself reports exactly one
whole-cognition observation via `Orchestrator.ObserveCognition` when it
finishes — and only on a completed termination (landed / model_done /
cap_exhausted); the failure family (admission_refused / provider_error /
ctx_done) feeds the estimator nothing, mirroring the single-shot worker's own
successes-only doctrine so a governor observation is always a completed
cognition's true cost, never a fragment of one or a fast failure.
[[daemon-lifecycle]] runs `ValidateKinds` before any model is reachable and
seeds the estimators from the profile; [[cli-promptworld]]'s `calibrate`
subcommand benchmarks the host+model and writes the profile, delegating its
own horizon printout to `HorizonSummary` ([[cognition-horizon-telemetry]]).
`llm.Orchestrator.SeedCalibration` reads `Calibrated` to stamp each provider's
`calibratedAt`.

No environment variables; the only file this package reads is
`calibration.json` in the world directory, and only `promptworld calibrate`
writes it. With no profile (or an unreadable one), bootstrap defaults (local
20 s/pt, cloud 10 s/pt) apply and — since spec 035 — the daemon prints a full
boot WARNING block (uncalibrated statement, `HorizonSummary` at the bootstrap
seeds, and the exact `promptworld calibrate <world>` command), not just a
one-line reminder ([[daemon-lifecycle]]). Estimator drift surfaces as
`cog.recalibration_recommended` (fires once per breach episode, and since spec
031 the same episode adopts — the payload's additive `prior_s_per_pt` →
`adopted_s_per_pt` fields carry the re-seed arithmetic, `estimate_s_per_pt`
remaining "current estimate at emission", i.e. the adopted value);
`Estimator.Stats` exposes estimate, rolling spike rate, and lifetime
sample/spike counts.

Back to [[cognition]] for the decision-class registry, the pure routing
arithmetic, and the delivery-gate doctrine that consumes these seconds-per-point
values.
