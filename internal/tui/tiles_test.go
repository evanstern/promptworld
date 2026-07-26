package tui

// Registry invariant tests (spec 068 T006/T007): the no-stray-literals sweep
// (C3/SC-004), the token classing rule (C4), the states-are-variants rule
// (C5), the registered-tile round trip (C1/SC-002), and full decode
// (C2/SC-005).

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/muesli/termenv"

	"github.com/evanstern/promptworld/internal/worldmap"
)

// TestTileNoStyleLiteralsOutsideRegistry (C3, SC-004 — the TASK-121
// fiction-literal sweep precedent, mechanical form): no non-test source file
// in this package except tiles.go may construct a lipgloss color from the
// registry's material tile palette. The denylist is derived from the token
// table itself, so a tile color added to the registry is automatically swept
// for strays. (Semantic-16 slots are shared with non-tile chrome — header
// families, alerts — so the material palette is the per-tile fingerprint;
// the byte-identity pin and the round-trip test hold the rest of C3's line.)
func TestTileNoStyleLiteralsOutsideRegistry(t *testing.T) {
	material := map[string]bool{}
	for _, tok := range styleTokens {
		if tok.class == classMaterial256 {
			material[tok.color] = true
		}
	}
	if len(material) == 0 {
		t.Fatal("token table has no material256 tokens — the sweep would be vacuous")
	}
	colorRe := regexp.MustCompile(`lipgloss\.Color\("(\d+)"\)`)
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") || f == "tiles.go" {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range colorRe.FindAllStringSubmatch(string(src), -1) {
			if material[m[1]] {
				t.Errorf("%s constructs tile material color %s outside the registry (C3: per-tile styles live in tiles.go's token table)", f, m[1])
			}
		}
	}
}

// TestTileStyleTokenClasses (C4): every token is classed semantic16 or
// material256; semantic16 tokens use only the 16 themeable ANSI slots (or
// the terminal's default foreground); material256 tokens use the fixed
// extended palette. Names are unique — a token is addressable.
func TestTileStyleTokenClasses(t *testing.T) {
	seen := map[string]bool{}
	for _, tok := range styleTokens {
		if tok.name == "" {
			t.Error("token with empty name")
		}
		if seen[tok.name] {
			t.Errorf("token name %q declared twice", tok.name)
		}
		seen[tok.name] = true
		switch tok.class {
		case classSemantic16:
			// "" = terminal default foreground — themeable by definition.
			if tok.color != "" && !ansiSlot(tok.color) {
				t.Errorf("semantic16 token %q uses color %q — not an ANSI 0-15 slot (C4)", tok.name, tok.color)
			}
		case classMaterial256:
			if ansiSlot(tok.color) || tok.color == "" {
				t.Errorf("material256 token %q uses color %q — the fixed 256 palette starts at 16 (C4)", tok.name, tok.color)
			}
		default:
			t.Errorf("token %q has unknown class %q", tok.name, tok.class)
		}
	}
	// Every registry row resolves through a classed token.
	for i := range mapGlyphs {
		if mapGlyphs[i].Token == nil {
			t.Errorf("registry row %q has no style token", mapGlyphs[i].Name)
		}
		for state, v := range mapGlyphs[i].Variants {
			if v == nil {
				t.Errorf("registry row %q variant %q has no style token", mapGlyphs[i].Name, state)
			}
		}
	}
	if groundTile.Token == nil {
		t.Error("ground tile has no style token")
	}
}

func ansiSlot(color string) bool {
	switch len(color) {
	case 1:
		return color[0] >= '0' && color[0] <= '9'
	case 2:
		return color[0] == '1' && color[1] >= '0' && color[1] <= '5'
	}
	return false
}

// TestTileVariantsReuseBaseGlyph (C5, FR-003): a state variant restyles its
// row's own glyph, never a different character — checked on the rendered
// output for every binding × variant.
func TestTileVariantsReuseBaseGlyph(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	stripANSI := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	check := func(where string, r tileRef) {
		base := stripANSI.ReplaceAllString(r.render(""), "")
		for state := range r.entry.Variants {
			got := stripANSI.ReplaceAllString(r.render(state), "")
			if got != base {
				t.Errorf("%s: variant %q draws %q, base draws %q — a variant may change style only (FR-003)", where, state, got, base)
			}
			if r.style(state).Render("x") == r.style("").Render("x") {
				t.Errorf("%s: variant %q renders identically to the base style — a vacuous variant", where, state)
			}
		}
	}
	for k, r := range terrainTiles {
		check("terrain "+r.entry.Name, r)
		_ = k
	}
	for key, r := range keyTiles {
		check("key "+key, r)
	}
}

