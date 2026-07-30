package sim

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Spec 104 test surface: the derived-progress engine's unit contracts
// (ordering, idempotence, monotonicity, legacy inertia — T002), the sighting
// equivalence harness (ruling 2 / FR-004 — T006), the needs thinning proofs
// (double-decay guard, K=1 byte parity, K=10 minute equality, crossing
// latency — T007/T008/T009), gru derivation + beat ordering (T010), and the
// executor emission rewire (T005).

// flatMap is an all-grass w×h map — a controlled stage for scripted walks.
func flatMap(w, h int) *worldmap.Map {
	tiles := make([]worldmap.TileKind, w*h)
	for i := range tiles {
		tiles[i] = worldmap.Grass
	}
	return &worldmap.Map{W: w, H: h, Tiles: tiles}
}

// islandMap is all water except one grass tile per island point — agents are
// placed on the islands and can never move, forage, or reach each other, so
// needs decay is the only physics (the needs-equivalence stages).
func islandMap(w, h int, islands ...worldmap.Point) *worldmap.Map {
	tiles := make([]worldmap.TileKind, w*h)
	for i := range tiles {
		tiles[i] = worldmap.Water
	}
	m := &worldmap.Map{W: w, H: h, Tiles: tiles}
	for _, p := range islands {
		tiles[p.Y*w+p.X] = worldmap.Grass
	}
	return m
}

// tuningEventK builds a sim.tuning_applied event carrying the default dial
// set with needs_checkpoint_minutes = k (k=0 keeps the LEGACY regime) and the
// gru emergence roll silenced, so paired worlds never diverge on predator
// luck.
func tuningEventK(tick, k int64) store.Event {
	tun := defaultTuning()
	tun.NeedsCheckpointMinutes = k
	tun.GruEmergePerMille = 0
	return NewTuningEvent(tick, tun)
}

// coalescedState is a fresh world with the spec-104 regime pinned at genesis
// (the new-world posture): NewState + the K-carrying tuning event at tick 0.
func coalescedState(t *testing.T, seed uint64, m *worldmap.Map, k int64) *State {
	t.Helper()
	s := NewState(seed, m)
	if err := s.Apply(tuningEventK(0, k)); err != nil {
		t.Fatalf("apply tuning: %v", err)
	}
	return s
}

// --- T002: engine unit contracts ------------------------------------------

// TestAdvanceLegacyStateInert is the old-log posture: on a world whose
// regime marker is off (nil tuning — every pre-104 world), AdvanceTo
// contributes NOTHING, ever — needs, gru, everything byte-identical however
// far it is driven. This is the structural half of FR-002 (the recorded
// pre-086 fixture hash, TestPre086ReplayByteIdentity, is the cross-build
// half).
func TestAdvanceLegacyStateInert(t *testing.T) {
	m := testMap(42)
	s := NewState(42, m)
	s.Gru = &Gru{X: 3, Y: 3}
	s.Night = true
	before := s.Marshal()
	s.AdvanceTo(500000)
	if !bytes.Equal(before, s.Marshal()) {
		t.Fatal("AdvanceTo mutated a legacy (regime-off) state")
	}
}

// TestAdvanceIdempotentMonotone: re-advancing to the same or a lower target
// changes nothing; each scheduled item runs exactly once.
func TestAdvanceIdempotentMonotone(t *testing.T) {
	m := flatMap(32, 32)
	s := coalescedState(t, 42, m, 10)
	isolateAgents(s)
	a := reviveAt(s, 5, 5)
	a.Intent = &Intent{Goal: "wander", TargetX: 12, TargetY: 5}
	if err := s.Apply(store.Event{Tick: 3, Type: "agent.path_started", Payload: mustPayload(PathStartedPayload{
		Agent: Ref(0), Path: []Point{{X: 6, Y: 5}, {X: 7, Y: 5}, {X: 8, Y: 5}},
		MoveEvery: moveEveryTicks, Phase: 0,
	})}); err != nil {
		t.Fatalf("apply path_started: %v", err)
	}
	s.AdvanceTo(9) // beats at 5 (tile 6,5) — steps strictly after tick 3
	first := s.Marshal()
	if a.X != 6 || a.Y != 5 {
		t.Fatalf("expected one step to (6,5), at (%d,%d)", a.X, a.Y)
	}
	s.AdvanceTo(9)
	if !bytes.Equal(first, s.Marshal()) {
		t.Fatal("re-advancing to the same target changed state")
	}
	s.AdvanceTo(4) // lower target: monotone no-op
	if !bytes.Equal(first, s.Marshal()) {
		t.Fatal("advancing to a lower target changed state")
	}
	s.AdvanceTo(11) // beat at 10 → (7,5)
	if a.X != 7 {
		t.Fatalf("expected second step to (7,5), at (%d,%d)", a.X, a.Y)
	}
	if a.Path == nil || a.Path.Next != 2 || a.Path.Done != 10 {
		t.Fatalf("segment watermark wrong: %+v", a.Path)
	}
	s.AdvanceTo(16) // beat at 15 → arrival tile (8,5); segment retires
	if a.X != 8 || a.Path != nil {
		t.Fatalf("expected arrival at (8,5) with segment retired, at (%d,%d) seg=%v", a.X, a.Y, a.Path)
	}
}

