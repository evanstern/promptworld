package sim

import (
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// --- spec 102 T005: replay determinism of the new guardian arms (FR-007) ---

// TestGuardianMemoryReplayByteIdentity folds the same recorded guardian
// memory/consolidation/dream/tuning sequence into two fresh states and
// requires byte-identical marshals — the reducer arms apply recorded
// outcomes and re-derive nothing (spec 092 doctrine), so live and replay can
// never diverge.
func TestGuardianMemoryReplayByteIdentity(t *testing.T) {
	tuned := defaultTuning()
	tuned.StewardCadenceTicks = 3600
	events := []store.Event{
		NewTuningEvent(0, tuned),
		gmEvent(t, "guardian.memory_added", 100, 11, GuardianMemoryPayload{Text: "Ash nearly starved", Salience: 7}),
		gmEvent(t, "guardian.memory_added", 200, 12, GuardianMemoryPayload{Text: "routine dawn", Salience: 3}),
		gmEvent(t, "guardian.memory_embedded", 210, 13, GuardianMemoryEmbeddedPayload{MemSeq: 11, Vec: []float32{1, 0}, Model: "m"}),
		gmEvent(t, "guardian.memory_promoted", 220, 14, GuardianMemoryPromotedPayload{MemTick: 100, TextHash: MemoryHash("Ash nearly starved"), Boost: 2}),
		gmEvent(t, "guardian.memory_faded", 230, 15, GuardianMemoryFadedPayload{MemTick: 200, TextHash: MemoryHash("routine dawn")}),
		gmEvent(t, "guardian.memory_added", 240, 16, GuardianMemoryPayload{Text: "the fire ran low twice", Salience: 4}),
		gmEvent(t, "guardian.salience_revised", 250, 17, GuardianSalienceRevisedPayload{MemTick: 240, TextHash: MemoryHash("the fire ran low twice"), Salience: 2, Reason: DreamReasonHabituation}),
		gmEvent(t, "guardian.consolidated", 260, 18, GuardianConsolidatedPayload{Night: 1, UpTo: 240, Outcome: ConsolidationAccepted, Promoted: 1, Faded: 1}),
	}
	fold := func() string {
		s := NewState(9, worldmap.Generate(9, 32, 32))
		for _, e := range events {
			if err := s.Apply(e); err != nil {
				t.Fatalf("apply %s: %v", e.Type, err)
			}
		}
		return string(s.Marshal())
	}
	if a, b := fold(), fold(); a != b {
		t.Fatal("guardian memory arm fold is not byte-deterministic")
	}
}

// TestPre102StateMarshalUnchanged pins the additive-fields compat half: a
// state that never saw a spec-102 event marshals WITHOUT any of the new
// fields (omitempty), so pre-102 snapshots and their hashes are untouched.
func TestPre102StateMarshalUnchanged(t *testing.T) {
	s := NewState(9, worldmap.Generate(9, 32, 32))
	b := string(s.Marshal())
	for _, field := range []string{"guardian_memories", "guardian_mem_up_to", "steward_cadence_ticks"} {
		if contains := (len(b) > 0 && (stringIndex(b, field) >= 0)); contains {
			t.Errorf("pre-102 marshal carries %q — additive-field discipline broken", field)
		}
	}
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
