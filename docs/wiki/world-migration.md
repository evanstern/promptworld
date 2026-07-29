---
name: world-migration
description: The offline migration driver in two modes (spec 094 decision rule) — the snapshot-cut chain (v1→v2→v3→v4 state shape) that carries a stopped world across a SEMANTIC break as one world.migrated event, and the TRANSLATING mode (v4/v5→v6) that rewrites a log's type column through the rename table for a PURE-RENAME break with every event preserved; per-step design pins and the transform algorithms live in two split-off children
kind: component
sources:
  - internal/sim/migrate.go
  - internal/world/migrate.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# World migration

`promptworld migrate` is the offline driver for BOTH migration modes the
spec 094 decision rule distinguishes:

- **snapshot-cut** (semantic breaks — the payload's meaning or the
  reducer's derivation changed): never replays old events under new rules;
  reads the source's covering snapshot, transforms it, and writes a fresh
  log whose single `world.migrated` event carries the whole transformed
  state. History is archived, not carried.
- **translation** (pure renames — spec 094's `metatron.*`→`guardian.*`
  rename, the first user): rewrites the log's type column through
  `sim.LogFormatV1Renames` with EVERY event's seq, tick, payload, and
  wall_time preserved byte-for-byte, copies snapshots and meta verbatim,
  stamps the new log format ([[event-log]]), and verifies the translated
  log replays (snapshot + tail, exactly what boot does) before swapping it
  in. Proven by the byte-identity harness
  (`internal/world/migrate_translate_test.go`): replay(source, old
  semantics) == replay(translated, new semantics) as state-hash sequences,
  per-event byte identity on disk, and the migrated world running forward
  on the new binary.

promptworld has broken its save format five times. Spec 012 (resources/food/crafting
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
Spec 094 (**v6**) renamed the 13 persisted `metatron.*` guardian event types to
`guardian.*` — a v5 log replayed under the renamed reducer arms would silently
hit no case for half the guardian history, so v6 is the first PURE-RENAME break
and the first user of the translating mode (with the log's own
`log_format_version` stamp, [[event-log]], enforcing the same refusal from
inside the DB).
Either way, `internal/world.Open`
refuses any manifest whose `format_version` isn't
the current one ([[world-save-directory]]). `promptworld migrate <world>`
([[cli-promptworld]]) is the one-time, offline door a stopped older world walks
through to keep running; it admits a v1 through v5 source and refuses an
already-current world outright ("nothing to migrate").

An older world chains every snapshot-cut step it needs in one `migrate` run
(1→2→3→4-shaped state for a v1 source) and lands on a fresh log that already
speaks the current vocabulary (stamped log format 2 directly — its only events
are `world.created`/`world.migrated`). A v4 or v5 source takes the translation
path instead: v4→v5 was manifest-only (spec 068), so both speak log format 1
and their logs are content-identical — the full history carries over with only
the type names rewritten, archived as `world.v4.db`/`world.v5.db`.

Two children (summary-style, corpus-spec v2) carry the step-by-step detail:

- [[world-migration-catalog]] — each step's design pin (v1→v2 "keep the
  people, reset the land"; v2→v3 carries everything verbatim; v3→v4 grants
  knowledge; v4/v5→v6 translates the vocabulary) and the write-mechanics
  that commit a successful migration to disk.
- [[world-migration-transforms]] — the transform algorithms themselves:
  `TransformV1Snapshot`/`TransformV2State`/`TransformV3State` and
  `migrateTranslate`'s type-column rewrite (the v4/v5→v6 leg; the old
  manifest-only v4→v5 branch was subsumed by it when v6 landed).

## How it works

The work splits across two packages by concern:

- **`internal/world/migrate.go`** is the ceremony for BOTH modes: resolve and
  validate the source world (`OpenForMigration`, which admits `format_version`
  1 through 5), refuse unsafe conditions, archive the original database, write
  the fresh (or translated) log, bump the manifest. The translation's rename
  table lives in `internal/sim` (`LogFormatV1Renames`, the vocabulary owner);
  the ceremony never touches sim semantics directly.
- **`internal/sim/migrate.go`** is the pure transforms: decode a v1 state shape and
  produce a v2 `sim.State` (`TransformV1Snapshot`/`MigrateState`), carry a v2
  `sim.State` into v3 (`TransformV2State`/`TransformV2Snapshot`), and grant a v3
  `sim.State` its mental maps into v4 (`TransformV3State`/`TransformV3Snapshot`).
  None runs on the live reducer path.