// TestAdvanceNeedsBeforeSteps pins the within-tick family order (research.md
// §2): the needs minute at tick u reads the walker's PRE-step position — a
// step at the same tick lands after the decay, so warmth gained from a fire
// the step walks into arrives one minute later, deterministically.
func TestAdvanceNeedsBeforeSteps(t *testing.T) {
	m := flatMap(32, 32)
	s := coalescedState(t, 42, m, 10)
	isolateAgents(s)
	a := reviveAt(s, 17, 5) // outside fireWarmRadius(2) of the fire at (20,5)
	a.Needs.Warmth = 500
	s.Structures = append(s.Structures, Structure{Kind: "fire", X: 20, Y: 5, FuelUntil: 1 << 40})
	// One step scheduled exactly at the minute tick 60: onto (18,5), still
	// outside the radius — then a second at 65 onto (19,5), inside it.
	if err := s.Apply(store.Event{Tick: 58, Type: "agent.path_started", Payload: mustPayload(PathStartedPayload{
		Agent: Ref(0), Path: []Point{{X: 18, Y: 5}, {X: 19, Y: 5}}, MoveEvery: moveEveryTicks, Phase: 0,
	})}); err != nil {
		t.Fatalf("apply: %v", err)
	}
	s.AdvanceTo(61) // minute 60 decays (pre-step position 17,5: day, no fire → +warmthGainDay), then the step fires
	if a.X != 18 {
		t.Fatalf("step at 60 did not land: at (%d,%d)", a.X, a.Y)
	}
	if got := a.Needs.Warmth; got != 500+warmthGainDay {
		t.Fatalf("minute 60 should decay from the pre-step tile (day gain %d), got warmth %d", warmthGainDay, got)
	}
	s.AdvanceTo(121) // step at 65 → (19,5), inside the radius; minute 120 warms by the fire
	if got := a.Needs.Warmth; got != 500+warmthGainDay+warmthGainFire {
		t.Fatalf("minute 120 should warm at the fire (+%d), got warmth %d", warmthGainFire, got)
	}
}

// --- T006: the sighting equivalence harness (ruling 2 / FR-004) -----------

// walkScript is one scripted walk: who, from where, along which tiles,
// departing at which tick.
type walkScript struct {
	agent int
	from  Point
	path  []Point
	start int64
}

func phaseOf(agent int) int64 { return (int64(agent) * 3) % moveEveryTicks }

// scheduleSteps mirrors the beat rule tick by tick (the harness's
// independent computation of when today's per-step emitter would have
// stepped): base slot every MoveEvery ticks, plus the path slot while the
// CURRENT tile is paved. paved is the scenario's static path-structure set.
func scheduleSteps(w walkScript, paved map[Point]bool) []struct {
	tick int64
	tile Point
} {
	var out []struct {
		tick int64
		tile Point
	}
	cur := w.from
	cursor := 0
	for t := w.start + 1; cursor < len(w.path); t++ {
		ph := (t + phaseOf(w.agent)) % moveEveryTicks
		if ph == 0 || (ph == pathSpeedSlot && paved[cur]) {
			cur = w.path[cursor]
			out = append(out, struct {
				tick int64
				tile Point
			}{t, cur})
			cursor++
		}
	}
	return out
}

// equivalenceStage builds the paired scenario base: a coalesced flat world
// (K=10 so no needs events are needed — derived decay runs identically on
// both sides) with every agent dead except the script participants.
func equivalenceStage(t *testing.T, participants map[int]Point, paved []Point) func() *State {
	t.Helper()
	return func() *State {
		m := flatMap(32, 32)
		s := coalescedState(t, 42, m, 10)
		isolateAgents(s)
		for idx, p := range participants {
			a := &s.Agents[idx]
			a.Dead = false
			a.X, a.Y = p.X, p.Y
			a.Needs = Needs{Health: 1000, Food: 900, Rest: 900, Warmth: 900, Morale: 900}
		}
		for _, p := range paved {
			s.Structures = append(s.Structures, Structure{Kind: "path", X: p.X, Y: p.Y})
		}
		return s
	}
}

// stripSegments marshals a state with its in-flight segments removed — the
// coalesced representation itself is the ONE intended byte difference
// against a per-step log mid-walk; everything semantic (positions, needs,
// explored bitmaps, peer sightings with Seen ticks, anchors, watermarks)
// must match byte-for-byte.
func stripSegments(s *State) []byte {
	saved := make([]*PathSegment, len(s.Agents))
	for i := range s.Agents {
		saved[i] = s.Agents[i].Path
		s.Agents[i].Path = nil
	}
	b := s.Marshal()
	for i := range s.Agents {
		s.Agents[i].Path = saved[i]
	}
	return b
}

