package sim

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tool"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// TestMiracleCostDerivedFromTool (spec 021 T004 / SC-004): the reducer's cost
// table IS tool.MiracleCostsByEvent() — a single authoritative source, not a
// second local copy. If anyone re-introduces a hand-written literal here that
// disagrees with the one table in internal/tool, this fails. It also pins the
// price doctrine (time_snap 2, others 1) against the event-keyed shape the
// reducer keys on.
func TestMiracleCostDerivedFromTool(t *testing.T) {
	if !reflect.DeepEqual(miracleCost, tool.MiracleCostsByEvent()) {
		t.Errorf("sim.miracleCost = %v, want the derived tool.MiracleCostsByEvent() = %v",
			miracleCost, tool.MiracleCostsByEvent())
	}
	want := map[string]int{
		"guardian.time_snapped":   2,
		"guardian.entity_moved":   1,
		"guardian.entity_removed": 1,
		"guardian.item_granted":   1,
	}
	if !reflect.DeepEqual(miracleCost, want) {
		t.Errorf("sim.miracleCost = %v, want %v", miracleCost, want)
	}
}

// TestGrantKindsMirrorTool (TASK-163 drift guard, the TestMiracleCostDerivedFromTool
// pattern): grantableKind — the give_item door's accept set — is derived from
// tool.GrantKinds(), the single authoritative grant vocabulary, not a second
// hand-written literal here. Accepts every listed kind, rejects the guessed
// forms a live guardian actually tried ("food", "forage") and the empty string.
func TestGrantKindsMirrorTool(t *testing.T) {
	for _, k := range tool.GrantKinds() {
		if !grantableKind(k) {
			t.Errorf("tool.GrantKinds() lists %q but grantableKind rejects it", k)
		}
	}
	for _, bad := range []string{"food", "forage", "", "spears", "axes", "gold"} {
		if grantableKind(bad) {
			t.Errorf("grantableKind(%q) = true, want false (not in tool.GrantKinds())", bad)
		}
	}
	if len(tool.GrantKinds()) != 10 {
		t.Errorf("tool.GrantKinds() = %v, want the ten grantable kinds", tool.GrantKinds())
	}
}

// Guardian miracles (spec 016 US1): the entity move/remove reducer arms.
// validate-not-clamp, reject-whole, no charge spent on rejection, no partial
// application, and a scripted move+remove sequence replays byte-identically.

// applyMiracleErr applies a miracle event and returns the reducer error (nil on
// success) — the reject cases need the error, so they cannot use applyEvent
// (which fails the test on any error).
func applyMiracleErr(s *State, tick int64, typ string, pl any) error {
	return s.Apply(store.Event{Tick: tick, Type: typ, Payload: mustPayload(pl)})
}

// passableTileExcept finds a passable tile not in the excluded set.
func passableTileExcept(m *worldmap.Map, s *State, ex ...Point) (Point, bool) {
	skip := map[Point]bool{}
	for _, p := range ex {
		skip[p] = true
	}
	for y := 0; y < m.H; y++ {
		for x := 0; x < m.W; x++ {
			p := Point{X: x, Y: y}
			if !skip[p] && passable(m, s, x, y) {
				return p, true
			}
		}
	}
	return Point{}, false
}

// firstTileOfKind finds a tile whose static base kind is k.
func firstTileOfKind(m *worldmap.Map, k worldmap.TileKind) (Point, bool) {
	for y := 0; y < m.H; y++ {
		for x := 0; x < m.W; x++ {
			if m.At(x, y) == k {
				return Point{X: x, Y: y}, true
			}
		}
	}
	return Point{}, false
}

func TestMiracleMoveVillager(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 3
	a := &s.Agents[0]
	src := Point{X: a.X, Y: a.Y}
	dst, ok := passableTileExcept(m, s, src)
	if !ok {
		t.Skip("no spare passable tile")
	}
	a.Intent = &Intent{Goal: "forage", TargetX: 9, TargetY: 9}

	if err := applyMiracleErr(s, 100, "guardian.entity_moved", EntityMovedPayload{
		Class: "villager", X: src.X, Y: src.Y, ToX: dst.X, ToY: dst.Y}); err != nil {
		t.Fatalf("villager move rejected: %v", err)
	}
	if a.X != dst.X || a.Y != dst.Y {
		t.Errorf("villager at (%d,%d), want (%d,%d)", a.X, a.Y, dst.X, dst.Y)
	}
	if a.Intent != nil {
		t.Error("move did not cancel the in-flight intent (cancel-and-replan)")
	}
	if a.IdleSince != 100 {
		t.Errorf("IdleSince = %d, want the landing tick 100", a.IdleSince)
	}
	if s.GuardianCharges != 2 {
		t.Errorf("charges = %d, want 2 (one spent)", s.GuardianCharges)
	}
}

func TestMiracleMoveStructureWhole(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 3
	bx, by, ok := findBuildTile(m, s)
	if !ok {
		t.Skip("no build tile")
	}
	dst := Point{X: bx, Y: by}
	// A fire somewhere else, carrying fuel that must ride along whole.
	src := Point{X: dst.X, Y: dst.Y}
	if p, ok2 := passableTileExcept(m, s, dst); ok2 {
		src = p
	}
	s.Structures = append(s.Structures, Structure{Kind: "fire", X: src.X, Y: src.Y, FuelUntil: 99999})

	if err := applyMiracleErr(s, 50, "guardian.entity_moved", EntityMovedPayload{
		Class: "structure", X: src.X, Y: src.Y, ToX: dst.X, ToY: dst.Y}); err != nil {
		t.Fatalf("structure move rejected: %v", err)
	}
	i := s.structureIndexAt(dst.X, dst.Y)
	if i < 0 {
		t.Fatal("structure not at destination")
	}
	if s.Structures[i].FuelUntil != 99999 || s.Structures[i].Kind != "fire" {
		t.Errorf("structure did not move whole: %+v", s.Structures[i])
	}
	if s.structureIndexAt(src.X, src.Y) >= 0 {
		t.Error("structure still at source")
	}
}

func TestMiracleMovePileMerges(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 3
	srcT, ok := passableTileExcept(m, s)
	if !ok {
		t.Skip("no passable tile")
	}
	dstT, ok := passableTileExcept(m, s, srcT)
	if !ok {
		t.Skip("no second passable tile")
	}
	// Source pile: 4 wood; destination already holds 2 wood → merges to 6.
	sp := s.pileFor(srcT.X, srcT.Y)
	sp.addNonFood("wood", 4)
	dp := s.pileFor(dstT.X, dstT.Y)
	dp.addNonFood("wood", 2)

	if err := applyMiracleErr(s, 70, "guardian.entity_moved", EntityMovedPayload{
		Class: "pile", X: srcT.X, Y: srcT.Y, ToX: dstT.X, ToY: dstT.Y}); err != nil {
		t.Fatalf("pile move rejected: %v", err)
	}
	if s.pileAt(srcT.X, srcT.Y) != nil {
		t.Error("source pile still present")
	}
	dest := s.pileAt(dstT.X, dstT.Y)
	if dest == nil || dest.Wood != 6 {
		t.Errorf("merged pile Wood = %v, want 6", dest)
	}
}

func TestMiracleMoveRejectsImpassableDestination(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 3
	water, ok := firstTileOfKind(m, worldmap.Water)
	if !ok {
		t.Skip("no water on this map")
	}
	a := &s.Agents[0]
	before := s.Marshal()
	err := applyMiracleErr(s, 40, "guardian.entity_moved", EntityMovedPayload{
		Class: "villager", X: a.X, Y: a.Y, ToX: water.X, ToY: water.Y})
	if err == nil {
		t.Fatal("move onto water should be rejected")
	}
	if string(s.Marshal()) != string(before) {
		t.Error("rejected move left a partial change / spent a charge")
	}
}

func TestMiracleMoveRejectsAbsentClass(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 3
	// A tile with no villager on it.
	empty, ok := passableTileExcept(m, s, agentPoints(s)...)
	if !ok {
		t.Skip("no empty passable tile")
	}
	dst, _ := passableTileExcept(m, s, empty)
	before := s.Marshal()
	err := applyMiracleErr(s, 40, "guardian.entity_moved", EntityMovedPayload{
		Class: "villager", X: empty.X, Y: empty.Y, ToX: dst.X, ToY: dst.Y})
	if err == nil {
		t.Fatal("moving a villager from an empty tile should be rejected")
	}
	if string(s.Marshal()) != string(before) {
		t.Error("rejected move mutated state")
	}
}

