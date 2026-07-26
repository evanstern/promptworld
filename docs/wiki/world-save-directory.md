---
name: world-save-directory
description: One directory = one world run — manifest (world.json), path helpers, create/open validation, clean separability, v1→v2→v3→v4 migration
kind: component
sources:
  - internal/world/world.go
  - internal/world/migrate.go
verified_against: 6318cf8b53e407765f0c9793f5355a7af4777ed7
---

# World save directory

`internal/world` defines the save-directory contract: one directory is one world run,
containing everything that run owns and nothing any other run touches. Copying a
stopped world's directory is a complete, restorable archive.

## How it works

`Manifest` (serialized as `world.json` at the dir root) carries `name`, `seed`
(uint64), `created_at` (RFC3339, metadata only — wall time never enters sim state),
`format_version` (currently **4** — spec 041's per-agent mental-map break
([[mental-maps]]) bumped it from 3: private spatial knowledge gates target
resolution, so a v3 world loaded with no seeded maps would leave every
villager knowing nothing (research D7) — on top of spec 013's inventory/storage
break that bumped it from 2, on top of spec 012's resources/food/crafting break
that bumped it from 1; a v1, v2, or v3 manifest is refused by `Open` with a
pointer to `promptworld migrate <world>` — [[world-migration]]), `tick_game_seconds` (fixed 1),
`map_width`/`map_height` (default 64×64; zero/absent values from older saves default
on `Open`), and an optional `meeting` block (TASK-36, `MeetingConfig`:
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
`World.Map()` regenerates the terrain from the seed and
dimensions — deterministic, so the map is never stored ([[worldmap-generation]]).

- `Create(dir, name, seed)` refuses any existing non-empty directory, creates
  `agents/` (empty — flat files for later features live there), and writes the
  manifest. The genesis `world.created` event is appended by the CLI `new` command,
  not here.
- `Open(dir)` reads and validates the manifest: unknown `format_version`, a
  `tick_game_seconds` other than 1, or a malformed `meeting` block (bad "HH:MM",
  or convene not strictly before open) is a hard error, so an old binary can
  never half-load a newer world.
- `SetTeaching(dir, on)` (spec 039) is the offline read-modify-write for the
  `Teaching` marker: `Open`s the manifest, flips the field, rewrites
  `world.json`. A running daemon reads `Teaching` only at boot, so this is a
  config edit the next start picks up, never a live toggle; `promptworld new
  --teaching` calls it right after `Create` (keeping `Create`'s own signature
  untouched for its other callers), and `promptworld teaching <world> on|off`
  is its standalone door ([[cli-promptworld]]).
- `SetStage(dir, stage, overridden, charterPreset)` (spec 046,
  [[curriculum-ladder]]) is `SetTeaching`'s write-mechanics sibling with the
  opposite lifecycle: exactly ONE caller — `promptworld new`, once,
  immediately after `Create` — because stage is a write-once birth fact (no
  `promptworld stage <world> …` toggle exists or ever will). It does not
  re-validate its arguments (the `SetTeaching` contract): callers pass
  already-`ValidStage`/`ValidCharterPreset`-checked values.
- Path helpers centralize layout: `DBPath()` → `world.db`, `LLMConfigPath()` →
  `llm.json` (the [[llm-orchestrator]] config, written by `new`, deletable to
  disable inference), `CalibrationPath()` → `calibration.json` (the
  seconds-per-point profile written only by `promptworld calibrate` —
  [[cognition]]; an absent file is legal, pessimistic bootstrap defaults apply),
  `EstimatorStatePath()` → `estimator_state.json` (the daemon-written snapshot
  of live latency estimates, TASK-113 — absent is legal, boot then seeds from
  calibration/bootstrap alone; never event-sourced, never read during replay),
  `SockPath()` → `daemon.sock`, `PidPath()` → `daemon.pid`,
  `LogPath()` → `daemon.log`, `CharterPath()` → `charter.md` (the player-editable
  prompt), `GuardianDir()` → `metatron/` (dir name frozen, spec 052 ruling 2 — the Guardian's soul and transcript —
  [[guardian]]), and `VillageCharterPath()` → `village_charter.md` (the village's
  scribe-rendered law, deliberately distinct from the Guardian's charter —
  [[governance]], TASK-13), `MorguePath()` → `morgue.md` (spec 044: the run's
  accumulating legacy document — one factual epitaph per death plus a run-end
  summary, scribe-rendered; a regenerable view over the event history, never a
  source of truth, exactly like the chronicle and village charter —
  [[morgue]]), and `BundlesDir()` → `bundles/` (spec 036: the
  drop-in persona/tool bundle root, discovered and boot-frozen by
  [[bundle-tools]]; absent means no bundles, never an error), and
  `TuningPath()` → `tuning.json` (spec 048: the optional, operator-
  authored, sparse per-world tuning manifest promoting doctrine constants
  to per-world dials — absent means every dial keeps its doctrine-constant
  default, exactly today's behavior; never written by `promptworld new`;
  [[world-tuning]] has the full mechanism).

Runtime files (`daemon.sock`, `daemon.pid`) exist only while a daemon runs and are
swept by [[daemon-lifecycle]] when stale. The full layout is documented in
`specs/001-world-daemon/contracts/storage.md`.

**Migration** (`migrate.go`, spec 012 US6 for v1→v2 + spec 013 for v2→v3 +
spec 041 for v3→v4 —
[[world-migration]] has the full design): `OpenForMigration(dir)` loads a world
manifest without the current version gate — it admits `format_version` 1, 2, or 3
(the sole purpose is migrating an older world this build otherwise can't `Open`) and
refuses an already-current world outright; `Migrate(dir)` runs the whole ceremony —
refuse a live daemon or an already-migrated source (the guard is keyed to the
*source* format: `V1DBPath()` → `world.v1.db` for a v1 source, `V2DBPath()` →
`world.v2.db` for a v2 source, `V3DBPath()` → `world.v3.db` for a v3 source),
read the source world's covering snapshot,
transform it (`internal/sim` — an older source chains every remaining transform
in one run, e.g. 1→2→3→4), archive the live `world.db` (and any `-wal`/`-shm` sidecars) to that
source-format archive **before** writing anything new (the archive is never
overwritten and never deleted), write a fresh log (`world.created` then
`world.migrated`) plus its covering snapshot, then bump `Manifest.FormatVersion` to
the current version last — a crash between the archive and the manifest bump
leaves a recoverable state (restore = rename the archive back, reset the
manifest).

## Connections

[[daemon-lifecycle]] opens the world and cross-checks the manifest against store meta;
[[event-log]] and [[snapshots]] live inside `world.db`; [[ipc-server]] binds the socket
at `SockPath()`. [[cli-promptworld]]'s `new` creates worlds and `migrate` upgrades
an older one ([[world-migration]]). [[mental-maps]] is the spec-041 subsystem the
current format version exists to support. [[curriculum-ladder]] is the spec-046
subsystem behind the `stage`/`stage_overridden`/`charter_preset` manifest
fields — the daemon reads them at boot and hands them boot-frozen to
[[guardian]] for the stage ceiling and the stage-1 instruction lock; the
per-user unlocks record that gates `promptworld new`'s default stage lives
outside the save directory (in the worlds home), advisory and never an
authority over anything in this directory. [[scenario-machinery]] is the
spec-054 subsystem that validates and consumes the `scenario` block this
note covers — the daemon reads it at boot the same boot-frozen way as
`stage`. [[world-tuning]] is the spec-048
subsystem `TuningPath()` fronts — a peer of `llm.json`/`calibration.json`,
consumed by the daemon's boot seed, never validated by this package.

## Operational notes

Seed and format version are immutable after creation (except across a migration,
which bumps `format_version` in place). There is deliberately no global
registry of worlds — the directory is the identity, per the grounding decision "never
global; runs cleanly separable" ([[design-grounding]]). Archiving = stop the daemon,
`cp -R` the directory. A migrated world's directory additionally carries the
source-format archive(s) (`world.v1.db` and/or `world.v2.db` and/or `world.v3.db`,
depending how far it
chained), the untouched original database(s) — deleting one is a deliberate,
irreversible acceptance of that step of the migration; `Migrate` never removes
either itself.
