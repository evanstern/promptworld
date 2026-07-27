package sim

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// The stranger's behavior suite (spec 077 US2 AS-3, T010/T011): arrival via
// the schedule, store-seeking movement, bounded takes, protection rules,
// dawn departure, gru coexistence, and genesis-replay byte-identity.

// strangerBorderTile finds a passable, unprotected border tile on m.
func strangerBorderTile(t *testing.T, m *worldmap.Map, s *State) (int, int) {
	t.Helper()
	for x := 0; x < m.W; x++ {
		if strangerEntryValid(s, m, x, 0) {
			return x, 0
		}
	}
	t.Fatal("no valid border entry tile on the test map")
	return 0, 0
}

// strangerScenario arms a schedule with one stranger_arrives entry at
// day 1 23:00 on the given seed's map, villagers parked dead for a pure
// schedule frame unless keep is set.
func strangerScenario(t *testing.T, seed uint64) (*State, *worldmap.Map, int64, int, int) {
	t.Helper()
	m := testMap(seed)
	probe := NewState(seed, m)
	ex, ey := strangerBorderTile(t, m, probe)
	def := ExerciseDefinition{ID: "first-night", Stage: "stage-1", Seed: seed,
		Schedule: []IncidentScheduleEntry{{Kind: IncidentStrangerArrives, Day: 1, Time: "23:00", X: ex, Y: ey}}}
	s := NewState(seed, m)
	if err := s.ArmScenario(def); err != nil {
		t.Fatal(err)
	}
	for i := range s.Agents {
		s.Agents[i].Dead = true
	}
	return s, m, clock.TickAt(1, 23, 0, 0), ex, ey
}

// TestStrangerArrivalTakesAndDawnDeparture drives the full night arc: the
// authored arrival lands once at its tick, the entity moves toward an
// unattended pile, takes bounded goods (ledger + pile decrement), and is
// gone by dawn (state nil) — with genesis replay reproducing the exact state.
func TestStrangerArrivalTakesAndDawnDeparture(t *testing.T) {
	s, m, authored, ex, ey := strangerScenario(t, 11)
	// An unattended larder in the open, near the entry tile: find a passable,
	// unprotected tile a few steps in and drop goods there.
	px, py := -1, -1
	for y := 0; y < m.H && px < 0; y++ {
		for x := 0; x < m.W; x++ {
			if strangerEntryValid(s, m, x, y) && abs(x-ex)+abs(y-ey) >= 2 && abs(x-ex)+abs(y-ey) <= 12 {
				px, py = x, y
				break
			}
		}
	}
	if px < 0 {
		t.Fatal("no pile tile near the entry")
	}
	pile := s.pileFor(px, py)
	pile.addNonFood("wood", 5)

	live := s
	log := []store.Event{}
	var lastSeq int64
	dawn2 := clock.TickAt(2, 6, 0, 0)
	live.Tick = authored - 10
	live.Night = true
	for live.Tick < dawn2+60 {
		next := live.Tick + 1
		for _, e := range stepEvents(live, m, next) {
			lastSeq++
			e.Seq = lastSeq
			if err := live.Apply(e); err != nil {
				t.Fatalf("apply %s: %v", e.Type, err)
			}
			log = append(log, e)
		}
		live.Tick = next
	}

	var arrivals, moves, takes, departures int
	for _, e := range log {
		switch e.Type {
		case "stranger.arrived":
			arrivals++
			if e.Tick != authored {
				t.Errorf("arrival at %d, want authored %d", e.Tick, authored)
			}
			var p StrangerArrivedPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				t.Fatal(err)
			}
			if p.X != ex || p.Y != ey {
				t.Errorf("arrival at (%d,%d), want authored (%d,%d)", p.X, p.Y, ex, ey)
			}
		case "stranger.moved":
			moves++
		case "stranger.took":
			takes++
			var p StrangerTookPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				t.Fatal(err)
			}
			if p.N < 1 || p.N > strangerTakeMax {
				t.Errorf("take of %d outside the bounded [1,%d]", p.N, strangerTakeMax)
			}
		case "stranger.departed":
			departures++
		}
	}
	if arrivals != 1 {
		t.Fatalf("%d arrivals, want exactly 1 (latch)", arrivals)
	}
	if moves == 0 {
		t.Error("the stranger never moved")
	}
	if takes == 0 {
		t.Error("the stranger never took from the unattended pile")
	}
	if departures != 1 {
		t.Fatalf("%d departures, want exactly 1 at dawn", departures)
	}
	if live.Stranger != nil {
		t.Error("stranger still on state after dawn")
	}
	if len(live.StrangerTakes) != takes {
		t.Errorf("ledger has %d records, want %d", len(live.StrangerTakes), takes)
	}
	// The pile lost what the ledger says (goods leave through the same
	// state shapes agent withdrawal uses).
	taken := 0
	for _, rec := range live.StrangerTakes {
		if rec.Kind == "wood" {
			taken += rec.N
		}
	}
	if got := 5 - taken; live.pileAt(px, py) != nil && live.pileAt(px, py).avail("wood") != got {
		t.Errorf("pile holds %d wood, want %d", live.pileAt(px, py).avail("wood"), got)
	}

	// Genesis replay (US2 AS-5): fold the recorded log over the same genesis
	// — UNARMED, no scenario runtime — and land on byte-identical state.
	replayed := NewState(11, m)
	for i := range replayed.Agents {
		replayed.Agents[i].Dead = true
	}
	rp := replayed.pileFor(px, py)
	rp.addNonFood("wood", 5)
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replayed.Tick = e.Tick
	}
	replayed.Tick = live.Tick
	if !bytes.Equal(live.Marshal(), replayed.Marshal()) {
		t.Fatal("stranger night replay diverged from live run")
	}
}

