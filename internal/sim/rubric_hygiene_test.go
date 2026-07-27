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
// one of the types the tutor lane emits, and (spec 085 FR-012, discharging
// the research R2 obligation this comment recorded) it may not be a faith.*
// type either: faith is the UNSCORED score, in-fiction only (the
// overjustification caution) — no exercise rubric may ever grade it, while
// prophecy.* events remain rubric-eligible world events.
func TestRubricHygieneNoTutorLaneTerms(t *testing.T) {
	banned := func(term string) bool {
		return strings.HasPrefix(term, "cog.") || strings.HasPrefix(term, "guardian.") ||
			strings.HasPrefix(term, "faith.")
	}
	for _, def := range ScenarioExercises {
		for _, term := range def.RubricTerms {
			if banned(term) {
				t.Errorf("exercise %q rubric term %q references tutor-lane telemetry or the faith score — tutoring is charge-free, faith-free, faith is unscored, and both are excluded from every rubric (spec 063 FR-003, spec 085 FR-012)",
					def.ID, term)
			}
		}
	}
}
