package mind

import (
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// TestIntentFailedRearmsMind is spec 096's US1-AS3: the mind's re-arm list
// treats a resolved-without-effect agent.intent_failed exactly like
// agent.intent_done/agent.build_failed — the actor is armed for its next
// planner thought, no stuck intent, and (the absorb switch dispatches on
// event Type alone) no double re-arm from a single event.
func TestIntentFailedRearmsMind(t *testing.T) {
	state := sim.NewState(42, worldmap.Generate(42, 64, 64))
	state.Paused = false
	state.Tick = 5000

	md := &Mind{replica: state}
	md.absorb([]store.Event{{Tick: 5000, Seq: 9, Type: "agent.intent_failed",
		Payload: mustJSON(t, sim.IntentFailedPayload{Agent: sim.Ref(0), Goal: "hunt", Reason: "target gone", X: 10, Y: 10})}})

	if !md.pending[0] {
		t.Error("agent.intent_failed did not re-arm its actor")
	}
	if md.pendingSeq[0] != 9 {
		t.Errorf("pendingSeq[0] = %d, want the arming event's own Seq (9)", md.pendingSeq[0])
	}
}
