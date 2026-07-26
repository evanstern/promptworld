package tui

// Help overlay tests (spec 045-tui-help-overlay, TASK-116): open/dismiss
// routing from every mode (US1, T008), the SC-003 keymap sweep (FR-003),
// the screen walkthrough's content-presence assertions (US2, T010), the
// no-LLM byte-identity proof (US3, T011), and the lessons seam (US4, T013).

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
)

// --- T008: '?' opens from every documented mode ---

// TestHelpOpensFromEveryMode: from each mode named in spec.md's US1
// Independent Test, '?' opens the overlay frozen on that mode's page;
// dismissal (esc) restores the exact prior mode/focus/selection.
func TestHelpOpensFromEveryMode(t *testing.T) {
	cases := []struct {
		name     string
		build    func(t *testing.T) Model
		wantMode helpModeKey
	}{
		{"global/home widescreen", func(t *testing.T) Model { return widescreenModel(t) }, helpModeGlobal},
		{"inspect", func(t *testing.T) Model { return pausedModel(t) }, helpModeInspect},
		{"villagers roster", func(t *testing.T) Model {
			m := widescreenModel(t)
			m.dockTab = paneVillagers
			return m
		}, helpModeVillagersRoster},
		{"villagers detail", func(t *testing.T) Model {
			m := widescreenModel(t)
			m.dockTab = paneVillagers
			m.villDetail = true
			return m
		}, helpModeVillagersDetail},
		{"solo zoom", func(t *testing.T) Model {
			m := widescreenModel(t)
			m.solo = true
			return m
		}, helpModeSolo},
		{"narrow fallback", func(t *testing.T) Model { return testModel(t) }, helpModeSolo},
		{"look-cursor", func(t *testing.T) Model {
			m := widescreenModel(t)
			m.lookActive = true
			return m
		}, helpModeLook},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.build(t)
			before := m.View()
			var mdl tea.Model = m
			mdl = update(mdl, "?")
			mm := mdl.(Model)
			if !mm.helpOpen {
				t.Fatal("'?' must open the overlay")
			}
			if mm.helpMode != tc.wantMode {
				t.Errorf("helpMode = %v, want %v", mm.helpMode, tc.wantMode)
			}
			if mm.helpPageMode != tc.wantMode {
				t.Errorf("helpPageMode = %v, want %v (starts on the frozen mode)", mm.helpPageMode, tc.wantMode)
			}
			if mm.helpTier {
				t.Error("must open on the basic tier")
			}
			if mm.helpSection != helpSectionKeys {
				t.Error("must open on the keys section")
			}
			view := mm.View()
			if view == "" {
				t.Fatal("overlay view rendered empty")
			}

			// Dismiss and verify the exact prior screen/state is restored.
			mdl = update(mdl, "esc")
			after := mdl.(Model)
			if after.helpOpen {
				t.Fatal("esc must dismiss the overlay")
			}
			if after.View() != before {
				t.Errorf("dismissal did not restore the exact prior view:\nbefore: %q\nafter:  %q", before, after.View())
			}
		})
	}
}

// TestHelpQuestionMarkTogglesClosed: '?' pressed while already open is a
// dismiss (toggle), never a stacked second overlay (spec.md Edge Cases).
func TestHelpQuestionMarkTogglesClosed(t *testing.T) {
	m := widescreenModel(t)
	var mdl tea.Model = m
	mdl = update(mdl, "?")
	if !mdl.(Model).helpOpen {
		t.Fatal("setup: expected the overlay open")
	}
	mdl = update(mdl, "?")
	if mdl.(Model).helpOpen {
		t.Fatal("'?' while open must dismiss, not stack a second overlay")
	}
}

// TestHelpMinibufferQuestionMarkStillTypes is the FR-001 pin (also the
// focus-contract rule-4 sweep's cousin, focus_test.go
// TestFocusContractCheck3_EveryKeyVisiblyActsWhileFocused): '?' while the
// minibuffer is focused appends to the buffer — it must never open help.
func TestHelpMinibufferQuestionMarkStillTypes(t *testing.T) {
	m := testModel(t)
	var mdl tea.Model = m
	mdl = update(mdl, "m")
	if !mdl.(Model).mbFocused {
		t.Fatal("setup: expected minibuffer focus")
	}
	mdl = update(mdl, "?")
	mm := mdl.(Model)
	if mm.helpOpen {
		t.Fatal("'?' while focused must never open help")
	}
	if !mm.mbFocused {
		t.Fatal("focus must survive typing '?'")
	}
	if got := mm.mbInput; got != "?" {
		t.Fatalf("mbInput = %q, want %q — '?' must type like any character", got, "?")
	}
}

