package guardian

// Spec 101 guardian-side suites: landCanonizeRegion (id minting, the door,
// refusal-to-counsel mapping, the charge pre-check, gratis structural
// absence) and the myth-briefing companion — plus the feature-kind mirror
// drift pin (the DesignationKindsMirrorSim/BuildableStructureKindsMirrorSim
// pattern, plans_test.go).

import (
	"reflect"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/tool"
)

// TestCanonizeFeatureKindsMirrorSim pins the tool package's hand-carried
// feature-kind Enum equal to sim's canonical vocabulary — the mirror cannot
// drift silently.
func TestCanonizeFeatureKindsMirrorSim(t *testing.T) {
	if got, want := tool.CanonizeFeatureKinds(), sim.CanonizeFeatureKinds(); !reflect.DeepEqual(got, want) {
		t.Errorf("tool.CanonizeFeatureKinds() = %v, want sim's %v", got, want)
	}
}

// TestLandCanonizeRegionLands: a name-only canonization lands through the
// door with the deterministic reg-<tick>-<seq> id and the reducer's active
// status; a second call (with a feature) mints the next same-tick seq.
func TestLandCanonizeRegionLands(t *testing.T) {
	mt, inj, grant := planFixture(t)
	inj.state.GuardianCharges = 10

	r, summary := mt.landCanonizeRegion(10, 10, 5, "Thornspire", "", 0, 0, false, 100, inj.state.GuardianCharges, grant)
	if summary == "" || r == nil {
		t.Fatalf("canonization refused: %q", summary)
	}
	if r.ID != "reg-100-0" {
		t.Errorf("id = %q, want reg-100-0", r.ID)
	}
	if got := len(inj.state.Regions); got != 1 {
		t.Fatalf("state holds %d regions, want 1", got)
	}
	if inj.state.Regions[0].Status != "active" {
		t.Errorf("status = %q, want active", inj.state.Regions[0].Status)
	}

	r2, why := mt.landCanonizeRegion(30, 30, 4, "Second Place", "path", 30, 30, true, 100, inj.state.GuardianCharges, grant)
	if why == "" || r2 == nil {
		t.Fatalf("second canonization refused: %q", why)
	}
	if r2.ID != "reg-100-1" {
		t.Errorf("second id = %q, want reg-100-1 (same-tick seq bump)", r2.ID)
	}
	if got := len(inj.state.Structures); got != 1 {
		t.Fatalf("structures = %d, want 1 (the feature)", got)
	}
}

// TestLandCanonizeRegionUngranted: structural absence — an ungranted world
// refuses at the door before any charge/state check.
func TestLandCanonizeRegionUngranted(t *testing.T) {
	mt, inj, _ := planFixture(t)
	inj.state.GuardianCharges = 10
	empty := grantSet{tools: map[string]bool{}}
	r, why := mt.landCanonizeRegion(10, 10, 5, "Thornspire", "", 0, 0, false, 100, inj.state.GuardianCharges, empty)
	if r != nil || why == "" {
		t.Fatalf("ungranted canonize_region should refuse, got (%v, %q)", r, why)
	}
	if len(inj.state.Regions) != 0 {
		t.Error("an ungranted refusal must not land anything")
	}
}

// TestLandCanonizeRegionChargeGate: the turn-level pre-check (landMiracle's
// "no charges banked" shape) refuses BEFORE reaching the door when the
// mirrored charge count is under the premium.
func TestLandCanonizeRegionChargeGate(t *testing.T) {
	mt, inj, grant := planFixture(t)
	r, why := mt.landCanonizeRegion(10, 10, 5, "Thornspire", "", 0, 0, false, 100, sim.GuardianRegionCharge-1, grant)
	if r != nil || why == "" {
		t.Fatalf("insufficient charges should refuse, got (%v, %q)", r, why)
	}
	if len(inj.state.Regions) != 0 {
		t.Error("a charge-gated refusal must not touch InjectSocial at all")
	}
}

// TestLandCanonizeRegionNeedsName: an empty name refuses with in-fiction
// counsel before reaching the door (the landIssueDirective empty-text shape).
func TestLandCanonizeRegionNeedsName(t *testing.T) {
	mt, inj, grant := planFixture(t)
	inj.state.GuardianCharges = 10
	r, why := mt.landCanonizeRegion(10, 10, 5, "   ", "", 0, 0, false, 100, inj.state.GuardianCharges, grant)
	if r != nil || why == "" {
		t.Fatal("a blank name should refuse")
	}
}

// TestLandCanonizeRegionFeatureNeedsSite: a feature_kind with no feature
// site refuses before reaching the door.
func TestLandCanonizeRegionFeatureNeedsSite(t *testing.T) {
	mt, inj, grant := planFixture(t)
	inj.state.GuardianCharges = 10
	r, why := mt.landCanonizeRegion(10, 10, 5, "Thornspire", "path", 0, 0, false, 100, inj.state.GuardianCharges, grant)
	if r != nil || why == "" {
		t.Fatal("a feature_kind with hasFeature=false should refuse")
	}
}

// TestCanonizeRegionDoorRefusalsMapToCounsel: door rejections (overlap, cap)
// translate to in-fiction counsel via canonizeRefusal, mirroring
// TestPlaceDesignationRejections' shape (plans_test.go).
func TestCanonizeRegionDoorRefusalsMapToCounsel(t *testing.T) {
	mt, inj, grant := planFixture(t)
	inj.state.GuardianCharges = 1000

	if _, why := mt.landCanonizeRegion(20, 20, 5, "Thornspire", "", 0, 0, false, 100, inj.state.GuardianCharges, grant); why == "" {
		t.Fatal("first canonization should land")
	}
	_, why := mt.landCanonizeRegion(22, 20, 5, "Overlap", "", 0, 0, false, 200, inj.state.GuardianCharges, grant)
	if why == "" {
		t.Fatal("overlapping canonization should refuse")
	}
	if !strings.Contains(why, "already named") {
		t.Errorf("overlap refusal = %q, want the already-named counsel", why)
	}
}

// TestHandleBriefMythsEmptyCorpus: the read-only tool never errors on an
// empty belief corpus — an honest "no candidates" notice instead.
func TestHandleBriefMythsEmptyCorpus(t *testing.T) {
	mt, _, _, _ := newTestGuardian(t, "ok")
	got := renderMythBriefing(mt.myths)
	if !strings.Contains(got, "no candidate") {
		t.Errorf("empty myth briefing = %q, want an honest empty notice", got)
	}
}

// TestRenderMythBriefing: the fact sheet lists each candidate's coordinate,
// wording, holder count, and confidence.
func TestRenderMythBriefing(t *testing.T) {
	got := renderMythBriefing([]sim.PlaceMythBriefing{
		{X: 10, Y: 10, Statement: "Thornspire at the forest's edge.", Holders: 3, Confidence: 70},
	})
	for _, want := range []string{"(10,10)", "Thornspire at the forest's edge.", "3 holder", "70%"} {
		if !strings.Contains(got, want) {
			t.Errorf("myth briefing = %q, missing %q", got, want)
		}
	}
}
