package mind

import (
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// TestSweepEncountersCrossingWalk (spec 104 T012): under the coalescing
// regime per-step agent.moved events vanish, so the encounter stimulus is
// the first-adjacency sweep over the ADVANCED replica. A scripted crossing
// walk must arm the pair at the adjacency moment, not re-arm while the pair
// lingers adjacent, and respect the per-pair cooldown on a later re-cross —
// the armEncounters semantics, replayed over derived steps.
func TestSweepEncountersCrossingWalk(t *testing.T) {
	tiles := make([]worldmap.TileKind, 32*32)
	for i := range tiles {
		tiles[i] = worldmap.Grass
	}
	m := &worldmap.Map{W: 32, H: 32, Tiles: tiles}
	replica := sim.NewState(42, m)
	set := replica.EffectiveTuning()
	set.NeedsCheckpointMinutes = 10 // regime ON
	set.EncounterCooldownTicks = 5000
	set.GruEmergePerMille = 0
	if err := replica.Apply(sim.NewTuningEvent(0, set)); err != nil {
		t.Fatal(err)
	}
	for i := range replica.Agents {
		replica.Agents[i].Dead = i > 1
	}
	a0, a1 := &replica.Agents[0], &replica.Agents[1]
	a0.X, a0.Y = 5, 5
	a1.X, a1.Y = 10, 5
	md := &Mind{replica: replica, pairSeen: map[[2]int]int64{}, pairAdjacent: map[[2]int]bool{}}
	md.pairSeen[[2]int{0, 1}] = -10000 // a genuine first meet

	// Agent 0 walks east toward agent 1 (stationary): one path_started, the
	// steps derived. Absorb the event, then advance batch by batch as ticks
	// pass (the absorb loop's own AdvanceTo + sweep, inlined).
	ev := store.Event{Tick: 2, Seq: 1, Type: "agent.path_started", Payload: mustPayloadJSON(t, sim.PathStartedPayload{
		Agent: sim.Ref(0), Path: []sim.Point{{X: 6, Y: 5}, {X: 7, Y: 5}, {X: 8, Y: 5}, {X: 9, Y: 5}},
		MoveEvery: 5, Phase: 0,
	})}
	if err := replica.Apply(ev); err != nil {
		t.Fatal(err)
	}
	armedAt := int64(-1)
	for tick := int64(2); tick <= 30; tick++ {
		replica.Tick = tick
		replica.AdvanceTo(tick)
		md.sweepEncounters(tick, 1)
		if md.pending[0] && armedAt < 0 {
			armedAt = tick
		}
	}
	// The step onto (9,5) — adjacency with (10,5) — fires at tick 20 (phase
	// 0, beats at 5/10/15/20); the sweep sees it advancing to tick 21.
	if armedAt != 21 {
		t.Fatalf("encounter armed at tick %d, want 21 (the first sweep after the adjacency step)", armedAt)
	}
	if !md.pending[1] {
		t.Fatal("the stationary side of the pair was not armed")
	}
	if md.pairSeen[[2]int{0, 1}] != 21 {
		t.Fatalf("pairSeen not stamped at the adjacency sweep: %d", md.pairSeen[[2]int{0, 1}])
	}
	// Lingering adjacent: no re-arm (the transition tracking).
	md.pending[0], md.pending[1] = false, false
	md.sweepEncounters(25, 2)
	if md.pending[0] || md.pending[1] {
		t.Fatal("lingering adjacency re-armed the pair")
	}
	// Separate, then re-cross inside the cooldown: still gated.
	a0.X = 3
	md.sweepEncounters(30, 3)
	a0.X = 9
	md.sweepEncounters(40, 4)
	if md.pending[0] || md.pending[1] {
		t.Fatal("re-cross inside the cooldown re-armed the pair")
	}
	// Re-cross after the cooldown: arms again.
	a0.X = 3
	md.sweepEncounters(5000, 5)
	a0.X = 9
	md.sweepEncounters(5221, 6)
	if !md.pending[0] || !md.pending[1] {
		t.Fatal("re-cross after the cooldown did not re-arm the pair")
	}
}

func mustPayloadJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
