package sim

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/evanstern/promptworld/internal/store"
)

// The report card's world-visible surface (spec 063 US4): the stored
// attribution note. The card is one composed artifact — the deterministic
// rubric checklist (event-derived, TASK-127's renderer) plus this note, the
// cheap-chain critique tying outcomes to charter text with event citations.
// The note is event-sourced through the inject_social door and STORED ONCE:
// re-opening a card re-reads the recorded note, never re-grades (FR-006).
// Production (stopping-point triggers, chain call, citation validation)
// lives daemon-side in internal/guardian; only the recorded fact reaches
// here. The run-end card rides the existing morgue.epilogue channel instead
// (the ended door narrows to recorded prose), so this arm never needs the
// ended exception.

// reportCardNoteMaxRunes bounds the stored note — the chronicle-line class
// of prose cap; the producer's token budget keeps real notes far under it.
const reportCardNoteMaxRunes = 1200

// GuardianReportCardPayload is the guardian.report_card event payload:
// the charter revision the critique graded under (the fingerprint-at-effect
// discipline, the charter_observed precedent), the note's prose, and the
// event seqs it cites — validated against the recorded trail by the
// producer BEFORE injection, and re-validated structurally here.
type GuardianReportCardPayload struct {
	Fingerprint string  `json:"fingerprint"`
	Note        string  `json:"note"`
	Citations   []int64 `json:"citations,omitempty"`
}

// GuardianReportCard is the latest stored card on State — what the console
// card seam composes (beside the rubric checklist) and what a late-attaching
// client reads.
type GuardianReportCard struct {
	Tick        int64   `json:"tick"` // when the card landed
	Seq         int64   `json:"seq"`  // the card event's own seq (identity for the unseen badge)
	Fingerprint string  `json:"fingerprint"`
	Note        string  `json:"note"`
	Citations   []int64 `json:"citations,omitempty"`
}

// applyReportCard is the reducer arm for guardian.report_card: validates
// rather than clamps (the nudged-arm contract — the InjectSocial dry-run
// runs this on a state copy, so an invalid card is refused at the door and
// a recorded one always re-applies cleanly in replay), then keeps the
// LATEST card on state; the log keeps history.
func (s *State) applyReportCard(e store.Event) error {
	var p GuardianReportCardPayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return fmt.Errorf("apply %s: %w", e.Type, err)
	}
	if p.Fingerprint == "" {
		return fmt.Errorf("apply %s: empty charter fingerprint", e.Type)
	}
	if strings.TrimSpace(p.Note) == "" {
		return fmt.Errorf("apply %s: empty note", e.Type)
	}
	if n := len([]rune(p.Note)); n > reportCardNoteMaxRunes {
		return fmt.Errorf("apply %s: note %d runes over the %d cap", e.Type, n, reportCardNoteMaxRunes)
	}
	// Citations must be recorded history: every cited seq precedes this
	// event's own position. The producer validated them against the actual
	// trail; this structural check keeps the door authoritative that a card
	// can never cite the future.
	for _, c := range p.Citations {
		if c <= 0 || (e.Seq > 0 && c >= e.Seq) {
			return fmt.Errorf("apply %s: citation seq %d is not recorded history", e.Type, c)
		}
	}
	s.GuardianReportCard = &GuardianReportCard{
		Tick: e.Tick, Seq: e.Seq,
		Fingerprint: p.Fingerprint, Note: p.Note,
		Citations: append([]int64(nil), p.Citations...),
	}
	return nil
}
