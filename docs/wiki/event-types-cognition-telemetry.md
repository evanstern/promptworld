---
name: event-types-cognition-telemetry
description: Cognition/planning event rows split from [[event-types]]: cog.thought/outcome, agent.intent_rejected, cog.recalibration_recommended/tool_call/memory_divergence, agent.plan_set, agent.plan_step_started/expired. Load when tracing the cog.* observability family, tool-call verdicts (spec 058 clamping), or plan landing/expiry.
kind: concept
sources:
  - internal/sim/loop.go
  - internal/sim/landing.go
verified_against: 9b4ed5aef5bfea50b67fac10f8e2153f065a814d
---

# Event types — cognition telemetry & planning events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.

Spec 086 (agent-named payloads): every agent-referencing field in this
family's payloads is a `sim.AgentRef` — the wire carries
`{"id":N,"name":"…"}` objects (lists element-wise), the name stamped at
emission from the fixed roster via `Ref`/`Refs`; sentinels marshal
`{"id":-1,"name":""}`. Legacy bare-int rows decode through the dual-shape
unmarshal forever and reducer arms fold `.ID`s only — the conventions and
the normative back-compat matrix live on [[event-types]] ("Agent
references are named refs").
Spec 043 (context grounding — [[decision-context]]) adds NO new event type at
all: `Agent` gains `omitempty` `IntentLog []IntentRecord` (the recent-intent
ring, cap 8) and `NeedsAnchor`/`NeedsAnchorTick` (the need-trajectory window
anchor), both maintained by EXISTING intent-lifecycle and needs arms (rows
below note each), so every pre-043 snapshot round-trips byte-identically —
and `agent.intent_rejected`, formerly a reducer no-op, becomes state-mutating
(it appends an already-closed ring record). `CogThoughtPayload` gains three
additive LAST `omitempty` fields — `prompt_bytes`/`block_bytes`/
`dropped_blocks`, the assembled decision-prompt size, per-block byte
breakdown, and budget-dropped blocks — stamped ONLY on planner-class
emissions, so pre-043 events and every non-planner emission marshal
byte-identically.

Spec 058 (tool surface hygiene — clamp expressive text, prune dead verbs —
[[tool-registry]], [[tool-loop]], TASK-110) likewise adds no new event type:
every landed event already carried post-validation text, so a clamped
expressive field (`say`/`gist`/`muse`/`reason`) or an oversized `set_plan`
lands with EXACTLY the truncated value the model's over-cap submission never
gets to see verbatim — replay and downstream consumers see only what was
accepted, unchanged. The feature's visibility rides EXISTING vocabulary
instead: `toolloop.VerdictLandedClamped` (a new `Verdict` string,
`"landed_clamped"`) distinguishes a clamped `cog.tool_call` from a clean
`landed` one, and `sim.OutcomeClamped` (a new `cog.outcome`/landing-decision
string, `"clamped"`) does the same for a `set_plan` landing — both are enum
additions on existing payload fields, not new event types. `collect_water`/
`bathe` leaving the villager-facing tool surface (roster/gloss only) touches
no event or payload shape at all — the sim executor keeps honoring both
verbs unchanged, so historical events of either kind replay exactly.

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `cog.thought` | `CogThoughtPayload{job, class, agent, snapshot_tick, generation, trigger_seq, points, predicted_wall_ms, predicted_land_tick, prompt_bytes?, block_bytes?, dropped_blocks?}` in `internal/sim/cognition.go` — the spec 043 tail (FR-009/FR-010) is additive, LAST, `omitempty`: assembled user-prompt size, per-block byte breakdown keyed by contract block name, and the blocks the size budget dropped in drop order; stamped ONLY on planner-class thoughts, zero-valued on every other class ([[decision-context]]) | mind driver (injected) when a call passes the router; `trigger_seq` is the log seq of the arming stimulus (0 = pure cadence) | none (telemetry, TASK-32, [[cognition]]) |
| `cog.outcome` | `CogOutcomePayload{job, class, agent, outcome, snapshot_tick, landing_tick, staleness_ticks, predicted_wall_ms, actual_wall_ms, kind?, reason?}` | loop landing ladder (landed/adapted/rejected-* /superseded/clamped) or mind driver (suppressed/expired/unusable — router suppressions have no matching `cog.thought`); since spec 061 (TASK-109) the mind driver's novelty SHIM also lands a `suppressed` outcome, distinguished from the router's by `reason` (`"nothing new since last exchange"`, `emitNothingNew`, [[social-fabric]]); since spec 058 (US2, FR-003) the landing ladder's `clamped` (`OutcomeClamped`) marks a `set_plan` whose oversized submission was truncated rather than rejected; the `retried` outcome is the one NON-terminal use — a marker for a consumed one-shot retry (conversation sites since TASK-42; the tool-loop's transport retry since spec 025, emitted by mind's `runPlan` and metatron's `Turn` alongside whatever terminal the run earns) | none — the terminal record of every thought (plus the non-terminal `retried` marker); rejections carry `kind` `prediction-miss` or `world-change` |
| `agent.intent_rejected` | `IntentRejectedPayload{agent, goal, reason, staleness_ticks}` in `internal/sim/cognition.go` | loop, when the landing ladder refuses a metered intent (alongside its `cog.outcome`) | since spec 043 US1 STATE-MUTATING (split out of the `cog.*` no-op arm): appends an already-closed `"rejected"` `IntentRecord` (source `planner`) to the ring — the refused intent never landed, so `Intent`/`IdleSince` stay untouched, but the next thought sees the attempt was made and refused; still its own type so the villagers tab/chronicle can notice refused intentions without parsing `cog.*` |
| `cog.recalibration_recommended` | `RecalibrationPayload{tier, estimate_s_per_pt, spike_rate, window}` in `internal/sim/cognition.go` | mind driver (injected) when a tier's live estimator breaches the spike-rate drift threshold (once per breach episode) | none (telemetry) |
| `cog.tool_call` (spec 017, FR-007) | `CogToolCallPayload{job, ordinal, tool, args?, verdict, reason?, tier, snapshot_tick}` in `internal/sim/cognition.go` | mind/metatron (injected), one per tool call a cognition's [[tool-loop]] saw — landed, landed_clamped (spec 058 FR-001/FR-003, `VerdictLandedClamped`, [[tool-loop]]), rejected, read, or unlanded; `{job, ordinal}` is the correlation key (1-based, dense per job, model-emission order) | none — recorded observability, reducer no-op, whitelisted alongside the other `cog.*` types |
| `cog.memory_divergence` (spec 042 US2) | `MemoryDivergencePayload{agent, tick, mode, legacy, augmented, overlap, displacement, vectorless, sit_tick}` in `internal/sim/cognition.go` | mind driver (injected via `InjectSocial`), one per memory selection while the world's `memory_relevance` flag is `shadow` or `on` — `legacy`/`augmented` are both windows' memory `Seq`s in window order ([[memory-retrieval]]) | none — recorded observability, reducer no-op, whitelisted alongside the other `cog.*` types |
| `agent.plan_set` | `PlanSetPayload{agent, job, steps}` in `internal/sim/plan.go` | loop, on a guarded plan landing (TASK-32 US4); since spec 058 (US2, FR-003) a submission longer than `PlanStepCap` is truncated to its first `PlanStepCap` steps IN PLACE before this event is built, so `steps` always carries exactly what landed, never the model's oversized submission (`OutcomeClamped` on the landing decision, [[sim-loop]]) | `Agent.Plan` replaced with the steps |
| `agent.plan_step_started` / `agent.plan_expired` | `PlanStepPayload{agent, job, step, reason?}` in `internal/sim/plan.go` | executor (`planStepEvents`) on an idle agent's head step firing / window closing or resolve failing | head step popped / whole plan cleared (a broken sequence is not resumed); spec 043: `plan_expired` also stamps the expired step into the `IntentLog` ring (`stampOrAppendExpired` — an open record matching the step's goal closes `"expired"`, otherwise a closed record is appended for a step that expired before ever firing) |

## Connections

The mind driver and the loop's landing ladder emit the `cog.*` telemetry
([[cognition]]); [[decision-context]] covers the spec 043 ring/anchor
surfaces the intent-lifecycle and needs rows maintain, and the prompt whose
size `cog.thought`'s `prompt_bytes` tail records; [[social-fabric]] owns the
mind-side novelty SHIM and its `cog.outcome{suppressed}` reason;
[[tool-loop]] owns `VerdictLandedClamped` end to end (spec 058);
[[tool-registry]] owns the `Clamp`-flagged `Param`s and the dormant-verb
roster prune the feature's tool surface reflects.
