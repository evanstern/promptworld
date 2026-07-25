---
name: sim-loop
description: The single-goroutine fixed-timestep loop — tick execution, command intents at tick boundaries, pacing, auto-slow degradation
kind: component
sources:
  - internal/sim/loop.go
  - internal/sim/landing.go
verified_against: e137b82bb699eb323eb26c6a69c3dc83ca474b27
---

# Sim loop

`sim.Loop` is the one goroutine that owns `State` and the write path to the store,
holding the static terrain (`worldmap.Map`, via `NewLoop(state, m, store, notify)`)
as read-only context for tick generation.
Everything external — pause, resume, speed changes, status reads — enters through a
command channel and is applied at a tick boundary, with every applied command recorded
as an event. That makes the [[event-log]] the complete input record of a run.

## How it works

`Loop.Run(ctx)` is a state machine over four modes:

- **Ended** (spec 044, checked first): once `State.Ended` latches, the terminal
  paused-mode posture — no timer, ever again; block on commands or ctx (ctx
  cancellation still takes the final snapshot). Unlike Paused there is no
  pacing to restart, because no command can resume an ended world (below).
- **Paused**: no timer; block on commands or ctx. Resume restarts pacing fresh.
- **Timed** (interval > 0): a timer fires per `Speed.Interval()`; each firing runs one
  tick and advances the schedule by exactly one interval. If the loop falls more than
  one interval behind, the schedule resets to now — **no catch-up bursts**; the world
  slows honestly instead of skipping (FR-012).
- **Max speed** (interval 0): spin ticks back-to-back with a non-blocking command
  check and a `runtime.Gosched()` every 1024 ticks.

`runTick`: compute `stepEvents(state, map, nextTick)` (pure), advance `state.Tick`,
pre-assign the batch's store seqs (`stampSeqs`, spec 042 — below), apply
each event through the reducer, `AppendEvents` in one transaction, then `notify`
(the [[ipc-server]] broadcast — must never block). Every `SnapshotEveryTicks = 3600`
ticks it snapshots and prunes.

`Loop.stampSeqs(events)` (spec 042) pre-computes each event's eventual store
seq as `LastSeq() + i + 1` and stamps it onto the batch BEFORE the reducer
applies it: `AppendEvents` otherwise only assigns seqs inside its own append
transaction, which runs AFTER `Apply` in `runTick`, `handleCommand`, and
`observeWindow` (all three call it on their event batch right before the
apply loop) — too late for the `agent.memory_added` arm's `Memory.Seq` stamp
([[sim-state-reducer]], [[memory-retrieval]]) to see a real value live. Since
the loop is the log's single writer while running, `AppendEvents` then
re-assigns the identical `last+i+1` values, so live state and a replayed
state (which reads each event's seq straight off the log) always agree — the
invariant `CheckContiguity` guards.

`handleCommand` implements idempotent semantics: pausing a paused world emits nothing;
`set_speed` to the current speed emits nothing; otherwise the `clock.*` event is
applied, appended, and broadcast, and a pause also triggers an immediate snapshot.
Replies carry a coherent `Status` snapshot (tick, game time, flags, last seq).
Emitted events now land regardless of the command's error — a rejected
`inject_intent` is the only command that pairs an error with events (its rejection
telemetry, so no failure is silent); every other error path emits nothing.

`handleCommand` also opens with the **ended gate** (spec 044 FR-002/FR-003): a
finished world (`State.Ended`) refuses every clock/world-mutating command —
`pause`, `resume`, `set_speed`, `govern`, `inject_intent` — with an explicit
"run has ended" error; reads (`status`/`state`) serve unchanged, and
`inject_social` narrows to `endedProseWhitelist` — recorded prose ABOUT the
ended run, exactly two types: `chronicle.entry` and `morgue.epilogue` (the
run-end epilogue lands AFTER `run.ended` by construction, the narrator
mourning asynchronously — [[morgue]]); any other social type in the batch
refuses the whole batch. `inject_operator` stays open (its whitelisted types
are reducer no-ops — daemon lifecycle, never world state — the same reason
shutdown is unaffected). The protocol `Status` gains additive `omitempty`
`Ended`/`EndedDay` fields (the run-over posture and its game day, for
rendering without a state fetch), and an ended world reports an effective
rate of 0 like a paused one.

Auto-slow (`observeWindow`): every `degradeWindow = 5s` the loop compares achieved
ticks/sec against the requested rate; sustained shortfall below 90% emits
`clock.degraded` (with the measured rate), recovery to ≥95% emits `clock.recovered`.
At max speed whatever is achieved is the contract — no degradation events.

`Loop.Govern(to, debt, jobs)` (spec 028 US2/US3) is the daemon governor sampler's
door onto the same command channel `set_speed` uses: it lands a
`clock.governor_shed`/`clock.governor_recovered` event exactly like a player
speed change, re-validating at the tick boundary before applying. A decision
that no longer applies by the time it lands — the world paused, `Speed`
already moved, `to` off the capped ladder, not exactly one notch from the
current speed, or a recover above the standing `RequestedSpeed` ceiling — is
dropped silently (no event, clean return); the daemon's sampler simply
re-evaluates next cadence, so there is never a merge to resolve. `Speed`
itself keeps meaning "the speed the loop paces at" — since spec 028 that is
specifically the EFFECTIVE speed, with `RequestedSpeed` carrying the player's
ceiling only while governed ([[cognition]], [[sim-state-reducer]]).

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
3. staleness over the class's `BudgetTicks` (looked up via
   `cognition.ClassFor`) → `rejected-stale`;
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
   second event. A
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

