---
title: Frontier-Based Exploration
aliases: [Frontier Exploration]
tags: [exploration, frontiers, robotics]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Mental-Maps]]", "[[Occupancy-Grids-and-Belief-Maps]]", "[[Fog-of-War-and-Limited-Perception]]"]
---

# Frontier-Based Exploration

The standard algorithm for turning "I don't know where X is" into directed search instead of
random wandering.

## Frontiers

A **frontier** is the boundary between explored and unexplored space — formally, a known free
cell adjacent to at least one unknown cell
([Frontiers in Robotics & AI](https://www.frontiersin.org/journals/robotics-and-ai/articles/10.3389/frobt.2021.616470/full),
[[_grounding]] §8). The definition presupposes a map with an explicit *unknown* state, which
is exactly what occupancy grids and fog-of-war visibility grids provide
([[Occupancy-Grids-and-Belief-Maps]], [[Fog-of-War-and-Limited-Perception]]).

## The algorithm

Frontier-based exploration (Yamauchi's classic method) loops
([Topiwala](https://arxiv.org/pdf/1806.03581), [Gu & Xu](https://cs.unh.edu/~tg1034/project/TianyiGu_AutonomousMapping.pdf),
[[_grounding]] §8):

1. Scan the grid for free cells adjacent to unknown cells (frontier detection — often BFS).
2. Cluster contiguous frontier cells into frontier regions/contours.
3. Pick a frontier (nearest, largest, or most informative) and navigate to it.
4. Perceive, update the map, repeat — "until there are no more frontiers and therefore no more
   unknown regions."

## Variants

- **Topological frontier-based exploration** attaches semantic/topological structure to
  frontier selection ([PMC](https://www.ncbi.nlm.nih.gov/pmc/articles/PMC6832505/),
  [[_grounding]] §8).
- **Uncertainty-aware information prediction** ranks frontiers by expected information gain
  rather than distance ([arXiv 2412.12825](https://arxiv.org/pdf/2412.12825), [[_grounding]] §8).
- Efficient frontier detection itself is a studied subproblem (avoiding full-grid rescans each
  cycle) ([Frontiers in Robotics & AI](https://www.frontiersin.org/journals/robotics-and-ai/articles/10.3389/frobt.2021.616470/full),
  [[_grounding]] §8) — though at 64x64 (4,096 cells) a full scan is trivially cheap by the
  update-throughput numbers in [[Occupancy-Grids-and-Belief-Maps]].

## Grounding

- [[_grounding]] — §8 (frontier-based exploration)
- [Topiwala — Frontier Based Exploration for Autonomous Robot (arXiv)](https://arxiv.org/pdf/1806.03581)
- [Frontiers in Robotics & AI — Detecting Frontier Cells](https://www.frontiersin.org/journals/robotics-and-ai/articles/10.3389/frobt.2021.616470/full)
