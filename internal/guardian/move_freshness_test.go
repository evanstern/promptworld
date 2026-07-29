package guardian

// Tests for spec 091 (move-miracle target freshness): door-side name
// re-resolution of a villager move's source position (decision (a) — x/y become
// advisory once a villager name is supplied). FR-005 coverage: a raced move
// lands when name-addressed, an unknown/dead name refuses before the charge, and
// the coordinate-only and structure/pile paths stay byte-identical to today.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// findAdjacentPassable returns a map-passable tile 4-adjacent to (x,y) — the
// TestMiracleFromDigestCoordinatesPassesDoor scan, reused here to place a
// villager's "post-race" position one step from where it was surveyed.
func findAdjacentPassable(m *worldmap.Map, x, y int) (int, int, bool) {
	for _, d := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		nx, ny := x+d[0], y+d[1]
		if m.Passable(nx, ny) {
			return nx, ny, true
		}
	}
	return 0, 0, false
}

// findPassableExcept scans the map for a passable tile that is none of the
// given points — used to pick a move destination distinct from the villager's
// source (surveyed or resolved) so a landed move is unambiguous.
func findPassableExcept(m *worldmap.Map, except ...[2]int) (int, int, bool) {
	skip := map[[2]int]bool{}
	for _, p := range except {
		skip[p] = true
	}
	for y := 0; y < m.H; y++ {
		for x := 0; x < m.W; x++ {
			if skip[[2]int{x, y}] {
				continue
			}
			if m.Passable(x, y) {
				return x, y, true
			}
		}
	}
	return 0, 0, false
}

// firstLivingVillager returns the index of a living villager from the
// guardian's own mirror (mt.agentXY / mt.alive), read under stateMu.
func firstLivingVillager(mt *Guardian) (idx, x, y int, ok bool) {
	mt.stateMu.Lock()
	defer mt.stateMu.Unlock()
	for i := range mt.agentXY {
		if mt.alive[i] {
			return i, mt.agentXY[i][0], mt.agentXY[i][1], true
		}
	}
	return -1, 0, 0, false
}

// TestLandMiracleMoveNameResolvesLivePosition (spec 091 FR-001/FR-002/SC-001): a
// villager walks away from the surveyed tile between survey and the call — the
// door re-resolves the NAMED villager's live position and the move lands there,
// instead of refusing with "no living villager at (x,y)" against the stale
// surveyed coordinates. The recorded event carries the RESOLVED coordinates.
func TestLandMiracleMoveNameResolvesLivePosition(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "It is moved.")

	vi, staleX, staleY, ok := firstLivingVillager(mt)
	if !ok {
		t.Fatal("no living villager in fixture")
	}
	nx, ny, ok := findAdjacentPassable(mt.m, staleX, staleY)
	if !ok {
		t.Skip("no adjacent passable tile for this seed")
	}

	// Simulate the race: the villager has ALREADY moved to (nx,ny) in both the
	// authoritative world state and the guardian's own absorb-mirrored copy —
	// exactly what a real metatron.entity_moved absorb would have produced —
	// while the surveyed coordinates the model still holds are the stale
	// (staleX,staleY).
	inj.state.Agents[vi].X, inj.state.Agents[vi].Y = nx, ny
	mt.stateMu.Lock()
	mt.agentXY[vi] = [2]int{nx, ny}
	mt.stateMu.Unlock()

	tx, ty, ok := findPassableExcept(mt.m, [2]int{nx, ny}, [2]int{staleX, staleY})
	if !ok {
		t.Skip("no destination tile distinct from source")
	}

	before := inj.state.GuardianCharges
	miracle, why := mt.landMiracle(miracleArgs{
		Kind: "move", Class: "villager", Villager: sim.AgentNames[vi],
		X: staleX, Y: staleY, ToX: tx, ToY: ty,
	}, before, fullGrant())
	if why != "" {
		t.Fatalf("name-addressed move rejected despite a raced survey: %q", why)
	}
	if miracle == nil {
		t.Fatal("no miracle returned despite a valid name-addressed move")
	}
	if inj.state.Agents[vi].X != tx || inj.state.Agents[vi].Y != ty {
		t.Errorf("villager did not land at (%d,%d): at (%d,%d)",
			tx, ty, inj.state.Agents[vi].X, inj.state.Agents[vi].Y)
	}
	if inj.state.GuardianCharges != before-1 {
		t.Errorf("move spent %d charges, want 1", before-inj.state.GuardianCharges)
	}
	// The recorded event carries the RESOLVED coordinates (emitter-computes,
	// FR-002), never the stale surveyed ones.
	found := false
	for _, b := range inj.batches {
		for _, e := range b {
			if e.Type != "guardian.entity_moved" {
				continue
			}
			var p sim.EntityMovedPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				t.Fatalf("decode entity_moved payload: %v", err)
			}
			if p.X != nx || p.Y != ny {
				t.Errorf("recorded event source = (%d,%d), want the resolved live position (%d,%d)", p.X, p.Y, nx, ny)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("no guardian.entity_moved event landed")
	}
}

