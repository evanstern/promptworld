package metatron

import (
	"testing"

	"github.com/evanstern/promptworld/internal/skin"
)

// TestStatusCarriesSkinFacts (spec 052 T005, contract §7): the model-free
// status peek carries the resolved skin display facts — identity fields
// always (resolved against the default table), override maps only when a
// world skin overrides — so clients render skin vocabulary without reading
// world files (FR-012).
func TestStatusCarriesSkinFacts(t *testing.T) {
	mt, _, _, _ := newTestAngel(t, "watching")

	// Never-installed skin (every pre-052 world): default facts, no maps.
	s := mt.Status()
	if s.SkinName != "Guardian" || s.SkinEpithet != "guardian" || s.SkinTabLabel != "guardian" || s.SkinFamilyLabel != "guardian" {
		t.Errorf("default skin facts wrong: %q/%q/%q/%q", s.SkinName, s.SkinEpithet, s.SkinTabLabel, s.SkinFamilyLabel)
	}
	if s.SkinStrings != nil || s.SkinStages != nil {
		t.Errorf("default skin must omit override maps: %v %v", s.SkinStrings, s.SkinStages)
	}

	// A world skin's facts ride through.
	sk, notices := skin.Parse([]byte(`{"name": "Raven", "epithet": "raven", "tab_label": "raven",
		"stages": {"stage-1": {"name": "The Whisper"}}}`))
	if len(notices) != 0 {
		t.Fatalf("unexpected notices: %v", notices)
	}
	mt.SetSkin(sk)
	s = mt.Status()
	if s.SkinName != "Raven" || s.SkinEpithet != "raven" || s.SkinTabLabel != "raven" {
		t.Errorf("world skin facts wrong: %q/%q/%q", s.SkinName, s.SkinEpithet, s.SkinTabLabel)
	}
	if s.SkinStages["stage-1"].Name != "The Whisper" {
		t.Errorf("stage override missing from status: %+v", s.SkinStages)
	}
	if _, ok := s.SkinStrings[skin.TokenName]; !ok {
		t.Errorf("identity overrides must ride SkinStrings: %v", s.SkinStrings)
	}
}
