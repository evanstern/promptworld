---
title: Agent-Mental-Maps
aliases: [Mental Maps, Spatial Memory]
tags: [moc, spatial-memory, grid-worlds]
type: moc
created: 2026-07-24
updated: 2026-07-24
related: []
---

# Agent-Mental-Maps

How agents build and maintain private "mental maps" of 2D grid worlds, with attention to
structures and algorithms that extend to 3D (layered grids / voxel volumes). Gathered to
ground the replacement of promptworld's omniscient nearest-target resolution with per-agent
spatial knowledge on its 64x64 map.

## Scope

**In:** hierarchical spatial data structures (quadtrees, octrees, k-d trees, BVHs);
occupancy-grid / belief-map cell semantics from robotics; cognitive-map, topological, and
landmark-graph models; fog-of-war and knowledge-limited perception designs; belief decay,
staleness, and provenance; hierarchical pathfinding over known space (HPA\*, quadtree A\*);
frontier-based exploration; multi-agent map sharing/merging; memory footprint of per-agent
maps. **Out:** any recommendation of which design promptworld should adopt (analysis phase);
Go library surveys; rendering/visualization of maps; full SLAM (localization is a non-problem
here — agents know their own coordinates).

## What is known

Three representational families recur across game AI and robotics, and they compose rather
than compete: **metric belief grids** (per-cell occupied/free/unknown with log-odds updates —
[[Occupancy-Grids-and-Belief-Maps]]), **hierarchical trees** over those grids (quadtrees in 2D,
octrees in 3D, with the probabilistic payload unchanged across dimensions —
[[Hierarchical-Spatial-Trees]]), and **sparse topological/landmark graphs** modeled on human
spatial memory ([[Topological-and-Landmark-Maps]]). Games gate knowledge behind perception
with three-state per-player visibility grids ([[Fog-of-War-and-Limited-Perception]]); the
explicit *unknown* state makes directed exploration possible ([[Frontier-Based-Exploration]]).
Remembered content goes stale in dynamic worlds, and the literature handles this with decay
toward uncertainty, staleness-weighted retrieval, and provenance-conditioned updates
([[Belief-Decay-and-Stale-Knowledge]]) — the same machinery that governs merging another
agent's reports ([[Multi-Agent-Map-Sharing]]). The hierarchy that stores a map can also plan
over it (HPA\* clusters, quadtree-gate A\* — [[Hierarchical-Pathfinding-Over-Known-Space]]),
and the consensus 3D lift for building-like worlds is per-layer metric maps joined by a
portal/stair topology, with full voxel octrees (OctoMap) as the upper bound
([[Extending-to-3D-and-Layered-Grids]]). At 64x64 scale, every candidate representation is
memory-trivial even multiplied across agents and layers
([[Memory-Footprint-for-Many-Private-Maps]]).

## Notes

- [[Brief-and-Assumptions]] — the request, promptworld context, assumptions, and flagged ambiguities
- [[Hierarchical-Spatial-Trees]] — quadtrees, octrees, k-d trees, BVHs; the 2D→3D relationship
- [[Occupancy-Grids-and-Belief-Maps]] — log-odds cell beliefs; occupied/free/unknown; OctoMap continuity
- [[Topological-and-Landmark-Maps]] — cognitive maps, place graphs, hybrid metric+topological+semantic
- [[Fog-of-War-and-Limited-Perception]] — three-state visibility grids; partial observability in games
- [[Frontier-Based-Exploration]] — directed exploration of unknown space via frontier cells
- [[Belief-Decay-and-Stale-Knowledge]] — decay toward uncertainty, staleness penalties, provenance
- [[Hierarchical-Pathfinding-Over-Known-Space]] — HPA\*, quadtree A\*, gates; storage doubling as search graph
- [[Extending-to-3D-and-Layered-Grids]] — 2.5D → layered floors → voxel octrees; floor-stair topology
- [[Multi-Agent-Map-Sharing]] — gossip, CRDTs, map merging, reliability-conditioned trust
- [[Memory-Footprint-for-Many-Private-Maps]] — concrete sizes at 64x64; bitsets, compression, structure sharing

## Analyses

_Opinionated evaluations built on this branch (added by the analyze phase). Empty until then._

## Open questions

- Should mental maps gate only *target resolution* (knowing where things are) or also
  *pathfinding* (routing only through remembered-passable space)?
- Do promptworld agents share map knowledge when they talk (spatial gossip), and at what trust
  level relative to firsthand witnessing?
- How should an LLM planner consume a mental map — rendered text, read-only query tools, or
  both? (The gathered literature is mostly about non-LLM planners.)
- For the future layered grids: fully independent per-layer grids joined at portals, or a true
  voxel volume? The literature's floor-stair topology assumes the former.
- Quantitative decay rates: sources establish *that* beliefs should decay toward uncertainty,
  but appropriate half-lives for a game-speed world are not established.

## Grounding

- [[_grounding]] — the research pass this branch is built on (web-search fan-out, 2026-07-24, 56 sources)
- [Hornung et al. — OctoMap (Autonomous Robots 2013)](https://link.springer.com/article/10.1007/s10514-012-9321-0)
- [Botea & Müller — Near Optimal Hierarchical Path-Finding (HPA\*)](https://www.researchgate.net/publication/228785110_Near_optimal_hierarchical_path-finding_HPA)
- [Amit Patel — Map Representations](http://theory.stanford.edu/~amitp/GameProgramming/MapRepresentations.html)
- [Samet — The Quadtree and Related Hierarchical Data Structures (1984)](http://www.cs.umd.edu/~hjs/pubs/SameCSUR84-ocr.pdf)
