package sim

// Spec 083 (neglect detector) tests: the reducer-arm units (T004), the
// detector mechanism (T009 — once per episode, defer on class intent, re-arm
// on recovery, asleep skip, generation bump, live-vs-replay identity), and
// the recorded fixtures derived from the documented world-01 evidence (T010
// Oak's death window fires with runway; T011 healthy windows stay silent).
// The raw world-01 log is NOT in-repo — these fixtures are the binding CI
// validation (spec FR-007); the env-guarded probe over the real log lives in
// internal/daemon (FR-008).

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// neglectNeedsChanged builds an absolute-values needs event (the executor
// heartbeat's shape) — the fixture scripts slides with it.
func neglectNeedsChanged(tick int64, agent int, n Needs) store.Event {
	return store.Event{Tick: tick, Type: "agent.needs_changed", Payload: mustPayload(NeedsPayload{
		Agent: agent, Health: n.Health, Food: n.Food, Rest: n.Rest, Warmth: n.Warmth, Morale: n.Morale,
	})}
}

func neglectIntentSet(tick int64, agent int, goal, source string) store.Event {
	return store.Event{Tick: tick, Type: "agent.intent_set", Payload: mustPayload(IntentSetPayload{
		Agent: agent, Goal: goal, TargetX: 1, TargetY: 1, Source: source,
	})}
}

func neglectDetected(tick int64, agent int, need string, level int, since int64) store.Event {
	return store.Event{Tick: tick, Type: "sim.neglect_detected", Payload: mustPayload(NeglectDetectedPayload{
		Agent: agent, Need: need, Level: level, Since: since,
	})}
}

func mustApply(t *testing.T, s *State, evs ...store.Event) {
	t.Helper()
	for _, e := range evs {
		if err := s.Apply(e); err != nil {
			t.Fatalf("apply %s at tick %d: %v", e.Type, e.Tick, err)
		}
	}
}

// --- T004: reducer-arm units -------------------------------------------------

// TestNeglectNeedsArmAnchors: the needs arm sets the band-entry anchor on the
// downward crossing, leaves it alone while the need stays below, and clears
// anchor + fired latch together on recovery to/above the band.
func TestNeglectNeedsArmAnchors(t *testing.T) {
	s := NewState(1, testMap(1))
	a := &s.Agents[0]

	// Genesis needs are healthy — no allocation, pre-083 bytes.
	if a.Neglect != nil {
		t.Fatal("genesis agent must carry a nil Neglect")
	}

	// Downward crossing sets the anchor at the crossing tick.
	mustApply(t, s, neglectNeedsChanged(100, 0, Needs{Health: 1000, Food: 600, Rest: 800, Warmth: 300, Morale: 600}))
	if got := a.Neglect.Since("warmth"); got != 100 {
		t.Fatalf("warmth Since = %d, want 100", got)
	}
	if a.Neglect.Since("food") != 0 || a.Neglect.Since("rest") != 0 {
		t.Error("healthy needs must not carry band anchors")
	}

	// Staying below never re-stamps (the anchor is the ENTRY tick).
	mustApply(t, s, neglectNeedsChanged(160, 0, Needs{Health: 1000, Food: 600, Rest: 800, Warmth: 200, Morale: 600}))
	if got := a.Neglect.Since("warmth"); got != 100 {
		t.Fatalf("warmth Since re-stamped to %d, want 100", got)
	}

	// Recovery clears anchor AND fired latch together (episode over, re-armed).
	mustApply(t, s, neglectDetected(220, 0, "warmth", 200, 100))
	if !a.Neglect.Fired("warmth") {
		t.Fatal("detected arm must set the fired latch")
	}
	mustApply(t, s, neglectNeedsChanged(280, 0, Needs{Health: 1000, Food: 600, Rest: 800, Warmth: 400, Morale: 600}))
	if a.Neglect.Since("warmth") != 0 || a.Neglect.Fired("warmth") {
		t.Errorf("recovery must clear anchor + latch: since=%d fired=%v",
			a.Neglect.Since("warmth"), a.Neglect.Fired("warmth"))
	}
}

