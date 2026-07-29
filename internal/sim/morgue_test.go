package sim

import (
	"fmt"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// TestCharterObservedArm (spec 044 US2, T014; spec 072 FR-006): the reducer
// records the current charter fingerprint AND its authorship
// (CharterCustom = !Default, latest observation wins), and refuses an empty
// fingerprint at the door.
func TestCharterObservedArm(t *testing.T) {
	s := NewState(7, testMap(7))
	if s.CharterCustom {
		t.Error("genesis CharterCustom = true, want the conservative false zero value")
	}
	if err := s.Apply(store.Event{Tick: 10, Type: "guardian.charter_observed",
		Payload: mustPayload(CharterObservedPayload{Fingerprint: "aaaa11112222", Default: true})}); err != nil {
		t.Fatalf("first observation: %v", err)
	}
	if s.CharterFingerprint != "aaaa11112222" {
		t.Errorf("CharterFingerprint = %q, want aaaa11112222", s.CharterFingerprint)
	}
	if s.CharterCustom {
		t.Error("CharterCustom = true after a Default observation, want false")
	}
	if err := s.Apply(store.Event{Tick: 20, Type: "guardian.charter_observed",
		Payload: mustPayload(CharterObservedPayload{Fingerprint: "bbbb33334444", Default: false})}); err != nil {
		t.Fatalf("second observation: %v", err)
	}
	if s.CharterFingerprint != "bbbb33334444" {
		t.Errorf("CharterFingerprint = %q, want bbbb33334444 (latest wins)", s.CharterFingerprint)
	}
	if !s.CharterCustom {
		t.Error("CharterCustom = false after a custom observation, want true")
	}
	// A revert to the default charter flips authorship back off (latest wins
	// — spec 072's "in force" reading).
	if err := s.Apply(store.Event{Tick: 25, Type: "guardian.charter_observed",
		Payload: mustPayload(CharterObservedPayload{Fingerprint: "cccc55556666", Default: true})}); err != nil {
		t.Fatalf("third observation: %v", err)
	}
	if s.CharterCustom {
		t.Error("CharterCustom = true after reverting to the default charter, want false")
	}
	if err := s.Apply(store.Event{Tick: 30, Type: "guardian.charter_observed",
		Payload: mustPayload(CharterObservedPayload{})}); err == nil {
		t.Error("empty fingerprint applied, want a rejection")
	}
}

// TestMorgueEpilogueArm (spec 044 US2, T015): the ring appends in event
// order, accepts the run-end sentinel (-1), refuses out-of-range agents and
// empty text, and stays bounded at morgueEpilogueCap.
func TestMorgueEpilogueArm(t *testing.T) {
	s := NewState(7, testMap(7))
	ok := []MorgueEpiloguePayload{
		{Agent: Ref(0), Text: "They kept the fire."},
		{Agent: Ref(-1), Text: "The village is quiet now."},
	}
	for i, p := range ok {
		if err := s.Apply(store.Event{Tick: int64(100 + i), Type: "morgue.epilogue", Payload: mustPayload(p)}); err != nil {
			t.Fatalf("epilogue %d: %v", i, err)
		}
	}
	if len(s.MorgueEpilogues) != 2 ||
		s.MorgueEpilogues[0].Agent != 0 || s.MorgueEpilogues[1].Agent != -1 ||
		s.MorgueEpilogues[0].Tick != 100 {
		t.Errorf("ring = %+v, want the two epilogues in event order", s.MorgueEpilogues)
	}
	bad := []MorgueEpiloguePayload{
		{Agent: Ref(len(s.Agents)), Text: "x"}, // out of range
		{Agent: Ref(-2), Text: "x"},            // below the run-end sentinel
		{Agent: Ref(0), Text: "   "},           // blank text
	}
	for i, p := range bad {
		if err := s.Apply(store.Event{Tick: 200, Type: "morgue.epilogue", Payload: mustPayload(p)}); err == nil {
			t.Errorf("bad epilogue %d applied, want a rejection: %+v", i, p)
		}
	}
	// Ring bound: overflow drops oldest.
	for i := 0; i < morgueEpilogueCap+3; i++ {
		p := MorgueEpiloguePayload{Agent: Ref(0), Text: fmt.Sprintf("entry %d", i)}
		if err := s.Apply(store.Event{Tick: int64(300 + i), Type: "morgue.epilogue", Payload: mustPayload(p)}); err != nil {
			t.Fatal(err)
		}
	}
	if len(s.MorgueEpilogues) != morgueEpilogueCap {
		t.Errorf("ring size = %d, want cap %d", len(s.MorgueEpilogues), morgueEpilogueCap)
	}
	if got := s.MorgueEpilogues[len(s.MorgueEpilogues)-1].Text; got != fmt.Sprintf("entry %d", morgueEpilogueCap+2) {
		t.Errorf("newest ring entry = %q, want the last applied", got)
	}
}

// TestEndedDoorAcceptsMorgueEpilogue (spec 044 US2, T015): the ended world's
// narrowed inject_social door accepts morgue.epilogue — the run-end epilogue
// lands AFTER run.ended by construction — while other whitelisted types stay
// refused (the US1 gating test covers the broad refusal set).
func TestEndedDoorAcceptsMorgueEpilogue(t *testing.T) {
	l := newEndedHarness(t)
	seqBefore := l.st.LastSeq()

	res := runCommand(t, l, command{name: "inject_social", social: []store.Event{
		{Type: "morgue.epilogue", Payload: mustPayload(MorgueEpiloguePayload{Agent: Ref(-1), Text: "The village is quiet now."})},
	}})
	if res.err != nil {
		t.Errorf("morgue.epilogue on an ended world refused: %v", res.err)
	}
	if l.st.LastSeq() != seqBefore+1 {
		t.Errorf("morgue.epilogue did not land: last seq %d, want %d", l.st.LastSeq(), seqBefore+1)
	}
	if n := len(l.state.MorgueEpilogues); n != 1 {
		t.Errorf("State.MorgueEpilogues = %d entries, want 1", n)
	}

	// A non-prose whitelisted type is still refused after run end.
	res = runCommand(t, l, command{name: "inject_social", social: []store.Event{
		{Type: "guardian.charter_observed", Payload: mustPayload(CharterObservedPayload{Fingerprint: "aaaa11112222"})},
	}})
	if res.err == nil || !strings.Contains(res.err.Error(), "run has ended") {
		t.Errorf("charter_observed on an ended world: err = %v, want a \"run has ended\" refusal", res.err)
	}
}
