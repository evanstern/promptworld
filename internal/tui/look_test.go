package tui

// Spec 074-look-cursor tests: mode entry/exit, cursor movement + camera
// push/snap, the esc-release chain (SC-004), the dock-borrow's state
// preservation (data-model.md invariant 4), the TILE view's fixed hierarchy
// (AC #9) and registry round-trip (SC-002), the env header (FR-007), mouse
// parity (US4), and the fixed-geometry pin (SC-003).

import (
	"encoding/json"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// --- entry / exit (US1 AS1, FR-001, edge cases) ---

func TestLookTogglesOnAndOff(t *testing.T) {
	m := widescreenModel(t)
	var mdl tea.Model = m
	mdl = update(mdl, "v")
	mm := mdl.(Model)
	if !mm.lookActive {
		t.Fatal("'v' must enter the look-cursor mode")
	}
	cx, cy := mm.wandererCentroid()
	if mm.lookX != cx || mm.lookY != cy {
		t.Errorf("cursor should spawn at the camera center (%d,%d), got (%d,%d)", cx, cy, mm.lookX, mm.lookY)
	}
	mdl = update(mdl, "v")
	mm = mdl.(Model)
	if mm.lookActive {
		t.Fatal("'v' again must exit the mode")
	}
	if mm.panX != 0 || mm.panY != 0 {
		t.Error("exiting must resume centroid-following (pan reset to 0,0)")
	}
}

func TestLookNoWorldStateNoOp(t *testing.T) {
	m := widescreenModel(t)
	m.gameMap = nil
	mdl := update(m, "v")
	if mdl.(Model).lookActive {
		t.Fatal("'v' with no gameMap must be a strict no-op (the x-key precedent)")
	}
}

func TestLookNoOpWhileSoloOrConsole(t *testing.T) {
	t.Run("solo", func(t *testing.T) {
		m := widescreenModel(t)
		m.solo = true
		mdl := update(m, "v")
		if mdl.(Model).lookActive {
			t.Fatal("'v' must not enter the mode while the dock is solo'd (map isn't visible)")
		}
	})
	t.Run("console", func(t *testing.T) {
		m := widescreenModel(t)
		m.console = true
		mdl := update(m, "v")
		if mdl.(Model).lookActive {
			t.Fatal("'v' must not enter the mode while the console owns the body")
		}
	})
	t.Run("narrow non-map pane", func(t *testing.T) {
		m := testModel(t)
		m.active = paneVillagers
		mdl := update(m, "v")
		if mdl.(Model).lookActive {
			t.Fatal("'v' must not enter the mode in narrow unless the map pane is active")
		}
	})
}

// --- movement + clamp + jump (US1 AS2, edge cases) ---

func TestLookMovementClampAndJump(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = nil // keep the centroid (and thus the spawn point) at map center
	var mdl tea.Model = m
	mdl = update(mdl, "v")
	mm := mdl.(Model)
	x0, y0 := mm.lookX, mm.lookY

	mdl = update(mdl, "l")
	if mdl.(Model).lookX != x0+1 || mdl.(Model).lookY != y0 {
		t.Errorf("'l' should move right by 1: got (%d,%d)", mdl.(Model).lookX, mdl.(Model).lookY)
	}
	mdl = update(mdl, "right")
	if mdl.(Model).lookX != x0+2 {
		t.Errorf("right-arrow should also move right by 1: got x=%d", mdl.(Model).lookX)
	}
	mdl = update(mdl, "L")
	if mdl.(Model).lookX != x0+10 {
		t.Errorf("'L' should jump right by 8: got x=%d", mdl.(Model).lookX)
	}
	mdl = update(mdl, "h")
	mdl = update(mdl, "H")
	if mdl.(Model).lookX != x0+1 {
		t.Errorf("h then H (jump left 8): got x=%d, want %d", mdl.(Model).lookX, x0+1)
	}
}

func TestLookCursorClampsToWorldEdgeNeverWraps(t *testing.T) {
	m := widescreenModel(t)
	var mdl tea.Model = update(m, "v")
	mm := mdl.(Model)
	mm.lookX, mm.lookY = 0, 0
	mdl = mm
	mdl = update(mdl, "H") // shift-jump left 8 from x=0 must clamp, never wrap
	if got := mdl.(Model).lookX; got != 0 {
		t.Errorf("shift-jump past the left edge must clamp to 0, got %d", got)
	}
	mdl = update(mdl, "K")
	if got := mdl.(Model).lookY; got != 0 {
		t.Errorf("shift-jump past the top edge must clamp to 0, got %d", got)
	}

	w := mm.gameMap.W
	hgt := mm.gameMap.H
	mm2 := mdl.(Model)
	mm2.lookX, mm2.lookY = w-1, hgt-1
	mdl = mm2
	mdl = update(mdl, "L")
	if got := mdl.(Model).lookX; got != w-1 {
		t.Errorf("shift-jump past the right edge must clamp to %d, got %d", w-1, got)
	}
}

func TestLookArrowsMoveCursorNeverFreePan(t *testing.T) {
	// AS1.5: arrow keys move the cursor while the mode is active — never
	// the free camera-pan they drive outside it.
	m := widescreenModel(t)
	var mdl tea.Model = update(m, "v")
	before := mdl.(Model)
	mdl = update(mdl, "down")
	after := mdl.(Model)
	if after.lookY != before.lookY+1 {
		t.Fatalf("down-arrow should move the cursor down by 1, got lookY %d -> %d", before.lookY, after.lookY)
	}
	// panX/panY only change as a side effect of the camera PUSH margin, not
	// as an independent free-pan accumulator — a single tile of movement
	// starting from the camera center should never trigger a push (comfortably
	// inside the margin on a normal-sized viewport).
	if after.panX != before.panX || after.panY != before.panY {
		t.Errorf("a single interior move should not have pushed the camera: pan %d,%d -> %d,%d", before.panX, before.panY, after.panX, after.panY)
	}
}

// --- camera push + snap (research R3) ---

func TestLookCameraPushesAtMargin(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = nil
	var mdl tea.Model = update(m, "v")
	mm := mdl.(Model)
	vw, _ := mm.mapViewportDims()
	x0, _ := mm.cameraOrigin(vw, 0)
	// Walk the cursor to the viewport's right margin and one step beyond —
	// the camera must pan right to keep the cursor >= lookCameraMargin
	// tiles inside the viewport.
	steps := vw/2 + vw - lookCameraMargin
	for i := 0; i < steps; i++ {
		mdl = update(mdl, "l")
	}
	after := mdl.(Model)
	newX0, _ := after.cameraOrigin(vw, 0)
	if newX0 <= x0 {
		t.Errorf("camera should have panned right as the cursor approached the viewport edge: x0 %d -> %d", x0, newX0)
	}
	if after.lookX < newX0+lookCameraMargin-1 {
		t.Errorf("cursor should stay >= %d tiles inside the viewport once pushed", lookCameraMargin)
	}
}

func TestLookCameraStopsAtWorldEdge(t *testing.T) {
	m := widescreenModel(t)
	var mdl tea.Model = update(m, "v")
	mm := mdl.(Model)
	mm.lookX = mm.gameMap.W - 1
	mdl = mm
	// Push further right — the camera clamps at the world edge (cameraOrigin's
	// own clamp), never panics, never produces an out-of-range origin.
	mdl = update(mdl, "l")
	after := mdl.(Model)
	if after.lookX != mm.gameMap.W-1 {
		t.Errorf("cursor already at the world edge should not move further: got %d", after.lookX)
	}
	vw, vh := after.mapViewportDims()
	x0, _ := after.cameraOrigin(vw, vh)
	if x0 < 0 || x0+vw > after.gameMap.W {
		t.Errorf("camera origin out of bounds at the world edge: x0=%d vw=%d W=%d", x0, vw, after.gameMap.W)
	}
}

func TestLookCSnapsCameraOntoCursor(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: 10, Y: 10}}
	var mdl tea.Model = update(m, "v")
	mm := mdl.(Model)
	mm.lookX, mm.lookY = 30, 30
	mdl = mm
	mdl = update(mdl, "c")
	after := mdl.(Model)
	cx, cy := after.wandererCentroid()
	if cx+after.panX != 30 || cy+after.panY != 30 {
		t.Errorf("'c' should snap the camera so centroid+pan lands on the cursor: got (%d,%d)", cx+after.panX, cy+after.panY)
	}

	// Outside the mode, 'c' keeps its pre-existing recenter-on-wanderers
	// meaning untouched.
	mdl = update(mdl, "v") // exit
	mdl = update(mdl, "c")
	after2 := mdl.(Model)
	if after2.panX != 0 || after2.panY != 0 {
		t.Error("'c' outside the mode should still just recenter on the wanderers (pan -> 0,0)")
	}
}

