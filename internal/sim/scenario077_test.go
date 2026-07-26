package sim

import (
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Spec 077 catalog + emitter-generalization tests (US1, T015-T020): the
// nine-exercise catalog's shape and content pins, the per-exercise rubric
// arms, the boundary vocabulary, evidence assembly through the sanctioned
// constructors, and the pass-emission fixtures (including the-law's
// completed emission and the stage-3→4 gate's first production grant).

// TestScenarioCatalogShape (FR-001, SC-001): nine exercises, 3/2/2/2 by
// stage, unique seeds 46101-46109.
func TestScenarioCatalogShape(t *testing.T) {
	if len(ScenarioExercises) != 9 {
		t.Fatalf("catalog has %d exercises, want 9", len(ScenarioExercises))
	}
	byStage := map[string]int{}
	seeds := map[uint64]string{}
	for _, def := range ScenarioExercises {
		byStage[def.Stage]++
		if prev, dup := seeds[def.Seed]; dup {
			t.Errorf("seed %d shared by %s and %s", def.Seed, prev, def.ID)
		}
		seeds[def.Seed] = def.ID
		if def.Seed < 46101 || def.Seed > 46109 {
			t.Errorf("%s: seed %d outside the pinned 46101-46109 band", def.ID, def.Seed)
		}
	}
	want := map[string]int{"stage-1": 3, "stage-2": 2, "stage-3": 2, "stage-4": 2}
	for stage, n := range want {
		if byStage[stage] != n {
			t.Errorf("stage %s has %d exercises, want %d", stage, byStage[stage], n)
		}
	}
}

// TestNoCatalogedExerciseReachesDefaultArm (FR-002, SC-001): every cataloged
// id gets a production EvaluateRubric arm — the default render-pending arm
// (whose terms echo the raw event type as their label) is reached by NONE.
func TestNoCatalogedExerciseReachesDefaultArm(t *testing.T) {
	for _, def := range ScenarioExercises {
		s := NewState(def.Seed, testMap(def.Seed))
		terms := EvaluateRubric(s, def, 0)
		if len(terms) < 2 {
			t.Errorf("%s: %d terms, want a real rubric", def.ID, len(terms))
		}
		for _, term := range terms {
			if term.Label == term.Event {
				t.Errorf("%s: term %q fell through to the default pending arm", def.ID, term.Label)
			}
		}
	}
}

// TestSchedulePositionsValidPerSeed (FR-008, the
// TestFirstNightSchedulePositionValid precedent generalized): every authored
// position is valid on its OWN exercise's seed's map — gru/stranger entries
// on passable, unprotected border tiles; blight centers over a genesis-
// blightable patch.
func TestSchedulePositionsValidPerSeed(t *testing.T) {
	for _, def := range ScenarioExercises {
		m := worldmap.Generate(def.Seed, worldmap.DefaultSize, worldmap.DefaultSize)
		s := NewState(def.Seed, m)
		for _, e := range def.Schedule {
			switch e.Kind {
			case IncidentGruEmerges, IncidentStrangerArrives:
				if !passable(m, s, e.X, e.Y) || gruProtected(s, e.X, e.Y) {
					t.Errorf("%s: %s position (%d,%d) not passable+unprotected on its own map",
						def.ID, e.Kind, e.X, e.Y)
				}
				if e.X != 0 && e.Y != 0 && e.X != m.W-1 && e.Y != m.H-1 {
					t.Errorf("%s: %s position (%d,%d) is not a border tile", def.ID, e.Kind, e.X, e.Y)
				}
			case IncidentForageBlight:
				if tiles := blightableTiles(m, s, e.X, e.Y, e.Radius); len(tiles) == 0 {
					t.Errorf("%s: blight center (%d,%d) r%d has no blightable forage at genesis",
						def.ID, e.X, e.Y, e.Radius)
				}
			}
		}
	}
}

// TestBoundaryDueVocabulary (FR-003, research R6): fixed = dawn of day N
// only; rolling = every dawn from day 2; never a non-dawn tick.
func TestBoundaryDueVocabulary(t *testing.T) {
	fixed := ExerciseDefinition{BoundaryDay: 3}
	rolling := ExerciseDefinition{}
	dawn := func(day int64) int64 { return clock.TickAt(day, 6, 0, 0) }
	cases := []struct {
		name string
		def  ExerciseDefinition
		tick int64
		want bool
	}{
		{"fixed at its dawn", fixed, dawn(3), true},
		{"fixed at an earlier dawn", fixed, dawn(2), false},
		{"fixed at a later dawn (missed forever)", fixed, dawn(4), false},
		{"fixed off-dawn same day", fixed, dawn(3) + 60, false},
		{"rolling at dawn 2", rolling, dawn(2), true},
		{"rolling at dawn 5", rolling, dawn(5), true},
		{"rolling at dawn 1 (genesis morning excluded)", rolling, dawn(1), false},
		{"rolling off-dawn", rolling, dawn(2) + 1, false},
	}
	for _, c := range cases {
		if got := boundaryDue(c.def, c.tick); got != c.want {
			t.Errorf("%s: boundaryDue = %v, want %v", c.name, got, c.want)
		}
	}
}

// skillsEvent builds a reducer-valid metatron.skills_observed fixture.
func skillsEvent(tick, seq int64) store.Event {
	return store.Event{Tick: tick, Seq: seq, Type: "metatron.skills_observed",
		Payload: mustPayload(SkillsObservedPayload{Fingerprint: "ab12cd34ef56", Names: []string{"10-watch.md"}})}
}

// playerOrderAt builds a player metatron.order_placed at tick with the seq.
func playerOrderAt(tick, seq int64) store.Event {
	e := watchOrderEvent(tick)
	e.Seq = seq
	return e
}

// --- rubric arm table tests (the TestTheLawRubricTable model) ---

func TestColdDawnRubricTable(t *testing.T) {
	m := testMap(ColdDawnExercise.Seed)
	dawn2 := clock.TickAt(2, 6, 0, 0)
	base := func() *State { return NewState(ColdDawnExercise.Seed, m) }

	// Exposure death flips only the freeze term's Met (and survival, via the
	// ledger); a gru death leaves "no villager freezes" standing.
	frozen := base()
	frozen.Deaths = append(frozen.Deaths, DeathRecord{Agent: 0, Tick: 60000, Cause: "exposure"})
	mauled := base()
	mauled.Deaths = append(mauled.Deaths, DeathRecord{Agent: 0, Tick: 60000, Cause: "gru"})
	watched := base()
	if err := watched.Apply(playerOrderAt(100, 7)); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		s    *State
		tick int64
		want [3]bool // survive, no-freeze, watch
	}{
		{"genesis", base(), 0, [3]bool{false, true, false}},
		{"watch placed, dawn reached", watched, dawn2, [3]bool{true, true, true}},
		{"exposure death", frozen, dawn2, [3]bool{false, false, false}},
		{"gru death: freeze term still stands", mauled, dawn2, [3]bool{false, true, false}},
	}
	for _, c := range cases {
		terms := EvaluateRubric(c.s, ColdDawnExercise, c.tick)
		if len(terms) != 3 {
			t.Fatalf("%s: %d terms, want 3", c.name, len(terms))
		}
		for i, term := range terms {
			if term.Met != c.want[i] {
				t.Errorf("%s: term %q Met=%v, want %v", c.name, term.Label, term.Met, c.want[i])
			}
		}
	}
}

