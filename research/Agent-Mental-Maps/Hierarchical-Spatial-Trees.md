---
title: Hierarchical Spatial Trees
aliases: [Quadtrees and Octrees]
tags: [data-structures, quadtree, octree]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Mental-Maps]]", "[[Occupancy-Grids-and-Belief-Maps]]", "[[Hierarchical-Pathfinding-Over-Known-Space]]", "[[Extending-to-3D-and-Layered-Grids]]"]
---

# Hierarchical Spatial Trees

The tree structures that organize grid-space hierarchically, and how the 2D and 3D variants
relate.

## Quadtrees (2D)

A quadtree recursively divides a 2D region into four quadrants; subdivision stops where a
region is homogeneous, so uniform areas collapse into single leaves
([Samet 1984](http://www.cs.umd.edu/~hjs/pubs/SameCSUR84-ocr.pdf), [[_grounding]] §1).
Two properties matter for mental maps:

- **Compression of coherent space.** Region quadtrees give "compact representations of large
  two-dimensional binary arrays" — a mostly-unexplored or mostly-open map costs few nodes
  ([[_grounding]] §1, §11).
- **Probabilistic variant.** Probabilistic quadtrees hold occupancy *beliefs* rather than
  binary values, "hierarchically subdividing non-uniform regions … while merging homogeneous
  areas" ([arXiv 2403.06494](https://arxiv.org/pdf/2403.06494), [[_grounding]] §1) — the tree
  and the belief map (see [[Occupancy-Grids-and-Belief-Maps]]) are compatible, not competing.

For a 64x64 world the full tree is at most 7 levels deep (64 → 32 → 16 → 8 → 4 → 2 → 1), so
depth is bounded and small.

## Octrees (3D)

An octree is the direct 3D analog: each internal node has exactly eight children, bounding
cubes instead of bounding squares ([GameDev.net](https://www.gamedev.net/articles/programming/general-and-gameplay-programming/introduction-to-octrees-r3529/),
[[_grounding]] §1). The quadtree→octree step changes the branching factor and nothing else
structurally, which is the main reason quadtree-based designs are described as 3D-extensible.
OctoMap (see [[Occupancy-Grids-and-Belief-Maps]] and [[Extending-to-3D-and-Layered-Grids]]) is
the standard demonstration that the probabilistic-occupancy payload survives the lift to 3D
unchanged ([[_grounding]] §3).

## The quadtree-plus-layers middle path

Game-development practice distinguishes when each fits: quadtrees for "sprawling topology that
is roughly-2D in nature," octrees for scenes extending in all directions ([[_grounding]] §1).
A documented intermediate for mostly-flat worlds with limited verticality keeps a quadtree and
adds "additional information or layers … to the leaf nodes to specify 3D aspects," acceptable
"when the additional dimension (e.g., height) is not as detailed or complex as the other
dimensions" ([GameDev.net forum](https://gamedev.net/forums/topic/463418-using-quadtrees-in-maps/),
[[_grounding]] §1). This sits between pure 2D quadtrees and full octrees and corresponds to
"layered grids" — see [[Extending-to-3D-and-Layered-Grids]].

## Other tree structures

- **k-d trees** split space one dimension at a time with arbitrary split positions; like
  quadtrees they are hierarchical space partitions, used mainly for nearest-neighbor queries
  over point sets rather than region occupancy ([Samet 1984](http://www.cs.umd.edu/~hjs/pubs/SameCSUR84-ocr.pdf)).
- **BVHs** (bounding volume hierarchies) group *objects* rather than partitioning *space*;
  they appear in the same spatial-partitioning literature (collision, rendering) but do not
  represent unknown/known regions, which is the defining need of a mental map
  ([[_grounding]] §1).

## Grounding

- [[_grounding]] — §1 (hierarchical trees), §3 (OctoMap), §11 (compression)
- [Samet — The Quadtree and Related Hierarchical Data Structures (1984)](http://www.cs.umd.edu/~hjs/pubs/SameCSUR84-ocr.pdf)
- [GameDev.net — Introduction to Octrees](https://www.gamedev.net/articles/programming/general-and-gameplay-programming/introduction-to-octrees-r3529/)
