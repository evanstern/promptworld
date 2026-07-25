package worlds

import (
	"os"
	"testing"
)

// TestLoadLessonsSeenMissingFileIsEmpty (T002): the seen-state contract's
// first obligation — a missing lessons-seen.json is an empty record, not an
// error.
func TestLoadLessonsSeenMissingFileIsEmpty(t *testing.T) {
	setHome(t)
	l := LoadLessonsSeen()
	if len(l.Entries) != 0 {
		t.Errorf("expected empty lessons-seen record, got %v", l.Entries)
	}
	if l.Seen("first-suppression") {
		t.Error("a fresh player should have nothing seen")
	}
}

// TestLoadLessonsSeenCorruptFileIsEmpty (T002): malformed JSON degrades to
// empty, never an error.
func TestLoadLessonsSeenCorruptFileIsEmpty(t *testing.T) {
	setHome(t)
	path, err := LessonsSeenPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(rootDirOf(t, path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	l := LoadLessonsSeen()
	if len(l.Entries) != 0 {
		t.Errorf("expected empty lessons-seen record from corrupt file, got %v", l.Entries)
	}
}

// TestLoadLessonsSeenUnknownVersionIsEmpty (T002): an unrecognized version
// number degrades to empty rather than attempting to interpret a shape it
// doesn't understand.
func TestLoadLessonsSeenUnknownVersionIsEmpty(t *testing.T) {
	setHome(t)
	path, err := LessonsSeenPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(rootDirOf(t, path), 0o755); err != nil {
		t.Fatal(err)
	}
	raw := `{"version":99,"seen":{"first-suppression":{"first_shown":"2026-07-25T18:00:00Z","world":"w"}}}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	l := LoadLessonsSeen()
	if len(l.Entries) != 0 {
		t.Errorf("expected empty record for an unknown version, got %v", l.Entries)
	}
}

// TestMarkLessonSeenWritesAndReloads (T002): the write/reload round-trip and
// the append-shaped doctrine — marking one lesson never disturbs another's
// entry.
func TestMarkLessonSeenWritesAndReloads(t *testing.T) {
	setHome(t)
	MarkLessonSeen("first-suppression", "world-01")

	l := LoadLessonsSeen()
	if !l.Seen("first-suppression") {
		t.Fatal("expected first-suppression seen after MarkLessonSeen")
	}
	if got := l.Entries["first-suppression"]; got.World != "world-01" || got.FirstShown == "" {
		t.Errorf("first-suppression entry = %+v", got)
	}

	// Marking a second lesson must not disturb the first.
	MarkLessonSeen("first-gru-attack", "world-01")
	l = LoadLessonsSeen()
	if !l.Seen("first-suppression") || !l.Seen("first-gru-attack") {
		t.Errorf("expected both lessons seen, got %v", l.Entries)
	}
}

// TestMarkLessonSeenIsUpsert (T002): re-marking an already-seen id (e.g. a
// defensive double-call) updates that id's entry rather than duplicating it
// or erroring — upsert semantics, not "first write wins".
func TestMarkLessonSeenIsUpsert(t *testing.T) {
	setHome(t)
	MarkLessonSeen("first-suppression", "world-01")
	MarkLessonSeen("first-suppression", "world-02")

	l := LoadLessonsSeen()
	if got := l.Entries["first-suppression"].World; got != "world-02" {
		t.Errorf("first-suppression world after re-mark = %q, want %q", got, "world-02")
	}
	if len(l.Entries) != 1 {
		t.Errorf("re-marking the same id should not duplicate entries, got %v", l.Entries)
	}
}

// TestMarkLessonSeenPerUserCrossWorld (T002/SC-001): a lesson marked seen in
// one world stays seen when checked against a record loaded independent of
// any particular world — the contract's "no world component in path or key".
func TestMarkLessonSeenPerUserCrossWorld(t *testing.T) {
	setHome(t)
	MarkLessonSeen("first-death", "world-alpha")
	l := LoadLessonsSeen()
	if !l.Seen("first-death") {
		t.Error("a lesson seen in one world must read as seen regardless of which world checks it")
	}
}

// TestMarkLessonSeenUnresolvableHomeWarnsAndContinues (T002, unlocks.go
// precedent): when the home directory cannot be resolved, MarkLessonSeen
// warns (never panics, never blocks the caller) and LoadLessonsSeen degrades
// to an empty record rather than erroring.
func TestMarkLessonSeenUnresolvableHomeWarnsAndContinues(t *testing.T) {
	var warned bool
	orig := lessonsWarnf
	lessonsWarnf = func(format string, args ...any) { warned = true }
	defer func() { lessonsWarnf = orig }()

	t.Setenv("PROMPTWORLD_HOME", "")
	t.Setenv("HOME", "")
	MarkLessonSeen("first-suppression", "w")
	if !warned {
		t.Skip("this platform resolves a home directory even with $HOME unset — nothing to assert")
	}
	l := LoadLessonsSeen()
	if l.Seen("first-suppression") {
		t.Error("an unresolvable-home mark must not have recorded anything")
	}
}

// TestMarkLessonSeenReadOnlyDirSwallowsWriteFailure (T002): a write failure
// (e.g. a read-only home directory) is swallowed — advisory, never authority
// — never surfaced as an error to the caller.
func TestMarkLessonSeenReadOnlyDirSwallowsWriteFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: permission bits don't block writes")
	}
	dir := setHome(t)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) }) // TempDir cleanup needs write perms back

	var warned bool
	orig := lessonsWarnf
	lessonsWarnf = func(format string, args ...any) { warned = true }
	defer func() { lessonsWarnf = orig }()

	MarkLessonSeen("first-suppression", "w") // must not panic
	if !warned {
		t.Error("expected a warning when the write fails against a read-only home dir")
	}
}