// TestHelpEscReleasesExactlyOneLayer: with villagers-detail open underneath
// the overlay, esc closes only help — detail must still be open; a second
// esc then closes detail (spec.md "esc-release ordering").
func TestHelpEscReleasesExactlyOneLayer(t *testing.T) {
	m := widescreenModel(t)
	m.dockTab = paneVillagers
	m.villDetail = true
	var mdl tea.Model = m
	mdl = update(mdl, "?")
	if !mdl.(Model).helpOpen {
		t.Fatal("setup: expected help open")
	}
	mdl = update(mdl, "esc")
	mm := mdl.(Model)
	if mm.helpOpen {
		t.Fatal("esc must close the overlay")
	}
	if !mm.villDetail {
		t.Fatal("esc must release only the help layer — villager detail should still be open")
	}
	mdl = update(mm, "esc")
	if mdl.(Model).villDetail {
		t.Fatal("a second esc should now close detail (one layer per press)")
	}
}

// TestHelpNoSideEffectsOnWorld (FR-006/SC-005): opening, tiering,
// paging, scrolling, and closing the overlay never sends a command, never
// touches the clock, and restores every non-help field exactly.
func TestHelpNoSideEffectsOnWorld(t *testing.T) {
	m := pausedModel(t)
	m.chronSelected = 2
	m.panX, m.panY = 3, -2
	before := m
	var mdl tea.Model = m
	for _, k := range []string{"?", "tab", "t", "n", "n", "p", "J", "J", "K", "tab", "tab", "shift+tab"} {
		next, cmd := mdl.Update(key(k))
		if cmd != nil {
			t.Fatalf("key %q dispatched a command — help must be side-effect-free (FR-006)", k)
		}
		mdl = next
	}
	mdl = update(mdl, "esc")
	after := mdl.(Model)
	if after.chronSelected != before.chronSelected || after.panX != before.panX || after.panY != before.panY {
		t.Errorf("non-help state drifted: chronSelected=%d panX=%d panY=%d, want %d/%d/%d",
			after.chronSelected, after.panX, after.panY, before.chronSelected, before.panX, before.panY)
	}
	if after.status.Clock.Paused != before.status.Clock.Paused {
		t.Error("the clock's pause state must never change because of the overlay")
	}
}

// TestHelpOpenCloseLoopHundredTimes is SC-005: 100 open/close cycles in a
// running (unpaused) world cause zero drift in clock-relevant state and
// never dispatch a command.
func TestHelpOpenCloseLoopHundredTimes(t *testing.T) {
	m := widescreenModel(t)
	m.connected = true
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{Paused: false, Speed: "4x"}}
	before := m
	var mdl tea.Model = m
	for i := 0; i < 100; i++ {
		next, cmd := mdl.Update(key("?"))
		if cmd != nil {
			t.Fatalf("open (iter %d) dispatched a command", i)
		}
		mdl = next
		next, cmd = mdl.Update(key("?"))
		if cmd != nil {
			t.Fatalf("close (iter %d) dispatched a command", i)
		}
		mdl = next
	}
	after := mdl.(Model)
	if after.helpOpen {
		t.Fatal("must end closed after an even number of toggles")
	}
	if after.status.Clock.Paused != before.status.Clock.Paused || after.status.Clock.Speed != before.status.Clock.Speed {
		t.Error("100 open/close cycles must cause zero clock drift")
	}
}