func TestStrangerAtTheGateRubricTable(t *testing.T) {
	m := testMap(StrangerAtTheGateExercise.Seed)
	dawn2 := clock.TickAt(2, 6, 0, 0)
	base := func() *State { return NewState(StrangerAtTheGateExercise.Seed, m) }

	// "nothing is taken" is zero-wanted: Met at genesis, flipped by a take.
	robbed := base()
	if err := robbed.Apply(store.Event{Tick: 70000, Seq: 3, Type: "stranger.took",
		Payload: mustPayload(StrangerTookPayload{X: 5, Y: 5, Kind: "wood", N: 2})}); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		s    *State
		tick int64
		want [3]bool // survive, no-deaths, nothing-taken
	}{
		{"genesis: zero-wanted term already Met", base(), 0, [3]bool{false, true, true}},
		{"clean dawn", base(), dawn2, [3]bool{true, true, true}},
		{"a take flips the term", robbed, dawn2, [3]bool{true, true, false}},
	}
	for _, c := range cases {
		terms := EvaluateRubric(c.s, StrangerAtTheGateExercise, c.tick)
		for i, term := range terms {
			if term.Met != c.want[i] {
				t.Errorf("%s: term %q Met=%v, want %v", c.name, term.Label, term.Met, c.want[i])
			}
		}
	}
	if terms := EvaluateRubric(robbed, StrangerAtTheGateExercise, dawn2); terms[2].Count != 1 {
		t.Errorf("take count = %d, want the ledger's 1", terms[2].Count)
	}
}

