---
name: event-types-clock-world
description: Clock/scheduler and world-lifecycle event rows split from [[event-types]]: world.created/migrated/forked, clock.paused/resumed/speed_set/governor_shed/governor_recovered/degraded/recovered, sim.day_started/night_started/forage_regrown, daemon.started/stopped/llm_warning. Load when tracing pause/resume/speed/governor mechanics, world genesis/migration/forking, or daemon lifecycle and provider-health events.
kind: concept
sources:
  - internal/sim/state.go
  - internal/sim/loop.go
  - internal/daemon/daemon.go
verified_against: fc1a8314f3f71a33c5e2145c914d5cbb511d9196
---

# Event types — clock & world events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.


Spec 086 (agent-named payloads): every agent-referencing field in this
family's payloads is a `sim.AgentRef` — the wire carries
`{"id":N,"name":"…"}` objects (lists element-wise), the name stamped at
emission from the fixed roster via `Ref`/`Refs`; sentinels marshal
`{"id":-1,"name":""}`. Legacy bare-int rows decode through the dual-shape
unmarshal forever and reducer arms fold `.ID`s only — the conventions and
the normative back-compat matrix live on [[event-types]] ("Agent
references are named refs"). `world.migrated` is deliberately untouched: its
payload embeds the full canonical `sim.State`, and no `AgentRef` is ever
reachable from `State` (`TestNoAgentRefInState`), so migrated worlds'
shapes and hashes are byte-identical.
Spec 028 (adaptive throttle) likewise adds **no** format bump: `State` gains
`RequestedSpeed` (`omitempty` — absent means ungoverned, so every pre-028
snapshot is a valid ungoverned state), and two new reducer-applied types,
`clock.governor_shed`/`clock.governor_recovered`, land the governor's
speed-ladder decisions ([[cognition]]).

Spec 034 (provider health conditions + preflight — [[llm-provider-health]]) is
also format-stable: one new whitelisted type, `daemon.llm_warning`
(`LLMWarningPayload{provider, kind, detail, remedy, active}`), rides the
daemon's operator-event door ([[sim-loop]]'s `InjectOperator`) on every
provider-health transition (raise/reclassify/clear); it is operator-facing
only, alongside `daemon.started`/`stopped` under the existing `daemon.*`
no-op convention.

Spec 039 (teaching-world speed posture — [[daemon-lifecycle]], [[cognition]])
adds no new event type and no new emission door: on a teaching world with an
orchestrator, boot computes the planner-safe posture rung and applies it
through the loop's ordinary `set_speed` command, so it lands as an EXISTING
`clock.speed_set` (below) with an ordinary `SpeedSetPayload{speed}` — a new
emitter of that type, not a new type, which is why the format stays put.

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `world.created` | `WorldCreatedPayload{name, seed}` | CLI `new`, tick 0 | none (genesis marker) |
| `world.migrated` | `WorldMigratedPayload{from_format, source_events, source_tick, state}` (`state` embeds the full canonical `sim.State`) | `promptworld migrate` (client-side, offline — [[world-migration]]), once, right after a fresh `world.created` | replaces `State` wholesale (after checking `state.Seed` matches — a foreign payload is a no-op); the log alone (`world.created` → `world.migrated`) reproduces the migrated world with zero snapshots |
| `world.forked` (spec 076) | `WorldForkedPayload{parent_name, parent_seed, parent_created_at, fork_tick, fork_seq}` | `world.Fork` (client-side, offline — [[world-forking]]), exactly once per fork, at `(boundary.tick, boundary.seq+1)` right after the carried parent prefix | none (recorded-history no-op, the `world.created` posture) — the no-op is what keeps a fork's state at the fork tick byte-identical to its parent's; the manifest `lineage` block mirrors it for offline readers |
| `clock.paused` / `clock.resumed` | `{}` | loop command | pause flag (+ snapshot on pause) |
| `clock.speed_set` | `SpeedSetPayload{speed}` | loop command (player `set_speed`, or — since spec 039 — the daemon's teaching-posture default applying its computed rung through the same command at boot, [[daemon-lifecycle]]) | `Speed` updated; since spec 028 also clears `State.RequestedSpeed` — a player command always collapses governed state (FR-009) |
| `clock.governor_shed` / `clock.governor_recovered` (spec 028 FR-008) | `GovernorPayload{requested, from, to, debt, jobs}`, shared by both | the daemon's governor sampler via the loop's `govern` command ([[cognition]], [[daemon-lifecycle]]) | `Speed = to`; `RequestedSpeed = requested` (shed) or cleared when `to == requested` (recovered reaching the ceiling); `EffectiveRate` follows `to` unless `Degraded` — never silent, so an operator can reconstruct every governed interval from the log alone (SC-005) |
| `clock.degraded` / `clock.recovered` | `DegradedPayload{effective_rate}` / `{}` | loop auto-slow | degradation flags |
| `sim.day_started` / `sim.night_started` | `DayPayload{day}` | executor, 06:00/22:00 | `Night` flag only — waking is explicit |
| `sim.forage_regrown` | `RegrownPayload{x, y}` | executor, regrow tick | harvest overlay removed |
| `daemon.started` / `daemon.stopped` | `DaemonStartedPayload` / `DaemonStoppedPayload` | daemon lifecycle | none |
| `daemon.llm_warning` (spec 034) | `LLMWarningPayload{provider, kind, detail, remedy?, active}` | daemon condition hook, via [[sim-loop]]'s `InjectOperator` door, on a provider-health raise/reclassify/clear ([[llm-provider-health]]) | none — operator-facing only, same no-op class as `daemon.started`/`stopped` |

## Connections

[[sim-state-reducer]] applies these; the [[executor]], [[reflex-policy]], and
[[sim-loop]] emit the sim/agent/clock families — since spec 039,
[[daemon-lifecycle]]'s teaching-posture boot default is a second caller of the
loop's `set_speed` command that lands `clock.speed_set`, alongside the player;
`promptworld migrate`
([[cli-promptworld]], [[world-migration]]) emits `world.migrated`;
`promptworld fork` ([[world-forking]]) emits `world.forked`;

[[daemon-lifecycle]]
emits `daemon.*`;

[[llm-provider-health]]
emits `daemon.llm_warning` through [[sim-loop]]'s `InjectOperator` door.
