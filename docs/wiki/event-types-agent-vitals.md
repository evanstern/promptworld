---
name: event-types-agent-vitals
description: Agent vitals/mortality event rows split from [[event-types]]: agent.needs_changed/died, run.ended, agent.slept/woke. Load when tracing needs decay/anchoring, death causes (starvation/exposure/collapse/gru), the death ledger and grave placement, or run-ending/postmortem posture.
kind: concept
sources:
  - internal/sim/executor.go
  - internal/sim/state.go
  - internal/sim/gru.go
  - internal/sim/morgue.go
verified_against: 0fd2104c59c54be8e8071d319fa4ce192083faf3
---

# Event types — agent vitals & mortality

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.

Spec 044 (run outcomes, morgue, gru escalation, graves — [[morgue]]) is also
format-stable: `State` gains `omitempty` `Deaths` (the run's death ledger),
`Ended`/`RunEnd` (the terminal run-over latch and summary),
`CharterFingerprint`, and the bounded `MorgueEpilogues` ring, and `Structure`
gains a `grave` kind (reducer-placed at a death tile, never player-built).
THREE new event types land the feature — executor-emitted `run.ended` (never
injectable: it opens two whole new namespaces, `run.*` and, via the
whitelisted pair, `morgue.*`) plus whitelisted `metatron.charter_observed` and
`morgue.epilogue`, the only two additions `TestWhitelistDiffIdentical` accepts
against the spec's declared boundary widening. Two existing rows changed
semantics rather than shape: `agent.died`'s cause domain gains `"gru"` (an
escalated kill emitted by `gruStep` itself) and its reducer arm gains the
ledger append + grave placement, and `gru.attacked`'s `health` may now be 0
when the target was already weakened below `nearDeathBelow` ([[gru]]).

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `agent.needs_changed` | `NeedsPayload{agent, …}` | per-game-minute heartbeat | needs set to absolute values; spec 043 US2: once `tick − NeedsAnchorTick ≥ trajectoryWindowTicks` (1800) the arm snapshots the current needs into `Agent.NeedsAnchor`/`NeedsAnchorTick` — the window-edge anchor the decision prompt diffs against for rising/falling/steady arrows ([[decision-context]]); anchor unset (tick 0) until the first full window elapses |
| `agent.died` | `DiedPayload{agent, cause}` — cause ∈ `starvation`\|`exposure`\|`collapse`\|`gru` (spec 044 US3) | heartbeat at 0 health, **or** `gruStep` (spec 044 US3) on an escalated gru kill — emitted immediately after that attack's `gru.attacked`, cause `"gru"`, with an inline witness-death memory loop (gru attacks land off the %60 needs heartbeat, so the executor's own witness-death block never runs for them) | `Dead`, intent cleared; spec 013 (US2, FR-006, research R7): the agent's entire carried inventory spills into a pile at the death tile (created/merged, food batches stamped `tick + rotWindowTicks`), emptying `Inv` — reducer-internal, no new event; spec 044 US1: the death is also appended to the `State.Deaths` ledger (`{agent, tick, cause}`, application = event order) so the run-end declaration can carry the run's full death history without a log scan; spec 044 US4 (FR-017): a `Structure{Kind:"grave"}` is placed at the death tile — unconditional (no dedup against whatever else already occupies the tile; structures already coexist per-tile by kind) — visible on the map ([[tui-client]]), knowable to villagers ([[agent-mind]] known-places), and blocking future `buildSite` on that tile (deliberate, research R10) |
| `run.ended` (spec 044 US1) | `RunEndedPayload{tick, deaths, final_cause}` — `deaths` is the whole run's ledger (`{agent, tick, cause}`, event order) | executor (`stepEvents`), in the same batch as the run's final `agent.died`, ordered after every same-tick death and its witness memories, guarded by `!State.Ended` — exactly once per world, ever; never injectable (no whitelist entries) | sets `State.Ended` (terminal latch — no event clears it, so restart/replay lands back ended) and `State.RunEnd`; the loop idles in the paused-mode posture (mutating commands refused, reads served), status surfaces gain `ended`/`ended_day`, and the TUI enters postmortem posture ([[tui-client]], [[sim-loop]]) |
| `agent.slept` / `agent.woke` | `AgentPayload{agent}` | executor | sleep flag (slept clears intent) |
