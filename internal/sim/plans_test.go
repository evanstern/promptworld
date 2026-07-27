package sim

// Spec 084 Phase 2 suites (T008): reducer arm validation tables, one-way
// transition races, prune determinism, the announcement grant, the executor
// sweeps (once-only, fixed order, one-tick lag, single terminal, ended-world
// silence), the rebase taxonomy, and from-genesis replay byte-identity over a
// full designation/directive lifecycle log.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// planState builds a fresh mapped state for arm/sweep tests: living agents
// with mental maps, world map installed (bounds checks live).
func planState(t *testing.T, seed uint64) *State {
	t.Helper()
	m := testMap(seed)
	return NewState(seed, m)
}

// validSite returns a legal structure-site designation payload.
func validSite(id string, tick int64) Designation {
	return Designation{ID: id, Kind: DesignationStructureSite, X: 10, Y: 10, X2: 10, Y2: 10,
		StructureKind: "shelter", PlacedTick: tick}
}

// validLine returns a legal wall-line designation payload.
func validLine(id string, tick int64) Designation {
	return Designation{ID: id, Kind: DesignationWallLine, X: 12, Y: 10, X2: 14, Y2: 10, PlacedTick: tick}
}

// validZone returns a legal settlement-zone designation payload.
func validZone(id string, tick int64, min int) Designation {
	return Designation{ID: id, Kind: DesignationSettlementZone, X: 20, Y: 20, X2: 23, Y2: 23,
		MinStructures: min, PlacedTick: tick}
}

func placedEvent(d Designation, seq, tick int64) store.Event {
	return store.Event{Seq: seq, Tick: tick, Type: "designation.placed", Payload: mustPayload(d)}
}

// validDirective returns a legal directive payload bound to designation dsg.
func validDirective(id, dsg string, targets []int, tick int64) Directive {
	return Directive{ID: id, DesignationID: dsg, Targets: targets,
		Text: "Raise what I have marked.", IssuedTick: tick, ExpiresTick: tick + 3*ticksPerGameDay}
}

func issuedEvent(d Directive, seq, tick int64) store.Event {
	return store.Event{Seq: seq, Tick: tick, Type: "directive.issued", Payload: mustPayload(d)}
}

func idEvent(typ, id string, tick int64) store.Event {
	return store.Event{Tick: tick, Type: typ, Payload: mustPayload(OrderIDPayload{ID: id})}
}

// TestDesignationPlacedLandsActive: the payload's Status and PlacedSeq are
// IGNORED — the reducer lands active and stamps PlacedSeq from the event's
// own store seq (contracts §1) — and the announcement grant reaches every
// living villager's map (anchor tile, provenance revealed, Seen = e.Tick),
// skipping the dead and the map-less (the reducer stays total).
func TestDesignationPlacedLandsActive(t *testing.T) {
	s := planState(t, 42)
	s.Agents[1].Dead = true
	s.Agents[2].Map = nil

	d := validSite("dsg-100-0", 100)
	d.Status = "fulfilled" // ignored on the wire
	d.PlacedSeq = 999      // ignored on the wire
	d.Label = "north shelter"
	if err := s.Apply(placedEvent(d, 77, 100)); err != nil {
		t.Fatalf("placed: %v", err)
	}
	if len(s.Designations) != 1 {
		t.Fatalf("designations = %d, want 1", len(s.Designations))
	}
	got := s.Designations[0]
	if got.Status != "active" || got.PlacedSeq != 77 {
		t.Errorf("landed status=%q placedSeq=%d, want active/77 (payload values must be ignored)", got.Status, got.PlacedSeq)
	}
	for i := range s.Agents {
		a := &s.Agents[i]
		f, ok := PlaceFact{}, false
		if a.Map != nil {
			f, ok = a.Map.factAt("designation", 10, 10)
		}
		switch {
		case i == 1: // dead — skipped
			if ok {
				t.Errorf("dead agent %d received the grant", i)
			}
		case i == 2: // map-less — skipped, no panic
			// nothing to assert beyond not panicking
		default:
			if !ok {
				t.Errorf("living agent %d missing the designation fact", i)
				continue
			}
			if f.Seen != 100 || f.Provenance != ProvenanceRevealed {
				t.Errorf("agent %d fact = %+v, want Seen 100 / revealed", i, f)
			}
		}
	}
	// The designation horizon is the max directive TTL (data-model §5).
	if factHorizon("designation") != GuardianOrderTTLMaxDays*ticksPerGameDay {
		t.Errorf("factHorizon(designation) = %d, want %d", factHorizon("designation"), GuardianOrderTTLMaxDays*ticksPerGameDay)
	}
}

