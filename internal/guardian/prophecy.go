package guardian

// The prophesy tool handler (spec 085 US3): the guardian stakes its
// credibility. Doctrine is the plan layer's (plans.go): the reducer dry-run
// is the door authority for every lifecycle rule — targets, text, TTL, the
// cap, the closed claim vocabulary, the already-true and active-duplicate
// refusals — and this helper only parses the call surface (kind-conditional
// claim params, partial or foreign args refused — the parseReveal shape),
// mints the deterministic id, lands ONE atomic batch through InjectSocial
// (prophecy.declared + one companion OriginOmen memory per target — the
// vision-memory shape, so the prompt firewall holds exactly as for visions),
// and maps a door rejection to in-fiction counsel. There is NO cancel
// handler: the word, once given, stands (research R8).

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/toolloop"
)

// prophesyArgs is the parsed prophesy call surface, claim params still raw
// (presence-tracked so foreign args for the chosen kind refuse cleanly).
type prophesyArgs struct {
	targets       string
	text          string
	claimKind     string
	designationID string
	structureKind string
	min           int
	hasMin        bool
	agentName     string
	deadlineDays  int
}

// assembleClaim builds the normalized sim.ProphecyClaim from the
// kind-conditional params, refusing partial or foreign argument sets before
// anything reaches the reducer (the parseReveal partial-triple shape). The
// reducer's claim door re-validates everything semantically.
func assembleClaim(a *prophesyArgs) (sim.ProphecyClaim, string) {
	c := sim.ProphecyClaim{Kind: strings.TrimSpace(a.claimKind)}
	foreign := func(fields string) string {
		return fmt.Sprintf("a %s claim takes %s and nothing else", c.Kind, fields)
	}
	switch c.Kind {
	case sim.ProphecyDesignationFulfilled:
		if a.structureKind != "" || a.hasMin || a.agentName != "" {
			return c, foreign("designation_id")
		}
		c.DesignationID = strings.TrimSpace(a.designationID)
		if c.DesignationID == "" {
			return c, "name the designation the claim stakes on (designation_id)"
		}
	case sim.ProphecyStructureCount:
		if a.designationID != "" || a.agentName != "" {
			return c, foreign("structure_kind and min")
		}
		c.StructureKind = strings.TrimSpace(a.structureKind)
		if c.StructureKind == "" {
			return c, "name the structure kind the claim counts (structure_kind)"
		}
		if !a.hasMin {
			return c, "give the count the claim must reach (min)"
		}
		c.Min = a.min
	case sim.ProphecyPopulationAtLeast:
		if a.designationID != "" || a.structureKind != "" || a.agentName != "" {
			return c, foreign("min")
		}
		if !a.hasMin {
			return c, "give the living count the claim must reach (min)"
		}
		c.Min = a.min
	case sim.ProphecySurvives:
		if a.designationID != "" || a.structureKind != "" || a.hasMin {
			return c, foreign("agent")
		}
		name := strings.TrimSpace(a.agentName)
		if name == "" {
			return c, "name the villager who must live to the deadline (agent)"
		}
		idx := agentIndexByName(name)
		if idx < 0 {
			return c, fmt.Sprintf("no villager named %q", name)
		}
		c.Agent = idx
	default:
		return c, fmt.Sprintf("I know no claim of kind %q — designation_fulfilled, structure_count, population_at_least, or survives", c.Kind)
	}
	return c, ""
}

// landProphesy resolves the targets (the send_omen vocabulary), assembles the
// claim, mints the id, and lands the ONE atomic batch: prophecy.declared plus
// one companion agent.memory_added per target (OriginOmen, dream band — the
// directive-companion shape). The charge spend rides the declaration itself:
// the prophecy.declared reducer arm validates and decrements the bank (the
// guardian.nudged contract), so the turn-side charges check here is only the
// polite pre-check — the door stays the authority. Returns the declared
// prophecy or (nil, in-fiction refusal).
func (mt *Guardian) landProphesy(a *prophesyArgs, charges int, tick int64, alive map[int]bool, grant grantSet) (*sim.Prophecy, string) {
	if !grant.allows("prophesy") {
		return nil, "that power is not granted in this world"
	}
	if charges <= 0 {
		return nil, "no charges are banked — a prophecy needs a stake"
	}
	targets, why := resolveOmenTargets(a.targets, alive)
	if why != "" {
		return nil, why
	}
	sort.Ints(targets) // the reducer requires ascending-unique indices
	text := strings.TrimSpace(a.text)
	if text == "" {
		return nil, "give me the word to declare — what will come to pass?"
	}
	claim, why := assembleClaim(a)
	if why != "" {
		return nil, why
	}
	days := a.deadlineDays
	if days == 0 {
		days = 3 // the tool-door default (data-model §4, the ttl_days shape)
	}
	if days < sim.GuardianOrderTTLMinDays || days > sim.GuardianOrderTTLMaxDays {
		return nil, fmt.Sprintf("a prophecy may stand for %d to %d days", sim.GuardianOrderTTLMinDays, sim.GuardianOrderTTLMaxDays)
	}
	p := sim.Prophecy{
		ID:           mt.nextPlanID("pro", tick),
		Targets:      targets,
		Village:      strings.EqualFold(strings.TrimSpace(a.targets), "everyone"),
		Text:         text,
		Claim:        claim,
		DeclaredTick: tick,
		DeadlineTick: tick + int64(days)*ticksPerGameDay,
	}
	// FROZEN recorded-at-emission prefix (spec 052 ruling 1): the companion
	// memory lands in agent.memory_added payloads — the event log is
	// skin-free, so the wording is fixed mechanics vocabulary (data-model §8).
	batch := []store.Event{{Type: "prophecy.declared", Payload: mustJSON(p.DeclaredPayload())}}
	for _, t := range targets {
		batch = append(batch, store.Event{Type: "agent.memory_added", Payload: mustJSON(sim.MemoryAddedPayload{
			Agent: sim.Ref(t), Text: "The Guardian foretells: " + text, Salience: sim.SalDream,
			Subject: sim.Ref(-1), Origin: sim.OriginOmen})})
	}
	if err := mt.social.InjectSocial(batch); err != nil {
		log.Printf("guardian: prophecy rejected at the door: %v", err)
		return nil, prophecyRefusal(err)
	}
	names := make([]string, len(targets))
	for i, t := range targets {
		names[i] = sim.AgentNames[t]
	}
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — I prophesied to %s (%s, judged by %s): %q\n",
		clock.Format(mt.replicaTickSafe()), strings.Join(names, ", "), p.ID, describeClaim(&p.Claim), text))
	return &p, ""
}

