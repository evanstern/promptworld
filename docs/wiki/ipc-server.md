---
name: ipc-server
description: Daemon-side sessions — dispatch, the miracle door, broadcast/gapless delivery, reply ceiling, long-path sockets; StatusData signal composition split to ipc-status-signals
kind: component
sources:
  - internal/ipc/server.go
  - internal/ipc/socket.go
verified_against: 657c770f87404b936a0587db1f6b00e81b9f0ee6
---

# IPC server

`ipc.Server` hosts the protocol for one world. Its governing invariant is FR-011:
session lifecycles are fully decoupled from the sim — a client can die mid-write,
spam garbage, or subscribe and stall, and the loop never notices.

## How it works

`Serve` accepts connections; each `session` runs its own reader goroutine with a
line-scanner capped at `maxRequestBytes` (1 MiB — requests are small). Malformed
JSON closes that connection; unknown commands return `ok:false` and keep it open. Time-control and status commands go
through `Loop.Do` and reply with the full `StatusData` (built by `statusData`, which
adds world/daemon/log sections around the loop's clock snapshot); `state` goes
through `Loop.DoState` and replies with `StateData` — the canonical world-state JSON
plus the `last_seq` it reflects. `llm_call` submits to the optional
[[llm-orchestrator]] (`SetLLM`; 2-minute timeout per call) — a slow or dead model
blocks only the calling session, never the loop; `statusDataFull` appends the
orchestrator's snapshot to status responses. `metatron_chat`/`metatron_status`
(frozen wire command names, spec 052 ruling 2) dispatch to the optional
guardian through the `Guardian` interface (`SetGuardian`,
[[guardian]]) — same posture: a long console turn occupies only its session, and
worlds without an LLM config answer with a clean "not present" error. Since
spec 052 (FR-003), `SetSkin` attaches the world's boot-frozen display skin
the same singleton-attachment way (`SetLoop`/`SetLLM`'s precedent) — nil-safe,
so a no-skin world (and every pre-052 test) needs no call at all; `statusData`
composes the resolved identity fields (`SkinName`/`SkinEpithet`/
`SkinTabLabel`/`SkinFamilyLabel`) plus any override maps (`SkinStrings`/
`SkinStages`) into every `StatusData` reply ([[ipc-protocol]]).

The optional signals folded into `StatusData` from there — `set_speed`'s
uncalibrated/teaching-posture warnings, the governor/morgue/stage/scenario
snapshot folds, and `horizonClasses` — split out to
[[ipc-status-signals]] (which links back here).

`miracle` (spec 016, [[guardian-miracles]]) dispatches to `handleMiracle`, which
needs only `srv.loop` — never `srv.llm` or `srv.guardian` — so it works on
pure-sim worlds with no guardian or orchestrator configured. It fetches the current
state via `loop.DoState` (to resolve door-side name/tile lookups: a `give_item`
villager name through `sim.AgentIndexByName`, a `time_snap` day/`HH:MM` through
`clock.ParseTimeOfDay`/`clock.TickAt`), builds `guardian.MiracleParams` from the
kind-specific args, calls the shared `guardian.BuildMiracleBatch` (the same
batch-builder the guardian's turn uses) to compose the miracle event plus its
perception-memory companions, and lands it through `loop.InjectSocial` — so the
door is validated by the exact same dry-run/reducer path as every other injected
batch. Replies with the post-land charge bank and a one-line summary
(`MiracleData`); an invalid kind, unresolvable name/tile, or reducer rejection
(insufficient charges, bad destination, …) returns `ok:false` and nothing lands.

**Broadcast path**: the loop's notify callback is `Server.Broadcast`, which offers
committed events to each session under a non-blocking send into a
`pushBufferSize = 1024` channel. On overflow the subscription is canceled and a
`{"push":"dropped","last_seq":N}` is sent from a fresh goroutine — the loop is never
blocked by a slow client.

**Gapless delivery**: each subscription runs a pusher goroutine with a `cursor`. It
first fills from the store up to the log head at subscribe time (`subscribe{since}`
replay), then consumes the live channel; any seq jump ahead of `cursor+1` triggers a
store gap-fill (`EventsSince`) before delivery. This closes the race between opening
the live buffer and reading history — events are always delivered in seq order with
no gaps for the life of a subscription.

**Reply ceiling** (TASK-19): server→client lines are bounded by
`maxReplyBytes` (64 MiB), split from the request cap because the `state` reply
carries the whole world state on one line and outgrew the old shared 1 MiB
`maxLineBytes` on long runs. `writeResponse` never emits a longer line: an
oversized reply is replaced by an `ok:false` error on the same ID whose message
starts with `replyTooLargePrefix` ("reply too large") and names the byte
counts — the [[ipc-client]] classifies that prefix as fatal
(`ErrReplyTooLarge`) instead of retrying. `writePush` drops an over-cap push
outright (a single event cannot realistically hit the cap); both funnel into
`writeLine` for the deadline-guarded write.

**Long socket paths** (`socket.go`): `sockaddr_un` caps paths (~104 bytes on darwin).
`listenUnix`/`dialUnix` transparently chdir into the socket's directory and use its
basename when the path exceeds `maxSockPath = 100`, serialized under a mutex with cwd
restored immediately — save directories can live at any depth.

`shutdown` replies ok, then invokes the daemon's cancel function. `Close` unwinds the
listener and every session and removes the socket file.

## Connections

[[sim-loop]] feeds `Broadcast` and receives `Do` calls; [[event-log]] backs replay and
gap-fill; [[ipc-protocol]] defines the wire shapes; [[daemon-lifecycle]] constructs the
server (with `SetLoop` breaking the mutual reference) and calls `Close` on exit;
`handleMiracle` is one of the two doors into [[guardian-miracles]] (the other is
the guardian's turn reply, [[guardian]]). See [[ipc-status-signals]] for the
warning/posture/horizon-fold connections.

## Operational notes

Writes carry a 10 s deadline; a dead client's connection is closed and its reader
unwinds. Multiple concurrent clients are allowed and equal. Subscriber count in status
counts subscribed sessions only, not mere connections.
