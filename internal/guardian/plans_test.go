package guardian

// Spec 084 guardian-side suites: the plan tool handlers (T013/T016) — locus
// parsing through target.ParseLocus, the kind×form matrix, deterministic id
// minting, the atomic issue batch with its per-target companion memories (the
// firewall's recorded-data-only channel), door rejections as counsel — plus
// the mirror drift pins (structure kinds, designation kinds).

import (
	"reflect"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/tool"
)

// TestBuildableStructureKindsMirrorSim pins the tool package's hand-carried
// structure_kind Enum equal to sim's recipes-derived list (the
// TestMiracleKindsMirrorTool pattern) — the mirror cannot drift silently.
func TestBuildableStructureKindsMirrorSim(t *testing.T) {
	if got, want := tool.BuildableStructureKinds(), sim.BuildableStructureKinds(); !reflect.DeepEqual(got, want) {
		t.Errorf("tool.BuildableStructureKinds() = %v, want sim's %v", got, want)
	}
}

// TestDesignationKindsMirrorSim pins place_designation's kind Enum equal to
// the sim entity vocabulary (and the handler's form matrix covers exactly it).
func TestDesignationKindsMirrorSim(t *testing.T) {
	want := []string{sim.DesignationSettlementZone, sim.DesignationStructureSite, sim.DesignationWallLine}
	if got := tool.DesignationKinds(); !reflect.DeepEqual(got, want) {
		t.Errorf("tool.DesignationKinds() = %v, want %v", got, want)
	}
	for _, k := range want {
		if _, ok := designationForms[k]; !ok {
			t.Errorf("handler form matrix missing kind %q", k)
		}
	}
}

// planFixture builds the standard guardian fixture with the full grant.
func planFixture(t *testing.T) (*Guardian, *stateInjector, grantSet) {
	t.Helper()
	mt, _, inj, _ := newTestGuardian(t, "so be it")
	return mt, inj, fullGrant()
}

// TestPlaceDesignationLands: each kind lands through the door with its
// normalized locus, the deterministic dsg-<tick>-<seq> id, and the reducer's
// active status; every living villager gains the announcement fact.
func TestPlaceDesignationLands(t *testing.T) {
	mt, inj, grant := planFixture(t)

	site, why := mt.landPlaceDesignation("structure_site", "4,5", "shelter", 0, false, "north shelter", 100, grant)
	if why != "" {
		t.Fatalf("site placement refused: %q", why)
	}
	if site.ID != "dsg-100-0" || site.X != 4 || site.Y != 5 || site.X2 != 4 || site.Y2 != 5 {
		t.Errorf("site = %+v, want dsg-100-0 at (4,5)", site)
	}
	line, why := mt.landPlaceDesignation("wall_line", "2,9 -> 2,2", "", 0, false, "", 100, grant)
	if why != "" {
		t.Fatalf("line placement refused: %q", why)
	}
	// Endpoint (author) order preserved — direction is intent.
	if line.ID != "dsg-100-1" || line.X != 2 || line.Y != 9 || line.X2 != 2 || line.Y2 != 2 {
		t.Errorf("line = %+v, want dsg-100-1 (2,9)->(2,2)", line)
	}
	zone, why := mt.landPlaceDesignation("settlement_zone", "8,8..1,1", "", 0, false, "", 100, grant)
	if why != "" {
		t.Fatalf("zone placement refused: %q", why)
	}
	// Rect normalized min/max; MinStructures carries the tool-door default 3.
	if zone.X != 1 || zone.Y != 1 || zone.X2 != 8 || zone.Y2 != 8 || zone.MinStructures != 3 {
		t.Errorf("zone = %+v, want normalized (1,1)..(8,8) min 3", zone)
	}

	if got := len(inj.state.Designations); got != 3 {
		t.Fatalf("state holds %d designations, want 3", got)
	}
	for _, d := range inj.state.Designations {
		if d.Status != "active" {
			t.Errorf("%s status %q, want active", d.ID, d.Status)
		}
	}
	// The announcement grant: every living villager knows the anchors (the
	// fact shape itself is pinned sim-side; this proves the fan-out through
	// the real door).
	for i := range inj.state.Agents {
		a := &inj.state.Agents[i]
		if a.Dead || a.Map == nil {
			continue
		}
		facts := a.Map.KnownFresh("designation", 0)
		if len(facts) != 3 {
			t.Errorf("agent %d holds %d designation facts, want 3", i, len(facts))
		}
	}
}

