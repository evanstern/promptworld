package sim

import (
	"encoding/json"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// The executor: the deterministic layer that runs agent bodies unattended
// between planner calls (TASK-5). stepEvents is a pure function of
// (state, map, next tick) — it must not mutate s; the loop applies its
// returned events through the reducer.

const (
	nightStartSecond = 22 * 3600 // 22:00
	dayStartSecond   = 6 * 3600  // 06:00
)

func stepEvents(s *State, m *worldmap.Map, nextTick int64) []store.Event {
	// Run-end guard (spec 044 FR-002): an ended world emits nothing, ever.
	// stepEvents is the only sim emitter, so this single latch freezes
	// simulated time while the loop keeps serving reads (contracts/events.md
	// ordering guarantee 2).
	if s.Ended {
		return nil
	}
	var events []store.Event
	emit := func(typ string, payload any) {
		events = append(events, store.Event{Tick: nextTick, Type: typ, Payload: mustPayload(payload)})
	}

	// Day/night boundaries.
	day, _, _, _ := clock.GameTime(nextTick)
	night := s.Night
	switch clock.SecondOfDay(nextTick) {
	case nightStartSecond:
		emit("sim.night_started", DayPayload{Day: day})
		night = true
	case dayStartSecond:
		emit("sim.day_started", DayPayload{Day: day})
		night = false
		for i := range s.Agents {
			a := &s.Agents[i]
			if !a.Dead && a.Needs.Warmth < coldNightBelow {
				events = append(events, situatedMemoryEvent(nextTick, i, salColdNight,
					PlaceAt(s, a.X, a.Y), "", OriginAction, "Survived a freezing night in the open."))
			}
		}
	}

	// Guardian charge regeneration (TASK-12; spec 085 FR-004 rewrote the
	// cadence): absolute boundaries of the faith-band cadence — a pure
	// function of (faith score, scenario presence, tick), both inputs
	// event-sourced/boot-frozen, so replay is inherited. Same check shape,
	// same event, same empty payload as the fixed-6h original; a world that
	// has never folded a faith event sits in the steady band and fires on a
	// byte-identical schedule. Cadence 0 (the scenario forsaken band — the
	// FR-005 spiral posture, see FaithRegenCadenceTicks) means the check
	// never fires.
	if c := FaithRegenCadenceTicks(s.FaithScore(), s.scenario != nil); c > 0 &&
		nextTick%c == 0 && s.GuardianCharges < GuardianChargeCap {
		emit("guardian.charge_regenerated", ChargeRegeneratedPayload{})
	}

	// Guardian standing-order expiry (spec 029): a pure function of (state, tick),
	// exactly the charge_regenerated pattern — an active order whose TTL has
	// elapsed emits guardian.order_expired, which the reducer transitions to
	// expired (freeing the player-cap slot). Emitted once: the same event marks
	// the order non-active, so the next tick no longer sees it active. Replay
	// reproduces it deterministically without the guardian running — unlike a
	// trigger, which is a live-only injection (a matched condition, never replay).
	for i := range s.GuardianOrders {
		// A survival watch (spec 059) never expires — it is the guardian's standing
		// nature, not a timed order — so the expiry sweep skips it (origin-keyed
		// TTL exemption, matching the reducer's order_placed arm).
		if o := &s.GuardianOrders[i]; o.Status == "active" && o.Survival == "" && nextTick >= o.ExpiresTick {
			emit("guardian.order_expired", OrderIDPayload{ID: o.ID})
		}
	}

	// Designation-fulfillment sweep (spec 084 FR-004, research R14): an active
	// designation whose structural predicate holds emits designation.fulfilled
	// — a pure function of (state, tick), the charge_regenerated/order_expired
	// idiom. Emitted once: the same event flips the designation non-active, so
	// the next tick's sweep skips it. Replay reproduces it deterministically
	// with no guardian running. Sweep order within the tick is fixed:
	// designations first (slice order), then directives — directive
	// fulfillment reads designation STATUS from pre-tick state, so a
	// designation fulfilled at tick T yields its bound directives'
	// directive.fulfilled at T+1's sweep (the documented one-tick lag), never
	// an order-dependent same-tick race.
	for i := range s.Designations {
		if d := &s.Designations[i]; d.Status == "active" && designationFulfilled(s, d) {
			emit("designation.fulfilled", OrderIDPayload{ID: d.ID})
		}
	}

	// Directive fulfillment/expiry sweep (spec 084 FR-009): per active
	// directive, fulfillment is checked BEFORE expiry so a directive eligible
	// for both at one boundary lands exactly ONE terminal (fulfilled wins —
	// the work was done; the reducer would refuse a second terminal in the
	// same batch). Fulfilled: the bound designation's status is "fulfilled".
	// Expired: the TTL elapsed OR no targeted villager remains alive (the
	// un-executable clause — a pure state check, no TTL wait). Both pure over
	// (state, tick), each fired once by the same flips-non-active argument.
	for i := range s.Directives {
		d := &s.Directives[i]
		if d.Status != "active" {
			continue
		}
		if dsg := s.designationByID(d.DesignationID); dsg != nil && dsg.Status == "fulfilled" {
			emit("directive.fulfilled", DirectiveFulfilledPayload{
				ID: d.ID, DesignationID: d.DesignationID, Targets: Refs(d.Targets), IssuedTick: d.IssuedTick})
			continue
		}
		if nextTick >= d.ExpiresTick || directiveTargetsAllDead(s, d) {
			emit("directive.expired", OrderIDPayload{ID: d.ID})
		}
	}

	// Prophecy verification sweep (spec 085 FR-008): active prophecies judged
	// against their recorded claims — fulfil before fail, once, pure over
	// (pre-tick state, tick), with companion OriginReport memories riding the
	// batch (prophecy.go). Placed with the other guardian sweeps, after the
	// directive sweep at a FIXED position: a designation fulfilling at T flips
	// status at apply, so a dependent designation_fulfilled claim reads it
	// directly at T+1's sweep — no extra lag rule needed, only this fixed slot.
	events = append(events, prophecyEvents(s, nextTick)...)

	// Scenario incidents (spec 054 US2): due authored emissions from the
	// boot-frozen incident source (scenario.go) — the executor emission
	// class, exactly the charge-regen idiom above: pure over (state, config,
	// tick), reducer-valid shapes only. Emitted BEFORE gruStep so a
	// scheduled gru.emerged precedes the predator's own turn in the batch;
	// the roll-preemption check inside gruStep keeps the dice out of a
	// scheduled night entirely (research R3). Ambient worlds (nil scenario)
	// never enter this branch — byte-identical behavior (contract §1.3).
	if s.scenario != nil {
		events = append(events, scenarioIncidentEvents(s, m, nextTick)...)
	}

	// The gru: nightly emergence, stalking, wounds, dawn withdrawal (gru.go).
	events = append(events, gruStep(s, m, night, nextTick)...)

	// The stranger (spec 077 US2): store-seeking movement, bounded takes,
	// dawn departure (stranger.go). Ordered after gruStep and before the
	// governance/social beats (pinned by TestStrangerStepsAfterGru); a world
	// where no stranger ever arrived returns nil on the first check —
	// byte-identical pre-077 behavior (FR-017).
	events = append(events, strangerStep(s, m, night, nextTick)...)

	// Governance (TASK-13): the meeting lifecycle (only once a convention is
	// established — TASK-36) and the per-minute violation detectors
	// (governance.go).
	events = append(events, governanceEvents(s, m, nextTick)...)

	// Forage regrowth.
	for _, h := range s.Harvested {
		if h.Regrow == nextTick {
			emit("sim.forage_regrown", RegrownPayload{X: h.X, Y: h.Y})
		}
	}

	// Fire fuel burnout (T019): a fire whose deadline falls in this tick's
	// window goes cold — emit sim.fire_burned_out exactly once on the
	// transition (tick-1 < FuelUntil <= tick). Pure function of (state, tick);
	// lit-ness stays derived from FuelUntil, so the event carries no state
	// effect. Refuel pushes FuelUntil forward, re-arming this detection.
	for _, st := range s.Structures {
		if st.Kind == "fire" && st.FuelUntil > nextTick-1 && st.FuelUntil <= nextTick {
			emit("sim.fire_burned_out", FireBurnedOutPayload{X: st.X, Y: st.Y})
			// Deferred Phase-4 item (contracts/events.md): a fire going cold
			// nearby is background texture, not formative — low salience,
			// purely personal (no gossip subject), same witness-radius idiom
			// as the oven-built/death witnessing above. Fixed agent
			// iteration order keeps this deterministic.
			for w := range s.Agents {
				if s.Agents[w].Dead {
					continue
				}
				if abs(s.Agents[w].X-st.X)+abs(s.Agents[w].Y-st.Y) <= witnessRadius {
					events = append(events, situatedMemoryEvent(nextTick, w, salFireOut,
						PlaceAt(s, s.Agents[w].X, s.Agents[w].Y), "", OriginAction, "Watched the fire burn out."))
				}
			}
		}
	}

	// Per-game-minute needs heartbeat: decay, warmth, death.
	if nextTick%60 == 0 {
		// Neglect detector (spec 083, FR-004): per living AWAKE agent, per
		// survival need in the fixed food→warmth→rest order, fire ONE
		// sim.neglect_detected + its companion percept memory when the need
		// has sat below its spec-062 danger band for neglectWindowTicks with
		// zero class intents over the same window (NeglectDue — pure over
		// PRE-tick state, the recoveryHoldEvents purity precedent). Sleeping
		// agents are skipped at the beat (their inaction is sleep; the
		// spec-064 wake ladder owns sleeping emergencies) — anchors keep
		// accruing, so a still-critical waker fires on its next heartbeat.
		// Emitted BEFORE this beat's needs_changed events on purpose: the
		// batch then folds latch-then-needs, so a need that recovers on this
		// very beat closes its episode cleanly (Since=0, Fired=false) instead
		// of stranding a stale latch that would silence the NEXT episode.
		// The companion memory (salNeglect = 9 = GenerationBumpSalience)
		// bumps the agent's generation — the interrupt IS the research §3
		// planner beat; the per-episode fired latch bounds the rate.
		for i := range s.Agents {
			a := &s.Agents[i]
			if a.Dead || a.Asleep {
				continue
			}
			for _, need := range neglectNeedOrder {
				if !NeglectDue(a, need, nextTick) {
					continue
				}
				emit("sim.neglect_detected", NeglectDetectedPayload{
					Agent: Ref(i), Need: need, Level: needValue(a.Needs, need), Since: a.Neglect.Since(need),
				})
				events = append(events, situatedMemoryEvent(nextTick, i, salNeglect,
					PlaceAt(s, a.X, a.Y), "", OriginWitness, "%s", neglectMemoryText(need)))
			}
		}
		for i := range s.Agents {
			a := &s.Agents[i]
			if a.Dead {
				continue
			}
			n := decayNeeds(a.Needs, a.Asleep, night, warmAt(s, a.X, a.Y, nextTick), s.structureAt("shelter", a.X, a.Y),
				coldSnapActive(s, nextTick))
			emit("agent.needs_changed", NeedsPayload{
				Agent: Ref(i), Health: n.Health, Food: n.Food, Rest: n.Rest, Warmth: n.Warmth, Morale: n.Morale,
			})
			// Own near-death is a formative memory, once per collapse (latch).
			if n.Health < nearDeathBelow && !a.NearDeath && n.Health > 0 {
				cause := "cold and hunger"
				switch {
				case n.Food == 0 && n.Warmth > 0:
					cause = "hunger"
				case n.Warmth == 0 && n.Food > 0:
					cause = "the cold"
				}
				if s.Gru != nil && s.Gru.LastVictim == i && nextTick-s.Gru.LastAttack <= 3600 {
					cause = "the gru"
				}
				events = append(events, situatedMemoryEvent(nextTick, i, salNearDeath, PlaceAt(s, a.X, a.Y), "", OriginAction, "Nearly died — %s almost took me.", cause))
			}
			if n.Health == 0 {
				cause := "collapse"
				switch {
				case n.Food == 0:
					cause = "starvation"
				case n.Warmth == 0:
					cause = "exposure"
				}
				emit("agent.died", DiedPayload{Agent: Ref(i), Cause: cause})
				// Death marks every witness close enough to see it.
				for w := range s.Agents {
					if w == i || s.Agents[w].Dead {
						continue
					}
					if abs(s.Agents[w].X-a.X)+abs(s.Agents[w].Y-a.Y) <= witnessRadius {
						events = append(events, situatedMemoryAboutEvent(nextTick, w, i, -80, salWitnessDeath,
							PlaceAt(s, s.Agents[w].X, s.Agents[w].Y), OriginWitness, "Watched %s die of %s.", a.Name, cause))
					}
				}
			}
		}
	}

	// Ground-food rot sweep (T032, US5): on the same per-game-minute boundary
	// the needs heartbeat uses, each ground pile's food batches whose absolute
	// deadline has arrived (SpoilAt <= tick) are removed as a visible,
	// event-sourced happening. A pure function of (state, tick) — the fuel-sweep
	// pattern: no bookkeeping state, the reducer's batch removal itself re-arms
	// the sweep. Pile iteration is State.Piles slice order; per pile, same-kind
	// spoiled batches merge into ONE sim.food_rotted, and kinds are walked in the
	// fixed canonical food order — both fixed orders keep replay byte-identical.
	// A world with no piles emits nothing (degraded-mode safe); chest food
	// carries no deadlines and never reaches here (FR-010). Placed among the
	// world sweeps (regrowth, burnout, needs) before per-agent execution, so a
	// pickup completing this same tick lands after the rot events in the batch
	// and the reducer's clamp resolves the contest (spec edge "Rot mid-pickup").
	if nextTick%60 == 0 {
		for pi := range s.Piles {
			pile := &s.Piles[pi]
			for _, kind := range foodKinds {
				n := 0
				for _, b := range pile.Food {
					if b.Kind == kind && b.SpoilAt <= nextTick {
						n += b.N
					}
				}
				if n > 0 {
					emit("sim.food_rotted", FoodRottedPayload{X: pile.X, Y: pile.Y, Kind: kind, N: n})
				}
			}
		}
	}

	// Hail sweep (TASK-47): resolve outstanding pauses — met (hailer arrived)
	// or expired — before anyone moves this tick, so met-vs-expired is a
	// deterministic race with met winning ties (research D4).
	events = append(events, hailStep(s, nextTick)...)

	// Perception sweep (spec 041, T007): each awake villager diffs ground
	// truth within the witness radius against its mental map and emits
	// agent.saw for the new/changed facts (perceptionEvents below).
	events = append(events, perceptionEvents(s, m, nextTick)...)

	// Per-agent execution. Uses current state s (pre-tick); all effects
	// land as events.
	for i := range s.Agents {
		a := &s.Agents[i]
		if a.Dead {
			continue
		}

		if a.Asleep {
			if wakeReason(s, m, i, night, nextTick) {
				emit("agent.woke", AgentPayload{Agent: Ref(i)})
			}
			continue
		}

		// Meeting pinning (TASK-13): while the village convenes, attendees
		// drop what they're doing and head for the meeting place. The goal
		// is executor-set only — never planner-choosable — and stale pins
		// clear once the meeting ends.
		if meetingActive(s) && s.MeetingPlace != nil && attendCandidate(s, i) &&
			(a.Intent == nil || a.Intent.Goal != "attend_meeting") {
			emit("agent.intent_set", IntentSetPayload{
				Agent: Ref(i), Goal: "attend_meeting",
				TargetX: s.MeetingPlace.X, TargetY: s.MeetingPlace.Y,
				Source: "meeting",
			})
			continue
		}
		if !meetingActive(s) && a.Intent != nil && a.Intent.Goal == "attend_meeting" {
			emit("agent.intent_done", AgentPayload{Agent: Ref(i)})
			continue
		}

		// Hail pause (TASK-47): a flagged-down agent stands still — no reflex,
		// no plan-step evaluation, no stepping en route — until the window
		// lifts. Its needs still decay (heartbeat above), it still takes part
		// in social beats, and stationary work at a tile it already stands on
		// continues; intent and plan are left exactly as they were (FR-004).
		paused := hailPaused(a, nextTick)

		if a.Intent == nil {
			if paused {
				continue
			}
			// Guarded plan steps (TASK-32 US4) own an idle agent while the
			// head step's window is open: holding emits nothing, firing
			// sets the intent, expiry clears the plan — all deterministic,
			// no model at firing time (FR-017).
			if len(a.Plan) > 0 {
				events = append(events, planStepEvents(s, m, i, nextTick)...)
				continue
			}
			// The reflex is the fallback mind (TASK-7): it acts only on
			// agents idle past the grace window, leaving room for planner
			// injections; with no planner it remains the permanent
			// degraded mode. Staggered so agents don't all think at once.
			if nextTick-a.IdleSince >= reflexGraceTicks && (nextTick+int64(i)*7)%20 == 0 {
				d := decideIntent(s, m, i, nextTick)
				switch {
				case d.directEvent == "agent.ate":
					if p, ok := eatOutcome(a); ok {
						p.Agent = Ref(i)
						emit("agent.ate", p)
					}
				case d.intent != nil:
					emit("agent.intent_set", IntentSetPayload{
						Agent: Ref(i), Goal: d.intent.Goal,
						TargetX: d.intent.TargetX, TargetY: d.intent.TargetY,
						ResX: d.intent.ResX, ResY: d.intent.ResY,
						Source: "reflex",
						// Spec 064 R3: the reflex warmth rungs (day + night) issue
						// goto_warmth WITH a completion condition — carry it so the
						// intent holds at the fire instead of arrive-idle-wander.
						// Zero on every other reflex intent (arrive-and-done, unchanged).
						UntilNeed: d.intent.UntilNeed, UntilValue: d.intent.UntilValue,
					})
				}
			}
			continue
		}

		in := a.Intent
		if a.X == in.TargetX && a.Y == in.TargetY {
			events = append(events, executeAtTarget(s, m, i, nextTick)...)
			continue
		}
		if paused {
			continue // frozen in place: no stepping toward the target
		}

		// En route: one tile per moveEveryTicks, staggered like decisions. Spec 032
		// US3 (research R3): a second stateless cadence slot at phase 2 fires only
		// when the agent is standing ON a path tile, so stepping FROM a path moves
		// at exactly 2x (two steps per 5-tick window) — the tile stepped from
		// decides, matching the spec. Phase 0 always steps (baseline speed intact);
		// off-path agents never see phase 2, so no existing behavior changes.
		phase := (nextTick + int64(i)*3) % moveEveryTicks
		if phase == 0 || (phase == 2 && pathAt(s, a.X, a.Y)) {
			nx, ny := nextStep(m, s, a.X, a.Y, in.TargetX, in.TargetY)
			if nx == a.X && ny == a.Y {
				emit("agent.intent_done", AgentPayload{Agent: Ref(i)}) // unreachable
				continue
			}
			emit("agent.moved", AgentMovedPayload{Agent: Ref(i), X: nx, Y: ny})
			// Spec 097 (D1): ONLY the step that lands the walker ON its
			// intent's chosen destination is the intent-completing arrival
			// — the one moment the grounded observation channel fires
			// (observe.go). Never per wander step (the guard below is the
			// card's flood worry made code — a mid-walk tile observes
			// nothing), never for a zero-distance intent (no walk, no fresh
			// arrival — the ambient perception sweep covers standing
			// still). Emitted uniformly for every goal, reflex and planner
			// alike: reason interpretation stays out of the executor (no
			// LLM in the emission path); the mind weighs relevance (D3,
			// internal/mind/reconcile.go).
			if nx == in.TargetX && ny == in.TargetY {
				events = append(events, placeObservedEvents(s, m, i, nx, ny, nextTick)...)
			}
		}
	}

	// Adjacent idle agents: give/repay first (debts bind), then talk with a
	// verbatim rumor fallback (the social fabric's model-free floor).
	if nextTick%60 == 30 {
		events = append(events, socialEvents(s, nextTick)...)
	}

	// Hourly ledger due-check: overdue open debts break, permanently.
	if nextTick%3600 == 0 {
		repayNorm := activeNormOfKind(s, NormRepayDebts)
		for _, d := range s.Debts {
			if d.Status == "open" && nextTick > d.Due {
				events = append(events,
					store.Event{Tick: nextTick, Type: "social.promise_broken",
						Payload: mustPayload(PromiseBrokenPayload{ID: d.ID})},
					store.Event{Tick: nextTick, Type: "social.relation_changed",
						Payload: mustPayload(RelationChangedPayload{
							A: Ref(d.Creditor), B: Ref(d.Debtor),
							TrustDelta: brokenTrustPenalty, AffectionDelta: brokenAffectPenalty,
							Reason: "promise broken"})},
					situatedMemoryAboutEvent(nextTick, d.Creditor, d.Debtor, toneNeverPaid, salNeverPaid,
						PlaceAt(s, s.Agents[d.Creditor].X, s.Agents[d.Creditor].Y),
						OriginWitness, "%s never repaid the food I gave them.", s.Agents[d.Debtor].Name))
				// A repay-debts norm in force makes the broken promise a
				// witnessed crime too (TASK-13).
				if repayNorm != nil {
					events = append(events, violationEvents(s, repayNorm, d.Debtor, nextTick)...)
				}
			}
		}
	}

	// Faith accounting sweep (spec 085 FR-003): scan THIS tick's batch for the
	// five faith-source events (directive fulfilled/expired, deaths, prophecy
	// terminals) and emit one faith.changed per source in batch order — pure
	// over (pre-tick state, this batch, tick), the run-end detector's own
	// idiom below. Positioned AFTER every faith-source emitter (the directive
	// and prophecy sweeps, gruStep, the needs heartbeat) and BEFORE the
	// scenario rubric and run-end detection, so every source's faith companion
	// lands in the same batch and run.ended stays the batch's last event.
	events = append(events, faithEvents(s, events, nextTick)...)

	// Scenario rubric (spec 054 US1): the pass boundary evaluation, pure over
	// (pre-tick state, boot-frozen definition, next tick) plus THIS batch —
	// the run-end detector's own idiom below: same-tick deaths are not yet
	// folded into s, and an all-dead dawn must be a fail, not a photo-finish
	// pass (spec edge case). Placed after every emitter and before run-end
	// detection so deaths precede the evaluation and run.ended stays the
	// batch's last event. Emits exercise_passed (+ same-batch stage_unlocked,
	// pass first) exactly once, via state latches — see scenario.go.
	if s.scenario != nil {
		events = append(events, scenarioRubricEvents(s, nextTick, events)...)
	}

	// Run-end detection (spec 044 R1): a pure function of (pre-tick state,
	// this batch). When every villager still living at the tick's start died
	// within the batch, declare the run over — as the batch's LAST event, so
	// every same-tick agent.died (heartbeat or gru) and its witness memories
	// precede it and no sim event trails it: the per-agent loops above still
	// act for agents whose death is only in this batch (pre-tick state shows
	// them alive), so emitting mid-batch would let their events follow the
	// declaration (contracts/events.md ordering). The !s.Ended check is the
	// exactly-once latch belt to the top-of-function guard's braces. The
	// payload carries the whole run's deaths — the State.Deaths ledger plus
	// this batch's — so no consumer ever scans the log for them.
	if !s.Ended {
		var batch []DeathRecord
		for _, e := range events {
			if e.Type != "agent.died" {
				continue
			}
			var p DiedPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				continue // struct-built above; cannot fail
			}
			batch = append(batch, DeathRecord{Agent: p.Agent.ID, Tick: e.Tick, Cause: p.Cause})
		}
		if len(batch) > 0 && len(batch) == livingCount(s) {
			deaths := make([]DeathRecord, 0, len(s.Deaths)+len(batch))
			deaths = append(deaths, s.Deaths...)
			deaths = append(deaths, batch...)
			emit("run.ended", RunEndedPayload{
				Tick:       nextTick,
				Deaths:     DeathRefs(deaths),
				FinalCause: deaths[len(deaths)-1].Cause,
			})
		}
	}

	return events
}

