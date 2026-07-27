package sim

// Spec 085 faith suites (T004/T006/T007/T009): the faith.changed fold table
// and nil-safe accessor, pre-085 snapshot byte-identity, the faith accounting
// sweep (per-source emission, batch-order determinism, the cannot-move gate,
// the excluded sources), from-genesis replay proofs, the regen band × posture
// matrix, and the genesis-band schedule byte-identity pin.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

func faithEvent(tick int64, delta int, reason, sourceID string) store.Event {
	return store.Event{Tick: tick, Type: "faith.changed",
		Payload: mustPayload(FaithChangedPayload{Delta: delta, Reason: reason, SourceID: sourceID})}
}

// TestFaithNilAccessorGenesis (FR-001): a fresh world holds no faith state at
// all — nil field, genesis score through the accessor, and no "faith" key in
// the canonical bytes (omitempty; the pre-085 byte-identity half is
// TestPre085FaithSnapshotCompat below).
func TestFaithNilAccessorGenesis(t *testing.T) {
	s := NewState(42, testMap(42))
	if s.Faith != nil {
		t.Fatal("genesis state materialized Faith")
	}
	if got := s.FaithScore(); got != FaithGenesis {
		t.Fatalf("FaithScore() = %d, want genesis %d", got, FaithGenesis)
	}
}

// TestFaithChangedFoldTable (FR-002): the reducer arm validates the reason
// domain and the delta's sign against the doctrine table (magnitudes are
// dial-ready and fold as recorded), folds with clamping at both ends, and
// materializes Faith on the first fold.
func TestFaithChangedFoldTable(t *testing.T) {
	cases := []struct {
		name    string
		pre     *FaithState // nil = genesis
		delta   int
		reason  string
		wantErr string // "" = accepted
		want    int    // folded score when accepted
	}{
		{"first fold materializes from genesis", nil, FaithDeltaDirectiveFulfilled, FaithReasonDirectiveFulfilled, "", 58},
		{"negative fold", &FaithState{Score: 30}, FaithDeltaVillagerDied, FaithReasonVillagerDied, "", 24},
		{"clamp at 100", &FaithState{Score: 95}, FaithDeltaProphecyFulfilled, FaithReasonProphecyFulfilled, "", 100},
		{"clamp at 0", &FaithState{Score: 3}, FaithDeltaVillagerDied, FaithReasonVillagerDied, "", 0},
		{"dial-ready magnitude folds as recorded", &FaithState{Score: 50}, 5, FaithReasonDirectiveFulfilled, "", 55},
		{"unknown reason refused", &FaithState{Score: 50}, 8, "tutoring", "unknown reason", 0},
		{"zero delta refused", &FaithState{Score: 50}, 0, FaithReasonDirectiveFulfilled, "zero delta", 0},
		{"sign mismatch refused", &FaithState{Score: 50}, -8, FaithReasonDirectiveFulfilled, "contradicts", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := NewState(42, testMap(42))
			s.Faith = c.pre
			err := s.Apply(faithEvent(100, c.delta, c.reason, "src"))
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("err = %v, want %q", err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("fold refused: %v", err)
			}
			if s.Faith == nil || s.Faith.Score != c.want {
				t.Fatalf("Faith = %+v, want score %d", s.Faith, c.want)
			}
		})
	}
}

// TestPre085FaithSnapshotCompat: a snapshot with neither spec-085 field
// unmarshals to nil and re-marshals byte-identically — omitempty, no format
// bump (the spec-084 compat shape).
func TestPre085FaithSnapshotCompat(t *testing.T) {
	s := NewState(42, testMap(42))
	before := s.Marshal()
	for _, key := range []string{"\"faith\"", "\"prophecies\""} {
		if strings.Contains(string(before), key) {
			t.Fatalf("empty spec-085 state leaks %s into canonical bytes (omitempty broken)", key)
		}
	}
	restored := NewState(42, s.m)
	if err := json.Unmarshal(before, restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.Faith != nil || restored.Prophecies != nil {
		t.Error("pre-085 snapshot resurrected faith/prophecy state")
	}
	if string(restored.Marshal()) != string(before) {
		t.Error("pre-085 snapshot did not round-trip byte-identically")
	}
}

// faithEventsIn filters a batch to its faith.changed payloads, decoded.
func faithEventsIn(t *testing.T, evs []store.Event) []FaithChangedPayload {
	t.Helper()
	var out []FaithChangedPayload
	for _, e := range evs {
		if e.Type != "faith.changed" {
			continue
		}
		var p FaithChangedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			t.Fatal(err)
		}
		out = append(out, p)
	}
	return out
}

