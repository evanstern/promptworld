package sim

// Spec 085 prophecy suites (T015): the prophecy.declared door rejection table,
// the claim predicate table per cell (incl. survives fail-fast and
// late-truth-after-failed), terminal races, in-batch faith companions
// (+12/−15), companion-memory provenance stamps, retention prune, ended-world
// silence, and from-genesis replay byte-identity over the full lifecycle.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// validProphecy returns a legal prophecy payload with a structure_count claim
// (no shelter stands at genesis, so the claim is never already-true) and the
// default 3-game-day deadline.
func validProphecy(id string, tick int64, targets []int) Prophecy {
	return Prophecy{
		ID: id, Targets: targets, Text: "Before three dawns a shelter will stand.",
		Claim:        ProphecyClaim{Kind: ProphecyStructureCount, StructureKind: "shelter", Min: 1},
		DeclaredTick: tick, DeadlineTick: tick + 3*ticksPerGameDay,
	}
}

func declaredEvent(p Prophecy, seq, tick int64) store.Event {
	return store.Event{Seq: seq, Tick: tick, Type: "prophecy.declared", Payload: mustPayload(p)}
}

// TestProphecyDeclaredLandsActive: payload Status/PlacedSeq are IGNORED — the
// reducer lands active and stamps PlacedSeq from the event's own store seq
// (the designation.placed contract).
func TestProphecyDeclaredLandsActive(t *testing.T) {
	s := planState(t, 42)
	p := validProphecy("pro-100-0", 100, []int{0, 2})
	p.Status = "fulfilled" // ignored on the wire
	p.PlacedSeq = 999      // ignored on the wire
	if err := s.Apply(declaredEvent(p, 77, 100)); err != nil {
		t.Fatalf("declared: %v", err)
	}
	if len(s.Prophecies) != 1 {
		t.Fatalf("prophecies = %d, want 1", len(s.Prophecies))
	}
	got := s.Prophecies[0]
	if got.Status != "active" || got.PlacedSeq != 77 {
		t.Errorf("landed status=%q placedSeq=%d, want active/77", got.Status, got.PlacedSeq)
	}
	// The stake: the declaration spent the genesis charge (US3 AS-1).
	if s.GuardianCharges != GuardianGenesisCharges-1 {
		t.Errorf("charges = %d, want %d (a prophecy spends one)", s.GuardianCharges, GuardianGenesisCharges-1)
	}
	// And an empty bank refuses the next word at the door.
	p2 := validProphecy("pro-200-0", 200, []int{0})
	p2.Claim.Min = 2
	if err := s.Apply(declaredEvent(p2, 78, 200)); err == nil || !strings.Contains(err.Error(), "no charges banked") {
		t.Fatalf("err = %v, want the charge-gate refusal", err)
	}
}