// TestDesignationPlacedValidationTable pins every contracts-§1 rejection arm
// (validate-not-clamp — the dry-run is the door).
func TestDesignationPlacedValidationTable(t *testing.T) {
	occupied := validSite("dsg-1-0", 1) // a different-kind structure on the tile
	line := validLine("dsg-1-1", 1)
	zone := validZone("dsg-1-2", 1, 3)

	cases := []struct {
		name string
		prep func(s *State)
		d    Designation
		want string
	}{
		{"empty id", nil, Designation{Kind: DesignationStructureSite, StructureKind: "shelter"}, "empty designation id"},
		{"unknown kind", nil, Designation{ID: "d", Kind: "camp"}, "unknown designation kind"},
		{"site not a point", nil, Designation{ID: "d", Kind: DesignationStructureSite, X: 1, Y: 1, X2: 2, Y2: 1, StructureKind: "shelter"}, "one tile"},
		{"site missing structure_kind", nil, Designation{ID: "d", Kind: DesignationStructureSite, X: 1, Y: 1, X2: 1, Y2: 1}, "needs a structure_kind"},
		{"site unknown structure_kind", nil, Designation{ID: "d", Kind: DesignationStructureSite, X: 1, Y: 1, X2: 1, Y2: 1, StructureKind: "palace"}, "unknown structure kind"},
		{"site min_structures", nil, Designation{ID: "d", Kind: DesignationStructureSite, X: 1, Y: 1, X2: 1, Y2: 1, StructureKind: "shelter", MinStructures: 3}, "settlement zones only"},
		{"diagonal line", nil, Designation{ID: "d", Kind: DesignationWallLine, X: 1, Y: 1, X2: 3, Y2: 3}, "axis-aligned"},
		{"line too long", nil, Designation{ID: "d", Kind: DesignationWallLine, X: 0, Y: 1, X2: 40, Y2: 1}, "exceeds the 32-tile cap"},
		{"line bad narrowing", nil, Designation{ID: "d", Kind: DesignationWallLine, X: 1, Y: 1, X2: 3, Y2: 1, StructureKind: "shelter"}, "not a wall kind"},
		{"zone unnormalized", nil, Designation{ID: "d", Kind: DesignationSettlementZone, X: 5, Y: 5, X2: 3, Y2: 6, MinStructures: 3}, "not normalized"},
		{"zone too big", nil, Designation{ID: "d", Kind: DesignationSettlementZone, X: 0, Y: 0, X2: 20, Y2: 20, MinStructures: 3}, "exceeds the 256-tile cap"},
		{"zone structure_kind", nil, Designation{ID: "d", Kind: DesignationSettlementZone, X: 1, Y: 1, X2: 2, Y2: 2, StructureKind: "shelter", MinStructures: 3}, "takes no structure_kind"},
		{"zone min out of bounds", nil, Designation{ID: "d", Kind: DesignationSettlementZone, X: 1, Y: 1, X2: 2, Y2: 2, MinStructures: 13}, "outside 1..12"},
		{"zone min zero", nil, Designation{ID: "d", Kind: DesignationSettlementZone, X: 1, Y: 1, X2: 2, Y2: 2}, "outside 1..12"},
		{"negative coordinate", nil, Designation{ID: "d", Kind: DesignationStructureSite, X: -1, Y: 1, X2: -1, Y2: 1, StructureKind: "shelter"}, "negative locus"},
		{"out of bounds", nil, Designation{ID: "d", Kind: DesignationStructureSite, X: 999, Y: 999, X2: 999, Y2: 999, StructureKind: "shelter"}, "outside the world"},
		{"label too long", nil, func() Designation {
			d := validSite("d", 1)
			d.Label = strings.Repeat("x", 81)
			return d
		}(), "label over 80 runes"},
		{"duplicate id any status", func(s *State) {
			if err := s.Apply(placedEvent(validSite("dsg-1-0", 1), 1, 1)); err != nil {
				t.Fatal(err)
			}
			if err := s.Apply(idEvent("designation.cancelled", "dsg-1-0", 2)); err != nil {
				t.Fatal(err)
			}
		}, occupied, "duplicate designation id"},
		{"site occupied by different kind", func(s *State) {
			s.Structures = append(s.Structures, Structure{Kind: "oven", X: 10, Y: 10})
		}, occupied, "already stands at (10,10)"},
		{"line tile holds non-wall", func(s *State) {
			s.Structures = append(s.Structures, Structure{Kind: "oven", X: 13, Y: 10})
		}, line, "stands on the line"},
		{"zone occupancy is free", func(s *State) {
			s.Structures = append(s.Structures, Structure{Kind: "oven", X: 21, Y: 21})
		}, zone, ""}, // accepted — zones are never occupancy-checked
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := planState(t, 42)
			if tc.prep != nil {
				tc.prep(s)
			}
			err := s.Apply(placedEvent(tc.d, 5, 5))
			if tc.want == "" {
				if err != nil {
					t.Fatalf("want accept, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want substring %q", err, tc.want)
			}
		})
	}
}