func TestMiracleRemoveVillagerRejected(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 3
	a := &s.Agents[0]
	before := s.Marshal()
	err := applyMiracleErr(s, 40, "guardian.entity_removed", EntityRemovedPayload{
		Class: "villager", X: a.X, Y: a.Y})
	if err == nil {
		t.Fatal("removing a villager must be rejected (v1 doctrine)")
	}
	if string(s.Marshal()) != string(before) {
		t.Error("rejected villager-remove mutated state")
	}
}

func TestMiracleRemoveChestSpillsContents(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 3
	tile, ok := passableTileExcept(m, s)
	if !ok {
		t.Skip("no passable tile")
	}
	store := &Inventory{Wood: 5, FoodRaw: 3, Spears: []int{2}}
	s.Structures = append(s.Structures, Structure{Kind: "chest", X: tile.X, Y: tile.Y, Owner: 0, Store: store})

	if err := applyMiracleErr(s, 200, "guardian.entity_removed", EntityRemovedPayload{
		Class: "structure", X: tile.X, Y: tile.Y}); err != nil {
		t.Fatalf("chest remove rejected: %v", err)
	}
	if s.structureIndexAt(tile.X, tile.Y) >= 0 {
		t.Error("chest not removed")
	}
	pile := s.pileAt(tile.X, tile.Y)
	if pile == nil {
		t.Fatal("chest contents were not spilled to a pile")
	}
	if pile.Wood != 5 {
		t.Errorf("spilled Wood = %d, want 5", pile.Wood)
	}
	if pile.avail("food_raw") != 3 {
		t.Errorf("spilled food_raw = %d, want 3", pile.avail("food_raw"))
	}
	if len(pile.Spears) != 1 || pile.Spears[0] != 2 {
		t.Errorf("spilled Spears = %v, want [2]", pile.Spears)
	}
	// Food spilled to the ground gains a rot deadline (death-spill vocabulary).
	if len(pile.Food) != 1 || pile.Food[0].SpoilAt != 200+rotWindowTicks {
		t.Errorf("spilled food batch = %+v, want spoil_at %d", pile.Food, 200+rotWindowTicks)
	}
}

func TestMiracleRemovePileDestroysContents(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 3
	tile, ok := passableTileExcept(m, s)
	if !ok {
		t.Skip("no passable tile")
	}
	p := s.pileFor(tile.X, tile.Y)
	p.addNonFood("stone", 7)

	if err := applyMiracleErr(s, 90, "guardian.entity_removed", EntityRemovedPayload{
		Class: "pile", X: tile.X, Y: tile.Y}); err != nil {
		t.Fatalf("pile remove rejected: %v", err)
	}
	if s.pileAt(tile.X, tile.Y) != nil {
		t.Error("pile not removed")
	}
	if s.GuardianCharges != 2 {
		t.Errorf("charges = %d, want 2", s.GuardianCharges)
	}
}

func TestMiracleRemoveTerrainRouting(t *testing.T) {
	const seed = 42
	m := testMap(seed)

	cases := []struct {
		kind  worldmap.TileKind
		label string
		check func(s *State, p Point) bool
	}{
		{worldmap.Tree, "tree", func(s *State, p Point) bool { return effectiveKind(m, s, p.X, p.Y) == worldmap.Grass }},
		{worldmap.Forage, "forage", func(s *State, p Point) bool {
			for _, h := range s.Harvested {
				if h.X == p.X && h.Y == p.Y && h.Regrow == 30+forageRegrowSec {
					return true
				}
			}
			return false
		}},
		{worldmap.Rock, "rock", func(s *State, p Point) bool { return effectiveKind(m, s, p.X, p.Y) == worldmap.Depleted }},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			p, ok := firstTileOfKind(m, c.kind)
			if !ok {
				t.Skipf("no %s tile", c.label)
			}
			s := NewState(seed, m)
			s.GuardianCharges = 3
			if err := applyMiracleErr(s, 30, "guardian.entity_removed", EntityRemovedPayload{
				Class: "terrain", X: p.X, Y: p.Y}); err != nil {
				t.Fatalf("%s remove rejected: %v", c.label, err)
			}
			if !c.check(s, p) {
				t.Errorf("%s remove did not route to the right overlay", c.label)
			}
			// Removing an already-overlaid tile is a no-op target → rejected.
			before := s.Marshal()
			if err := applyMiracleErr(s, 31, "guardian.entity_removed", EntityRemovedPayload{
				Class: "terrain", X: p.X, Y: p.Y}); err == nil {
				t.Errorf("already-overlaid %s should be rejected", c.label)
			}
			if string(s.Marshal()) != string(before) {
				t.Errorf("rejected re-remove mutated state")
			}
		})
	}
}

func TestMiracleInsufficientChargeRejected(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 0
	a := &s.Agents[0]
	dst, ok := passableTileExcept(m, s, Point{X: a.X, Y: a.Y})
	if !ok {
		t.Skip("no spare passable tile")
	}
	before := s.Marshal()
	err := applyMiracleErr(s, 40, "guardian.entity_moved", EntityMovedPayload{
		Class: "villager", X: a.X, Y: a.Y, ToX: dst.X, ToY: dst.Y})
	if err == nil {
		t.Fatal("move with an empty bank should be rejected")
	}
	if string(s.Marshal()) != string(before) {
		t.Error("charge-starved reject mutated state")
	}
}

func TestMiracleGratisWaivesChargeOnly(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 0 // empty bank
	a := &s.Agents[0]
	dst, ok := passableTileExcept(m, s, Point{X: a.X, Y: a.Y})
	if !ok {
		t.Skip("no spare passable tile")
	}
	// Gratis lands with a zero bank (charge waived) — but validation still runs.
	if err := applyMiracleErr(s, 40, "guardian.entity_moved", EntityMovedPayload{
		Class: "villager", X: a.X, Y: a.Y, ToX: dst.X, ToY: dst.Y, Gratis: true}); err != nil {
		t.Fatalf("gratis move with empty bank rejected: %v", err)
	}
	if s.GuardianCharges != 0 {
		t.Errorf("gratis spent a charge: bank = %d", s.GuardianCharges)
	}
	// Gratis does NOT waive the destination check.
	water, ok := firstTileOfKind(m, worldmap.Water)
	if ok {
		before := s.Marshal()
		if err := applyMiracleErr(s, 41, "guardian.entity_moved", EntityMovedPayload{
			Class: "villager", X: a.X, Y: a.Y, ToX: water.X, ToY: water.Y, Gratis: true}); err == nil {
			t.Error("gratis move onto water should still be rejected")
		}
		if string(s.Marshal()) != string(before) {
			t.Error("rejected gratis move mutated state")
		}
	}
}

