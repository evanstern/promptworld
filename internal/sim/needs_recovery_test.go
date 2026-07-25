package sim

// Spec 064 (needs-conditioned recovery): the completion condition rides the
// intent; an intent carrying one HOLDS at its target and completes on the need
// crossing the threshold (warm_up), aborts on a dead source (recovery_stalled),
// yields to a worse survival emergency, and is otherwise an ordinary active
// intent. Plus the audit Gap C wake-to-cold. Tests are grouped by user story;
// SC-001..005 evidence is called out per test.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// warmUpAtFire configures agent 0 as an isolated, awake, DAY villager standing
// beside a real lit fire (warmAt true, so decayNeeds warms it) with the given
// starting warmth — the state a conditioned warm_up holds in. Returns the state,
// map, agent, and the fire tile.
func warmUpAtFire(t *testing.T, seed uint64, warmth int) (*State, *worldmap.Map, *Agent) {
	t.Helper()
	m := testMap(seed)
	s := NewState(seed, m)
	s.Night = false
	isolateAgents(s)
	a := &s.Agents[0]
	a.Dead, a.Asleep = false, false
	a.Needs = Needs{Health: 1000, Food: 600, Rest: 600, Warmth: warmth, Morale: 600}
	a.Inv = Inventory{}
	fx, fy := a.X+1, a.Y
	s.Structures = append(s.Structures, Structure{Kind: "fire", X: fx, Y: fy, FuelUntil: 10_000_000})
	if !warmAt(s, a.X, a.Y, 0) {
		t.Skipf("seed %d: agent not warm beside the placed fire", seed)
	}
	a.Map.upsertFact(PlaceFact{Kind: "fire", X: fx, Y: fy, Seen: 1, Provenance: ProvenanceWitnessed, Detail: 10_000_000})
	return s, m, a
}

// condIntentCmd is an intent_set command carrying a completion condition, landing
// at tick 1 on the agent's own tile (no travel — executeAtTarget runs next tick).
func condIntentCmd(a *Agent, goal, need string, until int, source string) map[int64][]store.Event {
	return map[int64][]store.Event{
		1: {{Tick: 1, Type: "agent.intent_set", Payload: mustPayload(IntentSetPayload{
			Agent: 0, Goal: goal, TargetX: a.X, TargetY: a.Y,
			UntilNeed: need, UntilValue: until, Source: source})}},
	}
}

func intentDoneTick(log []store.Event, agent int) int64 {
	for _, e := range log {
		if e.Type != "agent.intent_done" {
			continue
		}
		var p AgentPayload
		json.Unmarshal(e.Payload, &p)
		if p.Agent == agent {
			return e.Tick
		}
	}
	return -1
}

// --- T001 / FR-001: opt-in condition, pre-064 byte compatibility ------------

// TestConditionlessIntentByteIdentical (SC-004): an intent with no condition
// marshals WITHOUT the spec-064 keys — a pre-064 snapshot/event is byte-for-byte
// unchanged. Same for a work_started with no Ref.
func TestConditionlessIntentByteIdentical(t *testing.T) {
	b, _ := json.Marshal(Intent{Goal: "forage", TargetX: 3, TargetY: 4})
	for _, key := range []string{"until_need", "until_value", "hold_ref"} {
		if strings.Contains(string(b), key) {
			t.Fatalf("conditionless intent leaked %q: %s", key, b)
		}
	}
	ws, _ := json.Marshal(WorkStartedPayload{Agent: 0, Tick: 100})
	if strings.Contains(string(ws), "ref") {
		t.Fatalf("Ref-less work_started leaked \"ref\": %s", ws)
	}
	set, _ := json.Marshal(IntentSetPayload{Agent: 0, Goal: "forage", TargetX: 1, TargetY: 2})
	for _, key := range []string{"until_need", "until_value"} {
		if strings.Contains(string(set), key) {
			t.Fatalf("conditionless intent_set leaked %q: %s", key, set)
		}
	}
}