// perceptionEvents is the spec-041 perception sweep (T007, research D2): for
// each awake living villager on its perception beat, diff ground truth within
// the witness radius against the agent's mental map and emit ONE agent.saw
// carrying the new/changed facts, fully baked (Seen = this tick, provenance
// witnessed, Detail as perceived) and sorted (Kind, X, Y). A remembered fact
// is skipped only when it is witnessed, detail-current, AND still fresh — so
// re-perception refreshes a fact exactly when its read-time horizon would
// stale it, and a told/revealed fact upgrades to witnessed on first sight.
// Pure function of (state, map, nextTick); stepEvents doctrine: reads s,
// never mutates it.
//
// The perception beat is the movement cadence — the same per-agent stagger
// slot as stepping ((tick + i*3) % moveEveryTicks == 0) — so a walker looks
// around exactly as often as it moves and a stationary agent notices changes
// within the same window, at a fifth of the per-tick sweep cost.
//
// Gated kinds (data-model.md): every structure kind as-is (fires bake their
// FuelUntil into Detail), ground piles ("pile"), and the resource tiles —
// standing trees, unharvested forage, unquarried rock ("tree"/"forage"/
// "rock", overlay-aware), water shoreline ("water_edge": a water tile with a
// statically-passable neighbor — terrain shape, so walls never churn it),
// and dens ("den"; a den on cooldown still exists).
func perceptionEvents(s *State, m *worldmap.Map, nextTick int64) []store.Event {
	var events []store.Event
	// Beat eligibility first (T034 hot-path finding): the overlay sets below
	// are worth building only when some agent actually looks around this
	// tick — at night the whole village sleeps and the sweep must be free.
	onBeat := false
	for i := range s.Agents {
		a := &s.Agents[i]
		if !a.Dead && !a.Asleep && a.Map != nil && (nextTick+int64(i)*3)%moveEveryTicks == 0 {
			onBeat = true
			break
		}
	}
	if !onBeat {
		return nil
	}

	for i := range s.Agents {
		a := &s.Agents[i]
		if a.Dead || a.Asleep || a.Map == nil {
			continue
		}
		if (nextTick+int64(i)*3)%moveEveryTicks != 0 {
			continue // not this agent's perception beat
		}
		// Per-beat local membership sets (T034 hot-path finding): the whole-
		// world overlay sets this sweep used to build EVERY tick were steady
		// allocation churn (O(overlays) map inserts per tick, forever
		// growing). Each beat instead scans the overlay slices once — no
		// allocation for the misses — and keeps only the handful of points
		// inside this agent's diamond. Lookup-only, so map iteration order
		// never matters.
		cleared := make(map[Point]bool)
		for _, p := range s.Cleared {
			if abs(p.X-a.X)+abs(p.Y-a.Y) <= witnessRadius {
				cleared[p] = true
			}
		}
		harvested := make(map[Point]bool)
		for _, h := range s.Harvested {
			if abs(h.X-a.X)+abs(h.Y-a.Y) <= witnessRadius {
				harvested[Point{X: h.X, Y: h.Y}] = true
			}
		}
		quarried := make(map[Point]bool)
		for _, p := range s.Quarried {
			if abs(p.X-a.X)+abs(p.Y-a.Y) <= witnessRadius {
				quarried[p] = true
			}
		}
		var news []PlaceFact
		note := func(kind string, x, y int, detail int64) {
			if rem, ok := a.Map.factAt(kind, x, y); ok &&
				rem.Provenance == ProvenanceWitnessed && rem.Detail == detail &&
				factFresh(rem, nextTick) {
				return // known, current, and fresh — nothing new to record
			}
			news = append(news, PlaceFact{
				Kind: kind, X: x, Y: y, Seen: nextTick,
				Provenance: ProvenanceWitnessed, Detail: detail,
			})
		}
		// Resource tiles within the Manhattan radius (row-major scan; the
		// payload is sorted below regardless).
		for y := a.Y - witnessRadius; y <= a.Y+witnessRadius; y++ {
			for x := a.X - witnessRadius; x <= a.X+witnessRadius; x++ {
				if !m.InBounds(x, y) || abs(x-a.X)+abs(y-a.Y) > witnessRadius {
					continue
				}
				pt := Point{X: x, Y: y}
				switch m.At(x, y) {
				case worldmap.Tree:
					if !cleared[pt] {
						note("tree", x, y, 0)
					}
				case worldmap.Forage:
					if !harvested[pt] {
						note("forage", x, y, 0)
					}
				case worldmap.Rock:
					if !quarried[pt] {
						note("rock", x, y, 0)
					}
				case worldmap.Water:
					if waterEdge(m, x, y) {
						note("water_edge", x, y, 0)
					}
					// Marsh and Sand (spec 068, C13) fall through deliberately:
					// they carry no resource affordances, so there is no
					// mental-map fact kind to record for them — like grass,
					// they are open ground, not a resource.
				}
			}
		}
		for _, d := range m.Dens {
			if abs(d.X-a.X)+abs(d.Y-a.Y) <= witnessRadius {
				note("den", d.X, d.Y, 0)
			}
		}
		for _, st := range s.Structures {
			if abs(st.X-a.X)+abs(st.Y-a.Y) <= witnessRadius {
				// FuelUntil is zero for every non-fire kind, so Detail stays
				// omitted for them (data-model.md: fires bake it, piles 0).
				note(st.Kind, st.X, st.Y, st.FuelUntil)
			}
		}
		for _, p := range s.Piles {
			if abs(p.X-a.X)+abs(p.Y-a.Y) <= witnessRadius {
				note("pile", p.X, p.Y, 0)
			}
		}
		if len(news) > 0 {
			sortFacts(news)
			events = append(events, store.Event{Tick: nextTick, Type: "agent.saw",
				Payload: mustPayload(SawPayload{Agent: Ref(i), Facts: news})})
		}

		// The correction half (spec 041 US3, T019): remembered FRESH facts
		// within the radius that are ABSENT from ground truth are perceived
		// gone. Absence is about the PLACE, not its availability — a
		// harvested forage spot or a cooling den still exists (only its
		// availability lapsed, the resolvers' ground-condition class), while
		// a chopped tree, a quarried-out outcrop, a drained pile, or a
		// removed structure is genuinely no more (groundFactPresent). Gone
		// facts ride verbatim (as remembered) in canonical order — Facts is
		// already (Kind,X,Y)-sorted. Stale facts are invisible to read paths
		// and left untouched; re-perception (agent.saw above) covers their
		// return. Emitted after the agent's saw event, a fixed order; the two
		// batches are disjoint by construction (a fact absent from ground
		// truth is never in the saw diff).
		var gone []PlaceFact
		clearedAt := func(x, y int) bool { return cleared[Point{X: x, Y: y}] }
		quarriedAt := func(x, y int) bool { return quarried[Point{X: x, Y: y}] }
		for _, f := range a.Map.Facts {
			if abs(f.X-a.X)+abs(f.Y-a.Y) > witnessRadius || !factFresh(f, nextTick) {
				continue
			}
			if !groundFactPresentIn(s, m, f, clearedAt, quarriedAt) {
				gone = append(gone, f)
			}
		}
		if len(gone) > 0 {
			events = append(events, store.Event{Tick: nextTick, Type: "agent.map_corrected",
				Payload: mustPayload(MapCorrectedPayload{Agent: Ref(i), Gone: gone})})
			// The situated discoveries ride the same batch as companion
			// memory events, one per gone fact (the buildFailedEvents shape —
			// memories accrete ONLY via agent.memory_added, TestMemoriesAccrete).
			// Salience sits below the generation bump: re-arming on a matching
			// intent target is the absorb trigger's job, not an interrupt.
			for _, f := range gone {
				events = append(events, situatedMemoryEvent(nextTick, i, salMapCorrected,
					PlaceAt(s, a.X, a.Y), "", OriginWitness, "%s", mapCorrectedText(f)))
			}
		}
	}
	return events
}

