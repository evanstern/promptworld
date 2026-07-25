package guardian

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/evanstern/promptworld/internal/bundle"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/toolloop"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// TestDogfoodMiracleMoveEquivalence is US2 / quickstart Scenario 5 (T020, SC-004):
// the shipped dogfood bundle (examples/bundles/dogfood-move) re-expresses the
// built-in work_miracle{kind:move}. On identical-seed twin worlds, the built-in
// door (BuildMiracleBatch — the shared builder BOTH channels compose through) and
// the bundle door (the handler factory over the miracle_move manifest) must
// produce byte-identical event batches for the same logical move, and applying
// each to its twin must leave byte-identical state (same entity_moved payload,
// same perception memory, same reducer-enforced charge deduction).
//
// The equivalence is proven at the batch level (Type + Payload bytes) AND at the
// state level (State.Hash() + charge balance), because the reducer keys the
// charge spend off the arriving event type, so identical batches ⇒ identical
// spends without any turn-side charge gate.
func TestDogfoodMiracleMoveEquivalence(t *testing.T) {
	const seed = 42
	m := worldmap.Generate(seed, 64, 64)
	genesis := func() *sim.State {
		s := sim.NewState(seed, m)
		s.GuardianCharges = sim.GuardianChargeCap
		return s
	}

	builtinState := genesis()
	bundleState := genesis()

	// The logical move: lift Ash (index 0) onto Birch's (index 1) living, passable
	// tile. The built-in addresses the move by SOURCE TILE (x,y); the bundle
	// addresses it by TARGET NAME — the same villager, so both resolve to the same
	// source tile and the same moved-villager memory recipient.
	ax, ay := builtinState.Agents[0].X, builtinState.Agents[0].Y
	bx, by := builtinState.Agents[1].X, builtinState.Agents[1].Y
	if ax == bx && ay == by {
		t.Skip("fixture seed placed Ash and Birch on the same tile")
	}

	// --- Built-in channel: the shared batch builder both doors compose through. ---
	builtinBatch, err := BuildMiracleBatch(builtinState, "move",
		MiracleParams{Class: "villager", X: ax, Y: ay, ToX: bx, ToY: by}, false)
	if err != nil {
		t.Fatalf("BuildMiracleBatch: %v", err)
	}

	// --- Bundle channel: the miracle_move handler over the shipped dogfood bundle. ---
	bs, err := bundle.Discover(filepath.Join("..", "..", "examples"))
	if err != nil {
		t.Fatalf("Discover(examples): %v", err)
	}
	if _, ok := hasTool(bs, "miracle_move"); !ok {
		t.Fatalf("dogfood bundle missing miracle_move: roster=%v, report=%+v", rosterOf(bs), bs.BootReport())
	}

	var bundleBatch []store.Event
	inject := func(evs []store.Event) error {
		for _, e := range evs {
			if err := bundleState.Apply(e); err != nil {
				return err
			}
		}
		bundleBatch = append(bundleBatch, evs...)
		return nil
	}
	probe := &sim.State{Agents: make([]sim.Agent, len(bundleState.Agents))}
	for i := range bundleState.Agents {
		probe.Agents[i] = sim.Agent{
			Name: bundleState.Agents[i].Name, X: bundleState.Agents[i].X,
			Y: bundleState.Agents[i].Y, Dead: bundleState.Agents[i].Dead}
	}
	ic := bundle.InvocationContext{State: probe, Tick: 0, Invoker: "the guardian", Inject: inject}
	h := bs.Handlers(ic)["miracle_move"]
	if h == nil {
		t.Fatal("no handler for miracle_move")
	}
	args, _ := json.Marshal(map[string]any{"target": "Ash", "to_x": bx, "to_y": by})
	out := h(context.Background(), llm.ToolCall{ID: "c1", Name: "miracle_move", Args: args})
	if out.Verdict != toolloop.VerdictLanded {
		t.Fatalf("miracle_move verdict = %q (%s), want landed", out.Verdict, out.ResultForModel)
	}

	// Byte-identical batches: same event types in the same order, same payloads.
	if len(builtinBatch) != len(bundleBatch) {
		t.Fatalf("batch lengths differ: built-in %d, bundle %d", len(builtinBatch), len(bundleBatch))
	}
	for i := range builtinBatch {
		if builtinBatch[i].Type != bundleBatch[i].Type || !bytes.Equal(builtinBatch[i].Payload, bundleBatch[i].Payload) {
			t.Errorf("event %d differs:\n built-in: %s %s\n bundle:   %s %s",
				i, builtinBatch[i].Type, builtinBatch[i].Payload, bundleBatch[i].Type, bundleBatch[i].Payload)
		}
	}

	// The built-in twin applies the built-in batch; the bundle twin already applied
	// its batch through the injector. Identical state ⇒ same move, same memory, same
	// charge spend (the reducer deducts on metatron.entity_moved for both).
	for _, e := range builtinBatch {
		if err := builtinState.Apply(e); err != nil {
			t.Fatalf("built-in apply %s: %v", e.Type, err)
		}
	}
	if builtinState.Hash() != bundleState.Hash() {
		t.Fatalf("twin states diverged:\n built-in: %s\n bundle:   %s", builtinState.Marshal(), bundleState.Marshal())
	}
	// The move actually spent a charge (an equivalence over a no-op would be hollow).
	if builtinState.GuardianCharges != sim.GuardianChargeCap-1 {
		t.Errorf("built-in charges = %d, want %d", builtinState.GuardianCharges, sim.GuardianChargeCap-1)
	}
	if bundleState.GuardianCharges != builtinState.GuardianCharges {
		t.Errorf("charge deduction diverged: built-in %d, bundle %d", builtinState.GuardianCharges, bundleState.GuardianCharges)
	}
}

// hasTool reports whether the set's roster carries a tool by name.
func hasTool(bs *bundle.BundleSet, name string) (int, bool) {
	for i, tl := range bs.Roster() {
		if tl.Name == name {
			return i, true
		}
	}
	return 0, false
}

func rosterOf(bs *bundle.BundleSet) []string {
	var out []string
	for _, tl := range bs.Roster() {
		out = append(out, tl.Name)
	}
	return out
}
