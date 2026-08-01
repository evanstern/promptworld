---
title: Sync Architectures
aliases: [Netcode Families, Lockstep vs State Sync]
tags: [netcode, architecture, lockstep, authority]
type: note
created: 2026-08-01
updated: 2026-08-01
related: ["[[Game-Multiplayer]]", "[[Determinism-and-Desync]]", "[[Latency-and-Responsiveness]]", "[[Scale-Ceilings-and-Interest-Management]]", "[[Promptworld-Baseline]]"]
---

# Sync Architectures

The surveyed games resolve "how do two machines agree on one world?" into three families. The
choice is not stylistic: each family fixes what must be deterministic, what travels on the wire,
what the failure mode looks like, and roughly how many participants fit.

## Family 1 — Deterministic lockstep (send inputs, replay the simulation everywhere)

Every participant runs the full simulation. Only **user inputs** cross the network; each machine
applies the same inputs in the same order and is expected to arrive at bit-identical state.

- **Factorio** is the canonical shipped example: it synchronizes "by sending only the user inputs
  that control that game, rather than networking the state of the objects in the game itself,"
  with "the full state of the map, player, entities, everything … simulated deterministically on
  all clients" ([[_grounding]] § Factorio).
- **RimWorld's Zetrith Multiplayer** applies the same model to a colony sim: "every pawn action,
  every random event must occur identically for everyone" ([[_grounding]] § RimWorld).

Properties, as stated by the sources:
- **Bandwidth is proportional to input volume, not to world size.** This is why Factorio can run
  megabase-scale worlds over consumer connections.
- **Determinism becomes a hard correctness requirement**, not a nice-to-have: "if functions
  produce random outputs, you can't use the lockstep architecture" ([[_grounding]] § Factorio).
- **The slowest peer sets the pace.** "The most fundamental limit of lock step architecture is
  that the game speed is limited by the slowest player, because to finish a frame input from all
  other peers needs to be processed" ([[_grounding]] § Factorio).
- **The failure mode is divergence, not lag** — see [[Determinism-and-Desync]].

## Family 2 — Server-authoritative state sync (one simulation, replicate results)

One process owns the world; clients send intents and receive state updates. Clients simulate
nothing authoritative.

- **Minecraft**: everything — "mob AI, redstone, chunk loading, player actions" — runs in a single
  server thread at a target 20 TPS ([[_grounding]] § Minecraft).
- **Terraria**: "clients connect to a server as a middleman, and clients cannot send or receive
  things directly from each other. The server is the owner of all NPCs" ([[_grounding]] §
  Terraria).
- **Space Station 13**: a single BYOND server with a Master Controller scheduling every
  subsystem ([[_grounding]] § SS13).
- **Eco**: one server simulating "a full ecosystem, economy, and government," default
  `MaxConnections` of 100 ([[_grounding]] § Eco).

Properties:
- **Determinism is not required for correctness** — there is only one authority. Terraria states
  the corollary duty plainly: "any non-deterministic decision must be synced to the clients,"
  because the client cannot re-derive it ([[_grounding]] § Terraria).
- **Bandwidth scales with the amount of changing state visible to each client**, which is why
  this family invests in interest management ([[Scale-Ceilings-and-Interest-Management]]).
- **The failure mode is degradation, not divergence.** When a Minecraft tick exceeds 50 ms, "the
  tick rate drops" — the world slows for everyone rather than splitting. SS13 makes the same
  tradeoff an operator dial: low ticklag means occasional stalls, high ticklag means "the game
  moves slow," with heuristics further modifying tick rate during spikes ([[_grounding]] §
  Minecraft, § SS13).

## Family 3 — Distributed / migrating authority (per-zone or per-entity ownership)

Authority is partitioned across participants rather than centralized or replicated.

- **Valheim** partitions by space: "the first player to enter an area acts as the host for the
  physics and logic in that zone" — described as "peer-to-peer within a dedicated server." Every
  entity is a **ZDO** (Zone Data Object) ([[_grounding]] § Valheim).
- **Project Zomboid** partitions by entity relevance: "ownership of zombies that are of relevance
  to the player at hand will be transferred — so the client impacted by the zombie in gameplay
  terms will have zero latency authority" ([[_grounding]] § Project Zomboid).

Properties:
- **Optimises local responsiveness**: the participant who cares most about an entity owns it, so
  their interaction with it has no round-trip.
- **Costs consistency at the seams.** PZ names the residual: attacking "a zombie that's chasing a
  friend" is owned by the friend's client and "will be at far higher risk of suffering from
  latency and will use client side prediction and cover-ups based on delayed information"
  ([[_grounding]] § Project Zomboid).
- **Replication efficiency dominates the ceiling.** Valheim resends whole objects — "every ZDO
  update resends the entire ZDO, regardless of what updated" — against a default ~64 KB/s cap
  ([[_grounding]] § Valheim).

## Family 4 (adjacent) — Persistent server, asynchronous participation

Not a synchronization scheme but an orthogonal posture that changes which problems apply. The
world "is always running, even when players aren't actively playing"; participants "log in at any
time and interact with an ongoing game world"; the family sits on "a spectrum from fully
concurrent to fully asynchronous" ([[_grounding]] § Asynchronous / persistent-world multiplayer).
A stated robustness property is that "server crashes have minimal impact on player experience."

Screeps is the worked example of the extreme end: a permanently running world where players
attach to read and to submit code, and where the per-tick cost of each player's participation is
metered ([[Per-Player-Compute-Budgets]]).

## How the families compare on the axes the brief names

| Axis | Lockstep | Server-authoritative | Migrating authority |
|---|---|---|---|
| What crosses the wire | inputs | state deltas | object state + ownership handoffs |
| Determinism required | yes, absolutely | no | no |
| Fidelity failure mode | silent divergence (desync) | none (single truth) | seam inconsistency |
| Performance failure mode | slowest-peer stall | tick-rate drop for all | bandwidth saturation per zone |
| Observed participant counts | small teams (RW MP), ~dozens (Factorio) | ~100 (Eco), server-bound | ~10 (Valheim), degrades at 5–6 |
| Scales with | input volume | visible changing state | entities per zone |

## Where promptworld already sits

promptworld's daemon is already Family 2 with Family 4 posture, before any multiplayer work:
a single authoritative goroutine, commands applied at tick boundaries, an append-only event log
as source of truth, and clients that receive committed events over a gapless subscription — see
[[Promptworld-Baseline]]. It additionally satisfies the lockstep precondition (deterministic RNG,
replayable log) that Family 1 requires, without currently using it for synchronization.

## Grounding

- [[_grounding]] § Factorio; § RimWorld; § Minecraft; § Terraria; § Valheim; § Project Zomboid;
  § SS13; § Eco; § Asynchronous / persistent-world multiplayer; § promptworld's own architecture
- [FFF #76 — MP inside out](https://www.factorio.com/blog/post/fff-76)
- [tModLoader Wiki — Basic Netcode](https://github.com/tModLoader/tModLoader/wiki/Basic-Netcode)
- [Edgegap — Valheim backend deep dive](https://edgegap.com/blog/valheim-multiplayer-game-backend-deep-dive)
