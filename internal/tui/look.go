package tui

// Look-cursor mode (spec 074-look-cursor, TASK-142): a client-side-only tile
// inspector layered on the map, reopening the deferred look-cursor feature
// (docs/design/tui/panels/map.md's "evaluated and deferred" note). `v`
// enters/exits; hjkl/arrows move the cursor one tile, shift jumps 8; the
// camera pushes at a 2-tile margin; `c` snaps the camera onto the cursor;
// while active the dock body is borrowed by a transient TILE view in the
// solo-zoom-seam style (never a numbered tab, never m.dockTab). Nothing here
// touches the replica, the reducer, or IPC — every field is presentation
// state (data-model.md "Look-cursor mode state").

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// lookFocusKind is which layer inside the mode holds the keyboard
// (data-model.md invariant 1: pane/drill focus exists only inside the mode).
type lookFocusKind int

const (
	lookFocusCursor lookFocusKind = iota
	lookFocusPane
	lookFocusDrill
)

// lookDrillKind is what an opened TILE-pane row's drill-in shows.
type lookDrillKind int

const (
	lookDrillNone lookDrillKind = iota
	lookDrillAgent
	lookDrillPile
	lookDrillChest
	lookDrillEvent
)

// lookDrillRef is a TILE-pane row's drill target — the zero value means "not
// drillable" (data-model.md invariant 2: non-zero implies lookFocusDrill).
// idx is the agent index (lookDrillAgent), the replica.Structures index
// (lookDrillChest — disambiguates the rare multi-chest-per-tile case), or
// the m.events index (lookDrillEvent); meaningless for lookDrillPile (at
// most one pile per tile, research R5/agents.go Pile comment).
type lookDrillRef struct {
	kind lookDrillKind
	idx  int
}

// tileBand is the TILE pane's fixed hierarchy (FR-009): agents → piles/
// chests → structures → terrain → recorded-position events, always in this
// order, never re-sorted per tile.
type tileBand int

const (
	bandAgents tileBand = iota
	bandPilesChests
	bandStructures
	bandTerrain
	bandEvents
)

// tileRow is one row of the TILE pane's assembled content — the single
// source rendering, keyboard selection (lookSel indexes this slice), and the
// tileHit mouse region all read, so the three can never disagree
// (data-model.md "TILE view row model").
type tileRow struct {
	band  tileBand
	label string
	drill lookDrillRef // zero for non-drillable rows (terrain, the gru, dens)
}

// mapHitRegion is the map grid's last-rendered click geometry (research R6,
// the chronHitRegion pointer/invalidation pattern): a pointer so a
// value-receiver View() can still record it, invalidated by default every
// frame and re-validated only when the map grid actually rendered this
// frame (mapPanelView/mapView).
type mapHitRegion struct {
	valid            bool
	originX, originY int // screen cell of tile (x0,y0)'s glyph
	x0, y0           int // world coords of the top-left rendered tile
	vw, vh           int // viewport size in tiles (stride: 2 screen columns per tile)
}

// tileHitRegion is the TILE pane's last-rendered row-list click geometry —
// the chronHitRegion sibling for the borrowed dock body.
type tileHitRegion struct {
	valid            bool
	originX, originY int   // screen cell of the row list's first line
	width            int   // column span of a hit
	rowIndex         []int // rowIndex[i] = tileRows() index for screen row originY+i; -1 = a band heading, not a row
}

const lookCameraMargin = 2

// --- entry / exit (data-model.md invariant 5) ---

// lookEntryAllowed gates `v` (and mouse map-tile-click entry): the map must
// actually be the thing on screen — widescreen home (never solo, research
// R2's covered-by-body-replacing-surfaces rule) or the narrow map pane — and
// a world must be attached (the `x`-key documented-no-op precedent, spec
// edge case "No world state").
func (m Model) lookEntryAllowed() bool {
	if m.gameMap == nil || m.console {
		return false
	}
	if isWidescreen(m.width) {
		return !m.solo
	}
	return m.active == paneMap
}

// enterLook activates the mode: the cursor spawns at the camera-center tile
// (data-model.md invariant 5). dockTab/solo/active are never touched — the
// borrow reads m.lookActive at render time instead (research R2).
func (m *Model) enterLook() {
	vw, vh := m.mapViewportDims()
	x0, y0 := m.cameraOrigin(vw, vh)
	m.lookActive = true
	m.lookFocus = lookFocusCursor
	m.lookSel = 0
	m.lookDrill = lookDrillRef{}
	m.lookDrillScroll = 0
	m.lookX, m.lookY = x0+vw/2, y0+vh/2
	if m.gameMap != nil {
		m.lookX = clampInt(m.lookX, 0, m.gameMap.W-1)
		m.lookY = clampInt(m.lookY, 0, m.gameMap.H-1)
	}
}

