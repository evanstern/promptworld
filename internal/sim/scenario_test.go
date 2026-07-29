package sim

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Scenario machinery tests (spec 054): the determinism contract
// (contracts/scenario-machinery.md §1) is the soul of this feature — twins,
// preemption, rubric tables, same-batch pass+unlock ordering, and replay
// equivalence all live here.

// firstNightMap is the first-night exercise's own terrain: its authored seed
// at the default dimensions, exactly what `promptworld new --scenario
// first-night` generates.
func firstNightMap() *worldmap.Map {
	return worldmap.Generate(FirstNightExercise.Seed, worldmap.DefaultSize, worldmap.DefaultSize)
}

// armedFirstNight is an armed first-night genesis on the exercise's own
// seed/map — the state a scenario world boots with.
func armedFirstNight(t *testing.T) (*State, *worldmap.Map) {
	t.Helper()
	m := firstNightMap()
	s := NewState(FirstNightExercise.Seed, m)
	if err := s.ArmScenario(FirstNightExercise); err != nil {
		t.Fatalf("arm first-night: %v", err)
	}
	return s, m
}

// driveTicksSeq is driveTicks with the loop's seq stamping (stampSeqs'
// last+i+1 arithmetic): scenario tests need real seqs because the
// metatron.order_placed reducer arm stamps PlacedSeq from the event
// envelope, and the pass payload's evidence re-locates the order by it.
func driveTicksSeq(t *testing.T, s *State, m *worldmap.Map, ticks int64, commands map[int64][]store.Event) []store.Event {
	t.Helper()
	var log []store.Event
	var lastSeq int64
	apply := func(evs []store.Event) {
		for i := range evs {
			lastSeq++
			evs[i].Seq = lastSeq
			if err := s.Apply(evs[i]); err != nil {
				t.Fatalf("apply %s at tick %d: %v", evs[i].Type, s.Tick, err)
			}
			log = append(log, evs[i])
		}
	}
	for s.Tick < ticks {
		apply(commands[s.Tick])
		next := s.Tick + 1
		evs := stepEvents(s, m, next)
		s.Tick = next
		apply(evs)
	}
	return log
}

// watchOrderEvent builds a reducer-valid player metatron.order_placed at
// tick — the watch the first-night rubric's direction term reads.
func watchOrderEvent(tick int64) store.Event {
	return store.Event{Tick: tick, Type: "guardian.order_placed", Payload: mustPayload(GuardianOrder{
		ID: "ord-test-watch", Origin: "player",
		Condition: "if the gru comes near the fire", Action: "wake the village",
		EventTypes: []string{"gru.sighted"}, Agent: -1,
		PlacedTick: tick, ExpiresTick: tick + 2*ticksPerGameDay,
	})}
}

// TestScenarioSchedulesCompile pins every cataloged exercise's authored
// schedule: arming must succeed (kinds in the closed enum, times parseable)
// — the boot-time compile error in ArmScenario is a can't-happen belt only
// because this test holds.
func TestScenarioSchedulesCompile(t *testing.T) {
	for _, def := range ScenarioExercises {
		s := NewState(def.Seed, testMap(def.Seed))
		if err := s.ArmScenario(def); err != nil {
			t.Errorf("%s: schedule does not compile: %v", def.ID, err)
		}
		if got := s.ScenarioExerciseID(); got != def.ID {
			t.Errorf("%s: armed id = %q", def.ID, got)
		}
	}
}

// TestFirstNightSchedulePositionValid pins the authored gru position against
// the exercise's own map: a passable, unprotected border tile — the same
// tile class the random emergence path draws from, so the scheduled event is
// indistinguishable in kind (contract §2).
func TestFirstNightSchedulePositionValid(t *testing.T) {
	m := firstNightMap()
	s := NewState(FirstNightExercise.Seed, m)
	for _, e := range FirstNightExercise.Schedule {
		if e.Kind != IncidentGruEmerges {
			continue
		}
		if !passable(m, s, e.X, e.Y) {
			t.Errorf("authored position (%d,%d) is not passable on the first-night map", e.X, e.Y)
		}
		if gruProtected(s, e.X, e.Y) {
			t.Errorf("authored position (%d,%d) is protected at genesis", e.X, e.Y)
		}
		if e.X != 0 && e.Y != 0 && e.X != m.W-1 && e.Y != m.H-1 {
			t.Errorf("authored position (%d,%d) is not a border tile", e.X, e.Y)
		}
	}
}

