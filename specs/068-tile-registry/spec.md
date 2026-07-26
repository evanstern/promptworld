# Feature Specification: Tile Registry + New Terrain Tiles

**Feature Branch**: `task-143-tile-registry`

**Created**: 2026-07-26

**Status**: Draft

**Input**: User description: "Tile registry + new terrain tiles (TASK-143): (1) extract the TUI map's hard-coded glyph/color style table into a data-driven tile registry — one tile table as the single source feeding the map renderer, the legend line, and the ? overlay glyph walkthrough, with colors expressed as tokens rather than scattered style literals; (2) add 2-3 new ground-cover terrain kinds from the CP437 shading tier of the Game-UI-UX tile-vocabulary-expansion analysis (e.g. marsh, sand) wired through worldgen and rendered via the new registry. Existing worlds and replays must render byte-identically for the pre-existing vocabulary."

**Grounding**: `research/Game-UI-UX/Analysis-Tile-Vocabulary-Expansion.md` (2026-07-26) and its
published briefing. The analysis's rules are treated as fixed constraints:

- Expand by borrowing conventions (roguelike dictionary → CP437 shading tier → novel Unicode
  last); every new glyph must stay distinct from the existing set when densely clustered at
  small sizes.
- Spend color before characters: a new *state* of an existing thing recolors the learned
  glyph, never introduces a new character.
- Semantic (meaning-bearing) colors live on the 16 themeable terminal colors; material
  (decorative) colors may use the fixed 256 palette.
- Never color alone for a distinction a player must act on.
- The map legend and the `?` overlay glyph walkthrough remain one shared source that can
  never silently diverge (the spec 045 discipline).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - One tile table drives map, legend, and overlay (Priority: P1)

As the maintainer (and as any future skin/tileset author), all knowledge about how a map
tile looks and is explained — its glyph, its short legend name, its plain-language overlay
meaning, and its visual style — lives in one data-driven registry. The map renderer, the
compact legend line, and the `?` overlay glyph walkthrough all read that registry. Adding
or re-skinning a tile is a data change in one place, not a code sweep.

**Why this priority**: This is the analysis's Rule 4 ("one grid model, swappable skins" —
the Dwarf Fortress Classic/Premium architecture) and the precondition for every future tile
addition, including User Story 2 and any later graphical/web tileset. Today the glyph names
and meanings are already one shared table, but the styles are ~20 scattered code literals
and the terrain-to-glyph mapping is an inline switch — three sources that can drift.

**Independent Test**: Can be fully tested on existing worlds alone: refactor with zero new
tiles and verify the rendered map, legend line, and overlay walkthrough are byte-identical
to the pre-refactor output for every existing fixture and golden test.

**Acceptance Scenarios**:

1. **Given** an existing world fixture rendered before the change, **When** the same fixture
   renders after the change, **Then** map grid, legend line, and `?` overlay output are
   byte-identical.
2. **Given** the registry, **When** a maintainer adds one new tile row (glyph, name, meaning,
   style), **Then** the tile renders on the map, appears in the compact legend, and appears
   in the overlay walkthrough with no other edit.
3. **Given** the registry's style definitions, **When** the codebase is swept for per-tile
   visual style literals outside the registry, **Then** none remain (state-overlay styles
   layered *on top of* tiles — night dimming, needs-critical, suppressed-mind, damaged —
   remain style transforms applied to registry styles, defined alongside them).
4. **Given** the registry, **When** styles are inspected, **Then** each is expressed through
   named tokens classed as *semantic* (themeable 16) or *material* (fixed 256), matching the
   analysis's palette rule.

---

### User Story 2 - New ground covers: marsh and sand (Priority: P2)

As a player creating a new world, the terrain vocabulary is richer: low-lying wet ground
near water reads as **marsh** and shoreline ground reads as **sand**, each with its own
glyph from the CP437 shading tier, its own material color, a legend entry, and a
plain-language overlay meaning. The world feels more varied at zero gameplay risk: both are
open, walkable ground with no resource affordances in this feature.

**Why this priority**: This is the "expand available tiles" payoff, and it exercises the
registry end-to-end (a real addition made the way all future additions will be made). It
depends on User Story 1's registry to land cleanly.

**Independent Test**: Create a new world from a fixed seed; assert both new kinds generate
at nonzero counts, render via the registry, and are decoded by legend + overlay. Open an
existing (pre-feature) world and assert its terrain is unchanged.

**Acceptance Scenarios**:

1. **Given** a new world created after this feature, **When** its map generates, **Then**
   marsh and sand tiles exist at nonzero counts, placed deterministically (same seed and
   dimensions → identical map, every platform).
2. **Given** a world created before this feature, **When** it is opened after the upgrade,
   **Then** its regenerated terrain is identical to what it was before the upgrade (no
   marsh, no sand, no shifted tiles).
3. **Given** a new-world map, **When** the player opens the `?` overlay glyph walkthrough,
   **Then** marsh and sand each have a row with glyph, name, and a plain-language meaning,
   and the compact legend carries matching tokens.
4. **Given** villagers pathing across a new world, **When** they traverse marsh or sand,
   **Then** both behave as open walkable ground (like grass), and agent-facing observations
   that name terrain use the new kinds' names rather than a fallback label.
5. **Given** a dense cluster of mixed tiles at minimum map size, **When** rendered, **Then**
   marsh and sand remain visually distinct from each other and from every existing glyph
   (different characters, not color-only — the shading-tier characters differ).

---

### Edge Cases