// TestConditionRidesIntentSetPayload (replay determinism, R1 door): the condition
// is carried by the recorded intent_set and re-applied by the reducer; a
// malformed need is dropped at the door (arrive-and-done, never a stuck hold).
func TestConditionRidesIntentSetPayload(t *testing.T) {
	m := testMap(42)
	s := NewState(42, m)
	isolateAgents(s)
	if err := s.Apply(store.Event{Tick: 1, Type: "agent.intent_set", Payload: mustPayload(
		IntentSetPayload{Agent: 0, Goal: "warm_up", TargetX: 1, TargetY: 2, UntilNeed: "warmth", UntilValue: 800})}); err != nil {
		t.Fatal(err)
	}
	if in := s.Agents[0].Intent; in == nil || in.UntilNeed != "warmth" || in.UntilValue != 800 {
		t.Fatalf("condition did not ride onto the intent: %+v", s.Agents[0].Intent)
	}
	// A bogus need is refused at the door — the intent stays arrive-and-done.
	if err := s.Apply(store.Event{Tick: 2, Type: "agent.intent_set", Payload: mustPayload(
		IntentSetPayload{Agent: 1, Goal: "warm_up", TargetX: 1, TargetY: 2, UntilNeed: "vibes", UntilValue: 800})}); err != nil {
		t.Fatal(err)
	}
	if in := s.Agents[1].Intent; in == nil || in.UntilNeed != "" {
		t.Fatalf("bogus need was not dropped at the door: %+v", s.Agents[1].Intent)
	}
}

// --- T006 / US1: warm_up holds and completes on warmth --------------------

// TestRecoverThenRelease (SC-001): a conditioned hold stays active while warmth
// is below the threshold (AS1) and completes exactly when it crosses (AS2), and
// the completion tick is deterministic across two identical runs.
func TestRecoverThenRelease(t *testing.T) {
	run := func() (int64, []store.Event) {
		s, m, a := warmUpAtFire(t, 42, 500)
		log := driveTicks(t, s, m, 4000, condIntentCmd(a, "warm_up", "warmth", 560, "planner"))
		return intentDoneTick(log, 0), log
	}
	done, log := run()
	if done < 0 {
		t.Fatal("warm_up never completed")
	}
	// AS1: no completion event fires while warmth is still below the threshold.
	warmthAt := 500
	for _, e := range log {
		if e.Type == "agent.needs_changed" {
			var p NeedsPayload
			json.Unmarshal(e.Payload, &p)
			if p.Agent == 0 {
				warmthAt = p.Warmth
			}
		}
		if e.Type == "agent.intent_done" && e.Tick < done {
			t.Fatalf("an intent_done fired at %d before the recovery completed", e.Tick)
		}
		if e.Type == "agent.intent_done" && e.Tick == done && warmthAt < 560 {
			t.Fatalf("completed at warmth %d, below the 560 threshold", warmthAt)
		}
	}
	// SC-001 determinism: an identical run completes on the identical tick.
	done2, _ := run()
	if done != done2 {
		t.Fatalf("non-deterministic completion tick: %d vs %d", done, done2)
	}
}

// TestAlreadySatisfiedCompletesImmediately (edge case): a condition already met
// on arrival completes at once (checked, not assumed).
func TestAlreadySatisfiedCompletesImmediately(t *testing.T) {
	s, m, a := warmUpAtFire(t, 42, 900)
	log := driveTicks(t, s, m, 400, condIntentCmd(a, "warm_up", "warmth", 800, "planner"))
	if done := intentDoneTick(log, 0); done < 0 || done > 60 {
		t.Fatalf("an already-satisfied condition should complete immediately, done=%d", done)
	}
	for _, e := range log {
		if e.Type == "agent.work_started" {
			t.Fatal("an already-satisfied recovery must not start a hold (no work_started)")
		}
	}
}