// TestMiracleReplayByteIdentity is SC-002 over US1: a scripted villager-move +
// chest-remove (with a spill) run replays from genesis (log only) to a
// byte-identical state hash — the recorded miracles re-apply cleanly.
func TestMiracleReplayByteIdentity(t *testing.T) {
	const seed = 42
	m := testMap(seed)

	base := NewState(seed, m)
	moveDst, ok := passableTileExcept(m, base, Point{X: base.Agents[0].X, Y: base.Agents[0].Y})
	if !ok {
		t.Skip("no spare passable tile")
	}
	chestTile, ok := passableTileExcept(m, base, Point{X: base.Agents[0].X, Y: base.Agents[0].Y}, moveDst)
	if !ok {
		t.Skip("no chest tile")
	}
	ax, ay := base.Agents[0].X, base.Agents[0].Y

	genesis := func() *State {
		s := NewState(seed, m)
		for i := 1; i < len(s.Agents); i++ {
			s.Agents[i].Dead = true // lone living villager keeps the run quiet
		}
		s.GuardianCharges = 3
		s.Structures = append(s.Structures, Structure{
			Kind: "chest", X: chestTile.X, Y: chestTile.Y, Owner: 0,
			Store: &Inventory{Wood: 5, FoodRaw: 3, Spears: []int{2}}})
		return s
	}

	pl := func(v any) []byte { return mustPayload(v) }
	commands := map[int64][]store.Event{
		10: {{Tick: 10, Type: "guardian.entity_moved", Payload: pl(EntityMovedPayload{
			Class: "villager", X: ax, Y: ay, ToX: moveDst.X, ToY: moveDst.Y})}},
		20: {{Tick: 20, Type: "guardian.entity_removed", Payload: pl(EntityRemovedPayload{
			Class: "structure", X: chestTile.X, Y: chestTile.Y})}},
	}

	const ticks = 60
	live := genesis()
	log := driveTicks(t, live, m, ticks, commands)

	var sawMove, sawRemove bool
	for _, e := range log {
		switch e.Type {
		case "guardian.entity_moved":
			sawMove = true
		case "guardian.entity_removed":
			sawRemove = true
		}
	}
	if !sawMove || !sawRemove {
		t.Fatalf("scripted miracles missing from the log (move %v, remove %v)", sawMove, sawRemove)
	}
	if s := live.pileAt(chestTile.X, chestTile.Y); s == nil || s.Wood != 5 {
		t.Fatalf("chest spill missing after the run: %+v", s)
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
		t.Fatalf("replay diverged:\nlive:     %s\nreplayed: %s", string(live.Marshal()), string(replay.Marshal()))
	}
}

// TestMoveFreshnessReplayByteIdentical (spec 091 FR-002/SC-002): the guardian
// door's move-miracle name re-resolution (TASK-166) is an EMITTER-side change —
// it decides which coordinates a recorded guardian.entity_moved carries, never
// how the reducer applies one. A recorded log of coordinate-only entity_moved
// events — the only shape any pre-fix (or post-fix, coordinate-addressed)
// recording ever carried, since the emitter always bakes concrete x/y into the
// event regardless of how it resolved them — must replay to a byte-identical
// state hash on the fixed binary. This pins applyEntityMoved
// (miracles.go:497, the "no living villager at (x,y)" arm) as untouched by
// TASK-166, mirroring TestMiracleReplayByteIdentity's genesis/log/replay
// skeleton with a two-hop villager move standing in for a short pre-fix
// history.
func TestMoveFreshnessReplayByteIdentical(t *testing.T) {
	const seed = 42
	m := testMap(seed)

	base := NewState(seed, m)
	hop1, ok := passableTileExcept(m, base, Point{X: base.Agents[0].X, Y: base.Agents[0].Y})
	if !ok {
		t.Skip("no spare passable tile")
	}
	hop2, ok := passableTileExcept(m, base, Point{X: base.Agents[0].X, Y: base.Agents[0].Y}, hop1)
	if !ok {
		t.Skip("no second spare passable tile")
	}
	ax, ay := base.Agents[0].X, base.Agents[0].Y

	genesis := func() *State {
		s := NewState(seed, m)
		for i := 1; i < len(s.Agents); i++ {
			s.Agents[i].Dead = true // lone living villager keeps the run quiet
		}
		s.GuardianCharges = 3
		return s
	}

	pl := func(v any) []byte { return mustPayload(v) }
	commands := map[int64][]store.Event{
		// Two moves in sequence — a coordinate-only history exactly like any
		// log recorded before TASK-166 shipped (the emitter always recorded
		// concrete coordinates; only the door's CHOICE of source coordinates
		// changed, not the event shape the reducer consumes).
		10: {{Tick: 10, Type: "guardian.entity_moved", Payload: pl(EntityMovedPayload{
			Class: "villager", X: ax, Y: ay, ToX: hop1.X, ToY: hop1.Y})}},
		20: {{Tick: 20, Type: "guardian.entity_moved", Payload: pl(EntityMovedPayload{
			Class: "villager", X: hop1.X, Y: hop1.Y, ToX: hop2.X, ToY: hop2.Y})}},
	}

	const ticks = 60
	live := genesis()
	log := driveTicks(t, live, m, ticks, commands)

	var moves int
	for _, e := range log {
		if e.Type == "guardian.entity_moved" {
			moves++
		}
	}
	if moves != 2 {
		t.Fatalf("scripted moves missing from the log: got %d, want 2", moves)
	}
	if live.Agents[0].X != hop2.X || live.Agents[0].Y != hop2.Y {
		t.Fatalf("villager did not land at the second hop: at (%d,%d), want (%d,%d)",
			live.Agents[0].X, live.Agents[0].Y, hop2.X, hop2.Y)
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
		t.Fatalf("replay diverged:\nlive:     %s\nreplayed: %s", string(live.Marshal()), string(replay.Marshal()))
	}
}

// agentPoints is the set of living villager tiles (for "empty tile" searches).
func agentPoints(s *State) []Point {
	var pts []Point
	for i := range s.Agents {
		if !s.Agents[i].Dead {
			pts = append(pts, Point{X: s.Agents[i].X, Y: s.Agents[i].Y})
		}
	}
	return pts
}

// --- US2 gratis (spec 016 T014) ---

// TestGratisValidationSurvives is US2-AS2 / T014: a forced (gratis) miracle is
// rejected on invalid input exactly as a charged one — gratis waives the charge
// only, never a validity rule. Paired charged/forced attempts on the same
// invalid inputs must leave the state byte-identical (no partial application).
func TestGratisValidationSurvives(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	water, ok := firstTileOfKind(m, worldmap.Water)
	if !ok {
		t.Skip("no water on this map")
	}
	empty, ok := passableTileExcept(m, NewState(seed, m), agentPoints(NewState(seed, m))...)
	if !ok {
		t.Skip("no empty passable tile")
	}

	invalids := []struct {
		label string
		typ   string
		mk    func(gratis bool) any
	}{
		{"move onto water", "guardian.entity_moved", func(g bool) any {
			return EntityMovedPayload{Class: "villager", Gratis: g} // X,Y filled per-state below
		}},
		{"remove absent structure", "guardian.entity_removed", func(g bool) any {
			return EntityRemovedPayload{Class: "structure", X: empty.X, Y: empty.Y, Gratis: g}
		}},
	}
	for _, c := range invalids {
		t.Run(c.label, func(t *testing.T) {
			// Charged and forced runs on identical fresh states. Each returns the
			// error and the after-state bytes; the before-state bytes are checked
			// in-run (no partial application).
			run := func(gratis bool) (after string, err error) {
				s := NewState(seed, m)
				s.GuardianCharges = 3
				a := &s.Agents[0]
				pl := c.mk(gratis)
				if mv, isMove := pl.(EntityMovedPayload); isMove {
					mv.X, mv.Y, mv.ToX, mv.ToY = a.X, a.Y, water.X, water.Y
					pl = mv
				}
				before := string(s.Marshal())
				err = applyMiracleErr(s, 40, c.typ, pl)
				after = string(s.Marshal())
				if after != before {
					t.Errorf("gratis=%v: rejected miracle mutated state (partial application)", gratis)
				}
				return after, err
			}
			charged, cerr := run(false)
			forced, ferr := run(true)
			if cerr == nil || ferr == nil {
				t.Fatalf("both charged and forced must reject: charged=%v forced=%v", cerr, ferr)
			}
			// Gratis is not a validity waiver: identical rejection, identical state.
			if charged != forced {
				t.Errorf("charged vs forced left different state:\n charged: %s\n forced:  %s", charged, forced)
			}
		})
	}
}

