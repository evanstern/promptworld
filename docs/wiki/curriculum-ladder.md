---
name: curriculum-ladder
description: The spec-046 four-stage curriculum ladder — staged worlds, earned capabilities: stage ceiling, stage-1 instruction lock, curriculum.* unlock events, per-user unlocks record, seeded exercises
kind: component
sources:
  - internal/sim/curriculum.go
  - internal/daemon/curriculum.go
  - internal/worlds/unlocks.go
  - internal/skin/skin.go
  - internal/world/world.go
  - internal/guardian/charter.go
  - cmd/promptworld/stages.go
verified_against: 31c893e0406653197e467a89b2fdb96f0bcf2ee0
---

# Curriculum ladder

The curriculum ladder (spec 046, TASK-68) turns promptworld into a staged
teaching game: a world is created AT a stage, the stage sets what the player
can do in it (which guardian tools exist, whether instruction files bind),
and playing well — passing seeded exercises — earns the next stage for the
player's FUTURE worlds. One prompt-engineering concept per stage, an identity
per stage rather than a difficulty label, and every earned claim auditable
back to events in the proving world's own log.

## How it works

**The four stages.** The substrate knows only neutral ids — `stage-1` ..
`stage-4`, the `world.Stage1`..`Stage4` constants ([[world-save-directory]]);
what the player SEES is skin data: [[skin]] (spec 052, TASK-121 — landed as
its own runtime substrate; ids and semantics never moved, the lookup just
grew a skin dimension) maps them to the default Guardian skin's
client-approved identities — **The Voice** ("you speak, it acts"), **The
Written Word** ("your law outlives the conversation"), **The Craft** ("you
shape what it can do"), **The Stewardship** ("a world in your care") —
via `skin.Stage(id)`/`StageName(id)`, an alternate skin free to re-voice all
four names while the ids and unlock semantics stay put.
`internal/world/world.go`'s `StagesLadder` (relocated here from
`cmd/promptworld/stages.go` by spec 063 T014 — [[grounded-feedback]]'s TUI
help overlay reads the same table, and `internal/tui` cannot import package
`main`; `stages.go` keeps `stageOrder`/`stagesLadder` as plain aliases onto
`world.StageOrder`/`world.StagesLadder` so its own rendering code is
unchanged) carries the skin-INDEPENDENT
ladder content beside those identities: per stage, the concept taught
(conversational prompting → instruction authoring → capability design →
mastery), what the world grants, and what evidence unlocks the next. Absent
stage (`""`) is a pre-ladder world: ungated, stage-4 semantics, so every
existing world loses nothing.

**Stage is a birth fact.** `Manifest.Stage`/`StageOverridden`/`CharterPreset`
are additive `omitempty` `world.json` fields, closed-vocabulary-validated at
`world.Open` and stamped exactly once by `promptworld new` via
`world.SetStage` right after `Create` — immutable for the world's lifetime,
no toggle command exists or will ([[world-save-directory]]). `new` resolves
the stage from the per-user unlocks record: default stage-1 for a new player,
else the highest earned stage; an explicit unearned `--stage` refuses with an
informed message naming the skipped concepts (skin names) unless
`--override`, which proceeds and records `stage_overridden: true` — the
honesty marker that keeps overridden runs comparable as overridden runs.
`promptworld stages` is the ladder's front door: all four stages always
visible (identity, concept, grants, unlock evidence, earned state with the
proving world), an informed identity table, never a difficulty menu.

