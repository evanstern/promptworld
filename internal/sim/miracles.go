package sim

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tool"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Guardian's miracles (spec 016): four recorded, charge-priced world edits that
// land through the same InjectSocial door the nudge uses. Like applyGuardian,
// these arms validate rather than clamp — the door's dry-run runs them on a
// state copy, so an invalid miracle is rejected before recording and a recorded
// miracle always re-applies cleanly at the same log position in replay
// (spec 016 R1). Charge pricing lives here in the reducer (not the console) so
// replay re-validates every spend identically (R2). `gratis` waives the charge
// and nothing else; it is unreachable from model output by construction (the
// guardian's turn contract has no gratis field).

// Miracle event payloads (canonical JSON, struct-ordered). No new persistent
// entities: miracles mutate existing fields only (data-model.md).
type (
	// TimeSnappedPayload jumps the clock forward to ToTick, re-basing the
	// relative-duration fields so remaining times are preserved (FR-008/009).
	TimeSnappedPayload struct {
		ToTick int64 `json:"to_tick"`
		Gratis bool  `json:"gratis"`
	}
	// ItemGrantedPayload provisions a living villager with known items,
	// reject-never-clamp at the bulk cap (FR-011).
	ItemGrantedPayload struct {
		Agent  AgentRef `json:"agent"`
		Kind   string   `json:"kind"`
		Qty    int      `json:"qty"`
		Gratis bool     `json:"gratis"`
	}
	// EntityMovedPayload relocates a villager, structure, or pile from (X,Y) to
	// (ToX,ToY) (FR-014).
	EntityMovedPayload struct {
		Class  string `json:"class"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
		ToX    int    `json:"to_x"`
		ToY    int    `json:"to_y"`
		Gratis bool   `json:"gratis"`
	}
	// EntityRemovedPayload deletes a structure, pile, or terrain overlay target
	// at (X,Y); villagers may never be removed (v1 doctrine).
	EntityRemovedPayload struct {
		Class  string `json:"class"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
		Gratis bool   `json:"gratis"`
	}
)

// miracleCost is the reducer's per-event charge table — DERIVED from the single
// authoritative source in internal/tool (spec 021 R7 / FR-009 / SC-004), not a
// second local copy. The time snap is the expensive one (2 charges), every
// other miracle costs 1; that doctrine now lives once in tool.miracleCosts, and
// tool.MiracleCostsByEvent maps it into the event-keyed shape this reducer keys
// on. A cost change in that one table propagates here with no second edit —
// TestMiracleCostDerivedFromTool pins the derivation so a re-introduced literal
// fails the build. Pricing is still doctrine, not caller input (a payload never
// carries its own price); keyed lookup only, never iterated into state.
var miracleCost = tool.MiracleCostsByEvent()

// spendMiracleCharge is the shared validate/spend helper for every miracle arm.
// It checks the bank against the event's cost and decrements it — UNLESS gratis,
// which waives the charge (and only the charge; every other validation still
// runs). It must be called only after all other validation has passed and
// before any mutation, so a rejected miracle spends nothing and leaves no
// partial application (validate-not-clamp, reject-whole).
func (s *State) spendMiracleCharge(eventType string, gratis bool) error {
	cost, ok := miracleCost[eventType]
	if !ok {
		return fmt.Errorf("apply %s: no cost defined", eventType)
	}
	if gratis {
		return nil
	}
	if s.GuardianCharges < cost {
		return fmt.Errorf("apply %s: need %d charge(s), only %d banked", eventType, cost, s.GuardianCharges)
	}
	s.GuardianCharges -= cost
	return nil
}

// applyMiracle is the reducer dispatcher for the four guardian.* miracle event
// types, routed here from State.Apply. Each arm validates, prices via
// spendMiracleCharge, and mutates — atomically, or errors with nothing changed.
func (s *State) applyMiracle(e store.Event) error {
	switch e.Type {
	case "guardian.time_snapped":
		return s.applyTimeSnapped(e)
	case "guardian.item_granted":
		return s.applyItemGranted(e)
	case "guardian.entity_moved":
		return s.applyEntityMoved(e)
	case "guardian.entity_removed":
		return s.applyEntityRemoved(e)
	}
	return fmt.Errorf("apply %s: unknown working type", e.Type)
}

