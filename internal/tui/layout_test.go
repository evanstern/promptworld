package tui

import "testing"

// TestIsWidescreen is the layout.md breakpoint: >=112 widescreen, <112 narrow.
func TestIsWidescreen(t *testing.T) {
	cases := []struct {
		width int
		want  bool
	}{
		{80, false},
		{111, false},
		{112, true},
		{113, true},
		{200, true},
	}
	for _, c := range cases {
		if got := isWidescreen(c.width); got != c.want {
			t.Errorf("isWidescreen(%d) = %v, want %v", c.width, got, c.want)
		}
	}
}

// TestComputeColumns is layout.md's "Column budget": map and dock split
// 50/50 (minus the 1-col gutter), with the map taking the odd leftover
// column when (total - gutter) is odd.
func TestComputeColumns(t *testing.T) {
	cases := []struct {
		total    int
		wantDock int
		wantMap  int
	}{
		{200, (200 - 1) / 2, 200 - 1 - (200-1)/2},
		{118, (118 - 1) / 2, 118 - 1 - (118-1)/2},
		{112, 55, 56}, // (112-1)=111 odd: dock floors to 55, map takes the extra column
		{113, 56, 56}, // 112 even: splits exactly
		{100, (100 - 1) / 2, 100 - 1 - (100-1)/2},
	}
	for _, c := range cases {
		got := computeColumns(c.total)
		if got.DockCols != c.wantDock {
			t.Errorf("computeColumns(%d).DockCols = %d, want %d", c.total, got.DockCols, c.wantDock)
		}
		if got.MapCols != c.wantMap {
			t.Errorf("computeColumns(%d).MapCols = %d, want %d", c.total, got.MapCols, c.wantMap)
		}
		if got.MapCols+got.Gutter+got.DockCols != c.total {
			t.Errorf("computeColumns(%d) columns don't sum to total: %+v", c.total, got)
		}
		if got.MapCols < got.DockCols {
			t.Errorf("computeColumns(%d): map (%d) should never be smaller than dock (%d) — map takes the odd column",
				c.total, got.MapCols, got.DockCols)
		}
	}
}

// TestComputeRows is layout.md's re-derived "Row budget": header 1, guardian
// strip 1 (spec 050), minibuffer 3, footer 1, body takes the remainder —
// never negative, and the strip is the last row to fold (ruling a step 4),
// so a total roomy enough for bodyMin (10) keeps the strip on.
func TestComputeRows(t *testing.T) {
	cases := []struct {
		total     int
		wantStrip int
		wantBody  int
	}{
		{40, 1, 40 - 1 - 1 - 3 - 1},
		{30, 1, 30 - 1 - 1 - 3 - 1},
		{16, 1, 10}, // exactly at the fold threshold: strip stays, body == bodyMin
		{15, 0, 10}, // one row short: strip folds, reclaiming the row keeps body == bodyMin
		{14, 0, 9},  // below the floor: no more foldable rows, body dips under bodyMin
		{3, 0, 0},   // starved: body floors at 0, never negative
	}
	for _, c := range cases {
		got := computeRows(c.total)
		if got.Header != 1 || got.Minibuffer != 3 || got.Footer != 1 {
			t.Errorf("computeRows(%d) chrome rows wrong: %+v", c.total, got)
		}
		if got.Strip != c.wantStrip {
			t.Errorf("computeRows(%d).Strip = %d, want %d", c.total, got.Strip, c.wantStrip)
		}
		if got.Body != c.wantBody {
			t.Errorf("computeRows(%d).Body = %d, want %d", c.total, got.Body, c.wantBody)
		}
		if got.Body < 0 {
			t.Errorf("computeRows(%d).Body went negative: %d", c.total, got.Body)
		}
	}
}

// TestComputeRowsInvariant sweeps a wide range of terminal heights and
// checks data-model.md's rowBudget invariant: Header+Strip+Body+Minibuffer+
// Footer == totalRows whenever totalRows covers the fixed chrome (i.e. the
// body hasn't floored at 0), and the strip is present whenever the body
// would otherwise be >= bodyMin (SC-003: fold order + row arithmetic must
// match the design reference at every height).
func TestComputeRowsInvariant(t *testing.T) {
	for total := 0; total <= 60; total++ {
		got := computeRows(total)
		sum := got.Header + got.Strip + got.Body + got.Minibuffer + got.Footer
		if got.Body > 0 && sum != total {
			t.Errorf("computeRows(%d) rows don't sum to total: %+v (sum %d)", total, got, sum)
		}
		if got.Strip != 0 && got.Strip != 1 {
			t.Errorf("computeRows(%d).Strip = %d, want 0 or 1", total, got.Strip)
		}
		// Never both folded AND roomy: if there was room for the strip
		// (body would be >= bodyMin with it present), the strip must be on.
		fixed := headerRows + minibufferRows + footerRows
		if bodyWithStrip := total - fixed - stripRows; bodyWithStrip >= bodyMin && got.Strip == 0 {
			t.Errorf("computeRows(%d): room for the strip (body would be %d >= bodyMin) but Strip folded", total, bodyWithStrip)
		}
	}
}

// TestMapViewportTiles: 2 terminal columns per tile, minus border+padding
// (B1: styleBox's Padding(0,1) eats 2 more columns beyond the border, a
// budget that was originally missed) and the legend row; never below 1x1.
func TestMapViewportTiles(t *testing.T) {
	w, h := mapViewportTiles(74, 30)
	if w != (74-4)/2 {
		t.Errorf("tilesW = %d, want %d", w, (74-4)/2)
	}
	if h != 30-3 {
		t.Errorf("tilesH = %d, want %d", h, 30-3)
	}
	if w2, h2 := mapViewportTiles(1, 1); w2 < 1 || h2 < 1 {
		t.Errorf("tiny panel must floor at 1x1, got %dx%d", w2, h2)
	}
}
