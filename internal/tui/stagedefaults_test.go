package tui

// Stage-shaped TUI layout defaults (spec 066, TASK-128): the authority-page
// parity sweep (contracts/stage-defaults-table.md, T004), resolve() unit
// tests (T005), the pre-feature pre-ladder golden-frame baseline (T002,
// SC-002), per-stage frame assertions (T008, SC-001), the reachability
// sweep (T009, SC-003), and US3's live-arrival/override plumbing (T011-014).

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/ipc"
)

// --- T004: authority-page parity sweep (contracts/stage-defaults-table.md,
// the digest.go TestCatalogSweep precedent) ---

// normalizeStageCell strips markdown emphasis/code-span decoration and
// collapses whitespace — formatting only, never meaning (contract §4).
func normalizeStageCell(s string) string {
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "*", "")
	return strings.Join(strings.Fields(s), " ")
}

// surfaceIDByDisplayName maps the authority page's first-column display
// name (its "(`ref`)"/"(D9)" annotation stripped) to the governed surface
// id constants (stagedefaults.go).
var surfaceIDByDisplayName = map[string]string{
	"Lesson row":                     surfaceLessonRow,
	"Guardian strip":                 surfaceGuardianStrip,
	"Villager strip":                 surfaceVillagerStrip,
	"Exercise tab":                   surfaceExerciseTab,
	"Incident-visibility vocabulary": surfaceIncidentVocabulary,
	"Systems tab":                    surfaceSystemsTab,
	"Guardian console":               surfaceGuardianConsole,
	"Help overlay guardian section":  surfaceHelpGuardianSection,
	"Unlock ceremony":                surfaceCeremony,
	"Postmortem":                     surfacePostmortem,
}

// splitTableRow splits one "| a | b | c |" markdown row into trimmed cells.
func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// stageDefaultsDisplayName extracts a first-column cell's plain surface
// name, dropping its trailing "(`ref`)"/"(D9)" annotation.
func stageDefaultsDisplayName(cell string) string {
	cell = strings.TrimSpace(cell)
	if i := strings.Index(cell, " ("); i >= 0 {
		return cell[:i]
	}
	return cell
}

// parseStageDefaultsPage parses the authority page's "Per-surface stage
// defaults" table into surface id -> six normalized column values,
// expanding the page's "″" ditto mark (repeats the cell immediately to its
// left in the same row, optionally with a suffix — e.g. "″ (3→4 only)")
// before normalizing.
func parseStageDefaultsPage(doc string) map[string][numStageColumns]string {
	out := map[string][numStageColumns]string{}
	inTable := false
	for _, line := range strings.Split(doc, "\n") {
		trimmed := strings.TrimSpace(line)
		if !inTable {
			if strings.HasPrefix(trimmed, "| Surface |") {
				inTable = true
			}
			continue
		}
		if !strings.HasPrefix(trimmed, "|") {
			break // table ended
		}
		if strings.HasPrefix(trimmed, "|---") || strings.HasPrefix(trimmed, "| ---") {
			continue // header separator row
		}
		cells := splitTableRow(trimmed)
		if len(cells) != 1+int(numStageColumns) {
			continue
		}
		name := stageDefaultsDisplayName(cells[0])
		id, ok := surfaceIDByDisplayName[name]
		if !ok {
			continue // reported separately below (unmapped page row)
		}
		var cols [numStageColumns]string
		prev := ""
		for i := 0; i < int(numStageColumns); i++ {
			raw := strings.TrimSpace(cells[i+1])
			switch {
			case raw == "″":
				raw = prev
			case strings.HasPrefix(raw, "″"):
				suffix := strings.TrimSpace(strings.TrimPrefix(raw, "″"))
				raw = strings.TrimSpace(prev + " " + suffix)
			}
			prev = raw
			cols[i] = normalizeStageCell(raw)
		}
		out[id] = cols
	}
	return out
}