// sweepEquivalence folds the per-step log (A) and the coalesced log (B) over
// identical bases and asserts semantic byte equality at EVERY tick boundary
// (fold events ≤ t, then AdvanceTo(t+1) — the state any tick-(t+1) observer
// sees; research.md §1), including each agent's canonical mental-map bytes.
func sweepEquivalence(t *testing.T, base func() *State, logA, logB []store.Event, from, to int64) {
	t.Helper()
	a, b := base(), base()
	ai, bi := 0, 0
	for tick := from; tick <= to; tick++ {
		for ai < len(logA) && logA[ai].Tick == tick {
			if err := a.Apply(logA[ai]); err != nil {
				t.Fatalf("A apply %s@%d: %v", logA[ai].Type, tick, err)
			}
			ai++
		}
		for bi < len(logB) && logB[bi].Tick == tick {
			if err := b.Apply(logB[bi]); err != nil {
				t.Fatalf("B apply %s@%d: %v", logB[bi].Type, tick, err)
			}
			bi++
		}
		a.Tick, b.Tick = tick, tick
		a.AdvanceTo(tick + 1)
		b.AdvanceTo(tick + 1)
		ab, bb := stripSegments(a), stripSegments(b)
		if !bytes.Equal(ab, bb) {
			t.Fatalf("state bytes diverge at tick boundary %d:\nA: %s\nB: %s", tick, ab, bb)
		}
		for i := range a.Agents {
			am, bm := a.Agents[i].Map, b.Agents[i].Map
			if (am == nil) != (bm == nil) {
				t.Fatalf("agent %d map presence diverges at tick %d", i, tick)
			}
			if am == nil {
				continue
			}
			amb, _ := json.Marshal(am)
			bmb, _ := json.Marshal(bm)
			if !bytes.Equal(amb, bmb) {
				t.Fatalf("agent %d mental-map bytes diverge at tick %d:\nA: %s\nB: %s", i, tick, amb, bmb)
			}
		}
	}
	if ai != len(logA) || bi != len(logB) {
		t.Fatalf("unconsumed log rows (A %d/%d, B %d/%d) — sweep window too short", ai, len(logA), bi, len(logB))
	}
}

// pairedLogs renders one walk script both ways: A as today's per-step
// agent.moved rows at the exact ticks the beat rule fires, B as the single
// agent.path_started. truncAt > 0 cuts the walk at that tick: A's step rows
// stop strictly before it, B gains an agent.path_truncated carrying the
// stop position.
func pairedLogs(w walkScript, paved map[Point]bool, truncAt int64) (logA, logB []store.Event) {
	logB = append(logB, store.Event{Tick: w.start, Type: "agent.path_started", Payload: mustPayload(PathStartedPayload{
		Agent: Ref(w.agent), Path: w.path, MoveEvery: moveEveryTicks, Phase: phaseOf(w.agent),
	})})
	stop := w.from
	for _, st := range scheduleSteps(w, paved) {
		if truncAt > 0 && st.tick >= truncAt {
			break
		}
		stop = st.tile
		logA = append(logA, store.Event{Tick: st.tick, Type: "agent.moved", Payload: mustPayload(AgentMovedPayload{
			Agent: Ref(w.agent), X: st.tile.X, Y: st.tile.Y,
		})})
	}
	if truncAt > 0 {
		logB = append(logB, store.Event{Tick: truncAt, Type: "agent.path_truncated", Payload: mustPayload(PathTruncatedPayload{
			Agent: Ref(w.agent), X: stop.X, Y: stop.Y,
		})})
	}
	return logA, logB
}

func mergeLogs(logs ...[]store.Event) []store.Event {
	var out []store.Event
	for _, l := range logs {
		out = append(out, l...)
	}
	// stable sort by tick, preserving per-log order within a tick (events
	// precede nothing here — derived items always run after the whole tick)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].Tick < out[j-1].Tick; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func line(x0, x1, y int) []Point {
	var p []Point
	step := 1
	if x1 < x0 {
		step = -1
	}
	for x := x0 + step; ; x += step {
		p = append(p, Point{X: x, Y: y})
		if x == x1 {
			return p
		}
	}
}

// TestEquivalenceSingleWalk: one agent, straight walk — explored bitmap
// growth per step, byte-identical.
func TestEquivalenceSingleWalk(t *testing.T) {
	w := walkScript{agent: 0, from: Point{X: 5, Y: 5}, path: line(5, 15, 5), start: 3}
	base := equivalenceStage(t, map[int]Point{0: w.from}, nil)
	logA, logB := pairedLogs(w, nil, 0)
	sweepEquivalence(t, base, logA, logB, 0, 90)
}

// TestEquivalenceCrossingWalks: two agents walking through each other's
// witness radius on interleaved beats — MUTUAL peer sightings with exact
// Seen ticks (the hard half of ruling 2).
func TestEquivalenceCrossingWalks(t *testing.T) {
	w0 := walkScript{agent: 0, from: Point{X: 5, Y: 10}, path: line(5, 15, 10), start: 2}
	w1 := walkScript{agent: 1, from: Point{X: 15, Y: 12}, path: line(15, 5, 12), start: 2}
	base := equivalenceStage(t, map[int]Point{0: w0.from, 1: w1.from}, nil)
	a0, b0 := pairedLogs(w0, nil, 0)
	a1, b1 := pairedLogs(w1, nil, 0)
	sweepEquivalence(t, base, mergeLogs(a0, a1), mergeLogs(b0, b1), 0, 90)
}

