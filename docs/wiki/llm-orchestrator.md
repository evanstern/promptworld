---
name: llm-orchestrator
description: Provider-registry call layer for all model traffic (spec 024) — package overview, StatusSnapshot, fresh-world defaults, and cross-package connections; declaration/config/transports, chain-walk dispatch, concurrency/leases, and budget/degraded-mode/estimation each split into a linked child note for file-level detail.
kind: component
sources:
  - internal/llm/llm.go
  - internal/llm/config.go
  - internal/llm/meter.go
  - internal/llm/health.go
  - internal/llm/providers.go
  - internal/llm/lease.go
  - internal/llm/pending.go
verified_against: d9d56cb030c15db8679e941a1ce1e4fb2a181009
---

# LLM orchestrator

`internal/llm` (TASK-6; generalized to a provider registry by spec 024 / TASK-35,
doctrine decision-5) is the single gateway for all model traffic. It lives entirely
**outside** the deterministic sim loop: LLM results reach the world only as recorded
inputs (TASK-7's job), so replay never re-calls a model — the determinism contract of
the substrate is structurally untouchable by inference.

## How it works

This note's file-level detail splits into four children by domain
(corpus-spec v2 summary-style split); each links back here.

**Providers, routes, embedding & transports** ([[llm-provider-registry]]):
model sources are declared as a named registry (`Config.Providers`) with
per-`Kind` ordered route chains (`Config.Routes`, spec 024); pricing class
(`provider.priced()`) replaces the retired "tier" concept, and the legacy
`local`/`cloud` two-provider config still loads byte-identical via
`deriveLegacy`. The embedding kind (`KindEmbedding`, spec 042) is a HEAD-ONLY
route outside `acceptedKinds`, dispatched directly by `Embed`/`WarmEmbedding`
with no queue/breaker/estimator. `openai_compat` and `anthropic` are the two
transports; `resolveRegistry`/`validateV2` (`config.go`) are the boot-time
validation authority, with a narrow backfill exception for kinds introduced
after the v2 format shipped.

**Chain-walk dispatch & the tool-loop wire** ([[llm-chain-walk-dispatch]]):
`Submit` walks a kind's chain in order, dispatching to the first admissible
candidate and recording every skip (`wallet-exhausted`/`circuit-open`/`busy`/
`queue-full`) on `Response.Skipped`; once accepted, a job's failure is final.
`Request.Tools`/`Turns`/`SkipObserve` and `Response.ToolCalls`/`Stop`
(TASK-52, spec 017) are the agent tool-use loop's wire shape; `tool_mode` is
per-provider, and `LoopMaxRounds`/per-kind `MaxTokens` budgets are top-level
kind-scoped knobs (spec 024 R9 / spec 025). `ObserveCognition` reports one
whole-cognition wall time per completed run.

**Concurrency, priority lanes & leases** ([[llm-concurrency-leases]]): each
provider owns `slots` worker goroutines (TASK-45) with slot-aware admission;
conversations ride a priority queue while `Request.BestEffort` calls yield
when busy. Advisory cross-process endpoint leases (`lease.go`, spec 024 US5)
cap combined in-flight calls across worlds sharing an endpoint via
flock-based slot files. The pending-thought registry (`pending.go`, spec
028) inventories every accepted-but-unfinished job for the daemon governor's
staleness-debt sampler.

**Budget, degraded mode & estimation** ([[llm-budget-degraded-mode]]): one
global `monthly_budget_usd` ceiling is checked at admission with
per-provider spend attribution (`meter.go`, spec 024 US4); a per-provider
circuit breaker (`health.go`) opens after 3 consecutive failures,
distinguishing genuine failures from caller-abandoned/busy. Each provider
carries a live seconds-per-point `cognition.Estimator`, seeded at boot and
reseeded from persisted state, with `calibratedAt` provenance tracked
independently of live drift. Router suppression counters (spec 037) tally
every suppression class the mind reports.

**Status** (`StatusSnapshot`, spec 024 US6): `Status{Providers []ProviderStatus,
Month, Spent, Budget}`, sorted by name — one shape for legacy and v2 worlds (legacy
shows rows `local`/`cloud`). `ProviderStatus{Name, Model, Endpoint, Up, Queue,
Inflight, Slots, Contended, SpentUSD}`, plus three `omitempty` operator-health
fields (`Condition`, `ConditionDetail`, `ConditionRemedy`, spec 034) a healthy
provider never populates — see [[llm-provider-health]] for the condition slot,
the preflight probe, and the tool-silence detector that feed them — and, since
spec 035 (FR-004/FR-008), an additive `omitempty` `CalibratedAt` string
(`json:"calibrated_at,omitempty"`): the provider's `calibratedAt` seed-provenance
field verbatim, empty for a bootstrap-seeded provider. [[cli-promptworld]]'s
`status` rendering turns this into one `providerCalibrationLine` per provider.

**Fresh-world defaults** (`DefaultConfig`, spec 034 R6): the local provider a
brand-new `llm.json` ships with is `{model: "cogito:3b", tool_mode: "json",
parallel: 4}` — the configuration the TASK-73 eval record proved live (three
8-game-hour soaks, 789–982 planner decisions each) — rather than the earlier
`gemma4:12b-mlx`/native default, a machine-local MLX build that never
reliably function-called out of the box and isn't a stock registry pull;
gemma-class models remain the documented upgrade path for operators who serve
them (docs/llm-providers.md). Existing worlds' `llm.json` files are untouched
by construction (config is per-world, read once at boot). `cmdNew`
([[cli-promptworld]]) prints the expected model and its pull command read
straight from `DefaultConfig()`.

## Connections

[[daemon-lifecycle]] starts it when config exists; [[ipc-server]] exposes `llm_call`
and folds `StatusSnapshot` into the protocol status; [[cli-promptworld]]'s `llm`
subcommand is the one-shot proof path and its `calibrate` iterates declared
providers; the [[tui-client]] guardian pane renders the provider table and spend;
the meter persists via [[event-log]]'s store. Agent minds (TASK-7), consolidation
(TASK-9), the narrator (TASK-11), and the guardian (TASK-12) are the callers.
[[tool-loop]] drives the tool-use loop wire through `Submit`, pinning its run's
provider via `ResolveProvider` and reporting latency via `ObserveCognition` — used
by both [[agent-mind]]'s `runPlan` and [[guardian]]'s `Turn`; [[social-fabric]]'s
conversation scenes pin per scene via the same field. [[daemon-lifecycle]]'s
governor sampler polls `PendingCognition` and feeds [[cognition]]'s `Debt`/
`Governor`. [[llm-provider-health]] builds on this package's provider/worker/
breaker machinery to make a dead or tool-silent provider operator-visible;
[[daemon-lifecycle]] wires its hook and preflight goroutine. Since spec 035,
[[ipc-server]]'s `set_speed` warning reads `EstimateForKind`/`CalibratedAt`
together, and [[cognition]]'s `Calibrated` is what seeds `calibratedAt`. Since
spec 037, [[agent-mind]]'s `emitSuppressed` feeds `RecordSuppression`, and
[[ipc-server]]'s `horizonClasses` reads `SuppressionCounts` alongside
`EstimateForKind`/`CalibratedAt` for the status wire's horizon.
[[memory-retrieval]]'s embedder driver is the sole caller of the embedding-kind
surface; [[daemon-lifecycle]] wires it at boot when routed. [[grounded-feedback]]
(spec 063) is `KindReportCard`'s sole caller.

## Operational notes

Tested against httptest mock providers: legacy-equivalence (the standing regression
suite — a legacy config's routing/refusals/metering/status pinned byte-identical),
the boot validation matrix, chain-walk skips per reason, pin admission,
no-redispatch, per-provider estimator attribution under concurrent two-provider load,
meter attribution summing (Σ providers + unattributed == total, across store
reopens), lease pools bounding combined in-flight across two orchestrators with
crash reclaim — all under `go test -race`. Live-verified (TASK-35 T019, real
Ollama): conversation → cogito:3b, planner → gemma4:12b-mlx concurrently loaded;
a dead provider's breaker opened after 3 hard failures and the next call recorded
`skipped: bogus (circuit-open)` — post-dispatch failure is final, exactly as
designed. Motivating measurements: one worker serialized everything (TASK-45: 130 s
queue waits behind 19 s calls) while the server ran 4 concurrent cogito calls in
0.98 s wall; 48–128-token structured outputs are 3B-viable while prose is not
(TASK-35 notes) — the division of labor the registry exists to express. Budget
reality check: nightly consolidation ≈ $34/month on the default cloud model, inside
the $100 ceiling.
