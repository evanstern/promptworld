package sim

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/store"
)

// governorEvent builds a reducer-applied governor event from a payload.
func governorEvent(typ string, p GovernorPayload) store.Event {
	return store.Event{Tick: 1, Type: typ, Payload: mustPayload(p)}
}

// TestGovernorShedReducer (spec 028 US2-AC1, contracts/events.md): applying a
// clock.governor_shed sets the effective Speed to `to`, records the player's
// ceiling in RequestedSpeed, and follows the new speed's tick rate.
func TestGovernorShedReducer(t *testing.T) {
	s := NewState(42, testMap(42))
	s.Speed = clock.Speed32x
	s.EffectiveRate = clock.Speed32x.TicksPerSecond()

	ev := governorEvent("clock.governor_shed", GovernorPayload{
		Requested: clock.Speed32x, From: clock.Speed32x, To: clock.Speed16x, Debt: 1.4, Jobs: 3,
	})
	if err := s.Apply(ev); err != nil {
		t.Fatalf("apply governor_shed: %v", err)
	}
	if s.Speed != clock.Speed16x {
		t.Errorf("Speed = %q, want 16x (effective drops one notch)", s.Speed)
	}
	if s.RequestedSpeed != clock.Speed32x {
		t.Errorf("RequestedSpeed = %q, want 32x (the ceiling is preserved)", s.RequestedSpeed)
	}
	if s.EffectiveRate != clock.Speed16x.TicksPerSecond() {
		t.Errorf("EffectiveRate = %v, want %v", s.EffectiveRate, clock.Speed16x.TicksPerSecond())
	}
}

// TestGovernorShedDegradedKeepsRate (contracts/events.md, "unless Degraded"): a
// shed applied to a host already reporting a degraded pace updates Speed but
// leaves EffectiveRate under the auto-slow observer's control.
func TestGovernorShedDegradedKeepsRate(t *testing.T) {
	s := NewState(42, testMap(42))
	s.Speed = clock.Speed32x
	s.Degraded = true
	s.EffectiveRate = 6.5 // the honest-slowdown observer's measured rate

	ev := governorEvent("clock.governor_shed", GovernorPayload{
		Requested: clock.Speed32x, From: clock.Speed32x, To: clock.Speed16x, Debt: 2.0, Jobs: 4,
	})
	if err := s.Apply(ev); err != nil {
		t.Fatalf("apply governor_shed: %v", err)
	}
	if s.Speed != clock.Speed16x {
		t.Errorf("Speed = %q, want 16x", s.Speed)
	}
	if s.RequestedSpeed != clock.Speed32x {
		t.Errorf("RequestedSpeed = %q, want 32x", s.RequestedSpeed)
	}
	if s.EffectiveRate != 6.5 {
		t.Errorf("EffectiveRate = %v, want 6.5 (degraded rate untouched by the governor)", s.EffectiveRate)
	}
}

// TestGovernorRecoveredReducer (spec 028 US3-AC1, contracts/events.md): a
// clock.governor_recovered raises Speed one notch; RequestedSpeed stays set
// while still below the ceiling and clears the instant it reaches the ceiling.
func TestGovernorRecoveredReducer(t *testing.T) {
	// Recover 8x -> 16x with a 32x ceiling: still governed, ceiling preserved.
	s := NewState(42, testMap(42))
	s.Speed = clock.Speed8x
	s.RequestedSpeed = clock.Speed32x
	s.EffectiveRate = clock.Speed8x.TicksPerSecond()
	if err := s.Apply(governorEvent("clock.governor_recovered", GovernorPayload{
		Requested: clock.Speed32x, From: clock.Speed8x, To: clock.Speed16x, Debt: 0.3, Jobs: 1,
	})); err != nil {
		t.Fatalf("apply governor_recovered: %v", err)
	}
	if s.Speed != clock.Speed16x {
		t.Errorf("Speed = %q, want 16x", s.Speed)
	}
	if s.RequestedSpeed != clock.Speed32x {
		t.Errorf("RequestedSpeed = %q, want 32x (still below the ceiling)", s.RequestedSpeed)
	}
	if s.EffectiveRate != clock.Speed16x.TicksPerSecond() {
		t.Errorf("EffectiveRate = %v, want %v", s.EffectiveRate, clock.Speed16x.TicksPerSecond())
	}

	// Recover 16x -> 32x reaching the 32x ceiling: governed state clears.
	if err := s.Apply(governorEvent("clock.governor_recovered", GovernorPayload{
		Requested: clock.Speed32x, From: clock.Speed16x, To: clock.Speed32x, Debt: 0.1, Jobs: 0,
	})); err != nil {
		t.Fatalf("apply governor_recovered (to ceiling): %v", err)
	}
	if s.Speed != clock.Speed32x {
		t.Errorf("Speed = %q, want 32x", s.Speed)
	}
	if s.RequestedSpeed != "" {
		t.Errorf("RequestedSpeed = %q, want empty (reached the ceiling, governed state cleared)", s.RequestedSpeed)
	}
	if s.EffectiveRate != clock.Speed32x.TicksPerSecond() {
		t.Errorf("EffectiveRate = %v, want %v", s.EffectiveRate, clock.Speed32x.TicksPerSecond())
	}
}

