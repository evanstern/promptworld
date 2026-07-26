package sim

// Spec 084 US4 suites (T020–T022): the DIRECTIVE rung matrix (survival
// preempts / directive preempts prep and wander / per-kind routing cells /
// orphan fall-through / inertness), the hail interruption-resume drive
// (zero new interruption code — AC #6/SC-002), and the reflex-only
// end-to-end lifecycle (SC-001).

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// directiveRungAgent configures agent 0 as a fed, warm, rested, awake villager
// by DAY — the state in which survival declines and the DIRECTIVE rung is the
// first decider (spec 084 FR-012).
func directiveRungAgent(t *testing.T, seed uint64) (*State, *worldmap.Map, *Agent) {
	t.Helper()
	m := testMap(seed)
	s := NewState(seed, m)
	s.Night = false
	s.Tick = 30000 // mid-morning, clear of every genesis special case
	a := &s.Agents[0]
	a.Needs = Needs{Health: 1000, Food: 600, Rest: 600, Warmth: 600, Morale: 600}
	a.Inv = Inventory{}
	return s, m, a
}

// bindDirective installs an active designation + directive addressing agent 0
// directly in state (the rung is a pure state reader; the door discipline is
// proven in plans_test.go).
func bindDirective(s *State, d Designation) {
	d.ID = "dsg-t-0"
	d.Status = "active"
	s.Designations = append(s.Designations, d)
	s.Directives = append(s.Directives, Directive{
		ID: "dir-t-0", DesignationID: "dsg-t-0", Targets: []int{0}, Text: "Do it.",
		IssuedTick: s.Tick, ExpiresTick: s.Tick + 3*ticksPerGameDay, Status: "active"})
}

// nearBuildSite finds a deterministic buildable tile near the agent.
func nearBuildSite(t *testing.T, s *State, m *worldmap.Map, a *Agent) Point {
	t.Helper()
	p, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool { return buildSite(m, s, x, y) })
	if !ok {
		t.Fatal("no build site near the agent")
	}
	return p
}

// farBuildSite finds a deterministic buildable tile at least min tiles away.
func farBuildSite(t *testing.T, s *State, m *worldmap.Map, a *Agent, min int) Point {
	t.Helper()
	p, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool {
		return buildSite(m, s, x, y) && abs(x-a.X)+abs(y-a.Y) >= min
	})
	if !ok {
		t.Fatal("no distant build site")
	}
	return p
}

