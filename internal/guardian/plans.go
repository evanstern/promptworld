package guardian

// The plan layer's tool handlers (spec 084 US2/US3): place_designation /
// cancel_designation / issue_directive / cancel_directive. Doctrine is the
// standing-order file's (orders.go): the reducer dry-run is the door authority
// for every lifecycle rule — form, bounds, occupancy, caps, TTLs, living
// targets, the one-way races — and these helpers only parse the call surface
// (loci through target.ParseLocus, the one-parser law), mint deterministic
// ids, land through InjectSocial, and map a door rejection to in-fiction
// counsel the loop feeds back as a rejected_gate.

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/target"
	"github.com/evanstern/promptworld/internal/toolloop"
)

// PlanReport is the console-facing summary of a landed plan act (spec 084):
// the entity id the player can name to cancel it, and a one-line human
// rendering. Additive omitempty on TurnResult.
type PlanReport struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
}

// nextPlanID assigns "<prefix>-<tick>-<seq>" (the nextOrderID shape, research
// R1): human-readable, deterministic, no RNG draw. seq disambiguates same-tick
// mints — the max seq already present at this tick in the mirror plus one,
// floored by the serialized per-prefix counter so a placement whose
// predecessor has not yet flowed back through Observe still gets a fresh id.
// Uniqueness is ultimately reducer-enforced (duplicate ids are refused).
func (mt *Guardian) nextPlanID(prefix string, tick int64) string {
	mt.stateMu.Lock()
	defer mt.stateMu.Unlock()
	seq := 0
	pre := fmt.Sprintf("%s-%d-", prefix, tick)
	bump := func(id string) {
		if strings.HasPrefix(id, pre) {
			if s, err := strconv.Atoi(strings.TrimPrefix(id, pre)); err == nil && s >= seq {
				seq = s + 1
			}
		}
	}
	switch prefix {
	case "dsg":
		for i := range mt.designations {
			bump(mt.designations[i].ID)
		}
	case "dir":
		for i := range mt.directives {
			bump(mt.directives[i].ID)
		}
	}
	if mt.planMintTick[prefix] == tick && mt.planMintSeq[prefix] >= seq {
		seq = mt.planMintSeq[prefix] + 1
	}
	mt.planMintTick[prefix] = tick
	mt.planMintSeq[prefix] = seq
	return fmt.Sprintf("%s-%d-%d", prefix, tick, seq)
}

// designationForms maps a designation kind to the one locus form it admits
// (spec 084 FR-003, data-model §1) — any other (kind × form) cell is a form
// error at this door, before anything reaches the reducer.
var designationForms = map[string]target.Form{
	sim.DesignationStructureSite:  target.FormPoint,
	sim.DesignationWallLine:       target.FormLine,
	sim.DesignationSettlementZone: target.FormRect,
}

// designationFormHint phrases the expected locus form for a kind, for counsel.
var designationFormHint = map[string]string{
	sim.DesignationStructureSite:  `one tile, like "4,5"`,
	sim.DesignationWallLine:       `an axis-aligned line, like "2,2->2,9"`,
	sim.DesignationSettlementZone: `a rectangle, like "1,1..8,8"`,
}