// TestLandMiracleMoveUnknownNameRefusesBeforeCharge (spec 091 FR-001): a move
// naming a villager who does not exist refuses with the existing "no villager
// named" shape, before any charge is spent.
func TestLandMiracleMoveUnknownNameRefusesBeforeCharge(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "It is moved.")
	before := inj.state.GuardianCharges

	miracle, why := mt.landMiracle(miracleArgs{
		Kind: "move", Class: "villager", Villager: "Nobody",
		X: 1, Y: 1, ToX: 2, ToY: 2,
	}, before, fullGrant())
	if why == "" {
		t.Fatal("expected a refusal for an unknown villager name")
	}
	if !strings.Contains(why, `no villager named`) {
		t.Errorf("refusal = %q, want the existing \"no villager named\" shape", why)
	}
	if miracle != nil {
		t.Error("no miracle should land on an unknown name")
	}
	if inj.state.GuardianCharges != before {
		t.Errorf("charges spent on a pre-charge refusal: before=%d after=%d", before, inj.state.GuardianCharges)
	}
}

// TestLandMiracleMoveDeadNameRefusesBeforeCharge (spec 091 FR-001): a move
// naming a villager who has since died refuses before the charge, mirroring
// landVision's "beyond reach" resolution refusal.
func TestLandMiracleMoveDeadNameRefusesBeforeCharge(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "It is moved.")

	vi, x, y, ok := firstLivingVillager(mt)
	if !ok {
		t.Fatal("no living villager in fixture")
	}
	inj.state.Agents[vi].Dead = true
	mt.stateMu.Lock()
	mt.alive[vi] = false
	mt.stateMu.Unlock()

	before := inj.state.GuardianCharges
	miracle, why := mt.landMiracle(miracleArgs{
		Kind: "move", Class: "villager", Villager: sim.AgentNames[vi],
		X: x, Y: y, ToX: x, ToY: y,
	}, before, fullGrant())
	if why == "" {
		t.Fatal("expected a refusal for a dead villager name")
	}
	if !strings.Contains(why, "beyond reach") {
		t.Errorf("refusal = %q, want the landVision-style \"beyond reach\" shape", why)
	}
	if miracle != nil {
		t.Error("no miracle should land naming a dead villager")
	}
	if inj.state.GuardianCharges != before {
		t.Errorf("charges spent on a pre-charge refusal: before=%d after=%d", before, inj.state.GuardianCharges)
	}
}

