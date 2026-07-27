package mind

import (
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// TestAbsorbHarvestRearmsWitness (spec 081 T013, FR-007 / contracts §5): a
// chop/quarry absorbed by the mind re-arms its actor AND any on-scene witness
// whose live intent targeted the cleared tile — the act-time silent map removal
// stands in for the map_corrected these witnesses used to re-arm on. A witness
// with an unrelated intent, or one out of witness radius, stays quiet.
func TestAbsorbHarvestRearmsWitness(t *testing.T) {
	state := sim.NewState(42, worldmap.Generate(42, 64, 64))
	state.Paused = false
	state.Tick = 5000

	const tx, ty = 20, 20
	// Actor 0 at the tile; witnesses positioned around it.
	place := func(i, x, y int) { state.Agents[i].X, state.Agents[i].Y = x, y }
	place(0, tx, ty)
	place(1, tx+3, ty)                   // in radius, intent targets the tile → re-arm
	place(2, tx+3, ty)                   // in radius, unrelated intent → quiet
	place(3, tx, ty+sim.WitnessRadius+4) // out of radius, intent targets the tile → quiet
	state.Agents[1].Intent = &sim.Intent{Goal: "chop", TargetX: tx, TargetY: ty}
	state.Agents[2].Intent = &sim.Intent{Goal: "forage", TargetX: 0, TargetY: 0}
	state.Agents[3].Intent = &sim.Intent{Goal: "chop", ResX: tx, ResY: ty}

	md := &Mind{replica: state}
	md.absorb([]store.Event{{Tick: 5000, Seq: 7, Type: "agent.chopped",
		Payload: mustJSON(t, sim.HarvestPayload{Agent: 0, X: tx, Y: ty})}})

	if !md.pending[0] {
		t.Error("actor was not re-armed by its own chop")
	}
	if !md.pending[1] {
		t.Error("in-radius witness whose intent targeted the felled tile was not re-armed")
	}
	if md.pending[2] {
		t.Error("in-radius witness with an unrelated intent was wrongly re-armed")
	}
	if md.pending[3] {
		t.Error("out-of-radius agent (fact never removed) was wrongly re-armed — it must discover on return")
	}
}
