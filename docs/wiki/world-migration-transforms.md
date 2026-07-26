---
name: world-migration-transforms
description: The four transform algorithms in detail — sim.TransformV1Snapshot (v1→v2), TransformV2State (v2→v3), TransformV3State (v3→v4), and Migrate's manifest-only v4→v5 branch — what each carries verbatim, resets, or grants
kind: component
sources:
  - internal/sim/migrate.go
  - internal/world/migrate.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# World migration: the four transforms

Split from [[world-migration]] (summary-style, corpus-spec v2): the actual
transform algorithms — decoder shapes, what each carries verbatim versus
resets versus grants, and why the v4→v5 step is different in kind (no
`sim.TransformV4State` exists at all).

## The transforms

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
snapshot-cut ceremony (see [[world-migration-catalog]]) runs, `Migrate` checks
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


## Connections

Back to [[world-migration]] and its sibling child [[world-migration-catalog]]
(design-pin summary + write-to-disk mechanics). [[sim-state-reducer]] applies
`world.migrated` as a wholesale replace; [[mental-maps]] owns the
`MentalMap`/`PlaceFact` types v3→v4 constructs; [[worldmap-generation]] is
what the v4→v5 bump guards.