// TestStrangerNeverEntersProtectedTiles (spec 077 FR-012): fire light and
// shelter are absolute — no recorded move lands on a protected tile, and a
// store inside the light is never touched (the gru's protection rules,
// shared predicates).
func TestStrangerNeverEntersProtectedTiles(t *testing.T) {
	s, m, authored, ex, ey := strangerScenario(t, 11)
	// A lit larder: pile beside a long fire, both near the entry.
	px, py := -1, -1
	for y := 0; y < m.H && px < 0; y++ {
		for x := 0; x < m.W; x++ {
			if strangerEntryValid(s, m, x, y) && abs(x-ex)+abs(y-ey) >= 2 && abs(x-ex)+abs(y-ey) <= 10 {
				px, py = x, y
				break
			}
		}
	}
	if px < 0 {
		t.Fatal("no pile tile near the entry")
	}
	pile := s.pileFor(px, py)
	pile.addNonFood("wood", 5)
	s.Structures = append(s.Structures, Structure{Kind: "fire", X: px, Y: py, FuelUntil: 1 << 40})

	dawn2 := clock.TickAt(2, 6, 0, 0)
	s.Tick = authored - 10
	s.Night = true
	for s.Tick < dawn2+60 {
		next := s.Tick + 1
		for _, e := range stepEvents(s, m, next) {
			switch e.Type {
			case "stranger.moved":
				var p StrangerMovedPayload
				if err := json.Unmarshal(e.Payload, &p); err != nil {
					t.Fatal(err)
				}
				if gruProtected(s, p.X, p.Y) {
					t.Fatalf("stranger stepped into a protected tile (%d,%d)", p.X, p.Y)
				}
			case "stranger.took":
				t.Fatal("stranger took from a fire-lit store")
			}
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		s.Tick = next
	}
	if pile := s.pileAt(px, py); pile == nil || pile.avail("wood") != 5 {
		t.Error("the lit larder was touched")
	}
}

// TestStrangerAttendedStoreIsSafe (research R4): a store with a living
// villager adjacent is attended — no take lands while the watcher stands by.
func TestStrangerAttendedStoreIsSafe(t *testing.T) {
	s, m, authored, ex, ey := strangerScenario(t, 11)
	px, py := -1, -1
	for y := 0; y < m.H && px < 0; y++ {
		for x := 0; x < m.W; x++ {
			if strangerEntryValid(s, m, x, y) && abs(x-ex)+abs(y-ey) >= 2 && abs(x-ex)+abs(y-ey) <= 10 {
				px, py = x, y
				break
			}
		}
	}
	if px < 0 {
		t.Fatal("no pile tile near the entry")
	}
	pile := s.pileFor(px, py)
	pile.addNonFood("wood", 5)
	// One living watcher standing on the pile, asleep (no reflex movement —
	// attendance is presence, not wakefulness).
	a := &s.Agents[0]
	a.Dead = false
	a.Asleep = true
	a.X, a.Y = px, py
	a.Needs = Needs{Health: 1000, Food: 900, Rest: 900, Warmth: 900, Morale: 500}

	dawn2 := clock.TickAt(2, 6, 0, 0)
	s.Tick = authored - 10
	s.Night = true
	for s.Tick < dawn2+60 {
		next := s.Tick + 1
		for _, e := range stepEvents(s, m, next) {
			if e.Type == "stranger.took" {
				t.Fatal("stranger took from an attended store")
			}
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		s.Tick = next
	}
}

// TestStrangerAndGruSameNightIndependent (spec 077 edge): both entities
// abroad the same night are independent latches — neither preempts the
// other's schedule entry, and both leave at dawn through their own arms.
func TestStrangerAndGruSameNightIndependent(t *testing.T) {
	m := testMap(11)
	probe := NewState(11, m)
	ex, ey := strangerBorderTile(t, m, probe)
	def := ExerciseDefinition{ID: "first-night", Stage: "stage-1", Seed: 11,
		Schedule: []IncidentScheduleEntry{
			{Kind: IncidentGruEmerges, Day: 1, Time: "22:30", X: ex, Y: ey},
			{Kind: IncidentStrangerArrives, Day: 1, Time: "23:00", X: ex, Y: ey},
		}}
	s := NewState(11, m)
	if err := s.ArmScenario(def); err != nil {
		t.Fatal(err)
	}
	for i := range s.Agents {
		s.Agents[i].Dead = true
	}
	dawn2 := clock.TickAt(2, 6, 0, 0)
	s.Tick = clock.TickAt(1, 22, 0, 0) - 10
	var emerged, arrived, withdrew, departed bool
	for s.Tick < dawn2+60 {
		next := s.Tick + 1
		batch := stepEvents(s, m, next)
		gruIdx, strangerIdx := -1, -1
		for i, e := range batch {
			switch e.Type {
			case "gru.emerged":
				emerged = true
			case "stranger.arrived":
				arrived = true
			case "gru.withdrew":
				withdrew = true
				gruIdx = i
			case "stranger.departed":
				departed = true
				strangerIdx = i
			}
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		// Order pin (T010): strangerStep runs after gruStep in the batch.
		if gruIdx >= 0 && strangerIdx >= 0 && gruIdx > strangerIdx {
			t.Fatal("stranger.departed preceded gru.withdrew — step order broken")
		}
		s.Tick = next
	}
	if !emerged || !arrived {
		t.Fatalf("emerged=%v arrived=%v — both entities should be abroad the same night", emerged, arrived)
	}
	if !withdrew || !departed {
		t.Fatalf("withdrew=%v departed=%v — both should leave at dawn", withdrew, departed)
	}
}

// TestStrangerWindowLapsesAndPreconditionSkips (US2 AS-2 class): a stranger
// already abroad blocks the entry (belt + latch), and a time snap past the
// window skips it silently forever.
func TestStrangerWindowLapsesAndPreconditionSkips(t *testing.T) {
	// A stranger already abroad at the authored tick.
	s, m, authored, _, _ := strangerScenario(t, 11)
	s.Tick = authored - 1
	s.Night = true
	s.Stranger = &Stranger{X: 5, Y: 5, Night: 1}
	for i := int64(0); i < 300; i++ {
		next := s.Tick + 1
		for _, e := range stepEvents(s, m, next) {
			if e.Type == "stranger.arrived" {
				t.Fatalf("arrival at %d despite a stranger abroad", next)
			}
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		s.Tick = next
	}

	// Snapped past the window: lapsed, never retried.
	s2, m2, _, _, _ := strangerScenario(t, 11)
	s2.Tick = clock.TickAt(2, 7, 0, 0)
	s2.Night = false
	for i := int64(0); i < 200; i++ {
		next := s2.Tick + 1
		for _, e := range stepEvents(s2, m2, next) {
			if e.Type == "stranger.arrived" {
				t.Fatalf("lapsed arrival fired at %d", next)
			}
			if err := s2.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		s2.Tick = next
	}
}

// TestStrangerTheftWitnessMemory (research R4): an awake villager near
// enough to see a take records the moment — rumor fuel through the gru's
// situated-memory idiom.
func TestStrangerTheftWitnessMemory(t *testing.T) {
	m := testMap(11)
	s := NewState(11, m)
	isolateAgents(s)
	// A witness two tiles from the pile, awake, alive — and far enough not
	// to attend the store (adjacency 1 is attendance; 2 is a witness).
	px, py := -1, -1
	probe := NewState(11, m)
	for y := 2; y < m.H-2 && px < 0; y++ {
		for x := 2; x < m.W-2; x++ {
			if strangerEntryValid(probe, m, x, y) && strangerEntryValid(probe, m, x+2, y) {
				px, py = x, y
				break
			}
		}
	}
	if px < 0 {
		t.Fatal("no open tile pair")
	}
	a := &s.Agents[0]
	a.Dead = false
	a.Asleep = false
	a.X, a.Y = px+2, py
	a.Needs = Needs{Health: 1000, Food: 900, Rest: 900, Warmth: 900, Morale: 500}
	a.IdleSince = 1 << 40 // park the reflex

	pile := s.pileFor(px, py)
	pile.addNonFood("wood", 3)
	s.Stranger = &Stranger{X: px, Y: py, Night: 1}
	s.Night = true
	s.Tick = clock.TickAt(1, 23, 0, 0)

	batch := stepEvents(s, m, s.Tick+1)
	var took, memory bool
	for _, e := range batch {
		switch e.Type {
		case "stranger.took":
			took = true
		case "agent.memory_added":
			var p MemoryAddedPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				t.Fatal(err)
			}
			if p.Agent.ID == 0 && p.Origin == OriginWitness {
				memory = true
			}
		}
	}
	if !took {
		t.Fatal("setup broken: no take landed")
	}
	if !memory {
		t.Error("a witnessed theft left no situated memory")
	}
}

// TestAmbientWorldCarriesNoIncidentState (spec 077 FR-017, T012): an ambient
// (no-scenario) world's stream and state carry NONE of the new vocabulary —
// no new event types, no new state keys — so its bytes are pre-077-identical
// by construction (no code path emits the kinds without an armed schedule).
func TestAmbientWorldCarriesNoIncidentState(t *testing.T) {
	const seed, ticks = 7, 30_000
	m := testMap(seed)
	s := NewState(seed, m)
	log := driveTicks(t, s, m, ticks, nil)
	newTypes := map[string]bool{
		"sim.cold_snap": true, "sim.forage_blighted": true,
		"stranger.arrived": true, "stranger.moved": true,
		"stranger.took": true, "stranger.departed": true,
		"metatron.skills_observed": true,
	}
	for _, e := range log {
		if newTypes[e.Type] {
			t.Fatalf("ambient world emitted %s", e.Type)
		}
	}
	got := s.Marshal()
	for _, key := range []string{`"cold_snap_until"`, `"stranger"`, `"stranger_takes"`,
		`"skills_fingerprint"`, `"charter_observed_seq"`} {
		if bytes.Contains(got, []byte(key)) {
			t.Errorf("ambient state carries the spec-077 key %s", key)
		}
	}
}
