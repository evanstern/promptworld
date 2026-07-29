---
name: tile-registry
description: The spec-068 tile registry (internal/tui/tiles.go) — the single data table carrying every map tile's glyph, legend name, plain-language overlay meaning, classed style token, state variants, and world binding, read by renderMapGrid, the compact legend line, and the ? overlay glyph walkthrough so the three surfaces can never silently diverge
kind: component
sources:
  - internal/tui/tiles.go
  - internal/tui/views.go
  - internal/tui/help.go
verified_against: 72f82f41f7aa2e345572105894cd0fb7c02fc0aa
---

# Tile registry

Spec 068 (TASK-143) grows the spec-045 shared glyph table (`internal/tui/help.go`'s
retired `glyphEntry`/`mapGlyphs`) into a full **tile registry**: one data-driven
table naming, styling, and binding every drawable map tile, so a tile's presentation
lives in exactly one place no matter how many surfaces render it. The table now
lives in `internal/tui/tiles.go`; `help.go` keeps only the two render call sites
that read it.

## How it works

**Three render surfaces, one source.** `renderMapGrid`'s `tile()` (views.go)
resolves a tile's glyph and style from the registry; `legendGlyphLine` (help.go)
renders the compact in-game legend from it; the `?` overlay's glyph walkthrough
renders the plain-language meaning from it. Previously `tile()` carried its
own literal glyph/style pairs and `help.go` a separate `mapGlyphs` table for
legend/overlay — two sources that could drift. The registry is now the only
source; `views.go`'s per-tile style-literal block (`styleWater`,
`styleFire`, `styleWall`, `styleGrave`, …) is deleted (FR-001/FR-002).

**Style tokens** (`styleToken`, `newToken`): a named color/emphasis definition,
classed `classSemantic16` (meaning-bearing colors on the 16 themeable ANSI slots
— water, tree, forage, den, the agent family) or `classMaterial256` (decorative,
the fixed 256-color palette — everything else). Classing is metadata only:
every token's color/bold/faint/underline values are the pre-registry literals
verbatim, in the old literals' exact property order (bold → faint → underline
→ foreground), so rendered bytes stay byte-identical to before (FR-004) — the
`tiles_identity` golden pin (below) is the gate. Named package-level views
(`styleAgent`, `styleGrave`, `styleFireDying`, `styleWallDamaged`, …) still
exist for direct-addressing call sites/tests but are now derived FROM a token
(`styleGrave = tokGrave.style`), never a literal of their own. Marsh
(material256 "65", muted wet green) and sand (material256 "180", pale warm
tan) are the feature's own two new tokens — distinct from tree's ANSI 2 /
forage's ANSI 3 and path 137 / pile 178 / chest 136 under the family-tint
discipline.

**Registry rows** (`tileEntry`): the spec-045 glyph-key triple (`Glyph`, `Name`,
`Meaning` — `Glyph`+`Name` concatenate with no separator into the compact
legend token, e.g. "~water", "▤▩wall"; `Meaning` is the overlay's sentence)
grown with a `Token` (base style), an optional `Variants` map (state name →
style-only token — `"dying"` on fire, `"damaged"` on wall), and a binding:
`Terrain []worldmap.TileKind` for a static-terrain row (water, tree, forage,
rock, the quarried/`Depleted` kind, and — since spec 068 — marsh/sand), or
`Keys []string` for a structure/marker row (`"fire"`, `"fire_cold"`,
`"shelter"`, `"oven"`, `"chest"`, `"wall_plank"`, `"wall_stone"`, `"path"`,
`"pile"`, `"den"`, `"gru"`, `"grave"`, spec 077's `"stranger"`, and —
spec 084 — `"designation_site"`/`"designation_line"`/`"designation_zone"`
(the guardian plan marks `◇`/`┄`/`◦`, one shared semantic-16 `designation`
token; rendered by `renderMapGrid` from `State.Designations`, ACTIVE only,
beneath every real entity — [[guardian-designations]]),
the night trickster's appended `S` row, violet 135 bold beside the gru's
red) — `KeyGlyphs` overrides the drawn
character per key when one row's `Keys` share a glyph slot (the wall row's
`▤`/`▩`). A **variant changes style only, never the glyph** (FR-003/C5 by
construction) — night dimming is a `Faint(true)` transform layered over the
token-resolved style in `tile()`, the same "state is a style transform" rule.