// groundFactPresent reports whether a remembered place-fact still names a
// real place (spec 041 US3): the correction's absence test. Kind-aware —
// terrain spots that merely regrow/cool (forage, dens, water) are permanent
// places and never correct; overlay-permanent removals (chopped tree,
// quarried rock), drained piles, and removed structures are gone.
func groundFactPresent(s *State, m *worldmap.Map, f PlaceFact) bool {
	return groundFactPresentIn(s, m, f,
		func(x, y int) bool {
			for _, p := range s.Cleared {
				if p.X == x && p.Y == y {
					return true
				}
			}
			return false
		},
		func(x, y int) bool {
			for _, p := range s.Quarried {
				if p.X == x && p.Y == y {
					return true
				}
			}
			return false
		})
}

// groundFactPresentIn is groundFactPresent with caller-supplied cleared /
// quarried membership tests (T034 hot-path finding): the perception sweep's
// correction half runs this once per remembered in-radius fact per beat, and
// the slice-scanning overlay checks (effectiveKind's shape) made it
// O(facts × overlays); the sweep passes its per-beat local sets instead. The
// two overlay tests must match the Cleared/Quarried semantics exactly —
// groundFactPresent above is the reference wiring.
func groundFactPresentIn(s *State, m *worldmap.Map, f PlaceFact, cleared, quarried func(x, y int) bool) bool {
	if !m.InBounds(f.X, f.Y) {
		return false
	}
	switch f.Kind {
	case "tree":
		// effectiveKind's overlay rule: a cleared tree tile reads Grass.
		return m.At(f.X, f.Y) == worldmap.Tree && !cleared(f.X, f.Y)
	case "rock":
		// effectiveKind's overlay rule: a quarried outcrop reads Depleted.
		return m.At(f.X, f.Y) == worldmap.Rock && !quarried(f.X, f.Y)
	case "forage":
		// Harvested spots regrow — the SPOT persists (static terrain).
		return m.At(f.X, f.Y) == worldmap.Forage
	case "water_edge":
		return m.At(f.X, f.Y) == worldmap.Water && waterEdge(m, f.X, f.Y)
	case "den":
		for _, d := range m.Dens {
			if d.X == f.X && d.Y == f.Y {
				return true
			}
		}
		return false
	case "pile":
		return s.pileAt(f.X, f.Y) != nil
	}
	return s.structureAt(f.Kind, f.X, f.Y)
}

