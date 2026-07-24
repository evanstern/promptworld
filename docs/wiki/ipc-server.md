---
name: ipc-server
description: Daemon-side sessions — gapless subscribe with store gap-fill, bounded push buffers that drop subscriptions rather than block the loop, long-path socket workaround
kind: component
sources:
  - internal/ipc/server.go
  - internal/ipc/socket.go
verified_against: 1e71b77f104dda982aa407b28ad2c994219e90d0
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
dispatch to the optional angel through the `Angel` interface (`SetMetatron`,
[[metatron]]) — same posture: a long console turn occupies only its session, and
worlds without an LLM config answer with a clean "not present" error. `set_speed` enforces the speed
policy (TASK-20): `max` is refused with an actionable error whenever the world
has an LLM configured (`llm != nil`) — uncapped ticking is for pure-sim worlds;
the watchable ceiling is 32x ([[game-clock]]); the spec 028 governor changes
nothing here — it never touches uncapped speed, so this refusal is unchanged.
Since spec 035 (FR-002), a `set_speed` reply that clears the `max` refusal
additionally carries `uncalibratedWarning(speed)` on `StatusData.Warning`:
non-empty only when the world has an orchestrator and the requested speed's
ticks/sec would suppress one or more of [[cognition]]'s watched classes
(`SuppressedAt`) evaluated at their CURRENT live estimates, gated to classes
whose serving provider is still bootstrap-seeded (`s.llm.CalibratedAt(name)
== ""` — a calibrated provider's live drift is the governor's signal, not
this one). The warning never blocks the speed change — it composes after the
`max` gate, which is unchanged and still evaluated first — and `status`,
`pause`, `resume` always pass `""` so their replies stay byte-identical.
`replyStatus(id, cmd, speed, warning)` carries the extra parameter through to
every caller; only the `set_speed` case supplies a non-empty value — since
spec 039 that value is `setSpeedWarning(speed)`, not `uncalibratedWarning`
directly.

Since spec 039 (US2/US4, `contracts/posture.md`), a teaching world
(`s.w.Manifest.Teaching`) layers a second, independent advisory on top:
`postureWarning(speed)` fires when the requested speed exceeds the
planner-safe posture — computed live via [[cognition]]'s `MaxSafeSpeed` over
the planner-serving provider's `EstimateForKind` estimate, the SAME call the
daemon's boot default uses ([[daemon-lifecycle]]) — and composes, for every
watched class `LiveHorizon` finds suppressed at the requested speed, the
router's `Verdict.Arithmetic` string verbatim plus a plain-language
`postureConsequence` (`"villagers will stop deep-thinking (reflex only)"` for
`planner`, `"conversations will be skipped"` for `conversation`, `"meetings
fall back to template speeches"` for `meeting`, a generic degrade phrase
otherwise), joined under an `above teaching posture Nx:` prefix. Unlike
`uncalibratedWarning`, this fires for a CALIBRATED teaching world too — that
is the point of a soft cap. `setSpeedWarning(speed)` composes the two,
posture first then uncalibrated, newline-joined when both fire (either may be
empty), and is what the `set_speed` handler now calls instead of
`uncalibratedWarning` directly. `postureStatus()` computes the teaching
world's `*PostureStatus{Rung, Calibrated}` the status-family reply carries
(`StatusData.Posture`, [[ipc-protocol]]) via the identical `MaxSafeSpeed` call
— nil when no provider serves the planner class, so the field stays absent
rather than reporting an ungrounded rung. `statusDataFull` sets it only when
`s.w.Manifest.Teaching`, alongside the existing `Horizon` assignment.

`statusData`/`statusDataFull` fold two optional per-world snapshots into the
`clock`/`llm` sections the same way: the orchestrator's `StatusSnapshot` when
`SetLLM` attached one, and — since spec 028 — the daemon governor's debt
reading through a local `Governor` interface (`GovernorStatus() (debt float64,
jobs int)`, `SetGovernor`, kept narrow like `Angel` so `ipc` never imports
`internal/daemon`); a nil governor (no-LLM world) leaves the clock section's
governor fields at their `omitempty` zero, byte-identical to pre-028
([[cognition]], [[daemon-lifecycle]]).

Since spec 037 (`contracts/status-horizon.md`), `statusDataFull` additionally
sets `StatusData.Horizon` via `horizonClasses(cs)` whenever an orchestrator is
attached — so `status`/`pause`/`resume`/`set_speed` all carry it alike, unlike
the `set_speed`-only `Warning`. `horizonClasses` delegates to
[[cognition]]'s `LiveHorizon` at the loop's EFFECTIVE speed
(`cs.Speed.TicksPerSecond()`, post-governor), resolving each watched class's
live estimate through `EstimateForKind` (a class whose kind has no admissible
serving provider is excluded, `ok=false`), and folds in
`s.llm.CalibratedAt(name) != ""` for the `Calibrated` flag and
`s.llm.SuppressionCounts()` for each entry's `SuppressedCount` — a class never
suppressed reads 0 from the map's zero value. Unlike `uncalibratedWarning`
below, which gates OUT calibrated classes, `horizonClasses` INCLUDES them
(research R4): calibration only changes the client's remedy phrasing, never
membership. Returns nil (never an empty slice) when nothing is included, so
`omitempty` keeps the field absent for a no-LLM world.

`miracle` (spec 016, [[metatron-miracles]]) dispatches to `handleMiracle`, which
needs only `srv.loop` — never `srv.llm` or `srv.metatron` — so it works on
pure-sim worlds with no angel or orchestrator configured. It fetches the current
state via `loop.DoState` (to resolve door-side name/tile lookups: a `give_item`
villager name through `sim.AgentIndexByName`, a `time_snap` day/`HH:MM` through
`clock.ParseTimeOfDay`/`clock.TickAt`), builds `metatron.MiracleParams` from the
kind-specific args, calls the shared `metatron.BuildMiracleBatch` (the same
batch-builder the angel's turn uses) to compose the miracle event plus its
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
`handleMiracle` is one of the two doors into [[metatron-miracles]] (the other is
the angel's turn reply, [[metatron]]). `uncalibratedWarning` reads
[[cognition]]'s `SuppressedAt` and [[llm-orchestrator]]'s `EstimateForKind`/
`CalibratedAt`, and its result rides `StatusData.Warning` ([[ipc-protocol]]),
rendered by [[cli-promptworld]]'s `setSpeedLine`. `horizonClasses` reads
[[cognition]]'s `LiveHorizon` and [[llm-orchestrator]]'s `EstimateForKind`/
`CalibratedAt`/`SuppressionCounts`, and its result rides
`StatusData.Horizon` ([[ipc-protocol]]), rendered by [[cli-promptworld]]'s
`horizonStatusLines` and [[tui-client]]'s header badge + metatron-pane
`horizonLines`. Since spec 039, `postureWarning`/`postureStatus` read
[[cognition]]'s `MaxSafeSpeed` and [[game-clock]]'s `SpeedForRate` (the same
call [[daemon-lifecycle]]'s boot default makes), and their results ride
`StatusData.Warning`/`StatusData.Posture` ([[ipc-protocol]]), rendered by
[[cli-promptworld]]'s `setSpeedLine`/`postureStatusLine`.

## Operational notes

Writes carry a 10 s deadline; a dead client's connection is closed and its reader
unwinds. Multiple concurrent clients are allowed and equal. Subscriber count in status
counts subscribed sessions only, not mere connections.