// TestStageDefaultsSweep is the contract's enforcer: the code table
// (stagedefaults.go) and the authority page must carry the same surface
// rows with the same six normalized column values in both directions — a
// default value change is a two-file change or this fails (contracts/
// stage-defaults-table.md).
func TestStageDefaultsSweep(t *testing.T) {
	doc, err := os.ReadFile("../../docs/design/tui/patterns/stage-defaults.md")
	if err != nil {
		t.Fatalf("reading docs/design/tui/patterns/stage-defaults.md: %v", err)
	}
	parsed := parseStageDefaultsPage(string(doc))

	seen := map[string]bool{}
	for _, row := range stageDefaultsTable {
		seen[row.id] = true
		pageCols, ok := parsed[row.id]
		if !ok {
			t.Errorf("code table carries surface %q but the authority page has no matching row", row.id)
			continue
		}
		for i := 0; i < int(numStageColumns); i++ {
			codeVal := normalizeStageCell(row.columns[i])
			if codeVal != pageCols[i] {
				t.Errorf("surface %q column %d: code table = %q, authority page = %q", row.id, i, codeVal, pageCols[i])
			}
		}
	}
	for id := range parsed {
		if !seen[id] {
			t.Errorf("authority page carries surface %q but the code table has no matching row", id)
		}
	}
}

// --- T002 (Foundational, must predate wiring): pre-ladder golden-frame
// corpus — the byte-identity baseline for SC-002/US2. Each entry hashes
// m.View() (sha256 hex) rather than embedding full frame text, so a single
// differing byte fails loudly without a multi-KB literal living in this
// file. Captured against the PRE-WIRING behavior (T003-T007 land after this
// test exists) and asserted unchanged for the rest of this feature's life —
// a hash that changes here means pre-ladder rendering regressed.

func frameHash(v string) string {
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}

// preLadderGoldenFrame is one entry in the corpus: build builds a
// pre-ladder Model (world.Stage/status.World.Stage left at their zero
// value, "") in a specific representative state, then View() is hashed.
type preLadderGoldenFrame struct {
	name  string
	want  string
	build func(t *testing.T) Model
}

var preLadderGoldenFrames = []preLadderGoldenFrame{
	{
		name: "widescreen-home",
		want: "f0327b1a9b7b633a5026c1ffb9bf6d761634d616564160f4725f50bcf1e0a9be",
		build: func(t *testing.T) Model {
			m := widescreenModel(t)
			seedEvents(&m, 20)
			return m
		},
	},
	{
		name: "narrow-home",
		// Re-pinned for spec 077 (TASK-151): the map legend legitimately grew
		// the appended "S stranger" token (tiles.go, the marsh/sand FR-009
		// precedent) — this narrow frame renders the legend line, so its hash
		// moved with the sanctioned vocabulary growth, nothing else
		// (TestTilesIdentityPin still guards the frozen prefix bytes).
		want: "0676b22e12e59fc683e6f7cdd8495c0c932fbfac87cc629e0b2ffc8f5d4006a1",
		build: func(t *testing.T) Model {
			m := testModel(t)
			seedEvents(&m, 20)
			return m
		},
	},
	{
		name: "widescreen-villagers-solo",
		want: "0e8d1b93efef4760ea36a8ecab97138122c3ef22433fe00247c83649833bb8c0",
		build: func(t *testing.T) Model {
			m := widescreenModel(t)
			m.dockTab = paneVillagers
			m.solo = true
			return m
		},
	},
	{
		name: "widescreen-guardian-strip-charges",
		want: "9eb5939e2bd54c496bac8bf1f02cab1fda38ff793a849592e3a47c71e98b220b",
		build: func(t *testing.T) Model {
			m := widescreenModel(t)
			m.connected = true
			m.status = &ipc.StatusData{Clock: ipc.ClockStatus{GuardianCharges: 2, Tick: 100}}
			return m
		},
	},
	{
		name: "widescreen-active-lesson-badge-only",
		want: "72fd8b525fb0e71b743167644a314e0871407e747c42f94cf49b843d6564517f",
		build: func(t *testing.T) Model {
			// Pre-ladder's lesson-row default is badge+overlay-only (the
			// table's own Pre-ladder column, matching the pre-055 "off"
			// posture) — an active lesson still renders only the header
			// badge, never the two-line row, at height 40.
			m := widescreenModel(t)
			m.applyEvent(lessonFixtureEvent(t, "first-death"))
			return m
		},
	},
	{
		name: "widescreen-scenario-exercise-tab",
		want: "3731790b637cdbfc51f9f5b023f1cddad96c5f51776b14620a7741b4bcc6e926",
		build: func(t *testing.T) Model {
			m := scenarioModel(t)
			m.w.Manifest.Stage = "" // this corpus is pre-ladder even though scenarioModel defaults to stage-1
			m.exBriefingDismissed = true
			return m
		},
	},
	{
		name: "help-overlay-open",
		want: "7e234c6e33ba9d43dbb2fe999496459ae4f88b9ef6a1714f7c53c5893f3c79a2",
		build: func(t *testing.T) Model {
			m := widescreenModel(t)
			m.helpOpen = true
			m.helpPageMode = helpModeGlobal
			m.helpSection = helpSectionKeys
			return m
		},
	},
}