// enterLookAt is the mouse map-click entry point (US4 AS1): the cursor
// spawns at the clicked tile instead of the camera center.
func (m *Model) enterLookAt(x, y int) {
	m.enterLook()
	if m.gameMap != nil {
		m.lookX = clampInt(x, 0, m.gameMap.W-1)
		m.lookY = clampInt(y, 0, m.gameMap.H-1)
	}
	m.pushCameraToCursor()
}

// exitLook deactivates the mode from any focus layer — every look field
// zeroes and the camera resumes centroid-following (data-model.md invariant
// 5: panX/panY reset to 0, "resume following" is literally the pre-existing
// following state).
func (m *Model) exitLook() {
	m.lookActive = false
	m.lookFocus = lookFocusCursor
	m.lookX, m.lookY = 0, 0
	m.lookSel = 0
	m.lookDrill = lookDrillRef{}
	m.lookDrillScroll = 0
	m.panX, m.panY = 0, 0
}

// --- movement + camera (research R3) ---

func (m *Model) moveLookCursor(dx, dy int) {
	if m.gameMap == nil {
		return
	}
	m.lookX = clampInt(m.lookX+dx, 0, m.gameMap.W-1)
	m.lookY = clampInt(m.lookY+dy, 0, m.gameMap.H-1)
	m.pushCameraToCursor()
}

// pushCameraToCursor adjusts panX/panY so the cursor sits at least
// lookCameraMargin tiles inside the CURRENT viewport (recomputed fresh via
// mapViewportDims/cameraOrigin every call — no cached geometry to go stale,
// research R3): the camera pans by the overshoot, degrading gracefully at
// the world edge, where cameraOrigin's own clamp stops the camera and the
// cursor may reach the viewport border (spec edge case).
func (m *Model) pushCameraToCursor() {
	vw, vh := m.mapViewportDims()
	x0, y0 := m.cameraOrigin(vw, vh)
	switch {
	case m.lookX < x0+lookCameraMargin:
		m.panX -= (x0 + lookCameraMargin) - m.lookX
	case m.lookX >= x0+vw-lookCameraMargin:
		m.panX += m.lookX - (x0 + vw - 1 - lookCameraMargin)
	}
	switch {
	case m.lookY < y0+lookCameraMargin:
		m.panY -= (y0 + lookCameraMargin) - m.lookY
	case m.lookY >= y0+vh-lookCameraMargin:
		m.panY += m.lookY - (y0 + vh - 1 - lookCameraMargin)
	}
}

// snapCameraToCursor is the in-mode `c` binding: the centerCameraOn
// jump-to-source formula (spec 049), one more caller of the same math —
// outside the mode `c` keeps its existing recenter-on-wanderers meaning
// untouched (handleGlobalKey's own "c" case).
func (m *Model) snapCameraToCursor() {
	m.centerCameraOn(m.lookX, m.lookY)
}

// --- key layer (research R1) ---

// lookDigitPane maps the exit-and-select digits to their dock pane — the
// same mapping handleGlobalKey's own "2".."6" cases use (dockTabKey
// reversed); "6" is only valid on a scenario world (spec 054 FR-008).
func (m Model) lookDigitPane(k string) (pane, bool) {
	switch k {
	case "2":
		return paneChronicle, true
	case "3":
		return paneGuardian, true
	case "4":
		return paneVillagers, true
	case "5":
		return paneSystems, true
	case "6":
		if m.exerciseID() != "" {
			return paneExercise, true
		}
	}
	return 0, false
}

// handleLookKey is patterns/keymap.md "Mode: look-cursor" — layered like
// handleInspectKey/handleVillagersKey (research R1), claiming the whole
// contested key set so the dormant inspect/villagers layers beneath never
// spuriously fire during the borrow (tui.go's chronicleVisible/
// villagersVisible/exerciseVisible guards already make them false, this is
// belt-and-braces at the dispatch site too). Every key claimed here either
// changes visible mode state or is a documented no-op (FR-004's "no key
// silently swallowed" reading) — unclaimed keys fall through to
// handleGlobalKey (space/[/]/m/q/p/? — FR-013).
func (m Model) handleLookKey(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	k := msg.String()

	// v is a hard exit from any focus layer (US1: "esc (or v again) exits").
	if k == "v" {
		m.exitLook()
		return m, nil, true
	}
	// G (console) exits the mode first (research R2's covered-by-
	// body-replacing-surfaces rule) but is deliberately NOT swallowed here:
	// handled=false lets it fall through to handleGlobalKey's own "G" case,
	// which actually opens the console — one console-open code path, not two.
	if k == "G" {
		m.exitLook()
		return m, nil, false
	}
	// Digits exit-and-select (research R2): claimed at every focus layer so
	// the borrow can never leave dockTab pointed somewhere the player never
	// actually chose.
	if p, ok := m.lookDigitPane(k); ok {
		m.exitLook()
		mdl, cmd := m.selectTab(p)
		return mdl, cmd, true
	}
	// tab/shift+tab are claimed at every layer — never falling through to
	// the global dock-cycle, which would silently change m.dockTab
	// underneath the borrow and break "prior state intact" on exit
	// (data-model.md invariant 4). From cursor focus they open the pane,
	// exactly like enter; from pane/drill focus they release one layer back
	// toward the cursor, the same direction esc releases (a second, wrist-
	// friendly way back that never needs a state-machine of its own: it is
	// esc's own transition, reused).
	if k == "tab" || k == "shift+tab" {
		switch m.lookFocus {
		case lookFocusCursor:
			m.focusTilePane()
		case lookFocusPane:
			m.lookFocus = lookFocusCursor
		case lookFocusDrill:
			m.lookFocus = lookFocusPane
			m.lookDrill = lookDrillRef{}
			m.lookDrillScroll = 0
		}
		return m, nil, true
	}

	switch m.lookFocus {
	case lookFocusCursor:
		return m.handleLookCursorKey(k)
	case lookFocusPane:
		return m.handleLookPaneKey(k)
	default: // lookFocusDrill
		return m.handleLookDrillKey(k)
	}
}

