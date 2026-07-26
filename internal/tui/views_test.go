package tui

// Header governed-speed surface tests (spec 028 T014, US4-AC1/FR-015). The
// exact-string pins here are the byte-identity regression for ungoverned
// worlds and the plain-language contract for governed ones —
// contracts/status-protocol.md "TUI" §.

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/llm"
)

// TestHeaderViewUngovernedUnchanged is the regression pin: a world with no
// RequestedSpeed set (the pre-028 shape, and any world the governor hasn't
// touched) renders its header byte-identically to before T014.
func TestHeaderViewUngovernedUnchanged(t *testing.T) {
	m := testModel(t)
	m.connected = true
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{
		Tick:          100,
		GameTime:      "Day 1, 06:00",
		Speed:         "16x",
		EffectiveRate: 16.0,
	}}
	got := m.headerView()
	want := "test — tick 100 · Day 1, 06:00 · running · speed 16x (16.0 t/s) [lesson] [8 villagers]" // spec 055 lesson badge + spec 060 villager-count badge (testModel's replica seeds 8 agents, narrow width never carries the strip itself)
	if got != want {
		t.Errorf("ungoverned header = %q, want %q", got, want)
	}
}

// TestHeaderViewGoverned pins the exact governed-speed string (FR-015): the
// speed segment gains "asked <requested> — <jobs> minds in flight, debt
// <P>%" once RequestedSpeed differs from the effective Speed.
func TestHeaderViewGoverned(t *testing.T) {
	m := testModel(t)
	m.connected = true
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{
		Tick:           100,
		GameTime:       "Day 1, 06:00",
		Speed:          "16x",
		EffectiveRate:  16.0,
		RequestedSpeed: "32x",
		GovernorDebt:   1.4,
		GovernorJobs:   3,
	}}
	got := m.headerView()
	want := "test — tick 100 · Day 1, 06:00 · running · speed 16x (16.0 t/s) asked 32x — 3 minds in flight, debt 140% [lesson] [8 villagers]" // spec 055/060 badges
	if got != want {
		t.Errorf("governed header = %q, want %q", got, want)
	}
}

// TestHeaderViewGovernedSingularMind: exactly one contributing thought reads
// "1 mind in flight", not "1 minds in flight".
func TestHeaderViewGovernedSingularMind(t *testing.T) {
	m := testModel(t)
	m.connected = true
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{
		Tick:           500,
		GameTime:       "Day 2, 14:30",
		Speed:          "8x",
		EffectiveRate:  8.0,
		RequestedSpeed: "16x",
		GovernorDebt:   0.5,
		GovernorJobs:   1,
	}}
	got := m.headerView()
	want := "test — tick 500 · Day 2, 14:30 · running · speed 8x (8.0 t/s) asked 16x — 1 mind in flight, debt 50% [lesson] [8 villagers]" // spec 055/060 badges
	if got != want {
		t.Errorf("governed header (singular) = %q, want %q", got, want)
	}
}

// TestHeaderViewRequestedEqualSpeedUngoverned: RequestedSpeed present but
// equal to Speed (a transient the reducer shouldn't produce per data-model.md
// invariants, but the header must still degrade gracefully) renders
// ungoverned — the suffix only appears when the two differ.
func TestHeaderViewRequestedEqualSpeedUngoverned(t *testing.T) {
	m := testModel(t)
	m.connected = true
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{
		Tick:           100,
		GameTime:       "Day 1, 06:00",
		Speed:          "16x",
		EffectiveRate:  16.0,
		RequestedSpeed: "16x",
	}}
	got := m.headerView()
	want := "test — tick 100 · Day 1, 06:00 · running · speed 16x (16.0 t/s) [lesson] [8 villagers]" // spec 055/060 badges
	if got != want {
		t.Errorf("header with RequestedSpeed==Speed = %q, want %q (no suffix)", got, want)
	}
}

// --- T010: header LLM-condition badge (spec 034, contracts/
// provider-conditions.md "Human surfaces": "[llm: <provider> <kind>]",
// pattern of the existing [degraded] badge) ---

