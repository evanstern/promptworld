package tui

// The tile registry (spec 068, FR-001/FR-002): the ONE data table carrying
// every map tile's presentation — glyph, compact legend name, plain-language
// overlay meaning, style token, state-variant styles, and the world thing it
// renders. renderMapGrid's tile resolution (views.go), legendGlyphLine
// (help.go), and the `?` overlay glyph walkthrough all read this table, so a
// tile added (or re-skinned) here reaches all three surfaces with no other
// edit — the analysis's "one grid model, swappable skins" rule, grown out of
// spec 045's shared glyph table rather than built beside it.
//
// Styles resolve through NAMED TOKENS, each classed per the tile-vocabulary
// analysis's palette rule: semantic (meaning-bearing) colors live on the 16
// themeable terminal slots; material (decorative) colors on the fixed 256
// palette. The classing is metadata — colors are byte-identical to the
// pre-registry literals (FR-004; the tiles_identity pin is the gate). State
// presentation (dying fire, damaged wall, night, the agent condition
// overlays) is always a style TRANSFORM of a learned glyph, never a new
// character (FR-003/C5): a Variants entry carries a token only — it cannot
// change the glyph — and night dimming stays a Faint(true) transform applied
// over the token-resolved style in tile().

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/evanstern/promptworld/internal/worldmap"
)

// --- style tokens (FR-002, C4) ---

// styleClass is a token's palette class: semantic16 tokens carry meaning and
// live on the 16 themeable terminal color slots (ANSI 0-15); material256
// tokens are decorative and use the fixed 256 palette. The unit a future
// skin would override (TASK-121 skin precedent, Cogmind externalized-colors
// precedent).
type styleClass string

const (
	classSemantic16  styleClass = "semantic16"
	classMaterial256 styleClass = "material256"
)

// styleToken is one named color/emphasis definition. color is the terminal
// color ref ("0".."15" for semantic16, a 256-palette code for material256,
// or "" for the terminal's default foreground — themeable by definition);
// style is the lipgloss value built from the fields, once, at init.
type styleToken struct {
	name      string
	class     styleClass
	color     string
	bold      bool
	faint     bool
	underline bool
	style     lipgloss.Style
}

// styleTokens is the whole token table, appended by newToken — the sweep
// tests (C3/C4) enumerate it.
var styleTokens []*styleToken

// newToken builds a token's lipgloss style in the exact property order the
// pre-registry literals used (bold → faint → underline → foreground), so the
// rendered bytes cannot drift from them (FR-004).
func newToken(name string, class styleClass, color string, bold, faint, underline bool) *styleToken {
	s := lipgloss.NewStyle()
	if bold {
		s = s.Bold(true)
	}
	if faint {
		s = s.Faint(true)
	}
	if underline {
		s = s.Underline(true)
	}
	if color != "" {
		s = s.Foreground(lipgloss.Color(color))
	}
	t := &styleToken{name: name, class: class, color: color, bold: bold, faint: faint, underline: underline, style: s}
	styleTokens = append(styleTokens, t)
	return t
}

// The token table. Colors and emphasis are the pre-registry literals
// verbatim (views.go's deleted per-tile block); the classing follows the
// tile-vocabulary analysis (research R2): semantic-16 = water 4, tree 2,
// forage 3, den 5, the agent family (awake 2, asleep 6, dead/critical 1),
// and plain ground (default fg); everything else is material-256.
var (
	tokWater  = newToken("water", classSemantic16, "4", false, false, false)
	tokTree   = newToken("tree", classSemantic16, "2", false, false, false)
	tokForage = newToken("forage", classSemantic16, "3", false, false, false)
	tokRock   = newToken("rock", classMaterial256, "245", false, false, false)
	// tokSpent is the shared "depleted/inert" register: quarried-out rock,
	// a cold fire, a damaged wall — all faint 240, one token (the classing
	// table's "depleted/cold/damaged 240" row).
	tokSpent     = newToken("spent", classMaterial256, "240", false, true, false)
	tokDen       = newToken("den", classSemantic16, "5", false, false, false)
	tokFire      = newToken("fire", classMaterial256, "208", true, false, false)
	tokFireDying = newToken("fire-dying", classMaterial256, "202", true, false, false)
	tokShelter   = newToken("shelter", classMaterial256, "130", true, false, false)
	tokOven      = newToken("oven", classMaterial256, "166", true, false, false)
	tokPile      = newToken("pile", classMaterial256, "178", true, false, false)
	tokChest     = newToken("chest", classMaterial256, "136", true, false, false)
	tokWall      = newToken("wall", classMaterial256, "250", true, false, false)
	tokPath      = newToken("path", classMaterial256, "137", false, false, false)
	tokGru       = newToken("gru", classMaterial256, "196", true, false, false)
	tokGrave     = newToken("grave", classMaterial256, "244", false, true, false)
	// tokGround: plain grass's dim "·" — no foreground (the terminal's own
	// default, themeable by definition), faint.
	tokGround = newToken("ground", classSemantic16, "", false, true, false)

	// The agent family: the letter/case resolution stays in tile() (the
	// glyph is an agent's own initial), but the STYLES it paints with are
	// registry tokens like every other tile's (data-model "Binding").
	tokAgent           = newToken("agent", classSemantic16, "2", true, false, false)
	tokAgentAsleep     = newToken("agent-asleep", classSemantic16, "6", false, false, false)
	tokAgentDead       = newToken("agent-dead", classSemantic16, "1", true, false, false)
	tokAgentCritical   = newToken("agent-critical", classSemantic16, "1", true, false, true)
	tokAgentSuppressed = newToken("agent-suppressed", classMaterial256, "135", false, true, false)
)