// applyTimeSnapped jumps the clock forward to ToTick with shift semantics
// (spec 016 US3, FR-008/009/010): the world is frozen and only re-labelled, so
// every in-progress relative duration is preserved (rebaseTicks) while history
// stays put. Forward-only — a target at or before the current tick is rejected
// whole, before any charge spend or mutation. The snap costs 2 charges (the
// dearest miracle) unless gratis. FR-010 (a snap mints no charges across the
// skipped regeneration boundaries) needs no code: regeneration is emitted only
// when the executor *processes* a boundary crossing, and a snap processes no
// interval — the skipped boundaries simply never fire.
func (s *State) applyTimeSnapped(e store.Event) error {
	var p TimeSnappedPayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return fmt.Errorf("apply %s: %w", e.Type, err)
	}
	if p.ToTick <= s.Tick {
		return fmt.Errorf("apply %s: target tick %d is not after the current tick %d (time only moves forward)", e.Type, p.ToTick, s.Tick)
	}
	if err := s.spendMiracleCharge(e.Type, p.Gratis); err != nil {
		return err
	}
	rebaseTicks(s, p.ToTick-s.Tick)
	s.Tick = p.ToTick
	return nil
}

// rebaseTicks shifts every relative-duration field in the state tree forward by
// delta, so a time snap preserves remaining durations while history stays put
// (FR-009). It is the SINGLE authority for shift semantics; State.Tick itself is
// set by the caller (applyTimeSnapped), never here.
//
// DOCTRINE — every tick-anchored int64 field in the sim state tree MUST be
// classified here, SHIFT or KEEP, and the taxonomy guard test
// (TestRebaseTaxonomyComplete) fails the build when a new int64 field appears in
// the state structs without a classification entry. The rule:
//
//	SHIFT (+delta) — a future deadline, or an anchor from which an elapsed/
//	  remaining duration is measured; shifting preserves that duration across
//	  the jump. A SHIFT field whose zero value is an "unset/never" sentinel is
//	  shifted ONLY when non-zero (shifting the sentinel would fabricate a value).
//	KEEP — a historical timestamp (when something happened) or an identity/
//	  counter; rewriting it would rewrite history or break a reference.
//
// SHIFT fields:
//
//	Agent.IdleSince      reflex-grace anchor (elapsed = tick-IdleSince); shifted
//	                     UNCONDITIONALLY — its zero is genesis-idle, a real tick
//	                     read by raw subtraction, not a "never" sentinel.
//	Agent.LastTalk       talk cooldown; ONLY non-zero (0 = never, canTalk-checked)
//	Agent.LastGive       gift cooldown; ONLY non-zero (0 = never, canGive-checked)
//	PairTalk.Tick        pair last-exchange cooldown anchor (spec 061: cooldown
//	                     elapsed = tick-Tick, the Agent.LastTalk shape); shifted
//	                     UNCONDITIONALLY — a PRESENT record is always a real
//	                     exchange tick (absence of the record is "never talked"),
//	                     so there is no zero sentinel to guard.
//	Agent.LastMindIntentDone  yield-window anchor (spec 062: elapsed =
//	                     tick-LastMindIntentDone gates prep deference); ONLY
//	                     non-zero (0 = never mind-driven — a no-planner world's
//	                     permanent sentinel — must stay 0 so the window never
//	                     spuriously arms after a snap; the Belief.Reinforced /
//	                     NeedsAnchorTick elapsed-anchor shape)
//	Intent.WorkStart     work-in-progress; ONLY non-zero (0 = not started)
//	AgentHail.Until       courtesy-pause deadline (a present hail is non-zero)
//	PlanStep.Until        plan-step validity deadline; ONLY when > 0 (0 = no
//	                      expiry). NOT in data-model.md — see NOTE.
//	Guard.Tick            after_tick/before_tick boundary; ONLY non-zero (0 for
//	                      the non-timed guard types). NOT in data-model.md — see NOTE.
//	Structure.FuelUntil   fire burn deadline; ONLY non-zero
//	Harvest.Regrow        forage regrowth deadline
//	DenUse.Ready          den cooldown deadline
//	FoodBatch.SpoilAt     ground-food rot deadline
//	Debt.Due              repayment deadline; ONLY non-zero
//	Belief.Reinforced     decay anchor (elapsed = tick-Reinforced); ONLY non-zero
//	                      (0 = legacy grandfather, must stay 0 so it never decays)
//	Gru.LastAttack        attack-cooldown anchor; ONLY non-zero (0 = never) and
//	                      only while the gru is abroad
//	State.ColdSnapUntil   cold-snap expiry deadline (spec 077): read live
//	                      against the clock (coldSnapActive), so a snap
//	                      forward preserves the snap's remaining window —
//	                      the Structure.FuelUntil shape; ONLY non-zero
//	                      (0 = no snap ever)
//	Stranger.LastMove     movement-cadence anchor (spec 077, the
//	Stranger.LastTake     Gru.LastAttack shape); ONLY non-zero (0 = fresh
//	                      arrival) and only while the stranger is abroad
//	Meeting.OpenedTick    assembly-phase anchor; ONLY non-zero (in-flight meeting)
//	Meeting.GatherStart   emergent-gathering-watch anchor; ONLY non-zero
//	GuardianOrder.ExpiresTick  standing-order expiry deadline (spec 029); shifted
//	                      ONLY for ACTIVE orders (a consumed order's deadline is a
//	                      spent artifact), so the remaining lifetime survives the jump
//	Directive.ExpiresTick directive TTL deadline (spec 084, data-model §10) —
//	                      the GuardianOrder.ExpiresTick classification verbatim:
//	                      shifted ONLY for ACTIVE directives
//	PlaceFact.Seen        mental-map freshness anchor (spec 041: fresh iff
//	                      now-Seen < horizon — the Belief.Reinforced shape); ONLY
//	                      non-zero. Left unshifted, a snap would instantly stale
//	                      every villager's knowledge — exactly the drift FR-009
//	                      forbids.
//	PeerSighting.Seen     mental-map sighting recency anchor (spec 041 T013) —
//	                      same shape as PlaceFact.Seen; ONLY non-zero (genesis
//	                      sightings carry tick 0 and stay put, like the
//	                      grandfathered Belief.Reinforced zero).
//	Agent.NeedsAnchorTick trajectory-window edge anchor (spec 043 US2: elapsed =
//	                      tick-NeedsAnchorTick gates the window roll in the
//	                      needs_changed arm); ONLY non-zero (0 = unset/first-
//	                      window sentinel). SHIFT preserves the window's remaining
//	                      time across the jump — exactly the Belief.Reinforced /
//	                      PlaceFact.Seen elapsed-anchor shape. Left unshifted, a
//	                      snap would push tick-NeedsAnchorTick past the window and
//	                      force an immediate anchor reset on the next needs
//	                      change, wiping every villager's trajectory sense (the
//	                      "when it happened" reading is superficial — the field is
//	                      READ live against the clock, unlike the KEEP history
//	                      timestamps below such as IntentRecord.Tick, so the
//	                      duration-anchor rule governs). NeedsAnchor itself holds
//	                      need LEVELS (0-1000 ints, not ticks) frozen with the
//	                      world, so it carries no tick field and needs no entry.
//	NeglectState.FoodSince/WarmthSince/RestSince  band-entry anchors (spec 083:
//	                      elapsed = tick-Since gates the neglect window in
//	                      NeglectDue); ONLY non-zero (0 = not in band). Left
//	                      unshifted, a snap would age every in-band episode past
//	                      the window and fire the detector the instant it landed.
//	NeglectState.FoodIntent/WarmthIntent/RestIntent  last-class-intent stamps
//	                      (spec 083: elapsed = tick-Intent is the zero-intent
//	                      clause — read live against the clock, the
//	                      Belief.Reinforced elapsed-anchor shape, NOT a history
//	                      record like IntentRecord.Tick); ONLY non-zero (0 =
//	                      never). The Fired latches are bools (episode state,
//	                      no tick field, no entry).
//
// KEEP (history/identity — never rewritten): Agent.Generation,
//
//	GuardianOrder.PlacedTick (spec 029: when the order was placed, history),
//	Designation.PlacedTick/PlacedSeq and Directive.IssuedTick/PlacedSeq
//	(spec 084, data-model §10: placement/issue history and event-seq
//	identities — a designation carries no future deadline at all),
//	Agent.LastGoalTick, Agent.LastConsolidatedNight, Agent.ConsolidatedUpTo,
//	Agent.LastConsolidateMark, IntentRecord.Tick / IntentRecord.OutcomeTick
//	(spec 043: when an intent landed / when its outcome landed — a historical
//	self-history log, like Memory.Tick; the ring records what happened, never a
//	future deadline), Memory.Tick, Memory.Conv (spec 019: a
//	conversation-ref identity, same founding-talk tick as ConvoRecord.Conv),
//	Memory.Seq (spec 042: the emitting event's store seq — an identity, never
//	a clock value), Agent.SitVecTick (spec 042: when the situation text was
//	rendered — history/audit, like Memory.Tick),
//	JournalEntry.Tick (spec 019: when the entry was written, history),
//	Belief.Tick, KnownRumor.Tick,
//	Guard.Generation, Rumor.OriginTick, ConvoRecord.Conv (identity — the
//	founding-talk tick doubles as the conversation id), ConvoRecord.Tick,
//	GuardianReportCard.Tick/Seq/Citations (spec 063: when the card landed +
//	the card event's identity + cited event seqs — history and identities,
//	never deadlines),
//	StrangerTake.Tick (spec 077: when the take happened — a historical
//	ledger fact, the DeathRecord.Tick shape), Stranger.Night (spec 077: the
//	1-based arrival night — identity, like Rumor.OriginTick),
//	State.CharterObservedSeq/CharterObservedTick and
//	State.SkillsObservedSeq/SkillsObservedTick (spec 077: log coordinates
//	re-locating recorded observation events — identities, the Memory.Seq /
//	EvidenceRef.Tick shape; rewriting them would point evidence at ticks
//	where nothing was recorded),
//	ChronicleEntry.Tick/Day/FromTick/ToTick, Meeting.LastMeetingDay,
//	MeetingConvention.EstablishedDay, Norm.DayPassed/DayRepealed/DayAmended,
//	NormViolation.Tick. Day-denominated governance fields re-arm naturally under
//	the new clock. PlaceFact.Detail (spec 041) is KEEP: a remembered value
//	baked at emission, never re-derived — for a fire it mirrors the (shifted)
//	FuelUntil as LAST SEEN, and rewriting what an agent remembers would rewrite
//	its history; the perception sweep simply re-witnesses the shifted reality
//	on the next look. Other kinds bake non-tick scalars here, so a blanket
//	shift would corrupt them.
//
// PHASE-ANCHORED behavior (day/night, meeting times of day, charge-regen
// boundaries) is a pure function of the absolute clock and stores no field here.
//
// NOTE (deviation from data-model.md, recorded for the planning tier):
// PlanStep.Until (internal/sim/plan.go:28) and Guard.Tick (internal/sim/guard.go:26)
// are reachable from State via Agent.Plan[].When but were NOT listed in the
// data-model.md taxonomy. They are genuine future deadlines — plan.go calls
// timed guards "the sole act-at-time-T mechanism" — so FR-009's catch-all ("any
// future duration-anchored state") requires shifting them: left unshifted, a
// snap that jumped past their absolute tick would expire a pending plan step or
// fire a timed guard the instant it landed, exactly the drift the feature forbids.
func rebaseTicks(s *State, delta int64) {
	shift := func(p *int64) { // shift a deadline, honoring the zero=never sentinel
		if *p != 0 {
			*p += delta
		}
	}
	for i := range s.Agents {
		a := &s.Agents[i]
		a.IdleSince += delta // unconditional: zero is genesis-idle, not "never"
		shift(&a.LastTalk)
		shift(&a.LastGive)
		shift(&a.LastMindIntentDone) // spec 062 US1: yield-window anchor (Belief.Reinforced shape); 0 = never mind-driven, stays 0
		shift(&a.NeedsAnchorTick)    // spec 043 US2: trajectory-window edge anchor; 0 = unset, stays 0
		if a.Neglect != nil {
			// Spec 083: the six neglect anchors are duration anchors (elapsed
			// against the live clock gates the window) — SHIFT, zero=never/
			// not-in-band sentinels stay zero. The Fired latches are episode
			// state (bools), untouched by construction.
			shift(&a.Neglect.FoodSince)
			shift(&a.Neglect.WarmthSince)
			shift(&a.Neglect.RestSince)
			shift(&a.Neglect.FoodIntent)
			shift(&a.Neglect.WarmthIntent)
			shift(&a.Neglect.RestIntent)
		}
		for j := range a.Beliefs {
			shift(&a.Beliefs[j].Reinforced) // spec 030: decay anchor (elapsed = tick-Reinforced); 0 = grandfather, stays 0
		}
		if a.Intent != nil {
			shift(&a.Intent.WorkStart)
		}
		if a.Hail != nil {
			shift(&a.Hail.Until)
		}
		for j := range a.Plan {
			shift(&a.Plan[j].Until)
			if a.Plan[j].When != nil {
				shift(&a.Plan[j].When.Tick)
			}
		}
		if a.Map != nil {
			for j := range a.Map.Facts {
				// Spec 041: the freshness anchor shifts so knowledge does not
				// stale across the jump; Detail is remembered history (KEEP).
				shift(&a.Map.Facts[j].Seen)
			}
			for j := range a.Map.Peers {
				shift(&a.Map.Peers[j].Seen)
			}
		}
	}
	for i := range s.Structures {
		shift(&s.Structures[i].FuelUntil)
	}
	for i := range s.Harvested {
		shift(&s.Harvested[i].Regrow)
	}
	for i := range s.DenUses {
		shift(&s.DenUses[i].Ready)
	}
	for i := range s.Piles {
		for j := range s.Piles[i].Food {
			shift(&s.Piles[i].Food[j].SpoilAt)
		}
	}
	for i := range s.Debts {
		shift(&s.Debts[i].Due)
	}
	for i := range s.GuardianOrders {
		// A standing order's expiry is a future deadline: shift only ACTIVE orders
		// so the remaining lifetime is preserved across the jump. PlacedTick is a
		// historical timestamp and is left unshifted (KEEP).
		if s.GuardianOrders[i].Status == "active" {
			shift(&s.GuardianOrders[i].ExpiresTick)
		}
	}
	for i := range s.Directives {
		// Spec 084 (data-model §10): a directive's expiry is a future deadline
		// — the GuardianOrders classification verbatim: shift only ACTIVE
		// directives so the remaining lifetime survives the jump. IssuedTick
		// and PlacedSeq are history/identity (KEEP). Designations carry no
		// future deadline at all — PlacedTick/PlacedSeq are history (KEEP,
		// untouched) — and the "designation" place facts' Seen anchors already
		// SHIFT in the per-agent a.Map.Facts loop above (no new code).
		if s.Directives[i].Status == "active" {
			shift(&s.Directives[i].ExpiresTick)
		}
	}
	for i := range s.Prophecies {
		// Spec 085 (data-model §9): an ACTIVE prophecy's DeadlineTick is a
		// future deadline — the Directive.ExpiresTick arm's clone: shift it so
		// the remaining judgment window survives the jump. DeclaredTick is
		// history (KEEP); a settled prophecy's ticks are spent artifacts
		// (KEEP). FaithState carries no tick fields at all — untouched.
		if s.Prophecies[i].Status == "active" {
			shift(&s.Prophecies[i].DeadlineTick)
		}
	}
	if s.Gru != nil {
		shift(&s.Gru.LastAttack)
	}
	// Spec 077: the snap's remaining window and the stranger's cadence
	// anchors survive the jump (SHIFT); the take ledger's ticks and the
	// observation coordinates are history/identity (KEEP — untouched).
	shift(&s.ColdSnapUntil)
	if s.Stranger != nil {
		shift(&s.Stranger.LastMove)
		shift(&s.Stranger.LastTake)
	}
	shift(&s.Meeting.OpenedTick)
	shift(&s.Meeting.GatherStart)
	for i := range s.PairTalks {
		// Spec 061 (TASK-109): a present pair record's Tick is a real exchange
		// anchor (cooldown elapsed = tick-Tick, the Agent.LastTalk shape), never
		// a "never" sentinel — absence of the record is "never talked" — so it
		// shifts UNCONDITIONALLY, preserving the remaining pair cooldown across
		// the snap exactly as Agent.LastTalk does.
		s.PairTalks[i].Tick += delta
	}
}