// --- esc release chain (SC-004, FR-003, AS 3.3) ---

func TestLookEscChainReleasesOneLayerPerPress(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = []sim.Agent{{Name: "Ash"}}
	var mdl tea.Model = update(m, "v")
	mm := mdl.(Model)
	mm.lookX, mm.lookY = mm.replica.Agents[0].X, mm.replica.Agents[0].Y
	mdl = mm

	// cursor -> pane
	mdl = update(mdl, "enter")
	if mdl.(Model).lookFocus != lookFocusPane {
		t.Fatal("enter from cursor focus must move to pane focus")
	}
	// pane -> drill (agent row is index 0 on an agent-occupied tile)
	mdl = update(mdl, "enter")
	if mdl.(Model).lookFocus != lookFocusDrill {
		t.Fatal("enter on the agent row must move to drill focus")
	}

	// esc #1: drill -> pane
	mdl = update(mdl, "esc")
	got := mdl.(Model)
	if got.lookFocus != lookFocusPane || !got.lookActive {
		t.Fatalf("esc #1 should release drill -> pane (still active): focus=%v active=%v", got.lookFocus, got.lookActive)
	}
	// esc #2: pane -> cursor
	mdl = update(mdl, "esc")
	got = mdl.(Model)
	if got.lookFocus != lookFocusCursor || !got.lookActive {
		t.Fatalf("esc #2 should release pane -> cursor (still active): focus=%v active=%v", got.lookFocus, got.lookActive)
	}
	// esc #3: cursor -> mode off
	mdl = update(mdl, "esc")
	got = mdl.(Model)
	if got.lookActive {
		t.Fatal("esc #3 should exit the mode entirely")
	}
	if got.panX != 0 || got.panY != 0 {
		t.Error("exiting via esc must resume centroid-following")
	}
	// esc #4: the normal global esc (nothing left from this feature) —
	// must not panic and must not resurrect any look state.
	mdl = update(mdl, "esc")
	if mdl.(Model).lookActive {
		t.Fatal("the 4th esc must not be swallowed back into look state")
	}
}

