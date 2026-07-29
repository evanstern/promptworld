package bundle

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/tool"
)

func discover(t *testing.T, world string) *BundleSet {
	t.Helper()
	bs, err := Discover(filepath.Join("testdata", "worlds", world))
	if err != nil {
		t.Fatalf("Discover(%s): %v", world, err)
	}
	return bs
}

func rosterNames(bs *BundleSet) []string {
	tools := bs.Roster()
	out := make([]string, len(tools))
	for i, t := range tools {
		out[i] = t.Name
	}
	return out
}

// findIssue returns the first BootReport entry matching rule and tool ("" = any
// tool), or nil.
func findIssue(bs *BundleSet, rule, tool string) *BootIssue {
	for i := range bs.report {
		if bs.report[i].Rule == rule && (tool == "" || bs.report[i].Tool == tool) {
			return &bs.report[i]
		}
	}
	return nil
}

// TestDiscoverValid: two valid bundles load; the roster is in deterministic
// order (bundle load order × ascending tool order) and no issues are reported.
func TestDiscoverValid(t *testing.T) {
	bs := discover(t, "valid")
	want := []string{"teleport", "whisper", "gift"} // aaa{teleport,whisper}, bbb{gift}
	if got := rosterNames(bs); !reflect.DeepEqual(got, want) {
		t.Errorf("roster = %v, want %v", got, want)
	}
	if len(bs.BootReport()) != 0 {
		t.Errorf("boot report = %+v, want empty", bs.BootReport())
	}
	// Templates were parsed and cached on each tool.
	for _, b := range bs.Bundles() {
		for _, tl := range b.Tools {
			if len(tl.Templates) == 0 {
				t.Errorf("tool %q has no cached templates", tl.Name)
			}
		}
	}
}

// TestDiscoverDeterministic: the same world dir yields byte-identical roster and
// report ordering on repeated discovery (replay + SC-005 determinism).
func TestDiscoverDeterministic(t *testing.T) {
	a := discover(t, "collision")
	b := discover(t, "collision")
	if !reflect.DeepEqual(rosterNames(a), rosterNames(b)) {
		t.Errorf("roster order not deterministic: %v vs %v", rosterNames(a), rosterNames(b))
	}
	if !reflect.DeepEqual(a.BootReport(), b.BootReport()) {
		t.Errorf("boot report order not deterministic")
	}
}

// TestOffWhitelistSkipsTool (T3): a tool declaring a non-injectable event type
// is skipped; its valid sibling stays on the roster; the report names the file
// and the offending event value (SC-005).
func TestOffWhitelistSkipsTool(t *testing.T) {
	bs := discover(t, "offwhitelist")
	if got := rosterNames(bs); !reflect.DeepEqual(got, []string{"goodtool"}) {
		t.Errorf("roster = %v, want [goodtool]", got)
	}
	iss := findIssue(bs, "T3", "healtool")
	if iss == nil {
		t.Fatal("no T3 issue for healtool")
	}
	if iss.File != filepath.Join("demo", "tools", "healtool", "tool.json") {
		t.Errorf("file = %q", iss.File)
	}
	if iss.Severity != "error" || !strings.Contains(iss.Message, "guardian.heal") || !strings.Contains(iss.Message, "not an injectable event type") {
		t.Errorf("message = %q", iss.Message)
	}
}

// TestLegacyEventVocabularyNormalizes (spec 094 D4's config-reference
// posture): a bundle authored against the pre-094 metatron.* vocabulary
// still loads — declared event names normalize through the log-format
// rename table before the T3 whitelist check, and the compiled tool carries
// the CURRENT names.
func TestLegacyEventVocabularyNormalizes(t *testing.T) {
	bs := discover(t, "legacyvocab")
	if got := rosterNames(bs); !reflect.DeepEqual(got, []string{"porter"}) {
		t.Fatalf("roster = %v, want [porter] (boot report: %+v)", got, bs.BootReport())
	}
	if len(bs.BootReport()) != 0 {
		t.Errorf("boot report = %+v, want empty", bs.BootReport())
	}
	porter := bs.Roster()[0]
	if !reflect.DeepEqual(porter.Events, []string{"guardian.entity_moved", "agent.memory_added"}) {
		t.Errorf("compiled events = %v, want the normalized current names", porter.Events)
	}
}

// TestBadSoulRejectsBundle (B2): an oversized SOUL.md rejects the whole bundle;
// its otherwise-valid tool never reaches the roster.
func TestBadSoulRejectsBundle(t *testing.T) {
	bs := discover(t, "badsoul")
	if got := rosterNames(bs); len(got) != 0 {
		t.Errorf("roster = %v, want empty (bundle rejected)", got)
	}
	iss := findIssue(bs, "B2", "")
	if iss == nil {
		t.Fatal("no B2 issue")
	}
	if iss.File != filepath.Join("persona", "SOUL.md") || !strings.Contains(iss.Message, "4001") {
		t.Errorf("issue = %+v", *iss)
	}
}

// TestBadCapsRejectsBundle (B3): a malformed capabilities.json (unknown key)
// rejects the whole bundle.
func TestBadCapsRejectsBundle(t *testing.T) {
	bs := discover(t, "badcaps")
	if got := rosterNames(bs); len(got) != 0 {
		t.Errorf("roster = %v, want empty (bundle rejected)", got)
	}
	iss := findIssue(bs, "B3", "")
	if iss == nil {
		t.Fatal("no B3 issue")
	}
	if iss.File != filepath.Join("persona", "capabilities.json") || iss.Severity != "error" {
		t.Errorf("issue = %+v", *iss)
	}
}