// TestFaithSweepDirectiveLifecycleInBatch (US1 AS-1/AS-2, SC-001): a directive
// fulfilling mints faith.changed{+8, directive_fulfilled, <id>} in the SAME
// tick batch as directive.fulfilled, and an expiring one mints −4 — driven
// through the real executor sweeps, then folded.
func TestFaithSweepDirectiveLifecycleInBatch(t *testing.T) {
	s := planState(t, 42)
	if err := s.Apply(placedEvent(validSite("dsg-1-0", 1), 1, 1)); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(issuedEvent(validDirective("dir-2-0", "dsg-1-0", []int{0}, 2), 2, 2)); err != nil {
		t.Fatal(err)
	}
	s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 10, Y: 10})
	s.Tick = 10

	// Tick 11: the designation fulfills — a bare designation.fulfilled mints
	// NOTHING (US1 AS-6, research R3).
	evs := stepEvents(s, s.m, 11)
	if countType(evs, "designation.fulfilled") != 1 || len(faithEventsIn(t, evs)) != 0 {
		t.Fatalf("designation fulfillment must mint no faith: %v", evs)
	}
	for _, e := range evs {
		if err := s.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	s.Tick = 11

	// Tick 12: the directive fulfills WITH its in-batch faith companion.
	evs = stepEvents(s, s.m, 12)
	if countType(evs, "directive.fulfilled") != 1 {
		t.Fatalf("directive did not fulfill: %v", evs)
	}
	faith := faithEventsIn(t, evs)
	if len(faith) != 1 || faith[0] != (FaithChangedPayload{Delta: FaithDeltaDirectiveFulfilled,
		Reason: FaithReasonDirectiveFulfilled, SourceID: "dir-2-0"}) {
		t.Fatalf("faith companion = %+v, want one {+8, directive_fulfilled, dir-2-0}", faith)
	}
	for _, e := range evs {
		if err := s.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	if got := s.FaithScore(); got != FaithGenesis+FaithDeltaDirectiveFulfilled {
		t.Fatalf("folded score = %d, want %d", got, FaithGenesis+FaithDeltaDirectiveFulfilled)
	}

	// Expiry: a second directive lapses at its TTL with the −4 companion.
	if err := s.Apply(placedEvent(validZone("dsg-3-0", 12, 12), 3, 12)); err != nil {
		t.Fatal(err)
	}
	d := validDirective("dir-13-0", "dsg-3-0", []int{0}, 13)
	d.ExpiresTick = 13 + ticksPerGameDay
	if err := s.Apply(issuedEvent(d, 4, 13)); err != nil {
		t.Fatal(err)
	}
	s.Tick = d.ExpiresTick - 1
	evs = stepEvents(s, s.m, d.ExpiresTick)
	if countType(evs, "directive.expired") != 1 {
		t.Fatalf("directive did not expire: %v", evs)
	}
	faith = faithEventsIn(t, evs)
	if len(faith) != 1 || faith[0] != (FaithChangedPayload{Delta: FaithDeltaDirectiveExpired,
		Reason: FaithReasonDirectiveExpired, SourceID: "dir-13-0"}) {
		t.Fatalf("expiry companion = %+v, want one {−4, directive_expired, dir-13-0}", faith)
	}
}

// TestFaithSweepDeathInBatch (US1 AS-3): a villager death mints one
// faith.changed{−6, villager_died, <index>} in the same batch as agent.died.
func TestFaithSweepDeathInBatch(t *testing.T) {
	s := NewState(42, testMap(42))
	s.Agents[1].Needs.Food = 0
	s.Agents[1].Needs.Health = 3
	log := driveTicks(t, s, s.m, 120, nil)
	deathTick := int64(-1)
	for _, e := range log {
		if e.Type == "agent.died" {
			deathTick = e.Tick
		}
	}
	if deathTick < 0 {
		t.Fatal("staged villager did not die")
	}
	faith := faithEventsIn(t, log)
	if len(faith) != 1 || faith[0] != (FaithChangedPayload{Delta: FaithDeltaVillagerDied,
		Reason: FaithReasonVillagerDied, SourceID: "1"}) {
		t.Fatalf("death companion = %+v, want one {−6, villager_died, \"1\"}", faith)
	}
	if got := s.FaithScore(); got != FaithGenesis+FaithDeltaVillagerDied {
		t.Fatalf("folded score = %d, want %d", got, FaithGenesis+FaithDeltaVillagerDied)
	}
}

// TestFaithSweepPileupOrderAndGate (FR-003, spec edge "same-boundary source
// pileup"): one faith.changed per source in the batch's own order; the
// cannot-move gate accounts for THIS batch's prior faith emissions; a partial
// clamp still emits the doctrine delta.
func TestFaithSweepPileupOrderAndGate(t *testing.T) {
	s := NewState(42, testMap(42))
	batch := []store.Event{
		{Tick: 9, Type: "directive.fulfilled", Payload: mustPayload(DirectiveFulfilledPayload{
			ID: "dir-a", DesignationID: "dsg-a", Targets: []int{0}, IssuedTick: 1})},
		{Tick: 9, Type: "directive.expired", Payload: mustPayload(OrderIDPayload{ID: "dir-b"})},
		{Tick: 9, Type: "agent.died", Payload: mustPayload(DiedPayload{Agent: Ref(2), Cause: "starvation"})},
	}

	t.Run("one per source, batch order", func(t *testing.T) {
		got := faithEventsIn(t, faithEvents(s, batch, 9))
		want := []FaithChangedPayload{
			{Delta: FaithDeltaDirectiveFulfilled, Reason: FaithReasonDirectiveFulfilled, SourceID: "dir-a"},
			{Delta: FaithDeltaDirectiveExpired, Reason: FaithReasonDirectiveExpired, SourceID: "dir-b"},
			{Delta: FaithDeltaVillagerDied, Reason: FaithReasonVillagerDied, SourceID: "2"},
		}
		if len(got) != len(want) {
			t.Fatalf("emitted %d faith events, want %d: %+v", len(got), len(want), got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("event %d = %+v, want %+v (batch order)", i, got[i], want[i])
			}
		}
	})

	t.Run("no-emit at the ceiling and the floor (US1 AS-4)", func(t *testing.T) {
		s.Faith = &FaithState{Score: 100}
		up := faithEvents(s, batch[:1], 9) // +8 at 100 moves nothing
		if len(up) != 0 {
			t.Errorf("score 100 + positive source emitted %d events, want 0", len(up))
		}
		s.Faith = &FaithState{Score: 0}
		down := faithEvents(s, batch[1:], 9) // −4 and −6 at 0 move nothing
		if len(down) != 0 {
			t.Errorf("score 0 + negative sources emitted %d events, want 0", len(down))
		}
	})

	t.Run("partial clamp still emits the doctrine delta", func(t *testing.T) {
		s.Faith = &FaithState{Score: 98}
		got := faithEventsIn(t, faithEvents(s, batch[:1], 9))
		if len(got) != 1 || got[0].Delta != FaithDeltaDirectiveFulfilled {
			t.Fatalf("partial clamp emitted %+v, want the doctrine +8 (the fold clamps)", got)
		}
	})

	t.Run("gate folds this batch's prior emissions", func(t *testing.T) {
		s.Faith = &FaithState{Score: 4}
		two := []store.Event{
			{Tick: 9, Type: "agent.died", Payload: mustPayload(DiedPayload{Agent: Ref(0), Cause: "starvation"})},
			{Tick: 9, Type: "agent.died", Payload: mustPayload(DiedPayload{Agent: Ref(1), Cause: "starvation"})},
		}
		got := faithEventsIn(t, faithEvents(s, two, 9))
		if len(got) != 1 || got[0].SourceID != "0" {
			t.Fatalf("got %+v, want exactly the first death (the second cannot move 0)", got)
		}
	})
}

// TestFaithReplayByteIdentical (T007, SC-001): a from-genesis replay of a log
// carrying the Phase-3 faith reasons (directive fulfilled, directive expired,
// villager death) reconstructs byte-identical state — the plans_test recovery
// idiom, with the death staging mirrored on both sides (the
// TestReplayRebuildsEnded pattern; the full five-reason lifecycle proof is
// TestProphecyLifecycleReplayByteIdentical).
func TestFaithReplayByteIdentical(t *testing.T) {
	const seed = 85
	const ticks = 90_000
	m := testMap(seed)
	stage := func(s *State) {
		s.Agents[3].Needs.Food = 0
		s.Agents[3].Needs.Health = 3
	}

	timeline := map[int64][]store.Event{
		50: {placedEvent(validSite("dsg-50-0", 50), 0, 50)},
		70: {placedEvent(validZone("dsg-70-0", 70, 12), 0, 70)},
		80: {issuedEvent(func() Directive {
			d := validDirective("dir-80-0", "dsg-50-0", []int{0, 1}, 80)
			d.ExpiresTick = 80 + 3*ticksPerGameDay
			return d
		}(), 0, 80)},
		90:  {{Tick: 90, Type: "metatron.item_granted", Payload: mustPayload(ItemGrantedPayload{Agent: Ref(0), Kind: "planks", Qty: 4})}},
		100: {{Tick: 100, Type: "agent.built", Payload: mustPayload(BuiltPayload{Agent: Ref(0), Kind: "shelter", X: 10, Y: 10})}},
		300: {issuedEvent(func() Directive {
			d := validDirective("dir-300-0", "dsg-70-0", []int{2}, 300)
			d.ExpiresTick = 300 + ticksPerGameDay
			return d
		}(), 0, 300)},
	}

	live := NewState(seed, m)
	stage(live)
	log := driveTicks(t, live, m, ticks, timeline)

	// Guard: the log carries all three Phase-3 reasons.
	reasons := map[string]int{}
	for _, p := range faithEventsIn(t, log) {
		reasons[p.Reason]++
	}
	for _, r := range []string{FaithReasonDirectiveFulfilled, FaithReasonDirectiveExpired, FaithReasonVillagerDied} {
		if reasons[r] == 0 {
			t.Fatalf("log carries no %s faith event (reasons: %v)", r, reasons)
		}
	}

	replayed := NewState(seed, m)
	stage(replayed)
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replayed.Tick = e.Tick
	}
	driveTicks(t, replayed, m, ticks, nil)

	if live.Hash() != replayed.Hash() {
		t.Fatalf("faith replay diverged from live:\nlive:     %s\nreplayed: %s",
			string(live.Marshal()), string(replayed.Marshal()))
	}
}

