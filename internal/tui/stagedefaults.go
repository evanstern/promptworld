package tui

// Stage-shaped TUI layout defaults (reorientation decision 3, spec 066,
// TASK-128). The authority for every per-surface, per-stage default-
// visibility value is docs/design/tui/patterns/stage-defaults.md — this file
// mirrors that page's "Per-surface stage defaults" table cell-for-cell
// (contracts/stage-defaults-table.md) and resolves it into the starting
// visible set the rest of the package composes with the existing fold order
// (layout.go's rowBudget/computeRows, unchanged by this feature).
//
// The parity sweep test (stagedefaults_test.go, TestStageDefaultsSweep,
// the digest.go TestCatalogSweep precedent) parses the authority page at
// test time and asserts equality with stageDefaultsTable below: a default
// changes on the page first, here second, or the build breaks.

import (
	"strings"
	"time"
)

// Governed surface ids — one per row of the authority table, in the page's
// own row order (T001).
const (
	surfaceLessonRow           = "lesson-row"
	surfaceGuardianStrip       = "guardian-strip"
	surfaceVillagerStrip       = "villager-strip"
	surfaceExerciseTab         = "exercise-tab"
	surfaceIncidentVocabulary  = "incident-vocabulary"
	surfaceSystemsTab          = "systems-tab"
	surfaceGuardianConsole     = "guardian-console"
	surfaceHelpGuardianSection = "help-guardian-section"
	surfaceCeremony            = "ceremony"
	surfacePostmortem          = "postmortem"
)

// stageColumn indexes the authority page's six columns, left to right.
type stageColumn int

const (
	colStage1 stageColumn = iota
	colStage2
	colStage3
	colStage4
	colPreLadder
	colNarrow
	numStageColumns
)

// stageDefaultRow is one governed surface's six column values, held in the
// authority page's own cell vocabulary, already normalized (markdown
// emphasis/backticks/ditto stripped and expanded — contracts/
// stage-defaults-table.md §4: normalizes formatting, never meaning). The
// sweep test parses the page fresh and applies the identical normalization
// before comparing, so this array is the page's prose, verbatim.
type stageDefaultRow struct {
	id      string
	columns [numStageColumns]string
}

// stageDefaultsTable mirrors docs/design/tui/patterns/stage-defaults.md's
// "Per-surface stage defaults" table, row for row, column for column
// (contracts/stage-defaults-table.md). Do not edit a value here to make a
// test pass — the page governs; amend it first (spec 047 gate), then this
// table, in the same commit.
var stageDefaultsTable = []stageDefaultRow{
	{surfaceLessonRow, [numStageColumns]string{
		"on",
		"on",
		"badge + overlay-only",
		"badge + overlay-only",
		"badge + overlay-only",
		"same as widescreen (carried, patterns/layout.md R3)",
	}},
	{surfaceGuardianStrip, [numStageColumns]string{
		"on",
		"on",
		"on",
		"on",
		"on",
		"on (carried, R3)",
	}},
	{surfaceVillagerStrip, [numStageColumns]string{
		"on",
		"on",
		"on",
		"on",
		"on",
		"off (folds to header count badge, R3 — never carried)",
	}},
	{surfaceExerciseTab, [numStageColumns]string{
		"present iff the world carries a scenario",
		"present iff the world carries a scenario",
		"present iff the world carries a scenario",
		"present iff the world carries a scenario",
		"present iff the world carries a scenario",
		"present iff the world carries a scenario (solo-view only, R3)",
	}},
	{surfaceIncidentVocabulary, [numStageColumns]string{
		"forecast",
		"forecast",
		"fog",
		"fog",
		"forecast (everything)",
		"same as widescreen",
	}},
	{surfaceSystemsTab, [numStageColumns]string{
		"on",
		"on",
		"on",
		"on",
		"on",
		"on",
	}},
	{surfaceGuardianConsole, [numStageColumns]string{
		"reachable (own key)",
		"reachable",
		"reachable",
		"reachable",
		"reachable",
		"reachable",
	}},
	{surfaceHelpGuardianSection, [numStageColumns]string{
		"shows stage 1's content",
		"shows stage 2's content",
		"shows stage 3's content",
		"shows stage 4's content",
		"shows the pre-ladder (all-verbs) variant",
		"unaffected by width",
	}},
	{surfaceCeremony, [numStageColumns]string{
		"fires stages 1→2, 2→3, 3→4",
		"fires stages 1→2, 2→3, 3→4",
		"fires stages 1→2, 2→3, 3→4 (3→4 only)",
		"never (stage 4 is terminal — nothing unlocks past it)",
		"never (no stage progression exists)",
		"fires identically (takeovers are layout-independent, R3)",
	}},
	{surfacePostmortem, [numStageColumns]string{
		"fires on run.ended, every world",
		"fires on run.ended, every world",
		"fires on run.ended, every world",
		"fires on run.ended, every world",
		"fires on run.ended, every world (ambient/pre-ladder worlds still get the takeover — FR-018's ambient ruling governs its content, not whether it fires)",
		"fires identically (R3)",
	}},
}

