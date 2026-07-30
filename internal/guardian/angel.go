package guardian

// The scheduled cognition lane (spec 102, D2/FR-001): the agentized guardian
// thinking on its OWN cadence, beside — never replacing — the event-driven
// doors (player console turns, standing-order matches, survival watches).
// The lane is opt-in per world (tuning.json angel_cadence_ticks, 0 = off,
// FR-007): a non-opted world constructs the machinery but never arms it, so
// pre-102 worlds behave byte-identically.
//
// Doctrine carried here:
//   - D2: the "angel" decision class gates every scheduled turn through the
//     SAME router (cognition.Route/RoutePaused) villager classes ride, with a
//     budget registered BELOW planner's so the angel sheds first under
//     saturation; suppressions are recorded cog.outcome events, exactly the
//     mind's emitSuppressed shape.
//   - D6: the scheduled turn may REVIEW watch state (the turn prompt already
//     lists standing orders/designations/prophecies) but never fires an
//     order's action — order triggering has ONE arbiter, the existing
//     matchOrders → triggerWorker pipeline; nothing on this lane emits
//     guardian.order_triggered (pinned by TestAngelNeverTouchesOrderDoor).
//   - The cadence arithmetic is the shared phase-preserving advance
//     (cognition.NextPhasePreservingDue) the villager planner cadence uses —
//     one schedule implementation, two drivers (SC-004).

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/cognition"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// angelClass is the scheduled lane's decision class (registered in
// internal/cognition/registry.go beside "metatron", spec 102 D2).
const angelClass = "angel"

// angelSeed is the scheduled turn's trailing directive — the guardian's OWN
// standing instruction, no player-text sink (the system-turn discipline).
// It review-only frames the order machinery (D6): watches fire through their
// own triggers, never through this lane.
const angelSeed = `Your own watch cadence has come round — no player message, no order due. ` +
	`Survey the village as it stands: the clock, the chronicle, your notes and memories, ` +
	`your watches and plans. If your charter permits an act on your own initiative and one is ` +
	`truly warranted, you may take it (one act at most, as ever); otherwise keep your notes and ` +
	`hold your counsel for the player's next visit — a quiet turn is a good turn. Standing ` +
	`orders fire through their own watch machinery when their conditions come true: review ` +
	`them here if you wish, but never carry out a standing order's action from this turn.`

// angelJob is one scheduled turn the absorb goroutine armed: the snapshot
// tick (the cog.thought identity) and the router's prediction for telemetry.
type angelJob struct {
	tick            int64
	predictedWallMs int64
	predictedLand   int64
}

// scheduleAngel arms/fires the cadence (absorb goroutine only — it reads the
// replica). Called from run() after the replica applied the batch and the
// mirror refreshed, the matchOrders position. The lane is inert (returns
// immediately) unless the world opted in via tuning.
func (mt *Guardian) scheduleAngel() {
	cadence := mt.replica.AngelCadence()
	if cadence <= 0 || mt.replica.Ended {
		return
	}
	tick := mt.replica.Tick
	if mt.angelDue == 0 {
		// First sighting of an opted-in world arms one full cadence out —
		// boot/opt-in never fires a turn at tick zero.
		mt.angelDue = tick + cadence
		return
	}
	if tick < mt.angelDue || mt.angelInFlight.Load() {
		return
	}
	// The cognition-horizon gate (D2, the mind.plan discipline): a scheduled
	// turn whose predicted drift exceeds the angel budget at this speed is
	// never attempted — DegradeSkip, recorded with its arithmetic.
	dc, ok := cognition.ClassFor(angelClass)
	if !ok {
		return // unregistered class: a code bug ValidateKinds would have caught at boot
	}
	spp := mt.secondsPerPoint(llm.KindAngel)
	var v cognition.Verdict
	switch {
	case mt.replica.Paused:
		v = cognition.RoutePaused(dc, spp)
	default:
		tps := mt.replica.Speed.TicksPerSecond()
		if tps <= 0 {
			v = cognition.Verdict{Allow: true, Class: angelClass, Points: dc.Points, BudgetTicks: dc.BudgetTicks}
		} else {
			v = cognition.Route(dc, tps, spp)
		}
	}
	if !v.Allow {
		mt.emitAngelSuppressed(tick, v)
		mt.angelDue = cognition.NextPhasePreservingDue(mt.angelDue, tick, cadence)
		return
	}
	job := angelJob{tick: tick, predictedWallMs: v.PredictedWallMs, predictedLand: tick + v.PredictedDriftTicks}
	mt.angelInFlight.Store(true)
	select {
	case mt.angelQ <- job:
		mt.angelDue = cognition.NextPhasePreservingDue(mt.angelDue, tick, cadence)
	default:
		mt.angelInFlight.Store(false) // queue full; the next batch retries
	}
}

// secondsPerPoint asks the orchestrator for the live per-kind estimate — the
// mind's estimating-seam pattern (internal/mind/telemetry.go). A test fake
// without the seam falls back to the pessimistic bootstrap seed (fail toward
// skip, never toward stale action).
func (mt *Guardian) secondsPerPoint(kind llm.Kind) float64 {
	if e, ok := mt.orch.(interface {
		EstimateForKind(kind llm.Kind) (string, float64, bool)
	}); ok {
		if _, spp, ok := e.EstimateForKind(kind); ok {
			return spp
		}
	}
	return cognition.SeedFor(nil, "", true)
}