`Loop.InjectSocial` is the second door — the mind's injection
door ([[social-fabric]], [[nightly-consolidation]], musings per [[agent-mind]],
narrator entries per [[chronicle]], nudges and miracles per [[guardian]] /
[[guardian-miracles]], standing orders per [[guardian-orders]], proposal rephrasing
per [[governance]], place-knowledge per [[mental-maps]] — `agent.thought` is
whitelisted as a reducer no-op, `chronicle.entry` appends the story ring,
`metatron.nudged` spends a charge with a validating reducer the dry-run enforces,
`metatron.place_revealed` (spec 041, FR-014) widens the boundary by one — a
vision's optional place grant, declared in `send_vision`'s `Events` so
`ValidateToolCoverage` pins it ⊆ this whitelist, whose dry-run enforces a
living target and a real place before anything lands — the four
`metatron.time_snapped`/`metatron.item_granted`/`metatron.entity_moved`/
`metatron.entity_removed` miracle types (spec 016) are whitelisted the same way —
their reducer arms enforce presence/destination/charge before anything lands,
the whitelist is only the isolation boundary — `metatron.order_placed`/
`metatron.order_cancelled`/`metatron.order_triggered` (spec 029) join the
whitelist the same way (placement/cancellation/trigger-match validation lives
in the reducer arm); `metatron.order_expired` needs no whitelist entry — it is
executor-emitted, never injected, the `charge_regenerated` precedent —
(since spec 036 whitelist membership is also readable from outside the package
via `InjectableSocialEvent(t)`, the single-source accessor both the tool
coverage gate and the bundle boot gate ([[bundle-tools]]) enforce against) —
`meeting.proposal_rephrased` swaps
an enacted norm's text and nothing else,
the `cog.*` telemetry — `cog.thought`, `cog.outcome`,
`cog.recalibration_recommended`, and (since spec 017) `cog.tool_call` (the
tool-use loop's per-call trace, [[tool-loop]]) — is whitelisted as reducer
no-ops so the [[cognition]] layer's observability is recorded, never silent,
and (since spec 019, US3) `journal.entry_written`/`journal.entry_deleted` —
the two mind-injectable journal mutations, whose reducer dry-run enforces the
rune budget (written) and entry existence (deleted) before either lands, and
(since spec 030 US2, FR-008) `agent.belief_reinforced` — the
grounded-observation seam that re-anchors a held belief's decay clock; spec 030
ships the whitelist entry and reducer arm only, no in-tree emitter yet), and
(since spec 042 US1/US2) three more: `agent.memory_embedded`/
`agent.situation_embedded` — the mind-side embedder's two vector companions
([[memory-retrieval]]), state-mutating unlike the `cog.*` telemetry (below) —
door ordering guarantees a memory's embedding companion never precedes the
memory itself, since the embedder only observes an `agent.memory_added` AFTER
it is committed and notified; and `cog.memory_divergence`, the shadow-mode
selector's rank-divergence record, riding the same reducer-no-op `cog.*`
isolation class as the telemetry types below), and (since spec 044 US2) two
more: `metatron.charter_observed` — the Guardian turn pipeline's
fingerprint-at-effect stamp, the event-sourced charter-revision timeline the
[[morgue]] aligns deaths against, whose reducer arm (and so the dry-run)
enforces a non-empty fingerprint — and `morgue.epilogue`, the narrator's
recorded mourning prose after a death or the run's end, appending only the
bounded `State.MorgueEpilogues` ring (never simulation state, which is why it
also survives the ended-world narrowing above)):
an atomic, whitelisted batch of conversation, consolidation, musing, chronicle,
nudge, miracle, phrasing, or telemetry effects, dry-run on a state copy before
applying — the dry-run probe is reconstructed from bytes and so carries no
unexported/unserialized state, so `handleCommand` re-attaches the loop's static
map (`probe.SetMap(l.m)`) before applying, letting miracle arms validate the
terrain vocabulary in the dry-run exactly as the real apply and replay will.
Model output enters
the sim only through these two doors, as recorded input. The protocol `Status`
carries `GuardianCharges` (JSON tag `metatron_charges`, frozen — spec 052 ruling 2)
so clients render the ⚡ bank without a state fetch.

