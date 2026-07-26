---
name: mental-map-perception
description: Child of [[mental-maps]] — how the map is READ for target resolution (nearestKnown/nearestKnownAdjacentTo, talk_to/seek) and GROWN by the perception sweep (agent.saw, agent.map_corrected); the search goal/frontier fallback; graves as a perceived fact kind. Load for what a resolver sees, how a fact gets witnessed or corrected, or the search verb's mechanics.
kind: component
sources:
  - internal/sim/policy.go
  - internal/sim/path.go
  - internal/sim/executor.go
  - internal/tool/registry.go
verified_against: 93837e1885bff17114df75e5382ac60dee24776a
---

# Mental map perception and resolution

Child of [[mental-maps]]: how the knowledge [[mental-map-model]] stores gets
read by target resolution and grown by perception — the search-goal frontier
fallback, the perception sweep that witnesses and corrects facts, and graves
as a perceived fact kind. [[mental-map-propagation]] covers how facts move
between agents instead of being perceived directly.

**Resolution** ([[reflex-policy]], `internal/sim/path.go`/`policy.go`): every
resolver that targets a place now searches the acting agent's fresh facts, not
ground truth. `nearestKnown`/`nearestKnownAdjacentTo` (`path.go`) are
knowledge-gated twins of `nearest`/`nearestAdjacentTo` — the identical BFS
geometry and tie-breaking, only the match closure differs, so
"nearest known" keeps every ground-truth search's determinism. Availability
that is not itself place knowledge — a harvested forage spot, a cooling den,
wall damage, chest contents, quarry depletion — stays layered on top as an
ordinary ground condition, checked at arrival exactly as before; a fully
resolved goal that fails on arrival re-validates the same way any contested
resource always has. `talk_to`/`seek` resolves to the target's last KNOWN
sighting (`peerSightingOf`) — a stale sighting walks honestly to where the
target was last seen, and the landing/arrival guards (`GuardTargetPresent`)
cover a miss; liveness (`Dead`) stays a live check.

**The search goal** (US4, research D4): `search` is a new World tool
(`internal/tool/registry.go`, `Effect: World, Gate: Resolvable, Cost.DurationTicks:
0, PlanStep: true, ReflexEligible: true`, appended after `build_path` so no
existing tool's registration position shifts) resolving to `nearestFrontier`
(`path.go`) — the closest reachable tile the agent's map marks EXPLORED that
4-neighbors at least one UNEXPLORED in-bounds tile (Yamauchi-style), decoding
the explored bitmap once per search; not found means the reachable world is
fully explored, the search verb's honest exhaustion. Completion is
wander-class ([[executor]]) — instant on arrival, since the walk itself did
the exploring (movement marks explored terrain and the perception sweep
witnesses what's there). The reflex ladder ([[reflex-policy]]) falls back to
`search` on the hungry rung ONLY when the agent knows of no forage and no
ready den — hunger-only, so a fed villager never mounts an expedition just to
top up the larder. Spec 062 (US3, 057 audit Gap A) adds a second,
independent reflex call site: a cold NIGHT with no known warmth, insufficient
wood, and no known tree to chop also falls back to `search` — one rung above
terminal sleep in [[reflex-policy]]'s bounded frontier-search fallback — so
`nearestFrontier` now backs two separate reflex triggers (hunger, and
cold-with-nothing-left-to-try), each still bounded by the same
fully-explored-fails-honestly floor.

**The perception sweep** (`internal/sim/executor.go`, `perceptionEvents`, T007):
each awake living villager, on the same staggered per-agent cadence movement
uses (a fifth of a full per-tick sweep, T034's hot-path relief), diffs ground
truth within `witnessRadius` against its map and emits at most one
`agent.saw` (new/changed structures, piles, standing trees, unharvested
forage, unquarried rock, water shoreline, dens — fully baked, `Seen` = this
tick, provenance `witnessed`) and, when a remembered fresh fact is genuinely
ABSENT from ground truth (`groundFactPresent` — a chopped tree, a
quarried-out outcrop, a drained pile, a removed structure; a merely-harvested
forage spot or cooling den still exists, only its availability lapsed), one
`agent.map_corrected` (US3, T019) naming the gone facts. A correction's gone
facts each ride a companion situated first-person discovery memory
(`mapCorrectedText`, `salMapCorrected` = 5) in the same batch as
`agent.memory_added` — memories accrete only via that event, never appended
directly by a reducer arm (a deviation from data-model.md's "reducer stamps a
situated memory" phrasing, recorded for the planning tier). `agent.saw` is
digest-only, deliberately no chronicle line (too chatty) and not an absorb
trigger; `agent.map_corrected` IS an absorb trigger — [[agent-mind]]'s `absorb`
re-arms the planner only when a removed fact matches the agent's OWN current
intent target or resolved coordinates, so a correction elsewhere in the map
stays quiet, carried into the next scheduled round as a memory instead.

**Graves** (spec 044 US4, [[morgue]]): a death leaves a persistent marker —
the `agent.died` reducer arm ([[sim-state-reducer]]) appends
`Structure{Kind: "grave"}` at the death tile, the same reducer-internal idiom
as the inventory spill, unconditionally (the `Structures` slice has no
per-tile uniqueness invariant outside the `buildSite` gate on NEW builds, so
a grave coexists with whatever already stands there; appended last, it wins
the map view's per-tile glyph, and it blocks future building via `buildSite`'s
blanket any-structure check). No new knowledge machinery: `grave` is simply a
new entry in the closed vocabulary [[mental-map-model]] owns (`placeFactKinds` in
`internal/tool/registry.go` mirrors it for `send_vision`'s `place_kind` Enum),
so the ordinary perception sweep witnesses a grave, talk can pass it on, a
vision can reveal it, and the prompt's landmark set names it individually.

## Connections

Parent [[mental-maps]] summarizes this note and links every sibling child;
[[mental-map-model]] is the type this note reads and grows; [[mental-map-propagation]]
covers facts arriving by telling or divine grant instead of perception;
[[reflex-policy]] is the primary resolver consumer; [[executor]] hosts
`perceptionEvents`; [[agent-mind]]'s `absorb` re-arms on a targeted
`agent.map_corrected`; [[morgue]] is the death flow that plants a grave;
[[tool-registry]] declares the `search` tool and `placeFactKinds`.
