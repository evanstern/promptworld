package guardian

// survey_site (spec 084 US1, FR-001): the guardian's free, deterministic site
// fact sheet — the explain + spec-059-targeting-digest pattern pointed at
// planning. The sheet assembles TURN-SIDE (the tool package has no world
// state to draw from — the buildTargetingDigest rationale) from the
// absorb-mirrored structure snapshot and the static worldmap: a pure function
// of (args, mirror, map) with fixed iteration orders and no clock reads, so
// identical calls return byte-identical text. Effect Read, Gate None: no
// charge moves, no event lands, and the loop driver's read exemption keeps it
// off the turn's one-act budget ("surveying is looking, not an act").

import (
	"context"
	"fmt"
	"strings"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/toolloop"
	"github.com/evanstern/promptworld/internal/worldmap"
)

const (
	surveyDefaultRadius = 4
	surveyMinRadius     = 1
	surveyMaxRadius     = 8
)

// surveyTerrainNames is the fixed terrain-mix rendering order — generated
// kinds only (Depleted is an effective-kind overlay the static map never
// holds). Fixed order = deterministic sheet bytes.
var surveyTerrainNames = []struct {
	kind worldmap.TileKind
	name string
}{
	{worldmap.Grass, "grass"},
	{worldmap.Water, "water"},
	{worldmap.Tree, "tree"},
	{worldmap.Forage, "forage"},
	{worldmap.Rock, "rock"},
	{worldmap.Marsh, "marsh"},
	{worldmap.Sand, "sand"},
}

// handleSurvey serves the read-only survey_site tool (the handleExplain
// dispatch shape): always read_ok — an out-of-bounds site returns an honest
// in-fiction miss naming the world bounds (the explain unknown-topic shape),
// which IS the data the model asked for, never a hard error.
func (mt *Guardian) handleSurvey(d *turnDispatch) toolloop.Handler {
	return func(_ context.Context, call llm.ToolCall) toolloop.Outcome {
		x, okX := argInt(call.Args, "x")
		y, okY := argInt(call.Args, "y")
		if !okX || !okY {
			return toolloop.Outcome{Verdict: toolloop.VerdictReadOK,
				ResultForModel: "name the site: survey_site needs both x and y"}
		}
		radius, okR := argInt(call.Args, "radius")
		if !okR {
			radius = surveyDefaultRadius
		}
		if radius < surveyMinRadius {
			radius = surveyMinRadius
		}
		if radius > surveyMaxRadius {
			radius = surveyMaxRadius
		}
		return toolloop.Outcome{Verdict: toolloop.VerdictReadOK, ResultForModel: mt.buildSurveySheet(x, y, radius)}
	}
}

// buildSurveySheet assembles the deterministic fact sheet: terrain-kind mix
// within the (2r+1)-square around the center (clipped to the world), nearest
// water/tree/rock Manhattan distances from the center (whole-map scan,
// row-major tie-break), structures present (kind + tile, mirror order = the
// replica's slice order), and a passability summary (walkable terrain minus
// tiles blocked by standing walls).
func (mt *Guardian) buildSurveySheet(x, y, radius int) string {
	m := mt.m
	if m == nil {
		return "I cannot see the land of this world"
	}
	if !m.InBounds(x, y) {
		return fmt.Sprintf("that site lies beyond the world — the land runs from (0,0) to (%d,%d)", m.W-1, m.H-1)
	}

	// The mirrored structure snapshot, read once under stateMu (the
	// landMiracle probe discipline) — the sheet never touches the replica.
	mt.stateMu.Lock()
	structXY := append([][2]int(nil), mt.structXY...)
	structKind := append([]string(nil), mt.structKind...)
	mt.stateMu.Unlock()

	x0, y0 := max(0, x-radius), max(0, y-radius)
	x1, y1 := min(m.W-1, x+radius), min(m.H-1, y+radius)

	// Terrain mix + terrain passability, one clipped-box walk.
	counts := map[worldmap.TileKind]int{}
	tiles, walkable := 0, 0
	for ty := y0; ty <= y1; ty++ {
		for tx := x0; tx <= x1; tx++ {
			tiles++
			counts[m.At(tx, ty)]++
			if m.Passable(tx, ty) {
				walkable++
			}
		}
	}

	// Standing walls inside the box block otherwise-walkable tiles.
	walled := 0
	var structLines []string
	for i := range structXY {
		sx, sy := structXY[i][0], structXY[i][1]
		if sx < x0 || sx > x1 || sy < y0 || sy > y1 {
			continue
		}
		structLines = append(structLines, fmt.Sprintf("%s at (%d,%d)", structKind[i], sx, sy))
		if structKind[i] == "wall_plank" || structKind[i] == "wall_stone" {
			walled++
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Survey of (%d,%d), radius %d (%d tiles):\n", x, y, radius, tiles)
	var mix []string
	for _, tn := range surveyTerrainNames {
		if n := counts[tn.kind]; n > 0 {
			mix = append(mix, fmt.Sprintf("%d %s", n, tn.name))
		}
	}
	fmt.Fprintf(&b, "- terrain: %s\n", strings.Join(mix, ", "))
	writeNearest := func(name string, kind worldmap.TileKind) {
		if d, nx, ny, ok := nearestKindFrom(m, x, y, kind); ok {
			fmt.Fprintf(&b, "- nearest %s: (%d,%d), %d tiles away\n", name, nx, ny, d)
		} else {
			fmt.Fprintf(&b, "- nearest %s: none in this world\n", name)
		}
	}
	writeNearest("water", worldmap.Water)
	writeNearest("tree", worldmap.Tree)
	writeNearest("rock", worldmap.Rock)
	if len(structLines) > 0 {
		fmt.Fprintf(&b, "- structures within: %s\n", strings.Join(structLines, "; "))
	} else {
		b.WriteString("- structures within: none\n")
	}
	fmt.Fprintf(&b, "- passability: %d of %d tiles are walkable terrain", walkable, tiles)
	if walled > 0 {
		fmt.Fprintf(&b, ", %d further blocked by standing walls", walled)
	}
	b.WriteString("\n")
	return b.String()
}

// nearestKindFrom scans the whole map for the nearest tile of kind by
// Manhattan distance from (x,y), row-major tie-break (deterministic).
func nearestKindFrom(m *worldmap.Map, x, y int, kind worldmap.TileKind) (dist, nx, ny int, ok bool) {
	best := -1
	for ty := 0; ty < m.H; ty++ {
		for tx := 0; tx < m.W; tx++ {
			if m.At(tx, ty) != kind {
				continue
			}
			d := absInt(tx-x) + absInt(ty-y)
			if best < 0 || d < best {
				best, nx, ny = d, tx, ty
			}
		}
	}
	if best < 0 {
		return 0, 0, 0, false
	}
	return best, nx, ny, true
}

// absInt is the sheet's Manhattan-distance helper.
func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