// --- dock-borrow state preservation (data-model.md invariant 4, R2) ---

func TestLookBorrowPreservesChronicleInspectState(t *testing.T) {
	m := widescreenModel(t)
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{Paused: true}}
	m.connected = true
	seedEvents(&m, 5)
	m.dockTab = paneChronicle
	m.chronSelected = 2
	m.chronDetailScroll = 1
	if !m.inspecting() {
		t.Fatal("test setup: expected inspect mode active before entering look")
	}

	var mdl tea.Model = update(m, "v")
	mm := mdl.(Model)
	if mm.dockTab != paneChronicle {
		t.Error("entering the mode must never write m.dockTab")
	}
	if mm.inspecting() {
		t.Error("chronicleVisible (and thus inspecting) must read false while the borrow is active")
	}
	// The dock body during the borrow is the TILE view, not the chronicle.
	if !strings.Contains(mm.dockTabContent(60, 20), "TILE") {
		t.Error("dockTabContent must render the TILE view while lookActive")
	}

	// Exit via esc: chronicle state untouched throughout.
	mdl = update(mdl, "esc")
	after := mdl.(Model)
	if after.chronSelected != 2 || after.chronDetailScroll != 1 {
		t.Errorf("chronicle selection/scroll must survive the borrow untouched: sel=%d scroll=%d", after.chronSelected, after.chronDetailScroll)
	}
	if !after.inspecting() {
		t.Error("exiting the mode should restore inspect mode exactly as it was")
	}
}

func TestLookDigitExitsAndSelectsTab(t *testing.T) {
	m := widescreenModel(t)
	m.dockTab = paneGuardian
	var mdl tea.Model = update(m, "v")
	mdl = update(mdl, "4") // villagers
	mm := mdl.(Model)
	if mm.lookActive {
		t.Fatal("a digit key must exit the mode")
	}
	if mm.dockTab != paneVillagers {
		t.Errorf("digit 4 should select the villagers tab, got %v", mm.dockTab)
	}
}

// --- visibility dormancy (research R1, FR-013) ---

