---
title: Hierarchical Pathfinding Over Known Space
aliases: [HPA-star, Quadtree Pathfinding]
tags: [pathfinding, hpa-star, quadtree]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Mental-Maps]]", "[[Hierarchical-Spatial-Trees]]", "[[Extending-to-3D-and-Layered-Grids]]", "[[Memory-Footprint-for-Many-Private-Maps]]"]
---

# Hierarchical Pathfinding Over Known Space

How the tree/cluster structure that stores a mental map can also *plan* over it — the two uses
share one hierarchy.

## HPA* — cluster abstraction

HPA\* (Botea et al. 2004) "abstracts a map into linked local clusters": optimal crossing
distances within each cluster are precomputed and cached; global search then "traverses
clusters in a single big step." Paths come out **within 1% of optimal**, and search is
"considerably faster than A\* in all cases"; the cost is a preprocessing phase that grows with
the number of abstraction levels
([Botea & Müller](https://www.researchgate.net/publication/228785110_Near_optimal_hierarchical_path-finding_HPA),
[A\* vs HPA\* analysis](https://www.researchgate.net/publication/272551633_The_Analysis_of_Efficiency_Dependence_of_the_Shortest_Path_Finding_Algorithms_A_and_HPA),
[[_grounding]] §7). The hierarchy extends beyond two levels: small clusters group into larger
ones, reusing the contained clusters' crossing distances ([[_grounding]] §7). A maintained
open-source implementation exists, benchmarked on Dragon Age: Origins maps
([hugoscurti/hierarchical-pathfinding](https://github.com/hugoscurti/hierarchical-pathfinding)).

## Quadtree pathfinding — the tree is the graph

On a quadtree map, "a section … contains no obstacles or all obstacles"; adjacent leaves
connect through **gates** (pairs of adjacent cells, one per side), and A\* or flow fields run
over the leaf-adjacency graph ([hit9/QuadtreePathfinding](https://github.com/hit9/QuadtreePathfinding),
[[_grounding]] §7). Reported speedup on an indoor benchmark: quadtree A\* "needs only 2% of
the time" of regular-grid A\* (0.75 s → ~0.015 s)
([Using Quadtrees for Realtime Pathfinding](https://www.researchgate.net/publication/235443232_Using_Quadtrees_for_Realtime_Pathfinding_in_Indoor_Environments),
[[_grounding]] §7). The general principle: "the fewer nodes in your map representation, the
faster A\* will be," and pathfinding cost scales worse than linearly with distance
([Amit Patel — Map Representations](http://theory.stanford.edu/~amitp/GameProgramming/MapRepresentations.html),
[[_grounding]] §7). This makes the quadtree of [[Hierarchical-Spatial-Trees]] double-duty:
storage *and* search graph.

## Scale context

These techniques are motivated by large maps (the cited benchmarks are commercial-game maps
far larger than 64x64 = 4,096 cells). Amit Patel's guidance is to "start with your existing
game world representation, then optimize if needed"; hierarchical representations trade some
path optimality for speed ([[_grounding]] §7). Their relevance at small scale is less about
raw speed than about (a) planning over *known-space only* maps whose content differs per
agent, and (b) carrying the same abstraction into multi-layer worlds, where HPA\*-style
cluster graphs and floor-stair topologies coincide ([[Extending-to-3D-and-Layered-Grids]]).

## Grounding

- [[_grounding]] — §7 (HPA\*, quadtree pathfinding, map-representation guidance)
- [Botea & Müller — Near Optimal Hierarchical Path-Finding](https://www.researchgate.net/publication/228785110_Near_optimal_hierarchical_path-finding_HPA)
- [Amit Patel — Map Representations](http://theory.stanford.edu/~amitp/GameProgramming/MapRepresentations.html)