// TestSpeedSetClearsGovernedState (spec 028 FR-009): a player speed command
// collapses any standing governor ceiling — RequestedSpeed clears and the
// requested speed becomes both ceiling and effective.
func TestSpeedSetClearsGovernedState(t *testing.T) {
	s := NewState(42, testMap(42))
	s.Speed = clock.Speed8x
	s.RequestedSpeed = clock.Speed32x // governed: asked 32x, running 8x
	s.EffectiveRate = clock.Speed8x.TicksPerSecond()

	if err := s.Apply(store.Event{Tick: 1, Type: "clock.speed_set",
		Payload: mustPayload(SpeedSetPayload{Speed: clock.Speed4x})}); err != nil {
		t.Fatalf("apply speed_set: %v", err)
	}
	if s.Speed != clock.Speed4x {
		t.Errorf("Speed = %q, want 4x", s.Speed)
	}
	if s.RequestedSpeed != "" {
		t.Errorf("RequestedSpeed = %q, want empty (a player command clears governed state)", s.RequestedSpeed)
	}
	if s.EffectiveRate != clock.Speed4x.TicksPerSecond() {
		t.Errorf("EffectiveRate = %v, want %v", s.EffectiveRate, clock.Speed4x.TicksPerSecond())
	}
}

// TestUngovernedSnapshotOmitsRequestedSpeed (spec 028 R2/R3, SC-001): an
// ungoverned State marshals with NO requested_speed key, so every pre-028
// snapshot byte shape is preserved. A governed State does carry the key.
func TestUngovernedSnapshotOmitsRequestedSpeed(t *testing.T) {
	s := NewState(42, testMap(42))
	if got := s.Marshal(); bytes.Contains(got, []byte("requested_speed")) {
		t.Errorf("ungoverned snapshot leaked requested_speed key:\n%s", got)
	}

	s.RequestedSpeed = clock.Speed32x
	if got := s.Marshal(); !bytes.Contains(got, []byte(`"requested_speed":"32x"`)) {
		t.Errorf("governed snapshot missing requested_speed key:\n%s", got)
	}
}

// TestStructureHPOmitemptyStable (spec 032 T002, research R7): the additive
// Structure.HP field is omitempty, so a non-wall structure (or any pre-032
// structure, which never set HP) marshals with NO "hp" key — pre-032 snapshot
// bytes are unchanged. A standing wall (HP ≥ 1) does carry the key.
func TestStructureHPOmitemptyStable(t *testing.T) {
	s := NewState(42, testMap(42))
	// A pre-032-shaped structure (fire) never sets HP → the key must not appear.
	s.Structures = []Structure{{Kind: "fire", X: 1, Y: 1, FuelUntil: 8 * 3600}}
	if got := s.Marshal(); bytes.Contains(got, []byte(`"hp"`)) {
		t.Errorf("non-wall structure leaked an hp key:\n%s", got)
	}
	// A standing wall carries its current health.
	s.Structures = []Structure{{Kind: "wall_plank", X: 2, Y: 2, HP: wallPlankHP}}
	if got := s.Marshal(); !bytes.Contains(got, []byte(`"hp":200`)) {
		t.Errorf("wall snapshot missing hp key:\n%s", got)
	}
}

// TestAxesOmitemptyStable (spec 032 T011, research R7): Inventory.Axes and
// Pile.Axes are omitempty, so an agent or pile carrying no axes marshals with NO
// "axes" key — pre-032 snapshot bytes are unchanged. Carried axes do serialize,
// sorted ascending.
func TestAxesOmitemptyStable(t *testing.T) {
	s := NewState(42, testMap(42))
	if got := s.Marshal(); bytes.Contains(got, []byte(`"axes"`)) {
		t.Errorf("a fresh (axe-less) world leaked an axes key:\n%s", got)
	}
	// Carried axes serialize under the "axes" key.
	s.Agents[0].Inv.Axes = []int{3, 10}
	if got := s.Marshal(); !bytes.Contains(got, []byte(`"axes":[3,10]`)) {
		t.Errorf("inventory axes not serialized:\n%s", got)
	}
	// A pile carrying only axes is non-empty (empty() sees them).
	p := Pile{X: 1, Y: 1, Axes: []int{5}}
	if p.empty() {
		t.Error("a pile holding an axe must not report empty")
	}
	if got := s.Marshal(); !bytes.Contains(got, []byte(`"axes"`)) {
		t.Error("pile/inventory axes key missing after adding axes")
	}
}

