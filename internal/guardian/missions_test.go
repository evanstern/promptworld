package guardian

// Spec 107 suites (FR-006, guardian half): acceptance/refusal at the tool
// door, the atomic mission_id link on the plan verbs, the prompt's mission
// section, and — the doctrine pin — the ceiling composition: a DEFAULT-
// ceiling scheduled turn with an ACTIVE mission regains exactly the pursuit
// surface at full competence (SC-003), while a mission-free one stays the
// modest read/counsel set (TestDefaultCharterCeilingCapsScheduledRoster,
// ceiling_test.go, is the without-mission half of that pin).

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/toolloop"
)

// seedActiveMission lands guardian.mission_accepted on BOTH fixture states
// (replica + injector, the optIn discipline) and refreshes the mirrors, so
// the turn worker sees an active mission exactly as a live absorb would.
func seedActiveMission(t *testing.T, mt *Guardian, inj *stateInjector, id, goal string, tick int64) {
	t.Helper()
	m := sim.Mission{ID: id, Goal: goal, AcceptedTick: tick, DeadlineTick: tick + 7*ticksPerGameDay}
	ev := store.Event{Seq: 1, Tick: tick, Type: "guardian.mission_accepted", Payload: mustJSON(m)}
	if err := mt.replica.Apply(ev); err != nil {
		t.Fatal(err)
	}
	if err := inj.state.Apply(ev); err != nil {
		t.Fatal(err)
	}
	mt.mirrorState()
}

// TestMissionPursuitGrantComposition pins the grant layer itself (spec 107
// D4/T003): under the DEFAULT ceiling, an active mission re-opens exactly
// the pursuit tools the world grants — at full miracle competence — while
// the clock triple, the order door, and the mission accept/cancel verbs stay
// exactly as the ceiling left them; no mission ⇒ identity.
func TestMissionPursuitGrantComposition(t *testing.T) {
	world := fullGrant()
	modest := applyAngelCeiling(world, false)

	// Without a mission: identity — the modest set stands untouched.
	same := applyMissionPursuitGrant(modest, world, false)
	if same.allows("place_designation") || same.allows("work_miracle") {
		t.Fatal("mission-free scheduled grant regained acting tools")
	}

	with := applyMissionPursuitGrant(modest, world, true)
	for _, name := range missionPursuitTools {
		if !with.allows(name) {
			t.Errorf("pursuit tool %q missing from the mission-bearing scheduled grant", name)
		}
	}
	// Full competence: the modest ceiling's empty miracle-kind set is
	// replaced by the WORLD's kind grant (unrestricted here).
	if !with.allowsKind("move") || !with.allowsKind("give_item") {
		t.Error("mission pursuit did not restore full miracle competence")
	}
	// The ceiling still caps everything initiative-shaped and the meta
	// surface: clock, orders, nudges, prophecy, canonization, and the
	// mission accept/cancel verbs (acceptance and stand-down are the
	// player's chat, never the lane's).
	for _, name := range []string{"pause", "start", "adjust_speed", "monitor_and_act", "cancel_order",
		"send_vision", "send_omen", "prophesy", "canonize_region", "accept_mission", "cancel_mission"} {
		if with.allows(name) {
			t.Errorf("mission pursuit leaked %q into the default-ceiling scheduled grant", name)
		}
	}
	// The pursuit layer never widens past the world's own doors: a world
	// that never granted work_miracle stays miracle-less under a mission.
	narrow := fullGrant()
	delete(narrow.tools, "work_miracle")
	got := applyMissionPursuitGrant(applyAngelCeiling(narrow, false), narrow, true)
	if got.allows("work_miracle") {
		t.Error("pursuit grant re-opened a tool the world never granted")
	}
}

// TestScheduledMissionPursuitAtDefaultCeiling is the runTurn-level pin
// (T003): a DEFAULT-charter world's scheduled turn — which without a mission
// declares only the modest read set (ceiling_test.go) — declares the pursuit
// surface once an active mission stands, composes the mission frame and the
// pursuit directive, and lists the mission in the user prompt.
func TestScheduledMissionPursuitAtDefaultCeiling(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "unused")
	seedActiveMission(t, mt, inj, "msn-5-0", "A shelter stands at (10,10).", 5)
	var j toolloop.Job
	captureJob(mt, &j)
	runAngelTurn(t, mt, inj)

	got := map[string]bool{}
	for _, n := range rosterNames(j) {
		got[n] = true
	}
	for _, name := range missionPursuitTools {
		if !got[name] {
			t.Errorf("scheduled roster missing pursuit tool %q under an active mission", name)
		}
	}
	for _, name := range []string{"pause", "start", "adjust_speed", "monitor_and_act", "accept_mission", "cancel_mission", "send_vision"} {
		if got[name] {
			t.Errorf("scheduled roster leaked %q — the ceiling must keep capping initiative", name)
		}
		if _, ok := j.Handlers[name]; ok {
			t.Errorf("capped tool %q still has a handler (door layer leak)", name)
		}
	}
	if !strings.Contains(j.System, "a MISSION the player charged you with still stands") {
		t.Fatal("mission frame missing from the scheduled system prompt")
	}
	if !strings.Contains(j.Seed, "Missions the player has charged you with") ||
		!strings.Contains(j.Seed, "msn-5-0") {
		t.Fatal("mission section missing from the scheduled user prompt")
	}
	if !strings.Contains(j.Seed, "carrying out the player's own standing instruction") {
		t.Fatal("pursuit directive missing from the scheduled seed")
	}
}

