package mind

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/evanstern/promptworld/internal/cognition"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// TestConvoRawTruncation (TASK-42 T004): oversized raw replies are cut on a rune
// boundary with a marker, stay ≤ cap, and remain valid UTF-8; small replies
// pass through untouched.
func TestConvoRawTruncation(t *testing.T) {
	if got := truncateRaw("short reply"); got != "short reply" {
		t.Errorf("small reply mutated: %q", got)
	}
	// A reply of multi-byte runes (é = 2 bytes) longer than the cap must cut
	// mid-string without splitting a rune.
	big := strings.Repeat("é", rawReplyCap) // 2*cap bytes
	got := truncateRaw(big)
	if len(got) > rawReplyCap {
		t.Errorf("truncated length %d exceeds cap %d", len(got), rawReplyCap)
	}
	if !strings.HasSuffix(got, rawTruncMarker) {
		t.Errorf("missing truncation marker: %q", got[len(got)-20:])
	}
	if !utf8.ValidString(got) {
		t.Error("truncation split a rune (invalid UTF-8)")
	}
	// Exactly at the cap: no truncation.
	exact := strings.Repeat("a", rawReplyCap)
	if got := truncateRaw(exact); got != exact {
		t.Error("reply at the cap should not be truncated")
	}
}

// TestPlannerTelemetryLanded (US1): a successful planner thought leaves a
// cog.thought and exactly one landed cog.outcome sharing its job id, with
// the prediction stamped at snapshot time.
func TestPlannerTelemetryLanded(t *testing.T) {
	h := newHarness(t, `{"goal":"forage","reason":"hungry"}`)
	thoughts := h.waitEvents(t, 20*time.Second, func(e store.Event) bool {
		if e.Type != "cog.thought" {
			return false
		}
		var p sim.CogThoughtPayload
		return json.Unmarshal(e.Payload, &p) == nil && p.Class == "planner"
	})
	if len(thoughts) == 0 {
		t.Fatal("no planner cog.thought recorded")
	}
	var tp sim.CogThoughtPayload
	if err := json.Unmarshal(thoughts[0].Payload, &tp); err != nil {
		t.Fatal(err)
	}
	if tp.Job == "" || tp.Points != 3 || tp.PredictedWallMs <= 0 {
		t.Errorf("thought payload incomplete: %+v", tp)
	}
	outcomes := h.waitEvents(t, 20*time.Second, func(e store.Event) bool {
		if e.Type != "cog.outcome" {
			return false
		}
		var p sim.CogOutcomePayload
		return json.Unmarshal(e.Payload, &p) == nil && p.Job == tp.Job
	})
	if len(outcomes) != 1 {
		t.Fatalf("job %s has %d outcomes, want exactly 1", tp.Job, len(outcomes))
	}
	var op sim.CogOutcomePayload
	if err := json.Unmarshal(outcomes[0].Payload, &op); err != nil {
		t.Fatal(err)
	}
	if op.Outcome != sim.OutcomeLanded {
		t.Errorf("outcome = %q, want landed (reason %q)", op.Outcome, op.Reason)
	}
	if op.SnapshotTick != tp.SnapshotTick {
		t.Errorf("outcome snapshot %d != thought snapshot %d", op.SnapshotTick, tp.SnapshotTick)
	}
}

// TestPlannerTelemetryUnusable (US1): garbage output still terminates in a
// recorded outcome — silent failure is gone (FR-015).
func TestPlannerTelemetryUnusable(t *testing.T) {
	h := newHarness(t, "I simply cannot decide!!")
	outcomes := h.waitEvents(t, 20*time.Second, func(e store.Event) bool {
		if e.Type != "cog.outcome" {
			return false
		}
		var p sim.CogOutcomePayload
		return json.Unmarshal(e.Payload, &p) == nil &&
			p.Class == "planner" && p.Outcome == sim.OutcomeUnusable
	})
	if len(outcomes) == 0 {
		t.Fatal("garbage planner reply left no recorded outcome")
	}
	var p sim.CogOutcomePayload
	json.Unmarshal(outcomes[0].Payload, &p)
	if p.Reason == "" {
		t.Error("unusable outcome carries no reason")
	}
}