// TestGratisIsLoggedVisible is US2-AS4 / SC-004 / T014: a landed forced miracle's
// recorded payload carries "gratis":true, enumerable after the fact from the
// event log — the reviewer can find every gratis act.
func TestGratisIsLoggedVisible(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	base := NewState(seed, m)
	dst, ok := passableTileExcept(m, base, Point{X: base.Agents[0].X, Y: base.Agents[0].Y})
	if !ok {
		t.Skip("no spare passable tile")
	}
	ax, ay := base.Agents[0].X, base.Agents[0].Y

	genesis := func() *State {
		s := NewState(seed, m)
		for i := 1; i < len(s.Agents); i++ {
			s.Agents[i].Dead = true
		}
		s.GuardianCharges = 0 // empty bank: only a gratis move can land
		return s
	}
	commands := map[int64][]store.Event{
		10: {{Tick: 10, Type: "guardian.entity_moved", Payload: mustPayload(EntityMovedPayload{
			Class: "villager", X: ax, Y: ay, ToX: dst.X, ToY: dst.Y, Gratis: true})}},
	}
	live := genesis()
	log := driveTicks(t, live, m, 30, commands)

	var found bool
	for _, e := range log {
		if e.Type != "guardian.entity_moved" {
			continue
		}
		found = true
		var p EntityMovedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			t.Fatal(err)
		}
		if !p.Gratis {
			t.Error("landed forced move payload does not carry gratis=true")
		}
		// The gratis flag is enumerable straight from the recorded JSON.
		if !jsonHasGratisTrue(e.Payload) {
			t.Errorf("recorded payload not enumerable as gratis: %s", e.Payload)
		}
	}
	if !found {
		t.Fatal("forced move missing from the log")
	}
	if live.GuardianCharges != 0 {
		t.Errorf("gratis move spent a charge from an empty bank: %d", live.GuardianCharges)
	}
}

func jsonHasGratisTrue(b []byte) bool {
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return false
	}
	g, ok := raw["gratis"].(bool)
	return ok && g
}

// --- US3 time snap (spec 016 T015-T019) ---

// TestSnapForwardOnly is FR-008 / US3-AS4: a target at or before the current
// tick is rejected whole, no charge spent, state unchanged.
func TestSnapForwardOnly(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.Tick = 5000
	s.GuardianCharges = 3
	for _, to := range []int64{5000, 4999, 0} {
		before := s.Marshal()
		if err := applyMiracleErr(s, 5000, "guardian.time_snapped", TimeSnappedPayload{ToTick: to}); err == nil {
			t.Errorf("snap to %d (<= current 5000) should be rejected", to)
		}
		if string(s.Marshal()) != string(before) {
			t.Errorf("rejected snap to %d mutated state / spent a charge", to)
		}
	}
}

// TestSnapPreservesRemainingDurations is US3-AS2 / SC-003 (arbitrary-delta
// variant) / T018(b): after a snap, every SHIFT field advanced by exactly delta
// (so its remaining duration is preserved), every zero=never sentinel stayed
// zero, and every KEEP (history/identity) field is untouched. This is the
// per-field, non-circular validation of the rebaseTicks taxonomy.
func TestSnapPreservesRemainingDurations(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	const old = int64(50000)
	const delta = int64(12345)
	const to = old + delta
	s.Tick = old
	s.GuardianCharges = 3

	a := &s.Agents[0]
	a.IdleSince = 40000
	a.LastTalk = 41000
	a.LastGive = 42000
	a.LastGoalTick = 43000
	a.Generation = 7
	a.LastConsolidatedNight = 3
	a.ConsolidatedUpTo = 44000
	a.LastConsolidateMark = 45000
	a.Intent = &Intent{Goal: "forage", WorkStart: 46000}
	a.Hail = &AgentHail{By: 1, Until: 47000}
	a.Memories = []Memory{{Text: "x", Salience: 5, Tick: 100, Subject: -1}}
	a.Beliefs = []Belief{{ID: 1, Tick: 200}}
	a.Known = []KnownRumor{{RumorID: 1, Text: "r", Tick: 300, From: -1}}
	a.Plan = []PlanStep{{Job: "j", Goal: "forage", Until: 48000,
		When: &Guard{Type: GuardAfterTick, Tick: 49000, Generation: 7}}}
	a.NeedsAnchor = &Needs{Warmth: 500} // levels: frozen, never shifted
	a.NeedsAnchorTick = 43500
	// Spec 083: two anchors set + one zero=sentinel per kind, plus a latch.
	a.Neglect = &NeglectState{WarmthSince: 43600, WarmthIntent: 43700, WarmthFired: true}

	// Sentinels: agent 1's not-started work and never-talked cooldown stay zero.
	a2 := &s.Agents[1]
	a2.Intent = &Intent{Goal: "wander", WorkStart: 0}
	a2.LastTalk = 0

	s.Structures = []Structure{
		{Kind: "fire", X: 1, Y: 1, FuelUntil: 51000},
		{Kind: "shelter", X: 2, Y: 2, FuelUntil: 0}, // cold/non-fire: stays zero
	}
	s.Harvested = []Harvest{{X: 3, Y: 3, Regrow: 52000}}
	s.DenUses = []DenUse{{X: 4, Y: 4, Ready: 53000}}
	s.Piles = []Pile{{X: 5, Y: 5, Food: []FoodBatch{{Kind: "food_raw", N: 2, SpoilAt: 54000}}}}
	s.Debts = []Debt{{ID: 1, Debtor: 0, Creditor: 1, Kind: "food", Due: 55000, Status: "open"}}
	s.Rumors = []Rumor{{ID: 1, Subject: 2, OriginAgent: 0, OriginTick: 400}}
	s.Gru = &Gru{X: 6, Y: 6, LastAttack: 56000}
	s.Conversations = []ConvoRecord{{Conv: 500, Tick: 600, Participants: []int{0, 1}}}
	s.PairTalks = []PairTalk{{A: 0, B: 1, Tick: 59000}} // spec 061: pair cooldown anchor, shifts
	s.Chronicle = []ChronicleEntry{{Tick: 700, Day: 1, FromTick: 650, ToTick: 700}}
	s.Meeting = MeetingState{Phase: "open", OpenedTick: 57000, GatherStart: 58000, LastMeetingDay: 2}
	s.MeetingConvention = &MeetingConvention{ConveneSecond: 100, OpenSecond: 200, EstablishedDay: 1}
	s.Norms = []Norm{{ID: 1, Kind: "k", DayPassed: 1, DayRepealed: 2, DayAmended: 3,
		Violations: []NormViolation{{Agent: 0, Tick: 800}}}}

	if err := applyMiracleErr(s, old, "guardian.time_snapped", TimeSnappedPayload{ToTick: to}); err != nil {
		t.Fatalf("snap rejected: %v", err)
	}
	if s.Tick != to {
		t.Fatalf("Tick = %d, want %d", s.Tick, to)
	}
	if s.GuardianCharges != 1 {
		t.Errorf("charges = %d, want 1 (2 spent on the snap)", s.GuardianCharges)
	}

	eq := func(label string, got, want int64) {
		t.Helper()
		if got != want {
			t.Errorf("%s = %d, want %d", label, got, want)
		}
	}
	// SHIFT (+delta): remaining duration preserved.
	eq("Agent.IdleSince", a.IdleSince, 40000+delta)
	eq("Agent.LastTalk", a.LastTalk, 41000+delta)
	eq("Agent.LastGive", a.LastGive, 42000+delta)
	eq("Intent.WorkStart", a.Intent.WorkStart, 46000+delta)
	eq("AgentHail.Until", a.Hail.Until, 47000+delta)
	eq("PlanStep.Until", a.Plan[0].Until, 48000+delta)
	eq("Guard.Tick", a.Plan[0].When.Tick, 49000+delta)
	eq("Structure.FuelUntil", s.Structures[0].FuelUntil, 51000+delta)
	eq("Harvest.Regrow", s.Harvested[0].Regrow, 52000+delta)
	eq("DenUse.Ready", s.DenUses[0].Ready, 53000+delta)
	eq("FoodBatch.SpoilAt", s.Piles[0].Food[0].SpoilAt, 54000+delta)
	eq("Debt.Due", s.Debts[0].Due, 55000+delta)
	eq("Gru.LastAttack", s.Gru.LastAttack, 56000+delta)
	eq("Meeting.OpenedTick", s.Meeting.OpenedTick, 57000+delta)
	eq("Meeting.GatherStart", s.Meeting.GatherStart, 58000+delta)
	eq("Agent.NeedsAnchorTick", a.NeedsAnchorTick, 43500+delta)
	eq("PairTalk.Tick", s.PairTalks[0].Tick, 59000+delta)
	// Spec 083: the in-band episode and its zero-intent clock survive the jump;
	// the not-in-band/never sentinels stay zero; the episode latch is untouched.
	eq("NeglectState.WarmthSince", a.Neglect.WarmthSince, 43600+delta)
	eq("NeglectState.WarmthIntent", a.Neglect.WarmthIntent, 43700+delta)
	eq("NeglectState.FoodSince(0)", a.Neglect.FoodSince, 0)
	eq("NeglectState.RestIntent(0)", a.Neglect.RestIntent, 0)
	if !a.Neglect.WarmthFired {
		t.Error("NeglectState.WarmthFired latch must survive a snap untouched")
	}
	// The anchor LEVELS ride the freeze untouched (need values, not ticks).
	if a.NeedsAnchor == nil || a.NeedsAnchor.Warmth != 500 {
		t.Errorf("NeedsAnchor levels changed across snap: %+v", a.NeedsAnchor)
	}
	// IdleSince shifts unconditionally: agent 1's genesis-zero becomes delta
	// (elapsed-idle is preserved, not a "never" sentinel).
	eq("Agent[1].IdleSince", s.Agents[1].IdleSince, delta)

	// Zero=never sentinels stay zero.
	eq("Agent[1].Intent.WorkStart(0)", a2.Intent.WorkStart, 0)
	eq("Agent[1].LastTalk(0)", a2.LastTalk, 0)
	eq("Structure.FuelUntil(0)", s.Structures[1].FuelUntil, 0)

	// KEEP: history/identity untouched.
	eq("Agent.Generation", a.Generation, 7)
	eq("Agent.LastGoalTick", a.LastGoalTick, 43000)
	eq("Agent.LastConsolidatedNight", a.LastConsolidatedNight, 3)
	eq("Agent.ConsolidatedUpTo", a.ConsolidatedUpTo, 44000)
	eq("Agent.LastConsolidateMark", a.LastConsolidateMark, 45000)
	eq("Memory.Tick", a.Memories[0].Tick, 100)
	eq("Belief.Tick", a.Beliefs[0].Tick, 200)
	eq("KnownRumor.Tick", a.Known[0].Tick, 300)
	eq("Guard.Generation", a.Plan[0].When.Generation, 7)
	eq("Rumor.OriginTick", s.Rumors[0].OriginTick, 400)
	eq("ConvoRecord.Conv", s.Conversations[0].Conv, 500)
	eq("ConvoRecord.Tick", s.Conversations[0].Tick, 600)
	eq("ChronicleEntry.Tick", s.Chronicle[0].Tick, 700)
	eq("ChronicleEntry.Day", s.Chronicle[0].Day, 1)
	eq("ChronicleEntry.FromTick", s.Chronicle[0].FromTick, 650)
	eq("ChronicleEntry.ToTick", s.Chronicle[0].ToTick, 700)
	eq("Meeting.LastMeetingDay", s.Meeting.LastMeetingDay, 2)
	eq("MeetingConvention.EstablishedDay", s.MeetingConvention.EstablishedDay, 1)
	eq("Norm.DayPassed", s.Norms[0].DayPassed, 1)
	eq("Norm.DayRepealed", s.Norms[0].DayRepealed, 2)
	eq("Norm.DayAmended", s.Norms[0].DayAmended, 3)
	eq("NormViolation.Tick", s.Norms[0].Violations[0].Tick, 800)
}

