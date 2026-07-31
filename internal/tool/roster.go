package tool

// Rosters express capability as membership (spec 014 US3, research R4): each
// agent kind has an ordered list of the registry tool names it may use. An
// action naming a tool outside the acting agent's roster is rejected at the
// door exactly like an unknown action. Rosters are data, not code branches.
//
// The villager roster's world-verb portion is DERIVED from the registry (every
// World tool, in registration order) so that adding a world verb is a single
// registry edit — it joins the roster and every derived surface at once
// (SC-001). The expressive membership is explicit, because say/muse/gist and
// converse/nudge_* are all Effect Expressive and only the roster distinguishes
// which agent kind holds which.

// villagerExpressive are the expressive tools a villager may use, in roster
// order (say, muse, gist).
var villagerExpressive = []string{"say", "muse", "gist"}

// dormantVillagerVerbs (spec 058 US3, TASK-110): collect_water and bathe are
// PRUNED from the villager-facing surfaces below — the tool-use loop's
// declared roster (LoopRosterVillager) and set_plan's step-goal enum
// (setPlanTool, registry.go) — because they are a non-choice today: water has
// no consumer beyond bathe, collection fell to 0/day once novelty wore off,
// and bathe fired once in 6 days of world-01. Every turn a model spends
// considering a verb it can't meaningfully use is pure loss (task diagnosis).
//
// Everything ELSE about both verbs is UNCHANGED and deliberately so: the
// registry Tool entries (worldToolsBase), the resolver/executor machinery
// (internal/sim policy.go/executor.go/recipes.go), RosterVillager (the sim
// door's OWN membership set), and PlanStepGoals (the plan-step accept set)
// all still know both verbs — so a historical world's collect_water/bathe
// events replay exactly, and bringing either back to the villager surface is
// a roster/gloss edit, not a rebuild.
//
// REVISIT CONDITION: reintroduce collect_water (and, by extension, bathe) the
// day a thirst need — or any other design reason water should matter to a
// villager — gives water a real consumer again.
var dormantVillagerVerbs = map[string]bool{"collect_water": true, "bathe": true}

// pruneDormant returns names with every dormantVillagerVerbs entry removed,
// order preserved — the shared filter both villager-facing surfaces
// (LoopRosterVillager here, setPlanTool's goal enum in registry.go) apply.
func pruneDormant(names []string) []string {
	out := make([]string, 0, len(names))
	for _, n := range names {
		if !dormantVillagerVerbs[n] {
			out = append(out, n)
		}
	}
	return out
}

// RosterVillager is the villager capability set: every legacy World tool
// (Effect World AND PlanStep true, isLegacyWorldTool in derive.go) in
// registration order, then the villager expressive tools. set_plan is
// deliberately excluded — it is Effect World but loop-only (PlanStep false);
// it appears only in LoopRosterVillager below.
var RosterVillager = func() []string {
	out := make([]string, 0, len(registry))
	for _, t := range registry {
		if isLegacyWorldTool(t) {
			out = append(out, t.Name)
		}
	}
	return append(out, villagerExpressive...)
}()

// RosterGuardian is the guardian capability set (the door name set): its
// converse channel and the acting tools it may use (spec 029 — send_vision/
// send_omen replace the retired nudge_dream/nudge_omen; monitor_and_act,
// cancel_order, work_miracle, and the meta tools pause/start/adjust_speed join
// it). It mirrors LoopRosterGuardian's names plus converse. Since spec 029 the
// nudge form is validated against the reducer's explicit form set, not this
// roster (contracts/events.md), so this set's live consumer is the boot-time
// name-resolution check in Validate; keeping it in step keeps that gate honest.
// The plan layer (spec 084) appends after explain: the four charge-free plan
// verbs and the survey read tool, granted at every stage (the monitor_and_act
// precedent — spec 084 Assumptions). prophesy (spec 085) appends last: the
// charge-priced declared claim, following send_vision's stage profile.
// canonize_region / brief_myths (spec 101) append last, so no existing
// tool's registration position shifts: the canonization miracle (premium
// charge-priced) and its read-only myth-briefing companion.
// The mission layer (spec 107) appends last again: accept_mission /
// note_mission_progress / cancel_mission — charge-free artifact verbs (the
// plan-layer posture; a mission records intent, the priced verbs still do
// all the acting).
var RosterGuardian = []string{"converse", "send_omen", "send_vision", "monitor_and_act", "cancel_order", "work_miracle", "pause", "start", "adjust_speed", "explain",
	"place_designation", "cancel_designation", "issue_directive", "cancel_directive", "survey_site", "prophesy", "canonize_region", "brief_myths",
	"accept_mission", "note_mission_progress", "cancel_mission"}

