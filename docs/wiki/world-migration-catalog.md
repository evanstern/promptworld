---
name: world-migration-catalog
description: The four migration steps' design pins (v1→v2 keep-people-reset-land, v2→v3 carry-everything, v3→v4 grant-knowledge, v4→v5 manifest-only) plus the write-mechanics that commit a migrated world to disk
kind: component
sources:
  - internal/sim/migrate.go
  - internal/world/migrate.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# World migration: step catalog & write mechanics

Split from [[world-migration]] (summary-style, corpus-spec v2): what each of
the four migration steps carries, resets, or grants — the design pin that
differs step to step because each step's INPUTS differ — and how a
successful migration commits its result to disk.

## The four migration steps

The four steps have different design pins because their inputs differ:

- **v1→v2** (research R10): **"keep the people, reset the land"** — terrain
  generation itself changed (rock outcrops), so every villager and the whole
  social/governance fabric carry over verbatim while the map and everything bound
  to it (structures, overlays, in-flight intents/plans) is reborn under v2 rules —
  migrated souls are re-placed via `genesisPlacement`.
- **v2→v3** (research R3): people **and** land carry verbatim — spec 013 changed no
  terrain generation and no map inputs, so there is nothing to reset. Agents keep
  their exact coordinates (no re-placement), structures/overlays/mid-flight intents
  ride through unchanged, and the only adjustment is the new bulk-cap invariant.
- **v3→v4** (spec 041, research D7): people and land carry verbatim, same as
  2→3 — the one addition is KNOWLEDGE. Villagers are NATIVES, not strangers:
  each living agent is granted explored terrain around its current position
  (perception radius) plus witnessed place-facts for every current structure
  and ground pile, all stamped at the migration tick — never a blank map that
  would have to re-discover a village it already lives in.
- **v4→v5** (spec 068, TASK-143, C11): **manifest-only** — the one step that
  transforms nothing. A carried-forward v4 world's event log, state, and terrain
  don't change: its `terrain_gen` stays absent, so it keeps generating LEGACY
  terrain, bit-identical to what it always has. The version bump exists solely so
  pre-068 software refuses to open it (rather than silently regenerating terrain
  under an algorithm the manifest didn't ask for); there is nothing to archive,
  transform, or cut a fresh snapshot for.

## Writing the result

**Writing the result**: after the transform succeeds, `archiveDB` renames
`world.db` (and any `-wal`/`-shm` sidecars) to the source-format archive —
`world.v1.db` for a v1 source, `world.v2.db` for a v2 source, `world.v3.db`
for a v3 source — the point of no easy
return, so every refusal has already run. A fresh `world.db` is opened and gets
exactly two events, both stamped at the continuation tick: `world.created` (same
name/seed) then `world.migrated`, whose payload (`WorldMigratedPayload` —
[[event-types]]) carries `FromFormat`, `SourceEvents` (the source log's last seq),
`SourceTick`, and the full transformed `State` embedded whole. A covering snapshot
is saved at the same tick — deleting it and replaying `world.created` →
`world.migrated` must reproduce the identical state (the determinism half of SC-007
in both specs). The manifest's `FormatVersion` is bumped to the current version
**last**: a crash between the archive and the manifest write leaves a recoverable
state (the source-format archive present, manifest still at the source version —
restore is the same rename-back).

## Connections

Back to [[world-migration]] for the ceremony overview and its sibling child
[[world-migration-transforms]] (the actual per-step transform algorithms
this catalog summarizes). [[event-types]] catalogs `WorldMigratedPayload`;
[[snapshots]] is the covering-snapshot mechanism every step but v4→v5
borrows.
