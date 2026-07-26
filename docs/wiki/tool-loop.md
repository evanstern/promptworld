---
name: tool-loop
description: The bounded agent tool-use loop driver (spec 017) — submit/dispatch/feed-back to one landed action or a hard cap, transport-agnostic and sim-agnostic, shared by the villager planner and the Guardian's console turn. Overview + connections here; the round/verdict contract, the CallRecord/termination machinery, and provider-pin/retry/estimator driver integration each split into their own child note.
kind: component
sources:
  - internal/toolloop/loop.go
  - internal/toolloop/record.go
verified_against: e137b82bb699eb323eb26c6a69c3dc83ca474b27
---

# Tool-use loop

`internal/toolloop` (spec 017, TASK-52) is the bounded loop driver every agent
cognition that "acts by calling tools" runs through: submit → dispatch → feed
results back → repeat, until an acting call lands, the model finishes on its
own, the round cap trips, or the transport fails. It replaces the pre-spec-017
pattern of one bare model call whose free-text reply a consumer package
hand-parsed against a hand-maintained vocabulary. The package is deliberately
transport-agnostic and sim-agnostic: it imports only `internal/llm` (the wire)
and `internal/tool` (the schema/roster source), and leaves handlers, artifact
recording, and event emission to the consumer — a shared leaf below both
[[agent-mind]] and [[guardian]] (research R1).

## How it works

This note's file-level detail splits into three children by domain
(corpus-spec v2 summary-style split); each links back here.

**Round contract and verdict taxonomy** — the TASK-52 request/event/gate/
executor doctrine, the `Run(ctx, orch, Job)` contract and its guarantees, the
one-landed-acting-call cardinality rule (Read tools exempt), and the driver's
full `Verdict` taxonomy including the spec-058 `VerdictLandedClamped`
clamp-with-notice shape — moved to [[tool-loop-round-verdict]].

**CallRecord sink, termination taxonomy, and the transcript invariant** — the
`CallRecord`/`Record` artifact sink (2 KiB arg-capping), the `Termination`
taxonomy (`TermLanded`/`TermModelDone`/`TermCapExhausted`/
`TermAdmissionRefused`/`TermProviderError`/`TermCtxDone`) and their error
semantics, and the one-assistant-turn-per-round transcript invariant load-
bearing for `openaiCompat`'s synthesized call IDs — moved to
[[tool-loop-records-termination]].

**Provider pin, transport retry, and roster/schema wiring** — the run-level
provider pin (`ResolveProvider` stamped once per run), the one-per-run
transport retry (`provider_error` classification, `Result.Retried`/
`RetryReason`), the successes-only whole-loop estimator feed to
`Orchestrator.ObserveCognition`, and how `Run` builds `[]llm.ToolDecl` from
the tool-registry roster — moved to [[tool-loop-driver-integration]].

## Connections

[[llm-orchestrator]] is the transport `Run` drives: `Request.Tools`/`Turns`/
`SkipObserve` out, `Response.ToolCalls`/`Stop` back, and
`Orchestrator.ObserveCognition` for the whole-loop latency feed.
[[tool-registry]] supplies the declared roster (`LoopRosterVillager`/
`LoopRosterGuardian`) and each tool's wire schema (`InputSchema`,
`InputSchemaJSON` overrides). [[agent-mind]]'s `runPlan` is the villager
consumer: it builds a `villagerDispatch`, wraps every acting tool's landing
door in `villagerHandlers` (`internal/mind/handlers.go`), and reads `res.Term`
to decide the terminal `cog.outcome` and rearm exactly as the pre-loop
rejection/failure paths did. [[guardian]]'s `Turn` is the console consumer:
its `turnHandlers` wrap the spec-029 agency surface (`send_vision`/`send_omen`,
`monitor_and_act`/`cancel_order`, `work_miracle`, and the meta tools
`pause`/`start`/`adjust_speed` — see [[guardian-orders]]), and `converse` is
deliberately NOT a declared tool — the model's closing prose
(`Result.Final`) is the transcript-only answer channel. [[cognition]] owns
the decision-class registry and staleness router both consumers gate on
before ever calling `Run`; [[event-types]] catalogs `cog.tool_call`, the
event both consumers land from buffered `CallRecord`s.

## Operational notes

The package has no environment variables and no persisted state of its own —
the transcript is ephemeral (never persisted, never replayed) and every
durable trace is the consumer's `CallRecord` emission. `contracts/loop-api.md`
and `data-model.md` (`specs/017-agent-tool-loop/`) are the authored contract
this note grounds; `loop_test.go`/`equivalence_test.go`/`governor_test.go`/
`adversarial_test.go` exercise the cardinality rule, the termination
taxonomy, the successes-only estimator feed, and adversarial model behavior
(over-cap calls, malformed args, an unknown tool name) respectively;
`retry_test.go` (spec 025) locks the transport-retry matrix — fail-once
recovery, fail-twice termination, admission/ctx/handler failures never
retried, round-cap and estimator invariance.
