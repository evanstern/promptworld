package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/metatron"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// TestWidescreenViewExactHeight is the B1 regression: the widescreen
// composite must render to EXACTLY m.height lines in every mode layered on
// top of it. Bubble Tea scrolls the top of the frame off-screen when a
// View() overflows the terminal height, which is what pushed the header
// row off-screen in the live tmux repro. Covers home, solo, and
// paused/inspect (both in the dock and solo'd), at several sizes straddling
// the breakpoint and the 50/50 column split.
func TestWidescreenViewExactHeight(t *testing.T) {
	sizes := []struct{ w, h int }{
		{112, 20}, {112, 30}, {113, 30}, {118, 30}, {140, 40}, {160, 50}, {200, 24},
	}
	for _, sz := range sizes {
		for _, state := range []string{
			"home", "solo", "inspect", "inspect-solo", "villagers-solo", "villagers-detail-solo", "metatron-solo",
			// spec 045 (TASK-116): the help overlay is a body-replacement
			// panel like solo zoom (R2) — it must hold the same exact-
			// height invariant at every size, in every section/tier.
			"help", "help-advanced", "help-walkthrough", "help-lessons",
		} {
			t.Run(fmt.Sprintf("%dx%d/%s", sz.w, sz.h, state), func(t *testing.T) {
				m := widescreenModel(t)
				m.width, m.height = sz.w, sz.h
				seedEvents(&m, 20)
				switch state {
				case "solo":
					m.solo = true
				case "inspect":
					m.connected = true
					m.status = &ipc.StatusData{Clock: ipc.ClockStatus{Paused: true}}
					m.chronSelected = 5
				case "inspect-solo":
					m.connected = true
					m.status = &ipc.StatusData{Clock: ipc.ClockStatus{Paused: true}}
					m.chronSelected = 5
					m.solo = true
				case "villagers-solo":
					m.dockTab = paneVillagers
					m.solo = true
				case "villagers-detail-solo":
					m.dockTab = paneVillagers
					m.solo = true
					m.villDetail = true
					m.replica.Agents[0].Beliefs = []sim.Belief{{Statement: "the fire needs tending", Confidence: 80}}
					m.replica.Agents[0].Narrative = "a long night watching the fire die and reviving it by hand."
					for i := 0; i < 20; i++ {
						m.replica.Agents[0].Memories = append(m.replica.Agents[0].Memories,
							sim.Memory{Text: "chopped wood at the treeline", Salience: 3, Tick: int64(i) * 60})
					}
				case "metatron-solo":
					m.dockTab = paneMetatron
					m.solo = true
					m.transcript = []string{"you: why is Rowan hoarding wood?", "angel: three cold nights, and Ash let the fire die each time."}
				case "help":
					m.helpOpen = true
					m.helpPageMode = helpModeGlobal
					m.helpSection = helpSectionKeys
				case "help-advanced":
					m.helpOpen = true
					m.helpPageMode = helpModeGlobal
					m.helpSection = helpSectionKeys
					m.helpTier = true
				case "help-walkthrough":
					m.helpOpen = true
					m.helpSection = helpSectionWalkthrough
				case "help-lessons":
					m.helpOpen = true
					m.helpSection = helpSectionLessons
				}
				v := m.View()
				lines := strings.Split(v, "\n")
				if len(lines) != sz.h {
					t.Errorf("View() = %d lines, want exactly %d — a taller frame scrolls the header off the top of a real terminal:\n%s",
						len(lines), sz.h, v)
				}
			})
		}
	}
}

// TestWidescreenViewExactHeightWithLongMinibufferInput is B3's regression:
// a long focused-minibuffer input must never wrap the box past its fixed
// 3-row budget.
func TestWidescreenViewExactHeightWithLongMinibufferInput(t *testing.T) {
	m := widescreenModel(t)
	m.width, m.height = 140, 40
	m.mbFocused = true
	m.mbInput = strings.Repeat("why does Rowan keep hoarding wood when the fire needs tending ", 5)
	v := m.View()
	lines := strings.Split(v, "\n")
	if len(lines) != m.height {
		t.Fatalf("View() with a long focused input = %d lines, want exactly %d:\n%s", len(lines), m.height, v)
	}
}

