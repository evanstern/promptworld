---
name: world-migration
description: The snapshot-cut migration chain (v1→v2→v3→v4→v5) that carries a stopped world's people, land, and knowledge across a save-format break as one world.migrated event; per-step design pins and the transform algorithms live in two split-off children
kind: component
sources:
  - internal/sim/migrate.go
  - internal/world/migrate.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# World migration

promptworld has broken its save format four times. Spec 012 (resources/food/crafting
v2) widened `Inventory` from the legacy `{wood, food}` pair to the full v2 resource
set and gave terrain generation rock outcrops — a v1 world's bytes simply don't mean
the same thing under a v2 build. Spec 013 (inventory/storage v3) added a bulk cap,
ground piles, chests, theft, and rot, which change how the reducer/executor treat
*existing* event shapes (yield truncation, death spill, the give-guard) — a v2 log
replayed under v3 code would diverge even though v2 and v3 land are identical.
Spec 041 ([[mental-maps]], v4) gates target resolution on a per-agent mental map
that a pre-041 world simply has none of — a v3 log loaded all-nil would leave
every villager knowing nothing, mass starvation. Spec 068 ([[worldmap-generation]],
[[tile-registry]], v5) adds a marsh/sand shoreline terrain pass selected by the
manifest's `terrain_gen` field — a pre-068 build reading that field's absence as
"nothing to see here" and regenerating LEGACY terrain under a v5 manifest would
silently strand agents and structures the daemon actually placed on marsh/sand
(FR-007), so the break exists purely to make that mismatch refuse loudly instead.
Either way, `internal/world.Open`
refuses any manifest whose `format_version` isn't
the current one ([[world-save-directory]]). `promptworld migrate <world>`
([[cli-promptworld]]) is the one-time, offline door a stopped older world walks
through to keep running; it admits a v1 through v4 source and refuses an
already-current world outright ("nothing to migrate").

An older world chains every step it needs in one `migrate` run (1→2→3→4→5 for a
v1 source); a v2 world runs the last three steps, a v3 world the last two, and a
v4 world only the manifest-only bump.

Two children (summary-style, corpus-spec v2) carry the step-by-step detail:

- [[world-migration-catalog]] — each step's design pin (v1→v2 "keep the
  people, reset the land"; v2→v3 carries everything verbatim; v3→v4 grants
  knowledge; v4→v5 is manifest-only) and the write-mechanics that commit a
  successful migration to disk.
- [[world-migration-transforms]] — the four transform algorithms themselves:
  `TransformV1Snapshot`/`TransformV2State`/`TransformV3State` and `Migrate`'s
  manifest-only v4→v5 branch.

## How it works

The work splits across two packages by concern:

- **`internal/world/migrate.go`** is the ceremony: resolve and validate the source
  world (`OpenForMigration`, which admits `format_version` 1, 2, or 3), refuse unsafe
  conditions, archive the original database, write the fresh log, bump the
  manifest. It never touches sim semantics directly.
- **`internal/sim/migrate.go`** is the pure transforms: decode a v1 state shape and
  produce a v2 `sim.State` (`TransformV1Snapshot`/`MigrateState`), carry a v2
  `sim.State` into v3 (`TransformV2State`/`TransformV2Snapshot`), and grant a v3
  `sim.State` its mental maps into v4 (`TransformV3State`/`TransformV3Snapshot`).
  None runs on the live reducer path.

