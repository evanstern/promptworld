package tui

// Mouse-parity sweep (spec 073-mouse-parity-sweep, TASK-154; reorient
// decision 8, "verification culture, extended"): converts the input-parity
// doctrine (docs/design/tui/patterns/keymap.md, "Input-parity doctrine" rule
// 1) from a hand-tracked promise into a gate. Every canonical-header control
// table under docs/design/tui/ is parsed; any non-'—' mouse half in its
// keys+mouse column (contracts/control-table.md) is a *shipped mouse claim*
// that must have a hand-audited oracle entry below whose live tea.MouseMsg
// dispatch proves the handler — both directions (a claim with no oracle
// entry, or an oracle entry the corpus no longer documents, both fail) — and
// every page carrying a keyed-but-mouseless row must carry a
// "**Parity rollout**" note (contract rule 4). Precedent: TestStageDefaultsSweep
// (stagedefaults_test.go) for the corpus-vs-code sweep shape; TestHelpKeymapSweep
// / TestHelpKeymapSweepLiveDispatch (help_test.go) for the oracle-plus-live-dispatch
// shape this feature composes the two into.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// controlTableCorpusRoot is the relative read from this package's directory
// (TestStageDefaultsSweep precedent).
const controlTableCorpusRoot = "../../docs/design/tui"

// canonicalControlTableHeader mirrors CANONICAL_HEADER in
// scripts/check-tui-design.mjs (contracts/control-table.md) byte-for-byte —
// the Go parser and the JS gate drift together or not at all (spec
// Assumptions).
const canonicalControlTableHeader = "| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |"

// mouseKeysCellClass is the keys+mouse column's classification — the
// corpus's cell inventory (spec grounding) has exactly three shapes.
type mouseKeysCellClass int

const (
	classDisplayOnly mouseKeysCellClass = iota
	classTrackedGap
	classShippedClaim
)

// classifyKeysMouse classifies one keys+mouse cell (plan D2). A cell with no
// " · " separator is display-only only when it starts with "—" ("—",
// "— (display-only)", "— (auto-follow)"); anything else with no separator is
// malformed (none exist in the shipped corpus — the contract's grammar
// allows only "—" or "<keys> · <mouse>"). Otherwise the mouse half is the
// LAST " · "-separated segment — dock.md's "solo → home / switch" row nests
// a second " · " inside the keys half itself (1/esc (home) · a different
// tab key (switch) · —) — exactly "—" is a tracked parity gap, anything
// else is a shipped mouse claim. A cell whose FIRST " · "-separated
// segment is bare "—" is also malformed (a display-only marker can't also
// carry a keys half) — defensive, none exist.
func classifyKeysMouse(cell string) (class mouseKeysCellClass, mouseClaim string, ok bool) {
	if !strings.Contains(cell, " · ") {
		if strings.HasPrefix(cell, "—") {
			return classDisplayOnly, "", true
		}
		return 0, "", false
	}
	if strings.TrimSpace(strings.SplitN(cell, " · ", 2)[0]) == "—" {
		return 0, "", false
	}
	idx := strings.LastIndex(cell, " · ")
	mouseHalf := strings.TrimSpace(cell[idx+len(" · "):])
	if mouseHalf == "—" {
		return classTrackedGap, "", true
	}
	return classShippedClaim, mouseHalf, true
}

// controlTableClaim is one shipped mouse claim found in the corpus: the
// oracle key triple (spec FR-002) plus the page it came from.
type controlTableClaim struct {
	page    string // corpus-relative path, e.g. "panels/chronicle.md"
	control string // column 1 (control/region)
	claim   string // mouse half, e.g. "click line"
}

// controlTableParseResult is everything TestMouseParitySweep needs from one
// walk of the corpus: every shipped mouse claim (direction 1's input), the
// corpus-relative pages that parsed >= 1 tracked-gap row (FR-005's
// rollout-note check), and totals for the non-vacuity assertion (FR-006).
type controlTableParseResult struct {
	claims           []controlTableClaim
	pagesNeedingNote map[string]bool
	tableCount       int
}

// parseControlTables walks root (FR-001, FR-006: t.Fatalf on an unreadable
// root rather than passing vacuously — the TestStageDefaultsSweep
// precedent), finds every canonical-header table (mirroring
// parseStageDefaultsPage's inTable state machine), and classifies each row's
// keys+mouse cell. splitTableRow (stagedefaults_test.go) does the `|`
// splitting; its column order is the JS gate's (cols[0] = control/region,
// cols[4] = keys+mouse).
func parseControlTables(t *testing.T, root string) controlTableParseResult {
	t.Helper()
	result := controlTableParseResult{pagesNeedingNote: map[string]bool{}}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		raw, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			return rerr
		}
		rel = filepath.ToSlash(rel)

		inTable := false
		for _, line := range strings.Split(string(raw), "\n") {
			trimmed := strings.TrimSpace(line)
			if !inTable {
				if !strings.HasPrefix(trimmed, "|") {
					continue
				}
				cells := splitTableRow(trimmed)
				if len(cells) != 7 {
					continue
				}
				if "| "+strings.Join(cells, " | ")+" |" != canonicalControlTableHeader {
					continue
				}
				inTable = true
				result.tableCount++
				continue
			}
			if !strings.HasPrefix(trimmed, "|") {
				inTable = false
				continue
			}
			if strings.HasPrefix(trimmed, "|---") || strings.HasPrefix(trimmed, "| ---") {
				continue // header separator row
			}
			cells := splitTableRow(trimmed)
			if len(cells) != 7 {
				inTable = false
				continue
			}
			control := cells[0]
			class, claim, ok := classifyKeysMouse(cells[4])
			if !ok {
				t.Fatalf("%s: malformed keys+mouse cell for control %q: %q", rel, control, cells[4])
			}
			switch class {
			case classTrackedGap:
				result.pagesNeedingNote[rel] = true
			case classShippedClaim:
				result.claims = append(result.claims, controlTableClaim{page: rel, control: control, claim: claim})
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("reading corpus root %s: %v", root, err)
	}
	return result
}

