package bundle

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// TestBundleToolReplayByteIdentity is SC-003 over US1, following the
// internal/sim/miracles_test.go:398 live-vs-replay pattern: a bundle tool's
// compiled effect batch, applied live and then re-applied from the recorded
// event log alone, yields a byte-identical State.Hash(). Because declarative
// bundle tools carry no runtime logic — CompileEffects is a pure function of
// (templates, args, state) — the recorded events are the whole story: replay
// never re-executes the tool, so it reproduces the exact same state even with the
// bundle absent (FR-011). The executor-side replay of these SAME event types
// (metatron.entity_moved, agent.memory_added) is already proven by
// TestMiracleReplayByteIdentity; this test pins the compiler + apply half that
// bundle tools add.
func TestBundleToolReplayByteIdentity(t *testing.T) {
	const seed = 42
	m := worldmap.Generate(seed, 64, 64)

	// genesis: a fresh world with charges banked. NewState keeps its map, so the
	// move reducer's passability check resolves.
	genesis := func() *sim.State {
		s := sim.NewState(seed, m)
		s.GuardianCharges = sim.GuardianChargeCap
		return s
	}

	// The teleport batch: move Ash onto Birch's (living, passable) tile and
	// narrate to every living villager. Compiled against a genesis snapshot so
	// target/recipient resolution is deterministic.
	base := genesis()
	bx, by := base.Agents[1].X, base.Agents[1].Y
	templates := teleportTemplates(t)
	in := CompileInput{
		State:    base,
		Tick:     0,
		Args:     map[string]string{"target": "Ash", "x": strconv.Itoa(bx), "y": strconv.Itoa(by)},
		Invoker:  "the guardian",
		Declared: map[string]bool{"metatron.entity_moved": true, "agent.memory_added": true},
	}
	effects, err := ExpandTemplates(templates, in)
	if err != nil {
		t.Fatalf("ExpandTemplates: %v", err)
	}
	batch, err := CompileEffects(effects, in)
	if err != nil {
		t.Fatalf("CompileEffects: %v", err)
	}

	// Compiler determinism: an independent compilation produces byte-identical
	// events (the replay guarantee starts here — the batch is a pure function).
	batch2, err := CompileEffects(mustExpand(t, templates, in), in)
	if err != nil {
		t.Fatalf("CompileEffects (2nd): %v", err)
	}
	assertSameEvents(t, batch, batch2)

	// Stamp ticks (the InjectSocial door does this live); the log is the stamped
	// batch. Both the live and replay states apply the identical log.
	log := make([]store.Event, len(batch))
	for i, e := range batch {
		e.Tick = 100
		log[i] = e
	}

	live := genesis()
	for _, e := range log {
		if err := live.Apply(e); err != nil {
			t.Fatalf("live apply %s: %v", e.Type, err)
		}
		live.Tick = e.Tick
	}
	// The batch actually changed the world (a no-op replay proof would be hollow).
	if live.Hash() == genesis().Hash() {
		t.Fatal("teleport batch left state unchanged")
	}

	replay := genesis()
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

// teleportTemplates parses the teleport fixture's effect templates.
func teleportTemplates(t *testing.T) []effectTemplate {
	t.Helper()
	bs := teleportSet(t)
	for _, b := range bs.Bundles() {
		for _, tl := range b.Tools {
			if tl.Name == "teleport" {
				return tl.Templates
			}
		}
	}
	t.Fatal("teleport tool not found")
	return nil
}

func mustExpand(t *testing.T, templates []effectTemplate, in CompileInput) []Effect {
	t.Helper()
	e, err := ExpandTemplates(templates, in)
	if err != nil {
		t.Fatalf("ExpandTemplates: %v", err)
	}
	return e
}

func assertSameEvents(t *testing.T, a, b []store.Event) {
	t.Helper()
	if len(a) != len(b) {
		t.Fatalf("event counts differ: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i].Type != b[i].Type || !bytes.Equal(a[i].Payload, b[i].Payload) {
			t.Errorf("event %d differs:\n a: %s %s\n b: %s %s", i, a[i].Type, a[i].Payload, b[i].Type, b[i].Payload)
		}
	}
}
