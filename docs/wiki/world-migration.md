---
name: world-migration
description: The snapshot-cut migration chain (spec 012 US6 v1→v2, spec 013 v2→v3, spec 041 v3→v4) plus the spec-068 manifest-only v4→v5 step — carries a stopped world's people (and, from v2 on, its land, and from v3 on, a granted mental map) across a save-format break as a single world.migrated event, or (v4→v5) just bumps the manifest
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

An older world chains every step it needs in one `migrate` run (1→2→3→4→5 for a
v1 source); a v2 world runs the last three steps, a v3 world the last two, and a
v4 world only the manifest-only bump.

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
ahead of the v4→v5 manifest-only short-circuit above — then, for a v1-v3 source,
refuses if the source-format archive already exists — `world.v1.db`
for a v1 source, `world.v2.db` for a v2 source, `world.v3.db` for a v3 source
(`archiveDBPath`, keyed to
`Manifest.FormatVersion`; the already-migrated guard, never overwritten); a v4
source never reaches this archive check at all, since its manifest-only
short-circuit above returns first. Keying the
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

**The v1→v2 transform** (`sim.TransformV1Snapshot` / `MigrateState`) decodes the v1
covering-snapshot JSON through a typed `legacyState`/`legacyAgent`/`legacyInventory`
shape — decoding straight into the v2 `Inventory` would silently drop `food` (the
one field where v1 and v2 diverge incompatibly; every other agent field either is
unchanged or is v2-added, so it decodes faithfully as its zero value from absent v1
JSON). The transform then:

