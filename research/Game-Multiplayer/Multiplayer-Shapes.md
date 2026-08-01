---
title: Multiplayer Shapes
aliases: [Shared World vs Parallel Settlements, MP Gameplay Shapes]
tags: [gameplay, shapes, shared-world, parallel-colonies, asynchronous]
type: note
created: 2026-08-01
updated: 2026-08-01
related: ["[[Game-Multiplayer]]", "[[Sync-Architectures]]", "[[Inter-Settlement-Interaction]]", "[[Deity-Multiplayer]]", "[[World-Size-and-Density]]", "[[Promptworld-Baseline]]"]
---

# Multiplayer Shapes

Before any netcode question, a colony/agent sim has to answer a gameplay question: *what do two
players share?* The survey shows four distinct answers, each with a shipped example, and — in
RimWorld's case — the same base game solved both dominant ways by two separate mods.

## Shape A — One settlement, many hands (shared control)

All players operate the same colony, same clock, same pawns.

- **Zetrith's RimWorld Multiplayer**: one colony, deterministic lockstep, "every pawn action,
  every random event must occur identically for everyone" ([[_grounding]] § RimWorld).
- **Space Station 13**: one station, dozens of players, each occupying a *role* rather than a
  territory ([[_grounding]] § SS13).

Consequences recorded in the sources:
- The shared clock is mandatory — you cannot run the same colony at two speeds.
- Fidelity becomes existential; the whole desync apparatus exists to serve this shape
  ([[Determinism-and-Desync]]).
- Player count is bounded by the sync scheme, and by how many hands one colony can absorb before
  players collide over the same objects.

## Shape B — Separate settlements, shared world (parallel play)

Each player runs their own settlement; the world and a verb set are shared.

- **RimWorld Together** is the direct example and states the design intent explicitly: play "in
  the same planet, at the same time, **while keeping separate colonies and pace**." Shared verbs:
  "visiting, raiding, trading, creating factions, roads, sites." Feature list includes "visit,
  raid, and spy other players," real-time trade and chat, and "create and manage custom factions."
  Colonies keep "independent progression for each player" ([[_grounding]] § RimWorld).
- **Haven & Hearth** is the long-running MMO version: villages as claims with radius, membership,
  and upkeep, plus higher-tier Realms ([[Inter-Settlement-Interaction]]).

Consequences:
- **No shared clock requirement** — "separate colonies and pace" is called out as a feature. This
  removes the slowest-peer problem entirely ([[Sync-Architectures]]).
- Determinism requirements drop sharply: settlements only need to agree at **interaction
  boundaries**, not tick-by-tick. RimWorld Together's stated ability to work "regardless of mods
  or DLCs" — impossible under lockstep — is direct evidence of how much weaker the coupling is.
- The shared surface is an **abstract world map**, not shared tactical space
  ([[World-Size-and-Density]]).

## Shape C — Shared world, opposed sides (competitive)

- The **god-game lineage**: "players may sometimes compete against other players with their own
  population of supporters. In Populous specifically, the user is a deity that must lead their
  followers against the followers of a rival deity in order to conquer and capture them"
  ([[Deity-Multiplayer]]).
- RimWorld Together includes the competitive verbs (raid, spy) alongside cooperative ones,
  showing the shapes are not mutually exclusive.

## Shape D — Shared world, shared stakes, politics as the mechanic

- **Eco**: ~100 players on one simulated planet, where the multiplayer content is governance and
  economy — a collectively decided "constitution that determines how laws are proposed and
  approved," player-created currencies, and laws that "restrict what other players can do" or
  incentivise via "taxes or government grants." A shared deadline (a meteor in thirty days) and a
  shared externality (environmental damage from play itself) supply the pressure
  ([[_grounding]] § Eco).

This shape is notable because the *conflict* is over collective decisions rather than territory,
and the simulation itself — the ecology — is the shared object players negotiate about.

## The orthogonal axis: synchronous or asynchronous participation

Independent of A–D, participation can be concurrent or not. The literature treats this as a
spectrum, not a binary: "games today exist on a spectrum from fully concurrent to fully
asynchronous and everything in between," where persistence "enables asynchronous interactions"
and the world "is always running, even when players aren't actively playing" ([[_grounding]] §
Asynchronous / persistent-world multiplayer).

Combining the axes yields the practical matrix:

| | Concurrent | Asynchronous |
|---|---|---|
| **Shared settlement** | RimWorld MP, SS13 | (rare — needs a shared clock) |
| **Separate settlements** | RimWorld Together, Eco | Screeps; play-by-visit persistent worlds |

The bottom-left cell is where an always-on daemon most naturally sits: the world runs
continuously, and operators attach and detach without stopping it.

## What each shape demands

| Requirement | A: shared settlement | B: parallel settlements | C: opposed | D: shared politics |
|---|---|---|---|---|
| Shared clock | required | not required | required if simultaneous | required |
| Determinism across instances | existential | boundary-only | boundary-only | n/a (one server) |
| Identity/attribution | needed for UX | needed for ownership | needed for scoring | needed for law/votes |
| Inter-settlement verbs | n/a | core content | core content | secondary |
| Conflict source | coordination friction | trade/raid/visit | direct opposition | collective decisions |
| Map growth pressure | none | world-map layer | world-map layer | one bigger shared world |

## Applicability to promptworld

- TASK-65 already frames the open decision in exactly the A-vs-B terms this survey found:
  "parallel villages vs shared village with per-player angels," and records that "the shape
  decision gates everything else" ([[Promptworld-Baseline]]).
- The player unit in promptworld is the **Guardian**, not a pawn — villagers are sealed and
  influence is indirect. That places the game in the god-game lineage for control style, whichever
  shape is chosen ([[Deity-Multiplayer]]).
- The always-on daemon posture matches the asynchronous column: the world advances 24/7 and
  clients "attach and detach without affecting it" ([[Promptworld-Baseline]]).
- Shape A is already partially realised in a single-player sense — the IPC server explicitly
  supports **multiple concurrent sessions** on one world, each with its own gapless event
  subscription and its own console turns. What is missing is not the transport but identity,
  attribution, and any gameplay rule about two operators acting on one village.

## Grounding

- [[_grounding]] § RimWorld; § SS13; § Eco; § God games; § Asynchronous / persistent-world
  multiplayer; § Haven & Hearth; § promptworld's own architecture
- [Steam Workshop — RimWorld Together](https://steamcommunity.com/sharedfiles/filedetails/?id=3005289691)
- [Eco Wiki — Collaboration](https://wiki.play.eco/en/Collaboration)
