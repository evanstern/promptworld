package tui

// The D9 guardian help section (spec 063 US5, T014; SC-005): stage-keyed,
// model-free — byte-identical across repeated renders of identical status,
// listing exactly the stage ceiling's verbs with one skin-resolved example
// ask per verb; nil status renders the pre-ladder variant (all verbs).

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/guardian"
	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/skin"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worlds"
)

// stagedModel returns a model whose status carries the given stage.
func stagedModel(t *testing.T, stage string) Model {
	t.Helper()
	m := testModel(t)
	m.status = &ipc.StatusData{}
	m.status.World.Stage = stage
	return m
}

// TestHelpGuardianByteIdenticalPerStatus (SC-005): for every stage (and the
// nil-status floor), two renders produce byte-identical lines — the no-LLM
// floor invariant extended to the stage-keyed section per the recorded
// spec-045 amendment.
func TestHelpGuardianByteIdenticalPerStatus(t *testing.T) {
	for _, stage := range []string{"", "stage-1", "stage-2", "stage-3", "stage-4"} {
		m := stagedModel(t, stage)
		if stage == "" {
			m.status = nil // the nil-status floor
		}
		a := strings.Join(m.helpGuardianLines(76), "\n")
		b := strings.Join(m.helpGuardianLines(76), "\n")
		if a != b {
			t.Errorf("stage %q: repeated renders differ", stage)
		}
		if a == "" {
			t.Errorf("stage %q: section rendered empty", stage)
		}
	}
	// Identical status values render identical bytes across models too.
	m1, m2 := stagedModel(t, "stage-1"), stagedModel(t, "stage-1")
	if strings.Join(m1.helpGuardianLines(76), "\n") != strings.Join(m2.helpGuardianLines(76), "\n") {
		t.Error("two models with identical status render different guardian sections")
	}
}

// TestHelpGuardianListsCeilingVerbs (SC-005's "exactly the world's effective
// verbs"): the section lists exactly guardian.StageCeilingVerbs(stage) —
// derived through the SAME intersection the turn's grant runs — each with
// its example-ask token resolved (never a raw token path).
func TestHelpGuardianListsCeilingVerbs(t *testing.T) {
	for _, stage := range []string{"stage-1", "stage-4"} {
		m := stagedModel(t, stage)
		body := strings.Join(m.helpGuardianLines(200), "\n")
		verbs := guardian.StageCeilingVerbs(stage)
		for _, v := range verbs {
			if !strings.Contains(body, v) {
				t.Errorf("stage %s: section omits ceiling verb %q", stage, v)
			}
			ask := skin.Default().ExampleAsk(v)
			if strings.HasPrefix(ask, "skin.") {
				t.Errorf("verb %q has no example-ask token in the default table", v)
			}
			if !strings.Contains(body, ask) {
				t.Errorf("stage %s: section omits %q's example ask %q", stage, v, ask)
			}
		}
	}
	// The stage-1 section teaches no beyond-ceiling verb.
	m := stagedModel(t, "stage-1")
	body := strings.Join(m.helpGuardianLines(200), "\n")
	for _, beyond := range []string{"work_miracle", "adjust_speed"} {
		if strings.Contains(body, beyond) {
			t.Errorf("stage-1 section teaches beyond-ceiling verb %q", beyond)
		}
	}
	// Stage identity renders skin-resolved.
	if !strings.Contains(body, "The Voice") {
		t.Error("stage-1 section omits the skin stage identity")
	}
	if !strings.Contains(body, "conversational prompting") {
		t.Error("stage-1 section omits the ladder concept")
	}
}

// TestHelpGuardianNilStatusPreLadder: nil status renders the pre-ladder
// variant — all verbs, no lock, never blank (the overlays/help.md rule).
func TestHelpGuardianNilStatusPreLadder(t *testing.T) {
	m := testModel(t)
	m.status = nil
	body := strings.Join(m.helpGuardianLines(200), "\n")
	if !strings.Contains(body, "pre-ladder") {
		t.Error("nil-status section does not name the pre-ladder posture")
	}
	for _, v := range guardian.StageCeilingVerbs("") {
		if !strings.Contains(body, v) {
			t.Errorf("pre-ladder section omits verb %q", v)
		}
	}
}