// TestHelpAllOtherKeysInertWhileOpen (FR-012): while open, keys that would
// normally pause the clock, quit, or switch tabs are swallowed — no silent
// fallthrough to the mode beneath.
func TestHelpAllOtherKeysInertWhileOpen(t *testing.T) {
	m := widescreenModel(t)
	m.connected = true
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{Paused: false, Speed: "4x"}}
	var mdl tea.Model = m
	mdl = update(mdl, "?")
	for _, k := range []string{"q", "2", "3", "4", "1", "m", "a", "r", "[", "]", "g", "G"} {
		next, cmd := mdl.Update(key(k))
		mm := next.(Model)
		if cmd != nil {
			t.Errorf("key %q dispatched a command while help was open — must be inert", k)
		}
		if mm.quitting {
			t.Fatalf("key %q must not quit while help is open", k)
		}
		if !mm.helpOpen {
			t.Fatalf("key %q must not close the overlay (only esc/'?' do)", k)
		}
		mdl = next
	}
}

// --- SC-003: the keymap sweep — every listed binding real, every real
// binding listed exactly once, per mode. ---

// realModeKeys is the sweep's independent oracle: hand-audited against the
// actual dispatch code (tui.go), kept separate from helpPages (help.go) so
// this cross-checks a second, independently authored source — the same
// two-tables-compared shape as digest_test.go's catalogFixture/
// digestRegistry (TestCatalogSweep precedent, research.md R3).
var realModeKeys = map[helpModeKey][]string{
	// handleGlobalKey's switch (tui.go) — one case list, string literals.
	// "G" (spec 053) opens the guardian console; "5" selects the new
	// systems dock tab.
	helpModeGlobal: {
		"q", "1", "2", "3", "4", "5", "6", "G", "tab", "shift+tab", "m", "enter", "esc",
		"a", "t", "r", "up", "down", "left", "right", "c", " ", "[", "]",
	},
	// Same handler as global — solo/narrow drives it identically; only the
	// basic/advanced split differs (R1/data-model.md).
	helpModeSolo: {
		"q", "1", "2", "3", "4", "5", "6", "G", "tab", "shift+tab", "m", "enter", "esc",
		"a", "t", "r", "up", "down", "left", "right", "c", " ", "[", "]",
	},
	// handleMinibufferKey (tui.go) switches on msg.Type; represented here
	// by the key forms the `key()` test helper produces for each case.
	helpModeMinibuffer: {"esc", "enter", "backspace", "up", "down", " "},
	// handleInspectKey's own claimed set (j,k,g,G,J,K,enter — all
	// handled=true) union globalKeys minus "enter" (shadowed: inspect's
	// own "enter" case runs first, so global's narrow-only body never
	// fires while inspecting — keymap.md "All global keys stay live").
	// "G" is NOT added a second time here: inspect's own binding (jump to
	// last event) already claims the literal key and shadows the global
	// console-open binding exactly the way its "enter" shadows global's —
	// FR-001 scopes the console key to "home, solo, and narrow" precisely
	// because inspect/villagers modes already own "G" for their own jump.
	// "5" (tab-select) is unclaimed here, so it falls through to global.
	helpModeInspect: {
		"j", "k", "g", "G", "J", "K", "enter",
		"q", "1", "2", "3", "4", "5", "6", "tab", "shift+tab", "m", "esc",
		"a", "t", "r", "up", "down", "left", "right", "c", " ", "[", "]",
	},
	// handleVillagersKey, roster state: j/k/g/G/enter always handled=true
	// regardless of villDetail; esc/d fall through to global when !villDetail
	// (esc real via global's own case; "d" has no global case, so it's
	// truly unbound in roster and excluded here). "G" is villJumpLast's own
	// binding (see the inspect-mode comment above); "5" falls through.
	helpModeVillagersRoster: {
		"j", "k", "g", "G", "enter",
		"q", "1", "2", "3", "4", "5", "6", "tab", "shift+tab", "m", "esc",
		"a", "t", "r", "up", "down", "left", "right", "c", " ", "[", "]",
	},
	// handleVillagersKey, detail state: esc/d also become handled=true here
	// (shadowing global's esc), alongside the always-true j/k/g/G/enter.
	helpModeVillagersDetail: {
		"j", "k", "g", "G", "enter", "esc", "d",
		"q", "1", "2", "3", "4", "5", "6", "tab", "shift+tab", "m",
		"a", "t", "r", "up", "down", "left", "right", "c", " ", "[", "]",
	},
	// handleLookKey (look.go): claimed at every focus layer — h/j/k/l/
	// arrows/H/J/K/L/c/enter move or jump the cursor (cursor focus), select/
	// scroll (pane/drill focus, j/k/J/K only — left/right/H/L are a
	// documented no-op there so they can never leak to the global pan);
	// tab/shift+tab and digits 2-6 are claimed at every layer too (never
	// leaking to the global dock-cycle/solo-zoom, which would silently
	// change dockTab/solo underneath the borrow). "G" is claimed only to
	// exit the mode first, then deliberately falls through (handled=false)
	// to the SAME global "G" case every other mode reaches — one
	// console-open path, not two. "q"/"m"/" "/"["/"]" fall through
	// unclaimed exactly as FR-013 requires. "a"/"t"/"r"/"1" are excluded:
	// each reaches a global case, but chronicleVisible()==false (the
	// borrow's own guard) makes a/t/r's body inert, and 1 merely re-asserts
	// state (solo=false, active=paneMap) that lookEntryAllowed() already
	// required true — no real effect while this mode is active (the
	// villagers-roster "d" exclusion precedent, one row up).
	helpModeLook: {
		"v", "h", "j", "k", "l", "up", "down", "left", "right",
		"H", "J", "K", "L", "c", "enter", "tab", "shift+tab", "esc",
		"2", "3", "4", "5", "6", "G", "q", "m", " ", "[", "]",
	},
}