// TestWidescreenViewExactHeightDenseChronicle is B1+B2+B5 together (spec
// 018: the always-on detail pane replaces the old ⏎-triggered expansion,
// but the row-budget invariant it guarded is the same one): a chronicle
// with far more events than fit (600 in a 30-row terminal), paused with a
// mid-ring selection — the composite must still render to exactly
// m.height lines, and the map panel (unrelated to how busy the dock is)
// must stay at its budgeted height.
func TestWidescreenViewExactHeightDenseChronicle(t *testing.T) {
	m := widescreenModel(t)
	m.width, m.height = 112, 30
	seedEvents(&m, 600)
	m.connected = true
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{Paused: true}}
	m.chronSelected = 300

	v := m.View()
	lines := strings.Split(v, "\n")
	if len(lines) != m.height {
		t.Fatalf("View() with 600 events, paused mid-ring = %d lines, want exactly %d:\n%s", len(lines), m.height, v)
	}

	cols := computeColumns(m.width)
	rows := computeRows(m.height)
	mapPanel := m.mapPanelView(cols.MapCols, rows.Body)
	if got := len(strings.Split(mapPanel, "\n")); got != rows.Body {
		t.Errorf("map panel = %d lines while the dock is dense+expanded, want exactly its budgeted %d — "+
			"the dock's content must never bleed into the map's row budget", got, rows.Body)
	}
}

// --- style-role tests (TASK-60 Phase 5, T022) ---
//
// go test's stdout isn't a TTY, so lipgloss's default renderer auto-detects
// no color and Render() becomes a no-op — these tests force a color
// profile for their duration (lipgloss.SetColorProfile is documented as
// existing "mostly for testing purposes") and restore it afterward so
// other tests in the package are unaffected.

func withColorProfile(t *testing.T, p termenv.Profile) {
	t.Helper()
	orig := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(p)
	t.Cleanup(func() { lipgloss.SetColorProfile(orig) })
}

// TestFamilyTintDistinctPerFamily: every family color role (contract §4)
// renders to a distinguishable style — the "roles, never raw colors"
// requirement means each named token must actually differ, not just exist.
func TestFamilyTintDistinctPerFamily(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	families := []eventFamily{
		familyWorld, familyClock, familySim, familyAgent, familySocial,
		familyGovernance, familyGru, familyChronicle, familyMetatron, familyCog,
	}
	seen := map[string]eventFamily{}
	for _, f := range families {
		rendered := familyTint(f).Render("TYPE")
		if !strings.Contains(rendered, "\x1b") {
			t.Errorf("family %v: tint produced no ANSI under a forced color profile: %q", f, rendered)
		}
		if prior, ok := seen[rendered]; ok {
			t.Errorf("family %v renders identically to family %v (%q) — tints must be distinguishable", f, prior, rendered)
		}
		seen[rendered] = f
	}
}

// TestPaintStyledLineRoleMapping: paintStyledLine maps each seg role to its
// documented style token (contract §4: name, speech, emphasis) and the
// prefix to the family tint passed in — role→token mapping, not just "some
// style got applied".
func TestPaintStyledLineRoleMapping(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	summary := []seg{
		{Text: "Ash", Role: segName},
		{Text: " said ", Role: segText},
		{Text: `"hi"`, Role: segSpeech},
		{Text: " x2", Role: segEmphasis},
	}
	lines := styleWrapLine("TYPE  ", summary, 60, 1)
	if len(lines) != 1 {
		t.Fatalf("want 1 styled line, got %d", len(lines))
	}
	out := paintStyledLine(lines[0], styleFamilyAgent, false)

	for _, want := range []string{
		styleFamilyAgent.Render("TYPE  "),
		styleFeedName.Render("Ash"),
		styleFeedSpeech.Render(`"hi"`),
		styleFeedEmphasis.Render(" x2"),
	} {
		if !strings.Contains(out, want) {
			t.Errorf("styled output missing expected fragment %q in:\n%q", want, out)
		}
	}
}

// TestRenderChronicleRowAlertWholeLine: the four alert types (contract §2)
// render the entire line in styleFeedAlert regardless of family.
func TestRenderChronicleRowAlertWholeLine(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	e := store.Event{Seq: 1, Tick: 60, Type: "agent.died", Payload: json.RawMessage(`{"agent":0,"cause":"starvation"}`)}
	l := formatChronicleLine(e, []string{"Ash"})
	cols := computeChronicleColumns([]chronicleLine{l}, false)
	got := renderChronicleRow(l, cols, 60, 1, false)
	want := styleFeedAlert.Render(plainChronicleLine(l, cols))
	if got != want {
		t.Errorf("alert row should be styled whole-line with styleFeedAlert:\ngot:  %q\nwant: %q", got, want)
	}
}

// TestRenderChronicleRowLabeledVoiceWholeLine: labeled-voice families
// (cog/clock/daemon, contract §2) tint the whole line, not per-segment.
func TestRenderChronicleRowLabeledVoiceWholeLine(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	e := store.Event{Seq: 1, Tick: 60, Type: "clock.speed_set", Payload: json.RawMessage(`{"speed":"4x"}`)}
	l := formatChronicleLine(e, nil)
	cols := computeChronicleColumns([]chronicleLine{l}, false)
	got := renderChronicleRow(l, cols, 60, 1, false)
	want := styleFeedClock.Render(plainChronicleLine(l, cols))
	if got != want {
		t.Errorf("labeled-voice row should be styled whole-line with the family tint:\ngot:  %q\nwant: %q", got, want)
	}
}