// mouseParityOracleEntry is one hand-audited proof: the corpus triple it
// covers, and an executable live-dispatch check that sends the claimed
// mouse event(s) through Model.Update and asserts the documented effect
// (spec FR-003). An entry may only be added alongside the corpus cell and
// code it proves — keyboard and mouse land together (keymap.md doctrine
// rule 1).
type mouseParityOracleEntry struct {
	page    string
	control string
	claim   string
	check   func(t *testing.T)
}

// mouseParityOracle is the sweep's hand-audited proof table. Today's one
// entry mirrors panels/chronicle.md's jump-to-source row, the corpus's only
// shipped mouse claim.
var mouseParityOracle = []mouseParityOracleEntry{
	{
		page:    "panels/chronicle.md",
		control: "jump-to-source",
		claim:   "click line",
		check:   checkChronicleJumpToSourceMouseClaim,
	},
}

// checkChronicleJumpToSourceMouseClaim proves panels/chronicle.md's "click
// line" claim: a left-release inside a recorded chronicle inspect-list row
// selects that row and applies jumpToSource's effect (camera centers) —
// exactly TestHandleMouseSelectsAndJumps (tui_test.go), the live-dispatch
// precedent this oracle entry reuses rather than duplicates.
func checkChronicleJumpToSourceMouseClaim(t *testing.T) {
	t.Helper()
	m := pausedModel(t) // widescreen, paused, chronicle dock tab, seeded events
	m.chronicleInspectBody(60, 20)
	if !m.chronHit.valid {
		t.Fatal("test setup: chronHit should be valid after rendering the inspect list")
	}
	originX, originY, _ := m.chronicleHitOrigin()
	const row = 1
	if row >= len(m.chronHit.rowEvent) {
		t.Fatalf("test setup: only %d list rows recorded, want at least %d", len(m.chronHit.rowEvent), row+1)
	}
	wantIdx := m.chronHit.rowEvent[row]

	mdl, _ := m.Update(mouseLeftRelease(originX+2, originY+row))
	mm := mdl.(Model)
	if mm.chronSelected != wantIdx {
		t.Fatalf("click line should select event %d, got %d", wantIdx, mm.chronSelected)
	}
	if mm.panX == 0 && mm.panY == 0 {
		t.Error("click line should also apply the jump (jumpToSource: camera should have moved)")
	}
}

// TestMouseParitySweep is the gate (board AC #1, SC-001): parses every
// canonical-header control table in the corpus and enforces, both
// directions, that the oracle above and the corpus's shipped mouse claims
// agree — plus the rollout-note honesty check (FR-005).
func TestMouseParitySweep(t *testing.T) {
	result := parseControlTables(t, controlTableCorpusRoot)

	if result.tableCount == 0 {
		t.Fatal("no canonical-header control tables found under docs/design/tui — corpus missing, moved, or the parser broke")
	}
	if len(result.claims) == 0 {
		t.Fatal("no shipped mouse claims found in the corpus — expected >= 1 (panels/chronicle.md jump-to-source); corpus missing/moved or the parser broke")
	}

	oracleByKey := map[string]mouseParityOracleEntry{}
	for _, e := range mouseParityOracle {
		oracleByKey[e.page+"|"+e.control+"|"+e.claim] = e
	}

	// Direction 1 (FR-002/FR-003): every shipped claim needs a proven oracle
	// entry, and every oracle entry's check actually runs, as a named subtest.
	seenOracleKeys := map[string]bool{}
	for _, claim := range result.claims {
		key := claim.page + "|" + claim.control + "|" + claim.claim
		entry, ok := oracleByKey[key]
		if !ok {
			t.Errorf("%s: control %q claims mouse target %q with no oracle entry proving it — add a mouseParityOracle entry + live-dispatch check (mouseparity_test.go)", claim.page, claim.control, claim.claim)
			continue
		}
		seenOracleKeys[key] = true
		t.Run(claim.page+"/"+claim.control, entry.check)
	}

	// Direction 2 (FR-004): a stale oracle entry — the corpus no longer
	// documents the claim it was written to prove.
	for _, e := range mouseParityOracle {
		key := e.page + "|" + e.control + "|" + e.claim
		if !seenOracleKeys[key] {
			t.Errorf("stale oracle entry: %s / %s claims %q but the corpus no longer documents that mouse target", e.page, e.control, e.claim)
		}
	}

	// FR-005: every page that parsed >= 1 tracked-gap row must carry the
	// "**Parity rollout**" note (contract rule 4's honesty requirement,
	// mechanized).
	for page := range result.pagesNeedingNote {
		raw, err := os.ReadFile(filepath.Join(controlTableCorpusRoot, filepath.FromSlash(page)))
		if err != nil {
			t.Errorf("%s: re-reading for the rollout-note check: %v", page, err)
			continue
		}
		if !strings.Contains(string(raw), "**Parity rollout**") {
			t.Errorf("%s: has a keyed-but-mouseless control but no \"**Parity rollout**\" note (contract rule 4)", page)
		}
	}
}