// cellFor returns the normalized table cell for id at col, "" if id isn't a
// governed surface (edge case: a row with no built surface is inert, never
// an error — callers that don't recognize an id simply get the zero value).
func cellFor(id string, col stageColumn) string {
	for _, row := range stageDefaultsTable {
		if row.id == id {
			return row.columns[col]
		}
	}
	return ""
}

// stageColumnFor selects the authority table's column for a world's stage
// value: one of the four numbered stages, or the Pre-ladder column for ""
// AND any unrecognized value — the fail-open rule (research.md R3, FR-003):
// an unrecognized stage never takes a narrower posture than pre-ladder.
func stageColumnFor(stage string) stageColumn {
	switch stage {
	case "stage-1":
		return colStage1
	case "stage-2":
		return colStage2
	case "stage-3":
		return colStage3
	case "stage-4":
		return colStage4
	default:
		return colPreLadder
	}
}

// startingVisibleSet is the resolved per-surface starting posture for one
// world at one moment (data-model.md "startingVisibleSet") — input to the
// existing fold pipeline (rowBudget), tab presence, and help-overlay
// section selection. Never consulted by capability machinery (FR-007).
type startingVisibleSet struct {
	// LessonRowOn is the lesson row's stage default: true ("on", eligible
	// for its full two-line form) or false ("badge + overlay-only", starts
	// folded to the header's quiet badge). Whether the row actually shows
	// this frame also depends on a lesson being active (wantsLessonRow,
	// views.go) — this field is only the STAGE default half.
	LessonRowOn bool

	// GuardianStripOn, VillagerStripOn, SystemsTabOn are always true per
	// the authority table (every column is "on") — carried here for
	// completeness and the parity sweep rather than gated anywhere today:
	// the guardian strip is unconditional (decision 7), the systems tab is
	// unconditional (D10), and the villager strip's own surface doesn't
	// exist yet (TASK-129 — an inert row, spec edge case: absence is
	// tolerated, never an error).
	GuardianStripOn bool
	VillagerStripOn bool
	SystemsTabOn    bool

	// ExerciseTabOn is world-shaped, not stage-shaped (FR-006): mirrors the
	// caller's hasScenario input independent of the resolved stage column.
	// The real presence gate stays exerciseID() != "" (tui.go); this field
	// mirrors that same fact for a single consolidated view of the
	// starting set (SC-001 frame assertions).
	ExerciseTabOn bool

	// IncidentVocabulary is the stage-keyed default word ("forecast" or
	// "fog") — the table's own stage-only half of internal/sim's
	// IncidentVisibilityFor (which also honors a per-definition override,
	// out of this function's scope). The real render call site
	// (internal/tui/exercise.go) keeps calling IncidentVisibilityFor
	// directly, since it alone has the ExerciseDefinition an override
	// needs; this field exists for a single consolidated view of the
	// starting set (SC-001) and is asserted consistent with it by
	// TestIncidentVocabularyMatchesSim (stagedefaults_test.go).
	IncidentVocabulary string

	// GuardianConsoleReachable is always true (every column says
	// "reachable") — carried for completeness/parity; the console's 'G'
	// key is unconditional in tui.go today.
	GuardianConsoleReachable bool

	// HelpGuardianVariant names which stage's guardian-section content
	// WOULD show (D9) — "stage-1".."stage-4" or "pre-ladder". The content
	// itself is a separate, not-yet-authored deliverable (docs/design/tui/
	// overlays/help.md lists it "unbuilt (wave 4)"); this field only
	// carries which variant the resolved column selects, for when that
	// content lands.
	HelpGuardianVariant string
}

