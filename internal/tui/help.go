package tui

// Help overlay content (spec 045-tui-help-overlay, TASK-116). Everything in
// this file is static data + pure rendering, the digest.go convention: no
// daemon/IPC/event/world-state reads, so the overlay is byte-identical on a
// no-LLM world by construction (R6, SC-004). The mode key tables are R3's
// "derived not duplicated" substrate — cross-checked against the real
// dispatch code by help_test.go's keymap sweep (FR-003/SC-003); the glyph
// table (below) is renderMapGrid's own legend source, so the map's key line
// and this file's glyph walkthrough page can never silently diverge
// (FR-005).

import (
	"fmt"
	"strings"
)

// --- T002: the shared map-glyph table (FR-005 anti-drift substrate) ---

// glyphEntry is one row of the map's glyph key. Glyph+Name concatenate with
// no separator to reproduce renderMapGrid's existing compact legend token
// (e.g. "~water", "▤▩wall" — two glyphs sharing one entry since the plank/
// stone walls have always rendered as one legend token); Meaning is the
// plain-language sentence the overlay's glyph walkthrough page renders
// instead (US2, FR-004).
type glyphEntry struct {
	Glyph   string
	Name    string
	Meaning string
}

// mapGlyphs is renderMapGrid's (views.go) legend line rendered *from* this
// table (legendGlyphLine below) — one source, so a glyph added here reaches
// both the compact in-game legend and the overlay's full walkthrough
// automatically (FR-005/SC-003; research.md R3/R8: this is also the seam
// that picks up spec 044's grave glyph, whenever it lands, without a second
// edit). The "G" gru row is a real gap this feature closes: the glyph has
// always been drawn (views.go tile(), styleGru) but was never in the legend
// text — SC-002 requires the overlay to decode 100% of what the map can
// draw, so it's added here (and, by construction, to the compact legend too).
//
// The "✝" grave row (spec 044 US4, ratified follow-up): every post-044
// death's grave shares its tile with the dead agent it belongs to, and
// renderMapGrid's tile() priority normally lets the agent glyph win a shared
// tile (the same rule the chest/path rows rely on) — so without a carve-out
// this glyph would advertise something the map could never actually show.
// The carve-out: a dead agent AT a tile that also holds a grave renders the
// grave glyph instead of the plain dead marker (the body becomes the grave);
// a graveless dead agent (pre-044 replay/history) is unaffected.
var mapGlyphs = []glyphEntry{
	{"~", "water", "water — impassable to foot travel"},
	{"♠", "wood", "a tree — choppable for wood"},
	{"\"", "forage", "wild forage — gatherable food"},
	{"^", "rock", "an intact rock outcrop — quarriable for stone"},
	{",", "quarried", "a depleted outcrop — passable, already quarried"},
	{"ᴥ", "den", "a gru's den"},
	{"▲", "fire", "a lit fire — warmth, cooking"},
	{"△", "cold", "a cold fire, out of fuel"},
	{"⌂", "shelter", "a built shelter"},
	{"▣", "oven", "a built oven"},
	{"%", "pile", "a ground stockpile of dropped goods"},
	{"☐", "chest", "a built chest, holding goods"},
	{"▤▩", "wall", "a built wall (▤ plank, ▩ stone); dim = damaged"},
	{"·", "path", "a paved path (tan) — distinct from plain ground's dim ·"},
	{"G", "gru", "the gru — a predator; approach at your peril"},
	{"✝", "grave", "a villager's grave — marks where a death occurred"},
}

// agentGlyphNote / mapControlNote are the legend's trailing free-text
// clauses — not table rows (they describe a per-agent-name convention and a
// control hint, not a single fixed glyph), but shared verbatim by
// renderMapGrid's legend and the overlay's glyph walkthrough page so the two
// can't drift on these either.
const (
	agentGlyphNote = "agents by initial (lowercase asleep, †dead)"
	mapControlNote = "arrows pan, c center"
)

