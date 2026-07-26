---
name: reflex-pathfinding
description: Child of [[reflex-policy]] — path.go's BFS geometry (fixed N/E/S/W neighbor order, FIFO frontier, deterministic tie-breaking), nearest/nearestAdjacentTo, and spec 041's knowledge-gated nearestKnown/nearestKnownAdjacentTo/nearestFrontier wrappers over the same search. Load for pathing/tie-break determinism questions or how a wall makes a tile impassable.
kind: component
sources:
  - internal/sim/path.go
verified_against: 9e8fb36e750dda15a633ba3ff8c44141f02debf2
---

# Reflex pathfinding

Child of [[reflex-policy]]: `path.go`'s pathfinding primitives, which every
resolver in [[reflex-goal-resolution]] and every SURVIVAL/PREP rung in
[[reflex-survival-rungs]]/[[reflex-prep-arbitration]] calls through.

Pathfinding (`path.go`, unchanged in its own geometry by spec 012, spec 032's
path-the-tile-improvement feature (a naming coincidence), or spec 041):
breadth-first search with **fixed neighbor order (N, E, S, W)** and FIFO
frontier, so shortest paths and nearest-match searches are identical on every
run. `nextStep` re-derives one hop per move from the shortest path (paths are
never stored in state — movement outcomes are evented, so replay needs no
path data); a standing wall (spec 032) makes its tile impassable via
[[executor]]'s `passable`, so BFS routes around walls with no change to
`path.go` itself — walls are just another obstacle the same search already
handles. `nearest` finds the closest reachable tile matching a predicate in
BFS order; `nearestAdjacentTo` finds a standing tile beside a resource —
chopping a tree, quarrying rock, drawing water, and (spec 032)
building/demolishing/repairing a wall all resolve through it. Spec 041 adds
three knowledge-gated wrappers on the SAME geometry, never touching the BFS
itself: `nearestKnown`/`nearestKnownAdjacentTo` layer a fresh-fact check onto
`nearest`/`nearestAdjacentTo`'s match closure (so knowledge-gated resolution
keeps every ground-truth search's exact tie-breaking), and `nearestFrontier`
(US4, Yamauchi-style) finds the closest reachable tile the agent's map marks
EXPLORED that 4-neighbors at least one UNEXPLORED in-bounds tile, decoding
the agent's `Explored` bitmap once per search — [[mental-maps]] owns the
bitmap and fact-freshness semantics these wrappers read. The escape clause
lets an agent standing on impassable terrain (pre-terrain saves) step out.

## Connections

Parent [[reflex-policy]] summarizes this note and links every sibling child;
[[reflex-goal-resolution]] is the primary consumer of `nearestAdjacentTo`/
`nearestKnownAdjacentTo`; [[mental-maps]] owns the `Explored` bitmap
`nearestFrontier` decodes; [[executor]] supplies `passable`.