// applyItemGranted provisions a living villager with known items, reject-never-
// clamp at the bulk cap (spec 016 US4, FR-011). Every validation — a valid,
// living agent, a known item kind, a positive quantity, and the cap check —
// precedes the charge spend and the mutation, so a rejected grant spends nothing
// and leaves no partial application (validate-not-clamp, reject-whole). The
// grant weighs one bulk per unit, exactly as bulk() weighs a carried item: a
// fresh spear counts one (durability lives in the slice, len(Spears) per bulk),
// so a grant of qty items always costs qty bulk regardless of kind.
func (s *State) applyItemGranted(e store.Event) error {
	var p ItemGrantedPayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return fmt.Errorf("apply %s: %w", e.Type, err)
	}
	if p.Agent.ID < 0 || p.Agent.ID >= len(s.Agents) {
		return fmt.Errorf("apply %s: no villager at index %d", e.Type, p.Agent.ID)
	}
	if s.Agents[p.Agent.ID].Dead {
		return fmt.Errorf("apply %s: %s is beyond a grant now", e.Type, s.Agents[p.Agent.ID].Name)
	}
	if !grantableKind(p.Kind) {
		return fmt.Errorf("apply %s: unknown item kind %q (grantable: %s)",
			e.Type, p.Kind, strings.Join(tool.GrantKinds(), ", "))
	}
	if p.Qty <= 0 {
		return fmt.Errorf("apply %s: grant quantity must be positive (got %d)", e.Type, p.Qty)
	}
	inv := &s.Agents[p.Agent.ID].Inv
	// One bulk per granted unit (a fresh spear weighs one, like every other
	// unit), so the grant's bulk is exactly qty — reject whole if it overflows
	// the carry cap, never clamp to a partial delivery (FR-011).
	if bulk(*inv)+p.Qty > bulkCap {
		return fmt.Errorf("apply %s: granting %d %s to %s would exceed the carry cap (%d/%d already used)",
			e.Type, p.Qty, p.Kind, s.Agents[p.Agent.ID].Name, bulk(*inv), bulkCap)
	}
	if err := s.spendMiracleCharge(e.Type, p.Gratis); err != nil {
		return err
	}
	switch p.Kind {
	case "spear":
		// Each granted unit is one fresh, full-durability spear; keep the
		// remaining-uses slice sorted ascending (hunts spend the most-worn first).
		for n := 0; n < p.Qty; n++ {
			inv.Spears = append(inv.Spears, spearDurability)
		}
		sort.Ints(inv.Spears)
	case "axe":
		// Spec 032 US2: each granted unit is one fresh, full-durability axe —
		// the spear-grant clone (durability lives in Inventory.Axes, so it has
		// no invField).
		for n := 0; n < p.Qty; n++ {
			inv.Axes = append(inv.Axes, axeDurability)
		}
		sort.Ints(inv.Axes)
	default:
		addItems(inv, []Item{{Kind: p.Kind, N: p.Qty}}, +1)
	}
	return nil
}