// TestEquivalencePathSpeedup: the spec-032 2x rule — paved tiles fire the
// extra pathSpeedSlot beat, evaluated against the tile stepped FROM.
func TestEquivalencePathSpeedup(t *testing.T) {
	paved := map[Point]bool{}
	var pavedList []Point
	for x := 5; x <= 10; x++ {
		p := Point{X: x, Y: 5}
		paved[p] = true
		pavedList = append(pavedList, p)
	}
	w := walkScript{agent: 0, from: Point{X: 5, Y: 5}, path: line(5, 14, 5), start: 3}
	base := equivalenceStage(t, map[int]Point{0: w.from}, pavedList)
	logA, logB := pairedLogs(w, paved, 0)
	sweepEquivalence(t, base, logA, logB, 0, 80)
}

// TestEquivalenceMidWalkTruncation: the walk is cut at a mid-walk tick — A's
// rows just stop, B records agent.path_truncated with the actual position.
func TestEquivalenceMidWalkTruncation(t *testing.T) {
	w := walkScript{agent: 0, from: Point{X: 5, Y: 5}, path: line(5, 15, 5), start: 3}
	base := equivalenceStage(t, map[int]Point{0: w.from}, nil)
	logA, logB := pairedLogs(w, nil, 27)
	sweepEquivalence(t, base, logA, logB, 0, 90)
}

// TestEquivalenceSleeperWaker: a bystander sleeps beside the walk and wakes
// mid-walk — sleepers record no sightings, the waker's agent.woke sighting
// pass and the walker's subsequent steps must interleave identically. The
// woke row precedes any same-tick step in A (events before derived items —
// the within-tick contract this harness enforces).
func TestEquivalenceSleeperWaker(t *testing.T) {
	w := walkScript{agent: 0, from: Point{X: 5, Y: 5}, path: line(5, 15, 5), start: 3}
	base := func() *State {
		s := equivalenceStage(t, map[int]Point{0: w.from, 2: {X: 10, Y: 6}}, nil)()
		s.Agents[2].Asleep = true
		return s
	}
	woke := store.Event{Tick: 25, Type: "agent.woke", Payload: mustPayload(AgentPayload{Agent: Ref(2)})}
	logA, logB := pairedLogs(w, nil, 0)
	sweepEquivalence(t, base, mergeLogs([]store.Event{woke}, logA), mergeLogs([]store.Event{woke}, logB), 0, 90)
}

// TestEquivalencePauseMidWalk: clock.paused mid-walk — B's pause arm
// truncates the segment, A's rows stop at the same boundary; the resume
// leaves both standing on the same tile with identical maps.
func TestEquivalencePauseMidWalk(t *testing.T) {
	w := walkScript{agent: 0, from: Point{X: 5, Y: 5}, path: line(5, 15, 5), start: 3}
	base := equivalenceStage(t, map[int]Point{0: w.from}, nil)
	pause := store.Event{Tick: 27, Type: "clock.paused", Payload: mustPayload(struct{}{})}
	resume := store.Event{Tick: 27, Type: "clock.resumed", Payload: mustPayload(struct{}{})}
	logA, _ := pairedLogs(w, nil, 27) // per-step rows stop at the pause
	_, logB := pairedLogs(w, nil, 0)  // segment installed; the pause arm truncates it
	sweepEquivalence(t, base,
		mergeLogs(logA, []store.Event{pause, resume}),
		mergeLogs(logB, []store.Event{pause, resume}), 0, 90)
}

// --- T008/T009: needs thinning proofs --------------------------------------

// needsIslands is the stationary stage: every agent imprisoned on a
// one-tile island (nothing to walk to, forage, or reach), so needs decay is
// the only physics and paired worlds cannot diverge on movement.
func needsIslands() (*worldmap.Map, []worldmap.Point) {
	pts := []worldmap.Point{
		{X: 2, Y: 2}, {X: 6, Y: 2}, {X: 10, Y: 2}, {X: 14, Y: 2},
		{X: 2, Y: 6}, {X: 6, Y: 6}, {X: 10, Y: 6}, {X: 14, Y: 6},
	}
	return islandMap(20, 10, pts...), pts
}

// needsEvents filters a driven log to the agent.needs_changed rows.
func needsEvents(log []store.Event) []store.Event {
	var out []store.Event
	for _, e := range log {
		if e.Type == "agent.needs_changed" {
			out = append(out, e)
		}
	}
	return out
}

// TestNeedsK1MatchesLegacyEmission is the escape hatch (FR-008): a K=1
// regime world emits the SAME agent.needs_changed stream — ticks and payload
// bytes — as a legacy world of the same seed. (The K=0-carrying tuning event
// keeps the control world legacy while pinning identical companion dials.)
func TestNeedsK1MatchesLegacyEmission(t *testing.T) {
	m, _ := needsIslands()
	legacy := NewState(7, m)
	if err := legacy.Apply(tuningEventK(0, 0)); err != nil {
		t.Fatal(err)
	}
	k1 := NewState(7, m)
	if err := k1.Apply(tuningEventK(0, 1)); err != nil {
		t.Fatal(err)
	}
	if legacy.AmbientCoalescing() || !k1.AmbientCoalescing() {
		t.Fatal("regime markers wrong way round")
	}
	logL := needsEvents(driveTicks(t, legacy, m, 7200, nil))
	logK := needsEvents(driveTicks(t, k1, m, 7200, nil))
	if len(logL) == 0 || len(logL) != len(logK) {
		t.Fatalf("needs stream lengths differ: legacy %d, k1 %d", len(logL), len(logK))
	}
	for i := range logL {
		if logL[i].Tick != logK[i].Tick || !bytes.Equal(logL[i].Payload, logK[i].Payload) {
			t.Fatalf("needs row %d diverges: legacy %d %s vs k1 %d %s",
				i, logL[i].Tick, logL[i].Payload, logK[i].Tick, logK[i].Payload)
		}
	}
}

