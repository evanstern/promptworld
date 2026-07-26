---
name: event-types-curriculum-events
description: Curriculum-ladder event rows split from [[event-types]]: curriculum.exercise_passed, curriculum.stage_unlocked. Load when tracing spec 046's staged-world unlock gates or the spec 054 scenario-rubric production emitter.
kind: concept
sources:
  - internal/sim/curriculum.go
  - internal/daemon/curriculum.go
verified_against: 4c66d240b2715706964f02cfd2396256c9957d8e
---

# Event types — curriculum-ladder events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.

Spec 046 (the curriculum ladder — staged worlds, earned capabilities,
[[curriculum-ladder]]) adds no
format bump: `State` gains `omitempty` `CurriculumPasses []CurriculumPass`
(ring-capped at 32) and `StagesUnlocked []string`, so a pre-046 snapshot with
both fields absent round-trips byte-identically. TWO new event types —
`curriculum.exercise_passed` and `curriculum.stage_unlocked` — are the
executor emission class (no whitelist entries, the `metatron.order_expired`
pattern): a stage's tool ceiling is a manifest-load-time intersection, never
an event, but crossing a ladder gate is (`sim.EvaluateUnlock` decides the
gate conjuncts; full shapes and reducer effects in the table below). The
`world.json` Manifest itself gains three additive fields outside this
event-sourced state — `stage`, `stage_overridden`, `charter_preset`
([[world-save-directory]], [[guardian]]) — validated at `world.Open` against
a closed vocabulary exactly like `memory_relevance` above.

Spec 054 (scenario incident-schedule machinery — [[scenario-machinery]],
TASK-119) adds no new event type and lands spec 046's still-awaited
production emitter: `GuardianOrder` gains `omitempty` `PlacedSeq` (the
placement event's store seq, reducer-stamped from the event envelope at
apply time, ignored on the incoming payload like `Status` — a pre-054 order
round-trips byte-identically), and `curriculum.exercise_passed`/
`stage_unlocked` (spec 046, below) now emit for real on a scenario world via
`scenario.go`'s `scenarioRubricEvents`, the same executor-emission-class
shapes test fixtures previously proved alone. `gru.emerged` also gains a
second emission path — an authored incident preempting that night's random
roll — with no payload change (the emitted shape is identical either way).
`ExerciseDefinition` (content, not event-sourced state) gains `Schedule`/
`IncidentVisibility`, carrying no event of their own.

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `curriculum.exercise_passed` (spec 046, [[curriculum-ladder]]) | `ExercisePassedPayload{exercise, stage, tick, evidence?}` in `internal/sim/curriculum.go`; `EvidenceRef{type, seq, tick, custom?}` | executor emission class (the `metatron.order_expired` pattern — pure function of state + tick, no whitelist entry); the scenario rubric machinery (`scenario.go`'s `scenarioRubricEvents`, [[scenario-machinery]]) is the production emitter for EVERY cataloged exercise since spec 077 (per-exercise dawn boundaries via `ExerciseDefinition.BoundaryDay`; evidence through the four sanctioned constructors — `OrderPlacedEvidence`, `CharterObservedEvidence`, and spec 077's state-sourced `CharterEvidenceFromState`/`SkillsObservedEvidence`) | appends a bounded `CurriculumPass` record (ring-capped at 32, `State.CurriculumPasses`) — the auditable proof a seeded exercise's event-derived rubric reached its pass signal |
| `curriculum.stage_unlocked` (spec 046, [[curriculum-ladder]]) | `StageUnlockedPayload{stage, exercise, tick}` in `internal/sim/curriculum.go` | same emission class as `exercise_passed`, derived from a pass whose evidence satisfies its stage's gate conjuncts (`sim.EvaluateUnlock` — stage-1: any pass; stage-2: a `metatron.charter_observed` evidence entry whose `custom` flag is the derived inverse of `CharterObservedPayload.Default` (`sim.CharterObservedEvidence` — spec 044 US2's fingerprint event, so a default/preset charter never opens the gate); stage-3: any `custom` evidence entry — in production, spec 077's `SkillsObservedEvidence` over the recorded `metatron.skills_observed`, custom by construction) | latches the stage into `State.StagesUnlocked` (once per world per stage — a duplicate is rejected; `stage-1`, the unearned floor, is never an unlockable stage); the daemon's curriculum observer (`internal/daemon/curriculum.go`, always-on, wired before the LLM gate) upserts the per-user `~/.promptworld/unlocks.json` record on observing it; both types render in the TUI digest under the metatron grammar family ([[tui-client]]), and the world's stage rides status as `ipc.WorldStatus.Stage`/`StageOverridden` (composed straight from the manifest, [[world-save-directory]]) |

## Connections

[[curriculum-ladder]] owns the spec 046 `curriculum.*` family end to
end — payloads and reducer arms in `internal/sim/curriculum.go` (executor
emission class, no whitelist entries; [[scenario-machinery]]'s rubric
machinery is the production emitter for all nine cataloged exercises since
spec 077), the
daemon's always-on unlock observer in `internal/daemon/curriculum.go`, and
the per-user unlocks record it projects.
[[event-types-scenario-incidents]] catalogs the spec-077 incident family
whose pressure these passes are earned under.
