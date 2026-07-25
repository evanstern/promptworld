package tui

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evanstern/promptworld/internal/guardian"
	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
)

// testModel defaults to a narrow (80-col) terminal — the widescreen tests
// set width explicitly (>= widescreenBreakpoint) where they need it.
func testModel(t *testing.T) Model {
	t.Helper()
	// Isolate the per-user lessons-seen record (spec 055, TASK-117): New()
	// now loads worlds.LoadLessonsSeen() at construction, which resolves
	// the REAL developer machine's ~/.promptworld absent this override —
	// every test going through New() must not read or write outside its
	// own t.TempDir() (the internal/worlds setHome(t) precedent).
	t.Setenv("PROMPTWORLD_HOME", t.TempDir()+"/home")
	w, err := world.Create(t.TempDir()+"/w", "test", 42)
	if err != nil {
		t.Fatal(err)
	}
	m := New(w)
	m.replica = sim.NewState(42, w.Map())
	m.width, m.height = 80, 30
	return m
}

func widescreenModel(t *testing.T) Model {
	t.Helper()
	m := testModel(t)
	m.width, m.height = 140, 40
	return m
}

func key(s string) tea.KeyMsg {
	switch s {
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func update(mdl tea.Model, k string) tea.Model {
	next, _ := mdl.Update(key(k))
	return next
}

// mouseLeftRelease builds the one mouse event this feature binds (research
// R2): a left-button release at screen cell (x,y).
func mouseLeftRelease(x, y int) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft}
}

func TestMapRendersWanderers(t *testing.T) {
	m := testModel(t)
	m.replica.Agents = []sim.Agent{
		{Name: "Ash", X: 3, Y: 4},
		{Name: "Birch", X: 10, Y: 2, Asleep: true},
	}
	view := m.mapView()
	lines := strings.Split(view, "\n")
	gridOnly := strings.Join(lines[:len(lines)-1], "\n") // drop the legend line
	if !strings.Contains(gridOnly, "A") {
		t.Error("awake wanderer A missing from map grid")
	}
	if !strings.Contains(gridOnly, "b") {
		t.Error("asleep wanderer should render lowercase b in map grid")
	}
	if len(lines) < 15 {
		t.Errorf("map viewport has %d lines, want a real window", len(lines))
	}
	if !strings.Contains(view, "~") {
		t.Error("terrain (water) missing from rendered window")
	}
}

// TestCenterCameraOnPanEquivalence (spec 049 T003, research R1):
// centerCameraOn is exactly a computed pan — panX/panY land the wanderer
// centroid + pan on the target, the map title flips to "panned" (identical
// to manual arrow-key panning), and 'c' (recenter) restores auto-follow
// exactly as it always has.
func TestCenterCameraOnPanEquivalence(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: 10, Y: 10}}
	cx, cy := m.wandererCentroid()
	if cx != 10 || cy != 10 {
		t.Fatalf("test setup: centroid = (%d,%d), want (10,10)", cx, cy)
	}
	m.centerCameraOn(20, 15)
	if m.panX != 10 || m.panY != 5 {
		t.Errorf("centerCameraOn(20,15) with centroid (10,10): panX=%d panY=%d, want 10,5", m.panX, m.panY)
	}
	if view := m.mapPanelView(60, 20); !strings.Contains(view, "MAP · panned (c to recenter)") {
		t.Errorf("a nonzero pan should flip the map title to panned: %q", view)
	}

	// 'c' (recenter) resets pan to 0,0 and restores follow — unchanged
	// existing semantics (handleGlobalKey's "c" case); centerCameraOn adds
	// no new state for it to know about.
	var mdl tea.Model = m
	mdl = update(mdl, "c")
	mm := mdl.(Model)
	if mm.panX != 0 || mm.panY != 0 {
		t.Errorf("c should reset pan to 0,0: panX=%d panY=%d", mm.panX, mm.panY)
	}
	if view := mm.mapPanelView(60, 20); !strings.Contains(view, "MAP · following centroid") {
		t.Errorf("recenter should restore the following-centroid title: %q", view)
	}
}

// TestCenterCameraOnClampsLikeManualPan: a jump to a target far outside the
// map is clamped identically to manual panning — centerCameraOn adds no new
// camera math or clamping (research R1); render-time clampInt (renderMapGrid)
// and the resize-path clampGeometry are the only bounds, unchanged.
func TestCenterCameraOnClampsLikeManualPan(t *testing.T) {
	m := testModel(t)
	m.centerCameraOn(-99999, -99999)
	grid, _ := m.renderMapGrid(10, 10)
	if grid == "" {
		t.Fatal("renderMapGrid must not panic/blank out on a pathological jump target")
	}
	m.clampGeometry()
	if m.panX < -m.gameMap.W || m.panX > m.gameMap.W || m.panY < -m.gameMap.H || m.panY > m.gameMap.H {
		t.Errorf("panX/panY not bounded by the existing clamp: panX=%d panY=%d", m.panX, m.panY)
	}
}

