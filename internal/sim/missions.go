package sim

// The guardian's mission layer (spec 107): a MISSION is the player's plain-
// words standing instruction made durable — accepted through guardian chat,
// decomposed and pursued via the EXISTING plan-layer verbs (D2: the mission
// artifact only records intent and derived progress; designations/directives/
// miracles stay the acting vocabulary), completed or failed by DERIVED
// predicates over recorded events (D3: never model prose). The entity clones
// the spec-084 Designation discipline verbatim (D1): deterministic human-
// readable ids minted guardian-side with no RNG, one-way status transitions
// resolved at the reducer door, payload Status/PlacedSeq ignored and reducer-
// stamped, caps validated in the arm (validate-not-clamp), and the shared
// active+recent-32 retention prune.
//
// Door split (the spec-084 R2 contract): guardian.mission_accepted /
// mission_progressed / mission_cancelled are INJECTED through InjectSocial
// (whitelist + these validating arms); guardian.mission_completed and
// guardian.mission_failed are EXECUTOR-emitted from stepEvents as pure
// functions of (state, tick) — the designation.fulfilled pattern — so replay
// reproduces the whole lifecycle with no guardian running, and whitelist
// absence refuses any injected forgery of a derived outcome.
//
// Doctrine (spec 107, not re-litigated here): a mission is durable
// PRE-AUTHORIZATION — the same legal shape as a standing order — so pursuit
// runs at FULL competence at any spec-102 ceiling (the ceiling caps
// initiative; a mission is the player's explicit instruction, not
// initiative). The grant composition lives guardian-side (ceiling.go).

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/evanstern/promptworld/internal/store"
)

const (
	// GuardianMissionCap bounds concurrent ACTIVE missions (spec 107 FR-001)
	// — the GuardianDirectiveCap number, same rationale: a bounded attention
	// economy for standing player instructions.
	GuardianMissionCap = 3
	// Mission TTL bounds, in game days: a mission spans multiple
	// designation/directive cycles, so its ceiling is double the directive
	// band (1..7); the default (7) is applied at the tool door (the
	// monitor_and_act TTL-default shape) — the reducer only validates.
	GuardianMissionTTLMinDays = 1
	GuardianMissionTTLMaxDays = 14
)

// missionGoalMaxRunes caps the guardian's goal rendering (FR-001) — the same
// 400-rune band the directive text carries. The goal is the GUARDIAN's own
// restatement in completion-predicate vocabulary where checkable (D3), never
// the player's literal words (the persona-firewall non-negotiable).
const missionGoalMaxRunes = 400

// Mission failure reasons — FROZEN recorded vocabulary (spec 052 ruling 2):
// they land in guardian.mission_failed payloads.
const (
	// MissionFailDeadline: the deadline elapsed with the completion
	// predicate unmet — linked work exists but did not fulfill in time.
	MissionFailDeadline = "deadline_unmet"
	// MissionFailNeverPursued: the deadline elapsed with NO linked
	// designation at all — the mission was accepted but never decomposed.
	MissionFailNeverPursued = "never_pursued"
)

// Mission is one event-sourced standing player instruction (spec 107 FR-001):
// durable, checkable, pursued across scheduled turns with no player in the
// loop. Designations/Directives accumulate from guardian.mission_progressed
// links — the completion predicate reads ONLY those linked entities, so
// progress is recorded intent, never inference. Its lifecycle
// (active → completed | failed | cancelled) is driven entirely by recorded
// events, so it reconstructs identically through snapshots, restart, replay.
type Mission struct {
	ID           string `json:"id"`   // "msn-<acceptedTick>-<seq>" (nextOrderID shape, no RNG)
	Goal         string `json:"goal"` // guardian's rendering, 1..400 runes
	AcceptedTick int64  `json:"accepted_tick"`
	DeadlineTick int64  `json:"deadline_tick"` // accepted + ttl_days game days; bounds 1..14
	// Linked pursuit artifacts (mission_progressed): designation/directive
	// ids in link order, deduplicated at the door. The completion predicate
	// requires every linked designation fulfilled (≥1 linked).
	Designations []string `json:"designations,omitempty"`
	Directives   []string `json:"directives,omitempty"`
	Status       string   `json:"status"`               // "active" | "completed" | "failed" | "cancelled" (one-way)
	PlacedSeq    int64    `json:"placed_seq,omitempty"` // reducer-stamped (the Designation.PlacedSeq contract)
}

