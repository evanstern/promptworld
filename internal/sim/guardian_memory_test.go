package sim

import (
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// --- spec 102: the guardian memory store's reducer arms ---

func gmEvent(t *testing.T, typ string, tick, seq int64, payload any) store.Event {
	t.Helper()
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return store.Event{Tick: tick, Seq: seq, Type: typ, Payload: b}
}

func TestGuardianMemoryLifecycle(t *testing.T) {
	s := NewState(7, worldmap.Generate(7, 32, 32))

	// Added: text/salience recorded, tick/seq stamped from the envelope.
	if err := s.Apply(gmEvent(t, "guardian.memory_added", 100, 11, GuardianMemoryPayload{Text: "Ash nearly starved", Salience: 7})); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(gmEvent(t, "guardian.memory_added", 200, 12, GuardianMemoryPayload{Text: "routine dawn", Salience: 3})); err != nil {
		t.Fatal(err)
	}
	if len(s.GuardianMemories) != 2 || s.GuardianMemories[0].Tick != 100 || s.GuardianMemories[0].Seq != 11 {
		t.Fatalf("store after adds: %+v", s.GuardianMemories)
	}

	// Embedded: vector attaches by seq.
	if err := s.Apply(gmEvent(t, "guardian.memory_embedded", 210, 13, GuardianMemoryEmbeddedPayload{MemSeq: 11, Vec: []float32{1, 0}, Model: "m"})); err != nil {
		t.Fatal(err)
	}
	if len(s.GuardianMemories[0].Vec) != 2 || s.GuardianMemories[0].VecModel != "m" {
		t.Fatalf("embed companion did not attach: %+v", s.GuardianMemories[0])
	}

	// Promoted: salience boost, clamped at MaxSalience.
	hash := MemoryHash("Ash nearly starved")
	if err := s.Apply(gmEvent(t, "guardian.memory_promoted", 220, 14, GuardianMemoryPromotedPayload{MemTick: 100, TextHash: hash, Boost: 9})); err != nil {
		t.Fatal(err)
	}
	if got := s.GuardianMemories[0].Salience; got != MaxSalience {
		t.Fatalf("promoted salience %d, want clamp at %d", got, MaxSalience)
	}

	// Faded: removed whole; vanished target is a no-op, never an error.
	if err := s.Apply(gmEvent(t, "guardian.memory_faded", 230, 15, GuardianMemoryFadedPayload{MemTick: 200, TextHash: MemoryHash("routine dawn")})); err != nil {
		t.Fatal(err)
	}
	if len(s.GuardianMemories) != 1 {
		t.Fatalf("fade left %d memories, want 1", len(s.GuardianMemories))
	}
	if err := s.Apply(gmEvent(t, "guardian.memory_faded", 231, 16, GuardianMemoryFadedPayload{MemTick: 999, TextHash: "deadbeef"})); err != nil {
		t.Fatalf("vanished fade target errored: %v", err)
	}

	// Dream outcomes: salience revision + merge over the guardian store.
	if err := s.Apply(gmEvent(t, "guardian.salience_revised", 240, 17, GuardianSalienceRevisedPayload{MemTick: 100, TextHash: hash, Salience: 4, Reason: DreamReasonHabituation})); err != nil {
		t.Fatal(err)
	}
	if got := s.GuardianMemories[0].Salience; got != 4 {
		t.Fatalf("salience_revised → %d, want 4", got)
	}

	// Consolidated: accepted advances the high-water mark; rejected does not.
	if err := s.Apply(gmEvent(t, "guardian.consolidated", 250, 18, GuardianConsolidatedPayload{Night: 1, UpTo: 100, Outcome: ConsolidationAccepted})); err != nil {
		t.Fatal(err)
	}
	if s.GuardianMemUpTo != 100 {
		t.Fatalf("GuardianMemUpTo = %d, want 100", s.GuardianMemUpTo)
	}
	if err := s.Apply(gmEvent(t, "guardian.consolidated", 260, 19, GuardianConsolidatedPayload{Night: 2, UpTo: 999, Outcome: ConsolidationRejected, Reason: "unparseable"})); err != nil {
		t.Fatal(err)
	}
	if s.GuardianMemUpTo != 100 {
		t.Fatalf("rejected night moved the high-water mark to %d", s.GuardianMemUpTo)
	}
	if buf := s.GuardianEpisodicBuffer(); len(buf) != 0 {
		t.Fatalf("buffer after accepted night: %v", buf)
	}
}

// TestGuardianMemoryCapDropsLowestSalience pins the deterministic overflow
// rule: past the cap, the lowest-salience (ties oldest) entry drops.
func TestGuardianMemoryCapDropsLowestSalience(t *testing.T) {
	s := NewState(7, worldmap.Generate(7, 32, 32))
	for i := 0; i < GuardianMemoryCap; i++ {
		sal := 5
		if i == 3 {
			sal = 1 // the designated victim
		}
		if err := s.Apply(gmEvent(t, "guardian.memory_added", int64(i+1), int64(i+1), GuardianMemoryPayload{Text: textN(i), Salience: sal})); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Apply(gmEvent(t, "guardian.memory_added", 9999, 9999, GuardianMemoryPayload{Text: "overflow", Salience: 5})); err != nil {
		t.Fatal(err)
	}
	if len(s.GuardianMemories) != GuardianMemoryCap {
		t.Fatalf("cap not enforced: %d", len(s.GuardianMemories))
	}
	for _, m := range s.GuardianMemories {
		if m.Text == textN(3) {
			t.Fatal("lowest-salience entry survived the cap")
		}
	}
}

func textN(i int) string {
	return "note " + string(rune('a'+i%26)) + string(rune('0'+i/26%10)) + string(rune('0'+i/260))
}

// TestGuardianMemoryGuards pins the reducer-side text guards (empty, over-cap).
func TestGuardianMemoryGuards(t *testing.T) {
	s := NewState(7, worldmap.Generate(7, 32, 32))
	if err := s.Apply(gmEvent(t, "guardian.memory_added", 1, 1, GuardianMemoryPayload{Text: "", Salience: 3})); err == nil {
		t.Fatal("empty text accepted")
	}
	long := make([]byte, GuardianMemoryTextMax+1)
	for i := range long {
		long[i] = 'x'
	}
	if err := s.Apply(gmEvent(t, "guardian.memory_added", 1, 1, GuardianMemoryPayload{Text: string(long), Salience: 3})); err == nil {
		t.Fatal("over-cap text accepted")
	}
}

// TestSharedConsolidationParsers pins the moved shared helpers (SC-004): the
// same label vocabulary both nightly drivers reference buffers by.
func TestSharedConsolidationParsers(t *testing.T) {
	if got := ParseMemLabel("m3", 5); got != 2 {
		t.Fatalf("ParseMemLabel(m3,5) = %d", got)
	}
	if got := ParseMemLabel("m9", 5); got != -1 {
		t.Fatalf("out-of-range label = %d, want -1", got)
	}
	if got := ParseRoutineLabels([]string{"g2", "g2", "gX", "g9"}, 3); len(got) != 1 || got[0] != 1 {
		t.Fatalf("ParseRoutineLabels = %v", got)
	}
	raw, err := FirstJSONObject("noise {\"a\": {\"b\": 1}} tail")
	if err != nil || raw != `{"a": {"b": 1}}` {
		t.Fatalf("FirstJSONObject = %q, %v", raw, err)
	}
}