// TestPreLadderGoldenFrames is the T002/SC-002 baseline: every entry's
// rendered frame must byte-hash exactly as it did before spec 066's
// stage-defaults RESOLUTION wiring landed — that refactor (resolve() +
// its call sites) must not move a single byte of pre-ladder rendering. If
// THIS refactor ever needs a new "want" value, the fix is in the wiring,
// never here.
//
// Re-pinned for spec 060 (TASK-129): the villager strip is a genuinely new
// chrome row (village lens completion, not a stage-defaults refactor), so
// every widescreen frame below legitimately gained a row (or, for the
// narrow/solo/help-overlay frames, the strip's absence still changed the
// header via the new `[N villagers]` badge). That is the intended, spec-
// sanctioned visual change this feature shipped — verified directly by
// villagerStripView's own tests (village_lens_test.go) — not a regression
// this baseline exists to catch. The hashes below are the new baseline;
// the "never re-pin for a resolution-only refactor" rule above still holds
// for spec 066's own wiring.
//
// Re-pinned again for spec 068 (TASK-143), five map-bearing frames only:
// these fixtures create NEW worlds (world.Create now stamps terrain_gen 2),
// whose maps legitimately carry the marsh ░ / sand ▒ vocabulary, and the
// map legend grew the two matching tokens (FR-009); the scenario frame's fixed-seed map likewise. Pre-existing-vocabulary
// byte identity is separately guaranteed by spec 068's own pin
// (TestTilesIdentityPin, a LEGACY-generation fixture) — this corpus's four
// new hashes reflect the sanctioned new-world visual change, nothing else.
func TestPreLadderGoldenFrames(t *testing.T) {
	for _, fx := range preLadderGoldenFrames {
		t.Run(fx.name, func(t *testing.T) {
			m := fx.build(t)
			if stage := m.currentStage(); stage != "" {
				t.Fatalf("fixture %q is not pre-ladder: currentStage() = %q", fx.name, stage)
			}
			got := frameHash(m.View())
			if got != fx.want {
				t.Errorf("%s: View() hash = %s, want %s (pre-ladder byte-identity regressed)", fx.name, got, fx.want)
			}
		})
	}
}

// --- T005: resolve() unit tests — all four stages, pre-ladder, unrecognized
// stage (fail-open, R3), scenario x stage independence (FR-006) ---

// TestResolveStageDefaultsPerStage sweeps every stage against the table's
// own columns (SC-001's unit-level twin — the full-frame version lives in
// TestLessonRowShownAtStage1And2/TestLessonBadgeAtStage3AndPreLadder,
// lessons_test.go, and TestStageDefaultsFrameMatchesResolve below).
func TestResolveStageDefaultsPerStage(t *testing.T) {
	cases := []struct {
		stage           string
		wantLessonOn    bool
		wantIncidentVoc string
	}{
		{"stage-1", true, "forecast"},
		{"stage-2", true, "forecast"},
		{"stage-3", false, "fog"},
		{"stage-4", false, "fog"},
		{"", false, "forecast"}, // pre-ladder
	}
	for _, c := range cases {
		t.Run("stage_"+c.stage, func(t *testing.T) {
			set := resolveStageDefaults(c.stage, false)
			if set.LessonRowOn != c.wantLessonOn {
				t.Errorf("stage %q: LessonRowOn = %v, want %v", c.stage, set.LessonRowOn, c.wantLessonOn)
			}
			if set.IncidentVocabulary != c.wantIncidentVoc {
				t.Errorf("stage %q: IncidentVocabulary = %q, want %q", c.stage, set.IncidentVocabulary, c.wantIncidentVoc)
			}
			if !set.GuardianStripOn || !set.VillagerStripOn || !set.SystemsTabOn || !set.GuardianConsoleReachable {
				t.Errorf("stage %q: always-on surfaces must all be true, got %+v", c.stage, set)
			}
		})
	}
}