// TestPlannerSuppressedAtHighSpeed (US2): at 32x under bootstrap calibration
// (20 s/pt), a planner thought's predicted drift (1920 ticks) exceeds its
// budget (1200) — no model call is made, the reflex floor covers, and the
// suppression is recorded with its arithmetic.
func TestPlannerSuppressedAtHighSpeed(t *testing.T) {
	h := newHarnessAt(t, `{"goal":"forage","reason":"hungry"}`, "32x")

	suppressed := h.waitEvents(t, 30*time.Second, func(e store.Event) bool {
		if e.Type != "cog.outcome" {
			return false
		}
		var p sim.CogOutcomePayload
		return json.Unmarshal(e.Payload, &p) == nil &&
			p.Class == "planner" && p.Outcome == sim.OutcomeSuppressed
	})
	if len(suppressed) == 0 {
		t.Fatal("no planner suppression recorded at 32x")
	}
	var p sim.CogOutcomePayload
	json.Unmarshal(suppressed[0].Payload, &p)
	if !strings.Contains(p.Reason, "> budget") {
		t.Errorf("suppression reason lacks arithmetic: %q", p.Reason)
	}
	h.model.mu.Lock()
	for _, k := range h.model.kinds {
		if k == llm.KindPlanner {
			t.Error("a planner model call was made despite suppression")
		}
	}
	h.model.mu.Unlock()
}

// TestRecalibrateSignalEmitsAdoption (spec 031 US3, FR-005): the drift hook
// records one cog.recalibration_recommended event carrying the breaching
// provider's name, spike rate, window, and the adoption arithmetic (prior →
// adopted), with the emission-time estimate equal to the adopted value.
func TestRecalibrateSignalEmitsAdoption(t *testing.T) {
	h := newHarnessAt(t, `{"goal":"wander","reason":"stretching"}`, "16x")
	h.mind.RecalibrateSignal("gemma", 11.8, 1.0, 0.524, 11.8)
	evs := h.waitEvents(t, 10*time.Second, func(e store.Event) bool {
		if e.Type != "cog.recalibration_recommended" {
			return false
		}
		var p sim.RecalibrationPayload
		return json.Unmarshal(e.Payload, &p) == nil && p.Tier == "gemma"
	})
	if len(evs) != 1 {
		t.Fatalf("emitted %d recalibration events for the driven breach, want exactly 1", len(evs))
	}
	var p sim.RecalibrationPayload
	if err := json.Unmarshal(evs[0].Payload, &p); err != nil {
		t.Fatal(err)
	}
	if p.SpikeRate != 1.0 || p.Window != cognition.WindowSize {
		t.Errorf("core fields: spike=%g window=%d", p.SpikeRate, p.Window)
	}
	if p.PriorSPerPt != 0.524 || p.AdoptedSPerPt != 11.8 || p.EstimateSPerPt != 11.8 {
		t.Errorf("adoption arithmetic: prior=%g adopted=%g est=%g (est must equal adopted)",
			p.PriorSPerPt, p.AdoptedSPerPt, p.EstimateSPerPt)
	}
}

// TestFutureDatedLine (US4): the helper states now and the landing estimate;
// no line when there is no meaningful prediction.
func TestFutureDatedLine(t *testing.T) {
	line := futureDated(0, 1800)
	if !strings.Contains(line, "day 1 06:00") || !strings.Contains(line, "day 1 06:30") {
		t.Errorf("future-dated line: %q", line)
	}
	if futureDated(1800, 1800) != "" || futureDated(1800, 0) != "" {
		t.Error("no-prediction cases must be empty")
	}
}

// TestPlanFormLandsAndExecutes (US4 integration): a plan reply parses, lands
// through the door as agent.plan_set, and the executor fires the steps with
// Source "plan" — no model at firing time.
func TestPlanFormLandsAndExecutes(t *testing.T) {
	h := newHarness(t, `{"plan":[{"goal":"wander"},{"goal":"forage","for_min":120}],"reason":"stretch then gather"}`)
	planSets := h.waitEvents(t, 20*time.Second, func(e store.Event) bool {
		return e.Type == "agent.plan_set"
	})
	if len(planSets) == 0 {
		t.Fatal("no plan landed")
	}
	started := h.waitEvents(t, 20*time.Second, func(e store.Event) bool {
		return e.Type == "agent.plan_step_started"
	})
	if len(started) == 0 {
		t.Fatal("plan never started executing")
	}
	intents := h.waitEvents(t, 20*time.Second, func(e store.Event) bool {
		if e.Type != "agent.intent_set" {
			return false
		}
		var p sim.IntentSetPayload
		return json.Unmarshal(e.Payload, &p) == nil && p.Source == "plan"
	})
	if len(intents) == 0 {
		t.Fatal("no plan-sourced intents executed")
	}
}

