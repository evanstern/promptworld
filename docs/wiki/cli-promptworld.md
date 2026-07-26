---
name: cli-promptworld
description: The single promptworld binary — subcommand dispatch, world management, daemon control, observation commands, v1→v2→v3 migration, exit discipline
kind: component
sources:
  - cmd/promptworld/main.go
  - cmd/promptworld/commands.go
  - cmd/promptworld/calibrate.go
  - cmd/promptworld/ps.go
  - cmd/promptworld/work.go
  - cmd/promptworld/divergence.go
  - cmd/promptworld/stages.go
verified_against: 31c893e0406653197e467a89b2fdb96f0bcf2ee0
---

# promptworld CLI

One binary serves every role: daemon, client tools, world management. `main.go` is a
plain dispatch table; behavior lives in `commands.go`, except `calibrate` in its own
`calibrate.go`, `ps` in `ps.go` ([[instance-manager]]), and the guardian's operator
door in `work.go`. The prose contract is
`specs/001-world-daemon/contracts/cli.md` (extended by
`specs/008-instance-manager/contracts/cli.md` for names/`ps`/`new`, and
`specs/007-cognition-horizon/contracts/cli.md` for `calibrate`).

Since spec 052 (TASK-121, FR-008) `dispatch` (`main.go`) advertises only the
canonical, guardian-voiced subcommand names in the printed usage text —
`guardian` and `work` — while the pre-052 fiction names `metatron` and
`miracle` stay registered as HIDDEN, fully functional compatibility aliases
(same handler function, so an old script can never drift from current
behavior; `TestGuardianWorkAliases`/`TestUsageShowsCanonicalOnly` in
`alias_test.go` pin both the equivalence and the usage text's silence on the
retired names).

## How it works

Exit discipline: 0 on success; 1 with a one-line `promptworld <cmd>: error` on stderr;
2 for usage errors.

Every per-world command takes `<world>` — a name or a path (TASK-43). Arguments
containing `/` or starting with `.`/`~` are paths and behave exactly as before;
bare names resolve through `resolveWorld` → `worlds.Resolve`
([[instance-manager]]: worlds home first, then the known-worlds registry;
ambiguous or unknown names exit 1). `worldArg`/`parseWorldFlags` wrap the older
`dirArg`/`parseDirFlags` with that resolution.

- `new <name> [--at DIR] [--seed] [--teaching] [--stage] [--override] [--charter-preset] [--scenario]` /
  `new <path> [--name] [--seed] [--teaching] [--stage] [--override] [--charter-preset] [--scenario]` — a bare-word
  argument is a name: the world is created at `<worlds-home>/<name>` (or exactly
  `--at DIR`, which also registers it in the known-worlds registry), manifest name =
  the argument, validated by `worlds.ValidateName`. A path-shaped argument keeps the
  legacy form byte-for-byte: create at that path, name from `--name` (validated) or
  the basename (unvalidated, backward-compatible). Both forms then run the same
  creation: `world.Create` + store + genesis `world.created`
  event plus the genesis tuning pin (spec 057: one `sim.tuning_applied` with
  the full default dial set, fixing the world's doctrine at birth —
  [[world-tuning]]), writes the default `llm.json`, seeds the eight personas and the
  guardian's
  charter (`persona.Genesis`, the one-and-only persona write — [[agent-mind]],
  [[guardian]]), and
  appends the tick-0 secret events ([[social-fabric]]). Random default seed (crypto-random,
  right-shifted 12 bits to stay comfortably printable). Since spec 034, the
  printed summary appends a line naming the fresh-world local model and its
  pull command (`local model: cogito:3b — pull it first if you haven't: ollama
  pull cogito:3b`), read from `llm.DefaultConfig()` itself so the hint can
  never drift from what was just written ([[llm-orchestrator]]). `--teaching`
  (spec 039) stamps the manifest's `Teaching` marker at birth via
  `world.SetTeaching` — set-after-create, so `world.Create`'s own signature
  stays untouched for its other callers — telling the daemon to default this
  world's speed to the highest planner-safe rung at every boot
  ([[daemon-lifecycle]], [[cognition]]). Since spec 046 ([[curriculum-ladder]]),
  `new` also resolves a curriculum **stage**: `--stage stage-1..stage-4`
  (validated against the `world.Stage1..4` constants), defaulting to the
  player's highest earned stage from the per-user unlocks record
  (`worlds.LoadUnlocks`) — stage-1 for a brand-new player. An unearned stage
  refuses with an informed error naming every skipped stage by its skin
  display name (`skin.StageName`) unless `--override`, which proceeds and
  records the override honestly. The resolved stage, override flag, and
  charter preset are stamped into the manifest set-after-create via
  `world.SetStage` (the `SetTeaching` pattern — write-once, no toggle
  command). `--charter-preset` (`default`|`tutor`, validated by
  `world.ValidCharterPreset`) picks the charter `persona.Genesis(dir,
  charterPreset)` seeds; a stage-1 world defaults it to `tutor` unless the
  player explicitly opts out with `--charter-preset default`. The printed
  summary gains a trailing `stage: <skin name> (<stage-id>)` line
  (`stageStatusLine`, `[overridden]` suffix when forced). Since spec 054
  ([[scenario-machinery]]), `--scenario <id>` resolves against the compiled
  `sim.ScenarioExercises` catalog (an unknown id refuses, listing every
  cataloged exercise); the scenario IMPLIES its stage and pins its authored
  seed, so an explicit `--stage`/`--seed` may only agree (a mismatch
  refuses) — the earned-stage gate above still applies to the implied
  stage, a scenario never bypasses it. A resolved scenario writes the
  manifest's `Scenario` block set-after-create via `world.SetScenario` (the
  `SetStage` pattern) and the summary gains a trailing `scenario: <id> —
  <concept>` line naming the exercise panel's key (`6`).