// TestDesignationSameKindPreexistingAccepted: the guardian may consecrate what
// stands — a same-kind structure on the site is legal, and the sweep fulfills
// at the next tick boundary (spec edge case).
func TestDesignationSameKindPreexistingAccepted(t *testing.T) {
	s := planState(t, 42)
	s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 10, Y: 10})
	if err := s.Apply(placedEvent(validSite("dsg-1-0", 1), 1, 1)); err != nil {
		t.Fatalf("consecrating placement refused: %v", err)
	}
	evs := stepEvents(s, s.m, s.Tick+1)
	if n := countType(evs, "designation.fulfilled"); n != 1 {
		t.Fatalf("next-boundary sweep emitted %d designation.fulfilled, want 1", n)
	}
}

// TestDesignationCap: at most 16 active designations; the 17th is refused at
// the door (validate-not-clamp), and cancelling one frees the slot.
func TestDesignationCap(t *testing.T) {
	s := planState(t, 42)
	for i := 0; i < GuardianDesignationCap; i++ {
		d := validZone("", 1, 3)
		d.ID = "dsg-1-" + string(rune('a'+i))
		d.X, d.X2 = i, i+1 // distinct, all in-bounds
		if err := s.Apply(placedEvent(d, int64(i), 1)); err != nil {
			t.Fatalf("placement %d: %v", i, err)
		}
	}
	over := validSite("dsg-2-0", 2)
	if err := s.Apply(placedEvent(over, 20, 2)); err == nil || !strings.Contains(err.Error(), "cap 16") {
		t.Fatalf("17th placement err = %v, want cap refusal", err)
	}
	if err := s.Apply(idEvent("designation.cancelled", "dsg-1-a", 3)); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(placedEvent(over, 21, 3)); err != nil {
		t.Fatalf("post-cancel placement refused: %v", err)
	}
}

