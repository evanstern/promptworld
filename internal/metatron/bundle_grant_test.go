package metatron

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/bundle"
)

// Unit tests for the persona grant-narrowing composition logic (spec 036 US4
// T030, data-model.md GrantNarrowing): a bundle's capabilities.json intersects
// with the world-level grant — it may exclude tools/miracle_kinds the world
// already grants, but it can never widen past what the world excludes. These
// tests exercise intersectGrant/narrowGrantForBundles directly, in isolation
// from disk and from the turn assembly; internal/bundle/load_test.go's
// TestDiscoverPersona and internal/metatron/bundle_integration_test.go cover
// the same semantics end-to-end through a real fixture (T031/T032).

// TestIntersectGrantNarrowsKinds: an unrestricted world (fullGrant, "all
// kinds") intersected with a persona's miracle_kinds list becomes exactly that
// list — the persona's list is the new ceiling since the world had none.
func TestIntersectGrantNarrowsKinds(t *testing.T) {
	g := fullGrant()
	got := intersectGrant(g, &bundle.GrantDoc{MiracleKinds: []string{"remove", "give_item"}})

	if !got.kindsRestricted {
		t.Fatal("kindsRestricted = false, want true after narrowing")
	}
	if got.allowsKind("move") {
		t.Error("move should be excluded — the persona's list did not name it")
	}
	if !got.allowsKind("remove") || !got.allowsKind("give_item") {
		t.Error("remove/give_item should remain granted — named by the persona")
	}
}

// TestIntersectGrantNeverWidensKinds: a persona naming a kind the world
// ALREADY excludes cannot resurrect it — intersection, never union.
func TestIntersectGrantNeverWidensKinds(t *testing.T) {
	g := fullGrant()
	g.kindsRestricted = true
	g.kinds = map[string]bool{"move": true} // world grants ONLY move

	got := intersectGrant(g, &bundle.GrantDoc{MiracleKinds: []string{"move", "remove"}})

	if !got.allowsKind("move") {
		t.Error("move should remain granted — named by both world and persona")
	}
	if got.allowsKind("remove") {
		t.Error("remove should stay excluded — the world never granted it; the persona cannot widen to it")
	}
}

// TestIntersectGrantNeverWidensTools: symmetric case for the "tools" key
// (spanning built-ins AND bundle tool names, same flat namespace loadManifest
// already uses). A persona naming a tool the world's own explicit list
// excludes — even the persona's OWN bundle tool — must not grant it; the
// world owner's explicit allowlist is authoritative.
func TestIntersectGrantNeverWidensTools(t *testing.T) {
	g := fullGrant()
	g.tools = map[string]bool{"send_vision": true} // world excludes work_miracle
	g.toolsConstrained = true
	g.bundleAllowed = map[string]bool{"send_vision": true} // world's list never named gandalf_bless

	got := intersectGrant(g, &bundle.GrantDoc{Tools: []string{"send_vision", "work_miracle", "gandalf_bless"}})

	if !got.allows("send_vision") {
		t.Error("send_vision should remain granted")
	}
	if got.allows("work_miracle") {
		t.Error("work_miracle should stay excluded — the world never granted it")
	}
	if got.allowsBundle("gandalf_bless") {
		t.Error("gandalf_bless should stay excluded — the world's own explicit tools list never named it, so the persona cannot widen to it")
	}
}

// TestIntersectGrantNarrowsToolsUnconstrainedWorld: when the world is
// UNCONSTRAINED (no capabilities.json, every built-in and bundle tool
// granted), a persona's explicit "tools" list becomes the effective roster —
// exactly like an equivalent world-level capabilities.json would (same
// schema, same semantics, reused rather than reinvented).
func TestIntersectGrantNarrowsToolsUnconstrainedWorld(t *testing.T) {
	g := fullGrant()

	got := intersectGrant(g, &bundle.GrantDoc{Tools: []string{"send_vision", "gandalf_bless"}})

	if !got.allows("send_vision") {
		t.Error("send_vision should remain granted — named by the persona")
	}
	if got.allows("work_miracle") {
		t.Error("work_miracle should be excluded — the persona's explicit list did not name it")
	}
	if !got.allowsBundle("gandalf_bless") {
		t.Error("gandalf_bless should be granted — the persona's own tool, explicitly named")
	}
}

