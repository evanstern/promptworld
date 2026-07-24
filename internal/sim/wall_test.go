package sim

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// openBlock scans for a 3x3 all-grass block centered on (cx,cy) — grass is both
// passable (a detour through the ring always exists when the center is walled)
// and buildable (buildSite accepts it, so a wall can actually be built there).
func openBlock(m *worldmap.Map) (cx, cy int, ok bool) {
	for y := 1; y < m.H-1; y++ {
		for x := 1; x < m.W-1; x++ {
			all := true
			for dy := -1; dy <= 1 && all; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if m.At(x+dx, y+dy) != worldmap.Grass {
						all = false
						break
					}
				}
			}
			if all {
				return x, y, true
			}
		}
	}
	return 0, 0, false
}

// TestWallBlocksAndReroutes is spec 032 US1 AC#2 / SC-002: a wall on the direct
// route makes its tile impassable, so nextStep detours around it and never
// steps onto the wall tile — while a route still exists.
func TestWallBlocksAndReroutes(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)

	cx, cy, ok := openBlock(m)
	if !ok {
		t.Fatal("no 3x3 open block found on seed 42")
	}
	// West-center → East-center, wall at the center: the direct step (the wall
	// tile) is barred, the detour through the ring is taken.
	fromX, fromY := cx-1, cy
	toX, toY := cx+1, cy

	// Without a wall the direct step is straight across (onto the center tile).
	if nx, ny := nextStep(m, s, fromX, fromY, toX, toY); nx != cx || ny != cy {
		t.Fatalf("baseline nextStep = (%d,%d), want the center tile (%d,%d)", nx, ny, cx, cy)
	}

	s.Structures = append(s.Structures, Structure{Kind: "wall_plank", X: cx, Y: cy, HP: wallPlankHP})
	if passable(m, s, cx, cy) {
		t.Fatal("a standing wall tile must be impassable")
	}
	nx, ny := nextStep(m, s, fromX, fromY, toX, toY)
	if nx == cx && ny == cy {
		t.Fatal("nextStep stepped onto the wall tile — it must detour")
	}
	if nx == fromX && ny == fromY {
		t.Fatal("nextStep froze — a detour exists through the open ring")
	}
	if !passable(m, s, nx, ny) {
		t.Errorf("nextStep chose an impassable tile (%d,%d)", nx, ny)
	}
}

// TestWallEnclosureUnreachable is spec 032 US1 (research R2): walls on all four
// neighbors of a tile enclose the agent standing on it — BFS finds no route, so
// nextStep returns the current tile (the caller resolves via intent_done).
func TestWallEnclosureUnreachable(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)

	cx, cy, ok := openBlock(m)
	if !ok {
		t.Fatal("no 3x3 open block found on seed 42")
	}
	// Wall the four orthogonal neighbors of the center: an agent on the center
	// is sealed in (diagonal moves are not allowed — 4-neighbor grid).
	for _, d := range neighborOrder {
		s.Structures = append(s.Structures, Structure{Kind: "wall_stone", X: cx + d[0], Y: cy + d[1], HP: wallStoneHP})
	}
	// A target two tiles east, outside the pen — unreachable.
	nx, ny := nextStep(m, s, cx, cy, cx+2, cy)
	if nx != cx || ny != cy {
		t.Errorf("enclosed agent's nextStep = (%d,%d), want its own tile (%d,%d) (no route)", nx, ny, cx, cy)
	}
}

// TestWallDetourReplayDeterminism is SC-005 over the passability change: a run
// with a wall placed (by direct injection, mirroring the structure-overlay
// idiom) replays to a byte-identical state hash — the wall scan in passable is
// tick-pure.
func TestWallDetourReplayDeterminism(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	cx, cy, ok := openBlock(NewState(seed, m).m)
	if !ok {
		t.Fatal("no open block on seed 42")
	}

	genesis := func() *State {
		s := NewState(seed, m)
		s.Structures = append(s.Structures, Structure{Kind: "wall_plank", X: cx, Y: cy, HP: wallPlankHP})
		return s
	}
	const ticks = 3000
	live := genesis()
	log := driveTicks(t, live, m, ticks, nil)

	replayed := genesis()
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replayed.Tick = e.Tick
	}
	driveTicks(t, replayed, m, ticks, nil)
	if live.Hash() != replayed.Hash() {
		t.Fatalf("replayed state diverged with a wall present")
	}
}

