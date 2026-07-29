---
name: reflex-policy-history
description: Child of [[reflex-policy]] — resolveGoal's spec-by-spec vocabulary growth (012 economy, 013 storage, 032 walls/axes/paths, 014/TASK-53's goalResolvers table restructuring) and spec 041's knowledge-gating rewrite (every place-targeting resolver now searches the agent's own fresh facts, never ground truth). Load for the "why" behind the current goal catalog.
kind: component
sources:
  - internal/sim/policy.go
  - internal/sim/path.go
verified_against: fc9566d527941d3950fdd307168556820bd0875b
---

# Reflex policy — resolveGoal's spec history

Child of [[reflex-policy]]: the spec-by-spec history of `resolveGoal`'s
growth and the spec-041 rewrite of what every resolver searches. The
current per-goal catalog lives in [[reflex-goal-resolution]] (consumables/
exploration) and [[reflex-goal-resolution-structures]] (storage/walls/
paths); this note explains how it got there.

## How it works

**Vocabulary growth**: Spec 012 widened `resolveGoal`'s goal set
considerably (quarrying, water, crafting, an oven, refueling, cooking,
bathing) while trimming the reflex ladder itself down to one addition —
refueling a dying fire — and one removal — shelter-building dropped out of
the reflex entirely once it was re-costed in planks. Spec 013 (inventory &
storage v1) widened `resolveGoal` again — a chest to build, goods to
drop/pick up/deposit/withdraw — and left the reflex ladder itself
completely untouched: all five new goals are planner/plan-only (FR-014),
added to `resolveGoal` but never reachable from `decideIntent`. Spec 032
(walls, axes, paths) widened it once more — two wall builds, a
demolish/repair pair on an existing wall, a fourth hand-craft
(`craft_axe`), and a path build — every one planner/plan-only, same
pattern: the reflex ladder itself gains nothing from spec 032 (an axe or a
wall is never something `decideIntent` reaches for on its own). Spec 014
(TASK-53) restructured `resolveGoal` from one large switch into
`goalResolvers`, a name-keyed resolver table with the old per-verb bodies
verbatim — the [[tool-registry]]'s boot-time coverage gate
(`sim.ValidateToolCoverage`) asserts every World tool on the villager
roster has a table entry, so a registered verb can never lack its
resolver. The plan-step accept set that once lived beside it
(`planGoals`) is gone: the sim door now derives it from the registry
([[sim-loop]]).

**Spec 041's knowledge-gating rewrite** ([[mental-maps]]): changed WHAT
every resolver above searches, not the search mechanics: every goal that
targets a place now resolves against the acting agent's own FRESH facts
(`nearestKnown`/`nearestKnownAdjacentTo`, `path.go` — knowledge-gated
twins of `nearest`/`nearestAdjacentTo` that keep the identical BFS
geometry and tie-breaking, only the match closure differs), not the
world's ground truth. A resolver that knows of nothing of the right kind
fails with an epistemic reason ("you know of no forage") rather than
resolving to a place it has never perceived. Availability that is not
itself place knowledge — a harvested forage spot, a cooling den, wall
damage, chest contents, quarry depletion — stays layered on top as an
ordinary ground condition: the agent knows the PLACE and walks there;
whether the walk pays off is checked at arrival, exactly as before. A new
goal, `search` (US4), resolves to the nearest exploration frontier
(`nearestFrontier`) instead of any resource — the deliberate answer to
knowing of nothing. `talk_to`/`seek` resolves to the target's last KNOWN
sighting (the mental map's peer record), never its live coordinates.

## Connections

Parent [[reflex-policy]] summarizes this history and links every sibling
child. [[reflex-goal-resolution]] and [[reflex-goal-resolution-structures]]
hold the resulting per-goal catalog this history explains; [[reflex-pathfinding]]
is the BFS geometry spec 041 kept unchanged; [[tool-registry]] owns the
boot-time coverage gate the TASK-53 restructuring depends on; [[sim-loop]]
derives the plan-step accept set from the registry; [[mental-maps]] owns
the knowledge store spec 041 gated every resolver against.
