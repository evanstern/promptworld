package guardian

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/toolloop"
)

// --- spec 102 T003: the deliberate-incompetence ceiling (D3, FR-003) ---

// captureJob installs a scripted loop that records the Job the turn composed
// (roster, handlers, system prompt, kind) and converses quietly.
func captureJob(mt *Guardian, sink *toolloop.Job) {
	mt.runLoop = func(_ context.Context, j toolloop.Job) (toolloop.Result, error) {
		*sink = j
		return toolloop.Result{Final: "so noted", Term: toolloop.TermModelDone}, nil
	}
}

func rosterNames(j toolloop.Job) []string {
	var out []string
	for _, t := range j.Roster {
		out = append(out, t.Name)
	}
	sort.Strings(out)
	return out
}

// runAngelTurn drives one scheduled turn synchronously through the real
// schedule path (opt-in → due → queue → runAngel).
func runAngelTurn(t *testing.T, mt *Guardian, inj *stateInjector) {
	t.Helper()
	optIn(t, mt, inj, 600)
	advanceTo(mt, 1000)
	mt.scheduleAngel()
	advanceTo(mt, 1700)
	mt.scheduleAngel()
	if !drainAngel(t, mt) {
		t.Fatal("cadence queued no turn")
	}
}

// TestDefaultCharterCeilingCapsScheduledRoster: under the compiled DEFAULT
// charter, a scheduled turn's whole surface is the modest read/counsel set —
// no charge spend, no watches, no plans, no clock — enforced structurally in
// all three gating layers (declaration, handlers, prose), and the turn rides
// the angel kind.
func TestDefaultCharterCeilingCapsScheduledRoster(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "unused")
	var j toolloop.Job
	captureJob(mt, &j)
	runAngelTurn(t, mt, inj)

	want := []string{"brief_myths", "explain", "survey_site"}
	if got := rosterNames(j); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("default-ceiling scheduled roster = %v, want %v", got, want)
	}
	for _, name := range []string{"send_vision", "send_omen", "work_miracle", "monitor_and_act",
		"cancel_order", "place_designation", "issue_directive", "prophesy", "canonize_region",
		"pause", "start", "adjust_speed"} {
		if _, ok := j.Handlers[name]; ok {
			t.Errorf("ceiling-capped tool %q still has a handler (door layer leak)", name)
		}
	}
	if !strings.Contains(j.System, "no initiative beyond watchfulness") {
		t.Fatal("modest frame missing from the scheduled system prompt")
	}
	// The ceiling caps INITIATIVE only — the frame says so, verbatim doctrine.
	if !strings.Contains(j.System, "you act with your full skill") {
		t.Fatal("modest frame lost the compliance-is-full sentence")
	}
	if j.Kind != llm.KindAngel {
		t.Fatalf("scheduled turn kind = %q, want %q", j.Kind, llm.KindAngel)
	}
}

// TestAuthoredCharterLiftsCeiling: a player-authored charter lifts the
// ceiling — the scheduled roster regains the acting surface (spending,
// watches, plans) while the clock triple stays structurally absent at ANY
// ceiling, and the lifted frame composes.
func TestAuthoredCharterLiftsCeiling(t *testing.T) {
	mt, _, inj, dir := newTestGuardian(t, "unused")
	authored := "You are a bold shepherd. Act early and often to keep the village fed and warm."
	if err := os.WriteFile(filepath.Join(dir, "charter.md"), []byte(authored), 0o644); err != nil {
		t.Fatal(err)
	}
	mt.charterFP = "" // let the new revision observe cleanly
	var j toolloop.Job
	captureJob(mt, &j)
	runAngelTurn(t, mt, inj)

	got := map[string]bool{}
	for _, n := range rosterNames(j) {
		got[n] = true
	}
	for _, name := range []string{"send_vision", "send_omen", "work_miracle", "monitor_and_act", "explain", "survey_site"} {
		if !got[name] {
			t.Errorf("lifted ceiling missing %q from the scheduled roster", name)
		}
	}
	for _, name := range angelClockTools {
		if got[name] {
			t.Errorf("clock tool %q on the scheduled roster — the clock is the player's at any ceiling", name)
		}
		if _, ok := j.Handlers[name]; ok {
			t.Errorf("clock tool %q has a scheduled handler", name)
		}
	}
	if !strings.Contains(j.System, "Your charter grants you initiative") {
		t.Fatal("lifted frame missing from the scheduled system prompt")
	}
}

// TestConsoleTurnFullCompetenceAtAnyCeiling pins the operator ruling's scope:
// the ceiling caps INITIATIVE only — a DEFAULT-charter world's CONSOLE turn
// keeps the full granted roster (clock included) and the ordinary restrictive
// frame; order compliance and tutor quality are never capped.
func TestConsoleTurnFullCompetenceAtAnyCeiling(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "at your service")
	optIn(t, mt, inj, 600) // agentized world, default charter — ceiling ON for the lane
	var j toolloop.Job
	captureJob(mt, &j)
	if _, err := mt.Turn(t.Context(), "send Ash a vision"); err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, n := range rosterNames(j) {
		got[n] = true
	}
	for _, name := range []string{"send_vision", "work_miracle", "monitor_and_act", "pause", "start", "adjust_speed", "explain"} {
		if !got[name] {
			t.Errorf("console roster missing %q — the ceiling must never cap compliance", name)
		}
	}
	if strings.Contains(j.System, "no initiative beyond watchfulness") ||
		strings.Contains(j.System, "Your charter grants you initiative") {
		t.Fatal("an angel frame leaked into a console turn")
	}
	if j.Kind != llm.KindGuardian {
		t.Fatalf("console turn kind = %q, want %q", j.Kind, llm.KindGuardian)
	}
}

// TestCeilingCompilesFromEffectiveCharter pins the data-not-degradation
// contract at the compile seam itself: preset text and both retired legacy
// seeds keep the ceiling ON; any other byte sequence lifts it.
func TestCeilingCompilesFromEffectiveCharter(t *testing.T) {
	if angelCharterLifted(presetCharter(""), "") {
		t.Fatal("default preset text lifted the ceiling")
	}
	if angelCharterLifted(presetCharter("tutor"), "tutor") {
		t.Fatal("tutor preset text lifted the ceiling")
	}
	if !angelCharterLifted("Anything the player wrote.", "") {
		t.Fatal("authored text did not lift the ceiling")
	}
}
