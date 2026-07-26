---
name: executor-needs-survival
description: The per-minute needs heartbeat, fire fuel/warmth, eating, needs-conditioned recovery holds (warm_up), and the run-end death detector — how an agent's Needs decay, recover, and can reach zero. Load for death causes, fuel/eat/recover mechanics, or the run.ended contract.
kind: component
sources:
  - internal/sim/executor.go
  - internal/sim/agents.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# Executor — needs, survival, and run end

Child of [[executor]] — the needs decay/recovery loop (heartbeat, fire fuel,
eating, needs-conditioned recovery holds) and the terminal run-end detector
that a fully-dead village triggers.

## How it works

**Heartbeat**: every game-minute (`tick%60 == 0`) each living agent's needs decay via
`decayNeeds`: food always falls; rest falls awake (or recovers asleep — at
`restRegenShelter` (6/minute) on a shelter tile, `restRegenSleep` (4) otherwise, the
plank economy's payoff for building one); warmth falls at night outdoors, recovers
near a **lit** fire or in shelter, drifts up by day. A fire is lit iff
`tick < Structure.FuelUntil` — `warmAt` takes the tick and checks it, so a burned-out
fire grants no warmth. Zero food or zero warmth drains health; health at 0 emits
`agent.died` with cause `starvation` / `exposure` / `collapse`. (A fourth cause,
`gru`, exists since spec 044 US3 but is never the heartbeat's: it is emitted by
`gruStep` itself when an escalated attack kills an already-weakened villager,
with its own inline witness-death memory loop, since gru attacks land off the
%60 heartbeat where the executor's witness-death block runs — [[gru]].) The new values land
as one absolute `agent.needs_changed` event per agent per minute (absolute values =
replay-safe).

**Fire fuel** (T019/T020): `build_fire` (still 2 wood) lights a fire for
`2×s.FireBurnPerWood()` (4 game-hours per wood by default, so 8 hours).
`refuel_fire` (instant on arrival, 1 wood) pushes `FuelUntil` forward by
`s.FireBurnPerWood()`, capped at `now + fireFuelCap` (12 hours); relighting a
cold fire starts the window from now. `fireBurnPerWood` is a spec-048
[[world-tuning]] per-world dial now — the default (still 4 game-hours) lives
in `internal/sim/tuning.go` as `defaultFireBurnPerWood`, and every call site
reads the tuned value through `State.FireBurnPerWood()`; `fireFuelCap` itself
is NOT promoted (research R6) and still truncates the effective deadline.
Every tick, `stepEvents` sweeps `Structures` for a fire
whose `FuelUntil` falls in the tick's window (`tick-1 < FuelUntil <= tick`) and emits
`sim.fire_burned_out` exactly once on that transition (no state effect — lit-ness
stays derived), plus a low-salience witness memory ("Watched the fire burn out.")
for every living agent within `witnessRadius`.

**Needs-conditioned recovery** (spec 064, from the TASK-101 spike Direction B):
`Intent` gains an OPTIONAL completion condition — `UntilNeed` (a member of the
closed set `warmth`/`rest`/`food`) and `UntilValue`, a need level — plus
`HoldRef`, the need level captured at the hold's anchor tick; all three
`omitempty`, so a conditionless intent (every pre-064 intent) marshals
byte-identically. When `UntilNeed` is set, `executeAtTarget` intercepts BEFORE
the per-goal switch and hands off to `recoveryHoldEvents` instead of the goal's
default arrive-and-done: the intent HOLDS at its target (visibly recovering,
not idle) and is checked every tick against the live need — already-satisfied
completes at once, a threshold crossing completes normally (arming the spec-062
yield window iff the ring source `isMindSource`, so a reflex-issued recovery
never arms it), a higher-priority survival need (the reflex ladder's order,
food > warmth > rest) crossing into ITS danger band ends the hold so the
agent re-decides (no new preemption immunity — a hold is LESS sticky than an
ordinary intent, never more), and no net gain over a full `recoveryStallTicks`
(300, ~5 needs heartbeats) window aborts with the distinct `agent.recovery_stalled`
outcome (dead fire, displaced source, unreachable threshold) rather than
loitering forever. `warm_up` is the evidenced consumer — a planner tool
resolving exactly like `goto_warmth` but carrying the condition, with an
optional `until_warmth` clamped (spec 058 clamp-with-notice posture) into
`[warmthRecoverFloor, needMax]` (`warmthRecoverFloor` = `dangerWarmthBelow`,
350; `needMax` 1000) via the single `clampWarmUp`/`ClampWarmUp` clamp home,
defaulting to the doctrine constant `warmthRecoverTo` (800, a healthy margin
above the danger band) when absent — and the [[reflex-policy]] day AND night
warmth rungs (`reachKnownWarmth`) now issue the same conditioned `goto_warmth`
at the doctrine default, so a reflex-driven recovery also holds at the fire
instead of arriving, idling, and wandering off cold (the world-01
arrive-idle-vacuum, this spec's Direction B). `wakeReason` (US4, the audit's
Gap C) gains a matching cold-emergency wake arm — a sleeper whose warmth falls
below `exposureWakeBelow` (150, the hunger-emergency wake's shape and
magnitude exactly, a deliberate deviation from the plan's nominated 350: an
emergency floor, not the routine-dip danger band, so a sleeper isn't roused
merely for being cold) wakes only when night AND the reflex's own warmth
ladder finds something actionable (the hunger wake's "food in hand" analog,
the churn bound); a cozy fire-side sleeper sleeps through untouched.
`wakeReason` now takes the state/map/tick it needs to run that ladder check,
rather than the bare `(agent, night)` it took before. Held-pinned villagers are excluded from the
emergent-gathering quorum ([[governance]]) — a survival hold is not an
elective assembly.

**Eating** (T018, `eatOutcome`): the reflex's `agent.ate` direct-event path (and the
planner's guarded-plan equivalent) now computes an outcome rather than emitting a
bare marker. `eatOutcome` consumes the most-nutritious form first — `Meals` →
`FoodCooked` → `FoodRaw` — one unit at a time until `Needs.Food` reaches `satietyAt`
(900) or the inventory runs dry, and returns `false` (nothing eaten, no event) if
already sated or carrying no food at all. Each form restores a different amount
(`mealRestore` 100, `foodCookedRestore` 80, `foodRawRestore` 40 — cooking roughly
doubles raw, a meal is the best food); the payload carries counts consumed per form
plus the absolute post-eat food need, so the reducer never re-derives arithmetic.
`wakeReason`'s hunger-emergency wake check now looks for *any* carried food form,
not just raw. `canGive` (the give-to-starving social rule) checks `Inv.FoodRaw`
specifically — raw is deliberately the form a subsistence village shares.

**Run end** (spec 044 US1): `stepEvents` opens with a terminal guard — an
ended world (`State.Ended`) returns nil, ever after; since `stepEvents` is the
sim's only emitter, this single latch freezes simulated time while [[sim-loop]]
keeps serving reads. At the end of the batch sits the matching detector: when
every villager still living at the tick's start (`livingCount`, `state.go`)
died within this batch, the run is declared over with a `run.ended` event
(`RunEndedPayload{tick, deaths, final_cause}`) emitted as the batch's LAST
event — after every same-tick `agent.died` (heartbeat or [[gru]]) and its
witness memories, so no sim event ever trails the declaration. The payload
carries the whole run's death history — the `State.Deaths` ledger plus this
batch's deaths, `final_cause` being the last death's cause — so no consumer
(status, the [[morgue]]) ever scans the log for them. A `!s.Ended` check belts
the top-of-function guard: exactly once per world, ever.

## Connections

Parent note: [[executor]]. [[world-tuning]] owns the fire-fuel dials this
section reads; [[reflex-policy]] issues the conditioned `goto_warmth`/`warm_up`
this section executes and owns the wake ladder `wakeReason` consults;
[[governance]] excludes held-pinned villagers from its gathering quorum;
[[morgue]] consumes the run-end death history; [[sim-loop]] is what
`run.ended` freezes into the ended posture.

## Operational notes

The v2 economy adds a full crafting chain (wood/stone → planks/refined_stone
→ spears/shelter/oven) and a fire that must be refueled or it goes cold —
`whole_feature_test.go` and `food_fire_test.go` exercise the chain and the
fuel sweep end-to-end.
