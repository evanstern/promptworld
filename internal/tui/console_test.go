package tui

// Guardian console tests (spec 053-guardian-console): the full-screen page
// (T003/T006/T007/T008/T009), the charter/skills read surface (T010), and
// the $EDITOR handoff (T011). The systems-tab split (T002/T004/T005) is
// covered in tui_test.go/render_test.go alongside the renderers it
// relocates.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evanstern/promptworld/internal/guardian"
	"github.com/evanstern/promptworld/internal/sim"
)

// --- T003: console page state — open/close, return target, esc ordering ---

// TestConsoleOpensFromEveryMode (spec.md US1 AS1): 'G' opens the console
// full-screen from home, solo, and narrow, without disturbing dockTab/solo/
// active — those already ARE the "prior view" (research R1), so closing
// needs nothing but the flag flip to restore them.
func TestConsoleOpensFromEveryMode(t *testing.T) {
	cases := []struct {
		name  string
		build func(t *testing.T) Model
	}{
		{"widescreen home", func(t *testing.T) Model { return widescreenModel(t) }},
		{"widescreen solo", func(t *testing.T) Model {
			// paneGuardian, not paneVillagers/paneChronicle: those tabs'
			// own key layers claim "G" for their own jump-to-last
			// (villagersVisible()/inspecting() are checked ahead of
			// handleGlobalKey regardless of solo — see
			// TestConsoleGShadowedByInspectAndVillagersJump) — this case
			// exercises the plain "solo, G reaches global" path.
			m := widescreenModel(t)
			m.solo = true
			m.dockTab = paneGuardian
			return m
		}},
		{"narrow", func(t *testing.T) Model { return testModel(t) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.build(t)
			before := m
			var mdl tea.Model = m
			mdl = update(mdl, "G")
			mm := mdl.(Model)
			if !mm.console {
				t.Fatal("'G' must open the console")
			}
			if mm.dockTab != before.dockTab || mm.solo != before.solo || mm.active != before.active {
				t.Errorf("opening the console must not touch dockTab/solo/active: got dockTab=%s solo=%v active=%s, want dockTab=%s solo=%v active=%s",
					paneNames[mm.dockTab], mm.solo, paneNames[mm.active],
					paneNames[before.dockTab], before.solo, paneNames[before.active])
			}
			if v := mm.View(); v == "" {
				t.Fatal("console view rendered empty")
			}

			// 'G' again closes and restores exactly the prior view.
			mdl = update(mdl, "G")
			after := mdl.(Model)
			if after.console {
				t.Fatal("'G' again must close the console")
			}
			if after.View() != before.View() {
				t.Errorf("closing the console did not restore the exact prior view:\nbefore: %q\nafter:  %q", before.View(), after.View())
			}
		})
	}
}

// TestConsoleClosesWithOneAndEsc: '1' and unfocused 'esc' both close the
// console (contract §1), same as 'G'.
func TestConsoleClosesWithOneAndEsc(t *testing.T) {
	for _, k := range []string{"1", "esc"} {
		t.Run(k, func(t *testing.T) {
			m := widescreenModel(t)
			var mdl tea.Model = m
			mdl = update(mdl, "G")
			if !mdl.(Model).console {
				t.Fatal("setup: expected the console open")
			}
			mdl = update(mdl, k)
			if mdl.(Model).console {
				t.Fatalf("%q must close the console", k)
			}
		})
	}
}

// TestConsoleGShadowedByInspectAndVillagersJump: FR-001 scopes the console
// key to "home, solo, and narrow" — inspect mode (chronJumpLast) and
// villagers mode (villJumpLast) already bind 'G' to their own jump, and
// keep it (deliberate scoping, not a regression): the console never opens
// out from under an in-progress jump-to-last.
func TestConsoleGShadowedByInspectAndVillagersJump(t *testing.T) {
	t.Run("inspect", func(t *testing.T) {
		m := pausedModel(t)
		var mdl tea.Model = m
		mdl = update(mdl, "G")
		mm := mdl.(Model)
		if mm.console {
			t.Fatal("'G' while inspecting must jump to last event, not open the console")
		}
		if mm.chronSelected != len(mm.events)-1 {
			t.Errorf("'G' should have jumped to the last event: chronSelected=%d", mm.chronSelected)
		}
	})
	t.Run("villagers roster", func(t *testing.T) {
		m := widescreenModel(t)
		m.dockTab = paneVillagers
		m.replica.Agents = append(m.replica.Agents, sim.Agent{Name: "Ash"}, sim.Agent{Name: "Birch"})
		wantIdx := len(m.replica.Agents) - 1
		var mdl tea.Model = m
		mdl = update(mdl, "G")
		mm := mdl.(Model)
		if mm.console {
			t.Fatal("'G' in the villagers roster must jump to last villager, not open the console")
		}
		if mm.villSelected != wantIdx {
			t.Errorf("'G' should have jumped to the last villager (%d): villSelected=%d", wantIdx, mm.villSelected)
		}
	})
}

