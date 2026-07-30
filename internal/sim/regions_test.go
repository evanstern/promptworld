package sim

// Spec 101 (TASK-81, the canonization miracle) test suites: door validation/
// refusal (FR-006), the artifact discipline (deterministic id, PlacedSeq
// stamping, the active cap), the D1 place-text integration (describePlace
// surfaces a region's coined name), the D3 observation-channel wiring (a
// canonized feature is exhaustively observable through the UNMODIFIED spec-097
// channel — reconciliation itself is covered by internal/mind's own suites),
// and from-genesis replay byte-identity over a canonization lifecycle log.

import (
	"fmt"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

func regionState(t *testing.T, seed uint64) *State {
	t.Helper()
	m := testMap(seed)
	return NewState(seed, m)
}

func regionNamedEvent(p RegionNamedPayload, seq, tick int64) store.Event {
	return store.Event{Seq: seq, Tick: tick, Type: "guardian.region_named", Payload: mustPayload(p)}
}

// TestRegionNamedLandsActive (spec 084 shape parity): the payload's absence
// of Status/PlacedSeq fields is moot (they don't ride the wire at all) — the
// reducer stamps Status "active" and PlacedSeq from the event's own store
// seq, and spends the D4 premium (2 charges) — genesis banks only
// GuardianGenesisCharges (1), so this test funds the bank itself rather than
// relying on day-1 grace to cover a 2-charge act.
func TestRegionNamedLandsActive(t *testing.T) {
	s := regionState(t, 42)
	s.GuardianCharges = 10
	before := s.GuardianCharges
	p := RegionNamedPayload{ID: "reg-100-0", X: 10, Y: 10, Radius: 5, Name: "Thornspire"}
	if err := s.Apply(regionNamedEvent(p, 77, 100)); err != nil {
		t.Fatalf("region_named: %v", err)
	}
	if len(s.Regions) != 1 {
		t.Fatalf("regions = %d, want 1", len(s.Regions))
	}
	got := s.Regions[0]
	if got.Status != "active" || got.PlacedSeq != 77 {
		t.Errorf("landed status=%q placedSeq=%d, want active/77", got.Status, got.PlacedSeq)
	}
	if got.Name != "Thornspire" || got.X != 10 || got.Y != 10 || got.Radius != 5 {
		t.Errorf("landed region = %+v, want the payload's fields verbatim", got)
	}
	if want := before - GuardianRegionCharge; s.GuardianCharges != want {
		t.Errorf("charges = %d, want %d (D4: %d flat)", s.GuardianCharges, want, GuardianRegionCharge)
	}
}

// TestRegionNamedGratisWaivesCharge: gratis waives ONLY the charge — every
// other validation still runs (the spendMiracleCharge contract, miracles.go).
func TestRegionNamedGratisWaivesCharge(t *testing.T) {
	s := regionState(t, 42)
	s.GuardianCharges = 0
	p := RegionNamedPayload{ID: "reg-100-0", X: 10, Y: 10, Radius: 5, Name: "Thornspire", Gratis: true}
	if err := s.Apply(regionNamedEvent(p, 0, 100)); err != nil {
		t.Fatalf("gratis region_named should land with zero charges banked: %v", err)
	}
	if s.GuardianCharges != 0 {
		t.Errorf("charges = %d, want 0 (gratis waives the spend, not a refund)", s.GuardianCharges)
	}
}

// TestRegionNamedValidationRefusals is the door table (validate-not-clamp):
// every listed payload must be rejected BEFORE any charge spends or state
// mutates — a rejected canonization leaves s.Regions/s.Structures/
// s.GuardianCharges completely untouched (reject-whole).
func TestRegionNamedValidationRefusals(t *testing.T) {
	cases := []struct {
		name string
		p    RegionNamedPayload
	}{
		{"empty id", RegionNamedPayload{X: 10, Y: 10, Radius: 5, Name: "Thornspire"}},
		{"radius too small", RegionNamedPayload{ID: "reg-1", X: 10, Y: 10, Radius: 1, Name: "Thornspire"}},
		{"radius too large", RegionNamedPayload{ID: "reg-1", X: 10, Y: 10, Radius: 25, Name: "Thornspire"}},
		{"empty name", RegionNamedPayload{ID: "reg-1", X: 10, Y: 10, Radius: 5, Name: "   "}},
		{"name too long", RegionNamedPayload{ID: "reg-1", X: 10, Y: 10, Radius: 5, Name: stringOfLen(81)}},
		{"center out of bounds", RegionNamedPayload{ID: "reg-1", X: 999, Y: 999, Radius: 5, Name: "Thornspire"}},
		{"unknown feature kind", RegionNamedPayload{ID: "reg-1", X: 10, Y: 10, Radius: 5, Name: "Thornspire",
			FeatureKind: "chest", FeatureX: 10, FeatureY: 10}},
		{"feature outside region", RegionNamedPayload{ID: "reg-1", X: 10, Y: 10, Radius: 5, Name: "Thornspire",
			FeatureKind: "path", FeatureX: 30, FeatureY: 30}},
		{"feature site out of bounds", RegionNamedPayload{ID: "reg-1", X: 10, Y: 10, Radius: 20, Name: "Thornspire",
			FeatureKind: "path", FeatureX: 999, FeatureY: 999}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := regionState(t, 42)
			before := s.GuardianCharges
			if err := s.Apply(regionNamedEvent(c.p, 0, 100)); err == nil {
				t.Fatalf("%s: expected a door refusal, landed clean", c.name)
			}
			if len(s.Regions) != 0 {
				t.Errorf("%s: a rejected canonization must not append a region", c.name)
			}
			if len(s.Structures) != 0 {
				t.Errorf("%s: a rejected canonization must not place a feature", c.name)
			}
			if s.GuardianCharges != before {
				t.Errorf("%s: a rejected canonization must not spend a charge (got %d, want %d)", c.name, s.GuardianCharges, before)
			}
		})
	}
}

