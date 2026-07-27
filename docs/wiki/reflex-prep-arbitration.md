---
name: reflex-prep-arbitration
description: Child of [[reflex-policy]] — the PREP yield gate (spec 062, "instinct yields to intelligence": yield window off LastMindIntentDone, danger-band override), the PREP rungs it guards (fire-knowledge build/chop, refuel, larder stock-up), and the wander/waking fallback. Load for prepYields mechanics, the doctrine-home danger constants, or why prep never fires mid-danger.
kind: component
sources:
  - internal/sim/policy.go
verified_against: fc9566d527941d3950fdd307168556820bd0875b
---

# Reflex PREP arbitration and wander

> Since spec 084, the DIRECTIVE rung (`directiveDecision`,
> [[guardian-designations]]) sits BETWEEN survival and this gate: a hard
> guardian directive preempts PREP and wander whenever it resolves, and is
> deliberately NOT gated by `prepYields` — the yield window exists so
> instinct doesn't counter-schedule the MIND, and a directive is the
> villager's current duty, not instinct noise (the planner sees the same
> directive through the context block, so rung and planner pull together).

Child of [[reflex-policy]]: the opportunistic-upkeep half of `decideIntent`'s
arbitration doctrine — `prepDecision`, which runs only when the yield gate
below says no — plus the idle `wanderDecision` filler. See
[[reflex-survival-rungs]] for the life-saving rungs that always run first and
are exempt from this gate.

### The PREP yield gate (spec 062 US1, FR-002/FR-003 — "instinct yields to intelligence")

Before `prepDecision` runs, `prepYields(s, a, tick)` (`agents.go`) checks two
independent clauses, either of which holds prep back:

- **Yield window**: `a.LastMindIntentDone != 0 && tick-a.LastMindIntentDone <
  prepYieldTicks` (1800 — one default planner cadence, deliberately the
  CONSTANT and not the tuned `PlannerCadence()` dial: the window is
  arbitration doctrine, not scheduling). `LastMindIntentDone` (`Agent`,
  `omitempty`) is the tick the agent's most recent NON-REFLEX intent
  completed — armed ONLY by the `agent.intent_done` reducer arm
  ([[sim-state-reducer]]), and only when the closed `IntentRecord`'s
  `Source` is `isMindSource` (`planner`/`plan`/`meeting`); `stampIntentOutcome`
  now returns that source so the arm can read it. A reflex completion never
  arms the window — instinct yielding to itself would deadlock prep in a
  no-planner world, so the sentinel stays a permanent 0 there and degraded
  mode never suppresses on this clause (SC-003, FR-007).
- **Danger band**: any need below its band — `dangerFoodBelow`/
  `dangerWarmthBelow`/`dangerRestBelow` (`agents.go`, doctrine-home constants
  beside the existing thresholds, R3), anchored EXACTLY at the existing
  survival-rung triggers with no padding: `dangerFoodBelow = hungryAt` (350),
  `dangerWarmthBelow = coldNightBelow` (350), `dangerRestBelow = tiredAt`
  (250) — "in danger" == "a survival rung would or will imminently act". The
  warmth band is the one that bites by day: it is what suppresses prep for a
  villager recovering AT a fire (`warmAt`, so the day-warmth rung above is
  skipped) whose warmth is still low.

Both clauses are named, dial-ready constants (FR-006) in a single doctrine
home, deliberately NOT `tuning.json` dials (earned by evidence, not
speculative). SURVIVAL rungs are exempt from both — they decide before this
gate is ever consulted. `yield_state_test.go` covers the window's arm/decay,
the danger-band override, reflex-never-arms, and the no-LLM parity drive
(SC-003: a planner-free reflex matches pre-062 intents except where a danger
band newly suppresses prep).

### PREP rungs (only when `prepYields` says no)

7. Knows of no fire (spec 041: `!knowsAnyFresh(a, "fire", tick)`, the agent's
   own belief rather than `!hasStructure("fire")` — a village fire someone
   else built and this agent has never seen still triggers this rung for
   THIS agent) → build/chop toward one.
8. `reflexRefuelIntent` again, unconditionally, to keep a known fire from
   dying down.
9. Stock the larder to `stockFoodRawTo` (8) units of raw food (`Inv.FoodRaw`).
   Shelter-building is gone from this ladder (T020): since spec 012 re-costed
   it in `Planks` (`shelterPlankCost`, 8) instead of raw wood, it joined the
   crafting economy and became planner-only — the reflex never enters
   `resolveGoal`'s `build_shelter`, `craft_*`, or `build_oven` cases.

### Wander

10. Neither group decided: a seeded short stroll
    (`rngAt(seed, "wander", tick, idx)`).

Waking (`wakeReason` in executor.go) mirrors this: day + decent rest
(`Rest >= 600`), or a hunger emergency the agent can act on — `Food < 150` and
`hasAnyFood`, the same triplet check as the eat rule above. Fully-rested agents
sleep through the night — the live-run sleep/wake churn bug is documented in the
TASK-5 notes. Actually eating the food (most-nutritious form first — `Meals` then
`FoodCooked` then `FoodRaw`, stopping at `satietyAt`) is the executor's
`eatOutcome`, detailed in [[executor]].

## Connections

Parent [[reflex-policy]] summarizes this rung group and links every sibling
child; [[reflex-survival-rungs]] is the unconditioned other half of
`decideIntent`; [[sim-state-reducer]] owns `Agent.LastMindIntentDone` and the
`agent.intent_done` arm that arms the yield window; [[executor]] hosts
`wakeReason` and `eatOutcome`.