// TestRenderChronicleRowSelectionReverse: selection reverse survives the
// segment-wise styling rework (R4/T021) for both alert and phrase-voice rows.
func TestRenderChronicleRowSelectionReverse(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	cases := []struct {
		typ     string
		payload string
	}{
		{"agent.died", `{"agent":0,"cause":"starvation"}`}, // alert path
		{"clock.speed_set", `{"speed":"4x"}`},              // labeled-voice path
		{"agent.moved", `{"agent":0,"x":1,"y":1}`},         // phrase-voice, segment-wise path
	}
	for _, c := range cases {
		e := store.Event{Seq: 1, Tick: 60, Type: c.typ, Payload: json.RawMessage(c.payload)}
		l := formatChronicleLine(e, []string{"Ash"})
		cols := computeChronicleColumns([]chronicleLine{l}, false)
		unselected := renderChronicleRow(l, cols, 60, 1, false)
		selected := renderChronicleRow(l, cols, 60, 1, true)
		if selected == unselected {
			t.Errorf("%s: selected row should render differently (reverse video) than unselected", c.typ)
		}
	}
}

// TestPureLayerEmitsNoANSI: the pure formatting layer (grammar.go/digest.go)
// never touches lipgloss (R4) — sweeps the full catalog fixture plus the
// fallback path for any stray ESC byte, regardless of color profile (this
// invariant must hold even when no profile is forced).
func TestPureLayerEmitsNoANSI(t *testing.T) {
	names := []string{"Ash", "Birch", "Cedar", "Rowan"}
	for typ, fx := range catalogFixture {
		e := store.Event{Seq: 1, Tick: 1, Type: typ, Payload: json.RawMessage(fx.payload)}
		l := formatChronicleLine(e, names)
		if strings.Contains(plainSegs(l.Summary), "\x1b") {
			t.Errorf("%s: digest summary contains an ESC byte — the pure layer must never emit ANSI", typ)
		}
		cols := computeChronicleColumns([]chronicleLine{l}, false)
		if strings.Contains(plainChronicleLine(l, cols), "\x1b") {
			t.Errorf("%s: plainChronicleLine contains an ESC byte", typ)
		}
	}
	fallback := store.Event{Seq: 1, Tick: 1, Type: "future.unknown_type", Payload: json.RawMessage(`{"agent":0}`)}
	l := formatChronicleLine(fallback, names)
	if strings.Contains(plainSegs(l.Summary), "\x1b") {
		t.Error("fallback summary contains an ESC byte")
	}
}

// --- llm provider table (spec 024 US6, contracts/status.md "TUI") ---

// TestLLMProviderLinesFields: every declared field lands in the row — name,
// model, up/down, queue, inflight/slots, contended, spend — for both an up
// and a down provider.
func TestLLMProviderLinesFields(t *testing.T) {
	st := &llm.Status{
		Providers: []llm.ProviderStatus{
			{Name: "cogito", Model: "cogito:3b", Up: true, Queue: 3, Inflight: 4, Slots: 4, Contended: false, SpentUSD: 0.42},
			{Name: "anthropic", Model: "claude-opus-4-8", Up: false, Queue: 0, Inflight: 0, Slots: 1, Contended: false, SpentUSD: 12.41},
		},
		Month: "2026-07", Spent: 12.83, Budget: 100,
	}
	lines := llmProviderLines(st)
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2 (no unattributed remainder — rows sum to Spent): %v", len(lines), lines)
	}
	up, down := lines[0], lines[1]
	for _, want := range []string{"cogito", "cogito:3b", "q3", "4/4", "$0.42"} {
		if !strings.Contains(up, want) {
			t.Errorf("up-provider row missing %q: %q", want, up)
		}
	}
	if strings.Contains(up, "●") == false {
		t.Errorf("an up provider should carry the up glyph: %q", up)
	}
	for _, want := range []string{"anthropic", "claude-opus-4-8", "q0", "0/1", "$12.41"} {
		if !strings.Contains(down, want) {
			t.Errorf("down-provider row missing %q: %q", want, down)
		}
	}
	if strings.Contains(down, "○") == false {
		t.Errorf("a down provider should carry the down glyph: %q", down)
	}
}

// TestLLMProviderLinesContendedMarker: the lease-wait flag renders distinctly
// from an uncontended row.
func TestLLMProviderLinesContendedMarker(t *testing.T) {
	st := &llm.Status{Providers: []llm.ProviderStatus{
		{Name: "local", Model: "gemma4:12b-mlx", Up: true, Slots: 4, Contended: true},
	}, Spent: 0, Budget: 100}
	lines := llmProviderLines(st)
	if len(lines) != 1 || !strings.Contains(lines[0], "⏳") {
		t.Errorf("contended provider should carry the contended marker: %v", lines)
	}
}

