---
title: Storyteller-Driven Pacing
aliases: [AI Director Pattern]
tags: [pacing, storyteller, drama]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Game-Gameplay-Patterns]]", "[[Simulation-vs-Director]]", "[[Difficulty-Dials-and-Dynamic-Depth]]"]
---

# Storyteller-Driven Pacing

RimWorld's central gameplay device: a director AI that watches the game state and injects
events to shape a dramatic arc. Lineage runs from Left 4 Dead's AI Director; RimWorld
markets itself as a "story generator," not a strategy game ([[_grounding]] § RimWorld).

## Mechanism

- **Story watchers + incident generators**: watchers monitor game state; generators trigger
  events in response. The storyteller "builds pressure, then releases it, creating emotional
  arcs similar to a narrative campaign" ([[_grounding]]).
- **Cassandra Classic** encodes the arc explicitly: challenges start small, scale with time
  and colony wealth, run on cooldowns (1–2 major threats per ~7–10 day quadrum, ≥2 days
  between major events), with recovery windows built in ([[_grounding]]).

## Persona as the pacing dial

The three storytellers are **named personalities that double as pacing/difficulty
settings** ([[_grounding]]):

| Storyteller | Pacing contract |
|---|---|
| Cassandra Classic | rising tension curve with breathing room |
| Phoebe Chillax | long recovery gaps; the de facto newcomer mode |
| Randy Random | no arc guarantee — "it's all drama to him" |

Selecting a narrator persona *is* selecting the difficulty and rhythm: "the AI storyteller
is the main mechanism used to determine game difficulty and play style" ([[_grounding]]).
The framing puts the choice in fiction terms (who tells your story) rather than in
easy/normal/hard terms — compare [[Difficulty-Dials-and-Dynamic-Depth]] on why framing
matters.

## Grounding

- [[_grounding]] — § "RimWorld — the AI storyteller"
- [AI Storytellers — RimWorld Wiki](https://rimworldwiki.com/wiki/AI_Storytellers)
- [Game Developer: procedurally generated storytelling](https://www.gamedeveloper.com/design/rimworld-dwarf-fortress-and-procedurally-generated-story-telling)