- `migrate <world>` — the one-time upgrade of an older world (v1 or v2) to the
  current format (spec 012 US6 for v1→v2, spec 013 for v2→v3 —
  [[world-migration]]): resolves `<world>` via `resolveWorldForMigrate`, which
  unlike `resolveWorld`/`worlds.Resolve` must reach older-format worlds that this
  build cannot `world.Open` — a path argument passes through verbatim, a bare name
  resolves against the worlds home then the known-worlds registry by manifest
  *presence* alone, never the version gate. Hands the whole
  archive/transform/rewrite ceremony to `world.Migrate`
  ([[world-save-directory]]), which admits a v1 or v2 source (a v1 world chains
  1→2→3 in one run; an already-current world is refused outright) and archives the
  original database under a name keyed to the source format (`world.v1.db` or
  `world.v2.db`). Prints a human summary (seed, villagers carried, continuation
  tick, source event count, archive path, and the `start` command to run next).
- `ps [--all] [--json]` — machine-wide listing of worlds with live-proven state
  ([[instance-manager]]): discovery over the worlds home + registry, concurrent
  bounded probes, `NAME STATE PID TICK GAME TIME SPEED LLM PATH` table or a JSON
  array reusing the `status --json` vocabulary. Default shows live-pid states
  (`running`/`paused`/`unresponsive`); `--all` adds `stopped`/`missing`/`unreadable`.
  Empty listing prints "no worlds running", exit 0.
- `daemon <world>` — the foreground primitive: `daemon.Run` directly.
- `start <world>` — detached start: re-execs itself (`os.Executable()` + `daemon <dir>`)
  with stdio appended to `daemon.log` and `Setsid` to leave the session, then polls
  the socket up to 5 s for a status round trip before reporting success. Never waits
  on the child.
- `stop <world>` — sends `shutdown` over the socket (falls back to SIGTERM if the
  socket is dead but the pid lives), waits ≤30 s for the pidfile to clear. Idempotent:
  "daemon not running" exits 0.
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
  Offline: last-known state reconstructed read-only from the store (via
  `worlds.OfflineSnapshot` — latest snapshot plus a fold of any newer events,
  [[instance-manager]]), clearly labeled "daemon not running"; an ended
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
- `guardian <world> [message...]` (canonical since spec 052 FR-008; the pre-052
  name `metatron` survives as a hidden compat alias, same handler) — the
  console one-shot ([[guardian]], TASK-12): with
  a message, one mediated turn (prints surfaced moments, the reply, any landed
  `⚡ vision/omen` line, a `👁 watch set`/`👁 watch released` line for a placed or
  cancelled standing order, a `⏲` line for a landed pause/start/adjust_speed
  meta tool call (spec 029, [[guardian-orders]]), and the charge bank); without,
  a model-free status peek (charges, charter provenance, a `--- standing
  orders ---` block via `orderStatusLine` — id, fuzzy marker, origin,
  remaining game-day, status, condition — when any order stands, and recent
  soul notes).
