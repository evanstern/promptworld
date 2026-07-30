package sim

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// --- T003: snapshot compatibility (spec 048, no format_version bump) ---

// TestTuningSnapshotCompat proves the pre-048 byte-identical guarantee: a
// snapshot with no "tuning" key unmarshals to Tuning == nil with every accessor
// returning its default, and a state whose Tuning is nil re-marshals with no
// "tuning" key at all (omitempty).
func TestTuningSnapshotCompat(t *testing.T) {
	m := worldmap.Generate(42, 64, 64)

	// A default state (nil Tuning) must marshal WITHOUT a tuning key.
	s := NewState(42, m)
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), "\"tuning\"") {
		t.Fatalf("nil Tuning marshaled a tuning key — breaks pre-048 byte-identity:\n%s", b)
	}

	// A pre-048 snapshot (no tuning key) unmarshals to nil Tuning + defaults.
	back := NewState(42, m)
	if err := json.Unmarshal(b, back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Tuning != nil {
		t.Fatalf("Tuning = %+v, want nil on a keyless snapshot", back.Tuning)
	}
	assertDefaults(t, back)
}

func assertDefaults(t *testing.T, s *State) {
	t.Helper()
	if got := s.RefuelDyingBelow(); got != defaultRefuelDyingBelow {
		t.Errorf("RefuelDyingBelow() = %d, want default %d", got, defaultRefuelDyingBelow)
	}
	if got := s.FireBurnPerWood(); got != defaultFireBurnPerWood {
		t.Errorf("FireBurnPerWood() = %d, want default %d", got, defaultFireBurnPerWood)
	}
	if got := s.GruEmergePerMille(); got != defaultGruEmergePerMille {
		t.Errorf("GruEmergePerMille() = %d, want default %d", got, defaultGruEmergePerMille)
	}
	if got := s.PlannerCadence(); got != defaultPlannerCadenceTicks {
		t.Errorf("PlannerCadence() = %d, want default %d", got, defaultPlannerCadenceTicks)
	}
	if got := s.EncounterCooldown(); got != defaultEncounterCooldownTicks {
		t.Errorf("EncounterCooldown() = %d, want default %d", got, defaultEncounterCooldownTicks)
	}
}

// TestTuningNilAccessorsMatchOldConstants pins the accessor defaults (nil
// Tuning) to the exact former doctrine-constant values — the "absent file ==
// current behavior" invariant, spec 048 FR-001.
func TestTuningNilAccessorsMatchOldConstants(t *testing.T) {
	m := worldmap.Generate(7, 64, 64)
	assertDefaults(t, NewState(7, m))

	want := map[string]int64{
		"refuel_dying_below":       10800, // spec 057 / TASK-108 raised this 3600 → 10800 (3 h)
		"fire_burn_per_wood":       14400,
		"gru_emerge_per_mille":     600,
		"planner_cadence_ticks":    1800,
		"encounter_cooldown_ticks": 7200,
	}
	got := defaultTuning()
	if int64(got.RefuelDyingBelow) != want["refuel_dying_below"] ||
		int64(got.FireBurnPerWood) != want["fire_burn_per_wood"] ||
		int64(got.GruEmergePerMille) != want["gru_emerge_per_mille"] ||
		int64(got.PlannerCadenceTicks) != want["planner_cadence_ticks"] ||
		int64(got.EncounterCooldownTicks) != want["encounter_cooldown_ticks"] {
		t.Fatalf("defaultTuning() = %+v, does not match documented doctrine defaults %+v", got, want)
	}
}

// --- T005: ParseTuning table (spec 048 US1, contracts/tuning.md) ---