// MissionProgressedPayload — guardian.mission_progressed: one recorded
// pursuit step. At least one of designation_id / directive_id / note is
// present (door-validated); linked ids must name entities that exist in
// state, so a link is always a real artifact, never a claim.
type MissionProgressedPayload struct {
	ID            string `json:"id"`
	DesignationID string `json:"designation_id,omitempty"`
	DirectiveID   string `json:"directive_id,omitempty"`
	Note          string `json:"note,omitempty"` // guardian framing, ≤400 runes
}

// MissionCompletedPayload — guardian.mission_completed (executor-emitted):
// the derived success terminal. Designations carries the fulfilled linked
// ids — the evidence trail the report card cites (D3).
type MissionCompletedPayload struct {
	ID           string   `json:"id"`
	AcceptedTick int64    `json:"accepted_tick"`
	Designations []string `json:"designations"`
}

// MissionEntityStatus is one linked entity's status at failure emission —
// recorded-event evidence of the blocker (statuses ARE derived from recorded
// events), never prose grading.
type MissionEntityStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// MissionFailedPayload — guardian.mission_failed (executor-emitted): the
// honest-failure terminal, carrying the frozen reason and every linked
// entity's status at the deadline.
type MissionFailedPayload struct {
	ID           string                `json:"id"`
	Reason       string                `json:"reason"`
	AcceptedTick int64                 `json:"accepted_tick"`
	DeadlineTick int64                 `json:"deadline_tick"`
	Designations []MissionEntityStatus `json:"designations,omitempty"`
	Directives   []MissionEntityStatus `json:"directives,omitempty"`
}

// missionByID returns the mission named id, or nil. Linear over the bounded
// slice (active ≤ 3 + retained 32), matching every other entity scan.
func (s *State) missionByID(id string) *Mission {
	for i := range s.Missions {
		if s.Missions[i].ID == id {
			return &s.Missions[i]
		}
	}
	return nil
}

// MissionByID is the exported lookup (the DesignationByID shape) — the TUI
// and guardian read linked-entity state from it.
func (s *State) MissionByID(id string) *Mission { return s.missionByID(id) }

// missionFulfilled is the mission completion predicate (spec 107 D3): a PURE
// state check — no clock, no RNG, no I/O, deterministic iteration (link
// order) — evaluated IDENTICALLY by the executor sweep (to emit
// guardian.mission_completed) and by the reducer arm (to validate before
// transitioning), so live and replay agree by construction. A mission with
// no linked designation is never fulfilled: intent alone completes nothing.
// A linked designation pruned out of retention reads as unfulfilled (nil
// lookup) — deterministic either way, since the prune is itself a pure
// function of the event order.
func missionFulfilled(s *State, m *Mission) bool {
	if len(m.Designations) == 0 {
		return false
	}
	for _, id := range m.Designations {
		d := s.designationByID(id)
		if d == nil || d.Status != "fulfilled" {
			return false
		}
	}
	return true
}

// missionFailedEvidence assembles the failure payload's linked-entity status
// evidence from state — deterministic (link order).
func missionFailedEvidence(s *State, m *Mission) ([]MissionEntityStatus, []MissionEntityStatus) {
	var dsgs, dirs []MissionEntityStatus
	for _, id := range m.Designations {
		st := "pruned"
		if d := s.designationByID(id); d != nil {
			st = d.Status
		}
		dsgs = append(dsgs, MissionEntityStatus{ID: id, Status: st})
	}
	for _, id := range m.Directives {
		st := "pruned"
		if d := s.directiveByID(id); d != nil {
			st = d.Status
		}
		dirs = append(dirs, MissionEntityStatus{ID: id, Status: st})
	}
	return dsgs, dirs
}