- `work <world> <snap-time|give|move|remove> ... [--force]` (canonical since
  spec 052 FR-008; the pre-052 name `miracle` survives as a hidden compat
  alias, `work.go`) — the operator door
  for the guardian's workings ([[guardian-miracles]], spec 016 R6), a dedicated
  subcommand family independent of the `guardian` conversational path: `snap-time
  <day> <HH:MM>`, `give <villager> <item> <qty>`, `move <class> <x,y> <x1,y1>`,
  `remove <class> <x,y>` (`<class>` is `villager|structure|pile|terrain`; terrain
  is remove-only, villagers cannot be removed). Dials the daemon and calls the
  `miracle` IPC command directly (the wire command name is FROZEN, spec 052
  ruling 2) — no LLM involved. `--force` sets the gratis flag
  that waives the charge cost, an override reachable only from this CLI door, never
  from the angel's own turn. Prints the miracle summary (`(forced)` suffix when
  gratis) and the remaining charge bank.
- `llm <world> <kind> <prompt...> [--system] [--max-tokens]` — one-shot model call via
  the daemon's `llm_call` command; `formatLLMOneShot` prints the serving PROVIDER
  (never a tier — spec 024 FR-011), model, tokens, cost, and latency, plus a
  `skipped: name (reason)` line whenever the chain-walk passed over candidates
  ([[llm-orchestrator]]). `new` also writes the default `llm.json` config (v2
  registry shape; its hint says "edit providers/routes/budget").
- `calibrate <world> [--provider <name>] [--samples N]` (`--tier local|cloud|all`
  kept as a deprecated alias — on a v2 registry `local`/`cloud` select every
  zero-priced/priced provider with a deprecation note; on a legacy config it
  behaves exactly as before) — the cognition horizon's
  setup stage ([[cognition]], TASK-32): benchmarks the DECLARED PROVIDERS (spec
  024 T020: a legacy config runs the untouched pre-024 path over its two derived
  providers, byte-identical output; any v2 registry — or `--provider` — iterates
  `orch.ProviderNames()`, each reference call pinned via `Request.Provider` so
  the sample measures the named provider regardless of what its kind's chain
  currently resolves to) against fixed reference prompt shapes (default 5
  samples per shape; priced-provider spend is opt-in and announced up front),
  takes the median seconds-per-point, writes/merges `calibration.json` (one
  profile entry per provider name — the shape `cognition.SeedFor` reads), and
  prints the horizon the hardware buys (e.g. "planner suppressed above 16x") via
  `horizonSummary`, which since spec 035 (R1/T004) delegates to
  `cognition.HorizonSummary` — the same function the daemon's boot warning and
  the `set_speed` warning read, so the printed horizon can never disagree with
  the router (FR-006, [[cognition]]) — evaluating the registry
  across the watchable speed ladder (`planner`/`conversation`/`meeting` — `musing`
  dropped from the ladder with its retirement as a scheduled kind, spec 017).
  Every run also prints `sequentialFloorDisclosure` exactly once, adjacent to
  the horizon summary when one printed (a cloud-only or priced-only run with
  no horizon line still gets it): a reminder that calibration measures one
  reference call at a time while a live world drives the same endpoint
  concurrently, so the measured seconds-per-point is a floor and the live
  estimator adapts upward at runtime (spec 035 US4/FR-005). Since
  spec 017 (FR-011) the `planner-3pt` shape is a LOOP probe, not a bare
  completion: `villagerProbeJob` drives `toolloop.Run` with the real
  `tool.LoopRosterVillager()` roster and a no-op handler per tool (every read
  reports `read_ok`, every acting call reports `landed` — ending the loop on the
  model's first action, since calibration measures round-trip latency, not
  landings) so the seeded seconds-per-point is measured in the SAME whole-loop
  unit `Orchestrator.ObserveCognition` later feeds live ([[llm-orchestrator]],
  [[tool-loop]]) — a representative tool loop's wall time, not one call's. The
  probe's round cap is `cfg.Rounds()` (the daemon's own `loop_max_rounds`), so the
  calibration and the live cognition share one horizon. Reference shapes select
  by PRICING CLASS (`refShapesFor(priced)`, spec 024 T020 generalizing the old
  tier branch): zero-priced providers get the loop probe, priced providers'
  `consolidation-5pt` shape stays a plain single-shot `Submit` (consolidation did
  not adopt the loop, FR-014) — the guardian IS the cloud's loop cognition, but
  calibrating it would drive extra metered cloud calls the spec 017 contract
  doesn't invite; its live whole-loop observations converge the cloud estimator
  at run time instead. Uses an in-memory meter so it never contends
  with a running daemon's store; a provider whose every sample fails is not
  written.
