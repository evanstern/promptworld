---
name: ipc-protocol
description: The wire contract — JSON-lines over a Unix socket; Request/Response/Push envelopes and the shared StatusData shape
kind: concept
sources:
  - internal/ipc/protocol.go
  - specs/001-world-daemon/contracts/client-protocol.md
verified_against: d23fbbfe471ec62c9b94ce79404870632a6eb60e
---

# IPC protocol

Clients talk to a world's daemon over the Unix socket in its save directory, one JSON
object per newline-delimited line. The Go types in `internal/ipc/protocol.go` are the
single source for both sides of the wire; the prose contract lives in
`specs/001-world-daemon/contracts/client-protocol.md`.

## How it works

Three envelopes:

- `Request{id, cmd, args}` — client → daemon; `id` is client-chosen and echoed back.
- `Response{id, ok, data | error}` — daemon → client.
- `Push{push, event | last_seq}` — daemon → subscribed client: `push: "event"`
  carries a `store.Event`; `push: "dropped"` carries `last_seq` and means the
  subscription was canceled on buffer overflow (re-subscribe with `since: last_seq`).

Clients demux on the presence of `id` vs `push` (`wireMsg` is the union used by the
client reader). Responses and pushes may interleave.

Commands: `status`, `state` (returns `StateData{state, last_seq}` — the full
canonical world-state JSON plus the log position it reflects, captured coherently in
one loop iteration; subscribe with `since: last_seq` for a gapless live replica),
`subscribe` (`SubscribeArgs{since}` — replay after that seq, then live, gapless),
`unsubscribe`, `pause`, `resume`, `set_speed` (`SetSpeedArgs{speed}`), `llm_call`
(`LLMCallArgs{kind, system, prompt, max_tokens}` → an `llm.Response` with tier,
model, tokens, cost, latency — errors when the world has no orchestrator), and
`shutdown`, and the Metatron console pair (TASK-12, [[metatron]]): `metatron_chat`
(`MetatronChatArgs{text}` → a `metatron.TurnResult` with reply, optional landed
nudge, charge bank, surfaced moments — a long call, one cloud round-trip) and
`metatron_status` (no args → the model-free `metatron.Status` peek), and `miracle`
(spec 016, [[metatron-miracles]]): `MiracleArgs{kind, day?, time?, villager?,
item?, qty?, class?, x?, y?, to_x?, to_y?, gratis?}` where `kind` selects
`time_snap`/`give_item`/`move`/`remove` and the remaining fields are that kind's
arguments → `MiracleData{kind, charges, gratis, summary}`. `miracle` is the
**only** surface that accepts `gratis` (the CLI's `--force` sets it, waiving the
charge — the angel's turn path has no equivalent field); the handler needs only
the sim loop, no LLM/angel presence, so it works in a pure-sim world. `StatusData`
gains an optional `llm` section (tier health, queue depths,
monthly spend vs budget) when the orchestrator is enabled.

`StatusData` is the shared response shape for status/pause/resume/set_speed, with four
sections: `world` (name, seed, format_version), `clock` (tick, game_time, paused,
speed, effective_rate, degraded, metatron_charges — the ⚡ bank, so clients need no
state fetch — plus, since spec 028, three additive `omitempty` adaptive-throttle
fields: `requested_speed` (the player's ceiling from sim state, empty when
ungoverned), `governor_debt`/`governor_jobs` (the daemon governor sampler's latest
staleness-debt reading, folded in exactly like the `llm` section — [[cognition]],
[[daemon-lifecycle]]); all three are zero/absent for a no-LLM world or an inert
governor, so pre-028 status bytes are unchanged), `daemon` (pid, uptime_seconds,
subscribers), `log`
(last_seq), and — since spec 035 — a top-level `Warning` string
(`json:"warning,omitempty"`) set ONLY on the `set_speed` reply (FR-002, FR-008):
the requested speed lands on a notch where a bootstrap-seeded provider's
watched cognition class is suppressed under its current estimate
([[ipc-server]] composes it via [[cognition]]'s `SuppressedAt`). `status`,
`pause`, and `resume` never set it, so their bytes stay unchanged; the
warning is purely advisory — the speed change has already applied by the
time it's set. `set_speed`'s existing refusal of uncapped `max` while an LLM is
configured is retained unchanged (spec 028 FR-012) — the governor only ever
moves `speed`/`effective_rate` along the capped ladder these fields describe,
never `max`.

Since spec 037 (`contracts/status-horizon.md`), `StatusData` also gains an
additive `omitempty` `Horizon []HorizonClass` — unlike `Warning`, this rides
`status`/`pause`/`resume`/`set_speed` alike (any world with an orchestrator,
composed in [[ipc-server]]'s `statusDataFull`), one entry per watched class
INCLUDED at the loop's CURRENT effective speed, never an empty slice (either
absent for a no-LLM world or ≥1 entry). `HorizonClass{Class, Suppressed,
Verdict, Calibrated, SuppressedCount}` carries the class name, whether it is
suppressed right now, [[cognition]]'s `Verdict.Arithmetic` string verbatim
(clients render it, never parse it), whether its serving provider is
calibrated (calibrated classes ARE included here — contrast the `Warning`
field above, which stays gated to bootstrap-seeded providers), and the
daemon-lifetime count of router suppressions [[llm-orchestrator]] has
recorded for that class.

Line caps (TASK-19): request lines are capped at 1 MiB, reply/push lines at
64 MiB. The daemon never emits a line over the cap — a reply that would exceed
it is substituted with an `ok:false` response whose `error` starts with
`reply too large` (carrying the byte counts). Clients classify that prefix —
and any raw over-long line — as `ipc.ErrReplyTooLarge`, a fatal error retrying
cannot fix.

Failure semantics: unknown cmd or bad args → `ok:false`, connection stays open;
malformed JSON → connection closed; daemon absent → socket connect fails fast;
oversized reply → the substituted `reply too large` error above.

## Connections

[[ipc-server]] implements the daemon side; [[ipc-client]] the attach side;
[[event-types]] defines what rides inside event pushes; [[cli-promptworld]] renders
`StatusData` for humans. The [[tui-client]] consumes `state` + `subscribe` to run its
live replica. `miracle` is the CLI/IPC operator door into [[metatron-miracles]].

## Operational notes

No authentication — trusted single-operator host, filesystem permissions on the socket
are the boundary. The protocol is debuggable with `nc -U <dir>/daemon.sock` and raw
JSON lines. Changing any field name here is a breaking wire change and must update the
contract doc in the same commit.
