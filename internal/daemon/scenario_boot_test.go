package daemon

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worldmap"
	"github.com/evanstern/promptworld/internal/worlds"
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

// TestSeedSurvivalWatchesReplaySeqConsistency pins the boot seeder's seq
// pre-stamping (spec 054 × spec 059): the metatron.order_placed reducer arm
// stamps GuardianOrder.PlacedSeq from the event envelope, so the seeder must
// apply with the SAME seq AppendEvents will record (the loop's stampSeqs
// contract) — or the live boot state diverges from replay and snapshot+tail
// splits from full replay (the e2e SC-003 failure this fixed).
func TestSeedSurvivalWatchesReplaySeqConsistency(t *testing.T) {
	dir := t.TempDir()
	w, err := world.Create(dir, "seqworld", 5)
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(filepath.Join(dir, "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	live := sim.NewState(w.Manifest.Seed, worldmap.Generate(w.Manifest.Seed, worldmap.DefaultSize, worldmap.DefaultSize))
	if err := seedSurvivalWatches(w, st, live); err != nil {
		t.Fatal(err)
	}
	replayed := sim.NewState(w.Manifest.Seed, worldmap.Generate(w.Manifest.Seed, worldmap.DefaultSize, worldmap.DefaultSize))
	if err := st.ReplayEvents(0, func(e store.Event) error { return replayed.Apply(e) }); err != nil {
		t.Fatal(err)
	}
	for i := range live.GuardianOrders {
		lo, ro := live.GuardianOrders[i], replayed.GuardianOrders[i]
		if lo.PlacedSeq == 0 {
			t.Errorf("live watch %s has no PlacedSeq — the seeder applied with seq 0", lo.ID)
		}
		if lo.PlacedSeq != ro.PlacedSeq {
			t.Errorf("watch %s PlacedSeq live=%d replay=%d — boot state diverges from replay", lo.ID, lo.PlacedSeq, ro.PlacedSeq)
		}
	}
	if live.Hash() != replayed.Hash() {
		t.Fatal("seeded boot state != replayed state")
	}
}

// TestScenarioPassEndToEndThroughLoopAndObserver is SC-003's backbone (spec
// 054 T008): an armed first-night world run through the REAL sim loop — a
// real store, the loop's seq stamping, the watch placed through the
// InjectSocial door exactly as the angel would land it — reaches its pass at
// max speed; the notify fan-out's curriculum observer (the same consumer the
// daemon wires) updates the per-user unlocks record; and the loop's status
// snapshot carries the model-free scenario facts (FR-007).
func TestScenarioPassEndToEndThroughLoopAndObserver(t *testing.T) {
	t.Setenv("PROMPTWORLD_HOME", t.TempDir())

	dir := t.TempDir()
	w, err := world.Create(dir, "fnworld", sim.FirstNightExercise.Seed)
	if err != nil {
		t.Fatal(err)
	}
	m := worldmap.Generate(w.Manifest.Seed, worldmap.DefaultSize, worldmap.DefaultSize)
	state := sim.NewState(w.Manifest.Seed, m)
	// Deterministic survival frame (the sim package's genesis-mutation
	// idiom): one well-provisioned villager beside a long-lived fire — lit
	// tiles hide her from the scheduled gru.
	for i := range state.Agents {
		state.Agents[i].Dead = true
	}
	a := &state.Agents[0]
	a.Dead = false
	a.Needs = sim.Needs{Health: 1000, Food: 900, Rest: 900, Warmth: 900, Morale: 600}
	a.Inv = sim.Inventory{FoodRaw: 20}
	state.Structures = append(state.Structures, sim.Structure{Kind: "fire", X: a.X, Y: a.Y + 1, FuelUntil: 3 * 86400})
	if _, err := armScenario(&world.World{Manifest: world.Manifest{
		Name: w.Manifest.Name, Seed: w.Manifest.Seed,
		Scenario: &world.ScenarioConfig{Exercise: "first-night"},
	}}, state); err != nil {
		t.Fatal(err)
	}

	st, err := store.Open(filepath.Join(dir, "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	observer := curriculumObserver(w)
	loop := sim.NewLoop(state, m, st, observer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- loop.Run(ctx) }()

	// The watch, through the mind's injection door (dry-run validated) —
	// PlacedTick 0 is before nightfall, satisfying the rubric's direction
	// term.
	order := sim.GuardianOrder{
		ID: "ord-0-0", Origin: "player",
		Condition: "if the gru comes near", Action: "wake the village",
		EventTypes: []string{"gru.sighted"}, Agent: -1,
		PlacedTick: 0, ExpiresTick: 2 * 86400,
	}
	if err := loop.InjectSocial([]store.Event{{Type: "metatron.order_placed", Payload: mustJSON(t, order)}}); err != nil {
		t.Fatalf("inject watch: %v", err)
	}
	if _, err := loop.Do("set_speed", "max"); err != nil {
		t.Fatal(err)
	}

	// Run to the pass (dawn of day 2 = 86400 ticks) — bounded by wall time.
	deadline := time.After(120 * time.Second)
	var cs sim.Status
	for {
		cs, err = loop.Do("status", "")
		if err != nil {
			t.Fatal(err)
		}
		if cs.ScenarioOutcome == sim.OutcomePassed {
			break
		}
		if cs.Ended {
			t.Fatalf("run ended without a pass at tick %d", cs.Tick)
		}
		select {
		case <-deadline:
			t.Fatalf("no pass by wall deadline (tick %d, outcome %q)", cs.Tick, cs.ScenarioOutcome)
		default:
		}
		time.Sleep(20 * time.Millisecond)
	}
	if cs.ScenarioExercise != "first-night" {
		t.Errorf("status exercise = %q", cs.ScenarioExercise)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("loop.Run: %v", err)
	}

	// The observer recorded the unlock with a pointer at the pass event.
	u := worlds.LoadUnlocks()
	if !u.Earned("stage-2") {
		t.Fatal("pass did not reach the per-user unlocks record")
	}
	entry := u.Entries["stage-2"]
	if entry.World != "fnworld" || entry.Exercise != "first-night" {
		t.Errorf("unlock entry = %+v", entry)
	}
	if len(entry.Evidence) != 1 || entry.Evidence[0].Type != "curriculum.exercise_passed" || entry.Evidence[0].Seq == 0 {
		t.Errorf("unlock evidence = %+v, want a seq-bearing pointer at the pass", entry.Evidence)
	}

	// The durable log carries the same-batch pair: pass immediately before
	// its unlock, both at the dawn boundary.
	var passSeq, unlockSeq, passTick int64
	st.ReplayEvents(0, func(e store.Event) error {
		switch e.Type {
		case "curriculum.exercise_passed":
			passSeq, passTick = e.Seq, e.Tick
		case "curriculum.stage_unlocked":
			unlockSeq = e.Seq
		}
		return nil
	})
	if passSeq == 0 || unlockSeq != passSeq+1 {
		t.Errorf("pass/unlock seqs %d/%d, want contiguous with pass first", passSeq, unlockSeq)
	}
	if passTick != 86400 {
		t.Errorf("pass tick %d, want the dawn-of-day-2 boundary 86400", passTick)
	}
}
