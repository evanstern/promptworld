---
id: TASK-159
title: >-
  Chop/quarry never updates mental maps: villagers rediscover their own
  harvesting as loss (75% of all memories)
status: In Progress
assignee: []
created_date: '2026-07-26 21:02'
updated_date: '2026-07-26 22:02'
labels: []
dependencies: []
priority: high
ordinal: 1
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
SHOWSTOPPER. Observed live on world "worldy": 75% of all agent.memory_added events (346 of 461) are "was gone when you looked" discoveries, flooding the WindowK=10 working-memory window and driving journals/conversations about the world "becoming barren" — which is false.

Root cause: the agent.chopped reducer (internal/sim/state.go:1057) appends the tile to s.Cleared but removes the tree fact from NO mental map — not the chopper's, not on-scene witnesses'. agent.quarried (state.go:1176) has the same gap. The spec 041 US3 perception sweep (perceptionEvents, internal/sim/executor.go:566-586) then finds the remembered fact absent from ground truth and emits agent.map_corrected plus a salience-5 memory "The tree at (x,y) had been felled when you looked." Verified: ALL 103 chops in worldy produced a self-correction within ~2-10 ticks (the actor "discovers" the tree they just felled), and every past witness gets their own copy on next approach — 120 real removals (103 chops + 17 quarries) fanned out into 346 loss memories. Chopping itself deliberately creates no memory (executor.go:1317 spam-avoidance precedent), so a villager's own labor is remembered EXCLUSIVELY as third-party-voiced destruction.

Fix direction (operator call, 2026-07-26): at chop/quarry time, remove the fact from the mental map of the actor and of every agent within witnessRadius (they watched it happen — no discovery pending); the actor's memory of the act is FIRST-PERSON ("Felled the tree at (x,y)."). What is stored as memory may be re-evaluated later. agent.map_corrected remains for its intended narrative: an agent who was elsewhere returns later and genuinely finds the place changed.

Corrections-pass reducer is fine (fact removed after correction, no repeats in the data). Touches the reducer contract and replay — needs a spec.

Spec: specs/081-first-person-harvest-memory
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The agent.chopped and agent.quarried reducers remove the corresponding place fact from the actor's mental map and from the map of every agent within witnessRadius at the event tick
- [x] #2 The actor receives a first-person situated memory of the act (e.g. "Felled the tree at (x,y).") instead of a later map_corrected discovery
- [x] #3 agent.map_corrected no longer fires for the actor or for agents who were within witnessRadius of the act when it happened; it still fires for agents who return later
- [x] #4 Replay determinism holds: a fresh replay of an existing event log reproduces identical state (reducer contract respected)
- [x] #5 Live verification on a fresh world: share of 'gone when you looked' memories drops to genuine return-discoveries only
- [ ] #6 Spec phase: Setup
- [ ] #7 Spec phase: Foundational (blocking prerequisites)
- [ ] #8 Spec phase: User Story 1 — My own harvest is mine, in the first person (P1) 🎯 MVP
- [ ] #9 Spec phase: User Story 2 — Watching a neighbor harvest is not a later mystery (P2)
- [ ] #10 Spec phase: User Story 3 — Genuine return-discovery still works (P3, regression guard)
- [ ] #11 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Spec'd + planned 2026-07-26: specs/081-first-person-harvest-memory (spec, plan, research, data-model, contracts/events.md, quickstart, tasks.md — 21 tasks). Implementation tier: Opus 4.8 per constitution V rubric — doctrine-adjacent behavior change (memory-formation/perception doctrine) + reducer/replay-contract surface spanning internal/sim executor+reducer and internal/mind absorb (research.md D7). Worktree .worktrees/task-159, branch task-159-first-person-harvest-memory (pushed).

Implemented by spec-implementer @ Opus 4.8 (code 8495b34, wiki+player docs 4f5c926) — reducer act-time removal (actor + awake in-radius witnesses, provenance-blind), salChop/salQuarry=4 first-person act memories, mind absorb parity (also fixed latent gap: quarry now arms its actor), 9 new tests, go test ./... green, tui-design no-op, player-docs 13 fresh, merge-drift pr exit 0. 50 wiki notes re-pinned (11 body-updated). T018 live validation (fresh world sc081, seed 42, 9 game days, reflex-only): 98 chops → 98 first-person memories (SC-004 exact); 0 self-corrections (SC-001, baseline 103/103); 0 awake on-scene corrections — 11 in-radius correctors were all asleep at act tick, the designed exception (SC-002); loss memories 28% of all (was 75%), every one a genuine absent/asleep return-discovery (SC-003). SC-006 journal spot-check deferred to next LLM-backed world (scratch run had no LLM → no journals); causal chain fully evidenced mechanically. Quarry act path covered by unit tests (reflex-only run mints no quarry intents).
<!-- SECTION:NOTES:END -->
