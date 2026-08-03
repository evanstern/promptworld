package main

// The design frame harness's CLI shell (spec 112, TASK-187): print one
// rendered frame, dump the whole matrix, list what's renderable, or open the
// real interactive client on a canned fixture.
//
// This file is deliberately thin. Everything that knows what a frame IS —
// the fixtures, the state registry, the posing, the ANSI seam — lives in
// internal/tui (design.go, fixtures.go), because the layout-bearing Model
// fields are unexported and exporting them to let a command pose a Model
// from outside would turn a private layout model into public API (plan.md
// F2). What lives HERE is flag shape, the dump matrix's size set, and file
// naming.
//
// Nothing on any path below starts a daemon, configures an LLM, or advances
// the sim (spec.md FR-002/AC #2): a fixture is a Go value and a frame is one
// pure View() call over it.

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evanstern/promptworld/internal/tui"
)

const framesUsage = `promptworld frames — render the TUI headlessly against a canned fixture world

Usage:
  promptworld frames --fixture <id> [--state <name>] [--size WxH] [--ansi]
  promptworld frames --list
  promptworld frames --dump [--out DIR]
  promptworld frames --fixture <id> [--state <name>] --interactive

Flags:
`

// defaultFramesDir is where --dump writes: the design corpus's generated
// wing (docs/design/tui/frames/README.md). Relative to the repo root, which
// is where the dump is meant to be run from.
const defaultFramesDir = "docs/design/tui/frames"

// frameSize is one terminal size in the dump matrix.
type frameSize struct{ W, H int }

func (s frameSize) String() string { return fmt.Sprintf("%dx%d", s.W, s.H) }

// frameSizes is the dump matrix's size set (spec.md FR-005, AC #5). Four
// sizes, not a width sweep: the matrix is 3 fixtures x len(States()) states
// x this list, and every size added multiplies the whole thing — a churny
// matrix is a stated risk (plan.md R2). Each of these earns its place by
// straddling a real layout boundary:
//
//	80x30   below the widescreen breakpoint (112) — the narrow single-pane
//	        fallback. NOTE (spec.md FR-008): narrow frames are CONTENT-height,
//	        not terminal-height, because the fallback has no fold arithmetic
//	        (patterns/layout.md ruling b). These files are legitimately
//	        shorter than 30 lines; that is not a defect to "fix".
//	112x30  exactly AT the breakpoint, and an odd column remainder — the
//	        composite's first width, with the map taking the leftover column
//	        (computeColumns).
//	113x30  one column wider: the even 50/50 map|dock split. The pair
//	        112/113 is the column-split straddle.
//	160x50  roomy and deep — wide enough for full pane content and tall
//	        enough that the fold cascade never fires, so all three teaching
//	        chrome rows are visible at once.
var frameSizes = []frameSize{
	{80, 30},
	{112, 30},
	{113, 30},
	{160, 50},
}

// framesMode is which of the four jobs a `frames` invocation is doing.
type framesMode int

const (
	frameModeSingle      framesMode = iota // print one frame to stdout
	frameModeList                          // enumerate fixtures, states, sizes
	frameModeDump                          // write the whole matrix to a directory
	frameModeInteractive                   // run the real TUI on a fixture
)

// framesModeOf resolves the mutually exclusive mode flags. Split out from
// cmdFrames so the exclusivity rule is unit-testable without a FlagSet.
func framesModeOf(list, dump, interactive bool) (framesMode, error) {
	var mode framesMode
	n := 0
	if list {
		mode, n = frameModeList, n+1
	}
	if dump {
		mode, n = frameModeDump, n+1
	}
	if interactive {
		mode, n = frameModeInteractive, n+1
	}
	if n > 1 {
		return 0, fmt.Errorf("--list, --dump and --interactive are mutually exclusive")
	}
	return mode, nil
}

// parseSize parses a WxH terminal size. Both separators are accepted because
// both get typed; everything else is a hard error rather than a silent
// default, since a mistyped size would otherwise dump a plausible-looking
// frame at the wrong dimensions.
func parseSize(s string) (int, int, error) {
	i := strings.IndexAny(s, "xX")
	if i < 0 {
		return 0, 0, fmt.Errorf("size %q: want WxH (e.g. 160x50)", s)
	}
	w, err := strconv.Atoi(strings.TrimSpace(s[:i]))
	if err != nil {
		return 0, 0, fmt.Errorf("size %q: width is not a number", s)
	}
	h, err := strconv.Atoi(strings.TrimSpace(s[i+1:]))
	if err != nil {
		return 0, 0, fmt.Errorf("size %q: height is not a number", s)
	}
	if w <= 0 || h <= 0 {
		return 0, 0, fmt.Errorf("size %q: width and height must both be positive", s)
	}
	return w, h, nil
}

// frameFileName is the ONE naming rule for a dumped frame, so the dump, the
// pruner, and anyone reading the directory agree. Double underscores because
// a state name already contains single hyphens (`help-walkthrough`) and a
// fixture id does too (`mid-game`).
func frameFileName(fixture, state string, sz frameSize) string {
	return fmt.Sprintf("%s__%s__%s.txt", fixture, state, sz)
}