// landPlaceDesignation parses a place_designation call, mints the id, and
// lands designation.placed through the door (spec 084 US2). Loci parse through
// target.ParseLocus — the SAME grammar/normalization/enumeration every other
// consumer uses (FR-003) — and land NORMALIZED ints in the payload (replay
// never re-parses). minStructures carries the tool-door default (3) when the
// caller omitted it; structure_site requires structure_kind HERE (the
// parseReveal partial-args shape). Returns the placed designation or
// (nil, in-fiction refusal).
func (mt *Guardian) landPlaceDesignation(kind, targetArg, structureKind string, minStructures int, hasMin bool, label string, tick int64, grant grantSet) (*sim.Designation, string) {
	if !grant.allows("place_designation") {
		return nil, "that power is not granted in this world"
	}
	kind = strings.TrimSpace(kind)
	wantForm, ok := designationForms[kind]
	if !ok {
		return nil, fmt.Sprintf("I know no designation of kind %q — settlement_zone, structure_site, or wall_line", kind)
	}
	addr, err := target.ParseLocus(targetArg)
	if err != nil {
		return nil, fmt.Sprintf("I could not read that site (%v)", err)
	}
	if addr.Form != wantForm {
		return nil, fmt.Sprintf("a %s wants %s", kind, designationFormHint[kind])
	}
	structureKind = strings.TrimSpace(structureKind)
	if kind == sim.DesignationStructureSite && structureKind == "" {
		return nil, "a structure site needs a structure_kind — what should stand there?"
	}
	d := sim.Designation{
		ID:            mt.nextPlanID("dsg", tick),
		Kind:          kind,
		X:             addr.X,
		Y:             addr.Y,
		X2:            addr.X2,
		Y2:            addr.Y2,
		StructureKind: structureKind,
		Label:         strings.TrimSpace(label),
		PlacedTick:    tick,
	}
	if addr.Form == target.FormPoint {
		d.X2, d.Y2 = addr.X, addr.Y
	}
	if kind == sim.DesignationSettlementZone {
		d.MinStructures = minStructures
		if !hasMin {
			d.MinStructures = 3 // the tool-door default (data-model §1)
		}
	}
	batch := []store.Event{{Type: "designation.placed", Payload: mustJSON(d)}}
	if err := mt.social.InjectSocial(batch); err != nil {
		log.Printf("guardian: designation rejected at the door: %v", err)
		return nil, planRefusal(err)
	}
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — I marked a %s (%s) at %s\n",
		clock.Format(mt.replicaTickSafe()), kind, d.ID, describeDesignationSite(&d)))
	return &d, ""
}

// landCancelDesignation lands designation.cancelled for the named id through
// the door. The reducer resolves the cancel/fulfil race — an unknown or
// non-active id refuses with counsel (the cancelOrder shape).
func (mt *Guardian) landCancelDesignation(id string, grant grantSet) string {
	if !grant.allows("cancel_designation") {
		return "that power is not granted in this world"
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return "name the designation you want me to withdraw"
	}
	batch := []store.Event{{Type: "designation.cancelled", Payload: mustJSON(sim.OrderIDPayload{ID: id})}}
	if err := mt.social.InjectSocial(batch); err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "unknown designation"):
			return fmt.Sprintf("I keep no designation called %q", id)
		case strings.Contains(msg, "not active"):
			return fmt.Sprintf("the designation %q has already run its course", id)
		default:
			return "the world would not let me withdraw it (" + msg + ")"
		}
	}
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — I withdrew a designation (%s)\n",
		clock.Format(mt.replicaTickSafe()), id))
	return ""
}

