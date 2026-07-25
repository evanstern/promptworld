package guardian

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/tool"
)

// Unit tests for the turnSystemPrompt seam's persona SOUL.md composition (spec
// 036 US4 T029): fragments compose immediately after the charter section, in
// bundle load order, still beneath the fixed frame (INV-1). The full
// through-a-real-bundle path (a real SOUL.md landing on a live Turn's system
// prompt) is covered by TestBundlePersonaComposesIdentityGrantAndTools in
// bundle_integration_test.go (T032); these tests isolate the composition
// function itself, following the existing prompt-assertion style
// (TestFixedFrameHolds, TestInitiativeFrameFixed) in metatron_test.go.

// TestSoulFragmentsAppearAfterCharter: souls compose right after the charter,
// in the order passed (bundle load order at the call site), and the frame
// still lands last regardless.
func TestSoulFragmentsAppearAfterCharter(t *testing.T) {
	roster := tool.LoopRosterGuardian()
	prompt := turnSystemPrompt("CHARTER-MARKER", nil, roster, "SOUL-FRAGMENT-ONE", "SOUL-FRAGMENT-TWO")

	charterAt := strings.Index(prompt, "CHARTER-MARKER")
	oneAt := strings.Index(prompt, "SOUL-FRAGMENT-ONE")
	twoAt := strings.Index(prompt, "SOUL-FRAGMENT-TWO")
	if charterAt < 0 || oneAt < 0 || twoAt < 0 {
		t.Fatalf("missing marker(s) in prompt:\n%s", prompt)
	}
	if !(charterAt < oneAt && oneAt < twoAt) {
		t.Errorf("souls not composed after the charter in load order: charter@%d one@%d two@%d", charterAt, oneAt, twoAt)
	}
	// Still beneath the fixed frame (INV-1): a persona fragment is editable
	// content like the charter/skills, never able to displace the frame.
	fixedFrameLast(t, prompt, "SOUL-FRAGMENT-ONE", "SOUL-FRAGMENT-TWO")
}

// TestSoulFragmentsBeforeSkills: souls compose between the charter and the
// skill files, not after them — "after the charter section" per the T029 seam.
func TestSoulFragmentsBeforeSkills(t *testing.T) {
	roster := tool.LoopRosterGuardian()
	skills := []skillFile{{name: "10-x.md", text: "SKILL-TEXT"}}
	prompt := turnSystemPrompt("CHARTER-MARKER", skills, roster, "SOUL-FRAGMENT")

	soulAt := strings.Index(prompt, "SOUL-FRAGMENT")
	skillAt := strings.Index(prompt, "SKILL-TEXT")
	if soulAt < 0 || skillAt < 0 {
		t.Fatalf("missing marker(s) in prompt:\n%s", prompt)
	}
	if soulAt > skillAt {
		t.Errorf("soul fragment (at %d) appears after the skill section (at %d)", soulAt, skillAt)
	}
}

// TestNoSoulsUnchangedFromPre036: omitting souls entirely — every pre-036
// call site in metatron_test.go does exactly this — composes deterministically
// and adds no persona separator, the "extend, don't break" contract for this
// seam (handoff note b).
func TestNoSoulsUnchangedFromPre036(t *testing.T) {
	roster := tool.LoopRosterGuardian()
	a := turnSystemPrompt("CHARTER", nil, roster)
	b := turnSystemPrompt("CHARTER", nil, roster)
	if a != b {
		t.Error("turnSystemPrompt is not deterministic with no souls")
	}
	if strings.Contains(a, "--- persona ---") {
		t.Error("no-souls prompt unexpectedly contains a persona separator")
	}
}
