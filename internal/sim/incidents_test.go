package sim

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Spec 077 incident-kind tests: reducer arms (apply + genesis-replay
// byte-identity, US2 AS-5), the named preconditions, scheduled emission,
// window lapse, and the pre-077 snapshot posture (US2 edge cases).

// applyStamped folds events into s with sequential seqs, failing the test on any
// reducer error — the fixture-side twin of driveTicksSeq's apply.
func applyStamped(t *testing.T, s *State, events []store.Event) {
	t.Helper()
	for i := range events {
		if events[i].Seq == 0 {
			events[i].Seq = int64(i + 1)
		}
		if err := s.Apply(events[i]); err != nil {
			t.Fatalf("apply %s: %v", events[i].Type, err)
		}
		s.Tick = events[i].Tick
	}
}

// TestColdSnapReducerLatchAndReplay (spec 077 FR-010, US2 AS-1/AS-5): the arm
// latches ColdSnapUntil, expiry is a read-time comparison (no end event
// exists to look for), and genesis replay of the recorded event reproduces
// byte-identical state with no scenario armed.
func TestColdSnapReducerLatchAndReplay(t *testing.T) {
	m := testMap(1)
	ev := store.Event{Tick: 100, Type: "sim.cold_snap",
		Payload: mustPayload(ColdSnapPayload{Night: 1, UntilTick: 100 + 8*3600})}

	live := NewState(1, m)
	applyStamped(t, live, []store.Event{ev})
	if live.ColdSnapUntil != 100+8*3600 {
		t.Fatalf("ColdSnapUntil = %d, want %d", live.ColdSnapUntil, 100+8*3600)
	}
	if !coldSnapActive(live, 200) {
		t.Error("snap should hold inside its window")
	}
	if coldSnapActive(live, 100+8*3600) {
		t.Error("snap should expire at its until_tick (read-time, no end event)")
	}

	replayed := NewState(1, m)
	applyStamped(t, replayed, []store.Event{ev})
	if !bytes.Equal(live.Marshal(), replayed.Marshal()) {
		t.Fatal("cold snap replay diverged from live fold")
	}

	// Door: a non-positive until_tick is a malformed fixture, rejected.
	bad := store.Event{Tick: 1, Type: "sim.cold_snap",
		Payload: mustPayload(ColdSnapPayload{Night: 1, UntilTick: 0})}
	if err := NewState(1, m).Apply(bad); err == nil {
		t.Error("zero until_tick applied cleanly, want a door rejection")
	}
}

// TestColdSnapHeartbeatRates (spec 077 FR-010, T009): while the snap holds,
// outdoor night warmth loss runs at the harsher rate through the SAME
// decayNeeds arithmetic; fire warmth still wins; past ColdSnapUntil the
// ambient rate resumes with no end event.
func TestColdSnapHeartbeatRates(t *testing.T) {
	n := Needs{Health: 1000, Food: 500, Rest: 500, Warmth: 500, Morale: 500}
	ambient := decayNeeds(n, false, true, false, false, false)
	snapped := decayNeeds(n, false, true, false, false, true)
	if got := n.Warmth - ambient.Warmth; got != warmthLossCold {
		t.Errorf("ambient night loss = %d, want %d", got, warmthLossCold)
	}
	if got := n.Warmth - snapped.Warmth; got != warmthLossColdSnap {
		t.Errorf("cold-snap night loss = %d, want %d", got, warmthLossColdSnap)
	}
	// Fire warmth beats the snap exactly as it beats ambient cold.
	byFire := decayNeeds(n, false, true, true, false, true)
	if byFire.Warmth <= n.Warmth {
		t.Error("a lit fire should out-warm the cold snap")
	}
	// Day is untouched by the flag (the snap only bites the night arm).
	day := decayNeeds(n, false, false, false, false, true)
	if day.Warmth < n.Warmth {
		t.Error("daytime warmth must not decay under a cold snap")
	}

	// Through the heartbeat: an outdoor sleeper under a live snap loses at the
	// harsher rate, and at the same tick arithmetic once the snap expires.
	m := testMap(1)
	s := NewState(1, m)
	isolateAgents(s)
	a := &s.Agents[0]
	a.Dead = false
	a.Needs = Needs{Health: 1000, Food: 500, Rest: 500, Warmth: 500, Morale: 500}
	s.Night = true
	s.Tick = clock.TickAt(1, 23, 0, 0) - 1 // next tick is on the %60 heartbeat
	s.ColdSnapUntil = s.Tick + 7200
	next := s.Tick + 1
	for _, e := range stepEvents(s, m, next) {
		if e.Type != "agent.needs_changed" {
			continue
		}
		var p NeedsPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			t.Fatal(err)
		}
		if got := a.Needs.Warmth - p.Warmth; got != warmthLossColdSnap {
			t.Errorf("heartbeat under snap lost %d warmth, want %d", got, warmthLossColdSnap)
		}
	}
	// Expired snap (read-time): the same boundary, ambient rate.
	s2 := NewState(1, m)
	isolateAgents(s2)
	a2 := &s2.Agents[0]
	a2.Dead = false
	a2.Needs = Needs{Health: 1000, Food: 500, Rest: 500, Warmth: 500, Morale: 500}
	s2.Night = true
	s2.Tick = clock.TickAt(1, 23, 0, 0) - 1
	s2.ColdSnapUntil = s2.Tick - 10 // lapsed
	for _, e := range stepEvents(s2, m, s2.Tick+1) {
		if e.Type != "agent.needs_changed" {
			continue
		}
		var p NeedsPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			t.Fatal(err)
		}
		if got := a2.Needs.Warmth - p.Warmth; got != warmthLossCold {
			t.Errorf("heartbeat past the snap lost %d warmth, want ambient %d", got, warmthLossCold)
		}
	}
}