// legendGlyphLine renders mapGlyphs compactly ("~water ♠wood ..." — T002):
// the exact token sequence renderMapGrid's legend line has always shown,
// plus the "Ggru" token this feature adds (see mapGlyphs' gru row comment).
func legendGlyphLine() string {
	parts := make([]string, len(mapGlyphs))
	for i, g := range mapGlyphs {
		parts[i] = g.Glyph + g.Name
	}
	return strings.Join(parts, " ")
}

// --- T009: screen walkthrough content (US2) ---

// headerAnatomyRow is one element or conditional badge headerView (views.go)
// can render.
type headerAnatomyRow struct {
	Element string
	Meaning string
}

// headerAnatomy enumerates every element headerView (views.go:99-135) can
// render, including every conditional badge — not just the ones currently
// visible (spec.md US2-AS2) — so the walkthrough decodes a header the
// player has never actually seen yet (budget exhausted, a governed speed,
// a provider outage) just as well as the common case.
var headerAnatomy = []headerAnatomyRow{
	{"world name", "the world's name, always shown"},
	{"disconnected", "the client lost its socket connection; it retries automatically, showing the last error"},
	{"tick / game time", "the current simulated tick and in-world clock time"},
	{"running / PAUSED", "the clock's state — PAUSED renders amber"},
	{"speed (t/s)", "the requested speed step and its effective ticks/second"},
	{"asked … in flight, debt …%", "governed speed: the operator asked for more speed than the governor currently allows"},
	{"[degraded]", "the clock's effective rate has fallen behind its target"},
	{"[llm: provider kind]", "the first LLM provider carrying an active health condition (see the guardian tab for the rest)"},
	{"[suppressed: classes]", "cognition classes currently being skipped at this speed (see the guardian tab for remedies)"},
}

// dockTabEntry is one dock tab's key/name/purpose (from paneNames/
// dockTabKey, tui.go:47-50).
type dockTabEntry struct {
	Key     string
	Name    string
	Purpose string
}

// dockTabEntries resolves the dock-tab walkthrough rows through the world
// skin (spec 052 FR-007): the guardian tab's label and epithet are skin
// data; the other tabs — the systems tab by D10 design — are non-fiction
// chrome.
func (m Model) dockTabEntries() []dockTabEntry {
	return []dockTabEntry{
		{dockTabKey[paneChronicle], paneNames[paneChronicle], "the event feed — narrated story or raw log; pauses into inspect mode when the clock is paused"},
		{dockTabKey[paneGuardian], m.paneName(paneGuardian), "the " + m.sk().Epithet() + "'s transcript and standing orders (fiction-layer content only)"},
		{dockTabKey[paneVillagers], paneNames[paneVillagers], "the roster — select a villager for full detail, decisions, and inventory"},
		{dockTabKey[paneSystems], paneNames[paneSystems], "engine telemetry — LLM provider health, spend, and the cognition horizon (spec 053, never skinned)"},
	}
}

// --- T012: the lessons pull-reference seam (US4, FR-010, SC-006) ---

// helpLesson is the whole seam contract (contracts/help-content.md "The
// lessons seam"): the future first-occurrence lesson projection appends
// entries here — id/title/body — and the lessons section renders them, with
// zero changes to this file's navigation or rendering code. Empty today;
// helpLessonsLines' placeholder line is what renders until the first entry
// lands.
type helpLesson struct {
	ID    string
	Title string
	Body  string
}

// helpLessons ships empty with this feature (US4's obligation is the seam,
// not lesson content) — append-only once the future feature lands.
var helpLessons []helpLesson

// --- T004/US1: mode key tables ---