// TestNeglectIntentArmStamps: the intent arm stamps exactly the classed
// goals' needs (source-agnostic), and unclassed goals allocate nothing.
func TestNeglectIntentArmStamps(t *testing.T) {
	// Unclassed goals never allocate: pre-083 bytes for a wander/chop-only agent.
	s := NewState(1, testMap(1))
	mustApply(t, s,
		neglectIntentSet(50, 0, "wander", "planner"),
		neglectIntentSet(110, 0, "chop", "reflex"))
	if s.Agents[0].Neglect != nil {
		t.Fatal("unclassed intents must not allocate NeglectState")
	}

	// Every classed goal stamps its need — the whole dictionary, all sources.
	tick := int64(100)
	for goal, need := range needClassGoals {
		s := NewState(1, testMap(1))
		mustApply(t, s, neglectIntentSet(tick, 2, goal, "reflex"))
		if got := s.Agents[2].Neglect.ClassIntent(need); got != tick {
			t.Errorf("%s: ClassIntent(%s) = %d, want %d", goal, need, got, tick)
		}
		for _, other := range neglectNeedOrder {
			if other != need && s.Agents[2].Neglect.ClassIntent(other) != 0 {
				t.Errorf("%s stamped unrelated need %s", goal, other)
			}
		}
	}
}

// TestNeglectDetectedArmLatch: the new arm sets exactly the payload need's
// latch and nothing else.
func TestNeglectDetectedArmLatch(t *testing.T) {
	s := NewState(1, testMap(1))
	mustApply(t, s, neglectDetected(500, 3, "rest", 200, 100))
	a := &s.Agents[3]
	if !a.Neglect.Fired("rest") {
		t.Fatal("rest latch not set")
	}
	if a.Neglect.Fired("food") || a.Neglect.Fired("warmth") {
		t.Error("other latches must stay clear")
	}
	if a.Neglect.Since("rest") != 0 || a.Neglect.ClassIntent("rest") != 0 {
		t.Error("the detected arm must touch only the latch")
	}
}

// --- T010 + T009: the Oak-shaped fixture -------------------------------------

// oakSlideNeeds is the documented world-01 death-window shape (research §1.3,
// thrash-detection-research.md finding 3): warmth 636→0 at the night rate
// (warmthLossCold 4/min) starting at tick 600, then health draining at
// healthLoss 3/min once warmth hits 0. Food/rest held healthy to isolate the
// warmth slide.
func oakSlideNeeds(tick int64) Needs {
	warmth := 636 - 4*int((tick-600)/60)
	if warmth < 0 {
		warmth = 0
	}
	health := 1000
	if warmth == 0 {
		zeroAt := int64(600 + 60*(636/4)) // tick 10140
		if tick > zeroAt {
			health -= 3 * int((tick-zeroAt)/60)
			if health < 0 {
				health = 0
			}
		}
	}
	return Needs{Health: health, Food: 600, Rest: 800, Warmth: warmth, Morale: 600}
}

// runNeglectSweep advances the fixture from tick from to tick to (heartbeat
// steps), evaluating the executor sweep against the pre-tick state at every
// heartbeat, applying ONLY the detector's own emitted pair (event + companion
// memory — asserted adjacent, event first), then folding the scripted needs
// for that beat. Returns the firings with their ticks.
func runNeglectSweep(t *testing.T, s *State, from, to int64, script func(tick int64) []store.Event) ([]NeglectDetectedPayload, []int64) {
	t.Helper()
	m := testMap(42)
	var firings []NeglectDetectedPayload
	var fireTicks []int64
	for tick := from; tick <= to; tick += 60 {
		evs := stepEvents(s, m, tick)
		for i, e := range evs {
			if e.Type != "sim.neglect_detected" {
				continue
			}
			var p NeglectDetectedPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				t.Fatalf("bad payload: %v", err)
			}
			firings = append(firings, p)
			fireTicks = append(fireTicks, tick)
			// FR-004: the companion memory rides the same batch immediately
			// after the event.
			if i+1 >= len(evs) || evs[i+1].Type != "agent.memory_added" {
				t.Fatalf("no companion agent.memory_added immediately after the firing at tick %d", tick)
			}
			var mp MemoryAddedPayload
			if err := json.Unmarshal(evs[i+1].Payload, &mp); err != nil {
				t.Fatalf("bad memory payload: %v", err)
			}
			if mp.Agent != p.Agent || mp.Salience != salNeglect || mp.Origin != OriginWitness ||
				mp.Where == nil || mp.Why != "" {
				t.Errorf("companion memory shape wrong: %+v", mp)
			}
			// situateText splices the where-clause before the trailing
			// period, so match the fixed text's stem.
			if !strings.HasPrefix(mp.Text, strings.TrimSuffix(neglectMemoryText(p.Need), ".")) {
				t.Errorf("memory text %q does not carry the fixed per-need text", mp.Text)
			}
			mustApply(t, s, e, evs[i+1])
		}
		s.Tick = tick
		if script != nil {
			mustApply(t, s, script(tick)...)
		}
	}
	return firings, fireTicks
}

