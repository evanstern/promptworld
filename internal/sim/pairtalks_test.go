package sim

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// TASK-109 / spec 061 Phase 2 (US2): the event-sourced pair last-exchange
// ledger. These pin FR-001 and the SC-004 replay/snapshot compat: the record
// updates on every talk (hail-founded and ambient alike), is unordered (one
// record per pair either direction), a pre-061 snapshot loads as never-talked
// and marshals byte-identically when empty, and a talk-heavy replay is
// deterministic.

func talkedEvent(a, b int, tick int64) store.Event {
	return store.Event{Tick: tick, Type: "agent.talked",
		Payload: mustPayload(TalkedPayload{A: Ref(a), B: Ref(b)})}
}

// TestPairTalkRecordUpdatesOnTalk (US2-AS1): the agent.talked arm records the
// pair's last-exchange tick; both orderings of the pair are the ONE record.
func TestPairTalkRecordUpdatesOnTalk(t *testing.T) {
	s := NewState(42, testMap(42))
	if got := s.PairLastTalk(1, 7); got != 0 {
		t.Fatalf("fresh world: PairLastTalk(1,7) = %d, want 0 (never talked)", got)
	}

	if err := s.Apply(talkedEvent(1, 7, 1000)); err != nil {
		t.Fatal(err)
	}
	if got := s.PairLastTalk(1, 7); got != 1000 {
		t.Errorf("PairLastTalk(1,7) = %d, want 1000", got)
	}
	// Unordered: the reversed pair reads the same record.
	if got := s.PairLastTalk(7, 1); got != 1000 {
		t.Errorf("PairLastTalk(7,1) = %d, want 1000 (unordered — one record)", got)
	}

	// A later talk in the OTHER direction updates the SAME record, not a new one.
	if err := s.Apply(talkedEvent(7, 1, 9000)); err != nil {
		t.Fatal(err)
	}
	if got := s.PairLastTalk(1, 7); got != 9000 {
		t.Errorf("PairLastTalk(1,7) = %d, want 9000 after reverse-order talk", got)
	}
	if len(s.PairTalks) != 1 {
		t.Errorf("PairTalks has %d records, want 1 (one per unordered pair): %+v", len(s.PairTalks), s.PairTalks)
	}
	// Stored A<B invariant.
	if s.PairTalks[0].A != 1 || s.PairTalks[0].B != 7 {
		t.Errorf("record not stored with A<B: %+v", s.PairTalks[0])
	}
}

// TestPairTalkRecordsBothHailAndAmbient (US2 / T002): the reducer arm is the
// single write site, so a talk founded by the ambient beat and a talk founded
// by the hail sweep both update the ledger identically — the arm is the same
// event either way (talkEvents emits agent.talked for both callers).
func TestPairTalkRecordsBothHailAndAmbient(t *testing.T) {
	m := testMap(42)

	// Ambient path: the executor's adjacency beat emits talkEvents for an
	// adjacent, past-cooldown pair. Drive two co-located idle agents and confirm
	// the ledger populated.
	ambient := NewState(42, m)
	for i := range ambient.Agents {
		if i > 1 {
			ambient.Agents[i].Dead = true // isolate the 0/1 pair
		}
	}
	ambient.Agents[0].X, ambient.Agents[0].Y = 20, 20
	ambient.Agents[1].X, ambient.Agents[1].Y = 20, 21 // Manhattan distance 1: the beat's adjacency
	log := driveTicks(t, ambient, m, 400, nil)
	if countType(log, "agent.talked") == 0 {
		t.Fatal("ambient beat never founded a talk between the co-located pair")
	}
	if ambient.PairLastTalk(0, 1) == 0 {
		t.Errorf("ambient talk did not update the pair ledger: %+v", ambient.PairTalks)
	}

	// Hail path: apply the hail-founded shape (social.hail_met + agent.talked, as
	// hailStep emits) and confirm the same arm records it.
	hail := NewState(42, m)
	if err := hail.Apply(store.Event{Tick: 5000, Type: "social.hail_met",
		Payload: mustPayload(HailMetPayload{From: Ref(0), To: Ref(1)})}); err != nil {
		t.Fatal(err)
	}
	if err := hail.Apply(talkedEvent(0, 1, 5000)); err != nil {
		t.Fatal(err)
	}
	if got := hail.PairLastTalk(0, 1); got != 5000 {
		t.Errorf("hail-founded talk ledger = %d, want 5000", got)
	}
}

