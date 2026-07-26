---
name: event-types-scenario-incidents
description: "Scenario-incident event rows split from [[event-types]] (spec 077): sim.cold_snap, sim.forage_blighted, and the stranger entity family (arrived/moved/took/departed). Load when tracing the exercise catalog's authored pressure or the stranger's night arc."
kind: concept
sources:
  - internal/sim/scenario.go
  - internal/sim/stranger.go
verified_against: b6a20eaa4da1073a69959a5aff69591d931103a9
---

# Event types — scenario incidents (spec 077)

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.

Spec 077 (the exercise catalog wave, TASK-151) grows the incident kind
vocabulary from one (`gru_emerges`, spec 054) to four, adding six event
types with no format bump: every new `State` field is `omitempty`
([[sim-state-world-fields]] — `ColdSnapUntil`, `Stranger`, `StrangerTakes`),
so pre-077 snapshots round-trip byte-identically. The doctrine is
**ambient indistinguishability**: NO payload carries an authored/scenario
marker — indistinguishability is a property of the recorded artifact and
the emission preconditions (named predicates a future ambient emitter,
TASK-28, calls verbatim), not of a co-shipped dice path. Recorded events
are the only persistence; replay arms no scenario runtime and re-applies
these through reducer-total arms ([[scenario-machinery]]).

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `sim.cold_snap` | `ColdSnapPayload{night, until_tick}` in `internal/sim/scenario.go` — `until_tick` derives from the AUTHORED coordinates (tick + hours), so a late firing still expires on schedule | `scenarioIncidentEvents` (precondition: `coldSnapActive(s, tick)` false — no snap already holds) | latches `State.ColdSnapUntil`; expiry is a READ-TIME comparison (no end event — the belief-decay precedent); while active, the needs heartbeat's outdoor night warmth loss runs at `warmthLossColdSnap` (double the ambient `warmthLossCold`) through the same `decayNeeds` arithmetic ([[executor-needs-survival]]); SHIFTs across a time snap ([[guardian-miracle-rebase-taxonomy]]) |
| `sim.forage_blighted` | `ForageBlightedPayload{x, y, radius, tiles, regrow_tick}` — ONE merged event per firing (the `sim.food_rotted` precedent); `tiles` in deterministic row-major patch-walk order (`blightableTiles`); `regrow_tick` = firing tick + `blightRegrowTicks` (4 game days) | `scenarioIncidentEvents` (precondition: ≥1 unharvested forage tile in the patch — an exhausted patch skips silently, never retried) | appends the EXISTING `Harvest{X, Y, Regrow}` overlay per tile (idempotent re-apply — already-harvested tiles skip): a blighted tile IS a harvested tile with a long regrow, so perception, `agent.map_corrected`, and `sim.forage_regrown` all work unchanged |
| `stranger.arrived` | `StrangerArrivedPayload{night, x, y}` in `internal/sim/stranger.go` | `scenarioIncidentEvents` (preconditions: no stranger abroad; `strangerEntryValid` — passable + unprotected, the gru's spawn tile class) | `State.Stranger = &Stranger{X, Y, Night}` — the entity latch (the `gru.emerged` shape); the map renders it as `S` ([[tui-map-view]], [[tile-registry]]) |
| `stranger.moved` | `StrangerMovedPayload{x, y}` | `strangerStep` (executor tick, after `gruStep` — [[executor-tick-subsystems]]): greedy store-seeking toward the nearest unattended store, else seeded prowl (`rngAt` purpose `"stranger-prowl"`, per-decision, no stream); never onto a protected tile (`gruProtected` shared with [[gru]], not duplicated) | position update + `LastMove` cadence stamp (SHIFT) |
| `stranger.took` | `StrangerTookPayload{x, y, kind, n}` — bounded (`strangerTakeMax` 2 per take, `strangerTakeCooldownTicks` 600 between takes), kinds walked in the fixed goods order (no spears/axes) | `strangerStep`, standing ON an unattended store (no living villager within Manhattan 1); witnesses within `witnessRadius` gain situated memories (rumor fuel — the gru's witness idiom) | decrements pile/chest stock through the same state shapes agent withdrawal uses (clamped, reducer-total); appends `StrangerTake{Tick, X, Y, Kind, N}` to the ring-bounded `State.StrangerTakes` ledger (retain 32) — what zero-wanted rubric terms count; whole-line ALERT tier in the chronicle beside `social.chest_taken` (theft is theft — [[tui-chronicle-feed]]) |
| `stranger.departed` | `StrangerDepartedPayload{day}` | `strangerStep` at dawn (the `gru.withdrew` shape — gone by dawn, always) | `State.Stranger = nil` |

## Connections

[[scenario-machinery]] owns the kind vocabulary, compile arms, emission
preconditions, and the TASK-28 ambient/preemption seam;
[[executor-tick-subsystems]] hosts `strangerStep` in the tick order (after
`gruStep`, before governance); [[gru]] shares the protection predicates
(`gruProtected` — fire light + shelter are absolute for both entities) and
the entity-not-phenomenon precedent; [[executor-needs-survival]] reads the
cold-snap severity; [[sim-state-world-fields]] catalogs the new state;
[[guardian-miracle-rebase-taxonomy]] classifies the new tick anchors
(`ColdSnapUntil`/`LastMove`/`LastTake` SHIFT; `Night`/`StrangerTake.Tick`
KEEP); [[tui-chronicle-feed]] renders all six types (stranger.* in the
gru/threat family voice); [[event-types-curriculum-events]] catalogs the
passes this pressure is earned under.

## Operational notes

Ambient worlds NEVER see these types in v1 (spec 077 FR-014/FR-017): no
dice path exists — an armed schedule is the only producer until TASK-28
adds the ambient rolls and their `gruScheduledTonight`-style preemption
twins as one move. The stranger and the gru may legally be abroad the same
night — independent entities, independent latches, neither preempts the
other's schedule entry.