// forageTileIn finds an unharvested forage tile on m, failing if none exists.
func forageTileIn(t *testing.T, m *worldmap.Map, s *State) (int, int) {
	t.Helper()
	for y := 0; y < m.H; y++ {
		for x := 0; x < m.W; x++ {
			if effectiveKind(m, s, x, y) == worldmap.Forage {
				return x, y
			}
		}
	}
	t.Fatal("no forage tile on the test map")
	return 0, 0
}

// TestForageBlightReducerOverlayAndReplay (spec 077 FR-011, US2 AS-2/AS-5):
// the arm marks each recorded tile with the EXISTING Harvest overlay at the
// far deadline — exactly what heavy picking produces — re-applies
// idempotently over a tile already harvested, and replays byte-identically.
func TestForageBlightReducerOverlayAndReplay(t *testing.T) {
	m := testMap(1)
	s := NewState(1, m)
	fx, fy := forageTileIn(t, m, s)
	tiles := blightableTiles(m, s, fx, fy, 2)
	if len(tiles) == 0 {
		t.Fatal("blightableTiles empty on a forage tile")
	}
	regrow := int64(100 + blightRegrowTicks)
	ev := store.Event{Tick: 100, Type: "sim.forage_blighted",
		Payload: mustPayload(ForageBlightedPayload{X: fx, Y: fy, Radius: 2, Tiles: tiles, RegrowTick: regrow})}
	applyStamped(t, s, []store.Event{ev})

	for _, tile := range tiles {
		found := false
		for _, h := range s.Harvested {
			if h.X == tile.X && h.Y == tile.Y && h.Regrow == regrow {
				found = true
			}
		}
		if !found {
			t.Errorf("tile (%d,%d) missing its Harvest overlay at the blight deadline", tile.X, tile.Y)
		}
		if effectiveKind(m, s, tile.X, tile.Y) != worldmap.Grass {
			t.Errorf("blighted tile (%d,%d) should read barren (Grass) through the overlay", tile.X, tile.Y)
		}
	}
	// The predicate is the latch: after the firing the patch has nothing left.
	if left := blightableTiles(m, s, fx, fy, 2); len(left) != 0 {
		t.Errorf("blightableTiles still returns %d tiles after the blight — the overlay latch failed", len(left))
	}

	// Idempotent re-apply (snapshot + replay overlap): a second application
	// adds no duplicate overlays.
	before := len(s.Harvested)
	if err := s.Apply(ev); err != nil {
		t.Fatalf("re-apply: %v", err)
	}
	if len(s.Harvested) != before {
		t.Errorf("re-apply grew Harvested %d → %d, want idempotent", before, len(s.Harvested))
	}

	// Genesis replay byte-identity.
	replayed := NewState(1, m)
	applyStamped(t, replayed, []store.Event{ev})
	if !bytes.Equal(s.Marshal(), replayed.Marshal()) {
		t.Fatal("blight replay diverged from live fold")
	}

	// Door: an empty tile list is a malformed fixture.
	bad := store.Event{Tick: 1, Type: "sim.forage_blighted",
		Payload: mustPayload(ForageBlightedPayload{X: fx, Y: fy, Radius: 2, RegrowTick: 10})}
	if err := NewState(1, m).Apply(bad); err == nil {
		t.Error("empty tile list applied cleanly, want a door rejection")
	}
}

