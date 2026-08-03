package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worlds"
)

// seedOperatorRecords writes a fully-populated per-user home: every lesson in
// the catalog already seen, and every stage already unlocked. This is the
// "veteran operator" machine — the one whose frames would differ from a fresh
// operator's if the harness read these records (plan.md F1).
func seedOperatorRecords(t *testing.T) string {
	t.Helper()
	home := filepath.Join(t.TempDir(), "home")
	t.Setenv("PROMPTWORLD_HOME", home)
	for i := range lessonCatalog {
		worlds.MarkLessonSeen(lessonCatalog[i].ID, "some-other-world")
	}
	for _, stage := range world.StageOrder {
		worlds.UpsertUnlock(stage, worlds.UnlockEntry{
			World: "some-other-world", Path: "/elsewhere", Exercise: "first-night", EarnedAt: "2026-01-01T00:00:00Z",
		})
	}
	if seen := worlds.LoadLessonsSeen(); len(seen.Entries) != len(lessonCatalog) {
		t.Fatalf("record setup: %d lessons recorded seen, want %d", len(seen.Entries), len(lessonCatalog))
	}
	return home
}

// emptyOperatorRecords is the fresh machine: a home directory with nothing in
// it at all.
func emptyOperatorRecords(t *testing.T) string {
	t.Helper()
	home := filepath.Join(t.TempDir(), "home")
	t.Setenv("PROMPTWORLD_HOME", home)
	return home
}

// TestFrameIgnoresPerUserRecordsOnDisk is the determinism fix's proof
// (plan.md F1, spec.md FR-003 / AC #1 and #3). tui.New() loads
// worlds.LoadLessonsSeen() and worlds.LoadUnlocks() from the operator's home
// directory; left alone, an identical fixture renders differently for a
// veteran operator than for a fresh one, and the "byte-identical matrix"
// guarantee fails silently on someone else's machine.
//
// Two halves, and the second is what makes the first mean anything:
//
//	(a) every fixture, in every state, renders byte-identically against a
//	    fully-populated home and against an empty one;
//	(b) the SAME fixture built the LIVE way — through the on-disk records —
//	    does differ, so (a) is a real seam doing real work rather than a
//	    vacuous comparison of two things that never depended on disk.
func TestFrameIgnoresPerUserRecordsOnDisk(t *testing.T) {
	states := []string{"home", "help-walkthrough", "help-lessons"}

	render := func(t *testing.T, f Fixture, state string, pu perUserState) string {
		t.Helper()
		m, err := f.buildWith(pu)
		if err != nil {
			t.Fatal(err)
		}
		m.width, m.height = 140, 40
		if err := poseState(&m, state); err != nil {
			t.Fatal(err)
		}
		restore := forceColorProfile(false)
		defer restore()
		return m.View()
	}

	for _, f := range Fixtures() {
		for _, state := range states {
			t.Run(f.ID+"/"+state, func(t *testing.T) {
				// (a) the harness path: canned records, both homes.
				seedOperatorRecords(t)
				veteran := render(t, f, state, fixturePerUserState())
				emptyOperatorRecords(t)
				fresh := render(t, f, state, fixturePerUserState())
				if veteran != fresh {
					t.Errorf("fixture frame depends on the operator's per-user records:\nveteran home:\n%s\nfresh home:\n%s",
						veteran, fresh)
				}

				// (b) the control: the live path, same two homes.
				seedOperatorRecords(t)
				liveVeteran := render(t, f, state, livePerUserState())
				emptyOperatorRecords(t)
				liveFresh := render(t, f, state, livePerUserState())
				if liveVeteran != liveFresh {
					return // the records genuinely reach this frame — (a) proved something
				}
				t.Logf("note: %s/%s renders identically through the live path too — "+
					"this combination does not exercise the seam", f.ID, state)
			})
		}
	}

	// The control must bite SOMEWHERE, or the whole test is vacuous. The
	// lesson row is the sharpest case: a veteran has already seen every
	// lesson, so the live path shows no row where the canned path always does.
	t.Run("control bites", func(t *testing.T) {
		f, ok := fixtureByID(FixtureMidGame)
		if !ok {
			t.Fatal("mid-game fixture missing")
		}
		seedOperatorRecords(t)
		veteran := render(t, f, "home", livePerUserState())
		emptyOperatorRecords(t)
		fresh := render(t, f, "home", livePerUserState())
		if veteran == fresh {
			t.Fatal("the live path renders identically against a populated and an empty home — " +
				"the per-user records no longer reach the frame, so this test proves nothing")
		}
		if strings.Contains(veteran, lessonPullSuffix) {
			t.Error("a veteran operator (every lesson seen) should see no lesson row through the live path")
		}
		if !strings.Contains(fresh, lessonPullSuffix) {
			t.Error("a fresh operator should see the lesson row through the live path")
		}
	})
}

