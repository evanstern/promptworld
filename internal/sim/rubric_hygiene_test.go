package sim

import (
	"strings"
	"testing"
)

// TestRubricHygieneNoTutorLaneTerms (spec 063 T006, FR-003 / US2 AS-2): no
// exercise rubric term may reference tutor-lane telemetry — the sweep that
// keeps future exercises from quietly grading tutoring. The tutor lane's
// only trail is:
//
//   - the `cog.*` observability channel (cog.tool_call is the ONLY event an
//     explain call produces — FR-002's "standard tool-call telemetry"; the
//     rest of cog.* is the same recorded-observability class), and
//   - the `guardian.*` feedback namespace (guardian.report_card — the
//     feedback layer's own output must never become a scoring input).
//
// This extends the existing catalog-sweep family: internal/tui's
// TestExerciseRubricTermsAreCatalogedEventTypes pins every term to a REAL
// cataloged event type; this sweep pins the complement — a term may not be
// one of the types the tutor lane emits. Faith (TASK-118) is unshipped;
// when a faith field/event exists, its type joins the banned set here (the
// research R2 obligation, recorded where it will be enforced).
func TestRubricHygieneNoTutorLaneTerms(t *testing.T) {
	tutorLane := func(term string) bool {
		return strings.HasPrefix(term, "cog.") || strings.HasPrefix(term, "guardian.")
	}
	for _, def := range ScenarioExercises {
		for _, term := range def.RubricTerms {
			if tutorLane(term) {
				t.Errorf("exercise %q rubric term %q references tutor-lane telemetry — tutoring is charge-free, faith-free, and excluded from every rubric (spec 063 FR-003)",
					def.ID, term)
			}
		}
	}
}
