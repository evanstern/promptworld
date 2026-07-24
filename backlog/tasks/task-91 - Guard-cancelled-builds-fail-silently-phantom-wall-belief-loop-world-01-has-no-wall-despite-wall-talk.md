---
id: TASK-91
title: >-
  Guard-cancelled builds fail silently: phantom-wall belief loop (world-01 has
  no wall despite wall talk)
status: In Progress
assignee: []
created_date: '2026-07-24 17:50'
updated_date: '2026-07-24 18:02'
labels:
  - bug
dependencies: []
priority: high
ordinal: 1
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
world-01's villagers have coordinated around a defensive wall for ~25k ticks, but zero walls exist: 10 agent.built events total (8 fires, 1 oven, 1 chest), no wall_plank/wall_stone.

Root causes (from world-01 event log, ~/.promptworld/worlds/world-01/world.db):

1. **Silent mid-work cancellation, no falsifiable feedback.** Hazel's build_wall_stone (intent_set seq 217117, tick 439939) landed and started the 600-tick work cycle. At tick 440287 Rowan — sent by the planner to socialize with the wall crew — pathed onto the wall's res tile (32,26). The per-tick re-validation guard (internal/sim/executor.go:657, `buildSite(...) && !agentAt(s, in.ResX, in.ResY)`) failed and the build resolved as a bare agent.intent_done (seq 217362): no wall, no spend, no intent_rejected, no failure memory. Meanwhile the conversation machinery had already written gist memories ("The trio works together to brace a structure…") and the chronicle wrote "built fortified walls" — so agents *believe* the wall exists. Cedar's later planner reasons say "The walls stand" (tick 434284) and generate check/repair/reinforce intents against a phantom wall (repair rejected: "no damaged wall reachable"). Nothing ever corrects the belief.

2. **Sociality is adversarial to adjacent-builds.** Wall work attracts "go join/help them" planner intents; helpers path to the builder and step through the res tile, tripping the never-entomb occupancy guard. Building a wall near other villagers — exactly the social activity walls are — is nearly impossible.

Secondary: the only other real attempt (Cedar, tick 459994) was gate-rejected for lacking 2 refined stone; the material pipeline rarely completes because stone is spent elsewhere. This is behavior, not a bug per se, but the silent-cancel loop above prevents retry pressure from ever building.

Spec: specs/038-loud-build-failure
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A build cancelled by mid-work re-validation (occupancy/site-vanished) emits a distinguishable failure event (not a bare agent.intent_done) — event type + payload documented in docs/wiki/event-types.md
- [ ] #2 The builder receives a situated failure memory (origin action) stating the build did NOT complete and why, so the belief is falsifiable
- [ ] #3 Transient occupancy no longer permanently kills a build: pathing avoids active build res tiles OR the build tolerates/waits out a passerby on the res tile (design choice recorded on this task)
- [ ] #4 Regression test: helper agent paths onto the res tile mid-build; assert failure event + failure memory (or successful wall under the chosen tolerance design), never a silent bare intent_done
- [ ] #5 Spec phase: Setup
- [ ] #6 Spec phase: Foundational (Blocking Prerequisites)
- [ ] #7 Spec phase: User Story 1 — Loud, distinguishable build failure (Priority: P1) 🎯 MVP
- [ ] #8 Spec phase: User Story 2 — Builder remembers the failure (Priority: P1)
- [ ] #9 Spec phase: User Story 3 — Passerby no longer kills the build (Priority: P2)
- [ ] #10 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Starting per constitution: full Spec Kit (specify → clarify → plan → tasks) before implementation; code exploration underway to ground the spec.

AC#3 design choice: BUILD TOLERATES PASSERBY (not pathing avoidance). Rationale from code exploration: pathing is unweighted BFS (internal/sim/path.go) with no reservation index — soft-avoid would require a new tile-cost concept plus a Res-tile registry consulted inside passable() (terrain.go:38), which every movement step calls; large blast radius. Tolerance is surgical: split the guard at executor.go:657 — buildSite false = genuine failure (new failure event + situated failure memory); agentAt true = transient occupancy → build waits (completion deferred, never entombs), with a grace timeout so a permanent squatter becomes a loud failure instead of an infinite wait. Failure event + memory must come from the executor (mind never writes memories from intent_done; mind.go:218 just re-arms the planner).

Implementation tier: Opus 4.8 (escalated from Sonnet default). Rubric lines: (1) cross-package change — internal/sim executor/reducer + internal/mind absorb re-arm + internal/tui digest; (2) doctrine-adjacent behavior change — alters belief/memory semantics (failure memories, falsifiability; adjacent to 030-epistemic-hygiene); plus deterministic replay byte-identity risk in the event-sourced hot path. Delegated to spec-implementer with model: opus on specs/038-loud-build-failure.
<!-- SECTION:NOTES:END -->
