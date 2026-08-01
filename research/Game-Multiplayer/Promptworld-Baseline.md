---
title: Promptworld Baseline
aliases: [Current Architecture, Starting Position]
tags: [baseline, architecture, current-state]
type: note
created: 2026-08-01
updated: 2026-08-01
related: ["[[Game-Multiplayer]]", "[[Brief-and-Assumptions]]", "[[Sync-Architectures]]", "[[Determinism-and-Desync]]", "[[Latency-and-Responsiveness]]", "[[Scale-Ceilings-and-Interest-Management]]", "[[Per-Player-Compute-Budgets]]", "[[World-Size-and-Density]]", "[[Multiplayer-Shapes]]", "[[Deity-Multiplayer]]", "[[Inter-Settlement-Interaction]]"]
---

# Promptworld Baseline

The starting position the multiplayer question lands on, stated as fact rather than judgment.
Recorded here so the branch can reason about applicability without cross-linking into the project
wiki (vault isolation). All claims are pinned to `docs/wiki/` at repo commit `1de512d9` and the
Backlog board as of 2026-08-01; the full quotations are in [[_grounding]] § promptworld's own
architecture.

## The part that already resembles a multiplayer server

- **Always-on daemon, attachable clients.** "A Go daemon advances the world 24/7 whether or not
  anyone is watching, and terminal clients attach and detach without affecting it." This is the
  persistent-world posture described in [[Multiplayer-Shapes]], already shipped.
- **Single authoritative simulation.** One goroutine owns all world state and advances it in
  deterministic ticks (1 tick = 1 game second); all external input enters as **commands applied at
  tick boundaries** and is recorded as events. This is the server-authoritative family in
  [[Sync-Architectures]].
- **Event log as source of truth.** Every event appends to a SQLite log in the world's save
  directory; state is a reducer over the log; snapshots bound recovery.
- **Log shipping with gapless replay already exists.** The IPC server broadcasts committed events
  to each session over a non-blocking send into a 1024-entry buffer; a subscription first fills
  from the store up to the log head (`subscribe{since}`), so a late or reconnecting client catches
  up without gaps.
- **Sessions are already isolated from the loop.** "A client can die mid-write, spam garbage, or
  subscribe and stall, and the loop never notices." Multiple concurrent sessions are supported;
  a long console turn "occupies only its session."
- **Deterministic RNG with no carried state** — per-decision PCG from `(seed, purpose, tick,
  index)` — plus a byte-identity replay testing discipline. Together these satisfy the strict
  determinism precondition discussed in [[Determinism-and-Desync]], without currently using it
  for synchronization.
- **World forking exists** (spec 076): a fork ceremony cutting a fresh prefix log at a snapshot
  boundary with `world.forked` lineage and the seed carried, plus fork/compare doors. Machine-wide
  `ps` enumerates every running world from live evidence.

## The part that is single-operator by construction

- **Transport is a Unix domain socket** inside the save directory — same-machine only. No network
  transport, authentication, or session security exists.
- **Sessions are anonymous.** Per TASK-65: no name, no id on an IPC session; event provenance is
  `Source: "planner"/"meeting"/"metatron"` with **no operator identity anywhere**. Nothing in the
  log can answer "whose prompt caused this."
- **The cost meter is one global ceiling** with per-provider attribution but **no per-operator
  attribution** — the gap [[Per-Player-Compute-Budgets]] compares against Screeps' per-account
  model.
- **One world = one save directory = at most one daemon process.** Multiple worlds mean multiple
  daemons.
- **No interest management.** Every subscribed session receives the full committed event stream;
  on buffer overflow the subscription is **cancelled** rather than degraded
  ([[Scale-Ceilings-and-Interest-Management]]).
- **One village per world, no world-map layer.** "Other villages" is not a modelled concept
  ([[Inter-Settlement-Interaction]]).

## The world itself