// --- US5: pause semantics — world freezes, minds catch up (FR-018) ---

// TestPauseInFlightThoughtLandsAtFrozenTick: a planner call in flight when
// the world pauses completes on the wall clock and lands at the frozen tick;
// the wall time spent paused adds zero game-tick staleness.
func TestPauseInFlightThoughtLandsAtFrozenTick(t *testing.T) {
	h := newHarnessAt(t, `{"goal":"wander","reason":"stretching"}`, "16x")
	gate := make(chan struct{})
	h.model.mu.Lock()
	h.model.planGate = gate
	h.model.mu.Unlock()

	// Wait for a planner call to be in flight (blocked on the gate).
	deadline := time.Now().Add(30 * time.Second)
	for h.model.calls.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if h.model.calls.Load() == 0 {
		t.Fatal("no planner call started")
	}
	st, err := h.loop.Do("pause", "")
	if err != nil {
		t.Fatal(err)
	}
	frozen := st.Tick
	time.Sleep(1500 * time.Millisecond) // wall time passes; ticks must not
	close(gate)                         // the mind finishes thinking mid-pause

	outcomes := h.waitEvents(t, 20*time.Second, func(e store.Event) bool {
		if e.Type != "cog.outcome" {
			return false
		}
		var p sim.CogOutcomePayload
		return json.Unmarshal(e.Payload, &p) == nil && p.Class == "planner" &&
			(p.Outcome == sim.OutcomeLanded || p.Outcome == sim.OutcomeAdapted)
	})
	if len(outcomes) == 0 {
		t.Fatal("in-flight thought never landed during pause")
	}
	if outcomes[0].Tick != frozen {
		t.Errorf("landed at tick %d, world frozen at %d", outcomes[0].Tick, frozen)
	}
	var p sim.CogOutcomePayload
	json.Unmarshal(outcomes[0].Payload, &p)
	if p.LandingTick != frozen {
		t.Errorf("landing_tick %d != frozen %d", p.LandingTick, frozen)
	}
	if p.StalenessTicks > frozen-p.SnapshotTick {
		t.Errorf("pause accrued staleness: %d > %d", p.StalenessTicks, frozen-p.SnapshotTick)
	}
}

// TestPauseStartsNoNewThoughts: scheduling is tick-driven — once a paused
// world quiesces, no new planner jobs start no matter how much wall time
// passes. (A landing batch arriving mid-pause may first settle one
// debounce-bounded catch-up round at zero staleness — FR-018 as refined by
// the live validation run; this test drains before measuring.)
func TestPauseStartsNoNewThoughts(t *testing.T) {
	h := newHarnessAt(t, `{"goal":"wander","reason":"stretching"}`, "16x")
	if _, err := h.loop.Do("pause", ""); err != nil {
		t.Fatal(err)
	}
	// Drain: give any pre-pause in-flight work a moment to finish.
	time.Sleep(1 * time.Second)
	before := h.model.calls.Load()
	time.Sleep(2 * time.Second)
	if after := h.model.calls.Load(); after != before {
		t.Errorf("model called %d times while paused", after-before)
	}
}

// TestPauseConversationLandsAtFrozenTick: a scene founded before the pause
// completes on the wall clock and lands atomically at the frozen tick.
func TestPauseConversationLandsAtFrozenTick(t *testing.T) {
	model := &scriptedModel{replies: convoScript(
		`{"gist": "talked shelter", "tone_a": 1, "tone_b": 1}`)}
	h, md := setupConvo(t, model)
	st, err := h.loop.Do("pause", "")
	if err != nil {
		t.Fatal(err)
	}
	frozen := st.Tick
	md.maybeStartConversation(store.Event{
		Tick: frozen, Type: "agent.talked",
		Payload: mustJSON(t, sim.TalkedPayload{A: sim.Ref(0), B: sim.Ref(1)}),
	}, 0)
	convs := h.waitEvents(t, 15*time.Second, func(e store.Event) bool {
		return e.Type == "social.conversation"
	})
	if len(convs) == 0 {
		t.Fatal("scene never landed during pause")
	}
	if convs[0].Tick != frozen {
		t.Errorf("scene landed at tick %d, world frozen at %d", convs[0].Tick, frozen)
	}
}