// TestConsoleAcceptAndCancelMission drives the accept_mission and
// cancel_mission handlers through a scripted console turn: acceptance lands
// guardian.mission_accepted with the guardian's own goal rendering and
// reports the id; the cancel lands the one-way terminal.
func TestConsoleAcceptAndCancelMission(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "as you will")
	grant := fullGrant()

	d := &turnDispatch{mt: mt, tick: 100, result: &TurnResult{}, grant: grant}
	out := mt.handleAcceptMission(d)(t.Context(), toolCall("accept_mission",
		`{"goal":"A second fire structure stands near (12,8) and stays fueled.","ttl_days":5}`))
	if out.Verdict != toolloop.VerdictLanded {
		t.Fatalf("accept verdict = %s (%s)", out.Verdict, out.ResultForModel)
	}
	if d.result.Mission == nil || d.result.Mission.ID == "" {
		t.Fatal("no MissionReport on the shared result")
	}
	id := d.result.Mission.ID
	m := inj.state.MissionByID(id)
	if m == nil || m.Status != "active" {
		t.Fatalf("mission %q not active on state", id)
	}
	if m.DeadlineTick != 100+5*ticksPerGameDay {
		t.Fatalf("DeadlineTick = %d, want the 5-day TTL", m.DeadlineTick)
	}

	// Door-refusal counsel: an empty goal never reaches the door; a bad TTL
	// names the honest bounds (repairable, the rejected_gate contract).
	if out := mt.handleAcceptMission(d)(t.Context(), toolCall("accept_mission", `{"goal":"  "}`)); out.Verdict != toolloop.VerdictRejectedGate {
		t.Fatalf("empty goal verdict = %s", out.Verdict)
	}
	if out := mt.handleAcceptMission(d)(t.Context(), toolCall("accept_mission", `{"goal":"x","ttl_days":99}`)); out.Verdict != toolloop.VerdictRejectedGate ||
		!strings.Contains(out.ResultForModel, "days") {
		t.Fatalf("bad ttl = %s (%s)", out.Verdict, out.ResultForModel)
	}

	// The player's stand-down: cancel lands the one-way terminal.
	if out := mt.handleCancelMission(d)(t.Context(), toolCall("cancel_mission", `{"id":"`+id+`"}`)); out.Verdict != toolloop.VerdictLanded {
		t.Fatalf("cancel verdict = %s (%s)", out.Verdict, out.ResultForModel)
	}
	if got := inj.state.MissionByID(id).Status; got != "cancelled" {
		t.Fatalf("status = %q, want cancelled", got)
	}
	// A second cancel refuses with counsel.
	if out := mt.handleCancelMission(d)(t.Context(), toolCall("cancel_mission", `{"id":"`+id+`"}`)); out.Verdict != toolloop.VerdictRejectedGate {
		t.Fatalf("re-cancel verdict = %s", out.Verdict)
	}
}