// prophecyRefusal maps a prophecy.declared door rejection to in-fiction
// counsel (the planRefusal shape): the reducer's error strings are the
// source, translated so the model hears a repairable reason.
func prophecyRefusal(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "no charges banked"):
		return "no charges are banked — a prophecy needs a stake"
	case strings.Contains(msg, "cap 3"):
		return "I have already staked as many prophecies as I may — the word must be spent carefully"
	case strings.Contains(msg, "already holds"):
		return "that has already come to pass — the past needs no prophet"
	case strings.Contains(msg, "already stakes that claim"):
		return "I have already staked that very claim — one word, one wager"
	case strings.Contains(msg, "unknown designation"):
		return "I keep no designation by that name"
	case strings.Contains(msg, "is dead"):
		return "one of those villagers is beyond reach now"
	case strings.Contains(msg, "living villager"):
		return "that villager is beyond reach now"
	case strings.Contains(msg, "unknown structure kind"):
		return "no such structure can stand in this world"
	case strings.Contains(msg, "text length"):
		return "the word is too long — speak it more briefly"
	case strings.Contains(msg, "game days"):
		return "a prophecy may stand only for a handful of days"
	case strings.Contains(msg, "outside"):
		return "that threshold is beyond what the world can count"
	default:
		return "the world would not let me declare it (" + msg + ")"
	}
}

// describeClaim renders a claim in plain words — the soul-append and prompt
// vocabulary, one home (the describeDesignationSite shape).
func describeClaim(c *sim.ProphecyClaim) string {
	switch c.Kind {
	case sim.ProphecyDesignationFulfilled:
		return fmt.Sprintf("the designation %s fulfilling", c.DesignationID)
	case sim.ProphecyStructureCount:
		return fmt.Sprintf("at least %d %s standing", c.Min, c.StructureKind)
	case sim.ProphecyPopulationAtLeast:
		return fmt.Sprintf("at least %d villagers living", c.Min)
	case sim.ProphecySurvives:
		name := "a villager"
		if c.Agent >= 0 && c.Agent < len(sim.AgentNames) {
			name = sim.AgentNames[c.Agent]
		}
		return fmt.Sprintf("%s living to the appointed day", name)
	}
	return c.Kind
}

// handleProphesy wraps landProphesy: door accept → landed (PlanReport on the
// shared result, the plan-verb report shape); refusal → rejected_gate the
// model may repair within the round cap.
func (mt *Guardian) handleProphesy(d *turnDispatch) toolloop.Handler {
	return func(_ context.Context, call llm.ToolCall) toolloop.Outcome {
		min, hasMin := argInt(call.Args, "min")
		days, _ := argInt(call.Args, "deadline_days")
		a := &prophesyArgs{
			targets:       argString(call.Args, "targets"),
			text:          argString(call.Args, "text"),
			claimKind:     argString(call.Args, "claim_kind"),
			designationID: argString(call.Args, "designation_id"),
			structureKind: argString(call.Args, "structure_kind"),
			min:           min,
			hasMin:        hasMin,
			agentName:     argString(call.Args, "agent"),
			deadlineDays:  days,
		}
		p, why := mt.landProphesy(a, d.charges, d.tick, d.alive, d.grant)
		if p == nil {
			return toolloop.Outcome{Verdict: toolloop.VerdictRejectedGate, ResultForModel: refusal(why)}
		}
		d.result.Plan = &PlanReport{ID: p.ID,
			Summary: fmt.Sprintf("prophesied to %d villager(s), judged by %s", len(p.Targets), describeClaim(&p.Claim))}
		return toolloop.Outcome{Verdict: toolloop.VerdictLanded,
			ResultForModel: "the word is given (" + p.ID + ") — it cannot be unsaid, and the world will judge it"}
	}
}
