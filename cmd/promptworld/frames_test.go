package main

// Flag/size parsing and dump plumbing for the frame harness CLI (spec 112
// T013). The heavier claims this shell rests on — determinism (AC #3) and
// fidelity to View() (AC #9) — are asserted where they belong, over the
// render API itself.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/tui"
)

func TestParseSize(t *testing.T) {
	ok := []struct {
		in   string
		w, h int
	}{
		{"160x50", 160, 50},
		{"80x30", 80, 30},
		{"112X30", 112, 30}, // capital X: both separators get typed
		{" 113 x 30 ", 113, 30},
		{"1x1", 1, 1},
	}
	for _, tc := range ok {
		w, h, err := parseSize(tc.in)
		if err != nil {
			t.Errorf("parseSize(%q) = error %v, want %dx%d", tc.in, err, tc.w, tc.h)
			continue
		}
		if w != tc.w || h != tc.h {
			t.Errorf("parseSize(%q) = %dx%d, want %dx%d", tc.in, w, h, tc.w, tc.h)
		}
	}

	// Every one of these must be a hard error rather than a silent default:
	// a mistyped size would otherwise dump a plausible frame at the wrong
	// dimensions, which is worse than no frame at all.
	bad := []string{"", "160", "160,50", "x50", "160x", "axb", "0x50", "160x0", "-8x30", "160x50x2"}
	for _, in := range bad {
		if w, h, err := parseSize(in); err == nil {
			t.Errorf("parseSize(%q) = %dx%d, want an error", in, w, h)
		}
	}
}

func TestFramesModeOf(t *testing.T) {
	cases := []struct {
		list, dump, interactive bool
		want                    framesMode
	}{
		{false, false, false, frameModeSingle},
		{true, false, false, frameModeList},
		{false, true, false, frameModeDump},
		{false, false, true, frameModeInteractive},
	}
	for _, tc := range cases {
		got, err := framesModeOf(tc.list, tc.dump, tc.interactive)
		if err != nil {
			t.Errorf("framesModeOf(%v,%v,%v) = error %v", tc.list, tc.dump, tc.interactive, err)
			continue
		}
		if got != tc.want {
			t.Errorf("framesModeOf(%v,%v,%v) = %v, want %v", tc.list, tc.dump, tc.interactive, got, tc.want)
		}
	}
	conflicts := [][3]bool{{true, true, false}, {true, false, true}, {false, true, true}, {true, true, true}}
	for _, c := range conflicts {
		if _, err := framesModeOf(c[0], c[1], c[2]); err == nil {
			t.Errorf("framesModeOf(%v,%v,%v) = no error, want the mutual-exclusion error", c[0], c[1], c[2])
		}
	}
}

func TestFrameFileName(t *testing.T) {
	got := frameFileName("mid-game", "help-walkthrough", frameSize{112, 30})
	want := "mid-game__help-walkthrough__112x30.txt"
	if got != want {
		t.Errorf("frameFileName = %q, want %q", got, want)
	}
	// The double separator is load-bearing: single hyphens already occur
	// inside both a fixture id and a state name, so a single-hyphen joiner
	// would make the name ambiguous to read back.
	if strings.Count(got, "__") != 2 {
		t.Errorf("frameFileName = %q, want exactly two %q separators", got, "__")
	}
}

func TestCmdFramesRejectsBadInput(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"no fixture", nil},
		{"unknown fixture", []string{"--fixture", "nope"}},
		{"unknown state", []string{"--fixture", "empty", "--state", "nope"}},
		{"bad size", []string{"--fixture", "empty", "--size", "big"}},
		{"mode conflict", []string{"--list", "--dump"}},
		{"positional argument", []string{"--list", "extra"}},
		{"interactive without fixture", []string{"--interactive"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := cmdFrames(tc.args); err == nil {
				t.Errorf("cmdFrames(%q) = nil, want an error", tc.args)
			}
		})
	}
}

// TestFrameSizesStraddleTheBoundaries pins the size set's REASON (spec.md
// FR-005): it must cross the widescreen breakpoint and cross the odd/even
// column split, or the matrix stops covering the layout boundaries it exists
// to cover. Trimming the set is fine; trimming it below these is not.
func TestFrameSizesStraddleTheBoundaries(t *testing.T) {
	var narrow, wide, oddSplit, evenSplit bool
	for _, sz := range frameSizes {
		if sz.W < 112 {
			narrow = true
		} else {
			wide = true
		}
		// computeColumns splits (W - 1 gutter col) in half; an odd remainder
		// gives the map the leftover column.
		if (sz.W-1)%2 == 1 {
			oddSplit = true
		} else {
			evenSplit = true
		}
	}
	if !narrow || !wide {
		t.Errorf("frameSizes must straddle the widescreen breakpoint (112): narrow=%v wide=%v", narrow, wide)
	}
	if !oddSplit || !evenSplit {
		t.Errorf("frameSizes must straddle the 50/50 column split: odd=%v even=%v", oddSplit, evenSplit)
	}
}

func TestDumpFramesWritesEveryCombination(t *testing.T) {
	dir := t.TempDir()
	names, err := dumpFrames(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := len(tui.Fixtures()) * len(tui.States()) * len(frameSizes)
	if len(names) != want {
		t.Errorf("dumpFrames wrote %d files, want %d (fixtures x states x sizes)", len(names), want)
	}
	seen := map[string]bool{}
	for _, n := range names {
		if seen[n] {
			t.Errorf("duplicate frame file name %q — two combinations collided", n)
		}
		seen[n] = true
		info, err := os.Stat(filepath.Join(dir, n))
		if err != nil {
			t.Errorf("%s: %v", n, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("%s is empty", n)
		}
	}
}

// TestDumpFramesPrunesStaleFrames: the directory must be exactly the current
// generation, never the union of every generation ever run — a renamed state
// leaving its old file behind would silently ship a frame no code produces.
// The authored README.md is never touched.
func TestDumpFramesPrunesStaleFrames(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, "gone__home__160x50.txt")
	if err := os.WriteFile(stale, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("authored\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := dumpFrames(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("stale frame %s survived the dump (stat err %v)", filepath.Base(stale), err)
	}
	if _, err := os.Stat(readme); err != nil {
		t.Errorf("the authored README.md must survive a dump: %v", err)
	}
}