// TestConsoleEscReleasesConsoleLayer: esc-release ordering gains the console
// layer (research R1: "minibuffer -> villager detail -> console -> solo ->
// home"). With villager detail left open underneath (dockTab/active
// untouched by opening the console), esc closes the visible console first;
// a second esc then releases the villager detail exactly as it always did.
// The console is opened from the chronicle tab (not villagers) — while
// villagersVisible() is true, "G" is villJumpLast's own key regardless of
// villDetail (see TestConsoleGShadowedByInspectAndVillagersJump), so there
// is no key-driven path to reach BOTH villager detail AND the console open
// at once; dockTab/villDetail are set directly here to construct that
// state, honestly documenting the gap rather than a keypress that cannot
// happen.
func TestConsoleEscReleasesConsoleLayer(t *testing.T) {
	m := widescreenModel(t)
	var mdl tea.Model = m
	mdl = update(mdl, "G")
	if !mdl.(Model).console {
		t.Fatal("setup: expected the console open")
	}
	mm := mdl.(Model)
	mm.dockTab = paneVillagers
	mm.villDetail = true
	mdl = mm
	mdl = update(mdl, "esc")
	mm = mdl.(Model)
	if mm.console {
		t.Fatal("esc must close the console (the visible layer) first")
	}
	if !mm.villDetail {
		t.Fatal("esc closing the console must not also release villager detail underneath — one layer per press")
	}
	mdl = update(mdl, "esc")
	if mdl.(Model).villDetail {
		t.Fatal("a second esc should now release villager detail")
	}
}

// TestConsoleScrollResetsOnClose (data-model.md "consoleScroll... reset on
// close" — "reading posture, not archive").
func TestConsoleScrollResetsOnClose(t *testing.T) {
	m := widescreenModel(t)
	var mdl tea.Model = m
	mdl = update(mdl, "G")
	mm := mdl.(Model)
	mm.consoleScroll = 5
	mdl = mm
	mdl = update(mdl, "G")
	if got := mdl.(Model).consoleScroll; got != 0 {
		t.Errorf("consoleScroll = %d, want 0 after close", got)
	}
}

// --- Focus contract: the console introduces no second input widget ---

// TestConsoleKeysTypeWhileMinibufferFocused (focus-contract.md "no silent
// stealing"; spec.md "Busy minibuffer when e pressed"): while the
// minibuffer is focused inside the console, G/e/5/J/K all type into the
// buffer instead of acting as console keys.
func TestConsoleKeysTypeWhileMinibufferFocused(t *testing.T) {
	m := widescreenModel(t)
	var mdl tea.Model = m
	mdl = update(mdl, "G")
	mdl = update(mdl, "m")
	if !mdl.(Model).mbFocused {
		t.Fatal("setup: expected the minibuffer focused")
	}
	for _, k := range []string{"G", "e", "5", "J", "K"} {
		mdl = update(mdl, k)
	}
	mm := mdl.(Model)
	if !mm.console || !mm.mbFocused {
		t.Fatal("typing must not have closed the console or released focus")
	}
	if got := mm.mbInput; got != "Ge5JK" {
		t.Fatalf("mbInput = %q, want the literal keys typed (no console key stole them)", got)
	}
}

