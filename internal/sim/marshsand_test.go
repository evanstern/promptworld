package sim

// Spec 068 T013 (C13, FR-008): marsh and sand behave as open walkable ground
// in the sim — passable, never buildable, no overlays, no resource facts —
// and every agent-facing surface names them correctly, never by a fallback
// label.

import (
	"testing"

	"github.com/evanstern/promptworld/internal/worldmap"
)

// marshSandFixture returns a GenMarshSand state plus one marsh and one sand
// coordinate from its map.
func marshSandFixture(t *testing.T) (s *State, marshX, marshY, sandX, sandY int) {
	t.Helper()
	m := worldmap.GenerateV(42, 64, 64, worldmap.GenMarshSand)
	s = NewState(42, m)
	marshX, sandX = -1, -1
	for y := 0; y < m.H; y++ {
		for x := 0; x < m.W; x++ {
			switch m.At(x, y) {
			case worldmap.Marsh:
				if marshX < 0 {
					marshX, marshY = x, y
				}
			case worldmap.Sand:
				if sandX < 0 {
					sandX, sandY = x, y
				}
			}
		}
	}
	if marshX < 0 || sandX < 0 {
		t.Fatal("fixture map carries no marsh or sand (the worldmap presence test gates this)")
	}
	return s, marshX, marshY, sandX, sandY
}

// TestMarshSandWalkableNotBuildable (C9/C13): both kinds pass the sim's own
// passable() (the second passability function beside worldmap.Passable) and
// refuse buildSite.
func TestMarshSandWalkableNotBuildable(t *testing.T) {
	s, mx, my, sx, sy := marshSandFixture(t)
	if !passable(s.m, s, mx, my) {
		t.Error("marsh must be walkable like grass (FR-008)")
	}
	if !passable(s.m, s, sx, sy) {
		t.Error("sand must be walkable like grass (FR-008)")
	}
	if buildSite(s.m, s, mx, my) {
		t.Error("marsh must not be a build site")
	}
	if buildSite(s.m, s, sx, sy) {
		t.Error("sand must not be a build site")
	}
}

// TestMarshSandEffectiveKindStable (C13): no overlay ever remaps the new
// kinds — their effective kind is their generated kind.
func TestMarshSandEffectiveKindStable(t *testing.T) {
	s, mx, my, sx, sy := marshSandFixture(t)
	// Overlay state at those coords must be inert for the new kinds even if
	// present (nothing produces it; belt-and-braces the merge order).
	s.Cleared = append(s.Cleared, Point{X: mx, Y: my})
	s.Quarried = append(s.Quarried, Point{X: sx, Y: sy})
	if got := effectiveKind(s.m, s, mx, my); got != worldmap.Marsh {
		t.Errorf("marsh effective kind = %d, want Marsh", got)
	}
	if got := effectiveKind(s.m, s, sx, sy); got != worldmap.Sand {
		t.Errorf("sand effective kind = %d, want Sand", got)
	}
}

// TestFeatureDescNamesMarshSand (C13, FR-008): the situated-memory feature
// description — the surface look-cursor text and memory grammar read — names
// the new kinds, never an empty/fallback label.
func TestFeatureDescNamesMarshSand(t *testing.T) {
	s, mx, my, sx, sy := marshSandFixture(t)
	if got := featureDesc(s, mx, my, ""); got != "the marsh" {
		t.Errorf("featureDesc(marsh) = %q, want %q", got, "the marsh")
	}
	if got := featureDesc(s, sx, sy, ""); got != "the sand flat" {
		t.Errorf("featureDesc(sand) = %q, want %q", got, "the sand flat")
	}
}

// TestRemoveTerrainRejectsMarshSand (C13): the terrain-removal miracle
// treats the new kinds like grass/water — not removable terrain, a clear
// refusal rather than a mis-handled overlay.
func TestRemoveTerrainRejectsMarshSand(t *testing.T) {
	s, mx, my, sx, sy := marshSandFixture(t)
	for _, c := range [][2]int{{mx, my}, {sx, sy}} {
		err := applyMiracleErr(s, 40, "guardian.entity_removed",
			EntityRemovedPayload{Class: "terrain", X: c[0], Y: c[1], Gratis: true})
		if err == nil {
			t.Errorf("(%d,%d): removing marsh/sand must refuse", c[0], c[1])
		}
	}
}
