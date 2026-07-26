---
name: tile-registry
description: The spec-068 tile registry (internal/tui/tiles.go) — the single data table carrying every map tile's glyph, legend name, plain-language overlay meaning, classed style token, state variants, and world binding, read by renderMapGrid, the compact legend line, and the ? overlay glyph walkthrough so the three surfaces can never silently diverge
kind: component
sources:
  - internal/tui/tiles.go
  - internal/tui/views.go
  - internal/tui/help.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# Tile registry

Spec 068 (TASK-143) grows the spec-045 shared glyph table (`internal/tui/help.go`'s
retired `glyphEntry`/`mapGlyphs`) into a full **tile registry**: one data-driven
table naming, styling, and binding every drawable map tile, so a tile's presentation
lives in exactly one place no matter how many surfaces render it. The table itself
now lives in `internal/tui/tiles.go`; `help.go` keeps only the two render call
sites that read it.

## How it works

**Three render surfaces, one source.** `renderMapGrid`'s `tile()` (views.go)
resolves a tile's glyph and style from the registry; `legendGlyphLine` (help.go)
renders the compact in-game legend from it; the `?` overlay's glyph walkthrough
page renders the plain-language meaning from it. Before this feature the map's
`tile()` switch carried its own literal glyph/style pairs and `help.go` carried a
SEPARATE `mapGlyphs` table for the legend/overlay — two sources that could drift.
The registry is now the only source; `views.go`'s entire per-tile style-literal
block (`styleWater`, `styleFire`, `styleWall`, `styleGrave`, …) is deleted and
`tile()` resolves through the registry instead (FR-001/FR-002).

**Style tokens** (`styleToken`, `newToken`): a named color/emphasis definition,
classed `classSemantic16` (meaning-bearing colors on the 16 themeable ANSI slots —
water, tree, forage, den, and the agent family) or `classMaterial256` (decorative,
the fixed 256-color palette — everything else). The classing is metadata only:
every token's actual color/bold/faint/underline values are the pre-registry
literals verbatim, built in the exact property order the old literals used (bold →
faint → underline → foreground), so the rendered bytes are byte-identical to
before (FR-004) — a `tiles_identity` golden-fixture pin (below) is the gate. Named
package-level views (`styleAgent`, `styleGrave`, `styleFireDying`, `styleWallDamaged`,
…) still exist for call sites and tests that address a style directly, but they are
now derived FROM a token (e.g. `styleGrave = tokGrave.style`), never a literal of
their own. Marsh (material256, "65", a muted wet green) and sand (material256,
"180", a pale warm tan) are the feature's own two new tokens — distinct from
tree's ANSI 2 / forage's ANSI 3 and from path 137 / pile 178 / chest 136 under the
family-tint discipline.

