package sim

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// Spec 043 US1 (T010): the recent-intent ring is maintained entirely by the
// intent-lifecycle reducer arms, so it is replay-safe by construction. These
// tests pin all five arms, the override-in-quick-succession ordering, ring
// wraparound at cap, and the rejected-never-landed shape (data-model.md).

func applyTo(t *testing.T, s *State, e store.Event) {
	t.Helper()
	if err := s.Apply(e); err != nil {
		t.Fatalf("apply %s: %v", e.Type, err)
	}
}

// TestIntentSetAppendsRecord: agent.intent_set appends a fresh open record with
// goal, source, reason, and landing tick.
func TestIntentSetAppendsRecord(t *testing.T) {
	s := NewState(42, testMap(42))
	applyTo(t, s, store.Event{Tick: 500, Type: "agent.intent_set",
		Payload: mustPayload(IntentSetPayload{Agent: Ref(2), Goal: "chop", Source: "planner", Reason: "need wood"})})
	log := s.Agents[2].IntentLog
	if len(log) != 1 {
		t.Fatalf("IntentLog len = %d, want 1", len(log))
	}
	r := log[0]
	if r.Goal != "chop" || r.Source != "planner" || r.Reason != "need wood" || r.Tick != 500 {
		t.Errorf("record = %+v, want {chop planner \"need wood\" 500}", r)
	}
	if r.Outcome != "" || r.OutcomeTick != 0 {
		t.Errorf("fresh record should be open, got outcome %q @ %d", r.Outcome, r.OutcomeTick)
	}
}

// TestIntentDoneStampsNewestOpen / TestBuildFailedStampsFailed: the closing
// events stamp the newest still-open record.
func TestIntentDoneStampsNewestOpen(t *testing.T) {
	s := NewState(42, testMap(42))
	applyTo(t, s, store.Event{Tick: 100, Type: "agent.intent_set",
		Payload: mustPayload(IntentSetPayload{Agent: Ref(0), Goal: "forage", Source: "reflex"})})
	applyTo(t, s, store.Event{Tick: 260, Type: "agent.intent_done",
		Payload: mustPayload(AgentPayload{Agent: Ref(0)})})
	r := s.Agents[0].IntentLog[0]
	if r.Outcome != "done" || r.OutcomeTick != 260 {
		t.Errorf("record = %+v, want done @ 260", r)
	}
}

func TestBuildFailedStampsFailed(t *testing.T) {
	s := NewState(42, testMap(42))
	applyTo(t, s, store.Event{Tick: 100, Type: "agent.intent_set",
		Payload: mustPayload(IntentSetPayload{Agent: Ref(0), Goal: "build_fire", Source: "planner"})})
	applyTo(t, s, store.Event{Tick: 340, Type: "agent.build_failed",
		Payload: mustPayload(BuildFailedPayload{Agent: Ref(0), Goal: "build_fire", Reason: "no wood"})})
	r := s.Agents[0].IntentLog[0]
	if r.Outcome != "failed" || r.OutcomeTick != 340 {
		t.Errorf("record = %+v, want failed @ 340", r)
	}
}

// TestIntentRejectedAppendsClosed (edge case: rejected/superseded before
// landing): a refused intent never had an open record, so it is appended
// already closed — source "planner", outcome "rejected" — and it does NOT touch
// the agent's live Intent (the refusal never landed).
func TestIntentRejectedAppendsClosed(t *testing.T) {
	s := NewState(42, testMap(42))
	applyTo(t, s, store.Event{Tick: 700, Type: "agent.intent_rejected",
		Payload: mustPayload(IntentRejectedPayload{Agent: Ref(3), Goal: "talk_to", Reason: "stale", StalenessTicks: 1646})})
	if s.Agents[3].Intent != nil {
		t.Errorf("intent_rejected must not set a live Intent: %+v", s.Agents[3].Intent)
	}
	log := s.Agents[3].IntentLog
	if len(log) != 1 {
		t.Fatalf("IntentLog len = %d, want 1", len(log))
	}
	r := log[0]
	if r.Goal != "talk_to" || r.Source != "planner" || r.Reason != "stale" ||
		r.Outcome != "rejected" || r.Tick != 700 || r.OutcomeTick != 700 {
		t.Errorf("rejected record = %+v, want closed rejected planner @ 700", r)
	}
}

// TestPlanExpiredStampsOpenStep: a plan step that FIRED (an open "plan"-source
// record exists) is closed "expired" on plan_expired, matching by goal.
func TestPlanExpiredStampsOpenStep(t *testing.T) {
	s := NewState(42, testMap(42))
	applyTo(t, s, store.Event{Tick: 1000, Type: "agent.intent_set",
		Payload: mustPayload(IntentSetPayload{Agent: Ref(0), Goal: "goto_warmth", Source: "plan"})})
	applyTo(t, s, store.Event{Tick: 1200, Type: "agent.plan_expired",
		Payload: mustPayload(PlanStepPayload{Agent: Ref(0), Job: "j", Step: "goto_warmth", Reason: "window closed"})})
	if n := len(s.Agents[0].IntentLog); n != 1 {
		t.Fatalf("IntentLog len = %d, want 1 (stamped, not appended)", n)
	}
	r := s.Agents[0].IntentLog[0]
	if r.Outcome != "expired" || r.OutcomeTick != 1200 {
		t.Errorf("record = %+v, want expired @ 1200", r)
	}
}