// TestDirectiveIssuedValidationTable pins the directive.issued door.
func TestDirectiveIssuedValidationTable(t *testing.T) {
	const tick = 100
	cases := []struct {
		name string
		mut  func(d *Directive, s *State)
		want string
	}{
		{"empty id", func(d *Directive, s *State) { d.ID = "" }, "empty directive id"},
		{"unknown designation", func(d *Directive, s *State) { d.DesignationID = "dsg-none" }, "unknown designation"},
		{"non-active designation", func(d *Directive, s *State) {
			if err := s.Apply(idEvent("designation.cancelled", d.DesignationID, tick)); err != nil {
				t.Fatal(err)
			}
		}, "not active"},
		{"no targets", func(d *Directive, s *State) { d.Targets = nil }, "no targets"},
		{"target out of range", func(d *Directive, s *State) { d.Targets = []int{99} }, "out of range"},
		{"targets not ascending", func(d *Directive, s *State) { d.Targets = []int{1, 0} }, "ascending"},
		{"targets duplicate", func(d *Directive, s *State) { d.Targets = []int{1, 1} }, "ascending"},
		{"dead target", func(d *Directive, s *State) { s.Agents[0].Dead = true }, "is dead"},
		{"empty text", func(d *Directive, s *State) { d.Text = "" }, "outside 1..400"},
		{"text too long", func(d *Directive, s *State) { d.Text = strings.Repeat("x", 401) }, "outside 1..400"},
		{"ttl too short", func(d *Directive, s *State) { d.ExpiresTick = d.IssuedTick + ticksPerGameDay/2 }, "outside 1..7"},
		{"ttl too long", func(d *Directive, s *State) { d.ExpiresTick = d.IssuedTick + 8*ticksPerGameDay }, "outside 1..7"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := planState(t, 42)
			if err := s.Apply(placedEvent(validSite("dsg-1-0", 1), 1, 1)); err != nil {
				t.Fatal(err)
			}
			d := validDirective("dir-100-0", "dsg-1-0", []int{0, 1}, tick)
			tc.mut(&d, s)
			err := s.Apply(issuedEvent(d, 9, tick))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want substring %q", err, tc.want)
			}
		})
	}

	// The happy path: lands active, PlacedSeq stamped, payload Status ignored.
	s := planState(t, 42)
	if err := s.Apply(placedEvent(validSite("dsg-1-0", 1), 1, 1)); err != nil {
		t.Fatal(err)
	}
	d := validDirective("dir-100-0", "dsg-1-0", []int{0, 1}, tick)
	d.Status = "expired" // ignored
	if err := s.Apply(issuedEvent(d, 55, tick)); err != nil {
		t.Fatalf("valid issue refused: %v", err)
	}
	if got := s.Directives[0]; got.Status != "active" || got.PlacedSeq != 55 {
		t.Errorf("landed status=%q seq=%d, want active/55", got.Status, got.PlacedSeq)
	}

	// Duplicate id in any status refused (issue again after cancel).
	if err := s.Apply(idEvent("directive.cancelled", "dir-100-0", tick+1)); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(issuedEvent(validDirective("dir-100-0", "dsg-1-0", []int{0}, tick+1), 56, tick+1)); err == nil ||
		!strings.Contains(err.Error(), "duplicate directive id") {
		t.Errorf("duplicate id err = %v", err)
	}
}

// TestDirectiveCap: at most 3 active directives.
func TestDirectiveCap(t *testing.T) {
	s := planState(t, 42)
	if err := s.Apply(placedEvent(validSite("dsg-1-0", 1), 1, 1)); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < GuardianDirectiveCap; i++ {
		d := validDirective("dir-2-"+string(rune('a'+i)), "dsg-1-0", []int{i}, 2)
		if err := s.Apply(issuedEvent(d, int64(i), 2)); err != nil {
			t.Fatalf("issue %d: %v", i, err)
		}
	}
	over := validDirective("dir-3-0", "dsg-1-0", []int{4}, 3)
	if err := s.Apply(issuedEvent(over, 9, 3)); err == nil || !strings.Contains(err.Error(), "cap 3") {
		t.Fatalf("4th issue err = %v, want cap refusal", err)
	}
}

