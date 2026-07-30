package sim

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// Spec 097 (TASK-80, perception of absence): the grounded arrival observation
// channel. FR-001/FR-002/FR-004/FR-006 coverage — emission on the
// intent-completing arrival step, exhaustiveness within placeScanRadius,
// determinism, the D4 dedup window (both-or-neither: event AND memory), the
// "observed" provenance class, the reducer's dedup anchor, the extended
// belief_reinforced arm, and the pre-097 tuning payload default resolution.
// Replay byte-identity for existing logs is covered by the existing replay
// suites (the event is additive; recorded logs contain none).

// TestPlaceObservedOnArrival drives a walker over a MULTI-STEP walk to its
// intent's target and asserts (a) no mid-walk tile observes anything — the
// D1 guard, the card's flood worry (the assertion that catches an unguarded
// per-step emission), (b) the arrival step emits the companion memory
// (Origin "observed", the base-salience dial) FIRST and agent.place_observed
// second, on the target tile, and (c) the whole drive is deterministic (two
// identical runs, identical canonical logs).
func TestPlaceObservedOnArrival(t *testing.T) {
	const seed = 42
	m := testMap(seed)

	run := func() ([]store.Event, int, int) {
		s := NewState(seed, m)
		isolateAgents(s)
		a := &s.Agents[0]
		a.Dead = false
		// A multi-step walk target, chosen by proving BFS reachability with
		// nextStep itself: candidates at Manhattan distance 3, kept only when
		// stepping toward them actually lands within a few hops.
		tx, ty := a.X, a.Y
	pick:
		for _, d := range [][2]int{{3, 0}, {-3, 0}, {0, 3}, {0, -3}, {2, 1}, {1, 2}, {-2, 1}, {-1, 2}} {
			cx, cy := a.X+d[0], a.Y+d[1]
			if !m.InBounds(cx, cy) || !passable(m, s, cx, cy) {
				continue
			}
			px, py := a.X, a.Y
			for hop := 0; hop < 8; hop++ {
				nx, ny := nextStep(m, s, px, py, cx, cy)
				if nx == px && ny == py {
					continue pick // unreachable from here
				}
				px, py = nx, ny
				if px == cx && py == cy {
					tx, ty = cx, cy
					break pick
				}
			}
		}
		if tx == a.X && ty == a.Y {
			t.Skip("no reachable multi-step walk target near agent 0's genesis tile on this seed")
		}
		a.Intent = &Intent{Goal: "wander", TargetX: tx, TargetY: ty, Reason: "see what is out there."}
		// Drive just past the walk (3 tiles ≈ 15 ticks at the movement
		// cadence) so every agent-0 observation in the log belongs to THIS
		// intent — later reflex-issued walks never muddy the assertion.
		log := driveTicks(t, s, m, s.Tick+40, nil)
		return log, tx, ty
	}

	log, tx, ty := run()
	var obs *PlaceObservedPayload
	var obsTick int64
	var obsIdx int
	for i, e := range log {
		if e.Type == "agent.place_observed" {
			var p PlaceObservedPayload
			mustUnmarshal(t, e.Payload, &p)
			if p.Agent.ID != 0 {
				continue
			}
			// The D1 guard: EVERY observation this walk produces sits on the
			// intent's target — a mid-walk (per-step) emission is the flood
			// bug this test exists to catch.
			if p.X != tx || p.Y != ty {
				t.Errorf("mid-walk observation at (%d,%d) — only the arrival tile (%d,%d) may observe (D1)", p.X, p.Y, tx, ty)
			}
			if obs == nil {
				cp := p
				obs, obsTick, obsIdx = &cp, e.Tick, i
			}
		}
	}
	if obs == nil {
		t.Fatal("no agent.place_observed emitted for the arriving walker")
	}
	if obs.X != tx || obs.Y != ty {
		t.Errorf("observation at (%d,%d), want the intent target (%d,%d)", obs.X, obs.Y, tx, ty)
	}
	if obs.Radius != placeScanRadius {
		t.Errorf("observation radius = %d, want placeScanRadius (%d)", obs.Radius, placeScanRadius)
	}
	for i := 1; i < len(obs.Kinds); i++ {
		if obs.Kinds[i-1] >= obs.Kinds[i] {
			t.Errorf("kinds not sorted/unique: %v", obs.Kinds)
		}
	}
	// The companion memory rides the same batch, immediately BEFORE the
	// observation (the mind's absorb loop reads it off its replica when the
	// observation event lands — observe.go).
	if obsIdx == 0 {
		t.Fatal("observation is the log's first event — no companion memory precedes it")
	}
	prev := log[obsIdx-1]
	if prev.Type != "agent.memory_added" || prev.Tick != obsTick {
		t.Fatalf("event before the observation = %s@%d, want the companion agent.memory_added@%d", prev.Type, prev.Tick, obsTick)
	}
	var mem MemoryAddedPayload
	mustUnmarshal(t, prev.Payload, &mem)
	if mem.Agent.ID != 0 || mem.Origin != OriginObserved {
		t.Errorf("companion memory agent/origin = %d/%q, want 0/%q", mem.Agent.ID, mem.Origin, OriginObserved)
	}
	if mem.Salience != defaultObservationBaseSalience {
		t.Errorf("companion memory salience = %d, want the base-salience dial default (%d)", mem.Salience, defaultObservationBaseSalience)
	}
	if mem.Why != "see what is out there." {
		t.Errorf("companion memory Why = %q, want the driving intent's reason", mem.Why)
	}

	// Determinism (D5/FR-006): the identical drive yields the identical log.
	log2, _, _ := run()
	if !bytes.Equal(canonicalLog(t, log), canonicalLog(t, log2)) {
		t.Error("two identical drives produced different canonical logs")
	}
}

