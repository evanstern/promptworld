package sim

// Prophecy — the guardian's charge-priced, uncancellable, deadline-bounded
// declared claim (spec 085 US3): the risk half of the faith economy. The
// verification rule (FR-007): a vision is 'true' exactly when its recorded
// claim — declared BEFORE the fact, from the closed predicate vocabulary
// below — is satisfied by recorded world state within its deadline, judged by
// pure (state, tick) predicates in the executor sweep and re-validated at the
// reducer door. Free text is never graded; no model output participates.
//
// Entity discipline clones spec 084's GuardianOrder family (plans.go):
// deterministic guardian-minted ids ("pro-<tick>-<seq>", no RNG), one-way
// status transitions resolved at the reducer door, payload Status/PlacedSeq
// ignored and reducer-stamped, the active cap validated in the arm, and the
// active+recent-32 retention prune. Door split: prophecy.declared is INJECTED
// through InjectSocial (whitelist + the validating arm below);
// prophecy.fulfilled and prophecy.failed are EXECUTOR-emitted from stepEvents
// (prophecyEvents) — the charge_regenerated pattern, so replay reproduces the
// whole lifecycle with no guardian running. There is NO cancel verb (research
// R8): the word, once given, stands.

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/evanstern/promptworld/internal/store"
)

// GuardianProphecyCap bounds concurrent ACTIVE prophecies (spec 085 FR-006) —
// the GuardianDirectiveCap shape: with the charge price it bounds claim spam
// independently of the −15 penalty.
const GuardianProphecyCap = 3

// Prophecy claim kinds (data-model §5) — the CLOSED verification vocabulary.
// FROZEN serialized identifiers (spec 052 ruling 2): they land in recorded
// prophecy.declared payloads. Each kind defines a pure fulfil condition and a
// pure fail condition (prophecyClaimFulfilled / prophecyClaimFailed below);
// growing the vocabulary is a one-row + one-predicate change per kind.
const (
	ProphecyDesignationFulfilled = "designation_fulfilled"
	ProphecyStructureCount       = "structure_count"
	ProphecyPopulationAtLeast    = "population_at_least"
	ProphecySurvives             = "survives"
)

// Prophecy claim bounds (data-model §5, door-validated).
const (
	prophecyStructureCountMin = 1
	prophecyStructureCountMax = 64
)

// Prophecy is one charge-priced, uncancellable, deadline-bounded declared
// claim (spec 085 FR-006). Targets are living villager indices RESOLVED AT
// DECLARATION (ascending, unique) — the payload is self-contained and replay
// never re-resolves names. Its lifecycle (active → fulfilled | failed) has
// two terminals, both one-way, judged ONLY by the claim's own conditions:
// unlike a directive, a prophecy never expires all-targets-dead — the word
// was spoken to the world, not contingent on its hearers.
type Prophecy struct {
	ID           string        `json:"id"`                   // "pro-<declaredTick>-<seq>", the nextOrderID shape, no RNG
	Targets      []int         `json:"targets"`              // living at declaration; ascending, unique
	Village      bool          `json:"village,omitempty"`    // declared to "everyone" (provenance marker)
	Text         string        `json:"text"`                 // 1..400 runes (NudgeTextMax — the registry TextCapBytes single source)
	Claim        ProphecyClaim `json:"claim"`                // stored NORMALIZED (data-model §5)
	DeclaredTick int64         `json:"declared_tick"`        // history → rebase KEEP
	DeadlineTick int64         `json:"deadline_tick"`        // future deadline → rebase SHIFT while active
	Status       string        `json:"status"`               // "active" | "fulfilled" | "failed" (one-way)
	PlacedSeq    int64         `json:"placed_seq,omitempty"` // reducer-stamped from e.Seq (the Designation contract)
}

// ProphecyClaim is the discriminated, machine-checkable claim (FR-007). Only
// the kind's own fields may be set (the door refuses extras); normalized-claim
// equality — plain field equality over this struct minus nothing — is the
// duplicate-rejection key. Agent participates only for the survives kind
// (index 0 is a legal villager; for every other kind it must be zero/absent).
type ProphecyClaim struct {
	Kind          string `json:"kind"`
	DesignationID string `json:"designation_id,omitempty"`
	StructureKind string `json:"structure_kind,omitempty"`
	Min           int    `json:"min,omitempty"`
	Agent         int    `json:"agent,omitempty"`
}

// prophecyByID returns the prophecy named id, or nil. Linear over the bounded
// slice (active ≤ 3 + retained 32), matching every other entity scan.
func (s *State) prophecyByID(id string) *Prophecy {
	for i := range s.Prophecies {
		if s.Prophecies[i].ID == id {
			return &s.Prophecies[i]
		}
	}
	return nil
}

