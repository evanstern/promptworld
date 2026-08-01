---
title: Game-Multiplayer
aliases: [Multiplayer, MP Architecture and Gameplay]
tags: [moc, multiplayer, netcode, colony-sim, sandbox, god-games]
type: moc
created: 2026-08-01
updated: 2026-08-01
related: []
---

# Game-Multiplayer

How multiplayer works in the games this one belongs to — colony sims, open-world sandboxes,
persistent shared simulations, and LLM-agent villages. Covers both halves of the question: the
**engineering** (synchronization, fidelity/determinism, latency, scale ceilings, per-player
compute) and the **gameplay** (what players share, how deities interact, how settlements relate,
whether the map has to grow). Descriptive throughout; no recommendation is made here.

## Scope

**In:** synchronization architectures and their failure modes; determinism and desync; latency
hiding and simulation-rate degradation; player-count ceilings and interest management; per-player
compute/cost metering; multiplayer gameplay shapes; god-game multiplayer; inter-settlement
interaction verbs; world-size and density evidence; promptworld's own current architecture as
grounded fact.

**Out:** any verdict on which shape promptworld should adopt (that is the analyze phase, and
TASK-65's AC #1); implementation planning; twitch-genre netcode (sub-100 ms hit registration);
accounts, matchmaking, anti-cheat, monetisation. Constraints and assumptions:
[[Brief-and-Assumptions]].

## What is known

- **The direct family has no official multiplayer, and the reason is architectural.** Dwarf
  Fortress ships none: the game "was built around a single thread," and the systems that resist
  parallelization are named as "AI, character logic, tasks and errands, pathfinding" — exactly
  what multiplayer would have to keep consistent. RimWorld ships none either; both of its
  multiplayer implementations are community mods ([[_grounding]] § Dwarf Fortress, § RimWorld).

- **Three synchronization families cover the field, and the choice fixes everything downstream.**
  Deterministic lockstep sends inputs and replays the sim everywhere (Factorio, RimWorld MP);
  server-authoritative state sync runs one simulation and replicates results (Minecraft, Terraria,
  SS13, Eco); migrating authority partitions ownership by zone or entity (Valheim, Project
  Zomboid). Each fixes what must be deterministic, what crosses the wire, and what breaks first
  ([[Sync-Architectures]]).

- **Fidelity is a lockstep problem specifically.** Under a single authority the desync class
  vanishes. Under lockstep it is existential, and the documented causes are mundane:
  "unsynchronized interface interactions" first, environment mismatch second. Mitigations are
  fixed seeds, RNG push/pop, continuous state fingerprinting, and an arbiter process; the general
  hazard catalogue names hash-map iteration order — including Go's explicitly — alongside
  unstable sorts and cross-machine floating point ([[Determinism-and-Desync]]).

- **Under load, every surveyed game sacrifices speed rather than consistency.** Minecraft's tick
  rate drops below 20 TPS; SS13 exposes ticklag as a dial and modifies it heuristically during
  spikes; Factorio's lockstep is "limited by the slowest player"; a Screeps tick "ends when all
  scripts of all players have been executed." Nobody desyncs to stay fast
  ([[Latency-and-Responsiveness]]).

- **Latency hiding is a bounded, self-correcting trick.** Factorio duplicates state into a
  "latency state" layer, resets it from authoritative state every tick, and replays unconfirmed
  local inputs — applied to movement, selection, GUIs, building and mining, but deliberately
  **not** to entity interaction or combat. Project Zomboid takes the alternative route, migrating
  entity ownership to the client that cares, and names the residual cost at the seams
  ([[Latency-and-Responsiveness]]).

- **Observed player ceilings are set by the replication scheme, not the design.** Valheim's
  10-player limit is "a networking decision, not a game design one" — whole-object ZDO resends
  against a ~64 KB/s cap, degrading at 5–6 players; Eco defaults to 100 and depends on "world size
  and server resources"; MMO zones host ~100. Interest management (zone-, aura-, visibility-based,
  with hysteresis) is the standard mitigation; sharding is the ceiling-raising move
  ([[Scale-Ceilings-and-Interest-Management]]).

- **Two published systems meter participation as a first-class resource.** Screeps gives each
  player a millisecond CPU allowance per tick (20 baseline, up to 300), an accumulating bucket to
  10,000 for bursts, hard fork-kill enforcement, and a per-shard split. The LLM-NPC literature
  meters money instead: $0.01–0.05 per session scaling linearly with engagement, 3–7 s round-trip
  latency, and an explicitly named multi-agent cost multiplier ([[Per-Player-Compute-Budgets]]).

- **Four gameplay shapes exist, and RimWorld's two mods took opposite ones.** Zetrith's
  Multiplayer puts many hands on one colony under a shared clock; RimWorld Together gives each
  player a separate colony on a shared planet, explicitly "keeping separate colonies and pace"
  and working "regardless of mods or DLCs" — evidence of how much weaker the coupling is. Add
  competitive shared worlds (god games) and shared-stakes politics (Eco). Synchronous vs
  asynchronous participation is an orthogonal axis ([[Multiplayer-Shapes]]).

- **God-game multiplayer means competing deities over separate populations.** The genre's grammar
  is deity → own followers → world → other followers; gods do not act on each other directly, and
  faith/belief is both the fuel and the scoreboard. Cooperative pantheons over one shared people
  are absent from the published record ([[Deity-Multiplayer]]).

