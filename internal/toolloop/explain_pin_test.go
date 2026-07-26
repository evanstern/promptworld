package toolloop

import (
	"reflect"
	"testing"

	"github.com/evanstern/promptworld/internal/tool"
)

// TestExplainDecisionClassesMirrorVerdicts (spec 063 T002): internal/tool's
// explain "decisions" fact sheet mirrors this package's Verdict vocabulary
// (tool is a leaf below toolloop); this pin holds the mirrored name list
// equal to the declared constants, in declaration order, so an explained
// verdict can never drift from the recorded one.
func TestExplainDecisionClassesMirrorVerdicts(t *testing.T) {
	want := []string{
		string(VerdictLanded), string(VerdictLandedClamped),
		string(VerdictRejectedGate), string(VerdictRejectedCardinality),
		string(VerdictRejectedUnknown), string(VerdictRejectedMalformed),
		string(VerdictReadOK), string(VerdictReadError),
		string(VerdictUnlanded),
	}
	if got := tool.DecisionClassNames(); !reflect.DeepEqual(got, want) {
		t.Errorf("tool decision-class mirror = %v, want the toolloop Verdict vocabulary %v", got, want)
	}
}