// prophecyClaimFulfilled is the per-kind fulfil condition (data-model §5): a
// PURE (state, tick) check — no clock reads, no RNG, deterministic iteration —
// evaluated IDENTICALLY by the executor sweep (to emit prophecy.fulfilled),
// by the reducer arm (to re-validate before transitioning), and by the
// declaration door (to refuse a claim already true — prophesying the past).
// The by-deadline kinds (rows 1–3) ignore the deadline here: the sweep only
// consults them while the prophecy is active, and failed latches one-way, so
// a condition turning true after the deadline mints nothing. survives is the
// at-deadline kind expressed in the same grammar: alive AT the deadline.
func prophecyClaimFulfilled(s *State, p *Prophecy, tick int64) bool {
	c := &p.Claim
	switch c.Kind {
	case ProphecyDesignationFulfilled:
		d := s.designationByID(c.DesignationID)
		return d != nil && d.Status == "fulfilled"
	case ProphecyStructureCount:
		n := 0
		for i := range s.Structures {
			if s.Structures[i].Kind == c.StructureKind {
				n++
			}
		}
		return n >= c.Min
	case ProphecyPopulationAtLeast:
		return livingCount(s) >= c.Min
	case ProphecySurvives:
		return tick >= p.DeadlineTick && agentAlive(s, c.Agent)
	}
	return false
}

// prophecyClaimFailed is the per-kind fail condition (data-model §5), pure
// like the fulfil half. The sweep and the arm both check fulfil FIRST, so at
// a shared boundary exactly one terminal ever lands (fulfilled wins — the
// directive sweep's precedent). survives fails FAST on death (the claim is
// already unsatisfiable — no deadline wait); every other kind fails at the
// first boundary ≥ deadline with the fulfil condition not holding.
func prophecyClaimFailed(s *State, p *Prophecy, tick int64) bool {
	if p.Claim.Kind == ProphecySurvives {
		return !agentAlive(s, p.Claim.Agent)
	}
	return tick >= p.DeadlineTick && !prophecyClaimFulfilled(s, p, tick)
}

// agentAlive reports whether idx names a living villager.
func agentAlive(s *State, idx int) bool {
	return idx >= 0 && idx < len(s.Agents) && !s.Agents[idx].Dead
}

// Prophecy terminal companion memories (data-model §8): word spreads — every
// living target learns the outcome as an honestly-secondhand OriginReport
// memory (DirectPerception stays false, so the spec-030 provenance gate never
// launders it into "witnessed"). Mid band: memorable social texture, well
// below the generation-interrupting and rumor-seed thresholds; the failed
// word carries a negative tone. Personal (Subject -1) — no gossip seeding in
// v1 (guardian-subject rumors are identified, deferred).
const (
	salProphecyReport  = 5
	toneProphecyFailed = -30
)

// prophecyEvents is the verification sweep (spec 085 FR-008): per active
// prophecy in slice order, each boundary tick, fulfil condition BEFORE fail
// condition — a pure function of (pre-tick state, tick), the
// designation-fulfillment sweep's idiom. Emitted once: the same event flips
// the prophecy non-active, so the next tick's sweep skips it. Each terminal
// rides with one companion OriginReport memory per living target in the same
// batch; the faith sweep (faithEvents) later in stepEvents observes the
// terminal and mints the faith companion in-batch too.
func prophecyEvents(s *State, nextTick int64) []store.Event {
	var events []store.Event
	for i := range s.Prophecies {
		p := &s.Prophecies[i]
		if p.Status != "active" {
			continue
		}
		switch {
		case prophecyClaimFulfilled(s, p, nextTick):
			events = append(events, store.Event{Tick: nextTick, Type: "prophecy.fulfilled",
				Payload: mustPayload(OrderIDPayload{ID: p.ID})})
			for _, t := range p.Targets {
				if !agentAlive(s, t) {
					continue // companions simply skip the dead (US3 AS-7)
				}
				a := &s.Agents[t]
				events = append(events, situatedMemoryEvent(nextTick, t, salProphecyReport,
					PlaceAt(s, a.X, a.Y), "", OriginReport,
					"%s", "The Guardian's foretelling came true — "+p.Text))
			}
		case prophecyClaimFailed(s, p, nextTick):
			events = append(events, store.Event{Tick: nextTick, Type: "prophecy.failed",
				Payload: mustPayload(OrderIDPayload{ID: p.ID})})
			for _, t := range p.Targets {
				if !agentAlive(s, t) {
					continue
				}
				a := &s.Agents[t]
				events = append(events, situatedMemoryToned(nextTick, t, salProphecyReport, toneProphecyFailed,
					PlaceAt(s, a.X, a.Y), "", OriginReport,
					"%s", "The appointed time passed; the Guardian's word did not come to pass."))
			}
		}
	}
	return events
}