// TestMissionLinkRidesPlanBatchAtomically: place_designation with mission_id
// lands designation.placed + guardian.mission_progressed as ONE batch (the
// link's existence check reads the placement in batch order), and a bad
// mission_id refuses the WHOLE act — no half-linked placement ever lands.
func TestMissionLinkRidesPlanBatchAtomically(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "so marked")
	seedActiveMission(t, mt, inj, "msn-5-0", "A shelter stands at (10,10).", 5)
	grant := fullGrant()

	dsg, why := mt.landPlaceDesignation("structure_site", "10,10", "shelter", 0, false, "", "msn-5-0", 100, grant)
	if why != "" {
		t.Fatalf("linked placement refused: %s", why)
	}
	last := inj.batches[len(inj.batches)-1]
	if len(last) != 2 || last[0].Type != "designation.placed" || last[1].Type != "guardian.mission_progressed" {
		t.Fatalf("batch types = %v, want [designation.placed guardian.mission_progressed]", eventTypes(last))
	}
	m := inj.state.MissionByID("msn-5-0")
	if len(m.Designations) != 1 || m.Designations[0] != dsg.ID {
		t.Fatalf("mission links = %v, want [%s]", m.Designations, dsg.ID)
	}

	// A bad mission_id refuses the whole batch: nothing landed, counsel names
	// the repair.
	before := len(inj.batches)
	_, why = mt.landPlaceDesignation("structure_site", "11,10", "shelter", 0, false, "", "msn-nope", 101, grant)
	if why == "" || !strings.Contains(why, "no mission") {
		t.Fatalf("bad mission link refusal = %q", why)
	}
	if len(inj.batches) != before {
		t.Fatal("a refused linked placement still landed a batch")
	}

	// The directive twin: issue_directive with mission_id links atomically.
	alive := map[int]bool{}
	for i := range sim.AgentNames {
		alive[i] = true
	}
	dir, why := mt.landIssueDirective(dsg.ID, sim.AgentNames[0], "Raise the shelter I have marked.", 0, "msn-5-0", 102, alive, grant)
	if why != "" {
		t.Fatalf("linked directive refused: %s", why)
	}
	m = inj.state.MissionByID("msn-5-0")
	if len(m.Directives) != 1 || m.Directives[0] != dir.ID {
		t.Fatalf("mission directive links = %v, want [%s]", m.Directives, dir.ID)
	}
}

// TestNoteMissionProgressHandler: the explicit progress verb links existing
// work and records obstacles; empty progress is counsel, not a door trip.
func TestNoteMissionProgressHandler(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "noted")
	seedActiveMission(t, mt, inj, "msn-5-0", "A shelter stands at (10,10).", 5)
	d := &turnDispatch{mt: mt, tick: 100, result: &TurnResult{}, grant: fullGrant()}

	if out := mt.handleNoteMissionProgress(d)(t.Context(), toolCall("note_mission_progress",
		`{"id":"msn-5-0","note":"the west bank floods; I will mark the site on higher ground"}`)); out.Verdict != toolloop.VerdictLanded {
		t.Fatalf("note verdict = %s (%s)", out.Verdict, out.ResultForModel)
	}
	if out := mt.handleNoteMissionProgress(d)(t.Context(), toolCall("note_mission_progress",
		`{"id":"msn-5-0"}`)); out.Verdict != toolloop.VerdictRejectedGate {
		t.Fatal("empty progress must counsel")
	}
	if out := mt.handleNoteMissionProgress(d)(t.Context(), toolCall("note_mission_progress",
		`{"id":"msn-5-0","designation_id":"dsg-nope"}`)); out.Verdict != toolloop.VerdictRejectedGate ||
		!strings.Contains(out.ResultForModel, "designation") {
		t.Fatal("unknown designation link must counsel")
	}
}

// TestMissionPromptSectionByteInert: a mission-free prompt is byte-identical
// with and without the missions parameter wired (FR-002's prompt-inert
// guarantee), and an active mission renders id, goal, days, and link status.
func TestMissionPromptSectionByteInert(t *testing.T) {
	alive := map[int]bool{0: true}
	base := turnUserPrompt(100, 1, sim.FaithGenesis, alive, nil, nil, nil, nil, nil, nil, nil, nil, "", "", "", "The player says:\nhello")
	if strings.Contains(base, "Missions") {
		t.Fatal("mission-free prompt mentions missions")
	}
	missions := []sim.Mission{{ID: "msn-5-0", Goal: "A shelter stands at (10,10).",
		AcceptedTick: 5, DeadlineTick: 5 + 7*ticksPerGameDay, Status: "active",
		Designations: []string{"dsg-9-0"}}}
	designations := []sim.Designation{{ID: "dsg-9-0", Kind: sim.DesignationStructureSite,
		X: 10, Y: 10, StructureKind: "shelter", Status: "fulfilled"}}
	got := turnUserPrompt(100, 1, sim.FaithGenesis, alive, nil, designations, nil, nil, missions, nil, nil, nil, "", "", "", "The player says:\nhello")
	for _, want := range []string{"Missions the player has charged you with", "msn-5-0",
		"A shelter stands at (10,10).", "dsg-9-0 fulfilled"} {
		if !strings.Contains(got, want) {
			t.Errorf("mission section missing %q:\n%s", want, got)
		}
	}
}

func eventTypes(evs []store.Event) []string {
	out := make([]string, len(evs))
	for i, e := range evs {
		out[i] = e.Type
	}
	return out
}
