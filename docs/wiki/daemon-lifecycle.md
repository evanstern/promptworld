---
name: daemon-lifecycle
description: Process lifecycle — startup recovery (snapshot+replay), pidfile with stale sweep, manifest↔meta validation, signal-driven graceful shutdown
kind: pipeline
sources:
  - internal/daemon/daemon.go
  - internal/daemon/curriculum.go
  - internal/daemon/estimator_persist.go
verified_against: aedcf52f680ed68910e185c3ccde44bd320517b6
---

# Daemon lifecycle

`daemon.Run(dir)` is the foreground primitive that turns a save directory into a
living world: validate, recover, bind, tick, and — on any exit path — leave the
directory in a state the next start can resume from losslessly.

## How it works

Startup sequence:

0. Tool-registry gates (spec 014, world-independent so they run first):
   `tool.Validate()` — the [[tool-registry]]'s internal consistency — and
   `sim.ValidateToolCoverage()` — every World tool on a roster has a sim
   resolver and duration, every Expressive tool's events are whitelisted. A
   malformed registry or roster aborts boot with a config error, never a
   tick-time failure.
1. `world.Open` — manifest validation ([[world-save-directory]]). Then, still
   before the pidfile, spec 036's bundle gate: `bundle.Discover(dir)` scans the
   world's `bundles/` folder once, validates every persona/tool bundle
   ([[bundle-tools]]), and freezes the result into a `BundleSet`. Each
   `BootReport` entry — a skipped tool or rejected bundle — prints one boot
   line naming its file, rule id, and offending value; a summary line
   (`daemon: bundles on (N tool(s) from M bundle(s))`) prints only when
   something loaded. Bundles are additive: an invalid bundle never bricks
   boot, an absent/empty `bundles/` changes nothing, and only an I/O failure
   reading the root is fatal.
2. `acquirePidfile` — one daemon per world: an existing pidfile with a live process
   (checked via `kill(pid, 0)`, EPERM counts as alive) is a hard error; a stale one
   (crash leftover) is swept along with the stale socket. Then `registerWorld`
   (TASK-43): a best-effort upsert into the advisory known-worlds registry
   ([[instance-manager]]) when the dir lives outside the worlds home — failures are
   logged and never block boot, and worlds inside the home are skipped (scan-owned).
3. `store.Open` + `validateMeta` — first run stamps `seed`/`format_version` into
   store meta; later runs must match the manifest exactly, catching save directories
   corrupted or spliced from two runs.
4. `CheckContiguity` — a holed event log refuses to run ([[event-log]]).
5. `recoverState` — newest hash-valid snapshot unmarshaled into
   `sim.NewState(seed, w.Map())` (genesis derives terrain-valid agent positions
   from [[worldmap-generation]]), then `ReplayEvents(seq > snapshot.seq)` through the
   reducer, bumping `Tick` to the highest event tick ([[snapshots]]). Recovery
   duration is measured and recorded. Then `seedMeetingConvention` (TASK-36):
   if the manifest declares a `meeting` block and recovered state carries no
   convention yet, a `meeting.convention_established` event (source `config`)
   is applied and appended at the recovered tick — landing in the log like
   genesis, so replay re-applies it and the seed never fires twice
   ([[governance]]). Then `seedTuning` (spec 048, [[world-tuning]]), the same
   build-event → `state.Apply` → `st.AppendEvents` shape: an absent
   `tuning.json` seeds nothing; a present one is parsed and clamped
   (`sim.ParseTuning`, one operator-visible warning per out-of-range field),
   failing boot on malformed JSON, wrong types, or unknown field names; the
   resolved effective set is compared against `state.EffectiveTuning()` and a
   `sim.tuning_applied` event lands only when they differ, so an unchanged
   file never grows the log on restart. This runs before the loop starts and
   before `mind.New`, so no tick and no planner schedule ever runs ahead of
   the tuned values. Then `seedSurvivalWatches` (spec 059, US1, the same
   `seedMeetingConvention`/`seedTuning` build-event → `state.Apply` →
   `st.AppendEvents` shape): if recovered state carries no ACTIVE
   system-origin survival watch yet, the three canonical watches
   (`sim.SurvivalWatchDefs` — near-death, starvation, exposure) land as
   `metatron.order_placed` events at the recovered tick; a fresh world's
   first boot seeds them, a pre-059 world's first boot after upgrade
   back-seeds them once, and every later boot finds them already active and
   injects nothing ([[guardian-orders]]). Then, still before the pidfile's
   recovery timing is stamped, `armScenario` (spec 054, [[scenario-machinery]]):
   if the manifest carries a `Scenario` block, resolves it via
   `sim.ExerciseByID` (a catalog miss here is a real corruption — `world.Open`
   already validated it — refused loudly) and calls `state.ArmScenario`,
   printing `daemon: scenario armed (<id>)`; an ambient world (no `Scenario`
   block) arms nothing. This runs before the loop exists, so no tick ever
   runs unarmed on a scenario world.
