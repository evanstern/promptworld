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
	"github.com/evanstern/promptworld/internal/skin"
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
			ask := skin.Default().Resolve("skin.guardian.example_ask." + v)
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
