---
name: sim-state-apply-agents-resources
description: Child of [[sim-state-apply-agents]] — sim.State.Apply's v2 crafting/gather-yield arms (quarried/collected_water/crafted/cooked/bathed/refueled/spear_broke, axe-vs-bare yields, removeHarvestedFact), the v3 storage events (dropped/picked_up/deposited/withdrew, food_rotted, chest_taken), and the spec-032 wall demolish/repair HP family.
kind: component
sources:
  - internal/sim/agents.go
  - internal/sim/recipes.go
  - internal/sim/terrain.go
verified_against: c61cd6c04ddfcd2a976c14a49ba071e8fd768a73
---

# Sim state: agent resource, storage, and wall Apply arms

Split from [[sim-state-apply-agents]] (summary-style, corpus-spec v2): the
`Apply` arms for gathered/crafted resources, the v3 storage economy, and
the spec-032 standing-wall HP family — as distinct from the parent's
genesis/clock/intent/movement/needs/death arms.

## How it works

The v2 resource/crafting events (`agent.quarried`/`collected_water`/
`crafted`/`cooked`/`bathed`/`refueled`/`spear_broke`, `sim.fire_burned_out`)
apply inventory deltas and structure/overlay changes, several by
re-deriving the recipe from `recipes.go` (the single source for craft/build
magnitudes — a duplicated number here would drift from the contract
table), and — since spec 013 — clamp their gather yields to the taker's
free bulk (`bulkCap − bulk(Inv)`); since spec 032 US2, `agent.chopped`/
`agent.quarried`'s own yield is no longer a single flat constant — it's
`chopYieldBare`/`quarryYieldBare` (1) bare-handed or `chopYieldAxe`/
`quarryYieldAxe` (3) carrying an axe, re-derived from the SAME
pre-mutation state the emitter checked (the spear-hunt precedent),
spending `Axes[0]`'s last use; a spent-to-zero axe's removal rides its own
companion `agent.axe_broke` (an `agent.spear_broke` clone), not this
payload; since spec 081 both arms ALSO call `removeHarvestedFact` (the
felled `tree` / quarried `rock` place-fact leaves the actor's map and
every awake in-radius witness's map at the act event — provenance-blind,
silent, [[mental-map-perception]]), positions read from the same
pre-mutation state as the yield derivation.

The v3 storage events (spec 013 US2/US3/US5) move goods between an
agent's `Inv` and a `Pile`/chest `Store`: `agent.dropped`/`agent.picked_up`
create-or-merge or drain a tile's pile (food oldest-batch-first, spears
AND axes (spec 032 US2) most-worn-first), `agent.deposited`/
`agent.withdrew` do the same against a chest's `Store`, and
`sim.food_rotted` drains a pile's spoiled food batches (`SpoilAt ≤ tick`)
— every one of these defensively re-clamps to what's actually
carried/held/available, so the reducer stays total even against a
contested or forged event, and an emptied pile is removed in the same
application; `social.chest_taken` is an effect-free record (its
consequences — the reason-`"theft"` `social.relation_changed` and the
owner/witness `agent.memory_added` events — ride the same companion
batch, [[social-fabric]]).

The wall family (spec 032 US1) maintains a standing wall's `HP`:
`agent.built`'s generic arm stamps a fresh `wall_plank`/`wall_stone` at
`HP = wallMaxHP(kind)` — full health, after the same recipe-delta spend
every structure gets — and three dedicated arms carry the multi-cycle
demolish/repair loop: `agent.wall_chipped` decrements `HP` by
`demolishChipHP` (clamped to never go below 1 — a standing wall never
serializes ≤ 0) and resets the demolisher's `Intent.WorkStart` to 0,
re-arming the executor's work gate for the next cycle under the SAME
intent (no new scheduling); `agent.wall_destroyed` (the final chip)
removes the structure — its tile is passable again by construction — and
clears the intent (`guardian.entity_removed` reaches the same end through
the miracle path); `agent.wall_repaired` consumes 1 unit of
`wallRepairMaterial(kind)` (planks for a plank wall, refined stone for a
stone wall — the same material each was built from) and restores `HP` by
`repairHPPerUnit`, clamped to the max, re-arming the work gate the same
way when still damaged with material in hand, otherwise clearing the
intent — `isWall`/`wallMaxHP`/`wallAt` (`terrain.go`, a `chestAt` sibling)
back every one of these arms, plus the movement `passable` check a
standing wall now fails.

## Connections

Parent [[sim-state-apply-agents]] summarizes this note and its sibling
arms. [[social-fabric]] owns the theft consequence batch and the chest
economy; [[mental-map-perception]] owns the harvested-fact removal these
chop/quarry arms trigger; [[executor]] owns the work-gate re-arming the
wall/demolish/repair cycle depends on; [[event-types]] catalogs every
payload shape here.