// TestFixtureRenderWritesNothingToHome is the write half of the same seam:
// applyEvent persists a surfaced lesson (worlds.MarkLessonSeen) on the live
// path, so a fixture that used the real writer would mutate the operator's
// record just by being LOOKED at — and would then render differently the
// second time.
func TestFixtureRenderWritesNothingToHome(t *testing.T) {
	home := emptyOperatorRecords(t)
	for _, f := range Fixtures() {
		if _, err := Frame(FrameOptions{Fixture: f.ID, State: "home", Width: 140, Height: 40}); err != nil {
			t.Fatal(err)
		}
	}
	if entries, err := os.ReadDir(home); err == nil && len(entries) > 0 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("rendering fixtures wrote to the operator's home directory: %v", names)
	}
}

// TestFixtureClockIsFrozen (plan.md F4): every fixture Model reads the frozen
// instant, never the wall clock — so the lessons projection's spacing and
// decay arithmetic can't turn on how long the harness took to build the feed.
func TestFixtureClockIsFrozen(t *testing.T) {
	for _, f := range Fixtures() {
		m, err := f.build()
		if err != nil {
			t.Fatal(err)
		}
		if got := m.wallNow(); !got.Equal(fixtureNow) {
			t.Errorf("%s: wallNow() = %v, want the frozen %v", f.ID, got, fixtureNow)
		}
	}
}

// TestNewKeepsReadingTheOperatorRecords: the seam must not change what a real
// client does — New() still loads the operator's own lessons-seen and unlocks
// records, exactly as before.
func TestNewKeepsReadingTheOperatorRecords(t *testing.T) {
	seedOperatorRecords(t)
	w, err := world.Create(t.TempDir()+"/w", "live", 42)
	if err != nil {
		t.Fatal(err)
	}
	m := New(w)
	if !m.lessons.seen[lessonCatalog[0].ID] {
		t.Error("New() no longer loads the operator's lessons-seen record")
	}
	if !m.unlocks.Earned(world.Stage4) {
		t.Error("New() no longer loads the operator's unlocks record")
	}
	if m.clientNow != nil || m.markSeen != nil {
		t.Error("New() must leave both harness seams nil — the live path is time.Now + worlds.MarkLessonSeen")
	}
}

