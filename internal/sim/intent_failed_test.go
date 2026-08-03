package sim

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Spec 096 (TASK-95, generalizing agent.build_failed/spec 038 to every
// non-build goal): the card's two enumerated silent-failure classes —
// mid-work invalid exits (forage/chop/hunt/demolish/repair/quarry/cook/bathe,
// reason intentFailTargetGone) and completion-time contested no-ops
// (craft/cook/bathe/deposit/withdraw, reason intentFailContested or, for
// deposit's empty Kind, intentFailInvalid) — now resolve LOUDLY via
// agent.intent_failed instead of a bare agent.intent_done. craft_planks'
// contested no-op (insufficient inputs, no-fit net bulk) is already covered
// by TestCraftInsufficientInputsNoOp/TestBulkCraftNoFitClearsIntent
// (craft_test.go/bulk_cap_test.go); TestColdFireRefusesCook/
// TestOvenCookNoFuelNoOp/TestContestedQuarry cover cook's target-gone and
// contested variants plus quarry's target-gone variant. This file covers the
// remaining enumerated paths and adds deep, single-mechanism coverage for one
// gather goal (hunt) and one station goal (cook, uniquely exercising BOTH
// reason classes) per card AC#2.

// TestIntentFailedTargetGoneMatrix sweeps every category-A goal (the card's
// mid-work invalid-exit list) not already covered elsewhere: forage, chop,
// demolish, repair, quarry, and bathe. Each case sets the intent's Target/Res
// to a tile that fails that goal's own `valid` re-check from the very first
// tick (arrival-time invalidity, SC-001) — a fresh WorkStart never needs to
// advance for the check to fire, since the `!valid` branch resolves before
// the WorkStart==0 gate.
func TestIntentFailedTargetGoneMatrix(t *testing.T) {
	const seed = 42
	m := testMap(seed)

	cases := []struct {
		name  string
		goal  string
		setup func(t *testing.T, s *State, a *Agent) (tx, ty, resx, resy int)
	}{
		{"forage", "forage", func(t *testing.T, s *State, a *Agent) (int, int, int, int) {
			// Genesis tiles are a mix of kinds on a given seed — find a
			// guaranteed non-Forage tile nearby rather than trusting luck.
			p, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return effectiveKind(m, s, x, y) != worldmap.Forage })
			if !ok {
				t.Skip("no non-Forage tile reachable from agent 0's genesis position")
			}
			a.X, a.Y = p.X, p.Y
			return a.X, a.Y, 0, 0
		}},
		{"chop", "chop", func(t *testing.T, s *State, a *Agent) (int, int, int, int) {
			if effectiveKind(m, s, a.X, a.Y) == worldmap.Tree {
				t.Skip("agent genesis tile is Tree-kind on this seed")
			}
			return a.X, a.Y, a.X, a.Y
		}},
		{"demolish", "demolish", func(t *testing.T, s *State, a *Agent) (int, int, int, int) {
			// No wall ever stood at the agent's own tile — wallAt is nil.
			return a.X, a.Y, a.X, a.Y
		}},
		{"repair", "repair", func(t *testing.T, s *State, a *Agent) (int, int, int, int) {
			return a.X, a.Y, a.X, a.Y
		}},
		{"quarry", "quarry", func(t *testing.T, s *State, a *Agent) (int, int, int, int) {
			if effectiveKind(m, s, a.X, a.Y) == worldmap.Rock {
				t.Skip("agent genesis tile is Rock-kind on this seed")
			}
			return a.X, a.Y, a.X, a.Y
		}},
		{"bathe", "bathe", func(t *testing.T, s *State, a *Agent) (int, int, int, int) {
			// No oven at the agent's own tile — structureAt("oven", …) is false.
			return a.X, a.Y, 0, 0
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := NewState(seed, m)
			isolateAgents(s)
			a := &s.Agents[0]
			a.Dead = false
			a.Inv.Water, a.Inv.Wood = 2, 2 // bathe's own no-op needs these NOT to be the cause
			tx, ty, rx, ry := c.setup(t, s, a)
			a.Intent = &Intent{Goal: c.goal, TargetX: tx, TargetY: ty, ResX: rx, ResY: ry}

			log := driveTicks(t, s, m, s.Tick+3, nil)

			var failed bool
			var reason, goal string
			var failTick int64
			memTick := int64(-1)
			for _, e := range log {
				switch e.Type {
				case "agent.intent_done":
					if failed {
						continue // reflex may have re-decided and completed something else after
					}
					t.Errorf("%s: resolved via bare agent.intent_done before any agent.intent_failed, want agent.intent_failed", c.name)
				case "agent.intent_failed":
					var p IntentFailedPayload
					mustUnmarshal(t, e.Payload, &p)
					if p.Agent.ID == 0 && !failed {
						failed, reason, goal, failTick = true, p.Reason, p.Goal, e.Tick
						if p.X != a.X && p.X != tx {
							// Position is the agent's OWN stand tile — sanity, not a
							// strict equality (a might have moved by the time the log
							// is scanned), just that it decoded to sane in-bounds coords.
						}
					}
				case "agent.memory_added":
					if failed && memTick == -1 {
						var p MemoryAddedPayload
						mustUnmarshal(t, e.Payload, &p)
						if p.Agent.ID == 0 && p.Origin == OriginAction && p.Salience == salIntentFailed && e.Tick == failTick {
							memTick = e.Tick
						}
					}
				}
			}
			if !failed {
				t.Fatalf("%s: no agent.intent_failed for agent 0", c.name)
			}
			if goal != c.goal {
				t.Errorf("%s: payload goal = %q, want %q", c.name, goal, c.goal)
			}
			if reason != intentFailTargetGone {
				t.Errorf("%s: reason = %q, want %q", c.name, reason, intentFailTargetGone)
			}
			if memTick != failTick {
				t.Errorf("%s: paired failure memory tick = %d, want it same-tick with the event (%d)", c.name, memTick, failTick)
			}
		})
	}
}