// TestBlightableTilesRowMajorOrder pins the deterministic walk order (spec
// 077 FR-011): tiles enumerate row-major (y outer, x inner) over the patch.
func TestBlightableTilesRowMajorOrder(t *testing.T) {
	m := testMap(1)
	s := NewState(1, m)
	fx, fy := forageTileIn(t, m, s)
	tiles := blightableTiles(m, s, fx, fy, 8)
	for i := 1; i < len(tiles); i++ {
		a, b := tiles[i-1], tiles[i]
		if a.Y > b.Y || (a.Y == b.Y && a.X >= b.X) {
			t.Fatalf("tiles not in row-major order: %v before %v", a, b)
		}
	}
}

// TestScheduledColdSnapFiresOnceWithLatch (spec 077 US2 AS-1, T011): an armed
// cold_snap entry fires at its authored tick, latches, never re-fires inside
// its window (the latch is the state), and lapses silently past its window
// after a time snap.
func TestScheduledColdSnapFiresOnceWithLatch(t *testing.T) {
	m := testMap(9)
	def := ExerciseDefinition{ID: "first-night", Stage: "stage-1", Seed: 9,
		Schedule: []IncidentScheduleEntry{{Kind: IncidentColdSnap, Day: 1, Time: "22:00", Hours: 6}}}
	s := NewState(9, m)
	if err := s.ArmScenario(def); err != nil {
		t.Fatal(err)
	}
	for i := range s.Agents {
		s.Agents[i].Dead = true // pure schedule frame
	}
	authored := clock.TickAt(1, 22, 0, 0)
	s.Tick = authored - 10
	var snaps []int64
	for s.Tick < authored+600 {
		next := s.Tick + 1
		for _, e := range stepEvents(s, m, next) {
			if e.Type == "sim.cold_snap" {
				snaps = append(snaps, e.Tick)
				var p ColdSnapPayload
				if err := json.Unmarshal(e.Payload, &p); err != nil {
					t.Fatal(err)
				}
				if p.UntilTick != authored+6*3600 {
					t.Errorf("until_tick = %d, want authored end %d", p.UntilTick, authored+6*3600)
				}
			}
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		s.Tick = next
	}
	if len(snaps) != 1 || snaps[0] != authored {
		t.Fatalf("cold snap fired at %v, want exactly once at %d", snaps, authored)
	}

	// Precondition-failed class: a snap already active at the authored tick
	// skips silently, never retried (US2 AS-2 class).
	s2 := NewState(9, m)
	if err := s2.ArmScenario(def); err != nil {
		t.Fatal(err)
	}
	for i := range s2.Agents {
		s2.Agents[i].Dead = true
	}
	s2.Tick = authored - 1
	s2.ColdSnapUntil = authored + 20 // an active snap
	for i := int64(0); i < 100; i++ {
		next := s2.Tick + 1
		for _, e := range stepEvents(s2, m, next) {
			if e.Type == "sim.cold_snap" {
				// Once the staged snap expires (authored+20) the entry is
				// STILL within its window [authored, authored+6h) — the
				// window is the snap's own end, so a late fire is legal
				// only while the window stands and no snap holds. Assert
				// it never fires while the staged snap holds.
				if next < authored+20 {
					t.Fatalf("cold snap fired at %d despite an active snap", next)
				}
			}
			if err := s2.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		s2.Tick = next
	}

	// Window lapse: a time snap past the whole window skips it silently.
	s3 := NewState(9, m)
	if err := s3.ArmScenario(def); err != nil {
		t.Fatal(err)
	}
	for i := range s3.Agents {
		s3.Agents[i].Dead = true
	}
	s3.Tick = authored + 6*3600 // at windowEnd — lapsed
	for i := int64(0); i < 200; i++ {
		next := s3.Tick + 1
		for _, e := range stepEvents(s3, m, next) {
			if e.Type == "sim.cold_snap" {
				t.Fatalf("lapsed cold snap fired at %d", next)
			}
			if err := s3.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		s3.Tick = next
	}
}

// TestScheduledBlightSkipsExhaustedPatch (spec 077 US2 AS-2): a blight entry
// over a patch with no unharvested forage fails its precondition and skips
// silently — the schedule proposes, the reducer-valid world disposes.
func TestScheduledBlightSkipsExhaustedPatch(t *testing.T) {
	m := testMap(9)
	probe := NewState(9, m)
	fx, fy := forageTileIn(t, m, probe)
	def := ExerciseDefinition{ID: "first-night", Stage: "stage-1", Seed: 9,
		Schedule: []IncidentScheduleEntry{{Kind: IncidentForageBlight, Day: 1, Time: "08:00", X: fx, Y: fy, Radius: 3}}}
	s := NewState(9, m)
	if err := s.ArmScenario(def); err != nil {
		t.Fatal(err)
	}
	for i := range s.Agents {
		s.Agents[i].Dead = true
	}
	// Exhaust the patch by hand: every blightable tile already harvested.
	for _, tile := range blightableTiles(m, s, fx, fy, 3) {
		s.Harvested = append(s.Harvested, Harvest{X: tile.X, Y: tile.Y, Regrow: 1 << 40})
	}
	authored := clock.TickAt(1, 8, 0, 0)
	s.Tick = authored - 10
	for s.Tick < authored+600 {
		next := s.Tick + 1
		for _, e := range stepEvents(s, m, next) {
			if e.Type == "sim.forage_blighted" {
				t.Fatalf("blight fired at %d over an exhausted patch", next)
			}
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		s.Tick = next
	}
}

// TestPre077SnapshotRoundTripByteIdentical (spec 077, data-model §3): a state
// carrying none of the new facts marshals with NONE of the new keys, and
// unmarshal + re-marshal is byte-identical — no format_version bump, the
// spec-072 CharterCustom precedent.
func TestPre077SnapshotRoundTripByteIdentical(t *testing.T) {
	s := NewState(42, testMap(42))
	got := s.Marshal()
	for _, key := range []string{`"cold_snap_until"`, `"stranger"`, `"stranger_takes"`,
		`"charter_observed_seq"`, `"charter_observed_tick"`,
		`"skills_fingerprint"`, `"skills_observed_seq"`, `"skills_observed_tick"`} {
		if bytes.Contains(got, []byte(key)) {
			t.Errorf("pre-077 state leaked the spec-077 key %s", key)
		}
	}
	s2 := NewState(42, testMap(42))
	if err := json.Unmarshal(got, s2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if again := s2.Marshal(); !bytes.Equal(got, again) {
		t.Errorf("pre-077 snapshot did not round-trip byte-identically")
	}
}

// TestCharterObservedStampsCoordinates (spec 077 FR-005): the charter arm
// additionally persists the envelope's (seq, tick) — the PlacedSeq precedent
// — and CharterEvidenceFromState re-locates the observation from state alone.
func TestCharterObservedStampsCoordinates(t *testing.T) {
	s := NewState(1, testMap(1))
	ev := theLawCharterEvent(500, "bbbb33334444", false)
	ev.Seq = 77
	if err := s.Apply(ev); err != nil {
		t.Fatal(err)
	}
	if s.CharterObservedSeq != 77 || s.CharterObservedTick != 500 {
		t.Fatalf("coordinates = (%d,%d), want (77,500)", s.CharterObservedSeq, s.CharterObservedTick)
	}
	ref, ok := CharterEvidenceFromState(s)
	if !ok {
		t.Fatal("CharterEvidenceFromState found no coordinates")
	}
	want := EvidenceRef{Type: "guardian.charter_observed", Seq: 77, Tick: 500, Custom: true}
	if ref != want {
		t.Errorf("evidence = %+v, want %+v", ref, want)
	}

	// The default-charter inverse: Custom rides CharterCustom, never freehand.
	if err := s.Apply(theLawCharterEvent(600, "cccc55556666", true)); err != nil {
		t.Fatal(err)
	}
	if ref, _ = CharterEvidenceFromState(s); ref.Custom {
		t.Error("a default-charter observation must derive Custom=false")
	}

	// Pre-077 honesty: no stamped coordinates ⇒ no evidence entry, ever.
	if _, ok := CharterEvidenceFromState(NewState(1, testMap(1))); ok {
		t.Error("a state with no observation coordinates produced evidence")
	}
}

// TestSkillsObservedReducerArm (spec 077 FR-006): the arm persists
// fingerprint + envelope coordinates; the door refuses an empty fingerprint
// or an empty name list (absence is not an observation); and
// SkillsObservedEvidence derives Custom: true by construction.
func TestSkillsObservedReducerArm(t *testing.T) {
	s := NewState(1, testMap(1))
	ev := store.Event{Tick: 900, Seq: 41, Type: "guardian.skills_observed",
		Payload: mustPayload(SkillsObservedPayload{Fingerprint: "ab12cd34ef56", Names: []string{"10-watch.md"}})}
	if err := s.Apply(ev); err != nil {
		t.Fatal(err)
	}
	if s.SkillsFingerprint != "ab12cd34ef56" || s.SkillsObservedSeq != 41 || s.SkillsObservedTick != 900 {
		t.Fatalf("skills state = (%q,%d,%d), want (ab12cd34ef56,41,900)",
			s.SkillsFingerprint, s.SkillsObservedSeq, s.SkillsObservedTick)
	}
	ref, ok := SkillsObservedEvidence(s)
	if !ok {
		t.Fatal("SkillsObservedEvidence found no coordinates")
	}
	want := EvidenceRef{Type: "guardian.skills_observed", Seq: 41, Tick: 900, Custom: true}
	if ref != want {
		t.Errorf("evidence = %+v, want %+v", ref, want)
	}
	if _, ok := SkillsObservedEvidence(NewState(1, testMap(1))); ok {
		t.Error("a state with no skills observation produced evidence")
	}

	// Door rejections.
	for name, payload := range map[string]SkillsObservedPayload{
		"empty fingerprint": {Names: []string{"a.md"}},
		"empty name list":   {Fingerprint: "ab12cd34ef56"},
	} {
		bad := store.Event{Tick: 1, Type: "guardian.skills_observed", Payload: mustPayload(payload)}
		if err := NewState(1, testMap(1)).Apply(bad); err == nil {
			t.Errorf("%s applied cleanly, want a door rejection", name)
		}
	}

	// Replay byte-identity (the recorded event is the only persistence).
	replayed := NewState(1, testMap(1))
	applyStamped(t, replayed, []store.Event{ev})
	live := NewState(1, testMap(1))
	applyStamped(t, live, []store.Event{ev})
	if !bytes.Equal(live.Marshal(), replayed.Marshal()) {
		t.Fatal("skills observation replay diverged")
	}
}

// TestSnapShiftsIncidentAnchors (spec 077 FR-015, plan D5): a time snap
// SHIFTs the snap's remaining window and the stranger's cadence anchors,
// KEEPs the take ledger's historical ticks and the observation coordinates.
func TestSnapShiftsIncidentAnchors(t *testing.T) {
	s := NewState(1, testMap(1))
	s.Tick = 1000
	s.ColdSnapUntil = 5000
	s.Stranger = &Stranger{X: 3, Y: 4, Night: 1, LastMove: 900, LastTake: 800}
	s.StrangerTakes = []StrangerTake{{Tick: 700, X: 3, Y: 4, Kind: "wood", N: 2}}
	s.CharterObservedSeq, s.CharterObservedTick = 9, 600
	s.SkillsObservedSeq, s.SkillsObservedTick = 11, 650

	rebaseTicks(s, 10_000)

	if s.ColdSnapUntil != 15_000 {
		t.Errorf("ColdSnapUntil = %d, want shifted 15000", s.ColdSnapUntil)
	}
	if s.Stranger.LastMove != 10_900 || s.Stranger.LastTake != 10_800 {
		t.Errorf("stranger anchors = (%d,%d), want shifted (10900,10800)", s.Stranger.LastMove, s.Stranger.LastTake)
	}
	if s.Stranger.Night != 1 {
		t.Errorf("Stranger.Night = %d, want KEEP 1", s.Stranger.Night)
	}
	if s.StrangerTakes[0].Tick != 700 {
		t.Errorf("StrangerTakes[0].Tick = %d, want KEEP 700", s.StrangerTakes[0].Tick)
	}
	if s.CharterObservedTick != 600 || s.SkillsObservedTick != 650 {
		t.Errorf("observation ticks = (%d,%d), want KEEP (600,650)", s.CharterObservedTick, s.SkillsObservedTick)
	}

	// The zero sentinels stay zero (no snap ever / fresh arrival).
	z := NewState(1, testMap(1))
	z.Stranger = &Stranger{X: 1, Y: 1, Night: 2}
	rebaseTicks(z, 10_000)
	if z.ColdSnapUntil != 0 || z.Stranger.LastMove != 0 || z.Stranger.LastTake != 0 {
		t.Error("zero sentinels were shifted")
	}
}
