---
name: sim-loop-landing-ladder
description: How a metered InjectIntent lands — Loop.Do/InjectIntent/InjectArgs, the six-rung landing ladder (unavailable/superseded/stale/pair-cooldown/guard-failed/adapted/success), and the resulting cog.outcome verdict. Load when tracing why a planner intent was accepted, rejected, adapted, or clamped.
kind: component
sources:
  - internal/sim/loop.go
  - internal/sim/landing.go
verified_against: 93837e1885bff17114df75e5382ac60dee24776a
---

# Sim loop — landing ladder

Child of [[sim-loop]] — the door a planner-issued intent lands through
(`Loop.InjectIntent`) and the six-rung ladder that judges it against the
world as it now stands.

## How it works

`Loop.Do(name, speed)` is the thread-safe entry used by IPC sessions — and, since
spec 029 (US5, [[guardian-orders]]), by the Guardian's `pause`/`start`/`adjust_speed`
meta tools through a `LoopControl` seam the daemon wires onto the same `*Loop`
([[daemon-lifecycle]]): an angel-issued clock command lands `clock.paused`/
`clock.resumed`/`clock.speed_set` indistinguishably from a console one. It
fails cleanly via the loop's `done` channel if the loop has stopped. `Loop.DoState()` answers the
protocol's `state` command with the canonical `State` JSON plus a status captured in
the same loop iteration — the returned `last_seq` is exactly the log position the
state reflects, which is what makes client-side replicas gapless.
`Loop.InjectIntent` (the `inject_intent` command) is the door for planner output
([[agent-mind]]). `InjectArgs` now carries cognition-horizon landing metadata
([[cognition]]): `Class`, `JobID`, `SnapshotTick`, `Generation`,
`PredictedWallMs`/`ActualWallMs`, `Guards`, and an optional `Plan` (mutually
exclusive with `Goal`), plus `Kind`/`Qty` (spec 013 R4) arguing a storage goal
(`drop`/`pick_up`/`deposit`/`withdraw`) when `Goal` is one of them — additive and
ignored otherwise, so pre-013 callers leave them zero. An empty `Class` means an
unmetered caller (tests, tooling): the ladder below is skipped and no telemetry
is emitted — the pre-TASK-32 contract.

At the boundary, a metered intent climbs the **landing ladder** against the
world as it is now (`staleness = state.Tick − SnapshotTick`, floored at 0).
Since TASK-70 the ladder lives in `internal/sim/landing.go` as
`Loop.landIntent` — the `inject_intent` case in `handleCommand` is a one-line
dispatch to it — with each doctrine rung a named function (`rungUnavailable`,
`rungSuperseded`, `rungStale`, `rungPairCooldown`, `rungHailRelaxed`,
`rungGuardFailed`, `rungAdapted`, `rungInRadiusHail`) and the guard walk
(`walkGuards`) producing
one explicit `landingDecision` (outcome, reason, hail target) instead of the
former cross-loop flags. The extraction is behavior-identical; the rungs:

1. dead/asleep agent → `rejected-unavailable`;
2. `Generation` mismatch with `Agent.Generation` → `superseded`;
3. staleness over the class's budget → `rejected-stale`. Since spec 067
   (TASK-141) the budget is `EffectiveBudgetTicks(tps)` — the class's 1x
   `BudgetTicks` (via `cognition.ClassFor`) scaled by the reducer's own
   event-sourced effective speed at the landing tick
   (`state.Speed.TicksPerSecond()`; unscaled at uncapped speed), keeping the
   gate a pure function of event-sourced state so replay determinism holds
   while a constant-wall-time thought is judged the same at every capped
   speed. The reason names the derivation (`staleness 9800 > budget 9600
   (1200 at 1x × 8x)`);
4. since spec 061 (TASK-109, [[social-fabric]]): `rungPairCooldown` — a
   `talk_to` landing whose living target the actor spoke with inside
   `EncounterCooldown()` (the [[world-tuning]] spec-048 dial, [[sim-state-reducer]]'s
   `State.PairLastTalk`/`pairCooled`) is refused `rejected-guard` with an
   informative "spoke recently" reason BEFORE the hail rungs — the planner
   founds no scene and learns why instead of burning turns re-hailing
   (FR-002). Non-`talk_to` goals, a dead target (left to the guard walk's own
   rejection), a never-talked pair, and a past-cooldown pair are all vacuous
   (`""`, no gate); the SAME predicate also backstops hail PLACEMENT
   (`hailable`, now `tick`-parameterized) and hail FOUNDING (`hailStep`'s
   sweep) so all three deliberate-talk routes close through one gate (SC-002)
   — `hailable`'s existing exemptions (dead/asleep/already-hailed/deadlock/
   meeting-pinned/radius) are unchanged, this is an ADDITIONAL check;