// landIssueDirective resolves the targets (comma-names or "everyone" — the
// send_omen vocabulary, all-living-or-reject), mints the id, and lands ONE
// atomic batch: directive.issued plus one companion agent.memory_added per
// target (spec 084 FR-009 — the vision-memory shape, so the prompt firewall
// holds exactly as for visions: guardian prose enters the sim only as recorded
// event data). The whole batch lands or nothing does. Returns the issued
// directive or (nil, in-fiction refusal).
func (mt *Guardian) landIssueDirective(designationID, targetsArg, text string, ttlDays int, tick int64, alive map[int]bool, grant grantSet) (*sim.Directive, string) {
	if !grant.allows("issue_directive") {
		return nil, "that power is not granted in this world"
	}
	designationID = strings.TrimSpace(designationID)
	if designationID == "" {
		return nil, "name the designation the directive should serve"
	}
	targets, why := resolveOmenTargets(targetsArg, alive)
	if why != "" {
		return nil, why
	}
	// The reducer requires ascending-unique indices (payloads are
	// self-contained); resolveOmenTargets preserves naming order, so sort.
	sort.Ints(targets)
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, "give me the charge in words — what should they do?"
	}
	if ttlDays == 0 {
		ttlDays = 3 // default (spec Assumption, the monitor_and_act shape)
	}
	if ttlDays < sim.GuardianOrderTTLMinDays || ttlDays > sim.GuardianOrderTTLMaxDays {
		return nil, fmt.Sprintf("a directive may stand for %d to %d days", sim.GuardianOrderTTLMinDays, sim.GuardianOrderTTLMaxDays)
	}
	d := sim.Directive{
		ID:            mt.nextPlanID("dir", tick),
		DesignationID: designationID,
		Targets:       targets,
		Village:       strings.EqualFold(strings.TrimSpace(targetsArg), "everyone"),
		Text:          text,
		IssuedTick:    tick,
		ExpiresTick:   tick + int64(ttlDays)*ticksPerGameDay,
	}
	// FROZEN recorded-at-emission prefix (spec 052 ruling 1): the companion
	// memory lands in agent.memory_added payloads — the event log is
	// skin-free, so the wording is fixed mechanics vocabulary.
	batch := []store.Event{{Type: "directive.issued", Payload: mustJSON(d)}}
	for _, t := range targets {
		batch = append(batch, store.Event{Type: "agent.memory_added", Payload: mustJSON(sim.MemoryAddedPayload{
			Agent: t, Text: "The Guardian charges you: " + text, Salience: sim.SalDream,
			Subject: -1, Origin: sim.OriginOmen})})
	}
	if err := mt.social.InjectSocial(batch); err != nil {
		log.Printf("guardian: directive rejected at the door: %v", err)
		return nil, planRefusal(err)
	}
	names := make([]string, len(targets))
	for i, t := range targets {
		names[i] = sim.AgentNames[t]
	}
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — I charged %s (%s, serving %s): %q\n",
		clock.Format(mt.replicaTickSafe()), strings.Join(names, ", "), d.ID, designationID, text))
	return &d, ""
}

// landCancelDirective lands directive.cancelled for the named id.
func (mt *Guardian) landCancelDirective(id string, grant grantSet) string {
	if !grant.allows("cancel_directive") {
		return "that power is not granted in this world"
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return "name the directive you want me to withdraw"
	}
	batch := []store.Event{{Type: "directive.cancelled", Payload: mustJSON(sim.OrderIDPayload{ID: id})}}
	if err := mt.social.InjectSocial(batch); err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "unknown directive"):
			return fmt.Sprintf("I keep no directive called %q", id)
		case strings.Contains(msg, "not active"):
			return fmt.Sprintf("the directive %q has already run its course", id)
		default:
			return "the world would not let me withdraw it (" + msg + ")"
		}
	}
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — I withdrew a directive (%s)\n",
		clock.Format(mt.replicaTickSafe()), id))
	return ""
}

// planRefusal maps a designation/directive door rejection to in-fiction
// counsel (the orderRefusal shape): the reducer's error strings are the
// source, translated so the model hears a repairable reason.
func planRefusal(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "cap 16"):
		return "I already hold as many designations as the world can bear — withdraw one and I will mark another"
	case strings.Contains(msg, "cap 3"):
		return "I already bind the village with as many directives as I may — withdraw one and I will issue another"
	case strings.Contains(msg, "outside the world"):
		return "that site lies beyond the world's edge"
	case strings.Contains(msg, "already stands"):
		return "something else already stands on that tile — it can never hold what you ask"
	case strings.Contains(msg, "stands on the line"):
		return "something that is not a wall stands on that line — it can never be walled whole"
	case strings.Contains(msg, "unknown designation"):
		return "I keep no designation by that name"
	case strings.Contains(msg, "is not active"):
		return "that designation has already run its course"
	case strings.Contains(msg, "is dead"):
		return "one of those villagers is beyond reach now"
	case strings.Contains(msg, "label over"):
		return "give it a shorter name and I will mark it"
	case strings.Contains(msg, "exceeds the"):
		return "that claim is too large — mark a smaller extent"
	case strings.Contains(msg, "ttl"):
		return "a directive may stand only for a handful of days"
	default:
		return "the world would not let me (" + msg + ")"
	}
}