// TestDirectiveRungRoutingMatrix walks the data-model §8 routing table cell by
// cell — every assertion calls decideIntent directly (the reflex matrix
// pattern) with no planner or model involved.
func TestDirectiveRungRoutingMatrix(t *testing.T) {
	t.Run("structure_site build with materials", func(t *testing.T) {
		s, m, a := directiveRungAgent(t, 42)
		site := farBuildSite(t, s, m, a, 5)
		bindDirective(s, Designation{Kind: DesignationStructureSite, X: site.X, Y: site.Y, X2: site.X, Y2: site.Y, StructureKind: "fire"})
		a.Inv.Wood = fireWoodCost
		d := decideIntent(s, m, 0, s.Tick)
		if d.intent == nil || d.intent.Goal != "build_fire" || d.intent.TargetX != site.X || d.intent.TargetY != site.Y {
			t.Fatalf("decision = %+v, want build_fire at the site", d.intent)
		}
	})
	t.Run("structure_site walk when not reflex-expressible", func(t *testing.T) {
		s, m, a := directiveRungAgent(t, 42)
		site := farBuildSite(t, s, m, a, 5)
		bindDirective(s, Designation{Kind: DesignationStructureSite, X: site.X, Y: site.Y, X2: site.X, Y2: site.Y, StructureKind: "shelter"})
		a.Inv.Planks = 100 // materials irrelevant: shelter is planner-only
		d := decideIntent(s, m, 0, s.Tick)
		if d.intent == nil || d.intent.Goal != "heed_directive" || d.intent.TargetX != site.X || d.intent.TargetY != site.Y {
			t.Fatalf("decision = %+v, want heed_directive to the site", d.intent)
		}
	})
	t.Run("structure_site walk when materials missing", func(t *testing.T) {
		s, m, a := directiveRungAgent(t, 42)
		site := farBuildSite(t, s, m, a, 5)
		bindDirective(s, Designation{Kind: DesignationStructureSite, X: site.X, Y: site.Y, X2: site.X, Y2: site.Y, StructureKind: "fire"})
		d := decideIntent(s, m, 0, s.Tick) // no wood in hand
		if d.intent == nil || d.intent.Goal != "heed_directive" {
			t.Fatalf("decision = %+v, want heed_directive", d.intent)
		}
	})
	t.Run("at site without means falls through", func(t *testing.T) {
		s, m, a := directiveRungAgent(t, 42)
		site := nearBuildSite(t, s, m, a)
		a.X, a.Y = site.X, site.Y
		bindDirective(s, Designation{Kind: DesignationStructureSite, X: site.X, Y: site.Y, X2: site.X, Y2: site.Y, StructureKind: "shelter"})
		d := decideIntent(s, m, 0, s.Tick)
		if d.intent != nil && (d.intent.Goal == "heed_directive" || d.intent.Goal == "build_shelter") {
			t.Fatalf("decision = %+v, want fall-through to the ladder (the planner's job)", d.intent)
		}
	})
	t.Run("wall_line builds the first gap with materials", func(t *testing.T) {
		s, m, a := directiveRungAgent(t, 42)
		site := farBuildSite(t, s, m, a, 4)
		// A 3-tile horizontal line; a wall already stands on the FIRST tile,
		// so the first gap in enumeration order is the second.
		lineY := site.Y
		bindDirective(s, Designation{Kind: DesignationWallLine, X: site.X, Y: lineY, X2: site.X + 2, Y2: lineY})
		s.Structures = append(s.Structures, Structure{Kind: "wall_plank", X: site.X, Y: lineY, HP: wallMaxHP("wall_plank")})
		a.Inv.Planks = wallPlankCost
		d := decideIntent(s, m, 0, s.Tick)
		if d.intent == nil || d.intent.Goal != "build_wall_plank" {
			t.Fatalf("decision = %+v, want build_wall_plank", d.intent)
		}
		if d.intent.ResX != site.X+1 || d.intent.ResY != lineY {
			t.Fatalf("wall lands at (%d,%d), want the first gap (%d,%d)", d.intent.ResX, d.intent.ResY, site.X+1, lineY)
		}
	})
	t.Run("wall_line walks toward the gap without materials", func(t *testing.T) {
		s, m, a := directiveRungAgent(t, 42)
		site := farBuildSite(t, s, m, a, 5)
		bindDirective(s, Designation{Kind: DesignationWallLine, X: site.X, Y: site.Y, X2: site.X + 2, Y2: site.Y})
		d := decideIntent(s, m, 0, s.Tick)
		if d.intent == nil || d.intent.Goal != "heed_directive" {
			t.Fatalf("decision = %+v, want heed_directive toward the line", d.intent)
		}
	})
	t.Run("wall_line fulfilled falls through", func(t *testing.T) {
		s, m, a := directiveRungAgent(t, 42)
		site := farBuildSite(t, s, m, a, 4)
		bindDirective(s, Designation{Kind: DesignationWallLine, X: site.X, Y: site.Y, X2: site.X + 1, Y2: site.Y})
		for _, p := range []Point{{site.X, site.Y}, {site.X + 1, site.Y}} {
			s.Structures = append(s.Structures, Structure{Kind: "wall_plank", X: p.X, Y: p.Y, HP: wallMaxHP("wall_plank")})
		}
		a.Inv.Planks = wallPlankCost
		d := decideIntent(s, m, 0, s.Tick)
		if d.intent != nil && (d.intent.Goal == "heed_directive" || d.intent.Goal == "build_wall_plank") {
			t.Fatalf("decision = %+v, want fall-through (line complete; the sweep stamps it)", d.intent)
		}
	})
	t.Run("zone walks in when outside", func(t *testing.T) {
		s, m, a := directiveRungAgent(t, 42)
		// A rect that deterministically excludes the agent.
		zx, zy := (a.X+20)%60, (a.Y+20)%60
		if zx <= a.X+2 && zx+3 >= a.X-2 && zy <= a.Y+2 && zy+3 >= a.Y-2 {
			zx, zy = (zx+10)%60, (zy+10)%60
		}
		bindDirective(s, Designation{Kind: DesignationSettlementZone, X: zx, Y: zy, X2: zx + 3, Y2: zy + 3, MinStructures: 3})
		d := decideIntent(s, m, 0, s.Tick)
		if d.intent == nil || d.intent.Goal != "heed_directive" {
			t.Fatalf("decision = %+v, want heed_directive into the zone", d.intent)
		}
		if d.intent.TargetX < zx || d.intent.TargetX > zx+3 || d.intent.TargetY < zy || d.intent.TargetY > zy+3 {
			t.Fatalf("walk target (%d,%d) outside the zone", d.intent.TargetX, d.intent.TargetY)
		}
	})
	t.Run("zone presence achieved falls through", func(t *testing.T) {
		s, m, a := directiveRungAgent(t, 42)
		bindDirective(s, Designation{Kind: DesignationSettlementZone, X: a.X - 1, Y: a.Y - 1, X2: a.X + 1, Y2: a.Y + 1, MinStructures: 3})
		d := decideIntent(s, m, 0, s.Tick)
		if d.intent != nil && d.intent.Goal == "heed_directive" {
			t.Fatalf("decision = %+v, want fall-through (presence achieved; what to build is mind work)", d.intent)
		}
	})
	t.Run("orphaned directive falls through", func(t *testing.T) {
		s, m, a := directiveRungAgent(t, 42)
		site := farBuildSite(t, s, m, a, 5)
		bindDirective(s, Designation{Kind: DesignationStructureSite, X: site.X, Y: site.Y, X2: site.X, Y2: site.Y, StructureKind: "fire"})
		s.Designations[0].Status = "cancelled" // orphan: the sweep will expire it
		a.Inv.Wood = fireWoodCost
		d := decideIntent(s, m, 0, s.Tick)
		// Fall-through means the LADDER owns the decision: the first-fire
		// prep rung legitimately builds NEAR THE AGENT with that wood — what
		// must never happen is the rung serving the orphaned site.
		if d.intent != nil && (d.intent.Goal == "heed_directive" ||
			(d.intent.Goal == "build_fire" && d.intent.TargetX == site.X && d.intent.TargetY == site.Y)) {
			t.Fatalf("decision = %+v, want fall-through for an orphan (never the site %v)", d.intent, site)
		}
	})
	t.Run("oldest directive wins", func(t *testing.T) {
		s, m, a := directiveRungAgent(t, 42)
		siteA := farBuildSite(t, s, m, a, 5)
		bindDirective(s, Designation{Kind: DesignationStructureSite, X: siteA.X, Y: siteA.Y, X2: siteA.X, Y2: siteA.Y, StructureKind: "shelter"})
		s.Designations = append(s.Designations, Designation{ID: "dsg-t-1", Kind: DesignationSettlementZone,
			X: 1, Y: 1, X2: 3, Y2: 3, MinStructures: 3, Status: "active"})
		s.Directives = append(s.Directives, Directive{ID: "dir-t-1", DesignationID: "dsg-t-1",
			Targets: []int{0}, Text: "Second.", IssuedTick: s.Tick + 1, ExpiresTick: s.Tick + ticksPerGameDay, Status: "active"})
		d := decideIntent(s, m, 0, s.Tick)
		if d.intent == nil || d.intent.TargetX != siteA.X || d.intent.TargetY != siteA.Y {
			t.Fatalf("decision = %+v, want the OLDEST directive's site (%d,%d)", d.intent, siteA.X, siteA.Y)
		}
	})
}