// helpModeKey identifies which mode's key page the overlay is showing —
// data-model.md's enumeration: global/home, minibuffer (content-only, since
// '?' types there instead of opening the overlay — FR-001), inspect,
// villagers roster, villagers detail (folded in with its decisions
// sub-view — neither the spec's mode list nor keymap.md splits them
// further), and solo zoom/narrow fallback (one page: both states drive the
// identical handleGlobalKey, differing only in which keys their footer
// hints hardcode as "basic").
type helpModeKey int

const (
	helpModeGlobal helpModeKey = iota
	helpModeMinibuffer
	helpModeInspect
	helpModeVillagersRoster
	helpModeVillagersDetail
	helpModeSolo
	helpModeCount
)

// nextHelpMode / prevHelpMode cycle every mode page in a fixed loop — the
// FR-001 seam: minibuffer's page is otherwise unreachable (its own '?'
// opens nothing, R1), so paging with 'n'/'p' from any other mode's overlay
// is the only way to read it.
func nextHelpMode(m helpModeKey) helpModeKey { return (m + 1) % helpModeCount }
func prevHelpMode(m helpModeKey) helpModeKey { return (m - 1 + helpModeCount) % helpModeCount }

// helpKeyRow is one real, working binding shown on a mode's page. Keys lists
// every literal msg.String() form the row covers (e.g. all four arrow keys
// in one row) — help_test.go's keymap sweep (SC-003) flattens Basic+Advanced
// across a mode's page and cross-checks the result against that mode's real
// dispatch code.
type helpKeyRow struct {
	Label  string
	Keys   []string
	Action string
}

// helpModePage is one mode's two-tier key listing (FR-002): Basic is the
// footer-hinted set (Assumptions: "basic tier ≈ footer-hinted keys");
// Advanced is everything else that mode's handler chain really recognizes
// (FR-003: every listed binding works, every working binding listed
// exactly once).
type helpModePage struct {
	Title    string
	Basic    []helpKeyRow
	Advanced []helpKeyRow
}

// --- shared rows: handleGlobalKey's switch (tui.go:544-637), one row per
// case label. helpModeGlobal and helpModeSolo are both driven by this exact
// handler (only which keys each mode's footer calls "basic" differs) so
// they partition the same row set between their two tiers rather than
// duplicating the Action text twice.
var (
	rowHome      = helpKeyRow{"1", []string{"1"}, "return to the map (exits solo, if solo'd)"}
	rowTab2      = helpKeyRow{"2", []string{"2"}, "select the chronicle tab; press again to solo it (or return home if already solo'd)"}
	rowTab3      = helpKeyRow{"3", []string{"3"}, "select the guardian tab; press again to solo it (or return home if already solo'd)"}
	rowTab4      = helpKeyRow{"4", []string{"4"}, "select the villagers tab; press again to solo it (or return home if already solo'd)"}
	rowTab5      = helpKeyRow{"5", []string{"5"}, "select the systems tab; press again to solo it (or return home if already solo'd)"}
	rowDockCycle = helpKeyRow{"tab/shift+tab", []string{"tab", "shift+tab"}, "cycle the dock tabs (alias for 2/3/4/5)"}
	// Spec 054: the exercise tab exists only on scenario worlds — on ambient
	// worlds 6 is inert (the key stays documented; the tab is world-shaped).
	rowTab6     = helpKeyRow{"6", []string{"6"}, "select the exercise tab (scenario worlds only; inert elsewhere); press again to solo it"}
	rowConsole  = helpKeyRow{"G", []string{"G"}, "open the guardian console — a full-screen page for the conversation, charter/skills, and $EDITOR (G/1/esc closes)"}
	rowAsk      = helpKeyRow{"m", []string{"m"}, "focus the minibuffer — ask the guardian"}
	rowPause    = helpKeyRow{"space", []string{" "}, "pause / resume the clock"}
	rowSpeed    = helpKeyRow{"[ ]", []string{"[", "]"}, "speed down / up"}
	rowPan      = helpKeyRow{"←↑↓→", []string{"up", "down", "left", "right"}, "pan the map"}
	rowRecenter = helpKeyRow{"c", []string{"c"}, "recenter the camera on the wanderers"}
	rowChronA   = helpKeyRow{"a", []string{"a"}, "chronicle: filter by agent (chronicle tab only)"}
	rowChronT   = helpKeyRow{"t", []string{"t"}, "chronicle: filter by thread (chronicle tab only)"}
	rowChronR   = helpKeyRow{"r", []string{"r"}, "chronicle: toggle raw ↔ narrated (chronicle tab only)"}
	rowExitSolo = helpKeyRow{"esc", []string{"esc"}, "exit solo zoom (no effect when already home)"}
	rowQuit     = helpKeyRow{"q", []string{"q"}, "quit"}
	rowNarrowMB = helpKeyRow{"enter", []string{"enter"}, "narrow fallback: focus the minibuffer from the guardian pane (no effect in the widescreen composite)"}
)

