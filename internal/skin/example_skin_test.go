package skin

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExampleRavenSkinLoads (spec 052 FR-014, T017): the in-repo example
// alternate skin is the format's living documentation — it must load with
// ZERO notices and re-theme the display surfaces end-to-end.
func TestExampleRavenSkinLoads(t *testing.T) {
	root := repoRoot(t)
	dir := t.TempDir()
	data, err := os.ReadFile(filepath.Join(root, "examples", "skins", "raven.json"))
	if err != nil {
		t.Fatalf("example skin missing: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skin.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	s, notices := Load(dir)
	if len(notices) != 0 {
		t.Fatalf("the shipped example must load clean, got notices: %v", notices)
	}
	if s.Name() != "the Raven" || s.Epithet() != "raven" || s.TabLabel() != "raven" || s.FamilyLabel() != "raven" {
		t.Errorf("identity fields: %q/%q/%q/%q", s.Name(), s.Epithet(), s.TabLabel(), s.FamilyLabel())
	}
	if s.WorkingNoun() != "trick" || s.FormNoun("vision") != "whisper" || s.FormNoun("omen") != "wingbeat" {
		t.Errorf("vocabulary overrides: %q/%q/%q", s.WorkingNoun(), s.FormNoun("vision"), s.FormNoun("omen"))
	}
	if s.Voice() == "" {
		t.Error("example skin should carry a voice")
	}
	if s.StageName("stage-2") != "The Bargain" || s.StageName("stage-4") != "The Long Flight" {
		t.Errorf("stage identities: %q/%q", s.StageName("stage-2"), s.StageName("stage-4"))
	}

	// The status round-trip re-themes a client identically (contract §7).
	rebuilt := FromFacts(s.StringOverrides(), s.StageOverrides())
	for _, tok := range []string{TokenName, TokenEpithet, TokenTabLabel, "skin.guardian.working_noun"} {
		if rebuilt.Resolve(tok) != s.Resolve(tok) {
			t.Errorf("round-trip drift on %s: %q vs %q", tok, rebuilt.Resolve(tok), s.Resolve(tok))
		}
	}
}
