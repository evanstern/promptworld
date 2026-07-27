package sim

// Faith — the village's event-sourced devotion score (spec 085): the endogenous
// mana loop's stock. Faith moves ONLY when recorded faith.changed events fold —
// no ambient accrual, no time decay, no model judgment anywhere in the loop —
// and charge regeneration becomes a pure band function of the folded score
// (FaithRegenCadenceTicks below), replacing the fixed 6-game-hour constant.
//
// Door discipline (research R1): faith.changed is EXECUTOR-emitted only
// (faithEvents, the charge_regenerated class — whitelist absence refuses any
// injected forgery), and the reducer arm below is the ONLY writer of
// State.Faith. Old events never mint faith retroactively: the sweep observes
// source events in the live batch and emits the faith event beside them; the
// fold reads only faith.changed, so pre-085 logs replay byte-identically.

import (
	"encoding/json"
	"fmt"

	"github.com/evanstern/promptworld/internal/store"
)

// FaithState is the village's event-sourced devotion score (spec 085 FR-001).
// nil State.Faith means the genesis default — the State.Tuning
// nil-means-default precedent; pre-085 snapshots round-trip byte-identically
// (omitempty, no format bump). Materialized by the first faith.changed fold.
// No tick fields → untouched by rebaseTicks (data-model §9).
type FaithState struct {
	Score int `json:"score"` // clamped 0..100 at fold time
}

// FaithGenesis is the score a world holds before any faith event ever folds —
// deliberately inside the steady band (below) so a world that has never folded
// a faith event regenerates charges byte-identically to pre-085. There is no
// boot seeding event (unlike sim.tuning_applied): nothing operator-authored
// exists to record; nil + the accessor is the whole compat story.
const FaithGenesis = 50

// The faith delta table (spec 085 FR-002, data-model §3) — ONE home for the
// closed reason vocabulary and its doctrine deltas. All constants are
// promoted-dial-READY (named, one place — the spec-059 survival-band
// discipline: dials are earned by evidence, deliberately NOT in tuning.json).
// The reason strings are FROZEN recorded vocabulary (spec 052 ruling 2): they
// land in faith.changed payloads. Sources deliberately EXCLUDED (research R3):
// designation.fulfilled alone (villager initiative is not the guardian's
// word), ambient accrual, time decay, tutoring (TASK-112 AC #6 — enforced by
// the rubric-hygiene faith.* ban), and metatron.nudged.
const (
	FaithReasonDirectiveFulfilled = "directive_fulfilled" // the primary endogenous source
	FaithReasonDirectiveExpired   = "directive_expired"   // the guardian's charge went unachieved
	FaithReasonVillagerDied       = "villager_died"       // the flock's suffering erodes faith
	FaithReasonProphecyFulfilled  = "prophecy_fulfilled"  // the declared word came true
	FaithReasonProphecyFailed     = "prophecy_failed"     // false prophet

	FaithDeltaDirectiveFulfilled = 8   // ~3 fulfilled directives climb one band
	FaithDeltaDirectiveExpired   = -4  // mild: half a death
	FaithDeltaVillagerDied       = -6  // the deliberate spiral feeder, one per death
	FaithDeltaProphecyFulfilled  = 12  // the strongest single faith act
	FaithDeltaProphecyFailed     = -15 // asymmetric vs +12 so claim-spam is negative-EV
)

// faithDeltaByReason is the reason → doctrine delta view of the table above —
// the sweep's emission source and the reducer arm's sign authority.
var faithDeltaByReason = map[string]int{
	FaithReasonDirectiveFulfilled: FaithDeltaDirectiveFulfilled,
	FaithReasonDirectiveExpired:   FaithDeltaDirectiveExpired,
	FaithReasonVillagerDied:       FaithDeltaVillagerDied,
	FaithReasonProphecyFulfilled:  FaithDeltaProphecyFulfilled,
	FaithReasonProphecyFailed:     FaithDeltaProphecyFailed,
}

// FaithChangedPayload — faith.changed (spec 085 FR-002): the ONE faith-movement
// event. delta is the doctrine delta at emission (the fold clamps — a partial
// clamp still records the doctrine value); reason is the closed vocabulary
// above; source_id names the source event's entity (directive id, prophecy id,
// or the dead villager's index as a decimal string). FROZEN recorded
// vocabulary (spec 052 ruling 2).
type FaithChangedPayload struct {
	Delta    int    `json:"delta"`
	Reason   string `json:"reason"`
	SourceID string `json:"source_id"`
}

// FaithScore is the nil-safe read path (FR-001): FaithGenesis when no faith
// event has ever folded. The ONLY consumption path — reducer, executor, daemon
// status, and tests all read the score through it, never the field.
func (s *State) FaithScore() int {
	if s.Faith == nil {
		return FaithGenesis
	}
	return s.Faith.Score
}