6. Notify fan-out + companions: the loop's notify goes to the IPC broadcast, the
   always-on soul scribe (which since spec 044 is constructed with the open
   store as its event source — `scribe.New(dir, seed, map, snapshot, st)` —
   because the [[morgue]] render is a pure fold over the FULL event history,
   which the boot snapshot alone cannot provide; reads are rare, per death or
   boot, and briefly serialize with the loop's appends on the store's single
   connection), the curriculum-ladder unlock observer (spec 046 US3,
   `curriculumObserver(w)` in `internal/daemon/curriculum.go` — always-on
   like the scribe and wired BEFORE the LLM gate, so a no-model world still
   records its unlocks: on observing `curriculum.stage_unlocked` it upserts
   the per-user `~/.promptworld/unlocks.json` record with the world's
   name/path and a pointer to the same batch's `curriculum.exercise_passed`
   event as evidence; `worlds.UpsertUnlock` warns-and-continues on any
   failure, so this advisory record can never perturb the loop — on a
   pre-scenario world (no rubric emitter reachable, [[curriculum-ladder]])
   it simply sits idle), and — when an orchestrator exists — the mind driver
   ([[agent-mind]]) and the Guardian component ([[guardian]], attached to the
   server via `SetGuardian` for the console); all consumers are non-blocking by
   contract. On a scenario world (spec 054, [[scenario-machinery]]), the
   scribe (`scr.SetScenario(exercise)`) and, when an orchestrator exists, the
   mind (`md.SetScenario(exercise)`) each receive the armed exercise id right
   after construction and before the loop starts — the scribe's call also
   re-renders the morgue immediately, so an already-ended scenario world's
   run summary carries the exercise line from the very first boot render on
   restart. The LLM
   orchestrator ([[llm-orchestrator]]) starts only when `llm.json` exists
   (`llm.LoadConfig` → `llm.New` → `srv.SetLLM`), closed on exit — config-gated,
   fully outside the loop, so inference failures can never touch the simulation.
   Inside that same conditional branch, spec 028's adaptive-throttle governor
   sampler is built and started: `newGovernorSampler(orch, loop)` wired to the
   server via `srv.SetGovernor` and run in its own goroutine
   (`go sampler.run(ctx)`), sampling aggregate staleness debt every
   `cognition.GovernorCadence` off the loop's non-blocking status door and
   issuing shed/recover decisions through the loop's `Govern` door — a no-LLM
   world builds zero governor machinery (FR-003, SC-004; see [[cognition]] for
   the debt arithmetic and controller, [[sim-loop]] for the `govern` command).
   In the same conditional branch, spec 034's provider-health surface is wired:
   `orch.SetConditionHook` installs a closure that prints a `daemon: WARNING
   llm provider …` (or recovered) log line and lands a durable
   `daemon.llm_warning` event through `loop.InjectOperator` — the loop's
   single-writer-preserving operator-event door ([[sim-loop]]); if the loop
   isn't running (the shutdown window) the durable leg is dropped and the log
   line is the sole record. `go orch.RunPreflight(ctx)` then starts the
   boot-time + periodic model-existence probe in its own goroutine, fired and
   forgotten under the shutdown ctx exactly like the governor sampler — boot
   never blocks or fails on its results (see [[llm-provider-health]]).
   Also in this branch, spec 042's embedding driver ([[memory-retrieval]]):
   `mind.New` now also takes `w.Manifest.MemoryRelevance` (the world's
   `memory_relevance` mode flag), and, ONLY when `orch.HasEmbedding()` reports
   `llm.json` routes the `embedding` kind, boot builds
   `mind.NewEmbedder(orch, loop, warnFn, w.Map(), w.Manifest.Seed,
   state.Marshal())` — a peer of the mind driver that watches committed
   `agent.memory_added` events — appends its `Observe` to the notify fan-out's
   consumers, and prints one boot line naming the embedding model and
   provider. An absent embedding route prints the off-switch line instead
   (`daemon: embedding off (no "embedding" route in llm.json — memories stay
   vectorless)`) and builds nothing — the same absence-is-the-feature-switch
   doctrine as a world with no `llm.json` at all. `warnFn`'s shape mirrors the
   provider-health hook just above (a daemon-log WARNING plus a durable
   `daemon.llm_warning`, `kind: "embedding-unavailable"`, through the same
   `loop.InjectOperator` door) but is debounced by the embedder driver itself,
   not by [[llm-provider-health]]'s detectors — which never RAISE a condition
   from embed traffic; since TASK-102 a successful embed does CLEAR a stale
   preflight condition on the embedding provider ([[llm-provider-health]]).
   On a teaching world (spec 039 US1/US3, `w.Manifest.Teaching` —
   [[world-save-directory]]), boot also derives and prints the teaching-posture
   default: `orch.EstimateForKind(llm.Kind("planner"))`'s live seconds-per-point
   feeds [[cognition]]'s `MaxSafeSpeed("planner", est)` for the highest
   planner-safe ladder rung, mapped to a `clock.Speed` via
   [[game-clock]]'s `SpeedForRate`; `teachingPostureBootLine` prints it in a
   calibrated flavor (the planner-serving provider's `CalibratedAt` is set) or
   a provisional one that also prompts `promptworld calibrate <world>` — the
   pessimistic bootstrap-seeded rung still applies either way, just honestly
   labeled. No planner-serving provider means no posture line and no default.
   Boot also surfaces the agent tool-use loop's config warnings the same
   warn-not-error way as the concurrency knob (`llmCfg.Local.Workers()`'s
   `workersWarn`): `llmCfg.Rounds()` (an out-of-range `loop_max_rounds`), both
   tiers' `ToolModeResolved()` (an unknown `tool_mode`), and — since spec 025
   (TASK-72) — the three per-kind token budgets (`llmCfg.PlannerTokens()`/
   `GuardianTurnTokens()`/`ConsolidationTokens()`, an out-of-range
   `max_tokens.<key>`) each print one line and
   clamp/default rather than aborting boot (TASK-52, [[llm-orchestrator]]). The
   normalized round cap and effective budgets then thread into both loop
   consumers: `mind.New(..., loopRounds, plannerTokens, consolidationTokens)`
   and `guardian.New(orch, loop, loop, ..., loopRounds, guardianTurnTokens)`
   (followed by `mt.SetBundles(bundleSet)` — spec 036 hands the boot-frozen
   bundle surface to the turn assembly, [[bundle-tools]] — and
   `mt.SetStage(w.Manifest.Stage, w.Manifest.CharterPreset)` — spec 046 US2
   hands the immutable stage + charter preset from the opened manifest the
   same boot-frozen way, so the stage tool ceiling and the stage-1
   instruction lock cannot be tampered mid-run, [[curriculum-ladder]] — and
   `mt.SetSkin(worldSkin)`, handing the same boot-frozen display skin
   `srv.SetSkin` (below) gave the status surface to the guardian turn
   assembly's prompts, spec 052 FR-003) —
   since spec 029 (US5) the loop is passed twice: once as the `Injector` it
   was always passed as, once as the new `LoopControl` seam Guardian's
   `pause`/`start`/`adjust_speed` meta tools drive ([[guardian-orders]],
   [[sim-loop]]'s `Loop.Do` — the same two-interfaces-one-value pattern
   `mind.New(loop, loop)` already used for the mind driver).
   Before the orchestrator is built, `cognition.ValidateKinds(llm.Kinds())` is a
   hard startup gate: every call kind must resolve to a registered decision class
   before a model is ever reachable ([[cognition]]). After it is built,
   `cognition.LoadProfile(w.CalibrationPath())` seeds the seconds-per-point
   estimators (`orch.SeedCalibration`, which since spec 035 also records each
   provider's `calibratedAt` from the profile — [[llm-orchestrator]]); a
   missing or unreadable `calibration.json` falls back to pessimistic
   bootstrap defaults
   (`cognition.BootstrapLocalSecPerPt`/`BootstrapCloudSecPerPt` — fail toward
   reflex, never toward stale action), and since spec 035 (FR-001,
   contracts/warnings.md §1) both branches print the full
   `uncalibratedBootWarning(worldName)` block instead of a one-line hint: the
   UNCALIBRATED statement, `cognition.HorizonSummary` evaluated at the
   bootstrap seeds (the identical string `promptworld calibrate` itself
   prints, FR-006 — [[cognition]]), and the exact `promptworld calibrate
   <world>` command to run. The profile-seeded branch is untouched and stays
   byte-identical (US2 AC2). After calibration seeding,
   `cognition.LoadEstimatorState(w.EstimatorStatePath())` +
   `orch.SeedPersisted` raise each provider's seed to any higher persisted
   live estimate (TASK-113, max(seed, persisted) — a malformed
   `estimator_state.json` downgrades to a warning, never a crash), and a
   daemon-side persister goroutine flushes `orch.SnapshotEstimators()` back to
   that file every 5 minutes plus once synchronously after `loop.Run(ctx)`
   returns, so learned drift survives restarts.
   `orch.SetRecalibrateHook(md.RecalibrateSignal)` wires
   the drift signal: a provider's estimator breaching its spike-rate threshold lands
   as `cog.recalibration_recommended` telemetry.
7. Wire-up: `ipc.NewServer(w, st, cancel)` where cancel is the
   `signal.NotifyContext(SIGTERM, SIGINT)` cancel — so the protocol `shutdown`
   command and Unix signals share one graceful path. Right after, the world's
   display skin (spec 052 FR-003) loads once — `skin.Load(dir)` — boot-frozen
   like the bundle set, with any loader notices printed as one
   `daemon: skin: <notice>` line each (the bundle `BootIssue` convention: a
   typo never bricks the world); `srv.SetSkin` hands it to the status/console
   surface, and — when an orchestrator exists — `mt.SetSkin` hands the same
   boot-frozen skin to the guardian turn assembly's prompts (above). `SetLoop`
   closes the loop↔server mutual reference. The stale socket is removed
   before `Listen`.
8. `daemon.started` event appended (payload carries tick and `recovery_ms`) and
   broadcast; then `srv.Serve()` in a goroutine, and — on a teaching world with
   a computed default — a goroutine applies the teaching-posture speed through
   the loop's normal `set_speed` command (`loop.Do("set_speed", sp)`) so it
   lands as a recorded `clock.speed_set` event just like a player's own speed
   change ([[event-types]]), replaying byte-identically; a failed apply (loop
   already stopping) only logs. Then `loop.Run(ctx)` in the foreground.

Shutdown: ctx cancellation (signal or `shutdown` cmd) returns from `Run` after the
loop's final snapshot; `daemon.stopped` is appended; deferred cleanup closes the
server (removing the socket), the store, and the pidfile — the pidfile only if it
is still ours (a slow shutdown can overlap a successor daemon that has already
claimed it; the CLI's stop wait is 30 s to match). SIGKILL skips all of this —
that is the crash path recovery is tested against.

`IsRunning(dir)` (used by CLI `start`/`stop`/`status`) reads the pidfile and probes
liveness without touching the world — deliberately, since TASK-147: it reads
`world.PidPathIn(dir)` (a pure path join, not a validating `Open`) rather than going
through a `*World`, so pidfile liveness stays checkable for a world this build can no
longer `world.Open` (e.g. an older `format_version`). Before this, `IsRunning` opened
the world first, so a running old-version daemon could never be detected — let alone
stopped — by a newer binary; `migrate` also refuses to touch a live daemon, so the two
gates combined could deadlock a world at an unsupported version.

`replayToTick(seed, m, st, cutoff)` (spec 043) sits beside `recoverState` as a
read-only reconstruction primitive the boot path never calls: it rebuilds
state as of an arbitrary tick by replaying the event log from genesis (a
snapshot may postdate the cutoff, so genesis replay is the only cutoff-correct
source), skipping — not stopping on — events past the cutoff, and tallying by
type, rather than aborting on, events the current reducer rejects, so a
legacy-format save whose manifest `world.Open` would refuse can still be
reconstructed from just its seed + map. Its consumer is the spec-043
replay-determinism harness (`internal/daemon/context_replay_test.go`,
[[decision-context]] / [[testing-strategy]]).

## Connections

[[cli-promptworld]] runs this via `daemon` and detaches it via `start`; [[sim-loop]]
is the foreground engine; [[ipc-server]] the concurrent face; [[event-types]] defines
the `daemon.*` bookkeeping events it emits; [[cognition]] supplies the startup kind
gate, the calibration profile it seeds into the orchestrator, (spec 035) the
`HorizonSummary` the boot warning block quotes verbatim, (spec 028)
the debt arithmetic and hysteresis controller the governor sampler drives,
and (spec 039) the `MaxSafeSpeed` the teaching-posture default computes from;
[[game-clock]]'s `SpeedForRate` turns that rung into the `clock.Speed` applied
through the loop's `set_speed` door;
[[guardian-orders]] is what the `LoopControl` seam wired here (spec 029) drives.
[[llm-provider-health]] is what the condition hook and preflight goroutine wired
here (spec 034) drive; its durable event rides [[sim-loop]]'s `InjectOperator`
door. [[memory-retrieval]] is the spec 042 embedding driver wired here only
when `orch.HasEmbedding()`; its failure warning shares [[sim-loop]]'s
`InjectOperator` door and the `daemon.llm_warning` event type with
[[llm-provider-health]] but is a separate, debounced-by-the-driver signal.
[[curriculum-ladder]] is what the always-on unlock observer and the
`SetStage` handoff wired here (spec 046) serve.
[[world-tuning]] is the spec 048 manifest `seedTuning` loads, clamps, and
seeds right after the meeting-convention seed. [[guardian-orders]] is what
`seedSurvivalWatches` (spec 059) seeds right after the tuning seed and before
the loop starts. [[scenario-machinery]] is what `armScenario` (spec 054)
arms right after the survival watches and before the loop starts, and what
the scribe's/mind's `SetScenario` handoffs (below, alongside `SetStage`)
carry the armed exercise id into.

## Operational notes

Measured recovery: 18 ms after kill -9 across 95k events. A world killed while paused
wakes paused (pause state lives in snapshots/replay). Startup prints one line with
tick, game time, recovery ms, and socket path to stdout — in detached mode that lands
in `daemon.log`.
