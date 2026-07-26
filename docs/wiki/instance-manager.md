---
name: instance-manager
description: World instance management — machine-wide ps with live-proven state, the default worlds home, name-or-path addressing, the advisory known-worlds registry
kind: component
sources:
  - internal/worlds/home.go
  - internal/worlds/registry.go
  - internal/worlds/resolve.go
  - internal/worlds/discover.go
  - internal/worlds/probe.go
  - cmd/promptworld/ps.go
verified_against: 0fd2104c59c54be8e8071d319fa4ce192083faf3
---

# Instance manager

`internal/worlds` is the client-side manager over many worlds (spec
`specs/008-instance-manager`, TASK-43): it gives `promptworld` docker/ollama ergonomics
— `ps` across the machine, creation by bare name, name-or-path addressing on every
per-world command — without weakening the grounded "one directory = one world, never
global" decision ([[design-grounding]], [[world-save-directory]]). Everything it keeps
outside a world directory is advisory and self-healing; a world runs with none of it
present.

## How it works

**Home** (`home.go`): the promptworld root is `~/.promptworld`, overridden by
`PROMPTWORLD_HOME`; derived paths are the worlds home `<root>/worlds` (where
`new <name>` creates) and the registry `<root>/known_worlds.json`. `ValidateName`
enforces creation-time name rules: non-empty, no `/`, no leading `-` or `.`.

**Registry** (`registry.go`): `known_worlds.json` is a `{"worlds":{name:{path}}}`
pointer cache for worlds living *outside* the worlds home (the home itself is
scan-owned). Reads tolerate a missing or corrupt file (⇒ empty, never an error); writes
are atomic (temp + rename) and prune entries whose directory no longer holds a readable
`world.json` or whose path is inside the current home. It is upserted from two places:
`daemon.Run` at boot ([[daemon-lifecycle]]) and `new --at`. Never authoritative for
liveness or content.

**Resolution** (`resolve.go`): an argument is a **path** iff it contains `/` or starts
with `.` or `~` — paths bypass the manager entirely and keep their historical behavior.
A bare **name** resolves worlds-home-first, then registry; a home/registry collision is
refused as ambiguous (both paths printed), an unknown name fails naming the home
searched and suggesting `ps --all`, and a registry entry whose directory vanished gets
a dedicated "last known at …, but that directory is gone" error (`ErrMissing`).
Candidates are classified by `probeWorld` rather than a plain open-succeeded check: a
directory whose `world.json` exists on disk but which `world.Open` rejects (e.g. an
unsupported `format_version`) is **unopenable**, not unknown — resolution returns
`ErrUnopenable` carrying the `world.Open` error verbatim (migrate hint included) instead
of falling through to not-found, in both directions (old binary vs newer-format world
and vice versa). An unopenable home candidate takes precedence over a healthy registry
candidate of the same name, matching home-first ordering.

**Discovery + probe** (`discover.go`, `probe.go`): `ps` candidates are the home scan ∪
registry entries. Each candidate is classified concurrently under a ~1s per-world
budget (`probeBudget`): a live pidfile pid (same `kill(pid, 0)` check as the daemon's,
duplicated locally to avoid a daemon↔worlds import cycle) is only a pre-filter — a
world is `running`/`paused` solely when its daemon answers a `status` round trip within
budget; a live pid that doesn't answer is `unresponsive`, never running. Dead-pid
candidates become `stopped` (last-known clock via `OfflineSnapshot`, the extracted
offline read `status` also uses), `unreadable` (corrupt manifest), or `missing`
(registry path gone). The state reconstruction itself is `OfflineState(w)`
(extracted by spec 076 so `status`/`ps` and `promptworld compare` share ONE
offline read — [[world-forking]]): the newest valid snapshot unmarshaled
into a fresh state, then — since spec 044 (FR-004) — any events newer than
that snapshot folded in (the snapshot cadence can trail the log, e.g. a
crash right after a final death), exactly as recovery replay would; under
WAL it is also safe against a RUNNING daemon's db (readers never block the
single writer — the read reflects the last committed batch), which is what
lets compare read live worlds. `OfflineSnapshot` wraps it, returning the
clock fields plus the two spec-044 run-outcome returns, `ended` and
`endedDay` (the game day of the run end via the state's `RunEnd`, 0 unless
ended), so a stopped ended world's `status` mirrors the live surface
([[morgue]]). `ps`'s own `fillOfflineSnapshot` discards the two run-outcome
returns — stopped rows render as before. `cmd/promptworld/ps.go` renders the table
(`NAME STATE PID TICK GAME TIME SPEED LLM PATH`) and the `--json` array reusing the
`status --json` vocabulary; default shows live-pid states, `--all` adds the rest;
inference on/off comes from `StatusData.LLM != nil` live, `llm.json` presence stopped.

## Connections

[[cli-promptworld]] threads every per-world command through `resolveWorld` and hosts
`ps`/`new`; [[daemon-lifecycle]] performs the boot-time registry upsert; probing rides
[[ipc-client]] `status` and [[world-save-directory]]'s pidfile/socket layout; offline
rows read [[snapshots]] + [[event-log]] via `OfflineSnapshot`;
[[world-forking]]'s `compare` is `OfflineState`'s second consumer.

## Operational notes

`PROMPTWORLD_HOME` moves home + registry together; all discovery and resolution honor
it. Registry write failures inside the daemon are logged and never block boot. `ps`
whole-listing time stays under ~2s regardless of wedged daemons (parallel probes). A
copied world directory needs no manager state anywhere — starting it (re)registers it.
