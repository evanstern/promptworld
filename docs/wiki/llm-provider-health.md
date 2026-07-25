---
name: llm-provider-health
description: Operator-facing provider health conditions (spec 034, TASK-84) — a boot + periodic model-existence preflight, a worker-side tool-silence detector, one condition slot per provider with precedence, and the daemon/status/TUI surfaces that make a dead or tool-silent local tier loud instead of silently brain-dead
kind: component
sources:
  - internal/llm/llm.go
  - internal/llm/preflight.go
  - internal/llm/config.go
  - internal/sim/loop.go
  - internal/sim/state.go
  - internal/daemon/daemon.go
  - cmd/promptworld/commands.go
  - internal/tui/views.go
verified_against: 1af833a2c4dab23932357d85cbf51e01089d66fc
---

# LLM provider health (preflight + tool-silence detection)

Spec 034 (TASK-84) closes a silent-failure gap in [[llm-orchestrator]]: a
fresh world whose configured local model is absent from its endpoint, or
whose model answers but never emits tool calls, produced zero visible signal
anywhere — villagers just sat at the reflex floor forever. This feature adds
one **operator-facing condition** per provider, two detection legs (a boot
+ periodic existence preflight, and a runtime tool-silence counter), and
wires both into the daemon log, the status wire, `promptworld status`, and
the TUI — all without touching the deterministic sim loop's semantics or
redesigning the existing circuit breaker.

## How it works

**The condition slot** (`internal/llm/llm.go`): each `provider` gains a
`condMu`-guarded `providerCondition{kind, detail, remedy, since}` beside its
existing `tierHealth` breaker — a deliberately separate piece of state: the
breaker counts transport failures (its own concern), while the condition is
the operator diagnosis. `ConditionKind` is `model-missing` |
`endpoint-unreachable` | `tool-silent` (empty = healthy). `condPrecedence`
ranks them `endpoint-unreachable` (3) > `model-missing` (2) > `tool-silent`
(1) — one slot, dominant problem wins: `raiseCondition` drops a raise ranked
below the active condition and no-ops an identical (kind, detail) repeat, so
a steady condition fires the transition hook exactly once; `clearCondition`
fires it again with `active:false`. `Orchestrator.SetConditionHook` installs
the sole consumer (mirroring `SetRecalibrateHook`); `fireCondition` always
runs it outside any provider lock.

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

