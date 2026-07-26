package tui

// The report-card console card (spec 063 T012): the attribution-note half of
// the ONE composed card artifact (standing resolution 1). The stored note is
// re-read from replica state (State.GuardianReportCard — the reducer keeps
// the latest card; the log keeps history) and composed into spec 053's card
// seam (consoleCard / Model.consoleCards / consoleCardLines). TASK-127's
// shared rubric-checklist renderer (reportCardView, spec 056 contract §4)
// composes ABOVE the note inside the same seam when rubric data exists —
// the checklist is authoritative, the note additive prose beneath it,
// clearly its own block; until that renderer merges, the note ships
// standalone behind the same seam (spec 063 assumption 1). Between stopping
// points the existing unseen-badge pattern announces a fresh card — never a
// takeover, never mid-run (FR-006).

import (
	"fmt"
	"strings"

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

// rebuildConsoleCards recomposes the console card seam from replica state —
// the stored-note re-read, never a re-grade (FR-006). Called on connect
// (late attach shows the stored card) and when a guardian.report_card event
// lands. Composition order per standing resolution 1: the rubric checklist
// (TASK-127's consoleCard wrapper over reportCardView) composes FIRST when
// rubric data exists; the attribution note beneath it. This task ships the
// note half; the checklist wrapper joins this exact seam when spec 056's
// renderer merges.
func (m *Model) rebuildConsoleCards() {
	m.consoleCards = m.consoleCards[:0]
	if c := m.buildNoteCard(m.sk()); c != nil {
		m.consoleCards = append(m.consoleCards, *c)
	}
}