// reviveAt revives agent 0 on a healthy footing at (x,y), isolating the rest —
// the shared setup for the wall lifecycle scenarios below.
func reviveAt(s *State, x, y int) *Agent {
	isolateAgents(s)
	a := &s.Agents[0]
	a.Dead = false
	a.X, a.Y = x, y
	a.Needs = Needs{Health: 1000, Food: 900, Rest: 900, Warmth: 900, Morale: 900}
	return a
}

// TestWallBuildStampsHP is spec 032 US1 AC#1: a completed plank-wall build
// stands a wall at the Res tile with full derived HP and deducts the planks.
func TestWallBuildStampsHP(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	cx, cy, ok := openBlock(m)
	if !ok {
		t.Fatal("no open block")
	}
	a := reviveAt(s, cx-1, cy)
	a.Inv = Inventory{Planks: wallPlankCost}
	// Adjacent-stand: stand at (cx-1,cy), build at (cx,cy); preset WorkStart so
	// the first completion tick lands (quarry_test shortcut).
	a.Intent = &Intent{Goal: "build_wall_plank", TargetX: cx - 1, TargetY: cy, ResX: cx, ResY: cy, WorkStart: 1 - buildWallTicks}

	driveTicks(t, s, m, 5, nil)
	w := wallAt(s, cx, cy)
	if w == nil {
		t.Fatal("no wall stood at the Res tile")
	}
	if w.HP != wallPlankHP {
		t.Errorf("wall HP = %d, want %d (full at build)", w.HP, wallPlankHP)
	}
	if a.Inv.Planks != 0 {
		t.Errorf("planks = %d, want 0 (spent on the wall)", a.Inv.Planks)
	}
	if passable(m, s, cx, cy) {
		t.Error("the wall tile must be impassable")
	}
	if a.Intent != nil {
		t.Error("intent should clear after the build")
	}
}

// TestWallDemolishCycles is spec 032 US1 AC#6 / quickstart scenario 3: a plank
// wall falls in 2 demolish cycles (1 chip + destroy), a stone wall in 6 (5 chips
// + destroy), all under ONE intent via the WorkStart-reset loop; the tile is
// passable once the wall collapses.
func TestWallDemolishCycles(t *testing.T) {
	cases := []struct {
		kind      string
		hp        int
		wantChips int
	}{
		{"wall_plank", wallPlankHP, 1}, // 200 → 100 (chip) → 0 (destroy)
		{"wall_stone", wallStoneHP, 5}, // 600 → …→100 (5 chips) → 0 (destroy)
	}
	for _, c := range cases {
		t.Run(c.kind, func(t *testing.T) {
			const seed = 42
			m := testMap(seed)
			s := NewState(seed, m)
			cx, cy, ok := openBlock(m)
			if !ok {
				t.Fatal("no open block")
			}
			a := reviveAt(s, cx-1, cy)
			s.Structures = append(s.Structures, Structure{Kind: c.kind, X: cx, Y: cy, HP: c.hp})
			a.Intent = &Intent{Goal: "demolish", TargetX: cx - 1, TargetY: cy, ResX: cx, ResY: cy}

			// 6 cycles * demolishTicks + arrival/start overhead.
			log := driveTicks(t, s, m, int64(c.wantChips+1)*demolishTicks+50, nil)
			chips, destroys := 0, 0
			for _, e := range log {
				switch e.Type {
				case "agent.wall_chipped":
					chips++
				case "agent.wall_destroyed":
					destroys++
				}
			}
			if chips != c.wantChips {
				t.Errorf("%s: chips = %d, want %d", c.kind, chips, c.wantChips)
			}
			if destroys != 1 {
				t.Errorf("%s: destroys = %d, want 1", c.kind, destroys)
			}
			if wallAt(s, cx, cy) != nil {
				t.Errorf("%s: wall should be gone after demolition", c.kind)
			}
			if !passable(m, s, cx, cy) {
				t.Errorf("%s: collapsed wall's tile must be passable again", c.kind)
			}
			if a.Intent != nil {
				t.Errorf("%s: intent should clear on destroy", c.kind)
			}
		})
	}
}