// TestPre085LogReplaysWithoutFaith (US1 AS-5): a log containing no faith
// events — the pre-085 world's log — replays under the new code with Faith
// still nil: faith derives EXCLUSIVELY from recorded faith.changed events,
// never retroactively from old directive.fulfilled rows.
func TestPre085LogReplaysWithoutFaith(t *testing.T) {
	s := NewState(42, testMap(42))
	// A hand-applied pre-085-style lifecycle: placement, issue, fulfillment —
	// applied WITHOUT the executor sweep, exactly as replaying an old log does.
	if err := s.Apply(placedEvent(validSite("dsg-1-0", 1), 1, 1)); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(issuedEvent(validDirective("dir-2-0", "dsg-1-0", []int{0}, 2), 2, 2)); err != nil {
		t.Fatal(err)
	}
	s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 10, Y: 10})
	if err := s.Apply(idEvent("designation.fulfilled", "dsg-1-0", 3)); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(idEvent("directive.fulfilled", "dir-2-0", 4)); err != nil {
		// directive.fulfilled carries the seam payload in real logs; the bare
		// id form is enough for the arm's re-validation here.
		t.Fatal(err)
	}
	if s.Faith != nil {
		t.Fatalf("replaying old events minted retroactive faith: %+v", s.Faith)
	}
	if got := s.FaithScore(); got != FaithGenesis {
		t.Fatalf("FaithScore() = %d, want untouched genesis %d", got, FaithGenesis)
	}
}