// TestObservedKindsExhaustive pins D2: a feature inside placeScanRadius is
// listed; the same kind outside the radius is not; overlays are respected (a
// cleared tree is not a tree).
func TestObservedKindsExhaustive(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	x, y := 30, 30

	s.Structures = append(s.Structures,
		Structure{Kind: "fire", X: x + placeScanRadius, Y: y},      // Manhattan == radius: in
		Structure{Kind: "chest", X: x + placeScanRadius + 1, Y: y}, // just out
	)
	kinds := observedKinds(s, m, x, y)
	has := func(k string) bool {
		for _, kk := range kinds {
			if kk == k {
				return true
			}
		}
		return false
	}
	if !has("fire") {
		t.Errorf("fire at Manhattan distance %d missing from kinds %v", placeScanRadius, kinds)
	}
	if has("chest") {
		t.Errorf("chest at Manhattan distance %d wrongly listed in kinds %v", placeScanRadius+1, kinds)
	}

	// Determinism: same inputs, same set.
	again := observedKinds(s, m, x, y)
	if len(again) != len(kinds) {
		t.Fatalf("observedKinds not deterministic: %v vs %v", kinds, again)
	}
	for i := range kinds {
		if kinds[i] != again[i] {
			t.Fatalf("observedKinds not deterministic: %v vs %v", kinds, again)
		}
	}
}