// TestNeedsK10DerivedMatchesK1Fold: on a K=10 world the DERIVED per-minute
// values equal the K=1 world's event-folded values at every minute (decay
// exactness), while emitting a fraction of the rows.
func TestNeedsK10DerivedMatchesK1Fold(t *testing.T) {
	m, _ := needsIslands()
	k1 := NewState(7, m)
	if err := k1.Apply(tuningEventK(0, 1)); err != nil {
		t.Fatal(err)
	}
	k10 := NewState(7, m)
	if err := k10.Apply(tuningEventK(0, 10)); err != nil {
		t.Fatal(err)
	}
	var rows1, rows10 int
	for minute := int64(1); minute <= 120; minute++ {
		rows1 += len(needsEvents(driveTicks(t, k1, m, minute*60, nil)))
		rows10 += len(needsEvents(driveTicks(t, k10, m, minute*60, nil)))
		// Compare at the tick boundary (research.md §1): items scheduled AT
		// the minute tick execute strictly after it — the same AdvanceTo the
		// next runTick would perform, idempotent when it re-runs there.
		k1.AdvanceTo(minute*60 + 1)
		k10.AdvanceTo(minute*60 + 1)
		for i := range k1.Agents {
			if k1.Agents[i].Needs != k10.Agents[i].Needs {
				t.Fatalf("minute %d agent %d: k1 %+v vs k10 %+v", minute, i, k1.Agents[i].Needs, k10.Agents[i].Needs)
			}
		}
	}
	if rows10 >= rows1/5 {
		t.Fatalf("K=10 should thin the needs stream several-fold: k1 %d rows, k10 %d rows", rows1, rows10)
	}
}

// TestNeedsDoubleDecayGuard: a per-minute event stream (the old-log shape)
// folded on a REGIME world contributes no derived decay on top — the folded
// values equal a legacy fold of the same rows, minute by minute.
func TestNeedsDoubleDecayGuard(t *testing.T) {
	m, pts := needsIslands()
	build := func(k int64) *State {
		s := NewState(7, m)
		if err := s.Apply(tuningEventK(0, k)); err != nil {
			t.Fatal(err)
		}
		return s
	}
	legacy, regime := build(0), build(1)
	// Hand-author an old-log-style stream: absolutes every minute for agent
	// 0, decayed independently of either state's engine.
	n := legacy.Agents[0].Needs
	for minute := int64(1); minute <= 30; minute++ {
		n = decayNeeds(n, false, false, false, false, false)
		ev := store.Event{Tick: minute * 60, Type: "agent.needs_changed", Payload: mustPayload(NeedsPayload{
			Agent: Ref(0), Health: n.Health, Food: n.Food, Rest: n.Rest, Warmth: n.Warmth, Morale: n.Morale,
		})}
		if err := legacy.Apply(ev); err != nil {
			t.Fatal(err)
		}
		if err := regime.Apply(ev); err != nil {
			t.Fatal(err)
		}
		regime.AdvanceTo(minute*60 + 1) // give derived decay every chance to double-fold
		if legacy.Agents[0].Needs != regime.Agents[0].Needs {
			t.Fatalf("minute %d: double decay — legacy %+v vs regime %+v", minute, legacy.Agents[0].Needs, regime.Agents[0].Needs)
		}
		if regime.Agents[0].NeedsSyncTick != minute*60 {
			t.Fatalf("watermark not stamped by the arm: %d", regime.Agents[0].NeedsSyncTick)
		}
	}
	_ = pts
}

