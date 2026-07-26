---
name: reflex-goal-resolution
description: Child of [[reflex-policy]] — resolveGoal's consumable/survival/exploration goals (eat, quarry, water, fire/shelter/oven builds, hand-crafts, refuel, cook, bathe, search, sleep/wander/goto_warmth, warm_up, talk_to/seek), one entry per goal naming its resolver, gate, and spec-041 knowledge-gating. See [[reflex-goal-resolution-structures]] for storage/wall/path goals.
kind: component
sources:
  - internal/sim/policy.go
  - internal/sim/path.go
verified_against: cffd9a79bbed61ccac573d97c6cf544565b40336
---

# Reflex goal resolution — consumables and exploration

Child of [[reflex-policy]]: `resolveGoal`'s per-goal catalog — the shared
target resolver every [[agent-mind]] planner goal and every reflex rung
resolves through to a coordinate. Covers the consumable, survival, and
exploration goals; [[reflex-goal-resolution-structures]] covers storage,
walls, axes, and paths. See [[reflex-policy]] for the spec 012/013/032/014
history that grew this vocabulary and [[reflex-pathfinding]] for the BFS
geometry every resolver below shares.

## resolveGoal's goal vocabulary (consumables and exploration)

`resolveGoal` grew from the original handful (`eat`, `forage`, `hunt`, `chop`,
`build_fire`, `build_shelter`, `sleep`, `goto_warmth`, `wander`, `talk_to`/`seek`)
to cover spec 012's full economy, still resolving every goal to a concrete
`Intent` or an error through the same `nearest`/`nearestAdjacentTo` helpers the
reflex uses:

- **`eat`** now refuses on two grounds — nothing to eat (`!hasAnyFood`) or already
  sated (`Needs.Food >= satietyAt`, 900) — so a planner-chosen eat is never wasted
  at the ceiling.
- **`quarry`** and **`collect_water`** are planner-only (never in the reflex
  ladder): both resolve via `nearestKnownAdjacentTo` (spec 041), the
  same beside-the-resource pattern `chop` uses, matching a KNOWN fresh
  `"rock"`/`"water_edge"` fact instead of ground-truth `worldmap.Rock`/
  `worldmap.Water` — knowing of neither fails honestly ("you know of no rock
  outcrops"/"you know of no water") before the search even runs; quarry
  depletion stays a ground condition layered on top (an outcrop's fact
  persists until US3's correction, the forage-overlay precedent).
- **`build_fire`** is unchanged: gated on `fireWoodCost` wood, resolved to the
  nearest `buildSite`.
- **`build_shelter`** is re-costed to `Planks` (`shelterPlankCost`, 8, was wood)
  and is planner-only now that the reflex dropped it.
- **`build_oven`** is new: gated on `recipeFor("build_oven")`'s inputs (refined
  stone plus planks, checked via `hasItems`) and resolved to a `buildSite` the
  same way as fire and shelter.
- **`craft_planks`**, **`craft_stone`**, and **`craft_spear`** are new hand-crafts
  that need no travel — each resolves to the agent's own tile once
  `recipeFor(goal)`'s inputs are satisfied.
- **`refuel_fire`** is the one goal both the reflex (`reflexRefuelIntent`) and the
  planner can choose (FR-020): it targets the nearest KNOWN fire (spec 041,
  `nearestKnown`) regardless of remembered lit state — a cold fire is relit
  on arrival, a dying one topped up. See
  [[executor]] for the fuel window (`s.FireBurnPerWood()`, `fireFuelCap`) the
  completion applies — the burn-per-wood side is also a spec-048
  [[world-tuning]] dial, `fireFuelCap` is not.
- **`cook`** targets the nearest station the agent KNOWS is valid (spec
  041): an oven fact, or a fire fact remembered lit — its remembered `Detail`
  (`FuelUntil` as last seen) still ahead of now, predicting burnout from the
  agent's own knowledge rather than reading the world's live fuel state; the
  fixed BFS neighbor order still makes the tie-break deterministic, and the
  station reached determines the output and duration (`food_cooked` vs.
  `meals`) at the executor.
- **`bathe`** is new and oven-only, gated on `recipeFor("bathe")`'s water/wood
  inputs — water's only v1 consumer; since spec 041 the oven itself must
  also be KNOWN (`knowsAnyFresh`/`nearestKnown`).
- **`search`** (spec 041 US4) is new: resolves to the nearest exploration
  frontier (`nearestFrontier` — an explored, passable tile adjacent to
  unexplored land, Yamauchi-style) regardless of resource kind; a fully
  explored reachable world fails honestly ("nothing left unexplored").
  Wander-class completion (below, [[executor]]) — the walk itself does the
  exploring.
- `sleep` and `wander` are unchanged. `goto_warmth` (spec 041) now resolves
  against `warmKnownPredicate` — a remembered-lit fire or a KNOWN shelter,
  never a live warmth read — failing honestly ("you know of no warm place")
  rather than falling through to build/chop when nothing known is warm.
- **`warm_up`** (spec 064 R3/FR-002) is the planner's warmth-RECOVERY verb,
  new and planner-only (not `ReflexEligible` — the reflex's own warmth rungs
  issue the equivalent conditioned `goto_warmth` themselves, above): target
  resolution is `goto_warmth`'s exactly (`warmKnownPredicate` + `nearest`),
  but the returned `Intent` carries a completion condition
  (`UntilNeed: "warmth"`), so it HOLDS at the fire and completes on warmth
  ([[executor]]) instead of arriving and finishing. The threshold rides in
  through the resolver's generic `qty` argument (the storage verbs'
  per-verb-use precedent): 0 defaults to the doctrine constant
  `warmthRecoverTo` (800); any other value is clamped into
  `[warmthRecoverFloor, needMax]` by the single `clampWarmUp` home
  (clamp-with-notice, the spec-058 posture) — the same clamp the mind
  handler's `ClampWarmUp` wrapper consults to phrase the model-facing notice,
  so the two can never drift.
  `talk_to`/`seek` (spec 041, T013) resolves to the target's LAST KNOWN
  sighting (`peerSightingOf`, the mental map's peer record) rather than the
  target's live coordinates — a stale sighting walks honestly to where the
  target was last seen, and the landing/arrival guards
  (`GuardTargetPresent`) cover a miss; liveness (`Dead`) stays a live check,
  since death-knowledge honesty is beyond this feature's place-fact scope.

## Connections

Parent [[reflex-policy]] summarizes this catalog and links every sibling
child; [[reflex-goal-resolution-structures]] covers the storage/wall/path
goals; [[reflex-pathfinding]] is the BFS/`nearest`/`nearestAdjacentTo`
geometry every resolver above uses; [[mental-maps]] is the knowledge store
gating every KNOWN-prefixed lookup; [[agent-mind]] is the planner that
chooses these goals; [[executor]] applies the completions each goal's
`Intent` names.
