---
title: Extending to 3D and Layered Grids
aliases: [Multi-Level Grids]
tags: [3d, octree, layers, floors]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Mental-Maps]]", "[[Hierarchical-Spatial-Trees]]", "[[Occupancy-Grids-and-Belief-Maps]]", "[[Topological-and-Landmark-Maps]]", "[[Hierarchical-Pathfinding-Over-Known-Space]]"]
---

# Extending to 3D and Layered Grids

What the literature says about lifting 2D grid mental maps into the third dimension — from
stacked layers to full voxel volumes.

## The spectrum of 3D-ness

1. **2.5D elevation maps** — a 2D grid with a continuous height value per cell. Cheapest, but
   "unsuitable for scenarios with overhanging objects or multi-floor structures"
   ([Point Cloud Tomography](https://arxiv.org/pdf/2403.07631), [[_grounding]] §9).
2. **Layered grids / floors** — independent 2D grids per level, connected at discrete portals
   (stairs, ladders, elevators). The dominant model for buildings in both games and robotics
   ([[_grounding]] §9).
3. **Quadtree with 3D-annotated leaves** — keep a 2D quadtree, add layer/height data to leaf
   nodes; documented as acceptable "when the additional dimension … is not as detailed or
   complex as the other dimensions" ([[_grounding]] §1, [[Hierarchical-Spatial-Trees]]).
4. **Full voxel octrees** — OctoMap-style: equal-sized voxels in an octree with per-leaf
   log-odds occupancy, explicitly modeling occupied/free/unknown 3D space, with child-merging
   compression ([Hornung et al. 2013](https://link.springer.com/article/10.1007/s10514-012-9321-0),
   [[_grounding]] §3).

## The layered-grid pattern (games and robotics converge)

- Game practice (Software Inc.): a **layered pathfinding algorithm** that "rather than
  measuring distance geometrically … asks 'what is the shortest amount of doors/elevators to
  traverse to get from room A to room B?'"
  ([ModDB](https://www.moddb.com/games/software-inc/features/pathfinding-on-multiple-floors),
  [[_grounding]] §9).
- Robotics: "transitions between floors don't require highly accurate 3D maps since they are
  typically connected only by elevators or stairs, so environments can be modeled as a
  **floor-stair topology**" ([Multi-Floor ZS Object Navigation](https://arxiv.org/pdf/2409.10906),
  [[_grounding]] §9). Stair regions are detected and represented as traversable waypoint areas
  ([Stairway to Success](https://arxiv.org/pdf/2505.23019), [[_grounding]] §9).
- Grid-game implementations build node grids per level and link them at stair nodes
  ([3D-A-Star-Pathfinding](https://github.com/olokobayusuf/3D-A-Star-Pathfinding),
  [[_grounding]] §9); multi-floor explorers combine per-floor metric maps with a cross-floor
  topological graph ([LITE](https://arxiv.org/pdf/2507.21517), [[_grounding]] §9).

The layered pattern is structurally the hybrid of [[Topological-and-Landmark-Maps]] (the
inter-layer portal graph) and per-layer metric maps ([[Occupancy-Grids-and-Belief-Maps]]),
and it is the same abstraction move as HPA\* clusters
([[Hierarchical-Pathfinding-Over-Known-Space]]) with layers as clusters and portals as gates.

## What carries over unchanged from 2D

- **Cell semantics**: log-odds occupied/free/unknown is identical in 2D grids, quadtrees, and
  octrees — OctoMap is literally the 2D occupancy model in an octree ([[_grounding]] §2, §3).
- **Tree shape**: quadtree → octree changes only the branching factor (4 → 8)
  ([[_grounding]] §1).
- **Frontiers and fog-of-war states** are defined on cell adjacency and lift to any dimension
  ([[_grounding]] §5, §8).

## Grounding

- [[_grounding]] — §9 (multi-level navigation), §3 (OctoMap), §1 (quadtree/octree relation)
- [Hornung et al. — OctoMap](https://link.springer.com/article/10.1007/s10514-012-9321-0)
- [ModDB — Software Inc. multi-floor pathfinding](https://www.moddb.com/games/software-inc/features/pathfinding-on-multiple-floors)
