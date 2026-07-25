package tui

// Layout math for the widescreen composite (TASK-34). Pure functions over
// terminal width/height — see docs/design/tui/patterns/layout.md. No panel
// measures the terminal itself; Update computes budgets once per
// tea.WindowSizeMsg and View hands each region its exact (width, height).

// widescreenBreakpoint is the width at/above which the composite home page
// (map ‖ dock) renders instead of today's single-pane narrow fallback.
const widescreenBreakpoint = 112

// Column budget constants (widescreen). The map and dock split the
// terminal 50/50 (patterns/layout.md "Column budget") — the map takes the
// odd leftover column when (width - gutter) doesn't divide evenly.
const gutterCols = 1

// Row budget constants (widescreen).
const (
	headerRows     = 1
	minibufferRows = 3
	footerRows     = 1
	// stripRows is the guardian strip's row cost when visible (spec 050;
	// patterns/layout.md re-derived row budget) — exactly 1, borderless.
	stripRows = 1
	// bodyMin is the fold threshold (patterns/layout.md ruling a): chrome
	// folds, one step at a time, until body >= bodyMin or the floor is
	// reached. The guardian strip is the only foldable row this feature
	// implements (the villager strip and lesson row are later waves), so it
	// folds at exactly this threshold — layout.md's total fold order names
	// it step 4, the last step, because decision 7 says the budget is
	// always visible; folding relocates it (minibufferView's dormant line)
	// rather than hiding it.
	bodyMin = 10
)

// isWidescreen reports whether width is enough for the composite home page.
func isWidescreen(width int) bool { return width >= widescreenBreakpoint }

// columnBudget is the widescreen composite's horizontal split.
type columnBudget struct {
	MapCols  int
	Gutter   int
	DockCols int
}

// computeColumns splits totalCols 50/50 between the map and the dock (a
// planning decision superseding the earlier fixed-44-col dock — see
// docs/design/tui/patterns/layout.md "Column budget"); the map takes the
// odd column when (totalCols - gutter) is odd.
func computeColumns(totalCols int) columnBudget {
	avail := totalCols - gutterCols
	if avail < 0 {
		avail = 0
	}
	dock := avail / 2
	mapCols := avail - dock
	return columnBudget{MapCols: mapCols, Gutter: gutterCols, DockCols: dock}
}

// rowBudget is the widescreen composite's vertical split.
type rowBudget struct {
	Header     int
	Strip      int // 0 or 1 — 0 once folded (patterns/layout.md ruling a step 4)
	Body       int
	Minibuffer int
	Footer     int
}

// computeRows splits totalRows between the fixed-height chrome (header,
// guardian strip, minibuffer, footer) and the body (map ‖ dock), which
// takes whatever is left. The guardian strip is the last chrome to fold
// (patterns/layout.md ruling a step 4): it stays visible as long as folding
// it isn't needed to keep the body at or above bodyMin, and is preferred
// over letting the body dip below that floor. Below the floor (no more
// foldable rows in this feature's code), body may still go under bodyMin,
// or even to 0 on a starved resize — the pre-reorientation "existing
// behavior" layout.md's Floor layout section keeps.
func computeRows(totalRows int) rowBudget {
	fixed := headerRows + minibufferRows + footerRows
	if bodyWithStrip := totalRows - fixed - stripRows; bodyWithStrip >= bodyMin {
		return rowBudget{Header: headerRows, Strip: stripRows, Body: bodyWithStrip, Minibuffer: minibufferRows, Footer: footerRows}
	}
	body := totalRows - fixed
	if body < 0 {
		body = 0
	}
	return rowBudget{Header: headerRows, Strip: 0, Body: body, Minibuffer: minibufferRows, Footer: footerRows}
}

// mapViewportTiles converts a panel's (cols, rows) into terrain tiles: 2
// terminal columns per tile (same family as today's vw/vh computation),
// minus room for the panel's border+padding and legend row.
func mapViewportTiles(panelCols, panelRows int) (tilesW, tilesH int) {
	tilesW = (panelCols - 4) / 2 // border (2) + the box style's Padding(0,1) (2) — B1
	if tilesW < 1 {
		tilesW = 1
	}
	tilesH = panelRows - 3 // border + legend row
	if tilesH < 1 {
		tilesH = 1
	}
	return tilesW, tilesH
}