// TestMapRendersPilesAndStockpileZones covers spec 013 T021 (US2-AS5,
// SC-006): the pile glyph appears on the map, adjacent piles are grouped
// into one stockpile zone by the render-side flood fill, and the legend
// (the map panel's one inspection surface — map.md "legend stays pinned as
// the panel's last row") reports each zone's/lone pile's contents as
// non-food counts + food batch totals.
func TestMapRendersPilesAndStockpileZones(t *testing.T) {
	m := testModel(t)
	cx, cy := m.gameMap.W/2, m.gameMap.H/2
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: cx, Y: cy}}
	m.replica.Piles = []sim.Pile{
		{X: cx, Y: cy, Wood: 3, Stone: 1},
		{X: cx + 1, Y: cy, Planks: 2}, // Manhattan-adjacent to the pile above → one zone
		{X: cx - 6, Y: cy - 6, Food: []sim.FoodBatch{{Kind: "food_raw", N: 5, SpoilAt: 100}}}, // isolated
	}
	view := m.mapView()
	lines := strings.Split(view, "\n")
	gridOnly := strings.Join(lines[:len(lines)-1], "\n")
	legend := lines[len(lines)-1]

	if !strings.Contains(gridOnly, "%") {
		t.Error("pile glyph % missing from map grid")
	}
	if !strings.Contains(legend, "zone[2]") {
		t.Errorf("legend should report the 2-pile adjacent zone, got: %s", legend)
	}
	if !strings.Contains(legend, "3w") || !strings.Contains(legend, "1st") || !strings.Contains(legend, "2pl") {
		t.Errorf("legend should summarize the zone's non-food counts, got: %s", legend)
	}
	if !strings.Contains(legend, "food 5r/0c/0m") {
		t.Errorf("legend should summarize the isolated pile's food batch totals, got: %s", legend)
	}
	if !strings.Contains(legend, "%pile") {
		t.Error("legend key should explain the % pile glyph")
	}
}

// TestPileZonesGroupsOnlyManhattanAdjacentPiles is a focused unit test on
// the flood fill itself (spec 013 T021): diagonal neighbors must NOT merge
// (data-model.md / spec.md restrict adjacency to the sim package's own
// Manhattan convention), while a chain of orthogonal drops does.
func TestPileZonesGroupsOnlyManhattanAdjacentPiles(t *testing.T) {
	piles := []sim.Pile{
		{X: 0, Y: 0, Wood: 1},
		{X: 1, Y: 0, Wood: 1}, // adjacent to (0,0)
		{X: 2, Y: 1, Wood: 1}, // diagonal to (1,0) only — not adjacent
		{X: 9, Y: 9, Wood: 1}, // far away, its own zone
	}
	zones := pileZones(piles)
	if len(zones) != 3 {
		t.Fatalf("want 3 zones (2-chain, diagonal-isolated, far-isolated), got %d: %+v", len(zones), zones)
	}
	if len(zones[0]) != 2 {
		t.Errorf("first zone should merge the two orthogonally adjacent piles, got %d piles", len(zones[0]))
	}
	if len(zones[1]) != 1 || zones[1][0].X != 2 {
		t.Errorf("diagonal neighbor must not merge into the chain, got %+v", zones[1])
	}
	if len(zones[2]) != 1 || zones[2][0].X != 9 {
		t.Errorf("far pile should be its own zone, got %+v", zones[2])
	}
}

// TestMapRendersChestGlyphAndInspection covers spec 013 T026 (SC-006): the
// chest glyph appears on the map, and the legend (the map panel's one
// inspection surface, T021's precedent) reports each visible chest's owner
// name, contents, and a fullness hint.
func TestMapRendersChestGlyphAndInspection(t *testing.T) {
	m := testModel(t)
	cx, cy := m.gameMap.W/2, m.gameMap.H/2
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: cx, Y: cy}, {Name: "Birch", X: cx + 2, Y: cy}}
	m.replica.Structures = []sim.Structure{
		// Off the agents' own tiles: the agent glyph outranks the structure
		// glyph on a shared tile (tile()'s priority order), so a chest at an
		// agent's own position would hide the glyph the test asserts on.
		{Kind: "chest", X: cx + 1, Y: cy, Owner: 1, Store: &sim.Inventory{Wood: 3, Planks: 2, FoodRaw: 5}},
	}
	view := m.mapView()
	lines := strings.Split(view, "\n")
	gridOnly := strings.Join(lines[:len(lines)-1], "\n")
	legend := lines[len(lines)-1]

	if !strings.Contains(gridOnly, "☐") {
		t.Error("chest glyph ☐ missing from map grid")
	}
	if !strings.Contains(legend, "☐chest") {
		t.Errorf("legend key should explain the ☐ chest glyph, got: %s", legend)
	}
	wantOwner := fmt.Sprintf("chest(%d,%d) [Birch]", cx+1, cy)
	if !strings.Contains(legend, wantOwner) {
		t.Errorf("legend should name the chest's owner by agent Name, got: %s", legend)
	}
	if !strings.Contains(legend, "3w") || !strings.Contains(legend, "2pl") || !strings.Contains(legend, "food 5r/0c/0m") {
		t.Errorf("legend should summarize the chest's contents, got: %s", legend)
	}
	if !strings.Contains(legend, "10/48") {
		t.Errorf("legend should show a fullness hint (3+2+5 = 10 of 48), got: %s", legend)
	}
}