// TestRebaseTaxonomyComplete is the drift-hazard tripwire (research R3, T017):
// a reflective walk of the marshalled State tree collects every tick-anchored
// int64 field; each MUST have a SHIFT/KEEP classification entry here (matching
// the rebaseTicks doctrine). A future field (e.g. a new `NewTimer int64` on
// Agent) with no entry fails the build, forcing a deliberate classification
// before it can silently drift across a snap.
func TestRebaseTaxonomyComplete(t *testing.T) {
	const shift, keep = "shift", "keep"
	classified := map[string]string{
		"State.Tick": keep, // the clock anchor: set by applyTimeSnapped, never rebased
		// SHIFT — future deadlines / duration anchors.
		"Agent.LastTalk":            shift,
		"Agent.LastGive":            shift,
		"PairTalk.Tick":             shift, // spec 061: pair last-exchange cooldown anchor (Agent.LastTalk shape)
		"Agent.IdleSince":           shift,
		"Agent.LastMindIntentDone":  shift, // spec 062 US1: yield-window anchor (Belief.Reinforced shape), 0 = never mind-driven
		"Intent.WorkStart":          shift,
		"AgentHail.Until":           shift,
		"PlanStep.Until":            shift, // deviation from data-model.md — see rebaseTicks NOTE
		"Guard.Tick":                shift, // deviation from data-model.md — see rebaseTicks NOTE
		"Structure.FuelUntil":       shift,
		"Harvest.Regrow":            shift,
		"DenUse.Ready":              shift,
		"FoodBatch.SpoilAt":         shift,
		"Debt.Due":                  shift,
		"Belief.Reinforced":         shift, // spec 030: decay anchor, non-zero shifts; 0 = grandfather stays 0
		"Gru.LastAttack":            shift,
		"MeetingState.OpenedTick":   shift,
		"MeetingState.GatherStart":  shift,
		"GuardianOrder.ExpiresTick": shift, // spec 029: a standing order's future expiry deadline
		"Directive.ExpiresTick":     shift, // spec 084: a directive's TTL deadline — the GuardianOrder.ExpiresTick classification verbatim (ACTIVE only)
		"PlaceFact.Seen":            shift, // spec 041: mental-map freshness anchor (Belief.Reinforced shape)
		"PeerSighting.Seen":         shift, // spec 041 T013: sighting recency anchor, same shape
		"Agent.NeedsAnchorTick":     shift, // spec 043 US2: trajectory-window edge anchor (Belief.Reinforced shape), 0 = unset
		"State.ColdSnapUntil":       shift, // spec 077: cold-snap expiry deadline, read live (Structure.FuelUntil shape)
		"Stranger.LastMove":         shift, // spec 077: movement-cadence anchor (Gru.LastAttack shape)
		"Stranger.LastTake":         shift, // spec 077: take-cooldown anchor (Gru.LastAttack shape)
		"NeglectState.FoodSince":    shift, // spec 083: band-entry anchor (Belief.Reinforced shape), 0 = not in band
		"NeglectState.WarmthSince":  shift, // spec 083: band-entry anchor, 0 = not in band
		"NeglectState.RestSince":    shift, // spec 083: band-entry anchor, 0 = not in band
		"NeglectState.FoodIntent":   shift, // spec 083: last-class-intent stamp (elapsed anchor), 0 = never
		"NeglectState.WarmthIntent": shift, // spec 083: last-class-intent stamp, 0 = never
		"NeglectState.RestIntent":   shift, // spec 083: last-class-intent stamp, 0 = never
		"Prophecy.DeadlineTick":     shift, // spec 085: a prophecy's future judgment deadline — the Directive.ExpiresTick classification verbatim (ACTIVE only)
		"ObservationMark.Tick":      shift, // spec 097: observation dedup anchor (elapsed gates the window — Belief.Reinforced shape), never zero once set
		// KEEP — history / identity / counters.
		"Agent.Generation":                 keep,
		"Agent.LastConsolidatedNight":      keep,
		"Agent.ConsolidatedUpTo":           keep,
		"Agent.LastConsolidateMark":        keep,
		"Agent.LastGoalTick":               keep,
		"IntentRecord.Tick":                keep, // spec 043: when the intent landed (history), like Memory.Tick
		"IntentRecord.OutcomeTick":         keep, // spec 043: when the outcome landed (history), like Memory.Tick
		"Memory.Tick":                      keep,
		"Memory.Conv":                      keep, // spec 019: conversation-ref identity (founding-talk tick), like ConvoRecord.Conv
		"Memory.Seq":                       keep, // spec 042: the emitting event's store seq — an identity, never a clock value
		"Agent.SitVecTick":                 keep, // spec 042: when the situation text was rendered (history/audit), like Memory.Tick
		"JournalEntry.Tick":                keep, // spec 019: when the entry was written (history), like Memory.Tick
		"Belief.Tick":                      keep,
		"KnownRumor.Tick":                  keep,
		"Guard.Generation":                 keep,
		"Rumor.OriginTick":                 keep,
		"ConvoRecord.Conv":                 keep,
		"ConvoRecord.Tick":                 keep,
		"ChronicleEntry.Tick":              keep,
		"ChronicleEntry.Day":               keep,
		"ChronicleEntry.FromTick":          keep,
		"ChronicleEntry.ToTick":            keep,
		"MeetingState.LastMeetingDay":      keep,
		"MeetingConvention.EstablishedDay": keep,
		"Norm.DayPassed":                   keep,
		"Norm.DayRepealed":                 keep,
		"Norm.DayAmended":                  keep,
		"NormViolation.Tick":               keep,
		"GuardianOrder.PlacedTick":         keep, // spec 029: when the order was placed (history)
		"GuardianOrder.PlacedSeq":          keep, // spec 054: the placement event's store seq — an identity, like Memory.Seq
		"Designation.PlacedTick":           keep, // spec 084: when the designation was placed (history) — no future deadline exists on a designation
		"Designation.PlacedSeq":            keep, // spec 084: the placement event's store seq — an identity (the GuardianOrder.PlacedSeq shape)
		"Directive.IssuedTick":             keep, // spec 084: when the directive was issued (history)
		"Directive.PlacedSeq":              keep, // spec 084: the issue event's store seq — an identity
		"Region.PlacedTick":                keep, // spec 101: when the region was christened (history) — no future deadline exists on a region (no terminal event in v1)
		"Region.PlacedSeq":                 keep, // spec 101: the christening event's store seq — an identity (the Designation.PlacedSeq shape)
		"Prophecy.DeclaredTick":            keep, // spec 085: when the word was given (history, Directive.IssuedTick shape)
		"Prophecy.PlacedSeq":               keep, // spec 085: the declaration event's store seq — an identity
		"PlaceFact.Detail":                 keep, // spec 041: remembered value baked at emission, never re-derived (see rebaseTicks)
		"RunEnd.Tick":                      keep, // spec 044: when the run ended (history; the world never ticks again)
		"DeathRecord.Tick":                 keep, // spec 044: when the death happened (history, like NormViolation.Tick)
		"MorgueEpilogue.Tick":              keep, // spec 044 US2: when the epilogue landed (history, ChronicleEntry.Tick shape)
		"CurriculumPass.Tick":              keep, // spec 046: when the pass was recorded (history), like Memory.Tick
		"Stranger.Night":                   keep, // spec 077: 1-based arrival night — identity, like Rumor.OriginTick
		"StrangerTake.Tick":                keep, // spec 077: when the take happened — ledger history (DeathRecord.Tick shape)
		"State.CharterObservedSeq":         keep, // spec 077: log coordinate of the recorded observation — identity, like Memory.Seq
		"State.CharterObservedTick":        keep, // spec 077: log coordinate — evidence pointer, like EvidenceRef.Tick
		"State.SkillsObservedSeq":          keep, // spec 077: log coordinate — identity, like Memory.Seq
		"State.SkillsObservedTick":         keep, // spec 077: log coordinate — evidence pointer, like EvidenceRef.Tick
		"EvidenceRef.Tick":                 keep, // spec 046: audit pointer at a recorded event's tick — history, never a deadline
		"EvidenceRef.Seq":                  keep, // spec 046: the evidence event's store seq — an identity, like Memory.Seq
		"GuardianReportCard.Tick":          keep, // spec 063: when the card landed (history, MorgueEpilogue.Tick shape)
		"GuardianReportCard.Seq":           keep, // spec 063: the card event's own seq — an identity, like Memory.Seq
		"GuardianReportCard.Citations":     keep, // spec 063: cited event seqs — identities into recorded history, like EvidenceRef.Seq
		// spec 048 tuning dials are DURATIONS (game-second spans), not absolute
		// tick timestamps — a timeline rebase never shifts them (rebaseTicks
		// leaves s.Tuning untouched). GruEmergePerMille is a uint64 probability,
		// not int64, so the taxonomy walk never flags it.
		"TuningState.RefuelDyingBelow":       keep,
		"TuningState.FireBurnPerWood":        keep,
		"TuningState.PlannerCadenceTicks":    keep,
		"TuningState.EncounterCooldownTicks": keep,
		// Spec 097 dials: durations/points, not clock anchors — KEEP like the
		// other TuningState fields.
		"TuningState.ObservationDedupTicks":         keep,
		"TuningState.ObservationBaseSalience":       keep,
		"TuningState.BeliefDisconfirmRetainPercent": keep,
		"TuningState.BeliefConfirmBoost":            keep,
		// spec 098 dream dials are per-mille ratios and a per-night count —
		// no tick anchors anywhere in the block (the TuningState duration
		// rationale above, one step further from the timeline).
		// Spec 104: the needs-checkpoint cadence is a duration (game-minute
		// span) doubling as the coalescing-regime marker — KEEP like every
		// other dial.
		"TuningState.NeedsCheckpointMinutes": keep,
		// Spec 104 advancement watermarks: decayed-through / beats-processed
		// anchors measured against the live clock — SHIFT, zero = pre-regime
		// sentinel stays zero.
		"Agent.NeedsSyncTick": shift,
		"Gru.Done":            shift,
		// Spec 104 in-flight walk segments are CLEARED by the rebase arm —
		// their beat schedule ((tick+Phase)%MoveEvery) is absolute-tick
		// arithmetic a delta would re-phase, so the snap truncates the walk
		// and the villager re-plans (research.md §7). MoveEvery/Phase are
		// cadence numbers, Done dies with the segment: all KEEP-class for
		// the walk's remaining lifetime of zero.
		"PathSegment.MoveEvery":             keep,
		"PathSegment.Phase":                 keep,
		"PathSegment.Done":                  keep,
		"DreamTuning.DensityPerMille":       keep,
		"DreamTuning.AmbiguousBandPerMille": keep,
		"DreamTuning.HabituationPerMille":   keep,
		"DreamTuning.MergeCapPerNight":      keep,
		"DreamTuning.JitterPerMille":        keep,
	}

	found := map[string]bool{}
	seen := map[reflect.Type]bool{}
	var walk func(rt reflect.Type)
	walk = func(rt reflect.Type) {
		for rt.Kind() == reflect.Ptr || rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
			rt = rt.Elem()
		}
		if rt.Kind() != reflect.Struct || seen[rt] {
			return
		}
		seen[rt] = true
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			if f.PkgPath != "" {
				continue // unexported (e.g. State.m) never serializes
			}
			ft := f.Type
			for ft.Kind() == reflect.Ptr || ft.Kind() == reflect.Slice || ft.Kind() == reflect.Array {
				ft = ft.Elem()
			}
			switch ft.Kind() {
			case reflect.Int64:
				found[rt.Name()+"."+f.Name] = true
			case reflect.Struct:
				walk(f.Type)
			}
		}
	}
	walk(reflect.TypeOf(State{}))

	for path := range found {
		if _, ok := classified[path]; !ok {
			t.Errorf("unclassified tick-anchored int64 field %q — classify it in rebaseTicks (SHIFT or KEEP) and add it to this table", path)
		}
	}
	for path := range classified {
		if !found[path] {
			t.Errorf("stale taxonomy entry %q — no such int64 field in the state tree anymore", path)
		}
	}
}

