---
name: executor-goals-and-intents
description: The Intent state machine (walk, instant-on-arrival, work-goal re-validation, duration lookup) and the full v2/spec-032 goal set it executes — gather/craft/build/station goals and their per-goal durations. Load to see what a goal IS before checking what it DOES on completion (executor-goal-completions).
kind: component
sources:
  - internal/sim/executor.go
  - internal/sim/agents.go
  - internal/sim/recipes.go
verified_against: 864d2a3bcff4b3113739d596befc72229a84d4b8
---

# Executor — goals and intents

Child of [[executor]] — how an `Intent` executes as a state machine, and the
full catalog of goals (gather, hand-craft, build, station, adjacent-build,
multi-cycle work) it can carry.

## How it works

**Intents**: `Intent{Goal, Target, Res, WorkStart}` executes as a state machine —
walk (one tile per 5 ticks, staggered per agent, next hop from [[reflex-policy]]'s
BFS), then on arrival: instant goals (`sleep`, `wander`, `goto_warmth`,
`refuel_fire`, since spec 084 `heed_directive` — the DIRECTIVE rung's
walk-to-site leg ([[guardian-designations]]), arrival IS the outcome and the
next idle decision picks the work leg; never model-facing, the `seek`-alias
class — and since spec 041 `search` — [[mental-maps]]'s deliberate-
exploration goal, wander-class because the walk itself does the exploring:
movement marks explored terrain and the perception beat witnesses what's
there, so arrival needs no extra work) complete immediately; work goals
re-validate the resource or station
(someone may have taken it, or a fire may have gone cold — the contested-resource
pattern, spec 012 FR-002/FR-014), emit `agent.work_started`, and after the goal's
duration (`workDuration`, below) emit the completion event, which the reducer turns
into inventory, overlays, structures, or needs. Since spec 038, a build goal
(`isBuildGoal`: `build_fire`/`build_shelter`/`build_oven`/`build_chest`/
`build_path`/`build_wall_plank`/`build_wall_stone`) whose mid-work re-validation
finds the site gone no longer falls through to the bare `agent.intent_done` — 
`buildFailedEvents` emits a distinct
`agent.build_failed{agent, goal, reason: buildFailSiteUnbuildable}` paired
same-tick with a situated first-person failure memory (`OriginAction`, shelter
salience, `buildStructureName`/`buildFailureCause` composing "My <structure> was
never built: <cause>."), so a cancelled build is never mistaken for a finished
one (the phantom-wall belief loop TASK-91 fixes). Spec 096 (TASK-95) then
generalized this to every OTHER work goal's own contested/invalid re-check:
`forage`/`chop`/`hunt`/`demolish`/`repair`/`quarry`/`cook`/`bathe`'s mid-work
site/resource re-validation, and `craft_*`/`cook`/`bathe`/`deposit`/`withdraw`'s
completion-time no-op recheck, now emit `agent.intent_failed` via
`intentFailedEvents` — the same shape (position + a paired failure memory at
the build-failure salience tier), reason drawn from a small closed vocabulary
(`intentFailTargetGone`/`intentFailContested`/`intentFailInvalid`) —
INSTEAD of the bare `agent.intent_done` these resolved through before;
`agent.intent_done` now means an unambiguous success (or an instant/
wander-class goal with no re-validation of its own — [[event-types-agent-intents]]
carries the full reason/payload table). Movement itself gets a second,
conditional cadence slot (spec 032 US3): the staggered phase-0 tick always steps,
but a phase-2 tick also steps when the agent is standing ON a path tile
(`pathAt`) — stepping FROM a paved tile doubles effective speed along it, while
an unpaved agent never sees the extra slot, so nothing else about movement changes.

**The v2 goal set** adds `quarry`/`collect_water` (gather, like forage/chop/hunt),
`craft_planks`/`craft_stone`/`craft_spear` (hand-crafts, `SiteAnywhere` — no travel,
work happens on the agent's own tile), `build_oven` (alongside `build_fire`/
`build_shelter`), and `cook`/`bathe`/`refuel_fire` (station actions at a fire or
oven). Spec 032 (walls, axes, paths) adds a fourth hand-craft, `craft_axe`
(alongside the other three), two ADJACENT-build goals — `build_wall_plank`/
`build_wall_stone` (the builder stands beside the tile the wall lands on, unlike
every stand-on-target build before it, so a wall can never entomb its own
builder), a stand-on-target build, `build_path` (the fire/oven/chest pattern), and
two multi-cycle work goals on an existing wall, `demolish` and `repair` — each
completion may re-arm the SAME intent for another work cycle rather than
finishing (below). Since spec 014 (TASK-53) `intentDuration` reads `intentDurations`, a table
built at init from the [[tool-registry]]'s per-tool `Cost.DurationTicks` (values
hand-equal to the sim constants, pinned by
`TestWorldToolDurationsMatchSimConstants`) — since spec 017, filtered to
GOAL-DOOR tools (`Effect World && PlanStep`, the same discriminator
[[tool-registry]]'s coverage check uses): `set_plan` is a World tool but never
reaches `intentDuration` by its own name (each of its plan steps names an
already-covered goal-door goal instead), so it is deliberately absent from this
table rather than carrying a meaningless zero-duration entry. Goals with no
registry duration — the instant verbs and the internal `seek` alias — complete
on arrival (0), exactly as the old switch's default did. `workDuration` overrides the plain
`intentDuration(goal)` lookup for two
context-dependent cases: a spear-carrying hunt takes `huntTicksSpear` (faster than
the bare-handed default) and cooking at an oven takes `cookOvenTicks` (slower than
at a fire) — both read off current state (`Agent.Inv.Spears`, the target structure),
never persisted on the `Intent`.

## Connections

Parent note: [[executor]]. [[executor-goal-completions]] is this note's
companion — what each of these goals does when its duration elapses.
[[reflex-policy]] and [[mental-maps]] choose/resolve the `search` goal;
[[tool-registry]] supplies `intentDurations` and the GOAL-DOOR discriminator
that filters it; [[cognition]] and [[sim-loop]] land intents through
`InjectIntent`'s landing ladder before they ever reach this state machine.
