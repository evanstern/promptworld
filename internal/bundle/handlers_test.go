package bundle

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/toolloop"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// teleportSet discovers the declarative teleport fixture world and returns its
// BundleSet (roster: [teleport]).
func teleportSet(t *testing.T) *BundleSet {
	t.Helper()
	bs, err := Discover(filepath.Join("testdata", "worlds", "declarative"))
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if names := rosterNames(bs); len(names) != 1 || names[0] != "teleport" {
		t.Fatalf("roster = %v, want [teleport]", names)
	}
	return bs
}

// captureInjector records the batch it is handed and applies it to a state
// through the real reducer (dry-run on a copy first), like the InjectSocial door.
type captureInjector struct {
	state   *sim.State
	batches [][]store.Event
	failErr error
}

func (ci *captureInjector) inject(events []store.Event) error {
	if ci.failErr != nil {
		return ci.failErr
	}
	for _, e := range events {
		if err := ci.state.Apply(e); err != nil {
			return err
		}
	}
	ci.batches = append(ci.batches, events)
	return nil
}

// teleportWorld builds a fresh sim world for the handler tests: a real map and
// state with all eight villagers alive, charges banked.
func teleportWorld(t *testing.T) (*sim.State, *worldmap.Map) {
	t.Helper()
	m := worldmap.Generate(42, 64, 64)
	s := sim.NewState(42, m)
	s.GuardianCharges = sim.GuardianChargeCap
	return s, m
}

// TestHandlerLandsBatch: a valid teleport invocation compiles to the declared
// events and lands them through the door — the target villager moves and every
// living villager gains the narration memory.
func TestHandlerLandsBatch(t *testing.T) {
	bs := teleportSet(t)
	s, _ := teleportWorld(t)
	ci := &captureInjector{state: s}

	// Move Ash (index 0) onto Birch's tile (index 1) — a passable tile another
	// living villager occupies, so the reducer's passability check always holds.
	ax, ay := s.Agents[0].X, s.Agents[0].Y
	bx, by := s.Agents[1].X, s.Agents[1].Y

	ic := InvocationContext{State: snapshot(s), Tick: 100, Invoker: "the guardian", Inject: ci.inject}
	h := bs.Handlers(ic)["teleport"]
	args, _ := json.Marshal(map[string]any{"target": "Ash", "x": bx, "y": by})
	out := h(context.Background(), llm.ToolCall{ID: "c1", Name: "teleport", Args: args})

	if out.Verdict != toolloop.VerdictLanded {
		t.Fatalf("verdict = %q (%s), want landed", out.Verdict, out.ResultForModel)
	}
	if s.Agents[0].X != bx || s.Agents[0].Y != by {
		t.Errorf("Ash at (%d,%d), want (%d,%d)", s.Agents[0].X, s.Agents[0].Y, bx, by)
	}
	if ax == bx && ay == by {
		t.Fatal("fixture precondition: Ash and Birch share a tile")
	}
	// One batch: the move + one memory per living villager, only the declared types.
	if len(ci.batches) != 1 {
		t.Fatalf("batches = %d, want 1", len(ci.batches))
	}
	var moves, mems int
	for _, e := range ci.batches[0] {
		switch e.Type {
		case "metatron.entity_moved":
			moves++
		case "agent.memory_added":
			mems++
		default:
			t.Errorf("undeclared event type %q landed", e.Type)
		}
	}
	if moves != 1 || mems != len(s.LivingAgents()) {
		t.Errorf("batch = %d moves, %d memories; want 1 move, %d memories", moves, mems, len(s.LivingAgents()))
	}
}

// TestHandlerRejectsUnknownTarget: a target that is not a living villager is an
// author-level failure — a rejected_gate with a specific reason, never an
// Outcome.Err, and nothing lands.
func TestHandlerRejectsUnknownTarget(t *testing.T) {
	bs := teleportSet(t)
	s, _ := teleportWorld(t)
	ci := &captureInjector{state: s}
	ic := InvocationContext{State: snapshot(s), Tick: 100, Invoker: "the guardian", Inject: ci.inject}
	h := bs.Handlers(ic)["teleport"]
	args, _ := json.Marshal(map[string]any{"target": "Nobody", "x": 1, "y": 1})
	out := h(context.Background(), llm.ToolCall{Name: "teleport", Args: args})
	if out.Verdict != toolloop.VerdictRejectedGate {
		t.Errorf("verdict = %q, want rejected_gate", out.Verdict)
	}
	if out.Err != nil {
		t.Errorf("Err = %v, want nil (author-level failure)", out.Err)
	}
	if len(ci.batches) != 0 {
		t.Error("a rejected invocation landed a batch")
	}
}

// TestHandlerDoorRefusalIsRejectedGate: a door refusal (the injector rejects the
// batch) surfaces as a rejected_gate carrying the door's reason, not an
// infrastructure error.
func TestHandlerDoorRefusalIsRejectedGate(t *testing.T) {
	bs := teleportSet(t)
	s, _ := teleportWorld(t)
	ci := &captureInjector{state: s, failErr: context.DeadlineExceeded}
	ic := InvocationContext{State: snapshot(s), Tick: 100, Invoker: "the guardian", Inject: ci.inject}
	h := bs.Handlers(ic)["teleport"]
	bx, by := s.Agents[1].X, s.Agents[1].Y
	args, _ := json.Marshal(map[string]any{"target": "Ash", "x": bx, "y": by})
	out := h(context.Background(), llm.ToolCall{Name: "teleport", Args: args})
	if out.Verdict != toolloop.VerdictRejectedGate || out.Err != nil {
		t.Errorf("verdict = %q, err = %v; want rejected_gate, nil", out.Verdict, out.Err)
	}
}

// snapshot builds a read-only probe with the fields the effect compiler resolves
// against (name/position/liveness), mirroring the guardian turn assembly's probe.
func snapshot(s *sim.State) *sim.State {
	p := &sim.State{Agents: make([]sim.Agent, len(s.Agents))}
	for i := range s.Agents {
		p.Agents[i] = sim.Agent{Name: s.Agents[i].Name, X: s.Agents[i].X, Y: s.Agents[i].Y, Dead: s.Agents[i].Dead}
	}
	return p
}