// TestSnapWholeDayNoDrift is SC-003 / US3-AS1 / T018(a): a whole-day (86400-tick,
// phase-preserving) snap leaves the world's subsequent behavior identical to an
// un-snapped control, modulo the clock offset. Comparison is the event stream
// normalized to each world's own clock base (Type + (tick-base) + payload).
//
// The drive window is deliberately RNG-free: rngAt seeds a PCG from the ABSOLUTE
// tick (rng.go), so agent reflex/wander and gru emergence are NOT phase-invariant
// across a whole-day offset. Only deterministic timer behavior is offset-invariant
// — a fire burning out, ground food rotting, forage regrowing (all pure tick
// comparisons), and an agent frozen mid-work (WorkStart, never completing in the
// window). If any of those SHIFT fields were misclassified, the corresponding
// event would fire at a different normalized tick (or the mid-work agent would
// complete instantly), diverging the streams. The remaining SHIFT fields are
// proven per-field by TestSnapPreservesRemainingDurations.
func TestSnapWholeDayNoDrift(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	const t0 = int64(1000)
	const delta = int64(86400) // exactly one game day: day/night phase preserved
	const window = int64(200)

	bx, by, ok := findBuildTile(m, NewState(seed, m))
	if !ok {
		t.Skip("no build tile")
	}
	fire := Point{X: (bx + 30) % m.W, Y: by}
	pile := Point{X: bx, Y: (by + 30) % m.H}
	harv := Point{X: (bx + 30) % m.W, Y: (by + 30) % m.H}

	genesis := func() *State {
		s := NewState(seed, m)
		s.Tick = t0
		s.GuardianCharges = 3
		for i := range s.Agents {
			s.Agents[i].Dead = true
		}
		// One living agent frozen mid-shelter-build (duration 1200 » window), so
		// it never completes and never idles (no reflex RNG) — but WorkStart must
		// shift, or the snapped copy would complete the build instantly.
		a := &s.Agents[0]
		a.Dead = false
		a.X, a.Y = bx, by
		a.Needs = Needs{Health: 1000, Food: 900, Rest: 900, Warmth: 900, Morale: 800}
		a.IdleSince = t0
		a.Intent = &Intent{Goal: "build_shelter", TargetX: bx, TargetY: by, WorkStart: t0}
		// Deterministic world timers with deadlines inside the window.
		s.Structures = []Structure{{Kind: "fire", X: fire.X, Y: fire.Y, FuelUntil: t0 + 50}}
		s.Harvested = []Harvest{{X: harv.X, Y: harv.Y, Regrow: t0 + 70}}
		s.Piles = []Pile{{X: pile.X, Y: pile.Y, Food: []FoodBatch{{Kind: "food_raw", N: 2, SpoilAt: t0 + 140}}}}
		// Spec 041: pre-settle the living agent's mental map (one applied
		// perception sweep at t0, identical in both copies) — agent.saw bakes
		// absolute Seen ticks, so an unsettled map would emit inside the
		// window and the payloads would differ by exactly the offset this
		// test normalizes away. PlaceFact.Seen is SHIFT (rebaseTicks), and
		// the durable horizon far exceeds the window, so the settled map
		// stays silent in both runs.
		for _, e := range perceptionEvents(s, m, t0) {
			if err := s.Apply(e); err != nil {
				t.Fatalf("pre-settle sweep: %v", err)
			}
		}
		return s
	}

	control := genesis()
	snapped := genesis()
	if err := snapped.Apply(store.Event{Tick: t0, Type: "guardian.time_snapped",
		Payload: mustPayload(TimeSnappedPayload{ToTick: t0 + delta, Gratis: true})}); err != nil {
		t.Fatalf("snap rejected: %v", err)
	}
	if snapped.Tick != t0+delta {
		t.Fatalf("snapped tick = %d, want %d", snapped.Tick, t0+delta)
	}

	ctrlLog := driveTicks(t, control, m, t0+window, nil)
	snapLog := driveTicks(t, snapped, m, t0+delta+window, nil)

	normalize := func(log []store.Event, base int64) []string {
		out := make([]string, len(log))
		for i, e := range log {
			out[i] = fmt.Sprintf("%s@%d %s", e.Type, e.Tick-base, string(e.Payload))
		}
		return out
	}
	cn := normalize(ctrlLog, t0)
	sn := normalize(snapLog, t0+delta)
	if len(cn) == 0 {
		t.Fatal("the drive produced no events — the timers never fired")
	}
	if !reflect.DeepEqual(cn, sn) {
		t.Fatalf("whole-day snap drifted:\n control (%d): %v\n snapped (%d): %v", len(cn), cn, len(sn), sn)
	}
	// The mid-work agent must still be building in both — WorkStart shifted, so
	// neither world completed the 1200-tick build inside the 200-tick window.
	if control.Agents[0].Intent == nil || snapped.Agents[0].Intent == nil {
		t.Fatal("the mid-work build should still be in flight in both worlds")
	}
}

