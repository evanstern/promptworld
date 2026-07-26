---
name: cli-world-lifecycle
description: CLI world lifecycle — new (creation/stamping), migrate (v1-v4 upgrade), ps (discovery), stages (ladder)
kind: component
sources:
  - cmd/promptworld/commands.go
  - cmd/promptworld/ps.go
  - cmd/promptworld/stages.go
verified_against: 801db7c1b15fb567732bc5c6063464e918353a4d
---

# CLI: world lifecycle commands

Split from [[cli-promptworld]] (full subcommand list there): `new`, `migrate`, `ps`,
`stages` — creating, upgrading, and enumerating worlds.

## How it works

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
- `migrate <world>` — the one-time upgrade of an older world (v1 through v4) to the
  current format (spec 012 US6 for v1→v2, spec 013 for v2→v3, spec 041 for v3→v4,
  spec 068 for v4→v5 —
  [[world-migration]]): resolves `<world>` via `resolveWorldForMigrate`, which
  unlike `resolveWorld`/`worlds.Resolve` must reach older-format worlds that this
  build cannot `world.Open` — a path argument passes through verbatim, a bare name
  resolves against the worlds home then the known-worlds registry by manifest
  *presence* alone, never the version gate. Hands the whole
  ceremony to `world.Migrate`
  ([[world-save-directory]]), which admits a v1-v4 source (an older world chains
  every remaining step in one run, e.g. 1→2→3→4→5; an already-current world is
  refused outright). For a v1-v3 source it archives the
  original database under a name keyed to the source format (`world.v1.db`,
  `world.v2.db`, or `world.v3.db`) and prints a human summary (seed, villagers
  carried, continuation
  tick, source event count, archive path, and the `start` command to run next). A
  v4 source instead takes the spec-068 manifest-only path (`MigrateResult.
  ManifestOnly`, [[world-migration]]) — no archive, no transform, since nothing
  about the world's log, state, or terrain changes — and `cmdMigrate` prints a
  distinct summary naming it a manifest-only upgrade whose event log and terrain
  carry over unchanged.
- `ps [--all] [--json]` — machine-wide listing of worlds with live-proven state
  ([[instance-manager]]): discovery over the worlds home + registry, concurrent
  bounded probes, `NAME STATE PID TICK GAME TIME SPEED LLM PATH` table or a JSON
  array reusing the `status --json` vocabulary. Default shows live-pid states
  (`running`/`paused`/`unresponsive`); `--all` adds `stopped`/`missing`/`unreadable`.
  Empty listing prints "no worlds running", exit 0.
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

## Connections

`new`: [[world-tuning]], [[agent-mind]], [[guardian]], [[social-fabric]],
[[llm-orchestrator]], [[daemon-lifecycle]], [[cognition]], [[curriculum-ladder]],
[[scenario-machinery]]. `migrate`: [[world-migration]], [[world-save-directory]].
`ps`: [[instance-manager]]. `stages`: unlocks record + [[skin]] stage table,
shared with [[grounded-feedback]]. See [[cli-promptworld]] for exit discipline,
arg resolution, and siblings ([[cli-runtime-control]], [[cli-guardian-ops]]).