// TestProphecyDeclaredValidationTable pins every data-model §4 rejection arm
// (validate-not-clamp — the dry-run is the door), including the already-true
// and active-duplicate refusals and the kind-conditional claim fields.
func TestProphecyDeclaredValidationTable(t *testing.T) {
	base := func() *State {
		s := planState(t, 42)
		s.GuardianCharges = GuardianChargeCap // multi-declaration cases need stakes
		if err := s.Apply(placedEvent(validSite("dsg-1-0", 1), 1, 1)); err != nil {
			t.Fatal(err)
		}
		return s
	}
	mutate := func(fn func(*Prophecy)) Prophecy {
		p := validProphecy("pro-100-0", 100, []int{0})
		fn(&p)
		return p
	}
	cases := []struct {
		name    string
		prep    func(*State)
		p       Prophecy
		wantErr string
	}{
		{"empty id", nil, mutate(func(p *Prophecy) { p.ID = "" }), "empty prophecy id"},
		{"no targets", nil, mutate(func(p *Prophecy) { p.Targets = nil }), "no targets"},
		{"target out of range", nil, mutate(func(p *Prophecy) { p.Targets = []int{99} }), "out of range"},
		{"targets not ascending-unique", nil, mutate(func(p *Prophecy) { p.Targets = []int{1, 1} }), "ascending-unique"},
		{"dead target", func(s *State) { s.Agents[0].Dead = true }, mutate(func(*Prophecy) {}), "is dead"},
		{"empty text", nil, mutate(func(p *Prophecy) { p.Text = "" }), "text length"},
		{"text over the cap", nil, mutate(func(p *Prophecy) { p.Text = strings.Repeat("x", NudgeTextMax+1) }), "text length"},
		{"ttl too short", nil, mutate(func(p *Prophecy) { p.DeadlineTick = p.DeclaredTick + ticksPerGameDay/2 }), "game days"},
		{"ttl too long", nil, mutate(func(p *Prophecy) { p.DeadlineTick = p.DeclaredTick + 8*ticksPerGameDay }), "game days"},
		{"unknown claim kind", nil, mutate(func(p *Prophecy) { p.Claim = ProphecyClaim{Kind: "weather"} }), "unknown claim kind"},
		{"structure_count missing kind", nil, mutate(func(p *Prophecy) { p.Claim.StructureKind = "" }), "needs a structure_kind"},
		{"structure_count unknown kind", nil, mutate(func(p *Prophecy) { p.Claim.StructureKind = "castle" }), "unknown structure kind"},
		{"structure_count min out of bounds", nil, mutate(func(p *Prophecy) { p.Claim.Min = 65 }), "outside 1..64"},
		{"structure_count foreign field", nil, mutate(func(p *Prophecy) { p.Claim.DesignationID = "dsg-1-0" }), "takes only"},
		{"designation_fulfilled missing id", nil, mutate(func(p *Prophecy) {
			p.Claim = ProphecyClaim{Kind: ProphecyDesignationFulfilled}
		}), "needs a designation_id"},
		{"designation_fulfilled unknown id", nil, mutate(func(p *Prophecy) {
			p.Claim = ProphecyClaim{Kind: ProphecyDesignationFulfilled, DesignationID: "dsg-none"}
		}), "unknown designation"},
		{"designation_fulfilled foreign field", nil, mutate(func(p *Prophecy) {
			p.Claim = ProphecyClaim{Kind: ProphecyDesignationFulfilled, DesignationID: "dsg-1-0", Min: 2}
		}), "takes only"},
		{"population min out of bounds", nil, mutate(func(p *Prophecy) {
			p.Claim = ProphecyClaim{Kind: ProphecyPopulationAtLeast, Min: 99}
		}), "outside 1.."},
		{"population already true", nil, mutate(func(p *Prophecy) {
			p.Claim = ProphecyClaim{Kind: ProphecyPopulationAtLeast, Min: 2}
		}), "already holds"},
		{"survives dead villager", func(s *State) { s.Agents[2].Dead = true }, mutate(func(p *Prophecy) {
			p.Claim = ProphecyClaim{Kind: ProphecySurvives, Agent: 2}
		}), "living villager"},
		{"survives foreign field", nil, mutate(func(p *Prophecy) {
			p.Claim = ProphecyClaim{Kind: ProphecySurvives, Agent: 1, Min: 3}
		}), "takes only"},
		{"already-true structural claim", func(s *State) {
			s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 3, Y: 3})
		}, mutate(func(*Prophecy) {}), "already holds"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := base()
			if c.prep != nil {
				c.prep(s)
			}
			err := s.Apply(declaredEvent(c.p, 1, 100))
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("err = %v, want %q", err, c.wantErr)
			}
			if len(s.Prophecies) != 0 {
				t.Fatal("a refused declaration landed")
			}
		})
	}

	t.Run("duplicate id in any status", func(t *testing.T) {
		s := base()
		if err := s.Apply(declaredEvent(validProphecy("pro-100-0", 100, []int{0}), 1, 100)); err != nil {
			t.Fatal(err)
		}
		s.Prophecies[0].Status = "failed" // even a settled prophecy holds its id
		dup := validProphecy("pro-100-0", 200, []int{0})
		dup.Claim.Min = 2 // different claim; the id is the collision
		if err := s.Apply(declaredEvent(dup, 2, 200)); err == nil || !strings.Contains(err.Error(), "duplicate prophecy id") {
			t.Fatalf("err = %v, want duplicate id refusal", err)
		}
	})

	t.Run("active duplicate claim refused", func(t *testing.T) {
		s := base()
		if err := s.Apply(declaredEvent(validProphecy("pro-100-0", 100, []int{0}), 1, 100)); err != nil {
			t.Fatal(err)
		}
		dup := validProphecy("pro-200-0", 200, []int{1, 2}) // same normalized claim, different targets
		if err := s.Apply(declaredEvent(dup, 2, 200)); err == nil || !strings.Contains(err.Error(), "already stakes that claim") {
			t.Fatalf("err = %v, want duplicate-claim refusal", err)
		}
		// A settled prophecy's claim is free to restake (the risk was consumed).
		s.Prophecies[0].Status = "failed"
		if err := s.Apply(declaredEvent(dup, 3, 300)); err != nil {
			t.Fatalf("restaking a settled claim refused: %v", err)
		}
	})

	t.Run("cap 3 active", func(t *testing.T) {
		s := base()
		s.GuardianCharges = 10    // enough stakes that the CAP is the refusal hit
		mins := []int{1, 2, 3, 4} // distinct claims, all unmet at genesis (no shelter)
		for i := 0; i < GuardianProphecyCap; i++ {
			p := validProphecy(idFor(i), 100, []int{0})
			p.Claim.Min = mins[i]
			if err := s.Apply(declaredEvent(p, int64(i+1), 100)); err != nil {
				t.Fatalf("declaration %d refused: %v", i, err)
			}
		}
		p := validProphecy("pro-100-9", 100, []int{0})
		p.Claim.Min = mins[3]
		if err := s.Apply(declaredEvent(p, 9, 100)); err == nil || !strings.Contains(err.Error(), "cap 3") {
			t.Fatalf("err = %v, want cap refusal", err)
		}
	})
}