// grantableKinds is the accepted-set built once from tool.GrantKinds() — the
// single authoritative grant vocabulary (TASK-163) — rather than a second
// hand-written literal here. Keyed, never iterated into state.
var grantableKinds = func() map[string]bool {
	m := make(map[string]bool, len(tool.GrantKinds()))
	for _, k := range tool.GrantKinds() {
		m[k] = true
	}
	return m
}()

// grantableKind reports whether kind is one a grant miracle may deliver: the
// Inventory key set plus "spear" (singular — a grant unit is one fresh spear;
// its durability lives in Inventory.Spears, so it has no invField). Derived
// from tool.GrantKinds(), work_miracle's declared `item` Enum, so the door's
// acceptance can never drift from the schema the model is shown.
func grantableKind(kind string) bool {
	return grantableKinds[kind]
}

// applyEntityMoved relocates a villager, structure, or pile (spec 016 US1,
// FR-014). The source class MUST be present at (x,y) and the destination MUST
// satisfy the class's placement rule (villager/pile → passable; structure →
// buildSite) — validated here so the dry-run rejects a bad move at the door and
// replay re-applies a recorded move cleanly. All validation precedes the charge
// spend and the mutation, so a rejected move spends nothing and leaves no
// partial application. Villagers may share a tile, so no destination-exclusivity
// check applies to a villager move.
func (s *State) applyEntityMoved(e store.Event) error {
	var p EntityMovedPayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return fmt.Errorf("apply %s: %w", e.Type, err)
	}
	switch p.Class {
	case "villager":
		idx := s.VillagerAt(p.X, p.Y)
		if idx < 0 {
			return fmt.Errorf("apply %s: no living villager at (%d,%d)", e.Type, p.X, p.Y)
		}
		if !passable(s.m, s, p.ToX, p.ToY) {
			return fmt.Errorf("apply %s: (%d,%d) is not passable", e.Type, p.ToX, p.ToY)
		}
		if err := s.spendMiracleCharge(e.Type, p.Gratis); err != nil {
			return err
		}
		a := &s.Agents[idx]
		a.X, a.Y = p.ToX, p.ToY
		// Cancel-and-replan (clarified): the moved villager's in-flight objective
		// is dropped and it becomes idle at the landing tick, exactly like every
		// other intent-clearing path.
		a.Intent = nil
		a.IdleSince = e.Tick
		// Spec 041 (research D2): a teleported villager knows where it landed —
		// the same derived explored-bit and peer-sighting bookkeeping as a
		// walked step.
		s.markExplored(a, p.ToX, p.ToY)
		s.notePresence(idx, e.Tick)
	case "structure":
		i := s.structureIndexAt(p.X, p.Y)
		if i < 0 {
			return fmt.Errorf("apply %s: no structure at (%d,%d)", e.Type, p.X, p.Y)
		}
		if !buildSite(s.m, s, p.ToX, p.ToY) {
			return fmt.Errorf("apply %s: (%d,%d) is not a valid build site", e.Type, p.ToX, p.ToY)
		}
		if err := s.spendMiracleCharge(e.Type, p.Gratis); err != nil {
			return err
		}
		// The struct moves whole — FuelUntil/Owner/Store ride along in the value.
		s.Structures[i].X, s.Structures[i].Y = p.ToX, p.ToY
	case "pile":
		if s.pileAt(p.X, p.Y) == nil {
			return fmt.Errorf("apply %s: no pile at (%d,%d)", e.Type, p.X, p.Y)
		}
		if !passable(s.m, s, p.ToX, p.ToY) {
			return fmt.Errorf("apply %s: (%d,%d) is not passable", e.Type, p.ToX, p.ToY)
		}
		if err := s.spendMiracleCharge(e.Type, p.Gratis); err != nil {
			return err
		}
		s.movePile(p.X, p.Y, p.ToX, p.ToY)
	default:
		return fmt.Errorf("apply %s: cannot move class %q", e.Type, p.Class)
	}
	return nil
}