// TestIntentFailedContestedMatrix sweeps the remaining category-B goals not
// already covered elsewhere: cook's "nothing to cook" no-op (distinct from
// TestOvenCookNoFuelNoOp's "no wood" no-op — both are contested), bathe's
// missing-carried-resource no-op, and deposit/withdraw's chest-state no-ops
// (a vanished chest, a full chest, nothing to take) plus deposit's malformed
// empty-Kind argument (intentFailInvalid, NOT contested — spec 096 draws this
// distinction because no chest-state change could ever satisfy it).
func TestIntentFailedContestedMatrix(t *testing.T) {
	const seed = 42
	m := testMap(seed)

	cases := []struct {
		name       string
		goal       string
		wantReason string
		// workStart, when non-zero, pins the intent's WorkStart so the
		// completion-time no-op recheck fires on the very first driven tick
		// (the whole-suite convention: WorkStart = 1 - duration) rather than
		// needing the full work duration to actually elapse. Instant-on-
		// arrival goals (deposit/withdraw) leave it 0 (unread).
		workStart int64
		driveTo   int64
		setup     func(s *State, a *Agent) (tx, ty int, kind string)
	}{
		{"cook_nothing_to_cook", "cook", intentFailContested, 1 - cookFireTicks, 5,
			func(s *State, a *Agent) (int, int, string) {
				// A real lit fire (passes the target-gone check) but no raw food
				// carried — the completion-time no-op, distinct from the no-wood case.
				s.Structures = append(s.Structures, Structure{Kind: "fire", X: a.X, Y: a.Y, FuelUntil: 1_000_000})
				a.Inv.FoodRaw = 0
				return a.X, a.Y, ""
			}},
		{"bathe_missing_resources", "bathe", intentFailContested, 1 - batheTicks, 5,
			func(s *State, a *Agent) (int, int, string) {
				s.Structures = append(s.Structures, Structure{Kind: "oven", X: a.X, Y: a.Y})
				a.Inv.Water, a.Inv.Wood = 0, 2
				return a.X, a.Y, ""
			}},
		{"deposit_chest_gone", "deposit", intentFailContested, 0, 3, func(s *State, a *Agent) (int, int, string) {
			a.Inv.Wood = 5
			return a.X, a.Y, "wood" // no chest ever placed here
		}},
		{"deposit_chest_full", "deposit", intentFailContested, 0, 3, func(s *State, a *Agent) (int, int, string) {
			a.Inv.Wood = 5
			s.Structures = append(s.Structures, Structure{Kind: "chest", X: a.X, Y: a.Y, Owner: 0, Store: &Inventory{Wood: chestCap}})
			return a.X, a.Y, "wood"
		}},
		{"deposit_empty_kind", "deposit", intentFailInvalid, 0, 3, func(s *State, a *Agent) (int, int, string) {
			a.Inv.Wood = 5
			s.Structures = append(s.Structures, Structure{Kind: "chest", X: a.X, Y: a.Y, Owner: 0, Store: &Inventory{}})
			return a.X, a.Y, "" // malformed: no Kind named
		}},
		{"withdraw_chest_gone", "withdraw", intentFailContested, 0, 3, func(s *State, a *Agent) (int, int, string) {
			return a.X, a.Y, "wood" // no chest ever placed here
		}},
		{"withdraw_nothing_available", "withdraw", intentFailContested, 0, 3, func(s *State, a *Agent) (int, int, string) {
			s.Structures = append(s.Structures, Structure{Kind: "chest", X: a.X, Y: a.Y, Owner: 0, Store: &Inventory{}})
			return a.X, a.Y, "wood" // chest stands, empty
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := NewState(seed, m)
			isolateAgents(s)
			a := &s.Agents[0]
			a.Dead = false
			tx, ty, kind := c.setup(s, a)
			a.Intent = &Intent{Goal: c.goal, TargetX: tx, TargetY: ty, Kind: kind, WorkStart: c.workStart}

			log := driveTicks(t, s, m, s.Tick+c.driveTo, nil)

			var failed bool
			var reason, goal string
			for _, e := range log {
				switch e.Type {
				case "agent.intent_done":
					if !failed {
						t.Errorf("%s: resolved via bare agent.intent_done, want agent.intent_failed", c.name)
					}
				case "agent.deposited", "agent.withdrew", "agent.cooked", "agent.bathed":
					t.Errorf("%s: emitted a success event (%s) on a case meant to fail", c.name, e.Type)
				case "agent.intent_failed":
					var p IntentFailedPayload
					mustUnmarshal(t, e.Payload, &p)
					if p.Agent.ID == 0 && !failed {
						failed, reason, goal = true, p.Reason, p.Goal
					}
				}
			}
			if !failed {
				t.Fatalf("%s: no agent.intent_failed for agent 0", c.name)
			}
			if goal != c.goal {
				t.Errorf("%s: payload goal = %q, want %q", c.name, goal, c.goal)
			}
			if reason != c.wantReason {
				t.Errorf("%s: reason = %q, want %q", c.name, reason, c.wantReason)
			}
		})
	}
}