// groundFactDetail is the kind-specific Detail scalar as ground truth holds it
// right now (spec 041 data-model: fires bake their FuelUntil; every other kind
// 0) — the guardian.place_revealed arm's stamp, mirroring what the perception
// sweep would bake had the agent seen the place itself.
func groundFactDetail(s *State, f PlaceFact) int64 {
	if f.Kind != "fire" {
		return 0
	}
	for i := range s.Structures {
		if st := &s.Structures[i]; st.Kind == "fire" && st.X == f.X && st.Y == f.Y {
			return st.FuelUntil
		}
	}
	return 0
}

// waterEdge reports whether a water tile touches statically-walkable ground —
// the shoreline a villager can draw from (and the only water worth a
// place-fact). Static terrain only: the shoreline is terrain shape, so
// dynamic overlays (walls) never churn the fact set.
func waterEdge(m *worldmap.Map, x, y int) bool {
	for _, d := range neighborOrder {
		if m.Passable(x+d[0], y+d[1]) {
			return true
		}
	}
	return false
}

// socialEvents runs the adjacency slot: repayment, gifts to the starving,
// or a talk (with the deterministic verbatim rumor fallback). One social
// beat per heartbeat keeps the fabric legible.
func socialEvents(s *State, nextTick int64) []store.Event {
	var events []store.Event
	give := func(from, to int) {
		f, t := &s.Agents[from], &s.Agents[to]
		events = append(events,
			store.Event{Tick: nextTick, Type: "social.gave",
				Payload: mustPayload(GavePayload{From: Ref(from), To: Ref(to), Kind: "food"})},
			store.Event{Tick: nextTick, Type: "social.relation_changed",
				Payload: mustPayload(RelationChangedPayload{
					A: Ref(to), B: Ref(from), TrustDelta: giveTrustToGiver, AffectionDelta: giveAffectionToGiver,
					Reason: "shared food"})},
			store.Event{Tick: nextTick, Type: "social.relation_changed",
				Payload: mustPayload(RelationChangedPayload{
					A: Ref(from), B: Ref(to), TrustDelta: 0, AffectionDelta: giveAffectionToRecv,
					Reason: "shared food"})},
			situatedMemoryAboutEvent(nextTick, to, from, toneSaved, salWasSaved,
				PlaceAt(s, t.X, t.Y), OriginWitness, "%s gave me food when I needed it.", f.Name),
			situatedMemoryEvent(nextTick, from, salGaveHelp, PlaceAt(s, f.X, f.Y), "", OriginAction, "Gave food to %s.", t.Name))
	}

	for i := range s.Agents {
		a := &s.Agents[i]
		if a.Dead || a.Asleep {
			continue
		}
		for j := i + 1; j < len(s.Agents); j++ {
			b := &s.Agents[j]
			if b.Dead || b.Asleep || abs(a.X-b.X)+abs(a.Y-b.Y) != 1 {
				continue
			}
			// 1) Repay an open debt when able.
			if deb, cred, ok := repayable(s, i, j, nextTick); ok {
				give(deb, cred)
				return events
			}
			// 2) Give to a starving neighbor.
			if giver, recv, ok := giveable(s, i, j, nextTick); ok {
				give(giver, recv)
				return events
			}
			// 3) Talk (+ verbatim rumor fallback). Villagers chat while
			// working — requiring mutual idleness starved the fabric once
			// planners kept everyone permanently tasked (cooldowns still
			// bound the chatter).
			if canTalk(a, nextTick) && canTalk(b, nextTick) {
				return append(events, talkEvents(s, i, j, nextTick)...)
			}
		}
	}
	return events
}

// talkEvents founds a talk between adjacent agents i and j: the morale/
// affection/memory shape plus the deterministic verbatim rumor floor (the
// better-stocked teller passes one rumor; the mind's conversations paraphrase
// instead when a model is available). Shared by the ambient social beat and
// the hail sweep — the sweep founds deliberately, bypassing the ambient
// cooldown (the caller here gates on canTalk; hailStep does not).
func talkEvents(s *State, i, j int, nextTick int64) []store.Event {
	a, b := &s.Agents[i], &s.Agents[j]
	events := []store.Event{
		{Tick: nextTick, Type: "agent.talked",
			Payload: mustPayload(TalkedPayload{A: Ref(i), B: Ref(j)})},
		{Tick: nextTick, Type: "social.relation_changed",
			Payload: mustPayload(RelationChangedPayload{
				A: Ref(i), B: Ref(j), AffectionDelta: talkAffection, Reason: "talked"})},
		{Tick: nextTick, Type: "social.relation_changed",
			Payload: mustPayload(RelationChangedPayload{
				A: Ref(j), B: Ref(i), AffectionDelta: talkAffection, Reason: "talked"})},
		situatedMemoryEvent(nextTick, i, salTalk, PlaceAt(s, a.X, a.Y), "", OriginAction, "Talked with %s.", b.Name),
		situatedMemoryEvent(nextTick, j, salTalk, PlaceAt(s, b.X, b.Y), "", OriginAction, "Talked with %s.", a.Name),
	}
	if tell, ok := TellableFor(s, i, j); ok {
		events = append(events, rumorTellEvent(nextTick, i, j, tell))
	} else if tell, ok := TellableFor(s, j, i); ok {
		events = append(events, rumorTellEvent(nextTick, j, i, tell))
	}
	// Spec 041 US5 (research D5): the place-knowledge exchange rides EVERY
	// founded talk beside the rumor slot — up to placeTellCap facts per
	// direction the other lacks or holds staler (tellablePlaces), one
	// social.place_told per direction with facts baked at emission, plus
	// companion situated memories both sides (the map_corrected shape:
	// memories accrete only via agent.memory_added). Direction order i→j
	// then j→i is fixed for determinism.
	for _, dir := range [2][2]int{{i, j}, {j, i}} {
		from, to := dir[0], dir[1]
		facts := tellablePlaces(s, from, to, nextTick)
		if len(facts) == 0 {
			continue
		}
		events = append(events,
			store.Event{Tick: nextTick, Type: "social.place_told",
				Payload: mustPayload(PlaceToldPayload{From: Ref(from), To: Ref(to), Facts: facts})},
			situatedMemoryEvent(nextTick, from, salPlaceTold,
				PlaceAt(s, s.Agents[from].X, s.Agents[from].Y), "", OriginAction,
				"%s", placeToldText(s.Agents[to].Name, facts, true)),
			situatedMemoryEvent(nextTick, to, salPlaceTold,
				PlaceAt(s, s.Agents[to].X, s.Agents[to].Y), "", OriginReport,
				"%s", placeToldText(s.Agents[from].Name, facts, false)),
		)
	}
	return events
}

func rumorTellEvent(tick int64, from, to int, tell Tellable) store.Event {
	return store.Event{Tick: tick, Type: "social.rumor_told",
		Payload: mustPayload(RumorToldPayload{
			From: Ref(from), To: Ref(to), RumorID: tell.RumorID, Subject: Ref(tell.Subject),
			Tone: tell.Tone, Text: tell.Text, Confidence: tell.Confidence,
		})}
}

// repayable: one of the pair owes the other and can spare a meal — and the
// creditor has bulk to receive it (T012: a gift into a full pouch is skipped
// under the cap, research R2; the debt simply stays open until there's room).
func repayable(s *State, i, j int, tick int64) (debtor, creditor int, ok bool) {
	for _, d := range s.Debts {
		if d.Status != "open" || d.Kind != "food" {
			continue
		}
		if d.Debtor == i && d.Creditor == j && canGive(&s.Agents[i], tick) && freeBulk(s.Agents[j].Inv) > 0 {
			return i, j, true
		}
		if d.Debtor == j && d.Creditor == i && canGive(&s.Agents[j], tick) && freeBulk(s.Agents[i].Inv) > 0 {
			return j, i, true
		}
	}
	return 0, 0, false
}

// giveable: one is starving, the other has spare food — and the starving
// receiver has free bulk (T012: never over the cap; a starving villager at the
// cap is carrying food already and would eat rather than receive).
func giveable(s *State, i, j int, tick int64) (giver, recv int, ok bool) {
	a, b := &s.Agents[i], &s.Agents[j]
	if a.Needs.Food < giveNeedBelow && canGive(b, tick) && freeBulk(a.Inv) > 0 {
		return j, i, true
	}
	if b.Needs.Food < giveNeedBelow && canGive(a, tick) && freeBulk(b.Inv) > 0 {
		return i, j, true
	}
	return 0, 0, false
}