func TestLookVisibilityDormancyAndUnclaimedGlobalKeys(t *testing.T) {
	m := widescreenModel(t)
	m.dockTab = paneVillagers
	m.replica.Agents = []sim.Agent{{Name: "Ash"}}
	m.connected = true
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{Paused: false, Speed: "4x"}}
	var mdl tea.Model = update(m, "v")
	mm := mdl.(Model)
	if mm.villagersVisible() || mm.chronicleVisible() || mm.exerciseVisible() || mm.guardianVisible() {
		t.Error("every other dock-tab visibility predicate must read false during the borrow")
	}

	// j/k must not move the (dormant) villagers roster selection while in
	// cursor focus (they move the cursor instead).
	before := mm.villSelected
	mdl2 := update(mdl, "j")
	if mdl2.(Model).villSelected != before {
		t.Error("villagers roster selection must not move during the borrow")
	}

	// FR-013: unclaimed global keys keep working.
	mdl3, cmd := mdl.(Model).Update(key(" "))
	if cmd == nil {
		t.Error("space (pause/resume) must still work while the mode is active")
	}
	_ = mdl3
	mdl4, cmd2 := mdl.(Model).Update(key("]"))
	if cmd2 == nil {
		t.Error("] (speed) must still work while the mode is active")
	}
	_ = mdl4
}

// --- TILE view: hierarchy order (AC #9) + registry round-trip (SC-002) ---

func TestTileRowsHierarchyOrder(t *testing.T) {
	m := widescreenModel(t)
	cx, cy := 20, 20
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: cx, Y: cy, Intent: &sim.Intent{Goal: "forage"}}}
	m.replica.Piles = []sim.Pile{{X: cx, Y: cy, Wood: 3}}
	m.replica.Structures = []sim.Structure{{Kind: "shelter", X: cx, Y: cy}}
	m.lookX, m.lookY = cx, cy

	rows := m.tileRows()
	var order []tileBand
	for _, r := range rows {
		if len(order) == 0 || order[len(order)-1] != r.band {
			order = append(order, r.band)
		}
	}
	want := []tileBand{bandAgents, bandPilesChests, bandStructures, bandTerrain}
	if len(order) != len(want) {
		t.Fatalf("band order = %v, want %v", order, want)
	}
	for i, b := range want {
		if order[i] != b {
			t.Errorf("band[%d] = %v, want %v (full order %v)", i, order[i], b, order)
		}
	}
}

func TestTileRowsEmptyTileShowsTerrainOnly(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = nil
	m.lookX, m.lookY = 5, 5
	// Make sure nothing else occupies (5,5).
	m.replica.Piles = nil
	m.replica.Structures = nil
	rows := m.tileRows()
	if len(rows) != 1 || rows[0].band != bandTerrain {
		t.Errorf("an empty tile should show exactly the terrain row, got %+v", rows)
	}
}

// TestTileRegistryRoundTrip (SC-002's fourth surface): a row registered
// after this feature reaches the TILE view's terrain row/header with zero
// renderer edits — the spec-068 seam extended.
func TestTileRegistryRoundTrip(t *testing.T) {
	const testKind = worldmap.TileKind(201)
	n := len(mapGlyphs)
	registerTile(tileEntry{
		Glyph: "¤", Name: "bog", Meaning: "a test-only bog tile (spec 074 round-trip)",
		Token: tokDen, Terrain: []worldmap.TileKind{testKind},
	})
	t.Cleanup(func() {
		mapGlyphs = mapGlyphs[:n]
		rebuildTileIndex()
	})

	m := widescreenModel(t)
	m.replica.Agents = nil
	cx, cy := m.gameMap.W/2, m.gameMap.H/2
	m.gameMap.Tiles[cy*m.gameMap.W+cx] = testKind
	m.lookX, m.lookY = cx, cy

	if got := m.terrainRowMeaning(); got != "a test-only bog tile (spec 074 round-trip)" {
		t.Errorf("terrainRowMeaning did not pick up the registered row: %q", got)
	}
	body := m.tileBody(60, 20)
	if !strings.Contains(body, "a test-only bog tile") {
		t.Errorf("TILE body header should carry the registered row's meaning:\n%s", body)
	}
}

// --- env header (FR-007) ---