// TestWallRepairMathAndReArm is spec 032 US1 AC#4 / quickstart scenario 3: a
// repair cycle consumes 1 matching material and restores repairHPPerUnit HP,
// clamped to the derived max; with material to spare and a still-damaged wall it
// runs a second cycle under one intent, ending at full health.
func TestWallRepairMathAndReArm(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	cx, cy, ok := openBlock(m)
	if !ok {
		t.Fatal("no open block")
	}
	a := reviveAt(s, cx-1, cy)
	// A plank wall chipped to 50 HP; two planks carried → two cycles: 50→150→200
	// (clamped), consuming both planks, then the intent clears at full health.
	s.Structures = append(s.Structures, Structure{Kind: "wall_plank", X: cx, Y: cy, HP: 50})
	a.Inv = Inventory{Planks: 2}
	a.Intent = &Intent{Goal: "repair", TargetX: cx - 1, TargetY: cy, ResX: cx, ResY: cy}

	log := driveTicks(t, s, m, 3*repairTicks+50, nil)
	repairs := 0
	for _, e := range log {
		if e.Type == "agent.wall_repaired" {
			repairs++
		}
	}
	if repairs != 2 {
		t.Errorf("repair events = %d, want 2 (two cycles under one intent)", repairs)
	}
	w := wallAt(s, cx, cy)
	if w == nil || w.HP != wallPlankHP {
		t.Fatalf("wall HP = %v, want %d (clamped to max)", w, wallPlankHP)
	}
	if a.Inv.Planks != 0 {
		t.Errorf("planks = %d, want 0 (1 per cycle)", a.Inv.Planks)
	}
	// repairs == 2 (not 3) proves the second cycle's clear-intent path fired: a
	// third cycle never started once the wall reached full health. (The intent
	// itself may be reflex-refilled over the long drive, so it is not asserted.)
}

// TestRepairFullHPNoResolve is spec 032 US1 / quickstart scenario 3: the repair
// resolver refuses a wall already at full health (nothing to repair), so the
// goal never produces an intent.
func TestRepairFullHPNoResolve(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	cx, cy, ok := openBlock(m)
	if !ok {
		t.Fatal("no open block")
	}
	a := reviveAt(s, cx-1, cy)
	a.Inv = Inventory{Planks: 5}
	s.Structures = append(s.Structures, Structure{Kind: "wall_plank", X: cx, Y: cy, HP: wallPlankHP})

	if _, _, err := resolveGoal(s, m, 0, "repair", -1, "", 0, 1); err == nil {
		t.Error("repair on a full-health wall should error (nothing to repair)")
	}
	// Damage it and the resolver now finds work.
	wallAt(s, cx, cy).HP = 100
	if in, _, err := resolveGoal(s, m, 0, "repair", -1, "", 0, 1); err != nil || in == nil {
		t.Errorf("repair on a damaged wall should resolve, got in=%v err=%v", in, err)
	}
}

// driveWithOccupant advances the sim tick by tick exactly like driveTicks, but
// BEFORE each tick it parks the occupant agent on the reserved tile (occupy) or
// moves it well clear — a deterministic way to script a passerby or a squatter
// crossing a wall's reserved tile across the build. The occupant is re-pinned at
// the START of every tick, so the builder (agent index 0, stepped first) always
// observes the scripted occupancy at its completion check. Occupancy is driven
// by direct placement rather than the reflex so the scenario is exact; these are
// guard-behavior tests, not replay tests (byte-identity is covered separately).
func driveWithOccupant(t *testing.T, s *State, m *worldmap.Map, ticks int64, occupant int, res, clear Point, occupy func(tick int64) bool) []store.Event {
	t.Helper()
	var log []store.Event
	for s.Tick < ticks {
		next := s.Tick + 1
		if occupy(next) {
			s.Agents[occupant].X, s.Agents[occupant].Y = res.X, res.Y
		} else {
			s.Agents[occupant].X, s.Agents[occupant].Y = clear.X, clear.Y
		}
		evs := stepEvents(s, m, next)
		s.Tick = next
		for _, e := range evs {
			if err := s.Apply(e); err != nil {
				t.Fatalf("apply %s at tick %d: %v", e.Type, s.Tick, err)
			}
		}
		log = append(log, evs...)
	}
	return log
}

// builderFailure scans a log for the builder's (agent 0) build-failure signal:
// the count of agent.build_failed, its reason and tick, whether a paired
// same-tick situated failure memory rode along, and the count of agent.built /
// bare agent.intent_done for the builder.
func builderFailure(t *testing.T, log []store.Event) (failed, built, done int, reason string, failTick int64, memory bool) {
	t.Helper()
	for _, e := range log {
		switch e.Type {
		case "agent.build_failed":
			var p BuildFailedPayload
			mustUnmarshal(t, e.Payload, &p)
			if p.Agent == 0 {
				failed++
				reason = p.Reason
				failTick = e.Tick
			}
		case "agent.built":
			built++
		case "agent.intent_done":
			var p AgentPayload
			mustUnmarshal(t, e.Payload, &p)
			if p.Agent == 0 {
				done++
			}
		case "agent.memory_added":
			var p MemoryAddedPayload
			mustUnmarshal(t, e.Payload, &p)
			if p.Agent == 0 && p.Origin == OriginAction && p.Salience == salShelter &&
				strings.Contains(p.Text, "never built") {
				memory = true
			}
		}
	}
	return
}