// TestSnapMintsNoCharges is FR-010 / US3-AS3 / T018(c): a snap across two or more
// charge-regeneration boundaries mints nothing — the skipped boundaries never
// fire. A charged snap costs its 2 and no more; a gratis snap changes the bank
// not at all.
func TestSnapMintsNoCharges(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	// From tick 5000, crossing the 21600 and 43200 boundaries (>= 2).
	across := int64(2 * chargeRegenTicks) // 43200
	to := int64(5000) + across + 100

	t.Run("charged pays only its price", func(t *testing.T) {
		s := NewState(seed, m)
		s.Tick = 5000
		s.GuardianCharges = 3
		if err := applyMiracleErr(s, 5000, "guardian.time_snapped", TimeSnappedPayload{ToTick: to}); err != nil {
			t.Fatalf("snap rejected: %v", err)
		}
		if s.GuardianCharges != 1 {
			t.Errorf("charges = %d, want 1 (only the 2-charge cost; skipped boundaries mint nothing)", s.GuardianCharges)
		}
	})
	t.Run("gratis leaves the bank untouched", func(t *testing.T) {
		s := NewState(seed, m)
		s.Tick = 5000
		s.GuardianCharges = 1
		if err := applyMiracleErr(s, 5000, "guardian.time_snapped", TimeSnappedPayload{ToTick: to, Gratis: true}); err != nil {
			t.Fatalf("gratis snap rejected: %v", err)
		}
		if s.GuardianCharges != 1 {
			t.Errorf("charges = %d, want 1 (gratis waives cost; skipped boundaries mint nothing)", s.GuardianCharges)
		}
	})
}

// TestSnapWhilePaused is the US3 edge case: a snap on a paused world re-labels
// the clock and leaves the world paused (the snap touches neither Paused nor the
// speed).
func TestSnapWhilePaused(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.Tick = 1000
	s.Paused = true
	s.GuardianCharges = 3
	if err := applyMiracleErr(s, 1000, "guardian.time_snapped", TimeSnappedPayload{ToTick: 5000}); err != nil {
		t.Fatalf("paused snap rejected: %v", err)
	}
	if s.Tick != 5000 {
		t.Errorf("Tick = %d, want 5000", s.Tick)
	}
	if !s.Paused {
		t.Error("snap must leave a paused world paused")
	}
	if s.GuardianCharges != 1 {
		t.Errorf("charges = %d, want 1", s.GuardianCharges)
	}
}

// TestMiracleSnapReplayByteIdentity is SC-002 over US3 / T019: a scripted snap in
// a driven log replays from genesis to a byte-identical state hash. The replay
// loop sets Tick BEFORE applying each event so the snap re-bases from the same
// tick the live loop snapped at (delta = to_tick - tick must match).
func TestMiracleSnapReplayByteIdentity(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	const snapAt = int64(50)
	const snapTo = int64(120) // delta 70
	const ticks = int64(220)

	genesis := func() *State {
		s := NewState(seed, m)
		for i := 1; i < len(s.Agents); i++ {
			s.Agents[i].Dead = true // lone living villager keeps the run quiet
		}
		s.GuardianCharges = 3
		return s
	}
	commands := map[int64][]store.Event{
		snapAt: {{Tick: snapAt, Type: "guardian.time_snapped",
			Payload: mustPayload(TimeSnappedPayload{ToTick: snapTo})}},
	}
	live := genesis()
	log := driveTicks(t, live, m, ticks, commands)

	var sawSnap bool
	for _, e := range log {
		if e.Type == "guardian.time_snapped" {
			sawSnap = true
		}
	}
	if !sawSnap {
		t.Fatal("scripted snap missing from the log")
	}
	if live.Tick != ticks {
		t.Fatalf("live tick = %d, want %d (snap jumped the clock then the run continued)", live.Tick, ticks)
	}

	replay := genesis()
	for _, e := range log {
		replay.Tick = e.Tick // set BEFORE apply: the snap re-bases from this tick
		if err := replay.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
	}
	driveTicks(t, replay, m, ticks, nil)
	if live.Hash() != replay.Hash() {
		t.Fatalf("snap replay diverged:\n live:     %s\n replayed: %s", string(live.Marshal()), string(replay.Marshal()))
	}
}

// --- US4 item grant (spec 016 T022) ---

