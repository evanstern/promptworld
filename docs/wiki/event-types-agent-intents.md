---
name: event-types-agent-intents
description: Agent intent-lifecycle event rows split from [[event-types]]: agent.intent_set/work_started/intent_done/recovery_stalled/build_failed/moved. Load when tracing how an intent is set, worked, completed, aborted (build failure or a stalled needs-conditioned recovery hold), or how movement drives mental-map sighting.
kind: concept
sources:
  - internal/sim/agents.go
  - internal/sim/executor.go
  - internal/sim/landing.go
verified_against: 93837e1885bff17114df75e5382ac60dee24776a
---

# Event types — agent intent lifecycle

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.

Spec 062 (instinct yields to intelligence — [[reflex-policy]], TASK-103) adds
NO new event type: `Agent` gains `omitempty` `LastMindIntentDone` (the
reflex's PREP-gate yield-window anchor), so a pre-062 snapshot with the field
absent round-trips byte-identically. No emitter changes either — the EXISTING
`agent.intent_done` reducer arm gains a silent DERIVED write with no new
event (the `markExplored`/`notePresence`/`PairTalks` precedent): it arms
`LastMindIntentDone = e.Tick` whenever the intent it just closed had a
non-reflex `IntentRecord.Source` (`planner`/`plan`/`meeting`, via the new
`isMindSource` classifier); `agent.build_failed`'s `"failed"` closure never
arms it (only a genuine completion counts), and neither does a
reflex-sourced `"done"` closure, so a no-planner world's anchor stays 0
forever (FR-007, SC-003). `stampIntentOutcome` (`agents.go`) now returns the
closed record's `Source` and whether one was open, so the arm can read it
without a second ring scan.

Spec 064 (needs-conditioned recovery — [[executor]], [[reflex-policy]]) adds
NO format bump: `Intent` gains `omitempty` `UntilNeed`/`UntilValue`/`HoldRef`
(a completion condition and its hold anchor — absent on every pre-064 intent,
so a conditionless intent marshals byte-identically), `IntentSetPayload`
gains `omitempty` `UntilNeed`/`UntilValue` (carried onto the intent only when
`UntilNeed` names a valid closed-set need — `warmth`/`rest`/`food` — a
malformed or absent need leaves the intent conditionless), and
`WorkStartedPayload` gains `omitempty` `Ref` (the need level captured at a
conditioned hold's anchor tick; 0 and unread for every ordinary work goal).
ONE new event type, `agent.recovery_stalled` (`RecoveryStalledPayload{agent,
goal, need}`), is the executor emission class (no whitelist entry, the
`metatron.order_expired`/`curriculum.*` pattern): a needs-conditioned hold
whose need shows no net gain across a full `recoveryStallTicks` window
(dead fire, displaced source, unreachable threshold) aborts with this
DISTINCT outcome — mirroring `agent.build_failed`'s state effect (`Intent`
cleared, `IdleSince` stamped) but stamping the intent ring `"stalled"`
(joining `done`/`failed`/`rejected`/`expired`) rather than a completion, so
it never arms the spec-062 yield window (the `build_failed` precedent: an
abort is not intelligence completing).

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `agent.intent_set` | `IntentSetPayload{agent, goal, target, res, source, kind?, qty?, job?, reason?, until_need?, until_value?}` | reflex (grace-gated), planner injection, or a plan step firing | intent installed; `source` (`reflex`/`planner`/`plan`/`meeting`) says which mind chose it; also stamps `Agent.LastGoal`/`LastGoalTick` (spec 015 — never cleared by any event, the villagers tab's past-objective line, [[tui-client]]); `job` (spec 017, omitempty) is set ONLY at the `inject_intent` landing site from `InjectArgs.JobID` — a planner-loop landing carries its job id, reflex/executor-authored intents carry none; `reason` (spec 019, omitempty) is likewise set ONLY at that landing site from `InjectArgs.Reason` — the planner's free-text reason, copied onto `Intent.Reason` by the reducer so it survives to completion where the executor bakes it into a memory's `why`; `until_need`/`until_value` (spec 064, omitempty, now the LAST fields) carry the optional needs-conditioned completion — set by the warm_up resolver and the reflex's conditioned warmth rungs — onto `Intent.UntilNeed`/`UntilValue`, but ONLY when `until_need` names a valid closed-set need (`warmth`/`rest`/`food`); every `omitempty` tail stays empty on reflex/executor emissions carrying none, so those marshal byte-identically to pre-feature; spec 043 US1: also appends an `IntentRecord{goal, source, reason, tick}` to `Agent.IntentLog` (ring, cap 8) — a previous still-open record stays open, so an override reads as open-then-new ([[decision-context]]) |
| `agent.work_started` | `WorkStartedPayload{agent, tick, ref?}` | executor at target | `WorkStart` stamped; since spec 064 (omitempty) `ref` also stamps `Intent.HoldRef` — a needs-conditioned hold's anchor need level, inert (0, unread) for every ordinary work goal |
| `agent.intent_done` | `AgentPayload{agent}` | executor (done/invalid/unreachable) | intent cleared — but since spec 038, a **build** goal (`build_fire`/`build_shelter`/`build_oven`/`build_chest`/`build_path`/`build_wall_plank`/`build_wall_stone`) whose mid-work re-validation fails no longer funnels through here; it emits the distinct `agent.build_failed` below instead, and since spec 064 a needs-conditioned recovery hold with a dead source aborts via the distinct `agent.recovery_stalled` below instead. `intent_done` remains the resolution for successful non-build completion paths (including a satisfied or threshold-crossed recovery hold), non-build no-ops (craft/cook/bathe/deposit contested re-checks), and every non-build goal's invalid/unreachable exit; spec 043 US1: also closes the newest still-open `IntentLog` record `"done"`; spec 062: when the closed record's `Source` `isMindSource` (`planner`/`plan`/`meeting`), also arms `Agent.LastMindIntentDone = e.Tick` — the reflex PREP gate's yield-window anchor ([[reflex-policy]]); a reflex-sourced closure never arms it (a reflex-issued recovery's completion is exempt on the same rule) |
| `agent.recovery_stalled` (spec 064) | `RecoveryStalledPayload{agent, goal, need}` | executor, a needs-conditioned hold ([[executor]]'s `recoveryHoldEvents`) whose need shows no net gain across a full `recoveryStallTicks` (300) window while holding at its target — dead fire, displaced source, or an unreachable threshold | intent cleared (`Intent = nil`, `IdleSince` stamped) — identical to `agent.intent_done`'s state effect, but the DISTINCT, honest abort outcome: closes the newest still-open `IntentLog` record `"stalled"` (never `"done"`) and — the `agent.build_failed` precedent — never arms `Agent.LastMindIntentDone` (an abort is not intelligence completing) |
| `agent.build_failed` (spec 038) | `BuildFailedPayload{agent, goal, reason}` | executor, mid-work re-validation of a build goal (the seven `build_*` above) — emitted **instead of** a bare `agent.intent_done` when the build genuinely fails: `reason` is `site no longer buildable` (any build goal whose `buildSite` re-check fails) or `site blocked too long` (walls only, once a reserved-tile occupant outlasts `wallOccupancyGraceTicks` past the due tick — [[executor]]) | intent cleared (`Intent = nil`, `IdleSince` stamped) — identical to `agent.intent_done`, so no material spent and no structure stands; a paired same-tick `agent.memory_added` (`OriginAction`, `salShelter`) rides along stating the build did NOT complete and why, so the builder can falsify a phantom-structure belief; the builder's mind re-arms its planner exactly as for `agent.intent_done` ([[agent-mind]]). Distinct from `agent.intent_rejected` (up-front landing refusal) — this clears an accepted intent; spec 043 US1: also closes the newest still-open `IntentLog` record `"failed"`. TUI digest renders it as a failure line (builder, goal, reason), never "finished" ([[tui-client]]) |

## Connections

[[executor]] and [[reflex-policy]] jointly own the spec 064
needs-conditioned recovery surface end to end — the `intent_set`/
`work_started` field additions and the new `agent.recovery_stalled` type
above;

[[governance]] excludes a recovery-held villager from the
emergent-gathering quorum, a reducer-side read of `Intent.UntilNeed` with no
event of its own.
