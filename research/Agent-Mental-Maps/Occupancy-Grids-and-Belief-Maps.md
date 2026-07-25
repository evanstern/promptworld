---
title: Occupancy Grids and Belief Maps
aliases: [Log-Odds Mapping]
tags: [robotics, occupancy-grid, belief]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Mental-Maps]]", "[[Hierarchical-Spatial-Trees]]", "[[Belief-Decay-and-Stale-Knowledge]]", "[[Frontier-Based-Exploration]]"]
---

# Occupancy Grids and Belief Maps

The robotics-standard model for what a mental map's *cells* actually store: a per-cell belief,
not a per-cell fact.

## The model

Occupancy grid mapping discretizes space into cells, each holding a probability of being
occupied, refined over time by a binary Bayes filter as observations arrive
([Freiburg slides](http://ais.informatik.uni-freiburg.de/teaching/ss16/robotics/slides/12-occupancy-mapping.pdf),
[[_grounding]] §2). Three effective states fall out of the probability:

- **occupied** (probability > 0.5 / positive log-odds)
- **free** (probability < 0.5 / negative log-odds)
- **unknown** (exactly the 0.5 prior — never observed)

Representing *unknown* explicitly is the load-bearing feature for knowledge-limited agents: it
is what distinguishes "I know this tile is empty" from "I have never looked"
([[_grounding]] §2, §5; see [[Frontier-Based-Exploration]] for how unknown cells drive
exploration).

## Log-odds updates

Cells store **log-odds** rather than probabilities. Each observation adds a constant — positive
for occupied evidence, negative for free evidence — so the whole Bayesian update "factorizes
… into a series of simple recursive additions"; updates are cheap enough that "a robot can
update millions of cells per second," and the representation avoids numerical underflow from
repeated probability multiplication
([CMU 16-831 notes](https://www.cs.cmu.edu/~16831-f12/notes/F12/16831_lecture05_vh.pdf),
[ThinkAutonomous](https://www.thinkautonomous.ai/blog/occupancy-grid-mapping/),
[[_grounding]] §2). A small integer (e.g. int8) per cell suffices in practice.

## Variants relevant to dynamic worlds

- **Confidence-aware occupancy grids** additionally track how much evidence supports each
  cell's estimate ([Agha et al.](https://karolhausman.github.io/pdf/agha17-ws-iros.pdf),
  [[_grounding]] §2).
- **Time-aware occupancy grids** account for the age of observations in dynamic environments
  ([USPTO 12,523,498](https://image-ppubs.uspto.gov/dirsearch-public/print/downloadPdf/12523498),
  [[_grounding]] §2) — the spatial version of the staleness problem covered in
  [[Belief-Decay-and-Stale-Knowledge]].

## Relationship to tree structures

The occupancy payload composes with hierarchical containers rather than competing with them:
probabilistic quadtrees hold occupancy beliefs in a quadtree ([[_grounding]] §1), and
**OctoMap** stores per-voxel log-odds occupancy in an octree, explicitly representing occupied,
free, and unknown 3D space; it adds a compression step merging identical children and is the
de-facto standard probabilistic 3D map in robotics
([Hornung et al. 2013](https://link.springer.com/article/10.1007/s10514-012-9321-0),
[[_grounding]] §3). The same cell semantics thus run unchanged from flat 2D grids through
quadtrees to 3D octrees — see [[Hierarchical-Spatial-Trees]] and
[[Extending-to-3D-and-Layered-Grids]].

## Grounding

- [[_grounding]] — §2 (occupancy grids, log-odds), §3 (OctoMap), §1 (probabilistic quadtrees)
- [CMU 16-831 — Occupancy Mapping lecture notes](https://www.cs.cmu.edu/~16831-f12/notes/F12/16831_lecture05_vh.pdf)
- [Hornung et al. — OctoMap (Autonomous Robots 2013)](https://link.springer.com/article/10.1007/s10514-012-9321-0)
