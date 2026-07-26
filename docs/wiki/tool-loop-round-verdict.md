---
name: tool-loop-round-verdict
description: Split from [[tool-loop]] (spec 017/025/029/058) — the TASK-52 request/event/gate/executor doctrine, the Run(ctx, orch, Job) contract and its guarantees, the one-landed-acting-call cardinality rule (Read tools exempt), and the driver's full Verdict taxonomy including the spec-058 VerdictLandedClamped clamp-with-notice shape.
kind: component
sources:
  - internal/toolloop/loop.go
  - internal/toolloop/clamp.go
verified_against: e137b82bb699eb323eb26c6a69c3dc83ca474b27
---

# Tool-use loop — round contract and verdict taxonomy

Split from [[tool-loop]] (spec 017, TASK-52): the doctrine, the `Run`
contract, the cardinality rule, and the verdict taxonomy.

**Doctrine, preserved verbatim from the TASK-52 design decisions**: a tool
call is a REQUEST; an event is the FACT; the gate decides; the executor
grounds work in time and space. The driver enforces bounds and RECORDS
requests — it never mutates world state itself. Every durable effect flows
through a handler that wraps an existing landing door (`InjectIntent`, the
`InjectSocial` whitelist), so the loop cannot manufacture a fact the gates
would not admit. Reads return data and ground nothing. Speaking, musing, and
thinking are tools too — game-state integrity applies to expression, not only
world mutation.

**`Run` contract** (`Run(ctx, orch *llm.Orchestrator, j Job) (Result, error)`,
delegating to an unexported `run` over a `submitter` interface — `Submit` +
`ObserveCognition` + `ResolveProvider` (the run-level pin seam, spec 024) — so
the control flow is unit-testable against a scripted
stub with no network or real orchestrator): a `Job` carries `JobID` (the
existing cognition job identifier, threading every `CallRecord`), `Kind`,
`System`, `Seed` (the initial user turn), `Roster []tool.Tool`, `Handlers
map[string]Handler`, `MaxRounds`, `MaxTokens`, an optional `Provider` (an
explicit pin — `promptworld calibrate` sets it so a reference sample measures a
NAMED provider; empty for every live mind/guardian caller), and
`Record func(CallRecord)`
(the artifact sink; the consumer buffers/lands records — never touched by the
driver beyond calling it). `MaxRounds <= 0` is defensively treated as 1 (the
real normalization is `llm.Config.Rounds()`, upstream). `Run` guarantees
(contracts/loop-api.md): it terminates within `MaxRounds` provider rounds; at
most one acting call lands; every model tool call yields exactly one
`CallRecord` via `j.Record` (ordinals 1-based, dense, emission-ordered); a
read-effect tool never consumes the action; `SkipObserve` rides every internal
`Submit`; the governor estimator is fed the whole-`Run` wall time only on
a completed termination (successes-only, below); and a transport-level
provider failure is retried EXACTLY ONCE per run (spec 025, below) before it
terminates.

**Cardinality — one landed acting call, reads exempt**: a tool is "acting"
(`isActing`) iff its `tool.EffectClass` is `World` or `Expressive`; a `Read`
tool does not consume the cognition's one action. Once an acting call has
landed within a response, EVERY remaining call in that same response —
including further reads — is rejected `rejected_cardinality`: the cognition's
one action is spent (FR-004, R8). A read-effect tool dispatched on a non-final
round returns its data and grounds nothing; dispatched on the FINAL round (at
the cap) it is instead recorded `unlanded` without ever calling its handler —
the loop is out of rounds to make use of what it would learn. An acting call,
by contrast, is dispatched on every round including the last — it can land as
the terminal answer without needing a follow-up round.

**Verdict taxonomy** (`Verdict`, data-model.md §5): the DRIVER owns
`rejected_unknown` (the call names a tool not on this cognition's roster, or
one with no registered handler), `rejected_malformed` (driver-side schema/
param validation — `validateArgs`, catching missing required args, wrong
scalar types, enum membership, number bounds, text caps; a tool with an
authored `InputSchemaJSON` override is instead validated against THAT schema
by a general schema-lite walker — `validateAuthored`/`walkSchema`, spec 029 R5
— NOT against `set_plan`'s shape as the retired `validateSetPlan` did), and
`rejected_cardinality`. A handler's returned `Outcome`
owns `landed`, `rejected_gate` (the door refused — stale, guard, scene,
charge), `read_ok`, and `read_error`. `unlanded` covers a call the loop
terminated before dispatching (cap reached, or a trailing call after an
infrastructure failure in the same batch). Since spec 058 (FR-001/FR-003,
TASK-110, `clamp.go`) a landed call has a second shape: `VerdictLandedClamped`
carries every `landed` control-flow consequence (consumes the action, ends the
loop) but marks that a Clamp-flagged expressive field — or set_plan's own step
count — was truncated to its cap rather than the whole call being rejected;
`validateArgs` returns `(finalArgs, clampNotice, rejectReason)` instead of a
bare reject string — `rejectReason != ""` still means reject exactly as
before, but a non-empty `clampNotice` on an otherwise-passing call means
`finalArgs` carries a Clamp-flagged `Text` param (or the authored-schema path's
top-level `reason`, clamped by field name via `clampTopLevelText` since an
`InputSchemaJSON` override bypasses `Param` derivation) truncated rune-safely
by `clamp.go`'s `ClampRunes`/`ClampBytes` (never splitting a UTF-8 sequence,
the `NormTextMax` idiom factored out for every byte-cap caller); `run` rewrites
`call.Args` to the clamped value in place BEFORE dispatch, so the handler, the
`CallRecord`, and the eventual event payload all see only the truncated text.
A clean `VerdictLanded` whose args were clamped upstream is upgraded to
`VerdictLandedClamped` by `run` itself once the handler returns; a handler may
also originate `VerdictLandedClamped` directly when its clamp condition is
domain-specific and the loop can't detect it generically (`set_plan`'s own
step-count clamp, [[agent-mind]]'s `handleSetPlan`) — either way the
model-facing result gains a `withClampNotice` suffix naming the field and the
clamp (FR-001's "the model can adapt"), and the recorded `Reason` carries that
same notice (or, when the handler originated the clamp itself, its own
`ResultForModel` phrasing) so a clamped acceptance is queryable exactly like a
rejection, never silent. Every model tool call ends with exactly one of these
verdicts.

## Connections

Part of [[tool-loop]]'s summary-style split (corpus-spec v2); see it for the
package overview and the other domains: [[tool-loop-records-termination]]
(`CallRecord`, termination taxonomy, the transcript invariant) and
[[tool-loop-driver-integration]] (the provider pin, transport retry, the
estimator feed, and roster/schema wiring). [[tool-registry]] is the source of
the `Param.Clamp` flag this note's `VerdictLandedClamped` consumes;
[[agent-mind]]/[[guardian]] are the consumers deciding a terminal `cog.outcome`
from the verdict a call lands with.
