---
name: llm-preflight-detection
description: Split from [[llm-provider-health]] (spec 034 TASK-84) — the boot + periodic model-existence preflight probe (probeModels, classification into healthy/missing/unreachable/unsupported) and how a successful embedding call clears a stale preflight condition without feeding the tool-silence detector. Load for preflight-probe or embedding-preflight-interaction questions.
kind: component
sources:
  - internal/llm/preflight.go
  - internal/llm/llm.go
verified_against: fc76d2ed3e6995779d392f794f889346704d0aca
---

# LLM provider health — preflight probe & embedding interaction

Split from [[llm-provider-health]] (spec 034, TASK-84): the boot + periodic
model-existence preflight leg, and how embedding traffic clears a stale
preflight condition without ever feeding the tool-silence detector.

**Preflight probe** (`internal/llm/preflight.go`, US1): `probeModels` does a
`GET {endpoint}/models` (the OpenAI-compat listing standard; `/api/tags` is
Ollama-proprietary and rejected as router-specific) with the same Bearer-if-
key auth rule as the chat transport, timeout `preflightTimeout` (5s, a
package var so tests can shrink it). It classifies into `probeHealthy` (the
configured model id is in `data[].id`), `probeMissing` (valid listing,
id absent), `probeUnreachable` (transport error/timeout), or
`probeUnsupported` (non-2xx, or 2xx whose body isn't the `{"data":[…]}`
shape — a router-variance skip, never a false `model-missing`, logged once
via `preflightLogf` and left for the runtime net to cover). `preflightProbe`
reconciles the classification through `setPreflightCondition`/
`clearPreflightCondition` — the former can reclassify DOWNWARD between the
two preflight kinds (unreachable ⇄ missing) by clearing before re-raising,
since `raiseCondition`'s precedence guard alone can't express that direction.
`preflightEligible` scopes probing to `openai_compat` transport providers,
name-sorted (`anthropic` has no local registry to list, FR-001 exempt).
`RunPreflight(ctx)` probes every eligible provider once at boot, then every
`preflightInterval` (60s, package var) re-probes **only** providers still
holding an active preflight condition — a healthy world makes zero
steady-state probe traffic — re-logging the standing warning each cadence
(repeat-loudness) independent of the transition hook.

**Embedding traffic clears preflight, skips tool-silence** (spec 042,
[[memory-retrieval]], TASK-102): `Orchestrator.Embed` calls the embedding
route's head provider's transport directly, bypassing the worker — so embeds
never reach `observeSuccess` and never feed the tool-silence detector
(embeddings carry no tool calls, so tool-silence has no meaning for them).
But a successful embed DOES clear a stale preflight condition
(`model-missing`/`endpoint-unreachable`) on the embedding provider via
`clearPreflightCondition`, mirroring `observeSuccess`'s tool-free branch: a
completed call proves the endpoint reachable and the model served. Before
TASK-102 a bare model alias that resolves fine at the endpoint (e.g.
`all-minilm` vs the listing's `all-minilm:latest`) preflighted against
`data[].id` and stuck with a spurious permanent warning; now that warning is
a transient boot-time blip that self-heals on the first successful embed
(regression-tested in `internal/llm/embed_test.go`). The embedding
subsystem's OWN failure signal is a separate mechanism riding the same wire
shape: `mind.NewEmbedder`'s debounced-per-episode failure callback (wired in
[[daemon-lifecycle]]) prints a daemon-log WARNING and lands a
`daemon.llm_warning` event (`kind: "embedding-unavailable"`) through the same
`Loop.InjectOperator` door this feature uses — the event type and door are
shared, and while this feature's detectors never RAISE from an embed call,
a successful embed does CLEAR its preflight conditions (above).

## Connections

Part of [[llm-provider-health]]'s summary-style split (corpus-spec v2). See
[[llm-provider-health]] for the condition slot, the tool-silence detector,
the daemon/status/TUI surfaces, and the overall doctrine this preflight leg
feeds. [[daemon-lifecycle]] starts `RunPreflight` in its own goroutine at
boot. [[memory-retrieval]]'s embedding traffic is the sole caller whose
success clears a preflight condition here (TASK-102).