// TestMidGameRosterCarriesAllThreeStates is AC #6's roster half: the dense
// fixture must show awake, asleep AND dead villagers, so the map glyphs, the
// villager strip and the roster's three renderings are all visible at once.
func TestMidGameRosterCarriesAllThreeStates(t *testing.T) {
	f, ok := fixtureByID(FixtureMidGame)
	if !ok {
		t.Fatal("mid-game fixture missing")
	}
	m, err := f.build()
	if err != nil {
		t.Fatal(err)
	}
	var awake, asleep, dead int
	for _, a := range m.replica.Agents {
		switch {
		case a.Dead:
			dead++
		case a.Asleep:
			asleep++
		default:
			awake++
		}
	}
	if awake == 0 || asleep == 0 || dead == 0 {
		t.Fatalf("roster = %d awake / %d asleep / %d dead, want at least one of each", awake, asleep, dead)
	}
	if len(m.replica.Deaths) != dead {
		t.Errorf("death ledger holds %d records for %d dead villagers", len(m.replica.Deaths), dead)
	}

	// The strip is where all three states show side by side: uppercase for
	// awake, lowercase for asleep, † for dead (panels/villager-strip.md).
	strip := m.villagerStripView(140)
	if !strings.Contains(strip, "†") {
		t.Errorf("villager strip carries no dead marker: %q", strip)
	}
}

// TestMidGameChronicleOverflows is AC #6's truncation half: the backlog must
// be deeper than any pane in the matrix can show, so the chronicle's
// truncation and the dock's row budget are exercised rather than merely
// asserted.
func TestMidGameChronicleOverflows(t *testing.T) {
	f, ok := fixtureByID(FixtureMidGame)
	if !ok {
		t.Fatal("mid-game fixture missing")
	}
	m, err := f.build()
	if err != nil {
		t.Fatal(err)
	}
	// The tallest matrix size is 50 rows; the chronicle body gets fewer than
	// that after the chrome. A backlog several times deeper guarantees
	// overflow at every size the harness renders.
	if len(m.events) < 200 {
		t.Fatalf("chronicle backlog is %d events — too shallow to overflow the pane", len(m.events))
	}
	m.width, m.height = 160, 50
	rendered := strings.Count(m.View(), "agent.")
	if rendered >= len(m.events) {
		t.Errorf("the pane rendered %d event rows out of a %d-event backlog — nothing truncated",
			rendered, len(m.events))
	}
}

// TestScenarioFixtureRendersExerciseTabAndLessonRow is AC #7: the exercise
// dock tab and the lesson row are the two surfaces the scenario fixture
// exists to show, and the exercise tab must be absent from the ambient
// fixtures — its presence is world-shaped (spec 054 FR-008), not a global.
func TestScenarioFixtureRendersExerciseTabAndLessonRow(t *testing.T) {
	frameOf := func(t *testing.T, id string) string {
		t.Helper()
		out, err := Frame(FrameOptions{Fixture: id, State: "home", Width: 160, Height: 50})
		if err != nil {
			t.Fatal(err)
		}
		return out
	}

	scenario := frameOf(t, FixtureScenario)
	if !strings.Contains(scenario, "exercise") {
		t.Errorf("the scenario fixture must render the exercise dock tab:\n%s", scenario)
	}
	if !strings.Contains(scenario, lessonPullSuffix) {
		t.Errorf("the scenario fixture must render the lesson row:\n%s", scenario)
	}

	for _, ambient := range []string{FixtureEmpty, FixtureMidGame} {
		if out := frameOf(t, ambient); strings.Contains(out, "exercise") {
			t.Errorf("ambient fixture %q rendered an exercise tab:\n%s", ambient, out)
		}
	}
}

// TestFixtureRegistry: the three shipped ids, each resolvable, each with a
// listing description, and the registry handed out as a copy.
func TestFixtureRegistry(t *testing.T) {
	want := []string{FixtureEmpty, FixtureMidGame, FixtureScenario}
	if got := FixtureIDs(); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("FixtureIDs() = %v, want %v", got, want)
	}
	for _, id := range want {
		f, ok := fixtureByID(id)
		if !ok {
			t.Fatalf("fixture %q does not resolve", id)
		}
		if f.Description == "" {
			t.Errorf("fixture %q has no listing description", id)
		}
	}
	Fixtures()[0].ID = "clobbered"
	if Fixtures()[0].ID != FixtureEmpty {
		t.Error("Fixtures() handed out the backing array")
	}
}