// TestConsoleMFocusesInPlaceWithoutTouchingActive: the console's 'm' must
// never call selectTab (which would overwrite m.active and corrupt the
// narrow "return to the pane you left" restore) — unlike the global
// focusMinibuffer, which does switch panes in the narrow fallback.
func TestConsoleMFocusesInPlaceWithoutTouchingActive(t *testing.T) {
	m := testModel(t) // narrow
	m.active = paneChronicle
	var mdl tea.Model = m
	mdl = update(mdl, "G")
	mdl = update(mdl, "m")
	mm := mdl.(Model)
	if !mm.mbFocused {
		t.Fatal("'m' must focus the minibuffer")
	}
	if mm.active != paneChronicle {
		t.Errorf("console 'm' must not touch active: got %s, want chronicle", paneNames[mm.active])
	}
}

// TestConsoleCountsAsGuardianVisible (spec.md US1 AS4): a reply landing
// while the console is open streams in place — no second badge system.
func TestConsoleCountsAsGuardianVisible(t *testing.T) {
	m := widescreenModel(t)
	m.dockTab = paneChronicle // NOT the guardian tab underneath
	var mdl tea.Model = m
	mdl = update(mdl, "G")
	mdl, _ = mdl.(Model).Update(consoleReplyMsg{result: &guardian.TurnResult{Reply: "hello"}})
	mm := mdl.(Model)
	if mm.guardianUnseen {
		t.Error("a reply landing while the console is open must not badge the guardian tab")
	}
	if mm.mbFlash != "" {
		t.Error("a reply landing while the console is open must not flash the minibuffer")
	}
}

// --- T006: document-style turn rendering ---

func TestConsoleTurnLinesLabelsAndWraps(t *testing.T) {
	transcript := []string{"you: why is Rowan hoarding wood?", transcriptGuardianPrefix + "Rowan's memory holds three nights of Ash letting the fire die."}
	lines := consoleTurnLines(transcript, 40, nil)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "you") {
		t.Errorf("missing 'you' label: %q", joined)
	}
	if !strings.Contains(joined, "guardian") {
		t.Errorf("missing guardian (epithet) label: %q", joined)
	}
	if !strings.Contains(joined, "why is Rowan hoarding wood?") {
		t.Errorf("missing turn text: %q", joined)
	}
	// Blank-line separation between turns (contract §2.2).
	blankFound := false
	for _, l := range lines {
		if l == "" {
			blankFound = true
		}
	}
	if !blankFound {
		t.Error("turns must be blank-line separated")
	}
}

// TestConsoleTurnLinesSpecialRowsInline (research R4): the special-row
// vocabulary (⚡/👁/⏲/») renders unlabeled, inline, exactly as classified by
// the shared classifyTranscriptLine the compact tab uses — one vocabulary,
// two renderings.
func TestConsoleTurnLinesSpecialRowsInline(t *testing.T) {
	transcript := []string{
		"⚡ vision → Ash: \"the fire dies\"",
		"👁 watch set (ord-3): \"gru sighted\"",
		"⏲ paused",
	}
	lines := consoleTurnLines(transcript, 60, nil)
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"⚡ vision", "👁 watch set", "⏲ paused"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing special row %q: %q", want, joined)
		}
	}
	// Special rows carry no "label\ntext" split — they render as one line,
	// same shape classifyTranscriptLine's label=="" case always has.
	for _, l := range lines {
		if l == "you" || l == "guardian" {
			t.Errorf("a special row must never be classified as a conversational turn: %q", joined)
		}
	}
}

// TestConsoleTurnLinesOmitTimestampHonestly (research R4): Model.transcript
// carries no per-entry timestamp in this client — the console never
// invents one.
func TestConsoleTurnLinesOmitTimestampHonestly(t *testing.T) {
	lines := consoleTurnLines([]string{"you: hello"}, 40, nil)
	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, ":") == false {
		// sanity: no assumption
	}
	// The label line itself must be exactly "you" — no invented "· HH:MM".
	found := false
	for _, l := range lines {
		if strings.TrimSpace(l) == "you" {
			found = true
		}
		if strings.Contains(l, "·") {
			t.Errorf("must not invent a timestamp suffix: %q", l)
		}
	}
	if !found {
		t.Error("expected a bare 'you' label line")
	}
}

// --- consoleScrollWindow: tail-anchored windowing ---