func (m Model) handleLookCursorKey(k string) (tea.Model, tea.Cmd, bool) {
	switch k {
	case "esc":
		m.exitLook()
	case "h", "left":
		m.moveLookCursor(-1, 0)
	case "l", "right":
		m.moveLookCursor(1, 0)
	case "k", "up":
		m.moveLookCursor(0, -1)
	case "j", "down":
		m.moveLookCursor(0, 1)
	case "H":
		m.moveLookCursor(-8, 0)
	case "L":
		m.moveLookCursor(8, 0)
	case "K":
		m.moveLookCursor(0, -8)
	case "J":
		m.moveLookCursor(0, 8)
	case "c":
		m.snapCameraToCursor()
	case "enter":
		m.focusTilePane()
	default:
		return m, nil, false
	}
	return m, nil, true
}

func (m Model) handleLookPaneKey(k string) (tea.Model, tea.Cmd, bool) {
	switch k {
	case "esc":
		m.lookFocus = lookFocusCursor
	case "j", "down":
		m.moveLookSel(1)
	case "k", "up":
		m.moveLookSel(-1)
	case "enter":
		m.drillSelectedRow()
	case "left", "right", "H", "L":
		// No meaning while the pane is focused — claimed only so these keys
		// never leak to the global free camera-pan behind this mode (the
		// AS1.5 "never the free camera-pan... outside the mode" reading
		// extended to every focus layer inside it), the same documented-
		// no-op shape as rowVillDetGG's "no effect here (roster-only)".
	default:
		return m, nil, false
	}
	return m, nil, true
}

func (m Model) handleLookDrillKey(k string) (tea.Model, tea.Cmd, bool) {
	switch k {
	case "esc":
		m.lookFocus = lookFocusPane
		m.lookDrill = lookDrillRef{}
		m.lookDrillScroll = 0
	case "J", "down":
		m.lookDrillScroll++ // clamped to content length at render time
	case "K", "up":
		if m.lookDrillScroll > 0 {
			m.lookDrillScroll--
		}
	case "left", "right", "H", "L":
		// Same documented no-op rationale as handleLookPaneKey above.
	default:
		return m, nil, false
	}
	return m, nil, true
}

// focusTilePane is `⏎`/`tab` from cursor focus (US3 AS1): the pane border
// draws amber (focus-contract rule 2 — panelFocusStyle at render time),
// selection clamped to the freshly-assembled row list.
func (m *Model) focusTilePane() {
	m.lookFocus = lookFocusPane
	if n := len(m.tileRows()); n > 0 {
		m.lookSel = clampInt(m.lookSel, 0, n-1)
	} else {
		m.lookSel = 0
	}
}

func (m *Model) moveLookSel(delta int) {
	rows := m.tileRows()
	if len(rows) == 0 {
		m.lookSel = 0
		return
	}
	m.lookSel = clampInt(m.lookSel+delta, 0, len(rows)-1)
}

// drillSelectedRow is `⏎` from pane focus (US3 AS2): opens the selected
// row's drill target, or is a documented no-op on a non-drillable row
// (terrain, the gru, a den) — the rowVillDetGG "no effect here" precedent.
func (m *Model) drillSelectedRow() {
	rows := m.tileRows()
	if m.lookSel < 0 || m.lookSel >= len(rows) {
		return
	}
	row := rows[m.lookSel]
	if row.drill.kind == lookDrillNone {
		return
	}
	m.lookFocus = lookFocusDrill
	m.lookDrill = row.drill
	m.lookDrillScroll = 0
}

// --- TILE view assembly (data-model.md "TILE view row model") ---

// registryMeaning resolves a structure/marker binding key's registry
// `meaning` prose directly (never through tileKey's ground-tile fallback,
// which would silently mislabel an unregistered kind) — the whatis source
// FR-010 requires.
func registryMeaning(key string) (string, bool) {
	if r, ok := keyTiles[key]; ok {
		return r.entry.Meaning, true
	}
	return "", false
}

