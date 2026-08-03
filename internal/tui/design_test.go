package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// frameSizes straddle the widescreen breakpoint (112) and the 50/50 column
// split, plus one narrow size below it — the shape FR-005's matrix takes.
var frameSizes = []struct{ w, h int }{
	{80, 30},  // narrow fallback
	{112, 30}, // just over the breakpoint
	{140, 40}, // the common working size
	{160, 50}, // wide and tall
	{200, 24}, // wide and short — the fold cascade under height pressure
}

// TestFrameRendersEveryCombination sweeps every (fixture, state, size) the
// harness can be asked for: each frame must render non-empty and carry no
// ANSI escape under the default plain-text posture (FR-004).
//
// The exact-height check is applied to WIDESCREEN sizes only, matching the
// invariant that actually holds: TestWidescreenViewExactHeight sweeps sizes
// at and above the breakpoint because the composite is the layout with a row
// budget. The narrow fallback has no fold arithmetic of its own (patterns/
// layout.md ruling b; the guardian-strip and lesson-row precedents both say
// so explicitly), so its frame is content-height, not terminal-height —
// asserting otherwise here would be inventing a guarantee the renderer has
// never made. Narrow frames are still swept for the two guarantees the
// harness itself owns.
func TestFrameRendersEveryCombination(t *testing.T) {
	for _, f := range Fixtures() {
		for _, state := range States() {
			for _, sz := range frameSizes {
				t.Run(fmt.Sprintf("%s/%s/%dx%d", f.ID, state, sz.w, sz.h), func(t *testing.T) {
					out, err := Frame(FrameOptions{Fixture: f.ID, State: state, Width: sz.w, Height: sz.h})
					if err != nil {
						t.Fatal(err)
					}
					if out == "" {
						t.Fatal("empty frame")
					}
					if lines := strings.Split(out, "\n"); isWidescreen(sz.w) && len(lines) != sz.h {
						t.Errorf("frame = %d lines, want exactly %d:\n%s", len(lines), sz.h, out)
					}
					if strings.Contains(out, "\x1b") {
						t.Errorf("plain-text frame carries an ANSI escape byte:\n%q", out)
					}
				})
			}
		}
	}
}

// TestFrameMatchesDirectView is the fidelity assertion (FR-007, AC #9): a
// dumped frame is not a re-rendering of the UI, it IS the UI — Frame's
// output must equal what the same fixture Model's View() produces at the
// same size and state. Every other guarantee in this feature rests on this
// one: the frames under docs/design/tui/frames/ are only evidence about the
// client if this holds.
//
// AC #9 asks for "at least one page per fixture", and this test began as
// exactly that — two states, one widescreen size. That floor is too low for
// what the assertion is load-bearing FOR (spec 112 T019, judged during phase
// 5). Frame() does three things View() does not: it poses the Model through
// poseState, it sets width/height itself, and it forces the color profile.
// Every one of those is per-(state, size), so a sample of two states at one
// widescreen size leaves the failure modes that matter uncovered — most
// sharply the NARROW fallback, which is a structurally different render
// branch (narrowView, no fold arithmetic — FR-008) and which the committed
// matrix ships a frame for at 80x30. A Frame/View divergence there would
// have been invisible.
//
// So the sweep is total: every fixture, every registered state, every size
// the harness renders at. Fidelity is not sampled, it is exhaustive over the
// harness's own domain — which is affordable precisely because a frame is a
// pure function of a Go value.
func TestFrameMatchesDirectView(t *testing.T) {
	for _, f := range Fixtures() {
		for _, state := range States() {
			for _, sz := range frameSizes {
				t.Run(fmt.Sprintf("%s/%s/%dx%d", f.ID, state, sz.w, sz.h), func(t *testing.T) {
					got, err := Frame(FrameOptions{Fixture: f.ID, State: state, Width: sz.w, Height: sz.h})
					if err != nil {
						t.Fatal(err)
					}

					m, err := f.build()
					if err != nil {
						t.Fatal(err)
					}
					m.width, m.height = sz.w, sz.h
					if err := poseState(&m, state); err != nil {
						t.Fatal(err)
					}
					restore := forceColorProfile(false)
					want := m.View()
					restore()

					if got != want {
						t.Errorf("Frame() diverged from a direct View() on the same posed Model:\ngot:\n%s\nwant:\n%s", got, want)
					}
				})
			}
		}
	}
}

// TestFrameDeterministic is FR-003 at the single-frame grain: the same
// combination rendered twice is byte-identical. (The whole-matrix form is
// the driver's own test.)
func TestFrameDeterministic(t *testing.T) {
	for _, f := range Fixtures() {
		for _, state := range States() {
			opts := FrameOptions{Fixture: f.ID, State: state, Width: 140, Height: 40}
			first, err := Frame(opts)
			if err != nil {
				t.Fatal(err)
			}
			second, err := Frame(opts)
			if err != nil {
				t.Fatal(err)
			}
			if first != second {
				t.Errorf("%s/%s: two renders of the same combination differ", f.ID, state)
			}
		}
	}
}