**The stage ceiling** (`internal/guardian/charter.go`): `stageCeiling` returns
a stage's capability ceiling as a `bundle.GrantDoc` — the same narrowing shape
a persona bundle's grant uses — and `applyStageCeiling` intersects it into the
world-level grant at every manifest load site (turn + status), immediately
after `loadManifest` and BEFORE `grantedRoster`, so declaration, prose, and
door all inherit it rather than re-implementing it ([[guardian]]). Stage-1 and
stage-2 pin the roster to `stage1CeilingTools` — `send_omen`, `send_vision`,
`monitor_and_act`, `cancel_order` (the ratified TASK-119 amendment put
standing orders in the stage-1 grant because the first-night exercise teaches
the watch), plus, since spec 063, `explain` (read-only, zero-cost, the
tutor guide's own grounding tool — [[grounded-feedback]]), with no miracle
kinds and the empty-intersection effect shutting
out bundle tools; stage-3, stage-4, and pre-ladder worlds have no ceiling.
`guardian.StageCeilingVerbs(stage)` exports a stage's granted loop-tool
names in registry order — the SAME intersection applied here — for the TUI
help overlay's D9 guardian section to teach from without a second
hand-maintained list. A
player's `capabilities.json` may narrow WITHIN the ceiling, never exceed it.
The daemon hands the manifest's stage + preset to the guardian boot-frozen via
`mt.SetStage` (the `SetBundles` discipline), so the ceiling cannot be
tampered mid-run ([[daemon-lifecycle]]).

**The stage-1 instruction lock + presets.** `stageCharter` forks `loadCharter`
by stage: at stage-1 the effective charter IS the world's preset constant —
`persona.DefaultCharter` for `""`/`"default"`, `persona.TutorCharter` (the
stage-1 orientation preset `promptworld new` seeds by default at stage-1,
opt-out via `--charter-preset default`) for `"tutor"` — sourced from the
compiled-in text, never the file, so the lock is tamper-proof rather than
advisory. An edited `charter.md` earns an honest notice naming the unlocking
stage (never silent ignoring); a missing file is restored to the preset with
no notice, since what binds never changed. `stageSkills` does the same for
skill files, which bind only from stage-3 — present-but-unbound files get one
notice. Every other stage behaves byte-identically to the pre-ladder loaders.
Crucially, `observeCharter`'s `default` flag is preset-aware: it compares the
effective text against the WORLD's preset constant, so a stage-1 tutor-preset
world's `metatron.charter_observed` records `default: true` — authored by the
game, never the player — which is exactly what keeps preset text from ever
opening the stage-2→3 gate below (and keeps the [[morgue]]'s charter-evidence
timeline honest).

**The unlock chain** ([[event-types]], [[sim-state-reducer]]). Two event
types, `curriculum.exercise_passed` (`ExercisePassedPayload{exercise, stage,
tick, evidence?}`, each `EvidenceRef{type, seq, tick, custom?}` a re-locatable
pointer into this world's log) and `curriculum.stage_unlocked`
(`StageUnlockedPayload{stage, exercise, tick}`), are the executor emission
class — pure functions of (state, tick), no whitelist entries, the
`metatron.order_expired` pattern; the spec-054 scenario rubric machinery
(`scenario.go`'s `scenarioRubricEvents`, TASK-119) is the production emitter
— see [[scenario-machinery]] — and test fixtures remain the only in-tree
emitter for exercises whose rubric evaluator is still unbuilt content work.
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
behavior ever reads it; its only consumers are `promptworld stages` and
`new`'s earned-stage check.

**The exercises** (`sim.ExerciseDefinition`) are CONTENT, not machinery —
stage, deterministic seed, framing, an event-derived rubric whose every term
must be a cataloged event type, the pass-signal shape, a chronicle
score-narrative framing (failure is a story, not a scold), and — since spec
054 — an optional authored `Schedule` of incidents plus an
`IncidentVisibility` override ([[scenario-machinery]]). Two ship
(`ScenarioExercises`): **first-night** (stage-1, seed 46101 — keep the
village alive through night one by directing the guardian: visions, omens,
and the watch; the only exercise with a production rubric evaluator today,
plus an authored night-one `gru_emerges` incident) and **the-law** (stage-2,
seed 46102 — get a norm adopted while a player-authored charter revision is
in force, the SC-004 conjunct; its rubric evaluator is still unbuilt content
work — state retains only the charter fingerprint, not the `Default` flag
the conjunct needs). The spec-054 scenario/rubric machinery consumes the
catalog end to end ([[scenario-machinery]]); `Manifest.Scenario` is its
consumed schema seam, no longer reserved ([[world-save-directory]]).
Surfacing: `ipc.WorldStatus.Stage`/`StageOverridden` ride status, `promptworld
status` renders a skin-named stage line (plus, on a scenario world, an
`exercise: <id> — <outcome>` line, [[scenario-machinery]]), and the TUI
digest narrates both `curriculum.*` types under the guardian grammar family
(the FROZEN `metatron` namespace's family-namespace mapping, [[tui-client]]).

## Connections

[[guardian]] applies the stage ceiling and the instruction lock inside its
turn/status assembly and emits the `metatron.charter_observed` events whose
`default` flag the gate derivation inverts; [[morgue]] aligns deaths against
that same observation timeline. [[sim-state-reducer]] owns the two
`curriculum.*` reducer arms and the `CurriculumPasses`/`StagesUnlocked`
state; [[event-types]] catalogs the payload shapes. [[world-save-directory]]
holds the `stage`/`stage_overridden`/`charter_preset` manifest facts and the
consumed `scenario` block; [[daemon-lifecycle]] wires the always-on unlock
observer and the boot-frozen `SetStage` handoff; [[cli-promptworld]] fronts
`promptworld stages` and `new --stage`/`new --scenario`; [[skin]] supplies the four stages'
player-visible identities (`Stage`/`StageName`) this note's ladder facts
pair with; [[scenario-machinery]] is the spec-054 production emitter for this
note's two event types; [[grounded-feedback]] (spec 063) relocated
`StagesLadder`/`StageOrder` here from `cmd/promptworld`, added `explain` to
`stage1CeilingTools`, and reads the stage ceiling via `StageCeilingVerbs` for
the TUI help overlay's D9 guardian section; [[testing-strategy]] catalogs the
per-layer suites (reducer, guardian stage gating, daemon observer, unlocks
record, CLI).

## Operational notes

The ladder gates the GUARDIAN's capabilities, never the villagers' world:
`TestCrossStageDeterminism` pins that the same seed ticks identically at
every stage. Deleting `~/.promptworld/unlocks.json` forgets earned
convenience, not truth — any proving world's log still carries its
`curriculum.stage_unlocked` events. Since spec 054 (TASK-119), the
`first-night` exercise's production rubric evaluator lands
`curriculum.exercise_passed`/`stage_unlocked` on a real scenario world
(`promptworld new --scenario first-night`) — see [[scenario-machinery]];
`the-law`'s evaluator remains unbuilt content work, so a stage-2 world still
needs `--override` until it lands.
