package mind

import (
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// stubInjector / stubSocial are inert loop seams for constructing a Mind whose
// only behavior under test is the boot-time cadence stagger.
type stubInjector struct{}

func (stubInjector) InjectIntent(sim.InjectArgs) error { return nil }

type stubSocial struct{}

func (stubSocial) InjectSocial([]store.Event) error { return nil }

// TestTunedPlannerCadenceShiftsStagger (spec 048 US3, SC-005): the mind's
// boot-time per-agent stagger is derived from the replica's PlannerCadence()
// accessor, so a tuned cadence — carried into the replica via the state JSON's
// sim.tuning_applied event — moves every nextDue, not the default constant.
func TestTunedPlannerCadenceShiftsStagger(t *testing.T) {
	m := worldmap.Generate(42, 64, 64)

	newMind := func(cadence int64) *Mind {
		t.Helper()
		state := sim.NewState(42, m)
		if cadence != 0 {
			set := sim.TuningState{
				RefuelDyingBelow:       3600,
				FireBurnPerWood:        14400,
				GruEmergePerMille:      600,
				PlannerCadenceTicks:    cadence,
				EncounterCooldownTicks: 7200,
			}
			if err := state.Apply(sim.NewTuningEvent(state.Tick, set)); err != nil {
				t.Fatal(err)
			}
		}
		md, err := New(&mockModel{reply: "{}"}, stubInjector{}, stubSocial{}, m, 42,
			state.Marshal(), [sim.AgentCount]string{}, testLoopRounds, testPlannerTokens, testConsolidationTokens, "", noopLoop)
		if err != nil {
			t.Fatal(err)
		}
		return md
	}

	// Default cadence (no tuning event): nextDue[i] = (i+1) * (1800/8).
	base := newMind(0)
	defer base.Close()
	for i := 0; i < sim.AgentCount; i++ {
		want := int64(i+1) * (1800 / int64(sim.AgentCount))
		if base.nextDue[i] != want {
			t.Fatalf("default stagger nextDue[%d] = %d, want %d", i, base.nextDue[i], want)
		}
	}

	// Tuned cadence 900 (half the default): every stagger step halves.
	tuned := newMind(900)
	defer tuned.Close()
	for i := 0; i < sim.AgentCount; i++ {
		want := int64(i+1) * (900 / int64(sim.AgentCount))
		if tuned.nextDue[i] != want {
			t.Fatalf("tuned stagger nextDue[%d] = %d, want %d (tuned cadence 900)", i, tuned.nextDue[i], want)
		}
		if tuned.nextDue[i] == int64(i+1)*(1800/int64(sim.AgentCount)) {
			t.Errorf("stagger nextDue[%d] used the DEFAULT cadence, not the tuned 900", i)
		}
	}
}

// TestTunedEncounterCooldownGates (spec 048 US3, SC-005): armEncounters gates a
// re-encounter on md.replica.EncounterCooldown(). Two agents adjacent again
// after a gap shorter than the tuned cooldown must NOT re-arm; after a gap that
// meets it, they must. Built as a struct literal (no goroutines) so the gate
// arithmetic is exercised in isolation.
func TestTunedEncounterCooldownGates(t *testing.T) {
	m := worldmap.Generate(42, 64, 64)

	// Replica with a tuned cooldown of 5000 ticks and agents 0,1 adjacent.
	state := sim.NewState(42, m)
	set := sim.TuningState{
		RefuelDyingBelow:       3600,
		FireBurnPerWood:        14400,
		GruEmergePerMille:      600,
		PlannerCadenceTicks:    1800,
		EncounterCooldownTicks: 5000,
	}
	if err := state.Apply(sim.NewTuningEvent(state.Tick, set)); err != nil {
		t.Fatal(err)
	}
	// Place agent 1 orthogonally adjacent to agent 0 (within encounterRadius=1).
	state.Agents[1].X = state.Agents[0].X + 1
	state.Agents[1].Y = state.Agents[0].Y
	state.Agents[1].Dead = false
	state.Agents[0].Dead = false

	md := &Mind{replica: state, pairSeen: map[[2]int]int64{}}
	if got := md.replica.EncounterCooldown(); got != 5000 {
		t.Fatalf("replica EncounterCooldown() = %d, want tuned 5000", got)
	}

	moved := func(tick int64) store.Event {
		p, _ := json.Marshal(sim.AgentMovedPayload{Agent: sim.Ref(0), X: state.Agents[0].X, Y: state.Agents[0].Y})
		return store.Event{Tick: tick, Seq: tick, Type: "agent.moved", Payload: p}
	}

	// First adjacency at tick 0 arms the pair (pairSeen starts at 0; 0-0 >= 5000
	// is false, so the FIRST encounter is actually gated too — seed pairSeen to a
	// negative baseline to model a genuine first meet).
	md.pairSeen[[2]int{0, 1}] = -10000 // last seen long ago → first meet arms
	md.armEncounters(moved(0))
	if md.pairSeen[[2]int{0, 1}] != 0 {
		t.Fatalf("first encounter did not arm the pair (pairSeen = %d)", md.pairSeen[[2]int{0, 1}])
	}

	// Re-encounter at tick 4000 (< tuned 5000): gated, pairSeen unchanged.
	md.armEncounters(moved(4000))
	if md.pairSeen[[2]int{0, 1}] != 0 {
		t.Errorf("re-encounter at 4000 (< cooldown 5000) re-armed — pairSeen = %d, want 0", md.pairSeen[[2]int{0, 1}])
	}

	// Re-encounter at tick 5000 (== tuned cooldown): admitted, pairSeen advances.
	md.armEncounters(moved(5000))
	if md.pairSeen[[2]int{0, 1}] != 5000 {
		t.Errorf("re-encounter at 5000 (>= cooldown) did not re-arm — pairSeen = %d, want 5000", md.pairSeen[[2]int{0, 1}])
	}
}
