package tui

// The report-card console rendering (spec 063 T012), across all five
// composition cases: {note-only, checklist-only, both, degraded, none}.
// The stored attribution note composes into spec 053's card seam from
// replica state — re-read, never re-graded — with the unseen badge between
// stopping points; the rubric checklist composes ABOVE it through TASK-127's
// shared renderer (reportCardView via its reportCard consoleCard wrapper,
// spec 056) whenever rubric data exists — one artifact, checklist
// authoritative, the note additive beneath (standing resolution 1).

import (
	"encoding/json"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
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

// scenarioCardModel returns a connected widescreen model on a first-night
// scenario world (the seeded exercise is a manifest fact, views.go
// scenarioExercise).
func scenarioCardModel(t *testing.T) Model {
	t.Helper()
	m := widescreenModel(t)
	m.connected = true
	m.w.Manifest.Scenario = &world.ScenarioConfig{Exercise: "first-night"}
	return m
}

// TestReportCardChecklistOnly (T012, standing resolution 1's checklist-only
// case): an exercise resolution with NO stored note composes exactly one
// card — TASK-127's shared reportCardView wrapped as its reportCard
// consoleCard — all terms met (the recorded pass is the instrument — spec
// 072 FR-002), backing preferring the pass's own Evidence, and no
// attribution block.
func TestReportCardChecklistOnly(t *testing.T) {
	m := scenarioCardModel(t)
	m.replica.CurriculumPasses = []sim.CurriculumPass{{
		Exercise: "first-night", Stage: "stage-1", Tick: 900,
		Evidence: []sim.EvidenceRef{
			{Type: "sim.day_started", Seq: 40, Tick: 890},
			{Type: "metatron.order_placed", Seq: 12, Tick: 200},
		},
	}}
	m.rebuildConsoleCards() // the curriculum.exercise_passed applyEvent site calls this
	if len(m.consoleCards) != 1 {
		t.Fatalf("consoleCards = %d, want 1 (checklist only)", len(m.consoleCards))
	}
	if _, ok := m.consoleCards[0].(reportCard); !ok {
		t.Fatalf("card is %T, want the spec-056 reportCard wrapper", m.consoleCards[0])
	}
	rendered := m.consoleCards[0].renderCard(90)
	if !strings.Contains(rendered, "report card · first-night") {
		t.Errorf("checklist card missing the shared renderer's title: %q", rendered)
	}
	// The recorded pass proves every term held at pass time: evaluator
	// labels, all met, evidence-backed where a term's event type matches.
	if !strings.Contains(rendered, "✓ village survives to dawn of day 2 (sim.day_started · seq 40)") ||
		!strings.Contains(rendered, "✓ a watch set before nightfall (metatron.order_placed · seq 12)") ||
		!strings.Contains(rendered, "✓ no villager dies") {
		t.Errorf("checklist rows wrong: %q", rendered)
	}
	if strings.Contains(rendered, "✗") || strings.Contains(rendered, "…") {
		t.Errorf("recorded-pass checklist must carry no unmet marker (re-read, never re-grade): %q", rendered)
	}
	if strings.Contains(rendered, "what your words did") {
		t.Error("checklist-only card carries an attribution block")
	}
}

// TestReportCardChecklistLiveMarkers (spec 072 FR-002's live leg): a stored
// attribution note is a stopping point with no pass and no run end — the
// checklist grades sim.EvaluateRubric live: met terms ✓, unmet terms the
// pending … marker, never the concluded ✗.
func TestReportCardChecklistLiveMarkers(t *testing.T) {
	m := scenarioCardModel(t)
	m.replica.GuardianReportCard = &sim.GuardianReportCard{Fingerprint: "ff00", Note: "a note"}
	m.rebuildConsoleCards()
	if len(m.consoleCards) != 2 {
		t.Fatalf("consoleCards = %d, want 2 (checklist + note)", len(m.consoleCards))
	}
	rendered := m.consoleCards[0].renderCard(90)
	if !strings.Contains(rendered, "… village survives to dawn of day 2") ||
		!strings.Contains(rendered, "✓ no villager dies (agent.died: 0)") ||
		!strings.Contains(rendered, "… a watch set before nightfall (metatron.order_placed: 0)") {
		t.Errorf("live checklist rows wrong: %q", rendered)
	}
	if strings.Contains(rendered, "✗") {
		t.Errorf("live checklist must never render the concluded ✗ marker: %q", rendered)
	}
}

// TestReportCardBothComposeChecklistAboveNote (T012, standing resolution 1's
// "both" case — ONE artifact): with rubric data AND a stored note, the seam
// composes the checklist card FIRST (authoritative) and the attribution
// note beneath it, clearly its own block, in the console view.
func TestReportCardBothComposeChecklistAboveNote(t *testing.T) {
	m := scenarioCardModel(t)
	m.transcript = []string{"you: hello"}
	m.replica.CurriculumPasses = []sim.CurriculumPass{{
		Exercise: "first-night", Stage: "stage-1", Tick: 900,
		Evidence: []sim.EvidenceRef{{Type: "sim.day_started", Seq: 40, Tick: 890}},
	}}
	m.applyEvent(cardPush(t, 2000, sim.GuardianReportCardPayload{
		Fingerprint: "a1b2c3d4e5f6",
		Note:        "Your charter's watch-first duty held (seq 40).",
		Citations:   []int64{40},
	}))
	if len(m.consoleCards) != 2 {
		t.Fatalf("consoleCards = %d, want 2 (checklist + note)", len(m.consoleCards))
	}
	if _, ok := m.consoleCards[0].(reportCard); !ok {
		t.Errorf("first card is %T, want the checklist (authoritative, composed first)", m.consoleCards[0])
	}
	if _, ok := m.consoleCards[1].(noteCard); !ok {
		t.Errorf("second card is %T, want the attribution note beneath", m.consoleCards[1])
	}
	var mdl tea.Model = m
	mdl = update(mdl, "G")
	view := mdl.(Model).consoleView()
	checklistIdx := strings.Index(view, "report card · first-night")
	noteIdx := strings.Index(view, "what your words did")
	readIdx := strings.Index(view, "charter · skills")
	if checklistIdx < 0 || noteIdx < 0 {
		t.Fatalf("both halves must render: checklist@%d note@%d", checklistIdx, noteIdx)
	}
	if !(checklistIdx < noteIdx && noteIdx < readIdx) {
		t.Errorf("composition order wrong: checklist@%d note@%d readSurface@%d", checklistIdx, noteIdx, readIdx)
	}
}

// TestReportCardChecklistStoppingPointGate: on a scenario world with NO
// stopping point on the record (no pass, no note, run alive), the seam
// stays empty — never a mid-run card (FR-006/SC-004's 0%-mid-run half).
func TestReportCardChecklistStoppingPointGate(t *testing.T) {
	m := scenarioCardModel(t)
	m.rebuildConsoleCards()
	if len(m.consoleCards) != 0 {
		t.Errorf("mid-run scenario world composed %d cards, want 0", len(m.consoleCards))
	}
	// The run ending is itself a stopping point: the checklist concludes.
	m.replica.Ended = true
	m.rebuildConsoleCards()
	if len(m.consoleCards) != 1 {
		t.Fatalf("ended run composed %d cards, want 1 (concluded checklist)", len(m.consoleCards))
	}
	if !strings.Contains(m.consoleCards[0].renderCard(90), "✗") {
		t.Error("ended-run checklist should carry concluded (✗) markers, not pending")
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