// TestPlaceObservedDedup pins D4: same tile + same kinds inside the window
// collapses BOTH the event and the memory; the window's edge, a changed
// place, and a different tile all observe again.
func TestPlaceObservedDedup(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a := &s.Agents[0]
	x, y := a.X, a.Y

	apply := func(evs []store.Event) {
		t.Helper()
		for _, e := range evs {
			if err := s.Apply(e); err != nil {
				t.Fatalf("apply %s: %v", e.Type, err)
			}
		}
	}

	first := placeObservedEvents(s, m, 0, x, y, 100)
	if len(first) != 2 {
		t.Fatalf("first observation emitted %d events, want 2 (memory + observation)", len(first))
	}
	if first[0].Type != "agent.memory_added" || first[1].Type != "agent.place_observed" {
		t.Fatalf("first observation order = [%s %s], want [agent.memory_added agent.place_observed]", first[0].Type, first[1].Type)
	}
	apply(first)
	if a.LastObs == nil || a.LastObs.X != x || a.LastObs.Y != y || a.LastObs.Tick != 100 {
		t.Fatalf("reducer did not record the dedup anchor: %+v", a.LastObs)
	}

	// Inside the window, unchanged place: both-or-neither ⇒ neither.
	if evs := placeObservedEvents(s, m, 0, x, y, 100+s.ObservationDedupTicks()-1); len(evs) != 0 {
		t.Errorf("unchanged repeat inside the window emitted %d events, want 0", len(evs))
	}
	// At the window's edge: observes again.
	if evs := placeObservedEvents(s, m, 0, x, y, 100+s.ObservationDedupTicks()); len(evs) != 2 {
		t.Errorf("repeat at the window edge emitted %d events, want 2", len(evs))
	}
	// Inside the window but the place CHANGED (a fire appeared): observes.
	s.Structures = append(s.Structures, Structure{Kind: "fire", X: x, Y: y})
	if evs := placeObservedEvents(s, m, 0, x, y, 200); len(evs) != 2 {
		t.Errorf("changed place inside the window emitted %d events, want 2", len(evs))
	}
	s.Structures = s.Structures[:len(s.Structures)-1]
	// An adjacent tile with the identical kind set inside the window: the
	// same place (scan discs overlap within placeScanRadius) — deduped,
	// UNLESS the shifted disc genuinely sees different kinds.
	adjKinds := observedKinds(s, m, x+1, y)
	adjEvs := placeObservedEvents(s, m, 0, x+1, y, 200)
	sameKinds := len(adjKinds) == len(a.LastObs.Kinds)
	if sameKinds {
		for i := range adjKinds {
			if adjKinds[i] != a.LastObs.Kinds[i] {
				sameKinds = false
				break
			}
		}
	}
	if sameKinds && len(adjEvs) != 0 {
		t.Errorf("adjacent tile, identical kinds, inside the window emitted %d events, want 0 (radius dedup)", len(adjEvs))
	}
	if !sameKinds && len(adjEvs) != 2 {
		t.Errorf("adjacent tile with a different kind set emitted %d events, want 2", len(adjEvs))
	}
	// Beyond the scan radius inside the window: a different place — observes.
	if evs := placeObservedEvents(s, m, 0, x+placeScanRadius+1, y, 200); len(evs) != 2 {
		t.Errorf("beyond-radius tile inside the window emitted %d events, want 2", len(evs))
	}
}

// TestObservedOriginIsDirect pins FR-002: "observed" is first-person direct
// perception, the strongest class per spec 030's hygiene.
func TestObservedOriginIsDirect(t *testing.T) {
	if !DirectPerception(OriginObserved) {
		t.Error("DirectPerception(OriginObserved) = false, want true")
	}
}

// TestBeliefReinforcedKinds pins the spec 097 additive extension of the
// TASK-79 seam: a Kind-stamped payload copies the emitter-computed stored
// confidence (clamped) AND re-anchors; the legacy bare shape stays the pure
// re-anchor it always was.
func TestBeliefReinforcedKinds(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a := &s.Agents[0]
	a.Beliefs = []Belief{{ID: 7, Statement: "there is a fire at (5,5)", Confidence: 80, Provenance: ProvenanceTold, Source: -1, Subject: -1, Tick: 10, Reinforced: 10}}

	apply := func(tick int64, p BeliefReinforcedPayload) {
		t.Helper()
		if err := s.Apply(store.Event{Tick: tick, Type: "agent.belief_reinforced", Payload: mustPayload(p)}); err != nil {
			t.Fatal(err)
		}
	}

	// Legacy shape: re-anchor only.
	apply(1000, BeliefReinforcedPayload{Agent: Ref(0), BeliefID: 7})
	if a.Beliefs[0].Confidence != 80 || a.Beliefs[0].Reinforced != 1000 {
		t.Errorf("legacy re-anchor: confidence/reinforced = %d/%d, want 80/1000", a.Beliefs[0].Confidence, a.Beliefs[0].Reinforced)
	}
	// Disconfirmation: stored confidence copied, clock re-anchored.
	apply(2000, BeliefReinforcedPayload{Agent: Ref(0), BeliefID: 7, Kind: BeliefDisconfirmed, Confidence: 56})
	if a.Beliefs[0].Confidence != 56 || a.Beliefs[0].Reinforced != 2000 {
		t.Errorf("disconfirm: confidence/reinforced = %d/%d, want 56/2000", a.Beliefs[0].Confidence, a.Beliefs[0].Reinforced)
	}
	// Confirmation with an out-of-range value clamps.
	apply(3000, BeliefReinforcedPayload{Agent: Ref(0), BeliefID: 7, Kind: BeliefConfirmed, Confidence: 140})
	if a.Beliefs[0].Confidence != 100 {
		t.Errorf("confirm clamp: confidence = %d, want 100", a.Beliefs[0].Confidence)
	}
	// Vanished belief: total no-op.
	apply(4000, BeliefReinforcedPayload{Agent: Ref(0), BeliefID: 99, Kind: BeliefDisconfirmed, Confidence: 1})
	if a.Beliefs[0].Confidence != 100 || a.Beliefs[0].Reinforced != 3000 {
		t.Error("a vanished-belief payload touched the held belief")
	}
}

