package worlds

import (
	"os"
	"testing"
)

// TestLoadUnlocksMissingFileIsEmpty (T015): the registry doctrine's first
// obligation — a missing unlocks.json is an empty record, not an error.
func TestLoadUnlocksMissingFileIsEmpty(t *testing.T) {
	setHome(t)
	u := LoadUnlocks()
	if len(u.Entries) != 0 {
		t.Errorf("expected empty unlocks record, got %v", u.Entries)
	}
	if u.Earned("stage-2") {
		t.Error("a fresh player should have nothing earned")
	}
}

// TestLoadUnlocksCorruptFileIsEmpty (T015): malformed JSON degrades to empty,
// never an error.
func TestLoadUnlocksCorruptFileIsEmpty(t *testing.T) {
	setHome(t)
	path, err := UnlocksPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(rootDirOf(t, path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	u := LoadUnlocks()
	if len(u.Entries) != 0 {
		t.Errorf("expected empty unlocks record from corrupt file, got %v", u.Entries)
	}
}

// TestUpsertUnlocksWritesAndReloads (T015): the write/reload round-trip and
// the append-shaped doctrine — earning one stage never disturbs another's
// entry.
func TestUpsertUnlocksWritesAndReloads(t *testing.T) {
	setHome(t)
	e2 := UnlockEntry{World: "demo", Path: "/worlds/demo", Exercise: "first-night",
		Evidence: []UnlockEvidenceRef{{Type: "curriculum.exercise_passed", Seq: 4812, Tick: 86400}},
		EarnedAt: "2026-07-25T18:00:00Z"}
	UpsertUnlock("stage-2", e2)

	u := LoadUnlocks()
	if !u.Earned("stage-2") {
		t.Fatal("expected stage-2 earned after upsert")
	}
	if got := u.Entries["stage-2"]; got.World != "demo" || got.Exercise != "first-night" {
		t.Errorf("stage-2 entry = %+v", got)
	}

	// Earning a second stage must not disturb the first.
	e3 := UnlockEntry{World: "demo2", Path: "/worlds/demo2", Exercise: "the-law", EarnedAt: "2026-07-26T09:00:00Z"}
	UpsertUnlock("stage-3", e3)
	u = LoadUnlocks()
	if !u.Earned("stage-2") || !u.Earned("stage-3") {
		t.Errorf("expected both stage-2 and stage-3 earned, got %v", u.Entries)
	}
}

// TestUpsertUnlocksOverwritesSameStage (T015): re-observing a pass for an
// already-earned stage (e.g. proven again in a different world) updates
// that stage's entry — upsert semantics, not "first write wins".
func TestUpsertUnlocksOverwritesSameStage(t *testing.T) {
	setHome(t)
	UpsertUnlock("stage-2", UnlockEntry{World: "first", Path: "/w/first", Exercise: "first-night", EarnedAt: "2026-07-25T18:00:00Z"})
	UpsertUnlock("stage-2", UnlockEntry{World: "second", Path: "/w/second", Exercise: "first-night", EarnedAt: "2026-07-26T18:00:00Z"})

	u := LoadUnlocks()
	if got := u.Entries["stage-2"].World; got != "second" {
		t.Errorf("stage-2 world after re-upsert = %q, want %q", got, "second")
	}
}

// TestLoadUnlocksHealsMalformedEntriesButKeepsMissingWorlds (T015): an entry
// missing required identity fields is dropped at load (malformed); an entry
// whose world path no longer exists on disk is KEPT — an archived or moved
// world is still historical proof (contract rule 3).
func TestLoadUnlocksHealsMalformedEntriesButKeepsMissingWorlds(t *testing.T) {
	setHome(t)
	path, err := UnlocksPath()
	if err != nil {
		t.Fatal(err)
	}
	raw := `{"unlocks":{
		"stage-2":{"world":"gone","path":"/does/not/exist","exercise":"first-night","earned_at":"2026-07-25T18:00:00Z"},
		"stage-3":{"world":"","path":"","exercise":"","earned_at":""}
	}}`
	if err := os.MkdirAll(rootDirOf(t, path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	u := LoadUnlocks()
	if !u.Earned("stage-2") {
		t.Error("an entry pointing at a since-removed world must survive (historical proof)")
	}
	if u.Earned("stage-3") {
		t.Error("a malformed entry (empty identity fields) must be dropped")
	}
}

// TestUpsertUnlocksUnresolvableHomeWarnsAndContinues (T015, lease.go
// precedent): when the home directory cannot be resolved, UpsertUnlock warns
// (never panics, never blocks the caller) and LoadUnlocks degrades to an
// empty record rather than erroring.
func TestUpsertUnlocksUnresolvableHomeWarnsAndContinues(t *testing.T) {
	var warned bool
	orig := unlocksWarnf
	unlocksWarnf = func(format string, args ...any) { warned = true }
	defer func() { unlocksWarnf = orig }()

	t.Setenv("PROMPTWORLD_HOME", "")
	t.Setenv("HOME", "")
	// os.UserHomeDir() on most platforms consults $HOME; forcing it empty
	// (and PROMPTWORLD_HOME empty) makes Root() fail deterministically.
	UpsertUnlock("stage-2", UnlockEntry{World: "x", Path: "/x", Exercise: "first-night", EarnedAt: "2026-07-25T18:00:00Z"})
	if !warned {
		t.Skip("this platform resolves a home directory even with $HOME unset — nothing to assert")
	}
	u := LoadUnlocks()
	if u.Earned("stage-2") {
		t.Error("an unresolvable-home upsert must not have recorded anything")
	}
}

// rootDirOf returns the directory portion of an unlocks.json path — a tiny
// helper so tests can pre-create the home dir before writing directly.
func rootDirOf(t *testing.T, path string) string {
	t.Helper()
	return path[:len(path)-len("/unlocks.json")]
}