func TestBlightedLarderRubricTable(t *testing.T) {
	m := testMap(BlightedLarderExercise.Seed)
	base := func() *State { return NewState(BlightedLarderExercise.Seed, m) }

	banked := base()
	if err := banked.Apply(theLawCharterEvent(10, "bbbb33334444", false)); err != nil {
		t.Fatal(err)
	}
	// A chest holding enough food, plus a ground batch — storedFoodTotal
	// sums both store shapes.
	banked.Structures = append(banked.Structures, Structure{Kind: "chest", X: 5, Y: 5, Owner: 0,
		Store: &Inventory{FoodRaw: blightedLarderFoodFloor - 2}})
	banked.pileFor(6, 5).addFood("meals", 2, 1<<40)

	starved := base()
	starved.Deaths = append(starved.Deaths, DeathRecord{Agent: 0, Tick: 60000, Cause: "starvation"})

	cases := []struct {
		name string
		s    *State
		want [3]bool // charter, no-starvation, larder
	}{
		{"genesis", base(), [3]bool{false, true, false}},
		{"charter + banked larder", banked, [3]bool{true, true, true}},
		{"a starvation death", starved, [3]bool{false, false, false}},
	}
	for _, c := range cases {
		terms := EvaluateRubric(c.s, BlightedLarderExercise, 0)
		for i, term := range terms {
			if term.Met != c.want[i] {
				t.Errorf("%s: term %q Met=%v, want %v", c.name, term.Label, term.Met, c.want[i])
			}
		}
	}
	if terms := EvaluateRubric(banked, BlightedLarderExercise, 0); terms[2].Count != blightedLarderFoodFloor {
		t.Errorf("larder count = %d, want storedFoodTotal %d", terms[2].Count, blightedLarderFoodFloor)
	}
}

func TestToolsmithRubricTable(t *testing.T) {
	m := testMap(ToolsmithExercise.Seed)
	base := func() *State { return NewState(ToolsmithExercise.Seed, m) }

	observed := base()
	if err := observed.Apply(skillsEvent(100, 5)); err != nil {
		t.Fatal(err)
	}
	acted := base()
	if err := acted.Apply(skillsEvent(100, 5)); err != nil {
		t.Fatal(err)
	}
	if err := acted.Apply(playerOrderAt(200, 6)); err != nil {
		t.Fatal(err)
	}
	// An order placed BEFORE the observation never counts as acting under it.
	stale := base()
	if err := stale.Apply(playerOrderAt(50, 4)); err != nil {
		t.Fatal(err)
	}
	if err := stale.Apply(skillsEvent(100, 5)); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		s    *State
		want [3]bool // skills, acts-under-it, no-deaths
	}{
		{"genesis", base(), [3]bool{false, false, true}},
		{"skills observed, no act yet", observed, [3]bool{true, false, true}},
		{"order predates the observation", stale, [3]bool{true, false, true}},
		{"observed then acted", acted, [3]bool{true, true, true}},
	}
	for _, c := range cases {
		terms := EvaluateRubric(c.s, ToolsmithExercise, c.s.Tick)
		for i, term := range terms {
			if term.Met != c.want[i] {
				t.Errorf("%s: term %q Met=%v, want %v", c.name, term.Label, term.Met, c.want[i])
			}
		}
	}
}