// TestWallOccupancyGuard is spec 038 US3 (FR-005/FR-006, SC-003): a squatter
// parked on the reserved tile at the completion moment DEFERS the build (never
// entomb) rather than cancelling it, and once the squatter outlasts the grace
// window the build fails LOUDLY — one agent.build_failed (site blocked too long)
// plus a paired failure memory — never a wall, never a spend, never silent.
func TestWallOccupancyGuard(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	cx, cy, ok := openBlock(m)
	if !ok {
		t.Fatal("no open block")
	}
	a := reviveAt(s, cx-1, cy)
	a.Inv = Inventory{Planks: wallPlankCost}
	// WorkStart preset so the due tick is tick 1; the loud fail then lands at
	// 1 + wallOccupancyGraceTicks (the bound is pure in WorkStart).
	a.Intent = &Intent{Goal: "build_wall_plank", TargetX: cx - 1, TargetY: cy, ResX: cx, ResY: cy, WorkStart: 1 - buildWallTicks}
	s.Agents[1].Dead = false
	res := Point{X: cx, Y: cy}
	clear := Point{X: cx + 1, Y: cy}
	wantFailTick := int64(1) + wallOccupancyGraceTicks

	// A permanent squatter: occupied on every tick.
	log := driveWithOccupant(t, s, m, wantFailTick+5, 1, res, clear, func(int64) bool { return true })

	failed, built, done, reason, failTick, memory := builderFailure(t, log)
	if failed != 1 {
		t.Fatalf("build_failed for the builder = %d, want exactly 1", failed)
	}
	if reason != buildFailSiteBlocked {
		t.Errorf("reason = %q, want %q", reason, buildFailSiteBlocked)
	}
	if built != 0 {
		t.Error("a wall must never be built onto an occupied tile (never entomb)")
	}
	if done != 0 {
		t.Error("a cancelled build must not resolve via a bare intent_done")
	}
	if !memory {
		t.Error("no paired situated failure memory for the builder")
	}
	if failTick != wantFailTick {
		t.Errorf("loud fail landed at tick %d, want %d (within the grace bound)", failTick, wantFailTick)
	}
	if wallAt(s, cx, cy) != nil {
		t.Error("no wall should stand on the blocked tile")
	}
	if a.Inv.Planks != wallPlankCost {
		t.Errorf("planks = %d, want %d (nothing spent on the failed build)", a.Inv.Planks, wallPlankCost)
	}
}

// TestWallBuildToleratesPasserby is spec 038 US3 (FR-004, SC-002): a second
// agent crossing the reserved tile MID-WORK and leaving before the due tick no
// longer kills the build — occupancy is ignored during work, so the wall
// completes normally, spends its planks, and never emits a build failure.
func TestWallBuildToleratesPasserby(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	cx, cy, ok := openBlock(m)
	if !ok {
		t.Fatal("no open block")
	}
	a := reviveAt(s, cx-1, cy)
	a.Inv = Inventory{Planks: wallPlankCost}
	// Due tick is 51 (work in progress for ticks 1..50); the passerby occupies
	// the reserved tile mid-work (ticks 10..30) and is gone well before due.
	a.Intent = &Intent{Goal: "build_wall_plank", TargetX: cx - 1, TargetY: cy, ResX: cx, ResY: cy, WorkStart: 51 - buildWallTicks}
	s.Agents[1].Dead = false
	res := Point{X: cx, Y: cy}
	clear := Point{X: cx + 1, Y: cy}

	log := driveWithOccupant(t, s, m, 60, 1, res, clear, func(tick int64) bool { return tick >= 10 && tick <= 30 })

	failed, built, _, _, _, _ := builderFailure(t, log)
	if failed != 0 {
		t.Fatalf("a passerby must not fail the build: build_failed = %d, want 0", failed)
	}
	if built != 1 {
		t.Fatalf("agent.built = %d, want 1 (the wall completes after the passerby leaves)", built)
	}
	w := wallAt(s, cx, cy)
	if w == nil || w.HP != wallPlankHP {
		t.Fatalf("wall = %v, want a plank wall at full HP", w)
	}
	if a.Inv.Planks != 0 {
		t.Errorf("planks = %d, want 0 (spent on the completed wall)", a.Inv.Planks)
	}
}