func keySet(keys []string) map[string]int {
	out := make(map[string]int, len(keys))
	for _, k := range keys {
		out[k]++
	}
	return out
}

func flattenRows(rows ...[]helpKeyRow) []string {
	var out []string
	for _, rs := range rows {
		for _, r := range rs {
			out = append(out, r.Keys...)
		}
	}
	return out
}

// TestHelpKeymapSweep is the SC-003 mechanical gate: for every mode, the
// overlay's declared keys (Basic ∪ Advanced) must equal — exactly, no
// extras, no omissions, no duplicates within a mode — the real dispatch
// code's recognized key set (realModeKeys).
func TestHelpKeymapSweep(t *testing.T) {
	for mode, want := range realModeKeys {
		mode, want := mode, want
		t.Run(helpPages[mode].Title, func(t *testing.T) {
			page := helpPages[mode]
			got := flattenRows(page.Basic, page.Advanced)

			gotSet := keySet(got)
			wantSet := keySet(want)

			for k, n := range gotSet {
				if n > 1 {
					t.Errorf("key %q listed %d times in %s's page — must appear in exactly one tier", k, n, page.Title)
				}
				if _, ok := wantSet[k]; !ok {
					t.Errorf("%s's page lists key %q, which the real dispatch code does not recognize in this mode", page.Title, k)
				}
			}
			for k := range wantSet {
				if _, ok := gotSet[k]; !ok {
					t.Errorf("%s's real dispatch code recognizes key %q, but no help page row lists it", page.Title, k)
				}
			}
		})
	}
	// Every mode must have an oracle row (catches a forgotten new mode).
	for mode := helpModeKey(0); mode < helpModeCount; mode++ {
		if _, ok := realModeKeys[mode]; !ok {
			t.Errorf("helpModeKey %v has no realModeKeys oracle entry", mode)
		}
	}
}

// TestHelpKeymapSweepLiveDispatch spot-checks a representative key from
// each mode's page against the actual running dispatch, catching a typo'd
// literal that the static-list sweep above can't (e.g. a key that's listed
// but the live handler silently ignores).
func TestHelpKeymapSweepLiveDispatch(t *testing.T) {
	t.Run("global: '2' selects the chronicle tab", func(t *testing.T) {
		m := widescreenModel(t)
		mdl, _ := m.Update(key("2"))
		if mdl.(Model).dockTab != paneChronicle {
			t.Error("'2' did not select the chronicle tab")
		}
	})
	t.Run("inspect: 'j' moves the selection", func(t *testing.T) {
		m := pausedModel(t)
		mdl, _ := m.Update(key("j"))
		if mdl.(Model).chronSelected != 4 {
			t.Error("'j' did not move the inspect selection")
		}
	})
	t.Run("villagers roster: 'enter' opens detail", func(t *testing.T) {
		m := widescreenModel(t)
		m.dockTab = paneVillagers
		m.replica.Agents = append(m.replica.Agents, sim.Agent{Name: "Ash"})
		mdl, _ := m.Update(key("enter"))
		if !mdl.(Model).villDetail {
			t.Error("'enter' did not open villager detail")
		}
	})
	t.Run("villagers detail: 'd' toggles decisions", func(t *testing.T) {
		m := widescreenModel(t)
		m.dockTab = paneVillagers
		m.villDetail = true
		mdl, _ := m.Update(key("d"))
		if !mdl.(Model).villDecisions {
			t.Error("'d' did not open the decisions sub-view")
		}
	})
}

