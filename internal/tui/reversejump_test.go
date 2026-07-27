package tui

// Reverse-jump rider tests (spec 086 US5, T024/T025): strip-glyph and
// roster-row clicks center the camera (real tea.MouseMsg dispatch through
// m.Update — the mouse-parity oracle's checks reuse these paths); `J` in
// villagers mode is the keyboard parity, incl. dead-villager grave
// coordinates and the nil-replica no-op; the strip's overflow marker is
// never a target.

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/evanstern/promptworld/internal/sim"
)

// reverseJumpModel: widescreen, connected, villager strip rendered, with
// villager positions spread so a jump always produces a nonzero pan.
func reverseJumpModel(t *testing.T) Model {
	t.Helper()
	m := widescreenModel(t)
	m.connected = true
	for i := range m.replica.Agents {
		m.replica.Agents[i].X = 5 + i*6
		m.replica.Agents[i].Y = 4 + i*3
	}
	return m
}

func TestStripGlyphClickCentersCamera(t *testing.T) {
	m := reverseJumpModel(t)
	_ = m.View() // renders the strip; records stripHit
	if !m.stripHit.valid {
		t.Fatal("stripHit should be valid after a widescreen render")
	}
	const i = 5
	if i >= len(m.stripHit.glyphX) {
		t.Fatalf("only %d glyphs recorded", len(m.stripHit.glyphX))
	}
	mdl, _ := m.Update(mouseLeftRelease(m.stripHit.glyphX[i], m.stripHit.originY))
	mm := mdl.(Model)
	a := mm.replica.Agents[i]
	cx, cy := mm.wandererCentroid()
	if mm.panX != a.X-cx || mm.panY != a.Y-cy {
		t.Fatalf("strip click pan = (%d,%d), want (%d,%d) — centerCameraOn(%d,%d)",
			mm.panX, mm.panY, a.X-cx, a.Y-cy, a.X, a.Y)
	}
}

func TestStripDeadVillagerClickJumpsToGrave(t *testing.T) {
	m := reverseJumpModel(t)
	m.replica.Agents[2].Dead = true // keeps X,Y — the grave
	_ = m.View()
	mdl, _ := m.Update(mouseLeftRelease(m.stripHit.glyphX[2], m.stripHit.originY))
	mm := mdl.(Model)
	a := mm.replica.Agents[2]
	cx, cy := mm.wandererCentroid()
	if mm.panX != a.X-cx || mm.panY != a.Y-cy {
		t.Fatal("clicking a † glyph must center on the grave coordinates")
	}
}

func TestStripNonGlyphClicksAreNoOps(t *testing.T) {
	m := reverseJumpModel(t)
	_ = m.View()
	// The count text (column 0) and the separator between glyphs 0 and 1.
	for _, x := range []int{0, m.stripHit.glyphX[0] + 1} {
		mdl, _ := m.Update(mouseLeftRelease(x, m.stripHit.originY))
		mm := mdl.(Model)
		if mm.panX != 0 || mm.panY != 0 {
			t.Fatalf("click at column %d moved the camera — only glyph cells are targets", x)
		}
	}
}

func TestStripOverflowMarkerIsNoTarget(t *testing.T) {
	m := reverseJumpModel(t)
	m.width = 30 // narrow enough that glyphs shed... but 30 < widescreen: strip never renders
	_ = m.View()
	if m.stripHit.valid {
		t.Fatal("the strip must not record a region in the narrow layout (it never renders there)")
	}
	// Widescreen with a pathologically small strip budget: force overflow by
	// rendering the strip view directly at a width that sheds glyphs.
	m = reverseJumpModel(t)
	_ = m.villagerStripView(16) // "8 villagers" + room for ~1 glyph + marker
	if !m.stripHit.valid {
		t.Fatal("stripHit should record the rendered prefix under overflow")
	}
	if len(m.stripHit.glyphX) >= 8 {
		t.Fatalf("overflow should shed glyphs: %d recorded", len(m.stripHit.glyphX))
	}
	// A click one column past the last rendered glyph (separator before the
	// …N marker) is a no-op.
	if len(m.stripHit.glyphX) > 0 {
		last := m.stripHit.glyphX[len(m.stripHit.glyphX)-1]
		mdl, _ := m.Update(mouseLeftRelease(last+2, m.stripHit.originY))
		if mm := mdl.(Model); mm.panX != 0 || mm.panY != 0 {
			t.Fatal("the overflow marker is not a click target — no guessing which hidden villager was meant")
		}
	}
}

