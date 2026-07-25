package worlds

// The per-user first-occurrence lessons seen-record (spec 055, TASK-117,
// contracts/seen-state-file.md): `<worlds.Root()>/lessons-seen.json`, managed
// with the SAME doctrine as unlocks.json (unlocks.go) — load-tolerant, atomic
// (.tmp+rename) writes, an unresolvable home directory WARNS and degrades
// rather than erroring (the endpoint-lease precedent). This record is purely
// advisory (D8/FR-006): the worst failure mode anywhere in its lifecycle is a
// repeated lesson, never a blocked boot or a player-visible error. No world
// behavior ever reads this file — its only consumer is the TUI's lesson-row
// projection (internal/tui/lessons.go).

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// lessonsSeenVersion is the current on-disk schema version (contract:
// "unknown versions load as empty"). Bumping this is a breaking change to the
// file's shape — every reader must still degrade to empty rather than error.
const lessonsSeenVersion = 1

// LessonsSeenPath returns <root>/lessons-seen.json.
func LessonsSeenPath() (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "lessons-seen.json"), nil
}

// LessonSeenEntry is one lesson id's provenance (data-model.md): read for
// diagnostics only, never for logic — presence in the map alone is the
// suppression signal (contract: "first_shown ... provenance only, never read
// for logic").
type LessonSeenEntry struct {
	FirstShown string `json:"first_shown"`
	World      string `json:"world"`
}

// lessonsSeenFile is the on-disk shape (contracts/seen-state-file.md).
type lessonsSeenFile struct {
	Version int                        `json:"version"`
	Seen    map[string]LessonSeenEntry `json:"seen"`
}

// LessonsSeen is the in-memory, load-tolerant view of the per-user record:
// lesson id -> the entry that recorded it seen.
type LessonsSeen struct {
	Entries map[string]LessonSeenEntry
}

// Seen reports whether id has ever been recorded shown — the only question
// this record is ever asked (mirrors Unlocks.Earned). A nil LessonsSeen (the
// unresolvable-home degrade) reports nothing seen, matching a fresh player's
// honest default — the worst outcome is a repeated lesson, never a false
// suppression.
func (l *LessonsSeen) Seen(id string) bool {
	if l == nil {
		return false
	}
	_, ok := l.Entries[id]
	return ok
}

// lessonsWarnf surfaces a lessons-record warning (warn-not-error, the
// unlocks.go/lease.go precedent): a player whose home directory cannot be
// resolved loses the never-repeat convenience for this run — never fatal,
// since the worst-case failure mode (a repeated lesson) is by contract
// tolerable. Overridable in tests.
var lessonsWarnf = func(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "promptworld: "+format+"\n", args...)
}

// LoadLessonsSeen reads the per-user lessons-seen record. Missing, corrupt,
// or unknown-version yields an empty record, never an error (contract: "Ever.");
// an unresolvable home directory warns and yields an empty record too.
func LoadLessonsSeen() *LessonsSeen {
	path, err := LessonsSeenPath()
	if err != nil {
		lessonsWarnf("could not resolve the lessons-seen record (%v) — proceeding as if nothing has been shown yet", err)
		return &LessonsSeen{Entries: map[string]LessonSeenEntry{}}
	}
	return loadLessonsSeenFrom(path)
}

func loadLessonsSeenFrom(path string) *LessonsSeen {
	l := &LessonsSeen{Entries: map[string]LessonSeenEntry{}}
	data, err := os.ReadFile(path)
	if err != nil {
		return l // missing ⇒ empty, not an error
	}
	var lf lessonsSeenFile
	if err := json.Unmarshal(data, &lf); err != nil {
		return l // corrupt ⇒ empty, not an error
	}
	if lf.Version != lessonsSeenVersion {
		return l // unknown version ⇒ empty, not an error (contract compatibility rule)
	}
	for id, entry := range lf.Seen {
		if id == "" {
			continue
		}
		l.Entries[id] = entry
	}
	return l
}

// MarkLessonSeen records id as shown (upsert semantics, data-model.md
// "absent -> seen only"), writing the record atomically (.tmp + rename,
// unlocks.go precedent). Append-shaped: existing entries for OTHER ids are
// preserved. A failure to resolve the home directory or to write — advisory,
// never authority — warns and returns; the caller's in-memory seen-set still
// prevents repeats for the rest of this session regardless (contract "Write
// timing").
func MarkLessonSeen(id, world string) {
	if id == "" {
		return
	}
	path, err := LessonsSeenPath()
	if err != nil {
		lessonsWarnf("could not resolve the lessons-seen record (%v) — %s not recorded (may repeat next session)", err, id)
		return
	}
	l := loadLessonsSeenFrom(path)
	l.Entries[id] = LessonSeenEntry{FirstShown: time.Now().UTC().Format(time.RFC3339), World: world}
	if err := writeLessonsSeen(path, l); err != nil {
		lessonsWarnf("could not write the lessons-seen record: %v", err)
	}
}

func writeLessonsSeen(path string, l *LessonsSeen) error {
	lf := lessonsSeenFile{Version: lessonsSeenVersion, Seen: l.Entries}
	if lf.Seen == nil {
		lf.Seen = map[string]LessonSeenEntry{}
	}
	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