func canGive(a *Agent, tick int64) bool {
	// Give-to-starving stays on raw food (T018 decision: simplest re-expression
	// of the pre-feature gift; the food triplet's least-nutritious form is what
	// a subsistence village shares — see social.go apply).
	return a.Inv.FoodRaw >= giveKeepsAtLeast &&
		(a.LastGive == 0 || tick-a.LastGive >= giveCooldownSec)
}

func canTalk(a *Agent, tick int64) bool {
	return a.LastTalk == 0 || tick-a.LastTalk >= talkCooldownSec
}

// recoveryHoldEvents runs the arrival/hold/complete/abort state machine for a
// needs-conditioned intent (spec 064 R2/R4) whose agent stands on its target.
// It is a PURE function of pre-tick state (s.Agents[i].Needs is the batch's
// start-of-tick value — the heartbeat only emits needs_changed, the reducer
// applies it after this batch), so replay reproduces every branch identically.
//
//   - need already at/over the threshold ⇒ complete now (agent.intent_done; the
//     already-satisfied-on-arrival edge and the mid-hold crossing are one check,
//     never assumed). The spec-062 yield window arms iff the ring source is
//     intelligence — automatic in the intent_done reducer — so a reflex-issued
//     recovery completing never arms it (source discipline, unchanged).
//   - first tick below threshold (WorkStart == 0) ⇒ anchor the hold: work_started
//     stamps the hold-since tick and captures the reference need level (Ref).
//   - a full recoveryStallTicks window elapsed since the anchor ⇒ judge progress:
//     net gain re-anchors (a live source keeps the hold alive); no net gain
//     aborts (agent.recovery_stalled) — the honest dead-source outcome.
//   - a higher-priority survival need in its danger band ⇒ yield (intent_done)
//     so the reflex's higher rung acts (US3 AS2 — survival preempts recovery).
//   - otherwise ⇒ hold silently (emit nothing); the agent stands recovering.
//
// No new preemption immunity (R7): a holding intent is an ordinary active intent
// — the caller's meeting pin, a planner intent_set override, staleness, the
// survival-preemption yield, and the abort are its only exits; the yield makes it
// LESS sticky than a moving intent, never more (no new immunity).
func recoveryHoldEvents(s *State, i int, nextTick int64) []store.Event {
	var events []store.Event
	emit := func(typ string, payload any) {
		events = append(events, store.Event{Tick: nextTick, Type: typ, Payload: mustPayload(payload)})
	}
	a := &s.Agents[i]
	in := a.Intent
	need := needValue(a.Needs, in.UntilNeed)

	if need >= in.UntilValue {
		emit("agent.intent_done", AgentPayload{Agent: Ref(i)})
		return events
	}
	// Survival preemption (spec 064 US3 AS2 / FR-004): a recovery for one need
	// never holds an agent THROUGH a worse emergency in another. When a
	// higher-priority survival need has crossed into its danger band, the hold
	// ends (intent_done) so the agent re-decides and the reflex's higher rung —
	// which runs BEFORE the warmth rung — owns the agent (forage/eat for hunger).
	// The OPPOSITE of immunity: it makes a hold LESS sticky, reusing the 062
	// danger-band doctrine (one home), so a hold is interruptible for a genuine
	// emergency exactly as the ladder interrupts any villager. Without it a
	// no-planner world's held villager can starve at a fire (the reflex only runs
	// on an idle agent). A reflex-sourced completion never arms the 062 window.
	if preemptsRecovery(a, in.UntilNeed) {
		emit("agent.intent_done", AgentPayload{Agent: Ref(i)})
		return events
	}
	if in.WorkStart == 0 {
		emit("agent.work_started", WorkStartedPayload{Agent: Ref(i), Tick: nextTick, Ref: need})
		return events
	}
	if nextTick-in.WorkStart >= recoveryStallTicks {
		if need > in.HoldRef {
			// Progress across the window — the source is delivering; slide the
			// anchor forward so a slow-but-steady recovery is never aborted.
			emit("agent.work_started", WorkStartedPayload{Agent: Ref(i), Tick: nextTick, Ref: need})
		} else {
			// No net gain across a whole window — the source is dead. Abort with
			// the distinct outcome; the agent re-decides (reflex or planner).
			emit("agent.recovery_stalled", RecoveryStalledPayload{Agent: Ref(i), Goal: in.Goal, Need: in.UntilNeed})
		}
	}
	return events
}