**Row order is the legend's frozen order** (`mapGlyphs []tileEntry`, tiles.go):
the shipped pre-068 vocabulary (water .. grave, 16 rows) keeps its exact
order — `registerTile` only appends, never reorders or removes — so the
legend/overlay's existing rows stay byte-stable; marsh/sand (spec 068 US2)
are appended AFTER the shipped 16 (FR-009/SC-002). `groundTile` is the one
row NOT in `mapGlyphs` — plain open ground (grass), the terrain fallback for
any kind with no row of its own, deliberately absent from the legend.

**Binding indexes** (`terrainTiles map[worldmap.TileKind]tileRef`,
`keyTiles map[string]tileRef`, `rebuildTileIndex`): derived views over the
table, rebuilt on init and every `registerTile` call — never a second source
of truth. `terrainTile(kind)` resolves a terrain kind (falling back to
`groundTile` for an unbound kind — grass, always); `tileKey(key)` resolves a
structure/marker binding the shipped vocabulary guarantees a row for; the
softer `renderTileKey(key, state)` lets a caller check `ok` (any future
structure kind with no bound row renders invisible, as before). `(tileRef).
render(state)` draws the binding's `Glyph` in its `state`-resolved style in
one call.

**tile()'s own priority** (views.go): gru > stranger (spec 077 — shares the
gru's tier; the gru wins a shared tile) > agents > structures > piles > dens
> path > depleted/terrain — the registry supplies WHAT each branch draws,
never the precedence itself. A structure kind not explicitly cased (the
`default` arm) now resolves through `renderTileKey` if a row binds it,
rather than staying invisible unconditionally (SC-002).

## Marsh and sand (the feature's shipped vocabulary growth)

Two new rows exercise the whole registry end to end: `{Glyph: "░", Name: "marsh",
Token: tokMarsh, Terrain: [worldmap.Marsh]}` and `{Glyph: "▒", Name: "sand", Token:
tokSand, Terrain: [worldmap.Sand]}` — CP437 shading characters, per the
tile-vocabulary rule that a new KIND gets a new character, never a color-only
distinction (so a 16-color terminal still tells marsh/sand apart from plain
grass and each other). They carry no `Variants` (no state changes them)
and are plain terrain rows, resolved through `terrainTile` exactly like water or
rock. [[worldmap-generation]] is the generator that produces `worldmap.Marsh`/
`worldmap.Sand` tiles (gated by the manifest's `terrain_gen` field,
[[world-save-directory]]); this note only owns how they're DRAWN once generated.
`internal/tool/explain.go`'s `explainGlyphs` mirror gains the matching rows
(`TestExplainGlyphsMirrorLegend` pins the two tables equal — [[grounded-feedback]]).

## Connections

[[tui-client]] hosts `Model.renderMapGrid`/`tile()` and the help overlay's
glyph walkthrough this registry feeds; [[worldmap-generation]] is the
generator whose `TileKind` vocabulary (incl. spec-068's `Marsh`/`Sand`) this
registry's `Terrain` bindings key against; [[world-save-directory]] carries
the `terrain_gen` manifest field gating whether a map ever contains
marsh/sand; [[executor]] treats marsh/sand as plain walkable ground with no
overlay of its own; [[mental-maps]] excludes marsh/sand from its closed
vocabulary; [[grounded-feedback]] mirrors this registry's glyph rows into
the guardian's `explain` tool so the explained legend can never drift from
the drawn one; [[village-lens]] draws its map condition overlays as
style-only variants over this registry's tokens, never new glyphs.

## Operational notes

The `tiles_identity` byte-identity pin (`internal/tui/tiles_identity_test.go`,
spec 068 T001, C6/FR-004/SC-001) is the feature's regression gate: a
LEGACY-world fixture's `renderMapGrid`/`legendGlyphLine`/glyph-walkthrough
output was captured as committed goldens BEFORE the registry refactor —
a failing pin means the renderer drifted, never the goldens; marsh/sand and
any future tile may GROW the legend/overlay past `pinnedGlyphRows` (16),
append-only (FR-009). `internal/tui/tiles_test.go` covers the mechanics
directly: no style literals survive outside the registry, token classing,
that a `Variants` entry never changes a row's own glyph, that a newly
`registerTile`-ed row reaches the map grid AND the legend AND the overlay
with no other edit, that the registry decodes cleanly, and (spec 068) that
marsh/sand's glyphs are distinct from every other row and both reach all
three surfaces, plus a `GenMarshSand` generation identity pin.