// Named views of the agent-family and state tokens, kept for the rendering
// call sites and tests that address a style directly (the overlay-priority
// tests, the identity fixture). These are registry data — resolved FROM the
// token table, never literals (FR-002).
var (
	styleAgent           = tokAgent.style
	styleAsleep          = tokAgentAsleep.style
	styleAgentDead       = tokAgentDead.style
	styleAgentCritical   = tokAgentCritical.style
	styleAgentSuppressed = tokAgentSuppressed.style
	styleFire            = tokFire.style
	styleFireDying       = tokFireDying.style
	styleFireCold        = tokSpent.style
	styleWall            = tokWall.style
	styleWallDamaged     = tokSpent.style
	stylePath            = tokPath.style
	styleGrave           = tokGrave.style
)

// --- the registry rows (grown glyphEntry, spec 045 → spec 068) ---

// tileEntry is one row of the registry: the spec 045 glyph-key triple
// (Glyph+Name concatenate into the compact legend token; Meaning is the `?`
// overlay walkthrough sentence) grown with its presentation and binding
// (spec 068 US1). Variants map a state name to a style-only token — the
// variant draws the row's OWN glyph, always (FR-003/C5). Terrain/Keys bind
// the row to what it renders: worldmap tile kinds, structure kinds, or
// marker keys ("pile", "den", "gru", "fire_cold"); KeyGlyphs picks the drawn
// character per key when Glyph lists several (the wall row's "▤▩").
type tileEntry struct {
	Glyph     string
	Name      string
	Meaning   string
	Token     *styleToken
	Variants  map[string]*styleToken
	Terrain   []worldmap.TileKind
	Keys      []string
	KeyGlyphs map[string]string
}

// mapGlyphs is the registry table itself — renderMapGrid's legend line and
// the overlay's glyph walkthrough render *from* it (legendGlyphLine,
// helpWalkthroughLines in help.go), and tile() resolves glyph+style through
// the binding indexes below — one source, so the map, the compact legend,
// and the overlay can never silently diverge (spec 045 FR-005, spec 068
// FR-001). Row order is the legend's token order: the shipped vocabulary's
// order is frozen (the tiles_identity pin), new rows append.
//
// The "G" gru and "✝" grave rows carry their spec 045/044 histories — see
// git for the original row commentary; the semantics are unchanged.
var mapGlyphs = []tileEntry{
	{Glyph: "~", Name: "water", Meaning: "water — impassable to foot travel",
		Token: tokWater, Terrain: []worldmap.TileKind{worldmap.Water}},
	{Glyph: "♠", Name: "wood", Meaning: "a tree — choppable for wood",
		Token: tokTree, Terrain: []worldmap.TileKind{worldmap.Tree}},
	{Glyph: "\"", Name: "forage", Meaning: "wild forage — gatherable food",
		Token: tokForage, Terrain: []worldmap.TileKind{worldmap.Forage}},
	{Glyph: "^", Name: "rock", Meaning: "an intact rock outcrop — quarriable for stone",
		Token: tokRock, Terrain: []worldmap.TileKind{worldmap.Rock}},
	{Glyph: ",", Name: "quarried", Meaning: "a depleted outcrop — passable, already quarried",
		Token: tokSpent, Terrain: []worldmap.TileKind{worldmap.Depleted}},
	{Glyph: "ᴥ", Name: "den", Meaning: "a gru's den",
		Token: tokDen, Keys: []string{"den"}},
	{Glyph: "▲", Name: "fire", Meaning: "a lit fire — warmth, cooking",
		Token: tokFire, Keys: []string{"fire"},
		Variants: map[string]*styleToken{"dying": tokFireDying}},
	// A cold fire has always been its own learned glyph ("△", hollow), not a
	// style variant — pre-registry vocabulary, preserved verbatim.
	{Glyph: "△", Name: "cold", Meaning: "a cold fire, out of fuel",
		Token: tokSpent, Keys: []string{"fire_cold"}},
	{Glyph: "⌂", Name: "shelter", Meaning: "a built shelter",
		Token: tokShelter, Keys: []string{"shelter"}},
	{Glyph: "▣", Name: "oven", Meaning: "a built oven",
		Token: tokOven, Keys: []string{"oven"}},
	{Glyph: "%", Name: "pile", Meaning: "a ground stockpile of dropped goods",
		Token: tokPile, Keys: []string{"pile"}},
	{Glyph: "☐", Name: "chest", Meaning: "a built chest, holding goods",
		Token: tokChest, Keys: []string{"chest"}},
	{Glyph: "▤▩", Name: "wall", Meaning: "a built wall (▤ plank, ▩ stone); dim = damaged",
		Token: tokWall, Keys: []string{"wall_plank", "wall_stone"},
		KeyGlyphs: map[string]string{"wall_plank": "▤", "wall_stone": "▩"},
		Variants:  map[string]*styleToken{"damaged": tokSpent}},
	{Glyph: "·", Name: "path", Meaning: "a paved path (tan) — distinct from plain ground's dim ·",
		Token: tokPath, Keys: []string{"path"}},
	{Glyph: "G", Name: "gru", Meaning: "the gru — a predator; approach at your peril",
		Token: tokGru, Keys: []string{"gru"}},
	{Glyph: "✝", Name: "grave", Meaning: "a villager's grave — marks where a death occurred",
		Token: tokGrave, Keys: []string{"grave"}},
}