// executeAtTarget runs the arrival/work/completion state machine for the
// agent standing on its intent target.
func executeAtTarget(s *State, m *worldmap.Map, i int, nextTick int64) []store.Event {
	var events []store.Event
	emit := func(typ string, payload any) {
		events = append(events, store.Event{Tick: nextTick, Type: typ, Payload: mustPayload(payload)})
	}
	a := &s.Agents[i]
	in := a.Intent

	// Needs-conditioned recovery (spec 064 R2): an intent carrying a completion
	// condition HOLDS at its target and completes on the need crossing the
	// threshold — not the goal's default arrive-and-done. Intercepted BEFORE the
	// per-goal switch so it is goal-agnostic: it governs the planner's warm_up AND
	// the reflex's goto_warmth-with-condition AND any future conditioned verb, all
	// through one check. A conditionless intent (UntilNeed == "") skips this
	// entirely and behaves exactly as it did pre-064 (opt-in mechanism, FR-001).
	if in.UntilNeed != "" {
		return recoveryHoldEvents(s, i, nextTick)
	}

	// Instant goals complete on arrival.
	switch in.Goal {
	case "sleep":
		emit("agent.slept", AgentPayload{Agent: Ref(i)})
		return events
	case "wander", "goto_warmth", "seek", "search", "heed_directive":
		// search (spec 041 US4) is wander-class: instant on arrival — the
		// walk itself did the exploring (movement marks explored terrain and
		// the perception beat witnesses what's there). heed_directive (spec
		// 084 research R13) is the DIRECTIVE rung's walk-to-site leg, the
		// same completion shape: arrival IS the outcome, and the next idle
		// decision picks the work leg (or the planner does).
		emit("agent.intent_done", AgentPayload{Agent: Ref(i)})
		return events
	case "refuel_fire":
		// T020: instant on arrival (like eat/sleep). Re-validate at completion
		// (contested pattern): fire still present and wood still carried, else
		// resolve with no effect. The new deadline is absolute and capped; a
		// cold fire relights. At-cap refuels are a no-op (edge case: consumes
		// and extends nothing) — detected as no gain over the current deadline.
		st, ok := fireStructAt(s, in.TargetX, in.TargetY)
		if !ok || a.Inv.Wood < 1 {
			emit("agent.intent_done", AgentPayload{Agent: Ref(i)})
			return events
		}
		base := st.FuelUntil
		if base < nextTick {
			base = nextTick // cold or expired: relight from now
		}
		deadline := base + s.FireBurnPerWood()
		if capAt := nextTick + fireFuelCap; deadline > capAt {
			deadline = capAt
		}
		if deadline <= st.FuelUntil {
			emit("agent.intent_done", AgentPayload{Agent: Ref(i)}) // already at the fuel cap
			return events
		}
		emit("agent.refueled", RefueledPayload{Agent: Ref(i), X: in.TargetX, Y: in.TargetY, FuelUntil: deadline})
		return events
	case "attend_meeting":
		// Assembled: stand at the meeting place until it closes (the
		// executor clears the pin once the meeting ends).
		return events
	case "drop":
		// T016 (spec 013 US2): instant on the agent's current tile. Emit
		// agent.dropped with the ACTUAL post-clamp count — min(Qty-or-all,
		// carried). Kind is required; an empty Kind or nothing carried resolves
		// via intent_done only (no pile is touched, contested-resource pattern).
		n := carriedCount(a.Inv, in.Kind)
		if in.Qty > 0 && in.Qty < n {
			n = in.Qty
		}
		if in.Kind == "" || n <= 0 {
			emit("agent.intent_done", AgentPayload{Agent: Ref(i)})
			return events
		}
		emit("agent.dropped", DroppedPayload{Agent: Ref(i), X: a.X, Y: a.Y, Kind: in.Kind, N: n})
		return events
	case "pick_up":
		// T017 (spec 013 US2): instant on arrival. Re-validate a pile on/
		// adjacent (it may have been drained while walking over) and emit ONE
		// agent.picked_up per kind actually moved, truncated cumulatively to
		// free bulk. Kind "" sweeps every kind in canonical field order (the
		// reducer drains food oldest-batch-first). Nothing moved ⇒ intent_done.
		pile := s.pileOnOrAdjacent(a.X, a.Y)
		if pile == nil {
			emit("agent.intent_done", AgentPayload{Agent: Ref(i)})
			return events
		}
		kinds := []string{in.Kind}
		if in.Kind == "" {
			kinds = canonicalKinds
		}
		free := freeBulk(a.Inv)
		moved := false
		for _, kind := range kinds {
			if free <= 0 {
				break
			}
			take := pile.avail(kind)
			if in.Kind != "" && in.Qty > 0 && in.Qty < take {
				take = in.Qty
			}
			if take > free {
				take = free
			}
			if take <= 0 {
				continue
			}
			emit("agent.picked_up", PickedUpPayload{Agent: Ref(i), X: pile.X, Y: pile.Y, Kind: kind, N: take})
			free -= take
			moved = true
		}
		if !moved {
			emit("agent.intent_done", AgentPayload{Agent: Ref(i)})
		}
		return events
	case "deposit":
		// T024 (spec 013 US3): instant on arrival at the chest. Re-validate the
		// chest still stands (contested pattern) and truncate the move to its free
		// space (chestCap − bulk(*Store)). Kind is required. Spec 096: a vanished
		// chest or a full chest resolve LOUDLY (agent.intent_failed, contested —
		// something else changed the chest out from under the intent); an empty
		// Kind resolves LOUDLY too but distinctly (invalid — a malformed intent no
		// chest state could ever satisfy). The payload carries the ACTUAL
		// post-clamp count.
		ch := s.chestAt(in.TargetX, in.TargetY)
		if ch == nil || ch.Store == nil {
			return append(events, intentFailedEvents(s, i, a, in, intentFailContested, nextTick)...)
		}
		if in.Kind == "" {
			return append(events, intentFailedEvents(s, i, a, in, intentFailInvalid, nextTick)...)
		}
		n := carriedCount(a.Inv, in.Kind)
		if in.Qty > 0 && in.Qty < n {
			n = in.Qty
		}
		if free := chestCap - bulk(*ch.Store); n > free {
			n = free
		}
		if n <= 0 {
			return append(events, intentFailedEvents(s, i, a, in, intentFailContested, nextTick)...)
		}
		emit("agent.deposited", DepositedPayload{Agent: Ref(i), X: in.TargetX, Y: in.TargetY, Kind: in.Kind, N: n})
		return events
	case "withdraw":
		// T024: instant on arrival. Re-validate the chest, then emit ONE
		// agent.withdrew per kind actually moved, truncated cumulatively to the
		// taker's free bulk and to what the chest holds. A named Kind honors Qty;
		// Kind "" sweeps every kind in canonical field order. Owner rides the
		// payload (the theft companion batch is US4, T029 — not emitted here).
		// Spec 096: a vanished chest or nothing moved resolve LOUDLY
		// (agent.intent_failed, contested) instead of the old bare intent_done.
		ch := s.chestAt(in.TargetX, in.TargetY)
		if ch == nil || ch.Store == nil {
			return append(events, intentFailedEvents(s, i, a, in, intentFailContested, nextTick)...)
		}
		kinds := []string{in.Kind}
		if in.Kind == "" {
			kinds = canonicalKinds
		}
		free := freeBulk(a.Inv)
		moved := false
		for _, kind := range kinds {
			if free <= 0 {
				break
			}
			take := carriedCount(*ch.Store, kind)
			if in.Kind != "" && in.Qty > 0 && in.Qty < take {
				take = in.Qty
			}
			if take > free {
				take = free
			}
			if take <= 0 {
				continue
			}
			emit("agent.withdrew", WithdrewPayload{Agent: Ref(i), X: in.TargetX, Y: in.TargetY, Kind: kind, N: take, Owner: Ref(ch.Owner)})
			free -= take
			moved = true
		}
		if !moved {
			return append(events, intentFailedEvents(s, i, a, in, intentFailContested, nextTick)...)
		}
		// T029 (spec 013 US4): a non-owner withdrawal is theft — never blocked
		// (the transfer above already stands), always marked. In THIS same batch,
		// after the agent.withdrew event(s) and in contract order (events.md
		// "Companion batch on a non-owner withdrawal"), the executor co-emits the
		// social consequences through the existing machinery: the taking record,
		// the reason-tagged relation delta, the owner's gossip-seed memory, and a
		// witness memory for each neighbor who saw it. Owner-from-own-chest ⇒
		// agent.withdrew alone (US4-AS4). All companions are additive; none can
		// undo the goods that already moved (FR-012).
		if owner := ch.Owner; owner != i && owner >= 0 && owner < len(s.Agents) {
			events = append(events, theftCompanions(s, owner, i, in.TargetX, in.TargetY, nextTick, a.Name)...)
		}
		return events
	}

	// Validity: the resource may have vanished while walking (someone else
	// got there first).
	valid := true
	switch in.Goal {
	case "forage":
		valid = effectiveKind(m, s, in.TargetX, in.TargetY) == worldmap.Forage
	case "chop":
		valid = effectiveKind(m, s, in.ResX, in.ResY) == worldmap.Tree
	case "hunt":
		valid = denReadyAt(s, in.TargetX, in.TargetY, nextTick)
	case "build_fire", "build_shelter", "build_oven", "build_chest", "build_path":
		// build_path is stand-on-target like fire/oven/chest (paths are walkable),
		// so it re-validates the same buildSite(Target) way (spec 032 US3, T018).
		valid = buildSite(m, s, in.TargetX, in.TargetY)
	case "build_wall_plank", "build_wall_stone":
		// Spec 032 US1 (FR-007, research R2): walls build ADJACENT — the wall
		// lands on the Res tile while the builder stands on Target. Spec 038: the
		// reserved-tile OCCUPANCY guard no longer cancels here — during work only
		// the SITE is re-validated (a vanished site fails loudly below), and the
		// never-entomb guard moves to the completion moment (defer, then a bounded
		// loud fail) so a passerby crossing the tile no longer kills the build.
		valid = buildSite(m, s, in.ResX, in.ResY)
	case "demolish":
		// Contested-wall (research R5): someone else may have destroyed this wall
		// while the demolisher worked — a vanished wall resolves via intent_done.
		valid = wallAt(s, in.ResX, in.ResY) != nil
	case "repair":
		// Repair needs a still-standing, still-damaged wall AND 1 matching
		// material still carried; any of those failing (wall gone, mended by
		// another, material spent) resolves via intent_done (research R5).
		w := wallAt(s, in.ResX, in.ResY)
		valid = w != nil && w.HP < wallMaxHP(w.Kind) && invField(a.Inv, wallRepairMaterial(w.Kind)) >= 1
	case "quarry":
		// Contested-resource pattern (FR-002, spec 012 AC#5): someone else may
		// have quarried this outcrop while this agent walked over.
		valid = effectiveKind(m, s, in.ResX, in.ResY) == worldmap.Rock
		// collect_water: no depletion check — water sources are inexhaustible.
	case "cook":
		// T031: the station must still be a lit fire OR an oven (ovens carry
		// no fuel window of their own) — a fire that went cold while walking
		// over (or during the work) yields no cooked food (edge case: fire
		// burns out mid-cook). Re-validated every tick.
		valid = litFireAt(s, in.TargetX, in.TargetY, nextTick) || s.structureAt("oven", in.TargetX, in.TargetY)
	case "bathe":
		// T032: the oven itself must still be there (it never goes cold —
		// only carried water/wood, checked at completion).
		valid = s.structureAt("oven", in.TargetX, in.TargetY)
	}
	if !valid {
		// Spec 038: a build goal whose site re-validation fails mid-work resolves
		// LOUDLY and distinctly — agent.build_failed + a situated failure memory —
		// so a cancelled build is never mistaken for a finished one (the root of
		// the phantom-wall belief loop). Spec 096 generalizes the same LOUD
		// resolution to every non-build goal in this switch (forage/chop/hunt/
		// demolish/repair/quarry/cook/bathe): agent.intent_failed with reason
		// intentFailTargetGone, instead of the bare intent_done these used to
		// resolve through silently.
		if isBuildGoal(in.Goal) {
			return append(events, buildFailedEvents(s, i, a, in, buildFailSiteUnbuildable, nextTick)...)
		}
		return append(events, intentFailedEvents(s, i, a, in, intentFailTargetGone, nextTick)...)
	}

	if in.WorkStart == 0 {
		emit("agent.work_started", WorkStartedPayload{Agent: Ref(i), Tick: nextTick})
		return events
	}
	if nextTick-in.WorkStart < workDuration(s, a, in) {
		return events // still working
	}

	// US1-AS1 zero-space guard (T011): a gather whose taker has no free bulk
	// does not happen — no harvest event and, crucially, no depletion (the
	// tree/den/outcrop/forage tile is left untouched for later). The intent
	// simply resolves. Same contested-resource re-validation as the vanished-
	// resource case above, keyed on the pouch instead of the world (research R2).
	switch in.Goal {
	case "forage", "chop", "hunt", "quarry", "collect_water":
		if freeBulk(a.Inv) == 0 {
			emit("agent.intent_done", AgentPayload{Agent: Ref(i)})
			return events
		}
	}

	// Spec 019 (US1): situate every memory this completion emits by the acting
	// agent's tile, and carry the driving intent's reason (in.Reason; "" for
	// reflex) into the memory's Why — baked at emission so replay reproduces the
	// identical situated text with no lookup.
	where := PlaceAt(s, a.X, a.Y)
	// axeBrokeIfLast co-emits agent.axe_broke immediately after a chop/quarry
	// completion when the most-worn carried axe is on its last use (pre-event
	// Axes[0] == 1) — the spear-broke precedent (T027). The harvest reducer
	// decrements Axes[0] to 0 first, then this companion removes it; a situated
	// memory rides alongside carrying no Why (the reason belongs to the harvest).
	// Judged against pre-mutation state, exactly what the reducer re-derives.
	axeBrokeIfLast := func() {
		if len(a.Inv.Axes) > 0 && a.Inv.Axes[0] == 1 {
			emit("agent.axe_broke", AxeBrokePayload{Agent: Ref(i)})
			events = append(events, situatedMemoryEvent(nextTick, i, salAxeBroke, where, "", OriginAction,
				"My axe broke at the work — I'll need to craft another."))
		}
	}
	switch in.Goal {
	case "forage":
		emit("agent.foraged", HarvestPayload{Agent: Ref(i), X: in.TargetX, Y: in.TargetY})
		if a.Needs.Food < 150 {
			events = append(events, situatedMemoryEvent(nextTick, i, salStarvingForage,
				where, in.Reason, OriginAction, "Found food when I was starving."))
		}
	case "chop":
		emit("agent.chopped", HarvestPayload{Agent: Ref(i), X: in.ResX, Y: in.ResY})
		// Spec 081 (FR-003): the actor's own harvest is a first-person memory in
		// the salHunt band, riding the same batch as the act (the hunt precedent
		// above) — replacing the third-party-voiced map_corrected discovery the
		// actor used to receive of its own felling.
		events = append(events, situatedMemoryEvent(nextTick, i, salChop, where, in.Reason, OriginAction,
			chopMemoryText, in.ResX, in.ResY))
		axeBrokeIfLast()
	case "hunt":
		// T027: carrying a spear (checked against pre-mutation state, exactly
		// what the reducer will independently re-derive when it applies this
		// event) raises the yield and spends the most-worn spear's last use.
		// Spent-to-zero breaks it — a companion agent.spear_broke rides the
		// same batch, immediately after, so apply order matches: the hunt
		// reducer decrements Spears[0] to 0, then spear_broke removes it.
		emit("agent.hunted", HarvestPayload{Agent: Ref(i), X: in.TargetX, Y: in.TargetY})
		events = append(events, situatedMemoryEvent(nextTick, i, salHunt, where, in.Reason, OriginAction,
			"Hunted at the den and came back with meat."))
		if len(a.Inv.Spears) > 0 && a.Inv.Spears[0] == 1 {
			emit("agent.spear_broke", SpearBrokePayload{Agent: Ref(i)})
			// The broken spear is an incidental consequence, not the driven act —
			// situated by place but carrying no Why (the reason belongs to the hunt
			// memory above), which also avoids a double em-dash in this base text.
			events = append(events, situatedMemoryEvent(nextTick, i, salSpearBroke, where, "", OriginAction,
				"My spear broke on the hunt — I'll need to craft another."))
		}
	case "build_fire":
		emit("agent.built", BuiltPayload{Agent: Ref(i), Kind: "fire", X: in.TargetX, Y: in.TargetY})
		events = append(events, situatedMemoryEvent(nextTick, i, salFire, placeForBuild(s, a.X, a.Y, "fire"), in.Reason, OriginAction, "Built a fire."))
	case "build_shelter":
		emit("agent.built", BuiltPayload{Agent: Ref(i), Kind: "shelter", X: in.TargetX, Y: in.TargetY})
		events = append(events, situatedMemoryEvent(nextTick, i, salShelter, placeForBuild(s, a.X, a.Y, "shelter"), in.Reason, OriginAction,
			"Raised a shelter with my own hands."))
	case "build_oven":
		// T030: the flagship station. "First oven" wording (research R8) is
		// accurate here — s.hasStructure checks the pre-mutation state, before
		// this very build lands. Village-visible: nearby living agents get a
		// witness memory too, same pattern as a witnessed death.
		first := !s.hasStructure("oven")
		emit("agent.built", BuiltPayload{Agent: Ref(i), Kind: "oven", X: in.TargetX, Y: in.TargetY})
		text := "Raised an oven for the village."
		if first {
			text = "Raised the village's first oven — meals and baths, at last."
		}
		events = append(events, situatedMemoryEvent(nextTick, i, salOvenBuilt, placeForBuild(s, a.X, a.Y, "oven"), in.Reason, OriginAction, "%s", text))
		for w := range s.Agents {
			if w == i || s.Agents[w].Dead {
				continue
			}
			if abs(s.Agents[w].X-in.TargetX)+abs(s.Agents[w].Y-in.TargetY) <= witnessRadius {
				events = append(events, situatedMemoryAboutEvent(nextTick, w, i, toneOvenBuilt, salOvenBuilt,
					PlaceAt(s, s.Agents[w].X, s.Agents[w].Y), OriginWitness, "Watched %s raise an oven for the village.", a.Name))
			}
		}
	case "build_chest":
		// T023/T030 (spec 013 US3/US4): the first owned container. Site re-validated
		// above (buildSite, including the pile-tile exclusion); the reducer consumes
		// the planks and stamps Owner + an empty Store. Village-visible, like the
		// oven: the builder remembers raising it, nearby living agents get a witness
		// memory. "First chest" wording checks the pre-mutation state (this build
		// hasn't landed yet), matching build_oven.
		first := !s.hasStructure("chest")
		emit("agent.built", BuiltPayload{Agent: Ref(i), Kind: "chest", X: in.TargetX, Y: in.TargetY})
		text := "Built a chest to keep the village's things."
		if first {
			text = "Built the village's first chest — a place to keep things safe."
		}
		events = append(events, situatedMemoryEvent(nextTick, i, salChestBuilt, placeForBuild(s, a.X, a.Y, "chest"), in.Reason, OriginAction, "%s", text))
		for w := range s.Agents {
			if w == i || s.Agents[w].Dead {
				continue
			}
			if abs(s.Agents[w].X-in.TargetX)+abs(s.Agents[w].Y-in.TargetY) <= witnessRadius {
				events = append(events, situatedMemoryAboutEvent(nextTick, w, i, toneChestBuilt, salChestBuilt,
					PlaceAt(s, s.Agents[w].X, s.Agents[w].Y), OriginWitness, "Watched %s build a chest for the village.", a.Name))
			}
		}
	case "build_wall_plank", "build_wall_stone":
		// Spec 038: at the completion moment an occupied reserved tile DEFERS the
		// build rather than cancelling it — never entomb an agent in a wall. The
		// deferral is bounded: once an occupant outlasts wallOccupancyGraceTicks
		// beyond the due tick (a permanent squatter), the build fails loudly
		// (site blocked too long) instead of waiting forever. The bound is a pure
		// function of WorkStart, so no new state is needed and replay is trivially
		// deterministic (research D2). Site loss is already caught loudly by the
		// validity switch above; here only occupancy remains.
		if agentAt(s, in.ResX, in.ResY) {
			if nextTick-in.WorkStart >= workDuration(s, a, in)+wallOccupancyGraceTicks {
				return append(events, buildFailedEvents(s, i, a, in, buildFailSiteBlocked, nextTick)...)
			}
			return events // defer: no event this tick; completion fires the first clear tick
		}
		// Spec 032 US1: the wall lands on the Res tile (adjacent-stand build). The
		// reducer stamps HP = wallMaxHP(kind) and spends the recipe inputs. A
		// situated builder memory rides along at the shelter salience tier
		// (events.md: "wall builds emit a situated builder memory"); the wall is
		// at Res, so the memory is situated by the builder's own stand tile.
		r, _ := recipeFor(in.Goal)
		emit("agent.built", BuiltPayload{Agent: Ref(i), Kind: r.Structure, X: in.ResX, Y: in.ResY})
		events = append(events, situatedMemoryEvent(nextTick, i, salShelter, where, in.Reason, OriginAction, "Built a wall."))
	case "build_path":
		// Spec 032 US3: the generic reducer arm spends the stone and appends the
		// path structure (no HP — isWall is false for "path"). Built on the Target
		// tile (stand-on-target). No builder memory — a path is not formative
		// (events.md: paths emit none, the forage/chop spam-avoidance precedent).
		emit("agent.built", BuiltPayload{Agent: Ref(i), Kind: "path", X: in.TargetX, Y: in.TargetY})
	case "demolish":
		// Spec 032 US1 (research R5): one chip per completed work cycle. When the
		// chip would leave the wall standing (HP − chip ≥ 1) emit wall_chipped and
		// the reducer re-arms the work gate (WorkStart = 0) for the next cycle;
		// otherwise emit wall_destroyed and the reducer removes the wall and clears
		// the intent. Validity above guarantees the wall still stands here. No
		// memory — a chip is not formative (spam avoidance, forage/chop precedent).
		if w := wallAt(s, in.ResX, in.ResY); w != nil && w.HP-demolishChipHP >= 1 {
			emit("agent.wall_chipped", WallWorkPayload{Agent: Ref(i), X: in.ResX, Y: in.ResY})
		} else {
			emit("agent.wall_destroyed", WallWorkPayload{Agent: Ref(i), X: in.ResX, Y: in.ResY})
		}
	case "repair":
		// Spec 032 US1 (research R5): one repair per completed work cycle. The
		// reducer consumes 1 matching material, clamps HP up to the derived max,
		// and either re-arms the work gate (still damaged AND material remains) or
		// clears the intent. Validity above guarantees a damaged wall + material.
		emit("agent.wall_repaired", WallWorkPayload{Agent: Ref(i), X: in.ResX, Y: in.ResY})
	case "quarry":
		emit("agent.quarried", HarvestPayload{Agent: Ref(i), X: in.ResX, Y: in.ResY})
		// Spec 081 (FR-003): quarry parity with the chop memory above.
		events = append(events, situatedMemoryEvent(nextTick, i, salQuarry, where, in.Reason, OriginAction,
			quarryMemoryText, in.ResX, in.ResY))
		axeBrokeIfLast()
	case "collect_water":
		emit("agent.collected_water", HarvestPayload{Agent: Ref(i), X: in.ResX, Y: in.ResY})
	case "craft_planks", "craft_stone", "craft_spear", "craft_axe":
		// T026: inputs re-validated at completion (contested-resource
		// pattern). Hand-crafts have no travel window (target = the agent's
		// own tile), so this is normally a formality, but the rule applies
		// uniformly with every other completion. Spec 096: insufficient
		// inputs resolve LOUDLY (agent.intent_failed, contested), no
		// agent.crafted.
		r, _ := recipeFor(in.Goal)
		if !hasItems(a.Inv, r.Inputs) {
			return append(events, intentFailedEvents(s, i, a, in, intentFailContested, nextTick)...)
		}
		// US1 (T012): a craft doesn't truncate — it either fits or it doesn't
		// happen. The completion re-validation extends to the net bulk delta
		// (outputs − inputs, the inputs freeing their own space first); if the
		// net won't fit, no agent.crafted, intent cleared LOUDLY (spec 096,
		// contested — same reason as insufficient inputs). Only craft_planks
		// has a positive net (research R2).
		if craftNetBulk(r) > freeBulk(a.Inv) {
			return append(events, intentFailedEvents(s, i, a, in, intentFailContested, nextTick)...)
		}
		emit("agent.crafted", CraftedPayload{Agent: Ref(i), Kind: craftKindFor(in.Goal)})
	case "cook":
		// T021/T031: convert up to a batch of raw food — fire produces
		// food_cooked (fuel-free, the fire's own fire burns); an oven
		// produces meals and additionally burns 1 carried wood fuel. Spec 096:
		// no carried wood at an oven (fuel required from day one, FR-017) or
		// nothing to cook (no raw carried) resolve LOUDLY (agent.intent_failed,
		// contested) instead of the old bare intent_done.
		atOven := s.structureAt("oven", in.TargetX, in.TargetY)
		if atOven && a.Inv.Wood < 1 {
			return append(events, intentFailedEvents(s, i, a, in, intentFailContested, nextTick)...)
		}
		consumed := a.Inv.FoodRaw
		if consumed > ovenBatchSize {
			consumed = ovenBatchSize
		}
		if consumed <= 0 {
			return append(events, intentFailedEvents(s, i, a, in, intentFailContested, nextTick)...)
		}
		if atOven {
			emit("agent.cooked", CookedPayload{
				Agent: Ref(i), Station: "oven", Consumed: consumed, Produced: consumed, Kind: "meals",
			})
		} else {
			emit("agent.cooked", CookedPayload{
				Agent: Ref(i), Station: "fire", Consumed: consumed, Produced: consumed, Kind: "food_cooked",
			})
		}
	case "bathe":
		// T032: re-validate carried water + wood at completion. Spec 096:
		// missing either resolves LOUDLY (agent.intent_failed, contested —
		// water's only v1 consumer) instead of the old bare intent_done.
		if a.Inv.Water < 1 || a.Inv.Wood < 1 {
			return append(events, intentFailedEvents(s, i, a, in, intentFailContested, nextTick)...)
		}
		morale := minInt(1000, a.Needs.Morale+bathMorale)
		warmth := minInt(1000, a.Needs.Warmth+bathWarmth)
		emit("agent.bathed", BathedPayload{Agent: Ref(i), MoraleAfter: morale, WarmthAfter: warmth})
		events = append(events, situatedMemoryToned(nextTick, i, salBath, toneBath, where, in.Reason, OriginAction,
			"Took a hot bath at the oven — warm, clean, and content."))
	}
	return events
}