// describeDesignationSite renders a designation's site in plain words — the
// soul-append and prompt-section vocabulary, one home.
func describeDesignationSite(d *sim.Designation) string {
	switch d.Kind {
	case sim.DesignationWallLine:
		return fmt.Sprintf("(%d,%d)->(%d,%d)", d.X, d.Y, d.X2, d.Y2)
	case sim.DesignationSettlementZone:
		return fmt.Sprintf("(%d,%d)..(%d,%d)", d.X, d.Y, d.X2, d.Y2)
	default:
		return fmt.Sprintf("(%d,%d)", d.X, d.Y)
	}
}

// --- tool-use loop handlers (the toolcalls.go dispatch shape) ---

// handlePlaceDesignation wraps landPlaceDesignation: door accept → landed
// (PlanReport on the shared result); refusal → rejected_gate the model may
// repair within the round cap.
func (mt *Guardian) handlePlaceDesignation(d *turnDispatch) toolloop.Handler {
	return func(_ context.Context, call llm.ToolCall) toolloop.Outcome {
		minStructures, hasMin := argInt(call.Args, "min_structures")
		dsg, why := mt.landPlaceDesignation(
			argString(call.Args, "kind"), argString(call.Args, "target"),
			argString(call.Args, "structure_kind"), minStructures, hasMin,
			argString(call.Args, "label"), d.tick, d.grant)
		if dsg == nil {
			return toolloop.Outcome{Verdict: toolloop.VerdictRejectedGate, ResultForModel: refusal(why)}
		}
		d.result.Plan = &PlanReport{ID: dsg.ID,
			Summary: fmt.Sprintf("marked a %s at %s", dsg.Kind, describeDesignationSite(dsg))}
		return toolloop.Outcome{Verdict: toolloop.VerdictLanded,
			ResultForModel: "the mark is set (" + dsg.ID + ") — every villager now knows of it"}
	}
}

// handleCancelDesignation wraps landCancelDesignation.
func (mt *Guardian) handleCancelDesignation(d *turnDispatch) toolloop.Handler {
	return func(_ context.Context, call llm.ToolCall) toolloop.Outcome {
		id := argString(call.Args, "id")
		if why := mt.landCancelDesignation(id, d.grant); why != "" {
			return toolloop.Outcome{Verdict: toolloop.VerdictRejectedGate, ResultForModel: refusal(why)}
		}
		d.result.Plan = &PlanReport{ID: id, Summary: "withdrew the designation " + id}
		return toolloop.Outcome{Verdict: toolloop.VerdictLanded, ResultForModel: "the mark is withdrawn"}
	}
}

// handleIssueDirective wraps landIssueDirective.
func (mt *Guardian) handleIssueDirective(d *turnDispatch) toolloop.Handler {
	return func(_ context.Context, call llm.ToolCall) toolloop.Outcome {
		ttl, _ := argInt(call.Args, "ttl_days")
		dir, why := mt.landIssueDirective(
			argString(call.Args, "designation_id"), argString(call.Args, "targets"),
			argString(call.Args, "text"), ttl, d.tick, d.alive, d.grant)
		if dir == nil {
			return toolloop.Outcome{Verdict: toolloop.VerdictRejectedGate, ResultForModel: refusal(why)}
		}
		d.result.Plan = &PlanReport{ID: dir.ID,
			Summary: fmt.Sprintf("bound %d villager(s) to %s", len(dir.Targets), dir.DesignationID)}
		return toolloop.Outcome{Verdict: toolloop.VerdictLanded,
			ResultForModel: "the charge is laid on them (" + dir.ID + ")"}
	}
}

// handleCancelDirective wraps landCancelDirective.
func (mt *Guardian) handleCancelDirective(d *turnDispatch) toolloop.Handler {
	return func(_ context.Context, call llm.ToolCall) toolloop.Outcome {
		id := argString(call.Args, "id")
		if why := mt.landCancelDirective(id, d.grant); why != "" {
			return toolloop.Outcome{Verdict: toolloop.VerdictRejectedGate, ResultForModel: refusal(why)}
		}
		d.result.Plan = &PlanReport{ID: id, Summary: "withdrew the directive " + id}
		return toolloop.Outcome{Verdict: toolloop.VerdictLanded, ResultForModel: "the charge is lifted"}
	}
}
