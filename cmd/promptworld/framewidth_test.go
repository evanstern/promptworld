package main

// Spec 114 (specs/114-map-legend-width) FR-009/FR-010, contract C1: no frame
// in the matrix may emit a line wider than the terminal it declares.
//
// This is the guard that turns "we fixed the legend" into something that stays
// fixed. It is deliberately matrix-WIDE rather than legend-only: a guard
// scoped to the one line this feature touched would leave every other surface
// free to regress in exactly the way the legend did.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/evanstern/promptworld/internal/tui"
)

// knownOverWideFrames is the registry of frames that still emit an over-width
// line for a defect spec 114 deliberately did NOT fix. Each entry names the
// board card that retires it.
//
// The registry is BIDIRECTIONAL, and that is the point: an unlisted frame that
// fails is a failure, and a listed frame that PASSES is also a failure. Debt
// that someone quietly fixes cannot linger here pretending to still be owed,
// and debt that grows cannot hide behind a stale allowance. The list can only
// shrink.
//
// Both remaining classes come from the same 2026-08-02 frame audit that found
// the legend defect (spec.md "Evidence"):
//
//   - Header row, 20 frames, line 1 at 81-83 columns. The status/header line
//     ("<World> — tick N · day D HH:MM · running · speed Nx (N t/s) [N villagers]")
//     is not clamped to the terminal width on the narrow path. Worst at 83
//     columns on the `scenario` fixture, whose world name is longest.
//   - Scenario keybind footer, 2 frames, line 30 at 121 columns. Only the
//     scenario fixture's home state, which carries extra lesson-row keybinds.
//
// TODO(TASK-192): header-row overflow — 20 frames.
// TODO(TASK-193): scenario keybind footer overflow — 2 frames.
var knownOverWideFrames = map[string]string{
	// Header row (line 1) — TASK-192.
	"mid-game__guardian-solo__80x30.txt":         "TASK-192",
	"mid-game__help__80x30.txt":                  "TASK-192",
	"mid-game__help-advanced__80x30.txt":         "TASK-192",
	"mid-game__help-lessons__80x30.txt":          "TASK-192",
	"mid-game__help-walkthrough__80x30.txt":      "TASK-192",
	"mid-game__home__80x30.txt":                  "TASK-192",
	"mid-game__solo__80x30.txt":                  "TASK-192",
	"mid-game__villagers-detail-solo__80x30.txt": "TASK-192",
	"mid-game__villagers-solo__80x30.txt":        "TASK-192",
	"scenario__guardian-solo__80x30.txt":         "TASK-192",
	"scenario__help__80x30.txt":                  "TASK-192",
	"scenario__help-advanced__80x30.txt":         "TASK-192",
	"scenario__help-lessons__80x30.txt":          "TASK-192",
	"scenario__help-walkthrough__80x30.txt":      "TASK-192",
	"scenario__home__80x30.txt":                  "TASK-192",
	"scenario__inspect__80x30.txt":               "TASK-192",
	"scenario__inspect-solo__80x30.txt":          "TASK-192",
	"scenario__solo__80x30.txt":                  "TASK-192",
	"scenario__villagers-detail-solo__80x30.txt": "TASK-192",
	"scenario__villagers-solo__80x30.txt":        "TASK-192",

	// Scenario keybind footer (line 30) — TASK-193.
	"scenario__home__112x30.txt": "TASK-193",
	"scenario__home__113x30.txt": "TASK-193",
}

// TestFramesNeverExceedDeclaredWidth is contract C1.
//
// Width is measured in DISPLAY COLUMNS via ansi.StringWidth, never in runes.
// Rune counting understates width wherever a double-width glyph appears (the
// guardian strip's ⚡ and 👁), and the original spec 114 audit used runes —
// which is why it could only ever produce false negatives. Declared width
// comes from frameSizes, not from parsing the filename, so the guard and the
// dumper can never disagree about what a size means.
func TestFramesNeverExceedDeclaredWidth(t *testing.T) {
	seen := map[string]bool{}

	for _, f := range tui.Fixtures() {
		for _, state := range tui.States() {
			for _, sz := range frameSizes {
				frame, err := tui.Frame(tui.FrameOptions{
					Fixture: f.ID, State: state, Width: sz.W, Height: sz.H,
				})
				if err != nil {
					t.Fatal(err)
				}
				name := frameFileName(f.ID, state, sz)

				var over []string
				for i, line := range strings.Split(frame, "\n") {
					if w := ansi.StringWidth(line); w > sz.W {
						over = append(over, lineReport(i+1, w, sz.W, line))
					}
				}

				card, allowed := knownOverWideFrames[name]
				switch {
				case len(over) > 0 && !allowed:
					t.Errorf("%s: %d line(s) exceed the declared width of %d columns — "+
						"a frame must never emit a line wider than the terminal it declares "+
						"(spec 114 contract C1)\n%s",
						name, len(over), sz.W, strings.Join(over, "\n"))
				case len(over) > 0 && allowed:
					seen[name] = true
				case len(over) == 0 && allowed:
					seen[name] = true
					t.Errorf("%s now fits within %d columns but is still listed in "+
						"knownOverWideFrames (%s). Remove the entry — the registry can only "+
						"shrink, and a stale allowance hides the next regression.",
						name, sz.W, card)
				}
			}
		}
	}

	// An entry naming a frame the matrix no longer produces is dead weight and
	// would silently excuse a future frame that happens to reuse the name.
	for name, card := range knownOverWideFrames {
		if !seen[name] {
			t.Errorf("knownOverWideFrames lists %s (%s), but the matrix produces no such frame — "+
				"remove the entry", name, card)
		}
	}
}

// lineReport renders one violation with enough context to act on without
// re-running the audit by hand.
func lineReport(lineNo, got, want int, line string) string {
	const preview = 60
	shown := ansi.Truncate(line, preview, "…")
	return fmt.Sprintf("    line %d: %d columns (over by %d): %s", lineNo, got, got-want, shown)
}