// TestPlaceDesignationRefusals: the kind×form matrix, locus parse misses,
// partial args, and door rejections all come back as repairable counsel.
func TestPlaceDesignationRefusals(t *testing.T) {
	mt, inj, grant := planFixture(t)
	cases := []struct {
		name             string
		kind, target, sk string
		min              int
		hasMin           bool
		wantWhy          string
	}{
		{"unknown kind", "camp", "4,5", "", 0, false, "no designation of kind"},
		{"site wants a point", "structure_site", "1,1..2,2", "shelter", 0, false, `one tile, like "4,5"`},
		{"line wants a line", "wall_line", "4,5", "", 0, false, "axis-aligned line"},
		{"zone wants a rect", "settlement_zone", "2,2->2,9", "", 0, false, "rectangle"},
		{"bad locus", "structure_site", "over yonder", "shelter", 0, false, "could not read that site"},
		{"diagonal line", "wall_line", "1,1->3,3", "", 0, false, "could not read that site"},
		{"site missing structure_kind", "structure_site", "4,5", "", 0, false, "needs a structure_kind"},
		{"out of bounds", "structure_site", "999,999", "shelter", 0, false, "beyond the world"},
		{"unbuildable kind", "structure_site", "4,5", "palace", 0, false, "the world would not let me"},
		{"zone too big", "settlement_zone", "0,0..20,20", "", 0, false, "too large"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := len(inj.batches)
			d, why := mt.landPlaceDesignation(tc.kind, tc.target, tc.sk, tc.min, tc.hasMin, "", 200, grant)
			if d != nil || !strings.Contains(why, tc.wantWhy) {
				t.Errorf("got (%v, %q), want refusal containing %q", d, why, tc.wantWhy)
			}
			// Handler-side refusals inject nothing; door-side ones dry-run
			// reject, so nothing ever lands either way.
			if len(inj.state.Designations) != 0 {
				t.Error("a refused placement landed")
			}
			_ = before
		})
	}

	// Ungranted: structural refusal.
	if _, why := mt.landPlaceDesignation("structure_site", "4,5", "shelter", 0, false, "", 200, grantSet{}); !strings.Contains(why, "not granted") {
		t.Errorf("ungranted why = %q", why)
	}
}

// TestCancelDesignationRaces: cancel lands once; the second cancel and an
// unknown id come back as counsel (the one-way door surfaces in-fiction).
func TestCancelDesignationRaces(t *testing.T) {
	mt, _, grant := planFixture(t)
	d, why := mt.landPlaceDesignation("structure_site", "4,5", "shelter", 0, false, "", 100, grant)
	if why != "" {
		t.Fatal(why)
	}
	if why := mt.landCancelDesignation(d.ID, grant); why != "" {
		t.Fatalf("first cancel refused: %q", why)
	}
	if why := mt.landCancelDesignation(d.ID, grant); !strings.Contains(why, "already run its course") {
		t.Errorf("second cancel why = %q", why)
	}
	if why := mt.landCancelDesignation("dsg-none", grant); !strings.Contains(why, "no designation called") {
		t.Errorf("unknown id why = %q", why)
	}
}

