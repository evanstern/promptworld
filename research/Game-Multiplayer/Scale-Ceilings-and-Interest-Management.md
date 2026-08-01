---
title: Scale Ceilings and Interest Management
aliases: [Player Limits, Area of Interest, AOI]
tags: [scale, bandwidth, interest-management, sharding]
type: note
created: 2026-08-01
updated: 2026-08-01
related: ["[[Game-Multiplayer]]", "[[Sync-Architectures]]", "[[World-Size-and-Density]]", "[[Per-Player-Compute-Budgets]]", "[[Promptworld-Baseline]]"]
---

# Scale Ceilings and Interest Management

What limits the number of participants, and what the surveyed games do to raise that limit. The
recurring finding is that **the observed ceiling is a property of the replication scheme, not of
the world's size or the game's design intent**.

## Observed ceilings

| Game | Practical ceiling | Stated binding constraint |
|---|---|---|
| Valheim | 10 (degrades at 5–6) | ZDO send/receive rate limits; whole-object resends; ~64 KB/s cap |
| Terraria | server-bound | server owns all NPCs; per-client state pushes |
| Project Zomboid | server-bound | server simulates "across the whole map" |
| Minecraft | server-bound | single-threaded tick; chunk generation |
| Eco | ~100 default `MaxConnections` | "depends on world size and server resources" |
| MMO zones (general) | ~100 per zone | zone hosting capacity |
| Screeps | thousands, across shards | per-player CPU allowance; per-shard databases |

Sources: [[_grounding]] § Valheim, § Terraria, § Project Zomboid, § Minecraft, § Eco,
§ Interest management, § Screeps.

## The Valheim case: a ceiling that is explicitly not a design decision

Valheim is the clearest documented example of netcode setting a gameplay limit. "Valheim's
10-player limit is a **networking decision, not a game design one**, as the ZDO object-sync
system degrades under higher player counts due to bandwidth pressure." Community teardown found
"hard-coded send and receive rate limits in the ZDO manager," and even with mods "performance
degrades noticeably approaching 5–6 players."

Two compounding causes are named:
1. **Replication granularity.** "Every ZDO update resends the entire ZDO, regardless of what
   updated."
2. **Entity count independent of player count.** "Large player-built structures and tamed animal
   populations add pressure" — a base with "5,000+ building pieces" saturates the ~64 KB/s
   default cap "instantly, causing everything to 'freeze'" ([[_grounding]] § Valheim).

The second point is the load-bearing one for any building-heavy simulation: **the world's own
complexity consumes the same budget as the players do.**

## Interest management — the standard mitigation

The problem statement from the MMO literature: "bandwidth limitations … make it infeasible for
all participants to receive all data from all others, so data must be filtered according to user
interest" ([[_grounding]] § Interest management).

Three canonical filtering schemes:
- **Zone-based** — spatial partitioning; you receive your zone.
- **Aura-based** — a subscription radius around each participant ("interested in any entity under
  100 m away").
- **Visibility-based** — filtering by what is actually perceivable.

Implementation notes recorded in the sources: the spatial query structure is independent of the
policy (quadtrees suit arbitrary object sizes); "only objects currently in range 'exist' for the
client to interact with"; **hysteresis prevents excessive updates** at boundaries; and the
subscription list lives "outside of a spatial query structure." Hierarchical dissemination "can
greatly save network bandwidth and alleviate each node's workload."

Zoning doubles as load distribution: "zoning is an efficient way to distribute server load,
frequently used in large sandbox games or open-world MMOs like World of Warcraft. Depending on the
game engine used, a zone can typically host ~100 players."

## Minecraft's version: interest management as operator dials

Minecraft exposes the same idea as two tunables with documented effects:
- **View distance** — "reducing view distance from the default 10 to 8 (or even 6) massively
  reduces chunk generation load."
- **Simulation distance** — "controls how far entities tick — 5 is the sweet spot."

Note the distinction: view distance governs *replication*, simulation distance governs *whether
the world updates at all* out there. The second is a simulation-fidelity tradeoff, not a
bandwidth one. Load attribution is stated bluntly: "chunk generation is the #1 cause of lag
spikes. When players explore new territory, the server has to generate terrain, caves,
structures, and biomes in real-time" ([[_grounding]] § Minecraft).

## Sharding — the ceiling-raising move

Screeps partitions the *world* rather than the view: "the consistent game world is divided into
**shards**, each with its own database of game objects, own game map, and own set of connected
runtime servers." A player's compute allowance is then divided across shards, and "their total
sum should match your account limit" ([[_grounding]] § Screeps).

The published infrastructure gives a sense of the cost of the approach: 40 quad-core runtime
servers (160 Xeon cores), and per-shard MongoDB on 24-core / 128 GB machines handling 30k updates
per second. Concurrency is avoided rather than managed: "one core synchronously processes one
room or player, preventing race conditions."

## The design-side lever: scale the rules, not just the plumbing

The game-design literature frames scaling as a rules problem: "scaling in game design means a game
retains a similar experience regardless of the number of players by **changing rules, numbers, or
other design elements based on player count**" ([[_grounding]] § World size, density, and player
count). This is an independent axis from replication — it addresses whether the game *plays* well
at N, not whether the wire can carry N.

## Applicability to promptworld

- The daemon already **broadcasts committed events to every subscribed session** with no spatial
  filtering: a non-blocking send into a 1024-entry buffer per session, with subscription
  cancellation on overflow, and gapless catch-up from the store at subscribe time
  ([[Promptworld-Baseline]]). That is a full-fidelity firehose — no interest management exists,
  because with one observer none was needed.
- The overflow behaviour is a *drop-the-subscriber* policy, which is a scale-relevant fact:
  under a fan-out of many sessions the existing backpressure design cancels rather than degrades.
- Event volume, not player count, would be the driver here: promptworld's per-tick output is
  agent/cognition events over a 64×64 grid, not per-entity position updates
  ([[World-Size-and-Density]]).
- The genuinely novel ceiling for this game is not bandwidth but **model spend and cognition
  throughput** — see [[Per-Player-Compute-Budgets]].

## Grounding

- [[_grounding]] § Valheim; § Minecraft; § Eco; § Interest management; § Screeps; § World size,
  density, and player count; § promptworld's own architecture
- [Boulanger — Interest Management for MMGs (PDF)](https://www.cs.mcgill.ca/~jboula2/thesis.pdf)
- [Edgegap — Valheim backend deep dive](https://edgegap.com/blog/valheim-multiplayer-game-backend-deep-dive)
- [Screeps Blog — World Shards Launched!](https://blog.screeps.com/2017/08/shards/)
