# Research — Tile Registry + New Terrain Tiles (spec 068)

Phase 0 findings. Every decision below resolves an unknown left open by spec.md; file:line
references are the evidence trail (verified 2026-07-26 at root, main = 070a21b).

## R1 — Current rendering anatomy (what the registry replaces)

- **Styles**: ~20 `lipgloss` literals in a package-level var block,
  `internal/tui/views.go:1581-1644` (terrain, structures, agent condition overlays).
- **Tile resolution**: `renderMapGrid`'s `tile(x, y)` closure, `views.go:1290-1330` —
  priority gru > agents > structures > piles > dens > path > quarried > base terrain switch
  (`worldmap.Water/Tree/Forage/Rock` → glyph+style, default grass `·` dim).
- **Names/meanings**: already centralized — `mapGlyphs []glyphEntry` in
  `internal/tui/help.go:57` (Glyph, Name, Meaning), feeding both `legendGlyphLine()`
  (compact legend) and the `?` overlay glyph walkthrough (spec 045 FR-005 anti-drift seam).
- **Decision**: the registry *grows `glyphEntry`* (adds style-token + kind-binding fields and
  a tile-resolution role) rather than inventing a parallel table — the single-source seam
  and its coverage tests already exist; we extend them.

## R2 — Style tokens (FR-002 shape)

- Precedent: TASK-121/spec 052 made fiction strings skin data; Cogmind externalizes message
  colors to a file. For this feature the tokens are **named Go-level constants/vars in one
  table** (e.g. `tokenWaterSemantic`, `tokenShelterMaterial`), each carrying its class
  (semantic-16 vs material-256) — *not* a user-editable file yet.
- **Rationale**: spec assumption limits scope to "the unit a future skin would override";
  file-based theming is future work. A `TestTileStyleTokenClasses` sweep asserts every
  registry style resolves through a classed token and no `lipgloss.Color(` per-tile literal
  survives outside the token table (SC-004; TASK-121 sweep precedent).
- Classing of the existing palette (from the analysis): semantic-16 = water(4), tree(2),
  forage(3), den(5), critical(1), dead/gru reds; material-256 = 130/136/137/166/178/196/202/
  208/240/244/245/250 + suppressed 135. The refactor must preserve exact output — classing
  is metadata, colors do not change (FR-004).

## R3 — New kinds: glyphs and colors (FR-005 finalization)

- **Marsh = `░` (U+2591 light shade), material color 256:65** (muted wet green, distinct
  from tree's ANSI 2 and forage's ANSI 3 under the family-tint discipline).
- **Sand = `▒` (U+2592 medium shade), material color 256:180** (pale warm tan, distinct
  from path 137, pile 178, chest 136).
- Different *characters*, so 16-color collapse keeps them apart (edge case honored); both
  distinct from every shipped glyph including grass/path `·` and depleted `,`.
- Alternatives considered: `≈` for marsh (rejected — too close to water `~` clustered);
  `.`-variants (rejected — color-only vs grass violates the grammar for a new *kind*).

## R4 — Generation placement (FR-006)

`worldmap.Generate` (worldmap.go:100-190) is percentile-driven over two fbm fields
(elevation, moisture), integer/hash noise, platform-deterministic.

- **Marsh**: dry-land tiles 4-adjacent to Water whose moisture exceeds a marshFraction
  percentile — "low-lying wet ground near water".
- **Sand**: remaining dry-land tiles 4-adjacent to Water — "shoreline ring" (marsh wins
  where both apply; sand is the drier shore).
- Applied AFTER water/tree/rock/forage passes so existing fractions shift only where the
  new kinds claim tiles; tuning constants alongside the existing `waterFraction` block.
- Nonzero-count guarantee (SC-003): every generated map has a water body (existing tests),
  so a shoreline always exists; test across a seed sample like `worldmap_test.go`'s
  fraction tests.

## R5 — Generation gating + old-software safety (FR-007 mechanism)

- Terrain is regenerated from the manifest on every open and never persisted
  (worldmap.go package doc; `Manifest` comment world.go:36).
- `world.Open` **rejects** any `format_version != FormatVersion` with a migrate hint
  (world.go:293) — this is the only existing hard gate old binaries enforce.
- **Decision**: bump `FormatVersion` 4 → 5 AND add an additive manifest field
  `terrain_gen` (int, `omitempty`): absent/0 ⇒ legacy generation (exactly today's
  algorithm); 2 ⇒ marsh+sand generation. `promptworld new` writes
  `format_version:5, terrain_gen:2`. `promptworld migrate` upgrades v4 worlds to v5
  *without* setting `terrain_gen` — their terrain stays bit-identical (FR-006, SC-006).
- **Why both**: the field alone is silently ignored by old readers (the exact
  mis-generation FR-007 forbids); the version bump alone can't distinguish
  migrated-legacy from new worlds. Old binary + new world ⇒ clean refusal at Open
  (SC-006's "refused, never silently re-generated"). Precedent: FormatVersion 4 was the
  same posture for spec 041's mental-maps break.
- `Generate` gains the version parameter (e.g. `Generate(seed, w, h)` keeps legacy;
  `GenerateV(seed, w, h, gen)` or an options struct — implementer's choice) — every caller
  (daemon, TUI replica, tests) resolves the version from the manifest.

## R6 — Sim semantics of marsh/sand (FR-008 boundaries)

- **Walkable**: two passability functions exist and BOTH must add the new kinds —
  `worldmap.Map.Passable` (worldmap.go:60, Grass||Forage) and sim `passable`
  (internal/sim/terrain.go:38, Grass||Forage||Depleted + wall check).
- **Not buildable**: `Buildable` stays Grass-only (worldmap.go:70) — marsh is wet, sand is
  loose; keeps every build/miracle path untouched. (Spec: "no resource affordances".)
- **Not a resource**: mental maps / policy keys ("tree","forage","rock" strings —
  policy.go, memory.go) gain no new keys; `effectiveKind` overlays (terrain.go:10) don't
  touch the new kinds.
- **Named to agents**: `featureDesc` (memory.go:107) returns "" for grass (ordinary
  ground); marsh and sand get real phrases ("the marsh", "the sand flat") — they are
  *notable* terrain. Sweep every `switch … worldmap.` site for exhaustiveness:
  executor.go:496 (scan/look), memory.go:126, miracles.go:524, terrain.go:13; add cases
  (or explicit fallthroughs) so no site mishandles the new kinds (FR-008's "no fallback
  label").

## R7 — Byte-identical verification (FR-004/SC-001 method)

- TUI package tests already assert rendered content (tui_test.go, digest_test.go,
  family-tint distinctness tests, help_test.go keymap/coverage sweeps) — they run
  unchanged and must pass.
- Add a pre/post pin: a fixture world (fixed seed, representative structures/agents)
  rendered via `renderMapGrid` + `legendGlyphLine` + overlay walkthrough, with the
  expected bytes captured BEFORE the refactor lands (US1 acceptance 1). The
  `worldmap.Map.Hash()` fingerprint (worldmap.go:85) pins terrain identity for the
  legacy-generation path (US2 acceptance 2 / SC-006).

## R8 — Model tier (constitution V)

Cross-package (tui + worldmap + world + sim), a format-version break with migration, and
determinism-sensitive generation → **senior implementation tier (Opus 4.8)** per the
escalation rubric ("cross-package or architectural changes"). Recorded on TASK-143.
