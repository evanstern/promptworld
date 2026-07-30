---
name: llm-provider-registry
description: Split from [[llm-orchestrator]] (spec 024) — how providers/routes are declared (Config.Providers/Routes, pricing class, legacy-config derivation), the embedding route's separate head-only dispatch (spec 042), the openai_compat/anthropic transports, and config.go's boot-time validation. Load for provider-declaration or config-shape questions.
kind: component
sources:
  - internal/llm/llm.go
  - internal/llm/config.go
  - internal/llm/providers.go
verified_against: fc76d2ed3e6995779d392f794f889346704d0aca
---

# LLM orchestrator — provider registry, embedding route & transports

Split from [[llm-orchestrator]] (spec 024, TASK-35): how model sources are
declared and validated, the embedding kind's separate dispatch path, and the
two wire transports.

**Providers and routes** (`llm.go`, `config.go`, spec 024): model sources are
**declared** as a named registry (`Config.Providers`) and every call `Kind` maps to an
**ordered chain** of provider names (`Config.Routes`). Chain order is the operator's
complete placement ruling — membership means "meets this kind's quality floor",
position means preference; no runtime scoring, no model ever chooses a model
(decision-5 extends decision-4's deterministic-routing doctrine one level down). Each
provider declares a transport (`openai_compat` — Ollama, LAN routers, 9router — or
`anthropic`, the official SDK), endpoint, model, pricing, `parallel`, per-provider
`reasoning_effort`/`tool_mode`, and optional `endpoint_capacity`. "Tier" retired as a
routing concept; the surviving local-vs-cloud distinction is **pricing class**
(`provider.priced()`): zero-priced providers are never budget-refused and seed
local-class latency bootstraps. The legacy two-entry config (`local`/`cloud`) loads
forever via `deriveLegacy` — a two-provider registry named `local`/`cloud` with the
pre-024 routes (planner/conversation/meeting → local; consolidation/narrator/drama/
metatron → cloud), byte-identical behavior; declaring both shapes in one file is a
load error. `KindMusing` retired with spec 017: musing is a roster tool inside the
planner loop ([[agent-mind]], [[tool-loop]]). Spec 029 adds `KindGuardianWatch`
(`"metatron_watch"` — the route-key string is FROZEN, spec 052 ruling 2) — the
guardian's fuzzy standing-order confirm, a single bare
yes/no `Submit` (never a tool loop, [[guardian-orders]]); it is the one kind
whose `defaultRoutes()` chain is MULTI-ENTRY (`["local","cloud"]` — cheap-first
local for the yes/no, cloud fallback), and it maps to [[cognition]]'s existing
`metatron` decision class (the class-name string is likewise frozen). Spec
063 ([[grounded-feedback]]) adds `KindReportCard` (`"report_card"`, frozen
from birth) on the identical shape — cheap-first `local→cloud`, mapped to
the same `metatron` decision class — for the guardian's report-card
critique: one bounded call per stopping point, never a tool loop. Both new
kinds are in `defaultBackfillKinds`, so a pre-063/pre-029 `llm.json`
backfills the route from `defaultRoutes()` with a boot log line rather than
failing to load.

**The embedding kind** (`llm.go`/`config.go`/`providers.go`, spec 042,
[[memory-retrieval]]): `KindEmbedding` (`"embedding"`) is a valid route key
that sits OUTSIDE `acceptedKinds` — it never dispatches through `Submit`'s
chat machinery or the cognition decision-class registry (`Submit` returns
`ErrUnknownKind` for it), and `validateV2` exempts it from the
route-completeness check: an ABSENT `embedding` route means the subsystem is
off (a vectorless world), never a boot error and never a warn-backfill — the
same "no llm.json → reflex-only" absence-is-the-switch doctrine. A PRESENT
route gets full chain validation plus one transport rule: naming an
`anthropic` provider anywhere in its chain is a boot config error, since the
Messages API serves no embeddings endpoint. `New` diverts the resolved route
into its own `embedding *provider` slot — HEAD ONLY, never a chain (a vector's
model identity travels with it, so silent chain fallback would mix models) —
and never adds it to the routes table Submit walks. `HasEmbedding`/
`EmbeddingProvider` are the daemon's wiring gate and boot-line source;
`Embed(ctx, texts)` calls the head provider's transport directly — no queue,
breaker, or estimator, since the embedder driver paces and debounces itself —
returning `ErrEmbeddingOff` when the subsystem is off; on success it clears
any stale preflight condition on the embedding provider (TASK-102, see
[[llm-provider-health]]); `WarmEmbedding` best-
effort pins the model resident. Both transports implement the `embedCaller`
surface: `openaiCompat.Embed` POSTs `{"model","input"}` to
`endpoint+"/embeddings"`, decoding `data[].embedding` into one index-ordered
vector per input text (defensively re-ordered by the wire's `index` field),
and `openaiCompat.WarmEmbed` hits the Ollama-NATIVE `/api/embed` with
`{"model","keep_alive":-1}` (the compat endpoint's own `keep_alive` is
ignored, so the native call is the only way to pin it); `anthropicCaller.Embed`
returns the typed `ErrEmbeddingUnsupported` sentinel — unreachable through a
validated registry, but present so both transports satisfy `embedCaller`.

**Transports** (`providers.go`): `openai_compat` speaks chat-completions over raw
HTTP, pins `stream: false` (some routers stream by default), and carries
`max_tokens` (from `Request.MaxTokens`, when positive) plus the provider's resolved
`reasoning_effort` (TASK-37: thinking-default models like gemma4 otherwise free-run
hidden chain-of-thought — live diagnosis measured 2–6 s calls inflated to 60–120 s);
`resolveReasoningEffort` keeps the nil/"" convention — zero-priced absent defaults
`"none"`, priced absent (and explicit `""` anywhere) sends nothing. `anthropic` uses
`anthropic-sdk-go` against the Messages API with `cache_control` on system blocks so
stable prompts (souls, charters) bill at cache-read rates. `newProviderCaller` builds
the right caller per declared transport. A TASK-58 `ResponseSchema`/`SchemaName`
structured-output path (deleted as dead code in TASK-71) is back, restored by spec
103/TASK-174 for the conversation route: `callNative` attaches `response_format
{type: json_schema}` iff `Request.ResponseSchema` is set and the request carries
no `Tools` (payload byte-identical otherwise); `anthropicCaller` never reads either
field. [[social-fabric-conversations]] stamps the two conversation-scene schemas.

**Config** (`config.go`): `llm.json` in the save directory, written v2 by
`promptworld new`; deleting the file disables the orchestrator entirely. Hosted keys
are never stored — only an env var NAME (`api_key_env`, default `ANTHROPIC_API_KEY`);
the optional inline `api_key` is for LAN-router keys only and wins when both are set.
`resolveRegistry` is the single validation authority (LoadConfig and `New` both call
it, dispatching to `validateV2` for v2): boot ERRORS name the offender for a route to
an undeclared provider, an accepted
kind with no route, an unknown kind key, a duplicate provider in a chain, an empty
chain, `no_fallback` with chain length > 1, missing transport/model, `openai_compat`
without endpoint, or both config shapes at once; tuning knobs clamp with warnings,
never errors. One narrow exception to the completeness check (spec 029, research
R8): kinds in `defaultBackfillKinds` — those introduced AFTER the v2 format shipped,
`KindGuardianWatch` and, since spec 063, `KindReportCard` — are BACKFILLED from `defaultRoutes()` with a
boot log line (`configWarnf`, warn-not-error) rather than failing boot, so a v2
`llm.json` written before the kind existed keeps booting on upgrade. This runs
before the completeness loop; a missing route for any OTHER kind is still fatal, and
an unknown route KEY is still a boot error. `RouteConfig.UnmarshalJSON` accepts the bare-array shorthand
(`["a","b"]`) and the `{chain, no_fallback}` object; `MarshalJSON` re-emits the
shorthand, and the shape-aware `Config.MarshalJSON` round-trips both shapes —
including top-level `max_tokens` — byte-for-byte.

## Connections

Part of [[llm-orchestrator]]'s summary-style split (corpus-spec v2). See
[[llm-orchestrator]] for the overall package doctrine and its other split-off
domains: [[llm-chain-walk-dispatch]] (chain-walk admission, the tool-use loop
wire shape, per-kind knobs), [[llm-concurrency-leases]] (worker concurrency,
priority lanes, endpoint leases, the pending-thought registry), and
[[llm-budget-degraded-mode]] (spend, the circuit breaker, latency estimation,
suppression counters). [[memory-retrieval]]'s mind-side embedder driver is the
sole caller of the embedding-kind surface; [[daemon-lifecycle]] wires it at
boot only when `llm.json` routes the kind.
