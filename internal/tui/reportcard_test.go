package tui

// The report-card console rendering (spec 063 T012): the stored attribution
// note composes into spec 053's card seam from replica state — re-read,
// never re-graded — with the unseen badge between stopping points, and
// degrades to no card when nothing is stored. The checklist-bearing cases
// ({checklist-only, both}) compose through TASK-127's shared renderer at
// this same seam when spec 056 merges; the note-only/degraded/none cases
// below are this task's half.

import (
	"encoding/json"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// cardPush builds a committed guardian.report_card event.
func cardPush(t *testing.T, seq int64, p sim.GuardianReportCardPayload) store.Event {
	t.Helper()
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	return store.Event{Seq: seq, Tick: 500, Type: "guardian.report_card", Payload: b}
}

// TestReportCardNoteRendersInConsoleSeam (note-only): a stored card renders
// between the turn stream and the read surface, labeled through the skin
// tokens, with its fingerprint, note, and citations.
func TestReportCardNoteRendersInConsoleSeam(t *testing.T) {
	m := widescreenModel(t)
	m.connected = true
	m.transcript = []string{"you: hello"}
	m.applyEvent(cardPush(t, 1000, sim.GuardianReportCardPayload{
		Fingerprint: "a1b2c3d4e5f6",
		Note:        "Your charter never mentions coordinates; the working was rejected twice (seq 812).",
		Citations:   []int64{812},
	}))
	if len(m.consoleCards) != 1 {
		t.Fatalf("consoleCards = %d, want 1", len(m.consoleCards))
	}
	var mdl tea.Model = m
	mdl = update(mdl, "G")
	view := mdl.(Model).consoleView()

	for _, want := range []string{
		"report card · under charter a1b2c3d4e5f6",
		"what your words did",
		"never mentions coordinates",
		"cites: seq 812",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("console view missing %q", want)
		}
	}
	turnIdx := strings.Index(view, "hello")
	cardIdx := strings.Index(view, "report card · under charter")
	readIdx := strings.Index(view, "charter · skills")
	if !(turnIdx < cardIdx && cardIdx < readIdx) {
		t.Errorf("card out of seam order: turn=%d card=%d read=%d", turnIdx, cardIdx, readIdx)
	}
}

// TestReportCardNoneRendersNothing (none/degraded): no stored card — the
// seam stays empty, exactly the pre-063 console; an empty-note card (the
// defensive degraded shape) also composes nothing rather than an empty box.
func TestReportCardNoneRendersNothing(t *testing.T) {
	m := widescreenModel(t)
	m.connected = true
	m.rebuildConsoleCards()
	if len(m.consoleCards) != 0 {
		t.Errorf("card-less replica composed %d cards", len(m.consoleCards))
	}
	// Degraded: a stored card whose note is blank (cannot arrive through the
	// reducer, which refuses empty notes — this is the render layer's own
	// defensive floor).
	m.replica.GuardianReportCard = &sim.GuardianReportCard{Fingerprint: "ff", Note: "   "}
	m.rebuildConsoleCards()
	if len(m.consoleCards) != 0 {
		t.Error("blank-note card composed a card")
	}
}

// TestReportCardUnseenBadge (US4 AS-3): a card landing while the guardian
// surface is NOT visible sets the existing unseen badge; landing while the
// console is open does not; opening the console clears it — no takeover, no
// interruption, at most the badge.
func TestReportCardUnseenBadge(t *testing.T) {
	m := widescreenModel(t)
	m.connected = true
	m.dockTab = paneChronicle
	m.applyEvent(cardPush(t, 20, sim.GuardianReportCardPayload{Fingerprint: "aa", Note: "n1", Citations: nil}))
	if !m.guardianUnseen {
		t.Error("fresh card while guardian hidden did not set the unseen badge")
	}
	var mdl tea.Model = m
	mdl = update(mdl, "G") // open console — clears the badge (existing pattern)
	m2 := mdl.(Model)
	if m2.guardianUnseen {
		t.Error("opening the console did not clear the badge")
	}
	m2.applyEvent(cardPush(t, 21, sim.GuardianReportCardPayload{Fingerprint: "aa", Note: "n2"}))
	if m2.guardianUnseen {
		t.Error("card landing while the console is open set the badge")
	}
}

// TestReportCardLateAttachRereadsStoredNote (FR-006's stored-never-regraded,
// client half): a replica arriving with a stored card (the connect path)
// composes it — the card seam re-reads recorded state, no fresh badge.
func TestReportCardLateAttachRereadsStoredNote(t *testing.T) {
	m := widescreenModel(t)
	m.connected = true
	m.replica.GuardianReportCard = &sim.GuardianReportCard{
		Tick: 900, Seq: 40, Fingerprint: "beadfeed0000",
		Note: "the watch fired before the death (seq 500)", Citations: []int64{500},
	}
	m.rebuildConsoleCards() // the connectedMsg site calls this after the replica lands
	if len(m.consoleCards) != 1 {
		t.Fatalf("stored card not composed on attach: %d cards", len(m.consoleCards))
	}
	rendered := m.consoleCards[0].renderCard(80)
	if !strings.Contains(rendered, "beadfeed0000") || !strings.Contains(rendered, "seq 500") {
		t.Errorf("attached card render = %q", rendered)
	}
}

// TestReportCardDigestRowRendersStoredNote (D1: the stored note is visible
// on the raw feed like every prose event): the chronicle digest row for a
// guardian.report_card event carries the label, fingerprint, and prose.
func TestReportCardDigestRow(t *testing.T) {
	e := cardPush(t, 30, sim.GuardianReportCardPayload{Fingerprint: "a1b2c3", Note: "short note"})
	segs, ok := digestRegistry["guardian.report_card"](e, []string{"Ash"}, nil)
	if !ok {
		t.Fatal("digest row failed to render")
	}
	line := plainSegs(segs)
	if !strings.Contains(line, "report card under charter a1b2c3") || !strings.Contains(line, "short note") {
		t.Errorf("digest line = %q", line)
	}
}
