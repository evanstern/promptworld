---
name: reflex-policy
description: decideIntent's deterministic reflex ladder for idle agents and resolveGoal's shared target resolver, overview + spec history. Detail lives in five children — [[reflex-survival-rungs]] (SURVIVAL rungs), [[reflex-prep-arbitration]] (PREP yield gate/rungs, wander), [[reflex-goal-resolution]] + [[reflex-goal-resolution-structures]] (resolveGoal's full goal catalog), [[reflex-pathfinding]] (BFS/nearest geometry). Load this note first for orientation, a child for mechanics.
kind: component
sources:
  - internal/sim/policy.go
  - internal/sim/path.go
verified_against: 048259bb42b03cc6ebeb13a49f367c2e3a7d4d37
---

# Reflex policy & pathfinding

`decideIntent` is the deterministic pure function that gives idle, awake agents
something to do — since TASK-7, only agents idle past the 120-tick grace (the
planner's injection window). It is the permanent degraded mode: when no planner
thoughts arrive, this keeps bodies alive. `resolveGoal` (same file) is the shared
target resolver: planner goals from [[agent-mind]] resolve to coordinates through
the exact same nearest-X helpers the reflex uses. Spec 012 widened `resolveGoal`'s
goal set considerably (quarrying, water, crafting, an oven, refueling, cooking,
bathing) while trimming the reflex ladder itself down to one addition — refueling
a dying fire — and one removal — shelter-building dropped out of the reflex
entirely once it was re-costed in planks. Spec 013 (inventory & storage v1) widened
`resolveGoal` again — a chest to build, goods to drop/pick up/deposit/withdraw —
and left the reflex ladder itself completely untouched: all five new goals are
planner/plan-only (FR-014), added to `resolveGoal` but never reachable from
`decideIntent`. Spec 032 (walls, axes, paths) widened it once more — two wall
builds, a demolish/repair pair on an existing wall, a fourth hand-craft
(`craft_axe`), and a path build — every one planner/plan-only, same pattern:
the reflex ladder itself gains nothing from spec 032 (an axe or a wall is
never something `decideIntent` reaches for on its own). Spec 014 (TASK-53) restructured `resolveGoal` from one large
switch into `goalResolvers`, a name-keyed resolver table with the old per-verb
bodies verbatim — the [[tool-registry]]'s boot-time coverage gate
(`sim.ValidateToolCoverage`) asserts every World tool on the villager roster has
a table entry, so a registered verb can never lack its resolver. The plan-step
accept set that once lived beside it (`planGoals`) is gone: the sim door now
derives it from the registry ([[sim-loop]]).

Spec 041 ([[mental-maps]]) changed WHAT every resolver above searches, not
the search mechanics: every goal that targets a place now resolves against
the acting agent's own FRESH facts (`nearestKnown`/`nearestKnownAdjacentTo`,
`path.go` — knowledge-gated twins of `nearest`/`nearestAdjacentTo` that keep
the identical BFS geometry and tie-breaking, only the match closure differs),
not the world's ground truth. A resolver that knows of nothing of the right
kind fails with an epistemic reason ("you know of no forage") rather than
resolving to a place it has never perceived. Availability that is not itself
place knowledge — a harvested forage spot, a cooling den, wall damage, chest
contents, quarry depletion — stays layered on top as an ordinary ground
condition: the agent knows the PLACE and walks there; whether the walk pays
off is checked at arrival, exactly as before. A new goal, `search`
(US4), resolves to the nearest exploration frontier
(`nearestFrontier`) instead of any resource — the deliberate answer to
knowing of nothing. `talk_to`/`seek` resolves to the target's last KNOWN
sighting (the mental map's peer record), never its live coordinates.

## How it works

Spec 062 (TASK-103, from the TASK-101 spike) restructured `decideIntent` into
two explicitly classified rung groups plus an idle filler, each its own
function — "instinct that yields to intelligence": the reflex ladder is a
safety net, not a scheduler, and the classification lives in **code
structure** (FR-001), not prose: `survivalDecision` (life-saving instinct,
runs first, unconditioned), `directiveDecision` (spec 084 — the guardian's
HARD directives, between survival and prep: unconditioned by the yield
window and danger bands, stateless — it re-derives the oldest active
directive addressing the agent from `State.Directives` at every idle
decision, so interruption-resume needs zero code and a directive-free world
is byte-identical; routing per kind in [[guardian-designations]], incl. the
new instant `heed_directive` walk goal), `prepDecision` (opportunistic
village upkeep, yields to a recent planner intent or a danger band), and
`wanderDecision` (the idle filler). Root cause this fixed (spike TASK-101, 057 audit): the
reflex's PREP rungs used to fire the instant a planner intent completed,
never checked warmth, and counter-scheduled the agent away from the fire the
planner just sent it to — world-01's forage↔goto_warmth thrash (Sage: 436
flips, 334 within ≤200 ticks). `thrash_regression_test.go` (US4) encodes both
the old flip and the new non-loop in one deterministic test.

This note now summarizes five split-off children that carry the doctrine and
catalog detail:

[[reflex-survival-rungs]] carries the six SURVIVAL rungs — eat, get food (with
its US4 frontier-search fallback), the night warmth ladder (+ its own US3
frontier-search fallback and terminal sleep), the daytime nap, and the day
warmth rung (spec 062 US2) — first-match-wins and unconditioned by the yield
gate below.

[[reflex-prep-arbitration]] carries the PREP yield gate (spec 062 US1 —
`prepYields`'s yield-window and danger-band clauses, "instinct yields to
intelligence"), the three PREP rungs it guards (fire-knowledge build/chop,
unconditional refuel, larder stock-up), the `wanderDecision` filler, and
waking (`wakeReason`).

[[reflex-goal-resolution]] carries `resolveGoal`'s consumable, survival, and
exploration goal catalog (`eat`, `quarry`/`collect_water`, the fire/shelter/
oven builds, the hand-crafts, `refuel_fire`, `cook`, `bathe`, `search`,
`sleep`/`wander`/`goto_warmth`, `warm_up`, `talk_to`/`seek`) — the goal set
spec 012's economy and spec 041's knowledge-gating grew.

[[reflex-goal-resolution-structures]] carries `resolveGoal`'s storage and
structural goal catalog (`build_chest`, `drop`/`pick_up`/`deposit`/`withdraw`,
`craft_axe`, `build_wall_plank`/`build_wall_stone`, `demolish`, `repair`,
`build_path`) — spec 013's storage economy and spec 032's walls/axes/paths.

[[reflex-pathfinding]] carries `path.go`'s BFS geometry: fixed neighbor order
(N, E, S, W), FIFO frontier, `nearest`/`nearestAdjacentTo`, and spec 041's
knowledge-gated `nearestKnown`/`nearestKnownAdjacentTo`/`nearestFrontier`
wrappers over the same search — unchanged in its own geometry by spec 012,
032, or 041.

## Connections

[[executor]] invokes decisions on a staggered cadence and executes the resulting
intents, including the fire-fuel and cooking/crafting mechanics several of the new
goals above key on; passability comes from [[executor]]'s terrain overlays over
[[worldmap-generation]]; randomness only via [[deterministic-rng]] purpose tags
(`wander`, plus `genesis` placement in [[sim-state-reducer]]); [[mental-maps]]
is the per-agent knowledge store every goal resolver now reads through —
`nearestKnown`/`nearestKnownAdjacentTo`/`nearestFrontier`, `knowsAnyFresh`,
`warmKnownPredicate`, and `peerSightingOf` all live in this note's files but
are gated entirely on facts [[mental-maps]] and the executor's perception
sweep populate. [[sim-state-reducer]] owns `Agent.LastMindIntentDone` and the
`agent.intent_done` arm that arms it (spec 062); [[event-types]] catalogs that
arm's effect; [[guardian-miracles]] classifies `LastMindIntentDone` SHIFT
(only-non-zero) in the `rebaseTicks` taxonomy; [[testing-strategy]] tracks
the spec-062 test files (`yield_state_test.go`, `day_warmth_test.go`,
`night_search_test.go`, `thrash_regression_test.go`) alongside the updated
`reflex_matrix_test.go`.

## Operational notes

BFS over a 64×64 map per decision/move is the current throughput ceiling — the
executor still clears >200k ticks/sec in the test harness, and auto-slow
([[sim-loop]]) degrades honestly under load. TASK-7 replaces this ladder with
planner-chosen goals; the ladder itself must remain reachable as the fallback.