// isBuildGoal reports whether goal is one of the seven build_* goals whose
// mid-work re-validation failures resolve LOUDLY via agent.build_failed (spec
// 038, research D5). Every other goal's invalid/contested resolution is
// generalized the same way, via agent.intent_failed (spec 096, below).
func isBuildGoal(goal string) bool {
	switch goal {
	case "build_fire", "build_shelter", "build_oven", "build_chest", "build_path",
		"build_wall_plank", "build_wall_stone":
		return true
	}
	return false
}

// buildStructureName is the player-facing noun for a build goal, for the
// builder's failure memory ("My stone wall was never built …").
func buildStructureName(goal string) string {
	switch goal {
	case "build_fire":
		return "fire"
	case "build_shelter":
		return "shelter"
	case "build_oven":
		return "oven"
	case "build_chest":
		return "chest"
	case "build_path":
		return "path"
	case "build_wall_plank":
		return "plank wall"
	case "build_wall_stone":
		return "stone wall"
	}
	return "structure"
}

// buildFailureCause renders a stable build-failure reason as a natural clause
// for the builder's memory text (no internal em-dash, so a driving Why appends
// its own single em-dash cleanly — the spear-broke double-dash precedent).
func buildFailureCause(reason string) string {
	if reason == buildFailSiteBlocked {
		return "the site stayed blocked too long"
	}
	return "the site was no longer buildable"
}