// TestDirectiveRungSurvivalPreempts (US4 AS2, AC #5): a survival need below
// its band decides FIRST — the rung sits after survivalDecision, so the
// directive waits while the villager keeps itself alive.
func TestDirectiveRungSurvivalPreempts(t *testing.T) {
	s, m, a := directiveRungAgent(t, 42)
	site := farBuildSite(t, s, m, a, 5)
	bindDirective(s, Designation{Kind: DesignationStructureSite, X: site.X, Y: site.Y, X2: site.X, Y2: site.Y, StructureKind: "fire"})
	a.Inv.Wood = fireWoodCost
	a.Needs.Food = hungryAt - 1
	a.Inv.FoodRaw = 2
	d := decideIntent(s, m, 0, s.Tick)
	if d.directEvent != "agent.ate" {
		t.Fatalf("decision = %+v, want the survival eat rung first", d)
	}
}

// TestDirectiveRungPreemptsPrepAndWander (US4 AS1, AC #5): when the rung
// resolves, prep and wander never run — even a prep rung that would otherwise
// fire (first-fire with wood in hand) defers to the directive.
func TestDirectiveRungPreemptsPrepAndWander(t *testing.T) {
	s, m, a := directiveRungAgent(t, 42)
	a.Inv.Wood = fireWoodCost // first-fire prep would build right here

	// Without a directive: prep owns the decision (build_fire near the agent).
	base := decideIntent(s, m, 0, s.Tick)
	if base.intent == nil || base.intent.Goal != "build_fire" {
		t.Fatalf("fixture broken: prep decision = %+v, want build_fire", base.intent)
	}

	// With a zone directive elsewhere: the rung preempts prep entirely.
	zx, zy := (a.X+25)%58, (a.Y+25)%58
	bindDirective(s, Designation{Kind: DesignationSettlementZone, X: zx, Y: zy, X2: zx + 2, Y2: zy + 2, MinStructures: 3})
	if a.X >= zx && a.X <= zx+2 && a.Y >= zy && a.Y <= zy+2 {
		t.Skip("agent landed inside the test zone (seed geometry)")
	}
	d := decideIntent(s, m, 0, s.Tick)
	if d.intent == nil || d.intent.Goal != "heed_directive" {
		t.Fatalf("decision = %+v, want the directive to preempt prep", d.intent)
	}
}