func idFor(i int) string { return "pro-100-" + string(rune('0'+i)) }

// TestProphecySweepFulfilWithCompanions (US3 AS-2): the claim coming true
// yields ONE prophecy.fulfilled, the in-batch faith companion {+12}, and one
// OriginReport memory per living target — once; the next sweep is silent.
func TestProphecySweepFulfilWithCompanions(t *testing.T) {
	s := planState(t, 42)
	p := validProphecy("pro-10-0", 10, []int{0, 1, 2})
	if err := s.Apply(declaredEvent(p, 1, 10)); err != nil {
		t.Fatal(err)
	}
	s.Agents[2].Dead = true // a target dying later: companions must skip the dead
	s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 10, Y: 10})
	s.Tick = 20

	evs := stepEvents(s, s.m, 21)
	if countType(evs, "prophecy.fulfilled") != 1 || countType(evs, "prophecy.failed") != 0 {
		t.Fatalf("sweep: %d fulfilled / %d failed, want 1/0", countType(evs, "prophecy.fulfilled"), countType(evs, "prophecy.failed"))
	}
	faith := faithEventsIn(t, evs)
	if len(faith) != 1 || faith[0] != (FaithChangedPayload{Delta: FaithDeltaProphecyFulfilled,
		Reason: FaithReasonProphecyFulfilled, SourceID: "pro-10-0"}) {
		t.Fatalf("faith companion = %+v, want one {+12, prophecy_fulfilled, pro-10-0}", faith)
	}
	var memories []MemoryAddedPayload
	for _, e := range evs {
		if e.Type != "agent.memory_added" {
			continue
		}
		var mp MemoryAddedPayload
		if err := json.Unmarshal(e.Payload, &mp); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(mp.Text, "foretelling came true") {
			memories = append(memories, mp)
		}
	}
	if len(memories) != 2 {
		t.Fatalf("companion memories = %d, want 2 (living targets only)", len(memories))
	}
	for _, mp := range memories {
		if mp.Origin != OriginReport {
			t.Errorf("companion origin = %q, want %q (word spreads secondhand)", mp.Origin, OriginReport)
		}
		if DirectPerception(mp.Origin) {
			t.Error("an OriginReport companion classified as direct perception — the provenance gate would launder it")
		}
		if mp.Agent.ID == 2 {
			t.Error("a dead target received a companion memory")
		}
	}

	applyAll(t, s, evs)
	s.Tick = 21
	if got := s.FaithScore(); got != FaithGenesis+FaithDeltaProphecyFulfilled {
		t.Fatalf("folded score = %d, want %d", got, FaithGenesis+FaithDeltaProphecyFulfilled)
	}
	if s.Prophecies[0].Status != "fulfilled" {
		t.Fatalf("status = %q, want fulfilled", s.Prophecies[0].Status)
	}
	if n := countType(stepEvents(s, s.m, 22), "prophecy.fulfilled"); n != 0 {
		t.Fatalf("sweep re-emitted %d terminals (once-only broken)", n)
	}
}