// TestNeglectOakShapedFixture (spec SC-001, FR-007): the fixture derived from
// Oak's documented death window — warmth 636→0 at 4/min with only reflex
// chop and planner wander intent records — fires sim.neglect_detected exactly
// once, at band-entry + neglectWindowTicks, with ≈5 game-hours of health
// runway before the death the same trajectory produces; a second full window
// without recovery adds nothing (SC-004 first half); the companion memory
// bumps Oak's generation (US1 AS5).
func TestNeglectOakShapedFixture(t *testing.T) {
	const oak = 6 // AgentNames[6] == "Oak"
	if AgentNames[oak] != "Oak" {
		t.Fatalf("agent %d is %q, not Oak", oak, AgentNames[oak])
	}
	s := NewState(42, testMap(42))
	s.Tick = 540

	// The slide's fixed arithmetic: band entry (first heartbeat with warmth
	// < 350) at tick 4920; warmth 0 at 10140; firing due at the first
	// heartbeat >= 4920+7200 = 12120; the trajectory's death (health 0) at
	// tick 30180 — 18060 ticks (~5 game-hours) after the firing.
	const bandEntry, fireAt, deathAt = 4920, 12120, 30180
	if n := oakSlideNeeds(bandEntry); n.Warmth >= dangerWarmthBelow {
		t.Fatalf("fixture arithmetic: warmth %d at band entry", n.Warmth)
	}
	if n := oakSlideNeeds(bandEntry - 60); n.Warmth < dangerWarmthBelow {
		t.Fatal("fixture arithmetic: band entered a beat early")
	}
	if n := oakSlideNeeds(deathAt); n.Health != 0 {
		t.Fatalf("fixture arithmetic: health %d at the death tick", n.Health)
	}

	gen0 := s.Agents[oak].Generation
	// Drive through the firing plus a full second window (12120+7200=19320).
	script := func(tick int64) []store.Event {
		evs := []store.Event{neglectNeedsChanged(tick, oak, oakSlideNeeds(tick))}
		// Oak's only intents in the window: reflex chop / planner wander,
		// alternating every 10 game-minutes (research §1.3 — neither is
		// warmth-class, so the zero-intent clock never resets).
		if (tick-600)%600 == 0 {
			goal, src := "chop", "reflex"
			if ((tick-600)/600)%2 == 1 {
				goal, src = "wander", "planner"
			}
			evs = append(evs, neglectIntentSet(tick, oak, goal, src))
		}
		return evs
	}
	firings, fireTicks := runNeglectSweep(t, s, 600, 19500, script)

	if len(firings) != 1 {
		t.Fatalf("firings = %d (%v at %v), want exactly 1", len(firings), firings, fireTicks)
	}
	p := firings[0]
	if fireTicks[0] != fireAt {
		t.Errorf("fired at %d, want band-entry+%d = %d", fireTicks[0], neglectWindowTicks, fireAt)
	}
	if p.Agent != oak || p.Need != "warmth" || p.Since != bandEntry {
		t.Errorf("payload = %+v, want agent %d / warmth / since %d", p, oak, bandEntry)
	}
	if p.Level != 0 {
		t.Errorf("pre-tick level at firing = %d, want 0 (warmth long gone)", p.Level)
	}
	// Runway (SC-001): the firing lands with health ≈ 900 — ~5 game-hours
	// before the death the same decay produces.
	if h := oakSlideNeeds(fireTicks[0] - 60).Health; h < 800 {
		t.Errorf("health at firing = %d — no runway left", h)
	}
	if deathAt-fireTicks[0] < 4*3600 {
		t.Errorf("runway = %d ticks, want >= 4 game-hours", deathAt-fireTicks[0])
	}
	// The latch held for the whole second window (drive ran to 19500 > 19320).
	if !s.Agents[oak].Neglect.Fired("warmth") {
		t.Error("fired latch not held after the firing")
	}
	// US1 AS5: the salience-9 companion bumped the generation.
	if got := s.Agents[oak].Generation; got != gen0+1 {
		t.Errorf("Generation = %d, want %d (one bump from the percept)", got, gen0+1)
	}
	mems := s.Agents[oak].Memories
	if len(mems) == 0 || mems[len(mems)-1].Salience != salNeglect {
		t.Error("percept memory did not accrete at salNeglect")
	}
}

