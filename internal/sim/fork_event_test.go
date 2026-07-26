package sim

// The world.forked reducer arm (spec 076 FR-007): a recorded-history no-op,
// exactly world.created's posture. The no-op is load-bearing — it is what
// makes a fork's state at the fork tick byte-identical to its parent's at
// the same (tick, seq) (US1 scenario 3), so it is pinned here by
// marshal-identity, not just by "no error".

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

func TestWorldForkedReducerNoOp(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	// Give the state some non-genesis shape so the identity check is not
	// trivially over an empty struct.
	if err := s.Apply(store.Event{Seq: 1, Tick: 0, Type: "world.created",
		Payload: mustJSON(t, WorldCreatedPayload{Name: "aria", Seed: seed})}); err != nil {
		t.Fatal(err)
	}
	before := s.Marshal()

	payload := mustJSON(t, WorldForkedPayload{
		ParentName:      "aria",
		ParentSeed:      seed,
		ParentCreatedAt: "2026-07-26T00:00:00Z",
		ForkTick:        97200,
		ForkSeq:         5000,
	})
	if err := s.Apply(store.Event{Seq: 2, Tick: 97200, Type: "world.forked", Payload: payload}); err != nil {
		t.Fatalf("world.forked must apply cleanly: %v", err)
	}
	if after := s.Marshal(); !bytes.Equal(before, after) {
		t.Errorf("world.forked mutated state:\nbefore: %s\nafter:  %s", before, after)
	}
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
