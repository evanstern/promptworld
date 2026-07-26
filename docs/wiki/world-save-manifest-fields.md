---
name: world-save-manifest-fields
description: world.json's field-by-field catalog — format_version history, tick_game_seconds, map dims, terrain_gen, the meeting block, teaching/memory_relevance/stage/stage_overridden/charter_preset/scenario/lineage additive fields, and their Open validation rules
kind: component
sources:
  - internal/world/world.go
verified_against: 0fd2104c59c54be8e8071d319fa4ce192083faf3
---

# World save manifest: field catalog

Split from [[world-save-directory]] (summary-style, corpus-spec v2): every
field `Manifest` (`world.json`) carries, in the order each spec added it —
what it means, its default/absent behavior, and the validation `Open`
applies to it.

`Manifest` (serialized as `world.json` at the dir root) carries `name`, `seed`
(uint64), `created_at` (RFC3339, metadata only — wall time never enters sim state),
`format_version` (currently **5** — spec 068's terrain-vocabulary break
([[worldmap-generation]], [[tile-registry]]) bumped it from 4: new worlds generate
marsh/sand terrain gated by the manifest's `terrain_gen` field, and software
predating that field would silently IGNORE it and regenerate different terrain
under the same manifest — agents and structures standing on water (FR-007); a
pre-068 build refuses a v5 world at `Open` with the migrate hint instead of
mis-generating (C10) — on top of spec 041's per-agent mental-map break
([[mental-maps]]) that bumped it from 3 to 4: private spatial knowledge gates target
resolution, so a v3 world loaded with no seeded maps would leave every
villager knowing nothing (research D7) — on top of spec 013's inventory/storage
break that bumped it from 2, on top of spec 012's resources/food/crafting break
that bumped it from 1; a v1-v4 manifest is refused by `Open` with a
pointer to `promptworld migrate <world>` — [[world-migration]]), `tick_game_seconds` (fixed 1),
`map_width`/`map_height` (default 64×64; zero/absent values from older saves default
on `Open`), an optional `terrain_gen` int (spec 068, `json:"terrain_gen,omitempty"`:
absent/0 = legacy terrain generation, bit-identical to every pre-068 world;
`worldmap.GenMarshSand` (2) = the marsh/sand shoreline pass — what `Create` writes for
every new world; `Open` refuses any other value outright, the same fail-closed
posture as `memory_relevance`, so a future generation version this build doesn't
implement is never silently mis-generated), and an optional `meeting` block (TASK-36, `MeetingConfig`:
`convene`/`open` as "HH:MM" 24-hour game clock times, optional `x`/`y` meeting
place) — the per-world meeting convention the daemon seeds on boot
([[governance]], [[daemon-lifecycle]]); `promptworld new` never writes it, so
emergent is the default. It also carries an optional `teaching` bool
(`Manifest.Teaching`, `omitempty`, decision-6/spec 039): absent means
non-teaching (a non-teaching `world.json` round-trips byte-identically, no
`FormatVersion` bump — an additive defaulting bool old readers ignore); when
true, the daemon defaults the world's speed to the highest planner-safe
ladder rung at every boot and surfaces the horizon arithmetic on override
([[daemon-lifecycle]], [[cognition]]). It also carries an optional
`memory_relevance` string (`Manifest.MemoryRelevance`, `omitempty`, spec 042,
[[memory-retrieval]]): absent (`""`) keeps today's salience+recency memory
window; `"shadow"` additionally computes the relevance-augmented window and
records rank divergence while prompts still see the legacy window; `"on"`
lets the augmented window feed prompts (divergence still recorded). `Open`
refuses any other value outright — a typo must never silently run as off —
and the field is additive `omitempty`, so a pre-042 `world.json` round-trips
byte-identically with no `FormatVersion` bump. Spec 046
([[curriculum-ladder]]) adds three more additive `omitempty` fields on the
same closed-vocabulary/no-bump pattern: `stage` (`Manifest.Stage`,
`stage-1`..`stage-4` via the `Stage1`..`Stage4` constants — the world's
curriculum-ladder stage, set once at creation and IMMUTABLE for the world's
lifetime: no mutation command exists or will, deliberately unlike
`SetTeaching`'s live toggle; absent (`""`) means a pre-ladder world, ungated
with stage-4 semantics, so existing worlds lose nothing; `Open` validates via
`ValidStage`), `stage_overridden` (`bool` — the honesty marker `promptworld
new --stage <id> --override` stamps when a world is created at an unearned
stage, making overridden runs comparable as overridden runs), and
`charter_preset` (`""`/`"default"` = the authored default charter, `"tutor"`
= the stage-1 orientation preset — the constant that seeds `charter.md` at
genesis and, at stage-1 where instruction files are locked, IS the effective
charter regardless of edits ([[guardian]]); `Open` validates via
`ValidCharterPreset`). A fourth addition, the optional `scenario` block (`ScenarioConfig{exercise}`,
naming a `sim.ExerciseDefinition.ID`), was RESERVED on the `meeting`-block
precedent through spec 046; spec 054 ([[scenario-machinery]]) consumes it:
`Open` now validates a present block against `ValidScenarioExercise` — a
LOCAL mirror of `sim.ScenarioExercises`' id set (the `validLadderStage`
twin-list precedent, in reverse: the deterministic core does not import this
save-directory package and this package does not import the core, so each
side keeps its own closed vocabulary; `TestScenarioVocabularyMirrorsSimCatalog`
pins the two in sync) — refusing an unknown exercise id with `corrupt
world.json: scenario exercise %q unknown` rather than silently booting
ambient. `SetScenario(dir, exercise)` (`SetStage`'s write-mechanics sibling
— exactly ONE caller, `promptworld new --scenario`, once, immediately after
`Create`) is the write-once stamp; it does not re-validate its argument
(callers pass an already-`ValidScenarioExercise`-checked id). The daemon
arms the boot-frozen scenario runtime from this block at every boot
(`sim.State.ArmScenario`, [[daemon-lifecycle]]) — the incident schedule,
rubric evaluator, status facts, and exercise tab all key off it; a world
with no `scenario` block stays byte-identical to pre-054 on every path.
Spec 076 ([[world-forking]]) adds the optional `lineage` block
(`LineageConfig{parent, parent_created_at, fork_tick}`, additive
`omitempty`, no `FormatVersion` bump — the `teaching` precedent): fork
provenance, written exactly once by `world.Fork` and never mutated — the
fast offline mirror of the authoritative `world.forked` event in the fork's
own log (compare's default window reads it). `Open` applies a structural
check only (a present block must carry a non-empty `parent` and a
`fork_tick >= 0`, else the standard corrupt-manifest error) — deliberately
not a closed vocabulary, unlike `terrain_gen`/`memory_relevance`. A world
that was never forked carries no `lineage` key and round-trips
byte-identically.

## Connections

Back to [[world-save-directory]] for `Create`/`Open`/path-helper mechanics
and its sibling child [[world-save-path-helpers]]. [[worldmap-generation]]
is what `terrain_gen` (current format version 5) exists to support;
[[mental-maps]] is the spec-041 subsystem that bumped it to 4 one break
earlier; [[curriculum-ladder]] owns `stage`/`stage_overridden`/
`charter_preset`; [[scenario-machinery]] validates and consumes `scenario`;
[[governance]] and [[daemon-lifecycle]] read the `meeting` block at boot;
[[memory-retrieval]] owns `memory_relevance`; [[world-migration]] is the
bridge a pre-current `format_version` walks through; [[world-forking]] owns
the `lineage` block and why the fork's `seed` is carried, never fresh.
