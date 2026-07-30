---
name: event-types-agent-vitals
description: Agent vitals/mortality event rows split from [[event-types]]: agent.needs_changed/died, run.ended, agent.slept/woke, sim.neglect_detected (spec 083). Load when tracing needs decay/anchoring, death causes (starvation/exposure/collapse/gru), the death ledger and grave placement, the neglect percept, or run-ending/postmortem posture.
kind: concept
sources:
  - internal/sim/executor.go
  - internal/sim/state.go
  - internal/sim/gru.go
  - internal/sim/morgue.go
verified_against: 9b4ed5aef5bfea50b67fac10f8e2153f065a814d
---

# Event types — agent vitals & mortality

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.


Spec 086 (agent-named payloads): every agent-referencing field in this
family's payloads is a `sim.AgentRef` — the wire carries
`{"id":N,"name":"…"}` objects (lists element-wise), the name stamped at
emission from the fixed roster via `Ref`/`Refs`; sentinels marshal
`{"id":-1,"name":""}`. Legacy bare-int rows decode through the dual-shape
unmarshal forever and reducer arms fold `.ID`s only — the conventions and
the normative back-compat matrix live on [[event-types]] ("Agent
references are named refs"). `run.ended`'s death ledger rides the wire as
`DeathRef` mirrors (`Agent AgentRef`, same tags) while the state
`DeathRecord` ledger keeps bare ints — the R2 split (`internal/sim/state.go`).
Spec 044 (run outcomes, morgue, gru escalation, graves — [[morgue]]) is also
format-stable: `State` gains `omitempty` `Deaths` (the run's death ledger),
`Ended`/`RunEnd` (the terminal run-over latch and summary),
`CharterFingerprint`, and the bounded `MorgueEpilogues` ring, and `Structure`
gains a `grave` kind (reducer-placed at a death tile, never player-built).
THREE new event types land the feature — executor-emitted `run.ended` (never
injectable: it opens two whole new namespaces, `run.*` and, via the
whitelisted pair, `morgue.*`) plus whitelisted `guardian.charter_observed` and
`morgue.epilogue`, the only two additions `TestWhitelistDiffIdentical` accepts
against the spec's declared boundary widening. Two existing rows changed
semantics rather than shape: `agent.died`'s cause domain gains `"gru"` (an
escalated kill emitted by `gruStep` itself) and its reducer arm gains the
ledger append + grave placement, and `gru.attacked`'s `health` may now be 0
when the target was already weakened below `nearDeathBelow` ([[gru]]).

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `agent.needs_changed` | `NeedsPayload{agent, …}` | per-game-minute heartbeat on LEGACY worlds; under the spec-104 coalescing regime ([[world-tuning]]'s `needs_checkpoint_minutes`, ruling 3) only every K game-minutes (checkpoint grid) PLUS immediately on any danger-band/near-death/zero crossing vs the last-EMITTED values (both directions, so guardian watches fire AND re-arm at the per-minute latency; K=1 = the legacy cadence byte-for-byte) — non-emitted minutes decay derivedly behind the `NeedsSyncTick` watermark ([[sim-state-reducer]] advancement) with the SAME latch/anchor folds | needs set to absolute values; spec 104: while coalescing, the arm also stamps `NeedsSyncTick` (the double-decay guard — a recorded minute is never re-decayed) and `NeedsEmitted` (the crossing baseline); spec 043 US2: once `tick − NeedsAnchorTick ≥ trajectoryWindowTicks` (1800) the arm snapshots the current needs into `Agent.NeedsAnchor`/`NeedsAnchorTick` — the window-edge anchor the decision prompt diffs against for rising/falling/steady arrows ([[decision-context]]); anchor unset (tick 0) until the first full window elapses |
| `agent.died` | `DiedPayload{agent, cause}` — cause ∈ `starvation`\|`exposure`\|`collapse`\|`gru` (spec 044 US3) | heartbeat at 0 health, **or** `gruStep` (spec 044 US3) on an escalated gru kill — emitted immediately after that attack's `gru.attacked`, cause `"gru"`, with an inline witness-death memory loop (gru attacks land off the %60 needs heartbeat, so the executor's own witness-death block never runs for them) | `Dead`, intent cleared; spec 013 (US2, FR-006, research R7): the agent's entire carried inventory spills into a pile at the death tile (created/merged, food batches stamped `tick + rotWindowTicks`), emptying `Inv` — reducer-internal, no new event; spec 044 US1: the death is also appended to the `State.Deaths` ledger (`{agent, tick, cause}`, application = event order) so the run-end declaration can carry the run's full death history without a log scan; spec 044 US4 (FR-017): a `Structure{Kind:"grave"}` is placed at the death tile — unconditional (no dedup against whatever else already occupies the tile; structures already coexist per-tile by kind) — visible on the map ([[tui-client]]), knowable to villagers ([[agent-mind]] known-places), and blocking future `buildSite` on that tile (deliberate, research R10) |
| `run.ended` (spec 044 US1) | `RunEndedPayload{tick, deaths, final_cause}` — `deaths` is the whole run's ledger (`{agent, tick, cause}`, event order) | executor (`stepEvents`), in the same batch as the run's final `agent.died`, ordered after every same-tick death and its witness memories, guarded by `!State.Ended` — exactly once per world, ever; never injectable (no whitelist entries) | sets `State.Ended` (terminal latch — no event clears it, so restart/replay lands back ended) and `State.RunEnd`; the loop idles in the paused-mode posture (mutating commands refused, reads served), status surfaces gain `ended`/`ended_day`, and the TUI enters postmortem posture ([[tui-client]], [[sim-loop]]) |
| `agent.slept` / `agent.woke` | `AgentPayload{agent}` | executor | sleep flag (slept clears intent) |
| `sim.neglect_detected` (spec 083) | `NeglectDetectedPayload{agent, need, level, since}` — need ∈ `food`\|`warmth`\|`rest`, level = pre-tick value, since = band-entry tick | needs heartbeat sweep (`stepEvents` %60, `NeglectDue` — pure over pre-tick state: living AWAKE agent, need below its spec-062 danger band for `neglectWindowTicks` (7200) with zero class-goal intents (`needClassGoals`, [[reflex-policy]]) over the same window, episode latch clear); executor emission class, never injectable; a salience-9 `agent.memory_added` companion (fixed per-need voice-of-evidence text, `OriginWitness`, generation-bumping) rides the same batch immediately after | sets the need's one-per-episode fired latch on `Agent.Neglect` ([[sim-state-agent-fields]]); the latch and band anchor clear together when `agent.needs_changed` folds the need back to/above its band |