// TestResolveStageDefaultsUnrecognizedStageFailsOpen (FR-003, research.md
// R3): a stage value that isn't one of the four ladder stages must take
// exactly the pre-ladder posture — never narrower — including a value that
// merely LOOKS like a stage id (e.g. a future "stage-5").
func TestResolveStageDefaultsUnrecognizedStageFailsOpen(t *testing.T) {
	preLadder := resolveStageDefaults("", false)
	for _, stage := range []string{"stage-5", "bogus", "Stage-1", " stage-1"} {
		got := resolveStageDefaults(stage, false)
		if got != preLadder {
			t.Errorf("resolveStageDefaults(%q, false) = %+v, want the pre-ladder posture %+v", stage, got, preLadder)
		}
	}
}

// TestResolveStageDefaultsExerciseTabIndependentOfStage (FR-006): the
// world-shaped exercise-tab axis mirrors hasScenario alone, at every stage
// including pre-ladder — never gated by the resolved stage column.
func TestResolveStageDefaultsExerciseTabIndependentOfStage(t *testing.T) {
	for _, stage := range []string{"", "stage-1", "stage-2", "stage-3", "stage-4", "unrecognized"} {
		if got := resolveStageDefaults(stage, true); !got.ExerciseTabOn {
			t.Errorf("stage %q, hasScenario=true: ExerciseTabOn = false, want true", stage)
		}
		if got := resolveStageDefaults(stage, false); got.ExerciseTabOn {
			t.Errorf("stage %q, hasScenario=false: ExerciseTabOn = true, want false", stage)
		}
	}
}

// TestResolveStageDefaultsHelpGuardianVariant (D9): the resolved column
// names the stage whose guardian-section content would show — the content
// itself is a separate, not-yet-authored deliverable (docs/design/tui/
// overlays/help.md: "unbuilt (wave 4)"); this only proves which variant the
// resolution selects.
func TestResolveStageDefaultsHelpGuardianVariant(t *testing.T) {
	cases := map[string]string{
		"stage-1": "stage-1",
		"stage-2": "stage-2",
		"stage-3": "stage-3",
		"stage-4": "stage-4",
		"":        "pre-ladder",
		"bogus":   "pre-ladder",
	}
	for stage, want := range cases {
		if got := resolveStageDefaults(stage, false).HelpGuardianVariant; got != want {
			t.Errorf("stage %q: HelpGuardianVariant = %q, want %q", stage, got, want)
		}
	}
}

// TestStageDefaultsFrameMatchesResolve (T008, SC-001): the lesson row's
// actual rendered posture (full row eligible vs header-badge-only) must
// match resolveStageDefaults' LessonRowOn field at every stage — the frame
// and the table-driven resolution are the same fact, never two.
func TestStageDefaultsFrameMatchesResolve(t *testing.T) {
	for _, stage := range []string{"", "stage-1", "stage-2", "stage-3", "stage-4"} {
		t.Run("stage_"+stage, func(t *testing.T) {
			m := withStage(widescreenModel(t), stage)
			m.applyEvent(lessonFixtureEvent(t, "first-death"))
			view := m.View()
			wantOn := resolveStageDefaults(stage, false).LessonRowOn
			hasFullRow := strings.Contains(view, lessonPullSuffix)
			if hasFullRow != wantOn {
				t.Errorf("stage %q: full lesson row rendered = %v, want %v (resolveStageDefaults.LessonRowOn)", stage, hasFullRow, wantOn)
			}
		})
	}
}

// --- T009: reachability sweep (SC-003, FR-002) — every governed surface
// stays reachable, with stage-independent full content, at every stage ---

// TestReachabilitySweepDockTabsStageIndependent: the dock-tab row and its
// keybindings (villagers/systems/guardian) never vary by stage — reaching
// them is identical regardless of the world's curriculum stage.
func TestReachabilitySweepDockTabsStageIndependent(t *testing.T) {
	var rows []string
	for _, stage := range []string{"", "stage-1", "stage-2", "stage-3", "stage-4"} {
		m := withStage(widescreenModel(t), stage)
		rows = append(rows, m.dockTabsRow())
	}
	for i := 1; i < len(rows); i++ {
		if rows[i] != rows[0] {
			t.Errorf("dockTabsRow differs by stage: %q vs %q", rows[0], rows[i])
		}
	}
}

// TestReachabilitySweepGuardianConsoleAlwaysOpens (guardian-console row:
// "reachable" at every stage): the 'G' key opens the console regardless of
// stage, with content that doesn't vary by stage (spec 046 doctrine, FR-007
// — capability gating is untouched by this feature; the console's own
// content is guardian.Status-driven, not stage-defaults-driven).
func TestReachabilitySweepGuardianConsoleAlwaysOpens(t *testing.T) {
	for _, stage := range []string{"", "stage-1", "stage-2", "stage-3", "stage-4"} {
		m := withStage(widescreenModel(t), stage)
		m = pressKey(t, m, "G")
		if !m.console {
			t.Errorf("stage %q: G should open the guardian console", stage)
		}
	}
}

