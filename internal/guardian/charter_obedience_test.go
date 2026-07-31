package guardian

// Spec 107 T004 (D5, FR-004): the EASY-mode default charter — the compiled
// default executes the player's missions without editorializing; counsel-
// first behavior is skinned-charter DATA, structurally absent from the
// default; and the retired counsel-first seed keeps reading as game-authored
// on upgrade (the spec-052 SC-003 discipline), so the spec-102 ceiling and
// the unlock gates never misclassify an untouched pre-107 world.

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/persona"
)

// TestDefaultCharterCarriesObedienceClause pins the D5 edit's presence: the
// default charter obeys direct orders and missions without editorializing,
// and its ONE sanctioned counsel case is the impossible-as-stated refusal
// that names the blocking fact.
func TestDefaultCharterCarriesObedienceClause(t *testing.T) {
	flat := strings.Join(strings.Fields(persona.DefaultCharter), " ")
	for _, want := range []string{
		"without editorializing",
		"Do not counsel in place of compliance",
		"IMPOSSIBLE as it stands",
		"name the exact blocking fact",
		"accept it and pursue it across your watches",
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("default charter missing the obedience clause fragment %q", want)
		}
	}
	// The counsel-loop's source text (TASK-166 behavioral note) is GONE from
	// the default: "propose a wiser method" as an unconditional duty was the
	// observed 4-turn counseling's charter grounding. Counsel-first behavior
	// is skinned-charter data now (TASK-121), demonstrably absent here.
	if strings.Contains(persona.DefaultCharter, "propose a wiser method") {
		t.Error("default charter still carries the counsel-first duty — the EASY-mode edit did not land")
	}
	if len(persona.DefaultCharter) > persona.CharterMaxChars {
		t.Errorf("default charter %d chars exceeds the %d cap", len(persona.DefaultCharter), persona.CharterMaxChars)
	}
}

// TestPre107DefaultStaysGameAuthored pins the upgrade posture: an untouched
// pre-107 world's charter.md (the counsel-first seed) still classifies as
// the game's own text — default for status/observation, ceiling ON for the
// scheduled lane — never reclassified player-authored by the D5 edit.
func TestPre107DefaultStaysGameAuthored(t *testing.T) {
	if !isLegacyDefault(persona.LegacyDefaultCharterCounsel) {
		t.Fatal("pre-107 counsel-first seed not recognized as a legacy default")
	}
	if angelCharterLifted(persona.LegacyDefaultCharterCounsel, "") {
		t.Fatal("pre-107 seed lifted the scheduled-lane ceiling — competence must still be bought with authorship")
	}
	if angelCharterLifted(persona.DefaultCharter, "") {
		t.Fatal("the new default lifted its own ceiling")
	}
	// A skinned/authored charter still lifts — the D5 edit narrows nothing
	// about authorship.
	if !angelCharterLifted("You are a stern keeper who counsels before every act.", "") {
		t.Fatal("authored text no longer lifts the ceiling")
	}
}