// TestNeglectRecoveryReArms (US1 AS4, SC-004 second half): after a fired
// episode, recovery to/above the band re-arms the detector; a fresh window
// below the band with zero class intents fires exactly once more.
func TestNeglectRecoveryReArms(t *testing.T) {
	s := NewState(42, testMap(42))
	s.Tick = 540
	// Phase 1: slide below the band with no class intents — fires at 600+7200.
	firings, fireTicks := runNeglectSweep(t, s, 600, 7800+120, func(tick int64) []store.Event {
		return []store.Event{neglectNeedsChanged(tick, 0, Needs{Health: 1000, Food: 600, Rest: 800, Warmth: 300, Morale: 600})}
	})
	if len(firings) != 1 || fireTicks[0] != 7800 {
		t.Fatalf("phase 1: firings %v at %v, want one at 7800", firings, fireTicks)
	}
	// Phase 2: recovery (episode closes, latch re-arms), then relapse at
	// 9000 — a fresh window fires once more at 9000+7200 = 16200.
	firings, fireTicks = runNeglectSweep(t, s, 8040, 16500, func(tick int64) []store.Event {
		w := 500 // recovered
		if tick >= 9000 {
			w = 300 // relapsed
		}
		return []store.Event{neglectNeedsChanged(tick, 0, Needs{Health: 1000, Food: 600, Rest: 800, Warmth: w, Morale: 600})}
	})
	if len(firings) != 1 || fireTicks[0] != 16200 {
		t.Fatalf("phase 2: firings %v at %v, want exactly one at 16200 (new episode)", firings, fireTicks)
	}
	if firings[0].Since != 9000 {
		t.Errorf("relapse Since = %d, want 9000", firings[0].Since)
	}
}

// --- T011: healthy fixtures stay silent --------------------------------------

// TestNeglectSilentWithClassIntents (US1 AS2, SC-001): a need below its band
// with class intents landing inside every window never fires — the class
// intent proves the mind engaged, whatever the outcome (Oak's productive
// day-4 shuttling shape). When the intents stop, the SLIDING zero-intent
// clock runs out a full window after the last one — pinning the deferral
// arithmetic, not just the silence.
func TestNeglectSilentWithClassIntents(t *testing.T) {
	s := NewState(42, testMap(42))
	s.Tick = 540
	const lastIntent = 15300 // goto_warmth every game-hour, 900..15300
	script := func(tick int64) []store.Event {
		evs := []store.Event{neglectNeedsChanged(tick, 0, Needs{Health: 1000, Food: 600, Rest: 800, Warmth: 300, Morale: 600})}
		if tick >= 900 && tick <= lastIntent && (tick-900)%3600 == 0 {
			evs = append(evs, neglectIntentSet(tick, 0, "goto_warmth", "reflex"))
		}
		return evs
	}
	// Three full windows of covered slide: silent throughout.
	firings, _ := runNeglectSweep(t, s, 600, lastIntent, script)
	if len(firings) != 0 {
		t.Fatalf("healthy shuttling fired %d times, want silence", len(firings))
	}
	// The clock is sliding: the firing lands one full window after the LAST
	// class intent (15300+7200 = 22500), not after band entry.
	firings, fireTicks := runNeglectSweep(t, s, lastIntent+60, 22800, script)
	if len(firings) != 1 || fireTicks[0] != lastIntent+neglectWindowTicks {
		t.Fatalf("post-intent firing %v at %v, want exactly one at %d", firings, fireTicks, lastIntent+neglectWindowTicks)
	}
}