// TestReachabilitySweepExerciseTabReachableWhenScenario (exercise-tab row:
// world-shaped presence, FR-006): key 6 reaches it at every stage on a
// scenario world, and its content (the panel body) does not vary by stage
// beyond the incident vocabulary the table itself governs.
func TestReachabilitySweepExerciseTabReachableWhenScenario(t *testing.T) {
	for _, stage := range []string{"", "stage-1", "stage-2", "stage-3", "stage-4"} {
		m := scenarioModel(t)
		m.w.Manifest.Stage = stage
		m.exBriefingDismissed = true
		if !strings.Contains(m.dockTabsRow(), "exercise") {
			t.Errorf("stage %q: exercise tab must be reachable on a scenario world", stage)
		}
		m2 := pressKey(t, m, "6")
		if m2.dockTab != paneExercise {
			t.Errorf("stage %q: key 6 should select the exercise tab", stage)
		}
	}
}

// TestReachabilitySweepHelpOverlayStageIndependent (D9 aside: the guardian
// section's own content variance is out of this feature's scope, see
// TestResolveStageDefaultsHelpGuardianVariant) — every OTHER help section
// (keys/walkthrough/lessons) must render identically regardless of stage,
// since defaults shape placement, never reachable content (FR-002). The
// header's own top line is excluded from the comparison: its "[lesson]"
// badge is stage-shaped BY DESIGN (spec 055, already shipped and covered
// by TestLessonBadgeAtStage3AndPreLadder) — a fact about the lesson-row
// surface, not the help overlay's content.
func TestReachabilitySweepHelpOverlayStageIndependent(t *testing.T) {
	var bodies []string
	for _, stage := range []string{"", "stage-1", "stage-2", "stage-3", "stage-4"} {
		m := withStage(widescreenModel(t), stage)
		m.helpOpen = true
		m.helpPageMode = helpModeGlobal
		m.helpSection = helpSectionKeys
		lines := strings.SplitN(m.View(), "\n", 2)
		if len(lines) != 2 {
			t.Fatalf("stage %q: expected a header line plus body", stage)
		}
		bodies = append(bodies, lines[1])
	}
	for i := 1; i < len(bodies); i++ {
		if bodies[i] != bodies[0] {
			t.Errorf("help overlay (keys section) body differs by stage")
		}
	}
}

// --- T012: session-only override precedence ---

// TestApplyOverridesPrecedence: an override outranks the freshly resolved
// value regardless of which way the default would have gone; a surface
// with no recorded override passes the resolved value through unchanged.
// No production caller sets an override today (no in-session command exists
// to toggle a governed surface) — this proves the mechanism against direct
// synthetic input.
func TestApplyOverridesPrecedence(t *testing.T) {
	base := resolveStageDefaults("stage-3", false) // LessonRowOn == false here
	if base.LessonRowOn {
		t.Fatal("fixture assumption broken: stage-3 lesson row should default off")
	}
	overridden := applyOverrides(base, surfaceOverrides{surfaceLessonRow: true})
	if !overridden.LessonRowOn {
		t.Error("an explicit override must outrank the resolved default")
	}
	// Everything else passes through untouched.
	if overridden.GuardianStripOn != base.GuardianStripOn || overridden.IncidentVocabulary != base.IncidentVocabulary {
		t.Errorf("applyOverrides touched a field it wasn't given an override for: %+v vs base %+v", overridden, base)
	}
	// No override recorded at all: identity.
	if got := applyOverrides(base, nil); got != base {
		t.Errorf("applyOverrides(base, nil) = %+v, want base %+v unchanged", got, base)
	}
}

// --- T013: newly-on-surface diff (forward-compatible plumbing — see
// stagedefaults.go's newlyOnSurfaces doc comment: no numbered-stage
// transition in the CURRENT table ever produces a real result here, since
// every governed row is constant-on or narrows going up, never widens) ---

