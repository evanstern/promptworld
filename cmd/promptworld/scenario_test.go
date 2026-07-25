package main

// Spec 054 US3 T007: `promptworld new --scenario <id>` — the one-command
// scenario world: exercise-implied stage + pinned seed + scenario block,
// unknown-id refusal listing the catalog, and the earned-stage gate applied
// to the implied stage unchanged. Plus the status line's scenario rendering
// (T009's CLI surface).

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worlds"
)

// TestCmdNewScenarioStampsManifest (US3 AS-1): a valid scenario id stamps
// the block, the exercise's stage, its pinned seed, and the stage's charter
// preset — ready for the daemon to arm at boot.
func TestCmdNewScenarioStampsManifest(t *testing.T) {
	home := isolatedHome(t)
	if err := cmdNew([]string{"fn", "--scenario", "first-night"}); err != nil {
		t.Fatal(err)
	}
	m := openManifest(t, home, "fn")
	if m.Scenario == nil || m.Scenario.Exercise != "first-night" {
		t.Errorf("Scenario = %+v, want first-night", m.Scenario)
	}
	if m.Stage != world.Stage1 {
		t.Errorf("stage = %q, want the exercise's %q", m.Stage, world.Stage1)
	}
	if m.Seed != sim.FirstNightExercise.Seed {
		t.Errorf("seed = %d, want the exercise's pinned %d", m.Seed, sim.FirstNightExercise.Seed)
	}
	if m.CharterPreset != world.CharterPresetTutor {
		t.Errorf("charter_preset = %q, want the stage-1 tutor default", m.CharterPreset)
	}
	if m.StageOverridden {
		t.Error("an earned-stage scenario creation must not record an override")
	}
}

// TestCmdNewScenarioUnknownRefusesWithCatalog (US3 AS-2): an unknown id
// refuses, listing every shipped exercise — the stage-gate refusal voice.
func TestCmdNewScenarioUnknownRefusesWithCatalog(t *testing.T) {
	isolatedHome(t)
	err := cmdNew([]string{"fn", "--scenario", "bogus"})
	if err == nil {
		t.Fatal("expected a refusal for an unknown scenario id")
	}
	msg := err.Error()
	for _, def := range sim.ScenarioExercises {
		if !strings.Contains(msg, def.ID) {
			t.Errorf("refusal should list the catalog entry %q: %q", def.ID, msg)
		}
	}
}

// TestCmdNewScenarioConflictingFlags: the scenario implies stage and pins
// seed — explicit flags may only agree.
func TestCmdNewScenarioConflictingFlags(t *testing.T) {
	isolatedHome(t)
	if err := cmdNew([]string{"fn", "--scenario", "first-night", "--stage", "stage-2"}); err == nil ||
		!strings.Contains(err.Error(), "--stage") {
		t.Errorf("conflicting --stage should refuse naming the flag, got %v", err)
	}
	if err := cmdNew([]string{"fn", "--scenario", "first-night", "--seed", "5"}); err == nil ||
		!strings.Contains(err.Error(), "--seed") {
		t.Errorf("conflicting --seed should refuse naming the flag, got %v", err)
	}
	// Agreeing flags pass (the pinned seed and the implied stage verbatim).
	if err := cmdNew([]string{"fn", "--scenario", "first-night", "--stage", "stage-1", "--seed", "46101"}); err != nil {
		t.Errorf("agreeing flags should create: %v", err)
	}
}

// TestCmdNewScenarioEarnGateApplies (US3 AS-3): a scenario whose stage the
// player hasn't earned hits the existing gate unchanged — implied stage,
// same refusal, same --override escape recorded honestly.
func TestCmdNewScenarioEarnGateApplies(t *testing.T) {
	home := isolatedHome(t)
	err := cmdNew([]string{"law", "--scenario", "the-law"})
	if err == nil || !strings.Contains(err.Error(), "--override") {
		t.Fatalf("unearned stage-2 scenario should refuse with the override hint, got %v", err)
	}
	if err := cmdNew([]string{"law", "--scenario", "the-law", "--override"}); err != nil {
		t.Fatal(err)
	}
	m := openManifest(t, home, "law")
	if m.Scenario == nil || m.Scenario.Exercise != "the-law" || m.Stage != world.Stage2 || !m.StageOverridden {
		t.Errorf("override creation manifest = %+v", m)
	}

	// An earned stage-2 needs no override.
	worlds.UpsertUnlock("stage-2", worlds.UnlockEntry{World: "proof", Path: "/x", Exercise: "first-night", EarnedAt: "2026-07-25T00:00:00Z"})
	if err := cmdNew([]string{"law2", "--scenario", "the-law"}); err != nil {
		t.Errorf("earned-stage scenario should create without --override: %v", err)
	}
}

// TestScenarioStatusLine (FR-007's human surface): the exercise + outcome
// line, absent for ambient worlds so their output is unchanged.
func TestScenarioStatusLine(t *testing.T) {
	ambient := &ipc.StatusData{}
	if got := scenarioStatusLine(ambient); got != "" {
		t.Errorf("ambient status line = %q, want empty", got)
	}
	live := &ipc.StatusData{World: ipc.WorldStatus{ScenarioExercise: "first-night", ScenarioOutcome: "in_progress"}}
	if got := scenarioStatusLine(live); got != "exercise: first-night — in progress" {
		t.Errorf("in-progress line = %q", got)
	}
	failed := &ipc.StatusData{World: ipc.WorldStatus{ScenarioExercise: "first-night", ScenarioOutcome: "failed"}}
	if got := scenarioStatusLine(failed); got != "exercise: first-night — failed (run ended)" {
		t.Errorf("failed line = %q", got)
	}
	if !strings.Contains(renderStatusHuman(live), "exercise: first-night — in progress") {
		t.Error("renderStatusHuman should carry the scenario line")
	}
	if strings.Contains(renderStatusHuman(ambient), "exercise:") {
		t.Error("ambient renderStatusHuman must not mention an exercise")
	}
}