// applyMission is the reducer arm family for the guardian.mission_* event
// vocabulary (spec 107 FR-001). Every arm validates rather than clamps — the
// InjectSocial dry-run runs these on a state copy, so an invalid acceptance/
// link is rejected at the door and recorded events always re-apply cleanly at
// the same position in replay (the applyPlan contract). The two executor-
// emitted terminals re-validate their emitting condition here, keeping the
// door authoritative even for sim-authored events.
func (s *State) applyMission(e store.Event) error {
	switch e.Type {
	case "guardian.mission_accepted":
		var m Mission
		if err := json.Unmarshal(e.Payload, &m); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if err := s.validateMissionAccepted(e.Type, &m); err != nil {
			return err
		}
		// Status is IGNORED on the payload — a mission always lands active;
		// PlacedSeq is reducer-stamped from the event's own store seq (the
		// designation.placed shape). Links are reducer-owned: acceptance
		// never carries them (validated empty above).
		m.Status = "active"
		m.PlacedSeq = e.Seq
		s.Missions = prunePlanEntities(append(s.Missions, m),
			func(x Mission) bool { return x.Status == "active" })
	case "guardian.mission_progressed":
		var p MissionProgressedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		return s.applyMissionProgressed(e.Type, &p)
	case "guardian.mission_completed":
		var p MissionCompletedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		// Re-validate the predicate against current state before
		// transitioning (the designation.fulfilled contract — the door stays
		// authoritative on the structural fact, even for the executor's own
		// emission).
		m := s.missionByID(p.ID)
		if m == nil {
			return fmt.Errorf("apply %s: unknown mission %q", e.Type, p.ID)
		}
		if !missionFulfilled(s, m) {
			return fmt.Errorf("apply %s: mission %q predicate does not hold", e.Type, p.ID)
		}
		return s.transitionMission(e.Type, p.ID, "completed")
	case "guardian.mission_failed":
		var p MissionFailedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		// Re-validate the emitting condition: deadline elapsed AND the
		// completion predicate does not hold (the directive.expired shape).
		m := s.missionByID(p.ID)
		if m == nil {
			return fmt.Errorf("apply %s: unknown mission %q", e.Type, p.ID)
		}
		if e.Tick < m.DeadlineTick {
			return fmt.Errorf("apply %s: mission %q is not past its deadline", e.Type, p.ID)
		}
		if missionFulfilled(s, m) {
			return fmt.Errorf("apply %s: mission %q predicate holds — it completed", e.Type, p.ID)
		}
		return s.transitionMission(e.Type, p.ID, "failed")
	case "guardian.mission_cancelled":
		var p OrderIDPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		return s.transitionMission(e.Type, p.ID, "cancelled")
	}
	return nil
}

// validateMissionAccepted is the guardian.mission_accepted door (validate-
// not-clamp): id discipline, goal runes, TTL bounds, no pre-linked entities,
// and the active cap.
func (s *State) validateMissionAccepted(eventType string, m *Mission) error {
	if m.ID == "" {
		return fmt.Errorf("apply %s: empty mission id", eventType)
	}
	// Duplicate id in ANY status is rejected — ids are assigned once and
	// consumed missions are retained (the designation-placed discipline).
	for i := range s.Missions {
		if s.Missions[i].ID == m.ID {
			return fmt.Errorf("apply %s: duplicate mission id %q", eventType, m.ID)
		}
	}
	if n := utf8.RuneCountInString(m.Goal); n == 0 || n > missionGoalMaxRunes {
		return fmt.Errorf("apply %s: goal length %d outside 1..%d runes", eventType, n, missionGoalMaxRunes)
	}
	if len(m.Designations) != 0 || len(m.Directives) != 0 {
		return fmt.Errorf("apply %s: acceptance carries links — links land only through %s", eventType, "guardian.mission_progressed")
	}
	if ttl := m.DeadlineTick - m.AcceptedTick; ttl < GuardianMissionTTLMinDays*ticksPerGameDay || ttl > GuardianMissionTTLMaxDays*ticksPerGameDay {
		return fmt.Errorf("apply %s: ttl %d ticks outside %d..%d game days", eventType, ttl, GuardianMissionTTLMinDays, GuardianMissionTTLMaxDays)
	}
	active := 0
	for i := range s.Missions {
		if s.Missions[i].Status == "active" {
			active++
		}
	}
	if active >= GuardianMissionCap {
		return fmt.Errorf("apply %s: %d missions already active (cap %d)", eventType, active, GuardianMissionCap)
	}
	return nil
}

