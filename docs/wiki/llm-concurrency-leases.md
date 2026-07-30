---
name: llm-concurrency-leases
description: Split from [[llm-orchestrator]] (spec 024 / TASK-45) — per-provider worker concurrency and slot-aware admission, the conversation priority lane vs BestEffort drop-when-busy, cross-process advisory endpoint leases (spec 024 US5), and the pending-thought registry the governor's debt sampler polls. Load for worker/queue or lease-capacity questions.
kind: component
sources:
  - internal/llm/llm.go
  - internal/llm/lease.go
  - internal/llm/pending.go
verified_against: d0645811c9783d1248dc65ed0fcf0b37524dd8fd
---

# LLM orchestrator — concurrency, priority lanes & leases

Split from [[llm-orchestrator]] (spec 024/TASK-45): per-provider worker
concurrency and admission, priority/best-effort lanes, cross-process endpoint
leases, and the pending-thought registry.

**Concurrency** (TASK-45, per-provider since spec 024): each provider owns `slots`
worker goroutines — N identical copies of one worker loop draining its two channels —
from its `parallel` via `Workers()` (absent/0 → 1, clamp 1–16 `maxLocalWorkers`,
warn-not-error; the daemon prints clamp warnings and the world always boots). An
`atomic.Int32` `inflight` per provider (incremented at dequeue, decremented on every
reply path) drives slot-aware best-effort admission.

**Priority lanes** (per-provider): conversations (`KindConversation`) ride a priority
queue idle workers drain first — dialogue is interactive, planner thoughts tolerate
staleness. The opposite extreme is caller-flagged: `Request.BestEffort` calls are
refused (`ErrTierBusy` / skip reason `busy`) when the candidate has queued work or no
idle slot — flavor yields to real cognition (`meeting.go`'s proposal rephrasing is
the current user; the caller-owned fairness-floor doctrine stands for any future
drop-when-busy kind). A worker-side hard cap (`workerCallCap`, 2 min) bounds any
single provider call so a hung transport can never wedge a provider. **Submit** is
synchronous with immediate admission control — that backpressure surface is what lets
local throughput cap effective sim speed. Bounded queues stay 32 per lane per
provider.

**Advisory endpoint leases** (`lease.go`, spec 024 US5 — closes TASK-24): a provider
declaring `endpoint_capacity` C joins a cross-process lease pool keyed by its
normalized endpoint (lowercased scheme+host, default ports and trailing slash
stripped; sha256[:16] names the pool dir under `~/.promptworld/endpoint-leases/`).
Acquisition is a non-blocking `syscall.Flock(LOCK_EX|LOCK_NB)` sweep over slot files
`slot-00…slot-(C-1)` with jittered ~100 ms retries, in the worker AFTER the
stale-skip check, inside the 2-min call cap, BEFORE the provider call — so combined
in-flight calls across all participating worlds never exceed C, and the TASK-24
mutual breaker-thrash cannot recur. Crash-safe by construction (the kernel frees a
dead process's flocks); lease waiting never strikes the breaker and the estimator
measures from post-acquisition start. A wait over 2 s sets the pool's `contended`
flag (cleared by a sub-2 s acquisition; the flag is POOL-scoped — endpoint congestion
is one truth shared by providers on that endpoint) and surfaces per provider in
status. Undeclared capacity = zero lease syscalls, exactly pre-024 behavior; a
missing home dir disables leases with a warning (warn-not-error).

**Pending-thought registry** (`pending.go`, spec 028 US1): a mutex-guarded
`pendingRegistry` inventories every accepted-but-unfinished job — the adaptive
throttle governor's debt signal. `Submit` `add`s an entry (keyed by a
monotonic id carried on the internal `job`) the instant a candidate accepts,
BEFORE the non-blocking channel send, so a worker that dequeues immediately
can always find it to stamp; the worker's `dispatch` stamps wall time at
dequeue (zero while still queued); a deferred `remove` on every terminal path
of `Submit` (reply, provider error, caller-abandoned ctx) drains the entry, so
the registry empties to zero once all work quiesces — a leaked entry would be
a bug. `Orchestrator.PendingCognition()` snapshots the registry (copy under
the lock, arithmetic outside it) into `[]PendingThought{Kind, Provider,
PredictedSec, ElapsedSec}`: `PredictedSec` is the job's class point cost ×
its provider's CURRENT live seconds-per-point estimate (recomputed at read
time, so it tracks the freshest estimator state including spike rejection),
`ElapsedSec` is wall time since dispatch (0 while queued). The daemon's
governor sampler ([[cognition]], [[daemon-lifecycle]]) is the sole consumer,
polling this every `GovernorCadence` to derive aggregate staleness debt; the
registry itself is orthogonal to routing/metering/breaker machinery and adds
no new call-admission behavior.

## Connections

Part of [[llm-orchestrator]]'s summary-style split (corpus-spec v2). See
[[llm-orchestrator]] for the overall package doctrine and its other split-off
domains: [[llm-provider-registry]] (declared providers/routes, the embedding
route, transports, config validation), [[llm-chain-walk-dispatch]] (chain-walk
admission, the tool-use loop wire shape, per-kind knobs), and
[[llm-budget-degraded-mode]] (spend, the circuit breaker, latency estimation,
suppression counters). [[daemon-lifecycle]]'s governor sampler polls
`PendingCognition` every `GovernorCadence` and feeds it to [[cognition]]'s
`Debt`/`Governor`.