// applyProphecy is the reducer arm family for the prophecy.* vocabulary
// (spec 085 FR-008). The declared arm validates rather than clamps — the
// InjectSocial dry-run runs it on a state copy, so an invalid declaration is
// rejected at the door and recorded events always re-apply cleanly at the
// same position in replay. The two executor-emitted terminals re-validate
// their emitting condition, keeping the door authoritative even for
// sim-authored events (the applyPlan contract).
func (s *State) applyProphecy(e store.Event) error {
	switch e.Type {
	case "prophecy.declared":
		var p Prophecy
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		// The stake (US3 AS-1): a prophecy spends one charge — the
		// metatron.nudged arm's contract, so the spend is event-sourced (the
		// declaration IS the spend's record) and replay reproduces the
		// economy. Validated before anything else lands (validate-not-clamp).
		if s.GuardianCharges <= 0 {
			return fmt.Errorf("apply %s: no charges banked", e.Type)
		}
		if err := s.validateProphecyDeclared(e.Type, &p); err != nil {
			return err
		}
		// Status is IGNORED on the payload — a prophecy always lands active;
		// PlacedSeq is reducer-stamped from the event's own store seq (the
		// designation.placed shape: identical live and in replay).
		p.Status = "active"
		p.PlacedSeq = e.Seq
		s.GuardianCharges--
		s.Prophecies = prunePlanEntities(append(s.Prophecies, p),
			func(x Prophecy) bool { return x.Status == "active" })
	case "prophecy.fulfilled":
		var p OrderIDPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		pro := s.prophecyByID(p.ID)
		if pro == nil {
			return fmt.Errorf("apply %s: unknown prophecy %q", e.Type, p.ID)
		}
		// Re-validate the fulfil condition against current state before
		// transitioning (the designation.fulfilled arm's contract).
		if !prophecyClaimFulfilled(s, pro, e.Tick) {
			return fmt.Errorf("apply %s: prophecy %q's claim does not hold", e.Type, p.ID)
		}
		return s.transitionProphecy(e.Type, p.ID, "fulfilled")
	case "prophecy.failed":
		var p OrderIDPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		pro := s.prophecyByID(p.ID)
		if pro == nil {
			return fmt.Errorf("apply %s: unknown prophecy %q", e.Type, p.ID)
		}
		// Re-validate the fail condition — and fulfil-first: a claim that
		// holds at apply time refuses the failure (exactly one terminal).
		if prophecyClaimFulfilled(s, pro, e.Tick) || !prophecyClaimFailed(s, pro, e.Tick) {
			return fmt.Errorf("apply %s: prophecy %q's fail condition does not hold", e.Type, p.ID)
		}
		return s.transitionProphecy(e.Type, p.ID, "failed")
	}
	return nil
}

// validateProphecyDeclared is the prophecy.declared door (data-model §4,
// validate-not-clamp — the dry-run is the door): id discipline, resolved
// living targets, text runes, the shared TTL bounds, the active cap, the
// closed claim vocabulary with kind-conditional fields, the already-true
// refusal (prophesying the past), and the active-duplicate refusal (faith
// cannot be farmed by restating one truth).
func (s *State) validateProphecyDeclared(eventType string, p *Prophecy) error {
	if p.ID == "" {
		return fmt.Errorf("apply %s: empty prophecy id", eventType)
	}
	// Duplicate id in ANY status is rejected — ids are assigned once and
	// consumed prophecies are retained (the order-placed discipline).
	for i := range s.Prophecies {
		if s.Prophecies[i].ID == p.ID {
			return fmt.Errorf("apply %s: duplicate prophecy id %q", eventType, p.ID)
		}
	}
	if len(p.Targets) == 0 {
		return fmt.Errorf("apply %s: prophecy has no targets", eventType)
	}
	for i, t := range p.Targets {
		if t < 0 || t >= len(s.Agents) {
			return fmt.Errorf("apply %s: target index %d out of range", eventType, t)
		}
		if i > 0 && p.Targets[i-1] >= t {
			return fmt.Errorf("apply %s: targets not ascending-unique", eventType)
		}
		if s.Agents[t].Dead {
			return fmt.Errorf("apply %s: target %s is dead", eventType, s.Agents[t].Name)
		}
	}
	if n := utf8.RuneCountInString(p.Text); n == 0 || n > NudgeTextMax {
		return fmt.Errorf("apply %s: text length %d outside 1..%d runes", eventType, n, NudgeTextMax)
	}
	if ttl := p.DeadlineTick - p.DeclaredTick; ttl < GuardianOrderTTLMinDays*ticksPerGameDay || ttl > GuardianOrderTTLMaxDays*ticksPerGameDay {
		return fmt.Errorf("apply %s: deadline %d ticks outside %d..%d game days", eventType, ttl, GuardianOrderTTLMinDays, GuardianOrderTTLMaxDays)
	}
	active := 0
	for i := range s.Prophecies {
		if s.Prophecies[i].Status == "active" {
			active++
		}
	}
	if active >= GuardianProphecyCap {
		return fmt.Errorf("apply %s: %d prophecies already active (cap %d)", eventType, active, GuardianProphecyCap)
	}
	if err := s.validateProphecyClaim(eventType, &p.Claim); err != nil {
		return err
	}
	// A claim already true at declaration is refused — prophesying the past
	// (US3 AS-4). The fulfil predicate at the declaration tick is the check;
	// for survives this can never hold before the deadline (deadline is
	// strictly future by the TTL bounds), so the door never trips on it.
	if prophecyClaimFulfilled(s, p, p.DeclaredTick) {
		return fmt.Errorf("apply %s: the claim already holds — the past needs no prophet", eventType)
	}
	// An identical claim to an already-active prophecy is refused
	// (normalized-claim equality: plain field equality over ProphecyClaim).
	for i := range s.Prophecies {
		if s.Prophecies[i].Status == "active" && s.Prophecies[i].Claim == p.Claim {
			return fmt.Errorf("apply %s: an active prophecy already stakes that claim", eventType)
		}
	}
	return nil
}