// TestMapRendersWallGlyphs covers spec 032 T010 (US1, SC-006): plank/stone wall
// glyphs appear on the map, a damaged wall renders dim (the cold-fire
// precedent), and the legend documents the wall key.
func TestMapRendersWallGlyphs(t *testing.T) {
	m := testModel(t)
	cx, cy := m.gameMap.W/2, m.gameMap.H/2
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: cx, Y: cy}}
	m.replica.Structures = []sim.Structure{
		// Full-health plank wall (normal glyph) and a damaged stone wall (dim),
		// both off the agent's own tile so the agent glyph doesn't mask them.
		{Kind: "wall_plank", X: cx + 1, Y: cy, HP: sim.WallMaxHP("wall_plank")},
		{Kind: "wall_stone", X: cx + 2, Y: cy, HP: 100},
	}
	view := m.mapView()
	lines := strings.Split(view, "\n")
	gridOnly := strings.Join(lines[:len(lines)-1], "\n")
	legend := lines[len(lines)-1]

	if !strings.Contains(gridOnly, styleWall.Render("▤")) {
		t.Error("full-health plank wall glyph ▤ (normal style) missing from map grid")
	}
	if !strings.Contains(gridOnly, styleWallDamaged.Render("▩")) {
		t.Error("damaged stone wall glyph ▩ should render dim (styleWallDamaged)")
	}
	if !strings.Contains(legend, "▤▩wall") {
		t.Errorf("legend key should document the wall glyphs, got: %s", legend)
	}
}

// TestMapRendersPathGlyph covers spec 032 T019 (US3): a path renders at terrain
// level ("·" in the path style), the agent glyph wins on a shared tile, and the
// legend documents the path key.
func TestMapRendersPathGlyph(t *testing.T) {
	m := testModel(t)
	cx, cy := m.gameMap.W/2, m.gameMap.H/2
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: cx, Y: cy}}
	m.replica.Structures = []sim.Structure{
		{Kind: "path", X: cx + 1, Y: cy}, // off the agent's tile
		{Kind: "path", X: cx, Y: cy},     // under the agent — agent glyph must win
	}
	view := m.mapView()
	lines := strings.Split(view, "\n")
	gridOnly := strings.Join(lines[:len(lines)-1], "\n")
	legend := lines[len(lines)-1]

	if !strings.Contains(gridOnly, stylePath.Render("·")) {
		t.Error("path glyph · (path style) missing from map grid")
	}
	if !strings.Contains(gridOnly, styleAgent.Render("A")) {
		t.Error("the agent glyph must win over the path on a shared tile")
	}
	if !strings.Contains(legend, "·path") {
		t.Errorf("legend key should document the path glyph, got: %s", legend)
	}
}

// TestMapRendersGraveGlyph covers spec 044 T025 (US4, FR-017) and the
// ratified render-priority follow-up: the grave glyph must actually render
// at a REAL death tile, not just somewhere a grave happens to sit unoccupied
// — every post-044 death places its grave at the SAME tile the dead agent's
// own frozen position occupies, so asserting the glyph only off that tile
// would test a case the map never produces. The honest version: a dead
// agent AND a grave at the same tile renders "✝" (the body becomes the
// grave), overriding the plain dead marker that agent would otherwise show.
// A graveless dead agent (pre-044 replay/history, or any hand-built replica
// that never placed one) is unaffected and still renders "†" — covered here
// at a second agent/tile so both branches are exercised in one test.
func TestMapRendersGraveGlyph(t *testing.T) {
	m := testModel(t)
	cx, cy := m.gameMap.W/2, m.gameMap.H/2
	m.replica.Agents = []sim.Agent{
		{Name: "Ash", X: cx, Y: cy, Dead: true},       // dies WITH a grave at its own tile
		{Name: "Birch", X: cx + 3, Y: cy, Dead: true}, // graveless dead agent (pre-044 shape)
	}
	m.replica.Structures = []sim.Structure{
		{Kind: "grave", X: cx, Y: cy}, // co-located with Ash's own tile
	}
	view := m.mapView()
	lines := strings.Split(view, "\n")
	gridOnly := strings.Join(lines[:len(lines)-1], "\n")
	legend := lines[len(lines)-1]

	if !strings.Contains(gridOnly, styleGrave.Render("✝")) {
		t.Error("grave glyph ✝ (grave style) missing from map grid at the dead agent's own death tile")
	}
	if !strings.Contains(gridOnly, styleErr.Render("†")) {
		t.Error("a graveless dead agent should still render the plain † marker")
	}
	if !strings.Contains(legend, "✝grave") {
		t.Errorf("legend key should document the grave glyph, got: %s", legend)
	}
}

// TestDescribeChestEmptyStore covers the empty-chest and out-of-range-owner
// edges of T026's inspection line: an empty Store reads "empty" rather than
// a blank/zero-padded contents string, and an owner index outside the
// roster (a defensive case, not one the sim package should ever produce)
// renders via agentName's "#N" fallback instead of panicking.
func TestDescribeChestEmptyStore(t *testing.T) {
	got := describeChest(sim.Structure{X: 1, Y: 2, Owner: 5, Store: &sim.Inventory{}}, []string{"Ash"})
	want := "chest(1,2) [#5] empty 0/48"
	if got != want {
		t.Errorf("describeChest empty store: got %q, want %q", got, want)
	}
}

// TestVillagersRosterShowsFullInventory covers SC-006 (spec 012 T043): the
// villagers roster must surface every carried resource kind —
// wood/stone/water/planks/refined stone, the food triplet, and the
// most-worn spear's remaining uses.
func TestVillagersRosterShowsFullInventory(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = []sim.Agent{
		{Name: "Ash", X: 3, Y: 4, Inv: sim.Inventory{
			Wood: 1, Stone: 2, Water: 3, Planks: 4, RefinedStone: 5,
			FoodRaw: 6, FoodCooked: 7, Meals: 8, Spears: []int{1, 3},
		}},
	}
	body := m.villagerRosterBody(m.width-6, m.height-6)
	want := "carry 1w 2st 3wt 4pl 5rs · food 6r/7c/8m · spear 2(1)"
	if !strings.Contains(body, want) {
		t.Errorf("villagers roster missing full inventory line %q, got:\n%s", want, body)
	}
}