func TestRosterRowClickSelectsAndCentersCamera(t *testing.T) {
	m := reverseJumpModel(t)
	m.dockTab = paneVillagers
	_ = m.View()
	if !m.rosterHit.valid {
		t.Fatal("rosterHit should be valid after rendering the villagers roster")
	}
	// Find villager 3's first band row.
	row := -1
	for r, idx := range m.rosterHit.rowAgent {
		if idx == 3 {
			row = r
			break
		}
	}
	if row < 0 {
		t.Fatal("villager 3 has no rendered roster row")
	}
	mdl, _ := m.Update(mouseLeftRelease(m.rosterHit.originX+2, m.rosterHit.originY+row))
	mm := mdl.(Model)
	if mm.villSelected != 3 {
		t.Fatalf("roster click should select row 3, got %d", mm.villSelected)
	}
	a := mm.replica.Agents[3]
	cx, cy := mm.wandererCentroid()
	if mm.panX != a.X-cx || mm.panY != a.Y-cy {
		t.Fatal("roster click should also center the camera (select + act, the chronicle click precedent)")
	}
}

func TestRosterClickNarrowSwitchesPaneToMap(t *testing.T) {
	m := reverseJumpModel(t)
	m.width = 80 // narrow
	m.active = paneVillagers
	_ = m.View()
	if !m.rosterHit.valid {
		t.Fatal("rosterHit should be valid after rendering the narrow villagers pane")
	}
	row := -1
	for r, idx := range m.rosterHit.rowAgent {
		if idx == 1 {
			row = r
			break
		}
	}
	if row < 0 {
		t.Fatal("villager 1 has no rendered roster row")
	}
	mdl, _ := m.Update(mouseLeftRelease(m.rosterHit.originX+1, m.rosterHit.originY+row))
	mm := mdl.(Model)
	if mm.active != paneMap {
		t.Fatal("a narrow roster jump must switch the active pane to the map (the jumpToSource FR-007 precedent)")
	}
	if mm.panX == 0 && mm.panY == 0 {
		t.Fatal("the narrow roster jump should have moved the camera")
	}
}

func TestVillagersKeyJCentersCamera(t *testing.T) {
	m := reverseJumpModel(t)
	m.dockTab = paneVillagers
	m.villSelected = 6
	mdl := update(m, "J")
	mm := mdl.(Model)
	a := mm.replica.Agents[6]
	cx, cy := mm.wandererCentroid()
	if mm.panX != a.X-cx || mm.panY != a.Y-cy {
		t.Fatalf("J pan = (%d,%d), want centerCameraOn(%d,%d)", mm.panX, mm.panY, a.X, a.Y)
	}

	// Detail view: J still jumps (roster and detail alike, FR-010).
	m2 := reverseJumpModel(t)
	m2.dockTab = paneVillagers
	m2.villSelected = 2
	m2.villDetail = true
	mm2 := update(m2, "J").(Model)
	if mm2.panX == 0 && mm2.panY == 0 {
		t.Fatal("J from the detail view should also jump")
	}

	// Dead villager: grave coordinates.
	m3 := reverseJumpModel(t)
	m3.dockTab = paneVillagers
	m3.villSelected = 4
	m3.replica.Agents[4].Dead = true
	mm3 := update(m3, "J").(Model)
	a3 := mm3.replica.Agents[4]
	cx3, cy3 := mm3.wandererCentroid()
	if mm3.panX != a3.X-cx3 || mm3.panY != a3.Y-cy3 {
		t.Fatal("J on a dead villager must center on the grave coordinates")
	}

	// No replica: a clean no-op, never a panic.
	m4 := widescreenModel(t)
	m4.dockTab = paneVillagers
	m4.replica = nil
	mm4 := update(m4, "J").(Model)
	if mm4.panX != 0 || mm4.panY != 0 {
		t.Fatal("J with no replica must be a no-op")
	}
	_ = sim.AgentCount
	var _ tea.Model = mm4
}
