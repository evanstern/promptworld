package tui

// Village lens completion tests (spec 060, TASK-129): the villager strip
// (US1) and the three map condition overlays (US2) — needs-critical,
// suppressed-mind, dying-fire.

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/evanstern/promptworld/internal/sim"
)

// healthyNeeds is well clear of every danger band (agents.go/guardian.go) —
// used by fixtures that don't want to accidentally trip needsCritical, which
// a bare sim.Agent{} zero-value would (0 is lethal territory on every need).
var healthyNeeds = sim.Needs{Health: 1000, Food: 1000, Rest: 1000, Warmth: 1000, Morale: 1000}

// --- US1: the villager strip ---

// TestVillagerStripViewBasic (FR-001, AC1): "N villagers" + the glyph run in
// roster order, styled exactly like the map's own agent glyphs.
func TestVillagerStripViewBasic(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = []sim.Agent{
		{Name: "Ash", Needs: healthyNeeds},
		{Name: "Birch", Asleep: true, Needs: healthyNeeds},
		{Name: "Cedar", Dead: true},
	}
	got := m.villagerStripView(80)
	want := "3 villagers " + styleAgent.Render("A") + " " + styleAsleep.Render("b") + " " + styleErr.Render("†")
	if got != want {
		t.Errorf("villagerStripView = %q, want %q", got, want)
	}
}

// TestVillagerStripViewEmptyRoster: no replica agents yet — just the count,
// no trailing space/glyph run artifact.
func TestVillagerStripViewEmptyRoster(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = nil
	got := m.villagerStripView(80)
	if got != "0 villagers" {
		t.Errorf("empty roster strip = %q, want %q", got, "0 villagers")
	}
}

// TestVillagerStripViewOverflowDropsFromEnd (FR-001 AC2): more villagers
// than the width allows sheds from the END with a trailing "…N" overflow
// count — never a mid-glyph truncation. The exact expected string pins the
// packing arithmetic (villagerStripView's backward-fit search).
func TestVillagerStripViewOverflowDropsFromEnd(t *testing.T) {
	m := widescreenModel(t)
	names := []string{"Ash", "Birch", "Cedar", "Dale", "Ember", "Fox", "Gale", "Holly", "Iris", "Jay"}
	agents := make([]sim.Agent, len(names))
	for i, n := range names {
		agents[i] = sim.Agent{Name: n, Needs: healthyNeeds}
	}
	m.replica.Agents = agents
	got := m.villagerStripView(20)
	want := "10 villagers A B " + styleDim.Render("…8")
	if got != want {
		t.Errorf("villagerStripView(20) = %q, want %q", got, want)
	}
}

// TestVillagerStripViewNeverExceedsWidth sweeps a range of widths (including
// pathologically narrow ones) with a large roster and checks the rendered
// strip never exceeds its budget and never wraps to a second line — the
// same discipline TestGuardianStripViewNeverExceedsWidth pins for the
// guardian strip.
func TestVillagerStripViewNeverExceedsWidth(t *testing.T) {
	m := widescreenModel(t)
	agents := make([]sim.Agent, 15)
	for i := range agents {
		agents[i] = sim.Agent{Name: sim.AgentNames[i%len(sim.AgentNames)], Needs: healthyNeeds}
	}
	m.replica.Agents = agents
	for _, w := range []int{1, 2, 5, 10, 20, 40, 80, 200} {
		got := m.villagerStripView(w)
		if strings.Contains(got, "\n") {
			t.Errorf("width %d: strip must never wrap to a second row: %q", w, got)
		}
		if lipgloss.Width(got) > w {
			t.Errorf("width %d: rendered strip width %d exceeds budget: %q", w, lipgloss.Width(got), got)
		}
	}
}

// TestVillagerStripRosterParity (SC-001): the strip's glyph run matches the
// villagers-tab roster 1:1 — same order, same awake/asleep/dead read — for
// a mixed fixture.
func TestVillagerStripRosterParity(t *testing.T) {
	m := widescreenModel(t)
	m.replica.Agents = []sim.Agent{
		{Name: "Ash", Needs: healthyNeeds},
		{Name: "Birch", Asleep: true, Needs: healthyNeeds},
		{Name: "Cedar", Dead: true},
		{Name: "Dale", Needs: healthyNeeds},
	}
	strip := m.villagerStripView(200)
	if !strings.HasPrefix(strip, "4 villagers ") {
		t.Fatalf("strip count should match roster length: %q", strip)
	}
	for i, a := range m.replica.Agents {
		roster := m.villagerRosterBody(200, 200)
		if !strings.Contains(roster, a.Name) {
			t.Fatalf("fixture sanity: roster missing %q", a.Name)
		}
		want := villagerStripGlyph(a)
		if !strings.Contains(strip, want) {
			t.Errorf("agent %d (%s): strip missing its roster-matching glyph %q: %q", i, a.Name, want, strip)
		}
	}
}