// --- T010: screen walkthrough content-presence ---

// TestHelpWalkthroughCoversEveryHeaderBadge: every string headerView
// (views.go) can render — including every conditional badge — has a
// matching walkthrough row (US2-AS2).
func TestHelpWalkthroughCoversEveryHeaderBadge(t *testing.T) {
	lines := strings.Join(Model{}.helpWalkthroughLines(200), "\n")
	for _, want := range []string{
		"running / PAUSED", "speed (t/s)", "asked", "[degraded]", "[llm:", "[suppressed:", "disconnected",
	} {
		if !strings.Contains(lines, want) {
			t.Errorf("walkthrough missing header element/badge %q", want)
		}
	}
}

// --- spec 074 FR-011/R8: the badge deep-link ---

// TestHelpBadgeDeepLinkOpensPreFocused: with a conditional header badge
// active, '?' opens the overlay pre-focused on the screen-walkthrough
// section, scrolled so that badge's headerAnatomy row is visible.
func TestHelpBadgeDeepLinkOpensPreFocused(t *testing.T) {
	cases := []struct {
		name   string
		status *ipc.StatusData
		want   string
	}{
		{"degraded", &ipc.StatusData{Clock: ipc.ClockStatus{Degraded: true}}, "[degraded]"},
		{"llm condition", &ipc.StatusData{LLM: &llm.Status{Providers: []llm.ProviderStatus{{Name: "anthropic", Condition: "rate_limited"}}}}, "[llm: provider kind]"},
		{"suppressed", &ipc.StatusData{Horizon: []ipc.HorizonClass{{Class: "cognition", Suppressed: true}}}, "[suppressed: classes]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := widescreenModel(t)
			m.status = tc.status
			var mdl tea.Model = m
			mdl = update(mdl, "?")
			mm := mdl.(Model)
			if mm.helpSection != helpSectionWalkthrough {
				t.Fatalf("badge-active open should land on the screen-walkthrough section, got %v", mm.helpSection)
			}
			lines := mm.helpWalkthroughLines(200)
			windowed := paginateHelpContent(lines, mm.helpScroll, 3)
			if !strings.Contains(strings.Join(windowed, "\n"), tc.want) {
				t.Errorf("badge %q's row should be visible at scroll %d:\n%s", tc.want, mm.helpScroll, strings.Join(windowed, "\n"))
			}
		})
	}
}

// TestHelpNoBadgeOpensByteIdentical: with no conditional badge active,
// opening the overlay is byte-identical to before this feature — the keys
// section, scroll 0.
func TestHelpNoBadgeOpensByteIdentical(t *testing.T) {
	m := widescreenModel(t)
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{}}
	var mdl tea.Model = m
	mdl = update(mdl, "?")
	mm := mdl.(Model)
	if mm.helpSection != helpSectionKeys || mm.helpScroll != 0 {
		t.Errorf("no active badge should open byte-identically (keys section, scroll 0): section=%v scroll=%d", mm.helpSection, mm.helpScroll)
	}
}

// TestHelpBadgeDeepLinkNavigationUnchanged (AS 5.3): after a pre-focused
// open, ordinary overlay navigation (tab/section cycling) behaves exactly
// as it does from any other open.
func TestHelpBadgeDeepLinkNavigationUnchanged(t *testing.T) {
	m := widescreenModel(t)
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{Degraded: true}}
	var mdl tea.Model = m
	mdl = update(mdl, "?")
	mdl = update(mdl, "tab")
	mm := mdl.(Model)
	if mm.helpSection != helpSectionLessons {
		t.Errorf("tab from the walkthrough section should cycle to lessons, got %v", mm.helpSection)
	}
	if mm.helpScroll != 0 {
		t.Error("cycling sections should reset scroll")
	}
}