// OnRoster reports whether name is on roster — the door membership check.
func OnRoster(roster []string, name string) bool {
	for _, n := range roster {
		if n == name {
			return true
		}
	}
	return false
}

// LoopRosterVillager returns the ordered declared-tool list the villager
// tool-use loop presents to the model (spec 017 contracts/loop-api.md,
// data-model.md §2): every legacy World tool in registration order — MINUS
// dormantVillagerVerbs (spec 058 US3: collect_water, bathe) — then set_plan,
// then muse. Unlike RosterVillager (name-only membership, for the door's
// roster check, unaffected by the prune), this returns full Tool values —
// InputSchema (derive.go) needs each tool's Params/InputSchemaJSON, not just
// its name.
//
// say/gist stay scene-gated and out of the loop roster this task (data-model
// §2): scenes remain driver-run, not model-initiated via the loop.
func LoopRosterVillager() []Tool {
	out := make([]Tool, 0, len(registry))
	for _, t := range registry {
		if isLegacyWorldTool(t) && !dormantVillagerVerbs[t.Name] {
			out = append(out, t)
		}
	}
	if sp, ok := Lookup("set_plan"); ok {
		out = append(out, sp)
	}
	if muse, ok := Lookup("muse"); ok {
		out = append(out, muse)
	}
	// Journal tools (spec 019, US3): the villager's private notebook — two
	// acting (Expressive) and two Read. Appended after muse so no existing
	// declared tool's position shifts; villager-only (the guardian roster is
	// untouched, journals are private).
	for _, n := range []string{"write_journal_entry", "delete_from_journal", "search_journal", "read_journal"} {
		if t, ok := Lookup(n); ok {
			out = append(out, t)
		}
	}
	return out
}

// loopGuardianTools is the ordered declared-tool list the guardian tool-use
// loop presents to the model (spec 017 T020; the agency surface, spec 029 R2):
// send_omen, send_vision, monitor_and_act, cancel_order, work_miracle, then the
// meta tools pause/start/adjust_speed. It is NOT RosterGuardian: converse is
// DELIBERATELY excluded. converse is the final-answer channel, not a callable
// tool — the guardian speaks by replying with text (toolloop Result Final), and the
// loop ends naturally (model_done) when it does. Declaring converse would trap a
// converse call as rejected_unknown (guardian installs no converse handler, by
// design: "converse is the transcript, not a door"), so it is offered only as
// the implicit text channel, never as a tool the model can call.
// explain (spec 063) is appended last: the read-only facts tool joins the
// declared loop surface like any other guardian tool — grant-gated through
// the same three layers — without shifting any existing tool's position.
// The plan layer (spec 084) appends after explain in registration-order
// discipline: place/cancel_designation, issue/cancel_directive (acting), and
// survey_site (Read — the explain dispatch class). prophesy (spec 085)
// appends last, so no existing tool's declared position shifts.
// The mission layer (spec 107) appends last again, the same discipline:
// three charge-free artifact verbs, no existing declared position shifts.
var loopGuardianTools = []string{"send_omen", "send_vision", "monitor_and_act", "cancel_order", "work_miracle", "pause", "start", "adjust_speed", "explain",
	"place_designation", "cancel_designation", "issue_directive", "cancel_directive", "survey_site", "prophesy", "canonize_region", "brief_myths",
	"accept_mission", "note_mission_progress", "cancel_mission"}

// LoopRosterGuardian returns the ordered declared-tool list the guardian
// tool-use loop presents to the model (loopGuardianTools), resolved to full
// Tool values — InputSchema (derive.go) needs each tool's Params, not just its
// name. RosterGuardian stays the pre-loop, name-only DOOR roster (landNudge's
// OnRoster check); this is the loop's DECLARED surface, which differs (converse
// excluded, work_miracle included).
func LoopRosterGuardian() []Tool {
	out := make([]Tool, 0, len(loopGuardianTools))
	for _, n := range loopGuardianTools {
		if t, ok := Lookup(n); ok {
			out = append(out, t)
		}
	}
	return out
}