// TestWallBuildDefersThenCompletes is spec 038 US3 (FR-005, never-entomb): an
// occupant present AT the due tick defers completion, and once it leaves within
// the grace window the wall completes on the first clear tick — and never on any
// tick the reserved tile was occupied.
func TestWallBuildDefersThenCompletes(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	cx, cy, ok := openBlock(m)
	if !ok {
		t.Fatal("no open block")
	}
	a := reviveAt(s, cx-1, cy)
	a.Inv = Inventory{Planks: wallPlankCost}
	// Due at tick 1; the occupant sits on the reserved tile ticks 1..50 (inside
	// the grace window) then leaves, so completion fires the first clear tick.
	a.Intent = &Intent{Goal: "build_wall_plank", TargetX: cx - 1, TargetY: cy, ResX: cx, ResY: cy, WorkStart: 1 - buildWallTicks}
	s.Agents[1].Dead = false
	res := Point{X: cx, Y: cy}
	clear := Point{X: cx + 1, Y: cy}
	const leaveAfter = 50

	log := driveWithOccupant(t, s, m, 60, 1, res, clear, func(tick int64) bool { return tick <= leaveAfter })

	var builtTicks []int64
	failed := 0
	for _, e := range log {
		switch e.Type {
		case "agent.built":
			builtTicks = append(builtTicks, e.Tick)
		case "agent.build_failed":
			failed++
		}
	}
	if failed != 0 {
		t.Fatalf("build_failed = %d, want 0 (the occupant left within the grace window)", failed)
	}
	if len(builtTicks) != 1 {
		t.Fatalf("agent.built count = %d, want exactly 1", len(builtTicks))
	}
	// Never entomb: completion never landed on a tick the tile was occupied.
	if builtTicks[0] <= leaveAfter {
		t.Errorf("wall built at tick %d, on/inside the occupied window (never entomb)", builtTicks[0])
	}
	if builtTicks[0] != leaveAfter+1 {
		t.Errorf("wall built at tick %d, want the first clear tick %d", builtTicks[0], leaveAfter+1)
	}
	if w := wallAt(s, cx, cy); w == nil || w.HP != wallPlankHP {
		t.Errorf("wall = %v, want a plank wall at full HP after the deferral cleared", w)
	}
	if a.Inv.Planks != 0 {
		t.Errorf("planks = %d, want 0 (spent once the deferral cleared)", a.Inv.Planks)
	}
}

// TestWallBuildSiteVanishedFailsLoud is spec 038 US1/US2 (FR-001/FR-003/FR-007,
// SC-001): a wall build whose reserved-tile SITE becomes invalid mid-work fails
// LOUDLY — one agent.build_failed (site no longer buildable) with the builder and
// goal, a paired first-person failure memory, no bare intent_done, no wall, and
// no material spent. Site loss is a genuine failure, never something to wait out.
func TestWallBuildSiteVanishedFailsLoud(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	cx, cy, ok := openBlock(m)
	if !ok {
		t.Fatal("no open block")
	}
	a := reviveAt(s, cx-1, cy)
	a.Inv = Inventory{RefinedStone: wallStoneCost}
	a.Intent = &Intent{Goal: "build_wall_stone", TargetX: cx - 1, TargetY: cy, ResX: cx, ResY: cy, WorkStart: 1 - buildWallTicks}
	// The reserved tile is no longer buildable — a structure now sits on it.
	s.Structures = append(s.Structures, Structure{Kind: "shelter", X: cx, Y: cy})

	log := driveTicks(t, s, m, 3, nil)

	failed := 0
	var reason, goal string
	var failTick int64
	memTick := int64(-1)
	built, done := 0, 0
	for _, e := range log {
		switch e.Type {
		case "agent.build_failed":
			var p BuildFailedPayload
			mustUnmarshal(t, e.Payload, &p)
			if p.Agent == 0 {
				failed++
				reason, goal, failTick = p.Reason, p.Goal, e.Tick
			}
		case "agent.built":
			built++
		case "agent.intent_done":
			var p AgentPayload
			mustUnmarshal(t, e.Payload, &p)
			if p.Agent == 0 {
				done++
			}
		case "agent.memory_added":
			var p MemoryAddedPayload
			mustUnmarshal(t, e.Payload, &p)
			if p.Agent == 0 && p.Origin == OriginAction && p.Salience == salShelter &&
				strings.Contains(p.Text, "never built") && strings.Contains(p.Text, "stone wall") {
				memTick = e.Tick
			}
		}
	}
	if failed != 1 {
		t.Fatalf("build_failed for the builder = %d, want exactly 1", failed)
	}
	if goal != "build_wall_stone" {
		t.Errorf("goal = %q, want build_wall_stone", goal)
	}
	if reason != buildFailSiteUnbuildable {
		t.Errorf("reason = %q, want %q", reason, buildFailSiteUnbuildable)
	}
	if built != 0 {
		t.Error("no wall should be built on a vanished site")
	}
	if done != 0 {
		t.Error("a vanished-site build must not resolve via a bare intent_done")
	}
	if memTick != failTick {
		t.Errorf("failure memory tick = %d, want it paired same-tick with the event (%d)", memTick, failTick)
	}
	if a.Inv.RefinedStone != wallStoneCost {
		t.Errorf("refined stone = %d, want %d (unspent on the failed build)", a.Inv.RefinedStone, wallStoneCost)
	}
}