- **Carries verbatim** (tick continuity intact): clock (`Tick`/`Paused`/`Speed`/
  `Night` — the migration tick *is* the carried v1 tick, so the clock simply
  continues), the whole social fabric (relations, debts, rumors + id counters), the
  conversation ring, the chronicle ring, the Guardian's charge bank, and governance
  (norms + their id counters — the village's lived law); per-agent people-state
  (needs, memories, beliefs, narrative, generation, consolidation marks, talk/give
  cooldowns, known rumors, `NearDeath`, `Dead` — a villager who died in the old
  world stays dead, not resurrected by the break).
- **Resets** (map-/session-bound, nil/zero): `Structures`, `Cleared`, `Harvested`,
  `DenUses`, `Quarried`, `Gru`, `MeetingConvention`/`MeetingPlace`/`Meeting` (the
  in-flight session — re-seeded from `world.json` on next boot, or re-emerges), and
  per-agent `Intent`/`Plan`/`Hail`/`Asleep` (everyone wakes standing, freshly idle
  at the migration tick via `IdleSince`).
- **Attaches the map** (spec 016): `MigrateState` sets the resulting `State`'s
  unexported `m *worldmap.Map` field (via the same construction path `NewState`
  uses, [[sim-state-reducer]]), so a migrated state is map-aware for the
  miracle reducer arms exactly like a fresh genesis or a live replica.
- **Re-places** every carried soul via `genesisPlacement` — the same deterministic
  placement `NewState` uses for a fresh genesis of that seed
  ([[sim-state-reducer]], [[deterministic-rng]]) — so migrated villagers land on
  passable v2 tiles (rock outcrops included) exactly where a brand-new world of the
  same seed would put them.
- **Re-expresses inventory**: `Wood` carries 1:1; the legacy `Food` count converts to
  `Meals` at a pinned rate (`legacyFoodToMeals` = 3 — a mild haircut, 350→300
  restore, flavored as preserved meals crossing the break); every new v2 kind
  (`Stone`/`Water`/`Planks`/`RefinedStone`/`FoodRaw`/`FoodCooked`/`Spears`) starts
  empty.

**The v2→v3 transform** (`sim.TransformV2State` / `TransformV2Snapshot`) needs no
distinct legacy decoder: every v3 addition (`State.Piles`, `Structure.Owner`/
`Store`, `Intent.Kind`/`Qty`) is additive and `omitempty`, so a v2 snapshot's JSON
decodes straight into the current `sim.State`, new fields landing on their zero
values. The transform is a pure function of the input (it copies the `Agents` and
`Piles` slices before mutating, so the caller's state is never touched) that:

- **Carries everything verbatim**: positions (no re-placement — v3 changed no
  terrain), structures, overlays (`Quarried`/`Cleared`/`Harvested`), mid-flight
  intents/plans (unlike 1→2's wipe — no map inputs changed, so targets stay valid
  and the bulk cap simply applies at completion), the whole social/governance
  fabric, and the clock (`Degraded` resets to `false` and `EffectiveRate` is
  refreshed from `Speed`, exactly as the v1→v2 step does — a stopped world carries
  no live drift across the break).
- **Applies only the bulk-cap invariant** (research R3 "Decisions taken at
  implementation"): a living agent whose carried bulk exceeds `bulkCap` spills the
  excess to a ground pile at its own tile; a dead agent spills its *entire* frozen
  inventory — the v3 death-spill invariant (see [[executor]]) carried forward, so a
  migrated world matches what v3 itself would have produced from that death.
  Spilling removes goods in canonical kind order: within food, least-nutritious
  first (`food_raw` → `food_cooked` → `meals`, so a capped villager keeps its best
  food); spears spill most-worn-first, mirroring the give/drop transfer idiom.
  Spilled food batches are stamped `SpoilAt` = migration tick + `rotWindowTicks`,
  same as any other fresh drop.

**The v3→v4 transform** (`sim.TransformV3State` / `TransformV3Snapshot`, spec 041
research D7) needs no distinct legacy decoder either — `Agent.Map` is additive and
`omitempty`, so a v3 snapshot's JSON decodes straight into the current `sim.State`
with every map nil. The transform (pure — the input and its slices are never
mutated; `m` must be the regeneration of the world's own seed, since it sizes
the explored bitmaps) then:

- **Carries everything verbatim**: exactly the 2→3 carry list (positions, land,
  overlays, mid-flight intents/plans, the whole social/governance fabric); the
  clock continues from the carried tick with `Degraded` reset and
  `EffectiveRate` refreshed, the same precedent as both earlier steps.
- **Grants each LIVING agent knowledge**: explored terrain around its current
  position at `witnessRadius` (the perception radius), witnessed place-facts
  for every current structure (a fire's `FuelUntil` baked as `Detail`) and
  ground pile — all stamped `Seen` at the migration tick, canonical
  `(Kind, X, Y)` order — and (T013) a peer sighting of every OTHER living
  villager at its current position, so `talk_to` stays viable across the
  break. Villagers are natives of the village they already live in, not
  strangers dropped into it.
- **Gives a DEAD agent an empty but non-nil map**, not the zero value left
  alone: genesis now seeds maps for every agent, and a replica/recovery
  unmarshal MERGES a snapshot over a genesis state — a map-ABSENT dead agent
  would silently resurrect the genesis map there, while a from-genesis replay
  (`world.created` → `world.migrated`) produces the transform's own value; an
  explicit empty map on every agent is what makes live-and-replay agree
  byte-for-byte (a deviation from `tasks.md`'s "living agents" phrasing,
  recorded for the planning tier — `data-model.md`'s "dead agents: map
  retained" invariant already wants the field present).

**The v4→v5 step** (spec 068 C11, `Migrate`'s own early branch — there is no
`sim.TransformV4State`, because nothing in `sim.State` changes): before any of the
snapshot-cut ceremony below runs, `Migrate` checks
`w.Manifest.FormatVersion == 4` and, if so, takes a SEPARATE short path: bump
`Manifest.FormatVersion` to the current version (5), write the manifest, and return
`MigrateResult{Name, Seed, ManifestOnly: true}` — every other `MigrateResult` field
(`AgentsCarried`, `Tick`, `SourceEvents`, `ArchivePath`) stays zero, since none of
that ceremony ran. No archive is created, no fresh log is written, no covering
snapshot is cut — cutting one here would imply a state break that does not exist,
and would demand a covering snapshot for what is otherwise a no-op. The world's
`terrain_gen` field stays exactly what it was (absent, for every world that predates
spec 068) — a migrated v5 world therefore keeps generating LEGACY terrain
([[worldmap-generation]]), the whole point being that only a `promptworld new`-born
world gets the marsh/sand pass.

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
needs neither.

## Operational notes

Migration is irreversible in practice (though recoverable mid-crash, above): the
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