// pileAtCursor returns the pile at the cursor tile, if any — at most one per
// tile (agents.go Pile: "the reducer merges drops onto an existing pile").
func (m Model) pileAtCursor() *sim.Pile {
	if m.replica == nil {
		return nil
	}
	for i := range m.replica.Piles {
		p := &m.replica.Piles[i]
		if p.X == m.lookX && p.Y == m.lookY {
			return p
		}
	}
	return nil
}

// agentTileLabel renders one agents-band row: name, status, current intent
// goal, and the same five need bars villagerIdentitySection shows.
func agentTileLabel(a sim.Agent) string {
	status := "awake"
	switch {
	case a.Dead:
		status = "dead"
	case a.Asleep:
		status = "asleep"
	}
	goal := "idle"
	if a.Intent != nil {
		goal = a.Intent.Goal
	}
	return fmt.Sprintf("%s · %s · %s · health %s food %s rest %s warmth %s morale %s",
		a.Name, status, goal, bar(a.Needs.Health), bar(a.Needs.Food), bar(a.Needs.Rest), bar(a.Needs.Warmth), bar(a.Needs.Morale))
}

// structureTileLabel renders one structures-band row — the registry
// `meaning` prose for every kind (SC-002's fourth surface: a row registered
// with no renderer edit reaches here through registryMeaning), with fire's
// three-state lit/dying/cold resolution (mirroring renderMapGrid's own
// FuelUntil/RefuelDyingBelow logic) and the wall family's damaged-state/HP
// detail special-cased on top.
func structureTileLabel(replica *sim.State, st sim.Structure) string {
	switch st.Kind {
	case "fire":
		switch {
		case replica != nil && replica.Tick < st.FuelUntil && st.FuelUntil-replica.Tick < replica.RefuelDyingBelow():
			return "a fire, dying — refuel it soon or it goes cold"
		case replica != nil && replica.Tick < st.FuelUntil:
			if meaning, ok := registryMeaning("fire"); ok {
				return meaning
			}
		default:
			if meaning, ok := registryMeaning("fire_cold"); ok {
				return meaning
			}
		}
		return "fire"
	case "wall_plank", "wall_stone":
		state := "intact"
		if st.HP < sim.WallMaxHP(st.Kind) {
			state = "damaged"
		}
		meaning, _ := registryMeaning(st.Kind)
		return fmt.Sprintf("%s (%s, %d hp)", meaning, state, st.HP)
	default:
		if meaning, ok := registryMeaning(st.Kind); ok {
			return meaning
		}
		return st.Kind
	}
}

// terrainRowMeaning resolves the TILE view's one terrain row — the same
// effective-kind resolution tile() (views.go, renderMapGrid) uses: a path
// overlay wins, then a quarried/depleted overlay, else the map's base
// terrain kind — read through the tile registry (FR-010) so a row added
// there reaches this surface with no edit here (SC-002's fourth surface).
func (m Model) terrainRowMeaning() string {
	if m.gameMap == nil {
		return "no terrain (world manifest missing?)"
	}
	x, y := m.lookX, m.lookY
	if m.replica != nil {
		for _, st := range m.replica.Structures {
			if st.Kind == "path" && st.X == x && st.Y == y {
				if meaning, ok := registryMeaning("path"); ok {
					return meaning
				}
			}
		}
		for _, q := range m.replica.Quarried {
			if q.X == x && q.Y == y {
				return terrainTile(worldmap.Depleted).entry.Meaning
			}
		}
	}
	return terrainTile(m.gameMap.At(x, y)).entry.Meaning
}

// isWaterTileAtCursor gates the "open water" terrain-flavor note (research
// R4: a registry-sourced note, never a sim claim) — checked against the
// map's base kind directly, since water is never overlaid (terrain.go
// effectiveKind's deliberate no-overlay-arm comment).
func (m Model) isWaterTileAtCursor() bool {
	return m.gameMap != nil && m.gameMap.At(m.lookX, m.lookY) == worldmap.Water
}

// tileEvents (research R5) filters m.events to entries whose subject-
// registry candidate carries an EXPLICIT recorded position equal to (x,y) —
// deliberately never resolveSubject's live-actor-preferring resolution: an
// event belongs to the tile where it was recorded, not wherever its actor
// stands now. Most-recent-first, bounded to a pane budget; never decodes an
// event type absent from subjectRegistry (the same bounded-work posture
// resolveSubject itself takes).
func (m Model) tileEvents(x, y int) []int {
	const budget = 20
	var out []int
	for i := len(m.events) - 1; i >= 0 && len(out) < budget; i-- {
		fn, ok := subjectRegistry[m.events[i].Type]
		if !ok {
			continue
		}
		cand, ok := fn(m.events[i])
		if !ok || !cand.hasPos {
			continue
		}
		if cand.x == x && cand.y == y {
			out = append(out, i)
		}
	}
	return out
}

