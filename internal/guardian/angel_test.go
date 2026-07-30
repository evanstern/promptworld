package guardian

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// --- spec 102: the scheduled cognition lane (T001) ---

// optIn applies the angel-cadence tuning to the fixture's replica AND its
// injector-backed state (the same event both sides would absorb live), so
// scheduleAngel sees an opted-in world.
func optIn(t *testing.T, mt *Guardian, inj *stateInjector, cadence int64) {
	t.Helper()
	tuned := sim.TuningState{}
	parsed, _, err := sim.ParseTuning([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	tuned = *parsed
	tuned.AngelCadenceTicks = cadence
	ev := sim.NewTuningEvent(mt.replica.Tick, tuned)
	if err := mt.replica.Apply(ev); err != nil {
		t.Fatal(err)
	}
	if err := inj.state.Apply(ev); err != nil {
		t.Fatal(err)
	}
	// Refresh the stateMu mirrors (angelOn, memWin) the worker-side paths
	// read — production's per-batch mirrorState.
	mt.mirrorState()
}

// advanceTo moves the fixture replica's clock (absorb-owned in production;
// tests drive absorb-side methods directly per the newTestGuardian contract).
func advanceTo(mt *Guardian, tick int64) {
	mt.replica.Tick = tick
	mt.mirrorState()
}

// drainAngel runs at most one queued scheduled turn synchronously (the
// worker is stopped by the fixture's Close).
func drainAngel(t *testing.T, mt *Guardian) bool {
	t.Helper()
	select {
	case job := <-mt.angelQ:
		mt.runAngel(job)
		return true
	default:
		return false
	}
}

// TestAngelLaneOffByDefault pins FR-007: a world that never opted in
// schedules nothing, however far the clock advances — the guardian stays
// purely event-driven.
func TestAngelLaneOffByDefault(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "quiet")
	advanceTo(mt, 100000)
	mt.scheduleAngel()
	advanceTo(mt, 200000)
	mt.scheduleAngel()
	if mt.angelDue != 0 {
		t.Fatalf("angelDue armed (%d) on a non-opted world", mt.angelDue)
	}
	if drainAngel(t, mt) {
		t.Fatal("a scheduled turn queued on a non-opted world")
	}
	if n := len(inj.batches); n != 0 {
		t.Fatalf("non-opted lane emitted %d batches, want 0", n)
	}
}

// TestAngelLaneSchedulesAndRuns is the FR-001/US1 happy path: an opted-in
// world arms one cadence out, fires when due, runs a full turn through the
// SHARED runTurn body, and the decision trail carries the chain (cog.thought
// then a terminal cog.outcome, job angel-metatron-<tick>).
func TestAngelLaneSchedulesAndRuns(t *testing.T) {
	mt, orch, inj, dir := newTestGuardian(t, "All is well; I keep the watch.")
	optIn(t, mt, inj, 600)

	advanceTo(mt, 1000)
	mt.scheduleAngel() // first sighting arms, never fires
	if mt.angelDue != 1600 {
		t.Fatalf("first sighting armed due=%d, want 1600", mt.angelDue)
	}
	if drainAngel(t, mt) {
		t.Fatal("armed-only pass queued a turn")
	}

	advanceTo(mt, 1700)
	mt.scheduleAngel()
	if !drainAngel(t, mt) {
		t.Fatal("due cadence queued no turn")
	}
	// Phase-preserving advance off the lane's OWN due (1600 + 600).
	if mt.angelDue != 2200 {
		t.Fatalf("due advanced to %d, want 2200", mt.angelDue)
	}

	// The turn ran through the shared body: prompt reached the model, the
	// transcript carries the [cadence] origin marker.
	reqs := orch.requests()
	if len(reqs) != 1 {
		t.Fatalf("want 1 model call, got %d", len(reqs))
	}
	if !strings.Contains(reqs[0].Prompt, "Your own watch cadence has come round") {
		t.Fatalf("angel directive missing from prompt:\n%s", reqs[0].Prompt)
	}
	transcript, err := os.ReadFile(mt.transcriptPath())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(transcript), "[cadence]") {
		t.Fatalf("transcript missing [cadence] marker:\n%s", transcript)
	}
	_ = dir

	// Decision trail (FR-008): one cog.thought and one terminal cog.outcome
	// (adapted — a quiet converse-only turn), sharing the angel job id.
	var thoughts, outcomes []sim.CogOutcomePayload
	var thoughtJobs []string
	for _, b := range inj.batches {
		for _, e := range b {
			switch e.Type {
			case "cog.thought":
				var p sim.CogThoughtPayload
				if err := json.Unmarshal(e.Payload, &p); err != nil {
					t.Fatal(err)
				}
				if p.Class == "angel" {
					thoughtJobs = append(thoughtJobs, p.Job)
					thoughts = append(thoughts, sim.CogOutcomePayload{Job: p.Job})
				}
			case "cog.outcome":
				var p sim.CogOutcomePayload
				if err := json.Unmarshal(e.Payload, &p); err != nil {
					t.Fatal(err)
				}
				if p.Class == "angel" {
					outcomes = append(outcomes, p)
				}
			}
		}
	}
	if len(thoughts) != 1 || len(outcomes) != 1 {
		t.Fatalf("want 1 angel thought + 1 outcome, got %d/%d", len(thoughts), len(outcomes))
	}
	if outcomes[0].Job != thoughtJobs[0] || !strings.HasPrefix(outcomes[0].Job, "angel-metatron-") {
		t.Fatalf("chain broken: thought %q vs outcome %q", thoughtJobs[0], outcomes[0].Job)
	}
	if outcomes[0].Outcome != sim.OutcomeAdapted {
		t.Fatalf("quiet turn outcome %q, want %q", outcomes[0].Outcome, sim.OutcomeAdapted)
	}
}