// TestLLMProviderLinesUnattributedRemainder (contracts/status.md: Σ rows ≤
// global spent_usd; the difference is legacy unattributed spend): a nonzero
// remainder gets its own trailing row; an exactly-accounted total does not.
func TestLLMProviderLinesUnattributedRemainder(t *testing.T) {
	st := &llm.Status{
		Providers: []llm.ProviderStatus{{Name: "cloud", Model: "claude-opus-4-8", Up: true, Slots: 1, SpentUSD: 5}},
		Spent:     8, Budget: 100,
	}
	lines := llmProviderLines(st)
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2 (provider row + unattributed remainder): %v", len(lines), lines)
	}
	if !strings.Contains(lines[1], "(unattributed)") || !strings.Contains(lines[1], "$3.00") {
		t.Errorf("remainder row = %q, want (unattributed) and $3.00", lines[1])
	}

	exact := &llm.Status{
		Providers: []llm.ProviderStatus{{Name: "cloud", Model: "claude-opus-4-8", Up: true, Slots: 1, SpentUSD: 5}},
		Spent:     5, Budget: 100,
	}
	if lines := llmProviderLines(exact); len(lines) != 1 {
		t.Errorf("fully attributed spend should not add a remainder row: %v", lines)
	}
}

// TestLLMProviderLinesEmpty: nil status and an empty registry both render no
// rows (rather than panicking or printing a bare remainder).
func TestLLMProviderLinesEmpty(t *testing.T) {
	if lines := llmProviderLines(nil); lines != nil {
		t.Errorf("nil status: %v", lines)
	}
	if lines := llmProviderLines(&llm.Status{}); lines != nil {
		t.Errorf("empty registry: %v", lines)
	}
}

// TestLLMProviderLinesConditionAnnotation (spec 034 T010, contracts/
// provider-conditions.md "Human surfaces"): a provider carrying an active
// condition gains an indented continuation line with the condition's detail
// and remedy immediately below its row; a healthy provider carries none.
func TestLLMProviderLinesConditionAnnotation(t *testing.T) {
	st := &llm.Status{Providers: []llm.ProviderStatus{
		{
			Name: "local", Model: "cogito:3b", Endpoint: "http://localhost:11434/v1", Up: false,
			Condition:       "model-missing",
			ConditionDetail: `model "cogito:3b" not served by http://localhost:11434/v1`,
			ConditionRemedy: "ollama pull cogito:3b",
		},
		{Name: "cloud", Model: "claude-opus-4-8", Up: true},
	}}
	lines := llmProviderLines(st)
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3 (affected row + its continuation + healthy row): %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "local") {
		t.Fatalf("lines[0] should be the local provider row: %q", lines[0])
	}
	for _, want := range []string{`model "cogito:3b" not served by http://localhost:11434/v1`, "ollama pull cogito:3b"} {
		if !strings.Contains(lines[1], want) {
			t.Errorf("condition continuation line missing %q: %q", want, lines[1])
		}
	}
	if !strings.Contains(lines[2], "cloud") {
		t.Fatalf("lines[2] should be the healthy cloud provider row: %q", lines[2])
	}
	if strings.Contains(lines[2], "ollama") || strings.Contains(lines[2], "not served") {
		t.Errorf("healthy provider row should carry no condition text: %q", lines[2])
	}
}

// --- T005: metatron-pane cognition-horizon block (spec 037 US1, FR-006/FR-007) ---

// TestHorizonLinesMixedStandings: a header plus one row per entry — thinking
// classes read "thinking at <speed>"; suppressed classes read "suppressed at
// <speed>" with the remedy split by calibration (uncalibrated → "calibrate or
// slow down"; calibrated → "slow down") and the router's verdict arithmetic as
// trailing verbatim detail.
func TestHorizonLinesMixedStandings(t *testing.T) {
	horizon := []ipc.HorizonClass{
		{Class: "planner", Suppressed: true, Calibrated: false, Verdict: "3pt x 20.0s/pt x 32x = 1920 ticks > budget 1200"},
		{Class: "conversation", Suppressed: true, Calibrated: true, Verdict: "13pt x 0.9s/pt x 32x = 374 ticks > budget 300"},
		{Class: "meeting", Suppressed: false, Calibrated: true},
	}
	lines := horizonLines(horizon, "32x")
	if len(lines) != 4 {
		t.Fatalf("got %d lines, want 4 (header + 3 rows): %v", len(lines), lines)
	}
	if !strings.Contains(lines[1], "planner suppressed at 32x") || !strings.Contains(lines[1], "calibrate or slow down") {
		t.Errorf("uncalibrated suppressed row wrong: %q", lines[1])
	}
	if !strings.Contains(lines[1], "3pt x 20.0s/pt x 32x = 1920 ticks > budget 1200") {
		t.Errorf("suppressed row must carry the verbatim verdict detail: %q", lines[1])
	}
	if !strings.Contains(lines[2], "conversation suppressed at 32x") || !strings.Contains(lines[2], "slow down") {
		t.Errorf("calibrated suppressed row wrong: %q", lines[2])
	}
	if strings.Contains(lines[2], "calibrate") {
		t.Errorf("a calibrated class must never be told to calibrate: %q", lines[2])
	}
	if !strings.Contains(lines[3], "meeting thinking at 32x") {
		t.Errorf("thinking row wrong: %q", lines[3])
	}
}

