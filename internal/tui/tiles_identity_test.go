package tui

// The byte-identity pin (spec 068 T001, C6/FR-004/SC-001): a representative
// legacy-world fixture rendered through renderMapGrid, legendGlyphLine, and
// the `?` overlay glyph walkthrough, with the exact output bytes committed
// as goldens in testdata/ BEFORE the tile-registry refactor landed. For the
// pre-existing vocabulary these bytes are a contract: if a later change
// breaks this pin, the change is wrong, not the pin.
//
// New vocabulary (marsh/sand, and any future tile) is allowed to GROW the
// legend and overlay — the spec's own FR-009 requires it — so the legend and
// overlay assertions pin the pre-existing vocabulary's bytes exactly and
// admit only append-only growth from registry rows added after the pinned
// ones (pinnedGlyphRows below). The map-grid goldens are a legacy fixture
// and therefore exact forever: legacy generation can never draw a new glyph.
//
// The -update-tile-goldens flag exists ONLY for a future deliberate,
// spec-ratified vocabulary redesign. It was run exactly once, against the
// pre-068 renderer, to mint the goldens. Never re-run it to "fix" a failing
// pin — a failing pin means the renderer's bytes drifted.

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/muesli/termenv"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/worldmap"
)

var updateTileGoldens = flag.Bool("update-tile-goldens", false,
	"rewrite testdata/tiles_identity_*.golden from the current renderer (deliberate vocabulary redesigns only — see file comment)")

// pinnedGlyphRows is the size of the shipped glyph vocabulary at pin time
// (pre-068: water..grave, 16 rows). Registry rows appended beyond this index
// are post-pin vocabulary: they may grow the legend/overlay, append-only,
// but the first pinnedGlyphRows rows' bytes are frozen by the goldens.
const pinnedGlyphRows = 16

func goldenPath(name string) string {
	return filepath.Join("testdata", name+".golden")
}

// pinGolden compares got against the committed golden bytes, writing them
// only under -update-tile-goldens.
func pinGolden(t *testing.T, name, got string) {
	t.Helper()
	path := goldenPath(name)
	if *updateTileGoldens {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing golden %s (T001 mints it once with -update-tile-goldens): %v", path, err)
	}
	if got != string(want) {
		t.Errorf("%s drifted from its pinned bytes (C6/SC-001)\n--- pinned ---\n%q\n--- got ---\n%q", name, string(want), got)
	}
}

func readGolden(t *testing.T, name string) string {
	t.Helper()
	want, err := os.ReadFile(goldenPath(name))
	if err != nil {
		t.Fatalf("missing golden %s: %v", goldenPath(name), err)
	}
	return string(want)
}

// tilesIdentityFixture builds the representative replica the pin renders: a
// LEGACY-generation map (the pre-068 Generate path, pinned by
// TestLegacyGenerationHashPin) carrying every glyph and state the shipped
// vocabulary can draw — agents awake/asleep/dead/critical/suppressed, a
// grave-bearing corpse, fire lit/dying/cold, walls intact/damaged (plank and
// stone), shelter, oven, chest, pile, den, path, quarried, and the gru — all
// inside the deterministic 24×16 viewport the tests render.
func tilesIdentityFixture(t *testing.T) Model {
	t.Helper()
	m := testModel(t)
	// The fixture is a pre-068 world: bind the map to the LEGACY generation
	// path explicitly (Generate's no-version signature), so the pin stays a
	// legacy fixture even after `promptworld new` starts stamping newer
	// terrain generations onto freshly created worlds.
	gm := worldmap.Generate(42, 64, 64)
	m.gameMap = gm
	m.replica = sim.NewState(42, gm)
	m.replica.Tick = 1000

	// Living agents: the centroid of (12..15, 12) is (13,12), so a 24×16
	// viewport spans x∈[1,24], y∈[4,19] — everything below sits inside it.
	m.replica.Agents = []sim.Agent{
		{Name: "Ash", X: 12, Y: 12, Needs: healthyNeeds},                                               // awake: "A"
		{Name: "Bo", X: 13, Y: 12, Asleep: true, Needs: healthyNeeds},                                  // asleep: "b"
		{Name: "Cyn", X: 14, Y: 12, Needs: sim.Needs{Health: 1000, Food: 1, Rest: 1000, Warmth: 1000}}, // needs-critical
		{Name: "Dot", X: 15, Y: 12, Needs: healthyNeeds},                                               // suppressed-mind (trace below)
		{Name: "Eve", X: 16, Y: 12, Dead: true, Needs: sim.Needs{}},                                    // dead on a grave → "✝"
		{Name: "Fyn", X: 17, Y: 12, Dead: true, Needs: sim.Needs{}},                                    // dead, no grave → "†"
	}
	// Dot's latest chain is a router suppression (spec 060 US2 AS2).
	m.applyEvent(outcomeEvent(1, "meeting-3-900", "meeting", 3, sim.OutcomeSuppressed, "budget exhausted"))

	dying := m.replica.Tick + m.replica.RefuelDyingBelow()/2 // lit, inside the dying window
	m.replica.Structures = []sim.Structure{
		{Kind: "fire", X: 12, Y: 14, FuelUntil: 1_000_000}, // lit "▲"
		{Kind: "fire", X: 13, Y: 14, FuelUntil: dying},     // dying "▲"
		{Kind: "fire", X: 14, Y: 14, FuelUntil: 500},       // cold "△"
		{Kind: "shelter", X: 15, Y: 14},
		{Kind: "oven", X: 16, Y: 14},
		{Kind: "chest", X: 17, Y: 14, Owner: 0},
		{Kind: "grave", X: 16, Y: 12}, // shares Eve's tile — the body becomes the grave
		{Kind: "wall_plank", X: 12, Y: 16, HP: sim.WallMaxHP("wall_plank")},
		{Kind: "wall_plank", X: 13, Y: 16, HP: 1}, // damaged → dim
		{Kind: "wall_stone", X: 14, Y: 16, HP: sim.WallMaxHP("wall_stone")},
		{Kind: "path", X: 15, Y: 16},
	}
	m.replica.Piles = []sim.Pile{{X: 16, Y: 16, Wood: 2}}
	m.replica.Quarried = append(m.replica.Quarried, sim.Point{X: 17, Y: 16})
	m.gameMap.Dens = append(m.gameMap.Dens, worldmap.Point{X: 18, Y: 16})
	m.replica.Gru = &sim.Gru{X: 19, Y: 16}
	return m
}