// TestScheduledEmissionsDeterministicTwins is SC-001's genesis half: two
// armed genesis runs of the same (seed, definition) produce byte-identical
// event sequences, with the scheduled gru.emerged at the authored tick and
// position — and only one spawn mechanism that night (the roll preempted).
func TestScheduledEmissionsDeterministicTwins(t *testing.T) {
	const ticks = 60_000 // past nightfall of night one (57600)
	a, m := armedFirstNight(t)
	b, _ := armedFirstNight(t)
	logA := driveTicks(t, a, m, ticks, nil)
	logB := driveTicks(t, b, m, ticks, nil)

	if !bytes.Equal(canonicalLog(t, logA), canonicalLog(t, logB)) {
		t.Fatal("armed twins diverged: scheduled emissions are not deterministic")
	}
	if a.Hash() != b.Hash() {
		t.Fatalf("armed twin state hashes diverged: %s vs %s", a.Hash(), b.Hash())
	}

	authoredTick := clock.TickAt(1, 22, 0, 0)
	emerged := 0
	for _, e := range logA {
		if e.Type != "gru.emerged" {
			continue
		}
		emerged++
		var p GruEmergedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			t.Fatal(err)
		}
		if e.Tick != authoredTick || p.X != 44 || p.Y != 0 {
			t.Errorf("gru.emerged at tick %d (%d,%d), want authored tick %d (44,0)",
				e.Tick, p.X, p.Y, authoredTick)
		}
	}
	if emerged != 1 {
		t.Fatalf("night one produced %d gru.emerged events, want exactly 1 (schedule preempts the roll)", emerged)
	}
}

// TestScheduledNightPreemptsRandomRoll is research R3's direct assertion: on
// a night with a scheduled emergence whose authored moment is LATER than
// nightfall, the nightfall roll is skipped even when the dice would have
// fired — one mechanism per night — and the incident then lands at its own
// authored tick.
func TestScheduledNightPreemptsRandomRoll(t *testing.T) {
	// Find a seed whose night-1 roll fires, so the suppression is observable.
	var seed uint64
	for cand := uint64(1); cand < 200; cand++ {
		if rngAt(cand, "gru-emerge", 1, 0).Uint64N(1000) < defaultGruEmergePerMille {
			seed = cand
			break
		}
	}
	if seed == 0 {
		t.Fatal("no seed under 200 rolls a night-1 emergence")
	}
	m := testMap(seed)

	// A border tile the scheduled emergence can honestly use on this map.
	s0 := NewState(seed, m)
	x, y := -1, 0
	for cx := 0; cx < m.W; cx++ {
		if passable(m, s0, cx, 0) && !gruProtected(s0, cx, 0) {
			x = cx
			break
		}
	}
	if x < 0 {
		t.Fatal("no passable border tile on test map")
	}

	def := ExerciseDefinition{ID: "first-night", Stage: "stage-1", Seed: seed,
		Schedule: []IncidentScheduleEntry{{Kind: IncidentGruEmerges, Day: 1, Time: "23:30", X: x, Y: y}}}
	s := NewState(seed, m)
	if err := s.ArmScenario(def); err != nil {
		t.Fatal(err)
	}
	log := driveTicks(t, s, m, clock.TickAt(2, 6, 0, 0), nil)

	authoredTick := clock.TickAt(1, 23, 30, 0)
	var emergedTicks []int64
	for _, e := range log {
		if e.Type == "gru.emerged" {
			emergedTicks = append(emergedTicks, e.Tick)
		}
	}
	if len(emergedTicks) != 1 || emergedTicks[0] != authoredTick {
		t.Fatalf("emergences at %v, want exactly one at the authored tick %d (roll suppressed)",
			emergedTicks, authoredTick)
	}
}