// TestNeedsCrossingEmitsAtTheMinute (ruling 3): on a K=10 world a danger-band
// crossing lands its agent.needs_changed at the exact minute the boundary is
// crossed — never deferred to the checkpoint — and the checkpoint grid still
// fires. Guardian survival watches and standing orders match this event type
// in their live absorb path, so crossing-at-the-minute IS the latency proof
// at the emission layer (the matcher itself is exercised in internal/
// guardian's order tests).
func TestNeedsCrossingEmitsAtTheMinute(t *testing.T) {
	m, pts := needsIslands()
	s := NewState(7, m)
	if err := s.Apply(tuningEventK(0, 10)); err != nil {
		t.Fatal(err)
	}
	isolateAgents(s)
	a := &s.Agents[0]
	a.Dead = false
	a.X, a.Y = pts[0].X, pts[0].Y
	a.Needs = Needs{Health: 1000, Food: dangerFoodBelow + 3, Rest: 900, Warmth: 900, Morale: 900}
	emitted := a.Needs
	a.NeedsEmitted = &emitted // regime baseline: pretend the current values were just emitted
	// Independent ladder: the first minute the food need sits strictly below
	// its danger band.
	expect := int64(0)
	probe := a.Needs
	for minute := int64(1); expect == 0; minute++ {
		probe = decayNeeds(probe, false, false, false, false, false)
		if probe.Food < dangerFoodBelow {
			expect = minute * 60
		}
	}
	log := needsEvents(driveTicks(t, s, m, expect+60, nil))
	var at []int64
	for _, e := range log {
		var p NeedsPayload
		if json.Unmarshal(e.Payload, &p) == nil && p.Agent.ID == 0 {
			at = append(at, e.Tick)
		}
	}
	// The crossing minute must appear; minutes before it that are neither
	// checkpoints nor crossings must not.
	found := false
	for _, tick := range at {
		if tick == expect {
			found = true
		}
		if tick < expect && (tick/60)%10 != 0 {
			t.Fatalf("unexpected off-checkpoint emission at %d before the crossing %d", tick, expect)
		}
	}
	if !found {
		t.Fatalf("crossing at %d not emitted at its minute; emissions: %v", expect, at)
	}
}

// --- T010: gru derivation ---------------------------------------------------

// TestGruDerivedMatchesLegacyDecision: with identical geometry, the derived
// gru walks EXACTLY the tiles the legacy emitter would have recorded — the
// shared gruMoveDecision, beat for beat (agents stationary, so the
// through-t/through-t-1 read difference is inert).
func TestGruDerivedMatchesLegacyDecision(t *testing.T) {
	m := flatMap(32, 32)
	build := func(k int64) *State {
		s := NewState(11, m)
		if err := s.Apply(tuningEventK(0, k)); err != nil {
			t.Fatal(err)
		}
		isolateAgents(s)
		a := &s.Agents[0]
		a.Dead = false
		a.X, a.Y = 5, 5
		a.Needs = Needs{Health: 1000, Food: 900, Rest: 900, Warmth: 900, Morale: 900}
		a.Asleep = true // no reflex wandering; sightings/attacks still possible but deterministic in both
		s.Night = true
		s.Gru = &Gru{X: 25, Y: 25}
		return s
	}
	legacy, derived := build(0), build(1)
	if derived.Gru.Done != 0 {
		t.Fatalf("gru Done should be 0 before any derived beat, got %d", derived.Gru.Done)
	}
	for tick := int64(1); tick <= 400; tick++ {
		// Legacy: fold the emitted gru.moved exactly as the loop would.
		for _, e := range stepEvents(legacy, m, tick) {
			if e.Type == "gru.moved" || e.Type == "gru.attacked" || e.Type == "agent.died" {
				if err := legacy.Apply(e); err != nil {
					t.Fatal(err)
				}
			}
		}
		legacy.Tick = tick
		derived.Tick = tick
		derived.AdvanceTo(tick + 1)
		if legacy.Gru == nil || derived.Gru == nil {
			t.Fatalf("gru vanished at %d", tick)
		}
		if legacy.Gru.X != derived.Gru.X || legacy.Gru.Y != derived.Gru.Y {
			t.Fatalf("tick %d: legacy gru (%d,%d) vs derived (%d,%d)",
				tick, legacy.Gru.X, legacy.Gru.Y, derived.Gru.X, derived.Gru.Y)
		}
	}
}

// TestGruBeatAfterAgentSteps pins the within-tick order (research.md §2 /
// plan D4): at a tick where an agent step and a gru beat coincide, the
// derived gru stalks the walker's POST-step position.
func TestGruBeatAfterAgentSteps(t *testing.T) {
	m := flatMap(32, 32)
	s := coalescedState(t, 42, m, 10)
	isolateAgents(s)
	a := reviveAt(s, 10, 5)
	a.Intent = &Intent{Goal: "wander", TargetX: 10, TargetY: 6}
	s.Night = true
	s.Gru = &Gru{X: 12, Y: 5, Done: 16}
	if err := s.Apply(store.Event{Tick: 16, Type: "agent.path_started", Payload: mustPayload(PathStartedPayload{
		Agent: Ref(0), Path: []Point{{X: 10, Y: 6}}, MoveEvery: moveEveryTicks, Phase: 0,
	})}); err != nil {
		t.Fatal(err)
	}
	// Tick 20: agent-0 beat (phase 0) AND gru beat (20%4==0). Stalking the
	// PRE-step target (10,5) picks W=(11,5); the POST-step target (10,6)
	// picks S=(12,6) — the order is observable in the chosen tile.
	s.AdvanceTo(21)
	if a.X != 10 || a.Y != 6 {
		t.Fatalf("walker step at 20 missing: at (%d,%d)", a.X, a.Y)
	}
	if s.Gru.X != 12 || s.Gru.Y != 6 {
		t.Fatalf("gru should stalk the post-step position to (12,6), went to (%d,%d)", s.Gru.X, s.Gru.Y)
	}
}