// stageResolve resolves the current model's starting visible set (with any
// session overrides layered on top) — a single consolidated read for
// frame-level assertions and any future call site that wants one fact
// rather than composing currentStage()/exerciseID() by hand. No production
// tab/row gate is rerouted through this today (see layout.go's
// lessonRowDefault for the one gate this feature DID consolidate); the
// method exists so a single source is available without rewiring every
// existing exerciseID()-gated call site, which already implements this
// same rule correctly (FR-006).
func (m Model) stageResolve() startingVisibleSet {
	set := resolveStageDefaults(m.currentStage(), m.exerciseID() != "")
	return applyOverrides(set, m.stageOverrides)
}

// resolveStageDefaults derives the starting visible set for one world at
// one moment (data-model.md "resolve(stage, world-shape)"): stage=="" or
// any unrecognized value takes the Pre-ladder column (fail-open, R3);
// hasScenario resolves the world-shaped exercise-tab axis independently of
// stage (FR-006).
func resolveStageDefaults(stage string, hasScenario bool) startingVisibleSet {
	col := stageColumnFor(stage)
	return startingVisibleSet{
		LessonRowOn:              cellFor(surfaceLessonRow, col) == "on",
		GuardianStripOn:          cellFor(surfaceGuardianStrip, col) == "on",
		VillagerStripOn:          cellFor(surfaceVillagerStrip, col) == "on",
		SystemsTabOn:             cellFor(surfaceSystemsTab, col) == "on",
		ExerciseTabOn:            hasScenario,
		IncidentVocabulary:       incidentVocabularyColumn(col),
		GuardianConsoleReachable: strings.HasPrefix(cellFor(surfaceGuardianConsole, col), "reachable"),
		HelpGuardianVariant:      helpGuardianVariantFor(col),
	}
}

// incidentVocabularyColumn reduces the table's prose cell ("forecast",
// "fog", or "forecast (everything)" for the pre-ladder "everything"
// posture) to the closed two-word vocabulary internal/sim.VisibilityForecast/
// VisibilityFog use — a prefix match keeps this table-driven rather than a
// second hardcoded stage switch (FR-001).
func incidentVocabularyColumn(col stageColumn) string {
	v := cellFor(surfaceIncidentVocabulary, col)
	if strings.HasPrefix(v, "fog") {
		return "fog"
	}
	return "forecast"
}

// helpGuardianVariantFor names which stage's guardian-section content the
// resolved column selects (D9) — the pre-ladder column is the "all-verbs"
// variant, every numbered column is its own stage's content.
func helpGuardianVariantFor(col stageColumn) string {
	switch col {
	case colStage1:
		return "stage-1"
	case colStage2:
		return "stage-2"
	case colStage3:
		return "stage-3"
	case colStage4:
		return "stage-4"
	default:
		return "pre-ladder"
	}
}

// --- US3 (T011-T013): live re-resolution, explicit overrides, first-
// occurrence arrival announcements ---

// surfaceOverrides records an in-session, per-surface visibility choice the
// player made explicitly (data-model.md "surfaceOverrides"): an override
// outranks re-resolution until the session ends and is never persisted
// (spec Assumptions). Keyed by the governed surface id constants above.
type surfaceOverrides map[string]bool

