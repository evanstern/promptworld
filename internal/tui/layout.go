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
	// lessonRowRows is the lesson row's row cost when showing (spec 055,
	// TASK-117; patterns/layout.md re-derived row budget) — exactly 2,
	// borderless (the 2-row budget has no room for a bordered box's own
	// top/bottom rule, panels/lesson-row.md).
	lessonRowRows = 2
	// villagerStripRows is the villager strip's row cost when showing (spec
	// 060, TASK-129; patterns/layout.md re-derived row budget) — exactly 1,
	// borderless, directly under the header (panels/villager-strip.md).
	// Unlike the lesson row it has no stage-off default: stage-defaults.md
	// rules it "on" at every stage in widescreen, so computeRows below never
	// takes a wantsVillagerStrip toggle — the only thing that ever hides it
	// is fold pressure (or narrow, which doesn't call computeRows at all).
	villagerStripRows = 1
	// bodyMin is the fold threshold (patterns/layout.md ruling a): chrome
	// folds, one step at a time, until body >= bodyMin or the floor is
	// reached. This feature's code implements three foldable rows — the
	// villager strip (spec 060), the lesson row (spec 055), and the
	// guardian strip (spec 050) — folding in that order (ruling a: the
	// villager strip folds SECOND, before the lesson row, before the
	// guardian strip; the map legend is body-internal, folding first but
	// outside this function's accounting). The guardian strip folds last
	// because decision 7 says its budget is always visible; folding
	// relocates it (minibufferView's dormant line) rather than hiding it.
	bodyMin = 10
)

// lessonRowDefault reports whether the lesson row's STAGE DEFAULT is "on"
// (patterns/stage-defaults.md): stages 1-2 only — stage 3+ and pre-ladder
// (stage == "", including "no status yet") default to the `[lesson]` header
// badge + overlay-only form instead. Absorbed into the shared stage-defaults
// table (spec 066, TASK-128 — the refactor research.md R6 flagged as this
// function's own future): the table's Lesson row entry (stagedefaults.go)
// is now the single source, resolved the same way every other governed
// surface is (resolveStageDefaults). Behavior is unchanged — this is a
// pure delegation, not a new default.
func lessonRowDefault(stage string) bool {
	return resolveStageDefaults(stage, false).LessonRowOn
}

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
	Header        int
	VillagerStrip int // 0 or 1 — 0 once folded to the header count badge (ruling a step 2, spec 060)
	Lesson        int // 0 or 2 — 0 when not wanted (stage default off / nothing active) or once folded (ruling a step 3, spec 055)
	Strip         int // 0 or 1 — 0 once folded (patterns/layout.md ruling a step 4) — the GUARDIAN strip
	Body          int
	Minibuffer    int
	Footer        int
}

// computeRows splits totalRows between the fixed-height chrome (header,
// villager strip, lesson row, guardian strip, minibuffer, footer) and the
// body (map ‖ dock), which takes whatever is left. wantsLesson is the
// lesson row's own eligibility to occupy its 2-row budget THIS frame — true
// iff the stage default is on (lessonRowDefault) AND a lesson is actually
// active (data-model.md "none" vs "showing": stage-eligible but nothing
// active is still 0 rows, not a blank 2-row block). The villager strip
// carries no such toggle (stage-defaults.md: "on" at every stage in
// widescreen — computeRows is never called at all in narrow, ruling b's
// "not carried" there), so it is always wanted going into the fold cascade.
//
// The three foldable rows reclaim their space in the ruled total order
// (patterns/layout.md ruling a steps 2–4), one step at a time, stopping the
// moment body clears bodyMin: villager strip first, then the lesson row,
// then the guardian strip last (decision 7: its budget is always visible,
// so folding it only relocates it into minibufferView's dormant line rather
// than hiding it). Body may still dip below bodyMin (or to 0 on a starved
// resize) once all three are folded — the pre-reorientation "existing
// behavior" layout.md's Floor layout section keeps. When wantsLesson is
// false, the lesson-row step is a no-op, reducing this to a two-step
// (villager-strip, then guardian-strip) cascade.
func computeRows(totalRows int, wantsLesson bool) rowBudget {
	fixed := headerRows + minibufferRows + footerRows
	lessonWant := 0
	if wantsLesson {
		lessonWant = lessonRowRows
	}
	// Everything visible.
	if body := totalRows - fixed - villagerStripRows - lessonWant - stripRows; body >= bodyMin {
		return rowBudget{Header: headerRows, VillagerStrip: villagerStripRows, Lesson: lessonWant, Strip: stripRows, Body: body, Minibuffer: minibufferRows, Footer: footerRows}
	}
	// Fold step 2 (ruling a): the villager strip reclaims its row first.
	if body := totalRows - fixed - lessonWant - stripRows; body >= bodyMin {
		return rowBudget{Header: headerRows, VillagerStrip: 0, Lesson: lessonWant, Strip: stripRows, Body: body, Minibuffer: minibufferRows, Footer: footerRows}
	}
	// Fold step 3: the lesson row reclaims its rows next, if it was wanted
	// at all — a no-op step when wantsLesson is false.
	if lessonWant > 0 {
		if body := totalRows - fixed - stripRows; body >= bodyMin {
			return rowBudget{Header: headerRows, VillagerStrip: 0, Lesson: 0, Strip: stripRows, Body: body, Minibuffer: minibufferRows, Footer: footerRows}
		}
	}
	// Fold step 4 (guardian strip, spec 050, unchanged): reclaim the strip too.
	body := totalRows - fixed
	if body < 0 {
		body = 0
	}
	return rowBudget{Header: headerRows, VillagerStrip: 0, Lesson: 0, Strip: 0, Body: body, Minibuffer: minibufferRows, Footer: footerRows}
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