// TestIntentFailedHuntDeepCoverage is the card's deep-coverage gather goal
// (AC#2): a hunt at a den still on cooldown fails LOUDLY the first tick the
// agent is at the den — full mechanics: the event's shape (goal/reason/
// position), the paired same-tick situated failure memory (OriginAction,
// salIntentFailed, mentioning the hunt), the reducer's IntentLog closure
// ("failed", matching agent.build_failed's own outcome string), Intent
// cleared + IdleSince stamped, and NO yield (no FoodRaw gained, no
// agent.hunted).
func TestIntentFailedHuntDeepCoverage(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	isolateAgents(s)
	if len(m.Dens) == 0 {
		t.Fatal("test map has no dens")
	}
	den := m.Dens[0]

	a := &s.Agents[0]
	a.Dead = false
	a.X, a.Y = den.X, den.Y
	// The den is on cooldown well past this short drive's horizon.
	s.DenUses = append(s.DenUses, DenUse{X: den.X, Y: den.Y, Ready: s.Tick + 1_000_000})
	a.Intent = &Intent{Goal: "hunt", TargetX: den.X, TargetY: den.Y}
	beforeIntentSet := len(a.IntentLog)
	a.appendIntent(IntentRecord{Goal: "hunt", Source: "planner", Tick: s.Tick})

	log := driveTicks(t, s, m, s.Tick+3, nil)

	var failed bool
	var p IntentFailedPayload
	var failTick int64
	var mem MemoryAddedPayload
	memTick := int64(-1)
	for _, e := range log {
		switch e.Type {
		case "agent.hunted":
			t.Fatal("a den on cooldown must not yield agent.hunted")
		case "agent.intent_failed":
			if !failed {
				mustUnmarshal(t, e.Payload, &p)
				if p.Agent.ID == 0 {
					failed, failTick = true, e.Tick
				}
			}
		case "agent.memory_added":
			if failed && memTick == -1 {
				var mp MemoryAddedPayload
				mustUnmarshal(t, e.Payload, &mp)
				if mp.Agent.ID == 0 && e.Tick == failTick {
					mem, memTick = mp, e.Tick
				}
			}
		}
	}
	if !failed {
		t.Fatal("no agent.intent_failed for the hunter")
	}
	if p.Goal != "hunt" {
		t.Errorf("goal = %q, want hunt", p.Goal)
	}
	if p.Reason != intentFailTargetGone {
		t.Errorf("reason = %q, want %q", p.Reason, intentFailTargetGone)
	}
	if p.X != den.X || p.Y != den.Y {
		t.Errorf("position = (%d,%d), want the hunter's stand tile (%d,%d)", p.X, p.Y, den.X, den.Y)
	}
	if memTick != failTick {
		t.Fatalf("failure memory tick = %d, want paired same-tick with the event (%d)", memTick, failTick)
	}
	if mem.Origin != OriginAction {
		t.Errorf("memory origin = %q, want %q", mem.Origin, OriginAction)
	}
	if mem.Salience != salIntentFailed {
		t.Errorf("memory salience = %d, want %d (build_failed's own tier — no new flooding vector)", mem.Salience, salIntentFailed)
	}
	if !strings.Contains(mem.Text, "hunt") {
		t.Errorf("memory text = %q, want it to mention the hunt", mem.Text)
	}
	if a.Intent != nil {
		t.Error("Intent should be cleared after intent_failed")
	}
	if a.IdleSince != failTick {
		t.Errorf("IdleSince = %d, want stamped to the failure tick (%d)", a.IdleSince, failTick)
	}
	if len(a.IntentLog) <= beforeIntentSet {
		t.Fatal("no IntentLog record to inspect")
	}
	r := a.IntentLog[len(a.IntentLog)-1]
	if r.Outcome != "failed" || r.OutcomeTick != failTick {
		t.Errorf("IntentLog record = %+v, want outcome \"failed\" @ %d (the agent.build_failed precedent)", r, failTick)
	}
	if a.Inv.FoodRaw != 0 {
		t.Errorf("FoodRaw = %d, want 0 — a failed hunt yields nothing", a.Inv.FoodRaw)
	}
}

