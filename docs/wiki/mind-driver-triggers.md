---
name: mind-driver-triggers
description: What arms a villager's planner thought and what its prompt contains: per-agent cadence with phase-preserving stagger, the trigger set (wake/completion/nightfall/encounter/map-correction/paused Guardian nudge), and the social-law/known-places/village-law prompt blocks. Split from [[agent-mind]] (The mind driver, triggers half).
kind: component
sources:
  - internal/mind/mind.go
  - internal/mind/prompt.go
  - internal/mind/parse.go
  - internal/mind/telemetry.go
verified_against: cb0eb0c0b00c7ecef9d0a6a88d49c3ee994953b4
---

# Mind driver cadence and prompt content

**The mind driver** (`internal/mind`): a replica fed by the loop's notify fan-out;
per-agent cadence (`replica.PlannerCadence()`, 1800 ticks by default — spec 048
promotes the cadence to a per-world [[world-tuning]] dial, the default living
in `internal/sim/tuning.go` as `defaultPlannerCadenceTicks` — staggered by
index; since TASK-44 the stagger is
phase-preserving — every re-arm steps in whole cadence multiples from the agent's
own due via `nextPhasePreservingDue`, never from the current tick, so a shared
stall cannot collapse agents into lockstep) plus triggers — wake, completion
idle, nightfall, first-adjacency encounters (`md.replica.EncounterCooldown()`,
2-game-hour pair cooldown by default — the same spec-048 dial family,
`defaultEncounterCooldownTicks`), a
mental-map correction that invalidates the agent's own current intent target
(spec 041 US3: `absorb`'s `agent.map_corrected` case arms the agent only when
one of the payload's gone facts matches the live intent's target or resolved
coordinates — a correction elsewhere in the map stays quiet, carried into the
next scheduled round as a memory instead; [[mental-maps]]) — since spec 081 the
same intent-match rule also arms on `agent.chopped`/`agent.quarried`: the actor
re-arms (a chop always did; a quarry now too), and so does any villager within
`sim.WitnessRadius` of the cleared tile whose live intent targeted it, standing
in for the `agent.map_corrected` those on-scene witnesses no longer receive
(the fact was removed silently at the act; [[mental-map-perception]]).
`agent.place_observed` (spec 097) is absorb-consumed but is NOT a planner
trigger: it drives the belief reconciler (`reconcilePlace` → a worker
injecting `agent.belief_reinforced`/`agent.memory_promoted` batches through
the door — [[executor-perception-observation]]), leaving cadence untouched.
Also armed — only while the replica is paused (spec 040, decision-6's paused
authoring chain) — a landed Guardian nudge (`guardian.nudged`), which arms each targeted
villager with the nudge event's seq as the causality edge; the game-time
debounce cannot reopen while frozen, so one nudge buys at most one round at
the frozen tick, and a nudge landed while running arms nothing — floored
by a 5-game-minute per-agent debounce (completion triggers otherwise form a
feedback loop that saturates the planner's provider). Planner prompts carry a social
context block (bonds, debts, reputation, loudest rumor, and the
last-conversation callback from the record ring — [[social-fabric]], TASK-22;
since TASK-42 scene replies get bounded parse-failure tolerance: `parse.go`'s
`lenientOutcome` repairs the observed unquoted-gist shape with zero extra
calls, and `telemetry.go`'s `cogSceneOutcome` variant carries the failed
reply's bounded `raw` text and a `retried` flag — the base `cogOutcomeEvent`
delegates there with the extras zeroed, keeping every other call site
byte-identical) and,
since TASK-13, a "Village law" block (`villageLaw` in prompt.go: active norms with
provenance, exile judgments — second-person for the exile — and the assembly call
while convening — since TASK-36 all rendered from the event-sourced meeting
convention's clock, with a bare "Village law:" header when none exists;
[[governance]]). Since spec 041, the prompt's world description is no longer
omniscient: `userPrompt`'s old blanket "Village: <first six structures>" line
and its bare-distance nearby-agent scan (`State.Structures`, all agents within
10 tiles regardless of what the agent has ever seen) are retired in favor of
`knownPlaces` (prompt.go), which renders only what the acting agent's OWN
mental map holds — landmark structures individually with provenance flavor
(witnessed/told/revealed), everything else place-shaped grouped by kind with
count + nearest, and an orientation line toward the nearest unexplored land;
the nearby-agent line itself now walks the map's peer sightings, so a peer who
slipped away unseen still renders where last seen rather than its live
position. Two villagers with different histories now see different worlds in
their own prompts ([[mental-maps]] owns the map subsystem this renders from).
Since spec 106 absorb also does two sleep-gating jobs alongside its arming
work: at batch end it refreshes the worker-facing per-agent unavailability
mirror (asleep|dead, one atomic word beside the `md.tick`/`md.tickRate`
mirrors) that [[tool-use-dispatch]]'s dequeue gate reads, and per event it
fires the agent's in-flight planner cancel slot on `agent.slept`/`agent.died`
(planner slot only — the consolidation that same `agent.slept` triggers is
untouched). The wake trigger is the gate's resumption path: `agent.woke` arms
the planner AND the same batch flips the mirror awake, so a villager whose
queued thought was skipped asleep re-thinks at the next `plan()` pass,
debounce permitting.
The driver also runs conversations (see
[[social-fabric]]). Villagers convened to the daily meeting are planner-suppressed
(`sim.AtMeeting`, checked in `plan()`) until close, their pending triggers left
armed — since musing no longer has a schedule of its own (spec 017, below), this
one gate now also covers it: a convened villager's tool-use loop simply never
runs, so `muse` cannot fire either. Since TASK-32 every trigger records its arming stimulus: `arm` takes the
event seq, kept in `pendingSeq` as the causality edge on the eventual telemetry.

## Connections

[[agent-mind]] is the parent note this child was split from; [[tool-use-dispatch]]
is its sibling child covering what happens next — the cognition-horizon gate
and the tool-use loop dispatch this trigger set feeds into; [[mental-maps]]
owns the map-correction arming condition and the `known_places` map state
this driver renders from; [[world-tuning]] owns the `PlannerCadence`/
`EncounterCooldown` per-world dials; [[social-fabric]] owns the bonds/debts/
reputation/rumor state the social context block renders; [[governance]] owns
the village-law/norms/exile state and the meeting convention this driver
suppresses planning during.

