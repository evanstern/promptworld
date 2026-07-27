package guardian

// The report-card producer (spec 063 T011, SC-004): stopping-point triggers
// (run end, exercise resolution, debounced+activity-gated pause episodes),
// grading inputs exactly the shared data source, citation validation
// against the fed trail, storage through the right door per lifecycle, and
// silent deterministic degradation without the chain — all driven with a
// stubbed chain (mockOrch), never a live model.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// seedTrail lands N guardian tool-call records into the trail ring via the
// real absorb-side collector, with committed seqs starting at base.
func seedTrail(mt *Guardian, base int64, verdicts ...string) {
	for i, v := range verdicts {
		p, _ := json.Marshal(sim.NewCogToolCallPayload(
			"turn-metatron-100", i+1, "work_miracle", nil, v, "coordinates not passable", "cloud", 100))
		mt.observeCardActivity(store.Event{Seq: base + int64(i), Tick: 100 + int64(i), Type: "cog.tool_call", Payload: p})
	}
}

// drainCard pops the one queued card job (t.Fatal when none).
func drainCard(t *testing.T, mt *Guardian) cardJob {
	t.Helper()
	select {
	case job := <-mt.cardQ:
		return job
	default:
		t.Fatal("no card job queued")
		return cardJob{}
	}
}

func queuedCards(mt *Guardian) int { return len(mt.cardQ) }

// TestReportCardProducerStoresValidatedNote: the happy path — a pause
// stopping point with activity queues a job whose trail carries the real
// seqs; the stubbed chain's note cites one; the producer validates and
// injects guardian.report_card with the graded fingerprint and the cited
// seqs; the reducer keeps it as the latest card; the graded window advances.
func TestReportCardProducerStoresValidatedNote(t *testing.T) {
	mt, orch, inj, _ := newTestGuardian(t, "ignored")
	orch.reply = "Your charter never mentions coordinates; the working was rejected for them (seq 812)."
	seedTrail(mt, 812, "rejected_gate", "rejected_gate")

	mt.observeCardTrigger(store.Event{Seq: 900, Tick: 200, Type: "clock.paused"})
	job := drainCard(t, mt)
	if job.reason != "pause" || len(job.trail) != 2 || job.trail[0].Seq != 812 {
		t.Fatalf("queued job = %+v", job)
	}
	if job.fingerprint == "" {
		t.Fatal("job carries no charter fingerprint")
	}

	// The grading prompt is exactly the shared data source: charter text +
	// revision, the seq-bearing trail, and the R1 fact sheet.
	prompt := mt.cardPrompt(job)
	for _, want := range []string{"revision " + job.fingerprint, "seq 812", "seq 813", "charges"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("grading prompt missing %q", want)
		}
	}

	mt.produceCard(job)
	// One injected batch: the stored card event.
	var cards []sim.GuardianReportCardPayload
	for _, b := range inj.batches {
		for _, e := range b {
			if e.Type == "guardian.report_card" {
				var p sim.GuardianReportCardPayload
				if err := json.Unmarshal(e.Payload, &p); err != nil {
					t.Fatal(err)
				}
				cards = append(cards, p)
			}
		}
	}
	if len(cards) != 1 {
		t.Fatalf("injected %d cards, want 1", len(cards))
	}
	if cards[0].Fingerprint != job.fingerprint || cards[0].Note != orch.reply {
		t.Errorf("stored card = %+v", cards[0])
	}
	if len(cards[0].Citations) != 1 || cards[0].Citations[0] != 812 {
		t.Errorf("citations = %v, want [812]", cards[0].Citations)
	}
	// The reducer kept it as the latest card on state.
	if inj.state.GuardianReportCard == nil || inj.state.GuardianReportCard.Note != orch.reply {
		t.Error("state does not carry the stored card")
	}
	// The graded window advanced: the same activity cannot be re-graded.
	mt.observeCardTrigger(store.Event{Seq: 901, Tick: 201, Type: "clock.resumed"})
	mt.observeCardTrigger(store.Event{Seq: 902, Tick: 202, Type: "clock.paused"})
	if queuedCards(mt) != 0 {
		t.Error("a stopping point with no NEW activity re-queued a card")
	}
}

// TestReportCardRejectsUnrecordedCitation (SC-004): a critique citing a seq
// outside the fed trail is dropped whole — no event lands, the graded
// window does NOT advance (the activity may still be graded by a later,
// honest critique).
func TestReportCardRejectsUnrecordedCitation(t *testing.T) {
	mt, orch, inj, _ := newTestGuardian(t, "ignored")
	orch.reply = "The charter caused a rejection — seq 9999 proves it."
	seedTrail(mt, 100, "rejected_gate")
	mt.observeCardTrigger(store.Event{Seq: 200, Tick: 50, Type: "clock.paused"})
	mt.produceCard(drainCard(t, mt))
	if len(landedBatches(inj)) != 0 {
		t.Error("a vibes critique landed anyway")
	}
	if mt.cardDoneSeq != 0 {
		t.Error("a dropped card advanced the graded window")
	}
}

