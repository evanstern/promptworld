---
name: reflex-survival-rungs
description: Child of [[reflex-policy]] — the five SURVIVAL rungs of decideIntent (eat, get food + US4 frontier-search fallback, the night warmth ladder + its own frontier-search fallback and terminal sleep, daytime nap, the day warmth rung). First-match-wins, unconditioned by the PREP yield gate, always run first. Load for eat/warmth/sleep/nap rung mechanics and their spec-041/062 knowledge-gating.
kind: component
sources:
  - internal/sim/policy.go
verified_against: 22bb41c887ef6a34c55a77b9b989b299f4dc6857
---

# Reflex survival rungs

Child of [[reflex-policy]]: the life-saving half of `decideIntent`'s arbitration
doctrine — `survivalDecision`, which runs FIRST and is **unconditioned** by the
PREP yield window or danger bands ([[reflex-prep-arbitration]]) — a life-saving
reflex never defers. See [[reflex-policy]] for the spec-062 restructuring that
carved this out as its own function and the thrash bug (world-01's
forage↔goto_warmth thrash, Sage: 436 flips) it fixes.

### SURVIVAL rungs (first match wins, always run)

1. **Eat** — hungry (`Food < hungryAt`, 350) and carrying any edible unit
   (`hasAnyFood`: `Inv.Meals + Inv.FoodCooked + Inv.FoodRaw > 0`) → instant
   `agent.ate`. The triplet check replaces the old raw-food-only check (T018) so an
   agent carrying only meals or only cooked food still eats reflexively.
2. **Get food** — hungry, nothing carried → `foodIntent`: nearest KNOWN
   fresh forage fact (spec 041 — availability still checked as a ground
   condition), else nearest KNOWN ready den (`hunt`); knowing of neither
   (US4, T026, FR-013 parity without omniscience) falls back to the nearest
   exploration frontier (`search`) — hunger-only, so a fed villager never
   mounts an expedition just to top up the larder.
3. **Night, cold** (`!warmAt`) — the night `warmthLadder` (spec 062 R5, the
   exact pre-062 body factored into a shared helper): `reachKnownWarmth`
   (reach a remembered-lit fire, `goto_warmth` — since spec 064, WITH a
   needs-conditioned completion (`UntilNeed: "warmth", UntilValue:
   warmthRecoverTo`), so the agent HOLDS at the fire and actually warms up
   instead of arriving, idling, and wandering off cold ([[executor]]'s
   `recoveryHoldEvents`; the world-01 arrive-idle-vacuum this spec kills from
   the recovery side, 062 having already killed it from the scheduling side)
   — via `warmKnownPredicate`; else
   `reflexRefuelIntent`, T020/FR-012, relighting/topping up a KNOWN cold or
   dying fire when carrying wood — cheaper than a fresh build) → build a fire
   with `fireWoodCost` (2) wood already in hand (`buildWarmthIfWood`) → chop
   the nearest KNOWN standing tree (yes, chopping in the cold dark — the
   day-1 drama working as designed). **New (US3, FR-005, 057 audit Gap A)**:
   if the ladder finds NOTHING — carrying under `fireWoodCost` wood AND no
   KNOWN tree to chop — rather than lying down cold, the agent searches
   toward the nearest exploration frontier (`nearestFrontier`, the
   hungry-search shape) one rung above terminal sleep; only when no frontier
   is reachable either does it fall through to sleep (today's floor of the
   fallback, `reflex_matrix_test.go`'s Gap-A cells: wood=0 cells now resolve
   to `search`, was `sleep`).
4. **Night, warm** — sleep where you stand.
5. **Exhausted by day** (`Rest < tiredAt`, 250) — nap, preferring a warm tile.
6. **Day warmth rung** (spec 062 US2, FR-004, 057 audit Gap B) — a
   cold-but-not-tired villager by day (`Needs.Warmth < dangerWarmthBelow`,
   350, and not already standing in warmth) runs `dayWarmthLadder`: the SAME
   `reachKnownWarmth` (since spec 064, holding conditioned) → `buildWarmthIfWood`
   rungs the night ladder uses (`R5`,
   shared helpers, no drift), BEFORE any PREP rung — so "Sage forages while
   freezing" becomes impossible at the reflex layer. **Deliberately omits**
   the night ladder's chop tail (a flagged plan deviation, recorded in
   `dayWarmthLadder`'s doc comment): built as the full night ladder including
   chop, the chop trip's ~300 ticks/trip stole a marginal villager's daytime
   larder-stocking time and starved a sleeper on seed 101 (regressing
   `TestDegradedModeVillageSurvives*` 8/8→7/8); since warmth passively
   regenerates by day (never a death spiral, unlike night), trekking to chop
   firewood for daytime warmth is unjustified subsistence-time theft —
   `TestDayWarmthDoesNotChopTheDeviation` (`day_warmth_test.go`) pins the
   no-chop case. The night branch keeps chop (night warmth IS a death spiral).

## Connections

Parent [[reflex-policy]] summarizes this rung group and links every sibling
child; [[reflex-prep-arbitration]] is the yield-gated other half of
`decideIntent`; [[executor]] hosts `recoveryHoldEvents` and the fire-fuel
mechanics these rungs key on; [[mental-maps]] is the knowledge store every
KNOWN-prefixed lookup above reads through.