func TestParseTuning(t *testing.T) {
	def := defaultTuning()

	t.Run("empty object resolves to the default set", func(t *testing.T) {
		ts, warns, err := ParseTuning([]byte(`{}`))
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if len(warns) != 0 {
			t.Fatalf("warns = %v, want none", warns)
		}
		if *ts != def {
			t.Fatalf("empty object = %+v, want default set %+v", *ts, def)
		}
	})

	t.Run("sparse file fills missing fields from defaults", func(t *testing.T) {
		ts, warns, err := ParseTuning([]byte(`{"fire_burn_per_wood": 28800}`))
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if len(warns) != 0 {
			t.Fatalf("warns = %v, want none", warns)
		}
		if ts.FireBurnPerWood != 28800 {
			t.Errorf("fire_burn_per_wood = %d, want 28800", ts.FireBurnPerWood)
		}
		if ts.RefuelDyingBelow != def.RefuelDyingBelow || ts.GruEmergePerMille != def.GruEmergePerMille ||
			ts.PlannerCadenceTicks != def.PlannerCadenceTicks || ts.EncounterCooldownTicks != def.EncounterCooldownTicks {
			t.Errorf("unset fields did not resolve to defaults: %+v", *ts)
		}
	})

	// Each field clamps at BOTH bounds with the documented warning text.
	clampCases := []struct {
		name     string
		json     string
		field    func(*TuningState) int64
		want     int64
		wantWarn string
	}{
		{"refuel below min", `{"refuel_dying_below": -5}`, func(t *TuningState) int64 { return t.RefuelDyingBelow }, 0,
			"tuning.json refuel_dying_below -5 out of range (min 0) — clamped to 0"},
		{"refuel above max", `{"refuel_dying_below": 999999}`, func(t *TuningState) int64 { return t.RefuelDyingBelow }, 86400,
			"tuning.json refuel_dying_below 999999 out of range (max 86400) — clamped to 86400"},
		{"fire below min", `{"fire_burn_per_wood": 10}`, func(t *TuningState) int64 { return t.FireBurnPerWood }, 600,
			"tuning.json fire_burn_per_wood 10 out of range (min 600) — clamped to 600"},
		{"fire above max", `{"fire_burn_per_wood": 999999}`, func(t *TuningState) int64 { return t.FireBurnPerWood }, 86400,
			"tuning.json fire_burn_per_wood 999999 out of range (max 86400) — clamped to 86400"},
		{"gru above max", `{"gru_emerge_per_mille": 5000}`, func(t *TuningState) int64 { return int64(t.GruEmergePerMille) }, 1000,
			"tuning.json gru_emerge_per_mille 5000 out of range (max 1000) — clamped to 1000"},
		{"cadence below min", `{"planner_cadence_ticks": 1}`, func(t *TuningState) int64 { return t.PlannerCadenceTicks }, 60,
			"tuning.json planner_cadence_ticks 1 out of range (min 60) — clamped to 60"},
		{"cadence above max", `{"planner_cadence_ticks": 999999}`, func(t *TuningState) int64 { return t.PlannerCadenceTicks }, 86400,
			"tuning.json planner_cadence_ticks 999999 out of range (max 86400) — clamped to 86400"},
		{"cooldown above max", `{"encounter_cooldown_ticks": 999999}`, func(t *TuningState) int64 { return t.EncounterCooldownTicks }, 86400,
			"tuning.json encounter_cooldown_ticks 999999 out of range (max 86400) — clamped to 86400"},
	}
	for _, c := range clampCases {
		t.Run(c.name, func(t *testing.T) {
			ts, warns, err := ParseTuning([]byte(c.json))
			if err != nil {
				t.Fatalf("err = %v", err)
			}
			if got := c.field(ts); got != c.want {
				t.Errorf("clamped value = %d, want %d", got, c.want)
			}
			if len(warns) != 1 || warns[0] != c.wantWarn {
				t.Errorf("warns = %v, want exactly [%q]", warns, c.wantWarn)
			}
		})
	}

	// Structural rejections (fail-closed).
	rejectCases := []struct {
		name string
		json string
	}{
		{"unknown key", `{"fire_burn_per_woods": 1}`},
		{"wrong type (string)", `{"fire_burn_per_wood": "hot"}`},
		{"negative into uint64 field", `{"gru_emerge_per_mille": -3}`}, // decode error, not a clamp
		{"malformed json", `{ not json`},
	}
	for _, c := range rejectCases {
		t.Run(c.name, func(t *testing.T) {
			_, _, err := ParseTuning([]byte(c.json))
			if err == nil {
				t.Fatalf("want boot-failing error, got nil")
			}
			if !strings.Contains(err.Error(), "tuning.json") {
				t.Errorf("error %q does not name the file", err)
			}
		})
	}
}

// --- T010: reducer / seed round-trip (spec 048 US2) ---