// TestPairTalkSnapshotCompat (US2-AS2 / SC-004): a state with no talks marshals
// with no pair_talks key (byte-identical to pre-061), and a pre-061 snapshot
// (key absent) loads as every-pair-never-talked.
func TestPairTalkSnapshotCompat(t *testing.T) {
	s := NewState(42, testMap(42))
	blob := s.Marshal()
	if bytes.Contains(blob, []byte(`"pair_talks"`)) {
		t.Fatal("a world with no talks must not carry a pair_talks key (pre-061 byte-identical)")
	}

	// A pre-061 snapshot is exactly this key-absent JSON; unmarshalling leaves
	// PairTalks nil and every pair reads never-talked.
	var back State
	if err := json.Unmarshal(blob, &back); err != nil {
		t.Fatal(err)
	}
	if back.PairTalks != nil {
		t.Errorf("key-absent snapshot loaded non-nil PairTalks: %+v", back.PairTalks)
	}
	if got := back.PairLastTalk(1, 7); got != 0 {
		t.Errorf("pre-061 snapshot: PairLastTalk = %d, want 0 (never talked)", got)
	}

	// Once populated, the record round-trips and the hash is stable.
	if err := s.Apply(talkedEvent(1, 7, 1000)); err != nil {
		t.Fatal(err)
	}
	pop := s.Marshal()
	if !bytes.Contains(pop, []byte(`"pair_talks":[{"a":1,"b":7,"tick":1000}]`)) {
		t.Fatalf("populated ledger did not serialize as expected: %s", pop)
	}
	var back2 State
	if err := json.Unmarshal(pop, &back2); err != nil {
		t.Fatal(err)
	}
	if back2.Hash() != s.Hash() {
		t.Error("populated pair ledger hash not stable across round-trip")
	}
}

// TestPairTalkReplayDeterministic (US2 / SC-004): replaying a talk-heavy event
// sequence twice yields byte-identical ledgers, and the sorted-slice invariant
// makes the marshalled bytes independent of the order pairs first talked.
func TestPairTalkReplayDeterministic(t *testing.T) {
	events := []store.Event{
		talkedEvent(7, 1, 100), // reverse order first
		talkedEvent(4, 6, 200),
		talkedEvent(1, 3, 300),
		talkedEvent(1, 7, 4000), // same pair again, ordered
		talkedEvent(2, 5, 500),
	}
	replay := func() []byte {
		s := NewState(42, testMap(42))
		for _, e := range events {
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		return s.Marshal()
	}
	a, b := replay(), replay()
	if !bytes.Equal(a, b) {
		t.Fatal("replay of the same talk sequence produced diverging ledgers")
	}

	// Applying the SAME set of final talks in a different arrival order yields
	// the identical sorted ledger (canonical bytes, R1).
	shuffled := []store.Event{
		talkedEvent(2, 5, 500),
		talkedEvent(1, 7, 4000),
		talkedEvent(4, 6, 200),
		talkedEvent(3, 1, 300),
	}
	ordered := []store.Event{
		talkedEvent(1, 3, 300),
		talkedEvent(1, 7, 4000),
		talkedEvent(2, 5, 500),
		talkedEvent(4, 6, 200),
	}
	build := func(evs []store.Event) []PairTalk {
		s := NewState(42, testMap(42))
		for _, e := range evs {
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		return s.PairTalks
	}
	sh, od := build(shuffled), build(ordered)
	shJSON, _ := json.Marshal(sh)
	odJSON, _ := json.Marshal(od)
	if !bytes.Equal(shJSON, odJSON) {
		t.Errorf("ledger depends on arrival order:\n shuffled=%s\n ordered =%s", shJSON, odJSON)
	}
}