// TestPlanOneWayRaces: exactly one terminal ever lands — the loser of any
// cancel/fulfil/expire race finds a non-active entity and is refused
// (the transitionGuardianOrder shape).
func TestPlanOneWayRaces(t *testing.T) {
	setup := func(t *testing.T) *State {
		s := planState(t, 42)
		if err := s.Apply(placedEvent(validSite("dsg-1-0", 1), 1, 1)); err != nil {
			t.Fatal(err)
		}
		d := validDirective("dir-2-0", "dsg-1-0", []int{0}, 2)
		if err := s.Apply(issuedEvent(d, 2, 2)); err != nil {
			t.Fatal(err)
		}
		return s
	}

	t.Run("designation cancel then fulfil", func(t *testing.T) {
		s := setup(t)
		s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 10, Y: 10})
		if err := s.Apply(idEvent("designation.cancelled", "dsg-1-0", 3)); err != nil {
			t.Fatal(err)
		}
		if err := s.Apply(idEvent("designation.fulfilled", "dsg-1-0", 3)); err == nil ||
			!strings.Contains(err.Error(), "not active") {
			t.Errorf("loser err = %v, want not-active refusal", err)
		}
	})
	t.Run("designation fulfil then cancel", func(t *testing.T) {
		s := setup(t)
		s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 10, Y: 10})
		if err := s.Apply(idEvent("designation.fulfilled", "dsg-1-0", 3)); err != nil {
			t.Fatal(err)
		}
		if err := s.Apply(idEvent("designation.cancelled", "dsg-1-0", 3)); err == nil ||
			!strings.Contains(err.Error(), "not active") {
			t.Errorf("loser err = %v", err)
		}
	})
	t.Run("designation double cancel", func(t *testing.T) {
		s := setup(t)
		if err := s.Apply(idEvent("designation.cancelled", "dsg-1-0", 3)); err != nil {
			t.Fatal(err)
		}
		if err := s.Apply(idEvent("designation.cancelled", "dsg-1-0", 4)); err == nil ||
			!strings.Contains(err.Error(), "not active") {
			t.Errorf("second cancel err = %v", err)
		}
	})
	t.Run("directive cancel then expire", func(t *testing.T) {
		s := setup(t)
		if err := s.Apply(idEvent("directive.cancelled", "dir-2-0", 3)); err != nil {
			t.Fatal(err)
		}
		e := idEvent("directive.expired", "dir-2-0", 2+8*ticksPerGameDay)
		if err := s.Apply(e); err == nil || !strings.Contains(err.Error(), "not active") {
			t.Errorf("loser err = %v", err)
		}
	})
	t.Run("unknown ids refused", func(t *testing.T) {
		s := setup(t)
		if err := s.Apply(idEvent("designation.cancelled", "dsg-none", 3)); err == nil ||
			!strings.Contains(err.Error(), "unknown designation") {
			t.Errorf("err = %v", err)
		}
		if err := s.Apply(idEvent("directive.cancelled", "dir-none", 3)); err == nil ||
			!strings.Contains(err.Error(), "unknown directive") {
			t.Errorf("err = %v", err)
		}
	})
}

// TestPlanTerminalRevalidation: the executor-emitted terminals re-validate
// their emitting condition at the door — a forged fulfilled/expired that does
// not hold structurally is refused even though the type itself is sim-authored.
func TestPlanTerminalRevalidation(t *testing.T) {
	s := planState(t, 42)
	if err := s.Apply(placedEvent(validSite("dsg-1-0", 1), 1, 1)); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(issuedEvent(validDirective("dir-2-0", "dsg-1-0", []int{0}, 2), 2, 2)); err != nil {
		t.Fatal(err)
	}
	// designation.fulfilled without the structure standing: predicate fails.
	if err := s.Apply(idEvent("designation.fulfilled", "dsg-1-0", 3)); err == nil ||
		!strings.Contains(err.Error(), "predicate does not hold") {
		t.Errorf("fulfilled err = %v", err)
	}
	// directive.fulfilled while the designation is still active: refused.
	fe := store.Event{Tick: 3, Type: "directive.fulfilled", Payload: mustPayload(DirectiveFulfilledPayload{
		ID: "dir-2-0", DesignationID: "dsg-1-0", Targets: []int{0}, IssuedTick: 2})}
	if err := s.Apply(fe); err == nil || !strings.Contains(err.Error(), "not fulfilled") {
		t.Errorf("directive.fulfilled err = %v", err)
	}
	// directive.expired before TTL with living targets: refused.
	if err := s.Apply(idEvent("directive.expired", "dir-2-0", 3)); err == nil ||
		!strings.Contains(err.Error(), "neither past its TTL nor targetless") {
		t.Errorf("expired err = %v", err)
	}
	// All targets dead: expiry lands even before the TTL (the un-executable
	// clause is a pure state check, no TTL wait).
	s.Agents[0].Dead = true
	if err := s.Apply(idEvent("directive.expired", "dir-2-0", 4)); err != nil {
		t.Errorf("all-dead expiry refused: %v", err)
	}
}