// TestMapOmitemptyStable (spec 041 T004, the TestAxesOmitemptyStable twin):
// Agent.Map is a pointer with omitempty, so a pre-041 snapshot (no "map" key,
// Map nil) round-trips BYTE-IDENTICALLY — unmarshal then re-marshal reproduces
// the exact bytes and leaks no "map" key. A present map does serialize, with
// its canonical field order.
func TestMapOmitemptyStable(t *testing.T) {
	s := NewState(42, testMap(42))
	// Fresh v4 worlds seed a map per agent at genesis; a pre-041 snapshot has
	// none — simulate its shape by clearing them.
	for i := range s.Agents {
		s.Agents[i].Map = nil
	}
	pre := s.Marshal()
	if bytes.Contains(pre, []byte(`"map":`)) {
		t.Fatalf("a map-less state leaked a map key:\n%s", pre)
	}
	var back State
	if err := json.Unmarshal(pre, &back); err != nil {
		t.Fatal(err)
	}
	if got := back.Marshal(); !bytes.Equal(got, pre) {
		t.Errorf("pre-041 snapshot did not round-trip byte-identically:\npre:  %s\npost: %s", pre, got)
	}
	// A present map serializes: explored bitmap first, facts omitted when nil.
	s.Agents[0].Map = newMentalMap(64, 64)
	got := s.Marshal()
	if !bytes.Contains(got, []byte(`"map":{"explored":`)) {
		t.Errorf("agent map not serialized:\n%s", got)
	}
	if bytes.Contains(got, []byte(`"facts"`)) {
		t.Errorf("a fact-less map leaked a facts key:\n%s", got)
	}
	s.Agents[0].Map.upsertFact(PlaceFact{Kind: "fire", X: 1, Y: 2, Seen: 100, Provenance: ProvenanceWitnessed, Detail: 400})
	if got := s.Marshal(); !bytes.Contains(got, []byte(`"facts":[{"kind":"fire","x":1,"y":2,"seen":100,"prov":"witnessed","detail":400}]`)) {
		t.Errorf("fact not serialized in canonical field order:\n%s", got)
	}
}

// memoryAddedEvent builds a reducer-applied agent.memory_added event carrying
// a store seq — the spec-042 identity the reducer stamps onto the Memory.
func memoryAddedEvent(seq, tick int64, agent int, text string) store.Event {
	return store.Event{Seq: seq, Tick: tick, Type: "agent.memory_added",
		Payload: mustPayload(MemoryAddedPayload{Agent: agent, Text: text, Salience: 3, Subject: -1})}
}