func TestConsoleScrollWindowTailAnchoredAndPads(t *testing.T) {
	content := []string{"a", "b", "c"}
	// Short content pads at the bottom (hugs the top, B1 discipline).
	got := consoleScrollWindow(content, 0, 5)
	want := []string{"a", "b", "c", "", ""}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("got %v, want %v", got, want)
	}

	long := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	tail := consoleScrollWindow(long, 0, 3)
	if strings.Join(tail, ",") != "8,9,10" {
		t.Errorf("scroll=0 should show the tail: got %v", tail)
	}
	scrolledUp := consoleScrollWindow(long, 3, 3)
	if strings.Join(scrolledUp, ",") != "5,6,7" {
		t.Errorf("scroll=3 should move back 3 lines from the tail: got %v", scrolledUp)
	}
	// Clamped: scroll past the head never produces fewer than `rows` lines
	// nor drifts blank.
	clamped := consoleScrollWindow(long, 999, 3)
	if strings.Join(clamped, ",") != "1,2,3" {
		t.Errorf("over-scroll should clamp at the head: got %v", clamped)
	}
}

// TestConsoleScrollKeysJAndK: K scrolls toward older content (increments),
// J scrolls back toward the tail (decrements, floored at 0).
func TestConsoleScrollKeysJAndK(t *testing.T) {
	m := widescreenModel(t)
	var mdl tea.Model = m
	mdl = update(mdl, "G")
	mdl = update(mdl, "K")
	mdl = update(mdl, "K")
	if got := mdl.(Model).consoleScroll; got != 2 {
		t.Fatalf("two K's: consoleScroll = %d, want 2", got)
	}
	mdl = update(mdl, "J")
	if got := mdl.(Model).consoleScroll; got != 1 {
		t.Fatalf("one J: consoleScroll = %d, want 1", got)
	}
	mdl = update(mdl, "J")
	mdl = update(mdl, "J")
	if got := mdl.(Model).consoleScroll; got != 0 {
		t.Fatalf("J past 0 should floor at 0: got %d", got)
	}
}

// --- T007: composer pairing — byte-identical minibuffer ---

// TestConsoleComposerIsTheSameMinibuffer: the console's composer renders
// via the exact same minibufferView call every other page uses — proven by
// asserting the console's rendered composer line is byte-identical to a
// standalone call at the same width, across all three states.
func TestConsoleComposerIsTheSameMinibuffer(t *testing.T) {
	m := widescreenModel(t)
	m.connected = true
	var mdl tea.Model = m
	mdl = update(mdl, "G")

	cases := []func(m Model) Model{
		func(m Model) Model { return m }, // dormant
		func(m Model) Model { m.mbFocused = true; return m },
		func(m Model) Model { m.mbBusy = true; return m },
	}
	for _, tc := range cases {
		mm := tc(mdl.(Model))
		want := mm.minibufferView(mm.width)
		view := mm.consoleView()
		if !strings.Contains(view, strings.TrimRight(want, "\n")) {
			t.Errorf("console composer diverges from the standalone minibufferView render:\nwant substring: %q\ngot view:       %q", want, view)
		}
	}
}

// --- T008: the card seam ---

// fakeConsoleCard pins the composition seam's position with a test fake
// (data-model.md "Tests pin the composition position with a test fake") —
// this feature ships zero real producers (TASK-127/115 land later).
type fakeConsoleCard struct{ label string }

func (f fakeConsoleCard) renderCard(width int) string { return "CARD:" + f.label }

// TestConsoleCardSeamComposesBetweenStreamAndReadSurface: a fake card
// injected into Model.consoleCards renders after the turn stream and before
// the read surface (contract §2.3).
func TestConsoleCardSeamComposesBetweenStreamAndReadSurface(t *testing.T) {
	m := widescreenModel(t)
	m.connected = true
	m.transcript = []string{"you: hello"}
	m.consoleCards = []consoleCard{fakeConsoleCard{label: "first-night"}}
	var mdl tea.Model = m
	mdl = update(mdl, "G")
	view := mdl.(Model).consoleView()

	cardIdx := strings.Index(view, "CARD:first-night")
	turnIdx := strings.Index(view, "hello")
	readSurfaceIdx := strings.Index(view, "charter · skills")
	if cardIdx < 0 {
		t.Fatal("card content missing from the console view")
	}
	if !(turnIdx < cardIdx && cardIdx < readSurfaceIdx) {
		t.Errorf("card seam must compose between the turn stream and the read surface: turn=%d card=%d readSurface=%d", turnIdx, cardIdx, readSurfaceIdx)
	}
}

