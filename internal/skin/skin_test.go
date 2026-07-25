package skin

import "testing"

// TestStageIdentities (spec 046 T003): the four ladder stages carry the
// client-approved guardian identities; unknown ids fall back safely.
func TestStageIdentities(t *testing.T) {
	want := map[string]StageIdentity{
		"stage-1": {Name: "The Voice", Line: "you speak, it acts"},
		"stage-2": {Name: "The Written Word", Line: "your law outlives the conversation"},
		"stage-3": {Name: "The Craft", Line: "you shape what it can do"},
		"stage-4": {Name: "The Stewardship", Line: "a world in your care"},
	}
	for id, w := range want {
		got, ok := Stage(id)
		if !ok || got != w {
			t.Errorf("Stage(%q) = %+v, %v; want %+v, true", id, got, ok, w)
		}
		if StageName(id) != w.Name {
			t.Errorf("StageName(%q) = %q, want %q", id, StageName(id), w.Name)
		}
	}
	if _, ok := Stage(""); ok {
		t.Error(`Stage("") should report no identity (pre-ladder world)`)
	}
	if got := StageName("stage-9"); got != "stage-9" {
		t.Errorf("StageName fallback = %q, want the id itself", got)
	}
}