func TestApplyEventUpdatesReplicaAndChronicle(t *testing.T) {
	m := testModel(t)
	m.lastSeq = 10

	// At-or-before the snapshot seq: already reflected, must be skipped.
	stale := store.Event{Seq: 10, Tick: 5, Type: "agent.moved",
		Payload: json.RawMessage(`{"agent":0,"x":9,"y":9}`)}
	m.applyEvent(stale)
	if len(m.events) != 0 || m.replica.Agents[0].X == 9 {
		t.Fatal("stale event must not apply")
	}

	fresh := store.Event{Seq: 11, Tick: 60, Type: "agent.moved",
		Payload: json.RawMessage(`{"agent":0,"x":7,"y":8}`)}
	m.applyEvent(fresh)
	if m.replica.Agents[0].X != 7 || m.replica.Agents[0].Y != 8 {
		t.Errorf("replica not updated: %+v", m.replica.Agents[0])
	}
	if m.replica.Tick != 60 {
		t.Errorf("replica tick = %d, want 60", m.replica.Tick)
	}
	if m.lastSeq != 11 || len(m.events) != 1 {
		t.Errorf("chronicle/cursor wrong: lastSeq=%d events=%d", m.lastSeq, len(m.events))
	}

	night := store.Event{Seq: 12, Tick: 16 * 3600, Type: "sim.night_started",
		Payload: json.RawMessage(`{"day":1}`)}
	m.applyEvent(night)
	if !m.replica.Night {
		t.Error("night event did not flip replica to night")
	}
}

func TestChronicleRingCap(t *testing.T) {
	m := testModel(t)
	for i := int64(1); i <= chronicleCap+50; i++ {
		m.applyEvent(store.Event{Seq: i, Tick: i, Type: "sim.day_started",
			Payload: json.RawMessage(`{"day":1}`)})
	}
	if len(m.events) != chronicleCap {
		t.Errorf("ring size = %d, want %d", len(m.events), chronicleCap)
	}
	if m.events[0].Seq != 51 {
		t.Errorf("ring dropped wrong end: oldest seq %d", m.events[0].Seq)
	}
}

func TestQuitDetaches(t *testing.T) {
	m := testModel(t)
	mdl, cmd := m.Update(key("q"))
	if cmd == nil {
		t.Fatal("q must produce tea.Quit")
	}
	if v := mdl.(Model).View(); !strings.Contains(v, "keeps running") {
		t.Errorf("quit view should reassure the world keeps running: %q", v)
	}
}

// TestCtrlCQuitsFromAnyState is focus-contract.md rule 3: "ctrl+c quits the
// app from any state whatsoever" — including while the minibuffer is
// focused and mid-input.
func TestCtrlCQuitsFromAnyState(t *testing.T) {
	m := testModel(t)
	var mdl tea.Model = m
	mdl = update(mdl, "m")
	mdl = update(mdl, "h")
	mdl = update(mdl, "i")
	mdl, cmd := mdl.(Model).Update(key("ctrl+c"))
	if cmd == nil {
		t.Fatal("ctrl+c while focused must still produce tea.Quit")
	}
	if !mdl.(Model).quitting {
		t.Fatal("ctrl+c while focused must set quitting")
	}
}

// TestReplyTooLargeQuitsInsteadOfRetrying is TASK-19 AC#1 at the TUI: a
// reply over the protocol cap used to feed the 2s retry loop forever; now it
// is fatal — quit, with the actionable reason in the final view (and in
// cmdUI's exit error via FatalErr).
func TestReplyTooLargeQuitsInsteadOfRetrying(t *testing.T) {
	m := testModel(t)
	mdl, cmd := m.Update(disconnectedMsg{err: fmt.Errorf("state: %w", ipc.ErrReplyTooLarge)})
	mm := mdl.(Model)
	if !mm.quitting || mm.FatalErr() == "" {
		t.Fatalf("oversized reply must be fatal: quitting=%v fatal=%q", mm.quitting, mm.FatalErr())
	}
	if cmd == nil {
		t.Fatal("fatal disconnect must produce tea.Quit, not a retry")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("want tea.QuitMsg, got %T", cmd())
	}
	if v := mm.View(); !strings.Contains(v, "reply cap") {
		t.Errorf("final view should carry the reason: %q", v)
	}

	// Transient failures keep the old behavior: not fatal, schedule a retry.
	m2 := testModel(t)
	mdl2, cmd2 := m2.Update(disconnectedMsg{err: errors.New("daemon not running")})
	mm2 := mdl2.(Model)
	if mm2.quitting || mm2.FatalErr() != "" {
		t.Fatal("transient disconnect must not be fatal")
	}
	if cmd2 == nil {
		t.Fatal("transient disconnect should schedule a retry")
	}
}

func TestDisconnectedHeaderShowsRetry(t *testing.T) {
	m := testModel(t)
	m.connected = false
	m.lastErr = "daemon not running"
	if v := m.headerView(); !strings.Contains(v, "disconnected") {
		t.Errorf("header should show disconnected state: %q", v)
	}
}