// TestFrameANSIOptIn is FR-004's two postures: plain by default so frames
// diff, escapes preserved on request so a live eyeballing keeps its color.
func TestFrameANSIOptIn(t *testing.T) {
	opts := FrameOptions{Fixture: FixtureMidGame, State: "home", Width: 140, Height: 40}
	plain, err := Frame(opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(plain, "\x1b") {
		t.Error("default frame must carry no escapes")
	}

	opts.ANSI = true
	colored, err := Frame(opts)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(colored, "\x1b") {
		t.Error("ANSI: true must preserve the color escapes")
	}
	if colored == plain {
		t.Error("the two postures produced identical output")
	}
}

// TestFrameRestoresColorProfile: the harness borrows a process-global
// lipgloss setting, so it must hand it back — otherwise one Frame call
// silently recolors every later render in the process (including the rest of
// this test binary).
func TestFrameRestoresColorProfile(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	before := lipgloss.ColorProfile()
	if _, err := Frame(FrameOptions{Fixture: FixtureEmpty, State: "home", Width: 140, Height: 40}); err != nil {
		t.Fatal(err)
	}
	if after := lipgloss.ColorProfile(); after != before {
		t.Errorf("color profile left at %v, want the prior %v", after, before)
	}
}

// TestFrameRejectsBadOptions: every selector is closed-vocabulary, and the
// error names the alternatives rather than failing mutely.
func TestFrameRejectsBadOptions(t *testing.T) {
	cases := []struct {
		name string
		opts FrameOptions
		want string
	}{
		{"unknown fixture", FrameOptions{Fixture: "nope", State: "home", Width: 80, Height: 24}, "unknown fixture"},
		{"unknown state", FrameOptions{Fixture: FixtureEmpty, State: "nope", Width: 80, Height: 24}, "unknown state"},
		{"zero width", FrameOptions{Fixture: FixtureEmpty, State: "home", Width: 0, Height: 24}, "must both be positive"},
		{"negative height", FrameOptions{Fixture: FixtureEmpty, State: "home", Width: 80, Height: -1}, "must both be positive"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := Frame(c.opts)
			if err == nil {
				t.Fatalf("want an error, got a %d-byte frame", len(out))
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q should mention %q", err, c.want)
			}
		})
	}
}

// TestStatesIsTheSingleRegistry: States() is the ONE list of state names
// (FR-005) — it is what render_test.go's exact-height sweep drives and what
// the matrix dumps, and every name it hands out must be poseable. The
// explicit expectation below is the spec's enumerated starting matrix: this
// test fails when a state is added or renamed, which is the moment the
// design docs and the matrix directory need the same edit.
func TestStatesIsTheSingleRegistry(t *testing.T) {
	want := []string{
		"home", "solo", "inspect", "inspect-solo",
		"villagers-solo", "villagers-detail-solo", "guardian-solo",
		"help", "help-advanced", "help-walkthrough", "help-lessons",
	}
	got := States()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("States() = %v, want %v", got, want)
	}
	for _, state := range got {
		m := testModel(t)
		if err := poseState(&m, state); err != nil {
			t.Errorf("registered state %q is not poseable: %v", state, err)
		}
	}
	// The registry is a copy — a caller mangling its slice must not corrupt
	// the harness for everyone else.
	States()[0] = "clobbered"
	if States()[0] != "home" {
		t.Error("States() handed out the backing array")
	}
}

// TestPoseInspectPreservesCannedStatus: posing a paused state must flip
// Paused without discarding the fixture's frozen tick/day/speed — a poser
// that overwrote the whole snapshot would render every inspect frame at
// tick 0 (and would silently mutate a shared canned value).
func TestPoseInspectPreservesCannedStatus(t *testing.T) {
	f, ok := fixtureByID(FixtureMidGame)
	if !ok {
		t.Fatal("mid-game fixture missing")
	}
	m, err := f.build()
	if err != nil {
		t.Fatal(err)
	}
	canned := m.status
	if err := poseState(&m, "inspect"); err != nil {
		t.Fatal(err)
	}
	if !m.status.Clock.Paused {
		t.Error("inspect must pause the clock")
	}
	if m.status.Clock.Tick != midGameTick {
		t.Errorf("posed status tick = %d, want the fixture's canned %d", m.status.Clock.Tick, midGameTick)
	}
	if canned.Clock.Paused {
		t.Error("posing mutated the fixture's own status snapshot in place")
	}
}