- **Inter-settlement interaction has a stable verb grammar** — trade, visit, raid/war, persistent
  diplomatic standing, territory, knowledge exchange — and the richest model comes from a
  single-player game. DF keeps sites on an abstract world map and makes contact episodic
  (seasonal caravans), stateful (attitude accumulates), and asymmetric (the other settlement is
  never fully simulated). RimWorld Together reuses nearly the same verbs with humans behind them;
  Haven & Hearth instead makes settlements spatial claims with radius, membership rights, and an
  authority upkeep that drains ([[Inter-Settlement-Interaction]]).

- **More players does not straightforwardly mean a bigger map.** Density beats extent for
  perceived scale; server cost tracks generation and simulated entities rather than area (Eco's
  ceiling *falls* as world size rises); spreading players out is implemented as zoning or
  sharding, not raw area; and the family's standard way to add settlements is a **separate
  abstract world-map layer**, not more tactical tiles. Scaling can also be a rules change rather
  than a space change ([[World-Size-and-Density]]).

- **promptworld already is a server-authoritative, event-sourced, persistent-world daemon** with
  gapless log shipping to multiple concurrent sessions, deterministic RNG, replay discipline, a
  full god-game vocabulary (charges, miracles, endogenous faith, charter, designations, prophecy,
  canonization), model-latency admission control, and world forking. What it lacks is network
  transport, operator identity, per-operator attribution, interest management, and any concept of
  a second village ([[Promptworld-Baseline]]).

## Notes

- [[Brief-and-Assumptions]] — the request restated, assumptions made, ambiguities flagged
- [[Promptworld-Baseline]] — the code-grounded current architecture the question lands on
- [[Sync-Architectures]] — lockstep vs server-authoritative vs migrating authority, compared
- [[Determinism-and-Desync]] — what breaks fidelity, how shipped games detect and recover
- [[Latency-and-Responsiveness]] — latency hiding, zero-latency authority, tick-rate degradation
- [[Scale-Ceilings-and-Interest-Management]] — observed player limits and how they are raised
- [[Per-Player-Compute-Budgets]] — Screeps' CPU model and LLM inference economics
- [[Multiplayer-Shapes]] — the four gameplay shapes and the sync/async axis
- [[Deity-Multiplayer]] — how god games handle multiple deities; what the genre leaves unanswered
- [[Inter-Settlement-Interaction]] — trade/visit/raid/diplomacy/territory across DF, RW, H&H, Eco
- [[World-Size-and-Density]] — whether the map must grow, and what actually costs

## Analyses

_Opinionated evaluations built on this branch (added by the analyze phase). Empty until then._

## Open questions

- **Cooperative multi-deity designs are unattested.** The god-game record covers rival deities
  over rival populations; no surveyed title puts two patrons over one people. Eco's ratified-
  constitution model is the nearest shipped analogue for many actors steering one population.
- **Asynchronous divine play is unattested.** No surveyed god game runs while a deity is away.
- **No published precedent for per-player LLM budgets in a shared simulation.** Screeps meters CPU;
  the LLM-NPC literature meters aggregate spend. Nothing meters *per operator in a shared world*.
- **Whether inter-village verbs need the other village simulated at full fidelity.** DF's answer
  is no (asymmetric abstraction); RimWorld Together's is yes (a real colony behind each site).
  The cost difference is not quantified in any source found.
- **Tick-rate/event-volume figures for promptworld under multiple subscribers** are not measured
  anywhere; the 1024-buffer overflow-cancels behaviour has not been characterised under fan-out.
- **Whether a world-map layer or world-forking/sharding is the cheaper route to "many villages"**
  — both primitives exist in adjacent form; neither has been costed.
- **What "same world, different pace" would mean for an LLM-agent sim**, where cognition cost is
  proportional to simulated time. RimWorld Together's independent-pace model has no cost analogue
  here.

## Grounding

- [[_grounding]] — the research pass this branch is built on (web-search fan-out, 2026-08-01)
- [FFF #83 — Hide the latency](https://www.factorio.com/blog/post/fff-83)
- [Factorio Wiki — Desynchronization](https://wiki.factorio.com/Desynchronization)
- [Zetrith/Multiplayer Wiki — Desyncs](https://github.com/Zetrith/Multiplayer/wiki/Desyncs)
- [Steam Workshop — RimWorld Together](https://steamcommunity.com/sharedfiles/filedetails/?id=3005289691)
- [Edgegap — Valheim multiplayer backend deep dive](https://edgegap.com/blog/valheim-multiplayer-game-backend-deep-dive)
- [The Indie Stone — Zed Clients](https://projectzomboid.com/blog/news/2020/06/zed-clients/)
- [Screeps Docs — Server-side architecture](https://docs.screeps.com/architecture.html)
- [Boulanger — Interest Management for MMGs (PDF)](https://www.cs.mcgill.ca/~jboula2/thesis.pdf)
- [DF Wiki — Civilization](https://dwarffortresswiki.org/index.php/Civilization)
- [Haven and Hearth Wiki — Village](https://havenandhearth.fandom.com/wiki/Village)
- [Inworld — LLM Inference Cost at Scale](https://inworld.ai/resources/llm-inference-cost-at-scale)