**Tool-silence detector** (`observeSuccess`, `internal/llm/llm.go`, US2): the
worker's success path (after `t.health.succeed()`) calls
`o.observeSuccess(t, len(j.req.Tools) > 0, len(cr.toolCalls) > 0)` on every
COMPLETED call. A landed tool call proves full health: it clears ANY
condition (including tool-silent) and resets `consecutiveToolFree` to 0. A
tool-free success on a tool-carrying call proves the model exists and
answers — so it clears a stale PREFLIGHT condition — but not that
tool-calling works, so it increments `consecutiveToolFree` and, at
`toolSilentThreshold` (8, a package var — TASK-73 soaks sustained 789–982
tool calls per 8 game-hours, so a run of 8 consecutive tool-free completions
never occurs on a healthy provider but flags a never-function-calling model
within minutes), raises `tool-silent` via the precedence-guarded
`raiseCondition` (a live preflight condition keeps the slot). A tool-free
success on a NON-tool call (conversation, meeting) touches neither the
counter nor a tool-silent condition — only kinds that declare
`Request.Tools` (planner, Metatron console turns via [[tool-loop]]) count,
per FR-005. `toolSilentRemedy` keys the remedy text to the RESOLVED
`tool_mode` (`config.go`'s `toolModeResolved`): native → suggest
`tool_mode: "json"`; json → the model itself is unsuited for tool work.
Transport failures never reach `observeSuccess` (the worker replies before
the success path) — they neither count nor reset; the breaker owns those.

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

**Fresh-world defaults** (US3) close the loop from the other end — see
[[llm-orchestrator]]'s `DefaultConfig` for the `cogito:3b` + `tool_mode:
"json"` default that makes this feature's warnings the safety net rather
than the first-run experience.

## Surfaces

- **Daemon log**: every transition prints `daemon: WARNING llm provider
  "<name>": <detail> — <remedy>` (or a recovered line on clear); the
  preflight's periodic re-probe additionally re-logs the standing warning
  every 60s while active, via `preflightLogf`, without re-firing the hook.
- **Durable event** (`daemon.llm_warning`, [[event-types]]): the daemon's
  condition hook (wired in [[daemon-lifecycle]]) marshals
  `sim.LLMWarningPayload{provider, kind, detail, remedy, active}` and lands
  it through [[sim-loop]]'s `Loop.InjectOperator` door — required because the
  hook fires from worker/preflight goroutines *while the loop runs*, and
  `store.AppendEvents` has no internal locking (the loop is the log's sole
  writer). Transitions only, so the durable stream stays quiet under a
  steady condition; if the loop isn't running (the shutdown window),
  `InjectOperator` errors and the hook degrades to the log line alone — the
  status fields still carry the condition.
- **Status wire** (`ProviderStatus`, [[llm-orchestrator]]): three `omitempty`
  fields, `Condition`/`ConditionDetail`/`ConditionRemedy`, populated by
  `StatusSnapshot` from each provider's `conditionSnapshot()`; a healthy
  provider marshals byte-identically to a pre-034 world for these three
  fields (spec 035 adds its own additive `CalibratedAt` field alongside them
  — [[llm-orchestrator]]).
- **`promptworld status`** ([[cli-promptworld]]): `renderStatusHuman` prints
  one `WARNING llm provider "<name>": <detail> — <remedy>` line
  (`llmConditionWarnings`) per affected provider, right after the clock line;
  no active condition renders byte-identical pre-034 output for this block.
  Since spec 035 (FR-004/US3), one `providerCalibrationLine` per provider
  follows right after — unconditionally, whenever the world has an LLM
  section, independent of condition state — so a healthy, uncalibrated world's
  human status output is no longer byte-identical to pre-034: it now always
  gains the calibration rows.
- **TUI** ([[tui-client]]): the header gains a red `[llm: <provider> <kind>]`
  badge (`firstLLMCondition`, the `[degraded]` badge's pattern) while any
  condition is active; the metatron pane's provider table
  (`llmProviderLines`) appends an indented, error-styled detail+remedy line
  beneath an affected provider's row.

## Connections

Builds on [[llm-orchestrator]]'s `provider`/worker/breaker machinery — the
condition slot sits beside `tierHealth`, the detector rides the worker's
existing success path, and `SetConditionHook` mirrors
`SetRecalibrateHook`. [[daemon-lifecycle]] wires the hook and starts
`RunPreflight` in its own goroutine at boot. [[sim-loop]]'s `InjectOperator`
door is the sole path the durable `daemon.llm_warning` event
([[event-types]]) rides while the loop runs; the reducer no-ops it exactly
like `daemon.started`/`stopped` ([[sim-state-reducer]]). [[cli-promptworld]]
and [[tui-client]] render the condition fields; [[cognition]]'s per-provider
estimator and this feature's condition slot are independent — a governed,
throttled world and a dead-tier warning can both be true at once. Spec 035's
calibration-UX surfaces (the boot `uncalibratedBootWarning`, `set_speed`'s
uncalibrated warning, and `providerCalibrationLine`) are a separate,
independent signal riding alongside this one on the same `renderStatusHuman`
output and the same `ProviderStatus` struct — a dead/tool-silent condition
and an uncalibrated provider are orthogonal facts that can both be true.
[[memory-retrieval]]'s embedding traffic never raises this feature's
conditions (successful embeds clear stale preflight ones, TASK-102) and
raises its own, differently-sourced `daemon.llm_warning` through the same
door.

## Operational notes

Tested (`internal/llm/preflight_test.go`, `condition_test.go`,
`detector_test.go`) against `httptest` fake OpenAI-compat servers: a
listing with/without the configured id, a 404 (listing-unsupported skip),
and connection-refused (unreachable), asserting condition kind, remedy text,
boot-never-fails, and clear-on-re-probe; the detector's raise-at-threshold /
reset-on-tool-call / clear-on-success / non-tool-kinds-never-count matrix;
and the status-wire/`cmdStatus`/TUI-badge rendering. `toolSilentThreshold`
and `preflightInterval`/`preflightTimeout` are package vars (not consts)
specifically so tests can cross the threshold or compress the clock without
a pathological real-time run.
