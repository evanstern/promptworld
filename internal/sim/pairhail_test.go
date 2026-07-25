package sim

import (
	"strings"
	"testing"
)

// TASK-109 / spec 061 Phase 3 (US1): the sim-side pair cooldown on the hail
// founding path. These pin SC-001 (the world-01 loop founds exactly once, the
// replan is refused informatively), SC-002 (three paths / three gates, no
// cross-regression), and the dial=0 vacuity.

// presentTalkTo builds a metered talk_to whose target_present guard HOLDS at the
// target's real tile (the in-radius hail case — the common loop shape).
func presentTalkTo(target int, tx, ty int) InjectArgs {
	args := meteredArgs(0, "talk_to")
	args.TargetAgent = target
	args.Guards = []Guard{
		{Type: GuardTargetAlive, Target: target},
		{Type: GuardTargetPresent, Target: target, X: tx, Y: ty},
	}
	return args
}

// TestPairCooldownRefusesReplannedTalk is SC-001: the world-01 loop shape. A
// pair that just spoke (a hail-founded talk recorded on the ledger) replans a
// talk_to at once; the landing refuses it with an informative "spoke recently"
// outcome the planner sees, and NO second hail is placed — so the self-
// sustaining scene→replan→talk→scene loop founds exactly once, not forever.
func TestPairCooldownRefusesReplannedTalk(t *testing.T) {
	h := newLadderHarness(t, func(s *State) {
		s.Agents[0].X, s.Agents[0].Y = 10, 10
		s.Agents[1].X, s.Agents[1].Y = 12, 10 // distance 2: present, would hail absent the cooldown
		// The hail-founded talk that just happened, 300 ticks ago (the observed
		// median inter-talk gap) — well inside the 7200 cooldown.
		s.recordPairTalk(0, 1, s.Tick-300)
	})

	if err := h.loop.InjectIntent(presentTalkTo(1, 12, 10)); err == nil {
		t.Fatal("a within-cooldown talk_to must be refused, not landed")
	}
	p, ok := h.lastOutcome(t)
	if !ok || p.Outcome != OutcomeRejectedGuard {
		t.Fatalf("outcome = %+v, want rejected-guard", p)
	}
	if !strings.Contains(p.Reason, "spoke") {
		t.Errorf("refusal reason = %q, want an informative 'spoke recently' message the planner sees", p.Reason)
	}
	evs, _ := h.st.EventsSince(0, 0)
	if n := countType(evs, "social.hailed"); n != 0 {
		t.Errorf("within-cooldown talk_to placed %d hails, want 0 (the leak closed)", n)
	}
}

// TestPairCooldownPastCooldownFounds is SC-001's other half: once the pair is
// PAST the cooldown, the same talk_to founds normally — TASK-47 hail semantics
// otherwise intact (the hail still bypasses the ambient cooldown by design).
func TestPairCooldownPastCooldownFounds(t *testing.T) {
	h := newLadderHarness(t, func(s *State) {
		s.Agents[0].X, s.Agents[0].Y = 10, 10
		s.Agents[1].X, s.Agents[1].Y = 12, 10
		s.recordPairTalk(0, 1, s.Tick-8000) // gap 8000 >= 7200: past cooldown
	})
	if err := h.loop.InjectIntent(presentTalkTo(1, 12, 10)); err != nil {
		t.Fatalf("a past-cooldown talk_to must found: %v", err)
	}
	if p, _ := h.lastOutcome(t); p.Outcome != OutcomeLanded {
		t.Errorf("outcome = %q, want landed (past-cooldown founds)", p.Outcome)
	}
	evs, _ := h.st.EventsSince(0, 0)
	if n := countType(evs, "social.hailed"); n != 1 {
		t.Errorf("past-cooldown talk_to placed %d hails, want 1", n)
	}
}

// TestPairCooldownFirstContactFounds: a never-talked pair (no ledger record) is
// vacuously past the cooldown — first contact always founds.
func TestPairCooldownFirstContactFounds(t *testing.T) {
	h := newLadderHarness(t, func(s *State) {
		s.Agents[0].X, s.Agents[0].Y = 10, 10
		s.Agents[1].X, s.Agents[1].Y = 12, 10
	})
	if err := h.loop.InjectIntent(presentTalkTo(1, 12, 10)); err != nil {
		t.Fatalf("first-contact talk_to must found: %v", err)
	}
	evs, _ := h.st.EventsSince(0, 0)
	if n := countType(evs, "social.hailed"); n != 1 {
		t.Errorf("first-contact talk_to placed %d hails, want 1", n)
	}
}

// TestPairCooldownDialZeroVacuous is the dial=0 edge (US1-AS4 / spec 048): a
// world whose tuning.json sets encounter_cooldown_ticks=0 makes the gate
// vacuous — a within-"cooldown" pair still founds, because the SAME dial the
// mind arming reads is 0.
func TestPairCooldownDialZeroVacuous(t *testing.T) {
	h := newLadderHarness(t, func(s *State) {
		s.Agents[0].X, s.Agents[0].Y = 10, 10
		s.Agents[1].X, s.Agents[1].Y = 12, 10
		s.recordPairTalk(0, 1, s.Tick-1) // spoke 1 tick ago
		td := defaultTuning()
		td.EncounterCooldownTicks = 0
		s.Tuning = &td
	})
	if err := h.loop.InjectIntent(presentTalkTo(1, 12, 10)); err != nil {
		t.Fatalf("dial=0 must make the gate vacuous, talk_to should found: %v", err)
	}
	evs, _ := h.st.EventsSince(0, 0)
	if n := countType(evs, "social.hailed"); n != 1 {
		t.Errorf("dial=0 within-cooldown talk_to placed %d hails, want 1 (vacuous gate)", n)
	}
}