// TestIssueDirectiveAtomicBatch (US3 AS1, FR-009): one batch carries
// directive.issued plus exactly one companion agent.memory_added per target —
// the vision-memory shape, the firewall's only channel: the guardian's words
// reach villagers as recorded event data and nothing else.
func TestIssueDirectiveAtomicBatch(t *testing.T) {
	mt, inj, grant := planFixture(t)
	if _, why := mt.landPlaceDesignation("structure_site", "4,5", "shelter", 0, false, "", 100, grant); why != "" {
		t.Fatal(why)
	}
	alive := map[int]bool{}
	for i := range inj.state.Agents {
		alive[i] = !inj.state.Agents[i].Dead
	}
	// Name targets in REVERSE index order — the payload must land ascending.
	targets := sim.AgentNames[1] + ", " + sim.AgentNames[0]
	dir, why := mt.landIssueDirective("dsg-100-0", targets, "Raise the shelter I have marked.", 0, 200, alive, grant)
	if why != "" {
		t.Fatalf("issue refused: %q", why)
	}
	if dir.ID != "dir-200-0" || !reflect.DeepEqual(dir.Targets, []int{0, 1}) || dir.Village {
		t.Errorf("directive = %+v, want dir-200-0 targets [0 1] village=false", dir)
	}
	if dir.ExpiresTick != 200+3*24*3600 {
		t.Errorf("ExpiresTick = %d, want the 3-day default", dir.ExpiresTick)
	}

	// The last injected batch is the atomic issue: issued + 2 companions.
	batch := inj.batches[len(inj.batches)-1]
	if len(batch) != 3 || batch[0].Type != "directive.issued" ||
		batch[1].Type != "agent.memory_added" || batch[2].Type != "agent.memory_added" {
		t.Fatalf("issue batch = %v, want issued + one memory per target", batch)
	}
	for _, e := range batch[1:] {
		if !strings.Contains(string(e.Payload), "The Guardian charges you: Raise the shelter I have marked.") {
			t.Errorf("companion memory lacks the framing text: %s", e.Payload)
		}
	}
	if got := inj.state.Directives[0].Status; got != "active" {
		t.Errorf("landed status %q", got)
	}

	// "everyone" resolves to all living, ascending, and marks Village.
	all, why := mt.landIssueDirective("dsg-100-0", "everyone", "All of you: build.", 0, 201, alive, grant)
	if why != "" {
		t.Fatalf("everyone refused: %q", why)
	}
	if !all.Village || len(all.Targets) != len(inj.state.Agents) {
		t.Errorf("everyone = %+v", all)
	}
}

// TestIssueDirectiveRejections (US3 AS2): dead target, unknown name, unknown/
// non-active designation, bad TTL, empty text — the whole batch rejects
// (nothing lands) with repairable counsel.
func TestIssueDirectiveRejections(t *testing.T) {
	mt, inj, grant := planFixture(t)
	if _, why := mt.landPlaceDesignation("structure_site", "4,5", "shelter", 0, false, "", 100, grant); why != "" {
		t.Fatal(why)
	}
	alive := map[int]bool{}
	for i := range inj.state.Agents {
		alive[i] = true
	}
	alive[2] = false // a dead villager

	cases := []struct {
		name         string
		dsg, targets string
		text         string
		ttl          int
		wantWhy      string
	}{
		{"dead target", "dsg-100-0", sim.AgentNames[0] + ", " + sim.AgentNames[2], "go", 0, "beyond reach"},
		{"unknown name", "dsg-100-0", "Nobody", "go", 0, "no villager named"},
		{"unknown designation", "dsg-none", sim.AgentNames[0], "go", 0, "no designation by that name"},
		{"ttl too long", "dsg-100-0", sim.AgentNames[0], "go", 9, "1 to 7 days"},
		{"empty text", "dsg-100-0", sim.AgentNames[0], "   ", 0, "give me the charge in words"},
		{"empty designation", "", sim.AgentNames[0], "go", 0, "name the designation"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, why := mt.landIssueDirective(tc.dsg, tc.targets, tc.text, tc.ttl, 300, alive, grant)
			if d != nil || !strings.Contains(why, tc.wantWhy) {
				t.Errorf("got (%v, %q), want refusal containing %q", d, why, tc.wantWhy)
			}
			if len(inj.state.Directives) != 0 {
				t.Error("a refused issue landed (atomicity broken)")
			}
		})
	}

	// Non-active designation: cancel it, then issue against it.
	if why := mt.landCancelDesignation("dsg-100-0", grant); why != "" {
		t.Fatal(why)
	}
	if d, why := mt.landIssueDirective("dsg-100-0", sim.AgentNames[0], "go", 0, 301, alive, grant); d != nil ||
		!strings.Contains(why, "run its course") {
		t.Errorf("non-active designation issue = (%v, %q)", d, why)
	}
}

