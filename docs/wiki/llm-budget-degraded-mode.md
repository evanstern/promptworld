---
name: llm-budget-degraded-mode
description: Split from [[llm-orchestrator]] (spec 024 US4 / TASK-32) — the single global spend ceiling with per-provider attribution, the per-provider circuit breaker (degraded mode), seconds-per-point latency estimation/calibration provenance, and the router suppression counters. Load for budget, breaker, or calibration questions.
kind: component
sources:
  - internal/llm/meter.go
  - internal/llm/health.go
  - internal/llm/llm.go
verified_against: 04ff15001bd8a74f7c2965889c0d318fc0dc03a9
---

# LLM orchestrator — budget, degraded mode & estimation

Split from [[llm-orchestrator]] (spec 024 US4, TASK-32): the one-wallet spend
ceiling, the per-provider circuit breaker, live latency estimation and its
calibration provenance, and router suppression counters.

**Spend: one wallet, per-provider attribution** (`meter.go`, spec 024 US4): a single
global `monthly_budget_usd` ceiling, checked at admission per priced candidate BEFORE
any HTTP. Cost uses the serving provider's declared pricing; `Add(provider, cost)`
writes the unchanged total key `llm_spend_YYYY-MM` AND `llm_spend_YYYY-MM:<provider>`
under one lock — restarts never forget money, per-provider rows sum to the total, and
pre-024 months surface their remainder as unattributed. Zero-priced providers are
never budget-refused (pricing class, not tier identity — the one deliberate
behavioral edge vs pre-024: a hypothetical zero-priced cloud router now serves past
the ceiling). The ceiling is per-WORLD by architecture: the meter persists in
the world's own meta table and the budget number comes from the world's own
`llm.json` — "one wallet / single global ceiling" means one wallet across
PROVIDERS within a world, never machine-global state ([[design-grounding]]'s
"never global; runs cleanly separable"). Spec 076 ([[world-forking]]) leans
on exactly that: a fork **inherits the parent's wallet as of fork time** —
`llm.json` copies verbatim (same ceiling) and every `llm_spend_*` meta key
(totals + per-provider attribution) copies into the fork's fresh store, so
the fork's meter opens at the parent's month/spend/ceiling and forking never
mints fresh budget. Thereafter each world meters independently — recorded
limitation: a duel's combined forward spend can exceed one ceiling by up to
the unspent remainder at fork time; attribution stays per-world and
per-provider, unchanged.

**Degraded mode** (`health.go`, per-provider): a circuit breaker — 3 consecutive
failures open it (15 s backoff doubling to 5 min), an open circuit refuses instantly
(and is skipped by the chain-walk), one half-open probe tests recovery. Busy is not
down (TASK-22): the worker skips queued jobs whose caller already gave up and never
counts a failure when the caller's own ctx died mid-call — only genuine provider
failures and the worker cap strike the breaker. A killed model degrades the AI layer;
the daemon and loop never notice.

**Latency estimation** (TASK-32, [[cognition]], per-provider): each provider carries
a live `cognition.Estimator` of seconds-per-point — the worker samples each
*successful* call's wall time normalized by the kind's point cost; with `parallel` >
1 samples include server-side contention, converging on true concurrent-rate cost.
Estimators seed from `cognition.SeedFor(profile, name, zeroPriced)` — profile keyed
by provider name (legacy worlds' derived `local`/`cloud` keep matching), miss falls
back by pricing class; `SeedCalibration` re-seeds all providers at daemon start and —
since spec 035 (R3) — also records each provider's seed PROVENANCE: `calibratedAt`
is set to the profile's `CalibratedAt` when `cognition.Calibrated(p, name)` finds a
usable entry, else cleared to `""` (bootstrap-seeded) — the exact presence test
`SeedFor` applies, so the seed value and its provenance can never disagree.
After `SeedCalibration`, `SeedPersisted(state)` raises each provider's seed to
`cognition.ReseedValue` = max(current seed, the provider's persisted live
estimate from `estimator_state.json`) — TASK-113's restart persistence; it
only ever raises, so a fresher human calibration or bootstrap floor is never
undercut, and `SnapshotEstimators()` exports the live estimates for the
daemon's periodic/shutdown flush ([[daemon-lifecycle]]).
`Orchestrator.CalibratedAt(name)` reads it back (`""` for an unknown provider or one
never seeded); it is set exactly once per `SeedCalibration` call and never mutated
by live estimator adaptation (spec 031) or drift — the seed-state fact and the live
estimate are deliberately independent (research R2). This is the read
[[ipc-server]]'s `set_speed` uncalibrated-warning gate and [[cli-promptworld]]'s
`providerCalibrationLine` both use. The
mind-facing exports (spec 024): `EstimateForKind(kind)` returns the kind's CURRENT
ADMISSIBLE chain head's name + estimate (`admissibleHead`, a non-mutating read —
falls back to the chain head when none admissible), `ResolveProvider(kind)` is the
pin-resolution dry walk, `ProviderNames()`/`ProviderConfig(name)` serve calibrate,
and `Kinds()` still feeds the cognition registry's completeness gate at daemon start.
Per-provider recalibrate hooks fire once per breach episode via `SetRecalibrateHook`;
since spec 031 a breach also ADOPTS (the estimator re-seeds to its window median), and
`feedEstimate` forwards the evidence — the hook signature carries `prior` and `adopted`
alongside the post-adoption estimate. The mind records `cog.recalibration_recommended`
(the provider name rides the recorded payload's `Tier` field, kept for replay-schema
stability; the adoption arithmetic rides additive omitempty fields).

**Suppression counters** (`llm.go`, spec 037 FR-004): `RecordSuppression(class)`
bumps a `suppMu`-guarded `map[string]int64` — one O(1) increment per router
suppression, so it never blocks the mind's absorb goroutine calling it.
`SuppressionCounts()` returns a defensive copy for the status composer to
range/mutate freely. Counters are daemon-lifetime (reset only by restart) and
count EVERY class the mind reports, watched or not; [[ipc-server]]'s
`horizonClasses` reads them back but keys out only the watched ones onto the
wire. [[agent-mind]]'s `emitSuppressed` is the sole caller, reporting through
the optional `suppressionCounting` seam (mirroring `estimating` below) rather
than a hard dependency, so a test fake or nil orchestrator is a silent no-op.

## Connections

Part of [[llm-orchestrator]]'s summary-style split (corpus-spec v2). See
[[llm-orchestrator]] for the overall package doctrine and its other split-off
domains: [[llm-provider-registry]] (declared providers/routes, the embedding
route, transports, config validation), [[llm-chain-walk-dispatch]] (chain-walk
admission, the tool-use loop wire shape, per-kind knobs), and
[[llm-concurrency-leases]] (worker concurrency, priority lanes, endpoint
leases, the pending-thought registry). [[llm-provider-health]] builds on this
package's `provider`/worker/breaker machinery (condition slot beside
`tierHealth`, the worker's success path, `SetConditionHook`) to make a dead or
tool-silent provider operator-visible. Since spec 035, [[ipc-server]]'s
`set_speed` warning reads `EstimateForKind` and `CalibratedAt` together, and
[[cognition]]'s `Calibrated` is what `SeedCalibration` reads to set
`calibratedAt` in the first place. Since spec 037, [[agent-mind]]'s
`emitSuppressed` feeds `RecordSuppression`, and [[ipc-server]]'s
`horizonClasses` reads `SuppressionCounts` alongside `EstimateForKind`/
`CalibratedAt` to compose the status wire's per-class horizon.