// TestConsoleCardSeamEmptyByDefault: this feature ships zero producers.
func TestConsoleCardSeamEmptyByDefault(t *testing.T) {
	m := widescreenModel(t)
	if len(m.consoleCards) != 0 {
		t.Error("consoleCards must ship empty this feature")
	}
}

// --- T009: footer hints ---

func TestConsoleFooterHints(t *testing.T) {
	m := widescreenModel(t)
	m.connected = true
	var mdl tea.Model = m
	mdl = update(mdl, "G")
	footer := mdl.(Model).footerView()
	for _, want := range []string{"G back", "esc back", "m ask", "space pause", "q quit", "? help"} {
		if !strings.Contains(footer, want) {
			t.Errorf("console footer missing %q: %q", want, footer)
		}
	}
}

func TestGlobalFooterAdvertisesConsole(t *testing.T) {
	m := widescreenModel(t)
	if !strings.Contains(m.footerView(), "G console") {
		t.Errorf("the global footer must advertise 'G' (FR-011): %q", m.footerView())
	}
	narrow := testModel(t)
	if !strings.Contains(narrow.footerView(), "G console") {
		t.Errorf("the narrow footer must advertise 'G' (FR-011): %q", narrow.footerView())
	}
}

// TestConsoleHelpOverlayReplacesBodyOnly (spec.md Edge Cases: "the console
// is a page, not an overlay — '?' over the console behaves as over any
// page"): '?' opens help over the console; the header/composer/footer chrome
// (outside the swapped body) survives, and dismissal restores the console
// exactly.
func TestConsoleHelpOverlayReplacesBodyOnly(t *testing.T) {
	m := widescreenModel(t)
	var mdl tea.Model = m
	mdl = update(mdl, "G")
	before := mdl.(Model).View()
	mdl = update(mdl, "?")
	mm := mdl.(Model)
	if !mm.helpOpen || !mm.console {
		t.Fatal("'?' over the console must open help without closing the console")
	}
	view := mm.View()
	if !strings.Contains(view, "HELP") {
		t.Errorf("help body must replace the console's reading region: %q", view)
	}
	mdl = update(mdl, "esc")
	after := mdl.(Model)
	if after.helpOpen {
		t.Fatal("esc must dismiss help")
	}
	if !after.console {
		t.Fatal("esc dismissing help must not also close the console (one layer per press)")
	}
	if after.View() != before {
		t.Errorf("dismissing help did not restore the exact prior console view:\nbefore: %q\nafter:  %q", before, after.View())
	}
}

// TestConsoleHelpModeRoutesToGlobalOrSolo: currentHelpMode() must not
// mis-route through a stale villagers/inspect state hiding underneath the
// console (dockTab/active persist unchanged while console is open). The
// console is opened first (from the chronicle tab — villagersVisible()
// would otherwise shadow 'G' with villJumpLast, same as
// TestConsoleEscReleasesConsoleLayer), then dockTab/villDetail are set
// directly to simulate villager detail left open underneath.
func TestConsoleHelpModeRoutesToGlobalOrSolo(t *testing.T) {
	m := widescreenModel(t)
	var mdl tea.Model = m
	mdl = update(mdl, "G")
	if !mdl.(Model).console {
		t.Fatal("setup: expected the console open")
	}
	mm := mdl.(Model)
	mm.dockTab = paneVillagers
	mm.villDetail = true
	mdl = mm
	mdl = update(mdl, "?")
	if got := mdl.(Model).helpMode; got != helpModeGlobal {
		t.Errorf("helpMode = %v, want helpModeGlobal (console open, widescreen) despite villDetail underneath", got)
	}
}

// --- T010: charter/skills read surface fixtures (SC-004) ---