// TestProphecySweepFailAtDeadline (US3 AS-3): the deadline passing unmet
// yields prophecy.failed, the −15 companion, negative-tone OriginReport
// memories — and a later truth mints NOTHING (one-way status; the "verifies
// after the TTL" edge).
func TestProphecySweepFailAtDeadline(t *testing.T) {
	s := planState(t, 42)
	p := validProphecy("pro-10-0", 10, []int{0})
	if err := s.Apply(declaredEvent(p, 1, 10)); err != nil {
		t.Fatal(err)
	}
	s.Tick = p.DeadlineTick - 1

	// One tick before the deadline: silence.
	if n := countType(stepEvents(s, s.m, p.DeadlineTick-1), "prophecy.failed"); n != 0 {
		t.Fatal("prophecy failed before its deadline")
	}
	evs := stepEvents(s, s.m, p.DeadlineTick)
	if countType(evs, "prophecy.failed") != 1 {
		t.Fatalf("deadline sweep emitted %d failed, want 1", countType(evs, "prophecy.failed"))
	}
	faith := faithEventsIn(t, evs)
	if len(faith) != 1 || faith[0] != (FaithChangedPayload{Delta: FaithDeltaProphecyFailed,
		Reason: FaithReasonProphecyFailed, SourceID: "pro-10-0"}) {
		t.Fatalf("faith companion = %+v, want one {−15, prophecy_failed}", faith)
	}
	sawMemory := false
	for _, e := range evs {
		if e.Type != "agent.memory_added" {
			continue
		}
		var mp MemoryAddedPayload
		if err := json.Unmarshal(e.Payload, &mp); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(mp.Text, "did not come to pass") {
			sawMemory = true
			if mp.Origin != OriginReport || mp.Tone >= 0 {
				t.Errorf("failed companion origin=%q tone=%d, want report/negative", mp.Origin, mp.Tone)
			}
		}
	}
	if !sawMemory {
		t.Fatal("no failed-word companion memory")
	}
	applyAll(t, s, evs)
	s.Tick = p.DeadlineTick

	// Late truth: the claim comes true AFTER failed latched — nothing mints.
	s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 10, Y: 10})
	late := stepEvents(s, s.m, s.Tick+1)
	if countType(late, "prophecy.fulfilled")+len(faithEventsIn(t, late)) != 0 {
		t.Fatalf("late truth minted something: %v", late)
	}
	// And an injected late fulfillment is refused by the one-way arm.
	err := s.Apply(store.Event{Tick: s.Tick + 1, Type: "prophecy.fulfilled",
		Payload: mustPayload(OrderIDPayload{ID: "pro-10-0"})})
	if err == nil || !strings.Contains(err.Error(), "not active") {
		t.Fatalf("late terminal err = %v, want one-way refusal", err)
	}
}

// TestProphecySurvivesFailFastAndFulfilAtDeadline (US3 AS-5): the survives
// kind fails FAST on death — no deadline wait — and fulfils at the first
// sweep tick ≥ deadline when the villager lives.
func TestProphecySurvivesFailFastAndFulfilAtDeadline(t *testing.T) {
	t.Run("fail-fast on death", func(t *testing.T) {
		s := planState(t, 42)
		p := validProphecy("pro-10-0", 10, []int{0})
		p.Claim = ProphecyClaim{Kind: ProphecySurvives, Agent: 1}
		if err := s.Apply(declaredEvent(p, 1, 10)); err != nil {
			t.Fatal(err)
		}
		s.Agents[1].Dead = true
		s.Tick = 20
		evs := stepEvents(s, s.m, 21) // far before the deadline
		if countType(evs, "prophecy.failed") != 1 {
			t.Fatalf("death did not fail-fast the survives claim: %v", evs)
		}
	})
	t.Run("fulfil at the deadline", func(t *testing.T) {
		s := planState(t, 42)
		p := validProphecy("pro-10-0", 10, []int{0})
		p.Claim = ProphecyClaim{Kind: ProphecySurvives, Agent: 1}
		if err := s.Apply(declaredEvent(p, 1, 10)); err != nil {
			t.Fatal(err)
		}
		s.Tick = p.DeadlineTick - 1
		if n := countType(stepEvents(s, s.m, p.DeadlineTick-1), "prophecy.fulfilled"); n != 0 {
			t.Fatal("survives fulfilled before its deadline")
		}
		if n := countType(stepEvents(s, s.m, p.DeadlineTick), "prophecy.fulfilled"); n != 1 {
			t.Fatal("survives did not fulfil at the deadline with the villager alive")
		}
	})
}