// TestHorizonLinesEmpty: a world with no horizon (no-LLM) renders no block —
// the metatron pane shows nothing extra.
func TestHorizonLinesEmpty(t *testing.T) {
	if lines := horizonLines(nil, "8x"); lines != nil {
		t.Errorf("no horizon should render no lines, got %v", lines)
	}
}

// TestHorizonLinesSkippedCount (spec 037 US2 / T009): the "skipped N" count
// rides every suppressed row, and every thinking row with a nonzero
// accumulated count; a thinking row with a zero count carries no count clutter.
func TestHorizonLinesSkippedCount(t *testing.T) {
	horizon := []ipc.HorizonClass{
		{Class: "planner", Suppressed: true, SuppressedCount: 214, Verdict: "3pt ..."},
		{Class: "conversation", Suppressed: false, SuppressedCount: 12}, // slowed down; count retained
		{Class: "meeting", Suppressed: false, SuppressedCount: 0},       // never suppressed
	}
	lines := horizonLines(horizon, "8x")
	if !strings.Contains(lines[1], "skipped 214") {
		t.Errorf("suppressed row missing count: %q", lines[1])
	}
	if !strings.Contains(lines[2], "conversation thinking at 8x") || !strings.Contains(lines[2], "skipped 12") {
		t.Errorf("thinking row with accumulated count wrong: %q", lines[2])
	}
	if strings.Contains(lines[3], "skipped") {
		t.Errorf("thinking row with zero count should omit the count: %q", lines[3])
	}
}

// TestOrderStatusLinesFields (spec 029 T023): the metatron pane's
// standing-orders block carries id, fuzzy marker, origin, expiry day, status,
// and condition for every order, headed by a count line.
func TestOrderStatusLinesFields(t *testing.T) {
	orders := []metatron.OrderStatus{
		{ID: "ord-120-1", Condition: "Rowan falls asleep", Origin: "player", ExpiresDay: 6, Status: "active"},
		{ID: "ord-130-1", Condition: "Rowan seems heartbroken", Origin: "system", Fuzzy: true, ExpiresDay: 7, Status: "active"},
	}
	lines := orderStatusLines(orders)
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3 (header + 2 orders): %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "2") {
		t.Errorf("header should carry the order count: %q", lines[0])
	}
	for _, want := range []string{"ord-120-1", "player", "day 6", "active", "Rowan falls asleep"} {
		if !strings.Contains(lines[1], want) {
			t.Errorf("structural order row missing %q: %q", want, lines[1])
		}
	}
	if strings.Contains(lines[1], "~") {
		t.Errorf("a structural (non-fuzzy) order should carry no fuzzy marker: %q", lines[1])
	}
	for _, want := range []string{"ord-130-1", "~", "system", "day 7", "Rowan seems heartbroken"} {
		if !strings.Contains(lines[2], want) {
			t.Errorf("fuzzy order row missing %q: %q", want, lines[2])
		}
	}
}

// TestOrderStatusLinesEmpty: no standing orders renders no rows.
func TestOrderStatusLinesEmpty(t *testing.T) {
	if lines := orderStatusLines(nil); lines != nil {
		t.Errorf("no orders: %v", lines)
	}
}

// TestMetatronViewRendersProviderTable: the console pane surfaces the
// per-provider table it's handed, not the old plain up/down line.
func TestMetatronViewRendersProviderTable(t *testing.T) {
	m := testModel(t)
	m.status = &ipc.StatusData{LLM: &llm.Status{
		Providers: []llm.ProviderStatus{
			{Name: "cogito", Model: "cogito:3b", Up: true, Queue: 1, Inflight: 1, Slots: 4, SpentUSD: 0},
		},
		Spent: 0, Budget: 100,
	}}
	view := m.metatronView()
	for _, want := range []string{"cogito", "cogito:3b", "q1", "1/4"} {
		if !strings.Contains(view, want) {
			t.Errorf("metatron view missing provider field %q:\n%s", want, view)
		}
	}
}