func TestTileEnvHeaderFireWarmthAndLight(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = nil
	m.replica.Structures = []sim.Structure{{Kind: "fire", X: 10, Y: 10, FuelUntil: 100000}}
	m.replica.Tick = 100
	m.lookX, m.lookY = 10, 10
	m.replica.Night = true

	lines := m.tileEnvHeader()
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "warm") || !strings.Contains(joined, "fire's radius") {
		t.Errorf("warmth line should name the fire source: %q", joined)
	}
}

func TestTileEnvHeaderDaylightNoSource(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Structures = nil
	m.replica.Night = false
	m.lookX, m.lookY = 40, 40
	joined := strings.Join(m.tileEnvHeader(), "\n")
	if !strings.Contains(joined, "mild") || !strings.Contains(joined, "daylight") {
		t.Errorf("daytime with no warmth source should read mild/daylight: %q", joined)
	}
	if !strings.Contains(joined, "bright") {
		t.Errorf("daytime light should read bright: %q", joined)
	}
}

// --- tileEvents: recorded-position filter (research R5) ---

func TestTileEventsFiltersToRecordedPosition(t *testing.T) {
	m := widescreenModel(t)
	m.events = []store.Event{
		{Seq: 1, Type: "agent.moved", Payload: json.RawMessage(`{"agent":0,"x":5,"y":5}`)},
		{Seq: 2, Type: "agent.moved", Payload: json.RawMessage(`{"agent":0,"x":6,"y":6}`)},
		{Seq: 3, Type: "agent.slept", Payload: json.RawMessage(`{"agent":0}`)}, // actor-only, no position
	}
	idx := m.tileEvents(5, 5)
	if len(idx) != 1 || m.events[idx[0]].Seq != 1 {
		t.Errorf("tileEvents(5,5) = %v, want just the seq-1 event", idx)
	}
	if len(m.tileEvents(6, 6)) != 1 {
		t.Error("tileEvents(6,6) should find the seq-2 event")
	}
	if len(m.tileEvents(0, 0)) != 0 {
		t.Error("tileEvents at an unrelated tile should be empty")
	}
}

// --- mouse parity (US4) ---

func TestLookMouseClickEntersModeAtClickedTile(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = nil
	_ = m.mapPanelView(m.width-70, 40) // render once to populate mapHit
	// Re-derive the panel size the same way widescreenView would.
	cols := computeColumns(m.width)
	rows := computeRows(m.height, m.wantsLessonRow())
	_ = m.mapPanelView(cols.MapCols, rows.Body) // record mapHit with the live layout's geometry

	hit := m.mapHit
	if hit == nil || !hit.valid {
		t.Fatal("test setup: mapHit should be valid after rendering the map panel")
	}
	clickX := hit.originX + 2 // tile 1 (stride 2 cols/tile)
	clickY := hit.originY + 1

	mdl, _ := m.Update(mouseLeftRelease(clickX, clickY))
	mm := mdl.(Model)
	if !mm.lookActive {
		t.Fatal("clicking a map tile should enter the look-cursor mode")
	}
	wantX, wantY := hit.x0+1, hit.y0+1
	if mm.lookX != wantX || mm.lookY != wantY {
		t.Errorf("cursor should land on the clicked tile (%d,%d), got (%d,%d)", wantX, wantY, mm.lookX, mm.lookY)
	}
}

func TestLookMouseClickMovesActiveCursor(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = nil
	cols := computeColumns(m.width)
	rows := computeRows(m.height, m.wantsLessonRow())
	_ = m.mapPanelView(cols.MapCols, rows.Body)

	var mdl tea.Model = update(m, "v")
	mm := mdl.(Model)
	// Re-record mapHit under the now-active mode (title/content differs but
	// geometry doesn't).
	_ = mm.mapPanelView(cols.MapCols, rows.Body)
	hit := mm.mapHit
	clickX := hit.originX + 6
	clickY := hit.originY + 2

	mdl2, _ := mm.Update(mouseLeftRelease(clickX, clickY))
	after := mdl2.(Model)
	if !after.lookActive {
		t.Fatal("mode should remain active (no churn) on a second click")
	}
	if after.lookX != hit.x0+3 || after.lookY != hit.y0+2 {
		t.Errorf("cursor should move to the newly clicked tile: got (%d,%d)", after.lookX, after.lookY)
	}
}

