---
id: TASK-135
title: 'Curriculum ladder production wiring: the unlock gate never runs outside tests'
status: To Do
assignee: []
created_date: '2026-07-25 19:29'
labels:
  - curriculum
  - review-2026-07-25
  - learning-game
dependencies: []
priority: medium
ordinal: 105000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Team review 2026-07-25 finding: the spec-046 curriculum ladder's evaluation layer is entirely unwired. Verified by grep — production callers of ZERO for all of:
  - EvaluateUnlock (internal/sim/curriculum.go:201) — the stage-2 -> stage-3 gate conjunct
  - nextLadderStage (curriculum.go:153)
  - CharterObservedEvidence (curriculum.go:242)
  - ScenarioExercises (curriculum.go:332) — the shipped exercise catalog

Their only callers are internal/sim/curriculum_test.go plus one TUI digest test. That is 332 lines of internal/sim held alive entirely by its own assertions, presenting on the board as shipped.

Consequence (Principle III): TASK-68 is Done with AC#6 ('pass signals are event-derived and surface in-game') ticked, and the tick's own note reads 'fixture-proven; production emission = TASK-119'. TASK-119 is scenario/incident machinery (spec 054) — it schedules emissions; it does not wire the unlock evaluator. So no task currently owns making the ladder actually evaluate in a running world. TASK-127 (takeover surfaces) CONSUMES curriculum.stage_unlocked; something must emit it.

Operator decision (2026-07-25): card the production wiring — do NOT delete the 332 lines.

Scope: call EvaluateUnlock on the real event path so a running world can actually cross stage-2 -> stage-3; emit curriculum.stage_unlocked from that evaluation; make ScenarioExercises reachable from a production path. Reconcile with TASK-119's scheduled-emission primitive and TASK-127's ceremony consumer so there is exactly one emitter.

Non-trivial: full Spec Kit before implementation.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 EvaluateUnlock is called on the production event path, not only from tests
- [ ] #2 A running world can cross stage-2 -> stage-3 and emits curriculum.stage_unlocked from that evaluation
- [ ] #3 ScenarioExercises is reachable from a production path (or explicitly re-scoped with the reason recorded)
- [ ] #4 Exactly one emitter of curriculum.stage_unlocked exists; TASK-127's ceremony consumes it and TASK-119's scheduler does not duplicate it
- [ ] #5 TASK-68's AC#6 fixture caveat is retired, or TASK-68 carries an explicit note that it closed on a fixture and this task completed it
<!-- AC:END -->
