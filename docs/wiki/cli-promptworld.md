---
name: cli-promptworld
description: The single promptworld binary — subcommand dispatch table, exit discipline, per-world arg resolution; routes to cli-world-lifecycle, cli-runtime-control, cli-guardian-ops for the subcommand families
kind: component
sources:
  - cmd/promptworld/main.go
  - cmd/promptworld/commands.go
verified_against: 9f7df6137c78506f9d5ab48809f6c2e4855da782
---

# promptworld CLI

One binary serves every role: daemon, client tools, world management. `main.go` is a
plain dispatch table; behavior lives in `commands.go`, except `calibrate` in its own
`calibrate.go`, `ps` in `ps.go` ([[instance-manager]]), the guardian's operator
door in `work.go`, and (spec 076) `fork`/`compare` in `fork.go`/`compare.go`
([[world-forking]]). The prose contract is
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
`dirArg`/`parseDirFlags` with that resolution. `parseDirFlags` accepts both
`cmd <arg> --flag` and `cmd --flag <arg>` orderings (`parseWorldFlags` adds name
resolution on top).

This note keeps the dispatch table, exit discipline, and shared argument
resolution; the subcommands themselves split into three family notes by
substance (summary-style split — each retains its own full detail and links
back here):

- [[cli-world-lifecycle]] — `new`, `migrate`, `fork`, `compare`, `ps`,
  `stages`: creating a world (with its tuning/teaching/stage/scenario
  stamping), upgrading an older one through the v1→v5 chain, forking a
  stopped world at its latest snapshot and comparing two runs (the spec 076
  duel — [[world-forking]]), and enumerating what's on the machine.
- [[cli-runtime-control]] — `daemon`, `start`, `stop`, `status`, `pause`/`resume`/
  `speed`, `teaching`, `ui`, `attach`, `tail`: starting/stopping a daemon, reading
  and steering its live status (WARNING/horizon/posture/stage/exercise lines,
  offline reconstruction), and the interactive/streaming clients over the socket.
- [[cli-guardian-ops]] — `guardian`, `work`, `llm`, `calibrate`, `divergence`: the
  guardian's conversational one-shot and operator miracle door, one-shot model
  calls, cognition-horizon benchmarking, and offline embedding-memory gate
  evidence.

One registered subcommand sits outside those three families because its
audience is a developer, not a player: `frames` (spec 112, `frames.go`)
renders the TUI headlessly against a canned in-process fixture world — one
frame to stdout, the whole design matrix to `docs/design/tui/frames/`
(`--dump`), the catalog of fixtures/states/sizes (`--list`), or the real
interactive client on a fixture (`--interactive`). It reaches no daemon, no
LLM and no sim, and it is the one per-world-looking command that takes no
`<world>` argument at all — a fixture is a Go value, so `resolveWorld` is
never involved. See [[tui-client]].

## Connections

[[daemon-lifecycle]] is what `daemon`/`start` run; [[instance-manager]] owns name
resolution, discovery, and the `ps` probe; [[ipc-client]] carries every online
command; [[world-save-directory]] and [[event-log]] back the offline paths;
[[game-clock]] formats times in clock/event lines. See each family note above
for the connections specific to its subcommands.

## Operational notes

`start` failure says "check daemon.log". Detached daemons survive terminal close
(Setsid); a machine reboot needs a manual `start` (launchd integration is future
work — the foreground `daemon` subcommand is what a plist would run).