// TestNarrowGrantForBundlesAbsentFile: a bundle set where no bundle declares a
// capabilities.json contributes NO narrowing — the world grant passes through
// unchanged. Covers both a nil BundleSet and one whose bundles carry no Grant.
func TestNarrowGrantForBundlesAbsentFile(t *testing.T) {
	g := fullGrant()

	if got := narrowGrantForBundles(g, nil); !equivalentGrants(got, g) {
		t.Errorf("nil BundleSet narrowed the grant: got %+v, want unchanged %+v", got, g)
	}

	bs := bundleWorld(t, "valid") // fixture bundles carry no capabilities.json
	if got := narrowGrantForBundles(g, bs); !equivalentGrants(got, g) {
		t.Errorf("a Grant-less BundleSet narrowed the grant: got %+v, want unchanged %+v", got, g)
	}
}

// TestNarrowGrantForBundlesAppliesPersonaFixture: the T031 persona fixture's
// capabilities.json (miracle_kinds narrowing only) applies through the full
// narrowGrantForBundles seam against an unrestricted world grant.
func TestNarrowGrantForBundlesAppliesPersonaFixture(t *testing.T) {
	g := fullGrant()
	bs := bundleWorld(t, "persona")

	got := narrowGrantForBundles(g, bs)
	if !got.kindsRestricted {
		t.Fatal("kindsRestricted = false, want true after the persona fixture's narrowing")
	}
	if got.allowsKind("move") {
		t.Error("move should be excluded — the gandalf fixture's capabilities.json does not name it")
	}
	for _, k := range []string{"remove", "give_item", "time_snap"} {
		if !got.allowsKind(k) {
			t.Errorf("%s should remain granted", k)
		}
	}
	// The fixture narrows kinds only — tools are untouched.
	if got.toolsConstrained {
		t.Error("toolsConstrained = true, want false — the fixture's capabilities.json does not touch tools")
	}
}

// TestLoadManifestSuppressesKnownBundleToolNotice (handoff fix, spec 036
// T030): a world-level capabilities.json naming a REAL bundle tool must not
// render a cosmetic "unknown tool … ignored" notice — allowsBundle already
// grants it correctly. A genuinely unknown name still gets its notice.
func TestLoadManifestSuppressesKnownBundleToolNotice(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, `{"tools":["send_vision","gandalf_bless","totally_unknown"]}`)

	g, notices := loadManifest(dir, "gandalf_bless")
	if !g.allowsBundle("gandalf_bless") {
		t.Error("gandalf_bless should be granted")
	}
	for _, n := range notices {
		if strings.Contains(n, "gandalf_bless") {
			t.Errorf("notice wrongly flags the known bundle tool gandalf_bless: %v", notices)
		}
	}
	var foundUnknown bool
	for _, n := range notices {
		if strings.Contains(n, "totally_unknown") {
			foundUnknown = true
		}
	}
	if !foundUnknown {
		t.Errorf("expected a notice naming the genuinely unknown tool totally_unknown: %v", notices)
	}
}

// equivalentGrants compares the grantSet fields that matter for narrowing
// (map-valued fields, so reflect.DeepEqual on the zero-value nils vs
// initialized-empty maps would over-fit); "no narrowing happened" is judged by
// gate behavior, not literal map identity.
func equivalentGrants(a, b grantSet) bool {
	if a.toolsConstrained != b.toolsConstrained || a.kindsRestricted != b.kindsRestricted {
		return false
	}
	for _, name := range []string{"send_vision", "send_omen", "work_miracle", "monitor_and_act", "cancel_order", "pause", "start", "adjust_speed"} {
		if a.allows(name) != b.allows(name) {
			return false
		}
	}
	for _, k := range []string{"move", "remove", "give_item", "time_snap"} {
		if a.allowsKind(k) != b.allowsKind(k) {
			return false
		}
	}
	return true
}