// angelWorker drains scheduled turns one at a time (the triggerWorker shape).
func (mt *Guardian) angelWorker() {
	defer mt.wg.Done()
	for {
		select {
		case <-mt.done:
			return
		case job := <-mt.angelQ:
			mt.runAngel(job)
		}
	}
}

// runAngel drives one scheduled guardian turn (FR-001): cog.thought opens the
// chain (decision-trail visibility, FR-008), the SHARED runTurn body does the
// work (same roster/handler/gate composition as every other origin — the D1
// "same construct" spine), and a terminal cog.outcome closes it. The
// single-flight turn slot serializes this lane with console and triggered
// turns, so the order door is never raced (D6).
func (mt *Guardian) runAngel(job angelJob) {
	defer mt.angelInFlight.Store(false)
	jobID := fmt.Sprintf("angel-metatron-%d", job.tick)
	mt.emitAngelCog(store.Event{Type: "cog.thought", Payload: mustJSON(sim.CogThoughtPayload{
		Job: jobID, Class: angelClass, Agent: sim.Ref(-1),
		SnapshotTick: job.tick, Points: angelPoints(),
		PredictedWallMs: job.predictedWallMs, PredictedLandTick: job.predictedLand,
	})})

	if !mt.acquireTurnBusy() {
		mt.emitAngelOutcome(jobID, job.tick, sim.OutcomeUnusable, "turn slot busy past the bounded wait", 0)
		return
	}
	start := time.Now()
	res, err := mt.runTurn(context.Background(), turnOrigin{system: true, angel: true, jobPrefix: "angel", jobID: jobID, seed: angelSeed})
	mt.turnBusy.Store(false)
	wallMs := time.Since(start).Milliseconds()

	if err != nil {
		mt.emitAngelOutcome(jobID, job.tick, sim.OutcomeUnusable, "scheduled turn failed: "+err.Error(), wallMs)
		return
	}
	// Terminal outcome: landed when an act reached a door this turn; adapted
	// when the guardian only observed/conversed — the quiet turn is a designed
	// outcome, not a failure, and the trail must be able to tell them apart.
	if acted, line := angelActLine(res); acted {
		mt.emitAngelOutcome(jobID, job.tick, sim.OutcomeLanded, line, wallMs)
		// The player learns what the caretaker did while they were away — the
		// triggered-turn moment discipline; quiet turns queue nothing (no
		// cadence spam).
		mt.queueMoment(fmt.Sprintf("%s — on my own watch: %s", clock.Format(job.tick), line))
		return
	}
	mt.emitAngelOutcome(jobID, job.tick, sim.OutcomeAdapted, "observed; no act taken", wallMs)
}

// angelPoints reads the registered class points (telemetry only).
func angelPoints() int {
	dc, _ := cognition.ClassFor(angelClass)
	return dc.Points
}

// angelActLine names the act a completed scheduled turn landed ("" when the
// turn only observed/conversed) — the moment/outcome vocabulary.
func angelActLine(r TurnResult) (bool, string) {
	switch {
	case r.Nudge != nil:
		return true, fmt.Sprintf("I sent a %s to %s.", r.Nudge.Form, joinNames(r.Nudge.Targets))
	case r.Miracle != nil:
		return true, "I worked a working: " + r.Miracle.Summary
	case r.Order != nil:
		return true, fmt.Sprintf("I set a watch (%s): %q.", r.Order.ID, r.Order.Condition)
	case r.Plan != nil:
		return true, "I laid a plan: " + r.Plan.Summary
	case r.Region != nil:
		return true, "I named a region: " + r.Region.Summary
	case len(r.Cancelled) > 0:
		return true, fmt.Sprintf("I released %d watch(es).", len(r.Cancelled))
	default:
		return false, ""
	}
}

func joinNames(names []string) string {
	out := ""
	for i, n := range names {
		if i > 0 {
			out += ", "
		}
		out += n
	}
	return out
}

// emitAngelSuppressed records a router suppression of the scheduled lane —
// the single terminal record of a turn never attempted (the mind's
// emitSuppressed shape, through the guardian's own social door).
func (mt *Guardian) emitAngelSuppressed(tick int64, v cognition.Verdict) {
	mt.emitAngelOutcomePayload(sim.CogOutcomePayload{
		Job:   fmt.Sprintf("angel-metatron-%d", tick),
		Class: angelClass, Agent: sim.Ref(-1),
		Outcome: sim.OutcomeSuppressed, SnapshotTick: tick,
		PredictedWallMs: v.PredictedWallMs, Reason: v.Arithmetic,
	})
}

// emitAngelOutcome records a scheduled turn's terminal outcome.
func (mt *Guardian) emitAngelOutcome(jobID string, tick int64, outcome, reason string, wallMs int64) {
	mt.emitAngelOutcomePayload(sim.CogOutcomePayload{
		Job: jobID, Class: angelClass, Agent: sim.Ref(-1),
		Outcome: outcome, SnapshotTick: tick,
		ActualWallMs: wallMs, Reason: reason,
	})
}

func (mt *Guardian) emitAngelOutcomePayload(p sim.CogOutcomePayload) {
	mt.emitAngelCog(store.Event{Type: "cog.outcome", Payload: mustJSON(p)})
}

// emitAngelCog lands angel telemetry through the social door; a rejected
// batch is logged, never fatal (the world outlives its observability).
func (mt *Guardian) emitAngelCog(events ...store.Event) {
	if mt.social == nil || len(events) == 0 {
		return
	}
	if err := mt.social.InjectSocial(events); err != nil {
		log.Printf("guardian: cadence telemetry rejected: %v", err)
	}
}