// TestProphecyFulfilBeforeFail (US3 AS-6): eligible for both terminals on one
// boundary — the claim turning true exactly at the deadline tick — fulfilled
// wins and exactly one terminal lands.
func TestProphecyFulfilBeforeFail(t *testing.T) {
	s := planState(t, 42)
	p := validProphecy("pro-10-0", 10, []int{0})
	if err := s.Apply(declaredEvent(p, 1, 10)); err != nil {
		t.Fatal(err)
	}
	s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 10, Y: 10})
	s.Tick = p.DeadlineTick - 1
	evs := stepEvents(s, s.m, p.DeadlineTick)
	f, x := countType(evs, "prophecy.fulfilled"), countType(evs, "prophecy.failed")
	if f != 1 || x != 0 {
		t.Fatalf("boundary sweep: fulfilled=%d failed=%d, want exactly one terminal (fulfilled)", f, x)
	}
}

// TestProphecyTerminalRaces: the one-way transition — a second terminal on a
// settled prophecy refuses at the door (the transitionGuardianOrder shape).
func TestProphecyTerminalRaces(t *testing.T) {
	s := planState(t, 42)
	p := validProphecy("pro-10-0", 10, []int{0})
	if err := s.Apply(declaredEvent(p, 1, 10)); err != nil {
		t.Fatal(err)
	}
	s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 10, Y: 10})
	if err := s.Apply(store.Event{Tick: 20, Type: "prophecy.fulfilled",
		Payload: mustPayload(OrderIDPayload{ID: "pro-10-0"})}); err != nil {
		t.Fatal(err)
	}
	err := s.Apply(store.Event{Tick: p.DeadlineTick, Type: "prophecy.failed",
		Payload: mustPayload(OrderIDPayload{ID: "pro-10-0"})})
	if err == nil {
		t.Fatal("second terminal was accepted (exactly one must ever land)")
	}
	if got := s.Prophecies[0].Status; got != "fulfilled" {
		t.Fatalf("status = %q, want the first terminal to stand", got)
	}
}

// TestProphecyTerminalRevalidation: the executor-emitted terminals re-validate
// their condition at the door — a forged fulfilled whose claim does not hold,
// or a forged failed before the deadline, refuses.
func TestProphecyTerminalRevalidation(t *testing.T) {
	s := planState(t, 42)
	p := validProphecy("pro-10-0", 10, []int{0})
	if err := s.Apply(declaredEvent(p, 1, 10)); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(store.Event{Tick: 20, Type: "prophecy.fulfilled",
		Payload: mustPayload(OrderIDPayload{ID: "pro-10-0"})}); err == nil {
		t.Fatal("fulfilled with an unmet claim was accepted")
	}
	if err := s.Apply(store.Event{Tick: 20, Type: "prophecy.failed",
		Payload: mustPayload(OrderIDPayload{ID: "pro-10-0"})}); err == nil {
		t.Fatal("failed before the deadline was accepted")
	}
}

// TestProphecyAllTargetsDeadStaysJudged (US3 AS-7): every hearer dying does
// NOT void the word — the prophecy stays active and is judged against the
// world; terminal companions simply skip the dead.
func TestProphecyAllTargetsDeadStaysJudged(t *testing.T) {
	s := planState(t, 42)
	p := validProphecy("pro-10-0", 10, []int{2})
	if err := s.Apply(declaredEvent(p, 1, 10)); err != nil {
		t.Fatal(err)
	}
	s.Agents[2].Dead = true
	s.Tick = 20
	if n := countType(stepEvents(s, s.m, 21), "prophecy.failed"); n != 0 {
		t.Fatal("all-targets-dead voided the prophecy (there is no such clause)")
	}
	s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 10, Y: 10})
	evs := stepEvents(s, s.m, 22)
	if countType(evs, "prophecy.fulfilled") != 1 {
		t.Fatal("orphaned prophecy was not judged")
	}
	if countType(evs, "agent.memory_added") != 0 {
		t.Error("a dead target received a terminal companion memory")
	}
}