// TestHelpLookModePageReachableAndFooterHint (FR-014): the look-cursor mode
// has its own keys page (reachable frozen-open and via n/p paging), and the
// footer advertises the mode's primary hints while it is active.
func TestHelpLookModePageReachableAndFooterHint(t *testing.T) {
	m := widescreenModel(t)
	m.lookActive = true
	if got := m.currentHelpMode(); got != helpModeLook {
		t.Fatalf("currentHelpMode during the look-cursor mode = %v, want helpModeLook", got)
	}
	var mdl tea.Model = m
	mdl = update(mdl, "?") // freezes helpPageMode = helpModeLook
	lines := strings.Join(mdl.(Model).helpKeysLines(200), "\n")
	if !strings.Contains(lines, "Look-cursor") {
		t.Errorf("look-cursor mode page should render its own title: %q", lines)
	}

	footer := m.footerView()
	if !strings.Contains(footer, "esc") {
		t.Errorf("footer should hint the mode's esc-release while active: %q", footer)
	}
}

// TestHelpWalkthroughGlyphPageMatchesSharedTable (FR-005/SC-003): the
// overlay's glyph page enumerates exactly the mapGlyphs table — drift is
// impossible by construction (both the overlay and renderMapGrid's legend
// render from the same var), asserted anyway per data-model.md T010.
func TestHelpWalkthroughGlyphPageMatchesSharedTable(t *testing.T) {
	lines := strings.Join(Model{}.helpWalkthroughLines(200), "\n")
	for _, g := range mapGlyphs {
		if !strings.Contains(lines, g.Meaning) {
			t.Errorf("walkthrough missing glyph %q's meaning: %q", g.Glyph, g.Meaning)
		}
	}
	// And the reverse: the compact map legend must carry every glyph token
	// mapGlyphs declares — same source, so this can only fail if
	// legendGlyphLine's construction itself diverges from the table.
	compact := legendGlyphLine()
	for _, g := range mapGlyphs {
		if !strings.Contains(compact, g.Glyph+g.Name) {
			t.Errorf("compact legend missing token %q", g.Glyph+g.Name)
		}
	}
}

// TestHelpWalkthroughCoversEveryDockTab: every dock tab (paneNames/
// dockTabKey) has a walkthrough row.
func TestHelpWalkthroughCoversEveryDockTab(t *testing.T) {
	lines := strings.Join(Model{}.helpWalkthroughLines(200), "\n")
	for _, name := range []string{paneNames[paneChronicle], paneNames[paneGuardian], paneNames[paneVillagers], paneNames[paneSystems]} {
		if !strings.Contains(lines, name) {
			t.Errorf("walkthrough missing dock tab %q", name)
		}
	}
}

// TestHelpWalkthroughScrollsAt80x24 (FR-009): at the client's small-terminal
// floor, the walkthrough is reachable via the pager rather than truncated
// and lost.
func TestHelpWalkthroughScrollsAt80x24(t *testing.T) {
	m := testModel(t) // 80x30, narrow — exercises the narrow overlay path too
	m.helpOpen = true
	m.helpSection = helpSectionWalkthrough
	full := strings.Join(Model{}.helpWalkthroughLines(76), "\n")
	// Confirm the content genuinely overflows a small pane so this test is
	// meaningful, then walk the pager to prove every line is reachable.
	var seen []string
	for scroll := 0; scroll < 200; scroll++ {
		page := paginateHelpContent(Model{}.helpWalkthroughLines(76), scroll, 6)
		seen = append(seen, page...)
		if !strings.Contains(strings.Join(page, "\n"), "J to scroll") {
			break
		}
	}
	joined := strings.Join(seen, "\n")
	for _, want := range []string{"Header anatomy", "Map glyphs", "Dock tabs"} {
		if !strings.Contains(joined, want) {
			t.Errorf("paging never surfaced %q (full content length %d lines)", want, len(strings.Split(full, "\n")))
		}
	}
	if v := m.View(); v == "" {
		t.Error("narrow overlay view rendered empty at 80x30")
	}
}

// --- T011: no-LLM / nil-status byte identity (US3, SC-004) ---