// TestGruNoEmissionUnderRegime: a coalesced night emits no gru.moved rows at
// all while the gru still prowls (derived), and the emergence/withdrawal/
// sighting/attack vocabulary is untouched.
func TestGruNoEmissionUnderRegime(t *testing.T) {
	m := flatMap(32, 32)
	s := coalescedState(t, 11, m, 10)
	isolateAgents(s)
	s.Night = true
	s.Gru = &Gru{X: 25, Y: 25, Done: 0}
	start := *s.Gru
	log := driveTicks(t, s, m, 600, nil)
	for _, e := range log {
		if e.Type == "gru.moved" {
			t.Fatalf("gru.moved emitted under the coalescing regime at %d", e.Tick)
		}
	}
	if s.Gru != nil && s.Gru.X == start.X && s.Gru.Y == start.Y && s.Gru.Done == 0 {
		t.Fatal("derived gru never moved")
	}
}

// --- T005: executor emission rewire ----------------------------------------

// TestCoalescedWalkEmission: one intent walk = one agent.path_started, zero
// agent.moved, exactly one spec-097 arrival observation at the arrival
// step's tick, and the walker lands on its target.
func TestCoalescedWalkEmission(t *testing.T) {
	m := flatMap(32, 32)
	s := coalescedState(t, 42, m, 10)
	isolateAgents(s)
	a := reviveAt(s, 5, 5)
	a.Intent = &Intent{Goal: "wander", TargetX: 13, TargetY: 5}
	log := driveTicks(t, s, m, 120, nil)
	started, moved, observed := 0, 0, 0
	var arrivalObsTick int64
	for _, e := range log {
		switch e.Type {
		case "agent.path_started":
			started++
		case "agent.moved":
			moved++
		case "agent.place_observed":
			observed++
			arrivalObsTick = e.Tick
		}
	}
	if started != 1 || moved != 0 {
		t.Fatalf("want 1 path_started / 0 agent.moved, got %d / %d", started, moved)
	}
	if observed != 1 {
		t.Fatalf("want exactly one arrival observation, got %d", observed)
	}
	if a.X != 13 || a.Y != 5 || a.Path != nil {
		t.Fatalf("walker did not arrive cleanly: (%d,%d) seg=%v", a.X, a.Y, a.Path)
	}
	// The observation rode the arrival step's own tick: the walker's LastObs
	// was stamped at that tick by the reducer arm.
	if a.LastObs == nil || a.LastObs.Tick != arrivalObsTick {
		t.Fatalf("arrival observation mark mismatch: %+v vs tick %d", a.LastObs, arrivalObsTick)
	}
}

// TestBlockedPathTruncatesAndReplans: a wall built across the declared path
// truncates the walk (the one executor-emitted deviation) and the next tick
// re-plans around it; the walker still arrives.
func TestBlockedPathTruncatesAndReplans(t *testing.T) {
	m := flatMap(32, 32)
	s := coalescedState(t, 42, m, 10)
	isolateAgents(s)
	a := reviveAt(s, 5, 5)
	a.Intent = &Intent{Goal: "wander", TargetX: 15, TargetY: 5}
	log := driveTicks(t, s, m, 20, nil)
	// Drop a wall a couple of tiles ahead of the cursor — NOT on the step
	// already pending at this very tick (that same-tick race is today's
	// behavior too and is documented in research.md §4); the executor's
	// blockage check meets it before any step does.
	if a.Path == nil || a.Path.Next+2 >= len(a.Path.Path) {
		t.Fatalf("no in-flight segment to block: %+v", a.Path)
	}
	block := a.Path.Path[a.Path.Next+2]
	s.Structures = append(s.Structures, Structure{Kind: "wall_stone", X: block.X, Y: block.Y, HP: wallMaxHP("wall_stone")})
	for !(a.X == 15 && a.Y == 5) && s.Tick < 600 {
		log = append(log, driveTicks(t, s, m, s.Tick+10, nil)...)
	}
	truncated, started := 0, 0
	for _, e := range log {
		switch e.Type {
		case "agent.path_truncated":
			truncated++
		case "agent.path_started":
			started++
		}
	}
	if truncated != 1 || started < 2 {
		t.Fatalf("want 1 truncation + a replanned path_started, got %d / %d", truncated, started)
	}
	if a.X != 15 || a.Y != 5 {
		t.Fatalf("walker never detoured to the target: (%d,%d)", a.X, a.Y)
	}
}

// TestPauseTruncatesAllWalks: clock.paused clears every in-flight segment
// (the pause arm's truncate-all) and the walk re-plans after resume.
func TestPauseTruncatesAllWalks(t *testing.T) {
	m := flatMap(32, 32)
	s := coalescedState(t, 42, m, 10)
	isolateAgents(s)
	a := reviveAt(s, 5, 5)
	a.Intent = &Intent{Goal: "wander", TargetX: 15, TargetY: 5}
	driveTicks(t, s, m, 20, nil)
	if a.Path == nil {
		t.Fatal("no in-flight segment before the pause")
	}
	if err := s.Apply(store.Event{Tick: s.Tick, Type: "clock.paused", Payload: mustPayload(struct{}{})}); err != nil {
		t.Fatal(err)
	}
	if a.Path != nil {
		t.Fatal("pause did not truncate the in-flight segment")
	}
	if err := s.Apply(store.Event{Tick: s.Tick, Type: "clock.resumed", Payload: mustPayload(struct{}{})}); err != nil {
		t.Fatal(err)
	}
	log := driveTicks(t, s, m, s.Tick+120, nil)
	replanned := false
	for _, e := range log {
		if e.Type == "agent.path_started" {
			replanned = true
		}
	}
	if !replanned || a.X != 15 {
		t.Fatalf("no replan after resume (replanned=%v, at (%d,%d))", replanned, a.X, a.Y)
	}
}