// globalRows is the complete flattened set the two rows above partition —
// help_test.go asserts helpModeGlobal and helpModeSolo each union back to
// exactly this.
var globalRows = []helpKeyRow{
	rowHome, rowTab2, rowTab3, rowTab4, rowTab5, rowTab6, rowDockCycle, rowConsole, rowAsk, rowPause, rowSpeed,
	rowPan, rowRecenter, rowChronA, rowChronT, rowChronR, rowExitSolo, rowQuit, rowNarrowMB,
}

// --- minibuffer-only rows (handleMinibufferKey, tui.go:724-767) ---
var (
	rowMbEsc  = helpKeyRow{"esc", []string{"esc"}, "release focus — back to global keys"}
	rowMbSend = helpKeyRow{"⏎", []string{"enter"}, "send (empty buffer: release focus instead)"}
	rowMbHist = helpKeyRow{"↑↓", []string{"up", "down"}, "step through input history"}
	rowMbType = helpKeyRow{"any printable key", []string{" "}, "append to the buffer, visibly — never hijacked, even '?'"}
	rowMbBS   = helpKeyRow{"backspace", []string{"backspace"}, "delete the last character"}
)

// --- inspect-only rows (handleInspectKey, tui.go:801-827) ---
var (
	rowInspSel   = helpKeyRow{"j/k", []string{"j", "k"}, "select next / previous event (resets detail scroll)"}
	rowInspJump  = helpKeyRow{"g/G", []string{"g", "G"}, "jump to first / last event"}
	rowInspScrl  = helpKeyRow{"J/K", []string{"J", "K"}, "scroll the always-on detail pane"}
	rowInspEnter = helpKeyRow{"⏎", []string{"enter"}, "jump: center the map camera on the selected event's subject (no location: shows a hint instead)"}
	rowResume    = helpKeyRow{"space", []string{" "}, "resume — exits inspect mode"}
)

// --- villagers roster rows (handleVillagersKey, tui.go:836-898, roster state) ---
var (
	rowVillSel  = helpKeyRow{"j/k", []string{"j", "k"}, "select next / previous villager (clamped)"}
	rowVillJump = helpKeyRow{"g/G", []string{"g", "G"}, "jump to first / last villager"}
	rowVillOpen = helpKeyRow{"⏎", []string{"enter"}, "open detail for the selected villager"}
)

// --- villagers detail rows (same handler, villDetail/villDecisions states) ---
var (
	rowVillDetD     = helpKeyRow{"d", []string{"d"}, "toggle the decisions sub-view"}
	rowVillDetEsc   = helpKeyRow{"esc", []string{"esc"}, "back — closes the decisions sub-view if open, else closes detail"}
	rowVillDetJK    = helpKeyRow{"j/k", []string{"j", "k"}, "scroll the decisions sub-view (only while it's open; no effect otherwise)"}
	rowVillDetGG    = helpKeyRow{"g/G", []string{"g", "G"}, "no effect here (roster-only)"}
	rowVillDetEnter = helpKeyRow{"⏎", []string{"enter"}, "no effect here (roster-only)"}
)