// TestTilesIdentityPin is the C6 pin proper: map grid (day and night), the
// full renderMapGrid legend line, the compact glyph line, and the overlay's
// glyph walkthrough, byte-compared against the committed goldens.
func TestTilesIdentityPin(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	m := tilesIdentityFixture(t)

	// --- map grid: a legacy fixture draws only pre-existing vocabulary, so
	// these goldens are exact forever.
	grid, legendDay := m.renderMapGrid(24, 16)
	pinGolden(t, "tiles_identity_grid_day", grid)

	m.replica.Night = true
	nightGrid, legendNight := m.renderMapGrid(24, 16)
	pinGolden(t, "tiles_identity_grid_night", nightGrid)

	// --- compact glyph line (legendGlyphLine): the pre-existing vocabulary's
	// bytes are frozen; growth is legal only as appended tokens belonging to
	// registry rows added after the pinned ones (FR-009 requires new tiles to
	// reach the legend — through this one seam, append-only).
	glyphLine := legendGlyphLine()
	var pinnedLine string
	if *updateTileGoldens {
		pinGolden(t, "tiles_identity_glyphline", glyphLine)
		pinnedLine = glyphLine
	} else {
		pinnedLine = readGolden(t, "tiles_identity_glyphline")
	}
	if !strings.HasPrefix(glyphLine, pinnedLine) {
		t.Errorf("legendGlyphLine no longer starts with the pinned pre-existing vocabulary bytes\n--- pinned ---\n%q\n--- got ---\n%q", pinnedLine, glyphLine)
	} else {
		var grown strings.Builder
		for _, g := range mapGlyphs[minInt(pinnedGlyphRows, len(mapGlyphs)):] {
			grown.WriteString(" " + g.Glyph + g.Name)
		}
		if rest := strings.TrimPrefix(glyphLine, pinnedLine); rest != grown.String() {
			t.Errorf("legendGlyphLine grew with something other than the appended registry rows: suffix %q, want %q", rest, grown.String())
		}
	}

	// --- full map legend line: everything around the vocabulary (phase,
	// camera coords, notes, pile/chest inspection) is pinned exactly; the
	// embedded glyph line is normalized back to its pinned form so legal
	// vocabulary growth doesn't shadow a drift elsewhere in the line.
	normalize := func(legend string) string {
		return strings.Replace(legend, glyphLine, pinnedLine, 1)
	}
	pinGolden(t, "tiles_identity_maplegend_day", normalize(legendDay))
	pinGolden(t, "tiles_identity_maplegend_night", normalize(legendNight))

	// --- `?` overlay glyph walkthrough: pinned exactly, modulo the appended
	// rows' own lines (filtered out by exact formatted-line match).
	lines := Model{}.helpWalkthroughLines(200)
	grownLines := map[string]bool{}
	for _, g := range mapGlyphs[minInt(pinnedGlyphRows, len(mapGlyphs)):] {
		grownLines[clipLine(fmt.Sprintf("%-4s %s", g.Glyph, g.Meaning), 200)] = true
	}
	var kept []string
	for _, l := range lines {
		if !grownLines[l] {
			kept = append(kept, l)
		}
	}
	pinGolden(t, "tiles_identity_overlay", strings.Join(kept, "\n"))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