// TestHelpGuardianSectionReachable: tab cycles reach the fourth section and
// its title resolves the skin epithet.
func TestHelpGuardianSectionReachable(t *testing.T) {
	m := stagedModel(t, "stage-1")
	m.helpOpen = true
	m.helpSection = helpSectionKeys
	for i := 0; i < int(helpSectionCount); i++ {
		if m.helpSection == helpSectionGuardian {
			break
		}
		m.helpSection = (m.helpSection + 1) % helpSectionCount
	}
	if m.helpSection != helpSectionGuardian {
		t.Fatal("section cycle never reaches the guardian section")
	}
	if got := m.helpSectionLabel(helpSectionGuardian); got != "the guardian" {
		t.Errorf("guardian section label = %q, want the skin epithet form", got)
	}
	view := m.helpPanelView(100, 24)
	if !strings.Contains(view, "HELP · the guardian") {
		t.Error("panel title does not carry the guardian section label")
	}
}

// --- spec 078: the forward-ladder block (reorient decision 6) ---

// ladderFixtureModel is the testModel(t) precedent (tui_test.go), but with a
// fixture unlocks-record entry written to the isolated PROMPTWORLD_HOME
// BEFORE New() runs — so the boot-loaded m.unlocks snapshot (spec 078
// FR-006) actually carries it, the way a real player's earned record would.
func ladderFixtureModel(t *testing.T, earnedStage string, entry worlds.UnlockEntry) Model {
	t.Helper()
	t.Setenv("PROMPTWORLD_HOME", t.TempDir()+"/home")
	worlds.UpsertUnlock(earnedStage, entry)
	w, err := world.Create(t.TempDir()+"/w", "test", 42)
	if err != nil {
		t.Fatal(err)
	}
	m := New(w)
	m.replica = sim.NewState(42, w.Map())
	m.width, m.height = 80, 30
	return m
}

// TestHelpLadderMatchesStagesJSONSubstrate is the deliverable's proof (spec
// 078 FR-005/SC-001): expected rows are computed at RUNTIME from the same
// substrate `stages --json` marshals (world.StageOrder × world.StagesLadder
// × worlds.Unlocks.StageEarned × skin.Stage) against a fixture unlocks
// record, then every field is asserted present in the rendered ladder block
// — zero hardcoded stage ids, counts, or catalog prose in this test body,
// so a TASK-151 catalog change flows through both surfaces and this test
// untouched (the spec's TASK-151 armor).
func TestHelpLadderMatchesStagesJSONSubstrate(t *testing.T) {
	// Fixture a record-earned stage that is NOT the unconditional floor —
	// world.Stage1 is always earned regardless of the record, so exercising
	// the audit-pointer path needs some other stage from the substrate's
	// own order.
	var fixtureStage string
	for _, id := range world.StageOrder {
		if id != world.Stage1 {
			fixtureStage = id
			break
		}
	}
	if fixtureStage == "" {
		t.Skip("substrate names no stage beyond the unconditional floor — nothing to fixture an audit pointer for")
	}

	entry := worlds.UnlockEntry{
		World: "fixture-proving-world", Path: "/worlds/fixture-proving-world",
		Exercise: "fixture-exercise", EarnedAt: "2026-07-25T18:00:00Z",
	}
	m := ladderFixtureModel(t, fixtureStage, entry)

	lines := m.helpGuardianLines(200)
	ladderStart := -1
	for i, l := range lines {
		if strings.Contains(l, "The ladder") {
			ladderStart = i
			break
		}
	}
	if ladderStart < 0 {
		t.Fatal("guardian section is missing its forward-ladder block")
	}
	ladder := lines[ladderStart:]

	// Locate each stage's row by its skin-resolved identity line, IN
	// world.StageOrder's own order (the production iteration order) — this
	// derives row boundaries from the substrate itself, never a literal
	// stage list.
	rowStart := make([]int, len(world.StageOrder))
	for i, id := range world.StageOrder {
		si, ok := skin.Stage(id)
		if !ok {
			t.Fatalf("default skin has no identity for substrate stage %q", id)
		}
		idx := -1
		for j, l := range ladder {
			if strings.Contains(l, si.Name) && strings.Contains(l, "("+id+")") {
				idx = j
				break
			}
		}
		if idx < 0 {
			t.Fatalf("ladder block is missing stage %q's identity row (name %q)", id, si.Name)
		}
		rowStart[i] = idx
	}

	nextID := ""
	for _, id := range world.StageOrder {
		if !m.unlocks.StageEarned(id) {
			nextID = id
			break
		}
	}

	for i, id := range world.StageOrder {
		end := len(ladder)
		if i+1 < len(rowStart) {
			end = rowStart[i+1]
		}
		row := strings.Join(ladder[rowStart[i]:end], "\n")
		info := world.StagesLadder[id]

		if !strings.Contains(row, info.Concept) {
			t.Errorf("stage %q row missing its substrate concept %q:\n%s", id, info.Concept, row)
		}
		if info.UnlockEvidence != "" {
			if !strings.Contains(row, info.UnlockEvidence) {
				t.Errorf("stage %q row missing its substrate unlock evidence %q:\n%s", id, info.UnlockEvidence, row)
			}
		} else if !strings.Contains(row, "graduation") {
			t.Errorf("stage %q (empty unlock evidence, the substrate's graduation case) row missing graduation wording:\n%s", id, row)
		}

		switch id {
		case fixtureStage:
			if !strings.Contains(row, entry.World) || !strings.Contains(row, entry.Exercise) {
				t.Errorf("record-earned stage %q row missing its audit pointer (want world=%q exercise=%q):\n%s",
					id, entry.World, entry.Exercise, row)
			}
		case world.Stage1:
			if !strings.Contains(row, "earned") {
				t.Errorf("unconditional-floor stage %q row missing an earned marker:\n%s", id, row)
			}
		case nextID:
			if !strings.Contains(row, "next") {
				t.Errorf("first-unearned stage %q row missing the next marker:\n%s", id, row)
			}
		}
	}
}