// tileRows assembles the cursor tile's content in DF's fixed hierarchy
// (FR-009): agents (the gru joins this band when abroad here, non-
// drillable) → piles/chests → structures (+ dens) → terrain (exactly one
// row) → recent recorded-position events. Empty bands contribute no rows and
// no heading (spec edge case: an empty tile shows header + terrain only).
func (m Model) tileRows() []tileRow {
	x, y := m.lookX, m.lookY
	var rows []tileRow

	if m.replica != nil {
		for i, a := range m.replica.Agents {
			if a.X != x || a.Y != y {
				continue
			}
			rows = append(rows, tileRow{band: bandAgents, label: agentTileLabel(a), drill: lookDrillRef{kind: lookDrillAgent, idx: i}})
		}
		if g := m.replica.Gru; g != nil && g.X == x && g.Y == y {
			meaning, _ := registryMeaning("gru")
			rows = append(rows, tileRow{band: bandAgents, label: meaning})
		}

		if p := m.pileAtCursor(); p != nil {
			rows = append(rows, tileRow{band: bandPilesChests, label: "pile — " + summarizePileContents([]sim.Pile{*p}), drill: lookDrillRef{kind: lookDrillPile}})
		}
		for i, st := range m.replica.Structures {
			if st.X != x || st.Y != y {
				continue
			}
			switch st.Kind {
			case "chest":
				rows = append(rows, tileRow{band: bandPilesChests, label: "chest — " + describeChest(st, m.agentNames()), drill: lookDrillRef{kind: lookDrillChest, idx: i}})
			case "path":
				// Terrain-level (renderMapGrid's own precedence) — not a
				// structures-band row.
			default:
				rows = append(rows, tileRow{band: bandStructures, label: structureTileLabel(m.replica, st)})
			}
		}
	}
	if m.gameMap != nil {
		for _, d := range m.gameMap.Dens {
			if d.X == x && d.Y == y {
				meaning, _ := registryMeaning("den")
				rows = append(rows, tileRow{band: bandStructures, label: meaning})
			}
		}
	}

	rows = append(rows, tileRow{band: bandTerrain, label: m.terrainRowMeaning()})

	if m.replica != nil {
		names := m.agentNames()
		for _, idx := range m.tileEvents(x, y) {
			e := m.events[idx]
			l := formatChronicleLine(e, names, m.sk())
			label := fmt.Sprintf("%s · %s: %s", l.Time, l.Type, plainSegs(l.Summary))
			rows = append(rows, tileRow{band: bandEvents, label: label, drill: lookDrillRef{kind: lookDrillEvent, idx: idx}})
		}
	}
	return rows
}

func tileBandLabel(b tileBand) string {
	switch b {
	case bandAgents:
		return "agents"
	case bandPilesChests:
		return "piles & chests"
	case bandStructures:
		return "structures"
	case bandTerrain:
		return "terrain"
	case bandEvents:
		return "recent events"
	}
	return ""
}

// --- env header (FR-007, research R4) ---

// envLevel is a discrete 3-step reading — the LEVEL is the contract
// (SC-006 reads the sim truth, not the glyph); the meter is presentation.
type envLevel int

const (
	envLow envLevel = iota
	envMid
	envHigh
)

func envMeter(l envLevel) string {
	switch l {
	case envHigh:
		return "▮▮▮"
	case envMid:
		return "▮▮▯"
	default:
		return "▮▯▯"
	}
}

// envWarmthLevel maps an EnvSample + day/night onto the warmth level/note
// table (data-model.md): a fire or shelter source always reads warm
// regardless of time of day; absent a source, daylight is mild and night is
// cold — never a duplicated radius, purely a presentation mapping of
// sim.EnvAt's own fields.
func envWarmthLevel(s sim.EnvSample, night bool) (envLevel, string) {
	switch {
	case s.WarmSource == "fire":
		return envHigh, "warm — in a lit fire's radius"
	case s.WarmSource == "shelter":
		return envHigh, "warm — shelter cover"
	case !night:
		return envMid, "mild — daylight"
	default:
		return envLow, "cold — night, no cover"
	}
}

// envLightLevel maps Night/Lit/on-shelter onto the light level/note table —
// the gru-safety notes restate gruProtected honestly (light OR shelter,
// gru.go), the teaching payoff of exposing light at all (research R4).
func envLightLevel(s sim.EnvSample, night, onShelter bool) (envLevel, string) {
	switch {
	case !night:
		return envHigh, "bright — daylight"
	case s.Lit:
		return envMid, "lit — firelight (gru-safe)"
	case onShelter:
		return envLow, "dark — indoors (gru-safe: shelter)"
	default:
		return envLow, "dark"
	}
}

