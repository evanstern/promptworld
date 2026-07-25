package tui

// The exercise tab (spec 054 US4, docs/design/tui/panels/exercise.md as
// authored): the framing, attach-time briefing, live rubric gauges,
// incident-visibility line, and pass/fail banner of the attached scenario
// world's seeded exercise. Present only on scenario worlds (world-shaped,
// not stage-shaped — patterns/stage-defaults.md); everything renders from
// the replica plus the compiled definition (sim.ExerciseByID) — same data
// every other tab reads, no extra IPC.

import (
	"fmt"
	"strings"

	"github.com/evanstern/promptworld/internal/sim"
)

// exerciseView is the narrow-fallback wrapper (the villagersView shape):
// the exercise tab is reachable as a solo/narrow pane exactly like every
// other dock tab, no narrow-specific rendering (panels/exercise.md).
func (m Model) exerciseView() string {
	body := m.exerciseBody(clampInt(m.width-6, 20, 500), clampInt(m.height-6, 4, 500))
	return styleBox.Render(body)
}

// exerciseBody renders the tab content: the briefing until dismissed (once
// per attach), then title/banner + gauge rows + the incident line per the
// visibility vocabulary.
func (m Model) exerciseBody(width, height int) string {
	id := m.exerciseID()
	if id == "" {
		return "" // unreachable: the tab only exists on scenario worlds
	}
	def, _ := sim.ExerciseByID(id)
	if m.exerciseBriefingShowing() {
		return m.exerciseBriefingBody(def, width)
	}
	if m.replica == nil {
		return styleHeader.Render(exerciseTitle(def.ID)) + "\n\n" + styleDim.Render("waiting for world state…")
	}

	outcome := m.exerciseOutcome(id)
	title := fmt.Sprintf("%s · %s", exerciseTitle(def.ID), exerciseOutcomeLabel(outcome))
	lines := []string{styleHeader.Render(title), ""}

	// Rubric gauges (contract §4): one row per evaluated term — plain
	// language, met/pending marker, and the backing event count — live at
	// every stage, the same sim.EvaluateRubric derivation the executor's
	// pass precondition reads, over the replica.
	for _, term := range sim.EvaluateRubric(m.replica, def, m.replica.Tick) {
		marker := styleDim.Render("…")
		if term.Met {
			marker = styleTabOn.Render("✓")
		}
		lines = append(lines, fmt.Sprintf("%s %s  %s", marker, term.Label,
			styleDim.Render(fmt.Sprintf("(%s: %d)", term.Event, term.Count))))
	}

	// Incident line per the visibility vocabulary (D4): forecast lists the
	// authored schedule with approximate times; fog omits the line entirely
	// — the gauges section never changes shape between the two.
	if line := exerciseIncidentLine(def, m.w.Manifest.Stage); line != "" {
		lines = append(lines, "", styleDim.Render(line))
	}

	// Pass/fail banner (contract §4): replaces the in-progress posture.
	switch outcome {
	case sim.OutcomePassed:
		lines = append(lines, "", styleTabOn.Render("✓ PASSED — "+def.ScoreNarrative))
	case sim.OutcomeFailed:
		lines = append(lines, "", styleFeedAlert.Render("✗ FAILED (run ended) — the postmortem carries the record"))
	}
	return clipContent(strings.Join(lines, "\n"), width)
}

// exerciseBriefingBody is the attach-time briefing (panels/exercise.md
// mockup 1): framing + this stage's incident-visibility mode, dismissed by
// any key (handleKey consumes exactly one press while this is visible).
func (m Model) exerciseBriefingBody(def sim.ExerciseDefinition, width int) string {
	lines := []string{styleHeader.Render(exerciseTitle(def.ID)), ""}
	lines = append(lines, wrapText(def.Framing, width)...)
	lines = append(lines, "")
	if sim.IncidentVisibilityFor(def, m.w.Manifest.Stage) == sim.VisibilityForecast {
		lines = append(lines, wrapText("Incidents ahead are forecast — you'll see the schedule before they land.", width)...)
	} else {
		lines = append(lines, wrapText("Incidents ahead are under fog — you'll meet them as they happen.", width)...)
	}
	lines = append(lines, "", styleDim.Render("any key — begin"))
	return clipContent(strings.Join(lines, "\n"), width)
}

// exerciseOutcome is the banner's dual-source derivation: the replica's
// recorded facts (sim.ExerciseOutcome) plus the status poll's Ended for a
// live transition the snapshot hasn't folded yet — the runEnded() posture,
// applied to the exercise's own outcome.
func (m Model) exerciseOutcome(id string) string {
	out := sim.ExerciseOutcome(m.replica, id)
	if out == sim.OutcomeInProgress && m.runEnded() {
		return sim.OutcomeFailed
	}
	return out
}

// exerciseIncidentLine renders the forecast line, or "" under fog (the line
// is omitted, never blanked — panels/exercise.md structure item 4).
func exerciseIncidentLine(def sim.ExerciseDefinition, stage string) string {
	if sim.IncidentVisibilityFor(def, stage) != sim.VisibilityForecast || len(def.Schedule) == 0 {
		return ""
	}
	parts := make([]string, 0, len(def.Schedule))
	for _, e := range def.Schedule {
		parts = append(parts, fmt.Sprintf("%s ~%s (day %d)", incidentNoun(e.Kind), e.Time, e.Day))
	}
	return "incidents (forecast): " + strings.Join(parts, " · ")
}

// incidentNoun is the closed incident vocabulary's player-facing phrasing
// (FR-020 audience ruling: plain language by default).
func incidentNoun(kind string) string {
	switch kind {
	case sim.IncidentGruEmerges:
		return "the gru emerges"
	}
	return strings.ReplaceAll(kind, "_", " ")
}

// exerciseTitle renders the exercise id as the panel's display name
// ("first-night" → "FIRST NIGHT").
func exerciseTitle(id string) string {
	return strings.ToUpper(strings.ReplaceAll(id, "-", " "))
}

// exerciseOutcomeLabel is the title's posture word (contract §4).
func exerciseOutcomeLabel(outcome string) string {
	switch outcome {
	case sim.OutcomePassed:
		return "passed"
	case sim.OutcomeFailed:
		return "failed (run ended)"
	}
	return "in progress"
}
