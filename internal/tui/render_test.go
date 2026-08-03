package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/evanstern/promptworld/internal/guardian"
	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
)

// TestWidescreenViewExactHeight is the B1 regression: the widescreen
// composite must render to EXACTLY m.height lines in every mode layered on
// top of it. Bubble Tea scrolls the top of the frame off-screen when a
// View() overflows the terminal height, which is what pushed the header
// row off-screen in the live tmux repro. Covers home, solo, and
// paused/inspect (both in the dock and solo'd), at several sizes straddling
// the breakpoint and the 50/50 column split.
//
// The state list and the posing logic are States()/poseState (spec 112 T007/
// T008, design.go) — the same registry and the same poser the frame harness
// drives, so this sweep and the dumped matrix can never disagree about what
// states exist or what one of them means.
func TestWidescreenViewExactHeight(t *testing.T) {
	sizes := []struct{ w, h int }{
		{112, 20}, {112, 30}, {113, 30}, {118, 30}, {140, 40}, {160, 50}, {200, 24},
	}
	for _, sz := range sizes {
		for _, state := range States() {
			t.Run(fmt.Sprintf("%dx%d/%s", sz.w, sz.h, state), func(t *testing.T) {
				m := widescreenModel(t)
				m.width, m.height = sz.w, sz.h
				seedEvents(&m, 20)
				if err := poseState(&m, state); err != nil {
					t.Fatal(err)
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
	rows := computeRows(m.height, false)
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
		familyGovernance, familyGru, familyChronicle, familyGuardian, familyCog,
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
	lines := styleWrapLine("TYPE  ", summary, 60, 1, 0)
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
	l := formatChronicleLine(e, []string{"Ash"}, nil)
	cols := computeChronicleColumns([]chronicleLine{l}, false)
	got := renderChronicleRow(l, cols, 60, 1, false)
	want := styleFeedAlert.Render(plainChronicleLine(l, cols))
	if got != want {
		t.Errorf("alert row should be styled whole-line with styleFeedAlert:\ngot:  %q\nwant: %q", got, want)
	}
}

// TestRenderChronicleRowNeglectAlertWholeLine (spec 083 FR-010, SC-005): the
// neglect percept renders whole-line in the alert role — membership in the
// shipped tier, the stranger.took precedent — and its summary carries the
// deterministic per-need wording (name + peril + inaction + level).
func TestRenderChronicleRowNeglectAlertWholeLine(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	e := store.Event{Seq: 1, Tick: 510180, Type: "sim.neglect_detected",
		Payload: json.RawMessage(`{"agent":0,"need":"warmth","level":0,"since":499320}`)}
	l := formatChronicleLine(e, []string{"Ash"}, nil)
	cols := computeChronicleColumns([]chronicleLine{l}, false)
	got := renderChronicleRow(l, cols, 120, 1, false)
	want := styleFeedAlert.Render(plainChronicleLine(l, cols))
	if got != want {
		t.Errorf("neglect row should be styled whole-line with styleFeedAlert:\ngot:  %q\nwant: %q", got, want)
	}
	if plain := plainChronicleLine(l, cols); !strings.Contains(plain, "Ash is dangerously cold and has done nothing about it (warmth 0)") {
		t.Errorf("neglect row wording drifted: %q", plain)
	}
}

// TestRenderChronicleRowLabeledVoiceWholeLine: labeled-voice families
// (cog/clock/daemon, contract §2) tint the whole line, not per-segment.
func TestRenderChronicleRowLabeledVoiceWholeLine(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	e := store.Event{Seq: 1, Tick: 60, Type: "clock.speed_set", Payload: json.RawMessage(`{"speed":"4x"}`)}
	l := formatChronicleLine(e, nil, nil)
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
		l := formatChronicleLine(e, []string{"Ash"}, nil)
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
		l := formatChronicleLine(e, names, nil)
		if strings.Contains(plainSegs(l.Summary), "\x1b") {
			t.Errorf("%s: digest summary contains an ESC byte — the pure layer must never emit ANSI", typ)
		}
		cols := computeChronicleColumns([]chronicleLine{l}, false)
		if strings.Contains(plainChronicleLine(l, cols), "\x1b") {
			t.Errorf("%s: plainChronicleLine contains an ESC byte", typ)
		}
	}
	fallback := store.Event{Seq: 1, Tick: 1, Type: "future.unknown_type", Payload: json.RawMessage(`{"agent":0}`)}
	l := formatChronicleLine(fallback, names, nil)
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
// the guardian pane shows nothing extra.
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

// TestOrderStatusLinesFields (spec 029 T023): the guardian pane's
// standing-orders block carries id, fuzzy marker, origin, expiry day, status,
// and condition for every order, headed by a count line.
func TestOrderStatusLinesFields(t *testing.T) {
	orders := []guardian.OrderStatus{
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

// TestSystemsViewRendersProviderTable (spec 053 US2/D10): the systems pane
// surfaces the per-provider table it's handed, not the old plain up/down
// line — the exact assertion TestGuardianViewRendersProviderTable made
// before the telemetry split, retargeted at its new home.
func TestSystemsViewRendersProviderTable(t *testing.T) {
	m := testModel(t)
	m.status = &ipc.StatusData{LLM: &llm.Status{
		Providers: []llm.ProviderStatus{
			{Name: "cogito", Model: "cogito:3b", Up: true, Queue: 1, Inflight: 1, Slots: 4, SpentUSD: 0},
		},
		Spent: 0, Budget: 100,
	}}
	view := m.systemsView()
	for _, want := range []string{"cogito", "cogito:3b", "q1", "1/4"} {
		if !strings.Contains(view, want) {
			t.Errorf("systems view missing provider field %q:\n%s", want, view)
		}
	}
}

// TestGuardianTabHasZeroTelemetryAfterSplit (spec 053 SC-002): the guardian
// tab (guardianView, narrow fallback) renders NONE of the relocated
// provider/spend/horizon content once an LLM-configured world would have
// rendered it pre-split — fiction-layer content only.
func TestGuardianTabHasZeroTelemetryAfterSplit(t *testing.T) {
	m := testModel(t)
	m.status = &ipc.StatusData{
		Clock: ipc.ClockStatus{Speed: "8x"},
		LLM: &llm.Status{
			Providers: []llm.ProviderStatus{
				{Name: "cogito", Model: "cogito:3b", Up: true, Queue: 1, Inflight: 1, Slots: 4, SpentUSD: 1.5},
			},
			Spent: 1.5, Budget: 100,
		},
		Horizon: []ipc.HorizonClass{{Class: "conversation", Suppressed: true, Calibrated: true}},
	}
	view := m.guardianView()
	for _, unwanted := range []string{"cogito", "cogito:3b", "spend $", "cognition horizon"} {
		if strings.Contains(view, unwanted) {
			t.Errorf("guardian tab must carry zero telemetry after the split, found %q:\n%s", unwanted, view)
		}
	}
}

// TestSystemsContentBodyNoLLMStatesHonestly (spec 053 US2 AS3, SC-002): a
// no-LLM world's systems tab states its absence honestly rather than
// rendering empty chrome.
func TestSystemsContentBodyNoLLMStatesHonestly(t *testing.T) {
	m := testModel(t)
	m.status = &ipc.StatusData{}
	body := m.systemsContentBody(80, 10)
	if !strings.Contains(body, "no LLM configured") {
		t.Errorf("no-LLM systems tab should state its absence honestly, got %q", body)
	}
	if strings.Contains(body, "cognition horizon") {
		t.Error("no-LLM systems tab must not render the horizon block")
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
		{"full bank omits regen", sim.GuardianChargeCap, 3, false, "👁 3 standing orders"},
		{"zero orders is a true zero", 1, 0, true, "👁 0 standing orders"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := testModel(t)
			m.status = &ipc.StatusData{Clock: ipc.ClockStatus{GuardianCharges: c.charges, Tick: 100}}
			m.replica.GuardianOrders = make([]sim.GuardianOrder, c.orders)
			for i := range m.replica.GuardianOrders {
				m.replica.GuardianOrders[i] = sim.GuardianOrder{ID: fmt.Sprintf("o%d", i), Status: "active"}
			}
			got := m.guardianStripView(200)
			// Spec 085 flipped the old absence pin to presence: a status with
			// NO faith field (an older daemon) renders the honest dashed form
			// — present, claiming nothing (guardian-strip.md §4).
			if !strings.Contains(got, "faith —") {
				t.Errorf("nil faith field must render the dashed segment: %q", got)
			}
			if hasRegen := strings.Contains(got, "next +1"); hasRegen != c.wantRegen {
				t.Errorf("regen segment presence = %v, want %v: %q", hasRegen, c.wantRegen, got)
			}
			if !strings.Contains(got, c.wantOrders) {
				t.Errorf("missing standing-order segment %q: %q", c.wantOrders, got)
			}
			bank := fmt.Sprintf("(%d/%d)", c.charges, sim.GuardianChargeCap)
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
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{GuardianCharges: 2, Tick: 100}}
	before := m.guardianStripView(200)
	if !strings.Contains(before, "(2/3)") {
		t.Fatalf("fixture setup wrong: %q", before)
	}

	// Spend a charge.
	m.status.Clock.GuardianCharges = 1
	afterSpend := m.guardianStripView(200)
	if afterSpend == before || !strings.Contains(afterSpend, "(1/3)") {
		t.Errorf("charge spend not reflected next frame: %q", afterSpend)
	}

	// Place a standing order (the same client-side replica mutation
	// applying a live metatron.order_placed event performs, tui.go
	// applyEvent/Apply).
	m.replica.GuardianOrders = append(m.replica.GuardianOrders, sim.GuardianOrder{ID: "ord-1", Status: "active"})
	afterOrder := m.guardianStripView(200)
	if !strings.Contains(afterOrder, "👁 1 standing orders") {
		t.Errorf("order add not reflected next frame: %q", afterOrder)
	}

	// An order transitioning to a terminal status is still retained
	// (pruneGuardianOrders keeps recent non-active ones alongside active
	// ones) — the count mirrors len(), exactly orderStatusLines' header
	// count, so it does not drop until the retention prune eventually runs.
	m.replica.GuardianOrders[0].Status = "expired"
	afterExpiry := m.guardianStripView(200)
	if !strings.Contains(afterExpiry, "👁 1 standing orders") {
		t.Errorf("retained-but-terminal order count should match orderStatusLines' len() semantics: %q", afterExpiry)
	}
}

// TestGuardianStripFaithStates (spec 085 US4 AS-1..3, SC-005): the §4 faith
// segment's three honest states — populated from the wire, dashed against an
// older daemon — plus the forecast honesty rules under the variable cadence:
// the wire cadence drives the boundary, cadence 0 (the scenario forsaken
// band) omits the segment, an older daemon falls back to the legacy exported
// constant, and faith drops FIRST under width pressure.
func TestGuardianStripFaithStates(t *testing.T) {
	faith := func(n int) *int { return &n }

	t.Run("populated from the wire", func(t *testing.T) {
		m := testModel(t)
		m.status = &ipc.StatusData{Clock: ipc.ClockStatus{GuardianCharges: 1, Tick: 100,
			Faith: faith(62), FaithRegenTicks: 12 * 3600}}
		got := m.guardianStripView(200)
		if !strings.Contains(got, "faith 62") {
			t.Errorf("populated segment missing: %q", got)
		}
		// The forecast reads the EFFECTIVE cadence off the wire: the next
		// absolute 12h boundary from tick 100 is tick 43200 — 18:00 on the
		// genesis clock (day 1 starts 06:00) — not the legacy 6h boundary.
		if !strings.Contains(got, "next +1 @ 18:00") {
			t.Errorf("forecast should use the wire cadence (12h): %q", got)
		}
	})

	t.Run("dashed against an older daemon, legacy forecast fallback", func(t *testing.T) {
		m := testModel(t)
		m.status = &ipc.StatusData{Clock: ipc.ClockStatus{GuardianCharges: 1, Tick: 100}}
		got := m.guardianStripView(200)
		if !strings.Contains(got, "faith —") {
			t.Errorf("dashed segment missing against a pre-085 daemon: %q", got)
		}
		// The legacy 6h boundary from tick 100 is tick 21600 — 12:00 on the
		// genesis clock.
		if !strings.Contains(got, "next +1 @ 12:00") {
			t.Errorf("forecast should fall back to the legacy 6h constant: %q", got)
		}
	})

	t.Run("cadence 0 omits the forecast (no regen scheduled)", func(t *testing.T) {
		m := testModel(t)
		m.status = &ipc.StatusData{Clock: ipc.ClockStatus{GuardianCharges: 0, Tick: 100,
			Faith: faith(5), FaithRegenTicks: 0}}
		got := m.guardianStripView(200)
		if strings.Contains(got, "next +1") {
			t.Errorf("scenario forsaken band must omit the forecast (never forecast an arrival that isn't scheduled): %q", got)
		}
		if !strings.Contains(got, "faith 5") {
			t.Errorf("faith segment still renders while regen is stopped: %q", got)
		}
	})

	t.Run("faith drops first under width pressure", func(t *testing.T) {
		m := testModel(t)
		m.status = &ipc.StatusData{Clock: ipc.ClockStatus{GuardianCharges: 1, Tick: 100,
			Faith: faith(62), FaithRegenTicks: 6 * 3600}}
		full := m.guardianStripView(200)
		if !strings.Contains(full, "faith 62") {
			t.Fatalf("fixture setup wrong: %q", full)
		}
		squeezed := m.guardianStripView(lipgloss.Width(full) - 1)
		if strings.Contains(squeezed, "faith") {
			t.Errorf("one column short must drop the faith segment first (§4 drop order): %q", squeezed)
		}
		if !strings.Contains(squeezed, "standing orders") {
			t.Errorf("orders must outlive faith under pressure: %q", squeezed)
		}
	})
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
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{GuardianCharges: 1, Tick: 100}}
	m.replica.GuardianOrders = []sim.GuardianOrder{{ID: "o1", Status: "active"}, {ID: "o2", Status: "active"}}
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
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{GuardianCharges: 2, Tick: 100}}
	m.replica.GuardianOrders = []sim.GuardianOrder{{ID: "o1", Status: "active"}}

	// Tall enough: strip stays on, dormant line is the plain placeholder.
	m.height = 40
	if rows := computeRows(m.height, false); rows.Strip == 0 {
		t.Fatalf("fixture setup: expected the strip to fit at height %d", m.height)
	}
	unfolded := m.minibufferView(m.width)
	if strings.Contains(unfolded, "👁") {
		t.Errorf("unfolded dormant line must not carry the budget prefix: %q", unfolded)
	}

	// Short enough: strip folds, dormant line carries the relocated prefix.
	m.height = 14
	if rows := computeRows(m.height, false); rows.Strip != 0 {
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
	base.status = &ipc.StatusData{Clock: ipc.ClockStatus{GuardianCharges: 2, Tick: 100}}
	base.replica.GuardianOrders = []sim.GuardianOrder{{ID: "o1", Status: "active"}}

	tall, short := base, base
	tall.height, short.height = 40, 14
	if computeRows(tall.height, false).Strip == 0 || computeRows(short.height, false).Strip != 0 {
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

// TestGuardianViewNarrowCarriesStrip (US3 AS-2; layout.md ruling b): the
// narrow fallback's only minibuffer (inside the guardian pane — narrow's
// other panes have no composer at all) carries the strip above it,
// unconditionally — narrow has no fold machinery of its own to drop it.
func TestGuardianViewNarrowCarriesStrip(t *testing.T) {
	m := testModel(t) // narrow: 80x30 (below widescreenBreakpoint)
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{GuardianCharges: 1, Tick: 100}}
	m.replica.GuardianOrders = []sim.GuardianOrder{{ID: "o1", Status: "active"}, {ID: "o2", Status: "active"}}

	view := m.guardianView()
	if !strings.Contains(view, "(1/3)") {
		t.Errorf("narrow guardian pane missing the strip's charge bank: %q", view)
	}
	if !strings.Contains(view, "👁 2 standing orders") {
		t.Errorf("narrow guardian pane missing the strip's standing-order count: %q", view)
	}
	stripIdx := strings.Index(view, "(1/3)")
	mbIdx := strings.Index(view, "⏎ m — speak with the guardian…")
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

// --- spec 056: takeover surfaces — the shared report-card renderer,
// the ambient/scored boundary, and exact-height ---

// TestReportCardThreeSiteEquivalence (SC-005): the same rubric fixture
// renders identical rows (term + backing) at every site — concluded vs live
// differ ONLY in the marker vocabulary for an unmet term — and composes
// unmodified into the console's card seam (spec 053).
func TestReportCardThreeSiteEquivalence(t *testing.T) {
	facts := []reportCardFact{
		{Term: "village survives to dawn", Met: false, Backing: "agent.died: 2"},
		{Term: "metatron order placed", Met: true, Backing: "guardian.order_placed: 1"},
	}
	concluded := reportCardView("first-night", facts, reportCardConcluded, 60)
	live := reportCardView("first-night", facts, reportCardLive, 60)

	for _, f := range facts {
		if !strings.Contains(concluded, f.Term) || !strings.Contains(concluded, f.Backing) {
			t.Errorf("concluded card missing term/backing %q/%q:\n%s", f.Term, f.Backing, concluded)
		}
		if !strings.Contains(live, f.Term) || !strings.Contains(live, f.Backing) {
			t.Errorf("live card missing term/backing %q/%q:\n%s", f.Term, f.Backing, live)
		}
	}
	if !strings.Contains(concluded, "✗") {
		t.Errorf("concluded mode's unmet marker should be ✗: %q", concluded)
	}
	if strings.Contains(live, "✗") {
		t.Errorf("live mode must never render the concluded ✗ marker: %q", live)
	}
	if !strings.Contains(live, "…") {
		t.Errorf("live mode's unmet marker should be …: %q", live)
	}

	// Seam composition (spec.md US3-AS2): a report card wrapped as a
	// consoleCard renders byte-identically to a direct call.
	card := reportCard{title: "first-night", facts: facts, mode: reportCardConcluded}
	if got, want := card.renderCard(60), concluded; got != want {
		t.Errorf("consoleCard wrapper diverged from a direct reportCardView call:\ngot:  %q\nwant: %q", got, want)
	}
}

// TestPostmortemReportCardFailedTermRendersMissed is spec 072's motivating
// regression (SC-001): a run-ended first-night world with two recorded
// deaths renders ✗ on "no villager dies" with the honest backing count —
// never the presence-derived ✓ the deleted generic builders produced
// ("agent died (agent.died: 2)" used to grade MET because the event type
// was merely present).
func TestPostmortemReportCardFailedTermRendersMissed(t *testing.T) {
	m := testModel(t)
	m.w.Manifest.Scenario = &world.ScenarioConfig{Exercise: "first-night"}
	m.replica.Deaths = []sim.DeathRecord{
		{Agent: 0, Tick: 90, Cause: "gru"},
		{Agent: 1, Tick: 95, Cause: "exposure"},
	}
	m.applyEvent(runEndedEvent(1, 100, "gru", m.replica.Deaths))

	card := m.postmortemReportCard(100)
	if !strings.Contains(card, "✗ no villager dies (agent.died: 2)") {
		t.Fatalf("failed death term must render ✗ with its honest count:\n%s", card)
	}
	if strings.Contains(card, "✓ no villager dies") {
		t.Fatalf("presence-derived ✓ on a failing zero-wanted term — the exact lie spec 072 deletes:\n%s", card)
	}
	// Labels are the evaluator's hand-authored plain language, not the
	// mechanical event-type gloss (US1-5).
	if strings.Contains(card, "agent died (") {
		t.Errorf("mechanical humanizeEventType gloss leaked back in:\n%s", card)
	}
	if !strings.Contains(card, "village survives to dawn of day 2") {
		t.Errorf("hand-authored rubric label missing:\n%s", card)
	}
	// The postmortem is a concluded surface: unmet terms carry ✗, never the
	// live pending marker.
	if strings.Contains(card, "…") {
		t.Errorf("concluded surface rendered a live pending marker:\n%s", card)
	}
}

// TestReportCardCrossSurfaceIdenticalRows (spec 072 US1-4): the postmortem
// overlay and the console checklist card render identical glyph+label rows
// for the same exercise at the same replica state — one shared fact
// resolver, by construction.
func TestReportCardCrossSurfaceIdenticalRows(t *testing.T) {
	m := testModel(t)
	m.w.Manifest.Scenario = &world.ScenarioConfig{Exercise: "first-night"}
	m.replica.Deaths = []sim.DeathRecord{{Agent: 0, Tick: 90, Cause: "gru"}}
	m.applyEvent(runEndedEvent(1, 100, "gru", m.replica.Deaths))

	rows := func(view string) []string {
		var out []string
		for _, line := range strings.Split(view, "\n") {
			if i := strings.IndexAny(line, "✓✗…"); i >= 0 {
				out = append(out, strings.TrimSpace(strings.Trim(line[i:], "│ ")))
			}
		}
		return out
	}
	postmortem := rows(m.postmortemReportCard(100))
	card := m.buildChecklistCard()
	if card == nil {
		t.Fatal("console checklist card absent on an ended scored run")
	}
	console := rows(card.renderCard(100))
	if len(postmortem) == 0 || fmt.Sprint(postmortem) != fmt.Sprint(console) {
		t.Errorf("surfaces disagree:\npostmortem: %v\nconsole:    %v", postmortem, console)
	}
}

// TestPostmortemAmbientScoredBoundary (SC-002): zero report cards on ambient
// runs, one on scored runs, across the fixture matrix — the report card
// never renders on missing/unrecognized exercise data (spec.md Edge Cases).
func TestPostmortemAmbientScoredBoundary(t *testing.T) {
	cases := []struct {
		name     string
		scenario *world.ScenarioConfig
		wantCard bool
	}{
		{"no scenario block (ambient)", nil, false},
		{"scenario names an unknown exercise", &world.ScenarioConfig{Exercise: "no-such-exercise"}, false},
		{"scenario names a cataloged exercise (scored)", &world.ScenarioConfig{Exercise: "first-night"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := testModel(t)
			m.w.Manifest.Scenario = tc.scenario
			m.applyEvent(runEndedEvent(1, 100, "exposure", []sim.DeathRecord{{Agent: 0, Tick: 100, Cause: "exposure"}}))
			view := m.postmortemView(100, 30)
			hasCard := strings.Contains(view, "report card")
			if hasCard != tc.wantCard {
				t.Errorf("report card present = %v, want %v:\n%s", hasCard, tc.wantCard, view)
			}
			// The morgue register renders regardless (FR-001: "always").
			if !strings.Contains(view, "morgue") {
				t.Errorf("morgue evidence section must always render: %s", view)
			}
		})
	}
}

// TestMorgueRowsCharterObservationHonesty (research R3): a death's closest
// charter observation resolves from the client's own chronicle ring; absent
// any such event, it renders "unknown" rather than guessing.
func TestMorgueRowsCharterObservationHonesty(t *testing.T) {
	m := testModel(t)
	m.replica.Agents = []sim.Agent{{Name: "Ash"}}
	m.applyEvent(store.Event{Seq: 1, Tick: 50, Type: "guardian.charter_observed",
		Payload: json.RawMessage(`{"fingerprint":"abc123","default":false}`)})
	m.applyEvent(runEndedEvent(2, 100, "exposure", []sim.DeathRecord{{Agent: 0, Tick: 100, Cause: "exposure"}}))
	rows := m.morgueRows()
	if len(rows) != 1 {
		t.Fatalf("morgueRows() = %v, want 1 row", rows)
	}
	if rows[0].Charter != "abc123" {
		t.Errorf("Charter = %q, want the recorded fingerprint", rows[0].Charter)
	}

	m2 := testModel(t)
	m2.replica.Agents = []sim.Agent{{Name: "Rowan"}}
	m2.applyEvent(runEndedEvent(1, 100, "exposure", []sim.DeathRecord{{Agent: 0, Tick: 100, Cause: "exposure"}}))
	rows2 := m2.morgueRows()
	if len(rows2) != 1 || rows2[0].Charter != "unknown" {
		t.Errorf("Charter = %v, want \"unknown\" with no recorded observation", rows2)
	}
}

// TestCeremonyReportCardPrefersRecordedEvidence: with the qualifying pass
// still retained, the ceremony's checklist is a re-read of that record
// (spec 072 FR-002) — every term met, backing drawn from the pass's own
// Evidence (authoritative — FR-019), labels from the evaluator's
// hand-authored terms, never a live re-grade.
func TestCeremonyReportCardPrefersRecordedEvidence(t *testing.T) {
	m := testModel(t)
	m.replica.CurriculumPasses = []sim.CurriculumPass{
		{Exercise: "first-night", Stage: "stage-1", Evidence: []sim.EvidenceRef{
			{Type: "sim.day_started", Seq: 40, Tick: 890},
			{Type: "guardian.order_placed", Seq: 12, Tick: 200},
		}},
	}
	m.applyEvent(stageUnlockedEvent(1, 200, "stage-2", "first-night"))
	view := m.ceremonyView(100, 30)
	if !strings.Contains(view, "report card · first-night") {
		t.Fatalf("ceremony should render the proving exercise's report card: %q", view)
	}
	// Evidence-backed rows carry the recorded refs; the recorded pass proves
	// every term held at pass time, so NO unmet marker may render even
	// though the replica's own live rubric would still grade terms pending.
	if !strings.Contains(view, "sim.day_started · seq 40") || !strings.Contains(view, "guardian.order_placed · seq 12") {
		t.Errorf("ceremony report card should carry the recorded pass's own Evidence refs: %q", view)
	}
	if !strings.Contains(view, "✓ village survives to dawn of day 2") || !strings.Contains(view, "✓ no villager dies") {
		t.Errorf("recorded pass must render all terms met with evaluator labels: %q", view)
	}
	if strings.Contains(view, "✗") || strings.Contains(view, "…") {
		t.Errorf("recorded-pass card must carry no unmet marker (re-read, never re-grade): %q", view)
	}
}

// TestCeremonyReportCardAgedOutPassFallsBackToRubric (spec 072 edge case,
// pinned as a decision): the qualifying pass has left the bounded
// CurriculumPasses retention, so the ceremony grades sim.EvaluateRubric
// over the current replica with concluded markers — current-state truth,
// honest about what it can still see.
func TestCeremonyReportCardAgedOutPassFallsBackToRubric(t *testing.T) {
	m := testModel(t)
	m.applyEvent(stageUnlockedEvent(1, 200, "stage-2", "first-night"))
	if len(m.replica.CurriculumPasses) != 0 {
		t.Fatal("fixture: retention must be empty (the aged-out posture)")
	}
	view := m.ceremonyView(100, 30)
	if !strings.Contains(view, "report card · first-night") {
		t.Fatalf("aged-out fallback should still render the proving exercise's card: %q", view)
	}
	// Concluded vocabulary over current state: at tick 0 the survival and
	// watch terms are not provably met — ✗, never a false ✓ and never the
	// live pending marker on this concluded surface.
	if !strings.Contains(view, "✗ village survives to dawn of day 2") || !strings.Contains(view, "✓ no villager dies (agent.died: 0)") {
		t.Errorf("aged-out card should grade the current replica with concluded markers: %q", view)
	}
	if strings.Contains(view, "…") {
		t.Errorf("concluded surface rendered a live pending marker: %q", view)
	}
}

// TestTakeoverExactHeight (B1/B5): both takeover kinds render to exactly
// m.height lines, in narrow and widescreen, across several sizes.
func TestTakeoverExactHeight(t *testing.T) {
	sizes := []struct{ w, h int }{
		{80, 24}, {80, 30}, {112, 20}, {140, 40}, {200, 24},
	}
	for _, sz := range sizes {
		for _, kind := range []string{"ceremony", "postmortem"} {
			t.Run(fmt.Sprintf("%dx%d/%s", sz.w, sz.h, kind), func(t *testing.T) {
				m := testModel(t)
				m.width, m.height = sz.w, sz.h
				switch kind {
				case "ceremony":
					m.applyEvent(stageUnlockedEvent(1, 100, "stage-2", "first-night"))
				case "postmortem":
					m.applyEvent(runEndedEvent(1, 100, "exposure", []sim.DeathRecord{{Agent: 0, Tick: 100, Cause: "exposure"}}))
				}
				view := m.View()
				if got := lipgloss.Height(view); got != sz.h {
					t.Errorf("View() height = %d, want %d:\n%s", got, sz.h, view)
				}
			})
		}
	}
}