// chronEntry appends a narrated entry to the test replica's ring.
func chronEntry(m *Model, day int64, text, thread string, agents ...int) {
	m.replica.Chronicle = append(m.replica.Chronicle, sim.ChronicleEntry{
		Tick: day * 86400, Day: day, FromTick: (day - 1) * 86400, ToTick: day * 86400,
		Text: text, Thread: thread, Agents: agents,
	})
}

// TestChronicleNarratedView is TASK-11 AC#1/#2 at the pane: narrated entries
// render, and the a/t keys filter by agent and thread.
func TestChronicleNarratedView(t *testing.T) {
	m := testModel(t)
	m.active = paneChronicle
	chronEntry(&m, 1, "Ash lit the first fire.", "cold-start", 0)
	chronEntry(&m, 2, "The gru circled Sage in the dark.", "gru", 7)

	view := m.chronicleView()
	if !strings.Contains(view, "Ash lit the first fire.") || !strings.Contains(view, "gru circled Sage") {
		t.Fatalf("narrated entries missing: %q", view)
	}

	// 'a' cycles to agent 0 (Ash): only entries mentioning Ash remain.
	var mdl tea.Model = m
	mdl = update(mdl, "a")
	view = mdl.(Model).chronicleView()
	if !strings.Contains(view, "first fire") || strings.Contains(view, "gru circled") {
		t.Errorf("agent filter leaked: %q", view)
	}

	// Back to all, then 't' cycles to the first thread (cold-start).
	for i := 0; i < len(m.replica.Agents); i++ {
		mdl = update(mdl, "a")
	}
	mdl = update(mdl, "t")
	mm := mdl.(Model)
	if mm.chronAgent != -1 || mm.chronThread != "cold-start" {
		t.Fatalf("filter state: agent=%d thread=%q", mm.chronAgent, mm.chronThread)
	}
	view = mm.chronicleView()
	if !strings.Contains(view, "first fire") || strings.Contains(view, "gru circled") {
		t.Errorf("thread filter leaked: %q", view)
	}

	// 't' again reaches "gru", once more wraps to all.
	mdl = update(mm, "t")
	if mdl.(Model).chronThread != "gru" {
		t.Errorf("thread cycle: %q", mdl.(Model).chronThread)
	}
	mdl = update(mdl, "t")
	if mdl.(Model).chronThread != "" {
		t.Errorf("thread cycle should wrap to all: %q", mdl.(Model).chronThread)
	}
}

// TestChronicleRawFallback: no narrated entries -> raw feed automatically;
// 'r' toggles back to raw even when narration exists.
func TestChronicleRawFallback(t *testing.T) {
	m := testModel(t)
	m.active = paneChronicle
	m.applyEvent(store.Event{Seq: 1, Tick: 60, Type: "agent.moved",
		Payload: json.RawMessage(`{"agent":0,"x":7,"y":8}`)})

	view := m.chronicleView()
	if !strings.Contains(view, "agent.moved") || !strings.Contains(view, "raw feed") {
		t.Fatalf("empty ring must fall back to the raw feed: %q", view)
	}

	chronEntry(&m, 1, "Ash lit the first fire.", "cold-start", 0)
	if view := m.chronicleView(); strings.Contains(view, "agent.moved") {
		t.Fatalf("narrated view should replace raw once entries exist: %q", view)
	}
	var mdl tea.Model = m
	mdl = update(mdl, "r")
	if view := mdl.(Model).chronicleView(); !strings.Contains(view, "agent.moved") {
		t.Errorf("'r' should show the raw feed: %q", view)
	}
}

// TestChronicleKeysScopedToPane: a/t/r do nothing outside the chronicle pane.
func TestChronicleKeysScopedToPane(t *testing.T) {
	m := testModel(t)
	m.active = paneMap
	chronEntry(&m, 1, "x", "cold-start", 0)
	var mdl tea.Model = m
	mdl = update(mdl, "a")
	mdl = update(mdl, "t")
	mdl = update(mdl, "r")
	mm := mdl.(Model)
	if mm.chronAgent != -1 || mm.chronThread != "" || mm.chronRaw {
		t.Errorf("filters changed outside the pane: %+v", mm)
	}
}

func TestWrapText(t *testing.T) {
	lines := wrapText("one two three four five", 9)
	if len(lines) != 3 || lines[0] != "one two" {
		t.Errorf("wrap: %v", lines)
	}
	if got := wrapText("", 10); got != nil {
		t.Errorf("empty wrap: %v", got)
	}
}

// TestMinibufferReply: a turn's reply, nudge, and moments land in the
// transcript and the busy flag clears; errors render honestly.
func TestMinibufferReply(t *testing.T) {
	m := testModel(t)
	m.active = paneGuardian
	m.dockTab = paneGuardian
	m.mbBusy = true
	var mdl tea.Model = m
	mdl, _ = mdl.(Model).Update(consoleReplyMsg{result: &guardian.TurnResult{
		Reply:   "It is done.",
		Nudge:   &guardian.Nudge{Form: "dream", Targets: []string{"Fern"}, Text: "a river of light"},
		Moments: []string{"day 3 — Ash died"},
		Charges: 0,
	}})
	mm := mdl.(Model)
	if mm.mbBusy {
		t.Fatal("busy flag not cleared")
	}
	view := mm.guardianView()
	for _, want := range []string{"It is done.", "dream", "Fern", "Ash died"} {
		if !strings.Contains(view, want) {
			t.Errorf("console view missing %q", want)
		}
	}
	mdl, _ = mm.Update(consoleReplyMsg{err: fmt.Errorf("tier is down")})
	if v := mdl.(Model).guardianView(); !strings.Contains(v, "unreachable") {
		t.Errorf("error not rendered honestly: %q", v)
	}
}

