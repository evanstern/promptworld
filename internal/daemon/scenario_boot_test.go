package daemon

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Boot arming tests (spec 054 T006): the daemon's manifest → armed-loop leg,
// isolated in armScenario so it is testable without a full daemon run.

func bootState(seed uint64) *sim.State {
	return sim.NewState(seed, worldmap.Generate(seed, worldmap.DefaultSize, worldmap.DefaultSize))
}

// TestArmScenarioFromManifest: a scenario-stamped manifest arms the state
// with its exercise (boot-frozen — ScenarioExerciseID reports it).
func TestArmScenarioFromManifest(t *testing.T) {
	w := &world.World{Manifest: world.Manifest{
		Name: "scen", Seed: sim.FirstNightExercise.Seed,
		Scenario: &world.ScenarioConfig{Exercise: "first-night"},
	}}
	state := bootState(w.Manifest.Seed)
	id, err := armScenario(w, state)
	if err != nil {
		t.Fatalf("armScenario: %v", err)
	}
	if id != "first-night" || state.ScenarioExerciseID() != "first-night" {
		t.Errorf("armed id = %q, state reports %q", id, state.ScenarioExerciseID())
	}
}

// TestArmScenarioAmbientNoop: no scenario block arms nothing — the ambient
// byte-identity contract's boot leg (§1.3).
func TestArmScenarioAmbientNoop(t *testing.T) {
	w := &world.World{Manifest: world.Manifest{Name: "ambient", Seed: 7}}
	state := bootState(7)
	id, err := armScenario(w, state)
	if err != nil {
		t.Fatalf("armScenario: %v", err)
	}
	if id != "" || state.ScenarioExerciseID() != "" {
		t.Errorf("ambient world armed %q / %q, want nothing", id, state.ScenarioExerciseID())
	}
}

// TestArmScenarioRefusesUncataloged: a manifest naming an exercise outside
// the compiled catalog is a boot failure (belt behind world.Open's mirror
// validation), never a silent ambient boot.
func TestArmScenarioRefusesUncataloged(t *testing.T) {
	w := &world.World{Manifest: world.Manifest{
		Name: "bad", Seed: 7,
		Scenario: &world.ScenarioConfig{Exercise: "not-shipped"},
	}}
	_, err := armScenario(w, bootState(7))
	if err == nil || !strings.Contains(err.Error(), "not-shipped") {
		t.Fatalf("want a refusal naming the exercise, got %v", err)
	}
}