`Loop.InjectOperator` (the `inject_operator` command, spec 034 R8) is a THIRD
door, distinct from both above: the daemon's operator-event door, whitelisted
to `daemon.llm_warning` only (`injectOperatorWhitelist`, kept separate from
`injectSocialWhitelist` — one door is the mind's model-output isolation
boundary, the other is the daemon's operator surface, and the two must never
share a whitelist). It exists because `store.AppendEvents` has no internal
locking and the loop is the log's single writer; `daemon.started`/`stopped`
append directly only because they run outside `Run`'s lifetime, but a
provider-health condition transition ([[llm-provider-health]]) fires from
worker/preflight goroutines *while the loop runs*, so its durable event must
ride this command door to keep seq assignment and tick-stamping inside the
loop goroutine. Every whitelisted type is a reducer no-op, so `handleCommand`
skips `InjectSocial`'s dry-run entirely — there is no world-state atomicity to
protect. It fails cleanly (mirroring `InjectSocial`) if the loop has stopped,
letting the daemon's condition hook degrade to a log line only.

## Connections

[[llm-provider-health]]'s condition hook is `InjectOperator`'s sole caller.
[[game-clock]] supplies intervals; the [[executor]] supplies tick events;
[[sim-state-reducer]] is the mutation path; [[event-log]] and [[snapshots]] persist;
[[ipc-server]] feeds commands in and broadcasts events out; [[daemon-lifecycle]] owns
the ctx whose cancellation triggers the final snapshot. The landing ladder's
budgets and classes come from [[cognition]] (`cognition.ClassFor`), whose router
and estimators produce the snapshot/landing metadata the ladder judges.
[[guardian-miracles]]'s four event types ride `InjectSocial`'s whitelist, as do
[[guardian-orders]]'s three injected order-lifecycle types and [[mental-maps]]'s
`metatron.place_revealed`. [[memory-retrieval]]'s embedder driver injects
`agent.memory_embedded`/`agent.situation_embedded` through the same door and
records `cog.memory_divergence` alongside the other `cog.*` telemetry;
`stampSeqs` exists specifically so its `Memory.Seq` targeting stays
replay-stable ([[sim-state-reducer]]).
[[tool-loop]] is the caller behind both doors' villager/guardian traffic since
spec 017 — its handlers wrap `InjectIntent` (world verbs, `set_plan`) and
`InjectSocial` (`muse`, and the Guardian's nudges/`work_miracle`), and its buffered
`CallRecord`s land as the `cog.tool_call` batch through the same social door.
The [[executor]]'s `run.ended` declaration is what flips the loop into the
ended posture; the [[morgue]] is the consumer of the two spec 044 whitelist
types and of the narrowed ended-world door. Since spec 061, `rungPairCooldown`
is this note's half of the conversation loop damper — [[social-fabric]] owns
the mind-side novelty SHIM one layer above it, and [[sim-state-reducer]] owns
the `PairTalks` ledger both read. Since spec 058, the plan rung's
`OutcomeClamped` path is this note's half of the clamp-with-notice feature —
[[tool-loop]] owns the matching `VerdictLandedClamped` for expressive text,
and [[tool-registry]] owns `set_plan`'s schema no longer declaring `maxItems`.

## Operational notes

Measured throughput at max speed on the target machine: ~1.65M ticks/sec, measured on
the TASK-2-era placeholder sim before the village systems landed (the full village does
more work per tick). Store errors inside the loop are fatal (the daemon exits) — an
unwritable log must never silently diverge from state.