// TestMinibufferReplyOrderAndClock (spec 029 T023): a landed standing order,
// a cancellation, and a meta-tool's clock line each render into the
// transcript alongside the reply/nudge report lines.
func TestMinibufferReplyOrderAndClock(t *testing.T) {
	m := testModel(t)
	m.active = paneGuardian
	m.dockTab = paneGuardian
	var mdl tea.Model = m
	mdl, _ = mdl.(Model).Update(consoleReplyMsg{result: &guardian.TurnResult{
		Reply:     "As you say.",
		Order:     &guardian.OrderReport{ID: "ord-120-1", Condition: "Rowan falls asleep"},
		Cancelled: []string{"ord-90-2"},
		Clock:     "the world moves again",
	}})
	view := mdl.(Model).guardianView()
	for _, want := range []string{"watch set", "ord-120-1", "Rowan falls asleep", "watch released", "ord-90-2", "the world moves again"} {
		if !strings.Contains(view, want) {
			t.Errorf("console view missing %q: %s", want, view)
		}
	}
}

// TestGuardianBadgeWhenTabNotVisible is minibuffer.md's reply-arrival rule:
// stream in place if the guardian tab/pane is visible, otherwise badge the
// dock tab and flash the minibuffer once — never steal the selected tab.
func TestGuardianBadgeWhenTabNotVisible(t *testing.T) {
	m := widescreenModel(t)
	m.dockTab = paneChronicle // guardian not visible
	mdl, _ := m.Update(consoleReplyMsg{result: &guardian.TurnResult{Reply: "the wood is dry"}})
	mm := mdl.(Model)
	if !mm.guardianUnseen {
		t.Error("tab should badge when guardian tab is not the visible one")
	}
	if mm.dockTab != paneChronicle {
		t.Error("arriving reply must not steal the selected tab")
	}
	if mm.mbFlash == "" {
		t.Error("minibuffer should flash once when the reply lands off-tab")
	}

	// Selecting the guardian tab clears the badge and flash.
	mdl2, _ := mm.selectTab(paneGuardian)
	mm2 := mdl2.(Model)
	if mm2.guardianUnseen || mm2.mbFlash != "" {
		t.Error("selecting the guardian tab should clear the badge/flash")
	}
}

// TestConsoleToolsSummary (spec 021 T021, SC-005): the console header's
// granted-tool summary is quiet for a full-grant default world, "none" for a
// conversation-only world, and the short-form set otherwise, carrying any
// miracle-kind restriction through.
func TestConsoleToolsSummary(t *testing.T) {
	cases := []struct {
		name string
		s    guardian.Status
		want string
	}{
		{"default is quiet", guardian.Status{ManifestDefault: true,
			GrantedTools: []string{"nudge_dream", "nudge_omen", "work_miracle"}}, ""},
		{"conversation-only", guardian.Status{ManifestDefault: false, GrantedTools: nil}, "tools: none"},
		{"subset short form", guardian.Status{ManifestDefault: false,
			GrantedTools: []string{"nudge_dream", "nudge_omen"}}, "tools: dream, omen"},
		{"restricted miracle kinds", guardian.Status{ManifestDefault: false,
			GrantedTools: []string{"nudge_dream", "work_miracle(move,give_item)"}},
			"tools: dream, workings(move,give_item)"},
		{"unrestricted miracles", guardian.Status{ManifestDefault: false,
			GrantedTools: []string{"work_miracle"}}, "tools: workings"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := c.s
			if got := consoleToolsSummary(&s, nil); got != c.want {
				t.Errorf("consoleToolsSummary = %q, want %q", got, c.want)
			}
		})
	}
}

// TestConsoleStageSummary (spec 046 T010): the guardian pane's stage line —
// absent for a pre-ladder/ungated world, the skin display name otherwise,
// with the charter-lock provenance appended at stage-1.
func TestConsoleStageSummary(t *testing.T) {
	cases := []struct {
		name string
		s    guardian.Status
		want string
	}{
		{"pre-ladder world is quiet", guardian.Status{Stage: ""}, ""},
		{"stage-2, unlocked", guardian.Status{Stage: "stage-2"}, "stage: The Written Word"},
		{"stage-1, locked", guardian.Status{Stage: "stage-1", CharterLocked: true, CharterPreset: "tutor"},
			"stage: The Voice (charter locked to tutor)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := c.s
			if got := consoleStageSummary(&s, nil); got != c.want {
				t.Errorf("consoleStageSummary = %q, want %q", got, c.want)
			}
		})
	}
}

// --- spec 049 T009/T010: chronicle-line click routing ---