// --- villagerCountBadge: the strip's fold-relocation form ---

// TestVillagerCountBadgeAbsentWhenStripShowing: a tall widescreen terminal
// keeps the strip on its own row — no redundant header badge.
func TestVillagerCountBadgeAbsentWhenStripShowing(t *testing.T) {
	m := widescreenModel(t)
	m.width, m.height = 140, 40
	m.replica.Agents = []sim.Agent{{Name: "Ash", Needs: healthyNeeds}}
	if got := m.villagerCountBadge(); got != "" {
		t.Errorf("badge should be absent while the strip itself is showing, got %q", got)
	}
}

// TestVillagerCountBadgeShownWhenStripFolds (patterns/layout.md ruling a
// step 2): a short widescreen terminal folds the strip — the header badge
// picks it up instead.
func TestVillagerCountBadgeShownWhenStripFolds(t *testing.T) {
	m := widescreenModel(t)
	m.width, m.height = 140, 16 // folds the villager strip (layout_test.go TestComputeRows)
	m.replica.Agents = []sim.Agent{{Name: "Ash", Needs: healthyNeeds}}
	got := m.villagerCountBadge()
	want := styleDim.Render("[1 villagers]")
	if got != want {
		t.Errorf("folded-strip badge = %q, want %q", got, want)
	}
}

// TestVillagerCountBadgeShownInNarrow (patterns/layout.md ruling b: "NOT
// carried" — narrow shows the badge form only).
func TestVillagerCountBadgeShownInNarrow(t *testing.T) {
	m := testModel(t) // narrow (80 cols)
	m.replica.Agents = []sim.Agent{{Name: "Ash", Needs: healthyNeeds}, {Name: "Birch", Needs: healthyNeeds}}
	got := m.villagerCountBadge()
	want := styleDim.Render("[2 villagers]")
	if got != want {
		t.Errorf("narrow badge = %q, want %q", got, want)
	}
}

// TestVillagerCountBadgeAbsentWithNoRoster: nothing to count yet (pre-
// connect/empty world) — no badge at all, in either layout.
func TestVillagerCountBadgeAbsentWithNoRoster(t *testing.T) {
	for _, m := range []Model{testModel(t), widescreenModel(t)} {
		m.replica.Agents = nil
		if got := m.villagerCountBadge(); got != "" {
			t.Errorf("empty-roster badge should be absent, got %q", got)
		}
	}
}

// --- US2: map condition overlays ---

// TestNeedsCritical pins the exact thresholds reused from sim's own danger
// bands (agents.go/guardian.go) — the roster gauges' existing critical
// vocabulary, not a new tuning surface.
func TestNeedsCritical(t *testing.T) {
	cases := []struct {
		name string
		n    sim.Needs
		want bool
	}{
		{"all healthy", healthyNeeds, false},
		{"health at floor", sim.Needs{Health: sim.SurvivalNearDeathBelow - 1, Food: 1000, Rest: 1000, Warmth: 1000}, true},
		{"health at rearm, not critical", sim.Needs{Health: sim.SurvivalNearDeathBelow, Food: 1000, Rest: 1000, Warmth: 1000}, false},
		{"food critical", sim.Needs{Health: 1000, Food: sim.SurvivalStarvingRearm - 1, Rest: 1000, Warmth: 1000}, true},
		{"warmth critical", sim.Needs{Health: 1000, Food: 1000, Rest: 1000, Warmth: sim.SurvivalFreezingRearm - 1}, true},
		{"rest critical", sim.Needs{Health: 1000, Food: 1000, Rest: sim.DangerRestBelow - 1, Warmth: 1000}, true},
		{"morale alone never critical (no sim danger band exists for it)", sim.Needs{Health: 1000, Food: 1000, Rest: 1000, Warmth: 1000, Morale: 0}, false},
	}
	for _, c := range cases {
		if got := needsCritical(c.n); got != c.want {
			t.Errorf("%s: needsCritical(%+v) = %v, want %v", c.name, c.n, got, c.want)
		}
	}
}