func TestFogWatchAndCapstoneRubricTables(t *testing.T) {
	dawn3 := clock.TickAt(3, 6, 0, 0)
	m := testMap(FogWatchExercise.Seed)
	ready := NewState(FogWatchExercise.Seed, m)
	if err := ready.Apply(skillsEvent(100, 5)); err != nil {
		t.Fatal(err)
	}
	terms := EvaluateRubric(ready, FogWatchExercise, dawn3)
	for i, want := range []bool{true, true, true} {
		if terms[i].Met != want {
			t.Errorf("fog-watch term %q Met=%v, want %v", terms[i].Label, terms[i].Met, want)
		}
	}

	// long-winter: survive to dawn 4 with an intact ledger and larder.
	m8 := testMap(LongWinterExercise.Seed)
	lw := NewState(LongWinterExercise.Seed, m8)
	dawn4 := clock.TickAt(4, 6, 0, 0)
	lwTerms := EvaluateRubric(lw, LongWinterExercise, dawn4)
	for i, want := range []bool{true, true, true} {
		if lwTerms[i].Met != want {
			t.Errorf("long-winter term %q Met=%v, want %v", lwTerms[i].Label, lwTerms[i].Met, want)
		}
	}

	// stewards-charge: the whole instruction ladder in force.
	m9 := testMap(StewardsChargeExercise.Seed)
	sc := NewState(StewardsChargeExercise.Seed, m9)
	for _, e := range []store.Event{theLawNormEvent(20), theLawCharterEvent(30, "bbbb33334444", false), skillsEvent(40, 5)} {
		if err := sc.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	scTerms := EvaluateRubric(sc, StewardsChargeExercise, 0)
	for i, want := range []bool{true, true, true, true} {
		if scTerms[i].Met != want {
			t.Errorf("stewards-charge term %q Met=%v, want %v", scTerms[i].Label, scTerms[i].Met, want)
		}
	}
}

// --- pass emission fixtures (T015/T020) ---

// isolateForBoundary parks every villager dead except one provisioned
// survivor beside a long fire — a deterministic frame whose dawn heartbeat
// kills no one.
func isolateForBoundary(s *State) {
	isolateAgents(s)
	a := &s.Agents[0]
	a.Dead = false
	a.Needs = Needs{Health: 1000, Food: 900, Rest: 900, Warmth: 900, Morale: 600}
	a.Inv = Inventory{FoodRaw: 10}
	s.Structures = append(s.Structures, Structure{Kind: "fire", X: a.X, Y: a.Y + 1, FuelUntil: 1 << 40})
}

// TestTheLawPassEmissionWithCharterEvidence is US1 AS-3 (SC-001): a the-law
// world where a norm is adopted while a player-authored charter is in force
// emits, at the next dawn boundary, curriculum.exercise_passed carrying a
// metatron.charter_observed evidence ref with Custom true (re-located via
// the persisted Seq/Tick) — and curriculum.stage_unlocked{stage-3} in the
// SAME batch, pass first. The spec-072 FR-009 guard is retired.
func TestTheLawPassEmissionWithCharterEvidence(t *testing.T) {
	m := testMap(TheLawExercise.Seed)
	s := NewState(TheLawExercise.Seed, m)
	if err := s.ArmScenario(TheLawExercise); err != nil {
		t.Fatal(err)
	}
	isolateForBoundary(s)
	charter := theLawCharterEvent(30_000, "bbbb33334444", false)
	charter.Seq = 41
	for _, e := range []store.Event{charter, theLawNormEvent(40_000)} {
		if err := s.Apply(e); err != nil {
			t.Fatal(err)
		}
	}

	dawn2 := clock.TickAt(2, 6, 0, 0)
	s.Tick = dawn2 - 1
	s.Night = true
	batch := stepEvents(s, m, dawn2)
	s.Tick = dawn2

	passIdx, unlockIdx := -1, -1
	for i, e := range batch {
		switch e.Type {
		case "curriculum.exercise_passed":
			passIdx = i
		case "curriculum.stage_unlocked":
			unlockIdx = i
		}
	}
	if passIdx < 0 || unlockIdx < 0 {
		t.Fatalf("boundary batch missing pass/unlock: %v", eventTypes(batch))
	}
	if !(passIdx < unlockIdx) {
		t.Fatal("unlock preceded its pass (daemon observer contract)")
	}
	var pass ExercisePassedPayload
	if err := json.Unmarshal(batch[passIdx].Payload, &pass); err != nil {
		t.Fatal(err)
	}
	if pass.Exercise != "the-law" || pass.Stage != "stage-2" {
		t.Errorf("pass payload = %+v", pass)
	}
	if len(pass.Evidence) != 1 || pass.Evidence[0].Type != "metatron.charter_observed" ||
		pass.Evidence[0].Seq != 41 || pass.Evidence[0].Tick != 30_000 || !pass.Evidence[0].Custom {
		t.Errorf("evidence = %+v, want the charter observation at seq 41 tick 30000, Custom true", pass.Evidence)
	}
	var unlock StageUnlockedPayload
	if err := json.Unmarshal(batch[unlockIdx].Payload, &unlock); err != nil {
		t.Fatal(err)
	}
	if unlock.Stage != "stage-3" || unlock.Exercise != "the-law" {
		t.Errorf("unlock payload = %+v", unlock)
	}

	// Latch: apply the batch; the boundary never re-fires (rolling included).
	for _, e := range batch {
		if err := s.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	if evs := scenarioRubricEvents(s, clock.TickAt(3, 6, 0, 0), nil); len(evs) != 0 {
		t.Errorf("rolling boundary re-emitted %d events after the pass", len(evs))
	}
}

// TestTheLawRollingBoundaryWaitsForSatisfaction: an unmet rubric at dawn 2
// emits nothing; once satisfied, the pass lands at the NEXT dawn — the first
// satisfying one — never retroactively.
func TestTheLawRollingBoundaryWaitsForSatisfaction(t *testing.T) {
	m := testMap(TheLawExercise.Seed)
	s := NewState(TheLawExercise.Seed, m)
	if err := s.ArmScenario(TheLawExercise); err != nil {
		t.Fatal(err)
	}
	isolateForBoundary(s)
	if evs := scenarioRubricEvents(s, clock.TickAt(2, 6, 0, 0), nil); len(evs) != 0 {
		t.Fatalf("unmet rubric emitted %d events at dawn 2", len(evs))
	}
	charter := theLawCharterEvent(90_000, "bbbb33334444", false)
	charter.Seq = 41
	for _, e := range []store.Event{charter, theLawNormEvent(95_000)} {
		if err := s.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	evs := scenarioRubricEvents(s, clock.TickAt(3, 6, 0, 0), nil)
	if len(evs) != 2 {
		t.Fatalf("satisfying dawn 3 emitted %d events, want pass+unlock", len(evs))
	}
}

// TestTheLawPre077SnapshotPassWaits is the pre-077 degradation edge: the
// rubric holds but the observation's coordinates are NOT on state (a
// snapshot whose charter arm predates the Seq/Tick stamp) — the pass WAITS;
// the next charter observation stamps the coordinates and the pass lands at
// the next dawn (honest, self-healing).
func TestTheLawPre077SnapshotPassWaits(t *testing.T) {
	m := testMap(TheLawExercise.Seed)
	s := NewState(TheLawExercise.Seed, m)
	if err := s.ArmScenario(TheLawExercise); err != nil {
		t.Fatal(err)
	}
	isolateForBoundary(s)
	if err := s.Apply(theLawNormEvent(40_000)); err != nil {
		t.Fatal(err)
	}
	// Simulate the pre-077 snapshot: authorship flags present, coordinates
	// absent (loaded state, not a reducer product).
	s.CharterFingerprint = "bbbb33334444"
	s.CharterCustom = true

	if evs := scenarioRubricEvents(s, clock.TickAt(2, 6, 0, 0), nil); len(evs) != 0 {
		t.Fatalf("pass emitted without state-derivable evidence coordinates (%d events)", len(evs))
	}

	// The next observation self-heals the stamp; the pass lands at the
	// next dawn.
	charter := theLawCharterEvent(100_000, "cccc55556666", false)
	charter.Seq = 77
	if err := s.Apply(charter); err != nil {
		t.Fatal(err)
	}
	evs := scenarioRubricEvents(s, clock.TickAt(3, 6, 0, 0), nil)
	if len(evs) != 2 {
		t.Fatalf("self-healed dawn emitted %d events, want pass+unlock", len(evs))
	}
	var pass ExercisePassedPayload
	if err := json.Unmarshal(evs[0].Payload, &pass); err != nil {
		t.Fatal(err)
	}
	if len(pass.Evidence) != 1 || pass.Evidence[0].Seq != 77 {
		t.Errorf("evidence = %+v, want the fresh observation's coordinates", pass.Evidence)
	}
}

// TestToolsmithPassUnlocksStageFour is US1 AS-4 (SC-001): a stage-3 pass
// whose evidence includes the Custom:true skills entry opens the stage-3→4
// gate — the first production satisfaction of EvaluateUnlock's stage-3
// conjunct.
func TestToolsmithPassUnlocksStageFour(t *testing.T) {
	m := testMap(ToolsmithExercise.Seed)
	s := NewState(ToolsmithExercise.Seed, m)
	if err := s.ArmScenario(ToolsmithExercise); err != nil {
		t.Fatal(err)
	}
	isolateForBoundary(s)
	for _, e := range []store.Event{skillsEvent(30_000, 41), playerOrderAt(40_000, 42)} {
		if err := s.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	dawn2 := clock.TickAt(2, 6, 0, 0)
	s.Tick = dawn2 - 1
	s.Night = true
	batch := stepEvents(s, m, dawn2)

	var pass ExercisePassedPayload
	var unlock StageUnlockedPayload
	havePass, haveUnlock := false, false
	for _, e := range batch {
		switch e.Type {
		case "curriculum.exercise_passed":
			havePass = true
			if err := json.Unmarshal(e.Payload, &pass); err != nil {
				t.Fatal(err)
			}
		case "curriculum.stage_unlocked":
			haveUnlock = true
			if err := json.Unmarshal(e.Payload, &unlock); err != nil {
				t.Fatal(err)
			}
		}
	}
	if !havePass || !haveUnlock {
		t.Fatalf("batch missing pass/unlock: %v", eventTypes(batch))
	}
	custom := false
	for _, ev := range pass.Evidence {
		if ev.Type == "metatron.skills_observed" && ev.Custom && ev.Seq == 41 {
			custom = true
		}
	}
	if !custom {
		t.Errorf("evidence = %+v, want a Custom:true skills entry at seq 41", pass.Evidence)
	}
	if unlock.Stage != "stage-4" || unlock.Exercise != "toolsmith" {
		t.Errorf("unlock = %+v, want stage-4 by toolsmith", unlock)
	}
}

// TestStageFourPassGraduatesWithoutUnlock is US1 AS-5: a stage-4 pass is
// recorded, but no curriculum.stage_unlocked follows — graduation, the
// existing nextLadderStage posture.
func TestStageFourPassGraduatesWithoutUnlock(t *testing.T) {
	m := testMap(StewardsChargeExercise.Seed)
	s := NewState(StewardsChargeExercise.Seed, m)
	if err := s.ArmScenario(StewardsChargeExercise); err != nil {
		t.Fatal(err)
	}
	isolateForBoundary(s)
	charter := theLawCharterEvent(20_000, "bbbb33334444", false)
	charter.Seq = 40
	for _, e := range []store.Event{theLawNormEvent(10_000), charter, skillsEvent(30_000, 41)} {
		if err := s.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	dawn2 := clock.TickAt(2, 6, 0, 0)
	s.Tick = dawn2 - 1
	s.Night = true
	batch := stepEvents(s, m, dawn2)
	pass, unlock := false, false
	for _, e := range batch {
		switch e.Type {
		case "curriculum.exercise_passed":
			pass = true
		case "curriculum.stage_unlocked":
			unlock = true
		}
	}
	if !pass {
		t.Fatalf("no pass in the boundary batch: %v", eventTypes(batch))
	}
	if unlock {
		t.Error("a stage-4 pass unlocked something — graduation must unlock nothing")
	}
}

// TestFixedBoundaryMissEmitsNothingForever is US1 AS-6 + the edge case: a
// fixed-boundary exercise whose terms are unmet at its boundary dawn emits
// nothing — then and forever; the outcome stays in_progress on a live world.
func TestFixedBoundaryMissEmitsNothingForever(t *testing.T) {
	m := testMap(ColdDawnExercise.Seed)
	s := NewState(ColdDawnExercise.Seed, m)
	if err := s.ArmScenario(ColdDawnExercise); err != nil {
		t.Fatal(err)
	}
	isolateForBoundary(s)
	// No watch placed: the rubric is unmet at dawn 2.
	if evs := scenarioRubricEvents(s, clock.TickAt(2, 6, 0, 0), nil); len(evs) != 0 {
		t.Fatalf("unmet fixed boundary emitted %d events", len(evs))
	}
	// A watch placed AFTER the boundary changes nothing: day 3+ dawns are
	// not this exercise's boundary.
	if err := s.Apply(playerOrderAt(100_000, 50)); err != nil {
		t.Fatal(err)
	}
	for day := int64(3); day <= 6; day++ {
		if evs := scenarioRubricEvents(s, clock.TickAt(day, 6, 0, 0), nil); len(evs) != 0 {
			t.Fatalf("missed fixed boundary emitted %d events at dawn %d", len(evs), day)
		}
	}
	if got := ExerciseOutcome(s, ColdDawnExercise.ID); got != OutcomeInProgress {
		t.Errorf("outcome = %q, want in_progress (failure is never an event)", got)
	}
}

// TestAllDeadDawnSuppressedForEveryExercise generalizes the photo-finish
// edge: a death landing in the boundary batch itself blocks the pass for a
// NON-first-night exercise too (the batch-death scan is the shared guard).
func TestAllDeadDawnSuppressedForEveryExercise(t *testing.T) {
	m := testMap(TheLawExercise.Seed)
	s := NewState(TheLawExercise.Seed, m)
	if err := s.ArmScenario(TheLawExercise); err != nil {
		t.Fatal(err)
	}
	charter := theLawCharterEvent(30_000, "bbbb33334444", false)
	charter.Seq = 41
	for _, e := range []store.Event{charter, theLawNormEvent(40_000)} {
		if err := s.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	// One villager dying ON the dawn heartbeat (the
	// TestFirstNightDawnDeathIsFailNotPass frame).
	isolateAgents(s)
	a := &s.Agents[0]
	a.Dead = false
	a.Needs = Needs{Health: healthLoss, Food: 0, Rest: 500, Warmth: 500, Morale: 500}

	dawn2 := clock.TickAt(2, 6, 0, 0)
	s.Tick = dawn2 - 1
	s.Night = true
	batch := stepEvents(s, m, dawn2)
	died, passed := false, false
	for _, e := range batch {
		switch e.Type {
		case "agent.died":
			died = true
		case "curriculum.exercise_passed":
			passed = true
		}
	}
	if !died {
		t.Fatalf("edge setup broken: no dawn death (batch %v)", eventTypes(batch))
	}
	if passed {
		t.Fatal("a dawn-batch death produced a pass — the shared batch-death guard failed")
	}
}

// TestColdDawnGenesisPassAndReplay drives cold-dawn from genesis through its
// authored snap to the dawn-2 pass — the full driveTicks path under the new
// incident — and proves genesis replay equivalence with no scenario armed.
func TestColdDawnGenesisPassAndReplay(t *testing.T) {
	m := worldmap.Generate(ColdDawnExercise.Seed, worldmap.DefaultSize, worldmap.DefaultSize)
	dawn2 := clock.TickAt(2, 6, 0, 0)
	genesis := func() *State {
		s := NewState(ColdDawnExercise.Seed, m)
		isolateForBoundary(s)
		return s
	}
	live := genesis()
	if err := live.ArmScenario(ColdDawnExercise); err != nil {
		t.Fatal(err)
	}
	commands := map[int64][]store.Event{30_000: {watchOrderEvent(30_000)}}
	log := driveTicksSeq(t, live, m, dawn2+3600, commands)

	var snapTick int64
	passes := 0
	for _, e := range log {
		switch e.Type {
		case "sim.cold_snap":
			snapTick = e.Tick
		case "curriculum.exercise_passed":
			passes++
			var p ExercisePassedPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				t.Fatal(err)
			}
			if p.Exercise != "cold-dawn" || len(p.Evidence) != 1 || p.Evidence[0].Type != "metatron.order_placed" {
				t.Errorf("pass = %+v, want cold-dawn with the watch evidence", p)
			}
		}
	}
	if snapTick != clock.TickAt(1, 22, 0, 0) {
		t.Errorf("cold snap fired at %d, want the authored 22:00", snapTick)
	}
	if passes != 1 {
		t.Fatalf("%d passes, want exactly 1", passes)
	}
	if got := ExerciseOutcome(live, "cold-dawn"); got != OutcomePassed {
		t.Errorf("outcome = %q, want passed", got)
	}

	replayed := genesis()
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replayed.Tick = e.Tick
	}
	replayed.Tick = live.Tick
	if live.Hash() != replayed.Hash() {
		t.Fatal("cold-dawn replay diverged from live run")
	}
}

// TestStrangerAtTheGateGenesisPass drives stranger-at-the-gate from genesis:
// the authored stranger prowls a storeless night, takes nothing, and the
// dawn-2 pass lands (zero-wanted term honest).
func TestStrangerAtTheGateGenesisPass(t *testing.T) {
	m := worldmap.Generate(StrangerAtTheGateExercise.Seed, worldmap.DefaultSize, worldmap.DefaultSize)
	dawn2 := clock.TickAt(2, 6, 0, 0)
	s := NewState(StrangerAtTheGateExercise.Seed, m)
	isolateForBoundary(s)
	if err := s.ArmScenario(StrangerAtTheGateExercise); err != nil {
		t.Fatal(err)
	}
	log := driveTicksSeq(t, s, m, dawn2+3600, nil)

	arrived, passes := false, 0
	for _, e := range log {
		switch e.Type {
		case "stranger.arrived":
			arrived = true
		case "stranger.took":
			t.Fatal("the stranger took from a storeless village")
		case "curriculum.exercise_passed":
			passes++
		}
	}
	if !arrived {
		t.Fatal("the authored stranger never arrived")
	}
	if passes != 1 {
		t.Fatalf("%d passes, want exactly 1", passes)
	}
	if len(s.StrangerTakes) != 0 {
		t.Errorf("ledger = %v, want empty", s.StrangerTakes)
	}
}