func TestCharterReadSurfaceFixtures(t *testing.T) {
	cases := []struct {
		name   string
		status *guardian.Status
		want   []string
		absent []string
	}{
		{
			name:   "default charter, pre-ladder",
			status: &guardian.Status{CharterDefault: true},
			want:   []string{"default, binds now"},
		},
		{
			name:   "player-authored, pre-ladder",
			status: &guardian.Status{CharterDefault: false, Skills: []string{"a.md", "b.md"}},
			want:   []string{"player-authored, binds now", "2 files, binds now"},
		},
		{
			name: "stage-1 tutor preset: charter locked, skills locked",
			status: &guardian.Status{
				Stage: "stage-1", CharterLocked: true, CharterPreset: "tutor", SkillsLocked: true,
			},
			want:   []string{"preset-locked to tutor", "unlocks instruction authoring", "skills/ — locked", "unlocks skill files"},
			absent: []string{"binds now"},
		},
		{
			name: "stage-2: charter unlocked, skills still locked",
			status: &guardian.Status{
				Stage: "stage-2", CharterDefault: true, SkillsLocked: true,
			},
			want:   []string{"default, binds now", "skills/ — locked"},
			absent: []string{"preset-locked"},
		},
		{
			name: "stage-3+: everything binds",
			status: &guardian.Status{
				Stage: "stage-3", CharterDefault: false, Skills: []string{"a.md"},
			},
			want:   []string{"player-authored, binds now", "1 file, binds now"},
			absent: []string{"locked"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := widescreenModel(t)
			mdl, _ := m.Update(consoleStatusMsg{status: c.status})
			mm := mdl.(Model)
			lines := strings.Join(mm.charterReadSurfaceLines(), "\n")
			for _, want := range c.want {
				if !strings.Contains(lines, want) {
					t.Errorf("missing %q in:\n%s", want, lines)
				}
			}
			for _, absent := range c.absent {
				if strings.Contains(lines, absent) {
					t.Errorf("unexpected %q in:\n%s", absent, lines)
				}
			}
		})
	}
}

// TestCharterReadSurfaceUnavailableBeforeStatus: no status peek yet ("" is
// consoleCharter's zero value) renders an honest absence, never an invented
// provenance.
func TestCharterReadSurfaceUnavailableBeforeStatus(t *testing.T) {
	m := widescreenModel(t)
	lines := strings.Join(m.charterReadSurfaceLines(), "\n")
	if !strings.Contains(lines, "unavailable") {
		t.Errorf("expected an honest 'unavailable' line before any status peek, got %q", lines)
	}
}

// --- T011: $EDITOR handoff ---

// TestEditorHandoffUnsetEditorHonestNotice: no $EDITOR set never shells out
// — an honest notice, no crash, no silent no-op (spec.md AS4).
func TestEditorHandoffUnsetEditorHonestNotice(t *testing.T) {
	t.Setenv("EDITOR", "")
	m := widescreenModel(t)
	mdl, cmd := m.startEditorHandoff()
	if cmd != nil {
		t.Error("no $EDITOR set must never dispatch a command (no subprocess)")
	}
	if got := mdl.(Model).consoleNotice; !strings.Contains(got, "EDITOR") {
		t.Errorf("expected an honest notice naming the missing $EDITOR, got %q", got)
	}
}

// TestEditorRoundTripMsgChanged / Unchanged / Error exercise the pure
// hash-comparison logic tea.ExecProcess's callback runs on return —
// isolated from bubbletea's own Program-level exec plumbing (execMsg is
// unexported and only intercepted inside a running Program, so it cannot be
// driven through Model.Update in a unit test; that half is bubbletea's own
// tested responsibility).
func TestEditorRoundTripMsgChanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "charter.md")
	os.WriteFile(path, []byte("before"), 0o644)
	before := hashFile(path)
	os.WriteFile(path, []byte("after"), 0o644) // simulates what $EDITOR would have done
	msg := editorRoundTripMsg(before, path, nil).(editorResultMsg)
	if !msg.changed || msg.err != nil {
		t.Fatalf("got %+v, want changed=true err=nil", msg)
	}
}

func TestEditorRoundTripMsgUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "charter.md")
	os.WriteFile(path, []byte("same"), 0o644)
	before := hashFile(path)
	msg := editorRoundTripMsg(before, path, nil).(editorResultMsg)
	if msg.changed || msg.err != nil {
		t.Fatalf("got %+v, want changed=false err=nil", msg)
	}
}