// TestResumeNoBurst: pause accrues no cognition debt — after resume, thought
// volume is cadence-normal, not a compensating flood.
func TestResumeNoBurst(t *testing.T) {
	// A real (finite) speed: at 16x, 2 wall-seconds after resume is only 32
	// game-ticks — far under one 1800-tick planner cadence — so anything
	// beyond stragglers IS a burst. (At max speed the same window spans
	// dozens of cadences and high volume is legitimate.)
	h := newHarnessAt(t, `{"goal":"wander","reason":"stretching"}`, "16x")
	// Let the world think a little, then pause for real wall time.
	h.waitEvents(t, 30*time.Second, func(e store.Event) bool { return e.Type == "cog.thought" })
	if _, err := h.loop.Do("pause", ""); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second)
	before := h.model.calls.Load()
	if _, err := h.loop.Do("resume", ""); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second)
	burst := h.model.calls.Load() - before
	if burst > int64(2*sim.AgentCount) {
		t.Errorf("resume burst: %d calls in 2s", burst)
	}
}

// --- spec 040 US1: a nudge wakes the nudged villager while paused (TASK-77) ---

// nudgeBatchEvents builds the landed-nudge batch the mind absorbs — the shape
// internal/guardian/turn.go's landNudgeBatch commits: one metatron.nudged
// carrying the arming Seq, then one prefixed agent.memory_added per target. Fed
// straight to the mind's Observe path: the mind consumes an ALREADY-landed nudge
// (the door's charge/night validation is upstream — spec 040 D1), so it arms
// every Target regardless of the reducer's form gates.
func nudgeBatchEvents(seq, tick int64, form, text string, targets ...int) []store.Event {
	nb, _ := json.Marshal(sim.GuardianNudgedPayload{Form: form, Targets: targets, Text: text})
	batch := []store.Event{{Seq: seq, Tick: tick, Type: "metatron.nudged", Payload: nb}}
	prefix := "You saw a vision: "
	if form == "omen" {
		prefix = "You witnessed an omen: "
	}
	for i, tgt := range targets {
		mb, _ := json.Marshal(sim.MemoryAddedPayload{
			Agent: sim.Ref(tgt), Text: prefix + text, Salience: sim.SalDream, Subject: sim.Ref(-1), Origin: sim.OriginOmen})
		batch = append(batch, store.Event{Seq: seq + int64(i) + 1, Tick: tick, Type: "agent.memory_added", Payload: mb})
	}
	return batch
}

// unplannedAt returns up to n agent indices with NO cog.thought in the store —
// villagers whose planner debounce window is provably open at a frozen tick
// ≥ planDebounceTicks. The paused-nudge tests pick their targets from this set
// instead of hardcoding high-index agents: which villagers get armed (by
// completions/encounters) into the very first tick-300 planner batch is
// EMERGENT world behavior, not a stable fixture premise — spec 041's
// knowledge-gated reflex, for one, changed it (early nearby completions arm
// high-index agents pre-freeze, closing their debounce for the whole pause,
// which is exactly the spec-040 "debounce is the only bound" doctrine).
func unplannedAt(t *testing.T, h *harness, n int) []int {
	t.Helper()
	evs, err := h.st.EventsSince(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	planned := map[int]bool{}
	for _, e := range evs {
		if e.Type != "cog.thought" {
			continue
		}
		var p sim.CogThoughtPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			planned[p.Agent] = true
		}
	}
	var out []int
	for i := sim.AgentCount - 1; i >= 0 && len(out) < n; i-- {
		if !planned[i] {
			out = append(out, i)
		}
	}
	if len(out) < n {
		t.Skipf("only %d villagers unplanned at freeze, need %d (whole village planned in the warm-up batch)", len(out), n)
	}
	return out
}

