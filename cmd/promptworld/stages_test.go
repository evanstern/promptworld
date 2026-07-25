package main

// Spec 046 US1 T011: creation flows (earned/unearned/override/default),
// manifest immutability, status rendering, absent-stage worlds unchanged;
// plus the `promptworld stages` command's human/--json surfaces (T008).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worlds"
)

func openManifest(t *testing.T, home, name string) world.Manifest {
	t.Helper()
	whome, err := worlds.WorldsHome()
	if err != nil {
		t.Fatal(err)
	}
	w, err := world.Open(filepath.Join(whome, name))
	if err != nil {
		t.Fatalf("Open %s: %v", name, err)
	}
	return w.Manifest
}

// TestCmdNewDefaultsToStage1ForNewPlayer (R9): with no --stage and no
// unlocks record, `new` defaults to stage-1 and seeds the tutor preset.
func TestCmdNewDefaultsToStage1ForNewPlayer(t *testing.T) {
	home := isolatedHome(t)
	if err := cmdNew([]string{"aria", "--seed", "1"}); err != nil {
		t.Fatal(err)
	}
	m := openManifest(t, home, "aria")
	if m.Stage != world.Stage1 {
		t.Errorf("default stage = %q, want %q", m.Stage, world.Stage1)
	}
	if m.StageOverridden {
		t.Error("a default-stage creation must not record an override")
	}
	if m.CharterPreset != world.CharterPresetTutor {
		t.Errorf("stage-1 default charter_preset = %q, want %q", m.CharterPreset, world.CharterPresetTutor)
	}
	charter, err := worlds.WorldsHome()
	if err != nil {
		t.Fatal(err)
	}
	data := readFile(t, filepath.Join(charter, "aria", "charter.md"))
	if data != persona.TutorCharter {
		t.Error("stage-1 default should seed persona.TutorCharter into charter.md")
	}
}

// TestCmdNewDefaultsToHighestEarnedStage (R9): with an unlocks record
// showing stage-3 earned, a bare `new` (no --stage) creates at stage-3 with
// no override recorded.
func TestCmdNewDefaultsToHighestEarnedStage(t *testing.T) {
	home := isolatedHome(t)
	worlds.UpsertUnlock("stage-2", worlds.UnlockEntry{World: "proof1", Path: "/x", Exercise: "first-night", EarnedAt: "2026-07-25T00:00:00Z"})
	worlds.UpsertUnlock("stage-3", worlds.UnlockEntry{World: "proof2", Path: "/y", Exercise: "the-law", EarnedAt: "2026-07-25T00:00:00Z"})

	if err := cmdNew([]string{"aria", "--seed", "1"}); err != nil {
		t.Fatal(err)
	}
	m := openManifest(t, home, "aria")
	if m.Stage != world.Stage3 {
		t.Errorf("default stage = %q, want %q (highest earned)", m.Stage, world.Stage3)
	}
	if m.StageOverridden {
		t.Error("creating at an EARNED stage must not record an override")
	}
	// Above stage-1, no preset is stamped by default — "" and "default" are
	// equivalent (ValidCharterPreset), and leaving it empty keeps the
	// manifest lean (only stage-1's tutor default, or an explicit
	// --charter-preset, ever stamps a non-empty value).
	if m.CharterPreset != "" {
		t.Errorf("stage-3 default charter_preset = %q, want empty", m.CharterPreset)
	}
}

// TestCmdNewUnearnedStageRequiresOverride (FR-003): requesting an unearned
// stage without --override refuses with an informed message naming the
// skipped concepts by their skin display names.
func TestCmdNewUnearnedStageRequiresOverride(t *testing.T) {
	isolatedHome(t)
	err := cmdNew([]string{"aria", "--seed", "1", "--stage", "stage-3"})
	if err == nil {
		t.Fatal("expected an error requesting an unearned stage without --override")
	}
	msg := err.Error()
	for _, want := range []string{"The Written Word", "The Craft", "--override"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message missing %q: %q", want, msg)
		}
	}
}

// TestCmdNewOverrideProceedsAndRecords (FR-003): --override proceeds at an
// unearned stage and the manifest honestly records the override.
func TestCmdNewOverrideProceedsAndRecords(t *testing.T) {
	home := isolatedHome(t)
	if err := cmdNew([]string{"aria", "--seed", "1", "--stage", "stage-3", "--override"}); err != nil {
		t.Fatal(err)
	}
	m := openManifest(t, home, "aria")
	if m.Stage != world.Stage3 {
		t.Errorf("stage = %q, want stage-3", m.Stage)
	}
	if !m.StageOverridden {
		t.Error("override must be recorded in the manifest")
	}
}