// helpPages is the whole overlay's key content (T004) — indexed by
// helpModeKey, each page partitioning its mode's real key set between the
// two tiers (FR-002/FR-003; help_test.go's keymap sweep is the mechanical
// gate, SC-003).
var helpPages = [helpModeCount]helpModePage{
	helpModeGlobal: {
		Title: "Global (home)",
		Basic: []helpKeyRow{rowTab2, rowTab3, rowTab4, rowTab5, rowConsole, rowAsk, rowPause, rowQuit},
		Advanced: []helpKeyRow{
			rowTab6, rowHome, rowDockCycle, rowNarrowMB, rowExitSolo,
			rowChronA, rowChronT, rowChronR, rowPan, rowRecenter, rowSpeed,
		},
	},
	helpModeMinibuffer: {
		Title:    "Minibuffer (focused)",
		Basic:    []helpKeyRow{rowMbEsc, rowMbSend, rowMbHist},
		Advanced: []helpKeyRow{rowMbType, rowMbBS},
	},
	helpModeInspect: {
		Title: "Inspect (paused, chronicle)",
		Basic: []helpKeyRow{rowInspSel, rowInspScrl, rowResume, rowAsk},
		Advanced: []helpKeyRow{
			rowInspJump, rowInspEnter,
			rowHome, rowTab2, rowTab3, rowTab4, rowTab5, rowTab6, rowDockCycle, rowSpeed,
			rowPan, rowRecenter, rowChronA, rowChronT, rowChronR, rowExitSolo, rowQuit,
		},
	},
	helpModeVillagersRoster: {
		Title: "Villagers (roster)",
		Basic: []helpKeyRow{rowVillSel, rowVillOpen, rowPause, rowQuit},
		Advanced: []helpKeyRow{
			rowVillJump, rowExitSolo,
			rowHome, rowTab2, rowTab3, rowTab4, rowTab5, rowTab6, rowDockCycle, rowAsk,
			rowSpeed, rowPan, rowRecenter, rowChronA, rowChronT, rowChronR,
		},
	},
	helpModeVillagersDetail: {
		Title: "Villagers (detail)",
		Basic: []helpKeyRow{rowVillDetD, rowVillDetEsc, rowVillDetJK, rowPause, rowQuit},
		Advanced: []helpKeyRow{
			rowVillDetGG, rowVillDetEnter,
			rowHome, rowTab2, rowTab3, rowTab4, rowTab5, rowTab6, rowDockCycle, rowAsk,
			rowSpeed, rowPan, rowRecenter, rowChronA, rowChronT, rowChronR,
		},
	},
	helpModeSolo: {
		Title: "Solo zoom / narrow fallback",
		Basic: []helpKeyRow{rowHome, rowTab2, rowTab3, rowTab4, rowTab5, rowConsole, rowPause, rowQuit},
		Advanced: []helpKeyRow{
			rowTab6, rowDockCycle, rowAsk, rowSpeed, rowPan, rowRecenter,
			rowChronA, rowChronT, rowChronR, rowExitSolo, rowNarrowMB,
		},
	},
}

// --- T005/T006: overlay sections + rendering ---

// helpSection is which of the overlay's sections is on screen (contracts/
// help-content.md "Sections & tiers"; spec 056 adds helpSectionCeremonies —
// overlays/help.md's "ceremony replay entries" row, classified
// status-derived like the lessons registry).
type helpSection int

const (
	helpSectionKeys helpSection = iota
	helpSectionWalkthrough
	helpSectionLessons
	helpSectionCeremonies
	helpSectionCount
)

var helpSectionTitle = [helpSectionCount]string{
	helpSectionKeys:        "keys",
	helpSectionWalkthrough: "the screen",
	helpSectionLessons:     "lessons",
	helpSectionCeremonies:  "ceremonies",
}