// TestAngelSuppressedAtSpeed pins the D2 router gate: at a speed where the
// angel budget cannot hold (16x at the bootstrap estimate), the scheduled
// turn is never attempted — a recorded cog.outcome suppressed with the
// router's arithmetic is the single terminal record, and the due advances.
func TestAngelSuppressedAtSpeed(t *testing.T) {
	mt, orch, inj, _ := newTestGuardian(t, "never called")
	optIn(t, mt, inj, 600)
	speedEv := store.Event{Type: "clock.speed_set", Payload: mustJSON(sim.SpeedSetPayload{Speed: "16x"})}
	if err := mt.replica.Apply(speedEv); err != nil {
		t.Fatal(err)
	}

	advanceTo(mt, 1000)
	mt.scheduleAngel() // arms
	advanceTo(mt, 1700)
	mt.scheduleAngel() // due, but suppressed at 16x
	if drainAngel(t, mt) {
		t.Fatal("suppressed lane queued a turn")
	}
	if len(orch.requests()) != 0 {
		t.Fatal("suppressed lane reached the model")
	}
	if mt.angelDue != 2200 {
		t.Fatalf("suppression did not advance due (got %d, want 2200)", mt.angelDue)
	}
	// The suppression record detaches from the absorb goroutine (the mind's
	// `go emitCog` discipline) — poll briefly for the landed batch.
	var found bool
	for wait := 0; wait < 200 && !found; wait++ {
		inj.mu.Lock()
		batches := append([][]store.Event(nil), inj.batches...)
		inj.mu.Unlock()
		for _, b := range batches {
			for _, e := range b {
				if e.Type != "cog.outcome" {
					continue
				}
				var p sim.CogOutcomePayload
				if err := json.Unmarshal(e.Payload, &p); err != nil {
					t.Fatal(err)
				}
				if p.Class == "angel" && p.Outcome == sim.OutcomeSuppressed {
					found = true
					if !strings.Contains(p.Reason, "budget") {
						t.Fatalf("suppression reason lacks the arithmetic: %q", p.Reason)
					}
				}
			}
		}
		if !found {
			time.Sleep(5 * time.Millisecond)
		}
	}
	if !found {
		t.Fatal("no recorded angel suppression")
	}
}

// TestAngelNeverTouchesOrderDoor is the D6 pin: a scheduled turn — even one
// that lands an act — emits NO guardian.order_triggered; order firing has one
// arbiter (matchOrders → triggerWorker), and the cadence lane cannot forge or
// race it. The charter is AUTHORED here (T003): only a lifted ceiling grants
// the scheduled lane an acting tool at all.
func TestAngelNeverTouchesOrderDoor(t *testing.T) {
	mt, _, inj, dir := newTestGuardian(t, "I send comfort.")
	if err := os.WriteFile(filepath.Join(dir, "charter.md"),
		[]byte("You are a bold shepherd; comfort the frightened without being asked."), 0o644); err != nil {
		t.Fatal(err)
	}
	mt.charterFP = "" // fresh revision observes cleanly
	optIn(t, mt, inj, 600)
	mt.replica.GuardianCharges = 3
	mt.runLoop = actLoop(mt, "send_vision", `{"target":"Ash","text":"Rest easy; the fire holds."}`)

	advanceTo(mt, 1000)
	mt.scheduleAngel()
	advanceTo(mt, 1700)
	mt.scheduleAngel()
	if !drainAngel(t, mt) {
		t.Fatal("due cadence queued no turn")
	}
	var landedVision bool
	for _, b := range inj.batches {
		for _, e := range b {
			if e.Type == "guardian.order_triggered" {
				t.Fatal("scheduled turn emitted guardian.order_triggered — the order door has one arbiter")
			}
			if e.Type == "guardian.nudged" {
				landedVision = true
			}
		}
	}
	if !landedVision {
		t.Fatal("acting scheduled turn landed nothing (fixture defect)")
	}
	// The landed act queued a player-facing moment (the triggered-turn
	// discipline) naming the cadence origin.
	mt.stateMu.Lock()
	moments := append([]string(nil), mt.moments...)
	mt.stateMu.Unlock()
	if len(moments) != 1 || !strings.Contains(moments[0], "on my own watch") {
		t.Fatalf("want one cadence moment, got %v", moments)
	}
}

// TestAngelEndedWorldSchedulesNothing: a finished run's caretaker rests — the
// lane checks Ended before arming or firing.
func TestAngelEndedWorldSchedulesNothing(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "quiet")
	optIn(t, mt, inj, 600)
	mt.replica.Ended = true
	advanceTo(mt, 5000)
	mt.scheduleAngel()
	if mt.angelDue != 0 || drainAngel(t, mt) {
		t.Fatal("ended world armed or fired the cadence lane")
	}
}
