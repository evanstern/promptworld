---
name: planner-telemetry
description: Per-call tool telemetry (cog.tool_call on every termination path), the single-goal InjectIntent landing path, and runPlan's terminal switch mirroring the pre-loop outcome paths (landed/rearmed/unusable). Split from [[agent-mind]] (Tool-call telemetry section).
kind: component
sources:
  - internal/mind/mind.go
  - internal/mind/telemetry.go
verified_against: ebc30ee1990dbcdf17576a2fbe81fdbeb175411d
---

# Planner outcome telemetry

**Tool-call telemetry** (`telemetry.go`, spec 017 FR-007/T018): every buffered
`CallRecord` the loop's `Record` sink collected lands as a `cog.tool_call`
event (`emitToolCalls`/`toolCallEvent`, via the shared `sim.NewCogToolCallPayload`
constructor also used by [[guardian]]) on EVERY termination path — landed,
rejected, capped, or errored — so a call that never grounded is still queryable
from the log (AC#5). Events are sorted by `Ordinal` before emission (the driver
already buffers them ordinal-dense; sorting here makes the mind's emission
order-independent of buffer order) and ride ONE dedicated `InjectSocial` batch,
separate from the terminal `cog.outcome`. A verdict requiring a non-empty
reason (every `rejected_*`, `read_error`, and — since spec 058 SC-005 —
`landed_clamped`, whose Reason is the queryable clamp notice: without it a
clamped acceptance would have no query value distinguishing it from a clean
`landed`) gets one backfilled from the
verdict name if a handler somehow left it blank, logged as the contract
violation it would be (`verdictRequiresReason`). Since spec 025 (TASK-72)
`runPlan` also surfaces the loop's one in-loop transport retry: when
`res.Retried` is set ([[tool-loop]]'s one-per-run transport retry), it emits a
NON-terminal `cog.outcome` carrying `sim.OutcomeRetried` and the first
failure's reason via `cogOutcomeEvent` — the TASK-42 marker vocabulary, so no
new event type — making every recovery countable from the trail; the terminal
outcome the run earns is still owned by the landing door or the terminal
switch below (the [[tui-client]] decision-trace projection skips the marker so
it never overwrites the earned terminal).

Single goals land via `Loop.InjectIntent` exactly as before — which validates,
resolves coordinates deterministically at the tick boundary (`resolveGoal`),
and records `agent.intent_set (source: planner)` + `agent.thought`, carrying
the landing metadata (`sim.InjectArgs`: Class, JobID, SnapshotTick,
Generation, Predicted/ActualWallMs, and since spec 013 Kind/Qty) and, for
`talk_to`, `GuardTargetAlive` + `GuardTargetPresent` guards — the loop (now
[[tool-loop]]'s `Run`, not `runPlan` itself) owns the round cardinality;
`Loop.InjectIntent`'s landing ladder still owns the landing verdict and its
outcome telemetry, unchanged. `runPlan`'s terminal switch, once the loop
returns, mirrors the pre-loop paths on `res.Term`: a `TermLanded` loop leaves
the sole `cog.outcome` to whichever door landed it (no rearm); a loop that
ended with `d.doorOutcome` true but nothing landed (a rejection, mirroring the
old rejection path) calls `rearmAgent` — the agent noticed the plan failed and
re-thinks at the next open debounce window, promptly but never hotly — with no
outcome added (the door's rejection is the record); and a loop that reached no
door at all (plain text, reads only, an unknown `talk_to` target, an infra
error — `TermModelDone`/`TermCapExhausted`/`TermAdmissionRefused`/`TermCtxDone`)
emits the terminal `cog.outcome{unusable}` itself, with `loopFailReason(res,
err)` naming which termination caused it (FR-015: no failure is silent) — the
reflex grace (120 ticks idle) remains the floor under every gap, and the
permanent degraded mode. Since spec 106 two sleep-gate records join this
vocabulary without any new event type or outcome word: a job skipped at
dequeue (agent asleep/dead per the mind's mirror — [[tool-use-dispatch]])
terminates in one `cog.outcome{suppressed}` with a dequeue reason and NO
matching `cog.thought` (the router-suppression shape), and a `TermCtxDone`
whose context absorb cancelled on `agent.slept`/`agent.died` swaps
`loopFailReason`'s generic "loop: context ended" for the slot's recorded
attribution ("cancelled in flight: agent slept"/"… agent died") — so a
sleep-cancel is countable from the trail, distinct from a plain
`callTimeout`, and the soak arithmetic (spec 106 SC-003) derives from the
log alone. A cancelled run also never feeds the estimator or the transport
retry: [[tool-loop]]'s ctx_done family observes nothing and never retries. The daemon also installs `RecalibrateSignal` as the
orchestrator's drift hook: an estimator spike-rate breach — which since spec
031 also adopts the window median as the new estimate — lands as
`cog.recalibration_recommended`, the payload carrying the adoption arithmetic
in additive `prior_s_per_pt`/`adopted_s_per_pt` fields.

## Connections

[[agent-mind]] is the parent note this child was split from; [[tool-use-dispatch]]
is the loop driver whose termination this telemetry records;
[[villager-tool-handlers]] is the sibling child whose handlers set
`doorOutcome` and hand back a verdict this switch consumes; [[guardian]]
shares the `sim.NewCogToolCallPayload` constructor; [[cognition]] owns the
latency estimate `RecalibrateSignal` recalibrates against; [[llm-orchestrator]]
is the per-provider `RecalibrateSignal` hook installs into; [[event-types]]
catalogs `cog.tool_call`/`cog.outcome`/`cog.recalibration_recommended`.

