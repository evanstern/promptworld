package bundle

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/toolloop"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// The spec-082 addressing fixture (T008/T010, SC-001/SC-003): the
// testdata/worlds/addressing bundle carries a declarative arg-templated
// structure move (relocate), a declarative text-param remove (smite), and a
// scripted tool composing a "pile@X,Y" address (raze). This file drives them
// end-to-end — a chest moves and later spills its contents when removed, a
// pile and a tree tile are erased (remove_entity demonstrably no longer
// inert) — then replays the recorded log against a fresh twin world with the
// bundle directory DELETED, proving the landed events are self-contained data
// (the script_replay_test.go bundle-independence pattern).

// addressingPlacements picks the fixture tiles as a pure function of the map:
// the first four grass tiles in row-major order (chest site, pile site, and
// two destinations — grass is passable AND a valid build site on a fresh
// state) and the first tree tile.
func addressingPlacements(t *testing.T, m *worldmap.Map) (chest, pile, destA, destB, tree [2]int) {
	t.Helper()
	var grass [][2]int
	tree = [2]int{-1, -1}
	for y := 0; y < m.H; y++ {
		for x := 0; x < m.W; x++ {
			switch m.At(x, y) {
			case worldmap.Grass:
				if len(grass) < 4 {
					grass = append(grass, [2]int{x, y})
				}
			case worldmap.Tree:
				if tree[0] < 0 {
					tree = [2]int{x, y}
				}
			}
		}
	}
	if len(grass) < 4 || tree[0] < 0 {
		t.Fatalf("fixture map lacks tiles: grass=%d tree=%v", len(grass), tree)
	}
	return grass[0], grass[1], grass[2], grass[3], tree
}

// addressingGenesis builds one deterministic fixture world: a chest with
// contents, a ground pile, banked charges. Called identically for the live
// and replay states so both start byte-identical.
func addressingGenesis(t *testing.T, seed uint64, m *worldmap.Map) *sim.State {
	t.Helper()
	s := sim.NewState(seed, m)
	chest, pile, _, _, _ := addressingPlacements(t, m)
	s.Structures = append(s.Structures, sim.Structure{
		Kind: "chest", X: chest[0], Y: chest[1], Store: &sim.Inventory{Wood: 3},
	})
	s.Piles = append(s.Piles, sim.Pile{X: pile[0], Y: pile[1], Wood: 2})
	s.GuardianCharges = 10 // headroom for four charged landings
	return s
}

// addressingProbe mirrors the guardian turn assembly's invocation probe shape
// (spec 082): villager roster plus structure/pile tiles plus the map for
// bounds — the read-only snapshot the effect compiler resolves against.
func addressingProbe(s *sim.State, m *worldmap.Map) *sim.State {
	p := &sim.State{Agents: make([]sim.Agent, len(s.Agents))}
	for i := range s.Agents {
		p.Agents[i] = sim.Agent{Name: s.Agents[i].Name, X: s.Agents[i].X, Y: s.Agents[i].Y, Dead: s.Agents[i].Dead}
	}
	p.Structures = make([]sim.Structure, len(s.Structures))
	for i := range s.Structures {
		p.Structures[i] = sim.Structure{Kind: s.Structures[i].Kind, X: s.Structures[i].X, Y: s.Structures[i].Y}
	}
	p.Piles = make([]sim.Pile, len(s.Piles))
	for i := range s.Piles {
		p.Piles[i] = sim.Pile{X: s.Piles[i].X, Y: s.Piles[i].Y}
	}
	p.SetMap(m)
	return p
}