// applyEntityRemoved deletes a structure, pile, or terrain overlay target
// (spec 016 US1). Villagers are never removable (v1 doctrine). A chest spills
// its Store to a ground pile before deletion (goods are never silently
// destroyed); a pile is removed with its contents (explicit, operator-visible
// destruction of the named target); terrain routes through the existing overlay
// vocabulary. All validation precedes the charge spend and the mutation.
func (s *State) applyEntityRemoved(e store.Event) error {
	var p EntityRemovedPayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return fmt.Errorf("apply %s: %w", e.Type, err)
	}
	switch p.Class {
	case "villager":
		return fmt.Errorf("apply %s: a villager can never be removed", e.Type)
	case "structure":
		i := s.structureIndexAt(p.X, p.Y)
		if i < 0 {
			return fmt.Errorf("apply %s: no structure at (%d,%d)", e.Type, p.X, p.Y)
		}
		if err := s.spendMiracleCharge(e.Type, p.Gratis); err != nil {
			return err
		}
		st := s.Structures[i]
		if st.Kind == "chest" && st.Store != nil {
			// Reuse the death-spill vocabulary: contents become a ground pile on
			// the tile, food stamped with a fresh rot deadline (never lost).
			s.spillInventory(p.X, p.Y, st.Store, e.Tick)
		}
		s.removeStructureAt(i)
	case "pile":
		if s.pileAt(p.X, p.Y) == nil {
			return fmt.Errorf("apply %s: no pile at (%d,%d)", e.Type, p.X, p.Y)
		}
		if err := s.spendMiracleCharge(e.Type, p.Gratis); err != nil {
			return err
		}
		s.removePileAt(p.X, p.Y)
	case "terrain":
		return s.removeTerrain(e, p)
	default:
		return fmt.Errorf("apply %s: cannot remove class %q", e.Type, p.Class)
	}
	return nil
}

