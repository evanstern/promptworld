---
name: event-types-mental-map
description: Perception and mental-map event rows split from [[event-types]]: agent.moved, agent.saw, agent.map_corrected, agent.place_observed, social.place_told, guardian.place_revealed. Load when tracing spec 041's per-agent spatial knowledge — perception sweeps, told directions, corrected/stale facts, grounded arrival observations, and vision-granted places.
kind: concept
sources:
  - internal/sim/executor.go
  - internal/sim/agents.go
  - internal/sim/observe.go
verified_against: 012f715f55d8d87317e601ad75686c599d277349
---

# Event types — perception & mental map events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.


Spec 086 (agent-named payloads): every agent-referencing field in this
family's payloads is a `sim.AgentRef` — the wire carries
`{"id":N,"name":"…"}` objects (lists element-wise), the name stamped at
emission from the fixed roster via `Ref`/`Refs`; sentinels marshal
`{"id":-1,"name":""}`. Legacy bare-int rows decode through the dual-shape
unmarshal forever and reducer arms fold `.ID`s only — the conventions and
the normative back-compat matrix live on [[event-types]] ("Agent
references are named refs").
Spec 041 ([[mental-maps]] — per-agent spatial knowledge) is likewise
format-stable: `Agent` gains `Map *MentalMap` (`omitempty`, the Journal/Hail
pointer precedent — a pre-041 snapshot with the field absent round-trips
byte-identically), and FOUR new whitelisted/reducer event types drive it —
`agent.saw` (perception sweep witnesses), `agent.map_corrected` (a remembered
place found gone), `social.place_told` (directions exchanged in talk), and
`guardian.place_revealed` (a vision's optional place grant) — full shapes and
reducer effects in the table below.

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `agent.moved` | `AgentMovedPayload{agent, x, y}` | executor pathing — **LEGACY worlds only since spec 104**: under the coalescing regime walks emit `agent.path_started` instead; this arm is retained forever so old logs replay unchanged | position updated; spec 041: the mover's surroundings (perception radius) are marked explored in its private mental map, and mover + awake bystanders within the witness radius record each other's positions as peer sightings (`MentalMap.Peers` — what `talk_to`/`seek` resolve against) — silent derived bookkeeping, no companion event (`agent.woke` and `guardian.entity_moved` villager moves run the same sighting pass) |
| `agent.path_started` (spec 104) | `PathStartedPayload{agent, path, move_every, phase}` — the FULL departure-time BFS route (tiles stepped onto, ending on the intent target) plus the cadence numbers baked at emission (spec 092: advancement never reads compiled constants) | executor pathing under the coalescing regime (`AmbientCoalescing` — the genesis-pinned `needs_checkpoint_minutes` marker): one event per walk instead of one per tile | installs `Agent.Path` (an in-flight `PathSegment`, `omitempty`); the derived-advancement engine (`internal/sim/advance.go`) then executes each step at its scheduled tick with EXACTLY the `agent.moved` arm's bookkeeping — position, explored bits, mutual tick-stamped peer sightings — at per-step fidelity (ruling 2); the arrival step retires the segment; every walk-invalidating event (`agent.intent_set`/`intent_done`/`intent_failed`/`build_failed`/`recovery_stalled`/`slept`/`died`, `gru.attacked`, `social.hailed`, `guardian.entity_moved`, `clock.paused`, `guardian.time_snapped`) truncates it in its own arm |
| `agent.path_truncated` (spec 104) | `PathTruncatedPayload{agent, x, y}` — the agent's ACTUAL position at truncation (outcome in payload, never recomputed) | executor, for the one walk deviation no other event records: the declared path's next tile became impassable (a wall built mid-segment); the next tick re-plans or resolves unreachable | sets position from the payload and clears `Agent.Path` |
| `agent.saw` (spec 041) | `SawPayload{agent, facts}` — `facts` are fully-baked `PlaceFact`s (kind, x, y, seen, prov, src?, detail?), sorted (kind, x, y) at emission | executor perception sweep, on the agent's movement-cadence beat: ground truth within the witness radius (structures, ground piles, standing trees / unharvested forage / unquarried rock, water shoreline, dens) diffed against the agent's mental map — new, changed (a fire's `FuelUntil` detail), horizon-staled, or provenance-upgraded facts only, so a settled map emits nothing | each fact upserted verbatim into `Agents[agent].Map.Facts` (provenance `witnessed`, `Seen` = event tick); digest-only — deliberately NO chronicle line (too chatty) and not an absorb trigger |
| `social.place_told` (spec 041 US5) | `PlaceToldPayload{from, to, facts}` — up to 2 facts per direction, fully baked at emission: told provenance, the TELLER's `seen` (secondhand is never fresher — staleness is the trust model), `src` = the immediate teller, `detail` carried verbatim; canonical (kind, x, y) order | executor talk sidecar (`talkEvents`, beside the `TellableFor` rumor slot, every founded talk incl. hail-founded): per direction, the teller's fresh facts the listener lacks-or-holds-staler, selected freshest → nearest-to-listener → coordinate order | upsert into the RECEIVER's map only where absent-or-staler (`Seen` compared — fresher knowledge never loses to secondhand); companion situated memories both sides ride the same batch ("Told Birch about the fire at (x,y)." / "Birch told you of a fire at (x,y)."); chronicle grammar line; not an absorb trigger |
| `agent.map_corrected` (spec 041 US3) | `MapCorrectedPayload{agent, gone}` — `gone` carries the facts AS REMEMBERED (verbatim from the agent's map, canonical order; context baked at emission for narration) | the same perception sweep, right after the agent's `agent.saw`: remembered FRESH facts within the witness radius that are ABSENT from ground truth (`groundFactPresent` — absence of the PLACE, not of its availability: a harvested forage spot or cooling den persists; a chopped tree, quarried-out outcrop, drained pile, or removed structure is gone) | each gone fact removed from the agent's map; a situated first-person discovery memory rides the SAME batch per fact as a companion `agent.memory_added` (`mapCorrectedText`, `salMapCorrected`, Origin witness — memories accrete only via `agent.memory_added`); chronicle grammar line; absorb trigger when a removed fact matches the agent's current intent target (the planner re-arms — [[agent-mind]]). **Spec 081**: this correction now fires only for agents dead/asleep/out-of-radius at removal time — the `agent.chopped`/`agent.quarried` arms remove the tree/rock fact from the actor and awake in-radius witnesses at the act event (no new event type; [[mental-map-perception]], [[sim-state-reducer]]), and those two harvest events became absorb triggers too, under the same intent-match re-arm rule |
| `guardian.place_revealed` (spec 041 FR-014) | `PlaceRevealedPayload{agent, facts}` in `internal/sim/mentalmap.go` — the emitter bakes only the place identity (kind, x, y, prov `revealed`); `seen` and `detail` are the reducer arm's NORMATIVE stamps | `send_vision`'s optional place grant (`place_kind`/`place_x`/`place_y` argument triple, all-or-none), riding the vision's atomic `InjectSocial` batch after the nudge memories | validates (living target, every fact names a REAL place — `groundFactPresent`, so the god reveals what is, never what isn't) then upserts into the target's mental map with `Seen` = landing tick, provenance `revealed`, `Detail` = ground truth at landing (a fire's `FuelUntil`); the companion Origin-omen memory ("The vision showed you the fire at (x,y).") rides the same batch as `agent.memory_added`; chronicle grammar line; absorb rides the batch's `guardian.nudged` (paused-authoring trigger) |
| `agent.place_observed` (spec 097) | `PlaceObservedPayload{agent, x, y, radius, kinds}` in `internal/sim/observe.go` — `kinds` is the COMPLETE sorted set of feature/entity kinds within `radius` (= `placeScanRadius`) at the arrival tick, the mental-map fact vocabulary; absence is implied by exhaustiveness (no "absence_of" field — expectation lives mind-side) | executor, on the movement step that lands a walker ON its intent's chosen target (intent-completing arrival — never per wander step, never for a zero-distance intent); suppressed entirely (event AND companion memory) when the same tile + identical kinds were observed inside the `observation_dedup_ticks` dial window | records the agent's `LastObs` dedup anchor (`ObservationMark{x, y, kinds, tick}`) — nothing else; the companion low-salience `agent.memory_added` (Origin `observed`, `observation_base_salience` dial) PRECEDES it in the batch so the mind's absorb reads it off the replica; chronicle grammar line; mind-side absorb trigger for belief reconciliation ([[nightly-consolidation]] hosts the belief substrate; `internal/mind/reconcile.go` judges confirm/disconfirm through the spec-030 `agent.belief_reinforced` seam) |

## Connections

[[mental-maps]] owns the `agent.saw`/`agent.map_corrected`/
`social.place_told`/`guardian.place_revealed` family end to end — the
executor's perception sweep and talk sidecar emit the first three, the
`send_vision` door the fourth, and `internal/sim/mentalmap.go` reduces all
four.