// TestReportCardActivityGate (edge case): a stopping point with no guardian
// activity since the last card queues nothing — there is nothing to
// attribute; the deterministic checklist stands alone.
func TestReportCardActivityGate(t *testing.T) {
	mt, _, _, _ := newTestGuardian(t, "ignored")
	mt.observeCardTrigger(store.Event{Seq: 10, Tick: 5, Type: "clock.paused"})
	mt.observeCardTrigger(store.Event{Seq: 11, Tick: 6, Type: "run.ended"})
	mt.observeCardTrigger(store.Event{Seq: 12, Tick: 7, Type: "curriculum.exercise_passed",
		Payload: mustJSON(sim.ExercisePassedPayload{Exercise: "first-night", Stage: "stage-1"})})
	if queuedCards(mt) != 0 {
		t.Errorf("%d cards queued with zero guardian activity", queuedCards(mt))
	}
}

// TestReportCardPauseDebounce (edge case "pause spam"): one card per pause
// episode — a second clock.paused in the same episode is silent; a
// resume+pause with NEW activity re-arms.
func TestReportCardPauseDebounce(t *testing.T) {
	mt, _, _, _ := newTestGuardian(t, "ignored")
	seedTrail(mt, 100, "landed")
	mt.observeCardTrigger(store.Event{Seq: 200, Tick: 50, Type: "clock.paused"})
	if queuedCards(mt) != 1 {
		t.Fatalf("first pause queued %d cards, want 1", queuedCards(mt))
	}
	mt.observeCardTrigger(store.Event{Seq: 201, Tick: 51, Type: "clock.paused"})
	if queuedCards(mt) != 1 {
		t.Error("same-episode pause queued a second card")
	}
	// Resume, new activity, pause again: a fresh episode with fresh activity.
	<-mt.cardQ // consume without producing: the window has NOT advanced
	mt.observeCardTrigger(store.Event{Seq: 202, Tick: 52, Type: "clock.resumed"})
	seedTrail(mt, 300, "landed")
	mt.observeCardTrigger(store.Event{Seq: 400, Tick: 60, Type: "clock.paused"})
	if queuedCards(mt) != 1 {
		t.Error("re-armed pause episode with new activity queued no card")
	}
}

// TestReportCardRunEndRidesEpilogue (research R5): the run-end card lands as
// a morgue.epilogue (agent −1) beside the narrator's — the ended door's
// recorded-prose channel — never as guardian.report_card.
func TestReportCardRunEndRidesEpilogue(t *testing.T) {
	mt, orch, inj, _ := newTestGuardian(t, "ignored")
	orch.reply = "The charter's watch-first duty held: the watch fired before the death (seq 500)."
	seedTrail(mt, 500, "landed")
	mt.observeCardTrigger(store.Event{Seq: 600, Tick: 900, Type: "run.ended"})
	job := drainCard(t, mt)
	if !job.ended {
		t.Fatal("run-end job not marked ended")
	}
	mt.produceCard(job)
	var epilogues []sim.MorgueEpiloguePayload
	for _, b := range inj.batches {
		for _, e := range b {
			switch e.Type {
			case "guardian.report_card":
				t.Error("run-end card rode guardian.report_card — the ended door refuses it")
			case "morgue.epilogue":
				var p sim.MorgueEpiloguePayload
				if err := json.Unmarshal(e.Payload, &p); err != nil {
					t.Fatal(err)
				}
				epilogues = append(epilogues, p)
			}
		}
	}
	if len(epilogues) != 1 || epilogues[0].Agent.ID != -1 {
		t.Fatalf("epilogues = %+v, want one agent -1 entry", epilogues)
	}
	if !strings.Contains(epilogues[0].Text, orch.reply) || !strings.Contains(epilogues[0].Text, job.fingerprint) {
		t.Errorf("epilogue text = %q", epilogues[0].Text)
	}
}

// TestReportCardDegradesSilently (US4 AS-2, FR-005): chain failure or an
// empty critique lands NOTHING — no event, no error theater; the graded
// window stays put.
func TestReportCardDegradesSilently(t *testing.T) {
	mt, orch, inj, _ := newTestGuardian(t, "ignored")
	seedTrail(mt, 100, "landed")
	mt.observeCardTrigger(store.Event{Seq: 200, Tick: 50, Type: "clock.paused"})
	job := drainCard(t, mt)

	orch.err = errBudget{}
	mt.produceCard(job)
	orch.err = nil
	orch.reply = "   "
	mt.produceCard(job)
	if len(inj.batches) != 0 {
		t.Errorf("degraded producer landed %d batches, want 0", len(inj.batches))
	}
}

// errBudget is a minimal transport error for the degradation test.
type errBudget struct{}

func (errBudget) Error() string { return "monthly cloud budget exhausted" }

// TestReportCardTrailIgnoresVillagerCognition: only guardian-correlated
// tool-call records enter the trail — a villager's cog.tool_call is not
// guardian activity and must not arm a card.
func TestReportCardTrailIgnoresVillagerCognition(t *testing.T) {
	mt, _, _, _ := newTestGuardian(t, "ignored")
	p, _ := json.Marshal(sim.NewCogToolCallPayload("planner-3-100", 1, "forage", nil, "landed", "", "local", 100))
	mt.observeCardActivity(store.Event{Seq: 10, Tick: 100, Type: "cog.tool_call", Payload: p})
	mt.observeCardTrigger(store.Event{Seq: 11, Tick: 101, Type: "clock.paused"})
	if queuedCards(mt) != 0 {
		t.Error("villager cognition armed a report card")
	}
}