- **64×64 tile map** (`DefaultSize = 64`), seeded and **regenerated rather than stored**. The
  whole map is simulated; there is no chunk streaming or simulation-distance concept
  ([[World-Size-and-Density]]).
- **Eight sealed villagers** with write-once personas, situated episodic memory, embedding-based
  retrieval, journals, and nightly consolidation.
- **A social and governance substrate**: relationships, rumors, debts, secrets, conversations;
  norms and votes; a daily meeting under an event-sourced convention; a village charter.
- **A scenario/incident scheduler** (spec 054) including the **stranger** entity family — an
  outside actor arriving as an incident, the nearest existing analogue to DF's caravan/liaison
  pattern.

## The player's role — already a god game

Per [[Deity-Multiplayer]], promptworld implements the genre's full vocabulary single-player:

- **Villagers are sealed**; the player acts only as the **Guardian**, through indirect channels —
  omens, visions, standing orders, designations, hard directives, prophecy, canonization of named
  regions, and missions expressed in plain words.
- **A charge economy** prices miracles (time snap, item grant, entity move/remove), with gratis
  and premium tiers.
- **An endogenous faith loop** (spec 085): village faith is event-sourced over a five-reason delta
  table, and **charge regen is a pure faith-band function** — divine capacity derives from the
  population's belief, exactly the genre property.
- **An editable charter** setting the Guardian's persona and a compiled competence ceiling; the
  Guardian is itself an agent with its own memory and scheduled cognition lane (spec 102).
- **A four-stage curriculum ladder** gating the Guardian's tool ceiling by stage.

## The latency and throughput regime

- **Model latency dominates, not network latency.** The **cognition horizon** deterministically
  gates every model-reaching decision by how stale its answer will be when it lands; the
  **adaptive governor** turns the player's speed setting into "a ceiling, not a promise," shedding
  cognition under debt. The published LLM-NPC round-trip band is 3–7 s
  ([[Latency-and-Responsiveness]]).
- **Admission control already exists**: per-provider worker concurrency, slot-aware admission, a
  priority lane for conversations, per-provider circuit breakers, and a degraded mode at the spend
  ceiling.
- **The world slows rather than diverging** under load — the same invariant every surveyed
  server-authoritative game holds, reached here through cognition throttling rather than tick-rate
  drop.

## The recorded board position

TASK-65 ("Operator identity and attribution groundwork") is **To Do**, labelled `deferred`, and
holds the decision this branch informs:

- It records that "single-player on-laptop is the likely v1 posture, with multiplayer (self-host /
  modest paid hosting) undecided."
- Its AC #1 is "**multiplayer shape decision recorded (parallel villages vs shared village with
  per-player angels) with client sign-off**," and the card states "the shape decision gates
  everything else."
- Its remaining scope is identity on IPC sessions threaded into event provenance, per-operator
  attribution on Guardian turns/nudges/miracles/LLM spend, and preservation of anonymous
  single-player operation.
- It also records a retired idea: "one agent per coworker" (each player authoring one villager
  persona) was ruled out — villagers stay sealed, and indirect influence via the Guardian "is the
  entire point." See [[Brief-and-Assumptions]] § A5.

## Grounding

- [[_grounding]] § promptworld's own architecture — full quotations and file pins
- `docs/wiki/`: `overview.md`, `ipc-server.md`, `ipc-protocol.md`, `event-log.md`,
  `deterministic-rng.md`, `cognition.md`, `cognition-governor-debt.md`,
  `llm-budget-degraded-mode.md`, `llm-concurrency-leases.md`, `worldmap-generation.md`,
  `guardian.md`, `guardian-faith.md`, `guardian-miracles.md`, `guardian-designations.md`,
  `guardian-canonization.md`, `guardian-missions.md`, `social-fabric.md`, `governance.md`,
  `scenario-machinery.md`, `world-forking.md`, `instance-manager.md` — pinned at `1de512d9`
- Backlog card TASK-65