// TestScheduledIncidentPreconditionSkips is US2 AS-2: an incident whose
// precondition fails at its tick is skipped silently — recorded nowhere,
// retried never. A gru already abroad blocks the emergence; the closed night
// window means it is never revived retroactively.
func TestScheduledIncidentPreconditionSkips(t *testing.T) {
	m := firstNightMap()
	s := NewState(FirstNightExercise.Seed, m)
	if err := s.ArmScenario(FirstNightExercise); err != nil {
		t.Fatal(err)
	}
	// A gru is already abroad when the authored tick arrives (staged by
	// hand, the gruTestState idiom).
	authoredTick := clock.TickAt(1, 22, 0, 0)
	s.Tick = authoredTick - 1
	s.Night = true
	s.Gru = &Gru{X: 10, Y: 20}
	for i := range s.Agents {
		s.Agents[i].Dead = true // no witnesses, no attacks — pure schedule frame
	}
	for i := int64(0); i < 300; i++ {
		next := s.Tick + 1
		for _, e := range stepEvents(s, m, next) {
			if e.Type == "gru.emerged" {
				t.Fatalf("scheduled emergence landed at tick %d despite a gru already abroad", next)
			}
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		s.Tick = next
	}
}

// TestScheduledIncidentWindowCloses pins the state-latch arithmetic
// (research R2): once the authored night's dawn passes, the entry is no
// longer due — fires late within its night, never twice, never after it.
func TestScheduledIncidentWindowCloses(t *testing.T) {
	s, m := armedFirstNight(t)
	// Jump the clock past the authored night entirely (the time-snap-past
	// edge): day 2, 07:00 — window [22:00 day1, 06:00 day2) has closed.
	s.Tick = clock.TickAt(2, 7, 0, 0)
	s.Night = false
	for i := range s.Agents {
		s.Agents[i].Dead = true
	}
	for i := int64(0); i < 200; i++ {
		next := s.Tick + 1
		for _, e := range stepEvents(s, m, next) {
			if e.Type == "gru.emerged" {
				t.Fatalf("lapsed incident fired at tick %d, outside its night window", next)
			}
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		s.Tick = next
	}
}

// TestFirstNightRubricTable is the rubric evaluator's table test (research
// R4): per-term satisfaction over crafted state facts.
func TestFirstNightRubricTable(t *testing.T) {
	m := firstNightMap()
	nightfall := clock.TickAt(1, 22, 0, 0)
	dawn2 := clock.TickAt(2, 6, 0, 0)

	base := func() *State { return NewState(FirstNightExercise.Seed, m) }
	withWatch := func(s *State, placed int64, origin string) *State {
		e := watchOrderEvent(placed)
		var o GuardianOrder
		if err := json.Unmarshal(e.Payload, &o); err != nil {
			t.Fatal(err)
		}
		o.Origin = origin
		e.Payload = mustPayload(o)
		e.Seq = 7
		if err := s.Apply(e); err != nil {
			t.Fatal(err)
		}
		return s
	}

	cases := []struct {
		name string
		s    *State
		tick int64
		want [3]bool // survive, no-deaths, watch
	}{
		{"genesis", base(), 0, [3]bool{false, true, false}},
		{"watch before nightfall", withWatch(base(), nightfall-100, "player"), nightfall, [3]bool{false, true, true}},
		{"watch after nightfall", withWatch(base(), nightfall+100, "player"), dawn2, [3]bool{true, true, false}},
		{"system order never counts", withWatch(base(), 100, "system"), dawn2, [3]bool{true, true, false}},
		{"dawn of day 2, all terms", withWatch(base(), 100, "player"), dawn2, [3]bool{true, true, true}},
	}
	// A death flips both the survival and zero-deaths terms.
	dead := withWatch(base(), 100, "player")
	dead.Deaths = append(dead.Deaths, DeathRecord{Agent: 0, Tick: 60000, Cause: "gru"})
	cases = append(cases, struct {
		name string
		s    *State
		tick int64
		want [3]bool
	}{"a death before dawn", dead, dawn2, [3]bool{false, false, true}})

	for _, c := range cases {
		terms := EvaluateRubric(c.s, FirstNightExercise, c.tick)
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

// TestFirstNightPassSameBatchOrdering is US1 AS-1 + the daemon observer's
// same-batch contract: at dawn of day 2 with the rubric satisfied, ONE batch
// carries curriculum.exercise_passed then curriculum.stage_unlocked (pass
// first), the evidence re-locates the watch's placement event, the reducer
// latches both, and the boundary never re-fires.
func TestFirstNightPassSameBatchOrdering(t *testing.T) {
	s, m := armedFirstNight(t)
	dawn2 := clock.TickAt(2, 6, 0, 0)

	// A player watch placed before nightfall, with the loop's seq stamping.
	watch := watchOrderEvent(30_000)
	watch.Seq = 41
	if err := s.Apply(watch); err != nil {
		t.Fatal(err)
	}

	// Step the exact boundary tick from the tick before it.
	s.Tick = dawn2 - 1
	s.Night = true
	batch := stepEvents(s, m, dawn2)
	s.Tick = dawn2

	passIdx, unlockIdx, dayIdx := -1, -1, -1
	for i, e := range batch {
		switch e.Type {
		case "curriculum.exercise_passed":
			passIdx = i
		case "curriculum.stage_unlocked":
			unlockIdx = i
		case "sim.day_started":
			dayIdx = i
		}
	}
	if dayIdx < 0 || passIdx < 0 || unlockIdx < 0 {
		t.Fatalf("boundary batch missing events: day=%d pass=%d unlock=%d (batch %d events)",
			dayIdx, passIdx, unlockIdx, len(batch))
	}
	if !(passIdx < unlockIdx) {
		t.Fatalf("pass at %d must precede unlock at %d (daemon observer contract)", passIdx, unlockIdx)
	}

	var pass ExercisePassedPayload
	if err := json.Unmarshal(batch[passIdx].Payload, &pass); err != nil {
		t.Fatal(err)
	}
	if pass.Exercise != "first-night" || pass.Stage != "stage-1" || pass.Tick != dawn2 {
		t.Errorf("pass payload = %+v", pass)
	}
	if len(pass.Evidence) != 1 || pass.Evidence[0].Type != "guardian.order_placed" ||
		pass.Evidence[0].Seq != 41 || pass.Evidence[0].Tick != 30_000 || pass.Evidence[0].Custom {
		t.Errorf("pass evidence = %+v, want the watch placement at seq 41 tick 30000, Custom false", pass.Evidence)
	}
	var unlock StageUnlockedPayload
	if err := json.Unmarshal(batch[unlockIdx].Payload, &unlock); err != nil {
		t.Fatal(err)
	}
	if unlock.Stage != "stage-2" || unlock.Exercise != "first-night" {
		t.Errorf("unlock payload = %+v", unlock)
	}

	// Apply the batch: the reducer latches, and the outcome flips.
	for _, e := range batch {
		if err := s.Apply(e); err != nil {
			t.Fatalf("apply %s: %v", e.Type, err)
		}
	}
	if len(s.CurriculumPasses) != 1 || len(s.StagesUnlocked) != 1 || s.StagesUnlocked[0] != "stage-2" {
		t.Errorf("latch state: passes=%v unlocked=%v", s.CurriculumPasses, s.StagesUnlocked)
	}
	if got := ExerciseOutcome(s, "first-night"); got != OutcomePassed {
		t.Errorf("outcome = %q, want %q", got, OutcomePassed)
	}

	// Once-only: the state latch blocks any re-evaluation at a later
	// boundary-shaped tick (crafted directly against the latch).
	if evs := scenarioRubricEvents(s, dawn2, nil); len(evs) != 0 {
		t.Errorf("re-evaluation after the latch emitted %d events, want 0", len(evs))
	}
}

// TestFirstNightDawnDeathIsFailNotPass is the photo-finish edge: a death
// landing in the dawn batch itself (not yet folded into state) blocks the
// pass — an all-dead dawn is a fail, not a pass.
func TestFirstNightDawnDeathIsFailNotPass(t *testing.T) {
	s, m := armedFirstNight(t)
	watch := watchOrderEvent(30_000)
	watch.Seq = 1
	if err := s.Apply(watch); err != nil {
		t.Fatal(err)
	}
	// Agent 0 will die of starvation on the dawn heartbeat: food 0, health
	// at the last loss step. Everyone else pre-parked dead (not via events,
	// so the deaths LEDGER stays empty — only the batch carries the death).
	for i := range s.Agents {
		s.Agents[i].Dead = true
	}
	a := &s.Agents[0]
	a.Dead = false
	a.Needs = Needs{Health: healthLoss, Food: 0, Rest: 500, Warmth: 500, Morale: 500}

	dawn2 := clock.TickAt(2, 6, 0, 0) // %60 == 0: the needs heartbeat runs
	s.Tick = dawn2 - 1
	s.Night = true
	batch := stepEvents(s, m, dawn2)

	died, passed, ended := false, false, false
	for _, e := range batch {
		switch e.Type {
		case "agent.died":
			died = true
		case "curriculum.exercise_passed":
			passed = true
		case "run.ended":
			ended = true
		}
	}
	if !died || !ended {
		t.Fatalf("edge setup broken: died=%v ended=%v (batch %v)", died, ended, eventTypes(batch))
	}
	if passed {
		t.Fatal("an all-dead dawn produced a pass — the batch-death check failed")
	}
}

// TestFirstNightGenesisPassAndReplayEquivalence is SC-001/SC-003's backbone
// (T008): a seeded scenario world driven from genesis to its pass — watch
// placed by command, villagers kept alive by an authored fire — lands
// exercise_passed + stage_unlocked in one batch at dawn of day 2; folding
// the recorded log over the same genesis (WITHOUT arming the scenario)
// reproduces the exact state, proving the pass is recorded history, not a
// live-only judgment; and the run continuing past the pass never re-emits.
func TestFirstNightGenesisPassAndReplayEquivalence(t *testing.T) {
	m := firstNightMap()
	dawn2 := clock.TickAt(2, 6, 0, 0)

	// genesis mutates identically on both sides (the whole_feature idiom):
	// one villager, well provisioned, beside a long-lived fire — lit tiles
	// hide her from the scheduled gru, so survival is deterministic.
	genesis := func() *State {
		s := NewState(FirstNightExercise.Seed, m)
		isolateAgents(s)
		a := &s.Agents[0]
		a.Dead = false
		a.Needs = Needs{Health: 1000, Food: 900, Rest: 900, Warmth: 900, Morale: 600}
		a.Inv = Inventory{FoodRaw: 20}
		s.Structures = append(s.Structures, Structure{Kind: "fire", X: a.X, Y: a.Y + 1, FuelUntil: dawn2 + 86400})
		return s
	}

	live := genesis()
	if err := live.ArmScenario(FirstNightExercise); err != nil {
		t.Fatal(err)
	}
	commands := map[int64][]store.Event{
		30_000: {watchOrderEvent(30_000)},
	}
	log := driveTicksSeq(t, live, m, dawn2+3600, commands)

	// The pass pair: same tick, pass first, exactly once.
	var passes, unlocks []int // indexes into log
	for i, e := range log {
		switch e.Type {
		case "curriculum.exercise_passed":
			passes = append(passes, i)
		case "curriculum.stage_unlocked":
			unlocks = append(unlocks, i)
		}
	}
	if len(passes) != 1 || len(unlocks) != 1 {
		t.Fatalf("passes=%d unlocks=%d, want exactly 1 each", len(passes), len(unlocks))
	}
	if log[passes[0]].Tick != dawn2 || log[unlocks[0]].Tick != dawn2 {
		t.Errorf("pass/unlock ticks %d/%d, want the dawn boundary %d",
			log[passes[0]].Tick, log[unlocks[0]].Tick, dawn2)
	}
	if !(passes[0] < unlocks[0]) {
		t.Error("unlock preceded its pass in the log")
	}
	if got := ExerciseOutcome(live, "first-night"); got != OutcomePassed {
		t.Errorf("live outcome %q, want passed", got)
	}

	// Replay equivalence (contract §1.2): fold the recorded log over the
	// same genesis — deliberately UNARMED, the recorded events are the only
	// latches — and land on the identical state.
	replayed := genesis()
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s (seq %d): %v", e.Type, e.Seq, err)
		}
		replayed.Tick = e.Tick
	}
	replayed.Tick = live.Tick
	if live.Hash() != replayed.Hash() {
		t.Fatalf("replayed state diverged from live:\nlive:     %s\nreplayed: %s",
			live.Marshal(), replayed.Marshal())
	}
}

// TestFirstNightFailureEmitsNothing is US1 AS-3 + contract §2: a run that
// dies before the boundary emits NO curriculum event — run.ended is the fail
// signal, and the outcome derives as failed.
func TestFirstNightFailureEmitsNothing(t *testing.T) {
	m := firstNightMap()
	s := NewState(FirstNightExercise.Seed, m)
	if err := s.ArmScenario(FirstNightExercise); err != nil {
		t.Fatal(err)
	}
	isolateAgents(s)
	a := &s.Agents[0]
	a.Dead = false
	// One heartbeat from death (food 0 ⇒ health loss every beat; nothing
	// carried, so no eat can intervene before the first %60 heartbeat).
	a.Needs = Needs{Health: 1, Food: 0, Rest: 100, Warmth: 0, Morale: 100}

	log := driveTicks(t, s, m, 7200, nil) // ample: death + run end land on the first heartbeat
	ended := false
	for _, e := range log {
		switch e.Type {
		case "curriculum.exercise_passed", "curriculum.stage_unlocked":
			t.Fatalf("failed run emitted %s — failure must emit nothing", e.Type)
		case "run.ended":
			ended = true
		}
	}
	if !ended {
		t.Fatal("setup broken: the run never ended")
	}
	if got := ExerciseOutcome(s, "first-night"); got != OutcomeFailed {
		t.Errorf("outcome %q, want failed", got)
	}
	if got := ExerciseOutcome(NewState(1, m), "first-night"); got != OutcomeInProgress {
		t.Errorf("fresh state outcome %q, want in_progress", got)
	}
}

// TestIncidentVisibilityVocabulary pins the D4 resolution: definition
// override wins; else forecast at stages 1–2/pre-ladder, fog from stage 3.
func TestIncidentVisibilityVocabulary(t *testing.T) {
	plain := ExerciseDefinition{ID: "x"}
	override := ExerciseDefinition{ID: "x", IncidentVisibility: VisibilityFog}
	cases := []struct {
		def   ExerciseDefinition
		stage string
		want  string
	}{
		{plain, "stage-1", VisibilityForecast},
		{plain, "stage-2", VisibilityForecast},
		{plain, "stage-3", VisibilityFog},
		{plain, "stage-4", VisibilityFog},
		{plain, "", VisibilityForecast}, // pre-ladder: everything
		{override, "stage-1", VisibilityFog},
	}
	for _, c := range cases {
		if got := IncidentVisibilityFor(c.def, c.stage); got != c.want {
			t.Errorf("visibility(%q override=%q) = %q, want %q", c.stage, c.def.IncidentVisibility, got, c.want)
		}
	}
}

// TestAmbientStatusCarriesNoScenarioFields is contract §1.3's status slice:
// an unarmed state composes a Status with both scenario fields absent.
func TestAmbientStatusCarriesNoScenarioFields(t *testing.T) {
	s := NewState(1, testMap(1))
	if id := s.ScenarioExerciseID(); id != "" {
		t.Fatalf("ambient state reports scenario %q", id)
	}
	b, err := json.Marshal(Status{Tick: s.Tick})
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(b, []byte("scenario")) {
		t.Fatalf("ambient status bytes leak scenario fields: %s", b)
	}
}

func eventTypes(evs []store.Event) []string {
	out := make([]string, len(evs))
	for i, e := range evs {
		out[i] = e.Type
	}
	return out
}

// theLawCharterEvent builds a metatron.charter_observed fixture event.
func theLawCharterEvent(tick int64, fp string, def bool) store.Event {
	return store.Event{Tick: tick, Type: "guardian.charter_observed",
		Payload: mustPayload(CharterObservedPayload{Fingerprint: fp, Default: def})}
}

// theLawNormEvent builds a passed meeting.proposal_resolved fixture event —
// the ONLY producer of State.Norms entries (resolveProposal), so the law
// term's ledger fills exactly the way production history fills it.
func theLawNormEvent(tick int64) store.Event {
	return store.Event{Tick: tick, Type: "meeting.proposal_resolved",
		Payload: mustPayload(ProposalResolvedPayload{
			ProposalPayload: ProposalPayload{ProposalID: 1, Kind: ProposeCurfew,
				Target: Ref(-1), Proposer: Ref(0), Text: "no one leaves the fires after dusk"},
			Passed: true,
		})}
}

// TestTheLawRubricTable is the-law's rubric evaluator table test (spec 072
// FR-007, US2 — the TestFirstNightRubricTable model): per-term satisfaction
// over state facts, with every mutation arriving through the reducer — the
// charter authorship flag enters state ONLY via metatron.charter_observed.
func TestTheLawRubricTable(t *testing.T) {
	m := testMap(TheLawExercise.Seed)
	fold := func(events ...store.Event) *State {
		s := NewState(TheLawExercise.Seed, m)
		for _, e := range events {
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
			s.Tick = e.Tick
		}
		return s
	}

	cases := []struct {
		name      string
		s         *State
		wantMet   [2]bool // law adopted, player-authored charter
		wantCount [2]int
	}{
		{"genesis: both pending, zero counts",
			fold(), [2]bool{false, false}, [2]int{0, 0}},
		{"default charter observed: authorship never met by the game's own hand",
			fold(theLawCharterEvent(10, "aaaa11112222", true)), [2]bool{false, false}, [2]int{0, 1}},
		{"custom charter observed: authorship met",
			fold(theLawCharterEvent(10, "bbbb33334444", false)), [2]bool{false, true}, [2]int{0, 1}},
		{"norm adopted: law met with the ledger count",
			fold(theLawNormEvent(20)), [2]bool{true, false}, [2]int{1, 0}},
		{"both terms met",
			fold(theLawCharterEvent(10, "bbbb33334444", false), theLawNormEvent(20)),
			[2]bool{true, true}, [2]int{1, 1}},
		{"revert to the default charter flips authorship back off (present force)",
			fold(theLawCharterEvent(10, "bbbb33334444", false), theLawCharterEvent(30, "cccc55556666", true)),
			[2]bool{false, false}, [2]int{0, 1}},
	}
	for _, c := range cases {
		terms := EvaluateRubric(c.s, TheLawExercise, c.s.Tick)
		if len(terms) != 2 {
			t.Fatalf("%s: %d terms, want 2", c.name, len(terms))
		}
		for i, term := range terms {
			if term.Met != c.wantMet[i] || term.Count != c.wantCount[i] {
				t.Errorf("%s: term %q Met=%v Count=%d, want Met=%v Count=%d",
					c.name, term.Label, term.Met, term.Count, c.wantMet[i], c.wantCount[i])
			}
		}
	}
}

// TestTheLawReplayEquivalence (spec 072 US2-5): folding the recorded events
// into a live state and replaying the same log over a fresh genesis land on
// byte-identical state and identical EvaluateRubric results — the charter
// flag has no writer outside the event-sourced reducer arm, so live fold and
// genesis replay cannot diverge.
func TestTheLawReplayEquivalence(t *testing.T) {
	m := testMap(TheLawExercise.Seed)
	log := []store.Event{
		theLawCharterEvent(10, "aaaa11112222", true),
		theLawNormEvent(20),
		theLawCharterEvent(30, "bbbb33334444", false),
	}
	for i := range log {
		log[i].Seq = int64(i + 1)
	}

	live := NewState(TheLawExercise.Seed, m)
	for _, e := range log {
		if err := live.Apply(e); err != nil {
			t.Fatalf("live apply %s: %v", e.Type, err)
		}
		live.Tick = e.Tick
	}
	liveTerms := EvaluateRubric(live, TheLawExercise, live.Tick)
	if !liveTerms[0].Met || !liveTerms[1].Met {
		t.Fatalf("live fold should satisfy both terms: %+v", liveTerms)
	}

	replayed := NewState(TheLawExercise.Seed, m)
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replayed.Tick = e.Tick
	}
	if live.Hash() != replayed.Hash() {
		t.Fatalf("replayed state diverged from live:\nlive:     %s\nreplayed: %s",
			live.Marshal(), replayed.Marshal())
	}
	replayTerms := EvaluateRubric(replayed, TheLawExercise, replayed.Tick)
	for i := range liveTerms {
		if liveTerms[i] != replayTerms[i] {
			t.Errorf("term %d diverged: live %+v, replayed %+v", i, liveTerms[i], replayTerms[i])
		}
	}
}