// removeTerrain overlays a tree/forage/rock tile through the SAME vocabulary the
// executor uses (chop→Cleared, forage→Harvested with regrow, quarry→Quarried),
// so a removed tile is a state the executor could already produce. A tile that
// is already overlaid is a no-op target → rejected. The charge is spent only
// after the base kind and the not-already-overlaid check pass.
func (s *State) removeTerrain(e store.Event, p EntityRemovedPayload) error {
	switch s.m.At(p.X, p.Y) {
	case worldmap.Tree:
		if effectiveKind(s.m, s, p.X, p.Y) != worldmap.Tree {
			return fmt.Errorf("apply %s: the tree at (%d,%d) is already cleared", e.Type, p.X, p.Y)
		}
		if err := s.spendMiracleCharge(e.Type, p.Gratis); err != nil {
			return err
		}
		// Mirror agent.chopped: Cleared is a bare Point (reverts to grass).
		s.Cleared = append(s.Cleared, Point{X: p.X, Y: p.Y})
	case worldmap.Forage:
		if effectiveKind(s.m, s, p.X, p.Y) != worldmap.Forage {
			return fmt.Errorf("apply %s: the forage at (%d,%d) is already harvested", e.Type, p.X, p.Y)
		}
		if err := s.spendMiracleCharge(e.Type, p.Gratis); err != nil {
			return err
		}
		// Mirror agent.foraged: standard forage regrow window.
		s.Harvested = append(s.Harvested, Harvest{X: p.X, Y: p.Y, Regrow: e.Tick + forageRegrowSec})
	case worldmap.Rock:
		if effectiveKind(s.m, s, p.X, p.Y) != worldmap.Rock {
			return fmt.Errorf("apply %s: the rock at (%d,%d) is already quarried", e.Type, p.X, p.Y)
		}
		if err := s.spendMiracleCharge(e.Type, p.Gratis); err != nil {
			return err
		}
		// Mirror agent.quarried: permanent depletion (no regrow entry).
		s.Quarried = append(s.Quarried, Point{X: p.X, Y: p.Y})
	default:
		// Marsh and Sand (spec 068, C13) land here deliberately alongside
		// grass/water: the new ground covers have no depleted state the
		// executor could produce, so they are not removable terrain.
		return fmt.Errorf("apply %s: (%d,%d) holds no removable terrain", e.Type, p.X, p.Y)
	}
	return nil
}