**Client-side and offline**: `Migrate(dir)` refuses a running daemon FIRST (a
pidfile liveness check duplicated from `internal/daemon` rather than imported, to
avoid an import cycle) — this check runs for every source version, v4 included,
ahead of the v4→v5 manifest-only short-circuit (see [[world-migration-transforms]]) — then, for a v1-v3 source,
refuses if the source-format archive already exists — `world.v1.db`
for a v1 source, `world.v2.db` for a v2 source, `world.v3.db` for a v3 source
(`archiveDBPath`, keyed to
`Manifest.FormatVersion`; the already-migrated guard, never overwritten); a v4
source never reaches this archive check at all, since its manifest-only
short-circuit (see [[world-migration-transforms]]) returns first. Keying the
guard to the *source* format means a v2 world produced by an earlier v1→2 migration
— which still carries a stale `world.v1.db` from that run — remains migratable to
v3 (or a v3 world from an earlier run remains migratable to v4): its own guard is
`world.v2.db`/`world.v3.db`, untouched until this run. `Migrate` never
replays old events under new rules; it reads only the source world's **covering
snapshot** — the clean-shutdown guarantee (`CheckContiguity` +
`LatestValidSnapshot`). A real daemon appends a `daemon.stopped` bookkeeping event
after its shutdown snapshot, so a one-event tail of pure `daemon.*` events past the
snapshot is tolerated (they carry zero sim state — nothing to lose); any
sim-affecting event past the snapshot means an unclean stop left un-snapshotted
history, and migration refuses with a remedy: start and stop the world once cleanly
under its own binary, then retry.

## Connections

[[cli-promptworld]]'s `migrate` command is the only caller; [[world-save-directory]]
defines the format-version gate this bridges and the `world.v1.db`/`world.v2.db`/
`world.v3.db` archive artifacts (v4→v5 creates none); [[sim-state-reducer]]'s `Apply` applies `world.migrated` as a
wholesale state replace (validated by matching `Seed`); [[event-types]] catalogs the
payload; [[executor]] is what the migrated agents' inventory (and, from v3, the bulk
cap and death-spill rule) belongs to; [[mental-maps]] is what the v3→v4 step grants
and owns the `MentalMap`/`PlaceFact` types the transform constructs directly;
[[worldmap-generation]] is what the v4→v5 step's bump makes an old build refuse to
mis-generate — the `terrain_gen` field and the marsh/sand pass it gates; [[snapshots]] is the general mechanism v1-v4's steps
borrow (a covering snapshot plus a minimal event tail) to make the migrated log
replay-provable with zero source-format history — the manifest-only v4→v5 step
needs neither. [[world-migration-catalog]] and [[world-migration-transforms]]
carry the per-step detail this note summarizes.

## Operational notes

Migration is irreversible in practice (though recoverable mid-crash — see [[world-migration-catalog]]): the
source-format archive (`world.v1.db`, `world.v2.db`, or `world.v3.db`) is kept, never
auto-deleted, as the human's escape hatch, but nothing in the codebase restores
from it automatically (the v4→v5 step creates no archive at all — there is nothing
to restore from, since nothing changed). A world with no valid covering snapshot (never cleanly
stopped) cannot be migrated at all — there is no path that migrates live event
history directly (this doesn't apply to the archive-free v4→v5 step, which never
reads a snapshot). `internal/sim/migrate_test.go` and `internal/world/migrate_test.go`
exercise all three snapshot-cut transforms and the full command against fixture v1, v2, and v3
worlds, including a v1 fixture that chains 1→2→3→4 in one run
(`TestTransformV3GrantsKnowledge`/`TestTransformV3ChainReducerReplay`,
`TestMigrateV3HappyPath`/`TestMigrateV3ReplayDeterminism` cover the v3→v4 leg
specifically, [[testing-strategy]]); `internal/sim/whole_feature_test.go`
covers the storage-surface event types this migration must also carry correctly.
`internal/world/terrain_gen_test.go` (spec 068, T012) covers the v4→v5 leg
specifically: `TestMigrateV4ManifestOnlyPreservesTerrain` proves a migrated world's
regenerated map `Hash()` is unchanged before and after and that the stand-in event
database bytes are never touched; `TestOpenRejectsV4WithMigrateHint` pins the v4
refusal a pre-migration v5 build shows; `TestNewWorldsMarkedWithTerrainGen`/
`TestOpenRejectsUnknownTerrainGen`/`TestOpenAbsentTerrainGenIsLegacy` cover the
`terrain_gen` field's write/validate/legacy-default behavior on the `Create`/`Open`
side of the same break ([[world-save-directory]]).
