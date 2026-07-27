package scribe

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// Morgue exercise-outcome tests (spec 054 US5, T014): a scenario world's run
// summary names the exercise and its outcome in the no-blame register; an
// ambient world's summary is byte-identically silent about exercises.

// endedHistory is a minimal ended run: one death, then the run declaration.
// extra events (e.g. a recorded pass) are folded before the end.
func endedHistory(t *testing.T, extra ...store.Event) []store.Event {
	t.Helper()
	events := []store.Event{
		{Seq: 1, Tick: 0, Type: "world.created",
			Payload: mustPayloadJSON(t, sim.WorldCreatedPayload{Name: "scen", Seed: 42})},
	}
	seq := int64(1)
	for _, e := range extra {
		seq++
		e.Seq = seq
		events = append(events, e)
	}
	deaths := []sim.DeathRecord{{Agent: 0, Tick: 60000, Cause: "gru"}}
	events = append(events,
		store.Event{Seq: seq + 1, Tick: 60000, Type: "agent.died",
			Payload: mustPayloadJSON(t, sim.DiedPayload{Agent: sim.Ref(0), Cause: "gru"})},
		store.Event{Seq: seq + 2, Tick: 60000, Type: "run.ended",
			Payload: mustPayloadJSON(t, sim.RunEndedPayload{Tick: 60000, Deaths: deaths, FinalCause: "gru"})},
	)
	return events
}

// TestMorgueRunSummaryNamesFailedExercise (FR-010): a failed scenario run's
// summary carries the exercise line, failed, evidence-register wording.
func TestMorgueRunSummaryNamesFailedExercise(t *testing.T) {
	scr, _, dir := newMorgueScribe(t, endedHistory(t))
	scr.SetScenario("first-night")
	s := readMorgue(t, dir)
	want := "- **The exercise**: first-night — failed — the run ended before its rubric was met."
	if !strings.Contains(s, want) {
		t.Errorf("morgue missing %q; content:\n%s", want, s)
	}
	if !strings.Contains(s, "## The run — ended day 1") {
		t.Errorf("run summary section missing:\n%s", s)
	}
}

// TestMorgueRunSummaryNamesPassedExercise: a pass recorded before the run's
// end renders as passed — the world went on, and then it didn't.
func TestMorgueRunSummaryNamesPassedExercise(t *testing.T) {
	pass := store.Event{Tick: 30000, Type: "curriculum.exercise_passed",
		Payload: mustPayloadJSON(t, sim.ExercisePassedPayload{Exercise: "first-night", Stage: "stage-1", Tick: 30000})}
	scr, _, dir := newMorgueScribe(t, endedHistory(t, pass))
	scr.SetScenario("first-night")
	s := readMorgue(t, dir)
	if !strings.Contains(s, "- **The exercise**: first-night — passed") {
		t.Errorf("morgue should name the passed exercise; content:\n%s", s)
	}
}

// TestMorgueRunSummaryAmbientNoExerciseLine: no scenario, no line — the
// ambient morgue is unchanged (contract §1.3).
func TestMorgueRunSummaryAmbientNoExerciseLine(t *testing.T) {
	_, _, dir := newMorgueScribe(t, endedHistory(t))
	s := readMorgue(t, dir)
	if strings.Contains(s, "The exercise") {
		t.Errorf("ambient morgue must not mention an exercise; content:\n%s", s)
	}
}
