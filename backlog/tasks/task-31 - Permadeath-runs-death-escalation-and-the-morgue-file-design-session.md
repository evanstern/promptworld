---
id: TASK-31
title: 'Permadeath runs, death escalation, and the morgue file: design session'
status: In Progress
assignee: []
created_date: '2026-07-20 19:55'
updated_date: '2026-07-25 07:05'
labels:
  - design
  - learning-game
dependencies: []
ordinal: 19000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Roguelike survival design (user, 2026-07-20; see decision-3 strife doctrine). Per-agent permadeath already exists: Health 0 fires agent.died with a cause (starvation/exposure/collapse), the reducer marks the agent Dead forever, nothing respawns, and witnesses within radius 8 form memories (docs/wiki/executor.md). But nothing is at stake at the run level: a world where all 8 agents die just keeps ticking, and the only lethal threats are neglect — the gru wounds (-250 health) but floors health at 1 and never kills (docs/wiki/gru.md). Socratic/spec session covering: (1) Run outcomes — define the end of a run: all agents dead emits a run.ended event; decide what the daemon does after (keep ticking an empty world? go idle? mark the save dir closed?). A save directory is already one run with no reset command (docs/wiki/world-save-directory.md), which fits roguelike semantics: new run = new world dir, old runs are archives. (2) The morgue file — the roguelike epitaph, written into the save dir as a durable artifact: per-death and at run end, days survived, cause, notable memories, relationships, debts owed and owing, deeds; the chronicle narrates the moment, the morgue file is the legacy document a human reads afterward. (3) Graves — a grave overlay where an agent fell; candidates: mourning morale effects, agents visiting graves, grief entering rumors. (4) Death escalation — decide whether the gru can kill (remove the health-1 floor entirely, or only for wounded/NearDeath agents, or only in the cold season per TASK-28), and whether untreated wounds fester without sleep (with TASK-30 healing). (5) Candidate: a difficulty preset recorded in the world.json manifest so hard runs are reproducible and comparable by seed. Output: a spec under specs/ linked to the board via spec-bridge.

Spec: specs/044-run-outcomes-morgue
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A grounding/design session produces a spec directory for run outcomes, death escalation, and the morgue file, linked on the board via spec-bridge
- [x] #2 Spec phase: Setup
- [x] #3 Spec phase: Foundational (blocking prerequisites)
- [x] #4 Spec phase: User Story 1 — The run ends, and the story survives it (P1) 🎯 MVP
- [x] #5 Spec phase: User Story 2 — The morgue file (P2)
- [ ] #6 Spec phase: User Story 3 — The gru can finish the wounded (P3)
- [ ] #7 Spec phase: User Story 4 — Graves on the map, grief in the village (P4)
- [ ] #8 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Pre-session decisions (user, 2026-07-20): (1) Permadeath enabled per agent — already true in code; this task makes it consequential at run level. (2) Death should be a real risk from more than neglect: the gru or wounds should be able to finish someone. (3) Per decision-3, deaths should generate social material (grief, blame, inheritance of stockpiles), not just remove an agent.

Re-grounding 2026-07-22: gru wound mechanics hold (gruWound=250 gru.go:38; wound floored at 1 health, gru.go:124). Spec 012 changes no death rules. Spec 013 (TASK-51) now owns the physical death-drop mechanic (FR-006: carried bulk spills as a ground pile at the death site) — defer that piece to 013; keep morgue file, graves, run.ended, and death escalation scope here.

Drift audit 2026-07-23: verified exact — gruWound=250 gru.go:38, gruWoundFloor=1 gru.go:39, floor applied gru.go:124; 'wounds, never kills' doctrine comment gru.go:15-16. Spec 013 (death-drop FR-006) is now Done/merged — the deferral of the physical death-drop piece to 013 has happened; remaining scope here (morgue file, graves, run.ended, escalation) untouched.

spec-bridge sync: Setup: 0/1 · Foundational (blocking prerequisites): 0/2 · User Story 1 — The run ends, and the story survives it (P1) 🎯 MVP: 0/9 · User Story 2 — The morgue file (P2): 0/7 · User Story 3 — The gru can finish the wounded (P3): 0/3 · User Story 4 — Graves on the map, grief in the village (P4): 0/4 · Polish & Cross-Cutting: 0/3