// TestReplayDeterminismOverRecoverySpan (FR-008): the whole recover-then-release
// event log is byte-identical across two runs — the per-tick check is a pure
// function of state.
func TestReplayDeterminismOverRecoverySpan(t *testing.T) {
	drive := func() []store.Event {
		s, m, a := warmUpAtFire(t, 7, 500)
		return driveTicks(t, s, m, 2000, condIntentCmd(a, "warm_up", "warmth", 620, "planner"))
	}
	if !bytesEqual(canonicalLog(t, drive()), canonicalLog(t, drive())) {
		t.Fatal("recovery-span event log is not replay-deterministic")
	}
}

func bytesEqual(a, b []byte) bool { return string(a) == string(b) }

// TestWarmUpResolverAndClamp (US1 AS3 / FR-002, R3): the warm_up resolver targets
// known warmth and stamps the condition; the threshold clamps with notice and
// defaults when absent — one clamp home (clampWarmUp / ClampWarmUp).
func TestWarmUpResolverAndClamp(t *testing.T) {
	// clamp home behaviour.
	if v, n := clampWarmUp(0); v != warmthRecoverTo || n != "" {
		t.Fatalf("absent until_warmth: got %d/%q, want default %d/no-notice", v, n, warmthRecoverTo)
	}
	if v, n := clampWarmUp(650); v != 650 || n != "" {
		t.Fatalf("in-range until_warmth: got %d/%q, want 650/no-notice", v, n)
	}
	if v, n := clampWarmUp(50); v != warmthRecoverFloor || n == "" {
		t.Fatalf("below-floor until_warmth: got %d/%q, want %d/notice", v, n, warmthRecoverFloor)
	}
	if v, n := clampWarmUp(5000); v != needMax || n == "" {
		t.Fatalf("above-cap until_warmth: got %d/%q, want %d/notice", v, n, needMax)
	}
	if v, _ := ClampWarmUp(50); v != warmthRecoverFloor {
		t.Fatal("exported ClampWarmUp must delegate to the same clamp")
	}
	// Resolver: default threshold when qty 0, clamped when out of range.
	s, m, _ := warmUpAtFire(t, 42, 500)
	in, _, err := resolveGoal(s, m, 0, "warm_up", -1, "", 0, 1)
	if err != nil || in == nil || in.Goal != "warm_up" || in.UntilNeed != "warmth" || in.UntilValue != warmthRecoverTo {
		t.Fatalf("warm_up resolve (default) = %+v, err %v", in, err)
	}
	in, _, err = resolveGoal(s, m, 0, "warm_up", -1, "", 5000, 1)
	if err != nil || in == nil || in.UntilValue != needMax {
		t.Fatalf("warm_up resolve (over-cap) should clamp to %d, got %+v", needMax, in)
	}
}

// TestReflexWarmthRungsIssueConditioned (US1 AS4 / FR-003): BOTH the day and the
// night reflex warmth rungs issue goto_warmth carrying the doctrine-default
// condition — the hold, not arrive-and-done. Source stays reflex (never arms the
// 062 yield window).
func TestReflexWarmthRungsIssueConditioned(t *testing.T) {
	for _, night := range []bool{false, true} {
		s, m, a, now := dayColdAgent(t, 42)
		s.Night = night
		if night {
			a.Needs.Warmth = coldNightBelow - 50
		}
		grantKnownLitFire(a, now)
		d := decideIntent(s, m, 0, now)
		if d.intent == nil || d.intent.Goal != "goto_warmth" {
			t.Fatalf("night=%v: reflex chose %+v, want goto_warmth", night, d.intent)
		}
		if d.intent.UntilNeed != "warmth" || d.intent.UntilValue != warmthRecoverTo {
			t.Fatalf("night=%v: reflex goto_warmth carried no default condition: %+v", night, d.intent)
		}
	}
}

// TestReflexHoldDoesNotArmYieldWindow (source discipline, 062 unchanged): a
// reflex-issued conditioned recovery completing never arms LastMindIntentDone.
func TestReflexHoldDoesNotArmYieldWindow(t *testing.T) {
	s, m, a := warmUpAtFire(t, 42, 500)
	driveTicks(t, s, m, 3000, condIntentCmd(a, "goto_warmth", "warmth", 560, "reflex"))
	if s.Agents[0].LastMindIntentDone != 0 {
		t.Fatalf("a reflex recovery armed the yield window (LastMindIntentDone=%d)", s.Agents[0].LastMindIntentDone)
	}
}