// TestPlanExpiredAppendsUnfiredStep: a step that expired BEFORE firing has no
// open record, so plan_expired appends a closed "expired" record (plan end
// visible at the next thought, FR-005).
func TestPlanExpiredAppendsUnfiredStep(t *testing.T) {
	s := NewState(42, testMap(42))
	applyTo(t, s, store.Event{Tick: 1200, Type: "agent.plan_expired",
		Payload: mustPayload(PlanStepPayload{Agent: Ref(0), Job: "j", Step: "hunt", Reason: "window closed"})})
	log := s.Agents[0].IntentLog
	if len(log) != 1 {
		t.Fatalf("IntentLog len = %d, want 1 (appended)", len(log))
	}
	r := log[0]
	if r.Goal != "hunt" || r.Source != "plan" || r.Outcome != "expired" || r.OutcomeTick != 1200 {
		t.Errorf("appended record = %+v, want closed expired plan hunt @ 1200", r)
	}
}

// TestOverrideQuickSuccession (edge case: instinct override during
// deliberation): two intents landing back-to-back appear as consecutive
// records in landing order, both left OPEN; a subsequent done closes only the
// newest, so the open-then-superseded shape is preserved.
func TestOverrideQuickSuccession(t *testing.T) {
	s := NewState(42, testMap(42))
	applyTo(t, s, store.Event{Tick: 100, Type: "agent.intent_set",
		Payload: mustPayload(IntentSetPayload{Agent: Ref(0), Goal: "forage", Source: "planner"})})
	applyTo(t, s, store.Event{Tick: 101, Type: "agent.intent_set",
		Payload: mustPayload(IntentSetPayload{Agent: Ref(0), Goal: "goto_warmth", Source: "reflex"})})
	log := s.Agents[0].IntentLog
	if len(log) != 2 || log[0].Goal != "forage" || log[1].Goal != "goto_warmth" {
		t.Fatalf("override order not preserved: %+v", log)
	}
	if log[0].Outcome != "" || log[1].Outcome != "" {
		t.Fatalf("both records should be open before any close: %+v", log)
	}
	applyTo(t, s, store.Event{Tick: 260, Type: "agent.intent_done",
		Payload: mustPayload(AgentPayload{Agent: Ref(0)})})
	log = s.Agents[0].IntentLog
	if log[1].Outcome != "done" {
		t.Errorf("newest (override) record should close done: %+v", log[1])
	}
	if log[0].Outcome != "" {
		t.Errorf("superseded record must stay open (open-then-superseded shape): %+v", log[0])
	}
}

// TestRingWraparoundAtCap: past intentLogCap the oldest records drop; the ring
// holds exactly the most recent intentLogCap, in order.
func TestRingWraparoundAtCap(t *testing.T) {
	s := NewState(42, testMap(42))
	total := intentLogCap + 5
	goals := make([]string, total)
	for i := 0; i < total; i++ {
		goals[i] = "g" + string(rune('A'+i))
		applyTo(t, s, store.Event{Tick: int64(100 + i), Type: "agent.intent_set",
			Payload: mustPayload(IntentSetPayload{Agent: Ref(0), Goal: goals[i], Source: "planner"})})
	}
	log := s.Agents[0].IntentLog
	if len(log) != intentLogCap {
		t.Fatalf("ring len = %d, want cap %d", len(log), intentLogCap)
	}
	// The retained window is the last intentLogCap goals, in order.
	want := goals[total-intentLogCap:]
	for i, g := range want {
		if log[i].Goal != g {
			t.Errorf("ring[%d].Goal = %q, want %q (oldest should have dropped)", i, log[i].Goal, g)
		}
	}
}

// TestIntentLogByteStabilityAndReplay: a never-acted agent carries no
// intent_log key (byte-stable vs pre-043 snapshots), and a set→done→set
// timeline round-trips through marshal/unmarshal with a stable hash.
func TestIntentLogByteStabilityAndReplay(t *testing.T) {
	s := NewState(42, testMap(42))
	if bytes.Contains(s.Marshal(), []byte(`"intent_log"`)) {
		t.Fatal("a fresh state should carry no intent_log key")
	}
	applyTo(t, s, store.Event{Tick: 100, Type: "agent.intent_set",
		Payload: mustPayload(IntentSetPayload{Agent: Ref(1), Goal: "chop", Source: "planner"})})
	applyTo(t, s, store.Event{Tick: 200, Type: "agent.intent_done",
		Payload: mustPayload(AgentPayload{Agent: Ref(1)})})
	applyTo(t, s, store.Event{Tick: 300, Type: "agent.intent_set",
		Payload: mustPayload(IntentSetPayload{Agent: Ref(1), Goal: "hunt", Source: "reflex"})})

	blob := s.Marshal()
	var back State
	if err := json.Unmarshal(blob, &back); err != nil {
		t.Fatal(err)
	}
	if back.Hash() != s.Hash() {
		t.Error("state hash not stable across round-trip with an IntentLog")
	}
	log := back.Agents[1].IntentLog
	if len(log) != 2 || log[0].Goal != "chop" || log[0].Outcome != "done" ||
		log[1].Goal != "hunt" || log[1].Outcome != "" {
		t.Errorf("round-tripped IntentLog wrong: %+v", log)
	}
}