// TestHeaderViewLLMConditionBadge: a provider carrying an active condition
// appends the red "[llm: <provider> <kind>]" badge to the header.
func TestHeaderViewLLMConditionBadge(t *testing.T) {
	m := testModel(t)
	m.connected = true
	m.status = &ipc.StatusData{
		Clock: ipc.ClockStatus{Tick: 100, GameTime: "Day 1, 06:00", Speed: "16x", EffectiveRate: 16.0},
		LLM: &llm.Status{Providers: []llm.ProviderStatus{
			{Name: "local", Model: "cogito:3b", Condition: "model-missing"},
		}},
	}
	got := m.headerView()
	if !strings.Contains(got, "[llm: local model-missing]") {
		t.Errorf("header missing llm condition badge: %q", got)
	}
}

// TestHeaderViewLLMConditionBadgeFirstAffected: with several providers, the
// badge names the first one carrying an active condition (wire order), not
// a healthy provider ahead of it.
func TestHeaderViewLLMConditionBadgeFirstAffected(t *testing.T) {
	m := testModel(t)
	m.connected = true
	m.status = &ipc.StatusData{
		Clock: ipc.ClockStatus{Tick: 100, GameTime: "Day 1, 06:00", Speed: "16x", EffectiveRate: 16.0},
		LLM: &llm.Status{Providers: []llm.ProviderStatus{
			{Name: "cloud", Model: "claude-opus-4-8"},
			{Name: "local", Model: "cogito:3b", Condition: "endpoint-unreachable"},
		}},
	}
	got := m.headerView()
	if !strings.Contains(got, "[llm: local endpoint-unreachable]") {
		t.Errorf("header should badge the first affected provider: %q", got)
	}
}

// TestHeaderViewNoLLMConditionUnchanged: no LLM status, and an LLM status
// with every provider healthy, both render with no llm badge — the T010
// regression pin alongside the pre-034 header tests above.
func TestHeaderViewNoLLMConditionUnchanged(t *testing.T) {
	m := testModel(t)
	m.connected = true
	m.status = &ipc.StatusData{
		Clock: ipc.ClockStatus{Tick: 100, GameTime: "Day 1, 06:00", Speed: "16x", EffectiveRate: 16.0},
		LLM:   &llm.Status{Providers: []llm.ProviderStatus{{Name: "local", Model: "cogito:3b", Up: true}}},
	}
	got := m.headerView()
	if strings.Contains(got, "[llm:") {
		t.Errorf("healthy LLM status should render no llm badge: %q", got)
	}
}

// --- T004: header suppression badge (spec 037 US1, FR-005) ---

// TestHeaderViewSuppressionBadge: ≥1 suppressed horizon entry appends the
// "[suppressed: <classes>]" badge, listing the suppressed classes in wire
// order (thinking classes excluded). headerView is shared by the widescreen
// and narrow layouts, so this pins the badge for both.
func TestHeaderViewSuppressionBadge(t *testing.T) {
	m := testModel(t)
	m.connected = true
	m.status = &ipc.StatusData{
		Clock: ipc.ClockStatus{Tick: 100, GameTime: "Day 1, 06:00", Speed: "32x", EffectiveRate: 32.0},
		Horizon: []ipc.HorizonClass{
			{Class: "planner", Suppressed: true},
			{Class: "conversation", Suppressed: true},
			{Class: "meeting", Suppressed: false},
		},
	}
	got := m.headerView()
	if !strings.Contains(got, "[suppressed: planner, conversation]") {
		t.Errorf("header missing suppression badge (wire-ordered, thinking excluded): %q", got)
	}
}

// --- spec 044 US1: postmortem posture (ENDED header token, inert clock keys,
// footer hint) — the badge-test pattern above ---

