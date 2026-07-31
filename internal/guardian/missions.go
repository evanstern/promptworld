package guardian

// The mission layer's tool handlers (spec 107): accept_mission /
// note_mission_progress / cancel_mission. Doctrine is the plan-layer file's
// (plans.go): the reducer dry-run (internal/sim/missions.go) is the door
// authority for every lifecycle rule — goal, TTL, cap, link existence, the
// one-way races — and these helpers only parse the call surface, mint
// deterministic ids, land through InjectSocial, and map a door rejection to
// in-fiction counsel the loop feeds back as a rejected_gate.
//
// Doctrine carried here (spec 107, ratified — not re-litigated):
//   - A mission is durable PRE-AUTHORIZATION, the standing order's legal
//     shape: pursuit is the player's own instruction, so it runs at FULL
//     competence at any spec-102 ceiling (ceiling.go's pursuit grant).
//   - D2: decomposition through EXISTING verbs only — the mission verbs
//     record intent and derived progress; designations/directives/surveys/
//     workings stay the whole acting vocabulary. The one additive hook is
//     place_designation/issue_directive's optional mission_id (plans.go),
//     which links the act atomically so pursuit never costs a second act.
//   - D5's one sanctioned counsel case: an IMPOSSIBLE-as-stated mission is
//     refused at acceptance, naming the blocking fact — charter + tool-gloss
//     doctrine (the model refuses by not calling accept_mission), with the
//     door's own validation behind it.

import (
	"context"
	"fmt"
	"strings"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/toolloop"
)

// MissionReport is the console-facing summary of a landed mission act (spec
// 107): the mission id the player can name to cancel it, and a one-line
// human rendering. Additive omitempty on TurnResult, the PlanReport shape.
type MissionReport struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
}

// landAcceptMission parses an accept_mission call, mints the id, and lands
// guardian.mission_accepted through the door (spec 107 FR-001). The goal is
// the GUARDIAN's own rendering — the player's literal words never enter a
// recorded payload (the persona-firewall non-negotiable; the directive-text
// precedent). Returns the accepted mission or (nil, in-fiction refusal).
func (mt *Guardian) landAcceptMission(goal string, ttlDays int, tick int64, grant grantSet) (*sim.Mission, string) {
	if !grant.allows("accept_mission") {
		return nil, "that power is not granted in this world"
	}
	goal = strings.TrimSpace(goal)
	if goal == "" {
		return nil, "give me the mission in words I can check the world against — what should stand true when it is done?"
	}
	if ttlDays == 0 {
		ttlDays = 7 // the tool-door default (the monitor_and_act TTL shape)
	}
	if ttlDays < sim.GuardianMissionTTLMinDays || ttlDays > sim.GuardianMissionTTLMaxDays {
		return nil, fmt.Sprintf("a mission may stand for %d to %d days", sim.GuardianMissionTTLMinDays, sim.GuardianMissionTTLMaxDays)
	}
	m := sim.Mission{
		ID:           mt.nextPlanID("msn", tick),
		Goal:         goal,
		AcceptedTick: tick,
		DeadlineTick: tick + int64(ttlDays)*ticksPerGameDay,
	}
	batch := []store.Event{{Type: "guardian.mission_accepted", Payload: mustJSON(m)}}
	if err := mt.social.InjectSocial(batch); err != nil {
		return nil, missionRefusal(err)
	}
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — I accepted a mission (%s): %q\n",
		clock.Format(mt.replicaTickSafe()), m.ID, goal))
	// Agentized memory (spec 102 D5): the standing instruction enters the
	// guardian's own store — fixed mechanics vocabulary (the event log and
	// memory payloads are skin-free, ruling 1).
	mt.recordMemory(fmt.Sprintf("I accepted a mission (%s): %q", m.ID, goal), salGuardianAct)
	return &m, ""
}