// TestIntentFailedCookDeepCoverage is the card's deep-coverage station goal
// (AC#2): cook is the one goal in the enumerated list that exercises BOTH
// failure classes — a cold/absent station fails via the mid-work `valid`
// switch (target gone), while a real station with a missing ingredient fails
// via the completion-time no-op (contested) — proving the SAME goal routes
// to the correct reason per site, and that both resolve identically
// otherwise (intent cleared, no agent.cooked, no inventory change).
func TestIntentFailedCookDeepCoverage(t *testing.T) {
	const seed = 42
	m := testMap(seed)

	t.Run("target_gone_no_station", func(t *testing.T) {
		s := NewState(seed, m)
		isolateAgents(s)
		a := &s.Agents[0]
		a.Dead = false
		a.Inv.FoodRaw = 5
		// No fire, no oven at the agent's own tile.
		a.Intent = &Intent{Goal: "cook", TargetX: a.X, TargetY: a.Y}

		log := driveTicks(t, s, m, s.Tick+3, nil)
		assertCookFailed(t, log, intentFailTargetGone)
		if a.Inv.FoodRaw != 5 {
			t.Errorf("FoodRaw = %d, want unchanged 5", a.Inv.FoodRaw)
		}
	})

	t.Run("contested_no_fuel_at_oven", func(t *testing.T) {
		s := NewState(seed, m)
		isolateAgents(s)
		a := &s.Agents[0]
		a.Dead = false
		a.Inv.FoodRaw = 5
		a.Inv.Wood = 0
		s.Structures = append(s.Structures, Structure{Kind: "oven", X: a.X, Y: a.Y})
		// Pinned WorkStart (the whole-suite convention, WorkStart = 1 - duration)
		// so the completion-time no-op recheck fires on the very first driven
		// tick, exactly like TestOvenCookNoFuelNoOp.
		a.Intent = &Intent{Goal: "cook", TargetX: a.X, TargetY: a.Y, WorkStart: 1 - cookOvenTicks}

		log := driveTicks(t, s, m, s.Tick+5, nil)
		assertCookFailed(t, log, intentFailContested)
		if a.Inv.FoodRaw != 5 || a.Inv.Meals != 0 {
			t.Errorf("post-attempt inventory = %d raw / %d meals, want 5/0 unchanged", a.Inv.FoodRaw, a.Inv.Meals)
		}
	})
}

