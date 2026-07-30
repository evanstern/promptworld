---
name: executor
description: The deterministic agent-body layer's entry point — overview and orchestration of `stepEvents`; behavior detail lives in seven child notes (agent state, needs/survival, goals/intents, goal completions, world state, social/perception, tick subsystems). Load first for orientation; route to a child for a specific mechanic.
kind: component
sources:
  - internal/sim/executor.go
  - internal/sim/agents.go
verified_against: 376afd4cee54839a545bc88409f3c485c2f5149d
---

# Executor

The executor (TASK-5) replaced the placeholder wanderers: agents are now
deterministic bodies with needs, inventories, and multi-step intents, run unattended
by `stepEvents` between planner calls. The LLM planner (TASK-7) will *choose* goals;
the executor is what makes goals physically happen — and it must keep bodies alive
with no planner at all (the degraded-mode contract from the grounding session).
Spec 012 (resources/food/crafting v2) widened the body's economy substantially:
finer-grained resources, a crafting chain, fire fuel with burnout, spear-armed
hunts, and a shelter rest bonus. Spec 013 (inventory & storage v1) added a carried
bulk cap, ground piles, builder-owned chests, and food rot — this note covers
that v3 shape. Spec 032 (walls, axes, paths) layered in a fifth harvest tool
(the axe, tripling chop/quarry yield), a new impassable-structure family
(player-built walls, multi-cycle demolish/repair), and a walkable tile
improvement (paths, which double movement speed) — all additive `omitempty`
fields, so `format_version` stays 3 and no migration is needed to carry them.
Spec 038 (loud build failure & occupancy tolerance, TASK-91) changed how a
build goal's mid-work re-validation resolves: every `build_*` goal's
site-vanished path now emits a distinct `agent.build_failed` instead of
funneling through the same silent `agent.intent_done` a completion uses, and
a wall's reserved-tile occupancy check moved from a mid-work insta-cancel to
a bounded completion-time deferral — a passerby crossing the tile no longer
kills the build, only a squatter that outlasts the grace period does (below).
Spec 096 (TASK-95) generalized that pattern to every non-build goal's own
invalid-exit/contested resolution — `agent.intent_failed`, the same
loud-failure shape, now covers the goals `build_failed` never did
([[executor-goals-and-intents]], [[executor-goal-completions]]).

## How it works

The executor's behavior is organized into seven child notes, each covering a
focused behavior domain (every child ≥1,500 chars of substance):

- **Agent state** ([[executor-agent-state]]): the eight named bodies, their
  integer Needs/Inventory economy, and the pointer fields (`Map`, `SitVec*`,
  `IntentLog`, `NeedsAnchor`, `LastMindIntentDone`) that other subsystems own
  end to end but the executor threads through — plus where the v2/spec-032
  tuning constants live.
- **Needs, survival, and run end** ([[executor-needs-survival]]): the
  per-minute needs heartbeat and death causes, fire fuel/warmth, eating,
  needs-conditioned recovery holds (`warm_up`), and the `run.ended` terminal
  detector that freezes [[sim-loop]] once every villager has died.
- **Goals and intents** ([[executor-goals-and-intents]]): the `Intent` state
  machine (walk, instant-on-arrival, work-goal re-validation, duration
  lookup) and the full v2/spec-032 goal catalog it executes.
- **Goal completions** ([[executor-goal-completions]]): what each goal in
  that catalog emits and mutates on completion — gather yields, hand-craft
  outputs, build/demolish/repair on walls and other structures, cook/bathe/
  refuel.
- **World state** ([[executor-world-state]]): the v3 storage economy
  (carried bulk cap, ground piles, drop/pick_up, builder-owned chests, food
  rot) and the terrain/structure overlays (cleared/harvested/quarried tiles,
  walls, paths, fire/shelter/oven/chest/grave).
- **Social, perception, and memory provenance** ([[executor-social-perception]]):
  guarded plans, hails (the `talk_to` courtesy pause), the per-agent
  perception sweep, and the situated-memory/`origin` provenance every
  emitted memory carries.
- **Tick subsystems** ([[executor-tick-subsystems]]): the ancillary
  subsystems `stepEvents` drives each tick beyond agent bodies and goals —
  Guardian charge/order upkeep, scenario incidents/rubric, the gru's turn,
  the social beat, and the governance layer.

`stepEvents` stays a pure function of (pre-tick state, map, next tick) across
all seven; every effect is an event through [[sim-state-reducer]] — the
determinism and replay guarantees of the substrate hold unchanged over the
whole layer. TASK-7 replaces goal *selection*, never execution.

## Connections

[[reflex-policy]] decides what idle agents do — the goal-door surface
[[executor-goals-and-intents]] executes on landing. [[sim-loop]] drives the
tick and is what the run-end latch ([[executor-needs-survival]]) freezes.
[[event-types]] catalogs the event families every child note emits into.
[[sim-state-reducer]] is the mutation path shared by all seven children.
[[agent-mind]] authors the personas [[executor-agent-state]] carries.

## Operational notes

A fresh village (seed 42) builds fires within the first game-hour and
survives multiple days unattended. Known day-1 quirk: agents can't see
construction in progress, so several may each build a fire in the same
window — wasteful, harmless. Event volume: ~8 needs events/game-minute (one
per living agent) plus movement bursts; a two-day run is ~100k events. See
each child note's own "Operational notes" for the test suites exercising its
slice of the behavior.
