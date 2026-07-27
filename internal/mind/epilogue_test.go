package mind

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
)

// TestQueueEpilogueOnDeath (spec 044 US2, T017): absorbing a death queues an
// epilogue job on the narrator queue — fact sheet from the replica (identity,
// cause, bonds, top memories), no chapter machinery touched.
func TestQueueEpilogueOnDeath(t *testing.T) {
	md, _, _ := narrMind(t)
	md.replica.Agents[0].Dead = true
	md.replica.Agents[0].Memories = []sim.Memory{
		{Tick: 100, Salience: 9, Text: "Came within a breath of the gru."},
		{Tick: 200, Salience: 3, Text: "A quiet afternoon."},
	}
	md.replica.Relations = []sim.Relation{{From: 0, To: 1, Trust: 40, Affection: 10}}

	md.queueEpilogue(mustEvent(t, 95000, "agent.died", sim.DiedPayload{Agent: sim.Ref(0), Cause: "starvation"}))

	var job narrJob
	select {
	case job = <-md.narrQ:
	default:
		t.Fatal("death did not queue an epilogue job")
	}
	if !job.epilogue || job.agent != 0 {
		t.Fatalf("job = %+v, want an epilogue job for agent 0", job)
	}
	facts := strings.Join(job.lines, "\n")
	for _, want := range []string{
		"Ash died of starvation on day 2.",
		"Bond: Birch (trust +40, affection +10).",
		"They remembered: Came within a breath of the gru.",
	} {
		if !strings.Contains(facts, want) {
			t.Errorf("fact sheet missing %q:\n%s", want, facts)
		}
	}
	// The high-salience memory ranks before the quiet one.
	if strings.Index(facts, "breath of the gru") > strings.Index(facts, "quiet afternoon") {
		t.Error("memories not salience-ranked")
	}
}

// TestQueueEpilogueOnRunEnd: the run-end declaration queues the run-level
// epilogue (agent -1) with the whole death ledger as facts.
func TestQueueEpilogueOnRunEnd(t *testing.T) {
	md, _, _ := narrMind(t)
	md.queueEpilogue(mustEvent(t, 97000, "run.ended", sim.RunEndedPayload{
		Tick: 97000, FinalCause: "exposure",
		Deaths: sim.DeathRefs([]sim.DeathRecord{
			{Agent: 2, Tick: 50000, Cause: "exposure"},
			{Agent: 1, Tick: 95000, Cause: "starvation"},
		})}))
	var job narrJob
	select {
	case job = <-md.narrQ:
	default:
		t.Fatal("run end did not queue an epilogue job")
	}
	if !job.epilogue || job.agent != -1 {
		t.Fatalf("job = %+v, want the run-end epilogue (agent -1)", job)
	}
	facts := strings.Join(job.lines, "\n")
	for _, want := range []string{
		"the last villager died of exposure on day 2",
		"Cedar died of exposure on day 1.",
		"Birch died of starvation on day 2.",
	} {
		if !strings.Contains(facts, want) {
			t.Errorf("run fact sheet missing %q:\n%s", want, facts)
		}
	}
}

// TestRunEpilogueLands: good narrator prose lands as ONE morgue.epilogue
// event through the injection door, capped and agent-tagged.
func TestRunEpilogueLands(t *testing.T) {
	md, social, model := narrMind(t)
	model.narrReply = "Ash kept the fire while the others slept, and the cold took them anyway."

	md.runNarration(narrJob{epilogue: true, agent: 0, label: "epilogue for Ash", toTick: 95000,
		lines: []string{"Ash died of starvation on day 2."}})

	if len(social.batches) != 1 || len(social.batches[0]) != 1 {
		t.Fatalf("batches = %+v, want one single-event injection", social.batches)
	}
	e := social.batches[0][0]
	if e.Type != "morgue.epilogue" {
		t.Fatalf("landed type = %q, want morgue.epilogue", e.Type)
	}
	var p sim.MorgueEpiloguePayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		t.Fatal(err)
	}
	if p.Agent.ID != 0 || p.Text != model.narrReply {
		t.Errorf("payload = %+v", p)
	}
}

// TestRunEpilogueFailureIsGap (FR-010): a narrator failure injects nothing
// and carries nothing — a gap in the prose, never a stall or a retry loop
// (the chronicle's carry machinery is chapter-only).
func TestRunEpilogueFailureIsGap(t *testing.T) {
	md, social, _ := narrMind(t)
	// narrReply empty → ErrTierDown from the mock.
	md.runNarration(narrJob{epilogue: true, agent: 0, label: "epilogue for Ash", toTick: 95000,
		lines: []string{"Ash died of starvation on day 2."}})
	if len(social.batches) != 0 {
		t.Fatal("failed epilogue must inject nothing")
	}
	select {
	case <-md.narrRetry:
		t.Fatal("epilogue failure must not carry into the chapter retry")
	default:
	}
}
