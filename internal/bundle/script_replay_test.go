package bundle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// randScript drives its effect batch off world.rand — the ONLY randomness a script
// may draw, coordinate-seeded from (world seed, "bundle:randtool:when", tick, 0).
// The forward time delta and the resulting batch are therefore a pure function of
// (seed, tick), so two executions are byte-identical and replay reproduces the
// state from the recorded events alone.
const randScript = `
def apply(args, world):
    r = world.rand("when", 0)
    dt = 100 + int(r * 100)
    return [
        {"kind": "snap_time", "to_tick": world.tick + dt},
        {"kind": "narrate", "text": "Time lurches forward.", "recipients": "all_living"},
    ]
`

// TestScriptedReplayByteIdentity is SC-003 / FR-011 over US3 (T028), the scripted
// variant of TestBundleToolReplayByteIdentity exercising world.rand. It proves
// two things: (1) determinism — a rand-driven script compiles to a byte-identical
// batch on repeated execution, and applying it live vs. replaying the recorded log
// yields an identical State.Hash(); (2) bundle-independence — after the batch is
// recorded, the bundle directory is DELETED, and replay still reproduces the hash,
// because replay applies self-contained event data and never re-executes the script.
func TestScriptedReplayByteIdentity(t *testing.T) {
	const seed = 42
	m := worldmap.Generate(seed, 64, 64)
	genesis := func() *sim.State {
		s := sim.NewState(seed, m)
		s.GuardianCharges = sim.GuardianChargeCap
		return s
	}

	// Author a rand-driven scripted tool in a throwaway bundle dir and compile it.
	dir := t.TempDir()
	toolDir := filepath.Join(dir, "randtool")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "tool.star"), []byte(randScript), 0o644); err != nil {
		t.Fatal(err)
	}
	sp, err := compileScript(toolDir, "tool.star")
	if err != nil {
		t.Fatalf("compileScript: %v", err)
	}

	declared := map[string]bool{"metatron.time_snapped": true, "agent.memory_added": true}
	compile := func() []store.Event {
		s := genesis()
		world := newWorldView("randtool", 0, seed, m.W, m.H, agentInfos(s))
		args, _ := scriptArgs(&Manifest{}, nil)
		effects, err := sp.execute(args, world, defaultMaxSteps)
		if err != nil {
			t.Fatalf("execute: %v", err)
		}
		batch, err := CompileEffects(effects, CompileInput{State: s, Tick: 0, Declared: declared})
		if err != nil {
			t.Fatalf("CompileEffects: %v", err)
		}
		return batch
	}

	// world.rand determinism: two independent executions are byte-identical.
	batch := compile()
	assertSameEvents(t, batch, compile())
	if len(batch) < 2 {
		t.Fatalf("rand-driven batch has %d events, want the snap plus narration", len(batch))
	}

	// Record the log, then DELETE the bundle — replay must not need it (FR-011).
	log := make([]store.Event, len(batch))
	for i, e := range batch {
		e.Tick = 50
		log[i] = e
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}

	live := genesis()
	for _, e := range log {
		if err := live.Apply(e); err != nil {
			t.Fatalf("live apply %s: %v", e.Type, err)
		}
	}
	if live.Hash() == genesis().Hash() {
		t.Fatal("rand-driven batch left state unchanged")
	}

	replay := genesis()
	for _, e := range log {
		if err := replay.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
	}
	if live.Hash() != replay.Hash() {
		t.Fatalf("replay diverged:\nlive:     %s\nreplayed: %s", live.Marshal(), replay.Marshal())
	}
}