func TestLookMouseGuardsNoOpDuringHelpOrMinibuffer(t *testing.T) {
	m := widescreenModel(t)
	cols := computeColumns(m.width)
	rows := computeRows(m.height, m.wantsLessonRow())
	_ = m.mapPanelView(cols.MapCols, rows.Body)
	hit := m.mapHit
	clickX, clickY := hit.originX+2, hit.originY+1

	help := m
	help.helpOpen = true
	if mdl, _ := help.Update(mouseLeftRelease(clickX, clickY)); mdl.(Model).lookActive {
		t.Error("a map click while help is open must be a no-op")
	}
	mb := m
	mb.mbFocused = true
	if mdl, _ := mb.Update(mouseLeftRelease(clickX, clickY)); mdl.(Model).lookActive {
		t.Error("a map click while the minibuffer is focused must be a no-op")
	}
}

func TestLookMouseTilePaneClickSelectsThenDrills(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: 8, Y: 8}}
	var mdl tea.Model = update(m, "v")
	mm := mdl.(Model)
	mm.lookX, mm.lookY = 8, 8
	mdl = mm

	dockPanel := (mdl.(Model)).dockPanelView(computeColumns(mdl.(Model).width).DockCols, computeRows(mdl.(Model).height, mdl.(Model).wantsLessonRow()).Body)
	_ = dockPanel // render to populate tileHit

	hit := mdl.(Model).tileHit
	if hit == nil || !hit.valid {
		t.Fatal("test setup: tileHit should be valid after rendering the dock panel")
	}
	// Find the first selectable row (the agent row) in the recorded region.
	rowOffset := -1
	for i, ri := range hit.rowIndex {
		if ri >= 0 {
			rowOffset = i
			break
		}
	}
	if rowOffset < 0 {
		t.Fatal("test setup: expected at least one selectable TILE-pane row")
	}
	clickX, clickY := hit.originX, hit.originY+rowOffset

	mdl2, _ := mdl.(Model).Update(mouseLeftRelease(clickX, clickY))
	sel1 := mdl2.(Model)
	if sel1.lookFocus != lookFocusPane || sel1.lookSel != hit.rowIndex[rowOffset] {
		t.Fatalf("first click should select the row and acquire pane focus: focus=%v sel=%d", sel1.lookFocus, sel1.lookSel)
	}

	mdl3, _ := sel1.Update(mouseLeftRelease(clickX, clickY))
	sel2 := mdl3.(Model)
	if sel2.lookFocus != lookFocusDrill {
		t.Error("a second click on the already-selected row should drill in")
	}
}

// --- narrow fallback (FR-012, research R7): entry / body swap / unwind ---

func TestLookNarrowEntrySwapAndUnwind(t *testing.T) {
	m := testModel(t) // narrow (80x30), active defaults to paneMap
	if isWidescreen(m.width) {
		t.Fatal("test setup: expected narrow layout")
	}
	if m.active != paneMap {
		t.Fatalf("test setup: expected active pane to default to paneMap, got %v", m.active)
	}

	var mdl tea.Model = update(m, "v")
	mm := mdl.(Model)
	if !mm.lookActive || mm.lookFocus != lookFocusCursor {
		t.Fatal("'v' should raise the cursor on the narrow map pane")
	}
	// Cursor focus: narrowView still renders the map itself (with the cursor
	// highlight), not the TILE view — content swap only happens once focus
	// moves into the pane (FR-012).
	view := mm.View()
	if strings.Contains(view, "TILE (") {
		t.Error("narrow cursor focus should still show the map, not the TILE view")
	}

	mdl = update(mdl, "enter")
	afterFocus := mdl.(Model)
	if afterFocus.lookFocus != lookFocusPane {
		t.Fatal("enter should move focus into the TILE pane in narrow too")
	}
	swappedView := afterFocus.View()
	if !strings.Contains(swappedView, "TILE (") {
		t.Errorf("narrow pane focus should swap the map pane's body to the TILE view:\n%s", swappedView)
	}

	// Unwind: esc releases pane -> cursor -> mode off, each one layer.
	mdl = update(mdl, "esc")
	if mdl.(Model).lookFocus != lookFocusCursor || !mdl.(Model).lookActive {
		t.Fatal("esc from narrow pane focus should release to cursor focus, mode still active")
	}
	mdl = update(mdl, "esc")
	if mdl.(Model).lookActive {
		t.Fatal("esc from narrow cursor focus should exit the mode entirely")
	}
}

