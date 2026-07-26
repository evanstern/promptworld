---
name: curriculum-ladder
description: The spec-046 four-stage curriculum ladder's identity — immutable per-world birth-fact stage, skin-supplied stage identities, the guardian's stage ceiling, and the stage-1 instruction lock/presets. How progression is earned (unlock events, the per-user record, exercise content) split to [[curriculum-ladder-progression]].
kind: component
sources:
  - internal/skin/skin.go
  - internal/world/world.go
  - internal/guardian/charter.go
  - cmd/promptworld/stages.go
  - internal/worlds/unlocks.go
verified_against: 93837e1885bff17114df75e5382ac60dee24776a
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
proving world), an informed identity table, never a difficulty menu. The
earned rule (stage-1's floor, else a record entry) is now
`worlds.(*Unlocks).StageEarned` (spec 078, relocated from this file's
former `stageEarned` — StagesLadder's T014 move repeated for earned
state); `cmdStages` and the TUI ladder ([[grounded-feedback]]) both call it.

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

**Earning the next stage** — split into
[[curriculum-ladder-progression]]: two executor-emitted events,
`curriculum.exercise_passed` and `curriculum.stage_unlocked`, latch
`State.StagesUnlocked` once per (world, stage) under `sim.EvaluateUnlock`'s
per-transition gate conjuncts (stage-2→3 and stage-3→4 require
`Custom`-derived charter/tool evidence, never freehand-asserted); a per-user
`~/.promptworld/unlocks.json` convenience record mirrors earned stages
without ever being read by world behavior; and `sim.ExerciseDefinition`
content — nine exercises since spec 077, 3/2/2/2 by stage — supplies the
seeded rubrics the scenario machinery evaluates and EMITS for; the stage-3
gate's `Custom` tool evidence is spec 077's `metatron.skills_observed`
observation (skill files are player-authored and stage-3+ by construction).
See the child for the event/evidence/exercise detail.

## Connections

[[guardian]] applies the stage ceiling and the instruction lock inside its
turn/status assembly and emits the `metatron.charter_observed` events whose
`default` flag the progression child's gate derivation inverts; [[morgue]]
aligns deaths against that same observation timeline.
[[world-save-directory]] holds the `stage`/`stage_overridden`/`charter_preset`
manifest facts; [[daemon-lifecycle]] hands the boot-frozen `SetStage`
handoff; [[cli-promptworld]] fronts `promptworld stages` and `new --stage`;
[[skin]] supplies the four stages' player-visible identities
(`Stage`/`StageName`) this note's ladder facts pair with;
[[grounded-feedback]] (spec 063) relocated `StagesLadder`/`StageOrder` here
from `cmd/promptworld`, added `explain` to `stage1CeilingTools`, and reads
the stage ceiling via `StageCeilingVerbs` for the D9 guardian section,
whose spec-078 ladder block also reads `StageEarned` above;
[[curriculum-ladder-progression]] is the split-off child covering how a
stage is actually earned; [[testing-strategy]] catalogs the per-layer
suites.

## Operational notes

The ladder gates the GUARDIAN's capabilities, never the villagers' world:
`TestCrossStageDeterminism` pins that the same seed ticks identically at
every stage. See [[curriculum-ladder-progression]] for earned-progression
operational notes (the unlocks record, exercise evaluator status).
