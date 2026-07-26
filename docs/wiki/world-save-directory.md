---
name: world-save-directory
description: One directory = one world run — manifest (world.json), create/open validation, path helpers, clean separability, v1→v2→v3→v4→v5 migration; the manifest field-by-field catalog and the path-helper catalog live in two split-off children
kind: component
sources:
  - internal/world/world.go
  - internal/world/migrate.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# World save directory

`internal/world` defines the save-directory contract: one directory is one world run,
containing everything that run owns and nothing any other run touches. Copying a
stopped world's directory is a complete, restorable archive.

Two children (summary-style, corpus-spec v2) carry the field-by-field and
file-by-file detail:

- [[world-save-manifest-fields]] — every `Manifest` (`world.json`) field, in
  the order each spec added it: `format_version` history, `tick_game_seconds`,
  map dims, `terrain_gen`, the `meeting` block, and the additive
  `teaching`/`memory_relevance`/`stage`/`stage_overridden`/`charter_preset`/
  `scenario` fields, plus `Open`'s validation of each.
- [[world-save-path-helpers]] — the path-helper catalog: every well-known
  file a save directory can contain (`world.db`, `llm.json`,
  `calibration.json`, sockets/pidfile/log, `charter.md`, `metatron/`,
  `village_charter.md`, `morgue.md`, `bundles/`, `tuning.json`) and the
  runtime-only files swept between daemon runs.

## How it works

Since spec 063 ([[grounded-feedback]], T014), `world.go` also hosts the
curriculum ladder's SKIN-INDEPENDENT content — `StageLadderInfo{Concept,
Grants, UnlockEvidence}`, `StageOrder`, and `StagesLadder` — relocated here
from `cmd/promptworld/stages.go` (package `main`, which `internal/tui`
cannot import) so the TUI help overlay's D9 guardian section and
`promptworld stages` read the exact same table; `stages.go` keeps its own
`stageOrder`/`stagesLadder` names as plain aliases so its rendering code is
unchanged. This is data only, unrelated to the manifest fields (see [[world-save-manifest-fields]]).
`World.Map()` regenerates the terrain from the seed,
dimensions, and (spec 068) the manifest's `terrain_gen` — deterministic, so the map
is never stored ([[worldmap-generation]]).

- `Create(dir, name, seed)` refuses any existing non-empty directory, creates
  `agents/` (empty — flat files for later features live there), and writes the
  manifest — since spec 068 (C12) stamping `TerrainGen: worldmap.GenMarshSand` so
  every NEW world is born on the current terrain generation; only a migrated
  legacy world carries an absent `terrain_gen`. The genesis `world.created` event
  is appended by the CLI `new` command, not here.
- `Open(dir)` reads and validates the manifest: unknown `format_version`, a
  `tick_game_seconds` other than 1, a malformed `meeting` block (bad "HH:MM",
  or convene not strictly before open), or (spec 068) a `terrain_gen` outside
  `{absent/0, worldmap.GenMarshSand}` is a hard error, so an old binary can
  never half-load a newer world. Since TASK-147, the `format_version` mismatch
  case alone is a distinguishable typed error, `*world.ErrFormatVersionMismatch{Got,
  Want}` (every other `Open` failure — corrupt JSON, a bad `meeting` block, a
  directory that was never a world — stays a plain wrapped error): the manifest
  parsed fine, it's just a version this build doesn't support, so the world itself
  may be perfectly healthy. Callers that only need daemon reachability (not content —
  [[daemon-lifecycle]], [[cli-runtime-control]]'s `stop`/`status`) match it with
  `errors.As` to tell "can't read this world's content" apart from a genuine open
  failure.
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

Runtime files (`daemon.sock`, `daemon.pid`) exist only while a daemon runs and are
swept by [[daemon-lifecycle]] when stale. The full layout is documented in
`specs/001-world-daemon/contracts/storage.md`.

**Migration** (`migrate.go`, spec 012 US6 for v1→v2 + spec 013 for v2→v3 +
spec 041 for v3→v4 + spec 068 for v4→v5 —
[[world-migration]] has the full design): `OpenForMigration(dir)` loads a world
manifest without the current version gate — it admits `format_version` 1 through 4
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
manifest). The v4→v5 step (spec 068, [[world-migration]]) is different in kind: it
is manifest-only — no state transform, no snapshot cut, no archive — since a
carried-forward v4 world's event log, state, and terrain (`terrain_gen` stays
absent, so it keeps legacy generation) do not change at all; `Migrate` detects a
v4 source, bumps `Manifest.FormatVersion` to 5, and returns
`MigrateResult{ManifestOnly: true}`.

## Connections

[[daemon-lifecycle]] opens the world and cross-checks the manifest against store meta;
[[event-log]] and [[snapshots]] live inside `world.db`; [[ipc-server]] binds the socket
at `SockPath()`. [[cli-promptworld]]'s `new` creates worlds and `migrate` upgrades
an older one ([[world-migration]]). [[world-save-manifest-fields]] and
[[world-save-path-helpers]] carry the field-by-field and file-by-file detail
this note summarizes — their own Connections sections link the subsystem
each field/path belongs to (worldmap-generation, mental-maps,
curriculum-ladder, scenario-machinery, grounded-feedback, world-tuning,
llm-orchestrator, cognition, guardian, governance, morgue, bundle-tools).

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