// --- guardian strip (panels/guardian-strip.md, spec 050) ---

// TestGuardianStripViewNoStatus: pre-status (connecting) renders a blank —
// but present — row: no invented zeros (US2 AS-3, contract §2).
func TestGuardianStripViewNoStatus(t *testing.T) {
	m := testModel(t)
	if got := m.guardianStripView(200); got != "" {
		t.Errorf("no status snapshot: guardianStripView = %q, want blank", got)
	}
}

// TestGuardianStripViewPresenceMatrix sweeps data-model.md's presence
// matrix (SC-002): every rendered segment must be a true statement, every
// absent segment genuinely absent — no dash, no invented zero.
func TestGuardianStripViewPresenceMatrix(t *testing.T) {
	cases := []struct {
		name       string
		charges    int
		orders     int
		wantRegen  bool
		wantOrders string
	}{
		{"partial bank, N orders", 1, 2, true, "👁 2 standing orders"},
		{"full bank omits regen", sim.MetatronChargeCap, 3, false, "👁 3 standing orders"},
		{"zero orders is a true zero", 1, 0, true, "👁 0 standing orders"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := testModel(t)
			m.status = &ipc.StatusData{Clock: ipc.ClockStatus{MetatronCharges: c.charges, Tick: 100}}
			m.replica.MetatronOrders = make([]sim.MetatronOrder, c.orders)
			for i := range m.replica.MetatronOrders {
				m.replica.MetatronOrders[i] = sim.MetatronOrder{ID: fmt.Sprintf("o%d", i), Status: "active"}
			}
			got := m.guardianStripView(200)
			if strings.Contains(got, "faith") {
				t.Errorf("no faith segment must ever appear (TASK-118 unshipped): %q", got)
			}
			if hasRegen := strings.Contains(got, "next +1"); hasRegen != c.wantRegen {
				t.Errorf("regen segment presence = %v, want %v: %q", hasRegen, c.wantRegen, got)
			}
			if !strings.Contains(got, c.wantOrders) {
				t.Errorf("missing standing-order segment %q: %q", c.wantOrders, got)
			}
			bank := fmt.Sprintf("(%d/%d)", c.charges, sim.MetatronChargeCap)
			if !strings.Contains(got, bank) {
				t.Errorf("missing numeric bank form %q: %q", bank, got)
			}
		})
	}
}

// TestGuardianStripViewLiveUpdate (US1 AS-2): a charge spend or a standing
// order add is reflected on the very next frame — guardianStripView reads
// m.status/m.replica directly, the same live fields the guardian tab header
// already uses, no polling required.
func TestGuardianStripViewLiveUpdate(t *testing.T) {
	m := testModel(t)
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{MetatronCharges: 2, Tick: 100}}
	before := m.guardianStripView(200)
	if !strings.Contains(before, "(2/3)") {
		t.Fatalf("fixture setup wrong: %q", before)
	}

	// Spend a charge.
	m.status.Clock.MetatronCharges = 1
	afterSpend := m.guardianStripView(200)
	if afterSpend == before || !strings.Contains(afterSpend, "(1/3)") {
		t.Errorf("charge spend not reflected next frame: %q", afterSpend)
	}

	// Place a standing order (the same client-side replica mutation
	// applying a live metatron.order_placed event performs, tui.go
	// applyEvent/Apply).
	m.replica.MetatronOrders = append(m.replica.MetatronOrders, sim.MetatronOrder{ID: "ord-1", Status: "active"})
	afterOrder := m.guardianStripView(200)
	if !strings.Contains(afterOrder, "👁 1 standing orders") {
		t.Errorf("order add not reflected next frame: %q", afterOrder)
	}

	// An order transitioning to a terminal status is still retained
	// (pruneMetatronOrders keeps recent non-active ones alongside active
	// ones) — the count mirrors len(), exactly orderStatusLines' header
	// count, so it does not drop until the retention prune eventually runs.
	m.replica.MetatronOrders[0].Status = "expired"
	afterExpiry := m.guardianStripView(200)
	if !strings.Contains(afterExpiry, "👁 1 standing orders") {
		t.Errorf("retained-but-terminal order count should match orderStatusLines' len() semantics: %q", afterExpiry)
	}
}