// TestAddressingFixtureEndToEnd is SC-001 + SC-003 over US1/US2: every
// class+tile form lands through the real handler pipeline and the recorded
// log replays byte-identically with the bundle gone.
func TestAddressingFixtureEndToEnd(t *testing.T) {
	const seed = 42
	m := worldmap.Generate(seed, 64, 64)
	chest, pile, destA, _, tree := addressingPlacements(t, m)

	// Copy the fixture bundle to a temp dir so it can be DELETED before
	// replay (bundle-independence, FR-010's delete-bundle-dir proof).
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS(filepath.Join("testdata", "worlds", "addressing"))); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	bs, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if names := rosterNames(bs); len(names) != 3 {
		t.Fatalf("roster = %v, want [raze relocate smite]", names)
	}

	live := addressingGenesis(t, seed, m)
	var log []store.Event
	tick := int64(100)
	invoke := func(tool string, args map[string]any) toolloop.Outcome {
		t.Helper()
		// Stamp ticks BEFORE applying (the InjectSocial door's job live), so
		// the live application and the recorded log are the same bytes —
		// tick-dependent reducer arms (a chest spill's rot deadlines) then
		// replay identically. Fixture batches are single events, so the
		// door's atomicity is not exercised here (the compiler's
		// whole-invocation rejection is proven separately).
		inject := func(evs []store.Event) error {
			for _, e := range evs {
				e.Tick = tick
				if err := live.Apply(e); err != nil {
					return err
				}
				live.Tick = e.Tick
				log = append(log, e)
			}
			return nil
		}
		ic := InvocationContext{
			State: addressingProbe(live, m), Tick: tick, Invoker: "the guardian",
			Inject: inject, Seed: seed, MapWidth: m.W, MapHeight: m.H,
		}
		raw, _ := json.Marshal(args)
		out := bs.Handlers(ic)[tool](context.Background(), llm.ToolCall{Name: tool, Args: raw})
		tick += 100
		return out
	}
	mustLand := func(tool string, args map[string]any) {
		t.Helper()
		if out := invoke(tool, args); out.Verdict != toolloop.VerdictLanded {
			t.Fatalf("%s(%v): verdict = %q (%s)", tool, args, out.Verdict, out.ResultForModel)
		}
	}

	// US1 AC1: the declarative arg-templated structure move.
	mustLand("relocate", map[string]any{"x": chest[0], "y": chest[1], "to_x": destA[0], "to_y": destA[1]})
	if !live.HasStructureAt(destA[0], destA[1]) || live.HasStructureAt(chest[0], chest[1]) {
		t.Fatal("relocate did not move the chest")
	}
	// FR-007: the scripted tool composes the same grammar (pile removal).
	mustLand("raze", map[string]any{"x": pile[0], "y": pile[1]})
	if live.HasPileAt(pile[0], pile[1]) {
		t.Fatal("raze did not remove the pile")
	}
	// US2 AC1: terrain removal through the overlay vocabulary (tree → Cleared).
	mustLand("smite", map[string]any{"target": fmt.Sprintf("terrain@%d,%d", tree[0], tree[1])})
	cleared := false
	for _, p := range live.Cleared {
		if p.X == tree[0] && p.Y == tree[1] {
			cleared = true
		}
	}
	if !cleared {
		t.Fatal("smite did not clear the tree tile")
	}
	// US1: chest removal spills its Store to a ground pile (reducer unchanged).
	mustLand("smite", map[string]any{"target": fmt.Sprintf("structure@%d,%d", destA[0], destA[1])})
	if live.HasStructureAt(destA[0], destA[1]) {
		t.Fatal("smite did not remove the chest")
	}
	if !live.HasPileAt(destA[0], destA[1]) {
		t.Fatal("removing the chest did not spill its contents to a ground pile")
	}

	// US2 AC3 shape + spec-036 invariant: a reducer-side rejection (grass is
	// not removable terrain) is whole-invocation — nothing lands, nothing is
	// spent, state is untouched. destA now holds the spilled pile, so use a
	// fresh grass probe: the chest's ORIGINAL tile is grass and empty again.
	before, chargesBefore := live.Hash(), live.GuardianCharges
	out := invoke("smite", map[string]any{"target": fmt.Sprintf("terrain@%d,%d", chest[0], chest[1])})
	if out.Verdict != toolloop.VerdictRejectedGate {
		t.Fatalf("grass removal verdict = %q, want rejected_gate", out.Verdict)
	}
	if live.Hash() != before || live.GuardianCharges != chargesBefore {
		t.Fatal("a rejected invocation mutated state or spent a charge")
	}

	// SC-003: delete the bundle, replay the recorded log on a fresh twin.
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	replay := addressingGenesis(t, seed, m)
	for _, e := range log {
		if err := replay.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replay.Tick = e.Tick
	}
	if live.Hash() != replay.Hash() {
		t.Fatalf("replay diverged:\nlive:     %s\nreplayed: %s", live.Marshal(), replay.Marshal())
	}
}
