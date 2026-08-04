---
name: sim-state-intent-lifecycle
description: sim.State.Apply's intent-ring closure arms (build_failed, intent_failed, spec-043/062/064 stampIntentOutcome, recovery_stalled) plus the hail family and agent.died's spill/Deaths/grave/run.ended effects
kind: component
sources:
  - internal/sim/state.go
  - internal/sim/agents.go
verified_against: 5761edb18e2b5fb49c6a03a050b0d871f5546c05
---

# Sim state: intent-ring & lifecycle-end arms

Split from [[sim-state-reducer]] (summary-style, corpus-spec v2): the arms
that close an `IntentRecord` honestly — `agent.build_failed`, its spec-096
generalization `agent.intent_failed`, the spec-043 completion/rejection
stamps, the spec-062 reflex yield-window anchor, and the spec-064
needs-conditioned `recovery_stalled` abort — plus the hail family and the
agent-death effects that end a life: inventory spill, the `Deaths` ledger,
grave placement, and the `run.ended` terminal latch.

`agent.build_failed` (spec 038, `BuildFailedPayload{agent, goal, reason}`) is a
NEW arm with a state effect identical to `agent.intent_done` — `Intent = nil`,
`IdleSince` stamped — the executor emits it, instead of the bare completion
type, whenever a `build_*` goal's mid-work re-validation genuinely fails (site
gone, or a wall's reserved-tile occupant outlasting the grace period); the
reducer itself carries no build-specific logic, it only clears the intent the
same way completion does, so no material is spent and no structure stands
([[executor]], [[event-types]]); since spec 043 US1 both completion arms also
close the ring: `agent.intent_done` stamps the newest still-open
`IntentRecord` `"done"` and `agent.build_failed` stamps it `"failed"` (via
`stampIntentOutcome` — the newest open record IS the current intent; an older
record left open by an override stays open, the open-then-superseded shape
the alternation view preserves), while `agent.intent_rejected` — formerly a
pure telemetry no-op — now appends an ALREADY-CLOSED `"rejected"` record
(source `planner`), so the next thought can see an attempt was refused before
ever landing; since spec 062 US1 ([[reflex-policy]]) `stampIntentOutcome`
additionally returns the closed record's `Source` and whether one was open,
and `agent.intent_done`'s arm arms `Agent.LastMindIntentDone = e.Tick`
whenever that source `isMindSource` (`planner`/`plan`/`meeting`) — the reflex
PREP gate's yield-window anchor; a `agent.build_failed` closure (`"failed"`,
never a completion) does not arm it, and neither does a reflex-sourced
`"done"` closure (`isMindSource("reflex")` is false), so a no-planner world's
anchor stays the permanent 0 sentinel;
since spec 096 ([[executor]], [[event-types-agent-intents]]) a NEW arm,
`agent.intent_failed` (`IntentFailedPayload{agent, goal, reason, x, y}`),
generalizes `agent.build_failed`'s pattern to every non-build goal's
invalid-exit/contested resolution (the card's enumerated `forage`/`chop`/
`hunt`/`demolish`/`repair`/`quarry`/`cook`/`bathe` mid-work exits and
`craft_*`/`cook`/`bathe`/`deposit`/`withdraw` completion-time no-op
rechecks) — a byte-for-byte IDENTICAL state effect to `agent.build_failed`'s
(`Intent = nil`, `IdleSince` stamped, `stampIntentOutcome("failed", …)`, never
arming `LastMindIntentDone`), so the reducer gained no new logic beyond one
more `case` calling the same helpers; the only new axis is the payload's own
`x`/`y` (the acting agent's stand tile), since intent_failed spans goals with
no single shared Target-vs-Res addressing convention the way builds do;
since spec 064 ([[executor]], [[reflex-policy]]) `agent.intent_set` carries an
OPTIONAL completion condition onto the intent (described in
[[sim-state-apply-agents]], as is its spec-083 neglect class-intent stamp —
the intent-SET side of the lifecycle feeds the neglect detector's
zero-intent clock; closures never touch it), `agent.work_started`
gains a companion `Ref` field (`omitempty`) — a conditioned hold's work-started
doubles as its hold anchor, so `Ref` captures the need level the per-tick
no-net-gain check baselines against (0 and unread for every ordinary work
goal, byte-inert); and a NEW arm, `agent.recovery_stalled`
(`RecoveryStalledPayload{agent, goal, need}`), mirrors `agent.build_failed`'s
state effect (`Intent = nil`, `IdleSince` stamped) for a needs-conditioned
hold whose need showed no net gain across a full `recoveryStallTicks` window
(dead fire, displaced source, unreachable threshold) — an honest abort, not a
completion, so `stampIntentOutcome` closes the ring `"stalled"` rather than
`"done"` and, like `build_failed`, never arms `LastMindIntentDone` (the
build_failed precedent: an abort is not intelligence completing);

expired before ever firing). The hail family (TASK-47) maintains `Agent.Hail *AgentHail`
(`{By, Until}`, `omitempty` so pre-TASK-47 snapshots and un-hailed agents stay
byte-stable): `social.hailed` sets it, `social.hail_met`/`social.hail_expired`
clear it, and `agent.died`/`agent.slept` also clear it (the dead and the
sleeping shed hails). `agent.died` also spills the dying agent's entire carried
`Inv` onto a pile at its own tile (create-or-merge, food batches stamped
`tick + rotWindowTicks`), emptying `Inv` — reducer-internal, no new event (spec
013 US2, FR-006, research R7's debt-opening precedent) — and, since spec 044,
two more reducer-internal effects in the same arm: the death is appended to the
`State.Deaths` ledger (`{agent, tick, cause}` — cause now includes `"gru"`, the
[[gru]]'s escalated kill), and a `Structure{Kind: "grave"}` is placed at the
death tile (US4, FR-017, research R10). The grave placement is deliberately
unconditional — no dedup against whatever already occupies the tile: the
`Structures` slice has no per-tile uniqueness invariant outside the `buildSite`
gate governing NEW player-directed builds, and `structureAt` filters by kind,
so coexisting entries are an established pattern; appended last, the grave
also wins the TUI's last-write-wins per-tile glyph, and it blocks future
`buildSite` on the tile via the blanket any-structure check. The `run.ended`
arm (spec 044 US1) is the terminal latch: it sets `State.Ended` — which no
event ever clears, so replay/restart lands back in the ended posture and
migration tooling cannot resurrect a finished run without rewriting history —
and copies the payload verbatim onto `State.RunEnd` ([[sim-loop]] holds the

## Connections

Back to [[sim-state-reducer]] and its other five split-off notes.
[[decision-context]] consumes the `IntentLog` ring this closes;
[[reflex-policy]] consumes `LastMindIntentDone` and the PREP gate's
conditioned-hold semantics; [[executor]] emits `build_failed`,
`intent_failed`, and the recovery-stall class; [[morgue]] covers the death
ledger/grave/epilogue
mechanism these effects feed; [[sim-loop]] holds the matching `run.ended`
posture. [[sim-state-apply-agents]] is where `agent.intent_set`'s
completion condition is set.
