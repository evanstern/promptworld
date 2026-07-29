package tui

// Spec 054 US4 (T010–T012): the exercise dock tab — presence only on
// scenario worlds (key 6, inert elsewhere), the panel grammar (contract §4),
// the attach-once briefing with its one-keypress any-key dismiss, live
// gauges over the replica, forecast/fog visibility, and the pass/fail
// banner.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
)

func mustPayload(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// scenarioModel is testModel on a first-night scenario world: the manifest
// carries the block (what world.Open would load), the replica runs the
// exercise's own seed.
func scenarioModel(t *testing.T) Model {
	t.Helper()
	w, err := world.Create(t.TempDir()+"/w", "fn", sim.FirstNightExercise.Seed)
	if err != nil {
		t.Fatal(err)
	}
	w.Manifest.Scenario = &world.ScenarioConfig{Exercise: "first-night"}
	w.Manifest.Stage = world.Stage1
	m := New(w)
	m.replica = sim.NewState(w.Manifest.Seed, w.Map())
	m.width, m.height = 140, 40
	return m
}

func pressKey(t *testing.T, m Model, k string) Model {
	t.Helper()
	mdl, _ := m.Update(key(k))
	return mdl.(Model)
}

// TestExerciseTabAbsentOnAmbientWorlds (US4 AS-6, contract §4): no tab in
// either layout's row, and 6 falls through inert.
func TestExerciseTabAbsentOnAmbientWorlds(t *testing.T) {
	m := widescreenModel(t)
	if strings.Contains(m.dockTabsRow(), "exercise") {
		t.Error("ambient dock row must not carry an exercise tab")
	}
	m2 := pressKey(t, m, "6")
	if m2.dockTab != m.dockTab || m2.solo {
		t.Error("6 must be inert on an ambient world")
	}
	narrow := testModel(t)
	if strings.Contains(narrow.tabsView(), "exercise") {
		t.Error("ambient narrow tab row must not carry an exercise tab")
	}
	if got := pressKey(t, narrow, "6"); got.active != narrow.active {
		t.Error("6 must be inert in the ambient narrow fallback")
	}
	if strings.Contains(m.footerView(), "6 exercise") {
		t.Error("ambient footer must not advertise key 6")
	}
}

// TestExerciseTabGrammar (contract §4 row 1): 6 selects; again solos; again
// returns home — the standard solo-views state machine; narrow reaches the
// pane like any other tab.
func TestExerciseTabGrammar(t *testing.T) {
	m := scenarioModel(t)
	m.exBriefingDismissed = true // grammar test: past the briefing
	if !strings.Contains(m.dockTabsRow(), "exercise") {
		t.Fatal("scenario dock row missing the exercise tab")
	}
	m = pressKey(t, m, "6")
	if m.dockTab != paneExercise || m.solo {
		t.Fatalf("6 should select the exercise tab (dockTab=%d solo=%v)", m.dockTab, m.solo)
	}
	m = pressKey(t, m, "6")
	if !m.solo {
		t.Fatal("6 again should solo the exercise tab")
	}
	m = pressKey(t, m, "6")
	if m.solo {
		t.Fatal("6 a third time should return home")
	}
	if !strings.Contains(m.footerView(), "6 exercise") {
		t.Error("scenario footer should advertise key 6")
	}

	narrow := scenarioModel(t)
	narrow.width = 80
	narrow.exBriefingDismissed = true
	narrow = pressKey(t, narrow, "6")
	if narrow.active != paneExercise {
		t.Fatal("narrow 6 should switch the active pane to exercise")
	}
	if !strings.Contains(narrow.View(), "FIRST NIGHT") {
		t.Error("narrow exercise pane should render the panel")
	}
}

// TestExerciseBriefingOncePerAttach (US4 AS-1, contract §4 row 2): first
// render shows framing + visibility mode; ANY key dismisses it (consumed —
// even q must not quit); reconnect shows it again.
func TestExerciseBriefingOncePerAttach(t *testing.T) {
	m := scenarioModel(t)
	m = pressKey(t, m, "6") // select the tab: the briefing is now visible
	body := m.exerciseBody(80, 20)
	for _, want := range []string{"FIRST NIGHT", "seeded world", "forecast", "any key — begin"} {
		if !strings.Contains(body, want) {
			t.Errorf("briefing missing %q:\n%s", want, body)
		}
	}

	// Any key — even q — dismisses and is consumed (never quits).
	m = pressKey(t, m, "q")
	if m.quitting {
		t.Fatal("the briefing's any-key dismiss must consume the key, not quit")
	}
	if !m.exBriefingDismissed {
		t.Fatal("briefing not dismissed")
	}
	if body := m.exerciseBody(80, 20); strings.Contains(body, "any key — begin") {
		t.Error("dismissed briefing still rendering")
	}
	// Now keys act normally again: q quits.
	if got := pressKey(t, m, "q"); !got.quitting {
		t.Error("after dismissal, q should quit as usual")
	}

	// The eater is scoped: with another tab visible, keys pass through.
	fresh := scenarioModel(t)
	fresh = pressKey(t, fresh, "4")
	if fresh.dockTab != paneVillagers {
		t.Fatal("keys must not be eaten while the exercise tab is not visible")
	}

	// Reconnect (a fresh attach) shows the briefing again.
	m2, _ := m.Update(connectedMsg{replica: m.replica, lastSeq: 0})
	if m2.(Model).exBriefingDismissed {
		t.Error("reconnect must reset the briefing (once per attach)")
	}
}

// TestExerciseBriefingYieldsToConsole (spec 053 × spec 054 interaction): the
// guardian console is a whole-body takeover, so while it is open the
// briefing is NOT the thing on screen — its any-key eater must yield, keys
// reach the console (esc closes it), and the briefing survives undismissed
// for when the tab is actually visible again.
func TestExerciseBriefingYieldsToConsole(t *testing.T) {
	m := scenarioModel(t)
	m = pressKey(t, m, "6") // exercise tab selected, briefing pending
	mdl, _ := m.openConsole()
	m = mdl.(Model)
	if !m.console {
		t.Fatal("setup: console did not open")
	}
	m = pressKey(t, m, "esc") // must reach the console (close), not the eater
	if m.console {
		t.Fatal("esc was eaten by the briefing instead of closing the console")
	}
	if m.exBriefingDismissed {
		t.Fatal("the briefing must survive a console session undismissed")
	}
	// Back on the exercise tab, the next key dismisses as usual.
	m = pressKey(t, m, "j")
	if !m.exBriefingDismissed {
		t.Fatal("briefing should dismiss normally once the console is closed")
	}
}

// TestExerciseGaugesTrackReplica (US4 AS-2): rubric-relevant events folded
// into the replica flip the gauges — same data, no extra IPC.
func TestExerciseGaugesTrackReplica(t *testing.T) {
	m := scenarioModel(t)
	m.exBriefingDismissed = true
	body := m.exerciseBody(100, 20)
	if !strings.Contains(body, "… a watch set before nightfall") ||
		!strings.Contains(body, "(guardian.order_placed: 0)") {
		t.Errorf("pending watch gauge missing:\n%s", body)
	}
	if !strings.Contains(body, "✓") || !strings.Contains(body, "no villager dies") ||
		!strings.Contains(body, "(agent.died: 0)") {
		t.Errorf("zero-deaths gauge missing:\n%s", body)
	}

	// A watch order lands (the replica fold applyEvent runs on push).
	order := sim.GuardianOrder{
		ID: "ord-100-0", Origin: "player",
		Condition: "if the gru nears", Action: "wake everyone",
		EventTypes: []string{"gru.sighted"}, Agent: -1,
		PlacedTick: 100, ExpiresTick: 100 + 2*86400,
	}
	m.applyEvent(store.Event{Seq: 1, Tick: 100, Type: "guardian.order_placed", Payload: mustPayload(t, order)})
	body = m.exerciseBody(100, 20)
	if strings.Contains(body, "… a watch set before nightfall") ||
		!strings.Contains(body, "(guardian.order_placed: 1)") {
		t.Errorf("watch gauge did not flip met:\n%s", body)
	}
}

// theLawModel is scenarioModel's the-law twin: a stage-2 world seeded with
// the second cataloged exercise.
func theLawModel(t *testing.T) Model {
	t.Helper()
	w, err := world.Create(t.TempDir()+"/w", "law", sim.TheLawExercise.Seed)
	if err != nil {
		t.Fatal(err)
	}
	w.Manifest.Scenario = &world.ScenarioConfig{Exercise: "the-law"}
	w.Manifest.Stage = world.Stage2
	m := New(w)
	m.replica = sim.NewState(w.Manifest.Seed, w.Map())
	m.width, m.height = 140, 40
	m.exBriefingDismissed = true
	return m
}

// TestExerciseGaugesTheLawEvaluateForReal (spec 072 US2, SC-002): on a
// the-law world the gauges evaluate from state facts — no term renders
// permanently pending. The charter gauge flips ONLY on a custom (player-
// authored) charter observation, never the default's (US2-2); the law gauge
// flips when a passed proposal appends a norm. Zero code change in this tab
// — the truth arrives through sim.EvaluateRubric's new production arm.
func TestExerciseGaugesTheLawEvaluateForReal(t *testing.T) {
	m := theLawModel(t)
	body := m.exerciseBody(120, 20)
	if !strings.Contains(body, "… a village law adopted") ||
		!strings.Contains(body, "(meeting.proposal_resolved: 0)") ||
		!strings.Contains(body, "… a player-authored charter in force") ||
		!strings.Contains(body, "(guardian.charter_observed: 0)") {
		t.Errorf("genesis gauges should be pending with zero counts:\n%s", body)
	}

	// The game's own (default) charter is observed: authorship stays pending
	// — the game's authorship never satisfies the player-authorship term.
	m.applyEvent(store.Event{Seq: 1, Tick: 100, Type: "guardian.charter_observed",
		Payload: mustPayload(t, sim.CharterObservedPayload{Fingerprint: "aaaa11112222", Default: true})})
	body = m.exerciseBody(120, 20)
	if !strings.Contains(body, "… a player-authored charter in force") ||
		!strings.Contains(body, "(guardian.charter_observed: 1)") {
		t.Errorf("default charter must leave the authorship gauge pending (count 1):\n%s", body)
	}

	// A player-authored revision lands: the gauge flips met.
	m.applyEvent(store.Event{Seq: 2, Tick: 200, Type: "guardian.charter_observed",
		Payload: mustPayload(t, sim.CharterObservedPayload{Fingerprint: "bbbb33334444", Default: false})})
	body = m.exerciseBody(120, 20)
	if strings.Contains(body, "… a player-authored charter in force") ||
		!strings.Contains(body, "a player-authored charter in force") {
		t.Errorf("custom charter observation should flip the authorship gauge met:\n%s", body)
	}

	// A passed proposal appends a norm: the law gauge flips with the count.
	m.applyEvent(store.Event{Seq: 3, Tick: 300, Type: "meeting.proposal_resolved",
		Payload: mustPayload(t, sim.ProposalResolvedPayload{
			ProposalPayload: sim.ProposalPayload{ProposalID: 1, Kind: sim.ProposeCurfew,
				Target: sim.Ref(-1), Proposer: sim.Ref(0), Text: "no one leaves the fires after dusk"},
			Passed: true,
		})})
	body = m.exerciseBody(120, 20)
	if strings.Contains(body, "… a village law adopted") ||
		!strings.Contains(body, "(meeting.proposal_resolved: 1)") {
		t.Errorf("an adopted norm should flip the law gauge with its ledger count:\n%s", body)
	}
}

// TestExerciseForecastVsFog (US4 AS-3, D4): the incident line forecasts the
// authored schedule at stages 1–2/pre-ladder and is omitted under fog
// (stage 3+); a definition override wins over the stage default.
func TestExerciseForecastVsFog(t *testing.T) {
	m := scenarioModel(t)
	m.exBriefingDismissed = true
	if body := m.exerciseBody(100, 20); !strings.Contains(body, "incidents (forecast): the gru emerges ~22:00 (day 1)") {
		t.Errorf("stage-1 panel missing the forecast line:\n%s", body)
	}
	m.w.Manifest.Stage = world.Stage3
	if body := m.exerciseBody(100, 20); strings.Contains(body, "incidents") {
		t.Errorf("stage-3 panel must omit the incident line (fog):\n%s", body)
	}
	// Vocabulary resolution, override included, is pinned in
	// sim.TestIncidentVisibilityVocabulary; here only the render seam.
	if line := exerciseIncidentLine(sim.FirstNightExercise, ""); line == "" {
		t.Error("pre-ladder worlds forecast (everything)")
	}
}

// TestExerciseBanner (US4 AS-4/AS-5): the pass banner on a recorded pass;
// the failed banner once the run ends with none.
func TestExerciseBanner(t *testing.T) {
	m := scenarioModel(t)
	m.exBriefingDismissed = true
	if body := m.exerciseBody(100, 20); !strings.Contains(body, "in progress") {
		t.Errorf("default posture should be in progress:\n%s", body)
	}

	passed := scenarioModel(t)
	passed.exBriefingDismissed = true
	passed.replica.CurriculumPasses = append(passed.replica.CurriculumPasses,
		sim.CurriculumPass{Exercise: "first-night", Stage: "stage-1", Tick: 86400})
	if body := passed.exerciseBody(100, 20); !strings.Contains(body, "PASSED") || !strings.Contains(body, "· passed") {
		t.Errorf("pass banner missing:\n%s", body)
	}

	failed := scenarioModel(t)
	failed.exBriefingDismissed = true
	failed.replica.Ended = true
	if body := failed.exerciseBody(100, 20); !strings.Contains(body, "FAILED (run ended)") ||
		!strings.Contains(body, "failed (run ended)") {
		t.Errorf("failed banner missing:\n%s", body)
	}
}

// TestExerciseDockCycleIncludesTab: tab/shift+tab cycle through the
// exercise tab — after the systems tab, matching its "6" position — exactly
// when the world carries one; ambient worlds keep the 4-tab cycle.
func TestExerciseDockCycleIncludesTab(t *testing.T) {
	m := scenarioModel(t)
	if got := m.nextDockTab(paneSystems); got != paneExercise {
		t.Errorf("scenario cycle after systems = %d, want exercise", got)
	}
	if got := m.nextDockTab(paneExercise); got != paneChronicle {
		t.Errorf("scenario cycle after exercise = %d, want chronicle", got)
	}
	if got := m.prevDockTab(paneChronicle); got != paneExercise {
		t.Errorf("scenario reverse cycle before chronicle = %d, want exercise", got)
	}
	if got := m.prevDockTab(paneExercise); got != paneSystems {
		t.Errorf("scenario reverse cycle before exercise = %d, want systems", got)
	}
	ambient := widescreenModel(t)
	if got := ambient.nextDockTab(paneSystems); got != paneChronicle {
		t.Errorf("ambient cycle after systems = %d, want chronicle", got)
	}
	if got := ambient.prevDockTab(paneChronicle); got != paneSystems {
		t.Errorf("ambient reverse cycle before chronicle = %d, want systems", got)
	}
}
