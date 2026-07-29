---
name: llm-provider-health
description: Operator-facing provider health conditions (spec 034, TASK-84) — the condition slot with precedence, the worker-side tool-silence detector, and the daemon/status/TUI surfaces that make a dead or tool-silent local tier loud instead of silently brain-dead; the boot+periodic preflight probe and its embedding-traffic interaction split into a linked child note.
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
verified_against: 1603d5ac22d9be35469ec88bf2355b7d2f9500bc
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

**Preflight probe & embedding interaction** ([[llm-preflight-detection]]):
`probeModels` (`GET {endpoint}/models`, US1) classifies each `openai_compat`
provider healthy/missing/unreachable/unsupported; `RunPreflight` probes once
at boot then re-probes only providers with an active condition every 60s. A
successful embedding call (spec 042, TASK-102) skips the tool-silence
detector entirely but still clears a stale preflight condition — the
TASK-102 fix for a bare-model-alias false positive.

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
`Request.Tools` (planner, Guardian console turns via [[tool-loop]]) count,
per FR-005. `toolSilentRemedy` keys the remedy text to the RESOLVED
`tool_mode` (`config.go`'s `toolModeResolved`): native → suggest
`tool_mode: "json"`; json → the model itself is unsuited for tool work.
Transport failures never reach `observeSuccess` (the worker replies before
the success path) — they neither count nor reset; the breaker owns those.

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
  (`llmConditionWarnings`) per affected provider, right after the clock line.
  Since spec 035, one `providerCalibrationLine` per provider follows right
  after, unconditionally — so even a healthy, uncalibrated world's status
  output now always gains the calibration rows.
- **TUI** ([[tui-client]]): the header gains a red `[llm: <provider> <kind>]`
  badge (`firstLLMCondition`, the `[degraded]` badge's pattern) while any
  condition is active; the provider table (`llmProviderLines` — since
  spec 053 rendered on the dedicated **systems** dock tab, key `5`, rather
  than the guardian pane, per the D10 telemetry split) appends an
  indented, error-styled detail+remedy line beneath an affected provider's
  row.

## Connections

Builds on [[llm-orchestrator]]'s provider/worker/breaker machinery — the
condition slot sits beside `tierHealth`, the detector rides the worker's
success path, and `SetConditionHook` mirrors `SetRecalibrateHook`.
[[daemon-lifecycle]] wires the hook and starts `RunPreflight` at boot.
[[sim-loop]]'s `InjectOperator` door is the sole path the durable
`daemon.llm_warning` event ([[event-types]]) rides while the loop runs; the
reducer no-ops it like `daemon.started`/`stopped` ([[sim-state-reducer]]).
[[cli-promptworld]] and [[tui-client]] render the condition fields;
[[cognition]]'s per-provider estimator and this feature's condition slot are
independent — a governed world and a dead-tier warning can both be true.
Spec 035's calibration-UX surfaces ride alongside this one on the same
status output — orthogonal facts, both can be true. [[memory-retrieval]]'s
embedding traffic never raises this feature's conditions (successful embeds
clear stale preflight ones, TASK-102) and raises its own,
differently-sourced `daemon.llm_warning` through the same door.

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