// TestSegmentSnapshotRoundTrip is the mid-walk kill-9 shape at the sim
// layer: a state snapshotted mid-segment and mid-needs-window, restored from
// bytes (plus SetMap, as recovery does), continues to byte-identical states.
func TestSegmentSnapshotRoundTrip(t *testing.T) {
	m := flatMap(32, 32)
	s := coalescedState(t, 42, m, 10)
	isolateAgents(s)
	a := reviveAt(s, 5, 5)
	a.Intent = &Intent{Goal: "wander", TargetX: 15, TargetY: 5}
	driveTicks(t, s, m, 23, nil) // mid-walk, mid-minute
	if a.Path == nil {
		t.Fatal("expected an in-flight segment at the snapshot point")
	}
	restored := &State{}
	if err := json.Unmarshal(s.Marshal(), restored); err != nil {
		t.Fatal(err)
	}
	restored.SetMap(m)
	driveTicks(t, s, m, 400, nil)
	driveTicks(t, restored, m, 400, nil)
	if !bytes.Equal(s.Marshal(), restored.Marshal()) {
		t.Fatalf("mid-segment snapshot round-trip diverged:\nlive:     %s\nrestored: %s", s.Marshal(), restored.Marshal())
	}
}

// TestCoalescedDeterminismSameSeedSameTimeline is the spec-104 sibling of
// TestDeterminismSameSeedSameTimeline: a full coalesced world (regime pinned
// at genesis) is self-deterministic — same seed, same command timeline, same
// event bytes, same state hash, same per-agent canonical mental-map bytes —
// across the day/night boundary, walks, derived needs decay, and derived gru
// motion.
func TestCoalescedDeterminismSameSeedSameTimeline(t *testing.T) {
	const seed, ticks = 7, 30_000
	m := testMap(seed)
	build := func() *State {
		return coalescedState(t, seed, m, 10)
	}
	a, b := build(), build()
	logA := driveTicks(t, a, m, ticks, commandTimeline())
	logB := driveTicks(t, b, m, ticks, commandTimeline())
	if len(logA) == 0 {
		t.Fatal("30k coalesced ticks should produce events")
	}
	if !bytes.Equal(canonicalLog(t, logA), canonicalLog(t, logB)) {
		t.Fatal("same seed + same commands produced different coalesced event sequences")
	}
	if a.Hash() != b.Hash() {
		t.Fatalf("coalesced state hashes diverged: %s vs %s", a.Hash(), b.Hash())
	}
	for i := range a.Agents {
		am, bm := mustPayload(a.Agents[i].Map), mustPayload(b.Agents[i].Map)
		if !bytes.Equal(am, bm) {
			t.Errorf("agent %d mental map diverged across same-seed coalesced runs:\n%s\n%s", i, am, bm)
		}
	}
	moved, started := 0, 0
	for _, e := range logA {
		switch e.Type {
		case "agent.moved":
			moved++
		case "agent.path_started":
			started++
		}
	}
	if moved != 0 || started == 0 {
		t.Fatalf("coalesced world emitted %d agent.moved / %d path_started", moved, started)
	}
}

// TestCoalescedReplayRebuildsState is TestReplayRebuildsState under the
// regime: folding the coalesced log over genesis (Apply's AdvanceTo hook
// interleaving derived steps exactly), then re-living the quiet tail, lands
// the exact live state — the replay-to-cutoff / recovery contract for
// derived progress (FR-002's new-world half).
func TestCoalescedReplayRebuildsState(t *testing.T) {
	const seed, ticks = 99, 40_000
	m := testMap(seed)
	live := coalescedState(t, seed, m, 10)
	log := driveTicks(t, live, m, ticks, commandTimeline())

	replayed := coalescedState(t, seed, m, 10)
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replayed.Tick = e.Tick
	}
	driveTicks(t, replayed, m, ticks, nil)
	if live.Hash() != replayed.Hash() {
		t.Fatalf("coalesced replay diverged from live:\nlive:     %s\nreplayed: %s",
			string(live.Marshal()), string(replayed.Marshal()))
	}
}

// TestLegacyStateCarriesNoNewFields: a legacy world's canonical bytes never
// mention the spec-104 fields (pre-change snapshot byte-identity, the
// omitempty contract — the recorded pre-086 fixture hash is the cross-build
// anchor).
func TestLegacyStateCarriesNoNewFields(t *testing.T) {
	m := testMap(42)
	s := NewState(42, m)
	driveTicks(t, s, m, 400, nil)
	b := s.Marshal()
	for _, key := range []string{"needs_sync_tick", "needs_emitted", `"path"`, "needs_checkpoint_minutes", `"done"`} {
		if bytes.Contains(b, []byte(key)) {
			t.Fatalf("legacy state bytes carry %q", key)
		}
	}
}