// TestTuningAppliedReducer proves the sim.tuning_applied arm sets Tuning so the
// accessors return payload values, is idempotent under re-application, and
// survives a snapshot round-trip.
func TestTuningAppliedReducer(t *testing.T) {
	m := worldmap.Generate(42, 64, 64)
	s := NewState(42, m)

	set := TuningState{
		RefuelDyingBelow:       1200,
		FireBurnPerWood:        28800,
		GruEmergePerMille:      300,
		PlannerCadenceTicks:    900,
		EncounterCooldownTicks: 3600,
	}
	ev := NewTuningEvent(s.Tick, set)
	if err := s.Apply(ev); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if s.RefuelDyingBelow() != 1200 || s.FireBurnPerWood() != 28800 || s.GruEmergePerMille() != 300 ||
		s.PlannerCadence() != 900 || s.EncounterCooldown() != 3600 {
		t.Fatalf("accessors did not reflect payload: %+v", s.Tuning)
	}

	// Idempotent re-apply.
	if err := s.Apply(ev); err != nil {
		t.Fatalf("re-apply: %v", err)
	}
	if !s.Tuning.Equal(set) { // Equal: spec 098's Dream block resolves by value
		t.Fatalf("re-apply changed state: %+v", *s.Tuning)
	}

	// Snapshot round-trip carries Tuning.
	b, _ := json.Marshal(s)
	if !strings.Contains(string(b), "\"tuning\"") {
		t.Fatalf("tuned state marshaled without a tuning key")
	}
	back := NewState(42, m)
	if err := json.Unmarshal(b, back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Tuning == nil || !back.Tuning.Equal(set) {
		t.Fatalf("round-trip Tuning = %+v, want %+v", back.Tuning, set)
	}
}

// --- T011: replay determinism with tuning (spec 048 US2, SC-003) ---

// TestTuningReplayDeterminism drives a tuned world through fire activity that
// depends on the tuned dial (a refuel deadline computed from a non-default
// FireBurnPerWood), captures the event log — which carries the
// sim.tuning_applied event as its ONLY record of the tuning — then replays that
// log into a FRESH genesis state (nil Tuning, no file anywhere) and asserts the
// state hash matches. If replay ignored the event and fell back to defaults, the
// tuning-dependent FuelUntil would diverge and the hashes would differ.
func TestTuningReplayDeterminism(t *testing.T) {
	const seed = 42
	m := testMap(seed)

	tuned := defaultTuning()
	tuned.FireBurnPerWood = 28800 // 2× default — moves every refuel deadline

	genesis := func() *State {
		s := NewState(seed, m)
		isolateAgents(s)
		a := &s.Agents[0]
		a.Dead = false
		a.Needs = Needs{Health: 1000, Food: 600, Rest: 600, Warmth: 600, Morale: 600}
		a.Inv.Wood = 3
		s.Structures = []Structure{{Kind: "fire", X: a.X, Y: a.Y, FuelUntil: 0}} // cold fire on the agent's tile
		a.Intent = &Intent{Goal: "refuel_fire", TargetX: a.X, TargetY: a.Y}
		return s
	}

	live := genesis()
	var log []store.Event
	// Seed the tuning event first (the boot seed's job), then drive.
	tev := NewTuningEvent(live.Tick, tuned)
	if err := live.Apply(tev); err != nil {
		t.Fatalf("apply tuning event: %v", err)
	}
	log = append(log, tev)
	log = append(log, driveTicks(t, live, m, live.Tick+10, nil)...)

	// The refuel deadline must reflect the tuned burn, proving the dial drove a
	// hashed reducer field (not the default).
	var refueled bool
	for _, e := range log {
		if e.Type == "agent.refueled" {
			var p RefueledPayload
			mustUnmarshal(t, e.Payload, &p)
			if p.FuelUntil != e.Tick+tuned.FireBurnPerWood {
				t.Fatalf("refuel FuelUntil = %d, want %d (refuel tick + tuned burn %d)", p.FuelUntil, e.Tick+tuned.FireBurnPerWood, tuned.FireBurnPerWood)
			}
			refueled = true
		}
	}
	if !refueled {
		t.Fatal("scenario produced no agent.refueled event")
	}

	// Replay the whole log — including the tuning event — into a fresh genesis
	// whose Tuning starts nil. No file is consulted; sim has no file access.
	replay := genesis()
	for _, e := range log {
		if err := replay.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		if e.Tick > replay.Tick {
			replay.Tick = e.Tick
		}
	}
	if replay.Tuning == nil || !replay.Tuning.Equal(tuned) {
		t.Fatalf("replay Tuning = %+v, want the logged set %+v (must come from the event)", replay.Tuning, tuned)
	}
	// Catch replay's tick up over any trailing event-free ticks (the codebase
	// replay idiom, axe_test.go): those ticks generated nothing in the live run
	// either, so this only advances the clock — no new events, no divergence.
	driveTicks(t, replay, m, live.Tick, nil)
	if live.Hash() != replay.Hash() {
		t.Fatalf("replay hash %s != live hash %s — tuned behavior did not reproduce from the log", replay.Hash(), live.Hash())
	}
}

// --- T015: per-dial proofs (spec 048 US3, SC-005) ---

// TestTunedFireBurnPerWoodMovesDeadline proves a tuned FireBurnPerWood — not the
// default constant — computes a refuel's FuelUntil.
func TestTunedFireBurnPerWoodMovesDeadline(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	isolateAgents(s)

	tuned := defaultTuning()
	tuned.FireBurnPerWood = 28800
	if err := s.Apply(NewTuningEvent(s.Tick, tuned)); err != nil {
		t.Fatal(err)
	}

	a := &s.Agents[0]
	a.Dead = false
	a.Inv.Wood = 3
	fx, fy := a.X, a.Y
	s.Structures = append(s.Structures, Structure{Kind: "fire", X: fx, Y: fy, FuelUntil: 0})
	a.Intent = &Intent{Goal: "refuel_fire", TargetX: fx, TargetY: fy}

	log := driveTicks(t, s, m, s.Tick+5, nil)
	var got, at int64
	found := false
	for _, e := range log {
		if e.Type == "agent.refueled" {
			var p RefueledPayload
			mustUnmarshal(t, e.Payload, &p)
			got, at, found = p.FuelUntil, e.Tick, true
		}
	}
	if !found {
		t.Fatal("no agent.refueled event")
	}
	if got != at+28800 {
		t.Errorf("FuelUntil = %d, want %d (tuned burn 28800); default %d would give %d", got, at+28800, defaultFireBurnPerWood, at+defaultFireBurnPerWood)
	}
	if got == at+defaultFireBurnPerWood {
		t.Error("FuelUntil used the DEFAULT burn, not the tuned value")
	}
}

// TestTunedRefuelDyingBelowMovesTriggerWindow proves a tuned RefuelDyingBelow
// widens the reflex's refuel trigger window: a fire the default window would
// treat as healthy triggers a refuel under the tuned window.
func TestTunedRefuelDyingBelowMovesTriggerWindow(t *testing.T) {
	const seed = 42
	m := testMap(seed)

	// A fire with 15000 ticks of fuel left: healthy under the default 10800
	// window (spec 057), dying under a tuned 20000 window.
	const fuelLeft = 15000
	setup := func(s *State) {
		a := &s.Agents[0]
		a.Needs = Needs{Health: 1000, Food: 600, Rest: 600, Warmth: 600, Morale: 600}
		a.Inv.Wood = 2
		s.Structures = []Structure{{Kind: "fire", X: a.X, Y: a.Y, FuelUntil: fuelLeft}}
		grantStructureFacts(s, 0)
	}

	// Default window: no refuel (15000 >= 10800).
	base := NewState(seed, m)
	setup(base)
	if d := decideIntent(base, m, 0, 0); d.intent != nil && d.intent.Goal == "refuel_fire" {
		t.Fatalf("default window (%d) should treat %d-fuel fire as healthy — got refuel", defaultRefuelDyingBelow, fuelLeft)
	}

	// Tuned window 20000: the same fire is now dying → refuel.
	tunedS := NewState(seed, m)
	tt := defaultTuning()
	tt.RefuelDyingBelow = 20000
	if err := tunedS.Apply(NewTuningEvent(tunedS.Tick, tt)); err != nil {
		t.Fatal(err)
	}
	setup(tunedS)
	if d := decideIntent(tunedS, m, 0, 0); d.intent == nil || d.intent.Goal != "refuel_fire" {
		t.Fatalf("tuned window (20000) should treat %d-fuel fire as dying — got %+v", fuelLeft, d.intent)
	}
}

// --- spec 057 / TASK-108 US1: the refuel-default raise (SC-001) --------------

// TestRefuelDefaultRaisedToThreeHours is the US1 acceptance test: a fire with
// 2.5 game-hours of fuel left and a wood-carrying villager gets a refuel intent
// under the new default (10800); under a tuning.json pinning the OLD value
// (3600) the same fire is healthy and no refuel arms — the dial still wins over
// the doctrine default (dial semantics unchanged, spec 048 preserved).
func TestRefuelDefaultRaisedToThreeHours(t *testing.T) {
	const seed = 42
	m := testMap(seed)

	// 2.5 game-hours remaining: dying under the new 10800 default, healthy under
	// the old 3600 window.
	const fuelLeft = 9000 // 2.5 * 3600
	setup := func(s *State) {
		a := &s.Agents[0]
		a.Needs = Needs{Health: 1000, Food: 600, Rest: 600, Warmth: 600, Morale: 600}
		a.Inv.Wood = 2
		s.Structures = []Structure{{Kind: "fire", X: a.X, Y: a.Y, FuelUntil: fuelLeft}}
		grantStructureFacts(s, 0)
	}

	// New default (nil Tuning ⇒ 10800): the fire is dying → refuel.
	def := NewState(seed, m)
	setup(def)
	if got := def.RefuelDyingBelow(); got != 10800 {
		t.Fatalf("default RefuelDyingBelow() = %d, want 10800 (spec 057)", got)
	}
	if d := decideIntent(def, m, 0, 0); d.intent == nil || d.intent.Goal != "refuel_fire" {
		t.Fatalf("new default (10800) should treat a %d-fuel fire as dying — got %+v", fuelLeft, d.intent)
	}

	// A tuning.json pinning the OLD default (3600): the same fire is healthy →
	// no refuel. The dial overrides the doctrine default (spec 048 semantics).
	pinned := NewState(seed, m)
	old := defaultTuning()
	old.RefuelDyingBelow = 3600
	if err := pinned.Apply(NewTuningEvent(pinned.Tick, old)); err != nil {
		t.Fatal(err)
	}
	setup(pinned)
	if d := decideIntent(pinned, m, 0, 0); d.intent != nil && d.intent.Goal == "refuel_fire" {
		t.Fatalf("manifest pinning 3600 should treat a %d-fuel fire as healthy — got a refuel (dial did not win)", fuelLeft)
	}
}

// TestTunedGruEmergePerMilleFlipsRoll proves the tuned per-mille drives the
// nightly emergence gate: 0 disables emergence entirely, 1000 forces it.
func TestTunedGruEmergePerMilleFlipsRoll(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	const nightTick = 22 * 3600 // first night

	never := NewState(seed, m)
	nt := defaultTuning()
	nt.GruEmergePerMille = 0
	if err := never.Apply(NewTuningEvent(never.Tick, nt)); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := gruEmergence(never, m, nightTick); ok {
		t.Error("per-mille 0 must never emerge")
	}

	always := NewState(seed, m)
	at := defaultTuning()
	at.GruEmergePerMille = 1000
	if err := always.Apply(NewTuningEvent(always.Tick, at)); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := gruEmergence(always, m, nightTick); !ok {
		t.Error("per-mille 1000 must always emerge (a spawn tile exists on a 64×64 map)")
	}
}

// --- T012: old-log compatibility (spec 048 FR-007) ---

// TestPreO48LogReplaysDefaults proves a pre-048 event sequence (no tuning
// events) leaves Tuning nil and every accessor on its default, and that the
// unknown-arm-tolerant Apply dispatch does not choke replaying such a log.
func TestPreO48LogReplaysDefaults(t *testing.T) {
	m := worldmap.Generate(42, 64, 64)
	s := NewState(42, m)
	// A handful of ordinary events with no tuning among them.
	log := []store.Event{
		{Tick: 1, Type: "clock.resumed"},
		{Tick: 2, Type: "world.created"},
	}
	for _, e := range log {
		if err := s.Apply(e); err != nil {
			t.Fatalf("apply %s: %v", e.Type, err)
		}
	}
	if s.Tuning != nil {
		t.Fatalf("Tuning = %+v after a pre-048 log, want nil", s.Tuning)
	}
	assertDefaults(t, s)
}

// --- spec 057 / TASK-108 US2: genesis-pin replay independence (SC-002) -------

// TestGenesisPinReplayIndependentOfCompiledDefault is the SC-002 acceptance
// test: a post-057 world's replay is hash-identical even after a compiled
// default is changed out from under it, because the genesis pin (a
// sim.tuning_applied event in the log) carries the values — replay derives
// tuning exclusively from the log, never from the binary's current default*
// constant. A "future binary whose defaults differ" is simulated by pre-seeding
// the replay state with a deliberately-wrong tuning set (a stand-in for changed
// default* constants) before replaying the real log: the genesis pin event must
// overwrite it back to the recorded set.
func TestGenesisPinReplayIndependentOfCompiledDefault(t *testing.T) {
	const seed = 42
	m := testMap(seed)

	// A pinned post-057 world with fire activity that depends on the tuned dials
	// (a refuel deadline computed from FireBurnPerWood — a hashed reducer field).
	genesis := func() *State {
		s := NewState(seed, m)
		isolateAgents(s)
		a := &s.Agents[0]
		a.Dead = false
		a.Needs = Needs{Health: 1000, Food: 600, Rest: 600, Warmth: 600, Morale: 600}
		a.Inv.Wood = 3
		s.Structures = []Structure{{Kind: "fire", X: a.X, Y: a.Y, FuelUntil: 0}} // cold fire on the agent's tile
		a.Intent = &Intent{Goal: "refuel_fire", TargetX: a.X, TargetY: a.Y}
		return s
	}

	live := genesis()
	pin := GenesisTuningEvent(live.Tick) // payload = the compiled defaults NOW
	if err := live.Apply(pin); err != nil {
		t.Fatalf("apply genesis pin: %v", err)
	}
	var log []store.Event
	log = append(log, pin)
	log = append(log, driveTicks(t, live, m, live.Tick+10, nil)...)

	// The refuel deadline must reflect the pinned FireBurnPerWood, proving the
	// pinned dial drove a hashed field.
	refueled := false
	for _, e := range log {
		if e.Type == "agent.refueled" {
			var p RefueledPayload
			mustUnmarshal(t, e.Payload, &p)
			if p.FuelUntil != e.Tick+defaultFireBurnPerWood {
				t.Fatalf("refuel FuelUntil = %d, want %d (refuel tick + pinned burn)", p.FuelUntil, e.Tick+defaultFireBurnPerWood)
			}
			refueled = true
		}
	}
	if !refueled {
		t.Fatal("scenario produced no agent.refueled event")
	}

	// Simulate a future binary whose compiled defaults differ: pre-seed the
	// fresh replay state with a deliberately-wrong tuning set, then replay the
	// real log. The genesis pin event (log[0]) must overwrite it.
	replay := genesis()
	wrong := defaultTuning()
	wrong.RefuelDyingBelow = 1  // stand-in for a changed default* constant
	wrong.FireBurnPerWood = 999 // "
	if err := replay.Apply(NewTuningEvent(replay.Tick, wrong)); err != nil {
		t.Fatalf("apply wrong-default stand-in: %v", err)
	}
	for _, e := range log {
		if err := replay.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		if e.Tick > replay.Tick {
			replay.Tick = e.Tick
		}
	}
	// The pin won: replay carries the recorded genesis set, not the wrong one.
	want := defaultTuning() // == the genesis pin's payload
	if replay.Tuning == nil || !replay.Tuning.Equal(want) {
		t.Fatalf("replay Tuning = %+v, want the pinned genesis set %+v (the pin must win over the binary default)", replay.Tuning, want)
	}
	driveTicks(t, replay, m, live.Tick, nil) // re-live the quiet tail, as recovery does
	if live.Hash() != replay.Hash() {
		t.Fatalf("replay hash %s != live hash %s — the pin did not reproduce behavior over a changed default", replay.Hash(), live.Hash())
	}
}

// TestPreO57LogReplaysUnderCompiledDefault is the FR-007 companion: a pre-057
// world carries NO genesis pin, so its log has no tuning event, its Tuning stays
// nil, and the accessors return the CURRENT compiled default — including the
// spec-057 refuel raise (10800). This is the documented, intended live effect:
// the default change reaches un-pinned worlds at their next boot.
func TestPreO57LogReplaysUnderCompiledDefault(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)

	// A pre-057 log: ordinary events, no sim.tuning_applied among them.
	log := []store.Event{
		{Tick: 0, Type: "world.created"},
		{Tick: 1, Type: "clock.resumed"},
	}
	for _, e := range log {
		if err := s.Apply(e); err != nil {
			t.Fatalf("apply %s: %v", e.Type, err)
		}
	}
	if s.Tuning != nil {
		t.Fatalf("Tuning = %+v after a pre-057 log, want nil (no pin)", s.Tuning)
	}
	// The compiled default reaches the un-pinned world — the raised 10800.
	if got := s.RefuelDyingBelow(); got != 10800 {
		t.Errorf("pre-057 world RefuelDyingBelow() = %d, want the compiled default 10800", got)
	}
}
