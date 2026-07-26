---
name: executor-needs-survival
description: The per-minute needs heartbeat, fire fuel/warmth, eating, needs-conditioned recovery holds (warm_up), and the run-end death detector — how an agent's Needs decay, recover, and can reach zero. Load for death causes, fuel/eat/recover mechanics, or the run.ended contract.
kind: component
sources:
  - internal/sim/executor.go
  - internal/sim/agents.go
verified_against: 8495b34ffb9ee5dc02e224025f0a23313bbab900
---

# Executor — needs, survival, and run end

Child of [[executor]] — the needs decay/recovery loop (heartbeat, fire fuel,
eating, needs-conditioned recovery holds) and the terminal run-end detector
a fully-dead village triggers.

## How it works

**Heartbeat**: every game-minute (`tick%60 == 0`) each living agent's needs decay via
`decayNeeds`: food always falls; rest falls awake (or recovers asleep — at
`restRegenShelter` (6/minute) on a shelter tile, `restRegenSleep` (4) otherwise, the
plank economy's payoff for building one); warmth falls at night outdoors —
at `warmthLossCold` (4/minute) ambiently, or `warmthLossColdSnap` (8, spec
077) while a scheduled cold snap holds (`coldSnapActive`: `tick <
State.ColdSnapUntil`, a read-time check, no end event — same
arithmetic path, harsher night; fire warmth beats the snap
as it beats ambient cold) — recovers
near a **lit** fire or in shelter, drifts up by day. A fire is lit iff
`tick < Structure.FuelUntil` — `warmAt` checks the tick, so a burned-out
fire grants no warmth. Zero food or zero warmth drains health; health at 0 emits
`agent.died` with cause `starvation` / `exposure` / `collapse`. (A fourth cause,
`gru` (spec 044 US3), is never the heartbeat's: `gruStep` itself emits it
when an escalated attack kills an already-weakened villager, with its own
inline witness-death memory loop — gru attacks land off the %60 heartbeat
where the executor's witness-death block runs; [[gru]].) New values land
as one absolute `agent.needs_changed` event per agent per minute (absolute =
replay-safe).

**Fire fuel** (T019/T020): `build_fire` (still 2 wood) lights a fire for
`2×s.FireBurnPerWood()` (4 game-hours per wood by default, so 8 hours).
`refuel_fire` (instant on arrival, 1 wood) pushes `FuelUntil` forward by
`s.FireBurnPerWood()`, capped at `now + fireFuelCap` (12 hours); relighting a
cold fire starts the window from now. `fireBurnPerWood` is now a spec-048
[[world-tuning]] per-world dial — the default (still 4 game-hours) is
`defaultFireBurnPerWood` in `internal/sim/tuning.go`, every call site reading
it via `State.FireBurnPerWood()`; `fireFuelCap` is NOT promoted
(research R6) and still truncates the effective deadline.
Every tick `stepEvents` sweeps `Structures` for a fire
whose `FuelUntil` falls in the window `tick-1 < FuelUntil <= tick`, emitting
`sim.fire_burned_out` exactly once on that transition (no state effect — lit-ness
stays derived), plus a low-salience witness memory ("Watched the fire burn out.")
for living agents within `witnessRadius`.

**Needs-conditioned recovery** (spec 064, from TASK-101 spike Direction B):
`Intent` gains an OPTIONAL completion condition — `UntilNeed` (closed set
`warmth`/`rest`/`food`) and `UntilValue`, a need level — plus `HoldRef`,
the need level at the hold's anchor tick; all three
`omitempty`, so a conditionless intent (every pre-064 intent) marshals
byte-identically. With `UntilNeed` set, `executeAtTarget` intercepts BEFORE
the per-goal switch, handing off to `recoveryHoldEvents` instead of the goal's
default arrive-and-done: the intent HOLDS at its target (visibly recovering,
not idle), checked every tick against the live need — already-satisfied
completes at once, a threshold crossing completes normally (arming the spec-062
yield window iff the ring source `isMindSource`, so a reflex-issued recovery
never arms it), a higher-priority survival need (reflex-ladder order:
food > warmth > rest) crossing into ITS danger band ends the hold so the
agent re-decides (no new preemption immunity — a hold is LESS sticky than an
ordinary intent, never more), and no net gain over a `recoveryStallTicks`
(300, ~5 needs heartbeats) window aborts with the distinct `agent.recovery_stalled`
outcome (dead fire, displaced source, unreachable threshold) rather than
loitering forever. `warm_up` is the evidenced consumer — a planner tool
resolving exactly like `goto_warmth` but carrying the condition, its
optional `until_warmth` clamped (spec 058 clamp-with-notice posture) into
`[warmthRecoverFloor, needMax]` (`warmthRecoverFloor` = `dangerWarmthBelow`,
350; `needMax` 1000) via the single `clampWarmUp`/`ClampWarmUp` clamp home,
defaulting to doctrine constant `warmthRecoverTo` (800, a healthy margin
above the danger band) when absent — and the [[reflex-policy]] day AND night
warmth rungs (`reachKnownWarmth`) now issue the same conditioned `goto_warmth`
at the doctrine default, so a reflex-driven recovery also holds at the fire
instead of arriving, idling, and wandering off cold (the world-01
arrive-idle-vacuum, Direction B). `wakeReason` (US4, the audit's
Gap C) gains a matching cold-emergency wake arm — a sleeper whose warmth falls
below `exposureWakeBelow` (150 — exactly the hunger-emergency wake's shape
and magnitude, deliberately deviating from the plan's nominated 350: an
emergency floor, not the routine-dip danger band, so a sleeper isn't roused
merely for being cold) wakes only when night AND the reflex's warmth
ladder finds something actionable (the hunger wake's "food in hand" analog,
the churn bound); a cozy fire-side sleeper sleeps through untouched.
`wakeReason` now takes the state/map/tick that ladder check needs, not the
bare `(agent, night)` of before. Held-pinned villagers are excluded from the
emergent-gathering quorum ([[governance]]) — a survival hold is not an
elective assembly.

**Eating** (T018, `eatOutcome`): the reflex's `agent.ate` direct-event path (and the
planner's guarded-plan equivalent) now computes an outcome rather than a
bare marker. `eatOutcome` consumes the most-nutritious form first — `Meals` →
`FoodCooked` → `FoodRaw` — one unit at a time until `Needs.Food` reaches `satietyAt`
(900) or the inventory runs dry, returning `false` (nothing eaten, no event)
if already sated or carrying no food. Restores differ per form
(`mealRestore` 100, `foodCookedRestore` 80, `foodRawRestore` 40 — cooking roughly
doubles raw, a meal is the best food); the payload carries per-form consumed
counts plus the absolute post-eat food need, so the reducer never re-derives
arithmetic.
`wakeReason`'s hunger-emergency wake check now looks for *any* carried food form,
not just raw. `canGive` (the give-to-starving social rule) checks `Inv.FoodRaw`
specifically — raw is deliberately the form a subsistence village shares.

**Run end** (spec 044 US1): `stepEvents` opens with a terminal guard — an
ended world (`State.Ended`) returns nil, ever after; `stepEvents` being the
sim's only emitter, the single latch freezes simulated time while [[sim-loop]]
keeps serving reads. The batch ends with the matching detector: when every
villager living at the tick's start (`livingCount`, `state.go`) died within
it, the run is declared over with a `run.ended` event
(`RunEndedPayload{tick, deaths, final_cause}`) as the batch's LAST event —
after every same-tick `agent.died` (heartbeat or [[gru]]) and its witness
memories, so no sim event ever trails the declaration. The payload
carries the whole run's death history — the `State.Deaths` ledger plus this
batch's, `final_cause` being the last death's cause — so no consumer
(status, the [[morgue]]) ever scans the log. A `!s.Ended` check belts
the top-of-function guard: exactly once per world, ever.

## Connections

Parent note: [[executor]]. [[world-tuning]] owns the fire-fuel dials read
here; [[reflex-policy]] issues the conditioned `goto_warmth`/`warm_up`
executed here and owns the wake ladder `wakeReason` consults;
[[governance]] excludes held-pinned villagers from its quorum;
[[morgue]] consumes the run-end death history; `run.ended` freezes
[[sim-loop]] into the ended posture.

## Operational notes

The v2 economy adds a full crafting chain (wood/stone → planks/refined_stone
→ spears/shelter/oven) and a fire that goes cold unless refueled —
`whole_feature_test.go` and `food_fire_test.go` exercise the chain and the
fuel sweep end-to-end.