// TestRegisteredTileReachesMapLegendOverlay (C1, SC-002): one appended
// registry row — glyph, name, meaning, style, binding — reaches the map
// renderer, the compact legend, and the `?` overlay walkthrough with zero
// renderer edits.
func TestRegisteredTileReachesMapLegendOverlay(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	const testKind = worldmap.TileKind(200)
	n := len(mapGlyphs)
	registerTile(tileEntry{
		Glyph:   "¤",
		Name:    "bog",
		Meaning: "a test-only bog tile",
		Token:   tokDen,
		Terrain: []worldmap.TileKind{testKind},
	})
	t.Cleanup(func() {
		mapGlyphs = mapGlyphs[:n]
		rebuildTileIndex()
	})

	// Map: a tile of the new kind at the camera's default center renders the
	// new glyph in the new row's token style.
	m := testModel(t)
	cx, cy := m.gameMap.W/2, m.gameMap.H/2
	m.gameMap.Tiles[cy*m.gameMap.W+cx] = testKind
	m.replica.Agents = nil
	grid, _ := m.renderMapGrid(10, 10)
	if !strings.Contains(grid, tokDen.style.Render("¤")) {
		t.Errorf("registered tile did not reach the map grid:\n%s", grid)
	}

	// Legend: the compact token appears, appended after the shipped set.
	if line := legendGlyphLine(); !strings.HasSuffix(line, " ¤bog") {
		t.Errorf("registered tile did not reach the compact legend: %q", line)
	}

	// Overlay: the walkthrough decodes it with its plain-language meaning.
	walkthrough := strings.Join(Model{}.helpWalkthroughLines(200), "\n")
	if !strings.Contains(walkthrough, "a test-only bog tile") {
		t.Errorf("registered tile did not reach the overlay walkthrough:\n%s", walkthrough)
	}
}

// TestTileRegistryFullDecode (C2, SC-005): every glyph the map can draw has
// a registry row with a non-empty name and meaning — every binding the
// renderer can resolve points at a decodable row, and every row is fully
// authored.
func TestTileRegistryFullDecode(t *testing.T) {
	inTable := map[*tileEntry]bool{}
	for i := range mapGlyphs {
		e := &mapGlyphs[i]
		inTable[e] = true
		if e.Glyph == "" || e.Name == "" || e.Meaning == "" {
			t.Errorf("registry row %d (%q) is not fully authored: glyph %q, name %q, meaning %q",
				i, e.Name, e.Glyph, e.Name, e.Meaning)
		}
	}
	check := func(where string, r tileRef) {
		if !inTable[r.entry] && r.entry != &groundTile {
			t.Errorf("%s resolves to an entry outside the registry table", where)
		}
		if r.glyph == "" {
			t.Errorf("%s draws an empty glyph", where)
		}
		if !strings.Contains(r.entry.Glyph, r.glyph) {
			t.Errorf("%s draws %q, which its row's legend glyph %q does not carry", where, r.glyph, r.entry.Glyph)
		}
	}
	for k, r := range terrainTiles {
		check(r.entry.Name+" (terrain)", r)
		_ = k
	}
	for key, r := range keyTiles {
		check(key+" (key)", r)
	}
	// Plain ground's "·" is deliberately not a legend row of its own; the
	// path row's meaning decodes it ("plain ground's dim ·") — pinned here
	// so the exception can never rot silently.
	ground := terrainTile(worldmap.Grass)
	if ground.entry != &groundTile {
		t.Fatal("grass no longer resolves to the ground tile")
	}
	decoded := false
	for i := range mapGlyphs {
		if strings.Contains(mapGlyphs[i].Meaning, "plain ground") {
			decoded = true
		}
	}
	if !decoded {
		t.Error("no legend row decodes plain ground's dim · (C2: the legend+overlay must decode everything the map draws)")
	}
}
