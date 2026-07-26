---
name: curriculum-ladder-progression
description: How the curriculum ladder is actually earned — the two curriculum.* events and their EvaluateUnlock gate conjuncts (per stage transition), the per-user ~/.promptworld/unlocks.json convenience record, and the nine-exercise ExerciseDefinition catalog (spec 077) whose rubrics the scenario machinery evaluates and EMITS for at every stage.
kind: component
sources:
  - internal/sim/curriculum.go
  - internal/daemon/curriculum.go
  - internal/worlds/unlocks.go
verified_against: 93837e1885bff17114df75e5382ac60dee24776a
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
`metatron.order_expired` pattern; the scenario rubric machinery
(`scenario.go`'s `scenarioRubricEvents`, generalized by spec 077 to
per-exercise dawn boundaries) is the production emitter for EVERY cataloged
exercise — see [[scenario-machinery]]; spec 072's the-law emission deferral
(FR-009) is complete.
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
stage-3→4, any `Custom` evidence entry — since spec 077 this is
`SkillsObservedEvidence`, `Custom: true` BY CONSTRUCTION (the long-open
"which tool" design slot, filled): `metatron.skills_observed` records the
bound skill-file set a guardian turn ran under, and skill files bind only
from stage-3 and only players author them, so a recorded observation is a
player-granted capability by structural necessity. Two further sanctioned
constructors joined `CharterObservedEvidence`/`OrderPlacedEvidence`:
`CharterEvidenceFromState` (reads the `State.CharterObservedSeq/Tick`
coordinates the charter arm persists since spec 077, `Custom` from the
persisted `CharterCustom`) and `SkillsObservedEvidence` (reads
`SkillsObservedSeq/Tick`) — both omit honestly when no coordinates are on
state (a pre-077 snapshot; the pass waits, self-healing on the next
observation). Stage-4 is graduation: nothing unlocks past it.

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
behavior ever reads it; its only consumers are `promptworld stages` and
`new`'s earned-stage check.

**The exercises** (`sim.ExerciseDefinition`) are CONTENT, not machinery —
stage, deterministic seed, framing, an event-derived rubric whose every term
must be a cataloged event type, the pass-signal shape, a chronicle
score-narrative framing (failure is a story, not a scold), and — since spec
054 — an optional authored `Schedule` of incidents plus an
`IncidentVisibility` override ([[scenario-machinery]]), and — since spec
077 — a `BoundaryDay` declaration (dawn of day N, or 0 = rolling from day
2). NINE ship (`ScenarioExercises`, spec 077 FR-001 — 3/2/2/2 by stage,
seeds 46101–46109): stage-1 **first-night**, **cold-dawn**,
**stranger-at-the-gate**; stage-2 **the-law** (rolling; emission real —
its pass carries `CharterEvidenceFromState` evidence and unlocks stage-3
without `--override`) and **blighted-larder**; stage-3 **toolsmith** (the
skills-evidence exercise — its pass first opens the stage-3→4 gate in
production) and **fog-watch**; stage-4 **long-winter** and
**stewards-charge** (graduation — passes record, nothing unlocks). Every
exercise has a production evaluator arm and real pass emission
([[scenario-machinery]] for the full table). The scenario/rubric machinery
consumes the catalog end to end; `Manifest.Scenario` is its consumed
schema seam, no longer reserved ([[world-save-directory]]).
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
events. Since spec 077 every cataloged exercise both evaluates AND emits
for real (`promptworld new --scenario <id>` for any of the nine ids) — the
stage-2→3 unlock no longer needs `--override` (the-law's pass carries the
charter evidence), and the stage-3→4 gate grants for the first time from
production machinery (toolsmith/fog-watch's `Custom: true` skills
evidence) — see [[scenario-machinery]].