// validateProphecyClaim enforces the closed vocabulary and its
// kind-conditional fields (data-model §5): required fields present and valid,
// foreign fields zero (extras refused — the parseReveal partial-args shape,
// re-checked reducer-side so the payload is authoritative).
func (s *State) validateProphecyClaim(eventType string, c *ProphecyClaim) error {
	switch c.Kind {
	case ProphecyDesignationFulfilled:
		if c.DesignationID == "" {
			return fmt.Errorf("apply %s: designation_fulfilled needs a designation_id", eventType)
		}
		if c.StructureKind != "" || c.Min != 0 || c.Agent != 0 {
			return fmt.Errorf("apply %s: designation_fulfilled takes only designation_id", eventType)
		}
		// Any active-or-past id is claimable — the claim may name in-flight
		// work; NOT-already-fulfilled is the shared already-true check.
		if s.designationByID(c.DesignationID) == nil {
			return fmt.Errorf("apply %s: unknown designation %q", eventType, c.DesignationID)
		}
	case ProphecyStructureCount:
		if c.StructureKind == "" {
			return fmt.Errorf("apply %s: structure_count needs a structure_kind", eventType)
		}
		if !buildableStructureKind(c.StructureKind) {
			return fmt.Errorf("apply %s: unknown structure kind %q", eventType, c.StructureKind)
		}
		if c.Min < prophecyStructureCountMin || c.Min > prophecyStructureCountMax {
			return fmt.Errorf("apply %s: min %d outside %d..%d", eventType, c.Min, prophecyStructureCountMin, prophecyStructureCountMax)
		}
		if c.DesignationID != "" || c.Agent != 0 {
			return fmt.Errorf("apply %s: structure_count takes only structure_kind and min", eventType)
		}
	case ProphecyPopulationAtLeast:
		if c.Min < 1 || c.Min > len(s.Agents) {
			return fmt.Errorf("apply %s: min %d outside 1..%d", eventType, c.Min, len(s.Agents))
		}
		if c.DesignationID != "" || c.StructureKind != "" || c.Agent != 0 {
			return fmt.Errorf("apply %s: population_at_least takes only min", eventType)
		}
	case ProphecySurvives:
		if !agentAlive(s, c.Agent) {
			return fmt.Errorf("apply %s: survives needs a living villager", eventType)
		}
		if c.DesignationID != "" || c.StructureKind != "" || c.Min != 0 {
			return fmt.Errorf("apply %s: survives takes only agent", eventType)
		}
	default:
		return fmt.Errorf("apply %s: unknown claim kind %q", eventType, c.Kind)
	}
	return nil
}

// transitionProphecy moves the prophecy named id from active to a terminal
// status (one-way) — the transitionGuardianOrder race shape: exactly one
// terminal lands, and the loser hits a non-active entity and refuses. A
// condition turning true AFTER failed latched therefore mints nothing (the
// "verifies after the TTL" edge).
func (s *State) transitionProphecy(eventType, id, to string) error {
	for i := range s.Prophecies {
		if s.Prophecies[i].ID != id {
			continue
		}
		if s.Prophecies[i].Status != "active" {
			return fmt.Errorf("apply %s: prophecy %q is not active (status %q)", eventType, id, s.Prophecies[i].Status)
		}
		s.Prophecies[i].Status = to
		return nil
	}
	return fmt.Errorf("apply %s: unknown prophecy %q", eventType, id)
}
