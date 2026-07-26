---
name: cli-runtime-control
description: promptworld CLI runtime control and monitoring commands — daemon/start/stop process control, status (WARNING/horizon/posture/stage/exercise lines, offline reconstruction), pause/resume/speed, teaching toggle, ui/attach/tail clients
kind: component
sources:
  - cmd/promptworld/commands.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# CLI: runtime control commands

Split from [[cli-promptworld]] (full subcommand list there): starting/stopping a
daemon, reading and steering its live status, and the interactive/streaming clients
that ride the same socket — `daemon`, `start`, `stop`, `status`, `pause`/`resume`/`speed`,
`teaching`, `ui`, `attach`, `tail`.

## How it works

- `daemon <world>` — the foreground primitive: `daemon.Run` directly.
- `start <world>` — detached start: re-execs itself (`os.Executable()` + `daemon <dir>`)
  with stdio appended to `daemon.log` and `Setsid` to leave the session, then polls
  the socket up to 5 s for a status round trip before reporting success. Never waits
  on the child.
- `stop <world>` — sends `shutdown` over the socket (falls back to SIGTERM if the
  socket is dead but the pid lives), waits ≤30 s for the pidfile to clear. Idempotent:
  "daemon not running" exits 0. Since TASK-147, deliberately version-agnostic: it
  checks the directory looks like a world (`requireWorldDir`, a bare manifest-presence
  stat) and reaches the daemon via `daemon.IsRunning`/the socket path alone, never
  `world.Open` — a running daemon on a world whose `format_version` this build can no
  longer `Open` (the v4-world stop/migrate deadlock a version-gated `Open` produced,
  since `migrate` also refuses a live daemon) is still stoppable.
