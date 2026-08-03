---
name: executor-needs-survival
description: The per-minute needs heartbeat, fire fuel/warmth, eating, and the run-end death detector — how an agent's Needs decay, recover, and can reach zero. Needs-conditioned recovery holds (warm_up) and the spec-083 neglect detector split to [[executor-needs-recovery-and-neglect]]. Load for death causes, fuel/eat mechanics, or the run.ended contract.
kind: component
sources:
  - internal/sim/executor.go
  - internal/sim/agents.go
verified_against: 012f715f55d8d87317e601ad75686c599d277349
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
replay-safe) — EXCEPT under the spec-104 coalescing regime
(`AmbientCoalescing()`, a new spec-048 tuning dial), where `needsEmitDue`
(`executor.go`) records the event only on the K-minute checkpoint grid
(`(tick/60)%K == 0`) or when a danger-band/near-death/zero boundary crossed
vs. the last-EMITTED values, in either direction (`needsBoundaryCrossed`) —
non-emitted minutes still decay, derivedly, behind the `NeedsSyncTick`
watermark (`advanceNeedsMinute`, `internal/sim/advance.go`), sharing the
SAME `foldNeedsAbsolutes` fold the recorded arm uses, so guardian survival
watches and standing-order hysteresis fire AND re-arm at the same one-minute
latency either way; K=1 reproduces every-minute emission byte-for-byte, and
a legacy world (no tuning dial, or the dial resolved to 0) always emits
every minute. Death detection and the near-death memory below still run
EVERY minute regardless of the regime — a death or near-death entry is
itself a crossing, so its needs event always rides the same batch.

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

Spec 064's needs-conditioned recovery holds (`warm_up`, `recoveryHoldEvents`,
the cold-emergency wake arm) and spec 083's neglect detector
(`sim.neglect_detected`, `Agent.Neglect`) split into
[[executor-needs-recovery-and-neglect]].

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

Parent note: [[executor]]. [[executor-needs-recovery-and-neglect]] is this
note's own split-off child — recovery holds and the neglect detector.
[[world-tuning]] owns the fire-fuel dials read here; [[morgue]] consumes
the run-end death history; `run.ended` freezes [[sim-loop]] into the ended
posture.

## Operational notes

The v2 economy adds a full crafting chain (wood/stone → planks/refined_stone
→ spears/shelter/oven) and a fire that goes cold unless refueled —
`whole_feature_test.go` and `food_fire_test.go` exercise the chain and the
fuel sweep end-to-end.