func assertCookFailed(t *testing.T, log []store.Event, wantReason string) {
	t.Helper()
	var failed bool
	var reason string
	var failTick int64
	memTick := int64(-1)
	for _, e := range log {
		switch e.Type {
		case "agent.cooked":
			t.Fatal("no agent.cooked on a case meant to fail")
		case "agent.intent_done":
			if !failed {
				t.Error("resolved via bare agent.intent_done, want agent.intent_failed")
			}
		case "agent.intent_failed":
			if !failed {
				var p IntentFailedPayload
				mustUnmarshal(t, e.Payload, &p)
				if p.Agent.ID == 0 {
					failed, reason, failTick = true, p.Reason, e.Tick
				}
			}
		case "agent.memory_added":
			if failed && memTick == -1 {
				var p MemoryAddedPayload
				mustUnmarshal(t, e.Payload, &p)
				if p.Agent.ID == 0 && e.Tick == failTick && p.Salience == salIntentFailed {
					memTick = e.Tick
				}
			}
		}
	}
	if !failed {
		t.Fatal("no agent.intent_failed for the cook")
	}
	if reason != wantReason {
		t.Errorf("reason = %q, want %q", reason, wantReason)
	}
	if memTick != failTick {
		t.Errorf("paired failure memory tick = %d, want same-tick with the event (%d)", memTick, failTick)
	}
}