// applyOverrides layers ov on top of a freshly resolved set: only the
// boolean surfaces an override can name are affected, and only when that
// surface actually has an override recorded — every other field of set
// passes through unchanged. (No production caller sets an override today;
// see stagedefaults_test.go's TestApplyOverridesPrecedence for the
// mechanism proven against synthetic input, and the package doc comment
// for the open question this leaves — no in-session UI command exists yet
// to toggle a governed surface.)
func applyOverrides(set startingVisibleSet, ov surfaceOverrides) startingVisibleSet {
	if v, ok := ov[surfaceLessonRow]; ok {
		set.LessonRowOn = v
	}
	if v, ok := ov[surfaceGuardianStrip]; ok {
		set.GuardianStripOn = v
	}
	if v, ok := ov[surfaceVillagerStrip]; ok {
		set.VillagerStripOn = v
	}
	if v, ok := ov[surfaceSystemsTab]; ok {
		set.SystemsTabOn = v
	}
	if v, ok := ov[surfaceExerciseTab]; ok {
		set.ExerciseTabOn = v
	}
	return set
}

// newlyOnSurfaces reports which boolean-gated governed surfaces went from
// off to on between two resolutions, in the fixed order lesson-row,
// guardian-strip, villager-strip, systems-tab — the input to first-
// occurrence arrival announcements (FR-005, SC-005). Exercise-tab is
// deliberately excluded: its presence is world-shaped (FR-006), never a
// stage-driven arrival. Under the CURRENT authority table no numbered-stage
// transition ever produces a true result here (every governed row is
// either constant-on or narrows going up, never widens) — this is
// forward-compatible plumbing, exercised directly by
// TestNewlyOnSurfacesDiff's synthetic fixtures rather than by any live
// stage transition today.
func newlyOnSurfaces(prev, next startingVisibleSet) []string {
	var out []string
	if !prev.LessonRowOn && next.LessonRowOn {
		out = append(out, surfaceLessonRow)
	}
	if !prev.GuardianStripOn && next.GuardianStripOn {
		out = append(out, surfaceGuardianStrip)
	}
	if !prev.VillagerStripOn && next.VillagerStripOn {
		out = append(out, surfaceVillagerStrip)
	}
	if !prev.SystemsTabOn && next.SystemsTabOn {
		out = append(out, surfaceSystemsTab)
	}
	return out
}

// surfaceArrivalLessonID maps a governed surface id to the lessonCatalog
// entry (lessons.go) that should announce its first appearance, if any.
// Empty today: no catalog entry is authored for a surface's stage-driven
// arrival (the current table never produces one — see newlyOnSurfaces),
// so this is an extension point, not live behavior. Wiring a real entry in
// here is a content-authoring decision for whichever future stage-defaults
// revision actually widens a surface, not this feature's to invent.
var surfaceArrivalLessonID = map[string]string{}

// announceSurfaceArrival routes a newly-on surface through the EXISTING
// first-occurrence lesson machinery (lessons.go lessonCatalog, spec 055,
// research.md R5) rather than a second dedupe mechanism: if id has a mapped
// catalog entry, it is offered to the row exactly as any other trigger
// would be (tryPromote-or-enqueue), so the catalog's own seen-map keeps it
// to exactly once (SC-005). A no-op when id has no mapped entry (true for
// every surface today).
func announceSurfaceArrival(lt *lessonTriggers, id string, now time.Time) *lessonEntry {
	lessonID, ok := surfaceArrivalLessonID[id]
	if !ok {
		return nil
	}
	lt.ensureSeen()
	if lt.seen[lessonID] || lt.isPendingOrActive(lessonID) {
		return nil
	}
	for i := range lessonCatalog {
		entry := &lessonCatalog[i]
		if entry.ID != lessonID {
			continue
		}
		if surfaced := lt.tryPromote(entry, now); surfaced != nil {
			return surfaced
		}
		lt.enqueue(entry, now)
		return nil
	}
	return nil
}
