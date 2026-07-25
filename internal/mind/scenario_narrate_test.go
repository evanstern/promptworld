package mind

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
)

// Scenario-cadence narration tests (spec 054 US5, T013): the exercise's
// pass/fail boundary closes an ADDITIONAL chapter; ambient cadence is
// byte-identical (the run.ended half is gated on the armed scenario, and
// exercise_passed never lands on ambient worlds at all).

// TestExercisePassClosesChapter: curriculum.exercise_passed lands a score
// line AND closes a chapter at the boundary, carrying that line.
func TestExercisePassClosesChapter(t *testing.T) {
	md, _, _ := narrMind(t)
	md.chronicleNote(mustEvent(t, 86400, "curriculum.exercise_passed",
		sim.ExercisePassedPayload{Exercise: "first-night", Stage: "stage-1", Tick: 86400}))

	var job narrJob
	select {
	case job = <-md.narrQ:
	default:
		t.Fatal("exercise pass did not close a chapter")
	}
	if !strings.Contains(job.label, "exercise passed") {
		t.Errorf("chapter label %q should name the pass boundary", job.label)
	}
	if len(job.lines) != 1 || !strings.Contains(job.lines[0], "first-night") {
		t.Errorf("chapter lines %v should carry the pass line", job.lines)
	}
}

// TestScenarioRunEndClosesChapter: on a scenario world, run.ended closes the
// exercise's chapter (the fail boundary) — additive to the epilogue path.
func TestScenarioRunEndClosesChapter(t *testing.T) {
	md, _, _ := narrMind(t)
	md.SetScenario("first-night")
	md.chronicleNote(mustEvent(t, 60000, "run.ended",
		sim.RunEndedPayload{Tick: 60000, Deaths: []sim.DeathRecord{{Agent: 0, Tick: 59000, Cause: "gru"}}, FinalCause: "gru"}))

	var job narrJob
	select {
	case job = <-md.narrQ:
	default:
		t.Fatal("scenario run end did not close a chapter")
	}
	if !strings.Contains(job.label, "exercise") {
		t.Errorf("chapter label %q should name the exercise boundary", job.label)
	}
	if len(job.lines) != 1 || !strings.Contains(job.lines[0], "run has ended") {
		t.Errorf("chapter lines %v should carry the run-ended line", job.lines)
	}
}

// TestAmbientRunEndCadenceUnchanged is the regression half (contract §5):
// with no scenario armed, run.ended adds its factual line to the buffer —
// exactly the pre-054 behavior — and closes NO chapter.
func TestAmbientRunEndCadenceUnchanged(t *testing.T) {
	md, _, _ := narrMind(t)
	md.chronicleNote(mustEvent(t, 60000, "run.ended",
		sim.RunEndedPayload{Tick: 60000, Deaths: []sim.DeathRecord{{Agent: 0, Tick: 59000, Cause: "gru"}}, FinalCause: "gru"}))

	select {
	case <-md.narrQ:
		t.Fatal("ambient run.ended must not close a chapter (cadence unchanged)")
	default:
	}
	if len(md.narrLines) != 1 || !strings.Contains(md.narrLines[0], "run has ended") {
		t.Errorf("buffered lines %v should carry only the factual run-ended line", md.narrLines)
	}
}