// TestCollisions (C1/C2): a bundle tool named like a built-in loses (C1); among
// two bundles claiming the same name, the first-loaded wins (C2). Both losers
// are warnings, not errors.
func TestCollisions(t *testing.T) {
	bs := discover(t, "collision")
	if got := rosterNames(bs); !reflect.DeepEqual(got, []string{"dup"}) {
		t.Errorf("roster = %v, want [dup] (say lost C1, ccc/dup lost C2)", got)
	}
	c1 := findIssue(bs, "C1", "say")
	if c1 == nil || c1.Severity != "warning" || !strings.Contains(c1.Message, "built-in") {
		t.Errorf("C1 issue = %+v", c1)
	}
	c2 := findIssue(bs, "C2", "dup")
	if c2 == nil || c2.Severity != "warning" || c2.Bundle != "ccc" || !strings.Contains(c2.Message, "bbb") {
		t.Errorf("C2 issue = %+v", c2)
	}
}

// TestScriptedToolValidation (T6): a script defining apply() loads; a script
// missing apply() is skipped with a T6 issue naming the problem.
func TestScriptedToolValidation(t *testing.T) {
	bs := discover(t, "scripted")
	// Tool dirs load in ascending name order: broken (skipped, no apply), then the
	// valid cast_light and lamp.
	if got := rosterNames(bs); !reflect.DeepEqual(got, []string{"cast_light", "lamp"}) {
		t.Errorf("roster = %v, want [cast_light lamp]", got)
	}
	iss := findIssue(bs, "T6", "broken")
	if iss == nil || !strings.Contains(iss.Message, "does not define apply()") {
		t.Errorf("T6 issue = %+v", iss)
	}
	// lamp is script mode: no cached templates, manifest carries the script ref.
	for _, b := range bs.Bundles() {
		for _, tl := range b.Tools {
			if tl.Name == "lamp" && (len(tl.Templates) != 0 || tl.Manifest.Script != "tool.star") {
				t.Errorf("lamp tool = %+v", tl)
			}
		}
	}
}

// TestBuiltinNameCollision (T021, US2 / clarification #2): a bundle tool named
// exactly like a built-in (work_miracle — the dogfood-relevant name) loses to C1,
// the built-in wins, and there is exactly ONE work_miracle system-wide (the
// built-in in tool.Lookup, never a second bundle roster entry). Its valid
// bundlemate survives, proving the collision skips only the offender.
func TestBuiltinNameCollision(t *testing.T) {
	bs := discover(t, "builtincollision")

	// work_miracle is skipped (C1); the valid sibling reaches the roster.
	if got := rosterNames(bs); !reflect.DeepEqual(got, []string{"aura"}) {
		t.Errorf("roster = %v, want [aura] (work_miracle lost C1)", got)
	}

	c1 := findIssue(bs, "C1", "work_miracle")
	if c1 == nil || c1.Severity != "warning" || !strings.Contains(c1.Message, "built-in") {
		t.Errorf("C1 issue = %+v, want a warning naming the built-in", c1)
	}

	// Exactly one work_miracle: the built-in remains registered, and the bundle
	// roster never carries a second one.
	if _, ok := tool.Lookup("work_miracle"); !ok {
		t.Fatal("built-in work_miracle is missing from the registry")
	}
	for _, n := range rosterNames(bs) {
		if n == "work_miracle" {
			t.Error("bundle roster leaked a second work_miracle entry — the built-in must win")
		}
	}
}

// TestAbsentBundlesDir: a world with no bundles/ dir yields an empty set and no
// error (bundles are additive).
func TestAbsentBundlesDir(t *testing.T) {
	bs, err := Discover(t.TempDir())
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(bs.Roster()) != 0 || len(bs.BootReport()) != 0 {
		t.Errorf("roster/report not empty: %v / %v", bs.Roster(), bs.BootReport())
	}
}

// TestDiscoverPersona (spec 036 US4, T031 fixture): the gandalf persona bundle
// surfaces its SOUL fragment and grant narrowing, keeps its valid tool, and
// per-tool-skips its deliberately broken sibling (clarification #1) — SOUL and
// grant stay intact even though one tool failed.
func TestDiscoverPersona(t *testing.T) {
	bs := discover(t, "persona")

	if got := rosterNames(bs); !reflect.DeepEqual(got, []string{"bless"}) {
		t.Fatalf("roster = %v, want [bless] (broken tool T1-skipped)", got)
	}

	bundles := bs.Bundles()
	if len(bundles) != 1 || bundles[0].Name != "gandalf" {
		t.Fatalf("bundles = %+v, want one bundle named gandalf", bundles)
	}
	b := bundles[0]
	if !strings.Contains(b.Soul, "Gandalf") {
		t.Errorf("Soul = %q, want it to contain the persona fragment", b.Soul)
	}
	if b.Grant == nil || !reflect.DeepEqual(b.Grant.MiracleKinds, []string{"remove", "give_item", "time_snap"}) {
		t.Errorf("Grant = %+v, want miracle_kinds [remove give_item time_snap]", b.Grant)
	}
	if b.Grant.Tools != nil {
		t.Errorf("Grant.Tools = %v, want nil (the fixture narrows kinds only)", b.Grant.Tools)
	}

	got := bs.SoulFragments()
	if len(got) != 1 || !strings.Contains(got[0], "Gandalf") {
		t.Errorf("SoulFragments() = %v, want one fragment naming Gandalf", got)
	}

	iss := findIssue(bs, "T1", "broken")
	if iss == nil {
		t.Fatalf("no T1 issue for the broken tool; report=%+v", bs.BootReport())
	}
	if iss.Severity != "error" || !strings.Contains(iss.Message, "brokentool") {
		t.Errorf("issue = %+v", *iss)
	}
}