// TestFixturesTouchNoWorldDirectory: a fixture is a Go value, not a
// directory. Nothing under its nominal Dir may need to exist — that is what
// sidesteps spec 046's earned-stage gate and makes AC #1's
// machine-independence hold by construction.
func TestFixturesTouchNoWorldDirectory(t *testing.T) {
	for _, f := range Fixtures() {
		w := f.world()
		if _, err := os.Stat(w.Dir); !os.IsNotExist(err) {
			t.Errorf("fixture %q claims a directory that exists on disk (%s) — "+
				"fixtures must never depend on generated world dirs", f.ID, w.Dir)
		}
		if _, err := f.build(); err != nil {
			t.Errorf("fixture %q failed to build without its directory: %v", f.ID, err)
		}
	}
}

// TestMaterializeFixtureIsTheSameSceneAsTheDump is T012's actual claim: the
// interactive session the operator drives and the frame the matrix dumps are
// the same model, not two constructions that happen to look alike. If they
// ever diverge, every frame under docs/design/tui/frames/ stops being
// evidence about what the client shows.
func TestMaterializeFixtureIsTheSameSceneAsTheDump(t *testing.T) {
	const w, h = 160, 50
	for _, f := range Fixtures() {
		for _, state := range []string{"home", "help-lessons"} {
			t.Run(f.ID+"/"+state, func(t *testing.T) {
				dir := filepath.Join(t.TempDir(), "world")
				m, err := MaterializeFixture(f.ID, state, dir)
				if err != nil {
					t.Fatal(err)
				}
				// Bubble Tea supplies the size live; the harness supplies it
				// as --size. That is the only difference there may be.
				m.width, m.height = w, h
				restore := forceColorProfile(false)
				got := m.View()
				restore()

				want, err := Frame(FrameOptions{Fixture: f.ID, State: state, Width: w, Height: h})
				if err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Errorf("the interactive model and the dumped frame differ for %s/%s", f.ID, state)
				}

				// The materialized directory must be a real world dir carrying
				// the fixture's OWN manifest — stage and scenario config
				// included, since those two fields decide whether the exercise
				// tab and the lesson row render at all.
				opened, err := world.Open(dir)
				if err != nil {
					t.Fatalf("materialized dir is not a world: %v", err)
				}
				if opened.Manifest.Stage != f.manifest.Stage {
					t.Errorf("stage = %q, want %q", opened.Manifest.Stage, f.manifest.Stage)
				}
				if (opened.Manifest.Scenario == nil) != (f.manifest.Scenario == nil) {
					t.Errorf("scenario config was not carried into the materialized world")
				}

				// Offline: connecting would fail instantly against a world with
				// no daemon, flip `connected` off and start a retry loop — i.e.
				// the interactive session would show something OTHER than the
				// frame above, within a frame or two of opening.
				if !m.offline {
					t.Error("a materialized fixture must be offline")
				}
				if cmd := m.Init(); cmd != nil {
					t.Error("Init() on a fixture model must issue no command")
				}
			})
		}
	}
}

