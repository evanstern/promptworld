package store

// Log-format stamp + load-time enforcement tests (spec 094 T002/T003,
// FR-001/FR-002): the stamp is written at genesis and readable without
// replay; a pre-stamp log resolves to the implicit legacy version; the
// verify gate refuses both directions — older with the migrate hint, newer
// with the upgrade posture.

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func openTemp(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func appendOne(t *testing.T, s *Store) {
	t.Helper()
	if err := s.AppendEvents([]Event{{Tick: 0, Type: "world.created", Payload: json.RawMessage(`{}`)}}); err != nil {
		t.Fatal(err)
	}
}

// TestStampLogFormatAtGenesis (US1 AS-1): a genesis-stamped log reports the
// current version via one meta read — no replay involved.
func TestStampLogFormatAtGenesis(t *testing.T) {
	s := openTemp(t)
	if err := s.StampLogFormat(); err != nil {
		t.Fatal(err)
	}
	appendOne(t, s)
	v, stamped, err := s.LogFormat()
	if err != nil {
		t.Fatal(err)
	}
	if !stamped || v != LogFormatVersion {
		t.Fatalf("LogFormat() = (%d, %v), want (%d, true)", v, stamped, LogFormatVersion)
	}
	if err := s.VerifyLogFormat(); err != nil {
		t.Fatalf("current-format log refused: %v", err)
	}
}

// TestUnstampedNonEmptyLogIsImplicitLegacy (US1 AS-2): a pre-stamp log with
// history resolves to LogFormatLegacy and is refused with the migrate hint —
// never silently replayed under the wrong vocabulary.
func TestUnstampedNonEmptyLogIsImplicitLegacy(t *testing.T) {
	s := openTemp(t)
	appendOne(t, s)
	v, stamped, err := s.LogFormat()
	if err != nil {
		t.Fatal(err)
	}
	if stamped || v != LogFormatLegacy {
		t.Fatalf("LogFormat() = (%d, %v), want (%d, false)", v, stamped, LogFormatLegacy)
	}
	err = s.VerifyLogFormat()
	var tooOld *ErrLogFormatTooOld
	if !errors.As(err, &tooOld) {
		t.Fatalf("VerifyLogFormat() = %v, want ErrLogFormatTooOld", err)
	}
	if tooOld.Got != LogFormatLegacy || tooOld.Want != LogFormatVersion {
		t.Fatalf("refusal carries (%d→%d), want (%d→%d)", tooOld.Got, tooOld.Want, LogFormatLegacy, LogFormatVersion)
	}
	if !strings.Contains(err.Error(), "promptworld migrate") {
		t.Fatalf("older-log refusal must carry the migrate hint, got: %v", err)
	}
	if !IsLogFormatMismatch(err) {
		t.Error("IsLogFormatMismatch must match the too-old refusal")
	}
}

// TestNewerLogRefusedWithUpgradePosture (US1 AS-3, SC-003): a log stamped
// newer than the binary is refused with the upgrade hint — the world.go
// FormatVersion posture, never a mis-replay.
func TestNewerLogRefusedWithUpgradePosture(t *testing.T) {
	s := openTemp(t)
	if err := s.SetMeta(logFormatKey, strconv.Itoa(LogFormatVersion+1)); err != nil {
		t.Fatal(err)
	}
	appendOne(t, s)
	err := s.VerifyLogFormat()
	var tooNew *ErrLogFormatTooNew
	if !errors.As(err, &tooNew) {
		t.Fatalf("VerifyLogFormat() = %v, want ErrLogFormatTooNew", err)
	}
	if !strings.Contains(err.Error(), "upgrade promptworld") {
		t.Fatalf("newer-log refusal must carry the upgrade posture, got: %v", err)
	}
	if !IsLogFormatMismatch(err) {
		t.Error("IsLogFormatMismatch must match the too-new refusal")
	}
}

// TestVerifyAdoptsEmptyUnstampedLog: an unstamped EMPTY log (mid-genesis) is
// adopted — stamped current and admitted, since nothing was ever recorded
// under another vocabulary.
func TestVerifyAdoptsEmptyUnstampedLog(t *testing.T) {
	s := openTemp(t)
	if err := s.VerifyLogFormat(); err != nil {
		t.Fatalf("empty unstamped log refused: %v", err)
	}
	v, stamped, err := s.LogFormat()
	if err != nil {
		t.Fatal(err)
	}
	if !stamped || v != LogFormatVersion {
		t.Fatalf("adoption did not stamp: LogFormat() = (%d, %v)", v, stamped)
	}
}

// TestCorruptStampRefused: a non-numeric stamp is a corrupt log, refused
// rather than guessed at.
func TestCorruptStampRefused(t *testing.T) {
	s := openTemp(t)
	if err := s.SetMeta(logFormatKey, "banana"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.LogFormat(); err == nil {
		t.Fatal("corrupt stamp accepted")
	}
	if err := s.VerifyLogFormat(); err == nil {
		t.Fatal("corrupt stamp passed verification")
	}
}