// clampFaith clamps a folded score to the 0..100 band.
func clampFaith(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// The regen band table (spec 085 FR-004/FR-005, data-model §6 — NORMATIVE).
// All cadences are divisors of the game day, so boundaries stay absolute
// multiples within any band; band changes take effect at the next check (a
// pure function of the folded score — replay-identical). chargeRegenTicks
// (guardian.go) survives as the steady band's value: genesis 50 lives there,
// so a world with no faith events keeps today's exact schedule.
const (
	faithFerventAt  = 75 // ≥ 75: the village believes; power comes easily
	faithSteadyAt   = 40 // 40..74: the old covenant pace (genesis band)
	faithWaveringAt = 15 // 15..39: doubt slows the flow; < 15: forsaken

	faithFerventCadence  = 4 * 3600  // 4 game hours
	faithWaveringCadence = 12 * 3600 // 12 game hours
	faithFloorCadence    = 24 * 3600 // the ambient forsaken floor: once per game day
)

// FaithRegenCadenceTicks is the pure faith → charge-regen cadence (spec 085
// FR-004): the executor's boundary check and the daemon's status projection
// both read THIS function — one home, exported so the wire value and the sim
// can never disagree. Returns 0 when no regen is scheduled at all (the
// scenario forsaken band; the caller's cadence-0 short-circuit).
//
// THE FR-005 POSTURE DECISION (spec 085 AC #4, research R5), recorded where
// the code lives: in the forsaken band (< 15) a SCENARIO world spirals
// authentically — no regen while the band holds; the run can die of it and
// the morgue teaches (the Hades doctrine: where the loop is run-shaped,
// failure must keep its teeth) — while an AMBIENT world floors at once per
// game day (a persistent homeworld has no run-reset; a zero-regen deadlock is
// a dead save, not a lesson; the charge-free plan verbs remain the endogenous
// exit). THE REVERSAL LEVER: the whole posture is this one band table plus
// the scenario fork argument — flipping ambient to authentic (or flooring
// scenarios) is a one-row change, and the recorded future promotion is
// tuning.json faith_floor_cadence_ticks (0 = no floor), a one-table change
// with no event or shape impact.
func FaithRegenCadenceTicks(score int, scenario bool) int64 {
	switch {
	case score >= faithFerventAt:
		return faithFerventCadence
	case score >= faithSteadyAt:
		return chargeRegenTicks // today's constant — the pre-085-identical genesis band
	case score >= faithWaveringAt:
		return faithWaveringCadence
	case scenario:
		return 0 // the authentic spiral: no regen while forsaken
	default:
		return faithFloorCadence // the ambient floor: slow, painful, never a hard lock
	}
}

// applyFaith is the faith.changed reducer arm (spec 085 FR-002) — the ONLY
// writer of State.Faith. Validates rather than clamps the vocabulary (the
// reason domain and the delta's sign against the doctrine table; magnitudes
// are dial-ready and fold as recorded), then folds with clamping,
// materializing Faith on the first fold. faith.changed is executor-emitted
// only — NOT on injectSocialWhitelist (loop.go) — so this arm never runs on
// injected model output.
func (s *State) applyFaith(e store.Event) error {
	var p FaithChangedPayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return fmt.Errorf("apply %s: %w", e.Type, err)
	}
	doctrine, ok := faithDeltaByReason[p.Reason]
	if !ok {
		return fmt.Errorf("apply %s: unknown reason %q", e.Type, p.Reason)
	}
	if p.Delta == 0 {
		return fmt.Errorf("apply %s: zero delta records no movement", e.Type)
	}
	if (p.Delta > 0) != (doctrine > 0) {
		return fmt.Errorf("apply %s: delta %d contradicts reason %q's sign", e.Type, p.Delta, p.Reason)
	}
	score := clampFaith(s.FaithScore() + p.Delta)
	if s.Faith == nil {
		s.Faith = &FaithState{}
	}
	s.Faith.Score = score
	return nil
}

// faithEvents is the faith accounting sweep (spec 085 FR-003): pure over
// (pre-tick state, this tick's batch, tick) — the run-end detector's own idiom
// (executor.go). It scans the batch for the five source events and emits one
// faith.changed per source, in the batch's own order, SKIPPING any emission
// whose fold could not move the clamped score computed from the pre-tick score
// plus this batch's PRIOR faith emissions folded in order (the
// charge_regenerated below-cap idiom: never record a movement that moves
// nothing; a partial clamp still emits the doctrine delta — the fold clamps).
// Positioned in stepEvents AFTER every faith-source emitter and BEFORE
// scenarioRubricEvents/run-end detection.
func faithEvents(s *State, batch []store.Event, nextTick int64) []store.Event {
	var events []store.Event
	running := s.FaithScore()
	emit := func(reason, sourceID string) {
		delta := faithDeltaByReason[reason]
		folded := clampFaith(running + delta)
		if folded == running {
			return // the fold could not move the clamped score — record nothing
		}
		running = folded
		events = append(events, store.Event{Tick: nextTick, Type: "faith.changed",
			Payload: mustPayload(FaithChangedPayload{Delta: delta, Reason: reason, SourceID: sourceID})})
	}
	for _, e := range batch {
		switch e.Type {
		case "directive.fulfilled":
			var p DirectiveFulfilledPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				continue // struct-built by the directive sweep; cannot fail
			}
			emit(FaithReasonDirectiveFulfilled, p.ID)
		case "directive.expired":
			var p OrderIDPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				continue
			}
			emit(FaithReasonDirectiveExpired, p.ID)
		case "agent.died":
			var p DiedPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				continue
			}
			emit(FaithReasonVillagerDied, fmt.Sprintf("%d", p.Agent.ID))
		case "prophecy.fulfilled":
			var p OrderIDPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				continue
			}
			emit(FaithReasonProphecyFulfilled, p.ID)
		case "prophecy.failed":
			var p OrderIDPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				continue
			}
			emit(FaithReasonProphecyFailed, p.ID)
		}
	}
	return events
}