// TestRegionNamedDuplicateIDRefused: ids are assigned once (the
// designation/directive discipline) — a duplicate is rejected regardless of
// the first region's status (there is no status a region can hold besides
// active in v1, but the door still checks by id, not by status).
func TestRegionNamedDuplicateIDRefused(t *testing.T) {
	s := regionState(t, 42)
	s.GuardianCharges = 10
	first := RegionNamedPayload{ID: "reg-100-0", X: 10, Y: 10, Radius: 3, Name: "Thornspire"}
	if err := s.Apply(regionNamedEvent(first, 0, 100)); err != nil {
		t.Fatalf("first: %v", err)
	}
	dup := RegionNamedPayload{ID: "reg-100-0", X: 40, Y: 40, Radius: 3, Name: "Somewhere Else"}
	if err := s.Apply(regionNamedEvent(dup, 0, 200)); err == nil {
		t.Fatal("duplicate region id should be refused")
	}
	if len(s.Regions) != 1 {
		t.Fatalf("regions = %d, want 1 (the duplicate must not append)", len(s.Regions))
	}
}

// TestRegionOverlapRefused is the spec's own named edge case: "second
// christening of an overlapping region refuses at the door (one name per
// ground truth; renames are future work)". Touching circles (distance ==
// sum of radii) are NOT overlap, so an edge-adjacent region still lands.
func TestRegionOverlapRefused(t *testing.T) {
	s := regionState(t, 42)
	s.GuardianCharges = 10
	first := RegionNamedPayload{ID: "reg-100-0", X: 20, Y: 20, Radius: 5, Name: "Thornspire"}
	if err := s.Apply(regionNamedEvent(first, 0, 100)); err != nil {
		t.Fatalf("first: %v", err)
	}
	overlapping := RegionNamedPayload{ID: "reg-200-0", X: 22, Y: 20, Radius: 5, Name: "Second Name"}
	if err := s.Apply(regionNamedEvent(overlapping, 0, 200)); err == nil {
		t.Fatal("an overlapping region should be refused")
	}
	if len(s.Regions) != 1 {
		t.Fatalf("regions = %d, want 1 (the overlap must not land)", len(s.Regions))
	}
	// Edge-adjacent (distance exactly radius sum, 10): legal.
	adjacent := RegionNamedPayload{ID: "reg-300-0", X: 30, Y: 20, Radius: 5, Name: "Neighbor"}
	if err := s.Apply(regionNamedEvent(adjacent, 0, 300)); err != nil {
		t.Errorf("edge-adjacent (touching, not overlapping) region should land: %v", err)
	}
}