// applyMissionProgressed is the guardian.mission_progressed door: the mission
// must be active; at least one of link/note present; a linked id must name an
// entity that exists in state (any status — the guardian may link work that
// already fulfilled: consecration, the pre-existing-structure precedent) and
// must not already be linked; the note rides the directive-text rune band.
func (s *State) applyMissionProgressed(eventType string, p *MissionProgressedPayload) error {
	m := s.missionByID(p.ID)
	if m == nil {
		return fmt.Errorf("apply %s: unknown mission %q", eventType, p.ID)
	}
	if m.Status != "active" {
		return fmt.Errorf("apply %s: mission %q is not active (status %q)", eventType, p.ID, m.Status)
	}
	if p.DesignationID == "" && p.DirectiveID == "" && p.Note == "" {
		return fmt.Errorf("apply %s: empty progress — link a designation/directive or record a note", eventType)
	}
	if n := utf8.RuneCountInString(p.Note); n > directiveTextMaxRunes {
		return fmt.Errorf("apply %s: note over %d runes", eventType, directiveTextMaxRunes)
	}
	if p.DesignationID != "" {
		if s.designationByID(p.DesignationID) == nil {
			return fmt.Errorf("apply %s: unknown designation %q", eventType, p.DesignationID)
		}
		for _, id := range m.Designations {
			if id == p.DesignationID {
				return fmt.Errorf("apply %s: designation %q already linked to %s", eventType, p.DesignationID, p.ID)
			}
		}
	}
	if p.DirectiveID != "" {
		if s.directiveByID(p.DirectiveID) == nil {
			return fmt.Errorf("apply %s: unknown directive %q", eventType, p.DirectiveID)
		}
		for _, id := range m.Directives {
			if id == p.DirectiveID {
				return fmt.Errorf("apply %s: directive %q already linked to %s", eventType, p.DirectiveID, p.ID)
			}
		}
	}
	if p.DesignationID != "" {
		m.Designations = append(m.Designations, p.DesignationID)
	}
	if p.DirectiveID != "" {
		m.Directives = append(m.Directives, p.DirectiveID)
	}
	// A note-only progress event mutates nothing on the entity: the recorded
	// event IS the note's durable home (the decision-trail reads it there).
	return nil
}

// transitionMission moves the mission named id from active to a terminal
// status (one-way) — the transitionDesignation race shape: exactly one
// terminal lands, and the loser hits a non-active entity and refuses.
func (s *State) transitionMission(eventType, id, to string) error {
	for i := range s.Missions {
		if s.Missions[i].ID != id {
			continue
		}
		if s.Missions[i].Status != "active" {
			return fmt.Errorf("apply %s: mission %q is not active (status %q)", eventType, id, s.Missions[i].Status)
		}
		s.Missions[i].Status = to
		return nil
	}
	return fmt.Errorf("apply %s: unknown mission %q", eventType, id)
}

// missionEvents is the mission outcome sweep (spec 107 FR-003), called from
// stepEvents beside the designation/directive/prophecy sweeps: per active
// mission, completion is checked BEFORE failure so a mission eligible for
// both at one boundary lands exactly ONE terminal (completed wins — the work
// was done). Both pure over (pre-tick state, tick), each fired once by the
// same flips-non-active argument as every other sweep. Sweep position is
// FIXED after the directive sweep: a designation fulfilled at T flips status
// at apply, so a dependent mission completes at T+1's sweep — the documented
// one-tick lag, never an order-dependent same-tick race.
func missionEvents(s *State, nextTick int64) []store.Event {
	var events []store.Event
	emit := func(typ string, payload any) {
		b, _ := json.Marshal(payload)
		events = append(events, store.Event{Tick: nextTick, Type: typ, Payload: b})
	}
	for i := range s.Missions {
		m := &s.Missions[i]
		if m.Status != "active" {
			continue
		}
		if missionFulfilled(s, m) {
			emit("guardian.mission_completed", MissionCompletedPayload{
				ID: m.ID, AcceptedTick: m.AcceptedTick,
				Designations: append([]string(nil), m.Designations...)})
			continue
		}
		if nextTick >= m.DeadlineTick {
			reason := MissionFailDeadline
			if len(m.Designations) == 0 {
				reason = MissionFailNeverPursued
			}
			dsgs, dirs := missionFailedEvidence(s, m)
			emit("guardian.mission_failed", MissionFailedPayload{
				ID: m.ID, Reason: reason, AcceptedTick: m.AcceptedTick,
				DeadlineTick: m.DeadlineTick, Designations: dsgs, Directives: dirs})
		}
	}
	return events
}