// TestChronicleHitOriginFormulas: the three layouts' chrome-row/column
// offsets (research R3: stable, non-wrap-dependent geometry, safe to derive
// directly rather than duplicating chronicleInspectBody's own windowing).
func TestChronicleHitOriginFormulas(t *testing.T) {
	narrow := testModel(t) // 80x30 — below widescreenBreakpoint
	if ox, oy, w := narrow.chronicleHitOrigin(); ox != 0 || oy != 3 || w != narrow.width {
		t.Errorf("narrow origin = (%d,%d,%d), want (0,3,%d)", ox, oy, w, narrow.width)
	}

	dock := widescreenModel(t) // 140x40, not solo
	cols := computeColumns(dock.width)
	if ox, oy, w := dock.chronicleHitOrigin(); ox != cols.MapCols+cols.Gutter || oy != 4 || w != cols.DockCols {
		t.Errorf("widescreen dock origin = (%d,%d,%d), want (%d,4,%d)", ox, oy, w, cols.MapCols+cols.Gutter, cols.DockCols)
	}

	solo := widescreenModel(t)
	solo.solo = true
	if ox, oy, w := solo.chronicleHitOrigin(); ox != 0 || oy != 3 || w != solo.width {
		t.Errorf("solo origin = (%d,%d,%d), want (0,3,%d)", ox, oy, w, solo.width)
	}
}

// TestRecordChronHitRowToEventMapping (T009): chronicleInspectBody records
// one row->event mapping entry per rendered list row, in the same order the
// list itself shows them — the geometry handleMouse consumes next Update().
func TestRecordChronHitRowToEventMapping(t *testing.T) {
	m := pausedModel(t) // 5 events, indices 0..4
	m.chronSelected = 4
	m.chronicleInspectBody(60, 20)
	hit := m.chronHit
	if hit == nil || !hit.valid {
		t.Fatal("chronHit should be valid after rendering the inspect list")
	}
	if len(hit.rowEvent) != 5 {
		t.Fatalf("rowEvent has %d entries, want 5 (all seeded events fit the list budget)", len(hit.rowEvent))
	}
	for i, idx := range hit.rowEvent {
		if idx != i {
			t.Errorf("rowEvent[%d] = %d, want %d", i, idx, i)
		}
	}
}

// TestChronHitInvalidatedWhenChronicleNotRendered (data-model.md "State
// transitions"): a frame that doesn't render the chronicle inspect list
// (different tab) must invalidate any previously-recorded hit region —
// View() defaults to invalid every call; only chronicleInspectBody
// re-validates.
func TestChronHitInvalidatedWhenChronicleNotRendered(t *testing.T) {
	m := pausedModel(t)
	m.View()
	if !m.chronHit.valid {
		t.Fatal("chronHit should be valid right after rendering with the chronicle dock tab visible+paused")
	}
	m.dockTab = paneVillagers
	m.View()
	if m.chronHit.valid {
		t.Error("chronHit must invalidate once the chronicle inspect list is no longer part of the frame")
	}
}

// TestHandleMouseSelectsAndJumps (US2 AS-1): a left-release inside a
// recorded chronicle list row selects that row and applies the same jump
// rules ⏎ does.
func TestHandleMouseSelectsAndJumps(t *testing.T) {
	m := pausedModel(t) // widescreen, paused, chronicle dock tab, 5 events
	m.chronicleInspectBody(60, 20)
	if !m.chronHit.valid {
		t.Fatal("test setup: chronHit should be valid")
	}
	originX, originY, _ := m.chronicleHitOrigin()
	const row = 1
	if row >= len(m.chronHit.rowEvent) {
		t.Fatalf("test setup: only %d list rows recorded, want at least %d", len(m.chronHit.rowEvent), row+1)
	}
	wantIdx := m.chronHit.rowEvent[row]

	mdl, _ := m.Update(mouseLeftRelease(originX+2, originY+row))
	mm := mdl.(Model)
	if mm.chronSelected != wantIdx {
		t.Fatalf("click should select event %d, got %d", wantIdx, mm.chronSelected)
	}
	if mm.chronDetailScroll != 0 {
		t.Errorf("click-select should reset detail scroll like j/k: got %d", mm.chronDetailScroll)
	}
	// seedEvents' fixture (agent.moved{agent:0,x:1,y:1}) resolves via Ash's
	// live position in the default 8-agent replica (sim.NewState) — always
	// locatable, so the click's jump half should have fired too.
	if mm.panX == 0 && mm.panY == 0 {
		t.Error("click should also apply the jump (camera should have moved)")
	}
}

// TestHandleMouseOutOfRegionNoOp (contract §1: "left-click outside the
// chronicle list rows ... no-op").
func TestHandleMouseOutOfRegionNoOp(t *testing.T) {
	m := pausedModel(t)
	m.chronicleInspectBody(60, 20)
	before := m.chronSelected
	mdl, _ := m.Update(mouseLeftRelease(0, 9999)) // far below any recorded row
	mm := mdl.(Model)
	if mm.chronSelected != before || mm.panX != 0 || mm.panY != 0 {
		t.Error("a click outside the chronicle list rows must be a no-op")
	}
}

// TestHandleMouseRunningClockNoOp (US2 AS-2, FR-004): "Given the clock is
// running, When the player clicks a chronicle line, Then nothing happens."
func TestHandleMouseRunningClockNoOp(t *testing.T) {
	m := pausedModel(t)
	m.chronicleInspectBody(60, 20) // record geometry while paused
	originX, originY, _ := m.chronicleHitOrigin()
	m.status.Clock.Paused = false // now running: inspecting() goes false
	before := m.chronSelected

	mdl, _ := m.Update(mouseLeftRelease(originX+2, originY))
	mm := mdl.(Model)
	if mm.chronSelected != before || mm.panX != 0 || mm.panY != 0 {
		t.Error("a click while the clock runs must be a no-op, even against stale recorded geometry")
	}
}