// tileEnvHeader renders the warmth/light meter lines (FR-007) — always
// exactly 2 lines (a "no world state" placeholder when there is no replica
// yet) so the TILE body's fixed header-line count never varies frame to
// frame (tileHitOrigin's headerLines constant depends on this).
func (m Model) tileEnvHeader() []string {
	if m.replica == nil {
		return []string{
			styleDim.Render("warmth  unknown — no world state"),
			styleDim.Render("light   unknown — no world state"),
		}
	}
	x, y := m.lookX, m.lookY
	sample := sim.EnvAt(m.replica, x, y, m.replica.Tick)
	onShelter := false
	for _, st := range m.replica.Structures {
		if st.Kind == "shelter" && st.X == x && st.Y == y {
			onShelter = true
			break
		}
	}
	night := m.replica.Night
	wLevel, wNote := envWarmthLevel(sample, night)
	lLevel, lNote := envLightLevel(sample, night, onShelter)
	if m.isWaterTileAtCursor() {
		wNote += " · open water"
	}
	return []string{
		fmt.Sprintf("warmth %s %s", envMeter(wLevel), wNote),
		fmt.Sprintf("light  %s %s", envMeter(lLevel), lNote),
	}
}

// --- TILE body rendering + hit region (research R6) ---

// tileHitHeaderLines is the count of fixed lines tileBody always emits
// before its row list — title(1) + tileEnvHeader (always 2) + blank(1).
// tileHitOrigin adds this to the borrow seam's own chrome offset so a mouse
// click's row maps to the exact tileRows() index the click landed on.
const tileHitHeaderLines = 4

// tileHitOrigin computes where the TILE pane's row list begins on screen —
// the chronicleHitOrigin sibling (research R6): pure chrome-row/column
// arithmetic per the layout currently active, recomputed fresh every frame
// (no cached geometry).
func (m Model) tileHitOrigin() (originX, originY, width int) {
	if !isWidescreen(m.width) {
		// narrowView: headerView(1) + tabsView(1) + blank(1) [+ lesson row
		// (2) if showing] + box top border(1), then this body's own header.
		row := 3
		if m.wantsLessonRow() {
			row += 2
		}
		return 2, row + 1 + tileHitHeaderLines, m.width
	}
	// widescreenView: headerView(1) + villager strip (0/1) + box top
	// border(1) + tab row(1) + divider(1), then this body's own header.
	rows := computeRows(m.height, m.wantsLessonRow())
	cols := computeColumns(m.width)
	return cols.MapCols + cols.Gutter + 2, headerRows + rows.VillagerStrip + 3 + tileHitHeaderLines, cols.DockCols
}

func (m Model) recordTileHit(rowIndex []int) {
	if m.tileHit == nil {
		return
	}
	originX, originY, width := m.tileHitOrigin()
	*m.tileHit = tileHitRegion{valid: true, originX: originX, originY: originY, width: width, rowIndex: append([]int(nil), rowIndex...)}
}

// tileBody is the TILE view's renderer — dockTabContent's borrow branch
// (views.go) and the narrow single-pane swap (FR-012) both call this.
// Drill focus swaps the WHOLE body to the drill renderer (content swap
// only, FR-008); cursor/pane focus share the header + banded row list, the
// pane's selection marker appearing only once focus has actually moved off
// the cursor.
func (m Model) tileBody(width, height int) string {
	if height < 3 {
		height = 3
	}
	if m.lookFocus == lookFocusDrill {
		return m.tileDrillBody(width, height)
	}

	meaning := m.terrainRowMeaning()
	lines := []string{styleHeader.Render(fmt.Sprintf("TILE (%d,%d)", m.lookX, m.lookY)) + " · " + meaning}
	lines = append(lines, m.tileEnvHeader()...)
	lines = append(lines, "")

	rows := m.tileRows()
	sel := 0
	if len(rows) > 0 {
		sel = clampInt(m.lookSel, 0, len(rows)-1)
	}

	var screenLines []string
	var rowIndex []int
	lastBand := tileBand(-1)
	for i, r := range rows {
		if r.band != lastBand {
			screenLines = append(screenLines, styleDim.Render(tileBandLabel(r.band)))
			rowIndex = append(rowIndex, -1)
			lastBand = r.band
		}
		cursor := "  "
		if m.lookFocus == lookFocusPane && i == sel {
			cursor = styleFeedSelect.Render("▌") + " "
		}
		screenLines = append(screenLines, cursor+clipLine(r.label, width-2))
		rowIndex = append(rowIndex, i)
	}

	budget := height - len(lines)
	if budget < 1 {
		budget = 1
	}
	start := 0
	if len(screenLines) > budget {
		selLine := 0
		for i, ri := range rowIndex {
			if ri == sel {
				selLine = i
				break
			}
		}
		start = selLine - budget/2
		if start < 0 {
			start = 0
		}
		if start+budget > len(screenLines) {
			start = len(screenLines) - budget
		}
		if start < 0 {
			start = 0
		}
	}
	end := start + budget
	if end > len(screenLines) {
		end = len(screenLines)
	}
	visLines := screenLines[start:end]
	visIndex := rowIndex[start:end]
	m.recordTileHit(visIndex) // spec 074 research R6: this frame's click geometry

	lines = append(lines, visLines...)
	for i, l := range lines {
		lines[i] = clipLine(l, width)
	}
	return strings.Join(lines, "\n")
}