// TestProphecyCancelledDesignationFailsAtDeadline (spec edge): a cancelled
// designation under a designation_fulfilled claim needs no special case —
// fulfil can never hold (one-way statuses), fail fires at the deadline.
func TestProphecyCancelledDesignationFailsAtDeadline(t *testing.T) {
	s := planState(t, 42)
	if err := s.Apply(placedEvent(validSite("dsg-1-0", 1), 1, 1)); err != nil {
		t.Fatal(err)
	}
	p := validProphecy("pro-10-0", 10, []int{0})
	p.Claim = ProphecyClaim{Kind: ProphecyDesignationFulfilled, DesignationID: "dsg-1-0"}
	if err := s.Apply(declaredEvent(p, 2, 10)); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(idEvent("designation.cancelled", "dsg-1-0", 20)); err != nil {
		t.Fatal(err)
	}
	s.Tick = p.DeadlineTick - 1
	if n := countType(stepEvents(s, s.m, p.DeadlineTick-1), "prophecy.failed"); n != 0 {
		t.Fatal("cancelled designation failed the prophecy before its deadline")
	}
	if n := countType(stepEvents(s, s.m, p.DeadlineTick), "prophecy.failed"); n != 1 {
		t.Fatal("cancelled designation's claim did not fail at the deadline")
	}
}

// TestProphecyEndedWorldSilence: the run-end latch freezes the verification
// sweep with everything else.
func TestProphecyEndedWorldSilence(t *testing.T) {
	s := planState(t, 42)
	p := validProphecy("pro-10-0", 10, []int{0})
	if err := s.Apply(declaredEvent(p, 1, 10)); err != nil {
		t.Fatal(err)
	}
	s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 10, Y: 10})
	s.Ended = true
	if evs := stepEvents(s, s.m, s.Tick+1); evs != nil {
		t.Fatalf("ended world emitted %d events", len(evs))
	}
}

// TestProphecyPruneDeterminism: active + most recent 32 settled retained (the
// prunePlanEntities discipline).
func TestProphecyPruneDeterminism(t *testing.T) {
	s := planState(t, 42)
	for i := 0; i < guardianOrderRetain+2; i++ {
		s.Prophecies = append(s.Prophecies, Prophecy{ID: "old-" + string(rune('A'+i%26)) + string(rune('0'+i/26)), Status: "failed"})
	}
	s.Prophecies = append(s.Prophecies, Prophecy{ID: "live", Status: "active",
		Claim: ProphecyClaim{Kind: ProphecyStructureCount, StructureKind: "shelter", Min: 2}})
	if err := s.Apply(declaredEvent(validProphecy("pro-99-0", 99, []int{0}), 1, 99)); err != nil {
		t.Fatal(err)
	}
	active, settled := 0, 0
	for i := range s.Prophecies {
		if s.Prophecies[i].Status == "active" {
			active++
		} else {
			settled++
		}
	}
	if active != 2 || settled != guardianOrderRetain {
		t.Fatalf("after prune: %d active / %d settled, want 2 / %d", active, settled, guardianOrderRetain)
	}
}

// TestProphecyRebaseTaxonomy (data-model §9): an ACTIVE prophecy's
// DeadlineTick SHIFTs across a time snap; DeclaredTick and settled prophecies
// KEEP; FaithState carries no tick and is untouched.
func TestProphecyRebaseTaxonomy(t *testing.T) {
	s := planState(t, 42)
	s.Faith = &FaithState{Score: 62}
	s.Prophecies = []Prophecy{
		{ID: "a", DeclaredTick: 100, DeadlineTick: 100 + ticksPerGameDay, Status: "active"},
		{ID: "b", DeclaredTick: 100, DeadlineTick: 100 + ticksPerGameDay, Status: "failed"},
	}
	rebaseTicks(s, 5000)
	if got := s.Prophecies[0].DeadlineTick; got != 100+ticksPerGameDay+5000 {
		t.Errorf("active DeadlineTick = %d, want shifted", got)
	}
	if s.Prophecies[0].DeclaredTick != 100 {
		t.Error("DeclaredTick moved (must KEEP)")
	}
	if got := s.Prophecies[1].DeadlineTick; got != 100+ticksPerGameDay {
		t.Errorf("settled DeadlineTick = %d, want unshifted (KEEP)", got)
	}
	if s.Faith.Score != 62 {
		t.Error("FaithState moved across a rebase (it carries no tick fields)")
	}
}