// TestAgentSuppressedMindClearsOnLaterNonSuppressedOutcome (US2 AS2): the
// latest chain's Suppressed flag is what the map reads, and it clears the
// moment a later, non-suppressed outcome lands for that agent.
func TestAgentSuppressedMindClearsOnLaterNonSuppressedOutcome(t *testing.T) {
	m := testModel(t)
	m.replica.Agents = []sim.Agent{{Name: "Ash", Needs: healthyNeeds}}

	if m.agentSuppressedMind(0) {
		t.Fatal("no chains yet — must not be marked suppressed")
	}

	m.applyEvent(outcomeEvent(1, "meeting-0-900", "meeting", 0, sim.OutcomeSuppressed, "budget exhausted"))
	if !m.agentSuppressedMind(0) {
		t.Fatal("a suppression-only chain should mark the agent suppressed")
	}

	// A later, ordinary cognition (thought + call + landed outcome) clears it.
	m.applyEvent(thoughtEvent(2, 1000, "reflex-0-1000", "reflex", 0, 0))
	m.applyEvent(toolCallEvent(3, "reflex-0-1000", 1, "gather", "landed", ""))
	m.applyEvent(outcomeEvent(4, "reflex-0-1000", "reflex", 0, "landed", ""))
	if m.agentSuppressedMind(0) {
		t.Error("a later non-suppressed outcome should clear the suppressed-mind mark")
	}
}

// --- renderMapGrid overlay integration ---

// TestRenderMapGridNeedsCriticalOverlay (US2 AS1): a living agent with a
// need in its danger band renders in the needs-critical style; recovery
// clears it next frame.
func TestRenderMapGridNeedsCriticalOverlay(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	m := testModel(t)
	cx, cy := m.gameMap.W/2, m.gameMap.H/2
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: cx, Y: cy, Needs: sim.Needs{Health: 1000, Food: 100, Rest: 1000, Warmth: 1000}}}

	grid, _ := m.renderMapGrid(10, 10)
	if !strings.Contains(grid, styleAgentCritical.Render("A")) {
		t.Errorf("starving agent should render in the needs-critical style: %q", grid)
	}

	m.replica.Agents[0].Needs.Food = 1000
	grid, _ = m.renderMapGrid(10, 10)
	if strings.Contains(grid, styleAgentCritical.Render("A")) {
		t.Error("recovered needs should clear the needs-critical style next frame")
	}
	if !strings.Contains(grid, styleAgent.Render("A")) {
		t.Errorf("recovered agent should render plainly: %q", grid)
	}
}

// TestRenderMapGridSuppressedMindOverlay (US2 AS2): a living agent whose
// latest decision outcome was a router suppression renders in the
// suppressed-mind style.
func TestRenderMapGridSuppressedMindOverlay(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	m := testModel(t)
	cx, cy := m.gameMap.W/2, m.gameMap.H/2
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: cx, Y: cy, Needs: healthyNeeds}}
	m.applyEvent(outcomeEvent(1, "meeting-0-900", "meeting", 0, sim.OutcomeSuppressed, "budget exhausted"))

	grid, _ := m.renderMapGrid(10, 10)
	if !strings.Contains(grid, styleAgentSuppressed.Render("A")) {
		t.Errorf("suppressed-mind agent should render in the suppressed-mind style: %q", grid)
	}
}

// TestRenderMapGridConditionPriorityNeedsCriticalWins (US2 AS4): needs-
// critical wins over suppressed-mind when both hold (physical danger over
// cognitive telemetry).
func TestRenderMapGridConditionPriorityNeedsCriticalWins(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	m := testModel(t)
	cx, cy := m.gameMap.W/2, m.gameMap.H/2
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: cx, Y: cy, Needs: sim.Needs{Health: 1000, Food: 50, Rest: 1000, Warmth: 1000}}}
	m.applyEvent(outcomeEvent(1, "meeting-0-900", "meeting", 0, sim.OutcomeSuppressed, "budget exhausted"))

	if !needsCritical(m.replica.Agents[0].Needs) || !m.agentSuppressedMind(0) {
		t.Fatal("fixture setup: agent should be both needs-critical and suppressed-mind")
	}
	grid, _ := m.renderMapGrid(10, 10)
	if !strings.Contains(grid, styleAgentCritical.Render("A")) {
		t.Errorf("needs-critical should win when both conditions hold: %q", grid)
	}
	if strings.Contains(grid, styleAgentSuppressed.Render("A")) {
		t.Errorf("suppressed-mind style must not render when needs-critical also applies: %q", grid)
	}
}

// TestRenderMapGridDeadAgentNeverGetsConditionOverlay: a dead agent renders
// plain † regardless of its frozen needs (which are typically 0/critical at
// death) — overlays only ever apply to the living (edge case: "map overlays
// never apply to the dead — no needs, no mind").
func TestRenderMapGridDeadAgentNeverGetsConditionOverlay(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	m := testModel(t)
	cx, cy := m.gameMap.W/2, m.gameMap.H/2
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: cx, Y: cy, Dead: true}}
	grid, _ := m.renderMapGrid(10, 10)
	if !strings.Contains(grid, styleErr.Render("†")) {
		t.Errorf("dead agent should render the plain dead marker, no overlay: %q", grid)
	}
	if strings.Contains(grid, styleAgentCritical.Render("A")) || strings.Contains(grid, styleAgentSuppressed.Render("A")) {
		t.Errorf("a dead agent must never carry a condition overlay: %q", grid)
	}
}