func cmdFrames(args []string) error {
	fs := flag.NewFlagSet("frames", flag.ContinueOnError)
	var (
		fixture     = fs.String("fixture", "", "fixture id to render (see --list)")
		state       = fs.String("state", "home", "screen state to pose (see --list)")
		size        = fs.String("size", "160x50", "terminal size as WxH")
		ansi        = fs.Bool("ansi", false, "keep the color escapes (default: plain text, so frames diff)")
		list        = fs.Bool("list", false, "list the fixtures, states and dump sizes")
		dump        = fs.Bool("dump", false, "write every (fixture, state, size) combination to --out")
		interactive = fs.Bool("interactive", false, "open the real interactive TUI on the fixture")
		out         = fs.String("out", defaultFramesDir, "directory --dump writes into")
	)
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), framesUsage)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q (frames takes flags only)", fs.Arg(0))
	}

	mode, err := framesModeOf(*list, *dump, *interactive)
	if err != nil {
		return err
	}

	switch mode {
	case frameModeList:
		printFrameCatalog(os.Stdout)
		return nil

	case frameModeDump:
		names, err := dumpFrames(*out)
		if err != nil {
			return err
		}
		fmt.Printf("wrote %d frames to %s\n", len(names), *out)
		return nil

	case frameModeInteractive:
		if *fixture == "" {
			return fmt.Errorf("--interactive needs --fixture (see --list)")
		}
		return runFramesInteractive(*fixture, *state)
	}

	if *fixture == "" {
		return fmt.Errorf("--fixture is required (see --list)")
	}
	w, h, err := parseSize(*size)
	if err != nil {
		return err
	}
	frame, err := tui.Frame(tui.FrameOptions{
		Fixture: *fixture, State: *state, Width: w, Height: h, ANSI: *ansi,
	})
	if err != nil {
		return err
	}
	fmt.Println(frame)
	return nil
}

// printFrameCatalog enumerates what is renderable. Fixtures and states are
// read from tui.Fixtures()/tui.States() — never a second hand-written list
// here, which is the whole reason States() exists (plan.md F3).
func printFrameCatalog(w io.Writer) {
	fmt.Fprintln(w, "fixtures:")
	for _, f := range tui.Fixtures() {
		fmt.Fprintf(w, "  %-9s %s\n", f.ID, f.Description)
	}
	fmt.Fprintln(w, "\nstates:")
	for _, s := range tui.States() {
		fmt.Fprintf(w, "  %s\n", s)
	}
	fmt.Fprintln(w, "\ndump sizes:")
	for _, s := range frameSizes {
		fmt.Fprintf(w, "  %s\n", s)
	}
}

// dumpFrames writes every (fixture, state, size) combination under dir, one
// file per combination (spec.md FR-005), and returns the file names written
// in write order.
//
// A frame file is exactly the frame plus a trailing newline — no header, no
// metadata. The filename carries the identity, and a bare frame keeps a diff
// of two matrices a diff of what actually changed on screen.
//
// Stale `.txt` files are pruned, so the directory is exactly the current
// generation rather than the union of every generation ever run: if a state
// is renamed or a size dropped, its old files go. Only `.txt` is touched —
// the directory's README.md is authored, not generated.
func dumpFrames(dir string) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	var names []string
	written := map[string]bool{}
	for _, f := range tui.Fixtures() {
		for _, state := range tui.States() {
			for _, sz := range frameSizes {
				frame, err := tui.Frame(tui.FrameOptions{
					Fixture: f.ID, State: state, Width: sz.W, Height: sz.H,
				})
				if err != nil {
					return nil, err
				}
				name := frameFileName(f.ID, state, sz)
				if err := os.WriteFile(filepath.Join(dir, name), []byte(frame+"\n"), 0o644); err != nil {
					return nil, err
				}
				names = append(names, name)
				written[name] = true
			}
		}
	}
	if err := pruneFrames(dir, written); err != nil {
		return nil, err
	}
	return names, nil
}

// pruneFrames removes the `.txt` files in dir that this generation did not
// write.
func pruneFrames(dir string, written map[string]bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") || written[e.Name()] {
			continue
		}
		if err := os.Remove(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

// runFramesInteractive opens the real client on a fixture (spec.md FR-006,
// AC #8). The tea.NewProgram construction is character-for-character cmdUI's
// (commands.go), including the FatalErr surfacing, so what the operator
// drives here is the shipped client and not a harness lookalike.
func runFramesInteractive(fixture, state string) error {
	dir, err := os.MkdirTemp("", "promptworld-frames-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	fixtureModel, err := tui.MaterializeFixture(fixture, state, dir)
	if err != nil {
		return err
	}
	m, err := tea.NewProgram(fixtureModel, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
	if err != nil {
		return err
	}
	if fm, ok := m.(tui.Model); ok && fm.FatalErr() != "" {
		return fmt.Errorf("%s", fm.FatalErr())
	}
	return nil
}