// TestHeaderViewEndedFromStatus: the status poll / push-refreshed
// ClockStatus.Ended source of the dual derivation (research R12) — ENDED
// replaces the running token.
func TestHeaderViewEndedFromStatus(t *testing.T) {
	m := testModel(t)
	m.connected = true
	m.replica = nil // isolate the status source
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{
		Tick: 100, GameTime: "Day 1, 06:00", Speed: "16x", Ended: true, EndedDay: 1,
	}}
	got := m.headerView()
	if !strings.Contains(got, "ENDED") {
		t.Errorf("header missing ENDED token: %q", got)
	}
	if strings.Contains(got, "running") || strings.Contains(got, "PAUSED") {
		t.Errorf("ENDED must replace the running/PAUSED token: %q", got)
	}
}

// TestHeaderViewEndedFromReplica: the replica-State source — a client
// attaching to an already-ended world sees ENDED from the state snapshot
// alone, before any status carries the flag (the snapshot path never replays
// folded events, research R12).
func TestHeaderViewEndedFromReplica(t *testing.T) {
	m := testModel(t)
	m.connected = true
	m.replica.Ended = true
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{
		Tick: 100, GameTime: "Day 1, 06:00", Speed: "16x", Paused: true, // ended outranks PAUSED
	}}
	got := m.headerView()
	if !strings.Contains(got, "ENDED") {
		t.Errorf("header missing ENDED token from the replica source: %q", got)
	}
	if strings.Contains(got, "PAUSED") {
		t.Errorf("ENDED must outrank PAUSED: %q", got)
	}
}

// TestHeaderViewNotEndedUnchanged: a living world renders no ENDED token —
// the regression pin beside the ungoverned-header pin above.
func TestHeaderViewNotEndedUnchanged(t *testing.T) {
	m := testModel(t)
	m.connected = true
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{
		Tick: 100, GameTime: "Day 1, 06:00", Speed: "16x", EffectiveRate: 16.0,
	}}
	if got := m.headerView(); strings.Contains(got, "ENDED") {
		t.Errorf("living world's header carries an ENDED token: %q", got)
	}
}

// TestFooterViewEndedHint: the clock keys are inert on an ended world, so
// the footer's pause hint gives way to the run-ended hint.
func TestFooterViewEndedHint(t *testing.T) {
	m := testModel(t)
	m.connected = true
	m.replica.Ended = true
	got := m.footerView()
	if !strings.Contains(got, "run ended") {
		t.Errorf("footer missing the run-ended hint: %q", got)
	}
	if strings.Contains(got, "space pause") {
		t.Errorf("footer still advertises the inert pause key: %q", got)
	}
}

// TestEndedClockKeysInert: space and the speed brackets issue no command on
// an ended world — gated client-side so the daemon's refusal error is never
// mistaken for a disconnect.
func TestEndedClockKeysInert(t *testing.T) {
	m := testModel(t)
	m.connected = true
	m.replica.Ended = true
	m.status = &ipc.StatusData{Clock: ipc.ClockStatus{Speed: "4x", Ended: true, EndedDay: 1}}
	for _, k := range []string{" ", "[", "]"} {
		_, cmd := m.handleGlobalKey(key(k))
		if cmd != nil {
			t.Errorf("key %q issued a command on an ended world", k)
		}
	}
}

// TestHeaderViewNoSuppressionBadge: a horizon with everything thinking, and a
// world with no horizon at all, both render no suppression badge — the FR-005
// "MUST NOT show it otherwise" pin.
func TestHeaderViewNoSuppressionBadge(t *testing.T) {
	for _, tc := range []struct {
		name    string
		horizon []ipc.HorizonClass
	}{
		{"all thinking", []ipc.HorizonClass{{Class: "planner"}, {Class: "conversation"}, {Class: "meeting"}}},
		{"no horizon", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := testModel(t)
			m.connected = true
			m.status = &ipc.StatusData{
				Clock:   ipc.ClockStatus{Tick: 100, GameTime: "Day 1, 06:00", Speed: "8x", EffectiveRate: 8.0},
				Horizon: tc.horizon,
			}
			if got := m.headerView(); strings.Contains(got, "[suppressed:") {
				t.Errorf("%s: header should carry no suppression badge: %q", tc.name, got)
			}
		})
	}
}