// TestFaithRegenCadenceTable (FR-004/FR-005, SC-002/SC-003): the band ×
// posture matrix — every cell of data-model §6, including the posture fork's
// both ways (the AC #4 proof at the function level; the drive-level proofs
// follow below).
func TestFaithRegenCadenceTable(t *testing.T) {
	cases := []struct {
		score    int
		scenario bool
		want     int64
	}{
		{100, false, 4 * 3600}, {75, false, 4 * 3600}, {75, true, 4 * 3600}, // fervent
		{74, false, 6 * 3600}, {50, false, 6 * 3600}, {40, true, 6 * 3600}, // steady (genesis band)
		{39, false, 12 * 3600}, {15, true, 12 * 3600}, // wavering
		{14, true, 0}, {0, true, 0}, // forsaken, scenario: the authentic spiral
		{14, false, 24 * 3600}, {0, false, 24 * 3600}, // forsaken, ambient: the floor
	}
	for _, c := range cases {
		if got := FaithRegenCadenceTicks(c.score, c.scenario); got != c.want {
			t.Errorf("FaithRegenCadenceTicks(%d, %v) = %d, want %d", c.score, c.scenario, got, c.want)
		}
	}
	if FaithRegenCadenceTicks(FaithGenesis, false) != GuardianChargeRegenTicks {
		t.Error("the genesis band must be today's exported constant (pre-085-identical schedule)")
	}
}