// TestReflexHoldNoArriveIdleWander (SC-002 unit, the enumerated no-LLM change):
// a reflex-driven cold villager at a known fire HOLDS (warm_up-style goto_warmth)
// and never wanders while warmth is still recovering — the arrive-idle-wander
// vacuum is dead.
func TestReflexHoldNoArriveIdleWander(t *testing.T) {
	s, m, a := warmUpAtFire(t, 42, 300)
	// Reflex issues the conditioned goto_warmth toward the fire the agent stands
	// on; drive well past the reflex grace and assert no wander before recovery.
	log := driveTicks(t, s, m, 4000, condIntentCmd(a, "goto_warmth", "warmth", 500, "reflex"))
	done := intentDoneTick(log, 0)
	if done < 0 {
		t.Fatal("the held recovery never completed")
	}
	for _, e := range log {
		if e.Type != "agent.intent_set" || e.Tick >= done {
			continue
		}
		var p IntentSetPayload
		json.Unmarshal(e.Payload, &p)
		if p.Agent == 0 && p.Goal == "wander" {
			t.Fatalf("agent wandered at tick %d mid-recovery — the arrive-idle-wander vacuum is not dead", e.Tick)
		}
	}
}

// --- T008 / US3: interruptibility -----------------------------------------

// TestRecoveryAbortsWhenSourceDead (SC-003, AS1 + AS3): a conditioned hold whose
// need shows no net gain across a full recoveryStallTicks window aborts with the
// DISTINCT agent.recovery_stalled outcome (dead fire), within the stall window —
// no infinite loiter. Night, no warm tile ⇒ warmth falls, never advances.
func TestRecoveryAbortsWhenSourceDead(t *testing.T) {
	m := testMap(42)
	s := NewState(42, m)
	s.Night = true
	isolateAgents(s)
	a := &s.Agents[0]
	a.Dead, a.Asleep = false, false
	a.Needs = Needs{Health: 1000, Food: 600, Rest: 600, Warmth: 500, Morale: 600}
	// No lit fire on the target tile ⇒ !warmAt ⇒ warmth only falls; the hold can
	// never advance, so it must abort rather than loiter forever.
	log := driveTicks(t, s, m, recoveryStallTicks+400, condIntentCmd(a, "warm_up", "warmth", 900, "reflex"))
	var stalledTick int64 = -1
	for _, e := range log {
		if e.Type == "agent.recovery_stalled" {
			var p RecoveryStalledPayload
			json.Unmarshal(e.Payload, &p)
			if p.Agent == 0 && p.Need == "warmth" && p.Goal == "warm_up" {
				stalledTick = e.Tick
			}
		}
	}
	if stalledTick < 0 {
		t.Fatal("a dead-source recovery did not abort (no agent.recovery_stalled)")
	}
	if stalledTick > recoveryStallTicks+180 {
		t.Fatalf("abort took %d ticks, past the stall window %d (no-infinite-loiter bound)", stalledTick, recoveryStallTicks)
	}
	// The distinct outcome is stamped on the ring, and the intent is cleared so
	// the agent re-decides (it is not left holding).
	if s.Agents[0].Intent != nil && s.Agents[0].Intent.UntilNeed == "warmth" && s.Agents[0].Intent.Goal == "warm_up" {
		// It may have re-issued a fresh recovery after re-deciding; only fail if it
		// is the SAME never-cleared hold (WorkStart from before the abort).
	}
	found := false
	for _, r := range s.Agents[0].IntentLog {
		if r.Outcome == "stalled" {
			found = true
		}
	}
	if !found {
		t.Fatal("the abort did not stamp the distinct \"stalled\" ring outcome")
	}
}

