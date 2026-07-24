---
title: Fog of War and Limited Perception
aliases: [Knowledge-Limited Perception]
tags: [fog-of-war, perception, games]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Mental-Maps]]", "[[Occupancy-Grids-and-Belief-Maps]]", "[[Belief-Decay-and-Stale-Knowledge]]", "[[Frontier-Based-Exploration]]"]
---

# Fog of War and Limited Perception

How games and multi-agent simulations gate knowledge behind perception — the design pattern
for replacing omniscient world access.

## The mechanic

Fog of war "simulate[s] limited knowledge or visibility of the game world," and its design
purpose is exactly the behavioral change sought from knowledge-limited agents: it "requires
players to gather information, scout, and make decisions based on partial knowledge, creating
uncertainty and encouraging exploration"
([Machinations.io](https://machinations.io/glossary/fog-of-war), [[_grounding]] §5).

## The standard tile implementation: a three-state visibility grid

The common tile-based implementation keeps, per player/agent, a visibility grid with three
states ([Didac Romero tutorial](https://didacromero.github.io/Fog-of-War/), [[_grounding]] §5):

1. **Unexplored** — never seen; nothing known.
2. **Explored but not currently visible** — static content (terrain, structures) is remembered
   as last seen; dynamic entities are hidden, so remembered content can be **stale**.
3. **Currently visible** — inside some unit's sight radius this tick; ground truth.

The grid is recomputed from sight radii each tick; state 2 is precisely a *mental map* — a
private, possibly-outdated copy of the world. This three-state scheme is structurally
identical to the occupied/free/unknown semantics of occupancy grids
([[Occupancy-Grids-and-Belief-Maps]]), with "explored-but-stale" adding the time dimension
covered in [[Belief-Decay-and-Stale-Knowledge]].

## Consequences for agent design

In RTS AI research, fog of war makes the environment **partially observable**: "terrain is
partly unknown and must be explored; bots must be able to update their knowledge about terrain
and have units explore unknown terrain to locate hidden enemy bases"
([Hagelbäck & Johansson](https://www.researchgate.net/publication/224491323_Dealing_with_Fog_of_War_in_a_Real_Time_Strategy_Game_Environment),
[[_grounding]] §5). Agents facing it maintain explicit internal world models and reason under
uncertainty — from Bayesian networks over hidden state to learned predictors of unseen enemy
units ([[_grounding]] §5). Multi-agent programming contest teams use per-agent "visibility
grids initialized with values denoting unobserved cells … recording what parts of the
environment have been explored and what information has been observed"
([MAPC 2019](https://arxiv.org/pdf/2006.02816), [[_grounding]] §5).

## Design dimensions observed in the literature

- **What perception grants:** a sight *radius* (RTS convention) vs. line-of-sight visibility;
  single-pass grid line-of-sight algorithms exist for the latter
  ([arXiv 2403.06494](https://arxiv.org/pdf/2403.06494), [[_grounding]] §11).
- **What memory retains:** static layers (terrain) are typically remembered; dynamic layers
  (units, resources that move or deplete) are remembered *as last seen* or not at all
  ([[_grounding]] §5).
- **Scope of sharing:** classic RTS fog is per-*team* (allied units pool vision); per-agent
  private maps are the stricter variant used in multi-agent contests ([[_grounding]] §5,
  and see [[Multi-Agent-Map-Sharing]]).

## Grounding

- [[_grounding]] — §5 (fog of war, partial observability), §11 (visibility computation)
- [Didac Romero — tile-based Fog of War guide](https://didacromero.github.io/Fog-of-War/)
- [Dealing with Fog of War in a RTS Game Environment](https://www.researchgate.net/publication/224491323_Dealing_with_Fog_of_War_in_a_Real_Time_Strategy_Game_Environment)