// TestRegionActiveCap: the door refuses the (cap+1)th region (GuardianRegionCap).
func TestRegionActiveCap(t *testing.T) {
	s := regionState(t, 42)
	s.GuardianCharges = 1000
	// GuardianRegionCap (16) non-overlapping radius-2 regions: two rows of 8,
	// x spaced 5 apart (> the 4-tile overlap-refusal threshold), rows 6 apart
	// vertically (also > threshold) — all comfortably within the 64x64 map.
	for i := 0; i < GuardianRegionCap; i++ {
		x := 2 + (i%8)*5
		y := 2 + (i/8)*6
		p := RegionNamedPayload{ID: idStr("reg", i), X: x, Y: y, Radius: 2, Name: idStr("Place", i)}
		if err := s.Apply(regionNamedEvent(p, 0, int64(100+i))); err != nil {
			t.Fatalf("region %d at (%d,%d): %v", i, x, y, err)
		}
	}
	over := RegionNamedPayload{ID: "reg-over", X: 2, Y: 60, Radius: 2, Name: "OneTooMany"}
	if err := s.Apply(regionNamedEvent(over, 0, 500)); err == nil {
		t.Fatal("the cap+1th region should be refused")
	}
	if len(s.Regions) != GuardianRegionCap {
		t.Fatalf("regions = %d, want %d (cap)", len(s.Regions), GuardianRegionCap)
	}
}

// TestRegionFeaturePlacement: an optional feature lands as a fresh Structure
// at the given site (mirroring agent.built's construction, minus the
// fuel/owner entanglements canonizeFeatureKinds excludes) — a wall gets full
// health, everything else no special fields.
func TestRegionFeaturePlacement(t *testing.T) {
	s := regionState(t, 42)
	s.GuardianCharges = 10
	p := RegionNamedPayload{ID: "reg-100-0", X: 10, Y: 10, Radius: 5, Name: "Thornspire",
		FeatureKind: "wall_stone", FeatureX: 10, FeatureY: 10}
	if err := s.Apply(regionNamedEvent(p, 0, 100)); err != nil {
		t.Fatalf("region_named with feature: %v", err)
	}
	if len(s.Structures) != 1 {
		t.Fatalf("structures = %d, want 1", len(s.Structures))
	}
	st := s.Structures[0]
	if st.Kind != "wall_stone" || st.X != 10 || st.Y != 10 {
		t.Errorf("feature = %+v, want wall_stone at (10,10)", st)
	}
	if st.HP != wallMaxHP("wall_stone") {
		t.Errorf("feature HP = %d, want %d (a fresh wall stands at full health)", st.HP, wallMaxHP("wall_stone"))
	}
}

// TestRegionFeatureRequiresBuildSite: a feature site already holding a
// structure is rejected (buildSite reuse — the existing entity/build
// placement rule, never re-derived).
func TestRegionFeatureRequiresBuildSite(t *testing.T) {
	s := regionState(t, 42)
	s.Structures = append(s.Structures, Structure{Kind: "fire", X: 10, Y: 10})
	p := RegionNamedPayload{ID: "reg-100-0", X: 10, Y: 10, Radius: 5, Name: "Thornspire",
		FeatureKind: "path", FeatureX: 10, FeatureY: 10}
	if err := s.Apply(regionNamedEvent(p, 0, 100)); err == nil {
		t.Fatal("a feature site already occupied should be refused")
	}
	if len(s.Regions) != 0 {
		t.Error("the region must not land either — reject-whole")
	}
}