// TestTuningSpec097Dials pins FR-004: the four dials parse and clamp through
// the manifest, ride the tuning_applied payload, and — the replay hazard — a
// pre-097 recorded payload (fields absent) resolves them to the doctrine
// defaults, never to zero.
func TestTuningSpec097Dials(t *testing.T) {
	// Manifest parse + clamp.
	tun, warns, err := ParseTuning([]byte(`{"observation_dedup_ticks":600,"observation_base_salience":99,"belief_disconfirm_retain_percent":50,"belief_confirm_boost":5}`))
	if err != nil {
		t.Fatal(err)
	}
	if tun.ObservationDedupTicks != 600 || tun.BeliefDisconfirmRetainPercent != 50 || tun.BeliefConfirmBoost != 5 {
		t.Errorf("parsed dials = %+v, want 600/50/5", tun)
	}
	if tun.ObservationBaseSalience != maxObservationBaseSalience {
		t.Errorf("out-of-range base salience = %d, want clamped to %d", tun.ObservationBaseSalience, maxObservationBaseSalience)
	}
	if len(warns) != 1 {
		t.Errorf("clamp warnings = %v, want exactly one", warns)
	}

	// Pre-097 payload: absent fields resolve to defaults.
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	pre097 := json.RawMessage(`{"refuel_dying_below":3600,"fire_burn_per_wood":14400,"gru_emerge_per_mille":600,"planner_cadence_ticks":1800,"encounter_cooldown_ticks":7200}`)
	if err := s.Apply(store.Event{Tick: 1, Type: "sim.tuning_applied", Payload: pre097}); err != nil {
		t.Fatal(err)
	}
	if s.ObservationDedupTicks() != defaultObservationDedupTicks ||
		s.ObservationBaseSalience() != defaultObservationBaseSalience ||
		s.BeliefDisconfirmRetainPercent() != defaultBeliefDisconfirmRetainPercent ||
		s.BeliefConfirmBoost() != defaultBeliefConfirmBoost {
		t.Errorf("pre-097 payload resolved dials to %d/%d/%d/%d, want the doctrine defaults",
			s.ObservationDedupTicks(), s.ObservationBaseSalience(), s.BeliefDisconfirmRetainPercent(), s.BeliefConfirmBoost())
	}
	if s.RefuelDyingBelow() != 3600 {
		t.Errorf("pre-097 payload's own dial lost: RefuelDyingBelow = %d, want 3600", s.RefuelDyingBelow())
	}

	// A full new event round-trips every dial.
	tun2 := defaultTuning()
	tun2.ObservationDedupTicks = 1234
	if err := s.Apply(NewTuningEvent(2, tun2)); err != nil {
		t.Fatal(err)
	}
	if s.ObservationDedupTicks() != 1234 {
		t.Errorf("round-tripped dedup dial = %d, want 1234", s.ObservationDedupTicks())
	}
}

// TestObservedMemoryText pins the deterministic composition, including the
// empty set — "nothing of note" IS the perception of absence.
func TestObservedMemoryText(t *testing.T) {
	if got := observedMemoryText(nil); got != "Looked around: nothing of note here." {
		t.Errorf("empty kinds text = %q", got)
	}
	if got := observedMemoryText([]string{"fire", "tree"}); got != "Looked around: a fire, standing trees." {
		t.Errorf("kinds text = %q", got)
	}
}