// TestProphecyLifecycleReplayByteIdentical (SC-001, FR-014): a from-genesis
// replay of a log carrying the FULL faith lifecycle — every faith reason
// (directive fulfilled/expired, a death, prophecy fulfilled AND failed) —
// reconstructs byte-identical state with no guardian running.
func TestProphecyLifecycleReplayByteIdentical(t *testing.T) {
	const seed = 118
	const ticks = 90_000 // past the failing prophecy's 1-day deadline (86,460) and dir-300's TTL (86,700)
	m := testMap(seed)
	stage := func(s *State) {
		s.Agents[3].Needs.Food = 0
		s.Agents[3].Needs.Health = 3
		// Two prophecies and the miracle grant all land before the first regen
		// boundary — bank the stakes on BOTH sides (the death-staging pattern:
		// test scaffolding standing in for recorded history).
		s.GuardianCharges = 3
	}

	fulfilling := Prophecy{ // structure_count met when the shelter builds
		ID: "pro-40-0", Targets: []int{0, 1}, Text: "Before three dawns a shelter will stand.",
		Claim:        ProphecyClaim{Kind: ProphecyStructureCount, StructureKind: "shelter", Min: 1},
		DeclaredTick: 40, DeadlineTick: 40 + 3*ticksPerGameDay,
	}
	failing := Prophecy{ // an oven never comes; fails at 60 + 1 game day
		ID: "pro-60-0", Targets: []int{0}, Text: "Before dawn an oven will stand.",
		Claim:        ProphecyClaim{Kind: ProphecyStructureCount, StructureKind: "oven", Min: 1},
		DeclaredTick: 60, DeadlineTick: 60 + ticksPerGameDay,
	}
	timeline := map[int64][]store.Event{
		40: {declaredEvent(fulfilling, 0, 40)},
		50: {placedEvent(validSite("dsg-50-0", 50), 0, 50)},
		60: {declaredEvent(failing, 0, 60)},
		80: {issuedEvent(func() Directive {
			d := validDirective("dir-80-0", "dsg-50-0", []int{0, 1}, 80)
			d.ExpiresTick = 80 + 3*ticksPerGameDay
			return d
		}(), 0, 80)},
		90:  {{Tick: 90, Type: "guardian.item_granted", Payload: mustPayload(ItemGrantedPayload{Agent: Ref(0), Kind: "planks", Qty: 4})}},
		100: {{Tick: 100, Type: "agent.built", Payload: mustPayload(BuiltPayload{Agent: Ref(0), Kind: "shelter", X: 10, Y: 10})}},
		200: {placedEvent(validZone("dsg-200-0", 200, 12), 0, 200)},
		300: {issuedEvent(func() Directive {
			d := validDirective("dir-300-0", "dsg-200-0", []int{2}, 300)
			d.ExpiresTick = 300 + ticksPerGameDay
			return d
		}(), 0, 300)},
	}

	live := NewState(seed, m)
	stage(live)
	log := driveTicks(t, live, m, ticks, timeline)

	// Guard: the log carries the full lifecycle — every reason exactly once.
	reasons := map[string]int{}
	for _, p := range faithEventsIn(t, log) {
		reasons[p.Reason]++
	}
	for _, r := range []string{FaithReasonDirectiveFulfilled, FaithReasonDirectiveExpired,
		FaithReasonVillagerDied, FaithReasonProphecyFulfilled, FaithReasonProphecyFailed} {
		if reasons[r] != 1 {
			t.Fatalf("log carries %d %s faith events, want exactly 1 (reasons: %v)", reasons[r], r, reasons)
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
		t.Fatalf("full-lifecycle replay diverged from live:\nlive:     %s\nreplayed: %s",
			string(live.Marshal()), string(replayed.Marshal()))
	}
	// The final score is the arithmetic of the recorded lifecycle.
	want := clampFaith(FaithGenesis + FaithDeltaDirectiveFulfilled + FaithDeltaDirectiveExpired +
		FaithDeltaVillagerDied + FaithDeltaProphecyFulfilled + FaithDeltaProphecyFailed)
	if got := live.FaithScore(); got != want {
		t.Fatalf("final score = %d, want %d", got, want)
	}
}