// TestDescribePlaceUsesRegionName is the D1 place-text integration: a tile
// inside a canonized region situates by its coined name, taking priority
// over the generic structure/terrain phrase underneath it — and a tile
// outside every region is unaffected (the pre-101 behavior, byte-identical).
func TestDescribePlaceUsesRegionName(t *testing.T) {
	s := regionState(t, 42)
	s.GuardianCharges = 10
	// A far tile's description BEFORE the canonization — captured so the
	// "unaffected" assertion below doesn't need to guess what's notable way
	// out at (50,50) on this map/seed, only that canonizing elsewhere never
	// changes it.
	farBefore := describePlace(s, 50, 50)
	// A structure at the SAME tile as the region center would otherwise win
	// featureDesc's structure-first scan — proving the region check runs
	// ahead of it.
	s.Structures = append(s.Structures, Structure{Kind: "fire", X: 10, Y: 10})
	p := RegionNamedPayload{ID: "reg-100-0", X: 10, Y: 10, Radius: 5, Name: "Thornspire"}
	if err := s.Apply(regionNamedEvent(p, 0, 100)); err != nil {
		t.Fatalf("region_named: %v", err)
	}
	if got := describePlace(s, 10, 10); got != "Thornspire" {
		t.Errorf("describePlace at the region center = %q, want %q (region name over the fire beneath it)", got, "Thornspire")
	}
	// A tile within the radius but not the exact center still resolves to
	// the region name (a circle, not a point).
	if got := describePlace(s, 12, 10); got != "Thornspire" {
		t.Errorf("describePlace within the region = %q, want %q", got, "Thornspire")
	}
	// Far outside the region: unaffected, pre-101 behavior.
	if got := describePlace(s, 50, 50); got != farBefore {
		t.Errorf("describePlace far outside every region = %q, want %q (unchanged by a distant canonization)", got, farBefore)
	}
}

// TestCanonizedFeatureObservable is the D3 wiring proof: a feature the
// canonize working raises is exhaustively observable through the UNMODIFIED
// spec-097 channel — no new perception machinery (observedKinds already
// scans s.Structures; a canonized wall participates identically to any
// villager-built one). Belief-vs-observation matching itself (confirm/
// disconfirm) is internal/mind/reconcile.go's own suite — this test proves
// only that the ground truth a matching belief would need is actually
// present in the emitted Kinds set, the seam D3 rests on.
func TestCanonizedFeatureObservable(t *testing.T) {
	s := regionState(t, 42)
	s.GuardianCharges = 10
	p := RegionNamedPayload{ID: "reg-100-0", X: 10, Y: 10, Radius: 5, Name: "Thornspire",
		FeatureKind: "wall_stone", FeatureX: 10, FeatureY: 10}
	if err := s.Apply(regionNamedEvent(p, 0, 100)); err != nil {
		t.Fatalf("region_named with feature: %v", err)
	}
	kinds := observedKinds(s, s.m, 10, 10)
	found := false
	for _, k := range kinds {
		if k == "wall_stone" {
			found = true
		}
	}
	if !found {
		t.Errorf("observedKinds at the canonized feature's site = %v, want it to include %q", kinds, "wall_stone")
	}
}

// TestRegionLifecycleReplayByteIdentical (FR-006): a from-genesis replay of a
// log carrying a region christening (with a feature) reconstructs
// byte-identical state with no guardian running — the additive-event
// contract (spec 094 doctrine, no format bump).
func TestRegionLifecycleReplayByteIdentical(t *testing.T) {
	const seed = 42
	const ticks = 5000
	m := testMap(seed)

	// Gratis (the operator-door escape hatch, D4): sidesteps the 2-charge
	// requirement — genesis banks only 1, and this test is proving replay
	// byte-identity of the REGION mechanism, not the charge economy (already
	// covered by TestRegionNamedLandsActive/TestRegionNamedGratisWaivesCharge).
	timeline := map[int64][]store.Event{
		100: {regionNamedEvent(RegionNamedPayload{
			ID: "reg-100-0", X: 10, Y: 10, Radius: 5, Name: "Thornspire",
			FeatureKind: "wall_stone", FeatureX: 10, FeatureY: 10, Gratis: true,
		}, 0, 100)},
	}

	live := NewState(seed, m)
	log := driveTicks(t, live, m, ticks, timeline)

	if n := countType(log, "guardian.region_named"); n != 1 {
		t.Fatalf("log carries %d region_named, want exactly 1", n)
	}

	replayed := NewState(seed, m)
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replayed.Tick = e.Tick
	}
	driveTicks(t, replayed, m, ticks, nil)

	if string(live.Marshal()) != string(replayed.Marshal()) {
		t.Fatalf("replay diverged:\nlive:     %s\nreplayed: %s", string(live.Marshal()), string(replayed.Marshal()))
	}
}

func stringOfLen(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'x'
	}
	return string(b)
}

func idStr(prefix string, i int) string {
	return fmt.Sprintf("%s-%d", prefix, i)
}