func TestNewlyOnSurfacesDiff(t *testing.T) {
	prev := startingVisibleSet{}
	next := startingVisibleSet{LessonRowOn: true, GuardianStripOn: true}
	got := newlyOnSurfaces(prev, next)
	want := []string{surfaceLessonRow, surfaceGuardianStrip}
	if len(got) != len(want) {
		t.Fatalf("newlyOnSurfaces = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("newlyOnSurfaces[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	// Nothing changed, or a surface narrowed (on->off): no arrivals.
	if got := newlyOnSurfaces(next, next); len(got) != 0 {
		t.Errorf("identical resolutions must report no arrivals, got %v", got)
	}
	narrowed := startingVisibleSet{LessonRowOn: false}
	if got := newlyOnSurfaces(next, narrowed); len(got) != 0 {
		t.Errorf("a surface narrowing (on->off) must not report an arrival, got %v", got)
	}
	// ExerciseTabOn is world-shaped, never a stage arrival (FR-006) — even
	// a synthetic off->on transition on that field alone must not surface.
	exOnly := startingVisibleSet{ExerciseTabOn: true}
	if got := newlyOnSurfaces(startingVisibleSet{}, exOnly); len(got) != 0 {
		t.Errorf("ExerciseTabOn must never be treated as a stage arrival, got %v", got)
	}
}

// --- T014: fold-pressure composition (edge case: a newly-on row on a
// short terminal enters through the normal fold order only, never forcing
// body below bodyMin) ---

// TestStageDefaultsComposeWithFoldOrder: at a stage whose default starts
// the lesson row "on" (stage-1/2), a short terminal folds it through
// patterns/layout.md ruling (a) exactly as any other on-by-default row
// would — body never dips below bodyMin while a fold step remains
// available, and never negative once every foldable row is exhausted. This
// is the same arithmetic TestLessonRowFoldsBeforeGuardianStripUnderHeight
// Pressure (lessons_test.go) already proves for one stage; here it sweeps
// every stage's resolved starting posture to confirm stage-defaults never
// produces a narrower fold order than layout.md states (patterns/
// stage-defaults.md "Composition with the fold order").
func TestStageDefaultsComposeWithFoldOrder(t *testing.T) {
	for _, stage := range []string{"", "stage-1", "stage-2", "stage-3", "stage-4"} {
		t.Run("stage_"+stage, func(t *testing.T) {
			wantsLesson := resolveStageDefaults(stage, false).LessonRowOn
			for _, h := range []int{9, 10, 14, 15, 16, 17, 18, 20, 30, 40, 60} {
				rows := computeRows(h, wantsLesson)
				if rows.Body < 0 {
					t.Fatalf("height %d: Body went negative: %+v", h, rows)
				}
				sum := rows.Header + rows.VillagerStrip + rows.Lesson + rows.Strip + rows.Body + rows.Minibuffer + rows.Footer
				if rows.Body > 0 && sum != h {
					t.Errorf("height %d: rows don't sum to total: %+v (sum %d)", h, rows, sum)
				}
				if rows.Lesson != 0 && rows.Lesson != lessonRowRows {
					t.Errorf("height %d: Lesson = %d, want 0 or %d", h, rows.Lesson, lessonRowRows)
				}
			}
		})
	}
}

// TestModelStageResolveConsolidatesCurrentStageAndOverrides: the model
// method matches resolveStageDefaults(m.currentStage(), hasScenario) with
// overrides layered on top — one consolidated read.
func TestModelStageResolveConsolidatesCurrentStageAndOverrides(t *testing.T) {
	m := withStage(widescreenModel(t), "stage-3")
	want := resolveStageDefaults("stage-3", false)
	if got := m.stageResolve(); got != want {
		t.Errorf("stageResolve() = %+v, want %+v", got, want)
	}
	m.stageOverrides = surfaceOverrides{surfaceLessonRow: true}
	if got := m.stageResolve(); !got.LessonRowOn {
		t.Error("stageResolve() must reflect an explicit override")
	}
}

// TestSurfaceIDByDisplayNameCoversAllGovernedIDs guards the sweep test
// itself: every governed surface id constant must have a display-name
// mapping, or a page row could silently fail to match (falling through
// "unmapped page row" rather than being compared at all).
func TestSurfaceIDByDisplayNameCoversAllGovernedIDs(t *testing.T) {
	want := map[string]bool{}
	for _, row := range stageDefaultsTable {
		want[row.id] = true
	}
	got := map[string]bool{}
	for _, id := range surfaceIDByDisplayName {
		got[id] = true
	}
	for id := range want {
		if !got[id] {
			t.Errorf("surface id %q has no entry in surfaceIDByDisplayName", id)
		}
	}
}