**Client-side and offline**: `Migrate(dir)` refuses a running daemon FIRST (a
pidfile liveness check duplicated from `internal/daemon` rather than imported, to
avoid an import cycle) — this check runs for every source version, ahead of the
mode dispatch — then refuses if the source-format archive already exists:
`world.v1.db` … `world.v5.db` (`archiveDBPath`, keyed to
`Manifest.FormatVersion`; the already-migrated guard, never overwritten).
Keying the
guard to the *source* format means a v2 world produced by an earlier v1→2 migration
— which still carries a stale `world.v1.db` from that run — remains migratable
onward: its own guard is the archive of ITS source version, untouched until this
run. The snapshot-cut modes never
replay old events under new rules; they read only the source world's **covering
snapshot** — the clean-shutdown guarantee (`CheckContiguity` +
`LatestValidSnapshot`). A real daemon appends a `daemon.stopped` bookkeeping event
after its shutdown snapshot, so a one-event tail of pure `daemon.*` events past the
snapshot is tolerated (they carry zero sim state — nothing to lose); any
sim-affecting event past the snapshot means an unclean stop left un-snapshotted
history, and migration refuses with a remedy: start and stop the world once cleanly
under its own binary, then retry. The TRANSLATING mode (`migrateTranslate`)
needs no covering snapshot at all — it carries the whole log — but demands
contiguity, builds the translated DB at a temp path
(`world.db.translating`), verifies it replays exactly as boot will
(latest snapshot + tail through `sim.State.Apply`), and only then archives the
source and swaps the verified translation into place; the manifest bumps last
(the same crash posture: archive present + manifest unbumped = restore by
renaming back).

## Connections

[[cli-promptworld]]'s `migrate` command is the only caller; [[world-save-directory]]
defines the format-version gate this bridges and the `world.v1.db` …
`world.v5.db` archive artifacts; [[event-log]] owns the log-format stamp the
translation writes and the load gate that refuses an untranslated log;
[[sim-state-reducer]]'s `Apply` applies `world.migrated` as a
wholesale state replace (validated by matching `Seed`) and states the
rename/re-derivation doctrine this driver discharges; [[event-types]] catalogs the
payload; [[executor]] is what the migrated agents' inventory (and, from v3, the bulk
cap and death-spill rule) belongs to; [[mental-maps]] is what the v3→v4 step grants
and owns the `MentalMap`/`PlaceFact` types the transform constructs directly;
[[worldmap-generation]] is what the spec-068 bump makes an old build refuse to
mis-generate — the `terrain_gen` field and the marsh/sand pass it gates; [[snapshots]] is the general mechanism the snapshot-cut steps
borrow (a covering snapshot plus a minimal event tail) to make the migrated log
replay-provable with zero source-format history — the translating step instead
carries the source's own latest verified snapshot verbatim.
[[world-migration-catalog]] and [[world-migration-transforms]]
carry the per-step detail this note summarizes.

## Operational notes

Migration is irreversible in practice (though recoverable mid-crash — see [[world-migration-catalog]]): the
source-format archive (`world.v1.db` … `world.v5.db`) is kept, never
auto-deleted, as the human's escape hatch, but nothing in the codebase restores
from it automatically. A world with no valid covering snapshot (never cleanly
stopped) cannot take a SNAPSHOT-CUT step — there is no path that migrates live
event history through a transform — while the translating mode carries the
whole log and needs no snapshot (one carried over verbatim when present).
`internal/sim/migrate_test.go` and `internal/world/migrate_test.go`
exercise all three snapshot-cut transforms and the full command against fixture v1, v2, and v3
worlds, including a v1 fixture that chains the snapshot-cut steps in one run
(`TestTransformV3GrantsKnowledge`/`TestTransformV3ChainReducerReplay`,
`TestMigrateV3HappyPath`/`TestMigrateV3ReplayDeterminism` cover the v3→v4 leg
specifically, [[testing-strategy]]); `internal/sim/whole_feature_test.go`
covers the storage-surface event types this migration must also carry correctly.
`internal/world/migrate_translate_test.go` (spec 094) is the translation
suite: `TestTranslatingMigrationByteIdentity` (a 13-type v5 fixture —
refusal, translation, per-event byte identity, state-hash-sequence identity,
snapshot+tail boot, running forward, idempotence + the crash-posture archive
guard), `TestTranslatingMigrationRefusesLiveDaemon`,
`TestTranslatingMigrationV4Source`, and `TestRenameMapMatchesCatalog` (the
rename table pinned against `sim.PayloadCatalog`).
`internal/world/terrain_gen_test.go` (spec 068 T012, amended by spec 094)
keeps the terrain contract on the v4 leg: `TestMigrateV4PreservesTerrain`
proves a migrated v4 world's regenerated map `Hash()` is unchanged (migrate
still never sets `terrain_gen`); `TestOpenRejectsV4WithMigrateHint` pins the
old-format refusal; `TestNewWorldsMarkedWithTerrainGen`/
`TestOpenRejectsUnknownTerrainGen`/`TestOpenAbsentTerrainGenIsLegacy` cover the
`terrain_gen` field's write/validate/legacy-default behavior on the `Create`/`Open`
side ([[world-save-directory]]).
