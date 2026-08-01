---
title: Per-Player Compute Budgets
aliases: [Screeps CPU Model, LLM Cost Scaling]
tags: [economics, cpu-budget, llm-cost, screeps]
type: note
created: 2026-08-01
updated: 2026-08-01
related: ["[[Game-Multiplayer]]", "[[Latency-and-Responsiveness]]", "[[Scale-Ceilings-and-Interest-Management]]", "[[Promptworld-Baseline]]"]
---

# Per-Player Compute Budgets

In every traditional multiplayer game surveyed, an additional player costs bandwidth and a share
of a fixed server tick. In a game whose inhabitants think via model calls, an additional player
can also cost **money per turn**. Only two published systems in this survey meter participation
as a first-class resource: Screeps (CPU) and the commercial LLM-NPC literature (tokens).

## Screeps — the published model for metering participation

Screeps is a persistent, always-on world where each player's *code* runs on the server every
tick. Its resource model is the closest published analogue to per-player agent cognition
([[_grounding]] § Screeps).

**Tick structure.** Two sequential stages: "(1) player scripts calculation — all active player
scripts execute; (2) commands processing — game world rooms process the resulting commands."
Database updates are then applied in bulk. Critically, "**a tick ends when all scripts of all
players have been executed to the end**" — tick duration is an *output* of participation, not a
fixed budget the world enforces.

**Per-player allowance.** "Every player gets a basic amount of CPU (20 at the time of writing),
and can unlock 10 more per GCL, up to a maximum (300 currently)." The unit is wall time: "CPU
time limit is a duration of time in milliseconds during which your game script is allowed to run
within one tick. The CPU limit 100 means that after 100 ms execution of your script will be
terminated even if it has not accomplished some work yet."

**Smoothing via a bucket.** Unused allowance accrues: "if a script during a tick worked less time
than the account CPU baseline limit, the resulting difference is added to a cumulative bucket. You
may accumulate up to 10,000 CPU. If the bucket contains any accumulation, your script can overrun
your CPU limit using up to 500 CPU per tick." This lets expensive occasional deliberation coexist
with a low steady-state allowance — a burst budget on top of a rate limit.

**Enforcement is hard, not cooperative.** Scripts run in Node `vm` contexts in forked processes;
"if execution exceeds the player-specific timeout, **the entire fork terminates** rather than
gracefully stopping — requiring context recreation."

**Isolation for correctness.** "One core synchronously processes one room or player, preventing
race conditions."

**Sharding splits the allowance.** Each shard has its own database, map, and runtime servers; the
player sets a per-shard CPU limit and "their total sum should match your account limit."

## LLM inference economics at multiplayer scale

Published figures ([[_grounding]] § LLM inference cost and latency):

- **Per-session**: "cloud AI creates a per-session cost of **$0.01–0.05** that scales linearly
  with player engagement."
- **At commercial scale**: 100,000 DAU × ten NPC conversations per session ≈ **$500k–$2M/year**
  at 2026 token rates (cited bands: $0.50–1.00 and $0.75–1.50 per million tokens).
- **The multi-agent multiplier is named explicitly**: "one of the main bottlenecks is the high
  computational cost of real-time LLM inference, especially in **multi-agent settings where
  several NPCs must perceive, reason, and respond simultaneously**."
- **Latency**: "three to seven seconds of round-trip time" for cloud NPC turns.
- **The local alternative**: edge/local inference gives "zero marginal cost per user and scales
  infinitely with the player base," at the cost of moving to "optimized smaller models."

The structural point is that **cost scales with simulated agents × cognition frequency**, and
only indirectly with player count — which makes the multiplayer *shape* an economic decision as
much as a design one ([[Multiplayer-Shapes]]).

## How the two metering models differ

| | Screeps CPU | LLM tokens |
|---|---|---|
| Scarce resource | server wall time per tick | money per call |
| Who consumes it | the player's own script | the simulated agents |
| Enforcement | hard timeout, fork killed | budget ceiling / degraded mode |
| Burst handling | accumulating bucket (max 10,000) | none published |
| Effect of exhaustion | that player's turn is cut | that decision is not made |

## Applicability to promptworld

promptworld already implements the second column, for a single operator
([[Promptworld-Baseline]]):

- **A single global spend ceiling** with per-provider attribution, per-provider circuit breakers,
  and a **degraded mode** when the ceiling is reached.
- **The cognition horizon** as a hard admission gate: a decision is routed to a model only if its
  answer will still be fresh when it lands; otherwise it falls back to reflex.
- **An adaptive governor** with a debt feedback controller that turns the requested speed into a
  ceiling rather than a promise — structurally the same "tick ends when the work is done" posture
  Screeps takes, expressed as speed suppression.
- **Per-provider worker concurrency and slot-aware admission**, with a **priority lane** for
  conversations — i.e. work-class prioritisation already exists.

What does **not** exist today, per TASK-65: any notion of **who** consumed the budget. "The cost
meter is one global monthly ceiling with **no per-operator attribution**," and IPC sessions carry
no identity, so no event's provenance can answer "whose prompt caused this." The Screeps model —
per-account allowance, per-shard split, accumulating bucket, hard enforcement — is the published
prior art for what per-operator metering looks like when participation is the cost centre.

## Grounding

- [[_grounding]] § Screeps; § LLM inference cost and latency; § promptworld's own architecture
- [Screeps Docs — Server-side architecture](https://docs.screeps.com/architecture.html)
- [Screeps Docs — How does CPU limit work](https://docs.screeps.com/cpu-limit.html)
- [Inworld — LLM Inference Cost at Scale](https://inworld.ai/resources/llm-inference-cost-at-scale)
- [Naavik — AI NPCs](https://naavik.co/digest/ai-npcs-the-future-of-game-characters/)