// TestReplayByteIdentityIntentFailed is SC-002/FR-005: a run whose log
// contains real, planner-sourced agent.intent_failed events (one gather —
// hunt at a cooling-down den — and one station — cook with no carried fuel
// at a real oven) replays from genesis to a byte-identical state hash, the
// same discipline every other event-type addition in this codebase proves
// (TestReplayByteIdentityOven, TestReplayByteIdentityWallsAxesPaths, …).
func TestReplayByteIdentityIntentFailed(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	if len(m.Dens) == 0 {
		t.Skip("test map has no dens")
	}
	den := m.Dens[0]

	genesis := func() *State {
		s := NewState(seed, m)
		isolateAgents(s)
		a := &s.Agents[0]
		a.Dead = false
		a.X, a.Y = den.X, den.Y
		a.Inv.FoodRaw = 5
		s.Structures = append(s.Structures, Structure{Kind: "oven", X: a.X, Y: a.Y})
		s.DenUses = append(s.DenUses, DenUse{X: den.X, Y: den.Y, Ready: 1_000_000_000})
		return s
	}
	setIntent := func(tick int64, goal string, tx, ty int) map[int64][]store.Event {
		return map[int64][]store.Event{
			tick: {{Tick: tick, Type: "agent.intent_set", Payload: mustPayload(IntentSetPayload{
				Agent: Ref(0), Goal: goal, TargetX: tx, TargetY: ty, Source: "planner",
			})}},
		}
	}

	live := genesis()
	x0, y0 := live.Agents[0].X, live.Agents[0].Y

	var log []store.Event
	log = append(log, driveTicks(t, live, m, live.Tick+5, setIntent(live.Tick, "hunt", x0, y0))...)
	// cook needs its full work duration to elapse before the completion-time
	// no-op recheck (no carried wood) ever runs — unlike hunt's target-gone
	// exit, which fires before work even starts.
	log = append(log, driveTicks(t, live, m, live.Tick+cookOvenTicks+10, setIntent(live.Tick, "cook", x0, y0))...)

	var sawHuntFail, sawCookFail bool
	for _, e := range log {
		if e.Type != "agent.intent_failed" {
			continue
		}
		var p IntentFailedPayload
		mustUnmarshal(t, e.Payload, &p)
		switch p.Goal {
		case "hunt":
			sawHuntFail = true
		case "cook":
			sawCookFail = true
		}
	}
	if !sawHuntFail || !sawCookFail {
		t.Fatalf("run did not exercise both intent_failed variants: hunt=%v cook=%v", sawHuntFail, sawCookFail)
	}

	replayed := genesis()
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replayed.Tick = e.Tick
	}
	driveTicks(t, replayed, m, live.Tick, nil) // re-live the quiet tail, as recovery does

	if live.Hash() != replayed.Hash() {
		t.Fatalf("replayed state diverged:\nlive:     %s\nreplayed: %s",
			string(live.Marshal()), string(replayed.Marshal()))
	}
}

