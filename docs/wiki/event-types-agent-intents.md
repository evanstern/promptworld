---
name: event-types-agent-intents
description: Agent intent-lifecycle event rows split from [[event-types]]: agent.intent_set/work_started/intent_done/recovery_stalled/build_failed/intent_failed/moved. Load when tracing how an intent is set, worked, completed, aborted (build failure, the generalized non-build intent failure, or a stalled needs-conditioned recovery hold), or how movement drives mental-map sighting.
kind: concept
sources:
  - internal/sim/agents.go
  - internal/sim/executor.go
  - internal/sim/landing.go
verified_against: 376afd4cee54839a545bc88409f3c485c2f5149d
---

# Event types — agent intent lifecycle

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.

Spec 086 (agent-named payloads): every agent-referencing field in this
family's payloads is a `sim.AgentRef` — the wire carries
`{"id":N,"name":"…"}` objects (lists element-wise), the name stamped at
emission from the fixed roster via `Ref`/`Refs`; sentinels marshal
`{"id":-1,"name":""}`. Legacy bare-int rows decode through the dual-shape
unmarshal forever and reducer arms fold `.ID`s only — the conventions and
the normative back-compat matrix live on [[event-types]].
Spec 062 (instinct yields to intelligence — [[reflex-policy]], TASK-103) adds
NO new event type: `Agent` gains `omitempty` `LastMindIntentDone` (the
reflex's PREP-gate yield-window anchor), so a pre-062 snapshot with the
field absent round-trips byte-identically. No emitter changes either — the
EXISTING `agent.intent_done` reducer arm gains a silent DERIVED write with
no new event (the `markExplored`/`notePresence`/`PairTalks` precedent): it
arms `LastMindIntentDone = e.Tick` whenever the closed intent had a
non-reflex `IntentRecord.Source` (`planner`/`plan`/`meeting`, via the new
`isMindSource` classifier); `agent.build_failed`'s `"failed"` closure never
arms it, nor does a reflex-sourced `"done"` closure, so a no-planner
world's anchor stays 0 forever (FR-007, SC-003). `stampIntentOutcome`
(`agents.go`) now returns the closed record's `Source` and whether one was
open, so the arm reads it without a second ring scan.

Spec 064 (needs-conditioned recovery — [[executor]], [[reflex-policy]]) adds
NO format bump: `Intent` gains `omitempty` `UntilNeed`/`UntilValue`/`HoldRef`
(a completion condition and its hold anchor — absent on every pre-064
intent, so a conditionless intent marshals byte-identically),
`IntentSetPayload` gains `omitempty` `UntilNeed`/`UntilValue` (carried onto
the intent only when `UntilNeed` names a valid closed-set need —
`warmth`/`rest`/`food` — a malformed or absent need leaves the intent
conditionless), and `WorkStartedPayload` gains `omitempty` `Ref` (the need
level at a conditioned hold's anchor tick; 0 and unread for an ordinary
work goal). ONE new event type, `agent.recovery_stalled`
(`RecoveryStalledPayload{agent, goal, need}`), is the executor emission
class (no whitelist entry, the `guardian.order_expired`/`curriculum.*`
pattern): a needs-conditioned hold whose need shows no net gain across a
full `recoveryStallTicks` window (dead fire, displaced source, unreachable
threshold) aborts with this DISTINCT outcome — mirroring
`agent.build_failed`'s state effect (`Intent` cleared, `IdleSince` stamped)
but stamping the ring `"stalled"` (joining `done`/`failed`/`rejected`/
`expired`) rather than a completion, so it never arms the spec-062 yield
window (an abort is not intelligence completing).

Spec 096 (TASK-95) generalizes `agent.build_failed` to every non-build goal
— ONE new type, `agent.intent_failed`, row below. Additive, no format bump
(spec 094: bumps trigger on renames, not new types).

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `agent.intent_set` | `IntentSetPayload{agent, goal, target, res, source, kind?, qty?, job?, reason?, until_need?, until_value?}` | reflex (grace-gated), planner injection, or a plan step firing | intent installed; `source` (`reflex`/`planner`/`plan`/`meeting`) says which mind chose it; stamps `Agent.LastGoal`/`LastGoalTick` (spec 015 — never cleared by any event, the villagers tab's past-objective line, [[tui-client]]); `job` (spec 017, omitempty) is set ONLY at the `inject_intent` landing site from `InjectArgs.JobID` — a planner-loop landing carries its job id, reflex/executor-authored intents carry none; `reason` (spec 019, omitempty) is likewise set ONLY at that landing site from `InjectArgs.Reason` — the planner's free-text reason, copied onto `Intent.Reason` by the reducer so it survives to completion where the executor bakes it into a memory's `why`; `until_need`/`until_value` (spec 064, omitempty, the LAST fields) carry the optional needs-conditioned completion — set by the warm_up resolver and the reflex's conditioned warmth rungs — onto `Intent.UntilNeed`/`UntilValue`, only when `until_need` names a valid closed-set need (`warmth`/`rest`/`food`); every `omitempty` tail stays empty on reflex/executor emissions carrying none, so those marshal byte-identically to pre-feature; spec 043 US1: appends an `IntentRecord{goal, source, reason, tick}` to `Agent.IntentLog` (ring, cap 8) — a previous still-open record stays open, so an override reads as open-then-new ([[decision-context]]) |
| `agent.work_started` | `WorkStartedPayload{agent, tick, ref?}` | executor at target | `WorkStart` stamped; since spec 064 (omitempty) `ref` also stamps `Intent.HoldRef` — a needs-conditioned hold's anchor need level, inert (0, unread) for an ordinary work goal |
| `agent.intent_done` | `AgentPayload{agent}` | executor (successful completion, or an instant/wander-class goal with no re-validation) | intent cleared — build failures (038) use `agent.build_failed` below, dead-source recovery holds (064) use `agent.recovery_stalled` below, every OTHER goal's invalid/contested exit (096) uses `agent.intent_failed` below; `intent_done` no longer covers any of the three; spec 043 US1: closes the newest open `IntentLog` record `"done"`; spec 062: `isMindSource` (`planner`/`plan`/`meeting`) also arms `Agent.LastMindIntentDone = e.Tick` ([[reflex-policy]]'s yield anchor); reflex-sourced closures never arm it |
| `agent.recovery_stalled` (spec 064) | `RecoveryStalledPayload{agent, goal, need}` | executor, a needs-conditioned hold ([[executor]]'s `recoveryHoldEvents`) whose need shows no net gain across a full `recoveryStallTicks` (300) window at its target — dead fire, displaced source, or unreachable threshold | intent cleared, identical to `intent_done`'s effect, but the DISTINCT honest-abort outcome: closes the `IntentLog` record `"stalled"` (never `"done"`) and — the `build_failed` precedent — never arms `LastMindIntentDone` (an abort isn't intelligence completing) |
| `agent.build_failed` (spec 038) | `BuildFailedPayload{agent, goal, reason}` | executor, mid-work re-validation of a build goal (the seven `build_*` above) — instead of a bare `agent.intent_done`: `reason` is `site no longer buildable` (`buildSite` re-check fails) or `site blocked too long` (walls only, a reserved-tile occupant outlasting `wallOccupancyGraceTicks` — [[executor]]) | intent cleared (`Intent = nil`, `IdleSince` stamped), identical to `intent_done` — no material spent, no structure stands; a paired same-tick `agent.memory_added` (`OriginAction`, `salShelter`) states the build did NOT complete and why; mind re-arms as for `intent_done` ([[agent-mind]]). Distinct from `agent.intent_rejected` (up-front refusal); closes the `IntentLog` record `"failed"`; TUI digest renders a failure line, never "finished" ([[tui-client]]) |
| `agent.intent_failed` (spec 096) | `IntentFailedPayload{agent, goal, reason, x, y}` | executor, generalizing `agent.build_failed` — the mid-work `valid` exit (`forage`/`chop`/`hunt`/`demolish`/`repair`/`quarry`/`cook`/`bathe`, `"target gone"`) and the completion-time no-op recheck (`craft_*`/`cook`/`bathe`/`deposit`/`withdraw`, `"contested"`, or `"invalid"` for `deposit`'s empty `Kind`) | intent cleared, identical to `build_failed`'s effect; paired memory (`salIntentFailed` = `salShelter`); mind re-arms the same way; `x`/`y` are the actor's own stand tile (no shared Target-vs-Res convention here); TUI digest renders a failure line |

## Connections

[[executor]]/[[reflex-policy]] jointly own the spec 064 needs-conditioned
recovery surface end to end — the `intent_set`/`work_started` field
additions and the new `agent.recovery_stalled` type above; [[governance]]
excludes a recovery-held villager from the emergent-gathering quorum, a
reducer-side read of `Intent.UntilNeed` with no event of its own.