// TestPausedNudgeWakesTargetOnce (spec 040 US1, FR-001/FR-002; contracts C2/C3):
// while paused, a landed nudge arms the target for exactly one frozen-tick
// planner round whose trigger_seq is the nudge event's Seq and whose outcome
// lands at zero staleness; a second nudge in the same pause buys no second
// thought — the 300-game-tick planner debounce cannot reopen while the clock is
// frozen (the bound is by construction, not a counter).
func TestPausedNudgeWakesTargetOnce(t *testing.T) {
	h := newHarnessAt(t, `{"goal":"wander","reason":"stretching"}`, "16x")

	// Warm past the debounce floor: agent 0 plans first at tick 300, proving the
	// world is warm; high-index agents (nextDue ≥ 900) have never planned, so
	// their debounce window is open at any frozen tick ≥ 300.
	if got := h.waitEvents(t, 40*time.Second, func(e store.Event) bool {
		return e.Type == "cog.thought"
	}); len(got) == 0 {
		t.Fatal("world never warmed to a first planner thought")
	}
	st, err := h.loop.Do("pause", "")
	if err != nil {
		t.Fatal(err)
	}
	frozen := st.Tick
	time.Sleep(1 * time.Second) // drain any in-flight pre-pause cognition

	target := unplannedAt(t, h, 1)[0] // debounce provably open at the frozen tick
	const nudgeSeq = int64(900001)
	h.mind.Observe(nudgeBatchEvents(nudgeSeq, frozen, "vision", "the river is rising", target))

	thoughts := h.waitEvents(t, 20*time.Second, func(e store.Event) bool {
		if e.Type != "cog.thought" {
			return false
		}
		var p sim.CogThoughtPayload
		return json.Unmarshal(e.Payload, &p) == nil && p.Class == "planner" &&
			p.Agent == target && p.TriggerSeq == nudgeSeq
	})
	if len(thoughts) != 1 {
		t.Fatalf("nudge produced %d planner thoughts for the target, want exactly 1", len(thoughts))
	}
	var tp sim.CogThoughtPayload
	json.Unmarshal(thoughts[0].Payload, &tp)
	if tp.SnapshotTick != frozen {
		t.Errorf("thought snapshot tick %d != frozen %d", tp.SnapshotTick, frozen)
	}

	outcomes := h.waitEvents(t, 20*time.Second, func(e store.Event) bool {
		if e.Type != "cog.outcome" {
			return false
		}
		var p sim.CogOutcomePayload
		return json.Unmarshal(e.Payload, &p) == nil && p.Job == tp.Job
	})
	if len(outcomes) != 1 {
		t.Fatalf("nudge thought %s has %d outcomes, want exactly 1", tp.Job, len(outcomes))
	}
	var op sim.CogOutcomePayload
	json.Unmarshal(outcomes[0].Payload, &op)
	if op.LandingTick != frozen || op.StalenessTicks != 0 {
		t.Errorf("frozen-tick landing wrong: landing=%d staleness=%d (frozen %d)", op.LandingTick, op.StalenessTicks, frozen)
	}

	// A second nudge to the same villager while still frozen arms pending again,
	// but the debounce cannot reopen (game time is frozen), so no second thought.
	const nudge2 = int64(900002)
	h.mind.Observe(nudgeBatchEvents(nudge2, frozen, "vision", "and rising still", target))
	if second := h.waitEvents(t, 3*time.Second, func(e store.Event) bool {
		if e.Type != "cog.thought" {
			return false
		}
		var p sim.CogThoughtPayload
		return json.Unmarshal(e.Payload, &p) == nil && p.TriggerSeq == nudge2
	}); len(second) != 0 {
		t.Fatalf("second nudge in the same pause produced %d thoughts, want 0 (debounce bound)", len(second))
	}
}

