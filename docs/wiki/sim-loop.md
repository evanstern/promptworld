---
name: sim-loop
description: The single-goroutine fixed-timestep loop — tick execution, pacing modes, auto-slow degradation, and command handling. Intent landing (the metered ladder) and the InjectSocial/InjectOperator doors are children — load this note first for loop modes/pacing, route to a child for injection-door detail.
kind: component
sources:
  - internal/sim/loop.go
verified_against: 5761edb18e2b5fb49c6a03a050b0d871f5546c05
size_budget_exempt: pre-existing overage (predates spec 101's one-line injectSocialWhitelist addition, the only touch this pass made) — a full pacing/degradation vs. command-handling summary-style split is a dedicated future pass, not this task's scope
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

`runTick`: since spec 104, opens with `l.state.AdvanceTo(nextTick)` — the
derived-progress engine (in-flight walk segments, needs decay, gru motion,
[[sim-state-agent-fields]]) runs everything scheduled STRICTLY BEFORE
`nextTick`, i.e. everything due at the PREVIOUS tick, so it executes after
every event this loop (or a command-door injection landing between
`runTick` calls) recorded at that tick and before `stepEvents` reads state
for the new tick — then compute `stepEvents(state, map, nextTick)` (pure), advance `state.Tick`,
pre-assign the batch's store seqs (`stampSeqs`, spec 042 — below), apply
each event through the reducer (`Apply` itself opens with the same
`AdvanceTo(e.Tick)` hook, so every fold path — recovery, `replayToTick`,
the mind/TUI replicas — gets the identical strictly-before interleaving for
free with no per-consumer wiring), `AppendEvents` in one transaction, then `notify`
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
rate of 0 like a paused one. Since spec 054, `status()` also composes
additive `omitempty` `ScenarioExercise`/`ScenarioOutcome` fields — only when
`s.ScenarioExerciseID()` reports an armed scenario — from
`sim.ExerciseOutcome(s, id)`, inside the same loop goroutine as `Tick`, so
the pair is always coherent with it; an ambient world's status bytes are
unchanged ([[scenario-machinery]]).

Auto-slow (`observeWindow`): every `degradeWindow = 5s` the loop compares achieved
ticks/sec against the requested rate; sustained shortfall below 90% emits
`clock.degraded` (with the measured rate), recovery to ≥95% emits `clock.recovered`.
At max speed whatever is achieved is the contract — no degradation events.

**Determinism scope (spec 092/TASK-75)**: the measured rate `clock.degraded`
carries (`l.measured`, `windowTicks / elapsed`) is wall-clock, so two
INDEPENDENT live runs from the same seed can diverge in state hash even
though every RNG draw agrees ([[deterministic-rng]] has the full scope
note). It is still payload-carried, never re-derived, so REPLAY of a
recorded log stays exact — this loop's only non-determinism is per-run,
never per-replay.

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

Two further topics are split into child notes: [[sim-loop-landing-ladder]]
covers `Loop.Do`/`Loop.InjectIntent`/`InjectArgs` and the six-rung landing
ladder a metered intent climbs before it lands, plus the resulting
`cog.outcome` verdict; [[sim-loop-injection-doors]] covers `Loop.InjectSocial`
(the mind's whitelisted conversation/consolidation/musing/chronicle/nudge/
miracle/telemetry door) and `Loop.InjectOperator` (the daemon's separate
operator-event door).

## Connections

[[game-clock]] supplies intervals; the [[executor]] supplies tick events;
[[sim-state-reducer]] is the mutation path; [[event-log]] and [[snapshots]]
persist; [[ipc-server]] feeds commands in and broadcasts events out;
[[daemon-lifecycle]] owns the ctx whose cancellation triggers the final
snapshot. [[scenario-machinery]]'s `sim.ExerciseOutcome` is what `status()`
reads for the spec-054 `ScenarioExercise`/`ScenarioOutcome` status fields.
The [[executor]]'s `run.ended` declaration is what flips the loop into the
ended posture; the [[morgue]] is a consumer of that narrowed ended-world
posture (see [[sim-loop-injection-doors]] for the two whitelist types it
reads). [[sim-loop-landing-ladder]] and [[sim-loop-injection-doors]] carry
the rest of this note's Connections (cognition, tool-loop, guardian family,
social-fabric, memory-retrieval, grounded-feedback, bundle-tools).

## Operational notes

Measured throughput at max speed on the target machine: ~1.65M ticks/sec, measured on
the TASK-2-era placeholder sim before the village systems landed (the full village does
more work per tick). Store errors inside the loop are fatal (the daemon exits) — an
unwritable log must never silently diverge from state.

**Store-error posture is a ratified decision, not an accident (spec 099 D2, TASK-76).**
Every `AppendEvents` call inside the loop (`runTick`, `handleCommand`,
`observeWindow`) returns its error straight up through `Run` to the daemon's
caller with NO retry — the daemon process exits rather than continuing on
in-memory state the log can no longer corroborate. Rationale: a bounded-retry
seam would trade a clean, loud failure for a liveness/consistency tradeoff
(retry-and-hope vs. give up cleanly) that nobody has needed at this repo's
current scale — there is no observed transient-write incident on record.
Fatal-by-doctrine STANDS as of spec 099; re-open only on one of two named
triggers: (1) a real-world transient-write incident actually observed on this
repo, or (2) multi-world hosting (concurrent worlds sharing infrastructure
change the failure-mode calculus enough to revisit). Until then, no retry code
ships. Site comments at all three `AppendEvents` call sites point back here.
