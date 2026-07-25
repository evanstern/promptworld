---
title: Multi-Agent Map Sharing
aliases: [Map Merging, Spatial Gossip]
tags: [multi-agent, gossip, map-merging]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Mental-Maps]]", "[[Belief-Decay-and-Stale-Knowledge]]", "[[Fog-of-War-and-Limited-Perception]]"]
---

# Multi-Agent Map Sharing

How private maps become partially shared knowledge when agents communicate.

## Gossip protocols

Gossip protocols are "decentralized communication mechanisms where each node periodically
exchanges state information with random neighbors, gradually propagating knowledge through the
network," giving "scalable, fault-tolerant communication and emergent knowledge convergence
without central control" ([Revisiting Gossip Protocols](https://arxiv.org/html/2508.01531v1),
[[_grounding]] §10). Through "probabilistic propagation, partial-view sampling, and
anti-entropy reconciliation," gossip yields "gradual convergence of local world models"
([[_grounding]] §10) — i.e., agents' mental maps drift toward agreement at the rate they meet.

## Merge mechanics

- **CRDT-layered gossip** provides eventual consistency of shared state without central
  coordination ([Gossip-Enhanced Communication Substrate](https://arxiv.org/pdf/2512.03285),
  [[_grounding]] §10).
- **Multi-robot SLAM merging**: decentralized systems detect map overlap, merge, then keep
  "sharing keyframes and map points to expand the shared map"
  ([DVM-SLAM](https://arxiv.org/pdf/2503.04126), [DRACo-SLAM](https://arxiv.org/pdf/2210.00867),
  [[_grounding]] §10).
- For grid/belief maps specifically, merge is a per-cell belief update — the same evidence
  merge used for the agent's own observations (noisy-OR in BeliefMem;
  log-odds addition in occupancy grids), applied to another agent's report
  ([[_grounding]] §6, §10; [[Belief-Decay-and-Stale-Knowledge]]).

## Trust and provenance in merges

Secondhand map content is weaker evidence than firsthand perception:
**reliability-conditional updating** conditions the belief update on the reporting source's
reliability, defending against wrong or poisoned reports
([When Does Belief-Based Agent Memory Help?](https://arxiv.org/html/2606.22030),
[[_grounding]] §10, §6). This is the mechanism-level counterpart of witnessed/told/inferred
provenance distinctions.

## Scope note

Classic team fog-of-war (allied units pooling vision into one shared map,
[[Fog-of-War-and-Limited-Perception]]) is the degenerate case: instantaneous, full-trust
merge within a team. Gossip/CRDT approaches cover the general case of intermittent, pairwise,
partial-trust exchange ([[_grounding]] §10).

## Grounding

- [[_grounding]] — §10 (gossip, CRDTs, map merging), §6 (reliability conditioning)
- [Revisiting Gossip Protocols (arXiv)](https://arxiv.org/html/2508.01531v1)
- [DVM-SLAM (arXiv)](https://arxiv.org/pdf/2503.04126)
