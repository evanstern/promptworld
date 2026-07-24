---
id: TASK-96
title: >-
  Per-agent mental maps: knowledge-gated spatial memory replaces omniscient
  nearest-target resolution
status: In Progress
assignee: []
created_date: '2026-07-24 19:03'
updated_date: '2026-07-24 22:23'
labels:
  - epistemics
  - spatial-memory
  - design-session
dependencies: []
references:
  - research/Agent-Mental-Maps/Agent-Mental-Maps.md
  - research/Agent-Mental-Maps/_grounding.md
  - internal/sim/policy.go
  - internal/sim/path.go
  - internal/mind/prompt.go
  - internal/sim/memory.go
priority: medium
ordinal: 80000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
<!-- SECTION:DESCRIPTION:BEGIN -->
THE ISSUE (traced 2026-07-24): agents do not perceive the world — they name a verb and the sim omnisciently resolves it. (1) Every goal resolver BFS-scans the ENTIRE map for the nearest matching tile/structure: resolveGoal/goalResolvers in internal/sim/policy.go:160-488 via nearest/nearestAdjacentTo in internal/sim/path.go:75-101 — no fog of war, no discovery requirement. (2) The planner prompt lists ALL village structures with exact global coordinates whether or not the agent ever saw them (internal/mind/prompt.go:95-104; also capped at first 6 in slice order, so newer structures silently vanish). (3) talk_to resolves any living agents current position anywhere on the map (policy.go:207-216). Locality exists only in passive channels (witnessRadius 8, encounter radius 1, nearby-agent listing 10) and in memory flavor: situated memories carry Where{X,Y,Desc} (internal/sim/memory.go:65-181) and beliefs carry witnessed/told/inferred provenance — but NONE of that knowledge is load-bearing for navigation. Net effect: "find X" is never a search problem; exploration, scouting, being lost, stale knowledge, and spatial rumor value are all impossible by construction.

THE DESIRED RESULT: each agent holds a private, tree-organized mental map of the 64x64 grid with an explicit UNKNOWN state; target resolution consults the agents knowledge instead of ground truth. An agent can only path to a fire it has witnessed, been told about, or can see; unknown space is explorable (frontier-style directed wander); remembered content can be stale and get corrected on arrival; spatial knowledge spreads socially through talk. The structure must extend cleanly to the planned layered grids (multi-level worlds): the quadtree-over-belief-grid family lifts to octrees/per-layer grids joined by a portal (stairs) topology with the same cell semantics. Prompt rendering and any read-only map-query tools follow from the spec.

RESEARCH (grounded corpus, built 2026-07-24): research/Agent-Mental-Maps/ vault branch — 10 interlinked notes over a 56-source grounding pass (_grounding.md). Key established facts: occupancy belief grids with log-odds occupied/free/unknown cells are the standard robotics model and the explicit unknown state is the load-bearing feature; quadtrees add hierarchy in 2D and the SAME payload lifts unchanged to octrees in 3D (OctoMap); games implement knowledge gating as three-state per-player visibility grids (unexplored / explored-but-stale / visible); staleness fails asymmetrically (agents act confidently on invalidated facts) and is handled by decay TOWARD UNCERTAINTY plus provenance-conditioned updates; frontier-based exploration turns unknown-space search into directed movement; the storage hierarchy doubles as the pathfinding abstraction (quadtree-gate A*, HPA* clusters); at 64x64 memory is trivial (4 KiB/byte-layer/agent) so representation choice is about semantics and 3D-extensibility, not size. Open questions flagged on the MOC for the spec to resolve: gate target-resolution only vs also pathfinding; spatial gossip on talk and its trust level; LLM consumption (rendered text vs query tools); portal-connected layers vs true voxel volume; decay half-lives at game speed.

CODEBASE AREAS OF RELEVANCE: internal/sim/policy.go (goalResolvers, decideIntent reflex fallback — both use global-nearest helpers); internal/sim/path.go (bfs, nextStep, nearest, nearestAdjacentTo); internal/sim/executor.go (movement stepping 277-285, witness stamping, arrival seam); internal/mind/prompt.go (structure list 95-104, nearby agents 106-122, memory window); internal/sim/memory.go (situated memories, placeScanRadius — natural write-path substrate); internal/sim/consolidate.go + internal/mind/consolidate.go (beliefs, confidence half-life — decay machinery already exists); internal/sim/agents.go (Agent state, cadence constants); worldmap generation (docs/wiki/worldmap-generation.md).

