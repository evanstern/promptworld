package mind

import (
	"encoding/json"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// Belief reconciliation against grounded arrival observations (spec 097 D3,
// FR-003): when an agent.place_observed lands on the replica, the beliefs of
// the observing agent that reference the observed location move —
//
//   - CONFIRMATION: every feature the belief names is present in the
//     observation ⇒ effective confidence + BeliefConfirmBoost (capped 100),
//     decay clock re-anchored;
//   - DISCONFIRMATION: the place was observed and a named feature is absent
//     (the observation is exhaustive within its radius, so absence is
//     definite) ⇒ effective confidence × BeliefDisconfirmRetainPercent —
//     faster than the ambient 8-game-day silence half-life but bounded, so a
//     myth survives several visits before trending under the floor (dials,
//     not cliffs); the observation memory also gets the D4 surprise bump
//     (memory_promoted — surprise is memorable);
//   - SILENCE: a belief that never gets its place visited keeps today's decay
//     untouched — no event, no movement.
//
// Everything lands through the TASK-79 seam (agent.belief_reinforced via the
// whitelisted injection door, extended additively with Kind + emitter-computed
// Confidence — internal/sim/consolidate.go): the mind computes, the reducer
// copies, replay never re-derives. Matching lives HERE, mind-side only (D3):
// the executor emitted uniformly with no reason interpretation, and while the
// spec permits an LLM for matching, this implementation is a deterministic
// coordinate + feature-vocabulary matcher — cheaper, testable, and replaceable
// behind reconcileObservation without touching the sim. No LLM ever enters
// the sim emission path either way.

// disconfirmSalienceBoost is the D4 surprise bump: a disconfirming
// observation's memory is promoted from the low observation base salience
// into the salHunt band — memorable, still far below the generation-
// interrupting tier. One bump per observation, however many beliefs it
// disconfirmed.
const disconfirmSalienceBoost = 2

// statementCoordRe extracts "(x,y)" coordinate pairs — the form every
// situated memory text bakes (sim.situateText), and therefore the form
// consolidation-authored place-beliefs overwhelmingly carry.
var statementCoordRe = regexp.MustCompile(`\((\d+)\s*,\s*(\d+)\)`)

// beliefFeatureVocab maps statement word tokens to the observation kind(s)
// that would ground them — the mental-map fact vocabulary the payload's Kinds
// carries. An entry's alternatives are ANY-OF (a "wall" is grounded by either
// wall kind). Deliberately closed and conservative: a token outside this
// vocabulary contributes no expectation, and a belief with NO recognized
// feature token is unjudgeable — silence, never a guess.
var beliefFeatureVocab = map[string][]string{
	"fire": {"fire"}, "fires": {"fire"}, "campfire": {"fire"},
	"shelter": {"shelter"}, "shelters": {"shelter"},
	"oven": {"oven"}, "ovens": {"oven"},
	"chest": {"chest"}, "chests": {"chest"},
	"wall": {"wall_plank", "wall_stone"}, "walls": {"wall_plank", "wall_stone"},
	"path": {"path"}, "paths": {"path"},
	"grave": {"grave"}, "graves": {"grave"},
	"pile": {"pile"}, "piles": {"pile"}, "stockpile": {"pile"}, "goods": {"pile"},
	"tree": {"tree"}, "trees": {"tree"}, "woods": {"tree"}, "forest": {"tree"},
	"forage": {"forage"}, "berries": {"forage"}, "berry": {"forage"},
	"rock": {"rock"}, "rocks": {"rock"}, "outcrop": {"rock"}, "outcrops": {"rock"},
	"stone": {"rock"}, "stones": {"rock"}, "boulder": {"rock"}, "boulders": {"rock"},
	"water": {"water_edge"}, "river": {"water_edge"}, "lake": {"water_edge"},
	"pond": {"water_edge"}, "shore": {"water_edge"}, "shoreline": {"water_edge"},
	"den": {"den"}, "dens": {"den"},
}

// statementReferencesPlace reports whether a belief statement names a
// coordinate inside the observed disc: the observation is exhaustive only
// within its radius, so ONLY a belief about a tile in that disc is testable
// by it — anything else is silence (correct epistemics, not caution).
func statementReferencesPlace(statement string, x, y, radius int) bool {
	for _, m := range statementCoordRe.FindAllStringSubmatch(statement, -1) {
		bx, err1 := strconv.Atoi(m[1])
		by, err2 := strconv.Atoi(m[2])
		if err1 != nil || err2 != nil {
			continue
		}
		if absInt(bx-x)+absInt(by-y) <= radius {
			return true
		}
	}
	return false
}

// statementExpectations extracts the belief's checkable feature expectations:
// one any-of alternative set per recognized feature token (deduplicated by
// canonical head kind, in first-appearance order — deterministic).
func statementExpectations(statement string) [][]string {
	seen := map[string]bool{}
	var out [][]string
	for _, tok := range strings.FieldsFunc(strings.ToLower(statement), func(r rune) bool {
		return r < 'a' || r > 'z'
	}) {
		alts, ok := beliefFeatureVocab[tok]
		if !ok || seen[alts[0]] {
			continue
		}
		seen[alts[0]] = true
		out = append(out, alts)
	}
	return out
}

// reconcileObservation is the pure decision core (unit-tested directly): given
// the observing agent's replica state, one place_observed payload, the event's
// tick, and the two belief dials, it returns the belief-movement events to
// inject — belief_reinforced (confirmed/disconfirmed, emitter-computed new
// stored confidence) per judgeable matching belief, plus at most ONE
// memory_promoted surprise bump for the observation's companion memory when
// anything was disconfirmed. Beliefs that reference other places, carry no
// coordinates, or name no recognized feature produce nothing (silence).
func reconcileObservation(a *sim.Agent, agentID int, p sim.PlaceObservedPayload, tick int64, confirmBoost, retainPct int64) []store.Event {
	present := map[string]bool{}
	for _, k := range p.Kinds {
		present[k] = true
	}
	var out []store.Event
	disconfirmed := false
	for _, b := range a.Beliefs {
		if !statementReferencesPlace(b.Statement, p.X, p.Y, p.Radius) {
			continue
		}
		expects := statementExpectations(b.Statement)
		if len(expects) == 0 {
			continue // nothing checkable: silence
		}
		confirmed := true
		for _, alts := range expects {
			ok := false
			for _, k := range alts {
				if present[k] {
					ok = true
					break
				}
			}
			if !ok {
				confirmed = false
				break
			}
		}
		eff := sim.EffectiveConfidence(b, tick)
		var kind string
		var newConf int
		if confirmed {
			kind = sim.BeliefConfirmed
			newConf = eff + int(confirmBoost)
			if newConf > 100 {
				newConf = 100
			}
		} else {
			kind = sim.BeliefDisconfirmed
			newConf = int((int64(eff)*retainPct + 50) / 100)
			disconfirmed = true
		}
		payload, err := json.Marshal(sim.BeliefReinforcedPayload{
			Agent: sim.Ref(agentID), BeliefID: b.ID, Kind: kind, Confidence: newConf,
		})
		if err != nil {
			continue // struct-built; cannot fail
		}
		out = append(out, store.Event{Tick: tick, Type: "agent.belief_reinforced", Payload: payload})
	}
	if disconfirmed {
		// The D4 surprise bump: promote the observation's OWN companion memory
		// (same tick, "observed" provenance — it precedes the observation in
		// its batch, so the replica already holds it). Newest match wins.
		for i := len(a.Memories) - 1; i >= 0; i-- {
			m := a.Memories[i]
			if m.Tick == tick && m.Origin == sim.OriginObserved {
				payload, err := json.Marshal(sim.MemoryPromotedPayload{
					Agent: sim.Ref(agentID), MemTick: m.Tick,
					TextHash: sim.MemoryHash(m.Text), Boost: disconfirmSalienceBoost,
				})
				if err == nil {
					out = append(out, store.Event{Tick: tick, Type: "agent.memory_promoted", Payload: payload})
				}
				break
			}
		}
	}
	return out
}

// reconcilePlace is the absorb-side trigger (runs on the absorb goroutine,
// which owns the replica): decode the observation, compute the movement
// events against the already-applied replica, and hand them to the
// reconciliation worker — absorb never blocks on the injection door.
func (md *Mind) reconcilePlace(e store.Event) {
	var p sim.PlaceObservedPayload
	if json.Unmarshal(e.Payload, &p) != nil {
		return
	}
	if p.Agent.ID < 0 || p.Agent.ID >= len(md.replica.Agents) {
		return
	}
	a := &md.replica.Agents[p.Agent.ID]
	evs := reconcileObservation(a, p.Agent.ID, p, e.Tick,
		md.replica.BeliefConfirmBoost(), md.replica.BeliefDisconfirmRetainPercent())
	if len(evs) == 0 {
		return
	}
	select {
	case md.reconQ <- evs:
	default:
		// Bounded and self-healing: a dropped batch means this observation
		// moved no beliefs — the next un-deduped visit re-judges from the
		// same recorded evidence.
		log.Printf("mind: belief reconciliation queue full — dropped %d events", len(evs))
	}
}

// reconcileWorker drains reconciliation batches into the injection door,
// keeping the absorb goroutine free (the meetingWorker shape). Each batch is
// atomic: the door dry-runs and lands it all-or-nothing.
func (md *Mind) reconcileWorker() {
	for {
		select {
		case <-md.done:
			return
		case evs := <-md.reconQ:
			if err := md.social.InjectSocial(evs); err != nil {
				log.Printf("mind: belief reconciliation inject failed: %v", err)
			}
		}
	}
}