// regenFires probes whether stepEvents emits charge_regenerated at tick —
// pure, non-mutating (the TestGuardianChargeRegenTicksMatchesExecutor shape).
func regenFires(s *State, tick int64) bool {
	for _, e := range stepEvents(s, s.m, tick) {
		if e.Type == "metatron.charge_regenerated" {
			return true
		}
	}
	return false
}

// TestFaithRegenBandBoundaryDrive (US2 AS-1..4): per band, the band cadence's
// absolute boundary fires and the off-boundary tick does not — the
// TestChargeRegen clone over the whole curve, plus the at-cap gate.
func TestFaithRegenBandBoundaryDrive(t *testing.T) {
	bands := []struct {
		name    string
		score   int
		cadence int64
	}{
		{"fervent 4h", 90, 4 * 3600},
		{"steady 6h", FaithGenesis, 6 * 3600},
		{"wavering 12h", 20, 12 * 3600},
		{"ambient forsaken 24h floor", 5, 24 * 3600},
	}
	for _, b := range bands {
		t.Run(b.name, func(t *testing.T) {
			s := NewState(11, testMap(11))
			s.GuardianCharges = 1
			if b.score != FaithGenesis {
				s.Faith = &FaithState{Score: b.score}
			}
			if !regenFires(s, b.cadence) {
				t.Errorf("cadence boundary %d did not fire", b.cadence)
			}
			if regenFires(s, b.cadence+1) {
				t.Error("off-boundary tick fired")
			}
			s.GuardianCharges = GuardianChargeCap
			if regenFires(s, 2*b.cadence) {
				t.Error("at-cap boundary fired (the below-cap gate must hold in every band)")
			}
		})
	}
}

