---
name: daemon-boot-recovery
description: Daemon boot sequence steps 0-5 and 7-8 — tool-registry gates, world.Open/bundle gate, pidfile/registry, store/meta validation, contiguity, snapshot+replay recovery and seeding, IPC wire-up, daemon.started, loop.Run
kind: pipeline
sources:
  - internal/daemon/daemon.go
verified_against: 801db7c1b15fb567732bc5c6063464e918353a4d
---

# Daemon boot: validate, recover, wire-up

Split from [[daemon-lifecycle]] (which retains the overview and keeps Shutdown,
`IsRunning`, `replayToTick`): the boot/preflight half of `daemon.Run(dir)`'s
startup sequence — steps 0 through 5 (validation through recovery/seeding) and
steps 7-8 (wire-up and serve). Step 6 (LLM orchestrator, governor, embedding,
teaching posture, calibration) is its own substance and splits to
[[daemon-orchestrator-startup]] and [[daemon-cognition-calibration]], which run
between step 5 and step 7 in the real sequence.

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

Step 6 — notify fan-out, the LLM orchestrator gate, governor/preflight/embedding
wiring, teaching posture, tool-loop warnings, and calibration/estimator seeding
— runs here in the real sequence; see [[daemon-orchestrator-startup]] and
[[daemon-cognition-calibration]] for its substance.

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

## Connections

[[cli-runtime-control]] runs this via `daemon` and detaches it via `start`;
[[sim-loop]] is the foreground engine; [[ipc-server]] the concurrent face;
[[event-types]] defines the `daemon.*` bookkeeping events it emits;
[[tool-registry]] and [[bundle-tools]] gate boot before the pidfile;
[[world-save-directory]], [[event-log]], [[snapshots]], [[worldmap-generation]],
and [[governance]] back recovery; [[world-tuning]] is what `seedTuning` loads,
clamps, and seeds; [[guardian-orders]] is what `seedSurvivalWatches` seeds; and
[[scenario-machinery]] is what `armScenario` arms — all before the loop starts.
See [[daemon-lifecycle]] for Shutdown, `IsRunning`, and `replayToTick`.

## Operational notes

Measured recovery: 18 ms after kill -9 across 95k events. A world killed while paused
wakes paused (pause state lives in snapshots/replay). Startup prints one line with
tick, game time, recovery ms, and socket path to stdout — in detached mode that lands
in `daemon.log`.