// TestPausedOmenArmsOnlyTargets (spec 040 US1, scenario 3): a paused multi-target
// omen gives each living, awake target its own single frozen-tick round; a
// villager the omen did not target is never armed.
func TestPausedOmenArmsOnlyTargets(t *testing.T) {
	h := newHarnessAt(t, `{"goal":"wander","reason":"stretching"}`, "16x")
	if got := h.waitEvents(t, 40*time.Second, func(e store.Event) bool {
		return e.Type == "cog.thought"
	}); len(got) == 0 {
		t.Fatal("world never warmed to a first planner thought")
	}
	st, err := h.loop.Do("pause", "")
	if err != nil {
		t.Fatal(err)
	}
	frozen := st.Tick
	time.Sleep(1 * time.Second)

	// Two villagers with provably open debounce windows at the freeze (never
	// planned); any villager outside the target set is the untargeted control.
	targets := unplannedAt(t, h, 2)
	control := 0
	for control == targets[0] || control == targets[1] {
		control++
	}
	const omenSeq = int64(910000)
	h.mind.Observe(nudgeBatchEvents(omenSeq, frozen, "omen", "the long night comes", targets[0], targets[1]))

	for _, tgt := range targets {
		got := h.waitEvents(t, 20*time.Second, func(e store.Event) bool {
			if e.Type != "cog.thought" {
				return false
			}
			var p sim.CogThoughtPayload
			return json.Unmarshal(e.Payload, &p) == nil && p.Agent == tgt && p.TriggerSeq == omenSeq
		})
		if len(got) != 1 {
			t.Fatalf("omen target %d got %d thoughts, want exactly 1", tgt, len(got))
		}
	}
	if bystander := h.waitEvents(t, 2*time.Second, func(e store.Event) bool {
		if e.Type != "cog.thought" {
			return false
		}
		var p sim.CogThoughtPayload
		return json.Unmarshal(e.Payload, &p) == nil && p.Agent == control && p.TriggerSeq == omenSeq
	}); len(bystander) != 0 {
		t.Fatalf("untargeted villager %d was armed by the omen (%d thoughts)", control, len(bystander))
	}
}

// --- spec 040 US2: paused routing tells the truth (TASK-77) ---

// TestRouteVerdictPausedAllowsAtSuppressingSpeed (spec 040 US2, FR-004/FR-005):
// a paused replica routes at zero drift (allow) with the C1 arithmetic naming
// the paused state — even at a SET speed that suppresses while running, and at
// uncapped max where the paused branch must precede the tps<=0 branch (scenario
// 3). Running at the same speed is byte-identical to cognition.Route.
func TestRouteVerdictPausedAllowsAtSuppressingSpeed(t *testing.T) {
	state := sim.NewState(42, worldmap.Generate(42, 64, 64))
	md := &Mind{orch: &mockModel{}, replica: state}
	dc, _ := cognition.ClassFor("planner")
	spp := md.secondsPerPoint(llm.KindPlanner)

	// Paused from a suppressing set speed (32x: 3pt×20×32 = 1920 > 1200): paused
	// wins — allow at zero drift, C1 arithmetic.
	state.Paused = true
	state.Speed = "32x"
	if v := md.routeVerdict("planner", llm.KindPlanner); v != cognition.RoutePaused(dc, spp) {
		t.Errorf("paused @32x verdict\n got %+v\nwant RoutePaused %+v", v, cognition.RoutePaused(dc, spp))
	} else if !v.Allow || v.PredictedDriftTicks != 0 || !strings.Contains(v.Arithmetic, "paused") {
		t.Errorf("paused @32x: allow=%v drift=%d arith=%q", v.Allow, v.PredictedDriftTicks, v.Arithmetic)
	}

	// Uncapped max while paused: the paused branch precedes tps<=0, so paused
	// still wins (a frozen world does not drift, whatever the set speed).
	state.Speed = "max"
	if v := md.routeVerdict("planner", llm.KindPlanner); v != cognition.RoutePaused(dc, spp) {
		t.Errorf("paused @uncapped verdict %+v, want RoutePaused (paused wins at uncapped)", v)
	}

	// Running at the same suppressing speed is byte-identical to Route (FR-005).
	state.Paused = false
	state.Speed = "32x"
	if got, want := md.routeVerdict("planner", llm.KindPlanner), cognition.Route(dc, 32, spp); got != want {
		t.Errorf("running @32x verdict\n got %+v\nwant %+v (byte-identical to Route)", got, want)
	}
}

