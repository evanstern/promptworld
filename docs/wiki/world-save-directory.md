---
name: world-save-directory
description: One directory = one world run — manifest (world.json), path helpers, create/open validation, clean separability, v1→v2→v3→v4 migration
kind: component
sources:
  - internal/world/world.go
  - internal/world/migrate.go
verified_against: 3b7dd17b478ab5aa64e4c99c44b77bc565d71376
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
([[daemon-lifecycle]], [[cognition]]). `World.Map()` regenerates the terrain from the seed and
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
- Path helpers centralize layout: `DBPath()` → `world.db`, `LLMConfigPath()` →
  `llm.json` (the [[llm-orchestrator]] config, written by `new`, deletable to
  disable inference), `CalibrationPath()` → `calibration.json` (the
  seconds-per-point profile written only by `promptworld calibrate` —
  [[cognition]]; an absent file is legal, pessimistic bootstrap defaults apply),
  `SockPath()` → `daemon.sock`, `PidPath()` → `daemon.pid`,
  `LogPath()` → `daemon.log`, `CharterPath()` → `charter.md` (the player-editable
  prompt), `MetatronDir()` → `metatron/` (the angel's soul and transcript —
  [[metatron]]), and `VillageCharterPath()` → `village_charter.md` (the village's
  scribe-rendered law, deliberately distinct from Metatron's charter —
  [[governance]], TASK-13), and `BundlesDir()` → `bundles/` (spec 036: the
  drop-in persona/tool bundle root, discovered and boot-frozen by
  [[bundle-tools]]; absent means no bundles, never an error).

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
current format version exists to support.

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
