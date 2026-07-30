package sim

// Spec 101 D5 (the myth briefing) test suite: clustering, ranking, and the
// read-only/no-new-state contract — DominantPlaceMyths never mutates s and
// is a pure function of the current belief corpus.

import (
	"fmt"
	"testing"
)

func mythState(t *testing.T, seed uint64) *State {
	t.Helper()
	return NewState(seed, testMap(seed))
}

// TestDominantPlaceMythsClustersByCoordinate: two agents holding differently
// worded beliefs about nearby coordinates (within one mythClusterCell) rank
// as ONE candidate, with the most-repeated wording surfaced and Holders
// counting distinct believers.
func TestDominantPlaceMythsClustersByCoordinate(t *testing.T) {
	s := mythState(t, 42)
	s.Agents[0].Beliefs = []Belief{
		{ID: 1, Statement: "Thornspire stands at (10,10), where the trees grow thick.", Confidence: 80, Subject: -1},
	}
	s.Agents[1].Beliefs = []Belief{
		{ID: 1, Statement: "Thornspire stands at (10,10), where the trees grow thick.", Confidence: 60, Subject: -1},
	}
	s.Agents[2].Beliefs = []Belief{
		// Nearby but not identical coordinate (same mythClusterCell bucket)
		// and different wording — still one cluster.
		{ID: 1, Statement: "A stone circle waits near (12,11).", Confidence: 40, Subject: -1},
	}
	out := s.DominantPlaceMyths(5)
	if len(out) != 1 {
		t.Fatalf("myths = %d, want 1 (all three beliefs cluster): %+v", len(out), out)
	}
	m := out[0]
	if m.Holders != 3 {
		t.Errorf("holders = %d, want 3", m.Holders)
	}
	if m.Statement != "Thornspire stands at (10,10), where the trees grow thick." {
		t.Errorf("statement = %q, want the most-repeated wording", m.Statement)
	}
	if want := (80 + 60 + 40) / 3; m.Confidence != want {
		t.Errorf("confidence = %d, want %d (average)", m.Confidence, want)
	}
}

// TestDominantPlaceMythsRanksByHoldersThenConfidence: a myth held by more
// villagers outranks one held by fewer, even at lower confidence; among
// equal holder counts, higher confidence wins.
func TestDominantPlaceMythsRanksByHoldersThenConfidence(t *testing.T) {
	s := mythState(t, 42)
	s.Agents[0].Beliefs = []Belief{{ID: 1, Statement: "A shrine near (10,10).", Confidence: 30, Subject: -1}}
	s.Agents[1].Beliefs = []Belief{{ID: 1, Statement: "A shrine near (10,10).", Confidence: 30, Subject: -1}}
	s.Agents[2].Beliefs = []Belief{{ID: 1, Statement: "A den near (40,40).", Confidence: 90, Subject: -1}}
	out := s.DominantPlaceMyths(5)
	if len(out) != 2 {
		t.Fatalf("myths = %d, want 2: %+v", len(out), out)
	}
	if out[0].Holders != 2 {
		t.Errorf("top-ranked holders = %d, want 2 (holder count beats raw confidence)", out[0].Holders)
	}
}

// TestDominantPlaceMythsExcludesAgentSubjectBeliefs: a belief about a
// villager (Subject >= 0) is never a place-myth candidate — there is no
// rumor-to-place linkage to invent (Rumor.Subject is always an agent index).
func TestDominantPlaceMythsExcludesAgentSubjectBeliefs(t *testing.T) {
	s := mythState(t, 42)
	s.Agents[0].Beliefs = []Belief{{ID: 1, Statement: "Ash is trustworthy, seen near (10,10).", Confidence: 80, Subject: 0}}
	if out := s.DominantPlaceMyths(5); len(out) != 0 {
		t.Errorf("myths = %+v, want none (Subject >= 0 is a belief about a villager, not a place)", out)
	}
}

// TestDominantPlaceMythsExcludesDeadAndUncoordinated: a dead agent's beliefs
// never contribute (a "current consensus" reading), and a belief with no
// recognizable coordinate is silence, not a candidate — the
// statementReferencesPlace/statementExpectations judgeable-or-silent
// discipline internal/mind/reconcile.go itself follows.
func TestDominantPlaceMythsExcludesDeadAndUncoordinated(t *testing.T) {
	s := mythState(t, 42)
	s.Agents[0].Dead = true
	s.Agents[0].Beliefs = []Belief{{ID: 1, Statement: "Thornspire at (10,10).", Confidence: 90, Subject: -1}}
	s.Agents[1].Beliefs = []Belief{{ID: 1, Statement: "Something is true, somewhere.", Confidence: 90, Subject: -1}}
	if out := s.DominantPlaceMyths(5); len(out) != 0 {
		t.Errorf("myths = %+v, want none (dead holder + no coordinate)", out)
	}
}

// TestDominantPlaceMythsTopNBounds: topN caps the returned candidate count;
// topN <= 0 returns every cluster found.
func TestDominantPlaceMythsTopNBounds(t *testing.T) {
	s := mythState(t, 42)
	for i := 0; i < 4; i++ {
		x := 10 + i*20
		s.Agents[i].Beliefs = []Belief{{ID: 1, Statement: mustPlaceStatement(x), Confidence: 50, Subject: -1}}
	}
	if out := s.DominantPlaceMyths(2); len(out) != 2 {
		t.Errorf("myths capped at topN=2 = %d, want 2", len(out))
	}
	if out := s.DominantPlaceMyths(0); len(out) != 4 {
		t.Errorf("myths with topN<=0 = %d, want every cluster (4)", len(out))
	}
}

// TestDominantPlaceMythsPure: the derivation never mutates state (D5: no new
// events, no self-grading, computed fresh every call).
func TestDominantPlaceMythsPure(t *testing.T) {
	s := mythState(t, 42)
	s.Agents[0].Beliefs = []Belief{{ID: 1, Statement: "Thornspire at (10,10).", Confidence: 70, Subject: -1}}
	before := string(s.Marshal())
	_ = s.DominantPlaceMyths(5)
	_ = s.DominantPlaceMyths(5)
	if after := string(s.Marshal()); before != after {
		t.Error("DominantPlaceMyths mutated state — it must be a pure read")
	}
}

func mustPlaceStatement(x int) string {
	return fmt.Sprintf("A place near (%d,10).", x)
}