- **Old software opening a new world**: a world whose terrain was generated with the new
  vocabulary MUST NOT be silently re-generated *without* it by older software (agents and
  structures could end up standing on water). The world manifest must make new-terrain
  worlds unmistakable to old readers — the mechanism (format-version bump vs. a generation
  version field that old readers reject) is a plan decision, but silent mis-generation is
  unacceptable.
- **Night rendering**: night dims the palette rather than hiding the world; the dimming
  transform must apply to registry-driven styles exactly as it did to the literals,
  including the new kinds.
- **16-color terminals**: material colors (256 palette) may collapse; the new kinds must
  still be distinguishable by glyph alone (guaranteed by scenario US2-5's
  different-characters requirement).
- **Tile priority**: agents > structures > piles > dens > terrain ordering is unchanged;
  marsh/sand are terrain-level and lose to everything above, like grass.
- **Depleted/path re-checks**: the existing effective-kind overrides (depleted `,`, path
  `·`) must keep winning over base terrain in the registry-driven switch exactly as today.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A single tile registry MUST be the only source of per-tile presentation data:
  glyph, compact legend name, plain-language overlay meaning, and visual style. The map
  renderer, the compact legend line, and the `?` overlay glyph walkthrough MUST all derive
  their tile content from it.
- **FR-002**: Registry styles MUST be expressed as named tokens, each classed as *semantic*
  (one of the 16 themeable terminal colors) or *material* (fixed 256-palette color), per the
  analysis's palette rule. No per-tile style literal may exist outside the registry and its
  token definitions.
- **FR-003**: State presentation (night dimming, needs-critical, suppressed-mind, dying,
  damaged, cold/spent) MUST remain style *transforms/variants* of a learned glyph — the
  registry MUST NOT introduce new characters for states.
- **FR-004**: The refactor (User Story 1) MUST be presentation-neutral: for the pre-existing
  vocabulary, rendered map, legend, and overlay bytes are unchanged on every existing
  fixture.
- **FR-005**: Two new walkable ground-cover terrain kinds — **marsh** and **sand** — MUST be
  added, each with a distinct CP437 shading-tier character (candidates `░` and `▒`; final
  glyph/color assignment is a plan decision bound by the distinctness and palette rules).
- **FR-006**: New-kind generation MUST be deterministic (same seed + dimensions + generation
  version → identical map on every platform) and MUST occur only for worlds created after
  this feature; a pre-existing world's regenerated terrain is bit-identical to before.
- **FR-007**: A world carrying new-vocabulary terrain MUST be unmistakably marked in its
  manifest such that software predating this feature cannot silently regenerate its terrain
  differently (see the old-software edge case).
- **FR-008**: Marsh and sand MUST behave as open walkable ground with no resource
  affordances; every agent-facing surface that names terrain kinds MUST name the new kinds
  correctly (no empty/fallback labels in observations, look-cursor text, or prompts that
  enumerate terrain).
- **FR-009**: The legend line and `?` overlay MUST decode 100% of what the map can draw
  after the addition (the spec 045 coverage discipline extends to the new kinds).

### Key Entities

- **Tile registry entry**: one tile's identity and presentation — glyph, legend name,
  overlay meaning, style (token references), and the world/terrain kind or structure kind it
  binds to.
- **Style token**: a named color/emphasis definition classed semantic-16 or material-256;
  the unit a future skin would override (TASK-121 skin precedent, Cogmind externalized-
  colors precedent).
- **Terrain kind — marsh, sand**: new walkable ground covers; generated for new worlds
  only; no resource affordances.
- **World manifest generation marker**: whatever field/version makes new-terrain worlds
  unmistakable to old readers (mechanism decided in plan).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of existing rendering fixtures/golden tests pass unchanged after the
  registry refactor (byte-identical map, legend, overlay for the pre-existing vocabulary).
- **SC-002**: Adding a subsequent tile after this feature requires touching exactly one
  registry data location for presentation (demonstrated in a test that registers a tile and
  finds it on map + legend + overlay with no renderer edits).
- **SC-003**: On 100% of new-world seeds sampled in tests, marsh and sand both generate at
  nonzero counts and the map is identical across repeated generations and platforms.
- **SC-004**: Zero per-tile style literals remain outside the registry (verified by
  sweep/test, the TASK-121 fiction-literal sweep precedent).
- **SC-005**: The overlay glyph walkthrough decodes every glyph the map can draw — coverage
  stays at 100% including the two new kinds.
- **SC-006**: A pre-feature world opened post-upgrade shows zero terrain diffs; a
  post-feature world opened by pre-feature software is refused or safely handled, never
  silently re-generated (exercised in a compatibility test).

## Assumptions

- "2-3 new kinds" is resolved to exactly **two** (marsh, sand): each exercises a different
  generation placement (moisture-adjacent vs. shoreline), and two suffice to prove the
  registry path; a third adds surface area without new information.
- Marsh and sand are **cosmetic-plus-naming only** in this feature: walkable like grass, no
  movement cost, no resources, no forage/build interactions. Gameplay affordances for them
  (if any) are future work.
- The existing shared glyph name/meaning table and its coverage tests are the seam this
  feature grows into a full registry — the legend/overlay single-source discipline already
  exists and is preserved, not reinvented.
- Terrain is regenerated from the manifest seed on every open and never persisted; that is
  why FR-006/FR-007's generation gating is required, and why no data migration of tiles is
  needed.
- The TUI is the only renderer in scope. The registry's shape should not preclude a future
  web/graphical tileset skin (Rule 4), but building one is out of scope.
- Colorblind-safe palette validation and concrete font selection remain open research gaps
  per the analysis — out of scope here; the token classing (FR-002) is the hook that future
  work plugs into.
