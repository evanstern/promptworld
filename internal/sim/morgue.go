package sim

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/evanstern/promptworld/internal/store"
)

// The morgue's narrated epilogues (spec 044 US2): prose the narrator writes
// after a death or the run's end. Like chronicle entries, epilogues are
// events — the narrator's output enters the world only through the
// inject_social door (and it is one of the two prose types an ENDED world's
// door still accepts, endedProseWhitelist) — and the reducer keeps a bounded
// ring on State so the scribe replica and any attaching client can read them
// from state. The morgue document's FACTS never depend on these: narrator
// absence or failure is a gap in the prose, never a stall (FR-010).

// MorgueEpiloguePayload is the morgue.epilogue event payload. Agent is the
// villager the epilogue mourns, or -1 for the run-end epilogue. Text is
// bounded like chronicle lines (the narrator caps it at emission; the arm
// refuses an empty one).
type MorgueEpiloguePayload struct {
	Agent AgentRef `json:"agent"`
	Text  string   `json:"text"`
}

// MorgueEpilogue is one recorded epilogue in the State ring.
type MorgueEpilogue struct {
	Tick  int64  `json:"tick"` // when the epilogue landed
	Agent int    `json:"agent"`
	Text  string `json:"text"`
}

// morgueEpilogueCap bounds State.MorgueEpilogues. A run produces at most one
// epilogue per death plus one for the run end (≤ agentCount+1 in practice);
// the cap only guards the ring against a misbehaving narrator re-mourning.
const morgueEpilogueCap = 32

func (s *State) applyMorgueEpilogue(e store.Event) error {
	var p MorgueEpiloguePayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return fmt.Errorf("apply %s: %w", e.Type, err)
	}
	if p.Agent.ID < -1 || p.Agent.ID >= len(s.Agents) {
		return fmt.Errorf("apply %s: agent %d out of range", e.Type, p.Agent.ID)
	}
	if strings.TrimSpace(p.Text) == "" {
		return fmt.Errorf("apply %s: empty text", e.Type)
	}
	s.MorgueEpilogues = append(s.MorgueEpilogues, MorgueEpilogue{Tick: e.Tick, Agent: p.Agent.ID, Text: p.Text})
	if len(s.MorgueEpilogues) > morgueEpilogueCap {
		s.MorgueEpilogues = append(s.MorgueEpilogues[:0], s.MorgueEpilogues[len(s.MorgueEpilogues)-morgueEpilogueCap:]...)
	}
	return nil
}