// TestNeglectSilentOnDipAndRecover (US2 AS3, SC-001): routine dips that
// recover above the band before the window completes never fire, and each
// recovery resets the anchors.
func TestNeglectSilentOnDipAndRecover(t *testing.T) {
	s := NewState(42, testMap(42))
	s.Tick = 540
	// Dip 600..4140 (~1h, under the 2h window), recover 4200..6000, dip
	// again 6060..9600, recover to the end.
	warm := func(tick int64) int {
		switch {
		case tick <= 4140, tick >= 6060 && tick <= 9600:
			return 300
		default:
			return 450
		}
	}
	firings, _ := runNeglectSweep(t, s, 600, 18000, func(tick int64) []store.Event {
		return []store.Event{neglectNeedsChanged(tick, 0, Needs{Health: 1000, Food: 600, Rest: 800, Warmth: warm(tick), Morale: 600})}
	})
	if len(firings) != 0 {
		t.Fatalf("dip-and-recover fired %d times, want silence", len(firings))
	}
	if got := s.Agents[0].Neglect.Since("warmth"); got != 0 {
		t.Errorf("anchor after final recovery = %d, want 0 (reset)", got)
	}
}

// TestNeglectSkipsAsleepFiresOnWake (edge case "Asleep villager"): the sweep
// skips a sleeper at the beat — anchors keep accruing — and a still-critical
// waker fires on its next heartbeat.
func TestNeglectSkipsAsleepFiresOnWake(t *testing.T) {
	s := NewState(42, testMap(42))
	s.Tick = 540
	script := func(tick int64) []store.Event {
		return []store.Event{neglectNeedsChanged(tick, 0, Needs{Health: 1000, Food: 600, Rest: 800, Warmth: 300, Morale: 600})}
	}
	// In band from 600; asleep before the window completes.
	firings, _ := runNeglectSweep(t, s, 600, 7500, script)
	if len(firings) != 0 {
		t.Fatalf("premature firing")
	}
	mustApply(t, s, store.Event{Tick: 7500, Type: "agent.slept", Payload: mustPayload(AgentPayload{Agent: 0})})
	// Due at 7800, but asleep: skipped at the beat, predicate false.
	firings, _ = runNeglectSweep(t, s, 7560, 9000, script)
	if len(firings) != 0 {
		t.Fatalf("fired while asleep")
	}
	if NeglectDue(&s.Agents[0], "warmth", 9000) {
		t.Error("NeglectDue must be false for a sleeper")
	}
	// Wake still critical: the next heartbeat fires.
	mustApply(t, s, store.Event{Tick: 9030, Type: "agent.woke", Payload: mustPayload(AgentPayload{Agent: 0})})
	firings, fireTicks := runNeglectSweep(t, s, 9060, 9240, script)
	if len(firings) != 1 || fireTicks[0] != 9060 {
		t.Fatalf("waker firing %v at %v, want exactly one at 9060", firings, fireTicks)
	}
}

// --- T009: live-vs-replay byte identity --------------------------------------

// neglectReplayTimeline scripts a live drive into a natural executor firing:
// agent 0's warmth is pinned below the band (the day drift +2/min cannot
// reach 350 between re-pins) while far-waypoint wander intents keep it
// marching — never idle, so the reflex never runs and the zero-class-intent
// clause holds by construction (wander is unclassed — Oak's planner shape).
func neglectReplayTimeline() map[int64][]store.Event {
	tl := map[int64][]store.Event{}
	add := func(tick int64, typ string, payload any) {
		tl[tick] = append(tl[tick], store.Event{Tick: tick, Type: typ, Payload: mustPayload(payload)})
	}
	for tick := int64(30); tick <= 7500; tick += 1200 {
		add(tick, "agent.needs_changed", NeedsPayload{Agent: 0, Health: 1000, Food: 1000, Rest: 1000, Warmth: 200, Morale: 600})
	}
	for i, tick := 0, int64(10); tick <= 7500; i, tick = i+1, tick+100 {
		x := 2
		if i%2 == 1 {
			x = 62
		}
		add(tick, "agent.intent_set", IntentSetPayload{Agent: 0, Goal: "wander", TargetX: x, TargetY: 32, Source: "planner"})
	}
	return tl
}

