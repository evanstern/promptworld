package tui

// The report-card console card (spec 063 T012): the ONE composed card
// artifact (standing resolution 1). The stored attribution note is re-read
// from replica state (State.GuardianReportCard — the reducer keeps the
// latest card; the log keeps history) and composed into spec 053's card
// seam (consoleCard / Model.consoleCards / consoleCardLines); TASK-127's
// shared rubric-checklist renderer (reportCardView via its reportCard
// wrapper, spec 056 / views.go) composes ABOVE the note inside the same
// seam whenever rubric data exists — the checklist is authoritative, the
// note additive prose beneath it, clearly its own block. Between stopping
// points the existing unseen-badge pattern announces a fresh card — never a
// takeover, never mid-run (FR-006).

import (
	"fmt"
	"strings"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/skin"
)

// noteCard is the attribution note's consoleCard implementation: the stored
// prose, its graded charter fingerprint, and its citations — labels resolved
// through the world skin at build time (render is skin-free thereafter).
type noteCard struct {
	label       string // card box title (skin.guardian.report_card_label)
	attribution string // the note block's own header (skin.guardian.attribution_label)
	fingerprint string
	note        string
	citations   []int64
}

// renderCard renders the bordered card block (the charterReadSurfaceBox
// manner: styleBox, header as the first content line, wrapped body).
func (c noteCard) renderCard(width int) string {
	if width < 20 {
		width = 20
	}
	lines := []string{styleHeader.Render(c.label + " · under charter " + c.fingerprint)}
	lines = append(lines, styleDim.Render(c.attribution))
	for _, w := range wrapText(c.note, width-4) {
		lines = append(lines, "  "+w)
	}
	if len(c.citations) > 0 {
		refs := make([]string, len(c.citations))
		for i, s := range c.citations {
			refs[i] = fmt.Sprintf("seq %d", s)
		}
		lines = append(lines, styleDim.Render("cites: "+strings.Join(refs, ", ")))
	}
	return styleBox.Width(width - 2).Render(clipContent(strings.Join(lines, "\n"), width-2))
}

// buildNoteCard resolves the replica's stored card into a noteCard, or nil
// when no card is stored (or its note is empty — the degraded case renders
// nothing rather than an empty frame; the checklist half degrades
// independently).
func (m *Model) buildNoteCard(sk *skin.Skin) *noteCard {
	if m.replica == nil || m.replica.GuardianReportCard == nil {
		return nil
	}
	rc := m.replica.GuardianReportCard
	if strings.TrimSpace(rc.Note) == "" {
		return nil
	}
	return &noteCard{
		label:       sk.ReportCardLabel(),
		attribution: sk.AttributionLabel(),
		fingerprint: rc.Fingerprint,
		note:        rc.Note,
		citations:   rc.Citations,
	}
}

// buildChecklistCard resolves the rubric-checklist half of the card (spec
// 063 standing resolution 1) through TASK-127's shared renderer: the seeded
// exercise's rubric terms as reportCardFacts, wrapped as the spec-056
// reportCard consoleCard. nil on an ambient world (no rubric exists) and —
// the stopping-point discipline — nil until a stopping point is visible in
// durable state: a stored attribution note, a recorded pass for this
// exercise (exercise resolution), or the ended run. Facts prefer the
// recorded pass's own Evidence (the ceremony's authoritative source,
// reportCardFactsFromEvidence) and fall back to the chronicle-ring
// derivation; the marker vocabulary is live (met/pending) until the
// exercise concludes or the run ends.
func (m *Model) buildChecklistCard() *reportCard {
	def, ok := m.scenarioExercise()
	if !ok || m.replica == nil {
		return nil
	}
	var pass *sim.CurriculumPass
	for i := range m.replica.CurriculumPasses {
		if m.replica.CurriculumPasses[i].Exercise == def.ID {
			pass = &m.replica.CurriculumPasses[i]
			break
		}
	}
	if pass == nil && m.replica.GuardianReportCard == nil && !m.runEnded() {
		return nil // no stopping point on the record yet — no card, no theater
	}
	mode := reportCardLive
	var facts []reportCardFact
	switch {
	case pass != nil:
		mode = reportCardConcluded
		facts = reportCardFactsFromEvidence(def, pass.Evidence)
	default:
		if m.runEnded() {
			mode = reportCardConcluded
		}
		facts = reportCardFactsFromEvents(def, m.events)
	}
	return &reportCard{title: def.ID, facts: facts, mode: mode}
}

// rebuildConsoleCards recomposes the console card seam from replica state —
// the stored-note re-read, never a re-grade (FR-006). Called on connect
// (late attach shows the stored card) and when a guardian.report_card event
// lands. Composition order per standing resolution 1 — ONE artifact, two
// ingredient classes: the rubric checklist (TASK-127's reportCard wrapper
// over the shared reportCardView, spec 056) composes FIRST when rubric data
// exists — always authoritative; the attribution note (noteCard) is
// additive prose beneath it, clearly its own block, never a second scoring
// computation. Either half absent degrades to the other alone; both absent
// leaves the seam empty.
func (m *Model) rebuildConsoleCards() {
	m.consoleCards = m.consoleCards[:0]
	if c := m.buildChecklistCard(); c != nil {
		m.consoleCards = append(m.consoleCards, *c)
	}
	if c := m.buildNoteCard(m.sk()); c != nil {
		m.consoleCards = append(m.consoleCards, *c)
	}
}