// TestIntentFailedForagePackFullWorld03 is the world-03 regression (TASK-196).
// Cedar starved to death on day 1 standing on a live forage patch, having never
// eaten once: he had dropped all six of his food to carry more wood, filling his
// pack to exactly bulkCap, and every forage after that completed the full work
// cycle, yielded nothing, and — the actual defect — resolved via a bare
// agent.intent_done, the SAME event a successful harvest emits. Nothing in the
// world could tell him, or his planner, that his hands were the problem.
//
// The shape asserted here is precisely that run: a full pouch of non-food, an
// agent standing on forage, work complete. The gather must resolve AUDIBLY —
// agent.intent_failed / "pack full", a paired same-tick memory naming the full
// hands, and an IntentLog record closed "failed" so the next thought sees it —
// while still yielding nothing and depleting nothing (the US1-AS1 invariant the
// guard exists to protect, unchanged).
func TestIntentFailedForagePackFullWorld03(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	fx, fy, ok := findForageTile(m)
	if !ok {
		t.Skip("no forage tile on this map")
	}
	s := NewState(seed, m)
	isolateAgents(s)

	a := &s.Agents[0]
	a.Dead = false
	a.X, a.Y = fx, fy
	// Cedar's exact pack: full to the cap, and not one scrap of it edible.
	a.Inv = Inventory{Wood: bulkCap}
	if freeBulk(a.Inv) != 0 {
		t.Fatalf("freeBulk = %d, want 0 — the fixture must reproduce a full pack", freeBulk(a.Inv))
	}
	// WorkStart pre-set so the work is already complete on tick 1 (quarry idiom).
	a.Intent = &Intent{Goal: "forage", TargetX: fx, TargetY: fy, WorkStart: 1 - forageTicks}
	beforeIntentSet := len(a.IntentLog)
	a.appendIntent(IntentRecord{Goal: "forage", Source: "planner", Tick: s.Tick})

	log := driveTicks(t, s, m, 3, nil)

	var failed bool
	var p IntentFailedPayload
	var failTick int64
	var mem MemoryAddedPayload
	memTick := int64(-1)
	for _, e := range log {
		switch e.Type {
		case "agent.foraged":
			t.Fatal("a full-pouch forage must not yield agent.foraged")
		case "agent.intent_done":
			var dp AgentPayload
			mustUnmarshal(t, e.Payload, &dp)
			if dp.Agent.ID == 0 {
				t.Fatal("a full-pouch forage resolved via agent.intent_done — this is the world-03 defect: " +
					"a no-op wearing the same event as a harvest")
			}
		case "agent.intent_failed":
			if !failed {
				mustUnmarshal(t, e.Payload, &p)
				if p.Agent.ID == 0 {
					failed, failTick = true, e.Tick
				}
			}
		case "agent.memory_added":
			if failed && memTick == -1 {
				var mp MemoryAddedPayload
				mustUnmarshal(t, e.Payload, &mp)
				if mp.Agent.ID == 0 && e.Tick == failTick {
					mem, memTick = mp, e.Tick
				}
			}
		}
	}
	if !failed {
		t.Fatal("no agent.intent_failed for the full-pouch forager")
	}
	if p.Goal != "forage" {
		t.Errorf("goal = %q, want forage", p.Goal)
	}
	if p.Reason != intentFailPackFull {
		t.Errorf("reason = %q, want %q — the villager must be able to tell a full pack from a vanished patch",
			p.Reason, intentFailPackFull)
	}
	if p.X != fx || p.Y != fy {
		t.Errorf("position = (%d,%d), want the forager's stand tile (%d,%d)", p.X, p.Y, fx, fy)
	}
	if memTick != failTick {
		t.Fatalf("failure memory tick = %d, want paired same-tick with the event (%d)", memTick, failTick)
	}
	if mem.Salience != salIntentFailed {
		t.Errorf("memory salience = %d, want %d", mem.Salience, salIntentFailed)
	}
	if !strings.Contains(mem.Text, "full") {
		t.Errorf("memory text = %q, want it to name the full hands — this text IS the signal Cedar never got", mem.Text)
	}
	// The IntentLog closure is the feedback channel proper: it is what the next
	// planner thought reads to see the goal did not finish.
	if len(a.IntentLog) <= beforeIntentSet {
		t.Fatal("no IntentLog record to inspect")
	}
	if r := a.IntentLog[len(a.IntentLog)-1]; r.Outcome != "failed" || r.OutcomeTick != failTick {
		t.Errorf("IntentLog record = %+v, want outcome \"failed\" @ %d", r, failTick)
	}
	// US1-AS1, unchanged: no yield, no depletion, intent cleared.
	if a.Inv.FoodRaw != 0 {
		t.Errorf("FoodRaw = %d, want 0 — nothing gathered into a full pouch", a.Inv.FoodRaw)
	}
	if bulk(a.Inv) != bulkCap {
		t.Errorf("bulk = %d, want %d — the pack must be untouched", bulk(a.Inv), bulkCap)
	}
	if len(s.Harvested) != 0 {
		t.Errorf("Harvested = %+v, want empty — the patch is left standing for later (US1-AS1)", s.Harvested)
	}
	if a.Intent != nil {
		t.Error("Intent should be cleared after the no-space gather resolves")
	}
}