// landNoteMissionProgress lands guardian.mission_progressed for an explicit
// link/note step (spec 107 D2). The atomic mission_id hook on
// place_designation/issue_directive covers the common pursuit path; this
// verb links pre-existing work (consecration) or records an obstacle.
func (mt *Guardian) landNoteMissionProgress(id, designationID, directiveID, note string, grant grantSet) string {
	if !grant.allows("note_mission_progress") {
		return "that power is not granted in this world"
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return "name the mission the progress belongs to"
	}
	p := sim.MissionProgressedPayload{
		ID:            id,
		DesignationID: strings.TrimSpace(designationID),
		DirectiveID:   strings.TrimSpace(directiveID),
		Note:          strings.TrimSpace(note),
	}
	if p.DesignationID == "" && p.DirectiveID == "" && p.Note == "" {
		return "record something — a designation or directive serving the mission, or a note on the obstacle"
	}
	batch := []store.Event{{Type: "guardian.mission_progressed", Payload: mustJSON(p)}}
	if err := mt.social.InjectSocial(batch); err != nil {
		return missionRefusal(err)
	}
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — I recorded progress on %s\n",
		clock.Format(mt.replicaTickSafe()), id))
	return ""
}

// landCancelMission lands guardian.mission_cancelled for the named id — the
// player's stand-down. The reducer resolves the cancel/terminal race.
func (mt *Guardian) landCancelMission(id string, grant grantSet) string {
	if !grant.allows("cancel_mission") {
		return "that power is not granted in this world"
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return "name the mission you want me to stand down from"
	}
	batch := []store.Event{{Type: "guardian.mission_cancelled", Payload: mustJSON(sim.OrderIDPayload{ID: id})}}
	if err := mt.social.InjectSocial(batch); err != nil {
		return missionRefusal(err)
	}
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — I stood down from the mission %s\n",
		clock.Format(mt.replicaTickSafe()), id))
	return ""
}

// missionRefusal maps a guardian.mission_* door rejection to in-fiction
// counsel (the planRefusal shape): the reducer's error strings are the
// source, translated so the model hears a repairable reason.
func missionRefusal(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "cap"):
		return "I already carry as many missions as I can hold true — release or finish one and I will take another"
	case strings.Contains(msg, "unknown mission"):
		return "I keep no mission by that name"
	case strings.Contains(msg, "is not active"):
		return "that mission has already run its course"
	case strings.Contains(msg, "unknown designation"):
		return "I keep no designation by that name to link"
	case strings.Contains(msg, "unknown directive"):
		return "I keep no directive by that name to link"
	case strings.Contains(msg, "already linked"):
		return "that work is already counted toward the mission"
	case strings.Contains(msg, "goal length"):
		return "state the mission's goal more briefly and I will hold it"
	case strings.Contains(msg, "ttl"):
		return fmt.Sprintf("a mission may stand for %d to %d days", sim.GuardianMissionTTLMinDays, sim.GuardianMissionTTLMaxDays)
	case strings.Contains(msg, "not past its deadline"), strings.Contains(msg, "predicate"):
		return "the mission's outcome is the world's to judge, not mine to declare"
	default:
		return "the world would not let me (" + msg + ")"
	}
}

// --- tool-use loop handlers (the toolcalls.go dispatch shape) ---

// handleAcceptMission wraps landAcceptMission: door accept → landed (a
// MissionReport on the shared result); refusal → rejected_gate the model may
// repair within the round cap.
func (mt *Guardian) handleAcceptMission(d *turnDispatch) toolloop.Handler {
	return func(_ context.Context, call llm.ToolCall) toolloop.Outcome {
		ttl, _ := argInt(call.Args, "ttl_days")
		m, why := mt.landAcceptMission(argString(call.Args, "goal"), ttl, d.tick, d.grant)
		if m == nil {
			return toolloop.Outcome{Verdict: toolloop.VerdictRejectedGate, ResultForModel: refusal(why)}
		}
		d.result.Mission = &MissionReport{ID: m.ID,
			Summary: fmt.Sprintf("accepted a mission: %q", m.Goal)}
		return toolloop.Outcome{Verdict: toolloop.VerdictLanded,
			ResultForModel: "the mission is mine (" + m.ID + ") — I will pursue it on my own watches and report what the world records"}
	}
}

