---
title: Latency and Responsiveness
aliases: [Latency Hiding, Netcode Latency]
tags: [latency, prediction, tick-rate, responsiveness]
type: note
created: 2026-08-01
updated: 2026-08-01
related: ["[[Game-Multiplayer]]", "[[Sync-Architectures]]", "[[Determinism-and-Desync]]", "[[Per-Player-Compute-Budgets]]", "[[Promptworld-Baseline]]"]
---

# Latency and Responsiveness

Latency in this family of games is not the twitch-shooter problem. The surveyed games run at
10–20 ticks per second or slower, and their latency work is about (a) hiding the round-trip on
*player input*, and (b) deciding what the world does when the simulation itself cannot keep up.

## Two distinct latencies

1. **Network round-trip on player input** — the gap between acting and seeing the act take
   effect. Addressed by prediction / latency hiding.
2. **Simulation latency** — the tick itself taking longer than its budget. Addressed by slowing
   the world, dropping tick rate, or shedding work.

The surveyed games treat these with different machinery, and the second dominates in
simulation-heavy titles.

## Latency hiding — Factorio's mechanism, in detail

Factorio's account ([[_grounding]] § Factorio latency hiding) is the most precisely documented in
the survey. The problem: under lockstep "all the players need to apply all the user actions in the
same order," so an action can only execute after a round trip.

The mechanism:
- A separate **"latency state"** layer duplicates the relevant game state.
- **Every tick** that layer is **reset from the authoritative game state**.
- It then **replays all buffered local user actions not yet confirmed** by the server.
- The player sees their own action immediately; the authoritative world catches up.

Actions Factorio latency-hides: player movement; entity selection; opening/closing GUIs; building
and fast-replacing; mining resources and buildings; picking items to cursor.

Actions it deliberately does **not**: "we don't plan to do any latency hiding for interacting
with entities (apart from basic operations like opening, rotating, etc.) or fighting" — the
cascading state changes are too entangled to speculatively apply.

The property that makes it safe: it is **self-correcting**. Because the latency layer is
reinitialized from authoritative state every tick, a wrong local prediction is silently
discarded on the next tick and can never accumulate into a desync
([[Determinism-and-Desync]]).

## Zero-latency authority — Project Zomboid's alternative

Instead of predicting, PZ **moves authority to the client that cares**: "ownership of zombies
that are of relevance to the player at hand will be transferred — so the client impacted by the
zombie in gameplay terms will have zero latency authority over the zombie's actions. This aims to
make combat identical to that of the single player experience."

The cost is named in the same source: an entity owned by someone else's client — "a zombie that's
chasing a friend" — "will be at far higher risk of suffering from latency and will use client side
prediction and cover-ups based on delayed information" ([[_grounding]] § Project Zomboid). The
technique relocates the latency rather than removing it.

## What happens when the *simulation* is the bottleneck

This is the failure mode that dominates in simulation-heavy games, and every surveyed
server-authoritative title resolves it the same way: **slow the world, keep everyone together.**

- **Minecraft**: 20 TPS is a target, not a guarantee — "if the server is overloaded such that a
  tick takes more than 50 ms, the tick rate drops." Game time itself dilates
  ([[_grounding]] § Minecraft).
- **Space Station 13**: ticklag is an explicit operator dial with a stated tradeoff — "if you have
  a low tick lag the server sometimes lags while it calculates things whereas if you have a high
  tick lag there's no lag as such but the game moves slow" — plus heuristic tick-rate modification
  during spikes ([[_grounding]] § SS13).
- **Factorio (lockstep)**: has no such option; instead "the game speed is limited by the slowest
  player," which is the same outcome imposed by a different mechanism ([[_grounding]] § Factorio).
- **Screeps**: a tick simply "ends when all scripts of all players have been executed to the end"
  — tick duration is an output, not an input ([[_grounding]] § Screeps).

Across the survey the invariant is: **nobody desyncs to preserve speed; speed is what gets
sacrificed.**

## Transport contributes measurably

Relay hops are a documented latency and reliability tax:
- Terraria's "Host & Play routes traffic through Steam relay servers instead of a direct
  connection, which adds delay and raises packet loss risk" ([[_grounding]] § Terraria).
- Valheim's crossplay backend routes through Azure PlayFab relays, and "crossplay players are
  more likely to experience lag, timeouts, and disconnects" than the direct Steam backend
  ([[_grounding]] § Valheim).

Both ship *both* options and let the operator choose reach versus latency.

## The LLM-specific latency band

For a game whose agents think via model calls, a third latency appears that none of the
traditional netcode literature addresses. Published figures for LLM-driven NPCs: "current
cloud-based NPC setups average **three to seven seconds of round-trip time**" ([[_grounding]] §
LLM inference cost and latency). This is two to three orders of magnitude above the network
latencies the netcode literature optimises, which means in an LLM-agent game **model latency, not
network latency, sets the responsiveness floor** for anything gated on a model call.

Local/edge inference is the stated alternative, trading capability for "zero marginal cost per
user" and no cloud round-trip.

## Applicability to promptworld

- promptworld's clients are **observers plus an operator console**, not simulators: the world
  advances on the daemon and clients receive committed events. The Factorio-style prediction
  problem therefore applies only to operator input feedback, not to world simulation
  ([[Promptworld-Baseline]]).
- The game already implements the "sacrifice speed, not consistency" pattern that every surveyed
  server-authoritative title converges on — and does so *because of model latency rather than
  network latency*. The cognition horizon "deterministically gates every model-reaching decision
  by how stale its answer will be when it lands," and the adaptive governor "turns the player's
  speed setting into a ceiling, not a promise" ([[_grounding]] § promptworld's own architecture).
  This is structurally the same device as Minecraft's TPS drop and SS13's ticklag heuristics.
- The 3–7 s LLM round-trip band is already the dominant latency term in the single-player game;
  adding operators adds console turns and (depending on shape) more agent cognition to the same
  budget — see [[Per-Player-Compute-Budgets]].

## Grounding

- [[_grounding]] § Factorio latency hiding; § Project Zomboid; § Minecraft; § SS13; § Terraria;
  § Valheim; § Screeps; § LLM inference cost and latency; § promptworld's own architecture
- [FFF #83 — Hide the latency](https://www.factorio.com/blog/post/fff-83)
- [The Indie Stone — Zed Clients](https://projectzomboid.com/blog/news/2020/06/zed-clients/)
- [Veriprajna — Edge AI gaming latency](https://veriprajna.com/technical-whitepapers/gaming-ai-edge-computing-latency)