RELATED TASKS: TASK-76 (entity-lookup accessor seam — the world-side index this feature would query through; determinism-neutral seam should land first or together); TASK-80 (perception of absence / grounded arrival observations — that channel is effectively the WRITE path of a mental map and shares the describePlace substrate; this task is the READ path); TASK-79 (belief reinforcement seam — staleness correction feeds it); TASK-95 (loud failure for unresolvable goals — failure semantics change when resolution is knowledge-gated: "no fire I know of" vs "no fire exists").

SEQUENCING AND RIGOR: needs full Spec Kit (specify -> clarify -> plan -> tasks) linked via spec-bridge before implementation — this is doctrine-adjacent (changes what agents can know), cross-package (sim + mind), and concurrency/determinism-sensitive, so implementation tier is Opus per the constitution rubric. Replay determinism is a hard constraint: per-agent maps become reducer state and must replay bit-identically.
<!-- SECTION:DESCRIPTION:END -->

Spec: specs/041-agent-mental-maps
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Spec Kit spec produced, clarified, and linked to this task via spec-bridge before any implementation (constitution rigor)
- [ ] #2 The five open design questions from the research MOC (gating scope, spatial gossip, LLM consumption, 3D posture, decay rates) are resolved with recorded decisions in the spec
- [ ] #3 Per-agent spatial knowledge exists with an explicit unknown state; verb target resolution consults it instead of the global world scan (no omniscient nearest for knowledge-gated goals)
- [ ] #4 Planner prompt renders only agent-known structures/places (no global structure coordinates; first-6 truncation bug retired)
- [ ] #5 Unknown space is explorable: an agent with no known target can search deliberately rather than only random-wander
- [ ] #6 Replay determinism preserved: mental-map state is reducer-owned and the determinism harness proves bit-identical replay
- [ ] #7 Chosen representation has a documented 3D/layered-grid extension path in the spec
- [x] #8 Spec phase: Setup
- [x] #9 Spec phase: Foundational (blocking all stories)
- [x] #10 Spec phase: User Story 1 — Agents act only on places they know (P1) 🎯 MVP
- [x] #11 Spec phase: User Story 2 — Prompt renders only known places (P2)
- [x] #12 Spec phase: User Story 3 — Stale memories corrected by reality (P3)
- [x] #13 Spec phase: User Story 4 — Deliberate search of the unknown (P4)
- [x] #14 Spec phase: User Story 5 — Spatial knowledge spreads through talk (P5)
- [ ] #15 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Research grounding is durable: vault branch research/Agent-Mental-Maps/ (56-source _grounding.md + 10 notes, branch gate green) committed as c70c53f on task-96-agent-mental-maps (worktree .worktrees/task-96), cut from origin/main dee6c01. Task activated 2026-07-24. Next step per AC #1: Spec Kit specify -> clarify, then spec-bridge:link back to this task before any implementation.

Specify phase complete: specs/041-agent-mental-maps/spec.md + requirements checklist (all items pass) committed to main as 1bd5db5 and pushed. Task branch rebased onto it (research commit now 4101500); worktree feature.json points at specs/041-agent-mental-maps (f404233). Five starred defaults in spec Assumptions await /speckit-clarify (AC #2); spec-bridge:link after tasks phase (AC #1). [note recovered by the TASK-77 session: it had been appended to the contested task-96 file during the ID collision and was autostashed during the refile to TASK-100; original edit preserved in git stash]

US1 complete (T011-T016, worktree f55180a): omniscient resolution replaced for 14 verbs + reflex parity; full suite green except pre-existing TASK-100 red. Accepted deviations recorded in data-model.md addenda (PeerSighting, availability split, spec-040 fixture repair 9bff237). MVP checkpoint reached.

US2+US3 complete (T017-T022, worktree b446a21): known-places prompt (first-6 cap retired, provenance phrasing, frontier line, peers from sightings) + agent.map_corrected loop with chronicle/digest/wiki gates. Full suite green except pre-existing TASK-100; one unrelated load-flake noted (metatron TestDigestFailureCarries, passes isolated). Deviations recorded in data-model addenda.

US4+US5 complete (T023-T031): search verb (frontier BFS, exhaustion honesty, reflex get-food fallback) + social.place_told sidecar (≤2 facts/direction, told provenance, teller Seen). Race gate re-verified by orchestrator (agent watcher died post-suite-green). Remaining: polish T032-T040.
<!-- SECTION:NOTES:END -->