// TestPausedThoughtPredictsFrozenLanding (spec 040 US2, D3; contract C3): on a
// paused replica newMeta predicts the land tick at the snapshot tick (not the
// set-speed projection), so the FutureDated prompt prefix no-ops and the recorded
// cog.thought agrees the thought lands at the frozen tick — prompt, gate, and
// record never disagree.
func TestPausedThoughtPredictsFrozenLanding(t *testing.T) {
	state := sim.NewState(42, worldmap.Generate(42, 64, 64))
	state.Paused = true
	state.Speed = "32x" // a projection that would push the land tick forward while running
	md := &Mind{orch: &mockModel{}, replica: state}

	const snapshot = int64(5000)
	meta := md.newMeta("planner", 0, snapshot, 900001, llm.KindPlanner)
	if meta.predictedLandTick != snapshot {
		t.Errorf("paused predicted land tick %d, want snapshot %d", meta.predictedLandTick, snapshot)
	}
	if !meta.class.FutureDated {
		t.Fatal("planner must be FutureDated for this test to mean anything")
	}
	if pre := futureDated(snapshot, meta.predictedLandTick); pre != "" {
		t.Errorf("paused thought carried a future-dating prefix: %q", pre)
	}
	var tp sim.CogThoughtPayload
	json.Unmarshal(cogThoughtEvent(meta).Payload, &tp)
	if tp.PredictedLandTick != snapshot {
		t.Errorf("recorded predicted_land_tick %d, want %d (contract C3)", tp.PredictedLandTick, snapshot)
	}
}

// TestPausedNudgeThinksAtSuppressingSpeed (spec 040 US1+US2 together, the full
// decision-6 chain; SC-003): on a world paused from a planner-suppressing speed
// (32x), a landed nudge's thought is ATTEMPTED and LANDS at the frozen tick —
// the wake (US1) arms it and paused routing (US2) allows what set-speed routing
// would suppress. The nudged round is never suppressed while paused.
func TestPausedNudgeThinksAtSuppressingSpeed(t *testing.T) {
	h := newHarnessAt(t, `{"goal":"wander","reason":"stretching"}`, "32x")

	// Warm until 32x routing is live (agent 0 suppressed at tick 300): under
	// suppression lastPlanned is never set, so every agent keeps an open window.
	if got := h.waitEvents(t, 40*time.Second, func(e store.Event) bool {
		if e.Type != "cog.outcome" {
			return false
		}
		var p sim.CogOutcomePayload
		return json.Unmarshal(e.Payload, &p) == nil && p.Class == "planner" && p.Outcome == sim.OutcomeSuppressed
	}); len(got) == 0 {
		t.Fatal("no planner suppression at 32x — world never warmed")
	}
	st, err := h.loop.Do("pause", "")
	if err != nil {
		t.Fatal(err)
	}
	frozen := st.Tick
	time.Sleep(1 * time.Second)

	const target = 7
	const nudgeSeq = int64(920000)
	h.mind.Observe(nudgeBatchEvents(nudgeSeq, frozen, "vision", "the well ran dry", target))

	thoughts := h.waitEvents(t, 20*time.Second, func(e store.Event) bool {
		if e.Type != "cog.thought" {
			return false
		}
		var p sim.CogThoughtPayload
		return json.Unmarshal(e.Payload, &p) == nil && p.Agent == target && p.TriggerSeq == nudgeSeq
	})
	if len(thoughts) != 1 {
		t.Fatalf("paused nudge at 32x produced %d thoughts, want exactly 1 (paused routing must allow what set-speed suppresses)", len(thoughts))
	}
	var tp sim.CogThoughtPayload
	json.Unmarshal(thoughts[0].Payload, &tp)

	outcomes := h.waitEvents(t, 20*time.Second, func(e store.Event) bool {
		if e.Type != "cog.outcome" {
			return false
		}
		var p sim.CogOutcomePayload
		return json.Unmarshal(e.Payload, &p) == nil && p.Job == tp.Job
	})
	if len(outcomes) != 1 {
		t.Fatalf("nudge job %s has %d outcomes, want exactly 1", tp.Job, len(outcomes))
	}
	var op sim.CogOutcomePayload
	json.Unmarshal(outcomes[0].Payload, &op)
	if op.Outcome == sim.OutcomeSuppressed {
		t.Fatalf("the paused nudge round was suppressed (%q) — paused routing must allow it", op.Reason)
	}
	if op.LandingTick != frozen || op.StalenessTicks != 0 {
		t.Errorf("frozen-tick landing wrong: landing=%d staleness=%d (frozen %d)", op.LandingTick, op.StalenessTicks, frozen)
	}
}

// --- spec 037 US2 (T007): the suppressionCounting seam ---

// countingOrch is a fake orchestrator implementing both Submitter and the
// optional suppressionCounting seam — the daemon orchestrator's shape for the
// counter hook, minus everything emitSuppressed doesn't touch.
type countingOrch struct {
	scriptedModel
	mu     sync.Mutex
	counts map[string]int
}

