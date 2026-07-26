---
name: tool-loop-driver-integration
description: Split from [[tool-loop]] (spec 024/025) — the run-level provider pin (ResolveProvider stamped once per run), the one-per-run transport retry (provider_error classification, Result.Retried/RetryReason), the successes-only whole-loop estimator feed to Orchestrator.ObserveCognition, and how Run builds []llm.ToolDecl from the tool-registry roster.
kind: component
sources:
  - internal/toolloop/loop.go
verified_against: e137b82bb699eb323eb26c6a69c3dc83ca474b27
---

# Tool-use loop — provider pin, transport retry, and roster/schema wiring

Split from [[tool-loop]] (spec 017/024/025, TASK-52/TASK-72): the run-level
provider pin, the transport-retry matrix, the successes-only estimator feed,
and how `Run` wires the roster into wire-level tool declarations.

**Run-level provider pin** (spec 024 R9, the FR-008 extension; spec 025
composition): an empty `Job.Provider` is NOT unpinned — `run` resolves the
kind's provider ONCE at run start (`submitter.ResolveProvider`, a dry
chain-walk naming the current admissible chain head) and stamps
`Request.Provider` on EVERY round's `Submit`, so a multi-round cognition never
changes models mid-transcript. The transport retry below inherits the pin
structurally (it re-enters the same `Submit`), so a retry always re-hits the
SAME provider; a genuinely down pinned provider fails the run per spec 025
semantics and the NEXT cognition's resolve walks the chain to a fallback. This
is the cognition-run analog of [[social-fabric]]'s conversation scene pin: with
per-call chain-walking, the breaker strike from the very failure being retried
would itself be a walk trigger — a mid-run switch would mix native vs JSON
tool-call ID conventions and mis-attribute the whole-run observation. An
explicit `Job.Provider` is honored as-is, never re-resolved; a `ResolveProvider`
miss (a test stub without the seam) leaves the pin empty, falling back to
per-kind routing.

**Transport retry — one per run** (spec 025, TASK-72,
`specs/025-llm-robustness-knobs/contracts/loop-retry.md`): when a `Submit`
fails and `terminationForSubmitErr` classifies it `provider_error` (transport;
NOT the admission-ladder sentinels, NOT context death), the loop re-submits
the identical transcript once — a failed `Submit` appended nothing, so the
retry is byte-identical, and it consumes no round (`rounds` counts model
responses). On a second transport failure, or the first after the run's retry
is spent, the loop terminates `provider_error` with the latest error exactly
as a single failure did pre-025. Admission refusals and ctx-done never retry
(the governor spoke; busy-is-not-down), and a handler infrastructure failure
is not a transport failure (the model call succeeded; handlers are
side-effectful) — those paths are unchanged. `Result.Retried` /
`Result.RetryReason` (the FIRST failure's text; non-empty iff `Retried`)
report the consumed retry for the consumer to surface as a NON-terminal
`cog.outcome` carrying `sim.OutcomeRetried` — the TASK-42 conversation
vocabulary, so no new event type — making every recovery countable from the
trail alone. Estimator/breaker doctrine is untouched structurally: the
retried `Submit` rides `SkipObserve` like any round, a recovered run ends in
the success family and feeds exactly one `ObserveCognition`, a twice-failed
run feeds zero, and each `Submit` strikes the breaker as an independent call.

**Successes-only whole-loop estimator feeding**: `Run`'s deferred exit hook
always sets `res.TotalMillis` (part of `Result` regardless of outcome), but
feeds `Orchestrator.ObserveCognition(j.Kind, pin, res.TotalMillis)` — the run
PIN names the provider whose estimator receives the sample, exact by
construction since every round Submitted to it — ONLY on a
completed termination — `TermLanded`, `TermModelDone`, `TermCapExhausted` —
each of which measured completed model work (`TermCapExhausted` did N full
provider rounds). The failure family (`TermAdmissionRefused`/
`TermProviderError`/`TermCtxDone`) did no completed thought and feeds nothing,
so a refused or errored loop cannot skew the governor's EWMA toward zero —
mirroring [[llm-orchestrator]]'s own per-call worker doctrine ("a fast
failure is not a latency observation of completed thought"). Every
per-round `Submit` inside the loop sets `Request.SkipObserve: true` so no
fractional per-round sample separately reaches the estimator; the whole-`Run`
observation is the ONLY sample a loop cognition contributes, in the SAME unit
([[cognition]]'s `TierProfile.SecondsPerPoint` doctrine) a single-shot kind's
one-call wall time is.

**Roster and schema wiring**: `Run` builds the wire-level `[]llm.ToolDecl`
from `j.Roster` — `Name`, `Description: t.PromptGloss`, `InputSchema:
tool.InputSchema(t)` ([[tool-registry]]) — once per invocation; the roster
itself (`tool.LoopRosterVillager()` / `tool.LoopRosterGuardian()`) and its
authored or derived schemas are the tool registry's responsibility, not this
package's.

## Connections

Part of [[tool-loop]]'s summary-style split (corpus-spec v2); see it for the
package overview and the other domains: [[tool-loop-round-verdict]] (the
`Run` contract, cardinality, and verdict taxonomy) and
[[tool-loop-records-termination]] (`CallRecord`, termination taxonomy, the
transcript invariant). [[llm-orchestrator]] is the transport this note's pin
and retry drive, and the estimator (`ObserveCognition`) this note feeds;
[[tool-registry]] supplies the roster (`LoopRosterVillager`/
`LoopRosterGuardian`) and per-tool wire schema this note wires in.