Implementation dispatched 2026-07-25: Setup+Foundational+US1 (T001-T012) to spec-implementer on Opus 4.8 — rubric: concurrency/scheduling logic (sim loop halt semantics, executor emission ordering, reducer arms) + doctrine-adjacent (run-end永久halt, command gating). US3/US4 and US2 rendering earmarked Sonnet per plan.md tier notes. Branch task-31-run-outcomes-morgue in .worktrees/task-31.

spec-bridge sync: Setup: 1/1 · Foundational (blocking prerequisites): 2/2 · User Story 1 — The run ends, and the story survives it (P1) 🎯 MVP: 9/9 · User Story 2 — The morgue file (P2): 0/7 · User Story 3 — The gru can finish the wounded (P3): 0/3 · User Story 4 — Graves on the map, grief in the village (P4): 0/4 · Polish & Cross-Cutting: 0/3

US2 dispatched 2026-07-25: T013-T019 to spec-implementer on Opus 4.8 — rubric: doctrine-adjacent (charter fingerprint in the metatron turn pipeline, injection whitelist extension, narrator single-flight worker) + cross-package (world/metatron/sim/mind/scribe/tui). Same branch/worktree.

spec-bridge sync: Setup: 1/1 · Foundational (blocking prerequisites): 2/2 · User Story 1 — The run ends, and the story survives it (P1) 🎯 MVP: 9/9 · User Story 2 — The morgue file (P2): 7/7 · User Story 3 — The gru can finish the wounded (P3): 0/3 · User Story 4 — Graves on the map, grief in the village (P4): 0/4 · Polish & Cross-Cutting: 0/3

US2 gated + ratified 2026-07-25 (4 deviations, see spec commit 7cd7163). US3+US4+Polish dispatched: T020-T029 to spec-implementer on Sonnet — rubric: routine slices (single-mechanism gru conditional + tests; grave structure kind + TUI rendering + doc reconciliation); same branch/worktree, sequential to avoid worktree collision. Deed-vocabulary duplication (narrate.go/scribe morgue.go) recorded as drift surface for a future extract.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-07-25 04:42
---
Relabeled per learning-game synthesis (2026-07-25): learning-game PREREQUISITE — player-attributable failure is the honest grader. Absorbs the shareable-epitaph framing: the morgue file is the retellable artifact (Boatmurdered lesson — the celebrated DF story object is a retelling). Scenario runs (staged-session decision) will consume run.ended as their fail signal; the open 'what failure can the player cause and how is it attributed' operator question lands in this design session.
---

created: 2026-07-25 04:54
---
Design session opened 2026-07-25 in main session (Fable): permadeath/run-outcomes/morgue + the learning-game synthesis open question 1 (player-attributable failure state).
---

created: 2026-07-25 05:02
---
Design session round 1 decided (operator, 2026-07-25): (1) RUN END = archive — run.ended halts the sim loop; daemon keeps serving reads (chronicle/morgue/event log, TUI postmortem mode); new run = new world dir, old runs are browsable archives. (2) ESCALATION = gru kills the already-wounded — health-1 floor stays for healthy villagers; a hit on a wounded/NearDeath agent can kill, so lethality emerges from compounding preventable spirals, not one-hit randomness. (3) FAILURE/ATTRIBUTION = all-dead run.ended is THE hard failure everywhere (scenarios add rubric fails on top); the morgue aligns each death against the charter/orders revision timeline — evidence, never a blame score. Resolves learning-game synthesis open question 1. (4) DIFFICULTY = folded into TASK-68 stage / TASK-119 scenario presets; no standalone difficulty field (identities not sliders; scenario id + seed already make runs comparable). Round 2 (morgue mechanism, graves scope) next; spec to follow.
---

created: 2026-07-25 05:06
---
Design session round 2 decided (operator, 2026-07-25): (5) MORGUE MECHANISM = deterministic core + narrated epilogue — reducer writes the factual event-derived record per death and at run end (days, cause, relationships, debts, deeds, charter/orders timeline alignment), works LLM-off; chronicle narrator appends a prose epilogue when available; facts never depend on the model. (6) GRAVES v1 = marker + memory/rumor hooks riding existing systems (mental-map place-facts, social fabric grief rumors); mourning morale effects and grave-visiting behaviors deferred. Orchestrator judgment calls (flagged, unobjected): single accumulating morgue.md per world (per-death epitaphs + run-end summary); format designed export-ready, Boatmurdered HTML export stays a separate Wave-3 task. Session decisions complete — spec drafting next (AC #1).
---
<!-- COMMENTS:END -->