// TestMaterializeFixtureRejectsBadInput: an unknown fixture or state must
// fail loudly rather than open the client on a silently wrong scene.
func TestMaterializeFixtureRejectsBadInput(t *testing.T) {
	if _, err := MaterializeFixture("nope", "home", filepath.Join(t.TempDir(), "w")); err == nil {
		t.Error("unknown fixture: want an error")
	}
	if _, err := MaterializeFixture(FixtureEmpty, "nope", filepath.Join(t.TempDir(), "w")); err == nil {
		t.Error("unknown state: want an error")
	}
	occupied := t.TempDir()
	if err := os.WriteFile(filepath.Join(occupied, "something"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := MaterializeFixture(FixtureEmpty, "home", occupied); err == nil {
		t.Error("non-empty target directory: want an error")
	}
}

// --- spec 115: the raw feed wraps, aligned to the summary column ----------

// midGameSoloFrame renders the mid-game fixture's solo chronicle at a size,
// with ANSI stripped, as the line slice these tests assert over.
func midGameSoloFrame(t *testing.T, w, h int) []string {
	t.Helper()
	f, ok := fixtureByID(FixtureMidGame)
	if !ok {
		t.Fatal("mid-game fixture missing")
	}
	m, err := f.build()
	if err != nil {
		t.Fatal(err)
	}
	m.width, m.height = w, h
	m.solo = true
	m.dockTab = paneChronicle
	return strings.Split(ansiRe.ReplaceAllString(m.View(), ""), "\n")
}

// feedRowFor returns the index of the first line carrying the given event
// type in its type column, or -1.
func feedRowFor(lines []string, eventType string) int {
	for i, ln := range lines {
		if strings.Contains(ln, eventType) {
			return i
		}
	}
	return -1
}

// TestFeedWrapsLongProseInFull (spec 115 US1, SC-001/SC-002): the whole
// thought reaches the screen. Before this feature the summary was cut at the
// pane edge with "…" and the end of the sentence was unrecoverable — the feed
// being the only place it is shown.
func TestFeedWrapsLongProseInFull(t *testing.T) {
	lines := midGameSoloFrame(t, 160, 50)
	row := feedRowFor(lines, "agent.thought")
	if row < 0 {
		t.Fatal("the mid-game fixture no longer emits agent.thought — FR-012 regressed")
	}
	joined := strings.Join(lines[row:], " ")
	for _, fragment := range []string{
		"I keep coming back to the chest by the river",
		"while Rowan is in earshot",
	} {
		if !strings.Contains(joined, fragment) {
			t.Errorf("wrapped thought lost %q — the feed truncated it", fragment)
		}
	}
	if strings.Contains(lines[row], "…") {
		t.Errorf("a wrapped row must not be elided: %q", lines[row])
	}
}

// TestFeedContinuationAlignsToSummaryColumn (spec 115 US2, SC-003): every
// continuation line starts exactly where its own row's summary starts, and
// carries no tick, time or type content.
func TestFeedContinuationAlignsToSummaryColumn(t *testing.T) {
	lines := midGameSoloFrame(t, 160, 50)
	row := feedRowFor(lines, "agent.thought")
	if row < 0 {
		t.Fatal("no agent.thought row")
	}
	// Columns are counted in RUNES, not bytes: the pane border "\u2502" is three
	// bytes, so strings.Index would report a column two past the visual one.
	b := strings.Index(lines[row], "Ash thought:")
	if b < 0 {
		t.Fatalf("could not locate the summary column in %q", lines[row])
	}
	summaryCol := len([]rune(lines[row][:b]))
	cont := []rune(lines[row+1])
	contCol := 0
	for contCol < len(cont) && (cont[contCol] == ' ' || cont[contCol] == '\u2502') {
		contCol++
	}
	if contCol != summaryCol {
		t.Errorf("continuation starts at column %d, want the summary column %d\n row:  %q\n cont: %q",
			contCol, summaryCol, lines[row], string(cont))
	}
	rail := strings.TrimFunc(string(cont[:summaryCol]), func(r rune) bool { return r == ' ' || r == '\u2502' })
	if rail != "" {
		t.Errorf("continuation's left rail must carry no content, got %q", rail)
	}
}

// TestFeedWrapsWithoutSplittingWords (spec 115 FR-002): breaks land between
// words. Checked by rejoining the wrapped lines and comparing to the source —
// a mid-word split would leave a fragment that no longer reads as the text.
func TestFeedWrapsWithoutSplittingWords(t *testing.T) {
	lines := midGameSoloFrame(t, 160, 50)
	row := feedRowFor(lines, "social.conversation_turn")
	if row < 0 {
		t.Fatal("no social.conversation_turn row")
	}
	rejoined := strings.Join(strings.Fields(strings.Join(lines[row:row+2], " ")), " ")
	if !strings.Contains(rejoined, "we should say so at the meeting tonight") {
		t.Errorf("wrap split a word — rejoined text reads %q", rejoined)
	}
}

// TestFeedRespectsPaneWidthAtEverySize (spec 115 FR-007, SC-004) and the row
// budget (FR-008, SC-005): nothing overflows horizontally and the body still
// fits its rows once events occupy several lines each.
func TestFeedRespectsPaneWidthAtEverySize(t *testing.T) {
	// Scoped to the chronicle body — the surface spec 115 governs. The frame's
	// TITLE row overflows by one rune at 80 columns, which is PRE-EXISTING
	// (present in the committed pre-115 frame) and belongs to the spec-114
	// family of width clamps, not to this feature. Asserting over the whole
	// frame would silently adopt that bug as ours.
	for _, size := range []struct{ w, h int }{{112, 30}, {113, 30}, {160, 50}} {
		lines := midGameSoloFrame(t, size.w, size.h)
		start := -1
		for i, ln := range lines {
			if strings.Contains(ln, "raw feed") {
				start = i
				break
			}
		}
		if start < 0 {
			t.Fatalf("%dx%d rendered no raw feed — the test is no longer measuring the feed", size.w, size.h)
		}
		for i, ln := range lines[start:] {
			if n := len([]rune(ln)); n > size.w {
				t.Errorf("%dx%d feed line %d is %d runes, exceeds the pane: %q",
					size.w, size.h, start+i, n, ln)
			}
		}
		if len(lines) > size.h {
			t.Errorf("%dx%d rendered %d rows, over the %d-row budget", size.w, size.h, len(lines), size.h)
		}
	}
}

// TestNarrowFallbackFeedWrapsWithinWidth (spec 115 US3, T012): below the
// widescreen breakpoint the frame router shows the map rather than the feed,
// so the narrow-fallback chronicle renderer — the one T012 switched from
// truncate to unbounded wrap — is exercised directly. Its body must wrap long
// prose and stay inside the pane.
func TestNarrowFallbackFeedWrapsWithinWidth(t *testing.T) {
	f, ok := fixtureByID(FixtureMidGame)
	if !ok {
		t.Fatal("mid-game fixture missing")
	}
	m, err := f.build()
	if err != nil {
		t.Fatal(err)
	}
	m.width, m.height = 80, 30
	body := ansiRe.ReplaceAllString(m.chronicleView(), "")
	lines := strings.Split(body, "\n")
	// Lines 0-1 are the hint and its blank. The hint is deliberately unclamped
	// — in a real frame the pane border clips it, and calling the body renderer
	// directly bypasses that clipping. Only the feed rows are this test's
	// subject; chronicleView budgets them at m.width-4.
	const hintRows = 2
	if len(lines) <= hintRows {
		t.Fatalf("narrow fallback rendered no feed rows: %q", body)
	}
	bodyWidth := m.width - 4
	for i, ln := range lines[hintRows:] {
		if n := len([]rune(ln)); n > bodyWidth {
			t.Errorf("narrow fallback row %d is %d runes, exceeds body width %d: %q",
				i+hintRows, n, bodyWidth, ln)
		}
	}
	if !strings.Contains(body, "chest by the river") {
		t.Error("narrow fallback did not render the long thought at all")
	}
	if strings.Contains(body, "…") {
		t.Error("narrow fallback still truncates instead of wrapping")
	}
}

// TestFeedNewestRowStaysVisible (spec 115 FR-008): the feed auto-follows the
// tail, so multi-row events must not push the newest event out of view.
func TestFeedNewestRowStaysVisible(t *testing.T) {
	lines := midGameSoloFrame(t, 160, 50)
	if feedRowFor(lines, "social.conversation_turn") < 0 {
		t.Error("the newest event is not visible — wrapping pushed the tail off the pane")
	}
}