// TestJoinStripSegmentsTruncatesRightToLeft is joinStripSegments as a pure
// function (plain ASCII fixtures sidestep any emoji display-width
// ambiguity): dropping segments right-to-left with an ellipsis marker, the
// first (leftmost, "bank") segment the last thing standing, hard-clipped
// rather than wrapped if even it doesn't fit (contract §1/edge cases).
func TestJoinStripSegmentsTruncatesRightToLeft(t *testing.T) {
	segs := []string{"BANK", "REGEN", "ORDERS"}
	full := "BANK · REGEN · ORDERS"
	if got := joinStripSegments(segs, 100); got != full {
		t.Fatalf("full width = %q, want %q", got, full)
	}
	// lipgloss.Width, not len(): "·" is a multi-byte rune (still 1 display
	// column) — byte length and display width diverge for it.
	if got := joinStripSegments(segs, lipgloss.Width(full)-1); got != "BANK · REGEN…" {
		t.Errorf("one column short of full should drop the rightmost (orders) segment first: %q", got)
	}
	if got := joinStripSegments(segs, lipgloss.Width("BANK · REGEN…")-1); got != "BANK…" {
		t.Errorf("dropping the middle (regen) segment next, bank last standing: %q", got)
	}
	if got := joinStripSegments(segs, 2); got != "B…" {
		t.Errorf("even the bank alone doesn't fit: hard-clip, never wrap: %q", got)
	}
	if got := joinStripSegments(segs, 0); got != "" {
		t.Errorf("width <= 0: %q", got)
	}
}

// TestGuardianStripViewNeverExceedsWidth is the composed sanity check over
// joinStripSegments's real inputs (glyphs/emoji included): at every width
// the strip stays one line and never exceeds the budget (SC-003).
func TestGuardianStripViewNeverExceedsWidth(t *testing.T) {
	m := testModel(t)
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{MetatronCharges: 1, Tick: 100}}
	m.replica.MetatronOrders = []sim.MetatronOrder{{ID: "o1", Status: "active"}, {ID: "o2", Status: "active"}}
	for _, w := range []int{1, 2, 5, 10, 20, 40, 200} {
		got := m.guardianStripView(w)
		if strings.Contains(got, "\n") {
			t.Errorf("width %d: strip must never wrap to a second row: %q", w, got)
		}
		if lipgloss.Width(got) > w {
			t.Errorf("width %d: rendered strip width %d exceeds budget: %q", w, lipgloss.Width(got), got)
		}
	}
}

// TestMinibufferDormantFoldRelocation (US3 AS-1; layout.md ruling a step 4):
// once the widescreen row budget folds the guardian strip, the dormant
// minibuffer line gains the budget prefix; unfolded, it's the plain
// placeholder untouched.
func TestMinibufferDormantFoldRelocation(t *testing.T) {
	m := widescreenModel(t)
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{MetatronCharges: 2, Tick: 100}}
	m.replica.MetatronOrders = []sim.MetatronOrder{{ID: "o1", Status: "active"}}

	// Tall enough: strip stays on, dormant line is the plain placeholder.
	m.height = 40
	if rows := computeRows(m.height); rows.Strip == 0 {
		t.Fatalf("fixture setup: expected the strip to fit at height %d", m.height)
	}
	unfolded := m.minibufferView(m.width)
	if strings.Contains(unfolded, "👁") {
		t.Errorf("unfolded dormant line must not carry the budget prefix: %q", unfolded)
	}

	// Short enough: strip folds, dormant line carries the relocated prefix.
	m.height = 14
	if rows := computeRows(m.height); rows.Strip != 0 {
		t.Fatalf("fixture setup: expected the strip to fold at height %d", m.height)
	}
	folded := m.minibufferView(m.width)
	if !strings.Contains(folded, "👁1") {
		t.Errorf("folded dormant line must carry the relocated budget prefix (bank glyphs + order count): %q", folded)
	}
}

// TestMinibufferFocusedBusyUnaffectedByFold (US3 AS-1): the fold-relocation
// prefix only ever touches the dormant branch — focused and busy render
// byte-identically whether the guardian strip is folded or not.
func TestMinibufferFocusedBusyUnaffectedByFold(t *testing.T) {
	base := widescreenModel(t)
	base.status = &ipc.StatusData{Clock: ipc.ClockStatus{MetatronCharges: 2, Tick: 100}}
	base.replica.MetatronOrders = []sim.MetatronOrder{{ID: "o1", Status: "active"}}

	tall, short := base, base
	tall.height, short.height = 40, 14
	if computeRows(tall.height).Strip == 0 || computeRows(short.height).Strip != 0 {
		t.Fatalf("fixture setup: expected tall unfolded, short folded")
	}

	tall.mbFocused, short.mbFocused = true, true
	tall.mbInput, short.mbInput = "why did rowan lie about the wood", "why did rowan lie about the wood"
	if a, b := tall.minibufferView(tall.width), short.minibufferView(short.width); a != b {
		t.Errorf("focused minibuffer differs by fold state:\ntall:  %q\nshort: %q", a, b)
	}

	tall.mbFocused, short.mbFocused = false, false
	tall.mbBusy, short.mbBusy = true, true
	if a, b := tall.minibufferView(tall.width), short.minibufferView(short.width); a != b {
		t.Errorf("busy minibuffer differs by fold state:\ntall:  %q\nshort: %q", a, b)
	}
}

