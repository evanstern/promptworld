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