func (o *countingOrch) RecordSuppression(class string) {
	o.mu.Lock()
	if o.counts == nil {
		o.counts = map[string]int{}
	}
	o.counts[class]++
	o.mu.Unlock()
}

// TestEmitSuppressedReachesSeam: every class emitSuppressed records reaches the
// counting seam exactly once per call (watched and unwatched classes alike —
// the seam counts all, the wire filters). social is nil so the detached event
// emit no-ops; the count must still land (it is taken before the goroutine).
func TestEmitSuppressedReachesSeam(t *testing.T) {
	orch := &countingOrch{}
	md := &Mind{orch: orch, social: nil}
	classes := []string{"planner", "conversation", "meeting", "consolidation", "chronicle"}
	for _, c := range classes {
		md.emitSuppressed(c, 0, 100, cognition.Verdict{Class: c})
	}
	md.emitSuppressed("planner", 1, 200, cognition.Verdict{Class: "planner"}) // a second planner suppression
	orch.mu.Lock()
	defer orch.mu.Unlock()
	if orch.counts["planner"] != 2 {
		t.Errorf("planner reached the seam %d times, want 2", orch.counts["planner"])
	}
	for _, c := range []string{"conversation", "meeting", "consolidation", "chronicle"} {
		if orch.counts[c] != 1 {
			t.Errorf("%s reached the seam %d times, want 1", c, orch.counts[c])
		}
	}
}

// TestEmitSuppressedSeamlessOrchNoOp: an orchestrator lacking the seam (a bare
// Submitter) is a silent no-op — emitSuppressed neither panics nor blocks, so
// the absorb loop is never at risk when the counter is absent (e.g. test
// fakes, or any future orchestrator without the hook).
func TestEmitSuppressedSeamlessOrchNoOp(t *testing.T) {
	md := &Mind{orch: &scriptedModel{}, social: nil}
	md.emitSuppressed("planner", 0, 100, cognition.Verdict{Class: "planner"})
}

// TestMapCorrectionRearmsMatchingIntent (spec 041 US3, T021 / contracts §1):
// an agent.map_corrected absorb re-arms the planner ONLY when a removed fact
// matches the acting agent's current intent target (Target or Res tile) — a
// correction touching nothing the agent acts on stays quiet, and other
// agents are never armed by it.
func TestMapCorrectionRearmsMatchingIntent(t *testing.T) {
	state := sim.NewState(42, worldmap.Generate(42, 64, 64))
	md := &Mind{replica: state}

	corrected := func(agent int, x, y int, seq int64) store.Event {
		b, _ := json.Marshal(sim.MapCorrectedPayload{Agent: sim.Ref(agent), Gone: []sim.PlaceFact{
			{Kind: "fire", X: x, Y: y, Seen: 100, Provenance: "witnessed"},
		}})
		return store.Event{Seq: seq, Tick: 500, Type: "agent.map_corrected", Payload: b}
	}

	// No intent: quiet.
	md.absorb([]store.Event{corrected(0, 10, 10, 71)})
	if md.pending[0] {
		t.Fatal("correction with no intent in flight armed the planner")
	}

	// Intent elsewhere: quiet.
	state.Agents[0].Intent = &sim.Intent{Goal: "refuel_fire", TargetX: 30, TargetY: 30}
	md.absorb([]store.Event{corrected(0, 10, 10, 72)})
	if md.pending[0] {
		t.Fatal("correction not touching the intent target armed the planner")
	}

	// Removed fact at the intent target: re-armed with the correction's seq.
	md.absorb([]store.Event{corrected(0, 30, 30, 73)})
	if !md.pending[0] || md.pendingSeq[0] != 73 {
		t.Fatalf("matching correction did not re-arm: pending=%v seq=%d", md.pending[0], md.pendingSeq[0])
	}

	// A work goal's Res tile matches too; other agents stay quiet throughout.
	state.Agents[1].Intent = &sim.Intent{Goal: "chop", TargetX: 5, TargetY: 5, ResX: 6, ResY: 5}
	md.absorb([]store.Event{corrected(1, 6, 5, 74)})
	if !md.pending[1] || md.pendingSeq[1] != 74 {
		t.Fatal("Res-tile correction did not re-arm")
	}
	if md.pending[2] {
		t.Fatal("a correction armed an unrelated agent")
	}
}