// TestMetatronViewNarrowCarriesStrip (US3 AS-2; layout.md ruling b): the
// narrow fallback's only minibuffer (inside the guardian pane — narrow's
// other panes have no composer at all) carries the strip above it,
// unconditionally — narrow has no fold machinery of its own to drop it.
func TestMetatronViewNarrowCarriesStrip(t *testing.T) {
	m := testModel(t) // narrow: 80x30 (below widescreenBreakpoint)
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{MetatronCharges: 1, Tick: 100}}
	m.replica.MetatronOrders = []sim.MetatronOrder{{ID: "o1", Status: "active"}, {ID: "o2", Status: "active"}}

	view := m.metatronView()
	if !strings.Contains(view, "(1/3)") {
		t.Errorf("narrow metatron pane missing the strip's charge bank: %q", view)
	}
	if !strings.Contains(view, "👁 2 standing orders") {
		t.Errorf("narrow metatron pane missing the strip's standing-order count: %q", view)
	}
	stripIdx := strings.Index(view, "(1/3)")
	mbIdx := strings.Index(view, "⏎ m — speak with the angel…")
	if stripIdx < 0 || mbIdx < 0 || stripIdx > mbIdx {
		t.Errorf("strip must render ABOVE the minibuffer: strip@%d minibuffer@%d\n%s", stripIdx, mbIdx, view)
	}
}

// --- spec 049 T013: detail-pane actions bar render states ---

// TestChronicleDetailPaneActionsBarLocatable: a locatable event's actions
// bar names the jump binding plus the resolved subject/position (contract
// §3), even when the payload is short enough that no scroll footer is
// needed — the bar is a permanent bottom-right slot, not a byproduct of
// overflow.
func TestChronicleDetailPaneActionsBarLocatable(t *testing.T) {
	e := store.Event{Seq: 1, Tick: 1, Type: "agent.moved", Payload: json.RawMessage(`{"agent":0,"x":5,"y":6}`)}
	out := chronicleDetailPane(e, []string{"Ash"}, 0, 60, 8, "⏎ jump to Ash (5,6)")
	body := strings.Join(out, "\n")
	if !strings.Contains(body, "⏎ jump to Ash (5,6)") {
		t.Errorf("actions bar should show the jump affordance: %q", body)
	}
	if len(out) != 8 {
		t.Errorf("chronicleDetailPane must render exactly paneRows lines, got %d", len(out))
	}
}

// TestChronicleDetailPaneActionsBarUnlocatable: an unlocatable event's
// actions bar shows the honest absence (contract §3, US1 AS-3's hint —
// "one surface, not two" with US3 AS-2).
func TestChronicleDetailPaneActionsBarUnlocatable(t *testing.T) {
	e := store.Event{Seq: 1, Tick: 1, Type: "clock.paused", Payload: json.RawMessage(`{}`)}
	out := chronicleDetailPane(e, nil, 0, 60, 8, "no location for this event")
	body := strings.Join(out, "\n")
	if !strings.Contains(body, "no location for this event") {
		t.Errorf("actions bar should show the honest no-location hint: %q", body)
	}
}

// TestChronicleDetailPaneActionsBarNoScrollFooterNeeded: a payload that fits
// well within the pane still shows the actions bar — it must not depend on
// the scroll footer being present.
func TestChronicleDetailPaneActionsBarNoScrollFooterNeeded(t *testing.T) {
	e := store.Event{Seq: 1, Tick: 1, Type: "clock.resumed", Payload: json.RawMessage(`{}`)}
	out := chronicleDetailPane(e, nil, 0, 60, 10, "no location for this event")
	body := strings.Join(out, "\n")
	if strings.Contains(body, "more — J to scroll") {
		t.Errorf("a short payload should not need the scroll footer: %q", body)
	}
	if !strings.Contains(body, "no location for this event") {
		t.Errorf("the actions bar should still render without a scroll footer: %q", body)
	}
}

// TestChronicleDetailPaneActionsBarTruncatedWidth: a narrow pane still
// renders to exactly paneRows lines without panicking, even when the
// footer text and the action label together would overflow the width (the
// gap-clamping floor).
func TestChronicleDetailPaneActionsBarTruncatedWidth(t *testing.T) {
	e := store.Event{Seq: 1, Tick: 1, Type: "world.migrated",
		Payload: json.RawMessage(`{"from_format":2,"source_events":100,"source_tick":500,"state":{}}`)}
	out := chronicleDetailPane(e, nil, 0, 20, 3, "no location for this event")
	if len(out) != 3 {
		t.Fatalf("chronicleDetailPane must render exactly paneRows lines even under width pressure, got %d", len(out))
	}
}
