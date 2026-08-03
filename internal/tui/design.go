package tui

// The design frame harness (spec 112, TASK-187): render one (fixture, state,
// size) combination to the exact string the terminal would receive.
//
// This lives INSIDE package tui rather than in cmd/promptworld because the
// layout-bearing Model fields — width, height, solo, dockTab, chronSelected,
// status — are unexported (plan.md F2). Exporting them to let a command pose
// a Model from outside would turn a private layout model into public API; a
// single documented Frame() entry point is the smaller surface. A CLI stays
// a thin shell over this.
//
// The harness is a productization of the in-package test helpers, not a
// parallel implementation (plan.md F3): States() is the ONE registry of
// state names and poseState() the ONE way to drive a Model into one, both
// consumed by render_test.go as well as by Frame. The tests and the harness
// therefore cannot disagree about what a state means.

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/sim"
)

// FrameOptions selects one renderable combination.
type FrameOptions struct {
	// Fixture is a shipped fixture id (FixtureIDs / Fixtures).
	Fixture string
	// State is a name from States().
	State string
	// Width and Height are the terminal size in cells; both must be positive.
	Width, Height int
	// ANSI keeps the color escapes in the output. Default (false) suppresses
	// them, so frames diff (spec.md FR-004).
	ANSI bool
}

// stateNames is the registry (spec.md FR-005). It is the matrix the harness
// dumps AND the matrix render_test.go's exact-height sweep drives, so a state
// can never exist in one and not the other.
//
// One name differs from the list spec.md FR-005 enumerates: the guardian
// tab's soloed state is `guardian-solo`, not the `metatron-solo` render_test.go
// had been calling it. That old spelling predates the spec 052 skin rename and
// the spec 094 event rename, and it survived only because the fiction-denylist
// sweep (internal/lint/fiction_sweep_test.go) skips _test.go files — moving the
// list into a production file is exactly the moment that gate is supposed to
// catch it. The state poses paneGuardian either way; every other name in this
// list is already post-052 vocabulary.
var stateNames = []string{
	"home", "solo", "inspect", "inspect-solo",
	"villagers-solo", "villagers-detail-solo", "guardian-solo",
	// Spec 045 (TASK-116): the help overlay is a body-replacement panel like
	// solo zoom — every section/tier is its own state.
	"help", "help-advanced", "help-walkthrough", "help-lessons",
}

// States returns the renderable state names, in matrix order. The slice is a
// copy: the registry is not a caller-mutable surface.
func States() []string {
	out := make([]string, len(stateNames))
	copy(out, stateNames)
	return out
}

// Frame renders one combination and returns exactly the string View() would
// hand Bubble Tea (spec.md FR-007 — the fidelity that makes every dumped
// frame trustworthy). No daemon, no LLM, no sim advancing: the fixture is a
// Go value and the frame is one pure View() call over it.
//
// NOT safe for concurrent use: suppressing ANSI goes through lipgloss's
// process-global color profile, the same seam render_test.go forces (F3).
func Frame(opts FrameOptions) (string, error) {
	f, ok := fixtureByID(opts.Fixture)
	if !ok {
		return "", fmt.Errorf("unknown fixture %q (want one of: %s)", opts.Fixture, strings.Join(FixtureIDs(), ", "))
	}
	if opts.Width <= 0 || opts.Height <= 0 {
		return "", fmt.Errorf("size %dx%d: width and height must both be positive", opts.Width, opts.Height)
	}
	m, err := f.build()
	if err != nil {
		return "", err
	}
	m.width, m.height = opts.Width, opts.Height
	if err := poseState(&m, opts.State); err != nil {
		return "", err
	}
	restore := forceColorProfile(opts.ANSI)
	defer restore()
	return m.View(), nil
}

// forceColorProfile pins lipgloss's color profile for the duration of a
// render and returns the restore func (the withColorProfile pattern
// render_test.go already uses, hoisted out of the test files so the harness
// and the tests force the SAME seam — spec.md FR-004's "reuse that seam
// rather than inventing a second one").
//
// Ascii makes every Style.Render a no-op, so no escape byte can reach the
// output; TrueColor is the opt-in that keeps color for live eyeballing.
func forceColorProfile(ansi bool) func() {
	prior := lipgloss.ColorProfile()
	profile := termenv.Ascii
	if ansi {
		profile = termenv.TrueColor
	}
	lipgloss.SetColorProfile(profile)
	return func() { lipgloss.SetColorProfile(prior) }
}

// poseState drives a constructed Model into the named screen state. It is
// the single implementation of "what does `inspect-solo` mean" — the frame
// harness and render_test.go's exact-height sweep both call it, so neither
// can drift from the other (plan.md F3).
//
// Every arm mutates only presentation state the live Update() would have
// reached by keypresses; none of them touch the fixture's world, feed, or
// canned per-user record.
func poseState(m *Model, state string) error {
	switch state {
	case "home":
		// The default composite — nothing to pose.
	case "solo":
		m.solo = true
	case "inspect":
		m.connected = true
		m.pauseForInspect()
		m.chronSelected = 5
	case "inspect-solo":
		m.connected = true
		m.pauseForInspect()
		m.chronSelected = 5
		m.solo = true
	case "villagers-solo":
		m.dockTab = paneVillagers
		m.solo = true
	case "villagers-detail-solo":
		m.dockTab = paneVillagers
		m.solo = true
		m.villDetail = true
		poseVillagerDetail(m)
	case "guardian-solo":
		m.dockTab = paneGuardian
		m.solo = true
		m.transcript = append(m.transcript,
			"you: why is Rowan hoarding wood?",
			transcriptGuardianPrefix+"three cold nights, and Ash let the fire die each time.")
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
	default:
		return fmt.Errorf("unknown state %q (want one of: %s)", state, strings.Join(stateNames, ", "))
	}
	return nil
}

// pauseForInspect marks the model paused — inspect mode is exactly "paused
// and the chronicle visible" (tui.go inspecting()). A model with no status
// yet gains a bare paused one; a model carrying a canned snapshot keeps every
// other field of it, so a fixture's frozen tick/day/speed survive being
// posed. The snapshot is copied rather than mutated in place: the canned
// value must never be shared-and-modified across two frames.
func (m *Model) pauseForInspect() {
	st := ipc.StatusData{}
	if m.status != nil {
		st = *m.status
	}
	st.Clock.Paused = true
	m.status = &st
}

// poseVillagerDetail gives the selected villager enough interior life —
// beliefs, a narrative, an overflowing memory list — for the detail pane's
// sections and its scroll affordance to be visible at once. A roster-less
// model (a fixture with no replica yet) is left alone rather than panicking.
func poseVillagerDetail(m *Model) {
	if m.replica == nil || len(m.replica.Agents) == 0 {
		return
	}
	a := &m.replica.Agents[m.clampedVillSelected()]
	a.Beliefs = []sim.Belief{{Statement: "the fire needs tending", Confidence: 80}}
	a.Narrative = "a long night watching the fire die and reviving it by hand."
	for i := 0; i < 20; i++ {
		a.Memories = append(a.Memories,
			sim.Memory{Text: "chopped wood at the treeline", Salience: 3, Tick: int64(i) * 60})
	}
}

// clampedVillSelected is the roster cursor clamped into range — the same
// defensive clamp every roster read applies (the replica can arrive late or
// be swapped wholesale on reconnect).
func (m Model) clampedVillSelected() int {
	if m.villSelected < 0 || m.villSelected >= len(m.replica.Agents) {
		return 0
	}
	return m.villSelected
}
