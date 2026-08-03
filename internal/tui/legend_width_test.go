package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/evanstern/promptworld/internal/sim"
)

// Spec 114 (specs/114-map-legend-width) — the map legend must never exceed
// the width it renders into, and must say so when it was shortened.
// contracts/legend-width.md C1–C5.

// TestClipLegendFitsUnchanged pins C2's second half: a legend that fits is
// returned untouched — no ellipsis, no padding. An implementation that always
// appends a marker lies in the opposite direction, telling the player content
// was hidden when none was.
func TestClipLegendFitsUnchanged(t *testing.T) {
	const s = "night · ~water ♠wood"
	w := ansi.StringWidth(s)
	for _, budget := range []int{w, w + 1, w + 40} {
		got := clipLegend(s, budget)
		if got != s {
			t.Errorf("budget %d: legend that fits was modified\n got: %q\nwant: %q", budget, got, s)
		}
		if strings.Contains(got, legendEllipsis) {
			t.Errorf("budget %d: legend that fits gained an ellipsis: %q", budget, got)
		}
	}
}

// TestClipLegendTruncatesWithMarker pins C2's first half and C1: an over-wide
// legend is cut to the budget and ends in the marker.
func TestClipLegendTruncatesWithMarker(t *testing.T) {
	const s = "night · ~water ♠wood \"forage ^rock ,quarried ᴥden ▲fire △cold ⌂shelter"
	full := ansi.StringWidth(s)
	for _, budget := range []int{10, 20, 33, full - 1} {
		got := clipLegend(s, budget)
		if w := ansi.StringWidth(got); w > budget {
			t.Errorf("budget %d: result is %d columns wide, over budget: %q", budget, w, got)
		}
		if !strings.HasSuffix(got, legendEllipsis) {
			t.Errorf("budget %d: truncated legend does not end in %q: %q", budget, legendEllipsis, got)
		}
	}
}

// TestClipLegendPreservesANSI pins C5. The legend is always styleDim.Render'd,
// so a naive rune slice would sever an escape sequence and bleed styling into
// every row below it. Truncating a styled string must leave it renderable and
// must still respect the budget in DISPLAY columns, not bytes.
func TestClipLegendPreservesANSI(t *testing.T) {
	// Built from literal escapes rather than lipgloss: under `go test` there is
	// no TTY, so lipgloss degrades to the Ascii profile and emits no escapes at
	// all — which would silently skip the very property this test exists to
	// pin. In a real terminal the legend is always styled.
	const (
		bold  = "\x1b[1m"
		reset = "\x1b[0m"
	)
	styled := bold + "night · ~water ♠wood \"forage ^rock ,quarried" + reset

	// Escapes must not count toward the budget.
	if w := ansi.StringWidth(styled); w != ansi.StringWidth("night · ~water ♠wood \"forage ^rock ,quarried") {
		t.Fatalf("precondition: escapes leaked into the width measurement (%d)", w)
	}

	for _, budget := range []int{5, 20, 40} {
		got := clipLegend(styled, budget)
		if w := ansi.StringWidth(got); w > budget {
			t.Errorf("budget %d: styled legend is %d display columns wide: %q", budget, w, got)
		}
		// A severed escape leaves a dangling introducer with no final byte.
		if i := strings.LastIndex(got, "\x1b"); i >= 0 {
			tail := got[i:]
			if !strings.ContainsAny(tail, "mKHJ") {
				t.Errorf("budget %d: truncation severed an ANSI escape, dangling tail %q in %q",
					budget, tail, got)
			}
		}
	}
}

// TestClipLegendMeasuresDisplayColumns pins FR-007. Rune count understates the
// width of double-width glyphs; measuring in runes is exactly how the original
// audit produced false negatives.
func TestClipLegendMeasuresDisplayColumns(t *testing.T) {
	// Emoji are double-width: 8 runes, 16 display columns.
	const wide = "⚡⚡⚡⚡⚡⚡⚡⚡"
	if rw, dw := len([]rune(wide)), ansi.StringWidth(wide); rw == dw {
		t.Skipf("environment measures %q as %d columns for %d runes; no wide-char divergence to test", wide, dw, rw)
	}
	got := clipLegend(wide, 10)
	if w := ansi.StringWidth(got); w > 10 {
		t.Errorf("wide-glyph legend truncated to %d display columns, want <= 10: %q", w, got)
	}
}