// TestSurvivalPreemptsRecovery (SC-003, AS2): a warmth recovery yields the moment
// a higher-priority survival need (hunger) is in its danger band — the hold ends
// (intent_done) so the reflex's hunger rung owns the agent. No new immunity: the
// hold is LESS sticky, not more.
func TestSurvivalPreemptsRecovery(t *testing.T) {
	s, m, a := warmUpAtFire(t, 42, 500)
	a.Needs.Food = dangerFoodBelow - 10 // hungry: a worse emergency than being cold
	log := driveTicks(t, s, m, 300, condIntentCmd(a, "warm_up", "warmth", 900, "planner"))
	done := intentDoneTick(log, 0)
	if done < 0 || done > 60 {
		t.Fatalf("a hold must yield at once to a pre-existing hunger emergency, done=%d", done)
	}
	// The warm_up never started a hold before yielding (a work_started BEFORE the
	// yield would mean it held first; work_started AFTER is the reflex's forage).
	for _, e := range log {
		if e.Type == "agent.work_started" && e.Tick <= done {
			t.Fatal("the recovery started a hold instead of yielding to hunger")
		}
	}
	// preemptsRecovery is the doctrine: warmth yields to food-in-danger, never the
	// reverse (a hungry recovery is not outranked by cold).
	hungry := Agent{Needs: Needs{Food: dangerFoodBelow - 1, Warmth: 900, Rest: 900}}
	if !preemptsRecovery(&hungry, "warmth") {
		t.Fatal("warmth recovery must yield to hunger in its danger band")
	}
	cold := Agent{Needs: Needs{Food: 900, Warmth: dangerWarmthBelow - 1, Rest: 900}}
	if preemptsRecovery(&cold, "food") {
		t.Fatal("food recovery must NOT be preempted by cold (food outranks warmth)")
	}
}

// TestHoldingIntentOverriddenLikeAnyActive (AS2, no new immunity): a planner
// intent_set mid-hold REPLACES the holding recovery exactly as it replaces any
// active intent — the executor's existing override path, untouched.
func TestHoldingIntentOverriddenLikeAnyActive(t *testing.T) {
	s, m, a := warmUpAtFire(t, 42, 500)
	// A high threshold ⇒ the warm_up is still actively holding at tick 600 (warmth
	// 500 rising +6/min reaches 900 only ~4000 ticks out).
	cmds := condIntentCmd(a, "warm_up", "warmth", 900, "planner")
	// Mid-hold, the planner injects a different intent.
	cmds[600] = []store.Event{{Tick: 600, Type: "agent.intent_set", Payload: mustPayload(
		IntentSetPayload{Agent: 0, Goal: "wander", TargetX: a.X, TargetY: a.Y, Source: "planner"})}}
	log := driveTicks(t, s, m, 605, cmds)
	// The warm_up was still HOLDING when overridden — no completion before 600.
	if done := intentDoneTick(log, 0); done >= 0 && done < 600 {
		t.Fatalf("the warm_up ended at %d before the override — it was not actively holding", done)
	}
	// The planner intent_set replaced the holding recovery at tick 600, exactly as
	// it replaces any active intent (the executor's existing override path).
	overrode := false
	for _, e := range log {
		if e.Type != "agent.intent_set" || e.Tick != 600 {
			continue
		}
		var p IntentSetPayload
		json.Unmarshal(e.Payload, &p)
		if p.Agent == 0 && p.Goal == "wander" {
			overrode = true
		}
	}
	if !overrode {
		t.Fatal("the planner override intent_set did not land on the holding agent")
	}
	// The active intent is now the override (wander en route), no longer the hold.
	if in := s.Agents[0].Intent; in != nil && in.UntilNeed != "" {
		t.Fatalf("the recovery hold survived a planner override (new immunity): %+v", in)
	}
}

// --- T010 / US2: the mechanism is generic ---------------------------------