// TestCancelDirective: the cancel path and its counsel.
func TestCancelDirective(t *testing.T) {
	mt, inj, grant := planFixture(t)
	if _, why := mt.landPlaceDesignation("structure_site", "4,5", "shelter", 0, false, "", 100, grant); why != "" {
		t.Fatal(why)
	}
	alive := map[int]bool{}
	for i := range inj.state.Agents {
		alive[i] = true
	}
	d, why := mt.landIssueDirective("dsg-100-0", sim.AgentNames[0], "go", 0, 200, alive, grant)
	if why != "" {
		t.Fatal(why)
	}
	if why := mt.landCancelDirective(d.ID, grant); why != "" {
		t.Fatalf("cancel refused: %q", why)
	}
	if why := mt.landCancelDirective(d.ID, grant); !strings.Contains(why, "already run its course") {
		t.Errorf("second cancel why = %q", why)
	}
	if why := mt.landCancelDirective("dir-none", grant); !strings.Contains(why, "no directive called") {
		t.Errorf("unknown id why = %q", why)
	}
}

// TestPlanIDMintingDeterministic: same-tick mints disambiguate by seq even
// before the mirror reflects the first landing (the nextOrderID guarantee),
// and the two prefixes count independently.
func TestPlanIDMintingDeterministic(t *testing.T) {
	mt, _, _ := planFixture(t)
	if id := mt.nextPlanID("dsg", 500); id != "dsg-500-0" {
		t.Errorf("first = %s", id)
	}
	if id := mt.nextPlanID("dsg", 500); id != "dsg-500-1" {
		t.Errorf("second = %s", id)
	}
	if id := mt.nextPlanID("dir", 500); id != "dir-500-0" {
		t.Errorf("dir counts with dsg: %s", id)
	}
	if id := mt.nextPlanID("dsg", 501); id != "dsg-501-0" {
		t.Errorf("new tick = %s", id)
	}
}

// TestPlanPromptSections (FR-015): active designations and directives render
// in the turn user prompt with id/kind/site/days-left; a plan-free prompt is
// byte-identical to a pre-084 one (empty sections render nothing).
func TestPlanPromptSections(t *testing.T) {
	alive := map[int]bool{0: true}
	base := turnUserPrompt(100, 1, alive, nil, nil, nil, nil, nil, "", "", "", "The player says:\nhello")
	if strings.Contains(base, "Designations") || strings.Contains(base, "Directives") {
		t.Error("plan-free prompt carries plan sections")
	}

	designations := []sim.Designation{
		{ID: "dsg-1-0", Kind: sim.DesignationStructureSite, X: 4, Y: 5, X2: 4, Y2: 5,
			StructureKind: "shelter", Label: "north shelter", Status: "active"},
		{ID: "dsg-1-1", Kind: sim.DesignationSettlementZone, X: 1, Y: 1, X2: 8, Y2: 8,
			MinStructures: 3, Status: "cancelled"}, // consumed — must not render
	}
	directives := []sim.Directive{
		{ID: "dir-2-0", DesignationID: "dsg-1-0", Targets: []int{0}, Text: "Raise it.",
			IssuedTick: 100, ExpiresTick: 100 + 2*24*3600, Status: "active"},
	}
	got := turnUserPrompt(100, 1, alive, nil, designations, directives, nil, nil, "", "", "", "The player says:\nhello")
	for _, want := range []string{
		"Designations you have marked on the world:",
		`- dsg-1-0: structure_site at (4,5) (shelter) — "north shelter"`,
		"Directives you have laid on the village:",
		"bound to dsg-1-0",
		"(2 day(s) left)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("prompt missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "dsg-1-1") {
		t.Error("consumed designation rendered")
	}
}