// TestHelpLadderByteIdenticalForFixedInputs (spec 078 FR-010): extends the
// SC-005 guarantee to the ladder's own inputs — for a fixed (stage,
// override flag, unlocks snapshot, replica.StagesUnlocked, skin) tuple, the
// block's bytes are constant across repeated renders; overridden vs
// non-overridden are DIFFERENT fixed inputs (plan D5) whose renders need
// not (and should not) match each other.
func TestHelpLadderByteIdenticalForFixedInputs(t *testing.T) {
	m := stagedModel(t, world.Stage1)
	m.status.World.StageOverridden = true
	m.replica.StagesUnlocked = []string{world.Stage1}

	a := strings.Join(m.helpGuardianLines(76), "\n")
	b := strings.Join(m.helpGuardianLines(76), "\n")
	if a != b {
		t.Error("ladder block differs across repeated renders of identical (stage, override, unlocks, replica, skin) inputs")
	}

	m2 := stagedModel(t, world.Stage1)
	m2.replica.StagesUnlocked = []string{world.Stage1}
	if strings.Join(m2.helpGuardianLines(76), "\n") == a {
		t.Error("overridden and non-overridden renders must differ — the override marker is part of the fixed-input tuple")
	}
}

// TestHelpLadderNilInputsRenderFloor (spec 078 FR-008): nil status, nil
// replica, and no unlocks file on disk (an unresolvable/never-written
// record) must still render a non-empty, honest floor ladder — never a
// panic, never a blank block (TestHelpContentReadsNoStatusOrReplica's
// construction check extended to the ladder's own inputs).
func TestHelpLadderNilInputsRenderFloor(t *testing.T) {
	t.Setenv("PROMPTWORLD_HOME", t.TempDir()+"/home") // never written to — LoadUnlocks degrades to empty
	w, err := world.Create(t.TempDir()+"/w", "test", 42)
	if err != nil {
		t.Fatal(err)
	}
	m := New(w)
	m.status = nil
	m.replica = nil

	body := strings.Join(m.helpGuardianLines(200), "\n")
	if body == "" {
		t.Fatal("nil status/replica/unlocks-file rendered an empty guardian section")
	}
	if !strings.Contains(body, "The ladder") {
		t.Fatal("nil status/replica/unlocks-file dropped the ladder block entirely")
	}
	for _, id := range world.StageOrder {
		if id == world.Stage1 {
			continue
		}
		if m.unlocks.StageEarned(id) {
			t.Errorf("stage %q reports earned with no unlocks file ever written — expected nothing beyond the floor", id)
		}
	}
	if !strings.Contains(body, "every player's floor") {
		t.Error("floor ladder does not honestly state the unconditional-floor stage's earned reason")
	}
}

