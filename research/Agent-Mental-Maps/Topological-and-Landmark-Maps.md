---
title: Topological and Landmark Maps
aliases: [Cognitive Maps]
tags: [cognitive-map, topological, landmarks]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Mental-Maps]]", "[[Occupancy-Grids-and-Belief-Maps]]", "[[Extending-to-3D-and-Layered-Grids]]"]
---

# Topological and Landmark Maps

The alternative to metric grids: a sparse graph of *places* and *connections*, modeled on how
humans actually remember space.

## The cognitive-science basis

"Humans build approximate graphs of their environment, encoding relative distances between
landmarks" — and "save landmark features in their memory for navigation, instead of detailed
scene layouts" ([Memory Proxy Maps](https://arxiv.org/html/2411.09893),
[MemoNav](https://arxiv.org/pdf/2208.09610), [[_grounding]] §4). In robotics and embodied-AI
these translate directly into **topological graphs**: nodes are landmarks/places, edges are
known traversals.

## What the graph holds

For embodied navigation, "memory should organize online perception into a relational spatial
state that summarizes explored regions, traversed paths, landmarks, and their spatial
connectivity" ([SpaceVLN](https://arxiv.org/pdf/2606.08992), [[_grounding]] §4). Concrete
systems layer it:

- SpaceVLN keeps a **spatial waypoint graph** (global structure) plus a **local landmark
  memory** (what is near each waypoint), built online and shared between planner and executor
  ([[_grounding]] §4).
- Topological-memory navigation systems pair a sparse graph of previously visited places with
  a local controller that handles point-to-point motion
  ([IEEE/CAA JAS](https://www.ieee-jas.net/article/doi/10.1109/JAS.2024.124332), [[_grounding]] §4).
- Semantic-spatial systems (Meta-Memory) attach semantic content ("what is here") to spatial
  nodes for retrieval ([Meta-Memory](https://arxiv.org/html/2509.20754v1), [[_grounding]] §4).

## Hybrid architectures

The survey-level position is that no single layer suffices: research direction is "hybrid
cognitive map architectures that unify **topological, metric, and semantic** representations
into a flexible and context-aware spatial memory system"
([Frontiers in Comp. Neuroscience](https://www.frontiersin.org/journals/computational-neuroscience/articles/10.3389/fncom.2024.1498160/full),
[Mind Meets Space](https://arxiv.org/pdf/2509.09154), [[_grounding]] §4). In grid-world terms:
a metric layer (the grid cells one has seen — [[Occupancy-Grids-and-Belief-Maps]]), a
topological layer (named places and how they connect), and a semantic layer (what each place
is for). Multi-floor navigation work uses exactly this split, with per-floor metric maps
joined by a cross-floor topological graph ([[_grounding]] §9,
[[Extending-to-3D-and-Layered-Grids]]).

## Properties relative to metric grids

- Topological maps are **sparse**: memory scales with places visited, not with world area
  ([[_grounding]] §4).
- They are naturally **verbalizable** (nodes have identities like "the well", "the north
  woods"), which is why landmark-based memory dominates in language-driven navigation agents
  ([[_grounding]] §4).
- They do not by themselves record *unknown* space — that is the metric/occupancy layer's job
  ([[Occupancy-Grids-and-Belief-Maps]]).

## Grounding

- [[_grounding]] — §4 (cognitive/topological/landmark maps), §9 (floor-stair topology)
- [SpaceVLN (arXiv)](https://arxiv.org/pdf/2606.08992)
- [Frontiers — Learning dynamic cognitive map](https://www.frontiersin.org/journals/computational-neuroscience/articles/10.3389/fncom.2024.1498160/full)
