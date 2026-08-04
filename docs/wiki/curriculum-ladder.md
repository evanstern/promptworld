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
size_budget_exempt: spec 116 added inspect_pack to the stage-1 ceiling paragraph, which must stay on this note as the normative ladder authority (the ceiling roster and its rationale are the note's whole subject); a per-stage summary-style split is a dedicated future pass, not this task's scope
verified_against: 5761edb18e2b5fb49c6a03a050b0d871f5546c05
---

# Curriculum ladder

The curriculum ladder (spec 046, TASK-68) turns promptworld into a staged
teaching game: a world is created AT a stage, which sets what the player
can do in it (which guardian tools exist, whether instruction files bind),
and playing well — passing seeded exercises — earns the next stage for the
player's FUTURE worlds. One prompt-engineering concept per stage, an identity
rather than a difficulty label, and every earned claim auditable
to events in the proving world's own log.

## How it works

**The four stages.** The substrate knows only neutral ids —
`stage-1`..`stage-4`, the `world.Stage1`..`Stage4` constants ([[world-save-directory]]);
what the player SEES is skin data: [[skin]] (spec 052, TASK-121 — its own
runtime substrate; ids/semantics never moved, the lookup grew a skin
dimension) maps them to the default Guardian skin's client-approved
identities — **The Voice** ("you speak, it acts"), **The Written Word**
("your law outlives the conversation"), **The Craft** ("you shape what it
can do"), **The Stewardship** ("a world in your care") — via
`skin.Stage(id)`/`StageName(id)`; an alternate skin may re-voice all four
names, ids/unlock semantics staying put. `internal/world/world.go`'s
`StagesLadder` (relocated from `cmd/promptworld/stages.go` by spec 063 T014
— [[grounded-feedback]]'s TUI help overlay reads the same table and
`internal/tui` cannot import package `main`; `stages.go` keeps
`stageOrder`/`stagesLadder` as plain aliases onto
`world.StageOrder`/`world.StagesLadder`, rendering unchanged) carries the
skin-INDEPENDENT ladder content beside those identities: per stage, the
concept taught (conversational prompting → instruction authoring →
capability design → mastery), what the world grants and what evidence
unlocks the next. Absent stage (`""`) is a pre-ladder world: ungated,
stage-4 semantics, so existing worlds lose nothing.

**Stage is a birth fact.** `Manifest.Stage`/`StageOverridden`/`CharterPreset`
are additive `omitempty` `world.json` fields, closed-vocabulary-validated at
`world.Open` and stamped exactly once by `promptworld new` via
`world.SetStage` right after `Create` — immutable for the world's lifetime,
no toggle command exists or will ([[world-save-directory]]). `new` resolves stage from the per-user unlocks record: stage-1 for a new
player, else the highest earned; an explicit unearned `--stage` gets an
informed refusal naming the skipped concepts (skin names) unless
`--override`, which proceeds, recording `stage_overridden: true` — flagging
the run honestly. `promptworld stages` is the ladder's front door: all four
stages always visible (identity, concept, grants, unlock evidence, earned
state with the proving world), never a difficulty menu. The earned rule
(stage-1's floor, else a record entry) is `worlds.(*Unlocks).StageEarned`
(spec 078, relocated from this file's former `stageEarned` — the T014 move
repeated for earned state); `cmdStages` and the TUI ladder
([[grounded-feedback]]) call it.

**The stage ceiling** (`internal/guardian/charter.go`): `stageCeiling`
returns a stage's ceiling as a `bundle.GrantDoc` — the narrowing shape
persona bundle grants use — and `applyStageCeiling` intersects it into the
world-level grant at every manifest load site (turn + status), right after
`loadManifest` and BEFORE `grantedRoster`, so declaration, prose, and door
inherit it rather than re-implement it ([[guardian]]). Stage-1 and stage-2
pin the roster to `stage1CeilingTools` — `send_omen`, `send_vision`,
`monitor_and_act`, `cancel_order` (the ratified TASK-119 amendment put
standing orders in the stage-1 grant — the first-night exercise teaches the
watch), plus, since spec 063, `explain` (read-only, zero-cost, the tutor
guide's grounding tool — [[grounded-feedback]]), plus, since spec 084, the
five charge-free plan-layer tools (`place_designation`/`cancel_designation`/
`issue_directive`/`cancel_directive`/`survey_site` — the plan loop is a
teaching primitive, the monitor_and_act precedent; [[guardian-designations]]),
plus, since spec 085, `prophesy` (send_vision's stage profile — the same
influence verb with a wager; [[guardian-faith]]), plus, since spec 101,
`brief_myths` (a read-only lookup, the `survey_site`/`explain` profile —
`canonize_region` stays OUT, the `work_miracle` precedent —
[[guardian-canonization]]), plus, since spec 116, `inspect_pack` (the same
read profile: charge-free, event-free, never the turn's act, and it widens no
acting capability — it only lets the guardian see what a villager already
carries; its companion `take_item` stays OUT, a `work_miracle` kind —
[[guardian-miracles]]), with no miracle kinds and the empty-intersection
effect shutting out bundle tools; stage-3, stage-4, and pre-ladder worlds
have no ceiling. `guardian.StageCeilingVerbs(stage)`
exports a stage's granted loop-tool names in registry order — the SAME
intersection applied here — for the TUI help overlay's D9 guardian section
to teach from, not a second hand-maintained list. A player's
`capabilities.json` may narrow WITHIN the ceiling, never exceed it. The
daemon hands the manifest's stage + preset to the guardian boot-frozen via
`mt.SetStage` (the `SetBundles` discipline), so the ceiling can't be
tampered mid-run ([[daemon-lifecycle]]).

**The stage-1 instruction lock + presets.** `stageCharter` forks `loadCharter`
by stage: at stage-1 the effective charter IS the world's preset constant —
`persona.DefaultCharter` for `""`/`"default"`, `persona.TutorCharter` (the
orientation preset `promptworld new` seeds by default at stage-1, opt-out
via `--charter-preset default`) for `"tutor"` — sourced from the compiled-in
text, never the file, so the lock is tamper-proof, not advisory. An edited `charter.md` earns an honest notice naming the unlocking stage
(never silent ignoring); a missing file is restored to the preset without
notice — what binds never changed. `stageSkills` does the same for skill
files, binding only from stage-3 — present-but-unbound files get one
notice. Every other stage behaves byte-identically to pre-ladder.
`observeCharter`'s `default` flag is preset-aware: it compares the
effective text against the WORLD's preset constant, so a stage-1 tutor-preset
world's `guardian.charter_observed` records `default: true` — authored by the
game, never the player — so preset text never opens the stage-2→3 gate
below, and the [[morgue]]'s charter-evidence timeline stays honest.

**Earning the next stage** — split into [[curriculum-ladder-progression]]:
two executor-emitted events latch `State.StagesUnlocked` once per (world,
stage) under gate conjuncts requiring real, never freehand-asserted,
evidence; a per-user convenience record mirrors earned stages (never read
by world behavior); nine seeded exercises (spec 077) supply the rubrics the
scenario machinery evaluates and emits against. See the child for the full
event/evidence/exercise detail.

## Connections

[[guardian]] applies the stage ceiling and the instruction lock in its
turn/status assembly and emits the `guardian.charter_observed` events whose
`default` flag the progression child's gate derivation inverts; [[morgue]]
aligns deaths against the same observation timeline. [[world-save-directory]]
holds the `stage`/`stage_overridden`/`charter_preset` manifest facts;
[[daemon-lifecycle]] hands the boot-frozen `SetStage` handoff;
[[cli-promptworld]] fronts `promptworld stages` and `new --stage`; [[skin]]
supplies the four stages' player-visible identities (`Stage`/`StageName`)
the ladder facts here pair with; [[grounded-feedback]] (spec 063) relocated
`StagesLadder`/`StageOrder` here from `cmd/promptworld`, added `explain` to
`stage1CeilingTools`, and reads the ceiling via `StageCeilingVerbs` for the
D9 guardian section, whose spec-078 ladder block reads `StageEarned` above;
[[curriculum-ladder-progression]] is the split-off child — how a stage is
earned; [[testing-strategy]] catalogs the per-layer suites.

## Operational notes

The ladder gates the GUARDIAN's capabilities, never the villagers' world:
`TestCrossStageDeterminism` pins the same seed ticking identically at every
stage. See [[curriculum-ladder-progression]] for earned-progression
notes (the unlocks record, exercise evaluator status).
