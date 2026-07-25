---
title: Memory Footprint for Many Private Maps
aliases: [Per-Agent Map Cost]
tags: [memory, bitset, compression]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Mental-Maps]]", "[[Hierarchical-Spatial-Trees]]", "[[Occupancy-Grids-and-Belief-Maps]]"]
---

# Memory Footprint for Many Private Maps

What per-agent private maps cost in memory, and the known techniques for shrinking them.

## Baseline arithmetic at 64x64

A 64x64 grid is 4,096 cells. From the representation definitions in the grounding
([[_grounding]] §11):

| Representation | Per cell | Per layer (64x64) | 8 agents | 8 agents × 5 layers |
|---|---|---|---|---|
| Bitset (1 bit: seen/unseen) | 1 bit | 512 B | 4 KiB | 20 KiB |
| Byte grid (state enum or int8 log-odds) | 1 byte | 4 KiB | 32 KiB | 160 KiB |
| Two byte layers (state + timestamp/age) | 2 bytes | 8 KiB | 64 KiB | 320 KiB |

Even the heaviest flat layout is far below any practical constraint at this scale — the
motivating benchmarks for compressed representations are orders of magnitude larger maps
([[_grounding]] §7, §11).

## Techniques in the literature

- **Bitsets** store one boolean per bit and combine with fast logical operations
  ([Cleverence — BitSet](https://www.cleverence.com/articles/oracle-documentation/bitset-java-platform-se-8-4827/),
  [[_grounding]] §11) — suited to seen/unseen visibility layers, including team-vision unions
  via OR.
- **Flat per-agent visibility arrays** are the working baseline in multi-agent contest systems
  ("initialized with values denoting unobserved cells," [MAPC 2019](https://arxiv.org/pdf/2006.02816),
  [[_grounding]] §5, §11).
- **Quadtree compression** collapses spatially coherent regions — effective precisely when
  maps are mostly-unknown or mostly-open ([[_grounding]] §1, §11;
  [[Hierarchical-Spatial-Trees]]). Probabilistic quadtrees do this for belief maps.
- **Octree compression** (OctoMap) merges identical children to keep 3D models compact
  ([Hornung et al.](https://link.springer.com/article/10.1007/s10514-012-9321-0),
  [[_grounding]] §3, §11).
- **Parametric/continuous maps** (GMMap — Gaussian mixtures) and **non-uniform cell**
  representations cut memory when grid resolution exceeds task needs
  ([GMMap](https://arxiv.org/pdf/2306.03740), [[_grounding]] §11).

## Structure sharing

A distinction visible across the sources: static world structure (terrain, walls) is common to
all agents, while *knowledge of it* differs per agent. Fog-of-war implementations exploit this
by keeping one shared world map plus a thin per-player visibility/exploration layer
([[_grounding]] §5; [[Occupancy-Grids-and-Belief-Maps]] for what the thin layer stores) —
the per-agent cost is then only the knowledge layer, not a world copy.

## Grounding

- [[_grounding]] — §11 (compact representations), §5 (per-agent visibility grids), §3 (OctoMap compression)
- [Cleverence — Java BitSet Explained](https://www.cleverence.com/articles/oracle-documentation/bitset-java-platform-se-8-4827/)
- [GMMap (arXiv)](https://arxiv.org/pdf/2306.03740)
