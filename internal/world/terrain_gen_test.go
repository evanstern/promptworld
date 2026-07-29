package world

// Spec 068 manifest-compatibility tests (T012): the terrain_gen field, the
// v4→v5 manifest-only migration (C11), Open's refusal posture for unknown
// generations and old formats (C10/C12), and the new-world marking (C12).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// TestNewWorldsMarkedWithTerrainGen (C12): `promptworld new`'s Create writes
// the current format_version + terrain_gen 2, the manifest round-trips
// through Open, and the regenerated map actually carries the new vocabulary.
func TestNewWorldsMarkedWithTerrainGen(t *testing.T) {
	dir := t.TempDir() + "/w"
	w, err := Create(dir, "fresh", 42)
	if err != nil {
		t.Fatal(err)
	}
	if w.Manifest.FormatVersion != FormatVersion || w.Manifest.TerrainGen != worldmap.GenMarshSand {
		t.Fatalf("Create wrote format_version %d, terrain_gen %d — want %d, %d",
			w.Manifest.FormatVersion, w.Manifest.TerrainGen, FormatVersion, worldmap.GenMarshSand)
	}
	raw, err := os.ReadFile(filepath.Join(dir, ManifestName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"terrain_gen": 2`) {
		t.Errorf("world.json must carry the terrain_gen marker unmistakably:\n%s", raw)
	}
	got, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Manifest.TerrainGen != worldmap.GenMarshSand {
		t.Fatalf("terrain_gen did not round-trip: %d", got.Manifest.TerrainGen)
	}
	m := got.Map()
	if m.CountKind(worldmap.Marsh) == 0 || m.CountKind(worldmap.Sand) == 0 {
		t.Error("a new world's map must carry marsh and sand (US2-AS1)")
	}
}

// TestOpenRejectsUnknownTerrainGen (C12): a terrain generation this build
// does not implement is refused with a clear error — never silently
// re-generated under the wrong algorithm.
func TestOpenRejectsUnknownTerrainGen(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ManifestName),
		[]byte(`{"name":"x","seed":1,"format_version":6,"tick_game_seconds":1,"terrain_gen":3}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Open(dir)
	if err == nil {
		t.Fatal("Open must reject an unknown terrain_gen")
	}
	if !strings.Contains(err.Error(), "terrain_gen 3") {
		t.Errorf("refusal should name the unknown generation, got: %v", err)
	}
}

// TestOpenAbsentTerrainGenIsLegacy (C11's read side): a v5 manifest without
// terrain_gen opens as a legacy-terrain world — bit-identical generation.
func TestOpenAbsentTerrainGenIsLegacy(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ManifestName),
		[]byte(`{"name":"x","seed":42,"format_version":6,"tick_game_seconds":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if w.Manifest.TerrainGen != worldmap.GenLegacy {
		t.Fatalf("absent terrain_gen must read as legacy, got %d", w.Manifest.TerrainGen)
	}
	if got, want := w.Map().Hash(), worldmap.Generate(42, worldmap.DefaultSize, worldmap.DefaultSize).Hash(); got != want {
		t.Error("a legacy-terrain world's map must equal the legacy generator's output")
	}
}

// TestOpenRejectsV4WithMigrateHint (C10): the format bump is the refusal
// mechanism — an old-format world is refused at Open with the migrate hint,
// exactly the posture a pre-068 build shows a v5 world. Never a silent load.
func TestOpenRejectsV4WithMigrateHint(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ManifestName),
		[]byte(`{"name":"x","seed":1,"format_version":4,"tick_game_seconds":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Open(dir)
	if err == nil {
		t.Fatal("Open must reject a v4 world")
	}
	if !strings.Contains(err.Error(), "format_version 4 unsupported") || !strings.Contains(err.Error(), "promptworld migrate") {
		t.Errorf("refusal should name the version and the migrate remedy, got: %v", err)
	}
}

// TestMigrateV4PreservesTerrain (C11's surviving guarantee under spec 094):
// the v4→v6 step is now a LOG TRANSLATION, but the terrain contract is
// unchanged — migrate must NOT set terrain_gen, so the regenerated map's
// Hash is identical before and after. (The translation mechanics themselves
// are covered in migrate_translate_test.go.)
func TestMigrateV4PreservesTerrain(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ManifestName),
		[]byte(`{"name":"carried","seed":42,"format_version":4,"tick_game_seconds":1,"map_width":64,"map_height":64}`), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(filepath.Join(dir, "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(sim.WorldCreatedPayload{Name: "carried", Seed: 42})
	if err := st.AppendEvents([]store.Event{{Tick: 0, Type: "world.created", Payload: payload}}); err != nil {
		t.Fatal(err)
	}
	st.Close()

	before, err := OpenForMigration(dir)
	if err != nil {
		t.Fatal(err)
	}
	hashBefore := before.Map().Hash()

	res, err := Migrate(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Translated {
		t.Fatal("a v4 source must take the translation path")
	}

	after, err := Open(dir)
	if err != nil {
		t.Fatalf("migrated world must open under the current build: %v", err)
	}
	if after.Manifest.FormatVersion != FormatVersion {
		t.Fatalf("format_version %d after migrate, want %d", after.Manifest.FormatVersion, FormatVersion)
	}
	if after.Manifest.TerrainGen != worldmap.GenLegacy {
		t.Fatalf("migrate must NOT set terrain_gen (C11), got %d", after.Manifest.TerrainGen)
	}
	if got := after.Map().Hash(); got != hashBefore {
		t.Errorf("migration shifted terrain: %s → %s (C11/SC-006)", hashBefore, got)
	}
	raw, err := os.ReadFile(filepath.Join(dir, ManifestName))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "terrain_gen") {
		t.Errorf("terrain_gen must stay absent from a migrated manifest:\n%s", raw)
	}
	// Idempotence guard: a second run has nothing to migrate.
	if _, err := Migrate(dir); err == nil || !strings.Contains(err.Error(), "nothing to migrate") {
		t.Errorf("re-migrating a current world must refuse, got: %v", err)
	}
}

// TestManifestOmitsTerrainGenWhenLegacy: the omitempty contract — a legacy
// manifest round-trips byte-identically with no terrain_gen key.
func TestManifestOmitsTerrainGenWhenLegacy(t *testing.T) {
	data, err := json.Marshal(Manifest{Name: "x", Seed: 1, FormatVersion: 5, TickGameSeconds: 1})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "terrain_gen") {
		t.Errorf("zero TerrainGen must marshal to an absent key: %s", data)
	}
}
