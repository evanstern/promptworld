---
name: curriculum-ladder-progression
description: How the curriculum ladder is actually earned — the two curriculum.* events and their EvaluateUnlock gate conjuncts (per stage transition), the per-user ~/.promptworld/unlocks.json convenience record, and ExerciseDefinition content (first-night/the-law) whose rubric the spec-054 scenario machinery evaluates.
kind: component
sources:
  - internal/sim/curriculum.go
  - internal/daemon/curriculum.go
  - internal/worlds/unlocks.go
verified_against: ad2a6543a9caf51d1cd28af863291f3daa3bd4eb
---

# Curriculum ladder — progression (unlocks, record, exercises)

Split off [[curriculum-ladder]] (spec 071): how a player actually EARNS the
next stage — the event-sourced unlock chain, the convenience record that
remembers it, and the exercise content the chain is evaluated against — as
distinct from the parent's stage identities, birth-fact immutability, and
guardian-facing ceiling/lock.

## How it works

**The unlock chain** ([[event-types]], [[sim-state-reducer]]). Two event
types, `curriculum.exercise_passed` (`ExercisePassedPayload{exercise, stage,
tick, evidence?}`, each `EvidenceRef{type, seq, tick, custom?}` a re-locatable
pointer into this world's log) and `curriculum.stage_unlocked`
(`StageUnlockedPayload{stage, exercise, tick}`), are the executor emission
class — pure functions of (state, tick), no whitelist entries, the
`metatron.order_expired` pattern; the spec-054 scenario rubric machinery
(`scenario.go`'s `scenarioRubricEvents`, TASK-119) is the production emitter
— see [[scenario-machinery]] — and test fixtures remain the only in-tree
emitter for exercises without pass emission (`the-law` since spec 072 has a
production EVALUATOR but still no emitter — FR-009 content work).
The reducer appends passes to the bounded `State.CurriculumPasses` ring
(`curriculumPassRetain` 32) and latches unlocks into `State.StagesUnlocked`
once per (world, stage), rejecting duplicates and `stage-1` (the unearned
floor). `sim.EvaluateUnlock(state, pass)` decides the gate conjuncts at
emission time: stage-1→2, any stage-1 pass; stage-2→3, the pass's evidence
must include a `metatron.charter_observed` entry with `Custom == true` —
where `Custom` is NEVER asserted freehand but derived by
`sim.CharterObservedEvidence`, the single sanctioned constructor, as the
inverse of the recorded `CharterObservedPayload.Default` (spec 044 US2), so a
default/preset charter structurally cannot satisfy the gate (SC-004) —
stage-3→4, any `Custom` evidence entry (a player-granted tool's contributing
act; which tool is TASK-119's exercise design). Stage-4 is graduation:
nothing unlocks past it.

**The per-user unlocks record** (`internal/worlds/unlocks.go`):
`~/.promptworld/unlocks.json` (`<worlds.Root()>/unlocks.json`), a
`{"unlocks": {"stage-2": {world, path, exercise, evidence, earned_at}}}` map
written by the daemon's always-on observer (`internal/daemon/curriculum.go`,
wired onto the notify fan-out before the LLM gate — a no-model world still
records its unlocks, [[daemon-lifecycle]]) whenever it sees
`curriculum.stage_unlocked`, with a pointer to the same batch's pass event as
evidence. The doctrine is the worlds-registry's, plus one explicit softening:
load-tolerant (missing/corrupt → empty, never an error; malformed entries
dropped at load but entries whose world path no longer exists KEPT — an
archived world is still historical proof), atomic `.tmp`+rename writes, and
advisory-never-authority — an unresolvable home directory WARNS and degrades
(the endpoint-lease precedent), because the record is pure convenience
projected from world event histories, which remain the authority. No world
BEHAVIOR ever reads it; its consumers are `promptworld stages`, `new`'s
earned-stage check, and — since spec 078 (TASK-152) — the TUI's own `?`
guardian-section forward-ladder block, which loads it once at client boot
(`Model.unlocks`, the `populateHelpLessons`/`LoadLessonsSeen` precedent) and
unions it with the live `replica.StagesUnlocked` at render time rather than
re-reading disk every frame ([[grounded-feedback]]). The earned predicate
itself — stage-1's unconditional floor, or an entry in this record —
is `(*Unlocks).StageEarned`, single-sourced here so `promptworld stages`
and the TUI ladder can never disagree on what counts as earned (relocated
from `cmd/promptworld/stages.go`'s former package-main `stageEarned`).

**The exercises** (`sim.ExerciseDefinition`) are CONTENT, not machinery —
stage, deterministic seed, framing, an event-derived rubric whose every term
must be a cataloged event type, the pass-signal shape, a chronicle
score-narrative framing (failure is a story, not a scold), and — since spec
054 — an optional authored `Schedule` of incidents plus an
`IncidentVisibility` override ([[scenario-machinery]]). Two ship
(`ScenarioExercises`): **first-night** (stage-1, seed 46101 — keep the
village alive through night one by directing the guardian: visions, omens,
and the watch; a production rubric evaluator AND pass emission, plus an
authored night-one `gru_emerges` incident) and **the-law** (stage-2, seed
46102 — get a norm adopted while a player-authored charter revision is in
force, the SC-004 conjunct; since spec 072 it has a production rubric
evaluator too — `theLawRubric` over `State.Norms` and the persisted
`State.CharterCustom` authorship flag, [[scenario-machinery]] — though pass
emission remains content work). The spec-054 scenario/rubric machinery consumes the
catalog end to end ([[scenario-machinery]]); `Manifest.Scenario` is its
consumed schema seam, no longer reserved ([[world-save-directory]]).
Surfacing: `ipc.WorldStatus.Stage`/`StageOverridden` ride status, `promptworld
status` renders a skin-named stage line (plus, on a scenario world, an
`exercise: <id> — <outcome>` line, [[scenario-machinery]]), and the TUI
digest narrates both `curriculum.*` types under the guardian grammar family
(the FROZEN `metatron` namespace's family-namespace mapping, [[tui-client]]).

## Connections

[[curriculum-ladder]] is the parent note — the stage identities, birth-fact
immutability, and guardian ceiling/instruction-lock this chain's stages
gate. [[sim-state-reducer]] owns the two `curriculum.*` reducer arms and the
`CurriculumPasses`/`StagesUnlocked` state; [[event-types]] catalogs the
payload shapes; [[guardian]] emits the `metatron.charter_observed` events
whose `default` flag the gate derivation inverts; [[daemon-lifecycle]] wires
the always-on unlock observer; [[scenario-machinery]] is the spec-054
production emitter for this note's two event types and the consumer of the
exercise catalog; [[world-save-directory]] holds the consumed `scenario`
manifest block; [[testing-strategy]] catalogs the per-layer suites (reducer,
daemon observer, unlocks record).

## Operational notes

Deleting `~/.promptworld/unlocks.json` forgets earned convenience, not
truth — any proving world's log still carries its `curriculum.stage_unlocked`
events. Since spec 054 (TASK-119), the `first-night` exercise's production
rubric evaluator lands `curriculum.exercise_passed`/`stage_unlocked` on a
real scenario world (`promptworld new --scenario first-night`) — see
[[scenario-machinery]]; `the-law` evaluates for real since spec 072 (live
gauges and card surfaces), but its pass EMISSION remains unbuilt content
work, so a stage-3 unlock still needs `--override` until it lands.