// TestPlanPruneDeterminism: the retention prune keeps every active entity plus
// the most recent 32 non-active, dropping oldest first, order preserved — the
// pruneGuardianOrders algorithm generalized.
func TestPlanPruneDeterminism(t *testing.T) {
	var items []Designation
	items = append(items, Designation{ID: "active-early", Status: "active"})
	for i := 0; i < 40; i++ {
		items = append(items, Designation{ID: "done-" + string(rune('A'+i)), Status: "cancelled"})
	}
	items = append(items, Designation{ID: "active-late", Status: "active"})
	out := prunePlanEntities(items, func(d Designation) bool { return d.Status == "active" })
	if len(out) != 34 { // 2 active + 32 retained non-active
		t.Fatalf("len = %d, want 34", len(out))
	}
	if out[0].ID != "active-early" || out[len(out)-1].ID != "active-late" {
		t.Errorf("actives not preserved in order: first %q last %q", out[0].ID, out[len(out)-1].ID)
	}
	if out[1].ID != "done-"+string(rune('A'+8)) {
		t.Errorf("oldest retained non-active = %q, want done-%c (drop oldest 8 of 40)", out[1].ID, 'A'+8)
	}
}

// countType lives in hail_test.go (shared package-test helper).

// TestPlanSweepOnceOnlyAndLag: the fixed sweep order (designations first, then
// directives) means a designation fulfilled at tick T yields its bound
// directive's fulfilled at T+1 — the documented one-tick lag — and each event
// fires exactly once (the landed event flips the entity non-active).
func TestPlanSweepOnceOnlyAndLag(t *testing.T) {
	s := planState(t, 42)
	if err := s.Apply(placedEvent(validSite("dsg-1-0", 1), 1, 1)); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(issuedEvent(validDirective("dir-2-0", "dsg-1-0", []int{0}, 2), 2, 2)); err != nil {
		t.Fatal(err)
	}
	s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 10, Y: 10})
	s.Tick = 10

	// Tick 11: designation fulfills; the directive does NOT (it reads the
	// designation's pre-tick status, still active).
	evs := stepEvents(s, s.m, 11)
	if countType(evs, "designation.fulfilled") != 1 || countType(evs, "directive.fulfilled") != 0 {
		t.Fatalf("tick 11 sweep: %d dsg-fulfilled / %d dir-fulfilled, want 1/0 (one-tick lag)",
			countType(evs, "designation.fulfilled"), countType(evs, "directive.fulfilled"))
	}
	for _, e := range evs {
		if err := s.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	s.Tick = 11

	// Tick 12: the directive fulfills, once, with the TASK-118 seam payload.
	evs = stepEvents(s, s.m, 12)
	if countType(evs, "designation.fulfilled") != 0 || countType(evs, "directive.fulfilled") != 1 {
		t.Fatalf("tick 12 sweep: %d/%d, want 0 dsg / 1 dir",
			countType(evs, "designation.fulfilled"), countType(evs, "directive.fulfilled"))
	}
	for _, e := range evs {
		if e.Type == "directive.fulfilled" {
			want := string(mustPayload(DirectiveFulfilledPayload{
				ID: "dir-2-0", DesignationID: "dsg-1-0", Targets: []int{0}, IssuedTick: 2}))
			if string(e.Payload) != want {
				t.Errorf("payload = %s, want %s (the TASK-118 seam)", e.Payload, want)
			}
		}
		if err := s.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	s.Tick = 12

	// Tick 13: silence — both entities are non-active, the sweep skips them.
	evs = stepEvents(s, s.m, 13)
	if countType(evs, "designation.fulfilled")+countType(evs, "directive.fulfilled")+
		countType(evs, "directive.expired") != 0 {
		t.Fatalf("tick 13 sweep re-emitted plan events: %v", evs)
	}
}

// TestPlanSweepSingleTerminal: a directive eligible for BOTH fulfillment and
// expiry at one boundary lands exactly ONE terminal — fulfilled wins.
func TestPlanSweepSingleTerminal(t *testing.T) {
	s := planState(t, 42)
	if err := s.Apply(placedEvent(validSite("dsg-1-0", 1), 1, 1)); err != nil {
		t.Fatal(err)
	}
	d := validDirective("dir-2-0", "dsg-1-0", []int{0}, 2)
	d.ExpiresTick = 2 + ticksPerGameDay // shortest legal TTL
	if err := s.Apply(issuedEvent(d, 2, 2)); err != nil {
		t.Fatal(err)
	}
	s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 10, Y: 10})
	if err := s.Apply(idEvent("designation.fulfilled", "dsg-1-0", 3)); err != nil {
		t.Fatal(err)
	}
	s.Tick = 2 + 2*ticksPerGameDay // well past the directive's TTL

	evs := stepEvents(s, s.m, s.Tick+1)
	f, x := countType(evs, "directive.fulfilled"), countType(evs, "directive.expired")
	if f != 1 || x != 0 {
		t.Fatalf("sweep emitted fulfilled=%d expired=%d, want exactly one terminal (fulfilled)", f, x)
	}
}