// handleNoteMissionProgress wraps landNoteMissionProgress.
func (mt *Guardian) handleNoteMissionProgress(d *turnDispatch) toolloop.Handler {
	return func(_ context.Context, call llm.ToolCall) toolloop.Outcome {
		id := argString(call.Args, "id")
		why := mt.landNoteMissionProgress(id,
			argString(call.Args, "designation_id"), argString(call.Args, "directive_id"),
			argString(call.Args, "note"), d.grant)
		if why != "" {
			return toolloop.Outcome{Verdict: toolloop.VerdictRejectedGate, ResultForModel: refusal(why)}
		}
		d.result.Mission = &MissionReport{ID: id, Summary: "recorded progress on " + id}
		return toolloop.Outcome{Verdict: toolloop.VerdictLanded, ResultForModel: "the step is on the record"}
	}
}

// handleCancelMission wraps landCancelMission.
func (mt *Guardian) handleCancelMission(d *turnDispatch) toolloop.Handler {
	return func(_ context.Context, call llm.ToolCall) toolloop.Outcome {
		id := argString(call.Args, "id")
		if why := mt.landCancelMission(id, d.grant); why != "" {
			return toolloop.Outcome{Verdict: toolloop.VerdictRejectedGate, ResultForModel: refusal(why)}
		}
		d.result.Mission = &MissionReport{ID: id, Summary: "stood down from the mission " + id}
		return toolloop.Outcome{Verdict: toolloop.VerdictLanded, ResultForModel: "I stand down from it"}
	}
}

// writeMissions renders the active-mission block of the turn user prompt
// (spec 107 FR-002): id, the guardian's goal rendering, remaining game-days,
// and each linked entity's live status — so pursuit and counsel stay
// truthful to recorded state (the writeStandingOrders discipline). Empty
// when none are active, so a mission-free world's prompt is byte-unchanged.
func writeMissions(b *strings.Builder, tick int64, missions []sim.Mission, designations []sim.Designation, directives []sim.Directive) {
	var active []sim.Mission
	for _, m := range missions {
		if m.Status == "active" {
			active = append(active, m)
		}
	}
	if len(active) == 0 {
		return
	}
	statusOf := func(id string, kind string) string {
		switch kind {
		case "dsg":
			for i := range designations {
				if designations[i].ID == id {
					return designations[i].Status
				}
			}
		case "dir":
			for i := range directives {
				if directives[i].ID == id {
					return directives[i].Status
				}
			}
		}
		return "pruned"
	}
	b.WriteString("\nMissions the player has charged you with (standing instructions — pursue them on your own watches at your full skill):\n")
	for i := range active {
		m := &active[i]
		days := (m.DeadlineTick - tick) / ticksPerGameDay
		if days < 0 {
			days = 0
		}
		var linked []string
		for _, id := range m.Designations {
			linked = append(linked, fmt.Sprintf("%s %s", id, statusOf(id, "dsg")))
		}
		for _, id := range m.Directives {
			linked = append(linked, fmt.Sprintf("%s %s", id, statusOf(id, "dir")))
		}
		pursuit := "no pursuit linked yet — decompose it: survey, mark designations (with mission_id), bind directives"
		if len(linked) > 0 {
			pursuit = "linked: " + strings.Join(linked, ", ")
		}
		fmt.Fprintf(b, "- %s: %q (%d day(s) left; %s)\n", m.ID, m.Goal, days, pursuit)
	}
}

// guardianMissionPursuitDirective is the scheduled turn's mission-pursuit
// addendum (spec 107 D4): appended to the cadence seed ONLY when an active
// mission stands, so a mission-free scheduled turn's directive is
// byte-identical to spec 102's. It restates the doctrine the frame carries —
// pursuit is the player's standing instruction, never the guardian's own
// initiative — and points at the atomic mission_id link.
const guardianMissionPursuitDirective = `

A mission the player charged you with still stands (listed above). Pursuing it is carrying out the player's own standing instruction — not initiative — so act for it now at your full skill: survey what you must, mark designations and bind directives that serve the goal (name the mission's id as mission_id so the work is counted), and where the world grants it, work a working. One act at most, as ever: steady progress each watch. If the way is truly blocked, record the obstacle with note_mission_progress.`