// TestMiracleGrantHappy is US4-AS1: a grant to a living villager with free
// capacity lands the exact quantity and spends one charge.
func TestMiracleGrantHappy(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 3
	s.Agents[0].Inv = Inventory{} // known-empty pouch for an exact delta

	if err := applyMiracleErr(s, 100, "guardian.item_granted", ItemGrantedPayload{
		Agent: Ref(0), Kind: "food_raw", Qty: 3}); err != nil {
		t.Fatalf("grant rejected: %v", err)
	}
	if got := s.Agents[0].Inv.FoodRaw; got != 3 {
		t.Errorf("FoodRaw = %d, want 3", got)
	}
	if bulk(s.Agents[0].Inv) != 3 {
		t.Errorf("bulk = %d, want 3 (nothing else touched)", bulk(s.Agents[0].Inv))
	}
	if s.GuardianCharges != 2 {
		t.Errorf("charges = %d, want 2 (one spent)", s.GuardianCharges)
	}
}

// TestMiracleGrantOverCapWholeReject is US4-AS2 / FR-011: a grant that would
// overflow the carry cap is rejected whole — no partial delivery, no charge,
// inventory byte-identical to before.
func TestMiracleGrantOverCapWholeReject(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 3
	s.Agents[0].Inv = Inventory{Wood: bulkCap} // exactly full
	before := s.Marshal()

	err := applyMiracleErr(s, 40, "guardian.item_granted", ItemGrantedPayload{
		Agent: Ref(0), Kind: "wood", Qty: 1})
	if err == nil {
		t.Fatal("over-cap grant should be rejected whole")
	}
	if string(s.Marshal()) != string(before) {
		t.Error("rejected over-cap grant mutated state (partial application or charge spent)")
	}
	// Spec 095 T003 door regression: the digest headroom guidance added
	// turn-side (internal/guardian) and the give_item gloss (internal/tool)
	// must not have moved this door's own message one byte — applyItemGranted
	// (internal/sim/miracles.go) is untouched. (The event-type prefix says
	// guardian.item_granted since spec 094's rename — that is the ONE
	// sanctioned change to this string, carried by the log-format bump, not
	// by the spec-095 guidance this pin guards against.)
	wantMsg := fmt.Sprintf("apply guardian.item_granted: granting 1 wood to %s would exceed the carry cap (%d/%d already used)",
		AgentNames[0], bulkCap, bulkCap)
	if err.Error() != wantMsg {
		t.Errorf("door rejection message changed:\n got:  %q\n want: %q", err.Error(), wantMsg)
	}
}

// TestMiracleGrantUnknownKindReject is US4-AS3: an unknown item kind is rejected
// with nothing spent.
func TestMiracleGrantUnknownKindReject(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 3
	before := s.Marshal()

	err := applyMiracleErr(s, 40, "guardian.item_granted", ItemGrantedPayload{
		Agent: Ref(0), Kind: "gold", Qty: 1})
	if err == nil {
		t.Fatal("unknown item kind should be rejected")
	}
	// TASK-163: the door rejection must enumerate the grant vocabulary — the
	// live-measurement finding was a guardian repeatedly guessing kinds
	// ("food", "forage") because the rejection named only the bad guess, never
	// what WOULD have worked.
	for _, k := range tool.GrantKinds() {
		if !strings.Contains(err.Error(), k) {
			t.Errorf("door rejection %q does not enumerate grantable kind %q", err.Error(), k)
		}
	}
	if string(s.Marshal()) != string(before) {
		t.Error("rejected unknown-kind grant mutated state")
	}
	// "spears" (plural, the storage key) is NOT the grant vocabulary — a grant
	// names one fresh spear as "spear" (singular). The plural form is rejected.
	if err := applyMiracleErr(s, 41, "guardian.item_granted", ItemGrantedPayload{
		Agent: Ref(0), Kind: "spears", Qty: 1}); err == nil {
		t.Error(`"spears" (plural) should be rejected — the grant kind is "spear"`)
	}
}

// TestMiracleGrantDeadVillagerReject is US4-AS3: a grant to a dead villager (or
// an out-of-range index) is rejected with nothing spent.
func TestMiracleGrantDeadVillagerReject(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 3
	s.Agents[0].Dead = true
	before := s.Marshal()

	if err := applyMiracleErr(s, 40, "guardian.item_granted", ItemGrantedPayload{
		Agent: Ref(0), Kind: "food_raw", Qty: 1}); err == nil {
		t.Fatal("grant to a dead villager should be rejected")
	}
	if string(s.Marshal()) != string(before) {
		t.Error("rejected dead-villager grant mutated state")
	}
	if err := applyMiracleErr(s, 41, "guardian.item_granted", ItemGrantedPayload{
		Agent: Ref(len(s.Agents)), Kind: "food_raw", Qty: 1}); err == nil {
		t.Error("grant to an out-of-range agent index should be rejected")
	}
}

// TestMiracleGrantNonPositiveQtyReject: a zero or negative quantity is rejected.
func TestMiracleGrantNonPositiveQtyReject(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 3
	before := s.Marshal()

	for _, qty := range []int{0, -5} {
		if err := applyMiracleErr(s, 40, "guardian.item_granted", ItemGrantedPayload{
			Agent: Ref(0), Kind: "food_raw", Qty: qty}); err == nil {
			t.Errorf("grant of qty %d should be rejected", qty)
		}
	}
	if string(s.Marshal()) != string(before) {
		t.Error("rejected non-positive-qty grant mutated state")
	}
}

// TestMiracleGrantSpearShape: a spear grant appends that many fresh, full-
// durability spears, and the slice stays sorted ascending among any it already
// carries (hunts spend the most-worn first).
func TestMiracleGrantSpearShape(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 3
	s.Agents[0].Inv = Inventory{Spears: []int{1}} // one worn spear already carried

	if err := applyMiracleErr(s, 100, "guardian.item_granted", ItemGrantedPayload{
		Agent: Ref(0), Kind: "spear", Qty: 2}); err != nil {
		t.Fatalf("spear grant rejected: %v", err)
	}
	want := []int{1, spearDurability, spearDurability}
	if !reflect.DeepEqual(s.Agents[0].Inv.Spears, want) {
		t.Errorf("Spears = %v, want %v (two fresh full-use spears, ascending)", s.Agents[0].Inv.Spears, want)
	}
}

// TestMiracleGrantGratisZeroBank is US2 over US4: a forced grant lands with an
// empty bank (charge waived), and the bank stays at zero.
func TestMiracleGrantGratisZeroBank(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	s.GuardianCharges = 0
	s.Agents[0].Inv = Inventory{}

	if err := applyMiracleErr(s, 40, "guardian.item_granted", ItemGrantedPayload{
		Agent: Ref(0), Kind: "meals", Qty: 2, Gratis: true}); err != nil {
		t.Fatalf("gratis grant with an empty bank rejected: %v", err)
	}
	if s.Agents[0].Inv.Meals != 2 {
		t.Errorf("Meals = %d, want 2", s.Agents[0].Inv.Meals)
	}
	if s.GuardianCharges != 0 {
		t.Errorf("gratis grant spent a charge: bank = %d", s.GuardianCharges)
	}
}

// TestMiracleGrantReplayByteIdentity is SC-002 over US4: a scripted grant (incl.
// a spear grant) run replays from genesis (log only) to a byte-identical hash.
func TestMiracleGrantReplayByteIdentity(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	const ticks = 60

	genesis := func() *State {
		s := NewState(seed, m)
		for i := 1; i < len(s.Agents); i++ {
			s.Agents[i].Dead = true // lone living villager keeps the run quiet
		}
		s.GuardianCharges = 3
		s.Agents[0].Inv = Inventory{}
		return s
	}
	commands := map[int64][]store.Event{
		10: {{Tick: 10, Type: "guardian.item_granted", Payload: mustPayload(ItemGrantedPayload{
			Agent: Ref(0), Kind: "food_raw", Qty: 3})}},
		20: {{Tick: 20, Type: "guardian.item_granted", Payload: mustPayload(ItemGrantedPayload{
			Agent: Ref(0), Kind: "spear", Qty: 2})}},
	}

	live := genesis()
	log := driveTicks(t, live, m, ticks, commands)

	var grants int
	for _, e := range log {
		if e.Type == "guardian.item_granted" {
			grants++
		}
	}
	if grants != 2 {
		t.Fatalf("scripted grants missing from the log (saw %d, want 2)", grants)
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
		t.Fatalf("grant replay diverged:\nlive:     %s\nreplayed: %s", string(live.Marshal()), string(replay.Marshal()))
	}
}