// TestPlanSweepExpiry: TTL elapse and the all-targets-dead clause each expire
// an active directive, once.
func TestPlanSweepExpiry(t *testing.T) {
	t.Run("ttl", func(t *testing.T) {
		s := planState(t, 42)
		if err := s.Apply(placedEvent(validZone("dsg-1-0", 1, 12), 1, 1)); err != nil {
			t.Fatal(err)
		}
		d := validDirective("dir-2-0", "dsg-1-0", []int{0}, 2)
		d.ExpiresTick = 2 + ticksPerGameDay
		if err := s.Apply(issuedEvent(d, 2, 2)); err != nil {
			t.Fatal(err)
		}
		s.Tick = d.ExpiresTick - 1
		evs := stepEvents(s, s.m, s.Tick+1) // nextTick == ExpiresTick
		if countType(evs, "directive.expired") != 1 {
			t.Fatalf("sweep at the deadline emitted %d expired, want 1", countType(evs, "directive.expired"))
		}
		for _, e := range evs {
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
		s.Tick++
		if countType(stepEvents(s, s.m, s.Tick+1), "directive.expired") != 0 {
			t.Fatal("expiry re-emitted after the transition")
		}
	})
	t.Run("all targets dead", func(t *testing.T) {
		s := planState(t, 42)
		if err := s.Apply(placedEvent(validZone("dsg-1-0", 1, 12), 1, 1)); err != nil {
			t.Fatal(err)
		}
		if err := s.Apply(issuedEvent(validDirective("dir-2-0", "dsg-1-0", []int{0, 1}, 2), 2, 2)); err != nil {
			t.Fatal(err)
		}
		s.Agents[0].Dead = true
		s.Tick = 5
		if countType(stepEvents(s, s.m, 6), "directive.expired") != 0 {
			t.Fatal("directive expired while a target still lives")
		}
		s.Agents[1].Dead = true
		if countType(stepEvents(s, s.m, 6), "directive.expired") != 1 {
			t.Fatal("all-dead directive did not expire (the un-executable clause)")
		}
	})
}

// TestPlanSweepEndedWorldSilence: an ended world emits nothing — the run-end
// guard freezes the plan sweeps with everything else.
func TestPlanSweepEndedWorldSilence(t *testing.T) {
	s := planState(t, 42)
	if err := s.Apply(placedEvent(validSite("dsg-1-0", 1), 1, 1)); err != nil {
		t.Fatal(err)
	}
	s.Structures = append(s.Structures, Structure{Kind: "shelter", X: 10, Y: 10})
	s.Ended = true
	if evs := stepEvents(s, s.m, s.Tick+1); evs != nil {
		t.Fatalf("ended world emitted %d events", len(evs))
	}
}

// TestPlanRebaseTaxonomy: an ACTIVE directive's ExpiresTick SHIFTs across a
// time snap (remaining lifetime preserved); a consumed directive's deadline is
// a spent artifact (KEEP); issue/placement history never moves; designations
// are all-KEEP (no future deadline exists).
func TestPlanRebaseTaxonomy(t *testing.T) {
	s := planState(t, 42)
	s.Designations = []Designation{{ID: "dsg", Kind: DesignationStructureSite, X: 1, Y: 1, X2: 1, Y2: 1,
		StructureKind: "shelter", PlacedTick: 100, Status: "active", PlacedSeq: 7}}
	s.Directives = []Directive{
		{ID: "a", DesignationID: "dsg", Targets: []int{0}, Text: "t", IssuedTick: 100, ExpiresTick: 100 + ticksPerGameDay, Status: "active"},
		{ID: "b", DesignationID: "dsg", Targets: []int{0}, Text: "t", IssuedTick: 100, ExpiresTick: 100 + ticksPerGameDay, Status: "cancelled"},
	}
	rebaseTicks(s, 5000)
	if got := s.Directives[0].ExpiresTick; got != 100+ticksPerGameDay+5000 {
		t.Errorf("active ExpiresTick = %d, want shifted", got)
	}
	if got := s.Directives[1].ExpiresTick; got != 100+ticksPerGameDay {
		t.Errorf("consumed ExpiresTick = %d, want unshifted (KEEP)", got)
	}
	if s.Directives[0].IssuedTick != 100 || s.Designations[0].PlacedTick != 100 || s.Designations[0].PlacedSeq != 7 {
		t.Error("history/identity fields moved (must KEEP)")
	}
}

// TestPlanLifecycleReplayByteIdentical (SC-004): a from-genesis replay of a
// log carrying the full plan lifecycle — placement of all three kinds, a
// directive issued and fulfilled through the sweeps, a cancellation, and a TTL
// expiry — reconstructs byte-identical state with no guardian running.
func TestPlanLifecycleReplayByteIdentical(t *testing.T) {
	const seed = 84
	const ticks = 90_000 // past the second directive's 1-day TTL
	m := testMap(seed)

	pl := mustPayload
	timeline := map[int64][]store.Event{
		50: {placedEvent(validSite("dsg-50-0", 50), 0, 50)},
		60: {placedEvent(validLine("dsg-60-0", 60), 0, 60)},
		70: {placedEvent(validZone("dsg-70-0", 70, 12), 0, 70)},
		80: {issuedEvent(func() Directive {
			d := validDirective("dir-80-0", "dsg-50-0", []int{0, 1}, 80)
			d.ExpiresTick = 80 + 3*ticksPerGameDay
			return d
		}(), 0, 80)},
		// Genesis charge pays for the grant; the built shelter then fulfills
		// dsg-50-0 at the next boundary and dir-80-0 one tick later.
		90:  {{Tick: 90, Type: "metatron.item_granted", Payload: pl(ItemGrantedPayload{Agent: 0, Kind: "planks", Qty: 4})}},
		100: {{Tick: 100, Type: "agent.built", Payload: pl(BuiltPayload{Agent: Ref(0), Kind: "shelter", X: 10, Y: 10})}},
		200: {idEvent("designation.cancelled", "dsg-60-0", 200)},
		// A second directive against the (unfulfillable, min 12) zone: the
		// sweep expires it at 300 + 1 game day.
		300: {issuedEvent(func() Directive {
			d := validDirective("dir-300-0", "dsg-70-0", []int{2}, 300)
			d.ExpiresTick = 300 + ticksPerGameDay
			return d
		}(), 0, 300)},
	}

	live := NewState(seed, m)
	log := driveTicks(t, live, m, ticks, timeline)

	// Guard: the log must actually carry the full lifecycle.
	wantOnce := []string{"designation.fulfilled", "directive.fulfilled", "designation.cancelled", "directive.expired"}
	for _, typ := range wantOnce {
		if n := countType(log, typ); n != 1 {
			t.Fatalf("log carries %d %s, want exactly 1", n, typ)
		}
	}
	if n := countType(log, "designation.placed"); n != 3 {
		t.Fatalf("log carries %d placements, want 3", n)
	}

	// Replay from genesis: reduce the logged events, align the clock, re-live
	// the quiet tail — the recovery contract (the paused-nudge test's shape).
	replayed := NewState(seed, m)
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replayed.Tick = e.Tick
	}
	driveTicks(t, replayed, m, ticks, nil)

	if live.Hash() != replayed.Hash() {
		t.Fatalf("plan-lifecycle replay diverged from live:\nlive:     %s\nreplayed: %s",
			string(live.Marshal()), string(replayed.Marshal()))
	}
}

// TestPreSpec084SnapshotCompat: a snapshot with neither plan field unmarshals
// to nil slices and re-marshals byte-identically (omitempty — no format bump).
func TestPreSpec084SnapshotCompat(t *testing.T) {
	s := planState(t, 42)
	before := s.Marshal()
	if strings.Contains(string(before), "\"designations\"") || strings.Contains(string(before), "\"directives\"") {
		t.Fatal("empty plan slices leak into canonical bytes (omitempty broken)")
	}
	restored := NewState(42, s.m)
	if err := json.Unmarshal(before, restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.Designations != nil || restored.Directives != nil {
		t.Error("pre-084 snapshot resurrected plan entities")
	}
}