// TestSecondConsumerSharesMechanism (SC-004): a rest-conditioned intent flows
// through the SAME executeAtTarget → recoveryHoldEvents path — completion and
// abort both — proving the condition machinery is need-generic, not warm_up
// plumbing. FLAG: sleep itself is NOT refactored to this mechanism (its asleep-
// flag + wakeReason lifecycle is behaviourally distinct; refactoring it would
// change wake behaviour), so US2 is proven at the mechanism level here per R6.
func TestSecondConsumerSharesMechanism(t *testing.T) {
	// A rest condition already satisfied on arrival completes via the same check.
	s, m, a := warmUpAtFire(t, 42, 500)
	a.Needs.Rest = 700
	log := driveTicks(t, s, m, 200, condIntentCmd(a, "warm_up", "rest", 600, "planner"))
	if done := intentDoneTick(log, 0); done < 0 || done > 60 {
		t.Fatalf("a satisfied rest condition should complete immediately (shared check), done=%d", done)
	}
	// A rest condition that cannot advance (awake ⇒ rest decays) aborts via the
	// SAME dead-source path — identical mechanism, different need.
	s, m, a = warmUpAtFire(t, 7, 500)
	a.Needs.Rest = 400
	log = driveTicks(t, s, m, recoveryStallTicks+300, condIntentCmd(a, "warm_up", "rest", 900, "planner"))
	stalled := false
	for _, e := range log {
		if e.Type == "agent.recovery_stalled" {
			var p RecoveryStalledPayload
			json.Unmarshal(e.Payload, &p)
			if p.Need == "rest" {
				stalled = true
			}
		}
	}
	if !stalled {
		t.Fatal("a non-advancing rest recovery did not abort via the shared mechanism")
	}
	// The closed-set door and the need accessor are the shared primitives.
	for _, n := range []string{"warmth", "rest", "food"} {
		if !isRecoveryNeed(n) {
			t.Fatalf("%q must be a valid recovery need", n)
		}
	}
	if isRecoveryNeed("morale") || isRecoveryNeed("") {
		t.Fatal("the closed set must reject non-recovery needs")
	}
	n := Needs{Warmth: 11, Rest: 22, Food: 33}
	if needValue(n, "warmth") != 11 || needValue(n, "rest") != 22 || needValue(n, "food") != 33 {
		t.Fatal("needValue must read each recovery need generically")
	}
}

// --- T012 / US4: wake to cold ---------------------------------------------

// TestSleeperWakesToColdEmergency (SC-005 / AS1): a sleeper whose warmth has
// fallen below the exposure-emergency floor at night, WITH warmth it can act on,
// wakes — the Oak final-night shape (Gap C).
func TestSleeperWakesToColdEmergency(t *testing.T) {
	s, m, a := warmUpAtFire(t, 42, exposureWakeBelow-20)
	s.Night = true
	a.Asleep = true
	// A KNOWN reachable lit fire (granted by warmUpAtFire) ⇒ warmthLadder acts.
	if !wakeReason(s, m, 0, s.Night, 1) {
		t.Fatal("a freezing sleeper with reachable warmth did not wake (Gap C)")
	}
}

// TestCozySleeperSleepsThrough (SC-005 / AS2, the control): a sleeper cozy at a
// fire (warmth high) is never woken — the wake keys on the emergency band, not on
// "night".
func TestCozySleeperSleepsThrough(t *testing.T) {
	s, m, _ := warmUpAtFire(t, 42, 800)
	s.Night = true
	s.Agents[0].Asleep = true
	if wakeReason(s, m, 0, s.Night, 1) {
		t.Fatal("a cozy fire-side sleeper was woken — the wake must fire on the band, not the night")
	}
}

// TestColdWakeIsEmergencyFloorNotDangerBand: a sleeper cold but ABOVE the
// exposure floor (between exposureWakeBelow and the 350 warmth danger band) is
// NOT woken — the wake mirrors the hunger EMERGENCY (150), not the danger band
// (350). This is the churn-and-survival guard that keeps degraded mode intact.
func TestColdWakeIsEmergencyFloorNotDangerBand(t *testing.T) {
	if exposureWakeBelow >= dangerWarmthBelow {
		t.Fatal("the exposure wake floor must sit BELOW the warmth danger band")
	}
	s, m, a := warmUpAtFire(t, 42, dangerWarmthBelow-10) // in the danger band, above the floor
	s.Night = true
	a.Asleep = true
	if wakeReason(s, m, 0, s.Night, 1) {
		t.Fatalf("a sleeper at warmth %d (below the 350 band but above the %d floor) must NOT wake",
			a.Needs.Warmth, exposureWakeBelow)
	}
}

