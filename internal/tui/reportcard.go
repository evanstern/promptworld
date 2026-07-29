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

// --- the shared fact resolver (spec 072 FR-001/FR-002) ---
//
// All three card surfaces — the postmortem overlay, the ceremony takeover,
// and this console card — derive their per-term verdicts here, from ONE
// source: a recorded pass (the instrument) or sim.EvaluateRubric over the
// replica (the same pure derivation the pass emitter reads). The panel, the
// emitter, and every card can no longer disagree — and a failed term on a
// concluded surface renders ✗, never a presence-derived ✓.

// reportCardFactsFromRubric renders evaluated rubric terms as report-card
// rows: hand-authored labels ("no villager dies"), honest verdicts, the
// backing event count — RubricTerm is the single verdict currency (spec 072
// FR-001/FR-004).
func reportCardFactsFromRubric(terms []sim.RubricTerm) []reportCardFact {
	facts := make([]reportCardFact, len(terms))
	for i, term := range terms {
		facts[i] = reportCardFact{Term: term.Label, Met: term.Met,
			Backing: fmt.Sprintf("%s: %d", term.Event, term.Count)}
	}
	return facts
}

// reportCardFactsFromPass renders a recorded CurriculumPass: every term met
// — the pass is the instrument, and scenarioRubricEvents emits only when
// every term held, so the card is a re-read of that record, never a
// re-grade (spec 063 doctrine; spec 072 FR-002). Backing prefers the pass's
// own Evidence — the exact satisfying events the unlock derivation read —
// first match by event type; a term with no matching ref keeps the rubric
// backing.
func reportCardFactsFromPass(terms []sim.RubricTerm, evidence []sim.EvidenceRef) []reportCardFact {
	facts := make([]reportCardFact, len(terms))
	for i, term := range terms {
		backing := fmt.Sprintf("%s: %d", term.Event, term.Count)
		for _, ev := range evidence {
			// A migrated world's recorded evidence keeps its pre-094 type
			// strings (translation preserves payloads verbatim, spec 094
			// D4) — normalize before matching so old passes still back
			// their terms; the backing renders the recorded string as-is
			// (it is the auditable historical reference).
			if sim.CanonicalEventType(ev.Type) == term.Event {
				backing = fmt.Sprintf("%s · seq %d", ev.Type, ev.Seq)
				break
			}
		}
		facts[i] = reportCardFact{Term: term.Label, Met: true, Backing: backing}
	}
	return facts
}

// Exported surface (spec 076 FR-018/data-model §5): the same one resolver,
// replica-parametric, so consumers outside the TUI Model — the CLI duel is
// the fourth card surface (postmortem, ceremony, console card, duel) —
// derive facts through the identical switch. Type aliases keep every
// existing call site and test untouched.

// ReportCardFact is one already-resolved rubric row (views.go's
// reportCardFact, exported): the plain-language term, whether it's met, and
// the backing event reference.
type ReportCardFact = reportCardFact

// ReportCardMode selects the marker vocabulary: concluded (✓/✗) or live
// (✓/…).
type ReportCardMode = reportCardMode

// The exported mode constants (views.go's values, aliased types).
const (
	ReportCardConcluded ReportCardMode = reportCardConcluded
	ReportCardLive      ReportCardMode = reportCardLive
)

// ResolveRubricFacts is the ONE precedence switch behind every card
// surface (spec 072 FR-001/FR-002; exported replica-parametric by spec 076
// FR-018): a recorded pass wins (concluded, all-met, evidence-backed);
// otherwise sim.EvaluateRubric over the state grades — concluded once the
// run ended (s.Ended), live (… pending) while it runs. Labels always come
// from the evaluator's terms. A nil state yields nothing: a card invented
// without state is exactly the mechanism spec 072 deleted. Implementing a
// second precedence switch anywhere is a spec violation — one resolver was
// the entire point of spec 072.
func ResolveRubricFacts(s *sim.State, def sim.ExerciseDefinition, pass *sim.CurriculumPass) ([]ReportCardFact, ReportCardMode) {
	if s == nil {
		return nil, ReportCardLive
	}
	terms := sim.EvaluateRubric(s, def, s.Tick)
	switch {
	case pass != nil:
		return reportCardFactsFromPass(terms, pass.Evidence), ReportCardConcluded
	case s.Ended:
		return reportCardFactsFromRubric(terms), ReportCardConcluded
	}
	return reportCardFactsFromRubric(terms), ReportCardLive
}

// RecordedPassFor finds the exercise's retained CurriculumPass on state, or
// nil — the resolver's instrument lookup, shared by the console card, the
// postmortem, and the CLI duel.
func RecordedPassFor(s *sim.State, exercise string) *sim.CurriculumPass {
	if s == nil {
		return nil
	}
	for i := range s.CurriculumPasses {
		if s.CurriculumPasses[i].Exercise == exercise {
			return &s.CurriculumPasses[i]
		}
	}
	return nil
}

// RenderReportCard is the exported face of reportCardView (views.go) — the
// bordered ✓/✗/… card every surface (and the duel) renders through, so the
// duel card and the postmortem card are the same artifact by construction.
func RenderReportCard(title string, facts []ReportCardFact, mode ReportCardMode, width int) string {
	return reportCardView(title, facts, mode, width)
}

// resolveReportCardFacts is the Model-side thin wrapper over
// ResolveRubricFacts (spec 076 D5): m.replica drives the switch, and the
// old m.runEnded() conjunct folds into the state-driven s.Ended read — the
// replica IS where runEnded() learns the run is over (the status half of
// its dual-source posture only bridges the moments before a replica
// exists, when this resolver renders nothing anyway).
func (m Model) resolveReportCardFacts(def sim.ExerciseDefinition, pass *sim.CurriculumPass) ([]reportCardFact, reportCardMode) {
	return ResolveRubricFacts(m.replica, def, pass)
}

// recordedPassFor is the Model-side thin wrapper over RecordedPassFor.
func (m Model) recordedPassFor(exercise string) *sim.CurriculumPass {
	return RecordedPassFor(m.replica, exercise)
}

// buildChecklistCard resolves the rubric-checklist half of the card (spec
// 063 standing resolution 1) through TASK-127's shared renderer: the seeded
// exercise's evaluated rubric terms as reportCardFacts, wrapped as the
// spec-056 reportCard consoleCard. nil on an ambient world (no rubric
// exists) and — the stopping-point discipline — nil until a stopping point
// is visible in durable state: a stored attribution note, a recorded pass
// for this exercise (exercise resolution), or the ended run. Facts and
// marker vocabulary come from the shared resolver (spec 072): the recorded
// pass when one exists, else sim.EvaluateRubric — live (met/pending) until
// the exercise concludes or the run ends.
func (m *Model) buildChecklistCard() *reportCard {
	def, ok := m.scenarioExercise()
	if !ok || m.replica == nil {
		return nil
	}
	pass := m.recordedPassFor(def.ID)
	if pass == nil && m.replica.GuardianReportCard == nil && !m.runEnded() {
		return nil // no stopping point on the record yet — no card, no theater
	}
	facts, mode := m.resolveReportCardFacts(def, pass)
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