// TestEditorRoundTripMsgNonzeroExit (edge case: "$EDITOR exits nonzero:
// treat as no-change... plus an honest one-line notice"): an exec error
// always reports as an error, regardless of whether the file also changed.
func TestEditorRoundTripMsgNonzeroExit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "charter.md")
	os.WriteFile(path, []byte("before"), 0o644)
	msg := editorRoundTripMsg("irrelevant", path, errCommand("exit status 1")).(editorResultMsg)
	if msg.err == nil {
		t.Fatal("a nonzero exit must propagate as an error, never silently treated as unchanged-but-ok")
	}
}

type errCommand string

func (e errCommand) Error() string { return string(e) }

// TestEditorHandoffUpdatesNotice: the Update() handler for editorResultMsg
// sets exactly the contract-worded notice for each outcome.
func TestEditorHandoffUpdatesNotice(t *testing.T) {
	cases := []struct {
		name string
		msg  editorResultMsg
		want string
	}{
		{"changed", editorResultMsg{changed: true}, "charter changed — next turn binds it"},
		{"unchanged clears", editorResultMsg{changed: false}, ""},
		{"error", editorResultMsg{err: errCommand("boom")}, "error"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := widescreenModel(t)
			m.consoleNotice = "stale"
			mdl, _ := m.Update(c.msg)
			got := mdl.(Model).consoleNotice
			if c.want == "" {
				if got != "" {
					t.Errorf("consoleNotice = %q, want cleared", got)
				}
				return
			}
			if !strings.Contains(got, c.want) {
				t.Errorf("consoleNotice = %q, want it to contain %q", got, c.want)
			}
		})
	}
}

// TestEditorCommandRealRoundTripViaScript is the "scripted fake-editor
// round-trip test" the tasks call for: a real subprocess (a tiny shell
// script standing in for $EDITOR) actually mutates charter.md, and the
// exact hash-comparison + message-construction path startEditorHandoff
// wires together (editorCommand + hashFile + editorRoundTripMsg) detects
// it — everything in this package's control, short of tea.ExecProcess's own
// Program-level suspend/resume (bubbletea's tested responsibility).
func TestEditorCommandRealRoundTripViaScript(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "charter.md")
	if err := os.WriteFile(path, []byte("original charter"), 0o644); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(dir, "fake-editor.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho 'edited by fake editor' > \"$1\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	before := hashFile(path)
	cmd := editorCommand(script, path)
	err := cmd.Run()
	msg := editorRoundTripMsg(before, path, err).(editorResultMsg)
	if msg.err != nil {
		t.Fatalf("unexpected error: %v", msg.err)
	}
	if !msg.changed {
		t.Fatal("expected the scripted fake editor to have changed the file")
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != "edited by fake editor\n" {
		t.Errorf("file content = %q, want the fake editor's output", data)
	}
}

// TestEditorCommandRealRoundTripNoChange: a no-op fake editor leaves the
// hash — and therefore the notice decision — unchanged.
func TestEditorCommandRealRoundTripNoChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "charter.md")
	if err := os.WriteFile(path, []byte("charter text"), 0o644); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(dir, "fake-editor.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\ntrue\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	before := hashFile(path)
	cmd := editorCommand(script, path)
	err := cmd.Run()
	msg := editorRoundTripMsg(before, path, err).(editorResultMsg)
	if msg.err != nil {
		t.Fatalf("unexpected error: %v", msg.err)
	}
	if msg.changed {
		t.Fatal("a no-op editor must not report a change")
	}
}

// --- exact-height invariant (B1): the console page is a top-level render
// path like widescreenView/narrowView and must hold it too. ---

func TestConsoleViewExactHeight(t *testing.T) {
	m := widescreenModel(t) // 140x40
	m.connected = true
	var mdl tea.Model = m
	mdl = update(mdl, "G")
	v := mdl.(Model).View()
	lines := strings.Split(v, "\n")
	if len(lines) != mdl.(Model).height {
		t.Errorf("console View() = %d lines, want %d:\n%s", len(lines), mdl.(Model).height, v)
	}
}