**Registry rows** (`tileEntry`): the spec-045 glyph-key triple (`Glyph`, `Name`,
`Meaning` — `Glyph`+`Name` concatenate with no separator into the compact legend
token, e.g. "~water", "▤▩wall"; `Meaning` is the overlay walkthrough's sentence)
grown with a `Token` (the row's base style), an optional `Variants` map (a state
name → a style-only token — `"dying"` on the fire row, `"damaged"` on the wall
row), and a binding: `Terrain []worldmap.TileKind` for a static-terrain row (water,
tree, forage, rock, the quarried/`Depleted` effective kind, and — since spec 068 —
marsh/sand), or `Keys []string` for a structure/marker row addressed by a string
key (`"fire"`, `"fire_cold"`, `"shelter"`, `"oven"`, `"chest"`, `"wall_plank"`,
`"wall_stone"`, `"path"`, `"pile"`, `"den"`, `"gru"`, `"grave"`) — `KeyGlyphs`
overrides the drawn character per key when one row's `Keys` share a glyph slot
(the wall row's `▤`/`▩`). A **variant changes style only, never the glyph**
(FR-003/C5 by construction: `Variants` carries tokens, not characters) — night
dimming is a `Faint(true)` transform layered over the token-resolved style in
`tile()`, the same "state is a style transform" rule.

**Row order is the legend's frozen order** (`mapGlyphs []tileEntry`, tiles.go): the
shipped pre-068 vocabulary (water .. grave, 16 rows) keeps its exact order —
`registerTile` only appends, never reorders or removes — so the legend line and
the overlay walkthrough's existing rows are byte-stable; marsh and sand (spec 068
US2) are appended AFTER the shipped 16, growing both surfaces without touching a
single existing byte (FR-009/SC-002). `groundTile` is the one row NOT in
`mapGlyphs` — plain open ground (grass), the terrain fallback for any kind with no
row of its own, deliberately absent from the legend (the shipped legend has never
listed grass; the path row's own meaning text explains "plain ground's dim ·").

**Binding indexes** (`terrainTiles map[worldmap.TileKind]tileRef`,
`keyTiles map[string]tileRef`, `rebuildTileIndex`): derived views over the table,
rebuilt on init and on every `registerTile` call — never a second source of truth.
`terrainTile(kind)` resolves a terrain kind to its binding (falling back to
`groundTile` for an unbound kind — grass, always); `tileKey(key)` resolves a
structure/marker binding the shipped vocabulary guarantees a row for (a missing
row is impossible for shipped keys — the coverage sweep pins it, and the
`tiles_identity` golden would catch a regression instantly); `renderTileKey(key,
state)` is the softer form a caller can check `ok` against (any future structure
kind with no bound row renders invisible, exactly as before the registry existed).
`(tileRef).render(state)` draws the binding's `Glyph` in its `state`-resolved style
in one call — `tileKey("fire").render("dying")`, `tileKey(st.Kind).render("")` for
an ordinary structure, `terrainTile(gm.At(x,y))` for base terrain.

**tile()'s own priority is unchanged** (views.go): gru > agents > structures >
piles > dens > path > depleted/terrain — the registry supplies WHAT each branch
draws (leaves), never the precedence itself; the switch's own logic decides which
branch fires. A structure kind not explicitly cased (the `default` arm) now
resolves through `renderTileKey` if a row binds it, rather than staying invisible
unconditionally — a registry row reaches the map with no renderer edit (SC-002).

## Marsh and sand (the feature's shipped vocabulary growth)

Two new rows exercise the whole registry end to end: `{Glyph: "░", Name: "marsh",
Token: tokMarsh, Terrain: [worldmap.Marsh]}` and `{Glyph: "▒", Name: "sand", Token:
tokSand, Terrain: [worldmap.Sand]}` — CP437 shading characters, per the
tile-vocabulary analysis's rule that a new KIND gets a new character, never a
color-only distinction (so a 16-color terminal still tells marsh/sand apart from
plain grass and from each other). They carry no `Variants` (no state changes them)
and are plain terrain rows, resolved through `terrainTile` exactly like water or
rock. [[worldmap-generation]] is the generator that produces `worldmap.Marsh`/
`worldmap.Sand` tiles (gated by the manifest's `terrain_gen` field,
[[world-save-directory]]); this note only owns how they're DRAWN once generated.
`internal/tool/explain.go`'s `explainGlyphs` mirror gains the matching two rows
(`TestExplainGlyphsMirrorLegend` pins the two tables equal — [[grounded-feedback]]).

## Connections

[[tui-client]] hosts `Model.renderMapGrid`/`tile()`, the map region this registry's
resolution feeds, and the help overlay's glyph walkthrough page; [[worldmap-generation]]
is the generator whose `TileKind` vocabulary (including the spec-068 `Marsh`/`Sand`
additions) this registry's `Terrain` bindings key against; [[world-save-directory]]
carries the `terrain_gen` manifest field that gates whether a world's map ever
contains a marsh/sand tile for this registry to draw; [[executor]] is where
`effectiveKind`/`passable` treat marsh/sand as plain walkable ground with no
overlay of their own — a fact this registry's rendering side does not need to know,
since it draws whatever `TileKind` the map reports; [[mental-maps]] deliberately
excludes marsh/sand from its own closed vocabulary (no resource affordance to
record); [[grounded-feedback]] mirrors this registry's glyph rows into the
guardian's `explain` tool's `glyphs` fact sheet — the SAME table, so the explained
legend can never drift from the drawn one; [[village-lens]] draws its map condition
overlays (needs-critical, suppressed-mind, dying-fire) as style-only variants over
this registry's agent/fire tokens, never new glyphs of their own.

## Operational notes

The `tiles_identity` byte-identity pin (`internal/tui/tiles_identity_test.go`, spec
068 T001, C6/FR-004/SC-001) is the feature's own regression gate: a representative
LEGACY-world fixture's `renderMapGrid`/`legendGlyphLine`/glyph-walkthrough output
was captured as committed goldens BEFORE the registry refactor landed — for the
pre-existing 16-row vocabulary these bytes are a contract (a failing pin means the
renderer drifted, never the goldens); marsh/sand and any future tile are explicitly
allowed to GROW the legend/overlay past `pinnedGlyphRows` (16), append-only, since
FR-009 requires exactly that. `internal/tui/tiles_test.go` covers the registry
mechanics directly: no style literals survive outside it
(`TestTileNoStyleLiteralsOutsideRegistry`), token classing, that a `Variants` entry
never changes a row's own glyph, that a newly `registerTile`-ed row reaches the map
grid AND the legend AND the overlay with no other edit
(`TestRegisteredTileReachesMapLegendOverlay`), that the full registry decodes
cleanly, and (spec 068 specifically) that marsh/sand's glyphs are distinct from
every other row and that both reach all three surfaces
(`TestMarshSandGlyphsDistinct`/`TestMarshSandReachMapLegendOverlay`) plus a
`GenMarshSand`-generation identity pin (`TestGen2IdentityPin`).