// TestLookFoldPressureResizeMidMode (spec edge case): a resize that triggers
// the widescreen fold cascade while the mode is active must neither panic
// nor change any panel dimension beyond what the fold itself would already
// do mode-off — the mode participates in none of computeRows' fold
// accounting (it adds zero chrome rows).
func TestLookFoldPressureResizeMidMode(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: 5, Y: 5}}
	var mdl tea.Model = update(m, "v")
	mdl = update(mdl, "enter") // pane focus, to exercise a non-trivial mode state

	// Shrink height enough to trigger the fold cascade (villager strip, then
	// lesson row, then guardian strip reclaim their rows) while still
	// widescreen-wide.
	mdl2, _ := mdl.(Model).Update(tea.WindowSizeMsg{Width: 140, Height: 20})
	after := mdl2.(Model)

	if !after.lookActive || after.lookFocus != lookFocusPane {
		t.Fatal("a resize mid-mode should not itself change the mode's state")
	}
	view := after.View() // must not panic
	if view == "" {
		t.Fatal("view rendered empty after a fold-pressure resize mid-mode")
	}
	lines := strings.Split(view, "\n")
	if len(lines) != after.height {
		t.Errorf("View() line count = %d, want exactly height %d (B1 exact-height invariant) after a fold-pressure resize mid-mode", len(lines), after.height)
	}

	// Compare against the SAME resize with the mode off: the fold's own
	// panel-dimension outcome must be identical whether or not the mode is
	// active (the mode adds zero chrome rows, so it must never itself be
	// the reason a panel's size differs).
	mOff := widescreenModel(t)
	mOff.replica.Agents = m.replica.Agents
	mdlOff, _ := mOff.Update(tea.WindowSizeMsg{Width: 140, Height: 20})
	offAfter := mdlOff.(Model)
	cols := computeColumns(after.width)
	rows := computeRows(after.height, after.wantsLessonRow())
	rowsOff := computeRows(offAfter.height, offAfter.wantsLessonRow())
	if rows != rowsOff {
		t.Errorf("fold outcome differs mode-on vs mode-off at the same size: %+v vs %+v", rows, rowsOff)
	}
	mapPanel := after.mapPanelView(cols.MapCols, rows.Body)
	dockPanel := after.dockPanelView(cols.DockCols, rows.Body)
	if lipgloss.Height(mapPanel) != rows.Body || lipgloss.Height(dockPanel) != rows.Body {
		t.Errorf("panel heights should match the folded body budget %d: map=%d dock=%d", rows.Body, lipgloss.Height(mapPanel), lipgloss.Height(dockPanel))
	}
}

// --- SC-003: fixed geometry across mode entry / pane focus / drill / exit ---

func TestLookFixedGeometryAcrossModeStates(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: 5, Y: 5}}
	cols := computeColumns(m.width)
	rows := computeRows(m.height, m.wantsLessonRow())

	dims := func(mm Model) (mw, mh, dw, dh int) {
		mapPanel := mm.mapPanelView(cols.MapCols, rows.Body)
		dockPanel := mm.dockPanelView(cols.DockCols, rows.Body)
		return lipgloss.Width(mapPanel), lipgloss.Height(mapPanel), lipgloss.Width(dockPanel), lipgloss.Height(dockPanel)
	}

	mw0, mh0, dw0, dh0 := dims(m)

	var mdl tea.Model = update(m, "v")
	mm := mdl.(Model)
	mw1, mh1, dw1, dh1 := dims(mm)

	mm.lookX, mm.lookY = 5, 5
	mdl = mm
	mdl = update(mdl, "enter") // pane focus
	mw2, mh2, dw2, dh2 := dims(mdl.(Model))

	mdl = update(mdl, "enter") // drill (agent row)
	mw3, mh3, dw3, dh3 := dims(mdl.(Model))

	mdl = update(mdl, "esc")
	mdl = update(mdl, "esc")
	mdl = update(mdl, "esc") // back to mode-off
	mw4, mh4, dw4, dh4 := dims(mdl.(Model))

	states := [][4]int{{mw0, mh0, dw0, dh0}, {mw1, mh1, dw1, dh1}, {mw2, mh2, dw2, dh2}, {mw3, mh3, dw3, dh3}, {mw4, mh4, dw4, dh4}}
	for i := 1; i < len(states); i++ {
		if states[i] != states[0] {
			t.Errorf("panel geometry changed at state %d: got %v, want %v (mode-off)", i, states[i], states[0])
		}
	}
}