- `divergence <world> [--json]` (`cmd/promptworld/divergence.go`, spec 042 US2) —
  offline gate evidence for the embedding-memory shadow→on flip: reads
  `cog.memory_divergence` events straight from the store (the `tail` pattern, no
  daemon required) and aggregates per-agent/per-game-day rows plus a whole-run
  total — mean overlap@K against the legacy ranking, the promoted-memory share
  (selections where relevance pulled in at least one memory legacy excluded),
  mean displacement, and mean vectorless-fallback count. Refuses with a one-line
  remedy ("set world.json memory_relevance to \"shadow\"") when no divergence
  events exist yet. `--json` emits the same rows machine-readably. The printed
  table is recorded evidence only — flipping the world to non-shadow relevance
  is an operator decision made on the board task, never an automatic action.
- `stages [--json]` (`cmd/promptworld/stages.go`, spec 046 US1 —
  [[curriculum-ladder]]) — the ladder's front door: an informed identity
  table over all four stages (always all four, never a difficulty menu), each
  row pairing the active skin's display identity (`skin.Stage` — name +
  one-line identity) with the skin-independent ladder facts (`stagesLadder` —
  since spec 063 a plain alias onto `internal/world.StagesLadder`, relocated
  there so the TUI help overlay's D9 guardian section can read the same
  table without `internal/tui` importing package `main`, [[grounded-feedback]] —
  mirroring the spec's table: the concept taught, what the world grants, and
  the evidence that unlocks the next stage — stage-4's reads "nothing — this
  is graduation") plus the earned state from the per-user unlocks record
  (`worlds.LoadUnlocks`): stage-1 is every player's unconditional floor
  (`stageEarned`); any other stage is earned only by an unlocks-record entry,
  whose proving world and exercise the earned line names; an unearned stage's
  row points at `new --stage <id> --override`. `--json` emits the same rows
  machine-readably (`proving_world`/`exercise` audit pointers only when
  earned by an entry). A missing/corrupt/unresolvable unlocks record simply
  means nothing beyond stage-1 is earned — the command never fails on record
  state. `highestEarnedStage` over the same record is what `new`'s default
  `--stage` selection uses.

`parseDirFlags` accepts both `cmd <arg> --flag` and `cmd --flag <arg>` orderings
(`parseWorldFlags` adds name resolution on top).

## Connections

[[daemon-lifecycle]] is what `daemon`/`start` run; [[instance-manager]] owns name
resolution, discovery, and the `ps` probe; [[ipc-client]] carries every online
command; [[world-save-directory]] and [[event-log]] back the offline paths;
[[game-clock]] formats times in `clockLine`/`eventLine`; `calibrate` writes the
profile [[cognition]] routes with; `migrate` hands off to [[world-migration]];
`work` (hidden alias `miracle`) hands off to [[guardian-miracles]]; `guardian`'s standing-orders
rendering reads [[guardian-orders]]; `status`'s WARNING block and `new`'s
pull-command hint read [[llm-provider-health]] and [[llm-orchestrator]];
`status`'s calibration rows and `calibrate`'s horizon summary both read
[[cognition]]'s `Calibrated`/`HorizonSummary`; `speed`'s appended WARNING line
reads the [[ipc-server]]-composed `StatusData.Warning` ([[ipc-protocol]]).
`status`'s `horizonStatusLines` reads the same [[ipc-server]]-composed
`StatusData.Horizon`, the wire shape [[tui-client]]'s header badge and
guardian-pane block also render. `status`'s `postureStatusLine` and `new`'s
`--teaching` flag / the `teaching` subcommand read and write
[[world-save-directory]]'s `Manifest.Teaching`/`SetTeaching` and
[[ipc-protocol]]'s `StatusData.Posture` (spec 039). `divergence` reads
`cog.memory_divergence` events [[memory-retrieval]]'s shadow-mode selector
records, offline via the same store/`tail` pattern as `tail`/`status`'s
offline path. `stages` and `new`'s stage resolution read the per-user
unlocks record (`worlds.LoadUnlocks`) and [[skin]]'s stage identity table, and
`status`'s stage line renders [[ipc-protocol]]'s
`WorldStatus.Stage`/`StageOverridden` — the [[curriculum-ladder]]'s CLI
surfaces (spec 046). `new --scenario` and `status`'s exercise line read
[[scenario-machinery]]'s `sim.ExerciseByID` catalog and
[[ipc-protocol]]'s `WorldStatus.ScenarioExercise`/`ScenarioOutcome` (spec
054).

## Operational notes

`start` failure says "check daemon.log". Detached daemons survive terminal close
(Setsid); a machine reboot needs a manual `start` (launchd integration is future
work — the foreground `daemon` subcommand is what a plist would run).
