package sim

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

func cardEvent(t *testing.T, seq, tick int64, p GuardianReportCardPayload) store.Event {
	t.Helper()
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	return store.Event{Seq: seq, Tick: tick, Type: "guardian.report_card", Payload: b}
}

// TestReportCardReducer (spec 063 T010): the guardian.report_card arm keeps
// the LATEST card on state (the log keeps history), validates rather than
// clamps, and the type is injectable through the social door.
func TestReportCardReducer(t *testing.T) {
	m := worldmap.Generate(1, 32, 32)
	s := NewState(1, m)

	if !InjectableSocialEvent("guardian.report_card") {
		t.Fatal("guardian.report_card is not on the inject_social whitelist")
	}

	first := GuardianReportCardPayload{Fingerprint: "aaaa00000001", Note: "watch placed before nightfall", Citations: []int64{7}}
	if err := s.Apply(cardEvent(t, 10, 100, first)); err != nil {
		t.Fatal(err)
	}
	got := s.GuardianReportCard
	if got == nil || got.Note != first.Note || got.Fingerprint != first.Fingerprint ||
		got.Tick != 100 || got.Seq != 10 || len(got.Citations) != 1 || got.Citations[0] != 7 {
		t.Fatalf("stored card = %+v", got)
	}

	// A later card replaces the stored one — latest-card semantics.
	second := GuardianReportCardPayload{Fingerprint: "bbbb00000002", Note: "the working was rejected twice — seq 812, 907", Citations: []int64{812, 907}}
	if err := s.Apply(cardEvent(t, 1000, 5000, second)); err != nil {
		t.Fatal(err)
	}
	if s.GuardianReportCard.Note != second.Note || s.GuardianReportCard.Seq != 1000 {
		t.Errorf("latest card not kept: %+v", s.GuardianReportCard)
	}

	// Validation: empty fingerprint / empty note / a citation at-or-after the
	// card's own seq are all refused at the door.
	for name, bad := range map[string]store.Event{
		"empty fingerprint": cardEvent(t, 1001, 5000, GuardianReportCardPayload{Note: "x"}),
		"empty note":        cardEvent(t, 1001, 5000, GuardianReportCardPayload{Fingerprint: "cccc00000003", Note: "  "}),
		"future citation":   cardEvent(t, 1001, 5000, GuardianReportCardPayload{Fingerprint: "cccc00000003", Note: "x", Citations: []int64{1001}}),
		"zero citation":     cardEvent(t, 1001, 5000, GuardianReportCardPayload{Fingerprint: "cccc00000003", Note: "x", Citations: []int64{0}}),
	} {
		if err := s.Apply(bad); err == nil {
			t.Errorf("%s: applied, want refusal", name)
		}
	}
	if s.GuardianReportCard.Seq != 1000 {
		t.Error("a refused card perturbed the stored one")
	}
}

// TestReportCardReplay (T010): a state reconstructed from the same events —
// the snapshot+replay path — carries the identical stored card, and a
// serialized state round-trips it (the stored-never-regraded contract's
// mechanical half).
func TestReportCardReplay(t *testing.T) {
	m := worldmap.Generate(1, 32, 32)
	build := func() *State {
		s := NewState(1, m)
		if err := s.Apply(cardEvent(t, 50, 900, GuardianReportCardPayload{
			Fingerprint: "a1b2c3d4e5f6", Note: "under charter a1b2c3…", Citations: []int64{12, 30}})); err != nil {
			t.Fatal(err)
		}
		return s
	}
	a, b := build(), build()
	aj, bj := a.Marshal(), b.Marshal()
	if string(aj) != string(bj) {
		t.Fatal("replayed states diverge")
	}
	var round State
	if err := json.Unmarshal(aj, &round); err != nil {
		t.Fatal(err)
	}
	if round.GuardianReportCard == nil || round.GuardianReportCard.Note != "under charter a1b2c3…" ||
		round.GuardianReportCard.Seq != 50 || len(round.GuardianReportCard.Citations) != 2 {
		t.Errorf("round-tripped card = %+v", round.GuardianReportCard)
	}
	// Pre-063 shape discipline: a state with no card serializes without the
	// field (omitempty — old snapshots stay byte-identical).
	empty := NewState(1, m)
	if strings.Contains(string(empty.Marshal()), "guardian_report_card") {
		t.Error("card-less state serializes the guardian_report_card field")
	}
}
