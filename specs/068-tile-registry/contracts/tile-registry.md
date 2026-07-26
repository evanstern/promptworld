# Contract — Tile Registry Invariants + Manifest Compatibility (spec 068)

The promises other code (and tests) may rely on. Each is test-enforced (the Gates-over-
Assertions discipline); the enforcing test is named so tasks.md can mirror them.

## Registry invariants (internal/tui)

- **C1 — Single source**: `renderMapGrid`'s tile resolution, `legendGlyphLine()`, and the
  `?` overlay glyph walkthrough all derive per-tile glyph/name/meaning/style from the one
  registry table. Enforced: extend help_test.go's existing legend/overlay coverage sweep;
  a registered test-only tile appears on map + legend + overlay with no renderer edit
  (SC-002).
- **C2 — Full decode**: every glyph the map can draw has a registry row with a non-empty
  name and meaning; legend + overlay decode 100% of drawable glyphs (SC-005; spec 045
  precedent, extended to marsh/sand).
- **C3 — No stray literals**: no per-tile `lipgloss` style/color literal exists outside
  the registry + token table (`internal/tui/tiles.go`). Enforced: sweep test over the tui
  package source (TASK-121 fiction-literal sweep precedent) (SC-004).
- **C4 — Tokens are classed**: every StyleToken is `semantic16` or `material256`;
  semantic16 tokens use only ANSI 0–15 slots (SC-004 companion assertion).
- **C5 — States are variants**: state presentation (night, dying, cold, damaged,
  needs-critical, suppressed) reuses the base row's glyph — a variant may change style
  only, never the character (FR-003). Enforced: variant-table shape + test.
- **C6 — Byte identity**: for the pre-existing vocabulary, map grid, legend line, and
  overlay walkthrough bytes are identical to pre-refactor output on the pinned fixture
  (FR-004/SC-001).

## Worldmap invariants (internal/worldmap)

- **C7 — Determinism**: `(seed, w, h, gen)` → identical `Map` (incl. `Hash()`) on every
  platform, every run. gen=legacy reproduces today's maps exactly (existing determinism
  test extended with a pinned legacy Hash).
- **C8 — Presence**: for gen=2, every generated map has Marsh > 0 and Sand > 0 across the
  test seed sample (SC-003).
- **C9 — Semantics**: Marsh/Sand are `Passable` (both passability functions), not
  `Buildable`, carry no resource affordances, and never appear in gen=legacy maps.

## Manifest compatibility contract (internal/world)

- **C10 — Refusal, never silent regen**: `format_version` bumps to 5. Pre-feature software
  refuses a v5 world at `Open` with the migrate-hint error (its existing behavior for any
  mismatch); it can never load one and silently regenerate different terrain (FR-007,
  SC-006).
- **C11 — Migration preserves terrain**: `promptworld migrate` upgrades a v4 world to v5
  without setting `terrain_gen`; the regenerated map's `Hash()` before and after migration
  is identical (SC-006).
- **C12 — New worlds are marked**: `promptworld new` writes `format_version: 5,
  terrain_gen: 2`; `Open` treats absent/0 `terrain_gen` as legacy and rejects unknown
  future values with a clear error (forward-compat posture, same shape as the
  format_version check).

## Agent-surface contract (internal/sim)

- **C13 — Named, never mislabeled**: `featureDesc` returns "the marsh" / "the sand flat";
  every terrain-kind switch in sim (executor scan, miracles validation, memory) handles
  the new kinds deliberately (case or documented fall-through) — no site treats them as a
  wrong kind (FR-008).