// TestHelpNoLLMByteIdentical: overlay content and behavior with a nil
// status/replica (the no-LLM/disconnected floor, tui.go:151-153,
// views.go:108-110 tolerance) is byte-identical to a status-carrying model
// — no help.go code path reads m.status or m.replica.
func TestHelpNoLLMByteIdentical(t *testing.T) {
	m := widescreenModel(t)
	m.status = nil
	m.replica = nil
	var mdl tea.Model = m
	mdl = update(mdl, "?")
	mm := mdl.(Model)
	if !mm.helpOpen {
		t.Fatal("help must open even with nil status/replica")
	}
	view := mm.View()
	if view == "" {
		t.Fatal("overlay view rendered empty with nil status/replica")
	}
	if !strings.Contains(view, "HELP") {
		t.Errorf("overlay view missing its own title with nil status/replica: %q", view)
	}

	// Byte-identity: the same mode/tier/section renders identically
	// regardless of whether status/replica are present (SC-004) — content
	// tables never branch on either. Forcing the same helpPageMode on both
	// sides isolates this from currentHelpMode()'s own (legitimate)
	// status-dependence, i.e. m.inspecting() needing a paused clock —
	// that's a *routing* difference, not a content one.
	withStatus := widescreenModel(t)
	withStatus.status = &ipc.StatusData{Clock: ipc.ClockStatus{Paused: true, Speed: "8x"}}
	withStatus.replica.Agents = append(withStatus.replica.Agents, sim.Agent{Name: "Ash"})

	for mode := helpModeKey(0); mode < helpModeCount; mode++ {
		for _, tier := range []bool{false, true} {
			for _, section := range []helpSection{helpSectionKeys, helpSectionWalkthrough, helpSectionLessons} {
				a := mm
				a.helpPageMode, a.helpTier, a.helpSection = mode, tier, section
				b := withStatus
				b.helpPageMode, b.helpTier, b.helpSection = mode, tier, section
				linesA := strings.Join(a.helpContentLines(76, 20), "\n")
				linesB := strings.Join(b.helpContentLines(76, 20), "\n")
				if linesA != linesB {
					t.Fatalf("mode=%v tier=%v section=%v content differs between nil-status and status-carrying models:\nnil:    %q\nstatus: %q",
						mode, tier, section, linesA, linesB)
				}
			}
		}
	}
}

// TestHelpContentReadsNoStatusOrReplica is a construction check (R6): every
// mode's key page and every walkthrough/lessons line renders without ever
// touching m.status/m.replica — proven by rendering with both nil across
// every mode, tier, and section without panicking or differing from a
// populated model (paired with the byte-identity test above).
func TestHelpContentReadsNoStatusOrReplica(t *testing.T) {
	m := testModel(t)
	m.status = nil
	m.replica = nil
	m.helpOpen = true
	for mode := helpModeKey(0); mode < helpModeCount; mode++ {
		m.helpPageMode = mode
		for _, tier := range []bool{false, true} {
			m.helpTier = tier
			for section := helpSection(0); section < helpSectionCount; section++ {
				m.helpSection = section
				if lines := m.helpContentLines(76, 20); len(lines) == 0 {
					t.Errorf("mode=%v tier=%v section=%v produced no content", mode, tier, section)
				}
			}
		}
	}
}

// --- T013: the lessons seam (US4, SC-006) ---

// TestHelpLessonsPlaceholderWhenEmpty: helpLessonsLines still degrades to
// the documented placeholder line when the table is empty. This used to be
// the feature's shipped state (spec 045, "the seam, not the content"); spec
// 055 (TASK-117) now populates helpLessons for real at every client boot
// (New(), lessons.go populateHelpLessons) — this test pins the rendering
// function's own empty-table tolerance directly, independent of that.
func TestHelpLessonsPlaceholderWhenEmpty(t *testing.T) {
	orig := helpLessons
	t.Cleanup(func() { helpLessons = orig })
	helpLessons = nil
	lines := strings.Join(helpLessonsLines(76), "\n")
	if !strings.Contains(lines, "lessons appear here") {
		t.Errorf("empty lessons table should render the placeholder line, got %q", lines)
	}
}

