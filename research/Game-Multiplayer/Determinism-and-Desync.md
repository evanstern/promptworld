---
title: Determinism and Desync
aliases: [Fidelity, State Divergence]
tags: [determinism, desync, fidelity, replay]
type: note
created: 2026-08-01
updated: 2026-08-01
related: ["[[Game-Multiplayer]]", "[[Sync-Architectures]]", "[[Latency-and-Responsiveness]]", "[[Promptworld-Baseline]]"]
---

# Determinism and Desync

"Fidelity" in the multiplayer sense is the guarantee that every participant's world is the same
world. This note records what breaks that guarantee, how shipped games detect the break, and what
they do about it. Which of these problems apply at all depends on the architecture chosen
([[Sync-Architectures]]): they are existential for lockstep and largely absent under a single
authority.

## What determinism means operationally

Given the same starting state and the same ordered inputs, the simulation must produce the same
state — on every machine, every run. Factorio states the precondition negatively: "if functions
produce random outputs, you can't use the lockstep architecture, as the whole system screws up if
the functions that process things don't give the same results for each client, every time"
([[_grounding]] § Factorio).

## The catalogue of determinism hazards

From the general engineering literature ([[_grounding]] § Determinism hazards):

- **Unseeded or shared-state RNG.** Named as a primary source.
- **Hash-map iteration order.** Called out generically, and specifically for this codebase's
  language: "in Go, iterating over maps does not guarantee a consistent order" — documented as a
  live defect class in Cosmos SDK blockchain code, where determinism is a correctness
  requirement.
- **Sorts without stable tiebreakers.**
- **Uninitialized fields.**
- **Floating point across heterogeneous machines.** The nuance matters: "FP calculations are
  entirely deterministic, as per the IEEE Floating Point Standard, but that doesn't mean they're
  entirely reproducible across machines, compilers, OS's." Same-binary, same-hardware,
  same-execution-order replay does *not* suffer FP variance; cross-machine lockstep can.

## How the games mitigate

**Fixed seeds and RNG discipline.** RimWorld's Multiplayer mod uses "**fixed seeds** to ensure
deterministic outcomes across clients," plus a "**push/pop state pattern** to preserve RNG states
during non-deterministic operations" — i.e. the RNG state is explicitly saved and restored around
any code that would otherwise consume randomness off-schedule ([[_grounding]] § RimWorld).

**Continuous state fingerprinting.** The same mod does "**state tracking** to monitor RNG state
for desync detection" — divergence is caught by comparing a running hash rather than by waiting
for a visible symptom.

**A reference simulation (the arbiter).** RimWorld MP runs an arbiter process; desync reports
"after trying to resync are also useful, because they indicate there could be a problem with the
arbiter." Its role is to adjudicate *which* participant diverged rather than trusting a majority
([[_grounding]] § RimWorld desyncs).

**Making non-determinism explicit instead of eliminating it.** Under server authority the burden
inverts: Terraria does not try to make NPC AI reproducible, it requires that "any
non-deterministic decision must be synced to the clients," and desync "happens if random choice
values in AI code are not synced properly" ([[_grounding]] § Terraria).

## What actually causes desyncs in practice

RimWorld MP's own wiki is the most candid published account. Ranked as stated:

1. **Unsynchronized interface interactions** — "most of the time it happens due to unsynchronized
   interface interactions." A UI action that mutates simulation state on only one machine.
2. **Environment mismatch** — mod versions, mod *load order*, mod configuration, game version,
   corrupted files. Community framing: "if even one mod differs between players … you will
   desync."

The second category generalises beyond mods: any content, config, or data file that participates
in simulation and is not identical across participants is a desync source.

## Detection and diagnosis practice

- **Automatic capture.** RimWorld MP "automatically generates desync files when synchronization
  failures occur," written to an `MpDesyncs` folder beside saves, retaining the latest 10.
- **First report is the useful one.** "The most useful ones are the very first ones after a period
  of everything behaving correctly," because they pinpoint the triggering action. Later reports
  are noise from an already-diverged state.
- **Human context is part of the report.** Players are asked what they were "clicking, seeing"
  before the failure — the divergence trigger is frequently an interaction, not a background
  system.
- **Tick-by-tick state comparison** is the general debugging method: "record the complete initial
  state and all inputs from a real game session, then replay the simulation on different machines
  or builds and compare state dumps at every tick." The cheapest smoke test is stated as: "run the
  same recorded input sequence twice and confirm the final world state is identical down to the
  last value" ([[_grounding]] § Determinism hazards).

## Recovery

The surveyed material describes resync (rejoin from an authoritative state) rather than repair:
divergence is not merged, one side is discarded. RimWorld MP's guidance is overwhelmingly
preventative — environment parity — with resync as the in-session remedy and the desync file as
the artefact sent upstream for patching ([[_grounding]] § RimWorld desyncs).

## The architectural escape hatch

Under a single authoritative simulation the entire hazard class disappears, because there is no
second simulation to diverge from. That is the tradeoff Terraria, Minecraft, SS13, and Eco take:
they give up client-side simulation and pay in bandwidth and server load instead
([[Sync-Architectures]], [[Scale-Ceilings-and-Interest-Management]]).

## Applicability to promptworld

promptworld already meets the strict form of the lockstep precondition without needing it for
networking, and its replay discipline is the same one the desync-debugging literature prescribes:

- Deterministic per-decision RNG derived from `(seed, purpose, tick, index)` with **no RNG state
  carried**, which structurally removes the push/pop problem RimWorld had to patch around.
- An **append-only event log as the source of truth**, with state defined as a reducer over it —
  the exact "record inputs, replay, compare" substrate the general literature recommends.
- Byte-identity replay is already an enforced testing discipline in the repo's determinism
  harness, and the wiki records a standing doctrine note on reducer-constant replay hazards
  ([[Promptworld-Baseline]]).

The Go map-iteration hazard is the one item from the general catalogue that applies to this
codebase's language regardless of architecture ([[_grounding]] § Determinism hazards).

## Grounding

- [[_grounding]] § RimWorld; § RimWorld desyncs; § Factorio; § Terraria; § Determinism hazards;
  § promptworld's own architecture
- [Zetrith/Multiplayer Wiki — Desyncs](https://github.com/Zetrith/Multiplayer/wiki/Desyncs)
- [Gaffer On Games — Floating Point Determinism](https://gafferongames.com/post/floating_point_determinism/)
- [Ashouri — Go map iteration and determinism](https://ashourics.medium.com/the-challenge-of-gos-map-iteration-in-the-cosmos-sdk-blockchain-a-dive-into-determinism-bd5a99260519)
- [Bugnet — Debugging desync in deterministic lockstep games](https://bugnet.io/blog/how-to-debug-desync-in-deterministic-lockstep-games)
