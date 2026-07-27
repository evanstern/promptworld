---
name: sim-state-agent-fields
description: Per-agent field catalog on sim.State — clock, needs/intents/inventories/memories, Journal, mental-map pointer, IntentLog ring, NeedsAnchor trajectory, LastMindIntentDone reflex-yield anchor, Neglect detector substrate (spec 083)
kind: component
sources:
  - internal/sim/state.go
  - internal/sim/agents.go
  - internal/sim/journal.go
verified_against: fc9566d527941d3950fdd307168556820bd0875b
---

# Sim state: agent & clock fields

Split from [[sim-state-reducer]] (summary-style, corpus-spec v2): the clock
state and the per-agent field catalog — needs, intents, inventories,
memories, and the reducer-derived self-knowledge surfaces layered on top of
them release by release.

`sim.State` is the whole world in one struct: clock state (tick, paused, speed,
degraded, effective rate — plus, since spec 028, a `RequestedSpeed
clock.Speed` `omitempty` sitting beside `Speed`: the player's ceiling, present
only while the adaptive-throttle governor holds `Speed` below it; `Speed`
itself keeps its pre-028 meaning as the loop's pacing speed, now specifically
the EFFECTIVE speed, so the router and auto-slow observer need no change)
plus the living world — agents with needs/intents/
inventories (the v2 resource set, spec 012 — [[executor]]; spec 032 US2 adds
`Axes []int`, a `Spears` clone — remaining harvest uses per carried axe, sorted
ascending, tripling chop/quarry yield)/memories (with
`IdleSince` for the reflex grace, a `NearDeath`
latch, a `Generation` interrupt counter and pending `Plan` steps for the
[[cognition]] horizon — both `omitempty` so pre-TASK-32 snapshots stay
byte-stable — plus, since spec 019 US3, a self-authored `Journal *Journal`
(`journal.go`): a durable per-agent notebook mutated ONLY by the two `journal.*`
Apply arms; an `omitempty` pointer on the Hail precedent, so a never-journaling
agent stays byte-identical to a pre-019 snapshot; each `Memory` also now carries
`omitempty` situated context `Where`/`Why`/`Conv`/`Origin` (spec 030's
closed-vocabulary provenance class — the ONLY signal `DirectPerception`,
`memory.go`, reads to classify direct perception), byte-stable when absent,
and — since spec 042 — `omitempty` `Seq`/`Vec`/`VecModel`: `Seq` is the
emitting `agent.memory_added` event's store seq, copied onto the appended
`Memory` at Apply time as its stable identity (unique where `(agent, tick)`
is not) — live, [[sim-loop]]'s `stampSeqs` pre-assigns each batch's
contiguous seqs before `Apply` runs, since `AppendEvents` otherwise only
assigns them inside its own append transaction; on replay the events already
carry their seqs from the log, so both paths land the identical value;
`Vec`/`VecModel` are a recorded embedding attached verbatim by the new
`agent.memory_embedded` arm (see [[sim-state-cognition-arms]]), nil `Vec` meaning vectorless
([[memory-retrieval]]);
plus, since spec 041, a private spatial-knowledge store, `Map *MentalMap`
(`omitempty`, the Journal/Hail pointer precedent — a never-mapped agent, i.e.
every pre-041 snapshot, round-trips byte-identically; [[mental-maps]] owns
the type and its two knowledge-event Apply arms — [[sim-state-apply-world]]), plus, since spec 042,
a rolling situation (query) vector `SitVec`/`SitVecModel`/`SitVecTick`
(`omitempty`) the reducer sets verbatim from an `agent.situation_embedded`
companion — absent (nil `SitVec`) leaves selection on the legacy ranking
([[memory-retrieval]]) — plus, since spec 043 ([[decision-context]]), two
reducer-DERIVED self-knowledge surfaces maintained by existing arms with no
new event type: `IntentLog []IntentRecord` (US1, `omitempty`) — the
recent-intent ring, capacity `intentLogCap` (8), each record
`{Goal, Source, Reason, Tick, Outcome, OutcomeTick}` with `Outcome` empty
while executing then `done`/`failed`/`rejected`/`expired` (since spec 064,
also `stalled` — a needs-conditioned recovery's honest dead-source abort,
see [[sim-state-intent-lifecycle]]) — and
`NeedsAnchor *Needs`/`NeedsAnchorTick` (US2, `omitempty`; a POINTER on the
Journal/Hail precedent, deliberately deviating from the spec's value type so
a pre-043 snapshot round-trips byte-identically) — the trajectory window's
edge snapshot the decision prompt diffs current needs against to render
rising/falling/steady, `NeedsAnchorTick == 0` the unset first-window
sentinel — plus, since spec 062 ([[reflex-policy]]), `LastMindIntentDone
int64` (`omitempty`) — the tick the agent's most recent NON-REFLEX
(planner/plan/meeting) intent completed, the yield-window anchor the
reflex's PREP rungs consult to defer to a recent planner decision
(`prepYields`, elapsed = tick−LastMindIntentDone gated against
`prepYieldTicks`); 0 is the permanent sentinel for a never-mind-driven
agent — plus, since spec 083, `Neglect *NeglectState` (`omitempty` POINTER,
the Journal/Hail/Map precedent: a pre-083 snapshot round-trips
byte-identically) — the death-by-neglect detector's derived substrate: per
survival need, flat fields (no maps — fixed canonical JSON) for the
band-entry anchor (`*Since`, tick the need crossed below its spec-062 danger
band; 0 = not in band), the last-class-intent stamp (`*Intent`, tick a
`needClassGoals` goal landed; 0 = never), and the one-per-episode fired
latch (`*Fired`), written ONLY by three reducer arms (needs_changed /
intent_set / the new `sim.neglect_detected` — [[sim-state-apply-agents]],
[[sim-state-intent-lifecycle]]), lazily allocated on first non-zero write,
read by the executor's heartbeat sweep and the exported `NeglectDue`
predicate ([[executor-needs-survival]]); its six tick anchors are SHIFT
(only-non-zero) under the rebase taxonomy
([[guardian-miracle-rebase-taxonomy]]).

## Connections

Back to [[sim-state-reducer]] for the whole `State`/`Apply` picture and the
other five split-off notes. [[executor]] owns the needs/intent executor
semantics these fields carry; [[cognition]] and [[nightly-consolidation]]
cover the memory/consolidation surfaces; [[mental-maps]] owns the `Map`
field's type; [[memory-retrieval]] owns `Seq`/`Vec`/`VecModel`/`SitVec*`;
[[decision-context]] consumes `IntentLog`/`NeedsAnchor`; [[reflex-policy]]
consumes `LastMindIntentDone`. The Apply arms that mutate these fields live
in [[sim-state-apply-agents]], [[sim-state-intent-lifecycle]], and
[[sim-state-cognition-arms]].