- `status <world> [--json]` — online: full `StatusData` via the client
  (`renderStatusHuman`), which — since spec 034 — prints one
  `WARNING llm provider "<name>": <detail> — <remedy>` line per provider
  carrying an active health condition, right after the clock line
  (`llmConditionWarnings`, [[llm-provider-health]]); then — since spec 035
  (FR-004/US3) — one `providerCalibrationLine` per provider (`  llm <name>
  (<model>): calibrated <timestamp>` or `... : uncalibrated (bootstrap)` when
  the wire's `calibrated_at` is empty), so a provider's calibration state is
  visible for the whole life of the daemon, not just a boot line that
  scrolled away. A world with no LLM status or every provider healthy and
  calibrated renders byte-identical to pre-034 output plus these rows. Since
  spec 037 (US3, FR-008), `renderStatusHuman` also appends `horizonStatusLines`
  — one `horizon: <class> suppressed at <speed> — calibrate or slow down
  (skipped N)` / `horizon: <class> thinking at <speed>` line per watched
  class in the wire's `Horizon`, mirroring the TUI's calibrate-vs-slow-down
  remedy split (`horizonRemedy`); absent entirely when `Horizon` is absent
  (a no-LLM world's output is unchanged). Since spec 039 (US4), a teaching
  world's status additionally appends one `postureStatusLine` after the
  horizon block: `teaching posture: <rung> (calibrated)` or `... (provisional
  — run `promptworld calibrate <world>`)`, read from the wire's
  `StatusData.Posture` — absent (and so the line omitted) for any non-teaching
  or pure-sim world. Since spec 046 ([[curriculum-ladder]]), a ladder world's
  status appends one `stage: <skin name> (<stage-id>)` line (plus an
  `[overridden]` marker when creation skipped ahead) via the same
  `stageStatusLine` that `new` prints, right before the log line — read live
  from the wire's `StatusData.World.Stage`/`StageOverridden`; empty stage
  (every pre-046/pre-ladder world) omits the line, output unchanged. Since
  spec 054 ([[scenario-machinery]]), a scenario world's status likewise
  appends one `exercise: <id> — <outcome>` line (`scenarioStatusLine`,
  `failed` rendering `failed (run ended)`) right after the stage line, read
  live from the wire's `StatusData.World.ScenarioExercise`/`ScenarioOutcome`;
  empty (every ambient world or old daemon) omits the line. Since
  spec 044, `clockLine` renders the run-over posture:
  when the wire's `Clock.Ended` is set, the running/paused clock line is
  replaced entirely by `tick N (...) — run ended day N, all villagers dead;
  world is an archive (read-only)` ([[morgue]]).
  Reaching a live daemon is likewise version-agnostic (same `requireWorldDir` +
  socket-dial-first shape as `stop`, TASK-147): the dial happens before any
  `world.Open`, so a running daemon on a world this build can no longer `Open`
  still answers normally. Offline: last-known state reconstructed read-only from the
  store (via `worlds.OfflineSnapshot` — latest snapshot plus a fold of any newer
  events, [[instance-manager]]), clearly labeled "daemon not running"; a world whose
  manifest `format_version` this build can't `Open` (`world.ErrFormatVersionMismatch`,
  [[world-save-directory]]) also reports "daemon not running" rather than the
  migrate-hint error, since the dial above already ruled out a live daemon and the
  point of this check is reachability, not content. An ended
  world's offline output carries the same posture — a `run ended day N, all
  villagers dead; world is an archive (read-only)` line before the log line,
  and `--json` gains `ended`/`ended_day` clock keys present only when ended,
  mirroring the live `ClockStatus` omitempty fields so a living world's
  offline JSON is unchanged. The offline path carries the stage the same way:
  the human output prints `stageStatusLine` from the manifest, and `--json`'s
  `world` object gains `stage` (+`stage_overridden` when true) keys present
  only when the manifest carries a stage.
- `pause` / `resume` / `speed <v>` — one-shot time controls printing the resulting
  clock line; `speed` (and the `attach` REPL's `speed` command) go through
  `setSpeedLine` (spec 035 FR-002), which appends a `WARNING: <text>` line
  after the clock line whenever the `set_speed` reply carries
  `StatusData.Warning` — additive and non-blocking, since the speed change
  has already applied by the time the warning prints
  ([[ipc-server]], [[ipc-protocol]]). `pause`/`resume` replies never carry a
  warning.
- `teaching <world> [on|off]` (spec 039 US4) — offline manifest toggle: with no
  on/off argument, `cmdTeaching` reports the current marker; with one, it
  rewrites `world.json` via `world.SetTeaching` and says the change applies at
  the next daemon start (a running daemon reads `Teaching` only at boot). Never
  fails on state, only on IO/parse errors.
- `ui <world>` — the full-screen Bubble Tea client ([[tui-client]]): map, chronicle,
  guardian, villagers panes over a live world replica (villagers renamed from
  souls, spec 015); runs in the alternate screen.
  If the TUI quits on an unrecoverable protocol error (`Model.FatalErr()`, e.g. a
  reply over the IPC cap — TASK-19), the command returns it as a real error and
  exits non-zero.
- `attach <world>` — line-mode: status header, live subscribe streamed to stdout,
  stdin commands (`pause`, `resume`, `speed <v>`, `status`, `quit`); handles
  `dropped` pushes by re-subscribing. Quit detaches; the world keeps running.
- `tail <world> [--since SEQ] [--follow]` — history from the store (default last 20),
  works with no daemon; `--follow` additionally subscribes live and requires one.

## Connections

`status`'s WARNING block and calibration rows read [[llm-provider-health]] and
[[llm-orchestrator]]; its `horizonStatusLines` and `speed`'s appended WARNING both
read [[ipc-server]]-composed `StatusData.Horizon`/`Warning` ([[ipc-protocol]]);
its `postureStatusLine` reads [[world-save-directory]]'s `Manifest.Teaching` and
`StatusData.Posture`; its stage/exercise lines read [[curriculum-ladder]] and
[[scenario-machinery]] facts off [[ipc-protocol]]; its run-over posture reads
[[morgue]]. `ui` is [[tui-client]]; its header badge and guardian-pane render the
same `Horizon`/exercise data. `daemon`/`start` are what [[daemon-lifecycle]] runs.
See [[cli-promptworld]] for exit discipline, arg resolution, and the sibling
families ([[cli-world-lifecycle]], [[cli-guardian-ops]]).
