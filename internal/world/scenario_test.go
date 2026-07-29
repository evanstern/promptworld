package world

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
)

// Scenario-block tests (spec 054 T006): the manifest side of the machinery —
// validation at Open, the write-once SetScenario stamp, and the twin-list
// sync check against the sim catalog.

// TestScenarioVocabularyMirrorsSimCatalog pins ValidScenarioExercise's local
// list to sim.ScenarioExercises' ids, both directions — the two packages
// deliberately do not import each other (the validLadderStage twin-list
// precedent), so this test is what keeps the vocabularies one vocabulary.
func TestScenarioVocabularyMirrorsSimCatalog(t *testing.T) {
	catalog := map[string]bool{}
	for _, def := range sim.ScenarioExercises {
		catalog[def.ID] = true
		if !ValidScenarioExercise(def.ID) {
			t.Errorf("cataloged exercise %q is not in world.ValidScenarioExercise's mirror", def.ID)
		}
	}
	for _, id := range []string{"first-night", "the-law"} {
		if ValidScenarioExercise(id) && !catalog[id] {
			t.Errorf("world mirror accepts %q but the sim catalog does not ship it", id)
		}
	}
	if ValidScenarioExercise("") || ValidScenarioExercise("bogus") {
		t.Error("mirror must reject empty and unknown ids")
	}
}

// TestScenarioRoundTrip: a stamped scenario block survives Open, and a fresh
// Create never writes the key (omitempty — the stage-keys byte-compat
// precedent).
func TestScenarioRoundTrip(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "w")
	if _, err := Create(dir, "scen", 46101); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ManifestName))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "scenario") {
		t.Errorf("fresh manifest must not carry the scenario key, got:\n%s", data)
	}
	if err := SetScenario(dir, "first-night"); err != nil {
		t.Fatalf("SetScenario: %v", err)
	}
	w, err := Open(dir)
	if err != nil {
		t.Fatalf("Open after SetScenario: %v", err)
	}
	if w.Manifest.Scenario == nil || w.Manifest.Scenario.Exercise != "first-night" {
		t.Errorf("Scenario = %+v, want first-night", w.Manifest.Scenario)
	}
}

// TestOpenRejectsBadScenario: the block is a closed vocabulary — an unknown
// (or empty) exercise id is refused at Open naming the value, never silently
// booted ambient (the bad-stage precedent).
func TestOpenRejectsBadScenario(t *testing.T) {
	for _, bad := range []string{"bogus-exercise", ""} {
		dir := t.TempDir()
		manifest := `{"name":"x","seed":1,"format_version":6,"tick_game_seconds":1,` +
			`"scenario":{"exercise":"` + bad + `"}}`
		if err := os.WriteFile(filepath.Join(dir, ManifestName), []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := Open(dir)
		if err == nil {
			t.Fatalf("Open should reject scenario exercise %q", bad)
		}
		if !strings.Contains(err.Error(), "scenario") {
			t.Errorf("refusal should name the scenario block, got: %v", err)
		}
	}
}