// structureIndexAt returns the index of the structure on (x,y), or -1. At most
// one structure ever stands on a tile (buildSite forbids stacking), so the
// first match is the tile's structure.
func (s *State) structureIndexAt(x, y int) int {
	for i := range s.Structures {
		if s.Structures[i].X == x && s.Structures[i].Y == y {
			return i
		}
	}
	return -1
}

// removeStructureAt deletes structure index i, preserving the creation order of
// the survivors. Empties to nil so canonical bytes stay stable.
func (s *State) removeStructureAt(i int) {
	s.Structures = append(s.Structures[:i], s.Structures[i+1:]...)
	if len(s.Structures) == 0 {
		s.Structures = nil
	}
}

// removePileAt deletes the pile on (x,y) and its contents outright (the explicit
// destruction a remove pile miracle names). Preserves survivor creation order.
func (s *State) removePileAt(x, y int) {
	out := s.Piles[:0]
	for _, q := range s.Piles {
		if !(q.X == x && q.Y == y) {
			out = append(out, q)
		}
	}
	s.Piles = out
	if len(s.Piles) == 0 {
		s.Piles = nil
	}
}

// movePile relocates the pile on (fromX,fromY) to (toX,toY), merging onto any
// pile already there (one-pile-per-tile doctrine). Food batches keep their own
// SpoilAt; spears keep their durabilities, sorted ascending. Contents are copied
// by value before the source slice is mutated, so the merge is pointer-safe even
// when the destination append reallocates the pile slice.
func (s *State) movePile(fromX, fromY, toX, toY int) {
	src := s.pileAt(fromX, fromY)
	if src == nil {
		return
	}
	moved := *src
	s.removePileAt(fromX, fromY)
	dest := s.pileFor(toX, toY)
	dest.addNonFood("wood", moved.Wood)
	dest.addNonFood("stone", moved.Stone)
	dest.addNonFood("water", moved.Water)
	dest.addNonFood("planks", moved.Planks)
	dest.addNonFood("refined_stone", moved.RefinedStone)
	for _, b := range moved.Food {
		dest.addFood(b.Kind, b.N, b.SpoilAt)
	}
	if len(moved.Spears) > 0 {
		dest.Spears = append(dest.Spears, moved.Spears...)
		sort.Ints(dest.Spears)
	}
	if len(moved.Axes) > 0 {
		// Spec 032 US2: axes ride the pile move with their durabilities, sorted.
		dest.Axes = append(dest.Axes, moved.Axes...)
		sort.Ints(dest.Axes)
	}
}

