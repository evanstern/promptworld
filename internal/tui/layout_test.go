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

// TestComputeRows is layout.md's re-derived "Row budget": header 1, villager
// strip 1 (spec 060), guardian strip 1 (spec 050), minibuffer 3, footer 1,
// body takes the remainder — never negative. The fold cascade (ruling a)
// reclaims the villager strip's row FIRST (step 2), then the guardian
// strip's (step 4; the lesson-row step is a no-op here since these cases
// all pass wantsLesson=false) — so each successive row threshold keeps body
// pinned at exactly bodyMin until the floor (both folded) is reached.
func TestComputeRows(t *testing.T) {
	cases := []struct {
		total             int
		wantVillagerStrip int
		wantStrip         int
		wantBody          int
	}{
		{40, 1, 1, 40 - 1 - 1 - 1 - 3 - 1},
		{30, 1, 1, 30 - 1 - 1 - 1 - 3 - 1},
		{17, 1, 1, 10}, // exactly at the full-chrome fold threshold: both strips stay, body == bodyMin
		{16, 0, 1, 10}, // one row short: the villager strip folds first, body == bodyMin
		{15, 0, 0, 10}, // one row shorter still: the guardian strip folds too, body == bodyMin
		{14, 0, 0, 9},  // below the floor: no more foldable rows, body dips under bodyMin
		{3, 0, 0, 0},   // starved: body floors at 0, never negative
	}
	for _, c := range cases {
		got := computeRows(c.total, false)
		if got.Header != 1 || got.Minibuffer != 3 || got.Footer != 1 {
			t.Errorf("computeRows(%d) chrome rows wrong: %+v", c.total, got)
		}
		if got.VillagerStrip != c.wantVillagerStrip {
			t.Errorf("computeRows(%d).VillagerStrip = %d, want %d", c.total, got.VillagerStrip, c.wantVillagerStrip)
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
// checks data-model.md's rowBudget invariant: Header+VillagerStrip+Strip+
// Body+Minibuffer+Footer == totalRows whenever totalRows covers the fixed
// chrome (i.e. the body hasn't floored at 0), and each strip is present
// whenever the body would otherwise be >= bodyMin (SC-003: fold order + row
// arithmetic must match the design reference at every height).
func TestComputeRowsInvariant(t *testing.T) {
	for total := 0; total <= 60; total++ {
		got := computeRows(total, false)
		sum := got.Header + got.VillagerStrip + got.Strip + got.Body + got.Minibuffer + got.Footer
		if got.Body > 0 && sum != total {
			t.Errorf("computeRows(%d) rows don't sum to total: %+v (sum %d)", total, got, sum)
		}
		if got.VillagerStrip != 0 && got.VillagerStrip != 1 {
			t.Errorf("computeRows(%d).VillagerStrip = %d, want 0 or 1", total, got.VillagerStrip)
		}
		if got.Strip != 0 && got.Strip != 1 {
			t.Errorf("computeRows(%d).Strip = %d, want 0 or 1", total, got.Strip)
		}
		// Never both folded AND roomy: if there was room for both strips
		// (body would be >= bodyMin with everything present), neither may
		// be folded.
		fixed := headerRows + minibufferRows + footerRows
		if bodyFull := total - fixed - villagerStripRows - stripRows; bodyFull >= bodyMin {
			if got.VillagerStrip == 0 {
				t.Errorf("computeRows(%d): room for the villager strip (body would be %d >= bodyMin) but it folded", total, bodyFull)
			}
			if got.Strip == 0 {
				t.Errorf("computeRows(%d): room for the guardian strip (body would be %d >= bodyMin) but it folded", total, bodyFull)
			}
		}
		// The villager strip folds no later than the guardian strip
		// (ruling a: step 2 before step 4) — it must never be the one
		// still showing while the guardian strip has already folded.
		if got.VillagerStrip > 0 && got.Strip == 0 {
			t.Errorf("computeRows(%d): villager strip on but guardian strip folded — violates fold order", total)
		}
	}
}

// TestComputeRowsFoldOrderWithLessonRow (ruling a steps 2-3-4 together):
// with the lesson row also wanted, the villager strip still folds first,
// the lesson row second, the guardian strip last — never out of order.
func TestComputeRowsFoldOrderWithLessonRow(t *testing.T) {
	cases := []struct {
		total                                    int
		wantVillagerStrip, wantLesson, wantStrip int
		wantBody                                 int
	}{
		{19, 1, lessonRowRows, 1, 10}, // everything fits
		{18, 0, lessonRowRows, 1, 10}, // villager strip folds first
		{16, 0, 0, 1, 10},             // lesson row folds next (villager strip stays folded)
		{15, 0, 0, 0, 10},             // guardian strip folds last
		{14, 0, 0, 0, 9},              // below the floor
	}
	for _, c := range cases {
		got := computeRows(c.total, true)
		if got.VillagerStrip != c.wantVillagerStrip || got.Lesson != c.wantLesson || got.Strip != c.wantStrip || got.Body != c.wantBody {
			t.Errorf("computeRows(%d, true) = %+v, want {VillagerStrip:%d Lesson:%d Strip:%d Body:%d}",
				c.total, got, c.wantVillagerStrip, c.wantLesson, c.wantStrip, c.wantBody)
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
