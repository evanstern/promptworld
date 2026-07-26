---
name: tool-loop-records-termination
description: Split from [[tool-loop]] (spec 017/058) — the CallRecord/Record artifact sink (record.go, capArgs' 2 KiB truncation), the Termination taxonomy (TermLanded/TermModelDone/TermCapExhausted/TermAdmissionRefused/TermProviderError/TermCtxDone) and their error semantics, and the one-assistant-turn-per-round transcript invariant load-bearing for openaiCompat's synthesized call IDs.
kind: component
sources:
  - internal/toolloop/loop.go
  - internal/toolloop/record.go
verified_against: e137b82bb699eb323eb26c6a69c3dc83ca474b27
---

# Tool-use loop — CallRecord sink, termination taxonomy, and the transcript invariant

Split from [[tool-loop]] (spec 017, TASK-52): the `CallRecord`/`Record` sink,
the `Termination` taxonomy, and the transcript invariant.

**`CallRecord`/`Record` sink** (`record.go`): `CallRecord{JobID, Ordinal,
Tool, Args, Verdict, Reason, Tier}` is the first-class artifact for one model
tool call (FR-007); `{JobID, Ordinal}` is the correlation key. `Args` is a
capped copy (`capArgs`, 2 KiB `maxArgsBytes`) — within the cap, a fresh byte
copy (never aliasing the transcript's buffer); over the cap, it collapses to a
valid JSON string `{"_truncated":true,"prefix":"…"}` with a UTF-8-clean
prefix (a byte-boundary cut that splits a multi-byte rune drops the dangling
partial rather than let `json.Marshal` substitute `U+FFFD`). The driver calls
`j.Record` for every dispatch decision it makes — landed, every rejection
kind, every read outcome, every `unlanded` — so a consumer's telemetry (both
[[agent-mind]]'s mind and [[guardian]] land these as `cog.tool_call` events
via the shared `sim.NewCogToolCallPayload`, [[event-types]], [[cognition]])
can reconstruct the complete call trace even for a cognition where nothing
ever landed. Since spec 058, `verdictRequiresReason` ([[agent-mind]]'s
telemetry emitter) adds `VerdictLandedClamped` to the verdicts whose
`cog.tool_call` MUST carry a non-empty `Reason` — a clamped acceptance has no
query value at all without the notice naming what was truncated.

**Termination taxonomy** (`Termination`, data-model.md §4): `TermLanded` /
`TermModelDone` (the model produced no tool call — Run reports this honestly;
the CONSUMER decides how to record the failure, FR-015) / `TermCapExhausted`
return a nil error; `TermAdmissionRefused` (the submit-side admission ladder —
budget/queue/circuit/best-effort sentinels) / `TermProviderError` /
`TermCtxDone` (context canceled or deadline exceeded) return the underlying
error alongside. `terminationForSubmitErr` maps a `Submit` failure onto one of
the latter three (a `provider_error` `Submit` failure first passes through the
one-per-run transport retry above); a handler's infrastructure failure (`Outcome.Err != nil`)
always terminates the loop with `TermProviderError`, recording the failing
call and every trailing call in the same batch as `unlanded`
(`recordInfraFailure`) — every model tool call still yields exactly one
record even when the loop dies mid-batch.

**Transcript invariant — one assistant turn per round**: the transcript
(`[]llm.Turn`) opens with the seed user turn and each round appends exactly
ONE assistant turn (`assistantEcho`: the model's prose, if any, then one
`llm.Block{ToolUse: ...}` per emitted call, in emission order) followed by one
user turn carrying that round's tool results (`resultBlock` per call). This
one-assistant-turn-per-round shape is load-bearing, not cosmetic: the
`openaiCompat` json-mode fallback ([[llm-orchestrator]]'s `callJSON`)
synthesizes a per-round call ID as `"env-<round>"` from the COUNT OF
ASSISTANT TURNS already in the transcript (`jsonModeRound`), since the flat
envelope carries no ID of its own — any deviation from exactly one assistant
turn per round would collide synthesized IDs across rounds.

## Connections

Part of [[tool-loop]]'s summary-style split (corpus-spec v2); see it for the
package overview and the other domains: [[tool-loop-round-verdict]] (the
`Run` contract, cardinality, and verdict taxonomy) and
[[tool-loop-driver-integration]] (the provider pin, transport retry, the
estimator feed, and roster/schema wiring). [[agent-mind]]/[[guardian]] land
`CallRecord`s as `cog.tool_call` events via `sim.NewCogToolCallPayload`
([[event-types]], [[cognition]]); [[llm-orchestrator]]'s `openaiCompat` fallback
depends on the transcript invariant for its synthesized call IDs.