// TestDirectiveRungInert (SC-006, US4 AS6): with no active directive the rung
// decides nothing — decisions are byte-identical to a pre-084 world (the
// wider guarantee rides the whole untouched reflex suite; this pins the
// non-active-entities case explicitly).
func TestDirectiveRungInert(t *testing.T) {
	s, m, a := directiveRungAgent(t, 42)
	a.Inv.Wood = fireWoodCost
	base := decideIntent(s, m, 0, s.Tick)

	s2, m2, a2 := directiveRungAgent(t, 42)
	a2.Inv.Wood = fireWoodCost
	site := farBuildSite(t, s2, m2, a2, 5)
	bindDirective(s2, Designation{Kind: DesignationStructureSite, X: site.X, Y: site.Y, X2: site.X, Y2: site.Y, StructureKind: "fire"})
	s2.Directives[0].Status = "fulfilled" // consumed — the rung must skip it
	got := decideIntent(s2, m2, 0, s2.Tick)

	if !bytes.Equal(mustPayload(base.intent), mustPayload(got.intent)) || base.directEvent != got.directEvent {
		t.Fatalf("consumed-directive decision %+v differs from directive-free %+v", got.intent, base.intent)
	}
}

// TestDirectiveHailInterruptResume (AC #6 / SC-002, T021): a hail pauses a
// directed walk exactly as it pauses any intent — the intent bytes survive
// the window untouched and movement resumes at expiry — and the rung, being
// stateless, re-resolves toward the SAME site at the next idle decision after
// any interruption that consumed the intent. ZERO new interruption code: this
// test exercises only the pre-084 hail/pause machinery (the diff obligation
// is recorded in the PR body).
func TestDirectiveHailInterruptResume(t *testing.T) {
	s, m, a := directiveRungAgent(t, 42)
	site := farBuildSite(t, s, m, a, 8)
	bindDirective(s, Designation{Kind: DesignationStructureSite, X: site.X, Y: site.Y, X2: site.X, Y2: site.Y, StructureKind: "shelter"})

	// The rung resolves the walk leg; land it as the executor would.
	d := decideIntent(s, m, 0, s.Tick)
	if d.intent == nil || d.intent.Goal != "heed_directive" {
		t.Fatalf("decision = %+v, want heed_directive", d.intent)
	}
	if err := s.Apply(store.Event{Tick: s.Tick, Type: "agent.intent_set", Payload: mustPayload(IntentSetPayload{
		Agent: 0, Goal: d.intent.Goal, TargetX: d.intent.TargetX, TargetY: d.intent.TargetY, Source: "reflex"})}); err != nil {
		t.Fatal(err)
	}
	start := s.Tick

	// Walk a little, then a hail freezes the villager mid-walk.
	driveTicks(t, s, m, start+60, nil)
	intentBefore := mustPayload(s.Agents[0].Intent)
	s.Agents[0].Hail = &AgentHail{By: 1, Until: s.Tick + hailWindowTicks}

	pre := driveTicks(t, s, m, s.Tick+hailWindowTicks-1, nil)
	if countAgentType(pre, "agent.moved", "agent", 0) != 0 {
		t.Error("directed walk moved during the hail pause")
	}
	if !bytes.Equal(mustPayload(s.Agents[0].Intent), intentBefore) {
		t.Error("the directed intent changed across the pause (FR-004 held for every other intent)")
	}
	post := driveTicks(t, s, m, s.Tick+600, nil)
	if countAgentType(post, "agent.moved", "agent", 0) == 0 {
		t.Error("directed walk did not resume after the hail expired")
	}

	// An interruption that CONSUMED the intent (a conversation closing it):
	// the next idle decision re-resolves the same directive toward the same
	// site — resumption is a free consequence of the rung's statelessness.
	// Needs are topped back up (survival preempting a hungry villager is its
	// own correct doctrine, proven separately) and the villager is stood back
	// away from the site — the post-expiry drive may already have walked it
	// there, and an at-site villager correctly falls through to the ladder.
	s.Agents[0].Intent = nil
	s.Agents[0].Needs = Needs{Health: 1000, Food: 900, Rest: 900, Warmth: 900, Morale: 600}
	if p, ok := nearest(m, s, site.X, site.Y, func(x, y int) bool {
		return passable(m, s, x, y) && abs(x-site.X)+abs(y-site.Y) >= 6
	}); ok {
		s.Agents[0].X, s.Agents[0].Y = p.X, p.Y
	}
	again := decideIntent(s, m, 0, s.Tick)
	if again.intent == nil || again.intent.Goal != "heed_directive" ||
		again.intent.TargetX != site.X || again.intent.TargetY != site.Y {
		t.Fatalf("re-resolution = %+v, want heed_directive to the same site (%d,%d)", again.intent, site.X, site.Y)
	}
}