// TestClipLegendDegenerateBudgets pins the spec's Edge Cases: a starved resize
// must degrade, never panic and never return a negative-length slice.
func TestClipLegendDegenerateBudgets(t *testing.T) {
	const s = "night · ~water ♠wood"
	for _, budget := range []int{-10, -1, 0, 1, 2, 3} {
		got := clipLegend(s, budget) // must not panic
		if budget < 1 {
			if got != "" {
				t.Errorf("budget %d: want empty, got %q", budget, got)
			}
			continue
		}
		if w := ansi.StringWidth(got); w > budget {
			t.Errorf("budget %d: result is %d columns wide, over budget: %q", budget, w, got)
		}
	}
	if got := clipLegend("", 0); got != "" {
		t.Errorf("empty legend must stay empty, got %q", got)
	}
}

// TestNarrowLegendClampedToTerminalWidth is the regression test for the defect
// itself (FR-002, US1). Before spec 114 the narrow fallback appended the legend
// OUTSIDE the box with no clip at all, so a ~355-column line reached an
// 80-column terminal, soft-wrapped into roughly five rows, and pushed the map
// off the top of the screen.
func TestNarrowLegendClampedToTerminalWidth(t *testing.T) {
	for _, width := range []int{80, 100, 112, 160} {
		m := testModel(t)
		m.width = width
		view := m.mapView()
		for i, line := range strings.Split(view, "\n") {
			if w := ansi.StringWidth(line); w > width {
				t.Errorf("width %d: line %d is %d columns wide (over by %d): %q",
					width, i+1, w, w-width, line)
			}
		}
	}
}

// TestNarrowLegendStaysOneRow pins C4/FR-008. Wrapping to a second row is not
// an accepted remedy for overflow: it would trade a horizontal violation for a
// vertical one, and the map panel's rows are the scarcer resource.
func TestNarrowLegendStaysOneRow(t *testing.T) {
	m := testModel(t)
	m.width = 80
	before := strings.Count(m.mapView(), "\n")

	// Force the legend to its widest by putting piles and chests in view — the
	// inspection content that grows the line most (data-model.md segments 7-8).
	cx, cy := m.gameMap.W/2, m.gameMap.H/2
	m.replica.Piles = []sim.Pile{{X: cx, Y: cy, Wood: 3, Stone: 1}}
	m.replica.Structures = []sim.Structure{
		{Kind: "chest", X: cx + 1, Y: cy, Owner: 1, Store: &sim.Inventory{Wood: 3, Planks: 2, FoodRaw: 5}},
	}
	if after := strings.Count(m.mapView(), "\n"); after != before {
		t.Errorf("legend grew the view from %d rows to %d; it must stay exactly one row", before+1, after+1)
	}
}

// TestLegendMonotonicInWidth pins C3/FR-005: widening the terminal never shows
// less legend. This is what forbids a fixed cap that would satisfy C1 and C2
// while ignoring the space actually available.
func TestLegendMonotonicInWidth(t *testing.T) {
	widths := []int{80, 100, 112, 113, 140, 160, 200}
	prev := -1
	prevWidth := 0
	for _, width := range widths {
		m := testModel(t)
		m.width = width
		vw, vh := m.narrowViewport()
		_, legend := m.renderMapGrid(vw, vh)
		got := ansi.StringWidth(clipLegend(legend, width))
		if got < prev {
			t.Errorf("legend shrank from %d columns at width %d to %d columns at width %d",
				prev, prevWidth, got, width)
		}
		prev, prevWidth = got, width
	}
}

// TestLegendNoFalseEllipsisWhenEverythingFits is C2's second half at the render
// level rather than the helper level: given a terminal wide enough for the
// entire legend, no marker may appear.
func TestLegendNoFalseEllipsisWhenEverythingFits(t *testing.T) {
	m := testModel(t)
	vw, vh := m.narrowViewport()
	_, legend := m.renderMapGrid(vw, vh)
	roomy := ansi.StringWidth(legend) + 50
	if got := clipLegend(legend, roomy); strings.Contains(got, legendEllipsis) {
		t.Errorf("legend gained an ellipsis at width %d where it fits in %d: %q",
			roomy, ansi.StringWidth(legend), got)
	}
}