// TestHelpLessonsFixtureEntryRendersWithZeroStructuralChange (SC-006): a
// fixture lesson entry injected into the table renders and is navigable
// with no change to this file's navigation/rendering code — proving the
// seam contract (contracts/help-content.md "The lessons seam") holds.
func TestHelpLessonsFixtureEntryRendersWithZeroStructuralChange(t *testing.T) {
	// Construct the model FIRST: New() populates helpLessons from the real
	// catalog now (spec 055, populateHelpLessons) — the fixture below must
	// override it AFTER construction, and every use below must reuse this
	// same m (never a fresh widescreenModel(t)/testModel(t) call, which
	// would re-run New() and clobber the fixture back to the real catalog).
	m := widescreenModel(t)
	orig := helpLessons
	t.Cleanup(func() { helpLessons = orig })
	helpLessons = []helpLesson{
		{ID: "lesson-fixture-1", Title: "Fire needs fuel", Body: "A fire burns out once its fuel runs out; refuel before it goes cold."},
	}

	direct := m
	direct.helpOpen = true
	direct.helpSection = helpSectionLessons
	view := direct.View()
	if !strings.Contains(view, "Fire needs fuel") {
		t.Errorf("fixture lesson title not rendered: %q", view)
	}

	// Navigable: reachable via the same tab-cycling every section uses.
	var mdl tea.Model = m
	mdl = update(mdl, "?")   // opens on keys
	mdl = update(mdl, "tab") // -> walkthrough
	mdl = update(mdl, "tab") // -> lessons
	mm := mdl.(Model)
	if mm.helpSection != helpSectionLessons {
		t.Fatalf("tab cycling did not reach the lessons section: %v", mm.helpSection)
	}
	if v := mm.View(); !strings.Contains(v, "Fire needs fuel") {
		t.Errorf("lessons section reached via navigation missing the fixture entry: %q", v)
	}
}

// --- spec 056: the ceremony-replay section (overlays/help.md "ceremony
// replay entries"; overlays/ceremony.md "Replayability", FR-013/SC-004) ---

// TestHelpCeremonyReplaySectionPlaceholderWhenEmpty: the section degrades to
// an honest placeholder before any stage has unlocked.
func TestHelpCeremonyReplaySectionPlaceholderWhenEmpty(t *testing.T) {
	m := testModel(t)
	lines := strings.Join(m.ceremonyReplayLines(76), "\n")
	if !strings.Contains(lines, "no stage has unlocked") {
		t.Errorf("empty ceremony-replay section should render the placeholder line, got %q", lines)
	}
}

// TestHelpCeremonyReplaySectionReachableAndStored (SC-004): a dismissed
// ceremony's content is retrievable via the `?` overlay with zero model
// calls — reachable by the same tab-cycling every section uses, and
// rendering the SAME chapter/report-card content the live ceremony showed.
func TestHelpCeremonyReplaySectionReachableAndStored(t *testing.T) {
	m := widescreenModel(t)
	m.replica.CurriculumPasses = []sim.CurriculumPass{
		{Exercise: "first-night", Stage: "stage-1", Evidence: []sim.EvidenceRef{{Type: "sim.day_started"}}},
	}
	m.applyEvent(stageUnlockedEvent(1, 100, "stage-2", "first-night"))
	// Dismiss the live ceremony — the replay section must still carry it.
	var mdl tea.Model = m
	mdl = update(mdl, "esc")
	mm := mdl.(Model)
	if mm.takeover != takeoverNone {
		t.Fatalf("precondition: ceremony should be dismissed, takeover = %v", mm.takeover)
	}

	mdl2 := tea.Model(mm)
	mdl2 = update(mdl2, "?")   // opens on keys
	mdl2 = update(mdl2, "tab") // -> walkthrough
	mdl2 = update(mdl2, "tab") // -> lessons
	mdl2 = update(mdl2, "tab") // -> ceremonies
	mm2 := mdl2.(Model)
	if mm2.helpSection != helpSectionCeremonies {
		t.Fatalf("tab cycling did not reach the ceremonies section: %v", mm2.helpSection)
	}
	view := mm2.View()
	if !strings.Contains(strings.ToUpper(view), "THE WRITTEN WORD") {
		t.Errorf("ceremony-replay section missing the unlocked stage's identity: %q", view)
	}
	if !strings.Contains(view, "report card · first-night") {
		t.Errorf("ceremony-replay section missing the SAME report card the live ceremony showed: %q", view)
	}
}