// TestNeglectLiveVsReplayByteIdentical (FR-009, SC-003): a live drive whose
// log contains the detector's own sim.neglect_detected + companion memory
// replays from genesis to a byte-identical state — reducer-only writes, pure
// emission. The governor_replay_test idiom.
func TestNeglectLiveVsReplayByteIdentical(t *testing.T) {
	const seed, ticks = 42, 7600
	m := testMap(seed)
	live := NewState(seed, m)
	log := driveTicks(t, live, m, ticks, neglectReplayTimeline())

	// Guard: the log must actually carry the firing (band entry at tick 30,
	// due at the first heartbeat >= 7230 -> 7260) and its companion, and
	// agent 0 must have set NO warmth-class intent (or the run proves nothing).
	var fired, memAfter int
	for i, e := range log {
		switch e.Type {
		case "sim.neglect_detected":
			var p NeglectDetectedPayload
			json.Unmarshal(e.Payload, &p)
			if p.Agent != 0 || p.Need != "warmth" || p.Since != 30 {
				t.Fatalf("unexpected firing %+v", p)
			}
			fired++
			if i+1 < len(log) && log[i+1].Type == "agent.memory_added" {
				memAfter++
			}
		case "agent.intent_set":
			var p IntentSetPayload
			json.Unmarshal(e.Payload, &p)
			if p.Agent == 0 && needClassOf(p.Goal) == "warmth" {
				t.Fatalf("agent 0 set a warmth-class intent (%s at %d) — the fixture no longer isolates neglect", p.Goal, e.Tick)
			}
		}
	}
	if fired != 1 || memAfter != 1 {
		t.Fatalf("log carries %d firings (%d with adjacent companion), want exactly 1 and 1", fired, memAfter)
	}
	if !live.Agents[0].Neglect.Fired("warmth") || live.Agents[0].Generation == 0 {
		t.Fatal("live state missing the latch or the generation bump")
	}

	// Genesis replay: reduce the logged events, align the clock, re-live the
	// quiet tail — exactly the recovery contract.
	replayed := NewState(seed, m)
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replayed.Tick = e.Tick
	}
	driveTicks(t, replayed, m, ticks, nil)

	if live.Hash() != replayed.Hash() {
		t.Fatalf("replay diverged from live:\nlive:     %s\nreplayed: %s",
			string(live.Marshal()), string(replayed.Marshal()))
	}
	// The spec-083 fields specifically survive replay — named, not just
	// implied by the whole-state hash.
	lv, rp := live.Agents[0].Neglect, replayed.Agents[0].Neglect
	if lv == nil || rp == nil || *lv != *rp {
		t.Errorf("NeglectState diverged: live %+v replayed %+v", lv, rp)
	}
	if live.Agents[0].Generation != replayed.Agents[0].Generation {
		t.Errorf("Generation diverged: %d vs %d", live.Agents[0].Generation, replayed.Agents[0].Generation)
	}
}

// --- snapshot byte-identity ---------------------------------------------------

// TestPre083SnapshotRoundTripByteIdentical (Key Entities, SC-006): a state
// carrying no neglect substrate marshals with NO "neglect" key, and
// unmarshal + re-marshal is byte-identical — omitempty pointer, no
// format_version bump (the Journal/Hail/Map precedent).
func TestPre083SnapshotRoundTripByteIdentical(t *testing.T) {
	s := NewState(42, testMap(42))
	got := s.Marshal()
	if bytes.Contains(got, []byte(`"neglect"`)) {
		t.Error("pre-083 state leaked the neglect key")
	}
	s2 := NewState(42, testMap(42))
	if err := json.Unmarshal(got, s2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if again := s2.Marshal(); !bytes.Equal(got, again) {
		t.Error("pre-083 snapshot did not round-trip byte-identically")
	}
}
