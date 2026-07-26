package guardian

// Spec 084 US1 suites (T025 / SC-005): survey_site byte-identity, the
// no-charge/no-event neutrality, the acting-cardinality exemption (Effect
// Read, structural), the bounds miss, the radius clamp, and the guidance-path
// split (read paragraph, never the acting list).

import (
	"context"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/tool"
	"github.com/evanstern/promptworld/internal/toolloop"
)

// surveyCall drives the handleSurvey handler directly with raw args.
func surveyCall(t *testing.T, mt *Guardian, args string) toolloop.Outcome {
	t.Helper()
	d := &turnDispatch{mt: mt, result: &TurnResult{}, grant: fullGrant()}
	h := mt.handleSurvey(d)
	return h(context.Background(), llm.ToolCall{Name: "survey_site", Args: []byte(args)})
}

// TestSurveyByteIdentityAndNeutrality (SC-005, US1 AS1/AS2): identical
// (args, state) yields byte-identical sheets; no event is appended and the
// charge bank never moves — surveying is looking, not an act.
func TestSurveyByteIdentityAndNeutrality(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	chargesBefore := inj.state.GuardianCharges
	batchesBefore := len(inj.batches)

	a := surveyCall(t, mt, `{"x":10,"y":10,"radius":4}`)
	b := surveyCall(t, mt, `{"x":10,"y":10,"radius":4}`)
	if a.Verdict != toolloop.VerdictReadOK || b.Verdict != toolloop.VerdictReadOK {
		t.Fatalf("verdicts = %v/%v, want read_ok", a.Verdict, b.Verdict)
	}
	if a.ResultForModel != b.ResultForModel {
		t.Errorf("identical calls returned different sheets:\n%s\n---\n%s", a.ResultForModel, b.ResultForModel)
	}
	for _, want := range []string{"Survey of (10,10), radius 4", "- terrain:", "- nearest water:", "- nearest tree:", "- nearest rock:", "- structures within:", "- passability:"} {
		if !strings.Contains(a.ResultForModel, want) {
			t.Errorf("sheet missing %q:\n%s", want, a.ResultForModel)
		}
	}
	if len(inj.batches) != batchesBefore {
		t.Error("a survey injected events")
	}
	if inj.state.GuardianCharges != chargesBefore {
		t.Error("a survey moved the charge bank")
	}
	// The acting-cardinality exemption is structural: Effect Read is what the
	// loop driver's isActing check keys on (the explain precedent).
	if tl, _ := tool.Lookup("survey_site"); tl.Effect != tool.Read || tl.Gate != tool.None {
		t.Errorf("survey_site Effect/Gate = %v/%v, want Read/None", tl.Effect, tl.Gate)
	}
}

// TestSurveyStructuresAndWalls: mirrored structures render kind + tile, and
// standing walls join the passability summary.
func TestSurveyStructuresAndWalls(t *testing.T) {
	mt, _, _, _ := newTestGuardian(t, "behold")
	mt.replica.Structures = append(mt.replica.Structures,
		sim.Structure{Kind: "shelter", X: 10, Y: 10},
		sim.Structure{Kind: "wall_stone", X: 11, Y: 10},
		sim.Structure{Kind: "oven", X: 60, Y: 60}) // outside the box — must not render
	mt.mirrorState()

	sheet := surveyCall(t, mt, `{"x":10,"y":10,"radius":2}`).ResultForModel
	if !strings.Contains(sheet, "shelter at (10,10)") || !strings.Contains(sheet, "wall_stone at (11,10)") {
		t.Errorf("sheet missing structures:\n%s", sheet)
	}
	if strings.Contains(sheet, "oven") {
		t.Errorf("sheet lists a structure outside the radius:\n%s", sheet)
	}
	if !strings.Contains(sheet, "1 further blocked by standing walls") {
		t.Errorf("sheet missing the wall passability clause:\n%s", sheet)
	}
}

// TestSurveyBoundsMiss (US1 AS3): out-of-bounds coordinates return a
// repairable in-fiction miss naming the world bounds — read_ok, never a hard
// error (the explain unknown-topic shape).
func TestSurveyBoundsMiss(t *testing.T) {
	mt, _, _, _ := newTestGuardian(t, "behold")
	out := surveyCall(t, mt, `{"x":999,"y":2}`)
	if out.Verdict != toolloop.VerdictReadOK {
		t.Fatalf("verdict = %v, want read_ok", out.Verdict)
	}
	if !strings.Contains(out.ResultForModel, "the land runs from (0,0) to (63,63)") {
		t.Errorf("miss does not name the bounds: %q", out.ResultForModel)
	}
	// Missing coordinates: an honest repairable ask.
	if got := surveyCall(t, mt, `{}`).ResultForModel; !strings.Contains(got, "needs both x and y") {
		t.Errorf("missing-args reply = %q", got)
	}
}

// TestSurveyRadiusClamp (FR-001): absent radius defaults to 4; out-of-band
// values clamp to 1..8 (clamp-with-honesty: the sheet names the used radius).
func TestSurveyRadiusClamp(t *testing.T) {
	mt, _, _, _ := newTestGuardian(t, "behold")
	if got := surveyCall(t, mt, `{"x":10,"y":10}`).ResultForModel; !strings.Contains(got, "radius 4") {
		t.Errorf("default radius sheet: %q", got)
	}
	if got := surveyCall(t, mt, `{"x":10,"y":10,"radius":99}`).ResultForModel; !strings.Contains(got, "radius 8") {
		t.Errorf("over-cap radius sheet: %q", got)
	}
	if got := surveyCall(t, mt, `{"x":10,"y":10,"radius":0}`).ResultForModel; !strings.Contains(got, "radius 1") {
		t.Errorf("under-cap radius sheet: %q", got)
	}
}

// TestSurveyGuidancePath (US1 AS4): survey_site renders under the read-tools
// paragraph (GuardianReadGuidance) and never under the acting list.
func TestSurveyGuidancePath(t *testing.T) {
	roster := tool.LoopRosterGuardian()
	read := tool.GuardianReadGuidance(roster)
	if !strings.Contains(read, "survey_site(x, y, radius)") {
		t.Errorf("read guidance missing survey_site:\n%s", read)
	}
	acting := tool.GuardianToolGuidance(roster)
	if strings.Contains(acting, "survey_site") {
		t.Errorf("acting guidance lists the read tool:\n%s", acting)
	}
	// The four plan verbs DO render as acting bullets.
	for _, want := range []string{"place_designation", "cancel_designation", "issue_directive", "cancel_directive"} {
		if !strings.Contains(acting, want) {
			t.Errorf("acting guidance missing %s", want)
		}
	}
}