// TestCmdNewEarnedStageNeedsNoOverride: a stage the player HAS earned is
// offered normally — no --override required.
func TestCmdNewEarnedStageNeedsNoOverride(t *testing.T) {
	home := isolatedHome(t)
	worlds.UpsertUnlock("stage-2", worlds.UnlockEntry{World: "proof", Path: "/x", Exercise: "first-night", EarnedAt: "2026-07-25T00:00:00Z"})
	if err := cmdNew([]string{"aria", "--seed", "1", "--stage", "stage-2"}); err != nil {
		t.Fatalf("earned stage should not require --override: %v", err)
	}
	m := openManifest(t, home, "aria")
	if m.Stage != world.Stage2 || m.StageOverridden {
		t.Errorf("manifest = stage %q overridden %v, want stage-2 false", m.Stage, m.StageOverridden)
	}
}

// TestCmdNewInvalidStageRejected: a garbage --stage value is refused before
// anything is created.
func TestCmdNewInvalidStageRejected(t *testing.T) {
	isolatedHome(t)
	if err := cmdNew([]string{"aria", "--seed", "1", "--stage", "stage-99"}); err == nil {
		t.Error("expected an error for an invalid --stage value")
	}
}

// TestCmdNewCharterPresetOptOut (R6): --charter-preset default at stage-1
// opts OUT of the tutor default.
func TestCmdNewCharterPresetOptOut(t *testing.T) {
	home := isolatedHome(t)
	if err := cmdNew([]string{"aria", "--seed", "1", "--stage", "stage-1", "--charter-preset", "default"}); err != nil {
		t.Fatal(err)
	}
	m := openManifest(t, home, "aria")
	if m.CharterPreset != world.CharterPresetDefault {
		t.Errorf("charter_preset = %q, want %q (opt-out)", m.CharterPreset, world.CharterPresetDefault)
	}
	whome, _ := worlds.WorldsHome()
	data := readFile(t, filepath.Join(whome, "aria", "charter.md"))
	if data != persona.DefaultCharter {
		t.Error("opting out should seed persona.DefaultCharter, not the tutor preset")
	}
}

// TestCmdNewInvalidCharterPresetRejected: a garbage --charter-preset value
// is refused.
func TestCmdNewInvalidCharterPresetRejected(t *testing.T) {
	isolatedHome(t)
	if err := cmdNew([]string{"aria", "--seed", "1", "--charter-preset", "nonsense"}); err == nil {
		t.Error("expected an error for an invalid --charter-preset value")
	}
}

// TestStageStatusLineAbsentForPreLadderWorld: a world with no Stage set (the
// world.Create primitive directly, bypassing `new`'s stage stamping —
// simulating a pre-046 world) renders no stage line at all (byte-compat).
func TestStageStatusLineAbsentForPreLadderWorld(t *testing.T) {
	if line := stageStatusLine("", false); line != "" {
		t.Errorf("absent-stage line = %q, want empty", line)
	}
}

// TestStageStatusLineRendersNameAndOverride: the stage line names the skin
// display identity and flags an override.
func TestStageStatusLineRendersNameAndOverride(t *testing.T) {
	if got, want := stageStatusLine("stage-1", false), "stage: The Voice (stage-1)"; got != want {
		t.Errorf("stage line = %q, want %q", got, want)
	}
	if got := stageStatusLine("stage-3", true); !strings.Contains(got, "[overridden]") {
		t.Errorf("overridden stage line missing marker: %q", got)
	}
}

// TestCmdStagesHumanOutput: the human-readable `stages` listing names every
// stage's identity, concept, grants, unlock evidence, and earned state.
func TestCmdStagesHumanOutput(t *testing.T) {
	isolatedHome(t)
	worlds.UpsertUnlock("stage-2", worlds.UnlockEntry{World: "proofworld", Path: "/x", Exercise: "first-night", EarnedAt: "2026-07-25T00:00:00Z"})
	out := captureStdout(t, func() {
		if err := cmdStages(nil); err != nil {
			t.Fatal(err)
		}
	})
	for _, want := range []string{
		"The Voice", "The Written Word", "The Craft", "The Stewardship",
		"teaches:", "grants:", "unlocked by:",
		"proofworld", "first-night",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stages human output missing %q:\n%s", want, out)
		}
	}
}

// TestCmdStagesJSONOutput: the --json twin carries the same facts as
// structured rows, earned state included.
func TestCmdStagesJSONOutput(t *testing.T) {
	isolatedHome(t)
	out := captureStdout(t, func() {
		if err := cmdStages([]string{"--json"}); err != nil {
			t.Fatal(err)
		}
	})
	for _, want := range []string{`"id": "stage-1"`, `"earned": true`, `"id": "stage-4"`, `"earned": false`} {
		if !strings.Contains(out, want) {
			t.Errorf("stages --json output missing %q:\n%s", want, out)
		}
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