// helpPanelView is the overlay's body-replacement panel (R2): the solo-zoom
// precedent — this client has no z-compositing, so help renders *instead
// of* the map/dock body (widescreenView) or the active pane (narrowView),
// chrome (header/minibuffer/footer) staying visible. Sized with the same
// styleBox/clipContent discipline every other panel uses (mapPanelView,
// soloPanelView) so the exact-height invariant holds with the overlay open.
func (m Model) helpPanelView(cols, rows int) string {
	if rows < 5 { // B5: never let a starved resize drive Height() negative
		rows = 5
	}
	inner := cols - 4
	if inner < 10 {
		inner = 10
	}
	title := styleHeader.Render("HELP · " + helpSectionTitle[m.helpSection])
	contentRows := rows - 5 // title(1) + blank(1) + footer(1) inside the box's own -2 budget
	if contentRows < 1 {
		contentRows = 1
	}
	content := strings.Join(m.helpContentLines(inner, contentRows), "\n")
	footer := styleDim.Render(m.helpFooterHint())
	body := title + "\n" + content + "\n" + footer
	// clipContent: see mapPanelView's doc comment — never let a too-wide
	// line soft-wrap and grow the panel past its Height() budget.
	return styleBox.Width(cols - 2).Height(rows - 2).Render(clipContent(body, cols-2))
}

// helpContentLines renders the current section's content, windowed to
// exactly maxLines by the shared pager (R4/FR-009).
func (m Model) helpContentLines(width, maxLines int) []string {
	var raw []string
	switch m.helpSection {
	case helpSectionWalkthrough:
		raw = m.helpWalkthroughLines(width)
	case helpSectionLessons:
		raw = helpLessonsLines(width)
	case helpSectionCeremonies:
		raw = m.ceremonyReplayLines(width)
	default:
		raw = m.helpKeysLines(width)
	}
	return paginateHelpContent(raw, m.helpScroll, maxLines)
}

// helpKeysLines renders the frozen-open mode-keys section: the page
// currently being paged through (helpPageMode, which starts equal to
// helpMode and can be cycled — FR-001) at its current tier.
func (m Model) helpKeysLines(width int) []string {
	page := helpPages[m.helpPageMode]
	rows, tierLabel := page.Basic, "basic"
	if m.helpTier {
		rows, tierLabel = page.Advanced, "advanced"
	}
	lines := []string{
		styleHeader.Render(page.Title) + " — " + tierLabel + styleDim.Render(" (t: tier · n/p: mode)"),
		"",
	}
	for _, r := range rows {
		lines = append(lines, clipLine(fmt.Sprintf("%-16s %s", r.Label, r.Action), width))
	}
	lines = append(lines, "", styleDim.Render("ctrl+c — quit, from anywhere"))
	return lines
}

// helpWalkthroughLines is US2's screen walkthrough: header anatomy, the map
// glyph legend (rendered long-form from the same mapGlyphs table
// renderMapGrid's compact legend uses — FR-005), and the dock tabs (skin-
// resolved, spec 052).
func (m Model) helpWalkthroughLines(width int) []string {
	lines := []string{styleHeader.Render("Header anatomy")}
	for _, r := range headerAnatomy {
		lines = append(lines, clipLine(fmt.Sprintf("%-28s %s", r.Element, r.Meaning), width))
	}
	lines = append(lines, "", styleHeader.Render("Map glyphs"))
	for _, g := range mapGlyphs {
		lines = append(lines, clipLine(fmt.Sprintf("%-4s %s", g.Glyph, g.Meaning), width))
	}
	lines = append(lines,
		clipLine(fmt.Sprintf("%-4s %s", "A/a", "a villager's initial — uppercase awake, lowercase asleep ("+agentGlyphNote+")"), width),
		clipLine(fmt.Sprintf("%-4s %s", "†", "a dead villager"), width),
		clipLine(fmt.Sprintf("%-4s %s", "", mapControlNote), width),
	)
	lines = append(lines, "", styleHeader.Render("Dock tabs"))
	for _, d := range m.dockTabEntries() {
		lines = append(lines, clipLine(fmt.Sprintf("%s %-10s %s", d.Key, d.Name, d.Purpose), width))
	}
	return lines
}