// tileDrillBody renders the open drill-in target — reusing the existing
// renderer families rather than forking new ones (research R5's FR-020
// boundary): the villager-detail family for an agent row, the raw-JSON
// inspector family for an event row, and the same pile/chest contents
// summary the row itself showed for those.
func (m Model) tileDrillBody(width, height int) string {
	switch m.lookDrill.kind {
	case lookDrillAgent:
		if m.replica == nil || m.lookDrill.idx < 0 || m.lookDrill.idx >= len(m.replica.Agents) {
			return styleDim.Render("no such villager")
		}
		return m.villagerDetailBodyFor(m.replica.Agents[m.lookDrill.idx], width, height)
	case lookDrillEvent:
		if m.lookDrill.idx < 0 || m.lookDrill.idx >= len(m.events) {
			return styleDim.Render("no such event")
		}
		e := m.events[m.lookDrill.idx]
		names := m.agentNames()
		actionLabel := ""
		if actions := m.detailActions(e); len(actions) > 0 {
			actionLabel = actions[0].Label
		}
		return strings.Join(chronicleDetailPane(e, names, m.lookDrillScroll, width, height, actionLabel), "\n")
	case lookDrillPile:
		p := m.pileAtCursor()
		if p == nil {
			return styleDim.Render("no pile here")
		}
		header := styleHeader.Render(fmt.Sprintf("PILE (%d,%d)", p.X, p.Y))
		return header + "\n\n" + summarizePileContents([]sim.Pile{*p})
	case lookDrillChest:
		if m.replica == nil || m.lookDrill.idx < 0 || m.lookDrill.idx >= len(m.replica.Structures) {
			return styleDim.Render("no such chest")
		}
		st := m.replica.Structures[m.lookDrill.idx]
		header := styleHeader.Render(fmt.Sprintf("CHEST (%d,%d)", st.X, st.Y))
		return header + "\n\n" + describeChest(st, m.agentNames())
	default:
		return styleDim.Render("nothing selected")
	}
}

// --- mouse (research R6, US4) ---

func (m Model) recordMapHit(x0, y0, vw, vh int) {
	if m.mapHit == nil {
		return
	}
	originX, originY := m.mapHitOrigin()
	*m.mapHit = mapHitRegion{valid: true, originX: originX, originY: originY, x0: x0, y0: y0, vw: vw, vh: vh}
}

// mapHitOrigin computes the map grid's rendered top-left screen cell — the
// mapHitRegion counterpart to chronicleHitOrigin (research R6): pure
// chrome-row/column arithmetic, recomputed fresh every frame. Border(1) +
// Padding(0,1)(1 horizontally, 0 vertically) puts the glyph's first column
// 2 columns in, in both layouts.
func (m Model) mapHitOrigin() (originX, originY int) {
	if !isWidescreen(m.width) {
		row := 3 // headerView(1) + tabsView(1) + blank(1)
		if m.wantsLessonRow() {
			row += 2
		}
		return 2, row + 1 // + box top border
	}
	rows := computeRows(m.height, m.wantsLessonRow())
	return 2, headerRows + rows.VillagerStrip + 2 // + box top border + title row
}

// handleMouse routes tea.MouseMsg (research R6): guards (release+left,
// help/minibuffer no-ops) → TILE pane region (select row; a second click on
// the already-selected row drills) → map region (move the cursor / enter
// the mode) → the pre-existing chronicle path, unchanged. The chronicle
// path is unreachable during a borrow anyway (chronicleVisible's
// !lookActive guard makes m.inspecting() false), so there is no ambiguity
// between the TILE-pane branch and the chronicle branch.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionRelease || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}
	if m.helpOpen || m.mbFocused {
		return m, nil
	}
	if mdl, cmd, handled := m.handleTileHitClick(msg); handled {
		return mdl, cmd
	}
	if mdl, cmd, handled := m.handleMapHitClick(msg); handled {
		return mdl, cmd
	}
	// Reverse jump (spec 086 US5): strip glyph and roster row, after the
	// map/TILE regions (no overlap — the strip row sits above the map, the
	// roster inside the dock) and ahead of the chronicle path.
	if mdl, cmd, handled := m.handleStripHitClick(msg); handled {
		return mdl, cmd
	}
	if mdl, cmd, handled := m.handleRosterHitClick(msg); handled {
		return mdl, cmd
	}
	if !m.inspecting() {
		return m, nil
	}
	hit := m.chronHit
	if hit == nil || !hit.valid {
		return m, nil
	}
	row := msg.Y - hit.originY
	col := msg.X - hit.originX
	if row < 0 || row >= len(hit.rowEvent) || col < 0 || col >= hit.width {
		return m, nil
	}
	idx := hit.rowEvent[row]
	if idx < 0 || idx >= len(m.events) {
		return m, nil
	}
	m.chronSelected = idx
	m.chronDetailScroll = 0
	return m.jumpToSource()
}