// buildFailedEvents resolves a cancelled build LOUDLY (spec 038): agent.build_failed
// — which clears the intent exactly like intent_done, spending no material and
// standing no structure — paired SAME-TICK with a situated first-person failure
// memory (OriginAction, shelter salience so it competes with the success memory
// it falsifies). Every build-failure site routes through here so the event and
// its memory always travel together (data-model invariant 3). The memory is
// situated by the builder's own stand tile and carries the driving intent Reason
// as its Why, matching the wall-built success memory it contradicts.
func buildFailedEvents(s *State, i int, a *Agent, in *Intent, reason string, nextTick int64) []store.Event {
	return []store.Event{
		{Tick: nextTick, Type: "agent.build_failed", Payload: mustPayload(BuildFailedPayload{Agent: Ref(i), Goal: in.Goal, Reason: reason})},
		situatedMemoryEvent(nextTick, i, salShelter, PlaceAt(s, a.X, a.Y), in.Reason, OriginAction,
			"My %s was never built: %s.", buildStructureName(in.Goal), buildFailureCause(reason)),
	}
}

// Stable reason vocabulary for agent.intent_failed (spec 096): the small
// closed set generalizing agent.build_failed's own two-member set to every
// non-build goal. intentFailTargetGone is the mid-work re-validation exit
// shared by forage/chop/hunt/demolish/repair/quarry/cook/bathe — the `valid`
// switch above found the target/resource/wall/station gone or no longer
// matching (arrival check, one site per tick). intentFailContested is the
// completion-time no-op re-check shared by craft/cook/bathe/deposit/withdraw —
// the materials, the chest's free space, or the chest itself changed out from
// under the intent between landing and completion. intentFailInvalid is a
// malformed intent argument that no amount of world-state change could ever
// satisfy — never something another agent contested (deposit's empty Kind).
const (
	intentFailTargetGone = "target gone"
	intentFailContested  = "contested"
	intentFailInvalid    = "invalid"
)

// intentFailGoalNoun is the player-facing noun phrase for a failed non-build
// goal, for the actor's failure memory ("My hunt came to nothing: …") —
// mirroring buildStructureName's role for build_failed.
func intentFailGoalNoun(goal string) string {
	switch goal {
	case "forage":
		return "foraging"
	case "chop":
		return "chopping"
	case "hunt":
		return "hunt"
	case "demolish":
		return "demolition"
	case "repair":
		return "repair"
	case "quarry":
		return "quarrying"
	case "cook":
		return "cooking"
	case "bathe":
		return "bath"
	case "craft_planks", "craft_stone", "craft_spear", "craft_axe":
		return "crafting"
	case "deposit":
		return "delivery"
	case "withdraw":
		return "withdrawal"
	}
	return "task"
}

// intentFailureCause renders a stable intent_failed reason as a natural
// clause for the actor's memory text (no internal em-dash, matching
// buildFailureCause's convention) — one clause per reason, shared across
// every goal that reason applies to (the reason IS the semantic content; the
// goal noun above supplies the specificity).
func intentFailureCause(reason string) string {
	switch reason {
	case intentFailContested:
		return "it was gone by the time I got there"
	case intentFailInvalid:
		return "I never had what it needed"
	default: // intentFailTargetGone
		return "the target was gone"
	}
}

// intentFailedEvents resolves a non-build goal's invalid-exit or contested
// no-op LOUDLY (spec 096, generalizing agent.build_failed/spec 038):
// agent.intent_failed — which clears the intent exactly like intent_done, no
// yield, no side effect — paired SAME-TICK with a situated first-person
// failure memory at the build-failure salience tier (salIntentFailed — the
// card's "same salience shape, no new flooding vector" edge case), so a
// resolved-without-effect intent is never mistaken for a completed one.
// Position is the acting agent's own stand tile: intent_failed spans goals
// with no single shared site-vs-stand-tile convention the way builds do
// (Target for most, Res for the adjacent-work goals), so the one address
// every goal shares — where the agent itself is standing, matching where the
// paired memory is situated — is what the payload carries.
func intentFailedEvents(s *State, i int, a *Agent, in *Intent, reason string, nextTick int64) []store.Event {
	return []store.Event{
		{Tick: nextTick, Type: "agent.intent_failed", Payload: mustPayload(IntentFailedPayload{
			Agent: Ref(i), Goal: in.Goal, Reason: reason, X: a.X, Y: a.Y,
		})},
		situatedMemoryEvent(nextTick, i, salIntentFailed, PlaceAt(s, a.X, a.Y), in.Reason, OriginAction,
			"My %s came to nothing: %s.", intentFailGoalNoun(in.Goal), intentFailureCause(reason)),
	}
}

// workDuration is the completion-timing rule for the two goals whose
// duration depends on context rather than the goal string alone (spec 012):
// a spear-carrying hunt is faster, and cooking at an oven takes longer than
// at a fire. Both are derived from current state — Spears/the target
// structure — never persisted on the Intent, matching the codebase's
// "duration is encoded in WorkStart + completion timing, not the payload"
// convention (contracts/events.md).
func workDuration(s *State, a *Agent, in *Intent) int64 {
	switch in.Goal {
	case "hunt":
		if len(a.Inv.Spears) > 0 {
			return huntTicksSpear
		}
	case "cook":
		if s.structureAt("oven", in.TargetX, in.TargetY) {
			return cookOvenTicks
		}
	}
	return intentDuration(in.Goal)
}

func decayNeeds(n Needs, asleep, night, warm, onShelter, coldSnap bool) Needs {
	n.Food = maxInt(0, n.Food-foodDecay)
	if asleep {
		// T037: sleeping on a shelter tile recovers rest at the boosted rate
		// (the plank economy's payoff for the structure).
		regen := restRegenSleep
		if onShelter {
			regen = restRegenShelter
		}
		n.Rest = minInt(1000, n.Rest+regen)
	} else {
		n.Rest = maxInt(0, n.Rest-restDecayAwake)
	}
	switch {
	case warm:
		n.Warmth = minInt(1000, n.Warmth+warmthGainFire)
	case night && coldSnap:
		// Spec 077 (FR-010): a cold snap is "a colder night" — the SAME
		// outdoor-night arithmetic at the harsher rate; fire warmth (above)
		// beats it exactly as it beats ambient cold, and expiry is the
		// caller's read-time coldSnapActive check (no end event).
		n.Warmth = maxInt(0, n.Warmth-warmthLossColdSnap)
	case night:
		n.Warmth = maxInt(0, n.Warmth-warmthLossCold)
	default:
		n.Warmth = minInt(1000, n.Warmth+warmthGainDay)
	}
	if n.Food == 0 || n.Warmth == 0 {
		n.Health = maxInt(0, n.Health-healthLoss)
	} else if n.Food > 300 && n.Rest > 200 {
		n.Health = minInt(1000, n.Health+healthRegen)
	}
	if n.Food < 200 || n.Rest < 200 || n.Warmth < 200 {
		n.Morale = maxInt(0, n.Morale-1)
	} else if n.Morale < 700 {
		n.Morale++
	}
	return n
}

// wakeReason: day breaks with decent rest, a hunger emergency the agent can
// actually act on (food in hand), or (spec 064 US4) a COLD emergency the agent
// can act on. Fully-rested agents sleep through the night regardless — waking
// bored at 4am with nothing to do but sleep again churned sleep/wake events
// endlessly; the emergency arms keep that churn bound with an actionable guard.
func wakeReason(s *State, m *worldmap.Map, i int, night bool, tick int64) bool {
	a := &s.Agents[i]
	if !night && a.Needs.Rest >= 600 {
		return true
	}
	// Wake to a hunger emergency the agent can act on: any food in hand (T018).
	if a.Needs.Food < 150 && hasAnyFood(a) {
		return true
	}
	// Wake to a cold emergency the agent can act on (spec 064 US4, audit Gap C —
	// Oak's final night: warmth 636→0 while asleep, the wake gate blind to cold).
	// Mirrors the hunger-emergency shape EXACTLY: an EMERGENCY-floor value
	// (exposureWakeBelow 150 — a genuine exposure spiral, NOT the routine-dip
	// danger band; the hunger wake likewise fires at 150, not the hungry band 350)
	// AND an actionable remedy. The remedy guard is the warmth ladder returning an
	// actionable intent (reach a known fire, refuel, build, or chop) — the hunger
	// wake's "food in hand" analog. That guard is the churn bound Gap C was
	// deferred for: a sleeper who could only sleep in place is not woken to
	// re-sleep every beat. Night-gated because warmth is a death spiral only at
	// night (decayNeeds loses warmth only when night; by day it passively
	// regenerates), so a cozy fire-side sleeper — warmth rising at warmAt, never
	// near the floor — sleeps through, and the wake keys on the emergency band,
	// not on "night" alone (US4 AS2 control). The ladder read is pure (stepEvents
	// doctrine) and gated behind the cheap band test, so idle sleepers cost nothing.
	if night && a.Needs.Warmth < exposureWakeBelow {
		if _, ok := warmthLadder(s, m, a, tick); ok {
			return true
		}
	}
	return false
}

// eatOutcome computes the most-nutritious-first eat (T018, FR-007): Meals →
// FoodCooked → FoodRaw, one unit at a time, until the Food need reaches
// satietyAt or the inventory runs out. It returns the outcome payload
// (consumed counts per form + the absolute post-eat need) and whether anything
// is eaten — false when already sated or carrying no food, so no unit is ever
// consumed at satiety (the eating-overshoot edge case). The caller sets Agent.
func eatOutcome(a *Agent) (AtePayload, bool) {
	food := a.Needs.Food
	if food >= satietyAt {
		return AtePayload{}, false
	}
	availM, availC, availR := a.Inv.Meals, a.Inv.FoodCooked, a.Inv.FoodRaw
	var meals, cooked, raw int
	for food < satietyAt && (availM > 0 || availC > 0 || availR > 0) {
		switch {
		case availM > 0:
			availM--
			meals++
			food = minInt(1000, food+mealRestore)
		case availC > 0:
			availC--
			cooked++
			food = minInt(1000, food+foodCookedRestore)
		default: // availR > 0
			availR--
			raw++
			food = minInt(1000, food+foodRawRestore)
		}
	}
	if meals == 0 && cooked == 0 && raw == 0 {
		return AtePayload{}, false
	}
	return AtePayload{Meals: meals, Cooked: cooked, Raw: raw, FoodAfter: food}, true
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
