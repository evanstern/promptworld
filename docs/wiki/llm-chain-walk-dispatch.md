---
name: llm-chain-walk-dispatch
description: Split from [[llm-orchestrator]] (spec 024 US3 / TASK-52) — Submit's chain-walk admission and recorded skips, the agent tool-use loop's wire shape (Request.Tools/Turns, Response.ToolCalls/Stop), per-provider tool_mode and the kind-scoped LoopMaxRounds/MaxTokens knobs, and the whole-loop latency feed. Load for dispatch-order or tool-loop-wire questions.
kind: component
sources:
  - internal/llm/llm.go
  - internal/llm/providers.go
  - internal/llm/config.go
verified_against: 1fae0d8536eb43e43eaa7b747aaeaf0b6e05ac83
---

# LLM orchestrator — chain-walk dispatch & tool-loop wire

Split from [[llm-orchestrator]] (spec 024 US3, TASK-52): how `Submit` walks a
kind's provider chain, the wire shape the agent tool-use loop rides, and the
per-kind knobs and latency feed around it.

**Chain-walk admission** (`Submit`, spec 024 US3): submission walks the kind's chain
in order and dispatches to the first admissible candidate; a candidate is skipped only
for a mechanical, observable reason — `wallet-exhausted` (priced candidates when the
ceiling is hit), `circuit-open`, `busy` (best-effort only), `queue-full` — recorded in
order on `Response.Skipped` (`[]RouteSkip{Provider, Reason}`). All candidates
inadmissible → the CHAIN HEAD's refusal error (`refusalFor`: the same
`ErrBudgetExhausted`/`ErrTierDown`/`ErrTierBusy`/`ErrQueueFull` sentinels as ever —
single-entry chains behave byte-identically to pre-024). Once a provider accepts a
job its failure is final: never re-dispatched elsewhere. A route may declare
`no_fallback` (single-entry chain enforced at load); `Request.Provider` pins a call to
a named provider, bypassing the walk while honoring that provider's admission
(`ErrUnknownProvider` guards a bad name). Two continuity pins ride this field:
a conversation SCENE resolves its provider once at scene start
([[social-fabric]]), and a tool-loop RUN pins at run start — including across the
spec-025 retry — via `ResolveProvider` ([[tool-loop]]); a persona never switches
voices mid-dialogue, a thought never switches models mid-transcript.
`Response.Provider` always names the serving provider.

**Agent tool-use loop transport** (`llm.go`/`providers.go`, TASK-52, spec 017; every
field additive — a request that sets none marshals byte-identical to before):
`Request.Tools` (`[]ToolDecl{Name, Description, InputSchema}`) declares the round's
tools; `Request.Turns` (`[]Turn{Role, Blocks}`, a `Block` one of text /
`ToolUseBlock` / `ToolResultBlock`) is the ephemeral multi-turn transcript replacing
`Prompt` when non-nil (never persisted, never replayed); `Request.SkipObserve` marks a
loop-internal per-round `Submit` so the worker feeds no fractional per-call sample to
the estimator. `Response.ToolCalls` carries emitted calls in order;
`Response.Stop` (`end_turn`/`tool_use`/`max_tokens`/`other`) is the mapped stop reason
[[tool-loop]]'s driver reads. The Anthropic caller sends native `ToolUnionParam`s and
`tool_use`/`tool_result` blocks (`anthropicInputSchema` round-trips schema keywords
the SDK's typed struct would drop, via `ExtraFields`); `openaiCompat.call` picks a
path per the provider's resolved `tool_mode`: `callNative` sends OpenAI-style
`tools`/`tool_calls`; `callJSON` is the FR-010 fallback for backends whose native
function calling is unreliable — tool catalog appended to the system prompt, every
reply grammar-constrained to a flat `{"tool", "args", "say"}` envelope, per-round call
IDs synthesized (`"env-<round>"` from the assistant-turn count) — a fallback-mode
transcript must keep exactly one assistant turn per round or synthesized IDs collide.

**Tool-call strategy and kind-scoped knobs** (`config.go`): `tool_mode` is
**per-provider** (`ProviderConfig.ToolMode`; legacy `local.tool_mode`/
`cloud.tool_mode` map onto the derived providers), normalized warn-not-error by
`resolveToolMode`, honored only by the `openai_compat` transport — the Anthropic path
is always native. Measured live (TASK-52 T027): cogito:3b never function-calls
natively (88/88 unusable) — its provider entry needs `"json"` wherever a tool-loop
kind can resolve to it. The kind-scoped knobs stay TOP-LEVEL (a property of the
thought class, never the provider — spec 024 R9): `Config.LoopMaxRounds`
(`Rounds()`: absent/0 → 8, clamp 1–16, warn-not-error) and `Config.MaxTokens`
(`*TokenBudgets`, spec 025 / TASK-72) — three per-kind response budgets, `planner`
(default 512), `metatron_turn` (1024, the JSON key is FROZEN), `consolidation` (1024), each normalized by
`PlannerTokens()`/`GuardianTurnTokens()`/`ConsolidationTokens()` (absent/0 → default,
1–4096 verbatim, clamp with warning; a POINTER so `omitempty` genuinely suppresses
the object and pre-025 configs round-trip byte-for-byte — preserved by the
shape-aware v2 `Config.MarshalJSON`). The shared upper clamp is the exported
`llm.MaxTokenBudget` (4096; spec 105 exported the former unexported
`maxTokenBudget`, value unchanged) — it bounds both the knob normalization here
and `internal/mind`'s truncation-retry ladder ceiling
([[nightly-consolidation]]), one source so they can never drift. The daemon
resolves all three at boot and
threads them into `mind.New`/`guardian.New`; conversation (128/224), meeting (72),
narrator (800), and guardian digest (400) budgets are deliberately NOT governed by
these knobs (though since spec 105 the consolidation and narrator budgets are the
STARTS of that ladder, escalating on detected truncation up to the clamp).

**Whole-loop latency feed** (`ObserveCognition(kind, provider, totalMillis)`,
TASK-52/spec 024): the tool-use loop's per-round `Submit`s each ride `SkipObserve`;
the loop reports exactly one whole-cognition wall time, attributed to the named
serving provider (its run pin — empty falls back to the chain head). Both feeding
paths share `feedEstimate`, normalizing by the kind's registered point cost and
firing the same per-provider recalibrate hook; [[tool-loop]]'s `Run` calls this only
on a completed termination (landed / model_done / cap_exhausted), never on the
failure family — mirroring the worker's successes-only doctrine below.

## Connections

Part of [[llm-orchestrator]]'s summary-style split (corpus-spec v2). See
[[llm-orchestrator]] for the overall package doctrine and its other split-off
domains: [[llm-provider-registry]] (declared providers/routes, the embedding
route, transports, config validation), [[llm-concurrency-leases]] (worker
concurrency, priority lanes, endpoint leases, the pending-thought registry),
and [[llm-budget-degraded-mode]] (spend, the circuit breaker, latency
estimation, suppression counters). [[tool-loop]] is this wire shape's driver —
used by both [[agent-mind]]'s `runPlan` and [[guardian]]'s `Turn` — pinning its
run's provider via `ResolveProvider` and reporting whole-cognition latency via
`ObserveCognition`. [[social-fabric]]'s conversation scenes pin per scene
through the same `Request.Provider` field.