// helpLessonsLines is US4's pull-reference section (SC-006): the seam
// renders each helpLesson entry, or a placeholder line while the table
// ships empty.
func helpLessonsLines(width int) []string {
	if len(helpLessons) == 0 {
		return []string{styleDim.Render("lessons appear here as the village teaches them")}
	}
	var lines []string
	for i, l := range helpLessons {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, styleHeader.Render(l.Title))
		lines = append(lines, wrapText(l.Body, width)...)
	}
	return lines
}

// --- spec 056: the ceremony-replay section (overlays/ceremony.md
// "Replayability"; overlays/help.md "ceremony replay entries", classified
// status-derived) ---

// ceremonyReplayLines renders the `?` overlay's ceremony-replay section
// (spec 056 FR-004/FR-013): every stage this world has ever unlocked
// (replica.StagesUnlocked — the durable per-world facts, never a second
// event scan), each re-rendering the SAME narrated chapter + report card
// the live ceremony showed (research R5: "stored, never regenerated";
// SC-004: retrievable with zero model calls) — sharing ceremonyView's own
// helpers (CeremonyChapter, ceremonyReportCardFor) so the two can never
// show different content for the same stage. Degrades to an honest
// placeholder with nil/empty replica (TestHelpContentReadsNoStatusOrReplica
// requires non-empty content even then).
func (m Model) ceremonyReplayLines(width int) []string {
	if m.replica == nil || len(m.replica.StagesUnlocked) == 0 {
		return []string{styleDim.Render("no stage has unlocked yet in this world")}
	}
	var lines []string
	for i, stage := range m.replica.StagesUnlocked {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, styleHeader.Render(strings.ToUpper(m.sk().StageName(stage))+" — unlocked"))
		lines = append(lines, wrapText(m.sk().CeremonyChapter(stage), width)...)
		if card := m.ceremonyReportCardFor(stage, width); card != "" {
			lines = append(lines, "", card)
		}
	}
	return lines
}

// helpFooterHint is the overlay's own footer, replacing the mode-beneath's
// hint while open (FR-011/FR-012): dismissal and section-cycling always
// apply; tier/mode-paging only mean something in the keys section.
func (m Model) helpFooterHint() string {
	if m.helpSection == helpSectionKeys {
		return "esc/? close · tab section · t tier · n/p mode · J/K scroll"
	}
	return "esc/? close · tab section · J/K scroll"
}

// paginateHelpContent windows content to exactly maxLines lines — the
// chronicleDetailPane pager idiom (views.go) copied for the help overlay
// (R4): scroll clamps to content so J past the end (or K before the start)
// is a no-op; an overflow footer replaces the last row with a remaining-
// line count once content doesn't fit, the same "shed content, never
// overflow" discipline every other pane in this package uses.
func paginateHelpContent(content []string, scroll, maxLines int) []string {
	if maxLines < 1 {
		maxLines = 1
	}
	contentRows := maxLines
	footerNeeded := len(content) > maxLines
	if footerNeeded {
		contentRows = maxLines - 1
		if contentRows < 1 {
			contentRows = 1
		}
	}
	maxScroll := len(content) - contentRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}
	if scroll < 0 {
		scroll = 0
	}
	visEnd := scroll + contentRows
	if visEnd > len(content) {
		visEnd = len(content)
	}
	out := append([]string{}, content[scroll:visEnd]...)
	if footerNeeded {
		remaining := len(content) - visEnd
		out = append(out, styleDim.Render(fmt.Sprintf("… (+%d more — J to scroll)", remaining)))
	}
	for len(out) < maxLines {
		out = append(out, "")
	}
	return out
}
