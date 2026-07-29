package store

// The log-level format-version stamp (spec 094): one row in the meta table,
// key "log_format_version", written at genesis — so a log is self-describing
// independent of its world.json manifest, and a version-mismatched log is
// refused at load instead of mis-replayed. One row describes the whole
// single-file log; append paths need nothing.
//
// DOCTRINE (spec 094, reconciled with spec 092's emitter-computes doctrine):
// a change to any persisted event-type NAME, or to how the reducer RE-DERIVES
// state from a recorded payload, REQUIRES bumping LogFormatVersion and
// shipping a migration. Pure renames translate (the type column is rewritten,
// every seq/tick/payload preserved — world.Migrate's translating mode);
// semantic breaks snapshot-cut (fresh log carrying the transformed state).
// The world-manifest FormatVersion bumps alongside so world.Open carries the
// refusal at every surface.

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	// LogFormatLegacy is the implicit version of every log written before
	// the stamp existed: the metatron.* guardian vocabulary. Pre-stamp logs
	// carry no meta row and resolve to this version by construction.
	LogFormatLegacy = 1
	// LogFormatVersion 2 is the guardian-rename break (spec 094): the 13
	// persisted metatron.* event types became guardian.* — a pure rename,
	// migrated by translation (`promptworld migrate`), never aliased at
	// read. A v1 log replayed under v2 arms would hit no reducer case;
	// the load gate below makes that impossible instead of silent.
	LogFormatVersion = 2
)

// logFormatKey is the meta-table key carrying the stamp. It lives beside the
// daemon's seed/format_version manifest mirrors (validateMeta) but describes
// the LOG's vocabulary, not the world manifest — three distinct versions:
// world.json format_version (save-directory shape, world.FormatVersion), this
// log stamp (event vocabulary), and the store DDL (table shape, schema.go).
const logFormatKey = "log_format_version"

// ErrLogFormatTooOld: the log speaks an older vocabulary than this build.
// Refused with the migrate hint — the translating migration preserves the
// full history (spec 094 US2), so nothing is lost by upgrading.
type ErrLogFormatTooOld struct {
	Got, Want int
}

func (e *ErrLogFormatTooOld) Error() string {
	return fmt.Sprintf("event log format v%d predates this build (v%d); run 'promptworld migrate <world>' to translate it", e.Got, e.Want)
}

// ErrLogFormatTooNew: the log speaks a NEWER vocabulary than this build — the
// world.go FormatVersion posture: a future vocabulary must never be
// mis-replayed under old reducer arms, so the only remedy is a newer build.
type ErrLogFormatTooNew struct {
	Got, Want int
}

func (e *ErrLogFormatTooNew) Error() string {
	return fmt.Sprintf("event log format v%d is newer than this build supports (v%d); upgrade promptworld to open this world", e.Got, e.Want)
}

// LogFormat reports the log's format version, readable without replay (one
// meta SELECT): the stamped value when present (stamped=true); the implicit
// LogFormatLegacy for an unstamped log that already carries events; and
// (0, false) for an unstamped EMPTY log — a log mid-genesis, before its
// writer stamped it.
func (s *Store) LogFormat() (version int, stamped bool, err error) {
	v, err := s.GetMeta(logFormatKey)
	if err != nil {
		return 0, false, err
	}
	if v == "" {
		if s.LastSeq() > 0 {
			return LogFormatLegacy, false, nil
		}
		return 0, false, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return 0, false, fmt.Errorf("corrupt %s meta value %q", logFormatKey, v)
	}
	return n, true, nil
}

// StampLogFormat writes the current LogFormatVersion stamp. Genesis writers
// (promptworld new, world.Fork, world.Migrate's fresh/translated logs) call
// this exactly once on an empty-or-being-born log; re-stamping the same value
// is a harmless idempotent upsert.
func (s *Store) StampLogFormat() error {
	return s.SetMeta(logFormatKey, strconv.Itoa(LogFormatVersion))
}

// VerifyLogFormat is the load-time gate (spec 094 FR-002), called wherever a
// log is opened for REPLAY under this build's reducer (daemon boot). Older ⇒
// ErrLogFormatTooOld (migrate hint); newer ⇒ ErrLogFormatTooNew (upgrade
// posture). An unstamped empty log is adopted: stamped current and admitted
// (nothing has been recorded under any other vocabulary). Deliberately NOT
// wired into Open: the migration driver and read-only archive tooling must
// still open old logs.
func (s *Store) VerifyLogFormat() error {
	v, stamped, err := s.LogFormat()
	if err != nil {
		return err
	}
	if !stamped && v == 0 {
		return s.StampLogFormat()
	}
	switch {
	case v < LogFormatVersion:
		return &ErrLogFormatTooOld{Got: v, Want: LogFormatVersion}
	case v > LogFormatVersion:
		return &ErrLogFormatTooNew{Got: v, Want: LogFormatVersion}
	}
	return nil
}

// IsLogFormatMismatch reports whether err is either directional log-format
// refusal — the errors.As twin of world.ErrFormatVersionMismatch's contract,
// for callers that need "wrong vocabulary" apart from "broken log".
func IsLogFormatMismatch(err error) bool {
	var tooOld *ErrLogFormatTooOld
	var tooNew *ErrLogFormatTooNew
	return errors.As(err, &tooOld) || errors.As(err, &tooNew)
}