// spillInventory pours an inventory (a removed chest's Store) onto the ground
// pile at (x,y), create-or-merge, food stamped with a fresh rot deadline — the
// exact death-spill vocabulary (state.go agent.died), so a removed chest can
// never silently destroy goods (spec 016 R4).
func (s *State) spillInventory(x, y int, inv *Inventory, tick int64) {
	if bulk(*inv) == 0 {
		return
	}
	pile := s.pileFor(x, y)
	pile.addNonFood("wood", inv.Wood)
	pile.addNonFood("stone", inv.Stone)
	pile.addNonFood("water", inv.Water)
	pile.addNonFood("planks", inv.Planks)
	pile.addNonFood("refined_stone", inv.RefinedStone)
	pile.addFood("food_raw", inv.FoodRaw, tick+rotWindowTicks)
	pile.addFood("food_cooked", inv.FoodCooked, tick+rotWindowTicks)
	pile.addFood("meals", inv.Meals, tick+rotWindowTicks)
	if len(inv.Spears) > 0 {
		pile.Spears = append(pile.Spears, inv.Spears...)
		sort.Ints(pile.Spears)
	}
	if len(inv.Axes) > 0 {
		// Spec 032 US2: axes spill with their durabilities, like spears.
		pile.Axes = append(pile.Axes, inv.Axes...)
		sort.Ints(pile.Axes)
	}
}

// VillagerAt returns the index of the first living villager standing on (x,y),
// or -1 when none does. Villagers may share a tile, so "first by agent index"
// is the deterministic choice; the miracle move arm and the perception-memory
// builder both resolve through this one helper so they can never disagree on
// which villager a tile-addressed move refers to. Map-free.
func (s *State) VillagerAt(x, y int) int {
	for i := range s.Agents {
		if !s.Agents[i].Dead && s.Agents[i].X == x && s.Agents[i].Y == y {
			return i
		}
	}
	return -1
}

// HasStructureAt reports whether a structure stands on (x,y). At most one
// structure ever occupies a tile (the buildSite invariant structureIndexAt
// documents), so presence is the tile's whole structure story. Exported for
// the bundle effect compiler's class+tile target resolution (spec 082): the
// compiler probes presence through the SAME one-per-tile lookup the miracle
// reducer arms use, so a tile-addressed bundle effect and the reducer can
// never disagree about which structure a tile names. Read-only,
// deterministic, map-free — the VillagerAt discipline.
func (s *State) HasStructureAt(x, y int) bool { return s.structureIndexAt(x, y) >= 0 }

// HasPileAt reports whether a ground pile sits on (x,y) — one pile per tile
// (the reducer invariant pileAt documents). Exported for the bundle effect
// compiler's class+tile target resolution (spec 082), mirroring
// HasStructureAt. Read-only, deterministic, map-free.
func (s *State) HasPileAt(x, y int) bool { return s.pileAt(x, y) != nil }

// AgentIndexByName resolves a villager name (case-insensitive, trimmed) to its
// roster index, or -1 when no villager bears it. Exported for the IPC miracle
// door, which receives a give_item's target by NAME (contracts §2); it resolves
// against the same AgentNames roster the guardian package's own resolver walks,
// so both doors turn a name into the same index and cannot drift. Map-free.
func AgentIndexByName(name string) int {
	name = strings.ToLower(strings.TrimSpace(name))
	for i, n := range AgentNames {
		if strings.ToLower(n) == name {
			return i
		}
	}
	return -1
}

// LivingAgents returns the indices of every living villager, ascending — the
// recipients of a time-snap perception memory (every villager feels the lurch,
// data-model.md). Map-free.
func (s *State) LivingAgents() []int {
	var out []int
	for i := range s.Agents {
		if !s.Agents[i].Dead {
			out = append(out, i)
		}
	}
	return out
}