// TestRenderMapGridDyingFireOverlay (US2 AS3): a fire inside the sim's own
// dying-fuel window (State.RefuelDyingBelow) renders in the warn style —
// still lit, distinct from both plain-lit and cold; refueling clears it.
func TestRenderMapGridDyingFireOverlay(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	m := testModel(t)
	cx, cy := m.gameMap.W/2, m.gameMap.H/2
	m.replica.Tick = 100000
	window := m.replica.RefuelDyingBelow()

	m.replica.Structures = []sim.Structure{{Kind: "fire", X: cx, Y: cy, FuelUntil: m.replica.Tick + window/2}}
	grid, _ := m.renderMapGrid(10, 10)
	if !strings.Contains(grid, styleFireDying.Render("▲")) {
		t.Errorf("a fire inside the dying window should render in the warn style: %q", grid)
	}
	if strings.Contains(grid, styleFire.Render("▲")) {
		t.Errorf("a dying fire must not also render as plain lit: %q", grid)
	}

	// Comfortably lit: well outside the dying window.
	m.replica.Structures[0].FuelUntil = m.replica.Tick + window*10
	grid, _ = m.renderMapGrid(10, 10)
	if !strings.Contains(grid, styleFire.Render("▲")) {
		t.Errorf("a comfortably-fueled fire should render plain lit: %q", grid)
	}
	if strings.Contains(grid, styleFireDying.Render("▲")) {
		t.Errorf("a comfortably-fueled fire must not render as dying: %q", grid)
	}

	// Refueling (pushing FuelUntil back out) clears the dying style, and
	// running out entirely goes cold, not dying.
	m.replica.Structures[0].FuelUntil = m.replica.Tick - 1
	grid, _ = m.renderMapGrid(10, 10)
	if !strings.Contains(grid, styleFireCold.Render("△")) {
		t.Errorf("a fire past FuelUntil should render cold, not dying: %q", grid)
	}
}

// TestMapLegendNamesConditionOverlays (US2 AS5, FR-003): the legend
// discoverability requirement — the compact in-game legend and the help
// overlay's walkthrough both name the new marker styles.
func TestMapLegendNamesConditionOverlays(t *testing.T) {
	m := testModel(t)
	cx, cy := m.gameMap.W/2, m.gameMap.H/2
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: cx, Y: cy, Needs: healthyNeeds}}
	_, legend := m.renderMapGrid(10, 10)
	if !strings.Contains(legend, conditionOverlayNote) {
		t.Errorf("compact legend should name the condition overlays: %q", legend)
	}
	walkthrough := strings.Join(Model{}.helpWalkthroughLines(200), "\n")
	if !strings.Contains(walkthrough, conditionOverlayNote) {
		t.Error("help overlay walkthrough should name the condition overlays too")
	}
}

// TestConditionOverlayStylesDistinctColorProfile mirrors
// TestFamilyTintDistinctPerFamily's discipline (render_test.go): every new
// overlay style must actually render distinguishably from its neighbors
// under a forced color profile, not just exist as a separate token.
func TestConditionOverlayStylesDistinctColorProfile(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)

	agentStyles := map[string]lipgloss.Style{
		"awake":           styleAgent,
		"asleep":          styleAsleep,
		"dead":            styleErr,
		"needs-critical":  styleAgentCritical,
		"suppressed-mind": styleAgentSuppressed,
	}
	seen := map[string]string{}
	for name, st := range agentStyles {
		rendered := st.Render("A")
		if !strings.Contains(rendered, "\x1b") {
			t.Errorf("%s: style produced no ANSI under a forced color profile: %q", name, rendered)
		}
		if prior, ok := seen[rendered]; ok {
			t.Errorf("%s renders identically to %s (%q) — must be distinguishable", name, prior, rendered)
		}
		seen[rendered] = name
	}

	fireStyles := map[string]lipgloss.Style{
		"lit":   styleFire,
		"dying": styleFireDying,
		"cold":  styleFireCold,
	}
	seenFire := map[string]string{}
	for name, st := range fireStyles {
		rendered := st.Render("▲")
		if prior, ok := seenFire[rendered]; ok {
			t.Errorf("fire %s renders identically to %s (%q) — must be distinguishable", name, prior, rendered)
		}
		seenFire[rendered] = name
	}
}