5. any `Guard.Eval` failure → `rejected-guard` — EXCEPT the hail rung
   (TASK-47): a failed `target_present` on a `talk_to` landing whose living
   target is `hailable` (or is the actor's own hailer — mutual convergence)
   proceeds as **adapted** instead of rejecting; a `target_present` guard that
   holds but whose target moved likewise marks the landing **adapted** (the
   repair is `resolveGoal`'s re-resolution);
6. success: the goal must first be a World tool on the [[tool-registry]]'s
   villager roster (spec 014 US3 — an out-of-roster or unknown name rejects
   with the same `unknown goal` reason as before; real planner traffic is
   unaffected), then `resolveGoal` resolves coordinates deterministically,
   recorded as `agent.intent_set (source: planner, job: InjectArgs.JobID)` +
   `agent.thought` (since spec 017 the tool-use loop's job id threads onto the
   landed event's `Job` field at this single emission site), or — for a
   `Plan` — each step validated against `tool.PlanStepGoals()`,
   the registry-derived plan-step set (spec 014 FR-006; deriving it cured the
   TASK-55 drift where the old hand-maintained `planGoals` map silently
   rejected the nine spec-012 verbs — FR-012, the migration's sole behavioral
   delta; missing `Until` defaults to `state.Tick + PlanDefaultWindowTicks`).
   Since spec 058 (US2, FR-003), a `Plan` longer than `PlanStepCap` is no
   longer rejected whole: it is truncated to the first `PlanStepCap` steps
   IN PLACE, reducer-side, before per-step validation, and the decision's
   outcome becomes `OutcomeClamped` — so the landed `agent.plan_set` always
   carries exactly the steps that were accepted (deterministic, replay-safe)
   and the model-facing/telemetry trail can tell a clamped plan from a clean
   one; a structurally invalid step WITHIN the clamped window still rejects
   the whole landing exactly as before — the clamp only forgives length. Both
   shapes are recorded as `agent.plan_set`; a `resolveGoal` failure is itself
   `rejected-guard`. Since spec 019 (R2), a non-empty `InjectArgs.Reason` also
   rides onto the landed `agent.intent_set` event's `Reason` field (reflex-
   and executor-authored intent_set events carry none), so the planner's
   free-text reason survives to completion as recorded input rather than a
   second event. Since spec 064 R3, a resolved `warm_up` (or any other
   completion-conditioned resolution) also carries its `UntilNeed`/
   `UntilValue` onto the landed `agent.intent_set` — zero for every
   conditionless goal, unchanged. A
   successful `talk_to` landing with a `hailable` target additionally emits
   `social.hailed` (in- or out-of-radius — the courtesy pause is uniform;
   [[executor]] enforces it and resolves met/expiry).

Every metered verdict lands atomically as `cog.outcome` (rejections also emit
`agent.intent_rejected`), classified `prediction-miss` when
`ActualWallMs > PredictionMissFactor × PredictedWallMs` and `world-change`
otherwise — see [[event-types]] for payload shapes.

Both injection doors are deliberately pause-open (FR-018): pause means "the
world freezes and the minds catch up" — an in-flight thought completes on the
wall clock and lands at the frozen tick, where its game-tick staleness is zero
by construction.

## Connections

Parent note: [[sim-loop]]. [[cognition]] (`cognition.ClassFor`) supplies the
landing budgets and classes the ladder judges against, and its router/
estimators produce the snapshot/landing metadata this ladder reads.
[[social-fabric]] owns the mind-side novelty shim one layer above
`rungPairCooldown`, and [[sim-state-reducer]] owns the `PairTalks` ledger
both read. [[tool-registry]] owns `set_plan`'s schema (no longer declaring
`maxItems`) and the villager-roster/GOAL-DOOR discriminator a landing's
success rung checks against. [[tool-loop]] owns the matching
`VerdictLandedClamped` expressive text for the plan rung's `OutcomeClamped`
path. [[event-types]] catalogs the `cog.outcome`/`agent.intent_rejected`
payload shapes.
