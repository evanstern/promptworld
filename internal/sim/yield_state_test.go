package sim

// Spec 062 (instinct yields to intelligence). T001: the yield-window anchor
// (Agent.LastMindIntentDone) is event-sourced, snapshot-compatible (omitempty),
// and armed ONLY by non-reflex intent completions. T005 (below): the PREP gate
// — the yield window suppresses prep and decays, the danger bands suppress
// regardless of the window, reflex completions never arm, and a no-planner
// drive matches pre-062 behavior except the enumerated danger-band suppressions.

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// TestYieldAnchorOmitemptySnapshotCompat (T001): an agent that never completed
// a mind-sourced intent carries no last_mind_intent_done key, so its canonical
// bytes stay identical to a pre-062 snapshot; and a pre-062 snapshot (the key
// absent) round-trips to the 0 sentinel.
func TestYieldAnchorOmitemptySnapshotCompat(t *testing.T) {
	m := testMap(42)
	s := NewState(42, m)

	// Never mind-driven ⇒ the field is absent from the marshalled state.
	if got := s.Marshal(); bytes.Contains(got, []byte("last_mind_intent_done")) {
		t.Fatalf("a never-mind-driven world serialized the yield anchor:\n%s", got)
	}

	// A pre-062 agent snapshot (no last_mind_intent_done key) unmarshals to 0.
	pre062 := []byte(`{"name":"Old","x":1,"y":1,"needs":{"health":1000,"food":600,"rest":600,"warmth":600,"morale":600},"inv":{},"idle_since":0,"last_talk":0}`)
	var a Agent
	if err := json.Unmarshal(pre062, &a); err != nil {
		t.Fatalf("pre-062 agent snapshot did not round-trip: %v", err)
	}
	if a.LastMindIntentDone != 0 {
		t.Fatalf("absent yield anchor unmarshalled to %d, want the 0 sentinel", a.LastMindIntentDone)
	}

	// Once armed, the field serializes with its tick (the omitempty tail stays
	// compact for every un-armed agent).
	a.LastMindIntentDone = 4321
	if got, _ := json.Marshal(a); !bytes.Contains(got, []byte(`"last_mind_intent_done":4321`)) {
		t.Fatalf("an armed yield anchor did not serialize:\n%s", got)
	}
}

// TestYieldAnchorArmedOnlyByMindCompletion (T001/FR-003): the intent-completion
// reducer arm stamps LastMindIntentDone from the completing intent's ring-record
// source — set for planner/plan/meeting completions, NEVER for reflex ones.
func TestYieldAnchorArmedOnlyByMindCompletion(t *testing.T) {
	for _, tc := range []struct {
		source   string
		wantArmed bool
	}{
		{"planner", true},
		{"plan", true},
		{"meeting", true},
		{"reflex", false},
		{"", false},
	} {
		t.Run("source="+tc.source, func(t *testing.T) {
			m := testMap(42)
			s := NewState(42, m)
			a := &s.Agents[0]

			// The intent lands with the source under test, then completes.
			set := store.Event{Tick: 1000, Type: "agent.intent_set",
				Payload: mustPayload(IntentSetPayload{Agent: 0, Goal: "goto_warmth", TargetX: a.X, TargetY: a.Y, Source: tc.source})}
			done := store.Event{Tick: 1200, Type: "agent.intent_done",
				Payload: mustPayload(AgentPayload{Agent: 0})}
			if err := s.Apply(set); err != nil {
				t.Fatal(err)
			}
			if err := s.Apply(done); err != nil {
				t.Fatal(err)
			}

			if tc.wantArmed && a.LastMindIntentDone != 1200 {
				t.Fatalf("%q completion armed the window to %d, want 1200", tc.source, a.LastMindIntentDone)
			}
			if !tc.wantArmed && a.LastMindIntentDone != 0 {
				t.Fatalf("%q completion armed the window to %d, want it never armed (0)", tc.source, a.LastMindIntentDone)
			}
		})
	}
}