// TestScenarioForsakenNoRegen (US2 AS-4, SC-003): a scenario world in the
// forsaken band regenerates NOTHING across a multi-day sweep of every
// candidate boundary — the authentic spiral; the posture fork's scenario arm.
func TestScenarioForsakenNoRegen(t *testing.T) {
	s := NewState(11, testMap(11))
	if err := s.ArmScenario(FirstNightExercise); err != nil {
		t.Fatal(err)
	}
	s.GuardianCharges = 0
	s.Faith = &FaithState{Score: 5}
	for day := int64(0); day < 3; day++ {
		for _, boundary := range []int64{4 * 3600, 6 * 3600, 12 * 3600, 24 * 3600} {
			tick := day*ticksPerGameDay + boundary
			s.Tick = tick - 1
			if regenFires(s, tick) {
				t.Fatalf("scenario forsaken world regenerated at tick %d", tick)
			}
		}
	}
	// The band is the lock, not the world: recovering out of the band restores
	// the wavering cadence (the endogenous exit stays real).
	s.Faith = &FaithState{Score: 15}
	if !regenFires(s, 12*3600) {
		t.Error("recovered band did not resume regen")
	}
}

// TestGenesisBandScheduleByteIdentity (US2 AS-1, SC-002 — THE compat pin): a
// world that has never folded a faith event fires charge_regenerated on
// EXACTLY the pre-085 lattice — every multiple of the old 6-game-hour
// constant, and nothing else — probed across every candidate band boundary
// of two full game days, so any accidental band change or cadence drift for
// untouched worlds fails here first.
func TestGenesisBandScheduleByteIdentity(t *testing.T) {
	s := NewState(11, testMap(11))
	s.GuardianCharges = 1
	if s.Faith != nil {
		t.Fatal("fixture must have no faith state")
	}
	var probes []int64
	for day := int64(0); day < 2; day++ {
		for _, boundary := range []int64{4 * 3600, 6 * 3600, 8 * 3600, 12 * 3600, 16 * 3600, 18 * 3600, 24 * 3600} {
			tick := day*ticksPerGameDay + boundary
			probes = append(probes, tick, tick+1)
		}
	}
	const pre085Cadence = 6 * 3600 // the retired fixed constant, pinned literally
	for _, tick := range probes {
		want := tick%pre085Cadence == 0
		if got := regenFires(s, tick); got != want {
			t.Errorf("tick %d: fires = %v, want %v (pre-085 schedule)", tick, got, want)
		}
	}
}

// TestFaithBandShiftsRegenLive: folding faith events across a band edge
// changes the cadence at the next check — the curve reads the FOLDED score,
// deterministically.
func TestFaithBandShiftsRegenLive(t *testing.T) {
	s := NewState(11, testMap(11))
	s.GuardianCharges = 0
	// Two deaths from genesis: 50 → 38 (wavering).
	for i, src := range []string{"0", "1"} {
		if err := s.Apply(faithEvent(int64(i+1), FaithDeltaVillagerDied, FaithReasonVillagerDied, src)); err != nil {
			t.Fatal(err)
		}
	}
	if regenFires(s, 6*3600) {
		t.Error("wavering world still fires on the steady 6h boundary")
	}
	if !regenFires(s, 12*3600) {
		t.Error("wavering world did not fire on its 12h boundary")
	}
}

// TestRegenEventShapeUnchanged (FR-004): the emitted event keeps its type and
// EMPTY payload — only the firing cadence moved.
func TestRegenEventShapeUnchanged(t *testing.T) {
	s := NewState(11, testMap(11))
	s.GuardianCharges = 1
	for _, e := range stepEvents(s, s.m, 6*3600) {
		if e.Type == "metatron.charge_regenerated" {
			if string(e.Payload) != "{}" {
				t.Fatalf("payload = %s, want empty {}", e.Payload)
			}
			return
		}
	}
	t.Fatal("no charge_regenerated at the steady boundary")
}

// TestFaithInjectionRefused (FR-002): faith.changed is executor-emitted only —
// whitelist absence refuses an injected forgery at the mind's door.
func TestFaithInjectionRefused(t *testing.T) {
	for _, typ := range []string{"faith.changed", "prophecy.fulfilled", "prophecy.failed"} {
		if InjectableSocialEvent(typ) {
			t.Errorf("%s is injectable — it must be executor-emitted only", typ)
		}
	}
	if !InjectableSocialEvent("prophecy.declared") {
		t.Error("prophecy.declared must be injectable (the prophesy tool's door)")
	}
}