// TestPre042RoundTripByteIdentical (spec 042 T004, data-model invariants): a
// state built from pre-042 events marshals with NONE of the spec-042 JSON keys
// (seq/vec/vec_model on memories, sit_vec* on agents) when the events carry no
// seqs — so a pre-042 snapshot round-trips byte-identically, no FormatVersion
// bump. The 019 precedent test, one spec later.
func TestPre042RoundTripByteIdentical(t *testing.T) {
	s := NewState(42, testMap(42))
	// A pre-042 log: memory events with NO store seq (as a pre-042 snapshot's
	// reduced memories would round-trip — the field absent).
	if err := s.Apply(memoryAddedEvent(0, 100, 0, "Built a fire.")); err != nil {
		t.Fatal(err)
	}
	got := s.Marshal()
	for _, key := range []string{`"seq"`, `"vec"`, `"vec_model"`, `"sit_vec"`, `"sit_vec_model"`, `"sit_vec_tick"`} {
		if bytes.Contains(got, []byte(key)) {
			t.Errorf("pre-042 state leaked the spec-042 key %s:\n%s", key, got)
		}
	}
	// Round-trip: unmarshal + re-marshal is byte-identical (the recovery path).
	s2 := NewState(42, testMap(42))
	if err := json.Unmarshal(got, s2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if again := s2.Marshal(); !bytes.Equal(got, again) {
		t.Errorf("pre-042 snapshot did not round-trip byte-identically:\n%s\n---\n%s", got, again)
	}
}

// TestMemorySeqStampedFromEvent (spec 042 T005, contracts/embedding-events.md
// §3.3): the reducer stamps Memory.Seq from the emitting event's store seq,
// and two replays of the same stream agree byte-for-byte — seq stamping is
// replay-stable because seq is part of the recorded stream.
func TestMemorySeqStampedFromEvent(t *testing.T) {
	log := []store.Event{
		memoryAddedEvent(7, 100, 0, "Foraged by the river."),
		memoryAddedEvent(8, 100, 0, "Saw Birch there."),
		memoryAddedEvent(9, 200, 1, "Chopped wood."),
	}
	replay := func() *State {
		s := NewState(42, testMap(42))
		for _, e := range log {
			if err := s.Apply(e); err != nil {
				t.Fatalf("apply: %v", err)
			}
		}
		return s
	}
	s := replay()
	if got := s.Agents[0].Memories; len(got) != 2 || got[0].Seq != 7 || got[1].Seq != 8 {
		t.Errorf("agent 0 memory seqs = %+v, want 7 then 8", got)
	}
	if got := s.Agents[1].Memories; len(got) != 1 || got[0].Seq != 9 {
		t.Errorf("agent 1 memory seq = %+v, want 9", got)
	}
	if a, b := string(replay().Marshal()), string(replay().Marshal()); a != b {
		t.Errorf("two replays of the same stream diverged:\n%s\n---\n%s", a, b)
	}
}

// TestMemoryEmbeddedReducer (spec 042 T006, contracts/embedding-events.md §1):
// agent.memory_embedded attaches Vec/VecModel to the {agent, mem_seq} memory,
// copy-verbatim; a missing target is a NO-OP (agent died / memory consolidated
// away), and a zero mem_seq never matches pre-042 (seq-less) memories.
func TestMemoryEmbeddedReducer(t *testing.T) {
	s := NewState(42, testMap(42))
	if err := s.Apply(memoryAddedEvent(0, 50, 0, "A seq-less pre-042 memory.")); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(memoryAddedEvent(7, 100, 0, "Foraged by the river.")); err != nil {
		t.Fatal(err)
	}

	embed := func(agent int, memSeq int64, vec []float32) error {
		return s.Apply(store.Event{Seq: 20, Tick: 120, Type: "agent.memory_embedded",
			Payload: mustPayload(MemoryEmbeddedPayload{Agent: agent, MemSeq: memSeq, Vec: vec, Model: "all-minilm"})})
	}
	if err := embed(0, 7, []float32{0.1, 0.2}); err != nil {
		t.Fatalf("embed: %v", err)
	}
	m := s.Agents[0].Memories[1]
	if len(m.Vec) != 2 || m.Vec[0] != 0.1 || m.VecModel != "all-minilm" {
		t.Errorf("memory did not gain the vector verbatim: %+v", m)
	}

	// Absent target: a deliberate no-op, never an error.
	before := string(s.Marshal())
	if err := embed(0, 999, []float32{1}); err != nil {
		t.Fatalf("missing target must no-op, got: %v", err)
	}
	if got := string(s.Marshal()); got != before {
		t.Errorf("no-op companion mutated state:\n%s\n---\n%s", before, got)
	}
	// Zero mem_seq: must not attach to the seq-less pre-042 memory.
	if err := embed(0, 0, []float32{1}); err != nil {
		t.Fatalf("zero mem_seq must no-op, got: %v", err)
	}
	if pre := s.Agents[0].Memories[0]; pre.Vec != nil {
		t.Errorf("zero mem_seq matched a pre-042 memory: %+v", pre)
	}
}

// TestSituationEmbeddedReducer (spec 042 T006): agent.situation_embedded sets
// the agent's rolling SitVec/SitVecModel/SitVecTick; a later event overwrites
// (selection reads the latest at-or-before its tick).
func TestSituationEmbeddedReducer(t *testing.T) {
	s := NewState(42, testMap(42))
	sit := func(tick int64, vec []float32) error {
		return s.Apply(store.Event{Seq: 30, Tick: tick, Type: "agent.situation_embedded",
			Payload: mustPayload(SituationEmbeddedPayload{Agent: 2, Tick: tick, Text: "midday · by the river", Vec: vec, Model: "all-minilm"})})
	}
	if err := sit(100, []float32{0.5}); err != nil {
		t.Fatal(err)
	}
	a := &s.Agents[2]
	if a.SitVecTick != 100 || a.SitVecModel != "all-minilm" || len(a.SitVec) != 1 || a.SitVec[0] != 0.5 {
		t.Errorf("situation vector not stored: %+v", a)
	}
	if err := sit(400, []float32{0.9}); err != nil {
		t.Fatal(err)
	}
	if a.SitVecTick != 400 || a.SitVec[0] != 0.9 {
		t.Errorf("later situation vector did not overwrite: tick=%d vec=%v", a.SitVecTick, a.SitVec)
	}
	// cog.memory_divergence is a reducer no-op (telemetry, cog.* class).
	before := string(s.Marshal())
	if err := s.Apply(store.Event{Seq: 31, Tick: 401, Type: "cog.memory_divergence",
		Payload: mustPayload(map[string]any{"agent": 2, "tick": 401, "mode": "shadow"})}); err != nil {
		t.Fatalf("cog.memory_divergence must be a no-op, got: %v", err)
	}
	if got := string(s.Marshal()); got != before {
		t.Errorf("cog.memory_divergence mutated state")
	}
}
