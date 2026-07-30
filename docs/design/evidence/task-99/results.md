# TASK-99 / spec 098 — private dreams: seeded-world demonstration (SC-001)

**Date**: 2026-07-29 · **World**: `~/.promptworld/measure/task-99-demo` (preserved,
stopped) · **Seed**: 1337 · **Stage**: stage-4 (overridden) · **Branch**:
`task-99-private-dreams`

## Setup (local-only routes, zero paid spend)

- `llm.json`: villager routes (planner/conversation/consolidation/narrator/…)
  on local gemma (`http://mbpro-m1.local:11434/v1`, `gemma4:12b-mlx`);
  `embedding` on local ollama (`http://localhost:11434/v1`,
  `all-minilm:latest`); only `metatron` on the 9router proxy
  (`localhost:20128`, CC-subscription, $0 marginal). No paid API traffic; all
  consolidation markers carry no `cost_usd`.
- `world.json`: `memory_relevance: "on"` (embedder live).
- `tuning.json`: `{"dream_density_per_mille": 820, "dream_ambiguous_band_per_mille": 20}`
  — the spec-098 dials exercised through the spec-048 manifest (US3-AC1);
  remaining dream dials at doctrine defaults (habituation 500‰, merge cap 4,
  jitter 15‰). The boot log printed the full ten-dial applied line including
  the dream block.
- Run: day 1 06:00 → ~09:20 of natural play at 8–16x (governor-shed from a
  requested 32x under gemma cognition debt), one operator time-snap
  (`work snap-time 1 21:45 --force`) to reach night, then run through
  day 2 ~02:08. 90 `agent.memory_added`, 90 `agent.memory_embedded`
  (100% vector coverage).

## What the night produced

Duplicate-heavy streams formed naturally: repeated `Felled the tree at (x,y)`,
`Quarried the outcrop`, and `My foraging came to nothing` memories per agent
(Birch entered the night with 20 memories, ~half near-duplicates).

**Night 1: 8 slept, 6 consolidation markers landed (2 attempts still queued
behind gemma latency at stop — they retry next sleep), 8 dream events:
7 `agent.memory_merged` + 1 `agent.salience_revised`, across 7 agents.**

Store shapes (pre-dream → final):

| Agent | memories | dream outcome |
|---|---|---|
| Birch | 20 → 12 | 2 geometry merges (2+2 members) + 1 habituation (6★→3★) + **1 ambiguous group folded by the consult (`dream_folded:1` on the marker)** |
| Sage  | 13 → 9  | merge (2 members) |
| Fern  | 12 → 10 | merge (3 members: ticks 719/1154/5744 folded into 11614) |
| Ash   | 5 → 4   | merge (2 members) — **night REJECTED (`drift:idle`), merge landed anyway** |
| Cedar | 5 → 4   | merge (2 members) |
| Hazel | 7 → 6   | merge (2 members) |
| Oak   | 4 → 5   | no cluster ≥ 3 — untouched (distinct stores dream of nothing) |

Key event excerpts (seq/tick from the log):

```
seq 8906 t63062 agent.consolidated {"agent":"Birch","outcome":"accepted",...,"dream_folded":1}
seq 8905 t63062 agent.memory_merged {"agent":"Birch","kept":{"tick":1350,...},"merged":[...2 refs...]}
seq 8438 t61917 agent.salience_revised {"agent":"Birch","mem_tick":2700,"salience":3,"reason":"habituation"}
seq 9793 t65325 agent.consolidated {"agent":"Ash","outcome":"rejected","reason":"drift:idle"}
seq 9321 t64100 agent.memory_merged {"agent":"Ash",...}   <- geometry independent of the rejected night
```

## SC-001: the cluster collapsed; the distinct memory surfaces

Offline probe (state rebuilt twice from the log — once stopped before the
first dream event, once in full — then `sim.SelectMemories`, K=10, at the
final tick):

**Birch, pre-dream window** — crowded by the routine mass: four
`My foraging came to nothing` (6★) entries plus two `My chopping came to
nothing` (6★) entries filled 6 of 10 slots.

**Birch, post-night window** — the routine mass is gone (merged/faded);
the window now carries the distinct moments: the neglect percept (9★), the
accusation conversation with Sage (7★), the time-lurch omen (10★), the
evening chronicle gist, the surviving representative chops. Same shape for
Sage (13→9) and Fern (12→10): every merge kept the newest/most salient
representative vivid and removed only near-duplicates; distinct memories
(fires built, conversations, discoveries) were untouched in all 8 stores.

## Doctrine checks exercised live

- **D1 privacy**: every dream event's inputs were one agent's store (the
  pass runs on the per-agent `consolJob` snapshot); the sim-level
  perturbation proof is `TestPlanDreamPrivacyPerturbation`.
- **D2 geometry-first**: 7/8 outcomes decided by geometry alone (no LLM);
  exactly one ambiguous group consulted the existing consolidation call and
  landed as a recorded fold (`dream_folded:1`) — no new call class, zero
  extra calls.
- **D3 recorded outcomes**: all 8 outcomes are whitelisted injected events;
  Ash's rejected night proves the geometry batch is independent of the LLM
  night's fate. Replay byte-identity is gated by `TestDreamReplayByteIdentity`.
- **D4 noise**: jitter 15‰ live (default); zeroed-dial equivalence and
  cross-seed boundary variance are gated by `TestDreamNoiseZeroedEquivalence`
  / `TestDreamNoiseVariesAcrossSeeds`.

## Reproduce

```sh
promptworld new ~/.promptworld/measure/task-99-demo --seed 1337 --stage stage-4 --override
# llm.json / tuning.json / world.json memory_relevance as above
promptworld start ~/.promptworld/measure/task-99-demo && promptworld speed ... 32x
# accumulate a morning, then: promptworld work ... snap-time 1 21:45 --force
# after night 1: promptworld tail ... | grep -E 'salience_revised|memory_merged|consolidated'
```

The offline window probe replays the log through `sim.State.Apply` and prints
`sim.SelectMemories` pre/post — a throwaway `package main` over
`internal/{sim,store,world}` (not shipped; see this file's PR for the exact
listing in the task record).