// TestNoAgentEverOnWallTile is spec 032 SC-002: with a wall standing, no living
// agent ever occupies its tile across a long reflex-driven run.
func TestNoAgentEverOnWallTile(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	cx, cy, ok := openBlock(m)
	if !ok {
		t.Fatal("no open block")
	}
	s.Structures = append(s.Structures, Structure{Kind: "wall_stone", X: cx, Y: cy, HP: wallStoneHP})

	const ticks = 6000
	driveTicks(t, s, m, ticks, nil)
	// Final positions never coincide with the wall (the per-step BFS treats it as
	// impassable, so no agent.moved ever lands there either).
	for i := range s.Agents {
		if !s.Agents[i].Dead && s.Agents[i].X == cx && s.Agents[i].Y == cy {
			t.Fatalf("agent %d ended on the wall tile (%d,%d)", i, cx, cy)
		}
	}
}

// TestWallLifecycleReplay is quickstart scenario 7 over US1: a run that builds
// then fully demolishes a wall (all via planner-sourced intent_set) replays from
// genesis to a byte-identical state hash.
func TestWallLifecycleReplay(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	base := NewState(seed, m)
	cx, cy, ok := openBlock(m)
	if !ok {
		t.Fatal("no open block")
	}
	stand := Point{X: cx - 1, Y: cy}

	genesis := func() *State {
		s := NewState(seed, m)
		reviveAt(s, stand.X, stand.Y)
		s.Agents[0].Inv = Inventory{Planks: wallPlankCost}
		return s
	}
	_ = base
	pl := func(v any) []byte { return mustPayload(v) }
	commands := map[int64][]store.Event{
		0: {{Tick: 0, Type: "agent.intent_set", Payload: pl(IntentSetPayload{
			Agent: 0, Goal: "build_wall_plank", TargetX: stand.X, TargetY: stand.Y, ResX: cx, ResY: cy, Source: "planner"})}},
		// The build completes at ~tick 601 (agent starts on the stand tile). Land
		// demolish at 700 — inside the reflex grace (idle < 120), so the reflex
		// hasn't repurposed the still-adjacent builder — then it tears the wall
		// back down to open ground.
		700: {{Tick: 700, Type: "agent.intent_set", Payload: pl(IntentSetPayload{
			Agent: 0, Goal: "demolish", TargetX: stand.X, TargetY: stand.Y, ResX: cx, ResY: cy, Source: "planner"})}},
	}

	const ticks = 4000
	live := genesis()
	log := driveTicks(t, live, m, ticks, commands)

	var built, destroyed bool
	for _, e := range log {
		switch e.Type {
		case "agent.built":
			built = true
		case "agent.wall_destroyed":
			destroyed = true
		}
	}
	if !built || !destroyed {
		t.Fatalf("run did not exercise build+demolish: built=%v destroyed=%v", built, destroyed)
	}
	if wallAt(live, cx, cy) != nil {
		t.Error("wall should be demolished by the end of the run")
	}

	replay := genesis()
	for _, e := range log {
		if err := replay.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replay.Tick = e.Tick
	}
	driveTicks(t, replay, m, ticks, nil)
	if live.Hash() != replay.Hash() {
		t.Fatalf("wall lifecycle replay diverged:\nlive:     %s\nreplayed: %s",
			string(live.Marshal()), string(replay.Marshal()))
	}
}