// TestHailStepGatesFoundingWithinCooldown pins the FOUNDING path directly (the
// leak the diagnosis names: hailStep founds talkEvents bypassing the ambient
// cooldown). With a hailer already adjacent to its hailed target, hailStep
// founds a talk when the pair is past the pair cooldown, and refuses to found
// one when they are within it — the pause simply lifts on window expiry.
func TestHailStepGatesFoundingWithinCooldown(t *testing.T) {
	setup := func(mutate func(*State)) *State {
		s := NewState(42, testMap(42))
		s.Tick = 10000
		// Agent 0 has flagged down agent 1 and stands adjacent (met condition).
		s.Agents[0].X, s.Agents[0].Y = 10, 10
		s.Agents[1].X, s.Agents[1].Y = 10, 11
		s.Agents[1].Hail = &AgentHail{By: 0, Until: 10500}
		if mutate != nil {
			mutate(s)
		}
		return s
	}

	// Past cooldown: hailStep founds (hail_met + agent.talked).
	past := setup(func(s *State) { s.recordPairTalk(0, 1, 1000) }) // gap 9000
	if evs := hailStep(past, 10001); countType(evs, "agent.talked") != 1 {
		t.Errorf("past-cooldown hailStep founded %d talks, want 1", countType(evs, "agent.talked"))
	}

	// Within cooldown: hailStep refuses to found — no talk, no hail_met.
	within := setup(func(s *State) { s.recordPairTalk(0, 1, 9800) }) // gap 200
	evs := hailStep(within, 10001)
	if n := countType(evs, "agent.talked"); n != 0 {
		t.Errorf("within-cooldown hailStep founded %d talks, want 0 (founding gate)", n)
	}
	if n := countType(evs, "social.hail_met"); n != 0 {
		t.Errorf("within-cooldown hailStep emitted %d hail_met, want 0", n)
	}
}

// TestAmbientGateUnchangedByDamper is SC-002: the ambient adjacency beat still
// gates on canTalk (per-agent LastTalk) exactly as before — the damper touches
// only the deliberate hail path. A pair whose LastTalk is recent does NOT
// ambient-talk; once past canTalk's cooldown it does.
func TestAmbientGateUnchangedByDamper(t *testing.T) {
	base := func(lastTalk int64) *State {
		s := NewState(42, testMap(42))
		for i := range s.Agents {
			if i > 1 {
				s.Agents[i].Dead = true
			}
		}
		s.Agents[0].X, s.Agents[0].Y = 20, 20
		s.Agents[1].X, s.Agents[1].Y = 20, 21 // adjacency for the ambient beat
		s.Agents[0].LastTalk, s.Agents[1].LastTalk = lastTalk, lastTalk
		return s
	}
	m := testMap(42)
	// Recent LastTalk (canTalk blocks: gap < talkCooldownSec through the whole
	// genesis drive window): the ambient beat founds nothing new.
	blocked := base(1)
	log := driveTicks(t, blocked, m, 400, nil)
	if countType(log, "agent.talked") != 0 {
		t.Errorf("recent LastTalk: ambient beat founded %d talks, want 0 (canTalk unchanged)", countType(log, "agent.talked"))
	}
	// Never talked (canTalk passes): the ambient beat founds.
	open := base(0)
	log2 := driveTicks(t, open, m, 400, nil)
	if countType(log2, "agent.talked") == 0 {
		t.Error("open canTalk: ambient beat should have founded a talk (unchanged)")
	}
}

// TestHailBypassesAmbientCooldownPreserved is SC-002 / the TASK-47 invariant:
// the hail founding path still bypasses the AMBIENT (canTalk / per-agent
// LastTalk) cooldown — only the new PAIR cooldown gates it. An agent with a
// recent LastTalk (canTalk would refuse an ambient talk) is still hail-founded
// when the pair has no within-cooldown ledger record.
func TestHailBypassesAmbientCooldownPreserved(t *testing.T) {
	s := NewState(42, testMap(42))
	s.Tick = 10000
	s.Agents[0].X, s.Agents[0].Y = 10, 10
	s.Agents[1].X, s.Agents[1].Y = 10, 11
	s.Agents[1].Hail = &AgentHail{By: 0, Until: 10500}
	// Recent per-agent LastTalk: the ambient canTalk gate WOULD block this pair.
	s.Agents[0].LastTalk, s.Agents[1].LastTalk = 9990, 9990
	// But NO pair ledger record (never talked as a pair, or long ago) — the hail
	// founds regardless, exactly as TASK-47 intends.
	evs := hailStep(s, 10001)
	if countType(evs, "agent.talked") != 1 {
		t.Errorf("hail must bypass the ambient LastTalk cooldown: founded %d talks, want 1", countType(evs, "agent.talked"))
	}
}