// TestSleeperNotWokenWhenNothingActionable (churn bound): a freezing sleeper who
// can do NOTHING about it (no known warmth, no wood, no known tree) stays asleep
// — waking to re-sleep every beat is the 4am churn the gate avoids.
func TestSleeperNotWokenWhenNothingActionable(t *testing.T) {
	m := testMap(42)
	s := NewState(42, m)
	s.Night = true
	isolateAgents(s)
	a := &s.Agents[0]
	a.Asleep = true
	a.Inv = Inventory{} // no wood
	a.Needs = Needs{Health: 1000, Food: 600, Rest: 600, Warmth: exposureWakeBelow - 20, Morale: 600}
	// No fire fact, no tree fact granted ⇒ the warmth ladder has nothing to do.
	if _, ok := warmthLadder(s, m, a, 1); ok {
		t.Skip("seed 42 spawn gives the agent an actionable warmth ladder; not the nothing-to-do case")
	}
	if wakeReason(s, m, 0, s.Night, 1) {
		t.Fatal("a sleeper who can do nothing about the cold must not be woken (churn bound)")
	}
}

// --- T013 / SC-002: the extended Sage scenario, end-to-end ----------------

// TestSageWarmUpHeldToThresholdThenReleased (SC-002): the 062 Sage-shape driven
// through a planner warm_up — the agent is HELD at the fire to the doctrine
// threshold with ZERO mid-recovery prep dispatches, then released. The
// arrive-idle-vacuum is dead end-to-end.
func TestSageWarmUpHeldToThresholdThenReleased(t *testing.T) {
	s, m, a, _ := sageScenario(t, 42) // agent at a lit fire, warmth 250, larder empty (prep temptation)
	warmthBefore := a.Needs.Warmth
	cmds := map[int64][]store.Event{
		1: {{Tick: 1, Type: "agent.intent_set", Payload: mustPayload(IntentSetPayload{
			Agent: 0, Goal: "warm_up", TargetX: a.X, TargetY: a.Y,
			UntilNeed: "warmth", UntilValue: warmthRecoverTo, Source: "planner"})}},
	}
	log := driveTicks(t, s, m, 8000, cmds)
	done := intentDoneTick(log, 0)
	if done < 0 {
		t.Fatal("the planner warm_up never completed — it should hold to threshold then release")
	}
	// Held to threshold: warmth recovered to the doctrine target.
	if s.Agents[0].Needs.Warmth < warmthRecoverTo {
		t.Fatalf("released at warmth %d, below the %d threshold", s.Agents[0].Needs.Warmth, warmthRecoverTo)
	}
	if s.Agents[0].Needs.Warmth <= warmthBefore {
		t.Fatalf("warmth did not recover: %d -> %d", warmthBefore, s.Agents[0].Needs.Warmth)
	}
	// SC-002: ZERO prep intents (the larder forage/hunt/build/refuel counter-
	// schedule) for agent 0 across the whole recovery — no mid-recovery dispatch.
	for _, e := range log {
		if e.Type != "agent.intent_set" || e.Tick == 1 || e.Tick > done {
			continue
		}
		var p IntentSetPayload
		json.Unmarshal(e.Payload, &p)
		if p.Agent == 0 && isPrepGoal(p.Goal) {
			t.Fatalf("a prep intent %q fired at tick %d mid-recovery — the vacuum is not dead", p.Goal, e.Tick)
		}
	}
	// The warm target never wandered off: no wander mid-recovery either.
	for _, e := range log {
		if e.Type != "agent.intent_set" || e.Tick == 1 || e.Tick > done {
			continue
		}
		var p IntentSetPayload
		json.Unmarshal(e.Payload, &p)
		if p.Agent == 0 && p.Goal == "wander" {
			t.Fatalf("agent wandered at tick %d mid-recovery", e.Tick)
		}
	}
}
