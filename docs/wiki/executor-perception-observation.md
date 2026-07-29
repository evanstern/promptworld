---
name: executor-perception-observation
description: Split from [[executor-social-perception]] — the per-agent perception sweep (spec 041 agent.saw/agent.map_corrected) and the spec-097 grounded arrival observations (agent.place_observed, exhaustive-within-radius, dedup window, mind-side belief reconciliation). Load when tracing what an agent sees, what it discovers gone, or how a visit falsifies a place-belief.
kind: component
sources:
  - internal/sim/executor.go
  - internal/sim/observe.go
  - internal/sim/memory.go
verified_against: b35a7ffec46ba996741cdba4af9652fcfd163b32
---

# Executor — the perception sweep and grounded arrival observations

Child of [[executor-social-perception]] (which keeps guarded plans, hails,
and memory provenance): the two channels through which ground truth reaches
an agent's private knowledge — the ambient perception beat and the spec-097
arrival observation.

## The perception sweep (spec 041, `perceptionEvents`)

Each awake living villager, on the same staggered per-agent cadence movement
uses (a fifth of a full per-tick sweep, T034's hot-path relief), diffs ground
truth within `witnessRadius` against its own `Agent.Map` and emits at most
one `agent.saw` (new/changed structures, piles, standing trees, unharvested
forage, unquarried rock, water shoreline, dens) and one `agent.map_corrected`
(remembered fresh facts whose place has vanished — a chopped tree, a
quarried-out outcrop, a drained pile, a removed structure; a merely-harvested
forage spot or cooling den is not gone, only unavailable, so it stays). A
correction's gone facts each ride a companion situated first-person discovery
memory (`mapCorrectedText`, `salMapCorrected`) in the same batch — memories
accrete only via `agent.memory_added`, never appended directly by a reducer
arm. Pure function of (state, map, tick): `stepEvents` reads, never mutates.
[[mental-maps]] owns the mental-map subsystem this sweep populates and
corrects; the executor's role is only the perception beat driving it. Since
spec 081, an `agent.map_corrected` names only agents who were dead, asleep, or
outside `witnessRadius` at removal: the `agent.chopped`/`agent.quarried`
reducer arms remove the felled/quarried fact from the actor's and every awake
in-radius witness's map at the act event ([[mental-map-perception]],
[[sim-state-reducer]]), so the sweep finds nothing to correct on-scene. The
same-tick beat is safe by `stepEvents` ordering — `perceptionEvents` runs
before `executeAtTarget`, reading pre-batch state where the tree still
stands. The chop/quarry emit sites also mint the actor a first-person act
memory ("Felled the tree at (x,y)." / "Quarried the outcrop at (x,y).",
`salChop`/`salQuarry`), a companion `agent.memory_added` in the act's batch
(the hunt precedent) — see [[executor-social-perception]] Memory emission.

## Grounded arrival observations (spec 097, `observe.go`)

The sweep only ever reported what IS there; `agent.place_observed` adds the
perception of ABSENCE — the architectural fix for the Thornspire finding
(confabulated place-beliefs were unfalsifiable because nothing ever recorded
what a place lacked). On the movement step that lands a walker ON its
intent's chosen target (D1: the intent-completing arrival — never per wander
step, never for a zero-distance intent), the executor emits the COMPLETE
sorted set of feature/entity kinds within `placeScanRadius` (the mental-map
fact vocabulary, overlay-aware — a cleared tree is not a tree). Absence is
implied by exhaustiveness (D2): anything not listed was not there; there is
no "absence_of" field because the reducer cannot know what an agent expected
— expectation lives mind-side. No reason interpretation, no model, anywhere
in the emission path (D5: pure function of world state at the arrival tick,
payload fully baked, additive event type — no format-version bump).

A companion low-salience situated memory (Origin `observed` — first-person,
`DirectPerception` true; salience = the `observation_base_salience` dial)
PRECEDES the event in its batch, so the mind's absorb loop reads it off the
replica when the observation lands. Repeat observations of an unchanged tile
inside the `observation_dedup_ticks` dial window collapse entirely —
both-or-neither, no event and no memory (D4: the working window, the event
stream, and the reconciliation rate are all bounded by one window). Dedup is
replay-pure: the `agent.place_observed` reducer arm records
`Agent.LastObs` (`ObservationMark{x, y, kinds, tick}`), and the emission
site compares the next arrival against that event-sourced anchor.

Mind-side, `internal/mind/reconcile.go` judges each observation against the
observer's place-beliefs through the spec-030 `agent.belief_reinforced` seam
([[nightly-consolidation]]): confirmation boost, bounded disconfirmation
decay (faster than silence — the `belief_confirm_boost` and
`belief_disconfirm_retain_percent` dials), silence untouched; a
disconfirming observation's memory gets one `agent.memory_promoted` surprise
bump. Matching (deterministic coordinate + feature-vocabulary today; D3
permits an LLM there) never runs sim-side.

## Connections

Parent: [[executor-social-perception]] (guarded plans, hails, situated-memory
constructors, origin provenance). [[mental-maps]] owns the map the sweep
feeds; [[event-types-mental-map]] carries the payload rows;
[[nightly-consolidation]] hosts the belief substrate the reconciliation
moves; [[world-tuning]] carries the four spec-097 dials.

## Operational notes

`observe_test.go` pins arrival emission, exhaustiveness, dedup, and
determinism; `internal/mind/reconcile_test.go` pins the three belief paths
and the myth-dies-slowly shape (SC-001).