// handleTileHitClick is US4 AS3: click a TILE-pane row to select it
// (acquiring pane focus if the cursor held it); click the already-selected
// row again to drill in.
func (m Model) handleTileHitClick(msg tea.MouseMsg) (tea.Model, tea.Cmd, bool) {
	if !m.lookActive {
		return m, nil, false
	}
	hit := m.tileHit
	if hit == nil || !hit.valid {
		return m, nil, false
	}
	row := msg.Y - hit.originY
	col := msg.X - hit.originX
	if row < 0 || row >= len(hit.rowIndex) || col < 0 || col >= hit.width {
		return m, nil, false
	}
	idx := hit.rowIndex[row]
	if idx < 0 {
		return m, nil, true // a band heading — inside the pane, consumed, not a row
	}
	if m.lookFocus == lookFocusPane && m.lookSel == idx {
		m.drillSelectedRow()
		return m, nil, true
	}
	m.lookFocus = lookFocusPane
	m.lookSel = idx
	return m, nil, true
}

// handleMapHitClick is US4 AS1/AS2: click a map tile to move the cursor
// there, entering the mode first if it was inactive.
func (m Model) handleMapHitClick(msg tea.MouseMsg) (tea.Model, tea.Cmd, bool) {
	hit := m.mapHit
	if hit == nil || !hit.valid || m.gameMap == nil {
		return m, nil, false
	}
	row := msg.Y - hit.originY
	col := msg.X - hit.originX
	if row < 0 || row >= hit.vh || col < 0 {
		return m, nil, false
	}
	tx := col / 2
	if tx >= hit.vw {
		return m, nil, false
	}
	x, y := hit.x0+tx, hit.y0+row
	if !m.lookActive {
		if !m.lookEntryAllowed() {
			return m, nil, false
		}
		m.enterLookAt(x, y)
		return m, nil, true
	}
	m.lookX = clampInt(x, 0, m.gameMap.W-1)
	m.lookY = clampInt(y, 0, m.gameMap.H-1)
	m.lookFocus = lookFocusCursor
	m.pushCameraToCursor()
	return m, nil, true
}

// --- reverse jump (spec 086 US5 — the operator-placed rider) ---

// handleStripHitClick: clicking a villager-strip glyph centers the map
// camera on that villager (centerCameraOn — pan-based, so `c` recenter, the
// panned map title, and follow-suspend behave per spec 049). Strip order ==
// roster order == replica.Agents order, so the glyph index IS the villager
// index; dead villagers jump to their grave coordinates (agents keep X,Y
// after death). Clicks on the count text, separators, or the …N overflow
// marker are no-ops; an empty replica records no region at all.
func (m Model) handleStripHitClick(msg tea.MouseMsg) (tea.Model, tea.Cmd, bool) {
	hit := m.stripHit
	if hit == nil || !hit.valid || msg.Y != hit.originY {
		return m, nil, false
	}
	if m.replica == nil {
		return m, nil, false
	}
	for i, x := range hit.glyphX {
		if msg.X == x && i < len(m.replica.Agents) {
			a := m.replica.Agents[i]
			m.centerCameraOn(a.X, a.Y)
			return m, nil, true
		}
	}
	return m, nil, false
}

// handleRosterHitClick: clicking a villagers-tab roster row selects that
// row AND centers the camera on that villager — the chronicle click-line
// select+act precedent; in the narrow layout the active pane switches to
// the map so the jump's effect is visible (the jumpToSource FR-007
// precedent). Heading/blank lines inside the roster are consumed but jump
// nothing.
func (m Model) handleRosterHitClick(msg tea.MouseMsg) (tea.Model, tea.Cmd, bool) {
	hit := m.rosterHit
	if hit == nil || !hit.valid {
		return m, nil, false
	}
	row := msg.Y - hit.originY
	col := msg.X - hit.originX
	if row < 0 || row >= len(hit.rowAgent) || col < 0 || col >= hit.width {
		return m, nil, false
	}
	idx := hit.rowAgent[row]
	if idx < 0 {
		return m, nil, true // heading / band spacer — inside the roster, consumed
	}
	if m.replica == nil || idx >= len(m.replica.Agents) {
		return m, nil, true
	}
	m.villSelected = idx
	a := m.replica.Agents[idx]
	m.centerCameraOn(a.X, a.Y)
	if !isWidescreen(m.width) {
		m.active = paneMap // land where the jump's effect is visible (FR-007 precedent)
	}
	return m, nil, true
}