// TestHandleMouseWrongButtonOrActionNoOp (research R2): only a left-button
// *release* is bound — press, motion, and other buttons are inert.
func TestHandleMouseWrongButtonOrActionNoOp(t *testing.T) {
	m := pausedModel(t)
	m.chronicleInspectBody(60, 20)
	originX, originY, _ := m.chronicleHitOrigin()
	cases := []tea.MouseMsg{
		{X: originX, Y: originY, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
		{X: originX, Y: originY, Action: tea.MouseActionRelease, Button: tea.MouseButtonRight},
		{X: originX, Y: originY, Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft},
	}
	for _, msg := range cases {
		before := m.chronSelected
		mdl, _ := m.Update(msg)
		if mdl.(Model).chronSelected != before {
			t.Errorf("mouse event %+v should be a no-op (only left-release is bound)", msg)
		}
	}
}

// TestHandleMouseInertDuringHelpAndMinibuffer: the chronicle click must not
// fire while another surface owns the keyboard/attention (help overlay,
// minibuffer focus) even if stale geometry from an earlier paused-chronicle
// frame is still recorded.
func TestHandleMouseInertDuringHelpAndMinibuffer(t *testing.T) {
	m := pausedModel(t)
	m.chronicleInspectBody(60, 20)
	originX, originY, _ := m.chronicleHitOrigin()

	help := m
	help.helpOpen = true
	if mdl, _ := help.Update(mouseLeftRelease(originX, originY)); mdl.(Model).chronSelected != help.chronSelected {
		t.Error("a click must be inert while the help overlay is open")
	}

	mb := m
	mb.mbFocused = true
	if mdl, _ := mb.Update(mouseLeftRelease(originX, originY)); mdl.(Model).chronSelected != mb.chronSelected {
		t.Error("a click must be inert while the minibuffer is focused")
	}
}

// --- systems dock tab (spec 053 US2/D10, T002) ---
// Tab-grammar regression: existing keys 2/3/4 must keep selecting exactly
// what they always selected once the 5/systems tab is added — and 5 must
// behave exactly like every other dock-tab key (select / same-key solo /
// same-key-again home, narrow reachability).

func TestTabGrammarUnchangedAfterSystemsAdded(t *testing.T) {
	m := widescreenModel(t)
	var mdl tea.Model = m
	for key, want := range map[string]pane{"2": paneChronicle, "3": paneGuardian, "4": paneVillagers} {
		mdl = update(widescreenModel(t), key)
		if got := mdl.(Model).dockTab; got != want {
			t.Errorf("%q selected %s, want %s (2/3/4 must be unchanged)", key, paneNames[got], paneNames[want])
		}
	}
}

// TestSystemsTabSoloZoomStateMachine: "5" follows the exact same
// select/solo/home state machine 2/3/4 already do (dock.md, solo-views.md).
func TestSystemsTabSoloZoomStateMachine(t *testing.T) {
	m := widescreenModel(t)
	var mdl tea.Model = m
	mdl = update(mdl, "5")
	mm := mdl.(Model)
	if mm.solo || mm.dockTab != paneSystems {
		t.Fatalf("first '5' should select (not solo): solo=%v tab=%s", mm.solo, paneNames[mm.dockTab])
	}
	mdl = update(mdl, "5")
	mm = mdl.(Model)
	if !mm.solo || mm.dockTab != paneSystems {
		t.Fatalf("second '5' should solo-zoom: solo=%v tab=%s", mm.solo, paneNames[mm.dockTab])
	}
	mdl = update(mdl, "5")
	if mdl.(Model).solo {
		t.Fatal("third '5' should return home")
	}
}

// TestSystemsTabReachableInNarrowFallback: below the widescreen breakpoint,
// '5' selects the systems pane exactly like 2/3/4 (panels/systems.md
// "Narrow behavior").
func TestSystemsTabReachableInNarrowFallback(t *testing.T) {
	m := testModel(t) // narrow
	var mdl tea.Model = m
	mdl = update(mdl, "5")
	if got := mdl.(Model).active; got != paneSystems {
		t.Fatalf("'5' in the narrow fallback should select the systems pane, got %s", paneNames[got])
	}
	if v := mdl.(Model).View(); v == "" {
		t.Error("narrow systems view rendered empty")
	}
}

// TestDockTabCycleIncludesSystems: tab/shift+tab (dockTab cycling aliases)
// reach the systems tab too, in its "5" position at the end of the row.
func TestDockTabCycleIncludesSystems(t *testing.T) {
	m := widescreenModel(t)
	var mdl tea.Model = m
	order := []pane{paneChronicle, paneGuardian, paneVillagers, paneSystems, paneChronicle}
	for i := 1; i < len(order); i++ {
		mdl = update(mdl, "tab")
		if got := mdl.(Model).dockTab; got != order[i] {
			t.Fatalf("tab-cycle step %d: dockTab = %s, want %s", i, paneNames[got], paneNames[order[i]])
		}
	}
}

// TestSystemsTabNoUnseenBadge (D10 "no second badge system"): unlike the
// guardian tab, the systems tab never carries an unseen-reply badge — there
// is nothing to badge (it carries no conversational content).
func TestSystemsTabNoUnseenBadge(t *testing.T) {
	m := widescreenModel(t)
	m.dockTab = paneSystems
	m.guardianUnseen = true // unrelated to systems — must not leak onto its label
	row := m.dockTabsRow()
	if strings.Contains(row, "systems") && strings.Contains(row, "SYSTEMS •") {
		t.Error("the systems tab must never render the unseen-badge dot")
	}
}