// TestLandMiracleMoveCoordinateOnlyUnchanged (spec 091 FR-003): a villager move
// with NO name supplied takes today's exact coordinate-addressed path — it
// lands using the given x/y, unaffected by the name-resolution branch, and it
// still refuses (with the reducer's stale-coordinate message, now carrying the
// FR-004 suggestion) if the villager has since left those coordinates.
func TestLandMiracleMoveCoordinateOnlyUnchanged(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "It is moved.")

	vi, x, y, ok := firstLivingVillager(mt)
	if !ok {
		t.Fatal("no living villager in fixture")
	}
	tx, ty, ok := findAdjacentPassable(mt.m, x, y)
	if !ok {
		t.Skip("no adjacent passable tile for this seed")
	}

	before := inj.state.GuardianCharges
	miracle, why := mt.landMiracle(miracleArgs{
		Kind: "move", Class: "villager", X: x, Y: y, ToX: tx, ToY: ty,
	}, before, fullGrant())
	if why != "" {
		t.Fatalf("coordinate-only move rejected: %q", why)
	}
	if miracle == nil {
		t.Fatal("no miracle returned for a valid coordinate-only move")
	}
	if inj.state.Agents[vi].X != tx || inj.state.Agents[vi].Y != ty {
		t.Errorf("villager did not land at (%d,%d): at (%d,%d)", tx, ty, inj.state.Agents[vi].X, inj.state.Agents[vi].Y)
	}

	// Now race it: move the villager away from (tx,ty) via the world/mirror,
	// then repeat a coordinate-only call against the (now stale) source tile
	// the model still believes it at — must refuse exactly as today, with the
	// FR-004 name-preference suggestion appended.
	nx, ny, ok := findAdjacentPassable(mt.m, tx, ty)
	if !ok {
		t.Skip("no further adjacent passable tile for this seed")
	}
	inj.state.Agents[vi].X, inj.state.Agents[vi].Y = nx, ny
	mt.stateMu.Lock()
	mt.agentXY[vi] = [2]int{nx, ny}
	mt.stateMu.Unlock()
	// Top up the bank: the fixture genesis grants exactly 1 charge (spent by
	// the first move above) and this second call only tests the DOOR's
	// refusal, not charge scarcity.
	inj.state.GuardianCharges = 3

	before2 := inj.state.GuardianCharges
	dtx, dty, ok := findPassableExcept(mt.m, [2]int{tx, ty}, [2]int{nx, ny})
	if !ok {
		t.Skip("no destination tile distinct from source")
	}
	miracle2, why2 := mt.landMiracle(miracleArgs{
		Kind: "move", Class: "villager", X: tx, Y: ty, ToX: dtx, ToY: dty,
	}, before2, fullGrant())
	if why2 == "" {
		t.Fatal("expected the coordinate-only race to still refuse (FR-003: coordinates remain a legal, racy address form)")
	}
	if miracle2 != nil {
		t.Error("no miracle should land against a stale coordinate-only source")
	}
	if !strings.Contains(why2, "no living villager at") {
		t.Errorf("refusal = %q, want the reducer's unchanged \"no living villager at\" message", why2)
	}
	if !strings.Contains(why2, "name the villager") {
		t.Errorf("refusal = %q, want the FR-004 name-preference suggestion appended", why2)
	}
	if inj.state.GuardianCharges != before2 {
		t.Errorf("charges spent on a rejected move: before=%d after=%d", before2, inj.state.GuardianCharges)
	}
}

// TestLandMiracleMoveStructureIgnoresVillagerField (spec 091 FR-003): a
// structure (or pile) move never triggers name resolution, even if a stray
// villager name rides the call — class gates the whole mechanism.
func TestLandMiracleMoveStructureIgnoresVillagerField(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "It is moved.")

	sx, sy := 5, 5
	tx, ty, ok := findPassableExcept(mt.m, [2]int{sx, sy})
	if !ok {
		t.Skip("no destination tile for this seed")
	}
	inj.state.Structures = append(inj.state.Structures, sim.Structure{Kind: "chest", X: sx, Y: sy})

	before := inj.state.GuardianCharges
	miracle, why := mt.landMiracle(miracleArgs{
		Kind: "move", Class: "structure", Villager: sim.AgentNames[0],
		X: sx, Y: sy, ToX: tx, ToY: ty,
	}, before, fullGrant())
	if why != "" {
		t.Fatalf("structure move rejected: %q", why)
	}
	if miracle == nil {
		t.Fatal("no miracle returned for a valid structure move")
	}
	if !strings.Contains(miracle.Summary, "structure") {
		t.Errorf("summary = %q, want it to describe a structure move (not a villager)", miracle.Summary)
	}
}
