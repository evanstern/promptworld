package tui

// Takeover-family tests (spec 056, contracts/takeovers.md): the overlay-owner
// state machine (T002) — precedence interleavings (SC-003), dismiss/replay
// wiring (esc/p/q), and connect-time auto-open. Renderer/ambient-scored-
// matrix/exact-height tests live in render_test.go; the console-seam
// composition test lives in console_test.go; the `?` overlay replay section
// test lives in help_test.go.

import (
	"encoding/json"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

func runEndedEvent(seq, tick int64, finalCause string, deaths []sim.DeathRecord) store.Event {
	b, _ := json.Marshal(sim.RunEndedPayload{Tick: tick, Deaths: deaths, FinalCause: finalCause})
	return store.Event{Seq: seq, Tick: tick, Type: "run.ended", Payload: b}
}

func stageUnlockedEvent(seq, tick int64, stage, exercise string) store.Event {
	b, _ := json.Marshal(sim.StageUnlockedPayload{Stage: stage, Exercise: exercise, Tick: tick})
	return store.Event{Seq: seq, Tick: tick, Type: "curriculum.stage_unlocked", Payload: b}
}

// TestTakeoverCeremonyOpensOnStageUnlocked (spec.md US2-AS1): a landed
// curriculum.stage_unlocked opens the ceremony immediately, from no prior
// takeover.
func TestTakeoverCeremonyOpensOnStageUnlocked(t *testing.T) {
	m := testModel(t)
	m.applyEvent(stageUnlockedEvent(1, 100, "stage-2", "first-night"))
	if m.takeover != takeoverCeremony {
		t.Fatalf("takeover = %v, want takeoverCeremony", m.takeover)
	}
	if got := m.replica.StagesUnlocked; len(got) != 1 || got[0] != "stage-2" {
		t.Fatalf("StagesUnlocked = %v, want [stage-2]", got)
	}
}

// TestTakeoverPostmortemOpensOnRunEnded (spec.md US1-AS1): a landed
// run.ended opens the postmortem immediately.
func TestTakeoverPostmortemOpensOnRunEnded(t *testing.T) {
	m := testModel(t)
	m.applyEvent(runEndedEvent(1, 100, "exposure", []sim.DeathRecord{{Agent: 0, Tick: 100, Cause: "exposure"}}))
	if m.takeover != takeoverPostmortem {
		t.Fatalf("takeover = %v, want takeoverPostmortem", m.takeover)
	}
	if !m.runEnded() {
		t.Fatal("runEnded() should be true once run.ended has landed")
	}
}

// TestTakeoverPostmortemAlwaysWinsOverOpenCeremony (SC-003, one order):
// run.ended landing while the ceremony is open replaces it — postmortem
// always wins (contracts/takeovers.md §1).
func TestTakeoverPostmortemAlwaysWinsOverOpenCeremony(t *testing.T) {
	m := testModel(t)
	m.applyEvent(stageUnlockedEvent(1, 100, "stage-2", "first-night"))
	if m.takeover != takeoverCeremony {
		t.Fatalf("precondition: takeover = %v, want takeoverCeremony", m.takeover)
	}
	m.applyEvent(runEndedEvent(2, 200, "exposure", nil))
	if m.takeover != takeoverPostmortem {
		t.Fatalf("takeover = %v, want takeoverPostmortem (postmortem always wins)", m.takeover)
	}
}

// TestTakeoverStageUnlockedDefersWhilePostmortemOpen (SC-003, the other
// order): an unlock landing while the postmortem is open never interrupts
// it — deferred, replayable later (contracts/takeovers.md §1).
func TestTakeoverStageUnlockedDefersWhilePostmortemOpen(t *testing.T) {
	m := testModel(t)
	m.applyEvent(runEndedEvent(1, 100, "exposure", nil))
	if m.takeover != takeoverPostmortem {
		t.Fatalf("precondition: takeover = %v, want takeoverPostmortem", m.takeover)
	}
	m.applyEvent(stageUnlockedEvent(2, 200, "stage-2", "first-night"))
	if m.takeover != takeoverPostmortem {
		t.Fatalf("takeover = %v, want takeoverPostmortem (the ceremony must not interrupt it)", m.takeover)
	}
	if !m.ceremonyDeferred {
		t.Error("ceremonyDeferred should be true — the unlock arrived while the postmortem owned the slot")
	}
	// Still replayable: the fact is recorded on the replica regardless of
	// whether the live ceremony ever opened.
	if got := m.replica.StagesUnlocked; len(got) != 1 || got[0] != "stage-2" {
		t.Fatalf("StagesUnlocked = %v, want [stage-2] (deferred content stays reachable via replay)", got)
	}
}

// TestTakeoverSameKindCeremonyReplacement (spec.md Edge Cases "Multiple
// unlocks queued"): a second unlock while the ceremony is already open
// replaces it — same-kind takeovers don't stack, the newest wins.
func TestTakeoverSameKindCeremonyReplacement(t *testing.T) {
	m := testModel(t)
	m.applyEvent(stageUnlockedEvent(1, 100, "stage-2", "first-night"))
	m.applyEvent(stageUnlockedEvent(2, 200, "stage-3", "the-law"))
	if m.takeover != takeoverCeremony {
		t.Fatalf("takeover = %v, want takeoverCeremony", m.takeover)
	}
	if got := m.replica.StagesUnlocked; len(got) != 2 || got[len(got)-1] != "stage-3" {
		t.Fatalf("StagesUnlocked = %v, want [stage-2 stage-3] (both remain replayable)", got)
	}
	m.width, m.height = 140, 40
	view := m.ceremonyView(m.width, m.height-2)
	if !strings.Contains(strings.ToUpper(view), "THE CRAFT") {
		t.Errorf("ceremonyView should show the NEWEST unlock (stage-3, The Craft): %q", view)
	}
}

// TestTakeoverEscDismisses (spec.md US1-AS3/US2-AS2, contracts/takeovers.md
// §1): esc dismisses one layer everywhere; on the postmortem it also
// latches postmortemDismissed.
func TestTakeoverEscDismisses(t *testing.T) {
	t.Run("ceremony", func(t *testing.T) {
		m := testModel(t)
		m.applyEvent(stageUnlockedEvent(1, 100, "stage-2", "first-night"))
		var mdl tea.Model = m
		mdl = update(mdl, "esc")
		mm := mdl.(Model)
		if mm.takeover != takeoverNone {
			t.Fatalf("takeover = %v, want takeoverNone after esc", mm.takeover)
		}
	})
	t.Run("postmortem", func(t *testing.T) {
		m := testModel(t)
		m.applyEvent(runEndedEvent(1, 100, "exposure", nil))
		var mdl tea.Model = m
		mdl = update(mdl, "esc")
		mm := mdl.(Model)
		if mm.takeover != takeoverNone {
			t.Fatalf("takeover = %v, want takeoverNone after esc", mm.takeover)
		}
		if !mm.postmortemDismissed {
			t.Error("postmortemDismissed should be true after esc")
		}
		// Read-only clock keys stay inert after dismissal (spec 044 posture
		// persists regardless of the takeover, contracts/takeovers.md §1).
		mdl2, _ := mm.Update(tea.KeyMsg{Type: tea.KeySpace})
		mm2 := mdl2.(Model)
		if mm2.status != nil && !mm2.runEnded() {
			t.Error("runEnded() should still hold after dismissal")
		}
	})
}

// TestTakeoverQuitMessageFraming (spec.md US1-AS5/US2-AS2, View()'s quitting
// branch): the ceremony's q keeps the "world keeps running" framing (D13);
// the postmortem's q never does — the run really has ended.
func TestTakeoverQuitMessageFraming(t *testing.T) {
	t.Run("ceremony q keeps running", func(t *testing.T) {
		m := testModel(t)
		m.applyEvent(stageUnlockedEvent(1, 100, "stage-2", "first-night"))
		var mdl tea.Model = m
		mdl = update(mdl, "q")
		view := mdl.(Model).View()
		if !strings.Contains(view, "the world keeps running") {
			t.Errorf("ceremony q should keep the D13 framing: %q", view)
		}
	})
	t.Run("postmortem q is plain", func(t *testing.T) {
		m := testModel(t)
		m.applyEvent(runEndedEvent(1, 100, "exposure", nil))
		var mdl tea.Model = m
		mdl = update(mdl, "q")
		view := mdl.(Model).View()
		if strings.Contains(view, "keeps running") {
			t.Errorf("postmortem q must NOT claim the world keeps running: %q", view)
		}
		if !strings.Contains(view, "run has ended") {
			t.Errorf("postmortem q should render the honest ended framing: %q", view)
		}
	})
}

// TestTakeoverPReopensPostmortem (spec.md US1-AS3, contracts/takeovers.md
// §1): `p` reopens the dismissed postmortem from anywhere while the run is
// ended; inert on a live world.
func TestTakeoverPReopensPostmortem(t *testing.T) {
	m := testModel(t)
	m.applyEvent(runEndedEvent(1, 100, "exposure", nil))
	var mdl tea.Model = m
	mdl = update(mdl, "esc")
	if mdl.(Model).takeover != takeoverNone {
		t.Fatal("precondition: esc should have dismissed the postmortem")
	}
	mdl = update(mdl, "p")
	mm := mdl.(Model)
	if mm.takeover != takeoverPostmortem {
		t.Fatalf("takeover = %v, want takeoverPostmortem after p", mm.takeover)
	}
	if mm.postmortemDismissed {
		t.Error("postmortemDismissed should clear on p (un-dismiss)")
	}
}

// TestTakeoverPInertOnLiveWorld: `p` does nothing while the world hasn't
// ended (contracts/takeovers.md §1 "inert on live worlds").
func TestTakeoverPInertOnLiveWorld(t *testing.T) {
	m := testModel(t)
	var mdl tea.Model = m
	mdl = update(mdl, "p")
	if mdl.(Model).takeover != takeoverNone {
		t.Error("p should be inert on a live (non-ended) world")
	}
}

// TestTakeoverAutoOpenOnConnectToEndedWorld (spec.md US1-AS4): a fresh
// client attaching to an already-ended world opens the postmortem
// automatically on connect — the dual-source runEnded() posture, not only
// the live transition.
func TestTakeoverAutoOpenOnConnectToEndedWorld(t *testing.T) {
	m := testModel(t)
	ended := sim.NewState(42, m.gameMap)
	ended.Ended = true
	ended.RunEnd = &sim.RunEnd{Tick: 500, FinalCause: "exposure"}
	mdl, _ := m.Update(connectedMsg{replica: ended})
	mm := mdl.(Model)
	if mm.takeover != takeoverPostmortem {
		t.Fatalf("takeover = %v, want takeoverPostmortem on connect to an ended world", mm.takeover)
	}
}

// TestTakeoverDismissedPostmortemStaysDismissedAcrossReconnect: per-session
// (not per-reconnect) — a transient resync must not re-annoy a player who
// already dismissed the postmortem this client session.
func TestTakeoverDismissedPostmortemStaysDismissedAcrossReconnect(t *testing.T) {
	m := testModel(t)
	m.applyEvent(runEndedEvent(1, 100, "exposure", nil))
	var mdl tea.Model = m
	mdl = update(mdl, "esc")
	mm := mdl.(Model)
	if !mm.postmortemDismissed {
		t.Fatal("precondition: postmortemDismissed should be true")
	}
	ended := sim.NewState(42, mm.gameMap)
	ended.Ended = true
	ended.RunEnd = &sim.RunEnd{Tick: 100, FinalCause: "exposure"}
	mdl2, _ := mm.Update(connectedMsg{replica: ended})
	mm2 := mdl2.(Model)
	if mm2.takeover != takeoverNone {
		t.Errorf("takeover = %v, want takeoverNone (a resync must not re-open a dismissed postmortem this session)", mm2.takeover)
	}
}

// TestTakeoverWinsBodySlotOverHelpAndConsole (spec.md Edge Cases "Help
// overlay open when a takeover fires"): the takeover renders instead of
// help/console, but doesn't force either closed — dismissing the takeover
// reveals whatever was underneath, one esc-release layer at a time.
func TestTakeoverWinsBodySlotOverHelpAndConsole(t *testing.T) {
	m := widescreenModel(t)
	var mdl tea.Model = m
	mdl = update(mdl, "?")
	if !mdl.(Model).helpOpen {
		t.Fatal("precondition: help should be open")
	}
	mm := mdl.(Model)
	mm.applyEvent(runEndedEvent(1, 100, "exposure", nil))
	view := mm.View()
	if !strings.Contains(view, "THE RUN HAS ENDED") {
		t.Errorf("the takeover should win the body slot over an open help overlay: %q", view)
	}
	if strings.Contains(view, "HELP ·") {
		t.Errorf("help must not render while a takeover is open: %q", view)
	}
	// '?' is swallowed while a takeover is open (contracts/takeovers.md §1).
	mdl2, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	if mdl2.(Model).takeover == takeoverNone {
		t.Fatal("'?' must not close the takeover")
	}
	// esc releases the takeover only — help's own state (opened earlier)
	// is untouched, so it renders again underneath.
	mdl3 := update(mdl2, "esc")
	mm3 := mdl3.(Model)
	if mm3.takeover != takeoverNone {
		t.Fatalf("takeover = %v, want takeoverNone after esc", mm3.takeover)
	}
	if !mm3.helpOpen {
		t.Error("help should still be open underneath — esc releases one layer at a time")
	}
}

// TestTakeoverWinsOverConsolePage: the takeover renders even when the
// guardian console page is open.
func TestTakeoverWinsOverConsolePage(t *testing.T) {
	m := widescreenModel(t)
	m.console = true
	m.applyEvent(runEndedEvent(1, 100, "exposure", nil))
	view := m.View()
	if !strings.Contains(view, "THE RUN HAS ENDED") {
		t.Errorf("the takeover should win the body slot over the console page: %q", view)
	}
}
