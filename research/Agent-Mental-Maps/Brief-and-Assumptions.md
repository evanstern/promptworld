---
title: Brief and Assumptions
aliases: []
tags: [brief]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Mental-Maps]]"]
---

# Brief and Assumptions

## The request (restated)

> "On a 64x64 map we should be able to build some sort of tree structure to organize mental
> maps. I'd like to research this strategy a bit. Build a vault in our research about how
> mental maps of grid-space can be built. Spend special attention to algorithms which work in
> 3 dimensions as well as that's something we'll be adding (layered grids) but not for a while."

## Context (from the codebase, verified 2026-07-24)

- **promptworld** is a Go simulation: 8 LLM-driven villager agents on a 64x64 tile grid.
- Agents are currently **omniscient for navigation**: verb tools ("forage", "refuel_fire")
  resolve to the nearest matching tile via full-map BFS (`internal/sim/policy.go`,
  `internal/sim/path.go`); all village structures appear with global coordinates in every
  planner prompt (`internal/mind/prompt.go`). There is no fog of war and no per-agent
  spatial knowledge.
- Locality **does** already exist in non-spatial layers: witness memories (Manhattan radius 8),
  situated memories carrying `{X, Y, Desc}`, rumors via adjacency talk, beliefs with
  witnessed/told/inferred provenance and confidence half-life decay.
- The goal is **per-agent private mental maps** replacing omniscient resolution, organized by
  some tree structure; later the world gains **multiple stacked grid layers** (a constrained
  form of 3D), so structures/algorithms should have a credible 3D extension.

## Assumptions made

- "Tree structure" is a direction, not a mandate for a specific structure — quadtrees are the
  obvious 2D candidate (octrees in 3D), but the research also covers alternatives (occupancy
  grids, topological/landmark graphs, hierarchical pathfinding abstractions) so the eventual
  analysis can compare.
- Scale envelope: 64x64 = 4,096 tiles now; "layered grids" assumed to mean a small number of
  stacked 64x64 layers (order 2–10), not full voxel worlds — but voxel-scale techniques are
  in scope as the 3D upper bound.
- 8 agents each holding a private map is the memory envelope; techniques should stay cheap at
  ~10x that (dozens of agents), since agent count may grow.
- Mental maps must tolerate **staleness** (the world changes while the agent is elsewhere), so
  belief decay / stale-knowledge handling is in scope.
- Go implementation is the target, but this branch stays language-neutral; no library survey
  was requested.

## Open questions (flagged, not answered here)

- Should the mental map gate *target resolution only* (agent must know a fire to path to it),
  or also *pathfinding* (agent can only path through remembered-passable tiles)?
- Do agents share/merge maps when they talk (rumor-style spatial gossip), or stay fully private?
- How does the LLM consume the mental map — rendered text summary, read-only query tools, or
  both?
- For the future 3D: are layers fully independent grids connected at portals (stairs/ladders),
  or a true voxel volume?

## Grounding

- [[_grounding]] — the research pass this brief scoped