// TestHelpLadderMidSessionUnlockShowsEarnedWithoutAuditPointer (spec 078
// FR-002/FR-006, edge case 3): a stage latched in replica.StagesUnlocked
// mid-session — before (or regardless of whether) the per-user record write
// lands — shows earned immediately (no client restart, no per-frame disk
// read), but WITHOUT an audit pointer, since the record is the only place
// that pointer exists.
func TestHelpLadderMidSessionUnlockShowsEarnedWithoutAuditPointer(t *testing.T) {
	var fixtureStage string
	for _, id := range world.StageOrder {
		if id != world.Stage1 {
			fixtureStage = id
			break
		}
	}
	if fixtureStage == "" {
		t.Skip("substrate names no stage beyond the unconditional floor")
	}

	m := testModel(t) // isolated PROMPTWORLD_HOME, no unlocks record written
	m.replica.StagesUnlocked = append(m.replica.StagesUnlocked, fixtureStage)

	if m.unlocks.StageEarned(fixtureStage) {
		t.Fatal("test bug: the fixture stage must not be record-earned")
	}

	lines := m.helpGuardianLines(200)
	ladderStart := -1
	for i, l := range lines {
		if strings.Contains(l, "The ladder") {
			ladderStart = i
			break
		}
	}
	if ladderStart < 0 {
		t.Fatal("guardian section is missing its forward-ladder block")
	}
	ladder := lines[ladderStart:]
	si, ok := skin.Stage(fixtureStage)
	if !ok {
		t.Fatalf("default skin has no identity for %q", fixtureStage)
	}
	rowIdx := -1
	for i, l := range ladder {
		if strings.Contains(l, si.Name) && strings.Contains(l, "("+fixtureStage+")") {
			rowIdx = i
			break
		}
	}
	if rowIdx < 0 {
		t.Fatalf("ladder block is missing stage %q's identity row", fixtureStage)
	}
	end := rowIdx + 5
	if end > len(ladder) {
		end = len(ladder)
	}
	row := strings.Join(ladder[rowIdx:end], "\n")
	if !strings.Contains(row, "earned") {
		t.Errorf("replica-only mid-session unlock did not render as earned:\n%s", row)
	}
	if strings.Contains(row, "proven in") {
		t.Errorf("replica-only mid-session unlock rendered an audit pointer before the record entry exists:\n%s", row)
	}
}

// TestHelpLadderOverrideAnnotatesWithoutLaunderingEarned (spec 078 FR-007,
// edge case "stage-overridden world"): the you-are-here marker states the
// override honestly, but the row's earned state stays record-derived — an
// override is NEVER laundered into an earned claim.
func TestHelpLadderOverrideAnnotatesWithoutLaunderingEarned(t *testing.T) {
	var fixtureStage string
	for _, id := range world.StageOrder {
		if id != world.Stage1 {
			fixtureStage = id
			break
		}
	}
	if fixtureStage == "" {
		t.Skip("substrate names no stage beyond the unconditional floor")
	}

	m := stagedModel(t, fixtureStage)
	m.status.World.StageOverridden = true

	if m.unlocks.StageEarned(fixtureStage) {
		t.Fatal("test bug: the overridden stage must not be record-earned")
	}
	body := strings.Join(m.helpGuardianLines(200), "\n")
	if !strings.Contains(body, "by override — not earned") {
		t.Error("overridden world's you-are-here marker does not name the override honestly")
	}
	if strings.Contains(body, "proven in") {
		t.Error("an overridden, unearned stage must never render an audit pointer")
	}
}

// TestHelpLadderScrollsAt80x24 (spec 078 FR-009, the
// TestHelpWalkthroughScrollsAt80x24 precedent): the ladder's content grows
// several lines per stage, but pages fully through the shared pager at a
// small pane height — every stage's identity surfaces somewhere in the
// scroll, usable at 80×24.
func TestHelpLadderScrollsAt80x24(t *testing.T) {
	m := testModel(t)
	m.helpOpen = true
	m.helpSection = helpSectionGuardian
	full := m.helpGuardianLines(76)

	var seen []string
	for scroll := 0; scroll < 200; scroll++ {
		page := paginateHelpContent(full, scroll, 6)
		seen = append(seen, page...)
		if !strings.Contains(strings.Join(page, "\n"), "J to scroll") {
			break
		}
	}
	joined := strings.Join(seen, "\n")
	if !strings.Contains(joined, "The ladder") {
		t.Error("paging never surfaced the ladder header")
	}
	for _, id := range world.StageOrder {
		si, ok := skin.Stage(id)
		if !ok {
			continue
		}
		if !strings.Contains(joined, si.Name) {
			t.Errorf("paging never surfaced stage %q's identity (full content length %d lines)", id, len(full))
		}
	}
	if v := m.View(); v == "" {
		t.Error("narrow overlay view rendered empty at 80x30")
	}
}
