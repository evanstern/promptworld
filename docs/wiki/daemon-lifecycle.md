---
name: daemon-lifecycle
description: Process lifecycle — startup recovery (snapshot+replay), pidfile with stale sweep, manifest↔meta validation, signal-driven graceful shutdown
kind: pipeline
sources:
  - internal/daemon/daemon.go
verified_against: cc514f7ff456fefbcfe289471c5a1467b8e724df
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
   ([[governance]]).
6. Notify fan-out + companions: the loop's notify goes to the IPC broadcast, the
   always-on soul scribe, and — when an orchestrator exists — the mind driver
   ([[agent-mind]]) and the Metatron component ([[metatron]], attached to the
   server via `SetMetatron` for the console); all consumers are non-blocking by
   contract. The LLM
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
   not by [[llm-provider-health]]'s detectors — which never observe embed
   traffic at all (a known gap, TASK-102).
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
   `MetatronTurnTokens()`/`ConsolidationTokens()`, an out-of-range
   `max_tokens.<key>`) each print one line and
   clamp/default rather than aborting boot (TASK-52, [[llm-orchestrator]]). The
   normalized round cap and effective budgets then thread into both loop
   consumers: `mind.New(..., loopRounds, plannerTokens, consolidationTokens)`
   and `metatron.New(orch, loop, loop, ..., loopRounds, metatronTurnTokens)`
   (followed by `mt.SetBundles(bundleSet)` — spec 036 hands the boot-frozen
   bundle surface to the turn assembly, [[bundle-tools]]) —
   since spec 029 (US5) the loop is passed twice: once as the `Injector` it
   was always passed as, once as the new `LoopControl` seam Metatron's
   `pause`/`start`/`adjust_speed` meta tools drive ([[metatron-orders]],
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
   byte-identical (US2 AC2). `orch.SetRecalibrateHook(md.RecalibrateSignal)` wires
   the drift signal: a provider's estimator breaching its spike-rate threshold lands
   as `cog.recalibration_recommended` telemetry.
7. Wire-up: `ipc.NewServer(w, st, cancel)` where cancel is the
   `signal.NotifyContext(SIGTERM, SIGINT)` cancel — so the protocol `shutdown`
   command and Unix signals share one graceful path. `SetLoop` closes the
   loop↔server mutual reference. The stale socket is removed before `Listen`.
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

`IsRunning(dir)` (used by CLI `start`/`stop`) reads the pidfile and probes liveness
without touching the world.

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
[[metatron-orders]] is what the `LoopControl` seam wired here (spec 029) drives.
[[llm-provider-health]] is what the condition hook and preflight goroutine wired
here (spec 034) drive; its durable event rides [[sim-loop]]'s `InjectOperator`
door. [[memory-retrieval]] is the spec 042 embedding driver wired here only
when `orch.HasEmbedding()`; its failure warning shares [[sim-loop]]'s
`InjectOperator` door and the `daemon.llm_warning` event type with
[[llm-provider-health]] but is a separate, debounced-by-the-driver signal.

## Operational notes

Measured recovery: 18 ms after kill -9 across 95k events. A world killed while paused
wakes paused (pause state lives in snapshots/replay). Startup prints one line with
tick, game time, recovery ms, and socket path to stdout — in detached mode that lands
in `daemon.log`.