// TestDirectiveReflexEndToEnd (SC-001, T022): in a reflex-only drive (no
// planner anywhere), an issued structure-site directive takes a fed, warm
// villager to the site, the fire goes up, the designation fulfills via the
// sweep, and directive.fulfilled lands — all recorded, all replayable.
func TestDirectiveReflexEndToEnd(t *testing.T) {
	const seed = 84
	m := testMap(seed)
	probe := NewState(seed, m)
	site, ok := nearest(m, probe, probe.Agents[0].X, probe.Agents[0].Y, func(x, y int) bool {
		return buildSite(m, probe, x, y) && abs(x-probe.Agents[0].X)+abs(y-probe.Agents[0].Y) >= 4
	})
	if !ok {
		t.Fatal("no site near agent 0")
	}

	// Day tick 22000 (06:06): place the designation + directive, then grant
	// the wood (AFTER the directive, so the first-fire prep rung cannot spend
	// it first — the genesis charge pays for the grant).
	dsg := Designation{ID: "dsg-22000-0", Kind: DesignationStructureSite,
		X: site.X, Y: site.Y, X2: site.X, Y2: site.Y, StructureKind: "fire", PlacedTick: 22000}
	dir := Directive{ID: "dir-22050-0", DesignationID: "dsg-22000-0", Targets: []int{0},
		Text: "Raise a fire where I have marked.", IssuedTick: 22050, ExpiresTick: 22050 + 3*ticksPerGameDay}
	timeline := map[int64][]store.Event{
		22000: {{Tick: 22000, Type: "designation.placed", Payload: mustPayload(dsg)}},
		22050: {{Tick: 22050, Type: "directive.issued", Payload: mustPayload(dir)}},
		22100: {{Tick: 22100, Type: "metatron.item_granted", Payload: mustPayload(ItemGrantedPayload{Agent: 0, Kind: "wood", Qty: fireWoodCost})}},
	}

	s := NewState(seed, m)
	log := driveTicks(t, s, m, 60000, timeline)

	var builtAtSite, reflexIntent bool
	for _, e := range log {
		if e.Type == "agent.built" {
			var p BuiltPayload
			if json.Unmarshal(e.Payload, &p) == nil && p.Agent == 0 && p.Kind == "fire" && p.X == site.X && p.Y == site.Y {
				builtAtSite = true
			}
		}
		if e.Type == "agent.intent_set" {
			var p IntentSetPayload
			if json.Unmarshal(e.Payload, &p) == nil && p.Agent == 0 && p.Source == "reflex" &&
				(p.Goal == "build_fire" || p.Goal == "heed_directive") &&
				p.TargetX == site.X && p.TargetY == site.Y && e.Tick >= 22100 {
				reflexIntent = true
			}
		}
	}
	if !reflexIntent {
		t.Error("no reflex-sourced directive intent recorded (the rung never fired)")
	}
	if !builtAtSite {
		t.Fatal("the fire never went up at the site")
	}
	if countType(log, "designation.fulfilled") != 1 {
		t.Fatalf("designation.fulfilled recorded %d times, want 1", countType(log, "designation.fulfilled"))
	}
	if countType(log, "directive.fulfilled") != 1 {
		t.Fatalf("directive.fulfilled recorded %d times, want 1", countType(log, "directive.fulfilled"))
	}
	if s.designationByID("dsg-22000-0").Status != "fulfilled" || s.directiveByID("dir-22050-0").Status != "fulfilled" {
		t.Errorf("terminal statuses: dsg %q, dir %q, want fulfilled/fulfilled",
			s.designationByID("dsg-22000-0").Status, s.directiveByID("dir-22050-0").Status)
	}
}