// groundTile is plain open ground — the terrain fallback for any kind
// without a row of its own (grass, since genesis). Registry data like every
// other tile, but deliberately NOT a legend row: the shipped legend has
// never listed grass (the path row's meaning decodes "plain ground's dim ·")
// and the legend's bytes are pinned.
var groundTile = tileEntry{
	Glyph: "·", Name: "ground", Meaning: "plain open ground", Token: tokGround,
}

// --- binding indexes (rebuilt whenever the table changes) ---

// tileRef is one resolved binding: the registry row plus the single
// character this binding draws (KeyGlyphs override, else the row's Glyph).
type tileRef struct {
	entry *tileEntry
	glyph string
}

var (
	terrainTiles map[worldmap.TileKind]tileRef
	keyTiles     map[string]tileRef
)

// rebuildTileIndex derives the binding indexes from the table. Called at
// init and by registerTile — the indexes are views, never a second source.
func rebuildTileIndex() {
	terrainTiles = make(map[worldmap.TileKind]tileRef)
	keyTiles = make(map[string]tileRef)
	for i := range mapGlyphs {
		e := &mapGlyphs[i]
		for _, k := range e.Terrain {
			terrainTiles[k] = tileRef{entry: e, glyph: e.Glyph}
		}
		for _, key := range e.Keys {
			g := e.Glyph
			if kg, ok := e.KeyGlyphs[key]; ok {
				g = kg
			}
			keyTiles[key] = tileRef{entry: e, glyph: g}
		}
	}
}

func init() { rebuildTileIndex() }

// registerTile appends one row to the registry and rebuilds the indexes —
// SC-002's "one data change" seam: the new row reaches the map renderer, the
// compact legend, and the overlay walkthrough with no renderer edit.
func registerTile(e tileEntry) {
	mapGlyphs = append(mapGlyphs, e)
	rebuildTileIndex()
}

// --- resolution (the renderer's read surface) ---

// terrainTile resolves a terrain kind to its registry binding; kinds with
// no row of their own are plain ground.
func terrainTile(k worldmap.TileKind) tileRef {
	if r, ok := terrainTiles[k]; ok {
		return r
	}
	return tileRef{entry: &groundTile, glyph: groundTile.Glyph}
}

// style resolves a ref's style, under a state variant when the row defines
// one. A variant changes style ONLY — the glyph stays the row's own
// (FR-003/C5 by construction: Variants carries tokens, not characters).
func (r tileRef) style(state string) lipgloss.Style {
	if state != "" {
		if tok, ok := r.entry.Variants[state]; ok {
			return tok.style
		}
	}
	return r.entry.Token.style
}

// render draws the binding's glyph in its (possibly variant) style.
func (r tileRef) render(state string) string {
	return r.style(state).Render(r.glyph)
}

// renderTileKey resolves a structure/marker binding key and draws it; ok is
// false when no row binds the key (the caller keeps its own fallback).
func renderTileKey(key, state string) (string, bool) {
	r, ok := keyTiles[key]
	if !ok {
		return "", false
	}
	return r.render(state), true
}

// tileKey resolves a binding key the shipped vocabulary guarantees a row
// for; a missing row falls back to plain ground — impossible for shipped
// keys (the coverage sweep pins every drawable binding to a row), and
// instantly visible in the identity pin if it ever regresses.
func tileKey(key string) tileRef {
	if r, ok := keyTiles[key]; ok {
		return r
	}
	return tileRef{entry: &groundTile, glyph: groundTile.Glyph}
}
